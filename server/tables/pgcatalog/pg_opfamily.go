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

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgOpfamilyName is a constant to the pg_opfamily name.
const PgOpfamilyName = "pg_opfamily"

// InitPgOpfamily handles registration of the pg_opfamily handler.
func InitPgOpfamily() {
	tables.AddHandler(PgCatalogName, PgOpfamilyName, PgOpfamilyHandler{})
}

// PgOpfamilyHandler is the handler for the pg_opfamily table.
type PgOpfamilyHandler struct{}

var _ tables.Handler = PgOpfamilyHandler{}

// Name implements the interface tables.Handler.
func (p PgOpfamilyHandler) Name() string {
	return PgOpfamilyName
}

// RowIter implements the interface tables.Handler.
func (p PgOpfamilyHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	return &pgOpfamilyRowIter{
		families: defaultOperatorFamilies,
		idx:      0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgOpfamilyHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgOpfamilySchema,
		PkOrdinals: nil,
	}
}

// pgOpfamilySchema is the schema for pg_opfamily.
var pgOpfamilySchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOpfamilyName},
	{Name: "opfmethod", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOpfamilyName},
	{Name: "opfname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgOpfamilyName},
	{Name: "opfnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOpfamilyName},
	{Name: "opfowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOpfamilyName},
}

// operatorFamily describes a built-in operator family.
type operatorFamily struct {
	am   string
	name string
}

// oid returns the ID of this operator family, whose OIDs are registered in core/id/cache_operator_class_defaults.go.
func (f operatorFamily) oid() id.Id {
	return id.NewId(id.Section_OperatorFamily, f.am, f.name)
}

// defaultOperatorFamilies is the list of built-in operator families available in Postgres for the access methods and
// types that Doltgres supports. Their OIDs are registered in core/id/cache_operator_class_defaults.go, and match the
// fixed operator family OIDs assigned by Postgres.
// TODO: Postgres defines more operator families (gin, gist, brin, spgist, and additional types); add them as the
// related types and access methods gain support.
var defaultOperatorFamilies = []operatorFamily{
	{am: "btree", name: "array_ops"},
	{am: "btree", name: "bit_ops"},
	{am: "btree", name: "bool_ops"},
	{am: "btree", name: "bpchar_ops"},
	{am: "btree", name: "bytea_ops"},
	{am: "btree", name: "char_ops"},
	{am: "btree", name: "datetime_ops"},
	{am: "btree", name: "float_ops"},
	{am: "btree", name: "network_ops"},
	{am: "btree", name: "integer_ops"},
	{am: "btree", name: "interval_ops"},
	{am: "btree", name: "numeric_ops"},
	{am: "btree", name: "oid_ops"},
	{am: "btree", name: "text_ops"},
	{am: "btree", name: "time_ops"},
	{am: "btree", name: "timetz_ops"},
	{am: "btree", name: "varbit_ops"},
	{am: "btree", name: "uuid_ops"},
	{am: "btree", name: "record_ops"},
	{am: "btree", name: "jsonb_ops"},
	{am: "hash", name: "bpchar_ops"},
	{am: "hash", name: "char_ops"},
	{am: "hash", name: "datetime_ops"},
	{am: "hash", name: "bytea_ops"},
	{am: "hash", name: "float_ops"},
	{am: "hash", name: "integer_ops"},
	{am: "hash", name: "oid_ops"},
	{am: "hash", name: "text_ops"},
	{am: "hash", name: "numeric_ops"},
	{am: "hash", name: "bool_ops"},
	{am: "hash", name: "uuid_ops"},
	{am: "hash", name: "jsonb_ops"},
}

// pgOpfamilyRowIter is the sql.RowIter for the pg_opfamily table.
type pgOpfamilyRowIter struct {
	families []operatorFamily
	idx      int
}

var _ sql.RowIter = (*pgOpfamilyRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgOpfamilyRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.families) {
		return nil, io.EOF
	}
	iter.idx++
	family := iter.families[iter.idx-1]

	return sql.Row{
		family.oid(),                         // oid
		id.NewAccessMethod(family.am).AsId(), // opfmethod
		family.name,                          // opfname
		id.NewNamespace("pg_catalog").AsId(), // opfnamespace
		id.Null,                              // opfowner (TODO: object ownership is not tracked)
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgOpfamilyRowIter) Close(ctx *sql.Context) error {
	return nil
}
