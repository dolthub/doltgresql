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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/globalstate"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/sequences"
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
		// TODO: this needs a database name to support inserts into other databases (including inserts on other branches than the current one)
		relationName := val1.(string)
		newVal := val2.(int64)
		autoAdvance := val3.(bool)
		collection, err := core.GetSequencesCollectionFromContext(ctx, ctx.GetCurrentDatabase())
		if err != nil {
			return nil, err
		}
		db, err := getDb(ctx)
		if err != nil {
			return nil, err
		}
		root, err := db.GetRoot(ctx)
		if err != nil {
			return nil, err
		}
		var sequenceName doltdb.TableName
		relationBaseName, err := ParseRelationBaseName(ctx, relationName)
		if err != nil {
			return 0, err
		}
		sequenceName, sequence, found, err := resolve.Relation(ctx, root, relationBaseName, sequences.SequenceSource{})
		if err != nil {
			return 0, err
		}
		if !found {
			return 0, errors.Errorf(`sequence "%s" does not exist`, relationName)
		}

		seqId := id.NewSequence(sequenceName.Schema, sequenceName.Name)
		if err != nil {
			return nil, err
		}

		sequenceTracker, err := dsess.GetSequenceTracker(ctx, db.GetGlobalState(), sequences.SequenceTrackerKey)
		if err != nil {
			return nil, err
		}

		ws, err := db.GetWorkingSet(ctx)
		if err != nil {
			return nil, err
		}

		nextState := sequence.SequenceState.WithValue(newVal)
		if autoAdvance {
			_, _, nextState, err = nextState.Next()
			if err != nil {
				return nil, err
			}
		}

		// Set the global state for the sequence.
		// This returns a new Sequence object, but we don't need it.
		_, err = sequenceTracker.Set(ctx, sequenceName, sequence, ws.Ref(), nextState)

		if err != nil {
			return nil, err
		}

		// Set the state on the local version of the sequence too.
		err = collection.SetVal(ctx, seqId, val2.(int64), false, val3.(bool))
		if err != nil {
			return nil, err
		}
		return newVal, nil
	},
}

// ParseRelationNameWithCurrentSchema parses the schema and relation name from a relation name string, including trimming any
// identifier quotes used in the name. If the schema is not specified, the current schema is used.
// For example, passing in 'public."MyTable"' would return 'public' and 'MyTable'.
func ParseRelationNameWithCurrentSchema(ctx *sql.Context, name string) (schema string, relation string, err error) {
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

	// Trim any quotes from the schema and the relation name
	schema = strings.Trim(schema, `"`)
	relation = strings.Trim(relation, `"`)

	return schema, relation, nil
}

func ParseRelationBaseName(ctx *sql.Context, name string) (string, error) {
	pathElems := strings.Split(name, ".")
	var relation string
	switch len(pathElems) {
	case 1:
		relation = pathElems[0]
	case 2:
		relation = pathElems[1]
	case 3:
		// database is not used atm
		relation = pathElems[2]
	default:
		return "", errors.Errorf(`cannot parse relation: %s`, relation)
	}
	return strings.Trim(relation, `"`), nil
}

// ParseRelationName parses the schema and relation name from a relation name string, including trimming any
// identifier quotes used in the name. If the schema is not specified, an empty string is returned.
// For example, passing in 'public."MyTable"' would return 'public' and 'MyTable'.
func ParseRelationName(ctx *sql.Context, name string) (schema string, relation string, err error) {
	pathElems := strings.Split(name, ".")
	switch len(pathElems) {
	case 1:
		schema = ""
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

	// Trim any quotes from the schema and the relation name
	schema = strings.Trim(schema, `"`)
	relation = strings.Trim(relation, `"`)

	return schema, relation, nil
}

func getSequenceTracker(ctx *sql.Context) (*sequences.SequenceTracker, error) {
	sess := dsess.DSessFromSess(ctx.Session)
	db, err := sess.Provider().Database(ctx, sess.GetCurrentDatabase())
	if err != nil {
		return nil, err
	}
	globalStateProvider, ok := db.(globalstate.GlobalStateProvider)
	if !ok {
		return nil, fmt.Errorf("database %s does not implement globalstate.GlobalStateProvider", db.Name())
	}
	return dsess.GetSequenceTracker(ctx, globalStateProvider.GetGlobalState(), sequences.SequenceTrackerKey)
}

func getDb(ctx *sql.Context) (sqle.Database, error) {
	sess := dsess.DSessFromSess(ctx.Session)
	db, err := sess.Provider().Database(ctx, sess.GetCurrentDatabase())
	if err != nil {
		return sqle.Database{}, err
	}
	return db.(sqle.Database), nil
}
