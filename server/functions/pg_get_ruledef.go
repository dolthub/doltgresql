// Copyright 2025 Dolthub, Inc.
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

// initPgGetRuledef registers the functions to the catalog.
func initPgGetRuledef() {
	framework.RegisterFunction(pg_get_ruledef_oid)
	framework.RegisterFunction(pg_get_ruledef_oid_bool)
}

// pg_get_ruledef_oid represents the PostgreSQL function of the same name, taking the same parameters.
var pg_get_ruledef_oid = framework.Function1{
	Name:       "pg_get_ruledef",
	Return:     pgtypes.Text,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Oid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		oidVal := val1.(id.Id)
		def, err := getRuleDef(ctx, oidVal, false)
		return def, err
	},
}

// pg_get_ruledef_oid_bool represents the PostgreSQL function of the same name, taking the same parameters.
var pg_get_ruledef_oid_bool = framework.Function2{
	Name:       "pg_get_ruledef",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Bool},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		oidVal := val1.(id.Id)
		pretty := val2.(bool)
		def, err := getRuleDef(ctx, oidVal, pretty)
		return def, err
	},
}

// getRuleDef returns the definition of the rule for the given OID.
// Currently, returns NULL as rule support is not fully implemented.
func getRuleDef(ctx *sql.Context, oidVal id.Id, pretty bool) (any, error) {
	// TODO: Implement rule definition retrieval once rule infrastructure is available
	// For now, return NULL as no rules exist in the system
	return nil, nil
}
