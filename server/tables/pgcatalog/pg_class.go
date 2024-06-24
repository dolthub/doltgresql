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

	var class []Class

	for _, db := range c.AllDatabases(ctx) {
		// Get tables and table indexes
		err := sql.DBTableIter(ctx, db, func(t sql.Table) (cont bool, err error) {
			hasIndexes := false

			if it, ok := t.(sql.IndexAddressable); ok {
				idxs, err := it.GetIndexes(ctx)
				if err != nil {
					return false, err
				}
				for _, idx := range idxs {
					class = append(class, Class{name: idx.ID(), hasIndexes: false, kind: "i"})
				}

				if len(idxs) > 0 {
					hasIndexes = true
				}
			}

			class = append(class, Class{name: t.Name(), hasIndexes: hasIndexes, kind: "r"})

			return true, nil
		})
		if err != nil {
			return nil, err
		}

		// Get views
		if vdb, ok := db.(sql.ViewDatabase); ok {
			views, err := vdb.AllViews(ctx)
			if err != nil {
				return nil, err
			}

			for _, view := range views {
				class = append(class, Class{name: view.Name, hasIndexes: false, kind: "v"})
			}
		}

	}

	return &pgClassRowIter{
		class: class,
		idx:   0,
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

type Class struct {
	name       string
	hasIndexes bool
	kind       string // r = ordinary table, i = index, S = sequence, t = TOAST table, v = view, m = materialized view, c = composite type, f = foreign table, p = partitioned table, I = partitioned index
}

// pgClassRowIter is the sql.RowIter for the pg_class table.
type pgClassRowIter struct {
	class []Class
	idx   int
}

var _ sql.RowIter = (*pgClassRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgClassRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.class) {
		return nil, io.EOF
	}
	iter.idx++
	cl := iter.class[iter.idx-1]

	// TODO: Fill in the rest of the pg_class columns
	return sql.Row{
		int32(iter.idx), // oid
		cl.name,         // relname
		int32(0),        // relnamespace
		int32(0),        // reltype
		int32(0),        // reloftype
		int32(0),        // relowner
		int32(0),        // relam
		int32(0),        // relfilenode
		int32(0),        // reltablespace
		int32(0),        // relpages
		float32(0),      // reltuples
		int32(0),        // relallvisible
		int32(0),        // reltoastrelid
		cl.hasIndexes,   // relhasindex
		false,           // relisshared
		"p",             // relpersistence
		cl.kind,         // relkind
		int16(0),        // relnatts
		int16(0),        // relchecks
		false,           // relhasrules
		false,           // relhastriggers
		false,           // relhassubclass
		false,           // relrowsecurity
		false,           // relforcerowsecurity
		true,            // relispopulated
		"d",             // relreplident
		false,           // relispartition
		int32(0),        // relrewrite
		int32(0),        // relfrozenxid
		int32(0),        // relminmxid
		nil,             // relacl
		nil,             // reloptions
		nil,             // relpartbound
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgClassRowIter) Close(ctx *sql.Context) error {
	return nil
}
