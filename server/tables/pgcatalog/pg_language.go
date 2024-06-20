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

// PgLanguageName is a constant to the pg_language name.
const PgLanguageName = "pg_language"

// InitPgLanguage handles registration of the pg_language handler.
func InitPgLanguage() {
	tables.AddHandler(PgCatalogName, PgLanguageName, PgLanguageHandler{})
}

// PgLanguageHandler is the handler for the pg_language table.
type PgLanguageHandler struct{}

var _ tables.Handler = PgLanguageHandler{}

// Name implements the interface tables.Handler.
func (p PgLanguageHandler) Name() string {
	return PgLanguageName
}

// RowIter implements the interface tables.Handler.
func (p PgLanguageHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_language row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgLanguageHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgLanguageSchema,
		PkOrdinals: nil,
	}
}

// pgLanguageSchema is the schema for pg_language.
var pgLanguageSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgLanguageName},
	{Name: "lanname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgLanguageName},
	{Name: "lanowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgLanguageName},
	{Name: "lanispl", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgLanguageName},
	{Name: "lanpltrusted", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgLanguageName},
	{Name: "lanplcallfoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgLanguageName},
	{Name: "laninline", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgLanguageName},
	{Name: "lanvalidator", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgLanguageName},
	{Name: "lanacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgLanguageName}, // TODO: aclitem[] type
}

// pgLanguageRowIter is the sql.RowIter for the pg_language table.
type pgLanguageRowIter struct {
}

var _ sql.RowIter = (*pgLanguageRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgLanguageRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgLanguageRowIter) Close(ctx *sql.Context) error {
	return nil
}
