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
