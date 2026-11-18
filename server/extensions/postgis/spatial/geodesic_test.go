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

package spatial

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The expected values come from PostGIS 3.6.4 linked against PROJ 9.8.1.
func TestEllipsoid(t *testing.T) {
	distance, azimuth, _ := WGS84.Inverse(42.36, -71.06, 51.507222, -0.1275)
	require.Equal(t, 5278911.716981254, distance)
	require.Equal(t, 53.14657601624982, azimuth)

	lat, lon := WGS84.Direct(42.36, -71.06, 45, 100000)
	require.Equal(t, 42.9932789680445, lat)
	require.Equal(t, -70.19291710821963, lon)

	grs80 := NewEllipsoid(6378137, 1/298.257222101)
	distance, _, _ = grs80.Inverse(42.36, -71.06, 51.507222, -0.1275)
	require.Equal(t, 5278911.717024138, distance)
	require.Equal(t, 6378137.0, grs80.Radius())
	require.Equal(t, 1/298.257222101, grs80.Flattening())

	lat, lon = WGS84.Position(42.36, -71.06, 51.507222, -0.1275, 0.25*5278911.716981254)
	require.InDelta(t, 48.64189274462905, lat, 1e-12)
	require.InDelta(t, -56.681864176712075, lon, 1e-12)
	require.InDelta(t, 12308778361.469454, WGS84.PolygonArea([]float64{0, 0, 1, 1, 0}, []float64{0, 1, 1, 0, 0}), 1e-4)
	require.Equal(t, 0.0, WGS84.PolygonArea(nil, nil))
}
