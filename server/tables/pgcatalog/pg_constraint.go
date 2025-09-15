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

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgConstraintName is a constant to the pg_constraint name.
const PgConstraintName = "pg_constraint"

// InitPgConstraint handles registration of the pg_constraint handler.
func InitPgConstraint() {
	tables.AddHandler(PgCatalogName, PgConstraintName, PgConstraintHandler{})
}

// PgConstraintHandler is the handler for the pg_constraint table.
type PgConstraintHandler struct{}

var _ tables.Handler = PgConstraintHandler{}
var _ tables.IndexedTableHandler = PgConstraintHandler{}

// Name implements the interface tables.Handler.
func (p PgConstraintHandler) Name() string {
	return PgConstraintName
}

// RowIter implements the interface tables.Handler.
func (p PgConstraintHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// Use cached data from this session if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.pgConstraints == nil {
		err = cachePgConstraints(ctx, pgCatalogCache)
		if err != nil {
			return nil, err
		}
	}

	if constraintIdxPart, ok := partition.(inMemIndexPartition); ok {
		return &inMemIndexScanIter[*pgConstraint]{
			lookup:         constraintIdxPart.lookup,
			rangeConverter: p,
			btreeAccess:    pgCatalogCache.pgConstraints,
			rowConverter:   pgConstraintToRow,
			rangeIdx:       0,
			nextChan:       nil,
		}, nil
	}

	return &pgConstraintTableScanIter{
		constraintCache: pgCatalogCache.pgConstraints,
		idx:             0,
	}, nil
}

func getFKAction(action sql.ForeignKeyReferentialAction) string {
	switch action {
	case sql.ForeignKeyReferentialAction_NoAction:
		return "a"
	case sql.ForeignKeyReferentialAction_Restrict:
		return "r"
	case sql.ForeignKeyReferentialAction_Cascade:
		return "c"
	case sql.ForeignKeyReferentialAction_SetNull:
		return "n"
	case sql.ForeignKeyReferentialAction_SetDefault:
		return "d"
	default:
		return ""
	}
}

// PkSchema implements the interface tables.Handler.
func (p PgConstraintHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     PgConstraintSchema,
		PkOrdinals: nil,
	}
}

// Indexes implements tables.IndexedTableHandler.
func (p PgConstraintHandler) Indexes() ([]sql.Index, error) {
	return []sql.Index{
		pgCatalogInMemIndex{
			name:        "pg_constraint_oid_index",
			tblName:     "pg_constraint",
			dbName:      "pg_catalog",
			uniq:        true,
			columnExprs: []sql.ColumnExpressionType{{Expression: "pg_constraint.oid", Type: pgtypes.Oid}},
		},
		pgCatalogInMemIndex{
			name:        "pg_constraint_conrelid_index",
			tblName:     "pg_constraint",
			dbName:      "pg_catalog",
			uniq:        false,
			columnExprs: []sql.ColumnExpressionType{{Expression: "pg_constraint.conrelid", Type: pgtypes.Oid}},
		},
		pgCatalogInMemIndex{
			name:        "pg_constraint_conname_nsp_index",
			tblName:     "pg_constraint",
			dbName:      "pg_catalog",
			uniq:        true,
			columnExprs: []sql.ColumnExpressionType{
				{Expression: "pg_constraint.conname", Type: pgtypes.Name},
				{Expression: "pg_constraint.connamespace", Type: pgtypes.Oid},
			},
		},
	}, nil
}

// LookupPartitions implements tables.IndexedTableHandler.
func (p PgConstraintHandler) LookupPartitions(context *sql.Context, lookup sql.IndexLookup) (sql.PartitionIter, error) {
	return &inMemIndexPartIter{
		part: inMemIndexPartition{
			idxName: lookup.Index.(pgCatalogInMemIndex).name,
			lookup:  lookup,
		},
	}, nil
}

// getIndexScanRange implements the interface RangeConverter.
func (p PgConstraintHandler) getIndexScanRange(rng sql.Range, index sql.Index) (*pgConstraint, bool, *pgConstraint, bool) {
	var gte, lt *pgConstraint
	var hasLowerBound, hasUpperBound bool

	switch index.(pgCatalogInMemIndex).name {
	case "pg_constraint_oid_index":
		msrng := rng.(sql.MySQLRange)
		oidRng := msrng[0]
		if oidRng.HasLowerBound() {
			lb := sql.GetMySQLRangeCutKey(oidRng.LowerBound)
			if lb != nil {
				lowerRangeCutKey := lb.(id.Id)
				gte = &pgConstraint{
					oidNative: idToOid(lowerRangeCutKey),
				}
				hasLowerBound = true
			}
		}
		if oidRng.HasUpperBound() {
			ub := sql.GetMySQLRangeCutKey(oidRng.UpperBound)
			if ub != nil {
				upperRangeCutKey := ub.(id.Id)
				lt = &pgConstraint{
					oidNative: idToOid(upperRangeCutKey) + 1,
				}
				hasUpperBound = true
			}
		}

	case "pg_constraint_conrelid_index":
		msrng := rng.(sql.MySQLRange)
		oidRng := msrng[0]
		if oidRng.HasLowerBound() {
			lb := sql.GetMySQLRangeCutKey(oidRng.LowerBound)
			if lb != nil {
				lowerRangeCutKey := lb.(id.Id)
				gte = &pgConstraint{
					tableOidNative: idToOid(lowerRangeCutKey),
				}
				hasLowerBound = true
			}
		}
		if oidRng.HasUpperBound() {
			ub := sql.GetMySQLRangeCutKey(oidRng.UpperBound)
			if ub != nil {
				upperRangeCutKey := ub.(id.Id)
				lt = &pgConstraint{
					tableOidNative: idToOid(upperRangeCutKey) + 1,
				}
				hasUpperBound = true
			}
		}

	case "pg_constraint_conname_nsp_index":
		msrng := rng.(sql.MySQLRange)
		conNameRange := msrng[0]
		schemaOidRange := msrng[1]
		var conNameLower, conNameUpper string
		schemaOidLower := uint32(0)
		schemaOidUpper := uint32(math.MaxUint32)
		schemaOidUpperSet := false

		if conNameRange.HasLowerBound() {
			lb := sql.GetMySQLRangeCutKey(conNameRange.LowerBound)
			if lb != nil {
				conNameLower = lb.(string)
				hasLowerBound = true
			}
		}
		if conNameRange.HasUpperBound() {
			ub := sql.GetMySQLRangeCutKey(conNameRange.UpperBound)
			if ub != nil {
				conNameUpper = ub.(string)
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

		if conNameRange.HasLowerBound() || schemaOidRange.HasLowerBound() {
			gte = &pgConstraint{
				name:            conNameLower,
				schemaOidNative: schemaOidLower,
			}
		}

		if conNameRange.HasUpperBound() || schemaOidRange.HasUpperBound() {
			// our less-than upper bound depends on whether we have a prefix match or both fields were set
			if !schemaOidUpperSet {
				conNameUpper = fmt.Sprintf("%s%o", conNameUpper, rune(0))
			} else {
				schemaOidUpper = schemaOidUpper + 1
			}
			lt = &pgConstraint{
				name:            conNameUpper,
				schemaOidNative: schemaOidUpper,
			}
		}
	default:
		panic("unknown index name: " + index.(pgCatalogInMemIndex).name)
	}

	return gte, hasLowerBound, lt, hasUpperBound
}

// cachePgConstraints caches the pg_constraint data for the current database in the session.
func cachePgConstraints(ctx *sql.Context, pgCatalogCache *pgCatalogCache) error {
	var constraints []*pgConstraint
	tableOIDs := make(map[id.Id]map[string]id.Id)
	tableColToIdxMap := make(map[string]int16)
	oidIdx := NewUniqueInMemIndexStorage[*pgConstraint](lessConstraintOid)
	conrelidIdx := NewNonUniqueInMemIndexStorage[*pgConstraint](lessConstraintConrelid)
	connameNspIdx := NewUniqueInMemIndexStorage[*pgConstraint](lessConstraintConnameNsp)

	// We iterate over all tables first to obtain their OIDs, which we'll need to reference for foreign keys
	err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		Table: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable) (cont bool, err error) {
			inner, ok := tableOIDs[schema.OID.AsId()]
			if !ok {
				inner = make(map[string]id.Id)
				tableOIDs[schema.OID.AsId()] = inner
			}
			inner[table.Item.Name()] = table.OID.AsId()

			for i, col := range table.Item.Schema() {
				tableColToIdxMap[fmt.Sprintf("%s.%s", table.Item.Name(), col.Name)] = int16(i + 1)
			}
			return true, nil
		},
	})
	if err != nil {
		return err
	}

	// Then we iterate over everything to fill our constraints
	err = functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		Check: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, check functions.ItemCheck) (cont bool, err error) {
			constraint := &pgConstraint{
				oid:             check.OID.AsId(),
				oidNative:       id.Cache().ToOID(check.OID.AsId()),
				name:            check.Item.Name,
				schemaOid:       schema.OID.AsId(),
				schemaOidNative: id.Cache().ToOID(schema.OID.AsId()),
				conType:         "c",
				tableOid:        table.OID.AsId(),
				tableOidNative:  id.Cache().ToOID(table.OID.AsId()),
				idxOid:          id.Null,
				tableRefOid:     id.Null,
			}
			oidIdx.Add(constraint)
			conrelidIdx.Add(constraint)
			connameNspIdx.Add(constraint)
			constraints = append(constraints, constraint)
			return true, nil
		},
		ForeignKey: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, foreignKey functions.ItemForeignKey) (cont bool, err error) {
			conKey := make([]any, len(foreignKey.Item.Columns))
			for i, expr := range foreignKey.Item.Columns {
				conKey[i] = tableColToIdxMap[expr]
			}

			parentTableColToIdxMap := make(map[string]int16)
			parentTable, ok, err := schema.Item.GetTableInsensitive(ctx, foreignKey.Item.ParentTable)
			if err != nil {
				return false, err
			} else if ok {
				for i, col := range parentTable.Schema() {
					parentTableColToIdxMap[col.Name] = int16(i + 1)
				}
			}

			conFkey := make([]any, len(foreignKey.Item.ParentColumns))
			for i, expr := range foreignKey.Item.ParentColumns {
				conFkey[i] = parentTableColToIdxMap[expr]
			}

			constraint := &pgConstraint{
				oid:             foreignKey.OID.AsId(),
				oidNative:       id.Cache().ToOID(foreignKey.OID.AsId()),
				name:            foreignKey.Item.Name,
				schemaOid:       schema.OID.AsId(),
				schemaOidNative: id.Cache().ToOID(schema.OID.AsId()),
				conType:         "f",
				tableOid:        tableOIDs[schema.OID.AsId()][foreignKey.Item.Table],
				tableOidNative:  id.Cache().ToOID(tableOIDs[schema.OID.AsId()][foreignKey.Item.Table]),
				idxOid:          foreignKey.OID.AsId(),
				tableRefOid:     tableOIDs[schema.OID.AsId()][foreignKey.Item.ParentTable],
				fkUpdateType:    getFKAction(foreignKey.Item.OnUpdate),
				fkDeleteType:    getFKAction(foreignKey.Item.OnDelete),
				fkMatchType:     "s",
				conKey:          conKey,
				conFkey:         conFkey,
			}
			oidIdx.Add(constraint)
			conrelidIdx.Add(constraint)
			connameNspIdx.Add(constraint)
			constraints = append(constraints, constraint)
			return true, nil
		},
		Index: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, index functions.ItemIndex) (cont bool, err error) {
			conType := "p"
			if index.Item.ID() != "PRIMARY" {
				if index.Item.IsUnique() {
					conType = "u"
				} else {
					// If this isn't a primary key or a unique index, then it's a regular index, and not
					// a constraint, so we don't need to report it in the pg_constraint table.
					return true, nil
				}
			}

			conKey := make([]any, len(index.Item.Expressions()))
			for i, expr := range index.Item.Expressions() {
				conKey[i] = tableColToIdxMap[expr]
			}

			constraint := &pgConstraint{
				oid:             index.OID.AsId(),
				oidNative:       id.Cache().ToOID(index.OID.AsId()),
				name:            formatIndexName(index.Item),
				schemaOid:       schema.OID.AsId(),
				schemaOidNative: id.Cache().ToOID(schema.OID.AsId()),
				conType:         conType,
				tableOid:        table.OID.AsId(),
				tableOidNative:  id.Cache().ToOID(table.OID.AsId()),
				idxOid:          index.OID.AsId(),
				tableRefOid:     id.Null,
				conKey:          conKey,
				conFkey:         nil,
			}
			oidIdx.Add(constraint)
			conrelidIdx.Add(constraint)
			connameNspIdx.Add(constraint)
			constraints = append(constraints, constraint)
			return true, nil
		},
	})
	if err != nil {
		return err
	}

	pgCatalogCache.pgConstraints = &pgConstraintCache{
		constraints:     constraints,
		oidIdx:         oidIdx,
		conrelidIdx:    conrelidIdx,
		connameNspIdx:  connameNspIdx,
	}

	return nil
}

// lessConstraintOid is a sort function for pgConstraint based on oid.
func lessConstraintOid(a, b *pgConstraint) bool {
	return a.oidNative < b.oidNative
}

// lessConstraintConrelid is a sort function for pgConstraint based on conrelid.
func lessConstraintConrelid(a, b []*pgConstraint) bool {
	return a[0].tableOidNative < b[0].tableOidNative
}

// lessConstraintConnameNsp is a sort function for pgConstraint based on conname, then schemaOid.
func lessConstraintConnameNsp(a, b *pgConstraint) bool {
	if a.name == b.name {
		return a.schemaOidNative < b.schemaOidNative
	}
	return a.name < b.name
}


// PgConstraintSchema is the schema for pg_constraint.
var PgConstraintSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "connamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "contype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "condeferrable", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "condeferred", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "convalidated", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "contypid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conindid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conparentid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confupdtype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confdeltype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confmatchtype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conislocal", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "coninhcount", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "connoinherit", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conkey", Type: pgtypes.Int16Array, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "confkey", Type: pgtypes.Int16Array, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "conpfeqop", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "conppeqop", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "conffeqop", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "confdelsetcols", Type: pgtypes.Int16Array, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "conexclop", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "conbin", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgConstraintName}, // TODO: type pg_node_tree, collation C
}

// pgConstraint is the struct for the pg_constraint table.
// We store oids in their native format as well so that we can do range scans on them.
type pgConstraint struct {
	oid             id.Id
	oidNative       uint32
	name            string
	schemaOid       id.Id
	schemaOidNative uint32
	conType         string // c = check constraint, f = foreign key constraint, p = primary key constraint, u = unique constraint, t = constraint trigger, x = exclusion constraint
	tableOid        id.Id
	tableOidNative  uint32
	// typeOid      id.Id
	idxOid       id.Id
	tableRefOid  id.Id
	fkUpdateType string // a = no action, r = restrict, c = cascade, n = set null, d = set default
	fkDeleteType string // a = no action, r = restrict, c = cascade, n = set null, d = set default
	fkMatchType  string // f = full, p = partial, s = simple
	conKey       []any
	conFkey      []any
}

// pgConstraintTableScanIter is the sql.RowIter for the pg_constraint table.
type pgConstraintTableScanIter struct {
	constraintCache *pgConstraintCache
	idx             int
}

var _ sql.RowIter = (*pgConstraintTableScanIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgConstraintTableScanIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.constraintCache.constraints) {
		return nil, io.EOF
	}
	iter.idx++
	constraint := iter.constraintCache.constraints[iter.idx-1]

	return pgConstraintToRow(constraint), nil
}

// Close implements the interface sql.RowIter.
func (iter *pgConstraintTableScanIter) Close(ctx *sql.Context) error {
	return nil
}

// pgConstraintToRow converts a pgConstraint to a sql.Row.
func pgConstraintToRow(constraint *pgConstraint) sql.Row {
	var conKey interface{}
	if len(constraint.conKey) == 0 {
		conKey = nil
	} else {
		conKey = constraint.conKey
	}

	var conFkey interface{}
	if len(constraint.conFkey) == 0 {
		conFkey = nil
	} else {
		conFkey = constraint.conFkey
	}

	return sql.Row{
		constraint.oid,          // oid
		constraint.name,          // conname
		constraint.schemaOid,     // connamespace
		constraint.conType,       // contype
		false,                    // condeferrable
		false,                    // condeferred
		true,                     // convalidated
		constraint.tableOid,      // conrelid
		id.Null,                  // contypid
		constraint.idxOid,        // conindid
		id.Null,                  // conparentid
		constraint.tableRefOid,   // confrelid
		constraint.fkUpdateType,  // confupdtype
		constraint.fkDeleteType,  // confdeltype
		constraint.fkMatchType,   // confmatchtype
		true,                     // conislocal
		int16(0),                 // coninhcount
		true,                     // connoinherit
		conKey,                   // conkey
		conFkey,                  // confkey
		nil,                      // conpfeqop
		nil,                      // conppeqop
		nil,                      // conffeqop
		nil,                      // confdelsetcols
		nil,                      // conexclop
		nil,                      // conbin
	}
}
