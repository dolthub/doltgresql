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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initBool registers the functions to the catalog.
func initBool() {
	framework.RegisterFunction(boolin)
	framework.RegisterFunction(boolout)
	framework.RegisterFunction(boolrecv)
	framework.RegisterFunction(boolsend)
	framework.RegisterFunction(btboolcmp)
}

// boolin represents the PostgreSQL function of boolean type IO input.
var boolin = framework.Function1{
	Name:       "boolin",
	Return:     pgtypes.Bool,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		val = strings.TrimSpace(strings.ToLower(val.(string)))
		if val == "true" || val == "t" || val == "yes" || val == "on" || val == "1" {
			return true, nil
		} else if val == "false" || val == "f" || val == "no" || val == "off" || val == "0" {
			return false, nil
		} else {
			return nil, pgtypes.ErrInvalidSyntaxForType.New("boolean", val)
		}
	},
}

// boolout represents the PostgreSQL function of boolean type IO output.
var boolout = framework.Function1{
	Name:       "boolout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Bool},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		if val.(bool) {
			return "true", nil
		} else {
			return "false", nil
		}
	},
}

// boolrecv represents the PostgreSQL function of boolean type IO receive.
var boolrecv = framework.Function1{
	Name:       "boolrecv",
	Return:     pgtypes.Bool,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		return data[0] != 0, nil
	},
}

// boolsend represents the PostgreSQL function of boolean type IO send.
var boolsend = framework.Function1{
	Name:       "boolsend",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Bool},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		if val.(bool) {
			return []byte{1}, nil
		} else {
			return []byte{0}, nil
		}
	},
}

// btboolcmp represents the PostgreSQL function of boolean type byte compare.
var btboolcmp = framework.Function2{
	Name:       "btboolcmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Bool, pgtypes.Bool},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(bool)
		bb := val2.(bool)
		if ab == bb {
			return int32(0), nil
		} else if !ab {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}
