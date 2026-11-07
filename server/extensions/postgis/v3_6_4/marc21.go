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
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	"github.com/dolthub/doltgresql/third_party/spatial"
)

// marc21Namespace is the namespace of MARC21/XML documents.
const marc21Namespace = "http://www.loc.gov/MARC21/slim"

// marc21Format is how ST_AsMARC21 writes each coordinate: with a hemisphere letter or a sign, in degrees alone, in
// degrees and minutes, or in degrees, minutes, and seconds, with the last part carrying the decimals.
type marc21Format struct {
	hemisphere bool
	parts      int
	decimals   int
}

// parseMARC21Format reads a format such as `hdddmmss` or `ddd.dddddd`, returning false when it is not valid.
func parseMARC21Format(text string) (marc21Format, bool) {
	var f marc21Format
	if rest, ok := strings.CutPrefix(text, "h"); ok {
		f.hemisphere, text = true, rest
	}
	whole, fraction, hasFraction := strings.Cut(text, ".")
	switch whole {
	case "ddd":
		f.parts = 1
	case "dddmm":
		f.parts = 2
	case "dddmmss":
		f.parts = 3
	default:
		return f, false
	}
	if hasFraction {
		letter := whole[len(whole)-1:]
		if fraction == "" || strings.Trim(fraction, letter) != "" {
			return f, false
		}
		f.decimals = len(fraction)
	}
	return f, true
}

// format writes a longitude or latitude in the format. A part that rounds up to 60 carries into the part before it.
func (f marc21Format) format(value float64, longitude bool) string {
	prefix := ""
	if f.hemisphere {
		prefix = map[bool]string{true: "N", false: "S"}[value >= 0]
		if longitude {
			prefix = map[bool]string{true: "E", false: "W"}[value >= 0]
		}
	} else if value < 0 {
		prefix = "-"
	}
	value = math.Abs(value)
	scale := math.Pow(10, float64(f.decimals))
	round := func(v float64) float64 {
		return math.Round(v*scale) / scale
	}
	last := func(v float64, width int) string {
		if f.decimals > 0 {
			width += f.decimals + 1
		}
		return fmt.Sprintf("%0*.*f", width, f.decimals, v)
	}
	degrees := math.Floor(value)
	switch f.parts {
	case 1:
		return prefix + last(round(value), 3)
	case 2:
		minutes := round((value - degrees) * 60)
		if minutes >= 60 {
			degrees, minutes = degrees+1, minutes-60
		}
		return prefix + fmt.Sprintf("%03d", int(degrees)) + last(minutes, 2)
	default:
		totalMinutes := (value - degrees) * 60
		minutes := math.Floor(totalMinutes)
		seconds := round((totalMinutes - minutes) * 60)
		if seconds >= 60 {
			minutes, seconds = minutes+1, seconds-60
		}
		if minutes >= 60 {
			degrees, minutes = degrees+1, minutes-60
		}
		return prefix + fmt.Sprintf("%03d%02d", int(degrees), int(minutes)) + last(seconds, 2)
	}
}

// stAsMARC21 implements ST_AsMARC21, which writes the bounding box of a geometry in longitude and latitude as a
// MARC21/XML record, with a datafield for each member of a collection.
func stAsMARC21(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	g, formatText := args[0].(Geometry), args[1].(string)
	if g.SRID == sridUnknown {
		return nil, pgerror.New(pgcode.Internal, "ST_AsMARC21: Input geometry has unknown (0) SRID")
	}
	definitions, err := lookUpSRS(ctx, runner, g.SRID)
	if err != nil {
		return nil, err
	}
	geographic := false
	for _, definition := range definitions {
		if definition != "" {
			geographic = spatial.IsGeographic(definition)
			break
		}
	}
	if !geographic {
		return nil, pgerror.Newf(pgcode.Internal, "ST_AsMARC21: Unsupported SRID (%d). Only lon/lat coordinate systems are supported in MARC21/XML Documents.", g.SRID)
	}
	f, ok := parseMARC21Format(formatText)
	if !ok {
		return nil, pgerror.Newf(pgcode.Internal, "invalid output format: \"%s\"", formatText)
	}
	if g.Shape.isEmpty() {
		return nil, nil
	}
	parts := []shape{g.Shape}
	if len(g.Shape.members) > 0 && g.Shape.typeNum != polygonType {
		parts = g.Shape.members
	}
	var sb strings.Builder
	sb.WriteString(`<record xmlns="` + marc21Namespace + `">`)
	for _, part := range parts {
		b, ok := part.bounds()
		if !ok {
			continue
		}
		sb.WriteString(`<datafield tag="034" ind1="1" ind2=" "><subfield code="a">a</subfield>`)
		for _, subfield := range []struct {
			code      string
			value     float64
			longitude bool
		}{{"d", b.minX, true}, {"e", b.maxX, true}, {"f", b.maxY, false}, {"g", b.minY, false}} {
			sb.WriteString(`<subfield code="` + subfield.code + `">` + f.format(subfield.value, subfield.longitude) + `</subfield>`)
		}
		sb.WriteString(`</datafield>`)
	}
	sb.WriteString(`</record>`)
	return sb.String(), nil
}

// parseMARC21Coordinate reads a coordinate written with a hemisphere letter or sign, then degrees in three digits,
// optionally followed by two digits of minutes and two of seconds, where the last part may carry decimals.
func parseMARC21Coordinate(literal string) (float64, bool) {
	text := strings.TrimSpace(literal)
	sign := 1.0
	if text != "" && strings.ContainsRune("NE+SW-", rune(text[0])) {
		if strings.ContainsRune("SW-", rune(text[0])) {
			sign = -1
		}
		text = text[1:]
	}
	whole, fraction, hasFraction := strings.Cut(text, ".")
	isDigits := func(s string) bool {
		return strings.Trim(s, "0123456789") == ""
	}
	if !isDigits(whole) || !isDigits(fraction) || (hasFraction && fraction == "") {
		return 0, false
	}
	lastPart := func(digits string) float64 {
		if hasFraction {
			digits += "." + fraction
		}
		v, _ := strconv.ParseFloat(digits, 64)
		return v
	}
	whole3 := func(digits string) float64 {
		v, _ := strconv.ParseFloat(digits, 64)
		return v
	}
	switch len(whole) {
	case 3:
		return sign * lastPart(whole), true
	case 5:
		return sign * (whole3(whole[:3]) + lastPart(whole[3:])/60), true
	case 7:
		return sign * (whole3(whole[:3]) + whole3(whole[3:5])/60 + lastPart(whole[5:])/3600), true
	default:
		return 0, false
	}
}

// stGeomFromMARC21 implements ST_GeomFromMARC21, which reads the coordinates of each 034 datafield of a MARC21/XML
// record as a point, or as a polygon when they describe a box. Several datafields make a multipoint, a multipolygon,
// or a collection, and a record without any is null.
func stGeomFromMARC21(ctx *sql.Context, args ...any) (any, error) {
	root, ok := parseXMLTree(args[0].(string), marc21Namespace)
	if !ok {
		return nil, pgerror.New(pgcode.Internal, "invalid MARC21/XML document.")
	}
	if root.name != "record" {
		return nil, pgerror.Newf(pgcode.Internal, "invalid MARC21/XML document. Root element <record> expected but <%s> found.", root.name)
	}
	var parts []shape
	for _, datafield := range root.childrenNamed("datafield") {
		if datafield.attrs["tag"] != "034" {
			continue
		}
		values := make(map[string]float64)
		for _, subfield := range datafield.childrenNamed("subfield") {
			code := subfield.attrs["code"]
			if len(code) != 1 || !strings.Contains("defg", code) {
				continue
			}
			value, ok := parseMARC21Coordinate(subfield.text.String())
			if !ok {
				return nil, pgerror.Newf(pgcode.Internal, "parse error - invalid literal at 034$%s: \"%s\"", code, subfield.text.String())
			}
			values[code] = value
		}
		if len(values) < 4 {
			return nil, pgerror.New(pgcode.Internal, "parse error - the Coded Cartographic Mathematical Data (datafield:034) in the given MARC21/XML is incomplete. Coordinates for subfields \"$d\",\"$e\",\"$f\" and \"$g\" are expected.")
		}
		west, east, north, south := values["d"], values["e"], values["f"], values["g"]
		if west == east && north == south {
			parts = append(parts, shape{typeNum: pointType, coords: []float64{west, north}})
		} else {
			parts = append(parts, shape{typeNum: polygonType, rings: [][]float64{{west, north, east, north, east, south, west, south, west, north}}})
		}
	}
	if len(parts) == 0 {
		return nil, nil
	}
	if len(parts) == 1 {
		return Geometry{Shape: parts[0]}, nil
	}
	typeNum := map[int]int{pointType: multiPointType, polygonType: multiPolygonType}[parts[0].typeNum]
	for _, part := range parts[1:] {
		if part.typeNum != parts[0].typeNum {
			typeNum = collectionType
		}
	}
	return Geometry{Shape: shape{typeNum: typeNum, members: parts}}, nil
}

// marc21Routines returns the routines that read and write MARC21/XML records.
func marc21Routines() []extdef.Routine {
	return []extdef.Routine{
		{Name: "st_asmarc21", Symbol: "ST_AsMARC21", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "format", Type: "text", Default: `'hdddmmss'::text`}}, Returns: "text", Strict: true, QueryImpl: stAsMARC21},
		{Name: "st_geomfrommarc21", Symbol: "ST_GeomFromMARC21", Parameters: []extdef.Parameter{{Name: "marc21xml", Type: "text"}}, Returns: "geometry", Strict: true, Impl: stGeomFromMARC21},
	}
}
