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

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/types/oid"
)

// initNextVal registers the functions to the catalog.
func initNextVal() {
	framework.RegisterFunction(nextval_regclass)
}

// nextval_text represents the PostgreSQL function of the same name, taking the same parameters.
var nextval_regclass = framework.Function1{
	Name:               "nextval",
	Return:             pgtypes.Int64,
	Parameters:         [1]pgtypes.DoltgresType{pgtypes.Regclass},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		var schemaName, seqName string
		err := oid.RunCallback(ctx, val.(uint32), oid.Callbacks{
			Sequence: func(ctx *sql.Context, schema oid.ItemSchema, sequence oid.ItemSequence) (cont bool, err error) {
				schemaName = schema.Item.SchemaName()
				seqName = sequence.Item.Name
				return false, nil
			},
		})
		if err != nil {
			return nil, err
		}

		collection, err := core.GetCollectionFromContext(ctx)
		if err != nil {
			return nil, err
		}
		if schemaName == "" || seqName == "" {
			return nil, fmt.Errorf("cannot find sequence to get its nextval")
		}
		return collection.NextVal(schemaName, seqName)
	},
}
