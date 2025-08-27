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

// PgStatUserFunctionsName is a constant to the pg_stat_user_functions name.
const PgStatUserFunctionsName = "pg_stat_user_functions"

// InitPgStatUserFunctions handles registration of the pg_stat_user_functions handler.
func InitPgStatUserFunctions() {
	tables.AddHandler(PgCatalogName, PgStatUserFunctionsName, PgStatUserFunctionsHandler{})
}

// PgStatUserFunctionsHandler is the handler for the pg_stat_user_functions table.
type PgStatUserFunctionsHandler struct{}

var _ tables.Handler = PgStatUserFunctionsHandler{}

// Name implements the interface tables.Handler.
func (p PgStatUserFunctionsHandler) Name() string {
	return PgStatUserFunctionsName
}

// RowIter implements the interface tables.Handler.
func (p PgStatUserFunctionsHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_user_functions row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatUserFunctionsHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatUserFunctionsSchema,
		PkOrdinals: nil,
	}
}

// pgStatUserFunctionsSchema is the schema for pg_stat_user_functions.
var pgStatUserFunctionsSchema = sql.Schema{
	{Name: "funcid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatUserFunctionsName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatUserFunctionsName},
	{Name: "funcname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatUserFunctionsName},
	{Name: "calls", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserFunctionsName},
	{Name: "total_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatUserFunctionsName},
	{Name: "self_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatUserFunctionsName},
}

// pgStatUserFunctionsRowIter is the sql.RowIter for the pg_stat_user_functions table.
type pgStatUserFunctionsRowIter struct {
}

var _ sql.RowIter = (*pgStatUserFunctionsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatUserFunctionsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatUserFunctionsRowIter) Close(ctx *sql.Context) error {
	return nil
}
