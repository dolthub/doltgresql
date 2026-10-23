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
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// accumulation is the internal state of the aggregates that gather their geometries before combining them: the
// nonnull geometries in input order and the aggregate's extra argument, such as a grid size or a distance, which is
// -1 when the aggregate has none.
type accumulation struct {
	geometries []any
	extra      float64
}

// accumulate implements the transition functions that gather geometries, skipping nulls.
func accumulate(ctx *sql.Context, args ...any) (any, error) {
	state, _ := args[0].(*accumulation)
	if state == nil {
		state = &accumulation{extra: -1}
	}
	if len(args) > 2 && args[2] != nil {
		state.extra = args[2].(float64)
	}
	if args[1] != nil {
		state.geometries = append(state.geometries, args[1])
	}
	return state, nil
}

// combineAccumulations implements the combine function of the aggregates that gather their geometries, appending the
// second state's geometries to the first's.
func combineAccumulations(ctx *sql.Context, args ...any) (any, error) {
	first, _ := args[0].(*accumulation)
	second, _ := args[1].(*accumulation)
	if first == nil {
		return second, nil
	}
	if second != nil {
		first.geometries = append(first.geometries, second.geometries...)
	}
	return first, nil
}

// finalImpl returns a final function that hands the gathered geometries to an implementation that takes an array,
// returning null when nothing was gathered.
func finalImpl(impl extdef.Function) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		state, _ := args[0].(*accumulation)
		if state == nil || len(state.geometries) == 0 {
			return nil, nil
		}
		return impl(ctx, state.geometries, state.extra)
	}
}

// unionAll returns the union of the gathered geometries with the first geometry's SRID, snapped to the grid size
// unless it is negative. Its GEOS counterpart does not check the SRIDs. An empty union is an empty point, linestring,
// or polygon, matching the highest dimension of the inputs.
func unionAll(ctx *sql.Context, args ...any) (any, error) {
	return combineAll("GEOSUnaryUnion", func(g *spatial.GEOS, collection spatial.Geometry) (spatial.Geometry, error) {
		return g.UnaryUnion(collection, args[1].(float64))
	})(ctx, args...)
}

// coverageUnionAll implements ST_CoverageUnion, which returns the union of polygons that may only share edges, in two
// dimensions and without an SRID. It skips null elements, and returns null when the union is empty.
func coverageUnionAll(ctx *sql.Context, args ...any) (any, error) {
	collection := shape{typeNum: collectionType}
	for _, elem := range args[0].([]any) {
		if g, ok := elem.(Geometry); ok {
			collection.members = append(collection.members, forceDims(g.Shape, dimXY, 0, 0))
		}
	}
	geosCollection, err := collection.toGEOS()
	if err != nil {
		return nil, err
	}
	g := spatial.NewGEOS()
	defer g.Close()
	union, err := g.CoverageUnion(geosCollection)
	if err != nil {
		return nil, pgerror.Newf(pgcode.Internal, "Error computing coverage union: %s", err.Error())
	}
	result := fromGEOS(union, sridUnknown, false)
	if result.Shape.isEmpty() {
		return nil, nil
	}
	return result, nil
}

// combineAll returns an implementation that combines the gathered geometries into a collection and hands it to a GEOS
// operation, as described in unionAll.
func combineAll(function string, op func(g *spatial.GEOS, collection spatial.Geometry) (spatial.Geometry, error)) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		geometries := args[0].([]any)
		first := geometries[0].(Geometry)
		collection := shape{typeNum: collectionType, ctype: first.Shape.ctype}
		hasZ := false
		for _, elem := range geometries {
			g := elem.(Geometry)
			hasZ = hasZ || g.Shape.hasZ()
			collection.members = append(collection.members, g.Shape.withDims(g.Shape.ctype&dimXYZ))
		}
		collection.ctype &= dimXYZ
		for i := range collection.members {
			if collection.members[i].ctype != collection.ctype {
				collection.members[i] = forceDims(collection.members[i], collection.ctype, 0, 0)
			}
		}
		geosCollection, err := collection.toGEOS()
		if err != nil {
			return nil, err
		}
		g := spatial.NewGEOS()
		defer g.Close()
		result, err := op(g, geosCollection)
		if err != nil {
			return nil, geosError(function, err)
		}
		combined := fromGEOS(result, first.SRID, hasZ)
		if combined.Shape.isEmpty() {
			emptyTypes := []int{pointType, lineType, polygonType}
			combined.Shape = shape{typeNum: emptyTypes[min(collection.dimension(), 2)], ctype: combined.Shape.ctype}
		}
		return combined, nil
	}
}

// clusters groups the gathered geometries into clusters, where two geometries share a cluster when `isNear` holds
// for them directly or through other geometries. Each cluster becomes a geometry collection, in the order in which
// its first geometry was gathered.
func clusters(geometries []any, isNear func(group *spatial.Group, i int, j int) (bool, error)) ([]any, error) {
	converted := make([]spatial.Geometry, len(geometries))
	for i, elem := range geometries {
		geosGeometry, err := elem.(Geometry).Shape.toGEOS()
		if err != nil {
			return nil, err
		}
		converted[i] = geosGeometry
	}
	g := spatial.NewGEOS()
	defer g.Close()
	group, err := g.NewGroup(converted)
	if err != nil {
		return nil, geosError("LWGEOM2GEOS", err)
	}
	defer group.Close()
	parent := make([]int, len(geometries))
	for i := range parent {
		parent[i] = i
	}
	find := func(i int) int {
		for parent[i] != i {
			parent[i] = parent[parent[i]]
			i = parent[i]
		}
		return i
	}
	for i := range converted {
		for j := i + 1; j < len(converted); j++ {
			if find(i) == find(j) {
				continue
			}
			near, err := isNear(group, i, j)
			if err != nil {
				return nil, pgerror.New(pgcode.Internal, err.Error())
			}
			if near {
				parent[find(j)] = find(i)
			}
		}
	}
	first := geometries[0].(Geometry)
	var order []int
	members := make(map[int][]shape)
	for i, elem := range geometries {
		root := find(i)
		if _, ok := members[root]; !ok {
			order = append(order, root)
		}
		members[root] = append(members[root], elem.(Geometry).Shape)
	}
	result := make([]any, len(order))
	for i, root := range order {
		result[i] = Geometry{SRID: first.SRID, Shape: shape{typeNum: collectionType, ctype: first.Shape.ctype, members: members[root]}}
	}
	return result, nil
}

// clusterIntersecting implements ST_ClusterIntersecting over an array, which clusters geometries that intersect.
func clusterIntersecting(ctx *sql.Context, args ...any) (any, error) {
	return clusters(args[0].([]any), (*spatial.Group).Intersects)
}

// clusterWithin implements ST_ClusterWithin over an array, which clusters geometries within the distance.
func clusterWithin(ctx *sql.Context, args ...any) (any, error) {
	distance := args[1].(float64)
	return clusters(args[0].([]any), func(group *spatial.Group, i int, j int) (bool, error) {
		d, err := group.Distance(i, j)
		return d <= distance, err
	})
}

// aggregateRoutines returns the transition and final functions of the aggregates, and the array functions that share
// their final steps.
func aggregateRoutines() []extdef.Routine {
	accumParams := func(extra ...string) []extdef.Parameter {
		return params(append([]string{"internal", "geometry"}, extra...)...)
	}
	arrayImpl := func(impl extdef.Function) extdef.Function {
		return func(ctx *sql.Context, args ...any) (any, error) {
			var geometries []any
			for _, elem := range args[0].([]any) {
				if elem != nil {
					geometries = append(geometries, elem)
				}
			}
			if len(geometries) == 0 {
				return nil, nil
			}
			extra := -1.0
			if len(args) > 1 {
				extra = args[1].(float64)
			}
			return impl(ctx, geometries, extra)
		}
	}
	arrayOnly := func(impl extdef.Function) extdef.Function {
		return func(ctx *sql.Context, args ...any) (any, error) {
			return impl(ctx, args[0])
		}
	}
	return []extdef.Routine{
		{Name: "pgis_geometry_accum_transfn", Symbol: "pgis_geometry_accum_transfn", Parameters: accumParams(), Returns: "internal", Impl: accumulate},
		{Name: "pgis_geometry_accum_transfn", Symbol: "pgis_geometry_accum_transfn_float8", Parameters: accumParams("float8"), Returns: "internal", Impl: accumulate},
		{Name: "pgis_geometry_union_parallel_transfn", Symbol: "pgis_geometry_union_parallel_transfn", Parameters: accumParams(), Returns: "internal", Impl: accumulate},
		{Name: "pgis_geometry_union_parallel_transfn", Symbol: "pgis_geometry_union_parallel_transfn_float8", Parameters: accumParams("float8"), Returns: "internal", Impl: accumulate},
		{Name: "pgis_geometry_union_parallel_combinefn", Symbol: "pgis_geometry_union_parallel_combinefn", Parameters: params("internal", "internal"), Returns: "internal", Impl: combineAccumulations},
		{Name: "pgis_geometry_union_parallel_finalfn", Symbol: "pgis_geometry_union_parallel_finalfn", Parameters: params("internal"), Returns: "geometry", Impl: finalImpl(unionAll)},
		{Name: "pgis_geometry_collect_finalfn", Symbol: "pgis_geometry_collect_finalfn", Parameters: params("internal"), Returns: "geometry", Impl: finalImpl(arrayOnly(stCollectArray))},
		{Name: "pgis_geometry_makeline_finalfn", Symbol: "pgis_geometry_makeline_finalfn", Parameters: params("internal"), Returns: "geometry", Impl: finalImpl(arrayOnly(stMakeLineArray))},
		{Name: "pgis_geometry_clusterintersecting_finalfn", Symbol: "pgis_geometry_clusterintersecting_finalfn", Parameters: params("internal"), Returns: "_geometry", Impl: finalImpl(clusterIntersecting)},
		{Name: "pgis_geometry_clusterwithin_finalfn", Symbol: "pgis_geometry_clusterwithin_finalfn", Parameters: params("internal"), Returns: "_geometry", Impl: finalImpl(clusterWithin)},
		{Name: "pgis_geometry_coverageunion_finalfn", Symbol: "pgis_geometry_coverageunion_finalfn", Parameters: params("internal"), Returns: "geometry", Impl: finalImpl(coverageUnionAll)},
		{Name: "st_union", Symbol: "pgis_union_geometry_array", Parameters: params("_geometry"), Returns: "geometry", Strict: true, Impl: arrayImpl(unionAll)},
		{Name: "st_coverageunion", Symbol: "ST_CoverageUnion", Parameters: params("_geometry"), Returns: "geometry", Strict: true, Impl: coverageUnionAll},
		{Name: "st_clusterintersecting", Symbol: "clusterintersecting_garray", Parameters: params("_geometry"), Returns: "_geometry", Strict: true, Impl: arrayImpl(clusterIntersecting)},
		{Name: "st_clusterwithin", Symbol: "cluster_within_distance_garray", Parameters: params("_geometry", "float8"), Returns: "_geometry", Strict: true, Impl: arrayImpl(clusterWithin)},
	}
}

// aggregates returns the aggregates that the extension declares.
func aggregates() []extdef.Aggregate {
	return []extdef.Aggregate{
		{Name: "st_extent", Parameters: params("geometry"), Returns: "box2d", StateType: "box3d", Transition: "BOX3D_combine", Final: "BOX3D_to_BOX2D", Combine: "BOX3D_combine_BOX3D"},
		{Name: "st_3dextent", Parameters: params("geometry"), Returns: "box3d", StateType: "box3d", Transition: "BOX3D_combine", Combine: "BOX3D_combine_BOX3D"},
		{Name: "st_memcollect", Parameters: params("geometry"), Returns: "geometry", StateType: "geometry", Transition: "LWGEOM_collect", Combine: "LWGEOM_collect"},
		{Name: "st_memunion", Parameters: params("geometry"), Returns: "geometry", StateType: "geometry", Transition: "ST_Union", Combine: "ST_Union"},
		{Name: "st_coverageunion", Parameters: params("geometry"), Returns: "geometry", StateType: "internal", Transition: "pgis_geometry_accum_transfn", Final: "pgis_geometry_coverageunion_finalfn"},
		{Name: "st_union", Parameters: params("geometry"), Returns: "geometry", StateType: "internal", Transition: "pgis_geometry_union_parallel_transfn", Final: "pgis_geometry_union_parallel_finalfn", Combine: "pgis_geometry_union_parallel_combinefn"},
		{Name: "st_union", Parameters: []extdef.Parameter{{Type: "geometry"}, {Name: "gridsize", Type: "float8"}}, Returns: "geometry", StateType: "internal", Transition: "pgis_geometry_union_parallel_transfn_float8", Final: "pgis_geometry_union_parallel_finalfn", Combine: "pgis_geometry_union_parallel_combinefn"},
		{Name: "st_collect", Parameters: params("geometry"), Returns: "geometry", StateType: "internal", Transition: "pgis_geometry_accum_transfn", Final: "pgis_geometry_collect_finalfn"},
		{Name: "st_makeline", Parameters: params("geometry"), Returns: "geometry", StateType: "internal", Transition: "pgis_geometry_accum_transfn", Final: "pgis_geometry_makeline_finalfn"},
		{Name: "st_clusterintersecting", Parameters: params("geometry"), Returns: "_geometry", StateType: "internal", Transition: "pgis_geometry_accum_transfn", Final: "pgis_geometry_clusterintersecting_finalfn"},
		{Name: "st_polygonize", Parameters: params("geometry"), Returns: "geometry", StateType: "internal", Transition: "pgis_geometry_accum_transfn", Final: "pgis_geometry_polygonize_finalfn"},
		{Name: "st_clusterwithin", Parameters: params("geometry", "float8"), Returns: "_geometry", StateType: "internal", Transition: "pgis_geometry_accum_transfn_float8", Final: "pgis_geometry_clusterwithin_finalfn"},
	}
}
