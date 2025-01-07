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
	"testing"

	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/core/dataloader"
	"github.com/dolthub/doltgresql/server/initialization"
	"github.com/dolthub/doltgresql/server/types"
)

func TestTabDataLoader(t *testing.T) {
	db := memory.NewDatabase("mydb")
	provider := memory.NewDBProvider(db)
	initialization.Initialize(nil)

	ctx := &sql.Context{
		Context: context.Background(),
		Session: memory.NewSession(sql.NewBaseSession(), provider),
	}

	pkCols := []string{"pk", "c1", "c2"}
	pkSchema := sql.NewPrimaryKeySchema(sql.Schema{
		{Name: "pk", Type: types.Int64, Source: "source1"},
		{Name: "c1", Type: types.Int64, Source: "source1"},
		{Name: "c2", Type: types.VarChar, Source: "source1"},
	}, 0)

	// Tests that a basic tab delimited doc can be loaded as a single chunk.
	t.Run("basic case", func(t *testing.T) {
		dataLoader, err := dataloader.NewTabularDataLoader(pkCols, pkSchema.Schema, "\t", "\\N", false)
		require.NoError(t, err)

		var rows []sql.Row

		// Load all the data as a single chunk
		reader := bytes.NewReader([]byte("1\t100\tbar\n2\t200\tbash\n"))
		err = dataLoader.SetNextDataChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)
		rows = append(rows, loadAllRows(ctx, t, dataLoader)...)

		results, err := dataLoader.Finish(ctx)
		require.NoError(t, err)
		require.EqualValues(t, 2, results.RowsLoaded)

		assert.Equal(t, []sql.Row{
			{int64(1), int64(100), "bar"},
			{int64(2), int64(200), "bash"},
		}, rows)
	})

	// Tests when a record is split across two chunks of data, and the
	// partial record must be buffered and prepended to the next chunk.
	t.Run("record split across two chunks", func(t *testing.T) {
		dataLoader, err := dataloader.NewTabularDataLoader(pkCols, pkSchema.Schema, "\t", "\\N", false)
		require.NoError(t, err)

		var rows []sql.Row

		// Load the first chunk
		reader := bytes.NewReader([]byte("1	100	ba"))
		err = dataLoader.SetNextDataChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)
		rows = append(rows, loadAllRows(ctx, t, dataLoader)...)

		// Load the second chunk
		reader = bytes.NewReader([]byte("r\n2	200	bash\n"))
		err = dataLoader.SetNextDataChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)
		rows = append(rows, loadAllRows(ctx, t, dataLoader)...)

		// Finish
		results, err := dataLoader.Finish(ctx)
		require.NoError(t, err)
		require.EqualValues(t, 2, results.RowsLoaded)

		assert.Equal(t, []sql.Row{
			{int64(1), int64(100), "bar"},
			{int64(2), int64(200), "bash"},
		}, rows)
	})

	// Tests when a record is split across two chunks of data, and a
	// header row is present.
	t.Run("record split across two chunks, with header", func(t *testing.T) {
		dataLoader, err := dataloader.NewTabularDataLoader(pkCols, pkSchema.Schema, "\t", "\\N", true)
		require.NoError(t, err)

		var rows []sql.Row

		// Load the first chunk
		reader := bytes.NewReader([]byte("pk	c1	c2\n1	100	ba"))
		err = dataLoader.SetNextDataChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)
		rows = append(rows, loadAllRows(ctx, t, dataLoader)...)

		// Load the second chunk
		reader = bytes.NewReader([]byte("r\n2	200	bash\n"))
		err = dataLoader.SetNextDataChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)
		rows = append(rows, loadAllRows(ctx, t, dataLoader)...)

		// Finish
		results, err := dataLoader.Finish(ctx)
		require.NoError(t, err)
		require.EqualValues(t, 2, results.RowsLoaded)

		// Assert that the table contains the expected data
		assert.Equal(t, []sql.Row{
			{int64(1), int64(100), "bar"},
			{int64(2), int64(200), "bash"},
		}, rows)
	})

	// Tests a record that contains a quoted newline character and is split
	// across two chunks.
	t.Run("quoted newlines across two chunks", func(t *testing.T) {
		dataLoader, err := dataloader.NewTabularDataLoader(pkCols, pkSchema.Schema, "\t", "\\N", false)
		require.NoError(t, err)

		var rows []sql.Row

		// Load the first chunk
		reader := bytes.NewReader([]byte("1	100	\"baz\\nbar\\n"))
		err = dataLoader.SetNextDataChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)
		rows = append(rows, loadAllRows(ctx, t, dataLoader)...)

		// Load the second chunk
		reader = bytes.NewReader([]byte("bash\"\n2	200	bash\n"))
		err = dataLoader.SetNextDataChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)
		rows = append(rows, loadAllRows(ctx, t, dataLoader)...)

		// Finish
		results, err := dataLoader.Finish(ctx)
		require.NoError(t, err)
		require.EqualValues(t, 2, results.RowsLoaded)

		// Assert that the table contains the expected data
		assert.Equal(t, []sql.Row{
			{int64(1), int64(100), "\"baz\\nbar\\nbash\""},
			{int64(2), int64(200), "bash"},
		}, rows)
	})

	// Tests when a record is split across two chunks of data, and a
	// header row is present.
	t.Run("delimiter='|', record split across two chunks, with header", func(t *testing.T) {
		dataLoader, err := dataloader.NewTabularDataLoader(pkCols, pkSchema.Schema, "|", "\\N", true)
		require.NoError(t, err)

		var rows []sql.Row

		// Load the first chunk
		reader := bytes.NewReader([]byte("pk|c1|c2\n1|100|ba"))
		err = dataLoader.SetNextDataChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)
		rows = append(rows, loadAllRows(ctx, t, dataLoader)...)

		// Load the second chunk
		reader = bytes.NewReader([]byte("r\n2|200|bash\n"))
		err = dataLoader.SetNextDataChunk(ctx, bufio.NewReader(reader))
		require.NoError(t, err)
		rows = append(rows, loadAllRows(ctx, t, dataLoader)...)

		// Finish
		results, err := dataLoader.Finish(ctx)
		require.NoError(t, err)
		require.EqualValues(t, 2, results.RowsLoaded)

		// Assert that the table contains the expected data
		assert.Equal(t, []sql.Row{
			{int64(1), int64(100), "bar"},
			{int64(2), int64(200), "bash"},
		}, rows)
	})
}
