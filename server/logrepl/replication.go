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
	"fmt"
	"log"
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
	// lsn is the last WAL position we have received from the server, which we send back to the server via 
	// SendStandbyStatusUpdate after every message we get. Postgres tracks this LSN for each slot, which allows us to 
	// resume where we left off in the case of an interruption.
	lsn             pglogrepl.LSN
	running         bool
	stop            chan struct{}
	mu              *sync.Mutex
}

// NewLogicalReplicator creates a new logical replicator instance which connects to the primary and replication
// databases using the connection strings provided. The connection to the replica is established immediately, and the
// connection to the primary is established when StartReplication is called.
func NewLogicalReplicator(primaryDns string, replicationDns string) (*LogicalReplicator, error) {
	conn, err := pgx.Connect(context.Background(), replicationDns)
	if err != nil {
		return nil, err
	}

	return &LogicalReplicator{
		primaryDns:      primaryDns,
		replicationConn: conn,
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
	// if we haven't received any messages yet, we can't tell if we're caught up or not
	r.mu.Lock()
	if r.lsn == 0 {
		r.mu.Unlock()
		log.Printf("No replication messages received yet")
		return false, nil
	}
	r.mu.Unlock()
	
	conn, err := pgx.Connect(context.Background(), r.PrimaryDns())
	if err != nil {
		return false, err
	}

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
			log.Printf("Replication lag: %d", lag.Int.Int64())
			return lag.Int.Int64() == 0, nil
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

const maxConsecutiveFailures = 10

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

	// We fail after 3 consecutive network errors excluding timeouts. Any successful RPC resets the counter.
	connErrCnt := 0
	var primaryConn *pgconn.PgConn
	
	defer func() {
		if primaryConn != nil {
			_ = primaryConn.Close(context.Background())
		}
		r.mu.Lock()
		r.running = false
		r.mu.Unlock()
	}()
	
	sendStandbyStatusUpdate := func(lsn pglogrepl.LSN) error {
		err := pglogrepl.SendStandbyStatusUpdate(context.Background(), primaryConn, pglogrepl.StandbyStatusUpdate{WALWritePosition: lsn})
		if err != nil {
			connErrCnt++
			if connErrCnt < maxConsecutiveFailures {
				// re-establish connection on next pass through the loop
				_ = primaryConn.Close(context.Background())
				primaryConn = nil
				return nil
			}

			return err
		}

		connErrCnt = 0
		log.Printf("Sent Standby status message at %s\n", lsn.String())
		nextStandbyMessageDeadline = time.Now().Add(standbyMessageTimeout)
		return nil
	}

	r.mu.Lock()
	r.lsn = 0
	r.running = true
	r.stop = make(chan struct{})
	r.mu.Unlock()

	for {
		// Shutdown if requested
		select {
		case <-r.stop:
			r.shutdown()
			return nil
		default:
			// continue
		}

		if primaryConn == nil {
			// TODO: not sure if this retry logic is correct, with some failures we appear to miss events that aren't
			//  sent again
			var err error
			primaryConn, err = r.beginReplication(slotName)
			if err != nil {
				return err
			}
		}
		
		if time.Now().After(nextStandbyMessageDeadline) {
			r.mu.Lock()
			err := sendStandbyStatusUpdate(r.lsn)
			if err != nil {
				r.mu.Unlock()
				return err
			}
			r.mu.Unlock()
		}

		log.Printf("attempting to receive message from primary server with time now = %s, deadline = %s", time.Now().String(), nextStandbyMessageDeadline.String())
		ctx, cancel := context.WithDeadline(context.Background(), nextStandbyMessageDeadline)
		receiveMsgChan := make(chan rcvMsg)
		go func() {
			rawMsg, err := primaryConn.ReceiveMessage(ctx)
			receiveMsgChan <- rcvMsg{msg: rawMsg, err: err}
		}()

		var msgAndErr rcvMsg
		select {
		case <-r.stop:
			log.Println("Received stop signal")
			cancel()
			r.shutdown()
			return nil
		case <-ctx.Done():
			log.Println("Context done")
			cancel()
			continue
		case msgAndErr = <-receiveMsgChan:
			if msgAndErr.msg != nil {
				log.Println("Received message from primary server")
			} else {
				log.Printf("Error received from primary server: %s", msgAndErr.err.Error())
			}
			cancel()
		}

		if msgAndErr.err != nil {
			if pgconn.Timeout(msgAndErr.err) {
				continue
			} else {
				connErrCnt++
				if connErrCnt < maxConsecutiveFailures {
					// re-establish connection on next pass through the loop
					_ = primaryConn.Close(context.Background())
					primaryConn = nil
					continue
				}
			}

			return msgAndErr.err
		}

		rawMsg := msgAndErr.msg
		connErrCnt = 0
		if errMsg, ok := rawMsg.(*pgproto3.ErrorResponse); ok {
			return fmt.Errorf("received Postgres WAL error: %+v", errMsg)
		}

		msg, ok := rawMsg.(*pgproto3.CopyData)
		if !ok {
			log.Printf("Received unexpected message: %T\n", rawMsg)
			continue
		}

		switch msg.Data[0] {
		case pglogrepl.PrimaryKeepaliveMessageByteID:
			pkm, err := pglogrepl.ParsePrimaryKeepaliveMessage(msg.Data[1:])
			if err != nil {
				log.Fatalln("ParsePrimaryKeepaliveMessage failed:", err)
			}
			log.Println("Primary Keepalive Message =>", "ServerWALEnd:", pkm.ServerWALEnd, "ServerTime:", pkm.ServerTime, "ReplyRequested:", pkm.ReplyRequested)
			
			r.mu.Lock()
			if pkm.ServerWALEnd > r.lsn {
				r.lsn = pkm.ServerWALEnd
			}
			r.mu.Unlock()
			
			if pkm.ReplyRequested {
				// Send our reply the next time through the loop
				nextStandbyMessageDeadline = time.Time{}
			}
		case pglogrepl.XLogDataByteID:
			xld, err := pglogrepl.ParseXLogData(msg.Data[1:])
			if err != nil {
				return err
			}

			log.Printf("XLogData => WALStart %s ServerWALEnd %s ServerTime %s WALData:\n", xld.WALStart, xld.ServerWALEnd, xld.ServerTime)
			err = r.processMessage(xld.WALData, relationsV2, typeMap, &inStream)
			if err != nil {
				return err
			}

			// TODO: we have a two-phase commit race here: if the call to update the standby fails and the process crashes,
			//  we will receive a duplicate LSN the next time we start replication. We can mitigate this by writing the last
			//  processed LSN to a file locally, but that suffers from the same problem (if the process dies before we update
			//  the file). A better solution would be to write the LSN directly into the DoltCommit message, and then parsing
			//  this message back out when we begin replication next.
			r.mu.Lock()
			if xld.ServerWALEnd > r.lsn {
				r.lsn = xld.ServerWALEnd
				err := sendStandbyStatusUpdate(r.lsn)
				if err != nil {
					r.mu.Unlock()
					return err
				}
			}
			r.mu.Unlock()
			
		default:
			log.Printf("Received unexpected message: %T\n", rawMsg)
		}
	}
}

func (r *LogicalReplicator) shutdown() {
	log.Print("shutting down replicator")
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
func (r *LogicalReplicator) beginReplication(slotName string) (*pgconn.PgConn, error) {
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
	// TODO: pass this in instead
	err = pglogrepl.StartReplication(context.Background(), conn, slotName, pglogrepl.LSN(0), pglogrepl.StartReplicationOptions{PluginArgs: pluginArguments})
	if err != nil {
		return nil, err
	}
	log.Println("Logical replication started on slot", slotName)

	return conn, nil
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
func (r *LogicalReplicator) processMessage(
	walData []byte,
	relations map[uint32]*pglogrepl.RelationMessageV2,
	typeMap *pgtype.Map,
	inStream *bool,
) error {
	logicalMsg, err := pglogrepl.ParseV2(walData, *inStream)
	if err != nil {
		log.Fatalf("Parse logical replication message: %s", err)
	}
	log.Printf("Receive a logical replication message: %s", logicalMsg.Type())
	switch logicalMsg := logicalMsg.(type) {
	case *pglogrepl.RelationMessageV2:
		relations[logicalMsg.RelationID] = logicalMsg
	case *pglogrepl.BeginMessage:
		// Indicates the beginning of a group of changes in a transaction.
		// This is only sent for committed transactions. You won't get any events from rolled back transactions.
		log.Printf("BeginMessage: %d", logicalMsg.Xid)
	case *pglogrepl.CommitMessage:
		log.Printf("CommitMessage: %v", logicalMsg.CommitTime)
	case *pglogrepl.InsertMessageV2:
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
					return err
				}
				valuesStr.WriteString(colData)
			default:
				log.Printf("unknown column data type: %c", col.DataType)
			}
		}

		log.Printf("insert for xid %d\n", logicalMsg.Xid)
		err = r.replicateQuery(fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s)", rel.Namespace, rel.RelationName, columnStr.String(), valuesStr.String()))
		if err != nil {
			return err
		}
	case *pglogrepl.UpdateMessageV2:
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
					return err
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

		log.Printf("update for xid %d\n", logicalMsg.Xid)
		err = r.replicateQuery(fmt.Sprintf("UPDATE %s.%s SET %s%s", rel.Namespace, rel.RelationName, updateStr.String(), whereClause(whereStr)))
		if err != nil {
			return err
		}
	case *pglogrepl.DeleteMessageV2:
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
					return err
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

		log.Printf("delete for xid %d\n", logicalMsg.Xid)
		err = r.replicateQuery(fmt.Sprintf("DELETE FROM %s.%s WHERE %s", rel.Namespace, rel.RelationName, whereStr.String()))
		if err != nil {
			return err
		}
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
	
	return nil
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
