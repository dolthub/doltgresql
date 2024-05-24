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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initSetVal registers the functions to the catalog.
func initSetVal() {
	framework.RegisterFunction(setval_text_int64)
	framework.RegisterFunction(setval_text_int64_boolean)
}

// setval_text_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var setval_text_int64 = framework.Function2{
	Name:               "setval",
	Return:             pgtypes.Int64,
	Parameters:         []pgtypes.DoltgresType{pgtypes.Text, pgtypes.Int64},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		return setval_text_int64_boolean.Callable(ctx, val1, val2, true)
	},
}

// setval_text_int64_boolean represents the PostgreSQL function of the same name, taking the same parameters.
var setval_text_int64_boolean = framework.Function3{
	Name:               "setval",
	Return:             pgtypes.Int64,
	Parameters:         []pgtypes.DoltgresType{pgtypes.Text, pgtypes.Int64, pgtypes.Bool},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, val1 any, val2 any, val3 any) (any, error) {
		if val1 == nil || val2 == nil || val3 == nil {
			return nil, nil
		}
		collection, err := core.GetCollectionFromContext(ctx)
		if err != nil {
			return nil, err
		}
		// TODO: this should take a regclass as the parameter to determine the schema
		schema, err := core.GetCurrentSchema(ctx)
		if err != nil {
			return nil, err
		}
		
		return val2.(int64), collection.SetVal(schema, val1.(string), val2.(int64), val3.(bool))
	},
}
