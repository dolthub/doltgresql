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

//go:build !postgis_native

package spatial

// This file stands in for the native GEOS, PROJ, GeographicLib, and Wagyu bindings when building without the
// postgis_native tag. Functions that can return an error return ErrNativeUnavailable, PROJ's database reads as
// empty, and the others panic with ErrNativeUnavailable.

// GEOS is a GEOS context, which cannot run any operations without the native spatial libraries.
type GEOS struct{}

// Group is a group of converted geometries, which cannot exist without the native spatial libraries.
type Group struct{}

// Transformer transforms coordinates, which cannot exist without the native spatial libraries.
type Transformer struct{}

// Ellipsoid is an ellipsoid of revolution, which only reports its parameters without the native spatial libraries.
type Ellipsoid struct {
	radius     float64
	flattening float64
}

// WGS84 is the ellipsoid of the WGS 84 datum.
var WGS84 = NewEllipsoid(6378137, 1/298.257223563)

// GEOSVersion is unavailable without the native spatial libraries.
func GEOSVersion() string {
	panic(ErrNativeUnavailable)
}

// PROJVersion is unavailable without the native spatial libraries.
func PROJVersion() string {
	panic(ErrNativeUnavailable)
}

// NewGEOS returns a context whose operations all return ErrNativeUnavailable.
func NewGEOS() *GEOS {
	return &GEOS{}
}

// Close does nothing without the native spatial libraries.
func (g *GEOS) Close() {
}

// Intersection is unavailable without the native spatial libraries.
func (g *GEOS) Intersection(a Geometry, b Geometry, gridSize float64) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// Union is unavailable without the native spatial libraries.
func (g *GEOS) Union(a Geometry, b Geometry, gridSize float64) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// Difference is unavailable without the native spatial libraries.
func (g *GEOS) Difference(a Geometry, b Geometry, gridSize float64) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// SymDifference is unavailable without the native spatial libraries.
func (g *GEOS) SymDifference(a Geometry, b Geometry, gridSize float64) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// UnaryUnion is unavailable without the native spatial libraries.
func (g *GEOS) UnaryUnion(a Geometry, gridSize float64) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// ConvexHull is unavailable without the native spatial libraries.
func (g *GEOS) ConvexHull(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// Centroid is unavailable without the native spatial libraries.
func (g *GEOS) Centroid(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// PointOnSurface is unavailable without the native spatial libraries.
func (g *GEOS) PointOnSurface(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// Intersects is unavailable without the native spatial libraries.
func (g *GEOS) Intersects(a Geometry, b Geometry) (bool, error) {
	return false, ErrNativeUnavailable
}

// Disjoint is unavailable without the native spatial libraries.
func (g *GEOS) Disjoint(a Geometry, b Geometry) (bool, error) {
	return false, ErrNativeUnavailable
}

// Touches is unavailable without the native spatial libraries.
func (g *GEOS) Touches(a Geometry, b Geometry) (bool, error) {
	return false, ErrNativeUnavailable
}

// Crosses is unavailable without the native spatial libraries.
func (g *GEOS) Crosses(a Geometry, b Geometry) (bool, error) {
	return false, ErrNativeUnavailable
}

// Within is unavailable without the native spatial libraries.
func (g *GEOS) Within(a Geometry, b Geometry) (bool, error) {
	return false, ErrNativeUnavailable
}

// Contains is unavailable without the native spatial libraries.
func (g *GEOS) Contains(a Geometry, b Geometry) (bool, error) {
	return false, ErrNativeUnavailable
}

// Overlaps is unavailable without the native spatial libraries.
func (g *GEOS) Overlaps(a Geometry, b Geometry) (bool, error) {
	return false, ErrNativeUnavailable
}

// Equals is unavailable without the native spatial libraries.
func (g *GEOS) Equals(a Geometry, b Geometry) (bool, error) {
	return false, ErrNativeUnavailable
}

// Covers is unavailable without the native spatial libraries.
func (g *GEOS) Covers(a Geometry, b Geometry) (bool, error) {
	return false, ErrNativeUnavailable
}

// CoveredBy is unavailable without the native spatial libraries.
func (g *GEOS) CoveredBy(a Geometry, b Geometry) (bool, error) {
	return false, ErrNativeUnavailable
}

// Relate is unavailable without the native spatial libraries.
func (g *GEOS) Relate(a Geometry, b Geometry, boundaryNodeRule int) (string, error) {
	return "", ErrNativeUnavailable
}

// RelatePatternMatch is unavailable without the native spatial libraries.
func (g *GEOS) RelatePatternMatch(matrix string, pattern string) (bool, error) {
	return false, ErrNativeUnavailable
}

// Buffer is unavailable without the native spatial libraries.
func (g *GEOS) Buffer(a Geometry, distance float64, params BufferParams) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// IsValidDetail is unavailable without the native spatial libraries.
func (g *GEOS) IsValidDetail(a Geometry, flags int) (valid bool, reason string, location Geometry, err error) {
	return false, "", Geometry{}, ErrNativeUnavailable
}

// CoverageUnion is unavailable without the native spatial libraries.
func (g *GEOS) CoverageUnion(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// NewGroup is unavailable without the native spatial libraries.
func (g *GEOS) NewGroup(geometries []Geometry) (*Group, error) {
	return nil, ErrNativeUnavailable
}

// Close does nothing without the native spatial libraries.
func (group *Group) Close() {
}

// Intersects is unavailable without the native spatial libraries.
func (group *Group) Intersects(i int, j int) (bool, error) {
	return false, ErrNativeUnavailable
}

// Distance is unavailable without the native spatial libraries.
func (group *Group) Distance(i int, j int) (float64, error) {
	return 0, ErrNativeUnavailable
}

// Distance is unavailable without the native spatial libraries.
func (g *GEOS) Distance(a Geometry, b Geometry) (float64, error) {
	return 0, ErrNativeUnavailable
}

// MinimumRotatedRectangle is unavailable without the native spatial libraries.
func (g *GEOS) MinimumRotatedRectangle(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// IsSimple is unavailable without the native spatial libraries.
func (g *GEOS) IsSimple(a Geometry) (bool, error) {
	return false, ErrNativeUnavailable
}

// IsRing is unavailable without the native spatial libraries.
func (g *GEOS) IsRing(a Geometry) (bool, error) {
	return false, ErrNativeUnavailable
}

// ContainsProperly is unavailable without the native spatial libraries.
func (g *GEOS) ContainsProperly(a Geometry, b Geometry) (bool, error) {
	return false, ErrNativeUnavailable
}

// Normalize is unavailable without the native spatial libraries.
func (g *GEOS) Normalize(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// MakeValid is unavailable without the native spatial libraries.
func (g *GEOS) MakeValid(a Geometry, structure bool, keepCollapsed bool) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// SimplifyPreserveTopology is unavailable without the native spatial libraries.
func (g *GEOS) SimplifyPreserveTopology(a Geometry, tolerance float64) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// LineMerge is unavailable without the native spatial libraries.
func (g *GEOS) LineMerge(a Geometry, directed bool) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// Node is unavailable without the native spatial libraries.
func (g *GEOS) Node(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// SharedPaths is unavailable without the native spatial libraries.
func (g *GEOS) SharedPaths(a Geometry, b Geometry) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// Snap is unavailable without the native spatial libraries.
func (g *GEOS) Snap(a Geometry, b Geometry, tolerance float64) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// OffsetCurve is unavailable without the native spatial libraries.
func (g *GEOS) OffsetCurve(a Geometry, distance float64, params BufferParams) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// BuildArea is unavailable without the native spatial libraries.
func (g *GEOS) BuildArea(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// ConcaveHull is unavailable without the native spatial libraries.
func (g *GEOS) ConcaveHull(a Geometry, ratio float64, allowHoles bool) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// PolygonHullSimplify is unavailable without the native spatial libraries.
func (g *GEOS) PolygonHullSimplify(a Geometry, outer bool, vertexFraction float64) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// DelaunayTriangulation is unavailable without the native spatial libraries.
func (g *GEOS) DelaunayTriangulation(a Geometry, tolerance float64, onlyEdges bool) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// ConstrainedDelaunayTriangulation is unavailable without the native spatial libraries.
func (g *GEOS) ConstrainedDelaunayTriangulation(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// VoronoiDiagram is unavailable without the native spatial libraries.
func (g *GEOS) VoronoiDiagram(a Geometry, envelope *Geometry, tolerance float64, onlyEdges bool) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// MinimumClearance is unavailable without the native spatial libraries.
func (g *GEOS) MinimumClearance(a Geometry) (float64, error) {
	return 0, ErrNativeUnavailable
}

// MinimumClearanceLine is unavailable without the native spatial libraries.
func (g *GEOS) MinimumClearanceLine(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// FrechetDistance is unavailable without the native spatial libraries.
func (g *GEOS) FrechetDistance(a Geometry, b Geometry, densifyFraction float64) (float64, error) {
	return 0, ErrNativeUnavailable
}

// HausdorffDistance is unavailable without the native spatial libraries.
func (g *GEOS) HausdorffDistance(a Geometry, b Geometry, densifyFraction float64) (float64, error) {
	return 0, ErrNativeUnavailable
}

// ClipByRect is unavailable without the native spatial libraries.
func (g *GEOS) ClipByRect(a Geometry, minX float64, minY float64, maxX float64, maxY float64) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// MaximumInscribedCircle is unavailable without the native spatial libraries.
func (g *GEOS) MaximumInscribedCircle(a Geometry, tolerance float64) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// LargestEmptyCircle is unavailable without the native spatial libraries.
func (g *GEOS) LargestEmptyCircle(obstacles Geometry, boundary *Geometry, tolerance float64) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// ConcaveHullOfPolygons is unavailable without the native spatial libraries.
func (g *GEOS) ConcaveHullOfPolygons(a Geometry, ratio float64, tight bool, allowHoles bool) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// MinimumBoundingCircle is unavailable without the native spatial libraries.
func (g *GEOS) MinimumBoundingCircle(a Geometry) (circle Geometry, center Geometry, radius float64, err error) {
	return Geometry{}, Geometry{}, 0, ErrNativeUnavailable
}

// SetPrecision is unavailable without the native spatial libraries.
func (g *GEOS) SetPrecision(a Geometry, gridSize float64, flags int) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// Polygonize is unavailable without the native spatial libraries.
func (g *GEOS) Polygonize(geometries []Geometry) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// Boundary is unavailable without the native spatial libraries.
func (g *GEOS) Boundary(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNativeUnavailable
}

// DatabaseCRSInfo is unavailable without the native spatial libraries.
func DatabaseCRSInfo(authority string, within *[4]float64) ([]CRSInfo, error) {
	return nil, ErrNativeUnavailable
}

// CRSDefinitions is unavailable without the native spatial libraries.
func CRSDefinitions(authority string, code string) (wkt string, projString string, err error) {
	return "", "", ErrNativeUnavailable
}

// NewTransformer is unavailable without the native spatial libraries.
func NewTransformer(source string, target string) (*Transformer, error) {
	return nil, ErrNativeUnavailable
}

// NewPipelineTransformer is unavailable without the native spatial libraries.
func NewPipelineTransformer(pipeline string, forward bool) (*Transformer, error) {
	return nil, ErrNativeUnavailable
}

// Reads returns false, as there is no PROJ database to read definitions from without the native spatial libraries.
// Unlike the other stubs, it does not panic, since type modifiers are checked with it while analyzing a query.
func Reads(definition string) bool {
	return false
}

// Close does nothing without the native spatial libraries.
func (t *Transformer) Close() {
}

// Transform is unavailable without the native spatial libraries.
func (t *Transformer) Transform(coords []float64, dims int, hasZ bool) error {
	return ErrNativeUnavailable
}

// PROJSettings is unavailable without the native spatial libraries.
func PROJSettings() (networkEnabled bool, urlEndpoint string, writableDirectory string, databasePath string) {
	panic(ErrNativeUnavailable)
}

// IsGeographic returns false, as PROJ cannot read any definition without the native spatial libraries.
func IsGeographic(definition string) bool {
	return false
}

// NewEllipsoid returns the ellipsoid with the given equatorial radius and flattening.
func NewEllipsoid(radius float64, flattening float64) *Ellipsoid {
	return &Ellipsoid{radius: radius, flattening: flattening}
}

// Radius returns the equatorial radius of the ellipsoid.
func (e *Ellipsoid) Radius() float64 {
	return e.radius
}

// Flattening returns the flattening of the ellipsoid.
func (e *Ellipsoid) Flattening() float64 {
	return e.flattening
}

// Inverse is unavailable without the native spatial libraries.
func (e *Ellipsoid) Inverse(lat1 float64, lon1 float64, lat2 float64, lon2 float64) (distance float64, azimuth1 float64, azimuth2 float64) {
	panic(ErrNativeUnavailable)
}

// Direct is unavailable without the native spatial libraries.
func (e *Ellipsoid) Direct(lat float64, lon float64, azimuth float64, distance float64) (float64, float64) {
	panic(ErrNativeUnavailable)
}

// Position is unavailable without the native spatial libraries.
func (e *Ellipsoid) Position(lat1 float64, lon1 float64, lat2 float64, lon2 float64, distance float64) (float64, float64) {
	panic(ErrNativeUnavailable)
}

// PolygonArea is unavailable without the native spatial libraries.
func (e *Ellipsoid) PolygonArea(lats []float64, lons []float64) float64 {
	panic(ErrNativeUnavailable)
}

// WagyuClip is unavailable without the native spatial libraries.
func WagyuClip(rings [][]int64, minX int64, minY int64, maxX int64, maxY int64) ([][][]int64, error) {
	return nil, ErrNativeUnavailable
}
