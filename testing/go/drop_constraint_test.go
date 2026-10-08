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

// TestDropConstraintWithExpressionIndex covers https://github.com/dolthub/doltgresql/issues/3524
// and constraint types affected by the same table wrapper.
func TestDropConstraintWithExpressionIndex(t *testing.T) {
	var scripts []ScriptTest
	for _, drop := range []struct {
		name  string
		query string
	}{
		{"unique constraint", "ALTER TABLE users DROP CONSTRAINT users_email_key;"},
		{"unique constraint if exists", "ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;"},
		{
			"unique constraint in Auth migration DO block",
			`DO $$
BEGIN
  ALTER TABLE ONLY public.users DROP CONSTRAINT IF EXISTS users_email_key;
EXCEPTION
  WHEN SQLSTATE '2BP01' THEN
    RAISE NOTICE 'Unable to drop users_email_key constraint due to dependent objects';
END $$;`,
		},
	} {
		scripts = append(scripts, ScriptTest{
			Name: drop.name,
			SetUpScript: []string{
				"CREATE TABLE users (email text UNIQUE);",
				"CREATE INDEX users_lower_email_idx ON users (lower(email));",
				"INSERT INTO users VALUES ('Alice@example.com');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    drop.query,
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT conname FROM pg_constraint WHERE conrelid = 'users'::regclass AND conname = 'users_email_key';",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE schemaname = 'public' AND tablename = 'users' ORDER BY indexname;",
					Expected: []sql.Row{{"users_lower_email_idx"}},
				},
				{
					// Success must remove uniqueness enforcement, not just return an ALTER TABLE tag.
					Query:    "INSERT INTO users VALUES ('Alice@example.com'), ('alice@example.com');",
					Expected: []sql.Row{},
				},
				{
					// Exercise the retained expression index after writing duplicate email values.
					Query:    "SELECT email FROM users WHERE lower(email) = 'alice@example.com' ORDER BY email;",
					Expected: []sql.Row{{"Alice@example.com"}, {"Alice@example.com"}, {"alice@example.com"}},
				},
			},
		})
	}
	scripts = append(scripts,
		ScriptTest{
			Name: "missing constraint preserves existing constraints and indexes",
			SetUpScript: []string{
				"CREATE TABLE users (email text UNIQUE);",
				"CREATE INDEX users_lower_email_idx ON users (lower(email));",
				"INSERT INTO users VALUES ('alice@example.com');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           "ALTER TABLE users DROP CONSTRAINT missing_constraint;",
					ExpectedErr:     "does not exist",
					ExpectedErrCode: "42704",
				},
				{
					Query:    "ALTER TABLE users DROP CONSTRAINT IF EXISTS missing_constraint;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE schemaname = 'public' AND tablename = 'users' ORDER BY indexname;",
					Expected: []sql.Row{{"users_email_key"}, {"users_lower_email_idx"}},
				},
				{
					Query:           "INSERT INTO users VALUES ('alice@example.com');",
					ExpectedErr:     "duplicate unique key",
					ExpectedErrCode: "23505",
				},
			},
		},
		ScriptTest{
			Name: "check constraint",
			SetUpScript: []string{
				"CREATE TABLE users (email text, CONSTRAINT users_email_check CHECK (email <> 'blocked'));",
				"CREATE INDEX users_lower_email_idx ON users (lower(email));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE users DROP CONSTRAINT users_email_check;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT conname FROM pg_constraint WHERE conrelid = 'users'::regclass AND conname = 'users_email_check';",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE schemaname = 'public' AND tablename = 'users' ORDER BY indexname;",
					Expected: []sql.Row{{"users_lower_email_idx"}},
				},
				{
					Query:    "INSERT INTO users VALUES ('blocked');",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT email FROM users WHERE lower(email) = 'blocked';",
					Expected: []sql.Row{{"blocked"}},
				},
			},
		},
		ScriptTest{
			Name: "primary key constraint",
			SetUpScript: []string{
				"CREATE TABLE users (id int PRIMARY KEY, email text);",
				"CREATE INDEX users_lower_email_idx ON users (lower(email));",
				"INSERT INTO users VALUES (1, 'alice@example.com');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE users DROP CONSTRAINT users_pkey;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT conname FROM pg_constraint WHERE conrelid = 'users'::regclass AND conname = 'users_pkey';",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE schemaname = 'public' AND tablename = 'users' ORDER BY indexname;",
					Expected: []sql.Row{{"users_lower_email_idx"}},
				},
				{
					Query:    "INSERT INTO users VALUES (1, 'bob@example.com');",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT id, email FROM users WHERE lower(email) IN ('alice@example.com', 'bob@example.com') ORDER BY email;",
					Expected: []sql.Row{{1, "alice@example.com"}, {1, "bob@example.com"}},
				},
			},
		},
	)
	RunScripts(t, scripts)
}

func TestDropUniqueConstraintIfExists(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "existing unique constraint is removed without an expression index",
			SetUpScript: []string{
				"CREATE TABLE users (email text UNIQUE);",
				"CREATE INDEX users_email_idx ON users (email);",
				"INSERT INTO users VALUES ('alice@example.com');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT conname FROM pg_constraint WHERE conrelid = 'users'::regclass AND conname = 'users_email_key';",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE schemaname = 'public' AND tablename = 'users' ORDER BY indexname;",
					Expected: []sql.Row{{"users_email_idx"}},
				},
				{
					Query:    "INSERT INTO users VALUES ('alice@example.com');",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT email FROM users ORDER BY email;",
					Expected: []sql.Row{{"alice@example.com"}, {"alice@example.com"}},
				},
			},
		},
	})
}

func TestDropUniqueConstraintSchemaQualified(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Auth migration retains expression index outside search path",
			SetUpScript: []string{
				"SET search_path TO public;",
				"CREATE SCHEMA auth;",
				"CREATE TABLE auth.users (email text UNIQUE);",
				"CREATE INDEX users_lower_email_idx ON auth.users (lower(email));",
				"INSERT INTO auth.users VALUES ('alice@example.com');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `DO $$
BEGIN
  ALTER TABLE ONLY auth.users DROP CONSTRAINT IF EXISTS users_email_key;
EXCEPTION
  WHEN SQLSTATE '2BP01' THEN
    RAISE NOTICE 'Unable to drop users_email_key constraint due to dependent objects';
END $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE schemaname = 'auth' AND tablename = 'users' ORDER BY indexname;",
					Expected: []sql.Row{{"users_lower_email_idx"}},
				},
				{
					Query:    "INSERT INTO auth.users VALUES ('alice@example.com');",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT email FROM auth.users WHERE lower(email) = 'alice@example.com' ORDER BY email;",
					Expected: []sql.Row{{"alice@example.com"}, {"alice@example.com"}},
				},
			},
		},
		{
			Name: "target schema is outside search path",
			SetUpScript: []string{
				"SET search_path TO public;",
				"CREATE SCHEMA auth;",
				"CREATE TABLE auth.users (email text UNIQUE);",
				"INSERT INTO auth.users VALUES ('alice@example.com');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE ONLY auth.users DROP CONSTRAINT users_email_key;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT conname FROM pg_constraint WHERE conrelid = 'auth.users'::regclass AND conname = 'users_email_key';",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT indexname FROM pg_indexes WHERE schemaname = 'auth' AND tablename = 'users';",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO auth.users VALUES ('alice@example.com');",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT email FROM auth.users ORDER BY email;",
					Expected: []sql.Row{{"alice@example.com"}, {"alice@example.com"}},
				},
			},
		},
		{
			Name: "same table and constraint names in another schema",
			SetUpScript: []string{
				"SET search_path TO public;",
				"CREATE SCHEMA auth;",
				"CREATE TABLE public.users (email text UNIQUE);",
				"CREATE TABLE auth.users (email text UNIQUE);",
				"INSERT INTO public.users VALUES ('public@example.com');",
				"INSERT INTO auth.users VALUES ('auth@example.com');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE ONLY auth.users DROP CONSTRAINT users_email_key;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT conname FROM pg_constraint WHERE conrelid = 'auth.users'::regclass AND conname = 'users_email_key';",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT conname FROM pg_constraint WHERE conrelid = 'public.users'::regclass AND conname = 'users_email_key';",
					Expected: []sql.Row{{"users_email_key"}},
				},
				{
					Query:    "SELECT schemaname, indexname FROM pg_indexes WHERE schemaname IN ('auth', 'public') AND tablename = 'users' ORDER BY schemaname, indexname;",
					Expected: []sql.Row{{"public", "users_email_key"}},
				},
				{
					Query:    "INSERT INTO auth.users VALUES ('auth@example.com');",
					Expected: []sql.Row{},
				},
				{
					Query:           "INSERT INTO public.users VALUES ('public@example.com');",
					ExpectedErr:     "duplicate unique key",
					ExpectedErrCode: "23505",
				},
				{
					Query:    "SELECT email FROM auth.users ORDER BY email;",
					Expected: []sql.Row{{"auth@example.com"}, {"auth@example.com"}},
				},
				{
					Query:    "SELECT email FROM public.users ORDER BY email;",
					Expected: []sql.Row{{"public@example.com"}},
				},
			},
		},
	})
}
