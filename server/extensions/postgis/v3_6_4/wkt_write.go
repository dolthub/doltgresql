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
	"math"
	"strconv"
	"strings"
)

// defaultPrecision is the number of decimal digits that text output uses unless told otherwise.
const defaultPrecision = 15

// wktVariant selects how WKT output marks dimensions and the SRID.
type wktVariant uint8

const (
	// wktISO marks every tagged geometry's dimensions with a separate "Z", "M", or "ZM", and omits the SRID.
	wktISO wktVariant = iota
	// wktExtended marks only a measured geometry with an "M" joined to its tag, and prefixes a known SRID.
	wktExtended
)

// untaggedMembers holds, for each type whose members may be written without a tag, the member type that is written
// that way.
var untaggedMembers = map[int]int{
	multiPointType:        pointType,
	multiLineType:         lineType,
	multiPolygonType:      polygonType,
	compoundType:          lineType,
	curvePolyType:         lineType,
	multiCurveType:        lineType,
	multiSurfaceType:      polygonType,
	polyhedralSurfaceType: polygonType,
	tinType:               triangleType,
}

// wktWriter writes WKT with a given precision and variant.
type wktWriter struct {
	sb        strings.Builder
	precision int
	variant   wktVariant
}

// writeWKT returns the WKT of the geometry, printing ordinates with at most `precision` decimal digits.
func writeWKT(g Geometry, precision int, variant wktVariant) string {
	w := &wktWriter{precision: precision, variant: variant}
	if variant == wktExtended && g.SRID != sridUnknown {
		w.sb.WriteString("SRID=")
		w.sb.WriteString(strconv.Itoa(int(g.SRID)))
		w.sb.WriteByte(';')
	}
	w.writeTagged(g.Shape)
	return w.sb.String()
}

// writeTagged writes the shape's tag, its dimensions, and its body.
func (w *wktWriter) writeTagged(s shape) {
	w.sb.WriteString(typeTags[s.typeNum])
	if w.variant == wktExtended {
		if s.ctype == dimXYM {
			w.sb.WriteByte('M')
		}
	} else if s.ctype != dimXY {
		w.sb.WriteByte(' ')
		w.sb.WriteString(map[coordinatesType]string{dimXYZ: "Z", dimXYM: "M", dimXYZM: "ZM"}[s.ctype])
		w.sb.WriteByte(' ')
	}
	if s.isEmpty() && len(s.members) == 0 {
		if w.variant == wktExtended || s.ctype == dimXY {
			w.sb.WriteByte(' ')
		}
		w.sb.WriteString("EMPTY")
		return
	}
	w.writeBody(s)
}

// writeUntagged writes a member that its collection writes without a tag.
func (w *wktWriter) writeUntagged(s shape) {
	if s.isEmpty() && len(s.members) == 0 {
		w.sb.WriteString("EMPTY")
		return
	}
	w.writeBody(s)
}

// writeBody writes the parenthesized body of a non-empty shape.
func (w *wktWriter) writeBody(s shape) {
	w.sb.WriteByte('(')
	switch s.typeNum {
	case pointType, lineType, circStringType:
		w.writeSequence(s.coords, s.ctype)
	case triangleType:
		w.sb.WriteByte('(')
		w.writeSequence(s.coords, s.ctype)
		w.sb.WriteByte(')')
	case polygonType:
		for i, ring := range s.rings {
			if i > 0 {
				w.sb.WriteByte(',')
			}
			w.sb.WriteByte('(')
			w.writeSequence(ring, s.ctype)
			w.sb.WriteByte(')')
		}
	default:
		untagged, hasUntagged := untaggedMembers[s.typeNum]
		for i, member := range s.members {
			if i > 0 {
				w.sb.WriteByte(',')
			}
			if s.typeNum == multiPointType && w.variant == wktExtended && !member.isEmpty() {
				w.writeSequence(member.coords, member.ctype)
			} else if hasUntagged && member.typeNum == untagged {
				w.writeUntagged(member)
			} else {
				w.writeTagged(member)
			}
		}
	}
	w.sb.WriteByte(')')
}

// writeSequence writes the comma-separated points of an ordinate sequence.
func (w *wktWriter) writeSequence(coords []float64, ctype coordinatesType) {
	dims := ctype.Dimension()
	for i := 0; i+dims <= len(coords); i += dims {
		if i > 0 {
			w.sb.WriteByte(',')
		}
		for j := 0; j < dims; j++ {
			if j > 0 {
				w.sb.WriteByte(' ')
			}
			w.sb.WriteString(formatOrdinate(coords[i+j], w.precision))
		}
	}
}

// formatOrdinate formats a number the way PostGIS prints ordinates. The shortest decimal digits that identify the
// number are rounded (half to even) to at most `precision` digits after the decimal point. Numbers whose magnitude is
// at most 1e-8 or at least 1e15 are written in exponent form, where the digits after the mantissa's decimal point are
// limited instead, and the exponent has no leading zeros.
func formatOrdinate(val float64, precision int) string {
	if math.IsNaN(val) {
		return "NaN"
	}
	if math.IsInf(val, 1) {
		return "Infinity"
	}
	if math.IsInf(val, -1) {
		return "-Infinity"
	}
	if val == 0 {
		return "0"
	}
	precision = max(precision, 0)
	digits, exp := shortestDigits(val)
	sign := ""
	if val < 0 {
		sign = "-"
	}
	if abs := math.Abs(val); abs <= 1e-8 || abs >= 1e15 {
		scientific := exp + len(digits) - 1
		digits, exp = roundDigits(digits, 1-len(digits), precision)
		return sign + fixedString(digits, exp) + "e" + exponentString(scientific)
	}
	digits, exp = roundDigits(digits, exp, precision)
	if digits == "0" {
		return "0"
	}
	return sign + fixedString(digits, exp)
}

// shortestDigits returns the shortest decimal digits that identify the magnitude of `val`, and the power of ten that
// those digits are scaled by.
func shortestDigits(val float64) (string, int) {
	s := strconv.FormatFloat(math.Abs(val), 'e', -1, 64)
	mantissa, expPart, _ := strings.Cut(s, "e")
	exp, _ := strconv.Atoi(expPart)
	digits := strings.Replace(mantissa, ".", "", 1)
	return digits, exp - (len(digits) - 1)
}

// roundDigits rounds digits scaled by 10^exp to at most `precision` digits after the decimal point, rounding half to
// even, and then removes trailing zeros.
func roundDigits(digits string, exp int, precision int) (string, int) {
	if trim := -exp - precision; trim > 0 {
		if trim > len(digits) {
			return "0", 0
		}
		kept, rest := []byte(digits[:len(digits)-trim]), digits[len(digits)-trim:]
		half := "5" + strings.Repeat("0", len(rest)-1)
		lastOdd := len(kept) > 0 && (kept[len(kept)-1]-'0')%2 == 1
		if rest > half || (rest == half && lastOdd) {
			i := len(kept) - 1
			for ; i >= 0 && kept[i] == '9'; i-- {
				kept[i] = '0'
			}
			if i < 0 {
				kept = append([]byte{'1'}, kept...)
			} else {
				kept[i]++
			}
		}
		digits, exp = string(kept), exp+trim
		if digits == "" {
			return "0", 0
		}
	}
	for len(digits) > 1 && digits[len(digits)-1] == '0' {
		digits = digits[:len(digits)-1]
		exp++
	}
	return digits, exp
}

// fixedString writes digits scaled by 10^exp in fixed-point notation.
func fixedString(digits string, exp int) string {
	if exp >= 0 {
		return digits + strings.Repeat("0", exp)
	}
	point := len(digits) + exp
	if point > 0 {
		return digits[:point] + "." + digits[point:]
	}
	return "0." + strings.Repeat("0", -point) + digits
}

// exponentString writes the signed exponent without leading zeros.
func exponentString(exp int) string {
	if exp < 0 {
		return "-" + strconv.Itoa(-exp)
	}
	return "+" + strconv.Itoa(exp)
}
