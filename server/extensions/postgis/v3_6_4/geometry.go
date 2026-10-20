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

package v3_6_4

// Geometry is the value of the geometry and geography types: a shape and the identifier of its spatial reference
// system.
type Geometry struct {
	SRID  int32
	Shape shape
}

// coordinatesType is the set of dimensions that the points of a shape have, as a bit for Z and a bit for M.
type coordinatesType byte

// The coordinates types.
const (
	dimXY   coordinatesType = 0
	dimXYZ  coordinatesType = 1
	dimXYM  coordinatesType = 2
	dimXYZM coordinatesType = 3
)

// Is3D returns whether points have Z.
func (c coordinatesType) Is3D() bool {
	return c&dimXYZ != 0
}

// IsMeasured returns whether points have M.
func (c coordinatesType) IsMeasured() bool {
	return c&dimXYM != 0
}

// Dimension returns the number of ordinates in each point.
func (c coordinatesType) Dimension() int {
	dims := 2
	if c.Is3D() {
		dims++
	}
	if c.IsMeasured() {
		dims++
	}
	return dims
}

// shape is a geometry in the model that PostGIS uses, which includes the curve, surface, and triangle types. Points,
// linestrings, circular strings, and triangles hold their ordinates in `coords`, with a triangle holding its single
// ring there. Polygons hold the ordinates of each ring in `rings`. Every other type holds its members in `members`,
// and a curve polygon's members are its rings. An empty shape holds none of them.
type shape struct {
	typeNum int
	ctype   coordinatesType
	coords  []float64
	rings   [][]float64
	members []shape
}

// The spatial reference identifier limits, matching the values that PostGIS is built with.
const (
	sridUnknown     = 0
	sridDefault     = 4326
	sridMaximum     = 999999
	sridUserMaximum = 998999
)

// The geometry type numbers that PostGIS uses in type modifiers and in its own type names.
const (
	pointType             = 1
	lineType              = 2
	polygonType           = 3
	multiPointType        = 4
	multiLineType         = 5
	multiPolygonType      = 6
	collectionType        = 7
	circStringType        = 8
	compoundType          = 9
	curvePolyType         = 10
	multiCurveType        = 11
	multiSurfaceType      = 12
	polyhedralSurfaceType = 13
	triangleType          = 14
	tinType               = 15
)

// typeTags holds the upper-case name of every type number, which WKT uses as the type's tag.
var typeTags = map[int]string{
	pointType:             "POINT",
	lineType:              "LINESTRING",
	polygonType:           "POLYGON",
	multiPointType:        "MULTIPOINT",
	multiLineType:         "MULTILINESTRING",
	multiPolygonType:      "MULTIPOLYGON",
	collectionType:        "GEOMETRYCOLLECTION",
	circStringType:        "CIRCULARSTRING",
	compoundType:          "COMPOUNDCURVE",
	curvePolyType:         "CURVEPOLYGON",
	multiCurveType:        "MULTICURVE",
	multiSurfaceType:      "MULTISURFACE",
	polyhedralSurfaceType: "POLYHEDRALSURFACE",
	triangleType:          "TRIANGLE",
	tinType:               "TIN",
}

// typeNames holds the mixed-case name of every type number, which PostGIS uses in messages and type modifiers.
var typeNames = map[int]string{
	pointType:             "Point",
	lineType:              "LineString",
	polygonType:           "Polygon",
	multiPointType:        "MultiPoint",
	multiLineType:         "MultiLineString",
	multiPolygonType:      "MultiPolygon",
	collectionType:        "GeometryCollection",
	circStringType:        "CircularString",
	compoundType:          "CompoundCurve",
	curvePolyType:         "CurvePolygon",
	multiCurveType:        "MultiCurve",
	multiSurfaceType:      "MultiSurface",
	polyhedralSurfaceType: "PolyhedralSurface",
	triangleType:          "Triangle",
	tinType:               "Tin",
}

// typeNameForNumber returns the mixed-case name of the type number.
func typeNameForNumber(typeNum int) string {
	if name, ok := typeNames[typeNum]; ok {
		return name
	}
	return "Invalid type"
}

// hasSequence returns whether shapes of the type number hold their ordinates in a single sequence.
func hasSequence(typeNum int) bool {
	return typeNum == pointType || typeNum == lineType || typeNum == circStringType || typeNum == triangleType
}

// isEmpty returns whether the shape has no points. A collection is empty when each of its members is empty.
func (s shape) isEmpty() bool {
	if hasSequence(s.typeNum) {
		return len(s.coords) == 0
	}
	if s.typeNum == polygonType {
		return len(s.rings) == 0
	}
	for _, member := range s.members {
		if !member.isEmpty() {
			return false
		}
	}
	return true
}

// hasZ returns whether the shape carries Z coordinates.
func (s shape) hasZ() bool {
	return s.ctype.Is3D()
}

// hasM returns whether the shape carries M coordinates.
func (s shape) hasM() bool {
	return s.ctype.IsMeasured()
}

// withDims returns a copy of the shape whose dimensions are relabeled as the given coordinates type, which must have
// the same number of dimensions. Only the labels change, so XYZ ordinates become XYM ordinates with the same values.
func (s shape) withDims(ctype coordinatesType) shape {
	s.ctype = ctype
	if len(s.members) > 0 {
		members := make([]shape, len(s.members))
		for i, member := range s.members {
			members[i] = member.withDims(ctype)
		}
		s.members = members
	}
	return s
}

// sequences returns every ordinate sequence in the shape, in order.
func (s shape) sequences() [][]float64 {
	if hasSequence(s.typeNum) {
		return [][]float64{s.coords}
	}
	if s.typeNum == polygonType {
		return s.rings
	}
	var seqs [][]float64
	for _, member := range s.members {
		seqs = append(seqs, member.sequences()...)
	}
	return seqs
}

// numPoints returns the number of points in the shape.
func (s shape) numPoints() int {
	dims := s.ctype.Dimension()
	count := 0
	for _, seq := range s.sequences() {
		count += len(seq) / dims
	}
	return count
}

// mapCoords returns a copy of the shape where each point's ordinates have been replaced by the given function, which
// receives and returns the ordinates of one point.
func (s shape) mapCoords(fn func(point []float64) []float64) shape {
	dims := s.ctype.Dimension()
	mapSeq := func(seq []float64) []float64 {
		mapped := make([]float64, 0, len(seq))
		for i := 0; i+dims <= len(seq); i += dims {
			mapped = append(mapped, fn(append([]float64(nil), seq[i:i+dims]...))...)
		}
		return mapped
	}
	if hasSequence(s.typeNum) {
		s.coords = mapSeq(s.coords)
	} else if s.typeNum == polygonType {
		rings := make([][]float64, len(s.rings))
		for i, ring := range s.rings {
			rings[i] = mapSeq(ring)
		}
		s.rings = rings
	} else {
		members := make([]shape, len(s.members))
		for i, member := range s.members {
			members[i] = member.mapCoords(fn)
		}
		s.members = members
	}
	return s
}
