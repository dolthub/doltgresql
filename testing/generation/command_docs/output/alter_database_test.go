// Copyright 2023 Dolthub, Inc.
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

package output

import "testing"

func TestAlterDatabase(t *testing.T) {
	tests := []QueryParses{
		Parses("ALTER DATABASE name"),
		Parses("ALTER DATABASE name ALLOW_CONNECTIONS allowconn"),
		Parses("ALTER DATABASE name WITH ALLOW_CONNECTIONS allowconn"),
		Parses("ALTER DATABASE name CONNECTION LIMIT -1"),
		Parses("ALTER DATABASE name WITH CONNECTION LIMIT -1"),
		Parses("ALTER DATABASE name IS_TEMPLATE istemplate"),
		Parses("ALTER DATABASE name WITH IS_TEMPLATE istemplate"),
		Parses("ALTER DATABASE name ALLOW_CONNECTIONS allowconn ALLOW_CONNECTIONS allowconn"),
		Parses("ALTER DATABASE name WITH ALLOW_CONNECTIONS allowconn ALLOW_CONNECTIONS allowconn"),
		Parses("ALTER DATABASE name CONNECTION LIMIT -1 ALLOW_CONNECTIONS allowconn"),
		Parses("ALTER DATABASE name WITH CONNECTION LIMIT -1 ALLOW_CONNECTIONS allowconn"),
		Parses("ALTER DATABASE name IS_TEMPLATE istemplate ALLOW_CONNECTIONS allowconn"),
		Parses("ALTER DATABASE name WITH IS_TEMPLATE istemplate ALLOW_CONNECTIONS allowconn"),
		Parses("ALTER DATABASE name ALLOW_CONNECTIONS allowconn CONNECTION LIMIT -1"),
		Parses("ALTER DATABASE name WITH ALLOW_CONNECTIONS allowconn CONNECTION LIMIT -1"),
		Parses("ALTER DATABASE name CONNECTION LIMIT -1 CONNECTION LIMIT -1"),
		Parses("ALTER DATABASE name WITH CONNECTION LIMIT -1 CONNECTION LIMIT -1"),
		Parses("ALTER DATABASE name IS_TEMPLATE istemplate CONNECTION LIMIT -1"),
		Parses("ALTER DATABASE name WITH IS_TEMPLATE istemplate CONNECTION LIMIT -1"),
		Parses("ALTER DATABASE name ALLOW_CONNECTIONS allowconn IS_TEMPLATE istemplate"),
		Parses("ALTER DATABASE name WITH ALLOW_CONNECTIONS allowconn IS_TEMPLATE istemplate"),
		Parses("ALTER DATABASE name CONNECTION LIMIT -1 IS_TEMPLATE istemplate"),
		Parses("ALTER DATABASE name WITH CONNECTION LIMIT -1 IS_TEMPLATE istemplate"),
		Parses("ALTER DATABASE name IS_TEMPLATE istemplate IS_TEMPLATE istemplate"),
		Parses("ALTER DATABASE name WITH IS_TEMPLATE istemplate IS_TEMPLATE istemplate"),
		Parses("ALTER DATABASE name RENAME TO new_name"),
		Converts("ALTER DATABASE name OWNER TO new_owner"),
		Converts("ALTER DATABASE name OWNER TO CURRENT_ROLE"),
		Converts("ALTER DATABASE name OWNER TO CURRENT_USER"),
		Converts("ALTER DATABASE name OWNER TO SESSION_USER"),
		Parses("ALTER DATABASE name SET TABLESPACE new_tablespace"),
		Parses("ALTER DATABASE name REFRESH COLLATION VERSION"),
		Parses("ALTER DATABASE name SET configuration_parameter TO value"),
		Parses("ALTER DATABASE name SET configuration_parameter = value"),
		Parses("ALTER DATABASE name SET configuration_parameter TO DEFAULT"),
		Parses("ALTER DATABASE name SET configuration_parameter = DEFAULT"),
		Parses("ALTER DATABASE name SET configuration_parameter FROM CURRENT"),
		Parses("ALTER DATABASE name RESET configuration_parameter"),
		Parses("ALTER DATABASE name RESET ALL"),
	}
	
	RunTests(t, tests)
	// RewriteTests(t, tests, "alter_database_test.go")
}
