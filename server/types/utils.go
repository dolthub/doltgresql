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

package types

import (
	"strings"
	"unicode/utf8"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/vt/proto/query"
)

// QuoteString will quote the string according to the type given. This means that some types will quote, and others will
// not, or they may quote in a special way that is unique to that type.
func QuoteString(baseID DoltgresTypeBaseID, str string) string {
	switch baseID {
	case DoltgresTypeBaseID_Char, DoltgresTypeBaseID_Name, DoltgresTypeBaseID_Text, DoltgresTypeBaseID_VarChar, DoltgresTypeBaseID_Unknown:
		return `'` + strings.ReplaceAll(str, `'`, `''`) + `'`
	default:
		return str
	}
}

// truncateString returns a string that has been truncated to the given length. Uses the rune count rather than the
// byte count. Returns the input string if it's smaller than the length. Also returns the rune count of the string.
func truncateString(val string, runeLimit uint32) (string, uint32) {
	runeLength := uint32(utf8.RuneCountInString(val))
	if runeLength > runeLimit {
		// TODO: figure out if there's a faster way to truncate based on rune count
		startString := val
		for i := uint32(0); i < runeLimit; i++ {
			_, size := utf8.DecodeRuneInString(val)
			val = val[size:]
		}
		return startString[:len(startString)-len(val)], runeLength
	}
	return val, runeLength
}

// FromGmsType returns a DoltgresType that is most similar to the given GMS type.
func FromGmsType(typ sql.Type) DoltgresType {
	switch typ.Type() {
	case query.Type_INT8, query.Type_INT16, query.Type_INT24, query.Type_INT32, query.Type_YEAR, query.Type_ENUM:
		return Int32
	case query.Type_INT64, query.Type_SET, query.Type_BIT, query.Type_UINT8, query.Type_UINT16, query.Type_UINT24, query.Type_UINT32:
		return Int64
	case query.Type_UINT64:
		return Numeric
	case query.Type_FLOAT32:
		return Float32
	case query.Type_FLOAT64:
		return Float64
	case query.Type_DECIMAL:
		return Numeric
	case query.Type_DATE, query.Type_DATETIME, query.Type_TIMESTAMP:
		return Timestamp
	case query.Type_TIME:
		return Text
	case query.Type_CHAR, query.Type_VARCHAR, query.Type_TEXT, query.Type_BINARY, query.Type_VARBINARY, query.Type_BLOB:
		return Text
	case query.Type_JSON:
		return Json
	case query.Type_NULL_TYPE:
		return Unknown
	case query.Type_GEOMETRY:
		return Unknown
	default:
		return Unknown
	}
}
