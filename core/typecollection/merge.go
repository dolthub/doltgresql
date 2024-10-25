// Copyright 2024 Dolthub, Inc.
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
	"fmt"

	"github.com/dolthub/doltgresql/server/types"
)

// Merge handles merging sequences on our root and their root.
func Merge(ctx context.Context, ourCollection, theirCollection, ancCollection *TypeCollection) (*TypeCollection, error) {
	mergedCollection := ourCollection.Clone()
	err := theirCollection.IterateTypes(func(schema string, theirType *types.Type) error {
		// If we don't have the type, then we simply add it
		mergedType, exists := mergedCollection.GetType(schema, theirType.Name)
		if !exists {
			newSeq := *theirType
			return mergedCollection.CreateType(schema, &newSeq)
		}

		// Different types with the same name cannot be merged. (e.g.: 'domain' type and 'base' type with the same name)
		if mergedType.TypType != theirType.TypType {
			return fmt.Errorf(`cannot merge type "%s" because type types do not match: '%s' and '%s'"`, theirType.Name, mergedType.TypType, theirType.TypType)
		}

		switch theirType.TypType {
		case types.TypeType_Domain:
			if mergedType.BaseTypeOID != theirType.BaseTypeOID {
				// TODO: we can extend on this in the future (e.g.: maybe uses preferred type?)
				return fmt.Errorf(`base types of domain type "%s" do not match`, theirType.Name)
			}
			if mergedType.Default == "" {
				mergedType.Default = theirType.Default
			} else if theirType.Default != "" && mergedType.Default != theirType.Default {
				return fmt.Errorf(`default values of domain type "%s" do not match`, theirType.Name)
			}
			// if either of types defined as NOT NULL, take NOT NULL
			if mergedType.NotNull || theirType.NotNull {
				mergedType.NotNull = true
			}
			if len(theirType.Checks) > 0 {
				// TODO: check for duplicate check constraints
				mergedType.Checks = append(mergedType.Checks, theirType.Checks...)
			}
		default:
			// TODO: support merge for other types. (base, range, etc.)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return mergedCollection, nil
}
