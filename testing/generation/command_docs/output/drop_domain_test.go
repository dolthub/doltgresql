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

func TestDropDomain(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("DROP DOMAIN name"),
		Unimplemented("DROP DOMAIN IF EXISTS name"),
		Unimplemented("DROP DOMAIN name , name"),
		Unimplemented("DROP DOMAIN IF EXISTS name , name"),
		Unimplemented("DROP DOMAIN name CASCADE"),
		Unimplemented("DROP DOMAIN IF EXISTS name CASCADE"),
		Unimplemented("DROP DOMAIN name , name CASCADE"),
		Unimplemented("DROP DOMAIN IF EXISTS name , name CASCADE"),
		Unimplemented("DROP DOMAIN name RESTRICT"),
		Unimplemented("DROP DOMAIN IF EXISTS name RESTRICT"),
		Unimplemented("DROP DOMAIN name , name RESTRICT"),
		Unimplemented("DROP DOMAIN IF EXISTS name , name RESTRICT"),
	}
	RunTests(t, tests)
}
