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

package publications

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
)

// Collection contains the publication definitions on a database root.
type Collection struct {
	objinterface.RootObjectMap
	accessCache map[id.Publication]Publication // This cache is used for general access
	idCache     []id.Publication               // This cache simply contains the name of every loaded publication
}

// Publication represents a database-wide logical replication publication definition.
// TODO: add explicit table and schema membership when their ownership and dependency handling is supported.
type Publication struct {
	ID              id.Publication
	OwnerRoleID     uint64
	AllTables       bool
	PublishInsert   bool
	PublishUpdate   bool
	PublishDelete   bool
	PublishTruncate bool
	PublishViaRoot  bool
}

var _ objinterface.Collection = (*Collection)(nil)
var _ objinterface.RootObject = Publication{}

// NewCollection returns a new Collection.
func NewCollection(ctx context.Context, rom objinterface.RootObjectMap) (*Collection, error) {
	collection := &Collection{
		RootObjectMap: rom,
		accessCache:   make(map[id.Publication]Publication),
	}
	return collection, collection.reloadCaches(ctx)
}

// GetPublication returns the loaded publication with the given name. Returns a publication with an invalid ID if it
// cannot be found.
func (pgp *Collection) GetPublication(ctx context.Context, name id.Publication) (Publication, error) {
	if f, ok := pgp.accessCache[name]; ok {
		return f, nil
	}
	return Publication{}, nil
}

// HasPublication returns whether the publication has been loaded.
func (pgp *Collection) HasPublication(ctx context.Context, name id.Publication) bool {
	_, ok := pgp.accessCache[name]
	return ok
}

// CreatePublication creates a publication after its definition has been validated.
func (pgp *Collection) CreatePublication(ctx context.Context, p Publication) error {
	// First we'll check to see if it exists
	if !p.ID.IsValid() {
		return errors.New("invalid publication ID")
	}
	if _, ok := pgp.accessCache[p.ID]; ok {
		return errors.Errorf(`publication "%s" already exists`, p.ID.Name())
	}

	// Now we'll add the publication to our map
	data, err := p.Serialize(ctx)
	if err != nil {
		return err
	}
	h, err := pgp.NodeStore().WriteBytes(ctx, data)
	if err != nil {
		return err
	}
	mapEditor := pgp.Contents().Editor()
	if err = mapEditor.Add(ctx, string(p.ID), h); err != nil {
		return err
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pgp.SetContents(newMap)
	return pgp.reloadCaches(ctx)
}

// DropPublication removes the named publications, validating all names before changing the collection.
func (pgp *Collection) DropPublication(ctx context.Context, names ...id.Publication) error {
	if len(names) == 0 {
		return nil
	}
	// Check that each name exists before performing any deletions
	for _, name := range names {
		if _, ok := pgp.accessCache[name]; !ok {
			return errors.Errorf(`publication "%s" does not exist`, name.Name())
		}
	}

	// Now we'll remove the publications from the map
	mapEditor := pgp.Contents().Editor()
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
	pgp.SetContents(newMap)
	return pgp.reloadCaches(ctx)
}

// reloadCaches writes the underlying map's contents to the caches.
func (pgp *Collection) reloadCaches(ctx context.Context) error {
	count, err := pgp.Contents().Count()
	if err != nil {
		return err
	}

	clear(pgp.accessCache)
	pgp.idCache = make([]id.Publication, 0, count)

	return pgp.Contents().IterAll(ctx, func(_ string, h hash.Hash) error {
		if h.IsEmpty() {
			return nil
		}
		data, err := pgp.NodeStore().ReadBytes(ctx, h)
		if err != nil {
			return err
		}
		publication, err := DeserializePublication(ctx, data)
		if err != nil {
			return err
		}
		pgp.accessCache[publication.ID] = publication
		pgp.idCache = append(pgp.idCache, publication.ID)
		return nil
	})
}

// GetID implements the interface objinterface.RootObject.
func (p Publication) GetID() id.Id {
	return p.ID.AsId()
}

// GetRootObjectID implements the interface objinterface.RootObject.
func (p Publication) GetRootObjectID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Publications
}

// HashOf implements the interface objinterface.RootObject.
func (p Publication) HashOf(ctx context.Context) (hash.Hash, error) {
	data, err := p.Serialize(ctx)
	if err != nil {
		return hash.Hash{}, err
	}
	return hash.Of(data), nil
}

// Name implements the interface objinterface.RootObject.
func (p Publication) Name() doltdb.TableName {
	return publicationTableName(p.ID)
}
