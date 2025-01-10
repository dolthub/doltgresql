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
	// Use cached data from this process if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.pgClasses == nil {
		var classes []pgClass
		tableHasIndexes := make(map[uint32]struct{})

		err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
			Index: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, index functions.ItemIndex) (cont bool, err error) {
				tableHasIndexes[id.Cache().ToOID(table.OID.Internal())] = struct{}{}
				classes = append(classes, pgClass{
					oid:        index.OID.Internal(),
					name:       getIndexName(index.Item),
					hasIndexes: false,
					kind:       "i",
					schemaOid:  schema.OID.Internal(),
				})
				return true, nil
			},
			Table: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable) (cont bool, err error) {
				_, hasIndexes := tableHasIndexes[id.Cache().ToOID(table.OID.Internal())]
				classes = append(classes, pgClass{
					oid:        table.OID.Internal(),
					name:       table.Item.Name(),
					hasIndexes: hasIndexes,
					kind:       "r",
					schemaOid:  schema.OID.Internal(),
				})
				return true, nil
			},
			View: func(ctx *sql.Context, schema functions.ItemSchema, view functions.ItemView) (cont bool, err error) {
				classes = append(classes, pgClass{
					oid:        view.OID.Internal(),
					name:       view.Item.Name,
					hasIndexes: false,
					kind:       "v",
					schemaOid:  schema.OID.Internal(),
				})
				return true, nil
			},
			Sequence: func(ctx *sql.Context, schema functions.ItemSchema, sequence functions.ItemSequence) (cont bool, err error) {
				classes = append(classes, pgClass{
					oid:        sequence.OID.Internal(),
					name:       sequence.Item.Name.SequenceName(),
					hasIndexes: false,
					kind:       "S",
					schemaOid:  schema.OID.Internal(),
				})
				return true, nil
			},
		})
		if err != nil {
			return nil, err
		}
		pgCatalogCache.pgClasses = classes
	}

	return &pgClassRowIter{
		classes: pgCatalogCache.pgClasses,
		idx:     0,
	}, nil
}

// getIndexName returns the name of an index.
func getIndexName(idx sql.Index) string {
	if idx.ID() == "PRIMARY" {
		return fmt.Sprintf("%s_pkey", idx.Table())
	}
	return idx.ID()
	// TODO: Unnamed indexes should have below format
	// return fmt.Sprintf("%s_%s_key", idx.Table(), idx.ID())
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
	{Name: "relpersistence", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relkind", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relnatts", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relchecks", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relhasrules", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relhastriggers", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relhassubclass", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relrowsecurity", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relforcerowsecurity", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relispopulated", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relreplident", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relispartition", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relrewrite", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relfrozenxid", Type: pgtypes.Xid, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relminmxid", Type: pgtypes.Xid, Default: nil, Nullable: false, Source: PgClassName},
	{Name: "relacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgClassName},     // TODO: type aclitem[]
	{Name: "reloptions", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgClassName}, // TODO: collation C
	{Name: "relpartbound", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgClassName},    // TODO: type pg_node_tree, collation C
}

// pgClass represents a row in the pg_class table.
type pgClass struct {
	oid        id.Internal
	name       string
	schemaOid  id.Internal
	hasIndexes bool
	kind       string // r = ordinary table, i = index, S = sequence, t = TOAST table, v = view, m = materialized view, c = composite type, f = foreign table, p = partitioned table, I = partitioned index
}

// pgClassRowIter is the sql.RowIter for the pg_class table.
type pgClassRowIter struct {
	classes []pgClass
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

	// TODO: this is temporary definition of 'relam' field
	var relam = id.Null
	if class.kind == "i" {
		relam = id.NewInternalAccessMethod("btree").Internal()
	} else if class.kind == "r" || class.kind == "t" {
		relam = id.NewInternalAccessMethod("heap").Internal()
	}

	// TODO: Fill in the rest of the pg_class columns
	return sql.Row{
		class.oid,        // oid
		class.name,       // relname
		class.schemaOid,  // relnamespace
		id.Null,          // reltype
		id.Null,          // reloftype
		id.Null,          // relowner
		relam,            // relam
		id.Null,          // relfilenode
		id.Null,          // reltablespace
		int32(0),         // relpages
		float32(0),       // reltuples
		int32(0),         // relallvisible
		id.Null,          // reltoastrelid
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
		id.Null,          // relrewrite
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
