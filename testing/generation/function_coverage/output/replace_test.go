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

func Test_Replace(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "replace",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT replace( '1' ,  '' ,  '' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '' ,  '' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  ' ' ,  '' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'something' ,  ' ' ,  '' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '0' ,  '' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '123' ,  '0' ,  '' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  '0' ,  '' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '0' ,  '' ) ;",
					Expected: []sql.Row{{"123456789123456"}},
				},
				{
					Query:    "SELECT replace( '' ,  '1' ,  '' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '1' ,  '1' ,  '' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  '1' ,  '' ) ;",
					Expected: []sql.Row{{"23456789"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  'a' ,  '' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( '0' ,  'abc' ,  '' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  'abc' ,  '' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '123' ,  '' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  '123' ,  '' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '123' ,  '' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '123' ,  '' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'value' ,  '' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '123' ,  'value' ,  '' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  'value' ,  '' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( ' something' ,  'value' ,  '' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  'value' ,  '' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '12345' ,  '' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'something' ,  '' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '1' ,  ' something' ,  '' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  'something ' ,  '' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'something ' ,  '' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '' ,  '123456789' ,  '' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( ' ' ,  '123456789' ,  '' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( '0' ,  '123456789' ,  '' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  '123456789' ,  '' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '123456789' ,  '' ) ;",
					Expected: []sql.Row{{"0123456"}},
				},
				{
					Query:    "SELECT replace( '' ,  'a group of words' ,  '' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'a group of words' ,  '' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'a group of words' ,  '' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '1234567890123456' ,  '' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( '' ,  '' ,  ' ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '1' ,  '' ,  ' ' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  '' ,  ' ' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '' ,  ' ' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  '' ,  ' ' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  ' ' ,  ' ' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( '' ,  '0' ,  ' ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '0' ,  ' ' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '0' ,  ' ' ) ;",
					Expected: []sql.Row{{"123456789 123456"}},
				},
				{
					Query:    "SELECT replace( '' ,  '1' ,  ' ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '1' ,  ' ' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '1' ,  ' ' ) ;",
					Expected: []sql.Row{{" 2345"}},
				},
				{
					Query:    "SELECT replace( '0' ,  'a' ,  ' ' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '1' ,  'a' ,  ' ' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'a' ,  ' ' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'a' ,  ' ' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '0' ,  'abc' ,  ' ' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'abc' ,  ' ' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  'abc' ,  ' ' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'abc' ,  ' ' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '1' ,  '123' ,  ' ' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '123' ,  ' ' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( 'something' ,  'value' ,  ' ' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  'value' ,  ' ' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  'value' ,  ' ' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( '' ,  '12345' ,  ' ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'something' ,  '12345' ,  ' ' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '12345' ,  ' ' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  'something' ,  ' ' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( '0' ,  ' something' ,  ' ' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  ' something' ,  ' ' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  ' something' ,  ' ' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  ' something' ,  ' ' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( '0' ,  'something ' ,  ' ' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '1' ,  'something ' ,  ' ' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  '123456789' ,  ' ' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( '' ,  'a group of words' ,  ' ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( ' ' ,  'a group of words' ,  ' ' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( '1' ,  'a group of words' ,  ' ' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  'a group of words' ,  ' ' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '1234567890123456' ,  ' ' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '1234567890123456' ,  ' ' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '1234567890123456' ,  ' ' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  '1234567890123456' ,  ' ' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '' ,  '0' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  ' ' ,  '0' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '123' ,  ' ' ,  '0' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '0' ,  '0' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  '0' ,  '0' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '0' ,  '0' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '1' ,  '0' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '1' ,  '0' ) ;",
					Expected: []sql.Row{{"02345"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  '1' ,  '0' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  'a' ,  '0' ) ;",
					Expected: []sql.Row{{"v0lue"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  'a' ,  '0' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'a' ,  '0' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  'a' ,  '0' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  '123' ,  '0' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( '1' ,  '123' ,  '0' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '123' ,  '0' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  '123' ,  '0' ) ;",
					Expected: []sql.Row{{"0456789"}},
				},
				{
					Query:    "SELECT replace( '1' ,  'value' ,  '0' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'value' ,  '0' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  'value' ,  '0' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  '12345' ,  '0' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  '12345' ,  '0' ) ;",
					Expected: []sql.Row{{"06789"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  'something' ,  '0' ) ;",
					Expected: []sql.Row{{"0 "}},
				},
				{
					Query:    "SELECT replace( 'something' ,  ' something' ,  '0' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( '' ,  '123456789' ,  '0' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'a group of words' ,  '0' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  'a group of words' ,  '0' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '123' ,  '1234567890123456' ,  '0' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '1234567890123456' ,  '0' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '' ,  '1' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '123' ,  ' ' ,  '1' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  ' ' ,  '1' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  ' ' ,  '1' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '0' ,  '0' ,  '1' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  '0' ,  '1' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '0' ,  '1' ) ;",
					Expected: []sql.Row{{"1234567891123456"}},
				},
				{
					Query:    "SELECT replace( '' ,  '1' ,  '1' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '1' ,  'a' ,  '1' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'a' ,  '1' ) ;",
					Expected: []sql.Row{{"1bc"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'a' ,  '1' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  'abc' ,  '1' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  'abc' ,  '1' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'abc' ,  '1' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '0' ,  '123' ,  '1' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '123' ,  '1' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '123' ,  '1' ) ;",
					Expected: []sql.Row{{"145678901456"}},
				},
				{
					Query:    "SELECT replace( '0' ,  'value' ,  '1' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '' ,  '12345' ,  '1' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '1' ,  '12345' ,  '1' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'something' ,  '1' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'something' ,  '1' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'something' ,  '1' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '' ,  ' something' ,  '1' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( ' ' ,  'something ' ,  '1' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( 'something' ,  'something ' ,  '1' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  '123456789' ,  '1' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( '0' ,  'a group of words' ,  '1' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '1' ,  'a group of words' ,  '1' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'a group of words' ,  '1' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '' ,  ' ' ,  'a' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'a' ,  ' ' ,  'a' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  ' ' ,  'a' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  '0' ,  'a' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '0' ,  'a' ) ;",
					Expected: []sql.Row{{"123456789a123456"}},
				},
				{
					Query:    "SELECT replace( '' ,  '1' ,  'a' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '0' ,  '1' ,  'a' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '1' ,  'a' ) ;",
					Expected: []sql.Row{{"a234567890a23456"}},
				},
				{
					Query:    "SELECT replace( '1' ,  'a' ,  'a' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( '123' ,  'a' ,  'a' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  'a' ,  'a' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  'abc' ,  'a' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '123' ,  'a' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( '' ,  'value' ,  'a' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '0' ,  'value' ,  'a' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '1' ,  'value' ,  'a' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  '12345' ,  'a' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '12345' ,  'a' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '0' ,  ' something' ,  'a' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '123' ,  ' something' ,  'a' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  ' something' ,  'a' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  ' something' ,  'a' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  ' something' ,  'a' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( '123' ,  'something ' ,  'a' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  '123456789' ,  'a' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'a group of words' ,  'a' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '' ,  '1234567890123456' ,  'a' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( ' ' ,  '1234567890123456' ,  'a' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '1234567890123456' ,  'a' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '1234567890123456' ,  'a' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( 'something' ,  '' ,  'abc' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '' ,  'abc' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  '' ,  'abc' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '' ,  'abc' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  ' ' ,  'abc' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '123' ,  ' ' ,  'abc' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  ' ' ,  'abc' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '' ,  '0' ,  'abc' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '0' ,  'abc' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '0' ,  'abc' ) ;",
					Expected: []sql.Row{{"123456789abc123456"}},
				},
				{
					Query:    "SELECT replace( '0' ,  '1' ,  'abc' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '1' ,  'abc' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '123' ,  '1' ,  'abc' ) ;",
					Expected: []sql.Row{{"abc23"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  '1' ,  'abc' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  '1' ,  'abc' ) ;",
					Expected: []sql.Row{{"abc23456789"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  'a' ,  'abc' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( '' ,  'abc' ,  'abc' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'value' ,  'abc' ,  'abc' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'abc' ,  'abc' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'abc' ,  'abc' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '' ,  'value' ,  'abc' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '12345' ,  'abc' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  '12345' ,  'abc' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'something' ,  'abc' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  ' something' ,  'abc' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  'something ' ,  'abc' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( '' ,  '123456789' ,  'abc' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '123' ,  '123456789' ,  'abc' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  '123456789' ,  'abc' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '123456789' ,  'abc' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '123456789' ,  'abc' ) ;",
					Expected: []sql.Row{{"abc0123456"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'a group of words' ,  'abc' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'a group of words' ,  'abc' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  'a group of words' ,  'abc' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  'a group of words' ,  'abc' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '1234567890123456' ,  'abc' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '' ,  '' ,  '123' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '1' ,  ' ' ,  '123' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  ' ' ,  '123' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  ' ' ,  '123' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '0' ,  '123' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '0' ,  '123' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '1' ,  '123' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '0' ,  'a' ,  '123' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '1' ,  '123' ,  '123' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '123' ,  '123' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '' ,  'value' ,  '123' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'value' ,  '123' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'value' ,  '123' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'value' ,  '123' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '0' ,  '12345' ,  '123' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '12345' ,  '123' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  '12345' ,  '123' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'something' ,  '123' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  'something' ,  '123' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  'something' ,  '123' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  ' something' ,  '123' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '123' ,  ' something' ,  '123' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( '0' ,  '123456789' ,  '123' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '123456789' ,  '123' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '1' ,  '1234567890123456' ,  '123' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '' ,  'value' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( '' ,  ' ' ,  'value' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  ' ' ,  'value' ) ;",
					Expected: []sql.Row{{"somethingvalue"}},
				},
				{
					Query:    "SELECT replace( '' ,  '0' ,  'value' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  '1' ,  'value' ) ;",
					Expected: []sql.Row{{"value23456789"}},
				},
				{
					Query:    "SELECT replace( '' ,  'a' ,  'value' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'a' ,  'value' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  'a' ,  'value' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'abc' ,  'value' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'abc' ,  'value' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '123' ,  'value' ) ;",
					Expected: []sql.Row{{"value45"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  '123' ,  'value' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  '123' ,  'value' ) ;",
					Expected: []sql.Row{{"value456789"}},
				},
				{
					Query:    "SELECT replace( '' ,  'value' ,  'value' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'value' ,  'value' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  'value' ,  'value' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  'value' ,  'value' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '12345' ,  'value' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '12345' ,  'value' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( ' ' ,  'something' ,  'value' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( '1' ,  'something' ,  'value' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( '' ,  ' something' ,  'value' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '0' ,  ' something' ,  'value' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  ' something' ,  'value' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  ' something' ,  'value' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  ' something' ,  'value' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( '' ,  'something ' ,  'value' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( ' ' ,  'something ' ,  'value' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'something ' ,  'value' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  'something ' ,  'value' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '0' ,  '123456789' ,  'value' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '123' ,  '123456789' ,  'value' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '123456789' ,  'value' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( '0' ,  '1234567890123456' ,  'value' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  '' ,  '12345' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '' ,  '12345' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  '' ,  '12345' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '' ,  '12345' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  ' ' ,  '12345' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  ' ' ,  '12345' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  ' ' ,  '12345' ) ;",
					Expected: []sql.Row{{"12345something"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  ' ' ,  '12345' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '1' ,  '0' ,  '12345' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( '123' ,  '0' ,  '12345' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '0' ,  '12345' ) ;",
					Expected: []sql.Row{{"12345678912345123456"}},
				},
				{
					Query:    "SELECT replace( '1' ,  '1' ,  '12345' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  '1' ,  '12345' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  '1' ,  '12345' ) ;",
					Expected: []sql.Row{{"1234523456789"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '1' ,  '12345' ) ;",
					Expected: []sql.Row{{"123452345678901234523456"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  'a' ,  '12345' ) ;",
					Expected: []sql.Row{{"v12345lue"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  'a' ,  '12345' ) ;",
					Expected: []sql.Row{{"12345 group of words"}},
				},
				{
					Query:    "SELECT replace( '0' ,  'abc' ,  '12345' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'abc' ,  '12345' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  'abc' ,  '12345' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( '1' ,  '123' ,  '12345' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '123' ,  '12345' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  '123' ,  '12345' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( '' ,  'value' ,  '12345' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '123' ,  'value' ,  '12345' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '12345' ,  '12345' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '12345' ,  '12345' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  '12345' ,  '12345' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'something' ,  '12345' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'something' ,  '12345' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  'something' ,  '12345' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( '' ,  ' something' ,  '12345' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '123' ,  ' something' ,  '12345' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( '0' ,  'something ' ,  '12345' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  'something ' ,  '12345' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  'something ' ,  '12345' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'something ' ,  '12345' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '123456789' ,  '12345' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '0' ,  'a group of words' ,  '12345' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'a group of words' ,  '12345' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '1234567890123456' ,  '12345' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '123' ,  '' ,  'something' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  '' ,  'something' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  ' ' ,  'something' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  ' ' ,  'something' ) ;",
					Expected: []sql.Row{{"asomethinggroupsomethingofsomethingwords"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '0' ,  'something' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '123' ,  '0' ,  'something' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '0' ,  'something' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  '1' ,  'something' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '1' ,  'something' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  '1' ,  'something' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( '1' ,  'a' ,  'something' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( '1' ,  'abc' ,  'something' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'abc' ,  'something' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'abc' ,  'something' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  'abc' ,  'something' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  '123' ,  'something' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( '1' ,  '123' ,  'something' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '123' ,  'something' ) ;",
					Expected: []sql.Row{{"something4567890something456"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  'value' ,  'something' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( '0' ,  '12345' ,  'something' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '12345' ,  'something' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  '12345' ,  'something' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  '12345' ,  'something' ) ;",
					Expected: []sql.Row{{"something6789"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'something' ,  'something' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  'something' ,  'something' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  'something ' ,  'something' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( '0' ,  '123456789' ,  'something' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '123456789' ,  'something' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'a group of words' ,  'something' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'a group of words' ,  'something' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'a group of words' ,  'something' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  '1234567890123456' ,  'something' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '1234567890123456' ,  'something' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  ' ' ,  ' something' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '0' ,  ' something' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '0' ,  ' something' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '1' ,  ' something' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'a' ,  ' something' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'a' ,  ' something' ) ;",
					Expected: []sql.Row{{" somethingbc"}},
				},
				{
					Query:    "SELECT replace( '0' ,  'abc' ,  ' something' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '123' ,  'abc' ,  ' something' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  'abc' ,  ' something' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'abc' ,  ' something' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '' ,  '123' ,  ' something' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '123' ,  ' something' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '123' ,  '123' ,  ' something' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  '123' ,  ' something' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '0' ,  'value' ,  ' something' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  'value' ,  ' something' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'value' ,  ' something' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '12345' ,  ' something' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '12345' ,  ' something' ) ;",
					Expected: []sql.Row{{" something67890 something6"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  'something' ,  ' something' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( '1' ,  'something' ,  ' something' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  'something' ,  ' something' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  'something' ,  ' something' ) ;",
					Expected: []sql.Row{{"  something"}},
				},
				{
					Query:    "SELECT replace( '0' ,  ' something' ,  ' something' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  ' something' ,  ' something' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  'something ' ,  ' something' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( 'value' ,  'something ' ,  ' something' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  'something ' ,  ' something' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  'a group of words' ,  ' something' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( '123' ,  '1234567890123456' ,  ' something' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '' ,  'something ' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  ' ' ,  'something ' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  ' ' ,  'something ' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  ' ' ,  'something ' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  '0' ,  'something ' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  '0' ,  'something ' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( '0' ,  '1' ,  'something ' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '1' ,  'something ' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '1' ,  'something ' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'a' ,  'something ' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '0' ,  'abc' ,  'something ' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'abc' ,  'something ' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '123' ,  'abc' ,  'something ' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'abc' ,  'something ' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '123' ,  '123' ,  'something ' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '123' ,  'something ' ) ;",
					Expected: []sql.Row{{"something 45"}},
				},
				{
					Query:    "SELECT replace( '' ,  '12345' ,  'something ' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'value' ,  '12345' ,  'something ' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  'something' ,  'something ' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'something' ,  'something ' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'something' ,  'something ' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  ' something' ,  'something ' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  ' something' ,  'something ' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'something ' ,  'something ' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '123456789' ,  'something ' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( '1' ,  '1234567890123456' ,  'something ' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( '0' ,  '' ,  '123456789' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '' ,  '123456789' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  ' ' ,  '123456789' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  ' ' ,  '123456789' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '' ,  '0' ,  '123456789' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  '1' ,  '123456789' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  'a' ,  '123456789' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  'a' ,  '123456789' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( '' ,  'abc' ,  '123456789' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '0' ,  'abc' ,  '123456789' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'abc' ,  '123456789' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  '123' ,  '123456789' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '123' ,  '123456789' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '123' ,  '123456789' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  '123' ,  '123456789' ) ;",
					Expected: []sql.Row{{"123456789456789"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'value' ,  '123456789' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'value' ,  '123456789' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '123' ,  '12345' ,  '123456789' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '12345' ,  '123456789' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  '12345' ,  '123456789' ) ;",
					Expected: []sql.Row{{"1234567896789"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '12345' ,  '123456789' ) ;",
					Expected: []sql.Row{{"123456789678901234567896"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'something' ,  '123456789' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  'something' ,  '123456789' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  ' something' ,  '123456789' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  ' something' ,  '123456789' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '1' ,  'something ' ,  '123456789' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  'something ' ,  '123456789' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  'something ' ,  '123456789' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( '' ,  '123456789' ,  '123456789' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '123' ,  '1234567890123456' ,  '123456789' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '' ,  'a group of words' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  '' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"something "}},
				},
				{
					Query:    "SELECT replace( 'value' ,  ' ' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  ' ' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  ' ' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( '' ,  '0' ,  'a group of words' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( '1' ,  '0' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '0' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  '1' ,  'a group of words' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( ' ' ,  'a' ,  'a group of words' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'a' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"a group of wordsbc"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  'a' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"va group of wordslue"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  'a' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  'a' ,  'a group of words' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  'abc' ,  'a group of words' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( '1' ,  'abc' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'abc' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  'abc' ,  'a group of words' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '123' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '123' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"a group of words45"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '123' ,  'a group of words' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( '0' ,  'value' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'value' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  '12345' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '12345' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"a group of words67890a group of words6"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  'something' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  'something' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"a group of words "}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  'something' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  ' something' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  ' something' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  'something ' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '123456789' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '123456789' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '' ,  'a group of words' ,  'a group of words' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( ' ' ,  'a group of words' ,  'a group of words' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '1234567890123456' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  '1234567890123456' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '1234567890123456' ,  'a group of words' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  '1234567890123456' ,  'a group of words' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( '123' ,  '' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  ' ' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( '0' ,  ' ' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  ' ' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  ' ' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"1234567890123456something"}},
				},
				{
					Query:    "SELECT replace( '123' ,  '0' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  '0' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '0' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT replace( '' ,  '1' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '1' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '12345' ,  '1' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"12345678901234562345"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  'a' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  'a' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( '123' ,  'abc' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT replace( 'something' ,  'abc' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT replace( '0' ,  '123' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '123' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( '123456789' ,  'value' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT replace( ' ' ,  '12345' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( '0' ,  '12345' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT replace( '1' ,  '12345' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'a' ,  '12345' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT replace( 'value' ,  '12345' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '12345' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( '1234567890123456' ,  '12345' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"12345678901234566789012345678901234566"}},
				},
				{
					Query:    "SELECT replace( '' ,  'something' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT replace( ' ' ,  'something' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( '1' ,  'something' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT replace( 'something ' ,  'something' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"1234567890123456 "}},
				},
				{
					Query:    "SELECT replace( ' ' ,  '123456789' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT replace( 'abc' ,  '123456789' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT replace( ' something' ,  '123456789' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  '123456789' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT replace( 'a group of words' ,  'a group of words' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT replace( '' ,  '1234567890123456' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{""}},
				},
			},
		},
	})
}
