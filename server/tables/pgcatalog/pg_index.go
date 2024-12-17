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

// Name implements the interface tables.Handler.
func (p PgIndexHandler) Name() string {
	return PgIndexName
}

// RowIter implements the interface tables.Handler.
func (p PgIndexHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// Use cached data from this process if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.indexes == nil {
		if err := cacheIndexMetadata(ctx, pgCatalogCache); err != nil {
			return nil, err
		}
	}

	return &pgIndexRowIter{
		indexes: pgCatalogCache.indexes,
		idxOIDs: pgCatalogCache.indexOIDs,
		tblOIDs: pgCatalogCache.indexTableOIDs,
		idx:     0,
	}, nil
}

// Schema implements the interface tables.Handler.
func (p PgIndexHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgIndexSchema,
		PkOrdinals: nil,
	}
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
	{Name: "indoption", Type: pgtypes.Int16Array, Default: nil, Nullable: false, Source: PgIndexName},  // TODO: type int2vector
	{Name: "indexprs", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgIndexName},          // TODO: type pg_node_tree, collation C
	{Name: "indpred", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgIndexName},           // TODO: type pg_node_tree, collation C
}

// pgIndexRowIter is the sql.RowIter for the pg_index table.
type pgIndexRowIter struct {
	indexes []sql.Index
	idxOIDs []id.Internal
	tblOIDs []id.Internal
	idx     int
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
		[]any{},                                  // indkey
		[]any{},                                  // indcollation
		[]any{},                                  // indclass
		[]any{},                                  // indoption
		nil,                                      // indexprs
		nil,                                      // indpred
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgIndexRowIter) Close(ctx *sql.Context) error {
	return nil
}

// cacheIndexMetadata iterates over the indexes in the current database and caches their metadata in |cache|. This
// cache holds pg_catalog data for the duration of a single query so that we don't have to generate the same pg_catalog
// data when multiple tables are joined together.
func cacheIndexMetadata(ctx *sql.Context, cache *pgCatalogCache) error {
	var indexes []sql.Index
	var indexSchemas []string
	var indexOIDs []id.Internal
	var tableOIDs []id.Internal

	err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		Index: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, index functions.ItemIndex) (cont bool, err error) {
			indexes = append(indexes, index.Item)
			indexSchemas = append(indexSchemas, schema.Item.SchemaName())
			indexOIDs = append(indexOIDs, index.OID)
			tableOIDs = append(tableOIDs, table.OID)
			return true, nil
		},
	})
	if err != nil {
		return err
	}

	cache.indexes = indexes
	cache.indexOIDs = indexOIDs
	cache.indexTableOIDs = tableOIDs
	cache.indexSchemas = indexSchemas
	return nil
}
