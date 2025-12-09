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

package dtables

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/adapters"
	doltdtables "github.com/dolthub/dolt/go/libraries/doltcore/sqle/dtables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// DoltgresDoltStatusTableAdapter adapts the [doltdtables.StatusTable] into a Doltgres-compatible version.
//
// DoltgresDoltStatusTableAdapter implements the [adapters.TableAdapter] interface.
type DoltgresDoltStatusTableAdapter struct{}

var _ adapters.TableAdapter = DoltgresDoltStatusTableAdapter{}

// NewTable returns a new [sql.Table] for Doltgres' version of [doltdtables.StatusTable].
func (a DoltgresDoltStatusTableAdapter) NewTable(ctx *sql.Context, tableName string, ddb *doltdb.DoltDB, ws *doltdb.WorkingSet, rp env.RootsProvider[*sql.Context]) sql.Table {
	doltTable := doltdtables.NewStatusTableWithNoAdapter(ctx, tableName, ddb, ws, rp)
	return &doltgresDoltStatusTable{
		srcDoltStatus: doltTable.(*doltdtables.StatusTable),
	}
}

// TableName returns the table name for Doltgres' version of [doltdtables.StatusTable].
func (a DoltgresDoltStatusTableAdapter) TableName() string {
	return DoltgresDoltStatusTableName
}

// DoltgresDoltStatusTableName is the name of Dolt's status table following Doltgres' naming conventions.
const DoltgresDoltStatusTableName = "status"

// doltgresDoltStatusTable translates the [doltdtables.StatusTable] into a Doltgres-compatible version.
//
// doltgresDoltStatusTable implements the [sql.Table] and [sql.StatisticsTable] interfaces.
type doltgresDoltStatusTable struct {
	srcDoltStatus *doltdtables.StatusTable
}

var _ sql.Table = (*doltgresDoltStatusTable)(nil)
var _ sql.StatisticsTable = (*doltgresDoltStatusTable)(nil)

// Name returns the name of Doltgres' version of the Dolt status table.
func (w *doltgresDoltStatusTable) Name() string {
	return w.srcDoltStatus.Name()
}

// Schema returns the schema for Doltgres' version of the Dolt status table.
func (w *doltgresDoltStatusTable) Schema() sql.Schema {
	return []*sql.Column{
		{Name: "table_name", Type: pgtypes.Text, Source: DoltgresDoltStatusTableName, PrimaryKey: true, Nullable: false},
		{Name: "staged", Type: pgtypes.Bool, Source: DoltgresDoltStatusTableName, PrimaryKey: true, Nullable: false},
		{Name: "status", Type: pgtypes.Text, Source: DoltgresDoltStatusTableName, PrimaryKey: true, Nullable: false},
	}
}

// String returns the string representation of [doltdtables.StatusTable].
func (w *doltgresDoltStatusTable) String() string {
	return w.srcDoltStatus.String()
}

// Collation returns the [sql.CollationID] from [doltdtables.StatusTable].
func (w *doltgresDoltStatusTable) Collation() sql.CollationID {
	return w.srcDoltStatus.Collation()
}

// Partitions returns a [sql.PartitionIter] on the partitions of [doltdtables.StatusTable].
func (w *doltgresDoltStatusTable) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	return w.srcDoltStatus.Partitions(ctx)
}

// PartitionRows returns a wrapped [sql.RowIter] for the rows in |partition| from
// [doltdtables.StatusTable.PartitionRows] to later apply column transformations that match Doltgres' version of the
// Dolt status table schema.
func (w *doltgresDoltStatusTable) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	iter, err := w.srcDoltStatus.PartitionRows(ctx, partition)
	if err != nil {
		return nil, err
	}
	return &doltgresDoltStatusRowIter{w, iter}, nil
}

// DataLength returns the length of the data in bytes from [doltdtables.StatusTable].
func (w *doltgresDoltStatusTable) DataLength(ctx *sql.Context) (uint64, error) {
	return w.srcDoltStatus.DataLength(ctx)
}

// RowCount returns exact (true) or estimate (false) number of rows from [doltdtables.StatusTable].
func (w *doltgresDoltStatusTable) RowCount(ctx *sql.Context) (uint64, bool, error) {
	return w.srcDoltStatus.RowCount(ctx)
}

// doltgresDoltStatusRowIter wraps [doltdtables.StatusTable] [sql.RowIter] and applies transformations before returning
// its rows to make sure they're compatible with Doltgres' version of Dolt's status table.
type doltgresDoltStatusRowIter struct {
	doltStatusTable sql.Table
	rowIter         sql.RowIter
}

var _ sql.RowIter = (*doltgresDoltStatusRowIter)(nil)

// Next converts the 'staged' column from [doltdtables.StatusTable.Schema] from a byte into a bool since, unlike the
// MySQL wire protocol, Doltgres has a real bool type.
func (i *doltgresDoltStatusRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	row, err := i.rowIter.Next(ctx)
	if err != nil {
		return nil, err
	}

	// Dolt uses byte to avoid MySQL wire protocol ambiguity on tinyint(1) and bool.
	// See: https://github.com/dolthub/dolt/pull/10117
	stagedIndex := i.doltStatusTable.Schema().IndexOfColName("staged")
	stagedVal, ok := row[stagedIndex].(byte)
	if !ok {
		return nil, fmt.Errorf("expected staged column at index %d to be byte, got %T", stagedIndex, row[stagedIndex])
	}
	row[stagedIndex] = stagedVal != 0

	return row, nil
}

// Close closes the wrapped [doltdtables.StatusTable] [sql.RowIter].
func (i *doltgresDoltStatusRowIter) Close(ctx *sql.Context) error {
	return i.rowIter.Close(ctx)
}
