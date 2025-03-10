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
	"fmt"
	"io"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgConstraintName is a constant to the pg_constraint name.
const PgConstraintName = "pg_constraint"

// InitPgConstraint handles registration of the pg_constraint handler.
func InitPgConstraint() {
	tables.AddHandler(PgCatalogName, PgConstraintName, PgConstraintHandler{})
}

// PgConstraintHandler is the handler for the pg_constraint table.
type PgConstraintHandler struct{}

var _ tables.Handler = PgConstraintHandler{}

// Name implements the interface tables.Handler.
func (p PgConstraintHandler) Name() string {
	return PgConstraintName
}

// RowIter implements the interface tables.Handler.
func (p PgConstraintHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// Use cached data from this process if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.pgConstraints == nil {
		var constraints []pgConstraint
		tableOIDs := make(map[id.Id]map[string]id.Id)
		tableColToIdxMap := make(map[string]int16)

		// We iterate over all tables first to obtain their OIDs, which we'll need to reference for foreign keys
		err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
			Table: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable) (cont bool, err error) {
				inner, ok := tableOIDs[schema.OID.AsId()]
				if !ok {
					inner = make(map[string]id.Id)
					tableOIDs[schema.OID.AsId()] = inner
				}
				inner[table.Item.Name()] = table.OID.AsId()

				for i, col := range table.Item.Schema() {
					tableColToIdxMap[fmt.Sprintf("%s.%s", table.Item.Name(), col.Name)] = int16(i + 1)
				}
				return true, nil
			},
		})
		if err != nil {
			return nil, err
		}

		// Then we iterate over everything to fill our constraints
		err = functions.IterateCurrentDatabase(ctx, functions.Callbacks{
			Check: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, check functions.ItemCheck) (cont bool, err error) {
				constraints = append(constraints, pgConstraint{
					oid:         check.OID.AsId(),
					name:        check.Item.Name,
					schemaOid:   schema.OID.AsId(),
					conType:     "c",
					tableOid:    table.OID.AsId(),
					idxOid:      id.Null,
					tableRefOid: id.Null,
				})
				return true, nil
			},
			ForeignKey: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, foreignKey functions.ItemForeignKey) (cont bool, err error) {
				conKey := make([]any, len(foreignKey.Item.Columns))
				for i, expr := range foreignKey.Item.Columns {
					conKey[i] = tableColToIdxMap[expr]
				}

				parentTableColToIdxMap := make(map[string]int16)
				parentTable, ok, err := schema.Item.GetTableInsensitive(ctx, foreignKey.Item.ParentTable)
				if err != nil {
					return false, err
				} else if ok {
					for i, col := range parentTable.Schema() {
						parentTableColToIdxMap[col.Name] = int16(i + 1)
					}
				}

				conFkey := make([]any, len(foreignKey.Item.ParentColumns))
				for i, expr := range foreignKey.Item.ParentColumns {
					conFkey[i] = parentTableColToIdxMap[expr]
				}

				constraints = append(constraints, pgConstraint{
					oid:          foreignKey.OID.AsId(),
					name:         foreignKey.Item.Name,
					schemaOid:    schema.OID.AsId(),
					conType:      "f",
					tableOid:     tableOIDs[schema.OID.AsId()][foreignKey.Item.Table],
					idxOid:       foreignKey.OID.AsId(),
					tableRefOid:  tableOIDs[schema.OID.AsId()][foreignKey.Item.ParentTable],
					fkUpdateType: getFKAction(foreignKey.Item.OnUpdate),
					fkDeleteType: getFKAction(foreignKey.Item.OnDelete),
					fkMatchType:  "s",
					conKey:       conKey,
					conFkey:      conFkey,
				})
				return true, nil
			},
			Index: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, index functions.ItemIndex) (cont bool, err error) {
				conType := "p"
				if index.Item.ID() != "PRIMARY" {
					if index.Item.IsUnique() {
						conType = "u"
					} else {
						// If this isn't a primary key or a unique index, then it's a regular index, and not
						// a constraint, so we don't need to report it in the pg_constraint table.
						return true, nil
					}
				}

				conKey := make([]any, len(index.Item.Expressions()))
				for i, expr := range index.Item.Expressions() {
					conKey[i] = tableColToIdxMap[expr]
				}

				constraints = append(constraints, pgConstraint{
					oid:         index.OID.AsId(),
					name:        getIndexName(index.Item),
					schemaOid:   schema.OID.AsId(),
					conType:     conType,
					tableOid:    table.OID.AsId(),
					idxOid:      index.OID.AsId(),
					tableRefOid: id.Null,
					conKey:      conKey,
					conFkey:     nil,
				})
				return true, nil
			},
		})
		if err != nil {
			return nil, err
		}
		pgCatalogCache.pgConstraints = constraints
	}

	return &pgConstraintRowIter{
		constraints: pgCatalogCache.pgConstraints,
		idx:         0,
	}, nil
}

func getFKAction(action sql.ForeignKeyReferentialAction) string {
	switch action {
	case sql.ForeignKeyReferentialAction_NoAction:
		return "a"
	case sql.ForeignKeyReferentialAction_Restrict:
		return "r"
	case sql.ForeignKeyReferentialAction_Cascade:
		return "c"
	case sql.ForeignKeyReferentialAction_SetNull:
		return "n"
	case sql.ForeignKeyReferentialAction_SetDefault:
		return "d"
	default:
		return ""
	}
}

// Schema implements the interface tables.Handler.
func (p PgConstraintHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     PgConstraintSchema,
		PkOrdinals: nil,
	}
}

// PgConstraintSchema is the schema for pg_constraint.
var PgConstraintSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "connamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "contype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "condeferrable", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "condeferred", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "convalidated", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "contypid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conindid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conparentid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confupdtype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confdeltype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confmatchtype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conislocal", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "coninhcount", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "connoinherit", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conkey", Type: pgtypes.Int16Array, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "confkey", Type: pgtypes.Int16Array, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "conpfeqop", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "conppeqop", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "conffeqop", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "confdelsetcols", Type: pgtypes.Int16Array, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "conexclop", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "conbin", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgConstraintName}, // TODO: type pg_node_tree, collation C
}

// pgConstraint is the struct for the pg_constraint table.
type pgConstraint struct {
	oid       id.Id
	name      string
	schemaOid id.Id
	conType   string // c = check constraint, f = foreign key constraint, p = primary key constraint, u = unique constraint, t = constraint trigger, x = exclusion constraint
	tableOid  id.Id
	// typeOid      id.Id
	idxOid       id.Id
	tableRefOid  id.Id
	fkUpdateType string // a = no action, r = restrict, c = cascade, n = set null, d = set default
	fkDeleteType string // a = no action, r = restrict, c = cascade, n = set null, d = set default
	fkMatchType  string // f = full, p = partial, s = simple
	conKey       []any
	conFkey      []any
}

// pgConstraintRowIter is the sql.RowIter for the pg_constraint table.
type pgConstraintRowIter struct {
	constraints []pgConstraint
	idx         int
}

var _ sql.RowIter = (*pgConstraintRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgConstraintRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.constraints) {
		return nil, io.EOF
	}
	iter.idx++
	con := iter.constraints[iter.idx-1]

	var conKey interface{}
	if len(con.conKey) == 0 {
		conKey = nil
	} else {
		conKey = con.conKey
	}

	var conFkey interface{}
	if len(con.conFkey) == 0 {
		conFkey = nil
	} else {
		conFkey = con.conFkey
	}

	return sql.Row{
		con.oid,          // oid
		con.name,         // conname
		con.schemaOid,    // connamespace
		con.conType,      // contype
		false,            // condeferrable
		false,            // condeferred
		true,             // convalidated
		con.tableOid,     // conrelid
		id.Null,          // contypid
		con.idxOid,       // conindid
		id.Null,          // conparentid
		con.tableRefOid,  // confrelid
		con.fkUpdateType, // confupdtype
		con.fkDeleteType, // confdeltype
		con.fkMatchType,  // confmatchtype
		true,             // conislocal
		int16(0),         // coninhcount
		true,             // connoinherit
		conKey,           // conkey
		conFkey,          // confkey
		nil,              // conpfeqop
		nil,              // conppeqop
		nil,              // conffeqop
		nil,              // confdelsetcols
		nil,              // conexclop
		nil,              // conbin
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgConstraintRowIter) Close(ctx *sql.Context) error {
	return nil
}
