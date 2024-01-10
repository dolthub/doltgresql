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
		Unimplemented("ALTER DATABASE name"),
		Unimplemented("ALTER DATABASE name ALLOW_CONNECTIONS allowconn"),
		Unimplemented("ALTER DATABASE name WITH ALLOW_CONNECTIONS allowconn"),
		Unimplemented("ALTER DATABASE name CONNECTION LIMIT -1"),
		Unimplemented("ALTER DATABASE name WITH CONNECTION LIMIT -1"),
		Unimplemented("ALTER DATABASE name IS_TEMPLATE istemplate"),
		Unimplemented("ALTER DATABASE name WITH IS_TEMPLATE istemplate"),
		Unimplemented("ALTER DATABASE name ALLOW_CONNECTIONS allowconn ALLOW_CONNECTIONS allowconn"),
		Unimplemented("ALTER DATABASE name WITH ALLOW_CONNECTIONS allowconn ALLOW_CONNECTIONS allowconn"),
		Unimplemented("ALTER DATABASE name CONNECTION LIMIT -1 ALLOW_CONNECTIONS allowconn"),
		Unimplemented("ALTER DATABASE name WITH CONNECTION LIMIT -1 ALLOW_CONNECTIONS allowconn"),
		Unimplemented("ALTER DATABASE name IS_TEMPLATE istemplate ALLOW_CONNECTIONS allowconn"),
		Unimplemented("ALTER DATABASE name WITH IS_TEMPLATE istemplate ALLOW_CONNECTIONS allowconn"),
		Unimplemented("ALTER DATABASE name ALLOW_CONNECTIONS allowconn CONNECTION LIMIT -1"),
		Unimplemented("ALTER DATABASE name WITH ALLOW_CONNECTIONS allowconn CONNECTION LIMIT -1"),
		Unimplemented("ALTER DATABASE name CONNECTION LIMIT -1 CONNECTION LIMIT -1"),
		Unimplemented("ALTER DATABASE name WITH CONNECTION LIMIT -1 CONNECTION LIMIT -1"),
		Unimplemented("ALTER DATABASE name IS_TEMPLATE istemplate CONNECTION LIMIT -1"),
		Unimplemented("ALTER DATABASE name WITH IS_TEMPLATE istemplate CONNECTION LIMIT -1"),
		Unimplemented("ALTER DATABASE name ALLOW_CONNECTIONS allowconn IS_TEMPLATE istemplate"),
		Unimplemented("ALTER DATABASE name WITH ALLOW_CONNECTIONS allowconn IS_TEMPLATE istemplate"),
		Unimplemented("ALTER DATABASE name CONNECTION LIMIT -1 IS_TEMPLATE istemplate"),
		Unimplemented("ALTER DATABASE name WITH CONNECTION LIMIT -1 IS_TEMPLATE istemplate"),
		Unimplemented("ALTER DATABASE name IS_TEMPLATE istemplate IS_TEMPLATE istemplate"),
		Unimplemented("ALTER DATABASE name WITH IS_TEMPLATE istemplate IS_TEMPLATE istemplate"),
		Parses("ALTER DATABASE name RENAME TO new_name"),
		Parses("ALTER DATABASE name OWNER TO new_owner"),
		Parses("ALTER DATABASE name OWNER TO CURRENT_ROLE"),
		Parses("ALTER DATABASE name OWNER TO CURRENT_USER"),
		Parses("ALTER DATABASE name OWNER TO SESSION_USER"),
		Unimplemented("ALTER DATABASE name SET TABLESPACE new_tablespace"),
		Unimplemented("ALTER DATABASE name REFRESH COLLATION VERSION"),
		Unimplemented("ALTER DATABASE name SET configuration_parameter TO value"),
		Unimplemented("ALTER DATABASE name SET configuration_parameter = value"),
		Unimplemented("ALTER DATABASE name SET configuration_parameter TO DEFAULT"),
		Unimplemented("ALTER DATABASE name SET configuration_parameter = DEFAULT"),
		Unimplemented("ALTER DATABASE name SET configuration_parameter FROM CURRENT"),
		Unimplemented("ALTER DATABASE name RESET configuration_parameter"),
		Unimplemented("ALTER DATABASE name RESET ALL"),
	}
	RunTests(t, tests)
}
