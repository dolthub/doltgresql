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

	"github.com/jackc/pgx/v5/pgtype"

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

func TestIssue2386(t *testing.T) {
	// https://github.com/dolthub/doltgresql/issues/2386
	ctx, connection, controller := CreateServer(t, "postgres")
	defer controller.Stop()
	conn := connection.Default
	_, err := connection.Exec(ctx, "CREATE TABLE users (id INT PRIMARY KEY, name TEXT NOT NULL);")
	require.NoError(t, err)
	_, err = connection.Exec(ctx, "INSERT INTO users VALUES (1, 'alice'), (2, 'bob'), (3, 'carol'), (4, 'dave');")
	require.NoError(t, err)
	targetIDs := []int32{1, 3}
	rows, err := conn.Query(ctx,
		`SELECT id, name FROM users WHERE id = ANY($1)`,
		targetIDs,
	)
	require.NoError(t, err)
	defer rows.Close()
	i := 0
	for rows.Next() {
		var id int32
		var name string
		err = rows.Scan(&id, &name)
		require.NoError(t, err)
		switch i {
		case 0:
			require.Equal(t, int32(1), id)
			require.Equal(t, "alice", name)
		case 1:
			require.Equal(t, int32(3), id)
			require.Equal(t, "carol", name)
		default:
			t.FailNow()
		}
		i++
	}
}

func TestBindingWithTextArray(t *testing.T) {
	ctx, connection, controller := CreateServer(t, "postgres")
	defer controller.Stop()
	conn := connection.Default

	m := pgtype.NewMap()
	textArray := []string{"foo", "bar"}

	plan := m.PlanEncode(pgtype.TextArrayOID, pgtype.BinaryFormatCode, textArray)
	encodedArr, err := plan.Encode(textArray, nil)
	require.NoError(t, err)

	args := [][]byte{encodedArr}
	paramOIDs := []uint32{1009}
	paramFormats := []int16{1}
	sql := "SELECT $1::text[]"

	resultReader := conn.PgConn().ExecParams(ctx, sql, args, paramOIDs, paramFormats, nil)
	result := resultReader.Read()
	require.NoError(t, result.Err)
}
