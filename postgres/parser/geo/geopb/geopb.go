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

package geopb

import "fmt"

// EWKBHex returns the EWKB-hex version of this data type
func (b *SpatialObject) EWKBHex() string {
	return fmt.Sprintf("%X", b.EWKB)
}

// MultiType returns the corresponding multi-type for a shape type, or unset
// if there is no multi-type.
func (s ShapeType) MultiType() ShapeType {
	switch s {
	case ShapeType_Unset:
		return ShapeType_Unset
	case ShapeType_Point, ShapeType_MultiPoint:
		return ShapeType_MultiPoint
	case ShapeType_LineString, ShapeType_MultiLineString:
		return ShapeType_MultiLineString
	case ShapeType_Polygon, ShapeType_MultiPolygon:
		return ShapeType_MultiPolygon
	case ShapeType_GeometryCollection:
		return ShapeType_GeometryCollection
	default:
		return ShapeType_Unset
	}
}
