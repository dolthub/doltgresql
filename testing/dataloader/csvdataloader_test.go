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

package _dataloader

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/core/dataloader"
	"github.com/dolthub/doltgresql/server/initialization"
	"github.com/dolthub/doltgresql/server/types"
)

// TestCsvDataLoader tests the CsvDataLoader implementation.
func TestCsvDataLoader(t *testing.T) {
	db := memory.NewDatabase("mydb")
	provider := memory.NewDBProvider(db)
	initialization.Initialize(nil)

	ctx := &sql.Context{
		Context: context.Background(),
		Session: memory.NewSession(sql.NewBaseSession(), provider),
	}

	pkSchema := sql.NewPrimaryKeySchema(sql.Schema{
		{Name: "pk", Type: types.Int64, Source: "source1"},
		{Name: "c1", Type: types.Int64, Source: "source1"},
		{Name: "c2", Type: types.VarChar, Source: "source1"},
	}, 0)

	// Tests that a basic CSV document can be loaded as a single chunk.
	t.Run("basic case", func(t *testing.T) {
		table := memory.NewTable(db, "myTable", pkSchema, nil)
		dataLoader, err := dataloader.NewCsvDataLoader(nil, table, ",", false)
		require.NoError(t, err)

		// Load all the data as a single chunk
		reader := bytes.NewReader([]byte("1,100,bar\n2,200,bash\n"))
		err = dataLoader.LoadChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)
		results, err := dataLoader.Finish(ctx)
		require.NoError(t, err)
		require.EqualValues(t, 2, results.RowsLoaded)

		// Assert that the table contains the expected data
		assertRows(t, ctx, table, [][]any{
			{int64(1), int64(100), "bar"},
			{int64(2), int64(200), "bash"},
		})
	})

	// Tests when a CSV record is split across two chunks of data, and the
	// partial record must be buffered and prepended to the next chunk.
	t.Run("record split across two chunks", func(t *testing.T) {
		table := memory.NewTable(db, "myTable", pkSchema, nil)
		dataLoader, err := dataloader.NewCsvDataLoader(nil, table, ",", false)
		require.NoError(t, err)

		// Load the first chunk
		reader := bytes.NewReader([]byte("1,100,ba"))
		err = dataLoader.LoadChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)

		// Load the second chunk
		reader = bytes.NewReader([]byte("r\n2,200,bash\n"))
		err = dataLoader.LoadChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)

		// Finish
		results, err := dataLoader.Finish(ctx)
		require.NoError(t, err)
		require.EqualValues(t, 2, results.RowsLoaded)

		// Assert that the table contains the expected data
		assertRows(t, ctx, table, [][]any{
			{int64(1), int64(100), "bar"},
			{int64(2), int64(200), "bash"},
		})
	})

	// Tests when a CSV record is split across two chunks of data, and a
	// header row is present.
	t.Run("record split across two chunks, with header", func(t *testing.T) {
		table := memory.NewTable(db, "myTable", pkSchema, nil)
		dataLoader, err := dataloader.NewCsvDataLoader(nil, table, ",", true)
		require.NoError(t, err)

		// Load the first chunk
		reader := bytes.NewReader([]byte("pk,c1,c2\n1,100,ba"))
		err = dataLoader.LoadChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)

		// Load the second chunk
		reader = bytes.NewReader([]byte("r\n2,200,bash\n"))
		err = dataLoader.LoadChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)

		// Finish
		results, err := dataLoader.Finish(ctx)
		require.NoError(t, err)
		require.EqualValues(t, 2, results.RowsLoaded)

		// Assert that the table contains the expected data
		assertRows(t, ctx, table, [][]any{
			{int64(1), int64(100), "bar"},
			{int64(2), int64(200), "bash"},
		})
	})

	// Tests a CSV record that contains a quoted newline character and is split
	// across two chunks.
	t.Run("quoted newlines across two chunks", func(t *testing.T) {
		table := memory.NewTable(db, "myTable", pkSchema, nil)
		dataLoader, err := dataloader.NewCsvDataLoader(nil, table, ",", false)
		require.NoError(t, err)

		// Load the first chunk
		reader := bytes.NewReader([]byte("1,100,\"baz\nbar\n"))
		err = dataLoader.LoadChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)

		// Load the second chunk
		reader = bytes.NewReader([]byte("bash\"\n2,200,bash\n"))
		err = dataLoader.LoadChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)

		// Finish
		results, err := dataLoader.Finish(ctx)
		require.NoError(t, err)
		require.EqualValues(t, 2, results.RowsLoaded)

		// Assert that the table contains the expected data
		assertRows(t, ctx, table, [][]any{
			{int64(1), int64(100), "baz\nbar\nbash"},
			{int64(2), int64(200), "bash"},
		})
	})

	// Test that calling Abort() does not insert any data into the table.
	t.Run("abort cancels data load", func(t *testing.T) {
		table := memory.NewTable(db, "myTable", pkSchema, nil)
		dataLoader, err := dataloader.NewCsvDataLoader(nil, table, ",", false)
		require.NoError(t, err)

		// Load the first chunk
		reader := bytes.NewReader([]byte("1,100,bazbar\n"))
		err = dataLoader.LoadChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)

		// Load the second chunk
		reader = bytes.NewReader([]byte("2,200,bash\n"))
		err = dataLoader.LoadChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)

		// Abort
		err = dataLoader.Abort(ctx)
		require.NoError(t, err)

		// Assert that the table does not contain any of the data from the CSV load
		assertRows(t, ctx, table, [][]any{})
	})

	// Tests when a PSV (i.e. delimiter='|') record is split across two chunks of data,
	// and a header row is present.
	t.Run("delimiter='|', record split across two chunks, with header", func(t *testing.T) {
		table := memory.NewTable(db, "myTable", pkSchema, nil)
		dataLoader, err := dataloader.NewCsvDataLoader(nil, table, "|", true)
		require.NoError(t, err)

		// Load the first chunk
		reader := bytes.NewReader([]byte("pk|c1|c2\n1|100|ba"))
		err = dataLoader.LoadChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)

		// Load the second chunk
		reader = bytes.NewReader([]byte("r\n2|200|bash\n"))
		err = dataLoader.LoadChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)

		// Finish
		results, err := dataLoader.Finish(ctx)
		require.NoError(t, err)
		require.EqualValues(t, 2, results.RowsLoaded)

		// Assert that the table contains the expected data
		assertRows(t, ctx, table, [][]any{
			{int64(1), int64(100), "bar"},
			{int64(2), int64(200), "bash"},
		})
	})
}

// assertRows asserts that the rows in |table| match |expectedRows| and fails the test if the
// rows do not exactly match.
func assertRows(t *testing.T, ctx *sql.Context, table *memory.Table, expectedRows [][]any) {
	partitions, err := table.Partitions(ctx)
	require.NoError(t, err)

	expectedRowsIdx := 0

	for {
		partition, err := partitions.Next(ctx)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		rows := table.GetPartition(string(partition.Key()))
		for _, row := range rows {
			if len(expectedRows) <= expectedRowsIdx {
				t.Fatalf("Expected %d rows, got more", len(expectedRows))
			}

			if len(expectedRows[expectedRowsIdx]) != len(row) {
				t.Fatalf("Expected row length %d, got %d. expectedRows: %v, rows: %v",
					len(expectedRows), len(row), expectedRows, rows)
			}
			for i := range len(row) {
				if expectedRows[expectedRowsIdx][i] != row[i] {
					t.Fatalf("Expected row %v, got %v. expectedRows: %v, rows: %v",
						expectedRows[expectedRowsIdx], row, expectedRows, rows)
				}
			}

			expectedRowsIdx += 1
		}
	}

	if len(expectedRows) != expectedRowsIdx {
		t.Fatalf("Expected %d rows, got %d", len(expectedRows), expectedRowsIdx)
	}
}
