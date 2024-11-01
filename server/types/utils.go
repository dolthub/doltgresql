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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"
)

// QuoteString will quote the string according to the type given.
// This means that some types will quote, and others will
// not, or they may quote in a special way that is unique to that type.
func QuoteString(typOid oid.Oid, str string) string {
	switch typOid {
	case oid.T_char, oid.T_bpchar, oid.T_name, oid.T_text, oid.T_varchar, oid.T_unknown:
		return `'` + strings.ReplaceAll(str, `'`, `''`) + `'`
	default:
		return str
	}
}

// FromGmsType returns a DoltgresType that is most similar to the given GMS type.
func FromGmsType(typ sql.Type) DoltgresType {
	switch typ.Type() {
	case query.Type_INT8:
		// Special treatment for boolean types when we can detect them
		if typ == types.Boolean {
			return Bool
		}
		return Int32
	case query.Type_INT16, query.Type_INT24, query.Type_INT32, query.Type_YEAR, query.Type_ENUM:
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
