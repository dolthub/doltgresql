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
	"encoding/binary"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBindingWithOidZero tests the behavior of binding parameters when the client specifies a zero OID for any of
// the parameters.
func TestBindingWithOidZero(t *testing.T) {
	// Start up a test server
	ctx, connection, controller := CreateServer(t, "postgres")
	defer func() {
		connection.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	conn := connection.Default

	// Create a table to insert into
	_, err := connection.Exec(ctx, "CREATE TABLE my_table (id INT, name varchar(100));")
	require.NoError(t, err)

	args := [][]byte{
		[]byte(strconv.Itoa(42)),
		[]byte("Alice"),
	}
	paramOIDs := []uint32{0, pgtype.TextOID}
	paramFormats := []int16{0, 0}
	sql := "INSERT INTO my_table (id, name) VALUES ($1, $2);"

	// Execute a query with the zero OID and assert that we don't get an error
	resultReader := conn.PgConn().ExecParams(ctx, sql, args, paramOIDs, paramFormats, nil)
	result := resultReader.Read()
	require.NoError(t, result.Err)

	// The explicitly-typed parameter ($2) must be honored rather than discarded, and the
	// inferred parameter ($1) must resolve correctly, so the row must round-trip intact.
	var id int32
	var name string
	require.NoError(t, conn.QueryRow(ctx, "SELECT id, name FROM my_table").Scan(&id, &name))
	assert.Equal(t, int32(42), id)
	assert.Equal(t, "Alice", name)
}

// TestBindingInt4ToBigintWithUntypedNull is a regression test for
// https://github.com/dolthub/doltgresql/issues/2973: an int4-typed parameter bound to
// a BIGINT column was silently corrupted whenever the same statement also had an untyped
// (OID 0) NULL parameter.
func TestBindingInt4ToBigintWithUntypedNull(t *testing.T) {
	ctx, connection, controller := CreateServer(t, "postgres")
	defer func() {
		connection.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	conn := connection.Default

	_, err := connection.Exec(ctx, "CREATE TABLE t1 (a BIGINT, b INT, f BOOLEAN);")
	require.NoError(t, err)

	aBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(aBytes, 1)
	bBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bBytes, 2)

	// $1 is explicitly int4 and 4 bytes wide even though it targets a BIGINT column;
	// $3 is an untyped NULL (OID 0).
	args := [][]byte{aBytes, bBytes, nil}
	paramOIDs := []uint32{pgtype.Int4OID, pgtype.Int4OID, 0}
	paramFormats := []int16{1, 1, 1}
	sql := "INSERT INTO t1 (a, b, f) VALUES ($1, $2, $3);"

	resultReader := conn.PgConn().ExecParams(ctx, sql, args, paramOIDs, paramFormats, nil)
	result := resultReader.Read()
	require.NoError(t, result.Err)

	rows, err := conn.Query(ctx, "SELECT a, b, f FROM t1")
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())
	var a int64
	var b int32
	var f *bool
	require.NoError(t, rows.Scan(&a, &b, &f))
	assert.Equal(t, int64(1), a)
	assert.Equal(t, int32(2), b)
	assert.Nil(t, f)
	assert.False(t, rows.Next())
}

// TestBindingWhereClauseInt4ToBigintWithUntypedNull is a regression test for
// https://github.com/dolthub/doltgresql/issues/2973, triggered via a WHERE-clause comparison
// rather than an INSERT VALUES tuple. go-mysql-server's planbuilder (sql/planbuilder/scalar.go,
// buildScalar) assigns a bindvar's inferred type from whatever column it's compared against, so
// this is a distinct code path from the INSERT case but the same underlying corruption class.
func TestBindingWhereClauseInt4ToBigintWithUntypedNull(t *testing.T) {
	ctx, connection, controller := CreateServer(t, "postgres")
	defer func() {
		connection.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	conn := connection.Default

	_, err := connection.Exec(ctx, "CREATE TABLE t2 (a BIGINT, f BOOLEAN);")
	require.NoError(t, err)
	_, err = connection.Exec(ctx, "INSERT INTO t2 VALUES (1, true), (2, false);")
	require.NoError(t, err)

	aBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(aBytes, 1)

	// $1 is explicitly int4 and compared against a BIGINT column; $2 is an untyped NULL,
	// referenced elsewhere in the WHERE clause so the statement actually binds it.
	args := [][]byte{aBytes, nil}
	paramOIDs := []uint32{pgtype.Int4OID, 0}
	paramFormats := []int16{1, 1}
	sql := "SELECT a, f FROM t2 WHERE a = $1 AND (f = $2 OR f IS NOT NULL);"

	resultReader := conn.PgConn().ExecParams(ctx, sql, args, paramOIDs, paramFormats, nil)
	result := resultReader.Read()
	require.NoError(t, result.Err)
	require.Len(t, result.Rows, 1)
	assert.Equal(t, "1", string(result.Rows[0][0]))
	assert.Equal(t, "t", string(result.Rows[0][1]))
}

// TestBindingUpdateInt4ToBigintWithUntypedNull is a regression test for
// https://github.com/dolthub/doltgresql/issues/2973, triggered via an UPDATE ... SET assignment
// rather than an INSERT VALUES tuple.
func TestBindingUpdateInt4ToBigintWithUntypedNull(t *testing.T) {
	ctx, connection, controller := CreateServer(t, "postgres")
	defer func() {
		connection.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	conn := connection.Default

	_, err := connection.Exec(ctx, "CREATE TABLE t3 (id INT PRIMARY KEY, a BIGINT, f BOOLEAN);")
	require.NoError(t, err)
	_, err = connection.Exec(ctx, "INSERT INTO t3 VALUES (1, 0, true);")
	require.NoError(t, err)

	aBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(aBytes, 42)

	// $1 is explicitly int4 and assigned to a BIGINT column; $2 is an untyped NULL.
	args := [][]byte{aBytes, nil}
	paramOIDs := []uint32{pgtype.Int4OID, 0}
	paramFormats := []int16{1, 1}
	sql := "UPDATE t3 SET a = $1, f = $2 WHERE id = 1;"

	resultReader := conn.PgConn().ExecParams(ctx, sql, args, paramOIDs, paramFormats, nil)
	result := resultReader.Read()
	require.NoError(t, result.Err)

	var a int64
	var f *bool
	require.NoError(t, conn.QueryRow(ctx, "SELECT a, f FROM t3 WHERE id = 1").Scan(&a, &f))
	assert.Equal(t, int64(42), a)
	assert.Nil(t, f)
}

// TestBindingFloat4ToDoubleWithUntypedNull is a regression test for
// https://github.com/dolthub/doltgresql/issues/2973, for a non-integer narrowing type pair
// (float4 -> float8/DOUBLE PRECISION).
func TestBindingFloat4ToDoubleWithUntypedNull(t *testing.T) {
	ctx, connection, controller := CreateServer(t, "postgres")
	defer func() {
		connection.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	conn := connection.Default

	_, err := connection.Exec(ctx, "CREATE TABLE t4 (a DOUBLE PRECISION, f BOOLEAN);")
	require.NoError(t, err)

	aBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(aBytes, math.Float32bits(3.5))

	// $1 is explicitly float4 and targets a DOUBLE PRECISION (float8) column; $2 is an
	// untyped NULL.
	args := [][]byte{aBytes, nil}
	paramOIDs := []uint32{pgtype.Float4OID, 0}
	paramFormats := []int16{1, 1}
	sql := "INSERT INTO t4 (a, f) VALUES ($1, $2);"

	resultReader := conn.PgConn().ExecParams(ctx, sql, args, paramOIDs, paramFormats, nil)
	result := resultReader.Read()
	require.NoError(t, result.Err)

	var a float64
	var f *bool
	require.NoError(t, conn.QueryRow(ctx, "SELECT a, f FROM t4").Scan(&a, &f))
	assert.Equal(t, float64(3.5), a)
	assert.Nil(t, f)
}

// TestBindingInt8ToIntColumnWithUntypedNull is a regression test for
// https://github.com/dolthub/doltgresql/issues/2973, for the mirror-image (narrowing) direction:
// the client explicitly declares $1 as int8 (8 bytes) even though it targets an INT (int4)
// column.
func TestBindingInt8ToIntColumnWithUntypedNull(t *testing.T) {
	ctx, connection, controller := CreateServer(t, "postgres")
	defer func() {
		connection.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	conn := connection.Default

	_, err := connection.Exec(ctx, "CREATE TABLE t5 (a INT, f BOOLEAN);")
	require.NoError(t, err)

	aBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(aBytes, uint64(7))

	args := [][]byte{aBytes, nil}
	paramOIDs := []uint32{pgtype.Int8OID, 0}
	paramFormats := []int16{1, 1}
	sql := "INSERT INTO t5 (a, f) VALUES ($1, $2);"

	resultReader := conn.PgConn().ExecParams(ctx, sql, args, paramOIDs, paramFormats, nil)
	result := resultReader.Read()
	require.NoError(t, result.Err)

	var a int32
	var f *bool
	require.NoError(t, conn.QueryRow(ctx, "SELECT a, f FROM t5").Scan(&a, &f))
	assert.Equal(t, int32(7), a)
	assert.Nil(t, f)
}

// TestBindingShortParameterOIDsArray is a regression test for a defect found from
// https://github.com/dolthub/doltgresql/issues/2973: Postgres allows a Parse message's
// ParameterOIDs to specify types for only a prefix of the actual placeholders.
func TestBindingShortParameterOIDsArray(t *testing.T) {
	ctx, connection, controller := CreateServer(t, "postgres")
	defer func() {
		connection.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	conn := connection.Default

	_, err := connection.Exec(ctx, "CREATE TABLE t6 (a INT, b INT);")
	require.NoError(t, err)

	args := [][]byte{[]byte(strconv.Itoa(1)), []byte(strconv.Itoa(2))}
	// Only $1's OID is specified; $2 is omitted entirely rather than given an explicit 0.
	paramOIDs := []uint32{pgtype.Int4OID}
	paramFormats := []int16{0, 0}
	sql := "INSERT INTO t6 (a, b) VALUES ($1, $2);"

	resultReader := conn.PgConn().ExecParams(ctx, sql, args, paramOIDs, paramFormats, nil)
	result := resultReader.Read()
	require.NoError(t, result.Err)

	var a, b int32
	require.NoError(t, conn.QueryRow(ctx, "SELECT a, b FROM t6").Scan(&a, &b))
	assert.Equal(t, int32(1), a)
	assert.Equal(t, int32(2), b)
}

func TestIssue2386(t *testing.T) {
	// https://github.com/dolthub/doltgresql/issues/2386
	ctx, connection, controller := CreateServer(t, "postgres")
	defer func() {
		connection.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
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
	defer func() {
		connection.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
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

// TestBindingOnConflictDoUpdateWithUntypedNull tests bind var type inferrence for
// ON CONFLICT DO UPDATE statements.
func TestBindingOnConflictDoUpdateWithUntypedNull(t *testing.T) {
	ctx, connection, controller := CreateServer(t, "postgres")
	defer func() {
		connection.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	conn := connection.Default

	_, err := connection.Exec(ctx, "CREATE TABLE tconf (id INT PRIMARY KEY, a BIGINT, f BOOLEAN);")
	require.NoError(t, err)
	_, err = connection.Exec(ctx, "INSERT INTO tconf VALUES (1, 0, true);")
	require.NoError(t, err)

	aBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(aBytes, 42)

	// $1 is explicitly int4 and assigned to a BIGINT column via ON CONFLICT DO UPDATE;
	// $2 is an untyped NULL.
	args := [][]byte{aBytes, nil}
	paramOIDs := []uint32{pgtype.Int4OID, 0}
	paramFormats := []int16{1, 1}
	sql := "INSERT INTO tconf (id, a) VALUES (1, 0) ON CONFLICT (id) DO UPDATE SET a = $1, f = $2;"

	resultReader := conn.PgConn().ExecParams(ctx, sql, args, paramOIDs, paramFormats, nil)
	result := resultReader.Read()
	require.NoError(t, result.Err)

	var a int64
	var f *bool
	require.NoError(t, conn.QueryRow(ctx, "SELECT a, f FROM tconf WHERE id = 1").Scan(&a, &f))
	assert.Equal(t, int64(42), a)
	assert.Nil(t, f)
}

// TestBindingMultipleInferredAndExplicitOIDsInterleaved tests the OID-merge logic by interleaving
// multiple unspecified (OID 0) positions with multiple explicit, narrower-than-column OIDs.
func TestBindingMultipleInferredAndExplicitOIDsInterleaved(t *testing.T) {
	ctx, connection, controller := CreateServer(t, "postgres")
	defer func() {
		connection.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	conn := connection.Default

	_, err := connection.Exec(ctx, "CREATE TABLE t7 (a INT, b BIGINT, c INT, d DOUBLE PRECISION, e BOOLEAN);")
	require.NoError(t, err)

	bBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bBytes, 22)
	dBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(dBytes, math.Float32bits(4.5))

	// $1 (-> a) and $3 (-> c) are unspecified (OID 0); $2 is explicit int4 targeting a BIGINT
	// column, $4 is explicit float4 targeting a DOUBLE PRECISION column, $5 is an untyped NULL.
	args := [][]byte{[]byte(strconv.Itoa(11)), bBytes, []byte(strconv.Itoa(33)), dBytes, nil}
	paramOIDs := []uint32{0, pgtype.Int4OID, 0, pgtype.Float4OID, 0}
	paramFormats := []int16{0, 1, 0, 1, 1}
	sql := "INSERT INTO t7 (a, b, c, d, e) VALUES ($1, $2, $3, $4, $5);"

	resultReader := conn.PgConn().ExecParams(ctx, sql, args, paramOIDs, paramFormats, nil)
	result := resultReader.Read()
	require.NoError(t, result.Err)

	var a, c int32
	var b int64
	var d float64
	var e *bool
	require.NoError(t, conn.QueryRow(ctx, "SELECT a, b, c, d, e FROM t7").Scan(&a, &b, &c, &d, &e))
	assert.Equal(t, int32(11), a)
	assert.Equal(t, int64(22), b)
	assert.Equal(t, int32(33), c)
	assert.Equal(t, float64(4.5), d)
	assert.Nil(t, e)
}

// TestBindingJSONBExtractPathTextWithUntypedParam tests that an untyped parameter
// compared against the result of `jsonb #>> text[]` is correctly inferred as JSONB (not TEXT).
// https://github.com/dolthub/doltgresql/issues/3012.
func TestBindingJSONBExtractPathTextWithUntypedParam(t *testing.T) {
	ctx, connection, controller := CreateServer(t, "postgres")
	defer func() {
		connection.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	conn := connection.Default

	_, err := connection.Exec(ctx, "CREATE TABLE t8 (id TEXT PRIMARY KEY, props JSONB);")
	require.NoError(t, err)
	_, err = connection.Exec(ctx, `INSERT INTO t8 VALUES ('a', '{"name":"Alice"}');`)
	require.NoError(t, err)

	// $1 is untyped (OID 0) and must be inferred as text, since `#>>` returns text even
	// though its left operand (props) is jsonb.
	args := [][]byte{[]byte("Alice")}
	paramOIDs := []uint32{0}
	paramFormats := []int16{0}
	sql := "SELECT count(*) FROM t8 WHERE props #>> array['name']::text[] = $1"

	resultReader := conn.PgConn().ExecParams(ctx, sql, args, paramOIDs, paramFormats, nil)
	result := resultReader.Read()
	require.NoError(t, result.Err)
	require.Len(t, result.Rows, 1)
	assert.Equal(t, "1", string(result.Rows[0][0]))
}

// TestBindingExplicitCastToTimestamptzPreservesOffset is a regression test for
// https://github.com/dolthub/doltgresql/issues/3093
func TestBindingExplicitCastToTimestamptzPreservesOffset(t *testing.T) {
	ctx, connection, controller := CreateServer(t, "postgres")
	defer func() {
		connection.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	conn := connection.Default

	_, err := connection.Exec(ctx, "CREATE TABLE t9 (id INT PRIMARY KEY, ts timestamptz);")
	require.NoError(t, err)

	// Use an offset that doesn't match the test server's local zone, so a bug that drops the
	// offset entirely (rather than double-applying it) can't coincidentally still pass.
	const text = "2026-08-21 12:00:00+05:00"
	want := time.Date(2026, 8, 21, 12, 0, 0, 0, time.FixedZone("", 5*3600))
	_, err = connection.Exec(ctx, "INSERT INTO t9 VALUES (1, $1)", want)
	require.NoError(t, err)

	// Standalone cast: the placeholder's inferred type must be timestamptz, not timestamp, so
	// the embedded offset is honored rather than discarded.
	var castBack time.Time
	err = conn.QueryRow(ctx, "SELECT $1::timestamptz", text).Scan(&castBack)
	require.NoError(t, err)
	assert.True(t, castBack.Equal(want), "got %v, want %v", castBack, want)

	// The same cast used in a WHERE clause comparison must match the stored row.
	var count int
	err = conn.QueryRow(ctx, "SELECT count(*) FROM t9 WHERE ts = $1::timestamptz", text).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
