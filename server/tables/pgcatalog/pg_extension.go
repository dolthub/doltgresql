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

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/extensions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgExtensionName is a constant to the pg_extension name.
const PgExtensionName = "pg_extension"

// InitPgExtension handles registration of the pg_extension handler.
func InitPgExtension() {
	tables.AddHandler(PgCatalogName, PgExtensionName, PgExtensionHandler{})
}

// PgExtensionHandler is the handler for the pg_extension table.
type PgExtensionHandler struct{}

var _ tables.Handler = PgExtensionHandler{}

// Name implements the interface tables.Handler.
func (p PgExtensionHandler) Name() string {
	return PgExtensionName
}

// RowIter implements the interface tables.Handler.
func (p PgExtensionHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// Use cached data from this process if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.extensions == nil {
		err = cachePgExtensions(ctx, pgCatalogCache)
		if err != nil {
			return nil, err
		}
	}

	return &pgExtensionRowIter{
		extensions: pgCatalogCache.extensions,
		idx:        0,
	}, nil
}

// cachePgExtensions caches the pg_extension data for the current database in the session.
func cachePgExtensions(ctx *sql.Context, pgCatalogCache *pgCatalogCache) error {
	extCollection, err := core.GetExtensionsCollectionFromContext(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return err
	}
	var exts []extensions.Extension
	err = extCollection.IterAll(ctx, func(rootObj objinterface.RootObject) (stop bool, err error) {
		if ext, ok := rootObj.(extensions.Extension); ok {
			exts = append(exts, ext)
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	pgCatalogCache.extensions = exts
	return nil
}

// PkSchema implements the interface tables.Handler.
func (p PgExtensionHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgExtensionSchema,
		PkOrdinals: nil,
	}
}

// pgExtensionSchema is the schema for pg_extension.
var pgExtensionSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgExtensionName},
	{Name: "extname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgExtensionName},
	{Name: "extowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgExtensionName},
	{Name: "extnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgExtensionName},
	{Name: "extrelocatable", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgExtensionName},
	{Name: "extversion", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgExtensionName}, // TODO: collation C
	{Name: "extconfig", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgExtensionName},
	{Name: "extcondition", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgExtensionName}, // TODO: collation C
}

// pgExtensionRowIter is the sql.RowIter for the pg_extension table.
type pgExtensionRowIter struct {
	extensions []extensions.Extension
	idx        int
}

var _ sql.RowIter = (*pgExtensionRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgExtensionRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.extensions) {
		return nil, io.EOF
	}
	iter.idx++
	ext := iter.extensions[iter.idx-1]
	return sql.Row{
		ext.ExtName.AsId(),                   // oid
		ext.ExtName.Name(),                   // extname
		id.Null,                              // extowner // TODO: extension owner is not yet tracked
		ext.Namespace.AsId(),                 // extnamespace
		ext.Relocatable,                      // extrelocatable
		ext.LibIdentifier.Version().String(), // extversion
		nil,                                  // extconfig
		nil,                                  // extcondition
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgExtensionRowIter) Close(ctx *sql.Context) error {
	return nil
}
