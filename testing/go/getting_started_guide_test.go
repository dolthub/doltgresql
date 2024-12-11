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

// TestGettingStartedGuide tests that the steps in the Doltgres Getting Started Guide work correctly.
// https://github.com/dolthub/doltgresql?tab=readme-ov-file#getting-started
func TestGettingStartedGuide(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Doltgres Getting Started Guide",
			SetUpScript: []string{
				"create table employees (\n    id int8,\n    last_name text,\n    first_name text,\n    primary key(id));",
				"create table teams (\n    id int8,\n    team_name text,\n    primary key(id));",
				"create table employees_teams(\n    team_id int8,\n    employee_id int8,\n    primary key(team_id, employee_id),\n    foreign key (team_id) references teams(id),\n    foreign key (employee_id) references employees(id));",
			},
			Assertions: []ScriptTestAssertion{
				// Make a Dolt Commit
				{
					Query: "select * from dolt.status;",
					Expected: []sql.Row{
						{"public.employees", "f", "new table"},
						{"public.employees_teams", "f", "new table"},
						{"public.teams", "f", "new table"},
					},
				},
				{
					Query:    "call dolt_add('teams', 'employees', 'employees_teams');",
					Expected: []sql.Row{},
				},
				{
					Query: "select * from dolt.status;",
					Expected: []sql.Row{
						{"public.employees", "t", "new table"},
						{"public.employees_teams", "t", "new table"},
						{"public.teams", "t", "new table"},
					},
				},
				{
					Query:    "call dolt_commit('-m', 'Created initial schema');",
					Expected: []sql.Row{},
				},
				{
					// TODO: employees_teams is still marked as modified even though we staged and committed it. The diff
					//       in the working set shows a schema change of adding the FK references. For now, we reset to
					//       remove this artifact and prevent it from showing up in later test assertions, but this can
					//       be removed once the issue is fixed.
					//       https://github.com/dolthub/doltgresql/issues/734
					Query:    "call dolt_reset('--hard');",
					Expected: []sql.Row{},
				},
				{
					Query:    "select * from dolt.status;",
					Expected: []sql.Row{},
				},
				{
					Query:    "select count(*) from dolt.log;",
					Expected: []sql.Row{{3}},
				},

				// Insert Some Data
				{
					Query:    "insert into employees values (0, 'Sehn', 'Tim'), (1, 'Hendriks', 'Brian'), (2, 'Son','Aaron'), (3, 'Fitzgerald', 'Brian');",
					Expected: []sql.Row{},
				},
				{
					Query:    "insert into teams values (0, 'Engineering'), (1, 'Sales');",
					Expected: []sql.Row{},
				},
				{
					Query:    "insert into employees_teams(employee_id, team_id) values (0,0), (1,0), (2,0), (0,1), (3,1);",
					Expected: []sql.Row{},
				},
				{
					Query: `select first_name, last_name, team_name from employees
							join employees_teams on (employees.id=employees_teams.employee_id)
							join teams on (teams.id=employees_teams.team_id)
							where team_name='Engineering';`,
					Expected: []sql.Row{
						{"Tim", "Sehn", "Engineering"},
						{"Brian", "Hendriks", "Engineering"},
						{"Aaron", "Son", "Engineering"},
					},
				},
				{
					Skip:     true, // This returns no rows for some reason. See https://github.com/dolthub/doltgresql/issues/1063
					Query:    "select * from employees_teams where employee_id='0' and team_id='1';",
					Expected: []sql.Row{{1, 0}},
				},

				// Examine the Diff
				{
					Query: "select * from dolt.status order by table_name;",
					Expected: []sql.Row{
						{"public.employees", "f", "modified"},
						{"public.employees_teams", "f", "modified"},
						{"public.teams", "f", "modified"},
					},
				},
				{
					Query: "select to_last_name, to_first_name, to_id, to_commit, from_last_name, from_first_name, from_id, diff_type from dolt_diff_employees;",
					Expected: []sql.Row{
						{"Sehn", "Tim", 0, "WORKING", nil, nil, nil, "added"},
						{"Hendriks", "Brian", 1, "WORKING", nil, nil, nil, "added"},
						{"Son", "Aaron", 2, "WORKING", nil, nil, nil, "added"},
						{"Fitzgerald", "Brian", 3, "WORKING", nil, nil, nil, "added"},
					},
				},
				{
					Query:    "call dolt_commit('-am', 'Populated tables with data');",
					Expected: []sql.Row{},
				},
				{
					Query:    "select * from dolt.status order by table_name;",
					Expected: []sql.Row{},
				},
				{
					Query: "select message from dolt.log;",
					Expected: []sql.Row{
						{"Populated tables with data"},
						{"Created initial schema"},
						{"CREATE DATABASE"},
						{"Initialize data repository"},
					},
				},
				{
					Query: "select table_name, message, data_change, schema_change from dolt.diff order by date desc, table_name;",
					Expected: []sql.Row{
						{"public.employees", "Populated tables with data", "t", "f"},
						{"public.employees_teams", "Populated tables with data", "t", "f"},
						{"public.teams", "Populated tables with data", "t", "f"},
						{"public.employees", "Created initial schema", "f", "t"},
						{"public.employees_teams", "Created initial schema", "f", "t"},
						{"public.teams", "Created initial schema", "f", "t"},
					},
				},

				// Oh no! I made a mistake.
				{
					Query:    "drop table employees_teams;",
					Expected: []sql.Row{},
				},
				{
					Query:       "select count(*) from employees_teams;",
					ExpectedErr: "table not found: employees_teams",
				},
				{
					Query:    "call dolt_reset('--hard');",
					Expected: []sql.Row{},
				},
				{
					Query:    "select count(*) from employees_teams;",
					Expected: []sql.Row{{5}},
				},

				// Make changes on a branch
				{
					Query:    "call dolt_checkout('-b','modifications');",
					Expected: []sql.Row{},
				},
				{
					Query:    "update employees SET first_name='Timothy' where first_name='Tim';",
					Expected: []sql.Row{},
				},
				{
					Query:    "insert INTO employees (id, first_name, last_name) values (4,'Daylon', 'Wilkins');",
					Expected: []sql.Row{},
				},
				{
					Query:    "insert into employees_teams(team_id, employee_id) values (0,4);",
					Expected: []sql.Row{},
				},
				{
					Query:    "delete from employees_teams where employee_id=0 and team_id=1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "call dolt_commit('-am', 'Modifications on a branch');",
					Expected: []sql.Row{},
				},
				{
					Query:    "call dolt_checkout('main');",
					Expected: []sql.Row{},
				},
				{
					Query: "select name, latest_commit_message from dolt.branches;",
					Expected: []sql.Row{
						{"main", "Populated tables with data"},
						{"modifications", "Modifications on a branch"},
					},
				},
				{
					Query:    "select active_branch();",
					Expected: []sql.Row{{"main"}},
				},
				{
					Query:    "select * from employees;",
					Expected: []sql.Row{{0, "Sehn", "Tim"}, {1, "Hendriks", "Brian"}, {2, "Son", "Aaron"}, {3, "Fitzgerald", "Brian"}},
				},
				{
					Query:    "select * from employees as of 'modifications';",
					Expected: []sql.Row{{0, "Sehn", "Timothy"}, {1, "Hendriks", "Brian"}, {2, "Son", "Aaron"}, {3, "Fitzgerald", "Brian"}, {4, "Wilkins", "Daylon"}},
				},
				{
					// TODO: This query panics: runtime error: slice bounds out of range [:1233] with capacity 260
					//       https://github.com/dolthub/doltgresql/issues/735
					Skip:     true,
					Query:    "select * from dolt_diff('main', 'modifications', 'employees');",
					Expected: []sql.Row{},
				},

				// Make a schema change on another branch
				// TODO: Most ALTER TABLE statements aren't supported yet
				{
					Query:    "call dolt_checkout('-b', 'schema_changes');",
					Expected: []sql.Row{},
				},
			},
		},
	})
}
