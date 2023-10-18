// Copyright 2023 Dolthub, Inc.
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

// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package geo

// hilbertInverse converts (x,y) to d on a Hilbert Curve.
// Adapted from `xy2d` from https://en.wikipedia.org/wiki/Hilbert_curve#Applications_and_mapping_algorithms.
func hilbertInverse(n, x, y uint64) uint64 {
	var d uint64
	for s := n / 2; s > 0; s /= 2 {
		var rx uint64
		if (x & s) > 0 {
			rx = 1
		}
		var ry uint64
		if (y & s) > 0 {
			ry = 1
		}
		d += s * s * ((3 * rx) ^ ry)
		x, y = hilbertRotate(n, x, y, rx, ry)
	}
	return d
}

// hilberRoate rotates/flips a quadrant appropriately.
// Adapted from `rot` in https://en.wikipedia.org/wiki/Hilbert_curve#Applications_and_mapping_algorithms.
func hilbertRotate(n, x, y, rx, ry uint64) (uint64, uint64) {
	if ry == 0 {
		if rx == 1 {
			x = n - 1 - x
			y = n - 1 - y
		}

		x, y = y, x
	}
	return x, y
}
