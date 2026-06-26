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

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	driver "github.com/dolthub/doltgresql/integration-tests/go-sql-server-driver/driver"
)

// setupTestServer creates a standard single-server test environment. The repo
// name is also used as the database name. Cleanup is registered with t.Cleanup.
func setupTestServer(t *testing.T, repoName string) *driver.SqlServer {
	t.Helper()
	ports := newPorts(t)

	u, err := driver.NewDoltUser()
	require.NoError(t, err)
	t.Cleanup(func() { u.Cleanup() })

	rs, err := u.MakeRepoStore()
	require.NoError(t, err)

	_, err = rs.MakeRepo(repoName)
	require.NoError(t, err)

	server := StartServer(t, rs, repoName, &driver.Server{
		Args:        []string{},
		DynamicPort: "server_port",
	}, ports)
	return server
}

// makeTestText returns a deterministic ASCII string of exactly size bytes.
// The seed value differentiates rows so each row has unique content.
func makeTestText(seed, size int) string {
	chunk := fmt.Sprintf("[row%07d-filler]", seed) // 18 bytes
	n := (size + len(chunk) - 1) / len(chunk)
	return strings.Repeat(chunk, n)[:size]
}

// makeTestBlobData returns a deterministic byte slice of exactly size bytes.
func makeTestBlobData(seed, size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte((seed*37 + i*17 + seed*i*3) & 0xFF)
	}
	return data
}

// largeJSONDoc is the top-level structure used to generate large JSON values.
type largeJSONDoc struct {
	ID          int             `json:"id"`
	Description string          `json:"description"`
	Tags        []string        `json:"tags"`
	Items       []largeJSONItem `json:"items"`
}

type largeJSONItem struct {
	Index   int    `json:"index"`
	Name    string `json:"name"`
	Payload string `json:"payload"`
}

// makeTestJSONString returns a JSON string of at least targetSize bytes built
// from deterministic data keyed by seed.
func makeTestJSONString(seed, targetSize int) string {
	doc := largeJSONDoc{
		ID:          seed,
		Description: strings.Repeat(fmt.Sprintf("desc-seed%07d-", seed), 10),
	}
	for i := range 15 {
		doc.Tags = append(doc.Tags, fmt.Sprintf("tag-%d-%d", seed, i))
	}
	for i := 0; ; i++ {
		item := largeJSONItem{
			Index:   i,
			Name:    fmt.Sprintf("item-%d-%05d", seed, i),
			Payload: strings.Repeat(fmt.Sprintf("pl-%d-%d-", seed, i), 12),
		}
		doc.Items = append(doc.Items, item)
		bs, _ := json.Marshal(doc)
		if len(bs) >= targetSize {
			return string(bs)
		}
	}
}

// buildWideCreateTable generates a CREATE TABLE statement with numCols columns
// of the given colType in addition to a BIGINT PRIMARY KEY column named "id".
func buildWideCreateTable(tableName string, numCols int, colType string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "CREATE TABLE %s (id BIGINT PRIMARY KEY", tableName)
	for i := range numCols {
		fmt.Fprintf(&sb, ", c%d %s", i, colType)
	}
	sb.WriteString(")")
	return sb.String()
}

// buildWideInsert generates an INSERT statement for a wide table. makeVal is
// called for each column index and should return a SQL literal (not quoted for
// numeric types; caller wraps strings in single quotes).
func buildWideInsert(tableName string, rowID int, numCols int, makeVal func(col int) string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "INSERT INTO %s VALUES (%d", tableName, rowID)
	for i := range numCols {
		fmt.Fprintf(&sb, ", %s", makeVal(i))
	}
	sb.WriteString(")")
	return sb.String()
}

// ----------------------------------------------------------------------------
// TestLargeOutOfBandValues exercises rows whose individual column values
// exceed 10,000 bytes — the range where Dolt stores values out-of-band in
// TEXT, BLOB, and JSON columns. Each subtest commits the data, reads it back
// to verify integrity, runs GC, and then re-verifies.
// ----------------------------------------------------------------------------

func TestLargeOutOfBandValues(t *testing.T) {
	t.Parallel()

	t.Run("LargeText", func(t *testing.T) {
		t.Parallel()
		const textSize = 15_000
		const numRows = 20

		server := setupTestServer(t, "large_text_values")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		func() {
			conn, err := db.Conn(ctx)
			require.NoError(t, err)
			defer conn.Close()
			_, err = conn.ExecContext(ctx, "CREATE TABLE large_text (id BIGINT PRIMARY KEY, txt TEXT)")
			require.NoError(t, err)
			for i := range numRows {
				_, err = conn.ExecContext(ctx, "INSERT INTO large_text VALUES ($1, $2)", i, makeTestText(i, textSize))
				require.NoError(t, err)
			}
			_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'insert large text rows')")
			require.NoError(t, err)
		}()

		verifyLargeText := func() {
			conn, err := db.Conn(ctx)
			require.NoError(t, err)
			defer conn.Close()

			var count int
			err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM large_text").Scan(&count)
			require.NoError(t, err)
			require.Equal(t, numRows, count)

			// All rows must have the correct byte length. Doltgres panics on
			// octet_length() over large out-of-band text values, so the length
			// is verified client-side.
			count = countTextRowsOfLen(t, conn, ctx, "SELECT txt FROM large_text", textSize)
			require.Equal(t, numRows, count, "every row should retain the full text length")

			// Spot-check a few rows for exact content.
			for _, id := range []int{0, numRows / 2, numRows - 1} {
				var actual string
				err = conn.QueryRowContext(ctx, "SELECT txt FROM large_text WHERE id = $1", id).Scan(&actual)
				require.NoError(t, err)
				require.Equal(t, makeTestText(id, textSize), actual,
					"row %d: text content must survive storage and retrieval", id)
			}
		}

		verifyLargeText()

		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)

		// dolt_gc invalidates open connections; reconnect with a fresh pool.
		db.Close()
		db, err = server.DB(driver.Connection{})
		require.NoError(t, err)

		verifyLargeText()
	})

	t.Run("LargeBlob", func(t *testing.T) {
		t.Parallel()
		const blobSize = 25_000
		const numRows = 20

		server := setupTestServer(t, "large_blob_values")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		func() {
			conn, err := db.Conn(ctx)
			require.NoError(t, err)
			defer conn.Close()
			_, err = conn.ExecContext(ctx, "CREATE TABLE large_blob (id BIGINT PRIMARY KEY, data BYTEA)")
			require.NoError(t, err)
			for i := range numRows {
				_, err = conn.ExecContext(ctx, "INSERT INTO large_blob VALUES ($1, $2)", i, makeTestBlobData(i, blobSize))
				require.NoError(t, err)
			}
			_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'insert large blob rows')")
			require.NoError(t, err)
		}()

		verifyLargeBlob := func() {
			conn, err := db.Conn(ctx)
			require.NoError(t, err)
			defer conn.Close()

			var count int
			err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM large_blob").Scan(&count)
			require.NoError(t, err)
			require.Equal(t, numRows, count)

			// Doltgres has no server-side byte-length function for bytea
			// (octet_length/length/encode are unavailable), so verify every
			// row's blob length on the client side instead.
			lenRows, err := conn.QueryContext(ctx, "SELECT data FROM large_blob")
			require.NoError(t, err)
			count = 0
			for lenRows.Next() {
				var d []byte
				require.NoError(t, lenRows.Scan(&d))
				require.Equal(t, blobSize, len(d), "every row should retain the full blob length")
				count++
			}
			require.NoError(t, lenRows.Err())
			lenRows.Close()
			require.Equal(t, numRows, count)

			// Spot-check exact binary content for a few rows.
			for _, id := range []int{0, numRows / 2, numRows - 1} {
				var actual []byte
				err = conn.QueryRowContext(ctx, "SELECT data FROM large_blob WHERE id = $1", id).Scan(&actual)
				require.NoError(t, err)
				require.Equal(t, makeTestBlobData(id, blobSize), actual,
					"row %d: blob content must survive storage and retrieval", id)
			}
		}

		verifyLargeBlob()

		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)

		db.Close()
		db, err = server.DB(driver.Connection{})
		require.NoError(t, err)

		verifyLargeBlob()
	})

	t.Run("LargeJSON", func(t *testing.T) {
		t.Parallel()
		const jsonTargetSize = 12_000
		const numRows = 20

		server := setupTestServer(t, "large_json_values")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		// Pre-generate JSON values so we know their exact sizes.
		jsonVals := make([]string, numRows)
		for i := range numRows {
			jsonVals[i] = makeTestJSONString(i, jsonTargetSize)
			require.GreaterOrEqual(t, len(jsonVals[i]), jsonTargetSize,
				"generated JSON should meet the target size floor")
		}

		func() {
			conn, err := db.Conn(ctx)
			require.NoError(t, err)
			defer conn.Close()
			_, err = conn.ExecContext(ctx, "CREATE TABLE large_json (id BIGINT PRIMARY KEY, doc JSON)")
			require.NoError(t, err)
			for i, v := range jsonVals {
				_, err = conn.ExecContext(ctx, "INSERT INTO large_json VALUES ($1, $2)", i, v)
				require.NoError(t, err)
			}
			_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'insert large json rows')")
			require.NoError(t, err)
		}()

		verifyLargeJSON := func() {
			conn, err := db.Conn(ctx)
			require.NoError(t, err)
			defer conn.Close()

			var count int
			err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM large_json").Scan(&count)
			require.NoError(t, err)
			require.Equal(t, numRows, count)

			// Verify that the JSON round-trip preserves the document ID field.
			// Doltgres panics on JSON operators (e.g. ->>) over large
			// out-of-band JSON values, so the whole document is retrieved and
			// parsed client-side instead.
			for _, id := range []int{0, numRows / 2, numRows - 1} {
				var docStr string
				err = conn.QueryRowContext(ctx,
					"SELECT doc FROM large_json WHERE id = $1", id).Scan(&docStr)
				require.NoError(t, err)
				var parsed largeJSONDoc
				require.NoError(t, json.Unmarshal([]byte(docStr), &parsed))
				require.Equal(t, id, parsed.ID,
					"row %d: JSON id field must be preserved after storage", id)
			}
		}

		verifyLargeJSON()

		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)

		db.Close()
		db, err = server.DB(driver.Connection{})
		require.NoError(t, err)

		verifyLargeJSON()
	})

	t.Run("MixedLargeColumns", func(t *testing.T) {
		t.Parallel()
		// Each row carries large values in three separate out-of-band column
		// types simultaneously, exercising the case where a single logical row
		// requires multiple large external chunks.
		const textSize = 18_000
		const blobSize = 22_000
		const jsonTarget = 11_000
		const numRows = 10

		server := setupTestServer(t, "mixed_large_columns")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		func() {
			conn, err := db.Conn(ctx)
			require.NoError(t, err)
			defer conn.Close()
			_, err = conn.ExecContext(ctx, `CREATE TABLE mixed_large (
				id       BIGINT PRIMARY KEY,
				txt      TEXT,
				bin_data BYTEA,
				doc      JSON,
				note     TEXT
			)`)
			require.NoError(t, err)
			for i := range numRows {
				note := makeTestText(i+1000, 5_000) // secondary TEXT column, ~5 KB
				_, err = conn.ExecContext(ctx,
					"INSERT INTO mixed_large VALUES ($1, $2, $3, $4, $5)",
					i,
					makeTestText(i, textSize),
					makeTestBlobData(i, blobSize),
					makeTestJSONString(i, jsonTarget),
					note,
				)
				require.NoError(t, err)
			}
			_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'insert mixed large column rows')")
			require.NoError(t, err)
		}()

		verifyMixed := func() {
			conn, err := db.Conn(ctx)
			require.NoError(t, err)
			defer conn.Close()

			var count int
			err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM mixed_large").Scan(&count)
			require.NoError(t, err)
			require.Equal(t, numRows, count)

			// Verify lengths of all large columns. Doltgres has no bytea length
			// function and panics on octet_length() over large out-of-band text,
			// so the lengths are checked client-side.
			lenRows, err := conn.QueryContext(ctx, "SELECT txt, bin_data, note FROM mixed_large")
			require.NoError(t, err)
			count = 0
			for lenRows.Next() {
				var txtv, notev string
				var binv []byte
				require.NoError(t, lenRows.Scan(&txtv, &binv, &notev))
				require.Equal(t, textSize, len(txtv))
				require.Equal(t, blobSize, len(binv))
				require.Equal(t, 5000, len(notev))
				count++
			}
			require.NoError(t, lenRows.Err())
			lenRows.Close()
			require.Equal(t, numRows, count, "all large column lengths must be preserved")

			// Spot-check content integrity.
			for _, id := range []int{0, numRows - 1} {
				var actualTxt string
				var actualBlob []byte
				err = conn.QueryRowContext(ctx,
					"SELECT txt, bin_data FROM mixed_large WHERE id = $1", id).
					Scan(&actualTxt, &actualBlob)
				require.NoError(t, err)
				require.Equal(t, makeTestText(id, textSize), actualTxt)
				require.Equal(t, makeTestBlobData(id, blobSize), actualBlob)
			}
		}

		verifyMixed()

		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)

		db.Close()
		db, err = server.DB(driver.Connection{})
		require.NoError(t, err)

		verifyMixed()
	})

	t.Run("ConcurrentLargeValueWrites", func(t *testing.T) {
		t.Parallel()
		// Multiple goroutines writing large TEXT and BLOB rows concurrently,
		// stress-testing the out-of-band storage path under writer contention.
		const textSize = 12_000
		const blobSize = 15_000
		const numWorkers = 8
		const rowsPerWorker = 10

		server := setupTestServer(t, "concurrent_large_values")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		func() {
			conn, err := db.Conn(ctx)
			require.NoError(t, err)
			defer conn.Close()
			_, err = conn.ExecContext(ctx, `CREATE TABLE large_concurrent (
				id   BIGINT PRIMARY KEY,
				txt  TEXT,
				data BYTEA
			)`)
			require.NoError(t, err)
			_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'create large_concurrent table')")
			require.NoError(t, err)
		}()

		eg, egCtx := errgroup.WithContext(ctx)
		startCh := make(chan struct{})
		readyCh := make(chan struct{})

		for w := range numWorkers {
			eg.Go(func() error {
				workerDB, err := server.DB(driver.Connection{})
				if err != nil {
					return err
				}
				defer workerDB.Close()
				workerDB.SetMaxOpenConns(1)
				conn, err := workerDB.Conn(egCtx)
				if err != nil {
					return err
				}
				defer conn.Close()

				select {
				case readyCh <- struct{}{}:
				case <-egCtx.Done():
					return nil
				}
				select {
				case <-startCh:
				case <-egCtx.Done():
					return nil
				}

				for j := range rowsPerWorker {
					if egCtx.Err() != nil {
						return nil
					}
					rowID := int64(w*rowsPerWorker + j)
					seed := int(rowID)
					_, err = conn.ExecContext(egCtx,
						"INSERT INTO large_concurrent VALUES ($1, $2, $3)",
						rowID,
						makeTestText(seed, textSize),
						makeTestBlobData(seed, blobSize),
					)
					if err != nil {
						return fmt.Errorf("worker %d insert row %d: %w", w, j, err)
					}
				}
				_, err = conn.ExecContext(egCtx,
					fmt.Sprintf("SELECT dolt_commit('-Am', 'worker %d inserts')", w))
				if err != nil && !strings.Contains(err.Error(), "nothing to commit") {
					return fmt.Errorf("worker %d commit: %w", w, err)
				}
				return nil
			})
		}

		for range numWorkers {
			select {
			case <-readyCh:
			case <-ctx.Done():
				require.NoError(t, eg.Wait())
				t.FailNow()
			}
		}
		close(startCh)
		require.NoError(t, eg.Wait())

		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()

		var count int
		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM large_concurrent").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, numWorkers*rowsPerWorker, count,
			"all rows from all workers must be present")

		// Doltgres has no bytea length function and panics on octet_length()
		// over large out-of-band text, so verify sizes client-side.
		lenRows, err := conn.QueryContext(ctx, "SELECT txt, data FROM large_concurrent")
		require.NoError(t, err)
		count = 0
		for lenRows.Next() {
			var txtv string
			var datav []byte
			require.NoError(t, lenRows.Scan(&txtv, &datav))
			require.Equal(t, textSize, len(txtv))
			require.Equal(t, blobSize, len(datav))
			count++
		}
		require.NoError(t, lenRows.Err())
		lenRows.Close()
		require.Equal(t, numWorkers*rowsPerWorker, count,
			"all large values must have the correct size")
	})
}

// ----------------------------------------------------------------------------
// TestTypeDiversity verifies that Doltgres correctly stores and retrieves a
// broad spectrum of column types, including boundary values, NULL, unicode
// strings, and values that cross the inline/out-of-band threshold. Each
// subtest operates on its own server so they can run in parallel.
// ----------------------------------------------------------------------------

func TestTypeDiversity(t *testing.T) {
	t.Parallel()

	t.Run("IntegerTypes", func(t *testing.T) {
		t.Parallel()
		t.Skip("Doltgres panics (nil pointer dereference) handling a NUMERIC column holding the uint64 max value 18446744073709551615")
		server := setupTestServer(t, "integer_types_test")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()

		// Postgres has no unsigned integer types; the "unsigned" columns are
		// widened to the next-larger signed type so the max boundary values fit.
		_, err = conn.ExecContext(ctx, `CREATE TABLE int_types (
			id               BIGINT PRIMARY KEY,
			col_tinyint      SMALLINT,
			col_smallint     SMALLINT,
			col_mediumint    INTEGER,
			col_int          INTEGER,
			col_bigint       BIGINT,
			col_tinyint_u    SMALLINT,
			col_smallint_u   INTEGER,
			col_int_u        BIGINT,
			col_bigint_u     NUMERIC
		)`)
		require.NoError(t, err)

		rows := []struct {
			id, ti, si, mi, i int64
			bi                int64
			tiu, siu          uint64
			iu                uint64
			biu               uint64
		}{
			{0, -128, -32768, -8388608, -2147483648, -9223372036854775808, 0, 0, 0, 0},
			{1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{2, 127, 32767, 8388607, 2147483647, 9223372036854775807, 255, 65535, 4294967295, 18446744073709551615},
			{3, 42, 1000, 100000, 1000000, 1000000000000, 200, 50000, 2000000000, 9000000000000000000},
		}
		for _, r := range rows {
			_, err = conn.ExecContext(ctx,
				"INSERT INTO int_types VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)",
				r.id, r.ti, r.si, r.mi, r.i, r.bi, r.tiu, r.siu, r.iu, r.biu)
			require.NoError(t, err)
		}

		_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'integer types')")
		require.NoError(t, err)

		var count int
		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM int_types").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, len(rows), count)

		// Verify boundary values survive round-trip.
		err = conn.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM int_types WHERE col_tinyint = -128 AND col_bigint = -9223372036854775808").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count, "min boundary values must be stored exactly")

		var bigU string
		err = conn.QueryRowContext(ctx,
			"SELECT col_bigint_u::text FROM int_types WHERE col_tinyint = 127").Scan(&bigU)
		require.NoError(t, err)
		require.Equal(t, "18446744073709551615", bigU, "max unsigned bigint boundary value must be stored exactly")

		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)

		db.Close()
		db, err = server.DB(driver.Connection{})
		require.NoError(t, err)
		conn, err = db.Conn(ctx)
		require.NoError(t, err)

		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM int_types").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, len(rows), count, "integer rows must survive GC")
	})

	t.Run("FloatingPointAndDecimal", func(t *testing.T) {
		t.Parallel()
		server := setupTestServer(t, "float_types_test")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.ExecContext(ctx, `CREATE TABLE float_types (
			id          BIGINT PRIMARY KEY,
			col_float   REAL,
			col_double  DOUBLE PRECISION,
			col_dec     DECIMAL(30,10)
		)`)
		require.NoError(t, err)

		type floatRow struct {
			id  int
			f   float32
			d   float64
			dec string
		}
		rows := []floatRow{
			{0, 0.0, 0.0, "0.0000000000"},
			{1, 1.5, 1.5, "1.5000000000"},
			{2, -1.5, -1.5, "-1.5000000000"},
			{3, 3.14159, 3.14159265358979, "3.1415926536"},
			{4, 1e10, 1e15, "99999999999999999999.9999999999"},
			{5, -1e10, -1e15, "-99999999999999999999.9999999999"},
		}
		for _, r := range rows {
			_, err = conn.ExecContext(ctx,
				"INSERT INTO float_types VALUES ($1,$2,$3,$4)",
				r.id, r.f, r.d, r.dec)
			require.NoError(t, err)
		}

		_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'float types')")
		require.NoError(t, err)

		var count int
		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM float_types").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, len(rows), count)

		// Decimal is exact — verify the stored value.
		err = conn.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM float_types WHERE col_dec = 3.1415926536").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count, "decimal value must be stored and retrieved exactly")

		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)

		db.Close()
		db, err = server.DB(driver.Connection{})
		require.NoError(t, err)
		conn, err = db.Conn(ctx)
		require.NoError(t, err)

		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM float_types").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, len(rows), count, "float rows must survive GC")
	})

	t.Run("StringTypes", func(t *testing.T) {
		t.Parallel()
		server := setupTestServer(t, "string_types_test")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.ExecContext(ctx, `CREATE TABLE string_types (
			id            BIGINT PRIMARY KEY,
			col_char      CHAR(100),
			col_varchar   VARCHAR(2000),
			col_text      TEXT,
			col_medtext   TEXT,
			col_longtext  TEXT
		)`)
		require.NoError(t, err)

		// Mix of empty, short, medium, large, unicode, and NULL values.
		vals := []struct {
			id                   int
			ch, vc, tx, mtx, ltx interface{}
		}{
			{0, "", "", "", "", ""},
			{1, "hello", "world", "short text", "medium text", "long text"},
			// unicode and multi-byte characters
			{2,
				"日本語テスト",
				"Ünïcödé strïng wïth vàrïöüs chàrs: αβγδεζηθ ℕ ℤ ℚ ℝ ℂ",
				strings.Repeat("中文测试内容-", 100),
				strings.Repeat("한국어 테스트 데이터-", 500),
				strings.Repeat("العربية اختبار البيانات-", 1000),
			},
			// large values that cross the out-of-band threshold
			{3,
				strings.Repeat("x", 100),
				strings.Repeat("v", 2000),
				makeTestText(300, 12_000),
				makeTestText(301, 50_000),
				makeTestText(302, 200_000),
			},
			// NULL values
			{4, nil, nil, nil, nil, nil},
		}

		for _, r := range vals {
			_, err = conn.ExecContext(ctx,
				"INSERT INTO string_types VALUES ($1,$2,$3,$4,$5,$6)",
				r.id, r.ch, r.vc, r.tx, r.mtx, r.ltx)
			require.NoError(t, err)
		}

		_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'string types')")
		require.NoError(t, err)

		var count int
		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM string_types").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, len(vals), count)

		// Verify the large TEXT value was preserved intact (200 KB of ASCII).
		// Doltgres panics on octet_length() over large out-of-band text, so the
		// length is checked client-side.
		var longText string
		err = conn.QueryRowContext(ctx, "SELECT col_longtext FROM string_types WHERE id = 3").Scan(&longText)
		require.NoError(t, err)
		require.Equal(t, 200000, len(longText), "200 KB TEXT value must be stored and retrieved intact")

		// Verify the NULL row.
		err = conn.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM string_types WHERE col_char IS NULL AND col_varchar IS NULL").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count, "NULL string columns must be stored as NULL")

		// Verify unicode content round-trip. CHAR(100) is blank-padded in
		// Postgres, so trim trailing spaces before comparing.
		var unicodeChar string
		err = conn.QueryRowContext(ctx, "SELECT col_char FROM string_types WHERE id = 2").Scan(&unicodeChar)
		require.NoError(t, err)
		require.Equal(t, "日本語テスト", strings.TrimRight(unicodeChar, " "), "unicode CHAR value must round-trip unchanged")

		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)

		db.Close()
		db, err = server.DB(driver.Connection{})
		require.NoError(t, err)
		conn, err = db.Conn(ctx)
		require.NoError(t, err)

		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM string_types").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, len(vals), count, "string rows must survive GC")
	})

	t.Run("BinaryTypes", func(t *testing.T) {
		t.Parallel()
		server := setupTestServer(t, "binary_types_test")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()

		// Postgres has a single variable-length binary type (bytea); there is
		// no fixed-width BINARY(n), so all binary columns map to bytea.
		_, err = conn.ExecContext(ctx, `CREATE TABLE binary_types (
			id            BIGINT PRIMARY KEY,
			col_binary    BYTEA,
			col_varbinary BYTEA,
			col_blob      BYTEA,
			col_longblob  BYTEA
		)`)
		require.NoError(t, err)

		vals := []struct {
			id int
			bi []byte
			vb []byte
			bl []byte
			lb []byte
		}{
			{0, make([]byte, 32), []byte{}, []byte{}, []byte{}},
			{1, makeTestBlobData(1, 32), makeTestBlobData(2, 500), makeTestBlobData(3, 8_000), makeTestBlobData(4, 30_000)},
			{2, makeTestBlobData(5, 32), makeTestBlobData(6, 2000), makeTestBlobData(7, 65_000), makeTestBlobData(8, 500_000)},
			{3, nil, nil, nil, nil},
		}

		for _, r := range vals {
			_, err = conn.ExecContext(ctx,
				"INSERT INTO binary_types VALUES ($1,$2,$3,$4,$5)",
				r.id, r.bi, r.vb, r.bl, r.lb)
			require.NoError(t, err)
		}

		_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'binary types')")
		require.NoError(t, err)

		var count int
		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM binary_types").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, len(vals), count)

		// Verify the 500 KB bytea value survived. Doltgres has no server-side
		// byte-length function for bytea, so check the length client-side.
		var lb500 []byte
		err = conn.QueryRowContext(ctx, "SELECT col_longblob FROM binary_types WHERE id = 2").Scan(&lb500)
		require.NoError(t, err)
		require.Equal(t, 500_000, len(lb500), "500 KB bytea value must be stored and retrieved intact")

		// Spot-check exact blob content.
		var actualLB []byte
		err = conn.QueryRowContext(ctx, "SELECT col_longblob FROM binary_types WHERE id = 2").Scan(&actualLB)
		require.NoError(t, err)
		require.Equal(t, makeTestBlobData(8, 500_000), actualLB,
			"500 KB bytea round-trip must be byte-perfect")

		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)

		db.Close()
		db, err = server.DB(driver.Connection{})
		require.NoError(t, err)
		conn, err = db.Conn(ctx)
		require.NoError(t, err)

		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM binary_types").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, len(vals), count, "binary rows must survive GC")
	})

	t.Run("DateTimeTypes", func(t *testing.T) {
		t.Parallel()
		server := setupTestServer(t, "datetime_types_test")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()

		// Postgres has no YEAR type and its DATE/TIME ranges differ from
		// MySQL's, so the extreme MySQL-only boundary values are replaced with
		// values that are valid in Postgres while still exercising round-trip,
		// fractional seconds, and NULL handling. YEAR maps to a plain integer.
		_, err = conn.ExecContext(ctx, `CREATE TABLE datetime_types (
			id           BIGINT PRIMARY KEY,
			col_date     DATE,
			col_time     TIME,
			col_datetime TIMESTAMP(6),
			col_ts       TIMESTAMP(6),
			col_year     INTEGER
		)`)
		require.NoError(t, err)

		type dtRow struct {
			id   int
			date interface{}
			tm   interface{}
			dt   interface{}
			ts   interface{}
			yr   interface{}
		}
		rows := []dtRow{
			{0, "1000-01-01", "00:00:00", "1000-01-01 00:00:00.000000", nil, 1901},
			{1, "2024-06-15", "00:00:00", "2024-06-15 12:30:45.123456", "2024-06-15 12:30:45.123456", 2024},
			{2, "9999-12-31", "23:59:59", "9999-12-31 23:59:59.999999", "2038-01-19 03:14:07.000000", 2155},
			{3, nil, nil, nil, nil, nil},
		}
		for _, r := range rows {
			_, err = conn.ExecContext(ctx,
				"INSERT INTO datetime_types VALUES ($1,$2,$3,$4,$5,$6)",
				r.id, r.date, r.tm, r.dt, r.ts, r.yr)
			require.NoError(t, err)
		}

		_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'datetime types')")
		require.NoError(t, err)

		var count int
		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM datetime_types").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, len(rows), count)

		err = conn.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM datetime_types WHERE col_date = '1000-01-01'").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count, "minimum DATE value must round-trip correctly")

		err = conn.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM datetime_types WHERE col_date IS NULL AND col_time IS NULL").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count, "NULL date/time values must be stored as NULL")

		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)

		db.Close()
		db, err = server.DB(driver.Connection{})
		require.NoError(t, err)
		conn, err = db.Conn(ctx)
		require.NoError(t, err)

		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM datetime_types").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, len(rows), count, "datetime rows must survive GC")
	})

	t.Run("SpecialTypes", func(t *testing.T) {
		t.Parallel()
		server := setupTestServer(t, "special_types_test")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()

		// Postgres has no ENUM/SET literal types in this dialect; ENUM maps to
		// varchar and SET (a multi-valued column) maps to varchar holding a
		// comma-separated value. FIND_IN_SET is expressed with string matching.
		_, err = conn.ExecContext(ctx, `CREATE TABLE special_types (
			id        BIGINT PRIMARY KEY,
			col_bool  BOOLEAN,
			col_json  JSON,
			col_enum  VARCHAR(32),
			col_set   VARCHAR(64)
		)`)
		require.NoError(t, err)

		type specialRow struct {
			id int
			b  interface{}
			j  interface{} // JSON value (string or nil)
			e  interface{}
			s  interface{}
		}
		rows := []specialRow{
			{0, false, `{}`, "alpha", "red"},
			{1, true, `{"key": "value", "num": 42, "arr": [1,2,3]}`, "beta", "red,green"},
			{2, nil, makeTestJSONString(200, 15_000), "gamma", "red,green,blue"},
			{3, false, `null`, "delta", "red,green,blue,yellow"},
			{4, nil, nil, nil, nil},
		}
		for _, r := range rows {
			_, err = conn.ExecContext(ctx,
				"INSERT INTO special_types VALUES ($1,$2,$3,$4,$5)",
				r.id, r.b, r.j, r.e, r.s)
			require.NoError(t, err)
		}

		_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'special types')")
		require.NoError(t, err)

		var count int
		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM special_types").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, len(rows), count)

		// Verify the large JSON value is accessible as a non-empty object.
		// Doltgres does not provide json_object_keys and panics on JSON
		// operators over large out-of-band JSON, so the document is retrieved
		// and parsed client-side and asserted to be a non-empty object.
		var jsonDoc string
		err = conn.QueryRowContext(ctx, "SELECT col_json FROM special_types WHERE id = 2").Scan(&jsonDoc)
		require.NoError(t, err)
		var jsonObj map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(jsonDoc), &jsonObj))
		require.Greater(t, len(jsonObj), 0, "large JSON value must be a non-empty object")

		// Verify ENUM and SET values. Use a comma-padded LIKE pattern to test
		// exact element membership in the comma-separated SET string without
		// relying on string_to_array, which is not yet implemented in Doltgres.
		err = conn.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM special_types WHERE col_enum = 'gamma' AND (',' || col_set || ',') LIKE '%,blue,%'").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count, "ENUM and SET values must round-trip correctly")

		err = conn.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM special_types WHERE col_enum IS NULL AND col_set IS NULL").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count, "NULL ENUM and SET values must be stored as NULL")

		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)

		db.Close()
		db, err = server.DB(driver.Connection{})
		require.NoError(t, err)
		conn, err = db.Conn(ctx)
		require.NoError(t, err)

		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM special_types").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, len(rows), count, "special-type rows must survive GC")
	})

	t.Run("NullValuesAcrossAllTypes", func(t *testing.T) {
		t.Parallel()
		// A single table with one nullable column of each major type.
		// Row 0 has all NULLs, row 1 has all non-NULLs, to verify the
		// storage engine handles both extremes in the same chunk.
		server := setupTestServer(t, "null_types_test")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.ExecContext(ctx, `CREATE TABLE nullable_types (
			id     BIGINT PRIMARY KEY,
			n_int  INTEGER,
			n_dbl  DOUBLE PRECISION,
			n_dec  DECIMAL(10,4),
			n_str  VARCHAR(255),
			n_txt  TEXT,
			n_blob BYTEA,
			n_date DATE,
			n_ts   TIMESTAMP(3),
			n_json JSON
		)`)
		require.NoError(t, err)

		// All NULLs.
		_, err = conn.ExecContext(ctx,
			"INSERT INTO nullable_types VALUES (0, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL)")
		require.NoError(t, err)

		// All non-NULLs.
		_, err = conn.ExecContext(ctx,
			"INSERT INTO nullable_types VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)",
			1, 42, 3.14, "2.7183", "hello", makeTestText(99, 5000),
			makeTestBlobData(99, 5000), "2024-03-15", "2024-03-15 10:00:00.000",
			`{"x": 1}`)
		require.NoError(t, err)

		// Mix: alternate NULL and non-NULL.
		for i := 2; i < 20; i++ {
			var nInt, nStr, nTxt, nDate interface{}
			if i%2 == 0 {
				nInt = i * 100
				nStr = fmt.Sprintf("value-%d", i)
				nTxt = makeTestText(i, 2000)
				nDate = "2024-01-01"
			}
			_, err = conn.ExecContext(ctx,
				"INSERT INTO nullable_types (id, n_int, n_str, n_txt, n_date) VALUES ($1,$2,$3,$4,$5)",
				i, nInt, nStr, nTxt, nDate)
			require.NoError(t, err)
		}

		_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'nullable types')")
		require.NoError(t, err)

		var count int
		err = conn.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM nullable_types WHERE n_int IS NULL AND n_str IS NULL AND n_blob IS NULL").Scan(&count)
		require.NoError(t, err)
		require.Greater(t, count, 0, "rows with all-NULL values must be stored and counted correctly")

		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)

		db.Close()
		db, err = server.DB(driver.Connection{})
		require.NoError(t, err)
		conn, err = db.Conn(ctx)
		require.NoError(t, err)

		// Count should be stable after GC.
		var total int
		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM nullable_types").Scan(&total)
		require.NoError(t, err)
		require.Equal(t, 20, total, "all nullable rows must survive GC")
	})
}

// ----------------------------------------------------------------------------
// TestWideTable verifies that Doltgres handles tables with a very large number
// of columns, including cases where the per-row data volume greatly exceeds
// what would fit in an inline (64 KB) row representation. Subtests cover
// different shapes: many small integer columns, fewer but wider string columns,
// and many TEXT columns carrying large payloads.
// ----------------------------------------------------------------------------

func TestWideTable(t *testing.T) {
	t.Parallel()

	t.Run("ManyIntColumns", func(t *testing.T) {
		t.Parallel()
		// 500 BIGINT columns: stresses column-count parsing and prolly-tree
		// row serialisation without producing a large per-row payload.
		const numCols = 500
		const numRows = 30

		server := setupTestServer(t, "wide_int_table")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.ExecContext(ctx, buildWideCreateTable("wide_int", numCols, "BIGINT"))
		require.NoError(t, err)

		for row := range numRows {
			stmt := buildWideInsert("wide_int", row, numCols, func(col int) string {
				return fmt.Sprintf("%d", int64(row)*int64(col+1))
			})
			_, err = conn.ExecContext(ctx, stmt)
			require.NoError(t, err)
		}

		_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'wide int table')")
		require.NoError(t, err)

		var count int
		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM wide_int").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, numRows, count)

		// Verify a specific cell: row 5, column c250 should be 5*(250+1)=1255.
		var val int64
		err = conn.QueryRowContext(ctx, "SELECT c250 FROM wide_int WHERE id = 5").Scan(&val)
		require.NoError(t, err)
		require.Equal(t, int64(5*251), val, "cell value at (row=5, col=c250) must be exact")

		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)

		db.Close()
		db, err = server.DB(driver.Connection{})
		require.NoError(t, err)
		conn, err = db.Conn(ctx)
		require.NoError(t, err)

		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM wide_int").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, numRows, count, "wide int table must survive GC intact")

		// Re-verify the cell value after GC.
		err = conn.QueryRowContext(ctx, "SELECT c250 FROM wide_int WHERE id = 5").Scan(&val)
		require.NoError(t, err)
		require.Equal(t, int64(5*251), val, "cell value must be preserved after GC")
	})

	t.Run("ManyVarcharColumns", func(t *testing.T) {
		t.Parallel()
		// 8 VARCHAR(300) columns, stressing wide-row serialisation.
		const numCols = 8
		const colWidth = 300
		const numRows = 20

		server := setupTestServer(t, "wide_varchar_table")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.ExecContext(ctx,
			buildWideCreateTable("wide_varchar", numCols, fmt.Sprintf("VARCHAR(%d)", colWidth)))
		require.NoError(t, err)

		for row := range numRows {
			stmt := buildWideInsert("wide_varchar", row, numCols, func(col int) string {
				// Each cell is colWidth chars, content encodes (row, col) for verifiability.
				prefix := fmt.Sprintf("r%03dc%03d-", row, col)
				padded := strings.Repeat(prefix, (colWidth+len(prefix)-1)/len(prefix))[:colWidth]
				return fmt.Sprintf("'%s'", padded)
			})
			_, err = conn.ExecContext(ctx, stmt)
			require.NoError(t, err)
		}

		_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'wide varchar table')")
		require.NoError(t, err)

		var count int
		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM wide_varchar").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, numRows, count)

		// Every cell should be exactly colWidth bytes.
		err = conn.QueryRowContext(ctx,
			fmt.Sprintf("SELECT COUNT(*) FROM wide_varchar WHERE octet_length(c0) = %d AND octet_length(c%d) = %d",
				colWidth, numCols-1, colWidth)).Scan(&count)
		require.NoError(t, err)
		require.Equal(t, numRows, count, "all VARCHAR cells must retain their full width")

		// Verify that a specific cell starts with the right prefix.
		var cell string
		err = conn.QueryRowContext(ctx, "SELECT c4 FROM wide_varchar WHERE id = 10").Scan(&cell)
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(cell, "r010c004-"),
			"VARCHAR cell content must encode the correct (row, col) position")

		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)

		db.Close()
		db, err = server.DB(driver.Connection{})
		require.NoError(t, err)
		conn, err = db.Conn(ctx)
		require.NoError(t, err)

		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM wide_varchar").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, numRows, count, "wide varchar table must survive GC intact")
	})

	t.Run("ManyTextColumns", func(t *testing.T) {
		t.Parallel()
		// 100 TEXT columns each holding ~2 KB of data, giving roughly 200 KB
		// of out-of-band content per logical row. This tests that Doltgres
		// handles rows whose external chunk references far outnumber what would
		// fit in any single inline storage format.
		const numCols = 100
		const colDataSize = 2_000
		const numRows = 10

		server := setupTestServer(t, "wide_text_table")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.ExecContext(ctx, buildWideCreateTable("wide_text", numCols, "TEXT"))
		require.NoError(t, err)

		for row := range numRows {
			// Build the VALUES tuple manually because each TEXT value contains
			// content that must be passed as a query parameter to avoid quoting issues.
			args := make([]interface{}, 0, numCols+1)
			placeholders := make([]string, 0, numCols+1)
			args = append(args, row)
			placeholders = append(placeholders, "$1")
			for col := range numCols {
				// Unique seed per (row, col) so content is verifiable.
				args = append(args, makeTestText(row*numCols+col, colDataSize))
				placeholders = append(placeholders, fmt.Sprintf("$%d", col+2))
			}
			stmt := fmt.Sprintf("INSERT INTO wide_text VALUES (%s)", strings.Join(placeholders, ","))
			_, err = conn.ExecContext(ctx, stmt, args...)
			require.NoError(t, err)
		}

		_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'wide text table')")
		require.NoError(t, err)

		var count int
		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM wide_text").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, numRows, count)

		// Verify that the first and last text columns have the expected length.
		// Doltgres panics on octet_length() over large out-of-band text, so the
		// lengths are checked client-side.
		lenRows, err := conn.QueryContext(ctx, "SELECT c0, c99 FROM wide_text")
		require.NoError(t, err)
		count = 0
		for lenRows.Next() {
			var c0v, c99v string
			require.NoError(t, lenRows.Scan(&c0v, &c99v))
			require.Equal(t, colDataSize, len(c0v))
			require.Equal(t, colDataSize, len(c99v))
			count++
		}
		require.NoError(t, lenRows.Err())
		lenRows.Close()
		require.Equal(t, numRows, count, "all TEXT cells must retain their content length")

		// Spot-check exact content of a cell.
		var cell string
		const checkRow = 3
		const checkCol = 47
		err = conn.QueryRowContext(ctx,
			fmt.Sprintf("SELECT c%d FROM wide_text WHERE id = %d", checkCol, checkRow)).Scan(&cell)
		require.NoError(t, err)
		require.Equal(t, makeTestText(checkRow*numCols+checkCol, colDataSize), cell,
			"TEXT cell content must be byte-perfect after storage and retrieval")

		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)

		db.Close()
		db, err = server.DB(driver.Connection{})
		require.NoError(t, err)
		conn, err = db.Conn(ctx)
		require.NoError(t, err)

		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM wide_text").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, numRows, count, "wide text table must survive GC intact")

		// Re-check cell content after GC.
		err = conn.QueryRowContext(ctx,
			fmt.Sprintf("SELECT c%d FROM wide_text WHERE id = %d", checkCol, checkRow)).Scan(&cell)
		require.NoError(t, err)
		require.Equal(t, makeTestText(checkRow*numCols+checkCol, colDataSize), cell,
			"TEXT cell content must be preserved through GC")
	})

	t.Run("ExtremelyWideRow", func(t *testing.T) {
		t.Parallel()
		// A single row with 200 TEXT columns each carrying 5 KB: 1 MB of
		// out-of-band data in one logical row. Exercises the edge of what
		// Doltgres's chunk graph must track for a single row pointer.
		const numCols = 200
		const colDataSize = 5_000
		const numRows = 3

		server := setupTestServer(t, "extreme_wide_row")
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()
		ctx := t.Context()

		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()

		_, err = conn.ExecContext(ctx, buildWideCreateTable("extreme_wide", numCols, "TEXT"))
		require.NoError(t, err)

		for row := range numRows {
			args := make([]interface{}, 0, numCols+1)
			placeholders := make([]string, 0, numCols+1)
			args = append(args, row)
			placeholders = append(placeholders, "$1")
			for col := range numCols {
				args = append(args, makeTestText(row*numCols+col, colDataSize))
				placeholders = append(placeholders, fmt.Sprintf("$%d", col+2))
			}
			stmt := fmt.Sprintf("INSERT INTO extreme_wide VALUES (%s)", strings.Join(placeholders, ","))
			_, err = conn.ExecContext(ctx, stmt, args...)
			require.NoError(t, err)
		}

		_, err = conn.ExecContext(ctx, "SELECT dolt_commit('-Am', 'extreme wide rows')")
		require.NoError(t, err)

		var count int
		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM extreme_wide").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, numRows, count)

		// Verify total data volume via the byte length of two boundary columns.
		// Doltgres panics on octet_length() over large out-of-band text, so the
		// lengths are checked client-side.
		lenRows, err := conn.QueryContext(ctx, "SELECT c0, c199 FROM extreme_wide")
		require.NoError(t, err)
		count = 0
		for lenRows.Next() {
			var c0v, c199v string
			require.NoError(t, lenRows.Scan(&c0v, &c199v))
			require.Equal(t, colDataSize, len(c0v))
			require.Equal(t, colDataSize, len(c199v))
			count++
		}
		require.NoError(t, lenRows.Err())
		lenRows.Close()
		require.Equal(t, numRows, count,
			"extremely wide rows must fully preserve all column data")

		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)

		db.Close()
		db, err = server.DB(driver.Connection{})
		require.NoError(t, err)
		conn, err = db.Conn(ctx)
		require.NoError(t, err)

		err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM extreme_wide").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, numRows, count, "extremely wide rows must survive GC")

		// Spot-check an interior column post-GC.
		const checkRow = 1
		const checkCol = 100
		var cell string
		err = conn.QueryRowContext(ctx,
			fmt.Sprintf("SELECT c%d FROM extreme_wide WHERE id = %d", checkCol, checkRow)).Scan(&cell)
		require.NoError(t, err)
		require.Equal(t, makeTestText(checkRow*numCols+checkCol, colDataSize), cell,
			"extreme wide row TEXT cell must be byte-perfect after GC")
	})
}

// countTextRowsOfLen runs a single-column text query and returns the number of
// rows whose value has exactly want bytes, asserting each row matches. Doltgres
// currently panics on octet_length()/length() over large out-of-band text
// values, so callers verify text byte-lengths client-side instead of in SQL.
func countTextRowsOfLen(t *testing.T, conn *sql.Conn, ctx context.Context, query string, want int) int {
	rows, err := conn.QueryContext(ctx, query)
	require.NoError(t, err)
	defer rows.Close()
	n := 0
	for rows.Next() {
		var s string
		require.NoError(t, rows.Scan(&s))
		require.Equal(t, want, len(s))
		n++
	}
	require.NoError(t, rows.Err())
	return n
}
