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

package regression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestComments(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_comments)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_comments,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT 'trailing' AS first; -- trailing single line`,
				Results:   []sql.Row{{`trailing`}},
			},
			{
				Statement: `SELECT /* embedded single line */ 'embedded' AS second;`,
				Results:   []sql.Row{{`embedded`}},
			},
			{
				Statement: `SELECT /* both embedded and trailing single line */ 'both' AS third; -- trailing single line`,
				Results:   []sql.Row{{`both`}},
			},
			{
				Statement: `SELECT 'before multi-line' AS fourth;`,
				Results:   []sql.Row{{`before multi-line`}},
			},
			{
				Statement: `/* This is an example of SQL which should not execute:
 * select 'multi-line';`,
			},
			{
				Statement: ` */
SELECT 'after multi-line' AS fifth;`,
				Results: []sql.Row{{`after multi-line`}},
			},
			{
				Statement: `/*
SELECT 'trailing' as x1; -- inside block comment`,
			},
			{
				Statement: `*/
/* This block comment surrounds a query which itself has a block comment...
SELECT /* embedded single line */ 'embedded' AS x2;`,
			},
			{
				Statement: `*/
SELECT -- continued after the following block comments...
/* Deeply nested comment.
   This includes a single apostrophe to make sure we aren't decoding this part as a string.
SELECT 'deep nest' AS n1;`,
			},
			{
				Statement: `/* Second level of nesting...
SELECT 'deeper nest' as n2;`,
			},
			{
				Statement: `/* Third level of nesting...
SELECT 'deepest nest' as n3;`,
			},
			{
				Statement: `*/
Hoo boy. Still two deep...
*/
Now just one deep...
*/
'deeply nested example' AS sixth;`,
				Results: []sql.Row{{`deeply nested example`}},
			},
			{
				Statement: `/* and this is the end of the file */`,
			},
		},
	})
}
