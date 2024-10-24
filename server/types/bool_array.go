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

// BoolArray is the array variant of Bool.
var BoolArray = CreateArrayTypeFromBaseType(Bool)

//	createArrayTypeWithFuncs(Bool, SerializationID_BoolArray, oid.T__bool, arrayContainerFunctions{
//	SQL: func(ctx *sql.Context, ac arrayContainer, dest []byte, valInterface any) (sqltypes.Value, error) {
//		if valInterface == nil {
//			return sqltypes.NULL, nil
//		}
//		converted, _, err := ac.Convert(valInterface)
//		if err != nil {
//			return sqltypes.Value{}, err
//		}
//		vals := converted.([]any)
//		bb := bytes.Buffer{}
//		bb.WriteRune('{')
//		for i := range vals {
//			if i > 0 {
//				bb.WriteRune(',')
//			}
//			if vals[i] == nil {
//				bb.WriteString("NULL")
//			} else if vals[i].(bool) {
//				bb.WriteRune('t')
//			} else {
//				bb.WriteRune('f')
//			}
//		}
//		bb.WriteRune('}')
//		return sqltypes.MakeTrusted(sqltypes.Text, types.AppendAndSliceBytes(dest, bb.Bytes())), nil
//	},
//})
