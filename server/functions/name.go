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

// initName registers the functions to the catalog.
func initName() {
	framework.RegisterFunction(namein)
	framework.RegisterFunction(nameout)
	framework.RegisterFunction(namerecv)
	framework.RegisterFunction(namesend)
	framework.RegisterFunction(btnamecmp)
	framework.RegisterFunction(btnametextcmp)
}

// namein represents the PostgreSQL function of name type IO input.
var namein = framework.Function1{
	Name:       "namein",
	Return:     pgtypes.Name,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		input, _ = truncateString(input, pgtypes.NameLength)
		return input, nil
	},
}

// nameout represents the PostgreSQL function of name type IO output.
var nameout = framework.Function1{
	Name:       "nameout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Name},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		str, _ := truncateString(val.(string), pgtypes.NameLength)
		return str, nil
	},
}

// namerecv represents the PostgreSQL function of name type IO receive.
var namerecv = framework.Function1{
	Name:       "namerecv",
	Return:     pgtypes.Name,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		reader := utils.NewReader(data)
		return reader.String(), nil
	},
}

// namesend represents the PostgreSQL function of name type IO send.
var namesend = framework.Function1{
	Name:       "namesend",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Name},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		str := val.(string)
		writer := utils.NewWriter(uint64(len(str) + 1))
		writer.String(str)
		return writer.Data(), nil
	},
}

// btnamecmp represents the PostgreSQL function of name type compare.
var btnamecmp = framework.Function2{
	Name:       "btnamecmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Name},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(string)
		bb := val2.(string)
		if ab == bb {
			return int32(0), nil
		} else if ab < bb {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}

// btnametextcmp represents the PostgreSQL function of name type compare with text.
var btnametextcmp = framework.Function2{
	Name:       "btnametextcmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(string)
		bb := val2.(string)
		if ab == bb {
			return int32(0), nil
		} else if ab < bb {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}
