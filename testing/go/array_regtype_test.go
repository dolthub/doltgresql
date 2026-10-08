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

// TestArrayRegtype checks that all array ranks resolve to the same element array type.
func TestArrayRegtype(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name:        "array dimensionality does not change regtype",
		SetUpScript: []string{"CREATE TABLE array_type_names (id int PRIMARY KEY,name text);", "INSERT INTO array_type_names VALUES (1,'integer[][]'),(2,'varchar[][][]');"},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT 'integer[]'::regtype='integer[][]'::regtype,'varchar[][][]'::regtype='varchar[]'::regtype;",
				Expected: []sql.Row{{"t", "t"}},
			},
			{Query: "SELECT 'pg_catalog.int4[][]'::regtype;", Expected: []sql.Row{{"integer[]"}}},
			{
				Query:    "SELECT name::regtype FROM array_type_names ORDER BY id;",
				Expected: []sql.Row{{"integer[]"}, {"character varying[]"}},
			},
			{
				Query:    "SELECT (SELECT 'pg_catalog.int4[][]')::regtype;",
				Expected: []sql.Row{{"integer[]"}},
			},
		},
	}})
}
