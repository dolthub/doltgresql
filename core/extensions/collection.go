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
	"cmp"
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
)

// Collection contains a collection of loaded extensions.
type Collection struct {
	objinterface.RootObjectMap
	accessCache map[id.Extension]Extension // This cache is used for general access
	idCache     []id.Extension             // This cache simply contains the name of every loaded extension
}

// Extension represents an extension that has been installed into a database.
type Extension struct {
	ExtName     id.Extension
	Namespace   id.Namespace
	Relocatable bool
	Version     string
	// TODO: keep track of what it references so I can later delete them
}

var _ objinterface.Collection = (*Collection)(nil)
var _ objinterface.RootObject = Extension{}

// NewCollection returns a new Collection.
func NewCollection(ctx context.Context, rom objinterface.RootObjectMap) (*Collection, error) {
	collection := &Collection{
		RootObjectMap: rom,
		accessCache:   make(map[id.Extension]Extension),
	}
	return collection, collection.reloadCaches(ctx)
}

// GetLoadedExtension returns the loaded extension with the given name. Returns an extension with an invalid ID if it
// cannot be found.
func (pge *Collection) GetLoadedExtension(ctx context.Context, name id.Extension) (Extension, error) {
	if f, ok := pge.accessCache[name]; ok {
		return f, nil
	}
	return Extension{}, nil
}

// HasLoadedExtension returns whether the extension has been loaded.
func (pge *Collection) HasLoadedExtension(ctx context.Context, name id.Extension) bool {
	_, ok := pge.accessCache[name]
	return ok
}

// AddLoadedExtension adds a new extension, that has already been loaded, to the collection.
func (pge *Collection) AddLoadedExtension(ctx context.Context, ext Extension) error {
	// First we'll check to see if it exists
	if _, ok := pge.accessCache[ext.ExtName]; ok {
		return errors.Errorf(`extension "%s" already exists`, ext.ExtName)
	}

	// Now we'll add the extension to our map
	data, err := ext.Serialize(ctx)
	if err != nil {
		return err
	}
	h, err := pge.NodeStore().WriteBytes(ctx, data)
	if err != nil {
		return err
	}
	mapEditor := pge.Contents().Editor()
	if err = mapEditor.Add(ctx, string(ext.ExtName), h); err != nil {
		return err
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pge.SetContents(newMap)
	return pge.reloadCaches(ctx)
}

// DropLoadedExtension drops a loaded extension. This should be called when unloading an extension, but this function
// itself does not perform the necessary logic.
func (pge *Collection) DropLoadedExtension(ctx context.Context, names ...id.Extension) error {
	// TODO: should this also handle the unloading logic?
	if len(names) == 0 {
		return nil
	}
	// Check that each name exists before performing any deletions
	for _, name := range names {
		if _, ok := pge.accessCache[name]; !ok {
			return errors.Errorf(`extension "%s" does not exist`, name)
		}
	}

	// Now we'll remove the extensions from the map
	mapEditor := pge.Contents().Editor()
	for _, name := range names {
		err := mapEditor.Delete(ctx, string(name))
		if err != nil {
			return err
		}
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pge.SetContents(newMap)
	return pge.reloadCaches(ctx)
}

// reloadCaches writes the underlying map's contents to the caches.
func (pge *Collection) reloadCaches(ctx context.Context) error {
	count, err := pge.Contents().Count()
	if err != nil {
		return err
	}

	clear(pge.accessCache)
	pge.idCache = make([]id.Extension, 0, count)

	return pge.Contents().IterAll(ctx, func(_ string, h hash.Hash) error {
		if h.IsEmpty() {
			return nil
		}
		data, err := pge.NodeStore().ReadBytes(ctx, h)
		if err != nil {
			return err
		}
		ext, err := DeserializeExtension(ctx, data)
		if err != nil {
			return err
		}
		pge.accessCache[ext.ExtName] = ext
		pge.idCache = append(pge.idCache, ext.ExtName)
		return nil
	})
}

// CompareVersions compares the version of the extension versus the given extension, as opaque strings.
func (ext Extension) CompareVersions(other Extension) int {
	return cmp.Compare(ext.Version, other.Version)
}

// GetID implements the interface objinterface.RootObject.
func (ext Extension) GetID() id.Id {
	return ext.ExtName.AsId()
}

// GetRootObjectID implements the interface objinterface.RootObject.
func (ext Extension) GetRootObjectID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Extensions
}

// HashOf implements the interface objinterface.RootObject.
func (ext Extension) HashOf(ctx context.Context) (hash.Hash, error) {
	data, err := ext.Serialize(ctx)
	if err != nil {
		return hash.Hash{}, err
	}
	return hash.Of(data), nil
}

// Name implements the interface objinterface.RootObject.
func (ext Extension) Name() doltdb.TableName {
	return doltdb.TableName{Name: ext.ExtName.Name()}
}
