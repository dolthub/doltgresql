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

// PgPoliciesName is a constant to the pg_policies name.
const PgPoliciesName = "pg_policies"

// InitPgPolicies handles registration of the pg_policies handler.
func InitPgPolicies() {
	tables.AddHandler(PgCatalogName, PgPoliciesName, PgPoliciesHandler{})
}

// PgPoliciesHandler is the handler for the pg_policies table.
type PgPoliciesHandler struct{}

var _ tables.Handler = PgPoliciesHandler{}

// Name implements the interface tables.Handler.
func (p PgPoliciesHandler) Name() string {
	return PgPoliciesName
}

// RowIter implements the interface tables.Handler.
func (p PgPoliciesHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_policies row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgPoliciesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgPoliciesSchema,
		PkOrdinals: nil,
	}
}

// pgPoliciesSchema is the schema for pg_policies.
var pgPoliciesSchema = sql.Schema{
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgPoliciesName},
	{Name: "tablename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgPoliciesName},
	{Name: "policyname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgPoliciesName},
	{Name: "permissive", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgPoliciesName},
	{Name: "roles", Type: pgtypes.NameArray, Default: nil, Nullable: true, Source: PgPoliciesName},
	{Name: "cmd", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgPoliciesName},
	{Name: "qual", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgPoliciesName},       // TODO: collation C
	{Name: "with_check", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgPoliciesName}, // TODO: collation C
}

// pgPoliciesRowIter is the sql.RowIter for the pg_policies table.
type pgPoliciesRowIter struct {
}

var _ sql.RowIter = (*pgPoliciesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgPoliciesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgPoliciesRowIter) Close(ctx *sql.Context) error {
	return nil
}
