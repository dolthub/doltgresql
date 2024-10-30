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

// initToRegclass registers the functions to the catalog.
func initToRegclass() {
	framework.RegisterFunction(to_regclass_text)
}

// to_regclass_text represents the PostgreSQL function of the same name, taking the same parameters.
var to_regclass_text = framework.Function1{
	Name:               "to_regclass",
	Return:             pgtypes.Regclass,
	Parameters:         [1]pgtypes.DoltgresType{pgtypes.Text},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val1 any) (any, error) {
		// If the string just represents a number, then we return nil.
		if _, err := strconv.ParseUint(val1.(string), 10, 32); err == nil {
			return nil, nil
		}
		oid, err := framework.IoInput(ctx, pgtypes.Regclass, val1.(string))
		if err != nil {
			// Specifically for the "does not exist" error, we return nil instead of the error.
			// https://www.postgresql.org/docs/15/functions-info.html#FUNCTIONS-INFO-CATALOG-TABLE
			if strings.Contains(err.Error(), "does not exist") {
				return nil, nil
			}
			return nil, err
		}
		return oid, nil
	},
}
