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
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

// millisecondUnits is ordered from largest to smallest, as in PostgreSQL's GUC
// unit conversion table. Fractional inputs round to the next smaller unit first.
var millisecondUnits = []struct {
	name       string
	multiplier float64
}{
	{"d", 86400000}, {"h", 3600000}, {"min", 60000},
	{"s", 1000}, {"ms", 1}, {"us", 0.001},
}

var timeSettingPattern = regexp.MustCompile(`^[\t\n\r\f\v ]*([+-]?(?:0[xX][0-9a-fA-F]+|(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?))[\t\n\r\f\v ]*([a-zA-Z]*)[\t\n\r\f\v ]*$`)

// HasMillisecondUnit reports whether a parameter stores time in milliseconds.
func HasMillisecondUnit(name string) bool {
	p, ok := postgresConfigParameters[strings.ToLower(name)].(*Parameter)
	return ok && p.Unit == "ms"
}

// parseMilliseconds parses a PostgreSQL integer time setting, including units,
// base-prefixed integers, and fractional values rounded using ties-to-even.
func parseMilliseconds(name, input string) (int64, error) {
	invalid := func() (int64, error) {
		return 0, pgerror.Newf(pgcode.InvalidParameterValue, `invalid value for parameter "%s": "%s"`, name, input)
	}
	parts := timeSettingPattern.FindStringSubmatch(input)
	if parts == nil {
		return invalid()
	}
	number := parts[1]
	integer, err := strconv.ParseInt(number, 0, 64)
	value := float64(integer)
	if err != nil {
		// PostgreSQL first tries base-0 integer parsing, falling back to a
		// decimal float for fractional/exponent input or integer overflow.
		if !strings.ContainsAny(number, ".eE") && !errors.Is(err, strconv.ErrRange) {
			return invalid()
		}
		value, err = strconv.ParseFloat(number, 64)
		if err != nil {
			return invalid()
		}
		// strtod reports underflow; Go's ParseFloat silently returns zero or
		// a subnormal value. Reject nonzero inputs below the normal range.
		mantissa := strings.SplitN(strings.ToLower(number), "e", 2)[0]
		if math.Abs(value) < math.Float64frombits(0x0010000000000000) && strings.Trim(mantissa, "+-.0") != "" {
			return invalid()
		}
	}
	if parts[2] != "" {
		found := false
		for i, unit := range millisecondUnits {
			if parts[2] == unit.name {
				value *= unit.multiplier
				if i+1 < len(millisecondUnits) {
					smaller := millisecondUnits[i+1].multiplier
					value = math.RoundToEven(value/smaller) * smaller
				}
				found = true
				break
			}
		}
		if !found {
			return invalid()
		}
	}
	value = math.RoundToEven(value)
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > math.MaxInt32 {
		return invalid()
	}
	return int64(value), nil
}

// formatMilliseconds uses the largest unit that represents the value exactly.
// Zero disables the timeout and is displayed without a unit.
func formatMilliseconds(value int64) string {
	if value == 0 {
		return "0"
	}
	for _, unit := range millisecondUnits {
		if unit.multiplier >= 1 && value%int64(unit.multiplier) == 0 {
			return strconv.FormatInt(value/int64(unit.multiplier), 10) + unit.name
		}
	}
	return strconv.FormatInt(value, 10) + "ms"
}
