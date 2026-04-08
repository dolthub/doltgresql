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

// initConcat registers the functions to the catalog.
func initConcat() {
	framework.RegisterFunction(concat_any)
}

// concat_any represents the PostgreSQL function of the same name, taking the same parameters.
var concat_any = framework.Function1N{
	Name:       "concat",
	Return:     pgtypes.Text,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Any},
	Strict:     false,
	Callable: func(ctx *sql.Context, t []*pgtypes.DoltgresType, val1 any, vals []any) (any, error) {
		sb := strings.Builder{}
		if val1 != nil {
			output, err := t[0].IoOutput(ctx, val1)
			if err != nil {
				return nil, err
			}
			sb.WriteString(output)
		}
		for i, val := range vals {
			if val == nil {
				continue
			}
			valType := t[i+1]
			if valType.ID == pgtypes.Bool.ID {
				// Within this context, `bool` returns 't' rather than 'true'
				if val.(bool) {
					sb.WriteRune('t')
				} else {
					sb.WriteRune('f')
				}
			} else {
				output, err := valType.IoOutput(ctx, val)
				if err != nil {
					return nil, err
				}
				sb.WriteString(output)
			}
		}
		return sb.String(), nil
	},
}
