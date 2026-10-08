// Copyright 2026 Dolthub, Inc.
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

package functions

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgTsConfigIsVisible registers the functions to the catalog.
func initPgTsConfigIsVisible() {
	framework.RegisterFunction(pg_ts_config_is_visible_oid)
}

// pg_ts_config_is_visible_oid represents the PostgreSQL system schema visibility inquiry function.
var pg_ts_config_is_visible_oid = framework.Function1{
	Name:               "pg_ts_config_is_visible",
	Return:             pgtypes.Bool,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		oidVal := val.(id.Id)
		// Only the built-in text search configurations exist for now, and they all live in pg_catalog, which is
		// always part of the search path.
		// TODO: check the containing schema against the search path once user-defined configurations are supported
		return oidVal.Section() == id.Section_TextSearchConfig && id.Cache().Exists(oidVal), nil
	},
}
