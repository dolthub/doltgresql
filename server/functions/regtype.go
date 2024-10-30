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

// initRegtype registers the functions to the catalog.
func initRegtype() {
	framework.RegisterFunction(regtypein)
	framework.RegisterFunction(regtypeout)
	framework.RegisterFunction(regtyperecv)
	framework.RegisterFunction(regtypesend)
}

// regtypein represents the PostgreSQL function of regtype type IO input.
var regtypein = framework.Function1{
	Name:       "regtypein",
	Return:     pgtypes.Regtype,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Text}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return pgtypes.Regtype_IoInput(ctx, val.(string))
	},
}

// regtypeout represents the PostgreSQL function of regtype type IO output.
var regtypeout = framework.Function1{
	Name:       "regtypeout",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Regtype},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return pgtypes.Regtype_IoOutput(ctx, val.(uint32))
	},
}

// regtyperecv represents the PostgreSQL function of regtype type IO receive.
var regtyperecv = framework.Function1{
	Name:       "regtyperecv",
	Return:     pgtypes.Regtype,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		switch val := val.(type) {
		case uint32:
			return val, nil
		default:
			return nil, pgtypes.ErrUnhandledType.New("regtype", val)
		}
	},
}

// regtypesend represents the PostgreSQL function of regtype type IO send.
var regtypesend = framework.Function1{
	Name:       "regtypesend",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Regtype},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		str, err := pgtypes.Regtype_IoOutput(ctx, val.(uint32))
		if err != nil {
			return nil, err
		}
		return []byte(str), nil
	},
}
