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

func Test_Left(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "left",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT left( '' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' ' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '0' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '1' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'a' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'abc' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '123' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'value' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '12345' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'something' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' something' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'something ' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '123456789' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'a group of words' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '1234567890123456' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' ' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '0' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '1' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'a' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'abc' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"ab"}},
				},
				{
					Query:    "SELECT left( '123' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"12"}},
				},
				{
					Query:    "SELECT left( 'value' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"valu"}},
				},
				{
					Query:    "SELECT left( '12345' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"1234"}},
				},
				{
					Query:    "SELECT left( 'something' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"somethin"}},
				},
				{
					Query:    "SELECT left( ' something' ,  -1::int4 ) ;",
					Expected: []sql.Row{{" somethin"}},
				},
				{
					Query:    "SELECT left( 'something ' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT left( '123456789' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"12345678"}},
				},
				{
					Query:    "SELECT left( 'a group of words' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"a group of word"}},
				},
				{
					Query:    "SELECT left( '1234567890123456' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"123456789012345"}},
				},
				{
					Query:    "SELECT left( '' ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' ' ,  1::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT left( '0' ,  1::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT left( '1' ,  1::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT left( 'a' ,  1::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT left( 'abc' ,  1::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT left( '123' ,  1::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT left( 'value' ,  1::int4 ) ;",
					Expected: []sql.Row{{"v"}},
				},
				{
					Query:    "SELECT left( '12345' ,  1::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT left( 'something' ,  1::int4 ) ;",
					Expected: []sql.Row{{"s"}},
				},
				{
					Query:    "SELECT left( ' something' ,  1::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT left( 'something ' ,  1::int4 ) ;",
					Expected: []sql.Row{{"s"}},
				},
				{
					Query:    "SELECT left( '123456789' ,  1::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT left( 'a group of words' ,  1::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT left( '1234567890123456' ,  1::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT left( '' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' ' ,  2::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT left( '0' ,  2::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT left( '1' ,  2::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT left( 'a' ,  2::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT left( 'abc' ,  2::int4 ) ;",
					Expected: []sql.Row{{"ab"}},
				},
				{
					Query:    "SELECT left( '123' ,  2::int4 ) ;",
					Expected: []sql.Row{{"12"}},
				},
				{
					Query:    "SELECT left( 'value' ,  2::int4 ) ;",
					Expected: []sql.Row{{"va"}},
				},
				{
					Query:    "SELECT left( '12345' ,  2::int4 ) ;",
					Expected: []sql.Row{{"12"}},
				},
				{
					Query:    "SELECT left( 'something' ,  2::int4 ) ;",
					Expected: []sql.Row{{"so"}},
				},
				{
					Query:    "SELECT left( ' something' ,  2::int4 ) ;",
					Expected: []sql.Row{{" s"}},
				},
				{
					Query:    "SELECT left( 'something ' ,  2::int4 ) ;",
					Expected: []sql.Row{{"so"}},
				},
				{
					Query:    "SELECT left( '123456789' ,  2::int4 ) ;",
					Expected: []sql.Row{{"12"}},
				},
				{
					Query:    "SELECT left( 'a group of words' ,  2::int4 ) ;",
					Expected: []sql.Row{{"a "}},
				},
				{
					Query:    "SELECT left( '1234567890123456' ,  2::int4 ) ;",
					Expected: []sql.Row{{"12"}},
				},
				{
					Query:    "SELECT left( '' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' ' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '0' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '1' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'a' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'abc' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT left( '123' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT left( 'value' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"val"}},
				},
				{
					Query:    "SELECT left( '12345' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT left( 'something' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"somethi"}},
				},
				{
					Query:    "SELECT left( ' something' ,  -2::int4 ) ;",
					Expected: []sql.Row{{" somethi"}},
				},
				{
					Query:    "SELECT left( 'something ' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"somethin"}},
				},
				{
					Query:    "SELECT left( '123456789' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"1234567"}},
				},
				{
					Query:    "SELECT left( 'a group of words' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"a group of wor"}},
				},
				{
					Query:    "SELECT left( '1234567890123456' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"12345678901234"}},
				},
				{
					Query:    "SELECT left( '' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' ' ,  5::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT left( '0' ,  5::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT left( '1' ,  5::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT left( 'a' ,  5::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT left( 'abc' ,  5::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT left( '123' ,  5::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT left( 'value' ,  5::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT left( '12345' ,  5::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT left( 'something' ,  5::int4 ) ;",
					Expected: []sql.Row{{"somet"}},
				},
				{
					Query:    "SELECT left( ' something' ,  5::int4 ) ;",
					Expected: []sql.Row{{" some"}},
				},
				{
					Query:    "SELECT left( 'something ' ,  5::int4 ) ;",
					Expected: []sql.Row{{"somet"}},
				},
				{
					Query:    "SELECT left( '123456789' ,  5::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT left( 'a group of words' ,  5::int4 ) ;",
					Expected: []sql.Row{{"a gro"}},
				},
				{
					Query:    "SELECT left( '1234567890123456' ,  5::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT left( '' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' ' ,  10::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT left( '0' ,  10::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT left( '1' ,  10::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT left( 'a' ,  10::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT left( 'abc' ,  10::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT left( '123' ,  10::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT left( 'value' ,  10::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT left( '12345' ,  10::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT left( 'something' ,  10::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT left( ' something' ,  10::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT left( 'something ' ,  10::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT left( '123456789' ,  10::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT left( 'a group of words' ,  10::int4 ) ;",
					Expected: []sql.Row{{"a group of"}},
				},
				{
					Query:    "SELECT left( '1234567890123456' ,  10::int4 ) ;",
					Expected: []sql.Row{{"1234567890"}},
				},
				{
					Query:    "SELECT left( '' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' ' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '0' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '1' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'a' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'abc' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '123' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'value' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '12345' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'something' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' something' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'something ' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '123456789' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'a group of words' ,  -10::int4 ) ;",
					Expected: []sql.Row{{"a grou"}},
				},
				{
					Query:    "SELECT left( '1234567890123456' ,  -10::int4 ) ;",
					Expected: []sql.Row{{"123456"}},
				},
				{
					Query:    "SELECT left( '' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' ' ,  100::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT left( '0' ,  100::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT left( '1' ,  100::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT left( 'a' ,  100::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT left( 'abc' ,  100::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT left( '123' ,  100::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT left( 'value' ,  100::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT left( '12345' ,  100::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT left( 'something' ,  100::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT left( ' something' ,  100::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT left( 'something ' ,  100::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT left( '123456789' ,  100::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT left( 'a group of words' ,  100::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT left( '1234567890123456' ,  100::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT left( '' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' ' ,  21050::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT left( '0' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT left( '1' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT left( 'a' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT left( 'abc' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT left( '123' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT left( 'value' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT left( '12345' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT left( 'something' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT left( ' something' ,  21050::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT left( 'something ' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT left( '123456789' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT left( 'a group of words' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT left( '1234567890123456' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT left( '' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' ' ,  100000::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT left( '0' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT left( '1' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT left( 'a' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT left( 'abc' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT left( '123' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT left( 'value' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT left( '12345' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT left( 'something' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT left( ' something' ,  100000::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT left( 'something ' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT left( '123456789' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT left( 'a group of words' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT left( '1234567890123456' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT left( '' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' ' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '0' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '1' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'a' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'abc' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '123' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'value' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '12345' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'something' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' something' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'something ' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '123456789' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( 'a group of words' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '1234567890123456' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( '' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT left( '0' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT left( '1' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT left( 'a' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT left( 'abc' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT left( '123' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT left( 'value' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT left( '12345' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT left( 'something' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT left( ' something' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT left( 'something ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT left( '123456789' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT left( 'a group of words' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT left( '1234567890123456' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:       "SELECT left( '' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT left( ' ' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT left( '0' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT left( '1' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT left( 'a' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT left( 'abc' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT left( '123' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT left( 'value' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT left( '12345' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT left( 'something' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT left( ' something' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT left( 'something ' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT left( '123456789' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT left( 'a group of words' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT left( '1234567890123456' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT left( '' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT left( ' ' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT left( '0' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT left( '1' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT left( 'a' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT left( 'abc' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT left( '123' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT left( 'value' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT left( '12345' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT left( 'something' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT left( ' something' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT left( 'something ' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT left( '123456789' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT left( 'a group of words' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT left( '1234567890123456' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
			},
		},
	})
}
