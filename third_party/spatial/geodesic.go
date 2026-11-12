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

// #include <geodesic.h>
import "C"

// Ellipsoid computes geodesics on an ellipsoid of revolution with PROJ's implementation of GeographicLib. Angles are
// in degrees, and distances are in the units of the ellipsoid's radius.
type Ellipsoid struct {
	geodesic C.struct_geod_geodesic
}

// WGS84 is the ellipsoid of the WGS 84 datum.
var WGS84 = NewEllipsoid(6378137, 1/298.257223563)

// NewEllipsoid returns the ellipsoid with the given equatorial radius and flattening.
func NewEllipsoid(radius float64, flattening float64) *Ellipsoid {
	e := &Ellipsoid{}
	C.geod_init(&e.geodesic, C.double(radius), C.double(flattening))
	return e
}

// Radius returns the equatorial radius of the ellipsoid.
func (e *Ellipsoid) Radius() float64 {
	return float64(e.geodesic.a)
}

// Flattening returns the flattening of the ellipsoid.
func (e *Ellipsoid) Flattening() float64 {
	return float64(e.geodesic.f)
}

// Inverse returns the length of the geodesic between two points, along with its azimuths at both points.
func (e *Ellipsoid) Inverse(lat1 float64, lon1 float64, lat2 float64, lon2 float64) (distance float64, azimuth1 float64, azimuth2 float64) {
	var s12, azi1, azi2 C.double
	C.geod_inverse(&e.geodesic, C.double(lat1), C.double(lon1), C.double(lat2), C.double(lon2), &s12, &azi1, &azi2)
	return float64(s12), float64(azi1), float64(azi2)
}

// Direct returns the point at the given distance along the geodesic that leaves a point at the given azimuth.
func (e *Ellipsoid) Direct(lat float64, lon float64, azimuth float64, distance float64) (float64, float64) {
	var lat2, lon2 C.double
	C.geod_direct(&e.geodesic, C.double(lat), C.double(lon), C.double(azimuth), C.double(distance), &lat2, &lon2, nil)
	return float64(lat2), float64(lon2)
}

// Position returns the point at the given distance along the geodesic between two points.
func (e *Ellipsoid) Position(lat1 float64, lon1 float64, lat2 float64, lon2 float64, distance float64) (float64, float64) {
	var line C.struct_geod_geodesicline
	C.geod_inverseline(&line, &e.geodesic, C.double(lat1), C.double(lon1), C.double(lat2), C.double(lon2), C.GEOD_ALL)
	var lat, lon C.double
	C.geod_position(&line, C.double(distance), &lat, &lon, nil)
	return float64(lat), float64(lon)
}

// PolygonArea returns the signed area of the polygon whose vertices have the given latitudes and longitudes, which
// is positive when the vertices run counterclockwise.
func (e *Ellipsoid) PolygonArea(lats []float64, lons []float64) float64 {
	if len(lats) == 0 {
		return 0
	}
	var area C.double
	C.geod_polygonarea(&e.geodesic, (*C.double)(&lats[0]), (*C.double)(&lons[0]), C.int(len(lats)), &area, nil)
	return float64(area)
}
