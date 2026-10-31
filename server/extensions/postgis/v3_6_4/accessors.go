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

// geometryType implements GeometryType, which returns the upper-case type name, suffixed by "M" when the geometry
// has M coordinates but no Z coordinates.
func geometryType(ctx *sql.Context, args ...any) (any, error) {
	s := args[0].(Geometry).Shape
	if s.ctype == dimXYM {
		return typeTags[s.typeNum] + "M", nil
	}
	return typeTags[s.typeNum], nil
}

// stGeometryType implements ST_GeometryType, which returns the type name in the form "ST_Point".
func stGeometryType(ctx *sql.Context, args ...any) (any, error) {
	return "ST_" + typeNameForNumber(args[0].(Geometry).Shape.typeNum), nil
}

// stSRID implements ST_SRID.
func stSRID(ctx *sql.Context, args ...any) (any, error) {
	return args[0].(Geometry).SRID, nil
}

// stSetSRID implements ST_SetSRID.
func stSetSRID(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	g.SRID = clampSRID(ctx, args[1].(int32))
	return g, nil
}

// pointOrdinate returns the ordinate at the given index of a point, or nil when the point is empty or lacks that
// ordinate, which a negative index signals. The function name is used in the error for a geometry that is not a
// point.
func pointOrdinate(g Geometry, funcName string, index int) (any, error) {
	if g.Shape.typeNum != pointType {
		return nil, pgerror.Newf(pgcode.Internal, "Argument to %s() must have type POINT", funcName)
	}
	if index < 0 || len(g.Shape.coords) == 0 {
		return nil, nil
	}
	return g.Shape.coords[index], nil
}

// stX implements ST_X.
func stX(ctx *sql.Context, args ...any) (any, error) {
	return pointOrdinate(args[0].(Geometry), "ST_X", 0)
}

// stY implements ST_Y.
func stY(ctx *sql.Context, args ...any) (any, error) {
	return pointOrdinate(args[0].(Geometry), "ST_Y", 1)
}

// stZ implements ST_Z.
func stZ(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	index := -1
	if g.Shape.hasZ() {
		index = 2
	}
	return pointOrdinate(g, "ST_Z", index)
}

// stM implements ST_M.
func stM(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	index := -1
	if g.Shape.hasM() {
		index = g.Shape.ctype.Dimension() - 1
	}
	return pointOrdinate(g, "ST_M", index)
}

// accessorRoutines returns the routines that read properties of geometries and geographies.
func accessorRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "geometrytype", Symbol: "geometrytype_geometry", Parameters: params("geometry"), Returns: "text", Strict: true, Impl: geometryType},
		{Name: "st_geometrytype", Symbol: "geometry_geometrytype", Parameters: params("geometry"), Returns: "text", Strict: true, Impl: stGeometryType},
		{Name: "st_m", Symbol: "LWGEOM_m_point", Parameters: params("geometry"), Returns: "float8", Strict: true, Impl: stM},
		{Name: "st_setsrid", Symbol: "st_setsrid_geography_int4", Parameters: []extdef.Parameter{{Name: "geog", Type: "geography"}, {Name: "srid", Type: "int4"}}, Returns: "geography", Strict: true, Impl: stSetSRID},
		{Name: "st_setsrid", Symbol: "st_setsrid_geometry_int4", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "srid", Type: "int4"}}, Returns: "geometry", Strict: true, Impl: stSetSRID},
		{Name: "st_srid", Symbol: "st_srid_geography", Parameters: []extdef.Parameter{{Name: "geog", Type: "geography"}}, Returns: "int4", Strict: true, Impl: stSRID},
		{Name: "st_srid", Symbol: "st_srid_geometry", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}}, Returns: "int4", Strict: true, Impl: stSRID},
		{Name: "st_x", Symbol: "LWGEOM_x_point", Parameters: params("geometry"), Returns: "float8", Strict: true, Impl: stX},
		{Name: "st_y", Symbol: "LWGEOM_y_point", Parameters: params("geometry"), Returns: "float8", Strict: true, Impl: stY},
		{Name: "st_z", Symbol: "LWGEOM_z_point", Parameters: params("geometry"), Returns: "float8", Strict: true, Impl: stZ},
	}
}
