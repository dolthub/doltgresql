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

// initNextVal registers the functions to the catalog.
func initNextVal() {
	framework.RegisterFunction(nextval_text)
}

// nextval_text represents the PostgreSQL function of the same name, taking the same parameters.
var nextval_text = framework.Function1{
	Name:       "nextval",
	Return:     pgtypes.Int64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Text},
	Callable: func(ctx *sql.Context, val1 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		var outVal int64
		err := core.UseRootInSession(ctx, func(ctx *sql.Context, root *core.RootValue) (*core.RootValue, error) {
			collection, err := root.GetSequences(ctx)
			if err != nil {
				return nil, err
			}
			outVal, err = collection.NextVal(core.GetCurrentSchema(ctx), val1.(string))
			if err != nil {
				return nil, err
			}
			return root.PutSequences(ctx, collection)
		})
		return outVal, err
	},
}