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

// This file holds the GEOS and Wagyu operations that have not been ported to Go yet. Functions that can return an error
// return ErrNotSupported, and the others panic with it.

// GEOS is a context for running GEOS operations, which may only be used by one goroutine at a time.
type GEOS struct{}

// IsValidDetail is not supported yet.
func (g *GEOS) IsValidDetail(a Geometry, flags int) (valid bool, reason string, location Geometry, err error) {
	return false, "", Geometry{}, ErrNotSupported
}

// CoverageUnion is not supported yet.
func (g *GEOS) CoverageUnion(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// MakeValid is not supported yet.
func (g *GEOS) MakeValid(a Geometry, structure bool, keepCollapsed bool) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// SimplifyPreserveTopology is not supported yet.
func (g *GEOS) SimplifyPreserveTopology(a Geometry, tolerance float64) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// LineMerge is not supported yet.
func (g *GEOS) LineMerge(a Geometry, directed bool) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// Node is not supported yet.
func (g *GEOS) Node(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// SharedPaths is not supported yet.
func (g *GEOS) SharedPaths(a Geometry, b Geometry) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// Snap is not supported yet.
func (g *GEOS) Snap(a Geometry, b Geometry, tolerance float64) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// OffsetCurve is not supported yet.
func (g *GEOS) OffsetCurve(a Geometry, distance float64, params BufferParams) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// BuildArea is not supported yet.
func (g *GEOS) BuildArea(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// ConcaveHull is not supported yet.
func (g *GEOS) ConcaveHull(a Geometry, ratio float64, allowHoles bool) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// PolygonHullSimplify is not supported yet.
func (g *GEOS) PolygonHullSimplify(a Geometry, outer bool, vertexFraction float64) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// DelaunayTriangulation is not supported yet.
func (g *GEOS) DelaunayTriangulation(a Geometry, tolerance float64, onlyEdges bool) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// ConstrainedDelaunayTriangulation is not supported yet.
func (g *GEOS) ConstrainedDelaunayTriangulation(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// VoronoiDiagram is not supported yet.
func (g *GEOS) VoronoiDiagram(a Geometry, envelope *Geometry, tolerance float64, onlyEdges bool) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// MinimumClearance is not supported yet.
func (g *GEOS) MinimumClearance(a Geometry) (float64, error) {
	return 0, ErrNotSupported
}

// MinimumClearanceLine is not supported yet.
func (g *GEOS) MinimumClearanceLine(a Geometry) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// ClipByRect is not supported yet.
func (g *GEOS) ClipByRect(a Geometry, minX float64, minY float64, maxX float64, maxY float64) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// MaximumInscribedCircle is not supported yet.
func (g *GEOS) MaximumInscribedCircle(a Geometry, tolerance float64) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// LargestEmptyCircle is not supported yet.
func (g *GEOS) LargestEmptyCircle(obstacles Geometry, boundary *Geometry, tolerance float64) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// ConcaveHullOfPolygons is not supported yet.
func (g *GEOS) ConcaveHullOfPolygons(a Geometry, ratio float64, tight bool, allowHoles bool) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// SetPrecision is not supported yet.
func (g *GEOS) SetPrecision(a Geometry, gridSize float64, flags int) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// Polygonize is not supported yet.
func (g *GEOS) Polygonize(geometries []Geometry) (Geometry, error) {
	return Geometry{}, ErrNotSupported
}

// WagyuClip is not supported yet.
func WagyuClip(rings [][]int64, minX int64, minY int64, maxX int64, maxY int64) ([][][]int64, error) {
	return nil, ErrNotSupported
}
