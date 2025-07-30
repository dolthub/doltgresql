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

package sequences

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"

	"github.com/dolthub/doltgresql/core/id"
	pgmerge "github.com/dolthub/doltgresql/core/merge"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

const (
	FIELD_NAME_DATA_TYPE    = "data_type"
	FIELD_NAME_PERSISTENCE  = "persistence"
	FIELD_NAME_START        = "start"
	FIELD_NAME_CURRENT      = "current"
	FIELD_NAME_INCREMENT    = "increment"
	FIELD_NAME_MINIMUM      = "minimum"
	FIELD_NAME_MAXIMUM      = "maximum"
	FIELD_NAME_CACHE        = "cache"
	FIELD_NAME_CYCLE        = "cycle"
	FIELD_NAME_IS_AT_END    = "is_at_end"
	FIELD_NAME_OWNER_TABLE  = "owner_table"
	FIELD_NAME_OWNER_COLUMN = "owner_column"
)

// DeserializeRootObject implements the interface objinterface.Collection.
func (pgs *Collection) DeserializeRootObject(ctx context.Context, data []byte) (objinterface.RootObject, error) {
	return DeserializeSequence(ctx, data)
}

// DiffRootObjects implements the interface objinterface.Collection.
func (pgs *Collection) DiffRootObjects(ctx context.Context, fromHash string, o objinterface.RootObject, t objinterface.RootObject, a objinterface.RootObject) ([]objinterface.RootObjectDiff, objinterface.RootObject, error) {
	ours := o.(*Sequence)
	{
		copiedOurs := *ours
		ours = &copiedOurs
	}
	theirs := t.(*Sequence)
	var ancestor Sequence
	hasAncestor := false
	if ancestorPtr, ok := a.(*Sequence); ok {
		ancestor = *ancestorPtr
		hasAncestor = true
	}
	var diffs []objinterface.RootObjectDiff
	if ours.DataTypeID != theirs.DataTypeID {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Text,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_DATA_TYPE,
		}
		if pgmerge.DiffValues(&diff, ours.DataTypeID.TypeName(), theirs.DataTypeID.TypeName(), ancestor.DataTypeID.TypeName(), hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.DataTypeID = id.NewType(ours.DataTypeID.SchemaName(), diff.OurValue.(string))
		}
	}
	if ours.Persistence != theirs.Persistence {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Int32,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_PERSISTENCE,
		}
		if pgmerge.DiffValues(&diff, int32(ours.Persistence), int32(theirs.Persistence), int32(ancestor.Persistence), hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.Persistence = Persistence(diff.OurValue.(int32))
		}
	}
	if ours.Start != theirs.Start {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Int64,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_START,
		}
		if pgmerge.DiffValues(&diff, ours.Start, theirs.Start, ancestor.Start, hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.Start = diff.OurValue.(int64)
		}
	}
	if ours.Current != theirs.Current {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Int64,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_CURRENT,
		}
		if pgmerge.DiffValues(&diff, ours.Current, theirs.Current, ancestor.Current, hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.Current = diff.OurValue.(int64)
		}
	}
	if ours.Increment != theirs.Increment {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Int64,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_INCREMENT,
		}
		if pgmerge.DiffValues(&diff, ours.Increment, theirs.Increment, ancestor.Increment, hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.Increment = diff.OurValue.(int64)
		}
	}
	if ours.Minimum != theirs.Minimum {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Int64,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_MINIMUM,
		}
		if pgmerge.DiffValues(&diff, ours.Minimum, theirs.Minimum, ancestor.Minimum, hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.Minimum = diff.OurValue.(int64)
		}
	}
	if ours.Maximum != theirs.Maximum {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Int64,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_MAXIMUM,
		}
		if pgmerge.DiffValues(&diff, ours.Maximum, theirs.Maximum, ancestor.Maximum, hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.Maximum = diff.OurValue.(int64)
		}
	}
	if ours.Cache != theirs.Cache {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Int64,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_CACHE,
		}
		if pgmerge.DiffValues(&diff, ours.Cache, theirs.Cache, ancestor.Cache, hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.Cache = diff.OurValue.(int64)
		}
	}
	if ours.Cycle != theirs.Cycle {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Bool,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_CYCLE,
		}
		if pgmerge.DiffValues(&diff, ours.Cycle, theirs.Cycle, ancestor.Cycle, hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.Cycle = diff.OurValue.(bool)
		}
	}
	if ours.IsAtEnd != theirs.IsAtEnd {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Bool,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_IS_AT_END,
		}
		if pgmerge.DiffValues(&diff, ours.IsAtEnd, theirs.IsAtEnd, ancestor.IsAtEnd, hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.IsAtEnd = diff.OurValue.(bool)
		}
	}
	if ours.OwnerTable != theirs.OwnerTable {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Text,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_OWNER_TABLE,
		}
		if pgmerge.DiffValues(&diff, ours.OwnerTable.TableName(), theirs.OwnerTable.TableName(), ancestor.OwnerTable.TableName(), hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.OwnerTable = id.NewTable(ours.OwnerTable.SchemaName(), diff.OurValue.(string))
		}
	}
	if ours.OwnerColumn != theirs.OwnerColumn {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Text,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_OWNER_COLUMN,
		}
		if pgmerge.DiffValues(&diff, ours.OwnerColumn, theirs.OwnerColumn, ancestor.OwnerColumn, hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.OwnerColumn = diff.OurValue.(string)
		}
	}
	return diffs, ours, nil
}

// DropRootObject implements the interface objinterface.Collection.
func (pgs *Collection) DropRootObject(ctx context.Context, identifier id.Id) error {
	if identifier.Section() != id.Section_Sequence {
		return errors.Errorf(`sequence %s does not exist`, identifier.String())
	}
	return pgs.DropSequence(ctx, id.Sequence(identifier))
}

// GetFieldType implements the interface objinterface.Collection.
func (pgs *Collection) GetFieldType(ctx context.Context, fieldName string) *pgtypes.DoltgresType {
	switch fieldName {
	case FIELD_NAME_DATA_TYPE:
		return pgtypes.Text
	case FIELD_NAME_PERSISTENCE:
		return pgtypes.Int32
	case FIELD_NAME_START:
		return pgtypes.Int64
	case FIELD_NAME_CURRENT:
		return pgtypes.Int64
	case FIELD_NAME_INCREMENT:
		return pgtypes.Int64
	case FIELD_NAME_MINIMUM:
		return pgtypes.Int64
	case FIELD_NAME_MAXIMUM:
		return pgtypes.Int64
	case FIELD_NAME_CACHE:
		return pgtypes.Int64
	case FIELD_NAME_CYCLE:
		return pgtypes.Bool
	case FIELD_NAME_IS_AT_END:
		return pgtypes.Bool
	case FIELD_NAME_OWNER_TABLE:
		return pgtypes.Text
	case FIELD_NAME_OWNER_COLUMN:
		return pgtypes.Text
	default:
		return nil
	}
}

// GetID implements the interface objinterface.Collection.
func (pgs *Collection) GetID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Sequences
}

// GetRootObject implements the interface objinterface.Collection.
func (pgs *Collection) GetRootObject(ctx context.Context, identifier id.Id) (objinterface.RootObject, bool, error) {
	if identifier.Section() != id.Section_Sequence {
		return nil, false, nil
	}
	seq, err := pgs.GetSequence(ctx, id.Sequence(identifier))
	return seq, err == nil && seq != nil, err
}

// HasRootObject implements the interface objinterface.Collection.
func (pgs *Collection) HasRootObject(ctx context.Context, identifier id.Id) (bool, error) {
	if identifier.Section() != id.Section_Sequence {
		return false, nil
	}
	return pgs.HasSequence(ctx, id.Sequence(identifier)), nil
}

// IDToTableName implements the interface objinterface.Collection.
func (pgs *Collection) IDToTableName(identifier id.Id) doltdb.TableName {
	if identifier.Section() != id.Section_Sequence {
		return doltdb.TableName{}
	}
	seqID := id.Sequence(identifier)
	return doltdb.TableName{
		Name:   seqID.SequenceName(),
		Schema: seqID.SchemaName(),
	}
}

// IterAll implements the interface objinterface.Collection.
func (pgs *Collection) IterAll(ctx context.Context, callback func(rootObj objinterface.RootObject) (stop bool, err error)) error {
	return pgs.IterateSequences(ctx, func(seq *Sequence) (stop bool, err error) {
		return callback(seq)
	})
}

// IterIDs implements the interface objinterface.Collection.
func (pgs *Collection) IterIDs(ctx context.Context, callback func(identifier id.Id) (stop bool, err error)) error {
	return pgs.iterateIDs(ctx, func(seqID id.Sequence) (stop bool, err error) {
		return callback(seqID.AsId())
	})
}

// PutRootObject implements the interface objinterface.Collection.
func (pgs *Collection) PutRootObject(ctx context.Context, rootObj objinterface.RootObject) error {
	seq, ok := rootObj.(*Sequence)
	if !ok {
		return errors.Newf("invalid sequence root object: %T", rootObj)
	}
	return pgs.CreateSequence(ctx, seq)
}

// RenameRootObject implements the interface objinterface.Collection.
func (pgs *Collection) RenameRootObject(ctx context.Context, oldName id.Id, newName id.Id) error {
	if !oldName.IsValid() || !newName.IsValid() || oldName.Section() != newName.Section() || oldName.Section() != id.Section_Sequence {
		return errors.New("cannot rename sequence due to invalid name")
	}
	oldSeqName := id.Sequence(oldName)
	newSeqName := id.Sequence(newName)
	seq, err := pgs.GetSequence(ctx, oldSeqName)
	if err != nil {
		return err
	}
	if err = pgs.DropSequence(ctx, oldSeqName); err != nil {
		return err
	}
	newSeq := *seq
	newSeq.Id = newSeqName
	return pgs.CreateSequence(ctx, &newSeq)
}

// ResolveName implements the interface objinterface.Collection.
func (pgs *Collection) ResolveName(ctx context.Context, name doltdb.TableName) (doltdb.TableName, id.Id, error) {
	rawID, err := pgs.resolveName(ctx, name.Schema, name.Name)
	if err != nil || !rawID.IsValid() {
		return doltdb.TableName{}, id.Null, err
	}
	return doltdb.TableName{
		Name:   rawID.SequenceName(),
		Schema: rawID.SchemaName(),
	}, rawID.AsId(), nil
}

// TableNameToID implements the interface objinterface.Collection.
func (pgs *Collection) TableNameToID(name doltdb.TableName) id.Id {
	return id.NewSequence(name.Schema, name.Name).AsId()
}

// UpdateField implements the interface objinterface.Collection.
func (pgs *Collection) UpdateField(ctx context.Context, rootObject objinterface.RootObject, fieldName string, newValue any) (objinterface.RootObject, error) {
	return nil, errors.New("updating through the conflicts table for this object type is not yet supported")
}
