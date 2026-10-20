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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// stGeomFromText implements ST_GeomFromText, which warns when given EWKT. An SRID argument replaces the parsed SRID.
func stGeomFromText(ctx *sql.Context, args ...any) (any, error) {
	g, err := parseEWKT(args[0].(string))
	if err != nil {
		return nil, err
	}
	if g.SRID != sridUnknown {
		sendNotice(ctx, "WARNING", "OGC WKT expected, EWKT provided - use GeomFromEWKT() for this")
	}
	if len(args) > 1 {
		g.SRID = args[1].(int32)
	}
	g.SRID = clampSRID(ctx, g.SRID)
	return g, nil
}

// stMakePoint implements ST_MakePoint and the two-argument ST_Point, whose dimensions follow the number of arguments.
func stMakePoint(ctx *sql.Context, args ...any) (any, error) {
	ordinates := make([]float64, len(args))
	for i, arg := range args {
		ordinates[i] = arg.(float64)
	}
	ctype := dimXY
	switch len(ordinates) {
	case 3:
		ctype = dimXYZ
	case 4:
		ctype = dimXYZM
	}
	return Geometry{Shape: shape{typeNum: pointType, ctype: ctype, coords: ordinates}}, nil
}

// stPoint implements the three-argument ST_Point, which takes an SRID after the coordinates.
func stPoint(ctx *sql.Context, args ...any) (any, error) {
	point := shape{typeNum: pointType, coords: []float64{args[0].(float64), args[1].(float64)}}
	return Geometry{SRID: clampSRID(ctx, args[2].(int32)), Shape: point}, nil
}

// stGeomFromEWKT implements ST_GeomFromEWKT.
func stGeomFromEWKT(ctx *sql.Context, args ...any) (any, error) {
	g, err := parseEWKT(args[0].(string))
	if err != nil {
		return nil, err
	}
	g.SRID = clampSRID(ctx, g.SRID)
	return g, nil
}

// stGeomFromWKB implements ST_GeomFromWKB, which warns when given extended WKB with an SRID. An SRID argument
// replaces the decoded SRID.
func stGeomFromWKB(ctx *sql.Context, args ...any) (any, error) {
	g, err := decodeWKB(ctx, args[0].([]byte))
	if err != nil {
		return nil, err
	}
	if g.SRID != sridUnknown {
		sendNotice(ctx, "WARNING", "OGC WKB expected, EWKB provided - use GeometryFromEWKB() for this")
	}
	if len(args) > 1 {
		g.SRID = clampSRID(ctx, args[1].(int32))
	}
	return g, nil
}

// stGeogFromText implements ST_GeogFromText and ST_GeographyFromText.
func stGeogFromText(ctx *sql.Context, args ...any) (any, error) {
	g, err := stGeomFromEWKT(ctx, args...)
	if err != nil {
		return nil, err
	}
	return asGeography(ctx, g.(Geometry))
}

// collectionTypeFor returns the type of the collection that holds members of a single type: the matching
// multi-geometry when the type has one, and a geometry collection otherwise.
func collectionTypeFor(typeNum int) int {
	if multi := multiTypeNumber(typeNum); multi != typeNum {
		return multi
	}
	return collectionType
}

// stCollect implements the two-argument ST_Collect, which returns one geometry when the other is null. Two points,
// linestrings, polygons, circular strings, or triangles of the same type become the matching multi-geometry, and any
// other pair becomes a geometry collection.
func stCollect(ctx *sql.Context, args ...any) (any, error) {
	if args[0] == nil {
		return args[1], nil
	}
	if args[1] == nil {
		return args[0], nil
	}
	a, b := args[0].(Geometry), args[1].(Geometry)
	if err := checkSameSRID("LWGEOM_collect", a, b); err != nil {
		return nil, err
	}
	if a.Shape.ctype != b.Shape.ctype {
		return nil, pgerror.New(pgcode.Internal, "Cannot ST_Collect geometries with differing dimensionality.")
	}
	typeNum := collectionType
	if a.Shape.typeNum == b.Shape.typeNum && a.Shape.typeNum != compoundType && a.Shape.typeNum != curvePolyType {
		typeNum = collectionTypeFor(a.Shape.typeNum)
	}
	return a.withShape(shape{typeNum: typeNum, ctype: a.Shape.ctype, members: []shape{a.Shape, b.Shape}}), nil
}

// stCollectArray implements ST_Collect over an array, which skips nulls and returns null when nothing remains.
// Members of a single type become the matching multi-geometry, and mixed members become a geometry collection.
func stCollectArray(ctx *sql.Context, args ...any) (any, error) {
	var collected Geometry
	var typeNum int
	for _, elem := range args[0].([]any) {
		if elem == nil {
			continue
		}
		g := elem.(Geometry)
		if len(collected.Shape.members) == 0 {
			collected = Geometry{SRID: g.SRID, Shape: shape{ctype: g.Shape.ctype}}
			typeNum = g.Shape.typeNum
		} else if g.SRID != collected.SRID {
			return nil, pgerror.Newf(pgcode.Internal, "LWGEOM_collect_garray: Operation on mixed SRID geometries %s %d != %d",
				typeNameForNumber(g.Shape.typeNum), g.SRID, collected.SRID)
		} else if g.Shape.ctype != collected.Shape.ctype {
			return nil, pgerror.Newf(pgcode.Internal, "lwcollection_construct: mixed dimension geometries: %d/%d",
				zmFlag(collected.Shape.ctype), zmFlag(g.Shape.ctype))
		} else if g.Shape.typeNum != typeNum {
			typeNum = collectionType
		}
		collected.Shape.members = append(collected.Shape.members, g.Shape)
	}
	if len(collected.Shape.members) == 0 {
		return nil, nil
	}
	collected.Shape.typeNum = collectionTypeFor(typeNum)
	return collected, nil
}

// makeLine joins points and linestrings into one linestring that has every dimension of its parts, where a missing
// ordinate of a point is zero. Empty parts are skipped, and a linestring that starts where the line so far ends in two
// dimensions skips its first point.
func makeLine(parts []shape) (shape, error) {
	ctype := dimXY
	for _, part := range parts {
		ctype |= part.ctype
	}
	dims := ctype.Dimension()
	line := shape{typeNum: lineType, ctype: ctype}
	for _, part := range parts {
		if part.typeNum == pointType {
			if len(part.coords) > 0 {
				line.coords = append(line.coords, convertPoint(part.coords, part.ctype, ctype, 0, 0)...)
			}
			continue
		}
		if len(part.coords) == 0 {
			continue
		}
		if part.ctype != ctype {
			return shape{}, pgerror.New(pgcode.Internal, "ptarray_append_ptarray: appending mixed dimensionality is not allowed")
		}
		seq := part.coords
		if end := len(line.coords) - dims; end >= 0 && line.coords[end] == seq[0] && line.coords[end+1] == seq[1] {
			seq = seq[dims:]
		}
		line.coords = append(line.coords, seq...)
	}
	return line, nil
}

// stMakeLine implements the two-argument ST_MakeLine, which joins two points or linestrings.
func stMakeLine(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(Geometry), args[1].(Geometry)
	if err := checkSameSRID("LWGEOM_makeline", a, b); err != nil {
		return nil, err
	}
	for _, g := range []Geometry{a, b} {
		if g.Shape.typeNum != pointType && g.Shape.typeNum != lineType {
			return nil, pgerror.New(pgcode.Internal, "Input geometries must be points or lines")
		}
	}
	line, err := makeLine([]shape{a.Shape, b.Shape})
	if err != nil {
		return nil, err
	}
	return a.withShape(line), nil
}

// stMakeLineArray implements ST_MakeLine over an array, which joins its points, the points of its multipoints, and its
// linestrings, skipping everything else. It returns null with a notice when nothing can be joined.
func stMakeLineArray(ctx *sql.Context, args ...any) (any, error) {
	elems := args[0].([]any)
	if len(elems) == 0 {
		return nil, nil
	}
	var parts []shape
	var srid int32
	for _, elem := range elems {
		if elem == nil {
			continue
		}
		g := elem.(Geometry)
		switch g.Shape.typeNum {
		case pointType, lineType, multiPointType:
		default:
			continue
		}
		if len(parts) == 0 {
			srid = g.SRID
		} else if g.SRID != srid {
			return nil, pgerror.Newf(pgcode.Internal, "LWGEOM_makeline_garray: Operation on mixed SRID geometries %s %d != %d",
				typeNameForNumber(g.Shape.typeNum), g.SRID, srid)
		}
		if g.Shape.typeNum == multiPointType {
			parts = append(parts, g.Shape.members...)
		} else {
			parts = append(parts, g.Shape)
		}
	}
	if len(parts) == 0 {
		sendNotice(ctx, "NOTICE", "No points or linestrings in input array")
		return nil, nil
	}
	line, err := makeLine(parts)
	if err != nil {
		return nil, err
	}
	return Geometry{SRID: srid, Shape: line}, nil
}

// stMakePolygon implements ST_MakePolygon, which builds a polygon from a closed linestring shell and an optional array
// of closed linestring holes.
func stMakePolygon(ctx *sql.Context, args ...any) (any, error) {
	shell := args[0].(Geometry)
	if shell.Shape.typeNum != lineType {
		return nil, pgerror.New(pgcode.Internal, "Shell is not a line")
	}
	var holes []Geometry
	if len(args) > 1 {
		for i, elem := range args[1].([]any) {
			if hole, ok := elem.(Geometry); ok && hole.Shape.typeNum == lineType {
				holes = append(holes, hole)
			} else {
				return nil, pgerror.Newf(pgcode.Internal, "Hole %d is not a line", i)
			}
		}
	}
	dims := shell.Shape.ctype.Dimension()
	if len(shell.Shape.coords) < 4*dims {
		return nil, pgerror.New(pgcode.Internal, "lwpoly_from_lwlines: shell must have at least 4 points")
	}
	if !isClosed(shell.Shape.coords, shell.Shape.ctype, false) {
		return nil, pgerror.New(pgcode.Internal, "lwpoly_from_lwlines: shell must be closed")
	}
	polygon := shape{typeNum: polygonType, ctype: shell.Shape.ctype, rings: [][]float64{shell.Shape.coords}}
	for _, hole := range holes {
		if hole.SRID != shell.SRID {
			return nil, pgerror.New(pgcode.Internal, "lwpoly_from_lwlines: mixed SRIDs in input lines")
		}
		if len(hole.Shape.coords) < 4*hole.Shape.ctype.Dimension() {
			return nil, pgerror.New(pgcode.Internal, "lwpoly_from_lwlines: holes must have at least 4 points")
		}
		if !isClosed(hole.Shape.coords, hole.Shape.ctype, false) {
			return nil, pgerror.New(pgcode.Internal, "lwpoly_from_lwlines: holes must be closed")
		}
		polygon.rings = append(polygon.rings, hole.Shape.coords)
	}
	for _, hole := range holes {
		if hole.Shape.ctype != shell.Shape.ctype {
			return nil, pgerror.New(pgcode.Internal, "lwpoly_construct: mixed dimensioned rings")
		}
	}
	return shell.withShape(polygon), nil
}

// stPolygon implements ST_Polygon, which builds a polygon from a closed linestring and gives it the SRID.
func stPolygon(ctx *sql.Context, args ...any) (any, error) {
	polygon, err := stMakePolygon(ctx, args[0])
	if err != nil {
		return nil, err
	}
	return stSetSRID(ctx, polygon, args[1])
}

// typedConstructor returns an implementation of a constructor that reads text or WKB with `parse` and returns null
// unless GeometryType names the expected type, parsing twice like the SQL that PostGIS uses, which repeats its
// warnings.
func typedConstructor(parse extdef.Function, expected string) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		g, err := parse(ctx, args...)
		if err != nil {
			return nil, err
		}
		if name, _ := geometryType(ctx, g); name != expected {
			return nil, nil
		}
		return parse(ctx, args...)
	}
}

// typedConstructorRoutines returns the constructors that read a single type from text or WKB, each with an optional
// SRID except st_multipointfromtext and st_multilinefromwkb.
func typedConstructorRoutines() []extdef.Routine {
	var routines []extdef.Routine
	add := func(names []string, input string, parse extdef.Function, expected string) {
		for _, name := range names {
			impl := typedConstructor(parse, expected)
			routines = append(routines, extdef.Routine{Name: name, Symbol: name + "_" + input, Parameters: params(input), Returns: "geometry", Strict: true, Impl: impl})
			if name != "st_multipointfromtext" && name != "st_multilinefromwkb" {
				routines = append(routines, extdef.Routine{Name: name, Symbol: name + "_" + input + "_int4", Parameters: params(input, "int4"), Returns: "geometry", Strict: true, Impl: impl})
			}
		}
	}
	add([]string{"st_pointfromtext"}, "text", stGeomFromText, "POINT")
	add([]string{"st_linefromtext"}, "text", stGeomFromText, "LINESTRING")
	add([]string{"st_polyfromtext", "st_polygonfromtext"}, "text", stGeomFromText, "POLYGON")
	add([]string{"st_mpointfromtext", "st_multipointfromtext"}, "text", stGeomFromText, "MULTIPOINT")
	add([]string{"st_mlinefromtext", "st_multilinestringfromtext"}, "text", stGeomFromText, "MULTILINESTRING")
	add([]string{"st_mpolyfromtext", "st_multipolygonfromtext"}, "text", stGeomFromText, "MULTIPOLYGON")
	add([]string{"st_geomcollfromtext"}, "text", stGeomFromText, "GEOMETRYCOLLECTION")
	add([]string{"st_pointfromwkb"}, "bytea", stGeomFromWKB, "POINT")
	add([]string{"st_linefromwkb", "st_linestringfromwkb"}, "bytea", stGeomFromWKB, "LINESTRING")
	add([]string{"st_polyfromwkb", "st_polygonfromwkb"}, "bytea", stGeomFromWKB, "POLYGON")
	add([]string{"st_mpointfromwkb", "st_multipointfromwkb"}, "bytea", stGeomFromWKB, "MULTIPOINT")
	add([]string{"st_mlinefromwkb", "st_multilinefromwkb"}, "bytea", stGeomFromWKB, "MULTILINESTRING")
	add([]string{"st_mpolyfromwkb", "st_multipolyfromwkb"}, "bytea", stGeomFromWKB, "MULTIPOLYGON")
	add([]string{"st_geomcollfromwkb"}, "bytea", stGeomFromWKB, "GEOMETRYCOLLECTION")
	return routines
}

// constructorRoutines returns the routines that build geometries and geographies from coordinates, text, or WKB.
func constructorRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "geomfromewkb", Symbol: "geomfromewkb_bytea", Parameters: params("bytea"), Returns: "geometry", Strict: true, Impl: geometryFromBytea},
		{Name: "st_collect", Symbol: "LWGEOM_collect", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}, Returns: "geometry", Impl: stCollect},
		{Name: "st_collect", Symbol: "LWGEOM_collect_garray", Parameters: params("_geometry"), Returns: "geometry", Strict: true, Impl: stCollectArray},
		{Name: "st_makeline", Symbol: "LWGEOM_makeline", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}, Returns: "geometry", Strict: true, Impl: stMakeLine},
		{Name: "st_makeline", Symbol: "LWGEOM_makeline_garray", Parameters: params("_geometry"), Returns: "geometry", Strict: true, Impl: stMakeLineArray},
		{Name: "st_makepolygon", Symbol: "st_makepolygon_geometry", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stMakePolygon},
		{Name: "st_makepolygon", Symbol: "st_makepolygon_geometry_geometry", Parameters: params("geometry", "_geometry"), Returns: "geometry", Strict: true, Impl: stMakePolygon},
		{Name: "st_polygon", Symbol: "st_polygon_geometry_int4", Parameters: params("geometry", "int4"), Returns: "geometry", Strict: true, Impl: stPolygon},
		{Name: "geomfromewkt", Symbol: "geomfromewkt_text", Parameters: params("text"), Returns: "geometry", Strict: true, Impl: stGeomFromEWKT},
		{Name: "st_geogfromtext", Symbol: "st_geogfromtext_text", Parameters: params("text"), Returns: "geography", Strict: true, Impl: stGeogFromText},
		{Name: "st_geogfromwkb", Symbol: "st_geogfromwkb_bytea", Parameters: params("bytea"), Returns: "geography", Strict: true, Impl: geographyFromBytea},
		{Name: "st_geographyfromtext", Symbol: "st_geographyfromtext_text", Parameters: params("text"), Returns: "geography", Strict: true, Impl: stGeogFromText},
		{Name: "st_geometryfromtext", Symbol: "st_geometryfromtext_text", Parameters: params("text"), Returns: "geometry", Strict: true, Impl: stGeomFromText},
		{Name: "st_geometryfromtext", Symbol: "st_geometryfromtext_text_int4", Parameters: params("text", "int4"), Returns: "geometry", Strict: true, Impl: stGeomFromText},
		{Name: "st_geomfromewkb", Symbol: "st_geomfromewkb_bytea", Parameters: params("bytea"), Returns: "geometry", Strict: true, Impl: geometryFromBytea},
		{Name: "st_geomfromewkt", Symbol: "st_geomfromewkt_text", Parameters: params("text"), Returns: "geometry", Strict: true, Impl: stGeomFromEWKT},
		{Name: "st_geomfromtext", Symbol: "st_geomfromtext_text", Parameters: params("text"), Returns: "geometry", Strict: true, Impl: stGeomFromText},
		{Name: "st_geomfromtext", Symbol: "st_geomfromtext_text_int4", Parameters: params("text", "int4"), Returns: "geometry", Strict: true, Impl: stGeomFromText},
		{Name: "st_geomfromwkb", Symbol: "st_geomfromwkb_bytea", Parameters: params("bytea"), Returns: "geometry", Strict: true, Impl: stGeomFromWKB},
		{Name: "st_geomfromwkb", Symbol: "st_geomfromwkb_bytea_int4", Parameters: params("bytea", "int4"), Returns: "geometry", Strict: true, Impl: stGeomFromWKB},
		{Name: "st_makepoint", Symbol: "st_makepoint_float8_float8", Parameters: params("float8", "float8"), Returns: "geometry", Strict: true, Impl: stMakePoint},
		{Name: "st_makepoint", Symbol: "st_makepoint_float8_float8_float8", Parameters: params("float8", "float8", "float8"), Returns: "geometry", Strict: true, Impl: stMakePoint},
		{Name: "st_makepoint", Symbol: "st_makepoint_float8_float8_float8_float8", Parameters: params("float8", "float8", "float8", "float8"), Returns: "geometry", Strict: true, Impl: stMakePoint},
		{Name: "st_point", Symbol: "st_point_float8_float8", Parameters: params("float8", "float8"), Returns: "geometry", Strict: true, Impl: stMakePoint},
		{Name: "st_point", Symbol: "ST_Point", Parameters: []extdef.Parameter{{Type: "float8"}, {Type: "float8"}, {Name: "srid", Type: "int4"}}, Returns: "geometry", Strict: true, Impl: stPoint},
		{Name: "st_wkbtosql", Symbol: "st_wkbtosql_bytea", Parameters: []extdef.Parameter{{Name: "wkb", Type: "bytea"}}, Returns: "geometry", Strict: true, Impl: stGeomFromWKB},
		{Name: "st_wkttosql", Symbol: "st_wkttosql_text", Parameters: params("text"), Returns: "geometry", Strict: true, Impl: stGeomFromText},
	}
}
