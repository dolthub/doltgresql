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

package sequences

import (
	"context"
	"sync"
)

// Merge handles merging sequences on our root and their root.
func Merge(ctx context.Context, ourSequence, theirSequence, ancSequence *Collection) (*Collection, error) {
	// TODO: Add all sequences from their root that do not exist in our root.
	//  For sequences with the same name, if the increment is in the same direction, then take the lowest min and
	//  largest max, along with the lowest/largest value and smallest increment value.
	//  If the increments are in different directions, then just use our sequence in full.
	return &Collection{
		schemaMap: ourSequence.schemaMap,
		mutex:     &sync.Mutex{},
	}, nil
}
