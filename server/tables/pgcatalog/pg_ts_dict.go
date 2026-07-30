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

// PgTsDictName is a constant to the pg_ts_dict name.
const PgTsDictName = "pg_ts_dict"

// InitPgTsDict handles registration of the pg_ts_dict handler.
func InitPgTsDict() {
	tables.AddHandler(PgCatalogName, PgTsDictName, PgTsDictHandler{})
}

// PgTsDictHandler is the handler for the pg_ts_dict table.
type PgTsDictHandler struct{}

var _ tables.Handler = PgTsDictHandler{}

// Name implements the interface tables.Handler.
func (p PgTsDictHandler) Name() string {
	return PgTsDictName
}

// RowIter implements the interface tables.Handler.
func (p PgTsDictHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: this only includes the "simple" dictionary, since the language-specific stemming dictionaries
	//  (english_stem, etc.) rely on snowball support that Doltgres does not yet have
	return &pgTsDictRowIter{
		dicts: defaultPostgresTsDicts,
		idx:   0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgTsDictHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgTsDictSchema,
		PkOrdinals: nil,
	}
}

// pgTsDictSchema is the schema for pg_ts_dict.
var pgTsDictSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsDictName},
	{Name: "dictname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgTsDictName},
	{Name: "dictnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsDictName},
	{Name: "dictowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsDictName},
	{Name: "dicttemplate", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsDictName},
	{Name: "dictinitoption", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgTsDictName}, // TODO: collation C
}

// pgTsDictRowIter is the sql.RowIter for the pg_ts_dict table.
type pgTsDictRowIter struct {
	dicts []tsDict
	idx   int
}

var _ sql.RowIter = (*pgTsDictRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTsDictRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.dicts) {
		return nil, io.EOF
	}
	iter.idx++
	dict := iter.dicts[iter.idx-1]

	return sql.Row{
		dict.oid,       // oid
		dict.name,      // dictname
		dict.namespace, // dictnamespace
		id.Null,        // dictowner (TODO: owner)
		dict.template,  // dicttemplate
		nil,            // dictinitoption
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgTsDictRowIter) Close(ctx *sql.Context) error {
	return nil
}

// tsDict represents a row in the pg_ts_dict table.
type tsDict struct {
	oid       id.Id
	name      string
	namespace id.Id
	template  id.Id
}

// defaultPostgresTsDicts is the list of built-in text search dictionaries available in Postgres.
var defaultPostgresTsDicts = []tsDict{
	{
		oid:       id.NewId(id.Section_TextSearchDictionary, "pg_catalog", "simple"),
		name:      "simple",
		namespace: id.NewNamespace("pg_catalog").AsId(),
		template:  id.NewId(id.Section_TextSearchTemplate, "pg_catalog", "simple"),
	},
}
