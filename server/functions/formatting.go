// Copyright 2025 Dolthub, Inc.
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

package functions

import (
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/dolthub/doltgresql/postgres/parser/duration"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
)

// This formatting implementation comes from postgres/src/backend/utils/adt/formatting.c

// Time and date constants for formatting calculations
const (
	monthsPerYear = 12
	hoursPerDay   = 24
	minsPerHour   = 60
	usecsPerSecs  = 1000000
	maxTzdispHour = 15 // Maximum allowed hour part for timezone display
)

// formatNode represents a single element in a parsed format string
type formatNode struct {
	typ        formatNodeType // Type of the node (action, character, separator, space)
	characters []uint8        // Characters for literal text nodes
	suffix     *keywordSuffix // Prefix modifiers like FM, TM
	postfix    *keywordSuffix // Suffix modifiers like TH, SP
	key        *keyword       // Format keyword for action nodes
}

// formatNodeType defines the type of formatting node
type formatNodeType uint

const (
	nodeTypeACTION    formatNodeType = iota + 1 // Format action (e.g., YYYY, MM, DD)
	nodeTypeCHAR                                // Literal character
	nodeTypeSEPARATOR                           // Separator character
	nodeTypeSPACE                               // Whitespace
)

// fromCharDateMode defines the date parsing mode
type fromCharDateMode uint

const (
	fromCharDateNONE      fromCharDateMode = iota // Value does not affect date mode
	fromCharDateGREGORIAN                         // Gregorian (day, month, year) style date
	fromCharDateISOWEEK                           // ISO 8601 week date
)

// tmFromChar holds parsed date/time components from input string
// For char->date/time conversion
type tmFromChar struct {
	mode      fromCharDateMode // Date parsing mode (Gregorian, ISO week or neither)
	hh        int              // Hour (12 or 24 hour format)
	pm        int              // 0 if AM; 1 if PM
	mi        int              // Minutes
	ss        int              // Seconds
	ssss      int              // Seconds since midnight
	d         int              // Day of week (1-7, Sunday = 1, 0 means missing)
	dd        int              // Day of month
	ddd       int              // Day of year
	mm        int              // Month
	ms        int              // Milliseconds
	year      int              // Year
	bc        int              // BC/AD indicator
	ww        int              // Week of year
	w         int              // Week of month
	cc        int              // Century
	j         int              // Julian day
	us        int              // Microseconds
	yysz      int              // Year size (YY=2, YYYY=4)
	is12clock bool             // True for 12-hour clock format
	tzsign    int              // Timezone sign: +1, -1, or 0 if no TZH/TZM fields
	tzh       int              // Timezone hours
	tzm       int              // Timezone minutes
	ff        int              // Fractional precision
	// TODO: use these when we support TZ
	//has_tz    bool             // True if TZ field was present
	//gmtoffset int              // GMT offset of fixed-offset zone abbreviation
}

// keywordSuffix represents format modifiers like FM (fill mode) or TH (ordinal)
type keywordSuffix struct {
	name      string // Name of the suffix/prefix
	len       int    // Length of the modifier
	isPostFix bool   // True if this is a postfix modifier
}

var orderedKeywordSuffixes = []*keywordSuffix{
	{"FM", 2, false},
	{"fm", 2, false},
	{"TM", 2, false},
	{"tm", 2, false},
	{"TH", 2, true},
	{"th", 2, true},
	{"SP", 2, true},
}

// keyword represents a format specifier (like YYYY, MM, DD, etc.)
type keyword struct {
	name    string           // Name of the keyword (e.g., "YYYY", "MM")
	len     int              // Length of the keyword
	isDigit bool             // True if this keyword expects digits
	fcdMode fromCharDateMode // Date mode this keyword belongs to
}

var orderedDCHKeywords = []*keyword{
	{"A.D.", 4, false, fromCharDateNONE},
	{"A.M.", 4, false, fromCharDateNONE},
	{"AD", 2, false, fromCharDateNONE},
	{"AM", 2, false, fromCharDateNONE},
	{"B.C.", 4, false, fromCharDateNONE},
	{"BC", 2, false, fromCharDateNONE},
	{"CC", 2, true, fromCharDateNONE},
	{"DAY", 3, false, fromCharDateNONE},
	{"DDD", 3, true, fromCharDateGREGORIAN},
	{"DD", 2, true, fromCharDateGREGORIAN},
	{"DY", 2, false, fromCharDateNONE},
	{"Day", 3, false, fromCharDateNONE},
	{"Dy", 2, false, fromCharDateNONE},
	{"D", 1, true, fromCharDateGREGORIAN},
	{"FF1", 3, true, fromCharDateNONE},
	{"FF2", 3, true, fromCharDateNONE},
	{"FF3", 3, true, fromCharDateNONE},
	{"FF4", 3, true, fromCharDateNONE},
	{"FF5", 3, true, fromCharDateNONE},
	{"FF6", 3, true, fromCharDateNONE},
	{"FX", 2, false, fromCharDateNONE},
	{"HH24", 4, true, fromCharDateNONE},
	{"HH12", 4, true, fromCharDateNONE},
	{"HH", 2, true, fromCharDateNONE},
	{"IDDD", 4, true, fromCharDateISOWEEK},
	{"ID", 2, true, fromCharDateISOWEEK},
	{"IW", 2, true, fromCharDateISOWEEK},
	{"IYYY", 4, true, fromCharDateISOWEEK},
	{"IYY", 3, true, fromCharDateISOWEEK},
	{"IY", 2, true, fromCharDateISOWEEK},
	{"I", 1, true, fromCharDateISOWEEK},
	{"J", 1, true, fromCharDateNONE},
	{"MI", 2, true, fromCharDateNONE},
	{"MM", 2, true, fromCharDateGREGORIAN},
	{"MONTH", 5, false, fromCharDateGREGORIAN},
	{"MON", 3, false, fromCharDateGREGORIAN},
	{"MS", 2, true, fromCharDateNONE},
	{"Month", 5, false, fromCharDateGREGORIAN},
	{"Mon", 3, false, fromCharDateGREGORIAN},
	{"OF", 2, false, fromCharDateNONE},
	{"P.M.", 4, false, fromCharDateNONE},
	{"PM", 2, false, fromCharDateNONE},
	{"Q", 1, true, fromCharDateNONE},
	{"RM", 2, false, fromCharDateGREGORIAN},
	{"SSSSS", 5, true, fromCharDateNONE},
	{"SSSS", 4, true, fromCharDateNONE},
	{"SS", 2, true, fromCharDateNONE},
	{"TZH", 3, false, fromCharDateNONE},
	{"TZM", 3, true, fromCharDateNONE},
	{"TZ", 2, false, fromCharDateNONE},
	{"US", 2, true, fromCharDateNONE},
	{"WW", 2, true, fromCharDateGREGORIAN},
	{"W", 1, true, fromCharDateGREGORIAN},
	{"Y,YYY", 5, true, fromCharDateGREGORIAN},
	{"YYYY", 4, true, fromCharDateGREGORIAN},
	{"YYY", 3, true, fromCharDateGREGORIAN},
	{"YY", 2, true, fromCharDateGREGORIAN},
	{"Y", 1, true, fromCharDateGREGORIAN},
	{"a.d.", 4, false, fromCharDateNONE},
	{"a.m.", 4, false, fromCharDateNONE},
	{"ad", 2, false, fromCharDateNONE},
	{"am", 2, false, fromCharDateNONE},
	{"b.c.", 4, false, fromCharDateNONE},
	{"bc", 2, false, fromCharDateNONE},
	{"cc", 2, true, fromCharDateNONE},
	{"day", 3, false, fromCharDateNONE},
	{"ddd", 3, true, fromCharDateGREGORIAN},
	{"dd", 2, true, fromCharDateGREGORIAN},
	{"dy", 2, false, fromCharDateNONE},
	{"d", 1, true, fromCharDateGREGORIAN},
	{"ff1", 3, true, fromCharDateNONE},
	{"ff2", 3, true, fromCharDateNONE},
	{"ff3", 3, true, fromCharDateNONE},
	{"ff4", 3, true, fromCharDateNONE},
	{"ff5", 3, true, fromCharDateNONE},
	{"ff6", 3, true, fromCharDateNONE},
	{"fx", 2, false, fromCharDateNONE},
	{"hh24", 4, true, fromCharDateNONE},
	{"hh12", 4, true, fromCharDateNONE},
	{"hh", 2, true, fromCharDateNONE},
	{"iddd", 4, true, fromCharDateISOWEEK},
	{"id", 2, true, fromCharDateISOWEEK},
	{"iw", 2, true, fromCharDateISOWEEK},
	{"iyyy", 4, true, fromCharDateISOWEEK},
	{"iyy", 3, true, fromCharDateISOWEEK},
	{"iy", 2, true, fromCharDateISOWEEK},
	{"i", 1, true, fromCharDateISOWEEK},
	{"j", 1, true, fromCharDateNONE},
	{"mi", 2, true, fromCharDateNONE},
	{"mm", 2, true, fromCharDateGREGORIAN},
	{"month", 5, false, fromCharDateGREGORIAN},
	{"mon", 3, false, fromCharDateGREGORIAN},
	{"ms", 2, true, fromCharDateNONE},
	{"of", 2, false, fromCharDateNONE},
	{"p.m.", 4, false, fromCharDateNONE},
	{"pm", 2, false, fromCharDateNONE},
	{"q", 1, true, fromCharDateNONE},
	{"rm", 2, false, fromCharDateGREGORIAN},
	{"sssss", 5, true, fromCharDateNONE},
	{"ssss", 4, true, fromCharDateNONE},
	{"ss", 2, true, fromCharDateNONE},
	{"tzh", 3, false, fromCharDateNONE},
	{"tzm", 3, true, fromCharDateNONE},
	{"tz", 2, false, fromCharDateNONE},
	{"us", 2, true, fromCharDateNONE},
	{"ww", 2, true, fromCharDateGREGORIAN},
	{"w", 1, true, fromCharDateGREGORIAN},
	{"y,yyy", 5, true, fromCharDateGREGORIAN},
	{"yyyy", 4, true, fromCharDateGREGORIAN},
	{"yyy", 3, true, fromCharDateGREGORIAN},
	{"yy", 2, true, fromCharDateGREGORIAN},
	{"y", 1, true, fromCharDateGREGORIAN},
}

// isSpace returns true if the character is a space.
func isSpace(s uint8) bool {
	return s == ' ' || s == '\t'
}

// isSeperateChar returns true if the character is a printable ASCII separator (not letter/digit)
func isSeperateChar(s uint8) bool {
	return s > 0x20 && s < 0x7F && !unicode.IsUpper(rune(s)) && !unicode.IsLower(rune(s)) && !unicode.IsDigit(rune(s))
}

// suffSearch searches for format modifiers (prefixes/suffixes) at the beginning of the string
func suffSearch(str string) *keywordSuffix {
	l := len(str)
	for _, s := range orderedKeywordSuffixes {
		if l >= s.len && str[:s.len] == s.name {
			return s
		}
	}
	return nil
}

// keywordSearch searches for format keywords at the beginning of the string
func keywordSearch(str string) *keyword {
	l := len(str)
	for _, k := range orderedDCHKeywords {
		if l >= k.len && str[:k.len] == k.name {
			return k
		}
	}
	return nil
}

// parseFormat breaks down a format string into individual FormatNodes for processing
func parseFormat(format string) ([]*formatNode, error) {
	var f []*formatNode
	formatPos := 0
	for formatPos < len(format) {
		var suffix *keywordSuffix
		// prefix
		if suffix = suffSearch(format[formatPos:]); suffix != nil {
			formatPos += suffix.len
		}
		// keyword
		if k := keywordSearch(format[formatPos:]); k != nil {
			newNode := &formatNode{
				typ:    nodeTypeACTION,
				suffix: suffix,
				key:    k,
			}
			formatPos += k.len
			// postfix
			if postfix := suffSearch(format[formatPos:]); postfix != nil {
				newNode.postfix = postfix
				formatPos += postfix.len
			}
			f = append(f, newNode)
		} else {
			if format[formatPos] == '"' {
				// Process double-quoted literal string, if any
				formatPos++
				for formatPos < len(format) {
					if format[formatPos] == '"' {
						formatPos++
						break
					}
					// backslash quotes the next character, if any
					if format[formatPos] == '\\' && formatPos+1 < len(format) {
						formatPos++
					}
					newNode := &formatNode{
						typ:        nodeTypeCHAR,
						characters: []uint8{format[formatPos]},
					}
					f = append(f, newNode)
					formatPos++
				}
			} else {
				// Outside double-quoted strings, backslash is only special if
				// it immediately precedes a double quote.
				if format[formatPos] == '\\' && formatPos+1 < len(format) && format[formatPos+1] == '"' {
					formatPos++
				}
				newNode := &formatNode{}
				if isSeperateChar(format[formatPos]) {
					newNode.typ = nodeTypeSEPARATOR
				} else if isSpace(format[formatPos]) {
					newNode.typ = nodeTypeSPACE
				} else {
					newNode.typ = nodeTypeCHAR
				}
				newNode.characters = []uint8{format[formatPos]}
				f = append(f, newNode)
				formatPos++
			}
		}
	}
	return f, nil
}

// getDateTimeFromFormat parses an input string according to a format specification and returns a time.Time.
func getDateTimeFromFormat(ctx *sql.Context, input, format string) (time.Time, error) {
	formatNodes, err := parseFormat(format)
	if err != nil {
		return time.Time{}, err
	}

	inputPos := 0
	extraSkip := 0
	fxMode := false
	var tfc tmFromChar

	for i, n := range formatNodes {
		if inputPos >= len(input) {
			break
		}

		if !fxMode && (n.typ != nodeTypeACTION || (n.key.name != "fx" && n.key.name != "FX")) && (n.typ == nodeTypeACTION || n == formatNodes[0]) {
			for isSpace(input[inputPos]) {
				inputPos++
				extraSkip++
			}
		}

		if n.typ == nodeTypeSPACE || n.typ == nodeTypeSEPARATOR {
			if !fxMode {
				// In non FX (fixed format) mode one format string space or
				// separator match to one space or separator in input string.
				// Or match nothing if there is no space or separator in the
				// current position of input string.
				extraSkip--
				if isSpace(input[inputPos]) || isSeperateChar(input[inputPos]) {
					inputPos++
					extraSkip++
				}
			} else {
				// In FX mode, on format string space or separator we consume
				// exactly one character from input string.  Notice we don't
				// insist that the consumed character match the format's
				// character.
				inputPos++
			}
			continue
		} else if n.typ != nodeTypeACTION {
			if !fxMode {
				// In non FX mode we might have skipped some extra characters
				// (more than specified in format string) before.  In this
				// case we don't skip input string character, because it might
				// be part of field.
				if extraSkip > 0 {
					extraSkip--
				} else {
					inputPos++
				}
			} else {
				inputPos += len(n.characters)
			}
			continue
		}

		if n.key.fcdMode != fromCharDateNONE {
			if tfc.mode == fromCharDateNONE {
				tfc.mode = n.key.fcdMode
			} else if tfc.mode != n.key.fcdMode {
				return time.Time{}, errors.Errorf(`invalid combination of date conventions`)
			}
		}

		switch n.key.name {
		case "fx", "FX":
			fxMode = true
		case "a.m.", "A.M.", "p.m.", "P.M.":
			v, l, err := fromCharSeqSearch(n.key.name, input[inputPos:],
				[]string{"a.m.", "p.m.", "A.M.", "P.M."}, tfc.pm,
				func(i int) int { return i % 2 })
			if err != nil {
				return time.Time{}, err
			}
			tfc.pm = v
			inputPos += l
			tfc.is12clock = true
		case "am", "AM", "pm", "PM":
			v, l, err := fromCharSeqSearch(n.key.name, input[inputPos:],
				[]string{"am", "pm", "AM", "PM"}, tfc.pm,
				func(i int) int { return i % 2 })
			if err != nil {
				return time.Time{}, err
			}
			tfc.pm = v
			inputPos += l
			tfc.is12clock = true
		case "hh", "HH", "hh12", "HH12":
			v, l, err := fromCharParseIntLen(input[inputPos:], 2, formatNodes[i:], tfc.hh)
			if err != nil {
				return time.Time{}, err
			}
			tfc.hh = v
			inputPos += l
			tfc.is12clock = true
			inputPos += skipTh(n.postfix)
		case "hh24", "HH24":
			v, l, err := fromCharParseIntLen(input[inputPos:], 2, formatNodes[i:], tfc.hh)
			if err != nil {
				return time.Time{}, err
			}
			tfc.hh = v
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "mi", "MI":
			v, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], tfc.mi)
			if err != nil {
				return time.Time{}, err
			}
			tfc.mi = v
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "ss", "SS":
			v, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], tfc.ss)
			if err != nil {
				return time.Time{}, err
			}
			tfc.ss = v
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "ms", "MS":
			// millisecond
			v, l, err := fromCharParseIntLen(input[inputPos:], 3, formatNodes[i:], tfc.ms)
			if err != nil {
				return time.Time{}, err
			}
			// 25 is 0.25 and 250 is 0.25 too; 025 is 0.025 and not 0.25
			switch l {
			case 1:
				tfc.ms = v * 100
			case 2:
				tfc.ms = v * 10
			default:
				tfc.ms = v
			}
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "ff1", "FF1", "ff2", "FF2", "ff3", "FF3", "ff4", "FF4", "ff5", "FF5", "ff6", "FF6":
			switch n.key.name {
			case "ff1", "FF1":
				tfc.ff = 1
			case "ff2", "FF2":
				tfc.ff = 2
			case "ff3", "FF3":
				tfc.ff = 3
			case "ff4", "FF4":
				tfc.ff = 4
			case "ff5", "FF5":
				tfc.ff = 5
			case "ff6", "FF6":
				tfc.ff = 6
			}
			fallthrough
		case "us", "US":
			l := 6
			if tfc.ff != 0 {
				l = tfc.ff
			}
			// microsecond
			v, l, err := fromCharParseIntLen(input[inputPos:], l, formatNodes[i:], tfc.us)
			if err != nil {
				return time.Time{}, err
			}
			switch l {
			case 1:
				tfc.us = v * 100000
			case 2:
				tfc.us = v * 10000
			case 3:
				tfc.us = v * 1000
			case 4:
				tfc.us = v * 100
			case 5:
				tfc.us = v * 10
			default:
				tfc.us = v
			}
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "ssss", "sssss", "SSSS", "SSSSS":
			v, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], tfc.ssss)
			if err != nil {
				return time.Time{}, err
			}
			tfc.ssss = v
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "tz", "TZ":
			// TODO: implement this
			return time.Time{}, errors.Errorf(`formatting TZ is not supported yet`)
		case "of", "OF":
			// OF is equivalent to TZH or TZH:TZM
			// see TZH comments below
			if input[inputPos] == '-' {
				tfc.tzsign = -1
				inputPos++
			} else if input[inputPos] == '+' || input[inputPos] == ' ' {
				tfc.tzsign = +1
				inputPos++
			} else {
				if extraSkip > 0 && input[inputPos-1] == '-' {
					tfc.tzsign = -1
				} else {
					tfc.tzsign = +1
				}
			}
			v, l, err := fromCharParseIntLen(input[inputPos:], 2, formatNodes[i:], tfc.tzh)
			if err != nil {
				return time.Time{}, err
			}
			tfc.tzh = v
			inputPos += l
			if inputPos < len(input[inputPos:]) && input[inputPos] == ':' {
				inputPos++
				v, l, err = fromCharParseIntLen(input[inputPos:], 2, formatNodes[i:], tfc.tzm)
				if err != nil {
					return time.Time{}, err
				}
				tfc.tzm = v
				inputPos += l
			}
		case "tzh", "TZH":
			if input[inputPos] == '-' {
				tfc.tzsign = -1
				inputPos++
			} else if input[inputPos] == '+' || input[inputPos] == ' ' {
				tfc.tzsign = +1
				inputPos++
			} else {
				if extraSkip > 0 && input[inputPos-1] == '-' {
					tfc.tzsign = -1
				} else {
					tfc.tzsign = +1
				}
			}
			v, l, err := fromCharParseIntLen(input[inputPos:], 2, formatNodes[i:], tfc.tzh)
			if err != nil {
				return time.Time{}, err
			}
			tfc.tzh = v
			inputPos += l
		case "tzm", "TZM":
			// assign positive timezone sign if TZH was not seen before
			if tfc.tzsign == 0 {
				tfc.tzsign = +1
			}
			v, l, err := fromCharParseIntLen(input[inputPos:], 2, formatNodes[i:], tfc.tzm)
			if err != nil {
				return time.Time{}, err
			}
			tfc.tzm = v
			inputPos += l
		case "a.d.", "b.c.", "A.D.", "B.C.":
			v, l, err := fromCharSeqSearch(n.key.name, input[inputPos:],
				[]string{"a.d.", "b.c.", "A.D.", "B.C."}, tfc.bc,
				func(i int) int { return i % 2 })
			if err != nil {
				return time.Time{}, err
			}
			tfc.bc = v
			inputPos += l
		case "ad", "bc", "AD", "BC":
			v, l, err := fromCharSeqSearch(n.key.name, input[inputPos:],
				[]string{"ad", "bc", "AD", "BC"}, tfc.bc,
				func(i int) int { return i % 2 })
			if err != nil {
				return time.Time{}, err
			}
			tfc.bc = v
			inputPos += l
		case "month", "Month", "MONTH":
			v, l, err := fromCharSeqSearch(n.key.name, input[inputPos:],
				[]string{"January", "February", "March", "April", "May", "June", "July",
					"August", "September", "October", "November", "December"}, tfc.mm,
				func(i int) int { return i })
			if err != nil {
				return time.Time{}, err
			}
			tfc.mm = v + 1
			inputPos += l
		case "mon", "Mon", "MON":
			v, l, err := fromCharSeqSearch(n.key.name, input[inputPos:],
				[]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun",
					"Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}, tfc.mm,
				func(i int) int { return i })
			if err != nil {
				return time.Time{}, err
			}
			tfc.mm = v + 1
			inputPos += l
		case "mm", "MM":
			val, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], tfc.mm)
			if err != nil {
				return time.Time{}, err
			}
			tfc.mm = val
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "day", "Day", "DAY":
			v, l, err := fromCharSeqSearch(n.key.name, input[inputPos:],
				[]string{"Sunday", "Monday", "Tuesday", "Wednesday",
					"Thursday", "Friday", "Saturday"}, tfc.d,
				func(i int) int { return i })
			if err != nil {
				return time.Time{}, err
			}
			tfc.d = v + 1
			inputPos += l
		case "dy", "Dy", "DY":
			v, l, err := fromCharSeqSearch(n.key.name, input[inputPos:],
				[]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}, tfc.d,
				func(i int) int { return i })
			if err != nil {
				return time.Time{}, err
			}
			tfc.d = v + 1
			inputPos += l
		case "ddd", "DDD":
			v, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], tfc.ddd)
			if err != nil {
				return time.Time{}, err
			}
			tfc.ddd = v
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "iddd", "IDDD":
			v, l, err := fromCharParseIntLen(input[inputPos:], 3, formatNodes[i:], tfc.ddd)
			if err != nil {
				return time.Time{}, err
			}
			tfc.ddd = v
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "dd", "DD":
			v, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], tfc.dd)
			if err != nil {
				return time.Time{}, err
			}
			tfc.dd = v
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "d", "D":
			v, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], tfc.d)
			if err != nil {
				return time.Time{}, err
			}
			tfc.d = v
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "id", "ID":
			v, l, err := fromCharParseIntLen(input[inputPos:], 1, formatNodes[i:], tfc.dd)
			if err != nil {
				return time.Time{}, err
			}
			// Shift numbering to match Gregorian where Sunday = 1
			if v+1 > 7 {
				tfc.d = 1
			} else {
				tfc.d = v
			}
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "ww", "WW", "iw", "IW":
			v, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], tfc.ww)
			if err != nil {
				return time.Time{}, err
			}
			tfc.ww = v
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "q", "Q":
			// we parse but ignore
			_, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], 0)
			if err != nil {
				return time.Time{}, err
			}
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "cc", "CC":
			v, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], tfc.cc)
			if err != nil {
				return time.Time{}, err
			}
			tfc.cc = v
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "y,yyy", "Y,YYY":
			var millennia, years int

			parts := strings.Split(input[inputPos:], ",")
			if len(parts) < 2 {
				return time.Time{}, errors.Errorf(`invalid input string for "Y,YYY"`)
			}

			millenniaStr := strings.TrimSpace(parts[0])
			yearsStr := strings.TrimSpace(parts[1])

			matched := 0
			if m, err := strconv.Atoi(millenniaStr); err == nil {
				millennia = m
				matched++
			}

			used := 0
			for _, l := range yearsStr {
				if unicode.IsDigit(l) {
					used += 1
				} else {
					break
				}
			}

			if used != 3 {
				return time.Time{}, errors.Errorf(`invalid input string for "Y,YYY"`)
			}

			if y, err := strconv.Atoi(yearsStr[:used]); err == nil {
				years = y
				matched++
			}

			if matched < 2 {
				return time.Time{}, errors.Errorf(`invalid input string for "Y,YYY"`)
			}

			// Check for overflow on multiplication
			if millennia > math.MaxInt32/1000 || millennia < math.MinInt32/1000 {
				return time.Time{}, errors.Errorf(`value for "Y,YYY" in source string is out of range`)
			}
			millennia *= 1000

			// Check for overflow on addition
			if (years > 0 && millennia > math.MaxInt32-years) || (years < 0 && millennia < math.MinInt32-years) {
				return time.Time{}, errors.Errorf(`value for "Y,YYY" in source string is out of range`)
			}
			years += millennia

			tfc.year = years
			tfc.yysz = 4
			inputPos += 5
			inputPos += skipTh(n.postfix)
		case "yyyy", "YYYY", "iyyy", "IYYY":
			v, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], tfc.year)
			if err != nil {
				return time.Time{}, err
			}
			tfc.year = v
			tfc.yysz = 4
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "yyy", "YYY", "iyy", "IYY":
			v, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], tfc.year)
			if err != nil {
				return time.Time{}, err
			}
			if l < 4 {
				tfc.year = adjustPartialYearTo2020(v)
			} else {
				tfc.year = v
			}
			tfc.yysz = 3
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "yy", "YY", "iy", "IY":
			v, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], tfc.year)
			if err != nil {
				return time.Time{}, err
			}
			if l < 4 {
				tfc.year = adjustPartialYearTo2020(v)
			} else {
				tfc.year = v
			}
			tfc.yysz = 2
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "y", "Y", "i", "I":
			v, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], tfc.year)
			if err != nil {
				return time.Time{}, err
			}
			if l < 4 {
				tfc.year = adjustPartialYearTo2020(v)
			} else {
				tfc.year = v
			}
			tfc.yysz = 1
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "rm", "RM":
			v, l, err := fromCharSeqSearch(n.key.name, input[inputPos:],
				[]string{"xii", "xi", "x", "ix", "viii", "vii", "vi", "v", "iv", "iii", "ii", "i"}, tfc.mm,
				func(i int) int { return monthsPerYear - i })
			if err != nil {
				return time.Time{}, err
			}
			tfc.mm = v
			inputPos += l
		case "w", "W":
			v, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], tfc.w)
			if err != nil {
				return time.Time{}, err
			}
			tfc.w = v
			inputPos += l
			inputPos += skipTh(n.postfix)
		case "j", "J":
			v, l, err := fromCharParseIntLen(input[inputPos:], n.key.len, formatNodes[i:], tfc.j)
			if err != nil {
				return time.Time{}, err
			}
			tfc.j = v
			inputPos += l
			inputPos += skipTh(n.postfix)
		default:
		}

		/* Ignore all spaces after fields */
		if !fxMode {
			extraSkip = 0
			for inputPos < len(input) && isSpace(input[inputPos]) {
				inputPos++
				extraSkip++
			}
		}
	}

	for inputPos < len(input) && isSpace(input[inputPos]) {
		inputPos++
	}

	if inputPos < len(input) {
		return time.Time{}, errors.Errorf(`trailing characters remain in input string after datetime format`)
	}

	return getTime(ctx, tfc)
}

// getTime converts parsed time components from tmFromChar into a time.Time value
func getTime(ctx *sql.Context, t tmFromChar) (time.Time, error) {
	var (
		year       int
		month      int
		day        int
		hour       int
		minute     int
		second     int
		nanosecond int
		fsec       int
		gmtOffset  int
		has_tz     bool
	)

	if t.ssss != 0 {
		x := t.ssss
		hour = x / duration.SecsPerHour
		x %= duration.SecsPerHour
		minute = x / duration.SecsPerMinute
		x %= duration.SecsPerMinute
		second = x
	}

	if t.ss != 0 {
		second = t.ss
	}
	if t.mi != 0 {
		minute = t.mi
	}
	if t.hh != 0 {
		hour = t.hh
	}

	if t.is12clock {
		if hour < 1 || hour > hoursPerDay/2 {
			return time.Time{}, errors.Errorf(`hour "%d" is invalid for the 12-hour clock`, hour)
		}
		if t.pm != 0 && hour < hoursPerDay/2 {
			hour += hoursPerDay / 2
		} else if t.pm == 0 && hour == hoursPerDay/2 {
			hour = 0
		}
	}

	if t.year != 0 {
		// If CC and YY are provided, combine them considering century boundaries
		if t.cc != 0 && t.yysz <= 2 {
			if t.bc != 0 {
				t.cc = -t.cc
			}
			year = t.year % 100
			if year != 0 {
				if t.cc >= 0 {
					// year += (t.cc - 1) * 100
					tmp, err := checkForOverflow((t.cc - 1) * 100)
					if err != nil {
						return time.Time{}, err
					}
					year, err = checkForOverflow(year + tmp)
					if err != nil {
						return time.Time{}, err
					}
				} else {
					// year = (t.cc + 1) * 100 - year + 1
					tmp, err := checkForOverflow((t.cc + 1) * 100)
					if err != nil {
						return time.Time{}, err
					}
					tmp, err = checkForOverflow(tmp - year)
					if err != nil {
						return time.Time{}, err
					}
					year, err = checkForOverflow(tmp + 1)
					if err != nil {
						return time.Time{}, err
					}
				}
			} else {
				// find century year for dates ending in "00"
				year = t.cc * 100
				if t.cc < 0 {
					year = year + 1
				}
			}
		} else {
			// If a 4-digit year is provided, we use that and ignore CC.
			year = t.year
			if t.bc != 0 {
				year = -year
			}
			// correct for our representation of BC years
			if year < 0 {
				year++
			}
		}
	} else if t.cc != 0 {
		// use first year of century
		if t.bc != 0 {
			t.cc = -t.cc
		}
		if t.cc >= 0 {
			// Add 1 because 21st century started in 2001, not 2000
			// year = (t.cc - 1) * 100 + 1
			tmp, err := checkForOverflow((t.cc - 1) * 100)
			if err != nil {
				return time.Time{}, err
			}
			t.year, err = checkForOverflow(tmp + 1)
			if err != nil {
				return time.Time{}, err
			}
		} else {
			// Add 1 because year 599 represents 600 BC
			// year = t.cc * 100 + 1;
			tmp, err := checkForOverflow(t.cc * 100)
			if err != nil {
				return time.Time{}, err
			}
			t.year, err = checkForOverflow(tmp + 1)
			if err != nil {
				return time.Time{}, err
			}
		}
	}

	if t.j != 0 {
		year, month, day = j2Date(t.j)
	}

	if t.ww != 0 {
		if t.mode == fromCharDateISOWEEK {
			// If t.d is not set, then the date is left at the beginning of
			// the ISO week (Monday).
			if t.d != 0 {
				year, month, day = isoWeekDate2Date(year, t.ww, t.d)
			} else {
				year, month, day = isoWeek2Date(year, t.ww)
			}
		} else {
			// t.ddd = (t.ww - 1) * 7 + 1
			var err error
			if t.ddd, err = checkForOverflow(t.ww - 1); err != nil {
				return time.Time{}, err
			}
			if t.ddd, err = checkForOverflow(t.ddd * 7); err != nil {
				return time.Time{}, err
			}
			if t.ddd, err = checkForOverflow(t.ddd + 1); err != nil {
				return time.Time{}, err
			}
		}
	}

	if t.w != 0 {
		// t.dd = (t.w - 1) * 7 + 1
		var err error
		if t.dd, err = checkForOverflow(t.w - 1); err != nil {
			return time.Time{}, err
		}
		if t.dd, err = checkForOverflow(t.dd * 7); err != nil {
			return time.Time{}, err
		}
		if t.dd, err = checkForOverflow(t.dd + 1); err != nil {
			return time.Time{}, err
		}
	}

	if t.dd != 0 {
		day = t.dd
	}

	if t.mm != 0 {
		month = t.mm
	}

	if t.ddd != 0 && (month <= 1 || day <= 1) {
		if year == 0 && t.bc == 0 {
			return time.Time{}, errors.Errorf(`cannot calculate day of year without year information`)
		}
		if t.mode == fromCharDateISOWEEK {
			// zeroth day of the ISO year, in Julian
			j0 := isoWeek2J(year, 1) - 1
			year, month, day = j2Date(j0 + t.ddd)
		} else {
			ysum := [2][13]int{
				{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334, 365},
				{0, 31, 60, 91, 121, 152, 182, 213, 244, 274, 305, 335, 366},
			}

			var y []int
			if isLeapYear(year) {
				y = ysum[1][:]
			} else {
				y = ysum[0][:]
			}

			var i int
			for i = 1; i <= monthsPerYear; i++ {
				if t.ddd <= y[i] {
					break
				}
			}

			if month <= 1 {
				month = i // mon is 0-11, so it might need adjustment later
			}

			if day <= 1 {
				day = t.ddd - y[i-1]
			}
		}
	}

	if t.ms != 0 {
		// *fsec += t.ms * 1000
		var err error
		fsec, err = checkForOverflow(t.ms * 1000)
		if err != nil {
			return time.Time{}, err
		}
	}
	if t.us != 0 {
		fsec += t.us
	}

	if hour < 0 || hour >= hoursPerDay ||
		minute < 0 || minute >= 60 ||
		second < 0 || second >= duration.SecsPerMinute ||
		fsec < 0 || fsec > usecsPerSecs {
		return time.Time{}, errors.Errorf(`date/time field value out of range`)
	}

	if t.tzsign != 0 {
		// TZH and/or TZM fields
		if t.tzh < 0 || t.tzh > maxTzdispHour || t.tzm < 0 || t.tzm >= minsPerHour {
			return time.Time{}, errors.Errorf(`date/time field value out of range`)
		}
		has_tz = true
		gmtOffset = (t.tzh*minsPerHour + t.tzm) * duration.SecsPerMinute
		if t.tzsign > 0 {
			gmtOffset = -gmtOffset
		}
	}

	if fsec != 0 {
		nanosecond = fsec * int(time.Microsecond)
	}

	// Get server timezone to interpret the parsed timestamp
	serverLoc, err := GetServerLocation(ctx)
	if err != nil {
		return time.Time{}, err
	}

	if year == 0 {
		year = 1970
	}
	if month == 0 {
		month = 1
	}
	if day == 0 {
		day = 1
	}

	// Create the time in the server timezone
	result := time.Date(year, time.Month(month), day, hour, minute, second, nanosecond, serverLoc)

	// Validate the created time (handles invalid dates like Feb 30)
	if result.Year() != year || int(result.Month()) != month || result.Day() != day {
		return time.Time{}, errors.Errorf(`date/time field value out of range`)
	}

	// Convert the result to specified timezone
	if has_tz {
		_, serverLocOffset := result.Zone()
		result = result.Add(time.Duration(gmtOffset+serverLocOffset) * time.Second)
	}

	return result, nil
}

// fromCharSeqSearch finds a matching string from an array and returns its index and length
func fromCharSeqSearch(keyName string, input string, arr []string, curVal int, calculateVal func(int) int) (int, int, error) {
	var v int
	matchFound := false
	l := 0
	for i, elem := range arr {
		// TODO: support collation when matching
		if strings.HasPrefix(strings.ToLower(input), strings.ToLower(elem)) {
			v = i
			l += len(elem)
			matchFound = true
			break
		}
	}
	if !matchFound {
		s := strings.Split(input, " ")
		if len(s) > 0 {
			l = len(s[0])
		}
		return 0, 0, errors.Errorf(`invalid value "%s" for "%s"`, input[:l], keyName)
	}

	val, err := verifyVal(keyName, curVal, calculateVal(v))
	if err != nil {
		return 0, 0, err
	}
	return val, l, nil
}

// verifyVal checks if the current value conflicts with a new value for the same field
func verifyVal(keyName string, curVal, newVal int) (int, error) {
	if curVal != 0 && curVal != newVal {
		return 0, errors.Errorf(`conflicting values for "%s" field in formatting string`, keyName)
	}
	return newVal, nil
}

// checkForOverflow validates that a value fits within int32 range
func checkForOverflow(v int) (int, error) {
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, errors.Errorf(`date/time field value out of range`)
	}
	return v, nil
}

// isoWeekDate2Date converts an ISO week date (year, week, weekday) to a Gregorian date using PostgreSQL's algorithm
func isoWeekDate2Date(year, isoweek, wday int) (int, int, int) {
	jday := isoWeek2J(year, isoweek)
	if wday > 1 {
		jday += wday - 2
	} else {
		jday += 6
	}

	return j2Date(jday)
}

// isoWeek2Date converts a year and ISO week to the Gregorian date of the first day of that week
func isoWeek2Date(year, woy int) (int, int, int) {
	// 1. Calculate the Julian day for the start of the ISO week.
	jday := isoWeek2J(year, woy)

	// 2. Convert that Julian day number to a Gregorian date.
	return j2Date(jday)
}

// isoWeek2J returns the Julian day corresponding to the first day (Monday) of an ISO 8601 year and week
func isoWeek2J(year, week int) int {
	var day0, day4 int
	// fourth day of current year
	day4 = date2J(year, 1, 4)
	// day0 == offset to first day of week (Monday)
	day0 = j2Day(day4 - 1)
	return (week-1)*7 + (day4 - day0)
}

// date2J converts a Gregorian date to a Julian day number using PostgreSQL's algorithm
func date2J(year, month, day int) int {
	var julian int
	var century int

	if month > 2 {
		month++
		year += 4800
	} else {
		month += 13
		year += 4799
	}

	century = year / 100
	julian = year*365 - 32167
	julian += year/4 - century + century/4
	julian += 7834*month/256 + day

	return julian
}

// j2Day converts a Julian day number to a day of the week (0=Sunday, 6=Saturday)
// This function implements the specific PostgreSQL algorithm.
func j2Day(date int) int {
	date++
	date %= 7
	if date < 0 {
		date += 7
	}

	return date
}

// j2Date converts a Julian day number back to a Gregorian date.
// This function implements the inverse of the date2J algorithm.
func j2Date(jd int) (year, month, day int) {
	var julian uint
	var quad uint
	var extra uint
	var y int

	julian = uint(jd)
	julian += 32044
	quad = julian / 146097
	extra = (julian-quad*146097)*4 + 3
	julian += 60 + quad*3 + extra/146097
	quad = julian / 1461
	julian -= quad * 1461
	y = int(julian*4) / 1461

	if y != 0 {
		julian = (julian+305)%365 + 123
	} else {
		julian = (julian+306)%366 + 123
	}

	y += int(quad * 4)
	year = y - 4800
	quad = julian * 2141 / 65536
	day = int(julian) - 7834*int(quad)/256
	month = int(quad+10)%monthsPerYear + 1

	return year, month, day
}

// isLeapYear returns true if the given year is a leap year.
func isLeapYear(year int) bool {
	// A year is a leap year if it has 366 days.
	// We can check this by getting the Day of the Year for March 1st
	// of the given year. If it's day 61, it means February had 29 days.
	return time.Date(year, time.March, 1, 0, 0, 0, 0, time.UTC).YearDay() == 61
}

// adjustPartialYearTo2020 adjusts partial years to reasonable 4-digit years based on PostgreSQL rules
func adjustPartialYearTo2020(year int) int {
	/* Force 0-69 into the 2000's */
	if year < 70 {
		return year + 2000
	} else if year < 100 {
		/* Force 70-99 into the 1900's */
		return year + 1900
	} else if year < 520 {
		/* Force 100-519 into the 2000's */
		return year + 2000
	} else if year < 1000 {
		/* Force 520-999 into the 1000's */
		return year + 1000
	} else {
		return year
	}
}

// skipTh returns 2 if there is a "TH" or "th" postfix; otherwise returns 0
func skipTh(suff *keywordSuffix) int {
	if suff != nil && suff.isPostFix && (suff.name == "th" || suff.name == "TH") {
		return 2
	} else {
		return 0
	}
}

// fromCharParseIntLen parses an integer from input string with specified length constraints
func fromCharParseIntLen(input string, length int, nodes []*formatNode, curVal int) (int, int, error) {
	// Skip whitespace
	input = strings.TrimLeftFunc(input, unicode.IsSpace)
	n := nodes[0]

	isFMorIsNextSeperator := (n.suffix != nil && (n.suffix.name == "fm" || n.suffix.name == "FM")) || isNextSeperator(nodes)

	if !isFMorIsNextSeperator && len(input) < length {
		return 0, 0, errors.Errorf(`source string too short for "%s" formatting field`, n.key.name)
	}

	used := 0
	negative := false
	for i, l := range input {
		// if fill mode is true, slurp as many characters as we can get
		// otherwise, get only given length many
		if i == 0 && l == '-' {
			negative = true
		} else if unicode.IsDigit(l) && (isFMorIsNextSeperator || used < length) {
			used++
		} else {
			break
		}
	}

	s := input[:used]
	if negative {
		s = input[:used+1]
	}
	res, err := strconv.ParseInt(s, 10, 64)
	if err != nil || used == 0 || (!isFMorIsNextSeperator && used > 0 && used < length) {
		if used < length && len(input) >= length {
			used = length
		}
		return 0, 0, errors.Errorf(`invalid value "%s" for "%s"`, input[:used], n.key.name)
	}

	if res < math.MinInt32 || res > math.MaxInt32 {
		return 0, 0, errors.Errorf(`value for "%s" in source string is out of range`, n.key.name)
	}

	v, err := verifyVal(n.key.name, curVal, int(res))
	if err != nil {
		return 0, 0, err
	}
	// result, length of result
	return v, len(s), nil
}

// isNextSeperator checks if the next node in the format is a separator or end of format
func isNextSeperator(nodes []*formatNode) bool {
	n := nodes[0]
	if len(nodes) == 1 {
		return true
	}
	if n.typ == nodeTypeACTION && n.postfix != nil && n.postfix.isPostFix && (n.postfix.name == "th" || n.postfix.name == "TH") {
		return true
	}
	// next node
	n = nodes[1]
	if n.typ == nodeTypeACTION {
		return !n.key.isDigit
	} else if len(n.characters) != 0 && unicode.IsDigit(rune(n.characters[0])) {
		return false
	}
	return true
}
