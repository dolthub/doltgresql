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
	"math/rand/v2"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// isCurved returns whether the type number names one of the curved types.
func isCurved(typeNum int) bool {
	switch typeNum {
	case circStringType, compoundType, curvePolyType, multiCurveType, multiSurfaceType:
		return true
	default:
		return false
	}
}

// stMakeValid implements ST_MakeValid, which takes optional "key=value" parameters choosing the method and whether
// collapsed parts are kept. The result drops M coordinates.
func stMakeValid(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	structure, keepCollapsed := false, false
	if len(args) > 1 {
		for _, param := range strings.Fields(args[1].(string)) {
			key, val, _ := strings.Cut(param, "=")
			switch key {
			case "method":
				if val != "linework" && val != "structure" {
					return nil, pgerror.Newf(pgcode.Internal, "Unsupported value for 'method', '%s'. Use 'linework' or 'structure'.", val)
				}
				structure = val == "structure"
			case "keepcollapsed":
				if val != "true" && val != "false" {
					return nil, pgerror.Newf(pgcode.Internal, "Unsupported value for 'keepcollapsed', '%s'. Use 'true' or 'false'", val)
				}
				keepCollapsed = val == "true"
			}
		}
	}
	var unsupported int
	atomicMembers(g.Shape, true, func(member shape) {
		if unsupported == 0 && isCurved(member.typeNum) {
			unsupported = member.typeNum
		}
	})
	if unsupported != 0 {
		return nil, pgerror.Newf(pgcode.Internal, "ST_MakeValid: unsupported geometry type %s", typeNameForNumber(unsupported))
	}
	return unaryImpl("GEOSMakeValid", func(geos *spatial.GEOS, a spatial.Geometry) (spatial.Geometry, error) {
		return geos.MakeValid(a, structure, keepCollapsed)
	})(ctx, g)
}

// stSimplifyPreserveTopology implements ST_SimplifyPreserveTopology.
func stSimplifyPreserveTopology(ctx *sql.Context, args ...any) (any, error) {
	return unaryImpl("GEOSTopologyPreserveSimplify", func(geos *spatial.GEOS, a spatial.Geometry) (spatial.Geometry, error) {
		return geos.SimplifyPreserveTopology(a, args[1].(float64))
	})(ctx, args[0])
}

// stLineMerge implements ST_LineMerge, which keeps the direction of the lines when it is given true. An empty geometry
// is returned unchanged.
func stLineMerge(ctx *sql.Context, args ...any) (any, error) {
	if args[0].(Geometry).Shape.isEmpty() {
		return args[0], nil
	}
	directed := len(args) > 1 && args[1].(bool)
	return unaryImpl("GEOSLineMerge", func(geos *spatial.GEOS, a spatial.Geometry) (spatial.Geometry, error) {
		return geos.LineMerge(a, directed)
	})(ctx, args[0])
}

// stNode implements ST_Node, which splits the lines wherever they cross or touch. The noded pieces are merged back
// together and then split again at the input's line ends, so that a line only breaks where it meets another line or
// ends.
func stNode(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	result, err := unaryImpl("GEOSNode", func(geos *spatial.GEOS, a spatial.Geometry) (spatial.Geometry, error) {
		noded, err := geos.Node(a)
		if err != nil {
			return spatial.Geometry{}, err
		}
		return geos.LineMerge(noded, false)
	})(ctx, g)
	if err != nil {
		return nil, err
	}
	merged := result.(Geometry)
	ends := make(map[[2]float64]bool)
	atomicMembers(g.Shape, false, func(member shape) {
		if member.typeNum == lineType && len(member.coords) > 0 {
			dims := member.ctype.Dimension()
			ends[[2]float64{member.coords[0], member.coords[1]}] = true
			ends[[2]float64{member.coords[len(member.coords)-dims], member.coords[len(member.coords)-dims+1]}] = true
		}
	})
	split := shape{typeNum: multiLineType, ctype: merged.Shape.ctype}
	dims := merged.Shape.ctype.Dimension()
	atomicMembers(merged.Shape, false, func(line shape) {
		start := 0
		for i := dims; i+dims < len(line.coords); i += dims {
			if ends[[2]float64{line.coords[i], line.coords[i+1]}] {
				split.members = append(split.members, shape{typeNum: lineType, ctype: line.ctype, coords: line.coords[start : i+dims]})
				start = i
			}
		}
		split.members = append(split.members, shape{typeNum: lineType, ctype: line.ctype, coords: line.coords[start:]})
	})
	return merged.withShape(split), nil
}

// binaryImpl returns an implementation of a GEOS operation on two geometries with matching SRIDs, whose result keeps
// the first geometry's SRID and the Z coordinates of either input.
func binaryImpl(funcName string, function string, op func(geos *spatial.GEOS, a spatial.Geometry, b spatial.Geometry) (spatial.Geometry, error)) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].(Geometry), args[1].(Geometry)
		geosA, geosB, err := geosPair(funcName, a, b)
		if err != nil {
			return nil, err
		}
		geos := spatial.NewGEOS()
		defer geos.Close()
		result, err := op(geos, geosA, geosB)
		if err != nil {
			return nil, geosError(function, err)
		}
		return fromGEOS(result, a.SRID, a.Shape.hasZ() || b.Shape.hasZ()), nil
	}
}

// stSnap implements ST_Snap.
func stSnap(ctx *sql.Context, args ...any) (any, error) {
	return binaryImpl("ST_Snap", "GEOSSnap", func(geos *spatial.GEOS, a spatial.Geometry, b spatial.Geometry) (spatial.Geometry, error) {
		return geos.Snap(a, b, args[2].(float64))
	})(ctx, args[0], args[1])
}

// stOffsetCurve implements ST_OffsetCurve, which takes optional "key=value" parameters for the joins. A distance of
// zero returns the line unchanged.
func stOffsetCurve(ctx *sql.Context, args ...any) (any, error) {
	g, distance := args[0].(Geometry), args[1].(float64)
	params := spatial.BufferParams{QuadSegs: 8, Join: spatial.BufferRound, MitreLimit: 5}
	if len(args) > 2 {
		for _, param := range strings.Fields(args[2].(string)) {
			key, val, _ := strings.Cut(param, "=")
			switch key {
			case "join":
				switch val {
				case "round":
					params.Join = spatial.BufferRound
				case "mitre", "miter":
					params.Join = spatial.BufferMitre
				case "bevel":
					params.Join = spatial.BufferBevel
				default:
					return nil, pgerror.Newf(pgcode.Internal, "Invalid buffer end cap style: %s (accept: 'round', 'mitre', 'miter' or 'bevel')", val)
				}
			case "mitre_limit", "miter_limit":
				params.MitreLimit = parseCFloat(val)
			case "quad_segs":
				params.QuadSegs = parseCInt(val)
			default:
				return nil, pgerror.Newf(pgcode.Internal, "Invalid buffer parameter: %s (accept: 'join', 'mitre_limit', 'miter_limit and 'quad_segs')", key)
			}
		}
	}
	switch g.Shape.typeNum {
	case lineType, multiLineType, circStringType, compoundType, multiCurveType, collectionType:
	default:
		return nil, pgerror.Newf(pgcode.Internal, "lwgeom_offsetcurve: input is not linear (type %s)", typeNameForNumber(g.Shape.typeNum))
	}
	if distance == 0 {
		return g, nil
	}
	return unaryImpl("GEOSOffsetCurve", func(geos *spatial.GEOS, a spatial.Geometry) (spatial.Geometry, error) {
		return geos.OffsetCurve(a, distance, params)
	})(ctx, g)
}

// stBuildArea implements ST_BuildArea, which returns null when the lines enclose nothing.
func stBuildArea(ctx *sql.Context, args ...any) (any, error) {
	result, err := unaryImpl("GEOSBuildArea", (*spatial.GEOS).BuildArea)(ctx, args...)
	if err != nil || result.(Geometry).Shape.isEmpty() {
		return nil, err
	}
	return result, nil
}

// stConcaveHull implements ST_ConcaveHull, which fills the gaps between the polygons of a polygonal geometry, and
// otherwise wraps the geometry's vertices.
func stConcaveHull(ctx *sql.Context, args ...any) (any, error) {
	g, ratio, allowHoles := args[0].(Geometry), args[1].(float64), args[2].(bool)
	polygonal := g.Shape.typeNum == polygonType || g.Shape.typeNum == multiPolygonType
	return unaryImpl("GEOSConcaveHull", func(geos *spatial.GEOS, a spatial.Geometry) (spatial.Geometry, error) {
		if polygonal {
			return geos.ConcaveHullOfPolygons(a, ratio, false, allowHoles)
		}
		return geos.ConcaveHull(a, ratio, allowHoles)
	})(ctx, g)
}

// stSimplifyPolygonHull implements ST_SimplifyPolygonHull.
func stSimplifyPolygonHull(ctx *sql.Context, args ...any) (any, error) {
	return unaryImpl("GEOSPolygonHullSimplify", func(geos *spatial.GEOS, a spatial.Geometry) (spatial.Geometry, error) {
		return geos.PolygonHullSimplify(a, args[2].(bool), args[1].(float64))
	})(ctx, args[0])
}

// stDelaunayTriangles implements ST_DelaunayTriangles, whose flags choose a collection of triangles (0), a
// multilinestring of their edges (1), or a TIN (2).
func stDelaunayTriangles(ctx *sql.Context, args ...any) (any, error) {
	tolerance, flags := args[1].(float64), args[2].(int32)
	result, err := unaryImpl("GEOSDelaunayTriangulation", func(geos *spatial.GEOS, a spatial.Geometry) (spatial.Geometry, error) {
		return geos.DelaunayTriangulation(a, tolerance, flags == 1)
	})(ctx, args[0])
	if err != nil || flags != 2 {
		return result, err
	}
	g := result.(Geometry)
	tin := shape{typeNum: tinType, ctype: g.Shape.ctype}
	for _, member := range g.Shape.members {
		tin.members = append(tin.members, shape{typeNum: triangleType, ctype: member.ctype, coords: member.rings[0]})
	}
	return g.withShape(tin), nil
}

// stTriangulatePolygon implements ST_TriangulatePolygon.
func stTriangulatePolygon(ctx *sql.Context, args ...any) (any, error) {
	return unaryImpl("GEOSConstrainedDelaunayTriangulation", (*spatial.GEOS).ConstrainedDelaunayTriangulation)(ctx, args...)
}

// stVoronoi implements _ST_Voronoi, which returns the Voronoi diagram of the vertices as polygons or as lines,
// extended to the envelope of the clipping geometry when one is given. A null input returns null.
func stVoronoi(ctx *sql.Context, args ...any) (any, error) {
	if args[0] == nil || args[2] == nil || args[3] == nil {
		return nil, nil
	}
	g, tolerance, polygons := args[0].(Geometry), args[2].(float64), args[3].(bool)
	var envelope *spatial.Geometry
	if clip, ok := args[1].(Geometry); ok {
		if b, ok := clip.Shape.bounds(); ok {
			envelopeGEOS := flatBoxShape(b.minX, b.minY, b.maxX, b.maxY).linearToGEOS()
			envelope = &envelopeGEOS
		}
	}
	return unaryImpl("GEOSVoronoiDiagram", func(geos *spatial.GEOS, a spatial.Geometry) (spatial.Geometry, error) {
		return geos.VoronoiDiagram(a, envelope, tolerance, !polygons)
	})(ctx, g)
}

// voronoiImpl returns the implementation of ST_VoronoiPolygons or ST_VoronoiLines.
func voronoiImpl(polygons bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		return stVoronoi(ctx, args[0], args[2], args[1], polygons)
	}
}

// stMinimumClearance implements ST_MinimumClearance, which is infinite for a geometry without two distinct vertices.
func stMinimumClearance(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	geosG, err := g.Shape.toGEOS()
	if err != nil {
		return nil, err
	}
	geos := spatial.NewGEOS()
	defer geos.Close()
	clearance, err := geos.MinimumClearance(geosG)
	if err != nil {
		return nil, geosError("GEOSMinimumClearance", err)
	}
	return clearance, nil
}

// stMinimumClearanceLine implements ST_MinimumClearanceLine.
func stMinimumClearanceLine(ctx *sql.Context, args ...any) (any, error) {
	return unaryImpl("GEOSMinimumClearanceLine", (*spatial.GEOS).MinimumClearanceLine)(ctx, args...)
}

// distanceImpl returns an implementation of a GEOS distance between two geometries, after densifying their segments
// by the optional fraction.
func distanceImpl(function string, op func(geos *spatial.GEOS, a spatial.Geometry, b spatial.Geometry, densifyFraction float64) (float64, error)) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		geosA, err := args[0].(Geometry).Shape.toGEOS()
		if err != nil {
			return nil, err
		}
		geosB, err := args[1].(Geometry).Shape.toGEOS()
		if err != nil {
			return nil, err
		}
		densifyFraction := -1.0
		if len(args) > 2 {
			densifyFraction = args[2].(float64)
		}
		geos := spatial.NewGEOS()
		defer geos.Close()
		distance, err := op(geos, geosA, geosB, densifyFraction)
		if err != nil {
			return nil, geosError(function, err)
		}
		return distance, nil
	}
}

// stClipByBox2D implements ST_ClipByBox2D, which returns an empty geometry of the input's type when nothing is left.
func stClipByBox2D(ctx *sql.Context, args ...any) (any, error) {
	g, box := args[0].(Geometry), args[1].(box2d)
	clipped, err := unaryImpl("GEOSClipByRect", func(geos *spatial.GEOS, a spatial.Geometry) (spatial.Geometry, error) {
		return geos.ClipByRect(a, box.minX, box.minY, box.maxX, box.maxY)
	})(ctx, g)
	if err != nil {
		return nil, err
	}
	if clipped.(Geometry).Shape.isEmpty() {
		return g.withShape(shape{typeNum: g.Shape.typeNum, ctype: clipped.(Geometry).Shape.ctype}), nil
	}
	return clipped, nil
}

// defaultCircleTolerance returns the tolerance that the inscribed and empty circle functions use when they are given
// none: a thousandth of the larger side of the geometry's bounding box.
func defaultCircleTolerance(s shape) float64 {
	b, ok := s.bounds()
	if !ok {
		return 0
	}
	return math.Max(b.maxX-b.minX, b.maxY-b.minY) / 1000
}

// circleRecord returns the record of an inscribed or empty circle, given the line from its center to the nearest
// point of the input.
func circleRecord(ctx *sql.Context, srid int32, line Geometry) (any, error) {
	geometryType, err := extensionType(ctx, "geometry")
	if err != nil {
		return nil, err
	}
	center, nearest := shape{typeNum: pointType, ctype: line.Shape.ctype}, shape{typeNum: pointType, ctype: line.Shape.ctype}
	radius := 0.0
	if coords := line.Shape.coords; len(coords) >= 4 {
		center.coords, nearest.coords = coords[:2], coords[2:4]
		radius = math.Hypot(coords[2]-coords[0], coords[3]-coords[1])
	}
	return []pgtypes.RecordValue{
		{Value: Geometry{SRID: srid, Shape: center}, Type: geometryType},
		{Value: Geometry{SRID: srid, Shape: nearest}, Type: geometryType},
		{Value: radius, Type: pgtypes.Float64},
	}, nil
}

// stMaximumInscribedCircle implements ST_MaximumInscribedCircle.
func stMaximumInscribedCircle(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	tolerance := defaultCircleTolerance(g.Shape)
	line, err := unaryImpl("GEOSMaximumInscribedCircle", func(geos *spatial.GEOS, a spatial.Geometry) (spatial.Geometry, error) {
		return geos.MaximumInscribedCircle(a, tolerance)
	})(ctx, g)
	if err != nil {
		return nil, err
	}
	return circleRecord(ctx, g.SRID, line.(Geometry))
}

// stLargestEmptyCircle implements ST_LargestEmptyCircle, whose center lies within the boundary unless the boundary is
// empty, in which case it lies within the convex hull of the obstacles.
func stLargestEmptyCircle(ctx *sql.Context, args ...any) (any, error) {
	g, tolerance, boundary := args[0].(Geometry), args[1].(float64), args[2].(Geometry)
	if tolerance == 0 {
		tolerance = defaultCircleTolerance(g.Shape)
	}
	geosG, err := g.Shape.toGEOS()
	if err != nil {
		return nil, err
	}
	geosBoundary, err := boundary.Shape.toGEOS()
	if err != nil {
		return nil, err
	}
	var bounding *spatial.Geometry
	if !boundary.Shape.isEmpty() && boundary.Shape.hasFiniteCoords() {
		bounding = &geosBoundary
	}
	geos := spatial.NewGEOS()
	defer geos.Close()
	line, err := geos.LargestEmptyCircle(geosG, bounding, tolerance)
	if err != nil {
		return nil, pgerror.New(pgcode.Internal, "Error calculating GEOSLargestEmptyCircle.")
	}
	return circleRecord(ctx, g.SRID, fromGEOS(line, g.SRID, false))
}

// hasFiniteCoords returns whether every ordinate of the shape is a finite number.
func (s shape) hasFiniteCoords() bool {
	for _, seq := range s.sequences() {
		for _, val := range seq {
			if math.IsNaN(val) || math.IsInf(val, 0) {
				return false
			}
		}
	}
	return true
}

// minimumBoundingCircle returns the center and radius of the smallest circle holding the geometry, or false when the
// geometry is empty.
func minimumBoundingCircle(g Geometry) (shape, float64, bool, error) {
	if g.Shape.isEmpty() {
		return shape{}, 0, false, nil
	}
	geosG, err := g.Shape.toGEOS()
	if err != nil {
		return shape{}, 0, false, err
	}
	geos := spatial.NewGEOS()
	defer geos.Close()
	_, center, radius, err := geos.MinimumBoundingCircle(geosG)
	if err != nil {
		return shape{}, 0, false, geosError("GEOSMinimumBoundingCircle", err)
	}
	return shapeFromGEOS(center, false), radius, true, nil
}

// stMinimumBoundingCircle implements ST_MinimumBoundingCircle, which approximates the circle with the given number
// of segments per quarter circle. A circle of radius zero is its center, and an empty geometry gives an empty point.
func stMinimumBoundingCircle(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	center, radius, ok, err := minimumBoundingCircle(g)
	if err != nil || !ok {
		return g.withShape(shape{typeNum: pointType}), err
	}
	if radius == 0 {
		return g.withShape(center), nil
	}
	return stBuffer(ctx, g.withShape(center), radius, args[1])
}

// stMinimumBoundingRadius implements ST_MinimumBoundingRadius.
func stMinimumBoundingRadius(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	center, radius, ok, err := minimumBoundingCircle(g)
	if err != nil {
		return nil, err
	}
	if !ok {
		center = shape{typeNum: pointType}
	}
	geometryType, err := extensionType(ctx, "geometry")
	if err != nil {
		return nil, err
	}
	return []pgtypes.RecordValue{
		{Value: g.withShape(center), Type: geometryType},
		{Value: radius, Type: pgtypes.Float64},
	}, nil
}

// stReducePrecision implements ST_ReducePrecision, which rounds the coordinates to the grid while keeping the
// geometry valid.
func stReducePrecision(ctx *sql.Context, args ...any) (any, error) {
	return unaryImpl("GEOSGeom_setPrecision", func(geos *spatial.GEOS, a spatial.Geometry) (spatial.Geometry, error) {
		return geos.SetPrecision(a, args[1].(float64), 0)
	})(ctx, args[0])
}

// polygonizeAll returns the polygons that the lines of the gathered geometries enclose, as a collection with the
// first geometry's SRID.
func polygonizeAll(ctx *sql.Context, args ...any) (any, error) {
	geometries := args[0].([]any)
	converted := make([]spatial.Geometry, len(geometries))
	hasZ := false
	for i, elem := range geometries {
		g := elem.(Geometry)
		hasZ = hasZ || g.Shape.hasZ()
		geosG, err := g.Shape.toGEOS()
		if err != nil {
			return nil, err
		}
		converted[i] = geosG
	}
	geos := spatial.NewGEOS()
	defer geos.Close()
	result, err := geos.Polygonize(converted)
	if err != nil {
		return nil, geosError("GEOSPolygonize", err)
	}
	return fromGEOS(result, geometries[0].(Geometry).SRID, hasZ), nil
}

// bdPolyFromText implements ST_BdPolyFromText and ST_BdMPolyFromText, which build the area that the lines of a
// multilinestring in WKT enclose. The single polygon form rejects an area of several polygons.
func bdPolyFromText(multi bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		g, err := parseGeometryInput(ctx, args[0].(string))
		if err != nil {
			return nil, err
		}
		g.SRID = args[1].(int32)
		if g.Shape.typeNum != multiLineType {
			return nil, pgerror.New(pgcode.RaiseException, "Input is not a MultiLinestring")
		}
		area, err := stBuildArea(ctx, g)
		if err != nil || area == nil {
			return nil, err
		}
		result := area.(Geometry)
		if multi {
			return result.withShape(asMulti(result.Shape)), nil
		}
		if result.Shape.typeNum != polygonType {
			return nil, pgerror.New(pgcode.RaiseException, "Input returns more then a single polygon, try using BdMPolyFromText instead")
		}
		return result, nil
	}
}

// stGeneratePoints implements ST_GeneratePoints, which returns random points inside a polygon or multipolygon,
// sharing the requested number among a multipolygon's polygons by area. A seed makes the points repeatable. A count
// below one, or an empty geometry, gives null.
func stGeneratePoints(ctx *sql.Context, args ...any) (any, error) {
	g, count := args[0].(Geometry), args[1].(int32)
	source := rand.NewPCG(rand.Uint64(), rand.Uint64())
	if len(args) > 2 {
		seed := args[2].(int32)
		if seed <= 0 {
			return nil, pgerror.New(pgcode.Internal, "ST_GeneratePoints: seed must be greater than zero")
		}
		source = rand.NewPCG(uint64(seed), uint64(seed))
	}
	polygons := []shape{g.Shape}
	switch g.Shape.typeNum {
	case polygonType:
	case multiPolygonType:
		polygons = g.Shape.members
	default:
		return nil, pgerror.Newf(pgcode.Internal, "lwgeom_to_points: unsupported geometry type '%s'", typeNameForNumber(g.Shape.typeNum))
	}
	if count < 1 || g.Shape.isEmpty() {
		return nil, nil
	}
	areas := make([]float64, len(polygons))
	total := 0.0
	for i, polygon := range polygons {
		areas[i], _ = shapeArea(polygon)
		if areas[i] == 0 && !polygon.isEmpty() {
			return nil, pgerror.New(pgcode.Internal, "lwpoly_to_points: zero area input polygon, TBD")
		}
		total += areas[i]
	}
	random := rand.New(source)
	var points []shape
	for i, polygon := range polygons {
		b, ok := polygon.bounds()
		if !ok {
			continue
		}
		dims := polygon.ctype.Dimension()
		rings := make([][]point3, len(polygon.rings))
		for r, ring := range polygon.rings {
			for j := 0; j+dims <= len(ring); j += dims {
				rings[r] = append(rings[r], point3{ring[j], ring[j+1]})
			}
		}
		area := newPolygon3(rings)
		for n := int(math.Round(float64(count) * areas[i] / total)); n > 0; {
			p := point3{b.minX + random.Float64()*(b.maxX-b.minX), b.minY + random.Float64()*(b.maxY-b.minY)}
			if area.contains(p) {
				points = append(points, shape{typeNum: pointType, coords: []float64{p[0], p[1]}})
				n--
			}
		}
	}
	return g.withShape(shape{typeNum: multiPointType, members: points}), nil
}

// constructionRoutines returns the routines that build new geometries through GEOS.
func constructionRoutines() []extdef.Routine {
	circleOutputs := []extdef.Parameter{{Name: "center", Type: "geometry"}, {Name: "nearest", Type: "geometry"}, {Name: "radius", Type: "float8"}}
	voronoiParams := []extdef.Parameter{{Name: "g1", Type: "geometry"}, {Name: "tolerance", Type: "float8", Default: `0.0`}, {Name: "extend_to", Type: "geometry", Default: `NULL::geometry`}}
	return []extdef.Routine{
		{Name: "_st_voronoi", Symbol: "ST_Voronoi", Parameters: []extdef.Parameter{{Name: "g1", Type: "geometry"}, {Name: "clip", Type: "geometry", Default: `NULL::geometry`}, {Name: "tolerance", Type: "float8", Default: `0.0`}, {Name: "return_polygons", Type: "bool", Default: `true`}}, Returns: "geometry", Impl: stVoronoi},
		{Name: "st_bdmpolyfromtext", Symbol: "st_bdmpolyfromtext_text_int4", Parameters: params("text", "int4"), Returns: "geometry", Strict: true, Impl: bdPolyFromText(true)},
		{Name: "st_bdpolyfromtext", Symbol: "st_bdpolyfromtext_text_int4", Parameters: params("text", "int4"), Returns: "geometry", Strict: true, Impl: bdPolyFromText(false)},
		{Name: "st_buildarea", Symbol: "ST_BuildArea", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stBuildArea},
		{Name: "st_cleangeometry", Symbol: "ST_CleanGeometry", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stMakeValid},
		{Name: "st_clipbybox2d", Symbol: "ST_ClipByBox2d", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "box", Type: "box2d"}}, Returns: "geometry", Strict: true, Impl: stClipByBox2D},
		{Name: "st_concavehull", Symbol: "ST_ConcaveHull", Parameters: []extdef.Parameter{{Name: "param_geom", Type: "geometry"}, {Name: "param_pctconvex", Type: "float8"}, {Name: "param_allow_holes", Type: "bool", Default: `false`}}, Returns: "geometry", Strict: true, Impl: stConcaveHull},
		{Name: "st_delaunaytriangles", Symbol: "ST_DelaunayTriangles", Parameters: []extdef.Parameter{{Name: "g1", Type: "geometry"}, {Name: "tolerance", Type: "float8", Default: `0.0`}, {Name: "flags", Type: "int4", Default: `0`}}, Returns: "geometry", Strict: true, Impl: stDelaunayTriangles},
		{Name: "st_frechetdistance", Symbol: "ST_FrechetDistance", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}, {Type: "float8", Default: `'-1'::integer`}}, Returns: "float8", Strict: true, Impl: distanceImpl("GEOSFrechetDistance", (*spatial.GEOS).FrechetDistance)},
		{Name: "st_generatepoints", Symbol: "ST_GeneratePoints", Parameters: []extdef.Parameter{{Name: "area", Type: "geometry"}, {Name: "npoints", Type: "int4"}}, Returns: "geometry", Strict: true, Impl: stGeneratePoints},
		{Name: "st_generatepoints", Symbol: "st_generatepoints_geometry_int4_int4", Parameters: []extdef.Parameter{{Name: "area", Type: "geometry"}, {Name: "npoints", Type: "int4"}, {Name: "seed", Type: "int4"}}, Returns: "geometry", Strict: true, Impl: stGeneratePoints},
		{Name: "st_hausdorffdistance", Symbol: "hausdorffdistance", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}, Returns: "float8", Strict: true, Impl: distanceImpl("GEOSHausdorffDistance", (*spatial.GEOS).HausdorffDistance)},
		{Name: "st_hausdorffdistance", Symbol: "hausdorffdistancedensify", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}, {Type: "float8"}}, Returns: "float8", Strict: true, Impl: distanceImpl("GEOSHausdorffDistance", (*spatial.GEOS).HausdorffDistance)},
		{Name: "st_largestemptycircle", Symbol: "ST_LargestEmptyCircle", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "tolerance", Type: "float8", Default: `0.0`}, {Name: "boundary", Type: "geometry", Default: `'0101000000000000000000F87F000000000000F87F'::geometry`}}, Outputs: circleOutputs, Strict: true, Impl: stLargestEmptyCircle},
		{Name: "st_linemerge", Symbol: "st_linemerge_geometry", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stLineMerge},
		{Name: "st_linemerge", Symbol: "st_linemerge_geometry_bool", Parameters: params("geometry", "bool"), Returns: "geometry", Strict: true, Impl: stLineMerge},
		{Name: "st_makevalid", Symbol: "st_makevalid_geometry", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stMakeValid},
		{Name: "st_makevalid", Symbol: "st_makevalid_geometry_text", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "params", Type: "text"}}, Returns: "geometry", Strict: true, Impl: stMakeValid},
		{Name: "st_maximuminscribedcircle", Symbol: "ST_MaximumInscribedCircle", Parameters: params("geometry"), Outputs: circleOutputs, Strict: true, Impl: stMaximumInscribedCircle},
		{Name: "st_minimumboundingcircle", Symbol: "ST_MinimumBoundingCircle", Parameters: []extdef.Parameter{{Name: "inputgeom", Type: "geometry"}, {Name: "segs_per_quarter", Type: "int4", Default: `48`}}, Returns: "geometry", Strict: true, Impl: stMinimumBoundingCircle},
		{Name: "st_minimumboundingradius", Symbol: "ST_MinimumBoundingRadius", Parameters: params("geometry"), Outputs: []extdef.Parameter{{Name: "center", Type: "geometry"}, {Name: "radius", Type: "float8"}}, Strict: true, Impl: stMinimumBoundingRadius},
		{Name: "st_minimumclearance", Symbol: "ST_MinimumClearance", Parameters: params("geometry"), Returns: "float8", Strict: true, Impl: stMinimumClearance},
		{Name: "st_minimumclearanceline", Symbol: "ST_MinimumClearanceLine", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stMinimumClearanceLine},
		{Name: "st_node", Symbol: "ST_Node", Parameters: []extdef.Parameter{{Name: "g", Type: "geometry"}}, Returns: "geometry", Strict: true, Impl: stNode},
		{Name: "st_offsetcurve", Symbol: "ST_OffsetCurve", Parameters: []extdef.Parameter{{Name: "line", Type: "geometry"}, {Name: "distance", Type: "float8"}, {Name: "params", Type: "text", Default: `''::text`}}, Returns: "geometry", Strict: true, Impl: stOffsetCurve},
		{Name: "st_polygonize", Symbol: "polygonize_garray", Parameters: params("_geometry"), Returns: "geometry", Strict: true, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			var geometries []any
			for _, elem := range args[0].([]any) {
				if elem != nil {
					geometries = append(geometries, elem)
				}
			}
			if len(geometries) == 0 {
				return nil, nil
			}
			return polygonizeAll(ctx, geometries)
		}},
		{Name: "pgis_geometry_polygonize_finalfn", Symbol: "pgis_geometry_polygonize_finalfn", Parameters: params("internal"), Returns: "geometry", Impl: finalImpl(func(ctx *sql.Context, args ...any) (any, error) {
			return polygonizeAll(ctx, args[0])
		})},
		{Name: "st_reduceprecision", Symbol: "ST_ReducePrecision", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "gridsize", Type: "float8"}}, Returns: "geometry", Strict: true, Impl: stReducePrecision},
		{Name: "st_sharedpaths", Symbol: "ST_SharedPaths", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}, Returns: "geometry", Strict: true, Impl: binaryImpl("ST_SharedPaths", "GEOSSharedPaths", (*spatial.GEOS).SharedPaths)},
		{Name: "st_simplifypolygonhull", Symbol: "ST_SimplifyPolygonHull", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "vertex_fraction", Type: "float8"}, {Name: "is_outer", Type: "bool", Default: `true`}}, Returns: "geometry", Strict: true, Impl: stSimplifyPolygonHull},
		{Name: "st_simplifypreservetopology", Symbol: "topologypreservesimplify", Parameters: params("geometry", "float8"), Returns: "geometry", Strict: true, Impl: stSimplifyPreserveTopology},
		{Name: "st_snap", Symbol: "ST_Snap", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}, {Type: "float8"}}, Returns: "geometry", Strict: true, Impl: stSnap},
		{Name: "st_triangulatepolygon", Symbol: "ST_TriangulatePolygon", Parameters: []extdef.Parameter{{Name: "g1", Type: "geometry"}}, Returns: "geometry", Strict: true, Impl: stTriangulatePolygon},
		{Name: "st_voronoilines", Symbol: "st_voronoilines_geometry_float8_geometry", Parameters: voronoiParams, Returns: "geometry", Impl: voronoiImpl(false)},
		{Name: "st_voronoipolygons", Symbol: "st_voronoipolygons_geometry_float8_geometry", Parameters: voronoiParams, Returns: "geometry", Impl: voronoiImpl(true)},
	}
}
