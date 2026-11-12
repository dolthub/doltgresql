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

import (
	"math"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// indexBox is the float32 box that PostGIS' index operators compare, held as float64 values. A two-dimensional box
// has two dimensions, and an n-dimensional box has an entry for each of X, Y, Z, and M that the geometry has, where
// an M without a Z places M fourth and pads Z with the largest float32 range.
type indexBox struct {
	mins, maxs []float64
}

// paddedDimension is the bound that marks a dimension an n-dimensional box does not have.
const paddedDimension = math.MaxFloat32

// flatBox is the two-dimensional float32 box that PostGIS' 2D index operators compare, held as float64 values. An
// empty geometry's box holds NaN, which makes most comparisons with it false.
type flatBox struct {
	minX, minY, maxX, maxY float64
}

// box2df returns the two-dimensional index box of the geometry.
func box2df(g Geometry) flatBox {
	b, ok := g.Shape.storedBounds()
	if !ok {
		nan := math.NaN()
		return flatBox{minX: nan, minY: nan, maxX: nan, maxY: nan}
	}
	return flatBox{minX: floatDown(b.minX), minY: floatDown(b.minY), maxX: floatUp(b.maxX), maxY: floatUp(b.maxY)}
}

// isEmpty returns whether the box belongs to an empty geometry.
func (b flatBox) isEmpty() bool {
	return math.IsNaN(b.minX)
}

// gidx returns the n-dimensional index box of the geometry, or false when the geometry is empty.
func gidx(g Geometry) (indexBox, bool) {
	b, ok := g.Shape.storedBounds()
	if !ok {
		return indexBox{}, false
	}
	box := indexBox{
		mins: []float64{floatDown(b.minX), floatDown(b.minY)},
		maxs: []float64{floatUp(b.maxX), floatUp(b.maxY)},
	}
	if g.Shape.hasZ() {
		box.mins, box.maxs = append(box.mins, floatDown(b.minZ)), append(box.maxs, floatUp(b.maxZ))
	}
	if g.Shape.hasM() {
		if !g.Shape.hasZ() {
			box.mins, box.maxs = append(box.mins, -paddedDimension), append(box.maxs, paddedDimension)
		}
		box.mins, box.maxs = append(box.mins, floatDown(b.minM)), append(box.maxs, floatUp(b.maxM))
	}
	return box, true
}

// geographyGidx returns the geocentric index box of a geography, or false when the geography is empty.
func geographyGidx(g Geometry) (indexBox, bool) {
	if g.Shape.isEmpty() {
		return indexBox{}, false
	}
	minPt, maxPt := geocentricBox(g.Shape)
	box := indexBox{mins: make([]float64, 3), maxs: make([]float64, 3)}
	for i := range minPt {
		box.mins[i], box.maxs[i] = floatDown(minPt[i]), floatUp(maxPt[i])
	}
	return box, true
}

// sharedDimensions calls the given test for each dimension that both n-dimensional boxes have and neither has
// padded, stopping and returning false as soon as the test does.
func sharedDimensions(a indexBox, b indexBox, test func(dim int) bool) bool {
	for dim := 0; dim < min(len(a.mins), len(b.mins)); dim++ {
		if a.maxs[dim] != paddedDimension && b.maxs[dim] != paddedDimension && !test(dim) {
			return false
		}
	}
	return true
}

// ndOverlaps returns whether two n-dimensional boxes overlap in every dimension they share.
func ndOverlaps(a indexBox, b indexBox) bool {
	return sharedDimensions(a, b, func(dim int) bool {
		return a.mins[dim] <= b.maxs[dim] && b.mins[dim] <= a.maxs[dim]
	})
}

// ndContains returns whether the first n-dimensional box contains the second in every dimension they share.
func ndContains(a indexBox, b indexBox) bool {
	return sharedDimensions(a, b, func(dim int) bool {
		return a.mins[dim] <= b.mins[dim] && a.maxs[dim] >= b.maxs[dim]
	})
}

// ndEquals returns whether two n-dimensional boxes are equal in every dimension they share.
func ndEquals(a indexBox, b indexBox) bool {
	return sharedDimensions(a, b, func(dim int) bool {
		return a.mins[dim] == b.mins[dim] && a.maxs[dim] == b.maxs[dim]
	})
}

// predicate2D returns an implementation that compares the two-dimensional index boxes of two geometries.
func predicate2D(test func(a flatBox, b flatBox) bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		return test(box2df(args[0].(Geometry)), box2df(args[1].(Geometry))), nil
	}
}

// boxSides returns a two-dimensional test that is false when either box is empty, and otherwise compares a side of
// the first box with a side of the second.
func boxSides(aSide func(flatBox) float64, bSide func(flatBox) float64, compare func(float64, float64) bool) func(a flatBox, b flatBox) bool {
	return func(a flatBox, b flatBox) bool {
		return !a.isEmpty() && !b.isEmpty() && compare(aSide(a), bSide(b))
	}
}

// overlaps2D is the test of the && operator: neither box is empty, and they overlap.
func overlaps2D(a flatBox, b flatBox) bool {
	if a.isEmpty() || b.isEmpty() {
		return false
	}
	return !(a.minX > b.maxX || b.minX > a.maxX || a.minY > b.maxY || b.minY > a.maxY)
}

// contains2D is the test of the ~ operator: the first box contains the second. A non-empty box contains an empty
// one, and since comparisons with an empty box's NaN bounds are false, an empty box contains everything.
func contains2D(a flatBox, b flatBox) bool {
	if b.isEmpty() && !a.isEmpty() {
		return true
	}
	return !(a.minX > b.minX || a.maxX < b.maxX || a.minY > b.minY || a.maxY < b.maxY)
}

// within2D is the test of the @ operator: the first box is within the second. An empty box is within a non-empty one,
// and since comparisons with an empty box's NaN bounds are false, everything is within an empty box.
func within2D(a flatBox, b flatBox) bool {
	if a.isEmpty() && !b.isEmpty() {
		return true
	}
	return !(a.minX < b.minX || a.maxX > b.maxX || a.minY < b.minY || a.maxY > b.maxY)
}

// same2D is the test of the ~= operator: the boxes are equal, or both are empty.
func same2D(a flatBox, b flatBox) bool {
	if a.isEmpty() || b.isEmpty() {
		return a.isEmpty() && b.isEmpty()
	}
	return a == b
}

// predicateND returns an implementation that compares the n-dimensional index boxes of two geometries, or of two
// geographies when `geodetic` is true. The result is false when either is empty.
func predicateND(geodetic bool, test func(a indexBox, b indexBox) bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		boxOf := gidx
		if geodetic {
			boxOf = geographyGidx
		}
		a, okA := boxOf(args[0].(Geometry))
		b, okB := boxOf(args[1].(Geometry))
		return okA && okB && test(a, b), nil
	}
}

// predicate3D returns an implementation that compares the exact three-dimensional boxes of two geometries, which
// PostGIS cannot do when either geometry is empty.
func predicate3D(test func(a box3d, b box3d) bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, err := box3dFromGeometry(ctx, args[0])
		if err != nil {
			return nil, err
		}
		b, err := box3dFromGeometry(ctx, args[1])
		if err != nil {
			return nil, err
		}
		if a == nil || b == nil {
			return nil, pgerror.New(pgcode.Internal, "function returned NULL")
		}
		return test(a.(box3d), b.(box3d)), nil
	}
}

// box3dContains returns whether the first box contains the second.
func box3dContains(a box3d, b box3d) bool {
	return a.maxX >= b.maxX && a.minX <= b.minX && a.maxY >= b.maxY && a.minY <= b.minY && a.maxZ >= b.maxZ && a.minZ <= b.minZ
}

// box3dOverlaps returns whether the boxes overlap.
func box3dOverlaps(a box3d, b box3d) bool {
	return a.minX <= b.maxX && b.minX <= a.maxX && a.minY <= b.maxY && b.minY <= a.maxY && a.minZ <= b.maxZ && b.minZ <= a.maxZ
}

// box3dSame returns whether the boxes are equal within Postgres' geometric tolerance.
func box3dSame(a box3d, b box3d) bool {
	near := func(x float64, y float64) bool { return math.Abs(x-y) <= 1e-6 }
	return near(a.maxX, b.maxX) && near(a.minX, b.minX) && near(a.maxY, b.maxY) && near(a.minY, b.minY) &&
		near(a.maxZ, b.maxZ) && near(a.minZ, b.minZ)
}

// boxDistance implements the <#> operator, which returns the distance between the two-dimensional index boxes of two
// geometries: zero when they overlap, the distance between the nearest corners when one lies diagonally from the
// other, the gap along one axis when they lie side by side, and the largest float32 when either is empty.
func boxDistance(ctx *sql.Context, args ...any) (any, error) {
	a, b := box2df(args[0].(Geometry)), box2df(args[1].(Geometry))
	if overlaps2D(a, b) {
		return 0.0, nil
	}
	left, right := a.maxX < b.minX, a.minX > b.maxX
	below, above := a.maxY < b.minY, a.minY > b.maxY
	if a.isEmpty() || b.isEmpty() || !(left || right || below || above) {
		return float64(math.MaxFloat32), nil
	}
	dx, dy := 0.0, 0.0
	if left {
		dx = b.minX - a.maxX
	} else if right {
		dx = a.minX - b.maxX
	}
	if below {
		dy = b.minY - a.maxY
	} else if above {
		dy = a.minY - b.maxY
	}
	return math.Hypot(dx, dy), nil
}

// bboxRoutines returns the routines behind the bounding box operators.
func bboxRoutines() []extdef.Routine {
	geomParams := []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}
	less := func(x float64, y float64) bool { return x < y }
	lessOrEqual := func(x float64, y float64) bool { return x <= y }
	greater := func(x float64, y float64) bool { return x > y }
	greaterOrEqual := func(x float64, y float64) bool { return x >= y }
	minX := func(b flatBox) float64 { return b.minX }
	minY := func(b flatBox) float64 { return b.minY }
	maxX := func(b flatBox) float64 { return b.maxX }
	maxY := func(b flatBox) float64 { return b.maxY }
	contained := func(a box3d, b box3d) bool { return box3dContains(b, a) }
	within := func(a indexBox, b indexBox) bool { return ndContains(b, a) }
	return []extdef.Routine{
		{Name: "geometry_overlaps", Symbol: "gserialized_overlaps_2d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate2D(overlaps2D)},
		{Name: "geometry_contains", Symbol: "gserialized_contains_2d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate2D(contains2D)},
		{Name: "geometry_within", Symbol: "gserialized_within_2d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate2D(within2D)},
		{Name: "geometry_same", Symbol: "gserialized_same_2d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate2D(same2D)},
		{Name: "geometry_left", Symbol: "gserialized_left_2d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate2D(boxSides(maxX, minX, less))},
		{Name: "geometry_overleft", Symbol: "gserialized_overleft_2d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate2D(boxSides(maxX, maxX, lessOrEqual))},
		{Name: "geometry_right", Symbol: "gserialized_right_2d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate2D(boxSides(minX, maxX, greater))},
		{Name: "geometry_overright", Symbol: "gserialized_overright_2d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate2D(boxSides(minX, minX, greaterOrEqual))},
		{Name: "geometry_below", Symbol: "gserialized_below_2d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate2D(boxSides(maxY, minY, less))},
		{Name: "geometry_overbelow", Symbol: "gserialized_overbelow_2d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate2D(boxSides(maxY, maxY, lessOrEqual))},
		{Name: "geometry_above", Symbol: "gserialized_above_2d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate2D(boxSides(minY, maxY, greater))},
		{Name: "geometry_overabove", Symbol: "gserialized_overabove_2d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate2D(boxSides(minY, minY, greaterOrEqual))},
		{Name: "geometry_distance_box", Symbol: "gserialized_distance_box_2d", Parameters: geomParams, Returns: "float8", Strict: true, Impl: boxDistance},
		{Name: "geometry_overlaps_nd", Symbol: "gserialized_overlaps", Parameters: params("geometry", "geometry"), Returns: "bool", Strict: true, Impl: predicateND(false, ndOverlaps)},
		{Name: "geometry_contains_nd", Symbol: "gserialized_contains", Parameters: params("geometry", "geometry"), Returns: "bool", Strict: true, Impl: predicateND(false, ndContains)},
		{Name: "geometry_within_nd", Symbol: "gserialized_within", Parameters: params("geometry", "geometry"), Returns: "bool", Strict: true, Impl: predicateND(false, within)},
		{Name: "geometry_same_nd", Symbol: "gserialized_same", Parameters: params("geometry", "geometry"), Returns: "bool", Strict: true, Impl: predicateND(false, ndEquals)},
		{Name: "geometry_overlaps_3d", Symbol: "gserialized_overlaps_3d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate3D(box3dOverlaps)},
		{Name: "geometry_contains_3d", Symbol: "gserialized_contains_3d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate3D(box3dContains)},
		{Name: "geometry_contained_3d", Symbol: "gserialized_contained_3d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate3D(contained)},
		{Name: "geometry_same_3d", Symbol: "gserialized_same_3d", Parameters: geomParams, Returns: "bool", Strict: true, Impl: predicate3D(box3dSame)},
		{Name: "geography_overlaps", Symbol: "geography_overlaps", Parameters: params("geography", "geography"), Returns: "bool", Strict: true, Impl: predicateND(true, ndOverlaps)},
	}
}

// bboxOperators returns the bounding box operators.
func bboxOperators() []extdef.Operator {
	geomOp := func(symbol string, routine string, commutator string) extdef.Operator {
		return extdef.Operator{Symbol: symbol, Left: "geometry", Right: "geometry", Routine: routine, Commutator: commutator}
	}
	return []extdef.Operator{
		geomOp("&&", "gserialized_overlaps_2d", "&&"),
		geomOp("~", "gserialized_contains_2d", "@"),
		geomOp("@", "gserialized_within_2d", "~"),
		geomOp("~=", "gserialized_same_2d", "~="),
		geomOp("<<", "gserialized_left_2d", ">>"),
		geomOp("&<", "gserialized_overleft_2d", "&>"),
		geomOp(">>", "gserialized_right_2d", "<<"),
		geomOp("&>", "gserialized_overright_2d", "&<"),
		geomOp("<<|", "gserialized_below_2d", "|>>"),
		geomOp("&<|", "gserialized_overbelow_2d", "|&>"),
		geomOp("|>>", "gserialized_above_2d", "<<|"),
		geomOp("|&>", "gserialized_overabove_2d", "&<|"),
		geomOp("<#>", "gserialized_distance_box_2d", "<#>"),
		geomOp("&&&", "gserialized_overlaps", "&&&"),
		geomOp("~~", "gserialized_contains", "@@"),
		geomOp("@@", "gserialized_within", "~~"),
		geomOp("~~=", "gserialized_same", "~~="),
		geomOp("&/&", "gserialized_overlaps_3d", "&/&"),
		geomOp("@>>", "gserialized_contains_3d", "<<@"),
		geomOp("<<@", "gserialized_contained_3d", "@>>"),
		geomOp("~==", "gserialized_same_3d", "~=="),
		geomOp("<->", "geometry_distance_centroid_geometry_geometry", "<->"),
		geomOp("|=|", "geometry_distance_cpa_geometry_geometry", "|=|"),
		geomOp("<<->>", "gserialized_distance_nd", "<<->>"),
		{Symbol: "&&", Left: "geography", Right: "geography", Routine: "geography_overlaps", Commutator: "&&"},
		{Symbol: "<->", Left: "geography", Right: "geography", Routine: "geography_distance_knn", Commutator: "<->"},
	}
}
