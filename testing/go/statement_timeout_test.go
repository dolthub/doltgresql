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

// TestStatementTimeoutUnits checks PostgreSQL 15 time-setting semantics.
func TestStatementTimeoutUnits(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "statement_timeout units",
		Assertions: []ScriptTestAssertion{
			{Query: `SELECT set_config('statement_timeout', '12h', false);`, Expected: []sql.Row{{"12h"}}},
			{Query: `SELECT set_config('statement_timeout', '1d', false);`, Expected: []sql.Row{{"1d"}}},
			{Query: `SELECT set_config('statement_timeout', '90min', false);`, Expected: []sql.Row{{"90min"}}},
			{Query: `SELECT set_config('statement_timeout', '1.5s', false);`, Expected: []sql.Row{{"1500ms"}}},
			{Query: `SELECT set_config('statement_timeout', '1500ms', false);`, Expected: []sql.Row{{"1500ms"}}},
			{Query: `SELECT set_config('statement_timeout', '1500000us', false);`, Expected: []sql.Row{{"1500ms"}}},
			{Query: `SELECT set_config('statement_timeout', '  +2 h  ', false);`, Expected: []sql.Row{{"2h"}}},
			{Query: `SELECT set_config('statement_timeout', '42', false);`, Expected: []sql.Row{{"42ms"}}},
			{Query: `SELECT set_config('statement_timeout', '0', false);`, Expected: []sql.Row{{"0"}}},
			{Query: `SELECT set_config('statement_timeout', '0.5', false);`, Expected: []sql.Row{{"0"}}},
			{Query: `SELECT set_config('statement_timeout', '1.5', false);`, Expected: []sql.Row{{"2ms"}}},
			{Query: `SELECT set_config('statement_timeout', '2.5', false);`, Expected: []sql.Row{{"2ms"}}},
			{Query: `SELECT set_config('statement_timeout', '500us', false);`, Expected: []sql.Row{{"0"}}},
			{Query: `SELECT set_config('statement_timeout', '1500us', false);`, Expected: []sql.Row{{"2ms"}}},
			{Query: `SELECT set_config('statement_timeout', '2500us', false);`, Expected: []sql.Row{{"2ms"}}},
			{Query: `SELECT set_config('statement_timeout', '1.00001h', false);`, Expected: []sql.Row{{"1h"}}},
			{Query: `SELECT set_config('statement_timeout', '1.0005s', false);`, Expected: []sql.Row{{"1s"}}},
			{Query: `SELECT set_config('statement_timeout', '2147483647ms', false);`, Expected: []sql.Row{{"2147483647ms"}}},
			{Query: `SELECT set_config('statement_timeout', '0x10', false);`, Expected: []sql.Row{{"16ms"}}},
			{Query: `SELECT set_config('statement_timeout', '010', false);`, Expected: []sql.Row{{"8ms"}}},
			{Query: `SELECT set_config('statement_timeout', '1e3ms', false);`, Expected: []sql.Row{{"1s"}}},
			{Query: `SET statement_timeout = '-1';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = '-1ms';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = '2147483648';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = '25d';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = '1e100h';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = '1e-999ms';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = '12H';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = '1m';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = '1ns';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = '1h30min';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = '12hours';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = '';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = 'ms';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = 'NaN';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = 'Infinity';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = '08';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = '1_000';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SET statement_timeout = '1 s junk';`, ExpectedErr: `statement_timeout`, ExpectedErrCode: "22023"},
			{Query: `SHOW statement_timeout;`, Expected: []sql.Row{{"1s"}}},
			{Query: `SET statement_timeout = '12h';`},
			{Query: `SELECT setting FROM pg_settings WHERE name = 'statement_timeout';`, Expected: []sql.Row{{"43200000"}}},
			{Query: `SHOW statement_timeout;`, Expected: []sql.Row{{"12h"}}},
			{Query: `SELECT current_setting('statement_timeout');`, Expected: []sql.Row{{"12h"}}},
			{Query: `BEGIN;`},
			{Query: `SET LOCAL statement_timeout = '90min';`},
			{Query: `SHOW statement_timeout;`, Expected: []sql.Row{{"90min"}}},
			{Query: `COMMIT;`},
			{Query: `SHOW statement_timeout;`, Expected: []sql.Row{{"12h"}}},
			{Query: `BEGIN;`},
			{Query: `SELECT set_config('statement_timeout', '1d', false);`, Expected: []sql.Row{{"1d"}}},
			{Query: `ROLLBACK;`},
			{Query: `SHOW statement_timeout;`, Expected: []sql.Row{{"12h"}}},
			{Query: `RESET statement_timeout;`},
			{Query: `SHOW statement_timeout;`, Expected: []sql.Row{{"0"}}},
			{Query: `SET statement_timeout = 1000;`},
			{Query: `SHOW statement_timeout;`, Expected: []sql.Row{{"1s"}}},
			{Query: `SET statement_timeout = DEFAULT;`},
			{Query: `SHOW statement_timeout;`, Expected: []sql.Row{{"0"}}},
		},
	}})
}
