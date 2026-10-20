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
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// subdivider splits shapes into pieces with few enough vertices.
type subdivider struct {
	geos        *spatial.GEOS
	dimension   int
	maxVertices int
	gridSize    float64
	pieces      []shape
}

// subdivideMaxDepth is the depth of halving after which a piece is kept whatever its number of vertices.
const subdivideMaxDepth = 50

// subdivide adds the pieces of a shape to the subdivider's pieces. Collections are subdivided member by member, and
// members of a lower dimension than the input are dropped. A piece with too many vertices is cut in half across the
// longer side of its bounding box, and each half is clipped out of it, cleaned of collinear vertices, and subdivided
// in turn. A polygon is cut through the vertex of its exterior ring nearest to the middle, or through the vertex of
// its largest hole when its holes hold most of its vertices.
func (d *subdivider) subdivide(s shape, depth int) error {
	b, ok := s.bounds()
	if !ok {
		return nil
	}
	width, height := b.maxX-b.minX, b.maxY-b.minY
	if width == 0 && height == 0 {
		if s.typeNum == pointType && d.dimension == 0 {
			d.pieces = append(d.pieces, s)
		}
		return nil
	}
	if width == 0 {
		b.minX, b.maxX, width = b.minX-subdivideTolerance, b.maxX+subdivideTolerance, 2*subdivideTolerance
	}
	if height == 0 {
		b.minY, b.maxY, height = b.minY-subdivideTolerance, b.maxY+subdivideTolerance, 2*subdivideTolerance
	}
	if hasMembers(s.typeNum) && s.typeNum != multiPointType {
		for _, member := range s.members {
			if err := d.subdivide(member, depth); err != nil {
				return err
			}
		}
		return nil
	}
	if s.dimension() < d.dimension {
		return nil
	}
	vertices := s.numPoints()
	if vertices <= d.maxVertices || depth > subdivideMaxDepth {
		d.pieces = append(d.pieces, s)
		return nil
	}
	axis, low, high := 0, b.minX, b.maxX
	if width <= height {
		axis, low, high = 1, b.minY, b.maxY
	}
	middle := (low + high) / 2
	cut := middle
	if s.typeNum == polygonType {
		dims := s.ctype.Dimension()
		ring := s.rings[0]
		if vertices >= 2*len(ring)/dims {
			largest := 0.0
			for _, hole := range s.rings[1:] {
				if area := math.Abs(signedRingArea(hole, s.ctype)); area >= largest {
					largest, ring = area, hole
				}
			}
		}
		closest := math.Inf(1)
		for i := axis; i < len(ring); i += dims {
			if distance := math.Abs(ring[i] - middle); distance < closest {
				cut, closest = ring[i], distance
			}
		}
		if cut == low || cut == high {
			cut = middle
		}
	}
	for _, half := range [][2]float64{{low, cut}, {cut, high}} {
		box := b
		if axis == 0 {
			box.minX, box.maxX = half[0], half[1]
		} else {
			box.minY, box.maxY = half[0], half[1]
		}
		geosShape, err := s.toGEOS()
		if err != nil {
			return err
		}
		clipped, err := d.geos.Intersection(geosShape, flatBoxShape(box.minX, box.minY, box.maxX, box.maxY).linearToGEOS(), d.gridSize)
		if err != nil {
			return geosError("GEOSIntersection", err)
		}
		piece, err := simplifyShape(shapeFromGEOS(clipped, s.hasZ()), 0, true)
		if err != nil {
			return err
		}
		if !piece.isEmpty() {
			if err = d.subdivide(piece, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

// subdivideTolerance is how far a bounding box without width or height is widened on each side before it is cut.
const subdivideTolerance = 1e-12

// stSubdivide implements ST_Subdivide, which returns the pieces of a geometry that hold at most the given number of
// vertices each, clipping with the grid size unless it is negative.
func stSubdivide(ctx *sql.Context, args ...any) (any, error) {
	g, maxVertices, gridSize := args[0].(Geometry), int(args[1].(int32)), args[2].(float64)
	if g.Shape.typeNum == polyhedralSurfaceType || g.Shape.typeNum == tinType {
		return nil, pgerror.Newf(pgcode.Internal, "lwgeom_subdivide_recursive: unsupported geometry type '%s'", typeNameForNumber(g.Shape.typeNum))
	}
	if maxVertices < 5 {
		return nil, pgerror.New(pgcode.Internal, "lwgeom_subdivide_prec: cannot subdivide to fewer than 5 vertices per output")
	}
	geos := spatial.NewGEOS()
	defer geos.Close()
	d := &subdivider{geos: geos, dimension: g.Shape.dimension(), maxVertices: maxVertices, gridSize: gridSize}
	if err := d.subdivide(g.Shape, 0); err != nil {
		return nil, err
	}
	rows := make([]sql.Row, len(d.pieces))
	for i, piece := range d.pieces {
		rows[i] = sql.Row{g.withShape(piece)}
	}
	return sql.RowsToRowIter(rows...), nil
}

// subdivideRoutines returns the routines that subdivide geometries.
func subdivideRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "st_subdivide", Symbol: "ST_Subdivide", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "maxvertices", Type: "int4", Default: `256`}, {Name: "gridsize", Type: "float8", Default: `'-1.0'::numeric`}}, Returns: "geometry", Strict: true, SetOf: true, Impl: stSubdivide},
	}
}
