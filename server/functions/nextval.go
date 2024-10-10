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
	"fmt"
	"strings"

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
	Name:               "nextval",
	Return:             pgtypes.Int64,
	Parameters:         [1]pgtypes.DoltgresType{pgtypes.Text},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		var schema, sequence string
		var err error
		pathElems := strings.Split(val.(string), ".")
		switch len(pathElems) {
		case 1:
			schema, err = core.GetCurrentSchema(ctx)
			if err != nil {
				return nil, err
			}
			sequence = pathElems[0]
		case 2:
			schema = pathElems[0]
			sequence = pathElems[1]
		case 3:
			// database is not used atm
			schema = pathElems[1]
			sequence = pathElems[2]
		default:
			return nil, fmt.Errorf(`cannot find sequence "%s" to get its nextval`, val.(string))
		}
		collection, err := core.GetSequencesCollectionFromContext(ctx)
		if err != nil {
			return nil, err
		}
		return collection.NextVal(schema, sequence)
	},
}
