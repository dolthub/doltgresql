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

// TestPredefinedDataRoles verifies the PostgreSQL provisioning workflow through
// real client connections, including dynamic objects and denied mutations.
func TestPredefinedDataRoles(t *testing.T) {
	reader := func(query string, expected ...sql.Row) ScriptTestAssertion {
		return ScriptTestAssertion{Query: query, Username: "reader", Password: "password", Expected: expected}
	}
	denied := func(user, query, message string) ScriptTestAssertion {
		return ScriptTestAssertion{Query: query, Username: user, Password: "password", ExpectedErr: message}
	}
	writer := func(query string, expected ...sql.Row) ScriptTestAssertion {
		return ScriptTestAssertion{Query: query, Username: "writer", Password: "password", Expected: expected}
	}
	RunScripts(t, []ScriptTest{
		{
			Name: "reader covers existing and future objects without write access",
			SetUpScript: []string{
				`CREATE USER reader PASSWORD 'password';`,
				`CREATE SCHEMA private;`,
				`CREATE TABLE private.data (v integer);`,
				`INSERT INTO private.data VALUES (1);`,
				`CREATE VIEW private.data_view AS SELECT v FROM private.data;`,
				`CREATE SEQUENCE private.seq;`,
			},
			Assertions: []ScriptTestAssertion{
				denied("reader", `SELECT * FROM private.data;`, "permission denied"),
				{Query: `GRANT pg_read_all_data TO reader;`},
				reader(`SELECT * FROM private.data;`, sql.Row{1}),
				reader(`SELECT * FROM private.data_view;`, sql.Row{1}),
				reader(`SELECT last_value, is_called FROM private.seq;`, sql.Row{1, "f"}),
				denied("reader", `SELECT nextval('private.seq');`, "permission denied for sequence"),
				denied("reader", `SELECT setval('private.seq', 99);`, "permission denied for sequence"),
				{Query: `SELECT nextval('private.seq');`, Expected: []sql.Row{{1}}},
				reader(`SELECT last_value, is_called FROM private.seq;`, sql.Row{1, "t"}),
				reader(`SELECT last_value FROM pg_sequences WHERE schemaname = 'private' AND sequencename = 'seq';`, sql.Row{1}),
				{Query: `CREATE SCHEMA later;`},
				{Query: `CREATE TABLE later.data (v integer);`},
				{Query: `INSERT INTO later.data VALUES (2);`},
				reader(`SELECT * FROM later.data;`, sql.Row{2}),
				{Query: `SET search_path = later;`},
				{Query: `ALTER TABLE data RENAME TO renamed;`},
				{Query: `SET search_path = public;`},
				reader(`SELECT * FROM later.renamed;`, sql.Row{2}),
				denied("reader", `INSERT INTO private.data VALUES (3);`, "permission denied"),
				denied("reader", `UPDATE private.data SET v = 3;`, "permission denied"),
				denied("reader", `DELETE FROM private.data;`, "permission denied"),
				denied("reader", `TRUNCATE private.data;`, "permission denied"),
				denied("reader", `CREATE TABLE private.forbidden (v integer);`, "permission denied"),
				denied("reader", `DROP TABLE private.data;`, "permission denied"),
				denied("reader", `GRANT SELECT ON private.data TO public;`, "does not have permission"),
				{Query: `GRANT SELECT ON private.data TO reader;`},
				{Query: `REVOKE pg_read_all_data FROM reader;`},
				reader(`SELECT * FROM private.data;`, sql.Row{1}),
				denied("reader", `SELECT * FROM later.renamed;`, "permission denied"),
			},
		},
		{
			Name: "writer grants mutation without granting reads or DDL",
			SetUpScript: []string{
				`CREATE USER writer PASSWORD 'password';`,
				`CREATE SCHEMA private;`,
				`CREATE TABLE private.data (v integer);`,
				`CREATE SEQUENCE private.seq;`,
				`CREATE TABLE private.serial_data (v serial);`,
				`GRANT pg_write_all_data TO writer;`,
			},
			Assertions: []ScriptTestAssertion{
				writer(`INSERT INTO private.data VALUES (1);`),
				denied("writer", `UPDATE private.data SET v = v + 1;`, "permission denied"),
				denied("writer", `UPDATE private.data SET v = 2 WHERE v = 1;`, "permission denied"),
				denied("writer", `UPDATE private.data SET v = 2 RETURNING v;`, "permission denied"),
				denied("writer", `INSERT INTO private.data VALUES (2) RETURNING *;`, "permission denied"),
				denied("writer", `DELETE FROM private.data WHERE v = 1;`, "permission denied"),
				denied("writer", `DELETE FROM private.data RETURNING v;`, "permission denied"),
				writer(`UPDATE private.data SET v = 2;`),
				writer(`DELETE FROM private.data;`),
				denied("writer", `SELECT * FROM private.data;`, "permission denied"),
				denied("writer", `SELECT last_value FROM private.seq;`, "permission denied"),
				writer(`SELECT nextval('private.seq');`, sql.Row{1}),
				writer(`SELECT setval('private.seq', 10);`, sql.Row{10}),
				writer(`SELECT nextval('private.seq');`, sql.Row{11}),
				writer(`INSERT INTO private.serial_data DEFAULT VALUES;`),
				writer(`SELECT last_value FROM pg_sequences WHERE schemaname = 'private' AND sequencename = 'seq';`, sql.Row{nil}),
				denied("writer", `TRUNCATE private.data;`, "permission denied"),
				denied("writer", `CREATE TABLE private.forbidden (v integer);`, "permission denied"),
				{Query: `GRANT pg_read_all_data TO writer;`},
				writer(`INSERT INTO private.data VALUES (3) RETURNING v;`, sql.Row{3}),
				writer(`UPDATE private.data SET v = v + 1 WHERE v = 3 RETURNING v;`, sql.Row{4}),
				writer(`DELETE FROM private.data WHERE v = 4 RETURNING v;`, sql.Row{4}),
				writer(`SELECT * FROM private.serial_data;`, sql.Row{1}),
				writer(`SELECT last_value FROM private.seq;`, sql.Row{11}),
			},
		},
		{
			Name: "data-role membership covers branch objects",
			SetUpScript: []string{
				`CREATE USER reader PASSWORD 'password';`,
				`GRANT pg_read_all_data TO reader;`,
				`SELECT dolt_checkout('-b', 'reporting_branch');`,
				`CREATE SCHEMA branch_private;`,
				`CREATE TABLE branch_private.data (v integer);`,
				`INSERT INTO branch_private.data VALUES (7);`,
				`SELECT dolt_commit('-Am', 'branch data');`,
				`SELECT dolt_checkout('main');`,
			},
			Assertions: []ScriptTestAssertion{
				reader(`SELECT * FROM "postgres/reporting_branch".branch_private.data;`, sql.Row{7}),
				denied("reader", `INSERT INTO "postgres/reporting_branch".branch_private.data VALUES (8);`, "permission denied"),
				{Query: `REVOKE pg_read_all_data FROM reader;`},
				denied("reader", `SELECT * FROM "postgres/reporting_branch".branch_private.data;`, "permission denied"),
			},
		},
		{
			Name: "nested membership respects NOINHERIT and alternate paths",
			SetUpScript: []string{
				`CREATE USER reader PASSWORD 'password';`,
				`CREATE ROLE intermediate;`,
				`CREATE TABLE data (v integer);`,
				`INSERT INTO data VALUES (1);`,
				`GRANT pg_read_all_data TO intermediate;`,
				`GRANT intermediate TO reader;`,
			},
			Assertions: []ScriptTestAssertion{
				reader(`SELECT * FROM data;`, sql.Row{1}),
				{Query: `ALTER ROLE intermediate NOINHERIT;`},
				denied("reader", `SELECT * FROM data;`, "permission denied"),
				{Query: `GRANT pg_read_all_data TO reader;`},
				reader(`SELECT * FROM data;`, sql.Row{1}),
				{Query: `ALTER ROLE reader NOINHERIT;`},
				denied("reader", `SELECT * FROM data;`, "permission denied"),
				{Query: `ALTER ROLE reader INHERIT;`},
				reader(`SELECT * FROM data;`, sql.Row{1}),
			},
		},
		{
			Name: "built-in role catalogs lifecycle and CREATEROLE delegation",
			SetUpScript: []string{
				`CREATE USER manager CREATEROLE PASSWORD 'password';`,
				`CREATE USER reader PASSWORD 'password';`,
				`CREATE ROLE ordinary;`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT rolname, oid::bigint, rolcanlogin, rolsuper, rolinherit, rolbypassrls FROM pg_roles WHERE rolname IN ('pg_read_all_data', 'pg_write_all_data') ORDER BY rolname;`, Expected: []sql.Row{{"pg_read_all_data", 6181, "f", "f", "t", "f"}, {"pg_write_all_data", 6182, "f", "f", "t", "f"}}},
				{Query: `CREATE ROLE pg_custom;`, ExpectedErr: "reserved", ExpectedErrCode: "42939"},
				{Query: `ALTER ROLE pg_read_all_data LOGIN;`, ExpectedErr: "reserved"},
				{Query: `DROP ROLE ordinary, pg_read_all_data;`, ExpectedErr: "required by the database system", ExpectedErrCode: "2BP01"},
				{Query: `SELECT count(*) FROM pg_roles WHERE rolname = 'ordinary';`, Expected: []sql.Row{{1}}},
				{Query: `GRANT pg_read_all_data TO reader;`, Username: "manager", Password: "password"},
				{Query: `SELECT grantor.rolname FROM pg_auth_members m JOIN pg_roles r ON m.roleid = r.oid JOIN pg_roles member ON m.member = member.oid JOIN pg_roles grantor ON m.grantor = grantor.oid WHERE r.rolname = 'pg_read_all_data' AND member.rolname = 'reader';`, Expected: []sql.Row{{"manager"}}},
				denied("reader", `GRANT pg_read_all_data TO ordinary;`, "does not have permission"),
				{Query: `REVOKE pg_read_all_data FROM reader;`, Username: "manager", Password: "password"},
				{Query: `GRANT pg_read_all_data TO reader WITH ADMIN OPTION;`, Username: "manager", Password: "password"},
				reader(`GRANT pg_read_all_data TO ordinary;`),
				{Query: `GRANT ordinary TO reader;`},
				{Query: `GRANT reader TO ordinary;`, ExpectedErr: "cycle"},
			},
		},
		{
			Name: "sequence function authorization handles bound and dynamic names",
			SetUpScript: []string{
				`CREATE USER reader PASSWORD 'password';`,
				`CREATE USER writer PASSWORD 'password';`,
				`CREATE SEQUENCE seq;`,
				`GRANT pg_read_all_data TO reader;`,
				`GRANT pg_write_all_data TO writer;`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT nextval($1::text);`, BindVars: []any{"seq"}, Username: "reader", Password: "password", ExpectedErr: "permission denied for sequence"},
				{Query: `SELECT nextval($1::text);`, BindVars: []any{"seq"}, Username: "writer", Password: "password", Expected: []sql.Row{{1}}},
				writer(`SELECT nextval('s' || 'eq');`, sql.Row{2}),
				denied("reader", `SELECT setval('s' || 'eq', 99);`, "permission denied for sequence"),
				reader(`SELECT last_value FROM seq;`, sql.Row{2}),
				{Query: `REVOKE pg_write_all_data FROM writer;`},
				{Query: `SELECT nextval($1::text);`, BindVars: []any{"seq"}, Username: "writer", Password: "password", ExpectedErr: "permission denied for sequence"},
				reader(`SELECT last_value FROM seq;`, sql.Row{2}),
			},
		},
	})
}
