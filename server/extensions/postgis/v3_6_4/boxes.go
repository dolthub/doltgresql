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
	"encoding/binary"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// bbox is the bounding box of a shape's vertices in every dimension.
type bbox struct {
	minX, minY, minZ, minM float64
	maxX, maxY, maxZ, maxM float64
}

// box2d is the value of the box2d type.
type box2d struct {
	minX, minY, maxX, maxY float64
}

// box3d is the value of the box3d type, which also carries an SRID.
type box3d struct {
	minX, minY, minZ, maxX, maxY, maxZ float64
	srid                               int32
}

// bounds returns the bounding box of the shape's vertices, or false when the shape is empty. Dimensions that the
// shape lacks are zero.
func (s shape) bounds() (bbox, bool) {
	b := bbox{
		minX: math.Inf(1), minY: math.Inf(1), minZ: math.Inf(1), minM: math.Inf(1),
		maxX: math.Inf(-1), maxY: math.Inf(-1), maxZ: math.Inf(-1), maxM: math.Inf(-1),
	}
	dims := s.ctype.Dimension()
	found := false
	for _, seq := range s.sequences() {
		for i := 0; i+dims <= len(seq); i += dims {
			found = true
			b.minX, b.maxX = math.Min(b.minX, seq[i]), math.Max(b.maxX, seq[i])
			b.minY, b.maxY = math.Min(b.minY, seq[i+1]), math.Max(b.maxY, seq[i+1])
			next := i + 2
			if s.hasZ() {
				b.minZ, b.maxZ = math.Min(b.minZ, seq[next]), math.Max(b.maxZ, seq[next])
				next++
			}
			if s.hasM() {
				b.minM, b.maxM = math.Min(b.minM, seq[next]), math.Max(b.maxM, seq[next])
			}
		}
	}
	//TODO: include the full extent of circular arcs, which can bulge past their vertices
	if !found {
		return bbox{}, false
	}
	if !s.hasZ() {
		b.minZ, b.maxZ = 0, 0
	}
	if !s.hasM() {
		b.minM, b.maxM = 0, 0
	}
	return b, true
}

// floatRounded returns the box widened to the nearest float32 values, which is how PostGIS stores a geometry's box.
func (b bbox) floatRounded() bbox {
	return bbox{
		minX: floatDown(b.minX), minY: floatDown(b.minY), minZ: floatDown(b.minZ), minM: floatDown(b.minM),
		maxX: floatUp(b.maxX), maxY: floatUp(b.maxY), maxZ: floatUp(b.maxZ), maxM: floatUp(b.maxM),
	}
}

// storedBounds returns the box that PostGIS keeps for the shape, which is exact for the shapes that hasExactBox
// describes and widened to float32 values otherwise.
func (s shape) storedBounds() (bbox, bool) {
	b, ok := s.bounds()
	if ok && !hasExactBox(s) {
		b = b.floatRounded()
	}
	return b, ok
}

// floatDown returns the largest float32 value that is at most `val`.
func floatDown(val float64) float64 {
	f := float32(val)
	if float64(f) > val {
		f = math.Nextafter32(f, float32(math.Inf(-1)))
	}
	return float64(f)
}

// floatUp returns the smallest float32 value that is at least `val`.
func floatUp(val float64) float64 {
	f := float32(val)
	if float64(f) < val {
		f = math.Nextafter32(f, float32(math.Inf(1)))
	}
	return float64(f)
}

// scanFloat matches the longest prefix that C's strtod accepts.
var scanFloat = regexp.MustCompile(`^[+-]?(?i:inf(inity)?|nan|0x[0-9a-f]*\.?[0-9a-f]*(p[+-]?[0-9]+)?|([0-9]+\.?[0-9]*|\.[0-9]+)(e[+-]?[0-9]+)?)`)

// parseStrtod parses a number that scanFloat matched, accepting a hexadecimal number without an exponent as C's strtod
// does.
func parseStrtod(match string) (float64, error) {
	if lower := strings.ToLower(match); strings.Contains(lower, "0x") && !strings.Contains(lower, "p") {
		match += "p0"
	}
	return strconv.ParseFloat(match, 64)
}

// scanFormat reads numbers from the input the way C's sscanf does with a format whose conversions are all "%lf". A
// space in the format matches any amount of whitespace, a conversion skips leading whitespace and reads a number,
// and any other character must match exactly. It returns the numbers read before the first mismatch.
func scanFormat(input string, format string) []float64 {
	var vals []float64
	pos := 0
	for i := 0; i < len(format); i++ {
		switch c := format[i]; c {
		case ' ':
			for pos < len(input) && isCSpace(input[pos]) {
				pos++
			}
		case '%':
			i += 2
			for pos < len(input) && isCSpace(input[pos]) {
				pos++
			}
			match := scanFloat.FindString(input[pos:])
			val, err := parseStrtod(match)
			if match == "" || (err != nil && !errors.Is(err, strconv.ErrRange)) {
				return vals
			}
			vals = append(vals, val)
			pos += len(match)
		default:
			if pos >= len(input) || input[pos] != c {
				return vals
			}
			pos++
		}
	}
	return vals
}

// isCSpace returns whether the character is whitespace to C's isspace.
func isCSpace(c byte) bool {
	return c == ' ' || (c >= '\t' && c <= '\r')
}

// box2dIn implements box2d_in, which reads "BOX(xmin ymin,xmax ymax)" in any case.
func box2dIn(ctx *sql.Context, args ...any) (any, error) {
	vals := scanFormat(strings.ToLower(args[0].(string)), "box(%lf %lf,%lf %lf)")
	if len(vals) != 4 {
		return nil, pgerror.New(pgcode.Internal, "box2d parser - couldn't parse.  It should look like: BOX(xmin ymin,xmax ymax)")
	}
	b := box2d{minX: vals[0], minY: vals[1], maxX: vals[2], maxY: vals[3]}
	if b.minX > b.maxX {
		b.minX, b.maxX = b.maxX, b.minX
	}
	if b.minY > b.maxY {
		b.minY, b.maxY = b.maxY, b.minY
	}
	return b, nil
}

// box2dOut implements box2d_out.
func box2dOut(ctx *sql.Context, args ...any) (any, error) {
	b := args[0].(box2d)
	return "BOX(" + formatOrdinate(b.minX, defaultPrecision) + " " + formatOrdinate(b.minY, defaultPrecision) + "," +
		formatOrdinate(b.maxX, defaultPrecision) + " " + formatOrdinate(b.maxY, defaultPrecision) + ")", nil
}

// box3dIn implements box3d_in, which reads "BOX3D(xmin ymin zmin,xmax ymax zmax)" or "BOX3D(xmin ymin,xmax ymax)".
// Like PostGIS, it swaps reversed bounds through a float32, losing precision.
func box3dIn(ctx *sql.Context, args ...any) (any, error) {
	input := args[0].(string)
	if !strings.HasPrefix(input, "BOX3D(") {
		return nil, pgerror.New(pgcode.Internal, "BOX3D parser - doesn't start with BOX3D(")
	}
	var b box3d
	if vals := scanFormat(input, "BOX3D(%le %le %le ,%le %le %le)"); len(vals) == 6 {
		b = box3d{minX: vals[0], minY: vals[1], minZ: vals[2], maxX: vals[3], maxY: vals[4], maxZ: vals[5]}
	} else if vals = scanFormat(input, "BOX3D(%le %le ,%le %le)"); len(vals) == 4 {
		b = box3d{minX: vals[0], minY: vals[1], maxX: vals[2], maxY: vals[3]}
	} else {
		return nil, pgerror.New(pgcode.Internal, "BOX3D parser - couldn't parse.  It should look like: BOX3D(xmin ymin zmin,xmax ymax zmax) or BOX3D(xmin ymin,xmax ymax)")
	}
	if b.minX > b.maxX {
		b.minX, b.maxX = b.maxX, float64(float32(b.minX))
	}
	if b.minY > b.maxY {
		b.minY, b.maxY = b.maxY, float64(float32(b.minY))
	}
	if b.minZ > b.maxZ {
		b.minZ, b.maxZ = b.maxZ, float64(float32(b.minZ))
	}
	return b, nil
}

// box3dOut implements box3d_out.
func box3dOut(ctx *sql.Context, args ...any) (any, error) {
	b := args[0].(box3d)
	vals := []float64{b.minX, b.minY, b.minZ, b.maxX, b.maxY, b.maxZ}
	var sb strings.Builder
	sb.WriteString("BOX3D(")
	for i, val := range vals {
		if i == 3 {
			sb.WriteByte(',')
		} else if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(formatOrdinate(val, defaultPrecision))
	}
	sb.WriteByte(')')
	return sb.String(), nil
}

// box2dFromGeometry implements box2d(geometry), which returns null for an empty geometry.
func box2dFromGeometry(ctx *sql.Context, args ...any) (any, error) {
	b, ok := args[0].(Geometry).Shape.bounds()
	if !ok {
		return nil, nil
	}
	return box2d{minX: b.minX, minY: b.minY, maxX: b.maxX, maxY: b.maxY}, nil
}

// box3dFromGeometry implements box3d(geometry), which returns null for an empty geometry.
func box3dFromGeometry(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	b, ok := g.Shape.bounds()
	if !ok {
		return nil, nil
	}
	return box3d{minX: b.minX, minY: b.minY, minZ: b.minZ, maxX: b.maxX, maxY: b.maxY, maxZ: b.maxZ, srid: g.SRID}, nil
}

// box2dFromBox3d implements box2d(box3d).
func box2dFromBox3d(ctx *sql.Context, args ...any) (any, error) {
	b := args[0].(box3d)
	return box2d{minX: b.minX, minY: b.minY, maxX: b.maxX, maxY: b.maxY}, nil
}

// box3dFromBox2d implements box3d(box2d).
func box3dFromBox2d(ctx *sql.Context, args ...any) (any, error) {
	b := args[0].(box2d)
	return box3d{minX: b.minX, minY: b.minY, maxX: b.maxX, maxY: b.maxY}, nil
}

// rectangle returns a closed ring through the four corners in order.
func rectangle(corners ...[]float64) []float64 {
	var ring []float64
	for _, corner := range append(corners, corners[0]) {
		ring = append(ring, corner...)
	}
	return ring
}

// geometryFromBox2d implements geometry(box2d): a point when the box is a point, a linestring when the box is a
// line, and a polygon otherwise.
func geometryFromBox2d(ctx *sql.Context, args ...any) (any, error) {
	b := args[0].(box2d)
	return Geometry{Shape: flatBoxShape(b.minX, b.minY, b.maxX, b.maxY)}, nil
}

// flatBoxShape returns the two-dimensional shape that covers a box: a point when the box is a point, a linestring
// from its minimum to its maximum when it is a line, and a polygon otherwise.
func flatBoxShape(minX float64, minY float64, maxX float64, maxY float64) shape {
	if minX == maxX && minY == maxY {
		return shape{typeNum: pointType, coords: []float64{minX, minY}}
	}
	if minX == maxX || minY == maxY {
		return shape{typeNum: lineType, coords: []float64{minX, minY, maxX, maxY}}
	}
	ring := rectangle([]float64{minX, minY}, []float64{minX, maxY}, []float64{maxX, maxY}, []float64{maxX, minY})
	return shape{typeNum: polygonType, rings: [][]float64{ring}}
}

// geometryFromBox3d implements geometry(box3d): a point, a linestring, or a polygon when the box is flat in enough
// dimensions, and otherwise a polyhedral surface made of the box's six faces.
func geometryFromBox3d(ctx *sql.Context, args ...any) (any, error) {
	b := args[0].(box3d)
	sameX, sameY, sameZ := b.minX == b.maxX, b.minY == b.maxY, b.minZ == b.maxZ
	corner := func(x float64, y float64, z float64) []float64 { return []float64{x, y, z} }
	var s shape
	if sameX && sameY && sameZ {
		s = shape{typeNum: pointType, coords: corner(b.minX, b.minY, b.minZ)}
	} else if ((sameX || sameY) && sameZ) || ((sameX || sameZ) && sameY) || ((sameY || sameZ) && sameX) {
		s = shape{typeNum: lineType, coords: append(corner(b.minX, b.minY, b.minZ), corner(b.maxX, b.maxY, b.maxZ)...)}
	} else if sameX {
		s = shape{typeNum: polygonType, rings: [][]float64{rectangle(corner(b.minX, b.minY, b.minZ),
			corner(b.minX, b.maxY, b.minZ), corner(b.minX, b.maxY, b.maxZ), corner(b.minX, b.minY, b.maxZ))}}
	} else if sameY {
		s = shape{typeNum: polygonType, rings: [][]float64{rectangle(corner(b.minX, b.minY, b.minZ),
			corner(b.maxX, b.minY, b.minZ), corner(b.maxX, b.minY, b.maxZ), corner(b.minX, b.minY, b.maxZ))}}
	} else if sameZ {
		s = shape{typeNum: polygonType, rings: [][]float64{rectangle(corner(b.minX, b.minY, b.minZ),
			corner(b.minX, b.maxY, b.minZ), corner(b.maxX, b.maxY, b.minZ), corner(b.maxX, b.minY, b.minZ))}}
	} else {
		pts := [8][]float64{
			corner(b.minX, b.minY, b.minZ), corner(b.minX, b.maxY, b.minZ), corner(b.maxX, b.maxY, b.minZ),
			corner(b.maxX, b.minY, b.minZ), corner(b.minX, b.minY, b.maxZ), corner(b.minX, b.maxY, b.maxZ),
			corner(b.maxX, b.maxY, b.maxZ), corner(b.maxX, b.minY, b.maxZ),
		}
		faces := [6][4]int{{0, 1, 2, 3}, {4, 7, 6, 5}, {0, 4, 5, 1}, {3, 2, 6, 7}, {0, 3, 7, 4}, {1, 5, 6, 2}}
		s = shape{typeNum: polyhedralSurfaceType}
		for _, face := range faces {
			ring := rectangle(pts[face[0]], pts[face[1]], pts[face[2]], pts[face[3]])
			s.members = append(s.members, shape{typeNum: polygonType, ctype: dimXYZ, rings: [][]float64{ring}})
		}
	}
	s.ctype = dimXYZ
	return Geometry{Shape: s}, nil
}

// box3dOrdinate returns an implementation that reads one bound of a box3d, choosing the smaller or larger of the two
// values that the box holds for that dimension.
func box3dOrdinate(pick func(float64, float64) float64, ordinates func(box3d) (float64, float64)) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		return pick(ordinates(args[0].(box3d))), nil
	}
}

// stMakeBox2d implements ST_MakeBox2D, which builds the box spanning two points.
func stMakeBox2d(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(Geometry), args[1].(Geometry)
	if err := checkSameSRID("BOX2D_construct", a, b); err != nil {
		return nil, err
	}
	if a.Shape.typeNum != pointType || b.Shape.typeNum != pointType {
		return nil, pgerror.New(pgcode.Internal, "BOX2D_construct: arguments must be points")
	}
	if a.Shape.isEmpty() || b.Shape.isEmpty() {
		return nil, pgerror.New(pgcode.Internal, "BOX2D_construct: args can not be empty points")
	}
	return box2d{
		minX: math.Min(a.Shape.coords[0], b.Shape.coords[0]), minY: math.Min(a.Shape.coords[1], b.Shape.coords[1]),
		maxX: math.Max(a.Shape.coords[0], b.Shape.coords[0]), maxY: math.Max(a.Shape.coords[1], b.Shape.coords[1]),
	}, nil
}

// st3dMakeBox implements ST_3DMakeBox, which builds the box spanning two points, reading Z as zero when a point has
// none.
func st3dMakeBox(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(Geometry), args[1].(Geometry)
	if a.Shape.typeNum != pointType || b.Shape.typeNum != pointType {
		return nil, pgerror.New(pgcode.Internal, "BOX3D_construct: args must be points")
	}
	if a.Shape.isEmpty() || b.Shape.isEmpty() {
		return nil, pgerror.New(pgcode.Internal, "BOX3D_construct: args can not be empty points")
	}
	if err := checkSameSRID("BOX3D_construct", a, b); err != nil {
		return nil, err
	}
	aBox, _ := a.Shape.bounds()
	bBox, _ := b.Shape.bounds()
	return box3d{
		minX: math.Min(aBox.minX, bBox.minX), minY: math.Min(aBox.minY, bBox.minY), minZ: math.Min(aBox.minZ, bBox.minZ),
		maxX: math.Max(aBox.maxX, bBox.maxX), maxY: math.Max(aBox.maxY, bBox.maxY), maxZ: math.Max(aBox.maxZ, bBox.maxZ),
		srid: a.SRID,
	}, nil
}

// stMakeEnvelope implements ST_MakeEnvelope, which builds a rectangle through the given corners in the order given.
func stMakeEnvelope(ctx *sql.Context, args ...any) (any, error) {
	x1, y1, x2, y2 := args[0].(float64), args[1].(float64), args[2].(float64), args[3].(float64)
	ring := rectangle([]float64{x1, y1}, []float64{x1, y2}, []float64{x2, y2}, []float64{x2, y1})
	return Geometry{SRID: args[4].(int32), Shape: shape{typeNum: polygonType, rings: [][]float64{ring}}}, nil
}

// stExpandBox2d implements ST_Expand for box2d, by one distance in both dimensions or by one distance per dimension.
func stExpandBox2d(ctx *sql.Context, args ...any) (any, error) {
	b := args[0].(box2d)
	dx := args[1].(float64)
	dy := dx
	if len(args) > 2 {
		dy = args[2].(float64)
	}
	return box2d{minX: b.minX - dx, minY: b.minY - dy, maxX: b.maxX + dx, maxY: b.maxY + dy}, nil
}

// stExpandBox3d implements ST_Expand for box3d, by one distance in every dimension or by one distance per dimension.
func stExpandBox3d(ctx *sql.Context, args ...any) (any, error) {
	b := args[0].(box3d)
	dx := args[1].(float64)
	dy, dz := dx, dx
	if len(args) > 2 {
		dy, dz = args[2].(float64), args[3].(float64)
	}
	b.minX, b.minY, b.minZ = b.minX-dx, b.minY-dy, b.minZ-dz
	b.maxX, b.maxY, b.maxZ = b.maxX+dx, b.maxY+dy, b.maxZ+dz
	return b, nil
}

// stExpandGeometry implements ST_Expand for geometry, which returns the expanded bounding box as a polygon with the
// geometry's dimensions, or an empty geometry unchanged.
func stExpandGeometry(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	b, ok := g.Shape.bounds()
	if !ok {
		return g, nil
	}
	dx := args[1].(float64)
	dy, dz, dm := dx, dx, dx
	if len(args) > 2 {
		dy, dz, dm = args[2].(float64), args[3].(float64), args[4].(float64)
	}
	b.minX, b.minY, b.minZ, b.minM = b.minX-dx, b.minY-dy, b.minZ-dz, b.minM-dm
	b.maxX, b.maxY, b.maxZ, b.maxM = b.maxX+dx, b.maxY+dy, b.maxZ+dz, b.maxM+dm
	corner := func(x float64, y float64, z float64, m float64) []float64 {
		pt := []float64{x, y}
		if g.Shape.hasZ() {
			pt = append(pt, z)
		}
		if g.Shape.hasM() {
			pt = append(pt, m)
		}
		return pt
	}
	ring := rectangle(corner(b.minX, b.minY, b.minZ, b.minM), corner(b.minX, b.maxY, b.minZ, b.minM),
		corner(b.maxX, b.maxY, b.maxZ, b.maxM), corner(b.maxX, b.minY, b.maxZ, b.maxM))
	return Geometry{SRID: g.SRID, Shape: shape{typeNum: polygonType, ctype: g.Shape.ctype, rings: [][]float64{ring}}}, nil
}

// stEnvelope implements ST_Envelope, which returns the geometry's bounding box as a point, linestring, or polygon,
// or an empty geometry unchanged.
func stEnvelope(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	b, ok := g.Shape.bounds()
	if !ok {
		return g, nil
	}
	return Geometry{SRID: g.SRID, Shape: flatBoxShape(b.minX, b.minY, b.maxX, b.maxY)}, nil
}

// stCombineBox2d implements ST_CombineBBox(box2d, geometry), which is not strict. It widens the box to include the
// box that PostGIS stores for the geometry.
func stCombineBox2d(ctx *sql.Context, args ...any) (any, error) {
	var geomBox bbox
	hasGeomBox := false
	if args[1] != nil {
		geomBox, hasGeomBox = args[1].(Geometry).Shape.storedBounds()
	}
	if args[0] == nil {
		if !hasGeomBox {
			return nil, nil
		}
		return box2d{minX: geomBox.minX, minY: geomBox.minY, maxX: geomBox.maxX, maxY: geomBox.maxY}, nil
	}
	b := args[0].(box2d)
	if !hasGeomBox {
		return b, nil
	}
	return box2d{
		minX: math.Min(b.minX, geomBox.minX), minY: math.Min(b.minY, geomBox.minY),
		maxX: math.Max(b.maxX, geomBox.maxX), maxY: math.Max(b.maxY, geomBox.maxY),
	}, nil
}

// stCombineBox3d implements ST_CombineBBox(box3d, geometry), which is not strict. It widens the box to include the
// geometry's exact bounding box, and takes the geometry's SRID.
func stCombineBox3d(ctx *sql.Context, args ...any) (any, error) {
	if args[1] == nil {
		return args[0], nil
	}
	g := args[1].(Geometry)
	geomBox, ok := g.Shape.bounds()
	if !ok {
		return args[0], nil
	}
	if args[0] == nil {
		return box3d{minX: geomBox.minX, minY: geomBox.minY, minZ: geomBox.minZ,
			maxX: geomBox.maxX, maxY: geomBox.maxY, maxZ: geomBox.maxZ, srid: g.SRID}, nil
	}
	b := args[0].(box3d)
	return box3d{
		minX: math.Min(b.minX, geomBox.minX), minY: math.Min(b.minY, geomBox.minY), minZ: math.Min(b.minZ, geomBox.minZ),
		maxX: math.Max(b.maxX, geomBox.maxX), maxY: math.Max(b.maxY, geomBox.maxY), maxZ: math.Max(b.maxZ, geomBox.maxZ),
		srid: g.SRID,
	}, nil
}

// stCombineBox3dBox3d implements ST_CombineBBox(box3d, box3d), which is not strict and keeps the first box's SRID.
func stCombineBox3dBox3d(ctx *sql.Context, args ...any) (any, error) {
	if args[0] == nil || args[1] == nil {
		if args[0] == nil {
			return args[1], nil
		}
		return args[0], nil
	}
	a, b := args[0].(box3d), args[1].(box3d)
	return box3d{
		minX: math.Min(a.minX, b.minX), minY: math.Min(a.minY, b.minY), minZ: math.Min(a.minZ, b.minZ),
		maxX: math.Max(a.maxX, b.maxX), maxY: math.Max(a.maxY, b.maxY), maxZ: math.Max(a.maxZ, b.maxZ),
		srid: a.srid,
	}, nil
}

// checkSameSRID returns the error PostGIS raises when two geometries given to the named function have different
// SRIDs.
func checkSameSRID(funcName string, a Geometry, b Geometry) error {
	if a.SRID == b.SRID {
		return nil
	}
	return pgerror.Newf(pgcode.Internal, "%s: Operation on mixed SRID geometries (%s, %d) != (%s, %d)",
		funcName, typeNameForNumber(a.Shape.typeNum), a.SRID, typeNameForNumber(b.Shape.typeNum), b.SRID)
}

// box2dCodec stores a box2d as its four bounds.
var box2dCodec = floatsCodec(4, func(vals []float64) any {
	return box2d{minX: vals[0], minY: vals[1], maxX: vals[2], maxY: vals[3]}
}, func(val any) []float64 {
	b := val.(box2d)
	return []float64{b.minX, b.minY, b.maxX, b.maxY}
})

// box3dCodec stores a box3d as its six bounds followed by its SRID, held exactly in a float64.
var box3dCodec = floatsCodec(7, func(vals []float64) any {
	return box3d{minX: vals[0], minY: vals[1], minZ: vals[2], maxX: vals[3], maxY: vals[4], maxZ: vals[5], srid: int32(vals[6])}
}, func(val any) []float64 {
	b := val.(box3d)
	return []float64{b.minX, b.minY, b.minZ, b.maxX, b.maxY, b.maxZ, float64(b.srid)}
})

// floatsCodec builds a storage codec for a value that is a fixed number of float64 values, ordered by those values
// in sequence.
func floatsCodec(count int, decode func([]float64) any, encode func(any) []float64) *pgtypes.TypeCodec {
	deserialize := func(data []byte) (any, error) {
		if len(data) != count*8 {
			return nil, errors.Errorf("malformed serialized box of %d bytes", len(data))
		}
		vals := make([]float64, count)
		for i := range vals {
			vals[i] = math.Float64frombits(binary.LittleEndian.Uint64(data[i*8:]))
		}
		return decode(vals), nil
	}
	return &pgtypes.TypeCodec{
		Serialize: func(val any) ([]byte, error) {
			return appendFloats(nil, encode(val), binary.LittleEndian), nil
		},
		Deserialize: deserialize,
		SerializedCompare: func(left []byte, right []byte) (int, error) {
			a, err := deserialize(left)
			if err != nil {
				return 0, err
			}
			b, err := deserialize(right)
			if err != nil {
				return 0, err
			}
			aVals, bVals := encode(a), encode(b)
			for i := range aVals {
				if aVals[i] != bVals[i] {
					if aVals[i] < bVals[i] {
						return -1, nil
					}
					return 1, nil
				}
			}
			return 0, nil
		},
	}
}

// fixedLengthType declares one of the fixed-length types, such as the box types, which have only input and output
// functions.
func fixedLengthType(name string, length int16, align pgtypes.TypeAlignment, in string, out string, codec *pgtypes.TypeCodec) extdef.Type {
	def := pgtypes.NewBaseTypeDefinition()
	def.TypLength = length
	def.Align = align
	return extdef.Type{Name: name, Definition: def, Input: in, Output: out, Codec: codec}
}

// boxRoutines returns the routines of the box2d and box3d types.
func boxRoutines() []extdef.Routine {
	lesser, greater := math.Min, math.Max
	return []extdef.Routine{
		{Name: "box2d_in", Symbol: "BOX2D_in", Parameters: params("cstring"), Returns: "box2d", Strict: true, Impl: box2dIn},
		{Name: "box2d_out", Symbol: "BOX2D_out", Parameters: params("box2d"), Returns: "cstring", Strict: true, Impl: box2dOut},
		{Name: "box3d_in", Symbol: "BOX3D_in", Parameters: params("cstring"), Returns: "box3d", Strict: true, Impl: box3dIn},
		{Name: "box3d_out", Symbol: "BOX3D_out", Parameters: params("box3d"), Returns: "cstring", Strict: true, Impl: box3dOut},
		{Name: "box2d", Symbol: "BOX3D_to_BOX2D", Parameters: params("box3d"), Returns: "box2d", Strict: true, Impl: box2dFromBox3d},
		{Name: "box2d", Symbol: "LWGEOM_to_BOX2D", Parameters: params("geometry"), Returns: "box2d", Strict: true, Impl: box2dFromGeometry},
		{Name: "box3d", Symbol: "BOX2D_to_BOX3D", Parameters: params("box2d"), Returns: "box3d", Strict: true, Impl: box3dFromBox2d},
		{Name: "box3d", Symbol: "LWGEOM_to_BOX3D", Parameters: params("geometry"), Returns: "box3d", Strict: true, Impl: box3dFromGeometry},
		{Name: "geometry", Symbol: "BOX2D_to_LWGEOM", Parameters: params("box2d"), Returns: "geometry", Strict: true, Impl: geometryFromBox2d},
		{Name: "geometry", Symbol: "BOX3D_to_LWGEOM", Parameters: params("box3d"), Returns: "geometry", Strict: true, Impl: geometryFromBox3d},
		{Name: "st_3dmakebox", Symbol: "BOX3D_construct", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}, Returns: "box3d", Strict: true, Impl: st3dMakeBox},
		{Name: "st_combinebbox", Symbol: "BOX2D_combine", Parameters: params("box2d", "geometry"), Returns: "box2d", Impl: stCombineBox2d},
		{Name: "st_combinebbox", Symbol: "BOX3D_combine_BOX3D", Parameters: params("box3d", "box3d"), Returns: "box3d", Impl: stCombineBox3dBox3d},
		{Name: "st_combinebbox", Symbol: "BOX3D_combine", Parameters: params("box3d", "geometry"), Returns: "box3d", Impl: stCombineBox3d},
		{Name: "st_envelope", Symbol: "LWGEOM_envelope", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stEnvelope},
		{Name: "st_expand", Symbol: "st_expand_box2d_float8", Parameters: params("box2d", "float8"), Returns: "box2d", Strict: true, Impl: stExpandBox2d},
		{Name: "st_expand", Symbol: "st_expand_box3d_float8", Parameters: params("box3d", "float8"), Returns: "box3d", Strict: true, Impl: stExpandBox3d},
		{Name: "st_expand", Symbol: "st_expand_box2d_float8_float8", Parameters: []extdef.Parameter{{Name: "box", Type: "box2d"}, {Name: "dx", Type: "float8"}, {Name: "dy", Type: "float8"}}, Returns: "box2d", Strict: true, Impl: stExpandBox2d},
		{Name: "st_expand", Symbol: "st_expand_box3d_float8_float8_float8", Parameters: []extdef.Parameter{{Name: "box", Type: "box3d"}, {Name: "dx", Type: "float8"}, {Name: "dy", Type: "float8"}, {Name: "dz", Type: "float8", Default: `0`}}, Returns: "box3d", Strict: true, Impl: stExpandBox3d},
		{Name: "st_expand", Symbol: "st_expand_geometry_float8", Parameters: params("geometry", "float8"), Returns: "geometry", Strict: true, Impl: stExpandGeometry},
		{Name: "st_expand", Symbol: "st_expand_geometry_float8_float8_float8_float8", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "dx", Type: "float8"}, {Name: "dy", Type: "float8"}, {Name: "dz", Type: "float8", Default: `0`}, {Name: "dm", Type: "float8", Default: `0`}}, Returns: "geometry", Strict: true, Impl: stExpandGeometry},
		{Name: "st_makebox2d", Symbol: "BOX2D_construct", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}, Returns: "box2d", Strict: true, Impl: stMakeBox2d},
		{Name: "st_makeenvelope", Symbol: "ST_MakeEnvelope", Parameters: []extdef.Parameter{{Type: "float8"}, {Type: "float8"}, {Type: "float8"}, {Type: "float8"}, {Type: "int4", Default: `0`}}, Returns: "geometry", Strict: true, Impl: stMakeEnvelope},
		{Name: "st_xmax", Symbol: "BOX3D_xmax", Parameters: params("box3d"), Returns: "float8", Strict: true, Impl: box3dOrdinate(greater, func(b box3d) (float64, float64) { return b.minX, b.maxX })},
		{Name: "st_xmin", Symbol: "BOX3D_xmin", Parameters: params("box3d"), Returns: "float8", Strict: true, Impl: box3dOrdinate(lesser, func(b box3d) (float64, float64) { return b.minX, b.maxX })},
		{Name: "st_ymax", Symbol: "BOX3D_ymax", Parameters: params("box3d"), Returns: "float8", Strict: true, Impl: box3dOrdinate(greater, func(b box3d) (float64, float64) { return b.minY, b.maxY })},
		{Name: "st_ymin", Symbol: "BOX3D_ymin", Parameters: params("box3d"), Returns: "float8", Strict: true, Impl: box3dOrdinate(lesser, func(b box3d) (float64, float64) { return b.minY, b.maxY })},
		{Name: "st_zmax", Symbol: "BOX3D_zmax", Parameters: params("box3d"), Returns: "float8", Strict: true, Impl: box3dOrdinate(greater, func(b box3d) (float64, float64) { return b.minZ, b.maxZ })},
		{Name: "st_zmin", Symbol: "BOX3D_zmin", Parameters: params("box3d"), Returns: "float8", Strict: true, Impl: box3dOrdinate(lesser, func(b box3d) (float64, float64) { return b.minZ, b.maxZ })},
	}
}

// boxCasts returns the casts between the box types and geometry.
func boxCasts() []extdef.Cast {
	//TODO: add the casts to and from the built-in box type once Doltgres has it
	return []extdef.Cast{
		{Source: "geometry", Target: "box2d", Routine: "LWGEOM_to_BOX2D", CastType: casts.CastType_Implicit},
		{Source: "geometry", Target: "box3d", Routine: "LWGEOM_to_BOX3D", CastType: casts.CastType_Implicit},
		{Source: "box3d", Target: "box2d", Routine: "BOX3D_to_BOX2D", CastType: casts.CastType_Implicit},
		{Source: "box2d", Target: "box3d", Routine: "BOX2D_to_BOX3D", CastType: casts.CastType_Implicit},
		{Source: "box2d", Target: "geometry", Routine: "BOX2D_to_LWGEOM", CastType: casts.CastType_Implicit},
		{Source: "box3d", Target: "geometry", Routine: "BOX3D_to_LWGEOM", CastType: casts.CastType_Implicit},
	}
}
