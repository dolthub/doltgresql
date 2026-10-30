// Copyright 2025 Dolthub, Inc.
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
	"github.com/dolthub/dolt/go/libraries/doltcore/ref"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/globalstate"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/sequences"
)

// sequenceIDListener implements the performer and validator functions for sequences.
type sequenceIDListener struct{}

var _ id.Listener = sequenceIDListener{}

// OperationValidator is the internal ID validator for sequences.
func (sequenceIDListener) OperationValidator(ctx *sql.Context, operation id.Operation, databaseName string, originalID id.Id, newID id.Id) error {
	switch originalID.Section() {
	case id.Section_ColumnDefault, id.Section_Table:
		switch operation {
		case id.Operation_Rename, id.Operation_Delete, id.Operation_Delete_Cascade:
			return nil
		default:
			return errors.Errorf("sequence validator received unexpected operation `%s`", operation.String())
		}
	default:
		return errors.Errorf("sequence validator received unexpected section `%s`", originalID.Section().String())
	}
}

// OperationPerformer is the internal ID performer for sequences, which modifies the sequence collection in response to
// the given operation and section of the original ID.
func (sequenceIDListener) OperationPerformer(ctx *sql.Context, operation id.Operation, databaseName string, originalID id.Id, newID id.Id) error {
	switch originalID.Section() {
	case id.Section_ColumnDefault:
		originalIDCol := id.ColumnDefault(originalID)
		switch operation {
		case id.Operation_Rename:
			return nil
		case id.Operation_Delete, id.Operation_Delete_Cascade:
			collection, err := GetSequencesCollectionFromContext(ctx, databaseName)
			if err != nil {
				return err
			}
			sequences, err := collection.GetSequencesWithTable(ctx, doltdb.TableName{
				Name:   originalIDCol.TableName(),
				Schema: originalIDCol.SchemaName(),
			})
			if err != nil {
				return err
			}
			for _, sequence := range sequences {
				if sequence.OwnerColumn == originalIDCol.ColumnName() {
					if err = collection.DropSequence(ctx, sequence.Id); err != nil {
						return err
					}
				}
			}
			return nil
		default:
			return errors.Errorf("sequence performer received unexpected operation `%s`", operation.String())
		}
	case id.Section_Table:
		originalIDTable := id.Table(originalID)
		switch operation {
		case id.Operation_Rename, id.Operation_Delete, id.Operation_Delete_Cascade:
			collection, err := GetSequencesCollectionFromContext(ctx, databaseName)
			if err != nil {
				return err
			}
			sequences, err := collection.GetSequencesWithTable(ctx, doltdb.TableName{
				Name:   originalIDTable.TableName(),
				Schema: originalIDTable.SchemaName(),
			})
			if err != nil {
				return err
			}
			for _, sequence := range sequences {
				if err = collection.DropSequence(ctx, sequence.Id); err != nil {
					return err
				}
				if operation == id.Operation_Rename {
					sequence.OwnerTable = id.Table(newID)
					if err = collection.CreateSequence(ctx, sequence); err != nil {
						return err
					}
				}
			}
			return nil
		default:
			return errors.Errorf("sequence performer received unexpected operation `%s`", operation.String())
		}
	default:
		return errors.Errorf("sequence performer received unexpected section `%s`", originalID.Section().String())
	}
}

// DropSequenceFromTracker removes a dropped sequence from the database's sequence tracker, so that a new sequence with
// the same name does not continue from the dropped sequence's values. The tracker keeps the values of any sequence with
// the same name on other branches.
func DropSequenceFromTracker(ctx *sql.Context, databaseName string, seqID id.Sequence) error {
	sess := dsess.DSessFromSess(ctx.Session)
	db, err := sess.Provider().Database(ctx, databaseName)
	if err != nil {
		return err
	}
	tracker, err := dsess.GetSequenceTracker(ctx, db.(globalstate.GlobalStateProvider).GetGlobalState(), sequences.SequenceTrackerKey)
	if err != nil {
		return err
	}
	ws, err := sess.WorkingSet(ctx, databaseName)
	if err != nil {
		return err
	}
	ddb, ok := sess.GetDoltDB(ctx, databaseName)
	if !ok {
		return errors.Errorf(`database "%s" does not exist`, databaseName)
	}
	branches, err := ddb.GetBranches(ctx)
	if err != nil {
		return err
	}
	var otherWorkingSets []*doltdb.WorkingSet
	for _, branch := range branches {
		wsRef, err := ref.WorkingSetRefForHead(branch)
		if err != nil {
			return err
		}
		if wsRef == ws.Ref() {
			continue
		}
		otherWorkingSet, err := ddb.ResolveWorkingSet(ctx, wsRef)
		if errors.Is(err, doltdb.ErrWorkingSetNotFound) {
			continue
		} else if err != nil {
			return err
		}
		otherWorkingSets = append(otherWorkingSets, otherWorkingSet)
	}
	return tracker.DropRelation(ctx, doltdb.TableName{Name: seqID.SequenceName(), Schema: seqID.SchemaName()}, otherWorkingSets...)
}
