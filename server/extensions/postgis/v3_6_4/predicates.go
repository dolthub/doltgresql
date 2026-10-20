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

	"github.com/dolthub/doltgresql/server/extensions/extdef"
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// spatialPredicate returns an implementation of a GEOS spatial relationship test. The function name is the one
// PostGIS reports when the SRIDs differ, and `whenEmpty` is the result when either geometry is empty.
func spatialPredicate(funcName string, whenEmpty bool, test func(g *spatial.GEOS, a spatial.Geometry, b spatial.Geometry) (bool, error)) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b := args[0].(Geometry), args[1].(Geometry)
		geosA, geosB, err := geosPair(funcName, a, b)
		if err != nil {
			return nil, err
		}
		if a.Shape.isEmpty() || b.Shape.isEmpty() {
			return whenEmpty, nil
		}
		g := spatial.NewGEOS()
		defer g.Close()
		result, err := test(g, geosA, geosB)
		if err != nil {
			return nil, geosError("GEOS"+funcName, err)
		}
		return result, nil
	}
}

// stEquals implements ST_Equals, which returns whether two geometries are topologically equal. Two empty geometries
// are equal, and an empty geometry equals nothing else.
func stEquals(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(Geometry), args[1].(Geometry)
	if a.Shape.isEmpty() || b.Shape.isEmpty() {
		if err := checkSameSRID("ST_Equals", a, b); err != nil {
			return nil, err
		}
		return a.Shape.isEmpty() && b.Shape.isEmpty(), nil
	}
	return spatialPredicate("ST_Equals", false, (*spatial.GEOS).Equals)(ctx, a, b)
}

// stOrderingEquals implements ST_OrderingEquals, which returns whether two geometries are stored identically.
func stOrderingEquals(ctx *sql.Context, args ...any) (any, error) {
	return compareGeometries(args[0].(Geometry), args[1].(Geometry), false) == 0, nil
}

// stRelate implements ST_Relate without a pattern, which returns the DE-9IM intersection matrix of two geometries
// under the optional boundary node rule.
func stRelate(ctx *sql.Context, args ...any) (any, error) {
	a, b, err := geosPair("relate_full", args[0].(Geometry), args[1].(Geometry))
	if err != nil {
		return nil, err
	}
	boundaryNodeRule := 1
	if len(args) > 2 {
		boundaryNodeRule = int(args[2].(int32))
	}
	g := spatial.NewGEOS()
	defer g.Close()
	matrix, err := g.Relate(a, b, boundaryNodeRule)
	if err != nil {
		return nil, geosError("GEOSRelate", err)
	}
	return matrix, nil
}

// stRelatePattern implements ST_Relate with a pattern, which returns whether the DE-9IM intersection matrix of two
// geometries matches the pattern. A malformed pattern matches nothing.
func stRelatePattern(ctx *sql.Context, args ...any) (any, error) {
	a, b, err := geosPair("relate_pattern", args[0].(Geometry), args[1].(Geometry))
	if err != nil {
		return nil, err
	}
	g := spatial.NewGEOS()
	defer g.Close()
	matrix, err := g.Relate(a, b, 1)
	if err != nil {
		return nil, geosError("GEOSRelate", err)
	}
	matches, err := g.RelatePatternMatch(matrix, args[2].(string))
	return err == nil && matches, nil
}

// stRelateMatch implements ST_RelateMatch, which returns whether an intersection matrix matches a pattern. A
// malformed matrix or pattern matches nothing.
func stRelateMatch(ctx *sql.Context, args ...any) (any, error) {
	g := spatial.NewGEOS()
	defer g.Close()
	matches, err := g.RelatePatternMatch(args[0].(string), args[1].(string))
	return err == nil && matches, nil
}

// predicateRoutines returns the routines that test spatial relationships.
func predicateRoutines() []extdef.Routine {
	geomParams := []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}
	intersectsImpl := spatialPredicate("ST_Intersects", false, (*spatial.GEOS).Intersects)
	coversImpl := spatialPredicate("covers", false, (*spatial.GEOS).Covers)
	coveredByImpl := spatialPredicate("coveredby", false, (*spatial.GEOS).CoveredBy)
	containsImpl := spatialPredicate("contains", false, (*spatial.GEOS).Contains)
	return []extdef.Routine{
		{Name: "_st_contains", Symbol: "_st_contains_geometry_geometry", Parameters: geomParams, Returns: "bool", Strict: true, Impl: containsImpl},
		{Name: "_st_containsproperly", Symbol: "_st_containsproperly_geometry_geometry", Parameters: geomParams, Returns: "bool", Strict: true, Impl: spatialPredicate("containsproperly", false, (*spatial.GEOS).ContainsProperly)},
		{Name: "_st_coveredby", Symbol: "_st_coveredby_geometry_geometry", Parameters: geomParams, Returns: "bool", Strict: true, Impl: coveredByImpl},
		{Name: "_st_covers", Symbol: "_st_covers_geometry_geometry", Parameters: geomParams, Returns: "bool", Strict: true, Impl: coversImpl},
		{Name: "_st_crosses", Symbol: "_st_crosses_geometry_geometry", Parameters: geomParams, Returns: "bool", Strict: true, Impl: spatialPredicate("crosses", false, (*spatial.GEOS).Crosses)},
		{Name: "_st_equals", Symbol: "_st_equals_geometry_geometry", Parameters: geomParams, Returns: "bool", Strict: true, Impl: stEquals},
		{Name: "_st_intersects", Symbol: "_st_intersects_geometry_geometry", Parameters: geomParams, Returns: "bool", Strict: true, Impl: intersectsImpl},
		{Name: "_st_orderingequals", Symbol: "_st_orderingequals_geometry_geometry", Parameters: geomParams, Returns: "bool", Strict: true, Impl: stOrderingEquals},
		{Name: "_st_overlaps", Symbol: "_st_overlaps_geometry_geometry", Parameters: geomParams, Returns: "bool", Strict: true, Impl: spatialPredicate("overlaps", false, (*spatial.GEOS).Overlaps)},
		{Name: "_st_touches", Symbol: "_st_touches_geometry_geometry", Parameters: geomParams, Returns: "bool", Strict: true, Impl: spatialPredicate("touches", false, (*spatial.GEOS).Touches)},
		{Name: "_st_within", Symbol: "_st_within_geometry_geometry", Parameters: geomParams, Returns: "bool", Impl: func(ctx *sql.Context, args ...any) (any, error) {
			if args[0] == nil || args[1] == nil {
				return nil, nil
			}
			return containsImpl(ctx, args[1], args[0])
		}},
		{Name: "equals", Symbol: "equals_geometry_geometry", Parameters: geomParams, Returns: "bool", Strict: true, Impl: stEquals},
		{Name: "st_contains", Symbol: "contains", Parameters: geomParams, Returns: "bool", Strict: true, Impl: containsImpl},
		{Name: "st_containsproperly", Symbol: "containsproperly", Parameters: geomParams, Returns: "bool", Strict: true, Impl: spatialPredicate("containsproperly", false, (*spatial.GEOS).ContainsProperly)},
		{Name: "st_coveredby", Symbol: "coveredby", Parameters: geomParams, Returns: "bool", Strict: true, Impl: coveredByImpl},
		{Name: "st_coveredby", Symbol: "st_coveredby_text_text", Parameters: params("text", "text"), Returns: "bool", Impl: textArgsAsGeometry(2, coveredByImpl)},
		{Name: "st_covers", Symbol: "covers", Parameters: geomParams, Returns: "bool", Strict: true, Impl: coversImpl},
		{Name: "st_covers", Symbol: "st_covers_text_text", Parameters: params("text", "text"), Returns: "bool", Impl: textArgsAsGeometry(2, coversImpl)},
		{Name: "st_crosses", Symbol: "crosses", Parameters: geomParams, Returns: "bool", Strict: true, Impl: spatialPredicate("crosses", false, (*spatial.GEOS).Crosses)},
		{Name: "st_disjoint", Symbol: "disjoint", Parameters: geomParams, Returns: "bool", Strict: true, Impl: spatialPredicate("disjoint", true, (*spatial.GEOS).Disjoint)},
		{Name: "st_equals", Symbol: "ST_Equals", Parameters: geomParams, Returns: "bool", Strict: true, Impl: stEquals},
		{Name: "st_intersects", Symbol: "ST_Intersects", Parameters: geomParams, Returns: "bool", Strict: true, Impl: intersectsImpl},
		{Name: "st_intersects", Symbol: "st_intersects_text_text", Parameters: params("text", "text"), Returns: "bool", Impl: textArgsAsGeometry(2, intersectsImpl)},
		{Name: "st_orderingequals", Symbol: "LWGEOM_same", Parameters: geomParams, Returns: "bool", Strict: true, Impl: stOrderingEquals},
		{Name: "st_overlaps", Symbol: "overlaps", Parameters: geomParams, Returns: "bool", Strict: true, Impl: spatialPredicate("overlaps", false, (*spatial.GEOS).Overlaps)},
		{Name: "st_relate", Symbol: "relate_full", Parameters: geomParams, Returns: "text", Strict: true, Impl: stRelate},
		{Name: "st_relate", Symbol: "relate_full_int4", Parameters: append(geomParams, extdef.Parameter{Type: "int4"}), Returns: "text", Strict: true, Impl: stRelate},
		{Name: "st_relate", Symbol: "relate_pattern", Parameters: append(geomParams, extdef.Parameter{Type: "text"}), Returns: "bool", Strict: true, Impl: stRelatePattern},
		{Name: "st_relatematch", Symbol: "ST_RelateMatch", Parameters: params("text", "text"), Returns: "bool", Strict: true, Impl: stRelateMatch},
		{Name: "st_touches", Symbol: "touches", Parameters: geomParams, Returns: "bool", Strict: true, Impl: spatialPredicate("touches", false, (*spatial.GEOS).Touches)},
		{Name: "st_within", Symbol: "within", Parameters: geomParams, Returns: "bool", Strict: true, Impl: spatialPredicate("within", false, (*spatial.GEOS).Within)},
	}
}
