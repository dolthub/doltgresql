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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initOidvector registers the functions to the catalog.
func initOidvector() {
	framework.RegisterFunction(oidvectorin)
	framework.RegisterFunction(oidvectorout)
	framework.RegisterFunction(oidvectorrecv)
	framework.RegisterFunction(oidvectorsend)
	framework.RegisterFunction(btoidvectorcmp)
}

// oidvectorin represents the PostgreSQL function of oidvector type IO input.
var oidvectorin = framework.Function1{
	Name:       "oidvectorin",
	Return:     pgtypes.Oidvector,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		strValues := strings.Split(input, " ")
		var values = make([]any, len(strValues))
		for i, strValue := range strValues {
			innerValue, err := pgtypes.Oid.IoInput(ctx, strValue)
			if err != nil {
				return nil, err
			}
			values[i] = innerValue
		}
		return values, nil
	},
}

// oidvectorout represents the PostgreSQL function of oidvector type IO output.
var oidvectorout = framework.Function1{
	Name:       "oidvectorout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Oidvector},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return pgtypes.VectorToString(ctx, val.([]any), pgtypes.Oid)
	},
}

// oidvectorrecv represents the PostgreSQL function of oidvector type IO receive.
var oidvectorrecv = framework.Function1{
	Name:       "oidvectorrecv",
	Return:     pgtypes.Oidvector,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		return deserializeArray(ctx, data, pgtypes.Oid)
	},
}

// oidvectorsend represents the PostgreSQL function of oidvector type IO send.
var oidvectorsend = framework.Function1{
	Name:       "oidvectorsend",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Oidvector},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		return array_send.Callable(ctx, t, val)
	},
}

// btoidvectorcmp represents the PostgreSQL function of oidvector type IO input.
var btoidvectorcmp = framework.Function2{
	Name:       "btoidvectorcmp",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Oidvector, pgtypes.Oidvector},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		leftOidvector := val1.([]any)
		rightOidvector := val2.([]any)
		llen := len(leftOidvector)
		rlen := len(rightOidvector)
		minLen := min(llen, rlen)

		for i := 0; i < minLen; i++ {
			lOid := id.Cache().ToOID(leftOidvector[i].(id.Id))
			rOid := id.Cache().ToOID(rightOidvector[i].(id.Id))

			if lOid < rOid {
				return int32(-1), nil
			}
			if lOid > rOid {
				return int32(1), nil
			}
		}

		// Compare lengths if all elements matched
		if llen < rlen {
			return int32(-1), nil
		} else if llen > rlen {
			return int32(1), nil
		}

		return int32(0), nil
	},
}
