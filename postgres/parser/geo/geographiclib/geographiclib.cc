// Copyright 2023 Dolthub, Inc.
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

// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

#include "geodesic.h"
#include "geographiclib.h"

void CR_GEOGRAPHICLIB_InverseBatch(
  struct geod_geodesic* spheroid,
  double lats[],
  double lngs[],
  int len,
  double *result
) {
  *result = 0;
  for (int i = 0; i < len - 1; i++) {
    double s12, az1, az2;
    geod_inverse(spheroid, lats[i], lngs[i], lats[i+1], lngs[i+1], &s12, &az1, &az2);
    *result += s12;
  }
}
