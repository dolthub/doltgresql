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

package functions

import (
	"context"

	"github.com/cockroachdb/errors"
)

// Merge handles merging functions on our root and their root.
func Merge(ctx context.Context, ourCollection, theirCollection, ancCollection *Collection) (*Collection, error) {
	mergedCollection := ourCollection.Clone(ctx)
	err := theirCollection.IterateFunctions(ctx, func(theirFunc Function) (bool, error) {
		// If we don't have the sequence, then we simply add it
		if !mergedCollection.HasFunction(ctx, theirFunc.ID) {
			return false, mergedCollection.AddFunction(ctx, theirFunc)
		}
		// TODO: figure out a decent merge strategy
		return true, errors.Errorf(`unable to merge "%s"`, theirFunc.ID.AsId().String())
	})
	if err != nil {
		return nil, err
	}
	return mergedCollection, nil
}
