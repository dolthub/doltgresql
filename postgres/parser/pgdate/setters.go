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

// Copyright 2018 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package pgdate

import (
	"strconv"
	"strings"
	"time"

	"github.com/dolthub/doltgresql/postgres/parser/errorutil/unimplemented"
)

// The functions in this file are used by fieldExtract.Extract().

// A fieldSetter is a helper function to set one or more
// fields within a fieldExtract in response to user input.
// These functions are used by fieldExtract.Extract().
type fieldSetter func(p *fieldExtract, s string) error

var keywordSetters = map[string]fieldSetter{
	"jan":       fieldSetterMonth(1),
	"january":   fieldSetterMonth(1),
	"feb":       fieldSetterMonth(2),
	"february":  fieldSetterMonth(2),
	"mar":       fieldSetterMonth(3),
	"march":     fieldSetterMonth(3),
	"apr":       fieldSetterMonth(4),
	"april":     fieldSetterMonth(4),
	"may":       fieldSetterMonth(5),
	"jun":       fieldSetterMonth(6),
	"june":      fieldSetterMonth(6),
	"jul":       fieldSetterMonth(7),
	"july":      fieldSetterMonth(7),
	"aug":       fieldSetterMonth(8),
	"august":    fieldSetterMonth(8),
	"sep":       fieldSetterMonth(9),
	"sept":      fieldSetterMonth(9), /* 4-chars, too in pg */
	"september": fieldSetterMonth(9),
	"oct":       fieldSetterMonth(10),
	"october":   fieldSetterMonth(10),
	"nov":       fieldSetterMonth(11),
	"november":  fieldSetterMonth(11),
	"dec":       fieldSetterMonth(12),
	"december":  fieldSetterMonth(12),

	keywordYesterday: fieldSetterRelativeDate,
	keywordToday:     fieldSetterRelativeDate,
	keywordTomorrow:  fieldSetterRelativeDate,

	keywordEraAD: fieldSetterExact(fieldEra, fieldValueCE),
	keywordEraCE: fieldSetterExact(fieldEra, fieldValueCE),

	keywordEraBC:  fieldSetterExact(fieldEra, fieldValueBCE),
	keywordEraBCE: fieldSetterExact(fieldEra, fieldValueBCE),

	keywordAM: fieldSetterExact(fieldMeridian, fieldValueAM),
	keywordPM: fieldSetterExact(fieldMeridian, fieldValuePM),

	keywordAllBalls: fieldSetterUTC,
	keywordGMT:      fieldSetterUTC,
	keywordUTC:      fieldSetterUTC,
	keywordZ:        fieldSetterUTC,
	keywordZulu:     fieldSetterUTC,
}

// These abbreviations are taken from:
// https://github.com/postgres/postgres/blob/master/src/timezone/known_abbrevs.txt
var timezoneMapping = map[string]string{
	"ACDT": "+10:30",
	"ACST": "+09:30",
	"ADT":  "-03:00",
	"AEDT": "+11:00",
	"AEST": "+10:00",
	"AKDT": "-08:00",
	"AKST": "-09:00",
	"AST":  "-04:00",
	"AWST": "+08:00",
	"BST":  "+01:00",
	"CAT":  "+02:00",
	"CDT":  "-05:00",
	//"CDT":  "",
	"CEST": "+02:00",
	"CET":  "+01:00",
	"CST":  "-06:00",
	//"CST":  "",
	//"CST":  "",
	"ChST": "+10:00",

	"EAT":  "+03:00",
	"EDT":  "-04:00",
	"EEST": "+03:00",
	"EET":  "+02:00",
	"EST":  "-05:00",

	// GMT has been removed from this list.
	"HDT": "-09:00",
	"HKT": "+08:00",
	"HST": "-10:00",
	"IDT": "+03:00",
	"IST": "+02:00",
	//"IST":  "",
	//"IST":  "",
	"JST": "+09:00",
	"KST": "+09:00",

	"MDT":  "-06:00",
	"MEST": "", // TODO
	"MET":  "", // TODO
	"MSK":  "+03:00",
	"MST":  "-07:00",
	"NDT":  "-02:30",
	"NST":  "-03:30",
	"NZDT": "+13:00",
	"NZST": "+12:00",

	"PDT": "-07:00",
	"PKT": "+05:00",
	"PST": "-08:00",
	//"PST":  "",
	"SAST": "+02:00",
	"SST":  "-11:00",
	"UCT":  "", // TODO

	// UTC has been removed from this list.
	"WAT":  "+01:00",
	"WEST": "+01:00",
	"WET":  "+00:00",
	"WIB":  "+07:00",
	"WIT":  "+09:00",
	"WITA": "+08:00",
}

func init() {
	for tz, offset := range timezoneMapping {
		if offset == "" {
			keywordSetters[strings.ToLower(tz)] = fieldSetterUnsupportedAbbreviation
		} else {
			keywordSetters[strings.ToLower(tz)] = fieldSetterLocation(offset)
		}
	}
}

// fieldSetterExact returns a fieldSetter that unconditionally sets field to v.
func fieldSetterExact(field field, v int) fieldSetter {
	return func(p *fieldExtract, _ string) error {
		return p.Set(field, v)
	}
}

// fieldSetterJulianDate parses a value like "J2451187" to set
// the year, month, and day fields.
func fieldSetterJulianDate(fe *fieldExtract, s string) (bool, error) {
	if !strings.HasPrefix(s, "j") {
		return false, nil
	}
	date, err := strconv.Atoi(s[1:])
	if err != nil {
		return true, inputErrorf("could not parse julian date")
	}

	year, month, day := julianDayToDate(date)

	if err := fe.Set(fieldYear, year); err != nil {
		return true, err
	}
	if err := fe.Set(fieldMonth, month); err != nil {
		return true, err
	}
	return true, fe.Set(fieldDay, day)
}

// fieldSetterMonth returns a fieldSetter that unconditionally sets
// the month to the given value.
func fieldSetterMonth(month int) fieldSetter {
	return fieldSetterExact(fieldMonth, month)
}

// fieldSetterRelativeDate sets the year, month, and day
// in response to the inputs "yesterday", "today", and "tomorrow"
// relative to fieldExtract.now.
func fieldSetterRelativeDate(fe *fieldExtract, s string) error {
	var offset int
	switch s {
	case keywordYesterday:
		offset = -1
	case keywordToday:
	case keywordTomorrow:
		offset = 1
	}

	year, month, day := fe.now().AddDate(0, 0, offset).Date()

	if err := fe.Set(fieldYear, year); err != nil {
		return err
	}
	if err := fe.Set(fieldMonth, int(month)); err != nil {
		return err
	}
	return fe.Set(fieldDay, day)
}

// fieldSetterUTC unconditionally sets the timezone to UTC and
// removes the TZ fields from the wanted list.
func fieldSetterUTC(fe *fieldExtract, _ string) error {
	fe.location = time.UTC
	fe.wanted = fe.wanted.ClearAll(tzFields)
	return nil
}

// fieldSetterLocation unconditionally sets the timezone to given offset
// and removes the TZ fields from the wanted list.
func fieldSetterLocation(offset string) fieldSetter {
	tzSign := 1
	if offset[0] == '-' {
		tzSign = -1
	}
	v := strings.Split(offset[1:], ":")
	hour, err := strconv.ParseInt(v[0], 10, 32)
	if err != nil {
		return fieldSetterUnsupportedAbbreviation
	}
	minute, err := strconv.ParseInt(v[1], 10, 32)
	if err != nil {
		return fieldSetterUnsupportedAbbreviation
	}

	return func(fe *fieldExtract, _ string) error {
		err = fe.Set(fieldTZHour, int(hour))
		err = fe.Set(fieldTZMinute, int(minute))
		fe.tzSign = tzSign
		fe.location = fe.MakeLocation()
		fe.wanted = fe.wanted.ClearAll(tzFields)
		return nil
	}
}

// fieldSetterUnsupportedAbbreviation always returns an error, but
// captures the abbreviation in telemetry.
func fieldSetterUnsupportedAbbreviation(_ *fieldExtract, s string) error {
	return unimplemented.NewWithIssueDetail(31710, s, "timestamp abbreviations not supported")
}
