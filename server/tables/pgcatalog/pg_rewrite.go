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

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgRewriteName is a constant to the pg_rewrite name.
const PgRewriteName = "pg_rewrite"

// InitPgRewrite handles registration of the pg_rewrite handler.
func InitPgRewrite() {
	tables.AddHandler(PgCatalogName, PgRewriteName, PgRewriteHandler{})
}

// PgRewriteHandler is the handler for the pg_rewrite table.
type PgRewriteHandler struct{}

var _ tables.Handler = PgRewriteHandler{}

// Name implements the interface tables.Handler.
func (p PgRewriteHandler) Name() string {
	return PgRewriteName
}

// RowIter implements the interface tables.Handler.
func (p PgRewriteHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// Use cached data from this process if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.rewrites == nil {
		err = cachePgRewrites(ctx, pgCatalogCache)
		if err != nil {
			return nil, err
		}
	}

	return &pgRewriteRowIter{
		rewrites: pgCatalogCache.rewrites,
		idx:      0,
	}, nil
}

// cachePgRewrites caches the pg_rewrite data for the current database in the session.
func cachePgRewrites(ctx *sql.Context, pgCatalogCache *pgCatalogCache) error {
	// Doltgres does not support CREATE RULE, so the only rewrite rules that exist are the implicit
	// "_RETURN" rules that every view has.
	var rewrites []pgRewrite
	err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		View: func(ctx *sql.Context, schema functions.ItemSchema, view functions.ItemView) (cont bool, err error) {
			rewrites = append(rewrites, pgRewrite{
				// There is no dedicated id section for rewrite rules, so we use a trigger id to
				// derive an OID that is unique and distinct from the view's own OID.
				oid:     id.NewTrigger(schema.Item.SchemaName(), view.Item.Name, "_RETURN").AsId(),
				evClass: view.OID.AsId(),
			})
			return true, nil
		},
	})
	if err != nil {
		return err
	}
	pgCatalogCache.rewrites = rewrites
	return nil
}

// PkSchema implements the interface tables.Handler.
func (p PgRewriteHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgRewriteSchema,
		PkOrdinals: nil,
	}
}

// pgRewriteSchema is the schema for pg_rewrite.
var pgRewriteSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgRewriteName},
	{Name: "rulename", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgRewriteName},
	{Name: "ev_class", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgRewriteName},
	{Name: "ev_type", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgRewriteName},
	{Name: "ev_enabled", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgRewriteName},
	{Name: "is_instead", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgRewriteName},
	{Name: "ev_qual", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgRewriteName},   // TODO: pg_node_tree type, collation C
	{Name: "ev_action", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgRewriteName}, // TODO: pg_node_tree type, collation C
}

// pgRewrite represents a row in the pg_rewrite table.
type pgRewrite struct {
	oid     id.Id
	evClass id.Id
}

// pgRewriteRowIter is the sql.RowIter for the pg_rewrite table.
type pgRewriteRowIter struct {
	rewrites []pgRewrite
	idx      int
}

var _ sql.RowIter = (*pgRewriteRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgRewriteRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.rewrites) {
		return nil, io.EOF
	}
	iter.idx++
	rewrite := iter.rewrites[iter.idx-1]

	return sql.Row{
		rewrite.oid,     // oid
		"_RETURN",       // rulename
		rewrite.evClass, // ev_class
		"1",             // ev_type (SELECT)
		"O",             // ev_enabled
		true,            // is_instead
		"<>",            // ev_qual
		"<>",            // ev_action (TODO: emit the actual query node tree for the view)
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgRewriteRowIter) Close(ctx *sql.Context) error {
	return nil
}
