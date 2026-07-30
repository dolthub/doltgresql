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
func (p PgLanguageHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	return &pgLanguageRowIter{
		languages: defaultPostgresLanguages,
		idx:       0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgLanguageHandler) PkSchema() sql.PrimaryKeySchema {
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
	languages []language
	idx       int
}

var _ sql.RowIter = (*pgLanguageRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgLanguageRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.languages) {
		return nil, io.EOF
	}
	iter.idx++
	lang := iter.languages[iter.idx-1]

	return sql.Row{
		lang.oid,       // oid
		lang.name,      // lanname
		id.Null,        // lanowner (TODO: owner)
		lang.isPl,      // lanispl
		lang.plTrusted, // lanpltrusted
		id.Null,        // lanplcallfoid (TODO: call handler function)
		id.Null,        // laninline (TODO: inline handler function)
		id.Null,        // lanvalidator (TODO: validator function)
		nil,            // lanacl
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgLanguageRowIter) Close(ctx *sql.Context) error {
	return nil
}

// language represents a row in the pg_language table.
type language struct {
	oid       id.Id
	name      string
	isPl      bool
	plTrusted bool
}

// defaultPostgresLanguages is the list of built-in function languages available in Postgres.
var defaultPostgresLanguages = []language{
	{oid: id.NewFunctionLanguage("internal").AsId(), name: "internal", isPl: false, plTrusted: false},
	{oid: id.NewFunctionLanguage("c").AsId(), name: "c", isPl: false, plTrusted: false},
	{oid: id.NewFunctionLanguage("sql").AsId(), name: "sql", isPl: false, plTrusted: true},
	{oid: id.NewFunctionLanguage("plpgsql").AsId(), name: "plpgsql", isPl: true, plTrusted: true},
}
