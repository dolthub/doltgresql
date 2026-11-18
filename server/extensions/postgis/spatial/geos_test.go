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
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The expected values come from PostGIS 3.6.4 with GEOS 3.11.1.

// coordsOf returns a point, linestring, or ring geometry of the given coordinates.
func coordsOf(typ GeometryType, coords ...float64) Geometry {
	return Geometry{Type: typ, Coords: coords}
}

// polygon returns a polygon of the given rings.
func polygon(rings ...[]float64) Geometry {
	g := Geometry{Type: Polygon}
	for _, ring := range rings {
		g.Members = append(g.Members, Geometry{Type: LinearRing, Coords: ring})
	}
	return g
}

// multiPoint returns a multipoint of the given coordinates.
func multiPoint(coords ...float64) Geometry {
	g := Geometry{Type: MultiPoint}
	for i := 0; i+1 < len(coords); i += 2 {
		g.Members = append(g.Members, coordsOf(Point, coords[i], coords[i+1]))
	}
	return g
}

// text writes a geometry as compact WKT, enough to compare results.
func text(g Geometry) string {
	var sb strings.Builder
	seq := func(coords []float64, dims int) {
		sb.WriteByte('(')
		for i := 0; i+dims <= len(coords); i += dims {
			if i > 0 {
				sb.WriteByte(',')
			}
			for d := 0; d < dims; d++ {
				if d > 0 {
					sb.WriteByte(' ')
				}
				sb.WriteString(strconv.FormatFloat(coords[i+d], 'g', -1, 64))
			}
		}
		sb.WriteByte(')')
	}
	dims := 2
	if g.HasZ {
		dims = 3
	}
	names := map[GeometryType]string{Point: "POINT", LineString: "LINESTRING", LinearRing: "LINESTRING", Polygon: "POLYGON",
		MultiPoint: "MULTIPOINT", MultiLineString: "MULTILINESTRING", MultiPolygon: "MULTIPOLYGON", GeometryCollection: "GEOMETRYCOLLECTION"}
	sb.WriteString(names[g.Type])
	if g.HasZ {
		sb.WriteString(" Z ")
	}
	switch g.Type {
	case Point, LineString, LinearRing:
		seq(g.Coords, dims)
	default:
		sb.WriteByte('(')
		for i, member := range g.Members {
			if i > 0 {
				sb.WriteByte(',')
			}
			switch {
			case g.Type == Polygon:
				seq(member.Coords, dims)
			case g.Type == MultiPoint:
				seq(member.Coords, dims)
			default:
				sb.WriteString(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(text(member), "POLYGON"), "LINESTRING"), " Z "))
			}
		}
		sb.WriteByte(')')
	}
	return sb.String()
}

func TestConvexHull(t *testing.T) {
	g := NewGEOS()
	defer g.Close()
	for _, test := range []struct {
		input    Geometry
		expected string
	}{
		{multiPoint(0, 0, 2, 0, 1, 1, 0, 2), "POLYGON((0 0,0 2,2 0,0 0))"},
		{coordsOf(LineString, 0, 0, 5, 5, 10, 0, 5, 2, 3, 8), "POLYGON((0 0,3 8,10 0,0 0))"},
		{multiPoint(0, 0, 1, 1, 2, 2), "LINESTRING(0 0,2 2)"},
		{coordsOf(Point, 1, 2), "POINT(1 2)"},
		{Geometry{Type: Polygon, HasZ: true, Members: []Geometry{{Type: LinearRing, HasZ: true, Coords: []float64{0, 0, 1, 4, 0, 2, 4, 4, 3, 0, 4, 4, 0, 0, 1}}}},
			"POLYGON Z ((0 0 1,0 4 4,4 4 3,4 0 2,0 0 1))"},
	} {
		hull, err := g.ConvexHull(test.input)
		require.NoError(t, err)
		require.Equal(t, test.expected, text(hull))
	}

	// More than 50 points are first reduced to those outside their octagon
	var coords []float64
	for i := 1; i <= 80; i++ {
		coords = append(coords, float64((i*37)%101), float64((i*53)%97))
	}
	hull, err := g.ConvexHull(multiPoint(coords...))
	require.NoError(t, err)
	require.Equal(t, "POLYGON((3 1,1 77,30 89,48 95,75 86,99 76,100 38,95 16,74 9,3 1))", text(hull))
}

func TestMinimumRotatedRectangle(t *testing.T) {
	g := NewGEOS()
	defer g.Close()
	for _, test := range []struct {
		input    Geometry
		expected string
	}{
		{polygon([]float64{0, 0, 4, 0, 4, 1, 0, 1, 0, 0}), "POLYGON((0 1,0 0,4 0,4 1,0 1))"},
		{multiPoint(0, 0, 2, 2), "LINESTRING(0 0,2 2)"},
		{multiPoint(0, 0, 10, 1, 4, 7, 6, -3, 2, 5), "POLYGON((4 7,-1.5 1.5,4.5 -4.5,10 1,4 7))"},
	} {
		rectangle, err := g.MinimumRotatedRectangle(test.input)
		require.NoError(t, err)
		require.Equal(t, test.expected, text(rectangle))
	}

	// The two sides of this rectangle are equally wide, so which one GEOS measures from depends on rounding that differs
	// between platforms, and only the values are compared
	rectangle, err := g.MinimumRotatedRectangle(polygon([]float64{0, 0, 3, 1, 4, 4, 1, 3, 0, 0}))
	require.NoError(t, err)
	expected := []float64{0, 0, 2.4, -0.8, 4, 4.000000000000001, 1.6, 4.800000000000001, 0, 0}
	require.Len(t, rectangle.Members, 1)
	require.InDeltaSlice(t, expected, rectangle.Members[0].Coords, 1e-12)
}

func TestOverlay(t *testing.T) {
	g := NewGEOS()
	defer g.Close()
	a := polygon([]float64{0, 0, 4, 0, 4, 4, 0, 4, 0, 0})
	b := polygon([]float64{2, 2, 6, 2, 6, 6, 2, 6, 2, 2})
	intersection, err := g.Intersection(a, b, -1)
	require.NoError(t, err)
	require.Equal(t, "POLYGON((4 4,4 2,2 2,2 4,4 4))", text(intersection))
	union, err := g.Union(a, b, -1)
	require.NoError(t, err)
	require.Equal(t, "POLYGON((4 0,0 0,0 4,2 4,2 6,6 6,6 2,4 2,4 0))", text(union))
	difference, err := g.Difference(coordsOf(LineString, 0, 0, 10, 0), coordsOf(LineString, 5, -1, 5, 1), -1)
	require.NoError(t, err)
	require.Equal(t, "MULTILINESTRING((0 0,5 0),(5 0,10 0))", text(difference))

	// Points come out in the order of their coordinates, every time
	box := polygon([]float64{0, 0, 3, 0, 3, 3, 0, 3, 0, 0})
	for i := 0; i < 20; i++ {
		points, err := g.Intersection(multiPoint(3, 3, 0, 0, 10, 10, 1, 1, 2, 2, 4, 4), box, -1)
		require.NoError(t, err)
		require.Equal(t, "MULTIPOINT((0 0),(1 1),(2 2),(3 3))", text(points))
	}

	// GEOS interpolates the Z of overlay results and snaps to grids, which are not supported
	withZ := Geometry{Type: LineString, HasZ: true, Coords: []float64{0, 0, 0, 10, 0, 10}}
	_, err = g.Difference(withZ, coordsOf(LineString, 5, -1, 5, 1), -1)
	require.ErrorIs(t, err, ErrNotSupported)
	_, err = g.Intersection(a, b, 0.5)
	require.ErrorIs(t, err, ErrNotSupported)
}

func TestPredicatesAndMeasures(t *testing.T) {
	g := NewGEOS()
	defer g.Close()
	square := polygon([]float64{0, 0, 4, 0, 4, 4, 0, 4, 0, 0})
	matrix, err := g.Relate(square, coordsOf(LineString, 2, 2, 6, 6), 1)
	require.NoError(t, err)
	require.Equal(t, "1020F1102", matrix)
	matches, err := g.RelatePatternMatch(matrix, "1*T***T**")
	require.NoError(t, err)
	require.True(t, matches)
	_, err = g.Relate(square, square, 2)
	require.ErrorIs(t, err, ErrNotSupported)

	distance, err := g.Distance(coordsOf(LineString, 0, 0, 10, 0), coordsOf(Point, 5, 3))
	require.NoError(t, err)
	require.Equal(t, 3.0, distance)
	simple, err := g.IsSimple(coordsOf(LineString, 0, 0, 2, 2, 0, 2, 2, 0))
	require.NoError(t, err)
	require.False(t, simple)
	ring, err := g.IsRing(coordsOf(LineString, 0, 0, 1, 0, 1, 1, 0, 0))
	require.NoError(t, err)
	require.True(t, ring)
	properly, err := g.ContainsProperly(square, coordsOf(Point, 4, 2))
	require.NoError(t, err)
	require.False(t, properly)
	contains, err := g.Contains(square, coordsOf(Point, 2, 2))
	require.NoError(t, err)
	require.True(t, contains)

	buffer, err := g.Buffer(coordsOf(LineString, 0, 0, 10, 0), 1, BufferParams{QuadSegs: 2, EndCap: BufferFlat, Join: BufferMitre, MitreLimit: 5})
	require.NoError(t, err)
	require.Equal(t, "POLYGON((10 1,10 -1,0 -1,0 1,10 1))", text(buffer))

	group, err := g.NewGroup([]Geometry{square, coordsOf(Point, 5, 5), coordsOf(Point, 2, 2)})
	require.NoError(t, err)
	defer group.Close()
	intersects, err := group.Intersects(0, 2)
	require.NoError(t, err)
	require.True(t, intersects)
	distance, err = group.Distance(0, 1)
	require.NoError(t, err)
	require.InDelta(t, 1.4142135623730951, distance, 1e-15)
}

func TestOrientationIndex(t *testing.T) {
	require.Equal(t, 1, orientationIndex(xy{x: 0, y: 0}, xy{x: 1, y: 0}, xy{x: 0, y: 1}))
	require.Equal(t, -1, orientationIndex(xy{x: 0, y: 0}, xy{x: 1, y: 0}, xy{x: 0, y: -1}))
	require.Equal(t, 0, orientationIndex(xy{x: 0, y: 0}, xy{x: 1, y: 1}, xy{x: 2, y: 2}))
	// Nearly collinear points that double precision alone cannot order
	require.Equal(t, 0, orientationIndex(xy{x: 0.1, y: 0.1}, xy{x: 0.3, y: 0.3}, xy{x: 0.7, y: 0.7}))
	require.Equal(t, -1, orientationIndex(xy{x: 1, y: 1}, xy{x: 3, y: 3.0000000000000004}, xy{x: 2, y: 2}))
}

func TestNormalize(t *testing.T) {
	g := NewGEOS()
	defer g.Close()
	collection := Geometry{Type: GeometryCollection, Members: []Geometry{
		coordsOf(Point, 2, 3),
		{Type: MultiLineString, Members: []Geometry{coordsOf(LineString, 0, 0, 1, 1), coordsOf(LineString, 2, 2, 3, 3)}},
		polygon([]float64{0, 10, 0, 0, 10, 0, 10, 10, 0, 10}, []float64{1, 1, 2, 1, 2, 2, 1, 2, 1, 1}, []float64{5, 5, 6, 5, 6, 6, 5, 6, 5, 5}),
	}}
	normalized, err := g.Normalize(collection)
	require.NoError(t, err)
	require.Equal(t, Point, normalized.Members[2].Type)
	require.Equal(t, "POLYGON((0 0,0 10,10 10,10 0,0 0),(5 5,6 5,6 6,5 6,5 5),(1 1,2 1,2 2,1 2,1 1))", text(normalized.Members[0]))
	require.Equal(t, "MULTILINESTRING((2 2,3 3),(0 0,1 1))", text(normalized.Members[1]))
	for _, test := range []struct {
		input    Geometry
		expected string
	}{
		{coordsOf(LineString, 3, 3, 1, 1, 2, 2), "LINESTRING(2 2,1 1,3 3)"},
		{coordsOf(LineString, 0, 0, 1, 0, 1, 1, 0, 0), "LINESTRING(0 0,1 1,1 0,0 0)"},
		{multiPoint(1, 1, 0, 0, 2, 2), "MULTIPOINT((2 2),(1 1),(0 0))"},
	} {
		normalized, err := g.Normalize(test.input)
		require.NoError(t, err)
		require.Equal(t, test.expected, text(normalized))
	}
}

func TestDiscreteDistances(t *testing.T) {
	g := NewGEOS()
	defer g.Close()
	line := coordsOf(LineString, 0, 0, 2, 0)
	bent := coordsOf(LineString, 0, 1, 1, 2, 2, 1)
	distance, err := g.HausdorffDistance(line, bent, -1)
	require.NoError(t, err)
	require.Equal(t, 2.0, distance)
	distance, err = g.HausdorffDistance(line, bent, 0.5)
	require.NoError(t, err)
	require.Equal(t, 2.0, distance)
	distance, err = g.HausdorffDistance(polygon([]float64{0, 0, 0, 2, 1, 2, 2, 2, 2, 0, 0, 0}),
		polygon([]float64{0.5, 0.5, 0.5, 2.5, 1.5, 2.5, 2.5, 2.5, 2.5, 0.5, 0.5, 0.5}), -1)
	require.NoError(t, err)
	require.Equal(t, 0.7071067811865476, distance)
	_, err = g.HausdorffDistance(line, bent, 2)
	require.Error(t, err)

	straight := coordsOf(LineString, 0, 0, 100, 0)
	peaked := coordsOf(LineString, 0, 0, 50, 50, 100, 0)
	distance, err = g.FrechetDistance(straight, peaked, -1)
	require.NoError(t, err)
	require.Equal(t, 70.71067811865476, distance)
	distance, err = g.FrechetDistance(straight, peaked, 0.5)
	require.NoError(t, err)
	require.Equal(t, 50.0, distance)
	distance, err = g.FrechetDistance(coordsOf(LineString, 1, 1, 2, 3, 5, 4), coordsOf(LineString, 0, 0, 3, 1, 6, 6, 9, 2), 0.3)
	require.NoError(t, err)
	require.Equal(t, 4.47213595499958, distance)
}

func TestMinimumBoundingCircle(t *testing.T) {
	g := NewGEOS()
	defer g.Close()
	_, center, radius, err := g.MinimumBoundingCircle(multiPoint(0, 0, 4, 0, 2, 3))
	require.NoError(t, err)
	require.InDeltaSlice(t, []float64{2, 0.833333333333333}, center.Coords, 1e-15)
	require.Equal(t, 2.1666666666666665, radius)
	_, center, radius, err = g.MinimumBoundingCircle(coordsOf(LineString, 0, 0, 10, 0, 3, 1, 5, 2))
	require.NoError(t, err)
	require.Equal(t, []float64{5, 0}, center.Coords)
	require.Equal(t, 5.0, radius)
	circle, center, radius, err := g.MinimumBoundingCircle(coordsOf(Point, 1, 2))
	require.NoError(t, err)
	require.Equal(t, "POINT(1 2)", text(circle))
	require.Equal(t, []float64{1, 2}, center.Coords)
	require.Equal(t, 0.0, radius)
	require.Equal(t, "3.11.1-CAPI-1.17.1", GEOSVersion())
}
