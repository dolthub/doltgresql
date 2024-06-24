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

	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgTsTemplateName is a constant to the pg_ts_template name.
const PgTsTemplateName = "pg_ts_template"

// InitPgTsTemplate handles registration of the pg_ts_template handler.
func InitPgTsTemplate() {
	tables.AddHandler(PgCatalogName, PgTsTemplateName, PgTsTemplateHandler{})
}

// PgTsTemplateHandler is the handler for the pg_ts_template table.
type PgTsTemplateHandler struct{}

var _ tables.Handler = PgTsTemplateHandler{}

// Name implements the interface tables.Handler.
func (p PgTsTemplateHandler) Name() string {
	return PgTsTemplateName
}

// RowIter implements the interface tables.Handler.
func (p PgTsTemplateHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_ts_template row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgTsTemplateHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgTsTemplateSchema,
		PkOrdinals: nil,
	}
}

// pgTsTemplateSchema is the schema for pg_ts_template.
var pgTsTemplateSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsTemplateName},
	{Name: "tmplname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgTsTemplateName},
	{Name: "tmplnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsTemplateName},
	{Name: "tmplinit", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTsTemplateName},   // TODO: regproc type
	{Name: "tmpllexize", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTsTemplateName}, // TODO: regproc type
}

// pgTsTemplateRowIter is the sql.RowIter for the pg_ts_template table.
type pgTsTemplateRowIter struct {
}

var _ sql.RowIter = (*pgTsTemplateRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTsTemplateRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgTsTemplateRowIter) Close(ctx *sql.Context) error {
	return nil
}
