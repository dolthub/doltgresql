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

package spatial

// #cgo CXXFLAGS: -std=c++14
// #include <stdint.h>
// #include <stdlib.h>
// int spatial_wagyu_clip(const int64_t* coords, const int32_t* ring_sizes, int32_t ring_count, int64_t min_x,
//     int64_t min_y, int64_t max_x, int64_t max_y, int64_t** out_coords, int32_t** out_ring_sizes,
//     int32_t** out_polygon_sizes, int32_t* out_polygon_count, char** error);
import "C"

import (
	"errors"
	"unsafe"
)

// WagyuClip clips rings of integer points to a box with Wagyu, filling them by the even-odd rule, and returns the
// resulting polygons as their rings. Each ring holds its x and y values one after the other, and the first ring of
// each polygon is its exterior.
func WagyuClip(rings [][]int64, minX int64, minY int64, maxX int64, maxY int64) ([][][]int64, error) {
	var coords []C.int64_t
	sizes := make([]C.int32_t, len(rings))
	for i, ring := range rings {
		sizes[i] = C.int32_t(len(ring) / 2)
		for _, v := range ring {
			coords = append(coords, C.int64_t(v))
		}
	}
	coords = append(coords, 0)
	sizes = append(sizes, 0)
	var outCoords *C.int64_t
	var outRingSizes, outPolygonSizes *C.int32_t
	var outPolygonCount C.int32_t
	var message *C.char
	if C.spatial_wagyu_clip(&coords[0], &sizes[0], C.int32_t(len(rings)), C.int64_t(minX), C.int64_t(minY), C.int64_t(maxX),
		C.int64_t(maxY), &outCoords, &outRingSizes, &outPolygonSizes, &outPolygonCount, &message) != 0 {
		defer C.free(unsafe.Pointer(message))
		return nil, errors.New(C.GoString(message))
	}
	defer C.free(unsafe.Pointer(outCoords))
	defer C.free(unsafe.Pointer(outRingSizes))
	defer C.free(unsafe.Pointer(outPolygonSizes))
	polygonSizes := unsafe.Slice(outPolygonSizes, int(outPolygonCount)+1)
	totalRings := 0
	for _, size := range polygonSizes[:outPolygonCount] {
		totalRings += int(size)
	}
	ringSizes := unsafe.Slice(outRingSizes, totalRings+1)
	totalPoints := 0
	for _, size := range ringSizes[:totalRings] {
		totalPoints += int(size)
	}
	values := unsafe.Slice(outCoords, 2*totalPoints+1)
	polygons := make([][][]int64, outPolygonCount)
	ringIndex, valueIndex := 0, 0
	for i := range polygons {
		polygons[i] = make([][]int64, polygonSizes[i])
		for j := range polygons[i] {
			ring := make([]int64, 2*int(ringSizes[ringIndex]))
			for k := range ring {
				ring[k] = int64(values[valueIndex])
				valueIndex++
			}
			polygons[i][j] = ring
			ringIndex++
		}
	}
	return polygons, nil
}
