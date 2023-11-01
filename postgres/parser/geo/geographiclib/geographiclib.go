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

// Package geographiclib is a wrapper around the GeographicLib library.
package geographiclib

var (
	// WGS84Spheroid represents the default WGS84 ellipsoid.
	WGS84Spheroid = NewSpheroid(6378137, 1/298.257223563)
)

// Spheroid is an object that can perform geodesic operations
// on a given spheroid.
type Spheroid struct {
	Radius       float64
	Flattening   float64
	SphereRadius float64
}

// NewSpheroid creates a spheroid from a radius and flattening.
func NewSpheroid(radius float64, flattening float64) *Spheroid {
	minorAxis := radius - radius*flattening
	s := &Spheroid{
		Radius:       radius,
		Flattening:   flattening,
		SphereRadius: (radius*2 + minorAxis) / 3,
	}
	return s
}
