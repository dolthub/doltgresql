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
	"io"
	"math"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// hexagonHalfHeight is half the height of a hexagon whose sides have length one, as PostGIS writes it.
const hexagonHalfHeight = 0.8660254037844387

// rectanglePolygon returns the polygon of a rectangle, starting at its lower left corner and running clockwise.
func rectanglePolygon(minX float64, minY float64, maxX float64, maxY float64) shape {
	ring := rectangle([]float64{minX, minY}, []float64{minX, maxY}, []float64{maxX, maxY}, []float64{maxX, minY})
	return shape{typeNum: polygonType, rings: [][]float64{ring}}
}

// stTileEnvelope implements ST_TileEnvelope, which returns the polygon of a tile in the grid that divides the bounds
// into 2^zoom tiles on each side, counting rows down from the top, with an optional margin as a fraction of a tile.
func stTileEnvelope(ctx *sql.Context, args ...any) (any, error) {
	zoom, x, y := args[0].(int32), args[1].(int32), args[2].(int32)
	bounds, margin := args[3].(Geometry), args[4].(float64)
	if zoom < 0 || zoom >= 32 {
		return nil, pgerror.Newf(pgcode.Internal, "ST_TileEnvelope: Invalid tile zoom value, %d", zoom)
	}
	tiles := int64(1) << uint(zoom)
	if int64(x) < 0 || int64(x) >= tiles {
		return nil, pgerror.Newf(pgcode.Internal, "ST_TileEnvelope: Invalid tile x value, %d", x)
	}
	if int64(y) < 0 || int64(y) >= tiles {
		return nil, pgerror.Newf(pgcode.Internal, "ST_TileEnvelope: Invalid tile y value, %d", y)
	}
	b, ok := bounds.Shape.bounds()
	if !ok || b.maxX-b.minX <= 0 || b.maxY-b.minY <= 0 {
		return nil, pgerror.New(pgcode.Internal, "ST_TileEnvelope: Geometric bounds are too small")
	}
	if margin < -0.5 {
		return nil, pgerror.Newf(pgcode.Internal, "ST_TileEnvelope: Margin must not be less than -50%%, margin=%f", margin)
	}
	width, height := (b.maxX-b.minX)/float64(tiles), (b.maxY-b.minY)/float64(tiles)
	minX := b.minX + float64(width*(float64(x)-margin))
	maxX := b.minX + float64(width*(float64(x+1)+margin))
	minY := b.maxY - float64(height*(float64(y+1)+margin))
	maxY := b.maxY - float64(height*(float64(y)-margin))
	return bounds.withShape(rectanglePolygon(minX, minY, maxX, maxY)), nil
}

// cellOrigin returns the X and Y of the origin argument at the index, and the geometry that gives the cell its SRID.
func cellOrigin(ctx *sql.Context, args []any, index int) (Geometry, []float64, error) {
	origin := args[index].(Geometry)
	coords, err := pointArg(ctx, origin, "ST_Hexagon: origin must be a point")
	return origin, coords, err
}

// stSquare implements ST_Square, which returns the square cell in the given column and row of the grid of squares of
// the given size that has a corner at the origin.
func stSquare(ctx *sql.Context, args ...any) (any, error) {
	size, i, j := args[0].(float64), float64(args[1].(int32)), float64(args[2].(int32))
	origin, coords, err := cellOrigin(ctx, args, 3)
	if err != nil {
		return nil, err
	}
	return origin.withShape(squareCell(coords, size, i, j)), nil
}

// squareCell returns the square in the given column and row of the grid of squares of the given size that has a
// corner at the origin.
func squareCell(origin []float64, size float64, i float64, j float64) shape {
	return rectanglePolygon(origin[0]+float64(i*size), origin[1]+float64(j*size), origin[0]+float64((i+1)*size), origin[1]+float64((j+1)*size))
}

// stHexagon implements ST_Hexagon, which returns the flat-topped hexagon in the given column and row of the grid of
// hexagons with the given side length that has a hexagon centered on the origin. Odd columns sit half a hexagon
// higher.
func stHexagon(ctx *sql.Context, args ...any) (any, error) {
	size, i, j := args[0].(float64), args[1].(int32), args[2].(int32)
	origin, coords, err := cellOrigin(ctx, args, 3)
	if err != nil {
		return nil, err
	}
	return origin.withShape(hexagonCell(coords, size, i, j)), nil
}

// hexagonCell returns the hexagon in the given column and row of the grid of hexagons with the given side length that
// has a hexagon centered on the origin.
func hexagonCell(origin []float64, size float64, i int32, j int32) shape {
	var ring []float64
	for _, corner := range [][2]float64{{-1, 0}, {-0.5, -1}, {0.5, -1}, {1, 0}, {0.5, 1}, {-0.5, 1}, {-1, 0}} {
		ring = append(ring,
			origin[0]+float64(size*(float64(1.5*float64(i))+corner[0])),
			origin[1]+float64(size*hexagonHalfHeight*float64(2*j+(i&1)+int32(corner[1]))))
	}
	return shape{typeNum: polygonType, rings: [][]float64{ring}}
}

// gridIter returns the cells of a grid of squares or hexagons, cornered at or centered on the origin, that cover the
// box of the bounds, in order of column and then row. As in PostGIS, the bounds' box is widened to float32 values, the
// cells are chosen by dividing that box by the cell size, and a hexagon whose top edge only touches the box is left
// out.
type gridIter struct {
	bounds   Geometry
	box      bbox
	size     float64
	hexagons bool
	geometry *pgtypes.DoltgresType
	i, lastI int32
	j, lastJ int32
}

var _ sql.RowIter = (*gridIter)(nil)

// gridImpl returns the implementation of ST_SquareGrid or ST_HexagonGrid.
func gridImpl(hexagons bool) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		size, bounds := args[0].(float64), args[1].(Geometry)
		b, ok := bounds.Shape.bounds()
		if !ok || !(size > 0) {
			return sql.RowsToRowIter(), nil
		}
		geometryType, err := extensionType(ctx, "geometry")
		if err != nil {
			return nil, err
		}
		g := &gridIter{bounds: bounds, box: b.floatRounded(), size: size, hexagons: hexagons, geometry: geometryType, j: 1}
		first, last := math.Floor(g.box.minX/size), math.Floor(g.box.maxX/size)
		if hexagons {
			first, last = math.Ceil((g.box.minX-size)/(1.5*size)), math.Floor((g.box.maxX+size)/(1.5*size))
		}
		g.i, g.lastI = int32(first)-1, int32(last)
		return g, nil
	}
}

// cell returns the cell in the given column and row.
func (g *gridIter) cell(i int32, j int32) shape {
	if g.hexagons {
		return hexagonCell([]float64{0, 0}, g.size, i, j)
	}
	return squareCell([]float64{0, 0}, g.size, float64(i), float64(j))
}

// Next implements the interface sql.RowIter.
func (g *gridIter) Next(ctx *sql.Context) (sql.Row, error) {
	for g.j > g.lastJ {
		if g.i >= g.lastI {
			return nil, io.EOF
		}
		g.i++
		first, last := math.Floor(g.box.minY/g.size), math.Floor(g.box.maxY/g.size)
		if g.hexagons {
			halfHeight, odd := g.size*hexagonHalfHeight, float64(g.i&1)
			first, last = math.Floor((g.box.minY/halfHeight-odd-1)/2)+1, math.Floor((g.box.maxY/halfHeight-odd+1)/2)
		}
		g.j, g.lastJ = int32(first), int32(last)
	}
	row := sql.Row{[]pgtypes.RecordValue{
		{Value: g.bounds.withShape(g.cell(g.i, g.j)), Type: g.geometry},
		{Value: g.i, Type: pgtypes.Int32},
		{Value: g.j, Type: pgtypes.Int32},
	}}
	g.j++
	return row, nil
}

// Close implements the interface sql.RowIter.
func (g *gridIter) Close(*sql.Context) error {
	return nil
}

// gridRoutines returns the routines that build grid cells.
func gridRoutines() []extdef.Routine {
	origin := extdef.Parameter{Name: "origin", Type: "geometry", Default: `'010100000000000000000000000000000000000000'::geometry`}
	cell := []extdef.Parameter{{Name: "size", Type: "float8"}, {Name: "cell_i", Type: "int4"}, {Name: "cell_j", Type: "int4"}, origin}
	gridParams := []extdef.Parameter{{Name: "size", Type: "float8"}, {Name: "bounds", Type: "geometry"}}
	gridOutputs := []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "i", Type: "int4"}, {Name: "j", Type: "int4"}}
	return []extdef.Routine{
		{Name: "st_hexagon", Symbol: "ST_Hexagon", Parameters: cell, Returns: "geometry", Strict: true, Impl: stHexagon},
		{Name: "st_hexagongrid", Symbol: "st_hexagongrid_float8_geometry", Parameters: gridParams, Outputs: gridOutputs, Strict: true, SetOf: true, Impl: gridImpl(true)},
		{Name: "st_square", Symbol: "ST_Square", Parameters: cell, Returns: "geometry", Strict: true, Impl: stSquare},
		{Name: "st_squaregrid", Symbol: "st_squaregrid_float8_geometry", Parameters: gridParams, Outputs: gridOutputs, Strict: true, SetOf: true, Impl: gridImpl(false)},
		{Name: "st_tileenvelope", Symbol: "ST_TileEnvelope", Parameters: []extdef.Parameter{{Name: "zoom", Type: "int4"}, {Name: "x", Type: "int4"}, {Name: "y", Type: "int4"},
			{Name: "bounds", Type: "geometry", Default: `'0102000020110F00000200000093107C45F81B73C193107C45F81B73C193107C45F81B734193107C45F81B7341'::geometry`},
			{Name: "margin", Type: "float8", Default: `0.0`}}, Returns: "geometry", Strict: true, Impl: stTileEnvelope},
	}
}
