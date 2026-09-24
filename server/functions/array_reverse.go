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
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
)

func initArrayReverse() { framework.RegisterFunction(array_reverse) }

var array_reverse = framework.Function1{
	Name: "array_reverse", Return: pgtypes.AnyArray, Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyArray}, Strict: true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		vals := val.([]any)
		result := make([]any, len(vals))
		for i, v := range vals {
			result[len(vals)-1-i] = v
		}
		return result, nil
	},
}
