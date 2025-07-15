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

package conflicts

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
)

// DeserializeRootObject implements the interface objinterface.Collection.
func (pgc *Collection) DeserializeRootObject(ctx context.Context, data []byte) (objinterface.RootObject, error) {
	return DeserializeConflict(ctx, data)
}

// DiffRootObjects implements the interface objinterface.Collection.
func (pgc *Collection) DiffRootObjects(ctx context.Context, ours, theirs, ancestor objinterface.RootObject) ([]objinterface.RootObjectDiff, error) {
	// TODO: implement me
	return nil, nil
}

// DropRootObject implements the interface objinterface.Collection.
func (pgc *Collection) DropRootObject(ctx context.Context, identifier id.Id) error {
	return pgc.DropConflict(ctx, identifier)
}

// GetID implements the interface objinterface.Collection.
func (pgc *Collection) GetID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Conflicts
}

// GetRootObject implements the interface objinterface.Collection.
func (pgc *Collection) GetRootObject(ctx context.Context, identifier id.Id) (objinterface.RootObject, bool, error) {
	conflict, err := pgc.GetConflict(ctx, identifier)
	return conflict, err == nil, err
}

// HasRootObject implements the interface objinterface.Collection.
func (pgc *Collection) HasRootObject(ctx context.Context, identifier id.Id) (bool, error) {
	return pgc.HasConflict(ctx, identifier), nil
}

// IDToTableName implements the interface objinterface.Collection.
func (pgc *Collection) IDToTableName(identifier id.Id) doltdb.TableName {
	return doltdb.TableName{Name: string(identifier)}
}

// IterAll implements the interface objinterface.Collection.
func (pgc *Collection) IterAll(ctx context.Context, callback func(rootObj objinterface.RootObject) (stop bool, err error)) error {
	return pgc.IterateConflicts(ctx, func(conflict Conflict) (stop bool, err error) {
		return callback(conflict)
	})
}

// IterIDs implements the interface objinterface.Collection.
func (pgc *Collection) IterIDs(ctx context.Context, callback func(identifier id.Id) (stop bool, err error)) error {
	return pgc.iterateIDs(ctx, func(conflictID id.Id) (stop bool, err error) {
		return callback(conflictID)
	})
}

// PutRootObject implements the interface objinterface.Collection.
func (pgc *Collection) PutRootObject(ctx context.Context, rootObj objinterface.RootObject) error {
	conflict, ok := rootObj.(Conflict)
	if !ok {
		return errors.Newf("invalid conflict root object: %T", rootObj)
	}
	return pgc.AddConflict(ctx, conflict)
}

// RenameRootObject implements the interface objinterface.Collection.
func (pgc *Collection) RenameRootObject(ctx context.Context, oldName id.Id, newName id.Id) error {
	return errors.New("cannot rename root object conflicts")
}

// ResolveName implements the interface objinterface.Collection.
func (pgc *Collection) ResolveName(ctx context.Context, name doltdb.TableName) (doltdb.TableName, id.Id, error) {
	return doltdb.TableName{}, id.Null, nil
}

// TableNameToID implements the interface objinterface.Collection.
func (pgc *Collection) TableNameToID(name doltdb.TableName) id.Id {
	asID := id.Id(name.Name)
	if asID.IsValid() {
		return asID
	}
	return id.Null
}
