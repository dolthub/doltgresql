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

package core

import (
	"github.com/cockroachdb/errors"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
)

// RelationType states the type of the relation.
type RelationType byte

const (
	RelationType_DoesNotExist RelationType = iota
	RelationType_Table
	RelationType_Sequence
)

// GetRelationType returns whether the working root has the given relation, and what type of relation it is. According
// to the Postgres docs, a relation may be one of: table, sequence, index, view, materialized view, foreign table. This
// may also include composite types and partitions, but this hasn't been confirmed.
func GetRelationType(ctx *sql.Context, schema string, relation string) (RelationType, error) {
	// TODO: the schema isn't actually being used
	if len(schema) == 0 {
		var err error
		schema, err = GetCurrentSchema(ctx)
		if err != nil {
			return RelationType_DoesNotExist, err
		}
	}

	session := dsess.DSessFromSess(ctx.Session)
	state, ok, err := session.LookupDbState(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return RelationType_DoesNotExist, err
	}
	if !ok {
		return RelationType_DoesNotExist, errors.Errorf("GetRelationType cannot find the database")
	}
	return GetRelationTypeFromRoot(ctx, schema, relation, state.WorkingRoot().(*RootValue))
}

// GetRelationTypeFromRoot performs the same function as GetRelationType, except that it uses the given root rather than
// the working session's root.
func GetRelationTypeFromRoot(ctx *sql.Context, schema string, relation string, root *RootValue) (RelationType, error) {
	// Check tables first
	ok, err := root.HasTable(ctx, doltdb.TableName{Schema: schema, Name: relation})
	if err != nil {
		return RelationType_DoesNotExist, err
	}
	if ok {
		return RelationType_Table, nil
	}
	// Check sequences next
	collection, err := root.GetSequences(ctx)
	if err != nil {
		return RelationType_DoesNotExist, err
	}
	if collection.HasSequence(ctx, id.NewSequence(schema, relation)) {
		return RelationType_Sequence, nil
	}
	// TODO: the rest of the relations
	return RelationType_DoesNotExist, nil
}
