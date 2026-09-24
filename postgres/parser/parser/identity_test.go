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

func TestSetRoleRoundTrip(t *testing.T) {
	for _, query := range []string{
		"SET ROLE reader", "SET SESSION ROLE reader", "SET LOCAL ROLE reader",
		"SET ROLE NONE", "SET ROLE DEFAULT", "SET LOCAL ROLE DEFAULT", "RESET ROLE",
		"SET ROLE TO reader", "SET ROLE = reader", "SET ROLE TO DEFAULT",
	} {
		t.Run(query, func(t *testing.T) {
			stmt, err := ParseOne(query)
			if err != nil {
				t.Fatal(err)
			}
			formatted := tree.AsString(stmt.AST)
			if _, err := ParseOne(formatted); err != nil {
				t.Fatalf("%q formatted as %q, which cannot be parsed: %v", query, formatted, err)
			}
		})
	}
}

func TestSessionAuthorizationRoundTrip(t *testing.T) {
	for _, query := range []string{
		"SET SESSION AUTHORIZATION reader", "SET SESSION SESSION AUTHORIZATION reader",
		"SET LOCAL SESSION AUTHORIZATION reader", "SET SESSION AUTHORIZATION DEFAULT",
		"SET LOCAL SESSION AUTHORIZATION DEFAULT", "RESET SESSION AUTHORIZATION",
		`SET SESSION AUTHORIZATION "Mixed Case"`,
	} {
		t.Run(query, func(t *testing.T) {
			stmt, err := ParseOne(query)
			if err != nil {
				t.Fatal(err)
			}
			formatted := tree.AsString(stmt.AST)
			if _, err := ParseOne(formatted); err != nil {
				t.Fatalf("%q formatted as %q, which cannot be parsed: %v", query, formatted, err)
			}
		})
	}
}
