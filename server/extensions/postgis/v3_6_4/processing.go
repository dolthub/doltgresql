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

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// emptyOverlay decides the result of an overlay when either input is empty, returning false when neither is.
type emptyOverlay func(a Geometry, b Geometry) (Geometry, bool)

// Overlays with empty inputs return one of the inputs unchanged: the union and symmetric difference return the
// other input, the intersection returns the empty input, and the difference returns the first input.
var (
	unionWithEmpty = func(a Geometry, b Geometry) (Geometry, bool) {
		if a.Shape.isEmpty() {
			return b, true
		}
		return a, b.Shape.isEmpty()
	}
	intersectionWithEmpty = func(a Geometry, b Geometry) (Geometry, bool) {
		if b.Shape.isEmpty() {
			return b, true
		}
		return a, a.Shape.isEmpty()
	}
	differenceWithEmpty = func(a Geometry, b Geometry) (Geometry, bool) {
		return a, a.Shape.isEmpty() || b.Shape.isEmpty()
	}
)

// overlayImpl returns an implementation of a binary overlay operation, named as in the errors that PostGIS reports.
// The result keeps the first geometry's SRID. A positive grid size snaps the inputs and the result to the grid.
func overlayImpl(opName string, withEmpty emptyOverlay, op func(g *spatial.GEOS, a spatial.Geometry, b spatial.Geometry, gridSize float64) (spatial.Geometry, error)) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].(Geometry), args[1].(Geometry)
		if a.SRID != b.SRID {
			return nil, pgerror.Newf(pgcode.Internal, "lwgeom_%s_prec: Operation on mixed SRID geometries (%d != %d)", opName, a.SRID, b.SRID)
		}
		geosA, err := a.Shape.toGEOS()
		if err != nil {
			return nil, err
		}
		geosB, err := b.Shape.toGEOS()
		if err != nil {
			return nil, err
		}
		if result, ok := withEmpty(a, b); ok {
			return result, nil
		}
		gridSize := -1.0
		if len(args) > 2 {
			gridSize = args[2].(float64)
		}
		g := spatial.NewGEOS()
		defer g.Close()
		result, err := op(g, geosA, geosB, gridSize)
		if err != nil {
			geosNames := map[string]string{"union": "Union", "intersection": "Intersection", "difference": "Difference", "symdifference": "SymDifference"}
			return nil, geosError("GEOS"+geosNames[opName], err)
		}
		return fromGEOS(result, a.SRID, a.Shape.hasZ() || b.Shape.hasZ()), nil
	}
}

// unaryImpl returns an implementation of a GEOS operation on a single geometry, whose result keeps the input's SRID
// and Z coordinates.
func unaryImpl(function string, op func(g *spatial.GEOS, a spatial.Geometry) (spatial.Geometry, error)) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a := args[0].(Geometry)
		geosA, err := a.Shape.toGEOS()
		if err != nil {
			return nil, err
		}
		g := spatial.NewGEOS()
		defer g.Close()
		result, err := op(g, geosA)
		if err != nil {
			return nil, geosError(function, err)
		}
		return fromGEOS(result, a.SRID, a.Shape.hasZ()), nil
	}
}

// stUnaryUnion implements ST_UnaryUnion, which snaps the input and the result to a positive grid size.
func stUnaryUnion(ctx *sql.Context, args ...any) (any, error) {
	gridSize := args[1].(float64)
	return unaryImpl("GEOSUnaryUnion", func(g *spatial.GEOS, a spatial.Geometry) (spatial.Geometry, error) {
		return g.UnaryUnion(a, gridSize)
	})(ctx, args[0])
}

// stConvexHull implements ST_ConvexHull.
func stConvexHull(ctx *sql.Context, args ...any) (any, error) {
	return unaryImpl("GEOSConvexHull", (*spatial.GEOS).ConvexHull)(ctx, args...)
}

// stCentroid implements ST_Centroid.
func stCentroid(ctx *sql.Context, args ...any) (any, error) {
	return unaryImpl("GEOSGetCentroid", (*spatial.GEOS).Centroid)(ctx, args...)
}

// stNormalize implements ST_Normalize. As with GEOS, curves are linearized, triangles become polygons, and M
// coordinates are dropped.
func stNormalize(ctx *sql.Context, args ...any) (any, error) {
	return unaryImpl("GEOSNormalize", (*spatial.GEOS).Normalize)(ctx, args...)
}

// stPointOnSurface implements ST_PointOnSurface.
func stPointOnSurface(ctx *sql.Context, args ...any) (any, error) {
	return unaryImpl("GEOSPointOnSurface", (*spatial.GEOS).PointOnSurface)(ctx, args...)
}

// validity returns the reason that GEOS gives for an invalid geometry under the given flags, along with the location
// of the problem, or an empty reason for a valid geometry.
func validity(a Geometry, flags int32) (reason string, location Geometry, err error) {
	if a.Shape.isEmpty() {
		return "", Geometry{}, nil
	}
	geosA, err := a.Shape.toGEOS()
	if err != nil {
		return "", Geometry{}, err
	}
	g := spatial.NewGEOS()
	defer g.Close()
	valid, reason, point, err := g.IsValidDetail(geosA, int(flags))
	if err != nil || valid {
		return "", Geometry{}, err
	}
	return reason, fromGEOS(point, 0, true), nil
}

// formatLocation formats the coordinates of a validity location with the given number of significant digits.
func formatLocation(location []float64, precision int) string {
	return strings.ToLower(formatG(location[0], precision) + " " + formatG(location[1], precision))
}

// stIsValid implements ST_IsValid, which sends the reason for an invalid geometry as a notice unless it is given
// flags.
func stIsValid(ctx *sql.Context, args ...any) (any, error) {
	reason, location, err := validity(args[0].(Geometry), validityFlags(args))
	if err != nil {
		return nil, err
	}
	if reason != "" && len(args) == 1 {
		sendNotice(ctx, "NOTICE", reason+" at or near point "+formatLocation(location.Shape.coords, 17))
	}
	return reason == "", nil
}

// stIsValidReason implements ST_IsValidReason, which includes the location of the problem unless it is given flags.
func stIsValidReason(ctx *sql.Context, args ...any) (any, error) {
	reason, location, err := validity(args[0].(Geometry), validityFlags(args))
	if err != nil {
		return nil, err
	}
	if reason == "" {
		return "Valid Geometry", nil
	}
	if len(args) == 1 {
		reason += "[" + formatLocation(location.Shape.coords, 15) + "]"
	}
	return reason, nil
}

// stIsValidDetail implements ST_IsValidDetail.
func stIsValidDetail(ctx *sql.Context, args ...any) (any, error) {
	reason, location, err := validity(args[0].(Geometry), validityFlags(args))
	if err != nil {
		return nil, err
	}
	if reason == "" {
		return compositeValue(ctx, "valid_detail", true, nil, nil)
	}
	return compositeValue(ctx, "valid_detail", false, reason, location)
}

// validityFlags returns the optional flags argument of the validity functions, or zero when it is missing.
func validityFlags(args []any) int32 {
	if len(args) > 1 {
		return args[1].(int32)
	}
	return 0
}

// bufferParams holds the options of ST_Buffer.
type bufferParams struct {
	quadSegs   int
	endCap     string
	join       string
	mitreLimit float64
	side       string
}

// parseBufferParams parses the space-separated "key=value" options of ST_Buffer.
func parseBufferParams(text string) (bufferParams, error) {
	params := bufferParams{quadSegs: 8, endCap: "round", join: "round", mitreLimit: 5, side: "both"}
	for _, param := range strings.Fields(text) {
		key, val, ok := strings.Cut(param, "=")
		if !ok || val == "" {
			return params, pgerror.Newf(pgcode.Internal, "Missing value for buffer parameter %s", key)
		}
		switch key {
		case "endcap":
			if val == "butt" {
				val = "flat"
			}
			if val != "round" && val != "flat" && val != "square" {
				return params, pgerror.Newf(pgcode.Internal, "Invalid buffer end cap style: %s (accept: 'round', 'flat', 'butt' or 'square')", val)
			}
			params.endCap = val
		case "join":
			if val == "miter" {
				val = "mitre"
			}
			if val != "round" && val != "mitre" && val != "bevel" {
				return params, pgerror.Newf(pgcode.Internal, "Invalid buffer end cap style: %s (accept: 'round', 'mitre', 'miter'  or 'bevel')", val)
			}
			params.join = val
		case "mitre_limit", "miter_limit":
			params.mitreLimit = parseCFloat(val)
		case "quad_segs":
			params.quadSegs = parseCInt(val)
		case "side":
			if val != "both" && val != "left" && val != "right" {
				return params, pgerror.Newf(pgcode.Internal, "Invalid side parameter: %s (accept: 'right', 'left', 'both')", val)
			}
			params.side = val
		default:
			return params, pgerror.Newf(pgcode.Internal, "Invalid buffer parameter: %s (accept: 'endcap', 'join', 'mitre_limit', 'miter_limit', 'quad_segs' and 'side')", key)
		}
	}
	return params, nil
}

// parseCFloat reads the longest numeric prefix of the text the way C's atof does, returning zero when there is none.
func parseCFloat(text string) float64 {
	val, _ := parseStrtod(scanFloat.FindString(strings.TrimLeft(text, " \t\n\v\f\r")))
	return val
}

// parseCInt reads the integer prefix of the text the way C's atoi does, returning zero when there is none.
func parseCInt(text string) int {
	text = strings.TrimLeft(text, " \t\n\v\f\r")
	end := 0
	for end < len(text) && ((text[end] >= '0' && text[end] <= '9') || (end == 0 && (text[end] == '-' || text[end] == '+'))) {
		end++
	}
	val, _ := strconv.Atoi(text[:end])
	return val
}

// stBuffer implements ST_Buffer for geometry, with either an options string or a number of segments per quarter
// circle.
func stBuffer(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	distance := args[1].(float64)
	params := bufferParams{quadSegs: 8, endCap: "round", join: "round", mitreLimit: 5, side: "both"}
	if len(args) > 2 {
		if quadSegs, ok := args[2].(int32); ok {
			params.quadSegs = int(quadSegs)
		} else {
			var err error
			if params, err = parseBufferParams(args[2].(string)); err != nil {
				return nil, err
			}
		}
	}
	if g.Shape.isEmpty() {
		return Geometry{SRID: g.SRID, Shape: shape{typeNum: polygonType}}, nil
	}
	for _, seq := range g.Shape.sequences() {
		for _, val := range seq {
			if math.IsNaN(val) || math.IsInf(val, 0) {
				return nil, pgerror.New(pgcode.Internal, "Geometry contains invalid coordinates")
			}
		}
	}
	geosG, err := g.Shape.toGEOS()
	if err != nil {
		return nil, err
	}
	geosParams := spatial.BufferParams{QuadSegs: params.quadSegs, EndCap: spatial.BufferRound, Join: spatial.BufferRound, MitreLimit: params.mitreLimit}
	switch params.endCap {
	case "flat":
		geosParams.EndCap = spatial.BufferFlat
	case "square":
		geosParams.EndCap = spatial.BufferSquare
	}
	switch params.join {
	case "mitre":
		geosParams.Join = spatial.BufferMitre
	case "bevel":
		geosParams.Join = spatial.BufferBevel
	}
	if params.side != "both" {
		geosParams.SingleSided = true
		if params.side == "right" {
			distance = -distance
		}
	}
	geos := spatial.NewGEOS()
	defer geos.Close()
	result, err := geos.Buffer(geosG, distance, geosParams)
	if err != nil {
		return nil, geosError("GEOSBuffer", err)
	}
	return fromGEOS(result, g.SRID, g.Shape.hasZ()), nil
}

// processingRoutines returns the routines that derive new geometries through GEOS.
func processingRoutines() []extdef.Routine {
	geomParams := []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}
	withGridSize := func(defaultValue string) []extdef.Parameter {
		return append(append([]extdef.Parameter(nil), geomParams...), extdef.Parameter{Name: "gridsize", Type: "float8", Default: defaultValue})
	}
	unionImpl := overlayImpl("union", unionWithEmpty, (*spatial.GEOS).Union)
	intersectionImpl := overlayImpl("intersection", intersectionWithEmpty, (*spatial.GEOS).Intersection)
	bufferParamsList := []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "radius", Type: "float8"}, {Name: "options", Type: "text", Default: `''::text`}}
	return []extdef.Routine{
		{Name: "st_normalize", Symbol: "ST_Normalize", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}}, Returns: "geometry", Strict: true, Impl: stNormalize},
		{Name: "st_buffer", Symbol: "buffer", Parameters: bufferParamsList, Returns: "geometry", Strict: true, Impl: stBuffer},
		{Name: "st_buffer", Symbol: "st_buffer_geometry_float8_int4", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "radius", Type: "float8"}, {Name: "quadsegs", Type: "int4"}}, Returns: "geometry", Strict: true, Impl: stBuffer},
		{Name: "st_buffer", Symbol: "st_buffer_text_float8", Parameters: params("text", "float8"), Returns: "geometry", Strict: true, Impl: textArgsAsGeometry(1, stBuffer)},
		{Name: "st_buffer", Symbol: "st_buffer_text_float8_int4", Parameters: params("text", "float8", "int4"), Returns: "geometry", Strict: true, Impl: textArgsAsGeometry(1, stBuffer)},
		{Name: "st_buffer", Symbol: "st_buffer_text_float8_text", Parameters: params("text", "float8", "text"), Returns: "geometry", Strict: true, Impl: textArgsAsGeometry(1, stBuffer)},
		{Name: "st_centroid", Symbol: "centroid", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stCentroid},
		{Name: "st_centroid", Symbol: "st_centroid_text", Parameters: params("text"), Returns: "geometry", Strict: true, Impl: textArgsAsGeometry(1, stCentroid)},
		{Name: "st_convexhull", Symbol: "convexhull", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stConvexHull},
		{Name: "st_difference", Symbol: "ST_Difference", Parameters: withGridSize(`'-1.0'::numeric`), Returns: "geometry", Strict: true, Impl: overlayImpl("difference", differenceWithEmpty, (*spatial.GEOS).Difference)},
		{Name: "st_intersection", Symbol: "ST_Intersection", Parameters: withGridSize(`'-1'::integer`), Returns: "geometry", Strict: true, Impl: intersectionImpl},
		{Name: "st_intersection", Symbol: "st_intersection_text_text", Parameters: params("text", "text"), Returns: "geometry", Strict: true, Impl: textArgsAsGeometry(2, intersectionImpl)},
		{Name: "st_isvalid", Symbol: "isvalid", Parameters: params("geometry"), Returns: "bool", Strict: true, Impl: stIsValid},
		{Name: "st_isvalid", Symbol: "st_isvalid_geometry_int4", Parameters: params("geometry", "int4"), Returns: "bool", Strict: true, Impl: stIsValid},
		{Name: "st_isvalidreason", Symbol: "isvalidreason", Parameters: params("geometry"), Returns: "text", Strict: true, Impl: stIsValidReason},
		{Name: "st_isvalidreason", Symbol: "st_isvalidreason_geometry_int4", Parameters: params("geometry", "int4"), Returns: "text", Strict: true, Impl: stIsValidReason},
		{Name: "st_isvaliddetail", Symbol: "isvaliddetail", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "flags", Type: "int4", Default: `0`}}, Returns: "valid_detail", Strict: true, Impl: stIsValidDetail},
		{Name: "st_pointonsurface", Symbol: "pointonsurface", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stPointOnSurface},
		{Name: "st_symdifference", Symbol: "ST_SymDifference", Parameters: withGridSize(`'-1.0'::numeric`), Returns: "geometry", Strict: true, Impl: overlayImpl("symdifference", unionWithEmpty, (*spatial.GEOS).SymDifference)},
		{Name: "st_symmetricdifference", Symbol: "st_symmetricdifference_geometry_geometry", Parameters: geomParams, Returns: "geometry", Strict: true, Impl: overlayImpl("symdifference", unionWithEmpty, (*spatial.GEOS).SymDifference)},
		{Name: "st_unaryunion", Symbol: "ST_UnaryUnion", Parameters: []extdef.Parameter{{Type: "geometry"}, {Name: "gridsize", Type: "float8", Default: `'-1.0'::numeric`}}, Returns: "geometry", Strict: true, Impl: stUnaryUnion},
		{Name: "st_union", Symbol: "ST_Union", Parameters: geomParams, Returns: "geometry", Strict: true, Impl: unionImpl},
		{Name: "st_union", Symbol: "ST_Union_gridsize", Parameters: append(append([]extdef.Parameter(nil), geomParams...), extdef.Parameter{Name: "gridsize", Type: "float8"}), Returns: "geometry", Strict: true, Impl: unionImpl},
	}
}
