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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgCharToEncoding registers the functions to the catalog.
func initPgCharToEncoding() {
	framework.RegisterFunction(pg_char_to_encoding_name)
}

// pg_char_to_encoding_name represents the PostgreSQL system catalog information function.
var pg_char_to_encoding_name = framework.Function1{
	Name:               "pg_char_to_encoding",
	Return:             pgtypes.Int32,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Name},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		valStr, err := framework.UnwrapString(ctx, val)
		if err != nil {
			return nil, err
		}
		// Encoding names are normalized by lowercasing and dropping any characters that are not ASCII letters or
		// digits, matching Postgres (e.g. "UTF-8" matches "utf8").
		var sb strings.Builder
		for _, r := range strings.ToLower(valStr) {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				sb.WriteRune(r)
			}
		}
		switch sb.String() {
		case "utf8", "unicode":
			return int32(6), nil
		}
		// TODO: only UTF8 is supported for now; Postgres returns -1 for unrecognized encoding names
		return int32(-1), nil
	},
}
