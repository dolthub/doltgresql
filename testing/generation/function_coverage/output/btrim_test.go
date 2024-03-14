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

func Test_Btrim(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "btrim",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT btrim( '' ,  '' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' ' ,  '' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT btrim( '0' ,  '' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT btrim( '1' ,  '' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT btrim( 'a' ,  '' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT btrim( 'abc' ,  '' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT btrim( '123' ,  '' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT btrim( 'value' ,  '' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT btrim( '12345' ,  '' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT btrim( 'something' ,  '' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT btrim( ' something' ,  '' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT btrim( 'something ' ,  '' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT btrim( '123456789' ,  '' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT btrim( 'a group of words' ,  '' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT btrim( '1234567890123456' ,  '' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT btrim( '' ,  ' ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' ' ,  ' ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( '0' ,  ' ' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT btrim( '1' ,  ' ' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT btrim( 'a' ,  ' ' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT btrim( 'abc' ,  ' ' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT btrim( '123' ,  ' ' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT btrim( 'value' ,  ' ' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT btrim( '12345' ,  ' ' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT btrim( 'something' ,  ' ' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT btrim( ' something' ,  ' ' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT btrim( 'something ' ,  ' ' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT btrim( '123456789' ,  ' ' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT btrim( 'a group of words' ,  ' ' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT btrim( '1234567890123456' ,  ' ' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT btrim( '' ,  '0' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' ' ,  '0' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT btrim( '0' ,  '0' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( '1' ,  '0' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT btrim( 'a' ,  '0' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT btrim( 'abc' ,  '0' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT btrim( '123' ,  '0' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT btrim( 'value' ,  '0' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT btrim( '12345' ,  '0' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT btrim( 'something' ,  '0' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT btrim( ' something' ,  '0' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT btrim( 'something ' ,  '0' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT btrim( '123456789' ,  '0' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT btrim( 'a group of words' ,  '0' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT btrim( '1234567890123456' ,  '0' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT btrim( '' ,  '1' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' ' ,  '1' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT btrim( '0' ,  '1' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT btrim( '1' ,  '1' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'a' ,  '1' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT btrim( 'abc' ,  '1' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT btrim( '123' ,  '1' ) ;",
					Expected: []sql.Row{{"23"}},
				},
				{
					Query:    "SELECT btrim( 'value' ,  '1' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT btrim( '12345' ,  '1' ) ;",
					Expected: []sql.Row{{"2345"}},
				},
				{
					Query:    "SELECT btrim( 'something' ,  '1' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT btrim( ' something' ,  '1' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT btrim( 'something ' ,  '1' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT btrim( '123456789' ,  '1' ) ;",
					Expected: []sql.Row{{"23456789"}},
				},
				{
					Query:    "SELECT btrim( 'a group of words' ,  '1' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT btrim( '1234567890123456' ,  '1' ) ;",
					Expected: []sql.Row{{"234567890123456"}},
				},
				{
					Query:    "SELECT btrim( '' ,  'a' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' ' ,  'a' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT btrim( '0' ,  'a' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT btrim( '1' ,  'a' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT btrim( 'a' ,  'a' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'abc' ,  'a' ) ;",
					Expected: []sql.Row{{"bc"}},
				},
				{
					Query:    "SELECT btrim( '123' ,  'a' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT btrim( 'value' ,  'a' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT btrim( '12345' ,  'a' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT btrim( 'something' ,  'a' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT btrim( ' something' ,  'a' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT btrim( 'something ' ,  'a' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT btrim( '123456789' ,  'a' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT btrim( 'a group of words' ,  'a' ) ;",
					Expected: []sql.Row{{" group of words"}},
				},
				{
					Query:    "SELECT btrim( '1234567890123456' ,  'a' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT btrim( '' ,  'abc' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' ' ,  'abc' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT btrim( '0' ,  'abc' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT btrim( '1' ,  'abc' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT btrim( 'a' ,  'abc' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'abc' ,  'abc' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( '123' ,  'abc' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT btrim( 'value' ,  'abc' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT btrim( '12345' ,  'abc' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT btrim( 'something' ,  'abc' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT btrim( ' something' ,  'abc' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT btrim( 'something ' ,  'abc' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT btrim( '123456789' ,  'abc' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT btrim( 'a group of words' ,  'abc' ) ;",
					Expected: []sql.Row{{" group of words"}},
				},
				{
					Query:    "SELECT btrim( '1234567890123456' ,  'abc' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT btrim( '' ,  '123' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' ' ,  '123' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT btrim( '0' ,  '123' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT btrim( '1' ,  '123' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'a' ,  '123' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT btrim( 'abc' ,  '123' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT btrim( '123' ,  '123' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'value' ,  '123' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT btrim( '12345' ,  '123' ) ;",
					Expected: []sql.Row{{"45"}},
				},
				{
					Query:    "SELECT btrim( 'something' ,  '123' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT btrim( ' something' ,  '123' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT btrim( 'something ' ,  '123' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT btrim( '123456789' ,  '123' ) ;",
					Expected: []sql.Row{{"456789"}},
				},
				{
					Query:    "SELECT btrim( 'a group of words' ,  '123' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT btrim( '1234567890123456' ,  '123' ) ;",
					Expected: []sql.Row{{"4567890123456"}},
				},
				{
					Query:    "SELECT btrim( '' ,  'value' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' ' ,  'value' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT btrim( '0' ,  'value' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT btrim( '1' ,  'value' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT btrim( 'a' ,  'value' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'abc' ,  'value' ) ;",
					Expected: []sql.Row{{"bc"}},
				},
				{
					Query:    "SELECT btrim( '123' ,  'value' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT btrim( 'value' ,  'value' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( '12345' ,  'value' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT btrim( 'something' ,  'value' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT btrim( ' something' ,  'value' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT btrim( 'something ' ,  'value' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT btrim( '123456789' ,  'value' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT btrim( 'a group of words' ,  'value' ) ;",
					Expected: []sql.Row{{" group of words"}},
				},
				{
					Query:    "SELECT btrim( '1234567890123456' ,  'value' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT btrim( '' ,  '12345' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' ' ,  '12345' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT btrim( '0' ,  '12345' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT btrim( '1' ,  '12345' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'a' ,  '12345' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT btrim( 'abc' ,  '12345' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT btrim( '123' ,  '12345' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'value' ,  '12345' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT btrim( '12345' ,  '12345' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'something' ,  '12345' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT btrim( ' something' ,  '12345' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT btrim( 'something ' ,  '12345' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT btrim( '123456789' ,  '12345' ) ;",
					Expected: []sql.Row{{"6789"}},
				},
				{
					Query:    "SELECT btrim( 'a group of words' ,  '12345' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT btrim( '1234567890123456' ,  '12345' ) ;",
					Expected: []sql.Row{{"67890123456"}},
				},
				{
					Query:    "SELECT btrim( '' ,  'something' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' ' ,  'something' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT btrim( '0' ,  'something' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT btrim( '1' ,  'something' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT btrim( 'a' ,  'something' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT btrim( 'abc' ,  'something' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT btrim( '123' ,  'something' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT btrim( 'value' ,  'something' ) ;",
					Expected: []sql.Row{{"valu"}},
				},
				{
					Query:    "SELECT btrim( '12345' ,  'something' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT btrim( 'something' ,  'something' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' something' ,  'something' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT btrim( 'something ' ,  'something' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT btrim( '123456789' ,  'something' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT btrim( 'a group of words' ,  'something' ) ;",
					Expected: []sql.Row{{"a group of word"}},
				},
				{
					Query:    "SELECT btrim( '1234567890123456' ,  'something' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT btrim( '' ,  ' something' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' ' ,  ' something' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( '0' ,  ' something' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT btrim( '1' ,  ' something' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT btrim( 'a' ,  ' something' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT btrim( 'abc' ,  ' something' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT btrim( '123' ,  ' something' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT btrim( 'value' ,  ' something' ) ;",
					Expected: []sql.Row{{"valu"}},
				},
				{
					Query:    "SELECT btrim( '12345' ,  ' something' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT btrim( 'something' ,  ' something' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' something' ,  ' something' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'something ' ,  ' something' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( '123456789' ,  ' something' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT btrim( 'a group of words' ,  ' something' ) ;",
					Expected: []sql.Row{{"a group of word"}},
				},
				{
					Query:    "SELECT btrim( '1234567890123456' ,  ' something' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT btrim( '' ,  'something ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' ' ,  'something ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( '0' ,  'something ' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT btrim( '1' ,  'something ' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT btrim( 'a' ,  'something ' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT btrim( 'abc' ,  'something ' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT btrim( '123' ,  'something ' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT btrim( 'value' ,  'something ' ) ;",
					Expected: []sql.Row{{"valu"}},
				},
				{
					Query:    "SELECT btrim( '12345' ,  'something ' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT btrim( 'something' ,  'something ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' something' ,  'something ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'something ' ,  'something ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( '123456789' ,  'something ' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT btrim( 'a group of words' ,  'something ' ) ;",
					Expected: []sql.Row{{"a group of word"}},
				},
				{
					Query:    "SELECT btrim( '1234567890123456' ,  'something ' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT btrim( '' ,  '123456789' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' ' ,  '123456789' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT btrim( '0' ,  '123456789' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT btrim( '1' ,  '123456789' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'a' ,  '123456789' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT btrim( 'abc' ,  '123456789' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT btrim( '123' ,  '123456789' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'value' ,  '123456789' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT btrim( '12345' ,  '123456789' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'something' ,  '123456789' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT btrim( ' something' ,  '123456789' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT btrim( 'something ' ,  '123456789' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT btrim( '123456789' ,  '123456789' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'a group of words' ,  '123456789' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT btrim( '1234567890123456' ,  '123456789' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT btrim( '' ,  'a group of words' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' ' ,  'a group of words' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( '0' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT btrim( '1' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT btrim( 'a' ,  'a group of words' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'abc' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"bc"}},
				},
				{
					Query:    "SELECT btrim( '123' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT btrim( 'value' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT btrim( '12345' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT btrim( 'something' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"methin"}},
				},
				{
					Query:    "SELECT btrim( ' something' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"methin"}},
				},
				{
					Query:    "SELECT btrim( 'something ' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"methin"}},
				},
				{
					Query:    "SELECT btrim( '123456789' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT btrim( 'a group of words' ,  'a group of words' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( '1234567890123456' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT btrim( '' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( ' ' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT btrim( '0' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( '1' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'a' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT btrim( 'abc' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT btrim( '123' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'value' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT btrim( '12345' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'something' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT btrim( ' something' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT btrim( 'something ' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT btrim( '123456789' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT btrim( 'a group of words' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT btrim( '1234567890123456' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
			},
		},
	})
}
