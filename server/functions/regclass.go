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
	"encoding/binary"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initRegclass registers the functions to the catalog.
func initRegclass() {
	framework.RegisterFunction(regclassin)
	framework.RegisterFunction(regclassout)
	framework.RegisterFunction(regclassrecv)
	framework.RegisterFunction(regclasssend)
}

// regclassin represents the PostgreSQL function of regclass type IO input.
var regclassin = framework.Function1{
	Name:       "regclassin",
	Return:     pgtypes.Regclass,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return pgtypes.Regclass_IoInput(ctx, val.(string))
	},
}

// regclassout represents the PostgreSQL function of regclass type IO output.
var regclassout = framework.Function1{
	Name:       "regclassout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regclass},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return pgtypes.Regclass_IoOutput(ctx, val.(uint32))
	},
}

// regclassrecv represents the PostgreSQL function of regclass type IO receive.
var regclassrecv = framework.Function1{
	Name:       "regclassrecv",
	Return:     pgtypes.Regclass,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		return binary.BigEndian.Uint32(data), nil
	},
}

// regclasssend represents the PostgreSQL function of regclass type IO send.
var regclasssend = framework.Function1{
	Name:       "regclasssend",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regclass},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		retVal := make([]byte, 4)
		binary.BigEndian.PutUint32(retVal, val.(uint32))
		return retVal, nil
	},
}
