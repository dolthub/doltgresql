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

package extensions

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"

	framework "github.com/dolthub/doltgresql/testing/go"
)

func TestPostGIS(t *testing.T) {
	framework.RunScripts(t, []framework.ScriptTest{
		{
			Name: `postgis geometry input and output`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT 'POINT(1 2)'::geometry::text;`,
					Expected: []sql.Row{{`0101000000000000000000F03F0000000000000040`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('SRID=4326;POINT(1 2)'::geometry);`,
					Expected: []sql.Row{{`SRID=4326;POINT(1 2)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('POINT Z (1 2 3)'::geometry), ST_AsEWKT('POINT M (1 2 3)'::geometry), ST_AsEWKT('POINT ZM (1 2 3 4)'::geometry);`,
					Expected: []sql.Row{{`POINT(1 2 3)`, `POINTM(1 2 3)`, `POINT(1 2 3 4)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('LINESTRING(0 0,1 1,2 0)'::geometry);`,
					Expected: []sql.Row{{`LINESTRING(0 0,1 1,2 0)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('POLYGON((0 0,0 1,1 1,1 0,0 0),(0.2 0.2,0.2 0.4,0.4 0.4,0.2 0.2))'::geometry);`,
					Expected: []sql.Row{{`POLYGON((0 0,0 1,1 1,1 0,0 0),(0.2 0.2,0.2 0.4,0.4 0.4,0.2 0.2))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('MULTIPOINT((1 1),(2 2))'::geometry), ST_AsEWKT('MULTIPOINT(1 1,EMPTY)'::geometry);`,
					Expected: []sql.Row{{`MULTIPOINT(1 1,2 2)`, `MULTIPOINT(1 1,EMPTY)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('GEOMETRYCOLLECTION(POINT(1 1),LINESTRING EMPTY)'::geometry);`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(POINT(1 1),LINESTRING EMPTY)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('CIRCULARSTRING(0 0,1 1,2 0)'::geometry), ST_AsEWKT('COMPOUNDCURVE((0 0,1 1),CIRCULARSTRING(1 1,2 2,3 1))'::geometry);`,
					Expected: []sql.Row{{`CIRCULARSTRING(0 0,1 1,2 0)`, `COMPOUNDCURVE((0 0,1 1),CIRCULARSTRING(1 1,2 2,3 1))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('CURVEPOLYGON(CIRCULARSTRING(0 0,1 1,2 0,1 -1,0 0))'::geometry);`,
					Expected: []sql.Row{{`CURVEPOLYGON(CIRCULARSTRING(0 0,1 1,2 0,1 -1,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('TRIANGLE((0 0,0 1,1 0,0 0))'::geometry), ST_AsEWKT('TIN(((0 0,0 1,1 0,0 0)))'::geometry);`,
					Expected: []sql.Row{{`TRIANGLE((0 0,0 1,1 0,0 0))`, `TIN(((0 0,0 1,1 0,0 0)))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('POLYHEDRALSURFACE(((0 0,0 1,1 1,0 0)))'::geometry);`,
					Expected: []sql.Row{{`POLYHEDRALSURFACE(((0 0,0 1,1 1,0 0)))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('POINT EMPTY'::geometry), ST_AsEWKT('GEOMETRYCOLLECTION EMPTY'::geometry);`,
					Expected: []sql.Row{{`POINT EMPTY`, `GEOMETRYCOLLECTION EMPTY`}},
				},
				{
					Query:    `SELECT 'SRID=4326;POINT(1 2)'::geometry::text;`,
					Expected: []sql.Row{{`0101000020E6100000000000000000F03F0000000000000040`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('0101000020E6100000000000000000F03F0000000000000040'::geometry);`,
					Expected: []sql.Row{{`SRID=4326;POINT(1 2)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('{"type":"Point","coordinates":[1,2]}'::geometry);`,
					Expected: []sql.Row{{`POINT(1 2)`}},
				},
				{
					Query:           `SELECT 'POINT(1 2'::geometry;`,
					ExpectedErr:     `parse error - invalid geometry`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT 'LINESTRING(1 2)'::geometry;`,
					ExpectedErr:     `geometry requires more points`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT 'POLYGON((0 0,0 1,1 1,0 0.5))'::geometry;`,
					ExpectedErr:     `geometry contains non-closed rings`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT 'NOTAGEOMETRY'::geometry;`,
					ExpectedErr:     `parse error - invalid geometry`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsText('POINT(1.123456789012345 2)'::geometry), ST_AsText('POINT(1.123456789 2)'::geometry, 3);`,
					Expected: []sql.Row{{`POINT(1.123456789012345 2)`, `POINT(1.123 2)`}},
				},
				{
					Query:    `SELECT ST_AsText('POINT(1e20 -1e-20)'::geometry);`,
					Expected: []sql.Row{{`POINT(1e+20 -1e-20)`}},
				},
				{
					Query:    `SELECT encode(ST_AsBinary('POINT(1 2)'::geometry), 'hex'), encode(ST_AsEWKB('SRID=4326;POINT(1 2)'::geometry), 'hex');`,
					Expected: []sql.Row{{`0101000000000000000000f03f0000000000000040`, `0101000020e6100000000000000000f03f0000000000000040`}},
				},
				{
					Query:    `SELECT ST_AsHEXEWKB('SRID=4326;POINT(1 2)'::geometry), ST_AsHEXEWKB('POINT(1 2)'::geometry, 'XDR');`,
					Expected: []sql.Row{{`0101000020E6100000000000000000F03F0000000000000040`, `00000000013FF00000000000004000000000000000`}},
				},
				{
					Query:    `SELECT ST_AsText('POINT Z (1 2 3)'::geometry), ST_AsEWKT('POINT Z (1 2 3)'::geometry);`,
					Expected: []sql.Row{{`POINT Z (1 2 3)`, `POINT(1 2 3)`}},
				},
			},
		},
		{
			Name: `postgis geometry type modifiers`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
				`CREATE TABLE geo_typmod (id INT PRIMARY KEY, g geometry(Point,4326), m geometry(MultiPolygon), z geometry(PointZ));`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: `SELECT format_type(atttypid, atttypmod) FROM pg_attribute WHERE attrelid = 'geo_typmod'::regclass AND attnum > 1 ORDER BY attnum;`,
					Expected: []sql.Row{
						{`geometry(Point,4326)`},
						{`geometry(MultiPolygon)`},
						{`geometry(PointZ)`},
					},
				},
				{
					Query:    `INSERT INTO geo_typmod (id, g) VALUES (1, 'POINT(1 2)');`,
					Expected: []sql.Row{},
				},
				{
					Query:           `INSERT INTO geo_typmod (id, g) VALUES (2, 'SRID=3857;POINT(1 2)');`,
					ExpectedErr:     `Geometry SRID (3857) does not match column SRID (4326)`,
					ExpectedErrCode: `22023`,
				},
				{
					Query:           `INSERT INTO geo_typmod (id, g) VALUES (3, 'LINESTRING(0 0,1 1)');`,
					ExpectedErr:     `Geometry type (LineString) does not match column type (Point)`,
					ExpectedErrCode: `22023`,
				},
				{
					Query:    `INSERT INTO geo_typmod (id, m) VALUES (4, 'POLYGON((0 0,0 1,1 1,0 0))');`,
					Expected: []sql.Row{},
				},
				{
					Query:           `INSERT INTO geo_typmod (id, z) VALUES (5, 'POINT(1 2)');`,
					ExpectedErr:     `Column has Z dimension but geometry does not`,
					ExpectedErrCode: `22023`,
				},
				{
					Query:    `SELECT ST_AsEWKT('POINT(1 2)'::geometry(Point,4326)), ST_AsEWKT('POLYGON((0 0,0 1,1 1,0 0))'::geometry(MultiPolygon));`,
					Expected: []sql.Row{{`SRID=4326;POINT(1 2)`, `MULTIPOLYGON(((0 0,0 1,1 1,0 0)))`}},
				},
				{
					Query:           `SELECT 'POINT(1 2)'::geometry(Polygon);`,
					ExpectedErr:     `Geometry type (Point) does not match column type (Polygon)`,
					ExpectedErrCode: `22023`,
				},
				{
					Query:           `SELECT 'POINT(1 2)'::geometry(Pointy);`,
					ExpectedErr:     `Invalid geometry type modifier: pointy`,
					ExpectedErrCode: `22023`,
				},
				{
					Query: `SELECT postgis_typmod_type(atttypmod), postgis_typmod_srid(atttypmod), postgis_typmod_dims(atttypmod) FROM pg_attribute WHERE attrelid = 'geo_typmod'::regclass AND attnum > 1 ORDER BY attnum;`,
					Expected: []sql.Row{
						{`Point`, 4326, 2},
						{`MultiPolygon`, 0, 2},
						{`PointZ`, 0, 3},
					},
				},
			},
		},
		{
			Name: `postgis geography`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsEWKT('POINT(1 2)'::geography);`,
					Expected: []sql.Row{{`SRID=4326;POINT(1 2)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('SRID=4269;POINT(1 2)'::geography);`,
					Expected: []sql.Row{{`SRID=4269;POINT(1 2)`}},
				},
				{
					Query:           `SELECT 'SRID=3857;POINT(1 2)'::geography;`,
					ExpectedErr:     `Only lon/lat coordinate systems are supported in geography.`,
					ExpectedErrCode: `22023`,
				},
				{
					Query:    `SELECT ST_AsEWKT('POINT(200 10)'::geography);`,
					Expected: []sql.Row{{`SRID=4326;POINT(-160 10)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('POINT(1 2)'::geometry::geography), ST_AsEWKT('POINT(1 2)'::geography::geometry);`,
					Expected: []sql.Row{{`SRID=4326;POINT(1 2)`, `SRID=4326;POINT(1 2)`}},
				},
				{
					Query:    `SELECT ST_Distance('POINT(0 0)'::geography, 'POINT(1 1)'::geography), ST_Distance('POINT(0 0)'::geography, 'POINT(1 1)'::geography, false);`,
					Expected: []sql.Row{{156899.56829134, 157249.59776851}},
				},
				{
					Query:    `SELECT ST_Distance('LINESTRING(0 0,10 0)'::geography, 'POINT(5 1)'::geography);`,
					Expected: []sql.Row{{110574.3885578}},
				},
				{
					Query:    `SELECT ST_Distance('POLYGON((0 0,0 10,10 10,10 0,0 0))'::geography, 'POINT(5 5)'::geography);`,
					Expected: []sql.Row{{0.0}},
				},
				{
					Query:    `SELECT ST_DWithin('POINT(0 0)'::geography, 'POINT(1 1)'::geography, 157000), ST_DWithin('POINT(0 0)'::geography, 'POINT(1 1)'::geography, 158000);`,
					Expected: []sql.Row{{`t`, `t`}},
				},
				{
					Query:    `SELECT ST_Area('POLYGON((0 0,0 1,1 1,1 0,0 0))'::geography) > 12308778361, ST_Perimeter('POLYGON((0 0,0 1,1 1,1 0,0 0))'::geography);`,
					Expected: []sql.Row{{`t`, 443770.9172483019}},
				},
				{
					Query:    `SELECT round(ST_Length('LINESTRING(0 0,1 1)'::geography)::numeric, 6), ST_Length('LINESTRING(0 0,1 1)'::geography, false);`,
					Expected: []sql.Row{{framework.Numeric(`156899.568291`), 157249.5977685051}},
				},
				{
					Query:    `SELECT round(ST_Azimuth('POINT(0 0)'::geography, 'POINT(1 1)'::geography)::numeric, 12);`,
					Expected: []sql.Row{{framework.Numeric(`0.788680084526`)}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Project('POINT(0 0)'::geography, 100000, radians(45)), 12);`,
					Expected: []sql.Row{{`SRID=4326;POINT(0.635231029126 0.639472334729)`}},
				},
				{
					Query:    `SELECT ST_Intersects('POLYGON((0 0,0 10,10 10,10 0,0 0))'::geography, 'POINT(5 5)'::geography), ST_Covers('POLYGON((0 0,0 10,10 10,10 0,0 0))'::geography, 'POINT(15 5)'::geography);`,
					Expected: []sql.Row{{`t`, `f`}},
				},
				{
					Query:           `SELECT ST_Distance('SRID=4326;POINT(0 0)'::geography, 'SRID=4269;POINT(1 1)'::geography);`,
					ExpectedErr:     `geography_distance: Operation on mixed SRID geometries (Point, 4326) != (Point, 4269)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_LineInterpolatePoint('LINESTRING(0 0,10 10)'::geography, 0.5));`,
					Expected: []sql.Row{{`SRID=4326;POINT(4.961882922527124 5.01950880089445)`}},
				},
			},
		},
		{
			Name: `postgis constructors`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsEWKT(ST_MakePoint(1, 2)), ST_AsEWKT(ST_MakePoint(1, 2, 3)), ST_AsEWKT(ST_MakePoint(1, 2, 3, 4));`,
					Expected: []sql.Row{{`POINT(1 2)`, `POINT(1 2 3)`, `POINT(1 2 3 4)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Point(1, 2)), ST_AsEWKT(ST_Point(1, 2, 4326)), ST_AsEWKT(ST_PointZ(1, 2, 3, 4326)), ST_AsEWKT(ST_PointM(1, 2, 3)), ST_AsEWKT(ST_PointZM(1, 2, 3, 4));`,
					Expected: []sql.Row{{`POINT(1 2)`, `SRID=4326;POINT(1 2)`, `SRID=4326;POINT(1 2 3)`, `POINTM(1 2 3)`, `POINT(1 2 3 4)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_MakePointM(1, 2, 3));`,
					Expected: []sql.Row{{`POINTM(1 2 3)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromText('POINT(1 2)')), ST_AsEWKT(ST_GeomFromText('POINT(1 2)', 4326)), ST_AsEWKT(ST_GeomFromEWKT('SRID=3857;POINT(1 2)'));`,
					Expected: []sql.Row{{`POINT(1 2)`, `SRID=4326;POINT(1 2)`, `SRID=3857;POINT(1 2)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromWKB('\x0101000000000000000000f03f0000000000000040'::bytea, 4326));`,
					Expected: []sql.Row{{`SRID=4326;POINT(1 2)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_PointFromText('POINT(1 2)')), ST_AsEWKT(ST_PointFromText('LINESTRING(0 0,1 1)')), ST_AsEWKT(ST_PolygonFromText('POLYGON((0 0,0 1,1 1,0 0))', 4326));`,
					Expected: []sql.Row{{`POINT(1 2)`, nil, `SRID=4326;POLYGON((0 0,0 1,1 1,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGeoJSON('{"type":"Point","coordinates":[1,2]}')), ST_AsEWKT(ST_GeomFromGeoJSON('{"type":"LineString","coordinates":[[1,2,3],[3,4]]}'));`,
					Expected: []sql.Row{{`SRID=4326;POINT(1 2)`, `SRID=4326;LINESTRING(1 2 3,3 4 0)`}},
				},
				{
					Query:           `SELECT ST_GeomFromGeoJSON('{"type":"Point"}');`,
					ExpectedErr:     `Unable to find 'coordinates' in GeoJSON string`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_MakeLine('POINT(1 1)'::geometry, 'POINT(2 2)'::geometry)), ST_AsEWKT(ST_MakeLine(ARRAY['POINT(1 1)'::geometry, 'POINT(2 2)', 'POINT(3 3)']));`,
					Expected: []sql.Row{{`LINESTRING(1 1,2 2)`, `LINESTRING(1 1,2 2,3 3)`}},
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_MakeLine('POINT(1 1)'::geometry, 'POLYGON((0 0,0 1,1 1,0 0))'::geometry));`,
					ExpectedErr:     `Input geometries must be points or lines`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Collect('POINT(1 1)'::geometry, 'POINT(2 2)'::geometry)), ST_AsEWKT(ST_Collect('POINT(1 1)'::geometry, 'LINESTRING(0 0,1 1)'::geometry));`,
					Expected: []sql.Row{{`MULTIPOINT(1 1,2 2)`, `GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 0,1 1))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Collect(ARRAY['POINT(1 1)'::geometry, 'POINT(2 2)']));`,
					Expected: []sql.Row{{`MULTIPOINT(1 1,2 2)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_MakePolygon('LINESTRING(0 0,0 1,1 1,0 0)'::geometry)), ST_AsEWKT(ST_Polygon('LINESTRING(0 0,0 1,1 1,0 0)'::geometry, 4326));`,
					Expected: []sql.Row{{`POLYGON((0 0,0 1,1 1,0 0))`, `SRID=4326;POLYGON((0 0,0 1,1 1,0 0))`}},
				},
				{
					Query:           `SELECT ST_MakePolygon('LINESTRING(0 0,0 1,1 1)'::geometry);`,
					ExpectedErr:     `lwpoly_from_lwlines: shell must have at least 4 points`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_MakeEnvelope(0, 0, 1, 2, 4326)), ST_AsEWKT(ST_LineFromMultiPoint('MULTIPOINT(0 0,1 1)'::geometry));`,
					Expected: []sql.Row{{`SRID=4326;POLYGON((0 0,0 2,1 2,1 0,0 0))`, `LINESTRING(0 0,1 1)`}},
				},
			},
		},
		{
			Name: `postgis accessors`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT GeometryType('POINT(1 2)'::geometry), ST_GeometryType('POINT M (1 2 3)'::geometry), GeometryType('POINT M (1 2 3)'::geometry);`,
					Expected: []sql.Row{{`POINT`, `ST_Point`, `POINTM`}},
				},
				{
					Query:    `SELECT ST_SRID('SRID=4326;POINT(1 2)'::geometry), ST_AsEWKT(ST_SetSRID('POINT(1 2)'::geometry, 3857));`,
					Expected: []sql.Row{{4326, `SRID=3857;POINT(1 2)`}},
				},
				{
					Query:    `SELECT ST_X('POINT(1 2)'::geometry), ST_Y('POINT(1 2)'::geometry), ST_Z('POINT Z (1 2 3)'::geometry), ST_M('POINT M (1 2 3)'::geometry), ST_Z('POINT(1 2)'::geometry);`,
					Expected: []sql.Row{{1.0, 2.0, 3.0, 3.0, nil}},
				},
				{
					Query:           `SELECT ST_X('LINESTRING(0 0,1 1)'::geometry);`,
					ExpectedErr:     `Argument to ST_X() must have type POINT`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_NumGeometries('MULTIPOINT(1 1,2 2)'::geometry), ST_AsEWKT(ST_GeometryN('MULTIPOINT(1 1,2 2)'::geometry, 2)), ST_NPoints('POLYGON((0 0,0 1,1 1,0 0))'::geometry);`,
					Expected: []sql.Row{{2, `POINT(2 2)`, 4}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_StartPoint('LINESTRING(0 0,1 1,2 2)'::geometry)), ST_AsEWKT(ST_EndPoint('LINESTRING(0 0,1 1,2 2)'::geometry)), ST_AsEWKT(ST_PointN('LINESTRING(0 0,1 1,2 2)'::geometry, -1));`,
					Expected: []sql.Row{{`POINT(0 0)`, `POINT(2 2)`, `POINT(2 2)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_ExteriorRing('POLYGON((0 0,0 1,1 1,0 0),(0.1 0.1,0.2 0.1,0.2 0.2,0.1 0.1))'::geometry)), ST_NumInteriorRings('POLYGON((0 0,0 1,1 1,0 0),(0.1 0.1,0.2 0.1,0.2 0.2,0.1 0.1))'::geometry), ST_NRings('POLYGON((0 0,0 1,1 1,0 0),(0.1 0.1,0.2 0.1,0.2 0.2,0.1 0.1))'::geometry);`,
					Expected: []sql.Row{{`LINESTRING(0 0,0 1,1 1,0 0)`, 1, 2}},
				},
				{
					Query:    `SELECT ST_Dimension('POLYGON((0 0,0 1,1 1,0 0))'::geometry), ST_CoordDim('POINT ZM (1 2 3 4)'::geometry), ST_NDims('POINT M (1 2 3)'::geometry), ST_Zmflag('POINT Z (1 2 3)'::geometry);`,
					Expected: []sql.Row{{2, 4, 3, 2}},
				},
				{
					Query:    `SELECT ST_IsEmpty('POINT EMPTY'::geometry), ST_IsClosed('LINESTRING(0 0,1 1,0 0)'::geometry), ST_IsCollection('MULTIPOINT(1 1)'::geometry), ST_IsRing('LINESTRING(0 0,0 1,1 1,0 0)'::geometry), ST_IsSimple('LINESTRING(0 0,1 1,0 1,1 0)'::geometry);`,
					Expected: []sql.Row{{`t`, `t`, `t`, `t`, `f`}},
				},
				{
					Query:    `SELECT ST_HasArc('CIRCULARSTRING(0 0,1 1,2 0)'::geometry), ST_NumCurves('COMPOUNDCURVE((0 0,1 1),(1 1,2 0))'::geometry), ST_NumPatches('TIN(((0 0,0 1,1 0,0 0)))'::geometry);`,
					Expected: []sql.Row{{`t`, 2, 1}},
				},
				{
					Query:    `SELECT ST_Summary('SRID=4326;LINESTRING(0 0,1 1)'::geometry);`,
					Expected: []sql.Row{{`LineString[S] with 2 points`}},
				},
				{
					Query:    `SELECT ST_MemSize('POINT(1 2)'::geometry), ST_MemSize('LINESTRING(0 0,1 1,2 2)'::geometry);`,
					Expected: []sql.Row{{32, 80}},
				},
				{
					Query:    `SELECT ST_IsPolygonCW('POLYGON((0 0,0 1,1 1,0 0))'::geometry), ST_IsPolygonCCW('POLYGON((0 0,0 1,1 1,0 0))'::geometry);`,
					Expected: []sql.Row{{`t`, `f`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Points('LINESTRING(0 0,1 1,2 2)'::geometry));`,
					Expected: []sql.Row{{`MULTIPOINT(0 0,1 1,2 2)`}},
				},
			},
		},
		{
			Name: `postgis bounding boxes`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT Box2D('LINESTRING(0 0,1 2)'::geometry), Box3D('LINESTRING Z (0 0 1,1 2 3)'::geometry);`,
					Expected: []sql.Row{{`BOX(0 0,1 2)`, `BOX3D(0 0 1,1 2 3)`}},
				},
				{
					Query:    `SELECT 'BOX(0 0,1 2)'::box2d, 'BOX3D(0 0 0,1 2 3)'::box3d;`,
					Expected: []sql.Row{{`BOX(0 0,1 2)`, `BOX3D(0 0 0,1 2 3)`}},
				},
				{
					Query:    `SELECT ST_XMin('BOX(1 2,3 4)'::box2d), ST_YMax('BOX(1 2,3 4)'::box2d), ST_ZMax('BOX3D(1 2 3,4 5 6)'::box3d);`,
					Expected: []sql.Row{{1.0, 4.0, 6.0}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Envelope('LINESTRING(0 0,1 2)'::geometry)), ST_AsEWKT(ST_Expand('POINT(1 1)'::geometry, 1));`,
					Expected: []sql.Row{{`POLYGON((0 0,0 2,1 2,1 0,0 0))`, `POLYGON((0 0,0 2,2 2,2 0,0 0))`}},
				},
				{
					Query:    `SELECT ST_MakeBox2D('POINT(0 0)'::geometry, 'POINT(1 2)'::geometry), ST_3DMakeBox('POINT Z (0 0 0)'::geometry, 'POINT Z (1 2 3)'::geometry);`,
					Expected: []sql.Row{{`BOX(0 0,1 2)`, `BOX3D(0 0 0,1 2 3)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT('BOX(0 0,1 2)'::box2d::geometry);`,
					Expected: []sql.Row{{`POLYGON((0 0,0 2,1 2,1 0,0 0))`}},
				},
				{
					Query:    `SELECT 'POINT(1 1)'::geometry && 'LINESTRING(0 0,2 2)'::geometry, 'POINT(5 5)'::geometry && 'LINESTRING(0 0,2 2)'::geometry;`,
					Expected: []sql.Row{{`t`, `f`}},
				},
				{
					Query:    `SELECT 'LINESTRING(0 0,2 2)'::geometry ~ 'POINT(1 1)'::geometry, 'POINT(1 1)'::geometry @ 'LINESTRING(0 0,2 2)'::geometry, 'POINT(1 1)'::geometry ~= 'POINT(1 1)'::geometry;`,
					Expected: []sql.Row{{`t`, `t`, `t`}},
				},
				{
					Query:    `SELECT 'POINT(0 0)'::geometry << 'POINT(1 0)'::geometry, 'POINT(0 0)'::geometry >> 'POINT(1 0)'::geometry, 'POINT(0 0)'::geometry <<| 'POINT(0 1)'::geometry;`,
					Expected: []sql.Row{{`t`, `f`, `t`}},
				},
				{
					Query:    `SELECT 'POINT(0 0)'::geometry <#> 'POINT(3 4)'::geometry;`,
					Expected: []sql.Row{{5.0}},
				},
				{
					Query:    `SELECT 'POINT(0 0)'::geography && 'POINT(0 0)'::geography;`,
					Expected: []sql.Row{{`t`}},
				},
			},
		},
		{
			Name: `postgis comparisons`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT 'POINT(1 1)'::geometry = 'POINT(1 1)'::geometry, 'POINT(1 1)'::geometry = 'POINT(2 2)'::geometry, 'POINT(1 1)'::geometry < 'POINT(2 2)'::geometry;`,
					Expected: []sql.Row{{`t`, `f`, `t`}},
				},
				{
					Query: `SELECT ST_AsEWKT(g) FROM (VALUES ('POINT(2 2)'::geometry), ('POINT(1 1)'), ('POINT EMPTY'), ('LINESTRING(0 0,1 1)')) t(g) ORDER BY g;`,
					Expected: []sql.Row{
						{`POINT EMPTY`},
						{`LINESTRING(0 0,1 1)`},
						{`POINT(1 1)`},
						{`POINT(2 2)`},
					},
				},
				{
					Query:    `SELECT ST_OrderingEquals('LINESTRING(0 0,1 1)'::geometry, 'LINESTRING(0 0,1 1)'::geometry), ST_OrderingEquals('LINESTRING(0 0,1 1)'::geometry, 'LINESTRING(1 1,0 0)'::geometry);`,
					Expected: []sql.Row{{`t`, `f`}},
				},
				{
					Query:    `SELECT ST_Equals('LINESTRING(0 0,1 1)'::geometry, 'LINESTRING(1 1,0 0)'::geometry);`,
					Expected: []sql.Row{{`t`}},
				},
			},
		},
		{
			Name: `postgis measures and predicates`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_Area('POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry), ST_Perimeter('POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry), ST_Length('LINESTRING(0 0,3 4)'::geometry);`,
					Expected: []sql.Row{{4.0, 8.0, 5.0}},
				},
				{
					Query:    `SELECT ST_Length('CIRCULARSTRING(0 0,1 1,2 0)'::geometry);`,
					Expected: []sql.Row{{3.141592653589793}},
				},
				{
					Query:    `SELECT ST_Distance('POINT(0 0)'::geometry, 'POINT(3 4)'::geometry), ST_Distance('POINT(0 0)'::geometry, 'POINT EMPTY'::geometry);`,
					Expected: []sql.Row{{5.0, nil}},
				},
				{
					Query:           `SELECT ST_Distance('POINT(0 0)'::geometry, 'SRID=4326;POINT(3 4)'::geometry);`,
					ExpectedErr:     `ST_Distance: Operation on mixed SRID geometries (Point, 0) != (Point, 4326)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_DWithin('POINT(0 0)'::geometry, 'POINT(3 4)'::geometry, 5), ST_DWithin('POINT(0 0)'::geometry, 'POINT(3 4)'::geometry, 4.9);`,
					Expected: []sql.Row{{`t`, `f`}},
				},
				{
					Query:           `SELECT ST_DWithin('POINT(0 0)'::geometry, 'POINT(3 4)'::geometry, -1);`,
					ExpectedErr:     `Tolerance cannot be less than zero`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_Contains('POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry, 'POINT(1 1)'::geometry), ST_Within('POINT(1 1)'::geometry, 'POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry), ST_Intersects('POINT(5 5)'::geometry, 'POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry);`,
					Expected: []sql.Row{{`t`, `t`, `f`}},
				},
				{
					Query:    `SELECT ST_Touches('POINT(0 0)'::geometry, 'POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry), ST_Crosses('LINESTRING(0 0,2 2)'::geometry, 'LINESTRING(0 2,2 0)'::geometry), ST_Overlaps('POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry, 'POLYGON((1 1,1 3,3 3,3 1,1 1))'::geometry);`,
					Expected: []sql.Row{{`t`, `t`, `t`}},
				},
				{
					Query:    `SELECT ST_Covers('POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry, 'POINT(0 0)'::geometry), ST_CoveredBy('POINT(0 0)'::geometry, 'POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry), ST_ContainsProperly('POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry, 'POINT(0 0)'::geometry);`,
					Expected: []sql.Row{{`t`, `t`, `f`}},
				},
				{
					Query:    `SELECT ST_Disjoint('POINT(5 5)'::geometry, 'POINT EMPTY'::geometry), ST_Intersects('POINT(5 5)'::geometry, 'POINT EMPTY'::geometry);`,
					Expected: []sql.Row{{`t`, `f`}},
				},
				{
					Query:    `SELECT ST_Relate('POINT(1 1)'::geometry, 'POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry), ST_Relate('POINT(1 1)'::geometry, 'POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry, '0FFFFF212'), ST_RelateMatch('0FFFFF212', 'T*F**F***');`,
					Expected: []sql.Row{{`0FFFFF212`, `t`, `t`}},
				},
				{
					Query:    `SELECT ST_MaxDistance('POINT(0 0)'::geometry, 'LINESTRING(1 0,3 4)'::geometry), ST_AsEWKT(ST_LongestLine('POINT(0 0)'::geometry, 'LINESTRING(1 0,3 4)'::geometry));`,
					Expected: []sql.Row{{5.0, `LINESTRING(0 0,3 4)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_ClosestPoint('LINESTRING(0 0,10 0)'::geometry, 'POINT(5 3)'::geometry)), ST_AsEWKT(ST_ShortestLine('LINESTRING(0 0,10 0)'::geometry, 'POINT(5 3)'::geometry));`,
					Expected: []sql.Row{{`POINT(5 0)`, `LINESTRING(5 0,5 3)`}},
				},
				{
					Query:    `SELECT ST_DFullyWithin('LINESTRING(0 0,10 0)'::geometry, 'POINT(5 3)'::geometry, 3);`,
					Expected: []sql.Row{{`t`}},
				},
				{
					Query:    `SELECT degrees(ST_Azimuth('POINT(0 0)'::geometry, 'POINT(1 1)'::geometry)), ST_Azimuth('POINT(0 0)'::geometry, 'POINT(0 0)'::geometry);`,
					Expected: []sql.Row{{45.0, nil}},
				},
				{
					Query:    `SELECT degrees(ST_Angle('POINT(0 0)'::geometry, 'POINT(1 0)'::geometry, 'POINT(1 1)'::geometry)), degrees(ST_Angle('LINESTRING(0 0,1 0)'::geometry, 'LINESTRING(0 0,0 1)'::geometry));`,
					Expected: []sql.Row{{90.0, 270.0}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Project('POINT(0 0)'::geometry, 'POINT(1 1)'::geometry, 2));`,
					Expected: []sql.Row{{`POINT(2.414213562373095 2.414213562373095)`}},
				},
			},
		},
		{
			Name: `postgis overlays and processing`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsEWKT(ST_Intersection('POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry, 'POINT(1 1)'::geometry));`,
					Expected: []sql.Row{{`POINT(1 1)`}},
				},
				{
					Query:    `SELECT ST_Area(ST_Union('POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry, 'POLYGON((1 1,1 3,3 3,3 1,1 1))'::geometry)), ST_Area(ST_Difference('POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry, 'POLYGON((1 1,1 3,3 3,3 1,1 1))'::geometry)), ST_Area(ST_SymDifference('POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry, 'POLYGON((1 1,1 3,3 3,3 1,1 1))'::geometry));`,
					Expected: []sql.Row{{7.0, 3.0, 6.0}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Difference('POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry, 'POINT EMPTY'::geometry)), ST_AsEWKT(ST_Intersection('POINT(1 1)'::geometry, 'LINESTRING EMPTY'::geometry));`,
					Expected: []sql.Row{{`POLYGON((0 0,0 2,2 2,2 0,0 0))`, `LINESTRING EMPTY`}},
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_Union('POINT(1 1)'::geometry, 'SRID=4326;POINT(1 1)'::geometry));`,
					ExpectedErr:     `lwgeom_union_prec: Operation on mixed SRID geometries (0 != 4326)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_ConvexHull('MULTIPOINT(0 0,2 0,1 1,0 2)'::geometry)), ST_AsEWKT(ST_Centroid('POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry));`,
					Expected: []sql.Row{{`POLYGON((0 0,0 2,2 0,0 0))`, `POINT(1 1)`}},
				},
				{
					Query:    `SELECT ST_Area(ST_Buffer('POINT(0 0)'::geometry, 1)) BETWEEN 3.12 AND 3.13, ST_NPoints(ST_Buffer('POINT(0 0)'::geometry, 1, 2));`,
					Expected: []sql.Row{{`t`, 9}},
				},
				{
					Query:           `SELECT ST_Buffer('POINT(0 0)'::geometry, 1, 'endcap=nope');`,
					ExpectedErr:     `Invalid buffer end cap style: nope (accept: 'round', 'flat', 'butt' or 'square')`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_IsValid('POLYGON((0 0,0 2,2 2,2 0,0 0))'::geometry), ST_IsValid('POLYGON((0 0,2 2,0 2,2 0,0 0))'::geometry);`,
					Expected: []sql.Row{{`t`, `f`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Boundary('POLYGON((0 0,0 1,1 1,0 0))'::geometry)), ST_AsEWKT(ST_Boundary('LINESTRING(0 0,1 1,2 0)'::geometry)), ST_AsEWKT(ST_Boundary('POINT(1 1)'::geometry));`,
					Expected: []sql.Row{{`LINESTRING(0 0,0 1,1 1,0 0)`, `MULTIPOINT(0 0,2 0)`, `POINT EMPTY`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Simplify('LINESTRING(0 0,1 0.1,2 0,3 5,4 0)'::geometry, 0.5)), ST_AsEWKT(ST_Simplify('POLYGON((0 0,0 1,1 1,0 0))'::geometry, 5));`,
					Expected: []sql.Row{{`LINESTRING(0 0,2 0,3 5,4 0)`, nil}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_CurveToLine('CIRCULARSTRING(0 0,1 1,2 0)'::geometry, 2));`,
					Expected: []sql.Row{{`LINESTRING(0 0,0.292893218813453 0.707106781186548,1 1,1.707106781186548 0.707106781186548,2 0)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Normalize('POLYGON((0 1,0 0,1 0,1 1,0 1))'::geometry)), ST_AsEWKT(ST_Normalize('MULTIPOINT(1 1,3 3,2 2)'::geometry));`,
					Expected: []sql.Row{{`POLYGON((0 0,0 1,1 1,1 0,0 0))`, `MULTIPOINT(3 3,2 2,1 1)`}},
				},
			},
		},
		{
			Name: `postgis editors`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsEWKT(ST_Force2D('POINT ZM (1 2 3 4)'::geometry)), ST_AsEWKT(ST_Force3D('POINT(1 2)'::geometry)), ST_AsEWKT(ST_Force3DM('POINT Z (1 2 3)'::geometry, 9)), ST_AsEWKT(ST_Force4D('POINT(1 2)'::geometry, 7, 8));`,
					Expected: []sql.Row{{`POINT(1 2)`, `POINT(1 2 0)`, `POINTM(1 2 9)`, `POINT(1 2 7 8)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Multi('POINT(1 2)'::geometry)), ST_AsEWKT(ST_Multi('CIRCULARSTRING(0 0,1 1,2 0)'::geometry)), ST_AsEWKT(ST_ForceCollection('MULTIPOINT(1 1,2 2)'::geometry));`,
					Expected: []sql.Row{{`MULTIPOINT(1 2)`, `MULTICURVE(CIRCULARSTRING(0 0,1 1,2 0))`, `GEOMETRYCOLLECTION(POINT(1 1),POINT(2 2))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_ForceCurve('LINESTRING(0 0,1 1)'::geometry)), ST_AsEWKT(ST_ForceSFS('TRIANGLE((0 0,1 0,0 1,0 0))'::geometry)), ST_AsEWKT(ST_ForceSFS('TRIANGLE((0 0,1 0,0 1,0 0))'::geometry, '1.2'));`,
					Expected: []sql.Row{{`COMPOUNDCURVE((0 0,1 1))`, `POLYGON((0 0,1 0,0 1,0 0))`, `TRIANGLE((0 0,1 0,0 1,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Reverse('LINESTRING(0 0,1 1,2 3)'::geometry)), ST_AsEWKT(ST_FlipCoordinates('POINT(1 2)'::geometry)), ST_AsEWKT(ST_SwapOrdinates('POINT ZM (1 2 3 4)'::geometry, 'xz'));`,
					Expected: []sql.Row{{`LINESTRING(2 3,1 1,0 0)`, `POINT(2 1)`, `POINT(3 2 1 4)`}},
				},
				{
					Query:           `SELECT ST_SwapOrdinates('POINT(1 2)'::geometry, 'xz');`,
					ExpectedErr:     `Geometry does not have a Z ordinate`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_ForcePolygonCW('POLYGON((0 0,1 0,1 1,0 1,0 0))'::geometry)), ST_AsEWKT(ST_ForcePolygonCCW('POLYGON((0 0,0 1,1 1,1 0,0 0))'::geometry)), ST_AsEWKT(ST_ForceRHR('POLYGON((0 0,1 0,1 1,0 1,0 0))'::geometry));`,
					Expected: []sql.Row{{`POLYGON((0 0,0 1,1 1,1 0,0 0))`, `POLYGON((0 0,1 0,1 1,0 1,0 0))`, `POLYGON((0 0,0 1,1 1,1 0,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_CollectionExtract('GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 0,1 1),POINT(2 2))'::geometry, 1)), ST_AsEWKT(ST_CollectionExtract('GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 0,1 1))'::geometry));`,
					Expected: []sql.Row{{`MULTIPOINT(1 1,2 2)`, `MULTILINESTRING((0 0,1 1))`}},
				},
				{
					Query:           `SELECT ST_CollectionExtract('POINT(1 1)'::geometry, 4);`,
					ExpectedErr:     `ST_CollectionExtract: only point, linestring and polygon may be extracted`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_CollectionHomogenize('GEOMETRYCOLLECTION(POINT(1 1),POINT(2 2))'::geometry)), ST_AsEWKT(ST_CollectionHomogenize('GEOMETRYCOLLECTION(POINT(1 1))'::geometry));`,
					Expected: []sql.Row{{`MULTIPOINT(1 1,2 2)`, `POINT(1 1)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_RemoveRepeatedPoints('LINESTRING(0 0,0 0,1 1,1 1,2 2)'::geometry)), ST_AsEWKT(ST_RemoveRepeatedPoints('LINESTRING(0 0,1 0,2 0,3 0,3.5 0)'::geometry, 1.5));`,
					Expected: []sql.Row{{`LINESTRING(0 0,1 1,2 2)`, `LINESTRING(0 0,3.5 0)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Segmentize('LINESTRING(0 0,3 0)'::geometry, 1)), ST_AsEWKT(ST_Segmentize('LINESTRING Z (0 0 0,2 0 4)'::geometry, 1));`,
					Expected: []sql.Row{{`LINESTRING(0 0,1 0,2 0,3 0)`, `LINESTRING(0 0 0,1 0 2,2 0 4)`}},
				},
				{
					Query:           `SELECT ST_Segmentize('LINESTRING(0 0,1 0)'::geometry, 0);`,
					ExpectedErr:     `ST_Segmentize: invalid max_distance 0 (must be >= 0)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_SnapToGrid('LINESTRING(0 0,0.4 0.4,1.6 1.6,2 2)'::geometry, 1)), ST_AsEWKT(ST_SnapToGrid('POINT(1.4 2.6)'::geometry, 0.5, 2)), ST_AsEWKT(ST_SnapToGrid('POINT(-0.5 2.5)'::geometry, 1));`,
					Expected: []sql.Row{{`LINESTRING(0 0,2 2)`, `POINT(1.5 2)`, `POINT(0 2)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_AddPoint('LINESTRING(0 0,1 1)'::geometry, 'POINT(5 5)'::geometry)), ST_AsEWKT(ST_AddPoint('LINESTRING(0 0,1 1)'::geometry, 'POINT(5 5)'::geometry, 0));`,
					Expected: []sql.Row{{`LINESTRING(0 0,1 1,5 5)`, `LINESTRING(5 5,0 0,1 1)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_RemovePoint('LINESTRING(0 0,1 1,2 2)'::geometry, 1)), ST_AsEWKT(ST_SetPoint('LINESTRING(0 0,1 1,2 2)'::geometry, -1, 'POINT(9 9)'::geometry));`,
					Expected: []sql.Row{{`LINESTRING(0 0,2 2)`, `LINESTRING(0 0,1 1,9 9)`}},
				},
				{
					Query:           `SELECT ST_RemovePoint('LINESTRING(0 0,1 1)'::geometry, 0);`,
					ExpectedErr:     `Can't remove points from a single segment line`,
					ExpectedErrCode: `XX000`,
				},
			},
		},
		{
			Name: `postgis affine transformations`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsEWKT(ST_Translate('POINT(1 2)'::geometry, 1, 2)), ST_AsEWKT(ST_Translate('POINT Z (1 2 3)'::geometry, 1, 2, 3));`,
					Expected: []sql.Row{{`POINT(2 4)`, `POINT(2 4 6)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Scale('POINT(1 2)'::geometry, 2, 3)), ST_AsEWKT(ST_Scale('POINT ZM (1 2 3 4)'::geometry, 'POINT ZM (2 3 4 5)'::geometry)), ST_AsEWKT(ST_Scale('POINT(1 2)'::geometry, 'POINT(2 2)'::geometry, 'POINT(1 1)'::geometry));`,
					Expected: []sql.Row{{`POINT(2 6)`, `POINT(2 6 12 20)`, `POINT(1 3)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Affine('POINT(1 2)'::geometry, 1, 2, 3, 4, 5, 6)), ST_AsEWKT(ST_Affine('POINT Z (1 2 3)'::geometry, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12));`,
					Expected: []sql.Row{{`POINT(10 17)`, `POINT(24 43 62)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_TransScale('POINT(1 2)'::geometry, 1, 2, 3, 4)), ST_AsEWKT(ST_Rotate('POINT(1 2)'::geometry, pi()));`,
					Expected: []sql.Row{{`POINT(6 16)`, `POINT(-1 -2)`}},
				},
				{
					Query:           `SELECT ST_Scale('POINT(1 2)'::geometry, 'LINESTRING(0 0,1 1)'::geometry);`,
					ExpectedErr:     `Scale factor geometry parameter must be a point`,
					ExpectedErrCode: `XX000`,
				},
			},
		},
		{
			Name: `postgis linear referencing`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsEWKT(ST_LineInterpolatePoint('LINESTRING(0 0,1 0,1 2)'::geometry, 0.25)), ST_AsEWKT(ST_LineInterpolatePoint('LINESTRING ZM (0 0 0 0,4 0 4 8)'::geometry, 0.5));`,
					Expected: []sql.Row{{`POINT(0.75 0)`, `POINT(2 0 2 4)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_LineInterpolatePoints('LINESTRING(0 0,4 0)'::geometry, 0.25)), ST_AsEWKT(ST_LineInterpolatePoints('LINESTRING(0 0,4 0)'::geometry, 0.25, false));`,
					Expected: []sql.Row{{`MULTIPOINT(1 0,2 0,3 0,4 0)`, `POINT(1 0)`}},
				},
				{
					Query:           `SELECT ST_LineInterpolatePoint('LINESTRING(0 0,1 1)'::geometry, 1.5);`,
					ExpectedErr:     `line_interpolate_point: 2nd arg isn't within [0,1]`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_LineLocatePoint('LINESTRING(0 0,4 0)'::geometry, 'POINT(1 1)'::geometry), ST_LineLocatePoint('LINESTRING(0 0,1 0,1 2)'::geometry, 'POINT(2 1)'::geometry);`,
					Expected: []sql.Row{{0.25, 0.6666666666666666}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_LineSubstring('LINESTRING(0 0,1 0,1 2,3 2)'::geometry, 0.25, 0.75)), ST_AsEWKT(ST_LineSubstring('MULTILINESTRING((0 0,1 0),(5 5,6 5))'::geometry, 0.5, 0.5));`,
					Expected: []sql.Row{{`LINESTRING(1 0.25,1 2,1.75 2)`, `GEOMETRYCOLLECTION(POINT(1 0),POINT(5 5))`}},
				},
				{
					Query:           `SELECT ST_LineSubstring('LINESTRING(0 0,1 0)'::geometry, 0.7, 0.2);`,
					ExpectedErr:     `2nd arg must be smaller then 3rd arg`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_AddMeasure('LINESTRING(0 0,1 0,1 2,3 2)'::geometry, 1, 9)), ST_AsEWKT(ST_LocateAlong('LINESTRING M (0 0 0,4 0 4,4 4 8)'::geometry, 3));`,
					Expected: []sql.Row{{`LINESTRINGM(0 0 1,1 0 2.6,1 2 5.8,3 2 9)`, `MULTIPOINTM(3 0 3)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_LocateBetween('LINESTRING M (0 0 0,4 0 4,0 0 0)'::geometry, 1, 3)), ST_AsEWKT(ST_LocateBetweenElevations('LINESTRING Z (0 0 0,4 0 4,4 4 8)'::geometry, 1, 3));`,
					Expected: []sql.Row{{`MULTILINESTRINGM((1 0 1,3 0 3),(3 0 3,1 0 1))`, `MULTILINESTRING((1 0 1,3 0 3))`}},
				},
				{
					Query:           `SELECT ST_LocateAlong('LINESTRING(0 0,1 1)'::geometry, 3);`,
					ExpectedErr:     `Input geometry does not have a measure dimension`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_InterpolatePoint('LINESTRING M (0 0 0,4 0 4,4 4 8)'::geometry, 'POINT(1 1)'::geometry);`,
					Expected: []sql.Row{{1.0}},
				},
			},
		},
		{
			Name: `postgis output formats`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsGeoJSON('POINT(1 2)'::geometry), ST_AsGeoJSON('SRID=3857;LINESTRING(1 2,3 4)'::geometry), ST_AsGeoJSON('SRID=4326;LINESTRING(1.123456 2,3 4)'::geometry, 3, 1);`,
					Expected: []sql.Row{{`{"type":"Point","coordinates":[1,2]}`, `{"type":"LineString","crs":{"type":"name","properties":{"name":"EPSG:3857"}},"coordinates":[[1,2],[3,4]]}`, `{"type":"LineString","bbox":[1.123,2.000,3.000,4.000],"coordinates":[[1.123,2],[3,4]]}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON('POLYGON((0 0,0 1,1 1,0 0))'::geometry), ST_AsGeoJSON('GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 0,1 1))'::geometry);`,
					Expected: []sql.Row{{`{"type":"Polygon","coordinates":[[[0,0],[0,1],[1,1],[0,0]]]}`, `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,1]},{"type":"LineString","coordinates":[[0,0],[1,1]]}]}`}},
				},
				{
					Query:           `SELECT ST_AsGeoJSON('CIRCULARSTRING(0 0,1 1,2 0)'::geometry);`,
					ExpectedErr:     `lwgeom_to_geojson: 'CircularString' geometry type not supported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT 'SRID=4326;POINT(1 2)'::geometry::json::text;`,
					Expected: []sql.Row{{`{"type":"Point","crs":{"type":"name","properties":{"name":"EPSG:4326"}},"coordinates":[1,2]}`}},
				},
				{
					Query:    `SELECT ST_AsKML('SRID=4326;POLYGON((0 0,0 1,1 1,0 0))'::geometry), ST_AsKML('SRID=4326;POINT Z (1 2 3)'::geometry, 15, 'kml');`,
					Expected: []sql.Row{{`<Polygon><outerBoundaryIs><LinearRing><coordinates>0,0 0,1 1,1 0,0</coordinates></LinearRing></outerBoundaryIs></Polygon>`, `<kml:Point><kml:coordinates>1,2,3</kml:coordinates></kml:Point>`}},
				},
				{
					Query:           `SELECT ST_AsKML('POINT(1 2)'::geometry);`,
					ExpectedErr:     `ST_AsKML: Input geometry has unknown (0) SRID`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsSVG('POINT(1 2)'::geometry), ST_AsSVG('LINESTRING(0 0,1 1,2 3)'::geometry, 1), ST_AsSVG('POLYGON((0 0,0 1,1 1,0 0))'::geometry);`,
					Expected: []sql.Row{{`cx="1" cy="-2"`, `M 0 0 l 1 -1 1 -2`, `M 0 0 L 0 -1 1 -1 Z`}},
				},
				{
					Query:    `SELECT ST_GeoHash('POINT(-126 48)'::geometry), ST_GeoHash('POINT(-126 48)'::geometry, 5), ST_AsEWKT(ST_PointFromGeoHash('9qqj7')), ST_Box2dFromGeoHash('9qqj7');`,
					Expected: []sql.Row{{`c0w3hf1s70w3hf1s70w3`, `c0w3h`, `POINT(-115.15869140625 36.10107421875)`, `BOX(-115.1806640625 36.0791015625,-115.13671875 36.123046875)`}},
				},
				{
					Query:    "SELECT ST_AsEncodedPolyline('SRID=4326;LINESTRING(-120.2 38.5,-120.95 40.7,-126.453 43.252)'::geometry), ST_AsEWKT(ST_LineFromEncodedPolyline('_p~iF~ps|U_ulLnnqC_mqNvxq`@'));",
					Expected: []sql.Row{{"_p~iF~ps|U_ulLnnqC_mqNvxq`@", `SRID=4326;LINESTRING(-120.2 38.5,-120.95 40.7,-126.453 43.252)`}},
				},
				{
					Query:    `SELECT ST_AsLatLonText('POINT(-3.2342342 -2.32498)'::geometry), ST_AsLatLonText('POINT(-302.2342342 -792.32498)'::geometry, 'C DD MM SS');`,
					Expected: []sql.Row{{`2°19'29.928"S 3°14'3.243"W`, `S 72 19 30 E 57 45 57`}},
				},
				{
					Query:           `SELECT ST_AsLatLonText('POINT(1 2)'::geometry, 'M S');`,
					ExpectedErr:     `Bad format, degrees (DD.DDD) must be included.`,
					ExpectedErrCode: `XX000`,
				},
			},
		},
		{
			Name: `postgis aggregates`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
				`CREATE TABLE geo_agg (id INT PRIMARY KEY, g geometry);`,
				`INSERT INTO geo_agg VALUES (1, 'POINT(1 1)'), (2, 'LINESTRING(0 0,2 2)'), (3, NULL), (4, 'POLYGON((0 0,0 3,3 3,3 0,0 0))');`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_Extent(g), ST_3DExtent(g) FROM geo_agg WHERE id < 3;`,
					Expected: []sql.Row{{`BOX(0 0,2 2)`, `BOX3D(0 0 0,2 2 5e-324)`}},
					Skip:     true,
				},
				{
					Query:    `SELECT ST_Extent(g) FROM geo_agg;`,
					Expected: []sql.Row{{`BOX(0 0,3 3)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_MemUnion(g)) FROM geo_agg WHERE id IN (1, 4);`,
					Expected: []sql.Row{{`POLYGON((0 3,3 3,3 0,0 0,0 3))`}},
				},
				{
					Query:    `SELECT ST_Extent(g) FROM geo_agg WHERE id > 100;`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Union(g)) FROM geo_agg;`,
					Expected: []sql.Row{{`POLYGON((0 3,3 3,3 0,0 0,0 3))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Collect(g)) FROM geo_agg WHERE id IN (1, 2);`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 0,2 2))`}},
				},
			},
		},
		{
			Name: `postgis dumps`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: `SELECT path::text, ST_AsEWKT(geom) FROM ST_Dump('SRID=4326;GEOMETRYCOLLECTION(POINT(1 1),MULTIPOINT(2 2,3 3))'::geometry);`,
					Expected: []sql.Row{
						{`{1}`, `SRID=4326;POINT(1 1)`},
						{`{2,1}`, `SRID=4326;POINT(2 2)`},
						{`{2,2}`, `SRID=4326;POINT(3 3)`},
					},
				},
				{
					Query:    `SELECT path::text, ST_AsEWKT(geom) FROM ST_Dump('POINT(1 2)'::geometry);`,
					Expected: []sql.Row{{`{}`, `POINT(1 2)`}},
				},
				{
					Query: `SELECT path::text, ST_AsEWKT(geom) FROM ST_DumpPoints('POLYGON((0 0,0 1,1 1,0 0))'::geometry);`,
					Expected: []sql.Row{
						{`{1,1}`, `POINT(0 0)`},
						{`{1,2}`, `POINT(0 1)`},
						{`{1,3}`, `POINT(1 1)`},
						{`{1,4}`, `POINT(0 0)`},
					},
				},
				{
					Query: `SELECT path::text, ST_AsEWKT(geom) FROM ST_DumpRings('POLYGON((0 0,0 1,1 1,0 0),(0.1 0.1,0.2 0.1,0.2 0.2,0.1 0.1))'::geometry);`,
					Expected: []sql.Row{
						{`{0}`, `POLYGON((0 0,0 1,1 1,0 0))`},
						{`{1}`, `POLYGON((0.1 0.1,0.2 0.1,0.2 0.2,0.1 0.1))`},
					},
				},
				{
					Query: `SELECT path::text, ST_AsEWKT(geom) FROM ST_DumpSegments('LINESTRING(0 0,1 1,2 2)'::geometry);`,
					Expected: []sql.Row{
						{`{1}`, `LINESTRING(0 0,1 1)`},
						{`{2}`, `LINESTRING(1 1,2 2)`},
					},
				},
				{
					Query:           `SELECT path::text, ST_AsEWKT(geom) FROM ST_DumpRings('LINESTRING(0 0,1 1)'::geometry);`,
					ExpectedErr:     `Input is not a polygon`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query: `SELECT ST_AsEWKT((ST_Dump('MULTIPOINT(1 1,2 2)'::geometry)).geom);`,
					Expected: []sql.Row{
						{`POINT(1 1)`},
						{`POINT(2 2)`},
					},
				},
			},
		},
		{
			Name: `postgis versions`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT postgis_lib_version(), postgis_scripts_installed();`,
					Expected: []sql.Row{{`3.6.4`, `3.6.4 94d984b`}},
				},
				{
					Query:    `SELECT postgis_version();`,
					Expected: []sql.Row{{`3.6 USE_GEOS=1 USE_PROJ=1 USE_STATS=1`}},
				},
				{
					Query:    `SELECT postgis_type_name('POINT', 3), postgis_type_name('POINTM', 3), postgis_type_name('POINT', 3, false);`,
					Expected: []sql.Row{{`PointZ`, `PointM`, `POINT`}},
				},
				{
					Query:    `SELECT postgis_lib_revision(), postgis_scripts_released(), _postgis_pgsql_version(), _postgis_scripts_pgsql_version();`,
					Expected: []sql.Row{{`94d984b`, `3.6.4 94d984b`, `150`, `150`}},
				},
				{
					Query:    `SELECT postgis_svn_version();`,
					Expected: []sql.Row{{`94d984b`}},
				},
				{
					Query:    `SELECT _postgis_deprecate('old_name', 'new_name', '3.6.0') IS NULL, _postgis_deprecate('old_name', 'new_name', '2.5.0') IS NULL;`,
					Expected: []sql.Row{{`f`, `f`}},
				},
			},
		},
		{
			Name: `postgis grid cells`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsEWKT(ST_TileEnvelope(2, 1, 1)), ST_AsEWKT(ST_TileEnvelope(3, 1, 1, 'SRID=4326;LINESTRING(-180 -90,180 90)'::geometry));`,
					Expected: []sql.Row{{`SRID=3857;POLYGON((-10018754.171394622 0,-10018754.171394622 10018754.171394622,0 10018754.171394622,0 0,-10018754.171394622 0))`, `SRID=4326;POLYGON((-135 45,-135 67.5,-90 67.5,-90 45,-135 45))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_TileEnvelope(2, 1, 1, 'SRID=3857;LINESTRING(-20037508.342789244 -20037508.342789244,20037508.342789244 20037508.342789244)'::geometry, 0.125));`,
					Expected: []sql.Row{{`SRID=3857;POLYGON((-11271098.442818949 -1252344.271424327,-11271098.442818949 11271098.442818949,1252344.271424327 11271098.442818949,1252344.271424327 -1252344.271424327,-11271098.442818949 -1252344.271424327))`}},
				},
				{
					Query:           `SELECT ST_TileEnvelope(2, 5, 1);`,
					ExpectedErr:     `ST_TileEnvelope: Invalid tile x value, 5`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_TileEnvelope(-1, 0, 0);`,
					ExpectedErr:     `ST_TileEnvelope: Invalid tile zoom value, -1`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Square(1, 0, 0)), ST_AsEWKT(ST_Square(2, 1, -1, 'SRID=3857;POINT(10 10)'::geometry));`,
					Expected: []sql.Row{{`POLYGON((0 0,0 1,1 1,1 0,0 0))`, `SRID=3857;POLYGON((12 8,12 10,14 10,14 8,12 8))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Hexagon(1, 0, 0)), ST_AsEWKT(ST_Hexagon(1, 1, 1));`,
					Expected: []sql.Row{{`POLYGON((-1 0,-0.5 -0.866025403784439,0.5 -0.866025403784439,1 0,0.5 0.866025403784439,-0.5 0.866025403784439,-1 0))`, `POLYGON((0.5 2.598076211353316,1 1.732050807568877,2 1.732050807568877,2.5 2.598076211353316,2 3.464101615137755,1 3.464101615137755,0.5 2.598076211353316))`}},
				},
			},
		},
		{
			Name: `postgis reshaping`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsEWKT(ST_ShiftLongitude('SRID=4326;POINT(-170 10)'::geometry)), ST_AsEWKT(ST_ShiftLongitude('LINESTRING(170 0,-170 0)'::geometry));`,
					Expected: []sql.Row{{`SRID=4326;POINT(190 10)`, `LINESTRING(170 0,190 0)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_WrapX('LINESTRING(-10 0,10 0)'::geometry, 0, 360)), ST_AsEWKT(ST_WrapX('POLYGON((-10 0,-10 1,-5 1,-5 0,-10 0))'::geometry, 0, 360));`,
					Expected: []sql.Row{{`MULTILINESTRING((350 0,360 0),(0 0,10 0))`, `POLYGON((350 0,350 1,355 1,355 0,350 0))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Scroll('LINESTRING(0 0,1 1,2 2,0 0)'::geometry, 'POINT(1 1)'::geometry));`,
					Expected: []sql.Row{{`LINESTRING(1 1,2 2,0 0,1 1)`}},
				},
				{
					Query:           `SELECT ST_Scroll('LINESTRING(0 0,1 1,2 2)'::geometry, 'POINT(1 1)'::geometry);`,
					ExpectedErr:     `ptarray_scroll_in_place: input POINTARRAY is not closed`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_FilterByM('LINESTRING M (0 0 1,1 1 2,2 2 3,3 3 4)'::geometry, 2, 3)), ST_AsEWKT(ST_FilterByM('LINESTRING M (0 0 1,1 1 2,2 2 3,3 3 4)'::geometry, 2, 3, true));`,
					Expected: []sql.Row{{`LINESTRING(1 1,2 2)`, `LINESTRINGM(1 1 2,2 2 3)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_LineExtend('LINESTRING(0 0,1 0)'::geometry, 1, 2)), ST_AsEWKT(ST_LineExtend('LINESTRING Z (0 0 0,3 4 5)'::geometry, 5, 5));`,
					Expected: []sql.Row{{`LINESTRING(-2 0,0 0,1 0,2 0)`, `LINESTRING(-3 -4 -5,0 0 0,3 4 5,6 8 10)`}},
				},
				{
					Query:           `SELECT ST_LineExtend('LINESTRING(0 0,1 0)'::geometry, -1);`,
					ExpectedErr:     `lwline_extend: distances must be non-negative`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_BoundingDiagonal('LINESTRING(0 0,1 2,3 1)'::geometry)), ST_AsEWKT(ST_BoundingDiagonal('LINESTRING Z (0 0 5,1 2 3)'::geometry));`,
					Expected: []sql.Row{{`LINESTRING(0 0,3 2)`, `LINESTRING(0 0 3,1 2 5)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_ChaikinSmoothing('LINESTRING(0 0,8 8,0 16)'::geometry)), ST_AsEWKT(ST_ChaikinSmoothing('POLYGON((0 0,8 8,0 16,0 0))'::geometry)), ST_AsEWKT(ST_ChaikinSmoothing('POLYGON((0 0,8 8,0 16,0 0))'::geometry, 1, true));`,
					Expected: []sql.Row{{`LINESTRING(0 0,6 6,6 10,0 16)`, `POLYGON((2 2,6 6,6 10,2 14,0 12,0 4,2 2))`, `POLYGON((0 0,6 6,6 10,2 14,0 12,0 0))`}},
				},
				{
					Query:           `SELECT ST_ChaikinSmoothing('LINESTRING(0 0,8 8,0 16)'::geometry, 6);`,
					ExpectedErr:     `Number of iterations must be between 1 and 5 : LWGEOM_ChaikinSmoothing`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_SimplifyVW('LINESTRING(5 2,3 8,6 20,7 25,10 10)'::geometry, 30)), ST_AsEWKT(ST_SimplifyVW('POLYGON((0 0,0 10,1 11,10 10,10 0,0 0))'::geometry, 10));`,
					Expected: []sql.Row{{`LINESTRING(5 2,7 25,10 10)`, `POLYGON((0 0,1 11,10 10,10 0,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_SetEffectiveArea('LINESTRING(5 2,3 8,6 20,7 25,10 10)'::geometry));`,
					Expected: []sql.Row{{`LINESTRINGM(5 2 3.402823466385289e+38,3 8 29,6 20 1.5,7 25 49.5,10 10 3.402823466385289e+38)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_QuantizeCoordinates('POINT(100.123456 0.123456)'::geometry, 2)), ST_AsEWKT(ST_QuantizeCoordinates('POINT Z (100.123456 0.123456 3.14159)'::geometry, 1, 2, 3));`,
					Expected: []sql.Row{{`POINT(100.12109375 0.12109375)`, `POINT(100.09375 0.12109375 3.14111328125)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_RemoveSmallParts('MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((0 0,0 10,10 10,10 0,0 0)))'::geometry, 5, 5)), ST_AsEWKT(ST_RemoveSmallParts('POLYGON((0 0,0 10,10 10,10 0,0 0),(1 1,1 2,2 2,2 1,1 1))'::geometry, 5, 5));`,
					Expected: []sql.Row{{`MULTIPOLYGON(((0 0,0 10,10 10,10 0,0 0)))`, `POLYGON((0 0,0 10,10 10,10 0,0 0))`}},
				},
			},
		},
		{
			Name: `postgis trajectories and crossings`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_IsValidTrajectory('LINESTRING M (0 0 1,0 1 2)'::geometry), ST_IsValidTrajectory('LINESTRING M (0 0 1,0 1 1)'::geometry);`,
					Expected: []sql.Row{{`t`, `f`}},
				},
				{
					Query:    `SELECT ST_ClosestPointOfApproach('LINESTRING M (0 0 0,10 0 10)'::geometry, 'LINESTRING M (0 5 0,10 -5 10)'::geometry), ST_DistanceCPA('LINESTRING M (0 0 0,10 0 10)'::geometry, 'LINESTRING M (0 5 0,10 -5 10)'::geometry), ST_CPAWithin('LINESTRING M (0 0 0,10 0 10)'::geometry, 'LINESTRING M (0 5 0,10 -5 10)'::geometry, 1);`,
					Expected: []sql.Row{{5.0, 0.0, `t`}},
				},
				{
					Query:    `SELECT ST_ClosestPointOfApproach('LINESTRING M (0 0 0,10 0 10)'::geometry, 'LINESTRING M (0 5 20,10 -5 30)'::geometry);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:           `SELECT ST_ClosestPointOfApproach('LINESTRING M (0 0 0,10 0 10)'::geometry, 'LINESTRING(0 0,1 1)'::geometry);`,
					ExpectedErr:     `Both input geometries must have a measure dimension`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_LineCrossingDirection('LINESTRING(0 0,10 10)'::geometry, 'LINESTRING(0 10,10 0)'::geometry), ST_LineCrossingDirection('LINESTRING(0 10,10 0)'::geometry, 'LINESTRING(0 0,10 10)'::geometry), ST_LineCrossingDirection('LINESTRING(0 0,10 0)'::geometry, 'LINESTRING(0 5,10 5)'::geometry);`,
					Expected: []sql.Row{{1, -1, 0}},
				},
				{
					Query:    `SELECT ST_LineCrossingDirection('LINESTRING(0 0,5 5,0 10)'::geometry, 'LINESTRING(3 0,3 10)'::geometry), ST_LineCrossingDirection('LINESTRING(0 0,5 5,0 10,5 15)'::geometry, 'LINESTRING(3 0,3 15)'::geometry);`,
					Expected: []sql.Row{{-3, -2}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeometricMedian('MULTIPOINT((0 0),(10 0))'::geometry)), ST_AsEWKT(ST_GeometricMedian('POINT(1 1)'::geometry));`,
					Expected: []sql.Row{{`POINT(5 0)`, `POINT(1 1)`}},
				},
				{
					Query:           `SELECT ST_GeometricMedian('LINESTRING(0 0,1 1)'::geometry);`,
					ExpectedErr:     `lwgeom_median: Unsupported geometry type: LineString`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_OrientedEnvelope('POLYGON((0 0,4 0,4 1,0 1,0 0))'::geometry)), ST_AsEWKT(ST_OrientedEnvelope('MULTIPOINT(0 0,2 2)'::geometry));`,
					Expected: []sql.Row{{`POLYGON((0 1,0 0,4 0,4 1,0 1))`, `LINESTRING(0 0,2 2)`}},
				},
			},
		},
		{
			Name: `postgis geography extras`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsEWKT(ST_Segmentize('LINESTRING Z (0 0 0,10 0 10)'::geography, 400000));`,
					Expected: []sql.Row{{`SRID=4326;LINESTRING(0 0 0,2.5 0 2.5,4.999999999999999 0 5,7.500000000000001 0 7.5,10 0 10)`}},
					Skip:     true,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Centroid('POINT(1 1)'::geography)), GeometryType('POINT(1 1)'::geography);`,
					Expected: []sql.Row{{`SRID=4326;POINT(1 1)`, `POINT`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_ClosestPoint('LINESTRING(0 0,10 0)'::geography, 'POINT(5 1)'::geography));`,
					Expected: []sql.Row{{`SRID=4326;POINT(5 0)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_LineInterpolatePoint('LINESTRING(0 0,10 10)'::geography, 0.5));`,
					Expected: []sql.Row{{`SRID=4326;POINT(4.961882922527124 5.01950880089445)`}},
				},
			},
		},
		{
			Name: `postgis validity`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
				`CREATE TABLE geo_valid (id INT PRIMARY KEY, g geometry);`,
				`INSERT INTO geo_valid VALUES (1, 'POLYGON((0 0,10 0,10 10,0 10,0 0),(15 15,16 15,16 16,15 15))'), (2, 'LINESTRING(0 0,0 0)'), (3, 'POLYGON((0 0,10 0,10 10,0 10,0 0),(1 1,9 1,9 9,1 9,1 1),(2 2,3 2,3 3,2 3,2 2))'), (4, 'POLYGON((0 0,10 0,10 10,0 10,0 0),(5 5,15 5,15 6,5 6,5 5))'), (5, 'POLYGON((0 0,1 1,1 1,0 0))'), (6, 'POLYGON((0 0,10 0,10 10,0 10,0 0),(0 0,5 1,5 5,0 0))'), (7, 'POINT(1 1)'), (8, 'POLYGON EMPTY'), (9, 'MULTIPOLYGON(((0 0,10 0,10 10,0 10,0 0),(2 2,8 2,8 8,2 8,2 2)),((3 3,4 3,4 4,3 4,3 3)))');`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: `SELECT id, ST_IsValid(g), ST_IsValidReason(g), ST_IsValidDetail(g), ST_IsValid(g, 1), ST_IsValidReason(g, 1), ST_IsValidDetail(g, 1) FROM geo_valid ORDER BY id;`,
					Expected: []sql.Row{
						{1, `f`, `Hole lies outside shell[15 15]`, `(f,"Hole lies outside shell",01010000000000000000002E400000000000002E40)`, `f`, `Hole lies outside shell`, `(f,"Hole lies outside shell",01010000000000000000002E400000000000002E40)`},
						{2, `f`, `Too few points in geometry component[0 0]`, `(f,"Too few points in geometry component",010100000000000000000000000000000000000000)`, `f`, `Too few points in geometry component`, `(f,"Too few points in geometry component",010100000000000000000000000000000000000000)`},
						{3, `f`, `Holes are nested[2 2]`, `(f,"Holes are nested",010100000000000000000000400000000000000040)`, `f`, `Holes are nested`, `(f,"Holes are nested",010100000000000000000000400000000000000040)`},
						{4, `f`, `Self-intersection[10 5]`, `(f,Self-intersection,010100000000000000000024400000000000001440)`, `f`, `Self-intersection`, `(f,Self-intersection,010100000000000000000024400000000000001440)`},
						{5, `f`, `Too few points in geometry component[0 0]`, `(f,"Too few points in geometry component",010100000000000000000000000000000000000000)`, `f`, `Too few points in geometry component`, `(f,"Too few points in geometry component",010100000000000000000000000000000000000000)`},
						{6, `t`, `Valid Geometry`, `(t,,)`, `t`, `Valid Geometry`, `(t,,)`},
						{7, `t`, `Valid Geometry`, `(t,,)`, `t`, `Valid Geometry`, `(t,,)`},
						{8, `t`, `Valid Geometry`, `(t,,)`, `t`, `Valid Geometry`, `(t,,)`},
						{9, `t`, `Valid Geometry`, `(t,,)`, `t`, `Valid Geometry`, `(t,,)`},
					},
				},
				{
					Query:    `SELECT ST_IsValidReason('LINESTRING(1 1,1 1,1 1)'), ST_IsValidReason('MULTIPOLYGON(((0 0,1 0,1 1,0 0)),((5 5,6 6,6 6,5 5)))'), ST_IsValidReason('CIRCULARSTRING(0 0,1 1,0 0)'), ST_IsValidReason('POLYGON((0 0,1 0,1 1,0 1,0 0),(0.5 0.5,0.5 0.5,0.5 0.5,0.5 0.5))');`,
					Expected: []sql.Row{{`Too few points in geometry component[1 1]`, `Too few points in geometry component[5 5]`, `Valid Geometry`, `Too few points in geometry component[0.5 0.5]`}},
				},
				{
					Query:    `SELECT ST_IsValidReason('POINT(NaN 1)'::geometry), ST_IsValidReason(NULL::geometry), ST_IsValidDetail('SRID=4326;POLYGON((0 0,1 1,1 1,0 0))');`,
					Expected: []sql.Row{{`Invalid Coordinate[nan 1]`, nil, `(f,"Too few points in geometry component",010100000000000000000000000000000000000000)`}},
				},
			},
		},
		{
			Name: `postgis 3d lengths`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_3DLength('LINESTRING Z (0 0 0,1 1 1,2 0 3)'), ST_3DLength('LINESTRING(0 0,3 4)'), ST_3DLength('LINESTRING M (0 0 5,3 4 9)'), ST_3DLength('MULTILINESTRING Z ((0 0 0,0 0 1),(1 1 1,2 2 2))');`,
					Expected: []sql.Row{{4.1815405503520555, 5.0, 5.0, 2.732050807568877}},
				},
				{
					Query:    `SELECT ST_3DLength('POLYGON Z ((0 0 0,1 0 1,1 1 1,0 0 0))'), ST_3DLength('POINT Z (1 2 3)'), ST_3DLength('LINESTRING EMPTY'), ST_3DLength('GEOMETRYCOLLECTION Z (LINESTRING Z (0 0 0,0 0 2),POINT Z (1 1 1))');`,
					Expected: []sql.Row{{0.0, 0.0, 0.0, 2.0}},
				},
				{
					Query:    `SELECT ST_3DLength('CIRCULARSTRING Z (0 0 0,1 1 1,2 0 2)'), ST_3DLength('COMPOUNDCURVE Z ((0 0 0,1 0 5),CIRCULARSTRING Z (1 0 5,2 1 5,3 0 5))'), ST_3DLength('MULTICURVE Z ((0 0 0,0 0 3),CIRCULARSTRING Z (0 0 0,1 1 1,2 0 2))');`,
					Expected: []sql.Row{{3.141592653589793, 4.141592653589793, 6.141592653589793}},
				},
				{
					Query:    `SELECT ST_3DPerimeter('POLYGON Z ((0 0 0,1 0 1,1 1 1,0 0 0))'), ST_3DPerimeter('POLYGON((0 0,3 0,3 4,0 0))'), ST_3DPerimeter('MULTIPOLYGON Z (((0 0 0,1 0 0,1 1 1,0 0 0)),((5 5 5,6 5 5,6 6 6,5 5 5),(5.2 5.1 5,5.3 5.1 5,5.3 5.2 5,5.2 5.1 5)))'), ST_3DPerimeter('LINESTRING Z (0 0 0,1 1 1)');`,
					Expected: []sql.Row{{4.146264369941973, 12.0, 8.633950096121255, 0.0}},
				},
				{
					Query:    `SELECT ST_3DPerimeter('CURVEPOLYGON Z (COMPOUNDCURVE Z ((0 0 0,2 0 3),CIRCULARSTRING Z (2 0 3,1 1 3,0 0 0)))'), ST_3DPerimeter('CURVEPOLYGON Z ((0 0 0,1 0 5,1 1 5,0 0 0))'), ST_3DPerimeter('TRIANGLE Z ((0 0 0,1 0 1,1 1 1,0 0 0))'), ST_3DPerimeter('TIN Z (((0 0 0,1 0 1,1 1 1,0 0 0)))');`,
					Expected: []sql.Row{{5.141592653589793, 11.295171936299417, 4.146264369941973, 4.146264369941973}},
				},
				{
					Query:    `SELECT ST_3DLength('LINESTRING Z (0.1 0.2 0.3,1.7 2.9 3.3,-4.1 5.3 0.7)'), ST_3DLength('LINESTRING (0.11 0.27,1.73 2.91,-4.13 5.37,8.1 -3.3)'), ST_3DPerimeter('POLYGON ((0.13 0.29,1.71 2.97,-4.17 5.31,0.13 0.29))');`,
					Expected: []sql.Row{{11.135773769803691, 24.444216289226603, 16.049455322566786}},
				},
				{
					Query:    `SELECT ST_Length('LINESTRING (0.1 0.2,1.7 2.9,-4.1 5.3)'::geometry), ST_Length('CIRCULARSTRING(0.1 0.2,1.7 2.9,3.3 5.6)'::geometry), ST_Length('CIRCULARSTRING(0.1 0.2,1.7 2.9,0.1 0.2)'::geometry);`,
					Expected: []sql.Row{{9.41541289588513, 6.276941930590086, 9.859797328075773}},
				},
			},
		},
		{
			Name: `postgis spheroids`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
				`CREATE TABLE geo_spheroid (id INT PRIMARY KEY, s spheroid);`,
				`INSERT INTO geo_spheroid VALUES (1, 'SPHEROID["WGS 84",6378137,298.257223563]'), (2, 'SPHEROID["GRS_1980",6378137,298.257222101]'), (3, 'SPHEROID["Clarke 1866",6378206.4,294.9786982]');`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: `SELECT id, s FROM geo_spheroid ORDER BY id;`,
					Expected: []sql.Row{
						{1, `SPHEROID("WGS 84",6378137,298.257223563)`},
						{2, `SPHEROID("GRS_1980",6378137,298.257222101)`},
						{3, `SPHEROID("Clarke 1866",6378206.4,294.9786982)`},
					},
				},
				{
					Query:    `SELECT 'SPHEROID["x",6378137.5,298.25,]'::spheroid, 'SPHEROID["x",1e300,1.5e-3]'::spheroid, 'SPHEROID["x",1,2] junk'::spheroid, 'SPHEROID["nineteen characters",6378137,298.257223563]'::spheroid;`,
					Expected: []sql.Row{{`SPHEROID("x",6378137.5,298.25)`, `SPHEROID("x",1e+300,0.0015)`, `SPHEROID("x",1,2)`, `SPHEROID("nineteen characters",6378137,298.257223563)`}},
				},
				{
					Query:           `SELECT 'spheroid["x",1,2]'::spheroid;`,
					ExpectedErr:     `SPHEROID parser - doesn't start with SPHEROID`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT 'SPHEROID["",1,2]'::spheroid;`,
					ExpectedErr:     `SPHEROID parser - couldn't parse the spheroid`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT 'SPHEROID["a very long spheroid name over twenty chars",6378137,298.257223563]'::spheroid;`,
					ExpectedErr:     `SPHEROID parser - couldn't parse the spheroid`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT 'SPHEROID["x", 6378137 , 298.5 ]'::spheroid;`,
					ExpectedErr:     `SPHEROID parser - couldn't parse the spheroid`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query: `SELECT id, round(ST_LengthSpheroid('LINESTRING(-118.58 38.38,-118.20 38.45,-117.9 37.1)', s)::numeric, 4), round(ST_Length2DSpheroid('LINESTRING Z (-118.58 38.38 0,-118.20 38.45 1000,-117.9 37.1 0)', s)::numeric, 4) FROM geo_spheroid ORDER BY id;`,
					Expected: []sql.Row{
						{1, framework.Numeric(`186237.2028`), framework.Numeric(`186255.1550`)},
						{2, framework.Numeric(`186237.2028`), framework.Numeric(`186255.1550`)},
						{3, framework.Numeric(`186234.8519`), framework.Numeric(`186252.8039`)},
					},
				},
				{
					Query: `SELECT id, round(ST_LengthSpheroid('GEOMETRYCOLLECTION(LINESTRING(0 0,1 1),POLYGON((0 0,1 0,1 1,0 0)),POINT(5 5))', s)::numeric, 4), round(ST_LengthSpheroid('TRIANGLE((0 0,1 0,1 1,0 0))', s)::numeric, 4), ST_LengthSpheroid('CURVEPOLYGON((0 0,1 0,1 1,0 0))', s), ST_LengthSpheroid('LINESTRING EMPTY', s) FROM geo_spheroid ORDER BY id;`,
					Expected: []sql.Row{
						{1, framework.Numeric(`535693.0159`), framework.Numeric(`378793.4476`), 0.0, 0.0},
						{2, framework.Numeric(`535693.0159`), framework.Numeric(`378793.4476`), 0.0, 0.0},
						{3, framework.Numeric(`535678.9253`), framework.Numeric(`378783.4758`), 0.0, 0.0},
					},
				},
				{
					Query:           `SELECT ST_LengthSpheroid('CIRCULARSTRING(0 0,1 1,2 0)', 'SPHEROID["WGS 84",6378137,298.257223563]');`,
					ExpectedErr:     `unsupported type passed to lwgeom_length_spheroid`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query: `SELECT id, round(ST_DistanceSpheroid('POINT(-118.58 38.38)', 'POINT(-117.9 37.1)', s)::numeric, 4), round(ST_DistanceSpheroid('POINT(-118.58 38.38)', 'LINESTRING(-117.9 37.1,-116 39)', s)::numeric, 4), ST_DistanceSpheroid('POLYGON((0 0,1 0,1 1,0 1,0 0))', 'POINT(0.5 0.5)', s) FROM geo_spheroid ORDER BY id;`,
					Expected: []sql.Row{
						{1, framework.Numeric(`154193.3546`), framework.Numeric(`134097.6049`), 0.0},
						{2, framework.Numeric(`154193.3546`), framework.Numeric(`134097.6049`), 0.0},
						{3, framework.Numeric(`154191.0464`), framework.Numeric(`134098.6477`), 0.0},
					},
				},
				{
					Query:    `SELECT round(ST_DistanceSpheroid('POINT(-118.58 38.38)', 'POINT(-117.9 37.1)')::numeric, 4), ST_DistanceSpheroid('POINT EMPTY', 'POINT(-117.9 37.1)'), round(ST_DistanceSpheroid('MULTIPOINT(0 0,5 5)', 'POINT Z (1 1 9)')::numeric, 4);`,
					Expected: []sql.Row{{framework.Numeric(`154193.3546`), -1.0, framework.Numeric(`156899.5683`)}},
				},
				{
					Query:           `SELECT ST_DistanceSpheroid('SRID=4326;POINT(1 1)', 'POINT(2 2)');`,
					ExpectedErr:     `geometry_distance_spheroid: Operation on mixed SRID geometries (Point, 4326) != (Point, 0)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_DistanceSpheroid('GEOMETRYCOLLECTION(POINT(0 0))', 'POINT(1 1)');`,
					ExpectedErr:     `geometry_distance_spheroid: Only point/line/polygon supported.`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT s FROM geo_spheroid ORDER BY s;`,
					ExpectedErr:     `could not identify an ordering operator for type spheroid`,
					ExpectedErrCode: `42883`,
					Skip:            true,
				},
			},
		},
		{
			Name: `postgis fixed-length type updates`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
				`CREATE TABLE geo_fixed (id INT PRIMARY KEY, b box2d, c box3d, s spheroid);`,
				`INSERT INTO geo_fixed VALUES (1, 'BOX(0 0,1 1)', 'BOX3D(0 0 0,1 1 1)', 'SPHEROID["WGS 84",6378137,298.257223563]');`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `UPDATE geo_fixed SET b = 'BOX(0 0,2 2)', c = 'BOX3D(0 0 0,2 2 2)', s = 'SPHEROID["b",1,2]' WHERE id = 1 RETURNING id, b, c, s;`,
					Expected: []sql.Row{{1, `BOX(0 0,2 2)`, `BOX3D(0 0 0,2 2 2)`, `SPHEROID("b",1,2)`}},
				},
			},
		},
		{
			Name: `postgis gml output`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsGML('SRID=4326;POINT(1.123456789 2)'::geometry), ST_AsGML(3, 'SRID=4326;POINT(1.123456789 2)'::geometry, 15, 0, 'x', 'myid'), ST_AsGML(3, 'SRID=4326;POINT(1.123456789 2)'::geometry, 15, 6, '', 'i');`,
					Expected: []sql.Row{{`<gml:Point srsName="EPSG:4326"><gml:coordinates>1.123456789,2</gml:coordinates></gml:Point>`, `<x:Point srsName="EPSG:4326" x:id="myid"><x:pos srsDimension="2">1.123456789 2</x:pos></x:Point>`, `<Point srsName="EPSG:4326" id="i"><pos>1.123456789 2</pos></Point>`}},
				},
				{
					Query:    `SELECT ST_AsGML('LINESTRING(0 0,1 1)'::geometry), ST_AsGML(3, 'LINESTRING(0 0,1 1)'::geometry, 15, 0, 'x', 'myid'), ST_AsGML(3, 'LINESTRING(0 0,1 1)'::geometry, 15, 6, '', 'i');`,
					Expected: []sql.Row{{`<gml:LineString><gml:coordinates>0,0 1,1</gml:coordinates></gml:LineString>`, `<x:Curve x:id="myid"><x:segments><x:LineStringSegment><x:posList srsDimension="2">0 0 1 1</x:posList></x:LineStringSegment></x:segments></x:Curve>`, `<LineString id="i"><posList>0 0 1 1</posList></LineString>`}},
				},
				{
					Query:    `SELECT ST_AsGML('POLYGON((0 0,1 0,1 1,0 0),(0.1 0.1,0.2 0.1,0.2 0.2,0.1 0.1))'::geometry), ST_AsGML(3, 'POLYGON((0 0,1 0,1 1,0 0),(0.1 0.1,0.2 0.1,0.2 0.2,0.1 0.1))'::geometry, 15, 0, 'x', 'myid'), ST_AsGML(3, 'POLYGON((0 0,1 0,1 1,0 0),(0.1 0.1,0.2 0.1,0.2 0.2,0.1 0.1))'::geometry, 15, 6, '', 'i');`,
					Expected: []sql.Row{{`<gml:Polygon><gml:outerBoundaryIs><gml:LinearRing><gml:coordinates>0,0 1,0 1,1 0,0</gml:coordinates></gml:LinearRing></gml:outerBoundaryIs><gml:innerBoundaryIs><gml:LinearRing><gml:coordinates>0.1,0.1 0.2,0.1 0.2,0.2 0.1,0.1</gml:coordinates></gml:LinearRing></gml:innerBoundaryIs></gml:Polygon>`, `<x:Polygon x:id="myid"><x:exterior><x:LinearRing><x:posList srsDimension="2">0 0 1 0 1 1 0 0</x:posList></x:LinearRing></x:exterior><x:interior><x:LinearRing><x:posList srsDimension="2">0.1 0.1 0.2 0.1 0.2 0.2 0.1 0.1</x:posList></x:LinearRing></x:interior></x:Polygon>`, `<Polygon id="i"><exterior><LinearRing><posList>0 0 1 0 1 1 0 0</posList></LinearRing></exterior><interior><LinearRing><posList>0.1 0.1 0.2 0.1 0.2 0.2 0.1 0.1</posList></LinearRing></interior></Polygon>`}},
				},
				{
					Query:    `SELECT ST_AsGML('MULTIPOINT(1 2,EMPTY)'::geometry), ST_AsGML(3, 'MULTIPOINT(1 2,EMPTY)'::geometry, 15, 0, 'x', 'myid'), ST_AsGML(3, 'MULTIPOINT(1 2,EMPTY)'::geometry, 15, 6, '', 'i');`,
					Expected: []sql.Row{{`<gml:MultiPoint><gml:pointMember><gml:Point><gml:coordinates>1,2</gml:coordinates></gml:Point></gml:pointMember><gml:pointMember><gml:Point/></gml:pointMember></gml:MultiPoint>`, `<x:MultiPoint x:id="myid"><x:pointMember><x:Point x:id="myid"><x:pos srsDimension="2">1 2</x:pos></x:Point></x:pointMember><x:pointMember><x:Point x:id="myid"/></x:pointMember></x:MultiPoint>`, `<MultiPoint id="i"><pointMember><Point id="i"><pos>1 2</pos></Point></pointMember><pointMember><Point id="i"/></pointMember></MultiPoint>`}},
				},
				{
					Query:    `SELECT ST_AsGML('MULTILINESTRING((0 0,1 1),(2 2,3 3))'::geometry), ST_AsGML(3, 'MULTILINESTRING((0 0,1 1),(2 2,3 3))'::geometry, 15, 0, 'x', 'myid'), ST_AsGML(3, 'MULTILINESTRING((0 0,1 1),(2 2,3 3))'::geometry, 15, 6, '', 'i');`,
					Expected: []sql.Row{{`<gml:MultiLineString><gml:lineStringMember><gml:LineString><gml:coordinates>0,0 1,1</gml:coordinates></gml:LineString></gml:lineStringMember><gml:lineStringMember><gml:LineString><gml:coordinates>2,2 3,3</gml:coordinates></gml:LineString></gml:lineStringMember></gml:MultiLineString>`, `<x:MultiCurve x:id="myid"><x:curveMember><x:Curve x:id="myid"><x:segments><x:LineStringSegment><x:posList srsDimension="2">0 0 1 1</x:posList></x:LineStringSegment></x:segments></x:Curve></x:curveMember><x:curveMember><x:Curve x:id="myid"><x:segments><x:LineStringSegment><x:posList srsDimension="2">2 2 3 3</x:posList></x:LineStringSegment></x:segments></x:Curve></x:curveMember></x:MultiCurve>`, `<MultiCurve id="i"><curveMember><LineString id="i"><posList>0 0 1 1</posList></LineString></curveMember><curveMember><LineString id="i"><posList>2 2 3 3</posList></LineString></curveMember></MultiCurve>`}},
				},
				{
					Query:    `SELECT ST_AsGML('MULTIPOLYGON(((0 0,1 0,1 1,0 0)),EMPTY)'::geometry), ST_AsGML(3, 'MULTIPOLYGON(((0 0,1 0,1 1,0 0)),EMPTY)'::geometry, 15, 0, 'x', 'myid'), ST_AsGML(3, 'MULTIPOLYGON(((0 0,1 0,1 1,0 0)),EMPTY)'::geometry, 15, 6, '', 'i');`,
					Expected: []sql.Row{{`<gml:MultiPolygon><gml:polygonMember><gml:Polygon><gml:outerBoundaryIs><gml:LinearRing><gml:coordinates>0,0 1,0 1,1 0,0</gml:coordinates></gml:LinearRing></gml:outerBoundaryIs></gml:Polygon></gml:polygonMember><gml:polygonMember><gml:Polygon/></gml:polygonMember></gml:MultiPolygon>`, `<x:MultiSurface x:id="myid"><x:surfaceMember><x:Polygon x:id="myid"><x:exterior><x:LinearRing><x:posList srsDimension="2">0 0 1 0 1 1 0 0</x:posList></x:LinearRing></x:exterior></x:Polygon></x:surfaceMember><x:surfaceMember><x:Polygon x:id="myid"/></x:surfaceMember></x:MultiSurface>`, `<MultiSurface id="i"><surfaceMember><Polygon id="i"><exterior><LinearRing><posList>0 0 1 0 1 1 0 0</posList></LinearRing></exterior></Polygon></surfaceMember><surfaceMember><Polygon id="i"/></surfaceMember></MultiSurface>`}},
				},
				{
					Query:    `SELECT ST_AsGML('GEOMETRYCOLLECTION(LINESTRING EMPTY,POLYGON EMPTY,MULTIPOINT EMPTY,POINT(1 1))'::geometry), ST_AsGML(3, 'GEOMETRYCOLLECTION(LINESTRING EMPTY,POLYGON EMPTY,MULTIPOINT EMPTY,POINT(1 1))'::geometry, 15, 0, 'x', 'myid'), ST_AsGML(3, 'GEOMETRYCOLLECTION(LINESTRING EMPTY,POLYGON EMPTY,MULTIPOINT EMPTY,POINT(1 1))'::geometry, 15, 6, '', 'i');`,
					Expected: []sql.Row{{`<gml:MultiGeometry><gml:geometryMember><gml:LineString/></gml:geometryMember><gml:geometryMember><gml:Polygon/></gml:geometryMember><gml:geometryMember><gml:MultiPoint/></gml:geometryMember><gml:geometryMember><gml:Point><gml:coordinates>1,1</gml:coordinates></gml:Point></gml:geometryMember></gml:MultiGeometry>`, `<x:MultiGeometry x:id="myid"><x:geometryMember><x:Curve x:id="myid"/></x:geometryMember><x:geometryMember><x:Polygon x:id="myid"/></x:geometryMember><x:geometryMember><x:MultiPoint x:id="myid"/></x:geometryMember><x:geometryMember><x:Point x:id="myid"><x:pos srsDimension="2">1 1</x:pos></x:Point></x:geometryMember></x:MultiGeometry>`, `<MultiGeometry id="i"><geometryMember><LineString id="i"/></geometryMember><geometryMember><Polygon id="i"/></geometryMember><geometryMember><MultiPoint id="i"/></geometryMember><geometryMember><Point id="i"><pos>1 1</pos></Point></geometryMember></MultiGeometry>`}},
				},
				{
					Query:    `SELECT ST_AsGML('SRID=3857;LINESTRING Z (0 0 0,1 1 1)'::geometry), ST_AsGML(3, 'SRID=3857;LINESTRING Z (0 0 0,1 1 1)'::geometry, 15, 0, 'x', 'myid'), ST_AsGML(3, 'SRID=3857;LINESTRING Z (0 0 0,1 1 1)'::geometry, 15, 6, '', 'i');`,
					Expected: []sql.Row{{`<gml:LineString srsName="EPSG:3857"><gml:coordinates>0,0,0 1,1,1</gml:coordinates></gml:LineString>`, `<x:Curve srsName="EPSG:3857" x:id="myid"><x:segments><x:LineStringSegment><x:posList srsDimension="3">0 0 0 1 1 1</x:posList></x:LineStringSegment></x:segments></x:Curve>`, `<LineString srsName="EPSG:3857" id="i"><posList>0 0 0 1 1 1</posList></LineString>`}},
				},
				{
					Query:    `SELECT ST_AsGML('LINESTRING ZM (0 0 1 2,1 1 3 4)'::geometry), ST_AsGML(3, 'LINESTRING ZM (0 0 1 2,1 1 3 4)'::geometry, 15, 0, 'x', 'myid'), ST_AsGML(3, 'LINESTRING ZM (0 0 1 2,1 1 3 4)'::geometry, 15, 6, '', 'i');`,
					Expected: []sql.Row{{`<gml:LineString><gml:coordinates>0,0,1 1,1,3</gml:coordinates></gml:LineString>`, `<x:Curve x:id="myid"><x:segments><x:LineStringSegment><x:posList srsDimension="3">0 0 1 1 1 3</x:posList></x:LineStringSegment></x:segments></x:Curve>`, `<LineString id="i"><posList>0 0 1 1 1 3</posList></LineString>`}},
				},
				{
					Query:    `SELECT ST_AsGML(3, 'TRIANGLE((0 0,1 0,1 1,0 0))'::geometry, 15, 0, 'gml', 'i'), ST_AsGML(3, 'TRIANGLE((0 0,1 0,1 1,0 0))'::geometry, 15, 4);`,
					Expected: []sql.Row{{`<gml:Triangle gml:id="i"><gml:exterior><gml:LinearRing><gml:posList srsDimension="2">0 0 1 0 1 1 0 0</gml:posList></gml:LinearRing></gml:exterior></gml:Triangle>`, `<gml:Triangle><gml:exterior><gml:LinearRing><gml:posList srsDimension="2">0 0 1 0 1 1 0 0</gml:posList></gml:LinearRing></gml:exterior></gml:Triangle>`}},
				},
				{
					Query:           `SELECT ST_AsGML(2, 'TRIANGLE((0 0,1 0,1 1,0 0))'::geometry);`,
					ExpectedErr:     `Cannot convert Triangle to GML2. Try ST_AsGML(3, <geometry>) to generate GML3.`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsGML(3, 'TIN(((0 0,1 0,1 1,0 0)))'::geometry, 15, 0, 'gml', 'i'), ST_AsGML(3, 'TIN(((0 0,1 0,1 1,0 0)))'::geometry, 15, 4);`,
					Expected: []sql.Row{{`<gml:Tin gml:id="i"><gml:trianglePatches><gml:Triangle gml:id="i"><gml:exterior><gml:LinearRing><gml:posList srsDimension="2">0 0 1 0 1 1 0 0</gml:posList></gml:LinearRing></gml:exterior></gml:Triangle></gml:trianglePatches></gml:Tin>`, `<gml:Tin><gml:trianglePatches><gml:Triangle><gml:exterior><gml:LinearRing><gml:posList srsDimension="2">0 0 1 0 1 1 0 0</gml:posList></gml:LinearRing></gml:exterior></gml:Triangle></gml:trianglePatches></gml:Tin>`}},
				},
				{
					Query:           `SELECT ST_AsGML(2, 'TIN(((0 0,1 0,1 1,0 0)))'::geometry);`,
					ExpectedErr:     `Cannot convert Tin to GML2. Try ST_AsGML(3, <geometry>) to generate GML3.`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsGML(3, 'POLYHEDRALSURFACE(((0 0,1 0,1 1,0 0)))'::geometry, 15, 0, 'gml', 'i'), ST_AsGML(3, 'POLYHEDRALSURFACE(((0 0,1 0,1 1,0 0)))'::geometry, 15, 4);`,
					Expected: []sql.Row{{`<gml:PolyhedralSurface gml:id="i"><gml:polygonPatches><gml:PolygonPatch gml:id="i"><gml:exterior><gml:LinearRing><gml:posList srsDimension="2">0 0 1 0 1 1 0 0</gml:posList></gml:LinearRing></gml:exterior></gml:PolygonPatch></gml:polygonPatches></gml:PolyhedralSurface>`, `<gml:PolyhedralSurface><gml:polygonPatches><gml:PolygonPatch><gml:exterior><gml:LinearRing><gml:posList srsDimension="2">0 0 1 0 1 1 0 0</gml:posList></gml:LinearRing></gml:exterior></gml:PolygonPatch></gml:polygonPatches></gml:PolyhedralSurface>`}},
				},
				{
					Query:           `SELECT ST_AsGML(2, 'POLYHEDRALSURFACE(((0 0,1 0,1 1,0 0)))'::geometry);`,
					ExpectedErr:     `Cannot convert PolyhedralSurface to GML2. Try ST_AsGML(3, <geometry>) to generate GML3.`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsGML(3, 'COMPOUNDCURVE((0 0,1 0),CIRCULARSTRING(1 0,2 1,3 0))'::geometry, 15, 0, 'gml', 'i'), ST_AsGML(3, 'COMPOUNDCURVE((0 0,1 0),CIRCULARSTRING(1 0,2 1,3 0))'::geometry, 15, 4);`,
					Expected: []sql.Row{{`<gml:Curve gml:id="i"><gml:segments><gml:LineStringSegment><gml:posList srsDimension="2">0 0 1 0</gml:posList></gml:LineStringSegment><gml:ArcString><gml:posList srsDimension="2">1 0 2 1 3 0</gml:posList></gml:ArcString></gml:segments></gml:Curve>`, `<gml:Curve><gml:segments><gml:LineStringSegment><gml:posList srsDimension="2">0 0 1 0</gml:posList></gml:LineStringSegment><gml:ArcString><gml:posList srsDimension="2">1 0 2 1 3 0</gml:posList></gml:ArcString></gml:segments></gml:Curve>`}},
				},
				{
					Query:           `SELECT ST_AsGML(2, 'COMPOUNDCURVE((0 0,1 0),CIRCULARSTRING(1 0,2 1,3 0))'::geometry);`,
					ExpectedErr:     `lwgeom_to_gml2: 'CompoundCurve' geometry type not supported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsGML(3, 'CURVEPOLYGON(CIRCULARSTRING(0 0,1 1,2 0,1 -1,0 0),(0.5 0,0.6 0,0.6 0.1,0.5 0))'::geometry, 15, 0, 'gml', 'i'), ST_AsGML(3, 'CURVEPOLYGON(CIRCULARSTRING(0 0,1 1,2 0,1 -1,0 0),(0.5 0,0.6 0,0.6 0.1,0.5 0))'::geometry, 15, 4);`,
					Expected: []sql.Row{{`<gml:Polygon gml:id="i"><gml:exterior><gml:Ring><gml:curveMember><gml:Curve gml:id="i"><gml:segments><gml:ArcString><gml:posList srsDimension="2">0 0 1 1 2 0 1 -1 0 0</gml:posList></gml:ArcString></gml:segments></gml:Curve></gml:curveMember></gml:Ring></gml:exterior><gml:interior><gml:LinearRing><gml:posList srsDimension="2">0.5 0 0.6 0 0.6 0.1 0.5 0</gml:posList></gml:LinearRing></gml:interior></gml:Polygon>`, `<gml:Polygon><gml:exterior><gml:Ring><gml:curveMember><gml:Curve><gml:segments><gml:ArcString><gml:posList srsDimension="2">0 0 1 1 2 0 1 -1 0 0</gml:posList></gml:ArcString></gml:segments></gml:Curve></gml:curveMember></gml:Ring></gml:exterior><gml:interior><gml:LinearRing><gml:posList srsDimension="2">0.5 0 0.6 0 0.6 0.1 0.5 0</gml:posList></gml:LinearRing></gml:interior></gml:Polygon>`}},
				},
				{
					Query:           `SELECT ST_AsGML(2, 'CURVEPOLYGON(CIRCULARSTRING(0 0,1 1,2 0,1 -1,0 0),(0.5 0,0.6 0,0.6 0.1,0.5 0))'::geometry);`,
					ExpectedErr:     `lwgeom_to_gml2: 'CurvePolygon' geometry type not supported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsGML(3, 'MULTICURVE((0 0,1 1),CIRCULARSTRING(0 0,1 1,2 0))'::geometry, 15, 0, 'gml', 'i'), ST_AsGML(3, 'MULTICURVE((0 0,1 1),CIRCULARSTRING(0 0,1 1,2 0))'::geometry, 15, 4);`,
					Expected: []sql.Row{{`<gml:MultiCurve gml:id="i"><gml:curveMember><gml:Curve gml:id="i"><gml:segments><gml:LineStringSegment><gml:posList srsDimension="2">0 0 1 1</gml:posList></gml:LineStringSegment></gml:segments></gml:Curve></gml:curveMember><gml:curveMember><gml:Curve gml:id="i"><gml:segments><gml:ArcString><gml:posList srsDimension="2">0 0 1 1 2 0</gml:posList></gml:ArcString></gml:segments></gml:Curve></gml:curveMember></gml:MultiCurve>`, `<gml:MultiCurve><gml:curveMember><gml:LineString><gml:posList srsDimension="2">0 0 1 1</gml:posList></gml:LineString></gml:curveMember><gml:curveMember><gml:Curve><gml:segments><gml:ArcString><gml:posList srsDimension="2">0 0 1 1 2 0</gml:posList></gml:ArcString></gml:segments></gml:Curve></gml:curveMember></gml:MultiCurve>`}},
				},
				{
					Query:           `SELECT ST_AsGML(2, 'MULTICURVE((0 0,1 1),CIRCULARSTRING(0 0,1 1,2 0))'::geometry);`,
					ExpectedErr:     `lwgeom_to_gml2: 'MultiCurve' geometry type not supported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsGML(3, 'MULTISURFACE(((0 0,1 0,1 1,0 0)),CURVEPOLYGON(CIRCULARSTRING(0 0,1 1,2 0,1 -1,0 0)))'::geometry, 15, 0, 'gml', 'i'), ST_AsGML(3, 'MULTISURFACE(((0 0,1 0,1 1,0 0)),CURVEPOLYGON(CIRCULARSTRING(0 0,1 1,2 0,1 -1,0 0)))'::geometry, 15, 4);`,
					Expected: []sql.Row{{`<gml:MultiSurface gml:id="i"><gml:Polygon gml:id="i"><gml:exterior><gml:LinearRing><gml:posList srsDimension="2">0 0 1 0 1 1 0 0</gml:posList></gml:LinearRing></gml:exterior></gml:Polygon><gml:Polygon gml:id="i"><gml:exterior><gml:Ring><gml:curveMember><gml:Curve gml:id="i"><gml:segments><gml:ArcString><gml:posList srsDimension="2">0 0 1 1 2 0 1 -1 0 0</gml:posList></gml:ArcString></gml:segments></gml:Curve></gml:curveMember></gml:Ring></gml:exterior></gml:Polygon></gml:MultiSurface>`, `<gml:MultiSurface><gml:Polygon><gml:exterior><gml:LinearRing><gml:posList srsDimension="2">0 0 1 0 1 1 0 0</gml:posList></gml:LinearRing></gml:exterior></gml:Polygon><gml:Polygon><gml:exterior><gml:Ring><gml:curveMember><gml:Curve><gml:segments><gml:ArcString><gml:posList srsDimension="2">0 0 1 1 2 0 1 -1 0 0</gml:posList></gml:ArcString></gml:segments></gml:Curve></gml:curveMember></gml:Ring></gml:exterior></gml:Polygon></gml:MultiSurface>`}},
				},
				{
					Query:           `SELECT ST_AsGML(2, 'MULTISURFACE(((0 0,1 0,1 1,0 0)),CURVEPOLYGON(CIRCULARSTRING(0 0,1 1,2 0,1 -1,0 0)))'::geometry);`,
					ExpectedErr:     `lwgeom_to_gml2: 'MultiSurface' geometry type not supported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsGML(2, 'GEOMETRYCOLLECTION(TRIANGLE((0 0,1 0,1 1,0 0)))'::geometry), ST_AsGML(2, 'POINT EMPTY'::geometry);`,
					Expected: []sql.Row{{`<gml:MultiGeometry><gml:geometryMember></gml:geometryMember></gml:MultiGeometry>`, nil}},
				},
				{
					Query:           `SELECT ST_AsGML(3, 'GEOMETRYCOLLECTION(MULTICURVE((0 0,1 1)))'::geometry);`,
					ExpectedErr:     `asgml3_collection: unknown geometry type`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsGML(3, 'SRID=4326;LINESTRING(1 2,3 4)'::geometry, 15, 1), ST_AsGML(3, 'SRID=4326;LINESTRING(1 2,3 4)'::geometry, 15, 17), ST_AsGML(2, 'SRID=4326;LINESTRING(1 2,3 4)'::geometry, 15, 17);`,
					Expected: []sql.Row{{`<gml:Curve srsName="urn:ogc:def:crs:EPSG::4326"><gml:segments><gml:LineStringSegment><gml:posList srsDimension="2">1 2 3 4</gml:posList></gml:LineStringSegment></gml:segments></gml:Curve>`, `<gml:Curve srsName="urn:ogc:def:crs:EPSG::4326"><gml:segments><gml:LineStringSegment><gml:posList srsDimension="2">2 1 4 3</gml:posList></gml:LineStringSegment></gml:segments></gml:Curve>`, `<gml:LineString srsName="urn:ogc:def:crs:EPSG::4326"><gml:coordinates>1,2 3,4</gml:coordinates></gml:LineString>`}},
				},
				{
					Query:    `SELECT ST_AsGML(3, 'SRID=4326;POLYGON Z ((0 0 1,1 0 2,1 1 3,0 0 1))'::geometry, 15, 32), ST_AsGML(3, 'SRID=4326;POLYGON Z ((0 0 1,1 0 2,1 1 3,0 0 1))'::geometry, 15, 50), ST_AsGML(2, 'SRID=4326;LINESTRING(1 2,3 4)'::geometry, 15, 32);`,
					Expected: []sql.Row{{`<gml:Envelope srsName="EPSG:4326" srsDimension="3"><gml:lowerCorner>0 0 1</gml:lowerCorner><gml:upperCorner>1 1 3</gml:upperCorner></gml:Envelope>`, `<gml:Envelope srsName="EPSG:4326"><gml:lowerCorner>0 0 1</gml:lowerCorner><gml:upperCorner>1 1 3</gml:upperCorner></gml:Envelope>`, `<gml:Box srsName="EPSG:4326"><gml:coordinates>1,2 3,4</gml:coordinates></gml:Box>`}},
				},
				{
					Query:    `SELECT ST_AsGML(3, 'SRID=4326;LINESTRING EMPTY'::geometry, 15, 32), ST_AsGML(2, 'SRID=4326;LINESTRING EMPTY'::geometry, 15, 32), ST_AsGML(3, 'LINESTRING(0 0,1 1)'::geometry, 15, 32, 'gml', 'i');`,
					Expected: []sql.Row{{`<gml:Envelope srsName="EPSG:4326"/>`, `<gml:Box srsName="EPSG:4326"/>`, `<gml:Envelope srsDimension="2"><gml:lowerCorner>0 0</gml:lowerCorner><gml:upperCorner>1 1</gml:upperCorner></gml:Envelope>`}},
				},
				{
					Query:    `SELECT ST_AsGML(3, 'POINT(1.23456789 2.5)'::geometry, 3), ST_AsGML(3, 'POINT(1.23456789 2.5)'::geometry, -1), ST_AsGML(3, 'POINT(123456789012345678 2.5e-9)'::geometry), ST_AsGML(3, 'POINT(1.5 -0.0000001)'::geometry, 3);`,
					Expected: []sql.Row{{`<gml:Point><gml:pos srsDimension="2">1.235 2.5</gml:pos></gml:Point>`, `<gml:Point><gml:pos srsDimension="2">1 2</gml:pos></gml:Point>`, `<gml:Point><gml:pos srsDimension="2">1.234567890123457e+17 2.5e-9</gml:pos></gml:Point>`, `<gml:Point><gml:pos srsDimension="2">1.5 0</gml:pos></gml:Point>`}},
				},
				{
					Query:    `SELECT ST_AsGML(3, 'POINT(1.123456789012345678 2)'::geometry, NULL, 0), ST_AsGML(3, 'SRID=4326;POINT(1 2)'::geometry, 15, NULL), ST_AsGML(3, 'POINT(1 2)'::geometry, 15, 0, NULL, 'i'), ST_AsGML(3, NULL::geometry) IS NULL;`,
					Expected: []sql.Row{{`<gml:Point><gml:pos srsDimension="2">1.123456789012346 2</gml:pos></gml:Point>`, `<gml:Point srsName="EPSG:4326"><gml:pos srsDimension="2">1 2</gml:pos></gml:Point>`, `<gml:Point gml:id="i"><gml:pos srsDimension="2">1 2</gml:pos></gml:Point>`, `t`}},
				},
				{
					Query:           `SELECT ST_AsGML(4, 'POINT(1 2)'::geometry);`,
					ExpectedErr:     `Only GML 2 and GML 3 are supported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsGML('SRID=4326;POINT(1 2)'::geography), ST_AsGML(3, 'SRID=4326;POINT(1 2)'::geography, 15, 1), ST_AsGML(2, 'SRID=4326;POINT(1 2)'::geography, 15, 1);`,
					Expected: []sql.Row{{`<gml:Point srsName="EPSG:4326"><gml:coordinates>1,2</gml:coordinates></gml:Point>`, `<gml:Point srsName="urn:ogc:def:crs:EPSG::4326"><gml:pos srsDimension="2">2 1</gml:pos></gml:Point>`, `<gml:Point srsName="urn:ogc:def:crs:EPSG::4326"><gml:coordinates>1,2</gml:coordinates></gml:Point>`}},
				},
				{
					Query:           `SELECT ST_AsGML(3, 'SRID=4326;POINT(1 2)'::geography, 15, 17);`,
					ExpectedErr:     `Options 17 passed to ST_AsGML(geography) but are only applicable to ST_AsGML(geometry)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsGML('POINT(1 2)'), _ST_AsGML(3, 'POINT(1 2)'::geometry, 15, 0, 'gml', 'z');`,
					Expected: []sql.Row{{`<gml:Point><gml:coordinates>1,2</gml:coordinates></gml:Point>`, `<gml:Point gml:id="z"><gml:pos srsDimension="2">1 2</gml:pos></gml:Point>`}},
				},
				{
					Query:    `SELECT ST_AsText('POINT(1.5 -0.0000001)'::geometry, 3), ST_AsGeoJSON('POINT(1.5 -0.0000001)'::geometry, 3), ST_AsEWKT('POINT(-0.0000001 -0.4)'::geometry, 0);`,
					Expected: []sql.Row{{`POINT(1.5 0)`, `{"type":"Point","coordinates":[1.5,0]}`, `POINT(0 0)`}},
				},
			},
		},
		{
			Name: `postgis x3d output`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsX3D('POINT Z (1 2 3)'::geometry), _ST_AsX3D(3, 'POINT Z (1 2 3)'::geometry, 15, 0, 'myid'), ST_AsX3D('SRID=4326;POINT Z (1 2 3)'::geometry, 15, 1), ST_AsX3D('SRID=4326;POINT Z (1 2 3)'::geometry, 15, 3);`,
					Expected: []sql.Row{{`1 2 3`, `1 2 3`, `2 1 3`, `2 1 3`}},
				},
				{
					Query:    `SELECT ST_AsX3D('LINESTRING(0 0,1 1,2 0)'::geometry), _ST_AsX3D(3, 'LINESTRING(0 0,1 1,2 0)'::geometry, 15, 0, 'myid'), ST_AsX3D('SRID=4326;LINESTRING(0 0,1 1,2 0)'::geometry, 15, 1), ST_AsX3D('SRID=4326;LINESTRING(0 0,1 1,2 0)'::geometry, 15, 3);`,
					Expected: []sql.Row{{`<LineSet  vertexCount='3'><Coordinate point='0 0 1 1 2 0' /></LineSet>`, `<LineSet myid: vertexCount='3'><Coordinate point='0 0 1 1 2 0' /></LineSet>`, `<LineSet  vertexCount='3'><Coordinate point='0 0 1 1 0 2' /></LineSet>`, `<LineSet  vertexCount='3'><GeoCoordinate geoSystem='"GD" "WE" "latitude_first"' point='0 0 1 1 0 2' /></LineSet>`}},
				},
				{
					Query:    `SELECT ST_AsX3D('POLYGON((0 0,1 0,1 1,0 0))'::geometry), _ST_AsX3D(3, 'POLYGON((0 0,1 0,1 1,0 0))'::geometry, 15, 0, 'myid'), ST_AsX3D('SRID=4326;POLYGON((0 0,1 0,1 1,0 0))'::geometry, 15, 1), ST_AsX3D('SRID=4326;POLYGON((0 0,1 0,1 1,0 0))'::geometry, 15, 3);`,
					Expected: []sql.Row{{`<IndexedFaceSet  convex='false' coordIndex='0 1 2'>0 0 1 0 1 1 ' />`, `<IndexedFaceSet myid: convex='false' coordIndex='0 1 2'>0 0 1 0 1 1 ' />`, `<IndexedFaceSet  convex='false' coordIndex='0 1 2'>0 0 0 1 1 1 ' />`, `<IndexedFaceSet  convex='false' coordIndex='0 1 2'>0 0 0 1 1 1 ' />`}},
				},
				{
					Query:    `SELECT ST_AsX3D('POLYGON Z ((0 0 0,1 0 0,1 1 0,0 0 0),(0.1 0.1 0,0.2 0.1 0,0.2 0.2 0,0.1 0.1 0))'::geometry), _ST_AsX3D(3, 'POLYGON Z ((0 0 0,1 0 0,1 1 0,0 0 0),(0.1 0.1 0,0.2 0.1 0,0.2 0.2 0,0.1 0.1 0))'::geometry, 15, 0, 'myid'), ST_AsX3D('SRID=4326;POLYGON Z ((0 0 0,1 0 0,1 1 0,0 0 0),(0.1 0.1 0,0.2 0.1 0,0.2 0.2 0,0.1 0.1 0))'::geometry, 15, 1), ST_AsX3D('SRID=4326;POLYGON Z ((0 0 0,1 0 0,1 1 0,0 0 0),(0.1 0.1 0,0.2 0.1 0,0.2 0.2 0,0.1 0.1 0))'::geometry, 15, 3);`,
					Expected: []sql.Row{{`<IndexedFaceSet  convex='false' coordIndex='0 1 2 -1 3 4 5'><Coordinate point='0 0 0 1 0 0 1 1 0 0.1 0.1 0 0.2 0.1 0 0.2 0.2 0 ' /></IndexedFaceSet>`, `<IndexedFaceSet myid: convex='false' coordIndex='0 1 2 -1 3 4 5'><Coordinate point='0 0 0 1 0 0 1 1 0 0.1 0.1 0 0.2 0.1 0 0.2 0.2 0 ' /></IndexedFaceSet>`, `<IndexedFaceSet  convex='false' coordIndex='0 1 2 -1 3 4 5'><Coordinate point='0 0 0 0 1 0 1 1 0 0.1 0.1 0 0.1 0.2 0 0.2 0.2 0 ' /></IndexedFaceSet>`, `<IndexedFaceSet  convex='false' coordIndex='0 1 2 -1 3 4 5'><GeoCoordinate geoSystem='"GD" "WE" "latitude_first"' point='0 0 0 0 1 0 1 1 0 0.1 0.1 0 0.1 0.2 0 0.2 0.2 0 ' /></IndexedFaceSet>`}},
				},
				{
					Query:    `SELECT ST_AsX3D('MULTIPOINT(1 2,EMPTY)'::geometry), _ST_AsX3D(3, 'MULTIPOINT(1 2,EMPTY)'::geometry, 15, 0, 'myid'), ST_AsX3D('SRID=4326;MULTIPOINT(1 2,EMPTY)'::geometry, 15, 1), ST_AsX3D('SRID=4326;MULTIPOINT(1 2,EMPTY)'::geometry, 15, 3);`,
					Expected: []sql.Row{{`<Polypoint2D  point='1 2  ' />`, `<Polypoint2D myid: point='1 2  ' />`, `<Polypoint2D  point='2 1  ' />`, `<Polypoint2D  point='2 1  ' />`}},
				},
				{
					Query:    `SELECT ST_AsX3D('MULTIPOINT Z (1 2 3,4 5 6)'::geometry), _ST_AsX3D(3, 'MULTIPOINT Z (1 2 3,4 5 6)'::geometry, 15, 0, 'myid'), ST_AsX3D('SRID=4326;MULTIPOINT Z (1 2 3,4 5 6)'::geometry, 15, 1), ST_AsX3D('SRID=4326;MULTIPOINT Z (1 2 3,4 5 6)'::geometry, 15, 3);`,
					Expected: []sql.Row{{`<PointSet ><Coordinate point='1 2 3 4 5 6 ' /></PointSet>`, `<PointSet myid:><Coordinate point='1 2 3 4 5 6 ' /></PointSet>`, `<PointSet ><Coordinate point='2 1 3 5 4 6 ' /></PointSet>`, `<PointSet ><GeoCoordinate geoSystem='"GD" "WE" "latitude_first"' point='2 1 3 5 4 6 ' /></PointSet>`}},
				},
				{
					Query:    `SELECT ST_AsX3D('MULTILINESTRING((0 0,1 1),EMPTY)'::geometry), _ST_AsX3D(3, 'MULTILINESTRING((0 0,1 1),EMPTY)'::geometry, 15, 0, 'myid'), ST_AsX3D('SRID=4326;MULTILINESTRING((0 0,1 1),EMPTY)'::geometry, 15, 1), ST_AsX3D('SRID=4326;MULTILINESTRING((0 0,1 1),EMPTY)'::geometry, 15, 3);`,
					Expected: []sql.Row{{`<IndexedLineSet  coordIndex='0 1 -1 '>0 0 1 1  ' />`, `<IndexedLineSet myid: coordIndex='0 1 -1 '>0 0 1 1  ' />`, `<IndexedLineSet  coordIndex='0 1 -1 '>0 0 1 1  ' />`, `<IndexedLineSet  coordIndex='0 1 -1 '>0 0 1 1  ' />`}},
				},
				{
					Query:    `SELECT ST_AsX3D('MULTILINESTRING Z ((0 0 0,1 1 1),(2 2 2,3 3 3))'::geometry), _ST_AsX3D(3, 'MULTILINESTRING Z ((0 0 0,1 1 1),(2 2 2,3 3 3))'::geometry, 15, 0, 'myid'), ST_AsX3D('SRID=4326;MULTILINESTRING Z ((0 0 0,1 1 1),(2 2 2,3 3 3))'::geometry, 15, 1), ST_AsX3D('SRID=4326;MULTILINESTRING Z ((0 0 0,1 1 1),(2 2 2,3 3 3))'::geometry, 15, 3);`,
					Expected: []sql.Row{{`<IndexedLineSet  coordIndex='0 1 -1 2 3'><Coordinate point='0 0 0 1 1 1 2 2 2 3 3 3 ' /></IndexedLineSet>`, `<IndexedLineSet myid: coordIndex='0 1 -1 2 3'><Coordinate point='0 0 0 1 1 1 2 2 2 3 3 3 ' /></IndexedLineSet>`, `<IndexedLineSet  coordIndex='0 1 -1 2 3'><Coordinate point='0 0 0 1 1 1 2 2 2 3 3 3 ' /></IndexedLineSet>`, `<IndexedLineSet  coordIndex='0 1 -1 2 3'><GeoCoordinate geoSystem='"GD" "WE" "latitude_first"' point='0 0 0 1 1 1 2 2 2 3 3 3 ' /></IndexedLineSet>`}},
				},
				{
					Query:    `SELECT ST_AsX3D('MULTIPOLYGON(((0 0,1 0,1 1,0 0)),EMPTY)'::geometry), _ST_AsX3D(3, 'MULTIPOLYGON(((0 0,1 0,1 1,0 0)),EMPTY)'::geometry, 15, 0, 'myid'), ST_AsX3D('SRID=4326;MULTIPOLYGON(((0 0,1 0,1 1,0 0)),EMPTY)'::geometry, 15, 1), ST_AsX3D('SRID=4326;MULTIPOLYGON(((0 0,1 0,1 1,0 0)),EMPTY)'::geometry, 15, 3);`,
					Expected: []sql.Row{{`<IndexedFaceSet  convex='false' coordIndex='0 1 2 -1 '>0 0 1 0 1 1  ' />`, `<IndexedFaceSet myid: convex='false' coordIndex='0 1 2 -1 '>0 0 1 0 1 1  ' />`, `<IndexedFaceSet  convex='false' coordIndex='0 1 2 -1 '>0 0 0 1 1 1  ' />`, `<IndexedFaceSet  convex='false' coordIndex='0 1 2 -1 '>0 0 0 1 1 1  ' />`}},
				},
				{
					Query:    `SELECT ST_AsX3D('MULTIPOLYGON Z (((0 0 0,1 0 0,1 1 0,0 0 0),(0.1 0.1 0,0.2 0.1 0,0.2 0.2 0,0.1 0.1 0)),((2 2 0,3 2 0,3 3 0,2 2 0)))'::geometry), _ST_AsX3D(3, 'MULTIPOLYGON Z (((0 0 0,1 0 0,1 1 0,0 0 0),(0.1 0.1 0,0.2 0.1 0,0.2 0.2 0,0.1 0.1 0)),((2 2 0,3 2 0,3 3 0,2 2 0)))'::geometry, 15, 0, 'myid'), ST_AsX3D('SRID=4326;MULTIPOLYGON Z (((0 0 0,1 0 0,1 1 0,0 0 0),(0.1 0.1 0,0.2 0.1 0,0.2 0.2 0,0.1 0.1 0)),((2 2 0,3 2 0,3 3 0,2 2 0)))'::geometry, 15, 1), ST_AsX3D('SRID=4326;MULTIPOLYGON Z (((0 0 0,1 0 0,1 1 0,0 0 0),(0.1 0.1 0,0.2 0.1 0,0.2 0.2 0,0.1 0.1 0)),((2 2 0,3 2 0,3 3 0,2 2 0)))'::geometry, 15, 3);`,
					Expected: []sql.Row{{`<IndexedFaceSet  convex='false' coordIndex='0 1 2 -1 3 4 5 -1 6 7 8'><Coordinate point='0 0 0 1 0 0 1 1 0 0.1 0.1 0 0.2 0.1 0 0.2 0.2 0 2 2 0 3 2 0 3 3 0 ' /></IndexedFaceSet>`, `<IndexedFaceSet myid: convex='false' coordIndex='0 1 2 -1 3 4 5 -1 6 7 8'><Coordinate point='0 0 0 1 0 0 1 1 0 0.1 0.1 0 0.2 0.1 0 0.2 0.2 0 2 2 0 3 2 0 3 3 0 ' /></IndexedFaceSet>`, `<IndexedFaceSet  convex='false' coordIndex='0 1 2 -1 3 4 5 -1 6 7 8'><Coordinate point='0 0 0 0 1 0 1 1 0 0.1 0.1 0 0.1 0.2 0 0.2 0.2 0 2 2 0 2 3 0 3 3 0 ' /></IndexedFaceSet>`, `<IndexedFaceSet  convex='false' coordIndex='0 1 2 -1 3 4 5 -1 6 7 8'><GeoCoordinate geoSystem='"GD" "WE" "latitude_first"' point='0 0 0 0 1 0 1 1 0 0.1 0.1 0 0.1 0.2 0 0.2 0.2 0 2 2 0 2 3 0 3 3 0 ' /></IndexedFaceSet>`}},
				},
				{
					Query:    `SELECT ST_AsX3D('TRIANGLE Z ((0 0 0,1 0 0,1 1 0,0 0 0))'::geometry), _ST_AsX3D(3, 'TRIANGLE Z ((0 0 0,1 0 0,1 1 0,0 0 0))'::geometry, 15, 0, 'myid'), ST_AsX3D('SRID=4326;TRIANGLE Z ((0 0 0,1 0 0,1 1 0,0 0 0))'::geometry, 15, 1), ST_AsX3D('SRID=4326;TRIANGLE Z ((0 0 0,1 0 0,1 1 0,0 0 0))'::geometry, 15, 3);`,
					Expected: []sql.Row{{`0 0 0 1 0 0 1 1 0`, `0 0 0 1 0 0 1 1 0`, `0 0 0 0 1 0 1 1 0`, `0 0 0 0 1 0 1 1 0`}},
				},
				{
					Query:    `SELECT ST_AsX3D('TIN(((0 0,1 0,1 1,0 0)),((0 0,1 1,0 1,0 0)))'::geometry), _ST_AsX3D(3, 'TIN(((0 0,1 0,1 1,0 0)),((0 0,1 1,0 1,0 0)))'::geometry, 15, 0, 'myid'), ST_AsX3D('SRID=4326;TIN(((0 0,1 0,1 1,0 0)),((0 0,1 1,0 1,0 0)))'::geometry, 15, 1), ST_AsX3D('SRID=4326;TIN(((0 0,1 0,1 1,0 0)),((0 0,1 1,0 1,0 0)))'::geometry, 15, 3);`,
					Expected: []sql.Row{{`<IndexedTriangleSet  index='0 1 2 3 4 5'><Coordinate point='0 0 1 0 1 1 0 0 1 1 0 1'/></IndexedTriangleSet>`, `<IndexedTriangleSet myid: index='0 1 2 3 4 5'><Coordinate point='0 0 1 0 1 1 0 0 1 1 0 1'/></IndexedTriangleSet>`, `<IndexedTriangleSet  index='0 1 2 3 4 5'><Coordinate point='0 0 0 1 1 1 0 0 1 1 1 0'/></IndexedTriangleSet>`, `<IndexedTriangleSet  index='0 1 2 3 4 5'><GeoCoordinate geoSystem='"GD" "WE" "latitude_first"' point='0 0 0 1 1 1 0 0 1 1 1 0'/></IndexedTriangleSet>`}},
				},
				{
					Query:    `SELECT ST_AsX3D('POLYHEDRALSURFACE Z (((0 0 0,1 0 0,1 1 0,0 0 0),(0.1 0.1 0,0.2 0.1 0,0.2 0.2 0,0.1 0.1 0)),((5 5 5,6 5 5,6 6 5,5 5 5)))'::geometry), _ST_AsX3D(3, 'POLYHEDRALSURFACE Z (((0 0 0,1 0 0,1 1 0,0 0 0),(0.1 0.1 0,0.2 0.1 0,0.2 0.2 0,0.1 0.1 0)),((5 5 5,6 5 5,6 6 5,5 5 5)))'::geometry, 15, 0, 'myid'), ST_AsX3D('SRID=4326;POLYHEDRALSURFACE Z (((0 0 0,1 0 0,1 1 0,0 0 0),(0.1 0.1 0,0.2 0.1 0,0.2 0.2 0,0.1 0.1 0)),((5 5 5,6 5 5,6 6 5,5 5 5)))'::geometry, 15, 1), ST_AsX3D('SRID=4326;POLYHEDRALSURFACE Z (((0 0 0,1 0 0,1 1 0,0 0 0),(0.1 0.1 0,0.2 0.1 0,0.2 0.2 0,0.1 0.1 0)),((5 5 5,6 5 5,6 6 5,5 5 5)))'::geometry, 15, 3);`,
					Expected: []sql.Row{{`<IndexedFaceSet convex='false'  coordIndex='0 1 2 -1 3 4 5'><Coordinate point='0 0 0 1 0 0 1 1 0 0.1 0.1 0 0.2 0.1 0 0.2 0.2 0 5 5 5 6 5 5 6 6 5' /></IndexedFaceSet>`, `<IndexedFaceSet convex='false' myid: coordIndex='0 1 2 -1 3 4 5'><Coordinate point='0 0 0 1 0 0 1 1 0 0.1 0.1 0 0.2 0.1 0 0.2 0.2 0 5 5 5 6 5 5 6 6 5' /></IndexedFaceSet>`, `<IndexedFaceSet convex='false'  coordIndex='0 1 2 -1 3 4 5'><Coordinate point='0 0 0 0 1 0 1 1 0 0.1 0.1 0 0.1 0.2 0 0.2 0.2 0 5 5 5 5 6 5 6 6 5' /></IndexedFaceSet>`, `<IndexedFaceSet convex='false'  coordIndex='0 1 2 -1 3 4 5'><GeoCoordinate geoSystem='"GD" "WE" "latitude_first"' point='0 0 0 0 1 0 1 1 0 0.1 0.1 0 0.1 0.2 0 0.2 0.2 0 5 5 5 5 6 5 6 6 5' /></IndexedFaceSet>`}},
				},
				{
					Query:           `SELECT ST_AsX3D('GEOMETRYCOLLECTION(POINT EMPTY,POLYGON((0 0,1 0,1 1,0 1,0 0),(0.1 0.1,0.2 0.1,0.2 0.2,0.1 0.1)),MULTIPOINT Z (1 1 1),GEOMETRYCOLLECTION(POINT(1 1)))'::geometry), _ST_AsX3D(3, 'GEOMETRYCOLLECTION(POINT EMPTY,POLYGON((0 0,1 0,1 1,0 1,0 0),(0.1 0.1,0.2 0.1,0.2 0.2,0.1 0.1)),MULTIPOINT Z (1 1 1),GEOMETRYCOLLECTION(POINT(1 1)))'::geometry, 15, 0, 'myid'), ST_AsX3D('SRID=4326;GEOMETRYCOLLECTION(POINT EMPTY,POLYGON((0 0,1 0,1 1,0 1,0 0),(0.1 0.1,0.2 0.1,0.2 0.2,0.1 0.1)),MULTIPOINT Z (1 1 1),GEOMETRYCOLLECTION(POINT(1 1)))'::geometry, 15, 1), ST_AsX3D('SRID=4326;GEOMETRYCOLLECTION(POINT EMPTY,POLYGON((0 0,1 0,1 1,0 1,0 0),(0.1 0.1,0.2 0.1,0.2 0.2,0.1 0.1)),MULTIPOINT Z (1 1 1),GEOMETRYCOLLECTION(POINT(1 1)))'::geometry, 15, 3);`,
					ExpectedErr:     `Dimensions mismatch in lwcollection`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsX3D('POINT EMPTY'::geometry), _ST_AsX3D(3, 'POINT EMPTY'::geometry, 15, 0, 'myid'), ST_AsX3D('SRID=4326;POINT EMPTY'::geometry, 15, 1), ST_AsX3D('SRID=4326;POINT EMPTY'::geometry, 15, 3);`,
					Expected: []sql.Row{{``, ``, ``, ``}},
				},
				{
					Query:           `SELECT ST_AsX3D('CIRCULARSTRING(0 0,1 1,2 0)'::geometry);`,
					ExpectedErr:     `lwgeom_to_x3d3: 'CircularString' geometry type not supported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsX3D('GEOMETRYCOLLECTION(TRIANGLE((0 0,1 0,1 1,0 0)))'::geometry);`,
					ExpectedErr:     `asx3d3_collection_buf: unknown geometry type`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsX3D('POINT Z (1 2 3)'::geometry, 15, 2);`,
					ExpectedErr:     `Only SRID 4326 is supported for geocoordinates.`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsX3D('POINT(1.23456789 2)'::geometry, 3), ST_AsX3D('POINT(1.23456789 2)'::geometry, NULL), ST_AsX3D(NULL::geometry) IS NULL, ST_AsX3D('POINT(1 2)'::geometry, 15, NULL), ST_AsX3D('POINT(1e20 1)'::geometry);`,
					Expected: []sql.Row{{`1.235 2`, `1.23456789 2`, `t`, `1 2`, `1e+20 1`}},
				},
				{
					Query:           `SELECT _ST_AsX3D(2, 'POINT(1 2)'::geometry, 15, 0, '');`,
					ExpectedErr:     `Only X3D version 3 are supported`,
					ExpectedErrCode: `XX000`,
				},
			},
		},
		{
			Name: `postgis grids`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: `SELECT i, j, ST_AsEWKT(geom) FROM ST_SquareGrid(1, 'SRID=3857;LINESTRING(0.5 0.5,2.5 1.5)'::geometry);`,
					Expected: []sql.Row{
						{0, 0, `SRID=3857;POLYGON((0 0,0 1,1 1,1 0,0 0))`},
						{0, 1, `SRID=3857;POLYGON((0 1,0 2,1 2,1 1,0 1))`},
						{1, 0, `SRID=3857;POLYGON((1 0,1 1,2 1,2 0,1 0))`},
						{1, 1, `SRID=3857;POLYGON((1 1,1 2,2 2,2 1,1 1))`},
						{2, 0, `SRID=3857;POLYGON((2 0,2 1,3 1,3 0,2 0))`},
						{2, 1, `SRID=3857;POLYGON((2 1,2 2,3 2,3 1,2 1))`},
					},
				},
				{
					Query: `SELECT i, j, ST_AsText(geom) FROM ST_HexagonGrid(1, 'LINESTRING(0 0,2 2)'::geometry);`,
					Expected: []sql.Row{
						{0, 0, `POLYGON((-1 0,-0.5 -0.866025403784439,0.5 -0.866025403784439,1 0,0.5 0.866025403784439,-0.5 0.866025403784439,-1 0))`},
						{0, 1, `POLYGON((-1 1.732050807568877,-0.5 0.866025403784439,0.5 0.866025403784439,1 1.732050807568877,0.5 2.598076211353316,-0.5 2.598076211353316,-1 1.732050807568877))`},
						{1, 0, `POLYGON((0.5 0.866025403784439,1 0,2 0,2.5 0.866025403784439,2 1.732050807568877,1 1.732050807568877,0.5 0.866025403784439))`},
						{1, 1, `POLYGON((0.5 2.598076211353316,1 1.732050807568877,2 1.732050807568877,2.5 2.598076211353316,2 3.464101615137755,1 3.464101615137755,0.5 2.598076211353316))`},
						{2, 0, `POLYGON((2 0,2.5 -0.866025403784439,3.5 -0.866025403784439,4 0,3.5 0.866025403784439,2.5 0.866025403784439,2 0))`},
						{2, 1, `POLYGON((2 1.732050807568877,2.5 0.866025403784439,3.5 0.866025403784439,4 1.732050807568877,3.5 2.598076211353316,2.5 2.598076211353316,2 1.732050807568877))`},
					},
				},
				{
					Query:    `SELECT i, j, ST_AsText(geom) FROM ST_SquareGrid(0.3, 'POINT(-1.5 -0.8660254037844386)'::geometry);`,
					Expected: []sql.Row{{-5, -3, `POLYGON((-1.5 -0.9,-1.5 -0.6,-1.2 -0.6,-1.2 -0.9,-1.5 -0.9))`}},
				},
				{
					Query:    `SELECT i, j, ST_AsText(geom) FROM ST_HexagonGrid(0.3, 'POINT(-3 0)'::geometry);`,
					Expected: []sql.Row{{-7, 0, `POLYGON((-3.45 0.259807621135332,-3.3 0,-3 0,-2.85 0.259807621135332,-3 0.519615242270663,-3.3 0.519615242270663,-3.45 0.259807621135332))`}},
				},
				{
					Query: `SELECT i, j FROM ST_HexagonGrid(1, 'POINT(1 0)'::geometry);`,
					Expected: []sql.Row{
						{0, 0},
						{1, 0},
					},
				},
				{
					Query: `SELECT i, j FROM ST_HexagonGrid(1, 'POINT(0 0.8660254037844386)'::geometry);`,
					Expected: []sql.Row{
						{0, 0},
						{0, 1},
					},
				},
				{
					Query: `SELECT i, j FROM ST_HexagonGrid(2.5, 'SRID=3857;POLYGON((0.1 0.2,4.7 0.2,4.7 3.3,0.1 0.2))'::geometry);`,
					Expected: []sql.Row{
						{0, 0},
						{0, 1},
						{1, 0},
					},
				},
				{
					Query:    `SELECT count(*) FROM ST_SquareGrid(0.3, 'MULTIPOINT(-5.5 -5.5,7.25 3.1)'::geometry);`,
					Expected: []sql.Row{{1320}},
				},
				{
					Query:    `SELECT count(*) FROM ST_SquareGrid(0, 'POINT(1 1)'::geometry), ST_HexagonGrid(-1, 'POINT(1 1)'::geometry);`,
					Expected: []sql.Row{{0}},
				},
				{
					Query:    `SELECT count(*) FROM ST_HexagonGrid(1, 'POINT EMPTY'::geometry);`,
					Expected: []sql.Row{{0}},
				},
				{
					Query:    `SELECT * FROM ST_SquareGrid(1, 'SRID=4326;POINT(0.5 0.5)'::geometry);`,
					Expected: []sql.Row{{`0103000020E61000000100000005000000000000000000000000000000000000000000000000000000000000000000F03F000000000000F03F000000000000F03F000000000000F03F000000000000000000000000000000000000000000000000`, 0, 0}},
				},
				{
					Query: `SELECT ST_HexagonGrid(1, 'POINT(0.5 0.5)'::geometry);`,
					Expected: []sql.Row{
						{`(01030000000100000007000000000000000000F0BF0000000000000000000000000000E0BFAB4C58E87AB6EBBF000000000000E03FAB4C58E87AB6EBBF000000000000F03F0000000000000000000000000000E03FAB4C58E87AB6EB3F000000000000E0BFAB4C58E87AB6EB3F000000000000F0BF0000000000000000,0,0)`},
						{`(01030000000100000007000000000000000000E03FAB4C58E87AB6EB3F000000000000F03F0000000000000000000000000000004000000000000000000000000000000440AB4C58E87AB6EB3F0000000000000040AB4C58E87AB6FB3F000000000000F03FAB4C58E87AB6FB3F000000000000E03FAB4C58E87AB6EB3F,1,0)`},
					},
				},
				{
					Query:    `SELECT ST_AsText(ST_Square(0.3, 18, -19)), ST_AsText(ST_Square(0.3, 18, -19, 'POINT(0.1 0.2)'::geometry)), ST_AsText(ST_Hexagon(0.3, -7, -6)), ST_AsText(ST_Hexagon(2.5, -3, 7, 'POINT(1.1 -0.7)'::geometry));`,
					Expected: []sql.Row{{`POLYGON((5.4 -5.7,5.4 -5.4,5.7 -5.4,5.7 -5.7,5.4 -5.7))`, `POLYGON((5.499999999999999 -5.5,5.499999999999999 -5.199999999999999,5.8 -5.199999999999999,5.8 -5.5,5.499999999999999 -5.5))`, `POLYGON((-3.45 -2.857883832488648,-3.3 -3.11769145362398,-3 -3.11769145362398,-2.85 -2.857883832488648,-3 -2.598076211353316,-3.3 -2.598076211353316,-3.45 -2.857883832488648))`, `POLYGON((-12.65 31.77595264191645,-11.4 29.610889132455355,-8.9 29.610889132455355,-7.65 31.77595264191645,-8.9 33.941016151377546,-11.4 33.941016151377546,-12.65 31.77595264191645))`}},
				},
				{
					Query:    `SELECT pg_get_function_arguments(oid), pg_get_function_result(oid) FROM pg_proc WHERE proname = 'st_hexagongrid';`,
					Expected: []sql.Row{{`size double precision, bounds geometry, OUT geom geometry, OUT i integer, OUT j integer`, `SETOF record`}},
					Skip:     true,
				},
			},
		},
		{
			Name: `postgis kml input`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromKML('<Point><coordinates> 1 , 2 </coordinates></Point>'));`,
					Expected: []sql.Row{{`SRID=4326;POINT(1 2)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromKML('<kml:Point><kml:coordinates>1,2,3</kml:coordinates></kml:Point>'));`,
					Expected: []sql.Row{{`SRID=4326;POINT(1 2 3)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromKML('<Polygon><outerBoundaryIs><LinearRing><coordinates>0,0 1,0 1,1 0,0</coordinates></LinearRing></outerBoundaryIs><innerBoundaryIs><LinearRing><coordinates>0.1,0.1 0.2,0.1 0.2,0.2 0.1,0.1</coordinates></LinearRing></innerBoundaryIs></Polygon>'));`,
					Expected: []sql.Row{{`SRID=4326;POLYGON((0 0,1 0,1 1,0 0),(0.1 0.1,0.2 0.1,0.2 0.2,0.1 0.1))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromKML('<MultiGeometry><Polygon><outerBoundaryIs><LinearRing><coordinates>0,0 1,0 1,1 0,0</coordinates></LinearRing></outerBoundaryIs></Polygon><Point><coordinates>1,2</coordinates></Point><MultiGeometry><Point><coordinates>3,4</coordinates></Point></MultiGeometry><LineString><coordinates>0,0 1,1</coordinates></LineString></MultiGeometry>'));`,
					Expected: []sql.Row{{`SRID=4326;GEOMETRYCOLLECTION(MULTIPOINT(1 2,3 4),LINESTRING(0 0,1 1),POLYGON((0 0,1 0,1 1,0 0)))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromKML('<MultiGeometry><Point><coordinates>3,4,5</coordinates></Point><Point><coordinates>1,2</coordinates></Point></MultiGeometry>'));`,
					Expected: []sql.Row{{`SRID=4326;MULTIPOINT(3 4,1 2)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromKML('<MultiGeometry></MultiGeometry>'));`,
					Expected: []sql.Row{{`SRID=4326;GEOMETRYCOLLECTION EMPTY`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromKML('<Point xmlns="http://other"><extrude>1</extrude><!-- c --><coordinates><![CDATA[0x10,1e2]]></coordinates></Point>'));`,
					Expected: []sql.Row{{`SRID=4326;POINT(16 100)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromKML('<Polygon><outerBoundaryIs><LinearRing><coordinates>0,0 1,0 1,1 0,1</coordinates></LinearRing></outerBoundaryIs></Polygon>'));`,
					Expected: []sql.Row{{`SRID=4326;POLYGON((0 0,1 0,1 1,0 1,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromKML('<Point><coordinates>1,,2</coordinates></Point>'));`,
					Expected: []sql.Row{{`SRID=4326;POINT(1 2)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromKML('<k:Point xmlns:k="http://www.opengis.net/kml/2.2"><k:coordinates>1,2</k:coordinates></k:Point>'));`,
					Expected: []sql.Row{{`SRID=4326;POINT(1 2)`}},
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromKML('<o:Point xmlns:o="http://other"><o:coordinates>1,2</o:coordinates></o:Point>'));`,
					ExpectedErr:     `invalid KML representation`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromKML('<Point><coordinates>1,2,</coordinates></Point>'));`,
					ExpectedErr:     `invalid KML representation`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromKML('<LineString><coordinates>0,0,1 1,1</coordinates></LineString>'));`,
					ExpectedErr:     `invalid KML representation: mixed coordinates dimension`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromKML('<Polygon><outerBoundaryIs><LinearRing><coordinates>0,0 1,0 1,1</coordinates></LinearRing></outerBoundaryIs></Polygon>'));`,
					ExpectedErr:     `invalid KML representation`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromKML('<Placemark><Point><coordinates>1,2</coordinates></Point></Placemark>'));`,
					ExpectedErr:     `invalid KML representation`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromKML('<Point><coordinates>1,2</coordinates></Point>junk'));`,
					ExpectedErr:     `invalid KML representation`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromKML('<Point><coordinates>inf,2</coordinates></Point>'));`,
					ExpectedErr:     `invalid KML representation`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromKML('<LineString><coordinates>0,0  ,1,1</coordinates></LineString>'));`,
					ExpectedErr:     `invalid KML representation`,
					ExpectedErrCode: `XX000`,
				},
			},
		},
		{
			Name: `postgis gml input`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Point><gml:coordinates cs=";" ts="|" decimal=",">1,5;2</gml:coordinates></gml:Point>'));`,
					Expected: []sql.Row{{`POINT(1.5 2)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Point><gml:coord><gml:X>1</gml:X><gml:Y>2</gml:Y><gml:Z>3</gml:Z></gml:coord></gml:Point>'));`,
					Expected: []sql.Row{{`POINT(1 2 3)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Point><gml:pos dimension="3">1 2 3</gml:pos></gml:Point>'));`,
					Expected: []sql.Row{{`POINT(1 2 3)`}},
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Point><gml:pos>1 2 3</gml:pos></gml:Point>'));`,
					ExpectedErr:     `invalid GML representation`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Point><gml:pos>.5 2</gml:pos></gml:Point>'));`,
					ExpectedErr:     `invalid GML representation`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Point/>'));`,
					Expected: []sql.Row{{`POINT EMPTY`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:LineString><gml:pos>0 0</gml:pos><gml:posList>1 1 2 2</gml:posList><gml:coordinates>3,3</gml:coordinates><gml:pointProperty><gml:Point><gml:pos>4 4</gml:pos></gml:Point></gml:pointProperty></gml:LineString>'));`,
					Expected: []sql.Row{{`LINESTRING(0 0,1 1,2 2,3 3,4 4)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:LineString><gml:posList>0 0 0 1 1 1</gml:posList></gml:LineString>'));`,
					Expected: []sql.Row{{`LINESTRING(0 0,0 1,1 1)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:LineString><gml:coordinates>1,2,3 4,5</gml:coordinates></gml:LineString>'));`,
					Expected: []sql.Row{{`LINESTRING(1 2,4 5)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Curve><gml:segments><gml:LineStringSegment><gml:posList>0 0 1 1</gml:posList></gml:LineStringSegment><gml:LineStringSegment><gml:posList>1 1 2 2</gml:posList></gml:LineStringSegment></gml:segments></gml:Curve>'));`,
					Expected: []sql.Row{{`LINESTRING(0 0,1 1,2 2)`}},
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Curve><gml:segments><gml:ArcString><gml:posList>0 0 1 1 2 0</gml:posList></gml:ArcString></gml:segments></gml:Curve>'));`,
					ExpectedErr:     `invalid GML representation`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Polygon><gml:outerBoundaryIs><gml:LinearRing><gml:coordinates>0,0 1,0 1,1 0,0</gml:coordinates></gml:LinearRing></gml:outerBoundaryIs><gml:innerBoundaryIs><gml:LinearRing><gml:coordinates>0.1,0.1 0.2,0.1 0.2,0.2 0.1,0.1</gml:coordinates></gml:LinearRing></gml:innerBoundaryIs></gml:Polygon>'));`,
					Expected: []sql.Row{{`POLYGON((0 0,1 0,1 1,0 0),(0.1 0.1,0.2 0.1,0.2 0.2,0.1 0.1))`}},
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Polygon><gml:exterior><gml:LinearRing><gml:posList>0 0 1 0 1 1 0 1</gml:posList></gml:LinearRing></gml:exterior></gml:Polygon>'));`,
					ExpectedErr:     `invalid GML representation`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Surface><gml:patches><gml:PolygonPatch><gml:exterior><gml:LinearRing><gml:posList srsDimension="3">0 0 1 1 0 1 1 1 1 0 0 1</gml:posList></gml:LinearRing></gml:exterior></gml:PolygonPatch></gml:patches></gml:Surface>'));`,
					Expected: []sql.Row{{`POLYGON((0 0 1,1 0 1,1 1 1,0 0 1))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:TriangulatedSurface><gml:trianglePatches><gml:Triangle><gml:exterior><gml:LinearRing><gml:posList>0 0 1 0 1 1 0 0</gml:posList></gml:LinearRing></gml:exterior></gml:Triangle></gml:trianglePatches></gml:TriangulatedSurface>'));`,
					Expected: []sql.Row{{`TIN(((0 0,1 0,1 1,0 0)))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:PolyhedralSurface><gml:polygonPatches><gml:PolygonPatch><gml:exterior><gml:LinearRing><gml:posList>0 0 1 0 1 1 0 0</gml:posList></gml:LinearRing></gml:exterior></gml:PolygonPatch></gml:polygonPatches></gml:PolyhedralSurface>'));`,
					Expected: []sql.Row{{`POLYHEDRALSURFACE(((0 0,1 0,1 1,0 0)))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:LinearRing><gml:posList>0 0 1 0 1 1 0 0</gml:posList></gml:LinearRing>'));`,
					Expected: []sql.Row{{`POLYGON((0 0,1 0,1 1,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:MultiPoint><gml:pointMembers><gml:Point><gml:pos>1 2</gml:pos></gml:Point><gml:Point><gml:pos>3 4</gml:pos></gml:Point></gml:pointMembers></gml:MultiPoint>'));`,
					Expected: []sql.Row{{`MULTIPOINT(1 2,3 4)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:MultiCurve><gml:curveMember><gml:LineString><gml:posList>0 0 1 1</gml:posList></gml:LineString></gml:curveMember><gml:curveMember><gml:Curve><gml:segments><gml:LineStringSegment><gml:posList>2 2 3 3</gml:posList></gml:LineStringSegment></gml:segments></gml:Curve></gml:curveMember></gml:MultiCurve>'));`,
					Expected: []sql.Row{{`MULTILINESTRING((0 0,1 1),(2 2,3 3))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:MultiSurface><gml:Polygon><gml:exterior><gml:LinearRing><gml:posList>0 0 1 0 1 1 0 0</gml:posList></gml:LinearRing></gml:exterior></gml:Polygon></gml:MultiSurface>'));`,
					Expected: []sql.Row{{`MULTIPOLYGON EMPTY`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:MultiGeometry><gml:geometryMember><gml:MultiGeometry><gml:geometryMember><gml:Point><gml:pos>1 2</gml:pos></gml:Point></gml:geometryMember></gml:MultiGeometry></gml:geometryMember><gml:geometryMember><gml:Point/></gml:geometryMember></gml:MultiGeometry>'));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT(1 2)),POINT EMPTY)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:MultiPoint><gml:pointMember><gml:Point><gml:pos>1 2</gml:pos></gml:Point></gml:pointMember><gml:pointMember><gml:Point><gml:pos srsDimension="3">1 2 3</gml:pos></gml:Point></gml:pointMember></gml:MultiPoint>'));`,
					Expected: []sql.Row{{`MULTIPOINT(1 2,1 2)`}},
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:MultiGeometry><gml:geometryMember><gml:Point><gml:pos srsDimension="3">1 2 3</gml:pos></gml:Point></gml:geometryMember><gml:geometryMember><gml:Point/></gml:geometryMember></gml:MultiGeometry>'));`,
					ExpectedErr:     `Dimensions mismatch in lwcollection`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:MultiSurface><gml:surfaceMember><gml:Point><gml:pos>1 2</gml:pos></gml:Point></gml:surfaceMember></gml:MultiSurface>'));`,
					ExpectedErr:     `MultiPolygon cannot contain Point element`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Point srsName="urn:ogc:def:crs:EPSG::4326"><gml:pos>1 2</gml:pos></gml:Point>'));`,
					Expected: []sql.Row{{`SRID=4326;POINT(2 1)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Point srsName="urn:ogc:def:crs:EPSG:6.6:4258"><gml:coordinates>1,2</gml:coordinates></gml:Point>'));`,
					Expected: []sql.Row{{`SRID=4258;POINT(2 1)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Point srsName="http://www.opengis.net/gml/srs/epsg.xml#4326"><gml:pos>1 2</gml:pos></gml:Point>'));`,
					Expected: []sql.Row{{`SRID=4326;POINT(1 2)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:MultiPoint><gml:pointMember><gml:Point srsName="urn:ogc:def:crs:EPSG::3857"><gml:pos>1 2</gml:pos></gml:Point></gml:pointMember></gml:MultiPoint>'));`,
					Expected: []sql.Row{{`SRID=3857;MULTIPOINT(1 2)`}},
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Point srsName="EPSG:999999"><gml:pos>1 2</gml:pos></gml:Point>'));`,
					ExpectedErr:     `unknown spatial reference system`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:MultiPoint srsName="EPSG:4326"><gml:pointMember><gml:Point srsName="EPSG:3857"><gml:pos>1 2</gml:pos></gml:Point></gml:pointMember></gml:MultiPoint>'));`,
					Expected: []sql.Row{{`SRID=4326;MULTIPOINT(0.000008983152841 0.000017966305682)`}},
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromGML('<o:Point xmlns:o="http://other"><o:pos>1 2</o:pos></o:Point>'));`,
					ExpectedErr:     `invalid GML representation`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<Point xmlns="http://other"><name>x</name><pos>1 2</pos></Point>'));`,
					Expected: []sql.Row{{`POINT(1 2)`}},
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_GeomFromGML('garbage'));`,
					ExpectedErr:     `invalid GML representation`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Point srsName="EPSG:4326"><gml:pos>1 2</gml:pos></gml:Point>', 3857)), ST_AsEWKT(ST_GMLToSQL('<gml:Point><gml:pos>1 2</gml:pos></gml:Point>', 4326)), ST_AsEWKT(_ST_GeomFromGML('<gml:Point><gml:pos>1 2</gml:pos></gml:Point>', NULL));`,
					Expected: []sql.Row{{`SRID=3857;POINT(1 2)`, `SRID=4326;POINT(1 2)`, `POINT(1 2)`}},
				},
			},
		},
		{
			Name: `postgis distance operators`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
				`CREATE TABLE geo_knn (id INT PRIMARY KEY, g geometry, gg geography);`,
				`INSERT INTO geo_knn VALUES (1, 'POINT(0 0)', 'POINT(0 0)'), (2, 'POINT(5 5)', 'POINT(5 5)'), (3, 'LINESTRING(10 0,10 10)', 'LINESTRING(10 0,10 10)'), (4, 'POLYGON((-1 -1,1 -1,1 1,-1 -1))', 'POINT(1 1)');`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT 'POINT(0 0)'::geometry <-> 'POINT(3 4)'::geometry, 'LINESTRING(0 0,10 0)'::geometry <-> 'POINT(5 5)'::geometry, 'POLYGON((0 0,10 0,10 10,0 0))'::geometry <-> 'POINT(9 1)'::geometry;`,
					Expected: []sql.Row{{5.0, 5.0, 0.0}},
				},
				{
					Query:    `SELECT round(('POINT(0 0)'::geography <-> 'POINT(1 1)'::geography)::numeric, 4), round(('LINESTRING(0 0,10 0)'::geography <-> 'POINT(5 1)'::geography)::numeric, 4);`,
					Expected: []sql.Row{{framework.Numeric(`157249.5978`), framework.Numeric(`111195.0797`)}},
				},
				{
					Query:    `SELECT ('POINT EMPTY'::geography <-> 'POINT(1 1)'::geography) IS NULL, ('POINT EMPTY'::geometry <-> 'POINT(1 1)'::geometry) IS NULL;`,
					Expected: []sql.Row{{`t`, `t`}},
				},
				{
					Query:           `SELECT 'SRID=4326;POINT(0 0)'::geometry <-> 'POINT(1 1)'::geometry;`,
					ExpectedErr:     `ST_Distance: Operation on mixed SRID geometries (Point, 4326) != (Point, 0)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT 'SRID=4326;POINT(0 0)'::geography <-> 'SRID=4269;POINT(1 1)'::geography;`,
					ExpectedErr:     `geography_distance_knn: Operation on mixed SRID geometries (Point, 4326) != (Point, 4269)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT 'LINESTRING M (0 0 0,10 0 10)'::geometry |=| 'LINESTRING M (0 5 0,10 5 10)'::geometry, 'LINESTRING M (0 0 0,10 0 10)'::geometry |=| 'LINESTRING M (0 5 20,10 5 30)'::geometry;`,
					Expected: []sql.Row{{5.0, nil}},
				},
				{
					Query:           `SELECT 'LINESTRING(0 0,1 1)'::geometry |=| 'LINESTRING M (0 5 0,10 5 10)'::geometry;`,
					ExpectedErr:     `Both input geometries must have a measure dimension`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query: `SELECT id FROM geo_knn ORDER BY g <-> 'POINT(9 1)'::geometry;`,
					Expected: []sql.Row{
						{3},
						{2},
						{4},
						{1},
					},
				},
				{
					Query: `SELECT id FROM geo_knn ORDER BY gg <-> 'POINT(4 4)'::geography LIMIT 2;`,
					Expected: []sql.Row{
						{2},
						{4},
					},
				},
				{
					Query:    `SELECT 'LINESTRING Z (0 0 0,10 0 4)'::geometry <<->> 'POINT Z (20 0 0)'::geometry;`,
					Expected: []sql.Row{{10.770329614269007}},
				},
			},
		},
		{
			Name: `postgis internal functions`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT _st_contains('POLYGON((0 0,2 0,2 2,0 0))'::geometry, 'POINT(1.5 0.5)'::geometry), _st_containsproperly('POLYGON((0 0,2 0,2 2,0 0))'::geometry, 'POINT(1.5 0.5)'::geometry), _st_coveredby('POINT(1.5 0.5)'::geometry, 'POLYGON((0 0,2 0,2 2,0 0))'::geometry), _st_covers('POLYGON((0 0,2 0,2 2,0 0))'::geometry, 'POINT(2 0)'::geometry), _st_crosses('LINESTRING(0 0,2 2)'::geometry, 'LINESTRING(0 2,2 0)'::geometry), _st_equals('POINT(1 1)'::geometry, 'POINT(1 1)'::geometry), _st_intersects('POINT(1 1)'::geometry, 'LINESTRING(0 0,2 2)'::geometry);`,
					Expected: []sql.Row{{`t`, `t`, `t`, `t`, `t`, `t`, `t`}},
				},
				{
					Query:    `SELECT _st_orderingequals('LINESTRING(0 0,1 1)'::geometry, 'LINESTRING(1 1,0 0)'::geometry), _st_overlaps('POLYGON((0 0,2 0,2 2,0 0))'::geometry, 'POLYGON((1 0,3 0,3 2,1 0))'::geometry), _st_touches('POINT(0 0)'::geometry, 'LINESTRING(0 0,1 1)'::geometry), _st_within('POINT(1.5 0.5)'::geometry, 'POLYGON((0 0,2 0,2 2,0 0))'::geometry), _st_within(NULL, 'POINT(1 1)'::geometry) IS NULL;`,
					Expected: []sql.Row{{`f`, `t`, `t`, `t`, `t`}},
				},
				{
					Query:           `SELECT _st_within('POINT(1 1)'::geometry, 'SRID=4326;POLYGON((0 0,2 0,2 2,0 0))'::geometry);`,
					ExpectedErr:     `contains: Operation on mixed SRID geometries (Polygon, 4326) != (Point, 0)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT _st_dfullywithin('POINT(0 0)'::geometry, 'LINESTRING(0 0,1 1)'::geometry, 2), _st_dwithin('POINT(0 0)'::geometry, 'POINT(3 4)'::geometry, 5), ST_AsText(_st_longestline('POINT(0 0)'::geometry, 'LINESTRING(1 1,5 0)'::geometry)), _st_maxdistance('POINT(0 0)'::geometry, 'LINESTRING(1 1,5 0)'::geometry), _st_linecrossingdirection('LINESTRING(0 0,10 10)'::geometry, 'LINESTRING(0 10,10 0)'::geometry);`,
					Expected: []sql.Row{{`t`, `t`, `LINESTRING(0 0,5 0)`, 5.0, 1}},
				},
				{
					Query:    `SELECT round(_st_distanceuncached('POINT(0 0)'::geography, 'POINT(1 1)'::geography)::numeric, 6), round(_st_distanceuncached('POINT(0 0)'::geography, 'POINT(1 1)'::geography, 0, true)::numeric, 6), _st_distancetree('POINT(0 0)'::geography, 'POINT(1 1)'::geography), _st_distancetree('POINT(0 0)'::geography, 'POINT(1 1)'::geography, 0, false);`,
					Expected: []sql.Row{{framework.Numeric(`156899.568291`), framework.Numeric(`156899.568291`), 156899.56829134, 157249.59776851}},
				},
				{
					Query:    `SELECT round(_st_distanceuncached('POINT(0 0)'::geography, 'POINT(1 1)'::geography, false)::numeric, 4);`,
					Expected: []sql.Row{{framework.Numeric(`157249.5978`)}},
				},
				{
					Query:    `SELECT _st_dwithinuncached('POINT(0 0)'::geography, 'POINT(1 1)'::geography, 157000), _st_dwithinuncached('POINT(0 0)'::geography, 'POINT(1 1)'::geography, 157000, false), _st_dwithin('POINT(0 0)'::geography, 'POINT(1 1)'::geography, 157000), _st_covers('POLYGON((0 0,2 0,2 2,0 0))'::geography, 'POINT(1.5 0.5)'::geography), _st_coveredby('POINT(1.5 0.5)'::geography, 'POLYGON((0 0,2 0,2 2,0 0))'::geography);`,
					Expected: []sql.Row{{`t`, `f`, `t`, `t`, `t`}},
				},
			},
		},
		{
			Name: `postgis legacy functions`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_PointInsideCircle('POINT(1 1)'::geometry, 0, 0, 2), ST_PointInsideCircle('POINT(1 1)'::geometry, 0, 0, 1.4142135623730951), ST_PointInsideCircle('POINT(1 1)'::geometry, 0, 0, 1);`,
					Expected: []sql.Row{{`t`, `f`, `f`}},
				},
				{
					Query:    `SELECT ST_PointInsideCircle('LINESTRING(1 1,2 2)'::geometry, 0, 0, 2) IS NULL, ST_PointInsideCircle('POINT EMPTY'::geometry, 0, 0, 2) IS NULL;`,
					Expected: []sql.Row{{`t`, `t`}},
				},
				{
					Query:    `SELECT ST_PointInsideCircle('POINT(1 1)'::geometry, 0, 0, -1), ST_PointInsideCircle('POINT Z (1 1 5)'::geometry, 1, 1, 0.0001), ST_PointInsideCircle('POINT(0.1 0.2)'::geometry, 0.4, 0.6, 0.5);`,
					Expected: []sql.Row{{`f`, `t`, `f`}},
				},
				{
					Query:    `SELECT equals('POINT(1 1)'::geometry, 'POINT(1 1)'::geometry), equals('LINESTRING(0 0,1 1)'::geometry, 'LINESTRING(1 1,0 0)'::geometry);`,
					Expected: []sql.Row{{`t`, `t`}},
				},
			},
		},
		{
			Name: `postgis transform`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
				`INSERT INTO spatial_ref_sys VALUES (990000, 'custom', 1, NULL, '+proj=merc +datum=WGS84 +units=m +no_defs'), (990001, NULL, NULL, NULL, NULL), (990002, 'custom', 1, 'garbage', 'garbage');`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsEWKT(ST_Transform('SRID=4326;POINT(-71.06 42.36)'::geometry, 3857), 6), ST_AsEWKT(ST_Transform('SRID=4326;POINT Z (-71.06 42.36 10)'::geometry, 2249), 6), ST_AsEWKT(ST_Transform('SRID=4326;POINT M (-71.06 42.36 7)'::geometry, 3857), 6);`,
					Expected: []sql.Row{{`SRID=3857;POINT(-7910363.01577 5215059.174611)`, `SRID=2249;POINT(775086.491274 2956519.888036 10)`, `SRID=3857;POINTM(-7910363.01577 5215059.174611 7)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Transform('SRID=4326;LINESTRING(-71.06 42.36,-71.1 42.4)'::geometry, 26986), 6), ST_AsEWKT(ST_Transform('SRID=3857;POLYGON((0 0,100000 0,100000 100000,0 0))'::geometry, 4326), 9);`,
					Expected: []sql.Row{{`SRID=26986;LINESTRING(236246.834933 901149.064171,232930.857459 905575.880819)`, `SRID=4326;POLYGON((0 0,0.898315284 0,0.898315284 0.898278483,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Transform('SRID=4326;POINT EMPTY'::geometry, 3857)), ST_AsEWKT(ST_Transform('SRID=4326;POINT(1 2)'::geometry, 4326)), ST_AsEWKT(ST_Transform('SRID=123456;POINT(1 2)'::geometry, 123456));`,
					Expected: []sql.Row{{`SRID=3857;POINT EMPTY`, `SRID=4326;POINT(1 2)`, `SRID=123456;POINT(1 2)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Transform('SRID=4326;GEOMETRYCOLLECTION(POINT(1 2),LINESTRING(1 2,3 4))'::geometry, 3857), 6), ST_AsEWKT(ST_Transform('SRID=4326;CIRCULARSTRING(0 0,1 1,2 0)'::geometry, 3857), 6);`,
					Expected: []sql.Row{{`SRID=3857;GEOMETRYCOLLECTION(POINT(111319.490793 222684.208506),LINESTRING(111319.490793 222684.208506,333958.47238 445640.109656))`, `SRID=3857;CIRCULARSTRING(0 0,111319.490793 111325.142866,222638.981587 0)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Transform('SRID=4326;POINT(1 2 3 4)'::geometry, 4978), 6), ST_AsEWKT(ST_Transform('SRID=4326;POINT(1 2)'::geometry, 4979), 6), ST_AsEWKT(ST_Transform('SRID=4326;POINT(1 2)'::geometry, 102003), 6);`,
					Expected: []sql.Row{{`SRID=4978;POINT(6373309.762299 111246.535709 221104.65001 4)`, `SRID=4979;POINT(1 2)`, `SRID=102003;POINT(10311643.196487 1998603.959972)`}},
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_Transform('POINT(1 2)'::geometry, 3857));`,
					ExpectedErr:     `ST_Transform: Input geometry has unknown (0) SRID`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_Transform('SRID=4326;POINT(1 2)'::geometry, 0));`,
					ExpectedErr:     `ST_Transform: 0 is an invalid target SRID`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_Transform('SRID=4326;POINT(1 2)'::geometry, 123456));`,
					ExpectedErr:     `Cannot find SRID (123456) in spatial_ref_sys`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_Transform('SRID=4326;POINT(200 100)'::geometry, 3857));`,
					ExpectedErr:     `transform: Invalid coordinate (2049)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Transform('SRID=4326;POINT(1 2)'::geometry, 990000), 6);`,
					Expected: []sql.Row{{`SRID=990000;POINT(111319.490793 221194.077168)`}},
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_Transform('SRID=4326;POINT(1 2)'::geometry, 990001));`,
					ExpectedErr:     `got NULL for SRID (990001)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_Transform('SRID=4326;POINT(1 2)'::geometry, 990002));`,
					ExpectedErr:     `could not form projection (LWPROJ) from 'srid=4326' to 'srid=990002'`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_Transform('SRID=4326;POINT(1 2)'::geometry, '+proj=merc +datum=WGS84'), 6), ST_AsEWKT(ST_Transform('POINT(1 2)'::geometry, '+proj=longlat +datum=WGS84', '+proj=merc +datum=WGS84'), 6), ST_AsEWKT(ST_Transform('POINT(1 2)'::geometry, 'EPSG:4326', 3857), 6);`,
					Expected: []sql.Row{{`POINT(111319.490793 221194.077168)`, `POINT(111319.490793 221194.077168)`, `SRID=3857;POINT(111319.490793 222684.208506)`}},
				},
				{
					Query:    `SELECT ST_Transform('POINT(1 2)'::geometry, 'EPSG:4326', 123456) IS NULL, ST_Transform('SRID=123456;POINT(1 2)'::geometry, '+proj=merc') IS NULL;`,
					Expected: []sql.Row{{`t`, `t`}},
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_Transform('SRID=4326;POINT(1 2)'::geometry, 'garbage'));`,
					ExpectedErr:     `could not parse proj string 'garbage'`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_Transform('POINT(1 2)'::geometry, 'garbage', 3857));`,
					ExpectedErr:     `could not parse proj string 'garbage'`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_TransformPipeline('SRID=4326;POINT(2 49)'::geometry, 'urn:ogc:def:coordinateOperation:EPSG::16031'), 3), ST_AsEWKT(ST_TransformPipeline('SRID=4326;POINT(2 49)'::geometry, 'urn:ogc:def:coordinateOperation:EPSG::16031', 32631), 3), ST_AsEWKT(ST_InverseTransformPipeline('SRID=32631;POINT(426857.988 5427937.523)'::geometry, 'urn:ogc:def:coordinateOperation:EPSG::16031', 4326), 6);`,
					Expected: []sql.Row{{`POINT(426857.988 5427937.523)`, `SRID=32631;POINT(426857.988 5427937.523)`, `SRID=4326;POINT(2 49)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_TransformPipeline('SRID=4326;POINT(2 49)'::geometry, '+proj=pipeline +step +proj=unitconvert +xy_in=deg +xy_out=rad +step +proj=utm +zone=31 +ellps=WGS84'), 3), ST_AsEWKT(postgis_transform_geometry('POINT(1 2)'::geometry, 'EPSG:4326', 'EPSG:3857', 3857), 6), ST_AsEWKT(postgis_transform_pipeline_geometry('SRID=4326;POINT(2 49)'::geometry, 'urn:ogc:def:coordinateOperation:EPSG::16031', true, 1), 3);`,
					Expected: []sql.Row{{`POINT(426857.988 5427937.523)`, `SRID=3857;POINT(111319.490793 222684.208506)`, `SRID=1;POINT(426857.988 5427937.523)`}},
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_TransformPipeline('SRID=4326;POINT(2 49)'::geometry, 'garbage'));`,
					ExpectedErr:     `could not parse coordinate operation 'garbage'`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_TransformPipeline('SRID=4326;POINT(2 49)'::geometry, 'EPSG:4326'));`,
					ExpectedErr:     `lwgeom_transform_pipeline: Failed to transform`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query: `SELECT srid, auth_name, auth_srid, rtrim(proj4text) FROM spatial_ref_sys WHERE srid IN (4326, 3857, 102003) ORDER BY srid;`,
					Expected: []sql.Row{
						{3857, `EPSG`, 3857, `+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0 +k=1.0 +units=m +nadgrids=@null +wktext +no_defs`},
						{4326, `EPSG`, 4326, `+proj=longlat +datum=WGS84 +no_defs`},
						{102003, `ESRI`, 102003, `+proj=aea +lat_0=37.5 +lon_0=-96 +lat_1=29.5 +lat_2=45.5 +x_0=0 +y_0=0 +datum=NAD83 +units=m +no_defs`},
					},
					Skip: true,
				},
				{
					Query:    `SELECT srtext FROM spatial_ref_sys WHERE srid = 4326;`,
					Expected: []sql.Row{{`GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563,AUTHORITY["EPSG","7030"]],AUTHORITY["EPSG","6326"]],PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]],UNIT["degree",0.0174532925199433,AUTHORITY["EPSG","9122"]],AUTHORITY["EPSG","4326"]]`}},
				},
				{
					Query: `SELECT srid, auth_name, auth_srid, srtext IS NOT NULL, proj4text IS NOT NULL FROM spatial_ref_sys WHERE srid IN (2249, 3857, 4326, 26986, 32631, 102003) ORDER BY srid;`,
					Expected: []sql.Row{
						{2249, `EPSG`, 2249, `t`, `t`},
						{3857, `EPSG`, 3857, `t`, `t`},
						{4326, `EPSG`, 4326, `t`, `t`},
						{26986, `EPSG`, 26986, `t`, `t`},
						{32631, `EPSG`, 32631, `t`, `t`},
						{102003, `ESRI`, 102003, `t`, `t`},
					},
				},
				{
					Query:    `SELECT count(*) > 8000, min(srid), bool_and(srid = auth_srid) FROM spatial_ref_sys WHERE auth_name IN ('EPSG', 'ESRI');`,
					Expected: []sql.Row{{`t`, 2000, `t`}},
				},
				{
					Query:    `SELECT get_proj4_from_srid(4326);`,
					Expected: []sql.Row{{`+proj=longlat +datum=WGS84 +no_defs `}},
					Skip:     true,
				},
			},
		},
		{
			Name: `postgis best srid`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT _ST_BestSRID('POINT(-71 42)'::geography), _ST_BestSRID('POINT(-71 -42)'::geography), _ST_BestSRID('POINT(10 80)'::geography), _ST_BestSRID('POINT(10 -80)'::geography);`,
					Expected: []sql.Row{{999019, 999119, 999061, 999161}},
				},
				{
					Query:    `SELECT _ST_BestSRID('POINT(10 40)'::geography, 'POINT(12 41)'::geography), _ST_BestSRID('POINT(10 40)'::geography, 'POINT(40 41)'::geography), _ST_BestSRID('POINT(10 80)'::geography, 'POINT(100 85)'::geography), _ST_BestSRID('POINT(-179 0)'::geography, 'POINT(179 0)'::geography);`,
					Expected: []sql.Row{{999032, 999247, 999061, 999060}},
				},
				{
					Query:    `SELECT _ST_BestSRID(ST_MakeEnvelope(10, 40, 15, 45, 4326)::geography), _ST_BestSRID(ST_MakeEnvelope(10, 40, 30, 50, 4326)::geography), _ST_BestSRID(ST_MakeEnvelope(10, 10, 40, 12, 4326)::geography), _ST_BestSRID(ST_MakeEnvelope(-170, -65, -100, -55, 4326)::geography), _ST_BestSRID(ST_MakeEnvelope(-179, 46, 179, 89.9, 4326)::geography), _ST_BestSRID(ST_MakeEnvelope(-179, 45, 179, 89.9, 4326)::geography);`,
					Expected: []sql.Row{{999247, 999247, 999000, 999000, 999061, 999000}},
				},
				{
					Query:    `SELECT _ST_BestSRID('POLYGON((0 80,120 80,-120 80,0 80))'::geography), _ST_BestSRID('LINESTRING(0 89,180 89)'::geography);`,
					Expected: []sql.Row{{999061, 999061}},
				},
			},
		},
		{
			Name: `postgis geography buffer and intersection`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_NPoints(ST_Buffer('POINT(-71.06 42.36)'::geography, 1000)::geometry), ST_AsText(ST_SnapToGrid(ST_Buffer('POINT(-71.06 42.36)'::geography, 1000, 2)::geometry, 0.000001));`,
					Expected: []sql.Row{{33, `POLYGON((-71.047864 42.360217,-71.051212 42.35379,-71.059706 42.351,-71.068372 42.353481,-71.072135 42.359781,-71.06879 42.36621,-71.060294 42.369,-71.051626 42.366518,-71.047864 42.360217))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_SnapToGrid(ST_Buffer('LINESTRING(-71.06 42.36,-71 42.4)'::geography, 500, 'endcap=flat quad_segs=2')::geometry, 0.000001));`,
					Expected: []sql.Row{{`POLYGON((-71.00406 42.403348,-70.995941 42.396652,-71.05594 42.356654,-71.06406 42.363346,-71.00406 42.403348))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_SnapToGrid(ST_Intersection('POLYGON((0 0,2 0,2 2,0 2,0 0))'::geography, 'POLYGON((1 1,3 1,3 3,1 3,1 1))'::geography)::geometry, 0.000001));`,
					Expected: []sql.Row{{`POLYGON((2 2,2.000151 1.000153,1 1,1.000303 2.000307,2 2))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_SnapToGrid(ST_Buffer('POINT(10 80)'::geography, 10000, 1)::geometry, 0.000001)), ST_SRID(ST_Buffer('POINT(10 80)'::geography, 10000, 1));`,
					Expected: []sql.Row{{`POLYGON((10.327928 79.930965,9.608763 79.941977,9.667536 80.068698,10.395772 80.057547,10.327928 79.930965))`, 4326}},
				},
			},
		},
		{
			Name: `postgis library versions`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT postgis_geos_version(), postgis_geos_compiled_version(), split_part(postgis_proj_version(), ' ', 1), split_part(postgis_proj_version(), ' ', 2);`,
					Expected: []sql.Row{{`3.11.1-CAPI-1.17.1`, `3.11.1`, `9.8.1`, `NETWORK_ENABLED=OFF`}},
				},
				{
					Query:    `SELECT postgis_full_version() LIKE '%GEOS="3.11.1-CAPI-1.17.1" PROJ="9.8.1 NETWORK_ENABLED=OFF%';`,
					Expected: []sql.Row{{`t`}},
				},
				{
					Query:    `SELECT postgis_wagyu_version(), postgis_full_version() LIKE '%WAGYU="0.5.0 (Internal)"%';`,
					Expected: []sql.Row{{`0.5.0 (Internal)`, `t`}},
				},
				{
					Query:    `SELECT postgis_lib_build_date() ~ '^\d{4}-\d\d-\d\d \d\d:\d\d:\d\d$', postgis_scripts_build_date() ~ '^\d{4}-\d\d-\d\d \d\d:\d\d:\d\d$';`,
					Expected: []sql.Row{{`t`, `t`}},
				},
				{
					Query: `SELECT proname, pg_get_function_arguments(oid), prorettype::regtype FROM pg_proc WHERE proname IN ('postgis_lib_build_date', 'postgis_scripts_build_date') ORDER BY 1;`,
					Expected: []sql.Row{
						{`postgis_lib_build_date`, ``, `text`},
						{`postgis_scripts_build_date`, ``, `text`},
					},
				},
				{
					Query: `SELECT proname, provolatile, count(*) FROM pg_proc WHERE proname IN ('st_area', 'st_transform', 'st_generatepoints', 'st_symmetricdifference', 'find_srid', 'st_estimatedextent', 'addgeometrycolumn', 'postgis_lib_build_date', '_postgis_pgsql_version') GROUP BY 1, 2 ORDER BY 1, 2;`,
					Expected: []sql.Row{
						{`_postgis_pgsql_version`, `s`, 1},
						{`addgeometrycolumn`, `s`, 1},
						{`addgeometrycolumn`, `v`, 2},
						{`find_srid`, `s`, 1},
						{`postgis_lib_build_date`, `i`, 1},
						{`st_area`, `i`, 3},
						{`st_estimatedextent`, `s`, 3},
						{`st_generatepoints`, `i`, 1},
						{`st_generatepoints`, `v`, 1},
						{`st_symmetricdifference`, `v`, 1},
						{`st_transform`, `i`, 4},
					},
				},
			},
		},
		{
			Name: `postgis snap to grid rounding`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsText(ST_SnapToGrid('POINT(-71.0478641 42.3567571)'::geometry, 0.000001)), ST_AsText(ST_SnapToGrid('POINT(1.0000003 2.07)'::geometry, 0.1)), ST_AsText(ST_SnapToGrid('POINT(0.3 0.7)'::geometry, 0.1, 0.2)), ST_AsText(ST_SnapToGrid('POINT(5.55 2.333)'::geometry, 1, 1, 0.1, 0.01));`,
					Expected: []sql.Row{{`POINT(-71.047864 42.356757)`, `POINT(1 2.1)`, `POINT(0.3 0.8)`, `POINT(5.6 2.33)`}},
				},
			},
		},
		{
			Name: `postgis geos constructions`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsText(ST_MakeValid('POLYGON((0 0,2 2,2 0,0 2,0 0))'::geometry));`,
					Expected: []sql.Row{{`MULTIPOLYGON(((0 2,1 1,0 0,0 2)),((2 0,1 1,2 2,2 0)))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_MakeValid('POLYGON((0 0,2 2,2 0,0 2,0 0))'::geometry, 'method=structure'));`,
					Expected: []sql.Row{{`MULTIPOLYGON(((1 1,2 2,2 0,1 1)),((0 0,0 2,1 1,0 0)))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_MakeValid('LINESTRING(0 0,0 0)'::geometry)), ST_AsText(ST_MakeValid('LINESTRING(0 0,0 0)'::geometry, 'method=structure keepcollapsed=true'));`,
					Expected: []sql.Row{{`POINT(0 0)`, `POINT(0 0)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_MakeValid('POLYGON Z ((0 0 1,2 2 1,2 0 1,0 2 1,0 0 1))'::geometry)), ST_AsText(ST_MakeValid('POLYGON M ((0 0 1,2 2 1,2 0 1,0 2 1,0 0 1))'::geometry));`,
					Expected: []sql.Row{{`MULTIPOLYGON Z (((0 2 1,1 1 1,0 0 1,0 2 1)),((2 0 1,1 1 1,2 2 1,2 0 1)))`, `MULTIPOLYGON(((0 2,1 1,0 0,0 2)),((2 0,1 1,2 2,2 0)))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_MakeValid('SRID=4326;POINT(1 1)'::geometry)), ST_AsText(ST_MakeValid('POLYGON EMPTY'::geometry)), ST_AsText(ST_MakeValid('MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((1 1,3 1,3 3,1 3,1 1)))'::geometry));`,
					Expected: []sql.Row{{`SRID=4326;POINT(1 1)`, `POLYGON EMPTY`, `MULTIPOLYGON(((0 2,1 2,1 1,2 1,2 0,0 0,0 2)),((1 3,3 3,3 1,2 1,2 2,1 2,1 3)))`}},
				},
				{
					Query:           `SELECT ST_AsText(ST_MakeValid('POLYGON((0 0,2 2,2 0,0 2,0 0))'::geometry, 'method=foo'));`,
					ExpectedErr:     `Unsupported value for 'method', 'foo'. Use 'linework' or 'structure'.`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsText(ST_MakeValid('POLYGON((0 0,2 2,2 0,0 2,0 0))'::geometry, 'keepcollapsed=maybe'));`,
					ExpectedErr:     `Unsupported value for 'keepcollapsed', 'maybe'. Use 'true' or 'false'`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsText(ST_MakeValid('CIRCULARSTRING(0 0,1 1,2 0)'::geometry));`,
					ExpectedErr:     `ST_MakeValid: unsupported geometry type CircularString`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsText(ST_SimplifyPreserveTopology('LINESTRING(0 0,1 0.1,2 0,3 0.1,4 0)'::geometry, 0.5)), ST_AsText(ST_SimplifyPreserveTopology('POLYGON((0 0,10 0,10 10,5 10.1,0 10,0 0),(4 4,6 4,6 6,4 6,4 4))'::geometry, 1));`,
					Expected: []sql.Row{{`LINESTRING(0 0,4 0)`, `POLYGON((0 0,10 0,10 10,0 10,0 0),(4 4,6 4,6 6,4 6,4 4))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_SimplifyPreserveTopology('LINESTRING Z (0 0 1,1 0.1 2,2 0 3)'::geometry, 0.5)), ST_AsText(ST_SimplifyPreserveTopology('POINT(1 1)'::geometry, 1));`,
					Expected: []sql.Row{{`LINESTRING Z (0 0 1,2 0 3)`, `POINT(1 1)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_LineMerge('MULTILINESTRING((0 0,1 1),(1 1,2 2),(3 3,2 2))'::geometry)), ST_AsText(ST_LineMerge('MULTILINESTRING((0 0,1 1),(1 1,2 2),(3 3,2 2))'::geometry, true));`,
					Expected: []sql.Row{{`LINESTRING(0 0,1 1,2 2,3 3)`, `MULTILINESTRING((0 0,1 1,2 2),(3 3,2 2))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_LineMerge('LINESTRING(0 0,1 1)'::geometry)), ST_AsText(ST_LineMerge('POINT(0 0)'::geometry)), ST_AsText(ST_LineMerge('MULTILINESTRING EMPTY'::geometry));`,
					Expected: []sql.Row{{`LINESTRING(0 0,1 1)`, `GEOMETRYCOLLECTION EMPTY`, `MULTILINESTRING EMPTY`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Node('MULTILINESTRING((0 0,2 2),(0 2,2 0))'::geometry)), ST_AsText(ST_Node('LINESTRING Z (0 0 0,2 2 2,0 2 1,2 0 3)'::geometry));`,
					Expected: []sql.Row{{`MULTILINESTRING((0 0,1 1),(0 2,1 1),(1 1,2 2),(1 1,2 0))`, `MULTILINESTRING Z ((0 0 0,1 1 1.5),(1 1 1.5,2 2 2,0 2 1,1 1 1.5),(1 1 1.5,2 0 3))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_SharedPaths('MULTILINESTRING((26 125,26 200,126 200,126 125,26 125),(51 150,101 150,76 175,51 150))'::geometry, 'LINESTRING(151 100,126 156.25,126 125,90 161,76 175)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(MULTILINESTRING((126 156.25,126 125),(101 150,90 161),(90 161,76 175)),MULTILINESTRING EMPTY)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Snap('POLYGON((0 0,10 0,10 10,0 10,0 0))'::geometry, 'POINT(10.5 10.5)'::geometry, 1)), ST_AsText(ST_Snap('LINESTRING(0 0,5 0)'::geometry, 'LINESTRING(0 0.2,5 0.2)'::geometry, 0.5));`,
					Expected: []sql.Row{{`POLYGON((0 0,10 0,10.5 10.5,0 10,0 0))`, `LINESTRING(0 0.2,5 0.2)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_OffsetCurve('LINESTRING(0 0,10 0,10 10)'::geometry, 2)), ST_AsText(ST_OffsetCurve('LINESTRING(0 0,10 0,10 10)'::geometry, -2, 'join=mitre'));`,
					Expected: []sql.Row{{`LINESTRING(0 2,8 2,8 10)`, `LINESTRING(0 -2,12 -2,12 10)`}},
				},
				{
					Query:           `SELECT ST_AsText(ST_OffsetCurve('LINESTRING(0 0,10 0,10 10)'::geometry, 2, 'quad_segs=2 join=round')), ST_AsText(ST_OffsetCurve('LINESTRING(0 0,10 0)'::geometry, 0)), ST_AsText(ST_OffsetCurve('POINT(0 0)'::geometry, 1));`,
					ExpectedErr:     `lwgeom_offsetcurve: input is not linear (type Point)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsText(ST_OffsetCurve('LINESTRING(0 0,10 0,10 10)'::geometry, 2, 'join=bevel mitre_limit=2 endcap=round'));`,
					ExpectedErr:     `Invalid buffer parameter: endcap (accept: 'join', 'mitre_limit', 'miter_limit and 'quad_segs')`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsText(ST_OffsetCurve('LINESTRING(0 0,10 0)'::geometry, 2, 'foo=bar'));`,
					ExpectedErr:     `Invalid buffer parameter: foo (accept: 'join', 'mitre_limit', 'miter_limit and 'quad_segs')`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsText(ST_BuildArea('MULTILINESTRING((0 0,10 0,10 10,0 10,0 0),(2 2,4 2,4 4,2 4,2 2))'::geometry)), ST_AsText(ST_BuildArea('LINESTRING(0 0,1 1)'::geometry));`,
					Expected: []sql.Row{{`POLYGON((0 0,0 10,10 10,10 0,0 0),(2 2,4 2,4 4,2 4,2 2))`, nil}},
				},
				{
					Query:    `SELECT ST_AsText(ST_ConcaveHull('MULTIPOINT((0 0),(10 0),(10 10),(0 10),(5 5),(5 1))'::geometry, 0.5)), ST_AsText(ST_ConcaveHull('MULTIPOINT((0 0),(10 0),(10 10),(0 10),(5 5),(5 1))'::geometry, 1));`,
					Expected: []sql.Row{{`POLYGON((5 5,0 10,10 10,10 0,5 1,0 0,5 5))`, `POLYGON((0 0,0 10,10 10,10 0,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_ConcaveHull('POLYGON((0 0,10 0,10 10,6 10,6 2,4 2,4 10,0 10,0 0))'::geometry, 0.1, true)), ST_AsText(ST_SimplifyPolygonHull('POLYGON((0 0,10 0,10 10,6 10,6 2,4 2,4 10,0 10,0 0))'::geometry, 0.5)), ST_AsText(ST_SimplifyPolygonHull('POLYGON((0 0,10 0,10 10,6 10,6 2,4 2,4 10,0 10,0 0))'::geometry, 0.5, false));`,
					Expected: []sql.Row{{`POLYGON((6 10,10 10,10 0,0 0,0 10,4 10,6 10))`, `POLYGON((0 0,0 10,10 10,10 0,0 0))`, `POLYGON((0 0,4 2,0 10,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_DelaunayTriangles('MULTIPOINT((0 0),(10 0),(5 5),(0 10))'::geometry)), ST_AsText(ST_DelaunayTriangles('MULTIPOINT((0 0),(10 0),(5 5),(0 10))'::geometry, 0, 1)), ST_AsText(ST_DelaunayTriangles('MULTIPOINT((0 0),(10 0),(5 5),(0 10))'::geometry, 0, 2));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(POLYGON((0 10,0 0,5 5,0 10)),POLYGON((5 5,0 0,10 0,5 5)))`, `MULTILINESTRING((5 5,10 0),(0 10,5 5),(0 0,0 10),(0 0,10 0),(0 0,5 5))`, `TIN(((0 10,0 0,5 5,0 10)),((5 5,0 0,10 0,5 5)))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_DelaunayTriangles('MULTIPOINT Z ((0 0 1),(10 0 2),(5 5 3))'::geometry, 0, 2));`,
					Expected: []sql.Row{{`TIN Z (((0 0 1,10 0 2,5 5 3,0 0 1)))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_VoronoiPolygons('MULTIPOINT((0 0),(10 0),(5 5))'::geometry)), ST_AsText(ST_VoronoiLines('MULTIPOINT((0 0),(10 0),(5 5))'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(POLYGON((-10 -10,-10 15,5 0,5 -10,-10 -10)),POLYGON((20 15,5 0,-10 15,20 15)),POLYGON((20 -10,5 -10,5 0,20 15,20 -10)))`, `MULTILINESTRING((-10 15,5 0),(5 0,5 -10),(20 15,5 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_VoronoiPolygons('MULTIPOINT((0 0),(10 0),(5 5))'::geometry, 0, 'POLYGON((-100 -100,100 -100,100 100,-100 100,-100 -100))'::geometry)), ST_AsText(ST_VoronoiPolygons('POINT(1 1)'::geometry)), ST_AsText(ST_VoronoiPolygons(NULL));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(POLYGON((-100 -100,-100 100,-95 100,5 0,5 -100,-100 -100)),POLYGON((100 100,100 95,5 0,-95 100,100 100)),POLYGON((100 -100,5 -100,5 0,100 95,100 -100)))`, `GEOMETRYCOLLECTION EMPTY`, nil}},
				},
				{
					Query:    `SELECT ST_MinimumClearance('POLYGON((0 0,1 0,1 1,0.5 3.2e-4,0 0))'::geometry), ST_AsText(ST_MinimumClearanceLine('POLYGON((0 0,1 0,1 1,0.5 3.2e-4,0 0))'::geometry)), ST_MinimumClearance('POINT(1 1)'::geometry)::text, ST_AsText(ST_MinimumClearanceLine('POINT(1 1)'::geometry));`,
					Expected: []sql.Row{{0.00032, `LINESTRING(0.5 0.00032,0.5 0)`, `Infinity`, `LINESTRING EMPTY`}},
				},
				{
					Query:    `SELECT ST_FrechetDistance('LINESTRING(0 0,100 0)'::geometry, 'LINESTRING(0 0,50 50,100 0)'::geometry), ST_FrechetDistance('LINESTRING(0 0,100 0)'::geometry, 'LINESTRING(0 0,50 50,100 0)'::geometry, 0.5);`,
					Expected: []sql.Row{{70.71067811865476, 50.0}},
				},
				{
					Query:    `SELECT ST_HausdorffDistance('LINESTRING(0 0,2 0)'::geometry, 'LINESTRING(0 1,1 2,2 1)'::geometry), ST_HausdorffDistance('LINESTRING(0 0,2 0)'::geometry, 'LINESTRING(0 1,1 2,2 1)'::geometry, 0.5);`,
					Expected: []sql.Row{{2.0, 2.0}},
				},
				{
					Query:    `SELECT ST_HausdorffDistance('SRID=4326;POINT(0 0)'::geometry, 'POINT(1 1)'::geometry);`,
					Expected: []sql.Row{{1.4142135623730951}},
				},
				{
					Query:    `SELECT ST_AsText(ST_ClipByBox2D('POLYGON((0 0,10 0,10 10,0 10,0 0))'::geometry, ST_MakeEnvelope(5, 5, 15, 15))), ST_AsText(ST_ClipByBox2D('LINESTRING(0 0,10 10)'::geometry, 'BOX(2 2,4 4)'::box2d)), ST_AsText(ST_ClipByBox2D('POINT(0 0)'::geometry, 'BOX(2 2,4 4)'::box2d));`,
					Expected: []sql.Row{{`POLYGON((5 5,5 10,10 10,10 5,5 5))`, `LINESTRING(2 2,4 4)`, `POINT EMPTY`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_ClipByBox2D('SRID=4326;POLYGON((0 0,10 0,10 10,0 10,0 0))'::geometry, 'BOX(-1 -1,20 20)'::box2d)), ST_AsText(ST_ClipByBox2D('POLYGON EMPTY'::geometry, 'BOX(2 2,4 4)'::box2d));`,
					Expected: []sql.Row{{`SRID=4326;POLYGON((0 0,10 0,10 10,0 10,0 0))`, `POLYGON EMPTY`}},
				},
				{
					Query:    `SELECT ST_AsText(center), ST_AsText(nearest), radius FROM ST_MaximumInscribedCircle('POLYGON((0 0,10 0,10 10,0 10,0 0))'::geometry);`,
					Expected: []sql.Row{{`POINT(5 5)`, `POINT(5 0)`, 5.0}},
				},
				{
					Query:    `SELECT ST_AsText(center), ST_AsText(nearest), radius FROM ST_LargestEmptyCircle('MULTIPOINT((0 0),(10 0),(10 10),(0 10))'::geometry);`,
					Expected: []sql.Row{{`POINT(5 5)`, `POINT(0 0)`, 7.0710678118654755}},
				},
				{
					Query:           `SELECT ST_AsText(center), ST_AsText(nearest), radius FROM ST_LargestEmptyCircle('MULTIPOINT((0 0),(10 0),(10 10),(0 10))'::geometry, 0.1, 'POLYGON((0 0,4 0,4 4,0 4,0 0))'::geometry);`,
					ExpectedErr:     `Error calculating GEOSLargestEmptyCircle.`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsText(center), radius FROM ST_MinimumBoundingRadius('MULTIPOINT((0 0),(10 0),(10 10))'::geometry);`,
					Expected: []sql.Row{{`POINT(5 5)`, 7.0710678118654755}},
				},
				{
					Query:    `SELECT ST_NPoints(ST_MinimumBoundingCircle('MULTIPOINT((0 0),(10 0),(10 10))'::geometry)), ST_NPoints(ST_MinimumBoundingCircle('MULTIPOINT((0 0),(10 0),(10 10))'::geometry, 2)), ST_AsText(ST_MinimumBoundingCircle('POINT(1 1)'::geometry)), ST_AsText(ST_MinimumBoundingCircle('POINT EMPTY'::geometry));`,
					Expected: []sql.Row{{193, 9, `POINT(1 1)`, `POINT EMPTY`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_ReducePrecision('POINT(1.412 19.323)'::geometry, 0.1)), ST_AsText(ST_ReducePrecision('POLYGON((0 0,10 0,10 0.4,0 0.6,0 0))'::geometry, 1)), ST_AsText(ST_ReducePrecision('POINT(1.412 19.323)'::geometry, 0));`,
					Expected: []sql.Row{{`POINT(1.4 19.3)`, `POLYGON((0 0,0 1,10 0,0 0))`, `POINT(1.412 19.323)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_ReducePrecision('POINT(1.412 19.323)'::geometry, -1));`,
					Expected: []sql.Row{{`POINT(1 19)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Polygonize(ARRAY['LINESTRING(0 0,10 0,10 10)'::geometry, 'LINESTRING(10 10,0 10,0 0)'::geometry])), ST_AsText(ST_Polygonize(g)) FROM (VALUES ('LINESTRING(0 0,10 0,10 10,0 0)'::geometry)) t(g);`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(POLYGON((10 10,10 0,0 0,0 10,10 10)))`, `GEOMETRYCOLLECTION(POLYGON((0 0,10 10,10 0,0 0)))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_TriangulatePolygon('POLYGON((0 0,10 0,10 10,0 10,0 0))'::geometry)), ST_AsText(ST_TriangulatePolygon('POINT(0 0)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(POLYGON((0 0,0 10,10 10,0 0)),POLYGON((10 10,10 0,0 0,10 10)))`, `GEOMETRYCOLLECTION EMPTY`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_CleanGeometry('POLYGON((0 0,2 2,2 0,0 2,0 0))'::geometry)), ST_AsText(ST_BdPolyFromText('MULTILINESTRING((0 0,10 0,10 10,0 10,0 0))', 4326)), ST_AsEWKT(ST_BdMPolyFromText('MULTILINESTRING((0 0,10 0,10 10,0 10,0 0),(20 20,30 20,30 30,20 20))', 4326));`,
					Expected: []sql.Row{{`MULTIPOLYGON(((0 2,1 1,0 0,0 2)),((2 0,1 1,2 2,2 0)))`, `POLYGON((0 0,0 10,10 10,10 0,0 0))`, `SRID=4326;MULTIPOLYGON(((0 10,10 10,10 0,0 0,0 10)),((30 30,30 20,20 20,30 30)))`}},
				},
				{
					Query:           `SELECT ST_AsText(ST_BdPolyFromText('MULTILINESTRING((0 0,10 0,10 10,0 10,0 0),(20 20,30 20,30 30,20 20))', 4326));`,
					ExpectedErr:     `Input returns more then a single polygon, try using BdMPolyFromText instead`,
					ExpectedErrCode: `P0001`,
				},
				{
					Query:           `SELECT ST_AsText(ST_BdPolyFromText('POINT(0 0)', 4326));`,
					ExpectedErr:     `Input is not a MultiLinestring`,
					ExpectedErrCode: `P0001`,
				},
			},
		},
		{
			Name: `postgis split`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsText(ST_Split('LINESTRING(0 0,10 0)'::geometry, 'POINT(5 1)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(LINESTRING(0 0,10 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('LINESTRING(0 0,10 0)'::geometry, 'POINT(0 0)'::geometry)), ST_AsText(ST_Split('LINESTRING(0 0,10 0)'::geometry, 'POINT(10 0)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(LINESTRING(0 0,10 0))`, `GEOMETRYCOLLECTION(LINESTRING(0 0,10 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('LINESTRING(0 0,10 0,10 10)'::geometry, 'POINT(10 0)'::geometry)), ST_AsText(ST_Split('LINESTRING(0 0,10 0,10 10)'::geometry, 'POINT(10 5)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(LINESTRING(0 0,10 0),LINESTRING(10 0,10 10))`, `GEOMETRYCOLLECTION(LINESTRING(0 0,10 0,10 5),LINESTRING(10 5,10 10))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('LINESTRING(0 0,10 0)'::geometry, 'MULTIPOINT((7 0),(2 0))'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(LINESTRING(7 0,10 0),LINESTRING(2 0,7 0),LINESTRING(0 0,2 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('LINESTRING Z (0 0 0,10 0 10)'::geometry, 'POINT(5 0)'::geometry)), ST_AsText(ST_Split('LINESTRING M (0 0 0,10 0 10)'::geometry, 'POINT(4 0)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION Z (LINESTRING Z (0 0 0,5 0 5),LINESTRING Z (5 0 5,10 0 10))`, `GEOMETRYCOLLECTION M (LINESTRING M (0 0 0,4 0 4),LINESTRING M (4 0 4,10 0 10))`}},
				},
				{
					Query:           `SELECT ST_AsText(ST_Split('LINESTRING(0 0,10 0)'::geometry, 'LINESTRING(2 0,5 0)'::geometry));`,
					ExpectedErr:     `Splitter line has linear intersection with input`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('LINESTRING(0 0,10 0)'::geometry, 'MULTILINESTRING((2 -1,2 1),(8 -1,8 1))'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(LINESTRING(0 0,2 0),LINESTRING(2 0,8 0),LINESTRING(8 0,10 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('LINESTRING(0 0,10 0)'::geometry, 'POLYGON((2 -1,8 -1,8 1,2 1,2 -1))'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(LINESTRING(0 0,2 0),LINESTRING(2 0,8 0),LINESTRING(8 0,10 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('LINESTRING(0 0,10 0)'::geometry, 'MULTIPOLYGON(((2 -1,8 -1,8 1,2 1,2 -1)))'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(LINESTRING(0 0,2 0),LINESTRING(2 0,8 0),LINESTRING(8 0,10 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('MULTILINESTRING((0 0,10 0),(0 5,10 5))'::geometry, 'LINESTRING(5 -1,5 6)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(LINESTRING(0 0,5 0),LINESTRING(5 0,10 0),LINESTRING(0 5,5 5),LINESTRING(5 5,10 5))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('POLYGON((0 0,10 0,10 10,0 10,0 0),(4 4,6 4,6 6,4 6,4 4))'::geometry, 'LINESTRING(5 -1,5 11)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(POLYGON((5 0,0 0,0 10,5 10,5 6,4 6,4 4,5 4,5 0)),POLYGON((5 10,10 10,10 0,5 0,5 4,6 4,6 6,5 6,5 10)))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('POLYGON((0 0,10 0,10 10,0 10,0 0))'::geometry, 'LINESTRING(5 -1,5 5)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(POLYGON((5 0,0 0,0 10,10 10,10 0,5 0)))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('POLYGON((0 0,10 0,10 10,0 10,0 0))'::geometry, 'MULTILINESTRING((5 -1,5 11),(-1 5,11 5))'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(POLYGON((5 0,0 0,0 5,5 5,5 0)),POLYGON((10 5,10 0,5 0,5 5,10 5)),POLYGON((5 10,10 10,10 5,5 5,5 10)),POLYGON((0 5,0 10,5 10,5 5,0 5)))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('MULTIPOLYGON(((0 0,10 0,10 10,0 10,0 0)),((20 0,30 0,30 10,20 10,20 0)))'::geometry, 'LINESTRING(5 -1,5 11)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(POLYGON((5 0,0 0,0 10,5 10,5 0)),POLYGON((5 10,10 10,10 0,5 0,5 10)),POLYGON((20 0,20 10,30 10,30 0,20 0)))`}},
				},
				{
					Query:           `SELECT ST_AsText(ST_Split('POLYGON((0 0,10 0,10 10,0 10,0 0))'::geometry, 'POINT(5 5)'::geometry));`,
					ExpectedErr:     `Splitting a Polygon by a Point is unsupported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('GEOMETRYCOLLECTION(LINESTRING(0 0,10 0),POLYGON((0 0,10 0,10 10,0 10,0 0)))'::geometry, 'LINESTRING(5 -1,5 11)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(LINESTRING(0 0,5 0),LINESTRING(5 0,10 0),POLYGON((5 0,0 0,0 10,5 10,5 0)),POLYGON((5 10,10 10,10 0,5 0,5 10)))`}},
				},
				{
					Query:           `SELECT ST_AsText(ST_Split('LINESTRING(0 0,10 0)'::geometry, 'GEOMETRYCOLLECTION(POINT(5 0))'::geometry));`,
					ExpectedErr:     `Splitting a Line by a GeometryCollection is unsupported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsEWKT(ST_Split('SRID=4326;LINESTRING(0 0,10 0)'::geometry, 'POINT(5 0)'::geometry));`,
					ExpectedErr:     `ST_Split: Operation on mixed SRID geometries (LineString, 4326) != (Point, 0)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsText(ST_Split('CIRCULARSTRING(0 0,1 1,2 0)'::geometry, 'POINT(1 1)'::geometry));`,
					ExpectedErr:     `Splitting of CircularString geometries is unsupported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('LINESTRING(0 0,10 0)'::geometry, 'POINT(3.3333333333333335 0)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(LINESTRING(0 0,3.333333333333334 0),LINESTRING(3.333333333333334 0,10 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('LINESTRING(0 0,10 10)'::geometry, 'POINT(3.3333333333333335 3.3333333333333335)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(LINESTRING(0 0,3.333333333333334 3.333333333333334),LINESTRING(3.333333333333334 3.333333333333334,10 10))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('LINESTRING(0 0,10 0,10 10,0 10)'::geometry, 'LINESTRING(5 -1,5 11)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(LINESTRING(0 0,5 0),LINESTRING(5 0,10 0,10 10,5 10),LINESTRING(5 10,0 10))`}},
				},
				{
					Query:           `SELECT ST_AsText(ST_Split('POINT(0 0)'::geometry, 'POINT(0 0)'::geometry));`,
					ExpectedErr:     `Splitting of Point geometries is unsupported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsText(ST_Split('POLYGON((0 0,10 0,10 10,0 10,0 0))'::geometry, 'POLYGON((5 -1,6 -1,6 11,5 11,5 -1))'::geometry));`,
					ExpectedErr:     `Splitting a Polygon by a Polygon is unsupported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('LINESTRING(0 0,10 0)'::geometry, 'MULTIPOINT((7 0),(2 0),(5 0),(20 0))'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(LINESTRING(7 0,10 0),LINESTRING(5 0,7 0),LINESTRING(0 0,2 0),LINESTRING(2 0,5 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('LINESTRING(0 0,10 0,10 10,0 10,0 0)'::geometry, 'POINT(0 0)'::geometry)), ST_AsText(ST_Split('LINESTRING(0 0,10 0,10 10,0 10,0 0)'::geometry, 'POINT(10 10)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(LINESTRING(0 0,10 0,10 10,0 10,0 0))`, `GEOMETRYCOLLECTION(LINESTRING(0 0,10 0,10 10),LINESTRING(10 10,0 10,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('LINESTRING Z (0 0 0,10 0 10)'::geometry, 'LINESTRING Z (5 -1 0,5 1 0)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION Z (LINESTRING Z (0 0 0,5 0 2.5),LINESTRING Z (5 0 2.5,10 0 10))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Split('POLYGON Z ((0 0 1,10 0 1,10 10 1,0 10 1,0 0 1))'::geometry, 'LINESTRING(5 -1,5 11)'::geometry));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION Z (POLYGON Z ((5 0 1,0 0 1,0 10 1,5 10 1,5 0 1)),POLYGON Z ((5 10 1,10 10 1,10 0 1,5 0 1,5 10 1)))`}},
				},
			},
		},
		{
			Name: `postgis subdivide`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsText(g) FROM ST_Subdivide('POLYGON((0 0,10 0,10 10,0 10,0 0))'::geometry, 5) g;`,
					Expected: []sql.Row{{`POLYGON((0 0,10 0,10 10,0 10,0 0))`}},
				},
				{
					Query: `SELECT ST_AsText(g, 12) FROM ST_Subdivide(ST_Buffer('POINT(0 0)'::geometry, 10, 4), 8) g;`,
					Expected: []sql.Row{
						{`POLYGON((-3.826834323651 -9.238795325113,-7.071067811865 -7.071067811865,-9.238795325113 -3.826834323651,-10 -1.224646799147e-15,-10 0,6.123233995737e-16 0,6.123233995737e-16 -10,-3.826834323651 -9.238795325113))`},
						{`POLYGON((7.071067811865 -7.071067811865,3.826834323651 -9.238795325113,6.123233995737e-16 -10,6.123233995737e-16 0,10 0,9.238795325113 -3.826834323651,7.071067811865 -7.071067811865))`},
						{`POLYGON((-7.071067811865 7.071067811865,-3.826834323651 9.238795325113,-1.836970198721e-15 10,-1.836970198721e-15 0,-10 0,-9.238795325113 3.826834323651,-7.071067811865 7.071067811865))`},
						{`POLYGON((3.826834323651 9.238795325113,7.071067811865 7.071067811865,9.238795325113 3.826834323651,10 0,-1.836970198721e-15 0,-1.836970198721e-15 10,3.826834323651 9.238795325113))`},
					},
				},
				{
					Query: `SELECT ST_AsText(g) FROM ST_Subdivide('LINESTRING(0 0,1 1,2 2,3 3,4 4,5 5,6 6)'::geometry, 5) g;`,
					Expected: []sql.Row{
						{`LINESTRING(0 0,3 3)`},
						{`LINESTRING(3 3,6 6)`},
					},
				},
				{
					Query: `SELECT ST_AsText(g) FROM ST_Subdivide('LINESTRING(0 0,1 3,2 0,3 3,4 0,5 3,6 0,7 3,8 0,9 3,10 0)'::geometry, 5) g;`,
					Expected: []sql.Row{
						{`LINESTRING(0 0,1 3)`},
						{`LINESTRING(1 3,2 0)`},
						{`LINESTRING(2 0,3 3)`},
						{`LINESTRING(3 3,4 0)`},
						{`LINESTRING(4 0,5 3)`},
						{`LINESTRING(5 3,6 0)`},
						{`LINESTRING(6 0,7 3)`},
						{`LINESTRING(7 3,8 0)`},
						{`LINESTRING(8 0,9 3)`},
						{`LINESTRING(9 3,10 0)`},
					},
				},
				{
					Query: `SELECT ST_AsText(g) FROM ST_Subdivide('POLYGON((0 0,20 0,20 10,0 10,0 0),(2 2,4 2,4 4,2 4,2 2),(12 2,14 2,14 4,12 4,12 2))'::geometry, 6) g;`,
					Expected: []sql.Row{
						{`POLYGON((4 2,4 0,0 0,0 2,4 2))`},
						{`POLYGON((0 4,2 4,2 2,0 2,0 4))`},
						{`POLYGON((4 10,4 4,0 4,0 10,4 10))`},
						{`POLYGON((12 10,12 0,4 0,4 10,12 10))`},
						{`POLYGON((14 2,14 0,12 0,12 2,14 2))`},
						{`POLYGON((14 4,20 4,20 0,14 0,14 4))`},
						{`POLYGON((12 4,12 10,20 10,20 4,12 4))`},
					},
				},
				{
					Query: `SELECT ST_AsText(g) FROM ST_Subdivide('MULTIPOINT((0 0),(1 1),(2 2),(3 3),(4 4),(5 5),(6 6))'::geometry, 5) g;`,
					Expected: []sql.Row{
						{`MULTIPOINT((0 0),(1 1),(2 2),(3 3))`},
						{`MULTIPOINT((3 3),(4 4),(5 5),(6 6))`},
					},
				},
				{
					Query:    `SELECT ST_AsText(g) FROM ST_Subdivide('POINT(1 1)'::geometry, 5) g;`,
					Expected: []sql.Row{{`POINT(1 1)`}},
				},
				{
					Query:    `SELECT ST_AsText(g) FROM ST_Subdivide('POLYGON EMPTY'::geometry, 5) g;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT ST_AsText(g) FROM ST_Subdivide('GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 0,1 1,2 2,3 3,4 4,5 5,6 6),POLYGON((0 0,10 0,10 10,0 10,0 0)))'::geometry, 5) g;`,
					Expected: []sql.Row{{`POLYGON((0 0,10 0,10 10,0 10,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(g) FROM ST_Subdivide('POLYGON((0 0,10 0,10 10,0 10,0 0))'::geometry, 5, 1) g;`,
					Expected: []sql.Row{{`POLYGON((0 0,10 0,10 10,0 10,0 0))`}},
				},
				{
					Query: `SELECT ST_AsText(g) FROM ST_Subdivide(ST_Buffer('POINT(0 0)'::geometry, 10, 4), 8, 0.5) g;`,
					Expected: []sql.Row{
						{`POLYGON((-4 -9,-7 -7,-9 -4,-10 0,0 0,0 -10,-4 -9))`},
						{`POLYGON((7 -7,4 -9,0 -10,0 0,10 0,9 -4,7 -7))`},
						{`POLYGON((-7 7,-4 9,0 10,0 0,-10 0,-9 4,-7 7))`},
						{`POLYGON((4 9,7 7,9 4,10 0,0 0,0 10,4 9))`},
					},
				},
				{
					Query: `SELECT ST_AsText(g) FROM ST_Subdivide('LINESTRING(0 0,0 1,0 2,0 3,0 4,0 5,0 6)'::geometry, 5) g;`,
					Expected: []sql.Row{
						{`LINESTRING(0 0,0 3)`},
						{`LINESTRING(0 3,0 6)`},
					},
				},
				{
					Query: `SELECT ST_AsEWKT(g) FROM ST_Subdivide('SRID=4326;LINESTRING Z (0 0 0,1 1 1,2 2 2,3 3 3,4 4 4,5 5 5,6 6 6)'::geometry, 5) g;`,
					Expected: []sql.Row{
						{`SRID=4326;LINESTRING(0 0 0,3 3 3)`},
						{`SRID=4326;LINESTRING(3 3 3,6 6 6)`},
					},
				},
				{
					Query: `SELECT ST_AsText(g) FROM ST_Subdivide('MULTIPOLYGON(((0 0,1 0,1 1,0 1,0 0)),((5 5,6 5,6 6,5 6,5 5)))'::geometry, 5) g;`,
					Expected: []sql.Row{
						{`POLYGON((0 0,1 0,1 1,0 1,0 0))`},
						{`POLYGON((5 5,6 5,6 6,5 6,5 5))`},
					},
				},
				{
					Query:           `SELECT ST_AsText(g) FROM ST_Subdivide('TIN(((0 0,1 0,0 1,0 0)))'::geometry, 5) g;`,
					ExpectedErr:     `lwgeom_subdivide_recursive: unsupported geometry type 'Tin'`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsText(g) FROM ST_Subdivide('CIRCULARSTRING(0 0,1 1,2 0)'::geometry, 5) g;`,
					Expected: []sql.Row{{`CIRCULARSTRING(0 0,1 1,2 0)`}},
				},
				{
					Query: `SELECT ST_AsText(g) FROM ST_Subdivide('POLYGON((0 0,7 0,7 1,1 1,1 9,7 9,7 10,0 10,0 0))'::geometry, 5) g;`,
					Expected: []sql.Row{
						{`POLYGON((0 0,0 1,7 1,7 0,0 0))`},
						{`POLYGON((1 1,0 1,0 9,1 9,1 1))`},
						{`POLYGON((0 10,7 10,7 9,0 9,0 10))`},
					},
				},
				{
					Query:    `SELECT count(*), sum(ST_NPoints(g)), sum(ST_Area(g)) FROM ST_Subdivide('POLYGON((0 0,40 0,40 3,30 3,30 6,40 6,40 40,0 40,0 30,10 30,10 20,0 20,0 0))'::geometry, 6) g;`,
					Expected: []sql.Row{{5, 25, 1470.0}},
				},
			},
		},
		{
			Name: `postgis spatial reference systems in PROJ`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT count(*) FROM postgis_srs_codes('EPSG');`,
					Expected: []sql.Row{{6658}},
				},
				{
					Query:    `SELECT count(*) FROM postgis_srs_codes('ESRI');`,
					Expected: []sql.Row{{1380}},
				},
				{
					Query:    `SELECT count(*) FROM postgis_srs_codes('IAU_2015');`,
					Expected: []sql.Row{{2189}},
				},
				{
					Query:    `SELECT count(*) FROM postgis_srs_codes('NOPE');`,
					Expected: []sql.Row{{0}},
				},
				{
					Query: `SELECT postgis_srs_codes('EPSG') LIMIT 5;`,
					Expected: []sql.Row{
						{`10160`},
						{`10183`},
						{`10188`},
						{`10194`},
						{`10199`},
					},
				},
				{
					Query:    `SELECT auth_name, auth_srid, srname, srtext, proj4text, ST_AsEWKT(point_sw), ST_AsEWKT(point_ne) FROM postgis_srs('EPSG', '4326');`,
					Expected: []sql.Row{{`EPSG`, `4326`, `WGS 84`, `GEOGCS["WGS 84",DATUM["WGS_1984",SPHEROID["WGS 84",6378137,298.257223563,AUTHORITY["EPSG","7030"]],AUTHORITY["EPSG","6326"]],PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]],UNIT["degree",0.0174532925199433,AUTHORITY["EPSG","9122"]],AUTHORITY["EPSG","4326"]]`, `+proj=longlat +datum=WGS84 +no_defs +type=crs`, `SRID=4326;POINT(-180 -90)`, `SRID=4326;POINT(180 90)`}},
				},
				{
					Query:    `SELECT auth_name, auth_srid, srname, proj4text, ST_AsEWKT(point_sw), ST_AsEWKT(point_ne) FROM postgis_srs('ESRI', '102003');`,
					Expected: []sql.Row{{`ESRI`, `102003`, `USA_Contiguous_Albers_Equal_Area_Conic`, `+proj=aea +lat_0=37.5 +lon_0=-96 +lat_1=29.5 +lat_2=45.5 +x_0=0 +y_0=0 +datum=NAD83 +units=m +no_defs +type=crs`, `SRID=4326;POINT(-124.79 24.41)`, `SRID=4326;POINT(-66.91 49.38)`}},
				},
				{
					Query:    `SELECT count(*) FROM postgis_srs('EPSG', '999999');`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT count(*) FROM postgis_srs('NOPE', '1');`,
					Expected: []sql.Row{{1}},
				},
				{
					Query: `SELECT auth_name, auth_srid, srname, ST_AsText(point_sw), ST_AsText(point_ne) FROM postgis_srs_search('SRID=4326;POLYGON((-71.2 42.3,-71 42.3,-71 42.4,-71.2 42.4,-71.2 42.3))') ORDER BY auth_srid;`,
					Expected: []sql.Row{
						{`EPSG`, `10598`, `WGS 84 / GLANCE North America`, `POINT(-180 7.98)`, `POINT(180 90)`},
						{`EPSG`, `10820`, `WGS 84 / Agriculture Canada Albers`, `POINT(-139.04 41.67)`, `POINT(-52.54 60.01)`},
						{`EPSG`, `20050`, `NAD83(2011) / Amtrak NECCS21 (ft)`, `POINT(-78.51 38.64)`, `POINT(-70.48 42.74)`},
						{`EPSG`, `22219`, `NAD83(CSRS)v2 / UTM zone 19N`, `POINT(-72 40.8)`, `POINT(-66 84)`},
						{`EPSG`, `22319`, `NAD83(CSRS)v3 / UTM zone 19N`, `POINT(-72 40.8)`, `POINT(-66 84)`},
						{`EPSG`, `22419`, `NAD83(CSRS)v4 / UTM zone 19N`, `POINT(-72 40.8)`, `POINT(-66 84)`},
						{`EPSG`, `2249`, `NAD83 / Massachusetts Mainland (ftUS)`, `POINT(-73.5 41.46)`, `POINT(-69.86 42.89)`},
						{`EPSG`, `22619`, `NAD83(CSRS)v6 / UTM zone 19N`, `POINT(-72 40.8)`, `POINT(-66 84)`},
						{`EPSG`, `22719`, `NAD83(CSRS)v7 / UTM zone 19N`, `POINT(-72 40.8)`, `POINT(-66 84)`},
						{`EPSG`, `22819`, `NAD83(CSRS)v8 / UTM zone 19N`, `POINT(-72 40.8)`, `POINT(-66 84)`},
						{`EPSG`, `26719`, `NAD27 / UTM zone 19N`, `POINT(-72 33.61)`, `POINT(-66 83.17)`},
						{`EPSG`, `26786`, `NAD27 / Massachusetts Mainland`, `POINT(-73.5 41.46)`, `POINT(-69.86 42.89)`},
						{`EPSG`, `26919`, `NAD83 / UTM zone 19N`, `POINT(-72 14.92)`, `POINT(-66 84)`},
						{`EPSG`, `26986`, `NAD83 / Massachusetts Mainland`, `POINT(-73.5 41.46)`, `POINT(-69.86 42.89)`},
						{`EPSG`, `27705`, `WGS 84 / Equi7 North America`, `POINT(-180 7.98)`, `POINT(180 90)`},
						{`EPSG`, `2805`, `NAD83(HARN) / Massachusetts Mainland`, `POINT(-73.5 41.46)`, `POINT(-69.86 42.89)`},
						{`EPSG`, `2894`, `NAD83(HARN) / Massachusetts Mainland (ftUS)`, `POINT(-73.5 41.46)`, `POINT(-69.86 42.89)`},
						{`EPSG`, `2960`, `NAD83(CSRS) / UTM zone 19N`, `POINT(-72 40.8)`, `POINT(-66 84)`},
						{`EPSG`, `3175`, `NAD83 / Great Lakes and St Lawrence Albers`, `POINT(-93.17 40.99)`, `POINT(-54.75 52.22)`},
						{`EPSG`, `32219`, `WGS 72 / UTM zone 19N`, `POINT(-72 0)`, `POINT(-66 84)`},
						{`EPSG`, `32419`, `WGS 72BE / UTM zone 19N`, `POINT(-72 0)`, `POINT(-66 84)`},
						{`EPSG`, `32600`, `WGS 84 / UTM grid system (northern hemisphere)`, `POINT(-180 0)`, `POINT(180 84)`},
						{`EPSG`, `32619`, `WGS 84 / UTM zone 19N`, `POINT(-72 0)`, `POINT(-66 84)`},
						{`EPSG`, `3347`, `NAD83 / Statistics Canada Lambert`, `POINT(-141.01 38.21)`, `POINT(-40.73 86.46)`},
						{`EPSG`, `3348`, `NAD83(CSRS) / Statistics Canada Lambert`, `POINT(-141.01 38.21)`, `POINT(-40.73 86.46)`},
						{`EPSG`, `3395`, `WGS 84 / World Mercator`, `POINT(-180 -80)`, `POINT(180 84)`},
						{`EPSG`, `3408`, `NSIDC EASE-Grid North`, `POINT(-180 0)`, `POINT(180 90)`},
						{`EPSG`, `3410`, `NSIDC EASE-Grid Global`, `POINT(-180 -86)`, `POINT(180 86)`},
						{`EPSG`, `3585`, `NAD83(NSRS2007) / Massachusetts Mainland`, `POINT(-73.5 41.46)`, `POINT(-69.86 42.89)`},
						{`EPSG`, `3586`, `NAD83(NSRS2007) / Massachusetts Mainland (ftUS)`, `POINT(-73.5 41.46)`, `POINT(-69.86 42.89)`},
						{`EPSG`, `3726`, `NAD83(NSRS2007) / UTM zone 19N`, `POINT(-72 33.61)`, `POINT(-65.99 47.47)`},
						{`EPSG`, `3749`, `NAD83(HARN) / UTM zone 19N`, `POINT(-72 40.96)`, `POINT(-66.91 47.47)`},
						{`EPSG`, `3832`, `WGS 84 / PDC Mercator`, `POINT(98.69 -60)`, `POINT(-68 66.67)`},
						{`EPSG`, `3857`, `WGS 84 / Pseudo-Mercator`, `POINT(-180 -85.06)`, `POINT(180 85.06)`},
						{`EPSG`, `3978`, `NAD83 / Canada Atlas Lambert`, `POINT(-141.01 38.21)`, `POINT(-40.73 86.46)`},
						{`EPSG`, `3979`, `NAD83(CSRS) / Canada Atlas Lambert`, `POINT(-141.01 38.21)`, `POINT(-40.73 86.46)`},
						{`EPSG`, `4087`, `WGS 84 / World Equidistant Cylindrical`, `POINT(-180 -90)`, `POINT(180 90)`},
						{`EPSG`, `4419`, `NAD27 / BLM 19N (ftUS)`, `POINT(-72 33.61)`, `POINT(-65.99 47.47)`},
						{`EPSG`, `4439`, `NAD83 / BLM 19N (ftUS)`, `POINT(-72 33.61)`, `POINT(-65.99 47.47)`},
						{`EPSG`, `5069`, `NAD27 / Conus Albers`, `POINT(-124.79 24.41)`, `POINT(-66.91 49.38)`},
						{`EPSG`, `5070`, `NAD83 / Conus Albers`, `POINT(-124.79 24.41)`, `POINT(-66.91 49.38)`},
						{`EPSG`, `5071`, `NAD83(HARN) / Conus Albers`, `POINT(-124.79 24.41)`, `POINT(-66.91 49.38)`},
						{`EPSG`, `5072`, `NAD83(NSRS2007) / Conus Albers`, `POINT(-124.79 24.41)`, `POINT(-66.91 49.38)`},
						{`EPSG`, `6348`, `NAD83(2011) / UTM zone 19N`, `POINT(-72 33.61)`, `POINT(-65.99 47.47)`},
						{`EPSG`, `6350`, `NAD83(2011) / Conus Albers`, `POINT(-124.79 24.41)`, `POINT(-66.91 49.38)`},
						{`EPSG`, `6491`, `NAD83(2011) / Massachusetts Mainland`, `POINT(-73.5 41.46)`, `POINT(-69.86 42.89)`},
						{`EPSG`, `6492`, `NAD83(2011) / Massachusetts Mainland (ftUS)`, `POINT(-73.5 41.46)`, `POINT(-69.86 42.89)`},
						{`EPSG`, `6931`, `WGS 84 / NSIDC EASE-Grid 2.0 North`, `POINT(-180 0)`, `POINT(180 90)`},
						{`EPSG`, `6933`, `WGS 84 / NSIDC EASE-Grid 2.0 Global`, `POINT(-180 -86)`, `POINT(180 86)`},
						{`EPSG`, `8857`, `WGS 84 / Equal Earth Greenwich`, `POINT(-180 -90)`, `POINT(180 90)`},
						{`EPSG`, `8858`, `WGS 84 / Equal Earth Americas`, `POINT(90.01 -90)`, `POINT(89.99 90)`},
						{`EPSG`, `8859`, `WGS 84 / Equal Earth Asia-Pacific`, `POINT(-29.99 -90)`, `POINT(-30.01 90)`},
						{`EPSG`, `9311`, `NAD27 / US National Atlas Equal Area`, `POINT(167.65 15.56)`, `POINT(-65.69 74.71)`},
					},
				},
				{
					Query:    `SELECT count(*) FROM postgis_srs_search('SRID=4326;POINT(2 49)');`,
					Expected: []sql.Row{{45}},
				},
				{
					Query: `SELECT auth_name, auth_srid FROM postgis_srs_search('SRID=4326;POINT(2 49)', 'ESRI') ORDER BY auth_srid;`,
					Expected: []sql.Row{
						{`ESRI`, `102013`},
						{`ESRI`, `102014`},
						{`ESRI`, `102016`},
						{`ESRI`, `102017`},
						{`ESRI`, `102018`},
						{`ESRI`, `102031`},
						{`ESRI`, `102034`},
						{`ESRI`, `102035`},
						{`ESRI`, `102038`},
						{`ESRI`, `102068`},
						{`ESRI`, `102069`},
						{`ESRI`, `102299`},
						{`ESRI`, `102423`},
						{`ESRI`, `102498`},
						{`ESRI`, `102581`},
						{`ESRI`, `102582`},
						{`ESRI`, `102585`},
						{`ESRI`, `102586`},
						{`ESRI`, `53001`},
						{`ESRI`, `53002`},
						{`ESRI`, `53003`},
						{`ESRI`, `53004`},
						{`ESRI`, `53008`},
						{`ESRI`, `53009`},
						{`ESRI`, `53010`},
						{`ESRI`, `53011`},
						{`ESRI`, `53012`},
						{`ESRI`, `53013`},
						{`ESRI`, `53014`},
						{`ESRI`, `53015`},
						{`ESRI`, `53016`},
						{`ESRI`, `53017`},
						{`ESRI`, `53018`},
						{`ESRI`, `53019`},
						{`ESRI`, `53021`},
						{`ESRI`, `53022`},
						{`ESRI`, `53023`},
						{`ESRI`, `53024`},
						{`ESRI`, `53025`},
						{`ESRI`, `53026`},
						{`ESRI`, `53027`},
						{`ESRI`, `53028`},
						{`ESRI`, `53029`},
						{`ESRI`, `53030`},
						{`ESRI`, `53031`},
						{`ESRI`, `53032`},
						{`ESRI`, `53034`},
						{`ESRI`, `53035`},
						{`ESRI`, `53036`},
						{`ESRI`, `53037`},
						{`ESRI`, `53042`},
						{`ESRI`, `53043`},
						{`ESRI`, `53044`},
						{`ESRI`, `53045`},
						{`ESRI`, `53046`},
						{`ESRI`, `53048`},
						{`ESRI`, `53049`},
						{`ESRI`, `53074`},
						{`ESRI`, `53075`},
						{`ESRI`, `53076`},
						{`ESRI`, `53077`},
						{`ESRI`, `53078`},
						{`ESRI`, `53079`},
						{`ESRI`, `53080`},
						{`ESRI`, `54001`},
						{`ESRI`, `54002`},
						{`ESRI`, `54003`},
						{`ESRI`, `54004`},
						{`ESRI`, `54008`},
						{`ESRI`, `54009`},
						{`ESRI`, `54010`},
						{`ESRI`, `54011`},
						{`ESRI`, `54012`},
						{`ESRI`, `54013`},
						{`ESRI`, `54014`},
						{`ESRI`, `54015`},
						{`ESRI`, `54016`},
						{`ESRI`, `54017`},
						{`ESRI`, `54018`},
						{`ESRI`, `54019`},
						{`ESRI`, `54021`},
						{`ESRI`, `54022`},
						{`ESRI`, `54023`},
						{`ESRI`, `54024`},
						{`ESRI`, `54025`},
						{`ESRI`, `54026`},
						{`ESRI`, `54027`},
						{`ESRI`, `54028`},
						{`ESRI`, `54029`},
						{`ESRI`, `54030`},
						{`ESRI`, `54031`},
						{`ESRI`, `54032`},
						{`ESRI`, `54034`},
						{`ESRI`, `54042`},
						{`ESRI`, `54043`},
						{`ESRI`, `54044`},
						{`ESRI`, `54045`},
						{`ESRI`, `54046`},
						{`ESRI`, `54048`},
						{`ESRI`, `54049`},
						{`ESRI`, `54050`},
						{`ESRI`, `54051`},
						{`ESRI`, `54052`},
						{`ESRI`, `54053`},
						{`ESRI`, `54074`},
						{`ESRI`, `54075`},
						{`ESRI`, `54076`},
						{`ESRI`, `54077`},
						{`ESRI`, `54078`},
						{`ESRI`, `54079`},
						{`ESRI`, `54080`},
						{`ESRI`, `54090`},
						{`ESRI`, `54091`},
						{`ESRI`, `54098`},
						{`ESRI`, `54099`},
						{`ESRI`, `54100`},
						{`ESRI`, `54101`},
					},
				},
				{
					Query:           `SELECT count(*) FROM postgis_srs_search('POINT(2 49)'::geometry);`,
					ExpectedErr:     `Cannot find SRID (0) in spatial_ref_sys`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT count(*) FROM postgis_srs_search('SRID=3857;POINT(222638 6274861)'::geometry);`,
					Expected: []sql.Row{{42}},
				},
				{
					Query: `SELECT auth_name, count(*) FROM postgis_srs_all() GROUP BY 1 ORDER BY 1;`,
					Expected: []sql.Row{
						{`EPSG`, 6658},
						{`ESRI`, 1380},
						{`IAU_2015`, 2189},
					},
				},
			},
		},
		{
			Name: `postgis column views`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
				`CREATE TABLE gc_test (id int primary key, g geometry(Point, 4326), g2 geometry, gz geometry(LineStringZ, 3857), gzm geometry(PolygonZM, 0), gm geometry(MultiPointM), gg geography(Polygon), gg2 geography, ggz geography(PointZ, 4269));`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: `SELECT f_table_schema, f_table_name, f_geometry_column, coord_dimension, srid, type FROM geometry_columns WHERE f_table_name = 'gc_test' ORDER BY f_geometry_column;`,
					Expected: []sql.Row{
						{`public`, `gc_test`, `g`, 2, 4326, `POINT`},
						{`public`, `gc_test`, `g2`, 2, 0, `GEOMETRY`},
						{`public`, `gc_test`, `gm`, 3, 0, `MULTIPOINTM`},
						{`public`, `gc_test`, `gz`, 3, 3857, `LINESTRING`},
						{`public`, `gc_test`, `gzm`, 4, 0, `POLYGON`},
					},
				},
				{
					Query: `SELECT f_table_schema, f_table_name, f_geography_column, coord_dimension, srid, type FROM geography_columns WHERE f_table_name = 'gc_test' ORDER BY f_geography_column;`,
					Expected: []sql.Row{
						{`public`, `gc_test`, `gg`, 2, 4326, `Polygon`},
						{`public`, `gc_test`, `gg2`, nil, 0, `Geometry`},
						{`public`, `gc_test`, `ggz`, 3, 4269, `PointZ`},
					},
				},
				{
					Query:    `SELECT f_table_catalog = current_database() FROM geometry_columns WHERE f_table_name = 'gc_test' LIMIT 1;`,
					Expected: []sql.Row{{`t`}},
				},
			},
		},
		{
			Name: `postgis geometry column management`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
				`CREATE TABLE mg (id int primary key);`,
				`SELECT AddGeometryColumn('mg', 'g', 4326, 'POINT', 2);`,
				`SELECT AddGeometryColumn('public', 'mg', 'g2', 3857, 'LINESTRINGZ', 3);`,
				`SELECT AddGeometryColumn('', 'public', 'mg', 'g3', 0, 'GEOMETRY', 2);`,
				`SELECT AddGeometryColumn('mg', 'g4', 4326, 'POINTM', 3);`,
				`SELECT AddGeometryColumn('mg', 'g5', 4326, 'point', 4);`,
				`SELECT AddGeometryColumn('mg', 'Mixed', 4326, 'MultiPolygon', 3);`,
				`INSERT INTO mg (id, g, g2) VALUES (1, 'SRID=4326;POINT(1 2)', 'SRID=3857;LINESTRING Z (0 0 0,1 1 1)');`,
				`SELECT UpdateGeometrySRID('mg', 'g2', 4326);`,
				`CREATE TABLE mg2 (id int primary key);`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: `SELECT f_geometry_column, coord_dimension, srid, type FROM geometry_columns WHERE f_table_name = 'mg' ORDER BY 1;`,
					Expected: []sql.Row{
						{`Mixed`, 3, 4326, `MULTIPOLYGON`},
						{`g`, 2, 4326, `POINT`},
						{`g2`, 3, 4326, `LINESTRING`},
						{`g3`, 2, 0, `GEOMETRY`},
						{`g4`, 3, 4326, `POINTM`},
						{`g5`, 4, 4326, `POINT`},
					},
				},
				{
					Query:    `SELECT ST_AsEWKT(g2) FROM mg;`,
					Expected: []sql.Row{{`SRID=4326;LINESTRING(0 0 0,1 1 1)`}},
				},
				{
					Query:    `SELECT Find_SRID('public', 'mg', 'g'), Find_SRID('', 'mg', 'g2'), Find_SRID('', 'public.mg', 'g2');`,
					Expected: []sql.Row{{4326, 4326, 4326}},
				},
				{
					Query:           `SELECT Find_SRID('public', 'mg', 'nope');`,
					ExpectedErr:     `find_srid() - could not find the corresponding SRID - is the geometry registered in the GEOMETRY_COLUMNS table?  Is there an uppercase/lowercase mismatch?`,
					ExpectedErrCode: `P0001`,
				},
				{
					Query:           `SELECT AddGeometryColumn('mg', 'g6', 4326, 'FOO', 2);`,
					ExpectedErr:     `Invalid type name "FOO(2)" - valid ones are:`,
					ExpectedErrCode: `P0001`,
				},
				{
					Query:           `SELECT AddGeometryColumn('mg', 'g7', 123456, 'POINT', 2);`,
					ExpectedErr:     `AddGeometryColumn() - invalid SRID`,
					ExpectedErrCode: `P0001`,
				},
				{
					Query:           `SELECT AddGeometryColumn('nope', 'g', 4326, 'POINT', 2);`,
					ExpectedErr:     `Table 'nope' does not occur in the search_path`,
					ExpectedErrCode: `P0001`,
				},
				{
					Query:           `SELECT AddGeometryColumn('mg', 'g8', 4326, 'POINTZM', 4);`,
					ExpectedErr:     `TypeM needs 3 dimensions`,
					ExpectedErrCode: `P0001`,
				},
				{
					Query:           `SELECT AddGeometryColumn('mg', 'g8', 4326, 'POINTM', 2);`,
					ExpectedErr:     `Invalid type name "POINTM(2)" - valid ones are:`,
					ExpectedErrCode: `P0001`,
				},
				{
					Query:           `SELECT AddGeometryColumn('nope', 'mg', 'x', 4326, 'POINT', 2);`,
					ExpectedErr:     `Schema 'nope' is not a valid schemaname`,
					ExpectedErrCode: `P0001`,
				},
				{
					Query:           `SELECT UpdateGeometrySRID('mg', 'nope', 4326);`,
					ExpectedErr:     `column not found in geometry_columns table`,
					ExpectedErrCode: `P0001`,
				},
				{
					Query:           `SELECT UpdateGeometrySRID('mg', 'g4', 123456);`,
					ExpectedErr:     `invalid SRID: 123456 not found in spatial_ref_sys`,
					ExpectedErrCode: `P0001`,
				},
				{
					Query:           `SELECT DropGeometryColumn('mg', 'nope');`,
					ExpectedErr:     `column not found in geometry_columns table`,
					ExpectedErrCode: `P0001`,
				},
				{
					Query:    `SELECT UpdateGeometrySRID('public', 'mg', 'g', 3857), UpdateGeometrySRID('', 'public', 'mg', 'g3', 4269);`,
					Expected: []sql.Row{{`public.mg.g SRID changed to 3857`, `public.mg.g3 SRID changed to 4269`}},
				},
				{
					Query:    `SELECT DropGeometryColumn('mg', 'Mixed');`,
					Expected: []sql.Row{{`public.mg.Mixed effectively removed.`}},
				},
				{
					Query:    `SELECT DropGeometryColumn('public', 'mg', 'g4'), DropGeometryColumn('', 'public', 'mg', 'g5');`,
					Expected: []sql.Row{{`public.mg.g4 effectively removed.`, `public.mg.g5 effectively removed.`}},
				},
				{
					Query:    `SELECT DropGeometryTable('public', 'mg2');`,
					Expected: []sql.Row{{`public.mg2 dropped.`}},
				},
				{
					Query:    `SELECT DropGeometryTable('mg');`,
					Expected: []sql.Row{{`public.mg dropped.`}},
				},
			},
		},
		{
			Name: `postgis gml spatial reference systems`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Polygon srsName="EPSG:4326"><gml:exterior><gml:LinearRing srsName="EPSG:3857"><gml:posList>0 0 100000 0 100000 100000 0 0</gml:posList></gml:LinearRing></gml:exterior></gml:Polygon>'), 6);`,
					Expected: []sql.Row{{`SRID=4326;POLYGON((0 0,100000 0,100000 100000,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Point srsName="urn:ogc:def:crs:EPSG::4326"><gml:pos>1 2</gml:pos></gml:Point>')), ST_AsEWKT(ST_GeomFromGML('<gml:Point srsName="urn:ogc:def:crs:EPSG::3857"><gml:pos>1 2</gml:pos></gml:Point>'));`,
					Expected: []sql.Row{{`SRID=4326;POINT(2 1)`, `SRID=3857;POINT(1 2)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:MultiPoint srsName="EPSG:3857"><gml:pointMember><gml:Point srsName="urn:ogc:def:crs:EPSG::4326"><gml:pos>2 1</gml:pos></gml:Point></gml:pointMember></gml:MultiPoint>'), 4);`,
					Expected: []sql.Row{{`SRID=3857;MULTIPOINT(111319.4908 222684.2085)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:MultiPoint><gml:pointMember><gml:Point srsName="EPSG:3857"><gml:pos>1 2</gml:pos></gml:Point></gml:pointMember><gml:pointMember><gml:Point srsName="EPSG:4326"><gml:pos>1 2</gml:pos></gml:Point></gml:pointMember></gml:MultiPoint>'), 4);`,
					Expected: []sql.Row{{`SRID=3857;MULTIPOINT(1 2,111319.4908 222684.2085)`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromGML('<gml:Point srsName="EPSG:102003"><gml:pos>1 2</gml:pos></gml:Point>'));`,
					Expected: []sql.Row{{`SRID=102003;POINT(1 2)`}},
				},
			},
		},
		{
			Name: `postgis clustering window functions`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: `SELECT proname, prokind FROM pg_proc WHERE proname IN ('st_clusterdbscan','st_clusterintersectingwin','st_clusterwithinwin') ORDER BY 1;`,
					Expected: []sql.Row{
						{`st_clusterdbscan`, `w`},
						{`st_clusterintersectingwin`, `w`},
						{`st_clusterwithinwin`, `w`},
					},
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'POINT(0 0)'::geometry),(2,'POINT(0 1)'),(3,'POINT(10 10)'),(4,'POINT(10 11)'),(5,'POINT(50 50)'),(6,'POINT(0 2)'),(7,NULL),(8,'POINT EMPTY')) SELECT id, ST_ClusterDBSCAN(g, 1.5, 2) OVER (), ST_ClusterDBSCAN(g, 1.5, 3) OVER (), ST_ClusterIntersectingWin(ST_Buffer(g, 0.6)) OVER (), ST_ClusterWithinWin(g, 1.5) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 0, 0, 0, 0},
						{2, 0, 0, 0, 0},
						{3, 1, nil, 1, 1},
						{4, 1, nil, 1, 1},
						{5, nil, nil, 2, 2},
						{6, 0, 0, 0, 0},
						{7, nil, nil, nil, nil},
						{8, nil, nil, 4, 4},
					},
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'POINT(50 50)'::geometry),(2,NULL),(3,'POINT(0 0)'),(4,'POINT(0 1)'),(5,'POINT EMPTY'),(6,'POINT(10 10)'),(7,'POINT(10 11)')) SELECT id, ST_ClusterDBSCAN(g, 1.5, 2) OVER (), ST_ClusterWithinWin(g, 1.5) OVER (), ST_ClusterIntersectingWin(g) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, nil, 0, 0},
						{2, nil, nil, nil},
						{3, 0, 2, 2},
						{4, 0, 2, 3},
						{5, nil, 3, 4},
						{6, 1, 4, 5},
						{7, 1, 4, 6},
					},
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'POINT(0 0)'::geometry),(2,'POINT(1 0)'),(3,'POINT(2 0)'),(4,'POINT(5 0)'),(5,'POINT(4 0)'),(6,'POINT(3 0)'),(7,'POINT(3.5 3)')) SELECT id, ST_ClusterDBSCAN(g, 1, 3) OVER (), ST_ClusterDBSCAN(g, 1, 2) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 0, 0},
						{2, 0, 0},
						{3, 0, 0},
						{4, 0, 0},
						{5, 0, 0},
						{6, 0, 0},
						{7, nil, nil},
					},
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'LINESTRING(0 0,10 0)'::geometry),(2,'POINT(5 1)'),(3,'POLYGON((20 0,30 0,30 10,20 0))'),(4,'POINT(19 0)')) SELECT id, ST_ClusterDBSCAN(g, 1, 1) OVER (), ST_ClusterWithinWin(g, 1) OVER (), ST_ClusterIntersectingWin(ST_Buffer(g, 0.5)) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 0, 0, 0},
						{2, 0, 0, 0},
						{3, 1, 1, 1},
						{4, 1, 1, 2},
					},
				},
				{
					Query: `WITH t(id, grp, g) AS (VALUES (1,1,'POINT(0 0)'::geometry),(2,1,'POINT(0 1)'),(3,2,'POINT(5 5)'),(4,2,'POINT(100 100)')) SELECT id, ST_ClusterDBSCAN(g, 2, 1) OVER (PARTITION BY grp ORDER BY id DESC), ST_ClusterWithinWin(g, 2) OVER (ORDER BY id DESC) FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 0, 2},
						{2, 0, 2},
						{3, 1, 1},
						{4, 0, 0},
					},
				},
				{
					Query:           `SELECT ST_ClusterDBSCAN('POINT(0 0)'::geometry, 1, 1);`,
					ExpectedErr:     `window function st_clusterdbscan requires an OVER clause`,
					ExpectedErrCode: `42809`,
				},
				{
					Query:    `WITH t(g) AS (VALUES ('POINT(0 0)'::geometry), ('POINT(3 0)'::geometry)) SELECT ST_NPoints(ST_Union(ST_Buffer(g, 1))) FROM t;`,
					Expected: []sql.Row{{66}},
				},
			},
		},
		{
			Name: `postgis k-means clustering`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT proname, prokind, pg_get_function_arguments(oid) FROM pg_proc WHERE proname = 'st_clusterkmeans';`,
					Expected: []sql.Row{{`st_clusterkmeans`, `w`, `geom geometry, k integer, max_radius double precision DEFAULT NULL::double precision`}},
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'POINT(0 0)'::geometry),(2,'POINT(0 1)'),(3,NULL),(4,'POINT EMPTY'),(5,'POINT(10 10)')) SELECT id, ST_ClusterKMeans(g, 2) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 1},
						{2, 1},
						{3, -1},
						{4, -1},
						{5, 0},
					},
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'POINT(0 0)'::geometry),(2,'POINT(0 1)'),(3,'POINT(10 10)')) SELECT id, ST_ClusterKMeans(g, 0) OVER (), ST_ClusterKMeans(g, -1) OVER (), ST_ClusterKMeans(g, NULL) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, nil, nil, nil},
						{2, nil, nil, nil},
						{3, nil, nil, nil},
					},
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'POINT(0 0)'::geometry),(2,'POINT(0 1)'),(3,'POINT(10 10)')) SELECT id, ST_ClusterKMeans(g, 1) OVER (), ST_ClusterKMeans(g, 3) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 0, 1},
						{2, 0, 2},
						{3, 0, 0},
					},
				},
				{
					Query:           `WITH t(id, g) AS (VALUES (1,'POINT(0 0)'::geometry),(2,'POINT(0 1)')) SELECT id, ST_ClusterKMeans(g, 5) OVER () FROM t ORDER BY id;`,
					ExpectedErr:     `K (5) must be smaller than the number of rows in the group (2)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'POINT(0 0)'::geometry),(2,'POINT(5 0)'),(3,NULL)) SELECT id, ST_ClusterKMeans(g, 3) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 1},
						{2, 0},
						{3, -1},
					},
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'POINT(0 0)'::geometry),(2,'POINT(0 1)'),(3,'POINT(10 10)'),(4,'POINT(10 12)'),(5,'POINT(30 30)')) SELECT id, ST_ClusterKMeans(g, 1, 2) OVER (), ST_ClusterKMeans(g, 1, 100) OVER (), ST_ClusterKMeans(g, 2, 5) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 2, 0, 2},
						{2, 2, 0, 2},
						{3, 1, 0, 1},
						{4, 1, 0, 1},
						{5, 0, 0, 0},
					},
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'POINT(0 0)'::geometry),(2,'POINT(1 0)'),(3,'POINT(2 0)'),(4,'POINT(3 0)'),(5,'POINT(4 0)'),(6,'POINT(5 0)'),(7,'POINT(6 0)'),(8,'POINT(7 0)')) SELECT id, ST_ClusterKMeans(g, 1, 1) OVER (), ST_ClusterKMeans(g, 1, 2) OVER (), ST_ClusterKMeans(g, 1, 0) OVER (), ST_ClusterKMeans(g, 1, -1) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 3, 1, 0, 0},
						{2, 3, 1, 0, 0},
						{3, 1, 1, 0, 0},
						{4, 1, 1, 0, 0},
						{5, 2, 0, 0, 0},
						{6, 2, 0, 0, 0},
						{7, 0, 0, 0, 0},
						{8, 0, 0, 0, 0},
					},
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'POINT(0 0)'::geometry),(2,'POINT(0 0)'),(3,'POINT(0 0)'),(4,'POINT(10 10)'),(5,'POINT(5 5)')) SELECT id, ST_ClusterKMeans(g, 2) OVER (), ST_ClusterKMeans(g, 4) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 1, 1},
						{2, 1, 1},
						{3, 1, 1},
						{4, 0, 0},
						{5, 0, 2},
					},
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'POINT(0 0)'::geometry),(2,'POINT(0 0)'),(3,'POINT(0 0)'),(4,'POINT(0 0)')) SELECT id, ST_ClusterKMeans(g, 1) OVER (), ST_ClusterKMeans(g, 2) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 0, 0},
						{2, 0, 0},
						{3, 0, 0},
						{4, 0, 0},
					},
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'POINT Z (0 0 0)'::geometry),(2,'POINT Z (0 0 10)'),(3,'POINT Z (0 0 11)')) SELECT id, ST_ClusterKMeans(g, 2) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 1},
						{2, 0},
						{3, 0},
					},
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'LINESTRING Z (0 0 0,0 0 10)'::geometry),(2,'POINT Z (0 0 4)'),(3,'POINT Z (0 0 30)')) SELECT id, ST_ClusterKMeans(g, 2) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 0},
						{2, 0},
						{3, 1},
					},
					Skip: true,
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'POINT M (0 0 1)'::geometry),(2,'POINT M (1 0 100)'),(3,'POINT M (10 0 1)'),(4,'POINT M (9 0 1)')) SELECT id, ST_ClusterKMeans(g, 2) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 1},
						{2, 1},
						{3, 0},
						{4, 0},
					},
				},
				{
					Query:           `WITH t(id, g) AS (VALUES (1,'POINT M (0 0 -1)'::geometry),(2,'POINT M (1 0 1)')) SELECT id, ST_ClusterKMeans(g, 1) OVER () FROM t ORDER BY id;`,
					ExpectedErr:     `lwgeom_cluster_kmeans has an input point geometry with weight in M less or equal to 0`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'LINESTRING(0 0,2 0)'::geometry),(2,'POLYGON((0 1,2 1,2 3,0 1))'),(3,'POINT(10 10)'),(4,'POINT(11 10)')) SELECT id, ST_ClusterKMeans(g, 2) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 1},
						{2, 1},
						{3, 0},
						{4, 0},
					},
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'LINESTRING(0 10,0 0,10 0)'::geometry),(2,'POINT(1 1)'),(3,'POINT(7 7)'),(4,'POINT(100 100)')) SELECT id, ST_ClusterKMeans(g, 3) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 0},
						{2, 0},
						{3, 2},
						{4, 1},
					},
					Skip: true,
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'LINESTRING Z (0 10 0,0 0 0,10 0 0)'::geometry),(2,'POINT Z (1 1 0)'),(3,'POINT Z (7 7 0)'),(4,'POINT Z (100 100 0)')) SELECT id, ST_ClusterKMeans(g, 3) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 2},
						{2, 0},
						{3, 2},
						{4, 1},
					},
					Skip: true,
				},
				{
					Query: `WITH t(id, g) AS (VALUES (1,'MULTIPOINT((0 0),(1 1))'::geometry),(2,'POINT M (1 0 5)'),(3,'POINT ZM (2 0 0 1)')) SELECT id, ST_ClusterKMeans(g, 2) OVER () FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 1},
						{2, 1},
						{3, 0},
					},
				},
				{
					Query: `WITH t(id, grp, g) AS (VALUES (1,1,'POINT(0 0)'::geometry),(2,1,'POINT(0 1)'),(3,2,'POINT(5 5)'),(4,2,'POINT(100 100)'),(5,2,'POINT(99 99)')) SELECT id, ST_ClusterKMeans(g, 2) OVER (PARTITION BY grp) FROM t ORDER BY id;`,
					Expected: []sql.Row{
						{1, 1},
						{2, 0},
						{3, 1},
						{4, 0},
						{5, 0},
					},
				},
			},
		},
		{
			Name: `postgis populate_geometry_columns and constraint functions`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
				`CREATE TABLE pa (id INT PRIMARY KEY, g geometry);`,
				`INSERT INTO pa VALUES (1, 'SRID=3857;POINT(1 2)'), (2, 'SRID=3857;POINT(3 4)');`,
				`CREATE TABLE pb (id INT PRIMARY KEY, g geometry);`,
				`INSERT INTO pb VALUES (1, 'POINT(1 2)'), (2, 'LINESTRING(0 0,1 1)');`,
				`CREATE TABLE pc (id INT PRIMARY KEY, g geometry);`,
				`CREATE TABLE pd (id INT PRIMARY KEY, g geometry, h geometry(Point, 4326));`,
				`INSERT INTO pd VALUES (1, NULL, NULL), (2, 'POINT Z (1 2 3)', 'SRID=4326;POINT(1 1)');`,
				`CREATE TABLE pf (id INT PRIMARY KEY, g geometry);`,
				`INSERT INTO pf VALUES (1, 'MULTIPOINT((1 2))'), (2, 'POINT(3 4)');`,
				`CREATE TABLE pr (id INT PRIMARY KEY, g geometry);`,
				`INSERT INTO pr VALUES (1, 'POINT M (1 2 3)');`,
				`CREATE TABLE px (id INT PRIMARY KEY, "G g" geometry);`,
				`INSERT INTO px VALUES (1, 'SRID=4326;POINT(1 1)');`,
				`SELECT populate_geometry_columns('pa'::regclass);`,
				`SELECT populate_geometry_columns('pd'::regclass);`,
				`SELECT populate_geometry_columns('pf'::regclass);`,
				`SELECT populate_geometry_columns('pr'::regclass, true);`,
				`SELECT populate_geometry_columns('px'::regclass);`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: `SELECT proname, pg_get_function_arguments(oid), pg_get_function_result(oid), proisstrict FROM pg_proc WHERE proname LIKE 'postgis_constraint%' OR proname = 'populate_geometry_columns' ORDER BY 1, 2;`,
					Expected: []sql.Row{
						{`populate_geometry_columns`, `tbl_oid oid, use_typmod boolean DEFAULT true`, `integer`, `f`},
						{`populate_geometry_columns`, `use_typmod boolean DEFAULT true`, `text`, `f`},
						{`postgis_constraint_dims`, `geomschema text, geomtable text, geomcolumn text`, `integer`, `t`},
						{`postgis_constraint_srid`, `geomschema text, geomtable text, geomcolumn text`, `integer`, `t`},
						{`postgis_constraint_type`, `geomschema text, geomtable text, geomcolumn text`, `character varying`, `t`},
					},
				},
				{
					Query: `SELECT c.relname, a.attname, format_type(a.atttypid, a.atttypmod) FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid WHERE c.relname IN ('pa', 'pb', 'pc', 'pd', 'pf', 'pr', 'px') AND a.attnum > 1 ORDER BY 1, 2;`,
					Expected: []sql.Row{
						{`pa`, `g`, `geometry(Point,3857)`},
						{`pb`, `g`, `geometry`},
						{`pc`, `g`, `geometry`},
						{`pd`, `g`, `geometry(PointZ)`},
						{`pd`, `h`, `geometry(Point,4326)`},
						{`pf`, `g`, `geometry(MultiPoint)`},
						{`pr`, `g`, `geometry(PointM)`},
						{`px`, `G g`, `geometry(Point,4326)`},
					},
				},
				{
					Query: `SELECT f_table_name, f_geometry_column, coord_dimension, srid, type FROM geometry_columns WHERE f_table_name IN ('pa', 'pb', 'pc', 'pd', 'pf', 'pr', 'px') ORDER BY 1, 2;`,
					Expected: []sql.Row{
						{`pa`, `g`, 2, 3857, `POINT`},
						{`pb`, `g`, 2, 0, `GEOMETRY`},
						{`pc`, `g`, 2, 0, `GEOMETRY`},
						{`pd`, `g`, 3, 0, `POINT`},
						{`pd`, `h`, 2, 4326, `POINT`},
						{`pf`, `g`, 2, 0, `MULTIPOINT`},
						{`pr`, `g`, 3, 0, `POINTM`},
						{`px`, `G g`, 2, 4326, `POINT`},
					},
				},
				{
					Query: `SELECT id, ST_AsEWKT(g) FROM pf ORDER BY id;`,
					Expected: []sql.Row{
						{1, `MULTIPOINT(1 2)`},
						{2, `MULTIPOINT(3 4)`},
					},
				},
				{
					Query:    `SELECT populate_geometry_columns('pa'::regclass), populate_geometry_columns('pc'::regclass);`,
					Expected: []sql.Row{{0, 0}},
				},
				{
					Query:    `SELECT populate_geometry_columns('pb'::regclass);`,
					Expected: []sql.Row{{0}},
				},
				{
					Query:    `SELECT populate_geometry_columns('pb'::regclass, false);`,
					Expected: []sql.Row{{1}},
					Skip:     true,
				},
				{
					Query:    `SELECT postgis_constraint_srid('public', 'pa', 'g'), postgis_constraint_dims('public', 'pa', 'g'), postgis_constraint_type('public', 'pa', 'g');`,
					Expected: []sql.Row{{nil, nil, nil}},
				},
				{
					Query:    `SELECT postgis_constraint_srid('public', 'nope', 'g'), postgis_constraint_srid(NULL, 'pa', 'g');`,
					Expected: []sql.Row{{nil, nil}},
				},
			},
		},
		{
			Name: `postgis 3D distance functions`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: `SELECT proname, pg_get_function_arguments(oid) FROM pg_proc WHERE proname IN ('st_3ddistance','st_3ddwithin','_st_3ddwithin','st_3ddfullywithin','_st_3ddfullywithin','st_3dintersects','_st_3dintersects','st_3dclosestpoint','st_3dshortestline','st_3dlongestline','st_3dmaxdistance','geometry_distance_centroid_nd') ORDER BY 1;`,
					Expected: []sql.Row{
						{`_st_3ddfullywithin`, `geom1 geometry, geom2 geometry, double precision`},
						{`_st_3ddwithin`, `geom1 geometry, geom2 geometry, double precision`},
						{`_st_3dintersects`, `geom1 geometry, geom2 geometry`},
						{`geometry_distance_centroid_nd`, `geometry, geometry`},
						{`st_3dclosestpoint`, `geom1 geometry, geom2 geometry`},
						{`st_3ddfullywithin`, `geom1 geometry, geom2 geometry, double precision`},
						{`st_3ddistance`, `geom1 geometry, geom2 geometry`},
						{`st_3ddwithin`, `geom1 geometry, geom2 geometry, double precision`},
						{`st_3dintersects`, `geom1 geometry, geom2 geometry`},
						{`st_3dlongestline`, `geom1 geometry, geom2 geometry`},
						{`st_3dmaxdistance`, `geom1 geometry, geom2 geometry`},
						{`st_3dshortestline`, `geom1 geometry, geom2 geometry`},
					},
				},
				{
					Query:    `SELECT o.oprname, o.oprleft::regtype, o.oprright::regtype, c.oprname FROM pg_operator o JOIN pg_operator c ON c.oid = o.oprcom WHERE o.oprname = '<<->>';`,
					Expected: []sql.Row{{`<<->>`, `geometry`, `geometry`, `<<->>`}},
				},
				{
					Query:    `SELECT ST_3DDistance('POINT Z (0 0 0)', 'POINT Z (1 1 1)'), ST_3DDistance('POINT(0 0)', 'POINT Z (3 4 12)'), ST_3DDistance('POINT(0 0)', 'POINT(3 4)');`,
					Expected: []sql.Row{{1.7320508075688772, 5.0, 5.0}},
				},
				{
					Query:    `WITH p(g) AS (SELECT 'POLYGON Z ((0 0 0,10 0 0,10 10 0,0 10 0,0 0 0))'::geometry) SELECT round(ST_3DDistance('POINT Z (5 5 10)', g)::numeric, 9), round(ST_3DDistance('POINT Z (15 5 10)', g)::numeric, 9), round(ST_3DDistance('LINESTRING Z (2 2 5,8 8 5)', g)::numeric, 9), round(ST_3DDistance('LINESTRING Z (5 5 0,5 5 10)', g)::numeric, 9), round(ST_3DDistance('LINESTRING Z (5 5 -1,5 5 10)', g)::numeric, 9), round(ST_3DDistance('LINESTRING Z (20 5 3,5 5 3)', g)::numeric, 9) FROM p;`,
					Expected: []sql.Row{{framework.Numeric(`10.000000000`), framework.Numeric(`11.180339887`), framework.Numeric(`5.000000000`), framework.Numeric(`5.000000000`), framework.Numeric(`0.000000000`), framework.Numeric(`3.000000000`)}},
				},
				{
					Query:    `WITH p(g) AS (SELECT 'POLYGON Z ((0 -5 -5,0 5 -5,0 5 5,0 -5 5,0 -5 -5))'::geometry) SELECT round(ST_3DDistance('LINESTRING Z (0 0 0,10 0 0)', g)::numeric, 9), round(ST_3DDistance('LINESTRING Z (1 0 0,10 0 0)', g)::numeric, 9), round(ST_3DDistance('POINT Z (0 0 0)', g)::numeric, 9), round(ST_3DDistance('LINESTRING Z (0 0 0,10 0 1)', g)::numeric, 9), round(ST_3DDistance('LINESTRING Z (-1 0 0,10 0 0)', g)::numeric, 9) FROM p;`,
					Expected: []sql.Row{{framework.Numeric(`5.000000000`), framework.Numeric(`1.000000000`), framework.Numeric(`0.000000000`), framework.Numeric(`4.975185951`), framework.Numeric(`0.000000000`)}},
				},
				{
					Query:    `WITH p(g) AS (SELECT 'POLYGON Z ((0 0 0,10 0 0,10 10 0,0 10 0,0 0 0))'::geometry) SELECT round(ST_3DDistance('POLYGON Z ((2 2 5,8 2 5,8 8 5,2 2 5))', g)::numeric, 9), round(ST_3DDistance('POLYGON Z ((2 2 5,8 2 5,8 8 1,2 2 5))', g)::numeric, 9), round(ST_3DDistance('POLYGON Z ((5 5 -1,6 5 2,6 6 2,5 5 -1))', g)::numeric, 9) FROM p;`,
					Expected: []sql.Row{{framework.Numeric(`5.000000000`), framework.Numeric(`1.000000000`), framework.Numeric(`0.000000000`)}},
				},
				{
					Query:    `SELECT round(ST_3DDistance('POLYGON Z ((0 0 0,10 0 0,10 10 0,0 10 0,0 0 0),(4 4 0,6 4 0,6 6 0,4 6 0,4 4 0))', 'POINT Z (5 5 3)')::numeric, 9), round(ST_3DDistance('POLYGON Z ((0 0 0,10 0 10,10 10 10,0 10 0,0 0 0))', 'POINT Z (5 5 0)')::numeric, 9), ST_AsText(ST_3DClosestPoint('POLYGON Z ((0 0 0,10 0 10,10 10 10,0 10 0,0 0 0))', 'POINT Z (5 5 0)'), 9);`,
					Expected: []sql.Row{{framework.Numeric(`3.162277660`), framework.Numeric(`3.535533906`), `POINT Z (2.5 5 2.5)`}},
				},
				{
					Query:    `SELECT round(ST_3DDistance('TIN Z (((0 0 0,1 0 0,0 1 0,0 0 0)))', 'POINT Z (0.2 0.2 1)')::numeric, 9), round(ST_3DDistance('POLYHEDRALSURFACE Z (((0 0 0,0 1 0,1 1 0,1 0 0,0 0 0)))', 'POINT Z (0.5 0.5 2)')::numeric, 9), ST_3DDistance('MULTIPOINT Z ((0 0 0),(10 10 10))', 'GEOMETRYCOLLECTION Z (POINT Z (11 10 10), LINESTRING Z (0 0 5,0 0 6))');`,
					Expected: []sql.Row{{framework.Numeric(`1.000000000`), framework.Numeric(`2.000000000`), 1.0}},
				},
				{
					Query:           `SELECT ST_3DDistance('CIRCULARSTRING Z (0 0 0,1 1 0,2 0 0)', 'POINT Z (1 0 0)');`,
					ExpectedErr:     `lw_dist3d_distribute_bruteforce: Unsupported geometry type: CircularString`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_3DDistance('SRID=4326;POINT Z (0 0 0)', 'POINT Z (1 1 1)');`,
					ExpectedErr:     `ST_3DDistance: Operation on mixed SRID geometries (Point, 4326) != (Point, 0)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_3DDistance('POINT Z EMPTY', 'POINT Z (1 1 1)') IS NULL, ST_3DShortestLine('POINT EMPTY', 'POINT(1 1)') IS NULL, ST_3DMaxDistance('POINT EMPTY', 'POINT(1 1)') IS NULL;`,
					Expected: []sql.Row{{`t`, `t`, `t`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_3DShortestLine('LINESTRING Z (0 0 0,10 0 10)', 'LINESTRING Z (0 5 5,10 5 0)'), 9), round(ST_Z(ST_3DClosestPoint('POLYGON Z ((0 0 0,10 0 0,10 10 0,0 10 0,0 0 0))', 'POINT Z (5 5 10)'))::numeric, 9), ST_AsText(ST_3DClosestPoint('POINT Z (5 5 10)', 'POLYGON Z ((0 0 0,10 0 0,10 10 0,0 10 0,0 0 0))'));`,
					Expected: []sql.Row{{`LINESTRING Z (3.333333333 0 3.333333333,3.333333333 5 3.333333333)`, framework.Numeric(`0.000000000`), `POINT Z (5 5 10)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_3DLongestLine('LINESTRING Z (0 0 0,10 0 10)', 'POINT Z (0 5 5)')), ST_3DMaxDistance('LINESTRING Z (0 0 0,10 0 10)', 'POINT Z (0 5 5)');`,
					Expected: []sql.Row{{`LINESTRING Z (10 0 10,0 5 5)`, 12.24744871391589}},
				},
				{
					Query:    `SELECT ST_AsText(ST_3DShortestLine('POINT(0 0)', 'POINT Z (3 4 12)')), ST_AsText(ST_3DClosestPoint('POINT(0 0)', 'POINT Z (3 4 12)')), ST_AsText(ST_3DLongestLine('POINT(0 0)', 'POINT Z (3 4 12)')), ST_3DMaxDistance('POINT(0 0)', 'POINT Z (3 4 12)');`,
					Expected: []sql.Row{{`LINESTRING Z (0 0 12,3 4 12)`, `POINT Z (0 0 12)`, `LINESTRING Z (0 0 12,3 4 12)`, 5.0}},
				},
				{
					Query:    `SELECT ST_AsText(ST_3DClosestPoint('POINT Z (3 4 12)', 'POINT(0 0)')), ST_AsText(ST_3DShortestLine('LINESTRING(0 0,10 0)', 'POINT Z (5 5 5)')), ST_AsText(ST_3DShortestLine('POINT(0 0)', 'POINT(3 4)'));`,
					Expected: []sql.Row{{`POINT Z (3 4 12)`, `LINESTRING Z (5 0 5,5 5 5)`, `LINESTRING(0 0,3 4)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_3DClosestPoint('POLYGON((0 0,10 0,10 10,0 10,0 0))', 'POINT Z (5 5 7)')), ST_AsText(ST_3DClosestPoint('POLYGON Z ((0 0 0,10 0 10,10 10 10,0 10 0,0 0 0))', 'POINT(5 5)'), 9);`,
					Expected: []sql.Row{{`POINT Z (5 5 7)`, `POINT Z (5 5 5)`}},
				},
				{
					Query:    `SELECT ST_3DDWithin('POINT Z (0 0 0)', 'POINT Z (1 1 1)', 1.8), ST_3DDWithin('POINT Z (0 0 0)', 'POINT Z (1 1 1)', 1.7), ST_3DDFullyWithin('LINESTRING Z (0 0 0,10 0 10)', 'POINT Z (0 5 5)', 12), ST_3DDFullyWithin('LINESTRING Z (0 0 0,10 0 10)', 'POINT Z (0 5 5)', 13), _ST_3DDWithin('POINT Z (0 0 0)', 'POINT Z (1 1 1)', 2);`,
					Expected: []sql.Row{{`t`, `f`, `f`, `t`, `t`}},
				},
				{
					Query:    `SELECT ST_3DIntersects('LINESTRING Z (0 0 0,10 0 10)', 'LINESTRING Z (0 10 0,10 -10 10)'), ST_3DIntersects('LINESTRING Z (0 0 0,10 0 10)', 'LINESTRING Z (0 10 1,10 -10 9)'), ST_3DIntersects('POINT Z (5 5 0)', 'POLYGON Z ((0 0 0,10 0 0,10 10 0,0 10 0,0 0 0))'), ST_3DIntersects('LINESTRING Z (0 0 0,10 0 0)', 'POLYGON Z ((0 -5 -5,0 5 -5,0 5 5,0 -5 5,0 -5 -5))'), ST_3DIntersects('POINT(0 0)', 'POINT Z (0 0 5)');`,
					Expected: []sql.Row{{`t`, `t`, `f`, `f`, `t`}},
					Skip:     true,
				},
				{
					Query:    `SELECT ST_3DDWithin('POINT EMPTY', 'POINT Z (0 0 0)', 1), ST_3DIntersects('POINT Z EMPTY', 'POINT Z (0 0 0)');`,
					Expected: []sql.Row{{`f`, `f`}},
				},
				{
					Query:    `SELECT 'POINT Z (0 0 0)'::geometry <<->> 'POINT Z (1 1 1)'::geometry, 'POINT ZM (0 0 0 0)'::geometry <<->> 'POINT ZM (0 0 0 5)'::geometry, 'POINT M (0 0 1)'::geometry <<->> 'POINT M (3 4 13)'::geometry, 'POINT ZM (0 0 0 1)'::geometry <<->> 'POINT Z (0 0 3)'::geometry, 'POINT M (0 0 0)'::geometry <<->> 'POINT Z (3 4 12)'::geometry;`,
					Expected: []sql.Row{{1.7320508075688772, 5.0, 13.0, 3.0, 5.0}},
				},
				{
					Query:    `SELECT 'LINESTRING M (0 0 0,10 0 4)'::geometry <<->> 'POINT M (20 0 0)'::geometry, 'LINESTRING(0 0,10 0)'::geometry <<->> 'POINT(20 5)'::geometry, 'POLYGON((0 0,10 0,10 10,0 10,0 0))'::geometry <<->> 'POINT(5 5)'::geometry, 'LINESTRING Z (0 0 0,10 0 4)'::geometry <<->> 'LINESTRING Z (20 0 0,30 0 9)'::geometry;`,
					Expected: []sql.Row{{10.770329614269007, 11.180339887498949, 0.0, 10.770329614269007}},
				},
				{
					Query:    `SELECT 'POINT EMPTY'::geometry <<->> 'POINT(1 1)'::geometry, 'SRID=4326;POINT(0 0)'::geometry <<->> 'POINT(1 1)'::geometry;`,
					Expected: []sql.Row{{0.0, 1.4142135623730951}},
				},
			},
		},
		{
			Name: `postgis extents and coverage functions`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
				`CREATE TABLE ext1 (id INT PRIMARY KEY, g geometry);`,
				`INSERT INTO ext1 VALUES (1, 'POINT(1 2)'), (2, 'POINT(5 7)');`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_FindExtent('ext1', 'g'), ST_FindExtent('public', 'ext1', 'g');`,
					Expected: []sql.Row{{`BOX(1 2,5 7)`, `BOX(1 2,5 7)`}},
				},
				{
					Query:           `SELECT ST_FindExtent('nope', 'g');`,
					ExpectedErr:     `table not found: nope`,
					ExpectedErrCode: `42P01`,
				},
				{
					Query:           `SELECT ST_FindExtent('nope', 'g');`,
					ExpectedErr:     `relation "nope" does not exist`,
					ExpectedErrCode: `42P01`,
					Skip:            true,
				},
				{
					Query:    `SELECT ST_EstimatedExtent('ext1', 'g') IS NULL, ST_EstimatedExtent('public', 'ext1', 'g') IS NULL, ST_EstimatedExtent('public', 'ext1', 'g', true) IS NULL;`,
					Expected: []sql.Row{{`t`, `t`, `t`}},
				},
				{
					Query:           `WITH t(g) AS (VALUES ('POLYGON((0 0,1 0,1 1,0 0))'::geometry)) SELECT ST_CoverageInvalidEdges(g) OVER () FROM t;`,
					ExpectedErr:     `The GEOS version this PostGIS binary was compiled against (31101) does not include 'GEOSCoverageIsValid' function (3.12 or greater required)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `WITH t(g) AS (VALUES ('POLYGON((0 0,1 0,1 1,0 0))'::geometry)) SELECT ST_CoverageSimplify(g, 1) OVER () FROM t;`,
					ExpectedErr:     `The GEOS version this PostGIS binary was compiled against (31101) not include 'GEOSCoverageSimplifyVW' function (3.12 or greater required)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `WITH t(g) AS (VALUES ('POLYGON((0 0,1 0,1 1,0 0))'::geometry)) SELECT ST_CoverageClean(g) OVER () FROM t;`,
					ExpectedErr:     `The GEOS version this PostGIS binary was compiled against (31101) not include 'GEOSCoverageClean' function (3.14 or greater required)`,
					ExpectedErrCode: `XX000`,
				},
			},
		},
		{
			Name: `postgis ST_RemoveIrrelevantPointsForView`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsText(ST_RemoveIrrelevantPointsForView('LINESTRING(0 0,10 0,20 0,30 0,40 0)', ST_MakeEnvelope(12,-5,18,5))), ST_AsText(ST_RemoveIrrelevantPointsForView('LINESTRING(0 0,10 10,20 0,30 10,40 0,50 10)', ST_MakeEnvelope(12,-5,28,15)));`,
					Expected: []sql.Row{{`LINESTRING(10 0,20 0)`, `LINESTRING(10 10,20 0,30 10)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_RemoveIrrelevantPointsForView('POLYGON((0 0,100 0,100 100,0 100,0 0))', ST_MakeEnvelope(10,10,20,20))), ST_AsText(ST_RemoveIrrelevantPointsForView('POLYGON((0 0,10 0,20 0,30 0,40 0,40 40,30 40,20 40,10 40,0 40,0 0))', ST_MakeEnvelope(12,-5,28,45), true));`,
					Expected: []sql.Row{{`POLYGON((0 0,100 0,100 100,0 100,0 0))`, `POLYGON((10 0,20 0,30 0,30 40,20 40,10 40,10 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_RemoveIrrelevantPointsForView('POINT(50 50)', ST_MakeEnvelope(0,0,10,10))), ST_AsText(ST_RemoveIrrelevantPointsForView('MULTIPOINT((50 50),(5 5))', ST_MakeEnvelope(0,0,10,10))), ST_AsText(ST_RemoveIrrelevantPointsForView('LINESTRING(0 0,1 1)', ST_MakeEnvelope(10,10,20,20)));`,
					Expected: []sql.Row{{`POINT(50 50)`, `MULTIPOINT((50 50),(5 5))`, `LINESTRING EMPTY`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_RemoveIrrelevantPointsForView('LINESTRING(-10 5,-5 5,5 5,15 5,20 5,25 5)', ST_MakeEnvelope(0,0,10,10))), ST_AsText(ST_RemoveIrrelevantPointsForView('LINESTRING(5 -10,5 -5,5 5,5 15,5 20)', ST_MakeEnvelope(0,0,10,10))), ST_AsText(ST_RemoveIrrelevantPointsForView('LINESTRING(-10 -10,-5 -5,-5 15,15 15,15 -5,-5 -5)', ST_MakeEnvelope(0,0,10,10)));`,
					Expected: []sql.Row{{`LINESTRING(-5 5,5 5,15 5)`, `LINESTRING(5 -5,5 5,5 15)`, `LINESTRING(-5 15,15 15,15 -5)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_RemoveIrrelevantPointsForView('GEOMETRYCOLLECTION(LINESTRING(0 0,10 0,20 0,30 0,40 0),POINT(1 1))', ST_MakeEnvelope(12,-5,18,5))), ST_AsText(ST_RemoveIrrelevantPointsForView('MULTILINESTRING((0 0,10 0,20 0,30 0,40 0),(0 1,1 1))', ST_MakeEnvelope(12,-5,18,5)));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(LINESTRING(0 0,10 0,20 0,30 0,40 0),POINT(1 1))`, `MULTILINESTRING((10 0,20 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_RemoveIrrelevantPointsForView('MULTIPOLYGON(((0 0,10 0,20 0,30 0,40 0,40 40,0 40,0 0)),((100 100,101 100,101 101,100 100)))', ST_MakeEnvelope(12,-5,28,45))), ST_AsText(ST_RemoveIrrelevantPointsForView('POLYGON((100 100,101 100,101 101,100 100))', ST_MakeEnvelope(0,0,10,10)));`,
					Expected: []sql.Row{{`MULTIPOLYGON(((10 0,20 0,30 0,40 40,0 40,10 0)))`, `POLYGON EMPTY`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_RemoveIrrelevantPointsForView('POLYGON((0 0,10 0,10 10,0 10,0 0),(100 100,101 100,101 101,100 100))', ST_MakeEnvelope(0,0,20,20))), ST_AsText(ST_RemoveIrrelevantPointsForView('POLYGON((0 0,10 0,10 10,0 10,0 0),(100 100,101 100,101 101,100 100))', ST_MakeEnvelope(0,0,5,5)));`,
					Expected: []sql.Row{{`POLYGON((0 0,10 0,10 10,0 10,0 0),(100 100,101 100,101 101,100 100))`, `POLYGON((0 0,10 0,10 10,0 10,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_RemoveIrrelevantPointsForView('POLYGON((-20 -20,40 -20,40 40,-20 40,-20 -20),(-5 -5,-4 -5,-4 -4,-5 -5),(5 5,6 5,6 6,5 5),(15 -5,20 -5,20 15,15 -5))', ST_MakeEnvelope(0,0,10,10)));`,
					Expected: []sql.Row{{`POLYGON((-20 -20,40 -20,40 40,-20 40,-20 -20),(5 5,6 5,6 6,5 5))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_RemoveIrrelevantPointsForView('LINESTRING(-5 5,-4 20,-3 5,-2 20,5 5)', ST_MakeEnvelope(0,0,10,10))), ST_AsText(ST_RemoveIrrelevantPointsForView('LINESTRING(-5 5,-5 20,5 20)', ST_MakeEnvelope(0,0,10,10))), ST_AsText(ST_RemoveIrrelevantPointsForView('LINESTRING(15 5,16 5,5 5)', ST_MakeEnvelope(0,0,10,10)));`,
					Expected: []sql.Row{{`LINESTRING(-2 20,5 5)`, `LINESTRING EMPTY`, `LINESTRING(16 5,5 5)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_RemoveIrrelevantPointsForView('POLYGON((-10 -10,30 -10,30 -5,20 -5,20 30,-10 30,-10 -10))', ST_MakeEnvelope(0,0,10,10))), ST_AsText(ST_RemoveIrrelevantPointsForView('POLYGON((-5 2,-4 2,-3 2,-2 2,5 5,-5 3,-5 2))', ST_MakeEnvelope(0,0,10,10))), ST_AsText(ST_RemoveIrrelevantPointsForView('POLYGON((5 5,15 5,15 6,14 7,13 8,5 5))', ST_MakeEnvelope(0,0,10,10)));`,
					Expected: []sql.Row{{`POLYGON((-10 -10,20 -5,20 30,-10 30,-10 -10))`, `POLYGON((-2 2,5 5,-5 3,-2 2))`, `POLYGON((5 5,15 5,13 8,5 5))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_RemoveIrrelevantPointsForView('LINESTRING(-20 -20,-10 -10,-5 -15,0 -30,20 -20)', ST_MakeEnvelope(0,0,10,10))), ST_AsEWKT(ST_RemoveIrrelevantPointsForView('SRID=4326;LINESTRING Z (0 0 1,10 0 2,20 0 3,30 0 4)', ST_MakeEnvelope(12,-5,18,5))), ST_AsText(ST_RemoveIrrelevantPointsForView('LINESTRING(0 0,10 0)', ST_MakeEnvelope(0,0,10,10)));`,
					Expected: []sql.Row{{`LINESTRING EMPTY`, `SRID=4326;LINESTRING(10 0 2,20 0 3)`, `LINESTRING(0 0,10 0)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_RemoveIrrelevantPointsForView('MULTILINESTRING((-5 -5,-4 -4))', ST_MakeEnvelope(0,0,10,10))), ST_AsText(ST_RemoveIrrelevantPointsForView('MULTIPOLYGON(((11 2,12 2,12 3,11 2)))', ST_MakeEnvelope(0,0,10,10))), ST_AsText(ST_RemoveIrrelevantPointsForView('CIRCULARSTRING(0 0,1 1,2 0)', ST_MakeEnvelope(10,10,20,20))), ST_AsText(ST_RemoveIrrelevantPointsForView('POINT EMPTY', ST_MakeEnvelope(10,10,20,20)));`,
					Expected: []sql.Row{{`MULTILINESTRING EMPTY`, `MULTIPOLYGON EMPTY`, `CIRCULARSTRING(0 0,1 1,2 0)`, `POINT EMPTY`}},
				},
			},
		},
		{
			Name: `postgis MARC21`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsMARC21('SRID=4326;POINT(-4.504289 54.253312)'), ST_AsMARC21('SRID=4326;POINT(-4.504289 54.253312)', 'hdddmmss.sss');`,
					Expected: []sql.Row{{`<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="a">a</subfield><subfield code="d">W0043015</subfield><subfield code="e">W0043015</subfield><subfield code="f">N0541512</subfield><subfield code="g">N0541512</subfield></datafield></record>`, `<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="a">a</subfield><subfield code="d">W0043015.440</subfield><subfield code="e">W0043015.440</subfield><subfield code="f">N0541511.923</subfield><subfield code="g">N0541511.923</subfield></datafield></record>`}},
				},
				{
					Query:    `SELECT ST_AsMARC21('SRID=4326;POINT(-4.504289 54.253312)', 'dddmm.mmmm'), ST_AsMARC21('SRID=4326;POINT(-4.504289 54.253312)', 'ddd.dddddd'), ST_AsMARC21('SRID=4326;POINT(-4.504289 54.253312)', 'hddd.dddddd'), ST_AsMARC21('SRID=4326;POINT(-4.504289 54.253312)', 'hdddmm.mmmm');`,
					Expected: []sql.Row{{`<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="a">a</subfield><subfield code="d">-00430.2573</subfield><subfield code="e">-00430.2573</subfield><subfield code="f">05415.1987</subfield><subfield code="g">05415.1987</subfield></datafield></record>`, `<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="a">a</subfield><subfield code="d">-004.504289</subfield><subfield code="e">-004.504289</subfield><subfield code="f">054.253312</subfield><subfield code="g">054.253312</subfield></datafield></record>`, `<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="a">a</subfield><subfield code="d">W004.504289</subfield><subfield code="e">W004.504289</subfield><subfield code="f">N054.253312</subfield><subfield code="g">N054.253312</subfield></datafield></record>`, `<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="a">a</subfield><subfield code="d">W00430.2573</subfield><subfield code="e">W00430.2573</subfield><subfield code="f">N05415.1987</subfield><subfield code="g">N05415.1987</subfield></datafield></record>`}},
				},
				{
					Query:    `SELECT ST_AsMARC21('SRID=4326;LINESTRING(-4 54,-3 55)'), ST_AsMARC21('SRID=4326;POLYGON((-4 54,-3 54,-3 55,-4 55,-4 54))');`,
					Expected: []sql.Row{{`<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="a">a</subfield><subfield code="d">W0040000</subfield><subfield code="e">W0030000</subfield><subfield code="f">N0550000</subfield><subfield code="g">N0540000</subfield></datafield></record>`, `<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="a">a</subfield><subfield code="d">W0040000</subfield><subfield code="e">W0030000</subfield><subfield code="f">N0550000</subfield><subfield code="g">N0540000</subfield></datafield></record>`}},
				},
				{
					Query:    `SELECT ST_AsMARC21('SRID=4326;MULTIPOINT((-4 54),(5 -10))'), ST_AsMARC21('SRID=4326;GEOMETRYCOLLECTION(POINT(-4 54),LINESTRING(1 1,2 2))');`,
					Expected: []sql.Row{{`<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="a">a</subfield><subfield code="d">W0040000</subfield><subfield code="e">W0040000</subfield><subfield code="f">N0540000</subfield><subfield code="g">N0540000</subfield></datafield><datafield tag="034" ind1="1" ind2=" "><subfield code="a">a</subfield><subfield code="d">E0050000</subfield><subfield code="e">E0050000</subfield><subfield code="f">S0100000</subfield><subfield code="g">S0100000</subfield></datafield></record>`, `<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="a">a</subfield><subfield code="d">W0040000</subfield><subfield code="e">W0040000</subfield><subfield code="f">N0540000</subfield><subfield code="g">N0540000</subfield></datafield><datafield tag="034" ind1="1" ind2=" "><subfield code="a">a</subfield><subfield code="d">E0010000</subfield><subfield code="e">E0020000</subfield><subfield code="f">N0020000</subfield><subfield code="g">N0010000</subfield></datafield></record>`}},
				},
				{
					Query:           `SELECT ST_AsMARC21('POINT(1 2)');`,
					ExpectedErr:     `ST_AsMARC21: Input geometry has unknown (0) SRID`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsMARC21('SRID=3857;POINT(1 2)');`,
					ExpectedErr:     `ST_AsMARC21: Unsupported SRID (3857). Only lon/lat coordinate systems are supported in MARC21/XML Documents.`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsMARC21('SRID=4326;POINT EMPTY') IS NULL;`,
					Expected: []sql.Row{{`t`}},
				},
				{
					Query:           `SELECT ST_AsMARC21('SRID=4326;POINT(1 2)', 'foo');`,
					ExpectedErr:     `invalid output format: "foo"`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromMARC21('<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="d">W0043015</subfield><subfield code="e">W0043015</subfield><subfield code="f">N0541512</subfield><subfield code="g">N0541512</subfield></datafield></record>')), ST_AsEWKT(ST_GeomFromMARC21('<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="d">W0040000</subfield><subfield code="e">W0030000</subfield><subfield code="f">N0550000</subfield><subfield code="g">N0540000</subfield></datafield></record>'));`,
					Expected: []sql.Row{{`POINT(-4.504166666666666 54.25333333333333)`, `POLYGON((-4 55,-3 55,-3 54,-4 54,-4 55))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromMARC21('<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="d">W0040000</subfield><subfield code="e">W0040000</subfield><subfield code="f">N0540000</subfield><subfield code="g">N0540000</subfield></datafield><datafield tag="034" ind1="1" ind2=" "><subfield code="d">E0050000</subfield><subfield code="e">E0050000</subfield><subfield code="f">S0100000</subfield><subfield code="g">S0100000</subfield></datafield></record>')), ST_AsEWKT(ST_GeomFromMARC21('<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="d">W0040000</subfield><subfield code="e">W0040000</subfield><subfield code="f">N0540000</subfield><subfield code="g">N0540000</subfield></datafield><datafield tag="034" ind1="1" ind2=" "><subfield code="d">E0010000</subfield><subfield code="e">E0020000</subfield><subfield code="f">N0020000</subfield><subfield code="g">N0010000</subfield></datafield></record>'));`,
					Expected: []sql.Row{{`MULTIPOINT(-4 54,5 -10)`, `GEOMETRYCOLLECTION(POINT(-4 54),POLYGON((1 2,2 2,2 1,1 1,1 2)))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_GeomFromMARC21('<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="d">-004.504289</subfield><subfield code="e">W00430.2573</subfield><subfield code="f">N054.253312</subfield><subfield code="g">05415.1987</subfield></datafield></record>')), ST_AsEWKT(ST_GeomFromMARC21('<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="d">W0043015.440</subfield><subfield code="e">W0043015.440</subfield><subfield code="f">N0541511.923</subfield><subfield code="g">N0541511.923</subfield></datafield></record>'));`,
					Expected: []sql.Row{{`POLYGON((-4.504289 54.253312,-4.504288333333333 54.253312,-4.504288333333333 54.25331166666667,-4.504289 54.25331166666667,-4.504289 54.253312))`, `POINT(-4.504288888888889 54.25331194444444)`}},
				},
				{
					Query:    `SELECT ST_GeomFromMARC21('<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="035"><subfield code="d">W0040000</subfield></datafield></record>') IS NULL;`,
					Expected: []sql.Row{{`t`}},
				},
				{
					Query:           `SELECT ST_GeomFromMARC21('<foo/>');`,
					ExpectedErr:     `invalid MARC21/XML document. Root element <record> expected but <foo> found.`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_GeomFromMARC21('not xml');`,
					ExpectedErr:     `invalid MARC21/XML document.`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_GeomFromMARC21('<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="d">X0040000</subfield><subfield code="e">W0040000</subfield><subfield code="f">N0540000</subfield><subfield code="g">N0540000</subfield></datafield></record>');`,
					ExpectedErr:     `parse error - invalid literal at 034$d: "X0040000"`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_GeomFromMARC21('<record xmlns="http://www.loc.gov/MARC21/slim"><datafield tag="034" ind1="1" ind2=" "><subfield code="d">W0040000</subfield><subfield code="e">W0040000</subfield><subfield code="f">N0540000</subfield></datafield></record>');`,
					ExpectedErr:     `parse error - the Coded Cartographic Mathematical Data (datafield:034) in the given MARC21/XML is incomplete. Coordinates for subfields "$d","$e","$f" and "$g" are expected.`,
					ExpectedErrCode: `XX000`,
				},
			},
		},
		{
			Name: `postgis ST_GeneratePoints`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_NPoints(ST_GeneratePoints('POLYGON((0 0,10 0,10 10,0 10,0 0))', 5)), ST_GeometryType(ST_GeneratePoints('POLYGON((0 0,10 0,10 10,0 10,0 0))', 5)), ST_Within(ST_GeneratePoints('POLYGON((0 0,10 0,10 10,0 10,0 0),(2 2,8 2,8 8,2 8,2 2))', 50), 'POLYGON((0 0,10 0,10 10,0 10,0 0),(2 2,8 2,8 8,2 8,2 2))');`,
					Expected: []sql.Row{{5, `ST_MultiPoint`, `t`}},
				},
				{
					Query:    `SELECT ST_Equals(ST_GeneratePoints('POLYGON((0 0,10 0,10 10,0 10,0 0))', 4, 7), ST_GeneratePoints('POLYGON((0 0,10 0,10 10,0 10,0 0))', 4, 7)), ST_NPoints(ST_GeneratePoints('POLYGON((0 0,10 0,10 10,0 10,0 0))', 4, 7));`,
					Expected: []sql.Row{{`t`, 4}},
				},
				{
					Query:    `SELECT ST_AsText(ST_GeneratePoints('POLYGON((0 0,10 0,10 10,0 10,0 0))', 2, 1));`,
					Expected: []sql.Row{{`MULTIPOINT((1.393062441335203 8.429653859474033),(3.257723912553178 1.070658155254062))`}},
					Skip:     true,
				},
				{
					Query:    `SELECT ST_GeneratePoints('POLYGON((0 0,10 0,10 10,0 10,0 0))', 0) IS NULL, ST_GeneratePoints('POLYGON EMPTY', 3) IS NULL, ST_GeneratePoints('POLYGON((0 0,10 0,10 10,0 10,0 0))', -1) IS NULL;`,
					Expected: []sql.Row{{`t`, `t`, `t`}},
				},
				{
					Query:           `SELECT ST_GeneratePoints('LINESTRING(0 0,10 0)', 3);`,
					ExpectedErr:     `lwgeom_to_points: unsupported geometry type 'LineString'`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_GeneratePoints('POLYGON((0 0,10 0,10 10,0 10,0 0))', 3, 0);`,
					ExpectedErr:     `ST_GeneratePoints: seed must be greater than zero`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_SRID(ST_GeneratePoints('SRID=4326;POLYGON((0 0,10 0,10 10,0 10,0 0))', 3)), ST_NPoints(ST_GeneratePoints('MULTIPOLYGON(((0 0,1 0,1 1,0 0)),((5 5,6 5,6 6,5 5)))', 7)), ST_NPoints(ST_GeneratePoints('MULTIPOLYGON(((0 0,1 0,1 1,0 0)),((5 5,7 5,7 7,5 5)))', 5)), ST_NPoints(ST_GeneratePoints('MULTIPOLYGON(((0 0,1 0,1 1,0 0)),((5 5,7 5,7 7,5 5)))', 1));`,
					Expected: []sql.Row{{4326, 8, 5, 1}},
				},
				{
					Query:    `SELECT ST_CoordDim(ST_GeneratePoints('POLYGON Z ((0 0 1,10 0 1,10 10 1,0 10 1,0 0 1))', 1, 3));`,
					Expected: []sql.Row{{2}},
				},
				{
					Query:           `SELECT ST_GeneratePoints('POLYGON((0 0,10 0,10 0,0 0))', 2);`,
					ExpectedErr:     `lwpoly_to_points: zero area input polygon, TBD`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_GeneratePoints('GEOMETRYCOLLECTION(POLYGON((0 0,1 0,1 1,0 0)))', 2, 1);`,
					ExpectedErr:     `lwgeom_to_points: unsupported geometry type 'GeometryCollection'`,
					ExpectedErrCode: `XX000`,
				},
			},
		},
		{
			Name: `postgis ST_LineToCurve`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsText(ST_LineToCurve(ST_CurveToLine('CIRCULARSTRING(0 0,1 1,2 0)'))), ST_AsText(ST_LineToCurve(ST_Buffer('POINT(1 3)', 3)));`,
					Expected: []sql.Row{{`CIRCULARSTRING(0 0,1 1,2 0)`, `CURVEPOLYGON(CIRCULARSTRING(4 3,-2 3,4 3))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_LineToCurve('LINESTRING(0 0,1 1,2 0)')), ST_AsText(ST_LineToCurve('LINESTRING(0 0,1 0,2 0,3 0)'));`,
					Expected: []sql.Row{{`LINESTRING(0 0,1 1,2 0)`, `LINESTRING(0 0,1 0,2 0,3 0)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_LineToCurve(ST_CurveToLine('COMPOUNDCURVE(CIRCULARSTRING(0 0,1 1,2 0),(2 0,4 0))'))), ST_AsText(ST_LineToCurve(ST_CurveToLine('CIRCULARSTRING(0 0,1 1,2 0)', 2)));`,
					Expected: []sql.Row{{`COMPOUNDCURVE(CIRCULARSTRING(0 0,1 1,2 0),(2 0,4 0))`, `CIRCULARSTRING(0 0,1 1,2 0)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_LineToCurve(ST_CurveToLine('CIRCULARSTRING(0 0,1 1,2 0,3 -1,4 0)')));`,
					Expected: []sql.Row{{`COMPOUNDCURVE(CIRCULARSTRING(0 0,1 1,2 0),CIRCULARSTRING(2 0,3 -1,4 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_LineToCurve('POINT(1 2)')), ST_AsText(ST_LineToCurve('LINESTRING EMPTY')), ST_AsText(ST_LineToCurve('MULTILINESTRING((0 0,1 0),(2 2,3 3))')), ST_AsText(ST_LineToCurve('GEOMETRYCOLLECTION(LINESTRING(0 0,1 0))'));`,
					Expected: []sql.Row{{`POINT(1 2)`, `LINESTRING EMPTY`, `MULTILINESTRING((0 0,1 0),(2 2,3 3))`, `GEOMETRYCOLLECTION(LINESTRING(0 0,1 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_LineToCurve(ST_CurveToLine('MULTICURVE(CIRCULARSTRING(0 0,1 1,2 0),(5 5,6 6))'))), ST_AsText(ST_LineToCurve(ST_CurveToLine('CURVEPOLYGON(CIRCULARSTRING(0 0,4 0,4 4,0 4,0 0))')));`,
					Expected: []sql.Row{{`MULTICURVE(CIRCULARSTRING(0 0,1 1,2 0),(5 5,6 6))`, `CURVEPOLYGON(CIRCULARSTRING(0 0,4 4,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_LineToCurve(ST_CurveToLine('CIRCULARSTRING Z (0 0 1,1 1 2,2 0 3)'))), ST_AsText(ST_LineToCurve('POLYGON((0 0,1 0,1 1,0 1,0 0))'));`,
					Expected: []sql.Row{{`CIRCULARSTRING Z (0 0 1,1 1 2,2 0 3)`, `POLYGON((0 0,1 0,1 1,0 1,0 0))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_LineToCurve(ST_MakeLine(ARRAY(SELECT ST_MakePoint(cos(radians(a)), sin(radians(a))) FROM generate_series(0, 60, 30) a))), 9), ST_AsText(ST_LineToCurve(ST_MakeLine(ARRAY(SELECT ST_MakePoint(cos(radians(a)), sin(radians(a))) FROM generate_series(0, 90, 30) a))), 9);`,
					Expected: []sql.Row{{`LINESTRING(1 0,0.866025404 0.5,0.5 0.866025404)`, `CIRCULARSTRING(1 0,0.866025404 0.5,6.123233996e-17 1)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_LineToCurve(ST_MakeLine(ARRAY(SELECT ST_MakePoint(cos(radians(a)), sin(radians(a))) FROM unnest(ARRAY[0,30,60,90,110]) a))), 9);`,
					Expected: []sql.Row{{`COMPOUNDCURVE(CIRCULARSTRING(1 0,0.866025404 0.5,6.123233996e-17 1),(6.123233996e-17 1,-0.342020143 0.939692621))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_LineToCurve(ST_MakeLine(ARRAY['POINT(5 5)'::geometry] || ARRAY(SELECT ST_MakePoint(cos(radians(a)), sin(radians(a))) FROM unnest(ARRAY[0,30,60,90,120]) a))), 9);`,
					Expected: []sql.Row{{`COMPOUNDCURVE((5 5,1 0),CIRCULARSTRING(1 0,0.5 0.866025404,-0.5 0.866025404))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_LineToCurve(ST_MakeLine(ARRAY(SELECT ST_MakePoint(cos(radians(a)), sin(radians(a))) FROM unnest(ARRAY[0,10,20,30,40,50]) a) || ARRAY(SELECT ST_MakePoint(cos(radians(a))*3, sin(radians(a))*3) FROM unnest(ARRAY[60,70,80,90,100]) a))), 9);`,
					Expected: []sql.Row{{`COMPOUNDCURVE(CIRCULARSTRING(1 0,0.939692621 0.342020143,0.64278761 0.766044443),(0.64278761 0.766044443,1.5 2.598076211),CIRCULARSTRING(1.5 2.598076211,0.520944533 2.954423259,-0.520944533 2.954423259))`}},
				},
				{
					Query: `SELECT s, ST_GeometryType(ST_LineToCurve(ST_MakeLine(ARRAY(SELECT ST_MakePoint(cos(radians(a)), sin(radians(a))) FROM generate_series(0, 3*s, s) a)))) FROM unnest(ARRAY[10,44,45,46,60,90]) s;`,
					Expected: []sql.Row{
						{10, `ST_CircularString`},
						{44, `ST_CircularString`},
						{45, `ST_CircularString`},
						{46, `ST_LineString`},
						{60, `ST_LineString`},
						{90, `ST_LineString`},
					},
				},
			},
		},
		{
			Name: `postgis MVT`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT encode(ST_AsMVT(q, 'lyr'), 'hex') FROM (SELECT 1 AS id, 'a' AS name, 2.5::float8 AS f, -3 AS i, true AS b, ST_AsMVTGeom('POINT(25 17)'::geometry, ST_MakeBox2D(ST_Point(0,0), ST_Point(4096,4096))) AS geom) q;`,
					Expected: []sql.Row{{`1a4f0a036c79721214120a00000101020203030404180122040932de3f1a0269641a046e616d651a01661a01691a01622202280122030a0161220919000000000000044022023005220238012880207802`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_AsMVTGeom('POINT(25 17)'::geometry, ST_MakeBox2D(ST_Point(0,0), ST_Point(4096,4096))));`,
					Expected: []sql.Row{{`POINT(25 4079)`}},
				},
				{
					Query:    `SELECT encode(ST_AsMVT(q, 'lyr', 4096, 'geom', 'id'), 'hex') FROM (SELECT 7 AS id, 'a' AS name, ST_AsMVTGeom('LINESTRING(0 0,10 10,20 0)'::geometry, ST_MakeBox2D(ST_Point(0,0), ST_Point(4096,4096))) AS geom UNION ALL SELECT 8, 'a', ST_AsMVTGeom('POLYGON((0 0,10 0,10 10,0 10,0 0))'::geometry, ST_MakeBox2D(ST_Point(0,0), ST_Point(4096,4096)))) q;`,
					Expected: []sql.Row{{`1a420a036c797212130807120200001802220909008040121413141412160808120200001803220c090080401a0013140000140f1a046e616d6522030a01612880207802`}},
				},
				{
					Query:    `SELECT encode(ST_AsMVT(q), 'hex') FROM (SELECT 1::int2 AS a, 2::int8 AS b, 1.5::float4 AS c, 3.25::numeric AS d, 'x'::varchar AS e, NULL::text AS n, '2024-01-02'::date AS dt, '{"k":1,"s":"v","z":null}'::jsonb AS j, 'POINT(1 1)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`1a770a0764656661756c741219121000000101020203030404060507000806180122030902021a01611a01621a01631a01641a01651a016e1a0264741a016b1a01731a017a22022801220228022205150000c03f22060a04332e323522030a0178220c0a0a323032342d30312d303222030a01762880207802`}},
				},
				{
					Query:    `SELECT encode(ST_AsMVT(q, 'L', 10, 'g2'), 'hex') FROM (SELECT 'POINT(1 1)'::geometry AS g1, 'POINT(2 2)'::geometry AS g2) q;`,
					Expected: []sql.Row{{`1a460a014c120b12020000180122030904041a026731222c0a2a303130313030303030303030303030303030303030304630334630303030303030303030303046303346280a7802`}},
				},
				{
					Query:    `SELECT encode(ST_AsMVT(q, 'L'), 'hex') FROM (SELECT 1 AS a, NULL::geometry AS geom UNION ALL SELECT 2, 'POINT(3 3)'::geometry) q;`,
					Expected: []sql.Row{{`1a1c0a014c120b12020000180122030906061a0161220228022880207802`}},
				},
				{
					Query:    `SELECT encode(ST_AsMVT(q, 'L', 4096, 'geom', 'fid'), 'hex') FROM (SELECT -5 AS fid, 'POINT(3 3)'::geometry AS geom UNION ALL SELECT NULL, 'POINT(4 4)'::geometry) q;`,
					Expected: []sql.Row{{`1a1a0a014c1207180122030906061207180122030908082880207802`}},
				},
				{
					Query:    `SELECT length(ST_AsMVT(q)), ST_AsMVT(q) IS NULL FROM (SELECT 1 AS a, 'POINT(1 1)'::geometry AS geom WHERE false) q;`,
					Expected: []sql.Row{{0, `f`}},
				},
				{
					Query:           `SELECT encode(ST_AsMVT(q), 'hex') FROM (SELECT 1 AS a) q;`,
					ExpectedErr:     `parse_column_keys: no geometry column found`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT encode(ST_AsMVT(q, 'L', 4096, 'nope'), 'hex') FROM (SELECT 'POINT(1 1)'::geometry AS geom) q;`,
					ExpectedErr:     `parse_column_keys: no geometry column found`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT encode(ST_AsMVT(q), 'hex') FROM (SELECT 'POINT(2.6 3.4)'::geometry AS geom UNION ALL SELECT 'POINT(-2.6 -3.5)'::geometry) q;`,
					Expected: []sql.Row{{`1a200a0764656661756c741207180122030904061207180122030903052880207802`}},
				},
				{
					Query:           `SELECT encode(ST_AsMVT(q), 'hex') FROM (SELECT 'GEOMETRYCOLLECTION(POINT(1 1))'::geometry AS geom) q;`,
					ExpectedErr:     `encode_feature_geometry: 'GeometryCollection' geometry type not supported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT encode(ST_AsMVT(q), 'hex') FROM (SELECT '{"o":{"x":1},"a":[1,2],"f":1.5,"t":true,"bb":"s"}'::jsonb AS j, 'POINT(1 1)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`1a430a0764656661756c74120f1206010003010402180122030902021a01611a01661a016f1a01741a026262220919000000000000f83f2202380122030a01732880207802`}},
				},
				{
					Query:    `SELECT encode(ST_AsMVT(q), 'hex') FROM (SELECT 'MULTIPOINT((1 1),(3 3))'::geometry AS geom UNION ALL SELECT 'MULTILINESTRING((0 0,1 1),(5 5,6 6))'::geometry UNION ALL SELECT 'MULTIPOLYGON(((0 0,4 0,4 4,0 0)),((10 10,14 10,14 14,10 10)))'::geometry) q;`,
					Expected: []sql.Row{{`1a430a0764656661756c74120918012205110202040412101802220c0900000a02020908080a020212161803221209000012080000080f090c0c12080000080f2880207802`}},
				},
				{
					Query:           `SELECT encode(ST_AsMVT(q, 'L', 4096, 'geom', 'id'), 'hex') FROM (SELECT 'x'::text AS id, 'POINT(1 1)'::geometry AS geom) q;`,
					ExpectedErr:     `mvt_agg_transfn: Could not find column 'id' of integer type`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `WITH b(box) AS (SELECT ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1))) SELECT ST_AsText(ST_AsMVTGeom('POINT(0.25 0.25)', box, 10, 0, false)), ST_AsText(ST_AsMVTGeom('POINT(0.35 0.35)', box, 10, 0, false)), ST_AsText(ST_AsMVTGeom('POINT(0.45 0.45)', box, 10, 0, false)), ST_AsText(ST_AsMVTGeom('POINT(-0.25 -0.25)', box, 10, 0, false)) FROM b;`,
					Expected: []sql.Row{{`POINT(2 8)`, `POINT(4 6)`, `POINT(4 6)`, `POINT(-2 12)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_AsMVTGeom('LINESTRING(0 0,1 1)', ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1)), 10, 0, false)), ST_AsText(ST_AsMVTGeom('LINESTRING(0 0,1 1)', ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1)), 10));`,
					Expected: []sql.Row{{`LINESTRING(0 10,10 0)`, `LINESTRING(0 10,10 0)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_AsMVTGeom('LINESTRING(-5 0.5,5 0.5)', ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1)), 10, 2, true)), ST_AsText(ST_AsMVTGeom('LINESTRING(-5 0.5,5 0.5)', ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1)), 10, 2, false));`,
					Expected: []sql.Row{{`LINESTRING(-2 5,12 5)`, `LINESTRING(-50 5,50 5)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_AsMVTGeom('POINT(50 50)', ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1)), 10, 2, true)) IS NULL, ST_AsText(ST_AsMVTGeom('POINT(50 50)', ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1)), 10, 2, false));`,
					Expected: []sql.Row{{`t`, `POINT(500 -490)`}},
				},
				{
					Query:    `SELECT ST_AsMVTGeom('LINESTRING(0 0,0.01 0.01)', ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1)), 10) IS NULL, ST_AsMVTGeom('POLYGON((0 0,0.01 0,0.01 0.01,0 0))', ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1)), 10) IS NULL;`,
					Expected: []sql.Row{{`t`, `t`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_AsMVTGeom('MULTIPOINT(0.1 0.1,0.1 0.1,0.5 0.5)', ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1)), 10)), ST_AsText(ST_AsMVTGeom('GEOMETRYCOLLECTION(POINT(0.1 0.1),LINESTRING(0 0,1 1))', ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1)), 10));`,
					Expected: []sql.Row{{`MULTIPOINT((1 9),(5 5))`, `LINESTRING(0 10,10 0)`}},
				},
				{
					Query:    `SELECT ST_AsMVTGeom('POINT EMPTY', ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1))) IS NULL, ST_AsMVTGeom(NULL, ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1))) IS NULL;`,
					Expected: []sql.Row{{`t`, `t`}},
				},
				{
					Query:           `SELECT ST_AsMVTGeom('POINT(1 1)', ST_MakeBox2D(ST_Point(0,0), ST_Point(0,1)));`,
					ExpectedErr:     `ST_AsMVTGeom: Geometric bounds are too small`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsMVTGeom('POINT(1 1)', ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1)), 0);`,
					ExpectedErr:     `ST_AsMVTGeom: Extent must be greater than 0`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsText(ST_AsMVTGeom('POINT Z (0.5 0.5 9)', ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1)), 10)), ST_SRID(ST_AsMVTGeom('SRID=3857;POINT(0.5 0.5)', ST_MakeBox2D(ST_Point(0,0), ST_Point(1,1)), 10));`,
					Expected: []sql.Row{{`POINT Z (5 5 9)`, 3857}},
				},
				{
					Query:    `WITH b(box) AS (SELECT ST_MakeBox2D(ST_Point(0,0), ST_Point(10,10))) SELECT ST_AsText(ST_AsMVTGeom('LINESTRING(2 2,8 8,8 2)', box, 10, 0, true)), ST_AsText(ST_AsMVTGeom('LINESTRING(-5 5,15 5)', box, 10, 0, true)), ST_AsText(ST_AsMVTGeom('MULTILINESTRING((-5 5,15 5),(1 1,2 2))', box, 10, 0, true)), ST_AsText(ST_AsMVTGeom('LINESTRING(2 2,2 2.01,8 8)', box, 10, 0, false)) FROM b;`,
					Expected: []sql.Row{{`LINESTRING(2 8,8 2,8 8)`, `LINESTRING(0 5,10 5)`, `MULTILINESTRING((0 5,10 5),(1 9,2 8))`, `LINESTRING(2 8,8 2)`}},
				},
				{
					Query:    `WITH b(box) AS (SELECT ST_MakeBox2D(ST_Point(0,0), ST_Point(10,10))) SELECT ST_AsText(ST_AsMVTGeom('POLYGON((0 0,0 10,10 10,10 0,0 0))', box, 10, 0, false)), ST_AsText(ST_AsMVTGeom('POLYGON((0 0,10 0,10 10,0 10,0 0))', box, 10, 0, false)), ST_AsText(ST_AsMVTGeom('POLYGON((2 2,8 2,8 8,2 2))', box, 10, 0, false)) FROM b;`,
					Expected: []sql.Row{{`POLYGON((10 0,10 10,0 10,0 0,10 0))`, `POLYGON((0 10,0 0,10 0,10 10,0 10))`, `POLYGON((8 2,8 8,2 8,8 2))`}},
				},
				{
					Query:    `SELECT encode(ST_AsMVT(q), 'hex') FROM (SELECT 1::int2 AS a, 2::int8 AS b, 1.5::float4 AS c, 3.25::numeric AS d, 'x'::varchar AS e, NULL::text AS n, '2024-01-02'::date AS dt, '{"k":1,"s":"v","z":null}'::jsonb AS j, 'POINT(1 1)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`1a770a0764656661756c741219121000000101020203030404060507000806180122030902021a01611a01621a01631a01641a01651a016e1a0264741a016b1a01731a017a22022801220228022205150000c03f22060a04332e323522030a0178220c0a0a323032342d30312d303222030a01762880207802`}},
				},
				{
					Query:    `SELECT encode(ST_AsMVT(q), 'hex') FROM (SELECT '{"o":{"x":1},"a":[1,2],"f":1.5,"t":true,"bb":"s"}'::jsonb AS j, 'POINT(1 1)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`1a430a0764656661756c74120f1206010003010402180122030902021a01611a01661a016f1a01741a026262220919000000000000f83f2202380122030a01732880207802`}},
				},
				{
					Query:    `WITH b(box) AS (SELECT ST_MakeBox2D(ST_Point(0,0), ST_Point(10,10))) SELECT ST_AsText(ST_AsMVTGeom('MULTIPOLYGON(((0 0,4 0,4 4,0 4,0 0)),((2 2,6 2,6 6,2 6,2 2)))', box, 10, 0, false)), ST_AsText(ST_AsMVTGeom('POLYGON((0 0,10 10,10 0,0 10,0 0))', box, 10, 0, false)), ST_AsText(ST_AsMVTGeom('POLYGON Z ((0 0 1,10 0 1,10 10 1,0 0 1))', box, 10, 0, false)) FROM b;`,
					Expected: []sql.Row{{`MULTIPOLYGON(((4 8,4 6,2 6,2 4,6 4,6 8,4 8)),((4 8,4 10,0 10,0 6,2 6,2 8,4 8)))`, `MULTIPOLYGON(((10 0,10 10,5 5,10 0)),((0 0,5 5,0 10,0 0)))`, `POLYGON((10 0,10 10,0 10,10 0))`}},
				},
				{
					Query:    `WITH b(box) AS (SELECT ST_MakeBox2D(ST_Point(0,0), ST_Point(10,10))) SELECT ST_AsText(ST_AsMVTGeom('POLYGON((0 0,10 0,10 10,0 10,0 0),(2 2,2 8,8 8,8 2,2 2))', box, 10, 0)), ST_AsText(ST_AsMVTGeom('POLYGON((-5 -5,5 -5,5 5,-5 5,-5 -5))', box, 10, 0, true)), ST_AsText(ST_AsMVTGeom('POLYGON((-5 -5,5 -5,5 5,-5 5,-5 -5))', box, 10, 1, true)) FROM b;`,
					Expected: []sql.Row{{`POLYGON((0 10,0 0,10 0,10 10,0 10),(8 2,2 2,2 8,8 8,8 2))`, `POLYGON((0 10,0 5,5 5,5 10,0 10))`, `POLYGON((-1 11,-1 5,5 5,5 11,-1 11))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_AsMVTGeom('POLYGON((0 0,10 0,10 10,0 10,0 0))'::geometry, ST_MakeBox2D(ST_Point(0,0), ST_Point(4096,4096)))), ST_AsText(ST_AsMVTGeom('POLYGON((0 0,0 10,10 10,10 0,0 0))'::geometry, ST_MakeBox2D(ST_Point(0,0), ST_Point(4096,4096))));`,
					Expected: []sql.Row{{`POLYGON((0 4096,0 4086,10 4086,10 4096,0 4096))`, `POLYGON((10 4086,10 4096,0 4096,0 4086,10 4086))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_AsMVTGeom(ST_Buffer('POINT(5 5)'::geometry, 4), ST_MakeBox2D(ST_Point(0,0), ST_Point(10,10)), 100, 0));`,
					Expected: []sql.Row{{`POLYGON((50 10,58 11,65 13,72 17,78 22,83 28,87 35,89 42,90 50,89 58,87 65,83 72,78 78,72 83,65 87,58 89,50 90,42 89,35 87,28 83,22 78,17 72,13 65,11 58,10 50,11 42,13 35,17 28,22 22,28 17,35 13,42 11,50 10))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_AsMVTGeom(ST_Buffer('POINT(5 5)'::geometry, 4), ST_MakeBox2D(ST_Point(2,2), ST_Point(8,8)), 60, 5));`,
					Expected: []sql.Row{{`POLYGON((48 -5,52 -3,58 2,63 8,65 11,65 48,63 52,58 58,52 63,48 65,11 65,8 63,2 58,-3 52,-5 48,-5 11,-3 8,2 2,8 -3,11 -5,48 -5))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_AsMVTGeom('MULTIPOLYGON(((0 0,3 0,3 3,0 0)),((6 6,9 6,9 9,6 6)))'::geometry, ST_MakeBox2D(ST_Point(0,0), ST_Point(10,10)), 10, 0));`,
					Expected: []sql.Row{{`MULTIPOLYGON(((3 7,3 10,0 10,3 7)),((9 1,9 4,6 4,9 1)))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_AsMVTGeom('GEOMETRYCOLLECTION(POLYGON((0 0,4 0,4 4,0 0)),LINESTRING(0 0,9 9))'::geometry, ST_MakeBox2D(ST_Point(0,0), ST_Point(10,10)), 10, 0));`,
					Expected: []sql.Row{{`POLYGON((4 6,4 10,0 10,4 6))`}},
				},
				{
					Query:    `SELECT encode(ST_AsMVT(q), 'hex') FROM (SELECT 1 AS id, ST_AsMVTGeom('POLYGON((0 0,10 0,10 10,0 10,0 0),(2 2,2 8,8 8,8 2,2 2))'::geometry, ST_MakeBox2D(ST_Point(0,0), ST_Point(10,10)), 10, 0) AS geom) q;`,
					Expected: []sql.Row{{`1a360a0764656661756c74121e12020000180322160900141a0013140000140f09030f1a0b00000c0c000f1a026964220228012880207802`}},
				},
			},
		},
		{
			Name: `postgis Geobuf`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT encode(ST_AsGeobuf(q), 'hex') FROM (SELECT 1 AS id, 'a' AS name, 2.5::float8 AS f, -3 AS i, true AS b, 'POINT(1.5 2.25)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`0a0269640a046e616d650a01660a01690a0162180222350a330a0808001a04ac02c2036a0218016a030a01616a091100000000000004406a0220036a030a0174720a00000101020203030404`}},
				},
				{
					Query:    `SELECT encode(ST_AsGeobuf(q, 'g'), 'hex') FROM (SELECT 'LINESTRING(0 0,1 1,2 0)'::geometry AS g UNION ALL SELECT 'POLYGON((0 0,10 0,10 10,0 0),(1 1,2 1,2 2,1 1))'::geometry) q;`,
					Expected: []sql.Row{{`180022260a0c0a0a08021a060000020202010a160a140804120203031a0c000014000014020202000002`}},
				},
				{
					Query:    `SELECT encode(ST_AsGeobuf(q), 'hex') FROM (SELECT 'POINT(1.123456789 2)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`1801220a0a080a0608001a021828`}},
				},
				{
					Query:    `SELECT encode(ST_AsGeobuf(q), 'hex') FROM (SELECT 'POINT Z (1 2 3.5)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`10031801220b0a090a0708001a03142846`}},
				},
				{
					Query:    `SELECT encode(ST_AsGeobuf(q), 'hex') FROM (SELECT 'MULTIPOINT((1 1),(2 2))'::geometry AS geom UNION ALL SELECT 'MULTILINESTRING((0 0,1 1),(5 5,6 6))'::geometry UNION ALL SELECT 'MULTIPOLYGON(((0 0,4 0,4 4,0 0)),((10 10,14 10,14 14,10 10),(11 11,12 11,12 12,11 11)))'::geometry UNION ALL SELECT 'GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 0,1 1))'::geometry) q;`,
					Expected: []sql.Row{{`1800225a0a0a0a0808011a04020202020a120a100803120202021a08000002020a0a02020a200a1e080512060201030203031a120000080000081414080000081616020000020a160a140806220608001a020202220808021a0400000202`}},
				},
				{
					Query:    `SELECT encode(ST_AsGeobuf(q), 'hex') FROM (SELECT 1.5::float4 AS a, 3.25::numeric AS b, NULL::int AS n, 7::int8 AS c, '{"x":1}'::jsonb AS j, 'POINT(1 1)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`0a01610a01620a016e0a01630a016a180022370a350a0608001a0202026a0911000000000000f83f6a060a04332e32356a0218076a0a0a087b2278223a20317d72080000010103020403`}},
				},
				{
					Query:    `SELECT ST_AsGeobuf(q) IS NULL FROM (SELECT 'POINT(1 1)'::geometry AS geom WHERE false) q;`,
					Expected: []sql.Row{{`t`}},
				},
				{
					Query:    `SELECT encode(ST_AsGeobuf(q), 'hex') FROM (SELECT NULL::geometry AS geom UNION ALL SELECT 'POINT(1 1)'::geometry) q;`,
					Expected: []sql.Row{{`1800220a0a080a0608001a020202`}},
				},
				{
					Query:           `SELECT encode(ST_AsGeobuf(q), 'hex') FROM (SELECT 1 AS a) q;`,
					ExpectedErr:     `encode_keys: no geometry column found`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT encode(ST_AsGeobuf(q), 'hex') FROM (SELECT 'POINT EMPTY'::geometry AS geom) q;`,
					Expected: []sql.Row{{`180022060a040a020800`}},
				},
				{
					Query:    `SELECT encode(ST_AsGeobuf(q), 'hex') FROM (SELECT 'POLYGON((1 1,5 1,5 5,1 1))'::geometry AS geom) q;`,
					Expected: []sql.Row{{`1800220e0a0c0a0a08041a06020208000008`}},
				},
				{
					Query:    `SELECT encode(ST_AsGeobuf(q), 'hex') FROM (SELECT 'LINESTRING(0 0,0.25 0,0.123 0.5)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`1802220e0a0c0a0a08021a06000032001764`}},
				},
				{
					Query:    `SELECT encode(ST_AsGeobuf(q), 'hex') FROM (SELECT 'LINESTRING(0.123456789123 0.987654321987,0.123456789123 0.987654321987,0.123456789123 0.987654321987,0.123456789123 0.987654321987,0.123456789123 0.987654321987,0.123456789123 0.987654321987,0.123456789123 0.987654321987,0.123456789123 0.987654321987,0.123456789123 0.987654321987,0.123456789123 0.987654321987,0.123456789123 0.987654321987,0.123456789123 0.987654321987)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`22240a220a2008021a1c82890f8ec87800000000000000000000000000000000000000000000`}},
				},
				{
					Query:    `SELECT encode(ST_AsGeobuf(q), 'hex') FROM (SELECT 'LINESTRING(0.1 0.1, 0.11 0.1, 0.111 0, 0.1111 0)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`180422140a120a1008021a0cd00fd00fc8010014cf0f0200`}},
				},
			},
		},
		{
			Name: `postgis FlatGeobuf`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
				`SET TIME ZONE 'UTC';`,
				`CREATE TABLE fgbt (id int, geom geometry, num int, name text);`,
				`SELECT ST_FromFlatGeobufToTable('public', 'fgb_new', (SELECT ST_AsFlatGeobuf(q) FROM (SELECT 7 AS num, 'a'::text AS name, 2.5::float8 AS f, true AS b, 3::int2 AS s, 9::int8 AS l, 1.5::float4 AS r, 'POINT(1 2)'::geometry AS geom) q));`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 'POINT(1 2)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`6667620366676201240000001c00000018000800000000000500000000000000000000000000060018000000000100003c00000010000000000000000000060008000400060000000c0000000800080000000400080000000400000002000000000000000000f03f0000000000000040`}},
				},
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 1 AS id, 'a' AS name, 'POINT(1 2)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`66676203666762016c000000200000000000000018000c000000000005000000000000000000080000000600180000000001000004000000020000002800000004000000e8ffffff0000000b04000000040000006e616d650000000008000c0008000700080000000000000504000000020000006964000054000000100000000000000008000c00080004000800000008000000200000000d000000000001000000010001000000610000000800080000000400080000000400000002000000000000000000f03f0000000000000040`}},
				},
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q, true), 'hex') FROM (SELECT 1 AS id, 'POINT(1 2)'::geometry AS geom UNION ALL SELECT 2, 'POINT(3 4)'::geometry) q;`,
					Expected: []sql.Row{{`6667620366676201840000001c000000000016001c0000000800070000000000000000000c00100016000000000000011c00000010000000020000000000000000000000010000003400000004000000000000000000f03f0000000000000040000000000000084000000000000010400000000008000c00080007000800000000000005040000000200000069640000000000000000f03f000000000000004000000000000008400000000000001040010000000000000000000000000008400000000000001040000000000000084000000000000010400000000000000000000000000000f03f0000000000000040000000000000f03f000000000000004050000000000000004c000000100000000000000008000c00080004000800000008000000180000000600000000000200000000000800080000000400080000000400000002000000000000000000084000000000000010404c000000100000000000000008000c00080004000800000008000000180000000600000000000100000000000800080000000400080000000400000002000000000000000000f03f0000000000000040`}},
				},
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 'SRID=4326;POINT Z (1 2 3)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`66676203666762013c0000002000000000001a000c00000000000400050000000000000000000000060008001a000000010100000c000000080008000000040008000000e6100000540000001000000000000000000006000a000400060000001000000000000a000c000000040008000a00000018000000040000000100000000000000000008400000000002000000000000000000f03f0000000000000040`}},
				},
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q, true), 'hex') FROM (SELECT 1 AS id, 'SRID=4326;POINT ZM (1 2 3 4)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`66676203666762019c0000002000000000001a00200000000800050006000700000000000c001400000010001a00000000010101200000001400000048000000010000000000000000000000010000004400000004000000000000000000f03f0000000000000040000000000000f03f000000000000004000000000080008000000040008000000e610000008000c00080007000800000000000005040000000200000069640000000000000000f03f0000000000000040000000000000f03f00000000000000400100000000000000000000000000f03f0000000000000040000000000000f03f00000000000000400000000000000000740000000c00000008000c000800040008000000080000001c0000000600000000000100000000000c0010000000040008000c000c0000002c0000001800000004000000010000000000000000001040000000000100000000000000000008400000000002000000000000000000f03f0000000000000040`}},
				},
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 'POLYGON((0 0,4 0,4 4,0 0),(1 1,2 1,2 2,1 1))'::geometry AS geom) q;`,
					Expected: []sql.Row{{`6667620366676201240000001c0000001800080000000000050000000000000000000000000006001800000000030000bc0000000c0000000000060008000400060000000c00000008000c00040008000800000090000000040000001000000000000000000000000000000000000000000000000000104000000000000000000000000000001040000000000000104000000000000000000000000000000000000000000000f03f000000000000f03f0000000000000040000000000000f03f00000000000000400000000000000040000000000000f03f000000000000f03f00000000020000000400000008000000`}},
				},
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 'MULTIPOLYGON(((0 0,4 0,4 4,0 0)),((5 5,6 5,6 6,5 5)))'::geometry AS geom) q;`,
					Expected: []sql.Row{{`6667620366676201240000001c0000001800080000000000050000000000000000000000000006001800000000060000fc00000010000000000000000000060008000400060000001800000014000c0000000000000000000000000007000800140000000000000604000000020000007000000004000000aaffffff0000000304000000080000000000000000001440000000000000144000000000000018400000000000001440000000000000184000000000000018400000000000001440000000000000144000000000000012000c0000000800000000000000000007001200000000000003040000000800000000000000000000000000000000000000000000000000104000000000000000000000000000001040000000000000104000000000000000000000000000000000`}},
				},
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 'POINT M (1 2 4)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`6667620366676201240000001c0000001800080000000000040000000500000000000000000006001800000001010000540000001000000000000000000006000800040006000000100000000c000c0000000400000008000c00000018000000040000000100000000000000000010400000000002000000000000000000f03f0000000000000040`}},
				},
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 'MULTILINESTRING((0 0,1 1),(2 2,3 3))'::geometry AS geom) q;`,
					Expected: []sql.Row{{`6667620366676201240000001c00000018000800000000000500000000000000000000000000060018000000000500007c0000000c0000000000060008000400060000000c00000008000c00040008000800000050000000040000000800000000000000000000000000000000000000000000000000f03f000000000000f03f000000000000004000000000000000400000000000000840000000000000084000000000020000000200000004000000`}},
				},
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 'GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 0,1 1))'::geometry AS geom) q;`,
					Expected: []sql.Row{{`6667620366676201240000001c0000001800080000000000050000000000000000000000000006001800000000070000ac00000010000000000000000000060008000400060000001800000014000c0000000000000000000000000007000800140000000000000704000000020000005000000004000000caffffff00000002040000000400000000000000000000000000000000000000000000000000f03f000000000000f03f00000000000012000c00000008000000000000000000070012000000000000010400000002000000000000000000f03f000000000000f03f`}},
				},
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q, false, 'g'), 'hex') FROM (SELECT 'LINESTRING(0 0,1 1)'::geometry AS g) q;`,
					Expected: []sql.Row{{`6667620366676201240000001c00000018000800000000000500000000000000000000000000060018000000000200004c00000010000000000000000000060008000400060000000c000000080008000000040008000000040000000400000000000000000000000000000000000000000000000000f03f000000000000f03f`}},
				},
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 'MULTIPOINT((1 1),(2 2))'::geometry AS geom) q;`,
					Expected: []sql.Row{{`6667620366676201240000001c00000018000800000000000500000000000000000000000000060018000000000400004c00000010000000000000000000060008000400060000000c0000000800080000000400080000000400000004000000000000000000f03f000000000000f03f00000000000000400000000000000040`}},
				},
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 1::int2 AS a, 2::int8 AS b, 1.5::float4 AS c, 2.5::float8 AS d, true AS e, 'xy'::text AS f, '2024-01-02 03:04:05+00'::timestamptz AS i, NULL::int AS m, 'POINT(1 1)'::geometry AS geom) q;`,
					Expected: []sql.Row{{`6667620366676201f40000001c00000018000c00000000000500000000000000000008000000060018000000000100000400000008000000b4000000940000007c000000640000004c000000340000001c0000000400000074ffffff0000000504000000010000006d00000088ffffff0000000d0400000001000000690000009cffffff0000000b040000000100000066000000b0ffffff00000002040000000100000065000000c4ffffff0000000a040000000100000064000000d8ffffff00000009040000000100000063000000ecffffff0000000704000000010000006200000008000c000800070008000000000000030400000001000000610000008c000000100000000000000008000c000800040008000000080000005800000045000000000001000100020000000000000002000000c03f030000000000000004400400010500020000007879060016000000323032342d30312d30322030333a30343a30352b30300000000800080000000400080000000400000002000000000000000000f03f000000000000f03f`}},
				},
				{
					Query:    `SELECT ST_AsFlatGeobuf(q) IS NULL FROM (SELECT 'POINT(1 1)'::geometry AS geom WHERE false) q;`,
					Expected: []sql.Row{{`t`}},
				},
				{
					Query:           `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 'CIRCULARSTRING(0 0,1 1,2 0)'::geometry AS geom) q;`,
					ExpectedErr:     `flatgeobuf: get_geometrytype: 'CircularString' geometry type not supported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 'TRIANGLE((0 0,1 0,0 1,0 0))'::geometry AS geom) q;`,
					ExpectedErr:     `flatgeobuf: GeometryWriter::write: 'Triangle' geometry type not supported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 'POINT EMPTY'::geometry AS geom) q;`,
					Expected: []sql.Row{{`6667620366676201240000001c00000018000800000000000500000000000000000000000000060018000000000100000c000000080000000400040004000000`}},
				},
				{
					Query:           `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 1 AS a) q;`,
					ExpectedErr:     `no geom column found`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 'POINT(1 1)'::geometry AS geom UNION ALL SELECT NULL::geometry) q;`,
					Expected: []sql.Row{{`6667620366676201240000001c00000018000800000000000500000000000000000000000000060018000000000100003c00000010000000000000000000060008000400060000000c0000000800080000000400080000000400000002000000000000000000f03f000000000000f03f0c000000080000000400040004000000`}},
				},
				{
					Query:           `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT NULL::geometry AS geom UNION ALL SELECT 'POINT(1 1)'::geometry) q;`,
					ExpectedErr:     `mixed geometry type is not supported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 'POINT(1 1)'::geometry AS geom UNION ALL SELECT 'LINESTRING(0 0,1 1)'::geometry) q;`,
					ExpectedErr:     `mixed geometry type is not supported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT encode(ST_AsFlatGeobuf(q), 'hex') FROM (SELECT 3.25::numeric AS n, 'POINT(1 1)'::geometry AS geom) q;`,
					ExpectedErr:     `flatgeobuf: get_column_type: '1700' column type not supported`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT encode(ST_AsFlatGeobuf(q, true), 'hex') FROM (SELECT i AS id, ST_MakePoint(i % 7, i / 7) AS geom FROM generate_series(1, 40) i) q;`,
					Expected: []sql.Row{{`6667620366676201840000001c000000000016001c0000000800070000000000000000000c00100016000000000000011c0000001000000028000000000000000000000001000000340000000400000000000000000000000000000000000000000000000000184000000000000014400000000008000c0008000700080000000000000504000000020000006964000000000000000000000000000000000000000000000000184000000000000014400100000000000000000000000000104000000000000000000000000000001840000000000000144004000000000000000000000000000000000000000000004000000000000010400000000000001440140000000000000000000000000000000000000000000000000000000000084000000000000000402400000000000000000000000000184000000000000000000000000000001840000000000000000000000000000000000000000000001840000000000000f03f0000000000001840000000000000f03f50000000000000000000000000001440000000000000f03f0000000000001440000000000000f03fa0000000000000000000000000001440000000000000000000000000000014400000000000000000f000000000000000000000000000104000000000000000000000000000001040000000000000000040010000000000000000000000001040000000000000f03f0000000000001040000000000000f03f90010000000000000000000000001040000000000000004000000000000010400000000000000040e00100000000000000000000000014400000000000000040000000000000144000000000000000403002000000000000000000000000184000000000000000400000000000001840000000000000004080020000000000000000000000001840000000000000084000000000000018400000000000000840d00200000000000000000000000014400000000000000840000000000000144000000000000008402003000000000000000000000000184000000000000010400000000000001840000000000000104070030000000000000000000000001440000000000000144000000000000014400000000000001440c00300000000000000000000000014400000000000001040000000000000144000000000000010401004000000000000000000000000104000000000000010400000000000001040000000000000104060040000000000000000000000001040000000000000144000000000000010400000000000001440b00400000000000000000000000010400000000000000840000000000000104000000000000008400005000000000000000000000000084000000000000008400000000000000840000000000000084050050000000000000000000000000040000000000000084000000000000000400000000000000840a0050000000000000000000000000840000000000000104000000000000008400000000000001040f00500000000000000000000000008400000000000001440000000000000084000000000000014404006000000000000000000000000004000000000000014400000000000000040000000000000144090060000000000000000000000000040000000000000104000000000000000400000000000001040e006000000000000000000000000f03f0000000000001040000000000000f03f00000000000010403007000000000000000000000000f03f0000000000001440000000000000f03f000000000000144080070000000000000000000000000000000000000000144000000000000000000000000000001440d00700000000000000000000000000000000000000001040000000000000000000000000000010402008000000000000000000000000f03f0000000000000840000000000000f03f000000000000084070080000000000000000000000000000000000000000084000000000000000000000000000000840c00800000000000000000000000000000000000000000040000000000000000000000000000000401009000000000000000000000000f03f0000000000000040000000000000f03f000000000000004060090000000000000000000000000040000000000000004000000000000000400000000000000040b0090000000000000000000000000840000000000000004000000000000008400000000000000040000a0000000000000000000000000040000000000000f03f0000000000000040000000000000f03f500a0000000000000000000000000840000000000000f03f0000000000000840000000000000f03fa00a0000000000000000000000000840000000000000000000000000000008400000000000000000f00a0000000000000000000000000040000000000000000000000000000000400000000000000000400b000000000000000000000000f03f0000000000000000000000000000f03f0000000000000000900b000000000000000000000000f03f000000000000f03f000000000000f03f000000000000f03fe00b0000000000000000000000000000000000000000f03f0000000000000000000000000000f03f300c0000000000004c000000100000000000000008000c00080004000800000008000000180000000600000000000600000000000800080000000400080000000400000002000000000000000000184000000000000000004c000000100000000000000008000c00080004000800000008000000180000000600000000000d000000000008000800000004000800000004000000020000000000000000001840000000000000f03f4c000000100000000000000008000c00080004000800000008000000180000000600000000000c000000000008000800000004000800000004000000020000000000000000001440000000000000f03f4c000000100000000000000008000c00080004000800000008000000180000000600000000000500000000000800080000000400080000000400000002000000000000000000144000000000000000004c000000100000000000000008000c00080004000800000008000000180000000600000000000400000000000800080000000400080000000400000002000000000000000000104000000000000000004c000000100000000000000008000c00080004000800000008000000180000000600000000000b000000000008000800000004000800000004000000020000000000000000001040000000000000f03f4c000000100000000000000008000c00080004000800000008000000180000000600000000001200000000000800080000000400080000000400000002000000000000000000104000000000000000404c000000100000000000000008000c00080004000800000008000000180000000600000000001300000000000800080000000400080000000400000002000000000000000000144000000000000000404c000000100000000000000008000c00080004000800000008000000180000000600000000001400000000000800080000000400080000000400000002000000000000000000184000000000000000404c000000100000000000000008000c00080004000800000008000000180000000600000000001b00000000000800080000000400080000000400000002000000000000000000184000000000000008404c000000100000000000000008000c00080004000800000008000000180000000600000000001a00000000000800080000000400080000000400000002000000000000000000144000000000000008404c000000100000000000000008000c00080004000800000008000000180000000600000000002200000000000800080000000400080000000400000002000000000000000000184000000000000010404c000000100000000000000008000c00080004000800000008000000180000000600000000002800000000000800080000000400080000000400000002000000000000000000144000000000000014404c000000100000000000000008000c00080004000800000008000000180000000600000000002100000000000800080000000400080000000400000002000000000000000000144000000000000010404c000000100000000000000008000c00080004000800000008000000180000000600000000002000000000000800080000000400080000000400000002000000000000000000104000000000000010404c000000100000000000000008000c00080004000800000008000000180000000600000000002700000000000800080000000400080000000400000002000000000000000000104000000000000014404c000000100000000000000008000c00080004000800000008000000180000000600000000001900000000000800080000000400080000000400000002000000000000000000104000000000000008404c000000100000000000000008000c00080004000800000008000000180000000600000000001800000000000800080000000400080000000400000002000000000000000000084000000000000008404c000000100000000000000008000c00080004000800000008000000180000000600000000001700000000000800080000000400080000000400000002000000000000000000004000000000000008404c000000100000000000000008000c00080004000800000008000000180000000600000000001f00000000000800080000000400080000000400000002000000000000000000084000000000000010404c000000100000000000000008000c00080004000800000008000000180000000600000000002600000000000800080000000400080000000400000002000000000000000000084000000000000014404c000000100000000000000008000c00080004000800000008000000180000000600000000002500000000000800080000000400080000000400000002000000000000000000004000000000000014404c000000100000000000000008000c00080004000800000008000000180000000600000000001e00000000000800080000000400080000000400000002000000000000000000004000000000000010404c000000100000000000000008000c00080004000800000008000000180000000600000000001d00000000000800080000000400080000000400000002000000000000000000f03f00000000000010404c000000100000000000000008000c00080004000800000008000000180000000600000000002400000000000800080000000400080000000400000002000000000000000000f03f00000000000014404c000000100000000000000008000c00080004000800000008000000180000000600000000002300000000000800080000000400080000000400000002000000000000000000000000000000000014404c000000100000000000000008000c00080004000800000008000000180000000600000000001c00000000000800080000000400080000000400000002000000000000000000000000000000000010404c000000100000000000000008000c00080004000800000008000000180000000600000000001600000000000800080000000400080000000400000002000000000000000000f03f00000000000008404c000000100000000000000008000c00080004000800000008000000180000000600000000001500000000000800080000000400080000000400000002000000000000000000000000000000000008404c000000100000000000000008000c00080004000800000008000000180000000600000000000e00000000000800080000000400080000000400000002000000000000000000000000000000000000404c000000100000000000000008000c00080004000800000008000000180000000600000000000f00000000000800080000000400080000000400000002000000000000000000f03f00000000000000404c000000100000000000000008000c00080004000800000008000000180000000600000000001000000000000800080000000400080000000400000002000000000000000000004000000000000000404c000000100000000000000008000c00080004000800000008000000180000000600000000001100000000000800080000000400080000000400000002000000000000000000084000000000000000404c000000100000000000000008000c000800040008000000080000001800000006000000000009000000000008000800000004000800000004000000020000000000000000000040000000000000f03f4c000000100000000000000008000c00080004000800000008000000180000000600000000000a000000000008000800000004000800000004000000020000000000000000000840000000000000f03f4c000000100000000000000008000c00080004000800000008000000180000000600000000000300000000000800080000000400080000000400000002000000000000000000084000000000000000004c000000100000000000000008000c00080004000800000008000000180000000600000000000200000000000800080000000400080000000400000002000000000000000000004000000000000000004c000000100000000000000008000c00080004000800000008000000180000000600000000000100000000000800080000000400080000000400000002000000000000000000f03f00000000000000004c000000100000000000000008000c00080004000800000008000000180000000600000000000800000000000800080000000400080000000400000002000000000000000000f03f000000000000f03f4c000000100000000000000008000c000800040008000000080000001800000006000000000007000000000008000800000004000800000004000000020000000000000000000000000000000000f03f`}},
				},
				{
					Query: `SELECT id, ST_AsEWKT(geom), num, name FROM ST_FromFlatGeobuf(null::fgbt, (SELECT ST_AsFlatGeobuf(q) FROM (SELECT 5 AS num, 'a' AS name, 'SRID=4326;POINT(1 2)'::geometry AS geom UNION ALL SELECT 7, 'b', 'SRID=4326;POINT(3 4)'::geometry) q));`,
					Expected: []sql.Row{
						{0, `SRID=4326;POINT(1 2)`, 5, `a`},
						{1, `SRID=4326;POINT(3 4)`, 7, `b`},
					},
				},
				{
					Query: `SELECT id, ST_AsText(geom) FROM ST_FromFlatGeobuf(null::fgbt, (SELECT ST_AsFlatGeobuf(q, true) FROM (SELECT 5 AS num, 'a' AS name, 'POLYGON((0 0,4 0,4 4,0 0),(1 1,2 1,2 2,1 1))'::geometry AS geom UNION ALL SELECT 7, 'b', 'POLYGON((10 10,14 10,14 14,10 10))'::geometry) q));`,
					Expected: []sql.Row{
						{0, `POLYGON((10 10,14 10,14 14,10 10))`},
						{1, `POLYGON((0 0,4 0,4 4,0 0),(1 1,2 1,2 2,1 1))`},
					},
				},
				{
					Query:           `SELECT * FROM ST_FromFlatGeobuf(null::fgbt, '\x00'::bytea);`,
					ExpectedErr:     `Data is not FlatGeobuf`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT count(*) FROM ST_FromFlatGeobuf(null::fgbt, ''::bytea);`,
					Expected: []sql.Row{{0}},
				},
				{
					Query:           `SELECT * FROM ST_FromFlatGeobuf(null::fgbt, (SELECT ST_AsFlatGeobuf(q) FROM (SELECT 5 AS num, 'POINT(1 2)'::geometry AS geom) q));`,
					ExpectedErr:     `flatgeobuf: column count mismatch: file has 1 columns, target type has 2`,
					ExpectedErrCode: `42804`,
				},
				{
					Query:           `SELECT * FROM ST_FromFlatGeobuf(null::fgbt, (SELECT ST_AsFlatGeobuf(q) FROM (SELECT 'x' AS num, 'y' AS name, 'POINT(1 2)'::geometry AS geom) q));`,
					ExpectedErr:     `flatgeobuf: column "num" type mismatch: file type "String" is not compatible with PostgreSQL type integer`,
					ExpectedErrCode: `42804`,
				},
				{
					Query: `SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'fgb_new' ORDER BY ordinal_position;`,
					Expected: []sql.Row{
						{`id`, `integer`},
						{`geom`, `USER-DEFINED`},
						{`num`, `integer`},
						{`name`, `text`},
						{`f`, `double precision`},
						{`b`, `boolean`},
						{`s`, `smallint`},
						{`l`, `bigint`},
						{`r`, `real`},
					},
					Skip: true,
				},
				{
					Query:    `SELECT count(*) FROM fgb_new;`,
					Expected: []sql.Row{{0}},
				},
				{
					Query:           `SELECT ST_FromFlatGeobufToTable('public', 'fgb_dup', (SELECT ST_AsFlatGeobuf(q) FROM (SELECT 1 AS id, 'POINT(1 2)'::geometry AS geom) q));`,
					ExpectedErr:     "duplicate column name: `id`",
					ExpectedErrCode: `42701`,
				},
				{
					Query:           `SELECT ST_FromFlatGeobufToTable('public', 'fgb_dup', (SELECT ST_AsFlatGeobuf(q) FROM (SELECT 1 AS id, 'POINT(1 2)'::geometry AS geom) q));`,
					ExpectedErr:     `column "id" specified more than once`,
					ExpectedErrCode: `42701`,
					Skip:            true,
				},
			},
		},
		{
			Name: `postgis TWKB`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT encode(ST_AsTWKB('POINT(1 2)'::geometry), 'hex'), encode(ST_AsTWKB('POINT(1.234 2.567)'::geometry, 2), 'hex'), encode(ST_AsTWKB('SRID=4326;POINT(1.123456 2.654321)'::geometry), 'hex');`,
					Expected: []sql.Row{{`01000204`, `4100f6018204`, `a100b4db0db0b320`}},
				},
				{
					Query:    `SELECT encode(ST_AsTWKB('LINESTRING(0 0,1 1,2 0)'::geometry), 'hex'), encode(ST_AsTWKB('LINESTRING(0 0,1 1,1 1,2 0)'::geometry), 'hex'), encode(ST_AsTWKB('LINESTRING(0 0,0.001 0,1 1)'::geometry, 0), 'hex');`,
					Expected: []sql.Row{{`020003000002020201`, `020003000002020201`, `02000200000202`}},
				},
				{
					Query:    `SELECT encode(ST_AsTWKB('POLYGON((0 0,10 0,10 10,0 10,0 0),(2 2,2 4,4 4,4 2,2 2))'::geometry), 'hex');`,
					Expected: []sql.Row{{`03000205000014000014130000130504040004040000030300`}},
				},
				{
					Query:    `SELECT encode(ST_AsTWKB('MULTIPOINT((1 1),(2 2))'::geometry), 'hex'), encode(ST_AsTWKB('MULTILINESTRING((0 0,1 1),(5 5,6 6))'::geometry), 'hex'), encode(ST_AsTWKB('MULTIPOLYGON(((0 0,4 0,4 4,0 0)),((5 5,6 5,6 6,5 5)))'::geometry), 'hex');`,
					Expected: []sql.Row{{`04000202020202`, `05000202000002020208080202`, `0600020104000008000008070701040a0a020000020101`}},
				},
				{
					Query:    `SELECT encode(ST_AsTWKB('GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 0,1 1))'::geometry), 'hex'), encode(ST_AsTWKB('GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 0,1 1))'::geometry, 0, 0, 0, true, true), 'hex');`,
					Expected: []sql.Row{{`0700020100020202000200000202`, `07031a0002000202010306020002000202020309000200020200000202`}},
				},
				{
					Query:    `SELECT encode(ST_AsTWKB('POINT Z (1 2 3)'::geometry), 'hex'), encode(ST_AsTWKB('POINT M (1 2 3)'::geometry), 'hex'), encode(ST_AsTWKB('POINT ZM (1.5 2.5 3.25 4.125)'::geometry, 1, 2, 3), 'hex'), encode(ST_AsTWKB('LINESTRING Z (0 0 0,1 1 1)'::geometry, 0, 1), 'hex');`,
					Expected: []sql.Row{{`010801020406`, `010802020406`, `21086b1e328a05ba40`, `02080502000000020214`}},
				},
				{
					Query:    `SELECT encode(ST_AsTWKB('LINESTRING(0 0,1 1,2 0)'::geometry, 0, 0, 0, true), 'hex'), encode(ST_AsTWKB('LINESTRING(0 0,1 1,2 0)'::geometry, 0, 0, 0, false, true), 'hex'), encode(ST_AsTWKB('LINESTRING(0 0,1 1,2 0)'::geometry, 0, 0, 0, true, true), 'hex');`,
					Expected: []sql.Row{{`02020703000002020201`, `02010004000203000002020201`, `02030b0004000203000002020201`}},
				},
				{
					Query:    `SELECT encode(ST_AsTWKB('POLYGON((0 0,10 0,10 10,0 0))'::geometry, 0, 0, 0, true, true), 'hex'), encode(ST_AsTWKB('MULTIPOINT((1 1),(2 2))'::geometry, 0, 0, 0, true, true), 'hex');`,
					Expected: []sql.Row{{`03030e0014001401040000140000141313`, `040309020202020202020202`}},
				},
				{
					Query:    `SELECT encode(ST_AsTWKB('POINT EMPTY'::geometry), 'hex'), encode(ST_AsTWKB('LINESTRING EMPTY'::geometry, 0, 0, 0, true), 'hex'), encode(ST_AsTWKB('POINT Z EMPTY'::geometry), 'hex'), encode(ST_AsTWKB('GEOMETRYCOLLECTION EMPTY'::geometry), 'hex');`,
					Expected: []sql.Row{{`0110`, `021200`, `011801`, `0710`}},
				},
				{
					Query:    `SELECT encode(ST_AsTWKB('POINT(1234.5 -10)'::geometry, -2), 'hex'), encode(ST_AsTWKB('POINT(1 2)'::geometry, 7), 'hex');`,
					Expected: []sql.Row{{`31001800`, `e10080dac40980b48913`}},
				},
				{
					Query:    `SELECT encode(ST_AsTWKB(ARRAY['POINT(1 1)'::geometry, 'POINT(2 2)'::geometry], ARRAY[10, 20]::bigint[]), 'hex'), encode(ST_AsTWKB(ARRAY['POINT(1 1)'::geometry, 'LINESTRING(0 0,1 1)'::geometry], ARRAY[1, -2]::bigint[]), 'hex');`,
					Expected: []sql.Row{{`040402142802020202`, `07040202030100020202000200000202`}},
				},
				{
					Query:    `SELECT encode(ST_AsTWKB(ARRAY['LINESTRING(0 0,1 1)'::geometry, 'LINESTRING(5 5,6 6)'::geometry], ARRAY[1, 2]::bigint[], 0, 0, 0, true, true), 'hex');`,
					Expected: []sql.Row{{`050711000c000c02020402000002020208080202`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_GeomFromTWKB(ST_AsTWKB('POLYGON((0 0,10 0,10 10,0 10,0 0),(2 2,2 4,4 4,4 2,2 2))'::geometry))), ST_AsText(ST_GeomFromTWKB(ST_AsTWKB('POINT(1.234 2.567)'::geometry, 2)));`,
					Expected: []sql.Row{{`POLYGON((0 0,10 0,10 10,0 10,0 0),(2 2,2 4,4 4,4 2,2 2))`, `POINT(1.23 2.57)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_GeomFromTWKB(ST_AsTWKB('GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 0,1 1))'::geometry, 0, 0, 0, true, true))), ST_AsText(ST_GeomFromTWKB(ST_AsTWKB('POINT ZM (1.5 2.5 3.25 4.125)'::geometry, 1, 2, 3)));`,
					Expected: []sql.Row{{`GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 0,1 1))`, `POINT ZM (1.5 2.5 3.25 4.125)`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_GeomFromTWKB(ST_AsTWKB('MULTIPOLYGON(((0 0,4 0,4 4,0 0)),((5 5,6 5,6 6,5 5)))'::geometry))), ST_AsText(ST_GeomFromTWKB(ST_AsTWKB('POINT EMPTY'::geometry))), ST_AsText(ST_GeomFromTWKB(ST_AsTWKB(ARRAY['POINT(1 1)'::geometry, 'POINT(2 2)'::geometry], ARRAY[10, 20]::bigint[])));`,
					Expected: []sql.Row{{`MULTIPOLYGON(((0 0,4 0,4 4,0 0)),((5 5,6 5,6 6,5 5)))`, `POINT EMPTY`, `MULTIPOINT((1 1),(2 2))`}},
				},
				{
					Query:           `SELECT ST_AsText(ST_GeomFromTWKB('\x01'::bytea));`,
					ExpectedErr:     `twkb_parse_state_advance: TWKB structure does not match expected size!`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT ST_AsTWKB('CIRCULARSTRING(0 0,1 1,2 0)'::geometry);`,
					ExpectedErr:     `lwgeom_twkb_type: Unsupported geometry type: CircularString`,
					ExpectedErrCode: `XX000`,
				},
			},
		},
		{
			Name: `postgis letters`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
				`CREATE FUNCTION tf(g text) RETURNS text AS $$ SELECT encode(ST_AsTWKB(g::geometry), 'base64') $$ LANGUAGE sql;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT pg_get_function_arguments(oid), proisstrict FROM pg_proc WHERE proname = 'st_letters';`,
					Expected: []sql.Row{{`letters text, font json DEFAULT NULL::json`, `f`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Letters('ab', json_build_object('m', tf('POLYGON((0 0,600 0,600 500,0 500,0 0))'), 'a', tf('POLYGON((10 0,410 0,410 500,10 500,10 0))'), 'b', tf('POLYGON((0 0,300 0,0 700,0 0))'))));`,
					Expected: []sql.Row{{`MULTIPOLYGON(((0 0,40 0,40 50,0 50,0 0)),((45 0,75 0,45 70,45 0)))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Letters('a b', json_build_object('m', tf('POLYGON((0 0,600 0,600 500,0 500,0 0))'), 'a', tf('POLYGON((10 0,410 0,410 500,10 500,10 0))'), 'b', tf('POLYGON((0 0,300 0,0 700,0 0))'))));`,
					Expected: []sql.Row{{`MULTIPOLYGON(((0 0,40 0,40 50,0 50,0 0)),((67.14285714285715 0,97.14285714285715 0,67.14285714285715 70,67.14285714285715 0)))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Letters('azb', json_build_object('m', tf('POLYGON((0 0,600 0,600 500,0 500,0 0))'), 'a', tf('POLYGON((10 0,410 0,410 500,10 500,10 0))'), 'b', tf('POLYGON((0 0,300 0,0 700,0 0))'))));`,
					Expected: []sql.Row{{`MULTIPOLYGON(((0 0,40 0,40 50,0 50,0 0)),((67.14285714285715 0,97.14285714285715 0,67.14285714285715 70,67.14285714285715 0)))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Letters('ba', json_build_object('m', tf('POLYGON((0 0,600 0,600 500,0 500,0 0))'), 'a', tf('POLYGON((10 0,410 0,410 500,10 500,10 0))'), 'b', tf('POLYGON((0 0,300 0,0 700,0 0))'))));`,
					Expected: []sql.Row{{`MULTIPOLYGON(((0 0,30 0,0 70,0 0)),((35 0,75 0,75 50,35 50,35 0)))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_Letters('cmc', json_build_object('m', tf('POLYGON((0 0,600 0,600 500,0 500,0 0))'), 'c', tf('MULTIPOLYGON(((0 0,100 0,0 100,0 0)),((200 200,300 200,200 300,200 200)))'))));`,
					Expected: []sql.Row{{`MULTIPOLYGON(((0 0,10 0,0 10,0 0)),((20 20,30 20,20 30,20 20)),((35 0,95 0,95 50,35 50,35 0)),((100 0,110 0,100 10,100 0)),((120 20,130 20,120 30,120 20)))`}},
				},
				{
					Query:    `SELECT ST_Letters('', json_build_object('m', tf('POLYGON((0 0,600 0,600 500,0 500,0 0))'))) IS NULL, ST_Letters('zz', json_build_object('m', tf('POLYGON((0 0,600 0,600 500,0 500,0 0))'))) IS NULL, ST_Letters('a', '[1]'::json) IS NULL;`,
					Expected: []sql.Row{{`t`, `t`, `t`}},
				},
				{
					Query:           `SELECT ST_Letters('a', json_build_object('m', 'xyz'));`,
					ExpectedErr:     `invalid base64 end sequence`,
					ExpectedErrCode: `22023`,
				},
				{
					Query:           `SELECT ST_Letters('m', '{"m": 5}'::json);`,
					ExpectedErr:     `invalid base64 end sequence`,
					ExpectedErrCode: `22023`,
				},
				{
					Query:           `SELECT ST_Letters(NULL);`,
					ExpectedErr:     `FOREACH expression must not be null`,
					ExpectedErrCode: `22004`,
				},
				{
					Query:    `SELECT ST_NumGeometries(ST_Letters('PostGIS')), ST_GeometryType(ST_Letters('PostGIS')), ST_IsValid(ST_Letters('PostGIS')), ST_XMin(ST_Letters('PostGIS')), ST_YMin(ST_Letters('PostGIS'));`,
					Expected: []sql.Row{{7, `ST_MultiPolygon`, `t`, 0.0, -0.9}},
				},
				{
					Query:    `SELECT ST_Letters('') IS NULL, ST_Letters('   ') IS NULL, ST_NumGeometries(ST_Letters('a b')) = ST_NumGeometries(ST_Letters('ab'));`,
					Expected: []sql.Row{{`t`, `t`, `t`}},
				},
				{
					Query: "SELECT c, round(ST_XMax(ST_Letters(c))::numeric, 6), round(ST_YMin(ST_Letters(c))::numeric, 6), round(ST_YMax(ST_Letters(c))::numeric, 6) FROM unnest(string_to_array($$!\"&'()+,-./0123456789:;=?ABCDEFGHIJKLMNOPQRSTUVWXYZ`abcdefghijklmnopqrstuvwxyz$$, NULL)) c;",
					Expected: []sql.Row{
						{`!`, framework.Numeric(`18.800000`), framework.Numeric(`-0.500000`), framework.Numeric(`73.700000`)},
						{`"`, framework.Numeric(`31.200000`), framework.Numeric(`47.100000`), framework.Numeric(`73.700000`)},
						{`&`, framework.Numeric(`46.000000`), framework.Numeric(`-0.700000`), framework.Numeric(`75.100000`)},
						{`'`, framework.Numeric(`13.000000`), framework.Numeric(`47.100000`), framework.Numeric(`73.700000`)},
						{`(`, framework.Numeric(`21.900000`), framework.Numeric(`-5.100000`), framework.Numeric(`75.000000`)},
						{`)`, framework.Numeric(`21.600000`), framework.Numeric(`-5.100000`), framework.Numeric(`75.000000`)},
						{`+`, framework.Numeric(`29.000000`), framework.Numeric(`14.700000`), framework.Numeric(`43.800000`)},
						{`,`, framework.Numeric(`18.000000`), framework.Numeric(`-9.700000`), framework.Numeric(`17.000000`)},
						{`-`, framework.Numeric(`29.000000`), framework.Numeric(`23.200000`), framework.Numeric(`38.200000`)},
						{`.`, framework.Numeric(`18.400000`), framework.Numeric(`-0.400000`), framework.Numeric(`17.600000`)},
						{`/`, framework.Numeric(`34.600000`), framework.Numeric(`0.100000`), framework.Numeric(`73.700000`)},
						{`0`, framework.Numeric(`42.000000`), framework.Numeric(`-1.000000`), framework.Numeric(`74.900000`)},
						{`1`, framework.Numeric(`26.700000`), framework.Numeric(`0.000000`), framework.Numeric(`73.700000`)},
						{`2`, framework.Numeric(`42.400000`), framework.Numeric(`0.000000`), framework.Numeric(`74.800000`)},
						{`3`, framework.Numeric(`41.600000`), framework.Numeric(`-0.600000`), framework.Numeric(`74.700000`)},
						{`4`, framework.Numeric(`47.300000`), framework.Numeric(`0.000000`), framework.Numeric(`73.700000`)},
						{`5`, framework.Numeric(`38.700000`), framework.Numeric(`-0.400000`), framework.Numeric(`73.700000`)},
						{`6`, framework.Numeric(`39.600000`), framework.Numeric(`-0.600000`), framework.Numeric(`74.600000`)},
						{`7`, framework.Numeric(`40.800000`), framework.Numeric(`0.000000`), framework.Numeric(`73.700000`)},
						{`8`, framework.Numeric(`40.600000`), framework.Numeric(`-1.000000`), framework.Numeric(`74.800000`)},
						{`9`, framework.Numeric(`39.600000`), framework.Numeric(`-0.600000`), framework.Numeric(`74.600000`)},
						{`:`, framework.Numeric(`18.400000`), framework.Numeric(`-0.400000`), framework.Numeric(`48.500000`)},
						{`;`, framework.Numeric(`18.400000`), framework.Numeric(`-9.700000`), framework.Numeric(`48.500000`)},
						{`=`, framework.Numeric(`29.000000`), framework.Numeric(`19.500000`), framework.Numeric(`46.700000`)},
						{`?`, framework.Numeric(`39.200000`), framework.Numeric(`-0.400000`), framework.Numeric(`75.000000`)},
						{`A`, framework.Numeric(`55.800000`), framework.Numeric(`0.000000`), framework.Numeric(`73.700000`)},
						{`B`, framework.Numeric(`43.500000`), framework.Numeric(`0.000000`), framework.Numeric(`73.800000`)},
						{`C`, framework.Numeric(`36.400000`), framework.Numeric(`0.000000`), framework.Numeric(`73.800000`)},
						{`D`, framework.Numeric(`43.200000`), framework.Numeric(`0.000000`), framework.Numeric(`73.800000`)},
						{`E`, framework.Numeric(`36.300000`), framework.Numeric(`-0.900000`), framework.Numeric(`74.700000`)},
						{`F`, framework.Numeric(`35.200000`), framework.Numeric(`0.000000`), framework.Numeric(`74.700000`)},
						{`G`, framework.Numeric(`41.300000`), framework.Numeric(`-0.700000`), framework.Numeric(`74.000000`)},
						{`H`, framework.Numeric(`45.600000`), framework.Numeric(`0.000000`), framework.Numeric(`73.800000`)},
						{`I`, framework.Numeric(`18.300000`), framework.Numeric(`0.000000`), framework.Numeric(`73.800000`)},
						{`J`, framework.Numeric(`46.100000`), framework.Numeric(`-0.700000`), framework.Numeric(`73.800000`)},
						{`K`, framework.Numeric(`57.200000`), framework.Numeric(`-0.900000`), framework.Numeric(`73.800000`)},
						{`L`, framework.Numeric(`36.500000`), framework.Numeric(`0.000000`), framework.Numeric(`73.800000`)},
						{`M`, framework.Numeric(`54.400000`), framework.Numeric(`0.000000`), framework.Numeric(`74.700000`)},
						{`N`, framework.Numeric(`46.900000`), framework.Numeric(`-1.400000`), framework.Numeric(`74.500000`)},
						{`O`, framework.Numeric(`45.600000`), framework.Numeric(`-0.900000`), framework.Numeric(`74.900000`)},
						{`P`, framework.Numeric(`42.600000`), framework.Numeric(`0.000000`), framework.Numeric(`73.800000`)},
						{`Q`, framework.Numeric(`56.400000`), framework.Numeric(`-0.900000`), framework.Numeric(`74.900000`)},
						{`R`, framework.Numeric(`47.500000`), framework.Numeric(`0.000000`), framework.Numeric(`73.800000`)},
						{`S`, framework.Numeric(`42.000000`), framework.Numeric(`-0.400000`), framework.Numeric(`74.200000`)},
						{`T`, framework.Numeric(`44.200000`), framework.Numeric(`0.000000`), framework.Numeric(`73.800000`)},
						{`U`, framework.Numeric(`47.700000`), framework.Numeric(`-0.400000`), framework.Numeric(`73.800000`)},
						{`V`, framework.Numeric(`55.800000`), framework.Numeric(`0.000000`), framework.Numeric(`73.700000`)},
						{`W`, framework.Numeric(`83.600000`), framework.Numeric(`0.000000`), framework.Numeric(`73.800000`)},
						{`X`, framework.Numeric(`54.000000`), framework.Numeric(`0.000000`), framework.Numeric(`73.800000`)},
						{`Y`, framework.Numeric(`54.000000`), framework.Numeric(`0.000000`), framework.Numeric(`73.800000`)},
						{`Z`, framework.Numeric(`43.800000`), framework.Numeric(`0.000000`), framework.Numeric(`73.800000`)},
						{"`", framework.Numeric(`20.700000`), framework.Numeric(`62.300000`), framework.Numeric(`77.800000`)},
						{`a`, framework.Numeric(`42.200000`), framework.Numeric(`-0.300000`), framework.Numeric(`58.100000`)},
						{`b`, framework.Numeric(`42.400000`), framework.Numeric(`0.000000`), framework.Numeric(`73.600000`)},
						{`c`, framework.Numeric(`32.300000`), framework.Numeric(`-0.700000`), framework.Numeric(`58.400000`)},
						{`d`, framework.Numeric(`42.000000`), framework.Numeric(`-0.300000`), framework.Numeric(`73.600000`)},
						{`e`, framework.Numeric(`40.700000`), framework.Numeric(`-0.300000`), framework.Numeric(`57.800000`)},
						{`f`, framework.Numeric(`31.200000`), framework.Numeric(`0.000000`), framework.Numeric(`75.500000`)},
						{`g`, framework.Numeric(`41.400000`), framework.Numeric(`-24.900000`), framework.Numeric(`57.600000`)},
						{`h`, framework.Numeric(`43.100000`), framework.Numeric(`0.000000`), framework.Numeric(`73.600000`)},
						{`i`, framework.Numeric(`18.700000`), framework.Numeric(`0.000000`), framework.Numeric(`76.200000`)},
						{`j`, framework.Numeric(`26.300000`), framework.Numeric(`-24.000000`), framework.Numeric(`76.200000`)},
						{`k`, framework.Numeric(`45.600000`), framework.Numeric(`0.000000`), framework.Numeric(`73.100000`)},
						{`l`, framework.Numeric(`18.200000`), framework.Numeric(`0.000000`), framework.Numeric(`77.100000`)},
						{`m`, framework.Numeric(`68.500000`), framework.Numeric(`0.000000`), framework.Numeric(`57.800000`)},
						{`n`, framework.Numeric(`43.100000`), framework.Numeric(`0.000000`), framework.Numeric(`57.700000`)},
						{`o`, framework.Numeric(`41.400000`), framework.Numeric(`-0.500000`), framework.Numeric(`58.000000`)},
						{`p`, framework.Numeric(`42.300000`), framework.Numeric(`-24.900000`), framework.Numeric(`57.900000`)},
						{`q`, framework.Numeric(`42.300000`), framework.Numeric(`-24.900000`), framework.Numeric(`57.900000`)},
						{`r`, framework.Numeric(`38.500000`), framework.Numeric(`0.000000`), framework.Numeric(`57.800000`)},
						{`s`, framework.Numeric(`36.900000`), framework.Numeric(`-0.500000`), framework.Numeric(`57.900000`)},
						{`t`, framework.Numeric(`32.400000`), framework.Numeric(`-0.900000`), framework.Numeric(`75.400000`)},
						{`u`, framework.Numeric(`43.100000`), framework.Numeric(`-0.100000`), framework.Numeric(`57.600000`)},
						{`v`, framework.Numeric(`49.100000`), framework.Numeric(`0.000000`), framework.Numeric(`57.600000`)},
						{`w`, framework.Numeric(`66.000000`), framework.Numeric(`0.000000`), framework.Numeric(`57.600000`)},
						{`x`, framework.Numeric(`46.600000`), framework.Numeric(`0.000000`), framework.Numeric(`57.600000`)},
						{`y`, framework.Numeric(`43.300000`), framework.Numeric(`-24.900000`), framework.Numeric(`57.600000`)},
						{`z`, framework.Numeric(`41.100000`), framework.Numeric(`-0.600000`), framework.Numeric(`58.200000`)},
					},
				},
				{
					Query:    `SELECT ST_XMax(ST_Letters('m')), ST_YMax(ST_Letters('Hello')), ST_YMin(ST_Letters('gjpqy'));`,
					Expected: []sql.Row{{68.5, 77.10000000000001, -24.900000000000002}},
				},
			},
		},
		{
			Name: `postgis geometry columns without a type modifier`,
			Skip: true,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
				`CREATE TABLE nt (id INT PRIMARY KEY);`,
				`SELECT AddGeometryColumn('nt', 'g', 4326, 'POINT', 2, false);`,
				`CREATE TABLE np (id INT PRIMARY KEY, g geometry);`,
				`INSERT INTO np VALUES (1, 'SRID=3857;LINESTRING(0 0,1 1)');`,
				`SELECT populate_geometry_columns('np'::regclass, false);`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: `SELECT c.relname, format_type(a.atttypid, a.atttypmod) FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid WHERE c.relname IN ('nt', 'np') AND a.attname = 'g' ORDER BY 1;`,
					Expected: []sql.Row{
						{`np`, `geometry`},
						{`nt`, `geometry`},
					},
				},
				{
					Query: `SELECT f_table_name, f_geometry_column, coord_dimension, srid, type FROM geometry_columns WHERE f_table_name IN ('nt', 'np') ORDER BY 1;`,
					Expected: []sql.Row{
						{`np`, `g`, 2, 3857, `LINESTRING`},
						{`nt`, `g`, 2, 4326, `POINT`},
					},
				},
				{
					Query:    `SELECT postgis_constraint_srid('public', 'nt', 'g'), postgis_constraint_dims('public', 'nt', 'g'), postgis_constraint_type('public', 'nt', 'g');`,
					Expected: []sql.Row{{4326, 2, `POINT`}},
				},
				{
					Query:    `SELECT postgis_constraint_srid('public', 'np', 'g'), postgis_constraint_dims('public', 'np', 'g'), postgis_constraint_type('public', 'np', 'g');`,
					Expected: []sql.Row{{3857, 2, `LINESTRING`}},
				},
				{
					Query: `SELECT conname FROM pg_constraint WHERE conrelid IN ('nt'::regclass, 'np'::regclass) ORDER BY 1;`,
					Expected: []sql.Row{
						{`enforce_dims_g`},
						{`enforce_dims_g`},
						{`enforce_geotype_g`},
						{`enforce_geotype_g`},
						{`enforce_srid_g`},
						{`enforce_srid_g`},
						{`np_pkey`},
						{`nt_pkey`},
					},
				},
				{
					Query:           `INSERT INTO nt VALUES (1, 'SRID=3857;POINT(1 2)');`,
					ExpectedErr:     `new row for relation "nt" violates check constraint "enforce_srid_g"`,
					ExpectedErrCode: `23514`,
				},
				{
					Query:           `INSERT INTO nt VALUES (2, 'SRID=4326;LINESTRING(0 0,1 1)');`,
					ExpectedErr:     `new row for relation "nt" violates check constraint "enforce_geotype_g"`,
					ExpectedErrCode: `23514`,
				},
				{
					Query:           `INSERT INTO nt VALUES (3, 'SRID=4326;POINT Z (1 2 3)');`,
					ExpectedErr:     `new row for relation "nt" violates check constraint "enforce_dims_g"`,
					ExpectedErrCode: `23514`,
				},
				{
					Query:           `INSERT INTO np VALUES (2, 'SRID=4326;LINESTRING(0 0,1 1)');`,
					ExpectedErr:     `new row for relation "np" violates check constraint "enforce_srid_g"`,
					ExpectedErrCode: `23514`,
				},
				{
					Query:    `INSERT INTO nt VALUES (4, 'SRID=4326;POINT(1 2)');`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: `postgis geojson features`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT pg_get_function_arguments(oid), proisstrict FROM pg_proc WHERE proname = 'st_asgeojson' AND pg_get_function_arguments(oid) LIKE 'r %';`,
					Expected: []sql.Row{{`r record, geom_column text DEFAULT ''::text, maxdecimaldigits integer DEFAULT 9, pretty_bool boolean DEFAULT false, id_column text DEFAULT ''::text`, `t`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t) FROM (SELECT 1 AS id, 'a' AS name, 'POINT(1.123456789012 2)'::geometry AS geom) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1.123456789,2]}, "properties": {"id": 1, "name": "a"}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t, 'g', 2) FROM (SELECT 'POINT(1 2)'::geometry AS geom, 'SRID=4326;POINT(1.555 2)'::geometry AS g, 3.5::float8 AS f) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1.56,2]}, "properties": {"geom": {"type":"Point","coordinates":[1,2]}, "f": 3.5}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t, '', 9, true) FROM (SELECT 1 AS id, 'a' AS name, 'POINT(1 2)'::geometry AS geom, ARRAY[1,2] AS arr) t;`,
					Expected: []sql.Row{{"{\"type\": \"Feature\", \"geometry\": {\"type\":\"Point\",\"coordinates\":[1,2]}, \"properties\": {\"id\": 1,\n \"name\": \"a\",\n \"arr\": [1,2]}}"}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t, '', 9, false, 'id') FROM (SELECT 7 AS id, 'x' AS name, 'POINT(1 2)'::geometry AS geom) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1,2]}, "id": 7, "properties": {"name": "x"}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t, 'geom', 9, true, 'id') FROM (SELECT 1 AS id, 'a' AS name, 'POINT(1 2)'::geometry AS geom) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1,2]}, "id": 1, "properties": {"name": "a"}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t) FROM (SELECT 1 AS id, NULL::geometry AS geom) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": null, "properties": {"id": 1}}`}},
				},
				{
					Query:           `SELECT ST_AsGeoJSON(t) FROM (SELECT 1 AS id) t;`,
					ExpectedErr:     `geometry column is missing`,
					ExpectedErrCode: `22023`,
				},
				{
					Query:           `SELECT ST_AsGeoJSON(t, 'nope') FROM (SELECT 1 AS id, 'POINT(1 2)'::geometry AS geom) t;`,
					ExpectedErr:     `geometry column is missing`,
					ExpectedErrCode: `22023`,
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t) FROM (SELECT 'POINT(1 2)'::geography AS geog, 2 AS n) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1,2]}, "properties": {"n": 2}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t) FROM (SELECT 'POINT(1 2)'::geometry AS a, 'POINT(3 4)'::geometry AS b) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1,2]}, "properties": {"b": {"type":"Point","coordinates":[3,4]}}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t) FROM (SELECT '{"x":[1,2]}'::jsonb AS j, ARRAY[1,2] AS arr, true AS b, NULL::text AS n, '2024-01-02'::date AS d, 'POINT(1 2)'::geometry AS geom) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1,2]}, "properties": {"j": {"x": [1, 2]}, "arr": [1,2], "b": true, "n": null, "d": "2024-01-02"}}`}},
				},
				{
					Query:           `SELECT ST_AsGeoJSON(t, '', 9, false, 'nope') FROM (SELECT 7 AS id, 'POINT(1 2)'::geometry AS geom) t;`,
					ExpectedErr:     `Specified id column "nope" is missing`,
					ExpectedErrCode: `22023`,
				},
				{
					Query:           `SELECT ST_AsGeoJSON(t, '', 9, false, 'geom') FROM (SELECT 7 AS id, 'POINT(1 2)'::geometry AS geom) t;`,
					ExpectedErr:     `Specified id column "geom" is missing`,
					ExpectedErrCode: `22023`,
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t, '', 9, false, 'id') FROM (SELECT NULL::int AS id, 'POINT(1 2)'::geometry AS geom) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1,2]}, "id": null, "properties": {}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t, 'geom', 0) FROM (SELECT 'LINESTRING(1.6 2.4,3 4)'::geometry AS geom, 'he said "hi"'::text AS q) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"LineString","coordinates":[[2,2],[3,4]]}, "properties": {"q": "he said \"hi\""}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t) FROM (SELECT 'POINT EMPTY'::geometry AS geom) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[]}, "properties": {}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t) FROM (SELECT 1.50::numeric AS n, 'NaN'::float8 AS f, 'Infinity'::float8 AS i, 'POINT(1 2)'::geometry AS geom) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1,2]}, "properties": {"n": 1.50, "f": "NaN", "i": "Infinity"}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t) FROM (SELECT 'POINT(1.1234567891234 2)'::geometry AS a, 'SRID=4326;POINT(1.1234567891234 2)'::geometry AS b, 'POINT(1 2)'::geography AS c) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1.123456789,2]}, "properties": {"b": {"type":"Point","crs":{"type":"name","properties":{"name":"EPSG:4326"}},"coordinates":[1.1234567891234,2]}, "c": "0101000020E6100000000000000000F03F0000000000000040"}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t) FROM (SELECT 'SRID=3857;POINT(1 2)'::geometry AS geom) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","crs":{"type":"name","properties":{"name":"EPSG:3857"}},"coordinates":[1,2]}, "properties": {}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t, 'geom', 2) FROM (SELECT 'POINT(1.123 2)'::geography AS geom) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1.12,2]}, "properties": {}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t) FROM (SELECT 'POINT(1 2)'::geometry AS geom, ROW(1,'x') AS r, 'BOX(1 2,3 4)'::box2d AS b) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1,2]}, "properties": {"r": {"f1":1,"f2":"x"}, "b": "BOX(1 2,3 4)"}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t, '', 9, false, 'id') FROM (SELECT 'k'::text AS id, 'POINT(1 2)'::geometry AS geom) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1,2]}, "id": "k", "properties": {}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t, 'GEOM') FROM (SELECT 'POINT(1 2)'::geometry AS "GEOM", 'POINT(5 5)'::geometry AS geom) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1,2]}, "properties": {"geom": {"type":"Point","coordinates":[5,5]}}}`}},
					Skip:     true,
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t, 'geom', -1) FROM (SELECT 'POINT(1.123 2)'::geometry AS geom) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1,2]}, "properties": {}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t, 'geom', 20) FROM (SELECT 'POINT(1.1234567891234567 2)'::geometry AS geom) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1.1234567891234566,2]}, "properties": {}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON(t) FROM (SELECT 'POINT(1 2)'::geometry AS geom, 'x'::varchar(3) AS v, 1.5::float4 AS r, 10::int8 AS b, '12:30'::time AS tm, '2024-01-02 03:04:05'::timestamp AS ts) t;`,
					Expected: []sql.Row{{`{"type": "Feature", "geometry": {"type":"Point","coordinates":[1,2]}, "properties": {"v": "x", "r": 1.5, "b": 10, "tm": "12:30:00", "ts": "2024-01-02T03:04:05"}}`}},
				},
				{
					Query:    `SELECT ST_AsGeoJSON('POINT(1 2)'), ST_AsGeoJSON('POINT(1 2)'::geometry);`,
					Expected: []sql.Row{{`{"type":"Point","coordinates":[1,2]}`, `{"type":"Point","coordinates":[1,2]}`}},
				},
				{
					Query:    `SELECT 'SRID=4326;POINT(1.1234567891234 2)'::geometry::json, json('POINT(1.1234567891234567 2)'::geometry), 'POINT(1.12345678912345 2)'::geometry::jsonb;`,
					Expected: []sql.Row{{`{"type":"Point","crs":{"type":"name","properties":{"name":"EPSG:4326"}},"coordinates":[1.1234567891234,2]}`, `{"type":"Point","coordinates":[1.123456789123457,2]}`, `{"type": "Point", "coordinates": [1.12345678912345, 2]}`}},
				},
				{
					Query:    `SELECT 'POINT(0.1234567890123456789 123456.123456789012345)'::geometry::json, 'POINT(1e-20 1e20)'::geometry::json;`,
					Expected: []sql.Row{{`{"type":"Point","coordinates":[0.123456789012346,123456.123456789]}`, `{"type":"Point","coordinates":[1e-20,1e+20]}`}},
				},
			},
		},
		{
			Name: `postgis sphere distance with a radius`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query: `SELECT pg_get_function_arguments(oid), proisstrict FROM pg_proc WHERE proname = 'st_distancesphere' ORDER BY 1;`,
					Expected: []sql.Row{
						{`geom1 geometry, geom2 geometry`, `t`},
						{`geom1 geometry, geom2 geometry, radius double precision`, `t`},
					},
				},
				{
					Query:    `SELECT round(ST_DistanceSphere('POINT(0 0)','POINT(1 1)', 6371000)::numeric, 6), round(ST_DistanceSphere('POINT(0 0)','POINT(1 1)', 1)::numeric, 6), round(ST_DistanceSphere('POINT(0 0)','POINT(1 1)', 1000)::numeric, 6);`,
					Expected: []sql.Row{{framework.Numeric(`157249.381272`), framework.Numeric(`0.024682`), framework.Numeric(`24.682056`)}},
				},
				{
					Query:    `SELECT round(ST_DistanceSphere('LINESTRING(0 0,1 1)','POINT(5 5)', 1000)::numeric, 6), round(ST_DistanceSphere('POLYGON((0 0,1 0,1 1,0 0))','POINT(0.2 0.1)', 1000)::numeric, 6), round(ST_DistanceSphere('POINT(0 0)','POINT(180 0)', 2)::numeric, 6);`,
					Expected: []sql.Row{{framework.Numeric(`98.652985`), framework.Numeric(`0.000000`), framework.Numeric(`6.283185`)}},
				},
				{
					Query:    `SELECT ST_DistanceSphere('POINT EMPTY','POINT(1 1)', 1000);`,
					Expected: []sql.Row{{-1.0}},
				},
				{
					Query:           `SELECT ST_DistanceSphere('SRID=4326;POINT(0 0)','SRID=3857;POINT(1 1)', 1000);`,
					ExpectedErr:     `geometry_distance_spheroid: Operation on mixed SRID geometries (Point, 4326) != (Point, 3857)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT round(ST_DistanceSphere('POINT(0 0)','POINT(1 1)', -5)::numeric, 6), round(ST_DistanceSphere('POINT(0 0)','POINT(1 1)', 0)::numeric, 6);`,
					Expected: []sql.Row{{framework.Numeric(`-0.123410`), framework.Numeric(`0.000000`)}},
				},
				{
					Query:    `SELECT round(ST_DistanceSphere('POINT(0 100)','POINT(1 1)', 1000)::numeric, 6), round(ST_DistanceSphere('POINT Z (0 0 5)','POINT(1 1)', 1000)::numeric, 6);`,
					Expected: []sql.Row{{framework.Numeric(`1378.837047`), framework.Numeric(`24.682056`)}},
				},
				{
					Query:           `SELECT ST_DistanceSphere('GEOMETRYCOLLECTION(POINT(0 0))','POINT(1 1)', 1000);`,
					ExpectedErr:     `geometry_distance_spheroid: Only point/line/polygon supported.`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT round(ST_DistanceSphere('MULTIPOINT(0 0,3 3)','LINESTRING(10 10,11 12)', 6371008.7714)::numeric, 6), round(ST_DistanceSphere('POINT(-170 10)','POINT(170 -10)', 6378137)::numeric, 6);`,
					Expected: []sql.Row{{framework.Numeric(`1096886.054368`), framework.Numeric(`3140555.328387`)}},
				},
				{
					Query:    `SELECT ST_DistanceSphere('POINT(0 0)','POINT(1 1)', NULL) IS NULL;`,
					Expected: []sql.Row{{`t`}},
				},
			},
		},
		{
			Name: `postgis coverage union of arrays`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT ST_AsText(ST_CoverageUnion(ARRAY['POLYGON((0 0,1 0,1 1,0 1,0 0))'::geometry,NULL,'POLYGON((1 0,2 0,2 1,1 1,1 0))']));`,
					Expected: []sql.Row{{`POLYGON((2 1,2 0,1 0,0 0,0 1,1 1,2 1))`}},
				},
				{
					Query:    `SELECT ST_CoverageUnion(ARRAY[]::geometry[]) IS NULL, ST_CoverageUnion(ARRAY[NULL::geometry]) IS NULL, ST_CoverageUnion(ARRAY['POLYGON EMPTY'::geometry]) IS NULL;`,
					Expected: []sql.Row{{`t`, `t`, `t`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_CoverageUnion(ARRAY['SRID=4326;POLYGON((0 0,1 0,1 1,0 1,0 0))'::geometry,'SRID=3857;POLYGON((1 0,2 0,2 1,1 1,1 0))']));`,
					Expected: []sql.Row{{`POLYGON((2 1,2 0,1 0,0 0,0 1,1 1,2 1))`}},
				},
				{
					Query:           `SELECT ST_AsText(ST_CoverageUnion(ARRAY['POLYGON((0 0,1 0,1 1,0 1,0 0))'::geometry,'POINT(5 5)']));`,
					ExpectedErr:     `Error computing coverage union: IllegalArgumentException: Unhandled geometry type in CoverageUnion.`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsText(ST_CoverageUnion(ARRAY['POLYGON Z ((0 0 1,1 0 1,1 1 1,0 1 1,0 0 1))'::geometry,'POLYGON Z ((1 0 2,2 0 2,2 1 2,1 1 2,1 0 2))'])), ST_AsText(ST_CoverageUnion(ARRAY['POLYGON M ((0 0 1,1 0 1,1 1 1,0 1 1,0 0 1))'::geometry]));`,
					Expected: []sql.Row{{`POLYGON((2 1,2 0,1 0,0 0,0 1,1 1,2 1))`, `POLYGON((1 1,1 0,0 0,0 1,1 1))`}},
				},
				{
					Query:    `SELECT ST_AsText(ST_CoverageUnion(ARRAY['POLYGON((0 0,1 0,1 1,0 1,0 0))'::geometry,'POLYGON((5 5,6 5,6 6,5 6,5 5))'])), ST_AsText(ST_CoverageUnion(ARRAY['MULTIPOLYGON(((0 0,1 0,1 1,0 1,0 0)),((1 0,2 0,2 1,1 1,1 0)))'::geometry]));`,
					Expected: []sql.Row{{`MULTIPOLYGON(((6 6,6 5,5 5,5 6,6 6)),((1 1,1 0,0 0,0 1,1 1)))`, `POLYGON((2 1,2 0,1 0,0 0,0 1,1 1,2 1))`}},
				},
				{
					Query:    `SELECT ST_AsEWKT(ST_CoverageUnion(g)) FROM (VALUES ('SRID=4326;POLYGON((0 0,1 0,1 1,0 1,0 0))'::geometry),('SRID=4326;POLYGON((1 0,2 0,2 1,1 1,1 0))')) t(g);`,
					Expected: []sql.Row{{`POLYGON((2 1,2 0,1 0,0 0,0 1,1 1,2 1))`}},
				},
				{
					Query:           `SELECT ST_AsText(ST_CoverageUnion(g)) FROM (VALUES ('POLYGON((0 0,1 0,1 1,0 1,0 0))'::geometry),('POINT(5 5)')) t(g);`,
					ExpectedErr:     `Error computing coverage union: IllegalArgumentException: Unhandled geometry type in CoverageUnion.`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `SELECT ST_AsText(ST_Union(ARRAY['POLYGON((0 0,1 0,1 1,0 1,0 0))'::geometry,NULL]));`,
					Expected: []sql.Row{{`POLYGON((0 0,1 0,1 1,0 1,0 0))`}},
				},
			},
		},
		{
			Name: `postgis bounding box cache trigger`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
				`CREATE TABLE cb (id INT PRIMARY KEY, g geometry, h text);`,
				`INSERT INTO cb VALUES (1, 'POINT(1 1)', 'a');`,
				`CREATE TRIGGER cb_ok BEFORE INSERT OR UPDATE ON cb FOR EACH ROW EXECUTE FUNCTION postgis_cache_bbox('g');`,
				`CREATE TABLE cb2 (id INT PRIMARY KEY, g geometry, h text);`,
				`CREATE TRIGGER cb_none BEFORE INSERT ON cb2 FOR EACH ROW EXECUTE FUNCTION postgis_cache_bbox();`,
				`CREATE TABLE cb3 (id INT PRIMARY KEY, g geometry, h text);`,
				`CREATE TRIGGER cb_nope BEFORE INSERT ON cb3 FOR EACH ROW EXECUTE FUNCTION postgis_cache_bbox('nope');`,
				`CREATE TABLE cb4 (id INT PRIMARY KEY, g geometry, h text);`,
				`CREATE TRIGGER cb_text BEFORE INSERT ON cb4 FOR EACH ROW EXECUTE FUNCTION postgis_cache_bbox('h');`,
				`CREATE TABLE cb5 (id INT PRIMARY KEY, g geometry, h text);`,
				`CREATE TRIGGER cb_after AFTER INSERT ON cb5 FOR EACH ROW EXECUTE FUNCTION postgis_cache_bbox('g');`,
				`CREATE TABLE cb6 (id INT PRIMARY KEY, g geometry, h text);`,
				`INSERT INTO cb6 VALUES (1, 'POINT(1 1)', 'a');`,
				`CREATE TRIGGER cb_del BEFORE DELETE ON cb6 FOR EACH ROW EXECUTE FUNCTION postgis_cache_bbox('nope', 'extra');`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT pg_get_function_result(oid), provolatile FROM pg_proc WHERE proname = 'postgis_cache_bbox';`,
					Expected: []sql.Row{{`trigger`, `v`}},
				},
				{
					Query:    `INSERT INTO cb VALUES (2, 'LINESTRING(0 0,1 1)', 'b') RETURNING id, ST_AsText(g), h;`,
					Expected: []sql.Row{{2, `LINESTRING(0 0,1 1)`, `b`}},
				},
				{
					Query:    `UPDATE cb SET h = 'z' WHERE id = 1 RETURNING id, ST_AsText(g), h;`,
					Expected: []sql.Row{{1, `POINT(1 1)`, `z`}},
				},
				{
					Query:           `INSERT INTO cb2 VALUES (1, 'POINT(1 1)', 'x');`,
					ExpectedErr:     `trigger 'cache_bbox' must be called with one argument`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `INSERT INTO cb3 VALUES (1, 'POINT(1 1)', 'x');`,
					ExpectedErr:     `trigger cb_nope can't find attribute nope`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `INSERT INTO cb4 VALUES (1, 'POINT(1 1)', 'x');`,
					ExpectedErr:     `trigger cb_text requested to apply to a non-geometry field (h)`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `INSERT INTO cb5 VALUES (1, 'POINT(1 1)', 'x') RETURNING id;`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:           `DELETE FROM cb6 WHERE id = 1 RETURNING id;`,
					ExpectedErr:     `trigger 'cache_bbox' must be called with one argument`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:           `SELECT postgis_cache_bbox();`,
					ExpectedErr:     `cache_bbox: not called by trigger manager`,
					ExpectedErrCode: `XX000`,
				},
				{
					Query:    `CREATE TRIGGER cb_stmt BEFORE INSERT ON cb FOR EACH STATEMENT EXECUTE FUNCTION postgis_cache_bbox('g');`,
					Expected: []sql.Row{},
					Skip:     true,
				},
			},
		},
		{
			Name: `postgis extensions upgrade`,
			SetUpScript: []string{
				`CREATE EXTENSION postgis;`,
			},
			Assertions: []framework.ScriptTestAssertion{
				{
					Query:    `SELECT pg_get_function_arguments(oid), proisstrict, provolatile FROM pg_proc WHERE proname = 'postgis_extensions_upgrade';`,
					Expected: []sql.Row{{`target_version text DEFAULT NULL::text`, `f`, `v`}},
				},
				{
					Query:    `SELECT postgis_extensions_upgrade();`,
					Expected: []sql.Row{{`Upgrade to version 3.6.4 completed, run SELECT postgis_full_version(); for details`}},
				},
				{
					Query:    `SELECT postgis_extensions_upgrade('3.6.4');`,
					Expected: []sql.Row{{`Upgrade to version 3.6.4 completed, run SELECT postgis_full_version(); for details`}},
				},
				{
					Query:           `SELECT postgis_extensions_upgrade('9.9');`,
					ExpectedErr:     `extension "postgis" has no update path from version "ANY" to version "9.9"`,
					ExpectedErrCode: `22023`,
				},
				{
					Query:           `SELECT postgis_extensions_upgrade('');`,
					ExpectedErr:     `zero-length delimited identifier at or near """"`,
					ExpectedErrCode: `42601`,
				},
				{
					Query:    `SELECT extversion FROM pg_extension WHERE extname = 'postgis';`,
					Expected: []sql.Row{{`3.6.4`}},
				},
				{
					Query:    `ALTER EXTENSION postgis UPDATE;`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT ST_AsText(ST_Transform('SRID=4326;POINT(1 2)'::geometry, 3857)), (ST_Dump('MULTIPOINT(1 1,2 2)'::geometry)).path[1];`,
					Expected: []sql.Row{
						{`POINT(111319.49079327357 222684.20850554405)`, 1},
						{`POINT(111319.49079327357 222684.20850554405)`, 2},
					},
				},
				{
					Query:    `ALTER EXTENSION postgis UPDATE TO '3.6.4';`,
					Expected: []sql.Row{},
				},
			},
		},
	})
}
