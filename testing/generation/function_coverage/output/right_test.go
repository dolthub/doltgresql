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

func Test_Right(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "right",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT right( '' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' ' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '0' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '1' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'a' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'abc' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '123' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'value' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '12345' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'something' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' something' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'something ' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '123456789' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'a group of words' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '1234567890123456' ,  0::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' ' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '0' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '1' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'a' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'abc' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"bc"}},
				},
				{
					Query:    "SELECT right( '123' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"23"}},
				},
				{
					Query:    "SELECT right( 'value' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"alue"}},
				},
				{
					Query:    "SELECT right( '12345' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"2345"}},
				},
				{
					Query:    "SELECT right( 'something' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"omething"}},
				},
				{
					Query:    "SELECT right( ' something' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT right( 'something ' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"omething "}},
				},
				{
					Query:    "SELECT right( '123456789' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"23456789"}},
				},
				{
					Query:    "SELECT right( 'a group of words' ,  -1::int4 ) ;",
					Expected: []sql.Row{{" group of words"}},
				},
				{
					Query:    "SELECT right( '1234567890123456' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"234567890123456"}},
				},
				{
					Query:    "SELECT right( '' ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' ' ,  1::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT right( '0' ,  1::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT right( '1' ,  1::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT right( 'a' ,  1::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT right( 'abc' ,  1::int4 ) ;",
					Expected: []sql.Row{{"c"}},
				},
				{
					Query:    "SELECT right( '123' ,  1::int4 ) ;",
					Expected: []sql.Row{{"3"}},
				},
				{
					Query:    "SELECT right( 'value' ,  1::int4 ) ;",
					Expected: []sql.Row{{"e"}},
				},
				{
					Query:    "SELECT right( '12345' ,  1::int4 ) ;",
					Expected: []sql.Row{{"5"}},
				},
				{
					Query:    "SELECT right( 'something' ,  1::int4 ) ;",
					Expected: []sql.Row{{"g"}},
				},
				{
					Query:    "SELECT right( ' something' ,  1::int4 ) ;",
					Expected: []sql.Row{{"g"}},
				},
				{
					Query:    "SELECT right( 'something ' ,  1::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT right( '123456789' ,  1::int4 ) ;",
					Expected: []sql.Row{{"9"}},
				},
				{
					Query:    "SELECT right( 'a group of words' ,  1::int4 ) ;",
					Expected: []sql.Row{{"s"}},
				},
				{
					Query:    "SELECT right( '1234567890123456' ,  1::int4 ) ;",
					Expected: []sql.Row{{"6"}},
				},
				{
					Query:    "SELECT right( '' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' ' ,  2::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT right( '0' ,  2::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT right( '1' ,  2::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT right( 'a' ,  2::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT right( 'abc' ,  2::int4 ) ;",
					Expected: []sql.Row{{"bc"}},
				},
				{
					Query:    "SELECT right( '123' ,  2::int4 ) ;",
					Expected: []sql.Row{{"23"}},
				},
				{
					Query:    "SELECT right( 'value' ,  2::int4 ) ;",
					Expected: []sql.Row{{"ue"}},
				},
				{
					Query:    "SELECT right( '12345' ,  2::int4 ) ;",
					Expected: []sql.Row{{"45"}},
				},
				{
					Query:    "SELECT right( 'something' ,  2::int4 ) ;",
					Expected: []sql.Row{{"ng"}},
				},
				{
					Query:    "SELECT right( ' something' ,  2::int4 ) ;",
					Expected: []sql.Row{{"ng"}},
				},
				{
					Query:    "SELECT right( 'something ' ,  2::int4 ) ;",
					Expected: []sql.Row{{"g "}},
				},
				{
					Query:    "SELECT right( '123456789' ,  2::int4 ) ;",
					Expected: []sql.Row{{"89"}},
				},
				{
					Query:    "SELECT right( 'a group of words' ,  2::int4 ) ;",
					Expected: []sql.Row{{"ds"}},
				},
				{
					Query:    "SELECT right( '1234567890123456' ,  2::int4 ) ;",
					Expected: []sql.Row{{"56"}},
				},
				{
					Query:    "SELECT right( '' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' ' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '0' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '1' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'a' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'abc' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"c"}},
				},
				{
					Query:    "SELECT right( '123' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"3"}},
				},
				{
					Query:    "SELECT right( 'value' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"lue"}},
				},
				{
					Query:    "SELECT right( '12345' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"345"}},
				},
				{
					Query:    "SELECT right( 'something' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"mething"}},
				},
				{
					Query:    "SELECT right( ' something' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"omething"}},
				},
				{
					Query:    "SELECT right( 'something ' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"mething "}},
				},
				{
					Query:    "SELECT right( '123456789' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"3456789"}},
				},
				{
					Query:    "SELECT right( 'a group of words' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"group of words"}},
				},
				{
					Query:    "SELECT right( '1234567890123456' ,  -2::int4 ) ;",
					Expected: []sql.Row{{"34567890123456"}},
				},
				{
					Query:    "SELECT right( '' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' ' ,  5::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT right( '0' ,  5::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT right( '1' ,  5::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT right( 'a' ,  5::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT right( 'abc' ,  5::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT right( '123' ,  5::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT right( 'value' ,  5::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT right( '12345' ,  5::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT right( 'something' ,  5::int4 ) ;",
					Expected: []sql.Row{{"thing"}},
				},
				{
					Query:    "SELECT right( ' something' ,  5::int4 ) ;",
					Expected: []sql.Row{{"thing"}},
				},
				{
					Query:    "SELECT right( 'something ' ,  5::int4 ) ;",
					Expected: []sql.Row{{"hing "}},
				},
				{
					Query:    "SELECT right( '123456789' ,  5::int4 ) ;",
					Expected: []sql.Row{{"56789"}},
				},
				{
					Query:    "SELECT right( 'a group of words' ,  5::int4 ) ;",
					Expected: []sql.Row{{"words"}},
				},
				{
					Query:    "SELECT right( '1234567890123456' ,  5::int4 ) ;",
					Expected: []sql.Row{{"23456"}},
				},
				{
					Query:    "SELECT right( '' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' ' ,  10::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT right( '0' ,  10::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT right( '1' ,  10::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT right( 'a' ,  10::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT right( 'abc' ,  10::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT right( '123' ,  10::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT right( 'value' ,  10::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT right( '12345' ,  10::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT right( 'something' ,  10::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT right( ' something' ,  10::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT right( 'something ' ,  10::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT right( '123456789' ,  10::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT right( 'a group of words' ,  10::int4 ) ;",
					Expected: []sql.Row{{"p of words"}},
				},
				{
					Query:    "SELECT right( '1234567890123456' ,  10::int4 ) ;",
					Expected: []sql.Row{{"7890123456"}},
				},
				{
					Query:    "SELECT right( '' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' ' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '0' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '1' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'a' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'abc' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '123' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'value' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '12345' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'something' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' something' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'something ' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '123456789' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'a group of words' ,  -10::int4 ) ;",
					Expected: []sql.Row{{" words"}},
				},
				{
					Query:    "SELECT right( '1234567890123456' ,  -10::int4 ) ;",
					Expected: []sql.Row{{"123456"}},
				},
				{
					Query:    "SELECT right( '' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' ' ,  100::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT right( '0' ,  100::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT right( '1' ,  100::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT right( 'a' ,  100::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT right( 'abc' ,  100::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT right( '123' ,  100::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT right( 'value' ,  100::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT right( '12345' ,  100::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT right( 'something' ,  100::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT right( ' something' ,  100::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT right( 'something ' ,  100::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT right( '123456789' ,  100::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT right( 'a group of words' ,  100::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT right( '1234567890123456' ,  100::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT right( '' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' ' ,  21050::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT right( '0' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT right( '1' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT right( 'a' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT right( 'abc' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT right( '123' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT right( 'value' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT right( '12345' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT right( 'something' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT right( ' something' ,  21050::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT right( 'something ' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT right( '123456789' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT right( 'a group of words' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT right( '1234567890123456' ,  21050::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT right( '' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' ' ,  100000::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT right( '0' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT right( '1' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT right( 'a' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT right( 'abc' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT right( '123' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT right( 'value' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT right( '12345' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT right( 'something' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT right( ' something' ,  100000::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT right( 'something ' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT right( '123456789' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT right( 'a group of words' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT right( '1234567890123456' ,  100000::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT right( '' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' ' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '0' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '1' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'a' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'abc' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '123' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'value' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '12345' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'something' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' something' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'something ' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '123456789' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( 'a group of words' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '1234567890123456' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( '' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT right( '0' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT right( '1' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT right( 'a' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT right( 'abc' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT right( '123' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT right( 'value' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT right( '12345' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT right( 'something' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT right( ' something' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT right( 'something ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT right( '123456789' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT right( 'a group of words' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT right( '1234567890123456' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:       "SELECT right( '' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT right( ' ' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT right( '0' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT right( '1' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT right( 'a' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT right( 'abc' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT right( '123' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT right( 'value' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT right( '12345' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT right( 'something' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT right( ' something' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT right( 'something ' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT right( '123456789' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT right( 'a group of words' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT right( '1234567890123456' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT right( '' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT right( ' ' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT right( '0' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT right( '1' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT right( 'a' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT right( 'abc' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT right( '123' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT right( 'value' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT right( '12345' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT right( 'something' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT right( ' something' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT right( 'something ' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT right( '123456789' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT right( 'a group of words' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT right( '1234567890123456' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
			},
		},
	})
}
