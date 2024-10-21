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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestDomain(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:        "create domain",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `CREATE DOMAIN year AS integer CONSTRAINT not_null_c NOT NULL CONSTRAINT null_c  NULL;`,
					ExpectedErr: `conflicting NULL/NOT NULL constraints`,
				},
				{
					Query:       `CREATE DOMAIN year AS integer NULL NOT NULL;`,
					ExpectedErr: `conflicting NULL/NOT NULL constraints`,
				},
				{
					Query:    `CREATE DOMAIN year AS integer DEFAULT 1999 NOT NULL CONSTRAINT year_check CHECK (((VALUE >= 1901) AND (VALUE <= 2155)));`,
					Expected: []sql.Row{},
				},
				{
					Query:       `CREATE DOMAIN year AS integer CONSTRAINT year_check CHECK (((VALUE >= 1901) AND (VALUE <= 2155)));`,
					ExpectedErr: `type "year" already exists`,
				},
				{
					Query:    `CREATE DOMAIN year_with_check AS integer CONSTRAINT year_check CHECK (((VALUE >= 1901) AND (VALUE <= 2155)));`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE DOMAIN year_with_two_checks AS integer CONSTRAINT year_check_min CHECK (VALUE >= 1901) CONSTRAINT year_check_max CHECK (VALUE <= 2155);`,
					Expected: []sql.Row{},
				},
				{
					Query:       `CREATE TABLE test_table (id int primary key, v non_existing_domain);`,
					ExpectedErr: `type "non_existing_domain" does not exist`,
				},
			},
		},
		{
			Name: "create table with domain type",
			SetUpScript: []string{
				`CREATE DOMAIN year AS integer CONSTRAINT year_check CHECK (((VALUE >= 1901) AND (VALUE <= 2155)));`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE TABLE table_with_domain (pk int primary key, y year);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO table_with_domain VALUES (1, 1999)`,
					Expected: []sql.Row{},
				},
				{
					Query:       `INSERT INTO table_with_domain VALUES (2, 1899)`,
					ExpectedErr: `constraint "year_check"`,
				},
				{
					Query:    `SELECT * FROM table_with_domain`,
					Expected: []sql.Row{{1, 1999}},
				},
			},
		},
		{
			Name: "create table with domain type with default value",
			SetUpScript: []string{
				`CREATE DOMAIN year AS integer DEFAULT 2000;`,
				`CREATE TABLE table_with_domain_with_default (pk int primary key, y year);`,
				`INSERT INTO table_with_domain_with_default VALUES (1, 1999)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `INSERT INTO table_with_domain_with_default(pk) VALUES (2)`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM table_with_domain_with_default`,
					Expected: []sql.Row{{1, 1999}, {2, 2000}},
				},
			},
		},
		{
			Name: "create table with domain type with not null constraint",
			SetUpScript: []string{
				`CREATE DOMAIN year AS integer NOT NULL;`,
				`CREATE TABLE tbl_not_null (pk int primary key, y year);`,
				`INSERT INTO tbl_not_null VALUES (1, 1999)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// TODO: the correct error msg: `domain year does not allow null values`
					Query:       `INSERT INTO tbl_not_null VALUES (2, null)`,
					ExpectedErr: `column name 'y' is non-nullable but attempted to set a value of null`,
				},
				{
					// TODO: the correct error msg: `domain year does not allow null values`
					Query:       `INSERT INTO tbl_not_null(pk) VALUES (2)`,
					ExpectedErr: `Field 'y' doesn't have a default value`,
				},
				{
					Query:    `SELECT * FROM tbl_not_null`,
					Expected: []sql.Row{{1, 1999}},
				},
			},
		},
		{
			Name: "update on table with domain type",
			SetUpScript: []string{
				`CREATE DOMAIN year AS integer NOT NULL CONSTRAINT year_check_min CHECK (VALUE >= 1901) CONSTRAINT year_check_max CHECK (VALUE <= 2155);`,
				`CREATE TABLE test_table (pk int primary key, y year);`,
				`INSERT INTO test_table VALUES (1, 1999), (2, 2000)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `UPDATE test_table SET y = 1902 WHERE pk = 1;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `UPDATE test_table SET y = 1900 WHERE pk = 1;`,
					ExpectedErr: `constraint "year_check_min"`,
				},
				{
					// TODO: the correct error msg: `domain year does not allow null values`
					Query:       `UPDATE test_table SET y = null WHERE pk = 1;`,
					ExpectedErr: `column name 'y' is non-nullable but attempted to set a value of null`,
				},
				{
					Query:    `SELECT * FROM test_table`,
					Expected: []sql.Row{{1, 1902}, {2, 2000}},
				},
			},
		},
		{
			Name: "domain type as text type",
			SetUpScript: []string{
				`CREATE DOMAIN non_empty_string AS text NULL CONSTRAINT name_check CHECK (VALUE <> '');`,
				`CREATE TABLE non_empty_string (id int primary key, first_name non_empty_string, last_name non_empty_string);`,
				`INSERT INTO non_empty_string VALUES (1, 'John', 'Doe')`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `INSERT INTO non_empty_string VALUES (2, 'Jane', 'Doe')`,
					Expected: []sql.Row{},
				},
				{
					Query:       `UPDATE non_empty_string SET last_name = '' WHERE first_name = 'Jane'`,
					ExpectedErr: `Check constraint "name_check" violated`,
				},
				{
					Query:    `UPDATE non_empty_string SET last_name = NULL WHERE first_name = 'Jane'`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM non_empty_string`,
					Expected: []sql.Row{{1, "John", "Doe"}, {2, "Jane", nil}},
				},
			},
		},
		{
			Skip: true, // TODO: support DROP DOMAIN in separate PR
			Name: "drop domain",
			SetUpScript: []string{
				`CREATE DOMAIN year AS integer CONSTRAINT year_check CHECK (((VALUE >= 1901) AND (VALUE <= 2155)));`,
				`CREATE TABLE table_with_domain (pk int primary key, y year);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `DROP DOMAIN year;`,
					ExpectedErr: `cannot drop type year because other objects depend on it`,
				},
				{
					Query:    `DROP TABLE table_with_domain;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP DOMAIN year;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP DOMAIN IF EXISTS year;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `DROP DOMAIN non_existing_domain;`,
					ExpectedErr: `type "non_existing_domain" does not exist`,
				},
			},
		},
	})
}
