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

// initUnknown registers the functions to the catalog.
func initUnknown() {
	framework.RegisterFunction(unknownin)
	framework.RegisterFunction(unknownout)
	framework.RegisterFunction(unknownrecv)
	framework.RegisterFunction(unknownsend)
}

// unknownin represents the PostgreSQL function of unknown type IO input.
var unknownin = framework.Function1{
	Name:       "unknownin",
	Return:     pgtypes.Unknown,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Text}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return val.(string), nil
	},
}

// unknownout represents the PostgreSQL function of unknown type IO output.
var unknownout = framework.Function1{
	Name:       "unknownout",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Unknown},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return val.(string), nil
	},
}

// unknownrecv represents the PostgreSQL function of unknown type IO receive.
var unknownrecv = framework.Function1{
	Name:       "unknownrecv",
	Return:     pgtypes.Unknown,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		reader := utils.NewReader(data)
		return reader.String(), nil
	},
}

// unknownsend represents the PostgreSQL function of unknown type IO send.
var unknownsend = framework.Function1{
	Name:       "unknownsend",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Unknown},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		str := val.(string)
		writer := utils.NewWriter(uint64(len(str) + 4))
		writer.String(str)
		return writer.Data(), nil
	},
}
