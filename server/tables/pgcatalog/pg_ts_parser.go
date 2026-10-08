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

// PgTsParserName is a constant to the pg_ts_parser name.
const PgTsParserName = "pg_ts_parser"

// InitPgTsParser handles registration of the pg_ts_parser handler.
func InitPgTsParser() {
	tables.AddHandler(PgCatalogName, PgTsParserName, PgTsParserHandler{})
}

// PgTsParserHandler is the handler for the pg_ts_parser table.
type PgTsParserHandler struct{}

var _ tables.Handler = PgTsParserHandler{}

// Name implements the interface tables.Handler.
func (p PgTsParserHandler) Name() string {
	return PgTsParserName
}

// RowIter implements the interface tables.Handler.
func (p PgTsParserHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	return &pgTsParserRowIter{
		parsers: defaultPostgresTsParsers,
		idx:     0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgTsParserHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgTsParserSchema,
		PkOrdinals: nil,
	}
}

// pgTsParserSchema is the schema for pg_ts_parser.
var pgTsParserSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsParserName},
	{Name: "prsname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgTsParserName},
	{Name: "prsnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsParserName},
	{Name: "prsstart", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgTsParserName},
	{Name: "prstoken", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgTsParserName},
	{Name: "prsend", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgTsParserName},
	{Name: "prsheadline", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgTsParserName},
	{Name: "prslextype", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgTsParserName},
}

// pgTsParserRowIter is the sql.RowIter for the pg_ts_parser table.
type pgTsParserRowIter struct {
	parsers []tsParser
	idx     int
}

var _ sql.RowIter = (*pgTsParserRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTsParserRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.parsers) {
		return nil, io.EOF
	}
	iter.idx++
	parser := iter.parsers[iter.idx-1]

	return sql.Row{
		parser.oid,       // oid
		parser.name,      // prsname
		parser.namespace, // prsnamespace
		parser.start,     // prsstart
		parser.token,     // prstoken
		parser.end,       // prsend
		parser.headline,  // prsheadline
		parser.lextype,   // prslextype
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgTsParserRowIter) Close(ctx *sql.Context) error {
	return nil
}

// tsParser represents a row in the pg_ts_parser table.
type tsParser struct {
	oid       id.Id
	name      string
	namespace id.Id
	start     id.Id
	token     id.Id
	end       id.Id
	headline  id.Id
	lextype   id.Id
}

// defaultPostgresTsParsers is the list of built-in text search parsers available in Postgres.
var defaultPostgresTsParsers = []tsParser{
	{
		oid:       id.NewId(id.Section_TextSearchParser, "pg_catalog", "default"),
		name:      "default",
		namespace: id.NewNamespace("pg_catalog").AsId(),
		start:     id.NewFunction("pg_catalog", "prsd_start", pgtypes.Internal.ID).AsId(),
		token:     id.NewFunction("pg_catalog", "prsd_nexttoken", pgtypes.Internal.ID).AsId(),
		end:       id.NewFunction("pg_catalog", "prsd_end", pgtypes.Internal.ID).AsId(),
		headline:  id.NewFunction("pg_catalog", "prsd_headline", pgtypes.Internal.ID).AsId(),
		lextype:   id.NewFunction("pg_catalog", "prsd_lextype", pgtypes.Internal.ID).AsId(),
	},
}
