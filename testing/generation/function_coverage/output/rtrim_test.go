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

func Test_Rtrim(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "rtrim",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT rtrim( '' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( '0' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT rtrim( '1' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT rtrim( 'a' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT rtrim( '123' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT rtrim( 'value' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT rtrim( '12345' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT rtrim( 'something' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT rtrim( ' something' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT rtrim( '' ,  '' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ,  '' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT rtrim( '0' ,  '' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT rtrim( '1' ,  '' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT rtrim( 'a' ,  '' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ,  '' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT rtrim( '123' ,  '' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT rtrim( 'value' ,  '' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT rtrim( '12345' ,  '' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT rtrim( 'something' ,  '' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT rtrim( ' something' ,  '' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ,  '' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ,  '' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ,  '' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ,  '' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT rtrim( '' ,  ' ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ,  ' ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( '0' ,  ' ' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT rtrim( '1' ,  ' ' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT rtrim( 'a' ,  ' ' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ,  ' ' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT rtrim( '123' ,  ' ' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT rtrim( 'value' ,  ' ' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT rtrim( '12345' ,  ' ' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT rtrim( 'something' ,  ' ' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT rtrim( ' something' ,  ' ' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ,  ' ' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ,  ' ' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ,  ' ' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ,  ' ' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT rtrim( '' ,  '0' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ,  '0' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT rtrim( '0' ,  '0' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( '1' ,  '0' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT rtrim( 'a' ,  '0' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ,  '0' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT rtrim( '123' ,  '0' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT rtrim( 'value' ,  '0' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT rtrim( '12345' ,  '0' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT rtrim( 'something' ,  '0' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT rtrim( ' something' ,  '0' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ,  '0' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ,  '0' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ,  '0' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ,  '0' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT rtrim( '' ,  '1' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ,  '1' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT rtrim( '0' ,  '1' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT rtrim( '1' ,  '1' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'a' ,  '1' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ,  '1' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT rtrim( '123' ,  '1' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT rtrim( 'value' ,  '1' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT rtrim( '12345' ,  '1' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT rtrim( 'something' ,  '1' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT rtrim( ' something' ,  '1' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ,  '1' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ,  '1' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ,  '1' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ,  '1' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT rtrim( '' ,  'a' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ,  'a' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT rtrim( '0' ,  'a' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT rtrim( '1' ,  'a' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT rtrim( 'a' ,  'a' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ,  'a' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT rtrim( '123' ,  'a' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT rtrim( 'value' ,  'a' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT rtrim( '12345' ,  'a' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT rtrim( 'something' ,  'a' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT rtrim( ' something' ,  'a' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ,  'a' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ,  'a' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ,  'a' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ,  'a' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT rtrim( '' ,  'abc' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ,  'abc' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT rtrim( '0' ,  'abc' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT rtrim( '1' ,  'abc' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT rtrim( 'a' ,  'abc' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ,  'abc' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( '123' ,  'abc' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT rtrim( 'value' ,  'abc' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT rtrim( '12345' ,  'abc' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT rtrim( 'something' ,  'abc' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT rtrim( ' something' ,  'abc' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ,  'abc' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ,  'abc' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ,  'abc' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ,  'abc' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT rtrim( '' ,  '123' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ,  '123' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT rtrim( '0' ,  '123' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT rtrim( '1' ,  '123' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'a' ,  '123' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ,  '123' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT rtrim( '123' ,  '123' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'value' ,  '123' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT rtrim( '12345' ,  '123' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT rtrim( 'something' ,  '123' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT rtrim( ' something' ,  '123' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ,  '123' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ,  '123' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ,  '123' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ,  '123' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT rtrim( '' ,  'value' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ,  'value' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT rtrim( '0' ,  'value' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT rtrim( '1' ,  'value' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT rtrim( 'a' ,  'value' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ,  'value' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT rtrim( '123' ,  'value' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT rtrim( 'value' ,  'value' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( '12345' ,  'value' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT rtrim( 'something' ,  'value' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT rtrim( ' something' ,  'value' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ,  'value' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ,  'value' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ,  'value' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ,  'value' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT rtrim( '' ,  '12345' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ,  '12345' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT rtrim( '0' ,  '12345' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT rtrim( '1' ,  '12345' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'a' ,  '12345' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ,  '12345' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT rtrim( '123' ,  '12345' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'value' ,  '12345' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT rtrim( '12345' ,  '12345' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'something' ,  '12345' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT rtrim( ' something' ,  '12345' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ,  '12345' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ,  '12345' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ,  '12345' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ,  '12345' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT rtrim( '' ,  'something' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ,  'something' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT rtrim( '0' ,  'something' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT rtrim( '1' ,  'something' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT rtrim( 'a' ,  'something' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ,  'something' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT rtrim( '123' ,  'something' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT rtrim( 'value' ,  'something' ) ;",
					Expected: []sql.Row{{"valu"}},
				},
				{
					Query:    "SELECT rtrim( '12345' ,  'something' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT rtrim( 'something' ,  'something' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' something' ,  'something' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ,  'something' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ,  'something' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ,  'something' ) ;",
					Expected: []sql.Row{{"a group of word"}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ,  'something' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT rtrim( '' ,  ' something' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ,  ' something' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( '0' ,  ' something' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT rtrim( '1' ,  ' something' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT rtrim( 'a' ,  ' something' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ,  ' something' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT rtrim( '123' ,  ' something' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT rtrim( 'value' ,  ' something' ) ;",
					Expected: []sql.Row{{"valu"}},
				},
				{
					Query:    "SELECT rtrim( '12345' ,  ' something' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT rtrim( 'something' ,  ' something' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' something' ,  ' something' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ,  ' something' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ,  ' something' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ,  ' something' ) ;",
					Expected: []sql.Row{{"a group of word"}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ,  ' something' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT rtrim( '' ,  'something ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ,  'something ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( '0' ,  'something ' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT rtrim( '1' ,  'something ' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT rtrim( 'a' ,  'something ' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ,  'something ' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT rtrim( '123' ,  'something ' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT rtrim( 'value' ,  'something ' ) ;",
					Expected: []sql.Row{{"valu"}},
				},
				{
					Query:    "SELECT rtrim( '12345' ,  'something ' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT rtrim( 'something' ,  'something ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' something' ,  'something ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ,  'something ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ,  'something ' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ,  'something ' ) ;",
					Expected: []sql.Row{{"a group of word"}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ,  'something ' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT rtrim( '' ,  '123456789' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ,  '123456789' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT rtrim( '0' ,  '123456789' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT rtrim( '1' ,  '123456789' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'a' ,  '123456789' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ,  '123456789' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT rtrim( '123' ,  '123456789' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'value' ,  '123456789' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT rtrim( '12345' ,  '123456789' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'something' ,  '123456789' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT rtrim( ' something' ,  '123456789' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ,  '123456789' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ,  '123456789' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ,  '123456789' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ,  '123456789' ) ;",
					Expected: []sql.Row{{"1234567890"}},
				},
				{
					Query:    "SELECT rtrim( '' ,  'a group of words' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ,  'a group of words' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( '0' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT rtrim( '1' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT rtrim( 'a' ,  'a group of words' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT rtrim( '123' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT rtrim( 'value' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT rtrim( '12345' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT rtrim( 'something' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"somethin"}},
				},
				{
					Query:    "SELECT rtrim( ' something' ,  'a group of words' ) ;",
					Expected: []sql.Row{{" somethin"}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"somethin"}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ,  'a group of words' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT rtrim( '' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( ' ' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT rtrim( '0' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( '1' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'a' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT rtrim( 'abc' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT rtrim( '123' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'value' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT rtrim( '12345' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'something' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT rtrim( ' something' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT rtrim( 'something ' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT rtrim( '123456789' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT rtrim( 'a group of words' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT rtrim( '1234567890123456' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
			},
		},
	})
}
