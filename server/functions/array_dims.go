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
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initArrayDims registers the functions to the catalog.
func initArrayDims() {
	framework.RegisterFunction(array_dims_anyarray)
}

// array_dims_anyarray represents the PostgreSQL function of the same name, taking the same parameters.
var array_dims_anyarray = framework.Function1{
	Name:       "array_dims",
	Return:     pgtypes.Text,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		dims := pgtypes.ArrayDims(val1.([]any), t[0].ArrayBaseType())
		if len(dims) == 0 {
			return nil, nil
		}
		sb := strings.Builder{}
		for _, dim := range dims {
			sb.WriteString(fmt.Sprintf("[1:%d]", dim))
		}
		return sb.String(), nil
	},
}
