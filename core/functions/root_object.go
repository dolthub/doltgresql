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

package functions

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
)

// DropRootObject implements the interface objinterface.Collection.
func (pgf *Collection) DropRootObject(ctx context.Context, identifier id.Id) error {
	if pgf.isReadOnly {
		return errors.New("cannot modify a read-only collection")
	}
	if identifier.Section() != id.Section_Function {
		return errors.Errorf(`function %s does not exist`, identifier.String())
	}
	return pgf.DropFunction(ctx, id.Function(identifier))
}

// GetID implements the interface objinterface.Collection.
func (pgf *Collection) GetID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Functions
}

// GetRootObject implements the interface objinterface.Collection.
func (pgf *Collection) GetRootObject(ctx context.Context, identifier id.Id) (objinterface.RootObject, bool, error) {
	if identifier.Section() != id.Section_Function {
		return nil, false, nil
	}
	f, err := pgf.GetFunction(ctx, id.Function(identifier))
	return f, err == nil, err
}

// HasRootObject implements the interface objinterface.Collection.
func (pgf *Collection) HasRootObject(ctx context.Context, identifier id.Id) (bool, error) {
	if identifier.Section() != id.Section_Function {
		return false, nil
	}
	return pgf.HasFunction(ctx, id.Function(identifier)), nil
}

// IDToTableName implements the interface objinterface.Collection.
func (pgf *Collection) IDToTableName(identifier id.Id) doltdb.TableName {
	if identifier.Section() != id.Section_Function {
		return doltdb.TableName{}
	}
	return FunctionIDToTableName(id.Function(identifier))
}

// IsReadOnly implements the interface objinterface.Collection.
func (pgf *Collection) IsReadOnly() bool {
	return pgf.isReadOnly
}

// IterAll implements the interface objinterface.Collection.
func (pgf *Collection) IterAll(ctx context.Context, callback func(rootObj objinterface.RootObject) (stop bool, err error)) error {
	return pgf.IterateFunctions(ctx, func(f Function) (stop bool, err error) {
		return callback(f)
	})
}

// IterIDs implements the interface objinterface.Collection.
func (pgf *Collection) IterIDs(ctx context.Context, callback func(identifier id.Id) (stop bool, err error)) error {
	return pgf.iterateIDs(ctx, func(funcID id.Function) (stop bool, err error) {
		return callback(funcID.AsId())
	})
}

// PutRootObject implements the interface objinterface.Collection.
func (pgf *Collection) PutRootObject(ctx context.Context, rootObj objinterface.RootObject) error {
	if pgf.isReadOnly {
		return errors.New("cannot modify a read-only collection")
	}
	f, ok := rootObj.(Function)
	if !ok {
		return errors.Newf("invalid function root object: %T", rootObj)
	}
	return pgf.AddFunction(ctx, f)
}

// RenameRootObject implements the interface objinterface.Collection.
func (pgf *Collection) RenameRootObject(ctx context.Context, oldName id.Id, newName id.Id) error {
	if pgf.isReadOnly {
		return errors.New("cannot modify a read-only collection")
	}
	if !oldName.IsValid() || !newName.IsValid() || oldName.Section() != newName.Section() || oldName.Section() != id.Section_Function {
		return errors.New("cannot rename function due to invalid name")
	}
	oldFuncName := id.Function(oldName)
	newFuncName := id.Function(newName)
	if oldFuncName.ParameterCount() != newFuncName.ParameterCount() {
		return errors.Newf(`old function id had "%d" parameters, new function id has "%d" parameters`,
			oldFuncName.ParameterCount(), newFuncName.ParameterCount())
	}
	f, err := pgf.GetFunction(ctx, oldFuncName)
	if err != nil {
		return err
	}
	if err = pgf.DropFunction(ctx, oldFuncName); err != nil {
		return err
	}
	f.ID = newFuncName
	return pgf.AddFunction(ctx, f)
}

// ResolveName implements the interface objinterface.Collection.
func (pgf *Collection) ResolveName(ctx context.Context, name doltdb.TableName) (doltdb.TableName, id.Id, error) {
	rawID, err := pgf.resolveName(ctx, name.Schema, name.Name)
	if err != nil || !rawID.IsValid() {
		return doltdb.TableName{}, id.Null, err
	}
	return FunctionIDToTableName(rawID), rawID.AsId(), nil
}

// SetReadOnly implements the interface objinterface.Collection.
func (pgf *Collection) SetReadOnly() {
	pgf.isReadOnly = true
}

// TableNameToID implements the interface objinterface.Collection.
func (pgf *Collection) TableNameToID(name doltdb.TableName) id.Id {
	return pgf.tableNameToID(name.Schema, name.Name).AsId()
}
