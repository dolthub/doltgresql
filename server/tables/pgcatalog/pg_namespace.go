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

// PgNamespaceName is a constant to the pg_namespace name.
const PgNamespaceName = "pg_namespace"

// InitPgNamespace handles registration of the pg_namespace handler.
func InitPgNamespace() {
	tables.AddHandler(PgCatalogName, PgNamespaceName, PgNamespaceHandler{})
}

// PgNamespaceHandler is the handler for the pg_namespace table.
type PgNamespaceHandler struct{}

var _ tables.Handler = PgNamespaceHandler{}
var _ tables.IndexedTableHandler = PgNamespaceHandler{}

// Name implements the interface tables.Handler.
func (p PgNamespaceHandler) Name() string {
	return PgNamespaceName
}

// RowIter implements the interface tables.Handler.
func (p PgNamespaceHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// Use cached data from this session if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.pgNamespaces == nil {
		err = cachePgNamespaces(ctx, pgCatalogCache)
		if err != nil {
			return nil, err
		}
	}

	if namespaceIdxPart, ok := partition.(inMemIndexPartition); ok {
		return &inMemIndexScanIter[*pgNamespace]{
			lookup:         namespaceIdxPart.lookup,
			rangeConverter: p,
			btreeAccess:    pgCatalogCache.pgNamespaces,
			rowConverter:   pgNamespaceToRow,
			rangeIdx:       0,
			nextChan:       nil,
		}, nil
	}

	return &pgNamespaceTableScanIter{
		namespaceCache: pgCatalogCache.pgNamespaces,
		idx:            0,
	}, nil
}

// cachePgNamespaces caches the pg_namespace data for the current database in the session.
func cachePgNamespaces(ctx *sql.Context, pgCatalogCache *pgCatalogCache) error {
	var namespaces []*pgNamespace
	oidIdx := NewUniqueInMemIndexStorage[*pgNamespace](lessNamespaceOid)
	nameIdx := NewUniqueInMemIndexStorage[*pgNamespace](lessNamespaceName)

	err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		Schema: func(ctx *sql.Context, schema functions.ItemSchema) (cont bool, err error) {
			namespace := &pgNamespace{
				oid:       schema.OID.AsId(),
				oidNative: id.Cache().ToOID(schema.OID.AsId()),
				name:      schema.Item.SchemaName(),
			}
			oidIdx.Add(namespace)
			nameIdx.Add(namespace)
			namespaces = append(namespaces, namespace)
			return true, nil
		},
	})
	if err != nil {
		return err
	}

	pgCatalogCache.pgNamespaces = &pgNamespaceCache{
		namespaces: namespaces,
		oidIdx:     oidIdx,
		nameIdx:    nameIdx,
	}

	return nil
}

// getIndexScanRange implements the interface RangeConverter.
func (p PgNamespaceHandler) getIndexScanRange(rng sql.Range, index sql.Index) (*pgNamespace, bool, *pgNamespace, bool) {
	var gte, lt *pgNamespace
	var hasLowerBound, hasUpperBound bool

	switch index.(pgCatalogInMemIndex).name {
	case "pg_namespace_oid_index":
		msrng := rng.(sql.MySQLRange)
		oidRng := msrng[0]
		if oidRng.HasLowerBound() {
			lb := sql.GetMySQLRangeCutKey(oidRng.LowerBound)
			if lb != nil {
				lowerRangeCutKey := lb.(id.Id)
				gte = &pgNamespace{
					oidNative: idToOid(lowerRangeCutKey),
				}
				hasLowerBound = true
			}
		}
		if oidRng.HasUpperBound() {
			ub := sql.GetMySQLRangeCutKey(oidRng.UpperBound)
			if ub != nil {
				upperRangeCutKey := ub.(id.Id)
				lt = &pgNamespace{
					oidNative: idToOid(upperRangeCutKey) + 1,
				}
				hasUpperBound = true
			}
		}

	case "pg_namespace_nspname_index":
		msrng := rng.(sql.MySQLRange)
		nameRng := msrng[0]
		var nameLower, nameUpper string

		if nameRng.HasLowerBound() {
			lb := sql.GetMySQLRangeCutKey(nameRng.LowerBound)
			if lb != nil {
				nameLower = lb.(string)
				hasLowerBound = true
			}
		}
		if nameRng.HasUpperBound() {
			ub := sql.GetMySQLRangeCutKey(nameRng.UpperBound)
			if ub != nil {
				nameUpper = ub.(string)
				nameUpper = fmt.Sprintf("%s%o", nameUpper, rune(0))
				hasUpperBound = true
			}
		}

		if hasLowerBound {
			gte = &pgNamespace{
				name: nameLower,
			}
		}
		if hasUpperBound {
			lt = &pgNamespace{
				name: nameUpper,
			}
		}
	default:
		panic("unknown index name: " + index.(pgCatalogInMemIndex).name)
	}

	return gte, hasLowerBound, lt, hasUpperBound
}

// PkSchema implements the interface tables.Handler.
func (p PgNamespaceHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgNamespaceSchema,
		PkOrdinals: nil,
	}
}

// Indexes implements tables.IndexedTableHandler.
func (p PgNamespaceHandler) Indexes() ([]sql.Index, error) {
	return []sql.Index{
		pgCatalogInMemIndex{
			name:        "pg_namespace_oid_index",
			tblName:     "pg_namespace",
			dbName:      "pg_catalog",
			uniq:        true,
			columnExprs: []sql.ColumnExpressionType{{Expression: "pg_namespace.oid", Type: pgtypes.Oid}},
		},
		pgCatalogInMemIndex{
			name:        "pg_namespace_nspname_index",
			tblName:     "pg_namespace",
			dbName:      "pg_catalog",
			uniq:        true,
			columnExprs: []sql.ColumnExpressionType{{Expression: "pg_namespace.nspname", Type: pgtypes.Name}},
		},
	}, nil
}

// LookupPartitions implements tables.IndexedTableHandler.
func (p PgNamespaceHandler) LookupPartitions(context *sql.Context, lookup sql.IndexLookup) (sql.PartitionIter, error) {
	return &inMemIndexPartIter{
		part: inMemIndexPartition{
			idxName: lookup.Index.(pgCatalogInMemIndex).name,
			lookup:  lookup,
		},
	}, nil
}

// pgNamespaceSchema is the schema for pg_namespace.
var pgNamespaceSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgNamespaceName},
	{Name: "nspname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgNamespaceName},
	{Name: "nspowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgNamespaceName},
	{Name: "nspacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgNamespaceName}, // TODO: type aclitem[]
}

// lessNamespaceOid is a sort function for pgNamespace based on oid.
func lessNamespaceOid(a, b *pgNamespace) bool {
	return a.oidNative < b.oidNative
}

// lessNamespaceName is a sort function for pgNamespace based on name.
func lessNamespaceName(a, b *pgNamespace) bool {
	return a.name < b.name
}

// pgNamespaceTableScanIter is the sql.RowIter for the pg_namespace table.
type pgNamespaceTableScanIter struct {
	namespaceCache *pgNamespaceCache
	idx            int
}

var _ sql.RowIter = (*pgNamespaceTableScanIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgNamespaceTableScanIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.namespaceCache.namespaces) {
		return nil, io.EOF
	}
	iter.idx++
	namespace := iter.namespaceCache.namespaces[iter.idx-1]

	return pgNamespaceToRow(namespace), nil
}

// Close implements the interface sql.RowIter.
func (iter *pgNamespaceTableScanIter) Close(ctx *sql.Context) error {
	return nil
}

func pgNamespaceToRow(namespace *pgNamespace) sql.Row {
	// TODO: columns are incomplete
	return sql.Row{
		namespace.oid,  // oid
		namespace.name, // nspname
		id.Null,        // nspowner
		nil,            // nspacl
	}
}
