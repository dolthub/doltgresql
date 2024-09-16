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

package main

import (
	"bytes"
	"crypto/sha256"

	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/utils"
)

// CompareRowsOrdered compares the two rows, enforcing that the order matches between the two rows.
func CompareRowsOrdered(aRows []*pgproto3.DataRow, bRows []*pgproto3.DataRow) bool {
	if len(aRows) != len(bRows) {
		return false
	}
	for rowIdx := range aRows {
		if len(aRows[rowIdx].Values) != len(bRows[rowIdx].Values) {
			return false
		}
		for colIdx := range aRows[rowIdx].Values {
			if !bytes.Equal(aRows[rowIdx].Values[colIdx], bRows[rowIdx].Values[colIdx]) {
				return false
			}
		}
	}
	return true
}

// CompareRowsUnordered compares the two rows. Order is not enforced, however if there are any duplicate rows, then it
// is expected that the duplicate counts match.
func CompareRowsUnordered(aRows []*pgproto3.DataRow, bRows []*pgproto3.DataRow) bool {
	if len(aRows) != len(bRows) {
		return false
	}
	// It's possible that two different rows can hash to the same result, but we're not concerned with that.
	// The same row will always output the same hash, and that's the only property that we really care about.
	aMap := make(map[string]int)
	bMap := make(map[string]int)
	aHasher := sha256.New()
	bHasher := sha256.New()
	for rowIdx := range aRows {
		// Column counts should always match, so this is a sanity check
		if len(aRows[rowIdx].Values) != len(bRows[rowIdx].Values) {
			return false
		}
		aHasher.Reset()
		bHasher.Reset()
		for colIdx := range aRows[rowIdx].Values {
			_, _ = aHasher.Write(aRows[rowIdx].Values[colIdx])
			_, _ = bHasher.Write(bRows[rowIdx].Values[colIdx])
		}
		aHash := string(aHasher.Sum(nil))
		bHash := string(bHasher.Sum(nil))
		if count, ok := aMap[aHash]; ok {
			aMap[aHash] = count + 1
		} else {
			aMap[aHash] = 1
		}
		if count, ok := bMap[bHash]; ok {
			bMap[bHash] = count + 1
		} else {
			bMap[bHash] = 1
		}
	}
	aKVs := utils.GetMapKVsSorted(aMap)
	bKVs := utils.GetMapKVsSorted(bMap)
	// One map may have duplicates that the other does not have, so we have to do a length check again
	if len(aKVs) != len(bKVs) {
		return false
	}
	for i := range aKVs {
		if aKVs[i].Key != bKVs[i].Key || aKVs[i].Value != bKVs[i].Value {
			return false
		}
	}
	return true
}
