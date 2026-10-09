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
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestAlterDefaultPrivileges(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "ALTER DEFAULT PRIVILEGES nonexistent role and grantee returns error",
			SetUpScript: []string{
				"CREATE ROLE ownerrole2;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "ALTER DEFAULT PRIVILEGES FOR ROLE no_such_role GRANT SELECT ON TABLES TO postgres;",
					ExpectedErr: `role "no_such_role" does not exist`,
				},
				{
					Query:       "ALTER DEFAULT PRIVILEGES FOR ROLE ownerrole2 GRANT SELECT ON TABLES TO no_such_grantee;",
					ExpectedErr: `role "no_such_grantee" does not exist`,
				},
			},
		},
		{
			Name: "ALTER DEFAULT PRIVILEGES for multiple target roles",
			SetUpScript: []string{
				`CREATE USER multi_owner1 PASSWORD 'a';`,
				`CREATE USER multi_owner2 PASSWORD 'a';`,
				`CREATE USER multi_reader PASSWORD 'a';`,
				`GRANT USAGE, CREATE ON SCHEMA public TO multi_owner1, multi_owner2;`,
				`GRANT USAGE ON SCHEMA public TO multi_reader;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `ALTER DEFAULT PRIVILEGES FOR ROLE multi_owner1, no_such_role GRANT SELECT ON TABLES TO multi_reader;`,
					ExpectedErr: `role "no_such_role" does not exist`,
				},
				{
					Query:       `ALTER DEFAULT PRIVILEGES FOR ROLE multi_owner1, multi_owner2 GRANT SELECT ON TABLES TO multi_reader;`,
					Username:    `multi_owner1`,
					Password:    `a`,
					ExpectedErr: `permission denied for multi_owner2`,
				},
				{
					Query:    `ALTER DEFAULT PRIVILEGES FOR ROLE multi_owner1, multi_owner2 GRANT SELECT ON TABLES TO multi_reader;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE multi_t1 (pk INT4 PRIMARY KEY);`,
					Username: `multi_owner1`,
					Password: `a`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE multi_t2 (pk INT4 PRIMARY KEY);`,
					Username: `multi_owner2`,
					Password: `a`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM multi_t1;`,
					Username: `multi_reader`,
					Password: `a`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM multi_t2;`,
					Username: `multi_reader`,
					Password: `a`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: `ALTER DEFAULT PRIVILEGES`,
			SetUpScript: []string{
				authTestCreateSuperUser,
				`CREATE USER readonly_user PASSWORD 'a';`,
				`GRANT USAGE ON SCHEMA public TO readonly_user;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE TABLE test (pk INT4 PRIMARY KEY);`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM test;`,
					Username:    `readonly_user`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:    `ALTER DEFAULT PRIVILEGES FOR USER auth_test_super IN SCHEMA public GRANT SELECT ON TABLES TO readonly_user;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM test;`,
					Username:    `readonly_user`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:    `CREATE TABLE another_table (pk INT4 PRIMARY KEY);`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM another_table;`,
					Username: `readonly_user`,
					Password: `a`,
					Expected: []sql.Row{},
				},
				{
					Query:       `create table user_table (i int);`,
					Username:    `readonly_user`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:    `ALTER DEFAULT PRIVILEGES FOR USER auth_test_super IN SCHEMA public REVOKE SELECT ON TABLES FROM readonly_user;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM test;`,
					Username:    `readonly_user`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
			},
		},
		{
			Name: `ALTER DEFAULT PRIVILEGES applies to new sequences`,
			SetUpScript: []string{
				authTestCreateSuperUser,
				`CREATE USER seq_reader PASSWORD 'a';`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE SEQUENCE old_seq START WITH 1;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT nextval('old_seq');`,
					Username:    `seq_reader`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:    `ALTER DEFAULT PRIVILEGES FOR USER auth_test_super IN SCHEMA public GRANT USAGE ON SEQUENCES TO seq_reader;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT nextval('old_seq');`,
					Username:    `seq_reader`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:    `CREATE SEQUENCE new_seq START WITH 10;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT nextval('new_seq');`,
					Username: `seq_reader`,
					Password: `a`,
					Expected: []sql.Row{{10}},
				},
			},
		},
		{
			Name: `ALTER DEFAULT PRIVILEGES applies to new functions`,
			SetUpScript: []string{
				authTestCreateSuperUser,
				`CREATE USER func_reader PASSWORD 'a';`,
				`GRANT USAGE ON SCHEMA public TO func_reader;`,
				// EXECUTE on functions is granted to PUBLIC by default, so it must be revoked for this test
				`ALTER DEFAULT PRIVILEGES FOR ROLE auth_test_super REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE FUNCTION old_func() RETURNS int AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT old_func();`,
					Username:    `func_reader`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:    `ALTER DEFAULT PRIVILEGES FOR USER auth_test_super IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO func_reader;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT old_func();`,
					Username:    `func_reader`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:    `CREATE FUNCTION new_func() RETURNS int AS $$ BEGIN RETURN 42; END; $$ LANGUAGE plpgsql;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT new_func();`,
					Username: `func_reader`,
					Password: `a`,
					Expected: []sql.Row{{42}},
				},
			},
		},
		{
			Name: `functions are executable by PUBLIC by default`,
			SetUpScript: []string{
				authTestCreateSuperUser,
				`CREATE USER func_caller PASSWORD 'a';`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE FUNCTION public_func() RETURNS int AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT public_func();`,
					Username: `func_caller`,
					Password: `a`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `ALTER DEFAULT PRIVILEGES FOR ROLE auth_test_super REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE FUNCTION private_func() RETURNS int AS $$ BEGIN RETURN 2; END; $$ LANGUAGE plpgsql;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT private_func();`,
					Username:    `func_caller`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:    `SELECT public_func();`,
					Username: `func_caller`,
					Password: `a`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `ALTER DEFAULT PRIVILEGES FOR ROLE auth_test_super GRANT EXECUTE ON FUNCTIONS TO PUBLIC;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE FUNCTION public_again() RETURNS int AS $$ BEGIN RETURN 3; END; $$ LANGUAGE plpgsql;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT public_again();`,
					Username: `func_caller`,
					Password: `a`,
					Expected: []sql.Row{{3}},
				},
			},
		},
		{
			Name: `DROP ROLE removes default privileges referencing the role`,
			SetUpScript: []string{
				authTestCreateSuperUser,
				`CREATE USER dp_owner PASSWORD 'a';`,
				`CREATE USER dp_grantee PASSWORD 'a';`,
				`CREATE USER dp_other PASSWORD 'a';`,
				`GRANT CREATE ON SCHEMA public TO dp_owner;`,
				`ALTER DEFAULT PRIVILEGES FOR ROLE dp_owner REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;`,
				`ALTER DEFAULT PRIVILEGES FOR ROLE dp_owner GRANT EXECUTE ON FUNCTIONS TO dp_grantee;`,
				`ALTER DEFAULT PRIVILEGES FOR ROLE dp_owner GRANT SELECT ON TABLES TO dp_grantee;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `DROP ROLE dp_grantee;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE FUNCTION dp_func() RETURNS int AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;`,
					Username: `dp_owner`,
					Password: `a`,
					Expected: []sql.Row{},
				},
				{
					// Dropping the only grantee must not undo the revoke from PUBLIC
					Query:       `SELECT dp_func();`,
					Username:    `dp_other`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:    `DROP FUNCTION dp_func();`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP ROLE dp_owner;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: `ALTER DEFAULT PRIVILEGES FOR ROLE`,
			SetUpScript: []string{
				authTestCreateSuperUser,
				`create user another_super with superuser password 'another';`,
				`CREATE USER user1 PASSWORD 'a';`,
				`CREATE TABLE test (pk INT4 PRIMARY KEY);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `SELECT * FROM test;`,
					Username:    `user1`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					// It only applies to tables created after this command is executed.
					Query:    `ALTER DEFAULT PRIVILEGES FOR ROLE auth_test_super IN SCHEMA public GRANT SELECT ON TABLES TO user1;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM test;`,
					Username:    `user1`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:    `CREATE TABLE new_table (pk INT4 PRIMARY KEY);`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM new_table;`,
					Username: `user1`,
					Password: `a`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE by_another (pk INT4 PRIMARY KEY);`,
					Username: `another_super`,
					Password: `another`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM by_another;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					// cannot select from tables created by `another` user
					Query:       `SELECT * FROM by_another;`,
					Username:    `user1`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:       `INSERT INTO test VALUES (1), (5), (6);`,
					Username:    `user1`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					// It only applies to tables created after this command is executed.
					Query:    `ALTER DEFAULT PRIVILEGES FOR ROLE auth_test_super IN SCHEMA public GRANT INSERT ON TABLES TO user1;`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:       `INSERT INTO test VALUES (1), (5), (6);`,
					Username:    `user1`,
					Password:    `a`,
					ExpectedErr: `denied`,
				},
				{
					Query:    `CREATE TABLE different_test (pk INT4 PRIMARY KEY);`,
					Username: authTestSuperUser,
					Password: authTestSuperPass,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO different_test VALUES (1), (5), (6);`,
					Username: `user1`,
					Password: `a`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM different_test;`,
					Username: `user1`,
					Password: `a`,
					Expected: []sql.Row{{1}, {5}, {6}},
				},
			},
		},
	})
}

// TestAlterDefaultPrivilegesOverlappingGrantOptions checks both ways a grant option can be supplied by overlapping defaults.
func TestAlterDefaultPrivilegesOverlappingGrantOptions(t *testing.T) {
	var scripts []ScriptTest
	for _, globalGrantOption := range []bool{true, false} {
		globalOption, schemaOption := "", " WITH GRANT OPTION"
		if globalGrantOption {
			globalOption, schemaOption = schemaOption, globalOption
		}
		script := ScriptTest{
			Name: fmt.Sprintf("overlapping defaults retain grant option global=%t", globalGrantOption),
			SetUpScript: []string{
				`CREATE USER overlap_reader PASSWORD 'a';`,
				`CREATE USER overlap_delegate PASSWORD 'a';`,
				`GRANT USAGE ON SCHEMA public TO overlap_reader, overlap_delegate;`,
				fmt.Sprintf(`ALTER DEFAULT PRIVILEGES GRANT SELECT ON TABLES TO overlap_reader%s;`, globalOption),
				fmt.Sprintf(`ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO overlap_reader%s;`, schemaOption),
			},
		}
		// Exercise several fresh objects: merging overlapping ACLs must preserve the grant option in every application order.
		for i := 0; i < 16; i++ {
			table := fmt.Sprintf("overlap_%d", i)
			script.Assertions = append(script.Assertions,
				ScriptTestAssertion{Query: fmt.Sprintf(`CREATE TABLE %s (i INT PRIMARY KEY);`, table)},
				ScriptTestAssertion{Query: fmt.Sprintf(`GRANT SELECT ON %s TO overlap_delegate;`, table), Username: "overlap_reader", Password: "a"},
				ScriptTestAssertion{Query: fmt.Sprintf(`SELECT * FROM %s;`, table), Username: "overlap_delegate", Password: "a", Expected: []sql.Row{}},
			)
		}
		scripts = append(scripts, script)
	}
	RunScripts(t, scripts)
}

// TestAlterDefaultPrivilegesGrantOptions checks that repeating a plain grant preserves an existing grant option.
func TestAlterDefaultPrivilegesGrantOptions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "repeating a plain default grant retains an existing grant option",
			SetUpScript: []string{
				`CREATE USER option_reader PASSWORD 'a';`,
				`CREATE USER option_delegate PASSWORD 'a';`,
				`GRANT USAGE ON SCHEMA public TO option_reader, option_delegate;`,
				`ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO option_reader WITH GRANT OPTION;`,
				`ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO option_reader;`,
				`CREATE TABLE option_regrant (i INT PRIMARY KEY);`,
				`INSERT INTO option_regrant VALUES (3);`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT * FROM option_regrant;`, Username: "option_reader", Password: "a", Expected: []sql.Row{{3}}},
				{Query: `GRANT SELECT ON option_regrant TO option_delegate;`, Username: "option_reader", Password: "a"},
				{Query: `SELECT * FROM option_regrant;`, Username: "option_delegate", Password: "a", Expected: []sql.Row{{3}}},
			},
		},
	})
}

// TestAlterDefaultPrivilegesCreationEdges checks that no-op creation and routine replacement preserve existing ACLs.
func TestAlterDefaultPrivilegesCreationEdges(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "CREATE TABLE IF NOT EXISTS does not apply defaults to an existing table",
			SetUpScript: []string{
				`CREATE USER noop_reader PASSWORD 'a';`,
				`GRANT USAGE ON SCHEMA public TO noop_reader;`,
				`CREATE TABLE noop_existing (i INT PRIMARY KEY);`,
				`INSERT INTO noop_existing VALUES (1);`,
				`ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO noop_reader;`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT * FROM noop_existing;`, Username: "noop_reader", Password: "a", ExpectedErr: "denied"},
				{Query: `CREATE TABLE IF NOT EXISTS noop_existing (i INT PRIMARY KEY);`},
				{Query: `SELECT * FROM noop_existing;`, Username: "noop_reader", Password: "a", ExpectedErr: "denied"},
				{Query: `SELECT * FROM noop_existing;`, Expected: []sql.Row{{1}}},
				{Query: `CREATE TABLE noop_new (i INT PRIMARY KEY);`},
				{Query: `SELECT * FROM noop_new;`, Username: "noop_reader", Password: "a", Expected: []sql.Row{}},
			},
		},
		{
			Name: "CREATE OR REPLACE FUNCTION preserves existing ACL instead of applying new defaults",
			SetUpScript: []string{
				`CREATE USER replace_reader PASSWORD 'a';`,
				`CREATE USER replace_new_reader PASSWORD 'a';`,
				`GRANT USAGE ON SCHEMA public TO replace_reader, replace_new_reader;`,
				`ALTER DEFAULT PRIVILEGES REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;`,
				`CREATE FUNCTION replace_existing() RETURNS INT AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;`,
				`GRANT EXECUTE ON FUNCTION replace_existing() TO replace_reader;`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT replace_existing();`, Username: "replace_reader", Password: "a", Expected: []sql.Row{{1}}},
				{Query: `SELECT replace_existing();`, Username: "replace_new_reader", Password: "a", ExpectedErr: "denied"},
				{Query: `ALTER DEFAULT PRIVILEGES GRANT EXECUTE ON FUNCTIONS TO replace_new_reader;`},
				{Query: `CREATE OR REPLACE FUNCTION replace_existing() RETURNS INT AS $$ BEGIN RETURN 2; END; $$ LANGUAGE plpgsql;`},
				{Query: `SELECT replace_existing();`, Username: "replace_reader", Password: "a", Expected: []sql.Row{{2}}},
				{Query: `SELECT replace_existing();`, Username: "replace_new_reader", Password: "a", ExpectedErr: "denied"},
				{Query: `CREATE FUNCTION replace_new() RETURNS INT AS $$ BEGIN RETURN 3; END; $$ LANGUAGE plpgsql;`},
				{Query: `SELECT replace_new();`, Username: "replace_new_reader", Password: "a", Expected: []sql.Row{{3}}},
				{Query: `SELECT replace_new();`, Username: "replace_reader", Password: "a", ExpectedErr: "denied"},
			},
		},
	})
}

// TestAlterDefaultPrivilegesRoutineOverloads checks that default grants apply to individual routine signatures.
func TestAlterDefaultPrivilegesRoutineOverloads(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "new function overload defaults do not change an existing overload ACL",
			SetUpScript: []string{
				`CREATE USER overload_reader PASSWORD 'a';`,
				`GRANT USAGE ON SCHEMA public TO overload_reader;`,
				`ALTER DEFAULT PRIVILEGES REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;`,
				`CREATE FUNCTION overload_acl(i INT) RETURNS INT AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT overload_acl(1);`, Username: "overload_reader", Password: "a", ExpectedErr: "denied"},
				{Query: `ALTER DEFAULT PRIVILEGES GRANT EXECUTE ON FUNCTIONS TO overload_reader;`},
				{Query: `CREATE FUNCTION overload_acl(i TEXT) RETURNS INT AS $$ BEGIN RETURN 2; END; $$ LANGUAGE plpgsql;`},
				{Query: `SELECT overload_acl('x'::TEXT);`, Username: "overload_reader", Password: "a", Expected: []sql.Row{{2}}},
				{Query: `SELECT overload_acl(1);`, Username: "overload_reader", Password: "a", ExpectedErr: "denied"},
			},
		},
	})
}
