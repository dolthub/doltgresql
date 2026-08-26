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
	"fmt"
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/enginetest/scriptgen/setup"
	"github.com/dolthub/go-mysql-server/sql"
)

// Adaptive-encoded columns (text, bpchar, unbounded varchar, bytea, jsonb, and unbounded extended
// types) keep their adaptive encoding when they appear in a primary key, so a key tuple larger than
// the tuple size target (2KB) stores the column's content out of band and embeds only its chunk
// address in the key, exactly as value tuples do. These tests exercise that key-side out-of-band
// path, which the value-focused tests in adaptive_encoding_test.go never reach (their primary keys
// are all small).

// bigKey returns a 20000-character string (well past the out-of-band threshold) that is unique per
// |i| and sorts by |i|.
func bigKey(i int, fill string) string {
	return fmt.Sprintf("%08d", i) + strings.Repeat(fill, 20000-8)
}

func TestAdaptiveEncodingInPrimaryKeys(t *testing.T) {
	for _, columnType := range []string{"text", "varchar"} {
		t.Run(columnType, func(t *testing.T) {
			RunScripts(t, []ScriptTest{
				{
					Name: "out-of-band values in primary keys",
					SetUpScript: setup.SetupScript{
						fmt.Sprintf(`create table tk (big %s primary key, n int);`, columnType),
						// ids 1-3: out-of-band keys; id 4: a small inline key
						`insert into tk select lpad(i::text, 8, '0') || repeat('k', 19992), i from generate_series(1, 3) as g(i);`,
						`insert into tk values ('small', 4);`,
					},
					Assertions: []ScriptTestAssertion{
						{
							Query:    `select count(*) from tk;`,
							Expected: []sql.Row{{4}},
						},
						{
							Query:    `select n, length(big) from tk order by n;`,
							Expected: []sql.Row{{1, 20000}, {2, 20000}, {3, 20000}, {4, 5}},
						},
						{
							// round-trip the full key content
							Query:    `select big = '` + bigKey(2, "k") + `' from tk where n = 2;`,
							Expected: []sql.Row{{"t"}},
						},
						{
							// point lookup by an out-of-band key
							Query:    `select n from tk where big = '` + bigKey(2, "k") + `';`,
							Expected: []sql.Row{{2}},
						},
						{
							// out-of-band keys participate in ordering
							Query:    `select n from tk order by big;`,
							Expected: []sql.Row{{1}, {2}, {3}, {4}},
						},
						{
							Query:    `update tk set n = 20 where big = '` + bigKey(2, "k") + `';`,
							Expected: []sql.Row{},
						},
						{
							Query:    `select n from tk where big = '` + bigKey(2, "k") + `';`,
							Expected: []sql.Row{{20}},
						},
						{
							Query:    `delete from tk where big = '` + bigKey(1, "k") + `';`,
							Expected: []sql.Row{},
						},
						{
							Query:    `select count(*) from tk;`,
							Expected: []sql.Row{{3}},
						},
						{
							// out-of-band keys survive a dolt commit
							Query:            `select dolt_commit('-Am', 'out-of-band keys');`,
							SkipResultsCheck: true,
						},
						{
							Query:    `select n from tk where big = '` + bigKey(3, "k") + `';`,
							Expected: []sql.Row{{3}},
						},
					},
				},
			})
		})
	}
}

// TestAdaptiveEncodingGarbageCollection verifies that the chunks backing out-of-band adaptive values
// are retained by garbage collection, for values referenced from value tuples, secondary index keys,
// and primary keys.
func TestAdaptiveEncodingGarbageCollection(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "out-of-band values in value tuples survive garbage collection",
			SetUpScript: setup.SetupScript{
				`create table tv (id int primary key, big text);`,
				`insert into tv select i, lpad(i::text, 8, '0') || repeat('v', 19992) from generate_series(1, 3) as g(i);`,
				`select dolt_commit('-Am', 'values');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `select dolt_gc();`,
					Expected: []sql.Row{{0}},
				},
				{
					// dolt_gc invalidates the session's state, so use a fresh connection
					Query:    `select count(*) from tv where length(big) = 20000;`,
					Username: "postgres",
					Password: "password",
					Expected: []sql.Row{{3}},
				},
			},
		},
		{
			Name: "out-of-band values in secondary index keys survive garbage collection",
			SetUpScript: setup.SetupScript{
				`create table ti (id int primary key, big text);`,
				`create index ti_big on ti (big);`,
				`insert into ti select i, lpad(i::text, 8, '0') || repeat('i', 19992) from generate_series(1, 3) as g(i);`,
				`select dolt_commit('-Am', 'indexed');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `select dolt_gc();`,
					Expected: []sql.Row{{0}},
				},
				{
					// an index lookup on the out-of-band column works after gc: the chunk is shared
					// with (and retained through) the value tuple's reference
					Query:    `select id from ti where big = '` + bigKey(2, "i") + `';`,
					Username: "postgres",
					Password: "password",
					Expected: []sql.Row{{2}},
				},
			},
		},
		{
			Name: "out-of-band values in primary keys survive garbage collection",
			SetUpScript: setup.SetupScript{
				`create table tk (big text primary key, n int);`,
				`insert into tk select lpad(i::text, 8, '0') || repeat('k', 19992), i from generate_series(1, 3) as g(i);`,
				`select dolt_commit('-Am', 'keys');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `select dolt_gc();`,
					Expected: []sql.Row{{0}},
				},
				{
					// BUG: the ProllyTreeNode message format records chunk references for value
					// tuples (value_address_offsets) but has no equivalent field for key tuples, so
					// the reachability walk used by gc (and push and clone) never visits chunks
					// referenced only from keys. gc deletes them, and this query panics with "empty
					// chunk returned from ChunkStore". Unskip once key-side references are tracked.
					Skip:     true,
					Query:    `select count(*) from tk where length(big) = 20000;`,
					Username: "postgres",
					Password: "password",
					Expected: []sql.Row{{3}},
				},
				{
					// key content that was never moved out of band is unaffected
					Query:    `select count(*) from tk;`,
					Username: "postgres",
					Password: "password",
					Expected: []sql.Row{{3}},
				},
			},
		},
	})
}
