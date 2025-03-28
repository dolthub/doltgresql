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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
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
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Int64},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		var unusedTypes [4]*pgtypes.DoltgresType
		return setval_text_int64_boolean.Callable(ctx, unusedTypes, val1, val2, true)
	},
}

// setval_text_int64_boolean represents the PostgreSQL function of the same name, taking the same parameters.
var setval_text_int64_boolean = framework.Function3{
	Name:               "setval",
	Return:             pgtypes.Int64,
	Parameters:         [3]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Int64, pgtypes.Bool},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1 any, val2 any, val3 any) (any, error) {
		collection, err := core.GetSequencesCollectionFromContext(ctx)
		if err != nil {
			return nil, err
		}
		// TODO: this should take a regclass as the parameter to determine the schema
		schema, relation, err := parseRelationName(ctx, val1.(string))
		if err != nil {
			return nil, err
		}
		return val2.(int64), collection.SetVal(ctx, id.NewSequence(schema, relation), val2.(int64), val3.(bool))
	},
}

// parseRelationName parses the schema and relation name from a relation name string, including trimming any
// identifier quotes used in the name. For example, passing in 'public."MyTable"' would return 'public' and 'MyTable'.
func parseRelationName(ctx *sql.Context, name string) (schema string, relation string, err error) {
	pathElems := strings.Split(name, ".")
	switch len(pathElems) {
	case 1:
		schema, err = core.GetCurrentSchema(ctx)
		if err != nil {
			return "", "", err
		}
		relation = pathElems[0]
	case 2:
		schema = pathElems[0]
		relation = pathElems[1]
	case 3:
		// database is not used atm
		schema = pathElems[1]
		relation = pathElems[2]
	default:
		return "", "", errors.Errorf(`cannot parse relation: %s`, name)
	}

	// Trim any quotes from the relation name
	relation = strings.Trim(relation, `"`)

	return schema, relation, nil
}
