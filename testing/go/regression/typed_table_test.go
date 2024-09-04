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

func TestTypedTable(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_typed_table)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_typed_table,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement:   `CREATE TABLE ttable1 OF nothing;`,
				ErrorString: `type "nothing" does not exist`,
			},
			{
				Statement: `CREATE TYPE person_type AS (id int, name text);`,
			},
			{
				Statement: `CREATE TABLE persons OF person_type;`,
			},
			{
				Statement: `CREATE TABLE IF NOT EXISTS persons OF person_type;`,
			},
			{
				Statement: `SELECT * FROM persons;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `\d persons
              Table "public.persons"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 id     | integer |           |          | 
 name   | text    |           |          | 
Typed table of type: person_type
CREATE FUNCTION get_all_persons() RETURNS SETOF person_type
LANGUAGE SQL
AS $$
    SELECT * FROM persons;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `SELECT * FROM get_all_persons();`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `ALTER TABLE persons ADD COLUMN comment text;`,
				ErrorString: `cannot add column to typed table`,
			},
			{
				Statement:   `ALTER TABLE persons DROP COLUMN name;`,
				ErrorString: `cannot drop column from typed table`,
			},
			{
				Statement:   `ALTER TABLE persons RENAME COLUMN id TO num;`,
				ErrorString: `cannot rename column of typed table`,
			},
			{
				Statement:   `ALTER TABLE persons ALTER COLUMN name TYPE varchar;`,
				ErrorString: `cannot alter column type of typed table`,
			},
			{
				Statement: `CREATE TABLE stuff (id int);`,
			},
			{
				Statement:   `ALTER TABLE persons INHERIT stuff;`,
				ErrorString: `cannot change inheritance of typed table`,
			},
			{
				Statement:   `CREATE TABLE personsx OF person_type (myname WITH OPTIONS NOT NULL); -- error`,
				ErrorString: `column "myname" does not exist`,
			},
			{
				Statement: `CREATE TABLE persons2 OF person_type (
    id WITH OPTIONS PRIMARY KEY,
    UNIQUE (name)
);`,
			},
			{
				Statement: `\d persons2
              Table "public.persons2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 id     | integer |           | not null | 
 name   | text    |           |          | 
Indexes:
    "persons2_pkey" PRIMARY KEY, btree (id)
    "persons2_name_key" UNIQUE CONSTRAINT, btree (name)
Typed table of type: person_type
CREATE TABLE persons3 OF person_type (
    PRIMARY KEY (id),
    name WITH OPTIONS DEFAULT ''
);`,
			},
			{
				Statement: `\d persons3
              Table "public.persons3"
 Column |  Type   | Collation | Nullable | Default  
--------+---------+-----------+----------+----------
 id     | integer |           | not null | 
 name   | text    |           |          | ''::text
Indexes:
    "persons3_pkey" PRIMARY KEY, btree (id)
Typed table of type: person_type
CREATE TABLE persons4 OF person_type (
    name WITH OPTIONS NOT NULL,
    name WITH OPTIONS DEFAULT ''  -- error, specified more than once
);`,
				ErrorString: `column "name" specified more than once`,
			},
			{
				Statement:   `DROP TYPE person_type RESTRICT;`,
				ErrorString: `cannot drop type person_type because other objects depend on it`,
			},
			{
				Statement: `function get_all_persons() depends on type person_type
table persons2 depends on type person_type
table persons3 depends on type person_type
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DROP TYPE person_type CASCADE;`,
			},
			{
				Statement:   `CREATE TABLE persons5 OF stuff; -- only CREATE TYPE AS types may be used`,
				ErrorString: `type stuff is not a composite type`,
			},
			{
				Statement: `DROP TABLE stuff;`,
			},
			{
				Statement: `CREATE TYPE person_type AS (id int, name text);`,
			},
			{
				Statement: `CREATE TABLE persons OF person_type;`,
			},
			{
				Statement: `INSERT INTO persons VALUES (1, 'test');`,
			},
			{
				Statement: `CREATE FUNCTION namelen(person_type) RETURNS int LANGUAGE SQL AS $$ SELECT length($1.name) $$;`,
			},
			{
				Statement: `SELECT id, namelen(persons) FROM persons;`,
				Results:   []sql.Row{{1, 4}},
			},
			{
				Statement: `CREATE TABLE persons2 OF person_type (
    id WITH OPTIONS PRIMARY KEY,
    UNIQUE (name)
);`,
			},
			{
				Statement: `\d persons2
              Table "public.persons2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 id     | integer |           | not null | 
 name   | text    |           |          | 
Indexes:
    "persons2_pkey" PRIMARY KEY, btree (id)
    "persons2_name_key" UNIQUE CONSTRAINT, btree (name)
Typed table of type: person_type
CREATE TABLE persons3 OF person_type (
    PRIMARY KEY (id),
    name NOT NULL DEFAULT ''
);`,
			},
			{
				Statement: `\d persons3
              Table "public.persons3"
 Column |  Type   | Collation | Nullable | Default  
--------+---------+-----------+----------+----------
 id     | integer |           | not null | 
 name   | text    |           | not null | ''::text
Indexes:
    "persons3_pkey" PRIMARY KEY, btree (id)
Typed table of type: person_type`,
			},
		},
	})
}
