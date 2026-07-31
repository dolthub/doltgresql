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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/doltgresql/core/sequences"
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

func nextval(ctx *sql.Context, ait *sequences.SequenceTracker, relationName string) (int64, error) {
	// TODO: this needs a database name to support inserts into other databases (including inserts on other branches than the current one)
	collection, err := core.GetSequencesCollectionFromContext(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return 0, err
	}

	schema, sequenceString, err := ParseRelationName(ctx, relationName)
	if err != nil {
		return 0, err
	}

	// TODO: Include Schema Name
	// As a partial workaround, ensure that the sequence exists
	sequenceId := id.NewSequence(schema, sequenceString)
	sequenceName := doltdb.TableName{Name: sequenceString, Schema: schema}
	sequence, err := collection.GetSequence(ctx, sequenceId)
	if err != nil {
		return 0, err
	}
	if sequence == nil {
		return 0, errors.Errorf(`relation "%s" does not exist`, sequenceString)
	}

	next, err := ait.Next(ctx, sequenceName, nil)
	if err != nil {
		return 0, err
	}

	err = collection.SetVal(ctx, sequenceId, next, true, true)
	if err != nil {
		return 0, err
	}
	return next, err
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
		ait, err := getSequenceTracker(ctx)
		if err != nil {
			return 0, err
		}
		return nextval(ctx, ait, val.(string))
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
		ait, err := getSequenceTracker(ctx)
		if err != nil {
			return 0, err
		}
		return nextval(ctx, ait, relationName)
	},
}
