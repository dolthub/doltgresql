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
	"net/netip"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestInet(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "inet input and output",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '192.168.1.1'::inet, '192.168.1.1/24'::inet, '::1'::inet, '::ffff:1.2.3.4'::inet, '10/8'::inet;`,
					Expected: []sql.Row{{"192.168.1.1", "192.168.1.1/24", "::1", "::ffff:1.2.3.4", "10.0.0.0/8"}},
				},
				{
					Query:    `SELECT '1.2.3.4'::inet::text, '1.2.3.4/24'::inet::varchar, '::1'::inet::text, '1.2.3.4'::text::inet;`,
					Expected: []sql.Row{{"1.2.3.4/32", "1.2.3.4/24", "::1/128", "1.2.3.4"}},
				},
				{
					Query:    `SELECT pg_typeof('1.2.3.4'::inet), pg_typeof(ARRAY['1.2.3.4'::inet]), ARRAY['1.2.3.4', '::1/64']::inet[];`,
					Expected: []sql.Row{{"inet", "inet[]", "{1.2.3.4,::1/64}"}},
				},
				{
					Query:       `SELECT 'foo'::inet;`,
					ExpectedErr: `invalid input syntax for type inet: "foo"`,
				},
				{
					Query:       `SELECT '1.2.3.4/33'::inet;`,
					ExpectedErr: `invalid input syntax for type inet: "1.2.3.4/33"`,
				},
				{
					Query:       `SELECT '10'::inet;`,
					ExpectedErr: `invalid input syntax for type inet: "10"`,
				},
			},
		},
		{
			Name: "inet functions",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT text('1.2.3.4/24'::inet), abbrev('1.2.3.4/24'::inet), abbrev('1.2.3.4'::inet), host('1.2.3.4/24'::inet), masklen('1.2.3.4/24'::inet), family('::1'::inet), family('1.2.3.4'::inet);`,
					Expected: []sql.Row{{"1.2.3.4/24", "1.2.3.4/24", "1.2.3.4", "1.2.3.4", 24, 6, 4}},
				},
				{
					Query:    `SELECT netmask('1.2.3.4/24'::inet), hostmask('1.2.3.4/24'::inet), broadcast('1.2.3.4/24'::inet), set_masklen('1.2.3.4/24'::inet, 16), set_masklen('1.2.3.4/24'::inet, -1);`,
					Expected: []sql.Row{{"255.255.255.0", "0.0.0.255", "1.2.3.255/24", "1.2.3.4/16", "1.2.3.4"}},
				},
				{
					Query:    `SELECT network('192.168.1.5/24'::inet), network('10.1.2.3/8'::inet), network('1.2.3.4'::inet), network('0.0.0.0/0'::inet);`,
					Expected: []sql.Row{{"192.168.1.0/24", "10.0.0.0/8", "1.2.3.4", "0.0.0.0/0"}},
				},
				{
					Query:    `SELECT network('2001:db8:1234:5678::1/48'::inet), network('2001:db8::ff/120'::inet), network('::1'::inet);`,
					Expected: []sql.Row{{"2001:db8:1234::/48", "2001:db8::/120", "::1"}},
				},
				{
					Query:    `SELECT inet_same_family('1.2.3.4', '::1'), inet_same_family('1.2.3.4', '5.6.7.8'), inetnot('1.2.3.4');`,
					Expected: []sql.Row{{"f", "t", "254.253.252.251"}},
				},
				{
					Query:       `SELECT set_masklen('1.2.3.4'::inet, 40);`,
					ExpectedErr: "invalid mask length: 40",
				},
			},
		},
		{
			Name: "inet operators",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '1.2.3.4'::inet + 300, '1.2.3.4'::inet - 1, '1.2.3.4'::inet - '1.2.3.0'::inet, 5 + '1.2.3.4'::inet;`,
					Expected: []sql.Row{{"1.2.4.48", "1.2.3.3", 4, "1.2.3.9"}},
				},
				{
					Query:    `SELECT '1.2.3.4'::inet & '255.255.0.0'::inet, '1.2.3.4'::inet | '0.0.255.255'::inet;`,
					Expected: []sql.Row{{"1.2.0.0", "1.2.255.255"}},
				},
				{
					Query:    `SELECT '1.2.3.4'::inet << '1.2.0.0/16'::inet, '1.2.0.0/16'::inet <<= '1.2.0.0/16'::inet, '1.2.0.0/16'::inet >> '1.2.3.4'::inet, '1.2.0.0/16'::inet >>= '1.2.3.4'::inet, '1.2.0.0/16'::inet && '1.0.0.0/8'::inet;`,
					Expected: []sql.Row{{"t", "t", "t", "t", "t"}},
				},
				{
					Query:    `SELECT '1.2.0.0/16'::inet << '1.2.0.0/16'::inet, '1.2.3.4'::inet << '::/0'::inet, '1.2.3.4'::inet && '::/0'::inet;`,
					Expected: []sql.Row{{"f", "f", "f"}},
				},
				{
					Query:    `SELECT '1.2.3.4/24'::inet = '1.2.3.4/24', '1.2.3.4'::inet = '1.2.3.4/32', '1.2.3.4'::inet <> '1.2.3.4/24', '1.2.3.4'::inet < '::1'::inet, '10.0.0.1/8'::inet > '9.0.0.0/32'::inet;`,
					Expected: []sql.Row{{"t", "t", "t", "t", "t"}},
				},
				{
					Query:       `SELECT '255.255.255.255'::inet + 1;`,
					ExpectedErr: "result is out of range",
				},
				{
					Query:       `SELECT '1.2.3.4'::inet & '::1'::inet;`,
					ExpectedErr: "cannot AND inet values of different sizes",
				},
				{
					Query:       `SELECT '1.2.3.4'::inet - '::1'::inet;`,
					ExpectedErr: "cannot subtract inet values of different sizes",
				},
			},
		},
		{
			Name: "inet table",
			SetUpScript: []string{
				`CREATE TABLE inet_t (id INT PRIMARY KEY, ip INET);`,
				`INSERT INTO inet_t VALUES (1, '10.0.0.1/8'), (2, '9.0.0.0/32'), (3, '10.0.0.0/8'), (4, '10.0.0.0/16'), (5, '::1'), (6, '10.0.0.0/32'), (7, NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{ // Matches the Postgres ordering, which compares the network part before the mask length
					Query:    `SELECT id, ip FROM inet_t ORDER BY ip, id;`,
					Expected: []sql.Row{{2, "9.0.0.0"}, {3, "10.0.0.0/8"}, {1, "10.0.0.1/8"}, {4, "10.0.0.0/16"}, {6, "10.0.0.0"}, {5, "::1"}, {7, nil}},
				},
				{
					Query:    `SELECT id FROM inet_t WHERE ip << '10.0.0.0/8' ORDER BY id;`,
					Expected: []sql.Row{{4}, {6}},
				},
				{
					Query:    `SELECT family(ip), count(*) FROM inet_t WHERE ip IS NOT NULL GROUP BY family(ip) ORDER BY 1;`,
					Expected: []sql.Row{{4, 5}, {6, 1}},
				},
				{
					Query:    `CREATE INDEX inet_t_idx ON inet_t (ip);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT id FROM inet_t WHERE ip = '10.0.0.0/8';`,
					Expected: []sql.Row{{3}},
				},
				{
					Query:    `SELECT id FROM inet_t WHERE ip > '10.0.0.1/8' AND ip < '::1' ORDER BY ip;`,
					Expected: []sql.Row{{4}, {6}},
				},
				{
					Query:    `UPDATE inet_t SET ip = ip + 1 WHERE id = 2;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT ip FROM inet_t WHERE id = 2;`,
					Expected: []sql.Row{{"9.0.0.1"}},
				},
			},
		},
		{
			Name: "inet primary key",
			SetUpScript: []string{
				`CREATE TABLE inet_pk (ip INET PRIMARY KEY, v TEXT);`,
				`INSERT INTO inet_pk VALUES ('::1', 'a'), ('1.2.3.4/24', 'b'), ('1.2.3.4', 'c');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM inet_pk;`,
					Expected: []sql.Row{{"1.2.3.4/24", "b"}, {"1.2.3.4", "c"}, {"::1", "a"}},
				},
				{
					Query:       `INSERT INTO inet_pk VALUES ('1.2.3.4/32', 'd');`,
					ExpectedErr: "duplicate",
				},
				{ // Bound parameters use the binary wire format
					Query:    `INSERT INTO inet_pk VALUES ($1, 'e'), ($2, 'f');`,
					BindVars: []any{netip.MustParsePrefix("10.1.2.3/8"), netip.MustParsePrefix("2001:db8::1/64")},
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT ip, v FROM inet_pk WHERE ip = $1 OR ip = $2 ORDER BY ip;`,
					BindVars: []any{netip.MustParsePrefix("10.1.2.3/8"), netip.MustParsePrefix("2001:db8::1/64")},
					Expected: []sql.Row{{"10.1.2.3/8", "e"}, {"2001:db8::1/64", "f"}},
				},
				{
					Query:    `SELECT host(ip), masklen(ip) FROM inet_pk WHERE v = 'f';`,
					Expected: []sql.Row{{"2001:db8::1", 64}},
				},
			},
		},
		{
			Name: "inet catalog",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT typname, typcategory, typispreferred, typlen, typalign, typstorage FROM pg_type WHERE typname IN ('inet', '_inet') ORDER BY typname;`,
					Expected: []sql.Row{{"_inet", "A", "f", -1, "i", "x"}, {"inet", "I", "t", -1, "i", "m"}},
				},
				{
					Query:    `SELECT am.amname, c.opcname, f.opfname, c.opcdefault FROM pg_opclass c JOIN pg_am am ON am.oid = c.opcmethod JOIN pg_opfamily f ON f.oid = c.opcfamily WHERE c.opcintype = 'inet'::regtype ORDER BY 1;`,
					Expected: []sql.Row{{"btree", "inet_ops", "network_ops", "t"}, {"hash", "inet_ops", "network_ops", "t"}},
				},
			},
		},
	})
}
