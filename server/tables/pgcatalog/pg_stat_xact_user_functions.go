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

// PgStatXactUserFunctionsName is a constant to the pg_stat_xact_user_functions name.
const PgStatXactUserFunctionsName = "pg_stat_xact_user_functions"

// InitPgStatXactUserFunctions handles registration of the pg_stat_xact_user_functions handler.
func InitPgStatXactUserFunctions() {
	tables.AddHandler(PgCatalogName, PgStatXactUserFunctionsName, PgStatXactUserFunctionsHandler{})
}

// PgStatXactUserFunctionsHandler is the handler for the pg_stat_xact_user_functions table.
type PgStatXactUserFunctionsHandler struct{}

var _ tables.Handler = PgStatXactUserFunctionsHandler{}

// Name implements the interface tables.Handler.
func (p PgStatXactUserFunctionsHandler) Name() string {
	return PgStatXactUserFunctionsName
}

// RowIter implements the interface tables.Handler.
func (p PgStatXactUserFunctionsHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_stat_xact_user_functions row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatXactUserFunctionsHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatXactUserFunctionsSchema,
		PkOrdinals: nil,
	}
}

// pgStatXactUserFunctionsSchema is the schema for pg_stat_xact_user_functions.
var pgStatXactUserFunctionsSchema = sql.Schema{
	{Name: "funcid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatXactUserFunctionsName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatXactUserFunctionsName},
	{Name: "funcname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatXactUserFunctionsName},
	{Name: "calls", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactUserFunctionsName},
	{Name: "total_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatXactUserFunctionsName},
	{Name: "self_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatXactUserFunctionsName},
}

// pgStatXactUserFunctionsRowIter is the sql.RowIter for the pg_stat_xact_user_functions table.
type pgStatXactUserFunctionsRowIter struct {
}

var _ sql.RowIter = (*pgStatXactUserFunctionsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatXactUserFunctionsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatXactUserFunctionsRowIter) Close(ctx *sql.Context) error {
	return nil
}
