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

func TestCreateExtension(t *testing.T) {
	tests := []QueryParses{
		Converts("CREATE EXTENSION extension_name"),
		Converts("CREATE EXTENSION IF NOT EXISTS extension_name"),
		Converts("CREATE EXTENSION extension_name WITH"),
		Converts("CREATE EXTENSION IF NOT EXISTS extension_name WITH"),
		Parses("CREATE EXTENSION extension_name SCHEMA schema_name"),
		Parses("CREATE EXTENSION IF NOT EXISTS extension_name SCHEMA schema_name"),
		Parses("CREATE EXTENSION extension_name WITH SCHEMA schema_name"),
		Parses("CREATE EXTENSION IF NOT EXISTS extension_name WITH SCHEMA schema_name"),
		Parses("CREATE EXTENSION extension_name VERSION version"),
		Parses("CREATE EXTENSION IF NOT EXISTS extension_name VERSION version"),
		Parses("CREATE EXTENSION extension_name WITH VERSION version"),
		Parses("CREATE EXTENSION IF NOT EXISTS extension_name WITH VERSION version"),
		Parses("CREATE EXTENSION extension_name SCHEMA schema_name VERSION version"),
		Parses("CREATE EXTENSION IF NOT EXISTS extension_name SCHEMA schema_name VERSION version"),
		Parses("CREATE EXTENSION extension_name WITH SCHEMA schema_name VERSION version"),
		Parses("CREATE EXTENSION IF NOT EXISTS extension_name WITH SCHEMA schema_name VERSION version"),
		Parses("CREATE EXTENSION extension_name CASCADE"),
		Parses("CREATE EXTENSION IF NOT EXISTS extension_name CASCADE"),
		Parses("CREATE EXTENSION extension_name WITH CASCADE"),
		Parses("CREATE EXTENSION IF NOT EXISTS extension_name WITH CASCADE"),
		Parses("CREATE EXTENSION extension_name SCHEMA schema_name CASCADE"),
		Parses("CREATE EXTENSION IF NOT EXISTS extension_name SCHEMA schema_name CASCADE"),
		Parses("CREATE EXTENSION extension_name WITH SCHEMA schema_name CASCADE"),
		Parses("CREATE EXTENSION IF NOT EXISTS extension_name WITH SCHEMA schema_name CASCADE"),
		Parses("CREATE EXTENSION extension_name VERSION version CASCADE"),
		Parses("CREATE EXTENSION IF NOT EXISTS extension_name VERSION version CASCADE"),
		Parses("CREATE EXTENSION extension_name WITH VERSION version CASCADE"),
		Parses("CREATE EXTENSION IF NOT EXISTS extension_name WITH VERSION version CASCADE"),
		Parses("CREATE EXTENSION extension_name SCHEMA schema_name VERSION version CASCADE"),
		Parses("CREATE EXTENSION IF NOT EXISTS extension_name SCHEMA schema_name VERSION version CASCADE"),
		Parses("CREATE EXTENSION extension_name WITH SCHEMA schema_name VERSION version CASCADE"),
		Parses("CREATE EXTENSION IF NOT EXISTS extension_name WITH SCHEMA schema_name VERSION version CASCADE"),
	}
	RunTests(t, tests)
}
