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

// measureLine returns a copy of a linestring with M coordinates that grow from `start` to `end` in proportion to the
// distance traveled, where `before` is the distance traveled before the linestring and `total` is the distance over
// all linestrings. Without any distance to travel, the M coordinates grow evenly from point to point. An empty
// linestring sends the notice that PostGIS sends.
func measureLine(ctx *sql.Context, line shape, ctype coordinatesType, start float64, end float64, before float64, total float64) shape {
	if line.isEmpty() {
		sendNotice(ctx, "NOTICE", "lwgeom_api.c [228] called with n=0 and npoints=0")
	}
	dims := line.ctype.Dimension()
	count := len(line.coords) / dims
	measured := shape{typeNum: lineType, ctype: ctype}
	traveled := before
	for i := 0; i < count; i++ {
		point := line.coords[i*dims : (i+1)*dims]
		if i > 0 {
			traveled += math.Hypot(point[0]-line.coords[(i-1)*dims], point[1]-line.coords[(i-1)*dims+1])
		}
		m := start
		if total > 0 {
			m = start + float64((end-start)*(traveled/total))
		} else if count > 1 {
			m = start + float64((end-start)*(float64(i)/float64(count-1)))
		}
		converted := convertPoint(point, line.ctype, ctype, 0, 0)
		converted[len(converted)-1] = m
		measured.coords = append(measured.coords, converted...)
	}
	return measured
}

// stAddMeasure implements ST_AddMeasure, which gives linestrings M coordinates that grow along their length. Each
// member of a multilinestring covers the part of the range that matches its share of the total length.
func stAddMeasure(ctx *sql.Context, args ...any) (any, error) {
	g, start, end := args[0].(Geometry), args[1].(float64), args[2].(float64)
	s := g.Shape
	ctype := s.ctype | dimXYM
	switch s.typeNum {
	case lineType:
		return g.withShape(measureLine(ctx, s, ctype, start, end, 0, curveLength(s))), nil
	case multiLineType:
		measured := shape{typeNum: multiLineType, ctype: ctype}
		total := curveLength(s)
		before := 0.0
		for _, member := range s.members {
			measured.members = append(measured.members, measureLine(ctx, member, ctype, start, end, before, total))
			before += curveLength(member)
		}
		return g.withShape(measured), nil
	default:
		return nil, pgerror.New(pgcode.Internal, "Only LINESTRING and MULTILINESTRING are supported")
	}
}

// pointsAlong returns the points of a shape whose ordinate at `index` equals the value, moved sideways by the offset
// to the left of the direction of travel. A linestring yields a point where each segment reaches the value, or the
// middle of a segment that holds the value throughout, skipping a point that repeats the one before it.
func pointsAlong(s shape, index int, value float64, offset float64) ([]shape, error) {
	dims := s.ctype.Dimension()
	switch s.typeNum {
	case pointType:
		if len(s.coords) > 0 && s.coords[index] == value {
			return []shape{s}, nil
		}
		return nil, nil
	case lineType:
		var points []shape
		for i := dims; i+dims <= len(s.coords); i += dims {
			a, b := s.coords[i-dims:i], s.coords[i:i+dims]
			va, vb := a[index], b[index]
			var point []float64
			if va == value && vb == value {
				point = interpolatePoint(a, b, 0.5)
			} else if (va <= value && value <= vb) || (vb <= value && value <= va) {
				point = interpolatePoint(a, b, (value-va)/(vb-va))
			} else {
				continue
			}
			point[index] = value
			if offset != 0 {
				dx, dy := b[0]-a[0], b[1]-a[1]
				length := math.Hypot(dx, dy)
				point[0] -= float64(dy / length * offset)
				point[1] += float64(dx / length * offset)
			}
			if len(points) > 0 && equalOrdinates(points[len(points)-1].coords, point) {
				continue
			}
			points = append(points, shape{typeNum: pointType, ctype: s.ctype, coords: point})
		}
		return points, nil
	case multiPointType, multiLineType:
		var points []shape
		for _, member := range s.members {
			memberPoints, err := pointsAlong(member, index, value, offset)
			if err != nil {
				return nil, err
			}
			points = append(points, memberPoints...)
		}
		return points, nil
	default:
		return nil, pgerror.Newf(pgcode.Internal, "Only linear geometries are supported, %s provided.", typeNameForNumber(s.typeNum))
	}
}

// stLocateAlong implements ST_LocateAlong, which returns the multipoint of the locations with the given measure.
func stLocateAlong(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	if !g.Shape.hasM() {
		return nil, pgerror.New(pgcode.Internal, "Input geometry does not have a measure dimension")
	}
	points, err := pointsAlong(g.Shape, g.Shape.ctype.Dimension()-1, args[1].(float64), args[2].(float64))
	if err != nil {
		return nil, err
	}
	return g.withShape(shape{typeNum: multiPointType, ctype: g.Shape.ctype, members: points}), nil
}

// clipSequence returns the pieces of a sequence where the ordinate at `index` lies within the range, adding a point
// wherever the sequence crosses into or out of the range, interpolated back from the end of the segment. A piece that
// touches the range at one location holds a single point.
func clipSequence(seq []float64, ctype coordinatesType, index int, from float64, to float64) [][]float64 {
	dims := ctype.Dimension()
	inRange := func(v float64) bool {
		return v >= from && v <= to
	}
	var pieces [][]float64
	var current []float64
	for i := 0; i+dims <= len(seq); i += dims {
		b := seq[i : i+dims]
		if i > 0 {
			a := seq[i-dims : i]
			va, vb := a[index], b[index]
			bounds := []float64{from, to}
			if va > vb {
				bounds = []float64{to, from}
			}
			for _, bound := range bounds {
				if (va < bound && bound < vb) || (vb < bound && bound < va) {
					crossing := interpolatePoint(b, a, (vb-bound)/(vb-va))
					crossing[index] = bound
					if current == nil {
						current = crossing
					} else {
						pieces = append(pieces, appendDistinct(current, crossing, ctype))
						current = nil
					}
				}
			}
		}
		if inRange(b[index]) {
			current = appendDistinct(current, b, ctype)
		} else if current != nil {
			pieces = append(pieces, current)
			current = nil
		}
	}
	if current != nil {
		pieces = append(pieces, current)
	}
	return pieces
}

// clipShape returns the parts of a shape whose ordinate at `index` lies within the range, as points and linestrings.
// The function name is the one that PostGIS reports for geometries it does not accept.
func clipShape(s shape, index int, from float64, to float64) ([]shape, error) {
	switch s.typeNum {
	case pointType:
		if len(s.coords) > 0 && s.coords[index] >= from && s.coords[index] <= to {
			return []shape{s}, nil
		}
		return nil, nil
	case lineType:
		var parts []shape
		for _, piece := range clipSequence(s.coords, s.ctype, index, from, to) {
			typeNum := lineType
			if len(piece) == s.ctype.Dimension() {
				typeNum = pointType
			}
			parts = append(parts, shape{typeNum: typeNum, ctype: s.ctype, coords: piece})
		}
		return parts, nil
	case multiPointType, multiLineType, collectionType:
		var parts []shape
		for _, member := range s.members {
			memberParts, err := clipShape(member, index, from, to)
			if err != nil {
				return nil, err
			}
			parts = append(parts, memberParts...)
		}
		return parts, nil
	case circStringType, compoundType, curvePolyType, multiCurveType, multiSurfaceType:
		return nil, pgerror.Newf(pgcode.Internal, "This function does not accept %s geometries.", typeNameForNumber(s.typeNum))
	default:
		//TODO: clip polygons, triangles, and surfaces to the range
		return nil, pgerror.Newf(pgcode.FeatureNotSupported, "clipping %s geometries to a range is not yet supported", typeNameForNumber(s.typeNum))
	}
}

// clipToRange returns the parts of a geometry whose ordinate at `index` lies within the range, in either order. Points
// form a multipoint, and linestrings form a multilinestring unless a part is a single point, which makes a geometry
// collection, as does an input collection.
func clipToRange(g Geometry, index int, from float64, to float64) (any, error) {
	if from > to {
		from, to = to, from
	}
	parts, err := clipShape(g.Shape, index, from, to)
	if err != nil {
		return nil, err
	}
	clipped := shape{typeNum: multiLineType, ctype: g.Shape.ctype, members: parts}
	switch g.Shape.typeNum {
	case pointType, multiPointType:
		clipped.typeNum = multiPointType
	case collectionType:
		clipped.typeNum = collectionType
	default:
		for _, part := range parts {
			if part.typeNum == pointType {
				clipped.typeNum = collectionType
			}
		}
	}
	return g.withShape(clipped), nil
}

// stLocateBetween implements ST_LocateBetween, which returns the parts of a geometry whose measures lie within the
// range. A range of a single measure returns the same points as ST_LocateAlong. With an offset, clipped points are
// dropped with a notice, as in PostGIS.
func stLocateBetween(ctx *sql.Context, args ...any) (any, error) {
	g, from, to, offset := args[0].(Geometry), args[1].(float64), args[2].(float64), args[3].(float64)
	if !g.Shape.hasM() {
		return nil, pgerror.New(pgcode.Internal, "This function only accepts geometries that have an M dimension.")
	}
	if from == to {
		return stLocateAlong(ctx, g, from, offset)
	}
	if offset == 0 {
		return clipToRange(g, g.Shape.ctype.Dimension()-1, from, to)
	}
	if from > to {
		from, to = to, from
	}
	parts, err := clipShape(g.Shape, g.Shape.ctype.Dimension()-1, from, to)
	if err != nil {
		return nil, err
	}
	for _, part := range parts {
		if part.typeNum != pointType {
			//TODO: offset the clipped linestrings, which needs the offset curves of GEOS
			return nil, pgerror.New(pgcode.FeatureNotSupported, "an offset is not yet supported by ST_LocateBetween for linestrings")
		}
		sendNotice(ctx, "NOTICE", "lwgeom_clip_to_ordinate_range cannot offset a clipped point")
	}
	return g.withShape(shape{typeNum: multiLineType, ctype: dimXY}), nil
}

// stLocateBetweenElevations implements ST_LocateBetweenElevations, which returns the parts of a geometry whose Z
// coordinates lie within the range.
func stLocateBetweenElevations(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	if !g.Shape.hasZ() {
		return nil, pgerror.New(pgcode.Internal, "This function only accepts geometries with Z dimensions.")
	}
	return clipToRange(g, 2, args[1].(float64), args[2].(float64))
}

// stInterpolatePoint implements ST_InterpolatePoint, which returns the measure of a linestring where it comes closest
// to a point.
func stInterpolatePoint(ctx *sql.Context, args ...any) (any, error) {
	g, p := args[0].(Geometry), args[1].(Geometry)
	if err := checkSameSRID("ST_InterpolatePoint", g, p); err != nil {
		return nil, err
	}
	line, point := g.Shape, p.Shape
	if line.typeNum != lineType {
		return nil, pgerror.New(pgcode.Internal, "ST_InterpolatePoint: 1st argument isn't a line")
	}
	if point.typeNum != pointType {
		return nil, pgerror.New(pgcode.Internal, "ST_InterpolatePoint: 2nd argument isn't a point")
	}
	if !line.hasM() {
		return nil, pgerror.New(pgcode.Internal, "ST_InterpolatePoint only accepts geometries that have an M dimension")
	}
	if line.isEmpty() || point.isEmpty() {
		return nil, pgerror.New(pgcode.Internal, "Input geometry is empty")
	}
	dims := line.ctype.Dimension()
	closest, closestDistance := dims, math.Inf(1)
	for i := dims; i+dims <= len(line.coords); i += dims {
		if d := segmentDistanceSquared(point.coords, line.coords[i-dims:], line.coords[i:]); d < closestDistance {
			closest, closestDistance = i, d
		}
	}
	a, b := line.coords[closest-dims:closest], line.coords[closest:closest+dims]
	dx, dy := b[0]-a[0], b[1]-a[1]
	r := 0.0
	if dx != 0 || dy != 0 {
		r = (float64((point.coords[0]-a[0])*dx) + float64((point.coords[1]-a[1])*dy)) / (float64(dx*dx) + float64(dy*dy))
		r = math.Max(0, math.Min(1, r))
	}
	return a[dims-1] + float64((b[dims-1]-a[dims-1])*r), nil
}

// stLineFromMultiPoint implements ST_LineFromMultiPoint, which joins the points of a multipoint into a linestring.
func stLineFromMultiPoint(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	if g.Shape.typeNum != multiPointType {
		return nil, pgerror.New(pgcode.Internal, "makeline: input must be a multipoint")
	}
	line := shape{typeNum: lineType, ctype: g.Shape.ctype}
	for _, member := range g.Shape.members {
		if member.isEmpty() {
			return nil, pgerror.New(pgcode.Internal, "lwgeom_api.c [138] called with n=0 and npoints=0")
		}
		line.coords = append(line.coords, member.coords...)
	}
	return g.withShape(line), nil
}

// measuredRoutines returns the routines that work with the measures and elevations of linestrings.
func measuredRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "st_addmeasure", Symbol: "ST_AddMeasure", Parameters: params("geometry", "float8", "float8"), Returns: "geometry", Strict: true, Impl: stAddMeasure},
		{Name: "st_interpolatepoint", Symbol: "ST_InterpolatePoint", Parameters: []extdef.Parameter{{Name: "line", Type: "geometry"}, {Name: "point", Type: "geometry"}}, Returns: "float8", Strict: true, Impl: stInterpolatePoint},
		{Name: "st_linefrommultipoint", Symbol: "LWGEOM_line_from_mpoint", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stLineFromMultiPoint},
		{Name: "st_locatealong", Symbol: "ST_LocateAlong", Parameters: []extdef.Parameter{{Name: "geometry", Type: "geometry"}, {Name: "measure", Type: "float8"}, {Name: "leftrightoffset", Type: "float8", Default: `0.0`}}, Returns: "geometry", Strict: true, Impl: stLocateAlong},
		{Name: "st_locatebetween", Symbol: "ST_LocateBetween", Parameters: []extdef.Parameter{{Name: "geometry", Type: "geometry"}, {Name: "frommeasure", Type: "float8"}, {Name: "tomeasure", Type: "float8"}, {Name: "leftrightoffset", Type: "float8", Default: `0.0`}}, Returns: "geometry", Strict: true, Impl: stLocateBetween},
		{Name: "st_locatebetweenelevations", Symbol: "ST_LocateBetweenElevations", Parameters: []extdef.Parameter{{Name: "geometry", Type: "geometry"}, {Name: "fromelevation", Type: "float8"}, {Name: "toelevation", Type: "float8"}}, Returns: "geometry", Strict: true, Impl: stLocateBetweenElevations},
	}
}
