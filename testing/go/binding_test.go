// Copyright 2025 Dolthub, Inc.
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
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBindingWithOidZero tests the behavior of binding parameters when the client specifies a zero OID for any of
// the parameters.
func TestBindingWithOidZero(t *testing.T) {
	// Start up a test server
	ctx, connection, controller := CreateServer(t, "postgres")
	defer controller.Stop()
	conn := connection.Default

	// Create a table to insert into
	_, err := connection.Exec(ctx, "CREATE TABLE my_table (id INT, name varchar(100));")
	require.NoError(t, err)

	args := [][]byte{
		[]byte(strconv.Itoa(42)),
		[]byte("Alice"),
	}
	paramOIDs := []uint32{0, 123}
	paramFormats := []int16{0, 0}
	sql := "INSERT INTO my_table (id, name) VALUES ($1, $2);"

	// Execute a query with the zero OID and assert that we don't get an error
	resultReader := conn.PgConn().ExecParams(ctx, sql, args, paramOIDs, paramFormats, nil)
	result := resultReader.Read()
	require.NoError(t, result.Err)
}
