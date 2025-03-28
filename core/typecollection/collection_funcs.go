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
	"github.com/dolthub/dolt/go/libraries/doltcore/merge"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"

	"github.com/dolthub/doltgresql/core/id"
	merge2 "github.com/dolthub/doltgresql/core/merge"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/flatbuffers/gen/serial"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// storage is used to read from and write to the root.
var storage = objinterface.RootObjectSerializer{
	Bytes:        (*serial.RootValue).TypesBytes,
	RootValueAdd: serial.RootValueAddTypes,
}

// HandleMerge implements the interface objinterface.Collection.
func (*TypeCollection) HandleMerge(ctx context.Context, mro merge.MergeRootObject) (doltdb.RootObject, *merge.MergeStats, error) {
	ourType := mro.OurRootObj.(TypeWrapper).Type
	theirType := mro.TheirRootObj.(TypeWrapper).Type
	// Ensure that they have the same identifier
	if ourType.ID != theirType.ID {
		return nil, nil, errors.Newf("attempted to merge different types: `%s` and `%s`",
			ourType.ID.TypeName(), theirType.ID.TypeName())
	}
	// Different types with the same name cannot be merged. (e.g.: 'domain' type and 'base' type with the same name)
	if ourType.TypType != theirType.TypType {
		return nil, nil, errors.Errorf(`cannot merge type "%s" because type types do not match: '%s' and '%s'"`,
			theirType.ID.TypeName(), ourType.TypType, theirType.TypType)
	}
	// Check if an ancestor is present
	var ancType pgtypes.DoltgresType
	hasAncestor := false
	if mro.AncestorRootObj != nil {
		ancType = *(mro.AncestorRootObj.(TypeWrapper).Type)
		hasAncestor = true
	}
	mergedType := *ourType
	switch theirType.TypType {
	case pgtypes.TypeType_Domain:
		if ourType.BaseTypeID != theirType.BaseTypeID {
			// TODO: we can extend on this in the future (e.g.: maybe uses preferred type?)
			return nil, nil, errors.Errorf(`base types of domain type "%s" do not match`, theirType.ID.TypeName())
		}
		var err error
		mergedType.Default = merge2.ResolveMergeValues(ourType.Default, theirType.Default, ancType.Default, hasAncestor, func(ourDefault, theirDefault string) string {
			if ourType.Default == "" {
				return theirDefault
			} else if theirType.Default != "" && ourType.Default != theirType.Default {
				err = errors.Errorf(`default values of domain type "%s" do not match`, theirType.ID.TypeName())
				return ourDefault
			} else {
				return ourDefault
			}
		})
		if err != nil {
			return nil, nil, err
		}
		// if either of types defined as NOT NULL, take NOT NULL
		mergedType.NotNull = merge2.ResolveMergeValues(ourType.NotNull, theirType.NotNull, ancType.NotNull, hasAncestor, func(ourNotNull, theirNotNull bool) bool {
			return ourNotNull || theirNotNull
		})
		if len(theirType.Checks) > 0 {
			// TODO: check for duplicate check constraints
			ourType.Checks = append(ourType.Checks, theirType.Checks...)
		}
		return TypeWrapper{Type: &mergedType}, &merge.MergeStats{
			Operation:            merge.TableModified,
			Adds:                 0,
			Deletes:              0,
			Modifications:        1,
			DataConflicts:        0,
			SchemaConflicts:      0,
			ConstraintViolations: 0,
		}, nil
	default:
		// TODO: support merge for other types. (base, range, etc.)
		return nil, nil, errors.Newf("cannot merge `%s` due to unsupported type", ourType.ID.TypeName())
	}
}

// LoadCollection implements the interface objinterface.Collection.
func (*TypeCollection) LoadCollection(ctx context.Context, root objinterface.RootValue) (objinterface.Collection, error) {
	return LoadTypes(ctx, root)
}

// LoadCollectionHash implements the interface objinterface.Collection.
func (*TypeCollection) LoadCollectionHash(ctx context.Context, root objinterface.RootValue) (hash.Hash, error) {
	m, ok, err := storage.GetProllyMap(ctx, root)
	if err != nil || !ok {
		return hash.Hash{}, err
	}
	return m.HashOf(), nil
}

// LoadTypes loads the types collection from the given root.
func LoadTypes(ctx context.Context, root objinterface.RootValue) (*TypeCollection, error) {
	m, ok, err := storage.GetProllyMap(ctx, root)
	if err != nil {
		return nil, err
	}
	if !ok {
		m, err = prolly.NewEmptyAddressMap(root.NodeStore())
		if err != nil {
			return nil, err
		}
	}
	return &TypeCollection{
		accessedMap:   make(map[id.Type]*pgtypes.DoltgresType),
		underlyingMap: m,
		ns:            root.NodeStore(),
	}, nil
}

// Serializer implements the interface objinterface.Collection.
func (*TypeCollection) Serializer() objinterface.RootObjectSerializer {
	return storage
}

// UpdateRoot implements the interface objinterface.Collection.
func (pgs *TypeCollection) UpdateRoot(ctx context.Context, root objinterface.RootValue) (objinterface.RootValue, error) {
	m, err := pgs.Map(ctx)
	if err != nil {
		return nil, err
	}
	return storage.WriteProllyMap(ctx, root, m)
}
