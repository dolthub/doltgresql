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

package casts

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
func (pgc *Collection) DeserializeRootObject(ctx context.Context, data []byte) (objinterface.RootObject, error) {
	return DeserializeCast(ctx, data)
}

// DiffRootObjects implements the interface objinterface.Collection.
func (pgc *Collection) DiffRootObjects(ctx context.Context, fromHash string, ours objinterface.RootObject, theirs objinterface.RootObject, ancestor objinterface.RootObject) ([]objinterface.RootObjectDiff, objinterface.RootObject, error) {
	return nil, nil, errors.New("cast conflict detection has not yet been implemented")
}

// DropRootObject implements the interface objinterface.Collection.
func (pgc *Collection) DropRootObject(ctx context.Context, identifier id.Id) error {
	if identifier.Section() != id.Section_Cast {
		return errors.Errorf(`cast %s does not exist`, identifier.String())
	}
	return pgc.DropCast(ctx, id.Cast(identifier))
}

// GetFieldType implements the interface objinterface.Collection.
func (pgc *Collection) GetFieldType(ctx context.Context, fieldName string) *pgtypes.DoltgresType {
	return nil
}

// GetID implements the interface objinterface.Collection.
func (pgc *Collection) GetID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Casts
}

// GetRootObject implements the interface objinterface.Collection.
func (pgc *Collection) GetRootObject(ctx context.Context, identifier id.Id) (objinterface.RootObject, bool, error) {
	if identifier.Section() != id.Section_Cast {
		return nil, false, nil
	}
	c, err := pgc.getCast(ctx, id.Cast(identifier), nil, nil, CastType_Explicit)
	return c, err == nil && c.ID.IsValid(), err
}

// HasRootObject implements the interface objinterface.Collection.
func (pgc *Collection) HasRootObject(ctx context.Context, identifier id.Id) (bool, error) {
	if identifier.Section() != id.Section_Cast {
		return false, nil
	}
	return pgc.HasCast(ctx, id.Cast(identifier)), nil
}

// IDToTableName implements the interface objinterface.Collection.
func (pgc *Collection) IDToTableName(identifier id.Id) doltdb.TableName {
	if identifier.Section() != id.Section_Cast {
		return doltdb.TableName{}
	}
	return CastIDToTableName(id.Cast(identifier))
}

// IterAll implements the interface objinterface.Collection.
func (pgc *Collection) IterAll(ctx context.Context, callback func(rootObj objinterface.RootObject) (stop bool, err error)) error {
	return pgc.IterateCasts(ctx, func(c Cast) (stop bool, err error) {
		return callback(c)
	})
}

// IterIDs implements the interface objinterface.Collection.
func (pgc *Collection) IterIDs(ctx context.Context, callback func(identifier id.Id) (stop bool, err error)) error {
	err := pgc.underlyingMap.IterAll(ctx, func(k string, _ hash.Hash) error {
		stop, err := callback(id.Id(k))
		if err != nil {
			return err
		} else if stop {
			return io.EOF
		} else {
			return nil
		}
	})
	return err
}

// PutRootObject implements the interface objinterface.Collection.
func (pgc *Collection) PutRootObject(ctx context.Context, rootObj objinterface.RootObject) error {
	c, ok := rootObj.(Cast)
	if !ok {
		return errors.Newf("invalid cast root object: %T", rootObj)
	}
	return pgc.AddCast(ctx, c)
}

// RenameRootObject implements the interface objinterface.Collection.
func (pgc *Collection) RenameRootObject(ctx context.Context, oldName id.Id, newName id.Id) error {
	if !oldName.IsValid() || !newName.IsValid() || oldName.Section() != newName.Section() || oldName.Section() != id.Section_Cast {
		return errors.New("cannot rename cast due to invalid id")
	}
	oldCastName := id.Cast(oldName)
	newCastName := id.Cast(newName)
	c, err := pgc.getCast(ctx, oldCastName, nil, nil, CastType_Explicit)
	if err != nil {
		return err
	}
	if err = pgc.DropCast(ctx, newCastName); err != nil {
		return err
	}
	c.ID = newCastName
	return pgc.AddCast(ctx, c)
}

// ResolveName implements the interface objinterface.Collection.
func (pgc *Collection) ResolveName(ctx context.Context, name doltdb.TableName) (doltdb.TableName, id.Id, error) {
	rawID, err := pgc.resolveName(ctx, name.Schema, name.Name)
	if err != nil || !rawID.IsValid() {
		return doltdb.TableName{}, id.Null, err
	}
	return CastIDToTableName(rawID), rawID.AsId(), nil
}

// TableNameToID implements the interface objinterface.Collection.
func (pgc *Collection) TableNameToID(name doltdb.TableName) id.Id {
	return pgc.tableNameToID(name.Schema, name.Name).AsId()
}

// UpdateField implements the interface objinterface.Collection.
func (pgc *Collection) UpdateField(ctx context.Context, rootObject objinterface.RootObject, fieldName string, newValue any) (objinterface.RootObject, error) {
	return nil, errors.New("updating through the conflicts table for this object type is not yet supported")
}
