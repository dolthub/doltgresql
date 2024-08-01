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

// PgProcName is a constant to the pg_proc name.
const PgProcName = "pg_proc"

// InitPgProc handles registration of the pg_proc handler.
func InitPgProc() {
	tables.AddHandler(PgCatalogName, PgProcName, PgProcHandler{})
}

// PgProcHandler is the handler for the pg_proc table.
type PgProcHandler struct{}

var _ tables.Handler = PgProcHandler{}

// Name implements the interface tables.Handler.
func (p PgProcHandler) Name() string {
	return PgProcName
}

// RowIter implements the interface tables.Handler.
func (p PgProcHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_proc row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgProcHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgProcSchema,
		PkOrdinals: nil,
	}
}

// pgProcSchema is the schema for pg_proc.
var pgProcSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "proname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "pronamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "proowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "prolang", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "procost", Type: pgtypes.Float32, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "prorows", Type: pgtypes.Float32, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "provariadic", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "prosupport", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgProcName}, // TODO: type regproc
	{Name: "prokind", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "prosecdef", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "proleakproof", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "proisstrict", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "proretset", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "provolatile", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "proparallel", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "pronargs", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "pronargdefaults", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "prorettype", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgProcName},
	{Name: "proargtypes", Type: pgtypes.OidArray, Default: nil, Nullable: false, Source: PgProcName}, // TODO: type oidvector
	{Name: "proallargtypes", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgProcName},
	{Name: "proargmodes", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgProcName}, // TODO: type char[]
	{Name: "proargnames", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgProcName}, // TODO: collation C
	{Name: "proargdefaults", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgProcName},   // TODO: type pg_node_tree, collation C
	{Name: "protrftypes", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgProcName},
	{Name: "prosrc", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgProcName}, // TODO: collation C
	{Name: "probin", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgProcName},
	{Name: "prosqlbody", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgProcName},     // TODO: type pg_node_tree, collation C
	{Name: "proconfig", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgProcName}, // TODO: collation C
	{Name: "proacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgProcName},    // TODO: type aclitem[]
}

// pgProcRowIter is the sql.RowIter for the pg_proc table.
type pgProcRowIter struct {
}

var _ sql.RowIter = (*pgProcRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgProcRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgProcRowIter) Close(ctx *sql.Context) error {
	return nil
}
