// Copyright 2024 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logrepl

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dolthub/doltgresql/postgres/parser/uuid"
)

const outputPlugin = "pgoutput"

type rcvMsg struct {
	msg pgproto3.BackendMessage
	err error
}

type LogicalReplicator struct {
	primaryDns      string
	replicationConn *pgx.Conn

	walFilePath     string
	running         bool
	messageReceived bool
	stop            chan struct{}
	mu              *sync.Mutex
}

// NewLogicalReplicator creates a new logical replicator instance which connects to the primary and replication
// databases using the connection strings provided. The connection to the replica is established immediately, and the
// connection to the primary is established when StartReplication is called.
func NewLogicalReplicator(walFilePath string, primaryDns string, replicationDns string) (*LogicalReplicator, error) {
	conn, err := pgx.Connect(context.Background(), replicationDns)
	if err != nil {
		return nil, err
	}

	return &LogicalReplicator{
		primaryDns:      primaryDns,
		replicationConn: conn,
		walFilePath:     walFilePath,
		mu:              &sync.Mutex{},
	}, nil
}

// PrimaryDns returns the DNS for the primary database. Not suitable for RPCs used in replication e.g.
// StartReplication. See ReplicationDns.
func (r *LogicalReplicator) PrimaryDns() string {
	return r.primaryDns
}

// ReplicationDns returns the DNS for the primary database with the replication query parameter appended. Not suitable
// for normal query RPCs.
func (r *LogicalReplicator) ReplicationDns() string {
	if strings.Contains(r.primaryDns, "?") {
		return fmt.Sprintf("%s&replication=database", r.primaryDns)
	}
	return fmt.Sprintf("%s?replication=database", r.primaryDns)
}

// CaughtUp returns true if the replication slot is caught up to the primary, and false otherwise. This only works if
// there is only a single replication slot on the primary, so it's only suitable for testing.
func (r *LogicalReplicator) CaughtUp() (bool, error) {
	r.mu.Lock()
	if !r.messageReceived {
		r.mu.Unlock()
		// We can't query the replication state until after receiving our first message
		return false, nil
	}
	r.mu.Unlock()

	conn, err := pgx.Connect(context.Background(), r.PrimaryDns())
	if err != nil {
		return false, err
	}
	defer conn.Close(context.Background())

	result, err := conn.Query(context.Background(), "SELECT pg_wal_lsn_diff(write_lsn, sent_lsn) AS replication_lag FROM pg_stat_replication")
	if err != nil {
		return false, err
	}

	defer result.Close()

	for result.Next() {
		rows, err := result.Values()
		if err != nil {
			return false, err
		}

		row := rows[0]
		lag, ok := row.(pgtype.Numeric)
		if ok && lag.Valid {
			log.Printf("Current replication lag: %v", row)
			return lag.Int.Int64() >= 0, nil
		} else {
			log.Printf("Replication lag unknown: %v", row)
		}
	}

	if result.Err() != nil {
		return false, result.Err()
	}

	// if we got this far, then there is no running replication thread, which we interpret as caught up
	return true, nil
}

// maxConsecutiveFailures is the maximum number of consecutive RPC errors that can occur before we stop
// the replication thread
const maxConsecutiveFailures = 10

var errShutdownRequested = errors.New("shutdown requested")

// StartReplication starts the replication process for the given slot name. This function blocks until replication is
// stopped via the Stop method, or an error occurs.
func (r *LogicalReplicator) StartReplication(slotName string) error {
	standbyMessageTimeout := 10 * time.Second
	nextStandbyMessageDeadline := time.Now().Add(standbyMessageTimeout)
	relationsV2 := map[uint32]*pglogrepl.RelationMessageV2{}
	typeMap := pgtype.NewMap()

	// whenever we get StreamStartMessage we set inStream to true and then pass it to DecodeV2 function
	// on StreamStopMessage we set it back to false
	inStream := false
	
	// We selectively ignore messages that are from before our last flush, which can be resent by postgres in certain
	// crash scenarios. Postgres sends messages in batches based on changes in a transaction, beginning with a Begin
	// message that records the last WAL position of the transaction. The individual INSERT, UPDATE, DELETE messages are
	// sent, each tagged with the WAL position of that tuple write. This WAL position can be before the last flush LSN 
	// in some cases. Whether we ignore them or not has nothing to do with the WAL position of any individual write, but 
	// the final LSN of the transaction, as recorded in the Begin message. So for every Begin, we decide whether to
	// process or ignore all messages until a corresponding Commit message. 
	processMessages := false

	// lsn is the last WAL position we have received from the server, which we send back to the server via
	// SendStandbyStatusUpdate after every message we get. Postgres tracks this LSN for each slot, which allows us to
	// resume where we left off in the case of an interruption.
	var lsn pglogrepl.LSN
	lsn, err := r.readWALPosition()
	if err != nil {
		return err
	}

	var primaryConn *pgconn.PgConn
	defer func() {
		if primaryConn != nil {
			_ = primaryConn.Close(context.Background())
		}
		// We always shut down here and only here, so we do the cleanup on thread exit in exactly one place
		r.shutdown()
	}()

	connErrCnt := 0
	handleErrWithRetry := func(err error) error {
		if err != nil {
			connErrCnt++
			if connErrCnt < maxConsecutiveFailures {
				log.Printf("Error: %v. Retrying", err)
				_ = primaryConn.Close(context.Background())
				primaryConn = nil
				return nil
			}
		} else {
			connErrCnt = 0
		}

		return err
	}

	sendStandbyStatusUpdate := func(currentLSN pglogrepl.LSN) error {
		// The StatusUpdate message wants us to respond with the current position in the WAL + 1:
		// https://www.postgresql.org/docs/current/protocol-replication.html
		lsn := currentLSN + 1
		err := pglogrepl.SendStandbyStatusUpdate(context.Background(), primaryConn, pglogrepl.StandbyStatusUpdate{WALWritePosition: lsn + 1})
		if err != nil {
			return handleErrWithRetry(err)
		}

		log.Printf("Sent Standby status message at %s\n", lsn.String())
		nextStandbyMessageDeadline = time.Now().Add(standbyMessageTimeout)
		return nil
	}

	log.Println("Starting replicator")
	r.mu.Lock()
	r.running = true
	r.stop = make(chan struct{})
	r.mu.Unlock()

	for {
		err := func() error {
			// Shutdown if requested
			select {
			case <-r.stop:
				return errShutdownRequested
			default:
				// continue below
			}

			if primaryConn == nil {
				var err error
				primaryConn, err = r.beginReplication(slotName, lsn)
				if err != nil {
					// unlike other error cases, back off a little here, since we're likely to just get the same error again
					// on initial replication establishment
					time.Sleep(100 * time.Millisecond)
					return handleErrWithRetry(err)
				}
			}

			if time.Now().After(nextStandbyMessageDeadline) {
				err := sendStandbyStatusUpdate(lsn)
				if err != nil {
					return err
				}
				if primaryConn == nil {
					// if we've lost the connection, we'll re-establish it on the next pass through the loop
					return nil
				}
			}

			ctx, cancel := context.WithDeadline(context.Background(), nextStandbyMessageDeadline)
			receiveMsgChan := make(chan rcvMsg)
			go func() {
				rawMsg, err := primaryConn.ReceiveMessage(ctx)
				receiveMsgChan <- rcvMsg{msg: rawMsg, err: err}
			}()

			var msgAndErr rcvMsg
			select {
			case <-r.stop:
				cancel()
				return errShutdownRequested
			case <-ctx.Done():
				cancel()
				return nil
			case msgAndErr = <-receiveMsgChan:
				cancel()
			}

			if msgAndErr.err != nil {
				if pgconn.Timeout(msgAndErr.err) {
					return nil
				} else {
					return handleErrWithRetry(msgAndErr.err)
				}
			}

			r.mu.Lock()
			r.messageReceived = true
			r.mu.Unlock()

			rawMsg := msgAndErr.msg
			if errMsg, ok := rawMsg.(*pgproto3.ErrorResponse); ok {
				return fmt.Errorf("received Postgres WAL error: %+v", errMsg)
			}

			msg, ok := rawMsg.(*pgproto3.CopyData)
			if !ok {
				log.Printf("Received unexpected message: %T\n", rawMsg)
				return nil
			}

			switch msg.Data[0] {
			case pglogrepl.PrimaryKeepaliveMessageByteID:
				pkm, err := pglogrepl.ParsePrimaryKeepaliveMessage(msg.Data[1:])
				if err != nil {
					log.Fatalln("ParsePrimaryKeepaliveMessage failed:", err)
				}
				log.Println("Primary Keepalive Message =>", "ServerWALEnd:", pkm.ServerWALEnd, "ServerTime:", pkm.ServerTime, "ReplyRequested:", pkm.ReplyRequested)

				if pkm.ReplyRequested {
					if pkm.ServerWALEnd > lsn {
						lsn = pkm.ServerWALEnd
					}
					// Send our reply the next time through the loop
					nextStandbyMessageDeadline = time.Time{}
				}
			case pglogrepl.XLogDataByteID:
				xld, err := pglogrepl.ParseXLogData(msg.Data[1:])
				if err != nil {
					return err
				}

				// TODO next: need to track whether we have yet received any message past the last LSN we wrote to the WAL, 
				//  in order to handle the case where we get LSNs out of order 
				updateNeeded, err := r.processMessage(lsn, xld, relationsV2, typeMap, &inStream, &processMessages)
				if err != nil {
					// TODO: do we need more than one handler, one for each connection?
					return handleErrWithRetry(err)
				}

				// TODO: we have a two-phase commit race here: if the WAL file update doesn't happen before the process crashes,
				//  we will receive a duplicate LSN the next time we start replication. A better solution would be to write the
				//  LSN directly into the DoltCommit message, and then parsing this message back out when we begin replication
				//  next.
				if updateNeeded && xld.ServerWALEnd > lsn {
					lsn = xld.ServerWALEnd
					err := r.writeWALPosition(lsn)
					if err != nil {
						return err
					}
				} else {
					log.Printf("No update needed for LSN %s, r.lsn is %s\n", xld.ServerWALEnd.String(), lsn.String())
				}

				err = sendStandbyStatusUpdate(xld.ServerWALEnd)
				if err != nil {
					return err
				}

				if primaryConn == nil {
					// if we've lost the connection, we'll re-establish it on the next pass through the loop
					return nil
				}
			default:
				log.Printf("Received unexpected message: %T\n", rawMsg)
			}

			return nil
		}()

		if err != nil {
			if errors.Is(err, errShutdownRequested) {
				return nil
			}
			log.Println("Error during replication:", err)
			return err
		}
	}
}

func (r *LogicalReplicator) shutdown() {
	r.mu.Lock()
	defer r.mu.Unlock()
	log.Print("shutting down replicator")
	r.running = false
	close(r.stop)
}

// Running returns whether replication is currently running
func (r *LogicalReplicator) Running() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}

// Stop stops the replication process and blocks until clean shutdown occurs.
func (r *LogicalReplicator) Stop() {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return
	}
	r.mu.Unlock()

	log.Print("stopping replication...")
	r.stop <- struct{}{}
	// wait for the channel to be closed, acknowledging that the replicator has stopped
	<-r.stop
}

// replicateQuery executes the query provided on the replica connection
func (r *LogicalReplicator) replicateQuery(query string) error {
	log.Printf("replicating query: %s", query)
	_, err := r.replicationConn.Exec(context.Background(), query)
	return err
}

// beginReplication starts a new replication connection to the primary server and returns it along with the current
// log sequence number (LSN) for continued status updates to the primary.
func (r *LogicalReplicator) beginReplication(slotName string, lsn pglogrepl.LSN) (*pgconn.PgConn, error) {
	conn, err := pgconn.Connect(context.Background(), r.ReplicationDns())
	if err != nil {
		return nil, err
	}

	// streaming of large transactions is available since PG 14 (protocol version 2)
	// we also need to set 'streaming' to 'true'
	pluginArguments := []string{
		"proto_version '2'",
		fmt.Sprintf("publication_names '%s'", slotName),
		"messages 'true'",
		"streaming 'true'",
	}

	// LSN(0) is used to use the last confirmed LSN for this slot
	log.Printf("Starting logical replication on slot %s at WAL location %s", slotName, lsn)
	err = pglogrepl.StartReplication(context.Background(), conn, slotName, lsn, pglogrepl.StartReplicationOptions{PluginArgs: pluginArguments})
	if err != nil {
		return nil, err
	}
	log.Println("Logical replication started on slot", slotName)

	return conn, nil
}

// DropPublication drops the publication with the given name if it exists. Mostly useful for testing.
func DropPublication(primaryDns, slotName string) error {
	conn, err := pgconn.Connect(context.Background(), primaryDns)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	result := conn.Exec(context.Background(), fmt.Sprintf("DROP PUBLICATION IF EXISTS %s;", slotName))
	_, err = result.ReadAll()
	return err
}

// CreatePublication creates a publication with the given name if it does not already exist. Mostly useful for testing.
// Customers should run the CREATE PUBLICATION command on their primary server manually, specifying whichever tables
// they want to replicate.
func CreatePublication(primaryDns, slotName string) error {
	conn, err := pgconn.Connect(context.Background(), primaryDns)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	result := conn.Exec(context.Background(), fmt.Sprintf("CREATE PUBLICATION %s FOR ALL TABLES;", slotName))
	_, err = result.ReadAll()
	return err
}

// DropReplicationSlot drops the replication slot with the given name. Any error from the slot not existing is ignored.
func (r *LogicalReplicator) DropReplicationSlot(slotName string) error {
	conn, err := pgconn.Connect(context.Background(), r.ReplicationDns())
	if err != nil {
		return err
	}

	_ = pglogrepl.DropReplicationSlot(context.Background(), conn, slotName, pglogrepl.DropReplicationSlotOptions{})
	return nil
}

// CreateReplicationSlotIfNecessary creates the replication slot named if it doesn't already exist.
func (r *LogicalReplicator) CreateReplicationSlotIfNecessary(slotName string) error {
	conn, err := pgx.Connect(context.Background(), r.PrimaryDns())
	if err != nil {
		return err
	}

	rows, err := conn.Query(context.Background(), "select * from pg_replication_slots where slot_name = $1", slotName)
	if err != nil {
		return err
	}

	slotExists := false
	defer rows.Close()
	for rows.Next() {
		_, err := rows.Values()
		if err != nil {
			return err
		}
		slotExists = true
	}

	if rows.Err() != nil {
		return rows.Err()
	}

	// We need a different connection to create the replication slot
	conn, err = pgx.Connect(context.Background(), r.ReplicationDns())
	if err != nil {
		return err
	}

	if !slotExists {
		_, err = pglogrepl.CreateReplicationSlot(context.Background(), conn.PgConn(), slotName, outputPlugin, pglogrepl.CreateReplicationSlotOptions{})
		if err != nil {
			pgErr, ok := err.(*pgconn.PgError)
			if ok && pgErr.Code == "42710" {
				// replication slot already exists, we can ignore this error
			} else {
				return err
			}
		}

		log.Println("Created replication slot:", slotName)
	}

	return nil
}

// processMessage processes a logical replication message as appropriate. A couple important aspects:
//  1. Relation messages describe tables being replicated and are used to build a type map for decoding tuples
//  2. INSERT/UPDATE/DELETE messages describe changes to rows that must be applied to the replica.
//     These describe a row in the form of a tuple, and are used to construct a query to apply the change to the replica.
//
// Returns a boolean true if the message was a write that should be acknowledged to the server, and an error if one
// occurred.
func (r *LogicalReplicator) processMessage(
		lsn pglogrepl.LSN,
		xld pglogrepl.XLogData,
		relations map[uint32]*pglogrepl.RelationMessageV2,
		typeMap *pgtype.Map,
		inStream *bool,
		processMessages *bool,
) (bool, error) {
	walData := xld.WALData
	logicalMsg, err := pglogrepl.ParseV2(walData, *inStream)
	if err != nil {
		return false, err
	}

	log.Printf("XLogData (%T) => WALStart %s ServerWALEnd %s ServerTime %s", logicalMsg, xld.WALStart, xld.ServerWALEnd, xld.ServerTime)

	switch logicalMsg := logicalMsg.(type) {
	case *pglogrepl.RelationMessageV2:
		relations[logicalMsg.RelationID] = logicalMsg
	case *pglogrepl.BeginMessage:
		// Indicates the beginning of a group of changes in a transaction.
		// This is only sent for committed transactions. We won't get any events from rolled back transactions.
		
		if lsn > logicalMsg.FinalLSN {
			log.Printf("Received stale message, ignoring. Current LSN: %s Message LSN: %s", lsn, logicalMsg.FinalLSN)
			*processMessages = false
			return false, nil
		}
		
		*processMessages = true
		log.Printf("BeginMessage: %v", logicalMsg)
		err = r.replicateQuery("START TRANSACTION")
		if err != nil {
			return false, err
		}
	case *pglogrepl.CommitMessage:
		log.Printf("CommitMessage: %v", logicalMsg)
		err = r.replicateQuery("COMMIT")
		if err != nil {
			return false, err
		}
		*processMessages = false
	case *pglogrepl.InsertMessageV2:
		if !*processMessages {
			log.Printf("Received stale message, ignoring. Current LSN: %s Message LSN: %s", lsn, xld.ServerWALEnd)
			return false, nil
		}

		rel, ok := relations[logicalMsg.RelationID]
		if !ok {
			log.Fatalf("unknown relation ID %d", logicalMsg.RelationID)
		}

		columnStr := strings.Builder{}
		valuesStr := strings.Builder{}
		for idx, col := range logicalMsg.Tuple.Columns {
			if idx > 0 {
				columnStr.WriteString(", ")
				valuesStr.WriteString(", ")
			}

			colName := rel.Columns[idx].Name
			columnStr.WriteString(colName)

			switch col.DataType {
			case 'n': // null
				valuesStr.WriteString("NULL")
			case 't': // text

				// We have to round-trip the data through the encodings to get an accurate text rep back
				val, err := decodeTextColumnData(typeMap, col.Data, rel.Columns[idx].DataType)
				if err != nil {
					log.Fatalln("error decoding column data:", err)
				}
				colData, err := encodeColumnData(typeMap, val, rel.Columns[idx].DataType)
				if err != nil {
					return false, err
				}
				valuesStr.WriteString(colData)
			default:
				log.Printf("unknown column data type: %c", col.DataType)
			}
		}

		err = r.replicateQuery(fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s)", rel.Namespace, rel.RelationName, columnStr.String(), valuesStr.String()))
		if err != nil {
			return false, err
		}

		return true, nil
	case *pglogrepl.UpdateMessageV2:
		if !*processMessages {
			log.Printf("Received stale message, ignoring. Current LSN: %s Message LSN: %s", lsn, xld.ServerWALEnd)
			return false, nil
		}

		// TODO: this won't handle primary key changes correctly
		// TODO: this probably doesn't work for unkeyed tables
		rel, ok := relations[logicalMsg.RelationID]
		if !ok {
			log.Fatalf("unknown relation ID %d", logicalMsg.RelationID)
		}

		updateStr := strings.Builder{}
		whereStr := strings.Builder{}
		for idx, col := range logicalMsg.NewTuple.Columns {
			colName := rel.Columns[idx].Name
			colFlags := rel.Columns[idx].Flags

			var stringVal string
			switch col.DataType {
			case 'n': // null
				stringVal = "NULL"
			case 'u': // unchanged toast
			case 't': // text
				val, err := decodeTextColumnData(typeMap, col.Data, rel.Columns[idx].DataType)
				if err != nil {
					log.Fatalln("error decoding column data:", err)
				}

				stringVal, err = encodeColumnData(typeMap, val, rel.Columns[idx].DataType)
				if err != nil {
					return false, err
				}
			default:
				log.Printf("unknown column data type: %c", col.DataType)
			}

			// TODO: quote column names?
			if colFlags == 0 {
				if updateStr.Len() > 0 {
					updateStr.WriteString(", ")
				}
				updateStr.WriteString(fmt.Sprintf("%s = %v", colName, stringVal))
			} else {
				if whereStr.Len() > 0 {
					updateStr.WriteString(", ")
				}
				whereStr.WriteString(fmt.Sprintf("%s = %v", colName, stringVal))
			}
		}

		err = r.replicateQuery(fmt.Sprintf("UPDATE %s.%s SET %s%s", rel.Namespace, rel.RelationName, updateStr.String(), whereClause(whereStr)))
		if err != nil {
			return false, err
		}

		return true, nil
	case *pglogrepl.DeleteMessageV2:
		if !*processMessages {
			log.Printf("Received stale message, ignoring. Current LSN: %s Message LSN: %s", lsn, xld.ServerWALEnd)
			return false, nil
		}

		// TODO: this probably doesn't work for unkeyed tables
		rel, ok := relations[logicalMsg.RelationID]
		if !ok {
			log.Fatalf("unknown relation ID %d", logicalMsg.RelationID)
		}

		whereStr := strings.Builder{}
		for idx, col := range logicalMsg.OldTuple.Columns {
			colName := rel.Columns[idx].Name
			colFlags := rel.Columns[idx].Flags

			var stringVal string
			switch col.DataType {
			case 'n': // null
				stringVal = "NULL"
			case 'u': // unchanged toast
			case 't': // text
				val, err := decodeTextColumnData(typeMap, col.Data, rel.Columns[idx].DataType)
				if err != nil {
					log.Fatalln("error decoding column data:", err)
				}

				stringVal, err = encodeColumnData(typeMap, val, rel.Columns[idx].DataType)
				if err != nil {
					return false, err
				}
			default:
				log.Printf("unknown column data type: %c", col.DataType)
			}

			if colFlags == 0 {
				// nothing to do
			} else {
				if whereStr.Len() > 0 {
					whereStr.WriteString(", ")
				}
				whereStr.WriteString(fmt.Sprintf("%s = %v", colName, stringVal))
			}
		}

		err = r.replicateQuery(fmt.Sprintf("DELETE FROM %s.%s WHERE %s", rel.Namespace, rel.RelationName, whereStr.String()))
		if err != nil {
			return false, err
		}

		return true, nil
	case *pglogrepl.TruncateMessageV2:
		log.Printf("truncate for xid %d\n", logicalMsg.Xid)
	case *pglogrepl.TypeMessageV2:
		log.Printf("typeMessage for xid %d\n", logicalMsg.Xid)
	case *pglogrepl.OriginMessage:
		log.Printf("originMessage for xid %s\n", logicalMsg.Name)
	case *pglogrepl.LogicalDecodingMessageV2:
		log.Printf("Logical decoding message: %q, %q, %d", logicalMsg.Prefix, logicalMsg.Content, logicalMsg.Xid)
	case *pglogrepl.StreamStartMessageV2:
		*inStream = true
		log.Printf("Stream start message: xid %d, first segment? %d", logicalMsg.Xid, logicalMsg.FirstSegment)
	case *pglogrepl.StreamStopMessageV2:
		*inStream = false
		log.Printf("Stream stop message")
	case *pglogrepl.StreamCommitMessageV2:
		log.Printf("Stream commit message: xid %d", logicalMsg.Xid)
	case *pglogrepl.StreamAbortMessageV2:
		log.Printf("Stream abort message: xid %d", logicalMsg.Xid)
	default:
		log.Printf("Unknown message type in pgoutput stream: %T", logicalMsg)
	}

	return false, nil
}

// readWALPosition reads the recorded WAL position from the WAL position file
func (r *LogicalReplicator) readWALPosition() (pglogrepl.LSN, error) {
	walFileContents, err := os.ReadFile(r.walFilePath)
	if err != nil {
		// if the file doesn't exist, consider this a cold start and return 0
		if os.IsNotExist(err) {
			return pglogrepl.LSN(0), nil
		}
		return 0, err
	}

	return pglogrepl.ParseLSN(string(walFileContents))
}

// writeWALPosition writes the recorded WAL position to the WAL position file
func (r *LogicalReplicator) writeWALPosition(lsn pglogrepl.LSN) error {
	// We write a single byte past the last LSN we flushed because our next startup will use that as our starting point.
	// The LSN given to the StartReplication call is inclusive, so we need to exclude the last one we have processed.
	writeLsn := lsn + 1
	return os.WriteFile(r.walFilePath, []byte(writeLsn.String()), 0644)
}

// whereClause returns a WHERE clause string with the contents of the builder if it's non-empty, or the empty
// string otherwise
func whereClause(str strings.Builder) string {
	if str.Len() > 0 {
		return " WHERE " + str.String()
	}
	return ""
}

// decodeTextColumnData decodes the given data using the given data type OID and returns the result as a golang value
func decodeTextColumnData(mi *pgtype.Map, data []byte, dataType uint32) (interface{}, error) {
	if dt, ok := mi.TypeForOID(dataType); ok {
		return dt.Codec.DecodeValue(mi, dataType, pgtype.TextFormatCode, data)
	}
	return string(data), nil
}

// encodeColumnData encodes the given data using the given data type OID and returns the result as a string to be
// used in an INSERT or other DML query.
func encodeColumnData(mi *pgtype.Map, data interface{}, dataType uint32) (string, error) {
	var value string
	if dt, ok := mi.TypeForOID(dataType); ok {
		e := dt.Codec.PlanEncode(mi, dataType, pgtype.TextFormatCode, data)
		if e != nil {
			encoded, err := e.Encode(data, nil)
			if err != nil {
				return "", err
			}
			value = string(encoded)
		} else {
			// no encoder for this type, use the string representation
			value = fmt.Sprintf("%v", data)
		}
	} else {
		value = fmt.Sprintf("%v", data)
	}

	// Some types need additional quoting after encoding
	switch data := data.(type) {
	case string, time.Time, pgtype.Time, bool:
		return fmt.Sprintf("'%s'", value), nil
	case [16]byte:
		// TODO: should we actually register an encoder for this type?
		uid := uuid.UUID(data)
		return fmt.Sprintf("'%s'", uid.String()), nil
	default:
		return value, nil
	}
}
