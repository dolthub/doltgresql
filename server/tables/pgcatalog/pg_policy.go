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

// PgPolicyName is a constant to the pg_policy name.
const PgPolicyName = "pg_policy"

// InitPgPolicy handles registration of the pg_policy handler.
func InitPgPolicy() {
	tables.AddHandler(PgCatalogName, PgPolicyName, PgPolicyHandler{})
}

// PgPolicyHandler is the handler for the pg_policy table.
type PgPolicyHandler struct{}

var _ tables.Handler = PgPolicyHandler{}

// Name implements the interface tables.Handler.
func (p PgPolicyHandler) Name() string {
	return PgPolicyName
}

// RowIter implements the interface tables.Handler.
func (p PgPolicyHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_policy row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgPolicyHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgPolicySchema,
		PkOrdinals: nil,
	}
}

// pgPolicySchema is the schema for pg_policy.
var pgPolicySchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgPolicyName},
	{Name: "polname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgPolicyName},
	{Name: "polrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgPolicyName},
	{Name: "polcmd", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgPolicyName},
	{Name: "polpermissive", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgPolicyName},
	{Name: "polroles", Type: pgtypes.OidArray, Default: nil, Nullable: false, Source: PgPolicyName},
	{Name: "polqual", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgPolicyName},      // TODO: pg_node_tree type, collation C
	{Name: "polwithcheck", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgPolicyName}, // TODO: pg_node_tree type, collation C
}

// pgPolicyRowIter is the sql.RowIter for the pg_policy table.
type pgPolicyRowIter struct {
}

var _ sql.RowIter = (*pgPolicyRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgPolicyRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgPolicyRowIter) Close(ctx *sql.Context) error {
	return nil
}
