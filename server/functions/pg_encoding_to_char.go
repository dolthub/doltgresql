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

// initPgEncodingToChar registers the functions to the catalog.
func initPgEncodingToChar() {
	framework.RegisterFunction(pg_encoding_to_char_int)
}

// pg_encoding_to_char_int represents the PostgreSQL system catalog information function.
var pg_encoding_to_char_int = framework.Function1{
	Name:               "pg_encoding_to_char",
	Return:             pgtypes.Name,
	Parameters:         [1]pgtypes.DoltgresType{pgtypes.Int32},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		encoding := val.(int32)
		if encoding == int32(6) {
			return "UTF8", nil
		}
		// TODO: encoding is not supported yet; if invalid val, return empty
		return "", nil
	},
}
