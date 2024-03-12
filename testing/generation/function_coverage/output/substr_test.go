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

package output

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func Test_Substr(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "substr",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT substr( '' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  0::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( '0' ,  0::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT substr( '1' ,  0::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  0::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  0::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT substr( '123' ,  0::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT substr( 'value' ,  0::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT substr( '12345' ,  0::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT substr( 'something' ,  0::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT substr( ' something' ,  0::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  0::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  0::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  0::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  0::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT substr( '' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  -1::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( '0' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT substr( '1' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT substr( '123' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT substr( 'value' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT substr( '12345' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT substr( 'something' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT substr( ' something' ,  -1::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT substr( '' ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  1::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( '0' ,  1::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT substr( '1' ,  1::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  1::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  1::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT substr( '123' ,  1::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT substr( 'value' ,  1::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT substr( '12345' ,  1::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT substr( 'something' ,  1::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT substr( ' something' ,  1::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  1::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  1::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  1::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  1::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT substr( '' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  2::int4 ) ;",
					Expected: []sql.Row{{"bc"}},
				},
				{
					Query:    "SELECT substr( '123' ,  2::int4 ) ;",
					Expected: []sql.Row{{"23"}},
				},
				{
					Query:    "SELECT substr( 'value' ,  2::int4 ) ;",
					Expected: []sql.Row{{"alue"}},
				},
				{
					Query:    "SELECT substr( '12345' ,  2::int4 ) ;",
					Expected: []sql.Row{{"2345"}},
				},
				{
					Query:    "SELECT substr( 'something' ,  2::int4 ) ;",
					Expected: []sql.Row{{"omething"}},
				},
				{
					Query:    "SELECT substr( ' something' ,  2::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  2::int4 ) ;",
					Expected: []sql.Row{{"omething "}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  2::int4 ) ;",
					Expected: []sql.Row{{"23456789"}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  2::int4 ) ;",
					Expected: []sql.Row{{" group of words"}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  2::int4 ) ;",
					Expected: []sql.Row{{"234567890123456"}},
				},
				{
					Query:    "SELECT substr( '' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  -2::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( '0' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT substr( '1' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT substr( '123' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT substr( 'value' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT substr( '12345' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT substr( 'something' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT substr( ' something' ,  -2::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT substr( '' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  5::int4 ) ;",
					Expected: []sql.Row{{"e"}},
				},
				{
					Query:    "SELECT substr( '12345' ,  5::int4 ) ;",
					Expected: []sql.Row{{"5"}},
				},
				{
					Query:    "SELECT substr( 'something' ,  5::int4 ) ;",
					Expected: []sql.Row{{"thing"}},
				},
				{
					Query:    "SELECT substr( ' something' ,  5::int4 ) ;",
					Expected: []sql.Row{{"ething"}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  5::int4 ) ;",
					Expected: []sql.Row{{"thing "}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  5::int4 ) ;",
					Expected: []sql.Row{{"56789"}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  5::int4 ) ;",
					Expected: []sql.Row{{"oup of words"}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  5::int4 ) ;",
					Expected: []sql.Row{{"567890123456"}},
				},
				{
					Query:    "SELECT substr( '' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  10::int4 ) ;",
					Expected: []sql.Row{{"g"}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  10::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  10::int4 ) ;",
					Expected: []sql.Row{{"f words"}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  10::int4 ) ;",
					Expected: []sql.Row{{"0123456"}},
				},
				{
					Query:    "SELECT substr( '' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  -10::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( '0' ,  -10::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT substr( '1' ,  -10::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  -10::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  -10::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT substr( '123' ,  -10::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT substr( 'value' ,  -10::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT substr( '12345' ,  -10::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT substr( 'something' ,  -10::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT substr( ' something' ,  -10::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  -10::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  -10::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  -10::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  -10::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT substr( '' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( '0' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT substr( '1' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT substr( '123' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT substr( 'value' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT substr( '12345' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT substr( 'something' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT substr( ' something' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT substr( '' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:       "SELECT substr( '' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '0' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'abc' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'value' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '12345' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' something' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something ' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a group of words' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1234567890123456' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT substr( '' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  0::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  0::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  0::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  0::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  0::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  1::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  1::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  2::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  2::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  2::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  2::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '' ,  -2::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  -2::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  5::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  10::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  10::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  10::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  -10::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  -10::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  100::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  100::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  21050::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  21050::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  100000::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  100000::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  100000::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  100000::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  -1184280::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  2525280::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  2147483647::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  2147483647::int4 ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:       "SELECT substr( '123' ,  0::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something' ,  0::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a group of words' ,  0::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  -1::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '0' ,  -1::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a group of words' ,  -1::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  1::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something' ,  1::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' something' ,  1::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1' ,  2::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '12345' ,  2::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  2::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123' ,  -2::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1234567890123456' ,  -2::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1' ,  5::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '12345' ,  10::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something' ,  -10::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  100::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '0' ,  100::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' something' ,  100::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  100::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '' ,  21050::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123' ,  21050::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '12345' ,  21050::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a' ,  100000::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '' ,  -1184280::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  -1184280::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  2525280::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'abc' ,  2525280::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something ' ,  -2147483648::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123' ,  2147483647::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'value' ,  2147483647::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '12345' ,  2147483647::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something ' ,  2147483647::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1234567890123456' ,  2147483647::int4 ,  -1::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT substr( 'something' ,  -1::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  2::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  -2::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  -2::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  -2::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  5::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  5::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  5::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  5::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '' ,  10::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  10::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  10::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{"g"}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  -10::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '' ,  100::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  100::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  100::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  100::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  100::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  100::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  100::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  21050::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  21050::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  100000::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  100000::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  -1184280::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  -1184280::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  -1184280::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '' ,  2525280::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  2525280::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  2525280::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '' ,  2147483647::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  2147483647::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  2147483647::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  2147483647::int4 ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  0::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT substr( '12345' ,  0::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  0::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT substr( '1' ,  -1::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  -1::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  1::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( 'value' ,  2::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{"al"}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  2::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{"om"}},
				},
				{
					Query:    "SELECT substr( 'value' ,  -2::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  -2::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  -2::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  5::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  5::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{"et"}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  5::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{"56"}},
				},
				{
					Query:    "SELECT substr( ' ' ,  10::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  10::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  10::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  10::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{"g"}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  10::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  100::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  100::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  100::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  21050::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  100000::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  100000::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  100000::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  100000::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  -1184280::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  -1184280::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  -1184280::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  2525280::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  2525280::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:       "SELECT substr( 'a group of words' ,  -2147483648::int4 ,  2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT substr( 'something' ,  2147483647::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  2147483647::int4 ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:       "SELECT substr( '' ,  0::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'abc' ,  0::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a group of words' ,  0::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  1::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1' ,  2::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' something' ,  2::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something ' ,  2::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '12345' ,  -2::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '' ,  5::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  5::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'value' ,  5::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  -10::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '0' ,  -10::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  100::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '12345' ,  100::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a group of words' ,  21050::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  100000::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123' ,  100000::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something ' ,  100000::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1234567890123456' ,  100000::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'abc' ,  -1184280::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '12345' ,  -1184280::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  -1184280::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '0' ,  2525280::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'value' ,  2525280::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a' ,  -2147483648::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '12345' ,  -2147483648::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something' ,  -2147483648::int4 ,  -2::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT substr( '' ,  0::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  0::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{"some"}},
				},
				{
					Query:    "SELECT substr( ' something' ,  0::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{" som"}},
				},
				{
					Query:    "SELECT substr( '1' ,  -1::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  1::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT substr( 'value' ,  1::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  2::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  -2::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{"so"}},
				},
				{
					Query:    "SELECT substr( ' ' ,  5::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  5::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{"thing"}},
				},
				{
					Query:    "SELECT substr( ' ' ,  -10::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  -10::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  -10::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  -10::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  -10::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '' ,  100::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  100::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '' ,  21050::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  21050::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  21050::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  21050::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  100000::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  -1184280::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  2525280::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:       "SELECT substr( '' ,  -2147483648::int4 ,  5::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1' ,  -2147483648::int4 ,  5::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a' ,  -2147483648::int4 ,  5::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something' ,  -2147483648::int4 ,  5::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' something' ,  -2147483648::int4 ,  5::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT substr( '' ,  2147483647::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  2147483647::int4 ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  0::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  0::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  -1::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT substr( 'value' ,  -1::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT substr( '0' ,  1::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  2::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{"23456789"}},
				},
				{
					Query:    "SELECT substr( ' ' ,  -2::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  -2::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT substr( '' ,  5::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  5::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  5::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{"56789"}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  5::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{"5678901234"}},
				},
				{
					Query:    "SELECT substr( '' ,  10::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  -10::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  -10::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  -10::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  -10::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  100::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  100::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  21050::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  21050::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  21050::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  100000::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  100000::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  100000::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  100000::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  -1184280::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  -1184280::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  -1184280::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '' ,  2525280::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  2525280::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  2525280::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:       "SELECT substr( ' ' ,  -2147483648::int4 ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1' ,  -2147483648::int4 ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123' ,  -2147483648::int4 ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' something' ,  -2147483648::int4 ,  10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT substr( ' ' ,  2147483647::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  2147483647::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  2147483647::int4 ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:       "SELECT substr( '1' ,  0::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something ' ,  0::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  0::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' something' ,  -1::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something ' ,  1::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '' ,  2::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '0' ,  2::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'value' ,  2::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  2::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'abc' ,  -2::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  -2::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '' ,  5::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123' ,  5::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '' ,  10::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1' ,  10::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something ' ,  10::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a group of words' ,  10::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '0' ,  -10::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'value' ,  -10::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '' ,  100::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'value' ,  21050::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '' ,  100000::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  100000::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '0' ,  100000::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something' ,  100000::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '' ,  -1184280::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a group of words' ,  -1184280::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '0' ,  2525280::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a' ,  2525280::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'abc' ,  2525280::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something' ,  2525280::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  2525280::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  -2147483648::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1' ,  2147483647::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123' ,  2147483647::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something' ,  2147483647::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' something' ,  2147483647::int4 ,  -10::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT substr( '0' ,  0::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  0::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT substr( '123' ,  0::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT substr( 'value' ,  0::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT substr( ' something' ,  0::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT substr( '0' ,  -1::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  -1::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT substr( '0' ,  1::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT substr( '123' ,  1::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT substr( 'something' ,  1::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT substr( '12345' ,  2::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{"2345"}},
				},
				{
					Query:    "SELECT substr( '' ,  -2::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  -2::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( '123' ,  -2::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT substr( '' ,  5::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  5::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  5::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  5::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  5::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{"567890123456"}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  10::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( 'a' ,  -10::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT substr( '1' ,  100::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  100::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  100::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  21050::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  21050::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  21050::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  100000::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  100000::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  -1184280::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  -1184280::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  -1184280::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  2525280::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  2525280::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:       "SELECT substr( '' ,  -2147483648::int4 ,  100::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123' ,  -2147483648::int4 ,  100::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something' ,  -2147483648::int4 ,  100::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' something' ,  -2147483648::int4 ,  100::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT substr( '1' ,  2147483647::int4 ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '' ,  -1::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  -1::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT substr( ' ' ,  1::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( '0' ,  1::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT substr( '12345' ,  2::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{"2345"}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  2::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{"234567890123456"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  -2::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  -2::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT substr( '12345' ,  -2::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  -2::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT substr( '123' ,  5::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  5::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{"56789"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  10::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  100::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something' ,  100::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  100::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  21050::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  21050::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  21050::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  100000::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  -1184280::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  2525280::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  2147483647::int4 ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  -1::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( '1' ,  -1::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  -1::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT substr( 'something' ,  -1::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  1::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT substr( '' ,  2::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  2::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  2::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{"23"}},
				},
				{
					Query:    "SELECT substr( '12345' ,  2::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{"2345"}},
				},
				{
					Query:    "SELECT substr( 'value' ,  -2::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  -2::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  -2::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT substr( ' ' ,  5::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  5::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{"oup of words"}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  10::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{"f words"}},
				},
				{
					Query:    "SELECT substr( '' ,  -10::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  -10::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  -10::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT substr( '123' ,  -10::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  100::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  100::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  100::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  100::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  21050::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  21050::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  100000::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  100000::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  100000::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  -1184280::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1234567890123456' ,  -1184280::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '' ,  2525280::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a' ,  2525280::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  2525280::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  2525280::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:       "SELECT substr( '' ,  -2147483648::int4 ,  100000::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  -2147483648::int4 ,  100000::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123' ,  -2147483648::int4 ,  100000::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '12345' ,  -2147483648::int4 ,  100000::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  -2147483648::int4 ,  100000::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT substr( '12345' ,  2147483647::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  2147483647::int4 ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:       "SELECT substr( 'a' ,  0::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  -1::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a' ,  -1::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1234567890123456' ,  -1::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '12345' ,  1::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'value' ,  2::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a group of words' ,  2::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  -2::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '0' ,  -2::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1' ,  -2::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'abc' ,  5::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' something' ,  5::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1234567890123456' ,  5::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '' ,  10::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '0' ,  10::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '0' ,  -10::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123' ,  -10::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'value' ,  -10::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  -10::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1234567890123456' ,  100::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '0' ,  21050::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1234567890123456' ,  21050::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something ' ,  100000::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a group of words' ,  100000::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'abc' ,  -1184280::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something' ,  -1184280::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123' ,  2525280::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'value' ,  -2147483648::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something ' ,  -2147483648::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a group of words' ,  -2147483648::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1234567890123456' ,  -2147483648::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something' ,  2147483647::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  2147483647::int4 ,  -1184280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT substr( '12345' ,  0::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT substr( 'value' ,  -1::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  -1::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT substr( ' ' ,  1::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( 'value' ,  1::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  1::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT substr( '0' ,  2::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  2::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT substr( 'value' ,  -2::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT substr( 'something' ,  -2::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  5::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  5::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"thing "}},
				},
				{
					Query:    "SELECT substr( '1' ,  10::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'abc' ,  10::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  -10::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT substr( '123' ,  -10::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  -10::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  21050::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  21050::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  100000::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' ' ,  2525280::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123' ,  2525280::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  2525280::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:       "SELECT substr( '0' ,  -2147483648::int4 ,  2525280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' something' ,  -2147483648::int4 ,  2525280::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT substr( '123456789' ,  2147483647::int4 ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:       "SELECT substr( '' ,  0::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a' ,  0::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'abc' ,  0::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something ' ,  0::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a group of words' ,  0::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '' ,  -1::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'abc' ,  -1::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '12345' ,  -1::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'abc' ,  1::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123' ,  1::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something ' ,  1::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  2::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1' ,  -2::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  5::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a' ,  10::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a group of words' ,  -10::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1' ,  100::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123' ,  100::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something' ,  100::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' something' ,  100::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '' ,  100000::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'abc' ,  100000::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  100000::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '1234567890123456' ,  100000::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  -1184280::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'abc' ,  -1184280::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123' ,  -1184280::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something' ,  -1184280::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  -1184280::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '12345' ,  2525280::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'something' ,  2525280::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' something' ,  2525280::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'value' ,  -2147483648::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a group of words' ,  -2147483648::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' ' ,  2147483647::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( 'a' ,  2147483647::int4 ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT substr( '' ,  0::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  0::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  0::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT substr( '' ,  -1::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '' ,  1::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  1::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT substr( '1' ,  1::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT substr( ' ' ,  2::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '123456789' ,  2::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"23456789"}},
				},
				{
					Query:    "SELECT substr( '0' ,  5::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  5::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '1' ,  10::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  10::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"f words"}},
				},
				{
					Query:    "SELECT substr( '12345' ,  -10::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT substr( 'something ' ,  -10::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  -10::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT substr( 'a' ,  100::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  100::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  100::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  100::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '0' ,  21050::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  100000::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( '12345' ,  -1184280::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT substr( 'something' ,  -1184280::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT substr( '123' ,  2525280::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:       "SELECT substr( 'abc' ,  -2147483648::int4 ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( ' something' ,  -2147483648::int4 ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT substr( '123456789' ,  -2147483648::int4 ,  2147483647::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT substr( 'a' ,  2147483647::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'value' ,  2147483647::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( ' something' ,  2147483647::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT substr( 'a group of words' ,  2147483647::int4 ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
			},
		},
	})
}
