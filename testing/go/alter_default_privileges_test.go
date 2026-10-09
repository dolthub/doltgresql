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
