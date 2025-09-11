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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgAttributeName is a constant to the pg_attribute name.
const PgAttributeName = "pg_attribute"

// InitPgAttribute handles registration of the pg_attribute handler.
func InitPgAttribute() {
	tables.AddHandler(PgCatalogName, PgAttributeName, PgAttributeHandler{})
}

// PgAttributeHandler is the handler for the pg_attribute table.
type PgAttributeHandler struct{}

var _ tables.Handler = PgAttributeHandler{}
var _ tables.IndexedTableHandler = PgAttributeHandler{}

// Name implements the interface tables.Handler.
func (p PgAttributeHandler) Name() string {
	return PgAttributeName
}

// RowIter implements the interface tables.Handler.
func (p PgAttributeHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// Use cached data from this session if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.pgAttributes == nil {
		err = cachePgAttributes(ctx, pgCatalogCache)
		if err != nil {
			return nil, err
		}
	}

	if attrIdxPart, ok := partition.(inMemIndexPartition); ok {
		return &inMemIndexScanIter[*pgAttribute]{
			lookup:         attrIdxPart.lookup,
			rangeConverter: p,
			btreeAccess:    pgCatalogCache.pgAttributes,
			rowConverter:   pgAttributeToRow,
			rangeIdx:       0,
			nextChan:       nil,
		}, nil
	}

	return &pgAttributeTableScanIter{
		attributeCache: pgCatalogCache.pgAttributes,
		idx:            0,
	}, nil
}

// cachePgAttributes caches the pg_attribute data for the current database in the session.
func cachePgAttributes(ctx *sql.Context, pgCatalogCache *pgCatalogCache) error {
	var attributes []*pgAttribute
	attrelidIdx := NewUniqueInMemIndexStorage[*pgAttribute](lessAttNum)
	attrelidAttnameIdx := NewUniqueInMemIndexStorage[*pgAttribute](lessAttName)

	err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		Table: func(ctx *sql.Context, _ functions.ItemSchema, table functions.ItemTable) (cont bool, err error) {
			for i, col := range table.Item.Schema() {
				typeOid := id.Null
				if doltgresType, ok := col.Type.(*pgtypes.DoltgresType); ok {
					typeOid = doltgresType.ID.AsId()
				} else {
					// TODO: Remove once all information_schema tables are converted to use DoltgresType
					dt := pgtypes.FromGmsType(col.Type)
					typeOid = dt.ID.AsId()
				}

				generated := ""
				if col.Generated != nil {
					generated = "s"
				}

				dimensions := int16(0)
				if s, ok := col.Type.(sql.SetType); ok {
					dimensions = int16(s.NumberOfElements())
				}

				hasDefault := col.Default != nil

				attr := &pgAttribute{
					attrelid:       table.OID.AsId(),
					attrelidNative: id.Cache().ToOID(table.OID.AsId()),
					attname:        col.Name,
					atttypid:       typeOid,
					attnum:         int16(i + 1),
					attndims:       dimensions,
					attnotnull:     !col.Nullable,
					atthasdef:      hasDefault,
					attgenerated:   generated,
				}
				attrelidIdx.Add(attr)
				attrelidAttnameIdx.Add(attr)
				attributes = append(attributes, attr)
			}
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

		db := ctx.GetCurrentDatabase()
		for _, tblName := range systemTables {
			tbl, err := core.GetSqlTableFromContext(ctx, db, tblName)
			if err != nil {
				// Some of the system tables exist conditionally when accessed, so just skip them in this case
				if errors.Is(doltdb.ErrTableNotFound, err) {
					continue
				}
				return err
			}

			schema := tbl.Schema()
			for i, col := range schema {
				typeOid := id.Null
				if doltgresType, ok := col.Type.(*pgtypes.DoltgresType); ok {
					typeOid = doltgresType.ID.AsId()
				} else {
					dt := pgtypes.FromGmsType(col.Type)
					typeOid = dt.ID.AsId()
				}

				generated := ""
				if col.Generated != nil {
					generated = "s"
				}

				dimensions := int16(0)
				if s, ok := col.Type.(sql.SetType); ok {
					dimensions = int16(s.NumberOfElements())
				}

				hasDefault := col.Default != nil

				attr := &pgAttribute{
					attrelid:       id.NewTable(tblName.Schema, tblName.Name).AsId(),
					attrelidNative: id.Cache().ToOID(id.NewTable(tblName.Schema, tblName.Name).AsId()),
					attname:        col.Name,
					atttypid:       typeOid,
					attnum:         int16(i + 1),
					attndims:       dimensions,
					attnotnull:     !col.Nullable,
					atthasdef:      hasDefault,
					attgenerated:   generated,
				}
				attrelidIdx.Add(attr)
				attrelidAttnameIdx.Add(attr)
				attributes = append(attributes, attr)
			}
		}
	}

	pgCatalogCache.pgAttributes = &pgAttributeCache{
		attributes:         attributes,
		attrelidIdx:        attrelidIdx,
		attrelidAttnameIdx: attrelidAttnameIdx,
	}

	return nil
}

// getIndexScanRange implements the interface RangeConverter.
func (p PgAttributeHandler) getIndexScanRange(rng sql.Range, index sql.Index) (*pgAttribute, bool, *pgAttribute, bool) {
	var gte, lte *pgAttribute
	var hasLowerBound, hasUpperBound bool

	switch index.(pgCatalogInMemIndex).name {
	case "pg_attribute_relid_attnum_index":
		msrng := rng.(sql.MySQLRange)
		oidRng := msrng[0]
		attNumRng := msrng[1]

		var oidLower, oidUpper id.Id
		var attnumLower, attnumUpper int16

		if oidRng.HasLowerBound() {
			lb := sql.GetMySQLRangeCutKey(oidRng.LowerBound)
			if lb != nil {
				oidLower = lb.(id.Id)
				hasLowerBound = true
			}
		}

		if oidRng.HasUpperBound() {
			ub := sql.GetMySQLRangeCutKey(oidRng.UpperBound)
			if ub != nil {
				oidUpper = ub.(id.Id)
				hasUpperBound = true
			}
		}

		if attNumRng.HasLowerBound() {
			lb := sql.GetMySQLRangeCutKey(attNumRng.LowerBound)
			if lb != nil {
				attnumLower = lb.(int16)
			}
		}

		if attNumRng.HasUpperBound() {
			ub := sql.GetMySQLRangeCutKey(attNumRng.UpperBound)
			if ub != nil {
				attnumUpper = ub.(int16)
			}
		}

		if hasLowerBound {
			gte = &pgAttribute{
				attrelidNative: idToOid(oidLower),
				attnum:         attnumLower,
			}
		}

		if hasUpperBound {
			lte = &pgAttribute{
				attrelidNative: idToOid(oidUpper),
				attnum:         attnumUpper,
			}
		}

	case "pg_attribute_relid_attnam_index":
		msrng := rng.(sql.MySQLRange)
		attrelidRange := msrng[0]
		attnameRange := msrng[1]
		var attrelidLower, attrelidUpper uint32
		var attnameLower, attnameUpper string

		if attrelidRange.HasLowerBound() {
			lb := sql.GetMySQLRangeCutKey(attrelidRange.LowerBound)
			if lb != nil {
				lowerRangeCutKey := lb.(id.Id)
				attrelidLower = idToOid(lowerRangeCutKey)
				hasLowerBound = true
			}
		}
		if attrelidRange.HasUpperBound() {
			ub := sql.GetMySQLRangeCutKey(attrelidRange.UpperBound)
			if ub != nil {
				upperRangeCutKey := ub.(id.Id)
				attrelidUpper = idToOid(upperRangeCutKey)
				hasUpperBound = true
			}
		}

		if attnameRange.HasLowerBound() {
			lb := sql.GetMySQLRangeCutKey(attnameRange.LowerBound)
			if lb != nil {
				attnameLower = lb.(string)
			}
		}
		if attnameRange.HasUpperBound() {
			ub := sql.GetMySQLRangeCutKey(attnameRange.UpperBound)
			if ub != nil {
				attnameUpper = ub.(string)
			}
		}

		if attrelidRange.HasLowerBound() || attnameRange.HasLowerBound() {
			gte = &pgAttribute{
				attrelidNative: attrelidLower,
				attname:        attnameLower,
			}
		}

		if attrelidRange.HasUpperBound() || attnameRange.HasUpperBound() {
			lte = &pgAttribute{
				attrelidNative: attrelidUpper,
				attname:        attnameUpper,
			}
		}
	default:
		panic("unknown index name: " + index.(pgCatalogInMemIndex).name)
	}

	return gte, hasLowerBound, lte, hasUpperBound
}

// PkSchema implements the interface tables.Handler.
func (p PgAttributeHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAttributeSchema,
		PkOrdinals: nil,
	}
}

// Indexes implements tables.IndexedTableHandler.
func (p PgAttributeHandler) Indexes() ([]sql.Index, error) {
	return []sql.Index{
		pgCatalogInMemIndex{
			name:    "pg_attribute_relid_attnum_index",
			tblName: "pg_attribute",
			dbName:  "pg_catalog",
			uniq:    true,
			columnExprs: []sql.ColumnExpressionType{
				{Expression: "pg_attribute.attrelid", Type: pgtypes.Oid},
				{Expression: "pg_attribute.attnum", Type: pgtypes.Int16},
			},
		},
		pgCatalogInMemIndex{
			name:    "pg_attribute_relid_attnam_index",
			tblName: "pg_attribute",
			dbName:  "pg_catalog",
			uniq:    true,
			columnExprs: []sql.ColumnExpressionType{
				{Expression: "pg_attribute.attrelid", Type: pgtypes.Oid},
				{Expression: "pg_attribute.attname", Type: pgtypes.Name},
			},
		},
	}, nil
}

// LookupPartitions implements tables.IndexedTableHandler.
func (p PgAttributeHandler) LookupPartitions(context *sql.Context, lookup sql.IndexLookup) (sql.PartitionIter, error) {
	return &inMemIndexPartIter{
		part: inMemIndexPartition{
			idxName: lookup.Index.(pgCatalogInMemIndex).name,
			lookup:  lookup,
		},
	}, nil
}

// pgAttributeSchema is the schema for pg_attribute.
var pgAttributeSchema = sql.Schema{
	{Name: "attrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "atttypid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attlen", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attnum", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attcacheoff", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "atttypmod", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attndims", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attbyval", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attalign", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attstorage", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attcompression", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attnotnull", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "atthasdef", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "atthasmissing", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attidentity", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attgenerated", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attisdropped", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attislocal", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attinhcount", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attstattarget", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attcollation", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgAttributeName},        // TODO: type aclitem[]
	{Name: "attoptions", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgAttributeName},    // TODO: collation C
	{Name: "attfdwoptions", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgAttributeName}, // TODO: collation C
	{Name: "attmissingval", Type: pgtypes.AnyArray, Default: nil, Nullable: true, Source: PgAttributeName},
}

// pgAttribute represents a row in the pg_attribute table.
// We store oids in their native format as well so that we can do range scans on them.
type pgAttribute struct {
	attrelid       id.Id
	attrelidNative uint32
	attname        string
	atttypid       id.Id
	attnum         int16
	attndims       int16
	attnotnull     bool
	atthasdef      bool
	attgenerated   string
}

// lessAttNum is a sort function for pgAttribute based on attrelid.
func lessAttNum(a, b *pgAttribute) bool {
	// Some keys used for lookups set only the first column, which means we only compare the second if it's set for
	// both entries
	if a.attrelidNative == b.attrelidNative && a.attnum != 0 && b.attnum != 0 {
		return a.attnum < b.attnum
	}
	return a.attrelidNative < b.attrelidNative
}

// lessAttName is a sort function for pgAttribute based on attrelid, then attname.
func lessAttName(a, b *pgAttribute) bool {
	// Some keys used for lookups set only the first column, which means we only compare the second if it's set for
	// both entries
	if a.attrelidNative == b.attrelidNative && a.attname != "" && b.attname != "" {
		return a.attname < b.attname
	}
	return a.attrelidNative < b.attrelidNative
}

// pgAttributeTableScanIter is the sql.RowIter for the pg_attribute table.
type pgAttributeTableScanIter struct {
	attributeCache *pgAttributeCache
	idx            int
}

var _ sql.RowIter = (*pgAttributeTableScanIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAttributeTableScanIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.attributeCache.attributes) {
		return nil, io.EOF
	}
	iter.idx++
	attr := iter.attributeCache.attributes[iter.idx-1]

	return pgAttributeToRow(attr), nil
}

// Close implements the interface sql.RowIter.
func (iter *pgAttributeTableScanIter) Close(ctx *sql.Context) error {
	return nil
}

func pgAttributeToRow(attr *pgAttribute) sql.Row {
	// TODO: Fill in the rest of the pg_attribute columns
	return sql.Row{
		attr.attrelid,     // attrelid
		attr.attname,      // attname
		attr.atttypid,     // atttypid
		int16(0),          // attlen
		attr.attnum,       // attnum
		int32(-1),         // attcacheoff
		int32(-1),         // atttypmod
		attr.attndims,     // attndims
		false,             // attbyval
		"i",               // attalign
		"p",               // attstorage
		"",                // attcompression
		attr.attnotnull,   // attnotnull
		attr.atthasdef,    // atthasdef
		false,             // atthasmissing
		"",                // attidentity
		attr.attgenerated, // attgenerated
		false,             // attisdropped
		true,              // attislocal
		int16(0),          // attinhcount
		int16(-1),         // attstattarget
		id.Null,           // attcollation
		nil,               // attacl
		nil,               // attoptions
		nil,               // attfdwoptions
		nil,               // attmissingval
	}
}
