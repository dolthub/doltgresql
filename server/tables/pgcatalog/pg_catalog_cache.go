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

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/sequences"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/types/oid"
)

// pgCatalogCache is a session cache that stores the contents of pg_catalog tables. Since this cache instance is only
// ever used by a single session, it does not include any synchronization for concurrent data access. A pgCatalogCache
// only caches data for the length of a single query, using |pid| to identify the current query. This means that the
// initial read of data from a pg_catalog table will always be generated fresh, then stored in this cache for any other
// table reads needed as part of that same query.
type pgCatalogCache struct {
	// pid marks the process id for the query this cache data represents. pgCatalogCache instances currently only
	// cache data within the context of a single query – not across multiple queries.
	pid uint64

	// pg_classes
	pgClasses []pgClass

	// pg_constraints
	pgConstraints []pgConstraint

	// pg_namespace
	schemaNames []string
	schemaOids  []uint32

	// pg_attribute
	attributeCols      []*sql.Column
	attributeTableOIDs []uint32
	attributeColIdxs   []int

	// pg_index / pg_indexes
	indexes        []sql.Index
	indexOIDs      []uint32
	indexTableOIDs []uint32
	indexSchemas   []string

	// pg_sequence
	sequences    []*sequences.Sequence
	sequenceOids []uint32

	// pg_attrdef
	attrdefCols      []oid.ItemColumnDefault
	attrdefTableOIDs []uint32

	// pg_views
	views       []sql.ViewDefinition
	viewSchemas []string

	// pg_types
	types        []pgtypes.DoltgresType
	pgCatalogOid uint32

	// pg_tables
	tables       []sql.Table
	tableSchemas []string
}

// newPgCatalogCache creates a new pgCatalogCache, with the query/process ID set to |pid|. The PID is important,
// since pgCatalogCache instances only cache data for the duration of a single query.
func newPgCatalogCache(pid uint64) *pgCatalogCache {
	return &pgCatalogCache{
		pid: pid,
	}
}

// getPgCatalogCache returns the pgCatalogCache instance for the current query in this session. If no cache exists
// yet, then a new one is created and returned. Note that pgCatalogCache only caches catalog data for a single query,
// so if the PID of the current context does not match the PID of the context when the pgCatalogCache was created,
// then a new cache will be created.
func getPgCatalogCache(ctx *sql.Context) (*pgCatalogCache, error) {
	untypedPgCatalogCache, err := core.GetPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}
	if untypedPgCatalogCache == nil {
		return initializeNewPgCatalogCache(ctx)
	}

	cache, ok := untypedPgCatalogCache.(*pgCatalogCache)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T for pg_catalog cache", untypedPgCatalogCache)
	}
	if cache.pid != ctx.Pid() {
		return initializeNewPgCatalogCache(ctx)
	}
	return cache, nil
}

// initializeNewPgCatalogCache creates a new pgCatalogCache instance and sets it in the context. This function should
// not be used directly, and should only be used directly by getPgCatalogCache(*sql.Context).
func initializeNewPgCatalogCache(ctx *sql.Context) (*pgCatalogCache, error) {
	cache := newPgCatalogCache(ctx.Pid())
	if err := core.SetPgCatalogCache(ctx, cache); err != nil {
		return nil, err
	}
	return cache, nil
}
