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

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/sql"

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
	doltSession := dsess.DSessFromSess(ctx.Session)
	c := sqle.NewDefault(doltSession.Provider()).Analyzer.Catalog

	var constraints []pgConstraint

	_, err := currentDatabaseSchemaIter(ctx, c, func(db sql.DatabaseSchema) (bool, error) {
		schemaOid := genOid(db.Name(), db.SchemaName())

		err := sql.DBTableIter(ctx, db, func(t sql.Table) (cont bool, err error) {
			// Get indexes
			if it, ok := t.(sql.IndexAddressable); ok {
				idxs, err := it.GetIndexes(ctx)
				if err != nil {
					return false, err
				}
				for _, idx := range idxs {
					idxOid := genOid(db.Name(), db.SchemaName(), t.Name(), idx.ID())
					constraints = append(constraints, pgConstraint{
						oid:          idxOid,
						name:         idx.ID(),
						schemaOid:    schemaOid,
						conType:      "p",
						tableOid:     genOid(db.Name(), db.SchemaName(), t.Name()),
						idxOid:       idxOid,
						tableRefOid:  uint32(0),
						fkUpdateType: "",
						fkDeleteType: "",
					})
				}
			}

			// Get foreign keys
			if ft, ok := t.(sql.ForeignKeyTable); ok {
				fks, err := ft.GetDeclaredForeignKeys(ctx)
				if err != nil {
					return false, err
				}
				for _, fk := range fks {
					fkOid := genOid(db.Name(), db.SchemaName(), fk.Table, fk.Name)
					constraints = append(constraints, pgConstraint{
						oid:          fkOid,
						name:         fk.Name,
						schemaOid:    schemaOid,
						conType:      "f",
						tableOid:     genOid(db.Name(), db.SchemaName(), fk.Table),
						idxOid:       fkOid,
						tableRefOid:  genOid(db.Name(), db.SchemaName(), fk.ParentTable),
						fkUpdateType: getFKAction(fk.OnUpdate),
						fkDeleteType: getFKAction(fk.OnDelete),
					})
				}
			}

			// Get checks
			if ct, ok := t.(sql.CheckTable); ok {
				checks, err := ct.GetChecks(ctx)
				if err != nil {
					return false, err
				}
				for _, check := range checks {
					checkOid := genOid(db.Name(), db.SchemaName(), t.Name(), check.Name)
					constraints = append(constraints, pgConstraint{
						oid:          checkOid,
						name:         check.Name,
						schemaOid:    schemaOid,
						conType:      "c",
						tableOid:     genOid(db.Name(), db.SchemaName(), t.Name()),
						idxOid:       uint32(0),
						tableRefOid:  uint32(0),
						fkUpdateType: "",
						fkDeleteType: "",
					})
				}
			}

			return true, nil
		})
		if err != nil {
			return false, err
		}

		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return &pgConstraintRowIter{
		constraints: constraints,
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
	{Name: "contype", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "condeferrable", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "condeferred", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "convalidated", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "contypid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conindid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conparentid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confupdtype", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confdeltype", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confmatchtype", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgConstraintName},
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
	oid       uint32
	name      string
	schemaOid uint32
	conType   string // c = check constraint, f = foreign key constraint, p = primary key constraint, u = unique constraint, t = constraint trigger, x = exclusion constraint
	tableOid  uint32
	// typeOid      uint32
	idxOid       uint32
	tableRefOid  uint32
	fkUpdateType string // a = no action, r = restrict, c = cascade, n = set null, d = set default
	fkDeleteType string // a = no action, r = restrict, c = cascade, n = set null, d = set default
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

	return sql.Row{
		con.oid,          // oid
		con.name,         // conname
		con.schemaOid,    // connamespace
		con.conType,      // contype
		false,            // condeferrable
		false,            // condeferred
		true,             // convalidated
		con.tableOid,     // conrelid
		uint32(0),        // contypid
		con.idxOid,       // conindid
		uint32(0),        // conparentid
		con.tableRefOid,  // confrelid
		con.fkUpdateType, // confupdtype
		con.fkDeleteType, // confdeltype
		"f",              // confmatchtype
		true,             // conislocal
		int16(0),         // coninhcount
		false,            // connoinherit
		nil,              // conkey
		nil,              // confkey
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
