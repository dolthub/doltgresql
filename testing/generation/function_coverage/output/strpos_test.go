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

func Test_Strpos(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "strpos",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT strpos( '' ,  '' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( ' ' ,  '' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '0' ,  '' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '1' ,  '' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'a' ,  '' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'abc' ,  '' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '123' ,  '' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'value' ,  '' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '12345' ,  '' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'something' ,  '' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( ' something' ,  '' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'something ' ,  '' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '123456789' ,  '' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'a group of words' ,  '' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '1234567890123456' ,  '' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '' ,  ' ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' ' ,  ' ' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '0' ,  ' ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1' ,  ' ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a' ,  ' ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'abc' ,  ' ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123' ,  ' ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'value' ,  ' ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '12345' ,  ' ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something' ,  ' ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' something' ,  ' ' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'something ' ,  ' ' ) ;",
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    "SELECT strpos( '123456789' ,  ' ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a group of words' ,  ' ' ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT strpos( '1234567890123456' ,  ' ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '' ,  '0' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' ' ,  '0' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '0' ,  '0' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '1' ,  '0' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a' ,  '0' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'abc' ,  '0' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123' ,  '0' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'value' ,  '0' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '12345' ,  '0' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something' ,  '0' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' something' ,  '0' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something ' ,  '0' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123456789' ,  '0' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a group of words' ,  '0' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1234567890123456' ,  '0' ) ;",
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    "SELECT strpos( '' ,  '1' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' ' ,  '1' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '0' ,  '1' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1' ,  '1' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'a' ,  '1' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'abc' ,  '1' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123' ,  '1' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'value' ,  '1' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '12345' ,  '1' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'something' ,  '1' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' something' ,  '1' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something ' ,  '1' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123456789' ,  '1' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'a group of words' ,  '1' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1234567890123456' ,  '1' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '' ,  'a' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' ' ,  'a' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '0' ,  'a' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1' ,  'a' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a' ,  'a' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'abc' ,  'a' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '123' ,  'a' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'value' ,  'a' ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT strpos( '12345' ,  'a' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something' ,  'a' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' something' ,  'a' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something ' ,  'a' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123456789' ,  'a' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a group of words' ,  'a' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '1234567890123456' ,  'a' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '' ,  'abc' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' ' ,  'abc' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '0' ,  'abc' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1' ,  'abc' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a' ,  'abc' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'abc' ,  'abc' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '123' ,  'abc' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'value' ,  'abc' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '12345' ,  'abc' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something' ,  'abc' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' something' ,  'abc' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something ' ,  'abc' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123456789' ,  'abc' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a group of words' ,  'abc' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1234567890123456' ,  'abc' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '' ,  '123' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' ' ,  '123' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '0' ,  '123' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1' ,  '123' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a' ,  '123' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'abc' ,  '123' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123' ,  '123' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'value' ,  '123' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '12345' ,  '123' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'something' ,  '123' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' something' ,  '123' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something ' ,  '123' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123456789' ,  '123' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'a group of words' ,  '123' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1234567890123456' ,  '123' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '' ,  'value' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' ' ,  'value' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '0' ,  'value' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1' ,  'value' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a' ,  'value' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'abc' ,  'value' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123' ,  'value' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'value' ,  'value' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '12345' ,  'value' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something' ,  'value' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' something' ,  'value' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something ' ,  'value' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123456789' ,  'value' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a group of words' ,  'value' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1234567890123456' ,  'value' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '' ,  '12345' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' ' ,  '12345' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '0' ,  '12345' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1' ,  '12345' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a' ,  '12345' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'abc' ,  '12345' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123' ,  '12345' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'value' ,  '12345' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '12345' ,  '12345' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'something' ,  '12345' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' something' ,  '12345' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something ' ,  '12345' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123456789' ,  '12345' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'a group of words' ,  '12345' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1234567890123456' ,  '12345' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '' ,  'something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' ' ,  'something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '0' ,  'something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1' ,  'something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a' ,  'something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'abc' ,  'something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123' ,  'something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'value' ,  'something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '12345' ,  'something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something' ,  'something' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( ' something' ,  'something' ) ;",
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    "SELECT strpos( 'something ' ,  'something' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '123456789' ,  'something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a group of words' ,  'something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1234567890123456' ,  'something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '' ,  ' something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' ' ,  ' something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '0' ,  ' something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1' ,  ' something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a' ,  ' something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'abc' ,  ' something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123' ,  ' something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'value' ,  ' something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '12345' ,  ' something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something' ,  ' something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' something' ,  ' something' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'something ' ,  ' something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123456789' ,  ' something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a group of words' ,  ' something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1234567890123456' ,  ' something' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '' ,  'something ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' ' ,  'something ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '0' ,  'something ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1' ,  'something ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a' ,  'something ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'abc' ,  'something ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123' ,  'something ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'value' ,  'something ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '12345' ,  'something ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something' ,  'something ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' something' ,  'something ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something ' ,  'something ' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '123456789' ,  'something ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a group of words' ,  'something ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1234567890123456' ,  'something ' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '' ,  '123456789' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' ' ,  '123456789' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '0' ,  '123456789' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1' ,  '123456789' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a' ,  '123456789' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'abc' ,  '123456789' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123' ,  '123456789' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'value' ,  '123456789' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '12345' ,  '123456789' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something' ,  '123456789' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' something' ,  '123456789' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something ' ,  '123456789' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123456789' ,  '123456789' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( 'a group of words' ,  '123456789' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1234567890123456' ,  '123456789' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '' ,  'a group of words' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' ' ,  'a group of words' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '0' ,  'a group of words' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1' ,  'a group of words' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a' ,  'a group of words' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'abc' ,  'a group of words' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123' ,  'a group of words' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'value' ,  'a group of words' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '12345' ,  'a group of words' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something' ,  'a group of words' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' something' ,  'a group of words' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something ' ,  'a group of words' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123456789' ,  'a group of words' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a group of words' ,  'a group of words' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    "SELECT strpos( '1234567890123456' ,  'a group of words' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' ' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '0' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'abc' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'value' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '12345' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( ' something' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'something ' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '123456789' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( 'a group of words' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    "SELECT strpos( '1234567890123456' ,  '1234567890123456' ) ;",
					Expected: []sql.Row{{int32(1)}},
				},
			},
		},
	})
}
