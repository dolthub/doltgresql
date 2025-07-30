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

package triggers

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// DeserializeRootObject implements the interface objinterface.Collection.
func (pgt *Collection) DeserializeRootObject(ctx context.Context, data []byte) (objinterface.RootObject, error) {
	return DeserializeTrigger(ctx, data)
}

// DiffRootObjects implements the interface objinterface.Collection.
func (pgt *Collection) DiffRootObjects(ctx context.Context, fromHash string, ours objinterface.RootObject, theirs objinterface.RootObject, ancestor objinterface.RootObject) ([]objinterface.RootObjectDiff, objinterface.RootObject, error) {
	return nil, nil, errors.New("trigger conflict detection has not yet been implemented")
}

// DropRootObject implements the interface objinterface.Collection.
func (pgt *Collection) DropRootObject(ctx context.Context, identifier id.Id) error {
	if identifier.Section() != id.Section_Trigger {
		return errors.Errorf(`trigger %s does not exist`, identifier.String())
	}
	return pgt.DropTrigger(ctx, id.Trigger(identifier))
}

// GetFieldType implements the interface objinterface.Collection.
func (pgt *Collection) GetFieldType(ctx context.Context, fieldName string) *pgtypes.DoltgresType {
	return nil
}

// GetID implements the interface objinterface.Collection.
func (pgt *Collection) GetID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Triggers
}

// GetRootObject implements the interface objinterface.Collection.
func (pgt *Collection) GetRootObject(ctx context.Context, identifier id.Id) (objinterface.RootObject, bool, error) {
	if identifier.Section() != id.Section_Trigger {
		return nil, false, nil
	}
	f, err := pgt.GetTrigger(ctx, id.Trigger(identifier))
	return f, err == nil && f.ID.IsValid(), err
}

// HasRootObject implements the interface objinterface.Collection.
func (pgt *Collection) HasRootObject(ctx context.Context, identifier id.Id) (bool, error) {
	if identifier.Section() != id.Section_Trigger {
		return false, nil
	}
	return pgt.HasTrigger(ctx, id.Trigger(identifier)), nil
}

// IDToTableName implements the interface objinterface.Collection.
func (pgt *Collection) IDToTableName(identifier id.Id) doltdb.TableName {
	if identifier.Section() != id.Section_Trigger {
		return doltdb.TableName{}
	}
	return TriggerIDToTableName(id.Trigger(identifier))
}

// IterAll implements the interface objinterface.Collection.
func (pgt *Collection) IterAll(ctx context.Context, callback func(rootObj objinterface.RootObject) (stop bool, err error)) error {
	return pgt.IterateTriggers(ctx, func(t Trigger) (stop bool, err error) {
		return callback(t)
	})
}

// IterIDs implements the interface objinterface.Collection.
func (pgt *Collection) IterIDs(ctx context.Context, callback func(identifier id.Id) (stop bool, err error)) error {
	return pgt.iterateIDs(ctx, func(trigID id.Trigger) (stop bool, err error) {
		return callback(trigID.AsId())
	})
}

// PutRootObject implements the interface objinterface.Collection.
func (pgt *Collection) PutRootObject(ctx context.Context, rootObj objinterface.RootObject) error {
	t, ok := rootObj.(Trigger)
	if !ok {
		return errors.Newf("invalid trigger root object: %T", rootObj)
	}
	return pgt.AddTrigger(ctx, t)
}

// RenameRootObject implements the interface objinterface.Collection.
func (pgt *Collection) RenameRootObject(ctx context.Context, oldName id.Id, newName id.Id) error {
	if !oldName.IsValid() || !newName.IsValid() || oldName.Section() != newName.Section() || oldName.Section() != id.Section_Trigger {
		return errors.New("cannot rename trigger due to invalid name")
	}
	oldTriggerName := id.Trigger(oldName)
	newTriggerName := id.Trigger(newName)
	t, err := pgt.GetTrigger(ctx, oldTriggerName)
	if err != nil {
		return err
	}
	if err = pgt.DropTrigger(ctx, newTriggerName); err != nil {
		return err
	}
	t.ID = newTriggerName
	return pgt.AddTrigger(ctx, t)
}

// ResolveName implements the interface objinterface.Collection.
func (pgt *Collection) ResolveName(ctx context.Context, name doltdb.TableName) (doltdb.TableName, id.Id, error) {
	rawID, err := pgt.resolveName(ctx, name.Schema, name.Name)
	if err != nil || !rawID.IsValid() {
		return doltdb.TableName{}, id.Null, err
	}
	return TriggerIDToTableName(rawID), rawID.AsId(), nil
}

// TableNameToID implements the interface objinterface.Collection.
func (pgt *Collection) TableNameToID(name doltdb.TableName) id.Id {
	return pgt.tableNameToID(name.Schema, name.Name).AsId()
}

// UpdateField implements the interface objinterface.Collection.
func (pgt *Collection) UpdateField(ctx context.Context, rootObject objinterface.RootObject, fieldName string, newValue any) (objinterface.RootObject, error) {
	return nil, errors.New("updating through the conflicts table for this object type is not yet supported")
}
