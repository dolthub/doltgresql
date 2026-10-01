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

package parser

import (
	"strings"
	"testing"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

func TestSessionUserRoundTrip(t *testing.T) {
	stmt, err := ParseOne("SELECT SESSION_USER")
	if err != nil {
		t.Fatal(err)
	}
	formatted := tree.AsString(stmt.AST)
	if !strings.Contains(formatted, "session_user") {
		t.Fatalf("SESSION_USER formatted as %q", formatted)
	}
	if strings.Contains(formatted, "current_user") {
		t.Fatalf("SESSION_USER became current_user: %q", formatted)
	}
	if _, err := ParseOne(formatted); err != nil {
		t.Fatalf("formatted expression cannot be parsed: %v", err)
	}
}
