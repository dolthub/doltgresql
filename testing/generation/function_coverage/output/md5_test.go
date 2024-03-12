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

func Test_Md5(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "md5",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT md5( '' ) ;",
					Expected: []sql.Row{{"d41d8cd98f00b204e9800998ecf8427e"}},
				},
				{
					Query:    "SELECT md5( ' ' ) ;",
					Expected: []sql.Row{{"7215ee9c7d9dc229d2921a40e899ec5f"}},
				},
				{
					Query:    "SELECT md5( '0' ) ;",
					Expected: []sql.Row{{"cfcd208495d565ef66e7dff9f98764da"}},
				},
				{
					Query:    "SELECT md5( '1' ) ;",
					Expected: []sql.Row{{"c4ca4238a0b923820dcc509a6f75849b"}},
				},
				{
					Query:    "SELECT md5( 'a' ) ;",
					Expected: []sql.Row{{"0cc175b9c0f1b6a831c399e269772661"}},
				},
				{
					Query:    "SELECT md5( 'abc' ) ;",
					Expected: []sql.Row{{"900150983cd24fb0d6963f7d28e17f72"}},
				},
				{
					Query:    "SELECT md5( '123' ) ;",
					Expected: []sql.Row{{"202cb962ac59075b964b07152d234b70"}},
				},
				{
					Query:    "SELECT md5( 'value' ) ;",
					Expected: []sql.Row{{"2063c1608d6e0baf80249c42e2be5804"}},
				},
				{
					Query:    "SELECT md5( '12345' ) ;",
					Expected: []sql.Row{{"827ccb0eea8a706c4c34a16891f84e7b"}},
				},
				{
					Query:    "SELECT md5( 'something' ) ;",
					Expected: []sql.Row{{"437b930db84b8079c2dd804a71936b5f"}},
				},
				{
					Query:    "SELECT md5( ' something' ) ;",
					Expected: []sql.Row{{"b0103a16a3cc4ee062a2d0ccf61f9617"}},
				},
				{
					Query:    "SELECT md5( 'something ' ) ;",
					Expected: []sql.Row{{"d23a679e329e8d6c028bbc6692b41dd8"}},
				},
				{
					Query:    "SELECT md5( '123456789' ) ;",
					Expected: []sql.Row{{"25f9e794323b453885f5181f1b624d0b"}},
				},
				{
					Query:    "SELECT md5( 'a group of words' ) ;",
					Expected: []sql.Row{{"231d49f0f13a285a2e6c24e0bb47a860"}},
				},
				{
					Query:    "SELECT md5( '1234567890123456' ) ;",
					Expected: []sql.Row{{"abeac07d3c28c1bef9e730002c753ed4"}},
				},
			},
		},
	})
}
