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

// Name implements the interface tables.Handler.
func (p PgAttributeHandler) Name() string {
	return PgAttributeName
}

// RowIter implements the interface tables.Handler.
func (p PgAttributeHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// Use cached data from this process if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.attributeCols == nil {
		var cols []*sql.Column
		var tableOIDs []id.Id
		var colIdxs []int

		err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
			Table: func(ctx *sql.Context, _ functions.ItemSchema, table functions.ItemTable) (cont bool, err error) {
				for i, col := range table.Item.Schema() {
					cols = append(cols, col)
					colIdxs = append(colIdxs, i)
					tableOIDs = append(tableOIDs, table.OID.AsId())
				}
				return true, nil
			},
		})
		if err != nil {
			return nil, err
		}

		if includeSystemTables {
			_, root, err := core.GetRootFromContext(ctx)
			if err != nil {
				return nil, err
			}

			systemTables, err := resolve.GetGeneratedSystemTables(ctx, root)
			if err != nil {
				return nil, err
			}

			db := ctx.GetCurrentDatabase()
			for _, tblName := range systemTables {
				tbl, err := core.GetSqlTableFromContext(ctx, db, tblName)
				if err != nil {
					// Some of the system tables exist conditionally when accessed, so just skip them in this case
					if errors.Is(doltdb.ErrTableNotFound, err) {
						continue
					}
					return nil, err
				}

				schema := tbl.Schema()
				for i, col := range schema {
					cols = append(cols, col)
					colIdxs = append(colIdxs, i)
					tableOIDs = append(tableOIDs, id.NewTable(tblName.Schema, tblName.Name).AsId())
				}
			}
		}

		pgCatalogCache.attributeCols = cols
		pgCatalogCache.attributeColIdxs = colIdxs
		pgCatalogCache.attributeTableOIDs = tableOIDs
	}

	return &pgAttributeRowIter{
		cols:      pgCatalogCache.attributeCols,
		colIdxs:   pgCatalogCache.attributeColIdxs,
		tableOIDs: pgCatalogCache.attributeTableOIDs,
		idx:       0,
	}, nil
}

// Schema implements the interface tables.Handler.
func (p PgAttributeHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAttributeSchema,
		PkOrdinals: nil,
	}
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

// pgAttributeRowIter is the sql.RowIter for the pg_attribute table.
type pgAttributeRowIter struct {
	cols      []*sql.Column
	colIdxs   []int
	tableOIDs []id.Id
	idx       int
}

var _ sql.RowIter = (*pgAttributeRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAttributeRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.cols) {
		return nil, io.EOF
	}
	iter.idx++
	col := iter.cols[iter.idx-1]
	tableOid := iter.tableOIDs[iter.idx-1]
	colIdx := iter.colIdxs[iter.idx-1]

	generated := ""
	if col.Generated != nil {
		generated = "s"
	}

	dimensions := 0
	if s, ok := col.Type.(sql.SetType); ok {
		dimensions = int(s.NumberOfElements())
	}

	hasDefault := col.Default != nil

	typeOid := id.Null
	if doltgresType, ok := col.Type.(*pgtypes.DoltgresType); ok {
		typeOid = doltgresType.ID.AsId()
	} else {
		// TODO: Remove once all information_schema tables are converted to use DoltgresType
		dt := pgtypes.FromGmsType(col.Type)
		typeOid = dt.ID.AsId()
	}

	// TODO: Fill in the rest of the pg_attribute columns
	return sql.Row{
		tableOid,          // attrelid
		col.Name,          // attname
		typeOid,           // atttypid
		int16(0),          // attlen
		int16(colIdx + 1), // attnum
		int32(-1),         // attcacheoff
		int32(-1),         // atttypmod
		int16(dimensions), // attndims
		false,             // attbyval
		"i",               // attalign
		"p",               // attstorage
		"",                // attcompression
		!col.Nullable,     // attnotnull
		hasDefault,        // atthasdef
		false,             // atthasmissing
		"",                // attidentity
		generated,         // attgenerated
		false,             // attisdropped
		true,              // attislocal
		int16(0),          // attinhcount
		int16(-1),         // attstattarget
		id.Null,           // attcollation
		nil,               // attacl
		nil,               // attoptions
		nil,               // attfdwoptions
		nil,               // attmissingval
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgAttributeRowIter) Close(ctx *sql.Context) error {
	return nil
}
