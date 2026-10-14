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

package config

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

// memoryUnits are the multipliers of the memory units, in bytes, in the order that they're listed in hints.
var memoryUnits = []parameterUnit{
	{"B", 1},
	{"kB", 1 << 10},
	{"MB", 1 << 20},
	{"GB", 1 << 30},
	{"TB", 1 << 40},
}

// timeUnits are the multipliers of the time units, in microseconds, in the order that they're listed in hints.
var timeUnits = []parameterUnit{
	{"us", 1},
	{"ms", 1000},
	{"s", 1000 * 1000},
	{"min", 60 * 1000 * 1000},
	{"h", 60 * 60 * 1000 * 1000},
	{"d", 24 * 60 * 60 * 1000 * 1000},
}

// parameterUnit is a unit that a value may be given in, along with its multiplier relative to the smallest unit of
// the same kind.
type parameterUnit struct {
	name       string
	multiplier float64
}

// numericParameterValue matches a numeric value that is optionally followed by a unit, which is how Postgres accepts
// the values of integer and real parameters.
var numericParameterValue = regexp.MustCompile(`^\s*([+-]?(?:[0-9]+\.?[0-9]*|\.[0-9]+)(?:[eE][+-]?[0-9]+)?)\s*([A-Za-z]*)\s*$`)

// parseNumericValue parses the string value of an integer or real parameter, converting any unit that is given into
// the parameter's base unit. This matches Postgres's parse_int and parse_real. Integer values are rounded to the
// nearest integer and returned as an int64, while real values are returned as a float64.
func (p *Parameter) parseNumericValue(str string, isInt bool) (any, error) {
	matches := numericParameterValue.FindStringSubmatch(str)
	if matches == nil {
		return nil, p.invalidValueError(str, "")
	}
	val, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return nil, p.invalidValueError(str, "")
	}
	if unit := matches[2]; len(unit) > 0 {
		units, baseMultiplier, ok := p.units()
		if !ok {
			return nil, p.invalidValueError(str, "")
		}
		found := false
		for _, u := range units {
			if u.name == unit {
				val = val * u.multiplier / baseMultiplier
				found = true
				break
			}
		}
		if !found {
			return nil, p.invalidValueError(str, unitsHint(units))
		}
	}
	if !isInt {
		return val, nil
	}
	val = math.RoundToEven(val)
	if val > math.MaxInt32 || val < math.MinInt32 {
		return nil, p.invalidValueError(str, "Value exceeds integer range.")
	}
	return int64(val), nil
}

// units returns the units that the parameter's value may be given in, along with the multiplier of the parameter's
// base unit. Returns false if the parameter's value does not have a unit.
func (p *Parameter) units() (units []parameterUnit, baseMultiplier float64, ok bool) {
	if len(p.Unit) == 0 {
		return nil, 0, false
	}
	// Some memory parameters are measured in blocks, such as "8kB", so we split the base unit into its count and unit
	count := 1.0
	unitName := p.Unit
	if idx := strings.IndexFunc(unitName, func(r rune) bool { return r < '0' || r > '9' }); idx > 0 {
		count, _ = strconv.ParseFloat(unitName[:idx], 64)
		unitName = unitName[idx:]
	}
	for _, units := range [][]parameterUnit{memoryUnits, timeUnits} {
		for _, u := range units {
			if u.name == unitName {
				return units, count * u.multiplier, true
			}
		}
	}
	return nil, 0, false
}

// invalidValueError returns the error for an invalid parameter value, with the given hint if it is not empty.
func (p *Parameter) invalidValueError(str string, hint string) error {
	err := pgerror.Newf(pgcode.InvalidParameterValue, `invalid value for parameter "%s": "%s"`, p.Name, str)
	if len(hint) > 0 {
		return errors.WithHint(err, hint)
	}
	return err
}

// unitsHint returns the hint that lists the valid units.
func unitsHint(units []parameterUnit) string {
	sb := strings.Builder{}
	sb.WriteString("Valid units for this parameter are ")
	for i, u := range units {
		if i == len(units)-1 {
			sb.WriteString(", and ")
		} else if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(`"` + u.name + `"`)
	}
	sb.WriteString(".")
	return sb.String()
}

// formatWithUnit formats the value of an integer or real parameter using the largest unit that represents the value
// exactly, which matches Postgres's convert_int_from_base_unit and convert_real_from_base_unit. Values that aren't
// positive, as well as values of parameters without a unit, are formatted without a unit.
func (p *Parameter) formatWithUnit(val any) (string, bool) {
	var num float64
	isInt := true
	switch v := val.(type) {
	case int:
		num = float64(v)
	case int8:
		num = float64(v)
	case int16:
		num = float64(v)
	case int32:
		num = float64(v)
	case int64:
		num = float64(v)
	case uint8:
		num = float64(v)
	case uint16:
		num = float64(v)
	case uint32:
		num = float64(v)
	case uint64:
		num = float64(v)
	case float32:
		num, isInt = float64(v), false
	case float64:
		num, isInt = v, false
	default:
		return "", false
	}
	units, baseMultiplier, ok := p.units()
	if !ok || num <= 0 {
		return "", false
	}
	for i := len(units) - 1; i >= 0; i-- {
		multiplier := units[i].multiplier / baseMultiplier
		converted := num / multiplier
		if multiplier <= 1 || converted == math.Trunc(converted) {
			if isInt {
				return strconv.FormatInt(int64(converted), 10) + units[i].name, true
			}
			return strconv.FormatFloat(converted, 'g', -1, 64) + units[i].name, true
		}
	}
	return "", false
}
