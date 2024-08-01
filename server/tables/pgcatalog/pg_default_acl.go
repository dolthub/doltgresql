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

// PgDefaultAclName is a constant to the pg_default_acl name.
const PgDefaultAclName = "pg_default_acl"

// InitPgDefaultAcl handles registration of the pg_default_acl handler.
func InitPgDefaultAcl() {
	tables.AddHandler(PgCatalogName, PgDefaultAclName, PgDefaultAclHandler{})
}

// PgDefaultAclHandler is the handler for the pg_default_acl table.
type PgDefaultAclHandler struct{}

var _ tables.Handler = PgDefaultAclHandler{}

// Name implements the interface tables.Handler.
func (p PgDefaultAclHandler) Name() string {
	return PgDefaultAclName
}

// RowIter implements the interface tables.Handler.
func (p PgDefaultAclHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_default_acl row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgDefaultAclHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgDefaultAclSchema,
		PkOrdinals: nil,
	}
}

// pgDefaultAclSchema is the schema for pg_default_acl.
var pgDefaultAclSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDefaultAclName},
	{Name: "defaclrole", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDefaultAclName},
	{Name: "defaclnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDefaultAclName},
	{Name: "defaclobjtype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgDefaultAclName},
	{Name: "defaclacl", Type: pgtypes.TextArray, Default: nil, Nullable: false, Source: PgDefaultAclName}, // TODO: aclitem[] type
}

// pgDefaultAclRowIter is the sql.RowIter for the pg_default_acl table.
type pgDefaultAclRowIter struct {
}

var _ sql.RowIter = (*pgDefaultAclRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgDefaultAclRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgDefaultAclRowIter) Close(ctx *sql.Context) error {
	return nil
}
