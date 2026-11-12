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

// affineMatrix holds the coefficients of a three-dimensional affine transformation in the order that ST_Affine takes
// them: the three rows of the linear part, followed by the X, Y, and Z offsets.
type affineMatrix [12]float64

// apply returns a copy of the shape transformed by the matrix. Shapes without Z coordinates ignore the terms that
// involve Z, and M coordinates are left unchanged.
func (t affineMatrix) apply(s shape) shape {
	is3D := s.ctype.Is3D()
	return s.mapCoords(func(point []float64) []float64 {
		x, y := point[0], point[1]
		if is3D {
			z := point[2]
			point[0] = float64(t[0]*x) + float64(t[1]*y) + float64(t[2]*z) + t[9]
			point[1] = float64(t[3]*x) + float64(t[4]*y) + float64(t[5]*z) + t[10]
			point[2] = float64(t[6]*x) + float64(t[7]*y) + float64(t[8]*z) + t[11]
		} else {
			point[0] = float64(t[0]*x) + float64(t[1]*y) + t[9]
			point[1] = float64(t[3]*x) + float64(t[4]*y) + t[10]
		}
		return point
	})
}

// floatArgs returns the arguments starting at the given index as floats.
func floatArgs(args []any, start int) []float64 {
	vals := make([]float64, len(args)-start)
	for i := range vals {
		vals[i] = args[start+i].(float64)
	}
	return vals
}

// transformImpl returns an implementation that applies the matrix built from the arguments after the geometry.
func transformImpl(matrix func(vals []float64) affineMatrix) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		g := args[0].(Geometry)
		return g.withShape(matrix(floatArgs(args, 1)).apply(g.Shape)), nil
	}
}

// Each of these builds the matrix of an affine function from its arguments after the geometry.
var (
	affine2DMatrix = func(v []float64) affineMatrix {
		return affineMatrix{v[0], v[1], 0, v[2], v[3], 0, 0, 0, 1, v[4], v[5], 0}
	}
	affine3DMatrix = func(v []float64) affineMatrix {
		return affineMatrix(v)
	}
	translateMatrix = func(v []float64) affineMatrix {
		v = append(v, 0)
		return affineMatrix{1, 0, 0, 0, 1, 0, 0, 0, 1, v[0], v[1], v[2]}
	}
	scaleMatrix = func(v []float64) affineMatrix {
		v = append(v, 1)
		return affineMatrix{v[0], 0, 0, 0, v[1], 0, 0, 0, v[2], 0, 0, 0}
	}
	transScaleMatrix = func(v []float64) affineMatrix {
		return affineMatrix{v[2], 0, 0, 0, v[3], 0, 0, 0, 1, v[0] * v[2], v[1] * v[3], 0}
	}
	rotateMatrix = func(v []float64) affineMatrix {
		cos, sin := math.Cos(v[0]), math.Sin(v[0])
		v = append(v, 0, 0)
		x0, y0 := v[1], v[2]
		xOffset := x0 - float64(cos*x0) + float64(sin*y0)
		yOffset := y0 - float64(sin*x0) - float64(cos*y0)
		return affineMatrix{cos, -sin, 0, sin, cos, 0, 0, 0, 1, xOffset, yOffset, 0}
	}
	rotateXMatrix = func(v []float64) affineMatrix {
		cos, sin := math.Cos(v[0]), math.Sin(v[0])
		return affineMatrix{1, 0, 0, 0, cos, -sin, 0, sin, cos, 0, 0, 0}
	}
	rotateYMatrix = func(v []float64) affineMatrix {
		cos, sin := math.Cos(v[0]), math.Sin(v[0])
		return affineMatrix{cos, 0, sin, 0, 1, 0, -sin, 0, cos, 0, 0, 0}
	}
)

// stRotateAroundPoint implements ST_Rotate around the origin point, which is null when the point is empty.
func stRotateAroundPoint(ctx *sql.Context, args ...any) (any, error) {
	g, origin := args[0].(Geometry), args[2].(Geometry)
	x0, err := pointOrdinate(origin, "ST_X", 0)
	if x0 == nil || err != nil {
		return nil, err
	}
	matrix := rotateMatrix([]float64{args[1].(float64), x0.(float64), origin.Shape.coords[1]})
	return g.withShape(matrix.apply(g.Shape)), nil
}

// stScaleByPoint implements ST_Scale with a point of factors, whose missing Z or M factor is one, and an optional
// origin point to scale around. M coordinates are scaled too, and an origin that is not a point is ignored.
func stScaleByPoint(ctx *sql.Context, args ...any) (any, error) {
	g, factors := args[0].(Geometry), args[1].(Geometry).Shape
	if factors.typeNum != pointType {
		return nil, pgerror.New(pgcode.Internal, "Scale factor geometry parameter must be a point")
	}
	if factors.isEmpty() {
		return g, nil
	}
	factor := convertPoint(factors.coords, factors.ctype, dimXYZM, 1, 1)
	origin := make([]float64, 4)
	if len(args) > 2 {
		if o := args[2].(Geometry).Shape; o.typeNum == pointType && !o.isEmpty() {
			origin = convertPoint(o.coords, o.ctype, dimXYZM, 0, 0)
		}
	}
	ctype := g.Shape.ctype
	return g.withShape(g.Shape.mapCoords(func(point []float64) []float64 {
		full := convertPoint(point, ctype, dimXYZM, 0, 0)
		for i := range full {
			full[i] = float64((full[i]-origin[i])*factor[i]) + origin[i]
		}
		return convertPoint(full, dimXYZM, ctype, 0, 0)
	})), nil
}

// affineRoutines returns the routines that apply affine transformations.
func affineRoutines() []extdef.Routine {
	floats := func(count int) []extdef.Parameter {
		ps := params("geometry")
		for i := 0; i < count; i++ {
			ps = append(ps, extdef.Parameter{Type: "float8"})
		}
		return ps
	}
	return []extdef.Routine{
		{Name: "st_affine", Symbol: "st_affine_geometry_float8_float8_float8_float8_float8_float8", Parameters: floats(6), Returns: "geometry", Strict: true, Impl: transformImpl(affine2DMatrix)},
		{Name: "st_affine", Symbol: "LWGEOM_affine", Parameters: floats(12), Returns: "geometry", Strict: true, Impl: transformImpl(affine3DMatrix)},
		{Name: "st_rotate", Symbol: "st_rotate_geometry_float8", Parameters: floats(1), Returns: "geometry", Strict: true, Impl: transformImpl(rotateMatrix)},
		{Name: "st_rotate", Symbol: "st_rotate_geometry_float8_float8_float8", Parameters: floats(3), Returns: "geometry", Strict: true, Impl: transformImpl(rotateMatrix)},
		{Name: "st_rotate", Symbol: "st_rotate_geometry_float8_geometry", Parameters: params("geometry", "float8", "geometry"), Returns: "geometry", Strict: true, Impl: stRotateAroundPoint},
		{Name: "st_rotatex", Symbol: "st_rotatex_geometry_float8", Parameters: floats(1), Returns: "geometry", Strict: true, Impl: transformImpl(rotateXMatrix)},
		{Name: "st_rotatey", Symbol: "st_rotatey_geometry_float8", Parameters: floats(1), Returns: "geometry", Strict: true, Impl: transformImpl(rotateYMatrix)},
		{Name: "st_rotatez", Symbol: "st_rotatez_geometry_float8", Parameters: floats(1), Returns: "geometry", Strict: true, Impl: transformImpl(rotateMatrix)},
		{Name: "st_scale", Symbol: "st_scale_geometry_float8_float8", Parameters: floats(2), Returns: "geometry", Strict: true, Impl: transformImpl(scaleMatrix)},
		{Name: "st_scale", Symbol: "st_scale_geometry_float8_float8_float8", Parameters: floats(3), Returns: "geometry", Strict: true, Impl: transformImpl(scaleMatrix)},
		{Name: "st_scale", Symbol: "st_scale_geometry_geometry", Parameters: params("geometry", "geometry"), Returns: "geometry", Strict: true, Impl: stScaleByPoint},
		{Name: "st_scale", Symbol: "st_scale_geometry_geometry_geometry", Parameters: []extdef.Parameter{{Type: "geometry"}, {Type: "geometry"}, {Name: "origin", Type: "geometry"}}, Returns: "geometry", Strict: true, Impl: stScaleByPoint},
		{Name: "st_translate", Symbol: "st_translate_geometry_float8_float8", Parameters: floats(2), Returns: "geometry", Strict: true, Impl: transformImpl(translateMatrix)},
		{Name: "st_translate", Symbol: "st_translate_geometry_float8_float8_float8", Parameters: floats(3), Returns: "geometry", Strict: true, Impl: transformImpl(translateMatrix)},
		{Name: "st_transscale", Symbol: "st_transscale_geometry_float8_float8_float8_float8", Parameters: floats(4), Returns: "geometry", Strict: true, Impl: transformImpl(transScaleMatrix)},
	}
}
