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

// PgDbRoleSettingName is a constant to the pg_db_role_setting name.
const PgDbRoleSettingName = "pg_db_role_setting"

// InitPgDbRoleSetting handles registration of the pg_db_role_setting handler.
func InitPgDbRoleSetting() {
	tables.AddHandler(PgCatalogName, PgDbRoleSettingName, PgDbRoleSettingHandler{})
}

// PgDbRoleSettingHandler is the handler for the pg_db_role_setting table.
type PgDbRoleSettingHandler struct{}

var _ tables.Handler = PgDbRoleSettingHandler{}

// Name implements the interface tables.Handler.
func (p PgDbRoleSettingHandler) Name() string {
	return PgDbRoleSettingName
}

// RowIter implements the interface tables.Handler.
func (p PgDbRoleSettingHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_db_role_setting row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgDbRoleSettingHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgDbRoleSettingSchema,
		PkOrdinals: nil,
	}
}

// pgDbRoleSettingSchema is the schema for pg_db_role_setting.
var pgDbRoleSettingSchema = sql.Schema{
	{Name: "setdatabase", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDbRoleSettingName},
	{Name: "setrole", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDbRoleSettingName},
	{Name: "setconfig", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgDbRoleSettingName}, // TODO: collation C
}

// pgDbRoleSettingRowIter is the sql.RowIter for the pg_db_role_setting table.
type pgDbRoleSettingRowIter struct {
}

var _ sql.RowIter = (*pgDbRoleSettingRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgDbRoleSettingRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgDbRoleSettingRowIter) Close(ctx *sql.Context) error {
	return nil
}
