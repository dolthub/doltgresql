/// Copyright 2023 Dolthub, Inc.
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

func TestIssues(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Issue #25",
			SetUpScript: []string{
				"create table tbl (pk int);",
				"insert into tbl values (1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `select dolt_add(".");`,
					ExpectedErr: "could not be found in any table in scope",
				},
				{
					Query:    `select dolt_add('.');`,
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:       `select dolt_commit("-m", "look ma");`,
					ExpectedErr: "could not be found in any table in scope",
				},
				{
					Query:    `select length(dolt_commit('-m', 'look ma')::text);`,
					Expected: []sql.Row{{34}},
				},
				{
					Query:       `select dolt_branch("br1");`,
					ExpectedErr: "could not be found in any table in scope",
				},
				{
					Query:    `select dolt_branch('br1');`,
					Expected: []sql.Row{{"{0}"}},
				},
			},
		},
		{
			Name: "Issue #2030",
			SetUpScript: []string{
				`CREATE TABLE sub_entities (
  project_id VARCHAR(256) NOT NULL,
  entity_id  VARCHAR(256) NOT NULL,
  id         VARCHAR(256) NOT NULL,
  name       VARCHAR(256) NOT NULL,
  PRIMARY KEY (project_id, entity_id, id)
);
`,
				`
CREATE TABLE entities (
  project_id              VARCHAR(256) NOT NULL,
  id                      VARCHAR(256) NOT NULL,
  name                    VARCHAR(256) NOT NULL,
  default_sub_entity_id   VARCHAR(256),
  PRIMARY KEY (project_id, id)
);
`,
				`
CREATE TABLE conversations (
  id                 VARCHAR(256) NOT NULL,
  tenant_id          VARCHAR(256) NOT NULL,
  project_id         VARCHAR(256) NOT NULL,
  active_sub_agent_id VARCHAR(256) NOT NULL,
  PRIMARY KEY (tenant_id, project_id, id)
);
`,
				`INSERT INTO sub_entities (project_id, entity_id, id, name) VALUES
  ('projectA', 'entityA', 'subA1', 'Sub-Entity A1'),
  ('projectA', 'entityB', 'subB1', 'Sub-Entity B1');
`,
				`INSERT INTO entities (project_id, id, name, default_sub_entity_id) VALUES
  ('projectA', 'entityA', 'Entity A', 'subA1'),
  ('projectA', 'entityB', 'Entity B', 'subB1');
`,
				`INSERT INTO conversations (tenant_id, project_id, id, active_sub_agent_id) VALUES
  ('tenant1', 'projectA', 'conv1', 'subA1'),
  ('tenant1', 'projectA', 'conv2', 'subB1');
`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `select
  "entities"."project_id",
  "entities"."id",
  "entities"."name",
  "entities"."default_sub_entity_id",
  "entities_defaultSubEntity"."data" as "defaultSubEntity"
from "entities" "entities"
left join lateral (
  select json_build_array(
           "entities_defaultSubEntity"."project_id",
           "entities_defaultSubEntity"."entity_id",
           "entities_defaultSubEntity"."id",
           "entities_defaultSubEntity"."name"
         ) as "data"
  from (
    select * from "sub_entities" "entities_defaultSubEntity"
    where "entities_defaultSubEntity"."id" = "entities"."default_sub_entity_id"
    limit $1
  ) "entities_defaultSubEntity"
) "entities_defaultSubEntity" on true
where ("entities"."project_id" = $2 and "entities"."id" = $3)
limit $4`,
					BindVars: []any{
						int64(1),
						"projectA",
						"entityA",
						int64(1),
					},
					Expected: []sql.Row{
						{
							"projectA",
							"entityA",
							"Entity A",
							"subA1",
							`["projectA", "entityA", "subA1", "Sub-Entity A1"]`,
						},
					},
				},
				{
					Query: `select
  "entities"."project_id",
  "entities"."id",
  "entities"."name",
  "entities"."default_sub_entity_id",
  "entities_defaultSubEntity"."data" as "defaultSubEntity"
from "entities" "entities"
left join lateral (
  select json_build_array(
           "entities_defaultSubEntity"."project_id",
           "entities_defaultSubEntity"."entity_id",
           "entities_defaultSubEntity"."id",
           "entities_defaultSubEntity"."name"
         ) as "data"
  from (
    select * from "sub_entities" "entities_defaultSubEntity"
    where "entities_defaultSubEntity"."id" = "entities"."default_sub_entity_id"
    limit 1
  ) "entities_defaultSubEntity"
) "entities_defaultSubEntity" on true
where ("entities"."project_id" = 'projectA' and "entities"."id" = 'entityA')
limit 1`,
					Expected: []sql.Row{
						{
							"projectA",
							"entityA",
							"Entity A",
							"subA1",
							`["projectA", "entityA", "subA1", "Sub-Entity A1"]`,
						},
					},
				},
			},
		},
		{
			Name: "Issue #2049",
			SetUpScript: []string{
				`CREATE TABLE jsonb_test (id VARCHAR(256) NOT NULL PRIMARY KEY, "jsonbColumn" JSONB);`,
				`INSERT INTO jsonb_test VALUES ('test', '{"test": "value\n"}');`,
				`INSERT INTO jsonb_test VALUES ('test2', '{"test": "value\t"}');`,
				`INSERT INTO jsonb_test VALUES ('test3', '{"test": "value\r"}');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM jsonb_test;",
					// The pgx library incorrectly reinterprets our JSON value by replacing the individual newline
					// characters (ASCII 92,110) with the actual newline character (ASCII 10), which is incorrect for us.
					// Therefore, we have to use the raw returned values. To make it more clear, we aren't using a raw
					// string literal and instead escaping the characters in the byte slice. We also test other escape
					// characters that are replaced.
					ExpectedRaw: [][][]byte{
						{[]byte("test"), []byte("{\"test\": \"value\\n\"}")},
						{[]byte("test2"), []byte("{\"test\": \"value\\t\"}")},
						{[]byte("test3"), []byte("{\"test\": \"value\\r\"}")},
					},
				},
			},
		},
		{
			Name: "Issue #2197 Part 1",
			SetUpScript: []string{
				`CREATE TABLE t1 (a INT, b VARCHAR(3));`,
				`CREATE TABLE t2(id SERIAL, t1 t1);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `INSERT INTO t2(t1) VALUES (ROW(1, 'abc'));`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM t2;`,
					Expected: []sql.Row{{1, "(1,abc)"}},
				},
				{
					Query:       `INSERT INTO t2(t1) VALUES (ROW('a', 'def'));`,
					ExpectedErr: "invalid input syntax for type",
				},
				{
					Query:       `INSERT INTO t2(t1) VALUES (ROW(true, 'def'));`,
					ExpectedErr: "Cannot cast type",
				},
				{
					Query:       `INSERT INTO t2(t1) VALUES (ROW(2, 'def', 'ghi'));`,
					ExpectedErr: "cannot cast type",
				},
				{
					Query:       `INSERT INTO t2(t1) VALUES (ROW(2));`,
					ExpectedErr: "cannot cast type",
				},
			},
		},
	})
}
