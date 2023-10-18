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

package geogfn

import (
	"github.com/cockroachdb/errors"
	"github.com/golang/geo/s1"

	"github.com/dolthub/doltgresql/postgres/parser/geo"
)

// DWithin returns whether a is within distance d of b. If A or B contains empty
// Geography objects, this will return false. If inclusive, DWithin is
// equivalent to Distance(a, b) <= d. Otherwise, DWithin is instead equivalent
// to Distance(a, b) < d.
func DWithin(
	a geo.Geography,
	b geo.Geography,
	distance float64,
	useSphereOrSpheroid UseSphereOrSpheroid,
	exclusivity geo.FnExclusivity,
) (bool, error) {
	if a.SRID() != b.SRID() {
		return false, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	}
	if distance < 0 {
		return false, errors.Newf("dwithin distance cannot be less than zero")
	}
	spheroid, err := a.Spheroid()
	if err != nil {
		return false, err
	}

	angleToExpand := s1.Angle(distance / spheroid.SphereRadius)
	if useSphereOrSpheroid == UseSpheroid {
		angleToExpand *= (1 + SpheroidErrorFraction)
	}
	if !a.BoundingCap().Expanded(angleToExpand).Intersects(b.BoundingCap()) {
		return false, nil
	}

	aRegions, err := a.AsS2(geo.EmptyBehaviorError)
	if err != nil {
		if geo.IsEmptyGeometryError(err) {
			return false, nil
		}
		return false, err
	}
	bRegions, err := b.AsS2(geo.EmptyBehaviorError)
	if err != nil {
		if geo.IsEmptyGeometryError(err) {
			return false, nil
		}
		return false, err
	}
	maybeClosestDistance, err := distanceGeographyRegions(
		spheroid,
		useSphereOrSpheroid,
		aRegions,
		bRegions,
		a.BoundingRect().Intersects(b.BoundingRect()),
		distance,
		exclusivity,
	)
	if err != nil {
		return false, err
	}
	if exclusivity == geo.FnExclusive {
		return maybeClosestDistance < distance, nil
	}
	return maybeClosestDistance <= distance, nil
}
