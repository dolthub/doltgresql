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
	"fmt"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initLength registers the functions to the catalog.
func initLength() {
	framework.RegisterFunction(length_text)
}

// length_text represents the PostgreSQL function of the same name, taking the same parameters.
var length_text = framework.Function1{
	Name:       "length",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		val1str, ok, err := sql.Unwrap[string](ctx, val1)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("unexpected type for length input, expected string, got %T", val1)
		}
		return int32(len([]rune(val1str))), nil
	},
}
