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

func TestInitPrivs(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_init_privs)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_init_privs,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT count(*) > 0 FROM pg_init_privs;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `GRANT SELECT ON pg_proc TO CURRENT_USER;`,
			},
			{
				Statement: `GRANT SELECT (prosrc) ON pg_proc TO CURRENT_USER;`,
			},
			{
				Statement: `GRANT SELECT (rolname, rolsuper) ON pg_authid TO CURRENT_USER;`,
			},
		},
	})
}
