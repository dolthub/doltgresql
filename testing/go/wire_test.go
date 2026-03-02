// Copyright 2026 Dolthub, Inc.
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

package _go

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xdg-go/scram"
)

// TestWireTypes allows us to directly test what is received on the wire regarding types, ensuring that the wire
// protocol is correctly implemented. ANY changes made to ANY test must be validated against an external Postgres server
// using the `ExternalServerPort` field.
func TestWireTypes(t *testing.T) {
	RunWireScripts(t, []WireScriptTest{
		{
			Name: "Smoke Test",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT4 PRIMARY KEY);",
				"INSERT INTO test VALUES (7);",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Query{String: "SELECT * FROM test;"},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("pk"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          23,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.DataRow{Values: [][]byte{[]byte("7")}},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "BIT returning text format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 BIT(8), v2 BIT(3));",
				"INSERT INTO test VALUES (B'11011010', '101');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1560,
									DataTypeSize:         -1,
									TypeModifier:         8,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1560,
									DataTypeSize:         -1,
									TypeModifier:         3,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte("11011010"),
								[]byte("101"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "BIT returning binary format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 BIT(65), v2 BIT(3));",
				"INSERT INTO test VALUES (B'10101010001000110110110010110011000101010110101010010110101011001', '101');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1560,
									DataTypeSize:         -1,
									TypeModifier:         65,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1560,
									DataTypeSize:         -1,
									TypeModifier:         3,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								{0, 0, 0, 65, 170, 35, 108, 179, 21, 106, 150, 172, 128},
								{0, 0, 0, 3, 160},
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "BIT VARYING returning text format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 BIT VARYING, v2 BIT VARYING(5));",
				"INSERT INTO test VALUES (B'100101010110011001', '110');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1562,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1562,
									DataTypeSize:         -1,
									TypeModifier:         5,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte("100101010110011001"),
								[]byte("110"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "BIT VARYING returning binary format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 BIT VARYING, v2 BIT VARYING(5));",
				"INSERT INTO test VALUES (B'100101010110011001', '110');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1562,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1562,
									DataTypeSize:         -1,
									TypeModifier:         5,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								{0, 0, 0, 18, 149, 102, 64},
								{0, 0, 0, 3, 192},
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "BOOL returning text format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 BOOL, v2 BOOL);",
				"INSERT INTO test VALUES (true, false);",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          16,
									DataTypeSize:         1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          16,
									DataTypeSize:         1,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte("t"),
								[]byte("f"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "BOOL returning binary format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 BOOL, v2 BOOL);",
				"INSERT INTO test VALUES (true, false);",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          16,
									DataTypeSize:         1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          16,
									DataTypeSize:         1,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								{1},
								{0},
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "BPCHAR returning text format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 BPCHAR, v2 BPCHAR(7));",
				"INSERT INTO test VALUES ('', 'abc'), ('more text', 'text');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test ORDER BY v1;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1042,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1042,
									DataTypeSize:         -1,
									TypeModifier:         11,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								{},
								[]byte("abc    "),
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte("more text"),
								[]byte("text   "),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "BPCHAR returning binary format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 BPCHAR, v2 BPCHAR(7));",
				"INSERT INTO test VALUES ('', 'abc'), ('more text', 'text');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test ORDER BY v1;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1042,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1042,
									DataTypeSize:         -1,
									TypeModifier:         11,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								{},
								[]byte("abc    "),
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte("more text"),
								[]byte("text   "),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: `"char" returning text format`,
			SetUpScript: []string{
				`CREATE TABLE test (v1 "char", v2 "char");`,
				`INSERT INTO test VALUES ('123', 'v');`,
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          18,
									DataTypeSize:         1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          18,
									DataTypeSize:         1,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte("1"),
								[]byte("v"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: `"char" returning binary format`,
			SetUpScript: []string{
				`CREATE TABLE test (v1 "char", v2 "char");`,
				`INSERT INTO test VALUES ('123', 'v');`,
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          18,
									DataTypeSize:         1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          18,
									DataTypeSize:         1,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								{'1'},
								{'v'},
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "ENUM returning text format",
			SetUpScript: []string{
				"CREATE TYPE enumType AS ENUM ('eval1', 'eval2', 'eval3');",
				"CREATE TABLE test (v1 enumType, v2 enumType);",
				"INSERT INTO test VALUES ('eval1', 'eval3');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          0,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          0,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte("eval1"),
								[]byte("eval3"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "ENUM returning binary format",
			SetUpScript: []string{
				"CREATE TYPE enumType AS ENUM ('eval1', 'eval2', 'eval3');",
				"CREATE TABLE test (v1 enumType, v2 enumType);",
				"INSERT INTO test VALUES ('eval1', 'eval3');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          0,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          0,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte("eval1"),
								[]byte("eval3"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "INTERVAL returning text format",
			Skip: true, // TODO: need to fix our text output for intervals
			SetUpScript: []string{
				"CREATE TABLE test (v1 INTERVAL, v2 INTERVAL);",
				"INSERT INTO test VALUES ('@ 1 minute', '2 years 15 months 100 weeks 99 hours 123456789 milliseconds');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1186,
									DataTypeSize:         16,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1186,
									DataTypeSize:         16,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte("@ 1 min"),
								[]byte("@ 3 years 3 mons 700 days 133 hours 17 mins 36.789 secs"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "INTERVAL returning binary format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 INTERVAL, v2 INTERVAL);",
				"INSERT INTO test VALUES ('@ 1 minute', '2 years 15 months 100 weeks 99 hours 123456789 milliseconds');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1186,
									DataTypeSize:         16,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1186,
									DataTypeSize:         16,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								{0, 0, 0, 0, 3, 147, 135, 0, 0, 0, 0, 0, 0, 0, 0, 0},
								{0, 0, 0, 111, 185, 177, 134, 8, 0, 0, 2, 188, 0, 0, 0, 39},
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "JSON returning text format",
			SetUpScript: []string{
				`CREATE TABLE test (v1 JSON, v2 JSON, v3 INT4);`,
				`INSERT INTO test VALUES ('{"key1": {"key": "value"}}', '{}', 1), ('{"key1": {"key": [2, 3]}}', '[]', 2);`,
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT v1, v2 FROM test ORDER BY v3;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          114,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          114,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(`{"key1": {"key": "value"}}`),
								[]byte(`{}`),
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(`{"key1": {"key": [2, 3]}}`),
								[]byte(`[]`),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "JSON returning binary format",
			SetUpScript: []string{
				`CREATE TABLE test (v1 JSON, v2 JSON, v3 INT4);`,
				`INSERT INTO test VALUES ('{"key1": {"key": "value"}}', '{}', 1), ('{"key1": {"key": [2, 3]}}', '[]', 2);`,
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT v1, v2 FROM test ORDER BY v3;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          114,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          114,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(`{"key1": {"key": "value"}}`),
								[]byte(`{}`),
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(`{"key1": {"key": [2, 3]}}`),
								[]byte(`[]`),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "JSONB returning text format",
			SetUpScript: []string{
				`CREATE TABLE test (v1 JSONB, v2 JSONB, v3 INT4);`,
				`INSERT INTO test VALUES ('{"key1": {"key": "value"}}', '{}', 1), ('{"key1": {"key": [2, 3]}}', '[]', 2);`,
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT v1, v2 FROM test ORDER BY v3;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          3802,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          3802,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(`{"key1": {"key": "value"}}`),
								[]byte(`{}`),
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(`{"key1": {"key": [2, 3]}}`),
								[]byte(`[]`),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "JSONB returning binary format",
			SetUpScript: []string{
				`CREATE TABLE test (v1 JSONB, v2 JSONB, v3 INT4);`,
				`INSERT INTO test VALUES ('{"key1": {"key": "value"}}', '{}', 1), ('{"key1": {"key": [2, 3]}}', '[]', 2);`,
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT v1, v2 FROM test ORDER BY v3;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          3802,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          3802,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								append([]byte{1}, []byte(`{"key1": {"key": "value"}}`)...),
								{1, 123, 125},
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								append([]byte{1}, []byte(`{"key1": {"key": [2, 3]}}`)...),
								{1, 91, 93},
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "NAME returning text format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 NAME, v2 NAME);",
				"INSERT INTO test VALUES ('', 'abc'), (NULL, 'a\",c');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          19,
									DataTypeSize:         64,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          19,
									DataTypeSize:         64,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								nil,
								[]byte(`a",c`),
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(""),
								[]byte("abc"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "NAME returning binary format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 NAME, v2 NAME);",
				"INSERT INTO test VALUES ('', 'abc'), (NULL, 'a\",c');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          19,
									DataTypeSize:         64,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          19,
									DataTypeSize:         64,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								nil,
								[]byte(`a",c`),
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(""),
								[]byte("abc"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "NUMERIC returning text format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 NUMERIC, v2 NUMERIC(5,2), v3 NUMERIC(14,5));",
				"INSERT INTO test VALUES (0, -0.1, NULL), (12357232.456786653224768755799, 235.67, 4278.009);",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test ORDER BY v1;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1700,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1700,
									DataTypeSize:         -1,
									TypeModifier:         327686,
									Format:               0,
								},
								{
									Name:                 []byte("v3"),
									TableOID:             0,
									TableAttributeNumber: 3,
									DataTypeOID:          1700,
									DataTypeSize:         -1,
									TypeModifier:         917513,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(`0`),
								[]byte(`-0.10`),
								nil,
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(`12357232.456786653224768755799`),
								[]byte(`235.67`),
								[]byte(`4278.00900`),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "NUMERIC returning binary format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 NUMERIC, v2 NUMERIC(5,2), v3 NUMERIC(14,5));",
				"INSERT INTO test VALUES (0, -0.1, NULL), (12357232.456786653224768755799, 235.67, 4278.009);",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test ORDER BY v1;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1700,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1700,
									DataTypeSize:         -1,
									TypeModifier:         327686,
									Format:               0,
								},
								{
									Name:                 []byte("v3"),
									TableOID:             0,
									TableAttributeNumber: 3,
									DataTypeOID:          1700,
									DataTypeSize:         -1,
									TypeModifier:         917513,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								{0, 0, 0, 0, 0, 0, 0, 0},
								{0, 1, 255, 255, 64, 0, 0, 2, 3, 232},
								nil,
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								{0, 8, 0, 1, 0, 0, 0, 21, 4, 211, 28, 64, 17, 215, 33, 217, 12, 152, 30, 7, 21, 203, 35, 40},
								{0, 2, 0, 0, 0, 0, 0, 2, 0, 235, 26, 44},
								{0, 2, 0, 0, 0, 0, 0, 5, 16, 182, 0, 90},
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "OID returning text format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 OID, v2 OID);",
				"INSERT INTO test VALUES (1, 2483574913);",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          26,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          26,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte("1"),
								[]byte("2483574913"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "OID returning binary format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 OID, v2 OID);",
				"INSERT INTO test VALUES (1, 2483574913);",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          26,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          26,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								{0, 0, 0, 1},
								{148, 8, 88, 129},
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "RECORD returning text format",
			SetUpScript: []string{
				"CREATE TABLE pre1 (v1 TEXT, v2 INT8, v3 NUMERIC(6,1));",
				"CREATE TABLE pre2 (v1 VARCHAR, v2 OID, v3 BOOL);",
				"CREATE TABLE test (v1 pre1, v2 pre2);",
				"INSERT INTO test VALUES (ROW('abc'::TEXT, 1::INT8, '12345.6'::NUMERIC(6,1)), ROW('def'::VARCHAR, 2::OID, 't'::BOOL));",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          0, // User-defined OID will always differ
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          0, // User-defined OID will always differ
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(`(abc,1,12345.6)`),
								[]byte(`(def,2,t)`),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "RECORD returning binary format",
			SetUpScript: []string{
				"CREATE TABLE pre1 (v1 TEXT, v2 INT8, v3 NUMERIC(6,1));",
				"CREATE TABLE pre2 (v1 VARCHAR, v2 OID, v3 BOOL);",
				"CREATE TABLE test (v1 pre1, v2 pre2);",
				"INSERT INTO test VALUES (ROW('abc'::TEXT, 1::INT8, '12345.6'::NUMERIC(6,1)), ROW('def'::VARCHAR, 2::OID, 't'::BOOL));",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          0, // User-defined OID will always differ
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          0, // User-defined OID will always differ
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								{0, 0, 0, 3, 0, 0, 0, 25, 0, 0, 0, 3, 97, 98, 99, 0, 0, 0, 20, 0, 0, 0, 8, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 6, 164, 0, 0, 0, 14, 0, 3, 0, 1, 0, 0, 0, 1, 0, 1, 9, 41, 23, 112},
								{0, 0, 0, 3, 0, 0, 4, 19, 0, 0, 0, 3, 100, 101, 102, 0, 0, 0, 26, 0, 0, 0, 4, 0, 0, 0, 2, 0, 0, 0, 16, 0, 0, 0, 1, 1},
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "REGTYPE returning text format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 REGTYPE, v2 REGTYPE);",
				"INSERT INTO test VALUES ('numeric'::REGTYPE, 'text'::REGTYPE);",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          2206,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          2206,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte("numeric"),
								[]byte("text"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "REGTYPE returning binary format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 REGTYPE, v2 REGTYPE);",
				"INSERT INTO test VALUES ('numeric'::REGTYPE, 'text'::REGTYPE);",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          2206,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          2206,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								{0, 0, 6, 164},
								{0, 0, 0, 25},
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "TEXT returning text format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 TEXT, v2 TEXT);",
				"INSERT INTO test VALUES ('', 'abc'), (NULL, 'a\",c');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          25,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          25,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								nil,
								[]byte(`a",c`),
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(""),
								[]byte("abc"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "TEXT returning binary format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 TEXT, v2 TEXT);",
				"INSERT INTO test VALUES ('', 'abc'), (NULL, 'a\",c');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          25,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          25,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								nil,
								[]byte(`a",c`),
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(""),
								[]byte("abc"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "TEXT ARRAY returning text format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 TEXT[], v2 TEXT[]);",
				"INSERT INTO test VALUES (ARRAY[]::text[], ARRAY['a','bb','ccc']), (NULL, ARRAY['dd',NULL,'ee']);",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1009,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1009,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte("{}"),
								[]byte("{a,bb,ccc}"),
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								nil,
								[]byte("{dd,NULL,ee}"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "TEXT ARRAY returning binary format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 TEXT[], v2 TEXT[]);",
				"INSERT INTO test VALUES (ARRAY[]::text[], ARRAY['a','bb','ccc']), (NULL, ARRAY['dd',NULL,'ee']);",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1009,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1009,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 25},
								{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 25, 0, 0, 0, 3, 0, 0, 0, 1, 0, 0, 0, 1, 97, 0, 0, 0, 2, 98, 98, 0, 0, 0, 3, 99, 99, 99},
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								nil,
								{0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 25, 0, 0, 0, 3, 0, 0, 0, 1, 0, 0, 0, 2, 100, 100, 255, 255, 255, 255, 0, 0, 0, 2, 101, 101},
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "TIME returning text format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 TIME, v2 TIME);",
				"INSERT INTO test VALUES ('0:0', '04:05:06.789'), ('09:27 PM', '12:12');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1083,
									DataTypeSize:         8,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1083,
									DataTypeSize:         8,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(`00:00:00`),
								[]byte(`04:05:06.789`),
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(`21:27:00`),
								[]byte(`12:12:00`),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "TIME returning binary format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 TIME, v2 TIME);",
				"INSERT INTO test VALUES ('0:0', '04:05:06.789'), ('09:27 PM', '12:12');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1083,
									DataTypeSize:         8,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1083,
									DataTypeSize:         8,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								{0, 0, 0, 0, 0, 0, 0, 0},
								{0, 0, 0, 3, 108, 151, 202, 136},
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								{0, 0, 0, 17, 250, 171, 177, 0},
								{0, 0, 0, 10, 57, 214, 4, 0},
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "TIMETZ returning text format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 TIMETZ, v2 TIMETZ);",
				"INSERT INTO test VALUES ('0:0 PST', '04:05:06.789 MST'), ('09:27 PM CST', '12:12 EST');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1266,
									DataTypeSize:         12,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1266,
									DataTypeSize:         12,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(`00:00:00-08`),
								[]byte(`04:05:06.789-07`),
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(`21:27:00-06`),
								[]byte(`12:12:00-05`),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "TIMETZ returning binary format",
			Skip: true, // TODO: CI uses a different time zone, need to homogenize them somehow for testing
			SetUpScript: []string{
				"CREATE TABLE test (v1 TIMETZ, v2 TIMETZ);",
				"INSERT INTO test VALUES ('0:0', '04:05:06.789'), ('09:27 PM', '12:12');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1266,
									DataTypeSize:         12,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1266,
									DataTypeSize:         12,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 112, 128},
								{0, 0, 0, 3, 108, 151, 202, 136, 0, 0, 112, 128},
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								{0, 0, 0, 17, 250, 171, 177, 0, 0, 0, 112, 128},
								{0, 0, 0, 10, 57, 214, 4, 0, 0, 0, 112, 128},
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "VARCHAR returning text format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 VARCHAR, v2 VARCHAR(5));",
				"INSERT INTO test VALUES ('', 'abc'), (NULL, 'a\",c');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1043,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1043,
									DataTypeSize:         -1,
									TypeModifier:         9,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								nil,
								[]byte(`a",c`),
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(""),
								[]byte("abc"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "VARCHAR returning binary format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 VARCHAR, v2 VARCHAR(5));",
				"INSERT INTO test VALUES ('', 'abc'), (NULL, 'a\",c');",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          1043,
									DataTypeSize:         -1,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          1043,
									DataTypeSize:         -1,
									TypeModifier:         9,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								nil,
								[]byte(`a",c`),
							},
						},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte(""),
								[]byte("abc"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 2")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "XID returning text format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 XID, v2 XID);",
				"INSERT INTO test VALUES (1::TEXT::XID, 2483574913::TEXT::XID);",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          28,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          28,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{0},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								[]byte("1"),
								[]byte("2483574913"),
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
		{
			Name: "XID returning binary format",
			SetUpScript: []string{
				"CREATE TABLE test (v1 XID, v2 XID);",
				"INSERT INTO test VALUES (1::TEXT::XID, 2483574913::TEXT::XID);",
			},
			Assertions: []WireScriptTestAssertion{
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Parse{
							Name:  "stmt_name",
							Query: "SELECT * FROM test;",
						},
						&pgproto3.Describe{
							ObjectType: 'S',
							Name:       "stmt_name",
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.ParseComplete{},
						&pgproto3.ParameterDescription{ParameterOIDs: []uint32{}},
						&pgproto3.RowDescription{
							Fields: []pgproto3.FieldDescription{
								{
									Name:                 []byte("v1"),
									TableOID:             0,
									TableAttributeNumber: 1,
									DataTypeOID:          28,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
								{
									Name:                 []byte("v2"),
									TableOID:             0,
									TableAttributeNumber: 2,
									DataTypeOID:          28,
									DataTypeSize:         4,
									TypeModifier:         -1,
									Format:               0,
								},
							},
						},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
				{
					Send: []pgproto3.FrontendMessage{
						&pgproto3.Bind{
							DestinationPortal:    "",
							PreparedStatement:    "stmt_name",
							ParameterFormatCodes: nil,
							Parameters:           nil,
							ResultFormatCodes:    []int16{1},
						},
						&pgproto3.Execute{},
						&pgproto3.Close{
							ObjectType: 'P',
						},
						&pgproto3.Sync{},
					},
					Receive: []pgproto3.BackendMessage{
						&pgproto3.BindComplete{},
						&pgproto3.DataRow{
							Values: [][]byte{
								{0, 0, 0, 1},
								{148, 8, 88, 129},
							},
						},
						&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
						&pgproto3.CloseComplete{},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					},
				},
			},
		},
	})
}

// IgnoreMessageParameters is used to ignore certain fields within a backend message, as they may not yet be implemented
// and therefore will return incorrect results (or variable results, such as with non-stable OIDs).
func IgnoreMessageParameters(message pgproto3.BackendMessage) pgproto3.BackendMessage {
	switch message := message.(type) {
	case *pgproto3.RowDescription:
		for i := range message.Fields {
			message.Fields[i].TableOID = 0
			// User-defined types will have an OID outside the reserved range, so we set those to zero
			if message.Fields[i].DataTypeOID > 16383 {
				message.Fields[i].DataTypeOID = 0
			}
		}
		return message
	default:
		return message
	}
}

// WireScriptTest is used to test wire messages, ensuring that our wire protocol behaves as expected.
type WireScriptTest struct {
	// Name of the script.
	Name string
	// The database to create and use. If not provided, then it defaults to "postgres".
	Database string
	// The SQL statements to execute as setup, in order. Results are not checked, but statements must not error.
	SetUpScript []string
	// The set of assertions to make after setup, in order
	Assertions []WireScriptTestAssertion
	// When using RunScripts, setting this on one (or more) tests causes RunWireScripts to ignore all tests that have
	// this set to false (which is the default value). This allows a developer to easily "focus" on a specific test
	// without having to comment out other tests, pull it into a different function, etc. In addition, CI ensures that
	// this is false before passing, meaning this prevents the commented-out situation where the developer forgets to
	// uncomment their code.
	Focus bool
	// Skip is used to completely skip a test
	Skip bool
	// When non-zero, this specific test will be run on an external server on the port specified. This is primarily to
	// test against an actual Postgres server to ensure that results are correct. CI ensures that this is not set before
	// passing, since this is intended for use on a developer's local machine only.
	ExternalServerPort int
}

// WireScriptTestAssertion are the assertions upon which the script executes its main "testing" logic.
type WireScriptTestAssertion struct {
	// These are sent as a batch to the server
	Send []pgproto3.FrontendMessage
	// These are the expected results that are received from the server, and must match in both contents and order
	Receive []pgproto3.BackendMessage
	// This functions the same as Focus on WireScriptTest, except that it applies to the assertion
	Focus bool
	// This is used to skip an assertion
	Skip bool
}

// RawWireConnection is a connection that allows us to directly send and receive messages to a server.
type RawWireConnection struct {
	frontend   *pgproto3.Frontend
	connection net.Conn
	username   string
	password   string
	network    string
	timeout    time.Duration
	startup    *pgproto3.StartupMessage
	errChan    chan error
}

// NewRawWireConnection returns a new RawWireConnection.
func NewRawWireConnection(t *testing.T, host string, port int, username string, password string, timeout time.Duration) *RawWireConnection {
	network := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	connection, err := (&net.Dialer{}).Dial("tcp", network)
	require.NoError(t, err)
	if len(username) == 0 && len(password) == 0 {
		username = "postgres"
		password = "password"
	}
	c := &RawWireConnection{
		frontend:   pgproto3.NewFrontend(connection, connection),
		connection: connection,
		username:   username,
		password:   password,
		network:    network,
		timeout:    timeout,
		startup:    nil,
		errChan:    make(chan error),
	}
	c.init(t)
	return c
}

// Close closes the internal connection.
func (c *RawWireConnection) Close() {
	// We don't close the error channel at all, which is fine since this is only used for these specific tests. Since we
	// send to the channel from another thread, it's possible to receive a ping between the time that the connection is
	// closed and the channel is closed. This would then cause a panic by trying to send on a closed channel. This could
	// be architected around, but it's far easier to just never close it for tests.
	_ = c.connection.Close()
}

// EmptyReceiveBuffer empties the buffer used by Receive. Returns an error if the buffer was not empty.
func (c *RawWireConnection) EmptyReceiveBuffer() error {
	if c.frontend.ReadBufferLen() > 0 {
		for c.frontend.ReadBufferLen() > 0 {
			_, _ = c.frontend.Receive()
		}
		return errors.New("Doltgres sent additional messages after ReadyForQuery")
	}
	return nil
}

// Receive returns the next message from the backend.
func (c *RawWireConnection) Receive(t *testing.T) (pgproto3.BackendMessage, error) {
	var message pgproto3.BackendMessage
	go func() {
		var err error
		message, err = c.frontend.Receive()
		c.errChan <- err
	}()
	return message, c.handleErrorChannel(t, false)
}

// Send sends the given messages over the wire. If an error is returned, then the connection has been closed, and a new
// one should be opened.
func (c *RawWireConnection) Send(t *testing.T, messages ...pgproto3.FrontendMessage) error {
	if len(messages) == 0 {
		return nil
	}
	hasMessage := false
	for _, message := range messages {
		if message == nil {
			continue
		}
		hasMessage = true
		if startupMessage, ok := message.(*pgproto3.StartupMessage); ok {
			c.startup = startupMessage
		}
		c.frontend.Send(message)
	}
	if !hasMessage {
		return nil
	}
	go func() {
		c.errChan <- c.frontend.Flush()
	}()
	return c.handleErrorChannel(t, true)
}

// init handles the startup message, authentication, and initial messages from the server.
func (c *RawWireConnection) init(t *testing.T) {
	err := c.Send(t, &pgproto3.StartupMessage{
		ProtocolVersion: 196608,
		Parameters: map[string]string{
			"timezone":         "PST8PDT",
			"user":             "postgres",
			"database":         "postgres",
			"options":          " -c intervalstyle=postgres_verbose",
			"application_name": "pg_regress",
			"client_encoding":  "WIN1252",
			"datestyle":        "Postgres, MDY",
		},
	})
	require.NoError(t, err)
	postgresMessage, err := c.Receive(t)
	require.NoError(t, err)
	// First we'll check if authentication is disabled
	if _, ok := postgresMessage.(*pgproto3.AuthenticationOk); ok {
		for {
			postgresMessage, err = c.Receive(t)
			require.NoError(t, err)
			switch response := postgresMessage.(type) {
			case *pgproto3.BackendKeyData:
			case *pgproto3.ErrorResponse:
				t.Log(response.Message)
				t.FailNow()
			case *pgproto3.ParameterStatus:
			case *pgproto3.ReadyForQuery:
				return
			default:
				t.Logf("unknown StartupMessage response type: %T", response)
				t.FailNow()
			}
		}
	}
	// If authentication is not disabled, then we'll do the SASL authentication routine (only one we support for now)
	if saslInit, ok := postgresMessage.(*pgproto3.AuthenticationSASL); !ok {
		t.Logf("unknown StartupMessage response type: %T", postgresMessage)
		t.FailNow()
	} else {
		found := false
		for _, authMech := range saslInit.AuthMechanisms {
			switch authMech {
			case "SCRAM-SHA-256":
				found = true
			}
		}
		if !found {
			t.Logf("attempted SASL authentication but did not find a suitable mechanism: %v", saslInit.AuthMechanisms)
			t.FailNow()
		}
	}
	client, err := scram.SHA256.NewClient(c.username, c.password, "")
	require.NoError(t, err)
	conv := client.NewConversation()
	saslInitialResponse, err := conv.Step("")
	require.NoError(t, err)
	err = c.Send(t, &pgproto3.SASLInitialResponse{
		AuthMechanism: "SCRAM-SHA-256",
		Data:          []byte(saslInitialResponse),
	})
	require.NoError(t, err)
	postgresMessage, err = c.Receive(t)
	require.NoError(t, err)
	saslContinue, ok := postgresMessage.(*pgproto3.AuthenticationSASLContinue)
	if !ok {
		t.Logf(`received "%T" but expected "AuthenticationSASLContinue"`, postgresMessage)
		t.FailNow()
	}
	saslResponse, err := conv.Step(string(saslContinue.Data))
	require.NoError(t, err)
	err = c.Send(t, &pgproto3.SASLResponse{
		Data: []byte(saslResponse),
	})
	require.NoError(t, err)
	postgresMessage, err = c.Receive(t)
	require.NoError(t, err)
	saslFinal, ok := postgresMessage.(*pgproto3.AuthenticationSASLFinal)
	if !ok {
		t.Logf(`received "%T" but expected "AuthenticationSASLFinal"`, postgresMessage)
		t.FailNow()
	}
	_, err = conv.Step(string(saslFinal.Data))
	require.NoError(t, err)
	postgresMessage, err = c.Receive(t)
	require.NoError(t, err)
	_, ok = postgresMessage.(*pgproto3.AuthenticationOk)
	if !ok {
		t.Logf(`received "%T" but expected "AuthenticationOk"`, postgresMessage)
		t.FailNow()
	}
	for {
		postgresMessage, err = c.Receive(t)
		require.NoError(t, err)
		switch response := postgresMessage.(type) {
		case *pgproto3.BackendKeyData:
		case *pgproto3.ErrorResponse:
			t.Log(response.Message)
			t.FailNow()
		case *pgproto3.ParameterStatus:
		case *pgproto3.ReadyForQuery:
			return
		default:
			t.Logf("unknown StartupMessage response type: %T", response)
			t.FailNow()
		}
	}
}

// handleErrorChannel handles errors while ensuring that stuck queries do not cause an infinite loop via a timeout.
func (c *RawWireConnection) handleErrorChannel(t *testing.T, isSend bool) error {
	var err error
	select {
	case err = <-c.errChan:
	case <-time.After(c.timeout):
		if isSend {
			err = errors.New("timeout during Send")
		} else {
			err = errors.New("timeout during Receive")
		}
	}
	// On error, we must create a new connection since we cut the old one
	if err != nil {
		_ = c.connection.Close()
		connection, nErr := (&net.Dialer{}).Dial("tcp", c.network)
		if nErr != nil {
			panic(fmt.Errorf("Unable to create a new connection:\n%s\n\nOriginal error:\n%s", nErr.Error(), err.Error()))
		}
		c.connection = connection
		c.frontend = pgproto3.NewFrontend(connection, connection)
		c.init(t)
	}
	return err
}

// RunWireScripts runs the given collection of scripts.
func RunWireScripts(t *testing.T, scripts []WireScriptTest) {
	// First, we'll run through the scripts to check for the Focus variable. If it's true, then append it to the new slice.
	focusScripts := make([]WireScriptTest, 0, len(scripts))
	for _, script := range scripts {
		if script.Focus {
			// If this is running in GitHub Actions, then we'll panic, because someone forgot to disable it before committing
			if _, ok := os.LookupEnv("GITHUB_ACTION"); ok {
				panic(fmt.Sprintf("The wire script `%s` has Focus set to `true`. GitHub Actions requires that "+
					"all tests are run, which Focus circumvents, leading to this error. Please disable Focus on "+
					"all tests.", script.Name))
			}
			focusScripts = append(focusScripts, script)
		}
		if script.ExternalServerPort != 0 {
			// Same as with Focus, we want to panic in a GitHub Action since this is only for local debugging
			if _, ok := os.LookupEnv("GITHUB_ACTION"); ok {
				panic(fmt.Sprintf("The wire script `%s` has ExternalServerPort set to `%d`. GitHub Actions "+
					"requires that all tests are run against an in-memory Doltgres server, which ExternalServerPort "+
					"circumvents, leading to this error. Please remove ExternalServerPort on all tests.",
					script.Name, script.ExternalServerPort))
			}
		}
	}
	// If we have scripts with Focus set, then we replace the normal script slice with the new slice.
	if len(focusScripts) > 0 {
		scripts = focusScripts
	}

	for _, script := range scripts {
		t.Run(script.Name, func(t *testing.T) {
			if script.Skip {
				t.Skip()
			}

			var rawConn *RawWireConnection
			if script.ExternalServerPort == 0 {
				scriptDatabase := script.Database
				if len(scriptDatabase) == 0 {
					scriptDatabase = "postgres"
				}
				port, err := sql.GetEmptyPort()
				require.NoError(t, err)
				ctx, conn, controller := CreateServerWithPort(t, scriptDatabase, port)
				defer func() {
					controller.Stop()
					err := controller.WaitForStop()
					require.NoError(t, err)
				}()
				for _, query := range script.SetUpScript {
					_, err = conn.Exec(ctx, query)
					require.NoError(t, err, "error running setup query: %s", query)
				}
				conn.Close(ctx)
				rawConn = NewRawWireConnection(t, "localhost", port, "", "", 10*time.Second)
				defer rawConn.Close()
			} else {
				rawConn = NewRawWireConnection(t, "localhost", script.ExternalServerPort, "", "", 10*time.Second)
				defer rawConn.Close()
				// Some tests create tables in their setup, so we do a very basic check to first drop those tables on
				// the external server. While not rigorous, this at least lets us run (and rerun) most tests against an
				// external server without explicitly rewriting the setup scripts. We also do this for created types.
				var dropTables []string
				var dropTypes []string
				for _, query := range script.SetUpScript {
					if idx := strings.Index(strings.ToLower(query), "create table "); idx != -1 {
						dropTables = append(dropTables, query[idx+13:strings.Index(query, "(")])
					}
					if idx := strings.Index(strings.ToLower(query), "create type "); idx != -1 {
						dropTypes = append(dropTypes, query[idx+12:idx+12+strings.Index(query[idx+12:], " ")])
					}
				}
				if len(dropTypes) > 0 {
					script.SetUpScript = append([]string{fmt.Sprintf(`DROP TYPE IF EXISTS %s;`, strings.Join(dropTypes, ", "))}, script.SetUpScript...)
				}
				if len(dropTables) > 0 {
					script.SetUpScript = append([]string{fmt.Sprintf(`DROP TABLE IF EXISTS %s;`, strings.Join(dropTables, ", "))}, script.SetUpScript...)
				}
				for _, query := range script.SetUpScript {
					err := rawConn.Send(t, &pgproto3.Query{String: query})
					require.NoError(t, err)
				ExternalSetupLoop:
					for {
						postgresMessage, err := rawConn.Receive(t)
						require.NoError(t, err)
						switch response := postgresMessage.(type) {
						case *pgproto3.ErrorResponse:
							t.Log(response.Message)
							t.FailNow()
						case *pgproto3.ReadyForQuery:
							break ExternalSetupLoop
						}
					}
				}
			}

			// With everything set up, we now check for Focus on the assertions
			assertions := script.Assertions
			// First, we'll run through the scripts to check for the Focus variable. If it's true, then append it to the new slice.
			focusAssertions := make([]WireScriptTestAssertion, 0, len(script.Assertions))
			for _, assertion := range script.Assertions {
				if assertion.Focus {
					// If this is running in GitHub Actions, then we'll panic, because someone forgot to disable it before committing
					if _, ok := os.LookupEnv("GITHUB_ACTION"); ok {
						panic("A wire assertion has Focus set to `true`. GitHub Actions requires that " +
							"all non-skipped assertions are run, which Focus circumvents, leading to this error. " +
							"Please disable Focus on all wire assertions.")
					}
					focusAssertions = append(focusAssertions, assertion)
				}
			}
			// If we have assertions with Focus set, then we replace the original slice with the new slice.
			if len(focusAssertions) > 0 {
				assertions = focusAssertions
			}

			// Run the assertions
			for assertionIdx, assertion := range assertions {
				t.Run(fmt.Sprintf("%d", assertionIdx), func(t *testing.T) {
					if assertion.Skip {
						t.Skip("Skip has been set in the assertion")
					}
					err := rawConn.Send(t, assertion.Send...)
					require.NoError(t, err)
					for _, expectedMessage := range assertion.Receive {
						actualMessage, err := rawConn.Receive(t)
						require.NoError(t, err)
						if !assert.Equal(t, IgnoreMessageParameters(expectedMessage), IgnoreMessageParameters(actualMessage)) {
							// If the assertion fails, and it's a message that we don't expect, then we sync to the next
							// ReadyForQuery message
							if reflect.TypeOf(expectedMessage) != reflect.TypeOf(actualMessage) {
								if _, ok := actualMessage.(*pgproto3.ReadyForQuery); !ok {
									for {
										actualMessage, err := rawConn.Receive(t)
										require.NoError(t, err)
										if _, ok = actualMessage.(*pgproto3.ReadyForQuery); ok {
											return
										}
									}
								}
							}
						}
					}
					// We then ensure that there are no other messages that were not accounted for by the assertion
					// (which we consider an error)
					_ = assert.NoError(t, rawConn.EmptyReceiveBuffer())
				})
			}
		})
	}
}
