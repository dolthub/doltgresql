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

package geomfn

import (
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/postgres/parser/geo"
	"github.com/dolthub/doltgresql/postgres/parser/geo/geos"
)

// BufferParams is a wrapper around the geos.BufferParams.
type BufferParams struct {
	p geos.BufferParams
}

// MakeDefaultBufferParams returns the default BufferParams/
func MakeDefaultBufferParams() BufferParams {
	return BufferParams{
		p: geos.BufferParams{
			EndCapStyle:      geos.BufferParamsEndCapStyleRound,
			JoinStyle:        geos.BufferParamsJoinStyleRound,
			SingleSided:      false,
			QuadrantSegments: 8,
			MitreLimit:       5.0,
		},
	}
}

// WithQuadrantSegments returns a copy of the BufferParams with the quadrantSegments set.
func (b BufferParams) WithQuadrantSegments(quadrantSegments int) BufferParams {
	ret := b
	ret.p.QuadrantSegments = quadrantSegments
	return ret
}

// ParseBufferParams parses the given buffer params from a SQL string into
// the BufferParams form.
// The string must be of the same format as specified by https://postgis.net/docs/ST_Buffer.html.
// Returns the BufferParams, as well as the modified distance.
func ParseBufferParams(s string, distance float64) (BufferParams, float64, error) {
	p := MakeDefaultBufferParams()
	fields := strings.Fields(s)
	for _, field := range fields {
		fParams := strings.Split(field, "=")
		if len(fParams) != 2 {
			return BufferParams{}, 0, errors.Newf("unknown buffer parameter: %s", fParams)
		}
		f, val := fParams[0], fParams[1]
		switch strings.ToLower(f) {
		case "quad_segs":
			valInt, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return BufferParams{}, 0, errors.Wrapf(err, "invalid int for %s: %s", f, val)
			}
			p.p.QuadrantSegments = int(valInt)
		case "endcap":
			switch strings.ToLower(val) {
			case "round":
				p.p.EndCapStyle = geos.BufferParamsEndCapStyleRound
			case "flat", "butt":
				p.p.EndCapStyle = geos.BufferParamsEndCapStyleFlat
			case "square":
				p.p.EndCapStyle = geos.BufferParamsEndCapStyleSquare
			default:
				return BufferParams{}, 0, errors.Newf("unknown endcap: %s (accepted: round, flat, square)", val)
			}
		case "join":
			switch strings.ToLower(val) {
			case "round":
				p.p.JoinStyle = geos.BufferParamsJoinStyleRound
			case "mitre", "miter":
				p.p.JoinStyle = geos.BufferParamsJoinStyleMitre
			case "bevel":
				p.p.JoinStyle = geos.BufferParamsJoinStyleBevel
			default:
				return BufferParams{}, 0, errors.Newf("unknown join: %s (accepted: round, mitre, bevel)", val)
			}
		case "mitre_limit", "miter_limit":
			valFloat, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return BufferParams{}, 0, errors.Wrapf(err, "invalid float for %s: %s", f, val)
			}
			p.p.MitreLimit = valFloat
		case "side":
			switch strings.ToLower(val) {
			case "both":
				p.p.SingleSided = false
			case "left":
				p.p.SingleSided = true
			case "right":
				p.p.SingleSided = true
				distance *= -1
			default:
				return BufferParams{}, 0, errors.Newf("unknown side: %s (accepted: both, left, right)", val)
			}
		default:
			return BufferParams{}, 0, errors.Newf("unknown field: %s (accepted fields: quad_segs, endcap, join, mitre_limit, side)", f)
		}
	}
	return p, distance, nil
}

// Buffer buffers a given Geometry by the supplied parameters.
func Buffer(g geo.Geometry, params BufferParams, distance float64) (geo.Geometry, error) {
	bufferedGeom, err := geos.Buffer(g.EWKB(), params.p, distance)
	if err != nil {
		return geo.Geometry{}, err
	}
	return geo.ParseGeometryFromEWKB(bufferedGeom)
}
