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

// PgHbaFileRulesName is a constant to the pg_hba_file_rules name.
const PgHbaFileRulesName = "pg_hba_file_rules"

// InitPgHbaFileRules handles registration of the pg_hba_file_rules handler.
func InitPgHbaFileRules() {
	tables.AddHandler(PgCatalogName, PgHbaFileRulesName, PgHbaFileRulesHandler{})
}

// PgHbaFileRulesHandler is the handler for the pg_hba_file_rules table.
type PgHbaFileRulesHandler struct{}

var _ tables.Handler = PgHbaFileRulesHandler{}

// Name implements the interface tables.Handler.
func (p PgHbaFileRulesHandler) Name() string {
	return PgHbaFileRulesName
}

// RowIter implements the interface tables.Handler.
func (p PgHbaFileRulesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_hba_file_rules row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgHbaFileRulesHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgHbaFileRulesSchema,
		PkOrdinals: nil,
	}
}

// pgHbaFileRulesSchema is the schema for pg_hba_file_rules.
var pgHbaFileRulesSchema = sql.Schema{
	{Name: "line_number", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgHbaFileRulesName},
	{Name: "type", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgHbaFileRulesName},
	{Name: "database", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgHbaFileRulesName},
	{Name: "user_name", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgHbaFileRulesName},
	{Name: "address", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgHbaFileRulesName},
	{Name: "netmask", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgHbaFileRulesName},
	{Name: "auth_method", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgHbaFileRulesName},
	{Name: "options", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgHbaFileRulesName},
	{Name: "error", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgHbaFileRulesName},
}

// pgHbaFileRulesRowIter is the sql.RowIter for the pg_hba_file_rules table.
type pgHbaFileRulesRowIter struct {
}

var _ sql.RowIter = (*pgHbaFileRulesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgHbaFileRulesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgHbaFileRulesRowIter) Close(ctx *sql.Context) error {
	return nil
}
