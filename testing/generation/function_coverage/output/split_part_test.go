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

func Test_SplitPart(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "split_part",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT split_part( ' ' ,  '' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'a' ,  '' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '123456789' ,  '' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'abc' ,  ' ' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'value' ,  ' ' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '123456789' ,  ' ' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'a group of words' ,  ' ' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '123' ,  '0' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '1234567890123456' ,  '0' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '123456789' ,  '1' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'a group of words' ,  '1' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '0' ,  'a' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'something ' ,  'abc' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'a group of words' ,  'abc' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '0' ,  '123' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( ' something' ,  '123' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'something' ,  'value' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( ' something' ,  'value' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( ' something' ,  '12345' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '123456789' ,  '12345' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '1' ,  'something' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( ' something' ,  'something' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '1234567890123456' ,  'something' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '' ,  ' something' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'a' ,  ' something' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'value' ,  ' something' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '1234567890123456' ,  ' something' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'value' ,  '123456789' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '1234567890123456' ,  '123456789' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '1' ,  'a group of words' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '' ,  '1234567890123456' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( ' ' ,  '1234567890123456' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'abc' ,  '1234567890123456' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'value' ,  '1234567890123456' ,  0::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT split_part( '0' ,  ' ' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT split_part( '1' ,  ' ' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  ' ' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  ' ' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"words"}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  ' ' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  '0' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '0' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '0' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '0' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  '0' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"123456"}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '1' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT split_part( '' ,  'a' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  'a' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  'a' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  'a' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT split_part( '0' ,  'abc' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  'abc' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  'abc' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  '123' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '123' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  '123' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT split_part( '0' ,  'value' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  '12345' ,  -1::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT split_part( '1' ,  '12345' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '12345' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT split_part( '123' ,  '12345' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT split_part( '1' ,  'something' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  'something' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  'something' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT split_part( '' ,  ' something' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  ' something' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT split_part( '0' ,  'something ' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  'something ' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  'something ' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT split_part( '123' ,  '123456789' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  'a group of words' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  'a group of words' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  'a group of words' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  'a group of words' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"1234567890123456"}},
				},
				{
					Query:    "SELECT split_part( '' ,  '1234567890123456' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '1234567890123456' ,  -1::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  '1234567890123456' ,  -1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  '' ,  1::int4 ) ;",
					Expected: []sql.Row{{"0"}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  '' ,  1::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  '' ,  1::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  ' ' ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  ' ' ,  1::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  ' ' ,  1::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT split_part( '' ,  '0' ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  '0' ,  1::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  '0' ,  1::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '0' ,  1::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  '1' ,  1::int4 ) ;",
					Expected: []sql.Row{{" "}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  '1' ,  1::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '1' ,  1::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '1' ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  '1' ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  'a' ,  1::int4 ) ;",
					Expected: []sql.Row{{"123"}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  'abc' ,  1::int4 ) ;",
					Expected: []sql.Row{{"123456789"}},
				},
				{
					Query:    "SELECT split_part( '' ,  '123' ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '123' ,  1::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT split_part( '123' ,  '123' ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  '123' ,  1::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT split_part( '' ,  'value' ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  'value' ,  1::int4 ) ;",
					Expected: []sql.Row{{"abc"}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  'value' ,  1::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  '12345' ,  1::int4 ) ;",
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  '12345' ,  1::int4 ) ;",
					Expected: []sql.Row{{"value"}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  '12345' ,  1::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  '12345' ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  'something' ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  'something ' ,  1::int4 ) ;",
					Expected: []sql.Row{{"12345"}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  'something ' ,  1::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  'something ' ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  '123456789' ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  '123456789' ,  1::int4 ) ;",
					Expected: []sql.Row{{"something"}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  '123456789' ,  1::int4 ) ;",
					Expected: []sql.Row{{" something"}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '123456789' ,  1::int4 ) ;",
					Expected: []sql.Row{{"a group of words"}},
				},
				{
					Query:    "SELECT split_part( '1' ,  'a group of words' ,  1::int4 ) ;",
					Expected: []sql.Row{{"1"}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  '1234567890123456' ,  1::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  '' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  ' ' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  ' ' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  '0' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '0' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '0' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  '0' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '0' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  'a' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  'a' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  'abc' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  'abc' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  'abc' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  'value' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  '12345' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  'something' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  'something' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  'something' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  ' something' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  ' something' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  ' something' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  'something ' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  '123456789' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  '123456789' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  'a group of words' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  'a group of words' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  '1234567890123456' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  '1234567890123456' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '1234567890123456' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  '1234567890123456' ,  2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  '' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '0' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  '0' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  'a' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  'a' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  'a' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  'abc' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '123' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  'value' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '12345' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '12345' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  'something' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  'something' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  'something' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  ' something' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  'something ' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  'something ' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  'something ' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  'a group of words' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  'a group of words' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  '1234567890123456' ,  -2::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  '' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  ' ' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  ' ' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  '0' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  '1' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  '1' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '1' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  '1' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  'a' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  'a' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  'abc' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  'abc' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  '123' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  '123' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  'value' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  'value' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  'value' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  'something' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  'something' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  ' something' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  'something ' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  'something ' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '123456789' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '123456789' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  '123456789' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  'a group of words' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  'a group of words' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  'a group of words' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  '1234567890123456' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '1234567890123456' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  '1234567890123456' ,  5::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  ' ' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  ' ' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  ' ' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  ' ' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  ' ' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '0' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  '0' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  '1' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  '1' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  'a' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  'a' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  'abc' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  'abc' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  'abc' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  '123' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '123' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  'value' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  '12345' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  ' something' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  ' something' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  ' something' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  ' something' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  'something ' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  'something ' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  '123456789' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  '123456789' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  'a group of words' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  'a group of words' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  '1234567890123456' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  '1234567890123456' ,  10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  '' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  '' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  ' ' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  ' ' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  ' ' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '0' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  '0' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '1' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  '1' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  'a' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  'a' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  'a' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  'value' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  '12345' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  'something' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  ' something' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  'something ' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  'something ' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  'something ' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  '123456789' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  '123456789' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  '123456789' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  '123456789' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  'a group of words' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  'a group of words' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  'a group of words' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  '1234567890123456' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  '1234567890123456' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '1234567890123456' ,  -10::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  '' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  '' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  ' ' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  '0' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  '0' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '0' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '1' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '1' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  '1' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  '1' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  'a' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  'a' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  'a' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  'a' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  'abc' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  '123' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  '123' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  'value' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  'value' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  '12345' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  'something' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  '123456789' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  '123456789' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '123456789' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '123456789' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  'a group of words' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  'a group of words' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  '1234567890123456' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  '1234567890123456' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '1234567890123456' ,  100::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  '' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  '' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  '' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  ' ' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  '0' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  '1' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  '1' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  'abc' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  'abc' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  '123' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  'value' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  'value' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  'value' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  '12345' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  '12345' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  '12345' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  'something' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  'something' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  ' something' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  ' something' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  'something ' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  'something ' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  '123456789' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  '123456789' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  '123456789' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  'a group of words' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  'a group of words' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  '1234567890123456' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  '1234567890123456' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '1234567890123456' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '1234567890123456' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  '1234567890123456' ,  21050::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  '' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  '0' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  '0' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '0' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '0' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  '1' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  'a' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  'a' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  'a' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  'abc' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  'abc' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  'abc' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  '123' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  '123' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '123' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  '123' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  'value' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  '12345' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '12345' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  'something' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  'something' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  'something' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  'something' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  'something' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  'something' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  ' something' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  'something ' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  '123456789' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  '123456789' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  'a group of words' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '1234567890123456' ,  100000::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  ' ' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  ' ' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  ' ' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  '0' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  '0' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '0' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '1' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  '1' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  'a' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  'abc' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  'abc' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  'abc' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '123' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  'value' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  'value' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  'value' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  'value' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  '12345' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  '12345' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '12345' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  'something' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  'something' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  'something' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  ' something' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  ' something' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  ' something' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  'something ' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  'something ' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  '123456789' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  '123456789' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '123456789' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  'a group of words' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  'a group of words' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  'a group of words' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  'a group of words' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  '1234567890123456' ,  -1184280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  '' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  ' ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  ' ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  ' ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  ' ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  ' ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  ' ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  ' ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  '0' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  '0' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  '0' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  '0' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  '1' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '1' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  '1' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  'a' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  'a' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  'abc' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  '123' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  'value' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  'value' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  'value' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  'value' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '12345' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '12345' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  'something' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  'something' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  ' something' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  ' something' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  'something ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  'something ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  'something ' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  '123456789' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '123456789' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '123456789' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  '123456789' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  'a group of words' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  'a group of words' ,  2525280::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:       "SELECT split_part( ' ' ,  '' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'a' ,  '' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'a group of words' ,  '' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '1234567890123456' ,  '' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '1' ,  '0' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'a group of words' ,  '0' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'abc' ,  '1' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'value' ,  '1' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'a' ,  'a' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'a group of words' ,  'a' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'value' ,  '123' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '' ,  'value' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'value' ,  'value' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '1' ,  'something' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '1234567890123456' ,  'something' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( ' ' ,  ' something' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'a' ,  ' something' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'abc' ,  ' something' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'something ' ,  ' something' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( ' ' ,  'something ' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'a' ,  'something ' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( ' something' ,  'something ' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '123456789' ,  'something ' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( ' something' ,  '123456789' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'a group of words' ,  '123456789' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '123' ,  'a group of words' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '12345' ,  'a group of words' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '123456789' ,  '1234567890123456' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( 'a group of words' ,  '1234567890123456' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:       "SELECT split_part( '1234567890123456' ,  '1234567890123456' ,  -2147483648::int4 ) ;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT split_part( '' ,  ' ' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  ' ' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  ' ' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  ' ' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  '0' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '0' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  '0' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  '1' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  '1' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  '1' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '1' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  'a' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  'a' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  'a' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  'abc' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  'abc' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  'abc' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123' ,  '123' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  '123' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '123456789' ,  '123' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  'value' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'value' ,  'value' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  'value' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something ' ,  'value' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  'value' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  '12345' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  ' something' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  ' something' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  ' something' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' something' ,  ' something' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1' ,  'something ' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a' ,  'something ' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'abc' ,  '123456789' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  '123456789' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '' ,  'a group of words' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( ' ' ,  'a group of words' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '0' ,  'a group of words' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '12345' ,  'a group of words' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'something' ,  'a group of words' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( '1234567890123456' ,  'a group of words' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
				{
					Query:    "SELECT split_part( 'a group of words' ,  '1234567890123456' ,  2147483647::int4 ) ;",
					Expected: []sql.Row{{""}},
				},
			},
		},
	})
}
