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

// initObjDescription registers the functions to the catalog.
func initObjDescription() {
	framework.RegisterFunction(obj_description_oid)
	framework.RegisterFunction(obj_description_oid_name)
}

// obj_description_oid represents the PostgreSQL function of the same name, taking the same parameters.
var obj_description_oid = framework.Function1{
	Name:               "obj_description",
	Return:             pgtypes.Text,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// TODO: When we support comments this should return the comment for a
		// database object specified by its OID and the name of the containing
		// system catalog.
		return "", nil
	},
}

// obj_description_oid_name represents the PostgreSQL function of the same name, taking the same parameters.
var obj_description_oid_name = framework.Function2{
	Name:               "obj_description",
	Return:             pgtypes.Text,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Name},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		// TODO: When we support comments this should return the comment for a
		// database object specified by its OID and the name of the containing
		// system catalog.
		return "", nil
	},
}
