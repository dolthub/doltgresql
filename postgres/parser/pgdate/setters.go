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

type timezone struct {
	abbreviation string
	// there are more than one identifier for some time zones.
	identifier string
}

// TimezoneMapping is a map timezone abbreviations to their timezone offset.
// These abbreviations are taken from:
// https://github.com/postgres/postgres/blob/master/src/timezone/known_abbrevs.txt
var TimezoneMapping = map[string]timezone{
	"ACDT": {"+10:30", "Australia/Adelaide"},
	"ACST": {"+09:30", ""},
	"ADT":  {"-03:00", "America/Glace_Bay"},
	"AEDT": {"+11:00", "Australia/Sydney"},
	"AEST": {"+10:00", "Australia/Brisbane"},
	"AKDT": {"-08:00", "America/Anchorage"},
	"AKST": {"-09:00", "America/Anchorage"},
	"AST":  {"-04:00", "America/Barbados"},
	"AWST": {"+08:00", "Australia/Perth"},
	"BST":  {"+01:00", "Europe/London"},
	"CAT":  {"+02:00", "Africa/Juba"},
	"CDT":  {"-05:00", "America/Chicago"},
	//"CDT":  "",
	"CEST": {"+02:00", "Europe/Berlin"},
	"CET":  {"+01:00", "Africa/Algiers"},
	"CST":  {"-06:00", "Asia/Shanghai"},
	//"CST":  "",
	//"CST":  "",
	"ChST": {"+10:00", "Pacific/Guam"},

	"EAT":  {"+03:00", "Africa/Nairobi"},
	"EDT":  {"-04:00", "America/Detroit"},
	"EEST": {"+03:00", "Africa/Cairo"},
	"EET":  {"+02:00", "Africa/Tripoli"},
	"EST":  {"-05:00", "America/Cancun"},

	"GMT": {"+00:00", "Africa/Monrovia"},
	"HDT": {"-09:00", "America/Adak"},
	"HKT": {"+08:00", "Asia/Hong_Kong"},
	"HST": {"-10:00", "Pacific/Honolulu"},
	"IDT": {"+03:00", "Asia/Jerusalem"},
	"IST": {"+02:00", "Asia/Jerusalem"}, // TODO: multiple abbr
	//"IST": "",
	//"IST": "",
	"JST": {"+09:00", "Asia/Tokyo"},
	"KST": {"+09:00", "Asia/Seoul"},

	"MDT":  {"-06:00", "America/Edmonton"},
	"MEST": {"", ""}, // TODO
	"MET":  {"", ""}, // TODO
	"MSK":  {"+03:00", "Europe/Moscow"},
	"MST":  {"-07:00", "America/Edmonton"},
	"NDT":  {"-02:30", "America/St_Johns"},
	"NST":  {"-03:30", "America/St_Johns"},
	"NZDT": {"+13:00", "Pacific/Auckland"},
	"NZST": {"+12:00", "Pacific/Auckland"},

	"PDT": {"-07:00", "America/Los_Angeles"},
	"PKT": {"+05:00", "Asia/Manila"},
	"PST": {"-08:00", "America/Los_Angeles"},
	//"PST": { "",
	"SAST": {"+02:00", "Africa/Johannesburg"},
	"SST":  {"-11:00", "Pacific/Pago_Pago"},
	"UCT":  {"", ""}, // TODO

	// UTC has been removed from this list.
	"WEST": {"+01:00", "Atlantic/Canary"},
	"WET":  {"+00:00", "Atlantic/Canary"},
	"WIB":  {"+07:00", "Asia/Jakarta"},
	"WIT":  {"+09:00", "Asia/Jayapura"},
	"WITA": {"+08:00", "Asia/Makassar"},
}

func init() {
	for tz, offset := range TimezoneMapping {
		if offset.abbreviation == "" {
			keywordSetters[strings.ToLower(tz)] = fieldSetterUnsupportedAbbreviation
		} else {
			keywordSetters[strings.ToLower(tz)] = fieldSetterLocation(offset.abbreviation)
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

func GetTimezoneIdentifier(tz string) (string, string) {
	if tzid, ok := TimezoneMapping[tz]; ok {
		return tzid.abbreviation, tzid.identifier
	}
	return "", ""
}
