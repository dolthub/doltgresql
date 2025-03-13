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

	"github.com/dolthub/doltgresql/core/id"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initNextVal registers the functions to the catalog.
func initNextVal() {
	framework.RegisterFunction(nextval_text)
	framework.RegisterFunction(nextval_regclass)
}

// nextval_text represents the PostgreSQL function of the same name, taking the same parameters.
//
// TODO: Even though we can implicitly convert a text param to a regclass param, it's an expensive process
// to convert it to a regclass, then convert the regclass back into the relation name, so we provide an overload
// that takes a text param directly, in addition to the function form that takes a regclass. Once we can optimize
// the regclass to text conversion, we can potentially remove this overload.
var nextval_text = framework.Function1{
	Name:               "nextval",
	Return:             pgtypes.Int64,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Text},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		schema, sequence, err := parseRelationName(ctx, val.(string))
		if err != nil {
			return nil, err
		}

		collection, err := core.GetSequencesCollectionFromContext(ctx)
		if err != nil {
			return nil, err
		}
		return collection.NextVal(ctx, id.NewSequence(schema, sequence))
	},
}

// nextval_regclass represents the PostgreSQL function of the same name, taking the same parameters.
var nextval_regclass = framework.Function1{
	Name:               "nextval",
	Return:             pgtypes.Int64,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Regclass},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		relationName, err := pgtypes.Regclass.IoOutput(ctx, val)
		if err != nil {
			return nil, err
		}

		schema, sequence, err := parseRelationName(ctx, relationName)
		if err != nil {
			return nil, err
		}

		collection, err := core.GetSequencesCollectionFromContext(ctx)
		if err != nil {
			return nil, err
		}
		return collection.NextVal(ctx, id.NewSequence(schema, sequence))
	},
}
