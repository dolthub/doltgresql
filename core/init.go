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

package core

import (
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/types"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/plpgsql"
)

// Init initializes this package.
func Init() {
	doltdb.EmptyRootValue = emptyRootValue
	doltdb.NewRootValue = newRootValue
	types.DoltgresRootValueHumanReadableStringAtIndentationLevel = rootValueHumanReadableStringAtIndentationLevel
	types.DoltgresRootValueWalkAddrs = rootValueWalkAddrs
	plpgsql.GetTypesCollectionFromContext = GetTypesCollectionFromContext
	id.RegisterListener(sequenceIDListener{}, id.Section_Table)
}
