// Copyright 2024 Dolthub, Inc.
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
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initToChar registers the functions to the catalog.
func initToChar() {
	framework.RegisterFunction(to_char_timestamp)
}

// to_char_timestamp represents the PostgreSQL function of the same name, taking the same parameters.
// Postgres date formatting: https://www.postgresql.org/docs/8.1/functions-formatting.html
var to_char_timestamp = framework.Function2{
	Name:       "to_char",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Timestamp, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		timestamp := val1.(time.Time)
		format := val2.(string)

		year := timestamp.Format("2006")

		result := ""
		for len(format) > 0 {
			switch {
			case strings.HasPrefix(format, "hh24") || strings.HasPrefix(format, "HH24"):
				result += timestamp.Format("15")
				format = format[4:]
			case strings.HasPrefix(format, "hh12") || strings.HasPrefix(format, "HH12"):
				result += timestamp.Format("03")
				format = format[4:]
			case strings.HasPrefix(format, "hh") || strings.HasPrefix(format, "HH"):
				result += timestamp.Format("03")
				format = format[2:]

			case strings.HasPrefix(format, "mi") || strings.HasPrefix(format, "MI"):
				result += timestamp.Format("04")
				format = format[2:]

			case strings.HasPrefix(format, "ssss") || strings.HasPrefix(format, "SSSS") || strings.HasPrefix(format, "sssss") || strings.HasPrefix(format, "SSSSS"):
				return nil, errors.Errorf("seconds past midnight not supported")

			case strings.HasPrefix(format, "ss") || strings.HasPrefix(format, "SS"):
				result += timestamp.Format("05")
				format = format[2:]

			case strings.HasPrefix(format, "ms") || strings.HasPrefix(format, "MS"):
				result += fmt.Sprintf("%03d", timestamp.Nanosecond()/1_000_000)
				format = format[2:]

			case strings.HasPrefix(format, "us") || strings.HasPrefix(format, "US"):
				result += fmt.Sprintf("%06d", timestamp.Nanosecond()/1_000)
				format = format[2:]

			case strings.HasPrefix(format, "am") || strings.HasPrefix(format, "pm"):
				if timestamp.Hour() < 12 {
					result += "am"
				} else {
					result += "pm"
				}
				format = format[2:]
			case strings.HasPrefix(format, "AM") || strings.HasPrefix(format, "PM"):
				if timestamp.Hour() < 12 {
					result += "AM"
				} else {
					result += "PM"
				}
				format = format[2:]
			case strings.HasPrefix(format, "a.m.") || strings.HasPrefix(format, "p.m."):
				if timestamp.Hour() < 12 {
					result += "a.m."
				} else {
					result += "p.m."
				}
				format = format[4:]
			case strings.HasPrefix(format, "A.M.") || strings.HasPrefix(format, "P.M."):
				if timestamp.Hour() < 12 {
					result += "A.M."
				} else {
					result += "P.M."
				}
				format = format[4:]

			case strings.HasPrefix(format, "y,yyy") || strings.HasPrefix(format, "Y,YYY"):
				result += string(year[0]) + "," + year[1:]
				format = format[5:]
			case strings.HasPrefix(format, "yyyy") || strings.HasPrefix(format, "YYYY"):
				result += year
				format = format[4:]
			case strings.HasPrefix(format, "yyy") || strings.HasPrefix(format, "YYY"):
				result += year[1:]
				format = format[3:]
			case strings.HasPrefix(format, "yy") || strings.HasPrefix(format, "YY"):
				result += year[2:]
				format = format[2:]
			case strings.HasPrefix(format, "y") || strings.HasPrefix(format, "Y"):
				result += year[3:]
				format = format[1:]

			case strings.HasPrefix(format, "iyyy") || strings.HasPrefix(format, "IYYY"):
				return nil, errors.Errorf("ISO year not supported")
			case strings.HasPrefix(format, "iyy") || strings.HasPrefix(format, "IYY"):
				return nil, errors.Errorf("ISO year not supported")
			case strings.HasPrefix(format, "iy") || strings.HasPrefix(format, "IY"):
				return nil, errors.Errorf("ISO year not supported")

			case strings.HasPrefix(format, "bc") || strings.HasPrefix(format, "ad"):
				return nil, errors.Errorf("era indicator not supported")
			case strings.HasPrefix(format, "BC") || strings.HasPrefix(format, "AD"):
				return nil, errors.Errorf("era indicator not supported")
			case strings.HasPrefix(format, "b.c.") || strings.HasPrefix(format, "a.d."):
				return nil, errors.Errorf("era indicator not supported")
			case strings.HasPrefix(format, "B.C.") || strings.HasPrefix(format, "A.D."):
				return nil, errors.Errorf("era indicator not supported")

			case strings.HasPrefix(format, "MONTH"):
				result += strings.ToUpper(timestamp.Format("January"))
				format = format[5:]
			case strings.HasPrefix(format, "Month"):
				result += timestamp.Format("January")
				format = format[5:]
			case strings.HasPrefix(format, "month"):
				result += strings.ToLower(timestamp.Format("January"))
				format = format[5:]

			case strings.HasPrefix(format, "MON"):
				result += strings.ToUpper(timestamp.Format("Jan"))
				format = format[3:]
			case strings.HasPrefix(format, "Mon"):
				result += timestamp.Format("Jan")
				format = format[3:]
			case strings.HasPrefix(format, "mon"):
				result += strings.ToLower(timestamp.Format("Jan"))
				format = format[3:]

			case strings.HasPrefix(format, "mm") || strings.HasPrefix(format, "MM"):
				result += timestamp.Format("01")
				format = format[2:]

			case strings.HasPrefix(format, "DAY"):
				result += strings.ToUpper(timestamp.Format("Monday"))
				format = format[3:]
			case strings.HasPrefix(format, "Day"):
				result += timestamp.Format("Monday")
				format = format[3:]
			case strings.HasPrefix(format, "day"):
				result += strings.ToLower(timestamp.Format("Monday"))
				format = format[3:]

			case strings.HasPrefix(format, "DY"):
				result += strings.ToUpper(timestamp.Format("Mon"))
				format = format[2:]
			case strings.HasPrefix(format, "Dy"):
				result += timestamp.Format("Mon")
				format = format[2:]
			case strings.HasPrefix(format, "dy"):
				result += strings.ToLower(timestamp.Format("Mon"))
				format = format[2:]

			case strings.HasPrefix(format, "ddd") || strings.HasPrefix(format, "DDD"):
				result += timestamp.Format("002")
				format = format[3:]

			case strings.HasPrefix(format, "dd") || strings.HasPrefix(format, "DD"):
				result += timestamp.Format("02")
				format = format[2:]

			case strings.HasPrefix(format, "d") || strings.HasPrefix(format, "D"):
				result += fmt.Sprintf("%d", timestamp.Weekday()+1)
				format = format[1:]

			case strings.HasPrefix(format, "ww") || strings.HasPrefix(format, "WW"):
				return nil, errors.Errorf("week of year not supported")

			case strings.HasPrefix(format, "iw") || strings.HasPrefix(format, "IW"):
				_, week := timestamp.ISOWeek()
				result += fmt.Sprintf("%02d", week)
				format = format[2:]

			case strings.HasPrefix(format, "i") || strings.HasPrefix(format, "I"):
				return nil, errors.Errorf("ISO year not supported")

			case strings.HasPrefix(format, "w") || strings.HasPrefix(format, "W"):
				return nil, errors.Errorf("week of month not supported")

			case strings.HasPrefix(format, "cc") || strings.HasPrefix(format, "CC"):
				return nil, errors.Errorf("century not supported")

			case strings.HasPrefix(format, "j") || strings.HasPrefix(format, "J"):
				return nil, errors.Errorf("julian days not supported")

			case strings.HasPrefix(format, "q") || strings.HasPrefix(format, "Q"):
				switch timestamp.Month() {
				case time.January, time.February, time.March:
					result += "1"
				case time.April, time.May, time.June:
					result += "2"
				case time.July, time.August, time.September:
					result += "3"
				case time.October, time.November, time.December:
					result += "4"
				}
				format = format[1:]

			case strings.HasPrefix(format, "rm") || strings.HasPrefix(format, "RM"):
				return nil, errors.Errorf("roman numeral month not supported")

			case strings.HasPrefix(format, "tz") || strings.HasPrefix(format, "TZ"):
				return nil, errors.Errorf("time-zone name not supported")

			default:
				result += string(format[0])
				format = format[1:]
			}
		}

		return result, nil
	},
}
