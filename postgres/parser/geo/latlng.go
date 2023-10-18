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

import "math"

// NormalizeLatitudeDegrees normalizes latitudes to the range [-90, 90].
func NormalizeLatitudeDegrees(lat float64) float64 {
	// math.Remainder(lat, 360) returns in the range [-180, 180].
	lat = math.Remainder(lat, 360)
	// If we are above 90 degrees, we curve back to 0, e.g. 91 -> 89, 100 -> 80.
	if lat > 90 {
		return 180 - lat
	}
	// If we are below 90 degrees, we curve back towards 0, e.g. -91 -> -89, -100 -> -80.
	if lat < -90 {
		return -180 - lat
	}
	return lat
}

// NormalizeLongitudeDegrees normalizes longitude to the range [-180, 180].
func NormalizeLongitudeDegrees(lng float64) float64 {
	// math.Remainder(lng, 360) returns in the range [-180, 180].
	return math.Remainder(lng, 360)
}
