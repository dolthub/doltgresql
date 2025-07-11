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
	"io"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initUnnest registers the functions to the catalog.
func initUnnest() {
	framework.RegisterFunction(unnest)
}

// unnest represents the PostgreSQL function of the same name, taking the same parameters.
var unnest = framework.Function1{
	Name:       "unnest",
	Return:     pgtypes.AnyElement, // TODO: Should return setof AnyElement
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyArray},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		valArr := val1.([]interface{})

		var i = 0
		return pgtypes.NewSetReturningFunctionRowIter(func(ctx *sql.Context) (sql.Row, error) {
			defer func() {
				i++
			}()

			if i >= len(valArr) {
				return nil, io.EOF
			}
			return sql.Row{valArr[i]}, nil
		}), nil
	},
}
