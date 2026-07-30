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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initHasDatabasePrivilege registers the functions to the catalog.
func initHasDatabasePrivilege() {
	framework.RegisterFunction(has_database_privilege_name_text_text)
	framework.RegisterFunction(has_database_privilege_name_oid_text)
	framework.RegisterFunction(has_database_privilege_oid_name_text)
	framework.RegisterFunction(has_database_privilege_oid_oid_text)
	framework.RegisterFunction(has_database_privilege_text_text)
	framework.RegisterFunction(has_database_privilege_oid_text)
}

// has_database_privilege_name_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_database_privilege_name_text_text = framework.Function3{
	Name:       "has_database_privilege",
	Return:     pgtypes.Bool,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		// TODO does user have privilege for database
		return true, nil
	},
}

// has_database_privilege_name_oid_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_database_privilege_name_oid_text = framework.Function3{
	Name:       "has_database_privilege",
	Return:     pgtypes.Bool,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Oid, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		// TODO does user have privilege for database
		return true, nil
	},
}

// has_database_privilege_oid_name_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_database_privilege_oid_name_text = framework.Function3{
	Name:       "has_database_privilege",
	Return:     pgtypes.Bool,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Name, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		// TODO does user have privilege for database
		return true, nil
	},
}

// has_database_privilege_oid_oid_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_database_privilege_oid_oid_text = framework.Function3{
	Name:       "has_database_privilege",
	Return:     pgtypes.Bool,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Oid, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		// TODO does user have privilege for database
		return true, nil
	},
}

// has_database_privilege_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_database_privilege_text_text = framework.Function2{
	Name:       "has_database_privilege",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		// TODO does current user have privilege for database
		return true, nil
	},
}

// has_database_privilege_oid_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_database_privilege_oid_text = framework.Function2{
	Name:       "has_database_privilege",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		// TODO does current user have privilege for database
		return true, nil
	},
}
