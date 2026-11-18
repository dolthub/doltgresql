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
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Extension returns the definition of the emulated extension.
func Extension() *extdef.Extension {
	return &extdef.Extension{
		Name: "postgis",
		Control: extdef.Control{
			DefaultVersion: "3.6.4",
			Comment:        "PostGIS geometry and geography spatial types and functions",
			Superuser:      true,
		},
		Types: []extdef.Type{
			spatialType("geometry", "LWGEOM_in", "LWGEOM_out", "LWGEOM_recv", "LWGEOM_send", "geometry_typmod_in", "geometry_typmod_out_int4", "lwgeom_cmp"),
			spatialType("geography", "geography_in", "geography_out", "geography_recv", "geography_send", "geography_typmod_in", "geography_typmod_out_int4", "geography_cmp"),
			fixedLengthType("box2d", 65, pgtypes.TypeAlignment_Int, "BOX2D_in", "BOX2D_out", box2dCodec),
			fixedLengthType("box3d", 52, pgtypes.TypeAlignment_Double, "BOX3D_in", "BOX3D_out", box3dCodec),
			fixedLengthType("spheroid", 65, pgtypes.TypeAlignment_Double, "ellipsoid_in", "ellipsoid_out", spheroidCodec),
		},
		Composites: composites(),
		Routines:   routines(),
		Operators:  operators(),
		Casts:      append(ioCasts(), boxCasts()...),
		Aggregates: append(append(append(aggregates(), mvtAggregates()...), geobufAggregates()...), fgbAggregates()...),
		Tables:     []extdef.Table{spatialRefSysTable()},
		Views:      columnViews(),
	}
}

// spatialType declares the geometry or geography type, given the symbols of its support functions.
func spatialType(name string, in string, out string, recv string, send string, typmodIn string, typmodOut string, compare string) extdef.Type {
	def := pgtypes.NewBaseTypeDefinition()
	def.Align = pgtypes.TypeAlignment_Double
	def.Storage = pgtypes.TypeStorage_Main
	def.Delimiter = ":"
	return extdef.Type{
		Name:       name,
		Definition: def,
		Input:      in,
		Output:     out,
		Receive:    recv,
		Send:       send,
		ModIn:      typmodIn,
		ModOut:     typmodOut,
		Compare:    compare,
	}
}

// params builds an unnamed parameter list from the given type names.
func params(typeNames ...string) []extdef.Parameter {
	ps := make([]extdef.Parameter, len(typeNames))
	for i, typeName := range typeNames {
		ps[i] = extdef.Parameter{Type: typeName}
	}
	return ps
}

// routines returns every routine that the extension declares, each immutable unless routineVolatility says otherwise.
func routines() []extdef.Routine {
	var routines []extdef.Routine
	routines = append(routines, ioRoutines()...)
	routines = append(routines, outputRoutines()...)
	routines = append(routines, constructorRoutines()...)
	routines = append(routines, accessorRoutines()...)
	routines = append(routines, comparisonRoutines()...)
	routines = append(routines, boxRoutines()...)
	routines = append(routines, bboxRoutines()...)
	routines = append(routines, measureRoutines()...)
	routines = append(routines, predicateRoutines()...)
	routines = append(routines, structureRoutines()...)
	routines = append(routines, strokeRoutines()...)
	routines = append(routines, processingRoutines()...)
	routines = append(routines, editorRoutines()...)
	routines = append(routines, coercionRoutines()...)
	routines = append(routines, vertexRoutines()...)
	routines = append(routines, affineRoutines()...)
	routines = append(routines, linrefRoutines()...)
	routines = append(routines, measuredRoutines()...)
	routines = append(routines, geojsonRoutines()...)
	routines = append(routines, formatsRoutines()...)
	routines = append(routines, aggregateRoutines()...)
	routines = append(routines, dumpRoutines()...)
	routines = append(routines, miscRoutines()...)
	routines = append(routines, typedConstructorRoutines()...)
	routines = append(routines, closestRoutines()...)
	routines = append(routines, geodesyRoutines()...)
	routines = append(routines, gridRoutines()...)
	routines = append(routines, reshapeRoutines()...)
	routines = append(routines, trajectoryRoutines()...)
	routines = append(routines, spheroidRoutines()...)
	routines = append(routines, gmlRoutines()...)
	routines = append(routines, x3dRoutines()...)
	routines = append(routines, kmlRoutines()...)
	routines = append(routines, transformRoutines()...)
	routines = append(routines, constructionRoutines()...)
	routines = append(routines, splitRoutines()...)
	routines = append(routines, subdivideRoutines()...)
	routines = append(routines, srsRoutines()...)
	routines = append(routines, managementRoutines()...)
	routines = append(routines, clusterRoutines()...)
	routines = append(routines, distance3DRoutines()...)
	routines = append(routines, coverageRoutines()...)
	routines = append(routines, viewRoutines()...)
	routines = append(routines, marc21Routines()...)
	routines = append(routines, mvtRoutines()...)
	routines = append(routines, geobufRoutines()...)
	routines = append(routines, fgbRoutines()...)
	routines = append(routines, lettersRoutines()...)
	routines = append(routines, enclosingRoutines()...)
	for i := range routines {
		routines[i].Volatility = tree.VolatilityImmutable
		if volatility, ok := routineVolatility[routines[i].Symbol]; ok {
			routines[i].Volatility = volatility
		}
	}
	return routines
}

// routineVolatility holds the volatility of each routine that is not immutable, by its symbol.
var routineVolatility = map[string]tree.Volatility{
	"ST_AsGeoJsonRow":        tree.VolatilityStable,
	"ST_ClusterKMeans":       tree.VolatilityVolatile,
	"ST_GeneratePoints":      tree.VolatilityVolatile,
	"_postgis_pgsql_version": tree.VolatilityStable,
	"addgeometrycolumn_varchar_varchar_int4_varchar_int4_bool":                 tree.VolatilityVolatile,
	"addgeometrycolumn_varchar_varchar_varchar_int4_varchar_int4_bool":         tree.VolatilityStable,
	"addgeometrycolumn_varchar_varchar_varchar_varchar_int4_varchar_int4_bool": tree.VolatilityVolatile,
	"cache_bbox":                                              tree.VolatilityVolatile,
	"dropgeometrycolumn_varchar_varchar":                      tree.VolatilityVolatile,
	"dropgeometrycolumn_varchar_varchar_varchar":              tree.VolatilityVolatile,
	"dropgeometrycolumn_varchar_varchar_varchar_varchar":      tree.VolatilityVolatile,
	"dropgeometrytable_varchar":                               tree.VolatilityVolatile,
	"dropgeometrytable_varchar_varchar":                       tree.VolatilityVolatile,
	"dropgeometrytable_varchar_varchar_varchar":               tree.VolatilityVolatile,
	"find_srid_varchar_varchar_varchar":                       tree.VolatilityStable,
	"gserialized_estimated_extent":                            tree.VolatilityStable,
	"pgis_geometry_accum_transfn":                             tree.VolatilityVolatile,
	"pgis_geometry_accum_transfn_float8":                      tree.VolatilityVolatile,
	"pgis_geometry_clusterintersecting_finalfn":               tree.VolatilityVolatile,
	"pgis_geometry_clusterwithin_finalfn":                     tree.VolatilityVolatile,
	"pgis_geometry_collect_finalfn":                           tree.VolatilityVolatile,
	"pgis_geometry_coverageunion_finalfn":                     tree.VolatilityVolatile,
	"pgis_geometry_makeline_finalfn":                          tree.VolatilityVolatile,
	"pgis_geometry_polygonize_finalfn":                        tree.VolatilityVolatile,
	"populate_geometry_columns_bool":                          tree.VolatilityVolatile,
	"populate_geometry_columns_oid_bool":                      tree.VolatilityVolatile,
	"postgis_constraint_dims":                                 tree.VolatilityStable,
	"postgis_constraint_srid":                                 tree.VolatilityStable,
	"postgis_constraint_type":                                 tree.VolatilityStable,
	"postgis_extensions_upgrade":                              tree.VolatilityVolatile,
	"st_estimatedextent_text_text":                            tree.VolatilityStable,
	"st_estimatedextent_text_text_text":                       tree.VolatilityStable,
	"st_findextent_text_text":                                 tree.VolatilityStable,
	"st_findextent_text_text_text":                            tree.VolatilityStable,
	"st_symmetricdifference_geometry_geometry":                tree.VolatilityVolatile,
	"updategeometrysrid_varchar_varchar_int4":                 tree.VolatilityVolatile,
	"updategeometrysrid_varchar_varchar_varchar_int4":         tree.VolatilityVolatile,
	"updategeometrysrid_varchar_varchar_varchar_varchar_int4": tree.VolatilityVolatile,
}

// operators returns every operator that the extension declares.
func operators() []extdef.Operator {
	return append(comparisonOperators(), bboxOperators()...)
}
