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
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initOid registers the functions to the catalog.
func initOid() {
	framework.RegisterFunction(oidin)
	framework.RegisterFunction(oidout)
	framework.RegisterFunction(oidrecv)
	framework.RegisterFunction(oidsend)
	framework.RegisterFunction(btoidcmp)
}

// oidin represents the PostgreSQL function of oid type IO input.
var oidin = framework.Function1{
	Name:       "oidin",
	Return:     pgtypes.Oid,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Text}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		uVal, err := strconv.ParseInt(strings.TrimSpace(input), 10, 64)
		if err != nil {
			return nil, pgtypes.ErrInvalidSyntaxForType.New("oid", input)
		}
		// Note: This minimum is different (-4294967295) for Postgres 15.4 compiled by Visual C++
		if uVal > pgtypes.MaxUint32 || uVal < pgtypes.MinInt32 {
			return nil, pgtypes.ErrValueIsOutOfRangeForType.New(input, "oid")
		}
		return uint32(uVal), nil
	},
}

// oidout represents the PostgreSQL function of oid type IO output.
var oidout = framework.Function1{
	Name:       "oidout",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Oid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return strconv.FormatUint(uint64(val.(uint32)), 10), nil
	},
}

// oidrecv represents the PostgreSQL function of oid type IO receive.
var oidrecv = framework.Function1{
	Name:       "oidrecv",
	Return:     pgtypes.Oid,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		switch val := val.(type) {
		case uint32:
			return val, nil
		default:
			return nil, pgtypes.ErrUnhandledType.New("oid", val)
		}
	},
}

// oidsend represents the PostgreSQL function of oid type IO send.
var oidsend = framework.Function1{
	Name:       "oidsend",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Oid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return []byte(strconv.FormatUint(uint64(val.(uint32)), 10)), nil
	},
}

// btoidcmp represents the PostgreSQL function of oid type compare.
var btoidcmp = framework.Function2{
	Name:       "btoidcmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Oid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(uint32)
		bb := val2.(uint32)
		if ab == bb {
			return int32(0), nil
		} else if ab < bb {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}
