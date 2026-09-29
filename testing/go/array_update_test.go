// Copyright 2026 Dolthub, Inc.
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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

// TestArraySubscriptUpdate checks array updates, rejected assignments, and preservation of unaffected rows.
func TestArraySubscriptUpdate(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:        "array subscript updates",
			SetUpScript: []string{"CREATE TABLE t (id int PRIMARY KEY, a int[]);", "INSERT INTO t VALUES (1,ARRAY[[1,2,3],[4,5,6]]),(2,ARRAY[1,2,3]),(3,NULL);"},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "UPDATE t SET a[2][3]=60 WHERE id=1 RETURNING a;",
					Expected: []sql.Row{{"{{1,2,3},{4,5,60}}"}},
				},
				{
					Query:    "UPDATE t SET a[1:2][2:3]=ARRAY[[20,30],[50,60]] WHERE id=1 RETURNING a;",
					Expected: []sql.Row{{"{{1,20,30},{4,50,60}}"}},
				},
				{
					Query:    "UPDATE t SET a[6]=99 WHERE id=2 RETURNING a;",
					Expected: []sql.Row{{"{1,2,3,NULL,NULL,99}"}},
				},
				{
					Query:    "UPDATE t SET a[1][1]=7 WHERE id=3 RETURNING a;",
					Expected: []sql.Row{{"{{7}}"}},
				},
				{
					Query:    "UPDATE t SET a[:][2:3]=ARRAY[[2,3],[5,6]] WHERE id=1 RETURNING a;",
					Expected: []sql.Row{{"{{1,2,3},{4,5,6}}"}},
				},
				{
					Query:           "UPDATE t SET a[3][1]=7 WHERE id=1;",
					ExpectedErr:     "array subscript out of range",
					ExpectedErrCode: "2202E",
				},
				{
					Query:           "UPDATE t SET a[1:2][1:2]=ARRAY[[9,8]] WHERE id=1;",
					ExpectedErr:     "source array too small",
					ExpectedErrCode: "2202E",
				},
				{
					Query:           "UPDATE t SET a[NULL]=7 WHERE id=2;",
					ExpectedErr:     "must not be null",
					ExpectedErrCode: "22004",
				},
				{Query: "SELECT a FROM t WHERE id=1;", Expected: []sql.Row{{"{{1,2,3},{4,5,6}}"}}},
				{
					Query:           "UPDATE t SET a[NULL:2]=NULL WHERE id=1;",
					ExpectedErr:     "must not be null",
					ExpectedErrCode: "22004",
				},
			},
		},
		{
			Name: "array assignments from row expressions and subqueries",
			SetUpScript: []string{
				"CREATE TABLE t (id int PRIMARY KEY,a int[]);",
				"INSERT INTO t VALUES (1,ARRAY[[1,2,3],[4,5,6]]),(2,ARRAY[1,2,3,NULL,NULL,99]),(3,ARRAY[[7]]);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "UPDATE t SET a[id]=(SELECT 8) WHERE id=2 RETURNING a;",
					Expected: []sql.Row{{"{1,8,3,NULL,NULL,99}"}},
				},
				{
					Query:    "UPDATE t SET a[(SELECT 1):(SELECT 2)]=(SELECT ARRAY[9,10]) WHERE id=2 RETURNING a;",
					Expected: []sql.Row{{"{9,10,3,NULL,NULL,99}"}},
				},
				{
					Query:    "UPDATE t SET a[1:2]=NULL WHERE id=2 RETURNING a;",
					Expected: []sql.Row{{"{9,10,3,NULL,NULL,99}"}},
				},
				{
					Query:    "UPDATE t SET a[2:1]=NULL WHERE id=2 RETURNING a;",
					Expected: []sql.Row{{"{9,10,3,NULL,NULL,99}"}},
				},
				{
					Query:    "UPDATE t SET a[1:1][1:1]=ARRAY[NULL,99]::int[] WHERE id=1 RETURNING a;",
					Expected: []sql.Row{{"{{NULL,2,3},{4,5,6}}"}},
				},
				{
					Query:    "SELECT a FROM t WHERE id=3;",
					Expected: []sql.Row{{"{{7}}"}},
				},
				{
					Query:           "UPDATE t SET a[2:1]=ARRAY[1] WHERE id=2;",
					ExpectedErr:     "upper bound cannot be less than lower bound",
					ExpectedErrCode: "2202E",
				},

				{
					Query: "UPDATE t SET a[id]=(SELECT 8) WHERE id=2;",
				},
				{
					Query:    "SELECT a FROM t WHERE id=2;",
					Expected: []sql.Row{{"{9,8,3,NULL,NULL,99}"}},
				},
			},
		},
		{
			Name: "empty array assignments require explicit bounds",
			SetUpScript: []string{
				"CREATE TABLE t (a int[]);",
				"INSERT INTO t VALUES (ARRAY[]::int[]);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           "UPDATE t SET a[:]=ARRAY[1];",
					ExpectedErr:     "array slice subscript must provide both boundaries",
					ExpectedErrCode: "2202E",
				},
				{
					Query:    "SELECT a FROM t;",
					Expected: []sql.Row{{"{}"}},
				},
				{
					Query: "UPDATE t SET a[1:1]=NULL;",
				},
				{
					Query:    "SELECT a FROM t;",
					Expected: []sql.Row{{"{}"}},
				},
			},
		},
	})
}
