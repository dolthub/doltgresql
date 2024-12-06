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

	"github.com/dolthub/doltgresql/utils"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initVoid registers the functions to the catalog.
func initVoid() {
	framework.RegisterFunction(void_in)
	framework.RegisterFunction(void_out)
	framework.RegisterFunction(void_recv)
	framework.RegisterFunction(void_send)
}

// void_in represents the PostgreSQL function of void type IO input.
var void_in = framework.Function2{
	Name:       "void_in",
	Return:     pgtypes.Void,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Cstring, pgtypes.Oid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		return val1.(string), nil
	},
}

// void_out represents the PostgreSQL function of void type IO output.
var void_out = framework.Function1{
	Name:       "void_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Void},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return val.(string), nil
	},
}

// void_recv represents the PostgreSQL function of void type IO receive.
var void_recv = framework.Function2{
	Name:       "void_recv",
	Return:     pgtypes.Void,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		data := val1.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		reader := utils.NewReader(data)
		return reader.String(), nil
	},
}

// void_send represents the PostgreSQL function of void type IO send.
var void_send = framework.Function1{
	Name:       "void_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Void},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		str := val.(string)
		writer := utils.NewWriter(uint64(len(str) + 4))
		writer.String(str)
		return writer.Data(), nil
	},
}
