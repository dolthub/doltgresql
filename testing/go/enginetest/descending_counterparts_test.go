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

package enginetest

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dolthub/go-mysql-server/enginetest/queries"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
)

// Keep the shared scripts' data and assertions. Prefix indexes become explicit
// substring expressions; ORDER BY spells out MySQL's NULL placement in native SQL.
func descendingCounterpart(original queries.ScriptTest) queries.ScriptTest {
	s := original
	s.SetUpScript = append([]string(nil), s.SetUpScript...)
	s.Assertions = append([]queries.ScriptTestAssertion(nil), s.Assertions...)
	switch s.Name {
	case "descending index columns":
		s.SetUpScript[2] = "CREATE INDEX bc ON t(b, (left(c,10)) DESC)"
	case "descending prefix and expression indexes":
		s.SetUpScript = []string{"CREATE TABLE p(pk int PRIMARY KEY,s varchar(50),n int)", "CREATE INDEX idx1 ON p((left(s,3)) DESC)", "CREATE INDEX idx2 ON p((n * -1) DESC)", original.SetUpScript[2]}
	case "descending indexes on assorted types":
		s.SetUpScript = append([]string{
			"CREATE TYPE desc_enum AS ENUM ('a','b','c')",
			"CREATE TABLE ty(pk int PRIMARY KEY,d timestamp,dc decimal(10,2),f double precision,bt bigint,e desc_enum)",
			"CREATE INDEX dd ON ty(d DESC)", "CREATE INDEX decd ON ty(dc DESC)",
			"CREATE INDEX fd ON ty(f DESC)", "CREATE INDEX btd ON ty(bt DESC)", "CREATE INDEX ed ON ty(e DESC)",
		}, s.SetUpScript[1:]...)
	}
	s.SetUpScript = append(s.SetUpScript, "SET enable_seqscan=off")
	metadata := 0
	renamed := false
	for i := range s.Assertions {
		a := &s.Assertions[i]
		q := a.Query
		if strings.HasPrefix(q, "SHOW CREATE TABLE") || strings.HasPrefix(q, "SHOW INDEX") || strings.Contains(q, "information_schema.statistics") {
			metadata++
			table := map[string]string{"descending index columns": "t", "descending unique indexes": "u", "descending prefix and expression indexes": "p", "descending index on a keyless table": "k"}[s.Name]
			// Keys, direction flags and uniqueness are authoritative pg_index metadata.
			// A copied table gets native-generated index names, so compare definitions.
			var rows []sql.Row
			key := func(n int, flags string, unique, expr bool) sql.Row { return sql.Row{n, flags, unique, expr} }
			switch s.Name {
			case "descending index columns":
				rows = []sql.Row{key(1, "0", true, false), key(2, "3 0", false, false), key(2, "0 3", false, true)}
			case "descending unique indexes":
				rows = []sql.Row{key(1, "0", true, false), key(1, "3", true, false), key(2, "3 0", true, false)}
			case "descending index on a keyless table":
				rows = []sql.Row{key(2, "3 3", false, false)}
			case "descending prefix and expression indexes":
				rows = []sql.Row{key(1, "0", true, false)}
				if metadata <= 2 {
					rows = append(rows, key(1, "3", false, true), key(1, "3", false, true))
				} else if metadata == 3 {
					rows = append(rows, key(1, "3", false, true), key(2, "0 3", false, true))
				} else if metadata == 4 {
					rows = append(rows, key(1, "3", false, true), key(2, "0 3", false, true), key(2, "3 0", false, false), key(2, "0 3", false, false))
				} else {
					rows = append(rows, key(2, "3 0", false, false), key(2, "0 3", false, false))
					if strings.HasSuffix(q, "p2") {
						table = "p2"
					}
				}
			}
			sort.Slice(rows, func(i, j int) bool { return fmt.Sprint(rows[i]) < fmt.Sprint(rows[j]) })
			*a = queries.ScriptTestAssertion{Query: fmt.Sprintf("SELECT indnkeyatts, indoption::text, indisunique, indexprs IS NOT NULL FROM pg_index WHERE indrelid='%s'::regclass ORDER BY indnkeyatts, indoption::text, indisunique, (indexprs IS NOT NULL)", table), Expected: rows}
			continue
		}
		switch q {
		case "INSERT INTO u VALUES (7, 6, 1) ON DUPLICATE KEY UPDATE b = 9":
			a.Query = "INSERT INTO u VALUES (7,6,1) ON CONFLICT(a) DO UPDATE SET b=9"
		case "INSERT INTO u VALUES (8, 6, 2) ON DUPLICATE KEY UPDATE b = 9":
			a.Query = "INSERT INTO u VALUES (8,6,2) ON CONFLICT(a) DO UPDATE SET b=9"
			a.Expected = []sql.Row{{types.NewOkResult(1)}}
		case "REPLACE INTO u VALUES (9, 5, 7)":
			a.Query = "INSERT INTO u VALUES(9,5,7) ON CONFLICT(a) DO UPDATE SET pk=excluded.pk,b=excluded.b"
			a.Expected = []sql.Row{{types.NewOkResult(1)}}
		case "CREATE INDEX idx3 ON p ((n + 1), s(2) DESC)":
			a.Query = "CREATE INDEX idx3 ON p ((n+1),(left(s,2)) DESC)"
		case "ALTER TABLE p RENAME COLUMN n TO n2":
			a.ExpectedErr = nil
			if renamed {
				a.Query = "SELECT n2 FROM p WHERE pk=1"
				a.Expected = []sql.Row{{1}}
			} else {
				renamed = true
			}
		case "CREATE TABLE p2 LIKE p":
			a.Query = "CREATE TABLE p2 (LIKE p INCLUDING ALL)"
		}
		a.Query = descendingNullOrder(a.Query)
	}
	return s
}

func descendingNullOrder(q string) string {
	at := strings.LastIndex(q, " ORDER BY ")
	if at < 0 {
		return q
	}
	end := q[at+10:]
	suffix := ""
	if limit := strings.Index(end, " LIMIT "); limit >= 0 {
		suffix = end[limit:]
		end = end[:limit]
	}
	parts := strings.Split(end, ",")
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasSuffix(p, " DESC") {
			p += " NULLS LAST"
		} else {
			p += " NULLS FIRST"
		}
		parts[i] = p
	}
	return q[:at+10] + strings.Join(parts, ", ") + suffix
}
