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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
)

// Native index catalogs expose their actual AM, casts and operator classes.
// Keep these expectations separate from Doltgres; never derive them from results.
func indexCounterparts(scripts []ScriptTest) []ScriptTest {
	for i := range scripts {
		for j := range scripts[i].Assertions {
			a := &scripts[i].Assertions[j]
			if strings.Contains(a.Query, "array_to_string(indclass") {
				// Resolve classes by their declared input type / name, not installation OIDs.
				name := strings.Split(strings.Split(a.Query, "indexrelid = '")[1], "'")[0]
				query := "SELECT string_agg(c.opcintype::regtype::text || ':' || regexp_replace(c.opcname, '^dolt_', ''), ',' ORDER BY k.n) FROM pg_index i CROSS JOIN LATERAL unnest(i.indclass) WITH ORDINALITY k(oid,n) JOIN pg_opclass c ON c.oid=k.oid WHERE i.indexrelid='" + name + "'::regclass"
				expected := "text:varchar_pattern_ops"
				switch name {
				case "index_warning_newest":
					expected = "bigint:int8_ops,timestamp with time zone:timestamptz_ops,bigint:int8_ops"
				case "ok_expr":
					expected = "text:text_pattern_ops"
				}
				if strings.Contains(a.Query, "indoption") {
					query = "SELECT array_to_string(indoption, ',') || ' ' || (" + query + ") FROM pg_index WHERE indexrelid='" + name + "'::regclass"
					option := "0"
					if name == "ex_pattern" {
						option = "3"
					}
					expected = option + " " + expected
				}
				a.Postgres = &ScriptTestAssertion{Query: query, Expected: []sql.Row{{expected}}}
			}
			if strings.Contains(a.Query, "indexdef") || strings.Contains(a.Query, "pg_get_indexdef") {
				native := *a
				if a.Postgres != nil {
					native = *a.Postgres
					if native.Query == "" {
						native.Query = a.Query
					}
				}
				native.Postgres, native.PostgresFork = nil, nil
				originalRows := native.Expected
				native.Expected = make([]sql.Row, len(originalRows))
				for r, row := range originalRows {
					native.Expected[r] = append(sql.Row(nil), row...)
					for c, v := range row {
						if text, ok := v.(string); ok {
							// varchar -> text is shown explicitly by PostgreSQL's deparser.
							text = strings.ReplaceAll(text, "(lower(name))", "lower((name)::text)")
							text = strings.ReplaceAll(text, "(upper(name))", "upper(name)")
							text = strings.ReplaceAll(text, "WHERE (user_sessions.is_active", "WHERE (is_active")
							text = strings.ReplaceAll(text, "WHERE (t.a", "WHERE (a")
							if strings.Contains(text, "upper(name), age") {
								text = strings.ReplaceAll(text, "(c1 + c2)", "((c1 + c2))")
							}
							native.Expected[r][c] = text
						}
					}
				}
				fork := native
				fork.Expected = make([]sql.Row, len(native.Expected))
				for r, row := range native.Expected {
					fork.Expected[r] = append(sql.Row(nil), row...)
					for c, v := range row {
						if text, ok := v.(string); ok {
							text = strings.ReplaceAll(text, "USING btree", "USING dolt_prolly_index")
							if strings.HasPrefix(text, "CREATE INDEX ok_default ") {
								text = strings.ReplaceAll(text, "(name)", "(name text_ops)")
							}
							if strings.HasPrefix(text, "CREATE INDEX ok_expr_default ") {
								text = strings.ReplaceAll(text, "(lower((name)::text))", "(lower((name)::text) text_ops)")
							}
							fork.Expected[r][c] = text
						}
					}
				}
				a.Postgres, a.PostgresFork = &native, &fork
			}
			if strings.Contains(a.ExpectedErr, "does not exist") && strings.HasPrefix(a.Query, "ALTER INDEX") {
				a.Postgres = &ScriptTestAssertion{ExpectedErrCode: "42P01"}
			}
			if strings.Contains(a.ExpectedErr, `access method "btree"`) {
				a.PostgresFork = &ScriptTestAssertion{ExpectedErr: strings.ReplaceAll(a.ExpectedErr, `access method "btree"`, `access method "dolt_prolly_index"`)}
			}
			if strings.HasPrefix(a.Query, "SELECT dolt_commit") {
				a.Postgres = &ScriptTestAssertion{Skip: true}
				copy := *a
				copy.Postgres = nil
				a.PostgresFork = &copy
			}
		}
	}
	return scripts
}
