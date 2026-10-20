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
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// letterShape returns the shape of a character in a font that maps characters to base64-encoded TWKB, or false when
// the font lacks the character.
func letterShape(font map[string]string, character string) (shape, bool, error) {
	encoded, ok := font[character]
	if !ok {
		return shape{}, false, nil
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return shape{}, false, pgerror.New(pgcode.InvalidParameterValue, "invalid base64 end sequence")
	}
	reader := twkbReader{data: data}
	s, err := reader.read()
	if err != nil {
		return shape{}, false, err
	}
	return s, !s.isEmpty(), nil
}

// stLetters implements ST_Letters, which draws text as a multipolygon 100 units high, using a font that maps each
// character to its shape, 1000 units high, as base64-encoded TWKB. The characters are set left to right, each moved
// to start where the one before it ended, plus a gap of a twelfth of the width of "m". A character that comes farther
// than the gap from the one before it is moved back to the gap. Spaces, and characters that the font lacks, take a
// width of "m" divided by 3.5. Text without any drawn characters is null.
func stLetters(ctx *sql.Context, args ...any) (any, error) {
	if args[0] == nil {
		return nil, pgerror.New(pgcode.NullValueNotAllowed, "FOREACH expression must not be null")
	}
	font := lettersFont
	if args[1] != nil {
		document, err := args[1].(sql.JSONWrapper).ToInterface(ctx)
		if err != nil {
			return nil, err
		}
		font = make(map[string]string)
		object, _ := document.(map[string]any)
		for character, value := range object {
			if text, ok := value.(string); ok {
				font[character] = text
			} else {
				encoded, _ := json.Marshal(value)
				font[character] = string(encoded)
			}
		}
	}
	m, ok, err := letterShape(font, "m")
	if err != nil || !ok {
		return nil, err
	}
	mBounds, _ := m.bounds()
	mWidth := mBounds.maxX - mBounds.minX
	spacing := mWidth / 12
	translate := func(s shape, dx float64) shape {
		return s.mapCoords(func(point []float64) []float64 {
			point[0] += dx
			return point
		})
	}
	var polygons []shape
	var previous Geometry
	hasPrevious := false
	position := 0.0
	for _, character := range strings.ReplaceAll(args[0].(string), " ", "\t") {
		letter, ok, err := letterShape(font, string(character))
		if err != nil {
			return nil, err
		}
		width := mWidth / 3.5
		adjustment := 0.0
		if ok {
			b, _ := letter.bounds()
			width = b.maxX - b.minX
			letter = translate(translate(letter, -b.minX), position)
			if hasPrevious {
				distance, _, err := distance2D("ST_Distance", previous, Geometry{Shape: letter})
				if err != nil {
					return nil, err
				}
				if distance > spacing {
					adjustment = spacing - distance
					letter = translate(letter, adjustment)
				}
			}
			if letter.typeNum == polygonType {
				polygons = append(polygons, letter)
			} else {
				polygons = append(polygons, letter.members...)
			}
		}
		previous, hasPrevious = Geometry{Shape: letter}, ok
		position += width + spacing + adjustment
	}
	if len(polygons) == 0 {
		return nil, nil
	}
	scale := 100.0 / 1000.0
	word := shape{typeNum: multiPolygonType, members: polygons}.mapCoords(func(point []float64) []float64 {
		point[0], point[1] = point[0]*scale, point[1]*scale
		return point
	})
	return Geometry{Shape: word}, nil
}

// lettersRoutines returns the routines that draw text.
func lettersRoutines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "st_letters", Symbol: "st_letters_text_json", Parameters: []extdef.Parameter{{Name: "letters", Type: "text"}, {Name: "font", Type: "json", Default: "NULL::json"}}, Returns: "geometry", Impl: stLetters},
	}
}
