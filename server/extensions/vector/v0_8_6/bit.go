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

package v0_8_6

import (
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
)

// hammingDistance implements hamming_distance, which counts the differing bits.
func hammingDistance(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(string), args[1].(string)
	if len(a) != len(b) {
		return nil, errors.Errorf("different bit lengths %d and %d", len(a), len(b))
	}
	distance := 0
	for i := range a {
		if a[i] != b[i] {
			distance++
		}
	}
	return float64(distance), nil
}

// jaccardDistance implements jaccard_distance, which returns 1 when neither bit string has a set bit.
func jaccardDistance(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(string), args[1].(string)
	if len(a) != len(b) {
		return nil, errors.Errorf("different bit lengths %d and %d", len(a), len(b))
	}
	intersection, union := 0, 0
	for i := range a {
		if a[i] == '1' && b[i] == '1' {
			intersection++
		}
		if a[i] == '1' || b[i] == '1' {
			union++
		}
	}
	if union == 0 {
		return float64(1), nil
	}
	return 1 - float64(intersection)/float64(union), nil
}
