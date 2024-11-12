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

// initDomain registers the functions to the catalog.
func initDomain() {
	framework.RegisterFunction(domain_in)
	framework.RegisterFunction(domain_recv)
}

// domain_in represents the PostgreSQL function of domain type IO input.
var domain_in = framework.Function3{
	Name:       "domain_in",
	Return:     pgtypes.Any,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Cstring, pgtypes.Oid, pgtypes.Int32},
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		str := val1.(string)
		baseTypeOid := val2.(uint32)
		t := pgtypes.OidToBuildInDoltgresType[baseTypeOid]
		typmod := val3.(int32)
		t.AttTypMod = typmod
		return framework.IoInput(ctx, t, str)
	},
}

// domain_recv represents the PostgreSQL function of domain type IO receive.
var domain_recv = framework.Function3{
	Name:       "domain_recv",
	Return:     pgtypes.Any,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		data := val1.([]byte)
		baseTypeOid := val2.(uint32)
		t := pgtypes.OidToBuildInDoltgresType[baseTypeOid]
		typmod := val3.(int32)
		t.AttTypMod = typmod
		return framework.IoReceive(ctx, t, data)
	},
}
