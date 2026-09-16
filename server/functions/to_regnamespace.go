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
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initToRegnamespace registers the functions to the catalog.
func initToRegnamespace() {
	framework.RegisterFunction(to_regnamespace_text)
}

// to_regnamespace_text represents the PostgreSQL function of the same name, taking the same parameters.
var to_regnamespace_text = framework.Function1{
	Name:               "to_regnamespace",
	Return:             pgtypes.Regnamespace,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Text},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		val1Str, err := framework.UnwrapString(ctx, val1)
		if err != nil {
			return nil, err
		}
		if _, err := strconv.ParseUint(val1Str, 10, 32); err == nil {
			return nil, nil
		}
		oid, err := pgtypes.Regnamespace.IoInput(ctx, val1Str)
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				return nil, nil
			}
			return nil, err
		}
		return oid, nil
	},
}
