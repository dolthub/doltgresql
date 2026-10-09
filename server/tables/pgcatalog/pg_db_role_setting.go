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
	"github.com/dolthub/doltgresql/server/auth"
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
	var entries []auth.RoleSettingsEntry
	auth.LockRead(func() {
		entries = auth.AllRoleSettings()
	})
	return &pgDbRoleSettingRowIter{entries: entries}, nil
}

// PkSchema implements the interface tables.Handler.
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
	entries []auth.RoleSettingsEntry
	idx     int
}

var _ sql.RowIter = (*pgDbRoleSettingRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgDbRoleSettingRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.entries) {
		return nil, io.EOF
	}
	iter.idx++
	entry := iter.entries[iter.idx-1]
	// An invalid OID applies the settings to every database or role
	databaseOid := id.Null
	if len(entry.Key.Database) > 0 {
		databaseOid = id.NewDatabase(entry.Key.Database).AsId()
	}
	roleOID := id.Null
	if entry.Key.Role.IsValid() {
		roleOID = roleOid(entry.RoleName)
	}
	return sql.Row{
		databaseOid,                      // setdatabase
		roleOID,                          // setrole
		roleSettingsText(entry.Settings), // setconfig
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgDbRoleSettingRowIter) Close(ctx *sql.Context) error {
	return nil
}

// roleConfig returns the settings that apply to the role in every database, for rolconfig and useconfig. Returns nil
// if there are none. This handles locking internally.
func roleConfig(role auth.Role) any {
	var settings []auth.RoleSetting
	auth.LockRead(func() {
		settings = auth.RoleSettingsForKey(auth.RoleSettingKey{Role: role.ID()})
	})
	if len(settings) == 0 {
		return nil
	}
	return roleSettingsText(settings)
}

// roleSettingsText renders the settings as the name=value text array used by the catalog tables.
func roleSettingsText(settings []auth.RoleSetting) []any {
	config := make([]any, len(settings))
	for i, setting := range settings {
		config[i] = setting.Name + "=" + setting.Value
	}
	return config
}
