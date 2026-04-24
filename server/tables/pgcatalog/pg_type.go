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
	"math"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// pgTypeName is a constant to the pg_type name.
const (
	pgTypeName     = "pg_type"
	pgTypeOidIndex = "pg_type_oid_index"
	pgTypnameIndex = "pg_type_typname_nsp_index"
)

// InitPgType handles registration of the pg_type handler.
func InitPgType() {
	tables.AddHandler(PgCatalogName, pgTypeName, PgTypeHandler{})
}

// PgTypeHandler is the handler for the pg_type table.
type PgTypeHandler struct{}

var _ tables.Handler = PgTypeHandler{}
var _ tables.IndexedTableHandler = PgTypeHandler{}

// Name implements the interface tables.Handler.
func (p PgTypeHandler) Name() string {
	return pgTypeName
}

// RowIter implements the interface tables.Handler.
func (p PgTypeHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// Use cached data from this process if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.pgTypes == nil {
		err = cachePgTypes(ctx, pgCatalogCache)
		if err != nil {
			return nil, err
		}
	}

	if typeIdxPart, ok := partition.(inMemIndexPartition); ok {
		return &inMemIndexScanIter[*pgType]{
			lookup:         typeIdxPart.lookup,
			rangeConverter: p,
			btreeAccess:    pgCatalogCache.pgTypes,
			rowConverter:   pgTypeToRow,
		}, nil
	}

	return &pgTypeTableScanIter{
		typeCache: pgCatalogCache.pgTypes,
		idx:       0,
	}, nil
}

func cachePgTypes(ctx *sql.Context, pgCatalogCache *pgCatalogCache) error {
	var types []*pgType
	nameIdx := NewUniqueInMemIndexStorage[*pgType](lessTypeName)
	oidIdx := NewUniqueInMemIndexStorage[*pgType](lessTypeOid)

	schemasToOid := make(map[string]id.Namespace)
	err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		Schema: func(ctx *sql.Context, schema functions.ItemSchema) (cont bool, err error) {
			schemasToOid[schema.Item.SchemaName()] = schema.OID
			return true, nil
		},
	})
	if err != nil {
		return err
	}

	allTypes := pgtypes.GetAllBuitInTypes()
	typeColl, err := core.GetTypesCollectionFromContext(ctx)
	if err != nil {
		return err
	}
	userTypes, schemas, cnt, err := typeColl.GetAllTypes(ctx)
	if err != nil {
		return err
	}
	if cnt > 0 {
		for _, schema := range schemas {
			if schema != PgCatalogName {
				allTypes = append(allTypes, userTypes[schema]...)
			}
		}
	}

	for _, typ := range allTypes {
		schemaOid := schemasToOid[typ.ID.SchemaName()]
		t := &pgType{
			oid:             typ.ID.AsId(),
			oidNative:       id.Cache().ToOID(typ.ID.AsId()),
			name:            typ.Name(),
			schemaOid:       schemaOid.AsId(),
			schemaOidNative: id.Cache().ToOID(schemaOid.AsId()),
			typ:             typ,
		}
		oidIdx.Add(t)
		nameIdx.Add(t)
		types = append(types, t)
	}

	pgCatalogCache.pgTypes = &pgTypeCache{
		types:   types,
		nameIdx: nameIdx,
		oidIdx:  oidIdx,
	}

	return nil
}

// getIndexScanRange implements the interface RangeConverter.
func (p PgTypeHandler) getIndexScanRange(rng sql.Range, index sql.Index) (*pgType, bool, *pgType, bool) {
	var gte, lt *pgType
	var hasLowerBound, hasUpperBound bool

	switch index.(pgCatalogInMemIndex).name {
	case pgTypeOidIndex:
		msrng := rng.(sql.MySQLRange)
		oidRng := msrng[0]
		if oidRng.HasLowerBound() {
			lb := sql.GetMySQLRangeCutKey(oidRng.LowerBound)
			if lb != nil {
				lowerRangeCutKey := lb.(id.Id)
				gte = &pgType{
					oidNative: idToOid(lowerRangeCutKey),
				}
				hasLowerBound = true
			}
		}

		if oidRng.HasUpperBound() {
			ub := sql.GetMySQLRangeCutKey(oidRng.UpperBound)
			if ub != nil {
				upperRangeCutKey := ub.(id.Id)
				lt = &pgType{
					oidNative: idToOid(upperRangeCutKey) + 1,
				}
				hasUpperBound = true
			}
		}

	case pgTypnameIndex:
		msrng := rng.(sql.MySQLRange)
		typNameRange := msrng[0]
		schemaOidRange := msrng[1]
		var typnameLower, typnameUpper string
		schemaOidLower := uint32(0)
		schemaOidUpper := uint32(math.MaxUint32)
		schemaOidUpperSet := false

		if typNameRange.HasLowerBound() {
			lb := sql.GetMySQLRangeCutKey(typNameRange.LowerBound)
			if lb != nil {
				typnameLower = lb.(string)
				hasLowerBound = true
			}
		}
		if typNameRange.HasUpperBound() {
			ub := sql.GetMySQLRangeCutKey(typNameRange.UpperBound)
			if ub != nil {
				typnameUpper = ub.(string)
				hasUpperBound = true
			}
		}

		if schemaOidRange.HasLowerBound() {
			lb := sql.GetMySQLRangeCutKey(schemaOidRange.LowerBound)
			if lb != nil {
				lowerRangeCutKey := lb.(id.Id)
				schemaOidLower = idToOid(lowerRangeCutKey)
			}
		}
		if schemaOidRange.HasUpperBound() {
			ub := sql.GetMySQLRangeCutKey(schemaOidRange.UpperBound)
			if ub != nil {
				upperRangeCutKey := ub.(id.Id)
				schemaOidUpper = idToOid(upperRangeCutKey)
				schemaOidUpperSet = true
			}
		}

		if typNameRange.HasLowerBound() || schemaOidRange.HasLowerBound() {
			gte = &pgType{
				name:            typnameLower,
				schemaOidNative: schemaOidLower,
			}
		}

		if typNameRange.HasUpperBound() || schemaOidRange.HasUpperBound() {
			// our less-than upper bound depends on whether we have a prefix match or both fields were set
			if !schemaOidUpperSet {
				typnameUpper = fmt.Sprintf("%s%o", typnameUpper, rune(0))
			} else {
				schemaOidUpper = schemaOidUpper + 1
			}
			lt = &pgType{
				name:            typnameUpper,
				schemaOidNative: schemaOidUpper,
			}
		}
	default:
		panic("unknown index name: " + index.(pgCatalogInMemIndex).name)
	}

	return gte, hasLowerBound, lt, hasUpperBound
}

// Schema implements the interface tables.Handler.
func (p PgTypeHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgTypeSchema,
		PkOrdinals: nil,
	}
}

// Indexes implements tables.IndexedTableHandler.
func (p PgTypeHandler) Indexes() ([]sql.Index, error) {
	return []sql.Index{
		pgCatalogInMemIndex{
			name:        pgTypeOidIndex,
			tblName:     "pg_type",
			dbName:      "pg_catalog",
			uniq:        true,
			columnExprs: []sql.ColumnExpressionType{{Expression: "pg_type.oid", Type: pgtypes.Oid}},
		},
		pgCatalogInMemIndex{
			name:    pgTypnameIndex,
			tblName: "pg_type",
			dbName:  "pg_catalog",
			uniq:    true,
			columnExprs: []sql.ColumnExpressionType{
				{Expression: "pg_type.typname", Type: pgtypes.Name},
				{Expression: "pg_type.typnamespace", Type: pgtypes.Oid},
			},
		},
	}, nil
}

// LookupPartitions implements tables.IndexedTableHandler.
func (p PgTypeHandler) LookupPartitions(_ *sql.Context, lookup sql.IndexLookup) (sql.PartitionIter, error) {
	return &inMemIndexPartIter{
		part: inMemIndexPartition{
			idxName: lookup.Index.(pgCatalogInMemIndex).name,
			lookup:  lookup,
		},
	}, nil
}

// pgTypeSchema is the schema for pg_type.
var pgTypeSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typlen", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typbyval", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typtype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typcategory", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typispreferred", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typisdefined", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typdelim", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typsubscript", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgTypeName}, // TODO: type regproc
	{Name: "typelem", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typarray", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typinput", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgTypeName},   // TODO: type regproc
	{Name: "typoutput", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgTypeName},  // TODO: type regproc
	{Name: "typreceive", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgTypeName}, // TODO: type regproc
	{Name: "typsend", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgTypeName},    // TODO: type regproc
	{Name: "typmodin", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgTypeName},   // TODO: type regproc
	{Name: "typmodout", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgTypeName},  // TODO: type regproc
	{Name: "typanalyze", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgTypeName}, // TODO: type regproc
	{Name: "typalign", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typstorage", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typnotnull", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typbasetype", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typtypmod", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typndims", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typcollation", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: pgTypeName},
	{Name: "typdefaultbin", Type: pgtypes.Text, Default: nil, Nullable: true, Source: pgTypeName}, // TODO: type pg_node_tree, collation C
	{Name: "typdefault", Type: pgtypes.Text, Default: nil, Nullable: true, Source: pgTypeName},    // TODO: collation C
	{Name: "typacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: pgTypeName},   // TODO: type aclitem[]
}

// pgType represents a row in the pg_type table.
// We store oids in their native format as well so that we can do range scans on them.
type pgType struct {
	oid             id.Id
	oidNative       uint32
	name            string
	schemaOid       id.Id
	schemaOidNative uint32
	typ             *pgtypes.DoltgresType
}

// lessTypeOid is a sort function for pgType based on oid.
func lessTypeOid(a, b *pgType) bool {
	return a.oidNative < b.oidNative
}

// lessTypeName is a sort function for pgType based on name, then schemaOid.
func lessTypeName(a, b *pgType) bool {
	if a.name == b.name {
		return a.schemaOidNative < b.schemaOidNative
	}
	return a.name < b.name
}

// pgTypeTableScanIter is the sql.RowIter for the pg_type table.
type pgTypeTableScanIter struct {
	typeCache *pgTypeCache
	idx       int
}

var _ sql.RowIter = (*pgTypeTableScanIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTypeTableScanIter) Next(_ *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.typeCache.types) {
		return nil, io.EOF
	}
	iter.idx++
	nextType := iter.typeCache.types[iter.idx-1]

	return pgTypeToRow(nextType), nil
}

// Close implements the interface sql.RowIter.
func (iter *pgTypeTableScanIter) Close(_ *sql.Context) error {
	return nil
}

func pgTypeToRow(nextType *pgType) sql.Row {
	typAcl := []any(nil)

	return sql.Row{
		nextType.oid,
		nextType.name,
		nextType.schemaOid,
		id.Null,
		nextType.typ.TypLength,           // typlen
		nextType.typ.PassedByVal,         // typbyval
		string(nextType.typ.TypType),     // typtype
		string(nextType.typ.TypCategory), // typcategory
		nextType.typ.IsPreferred,         // typispreferred
		nextType.typ.IsDefined,           // typisdefined
		nextType.typ.Delimiter,           // typdelim
		nextType.typ.RelID,               // typrelid
		nextType.typ.SubscriptFuncName(), // typsubscript
		nextType.typ.Elem.AsId(),         // typelem
		nextType.typ.Array.AsId(),        // typarray
		nextType.typ.InputFuncName(),     // typinput
		nextType.typ.OutputFuncName(),    // typoutput
		nextType.typ.ReceiveFuncName(),   // typreceive
		nextType.typ.SendFuncName(),      // typsend
		nextType.typ.ModInFuncName(),     // typmodin
		nextType.typ.ModOutFuncName(),    // typmodout
		nextType.typ.AnalyzeFuncName(),   // typanalyze
		string(nextType.typ.Align),       // typalign
		string(nextType.typ.Storage),     // typstorage
		nextType.typ.NotNull,             // typnotnull
		nextType.typ.BaseTypeID.AsId(),   // typbasetype
		nextType.typ.TypMod,              // typtypmod
		nextType.typ.NDims,               // typndims
		nextType.typ.TypCollation.AsId(), // typcollation
		nextType.typ.DefaulBin,           // typdefaultbin
		nextType.typ.Default,             // typdefault
		typAcl,                           // typacl
	}
}
