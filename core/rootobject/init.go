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

package rootobject

import (
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/merge"

	"github.com/dolthub/doltgresql/core/conflicts"
	pgmerge "github.com/dolthub/doltgresql/core/merge"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/core/storage"
)

// Init initializes the package
func Init() {
	merge.MergeRootObjects = HandleMerge
	pgmerge.CreateConflict = CreateConflict
	conflicts.DeserializeRootObject = DeserializeRootObject
	conflicts.DiffRootObjects = DiffRootObjects
	conflicts.GetFieldType = GetFieldType
	conflicts.UpdateField = UpdateField
	doltdb.RootObjectDiffFromRow = objinterface.DiffFromRow
	for _, collFuncs := range globalCollections {
		if collFuncs == nil {
			continue
		}
		serializer := collFuncs.Serializer()
		storage.RootObjectSerializations = append(storage.RootObjectSerializations, storage.RootObjectSerialization{
			Bytes:        serializer.Bytes,
			RootValueAdd: serializer.RootValueAdd,
		})
	}
}
