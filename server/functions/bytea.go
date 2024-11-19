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
	"bytes"
	"encoding/hex"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initBytea registers the functions to the catalog.
func initBytea() {
	framework.RegisterFunction(byteain)
	framework.RegisterFunction(byteaout)
	framework.RegisterFunction(bytearecv)
	framework.RegisterFunction(byteasend)
	framework.RegisterFunction(byteacmp)
}

// byteain represents the PostgreSQL function of bytea type IO input.
var byteain = framework.Function1{
	Name:       "byteain",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		if strings.HasPrefix(input, `\x`) {
			return hex.DecodeString(input[2:])
		} else {
			return []byte(input), nil
		}
	},
}

// byteaout represents the PostgreSQL function of bytea type IO output.
var byteaout = framework.Function1{
	Name:       "byteaout",
	Return:     pgtypes.Cstring,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Bytea},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return `\x` + hex.EncodeToString(val.([]byte)), nil
	},
}

// bytearecv represents the PostgreSQL function of bytea type IO receive.
var bytearecv = framework.Function1{
	Name:       "bytearecv",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		reader := utils.NewReader(data)
		return reader.ByteSlice(), nil
	},
}

// byteasend represents the PostgreSQL function of bytea type IO send.
var byteasend = framework.Function1{
	Name:       "byteasend",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Bytea},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		str := val.([]byte)
		writer := utils.NewWriter(uint64(len(str) + 4))
		writer.ByteSlice(str)
		return writer.Data(), nil
	},
}

// byteacmp represents the PostgreSQL function of bytea type compare.
var byteacmp = framework.Function2{
	Name:       "byteacmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Bytea, pgtypes.Bytea},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		return int32(bytes.Compare(val1.([]byte), val2.([]byte))), nil
	},
}
