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

	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// outsideSides returns the sides of the box that a point lies beyond, as a bit for each of left, right, below, and
// above.
func outsideSides(view box2d, x float64, y float64) int {
	sides := 0
	if x < view.minX {
		sides |= 1
	}
	if x > view.maxX {
		sides |= 2
	}
	if y < view.minY {
		sides |= 4
	}
	if y > view.maxY {
		sides |= 8
	}
	return sides
}

// relevantPoints returns the ordinates of the points of a sequence that can affect what is drawn inside the box. A
// point is dropped when it lies beyond the same side of the box as the point before it and the point after it. The
// point before is the last point kept, or the point just before it when none has been kept yet. The first and last
// points of a linestring have only one neighbor, and a ring's first point follows its last.
func relevantPoints(view box2d, seq []float64, dims int, closed bool) []float64 {
	count := len(seq) / dims
	if closed {
		count--
	}
	sidesAt := func(i int) int {
		return outsideSides(view, seq[i*dims], seq[i*dims+1])
	}
	var kept []float64
	lastKept, hasKept := 0, false
	if closed && count > 0 {
		lastKept, hasKept = sidesAt(count-1), true
	}
	for i := 0; i < count; i++ {
		shared := sidesAt(i)
		if hasKept {
			shared &= lastKept
		} else if i > 0 {
			shared &= sidesAt(i - 1)
		}
		if i+1 < count {
			shared &= sidesAt(i + 1)
		} else if closed {
			shared &= sidesAt(0)
		}
		if shared != 0 {
			continue
		}
		kept = append(kept, seq[i*dims:(i+1)*dims]...)
		lastKept, hasKept = sidesAt(i), true
	}
	if closed && len(kept) > 0 {
		kept = append(kept, kept[:dims]...)
	}
	return kept
}

// removeIrrelevantPoints returns the shape without the points of its linestrings and polygons that cannot affect
// what is drawn inside the box. A polygon whose exterior ring lies inside the box is kept whole, and rings and
// members left with too few points are dropped. Every other type is kept as it is.
func removeIrrelevantPoints(view box2d, s shape) shape {
	dims := s.ctype.Dimension()
	switch s.typeNum {
	case lineType:
		kept := relevantPoints(view, s.coords, dims, false)
		if len(kept) < 2*dims {
			kept = nil
		}
		s.coords = kept
	case polygonType:
		if len(s.rings) == 0 {
			return s
		}
		exterior := shape{typeNum: lineType, ctype: s.ctype, coords: s.rings[0]}
		if b, ok := exterior.bounds(); ok && b.minX >= view.minX && b.maxX <= view.maxX && b.minY >= view.minY && b.maxY <= view.maxY {
			return s
		}
		var rings [][]float64
		for i, ring := range s.rings {
			kept := relevantPoints(view, ring, dims, true)
			if len(kept) < 4*dims {
				if i == 0 {
					break
				}
				continue
			}
			rings = append(rings, kept)
		}
		s.rings = rings
	case multiLineType, multiPolygonType:
		var members []shape
		for _, member := range s.members {
			if kept := removeIrrelevantPoints(view, member); !kept.isEmpty() {
				members = append(members, kept)
			}
		}
		s.members = members
	}
	return s
}

// stRemoveIrrelevantPointsForView implements ST_RemoveIrrelevantPointsForView, which drops the points that cannot
// affect what is drawn of a geometry inside the box.
func stRemoveIrrelevantPointsForView(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	//TODO: apply the further reductions that PostGIS makes when `cartesian_hint` is true
	return g.withShape(removeIrrelevantPoints(args[1].(box2d), g.Shape)), nil
}

// viewRoutines returns the routines that prepare geometries for drawing.
func viewRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "st_removeirrelevantpointsforview", Symbol: "ST_RemoveIrrelevantPointsForView", Parameters: []extdef.Parameter{{Type: "geometry"}, {Type: "box2d"}, {Type: "bool", Default: "false"}}, Returns: "geometry", Strict: true, Impl: stRemoveIrrelevantPointsForView},
	}
}
