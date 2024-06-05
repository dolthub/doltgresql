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

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/utils"
)

// Merge handles merging sequences on our root and their root.
func Merge(ctx context.Context, ourCollection, theirCollection, ancCollection *Collection) (*Collection, error) {
	mergedCollection := ourCollection.Clone()
	err := theirCollection.IterateSequences(func(schema string, theirSeq *Sequence) error {
		// If we don't have the sequence, then we simply add it
		if !mergedCollection.HasSequence(doltdb.TableName{Name: theirSeq.Name, Schema: schema}) {
			newSeq := *theirSeq
			return mergedCollection.CreateSequence(schema, &newSeq)
		}
		// Take the min/max of fields that aren't dependent on the increment direction
		mergedSeq := mergedCollection.GetSequence(doltdb.TableName{Name: theirSeq.Name, Schema: schema})
		mergedSeq.Minimum = utils.Min(mergedSeq.Minimum, theirSeq.Minimum)
		mergedSeq.Maximum = utils.Max(mergedSeq.Maximum, theirSeq.Maximum)
		mergedSeq.Cache = utils.Min(mergedSeq.Cache, theirSeq.Cache)
		mergedSeq.Cycle = mergedSeq.Cycle || theirSeq.Cycle
		// Take the largest type specified
		if (mergedSeq.DataTypeOID == uint32(oid.T_int2) && (theirSeq.DataTypeOID == uint32(oid.T_int4) || theirSeq.DataTypeOID == uint32(oid.T_int8))) ||
			(mergedSeq.DataTypeOID == uint32(oid.T_int4) && theirSeq.DataTypeOID == uint32(oid.T_int8)) {
			mergedSeq.DataTypeOID = theirSeq.DataTypeOID
		}
		// Handle the fields that are dependent on the increment direction.
		// We'll always take the increment size that's the smallest for the most granularity, along with the one that
		// has progressed the furthest.
		// For opposing increment directions, we'll take whatever is in our collection, therefore there's no else branch.
		if mergedSeq.Increment >= 0 && theirSeq.Increment >= 0 {
			mergedSeq.Increment = utils.Min(mergedSeq.Increment, theirSeq.Increment)
			mergedSeq.Start = utils.Min(mergedSeq.Start, theirSeq.Start)
			mergedSeq.Current = utils.Max(mergedSeq.Current, theirSeq.Current)
		} else if mergedSeq.Increment < 0 && theirSeq.Increment < 0 {
			mergedSeq.Increment = utils.Max(mergedSeq.Increment, theirSeq.Increment)
			mergedSeq.Start = utils.Max(mergedSeq.Start, theirSeq.Start)
			mergedSeq.Current = utils.Min(mergedSeq.Current, theirSeq.Current)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return mergedCollection, nil
}
