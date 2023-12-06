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

func TestDropIndex(t *testing.T) {
	tests := []QueryParses{
		Converts("DROP INDEX name"),
		Parses("DROP INDEX CONCURRENTLY name"),
		Converts("DROP INDEX IF EXISTS name"),
		Parses("DROP INDEX CONCURRENTLY IF EXISTS name"),
		Parses("DROP INDEX name , name"),
		Parses("DROP INDEX CONCURRENTLY name , name"),
		Parses("DROP INDEX IF EXISTS name , name"),
		Parses("DROP INDEX CONCURRENTLY IF EXISTS name , name"),
		Parses("DROP INDEX name CASCADE"),
		Parses("DROP INDEX CONCURRENTLY name CASCADE"),
		Parses("DROP INDEX IF EXISTS name CASCADE"),
		Parses("DROP INDEX CONCURRENTLY IF EXISTS name CASCADE"),
		Parses("DROP INDEX name , name CASCADE"),
		Parses("DROP INDEX CONCURRENTLY name , name CASCADE"),
		Parses("DROP INDEX IF EXISTS name , name CASCADE"),
		Parses("DROP INDEX CONCURRENTLY IF EXISTS name , name CASCADE"),
		Parses("DROP INDEX name RESTRICT"),
		Parses("DROP INDEX CONCURRENTLY name RESTRICT"),
		Parses("DROP INDEX IF EXISTS name RESTRICT"),
		Parses("DROP INDEX CONCURRENTLY IF EXISTS name RESTRICT"),
		Parses("DROP INDEX name , name RESTRICT"),
		Parses("DROP INDEX CONCURRENTLY name , name RESTRICT"),
		Parses("DROP INDEX IF EXISTS name , name RESTRICT"),
		Parses("DROP INDEX CONCURRENTLY IF EXISTS name , name RESTRICT"),
	}
	RunTests(t, tests)
}
