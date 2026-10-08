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
	"encoding/binary"
	"fmt"
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgproto3"
)

// TestResultBatchBoundaries exercises empty, partial, exact, and overflowing
// wire-result batches. In addition to row ordering, the nullable and text
// columns protect the lifetime of values retained until a batch is encoded.
func TestResultBatchBoundaries(t *testing.T) {
	const maxRows = 129
	values := make([]string, maxRows)
	for i := range values {
		id := i + 1
		textValue := fmt.Sprintf("'value-%03d'", id)
		if id%17 == 0 {
			textValue = "NULL"
		}
		values[i] = fmt.Sprintf("(%d, %s)", id, textValue)
	}

	assertions := make([]ScriptTestAssertion, 0, 5)
	for _, rowCount := range []int{0, 1, 127, 128, 129} {
		expected := make([]sql.Row, rowCount)
		for i := range expected {
			id := i + 1
			var textValue any = fmt.Sprintf("value-%03d", id)
			if id%17 == 0 {
				textValue = nil
			}
			expected[i] = sql.Row{int32(id), textValue}
		}
		assertions = append(assertions, ScriptTestAssertion{
			Query:    fmt.Sprintf("SELECT id, value FROM result_batch_test ORDER BY id LIMIT %d", rowCount),
			Expected: expected,
		})
	}

	RunScripts(t, []ScriptTest{{
		Name: "result batch boundaries",
		SetUpScript: []string{
			"CREATE TABLE result_batch_test (id INT PRIMARY KEY, value TEXT)",
			"INSERT INTO result_batch_test VALUES " + strings.Join(values, ","),
		},
		Assertions: assertions,
	}})
}

// TestResultBatchMixedFormats verifies that mixed text, binary, and NULL values
// remain correct when a result crosses the internal batch boundary.
func TestResultBatchMixedFormats(t *testing.T) {
	const rowCount = 129
	values := make([]string, rowCount)
	receive := make([]pgproto3.BackendMessage, 0, rowCount+4)
	receive = append(receive, &pgproto3.ParseComplete{}, &pgproto3.BindComplete{})
	for i := range values {
		id := i + 1
		textValue := fmt.Sprintf("'value-%03d'", id)
		var wireText []byte
		if id%17 == 0 || id == rowCount {
			textValue = "NULL"
		} else {
			wireText = []byte(fmt.Sprintf("value-%03d", id))
		}
		values[i] = fmt.Sprintf("(%d, %s)", id, textValue)
		wireID := make([]byte, 4)
		binary.BigEndian.PutUint32(wireID, uint32(id))
		receive = append(receive, &pgproto3.DataRow{Values: [][]byte{wireID, wireText}})
	}
	receive = append(receive,
		&pgproto3.CommandComplete{CommandTag: []byte("SELECT 129")},
		&pgproto3.ReadyForQuery{TxStatus: 'I'},
	)

	RunWireScripts(t, []WireScriptTest{{
		Name: "mixed formats across result batch boundary",
		SetUpScript: []string{
			"CREATE TABLE result_batch_mixed_test (id INT PRIMARY KEY, value TEXT)",
			"INSERT INTO result_batch_mixed_test VALUES " + strings.Join(values, ","),
		},
		Assertions: []WireScriptTestAssertion{{
			Send: []pgproto3.FrontendMessage{
				&pgproto3.Parse{Name: "result_batch_mixed", Query: "SELECT id, value FROM result_batch_mixed_test ORDER BY id"},
				&pgproto3.Bind{PreparedStatement: "result_batch_mixed", ResultFormatCodes: []int16{1, 0}},
				&pgproto3.Execute{},
				&pgproto3.Sync{},
			},
			Receive: receive,
		}},
	}})
}
