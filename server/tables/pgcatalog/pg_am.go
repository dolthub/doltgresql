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

// PgAmName is a constant to the pg_am name.
const PgAmName = "pg_am"

// InitPgAm handles registration of the pg_am handler.
func InitPgAm() {
	tables.AddHandler(PgCatalogName, PgAmName, PgAmHandler{})
}

// PgAmHandler is the handler for the pg_am table.
type PgAmHandler struct{}

var _ tables.Handler = PgAmHandler{}

// Name implements the interface tables.Handler.
func (p PgAmHandler) Name() string {
	return PgAmName
}

// RowIter implements the interface tables.Handler.
func (p PgAmHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	return &pgAmRowIter{
		ams: defaultPostgresAms,
		idx: 0,
	}, nil
}

// Schema implements the interface tables.Handler.
func (p PgAmHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAmSchema,
		PkOrdinals: nil,
	}
}

// pgAmSchema is the schema for pg_am.
var pgAmSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAmName},
	{Name: "amname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgAmName},
	{Name: "amhandler", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgAmName}, // TODO: type regproc
	{Name: "amtype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgAmName},
}

// pgAmRowIter is the sql.RowIter for the pg_am table.
type pgAmRowIter struct {
	ams []accessMethod
	idx int
}

var _ sql.RowIter = (*pgAmRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAmRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.ams) {
		return nil, io.EOF
	}
	iter.idx++
	am := iter.ams[iter.idx-1]

	return sql.Row{
		am.oid,     // oid
		am.name,    // amname
		am.handler, // amhandler
		am.typ,     // amtype
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgAmRowIter) Close(ctx *sql.Context) error {
	return nil
}

type accessMethod struct {
	oid     id.Id
	name    string
	handler string
	typ     string
}

// defaultPostgresAms is the list of default access methods available in Postgres.
var defaultPostgresAms = []accessMethod{
	{oid: id.NewAccessMethod("heap").AsId(), name: "heap", handler: "heap_tableam_handler", typ: "t"},
	{oid: id.NewAccessMethod("btree").AsId(), name: "btree", handler: "bthandler", typ: "i"},
	{oid: id.NewAccessMethod("hash").AsId(), name: "hash", handler: "hashhandler", typ: "i"},
	{oid: id.NewAccessMethod("gist").AsId(), name: "gist", handler: "gisthandler", typ: "i"},
	{oid: id.NewAccessMethod("gin").AsId(), name: "gin", handler: "ginhandler", typ: "i"},
	{oid: id.NewAccessMethod("spgist").AsId(), name: "spgist", handler: "spghandler", typ: "i"},
	{oid: id.NewAccessMethod("brin").AsId(), name: "brin", handler: "brinhandler", typ: "i"},
}
