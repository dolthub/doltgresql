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
	"bytes"
	"encoding/binary"
	"math"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// spheroid is the value of the spheroid type: a named ellipsoid given by its semi-major axis and its flattening.
type spheroid struct {
	name string
	a, f float64
}

// spheroidCodec stores a spheroid as its semi-major axis and flattening, followed by its name.
var spheroidCodec = &pgtypes.TypeCodec{
	Serialize: func(val any) ([]byte, error) {
		s := val.(spheroid)
		return append(appendFloats(nil, []float64{s.a, s.f}, binary.LittleEndian), s.name...), nil
	},
	Deserialize: func(data []byte) (any, error) {
		if len(data) < 16 {
			return nil, errors.Errorf("malformed serialized spheroid of %d bytes", len(data))
		}
		return spheroid{
			name: string(data[16:]),
			a:    math.Float64frombits(binary.LittleEndian.Uint64(data)),
			f:    math.Float64frombits(binary.LittleEndian.Uint64(data[8:])),
		}, nil
	},
	SerializedCompare: func(left []byte, right []byte) (int, error) {
		return bytes.Compare(left, right), nil
	},
}

// spheroidIn implements spheroid_in, which reads `SPHEROID["name",a,rf]`, where `a` is the semi-major axis and `rf`
// is the inverse flattening. The name holds at most 19 characters.
func spheroidIn(ctx *sql.Context, args ...any) (any, error) {
	text := args[0].(string)
	if !strings.HasPrefix(text, "SPHEROID") {
		return nil, pgerror.New(pgcode.Internal, "SPHEROID parser - doesn't start with SPHEROID")
	}
	rest, ok := strings.CutPrefix(text, `SPHEROID["`)
	nameLength := strings.IndexByte(rest, '"')
	if !ok || nameLength < 1 || nameLength > 19 {
		return nil, pgerror.New(pgcode.Internal, "SPHEROID parser - couldn't parse the spheroid")
	}
	vals := scanFormat(rest[nameLength:], `",%lf,%lf]`)
	if len(vals) < 2 {
		return nil, pgerror.New(pgcode.Internal, "SPHEROID parser - couldn't parse the spheroid")
	}
	return spheroid{name: rest[:nameLength], a: vals[0], f: 1 / vals[1]}, nil
}

// spheroidOut implements spheroid_out.
func spheroidOut(ctx *sql.Context, args ...any) (any, error) {
	s := args[0].(spheroid)
	return `SPHEROID("` + s.name + `",` + formatG(s.a, 15) + "," + formatG(1/s.f, 15) + ")", nil
}

// stLengthSpheroid implements ST_LengthSpheroid and ST_Length2DSpheroid, which sum the length in meters of every
// linestring and ring on the spheroid, including the change in Z when there is one.
func stLengthSpheroid(ctx *sql.Context, args ...any) (any, error) {
	ellipsoid := args[1].(spheroid).ellipsoid()
	var length func(s shape) (float64, error)
	length = func(s shape) (float64, error) {
		dims := s.ctype.Dimension()
		total := 0.0
		addSequence := func(seq []float64) {
			for i := dims; i+dims <= len(seq); i += dims {
				d := spheroidDistance(ellipsoid, seq[i-dims:], seq[i:])
				if s.hasZ() {
					dz := seq[i+2] - seq[i-dims+2]
					d = math.Sqrt(float64(d*d) + float64(dz*dz))
				}
				total += d
			}
		}
		switch s.typeNum {
		case lineType, triangleType:
			addSequence(s.coords)
		case polygonType:
			for _, ring := range s.rings {
				addSequence(ring)
			}
		case circStringType:
			return 0, pgerror.New(pgcode.Internal, "unsupported type passed to lwgeom_length_spheroid")
		case pointType, curvePolyType:
		default:
			for _, member := range s.members {
				memberLength, err := length(member)
				if err != nil {
					return 0, err
				}
				total += memberLength
			}
		}
		return total, nil
	}
	return length(args[0].(Geometry).Shape)
}

// stDistanceSpheroid implements ST_DistanceSpheroid, which returns the distance in meters between the closest points
// of two geometries on the spheroid, or on the WGS84 spheroid when none is given. It returns -1 when either geometry
// is empty.
func stDistanceSpheroid(ctx *sql.Context, args ...any) (any, error) {
	a, b := args[0].(Geometry), args[1].(Geometry)
	if err := checkSameSRID("geometry_distance_spheroid", a, b); err != nil {
		return nil, err
	}
	for _, s := range []shape{a.Shape, b.Shape} {
		switch s.typeNum {
		case pointType, lineType, polygonType, multiPointType, multiLineType, multiPolygonType:
		default:
			return nil, pgerror.New(pgcode.Internal, "geometry_distance_spheroid: Only point/line/polygon supported.")
		}
	}
	ellipsoid := spatial.WGS84
	if len(args) > 2 {
		ellipsoid = args[2].(spheroid).ellipsoid()
	}
	pa, pb, ok := sphericalClosest(a.Shape, b.Shape)
	if !ok {
		return -1.0, nil
	}
	return spheroidDistance(ellipsoid, pa, pb), nil
}

// ellipsoid returns the geodesic calculator for the spheroid.
func (s spheroid) ellipsoid() *spatial.Ellipsoid {
	return spatial.NewEllipsoid(s.a, s.f)
}

// spheroidRoutines returns the routines of the spheroid type.
func spheroidRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "spheroid_in", Symbol: "ellipsoid_in", Parameters: params("cstring"), Returns: "spheroid", Strict: true, Impl: spheroidIn},
		{Name: "spheroid_out", Symbol: "ellipsoid_out", Parameters: params("spheroid"), Returns: "cstring", Strict: true, Impl: spheroidOut},
		{Name: "st_distancespheroid", Symbol: "st_distancespheroid_geometry_geometry", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}}, Returns: "float8", Strict: true, Impl: stDistanceSpheroid},
		{Name: "st_distancespheroid", Symbol: "st_distancespheroid_geometry_geometry_spheroid", Parameters: []extdef.Parameter{{Name: "geom1", Type: "geometry"}, {Name: "geom2", Type: "geometry"}, {Type: "spheroid"}}, Returns: "float8", Strict: true, Impl: stDistanceSpheroid},
		{Name: "st_length2dspheroid", Symbol: "LWGEOM_length2d_ellipsoid", Parameters: params("geometry", "spheroid"), Returns: "float8", Strict: true, Impl: stLengthSpheroid},
		{Name: "st_lengthspheroid", Symbol: "LWGEOM_length_ellipsoid_linestring", Parameters: params("geometry", "spheroid"), Returns: "float8", Strict: true, Impl: stLengthSpheroid},
	}
}
