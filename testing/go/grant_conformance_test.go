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

// TestGrantConformance checks that GRANT, REVOKE, and ALTER DEFAULT PRIVILEGES resolve their targets the way
// PostgreSQL 15 does, and that unqualified routines and sequences are authorized against the schema they resolve to.
func TestGrantConformance(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "GRANT and REVOKE on nonexistent objects",
			SetUpScript: []string{
				`CREATE USER ne PASSWORD 'password';`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           `GRANT SELECT ON nosuch TO ne;`,
					ExpectedErr:     `relation "nosuch" does not exist`,
					ExpectedErrCode: "42P01",
				},
				{
					Query:           `GRANT SELECT ON public.nosuch TO ne;`,
					ExpectedErr:     `relation "public.nosuch" does not exist`,
					ExpectedErrCode: "42P01",
				},
				{
					Query:           `GRANT SELECT ON nosuchschema.t TO ne;`,
					ExpectedErr:     `schema "nosuchschema" does not exist`,
					ExpectedErrCode: "3F000",
				},
				{
					Query:           `GRANT USAGE ON SEQUENCE nosuchseq TO ne;`,
					ExpectedErr:     `relation "nosuchseq" does not exist`,
					ExpectedErrCode: "42P01",
				},
				{
					Query:           `GRANT EXECUTE ON FUNCTION nosuchf() TO ne;`,
					ExpectedErr:     `function nosuchf() does not exist`,
					ExpectedErrCode: "42883",
				},
				{
					Query:           `GRANT EXECUTE ON PROCEDURE nosuchp() TO ne;`,
					ExpectedErr:     `function nosuchp() does not exist`,
					ExpectedErrCode: "42883",
				},
				{
					Query:           `GRANT USAGE ON SCHEMA nosuchschema TO ne;`,
					ExpectedErr:     `schema "nosuchschema" does not exist`,
					ExpectedErrCode: "3F000",
				},
				{
					Query:           `GRANT CONNECT ON DATABASE nosuchdb TO ne;`,
					ExpectedErr:     `database "nosuchdb" does not exist`,
					ExpectedErrCode: "3D000",
				},
				{
					Query:           `GRANT SELECT ON ALL TABLES IN SCHEMA nosuchschema TO ne;`,
					ExpectedErr:     `schema "nosuchschema" does not exist`,
					ExpectedErrCode: "3F000",
				},
				{
					Query:           `GRANT USAGE ON ALL SEQUENCES IN SCHEMA nosuchschema TO ne;`,
					ExpectedErr:     `schema "nosuchschema" does not exist`,
					ExpectedErrCode: "3F000",
				},
				{
					Query:           `GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA nosuchschema TO ne;`,
					ExpectedErr:     `schema "nosuchschema" does not exist`,
					ExpectedErrCode: "3F000",
				},
				{
					Query:           `REVOKE SELECT ON nosuch FROM ne;`,
					ExpectedErr:     `relation "nosuch" does not exist`,
					ExpectedErrCode: "42P01",
				},
				{
					Query:           `REVOKE USAGE ON SCHEMA nosuchschema FROM ne;`,
					ExpectedErr:     `schema "nosuchschema" does not exist`,
					ExpectedErrCode: "3F000",
				},
				{
					// A missing object is reported before a missing role
					Query:           `GRANT SELECT ON TABLE nosuch2 TO nosuchrole;`,
					ExpectedErr:     `relation "nosuch2" does not exist`,
					ExpectedErrCode: "42P01",
				},
				{
					Query:           `ALTER DEFAULT PRIVILEGES IN SCHEMA nosuchschema GRANT SELECT ON TABLES TO ne;`,
					ExpectedErr:     `schema "nosuchschema" does not exist`,
					ExpectedErrCode: "3F000",
				},
			},
		},
		{
			Name: "GRANT and REVOKE on existing objects",
			SetUpScript: []string{
				`CREATE USER ex PASSWORD 'password';`,
				`CREATE SCHEMA s;`,
				`CREATE TABLE s.t (pk INT PRIMARY KEY);`,
				`INSERT INTO s.t VALUES (1);`,
				`CREATE VIEW s.v AS SELECT pk FROM s.t;`,
				`CREATE SEQUENCE s.q;`,
				`CREATE FUNCTION s.f() RETURNS INT AS $$ BEGIN RETURN 7; END; $$ LANGUAGE plpgsql;`,
				`CREATE PROCEDURE s.p() AS $$ BEGIN END; $$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `GRANT USAGE ON SCHEMA s TO ex;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `GRANT SELECT ON s.t, s.v TO ex;`,
					Expected: []sql.Row{},
				},
				{
					// A sequence is a relation, so it may be named by GRANT ... ON TABLE
					Query:    `GRANT SELECT ON s.q TO ex;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `GRANT USAGE ON SEQUENCE s.q TO ex;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `GRANT EXECUTE ON FUNCTION s.f() TO ex;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `GRANT EXECUTE ON PROCEDURE s.p() TO ex;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `GRANT EXECUTE ON FUNCTION pg_catalog.lower(text) TO ex;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `GRANT SELECT ON ALL TABLES IN SCHEMA s TO ex;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `GRANT CONNECT ON DATABASE postgres TO ex;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM s.v;`,
					Username: `ex`,
					Password: `password`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT s.f();`,
					Username: `ex`,
					Password: `password`,
					Expected: []sql.Row{{7}},
				},
				{
					Query:    `REVOKE SELECT ON s.t FROM ex;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `REVOKE SELECT ON ALL TABLES IN SCHEMA s FROM ex;`,
					Expected: []sql.Row{},
				},
				{
					Query:           `SELECT * FROM s.t;`,
					Username:        `ex`,
					Password:        `password`,
					ExpectedErr:     `permission denied for table t`,
					ExpectedErrCode: "42501",
				},
			},
		},
		{
			Name: "GRANT and REVOKE on objects in other databases",
			SetUpScript: []string{
				`CREATE USER cd PASSWORD 'password';`,
				`CREATE DATABASE otherdb;`,
				`CREATE TABLE t (pk INT PRIMARY KEY);`,
				`INSERT INTO t VALUES (1);`,
				`CREATE SEQUENCE q;`,
				`CREATE FUNCTION f() RETURNS INT AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           `GRANT SELECT ON otherdb.public.t TO cd;`,
					ExpectedErr:     `cross-database references are not implemented: "otherdb.public.t"`,
					ExpectedErrCode: "0A000",
				},
				{
					Query:           `REVOKE SELECT ON otherdb.public.t FROM cd;`,
					ExpectedErr:     `cross-database references are not implemented: "otherdb.public.t"`,
					ExpectedErrCode: "0A000",
				},
				{
					Query:           `GRANT USAGE ON SEQUENCE otherdb.public.q TO cd;`,
					ExpectedErr:     `cross-database references are not implemented: "otherdb.public.q"`,
					ExpectedErrCode: "0A000",
				},
				{
					Query:           `GRANT EXECUTE ON FUNCTION otherdb.public.f() TO cd;`,
					ExpectedErr:     `cross-database references are not implemented: otherdb.public.f`,
					ExpectedErrCode: "0A000",
				},
				{
					// The current database may be named explicitly
					Query:    `GRANT SELECT ON postgres.public.t TO cd;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `GRANT USAGE ON SEQUENCE postgres.public.q TO cd;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `GRANT EXECUTE ON FUNCTION postgres.public.f() TO cd;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM t;`,
					Username: `cd`,
					Password: `password`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT nextval('q');`,
					Username: `cd`,
					Password: `password`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "unqualified GRANT targets resolve on the search path",
			SetUpScript: []string{
				`CREATE USER sp PASSWORD 'password';`,
				`CREATE SCHEMA first_s;`,
				`CREATE TABLE secret (pk INT PRIMARY KEY);`,
				`INSERT INTO secret VALUES (1);`,
				`CREATE SEQUENCE secretseq;`,
				`ALTER DEFAULT PRIVILEGES REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;`,
				`CREATE FUNCTION secretfunc() RETURNS TEXT AS $$ BEGIN RETURN 'ran'; END; $$ LANGUAGE plpgsql;`,
				`SET search_path = first_s, public;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// first_s is the current schema, but the objects are in public
					Query:    `GRANT SELECT ON secret TO sp;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `GRANT USAGE ON SEQUENCE secretseq TO sp;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `GRANT EXECUTE ON FUNCTION secretfunc() TO sp;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM public.secret;`,
					Username: `sp`,
					Password: `password`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT nextval('public.secretseq');`,
					Username: `sp`,
					Password: `password`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT public.secretfunc();`,
					Username: `sp`,
					Password: `password`,
					Expected: []sql.Row{{"ran"}},
				},
			},
		},
		{
			Name: "unqualified routines and sequences are authorized in the schema they resolve to",
			SetUpScript: []string{
				`CREATE USER fo PASSWORD 'password';`,
				`ALTER DEFAULT PRIVILEGES REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;`,
				`CREATE FUNCTION secretfunc() RETURNS TEXT AS $$ BEGIN RETURN 'secretfunc-ran'; END; $$ LANGUAGE plpgsql;`,
				`CREATE SEQUENCE secretseq;`,
				`CREATE TABLE secret (pk INT PRIMARY KEY, ssn TEXT);`,
				`INSERT INTO secret VALUES (1, '111-11-1111');`,
				// The schema named after the user is first on the default search path
				`CREATE SCHEMA fo;`,
				`GRANT USAGE ON SCHEMA fo TO fo;`,
				`CREATE SCHEMA other_s;`,
				`CREATE SEQUENCE other_s.seq_usage_only;`,
				`GRANT USAGE ON SCHEMA other_s TO fo;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SHOW search_path;`,
					Username: `fo`,
					Password: `password`,
					Expected: []sql.Row{{"\"$user\", public"}},
				},
				{
					Query:           `SELECT secretfunc();`,
					Username:        `fo`,
					Password:        `password`,
					ExpectedErr:     "permission denied for routine secretfunc",
					ExpectedErrCode: "42501",
				},
				{
					Query:           `SELECT public.secretfunc();`,
					Username:        `fo`,
					Password:        `password`,
					ExpectedErr:     "permission denied for routine secretfunc",
					ExpectedErrCode: "42501",
				},
				{
					Query:           `SELECT nextval('secretseq');`,
					Username:        `fo`,
					Password:        `password`,
					ExpectedErr:     "permission denied for sequence secretseq",
					ExpectedErrCode: "42501",
				},
				{
					Query:           `SELECT nextval('public.secretseq');`,
					Username:        `fo`,
					Password:        `password`,
					ExpectedErr:     "permission denied for sequence secretseq",
					ExpectedErrCode: "42501",
				},
				{
					// TODO: PostgreSQL does not let USAGE ON SCHEMA grant sequence access, but existing grants rely on it
					Query:    `SELECT nextval('other_s.seq_usage_only');`,
					Username: `fo`,
					Password: `password`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:           `SELECT * FROM secret;`,
					Username:        `fo`,
					Password:        `password`,
					ExpectedErr:     "permission denied for table secret",
					ExpectedErrCode: "42501",
				},
				{
					Query:    `SET search_path = other_s, public;`,
					Username: `fo`,
					Password: `password`,
					Expected: []sql.Row{},
				},
				{
					Query:           `SELECT secretfunc();`,
					Username:        `fo`,
					Password:        `password`,
					ExpectedErr:     "permission denied for routine secretfunc",
					ExpectedErrCode: "42501",
				},
				{
					Query:           `SELECT nextval('secretseq');`,
					Username:        `fo`,
					Password:        `password`,
					ExpectedErr:     "permission denied for sequence secretseq",
					ExpectedErrCode: "42501",
				},
				{
					Query:    `GRANT EXECUTE ON FUNCTION public.secretfunc() TO fo;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `GRANT USAGE ON SEQUENCE public.secretseq TO fo;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT secretfunc();`,
					Username: `fo`,
					Password: `password`,
					Expected: []sql.Row{{"secretfunc-ran"}},
				},
				{
					Query:    `SELECT nextval('secretseq');`,
					Username: `fo`,
					Password: `password`,
					Expected: []sql.Row{{1}},
				},
				{
					// Built-in routines are still executable without a grant
					Query:    `SELECT lower('ABC');`,
					Username: `fo`,
					Password: `password`,
					Expected: []sql.Row{{"abc"}},
				},
			},
		},
	})
}
