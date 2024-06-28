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
func (p PgTsDictHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_ts_dict row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgTsDictHandler) Schema() sql.PrimaryKeySchema {
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
}

var _ sql.RowIter = (*pgTsDictRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTsDictRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgTsDictRowIter) Close(ctx *sql.Context) error {
	return nil
}
