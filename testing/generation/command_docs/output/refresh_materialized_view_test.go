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

func TestRefreshMaterializedView(t *testing.T) {
	tests := []QueryParses{
		Parses("REFRESH MATERIALIZED VIEW name"),
		Parses("REFRESH MATERIALIZED VIEW CONCURRENTLY name"),
		Parses("REFRESH MATERIALIZED VIEW name WITH DATA"),
		Parses("REFRESH MATERIALIZED VIEW CONCURRENTLY name WITH DATA"),
		Parses("REFRESH MATERIALIZED VIEW name WITH NO DATA"),
		Parses("REFRESH MATERIALIZED VIEW CONCURRENTLY name WITH NO DATA"),
	}
	RunTests(t, tests)
}
