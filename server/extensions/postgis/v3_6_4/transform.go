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
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// spatialRefSysTable declares the spatial_ref_sys table, which starts with the coordinate reference systems in PROJ's
// database. EPSG systems take their own codes as SRIDs, and so do ESRI systems whose codes EPSG does not use.
func spatialRefSysTable() extdef.Table {
	return extdef.Table{
		Name: "spatial_ref_sys",
		Definition: `(srid integer NOT NULL, auth_name varchar(256), auth_srid integer, srtext varchar(2048), ` +
			`proj4text varchar(2048), PRIMARY KEY (srid), ` +
			`CONSTRAINT spatial_ref_sys_srid_check CHECK (srid > 0 AND srid <= 998999))`,
		Rows: spatialRefSysRows,
	}
}

// columnViews declares the geometry_columns and geography_columns views, which describe every geometry and geography
// column from its type modifier.
func columnViews() []extdef.View {
	return []extdef.View{
		{Name: "geometry_columns", Query: `SELECT current_database()::character varying(256) AS f_table_catalog, n.nspname AS f_table_schema, ` +
			`c.relname AS f_table_name, a.attname AS f_geometry_column, COALESCE(postgis_typmod_dims(a.atttypmod), 2) AS coord_dimension, ` +
			`postgis_typmod_srid(a.atttypmod) AS srid, replace(replace(COALESCE(NULLIF(upper(postgis_typmod_type(a.atttypmod)), 'GEOMETRY'), ` +
			`'GEOMETRY'), 'ZM', ''), 'Z', '')::character varying(30) AS type FROM pg_class c JOIN pg_attribute a ON a.attrelid = c.oid AND ` +
			`NOT a.attisdropped JOIN pg_namespace n ON c.relnamespace = n.oid JOIN pg_type t ON a.atttypid = t.oid WHERE c.relkind IN ` +
			`('r', 'v', 'm', 'f', 'p') AND c.relname <> 'raster_columns' AND t.typname = 'geometry'`},
		{Name: "geography_columns", Query: `SELECT current_database() AS f_table_catalog, n.nspname AS f_table_schema, c.relname AS f_table_name, ` +
			`a.attname AS f_geography_column, postgis_typmod_dims(a.atttypmod) AS coord_dimension, postgis_typmod_srid(a.atttypmod) AS srid, ` +
			`postgis_typmod_type(a.atttypmod) AS type FROM pg_class c, pg_attribute a, pg_type t, pg_namespace n WHERE t.typname = 'geography' ` +
			`AND a.attisdropped = false AND a.atttypid = t.oid AND a.attrelid = c.oid AND c.relnamespace = n.oid AND c.relkind IN ` +
			`('r', 'v', 'm', 'f', 'p')`},
	}
}

// spatialRefSysRows returns the starting rows of spatial_ref_sys.
func spatialRefSysRows(ctx *sql.Context) ([][]any, error) {
	return databaseSRSRows()
}

// databaseSRSRows reads the starting rows of spatial_ref_sys from PROJ's database, which never changes, so it is only
// read once.
var databaseSRSRows = sync.OnceValues(func() ([][]any, error) {
	var rows [][]any
	taken := make(map[int64]bool)
	for _, authority := range []string{"EPSG", "ESRI"} {
		infos, err := spatial.DatabaseCRSInfo(authority, nil)
		if err != nil {
			return nil, err
		}
		for _, info := range infos {
			code, err := strconv.ParseInt(info.Code, 10, 32)
			if err != nil || info.Kind == spatial.OtherCRS || taken[code] || code > sridUserMaximum {
				continue
			}
			taken[code] = true
			wkt, projString, err := spatial.CRSDefinitions(info.Authority, info.Code)
			if err != nil {
				return nil, err
			}
			rows = append(rows, []any{int32(code), info.Authority, int32(code), wkt, strings.TrimSuffix(projString, " +type=crs")})
		}
	}
	return rows, nil
})

// databaseSRS describes how spatial_ref_sys starts out defining an SRID: the authority that files it, and whether it
// is a geographic system.
type databaseSRS struct {
	authority  string
	geographic bool
}

// databaseSRSCache holds the systems that databaseSRSFor has found, keyed by SRID.
var databaseSRSCache sync.Map

// databaseSRSFor returns how spatial_ref_sys starts out defining an SRID, reading PROJ's database the way the table is
// filled: EPSG files the SRID when it knows the code, and ESRI otherwise. It returns false when neither does.
func databaseSRSFor(srid int32) (databaseSRS, bool) {
	if cached, ok := databaseSRSCache.Load(srid); ok {
		found := cached.(databaseSRS)
		return found, found.authority != ""
	}
	var found databaseSRS
	for _, authority := range []string{"EPSG", "ESRI"} {
		if definition := fmt.Sprintf("%s:%d", authority, srid); spatial.Reads(definition) {
			found = databaseSRS{authority: authority, geographic: spatial.IsGeographic(definition)}
			break
		}
	}
	databaseSRSCache.Store(srid, found)
	return found, found.authority != ""
}

// srsKey identifies an SRID looked up during one query, since spatial_ref_sys is read once per query as PostGIS does.
type srsKey struct {
	pid  uint64
	srid int32
}

// srsDefinitions holds the ways that spatial_ref_sys defines an SRID for PROJ: the authority and code, the WKT, and
// the PROJ string. Each is empty when the row lacks it.
type srsDefinitions [3]string

// srsCache holds the definitions of the SRIDs that recent queries looked up, which is cleared once it grows large.
var srsCache = struct {
	sync.Mutex
	definitions map[srsKey]srsDefinitions
}{definitions: make(map[srsKey]srsDefinitions)}

// queryRows runs a query on the runner and returns its rows.
func queryRows(ctx *sql.Context, runner sql.StatementRunner, query string) ([]sql.Row, error) {
	return sql.RunInterpreted(ctx, func(subCtx *sql.Context) ([]sql.Row, error) {
		_, rowIter, _, err := runner.QueryWithBindings(subCtx, query, nil, nil, nil)
		if err != nil {
			return nil, err
		}
		return sql.RowIterToRows(subCtx, rowIter)
	})
}

// lookUpSRS returns the definitions of the SRID in spatial_ref_sys.
func lookUpSRS(ctx *sql.Context, runner sql.StatementRunner, srid int32) (srsDefinitions, error) {
	key := srsKey{pid: ctx.Pid(), srid: srid}
	srsCache.Lock()
	definitions, ok := srsCache.definitions[key]
	srsCache.Unlock()
	if ok {
		return definitions, nil
	}
	if projString, ok := reservedSRS(srid); ok {
		return srsDefinitions{2: projString}, nil
	}
	rows, err := queryRows(ctx, runner, "SELECT auth_name, auth_srid, srtext, proj4text FROM spatial_ref_sys WHERE srid = "+strconv.Itoa(int(srid)))
	if err != nil {
		return definitions, err
	}
	if len(rows) == 0 {
		return definitions, pgerror.Newf(pgcode.Internal, "Cannot find SRID (%d) in spatial_ref_sys", srid)
	}
	if rows[0][0] != nil && rows[0][1] != nil {
		definitions[0] = fmt.Sprintf("%s:%d", rows[0][0], rows[0][1])
	}
	for i, text := range rows[0][2:] {
		if text != nil {
			definitions[i+1] = text.(string)
		}
	}
	if definitions == (srsDefinitions{}) {
		return definitions, pgerror.Newf(pgcode.Internal, "got NULL for SRID (%d)", srid)
	}
	srsCache.Lock()
	if len(srsCache.definitions) > 4096 {
		srsCache.definitions = make(map[srsKey]srsDefinitions)
	}
	srsCache.definitions[key] = definitions
	srsCache.Unlock()
	return definitions, nil
}

// transformerPool holds idle transformers between pairs of definitions, which are expensive to create. A transformer
// is taken out of the pool while it is used.
var transformerPool = struct {
	sync.Mutex
	idle map[[2]string][]*spatial.Transformer
}{idle: make(map[[2]string][]*spatial.Transformer)}

// withTransformer runs the function with a transformer between two definitions, returning the error from creating the
// transformer unchanged.
func withTransformer(source string, target string, fn func(t *spatial.Transformer) error) error {
	key := [2]string{source, target}
	transformerPool.Lock()
	var t *spatial.Transformer
	if idle := transformerPool.idle[key]; len(idle) > 0 {
		t, transformerPool.idle[key] = idle[len(idle)-1], idle[:len(idle)-1]
	}
	transformerPool.Unlock()
	if t == nil {
		var err error
		if t, err = spatial.NewTransformer(source, target); err != nil {
			return err
		}
	}
	err := fn(t)
	transformerPool.Lock()
	if len(transformerPool.idle) > 256 {
		for _, idle := range transformerPool.idle {
			for _, unused := range idle {
				unused.Close()
			}
		}
		transformerPool.idle = make(map[[2]string][]*spatial.Transformer)
	}
	transformerPool.idle[key] = append(transformerPool.idle[key], t)
	transformerPool.Unlock()
	return err
}

// transformShape returns the shape with every point transformed, including the control points of curves. Z
// coordinates are transformed as heights, and M coordinates are kept.
func transformShape(s shape, t *spatial.Transformer) (shape, error) {
	dims := s.ctype.Dimension()
	transformSeq := func(seq []float64) ([]float64, error) {
		transformed := append([]float64(nil), seq...)
		return transformed, t.Transform(transformed, dims, s.hasZ())
	}
	var err error
	if hasSequence(s.typeNum) {
		s.coords, err = transformSeq(s.coords)
		return s, err
	}
	if s.typeNum == polygonType {
		rings := make([][]float64, len(s.rings))
		for i, ring := range s.rings {
			if rings[i], err = transformSeq(ring); err != nil {
				return shape{}, err
			}
		}
		s.rings = rings
		return s, nil
	}
	members := make([]shape, len(s.members))
	for i, member := range s.members {
		if members[i], err = transformShape(member, t); err != nil {
			return shape{}, err
		}
	}
	s.members = members
	return s, nil
}

// transformGeometry transforms a geometry between two definitions and gives the result the SRID. It returns false
// when PROJ cannot form a transformation between the definitions.
func transformGeometry(g Geometry, source string, target string, srid int32) (Geometry, bool, error) {
	var result Geometry
	formed := false
	err := withTransformer(source, target, func(t *spatial.Transformer) error {
		formed = true
		transformed, err := transformShape(g.Shape, t)
		var transformErr spatial.TransformError
		if errors.As(err, &transformErr) {
			return pgerror.Newf(pgcode.Internal, "transform: %s (%d)", transformErr.Message, transformErr.Number)
		}
		result = Geometry{SRID: srid, Shape: transformed}
		return err
	})
	if !formed {
		return Geometry{}, false, nil
	}
	return result, true, err
}

// stTransform implements ST_Transform with a target SRID, reading both systems from spatial_ref_sys. It tries their
// authorities and codes, then their WKT, and then their PROJ strings.
func stTransform(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	g, target := args[0].(Geometry), args[1].(int32)
	if target == sridUnknown {
		return nil, pgerror.Newf(pgcode.Internal, "ST_Transform: %d is an invalid target SRID", target)
	}
	if g.SRID == target {
		return g, nil
	}
	if g.SRID == sridUnknown {
		return nil, pgerror.New(pgcode.Internal, "ST_Transform: Input geometry has unknown (0) SRID")
	}
	sourceDefs, err := lookUpSRS(ctx, runner, g.SRID)
	if err != nil {
		return nil, err
	}
	targetDefs, err := lookUpSRS(ctx, runner, target)
	if err != nil {
		return nil, err
	}
	for i := range sourceDefs {
		if sourceDefs[i] == "" || targetDefs[i] == "" {
			continue
		}
		if result, ok, err := transformGeometry(g, sourceDefs[i], targetDefs[i], target); ok || err != nil {
			return result, err
		}
	}
	return nil, pgerror.Newf(pgcode.Internal, "could not form projection (LWPROJ) from 'srid=%d' to 'srid=%d'", g.SRID, target)
}

// proj4Text returns the PROJ string of an SRID in spatial_ref_sys, or false when the SRID has no row.
func proj4Text(ctx *sql.Context, runner sql.StatementRunner, srid int32) (any, bool, error) {
	rows, err := queryRows(ctx, runner, "SELECT proj4text FROM spatial_ref_sys WHERE srid = "+strconv.Itoa(int(srid)))
	if err != nil || len(rows) == 0 {
		return nil, false, err
	}
	return rows[0][0], true, nil
}

// getProj4FromSRID implements get_proj4_from_srid, which returns the PROJ string of an SRID in spatial_ref_sys, or null
// when the SRID has no row.
func getProj4FromSRID(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	projString, _, err := proj4Text(ctx, runner, args[0].(int32))
	return projString, err
}

// postgisTransformGeometry implements postgis_transform_geometry, which transforms a geometry between two definitions
// that PROJ reads and gives the result the SRID.
func postgisTransformGeometry(ctx *sql.Context, args ...any) (any, error) {
	g, source, target, srid := args[0].(Geometry), args[1].(string), args[2].(string), args[3].(int32)
	result, ok, err := transformGeometry(g, source, target, srid)
	if ok || err != nil {
		return result, err
	}
	for _, definition := range []string{source, target} {
		if !spatial.Reads(definition) {
			return nil, pgerror.Newf(pgcode.Internal, "could not parse proj string '%s'", definition)
		}
	}
	return nil, pgerror.New(pgcode.Internal, "lwgeom_transform_from_str: Failed to transform")
}

// stTransformToText implements ST_Transform with a target definition, taking the source from the PROJ string of the
// geometry's SRID. A geometry whose SRID has no row in spatial_ref_sys returns null.
func stTransformToText(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	g := args[0].(Geometry)
	source, ok, err := proj4Text(ctx, runner, g.SRID)
	if err != nil || !ok || source == nil {
		return nil, err
	}
	return postgisTransformGeometry(ctx, g, source, args[1], int32(0))
}

// stTransformFromText implements ST_Transform from a source definition to a target SRID, taking the target from the
// SRID's PROJ string. A target SRID without a row in spatial_ref_sys returns null.
func stTransformFromText(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	target, ok, err := proj4Text(ctx, runner, args[2].(int32))
	if err != nil || !ok || target == nil {
		return nil, err
	}
	return postgisTransformGeometry(ctx, args[0], args[1], target, args[2])
}

// stTransformTextToText implements ST_Transform between two definitions, whose result has an unknown SRID.
func stTransformTextToText(ctx *sql.Context, args ...any) (any, error) {
	return postgisTransformGeometry(ctx, args[0], args[1], args[2], int32(0))
}

// postgisTransformPipelineGeometry implements postgis_transform_pipeline_geometry, which runs a coordinate operation
// on a geometry, forward or in reverse, and gives the result the SRID.
func postgisTransformPipelineGeometry(ctx *sql.Context, args ...any) (any, error) {
	g, pipeline, forward, srid := args[0].(Geometry), args[1].(string), args[2].(bool), args[3].(int32)
	t, err := spatial.NewPipelineTransformer(pipeline, forward)
	if err != nil {
		return nil, pgerror.Newf(pgcode.Internal, "could not parse coordinate operation '%s'", pipeline)
	}
	defer t.Close()
	transformed, err := transformShape(g.Shape, t)
	if err != nil {
		return nil, pgerror.New(pgcode.Internal, "lwgeom_transform_pipeline: Failed to transform")
	}
	return Geometry{SRID: srid, Shape: transformed}, nil
}

// pipelineImpl returns the implementation of ST_TransformPipeline or ST_InverseTransformPipeline.
func pipelineImpl(forward bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		return postgisTransformPipelineGeometry(ctx, args[0], args[1], forward, args[2])
	}
}

// transformRoutines returns the routines that transform geometries between coordinate reference systems.
func transformRoutines() []extdef.Routine {
	pipelineParams := []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "pipeline", Type: "text"}, {Name: "to_srid", Type: "int4", Default: `0`}}
	return []extdef.Routine{
		{Name: "_st_bestsrid", Symbol: "_st_bestsrid_geography", Parameters: params("geography"), Returns: "int4", Strict: true, Impl: stBestSRID},
		{Name: "_st_bestsrid", Symbol: "_st_bestsrid_geography_geography", Parameters: params("geography", "geography"), Returns: "int4", Strict: true, Impl: stBestSRID},
		{Name: "st_buffer", Symbol: "st_buffer_geography_float8", Parameters: params("geography", "float8"), Returns: "geography", Strict: true, QueryImpl: geographyBuffer},
		{Name: "st_buffer", Symbol: "st_buffer_geography_float8_int4", Parameters: params("geography", "float8", "int4"), Returns: "geography", Strict: true, QueryImpl: geographyBuffer},
		{Name: "st_buffer", Symbol: "st_buffer_geography_float8_text", Parameters: params("geography", "float8", "text"), Returns: "geography", Strict: true, QueryImpl: geographyBuffer},
		{Name: "st_intersection", Symbol: "st_intersection_geography_geography", Parameters: params("geography", "geography"), Returns: "geography", Strict: true, QueryImpl: geographyIntersection},
		{Name: "get_proj4_from_srid", Symbol: "get_proj4_from_srid_int4", Parameters: params("int4"), Returns: "text", Strict: true, QueryImpl: getProj4FromSRID},
		{Name: "postgis_transform_geometry", Symbol: "transform_geom", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Type: "text"}, {Type: "text"}, {Type: "int4"}}, Returns: "geometry", Strict: true, Impl: postgisTransformGeometry},
		{Name: "postgis_transform_pipeline_geometry", Symbol: "transform_pipeline_geom", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "pipeline", Type: "text"}, {Name: "forward", Type: "bool"}, {Name: "to_srid", Type: "int4"}}, Returns: "geometry", Strict: true, Impl: postgisTransformPipelineGeometry},
		{Name: "st_inversetransformpipeline", Symbol: "st_inversetransformpipeline_geometry_text_int4", Parameters: pipelineParams, Returns: "geometry", Strict: true, Impl: pipelineImpl(false)},
		{Name: "st_transform", Symbol: "transform", Parameters: params("geometry", "int4"), Returns: "geometry", Strict: true, QueryImpl: stTransform},
		{Name: "st_transform", Symbol: "st_transform_geometry_text_text", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "from_proj", Type: "text"}, {Name: "to_proj", Type: "text"}}, Returns: "geometry", Strict: true, Impl: stTransformTextToText},
		{Name: "st_transform", Symbol: "st_transform_geometry_text_int4", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "from_proj", Type: "text"}, {Name: "to_srid", Type: "int4"}}, Returns: "geometry", Strict: true, QueryImpl: stTransformFromText},
		{Name: "st_transform", Symbol: "st_transform_geometry_text", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "to_proj", Type: "text"}}, Returns: "geometry", Strict: true, QueryImpl: stTransformToText},
		{Name: "st_transformpipeline", Symbol: "st_transformpipeline_geometry_text_int4", Parameters: pipelineParams, Returns: "geometry", Strict: true, Impl: pipelineImpl(true)},
	}
}

// reservedSRS returns the PROJ string of an SRID in the range that PostGIS reserves for the planar systems that
// _ST_BestSRID chooses, which spatial_ref_sys never holds: world Mercator, the UTM zones, the polar Lambert azimuthal
// equal-area systems, and the Lambert azimuthal equal-area systems of each latitude band's zones.
func reservedSRS(srid int32) (string, bool) {
	const wgs84 = " +ellps=WGS84 +datum=WGS84 +units=m +no_defs"
	switch {
	case srid == bestSRIDMercator:
		return "+proj=merc +lat_ts=0 +lon_0=0 +k=1.000000 +x_0=0 +y_0=0" + wgs84, true
	case srid > bestSRIDNorthUTM && srid <= bestSRIDNorthUTM+60:
		return fmt.Sprintf("+proj=utm +zone=%d", srid-bestSRIDNorthUTM) + wgs84, true
	case srid > bestSRIDSouthUTM && srid <= bestSRIDSouthUTM+60:
		return fmt.Sprintf("+proj=utm +zone=%d +south", srid-bestSRIDSouthUTM) + wgs84, true
	case srid == bestSRIDNorthPole:
		return "+proj=laea +lat_0=90 +lon_0=-40 +x_0=0 +y_0=0" + wgs84, true
	case srid == bestSRIDSouthPole:
		return "+proj=laea +lat_0=-90 +lon_0=0 +x_0=0 +y_0=0" + wgs84, true
	case srid >= bestSRIDBands && srid < bestSRIDBands+20*int32(len(bandZoneWidths)):
		band, zone := (srid-bestSRIDBands)/20, (srid-bestSRIDBands)%20
		width := bandZoneWidths[band]
		if zone >= 360/width {
			return "", false
		}
		return fmt.Sprintf("+proj=laea +ellps=WGS84 +datum=WGS84 +lat_0=%d +lon_0=%g +units=m +no_defs", -75+30*band, float64(-180+width*zone)+float64(width)/2), true
	default:
		return "", false
	}
}

// The SRIDs that _ST_BestSRID returns, which count up from these bases.
const (
	bestSRIDMercator  = 999000
	bestSRIDNorthUTM  = 999000
	bestSRIDNorthPole = 999061
	bestSRIDSouthUTM  = 999100
	bestSRIDSouthPole = 999161
	bestSRIDBands     = 999163
)

// bandZoneWidths holds the width in degrees of the zones that divide each 30-degree latitude band, from the south.
var bandZoneWidths = []int32{90, 45, 30, 30, 45, 90}

// geodeticBox returns the box on the unit sphere's geocentric coordinates that holds the shape, including the bulges
// of the great-circle edges between its points and any pole that a polygon holds. An empty shape has a zero box.
func geodeticBox(s shape) (lo vec3, hi vec3) {
	first := true
	add := func(p vec3) {
		for i := range p {
			if first || p[i] < lo[i] {
				lo[i] = p[i]
			}
			if first || p[i] > hi[i] {
				hi[i] = p[i]
			}
		}
		first = false
	}
	dims := s.ctype.Dimension()
	for _, seq := range s.sequences() {
		for i := 0; i+dims <= len(seq); i += dims {
			b := toVec3(seq[i:])
			add(b)
			if i == 0 {
				continue
			}
			a := toVec3(seq[i-dims:])
			normal, ok := a.cross(b).normalized()
			if !ok {
				continue
			}
			for axis := range 3 {
				var unit vec3
				unit[axis] = 1
				along := unit.dot(normal)
				extreme, ok := vec3{unit[0] - along*normal[0], unit[1] - along*normal[1], unit[2] - along*normal[2]}.normalized()
				if !ok {
					continue
				}
				for _, candidate := range []vec3{extreme, {-extreme[0], -extreme[1], -extreme[2]}} {
					if a.cross(candidate).dot(normal) >= 0 && candidate.cross(b).dot(normal) >= 0 {
						add(candidate)
					}
				}
			}
		}
	}
	if s.dimension() == 2 && !first {
		outside := pointOutsideBox(lo, hi)
		if holdsPoint(s, vec3{0, 0, 1}, outside) {
			hi[2] = 1
		}
		if holdsPoint(s, vec3{0, 0, -1}, outside) {
			lo[2] = -1
		}
	}
	return lo, hi
}

// pointOutsideBox returns a point on the unit sphere outside a geocentric box, found among the box's corners pushed
// slightly outward.
func pointOutsideBox(lo vec3, hi vec3) vec3 {
	for _, x := range []float64{lo[0] - 0.01, hi[0] + 0.01} {
		for _, y := range []float64{lo[1] - 0.01, hi[1] + 0.01} {
			for _, z := range []float64{lo[2] - 0.01, hi[2] + 0.01} {
				p, ok := vec3{x, y, z}.normalized()
				if ok && (p[0] < lo[0] || p[0] > hi[0] || p[1] < lo[1] || p[1] > hi[1] || p[2] < lo[2] || p[2] > hi[2]) {
					return p
				}
			}
		}
	}
	return vec3{-lo[0], -lo[1], -lo[2]}
}

// holdsPoint returns whether a point lies within the surfaces of the shape, taking a polygon's interior to be the side
// of its rings away from the given point outside it. The arc from the outside point to the point crosses the rings an
// odd number of times exactly when the point lies within them.
func holdsPoint(s shape, p vec3, outside vec3) bool {
	dims := s.ctype.Dimension()
	crossings := 0
	for _, seq := range s.sequences() {
		for i := dims; i+dims <= len(seq); i += dims {
			if _, ok := arcCrossing(outside, p, toVec3(seq[i-dims:]), toVec3(seq[i:])); ok {
				crossings++
			}
		}
	}
	return crossings%2 == 1
}

// bestSRID returns the reserved SRID of the planar system that suits the extent of the shapes best. Extents within
// 70 degrees of a pole use a polar system, extents less than six degrees wide use a UTM zone, extents that fit in a
// zone of a latitude band use that zone's system, and everything else uses world Mercator. The extent's center and
// size come from the corners of its geocentric box.
func bestSRID(shapes ...shape) int32 {
	var lo, hi vec3
	for i, s := range shapes {
		shapeLo, shapeHi := geodeticBox(s)
		for j := range lo {
			if i == 0 || shapeLo[j] < lo[j] {
				lo[j] = shapeLo[j]
			}
			if i == 0 || shapeHi[j] > hi[j] {
				hi[j] = shapeHi[j]
			}
		}
	}
	var center vec3
	var directions [][2]float64
	minZ, maxZ := math.Inf(1), math.Inf(-1)
	for _, x := range []float64{lo[0], hi[0]} {
		for _, y := range []float64{lo[1], hi[1]} {
			for _, z := range []float64{lo[2], hi[2]} {
				corner, ok := vec3{x, y, z}.normalized()
				if !ok {
					corner = vec3{x, y, z}
				}
				center = center.plus(corner)
				minZ, maxZ = min(minZ, corner[2]), max(maxZ, corner[2])
				if length := math.Hypot(corner[0], corner[1]); length > 0 {
					directions = append(directions, [2]float64{corner[0] / length, corner[1] / length})
				}
			}
		}
	}
	if normalized, ok := center.normalized(); ok {
		center = normalized
	}
	lat, lon := math.Asin(center[2])*180/math.Pi, math.Atan2(center[1], center[0])*180/math.Pi
	height := (math.Asin(maxZ) - math.Asin(minZ)) * 180 / math.Pi
	width := 0.0
	for _, a := range directions {
		for _, b := range directions {
			width = max(width, math.Acos(max(-1, min(1, a[0]*b[0]+a[1]*b[1]))))
		}
	}
	width *= 180 / math.Pi
	if lat > 70 && height < 45 {
		return bestSRIDNorthPole
	}
	if lat < -70 && height < 45 {
		return bestSRIDSouthPole
	}
	if width < 6 {
		zone := min(int32(math.Floor((lon+180)/6)), 59) + 1
		if lat < 0 {
			return bestSRIDSouthUTM + zone
		}
		return bestSRIDNorthUTM + zone
	}
	band := min(int32(math.Floor((lat+90)/30)), int32(len(bandZoneWidths)-1))
	if zoneWidth := bandZoneWidths[band]; height < 25 && width < float64(zoneWidth) {
		return bestSRIDBands + 20*band + int32(math.Floor((lon+180)/float64(zoneWidth)))
	}
	return bestSRIDMercator
}

// stBestSRID implements _ST_BestSRID, for one geography or for two together.
func stBestSRID(ctx *sql.Context, args ...any) (any, error) {
	shapes := make([]shape, len(args))
	for i, arg := range args {
		shapes[i] = arg.(Geometry).Shape
	}
	return bestSRID(shapes...), nil
}

// throughBestSRID runs a geometry operation on geographies after transforming them to the planar system that
// _ST_BestSRID chooses for all of them, and transforms the result back to the first geography's SRID.
func throughBestSRID(ctx *sql.Context, runner sql.StatementRunner, geographies []Geometry, op func(geometries []any) (any, error)) (any, error) {
	shapes := make([]shape, len(geographies))
	for i, g := range geographies {
		shapes[i] = g.Shape
	}
	srid := bestSRID(shapes...)
	geometries := make([]any, len(geographies))
	for i, g := range geographies {
		projected, err := stTransform(ctx, runner, g, srid)
		if err != nil {
			return nil, err
		}
		geometries[i] = projected
	}
	result, err := op(geometries)
	if err != nil {
		return nil, err
	}
	back, err := stTransform(ctx, runner, result, geographies[0].SRID)
	if err != nil {
		return nil, err
	}
	return asGeography(ctx, back.(Geometry))
}

// geographyBuffer implements ST_Buffer for geography, buffering in the planar system that suits the geography best.
func geographyBuffer(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	return throughBestSRID(ctx, runner, []Geometry{args[0].(Geometry)}, func(geometries []any) (any, error) {
		return stBuffer(ctx, append(geometries, args[1:]...)...)
	})
}

// geographyIntersection implements ST_Intersection for geography, intersecting in the planar system that suits both
// geographies best.
func geographyIntersection(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	return throughBestSRID(ctx, runner, []Geometry{args[0].(Geometry), args[1].(Geometry)}, func(geometries []any) (any, error) {
		return overlayImpl("intersection", intersectionWithEmpty, (*spatial.GEOS).Intersection)(ctx, geometries...)
	})
}
