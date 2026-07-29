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
	"github.com/dolthub/doltgresql/core/sequences"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgDependName is a constant to the pg_depend name.
const PgDependName = "pg_depend"

// Fixed OIDs of the system catalogs themselves, used in the classid/refclassid
// columns to identify which catalog the (referenced) object belongs to.
const (
	pgClassCatalogOID   = 1259 // pg_class
	pgAttrdefCatalogOID = 2604 // pg_attrdef
)

// InitPgDepend handles registration of the pg_depend handler.
func InitPgDepend() {
	tables.AddHandler(PgCatalogName, PgDependName, PgDependHandler{})
}

// PgDependHandler is the handler for the pg_depend table.
type PgDependHandler struct{}

var _ tables.Handler = PgDependHandler{}

// Name implements the interface tables.Handler.
func (p PgDependHandler) Name() string {
	return PgDependName
}

// RowIter implements the interface tables.Handler.
func (p PgDependHandler) RowIter(ctx *sql.Context, _ sql.Partition) (sql.RowIter, error) {
	// Use cached data from this process if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.dependRows == nil {
		err = cachePgDepends(ctx, pgCatalogCache)
		if err != nil {
			return nil, err
		}
	}

	return &pgDependRowIter{rows: pgCatalogCache.dependRows}, nil
}

// cachePgDepends caches the pg_depend data for the current database in the session.
func cachePgDepends(ctx *sql.Context, pgCatalogCache *pgCatalogCache) error {
	// pg_depend is a partial implementation containing only the dependencies that can currently be
	// derived from the database's schema:
	//   - sequence -> owning column ('a' auto dependencies, e.g. SERIAL columns and OWNED BY)
	//   - column default (pg_attrdef entry) -> table column ('a' auto dependencies)
	// TODO: emit the remaining dependency kinds: views -> tables/columns, foreign key and other
	//  constraints, indexes -> columns, functions -> types, extension membership ('e'), internal
	//  dependencies ('i'), and 'p' pinned rows for built-in objects.
	pgClassOid := id.NewOID(pgClassCatalogOID).AsId()
	pgAttrdefOid := id.NewOID(pgAttrdefCatalogOID).AsId()

	var rows []sql.Row
	var seqs []*sequences.Sequence
	tableSchemas := make(map[id.Table]sql.Schema)
	err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		Table: func(ctx *sql.Context, _ functions.ItemSchema, table functions.ItemTable) (cont bool, err error) {
			tableSchemas[table.OID] = table.Item.Schema(ctx)
			return true, nil
		},
		Sequence: func(ctx *sql.Context, _ functions.ItemSchema, sequence functions.ItemSequence) (cont bool, err error) {
			seqs = append(seqs, sequence.Item)
			return true, nil
		},
		ColumnDefault: func(ctx *sql.Context, _ functions.ItemSchema, table functions.ItemTable, colDefault functions.ItemColumnDefault) (cont bool, err error) {
			// Each pg_attrdef entry has an automatic dependency on the column it applies to
			rows = append(rows, sql.Row{
				pgAttrdefOid,                           // classid
				colDefault.OID.AsId(),                  // objid
				int32(0),                               // objsubid
				pgClassOid,                             // refclassid
				table.OID.AsId(),                       // refobjid
				int32(colDefault.Item.ColumnIndex + 1), // refobjsubid
				"a",                                    // deptype
			})
			return true, nil
		},
	})
	if err != nil {
		return err
	}

	// Sequences owned by a table column (SERIAL columns, OWNED BY clauses) have an automatic
	// dependency on that column. The owning table's schema is needed to resolve the column's
	// ordinal, so these rows are built after iteration completes.
	for _, seq := range seqs {
		if !seq.OwnerTable.IsValid() {
			continue
		}
		sch, ok := tableSchemas[seq.OwnerTable]
		if !ok {
			continue
		}
		for colIdx, col := range sch {
			if col.Name == seq.OwnerColumn {
				rows = append(rows, sql.Row{
					pgClassOid,            // classid
					seq.Id.AsId(),         // objid
					int32(0),              // objsubid
					pgClassOid,            // refclassid
					seq.OwnerTable.AsId(), // refobjid
					int32(colIdx + 1),     // refobjsubid
					"a",                   // deptype
				})
				break
			}
		}
	}

	pgCatalogCache.dependRows = rows
	return nil
}

// PkSchema implements the interface tables.Handler.
func (p PgDependHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgDependSchema,
		PkOrdinals: nil,
	}
}

// pgDependSchema is the schema for pg_depend.
var pgDependSchema = sql.Schema{
	{Name: "classid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDependName},
	{Name: "objid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDependName},
	{Name: "objsubid", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgDependName},
	{Name: "refclassid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDependName},
	{Name: "refobjid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDependName},
	{Name: "refobjsubid", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgDependName},
	{Name: "deptype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgDependName},
}

// pgDependRowIter is the sql.RowIter for the pg_depend table.
type pgDependRowIter struct {
	rows []sql.Row
	idx  int
}

var _ sql.RowIter = (*pgDependRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgDependRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.rows) {
		return nil, io.EOF
	}
	iter.idx++
	return iter.rows[iter.idx-1], nil
}

// Close implements the interface sql.RowIter.
func (iter *pgDependRowIter) Close(ctx *sql.Context) error {
	return nil
}
