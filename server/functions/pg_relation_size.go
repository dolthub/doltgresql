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

package functions

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgRelationSize registers the functions to the catalog.
func initPgRelationSize() {
	framework.RegisterFunction(pg_relation_size_regclass)
	framework.RegisterFunction(pg_relation_size_regclass_text)
}

// pg_relation_size_regclass represents the PostgreSQL function of the same name, taking the same parameters.
var pg_relation_size_regclass = framework.Function1{
	Name:               "pg_relation_size",
	Return:             pgtypes.Int64,
	Parameters:         [1]pgtypes.DoltgresType{pgtypes.Regclass},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO: on-disk size in bytes of one fork of that relation
		//  used by 'main' by default.
		return int64(0), nil
	},
}

// pg_relation_size_regclass_text represents the PostgreSQL function of the same name, taking the same parameters.
var pg_relation_size_regclass_text = framework.Function2{
	Name:               "pg_relation_size",
	Return:             pgtypes.Int64,
	Parameters:         [2]pgtypes.DoltgresType{pgtypes.Regclass, pgtypes.Text},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		// TODO: on-disk size in bytes of one fork of that relation
		//  used by the specified fork ('main', 'fsm', 'vm', or 'init')
		return int64(0), nil
	},
}
