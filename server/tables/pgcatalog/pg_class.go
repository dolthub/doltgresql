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

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/google/btree"

	"github.com/dolthub/doltgresql/core"
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
var _ tables.IndexedTableHandler = PgClassHandler{}

// Name implements the interface tables.Handler.
func (p PgClassHandler) Name() string {
	return PgClassName
}

// RowIter implements the interface tables.Handler.
func (p PgClassHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// Use cached data from this session if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.pgClasses == nil {
		err = cachePgClasses(ctx, pgCatalogCache)
		if err != nil {
			return nil, err
		}
	}

	if classIdxPart, ok := partition.(inMemIndexPartition); ok {
		return &inMemIndexScanIter[*pgClass]{
			lookup:         classIdxPart.lookup,
			rangeConverter: p,
			btreeAccess:    pgCatalogCache.pgClasses,
			rowConverter:   pgClassToRow,
			rangeIdx:       0,
			nextChan:       nil,
		}, nil
	}

	return &pgClassTableScanIter{
		classCache: pgCatalogCache.pgClasses,
		idx:        0,
	}, nil
}

// cachePgClasses caches the pg_class data for the current database in the session.
func cachePgClasses(ctx *sql.Context, pgCatalogCache *pgCatalogCache) error {
	var classes []*pgClass
	tableHasIndexes := make(map[uint32]struct{})
	nameIdx := btree.NewG[*pgClass](2, lessName)
	oidIdx := btree.NewG(2, lessOid)

	err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		Index: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, index functions.ItemIndex) (cont bool, err error) {
			tableHasIndexes[id.Cache().ToOID(table.OID.AsId())] = struct{}{}
			class := &pgClass{
				oid:             index.OID.AsId(),
				oidNative:       id.Cache().ToOID(index.OID.AsId()),
				name:            formatIndexName(index.Item),
				hasIndexes:      false,
				kind:            "i",
				schemaOid:       schema.OID.AsId(),
				schemaOidNative: id.Cache().ToOID(schema.OID.AsId()),
			}
			nameIdx.ReplaceOrInsert(class)
			oidIdx.ReplaceOrInsert(class)
			classes = append(classes, class)
			return true, nil
		},
		Table: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable) (cont bool, err error) {
			_, hasIndexes := tableHasIndexes[id.Cache().ToOID(table.OID.AsId())]
			class := &pgClass{
				oid:             table.OID.AsId(),
				oidNative:       id.Cache().ToOID(table.OID.AsId()),
				name:            table.Item.Name(),
				hasIndexes:      hasIndexes,
				kind:            "r",
				schemaOid:       schema.OID.AsId(),
				schemaOidNative: id.Cache().ToOID(schema.OID.AsId()),
			}
			nameIdx.ReplaceOrInsert(class)
			oidIdx.ReplaceOrInsert(class)
			classes = append(classes, class)
			return true, nil
		},
		View: func(ctx *sql.Context, schema functions.ItemSchema, view functions.ItemView) (cont bool, err error) {
			class := &pgClass{
				oid:             view.OID.AsId(),
				oidNative:       id.Cache().ToOID(view.OID.AsId()),
				name:            view.Item.Name,
				hasIndexes:      false,
				kind:            "v",
				schemaOid:       schema.OID.AsId(),
				schemaOidNative: id.Cache().ToOID(schema.OID.AsId()),
			}
			nameIdx.ReplaceOrInsert(class)
			oidIdx.ReplaceOrInsert(class)
			classes = append(classes, class)
			return true, nil
		},
		Sequence: func(ctx *sql.Context, schema functions.ItemSchema, sequence functions.ItemSequence) (cont bool, err error) {
			class := &pgClass{
				oid:             sequence.OID.AsId(),
				oidNative:       id.Cache().ToOID(sequence.OID.AsId()),
				name:            sequence.Item.Id.SequenceName(),
				hasIndexes:      false,
				kind:            "S",
				schemaOid:       schema.OID.AsId(),
				schemaOidNative: id.Cache().ToOID(schema.OID.AsId()),
			}
			nameIdx.ReplaceOrInsert(class)
			oidIdx.ReplaceOrInsert(class)
			classes = append(classes, class)
			return true, nil
		},
	})
	if err != nil {
		return err
	}

	if includeSystemTables {
		_, root, err := core.GetRootFromContext(ctx)
		if err != nil {
			return err
		}

		systemTables, err := resolve.GetGeneratedSystemTables(ctx, root)
		if err != nil {
			return err
		}

		for _, tblName := range systemTables {
			class := &pgClass{
				oid:       id.NewTable(tblName.Schema, tblName.Name).AsId(),
				name:      tblName.Name,
				schemaOid: id.NewNamespace(tblName.Schema).AsId(),
				kind:      "r",
			}
			nameIdx.ReplaceOrInsert(class)
			oidIdx.ReplaceOrInsert(class)
			classes = append(classes, class)
		}
	}

	pgCatalogCache.pgClasses = &pgClassCache{
		classes: classes,
		nameIdx: nameIdx,
		oidIdx:  oidIdx,
	}
	
	return nil
}

// formatIndexName returns the name of an index for display
func formatIndexName(idx sql.Index) string {
	if idx.ID() == "PRIMARY" {
		return fmt.Sprintf("%s_pkey", idx.Table())
	}
	return idx.ID()
	// TODO: Unnamed indexes should have below format
	// return fmt.Sprintf("%s_%s_key", idx.Table(), idx.ID())
}

// getIndexScanRange implements the interface RangeConverter.
func (p PgClassHandler) getIndexScanRange(rng sql.Range, index sql.Index) (*pgClass, bool, *pgClass, bool) {
	var gte, lte *pgClass
	var hasLowerBound, hasUpperBound bool

	switch index.(pgCatalogInMemIndex).name {
	case "pg_class_oid_index":
		msrng := rng.(sql.MySQLRange)
		oidRng := msrng[0]
		if oidRng.HasLowerBound() {
			lowerRangeCutKey := sql.GetMySQLRangeCutKey(oidRng.LowerBound)
			oidLower := uint32(lowerRangeCutKey.(int32))
			gte = &pgClass{
				oidNative: oidLower,
			}
			hasLowerBound = true
		}
		if oidRng.HasUpperBound() {
			upperRangeCutKey := sql.GetMySQLRangeCutKey(oidRng.UpperBound)
			oidUpper := uint32(upperRangeCutKey.(int32))
			lte = &pgClass{
				oidNative: oidUpper,
			}
			hasUpperBound = true
		}

	case "pg_class_relname_nsp_index":
		msrng := rng.(sql.MySQLRange)
		relNameRange := msrng[0]
		schemaOidRange := msrng[1]
		var relnameLower, relnameUpper string
		var schemaOidLower, schemaOidUpper uint32

		if relNameRange.HasLowerBound() {
			relnameLower = sql.GetMySQLRangeCutKey(relNameRange.LowerBound).(string)
			hasLowerBound = true
		}
		if relNameRange.HasUpperBound() {
			relnameUpper = sql.GetMySQLRangeCutKey(relNameRange.UpperBound).(string)
			hasUpperBound = true
		}

		if schemaOidRange.HasLowerBound() {
			lowerRangeCutKey := sql.GetMySQLRangeCutKey(schemaOidRange.LowerBound)
			schemaOidLower = uint32(lowerRangeCutKey.(int32))
		}
		if schemaOidRange.HasUpperBound() {
			upperRangeCutKey := sql.GetMySQLRangeCutKey(schemaOidRange.UpperBound)
			schemaOidUpper = uint32(upperRangeCutKey.(int32))
		}

		if relNameRange.HasLowerBound() || schemaOidRange.HasLowerBound() {
			gte = &pgClass{
				name:      relnameLower,
				schemaOidNative: schemaOidLower,
			}
		}

		if relNameRange.HasUpperBound() || schemaOidRange.HasUpperBound() {
			lte = &pgClass{
				name:      relnameUpper,
				schemaOidNative: schemaOidUpper,
			}
		}
	default:
		panic("unknown index name: " + index.(pgCatalogInMemIndex).name)
	}

	return gte, hasLowerBound, lte, hasUpperBound
}

// PkSchema implements the interface tables.Handler.
func (p PgClassHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgClassSchema,
		PkOrdinals: nil,
	}
}

// Indexes implements tables.IndexedTableHandler.
func (p PgClassHandler) Indexes() ([]sql.Index, error) {
	return []sql.Index{
		pgCatalogInMemIndex{
			name:        "pg_class_oid_index",
			tblName:     "pg_class",
			dbName:      "pg_catalog",
			uniq:        true,
			columnExprs: []sql.ColumnExpressionType{{Expression: "pg_class.oid", Type: pgtypes.Oid}},
		},
		pgCatalogInMemIndex{
			name:    "pg_class_relname_nsp_index",
			tblName: "pg_class",
			dbName:  "pg_catalog",
			uniq:    true,
			columnExprs: []sql.ColumnExpressionType{
				{Expression: "pg_class.relname", Type: pgtypes.Name},
				{Expression: "pg_class.relnamespace", Type: pgtypes.Oid},
			},
		},
	}, nil
}

// LookupPartitions implements tables.IndexedTableHandler.
func (p PgClassHandler) LookupPartitions(context *sql.Context, lookup sql.IndexLookup) (sql.PartitionIter, error) {
	return &inMemIndexPartIter{
		part: inMemIndexPartition{
			idxName: lookup.Index.(pgCatalogInMemIndex).name,
			lookup:  lookup,
		},
	}, nil
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
// We store oids in their native format as well so that we can do range scans on them.
type pgClass struct {
	oid             id.Id
	oidNative       uint32
	name            string
	schemaOid       id.Id
	schemaOidNative uint32
	hasIndexes      bool
	kind            string // r = ordinary table, i = index, S = sequence, t = TOAST table, v = view, m = materialized view, c = composite type, f = foreign table, p = partitioned table, I = partitioned index
}

// lessOid is a sort function for pgClass based on oid.
func lessOid(a, b *pgClass) bool {
	return a.oidNative < b.oidNative
}

// lessName is a sort function for pgClass based on name, then schemaOid.
func lessName(a, b *pgClass) bool {
	if a.name == b.name {
		return a.schemaOidNative < b.schemaOidNative
	}
	return a.name < b.name
}

// pgClassTableScanIter is the sql.RowIter for the pg_class table.
type pgClassTableScanIter struct {
	classCache *pgClassCache
	idx        int
	idxLookup  *sql.IndexLookup
}

var _ sql.RowIter = (*pgClassTableScanIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgClassTableScanIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.classCache.classes) {
		return nil, io.EOF
	}
	iter.idx++
	class := iter.classCache.classes[iter.idx-1]

	return pgClassToRow(class), nil
}

func pgClassToRow(class *pgClass) sql.Row {
	// TODO: this is temporary definition of 'relam' field
	var relam = id.Null
	if class.kind == "i" {
		relam = id.NewAccessMethod("btree").AsId()
	} else if class.kind == "r" || class.kind == "t" {
		relam = id.NewAccessMethod("heap").AsId()
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
	}
}

// Close implements the interface sql.RowIter.
func (iter *pgClassTableScanIter) Close(ctx *sql.Context) error {
	return nil
}
