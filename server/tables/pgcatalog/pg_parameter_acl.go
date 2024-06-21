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

// PgParameterAclName is a constant to the pg_parameter_acl name.
const PgParameterAclName = "pg_parameter_acl"

// InitPgParameterAcl handles registration of the pg_parameter_acl handler.
func InitPgParameterAcl() {
	tables.AddHandler(PgCatalogName, PgParameterAclName, PgParameterAclHandler{})
}

// PgParameterAclHandler is the handler for the pg_parameter_acl table.
type PgParameterAclHandler struct{}

var _ tables.Handler = PgParameterAclHandler{}

// Name implements the interface tables.Handler.
func (p PgParameterAclHandler) Name() string {
	return PgParameterAclName
}

// RowIter implements the interface tables.Handler.
func (p PgParameterAclHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_parameter_acl row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgParameterAclHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgParameterAclSchema,
		PkOrdinals: nil,
	}
}

// pgParameterAclSchema is the schema for pg_parameter_acl.
var pgParameterAclSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgParameterAclName},
	{Name: "parname", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgParameterAclName},    // TODO: collation C
	{Name: "paracl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgParameterAclName}, // TODO: aclitem[] type
}

// pgParameterAclRowIter is the sql.RowIter for the pg_parameter_acl table.
type pgParameterAclRowIter struct {
}

var _ sql.RowIter = (*pgParameterAclRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgParameterAclRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgParameterAclRowIter) Close(ctx *sql.Context) error {
	return nil
}
