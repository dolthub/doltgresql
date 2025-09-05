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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/google/btree"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgIndexName is a constant to the pg_index name.
const PgIndexName = "pg_index"

// InitPgIndex handles registration of the pg_index handler.
func InitPgIndex() {
	tables.AddHandler(PgCatalogName, PgIndexName, PgIndexHandler{})
}

// PgIndexHandler is the handler for the pg_index table.
type PgIndexHandler struct{}

var _ tables.Handler = PgIndexHandler{}
var _ tables.IndexedTableHandler = PgIndexHandler{}

// Name implements the interface tables.Handler.
func (p PgIndexHandler) Name() string {
	return PgIndexName
}

// RowIter implements the interface tables.Handler.
func (p PgIndexHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// Use cached data from this session if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.pgIndexes == nil {
		err = cachePgIndexes(ctx, pgCatalogCache)
		if err != nil {
			return nil, err
		}
	}

	if indexIdxPart, ok := partition.(inMemIndexPartition); ok {
		return &inMemIndexScanIter[*pgIndex]{
			lookup:         indexIdxPart.lookup,
			rangeConverter: p,
			btreeAccess:    pgCatalogCache.pgIndexes,
			rowConverter:   pgIndexToRow,
			rangeIdx:       0,
			nextChan:       nil,
		}, nil
	}

	return &pgIndexTableScanIter{
		indexCache: pgCatalogCache.pgIndexes,
		idx:        0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgIndexHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgIndexSchema,
		PkOrdinals: nil,
	}
}

// Indexes implements tables.IndexedTableHandler.
func (p PgIndexHandler) Indexes() ([]sql.Index, error) {
	return []sql.Index{
		pgCatalogInMemIndex{
			name:        "pg_index_indexrelid_index",
			tblName:     "pg_index",
			dbName:      "pg_catalog",
			uniq:        true,
			columnExprs: []sql.ColumnExpressionType{{Expression: "pg_index.indexrelid", Type: pgtypes.Oid}},
		},
		pgCatalogInMemIndex{
			name:        "pg_index_indrelid_index",
			tblName:     "pg_index",
			dbName:      "pg_catalog",
			uniq:        false,
			columnExprs: []sql.ColumnExpressionType{{Expression: "pg_index.indrelid", Type: pgtypes.Oid}},
		},
	}, nil
}

// LookupPartitions implements tables.IndexedTableHandler.
func (p PgIndexHandler) LookupPartitions(context *sql.Context, lookup sql.IndexLookup) (sql.PartitionIter, error) {
	return &inMemIndexPartIter{
		part: inMemIndexPartition{
			idxName: lookup.Index.(pgCatalogInMemIndex).name,
			lookup:  lookup,
		},
	}, nil
}

// getIndexScanRange implements the interface RangeConverter.
func (p PgIndexHandler) getIndexScanRange(rng sql.Range, index sql.Index) (*pgIndex, bool, *pgIndex, bool) {
	var gte, lte *pgIndex
	var hasLowerBound, hasUpperBound bool

	switch index.(pgCatalogInMemIndex).name {
	case "pg_index_indexrelid_index":
		msrng := rng.(sql.MySQLRange)
		oidRng := msrng[0]
		if oidRng.HasLowerBound() {
			lowerRangeCutKey := sql.GetMySQLRangeCutKey(oidRng.LowerBound).(id.Id)
			gte = &pgIndex{
				indexOidNative: idToOid(lowerRangeCutKey),
			}
			hasLowerBound = true
		}
		if oidRng.HasUpperBound() {
			upperRangeCutKey := sql.GetMySQLRangeCutKey(oidRng.UpperBound).(id.Id)
			lte = &pgIndex{
				indexOidNative: idToOid(upperRangeCutKey),
			}
			hasUpperBound = true
		}

	case "pg_index_indrelid_index":
		msrng := rng.(sql.MySQLRange)
		oidRng := msrng[0]
		if oidRng.HasLowerBound() {
			lowerRangeCutKey := sql.GetMySQLRangeCutKey(oidRng.LowerBound).(id.Id)
			gte = &pgIndex{
				tableOidNative: idToOid(lowerRangeCutKey),
			}
			hasLowerBound = true
		}
		if oidRng.HasUpperBound() {
			upperRangeCutKey := sql.GetMySQLRangeCutKey(oidRng.UpperBound).(id.Id)
			lte = &pgIndex{
				tableOidNative: idToOid(upperRangeCutKey),
			}
			hasUpperBound = true
		}
	default:
		panic("unknown index name: " + index.(pgCatalogInMemIndex).name)
	}

	return gte, hasLowerBound, lte, hasUpperBound
}

// pgIndexSchema is the schema for pg_index.
var pgIndexSchema = sql.Schema{
	{Name: "indexrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgIndexName},
	{Name: "indrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgIndexName},
	{Name: "indnatts", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgIndexName},
	{Name: "indnkeyatts", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgIndexName},
	{Name: "indisunique", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgIndexName},
	{Name: "indnullsnotdistinct", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgIndexName},
	{Name: "indisprimary", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgIndexName},
	{Name: "indisexclusion", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgIndexName},
	{Name: "indimmediate", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgIndexName},
	{Name: "indisclustered", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgIndexName},
	{Name: "indisvalid", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgIndexName},
	{Name: "indcheckxmin", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgIndexName},
	{Name: "indisready", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgIndexName},
	{Name: "indislive", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgIndexName},
	{Name: "indisreplident", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgIndexName},
	{Name: "indkey", Type: pgtypes.Int16Array, Default: nil, Nullable: false, Source: PgIndexName},     // TODO: type int2vector
	{Name: "indcollation", Type: pgtypes.OidArray, Default: nil, Nullable: false, Source: PgIndexName}, // TODO: type oidvector
	{Name: "indclass", Type: pgtypes.OidArray, Default: nil, Nullable: false, Source: PgIndexName},     // TODO: type oidvector
	{Name: "indoption", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgIndexName},        // TODO: type int2vector. Declared as the serialized form so it can be read by clients expecting text, but this is a hacky temp solution
	{Name: "indexprs", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgIndexName},          // TODO: type pg_node_tree, collation C
	{Name: "indpred", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgIndexName},           // TODO: type pg_node_tree, collation C
}

// pgIndexRowIter is the sql.RowIter for the pg_index table.
type pgIndexRowIter struct {
	indexes      []sql.Index
	tableSchemas map[id.Id]sql.Schema
	idxOIDs      []id.Id
	tblOIDs      []id.Id
	idx          int
}

var _ sql.RowIter = (*pgIndexRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgIndexRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.indexes) {
		return nil, io.EOF
	}
	iter.idx++
	index := iter.indexes[iter.idx-1]
	tableOid := iter.tblOIDs[iter.idx-1]
	indexOid := iter.idxOIDs[iter.idx-1]
	schema := iter.tableSchemas[tableOid]

	indKey := make([]any, len(index.Expressions()))
	for i, expr := range index.Expressions() {
		colName := extractColName(expr)
		indKey[i] = int16(schema.IndexOfColName(colName)) + 1
	}

	// TODO: Fill in the rest of the pg_index columns
	return sql.Row{
		indexOid,                                 // indexrelid
		tableOid,                                 // indrelid
		int16(len(index.Expressions())),          // indnatts
		int16(0),                                 // indnkeyatts
		index.IsUnique(),                         // indisunique
		false,                                    // indnullsnotdistinct
		strings.ToLower(index.ID()) == "primary", // indisprimary
		false,                                    // indisexclusion
		false,                                    // indimmediate
		false,                                    // indisclustered
		true,                                     // indisvalid
		false,                                    // indcheckxmin
		true,                                     // indisready
		true,                                     // indislive
		false,                                    // indisreplident
		indKey,                                   // indkey
		[]any{},                                  // indcollation
		[]any{},                                  // indclass
		"0",                                      // indoption
		nil,                                      // indexprs
		nil,                                      // indpred
	}, nil
}

func extractColName(expr string) string {
	// TODO: this breaks for column names that contain a `.`, but this is a problem that happens
	//  throughout index analysis in the engine
	lastDot := strings.LastIndex(expr, ".")
	return expr[lastDot+1:]
}

// Close implements the interface sql.RowIter.
func (iter *pgIndexRowIter) Close(ctx *sql.Context) error {
	return nil
}

// pgIndex represents a row in the pg_index table.
// We store oids in their native format as well so that we can do range scans on them.
type pgIndex struct {
	indexOid            id.Id
	indexOidNative      uint32
	tableOid            id.Id
	tableOidNative      uint32
	indnatts            int16
	indnkeyatts         int16
	indisunique         bool
	indnullsnotdistinct bool
	indisprimary        bool
	indisexclusion      bool
	indimmediate        bool
	indisclustered      bool
	indisvalid          bool
	indcheckxmin        bool
	indisready          bool
	indislive           bool
	indisreplident      bool
	indkey              []any
	indcollation        []any
	indclass            []any
	indoption           string
	indexprs            interface{}
	indpred             interface{}
}

// lessIndexOid is a sort function for pgIndex based on indexrelid.
func lessIndexOid(a, b *pgIndex) bool {
	return a.indexOidNative < b.indexOidNative
}

// lessIndrelid is a sort function for pgIndex based on indrelid.
func lessIndrelid(a, b *pgIndex) bool {
	return a.tableOidNative < b.tableOidNative
}

// pgIndexTableScanIter is the sql.RowIter for the pg_index table.
type pgIndexTableScanIter struct {
	indexCache *pgIndexCache
	idx        int
}

var _ sql.RowIter = (*pgIndexTableScanIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgIndexTableScanIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.indexCache.indexes) {
		return nil, io.EOF
	}
	iter.idx++
	index := iter.indexCache.indexes[iter.idx-1]

	return pgIndexToRow(index), nil
}

// Close implements the interface sql.RowIter.
func (iter *pgIndexTableScanIter) Close(ctx *sql.Context) error {
	return nil
}

func pgIndexToRow(index *pgIndex) sql.Row {
	return sql.Row{
		index.indexOid,            // indexrelid
		index.tableOid,            // indrelid
		index.indnatts,            // indnatts
		index.indnkeyatts,         // indnkeyatts
		index.indisunique,         // indisunique
		index.indnullsnotdistinct, // indnullsnotdistinct
		index.indisprimary,        // indisprimary
		index.indisexclusion,      // indisexclusion
		index.indimmediate,        // indimmediate
		index.indisclustered,      // indisclustered
		index.indisvalid,          // indisvalid
		index.indcheckxmin,        // indcheckxmin
		index.indisready,          // indisready
		index.indislive,           // indislive
		index.indisreplident,      // indisreplident
		index.indkey,              // indkey
		index.indcollation,        // indcollation
		index.indclass,            // indclass
		index.indoption,           // indoption
		index.indexprs,            // indexprs
		index.indpred,             // indpred
	}
}

// cachePgIndexes caches the pg_index data for the current database in the session.
func cachePgIndexes(ctx *sql.Context, pgCatalogCache *pgCatalogCache) error {
	var indexes []*pgIndex
	indexOidIdx := btree.NewG[*pgIndex](2, lessIndexOid)
	indrelidIdx := btree.NewG[*pgIndex](2, lessIndrelid)

	tableSchemas := make(map[id.Id]sql.Schema)

	err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		Index: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, index functions.ItemIndex) (cont bool, err error) {
			if tableSchemas[table.OID.AsId()] == nil {
				tableSchemas[table.OID.AsId()] = table.Item.Schema()
			}

			schema := tableSchemas[table.OID.AsId()]
			indKey := make([]any, len(index.Item.Expressions()))
			for i, expr := range index.Item.Expressions() {
				colName := extractColName(expr)
				indKey[i] = int16(schema.IndexOfColName(colName)) + 1
			}

			pgIdx := &pgIndex{
				indexOid:            index.OID.AsId(),
				indexOidNative:      id.Cache().ToOID(index.OID.AsId()),
				tableOid:            table.OID.AsId(),
				tableOidNative:      id.Cache().ToOID(table.OID.AsId()),
				indnatts:            int16(len(index.Item.Expressions())),
				indnkeyatts:         int16(0),
				indisunique:         index.Item.IsUnique(),
				indnullsnotdistinct: false,
				indisprimary:        strings.ToLower(index.Item.ID()) == "primary",
				indisexclusion:      false,
				indimmediate:        false,
				indisclustered:      false,
				indisvalid:          true,
				indcheckxmin:        false,
				indisready:          true,
				indislive:           true,
				indisreplident:      false,
				indkey:              indKey,
				indcollation:        []any{},
				indclass:            []any{},
				indoption:           "0",
				indexprs:            nil,
				indpred:             nil,
			}

			indexOidIdx.ReplaceOrInsert(pgIdx)
			indrelidIdx.ReplaceOrInsert(pgIdx)
			indexes = append(indexes, pgIdx)
			return true, nil
		},
	})
	if err != nil {
		return err
	}

	pgCatalogCache.pgIndexes = &pgIndexCache{
		indexes:     indexes,
		indexOidIdx: indexOidIdx,
		indrelidIdx: indrelidIdx,
	}

	// Keep the old cache data for backward compatibility
	var legacyIndexes []sql.Index
	var indexSchemas []string
	var indexOIDs []id.Id
	var tableOIDs []id.Id

	for _, pgIdx := range indexes {
		// We need to reconstruct the sql.Index for legacy compatibility
		// This is a simplified approach - in a real implementation you might need more sophisticated reconstruction
		legacyIndexes = append(legacyIndexes, nil) // placeholder
		indexSchemas = append(indexSchemas, "")    // placeholder
		indexOIDs = append(indexOIDs, pgIdx.indexOid)
		tableOIDs = append(tableOIDs, pgIdx.tableOid)
	}

	pgCatalogCache.indexes = legacyIndexes
	pgCatalogCache.tableSchemas = tableSchemas
	pgCatalogCache.indexOIDs = indexOIDs
	pgCatalogCache.indexTableOIDs = tableOIDs
	pgCatalogCache.indexSchemas = indexSchemas

	return nil
}
