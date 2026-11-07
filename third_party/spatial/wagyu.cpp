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

// Fused multiply-adds round differently from separate operations, so they are disabled to compute the same results on
// every platform. cgo does not accept the compiler flag that does this.
#if defined(__clang__)
#pragma clang fp contract(off)
#elif defined(__GNUC__)
#pragma GCC optimize("fp-contract=off")
#endif

#include <cstdint>
#include <cstdlib>
#include <cstring>
#include <exception>

#include <mapbox/geometry/wagyu/quick_clip.hpp>
#include <mapbox/geometry/wagyu/wagyu.hpp>

extern "C" {

// spatial_wagyu_clip clips rings of integer points to a box with Wagyu, filling them by the even-odd rule, and returns
// the resulting polygons. The input holds ring_count rings, whose point counts are in ring_sizes and whose x and y
// values follow each other in coords. The output is allocated with malloc and holds each polygon's ring count in
// out_polygon_sizes, each ring's point count in out_ring_sizes, and the points in out_coords. It returns zero on
// success, or one with a message in error, which is also allocated with malloc.
int spatial_wagyu_clip(const int64_t* coords, const int32_t* ring_sizes, int32_t ring_count, int64_t min_x, int64_t min_y,
                       int64_t max_x, int64_t max_y, int64_t** out_coords, int32_t** out_ring_sizes,
                       int32_t** out_polygon_sizes, int32_t* out_polygon_count, char** error) {
    using namespace mapbox::geometry;
    try {
        mapbox::geometry::polygon<int64_t> rings;
        for (int32_t i = 0; i < ring_count; i++) {
            linear_ring<int64_t> ring;
            for (int32_t j = 0; j < ring_sizes[i]; j++, coords += 2) {
                ring.push_back(point<int64_t>(coords[0], coords[1]));
            }
            rings.push_back(ring);
        }
        multi_polygon<int64_t> input;
        input.push_back(rings);
        box<int64_t> clip_box(point<int64_t>(min_x, min_y), point<int64_t>(max_x, max_y));
        multi_polygon<int64_t> result = wagyu::clip(input, clip_box, wagyu::fill_type_even_odd);
        size_t total_rings = 0, total_points = 0;
        for (auto const& polygon : result) {
            total_rings += polygon.size();
            for (auto const& ring : polygon) {
                total_points += ring.size();
            }
        }
        *out_polygon_count = static_cast<int32_t>(result.size());
        *out_polygon_sizes = static_cast<int32_t*>(malloc(sizeof(int32_t) * (result.size() + 1)));
        *out_ring_sizes = static_cast<int32_t*>(malloc(sizeof(int32_t) * (total_rings + 1)));
        *out_coords = static_cast<int64_t*>(malloc(sizeof(int64_t) * (2 * total_points + 1)));
        size_t ring_index = 0, coord_index = 0;
        for (size_t i = 0; i < result.size(); i++) {
            (*out_polygon_sizes)[i] = static_cast<int32_t>(result[i].size());
            for (auto const& ring : result[i]) {
                (*out_ring_sizes)[ring_index++] = static_cast<int32_t>(ring.size());
                for (auto const& p : ring) {
                    (*out_coords)[coord_index++] = p.x;
                    (*out_coords)[coord_index++] = p.y;
                }
            }
        }
        return 0;
    } catch (std::exception const& e) {
        *error = strdup(e.what());
        return 1;
    }
}
}
