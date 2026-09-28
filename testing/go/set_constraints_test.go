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

func TestSetConstraints(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "SET CONSTRAINTS",
			SetUpScript: []string{
				"CREATE TABLE parent_example (id INTEGER PRIMARY KEY, u INTEGER CONSTRAINT uniq_u UNIQUE, c INTEGER CONSTRAINT chk CHECK (c > 0));",
				"CREATE TABLE child_example (id INTEGER PRIMARY KEY, p2 INTEGER CONSTRAINT nd_fk REFERENCES parent_example(id));",
				"CREATE SCHEMA s2;",
				"CREATE TABLE s2.t2 (id INTEGER PRIMARY KEY);",
				"CREATE DOMAIN dom AS INTEGER CONSTRAINT dom_chk CHECK (VALUE > 0);",
				"CREATE TABLE named_pk_example (id INTEGER CONSTRAINT named_pk PRIMARY KEY);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SET CONSTRAINTS ALL IMMEDIATE;",
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "SET CONSTRAINTS can only be used in transaction blocks",
						},
					},
				},
				{
					Query: "SET CONSTRAINTS ALL DEFERRED;",
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "SET CONSTRAINTS can only be used in transaction blocks",
						},
					},
				},
				{
					Query:           "SET CONSTRAINTS nope IMMEDIATE;",
					ExpectedErr:     `constraint "nope" does not exist`,
					ExpectedErrCode: "42704",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "SET CONSTRAINTS ALL IMMEDIATE;",
				},
				{
					Query: "SET CONSTRAINTS ALL DEFERRED;",
				},
				{
					Query: "SET CONSTRAINTS nd_fk, parent_example_pkey, uniq_u, chk, dom_chk IMMEDIATE;",
				},
				{
					Query: "SET CONSTRAINTS ND_FK IMMEDIATE;",
				},
				{
					Query: "SET CONSTRAINTS public.nd_fk IMMEDIATE;",
				},
				{
					Query: "SET CONSTRAINTS postgres.public.nd_fk IMMEDIATE;",
				},
				{
					Query: "SET CONSTRAINTS s2.t2_pkey IMMEDIATE;",
				},
				{
					Query:           "SET CONSTRAINTS t2_pkey IMMEDIATE;",
					ExpectedErr:     `constraint "t2_pkey" does not exist`,
					ExpectedErrCode: "42704",
				},
				{
					Query: "ROLLBACK;",
				},
				{
					Query:           `SET CONSTRAINTS "ND_FK" IMMEDIATE;`,
					ExpectedErr:     `constraint "ND_FK" does not exist`,
					ExpectedErrCode: "42704",
				},
				{
					Query:           "SET CONSTRAINTS public.t2_pkey IMMEDIATE;",
					ExpectedErr:     `constraint "t2_pkey" does not exist`,
					ExpectedErrCode: "42704",
				},
				{
					Query:           "SET CONSTRAINTS nd_fk, nope IMMEDIATE;",
					ExpectedErr:     `constraint "nope" does not exist`,
					ExpectedErrCode: "42704",
				},
				{
					Query:           "SET CONSTRAINTS nosuch.nd_fk IMMEDIATE;",
					ExpectedErr:     `schema "nosuch" does not exist`,
					ExpectedErrCode: "3F000",
				},
				{
					Query:           "SET CONSTRAINTS otherdb.public.nd_fk IMMEDIATE;",
					ExpectedErr:     `cross-database references are not implemented: "otherdb.public.nd_fk"`,
					ExpectedErrCode: "0A000",
				},
				{
					Query:           "SET CONSTRAINTS nd_fk DEFERRED;",
					ExpectedErr:     `constraint "nd_fk" is not deferrable`,
					ExpectedErrCode: "42809",
				},
				{
					Query:           "SET CONSTRAINTS dom_chk DEFERRED;",
					ExpectedErr:     `constraint "dom_chk" is not deferrable`,
					ExpectedErrCode: "42809",
				},
				{
					Query:           "SET CONSTRAINTS nd_fk, nope DEFERRED;",
					ExpectedErr:     `constraint "nd_fk" is not deferrable`,
					ExpectedErrCode: "42809",
				},
				{
					Query: "SET search_path = s2, public;",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "SET CONSTRAINTS t2_pkey, nd_fk IMMEDIATE;",
				},
				{
					Query: "COMMIT;",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "SET CONSTRAINTS named_pk IMMEDIATE;",
					Skip:  true, // Primary key constraint names are not preserved
				},
				{
					Query: "ROLLBACK;",
				},
				{
					Query:           "SET CONSTRAINTS a.b.c.d IMMEDIATE;",
					ExpectedErr:     "improper qualified name (too many dotted names): a.b.c.d",
					ExpectedErrCode: "42601",
					Skip:            true, // The parser reports a generic syntax error
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "SET CONSTRAINTS ALL DEFERRED;",
				},
				{
					Query:           "INSERT INTO child_example VALUES (1, 99);",
					ExpectedErr:     "Foreign key violation on fk: `nd_fk`",
					ExpectedErrCode: "23503",
				},
				{
					Query: "ROLLBACK;",
				},
			},
		},
		{
			Name: "SET CONSTRAINTS with deferred foreign keys",
			Skip: true, // DEFERRABLE is not yet supported
			SetUpScript: []string{
				"CREATE TABLE parent_example (id INTEGER PRIMARY KEY);",
				"CREATE TABLE child_example (id INTEGER PRIMARY KEY, parent_id INTEGER CONSTRAINT child_parent_fk REFERENCES parent_example(id) DEFERRABLE INITIALLY DEFERRED);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN;",
				},
				{
					Query: "INSERT INTO child_example VALUES (1, 10);",
				},
				{
					Query: "INSERT INTO parent_example VALUES (10);",
				},
				{
					Query: "COMMIT;",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "INSERT INTO child_example VALUES (2, 20);",
				},
				{
					Query:           "COMMIT;",
					ExpectedErr:     "Foreign key violation on fk: `child_parent_fk`",
					ExpectedErrCode: "23503",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "INSERT INTO child_example VALUES (3, 30);",
				},
				{
					Query:           "SET CONSTRAINTS child_parent_fk IMMEDIATE;",
					ExpectedErr:     "Foreign key violation on fk: `child_parent_fk`",
					ExpectedErrCode: "23503",
				},
				{
					Query: "ROLLBACK;",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "SET CONSTRAINTS ALL IMMEDIATE;",
				},
				{
					Query:           "INSERT INTO child_example VALUES (4, 40);",
					ExpectedErr:     "Foreign key violation on fk: `child_parent_fk`",
					ExpectedErrCode: "23503",
				},
				{
					Query: "ROLLBACK;",
				},
				{
					Query:    "SELECT * FROM child_example;",
					Expected: []sql.Row{{1, 10}},
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "SET CONSTRAINTS child_parent_fk DEFERRED;",
				},
				{
					Query: "SET CONSTRAINTS child_parent_fk IMMEDIATE;",
				},
				{
					Query: "ROLLBACK;",
				},
			},
		},
		{
			Name: "SET CONSTRAINTS with deferrable constraints",
			Skip: true, // DEFERRABLE is not yet supported
			SetUpScript: []string{
				"CREATE TABLE p (id INTEGER PRIMARY KEY);",
				"CREATE TABLE fk_imm (id INTEGER PRIMARY KEY, pid INTEGER CONSTRAINT fk_imm_fk REFERENCES p(id) DEFERRABLE INITIALLY IMMEDIATE);",
				"CREATE TABLE u_def (id INTEGER PRIMARY KEY, v INTEGER CONSTRAINT u_def_u UNIQUE DEFERRABLE INITIALLY DEFERRED);",
				"CREATE TABLE pk_def (id INTEGER CONSTRAINT pk_def_pk PRIMARY KEY DEFERRABLE, v INTEGER);",
				"INSERT INTO u_def VALUES (1, 1), (2, 2);",
				"INSERT INTO pk_def VALUES (1, 1), (2, 2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN;",
				},
				{
					Query:           "INSERT INTO fk_imm VALUES (1, 10);",
					ExpectedErr:     "Foreign key violation on fk: `fk_imm_fk`",
					ExpectedErrCode: "23503",
				},
				{
					Query: "ROLLBACK;",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "SET CONSTRAINTS fk_imm_fk DEFERRED;",
				},
				{
					Query: "INSERT INTO fk_imm VALUES (1, 10);",
				},
				{
					Query: "INSERT INTO p VALUES (10);",
				},
				{
					Query: "COMMIT;",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "SET CONSTRAINTS ALL DEFERRED;",
				},
				{
					Query: "INSERT INTO fk_imm VALUES (2, 20);",
				},
				{
					Query: "INSERT INTO p VALUES (20);",
				},
				{
					Query: "COMMIT;",
				},
				{
					Query:    "SELECT * FROM fk_imm ORDER BY id;",
					Expected: []sql.Row{{1, 10}, {2, 20}},
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "UPDATE u_def SET v = 2 WHERE id = 1;",
				},
				{
					Query: "UPDATE u_def SET v = 1 WHERE id = 2;",
				},
				{
					Query: "COMMIT;",
				},
				{
					Query:    "SELECT * FROM u_def ORDER BY id;",
					Expected: []sql.Row{{1, 2}, {2, 1}},
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "UPDATE u_def SET v = 1 WHERE id = 1;",
				},
				{
					Query:           "COMMIT;",
					ExpectedErr:     "duplicate unique key",
					ExpectedErrCode: "23505",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "UPDATE u_def SET v = 1 WHERE id = 1;",
				},
				{
					Query:           "SET CONSTRAINTS u_def_u IMMEDIATE;",
					ExpectedErr:     "duplicate unique key",
					ExpectedErrCode: "23505",
				},
				{
					Query: "ROLLBACK;",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "SET CONSTRAINTS pk_def_pk DEFERRED;",
				},
				{
					Query: "UPDATE pk_def SET id = 2 WHERE v = 1;",
				},
				{
					Query: "UPDATE pk_def SET id = 1 WHERE v = 2;",
				},
				{
					Query: "COMMIT;",
				},
				{
					Query:    "SELECT * FROM pk_def ORDER BY v;",
					Expected: []sql.Row{{2, 1}, {1, 2}},
				},
				{
					Query: "UPDATE pk_def SET id = id + 10;",
				},
				{
					Query:    "SELECT * FROM pk_def ORDER BY v;",
					Expected: []sql.Row{{12, 1}, {11, 2}},
				},
			},
		},
	})
}
