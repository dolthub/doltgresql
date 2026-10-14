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

func TestInetRegression(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "inet complement operator",
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT ~ inet '192.168.1.6'`, Expected: []sql.Row{{"63.87.254.249"}}},
				{Query: `SELECT ~ inet '::1'`, Expected: []sql.Row{{"ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffe"}}},
			},
		},
		{
			Name: "IPv4 decimal octets with leading zeros",
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT host('192.168.001.010'::inet)`, Expected: []sql.Row{{"192.168.1.10"}}},
				{Query: `SELECT host('010.000.000.001/8'::inet)`, Expected: []sql.Row{{"10.0.0.1"}}},
			},
		},
		{
			Name: "explicit length limited casts",
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT '192.168.1.6'::inet::varchar(5)`, Expected: []sql.Row{{"192.1"}}},
				{Query: `SELECT '192.168.1.6'::inet::char(5)`, Expected: []sql.Row{{"192.1"}}},
			},
		},
		{
			Name:        "inet aggregates and joins",
			SetUpScript: []string{`CREATE TABLE ips (ip inet)`, `INSERT INTO ips VALUES ('1.2.3.4'), ('1.2.3.4/32'), ('1.2.3.4/24'), ('::1')`},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT host(ip), masklen(ip), count(*) FROM ips GROUP BY ip ORDER BY ip`, Expected: []sql.Row{{"1.2.3.4", 24, 1}, {"1.2.3.4", 32, 2}, {"::1", 128, 1}}},
				{Query: `SELECT count(*) FROM ips a JOIN ips b ON a.ip = b.ip`, Expected: []sql.Row{{6}}},
			},
		},
	})
}
