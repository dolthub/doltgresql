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
	"fmt"
	"math"
	"sort"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// stIsValidTrajectory implements ST_IsValidTrajectory, which returns whether a linestring's M coordinates strictly
// increase, sending a notice that explains a false result.
func stIsValidTrajectory(ctx *sql.Context, args ...any) (any, error) {
	line := args[0].(Geometry).Shape
	if line.typeNum != lineType {
		sendNotice(ctx, "NOTICE", "Geometry is not a LINESTRING")
		return false, nil
	}
	if !line.hasM() {
		sendNotice(ctx, "NOTICE", "Line does not have M dimension")
		return false, nil
	}
	dims := line.ctype.Dimension()
	for i := 2*dims - 1; i < len(line.coords); i += dims {
		if line.coords[i] <= line.coords[i-dims] {
			vertex := i / dims
			sendNotice(ctx, "NOTICE", fmt.Sprintf("Measure of vertex %d (%s) not bigger than measure of vertex %d (%s)",
				vertex, formatG(line.coords[i], 6), vertex-1, formatG(line.coords[i-dims], 6)))
			return false, nil
		}
	}
	return true, nil
}

// trackPosition returns the position of a trajectory at the time, interpolated between the vertices whose M
// coordinates surround it, using X, Y, and Z when the trajectory has them.
func trackPosition(track shape, time float64) []float64 {
	dims := track.ctype.Dimension()
	position := func(i int) []float64 {
		if track.hasZ() {
			return track.coords[i : i+3]
		}
		return track.coords[i : i+2]
	}
	for i := dims; i+dims <= len(track.coords); i += dims {
		start, end := track.coords[i-1], track.coords[i+dims-1]
		if time >= start && time <= end {
			fraction := 0.0
			if end > start {
				fraction = (time - start) / (end - start)
			}
			return interpolatePoint(position(i-dims), position(i), fraction)
		}
	}
	return position(len(track.coords) - dims)
}

// closestApproach returns the time at which two trajectories come closest within the time range they share, and the
// distance between them then, or false when they share no time.
func closestApproach(a shape, b shape) (float64, float64, bool) {
	measures := func(s shape) []float64 {
		dims := s.ctype.Dimension()
		var times []float64
		for i := dims - 1; i < len(s.coords); i += dims {
			times = append(times, s.coords[i])
		}
		return times
	}
	timesA, timesB := measures(a), measures(b)
	if len(timesA) == 0 || len(timesB) == 0 {
		return 0, 0, false
	}
	start, end := math.Max(timesA[0], timesB[0]), math.Min(timesA[len(timesA)-1], timesB[len(timesB)-1])
	if start > end {
		return 0, 0, false
	}
	times := []float64{start, end}
	for _, t := range append(timesA, timesB...) {
		if t > start && t < end {
			times = append(times, t)
		}
	}
	sort.Float64s(times)
	bestTime, bestDistance := start, math.Inf(1)
	distanceAt := func(t float64) float64 {
		pa, pb := trackPosition(a, t), trackPosition(b, t)
		sum := 0.0
		for i := range pa {
			sum += float64((pa[i] - pb[i]) * (pa[i] - pb[i]))
		}
		return math.Sqrt(sum)
	}
	for i := 0; i+1 < len(times) || i == 0; i++ {
		t0, t1 := times[i], times[min(i+1, len(times)-1)]
		candidates := []float64{t0}
		if t1 > t0 {
			p0a, p0b, p1a, p1b := trackPosition(a, t0), trackPosition(b, t0), trackPosition(a, t1), trackPosition(b, t1)
			var dot, speed float64
			for j := range p0a {
				offset := p0a[j] - p0b[j]
				velocity := (p1a[j] - p0a[j]) - (p1b[j] - p0b[j])
				dot += float64(offset * velocity)
				speed += float64(velocity * velocity)
			}
			if speed > 0 {
				if fraction := -dot / speed; fraction > 0 && fraction < 1 {
					candidates = append(candidates, t0+float64(fraction*(t1-t0)))
				}
			}
			candidates = append(candidates, t1)
		}
		for _, t := range candidates {
			if d := distanceAt(t); d < bestDistance {
				bestTime, bestDistance = t, d
			}
		}
		if len(times) == 1 {
			break
		}
	}
	return bestTime, bestDistance, true
}

// trajectoryArgs checks that both arguments are measured linestrings.
func trajectoryArgs(args []any) (shape, shape, error) {
	a, b := args[0].(Geometry).Shape, args[1].(Geometry).Shape
	if !a.hasM() || !b.hasM() {
		return shape{}, shape{}, pgerror.New(pgcode.Internal, "Both input geometries must have a measure dimension")
	}
	if a.typeNum != lineType || b.typeNum != lineType {
		return shape{}, shape{}, pgerror.New(pgcode.Internal, "Both input geometries must be linestrings")
	}
	return a, b, nil
}

// cpaImpl returns an implementation of a function of the closest point of approach of two trajectories, which is null
// when they share no time.
func cpaImpl(result func(time float64, distance float64, args []any) any) extdef.Function {
	return func(ctx *sql.Context, args ...any) (any, error) {
		a, b, err := trajectoryArgs(args)
		if err != nil {
			return nil, err
		}
		time, distance, ok := closestApproach(a, b)
		if !ok {
			return nil, nil
		}
		return result(time, distance, args), nil
	}
}

// crossingSide returns the side of the segment from `a` to `b` that `p` lies on: one for the left, minus one for the
// right, and zero on the line.
func crossingSide(a []float64, b []float64, p []float64) int {
	cross := float64((b[0]-a[0])*(p[1]-a[1])) - float64((b[1]-a[1])*(p[0]-a[0]))
	if cross > 0 {
		return 1
	}
	if cross < 0 {
		return -1
	}
	return 0
}

// stLineCrossingDirection implements ST_LineCrossingDirection, which describes how the second linestring crosses the
// first. A single crossing from the left of the first linestring to its right is one, and from its right to its left
// minus one. More crossings give three times the direction of the first when there is an even number, ending on the
// starting side, and two times the direction of the last otherwise. Lines that only touch do not cross.
func stLineCrossingDirection(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(Geometry), args[1].(Geometry)
	if err := checkSameSRID("ST_LineCrossingDirection", a, b); err != nil {
		return nil, err
	}
	if a.Shape.typeNum != lineType || b.Shape.typeNum != lineType {
		return nil, pgerror.New(pgcode.Internal, "Geometry types must both be LINESTRING")
	}
	var directions []int
	for _, sb := range segmentsOf(b.Shape) {
		type crossing struct {
			along     float64
			direction int
		}
		var crossings []crossing
		for _, sa := range segmentsOf(a.Shape) {
			startSide, endSide := crossingSide(sa[0], sa[1], sb[0]), crossingSide(sa[0], sa[1], sb[1])
			if startSide*endSide >= 0 || crossingSide(sb[0], sb[1], sa[0])*crossingSide(sb[0], sb[1], sa[1]) >= 0 {
				continue
			}
			point, ok := segmentCrossing(sb[0], sb[1], sa[0], sa[1])
			if !ok {
				continue
			}
			crossings = append(crossings, crossing{along: math.Hypot(point[0]-sb[0][0], point[1]-sb[0][1]), direction: startSide})
		}
		sort.SliceStable(crossings, func(i, j int) bool { return crossings[i].along < crossings[j].along })
		for _, c := range crossings {
			directions = append(directions, c.direction)
		}
	}
	switch {
	case len(directions) == 0:
		return int32(0), nil
	case len(directions) == 1:
		return int32(directions[0]), nil
	case len(directions)%2 == 0:
		return int32(3 * directions[0]), nil
	default:
		return int32(2 * directions[len(directions)-1]), nil
	}
}

// trajectoryRoutines returns the routines that work with trajectories and crossings.
func trajectoryRoutines() []extdef.Routine {
	distanceCPA := cpaImpl(func(time float64, distance float64, args []any) any {
		return distance
	})
	return []extdef.Routine{
		{Name: "_st_linecrossingdirection", Symbol: "_st_linecrossingdirection_geometry_geometry", Parameters: []extdef.Parameter{{Name: "line1", Type: "geometry"}, {Name: "line2", Type: "geometry"}}, Returns: "int4", Strict: true, Impl: stLineCrossingDirection},
		{Name: "geometry_distance_cpa", Symbol: "geometry_distance_cpa_geometry_geometry", Parameters: params("geometry", "geometry"), Returns: "float8", Strict: true, Impl: distanceCPA},
		{Name: "st_closestpointofapproach", Symbol: "ST_ClosestPointOfApproach", Parameters: params("geometry", "geometry"), Returns: "float8", Strict: true, Impl: cpaImpl(func(time float64, distance float64, args []any) any {
			return time
		})},
		{Name: "st_cpawithin", Symbol: "ST_CPAWithin", Parameters: params("geometry", "geometry", "float8"), Returns: "bool", Strict: true, Impl: cpaImpl(func(time float64, distance float64, args []any) any {
			return distance <= args[2].(float64)
		})},
		{Name: "st_distancecpa", Symbol: "st_distancecpa_geometry_geometry", Parameters: params("geometry", "geometry"), Returns: "float8", Strict: true, Impl: distanceCPA},
		{Name: "st_isvalidtrajectory", Symbol: "ST_IsValidTrajectory", Parameters: params("geometry"), Returns: "bool", Strict: true, Impl: stIsValidTrajectory},
		{Name: "st_linecrossingdirection", Symbol: "st_linecrossingdirection_geometry_geometry", Parameters: []extdef.Parameter{{Name: "line1", Type: "geometry"}, {Name: "line2", Type: "geometry"}}, Returns: "int4", Strict: true, Impl: stLineCrossingDirection},
	}
}
