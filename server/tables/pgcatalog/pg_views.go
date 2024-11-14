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

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/types/oid"
)

// PgViewsName is a constant to the pg_views name.
const PgViewsName = "pg_views"

// InitPgViews handles registration of the pg_views handler.
func InitPgViews() {
	tables.AddHandler(PgCatalogName, PgViewsName, PgViewsHandler{})
}

// PgViewsHandler is the handler for the pg_views table.
type PgViewsHandler struct{}

var _ tables.Handler = PgViewsHandler{}

// Name implements the interface tables.Handler.
func (p PgViewsHandler) Name() string {
	return PgViewsName
}

// RowIter implements the interface tables.Handler.
func (p PgViewsHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// Use cached data from this process if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.views == nil {
		var views []sql.ViewDefinition
		var viewSchemas []string
		err := oid.IterateCurrentDatabase(ctx, oid.Callbacks{
			View: func(ctx *sql.Context, schema oid.ItemSchema, view oid.ItemView) (cont bool, err error) {
				views = append(views, view.Item)
				viewSchemas = append(viewSchemas, schema.Item.SchemaName())
				return true, nil
			},
		})
		if err != nil {
			return nil, err
		}
		pgCatalogCache.views = views
		pgCatalogCache.viewSchemas = viewSchemas
	}

	return &pgViewsRowIter{
		views:   pgCatalogCache.views,
		schemas: pgCatalogCache.viewSchemas,
		idx:     0,
	}, nil
}

// Schema implements the interface tables.Handler.
func (p PgViewsHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgViewsSchema,
		PkOrdinals: nil,
	}
}

// pgViewsSchema is the schema for pg_views.
var pgViewsSchema = sql.Schema{
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgViewsName},
	{Name: "viewname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgViewsName},
	{Name: "viewowner", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgViewsName},
	{Name: "definition", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgViewsName},
}

// pgViewsRowIter is the sql.RowIter for the pg_views table.
type pgViewsRowIter struct {
	views   []sql.ViewDefinition
	schemas []string
	idx     int
}

var _ sql.RowIter = (*pgViewsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgViewsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.views) {
		return nil, io.EOF
	}
	iter.idx++
	view := iter.views[iter.idx-1]
	schema := iter.schemas[iter.idx-1]

	textDef := view.TextDefinition
	if textDef == "" {
		stmts, err := parser.Parse(view.CreateViewStatement)
		if err != nil {
			return nil, err
		}
		if len(stmts) == 0 {
			return nil, fmt.Errorf("expected Create View statement, got none")
		}
		cv, ok := stmts[0].AST.(*tree.CreateView)
		if !ok {
			return nil, fmt.Errorf("expected Create View statement, got %s", stmts[0].SQL)
		}

		textDef = cv.AsSource.String()
	}

	return sql.Row{
		schema,    // schemaname
		view.Name, // viewname
		"",        // viewowner
		textDef,   // definition
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgViewsRowIter) Close(ctx *sql.Context) error {
	return nil
}
