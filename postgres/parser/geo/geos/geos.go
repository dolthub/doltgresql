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

// Package geos is a wrapper around the spatial data types between the geo
// package and the GEOS C library. The GEOS library is dynamically loaded
// at init time.
// Operations will error if the GEOS library was not found.
package geos

import (
	"github.com/dolthub/doltgresql/postgres/parser/geo/geopb"
)

// WKTToEWKB parses a WKT into WKB using the GEOS library.
func WKTToEWKB(wkt geopb.WKT, srid geopb.SRID) (geopb.EWKB, error) {
	panic("WKT to EWKB support is not yet implemented")
}
