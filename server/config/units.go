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
	"slices"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

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
// the values of integer and real parameters. Whitespace matches C's isspace, which includes the vertical tab.
var numericParameterValue = regexp.MustCompile(`^[\t\n\v\f\r ]*([+-]?(?:0[xX][0-9a-fA-F]+|(?:[0-9]+\.?[0-9]*|\.[0-9]+)(?:[eE][+-]?[0-9]+)?))[\t\n\v\f\r ]*([A-Za-z]*)[\t\n\v\f\r ]*$`)

// minNormalFloat64 is the smallest positive normal float64. Smaller nonzero values underflow in C's strtod.
var minNormalFloat64 = math.Float64frombits(0x0010000000000000)

// HasUnit returns whether the given parameter has a unit, meaning that its value is displayed with a unit.
func HasUnit(name string) bool {
	p, ok := postgresConfigParameters[strings.ToLower(name)].(*Parameter)
	return ok && len(p.Unit) > 0
}

// HasUnitInRegistry returns whether the parameter, as defined in the given registry, has a unit.
func HasUnitInRegistry(vars sql.SystemVariableRegistry, name string) bool {
	definition, _, ok := vars.GetGlobal(strings.ToLower(name))
	if !ok {
		return false
	}
	p, ok := definition.(*Parameter)
	return ok && len(p.Unit) > 0
}

// parseNumericValue parses the string value of an integer or real parameter, converting any unit that is given into
// the parameter's base unit. This matches Postgres's parse_int and parse_real. Integer values are rounded to the
// nearest integer and returned as an int64, while real values are returned as a float64.
func (p *Parameter) parseNumericValue(str string, isInt bool) (any, error) {
	matches := numericParameterValue.FindStringSubmatch(str)
	if matches == nil {
		return nil, p.invalidValueError(str, "")
	}
	val, ok := parseNumber(matches[1], isInt)
	if !ok {
		return nil, p.invalidValueError(str, "")
	}
	if unit := matches[2]; len(unit) > 0 {
		units, baseMultiplier, ok := p.units()
		if !ok {
			return nil, p.invalidValueError(str, "")
		}
		idx := slices.IndexFunc(units, func(u parameterUnit) bool { return u.name == unit })
		if idx < 0 {
			return nil, p.invalidValueError(str, unitsHint(units))
		}
		val *= units[idx].multiplier
		// Fractional values are rounded to the nearest multiple of the next smaller unit, as in convert_to_base_unit
		if idx > 0 {
			val = math.RoundToEven(val/units[idx-1].multiplier) * units[idx-1].multiplier
		}
		val /= baseMultiplier
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

// parseNumber parses a number the same way as Postgres. Integer parameters first try strtol with a base of 0, which
// accepts hexadecimal and octal values, and fall back to strtod for fractional values, exponents, and overflow. Real
// parameters only use strtod.
func parseNumber(number string, isInt bool) (float64, bool) {
	if isInt {
		integer, err := strconv.ParseInt(number, 0, 64)
		if err == nil {
			return float64(integer), true
		}
		if !strings.ContainsAny(number, ".eE") && !errors.Is(err, strconv.ErrRange) {
			return 0, false
		}
	}
	val, err := strconv.ParseFloat(number, 64)
	if err != nil {
		return 0, false
	}
	// strtod reports underflow as an error, while ParseFloat silently returns zero or a subnormal value
	mantissa, _, _ := strings.Cut(strings.ToLower(number), "e")
	if math.Abs(val) < minNormalFloat64 && strings.Trim(mantissa, "+-.0") != "" {
		return 0, false
	}
	return val, true
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
// exactly (or nearly exactly, for real values), which matches Postgres's convert_int_from_base_unit and
// convert_real_from_base_unit. Values that aren't positive, as well as values of parameters without a unit, are
// formatted without a unit.
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
		if isInt {
			// Integer values are never displayed in a unit that is smaller than the base unit
			if multiplier <= 1 || converted == math.Trunc(converted) {
				return strconv.FormatInt(int64(converted), 10) + units[i].name, true
			}
		} else if i == 0 || math.Abs(math.RoundToEven(converted)/converted-1) <= 1e-8 {
			// Real values may use smaller units, falling back to the smallest unit when none are integral
			return strconv.FormatFloat(converted, 'g', -1, 64) + units[i].name, true
		}
	}
	return "", false
}
