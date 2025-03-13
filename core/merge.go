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

package core

import (
	"context"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/merge"
)

// MergeRootObjects handles merging root objects, which is primarily used by Doltgres.
func MergeRootObjects(ctx context.Context, mro merge.MergeRootObject) (doltdb.RootObject, *merge.MergeStats, error) {
	return mro.OurRootObj, &merge.MergeStats{
		Operation:            merge.TableUnmodified,
		Adds:                 0,
		Deletes:              0,
		Modifications:        0,
		DataConflicts:        0,
		SchemaConflicts:      0,
		ConstraintViolations: 0,
	}, nil
}
