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

import "github.com/tidwall/geodesic"

// Ellipsoid computes geodesics on an ellipsoid of revolution with a Go port of GeographicLib's geodesic routines, the
// same algorithms that PostGIS uses through PROJ. Angles are in degrees, and distances are in the units of the
// ellipsoid's radius.
type Ellipsoid struct {
	ellipsoid *geodesic.Ellipsoid
}

// WGS84 is the ellipsoid of the WGS 84 datum.
var WGS84 = NewEllipsoid(6378137, 1/298.257223563)

// NewEllipsoid returns the ellipsoid with the given equatorial radius and flattening.
func NewEllipsoid(radius float64, flattening float64) *Ellipsoid {
	return &Ellipsoid{ellipsoid: geodesic.NewEllipsoid(radius, flattening)}
}

// Radius returns the equatorial radius of the ellipsoid.
func (e *Ellipsoid) Radius() float64 {
	return e.ellipsoid.Radius()
}

// Flattening returns the flattening of the ellipsoid.
func (e *Ellipsoid) Flattening() float64 {
	return e.ellipsoid.Flattening()
}

// Inverse returns the length of the geodesic between two points, along with its azimuths at both points.
func (e *Ellipsoid) Inverse(lat1 float64, lon1 float64, lat2 float64, lon2 float64) (distance float64, azimuth1 float64, azimuth2 float64) {
	e.ellipsoid.Inverse(lat1, lon1, lat2, lon2, &distance, &azimuth1, &azimuth2)
	return distance, azimuth1, azimuth2
}

// Direct returns the point at the given distance along the geodesic that leaves a point at the given azimuth.
func (e *Ellipsoid) Direct(lat float64, lon float64, azimuth float64, distance float64) (float64, float64) {
	var lat2, lon2 float64
	e.ellipsoid.Direct(lat, lon, azimuth, distance, &lat2, &lon2, nil)
	return lat2, lon2
}

// Position returns the point at the given distance along the geodesic between two points.
func (e *Ellipsoid) Position(lat1 float64, lon1 float64, lat2 float64, lon2 float64, distance float64) (float64, float64) {
	line := e.ellipsoid.InverseLine(lat1, lon1, lat2, lon2, geodesic.All)
	var lat, lon float64
	line.Position(distance, &lat, &lon, nil)
	return lat, lon
}

// PolygonArea returns the signed area of the polygon whose vertices have the given latitudes and longitudes, which
// is positive when the vertices run counterclockwise.
func (e *Ellipsoid) PolygonArea(lats []float64, lons []float64) float64 {
	if len(lats) == 0 {
		return 0
	}
	polygon := e.ellipsoid.PolygonInit(false)
	for i := range lats {
		polygon.AddPoint(lats[i], lons[i])
	}
	var area float64
	polygon.Compute(false, true, &area, nil)
	return area
}
