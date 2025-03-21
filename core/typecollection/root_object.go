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

package typecollection

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// DropRootObject implements the interface rootobject.Collection.
func (pgs *TypeCollection) DropRootObject(ctx context.Context, identifier id.Id) error {
	if identifier.Section() != id.Section_Type {
		return errors.Errorf(`type %s does not exist`, identifier.String())
	}
	return pgs.DropType(ctx, id.Type(identifier))
}

// GetID implements the interface rootobject.Collection.
func (pgs *TypeCollection) GetID() rootobject.RootObjectID {
	return rootobject.RootObjectID_Types
}

// GetRootObject implements the interface rootobject.Collection.
func (pgs *TypeCollection) GetRootObject(ctx context.Context, identifier id.Id) (rootobject.RootObject, bool, error) {
	if identifier.Section() != id.Section_Type {
		return nil, false, nil
	}
	typ, err := pgs.GetType(ctx, id.Type(identifier))
	return TypeWrapper{Type: typ}, err == nil, err
}

// HasRootObject implements the interface rootobject.Collection.
func (pgs *TypeCollection) HasRootObject(ctx context.Context, identifier id.Id) (bool, error) {
	if identifier.Section() != id.Section_Type {
		return false, nil
	}
	return pgs.HasType(ctx, id.Type(identifier)), nil
}

// IDToTableName implements the interface rootobject.Collection.
func (pgs *TypeCollection) IDToTableName(identifier id.Id) doltdb.TableName {
	if identifier.Section() != id.Section_Type {
		return doltdb.TableName{}
	}
	typID := id.Type(identifier)
	return doltdb.TableName{
		Name:   typID.TypeName(),
		Schema: typID.SchemaName(),
	}
}

// IterateIDs implements the interface rootobject.Collection.
func (pgs *TypeCollection) IterateIDs(ctx context.Context, callback func(identifier id.Id) (stop bool, err error)) error {
	return pgs.iterateIDs(ctx, func(typID id.Type) (stop bool, err error) {
		return callback(typID.AsId())
	})
}

// IterateRootObjects implements the interface rootobject.Collection.
func (pgs *TypeCollection) IterateRootObjects(ctx context.Context, callback func(rootObj rootobject.RootObject) (stop bool, err error)) error {
	return pgs.IterateTypes(ctx, func(typ *pgtypes.DoltgresType) (stop bool, err error) {
		return callback(TypeWrapper{Type: typ})
	})
}

// PutRootObject implements the interface rootobject.Collection.
func (pgs *TypeCollection) PutRootObject(ctx context.Context, rootObj rootobject.RootObject) error {
	typ, ok := rootObj.(TypeWrapper)
	if !ok || typ.Type == nil {
		return errors.Newf("invalid type root object: %T", rootObj)
	}
	return pgs.CreateType(ctx, typ.Type)
}

// RenameRootObject implements the interface rootobject.Collection.
func (pgs *TypeCollection) RenameRootObject(ctx context.Context, oldName id.Id, newName id.Id) error {
	if !oldName.IsValid() || !newName.IsValid() || oldName.Section() != newName.Section() || oldName.Section() != id.Section_Type {
		return errors.New("cannot rename type due to invalid name")
	}
	oldTypeName := id.Type(oldName)
	newTypeName := id.Type(newName)
	typ, err := pgs.GetType(ctx, oldTypeName)
	if err != nil {
		return err
	}
	if err = pgs.DropType(ctx, oldTypeName); err != nil {
		return err
	}
	newType := *typ
	newType.ID = newTypeName
	return pgs.CreateType(ctx, &newType)
}

// ResolveName implements the interface rootobject.Collection.
func (pgs *TypeCollection) ResolveName(ctx context.Context, name doltdb.TableName) (doltdb.TableName, id.Id, error) {
	rawID, err := pgs.resolveName(ctx, name.Schema, name.Name)
	if err != nil || !rawID.IsValid() {
		return doltdb.TableName{}, id.Null, err
	}
	return doltdb.TableName{
		Name:   rawID.TypeName(),
		Schema: rawID.SchemaName(),
	}, rawID.AsId(), nil
}

// TableNameToID implements the interface rootobject.Collection.
func (pgs *TypeCollection) TableNameToID(name doltdb.TableName) id.Id {
	return id.NewType(name.Schema, name.Name).AsId()
}
