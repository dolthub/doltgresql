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

// PgTsConfigName is a constant to the pg_ts_config name.
const PgTsConfigName = "pg_ts_config"

// InitPgTsConfig handles registration of the pg_ts_config handler.
func InitPgTsConfig() {
	tables.AddHandler(PgCatalogName, PgTsConfigName, PgTsConfigHandler{})
}

// PgTsConfigHandler is the handler for the pg_ts_config table.
type PgTsConfigHandler struct{}

var _ tables.Handler = PgTsConfigHandler{}

// Name implements the interface tables.Handler.
func (p PgTsConfigHandler) Name() string {
	return PgTsConfigName
}

// RowIter implements the interface tables.Handler.
func (p PgTsConfigHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_ts_config row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgTsConfigHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgTsConfigSchema,
		PkOrdinals: nil,
	}
}

// pgTsConfigSchema is the schema for pg_ts_config.
var pgTsConfigSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsConfigName},
	{Name: "cfgname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgTsConfigName},
	{Name: "cfgnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsConfigName},
	{Name: "cfgowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsConfigName},
	{Name: "cfgparser", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsConfigName},
}

// pgTsConfigRowIter is the sql.RowIter for the pg_ts_config table.
type pgTsConfigRowIter struct {
}

var _ sql.RowIter = (*pgTsConfigRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTsConfigRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgTsConfigRowIter) Close(ctx *sql.Context) error {
	return nil
}
