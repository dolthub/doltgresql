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

func TestRoleattributes(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_roleattributes)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_roleattributes,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE ROLE regress_test_def_superuser;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_def_superuser';`,
				Results:   []sql.Row{{`regress_test_def_superuser`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `CREATE ROLE regress_test_superuser WITH SUPERUSER;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_superuser';`,
				Results:   []sql.Row{{`regress_test_superuser`, true, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `ALTER ROLE regress_test_superuser WITH NOSUPERUSER;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_superuser';`,
				Results:   []sql.Row{{`regress_test_superuser`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `ALTER ROLE regress_test_superuser WITH SUPERUSER;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_superuser';`,
				Results:   []sql.Row{{`regress_test_superuser`, true, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `CREATE ROLE regress_test_def_inherit;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_def_inherit';`,
				Results:   []sql.Row{{`regress_test_def_inherit`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `CREATE ROLE regress_test_inherit WITH NOINHERIT;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_inherit';`,
				Results:   []sql.Row{{`regress_test_inherit`, false, false, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `ALTER ROLE regress_test_inherit WITH INHERIT;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_inherit';`,
				Results:   []sql.Row{{`regress_test_inherit`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `ALTER ROLE regress_test_inherit WITH NOINHERIT;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_inherit';`,
				Results:   []sql.Row{{`regress_test_inherit`, false, false, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `CREATE ROLE regress_test_def_createrole;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_def_createrole';`,
				Results:   []sql.Row{{`regress_test_def_createrole`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `CREATE ROLE regress_test_createrole WITH CREATEROLE;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_createrole';`,
				Results:   []sql.Row{{`regress_test_createrole`, false, true, true, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `ALTER ROLE regress_test_createrole WITH NOCREATEROLE;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_createrole';`,
				Results:   []sql.Row{{`regress_test_createrole`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `ALTER ROLE regress_test_createrole WITH CREATEROLE;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_createrole';`,
				Results:   []sql.Row{{`regress_test_createrole`, false, true, true, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `CREATE ROLE regress_test_def_createdb;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_def_createdb';`,
				Results:   []sql.Row{{`regress_test_def_createdb`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `CREATE ROLE regress_test_createdb WITH CREATEDB;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_createdb';`,
				Results:   []sql.Row{{`regress_test_createdb`, false, true, false, true, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `ALTER ROLE regress_test_createdb WITH NOCREATEDB;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_createdb';`,
				Results:   []sql.Row{{`regress_test_createdb`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `ALTER ROLE regress_test_createdb WITH CREATEDB;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_createdb';`,
				Results:   []sql.Row{{`regress_test_createdb`, false, true, false, true, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `CREATE ROLE regress_test_def_role_canlogin;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_def_role_canlogin';`,
				Results:   []sql.Row{{`regress_test_def_role_canlogin`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `CREATE ROLE regress_test_role_canlogin WITH LOGIN;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_role_canlogin';`,
				Results:   []sql.Row{{`regress_test_role_canlogin`, false, true, false, false, true, false, false, -1, ``, ``}},
			},
			{
				Statement: `ALTER ROLE regress_test_role_canlogin WITH NOLOGIN;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_role_canlogin';`,
				Results:   []sql.Row{{`regress_test_role_canlogin`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `ALTER ROLE regress_test_role_canlogin WITH LOGIN;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_role_canlogin';`,
				Results:   []sql.Row{{`regress_test_role_canlogin`, false, true, false, false, true, false, false, -1, ``, ``}},
			},
			{
				Statement: `CREATE USER regress_test_def_user_canlogin;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_def_user_canlogin';`,
				Results:   []sql.Row{{`regress_test_def_user_canlogin`, false, true, false, false, true, false, false, -1, ``, ``}},
			},
			{
				Statement: `CREATE USER regress_test_user_canlogin WITH NOLOGIN;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_user_canlogin';`,
				Results:   []sql.Row{{`regress_test_user_canlogin`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `ALTER USER regress_test_user_canlogin WITH LOGIN;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_user_canlogin';`,
				Results:   []sql.Row{{`regress_test_user_canlogin`, false, true, false, false, true, false, false, -1, ``, ``}},
			},
			{
				Statement: `ALTER USER regress_test_user_canlogin WITH NOLOGIN;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_user_canlogin';`,
				Results:   []sql.Row{{`regress_test_user_canlogin`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `CREATE ROLE regress_test_def_replication;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_def_replication';`,
				Results:   []sql.Row{{`regress_test_def_replication`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `CREATE ROLE regress_test_replication WITH REPLICATION;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_replication';`,
				Results:   []sql.Row{{`regress_test_replication`, false, true, false, false, false, true, false, -1, ``, ``}},
			},
			{
				Statement: `ALTER ROLE regress_test_replication WITH NOREPLICATION;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_replication';`,
				Results:   []sql.Row{{`regress_test_replication`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `ALTER ROLE regress_test_replication WITH REPLICATION;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_replication';`,
				Results:   []sql.Row{{`regress_test_replication`, false, true, false, false, false, true, false, -1, ``, ``}},
			},
			{
				Statement: `CREATE ROLE regress_test_def_bypassrls;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_def_bypassrls';`,
				Results:   []sql.Row{{`regress_test_def_bypassrls`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `CREATE ROLE regress_test_bypassrls WITH BYPASSRLS;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_bypassrls';`,
				Results:   []sql.Row{{`regress_test_bypassrls`, false, true, false, false, false, false, true, -1, ``, ``}},
			},
			{
				Statement: `ALTER ROLE regress_test_bypassrls WITH NOBYPASSRLS;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_bypassrls';`,
				Results:   []sql.Row{{`regress_test_bypassrls`, false, true, false, false, false, false, false, -1, ``, ``}},
			},
			{
				Statement: `ALTER ROLE regress_test_bypassrls WITH BYPASSRLS;`,
			},
			{
				Statement: `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls, rolconnlimit, rolpassword, rolvaliduntil FROM pg_authid WHERE rolname = 'regress_test_bypassrls';`,
				Results:   []sql.Row{{`regress_test_bypassrls`, false, true, false, false, false, false, true, -1, ``, ``}},
			},
			{
				Statement: `DROP ROLE regress_test_def_superuser;`,
			},
			{
				Statement: `DROP ROLE regress_test_superuser;`,
			},
			{
				Statement: `DROP ROLE regress_test_def_inherit;`,
			},
			{
				Statement: `DROP ROLE regress_test_inherit;`,
			},
			{
				Statement: `DROP ROLE regress_test_def_createrole;`,
			},
			{
				Statement: `DROP ROLE regress_test_createrole;`,
			},
			{
				Statement: `DROP ROLE regress_test_def_createdb;`,
			},
			{
				Statement: `DROP ROLE regress_test_createdb;`,
			},
			{
				Statement: `DROP ROLE regress_test_def_role_canlogin;`,
			},
			{
				Statement: `DROP ROLE regress_test_role_canlogin;`,
			},
			{
				Statement: `DROP USER regress_test_def_user_canlogin;`,
			},
			{
				Statement: `DROP USER regress_test_user_canlogin;`,
			},
			{
				Statement: `DROP ROLE regress_test_def_replication;`,
			},
			{
				Statement: `DROP ROLE regress_test_replication;`,
			},
			{
				Statement: `DROP ROLE regress_test_def_bypassrls;`,
			},
			{
				Statement: `DROP ROLE regress_test_bypassrls;`,
			},
		},
	})
}
