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
	"strconv"
	"strings"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

// The errors that the WKT reader returns, worded as PostGIS words them.
var (
	errWKTInvalid       = pgerror.New(pgcode.Internal, "parse error - invalid geometry")
	errWKTMorePoints    = pgerror.New(pgcode.Internal, "geometry requires more points")
	errWKTTooManyPoints = pgerror.New(pgcode.Internal, "geometry has too many points")
	errWKTOddPoints     = pgerror.New(pgcode.Internal, "geometry must have an odd number of points")
	errWKTUnclosedRings = pgerror.New(pgcode.Internal, "geometry contains non-closed rings")
	errWKTMixedDims     = pgerror.New(pgcode.Internal, "can not mix dimensionality in a geometry")
	errWKTIncontinuous  = pgerror.New(pgcode.Internal, "incontinuous compound curve")
	errWKTTrianglePts   = pgerror.New(pgcode.Internal, "triangle must have exactly 4 points")
	errDimsMismatch     = pgerror.New(pgcode.Internal, "Dimensions mismatch in lwcollection")
)

// wktTags maps each WKT tag to its type number.
var wktTags = func() map[string]int {
	tags := make(map[string]int, len(typeTags))
	for typeNum, tag := range typeTags {
		tags[tag] = typeNum
	}
	return tags
}()

// wktReader parses the extended WKT that PostGIS accepts. Like PostGIS, it builds each geometry from the inside out:
// a point's dimensions come from its number of ordinates, and a "Z", "M", or "ZM" after a tag must then agree with
// the contents, which it relabels as needed.
type wktReader struct {
	input string
	pos   int
}

// parseEWKT parses WKT that may be preceded by an "SRID=<srid>;" prefix.
func parseEWKT(input string) (Geometry, error) {
	srid, body, err := splitSRIDPrefix(input)
	if err != nil {
		return Geometry{}, err
	}
	s, err := parseWKT(body)
	if err != nil {
		return Geometry{}, err
	}
	return Geometry{SRID: srid, Shape: s}, nil
}

// splitSRIDPrefix removes an "SRID=<srid>;" prefix from the input, returning the SRID it named.
func splitSRIDPrefix(input string) (int32, string, error) {
	if len(input) < 5 || !strings.EqualFold(input[:5], "SRID=") {
		return sridUnknown, input, nil
	}
	end := strings.IndexByte(input, ';')
	if end < 0 {
		return 0, "", errWKTInvalid
	}
	srid, err := strconv.Atoi(strings.TrimSpace(input[5:end]))
	if err != nil {
		return 0, "", errWKTInvalid
	}
	return int32(srid), input[end+1:], nil
}

// parseWKT parses WKT without an SRID prefix.
func parseWKT(input string) (shape, error) {
	r := &wktReader{input: input}
	s, err := r.readTagged()
	if err != nil {
		return shape{}, err
	}
	if r.peek() != "" {
		return shape{}, errWKTInvalid
	}
	return s, nil
}

// readTagged reads a geometry that begins with its tag, which may be joined to or followed by its dimensions.
func (r *wktReader) readTagged() (shape, error) {
	tag := strings.ToUpper(r.next())
	suffix := ""
	typeNum, ok := wktTags[tag]
	for _, dims := range []string{"ZM", "Z", "M"} {
		if !ok && strings.HasSuffix(tag, dims) {
			if typeNum, ok = wktTags[strings.TrimSuffix(tag, dims)]; ok {
				suffix = dims
			}
		}
	}
	if !ok {
		return shape{}, errWKTInvalid
	}
	if next := strings.ToUpper(r.peek()); suffix == "" && (next == "Z" || next == "M" || next == "ZM") {
		suffix = next
		r.next()
	}
	if strings.EqualFold(r.peek(), "EMPTY") {
		r.next()
		return shape{typeNum: typeNum, ctype: suffixDims(suffix)}, nil
	}
	s, err := r.readParenthesized(typeNum)
	if err != nil {
		return shape{}, err
	}
	if s, err = finalizeDims(s, suffix); err != nil {
		return shape{}, err
	}
	for _, member := range s.members {
		if member.ctype != s.ctype {
			return shape{}, errDimsMismatch
		}
	}
	return s, nil
}

// finalizeDims applies the dimensions given after a tag to a geometry read from its contents. The contents must have
// the same number of dimensions, and the members of a geometry collection must have exactly those dimensions.
func finalizeDims(s shape, suffix string) (shape, error) {
	if suffix == "" {
		return s, nil
	}
	ctype := suffixDims(suffix)
	if isCollectionLike(s.typeNum) {
		for _, member := range s.members {
			if member.isEmpty() {
				continue
			}
			if member.ctype.Dimension() != ctype.Dimension() || (s.typeNum == collectionType && member.ctype != ctype) {
				return shape{}, errWKTMixedDims
			}
		}
	} else if s.ctype.Dimension() != ctype.Dimension() {
		return shape{}, errWKTMixedDims
	}
	return s.withDims(ctype), nil
}

// isCollectionLike returns whether the type number is a multi-geometry, a geometry collection, or a surface made of
// patches, whose members' dimensions are checked individually.
func isCollectionLike(typeNum int) bool {
	switch typeNum {
	case multiPointType, multiLineType, multiPolygonType, collectionType, multiCurveType, multiSurfaceType,
		polyhedralSurfaceType, tinType:
		return true
	default:
		return false
	}
}

// readParenthesized reads the parenthesized contents of a tagged geometry.
func (r *wktReader) readParenthesized(typeNum int) (shape, error) {
	if err := r.expect("("); err != nil {
		return shape{}, err
	}
	s, err := r.readContents(typeNum)
	if err != nil {
		return shape{}, err
	}
	return s, r.expect(")")
}

// readContents reads what lies within a geometry's outer parentheses.
func (r *wktReader) readContents(typeNum int) (shape, error) {
	switch typeNum {
	case pointType:
		coords, ctype, err := r.readPoints()
		if err != nil {
			return shape{}, err
		}
		if len(coords) != ctype.Dimension() {
			return shape{}, errWKTTooManyPoints
		}
		return shape{typeNum: pointType, ctype: ctype, coords: coords}, nil
	case lineType, circStringType:
		return r.readCurveContents(typeNum)
	case triangleType:
		return r.readTriangleRing()
	case polygonType:
		return r.readRings(false)
	case compoundType:
		return r.readMembers(compoundType, r.readCompoundMember)
	case curvePolyType:
		return r.readMembers(curvePolyType, r.readCurveRing)
	case multiPointType:
		return r.readMembers(multiPointType, r.readMultiPointMember)
	case multiLineType:
		return r.readMembers(multiLineType, func() (shape, error) { return r.readUntagged(lineType) })
	case multiPolygonType:
		return r.readMembers(multiPolygonType, func() (shape, error) { return r.readUntagged(polygonType) })
	case multiCurveType:
		return r.readMembers(multiCurveType, r.readCurveMember)
	case multiSurfaceType:
		return r.readMembers(multiSurfaceType, r.readSurfaceMember)
	case polyhedralSurfaceType:
		return r.readMembers(polyhedralSurfaceType, r.readPatch)
	case tinType:
		return r.readMembers(tinType, r.readUntaggedTriangle)
	default:
		return r.readMembers(collectionType, r.readTagged)
	}
}

// readCurveContents reads the points of a linestring or circular string and checks their count.
func (r *wktReader) readCurveContents(typeNum int) (shape, error) {
	coords, ctype, err := r.readPoints()
	if err != nil {
		return shape{}, err
	}
	count := len(coords) / ctype.Dimension()
	if typeNum == lineType && count < 2 {
		return shape{}, errWKTMorePoints
	}
	if typeNum == circStringType {
		if count < 3 {
			return shape{}, errWKTMorePoints
		}
		if count%2 == 0 {
			return shape{}, errWKTOddPoints
		}
	}
	return shape{typeNum: typeNum, ctype: ctype, coords: coords}, nil
}

// readRings reads the parenthesized rings of a polygon, or of a polyhedral surface's patch when `patch` is true. Each
// ring needs four points and must be closed, which a patch checks in three dimensions and a polygon in two.
func (r *wktReader) readRings(patch bool) (shape, error) {
	s := shape{typeNum: polygonType}
	for {
		if err := r.expect("("); err != nil {
			return shape{}, err
		}
		coords, ctype, err := r.readPoints()
		if err != nil {
			return shape{}, err
		}
		if err = r.expect(")"); err != nil {
			return shape{}, err
		}
		if len(s.rings) == 0 {
			s.ctype = ctype
		} else if ctype.Dimension() != s.ctype.Dimension() {
			return shape{}, errWKTMixedDims
		}
		if len(coords)/ctype.Dimension() < 4 {
			return shape{}, errWKTMorePoints
		}
		if !isClosed(coords, ctype, patch) {
			return shape{}, errWKTUnclosedRings
		}
		s.rings = append(s.rings, coords)
		if r.peek() != "," {
			return s, nil
		}
		r.next()
	}
}

// readTriangleRing reads the single parenthesized ring of a triangle, which needs exactly four points and must be
// closed.
func (r *wktReader) readTriangleRing() (shape, error) {
	if err := r.expect("("); err != nil {
		return shape{}, err
	}
	coords, ctype, err := r.readPoints()
	if err != nil {
		return shape{}, err
	}
	if err = r.expect(")"); err != nil {
		return shape{}, err
	}
	if len(coords)/ctype.Dimension() != 4 {
		return shape{}, errWKTTrianglePts
	}
	if !isClosed(coords, ctype, true) {
		return shape{}, errWKTUnclosedRings
	}
	return shape{typeNum: triangleType, ctype: ctype, coords: coords}, nil
}

// readMembers reads the comma-separated members of a collection type with the given member reader. The collection
// takes the dimensions of its first member, and compound curves and curve polygons check their members as they are
// added.
func (r *wktReader) readMembers(typeNum int, readMember func() (shape, error)) (shape, error) {
	s := shape{typeNum: typeNum}
	for {
		member, err := readMember()
		if err != nil {
			return shape{}, err
		}
		if len(s.members) == 0 {
			s.ctype = member.ctype
		} else if (typeNum == compoundType || typeNum == curvePolyType) && member.ctype.Dimension() != s.ctype.Dimension() {
			return shape{}, errWKTMixedDims
		}
		if typeNum == compoundType && !continuesCompound(s, member) {
			return shape{}, errWKTIncontinuous
		}
		if typeNum == curvePolyType {
			if err = checkCurveRing(member); err != nil {
				return shape{}, err
			}
		}
		s.members = append(s.members, member)
		if r.peek() != "," {
			return s, nil
		}
		r.next()
	}
}

// continuesCompound returns whether a member may be appended to a compound curve: it must not be empty, and it must
// start where the curve so far ends.
func continuesCompound(compound shape, member shape) bool {
	if member.isEmpty() {
		return false
	}
	if len(compound.members) == 0 {
		return true
	}
	last := compound.members[len(compound.members)-1]
	dims := last.ctype.Dimension()
	lastX, lastY := last.coords[len(last.coords)-dims], last.coords[len(last.coords)-dims+1]
	return fpEquals(member.coords[0], lastX) && fpEquals(member.coords[1], lastY)
}

// checkCurveRing checks that a ring of a curve polygon has enough vertices and is closed. Linestring rings need four
// vertices, and the other curves need three.
func checkCurveRing(ring shape) error {
	needed := 3
	if ring.typeNum == lineType {
		needed = 4
	}
	if ring.numPoints() < needed {
		return errWKTMorePoints
	}
	if !isClosedCurve(ring) {
		return errWKTUnclosedRings
	}
	return nil
}

// readCompoundMember reads a member of a compound curve: an untagged linestring, or a tagged linestring or circular
// string.
func (r *wktReader) readCompoundMember() (shape, error) {
	if r.peek() == "(" || strings.EqualFold(r.peek(), "EMPTY") {
		return r.readUntagged(lineType)
	}
	member, err := r.readTagged()
	if err != nil {
		return shape{}, err
	}
	if member.typeNum != lineType && member.typeNum != circStringType {
		return shape{}, errWKTInvalid
	}
	return member, nil
}

// readCurveRing reads a ring of a curve polygon: an untagged linestring, or a tagged linestring, circular string, or
// compound curve.
func (r *wktReader) readCurveRing() (shape, error) {
	return r.readCurveMember()
}

// readCurveMember reads a member of a multicurve: an untagged linestring, or a tagged linestring, circular string,
// or compound curve.
func (r *wktReader) readCurveMember() (shape, error) {
	if r.peek() == "(" || strings.EqualFold(r.peek(), "EMPTY") {
		return r.readUntagged(lineType)
	}
	member, err := r.readTagged()
	if err != nil {
		return shape{}, err
	}
	if member.typeNum != lineType && member.typeNum != circStringType && member.typeNum != compoundType {
		return shape{}, errWKTInvalid
	}
	return member, nil
}

// readSurfaceMember reads a member of a multisurface: an untagged polygon, or a tagged polygon or curve polygon.
func (r *wktReader) readSurfaceMember() (shape, error) {
	if r.peek() == "(" || strings.EqualFold(r.peek(), "EMPTY") {
		return r.readUntagged(polygonType)
	}
	member, err := r.readTagged()
	if err != nil {
		return shape{}, err
	}
	if member.typeNum != polygonType && member.typeNum != curvePolyType {
		return shape{}, errWKTInvalid
	}
	return member, nil
}

// readPatch reads a patch of a polyhedral surface, which is an untagged polygon whose rings are checked for closure
// in three dimensions.
func (r *wktReader) readPatch() (shape, error) {
	if err := r.expect("("); err != nil {
		return shape{}, err
	}
	patch, err := r.readRings(true)
	if err != nil {
		return shape{}, err
	}
	return patch, r.expect(")")
}

// readUntaggedTriangle reads a member of a TIN, which is an untagged triangle.
func (r *wktReader) readUntaggedTriangle() (shape, error) {
	if err := r.expect("("); err != nil {
		return shape{}, err
	}
	triangle, err := r.readTriangleRing()
	if err != nil {
		return shape{}, err
	}
	return triangle, r.expect(")")
}

// readUntagged reads an untagged linestring or polygon, which is either EMPTY or a parenthesized body.
func (r *wktReader) readUntagged(typeNum int) (shape, error) {
	if strings.EqualFold(r.peek(), "EMPTY") {
		r.next()
		return shape{typeNum: typeNum}, nil
	}
	if err := r.expect("("); err != nil {
		return shape{}, err
	}
	var s shape
	var err error
	if typeNum == polygonType {
		s, err = r.readRings(false)
	} else {
		s, err = r.readCurveContents(typeNum)
	}
	if err != nil {
		return shape{}, err
	}
	return s, r.expect(")")
}

// readMultiPointMember reads a member of a multipoint, whose coordinates may appear with or without parentheses.
func (r *wktReader) readMultiPointMember() (shape, error) {
	if strings.EqualFold(r.peek(), "EMPTY") {
		r.next()
		return shape{typeNum: pointType}, nil
	}
	parenthesized := r.peek() == "("
	if parenthesized {
		r.next()
	}
	coords, ctype, err := r.readCoordinate()
	if err != nil {
		return shape{}, err
	}
	if parenthesized {
		if err = r.expect(")"); err != nil {
			return shape{}, err
		}
	}
	return shape{typeNum: pointType, ctype: ctype, coords: coords}, nil
}

// readPoints reads a comma-separated list of points, which must all have the same number of ordinates.
func (r *wktReader) readPoints() ([]float64, coordinatesType, error) {
	coords, ctype, err := r.readCoordinate()
	if err != nil {
		return nil, 0, err
	}
	for r.peek() == "," {
		r.next()
		point, pointCtype, err := r.readCoordinate()
		if err != nil {
			return nil, 0, err
		}
		if pointCtype != ctype {
			return nil, 0, errWKTMixedDims
		}
		coords = append(coords, point...)
	}
	return coords, ctype, nil
}

// readCoordinate reads the two to four ordinates of a single point. Three ordinates are read as XYZ, which a tag of
// "M" later relabels as XYM.
func (r *wktReader) readCoordinate() ([]float64, coordinatesType, error) {
	var ordinates []float64
	for {
		tok := r.peek()
		if tok == "," || tok == ")" || tok == "" {
			break
		}
		val, err := parseOrdinate(r.next())
		if err != nil {
			return nil, 0, err
		}
		ordinates = append(ordinates, val)
	}
	switch len(ordinates) {
	case 2:
		return ordinates, dimXY, nil
	case 3:
		return ordinates, dimXYZ, nil
	case 4:
		return ordinates, dimXYZM, nil
	default:
		return nil, 0, errWKTInvalid
	}
}

// parseOrdinate parses a single ordinate, which may be "NaN" but never infinite.
func parseOrdinate(tok string) (float64, error) {
	val, err := strconv.ParseFloat(tok, 64)
	if err != nil || math.IsInf(val, 0) {
		return 0, errWKTInvalid
	}
	if math.IsNaN(val) {
		return math.Float64frombits(quietNaNBits), nil
	}
	return val, nil
}

// expect consumes the next token, which must be the given punctuation.
func (r *wktReader) expect(tok string) error {
	if r.next() != tok {
		return errWKTInvalid
	}
	return nil
}

// peek returns the next token without consuming it.
func (r *wktReader) peek() string {
	pos := r.pos
	tok := r.next()
	r.pos = pos
	return tok
}

// next consumes and returns the next token: a word, a number, or a single punctuation character. It returns an
// empty string at the end of the input.
func (r *wktReader) next() string {
	for r.pos < len(r.input) && isWKTSpace(r.input[r.pos]) {
		r.pos++
	}
	if r.pos >= len(r.input) {
		return ""
	}
	start := r.pos
	if c := r.input[r.pos]; c == '(' || c == ')' || c == ',' {
		r.pos++
	} else if isWKTLetter(c) {
		for r.pos < len(r.input) && isWKTLetter(r.input[r.pos]) {
			r.pos++
		}
	} else {
		for r.pos < len(r.input) && !isWKTSpace(r.input[r.pos]) && strings.IndexByte("(),", r.input[r.pos]) < 0 {
			r.pos++
		}
	}
	return r.input[start:r.pos]
}

// isWKTSpace returns whether the character separates WKT tokens.
func isWKTSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// isWKTLetter returns whether the character can appear in a WKT tag.
func isWKTLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

// suffixDims returns the coordinates type that a "Z", "M", or "ZM" suffix declares, or XY for no suffix.
func suffixDims(suffix string) coordinatesType {
	switch suffix {
	case "Z":
		return dimXYZ
	case "M":
		return dimXYM
	case "ZM":
		return dimXYZM
	default:
		return dimXY
	}
}

// fpTolerance is the tolerance PostGIS uses when comparing ordinates for equality.
const fpTolerance = 1e-12

// fpEquals returns whether two ordinates are equal within PostGIS' tolerance.
func fpEquals(a float64, b float64) bool {
	return math.Abs(a-b) <= fpTolerance
}

// isClosed returns whether a sequence's first and last points are equal, also comparing Z when `checkZ` is true and
// the sequence has Z coordinates.
func isClosed(coords []float64, ctype coordinatesType, checkZ bool) bool {
	dims := ctype.Dimension()
	if len(coords) < dims {
		return false
	}
	last := coords[len(coords)-dims:]
	if coords[0] != last[0] || coords[1] != last[1] {
		return false
	}
	return !(checkZ && ctype.Is3D() && coords[2] != last[2])
}

// isClosedCurve returns whether a linestring, circular string, or compound curve ends where it starts, comparing Z
// when the curve has Z coordinates.
func isClosedCurve(curve shape) bool {
	if curve.typeNum != compoundType {
		return isClosed(curve.coords, curve.ctype, true)
	}
	if len(curve.members) == 0 {
		return false
	}
	first, last := curve.members[0], curve.members[len(curve.members)-1]
	dims := curve.ctype.Dimension()
	joined := append(append([]float64(nil), first.coords[:dims]...), last.coords[len(last.coords)-dims:]...)
	return isClosed(joined, curve.ctype, true)
}
