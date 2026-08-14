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

package operators

import (
	"context"
	"io"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// DeserializeRootObject implements the interface objinterface.Collection.
func (pgo *Collection) DeserializeRootObject(ctx context.Context, data []byte) (objinterface.RootObject, error) {
	return DeserializeOperator(ctx, data)
}

// DiffRootObjects implements the interface objinterface.Collection.
func (pgo *Collection) DiffRootObjects(ctx context.Context, fromHash string, ours objinterface.RootObject, theirs objinterface.RootObject, ancestor objinterface.RootObject) ([]objinterface.RootObjectDiff, objinterface.RootObject, error) {
	return nil, nil, errors.New("operator conflict detection has not yet been implemented")
}

// DropRootObject implements the interface objinterface.Collection.
func (pgo *Collection) DropRootObject(ctx context.Context, identifier id.Id) error {
	if identifier.Section() != id.Section_Operator {
		return errors.Errorf(`operator %s does not exist`, identifier.String())
	}
	return pgo.DropOperator(ctx, id.Operator(identifier))
}

// GetFieldType implements the interface objinterface.Collection.
func (pgo *Collection) GetFieldType(ctx context.Context, fieldName string) *pgtypes.DoltgresType {
	return nil
}

// GetID implements the interface objinterface.Collection.
func (pgo *Collection) GetID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Operators
}

// GetRootObject implements the interface objinterface.Collection.
func (pgo *Collection) GetRootObject(ctx context.Context, identifier id.Id) (objinterface.RootObject, bool, error) {
	if identifier.Section() != id.Section_Operator {
		return nil, false, nil
	}
	o, err := pgo.GetOperator(ctx, id.Operator(identifier))
	return o, err == nil && o.ID.IsValid(), err
}

// HasRootObject implements the interface objinterface.Collection.
func (pgo *Collection) HasRootObject(ctx context.Context, identifier id.Id) (bool, error) {
	if identifier.Section() != id.Section_Operator {
		return false, nil
	}
	return pgo.HasOperator(ctx, id.Operator(identifier)), nil
}

// IDToTableName implements the interface objinterface.Collection.
func (pgo *Collection) IDToTableName(identifier id.Id) doltdb.TableName {
	if identifier.Section() != id.Section_Operator {
		return doltdb.TableName{}
	}
	return OperatorIDToTableName(id.Operator(identifier))
}

// IterAll implements the interface objinterface.Collection.
func (pgo *Collection) IterAll(ctx context.Context, callback func(rootObj objinterface.RootObject) (stop bool, err error)) error {
	return pgo.IterateOperators(ctx, func(o Operator) (stop bool, err error) {
		return callback(o)
	})
}

// IterIDs implements the interface objinterface.Collection.
func (pgo *Collection) IterIDs(ctx context.Context, callback func(identifier id.Id) (stop bool, err error)) error {
	return pgo.Contents().IterAll(ctx, func(k string, _ hash.Hash) error {
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
func (pgo *Collection) PutRootObject(ctx context.Context, rootObj objinterface.RootObject) error {
	o, ok := rootObj.(Operator)
	if !ok {
		return errors.Newf("invalid operator root object: %T", rootObj)
	}
	return pgo.AddOperator(ctx, o)
}

// RenameRootObject implements the interface objinterface.Collection.
func (pgo *Collection) RenameRootObject(ctx context.Context, oldName id.Id, newName id.Id) error {
	if !oldName.IsValid() || !newName.IsValid() || oldName.Section() != newName.Section() || oldName.Section() != id.Section_Operator {
		return errors.New("cannot rename operator due to invalid id")
	}
	o, err := pgo.GetOperator(ctx, id.Operator(oldName))
	if err != nil {
		return err
	}
	if err = pgo.DropOperator(ctx, id.Operator(oldName)); err != nil {
		return err
	}
	o.ID = id.Operator(newName)
	return pgo.AddOperator(ctx, o)
}

// ResolveName implements the interface objinterface.Collection.
func (pgo *Collection) ResolveName(ctx context.Context, name doltdb.TableName) (doltdb.TableName, id.Id, error) {
	rawID, err := pgo.resolveName(ctx, name.Schema, name.Name)
	if err != nil || !rawID.IsValid() {
		return doltdb.TableName{}, id.Null, err
	}
	return OperatorIDToTableName(rawID), rawID.AsId(), nil
}

// TableNameToID implements the interface objinterface.Collection.
func (pgo *Collection) TableNameToID(name doltdb.TableName) id.Id {
	return pgo.tableNameToID(name.Schema, name.Name).AsId()
}

// UpdateField implements the interface objinterface.Collection.
func (pgo *Collection) UpdateField(ctx context.Context, rootObject objinterface.RootObject, fieldName string, newValue any) (objinterface.RootObject, error) {
	return nil, errors.New("updating through the conflicts table for this object type is not yet supported")
}
