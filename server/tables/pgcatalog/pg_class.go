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

// PgClassName is a constant to the pg_class name.
const PgClassName = "pg_class"

// InitPgClass handles registration of the pg_class handler.
func InitPgClass() {
	tables.AddHandler(PgCatalogName, PgClassName, PgClassHandler{})
}

// PgClassHandler is the handler for the pg_class table.
type PgClassHandler struct{}

var _ tables.Handler = PgClassHandler{}

// Name implements the interface tables.Handler.
func (p PgClassHandler) Name() string {
	return PgClassName
}

// RowIter implements the interface tables.Handler.
func (p PgClassHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	doltSession := dsess.DSessFromSess(ctx.Session)
	c := sqle.NewDefault(doltSession.Provider()).Analyzer.Catalog

	var classes []Class

	_, err := currentDatabaseSchemaIter(ctx, c, func(db sql.DatabaseSchema) (bool, error) {
		dbName := db.Name()
		schName := db.SchemaName()
		schOid := genOid(dbName, schName)

		// Get tables and table indexes
		err := sql.DBTableIter(ctx, db, func(t sql.Table) (cont bool, err error) {
			tableName := t.Name()
			hasIndexes := false

			if it, ok := t.(sql.IndexAddressable); ok {
				idxs, err := it.GetIndexes(ctx)
				if err != nil {
					return false, err
				}
				for _, idx := range idxs {
					classes = append(classes, Class{
						oid:        genOid(dbName, schName, tableName, idx.ID()),
						name:       idx.ID(),
						hasIndexes: false,
						kind:       "i",
						schemaOid:  schOid,
					})
				}

				if len(idxs) > 0 {
					hasIndexes = true
				}
			}

			classes = append(classes, Class{
				oid:        genOid(dbName, schName, tableName),
				name:       tableName,
				hasIndexes: hasIndexes,
				kind:       "r",
				schemaOid:  schOid,
			})

			return true, nil
		})
		if err != nil {
			return false, err
		}

		// Get views
		if vdb, ok := db.(sql.ViewDatabase); ok {
			views, err := vdb.AllViews(ctx)
			if err != nil {
				return false, err
			}

			for _, view := range views {
				classes = append(classes, Class{
					oid:        genOid(dbName, schName, view.Name),
					name:       view.Name,
					hasIndexes: false,
					kind:       "v",
					schemaOid:  schOid,
				})
			}
		}

		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return &pgClassRowIter{
		classes: classes,
		idx:     0,
	}, nil
}

// Schema implements the interface tables.Handler.
func (p PgClassHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgClassSchema,
		PkOrdinals: nil,
	}
}

// pgClassSchema is the schema for pg_class.
var pgClassSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "reltype", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "reloftype", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relam", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relfilenode", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "reltablespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relpages", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "reltuples", Type: pgtypes.Float32, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relallvisible", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "reltoastrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relhasindex", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relisshared", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relpersistence", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relkind", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relnatts", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relchecks", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relhasrules", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relhastriggers", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relhassubclass", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relrowsecurity", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relforcerowsecurity", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relispopulated", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relreplident", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relispartition", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relrewrite", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relfrozenxid", Type: pgtypes.Xid, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relminmxid", Type: pgtypes.Xid, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgClassName},     // TODO: type aclitem[]
	{Name: "reloptions", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgClassName}, // TODO: collation C
	{Name: "relpartbound", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgClassName},    // TODO: type pg_node_tree, collation C
}

// Class represents a row in the pg_class table.
type Class struct {
	oid        uint32
	name       string
	schemaOid  uint32
	hasIndexes bool
	kind       string // r = ordinary table, i = index, S = sequence, t = TOAST table, v = view, m = materialized view, c = composite type, f = foreign table, p = partitioned table, I = partitioned index
}

// pgClassRowIter is the sql.RowIter for the pg_class table.
type pgClassRowIter struct {
	classes []Class
	idx     int
}

var _ sql.RowIter = (*pgClassRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgClassRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.classes) {
		return nil, io.EOF
	}
	iter.idx++
	class := iter.classes[iter.idx-1]

	// TODO: Fill in the rest of the pg_class columns
	return sql.Row{
		class.oid,        // oid
		class.name,       // relname
		class.schemaOid,  // relnamespace
		uint32(0),        // reltype
		uint32(0),        // reloftype
		uint32(0),        // relowner
		uint32(0),        // relam
		uint32(0),        // relfilenode
		uint32(0),        // reltablespace
		int32(0),         // relpages
		float32(0),       // reltuples
		int32(0),         // relallvisible
		uint32(0),        // reltoastrelid
		class.hasIndexes, // relhasindex
		false,            // relisshared
		"p",              // relpersistence
		class.kind,       // relkind
		int16(0),         // relnatts
		int16(0),         // relchecks
		false,            // relhasrules
		false,            // relhastriggers
		false,            // relhassubclass
		false,            // relrowsecurity
		false,            // relforcerowsecurity
		true,             // relispopulated
		"d",              // relreplident
		false,            // relispartition
		uint32(0),        // relrewrite
		uint32(0),        // relfrozenxid
		uint32(0),        // relminmxid
		nil,              // relacl
		nil,              // reloptions
		nil,              // relpartbound
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgClassRowIter) Close(ctx *sql.Context) error {
	return nil
}
