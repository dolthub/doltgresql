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
	"bytes"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/lib/pq/oid"
)

// VarCharArray is the array variant of VarChar.
var VarCharArray = createArrayTypeWithFuncs(VarChar, SerializationID_VarCharArray, oid.T__varchar, arrayContainerFunctions{
	SQL: stringArraySQL,
})

// stringArraySQL is the SQL implementation for all string array types.
func stringArraySQL(ctx *sql.Context, ac arrayContainer, dest []byte, valInterface any) (sqltypes.Value, error) {
	if valInterface == nil {
		return sqltypes.NULL, nil
	}
	converted, _, err := ac.Convert(valInterface)
	if err != nil {
		return sqltypes.Value{}, err
	}
	vals := converted.([]any)
	if len(vals) == 0 {
		return sqltypes.MakeTrusted(ac.Type(), types.AppendAndSliceBytes(dest, []byte{'{', '}'})), nil
	}
	bb := bytes.Buffer{}
	bb.WriteRune('{')
	for i := range vals {
		if i > 0 {
			bb.WriteRune(',')
		}
		if vals[i] == nil {
			bb.WriteString("NULL")
			continue
		}
		sqlStringQuote(&bb, vals[i].(string))
	}
	bb.WriteRune('}')
	return sqltypes.MakeTrusted(sqltypes.Text, types.AppendAndSliceBytes(dest, bb.Bytes())), nil
}

// sqlStringQuote returns a string that has been quoted according to the rules defined by Postgres. Not all strings will
// result in quoting, therefore this function takes care of all situations where a string does or does not need to be
// quoted. Writes the result to the buffer.
func sqlStringQuote(bb *bytes.Buffer, val string) {
	containsDoubleQuote := strings.Contains(val, `"`)
	if containsDoubleQuote {
		val = strings.ReplaceAll(val, `"`, `\"`)
	}
	if containsDoubleQuote || strings.Contains(val, `,`) ||
		strings.Contains(val, `{`) || strings.Contains(val, `}`) {
		bb.WriteRune('"')
		bb.WriteString(val)
		bb.WriteRune('"')
	} else {
		bb.WriteString(val)
	}
}
