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

// stGeometricMedian implements ST_GeometricMedian with Weiszfeld's algorithm, starting from the centroid of the
// points and weighting each point by its M coordinate when the multipoint has them. It stops once a step moves less
// than the tolerance, which defaults to a tiny fraction of the largest coordinate, or after the maximum iterations.
func stGeometricMedian(ctx *sql.Context, args ...any) (any, error) {
	if args[0] == nil {
		return nil, nil
	}
	g := args[0].(Geometry)
	s := g.Shape
	if s.typeNum != pointType && s.typeNum != multiPointType {
		return nil, pgerror.Newf(pgcode.Internal, "lwgeom_median: Unsupported geometry type: %s", typeNameForNumber(s.typeNum))
	}
	dims := s.ctype.Dimension()
	spatial := 2
	if s.hasZ() {
		spatial = 3
	}
	var points [][]float64
	var weights []float64
	largest := 0.0
	for _, row := range dumpPoints(s, nil) {
		points = append(points, row.part.coords[:spatial])
		weight := 1.0
		if s.hasM() {
			weight = row.part.coords[dims-1]
		}
		weights = append(weights, weight)
		for _, c := range row.part.coords[:spatial] {
			largest = math.Max(largest, math.Abs(c))
		}
	}
	if len(points) == 0 {
		return g.withShape(shape{typeNum: pointType, ctype: s.ctype &^ dimXYM}), nil
	}
	tolerance := 1e-10 * largest
	if args[1] != nil {
		tolerance = args[1].(float64)
	}
	maxIterations := 10000
	if args[2] != nil {
		maxIterations = int(args[2].(int32))
	}
	median := make([]float64, spatial)
	totalWeight := 0.0
	for i, point := range points {
		for j := range median {
			median[j] += float64(point[j] * weights[i])
		}
		totalWeight += weights[i]
	}
	for j := range median {
		median[j] /= totalWeight
	}
	converged := false
	for iteration := 0; iteration < maxIterations && !converged; iteration++ {
		next := make([]float64, spatial)
		denominator := 0.0
		for i, point := range points {
			distance := 0.0
			for j := range median {
				distance += float64((point[j] - median[j]) * (point[j] - median[j]))
			}
			distance = math.Sqrt(distance)
			if distance == 0 {
				continue
			}
			for j := range next {
				next[j] += float64(point[j] * weights[i] / distance)
			}
			denominator += weights[i] / distance
		}
		if denominator == 0 {
			break
		}
		step := 0.0
		for j := range next {
			next[j] /= denominator
			step += float64((next[j] - median[j]) * (next[j] - median[j]))
		}
		median = next
		converged = math.Sqrt(step) <= tolerance
	}
	if !converged && args[3] != nil && args[3].(bool) {
		return nil, pgerror.Newf(pgcode.Internal, "Median failed to converge within %s after %d iterations.", formatG(tolerance, 6), maxIterations)
	}
	return g.withShape(shape{typeNum: pointType, ctype: s.ctype &^ dimXYM, coords: median}), nil
}

// stOrientedEnvelope implements ST_OrientedEnvelope, the rectangle of least width that holds the geometry. A geometry
// without area gives its convex hull instead.
func stOrientedEnvelope(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	geosG, err := g.Shape.toGEOS()
	if err != nil {
		return nil, err
	}
	if g.Shape.isEmpty() {
		return g.withShape(shape{typeNum: polygonType}), nil
	}
	geos := spatial.NewGEOS()
	defer geos.Close()
	rectangle, err := geos.MinimumRotatedRectangle(geosG)
	if err != nil {
		return nil, geosError("GEOSMinimumRotatedRectangle", err)
	}
	return fromGEOS(rectangle, g.SRID, false), nil
}

// enclosingRoutines returns the routines that find central points and enclosing shapes.
func enclosingRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "st_geometricmedian", Symbol: "ST_GeometricMedian", Parameters: []extdef.Parameter{{Name: "g", Type: "geometry"}, {Name: "tolerance", Type: "float8", Default: `NULL::double precision`}, {Name: "max_iter", Type: "int4", Default: `10000`}, {Name: "fail_if_not_converged", Type: "bool", Default: `false`}}, Returns: "geometry", Impl: stGeometricMedian},
		{Name: "st_orientedenvelope", Symbol: "ST_OrientedEnvelope", Parameters: params("geometry"), Returns: "geometry", Strict: true, Impl: stOrientedEnvelope},
	}
}
