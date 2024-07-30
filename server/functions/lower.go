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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initLower registers the functions to the catalog.
func initLower() {
	framework.RegisterFunction(lower_varchar)
}

// lower_varchar represents the PostgreSQL function of the same name, taking the same parameters.
var lower_varchar = framework.Function1{
	Name:       "lower",
	Return:     pgtypes.VarChar,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.VarChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val1 any, varargs ...any) (any, error) {
		// TODO: this doesn't respect collations
		return strings.ToLower(val1.(string)), nil
	},
}
