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

import "github.com/cockroachdb/errors"

// ErrNativeUnavailable is returned, or panicked with by functions that cannot return an error, when the binary was
// built without the native GEOS and PROJ libraries, which are only linked when building with the postgis_native tag.
var ErrNativeUnavailable = errors.New("this PostGIS function requires the native spatial libraries, which this build does not include")

// GeometryType is the type of a Geometry, numbered as GEOS numbers them.
type GeometryType int

// The geometry types that GEOS supports.
const (
	Point GeometryType = iota
	LineString
	LinearRing
	Polygon
	MultiPoint
	MultiLineString
	MultiPolygon
	GeometryCollection
)

// Geometry is a geometry that GEOS can operate on. Points, linestrings, and linear rings hold their interleaved
// ordinates in `Coords`, with three ordinates per position when `HasZ` is set and two otherwise. Polygons hold their
// rings in `Members`, the shell first, and collections hold their members there. An empty geometry holds neither.
type Geometry struct {
	Type    GeometryType
	HasZ    bool
	Coords  []float64
	Members []Geometry
}

// The end cap and join styles of a buffer, numbered as GEOS numbers them.
const (
	BufferRound  = 1
	BufferFlat   = 2
	BufferSquare = 3
	BufferMitre  = 2
	BufferBevel  = 3
)

// BufferParams holds the options of a buffer: the number of segments per quarter circle, the end cap and join
// styles, the mitre limit, and whether only one side is buffered, which is the left side for positive distances.
type BufferParams struct {
	QuadSegs    int
	EndCap      int
	Join        int
	MitreLimit  float64
	SingleSided bool
}

// CRSKind is the kind of a coordinate reference system.
type CRSKind int

// The kinds of coordinate reference systems in PROJ's database that the PostGIS functions list.
const (
	OtherCRS CRSKind = iota
	GeocentricCRS
	Geographic2DCRS
	Geographic3DCRS
	ProjectedCRS
	CompoundCRS
)

// CRSInfo describes a coordinate reference system in PROJ's database, along with the longitude and latitude bounds of
// its area of use when it has one.
type CRSInfo struct {
	Authority  string
	Code       string
	Name       string
	Kind       CRSKind
	Deprecated bool
	AreaValid  bool
	West       float64
	South      float64
	East       float64
	North      float64
}

// TransformError is an error that PROJ reported while transforming coordinates.
type TransformError struct {
	Message string
	Number  int
}

// Error implements the interface error.
func (e TransformError) Error() string {
	return e.Message
}
