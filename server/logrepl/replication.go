// Copyright 2023 Dolthub, Inc.
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
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
)

const outputPlugin = "pgoutput"

type rcvMsg struct {
  msg pgproto3.BackendMessage
	err error
}

type LogicalReplicator struct {
	primaryDns      string
	replicationConn *pgx.Conn
	receiveMsgChan chan rcvMsg
	stop chan struct{}
}

func NewLogicalReplicator(primaryDns string, replicationDns string) (*LogicalReplicator, error) {
	conn, err := pgx.Connect(context.Background(), replicationDns)
	if err != nil {
		return nil, err
	}
	
	return &LogicalReplicator{
		primaryDns: primaryDns,
		replicationConn: conn,
		stop: make(chan struct{}),
		receiveMsgChan: make(chan rcvMsg),
	}, nil
}

// SetupReplication sets up the replication slot and publication for the given database
func SetupReplication(connectionString string, publicationName string) error {
	conn, err := pgconn.Connect(context.Background(), connectionString)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	result := conn.Exec(context.Background(), fmt.Sprintf("DROP PUBLICATION IF EXISTS %s;", publicationName))
	_, err = result.ReadAll()
	if err != nil {
		return err
	}

	result = conn.Exec(context.Background(), fmt.Sprintf("CREATE PUBLICATION %s FOR ALL TABLES;", publicationName))
	_, err = result.ReadAll()
	return err
}

func (r *LogicalReplicator) StartReplication(slotName string) error {
	standbyMessageTimeout := time.Second * 10
	nextStandbyMessageDeadline := time.Now().Add(standbyMessageTimeout)
	relationsV2 := map[uint32]*pglogrepl.RelationMessageV2{}
	typeMap := pgtype.NewMap()

	// whenever we get StreamStartMessage we set inStream to true and then pass it to DecodeV2 function
	// on StreamStopMessage we set it back to false
	inStream := false

	connErrCnt := 0
	i := 0
	var primaryConn *pgconn.PgConn
	var clientXLogPos pglogrepl.LSN
	
	defer func() {
		if primaryConn != nil {
			_ = primaryConn.Close(context.Background())
		}
	}()
	
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
			primaryConn, clientXLogPos, err = r.beginReplication(slotName)
			if err != nil {
				return err
			}
		}
		
		if time.Now().After(nextStandbyMessageDeadline) {
			err := pglogrepl.SendStandbyStatusUpdate(context.Background(), primaryConn, pglogrepl.StandbyStatusUpdate{WALWritePosition: clientXLogPos})
			if err != nil {
				connErrCnt++
				if connErrCnt < 3 {
					// re-establish connection on next pass through the loop
					_ = primaryConn.Close(context.Background())
					primaryConn = nil
					continue
				}

				return err
			}
			
			connErrCnt = 0
			log.Printf("Sent Standby status message at %s\n", clientXLogPos.String())
			nextStandbyMessageDeadline = time.Now().Add(standbyMessageTimeout)
		}

		ctx, cancel := context.WithDeadline(context.Background(), nextStandbyMessageDeadline)
		go func() {
			rawMsg, err := primaryConn.ReceiveMessage(ctx)
			r.receiveMsgChan <- rcvMsg{msg: rawMsg, err: err}
		}()
		
		var msgAndErr rcvMsg
		select {
		case <-r.stop:
			cancel()
			r.shutdown()
			return nil
		case <-ctx.Done():
			continue
		case msgAndErr = <-r.receiveMsgChan:
		}

		if msgAndErr.err != nil {
			if pgconn.Timeout(msgAndErr.err) {
				continue
			} else {
				connErrCnt++
				if connErrCnt < 3 {
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
			if pkm.ServerWALEnd > clientXLogPos {
				clientXLogPos = pkm.ServerWALEnd
			}
			if pkm.ReplyRequested {
				nextStandbyMessageDeadline = time.Time{}
			}

		case pglogrepl.XLogDataByteID:
			xld, err := pglogrepl.ParseXLogData(msg.Data[1:])
			if err != nil {
				log.Fatalln("ParseXLogData failed:", err)
			}

			log.Printf("XLogData => WALStart %s ServerWALEnd %s ServerTime %s WALData:\n", xld.WALStart, xld.ServerWALEnd, xld.ServerTime)
			r.processMessage(xld.WALData, relationsV2, typeMap, &inStream)

			if xld.WALStart > clientXLogPos {
				clientXLogPos = xld.WALStart
			}
		default:
			// TODO: is this an error?
			log.Printf("Received unexpected message: %T\n", rawMsg)
		}

		i++
		if i%11 == 0 {
			// log.Printf("simulating connection failure\n")
			// _ = conn.Close(context.Background())
			// conn = nil
		}
	}
}

func (r *LogicalReplicator) shutdown() {
	log.Printf("shutting down replicator")
	close(r.stop)
}

func (r *LogicalReplicator) Stop() {
	log.Printf("stopping replication...")
	r.stop <- struct{}{}
	_, _ = <-r.stop
}

func (r *LogicalReplicator) replicateQuery(query string) error {
	log.Printf("replicating query: %s", query)
	_, err := r.replicationConn.Exec(context.Background(), query)
	return err
}

func (r *LogicalReplicator) beginReplication(slotName string) (*pgconn.PgConn, pglogrepl.LSN, error) {
	conn, err := pgconn.Connect(context.Background(), r.primaryDns)
	if err != nil {
		return nil, 0, err
	}

	// streaming of large transactions is available since PG 14 (protocol version 2)
	// we also need to set 'streaming' to 'true'
	pluginArguments := []string{
		"proto_version '2'",
		"publication_names 'pglogrepl_demo'",
		"messages 'true'",
		"streaming 'true'",
	}

	sysident, err := pglogrepl.IdentifySystem(context.Background(), conn)
	if err != nil {
		return nil, 0, err
	}
	log.Println("SystemID:", sysident.SystemID, "Timeline:", sysident.Timeline, "XLogPos:", sysident.XLogPos, "DBName:", sysident.DBName)

	_ = pglogrepl.DropReplicationSlot(context.Background(), conn, slotName, pglogrepl.DropReplicationSlotOptions{})
	_, err = pglogrepl.CreateReplicationSlot(context.Background(), conn, slotName, outputPlugin, pglogrepl.CreateReplicationSlotOptions{Temporary: true})
	if err != nil {
		pgErr, ok := err.(*pgconn.PgError)
		if ok && pgErr.Code == "42710" {
			// replication slot already exists, we can ignore this error
		} else {
			return nil, 0, err
		}
	}
	log.Println("Created temporary replication slot:", slotName)

	err = pglogrepl.StartReplication(context.Background(), conn, slotName, sysident.XLogPos, pglogrepl.StartReplicationOptions{PluginArgs: pluginArguments})
	if err != nil {
		return nil, 0, err
	}
	log.Println("Logical replication started on slot", slotName)
	
	return conn, sysident.XLogPos, nil
}

func (r *LogicalReplicator) processMessage(walData []byte, relations map[uint32]*pglogrepl.RelationMessageV2, typeMap *pgtype.Map, inStream *bool) {
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
					panic(err)
				}
				valuesStr.WriteString(fmt.Sprintf("%s", colData))
				// valuesStr.WriteString(fmt.Sprintf("'%s'", val))
			default:
				log.Printf("unknown column data type: %c", col.DataType)
			}
		}
		
		log.Printf("insert for xid %d\n", logicalMsg.Xid)
		err = r.replicateQuery(fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s)", rel.Namespace, rel.RelationName, columnStr.String(), valuesStr.String()))
		if err != nil {
			panic(err)
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
					panic(err)
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
			panic(err)
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
					panic(err)
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
			panic(err)
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
}

func whereClause(str strings.Builder) string {
	if str.Len() > 0 {
		return " WHERE " + str.String()
	}
	return ""
}

func decodeTextColumnData(mi *pgtype.Map, data []byte, dataType uint32) (interface{}, error) {
	if dt, ok := mi.TypeForOID(dataType); ok {
		return dt.Codec.DecodeValue(mi, dataType, pgtype.TextFormatCode, data)
	}
	return string(data), nil
}

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
			// TODO
			value = fmt.Sprintf("%v", data)
		}
	} else {
		value = fmt.Sprintf("%v", data)
	}

	// Some types need additional quoting after encoding	
	switch data.(type) {
	case string, time.Time, pgtype.Time, bool:
		return fmt.Sprintf("'%s'", value), nil
	default:
		return value, nil
	}
}