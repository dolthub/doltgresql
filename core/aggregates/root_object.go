// Copyright 2026 Dolthub, Inc.
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

package aggregates

import (
	"context"
	"io"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"

	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// DeserializeRootObject implements the interface objinterface.Collection.
func (pga *Collection) DeserializeRootObject(ctx context.Context, data []byte) (objinterface.RootObject, error) {
	return DeserializeAggregate(ctx, data)
}

// DiffRootObjects implements the interface objinterface.Collection.
func (pga *Collection) DiffRootObjects(ctx context.Context, fromHash string, ours objinterface.RootObject, theirs objinterface.RootObject, ancestor objinterface.RootObject) ([]objinterface.RootObjectDiff, objinterface.RootObject, error) {
	return nil, nil, errors.New("aggregate conflict detection has not yet been implemented")
}

// DropRootObject implements the interface objinterface.Collection.
func (pga *Collection) DropRootObject(ctx context.Context, identifier id.Id) error {
	if identifier.Section() != id.Section_Function {
		return errors.Errorf(`aggregate %s does not exist`, identifier.String())
	}
	return pga.DropAggregate(ctx, id.Function(identifier))
}

// GetFieldType implements the interface objinterface.Collection.
func (pga *Collection) GetFieldType(ctx context.Context, fieldName string) *pgtypes.DoltgresType {
	return nil
}

// GetID implements the interface objinterface.Collection.
func (pga *Collection) GetID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Aggregates
}

// GetRootObject implements the interface objinterface.Collection.
func (pga *Collection) GetRootObject(ctx context.Context, identifier id.Id) (objinterface.RootObject, bool, error) {
	if identifier.Section() != id.Section_Function {
		return nil, false, nil
	}
	a, err := pga.GetAggregate(ctx, id.Function(identifier))
	return a, err == nil && a.ID.IsValid(), err
}

// HasRootObject implements the interface objinterface.Collection.
func (pga *Collection) HasRootObject(ctx context.Context, identifier id.Id) (bool, error) {
	if identifier.Section() != id.Section_Function {
		return false, nil
	}
	return pga.HasAggregate(ctx, id.Function(identifier)), nil
}

// IDToTableName implements the interface objinterface.Collection.
func (pga *Collection) IDToTableName(identifier id.Id) doltdb.TableName {
	if identifier.Section() != id.Section_Function {
		return doltdb.TableName{}
	}
	return functions.FunctionIDToTableName(id.Function(identifier))
}

// IterAll implements the interface objinterface.Collection.
func (pga *Collection) IterAll(ctx context.Context, callback func(rootObj objinterface.RootObject) (stop bool, err error)) error {
	return pga.IterateAggregates(ctx, func(a Aggregate) (stop bool, err error) {
		return callback(a)
	})
}

// IterIDs implements the interface objinterface.Collection.
func (pga *Collection) IterIDs(ctx context.Context, callback func(identifier id.Id) (stop bool, err error)) error {
	return pga.Contents().IterAll(ctx, func(k string, _ hash.Hash) error {
		stop, err := callback(id.Id(k))
		if err != nil {
			return err
		} else if stop {
			return io.EOF
		} else {
			return nil
		}
	})
}

// PutRootObject implements the interface objinterface.Collection.
func (pga *Collection) PutRootObject(ctx context.Context, rootObj objinterface.RootObject) error {
	a, ok := rootObj.(Aggregate)
	if !ok {
		return errors.Newf("invalid aggregate root object: %T", rootObj)
	}
	return pga.AddAggregate(ctx, a)
}

// RenameRootObject implements the interface objinterface.Collection.
func (pga *Collection) RenameRootObject(ctx context.Context, oldName id.Id, newName id.Id) error {
	if !oldName.IsValid() || !newName.IsValid() || oldName.Section() != newName.Section() || oldName.Section() != id.Section_Function {
		return errors.New("cannot rename aggregate due to invalid id")
	}
	a, err := pga.GetAggregate(ctx, id.Function(oldName))
	if err != nil {
		return err
	}
	if err = pga.DropAggregate(ctx, id.Function(oldName)); err != nil {
		return err
	}
	a.ID = id.Function(newName)
	return pga.AddAggregate(ctx, a)
}

// ResolveName implements the interface objinterface.Collection.
func (pga *Collection) ResolveName(ctx context.Context, name doltdb.TableName) (doltdb.TableName, id.Id, error) {
	rawID, err := pga.resolveName(ctx, name.Schema, name.Name)
	if err != nil || !rawID.IsValid() {
		return doltdb.TableName{}, id.Null, err
	}
	return functions.FunctionIDToTableName(rawID), rawID.AsId(), nil
}

// TableNameToID implements the interface objinterface.Collection.
func (pga *Collection) TableNameToID(name doltdb.TableName) id.Id {
	return functions.TableNameToFunctionID(name.Schema, name.Name).AsId()
}

// UpdateField implements the interface objinterface.Collection.
func (pga *Collection) UpdateField(ctx context.Context, rootObject objinterface.RootObject, fieldName string, newValue any) (objinterface.RootObject, error) {
	return nil, errors.New("updating through the conflicts table for this object type is not yet supported")
}
