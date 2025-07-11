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

package extensions

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
)

// DropRootObject implements the interface objinterface.Collection.
func (pge *Collection) DropRootObject(ctx context.Context, identifier id.Id) error {
	if identifier.Section() != id.Section_Extension {
		return errors.Errorf(`extension "%s" does not exist`, identifier.String())
	}
	return pge.DropLoadedExtension(ctx, id.Extension(identifier))
}

// GetID implements the interface objinterface.Collection.
func (pge *Collection) GetID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Extensions
}

// GetRootObject implements the interface objinterface.Collection.
func (pge *Collection) GetRootObject(ctx context.Context, identifier id.Id) (objinterface.RootObject, bool, error) {
	if identifier.Section() != id.Section_Extension {
		return nil, false, nil
	}
	ext, err := pge.GetLoadedExtension(ctx, id.Extension(identifier))
	return ext, err == nil, err
}

// HasRootObject implements the interface objinterface.Collection.
func (pge *Collection) HasRootObject(ctx context.Context, identifier id.Id) (bool, error) {
	if identifier.Section() != id.Section_Extension {
		return false, nil
	}
	return pge.HasLoadedExtension(ctx, id.Extension(identifier)), nil
}

// IDToTableName implements the interface objinterface.Collection.
func (pge *Collection) IDToTableName(identifier id.Id) doltdb.TableName {
	if identifier.Section() != id.Section_Extension {
		return doltdb.TableName{}
	}
	return doltdb.TableName{Name: id.Extension(identifier).Name()}
}

// IterAll implements the interface objinterface.Collection.
func (pge *Collection) IterAll(ctx context.Context, callback func(rootObj objinterface.RootObject) (stop bool, err error)) error {
	for _, extID := range pge.idCache {
		stop, err := callback(pge.accessCache[extID])
		if err != nil {
			return err
		} else if stop {
			return nil
		}
	}
	return nil
}

// IterIDs implements the interface objinterface.Collection.
func (pge *Collection) IterIDs(ctx context.Context, callback func(identifier id.Id) (stop bool, err error)) error {
	for _, extID := range pge.idCache {
		stop, err := callback(extID.AsId())
		if err != nil {
			return err
		} else if stop {
			return nil
		}
	}
	return nil
}

// PutRootObject implements the interface objinterface.Collection.
func (pge *Collection) PutRootObject(ctx context.Context, rootObj objinterface.RootObject) error {
	ext, ok := rootObj.(Extension)
	if !ok {
		return errors.Newf("invalid extension root object: %T", rootObj)
	}
	return pge.AddLoadedExtension(ctx, ext)
}

// RenameRootObject implements the interface objinterface.Collection.
func (pge *Collection) RenameRootObject(ctx context.Context, oldName id.Id, newName id.Id) error {
	return errors.New(`extensions cannot be renamed`)
}

// ResolveName implements the interface objinterface.Collection.
func (pge *Collection) ResolveName(ctx context.Context, name doltdb.TableName) (doltdb.TableName, id.Id, error) {
	extID := id.NewExtension(name.Name)
	if pge.HasLoadedExtension(ctx, extID) {
		return doltdb.TableName{Name: name.Name}, extID.AsId(), nil
	}
	return doltdb.TableName{}, id.Null, nil
}

// TableNameToID implements the interface objinterface.Collection.
func (pge *Collection) TableNameToID(name doltdb.TableName) id.Id {
	return id.NewExtension(name.Name).AsId()
}
