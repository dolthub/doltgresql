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
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initToRegproc registers the functions to the catalog.
func initToRegproc() {
	framework.RegisterFunction(to_regproc_text)
}

// to_regproc_text represents the PostgreSQL function of the same name, taking the same parameters.
var to_regproc_text = framework.Function1{
	Name:               "to_regproc",
	Return:             pgtypes.Regproc,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Text},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		// If the string just represents a number, then we return nil.
		if _, err := strconv.ParseUint(val1.(string), 10, 32); err == nil {
			return nil, nil
		}
		oid, err := pgtypes.Regproc.IoInput(ctx, val1.(string))
		if err != nil {
			// Specifically for the "does not exist" and "more than one function" errors, we return nil instead of the error.
			// https://www.postgresql.org/docs/15/functions-info.html#FUNCTIONS-INFO-CATALOG-TABLE
			errStr := err.Error()
			if strings.Contains(errStr, "does not exist") || strings.Contains(errStr, "more than one function") {
				return nil, nil
			}
			return nil, err
		}
		return oid, nil
	},
}
