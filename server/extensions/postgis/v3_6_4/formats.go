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
)

// kmlWriter writes geometries as KML, prefixing every tag with the namespace prefix.
type kmlWriter struct {
	sb        strings.Builder
	precision int
	prefix    string
}

// writeKML returns the KML of a nonempty geometry, which must be in longitude and latitude.
func writeKML(g Geometry, precision int, prefix string) (any, error) {
	if g.SRID == sridUnknown {
		return nil, pgerror.New(pgcode.Internal, "ST_AsKML: Input geometry has unknown (0) SRID")
	}
	if g.Shape.isEmpty() {
		return nil, nil
	}
	if g.SRID != sridDefault {
		//TODO: transform the geometry to longitude and latitude once ST_Transform is supported
		return nil, pgerror.New(pgcode.FeatureNotSupported, "ST_AsKML is not yet supported for SRIDs other than 4326")
	}
	if prefix != "" {
		prefix += ":"
	}
	w := &kmlWriter{precision: precision, prefix: prefix}
	if err := w.write(g.Shape); err != nil {
		return nil, err
	}
	return w.sb.String(), nil
}

// tag writes an opening or closing tag with the namespace prefix.
func (w *kmlWriter) tag(name string, closing bool) {
	if closing {
		w.sb.WriteString("</" + w.prefix + name + ">")
	} else {
		w.sb.WriteString("<" + w.prefix + name + ">")
	}
}

// write writes a shape as a KML element. Multi-geometries and TINs become MultiGeometry elements.
func (w *kmlWriter) write(s shape) error {
	switch s.typeNum {
	case pointType, lineType:
		name := "Point"
		if s.typeNum == lineType {
			name = "LineString"
		}
		w.tag(name, false)
		w.writeCoordinates(s.coords, s.ctype)
		w.tag(name, true)
	case polygonType, triangleType:
		rings := s.rings
		if s.typeNum == triangleType {
			rings = [][]float64{s.coords}
		}
		w.tag("Polygon", false)
		for i, ring := range rings {
			boundary := "innerBoundaryIs"
			if i == 0 {
				boundary = "outerBoundaryIs"
			}
			w.tag(boundary, false)
			w.tag("LinearRing", false)
			w.writeCoordinates(ring, s.ctype)
			w.tag("LinearRing", true)
			w.tag(boundary, true)
		}
		w.tag("Polygon", true)
	case multiPointType, multiLineType, multiPolygonType, tinType:
		w.tag("MultiGeometry", false)
		for _, member := range s.members {
			if err := w.write(member); err != nil {
				return err
			}
		}
		w.tag("MultiGeometry", true)
	default:
		return pgerror.Newf(pgcode.Internal, "lwgeom_to_kml2: '%s' geometry type not supported", typeNameForNumber(s.typeNum))
	}
	return nil
}

// writeCoordinates writes a coordinates element, with X, Y, and Z when present separated by commas, and points
// separated by spaces.
func (w *kmlWriter) writeCoordinates(seq []float64, ctype coordinatesType) {
	w.tag("coordinates", false)
	dims := ctype.Dimension()
	for i := 0; i+dims <= len(seq); i += dims {
		if i > 0 {
			w.sb.WriteByte(' ')
		}
		w.sb.WriteString(formatOrdinate(seq[i], w.precision) + "," + formatOrdinate(seq[i+1], w.precision))
		if ctype.Is3D() {
			w.sb.WriteString("," + formatOrdinate(seq[i+2], w.precision))
		}
	}
	w.tag("coordinates", true)
}

// stAsKML implements ST_AsKML for geometry and geography.
func stAsKML(ctx *sql.Context, args ...any) (any, error) {
	return writeKML(args[0].(Geometry), int(args[1].(int32)), args[2].(string))
}

// svgWriter writes geometries as SVG path data, with Y coordinates negated. Relative output gives each point as the
// difference from the point before it, after both are rounded to the precision.
type svgWriter struct {
	sb        strings.Builder
	precision int
	relative  bool
}

// number formats a number at the writer's precision.
func (w *svgWriter) number(val float64) string {
	return formatOrdinate(val, w.precision)
}

// rounded returns a number rounded to the writer's precision.
func (w *svgWriter) rounded(val float64) float64 {
	rounded, _ := strconv.ParseFloat(w.number(val), 64)
	return rounded
}

// write writes a shape as SVG. Points are written as attributes, everything else as path data, and empty shapes as
// nothing.
func (w *svgWriter) write(s shape, isTop bool) error {
	if s.isEmpty() && s.typeNum != polyhedralSurfaceType && s.typeNum != triangleType && s.typeNum != tinType {
		return nil
	}
	dims := s.ctype.Dimension()
	switch s.typeNum {
	case pointType:
		if w.relative {
			w.sb.WriteString(`x="` + w.number(s.coords[0]) + `" y="` + w.number(-s.coords[1]) + `"`)
		} else {
			w.sb.WriteString(`cx="` + w.number(s.coords[0]) + `" cy="` + w.number(-s.coords[1]) + `"`)
		}
	case lineType:
		w.writePath(s.coords, dims, false)
	case polygonType:
		for i, ring := range s.rings {
			if i > 0 {
				w.sb.WriteByte(' ')
			}
			w.writePath(ring, dims, true)
		}
	case circStringType:
		w.writeCircularString(s.coords, dims)
	case compoundType:
		w.writeCompound(s.members, dims)
	case curvePolyType:
		for i, ring := range s.members {
			if i > 0 {
				w.sb.WriteByte(' ')
			}
			if ring.typeNum == lineType {
				w.writePath(ring.coords, dims, true)
				continue
			}
			if ring.typeNum == compoundType {
				w.writeCompound(ring.members, dims)
			} else {
				w.writeCircularString(ring.coords, dims)
			}
			w.close()
		}
	case multiPointType, multiLineType, multiPolygonType, multiCurveType, multiSurfaceType, collectionType:
		if s.typeNum == collectionType && !isTop {
			return pgerror.New(pgcode.Internal, "assvg_geom_buf: 'GeometryCollection' geometry type not supported.")
		}
		separator := " "
		if s.typeNum == multiPointType {
			separator = ","
		} else if s.typeNum == collectionType {
			separator = ";"
		}
		for i, member := range s.members {
			if i > 0 {
				w.sb.WriteString(separator)
			}
			if err := w.write(member, false); err != nil {
				return err
			}
		}
	default:
		return pgerror.Newf(pgcode.Internal, "lwgeom_to_svg: '%s' geometry type not supported", typeNameForNumber(s.typeNum))
	}
	return nil
}

// close writes the command that closes a ring.
func (w *svgWriter) close() {
	if w.relative {
		w.sb.WriteString(" z")
	} else {
		w.sb.WriteString(" Z")
	}
}

// writePath writes a sequence as a move to its first point followed by lines to the others. A ring leaves out its
// closing point and closes the path instead.
func (w *svgWriter) writePath(seq []float64, dims int, isRing bool) {
	end := len(seq)
	if isRing {
		end -= dims
	}
	w.sb.WriteString("M " + w.number(seq[0]) + " " + w.number(-seq[1]))
	if end > dims {
		if w.relative {
			w.sb.WriteString(" l")
		} else {
			w.sb.WriteString(" L")
		}
	}
	for i := dims; i < end; i += dims {
		if w.relative {
			w.sb.WriteString(" " + w.number(w.rounded(seq[i])-w.rounded(seq[i-dims])) + " " + w.number(w.rounded(-seq[i+1])-w.rounded(-seq[i-dims+1])))
		} else {
			w.sb.WriteString(" " + w.number(seq[i]) + " " + w.number(-seq[i+1]))
		}
	}
	if isRing {
		w.close()
	}
}

// writeCompound writes the members of a compound curve, where each member after the first continues the path from its
// own first point.
func (w *svgWriter) writeCompound(members []shape, dims int) {
	for i, member := range members {
		if i == 0 && member.typeNum == lineType {
			w.writePath(member.coords, dims, false)
			continue
		}
		if i == 0 {
			w.sb.WriteString("M ")
		} else {
			w.sb.WriteByte(' ')
		}
		w.sb.WriteString(w.number(member.coords[0]) + " " + w.number(-member.coords[1]))
		if member.typeNum == lineType {
			for j := dims; j < len(member.coords); j += dims {
				w.sb.WriteString(" " + w.number(member.coords[j]) + " " + w.number(-member.coords[j+1]))
			}
		} else {
			w.writeArcs(member.coords, dims)
		}
	}
}

// writeCircularString writes a circular string as a move to its first point followed by its arcs, or as a full circle
// alone when it is one.
func (w *svgWriter) writeCircularString(seq []float64, dims int) {
	if len(seq) != 3*dims || seq[0] != seq[2*dims] || seq[1] != seq[2*dims+1] {
		w.sb.WriteString("M " + w.number(seq[0]) + " " + w.number(-seq[1]))
	}
	w.writeArcs(seq, dims)
}

// writeArcs writes the arcs of a circular string as elliptical arc commands with equal radii. A full circle is
// written as a move to its center and two half circles.
func (w *svgWriter) writeArcs(seq []float64, dims int) {
	command := " A "
	if w.relative {
		command = " a "
	}
	for i := 2 * dims; i+dims <= len(seq); i += 2 * dims {
		start, mid, end := seq[i-2*dims:], seq[i-dims:], seq[i:]
		centerX, centerY, radius := circumcircle(start, mid, end)
		r := w.number(radius)
		if start[0] == end[0] && start[1] == end[1] {
			diameter := w.number(2 * radius)
			w.sb.WriteString("M " + w.number(centerX) + " " + w.number(-centerY) + " m " + r + " 0" +
				" a " + r + " " + r + " 0 1 0 -" + diameter + " 0" + " a " + r + " " + r + " 0 1 0 " + diameter + " 0")
			continue
		}
		sweep := "0"
		side := pointSide(start, end, mid)
		if side < 0 {
			sweep = "1"
		}
		large := "0"
		if pointSide(start, end, []float64{centerX, centerY}) == side {
			large = "1"
		}
		w.sb.WriteString(command + r + " " + r + " 0 " + large + " " + sweep + " " + w.number(end[0]) + " " + w.number(-end[1]))
	}
}

// stAsSVG implements ST_AsSVG for geometry and geography, which returns an empty string for an empty geometry.
func stAsSVG(ctx *sql.Context, args ...any) (any, error) {
	w := &svgWriter{relative: args[1].(int32) != 0, precision: int(args[2].(int32))}
	if err := w.write(args[0].(Geometry).Shape, true); err != nil {
		return nil, err
	}
	return w.sb.String(), nil
}

// geohashAlphabet holds the characters that geohashes use, in the order of their values.
const geohashAlphabet = "0123456789bcdefghjkmnpqrstuvwxyz"

// encodeGeohash returns the geohash of a longitude and latitude with the given number of characters. Each bit picks
// the upper half of the remaining range when the coordinate is at least its middle, alternating between longitude and
// latitude, starting with longitude.
func encodeGeohash(lon float64, lat float64, chars int) string {
	lonRange, latRange := [2]float64{-180, 180}, [2]float64{-90, 90}
	var sb strings.Builder
	for i := 0; i < chars; i++ {
		value := 0
		for bit := 0; bit < 5; bit++ {
			r, coord := &lonRange, lon
			if (i*5+bit)%2 == 1 {
				r, coord = &latRange, lat
			}
			mid := (r[0] + r[1]) / 2
			value <<= 1
			if coord >= mid {
				value |= 1
				r[0] = mid
			} else {
				r[1] = mid
			}
		}
		sb.WriteByte(geohashAlphabet[value])
	}
	return sb.String()
}

// geohashPrecision returns the number of geohash characters whose cell holds the whole box, up to twenty.
func geohashPrecision(b bbox) int {
	lonRange, latRange := [2]float64{-180, 180}, [2]float64{-90, 90}
	bits := 0
	for bits < 100 {
		r, low, high := &lonRange, b.minX, b.maxX
		if bits%2 == 1 {
			r, low, high = &latRange, b.minY, b.maxY
		}
		mid := (r[0] + r[1]) / 2
		if low > mid {
			r[0] = mid
		} else if high < mid {
			r[1] = mid
		} else {
			break
		}
		bits++
	}
	return bits / 5
}

// stGeoHash implements ST_GeoHash, which encodes the center of the geometry's box. Without a length, a point uses
// twenty characters and any other geometry uses the most characters whose cell holds its box.
func stGeoHash(ctx *sql.Context, args ...any) (any, error) {
	g, chars := args[0].(Geometry), int(args[1].(int32))
	b, ok := g.Shape.bounds()
	if !ok {
		return nil, nil
	}
	if b.minX < -180 || b.maxX > 180 || b.minY < -90 || b.maxY > 90 {
		return nil, pgerror.Newf(pgcode.Internal, "Geohash requires inputs in decimal degrees, got (%s %s, %s %s).",
			formatG(b.minX, 6), formatG(b.minY, 6), formatG(b.maxX, 6), formatG(b.maxY, 6))
	}
	if chars <= 0 {
		chars = 20
		if b.minX != b.maxX || b.minY != b.maxY {
			chars = geohashPrecision(b)
		}
	}
	return encodeGeohash((b.minX+b.maxX)/2, (b.minY+b.maxY)/2, chars), nil
}

// decodeGeohash returns the box of a geohash's cell, using at most the given number of characters unless it is
// negative.
func decodeGeohash(hash string, chars int) (box2d, error) {
	if chars >= 0 && chars < len(hash) {
		hash = hash[:chars]
	}
	lonRange, latRange := [2]float64{-180, 180}, [2]float64{-90, 90}
	bit := 0
	for _, c := range strings.ToLower(hash) {
		value := strings.IndexRune(geohashAlphabet, c)
		if value < 0 {
			return box2d{}, pgerror.Newf(pgcode.Internal, "decode_geohash_bbox: Invalid character '%c'", c)
		}
		for shift := 4; shift >= 0; shift-- {
			r := &lonRange
			if bit%2 == 1 {
				r = &latRange
			}
			mid := (r[0] + r[1]) / 2
			if value&(1<<shift) != 0 {
				r[0] = mid
			} else {
				r[1] = mid
			}
			bit++
		}
	}
	return box2d{minX: lonRange[0], minY: latRange[0], maxX: lonRange[1], maxY: latRange[1]}, nil
}

// geohashArgs decodes the geohash and optional length arguments, returning false for a null geohash. A null length
// uses every character.
func geohashArgs(args []any) (box2d, bool, error) {
	if args[0] == nil {
		return box2d{}, false, nil
	}
	chars := -1
	if len(args) > 1 && args[1] != nil {
		chars = int(args[1].(int32))
	}
	b, err := decodeGeohash(args[0].(string), chars)
	return b, err == nil, err
}

// stPointFromGeoHash implements ST_PointFromGeoHash, which returns the center of the geohash's cell.
func stPointFromGeoHash(ctx *sql.Context, args ...any) (any, error) {
	b, ok, err := geohashArgs(args)
	if !ok {
		return nil, err
	}
	point := shape{typeNum: pointType, coords: []float64{(b.minX + b.maxX) / 2, (b.minY + b.maxY) / 2}}
	return Geometry{Shape: point}, nil
}

// stGeomFromGeoHash implements ST_GeomFromGeoHash, which returns the polygon of the geohash's cell.
func stGeomFromGeoHash(ctx *sql.Context, args ...any) (any, error) {
	b, ok, err := geohashArgs(args)
	if !ok {
		return nil, err
	}
	ring := rectangle([]float64{b.minX, b.minY}, []float64{b.minX, b.maxY}, []float64{b.maxX, b.maxY}, []float64{b.maxX, b.minY})
	return Geometry{Shape: shape{typeNum: polygonType, rings: [][]float64{ring}}}, nil
}

// stBox2DFromGeoHash implements ST_Box2dFromGeoHash.
func stBox2DFromGeoHash(ctx *sql.Context, args ...any) (any, error) {
	b, ok, err := geohashArgs(args)
	if !ok {
		return nil, err
	}
	return b, nil
}

// stAsEncodedPolyline implements ST_AsEncodedPolyline, which writes a linestring or multipoint in longitude and
// latitude in the encoded polyline format: each latitude and longitude, scaled by ten to the precision and rounded,
// is written as its difference from the one before it in groups of five bits.
func stAsEncodedPolyline(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	if g.SRID != sridDefault {
		return nil, pgerror.New(pgcode.Internal, "Only SRID 4326 is supported.")
	}
	var seq []float64
	dims := g.Shape.ctype.Dimension()
	switch g.Shape.typeNum {
	case lineType:
		seq = g.Shape.coords
	case multiPointType:
		for _, member := range g.Shape.members {
			seq = append(seq, member.coords...)
		}
	default:
		return nil, pgerror.Newf(pgcode.Internal, "%s geometry type not supported by encoded polyline", typeNameForNumber(g.Shape.typeNum))
	}
	scale := math.Pow(10, float64(args[1].(int32)))
	var sb strings.Builder
	var prevLat, prevLon int64
	for i := 0; i+dims <= len(seq); i += dims {
		lat, lon := int64(math.Round(seq[i+1]*scale)), int64(math.Round(seq[i]*scale))
		writePolylineValue(&sb, lat-prevLat)
		writePolylineValue(&sb, lon-prevLon)
		prevLat, prevLon = lat, lon
	}
	return sb.String(), nil
}

// writePolylineValue writes a signed value in the encoded polyline format.
func writePolylineValue(sb *strings.Builder, value int64) {
	encoded := uint64(value << 1)
	if value < 0 {
		encoded = ^encoded
	}
	for encoded >= 0x20 {
		sb.WriteByte(byte((0x20 | (encoded & 0x1f)) + 63))
		encoded >>= 5
	}
	sb.WriteByte(byte(encoded + 63))
}

// stLineFromEncodedPolyline implements ST_LineFromEncodedPolyline, which reads a linestring in longitude and latitude
// from the encoded polyline format.
func stLineFromEncodedPolyline(ctx *sql.Context, args ...any) (any, error) {
	text := args[0].(string)
	scale := math.Pow(10, float64(args[1].(int32)))
	line := shape{typeNum: lineType}
	var values []int64
	var current uint64
	shift := 0
	for i := 0; i < len(text); i++ {
		chunk := uint64(text[i]) - 63
		current |= (chunk & 0x1f) << shift
		shift += 5
		if chunk < 0x20 {
			value := int64(current >> 1)
			if current&1 != 0 {
				value = ^value
			}
			values = append(values, value)
			current, shift = 0, 0
		}
	}
	var lat, lon int64
	for i := 0; i+1 < len(values); i += 2 {
		lat += values[i]
		lon += values[i+1]
		line.coords = append(line.coords, float64(lon)/scale, float64(lat)/scale)
	}
	return Geometry{SRID: sridDefault, Shape: line}, nil
}

// latLonToken is a piece of an ST_AsLatLonText format: literal text, or a unit with its width and decimal digits.
type latLonToken struct {
	unit     byte
	text     string
	width    int
	decimals int
}

// parseLatLonFormat splits an ST_AsLatLonText format into its pieces. Degrees (D), minutes (M), and seconds (S) are
// runs of their letter, optionally followed by a decimal point and more letters for the decimal digits, and C is the
// compass direction. Degrees are required, seconds require minutes, and no unit may appear twice.
func parseLatLonFormat(format string) ([]latLonToken, error) {
	if format == "" {
		format = `D°M'S.SSS"C`
	}
	names := map[byte]string{'D': "degrees (DD.DDD)", 'M': "minutes (MM.MMM)", 'S': "seconds (SS.SSS)", 'C': "compass dir (C)"}
	seen := make(map[byte]bool)
	var tokens []latLonToken
	for i := 0; i < len(format); {
		c := format[i]
		if names[c] == "" {
			if len(tokens) == 0 || tokens[len(tokens)-1].unit != 0 {
				tokens = append(tokens, latLonToken{})
			}
			tokens[len(tokens)-1].text += format[i : i+1]
			i++
			continue
		}
		if seen[c] {
			return nil, pgerror.Newf(pgcode.Internal, "Bad format, cannot include %s more than once.", names[c])
		}
		seen[c] = true
		token := latLonToken{unit: c}
		for i < len(format) && format[i] == c {
			token.width++
			i++
		}
		if c != 'C' && i+1 < len(format) && format[i] == '.' && format[i+1] == c {
			for i < len(format) && (format[i] == c || format[i] == '.') {
				if format[i] == c {
					token.decimals++
				}
				i++
			}
		}
		tokens = append(tokens, token)
	}
	if !seen['D'] {
		return nil, pgerror.New(pgcode.Internal, "Bad format, degrees (DD.DDD) must be included.")
	}
	if seen['S'] && !seen['M'] {
		return nil, pgerror.New(pgcode.Internal, "Bad format, cannot include seconds (SS.SSS) without including minutes (MM.MMM).")
	}
	return tokens, nil
}

// formatLatLon formats one coordinate with the format's pieces. Without a compass direction, a negative coordinate
// has a negative degree count. Seconds that round up to sixty carry into the minutes and degrees.
func formatLatLon(value float64, tokens []latLonToken, positive string, negative string) string {
	units := make(map[byte]latLonToken)
	for _, token := range tokens {
		units[token.unit] = token
	}
	_, hasM := units['M']
	_, hasS := units['S']
	_, hasC := units['C']
	abs := math.Abs(value)
	degrees, minutes, seconds := abs, 0.0, 0.0
	if hasM {
		degrees = math.Floor(abs)
		minutes = (abs - degrees) * 60
		if hasS {
			whole := math.Floor(minutes)
			seconds = (minutes - whole) * 60
			minutes = whole
			if rounded, _ := strconv.ParseFloat(strconv.FormatFloat(seconds, 'f', units['S'].decimals, 64), 64); rounded >= 60 {
				seconds, minutes = 0, minutes+1
				if minutes >= 60 {
					minutes, degrees = 0, degrees+1
				}
			}
		}
	}
	if value < 0 && !hasC {
		degrees = -degrees
	}
	var sb strings.Builder
	for _, token := range tokens {
		switch token.unit {
		case 'D':
			sb.WriteString(fmt.Sprintf("%*.*f", token.width, token.decimals, degrees))
		case 'M':
			sb.WriteString(fmt.Sprintf("%*.*f", token.width, token.decimals, minutes))
		case 'S':
			sb.WriteString(fmt.Sprintf("%*.*f", token.width, token.decimals, seconds))
		case 'C':
			if value < 0 {
				sb.WriteString(negative)
			} else {
				sb.WriteString(positive)
			}
		default:
			sb.WriteString(token.text)
		}
	}
	return sb.String()
}

// stAsLatLonText implements ST_AsLatLonText, which writes a point's latitude and then longitude with the format,
// after bringing them into range. A latitude beyond a pole is reflected back, which moves the longitude halfway
// around the world.
func stAsLatLonText(ctx *sql.Context, args ...any) (any, error) {
	g := args[0].(Geometry)
	if g.Shape.typeNum != pointType {
		return nil, pgerror.Newf(pgcode.Internal, "Only points are supported, you tried type %s.", typeNameForNumber(g.Shape.typeNum))
	}
	if g.Shape.isEmpty() {
		return nil, pgerror.New(pgcode.Internal, "Cannot convert an empty point into formatted text.")
	}
	tokens, err := parseLatLonFormat(args[1].(string))
	if err != nil {
		return nil, err
	}
	lon, lat := g.Shape.coords[0], g.Shape.coords[1]
	lat = math.Mod(lat, 360)
	if lat > 90 {
		lat, lon = 180-lat, lon+180
	} else if lat < -90 {
		lat, lon = -180-lat, lon+180
	}
	lon = math.Mod(lon, 360)
	if lon > 180 {
		lon -= 360
	} else if lon < -180 {
		lon += 360
	}
	return formatLatLon(lat, tokens, "N", "S") + " " + formatLatLon(lon, tokens, "E", "W"), nil
}

// isLatLong returns whether the SRID names a longitude and latitude coordinate system.
func isLatLong(srid int32) bool {
	return srid != sridUnknown && checkLatLongSRID(srid) == nil
}

// twkbShape returns a copy of the shape with every ordinate rounded to its number of decimal digits, half away from
// zero, and without the consecutive repeated points that rounding leaves in linestrings and rings. Types that TWKB
// lacks are rejected.
func twkbShape(s shape, precisions [4]int) (shape, error) {
	dims := s.ctype.Dimension()
	round := func(val float64, digits int) float64 {
		scale := math.Pow(10, float64(digits))
		return math.Round(val*scale) / scale
	}
	rounded := s.mapCoords(func(point []float64) []float64 {
		point[0], point[1] = round(point[0], precisions[0]), round(point[1], precisions[1])
		if s.hasZ() {
			point[2] = round(point[2], precisions[2])
		}
		if s.hasM() {
			point[dims-1] = round(point[dims-1], precisions[3])
		}
		return point
	})
	switch s.typeNum {
	case pointType, multiPointType:
		return rounded, nil
	case lineType:
		rounded.coords = removeRepeated(rounded.coords, dims, 0, 2)
		return rounded, nil
	case polygonType:
		for i, ring := range rounded.rings {
			rounded.rings[i] = removeRepeated(ring, dims, 0, 4)
		}
		return rounded, nil
	case multiLineType, multiPolygonType, collectionType:
		for i, member := range rounded.members {
			converted, err := twkbShape(member, precisions)
			if err != nil {
				return shape{}, err
			}
			rounded.members[i] = converted
		}
		return rounded, nil
	default:
		return shape{}, pgerror.Newf(pgcode.Internal, "lwgeom_twkb_type: Unsupported geometry type: %s", typeNameForNumber(s.typeNum))
	}
}

// writeTWKB returns the TWKB of a shape. The precision arguments start at the index: the X and Y precision, which
// defaults to five decimal digits for longitude and latitude and zero otherwise, the Z and M precisions, which default
// to zero, and whether to write sizes and bounding boxes. An empty shape is written without a bounding box.
func writeTWKB(s shape, srid int32, args []any, index int, ids []int64) ([]byte, error) {
	precisions := [4]int{}
	if isLatLong(srid) {
		precisions[0] = 5
	}
	if args[index] != nil {
		precisions[0] = int(args[index].(int32))
	}
	precisions[1] = precisions[0]
	for i := 1; i <= 2; i++ {
		if args[index+i] != nil {
			precisions[i+1] = int(args[index+i].(int32))
		}
	}
	rounded, err := twkbShape(s, precisions)
	if err != nil {
		return nil, err
	}
	withSizes := args[index+3] != nil && args[index+3].(bool)
	if rounded.isEmpty() && ids == nil {
		zigzag := byte((precisions[0] << 1) ^ (precisions[0] >> 31))
		twkb := []byte{byte(rounded.typeNum) | zigzag<<4, 0x10}
		if withSizes {
			twkb[1] |= 0x02
		}
		if rounded.hasZ() || rounded.hasM() {
			twkb[1] |= 0x08
			dims := byte(0)
			if rounded.hasZ() {
				dims |= 0x01 | byte(precisions[2])<<2
			}
			if rounded.hasM() {
				dims |= 0x02 | byte(precisions[3])<<5
			}
			twkb = append(twkb, dims)
		}
		if withSizes {
			twkb = append(twkb, 0)
		}
		return twkb, nil
	}
	writer := twkbWriter{precisions: precisions, withSizes: withSizes, withBoxes: args[index+4] != nil && args[index+4].(bool)}
	return writer.write(rounded, ids), nil
}

// stAsTWKB implements ST_AsTWKB for a geometry.
func stAsTWKB(ctx *sql.Context, args ...any) (any, error) {
	if args[0] == nil {
		return nil, nil
	}
	g := args[0].(Geometry)
	return writeTWKB(g.Shape, g.SRID, args, 1, nil)
}

// stAsTWKBArray implements ST_AsTWKB for an array of geometries with an array of identifiers, which it writes as a
// collection with an identifier list.
func stAsTWKBArray(ctx *sql.Context, args ...any) (any, error) {
	if args[0] == nil || args[1] == nil {
		return nil, nil
	}
	geometries, idValues := args[0].([]any), args[1].([]any)
	if len(geometries) != len(idValues) {
		return nil, pgerror.New(pgcode.Internal, "size of geometry[] and integer[] arrays must match")
	}
	collection := shape{typeNum: collectionType}
	ids := make([]int64, 0, len(idValues))
	srid := int32(sridUnknown)
	var typeNum int
	for i, elem := range geometries {
		if elem == nil || idValues[i] == nil {
			continue
		}
		g := elem.(Geometry)
		if len(collection.members) == 0 {
			srid, typeNum, collection.ctype = g.SRID, g.Shape.typeNum, g.Shape.ctype
		} else if g.Shape.typeNum != typeNum {
			typeNum = collectionType
		}
		collection.members = append(collection.members, g.Shape)
		ids = append(ids, idValues[i].(int64))
	}
	collection.typeNum = collectionTypeFor(typeNum)
	return writeTWKB(collection, srid, args, 2, ids)
}

// stGeomFromTWKB implements ST_GeomFromTWKB.
func stGeomFromTWKB(ctx *sql.Context, args ...any) (any, error) {
	reader := twkbReader{data: args[0].([]byte)}
	s, err := reader.read()
	if err != nil {
		return nil, err
	}
	return Geometry{Shape: s}, nil
}

// formatsRoutines returns the routines that read and write KML, SVG, geohashes, and encoded polylines.
func formatsRoutines() []extdef.Routine {
	precision15 := extdef.Parameter{Name: "maxdecimaldigits", Type: "int4", Default: `15`}
	return []extdef.Routine{
		{Name: "st_astwkb", Symbol: "TWKBFromLWGEOMArray", Parameters: []extdef.Parameter{{Name: "geom", Type: "_geometry"}, {Name: "ids", Type: "_int8"}, {Name: "prec", Type: "int4", Default: `NULL::integer`}, {Name: "prec_z", Type: "int4", Default: `NULL::integer`}, {Name: "prec_m", Type: "int4", Default: `NULL::integer`}, {Name: "with_sizes", Type: "bool", Default: `NULL::boolean`}, {Name: "with_boxes", Type: "bool", Default: `NULL::boolean`}}, Returns: "bytea", Impl: stAsTWKBArray},
		{Name: "st_astwkb", Symbol: "TWKBFromLWGEOM", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "prec", Type: "int4", Default: `NULL::integer`}, {Name: "prec_z", Type: "int4", Default: `NULL::integer`}, {Name: "prec_m", Type: "int4", Default: `NULL::integer`}, {Name: "with_sizes", Type: "bool", Default: `NULL::boolean`}, {Name: "with_boxes", Type: "bool", Default: `NULL::boolean`}}, Returns: "bytea", Impl: stAsTWKB},
		{Name: "st_geomfromtwkb", Symbol: "LWGEOMFromTWKB", Parameters: params("bytea"), Returns: "geometry", Strict: true, Impl: stGeomFromTWKB},
		{Name: "st_aslatlontext", Symbol: "LWGEOM_to_latlon", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "tmpl", Type: "text", Default: `''::text`}}, Returns: "text", Strict: true, Impl: stAsLatLonText},
		{Name: "st_asencodedpolyline", Symbol: "LWGEOM_asEncodedPolyline", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "nprecision", Type: "int4", Default: `5`}}, Returns: "text", Strict: true, Impl: stAsEncodedPolyline},
		{Name: "st_askml", Symbol: "geography_as_kml", Parameters: []extdef.Parameter{{Name: "geog", Type: "geography"}, precision15, {Name: "nprefix", Type: "text", Default: `''::text`}}, Returns: "text", Strict: true, Impl: stAsKML},
		{Name: "st_askml", Symbol: "LWGEOM_asKML", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, precision15, {Name: "nprefix", Type: "text", Default: `''::text`}}, Returns: "text", Strict: true, Impl: stAsKML},
		{Name: "st_askml", Symbol: "st_askml_text", Parameters: params("text"), Returns: "text", Strict: true, Impl: textArgsAsGeometry(1, func(ctx *sql.Context, args ...any) (any, error) {
			return writeKML(args[0].(Geometry), defaultPrecision, "")
		})},
		{Name: "st_assvg", Symbol: "geography_as_svg", Parameters: []extdef.Parameter{{Name: "geog", Type: "geography"}, {Name: "rel", Type: "int4", Default: `0`}, precision15}, Returns: "text", Strict: true, Impl: stAsSVG},
		{Name: "st_assvg", Symbol: "LWGEOM_asSVG", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "rel", Type: "int4", Default: `0`}, precision15}, Returns: "text", Strict: true, Impl: stAsSVG},
		{Name: "st_assvg", Symbol: "st_assvg_text", Parameters: params("text"), Returns: "text", Strict: true, Impl: textArgsAsGeometry(1, func(ctx *sql.Context, args ...any) (any, error) {
			return stAsSVG(ctx, args[0], int32(0), int32(defaultPrecision))
		})},
		{Name: "st_box2dfromgeohash", Symbol: "box2d_from_geohash", Parameters: []extdef.Parameter{{Type: "text"}, {Type: "int4", Default: `NULL::integer`}}, Returns: "box2d", Impl: stBox2DFromGeoHash},
		{Name: "st_geohash", Symbol: "st_geohash_geography_int4", Parameters: []extdef.Parameter{{Name: "geog", Type: "geography"}, {Name: "maxchars", Type: "int4", Default: `0`}}, Returns: "text", Strict: true, Impl: stGeoHash},
		{Name: "st_geohash", Symbol: "st_geohash_geometry_int4", Parameters: []extdef.Parameter{{Name: "geom", Type: "geometry"}, {Name: "maxchars", Type: "int4", Default: `0`}}, Returns: "text", Strict: true, Impl: stGeoHash},
		{Name: "st_geomfromgeohash", Symbol: "st_geomfromgeohash_text_int4", Parameters: []extdef.Parameter{{Type: "text"}, {Type: "int4", Default: `NULL::integer`}}, Returns: "geometry", Impl: stGeomFromGeoHash},
		{Name: "st_linefromencodedpolyline", Symbol: "line_from_encoded_polyline", Parameters: []extdef.Parameter{{Name: "txtin", Type: "text"}, {Name: "nprecision", Type: "int4", Default: `5`}}, Returns: "geometry", Strict: true, Impl: stLineFromEncodedPolyline},
		{Name: "st_pointfromgeohash", Symbol: "point_from_geohash", Parameters: []extdef.Parameter{{Type: "text"}, {Type: "int4", Default: `NULL::integer`}}, Returns: "geometry", Impl: stPointFromGeoHash},
	}
}
