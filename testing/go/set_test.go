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
	"fmt"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestSetStatements(t *testing.T) {
	RunScripts(t, setStmts)
}

// setStmts test on simple cases on setting and showing the config parameters.
// This includes setting the parameters successfully and
// showing the updated value if they are of context, `user` or `superuser`.
// If the parameters are of any other context (e.g. `sighup` or `postmaster`),
// it returns an error as those parameters can only be updated
// through configuration file and/or SIGHUP signal and/or having appropriate roles.
var setStmts = []ScriptTest{
	{
		Name:        "special case for TIME ZONE",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW timezone",
				Expected: []sql.Row{{"America/Los_Angeles"}},
			},
			{
				Query:    "SET timezone TO '+00:00';",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW timezone",
				Expected: []sql.Row{{"+00:00"}},
			},
			{
				Query:    "SET TIME ZONE LOCAL;",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW timezone",
				Expected: []sql.Row{{"America/Los_Angeles"}},
			},
			{
				Query:    "SET TIME ZONE '+00:00';",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW timezone",
				Expected: []sql.Row{{"+00:00"}},
			},
			{
				Query:    "SET TIME ZONE DEFAULT;",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW timezone",
				Expected: []sql.Row{{"America/Los_Angeles"}},
			},
			{
				Query:    "SELECT current_setting('timezone')",
				Expected: []sql.Row{{"America/Los_Angeles"}},
			},
		},
	},
	{
		Name:        "special case for SCHEMA",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW search_path",
				Expected: []sql.Row{{"\"$user\", public,"}},
			},
			{
				Query:    "SET SCHEMA 'postgres';",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW search_path",
				Expected: []sql.Row{{"postgres"}},
			},
			{
				Query:    "SET search_path = public, pg_catalog;",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW search_path",
				Expected: []sql.Row{{"public, pg_catalog"}},
			},
			{
				Query:    "SET search_path = postgres;",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW search_path",
				Expected: []sql.Row{{"postgres"}},
			},
			{
				Query:    "SELECT current_setting('search_path')",
				Expected: []sql.Row{{"postgres"}},
			},
		},
	},
	{
		Name:        "special case for NAMES",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW client_encoding",
				Expected: []sql.Row{{"UTF8"}},
			},
			{
				Query:    "SET NAMES 'LATIN1';",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW client_encoding;",
				Expected: []sql.Row{{"LATIN1"}},
			},
			{
				Query:    "SET client_encoding = DEFAULT;",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW client_encoding;",
				Expected: []sql.Row{{"UTF8"}},
			},
			{
				Query:    "SELECT current_setting('client_encoding')",
				Expected: []sql.Row{{"UTF8"}},
			},
		},
	},
	{
		Name:        "special case SEED",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW geqo_seed",
				Expected: []sql.Row{{float64(0)}},
			},
			{
				Query:    "SET SEED 1;",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW geqo_seed",
				Expected: []sql.Row{{float64(1)}},
			},
			{
				Query:    "SELECT current_setting('geqo_seed')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'allow_in_place_tablespaces' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW allow_in_place_tablespaces",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET allow_in_place_tablespaces TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW allow_in_place_tablespaces",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET allow_in_place_tablespaces TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW allow_in_place_tablespaces",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('allow_in_place_tablespaces')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'allow_system_table_mods' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW allow_system_table_mods",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET allow_system_table_mods TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW allow_system_table_mods",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET allow_system_table_mods TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW allow_system_table_mods",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('allow_system_table_mods')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'application_name' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW application_name",
				Expected: []sql.Row{{"psql"}},
			},
			{
				Query:    "SET application_name TO 'postgresql'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW application_name",
				Expected: []sql.Row{{"postgresql"}},
			},
			{
				Query:    "SET application_name TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW application_name",
				Expected: []sql.Row{{"psql"}},
			},
			{
				Query:    "SELECT current_setting('application_name')",
				Expected: []sql.Row{{"psql"}},
			},
		},
	},
	{
		Name:        "set 'archive_cleanup_command' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW archive_cleanup_command",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET archive_cleanup_command TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('archive_cleanup_command')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'archive_command' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW archive_command",
				Expected: []sql.Row{{"(disabled)"}},
			},
			{
				Query:       "SET archive_command TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('archive_command')",
				Expected: []sql.Row{{"(disabled)"}},
			},
		},
	},
	{
		Name:        "set 'archive_library' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW archive_library",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET archive_library TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('archive_library')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'archive_mode' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW archive_mode",
				Expected: []sql.Row{{"off"}},
			},
			{
				Query:       "SET archive_mode TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('archive_mode')",
				Expected: []sql.Row{{"off"}},
			},
		},
	},
	{
		Name:        "set 'archive_timeout' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW archive_timeout",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:       "SET archive_timeout TO '0'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('archive_timeout')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'array_nulls' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW array_nulls",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET array_nulls TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW array_nulls",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET array_nulls TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW array_nulls",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('array_nulls')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'authentication_timeout' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW authentication_timeout",
				Expected: []sql.Row{{int64(60)}},
			},
			{
				Query:       "SET authentication_timeout TO '120'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('authentication_timeout')",
				Expected: []sql.Row{{"60"}},
			},
		},
	},
	{
		Name:        "set 'autovacuum' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW autovacuum",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:       "SET autovacuum TO 'on'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('autovacuum')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'autovacuum_analyze_scale_factor' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW autovacuum_analyze_scale_factor",
				Expected: []sql.Row{{0.1}},
			},
			{
				Query:       "SET autovacuum_analyze_scale_factor TO '0.1'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('autovacuum_analyze_scale_factor')",
				Expected: []sql.Row{{"0.1"}},
			},
		},
	},
	{
		Name:        "set 'autovacuum_analyze_threshold' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW autovacuum_analyze_threshold",
				Expected: []sql.Row{{int64(50)}},
			},
			{
				Query:       "SET autovacuum_analyze_threshold TO '50'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('autovacuum_analyze_threshold')",
				Expected: []sql.Row{{"50"}},
			},
		},
	},
	{
		Name:        "set 'autovacuum_freeze_max_age' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW autovacuum_freeze_max_age",
				Expected: []sql.Row{{int64(2000000000)}},
			},
			{
				Query:       "SET autovacuum_freeze_max_age TO '200000000'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('autovacuum_freeze_max_age')",
				Expected: []sql.Row{{"2000000000"}},
			},
		},
	},
	{
		Name:        "set 'autovacuum_max_workers' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW autovacuum_max_workers",
				Expected: []sql.Row{{int64(3)}},
			},
			{
				Query:       "SET autovacuum_max_workers TO '3'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('autovacuum_max_workers')",
				Expected: []sql.Row{{"3"}},
			},
		},
	},
	{
		Name:        "set 'autovacuum_multixact_freeze_max_age' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW autovacuum_multixact_freeze_max_age",
				Expected: []sql.Row{{int64(400000000)}},
			},
			{
				Query:       "SET autovacuum_multixact_freeze_max_age TO '400000000'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('autovacuum_multixact_freeze_max_age')",
				Expected: []sql.Row{{"400000000"}},
			},
		},
	},
	{
		Name:        "set 'autovacuum_naptime' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW autovacuum_naptime",
				Expected: []sql.Row{{int64(60)}},
			},
			{
				Query:       "SET autovacuum_naptime TO '60'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('autovacuum_naptime')",
				Expected: []sql.Row{{"60"}},
			},
		},
	},
	{
		Name:        "set 'autovacuum_vacuum_cost_delay' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW autovacuum_vacuum_cost_delay",
				Expected: []sql.Row{{float64(2)}},
			},
			{
				Query:       "SET autovacuum_vacuum_cost_delay TO '2'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('autovacuum_vacuum_cost_delay')",
				Expected: []sql.Row{{"2"}},
			},
		},
	},
	{
		Name:        "set 'autovacuum_vacuum_cost_limit' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW autovacuum_vacuum_cost_limit",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:       "SET autovacuum_vacuum_cost_limit TO '-1'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('autovacuum_vacuum_cost_limit')",
				Expected: []sql.Row{{"-1"}},
			},
		},
	},
	{
		Name:        "set 'autovacuum_vacuum_insert_scale_factor' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW autovacuum_vacuum_insert_scale_factor",
				Expected: []sql.Row{{float64(0.2)}},
			},
			{
				Query:       "SET autovacuum_vacuum_insert_scale_factor TO '0.2'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('autovacuum_vacuum_insert_scale_factor')",
				Expected: []sql.Row{{"0.2"}},
			},
		},
	},
	{
		Name:        "set 'autovacuum_vacuum_insert_threshold' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW autovacuum_vacuum_insert_threshold",
				Expected: []sql.Row{{int64(1000)}},
			},
			{
				Query:       "SET autovacuum_vacuum_insert_threshold TO '1000'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('autovacuum_vacuum_insert_threshold')",
				Expected: []sql.Row{{"1000"}},
			},
		},
	},
	{
		Name:        "set 'autovacuum_vacuum_scale_factor' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW autovacuum_vacuum_scale_factor",
				Expected: []sql.Row{{float64(0.2)}},
			},
			{
				Query:       "SET autovacuum_vacuum_scale_factor TO '0.2'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('autovacuum_vacuum_scale_factor')",
				Expected: []sql.Row{{"0.2"}},
			},
		},
	},
	{
		Name:        "set 'autovacuum_vacuum_threshold' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW autovacuum_vacuum_threshold",
				Expected: []sql.Row{{int64(50)}},
			},
			{
				Query:       "SET autovacuum_vacuum_threshold TO '50'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('autovacuum_vacuum_threshold')",
				Expected: []sql.Row{{"50"}},
			},
		},
	},
	{
		Name:        "set 'autovacuum_work_mem' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW autovacuum_work_mem",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:       "SET autovacuum_work_mem TO '-1'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('autovacuum_work_mem')",
				Expected: []sql.Row{{"-1"}},
			},
		},
	},
	{
		Name:        "set 'backend_flush_after' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW backend_flush_after",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET backend_flush_after TO '256'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW backend_flush_after",
				Expected: []sql.Row{{int64(256)}},
			},
			{
				Query:    "SET backend_flush_after TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW backend_flush_after",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('backend_flush_after')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'backslash_quote' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW backslash_quote",
				Expected: []sql.Row{{"safe_encoding"}},
			},
			{
				Query:    "SET backslash_quote TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW backslash_quote",
				Expected: []sql.Row{{"on"}},
			},
			{
				Query:    "SET backslash_quote TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW backslash_quote",
				Expected: []sql.Row{{"safe_encoding"}},
			},
			{
				Query:    "SELECT current_setting('backslash_quote')",
				Expected: []sql.Row{{"safe_encoding"}},
			},
		},
	},
	{
		Name:        "set 'backtrace_functions' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW backtrace_functions",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SET backtrace_functions TO 'default'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW backtrace_functions",
				Expected: []sql.Row{{"default"}},
			},
			{
				Query:    "SET backtrace_functions TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW backtrace_functions",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SELECT current_setting('backtrace_functions')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'bgwriter_delay' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW bgwriter_delay",
				Expected: []sql.Row{{int64(200)}},
			},
			{
				Query:       "SET bgwriter_delay TO '200'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('bgwriter_delay')",
				Expected: []sql.Row{{"200"}},
			},
		},
	},
	{
		Name:        "set 'bgwriter_flush_after' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW bgwriter_flush_after",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:       "SET bgwriter_flush_after TO '0'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('bgwriter_flush_after')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'bgwriter_lru_maxpages' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW bgwriter_lru_maxpages",
				Expected: []sql.Row{{int64(100)}},
			},
			{
				Query:       "SET bgwriter_lru_maxpages TO '100'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('bgwriter_lru_maxpages')",
				Expected: []sql.Row{{"100"}},
			},
		},
	},
	{
		Name:        "set 'bgwriter_lru_multiplier' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW bgwriter_lru_multiplier",
				Expected: []sql.Row{{float64(2)}},
			},
			{
				Query:       "SET bgwriter_lru_multiplier TO '2'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('bgwriter_lru_multiplier')",
				Expected: []sql.Row{{"2"}},
			},
		},
	},
	{
		Name:        "set 'block_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW block_size",
				Expected: []sql.Row{{int64(8192)}},
			},
			{
				Query:       "SET block_size TO '8192'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('block_size')",
				Expected: []sql.Row{{"8192"}},
			},
		},
	},
	{
		Name:        "set 'bonjour' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW bonjour",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET bonjour TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('bonjour')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'bonjour_name' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW bonjour_name",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET bonjour_name TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('bonjour_name')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'bytea_output' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW bytea_output",
				Expected: []sql.Row{{"hex"}},
			},
			{
				Query:    "SET bytea_output TO 'escape'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW bytea_output",
				Expected: []sql.Row{{"escape"}},
			},
			{
				Query:    "SET bytea_output TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW bytea_output",
				Expected: []sql.Row{{"hex"}},
			},
			{
				Query:    "SELECT current_setting('bytea_output')",
				Expected: []sql.Row{{"hex"}},
			},
		},
	},
	{
		Name:        "set 'check_function_bodies' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW check_function_bodies",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET check_function_bodies TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW check_function_bodies",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET check_function_bodies TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW check_function_bodies",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('check_function_bodies')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'checkpoint_completion_target' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW checkpoint_completion_target",
				Expected: []sql.Row{{float64(0.9)}},
			},
			{
				Query:       "SET checkpoint_completion_target TO '0.9'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('checkpoint_completion_target')",
				Expected: []sql.Row{{"0.9"}},
			},
		},
	},
	{
		Name:        "set 'checkpoint_flush_after' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW checkpoint_flush_after",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:       "SET checkpoint_flush_after TO '0'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('checkpoint_flush_after')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'checkpoint_timeout' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW checkpoint_timeout",
				Expected: []sql.Row{{int64(300)}},
			},
			{
				Query:       "SET checkpoint_timeout TO '300'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('checkpoint_timeout')",
				Expected: []sql.Row{{"300"}},
			},
		},
	},
	{
		Name:        "set 'checkpoint_warning' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW checkpoint_warning",
				Expected: []sql.Row{{int64(30)}},
			},
			{
				Query:       "SET checkpoint_warning TO '30'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('checkpoint_warning')",
				Expected: []sql.Row{{"30"}},
			},
		},
	},
	{
		Name:        "set 'client_connection_check_interval' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW client_connection_check_interval",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET client_connection_check_interval TO 10",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW client_connection_check_interval",
				Expected: []sql.Row{{int64(10)}},
			},
			{
				Query:    "SET client_connection_check_interval TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW client_connection_check_interval",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('client_connection_check_interval')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'client_encoding' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW client_encoding",
				Expected: []sql.Row{{"UTF8"}},
			},
			{
				Query:    "SET client_encoding TO 'LATIN1'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW client_encoding",
				Expected: []sql.Row{{"LATIN1"}},
			},
			{
				Query:    "SET client_encoding TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW client_encoding",
				Expected: []sql.Row{{"UTF8"}},
			},
			{
				Query:    "SELECT current_setting('client_encoding')",
				Expected: []sql.Row{{"UTF8"}},
			},
		},
	},
	{
		Name:        "set 'client_min_messages' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW client_min_messages",
				Expected: []sql.Row{{"notice"}},
			},
			{
				Query:    "SET client_min_messages TO 'log'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW client_min_messages",
				Expected: []sql.Row{{"log"}},
			},
			{
				Query:    "SET client_min_messages TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW client_min_messages",
				Expected: []sql.Row{{"notice"}},
			},
			{
				Query:    "SELECT current_setting('client_min_messages')",
				Expected: []sql.Row{{"notice"}},
			},
		},
	},
	{
		Name:        "set 'cluster_name' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW cluster_name",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET cluster_name TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('cluster_name')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'commit_delay' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW commit_delay",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET commit_delay TO 100000",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW commit_delay",
				Expected: []sql.Row{{int64(100000)}},
			},
			{
				Query:    "SET commit_delay TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW commit_delay",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('commit_delay')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'commit_siblings' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW commit_siblings",
				Expected: []sql.Row{{int64(5)}},
			},
			{
				Query:    "SET commit_siblings TO '1000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW commit_siblings",
				Expected: []sql.Row{{int64(1000)}},
			},
			{
				Query:    "SET commit_siblings TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW commit_siblings",
				Expected: []sql.Row{{int64(5)}},
			},
			{
				Query:    "SELECT current_setting('commit_siblings')",
				Expected: []sql.Row{{"5"}},
			},
		},
	},
	{
		Name:        "set 'compute_query_id' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW compute_query_id",
				Expected: []sql.Row{{"auto"}},
			},
			{
				Query:    "SET compute_query_id TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW compute_query_id",
				Expected: []sql.Row{{"on"}},
			},
			{
				Query:    "SET compute_query_id TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW compute_query_id",
				Expected: []sql.Row{{"auto"}},
			},
			{
				Query:    "SELECT current_setting('compute_query_id')",
				Expected: []sql.Row{{"auto"}},
			},
		},
	},
	{
		Name:        "set 'config_file' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW config_file",
				Expected: []sql.Row{{"postgresql.conf"}},
			},
			{
				Query:       "SET config_file TO '/Users/postgres/postgresql.conf'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('config_file')",
				Expected: []sql.Row{{"postgresql.conf"}},
			},
		},
	},
	{
		Name:        "set 'constraint_exclusion' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW constraint_exclusion",
				Expected: []sql.Row{{"partition"}},
			},
			{
				Query:    "SET constraint_exclusion TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW constraint_exclusion",
				Expected: []sql.Row{{"on"}},
			},
			{
				Query:    "SET constraint_exclusion TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW constraint_exclusion",
				Expected: []sql.Row{{"partition"}},
			},
			{
				Query:    "SELECT current_setting('constraint_exclusion')",
				Expected: []sql.Row{{"partition"}},
			},
		},
	},
	{
		Name:        "set 'cpu_index_tuple_cost' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW cpu_index_tuple_cost",
				Expected: []sql.Row{{float64(0.005)}},
			},
			{
				Query:    "SET cpu_index_tuple_cost TO '0.01'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW cpu_index_tuple_cost",
				Expected: []sql.Row{{float64(0.01)}},
			},
			{
				Query:    "SET cpu_index_tuple_cost TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW cpu_index_tuple_cost",
				Expected: []sql.Row{{float64(0.005)}},
			},
			{
				Query:    "SELECT current_setting('cpu_index_tuple_cost')",
				Expected: []sql.Row{{"0.005"}},
			},
		},
	},
	{
		Name:        "set 'cpu_operator_cost' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW cpu_operator_cost",
				Expected: []sql.Row{{float64(0.0025)}},
			},
			{
				Query:    "SET cpu_operator_cost TO '0.005'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW cpu_operator_cost",
				Expected: []sql.Row{{float64(0.005)}},
			},
			{
				Query:    "SET cpu_operator_cost TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW cpu_operator_cost",
				Expected: []sql.Row{{float64(0.0025)}},
			},
			{
				Query:    "SELECT current_setting('cpu_operator_cost')",
				Expected: []sql.Row{{"0.0025"}},
			},
		},
	},
	{
		Name:        "set 'cpu_tuple_cost' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW cpu_tuple_cost",
				Expected: []sql.Row{{float64(0.01)}},
			},
			{
				Query:    "SET cpu_tuple_cost TO '0.02'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW cpu_tuple_cost",
				Expected: []sql.Row{{float64(0.02)}},
			},
			{
				Query:    "SET cpu_tuple_cost TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW cpu_tuple_cost",
				Expected: []sql.Row{{float64(0.01)}},
			},
			{
				Query:    "SELECT current_setting('cpu_tuple_cost')",
				Expected: []sql.Row{{"0.01"}},
			},
		},
	},
	{
		Name:        "set 'createrole_self_grant' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW createrole_self_grant",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SET createrole_self_grant TO 'inherit'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW createrole_self_grant",
				Expected: []sql.Row{{"inherit"}},
			},
			{
				Query:    "SET createrole_self_grant TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW createrole_self_grant",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SELECT current_setting('createrole_self_grant')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'cursor_tuple_fraction' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW cursor_tuple_fraction",
				Expected: []sql.Row{{float64(0.1)}},
			},
			{
				Query:    "SET cursor_tuple_fraction TO '0.2'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW cursor_tuple_fraction",
				Expected: []sql.Row{{float64(0.2)}},
			},
			{
				Query:    "SET cursor_tuple_fraction TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW cursor_tuple_fraction",
				Expected: []sql.Row{{float64(0.1)}},
			},
			{
				Query:    "SELECT current_setting('cursor_tuple_fraction')",
				Expected: []sql.Row{{"0.1"}},
			},
		},
	},
	{
		Name:        "set 'data_checksums' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW data_checksums",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET data_checksums TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('data_checksums')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'data_directory' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW data_directory",
				Expected: []sql.Row{{"postgres"}},
			},
			{
				Query:       "SET data_directory TO '/Users/postgres'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('data_directory')",
				Expected: []sql.Row{{"postgres"}},
			},
		},
	},
	{
		Name:        "set 'data_directory_mode' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW data_directory_mode",
				Expected: []sql.Row{{int64(448)}},
			},
			{
				Query:       "SET data_directory_mode TO '448'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('data_directory_mode')",
				Expected: []sql.Row{{"448"}},
			},
		},
	},
	{
		Name:        "set 'data_sync_retry' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW data_sync_retry",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET data_sync_retry TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('data_sync_retry')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'DateStyle' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW DateStyle",
				Expected: []sql.Row{{"ISO, MDY"}},
			},
			{
				Query:    "SET DateStyle TO 'ISO, DMY'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW DateStyle",
				Expected: []sql.Row{{"ISO, DMY"}},
			},
			{
				Query:    "SET DateStyle TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW DateStyle",
				Expected: []sql.Row{{"ISO, MDY"}},
			},
			{
				Query:    "SELECT current_setting('DateStyle')",
				Expected: []sql.Row{{"ISO, MDY"}},
			},
		},
	},
	{
		Name:        "set 'db_user_namespace' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW db_user_namespace",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET db_user_namespace TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('db_user_namespace')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'deadlock_timeout' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW deadlock_timeout",
				Expected: []sql.Row{{int64(1000)}},
			},
			{
				Query:    "SET deadlock_timeout TO '2000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW deadlock_timeout",
				Expected: []sql.Row{{int64(2000)}},
			},
			{
				Query:    "SET deadlock_timeout TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW deadlock_timeout",
				Expected: []sql.Row{{int64(1000)}},
			},
			{
				Query:    "SELECT current_setting('deadlock_timeout')",
				Expected: []sql.Row{{"1000"}},
			},
		},
	},
	{
		Name:        "set 'debug_assertions' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW debug_assertions",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET debug_assertions TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('debug_assertions')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'debug_discard_caches' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW debug_discard_caches",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET debug_discard_caches TO '0'", // cannot set it to anything other than 0
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW debug_discard_caches",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('debug_discard_caches')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'debug_io_direct' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW debug_io_direct",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET debug_io_direct TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('debug_io_direct')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'debug_logical_replication_streaming' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW debug_logical_replication_streaming",
				Expected: []sql.Row{{"buffered"}},
			},
			{
				Query:    "SET debug_logical_replication_streaming TO 'immediate'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW debug_logical_replication_streaming",
				Expected: []sql.Row{{"immediate"}},
			},
			{
				Query:    "SET debug_logical_replication_streaming TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW debug_logical_replication_streaming",
				Expected: []sql.Row{{"buffered"}},
			},
			{
				Query:    "SELECT current_setting('debug_logical_replication_streaming')",
				Expected: []sql.Row{{"buffered"}},
			},
		},
	},
	{
		Name:        "set 'debug_parallel_query' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW debug_parallel_query",
				Expected: []sql.Row{{"off"}},
			},
			{
				Query:    "SET debug_parallel_query TO 'regress'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW debug_parallel_query",
				Expected: []sql.Row{{"regress"}},
			},
			{
				Query:    "SET debug_parallel_query TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW debug_parallel_query",
				Expected: []sql.Row{{"off"}},
			},
			{
				Query:    "SELECT current_setting('debug_parallel_query')",
				Expected: []sql.Row{{"off"}},
			},
		},
	},
	{
		Name:        "set 'debug_pretty_print' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW debug_pretty_print",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET debug_pretty_print TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW debug_pretty_print",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET debug_pretty_print TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW debug_pretty_print",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('debug_pretty_print')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'debug_print_parse' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW debug_print_parse",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET debug_print_parse TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW debug_print_parse",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET debug_print_parse TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW debug_print_parse",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('debug_print_parse')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'debug_print_plan' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW debug_print_plan",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET debug_print_plan TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW debug_print_plan",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET debug_print_plan TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW debug_print_plan",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('debug_print_plan')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'debug_print_rewritten' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW debug_print_rewritten",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET debug_print_rewritten TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW debug_print_rewritten",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET debug_print_rewritten TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW debug_print_rewritten",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('debug_print_rewritten')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'default_statistics_target' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW default_statistics_target",
				Expected: []sql.Row{{int64(100)}},
			},
			{
				Query:    "SET default_statistics_target TO '10000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW default_statistics_target",
				Expected: []sql.Row{{int64(10000)}},
			},
			{
				Query:    "SET default_statistics_target TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW default_statistics_target",
				Expected: []sql.Row{{int64(100)}},
			},
			{
				Query:    "SELECT current_setting('default_statistics_target')",
				Expected: []sql.Row{{"100"}},
			},
		},
	},
	{
		Name:        "set 'default_table_access_method' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW default_table_access_method",
				Expected: []sql.Row{{"heap"}},
			},
			{
				Query:    "SET default_table_access_method TO 'heap'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW default_table_access_method",
				Expected: []sql.Row{{"heap"}},
			},
			{
				Query:    "SELECT current_setting('default_table_access_method')",
				Expected: []sql.Row{{"heap"}},
			},
		},
	},
	{
		Name:        "set 'default_tablespace' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW default_tablespace",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SET default_tablespace TO 'pg_default'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW default_tablespace",
				Expected: []sql.Row{{"pg_default"}},
			},
			{
				Query:    "SET default_tablespace TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW default_tablespace",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SELECT current_setting('default_tablespace')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'default_text_search_config' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW default_text_search_config",
				Expected: []sql.Row{{"pg_catalog.english"}},
			},
			{
				Query:    "SET default_text_search_config TO 'pg_catalog.spanish'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW default_text_search_config",
				Expected: []sql.Row{{"pg_catalog.spanish"}},
			},
			{
				Query:    "SET default_text_search_config TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW default_text_search_config",
				Expected: []sql.Row{{"pg_catalog.english"}},
			},
			{
				Query:    "SELECT current_setting('default_text_search_config')",
				Expected: []sql.Row{{"pg_catalog.english"}},
			},
		},
	},
	{
		Name:        "set 'default_toast_compression' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW default_toast_compression",
				Expected: []sql.Row{{"pglz"}},
			},
			{
				Query:    "SET default_toast_compression TO 'lz4'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW default_toast_compression",
				Expected: []sql.Row{{"lz4"}},
			},
			{
				Query:    "SET default_toast_compression TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW default_toast_compression",
				Expected: []sql.Row{{"pglz"}},
			},
			{
				Query:    "SELECT current_setting('default_toast_compression')",
				Expected: []sql.Row{{"pglz"}},
			},
		},
	},
	{
		Name:        "set 'default_transaction_deferrable' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW default_transaction_deferrable",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET default_transaction_deferrable TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW default_transaction_deferrable",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET default_transaction_deferrable TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW default_transaction_deferrable",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('default_transaction_deferrable')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'default_transaction_isolation' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW default_transaction_isolation",
				Expected: []sql.Row{{"read committed"}},
			},
			{
				Query:    "SET default_transaction_isolation TO 'serializable'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW default_transaction_isolation",
				Expected: []sql.Row{{"serializable"}},
			},
			{
				Query:    "SET default_transaction_isolation TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW default_transaction_isolation",
				Expected: []sql.Row{{"read committed"}},
			},
			{
				Query:    "SELECT current_setting('default_transaction_isolation')",
				Expected: []sql.Row{{"read committed"}},
			},
		},
	},
	{
		Name:        "set 'default_transaction_read_only' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW default_transaction_read_only",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET default_transaction_read_only TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW default_transaction_read_only",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET default_transaction_read_only TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW default_transaction_read_only",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('default_transaction_read_only')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'dynamic_library_path' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW dynamic_library_path",
				Expected: []sql.Row{{"$libdir"}},
			},
			{
				Query:    "SET dynamic_library_path TO ''",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW dynamic_library_path",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SET dynamic_library_path TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW dynamic_library_path",
				Expected: []sql.Row{{"$libdir"}},
			},
			{
				Query:    "SELECT current_setting('dynamic_library_path')",
				Expected: []sql.Row{{"$libdir"}},
			},
		},
	},
	{
		Name:        "set 'dynamic_shared_memory_type' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW dynamic_shared_memory_type",
				Expected: []sql.Row{{"posix"}},
			},
			{
				Query:       "SET dynamic_shared_memory_type TO 'posix'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('dynamic_shared_memory_type')",
				Expected: []sql.Row{{"posix"}},
			},
		},
	},
	{
		Name:        "set 'effective_cache_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW effective_cache_size",
				Expected: []sql.Row{{int64(524288)}},
			},
			{
				Query:    "SET effective_cache_size TO '400000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW effective_cache_size",
				Expected: []sql.Row{{int64(400000)}},
			},
			{
				Query:    "SET effective_cache_size TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW effective_cache_size",
				Expected: []sql.Row{{int64(524288)}},
			},
			{
				Query:    "SELECT current_setting('effective_cache_size')",
				Expected: []sql.Row{{"524288"}},
			},
		},
	},
	{
		Name:        "set 'effective_io_concurrency' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW effective_io_concurrency",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET effective_io_concurrency TO '100'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW effective_io_concurrency",
				Expected: []sql.Row{{int64(100)}},
			},
			{
				Query:    "SET effective_io_concurrency TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW effective_io_concurrency",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('effective_io_concurrency')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'enable_async_append' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_async_append",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_async_append TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_async_append",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_async_append TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_async_append",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_async_append')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_bitmapscan' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_bitmapscan",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_bitmapscan TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_bitmapscan",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_bitmapscan TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_bitmapscan",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_bitmapscan')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_gathermerge' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_gathermerge",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_gathermerge TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_gathermerge",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_gathermerge TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_gathermerge",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_gathermerge')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_hashagg' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_hashagg",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_hashagg TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_hashagg",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_hashagg TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_hashagg",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_hashagg')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_hashjoin' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_hashjoin",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_hashjoin TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_hashjoin",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_hashjoin TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_hashjoin",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_hashjoin')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_incremental_sort' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_incremental_sort",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_incremental_sort TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_incremental_sort",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_incremental_sort TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_incremental_sort",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_incremental_sort')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_indexonlyscan' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_indexonlyscan",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_indexonlyscan TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_indexonlyscan",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_indexonlyscan TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_indexonlyscan",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_indexonlyscan')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_indexscan' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_indexscan",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_indexscan TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_indexscan",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_indexscan TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_indexscan",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_indexscan')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_material' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_material",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_material TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_material",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_material TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_material",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_material')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_memoize' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_memoize",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_memoize TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_memoize",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_memoize TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_memoize",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_memoize')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_mergejoin' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_mergejoin",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_mergejoin TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_mergejoin",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_mergejoin TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_mergejoin",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_mergejoin')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_nestloop' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_nestloop",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_nestloop TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_nestloop",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_nestloop TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_nestloop",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_nestloop')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_parallel_append' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_parallel_append",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_parallel_append TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_parallel_append",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_parallel_append TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_parallel_append",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_parallel_append')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_parallel_hash' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_parallel_hash",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_parallel_hash TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_parallel_hash",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_parallel_hash TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_parallel_hash",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_parallel_hash')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_partition_pruning' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_partition_pruning",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_partition_pruning TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_partition_pruning",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_partition_pruning TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_partition_pruning",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_partition_pruning')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_partitionwise_aggregate' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_partitionwise_aggregate",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_partitionwise_aggregate TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_partitionwise_aggregate",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_partitionwise_aggregate TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_partitionwise_aggregate",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('enable_partitionwise_aggregate')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'enable_partitionwise_join' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_partitionwise_join",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_partitionwise_join TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_partitionwise_join",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_partitionwise_join TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_partitionwise_join",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('enable_partitionwise_join')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'enable_presorted_aggregate' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_presorted_aggregate",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_presorted_aggregate TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_presorted_aggregate",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_presorted_aggregate TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_presorted_aggregate",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_presorted_aggregate')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_seqscan' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_seqscan",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_seqscan TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_seqscan",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_seqscan TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_seqscan",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_seqscan')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_sort' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_sort",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_sort TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_sort",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_sort TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_sort",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_sort')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'enable_tidscan' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW enable_tidscan",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET enable_tidscan TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_tidscan",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET enable_tidscan TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW enable_tidscan",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('enable_tidscan')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'escape_string_warning' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW escape_string_warning",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET escape_string_warning TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW escape_string_warning",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET escape_string_warning TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW escape_string_warning",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('escape_string_warning')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'event_source' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW event_source",
				Expected: []sql.Row{{"PostgreSQL"}},
			},
			{
				Query:       "SET event_source TO 'PostgreSQL'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('event_source')",
				Expected: []sql.Row{{"PostgreSQL"}},
			},
		},
	},
	{
		Name:        "set 'exit_on_error' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW exit_on_error",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET exit_on_error TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW exit_on_error",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET exit_on_error TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW exit_on_error",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('exit_on_error')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'external_pid_file' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW external_pid_file",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET external_pid_file TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('external_pid_file')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'extra_float_digits' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW extra_float_digits",
				Expected: []sql.Row{{int64(1)}},
			},
			{
				Query:    "SET extra_float_digits TO -10",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW extra_float_digits",
				Expected: []sql.Row{{int64(-10)}},
			},
			{
				Query:    "SET extra_float_digits TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW extra_float_digits",
				Expected: []sql.Row{{int64(1)}},
			},
			{
				Query:    "SELECT current_setting('extra_float_digits')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'from_collapse_limit' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW from_collapse_limit",
				Expected: []sql.Row{{int64(8)}},
			},
			{
				Query:    "SET from_collapse_limit TO 100",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW from_collapse_limit",
				Expected: []sql.Row{{int64(100)}},
			},
			{
				Query:    "SET from_collapse_limit TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW from_collapse_limit",
				Expected: []sql.Row{{int64(8)}},
			},
			{
				Query:    "SELECT current_setting('from_collapse_limit')",
				Expected: []sql.Row{{"8"}},
			},
		},
	},
	{
		Name:        "set 'fsync' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW fsync",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:       "SET fsync TO 'on'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('fsync')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'full_page_writes' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW full_page_writes",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:       "SET full_page_writes TO 'on'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('full_page_writes')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'geqo' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW geqo",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET geqo TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW geqo",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET geqo TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW geqo",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('geqo')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'geqo_effort' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW geqo_effort",
				Expected: []sql.Row{{int64(5)}},
			},
			{
				Query:    "SET geqo_effort TO 10",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW geqo_effort",
				Expected: []sql.Row{{int64(10)}},
			},
			{
				Query:    "SET geqo_effort TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW geqo_effort",
				Expected: []sql.Row{{int64(5)}},
			},
			{
				Query:    "SELECT current_setting('geqo_effort')",
				Expected: []sql.Row{{"5"}},
			},
		},
	},
	{
		Name:        "set 'geqo_generations' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW geqo_generations",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET geqo_generations TO '100'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW geqo_generations",
				Expected: []sql.Row{{int64(100)}},
			},
			{
				Query:    "SET geqo_generations TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW geqo_generations",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('geqo_generations')",
				Expected: []sql.Row{{"0"}},
			},
			{
				Query:    "SELECT current_setting('geqo_generations')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'geqo_pool_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW geqo_pool_size",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET geqo_pool_size TO 1",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW geqo_pool_size",
				Expected: []sql.Row{{int64(1)}},
			},
			{
				Query:    "SET geqo_pool_size TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW geqo_pool_size",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('geqo_pool_size')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'geqo_seed' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW geqo_seed",
				Expected: []sql.Row{{float64(0)}},
			},
			{
				Query:    "SET geqo_seed TO 0.2",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW geqo_seed",
				Expected: []sql.Row{{float64(0.2)}},
			},
			{
				Query:    "SET geqo_seed TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW geqo_seed",
				Expected: []sql.Row{{float64(0)}},
			},
			{
				Query:    "SELECT current_setting('geqo_seed')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'geqo_selection_bias' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW geqo_selection_bias",
				Expected: []sql.Row{{float64(2)}},
			},
			{
				Query:    "SET geqo_selection_bias TO 1.7",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW geqo_selection_bias",
				Expected: []sql.Row{{float64(1.7)}},
			},
			{
				Query:    "SET geqo_selection_bias TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW geqo_selection_bias",
				Expected: []sql.Row{{float64(2)}},
			},
			{
				Query:    "SELECT current_setting('geqo_selection_bias')",
				Expected: []sql.Row{{"2"}},
			},
		},
	},
	{
		Name:        "set 'geqo_threshold' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW geqo_threshold",
				Expected: []sql.Row{{int64(12)}},
			},
			{
				Query:    "SET geqo_threshold TO 22",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW geqo_threshold",
				Expected: []sql.Row{{int64(22)}},
			},
			{
				Query:    "SET geqo_threshold TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW geqo_threshold",
				Expected: []sql.Row{{int64(12)}},
			},
			{
				Query:    "SELECT current_setting('geqo_threshold')",
				Expected: []sql.Row{{"12"}},
			},
		},
	},
	{
		Name:        "set 'gin_fuzzy_search_limit' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW gin_fuzzy_search_limit",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET gin_fuzzy_search_limit TO 2",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW gin_fuzzy_search_limit",
				Expected: []sql.Row{{int64(2)}},
			},
			{
				Query:    "SET gin_fuzzy_search_limit TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW gin_fuzzy_search_limit",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('gin_fuzzy_search_limit')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'gin_pending_list_limit' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW gin_pending_list_limit",
				Expected: []sql.Row{{int64(4096)}},
			},
			{
				Query:    "SET gin_pending_list_limit TO '4000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW gin_pending_list_limit",
				Expected: []sql.Row{{int64(4000)}},
			},
			{
				Query:    "SET gin_pending_list_limit TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW gin_pending_list_limit",
				Expected: []sql.Row{{int64(4096)}},
			},
			{
				Query:    "SELECT current_setting('gin_pending_list_limit')",
				Expected: []sql.Row{{"4096"}},
			},
		},
	},
	{
		Name:        "set 'gss_accept_delegation' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW gss_accept_delegation",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET gss_accept_delegation TO 'on'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('gss_accept_delegation')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'hash_mem_multiplier' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW hash_mem_multiplier",
				Expected: []sql.Row{{float64(2)}},
			},
			{
				Query:    "SET hash_mem_multiplier TO 20.1",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW hash_mem_multiplier",
				Expected: []sql.Row{{float64(20.1)}},
			},
			{
				Query:    "SET hash_mem_multiplier TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW hash_mem_multiplier",
				Expected: []sql.Row{{float64(2)}},
			},
			{
				Query:    "SELECT current_setting('hash_mem_multiplier')",
				Expected: []sql.Row{{"2"}},
			},
		},
	},
	{
		Name:        "set 'hba_file' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW hba_file",
				Expected: []sql.Row{{"pg_hba.conf"}},
			},
			{
				Query:       "SET hba_file TO '/Users/postgres/pg_hba.conf'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('hba_file')",
				Expected: []sql.Row{{"pg_hba.conf"}},
			},
		},
	},
	{
		Name:        "set 'hot_standby' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW hot_standby",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:       "SET hot_standby TO 'on'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('hot_standby')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'hot_standby_feedback' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW hot_standby_feedback",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET hot_standby_feedback TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('hot_standby_feedback')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'huge_page_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW huge_page_size",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:       "SET huge_page_size TO '0'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('huge_page_size')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'huge_pages' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW huge_pages",
				Expected: []sql.Row{{"try"}},
			},
			{
				Query:       "SET huge_pages TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('huge_pages')",
				Expected: []sql.Row{{"try"}},
			},
		},
	},
	{
		Name:        "set 'icu_validation_level' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW icu_validation_level",
				Expected: []sql.Row{{"warning"}},
			},
			{
				Query:    "SET icu_validation_level TO 'disabled'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW icu_validation_level",
				Expected: []sql.Row{{"disabled"}},
			},
			{
				Query:    "SET icu_validation_level TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW icu_validation_level",
				Expected: []sql.Row{{"warning"}},
			},
			{
				Query:    "SELECT current_setting('icu_validation_level')",
				Expected: []sql.Row{{"warning"}},
			},
		},
	},
	{
		Name:        "set 'ident_file' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ident_file",
				Expected: []sql.Row{{"pg_ident.conf"}},
			},
			{
				Query:       "SET ident_file TO '/Users/postgres/pg_ident.conf'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ident_file')",
				Expected: []sql.Row{{"pg_ident.conf"}},
			},
		},
	},
	{
		Name:        "set 'idle_in_transaction_session_timeout' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW idle_in_transaction_session_timeout",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET idle_in_transaction_session_timeout TO 2",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW idle_in_transaction_session_timeout",
				Expected: []sql.Row{{int64(2)}},
			},
			{
				Query:    "SET idle_in_transaction_session_timeout TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW idle_in_transaction_session_timeout",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('idle_in_transaction_session_timeout')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'idle_session_timeout' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW idle_session_timeout",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET idle_session_timeout TO '3'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW idle_session_timeout",
				Expected: []sql.Row{{int64(3)}},
			},
			{
				Query:    "SET idle_session_timeout TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW idle_session_timeout",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('idle_session_timeout')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'ignore_checksum_failure' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ignore_checksum_failure",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET ignore_checksum_failure TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW ignore_checksum_failure",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET ignore_checksum_failure TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW ignore_checksum_failure",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('ignore_checksum_failure')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'ignore_invalid_pages' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ignore_invalid_pages",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET ignore_invalid_pages TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ignore_invalid_pages')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'ignore_system_indexes' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ignore_system_indexes",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET ignore_system_indexes TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ignore_system_indexes')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'in_hot_standby' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW in_hot_standby",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET in_hot_standby TO 'on'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('in_hot_standby')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'integer_datetimes' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW integer_datetimes",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:       "SET integer_datetimes TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('integer_datetimes')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'IntervalStyle' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW IntervalStyle",
				Expected: []sql.Row{{"postgres"}},
			},
			{
				Query:    "SET IntervalStyle TO 'sql_standard'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW IntervalStyle",
				Expected: []sql.Row{{"sql_standard"}},
			},
			{
				Query:    "SET IntervalStyle TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW IntervalStyle",
				Expected: []sql.Row{{"postgres"}},
			},
			{
				Query:    "SELECT current_setting('IntervalStyle')",
				Expected: []sql.Row{{"postgres"}},
			},
		},
	},
	{
		Name:        "set 'jit' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW jit",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET jit TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW jit",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET jit TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW jit",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('jit')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'jit_above_cost' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW jit_above_cost",
				Expected: []sql.Row{{float64(100000)}},
			},
			{
				Query:    "SET jit_above_cost TO '100'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW jit_above_cost",
				Expected: []sql.Row{{float64(100)}},
			},
			{
				Query:    "SET jit_above_cost TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW jit_above_cost",
				Expected: []sql.Row{{float64(100000)}},
			},
			{
				Query:    "SELECT current_setting('jit_above_cost')",
				Expected: []sql.Row{{"100000"}},
			},
		},
	},
	{
		Name:        "set 'jit_debugging_support' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW jit_debugging_support",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET jit_debugging_support TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('jit_debugging_support')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'jit_dump_bitcode' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW jit_dump_bitcode",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET jit_dump_bitcode TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW jit_dump_bitcode",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET jit_dump_bitcode TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW jit_dump_bitcode",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('jit_dump_bitcode')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'jit_expressions' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW jit_expressions",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET jit_expressions TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW jit_expressions",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET jit_expressions TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW jit_expressions",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('jit_expressions')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'jit_inline_above_cost' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW jit_inline_above_cost",
				Expected: []sql.Row{{float64(500000)}},
			},
			{
				Query:    "SET jit_inline_above_cost TO '5000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW jit_inline_above_cost",
				Expected: []sql.Row{{float64(5000)}},
			},
			{
				Query:    "SET jit_inline_above_cost TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW jit_inline_above_cost",
				Expected: []sql.Row{{float64(500000)}},
			},
			{
				Query:    "SELECT current_setting('jit_inline_above_cost')",
				Expected: []sql.Row{{"500000"}},
			},
		},
	},
	{
		Name:        "set 'jit_optimize_above_cost' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW jit_optimize_above_cost",
				Expected: []sql.Row{{float64(500000)}},
			},
			{
				Query:    "SET jit_optimize_above_cost TO '5000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW jit_optimize_above_cost",
				Expected: []sql.Row{{float64(5000)}},
			},
			{
				Query:    "SET jit_optimize_above_cost TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW jit_optimize_above_cost",
				Expected: []sql.Row{{float64(500000)}},
			},
			{
				Query:    "SELECT current_setting('jit_optimize_above_cost')",
				Expected: []sql.Row{{"500000"}},
			},
		},
	},
	{
		Name:        "set 'jit_profiling_support' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW jit_profiling_support",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET jit_profiling_support TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('jit_profiling_support')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'jit_provider' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW jit_provider",
				Expected: []sql.Row{{"llvmjit"}},
			},
			{
				Query:       "SET jit_provider TO 'llvmjit'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('jit_provider')",
				Expected: []sql.Row{{"llvmjit"}},
			},
		},
	},
	{
		Name:        "set 'jit_tuple_deforming' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW jit_tuple_deforming",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET jit_tuple_deforming TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW jit_tuple_deforming",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET jit_tuple_deforming TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW jit_tuple_deforming",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('jit_tuple_deforming')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'join_collapse_limit' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW join_collapse_limit",
				Expected: []sql.Row{{int64(8)}},
			},
			{
				Query:    "SET join_collapse_limit TO '100'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW join_collapse_limit",
				Expected: []sql.Row{{int64(100)}},
			},
			{
				Query:    "SET join_collapse_limit TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW join_collapse_limit",
				Expected: []sql.Row{{int64(8)}},
			},
			{
				Query:    "SELECT current_setting('join_collapse_limit')",
				Expected: []sql.Row{{"8"}},
			},
		},
	},
	{
		Name:        "set 'krb_caseins_users' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW krb_caseins_users",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET krb_caseins_users TO 'on'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('krb_caseins_users')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'krb_server_keyfile' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW krb_server_keyfile",
				Expected: []sql.Row{{"FILE:/usr/local/etc/postgresql/krb5.keytab"}},
			},
			{
				Query:       "SET krb_server_keyfile TO 'FILE:'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('krb_server_keyfile')",
				Expected: []sql.Row{{"FILE:/usr/local/etc/postgresql/krb5.keytab"}},
			},
		},
	},
	{
		Name:        "set 'lc_messages' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW lc_messages",
				Expected: []sql.Row{{"en_US.UTF-8"}},
			},
			{
				Query:    "SET lc_messages TO 'en_US'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW lc_messages",
				Expected: []sql.Row{{"en_US"}},
			},
			{
				Query:    "SET lc_messages TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW lc_messages",
				Expected: []sql.Row{{"en_US.UTF-8"}},
			},
			{
				Query:    "SELECT current_setting('lc_messages')",
				Expected: []sql.Row{{"en_US.UTF-8"}},
			},
		},
	},
	{
		Name:        "set 'lc_monetary' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW lc_monetary",
				Expected: []sql.Row{{"en_US.UTF-8"}},
			},
			{
				Query:    "SET lc_monetary TO 'en_US'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW lc_monetary",
				Expected: []sql.Row{{"en_US"}},
			},
			{
				Query:    "SET lc_monetary TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW lc_monetary",
				Expected: []sql.Row{{"en_US.UTF-8"}},
			},
			{
				Query:    "SELECT current_setting('lc_monetary')",
				Expected: []sql.Row{{"en_US.UTF-8"}},
			},
		},
	},
	{
		Name:        "set 'lc_numeric' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW lc_numeric",
				Expected: []sql.Row{{"en_US.UTF-8"}},
			},
			{
				Query:    "SET lc_numeric TO 'en_US'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW lc_numeric",
				Expected: []sql.Row{{"en_US"}},
			},
			{
				Query:    "SET lc_numeric TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW lc_numeric",
				Expected: []sql.Row{{"en_US.UTF-8"}},
			},
			{
				Query:    "SELECT current_setting('lc_numeric')",
				Expected: []sql.Row{{"en_US.UTF-8"}},
			},
		},
	},
	{
		Name:        "set 'lc_time' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW lc_time",
				Expected: []sql.Row{{"en_US.UTF-8"}},
			},
			{
				Query:    "SET lc_time TO 'en_US'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW lc_time",
				Expected: []sql.Row{{"en_US"}},
			},
			{
				Query:    "SET lc_time TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW lc_time",
				Expected: []sql.Row{{"en_US.UTF-8"}},
			},
			{
				Query:    "SELECT current_setting('lc_time')",
				Expected: []sql.Row{{"en_US.UTF-8"}},
			},
		},
	},
	{
		Name:        "set 'listen_addresses' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW listen_addresses",
				Expected: []sql.Row{{"localhost"}},
			},
			{
				Query:       "SET listen_addresses TO 'localhost'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('listen_addresses')",
				Expected: []sql.Row{{"localhost"}},
			},
		},
	},
	{
		Name:        "set 'lo_compat_privileges' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW lo_compat_privileges",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET lo_compat_privileges TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW lo_compat_privileges",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET lo_compat_privileges TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW lo_compat_privileges",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('lo_compat_privileges')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'local_preload_libraries' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW local_preload_libraries",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SET local_preload_libraries TO '/'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW local_preload_libraries",
				Expected: []sql.Row{{"/"}},
			},
			{
				Query:    "SET local_preload_libraries TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW local_preload_libraries",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SELECT current_setting('local_preload_libraries')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'lock_timeout' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW lock_timeout",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET lock_timeout TO 20",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW lock_timeout",
				Expected: []sql.Row{{int64(20)}},
			},
			{
				Query:    "SET lock_timeout TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW lock_timeout",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('lock_timeout')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'log_autovacuum_min_duration' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_autovacuum_min_duration",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:       "SET log_autovacuum_min_duration TO '600'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_autovacuum_min_duration')",
				Expected: []sql.Row{{"-1"}},
			},
		},
	},
	{
		Name:        "set 'log_checkpoints' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_checkpoints",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:       "SET log_checkpoints TO 'on'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_checkpoints')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'log_connections' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_connections",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET log_connections TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_connections')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'log_destination' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_destination",
				Expected: []sql.Row{{"stderr"}},
			},
			{
				Query:       "SET log_destination TO 'jsonlog'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_destination')",
				Expected: []sql.Row{{"stderr"}},
			},
		},
	},
	{
		Name:        "set 'log_directory' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_directory",
				Expected: []sql.Row{{"log"}},
			},
			{
				Query:       "SET log_directory TO 'log'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_directory')",
				Expected: []sql.Row{{"log"}},
			},
		},
	},
	{
		Name:        "set 'log_disconnections' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_disconnections",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET log_disconnections TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_disconnections')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'log_duration' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_duration",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET log_duration TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_duration",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET log_duration TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_duration",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('log_duration')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'log_error_verbosity' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_error_verbosity",
				Expected: []sql.Row{{"default"}},
			},
			{
				Query:    "SET log_error_verbosity TO 'verbose'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_error_verbosity",
				Expected: []sql.Row{{"verbose"}},
			},
			{
				Query:    "SET log_error_verbosity TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_error_verbosity",
				Expected: []sql.Row{{"default"}},
			},
			{
				Query:    "SELECT current_setting('log_error_verbosity')",
				Expected: []sql.Row{{"default"}},
			},
		},
	},
	{
		Name:        "set 'log_executor_stats' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_executor_stats",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET log_executor_stats TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_executor_stats",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET log_executor_stats TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_executor_stats",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('log_executor_stats')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'log_file_mode' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_file_mode",
				Expected: []sql.Row{{int64(384)}},
			},
			{
				Query:       "SET log_file_mode TO '384'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_file_mode')",
				Expected: []sql.Row{{"384"}},
			},
		},
	},
	{
		Name:        "set 'log_filename' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_filename",
				Expected: []sql.Row{{"postgresql-%Y-%m-%d_%H%M%S.log"}},
			},
			{
				Query:       "SET log_filename TO 'postgresql-%Y-%m-%d_%H%M%S.log'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_filename')",
				Expected: []sql.Row{{"postgresql-%Y-%m-%d_%H%M%S.log"}},
			},
		},
	},
	{
		Name:        "set 'log_hostname' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_hostname",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET log_hostname TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_hostname')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'log_line_prefix' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_line_prefix",
				Expected: []sql.Row{{"%m [%p]"}},
			},
			{
				Query:       "SET log_line_prefix TO '%m [%p]'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_line_prefix')",
				Expected: []sql.Row{{"%m [%p]"}},
			},
		},
	},
	{
		Name:        "set 'log_lock_waits' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_lock_waits",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET log_lock_waits TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_lock_waits",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET log_lock_waits TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_lock_waits",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('log_lock_waits')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'log_min_duration_sample' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_min_duration_sample",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:    "SET log_min_duration_sample TO 1",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_min_duration_sample",
				Expected: []sql.Row{{int64(1)}},
			},
			{
				Query:    "SET log_min_duration_sample TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_min_duration_sample",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:    "SELECT current_setting('log_min_duration_sample')",
				Expected: []sql.Row{{"-1"}},
			},
		},
	},
	{
		Name:        "set 'log_min_duration_statement' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_min_duration_statement",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:    "SET log_min_duration_statement TO 10",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_min_duration_statement",
				Expected: []sql.Row{{int64(10)}},
			},
			{
				Query:    "SET log_min_duration_statement TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_min_duration_statement",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:    "SELECT current_setting('log_min_duration_statement')",
				Expected: []sql.Row{{"-1"}},
			},
		},
	},
	{
		Name:        "set 'log_min_error_statement' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_min_error_statement",
				Expected: []sql.Row{{"error"}},
			},
			{
				Query:    "SET log_min_error_statement TO 'debug5'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_min_error_statement",
				Expected: []sql.Row{{"debug5"}},
			},
			{
				Query:    "SET log_min_error_statement TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_min_error_statement",
				Expected: []sql.Row{{"error"}},
			},
			{
				Query:    "SELECT current_setting('log_min_error_statement')",
				Expected: []sql.Row{{"error"}},
			},
		},
	},
	{
		Name:        "set 'log_min_messages' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_min_messages",
				Expected: []sql.Row{{"warning"}},
			},
			{
				Query:    "SET log_min_messages TO 'info'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_min_messages",
				Expected: []sql.Row{{"info"}},
			},
			{
				Query:    "SET log_min_messages TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_min_messages",
				Expected: []sql.Row{{"warning"}},
			},
			{
				Query:    "SELECT current_setting('log_min_messages')",
				Expected: []sql.Row{{"warning"}},
			},
		},
	},
	{
		Name:        "set 'log_parameter_max_length' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_parameter_max_length",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:    "SET log_parameter_max_length TO '10'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_parameter_max_length",
				Expected: []sql.Row{{int64(10)}},
			},
			{
				Query:    "SELECT current_setting('log_parameter_max_length')",
				Expected: []sql.Row{{"10"}},
			},
			{
				Query:    "SET log_parameter_max_length TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_parameter_max_length",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:    "SELECT current_setting('log_parameter_max_length')",
				Expected: []sql.Row{{"-1"}},
			},
		},
	},
	{
		Name:        "set 'log_parameter_max_length_on_error' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_parameter_max_length_on_error",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET log_parameter_max_length_on_error TO '1'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_parameter_max_length_on_error",
				Expected: []sql.Row{{int64(1)}},
			},
			{
				Query:    "SET log_parameter_max_length_on_error TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_parameter_max_length_on_error",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('log_parameter_max_length_on_error')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'log_parser_stats' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_parser_stats",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET log_parser_stats TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_parser_stats",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET log_parser_stats TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_parser_stats",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('log_parser_stats')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'log_planner_stats' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_planner_stats",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET log_planner_stats TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_planner_stats",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET log_planner_stats TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_planner_stats",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('log_planner_stats')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'log_recovery_conflict_waits' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_recovery_conflict_waits",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET log_recovery_conflict_waits TO 'on'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_recovery_conflict_waits')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'log_replication_commands' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_replication_commands",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET log_replication_commands TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_replication_commands",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET log_replication_commands TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_replication_commands",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('log_replication_commands')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'log_rotation_age' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_rotation_age",
				Expected: []sql.Row{{int64(1440)}},
			},
			{
				Query:       "SET log_rotation_age TO '1440'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_rotation_age')",
				Expected: []sql.Row{{"1440"}},
			},
		},
	},
	{
		Name:        "set 'log_rotation_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_rotation_size",
				Expected: []sql.Row{{int64(10240)}},
			},
			{
				Query:       "SET log_rotation_size TO '10240'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_rotation_size')",
				Expected: []sql.Row{{"10240"}},
			},
		},
	},
	{
		Name:        "set 'log_startup_progress_interval' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_startup_progress_interval",
				Expected: []sql.Row{{int64(10000)}},
			},
			{
				Query:       "SET log_startup_progress_interval TO '10'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_startup_progress_interval')",
				Expected: []sql.Row{{"10000"}},
			},
		},
	},
	{
		Name:        "set 'log_statement' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_statement",
				Expected: []sql.Row{{"none"}},
			},
			{
				Query:    "SET log_statement TO 'ddl'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_statement",
				Expected: []sql.Row{{"ddl"}},
			},
			{
				Query:    "SET log_statement TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_statement",
				Expected: []sql.Row{{"none"}},
			},
			{
				Query:    "SELECT current_setting('log_statement')",
				Expected: []sql.Row{{"none"}},
			},
		},
	},
	{
		Name:        "set 'log_statement_sample_rate' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_statement_sample_rate",
				Expected: []sql.Row{{float64(1)}},
			},
			{
				Query:    "SET log_statement_sample_rate TO 0.5",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_statement_sample_rate",
				Expected: []sql.Row{{float64(0.5)}},
			},
			{
				Query:    "SET log_statement_sample_rate TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_statement_sample_rate",
				Expected: []sql.Row{{float64(1)}},
			},
			{
				Query:    "SELECT current_setting('log_statement_sample_rate')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'log_statement_stats' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_statement_stats",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET log_statement_stats TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_statement_stats",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET log_statement_stats TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_statement_stats",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('log_statement_stats')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'log_temp_files' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_temp_files",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:    "SET log_temp_files TO '100'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_temp_files",
				Expected: []sql.Row{{int64(100)}},
			},
			{
				Query:    "SET log_temp_files TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_temp_files",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:    "SELECT current_setting('log_temp_files')",
				Expected: []sql.Row{{"-1"}},
			},
		},
	},
	{
		Name:        "set 'log_timezone' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_timezone",
				Expected: []sql.Row{{"America/Los_Angeles"}},
			},
			{
				Query:       "SET log_timezone TO 'America/Los_Angeles'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_timezone')",
				Expected: []sql.Row{{"America/Los_Angeles"}},
			},
		},
	},
	{
		Name:        "set 'log_transaction_sample_rate' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_transaction_sample_rate",
				Expected: []sql.Row{{float64(0)}},
			},
			{
				Query:    "SET log_transaction_sample_rate TO '0.5'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_transaction_sample_rate",
				Expected: []sql.Row{{float64(0.5)}},
			},
			{
				Query:    "SET log_transaction_sample_rate TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW log_transaction_sample_rate",
				Expected: []sql.Row{{float64(0)}},
			},
			{
				Query:    "SELECT current_setting('log_transaction_sample_rate')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'log_truncate_on_rotation' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW log_truncate_on_rotation",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET log_truncate_on_rotation TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('log_truncate_on_rotation')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'logging_collector' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW logging_collector",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET logging_collector TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('logging_collector')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'logical_decoding_work_mem' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW logical_decoding_work_mem",
				Expected: []sql.Row{{int64(65536)}},
			},
			{
				Query:    "SET logical_decoding_work_mem TO '64000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW logical_decoding_work_mem",
				Expected: []sql.Row{{int64(64000)}},
			},
			{
				Query:    "SET logical_decoding_work_mem TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW logical_decoding_work_mem",
				Expected: []sql.Row{{int64(65536)}},
			},
			{
				Query:    "SELECT current_setting('logical_decoding_work_mem')",
				Expected: []sql.Row{{"65536"}},
			},
		},
	},
	{
		Name:        "set 'maintenance_io_concurrency' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW maintenance_io_concurrency",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET maintenance_io_concurrency TO '1'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW maintenance_io_concurrency",
				Expected: []sql.Row{{int64(1)}},
			},
			{
				Query:    "SET maintenance_io_concurrency TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW maintenance_io_concurrency",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('maintenance_io_concurrency')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'maintenance_work_mem' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW maintenance_work_mem",
				Expected: []sql.Row{{int64(65536)}},
			},
			{
				Query:    "SET maintenance_work_mem TO '64000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW maintenance_work_mem",
				Expected: []sql.Row{{int64(64000)}},
			},
			{
				Query:    "SET maintenance_work_mem TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW maintenance_work_mem",
				Expected: []sql.Row{{int64(65536)}},
			},
			{
				Query:    "SELECT current_setting('maintenance_work_mem')",
				Expected: []sql.Row{{"65536"}},
			},
		},
	},
	{
		Name:        "set 'max_connections' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_connections",
				Expected: []sql.Row{{int64(100)}},
			},
			{
				Query:       "SET max_connections TO '150'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_connections')",
				Expected: []sql.Row{{"100"}},
			},
		},
	},
	{
		Name:        "set 'max_files_per_process' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_files_per_process",
				Expected: []sql.Row{{int64(1000)}},
			},
			{
				Query:       "SET max_files_per_process TO '1000'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_files_per_process')",
				Expected: []sql.Row{{"1000"}},
			},
		},
	},
	{
		Name:        "set 'max_function_args' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_function_args",
				Expected: []sql.Row{{int64(100)}},
			},
			{
				Query:       "SET max_function_args TO '100'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_function_args')",
				Expected: []sql.Row{{"100"}},
			},
		},
	},
	{
		Name:        "set 'max_identifier_length' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_identifier_length",
				Expected: []sql.Row{{int64(63)}},
			},
			{
				Query:       "SET max_identifier_length TO '63'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_identifier_length')",
				Expected: []sql.Row{{"63"}},
			},
		},
	},
	{
		Name:        "set 'max_index_keys' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_index_keys",
				Expected: []sql.Row{{int64(32)}},
			},
			{
				Query:       "SET max_index_keys TO '32'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_index_keys')",
				Expected: []sql.Row{{"32"}},
			},
		},
	},
	{
		Name:        "set 'max_locks_per_transaction' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_locks_per_transaction",
				Expected: []sql.Row{{int64(64)}},
			},
			{
				Query:       "SET max_locks_per_transaction TO '64'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_locks_per_transaction')",
				Expected: []sql.Row{{"64"}},
			},
		},
	},
	{
		Name:        "set 'max_logical_replication_workers' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_logical_replication_workers",
				Expected: []sql.Row{{int64(4)}},
			},
			{
				Query:       "SET max_logical_replication_workers TO '4'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_logical_replication_workers')",
				Expected: []sql.Row{{"4"}},
			},
		},
	},
	{
		Name:        "set 'max_parallel_apply_workers_per_subscription' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_parallel_apply_workers_per_subscription",
				Expected: []sql.Row{{int64(2)}},
			},
			{
				Query:       "SET max_parallel_apply_workers_per_subscription TO '2'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_parallel_apply_workers_per_subscription')",
				Expected: []sql.Row{{"2"}},
			},
		},
	},
	{
		Name:        "set 'max_parallel_maintenance_workers' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_parallel_maintenance_workers",
				Expected: []sql.Row{{int64(2)}},
			},
			{
				Query:    "SET max_parallel_maintenance_workers TO '3'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW max_parallel_maintenance_workers",
				Expected: []sql.Row{{int64(3)}},
			},
			{
				Query:    "SET max_parallel_maintenance_workers TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW max_parallel_maintenance_workers",
				Expected: []sql.Row{{int64(2)}},
			},
			{
				Query:    "SELECT current_setting('max_parallel_maintenance_workers')",
				Expected: []sql.Row{{"2"}},
			},
		},
	},
	{
		Name:        "set 'max_parallel_workers' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_parallel_workers",
				Expected: []sql.Row{{int64(8)}},
			},
			{
				Query:    "SET max_parallel_workers TO 11",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW max_parallel_workers",
				Expected: []sql.Row{{int64(11)}},
			},
			{
				Query:    "SET max_parallel_workers TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW max_parallel_workers",
				Expected: []sql.Row{{int64(8)}},
			},
			{
				Query:    "SELECT current_setting('max_parallel_workers')",
				Expected: []sql.Row{{"8"}},
			},
		},
	},
	{
		Name:        "set 'max_parallel_workers_per_gather' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_parallel_workers_per_gather",
				Expected: []sql.Row{{int64(2)}},
			},
			{
				Query:    "SET max_parallel_workers_per_gather TO 3",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW max_parallel_workers_per_gather",
				Expected: []sql.Row{{int64(3)}},
			},
			{
				Query:    "SET max_parallel_workers_per_gather TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW max_parallel_workers_per_gather",
				Expected: []sql.Row{{int64(2)}},
			},
			{
				Query:    "SELECT current_setting('max_parallel_workers_per_gather')",
				Expected: []sql.Row{{"2"}},
			},
		},
	},
	{
		Name:        "set 'max_pred_locks_per_page' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_pred_locks_per_page",
				Expected: []sql.Row{{int64(2)}},
			},
			{
				Query:       "SET max_pred_locks_per_page TO '2'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_pred_locks_per_page')",
				Expected: []sql.Row{{"2"}},
			},
		},
	},
	{
		Name:        "set 'max_pred_locks_per_relation' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_pred_locks_per_relation",
				Expected: []sql.Row{{int64(-2)}},
			},
			{
				Query:       "SET max_pred_locks_per_relation TO '-2'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_pred_locks_per_relation')",
				Expected: []sql.Row{{"-2"}},
			},
		},
	},
	{
		Name:        "set 'max_pred_locks_per_transaction' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_pred_locks_per_transaction",
				Expected: []sql.Row{{int64(64)}},
			},
			{
				Query:       "SET max_pred_locks_per_transaction TO '64'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_pred_locks_per_transaction')",
				Expected: []sql.Row{{"64"}},
			},
		},
	},
	{
		Name:        "set 'max_prepared_transactions' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_prepared_transactions",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:       "SET max_prepared_transactions TO '0'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_prepared_transactions')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'max_replication_slots' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_replication_slots",
				Expected: []sql.Row{{int64(10)}},
			},
			{
				Query:       "SET max_replication_slots TO '10'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_replication_slots')",
				Expected: []sql.Row{{"10"}},
			},
		},
	},
	{
		Name:        "set 'max_slot_wal_keep_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_slot_wal_keep_size",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:       "SET max_slot_wal_keep_size TO '-1'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_slot_wal_keep_size')",
				Expected: []sql.Row{{"-1"}},
			},
		},
	},
	{
		Name:        "set 'max_stack_depth' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_stack_depth",
				Expected: []sql.Row{{int64(2048)}},
			},
			{
				Query:    "SET max_stack_depth TO '2000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW max_stack_depth",
				Expected: []sql.Row{{int64(2000)}},
			},
			{
				Query:    "SET max_stack_depth TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW max_stack_depth",
				Expected: []sql.Row{{int64(2048)}},
			},
			{
				Query:    "SELECT current_setting('max_stack_depth')",
				Expected: []sql.Row{{"2048"}},
			},
		},
	},
	{
		Name:        "set 'max_standby_archive_delay' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_standby_archive_delay",
				Expected: []sql.Row{{int64(30000)}},
			},
			{
				Query:       "SET max_standby_archive_delay TO '30'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_standby_archive_delay')",
				Expected: []sql.Row{{"30000"}},
			},
		},
	},
	{
		Name:        "set 'max_standby_streaming_delay' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_standby_streaming_delay",
				Expected: []sql.Row{{int64(30000)}},
			},
			{
				Query:       "SET max_standby_streaming_delay TO '30'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_standby_streaming_delay')",
				Expected: []sql.Row{{"30000"}},
			},
		},
	},
	{
		Name:        "set 'max_sync_workers_per_subscription' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_sync_workers_per_subscription",
				Expected: []sql.Row{{int64(2)}},
			},
			{
				Query:       "SET max_sync_workers_per_subscription TO '2'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_sync_workers_per_subscription')",
				Expected: []sql.Row{{"2"}},
			},
		},
	},
	{
		Name:        "set 'max_wal_senders' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_wal_senders",
				Expected: []sql.Row{{int64(10)}},
			},
			{
				Query:       "SET max_wal_senders TO '10'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_wal_senders')",
				Expected: []sql.Row{{"10"}},
			},
		},
	},
	{
		Name:        "set 'max_wal_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_wal_size",
				Expected: []sql.Row{{int64(1024)}},
			},
			{
				Query:       "SET max_wal_size TO '1000'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_wal_size')",
				Expected: []sql.Row{{"1024"}},
			},
		},
	},
	{
		Name:        "set 'max_worker_processes' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW max_worker_processes",
				Expected: []sql.Row{{int64(8)}},
			},
			{
				Query:       "SET max_worker_processes TO '8'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('max_worker_processes')",
				Expected: []sql.Row{{"8"}},
			},
		},
	},
	{
		Name:        "set 'min_dynamic_shared_memory' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW min_dynamic_shared_memory",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:       "SET min_dynamic_shared_memory TO '0'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('min_dynamic_shared_memory')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'min_parallel_index_scan_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW min_parallel_index_scan_size",
				Expected: []sql.Row{{int64(1024)}},
			},
			{
				Query:    "SET min_parallel_index_scan_size TO '512'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW min_parallel_index_scan_size",
				Expected: []sql.Row{{int64(512)}},
			},
			{
				Query:    "SET min_parallel_index_scan_size TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW min_parallel_index_scan_size",
				Expected: []sql.Row{{int64(1024)}},
			},
			{
				Query:    "SELECT current_setting('min_parallel_index_scan_size')",
				Expected: []sql.Row{{"1024"}},
			},
		},
	},
	{
		Name:        "set 'min_parallel_table_scan_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW min_parallel_table_scan_size",
				Expected: []sql.Row{{int64(1024)}},
			},
			{
				Query:    "SET min_parallel_table_scan_size TO '800'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW min_parallel_table_scan_size",
				Expected: []sql.Row{{int64(800)}},
			},
			{
				Query:    "SET min_parallel_table_scan_size TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW min_parallel_table_scan_size",
				Expected: []sql.Row{{int64(1024)}},
			},
			{
				Query:    "SELECT current_setting('min_parallel_table_scan_size')",
				Expected: []sql.Row{{"1024"}},
			},
		},
	},
	{
		Name:        "set 'min_wal_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW min_wal_size",
				Expected: []sql.Row{{int64(80)}},
			},
			{
				Query:       "SET min_wal_size TO '8000'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('min_wal_size')",
				Expected: []sql.Row{{"80"}},
			},
		},
	},
	{
		Name:        "set 'old_snapshot_threshold' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW old_snapshot_threshold",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:       "SET old_snapshot_threshold TO '-1'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('old_snapshot_threshold')",
				Expected: []sql.Row{{"-1"}},
			},
		},
	},
	{
		Name:        "set 'parallel_leader_participation' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW parallel_leader_participation",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET parallel_leader_participation TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW parallel_leader_participation",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET parallel_leader_participation TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW parallel_leader_participation",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('parallel_leader_participation')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'parallel_setup_cost' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW parallel_setup_cost",
				Expected: []sql.Row{{float64(1000)}},
			},
			{
				Query:    "SET parallel_setup_cost TO '10000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW parallel_setup_cost",
				Expected: []sql.Row{{float64(10000)}},
			},
			{
				Query:    "SET parallel_setup_cost TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW parallel_setup_cost",
				Expected: []sql.Row{{float64(1000)}},
			},
			{
				Query:    "SELECT current_setting('parallel_setup_cost')",
				Expected: []sql.Row{{"1000"}},
			},
		},
	},
	{
		Name:        "set 'parallel_tuple_cost' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW parallel_tuple_cost",
				Expected: []sql.Row{{float64(0.1)}},
			},
			{
				Query:    "SET parallel_tuple_cost TO '0.2'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW parallel_tuple_cost",
				Expected: []sql.Row{{float64(0.2)}},
			},
			{
				Query:    "SET parallel_tuple_cost TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW parallel_tuple_cost",
				Expected: []sql.Row{{float64(0.1)}},
			},
			{
				Query:    "SELECT current_setting('parallel_tuple_cost')",
				Expected: []sql.Row{{"0.1"}},
			},
		},
	},
	{
		Name:        "set 'password_encryption' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW password_encryption",
				Expected: []sql.Row{{"scram-sha-256"}},
			},
			{
				Query:    "SET password_encryption TO 'md5'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW password_encryption",
				Expected: []sql.Row{{"md5"}},
			},
			{
				Query:    "SET password_encryption TO 'scram-sha-256'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW password_encryption",
				Expected: []sql.Row{{"scram-sha-256"}},
			},
			{
				Query:    "SELECT current_setting('password_encryption')",
				Expected: []sql.Row{{"scram-sha-256"}},
			},
		},
	},
	{
		Name:        "set 'plan_cache_mode' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW plan_cache_mode",
				Expected: []sql.Row{{"auto"}},
			},
			{
				Query:    "SET plan_cache_mode TO 'force_generic_plan'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW plan_cache_mode",
				Expected: []sql.Row{{"force_generic_plan"}},
			},
			{
				Query:    "SET plan_cache_mode TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW plan_cache_mode",
				Expected: []sql.Row{{"auto"}},
			},
			{
				Query:    "SELECT current_setting('plan_cache_mode')",
				Expected: []sql.Row{{"auto"}},
			},
		},
	},
	{
		Name:        "set 'port' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW port",
				Expected: []sql.Row{{currentPort}},
			},
			{
				Query:       "SET port TO '5432'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('port')",
				Expected: []sql.Row{{fmt.Sprintf("%v", currentPort)}},
			},
		},
	},
	{
		Name:        "set 'post_auth_delay' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW post_auth_delay",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:       "SET post_auth_delay TO '0'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('post_auth_delay')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'pre_auth_delay' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW pre_auth_delay",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:       "SET pre_auth_delay TO '0'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('pre_auth_delay')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'primary_conninfo' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW primary_conninfo",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET primary_conninfo TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('primary_conninfo')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'primary_slot_name' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW primary_slot_name",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET primary_slot_name TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('primary_slot_name')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'quote_all_identifiers' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW quote_all_identifiers",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET quote_all_identifiers TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW quote_all_identifiers",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET quote_all_identifiers TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW quote_all_identifiers",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('quote_all_identifiers')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'random_page_cost' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW random_page_cost",
				Expected: []sql.Row{{float64(4)}},
			},
			{
				Query:    "SET random_page_cost TO 2.5",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW random_page_cost",
				Expected: []sql.Row{{float64(2.5)}},
			},
			{
				Query:    "SET random_page_cost TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW random_page_cost",
				Expected: []sql.Row{{float64(4)}},
			},
			{
				Query:    "SELECT current_setting('random_page_cost')",
				Expected: []sql.Row{{"4"}},
			},
		},
	},
	{
		Name:        "set 'recovery_end_command' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW recovery_end_command",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET recovery_end_command TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('recovery_end_command')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'recovery_init_sync_method' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW recovery_init_sync_method",
				Expected: []sql.Row{{"fsync"}},
			},
			{
				Query:       "SET recovery_init_sync_method TO 'fsync'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('recovery_init_sync_method')",
				Expected: []sql.Row{{"fsync"}},
			},
		},
	},
	{
		Name:        "set 'recovery_min_apply_delay' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW recovery_min_apply_delay",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:       "SET recovery_min_apply_delay TO '0'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('recovery_min_apply_delay')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'recovery_prefetch' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW recovery_prefetch",
				Expected: []sql.Row{{"try"}},
			},
			{
				Query:       "SET recovery_prefetch TO 'try'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('recovery_prefetch')",
				Expected: []sql.Row{{"try"}},
			},
		},
	},
	{
		Name:        "set 'recovery_target' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW recovery_target",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET recovery_target TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('recovery_target')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'recovery_target_action' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW recovery_target_action",
				Expected: []sql.Row{{"pause"}},
			},
			{
				Query:       "SET recovery_target_action TO 'pause'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('recovery_target_action')",
				Expected: []sql.Row{{"pause"}},
			},
		},
	},
	{
		Name:        "set 'recovery_target_inclusive' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW recovery_target_inclusive",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:       "SET recovery_target_inclusive TO 'on'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('recovery_target_inclusive')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'recovery_target_lsn' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW recovery_target_lsn",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET recovery_target_lsn TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('recovery_target_lsn')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'recovery_target_name' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW recovery_target_name",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET recovery_target_name TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('recovery_target_name')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'recovery_target_time' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW recovery_target_time",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET recovery_target_time TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('recovery_target_time')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'recovery_target_timeline' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW recovery_target_timeline",
				Expected: []sql.Row{{"latest"}},
			},
			{
				Query:       "SET recovery_target_timeline TO 'latest'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('recovery_target_timeline')",
				Expected: []sql.Row{{"latest"}},
			},
		},
	},
	{
		Name:        "set 'recovery_target_xid' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW recovery_target_xid",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET recovery_target_xid TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('recovery_target_xid')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'recursive_worktable_factor' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW recursive_worktable_factor",
				Expected: []sql.Row{{float64(10)}},
			},
			{
				Query:    "SET recursive_worktable_factor TO '1'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW recursive_worktable_factor",
				Expected: []sql.Row{{float64(1)}},
			},
			{
				Query:    "SET recursive_worktable_factor TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW recursive_worktable_factor",
				Expected: []sql.Row{{float64(10)}},
			},
			{
				Query:    "SELECT current_setting('recursive_worktable_factor')",
				Expected: []sql.Row{{"10"}},
			},
		},
	},
	{
		Name:        "set 'remove_temp_files_after_crash' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW remove_temp_files_after_crash",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET remove_temp_files_after_crash TO 'on'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('remove_temp_files_after_crash')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'reserved_connections' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW reserved_connections",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:       "SET reserved_connections TO '0'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('reserved_connections')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'restart_after_crash' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW restart_after_crash",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:       "SET restart_after_crash TO 'on'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('restart_after_crash')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'restore_command' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW restore_command",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET restore_command TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('restore_command')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'row_security' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW row_security",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET row_security TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW row_security",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET row_security TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW row_security",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('row_security')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'scram_iterations' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW scram_iterations",
				Expected: []sql.Row{{int64(4096)}},
			},
			{
				Query:    "SET scram_iterations TO '4000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW scram_iterations",
				Expected: []sql.Row{{int64(4000)}},
			},
			{
				Query:    "SET scram_iterations TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW scram_iterations",
				Expected: []sql.Row{{int64(4096)}},
			},
			{
				Query:    "SELECT current_setting('scram_iterations')",
				Expected: []sql.Row{{"4096"}},
			},
		},
	},
	{
		Name:        "set 'search_path' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW search_path",
				Expected: []sql.Row{{"\"$user\", public,"}},
			},
			{
				Query:    "SET search_path TO 'postgres'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW search_path",
				Expected: []sql.Row{{"postgres"}},
			},
			{
				Query:    "SET search_path TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW search_path",
				Expected: []sql.Row{{"\"$user\", public,"}},
			},
			{
				Query:    "SELECT current_setting('search_path')",
				Expected: []sql.Row{{"\"$user\", public,"}},
			},
		},
	},
	{
		Name:        "set 'segment_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW segment_size",
				Expected: []sql.Row{{int64(131072)}},
			},
			{
				Query:       "SET segment_size TO '131072'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('segment_size')",
				Expected: []sql.Row{{"131072"}},
			},
		},
	},
	{
		Name:        "set 'send_abort_for_crash' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW send_abort_for_crash",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET send_abort_for_crash TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('send_abort_for_crash')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'send_abort_for_kill' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW send_abort_for_kill",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET send_abort_for_kill TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('send_abort_for_kill')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'seq_page_cost' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW seq_page_cost",
				Expected: []sql.Row{{float64(1)}},
			},
			{
				Query:       "SET seq_page_cost TO '1'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('seq_page_cost')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'server_encoding' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW server_encoding",
				Expected: []sql.Row{{"UTF8"}},
			},
			{
				Query:       "SET server_encoding TO 'UTF8'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('server_encoding')",
				Expected: []sql.Row{{"UTF8"}},
			},
		},
	},
	{
		Name:        "set 'server_version' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW server_version",
				Expected: []sql.Row{{"16.1 (Homebrew)"}},
			},
			{
				Query:       "SET server_version TO '16.1 (Homebrew)'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('server_version')",
				Expected: []sql.Row{{"16.1 (Homebrew)"}},
			},
		},
	},
	{
		Name:        "set 'server_version_num' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW server_version_num",
				Expected: []sql.Row{{int64(160001)}},
			},
			{
				Query:       "SET server_version_num TO '160001'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('server_version_num')",
				Expected: []sql.Row{{"160001"}},
			},
		},
	},
	{
		Name:        "set 'session_preload_libraries' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW session_preload_libraries",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SET session_preload_libraries TO '/'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW session_preload_libraries",
				Expected: []sql.Row{{"/"}},
			},
			{
				Query:    "SET session_preload_libraries TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW session_preload_libraries",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SELECT current_setting('session_preload_libraries')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'session_replication_role' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW session_replication_role",
				Expected: []sql.Row{{"origin"}},
			},
			{
				Query:    "SET session_replication_role TO 'local'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW session_replication_role",
				Expected: []sql.Row{{"local"}},
			},
			{
				Query:    "SET session_replication_role TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW session_replication_role",
				Expected: []sql.Row{{"origin"}},
			},
			{
				Query:    "SELECT current_setting('session_replication_role')",
				Expected: []sql.Row{{"origin"}},
			},
		},
	},
	{
		Name:        "set 'shared_buffers' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW shared_buffers",
				Expected: []sql.Row{{int64(16384)}},
			},
			{
				Query:       "SET shared_buffers TO '128000'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('shared_buffers')",
				Expected: []sql.Row{{"16384"}},
			},
		},
	},
	{
		Name:        "set 'shared_memory_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW shared_memory_size",
				Expected: []sql.Row{{int64(143)}},
			},
			{
				Query:       "SET shared_memory_size TO '143000'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('shared_memory_size')",
				Expected: []sql.Row{{"143"}},
			},
		},
	},
	{
		Name:        "set 'shared_memory_size_in_huge_pages' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW shared_memory_size_in_huge_pages",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:       "SET shared_memory_size_in_huge_pages TO '-1'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('shared_memory_size_in_huge_pages')",
				Expected: []sql.Row{{"-1"}},
			},
		},
	},
	{
		Name:        "set 'shared_memory_type' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW shared_memory_type",
				Expected: []sql.Row{{"mmap"}},
			},
			{
				Query:       "SET shared_memory_type TO 'mmap'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('shared_memory_type')",
				Expected: []sql.Row{{"mmap"}},
			},
		},
	},
	{
		Name:        "set 'shared_preload_libraries' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW shared_preload_libraries",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET shared_preload_libraries TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('shared_preload_libraries')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'ssl' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ssl",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET ssl TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ssl')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'ssl_ca_file' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ssl_ca_file",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET ssl_ca_file TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ssl_ca_file')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'ssl_cert_file' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ssl_cert_file",
				Expected: []sql.Row{{"server.crt"}},
			},
			{
				Query:       "SET ssl_cert_file TO 'server.crt'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ssl_cert_file')",
				Expected: []sql.Row{{"server.crt"}},
			},
		},
	},
	{
		Name:        "set 'ssl_ciphers' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ssl_ciphers",
				Expected: []sql.Row{{"HIGH:MEDIUM:+3DES:!aNULL"}},
			},
			{
				Query:       "SET ssl_ciphers TO 'HIGH:MEDIUM:'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ssl_ciphers')",
				Expected: []sql.Row{{"HIGH:MEDIUM:+3DES:!aNULL"}},
			},
		},
	},
	{
		Name:        "set 'ssl_crl_dir' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ssl_crl_dir",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET ssl_crl_dir TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ssl_crl_dir')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'ssl_crl_file' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ssl_crl_file",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET ssl_crl_file TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ssl_crl_file')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'ssl_dh_params_file' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ssl_dh_params_file",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET ssl_dh_params_file TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ssl_dh_params_file')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'ssl_ecdh_curve' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ssl_ecdh_curve",
				Expected: []sql.Row{{"prime256v1"}},
			},
			{
				Query:       "SET ssl_ecdh_curve TO 'prime256v1'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ssl_ecdh_curve')",
				Expected: []sql.Row{{"prime256v1"}},
			},
		},
	},
	{
		Name:        "set 'ssl_key_file' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ssl_key_file",
				Expected: []sql.Row{{"server.key"}},
			},
			{
				Query:       "SET ssl_key_file TO 'server.key'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ssl_key_file')",
				Expected: []sql.Row{{"server.key"}},
			},
		},
	},
	{
		Name:        "set 'ssl_library' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ssl_library",
				Expected: []sql.Row{{"OpenSSL"}},
			},
			{
				Query:       "SET ssl_library TO 'OpenSSL'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ssl_library')",
				Expected: []sql.Row{{"OpenSSL"}},
			},
		},
	},
	{
		Name:        "set 'ssl_max_protocol_version' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ssl_max_protocol_version",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET ssl_max_protocol_version TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ssl_max_protocol_version')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'ssl_min_protocol_version' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ssl_min_protocol_version",
				Expected: []sql.Row{{"TLSv1.2"}},
			},
			{
				Query:       "SET ssl_min_protocol_version TO 'TLSv1.2'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ssl_min_protocol_version')",
				Expected: []sql.Row{{"TLSv1.2"}},
			},
		},
	},
	{
		Name:        "set 'ssl_passphrase_command' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ssl_passphrase_command",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET ssl_passphrase_command TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ssl_passphrase_command')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'ssl_passphrase_command_supports_reload' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ssl_passphrase_command_supports_reload",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET ssl_passphrase_command_supports_reload TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ssl_passphrase_command_supports_reload')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'ssl_prefer_server_ciphers' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW ssl_prefer_server_ciphers",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:       "SET ssl_prefer_server_ciphers TO 'on'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('ssl_prefer_server_ciphers')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'standard_conforming_strings' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW standard_conforming_strings",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET standard_conforming_strings TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW standard_conforming_strings",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET standard_conforming_strings TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW standard_conforming_strings",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('standard_conforming_strings')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'statement_timeout' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW statement_timeout",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET statement_timeout TO '42'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW statement_timeout",
				Expected: []sql.Row{{int64(42)}},
			},
			{
				Query:    "SELECT current_setting('statement_timeout')",
				Expected: []sql.Row{{"42"}},
			},
		},
	},
	{
		Name:        "set 'stats_fetch_consistency' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW stats_fetch_consistency",
				Expected: []sql.Row{{"cache"}},
			},
			{
				Query:    "SET stats_fetch_consistency TO 'snapshot'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW stats_fetch_consistency",
				Expected: []sql.Row{{"snapshot"}},
			},
			{
				Query:    "SET stats_fetch_consistency TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW stats_fetch_consistency",
				Expected: []sql.Row{{"cache"}},
			},
			{
				Query:    "SELECT current_setting('stats_fetch_consistency')",
				Expected: []sql.Row{{"cache"}},
			},
		},
	},
	{
		Name:        "set 'superuser_reserved_connections' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW superuser_reserved_connections",
				Expected: []sql.Row{{int64(3)}},
			},
			{
				Query:       "SET superuser_reserved_connections TO '3'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('superuser_reserved_connections')",
				Expected: []sql.Row{{"3"}},
			},
		},
	},
	{
		Name:        "set 'synchronize_seqscans' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW synchronize_seqscans",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET synchronize_seqscans TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW synchronize_seqscans",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET synchronize_seqscans TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW synchronize_seqscans",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('synchronize_seqscans')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'synchronous_commit' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW synchronous_commit",
				Expected: []sql.Row{{"on"}},
			},
			{
				Query:    "SET synchronous_commit TO 'local'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW synchronous_commit",
				Expected: []sql.Row{{"local"}},
			},
			{
				Query:    "SET synchronous_commit TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW synchronous_commit",
				Expected: []sql.Row{{"on"}},
			},
			{
				Query:    "SELECT current_setting('synchronous_commit')",
				Expected: []sql.Row{{"on"}},
			},
		},
	},
	{
		Name:        "set 'synchronous_standby_names' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW synchronous_standby_names",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET synchronous_standby_names TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('synchronous_standby_names')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'syslog_facility' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW syslog_facility",
				Expected: []sql.Row{{"local0"}},
			},
			{
				Query:       "SET syslog_facility TO 'local0'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('syslog_facility')",
				Expected: []sql.Row{{"local0"}},
			},
		},
	},
	{
		Name:        "set 'syslog_ident' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW syslog_ident",
				Expected: []sql.Row{{"postgres"}},
			},
			{
				Query:       "SET syslog_ident TO 'postgres'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('syslog_ident')",
				Expected: []sql.Row{{"postgres"}},
			},
		},
	},
	{
		Name:        "set 'syslog_sequence_numbers' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW syslog_sequence_numbers",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:       "SET syslog_sequence_numbers TO 'on'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('syslog_sequence_numbers')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'syslog_split_messages' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW syslog_split_messages",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:       "SET syslog_split_messages TO 'on'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('syslog_split_messages')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'tcp_keepalives_count' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW tcp_keepalives_count",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET tcp_keepalives_count TO 100",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW tcp_keepalives_count",
				Expected: []sql.Row{{int64(100)}},
			},
			{
				Query:    "SET tcp_keepalives_count TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW tcp_keepalives_count",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('tcp_keepalives_count')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'tcp_keepalives_idle' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW tcp_keepalives_idle",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET tcp_keepalives_idle TO 1",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW tcp_keepalives_idle",
				Expected: []sql.Row{{int64(1)}},
			},
			{
				Query:    "SET tcp_keepalives_idle TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW tcp_keepalives_idle",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('tcp_keepalives_idle')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'tcp_keepalives_interval' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW tcp_keepalives_interval",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET tcp_keepalives_interval TO 1",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW tcp_keepalives_interval",
				Expected: []sql.Row{{int64(1)}},
			},
			{
				Query:    "SET tcp_keepalives_interval TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW tcp_keepalives_interval",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('tcp_keepalives_interval')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'tcp_user_timeout' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW tcp_user_timeout",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SET tcp_user_timeout TO '100000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW tcp_user_timeout",
				Expected: []sql.Row{{int64(100000)}},
			},
			{
				Query:    "SET tcp_user_timeout TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW tcp_user_timeout",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:    "SELECT current_setting('tcp_user_timeout')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'temp_buffers' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW temp_buffers",
				Expected: []sql.Row{{int64(1024)}},
			},
			{
				Query:    "SET temp_buffers TO '8000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW temp_buffers",
				Expected: []sql.Row{{int64(8000)}},
			},
			{
				Query:    "SET temp_buffers TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW temp_buffers",
				Expected: []sql.Row{{int64(1024)}},
			},
			{
				Query:    "SELECT current_setting('temp_buffers')",
				Expected: []sql.Row{{"1024"}},
			},
		},
	},
	{
		Name:        "set 'temp_file_limit' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW temp_file_limit",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:    "SET temp_file_limit TO 100",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW temp_file_limit",
				Expected: []sql.Row{{int64(100)}},
			},
			{
				Query:    "SET temp_file_limit TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW temp_file_limit",
				Expected: []sql.Row{{int64(-1)}},
			},
			{
				Query:    "SELECT current_setting('temp_file_limit')",
				Expected: []sql.Row{{"-1"}},
			},
		},
	},
	{
		Name:        "set 'temp_tablespaces' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW temp_tablespaces",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SET temp_tablespaces TO 'pg_default'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW temp_tablespaces",
				Expected: []sql.Row{{"pg_default"}},
			},
			{
				Query:    "SET temp_tablespaces TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW temp_tablespaces",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SELECT current_setting('temp_tablespaces')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'TimeZone' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW TimeZone",
				Expected: []sql.Row{{"America/Los_Angeles"}},
			},
			{
				Query:    "SET TimeZone TO 'UTC'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW TimeZone",
				Expected: []sql.Row{{"UTC"}},
			},
			{
				Query:    "SET TimeZone TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW TimeZone",
				Expected: []sql.Row{{"America/Los_Angeles"}},
			},
			{
				Query:    "SELECT current_setting('TimeZone')",
				Expected: []sql.Row{{"America/Los_Angeles"}},
			},
		},
	},
	{
		Name:        "set 'timezone_abbreviations' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW timezone_abbreviations",
				Expected: []sql.Row{{"Default"}},
			},
			{
				Query:    "SET timezone_abbreviations TO ''",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW timezone_abbreviations",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SET timezone_abbreviations TO 'Default'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW timezone_abbreviations",
				Expected: []sql.Row{{"Default"}},
			},
			{
				Query:    "SELECT current_setting('timezone_abbreviations')",
				Expected: []sql.Row{{"Default"}},
			},
		},
	},
	{
		Name:        "set 'trace_notify' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW trace_notify",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET trace_notify TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW trace_notify",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET trace_notify TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW trace_notify",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('trace_notify')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'trace_recovery_messages' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW trace_recovery_messages",
				Expected: []sql.Row{{"log"}},
			},
			{
				Query:       "SET trace_recovery_messages TO 'log'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('trace_recovery_messages')",
				Expected: []sql.Row{{"log"}},
			},
		},
	},
	{
		Name:        "set 'trace_sort' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW trace_sort",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET trace_sort TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW trace_sort",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET trace_sort TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW trace_sort",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('trace_sort')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'track_activities' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW track_activities",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET track_activities TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW track_activities",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET track_activities TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW track_activities",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('track_activities')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'track_activity_query_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW track_activity_query_size",
				Expected: []sql.Row{{int64(1024)}},
			},
			{
				Query:       "SET track_activity_query_size TO '1024'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('track_activity_query_size')",
				Expected: []sql.Row{{"1024"}},
			},
		},
	},
	{
		Name:        "set 'track_commit_timestamp' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW track_commit_timestamp",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET track_commit_timestamp TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('track_commit_timestamp')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'track_counts' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW track_counts",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET track_counts TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW track_counts",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET track_counts TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW track_counts",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('track_counts')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'track_functions' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW track_functions",
				Expected: []sql.Row{{"none"}},
			},
			{
				Query:    "SET track_functions TO 'all'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW track_functions",
				Expected: []sql.Row{{"all"}},
			},
			{
				Query:    "SET track_functions TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW track_functions",
				Expected: []sql.Row{{"none"}},
			},
			{
				Query:    "SELECT current_setting('track_functions')",
				Expected: []sql.Row{{"none"}},
			},
		},
	},
	{
		Name:        "set 'track_io_timing' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW track_io_timing",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET track_io_timing TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW track_io_timing",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET track_io_timing TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW track_io_timing",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('track_io_timing')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'track_wal_io_timing' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW track_wal_io_timing",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET track_wal_io_timing TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW track_wal_io_timing",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET track_wal_io_timing TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW track_wal_io_timing",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('track_wal_io_timing')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'transaction_deferrable' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW transaction_deferrable",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET transaction_deferrable TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW transaction_deferrable",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET transaction_deferrable TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW transaction_deferrable",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('transaction_deferrable')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'transaction_isolation' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW transaction_isolation",
				Expected: []sql.Row{{"read committed"}},
			},
			{
				Query:    "SET transaction_isolation TO 'serializable'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW transaction_isolation",
				Expected: []sql.Row{{"serializable"}},
			},
			{
				Query:    "SET transaction_isolation TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW transaction_isolation",
				Expected: []sql.Row{{"read committed"}},
			},
			{
				Query:    "SELECT current_setting('transaction_isolation')",
				Expected: []sql.Row{{"read committed"}},
			},
		},
	},
	{
		Name:        "set 'transaction_read_only' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW transaction_read_only",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET transaction_read_only TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW transaction_read_only",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET transaction_read_only TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW transaction_read_only",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('transaction_read_only')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'transform_null_equals' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW transform_null_equals",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET transform_null_equals TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW transform_null_equals",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET transform_null_equals TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW transform_null_equals",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('transform_null_equals')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'unix_socket_directories' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW unix_socket_directories",
				Expected: []sql.Row{{"/tmp"}},
			},
			{
				Query:       "SET unix_socket_directories TO '/tmp'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('unix_socket_directories')",
				Expected: []sql.Row{{"/tmp"}},
			},
		},
	},
	{
		Name:        "set 'unix_socket_group' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW unix_socket_group",
				Expected: []sql.Row{{""}},
			},
			{
				Query:       "SET unix_socket_group TO ''",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('unix_socket_group')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'unix_socket_permissions' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW unix_socket_permissions",
				Expected: []sql.Row{{int64(511)}},
			},
			{
				Query:       "SET unix_socket_permissions TO '511'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('unix_socket_permissions')",
				Expected: []sql.Row{{"511"}},
			},
		},
	},
	{
		Name:        "set 'update_process_title' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW update_process_title",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET update_process_title TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW update_process_title",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET update_process_title TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW update_process_title",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('update_process_title')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'vacuum_buffer_usage_limit' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW vacuum_buffer_usage_limit",
				Expected: []sql.Row{{int64(256)}},
			},
			{
				Query:    "SET vacuum_buffer_usage_limit TO '512'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_buffer_usage_limit",
				Expected: []sql.Row{{int64(512)}},
			},
			{
				Query:    "SET vacuum_buffer_usage_limit TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_buffer_usage_limit",
				Expected: []sql.Row{{int64(256)}},
			},
			{
				Query:    "SELECT current_setting('vacuum_buffer_usage_limit')",
				Expected: []sql.Row{{"256"}},
			},
		},
	},
	{
		Name:        "set 'vacuum_cost_delay' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW vacuum_cost_delay",
				Expected: []sql.Row{{float64(0)}},
			},
			{
				Query:    "SET vacuum_cost_delay TO '0.2'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_cost_delay",
				Expected: []sql.Row{{float64(0.2)}},
			},
			{
				Query:    "SET vacuum_cost_delay TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_cost_delay",
				Expected: []sql.Row{{float64(0)}},
			},
			{
				Query:    "SELECT current_setting('vacuum_cost_delay')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'vacuum_cost_limit' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW vacuum_cost_limit",
				Expected: []sql.Row{{int64(200)}},
			},
			{
				Query:    "SET vacuum_cost_limit TO '400'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_cost_limit",
				Expected: []sql.Row{{int64(400)}},
			},
			{
				Query:    "SET vacuum_cost_limit TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_cost_limit",
				Expected: []sql.Row{{int64(200)}},
			},
			{
				Query:    "SELECT current_setting('vacuum_cost_limit')",
				Expected: []sql.Row{{"200"}},
			},
		},
	},
	{
		Name:        "set 'vacuum_cost_page_dirty' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW vacuum_cost_page_dirty",
				Expected: []sql.Row{{int64(20)}},
			},
			{
				Query:    "SET vacuum_cost_page_dirty TO '200'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_cost_page_dirty",
				Expected: []sql.Row{{int64(200)}},
			},
			{
				Query:    "SET vacuum_cost_page_dirty TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_cost_page_dirty",
				Expected: []sql.Row{{int64(20)}},
			},
			{
				Query:    "SELECT current_setting('vacuum_cost_page_dirty')",
				Expected: []sql.Row{{"20"}},
			},
		},
	},
	{
		Name:        "set 'vacuum_cost_page_hit' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW vacuum_cost_page_hit",
				Expected: []sql.Row{{int64(1)}},
			},
			{
				Query:    "SET vacuum_cost_page_hit TO '100'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_cost_page_hit",
				Expected: []sql.Row{{int64(100)}},
			},
			{
				Query:    "SET vacuum_cost_page_hit TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_cost_page_hit",
				Expected: []sql.Row{{int64(1)}},
			},
			{
				Query:    "SELECT current_setting('vacuum_cost_page_hit')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'vacuum_cost_page_miss' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW vacuum_cost_page_miss",
				Expected: []sql.Row{{int64(2)}},
			},
			{
				Query:    "SET vacuum_cost_page_miss TO '20'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_cost_page_miss",
				Expected: []sql.Row{{int64(20)}},
			},
			{
				Query:    "SET vacuum_cost_page_miss TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_cost_page_miss",
				Expected: []sql.Row{{int64(2)}},
			},
			{
				Query:    "SELECT current_setting('vacuum_cost_page_miss')",
				Expected: []sql.Row{{"2"}},
			},
		},
	},
	{
		Name:        "set 'vacuum_failsafe_age' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW vacuum_failsafe_age",
				Expected: []sql.Row{{int64(1600000000)}},
			},
			{
				Query:    "SET vacuum_failsafe_age TO '2100000000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_failsafe_age",
				Expected: []sql.Row{{int64(2100000000)}},
			},
			{
				Query:    "SET vacuum_failsafe_age TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_failsafe_age",
				Expected: []sql.Row{{int64(1600000000)}},
			},
			{
				Query:    "SELECT current_setting('vacuum_failsafe_age')",
				Expected: []sql.Row{{"1600000000"}},
			},
		},
	},
	{
		Name:        "set 'vacuum_freeze_min_age' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW vacuum_freeze_min_age",
				Expected: []sql.Row{{int64(50000000)}},
			},
			{
				Query:    "SET vacuum_freeze_min_age TO '20000000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_freeze_min_age",
				Expected: []sql.Row{{int64(20000000)}},
			},
			{
				Query:    "SET vacuum_freeze_min_age TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_freeze_min_age",
				Expected: []sql.Row{{int64(50000000)}},
			},
			{
				Query:    "SELECT current_setting('vacuum_freeze_min_age')",
				Expected: []sql.Row{{"50000000"}},
			},
		},
	},
	{
		Name:        "set 'vacuum_freeze_table_age' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW vacuum_freeze_table_age",
				Expected: []sql.Row{{int64(150000000)}},
			},
			{
				Query:    "SET vacuum_freeze_table_age TO '100000000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_freeze_table_age",
				Expected: []sql.Row{{int64(100000000)}},
			},
			{
				Query:    "SET vacuum_freeze_table_age TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_freeze_table_age",
				Expected: []sql.Row{{int64(150000000)}},
			},
			{
				Query:    "SELECT current_setting('vacuum_freeze_table_age')",
				Expected: []sql.Row{{"150000000"}},
			},
		},
	},
	{
		Name:        "set 'vacuum_multixact_failsafe_age' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW vacuum_multixact_failsafe_age",
				Expected: []sql.Row{{int64(1600000000)}},
			},
			{
				Query:    "SET vacuum_multixact_failsafe_age TO '1000000000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_multixact_failsafe_age",
				Expected: []sql.Row{{int64(1000000000)}},
			},
			{
				Query:    "SET vacuum_multixact_failsafe_age TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_multixact_failsafe_age",
				Expected: []sql.Row{{int64(1600000000)}},
			},
			{
				Query:    "SELECT current_setting('vacuum_multixact_failsafe_age')",
				Expected: []sql.Row{{"1600000000"}},
			},
		},
	},
	{
		Name:        "set 'vacuum_multixact_freeze_min_age' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW vacuum_multixact_freeze_min_age",
				Expected: []sql.Row{{int64(5000000)}},
			},
			{
				Query:    "SET vacuum_multixact_freeze_min_age TO '2000000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_multixact_freeze_min_age",
				Expected: []sql.Row{{int64(2000000)}},
			},
			{
				Query:    "SET vacuum_multixact_freeze_min_age TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_multixact_freeze_min_age",
				Expected: []sql.Row{{int64(5000000)}},
			},
			{
				Query:    "SELECT current_setting('vacuum_multixact_freeze_min_age')",
				Expected: []sql.Row{{"5000000"}},
			},
		},
	},
	{
		Name:        "set 'vacuum_multixact_freeze_table_age' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW vacuum_multixact_freeze_table_age",
				Expected: []sql.Row{{int64(150000000)}},
			},
			{
				Query:    "SET vacuum_multixact_freeze_table_age TO '120000000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_multixact_freeze_table_age",
				Expected: []sql.Row{{int64(120000000)}},
			},
			{
				Query:    "SET vacuum_multixact_freeze_table_age TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW vacuum_multixact_freeze_table_age",
				Expected: []sql.Row{{int64(150000000)}},
			},
			{
				Query:    "SELECT current_setting('vacuum_multixact_freeze_table_age')",
				Expected: []sql.Row{{"150000000"}},
			},
		},
	},
	{
		Name:        "set 'wal_block_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_block_size",
				Expected: []sql.Row{{int64(8192)}},
			},
			{
				Query:       "SET wal_block_size TO '8192'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('wal_block_size')",
				Expected: []sql.Row{{"8192"}},
			},
		},
	},
	{
		Name:        "set 'wal_buffers' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_buffers",
				Expected: []sql.Row{{int64(512)}},
			},
			{
				Query:       "SET wal_buffers TO '4000'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('wal_buffers')",
				Expected: []sql.Row{{"512"}},
			},
		},
	},
	{
		Name:        "set 'wal_compression' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_compression",
				Expected: []sql.Row{{"off"}},
			},
			{
				Query:    "SET wal_compression TO 'lz4'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW wal_compression",
				Expected: []sql.Row{{"lz4"}},
			},
			{
				Query:    "SET wal_compression TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW wal_compression",
				Expected: []sql.Row{{"off"}},
			},
			{
				Query:    "SELECT current_setting('wal_compression')",
				Expected: []sql.Row{{"off"}},
			},
		},
	},
	{
		Name:        "set 'wal_consistency_checking' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_consistency_checking",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SET wal_consistency_checking TO 'generic'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW wal_consistency_checking",
				Expected: []sql.Row{{"generic"}},
			},
			{
				Query:    "SET wal_consistency_checking TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW wal_consistency_checking",
				Expected: []sql.Row{{""}},
			},
			{
				Query:    "SELECT current_setting('wal_consistency_checking')",
				Expected: []sql.Row{{""}},
			},
		},
	},
	{
		Name:        "set 'wal_decode_buffer_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_decode_buffer_size",
				Expected: []sql.Row{{int64(524288)}},
			},
			{
				Query:       "SET wal_decode_buffer_size TO '524288'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('wal_decode_buffer_size')",
				Expected: []sql.Row{{"524288"}},
			},
		},
	},
	{
		Name:        "set 'wal_init_zero' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_init_zero",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET wal_init_zero TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW wal_init_zero",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET wal_init_zero TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW wal_init_zero",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('wal_init_zero')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'wal_keep_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_keep_size",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				Query:       "SET wal_keep_size TO '0'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('wal_keep_size')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'wal_level' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_level",
				Expected: []sql.Row{{"replica"}},
			},
			{
				Query:       "SET wal_level TO 'replica'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('wal_level')",
				Expected: []sql.Row{{"replica"}},
			},
		},
	},
	{
		Name:        "set 'wal_log_hints' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_log_hints",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET wal_log_hints TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('wal_log_hints')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'wal_receiver_create_temp_slot' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_receiver_create_temp_slot",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:       "SET wal_receiver_create_temp_slot TO 'off'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('wal_receiver_create_temp_slot')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name:        "set 'wal_receiver_status_interval' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_receiver_status_interval",
				Expected: []sql.Row{{int64(10)}},
			},
			{
				Query:       "SET wal_receiver_status_interval TO '10'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('wal_receiver_status_interval')",
				Expected: []sql.Row{{"10"}},
			},
		},
	},
	{
		Name:        "set 'wal_receiver_timeout' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_receiver_timeout",
				Expected: []sql.Row{{int64(60000)}},
			},
			{
				Query:       "SET wal_receiver_timeout TO '60'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('wal_receiver_timeout')",
				Expected: []sql.Row{{"60000"}},
			},
		},
	},
	{
		Name:        "set 'wal_recycle' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_recycle",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET wal_recycle TO 'off'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW wal_recycle",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET wal_recycle TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW wal_recycle",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SELECT current_setting('wal_recycle')",
				Expected: []sql.Row{{"1"}},
			},
		},
	},
	{
		Name:        "set 'wal_retrieve_retry_interval' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_retrieve_retry_interval",
				Expected: []sql.Row{{int64(5000)}},
			},
			{
				Query:       "SET wal_retrieve_retry_interval TO '5'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('wal_retrieve_retry_interval')",
				Expected: []sql.Row{{"5000"}},
			},
		},
	},
	{
		Name:        "set 'wal_segment_size' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_segment_size",
				Expected: []sql.Row{{int64(16777216)}},
			},
			{
				Query:       "SET wal_segment_size TO '16777216'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('wal_segment_size')",
				Expected: []sql.Row{{"16777216"}},
			},
		},
	},
	{
		Name:        "set 'wal_sender_timeout' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_sender_timeout",
				Expected: []sql.Row{{int64(60000)}},
			},
			{
				Query:    "SET wal_sender_timeout TO '100000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW wal_sender_timeout",
				Expected: []sql.Row{{int64(100000)}},
			},
			{
				Query:    "SET wal_sender_timeout TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW wal_sender_timeout",
				Expected: []sql.Row{{int64(60000)}},
			},
			{
				Query:    "SELECT current_setting('wal_sender_timeout')",
				Expected: []sql.Row{{"60000"}},
			},
		},
	},
	{
		Name:        "set 'wal_skip_threshold' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_skip_threshold",
				Expected: []sql.Row{{int64(2048)}},
			},
			{
				Query:    "SET wal_skip_threshold TO '2000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW wal_skip_threshold",
				Expected: []sql.Row{{int64(2000)}},
			},
			{
				Query:    "SET wal_skip_threshold TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW wal_skip_threshold",
				Expected: []sql.Row{{int64(2048)}},
			},
			{
				Query:    "SELECT current_setting('wal_skip_threshold')",
				Expected: []sql.Row{{"2048"}},
			},
		},
	},
	{
		Name:        "set 'wal_sync_method' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_sync_method",
				Expected: []sql.Row{{"open_datasync"}},
			},
			{
				Query:       "SET wal_sync_method TO 'open_datasync'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('wal_sync_method')",
				Expected: []sql.Row{{"open_datasync"}},
			},
		},
	},
	{
		Name:        "set 'wal_writer_delay' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_writer_delay",
				Expected: []sql.Row{{int64(200)}},
			},
			{
				Query:       "SET wal_writer_delay TO '200'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('wal_writer_delay')",
				Expected: []sql.Row{{"200"}},
			},
		},
	},
	{
		Name:        "set 'wal_writer_flush_after' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW wal_writer_flush_after",
				Expected: []sql.Row{{int64(128)}},
			},
			{
				Query:       "SET wal_writer_flush_after TO '1000'",
				ExpectedErr: "is a read only variable",
			},
			{
				Query:    "SELECT current_setting('wal_writer_flush_after')",
				Expected: []sql.Row{{"128"}},
			},
		},
	},
	{
		Name:        "set 'work_mem' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW work_mem",
				Expected: []sql.Row{{int64(4096)}},
			},
			{
				Query:    "SET work_mem TO '4000'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW work_mem",
				Expected: []sql.Row{{int64(4000)}},
			},
			{
				Query:    "SET work_mem TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW work_mem",
				Expected: []sql.Row{{int64(4096)}},
			},
			{
				Query:    "SELECT current_setting('work_mem')",
				Expected: []sql.Row{{"4096"}},
			},
		},
	},
	{
		Name:        "set 'xmlbinary' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW xmlbinary",
				Expected: []sql.Row{{"base64"}},
			},
			{
				Query:    "SET xmlbinary TO 'hex'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW xmlbinary",
				Expected: []sql.Row{{"hex"}},
			},
			{
				Query:    "SET xmlbinary TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW xmlbinary",
				Expected: []sql.Row{{"base64"}},
			},
			{
				Query:    "SELECT current_setting('xmlbinary')",
				Expected: []sql.Row{{"base64"}},
			},
		},
	},
	{
		Name:        "set 'xmloption' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW xmloption",
				Expected: []sql.Row{{"content"}},
			},
			{
				Query:    "SET xmloption TO 'document'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW xmloption",
				Expected: []sql.Row{{"document"}},
			},
			{
				Query:    "SET xmloption TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW xmloption",
				Expected: []sql.Row{{"content"}},
			},
			{
				Query:    "SELECT current_setting('xmloption')",
				Expected: []sql.Row{{"content"}},
			},
		},
	},
	{
		Name:        "set 'zero_damaged_pages' configuration variable",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SHOW zero_damaged_pages",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SET zero_damaged_pages TO 'on'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW zero_damaged_pages",
				Expected: []sql.Row{{int8(1)}},
			},
			{
				Query:    "SET zero_damaged_pages TO DEFAULT",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW zero_damaged_pages",
				Expected: []sql.Row{{int8(0)}},
			},
			{
				Query:    "SELECT current_setting('zero_damaged_pages')",
				Expected: []sql.Row{{"0"}},
			},
		},
	},
	{
		Name: "settings with namespaces",
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SET myvar.var_value TO 'value'",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW myvar.var_value",
				Expected: []sql.Row{{"value"}},
			},
			{
				Query:    "select current_setting('myvar.var_value')",
				Expected: []sql.Row{{"value"}},
			},
			{
				Query:       "select current_setting('unknown_var')",
				ExpectedErr: "unrecognized configuration parameter",
			},
			{
				Query:       "show myvar.unknown_var",
				ExpectedErr: "unrecognized configuration parameter",
			},
			{
				Query:    "set myvar.var_value to (select 'a')",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW myvar.var_value",
				Expected: []sql.Row{{"a"}},
			},
			{
				Query:    "set myvar.val2 to (select current_setting('myvar.var_value'))",
				Expected: []sql.Row{},
			},
			{
				Query:    "SHOW myvar.val2",
				Expected: []sql.Row{{"a"}},
			},
		},
	},
}
