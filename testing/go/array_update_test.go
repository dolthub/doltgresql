// Copyright 2025 Dolthub, Inc.
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
	"github.com/dolthub/go-mysql-server/sql"
	"testing"
)

func TestArraySubscriptUpdate(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name:        "array subscript updates",
		SetUpScript: []string{"CREATE TABLE t (id int PRIMARY KEY, a int[]);", "INSERT INTO t VALUES (1,ARRAY[[1,2,3],[4,5,6]]),(2,ARRAY[1,2,3]),(3,NULL);"},
		Assertions: []ScriptTestAssertion{
			{Query: "UPDATE t SET a[2][3]=60 WHERE id=1 RETURNING a;", Expected: []sql.Row{{"{{1,2,3},{4,5,60}}"}}},
			{Query: "UPDATE t SET a[1:2][2:3]=ARRAY[[20,30],[50,60]] WHERE id=1 RETURNING a;", Expected: []sql.Row{{"{{1,20,30},{4,50,60}}"}}},
			{Query: "UPDATE t SET a[6]=99 WHERE id=2 RETURNING a;", Expected: []sql.Row{{"{1,2,3,NULL,NULL,99}"}}},
			{Query: "UPDATE t SET a[1][1]=7 WHERE id=3 RETURNING a;", Expected: []sql.Row{{"{{7}}"}}},
			{Query: "UPDATE t SET a[:][2:3]=ARRAY[[2,3],[5,6]] WHERE id=1 RETURNING a;", Expected: []sql.Row{{"{{1,2,3},{4,5,6}}"}}},
			{Query: "UPDATE t SET a[3][1]=7 WHERE id=1;", ExpectedErr: "array subscript out of range", ExpectedErrCode: "2202E"},
			{Query: "UPDATE t SET a[1:2][1:2]=ARRAY[[9,8]] WHERE id=1;", ExpectedErr: "source array too small", ExpectedErrCode: "2202E"},
			{Query: "UPDATE t SET a[NULL]=7 WHERE id=2;", ExpectedErr: "must not be null", ExpectedErrCode: "22004"},
			{Query: "SELECT a FROM t WHERE id=1;", Expected: []sql.Row{{"{{1,2,3},{4,5,6}}"}}},
		},
	}})
}
