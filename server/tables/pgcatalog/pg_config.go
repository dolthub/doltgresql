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
	// TODO: Implement pg_config row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
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

// pgConfigRowIter is the sql.RowIter for the pg_config table.
type pgConfigRowIter struct {
}

var _ sql.RowIter = (*pgConfigRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgConfigRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgConfigRowIter) Close(ctx *sql.Context) error {
	return nil
}
