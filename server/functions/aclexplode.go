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

// initAclexplode registers the functions to the catalog.
func initAclexplode() {
	framework.RegisterFunction(aclexplode) // TODO: This breaks pgAdmin 4 because of the unsupported aclitem[] type
}

// aclexplodeName is the name for aclexplode function.
const aclexplodeName = "aclexplode"

// aclexplode represents the PostgreSQL function of the same name, taking the same parameters.
var aclexplode = framework.Function1{
	Name:       aclexplodeName,
	Return:     pgtypes.Record,                              // SETOF record
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.TextArray}, // TODO: type aclitem[]
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return nil, nil
	},
	OutParams: aclexplodeOutArgs,
}

// aclexplodeOutArgs is the schema for aclexplode table function. Each column is OUT argument.
var aclexplodeOutArgs = sql.Schema{
	{Name: "grantor", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: aclexplodeName},
	{Name: "grantee", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: aclexplodeName},
	{Name: "privilege_type", Type: pgtypes.Text, Default: nil, Nullable: false, Source: aclexplodeName},
	{Name: "is_grantable", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: aclexplodeName},
}
