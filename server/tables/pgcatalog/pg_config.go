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

	"github.com/dolthub/doltgresql/server/config"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgConfigName is a constant to the pg_config name.
const PgConfigName = "pg_config"

// InitPgConfig handles registration of the pg_config handler.
func InitPgConfig() {
	tables.AddHandler(PgCatalogName, PgConfigName, PgConfigHandler{})
}

// PgConfigHandler is the handler for the pg_config table.
type PgConfigHandler struct{}

var _ tables.Handler = PgConfigHandler{}

// Name implements the interface tables.Handler.
func (p PgConfigHandler) Name() string {
	return PgConfigName
}

// RowIter implements the interface tables.Handler.
func (p PgConfigHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	return &pgConfigRowIter{
		configs: defaultPgConfigs,
		idx:     0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgConfigHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgConfigSchema,
		PkOrdinals: nil,
	}
}

// pgConfigSchema is the schema for pg_config.
var pgConfigSchema = sql.Schema{
	{Name: "name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgConfigName},
	{Name: "setting", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgConfigName},
}

// pgConfigVersion returns the value for the VERSION row of pg_config, matching the Postgres version
// that Doltgres reports through the server_version configuration parameter.
func pgConfigVersion() string {
	if param, ok := config.PostgresConfigParameters()["server_version"].(*config.Parameter); ok {
		if version, ok := param.Default.(string); ok {
			return "PostgreSQL " + version
		}
	}
	return "PostgreSQL"
}

// pgConfig represents a row in the pg_config table.
type pgConfig struct {
	name    string
	setting string
}

// defaultPgConfigs is the fixed set of rows reported by pg_config, matching the 23 rows that
// Postgres reports.
// TODO: the settings are placeholders, since Doltgres is not built from the Postgres source tree.
var defaultPgConfigs = []pgConfig{
	{name: "BINDIR", setting: "/usr/local/pgsql/bin"},
	{name: "DOCDIR", setting: "/usr/local/pgsql/share/doc"},
	{name: "HTMLDIR", setting: "/usr/local/pgsql/share/doc"},
	{name: "INCLUDEDIR", setting: "/usr/local/pgsql/include"},
	{name: "PKGINCLUDEDIR", setting: "/usr/local/pgsql/include"},
	{name: "INCLUDEDIR-SERVER", setting: "/usr/local/pgsql/include/server"},
	{name: "LIBDIR", setting: "/usr/local/pgsql/lib"},
	{name: "PKGLIBDIR", setting: "/usr/local/pgsql/lib"},
	{name: "LOCALEDIR", setting: "/usr/local/pgsql/share/locale"},
	{name: "MANDIR", setting: "/usr/local/pgsql/share/man"},
	{name: "SHAREDIR", setting: "/usr/local/pgsql/share"},
	{name: "SYSCONFDIR", setting: "/usr/local/pgsql/etc"},
	{name: "PGXS", setting: "/usr/local/pgsql/lib/pgxs/src/makefiles/pgxs.mk"},
	{name: "CONFIGURE", setting: ""},
	{name: "CC", setting: "cc"},
	{name: "CPPFLAGS", setting: ""},
	{name: "CFLAGS", setting: ""},
	{name: "CFLAGS_SL", setting: ""},
	{name: "LDFLAGS", setting: ""},
	{name: "LDFLAGS_EX", setting: ""},
	{name: "LDFLAGS_SL", setting: ""},
	{name: "LIBS", setting: ""},
	{name: "VERSION", setting: pgConfigVersion()},
}

// pgConfigRowIter is the sql.RowIter for the pg_config table.
type pgConfigRowIter struct {
	configs []pgConfig
	idx     int
}

var _ sql.RowIter = (*pgConfigRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgConfigRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.configs) {
		return nil, io.EOF
	}
	iter.idx++
	cfg := iter.configs[iter.idx-1]

	return sql.Row{
		cfg.name,    // name
		cfg.setting, // setting
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgConfigRowIter) Close(ctx *sql.Context) error {
	return nil
}
