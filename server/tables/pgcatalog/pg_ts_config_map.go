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

// PgTsConfigMapName is a constant to the pg_ts_config_map name.
const PgTsConfigMapName = "pg_ts_config_map"

// InitPgTsConfigMap handles registration of the pg_ts_config_map handler.
func InitPgTsConfigMap() {
	tables.AddHandler(PgCatalogName, PgTsConfigMapName, PgTsConfigMapHandler{})
}

// PgTsConfigMapHandler is the handler for the pg_ts_config_map table.
type PgTsConfigMapHandler struct{}

var _ tables.Handler = PgTsConfigMapHandler{}

// Name implements the interface tables.Handler.
func (p PgTsConfigMapHandler) Name() string {
	return PgTsConfigMapName
}

// RowIter implements the interface tables.Handler.
func (p PgTsConfigMapHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_ts_config_map row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgTsConfigMapHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgTsConfigMapSchema,
		PkOrdinals: nil,
	}
}

// pgTsConfigMapSchema is the schema for pg_ts_config_map.
var pgTsConfigMapSchema = sql.Schema{
	{Name: "mapcfg", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsConfigMapName},
	{Name: "maptokentype", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgTsConfigMapName},
	{Name: "mapseqno", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgTsConfigMapName},
	{Name: "mapdict", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsConfigMapName},
}

// pgTsConfigMapRowIter is the sql.RowIter for the pg_ts_config_map table.
type pgTsConfigMapRowIter struct {
}

var _ sql.RowIter = (*pgTsConfigMapRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTsConfigMapRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgTsConfigMapRowIter) Close(ctx *sql.Context) error {
	return nil
}
