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
					Query:    `CREATE DOMAIN year AS integer DEFAULT 1;`,
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
			},
		},
		{
			Name:        "create table with domain type",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE DOMAIN year AS integer CONSTRAINT year_check CHECK (((VALUE >= 1901) AND (VALUE <= 2155)));`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE table_with_domain (pk int primary key, y year);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO table_with_domain VALUES (1, 1999)`,
					Expected: []sql.Row{},
				},
				{
					// TODO: the correct error msg: `value for domain year violates check constraint "year_check"`
					Query:       `INSERT INTO table_with_domain VALUES (2, 1899)`,
					ExpectedErr: `Check constraint "year_check" violated`,
				},
				{
					Query:    `SELECT * FROM table_with_domain`,
					Expected: []sql.Row{{1, 1999}},
				},
			},
		},
		{
			Name:        "create table with domain type with default value",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE DOMAIN year AS integer DEFAULT 2000;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE table_with_domain_with_default (pk int primary key, y year);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO table_with_domain_with_default VALUES (1, 1999)`,
					Expected: []sql.Row{},
				},
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
			Name: "drop domain",
			SetUpScript: []string{
				`CREATE DOMAIN year AS integer CONSTRAINT year_check CHECK (((VALUE >= 1901) AND (VALUE <= 2155)));`,
				`CREATE TABLE table_with_domain (pk int primary key, y year);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Skip:        true, // TODO
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
