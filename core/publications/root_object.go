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
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
)

// DeserializeRootObject implements the interface objinterface.Collection.
func (pgp *Collection) DeserializeRootObject(ctx context.Context, data []byte) (objinterface.RootObject, error) {
	return DeserializePublication(ctx, data)
}

// DropRootObject implements the interface objinterface.Collection.
func (pgp *Collection) DropRootObject(ctx context.Context, identifier id.Id) error {
	if identifier.Section() != id.Section_Publication {
		return errors.Errorf(`publication "%s" does not exist`, identifier.String())
	}
	return pgp.DropPublication(ctx, id.Publication(identifier))
}

// GetID implements the interface objinterface.Collection.
func (pgp *Collection) GetID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Publications
}

// GetRootObject implements the interface objinterface.Collection.
func (pgp *Collection) GetRootObject(ctx context.Context, identifier id.Id) (objinterface.RootObject, bool, error) {
	if identifier.Section() != id.Section_Publication {
		return nil, false, nil
	}
	publication, err := pgp.GetPublication(ctx, id.Publication(identifier))
	return publication, err == nil && publication.ID.IsValid(), err
}

// HasRootObject implements the interface objinterface.Collection.
func (pgp *Collection) HasRootObject(ctx context.Context, identifier id.Id) (bool, error) {
	if identifier.Section() != id.Section_Publication {
		return false, nil
	}
	return pgp.HasPublication(ctx, id.Publication(identifier)), nil
}

// IDToTableName implements the interface objinterface.Collection.
func (pgp *Collection) IDToTableName(identifier id.Id) doltdb.TableName {
	if identifier.Section() != id.Section_Publication {
		return doltdb.TableName{}
	}
	return publicationTableName(id.Publication(identifier))
}

// IterAll implements the interface objinterface.Collection.
func (pgp *Collection) IterAll(ctx context.Context, callback func(rootObj objinterface.RootObject) (stop bool, err error)) error {
	for _, publicationID := range pgp.idCache {
		stop, err := callback(pgp.accessCache[publicationID])
		if err != nil {
			return err
		} else if stop {
			return nil
		}
	}
	return nil
}

// IterIDs implements the interface objinterface.Collection.
func (pgp *Collection) IterIDs(ctx context.Context, callback func(identifier id.Id) (stop bool, err error)) error {
	for _, publicationID := range pgp.idCache {
		stop, err := callback(publicationID.AsId())
		if err != nil {
			return err
		} else if stop {
			return nil
		}
	}
	return nil
}

// PutRootObject implements the interface objinterface.Collection.
func (pgp *Collection) PutRootObject(ctx context.Context, rootObj objinterface.RootObject) error {
	publication, ok := rootObj.(Publication)
	if !ok {
		return errors.Newf("invalid publication root object: %T", rootObj)
	}
	return pgp.CreatePublication(ctx, publication)
}

// RenameRootObject implements the interface objinterface.Collection.
func (pgp *Collection) RenameRootObject(ctx context.Context, oldName id.Id, newName id.Id) error {
	return errors.New(`publications cannot be renamed`)
}

// ResolveName implements the interface objinterface.Collection.
func (pgp *Collection) ResolveName(ctx context.Context, name doltdb.TableName) (doltdb.TableName, id.Id, error) {
	identifier := pgp.TableNameToID(name)
	if !identifier.IsValid() || !pgp.HasPublication(ctx, id.Publication(identifier)) {
		return doltdb.TableName{}, id.Null, nil
	}
	return publicationTableName(id.Publication(identifier)), identifier, nil
}

// TableNameToID implements the interface objinterface.Collection.
func (*Collection) TableNameToID(name doltdb.TableName) id.Id {
	// Dolt staging qualifies names through the search path. The synthetic
	// publication spelling remains database-wide regardless of that schema.
	if !strings.HasPrefix(name.Name, "publication ") {
		return id.Null
	}
	publicationName, err := strconv.Unquote(strings.TrimPrefix(name.Name, "publication "))
	if err != nil {
		return id.Null
	}
	return id.NewPublication(publicationName).AsId()
}

// publicationTableName distinguishes publication root objects from SQL relation names.
func publicationTableName(identifier id.Publication) doltdb.TableName {
	return doltdb.TableName{Name: "publication " + strconv.Quote(identifier.Name())}
}
