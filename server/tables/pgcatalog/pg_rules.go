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

// PgRulesName is a constant to the pg_rules name.
const PgRulesName = "pg_rules"

// InitPgRules handles registration of the pg_rules handler.
func InitPgRules() {
	tables.AddHandler(PgCatalogName, PgRulesName, PgRulesHandler{})
}

// PgRulesHandler is the handler for the pg_rules table.
type PgRulesHandler struct{}

var _ tables.Handler = PgRulesHandler{}

// Name implements the interface tables.Handler.
func (p PgRulesHandler) Name() string {
	return PgRulesName
}

// RowIter implements the interface tables.Handler.
func (p PgRulesHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_rules row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgRulesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgRulesSchema,
		PkOrdinals: nil,
	}
}

// pgRulesSchema is the schema for pg_rules.
var pgRulesSchema = sql.Schema{
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgRulesName},
	{Name: "tablename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgRulesName},
	{Name: "rulename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgRulesName},
	{Name: "definition", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgRulesName},
}

// pgRulesRowIter is the sql.RowIter for the pg_rules table.
type pgRulesRowIter struct {
}

var _ sql.RowIter = (*pgRulesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgRulesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgRulesRowIter) Close(ctx *sql.Context) error {
	return nil
}
