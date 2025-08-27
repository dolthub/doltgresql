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

package pgcatalog

import (
	"io"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgPartitionedTableName is a constant to the pg_partitioned_table name.
const PgPartitionedTableName = "pg_partitioned_table"

// InitPgPartitionedTable handles registration of the pg_partitioned_table handler.
func InitPgPartitionedTable() {
	tables.AddHandler(PgCatalogName, PgPartitionedTableName, PgPartitionedTableHandler{})
}

// PgPartitionedTableHandler is the handler for the pg_partitioned_table table.
type PgPartitionedTableHandler struct{}

var _ tables.Handler = PgPartitionedTableHandler{}

// Name implements the interface tables.Handler.
func (p PgPartitionedTableHandler) Name() string {
	return PgPartitionedTableName
}

// RowIter implements the interface tables.Handler.
func (p PgPartitionedTableHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_partitioned_table row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgPartitionedTableHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgPartitionedTableSchema,
		PkOrdinals: nil,
	}
}

// pgPartitionedTableSchema is the schema for pg_partitioned_table.
var pgPartitionedTableSchema = sql.Schema{
	{Name: "partrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgPartitionedTableName},
	{Name: "partstrat", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgPartitionedTableName},
	{Name: "partnatts", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgPartitionedTableName},
	{Name: "partdefid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgPartitionedTableName},
	{Name: "partattrs", Type: pgtypes.Int16Array, Default: nil, Nullable: false, Source: PgPartitionedTableName},   // TODO: int2vector type
	{Name: "partclass", Type: pgtypes.OidArray, Default: nil, Nullable: false, Source: PgPartitionedTableName},     // TODO: oidvector type
	{Name: "partcollation", Type: pgtypes.OidArray, Default: nil, Nullable: false, Source: PgPartitionedTableName}, // TODO: oidvector type
	{Name: "partexprs", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgPartitionedTableName},          // TODO: pg_node_tree type, collation C
}

// pgPartitionedTableRowIter is the sql.RowIter for the pg_partitioned_table table.
type pgPartitionedTableRowIter struct {
}

var _ sql.RowIter = (*pgPartitionedTableRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgPartitionedTableRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgPartitionedTableRowIter) Close(ctx *sql.Context) error {
	return nil
}
