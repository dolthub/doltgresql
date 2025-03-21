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
	"github.com/dolthub/doltgresql/core/rootobject"
)

// DropRootObject implements the interface rootobject.Collection.
func (pgs *Collection) DropRootObject(ctx context.Context, identifier id.Id) error {
	if identifier.Section() != id.Section_Sequence {
		return errors.Errorf(`sequence %s does not exist`, identifier.String())
	}
	return pgs.DropSequence(ctx, id.Sequence(identifier))
}

// GetID implements the interface rootobject.Collection.
func (pgs *Collection) GetID() rootobject.RootObjectID {
	return rootobject.RootObjectID_Sequences
}

// GetRootObject implements the interface rootobject.Collection.
func (pgs *Collection) GetRootObject(ctx context.Context, identifier id.Id) (rootobject.RootObject, bool, error) {
	if identifier.Section() != id.Section_Sequence {
		return nil, false, nil
	}
	seq, err := pgs.GetSequence(ctx, id.Sequence(identifier))
	return seq, err == nil, err
}

// HasRootObject implements the interface rootobject.Collection.
func (pgs *Collection) HasRootObject(ctx context.Context, identifier id.Id) (bool, error) {
	if identifier.Section() != id.Section_Sequence {
		return false, nil
	}
	return pgs.HasSequence(ctx, id.Sequence(identifier)), nil
}

// IDToTableName implements the interface rootobject.Collection.
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

// IterateIDs implements the interface rootobject.Collection.
func (pgs *Collection) IterateIDs(ctx context.Context, callback func(identifier id.Id) (stop bool, err error)) error {
	return pgs.iterateIDs(ctx, func(seqID id.Sequence) (stop bool, err error) {
		return callback(seqID.AsId())
	})
}

// IterateRootObjects implements the interface rootobject.Collection.
func (pgs *Collection) IterateRootObjects(ctx context.Context, callback func(rootObj rootobject.RootObject) (stop bool, err error)) error {
	return pgs.IterateSequences(ctx, func(seq *Sequence) (stop bool, err error) {
		return callback(seq)
	})
}

// PutRootObject implements the interface rootobject.Collection.
func (pgs *Collection) PutRootObject(ctx context.Context, rootObj rootobject.RootObject) error {
	seq, ok := rootObj.(*Sequence)
	if !ok {
		return errors.Newf("invalid sequence root object: %T", rootObj)
	}
	return pgs.CreateSequence(ctx, seq)
}

// RenameRootObject implements the interface rootobject.Collection.
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

// ResolveName implements the interface rootobject.Collection.
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

// TableNameToID implements the interface rootobject.Collection.
func (pgs *Collection) TableNameToID(name doltdb.TableName) id.Id {
	return id.NewSequence(name.Schema, name.Name).AsId()
}
