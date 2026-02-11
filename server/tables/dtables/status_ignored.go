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

package dtables

import (
	"fmt"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/adapters"
	doltdtables "github.com/dolthub/dolt/go/libraries/doltcore/sqle/dtables"
	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// DoltgresDoltStatusIgnoredTableAdapter adapts the [doltdtables.StatusIgnoredTable] into a Doltgres-compatible version.
//
// DoltgresDoltStatusIgnoredTableAdapter implements the [adapters.TableAdapter] interface.
type DoltgresDoltStatusIgnoredTableAdapter struct{}

var _ adapters.TableAdapter = DoltgresDoltStatusIgnoredTableAdapter{}

// NewTable returns a new [sql.Table] for Doltgres' version of [doltdtables.StatusIgnoredTable].
func (a DoltgresDoltStatusIgnoredTableAdapter) NewTable(ctx *sql.Context, tableName string, ddb *doltdb.DoltDB, ws *doltdb.WorkingSet, rp env.RootsProvider[*sql.Context]) sql.Table {
	doltTable := doltdtables.NewStatusIgnoredTableWithNoAdapter(ctx, tableName, ddb, ws, rp)
	return &doltgresDoltStatusIgnoredTable{
		srcDoltStatusIgnored: doltTable.(*doltdtables.StatusIgnoredTable),
	}
}

// TableName returns the table name for Doltgres' version of [doltdtables.StatusIgnoredTable].
func (a DoltgresDoltStatusIgnoredTableAdapter) TableName() string {
	return DoltgresDoltStatusIgnoredTableName
}

// DoltgresDoltStatusIgnoredTableName is the name of Dolt's status_ignored table following Doltgres' naming conventions.
const DoltgresDoltStatusIgnoredTableName = "status_ignored"

// doltgresDoltStatusIgnoredTable translates the [doltdtables.StatusIgnoredTable] into a Doltgres-compatible version.
//
// doltgresDoltStatusIgnoredTable implements the [sql.Table] and [sql.StatisticsTable] interfaces.
type doltgresDoltStatusIgnoredTable struct {
	srcDoltStatusIgnored *doltdtables.StatusIgnoredTable
}

var _ sql.Table = (*doltgresDoltStatusIgnoredTable)(nil)
var _ sql.StatisticsTable = (*doltgresDoltStatusIgnoredTable)(nil)

// Name returns the name of Doltgres' version of the Dolt status_ignored table.
func (w *doltgresDoltStatusIgnoredTable) Name() string {
	return w.srcDoltStatusIgnored.Name()
}

// Schema returns the schema for Doltgres' version of the Dolt status_ignored table.
func (w *doltgresDoltStatusIgnoredTable) Schema() sql.Schema {
	return []*sql.Column{
		{Name: "table_name", Type: pgtypes.Text, Source: DoltgresDoltStatusIgnoredTableName, PrimaryKey: true, Nullable: false},
		{Name: "staged", Type: pgtypes.Bool, Source: DoltgresDoltStatusIgnoredTableName, PrimaryKey: true, Nullable: false},
		{Name: "status", Type: pgtypes.Text, Source: DoltgresDoltStatusIgnoredTableName, PrimaryKey: true, Nullable: false},
		{Name: "ignored", Type: pgtypes.Bool, Source: DoltgresDoltStatusIgnoredTableName, PrimaryKey: false, Nullable: false},
	}
}

// String returns the string representation of [doltdtables.StatusIgnoredTable].
func (w *doltgresDoltStatusIgnoredTable) String() string {
	return w.srcDoltStatusIgnored.String()
}

// Collation returns the [sql.CollationID] from [doltdtables.StatusIgnoredTable].
func (w *doltgresDoltStatusIgnoredTable) Collation() sql.CollationID {
	return w.srcDoltStatusIgnored.Collation()
}

// Partitions returns a [sql.PartitionIter] on the partitions of [doltdtables.StatusIgnoredTable].
func (w *doltgresDoltStatusIgnoredTable) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	return w.srcDoltStatusIgnored.Partitions(ctx)
}

// PartitionRows returns a wrapped [sql.RowIter] for the rows in |partition| from
// [doltdtables.StatusIgnoredTable.PartitionRows] to later apply column transformations that match Doltgres' version of the
// Dolt status_ignored table schema.
func (w *doltgresDoltStatusIgnoredTable) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	iter, err := w.srcDoltStatusIgnored.PartitionRows(ctx, partition)
	if err != nil {
		return nil, err
	}
	return &doltgresDoltStatusIgnoredRowIter{w, iter}, nil
}

// DataLength returns the length of the data in bytes from [doltdtables.StatusIgnoredTable].
func (w *doltgresDoltStatusIgnoredTable) DataLength(ctx *sql.Context) (uint64, error) {
	return w.srcDoltStatusIgnored.DataLength(ctx)
}

// RowCount returns exact (true) or estimate (false) number of rows from [doltdtables.StatusIgnoredTable].
func (w *doltgresDoltStatusIgnoredTable) RowCount(ctx *sql.Context) (uint64, bool, error) {
	return w.srcDoltStatusIgnored.RowCount(ctx)
}

// doltgresDoltStatusIgnoredRowIter wraps [doltdtables.StatusIgnoredTable] [sql.RowIter] and applies transformations before returning
// its rows to make sure they're compatible with Doltgres' version of Dolt's status_ignored table.
type doltgresDoltStatusIgnoredRowIter struct {
	doltStatusIgnoredTable sql.Table
	rowIter                sql.RowIter
}

var _ sql.RowIter = (*doltgresDoltStatusIgnoredRowIter)(nil)

// Next converts the 'staged' column from [doltdtables.StatusIgnoredTable.Schema] from byte into bool since,
// unlike the MySQL wire protocol, Doltgres has a real bool type.
func (i *doltgresDoltStatusIgnoredRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	row, err := i.rowIter.Next(ctx)
	if err != nil {
		return nil, err
	}

	// Dolt uses byte to avoid MySQL wire protocol ambiguity on tinyint(1) and bool.
	// See: https://github.com/dolthub/dolt/pull/10117
	stagedIndex := i.doltStatusIgnoredTable.Schema().IndexOfColName("staged")
	stagedVal, ok := row[stagedIndex].(byte)
	if !ok {
		return nil, fmt.Errorf("expected staged column at index %d to be byte, got %T", stagedIndex, row[stagedIndex])
	}
	row[stagedIndex] = stagedVal != 0

	return row, nil
}

// Close closes the wrapped [doltdtables.StatusIgnoredTable] [sql.RowIter].
func (i *doltgresDoltStatusIgnoredRowIter) Close(ctx *sql.Context) error {
	return i.rowIter.Close(ctx)
}
