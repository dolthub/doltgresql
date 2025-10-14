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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/sequences"
	"github.com/dolthub/doltgresql/server/functions"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// pgNamespace represents a row in the pg_namespace table.
// We store oids in their native format as well so that we can do range scans on them.
type pgNamespace struct {
	oid       id.Id
	oidNative uint32
	name      string
}

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
	pgClasses *pgClassCache

	// pg_constraints
	pgConstraints *pgConstraintCache

	// pg_namespace
	pgNamespaces *pgNamespaceCache

	// pg_attribute
	pgAttributes *pgAttributeCache

	// pg_index / pg_indexes
	pgIndexes *pgIndexCache

	// pg_sequence
	sequences    []*sequences.Sequence
	sequenceOids []id.Id

	// pg_attrdef
	attrdefCols      []functions.ItemColumnDefault
	attrdefTableOIDs []id.Id

	// pg_views
	views       []sql.ViewDefinition
	viewSchemas []string

	// pg_types
	types        []*pgtypes.DoltgresType
	schemasToOid map[string]id.Namespace

	// pg_tables
	tables       []sql.Table
	systemTables []doltdb.TableName
}

// pgClassCache holds cached data for the pg_class table, including two btree indexes for fast lookups by OID and
// by relname
type pgClassCache struct {
	classes []*pgClass
	nameIdx *inMemIndexStorage[*pgClass]
	oidIdx  *inMemIndexStorage[*pgClass]
}

// getIndex implements BTreeStorageAccess.
func (p pgClassCache) getIndex(name string) *inMemIndexStorage[*pgClass] {
	switch name {
	case "pg_class_oid_index":
		return p.oidIdx
	case "pg_class_relname_nsp_index":
		return p.nameIdx
	default:
		panic("unknown pg_class index: " + name)
	}
}

var _ BTreeStorageAccess[*pgClass] = &pgClassCache{}

// pgIndexCache holds cached data for the pg_index table, including two btree indexes for fast lookups by index OID
type pgIndexCache struct {
	indexes     []*pgIndex
	tableNames  map[id.Id]string
	indexOidIdx *inMemIndexStorage[*pgIndex]
	indrelidIdx *inMemIndexStorage[*pgIndex]
}

var _ BTreeStorageAccess[*pgIndex] = &pgIndexCache{}

// pgConstraintCache holds cached data for the pg_constraint table, including three btree indexes for fast lookups
type pgConstraintCache struct {
	constraints     []*pgConstraint
	oidIdx          *inMemIndexStorage[*pgConstraint]
	relidTypNameIdx *inMemIndexStorage[*pgConstraint]
	nameSchemaIdx   *inMemIndexStorage[*pgConstraint]
	typIdx          *inMemIndexStorage[*pgConstraint]
}

var _ BTreeStorageAccess[*pgConstraint] = &pgConstraintCache{}

// getIndex implements BTreeStorageAccess.
func (p pgConstraintCache) getIndex(name string) *inMemIndexStorage[*pgConstraint] {
	switch name {
	case "pg_constraint_oid_index":
		return p.oidIdx
	case "pg_constraint_conrelid_contypid_conname_index":
		return p.relidTypNameIdx
	case "pg_constraint_conname_nsp_index":
		return p.nameSchemaIdx
	case "pg_constraint_contypid_index":
		return p.typIdx
	default:
		panic("unknown pg_constraint index: " + name)
	}
}

// pgAttributeCache holds cached data for the pg_attribute table, including two btree indexes for fast lookups
type pgAttributeCache struct {
	attributes         []*pgAttribute
	attrelidIdx        *inMemIndexStorage[*pgAttribute]
	attrelidAttnameIdx *inMemIndexStorage[*pgAttribute]
}

var _ BTreeStorageAccess[*pgAttribute] = &pgAttributeCache{}

// pgNamespaceCache holds cached data for the pg_namespace table, including two btree indexes for fast lookups by OID and by name
type pgNamespaceCache struct {
	namespaces []*pgNamespace
	oidIdx     *inMemIndexStorage[*pgNamespace]
	nameIdx    *inMemIndexStorage[*pgNamespace]
}

var _ BTreeStorageAccess[*pgNamespace] = &pgNamespaceCache{}

// getIndex implements BTreeStorageAccess.
func (p pgNamespaceCache) getIndex(name string) *inMemIndexStorage[*pgNamespace] {
	switch name {
	case "pg_namespace_oid_index":
		return p.oidIdx
	case "pg_namespace_nspname_index":
		return p.nameIdx
	default:
		panic("unknown pg_namespace index: " + name)
	}
}

func (p pgAttributeCache) getIndex(name string) *inMemIndexStorage[*pgAttribute] {
	switch name {
	case "pg_attribute_relid_attnum_index":
		return p.attrelidIdx
	case "pg_attribute_relid_attnam_index":
		return p.attrelidAttnameIdx
	default:
		panic("unknown pg_attribute index: " + name)
	}
}

// getIndex implements BTreeStorageAccess.
func (p pgIndexCache) getIndex(name string) *inMemIndexStorage[*pgIndex] {
	switch name {
	case "pg_index_indexrelid_index":
		return p.indexOidIdx
	case "pg_index_indrelid_index":
		return p.indrelidIdx
	default:
		panic("unknown pg_index index: " + name)
	}
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
		return nil, errors.Errorf("unexpected type %T for pg_catalog cache", untypedPgCatalogCache)
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
