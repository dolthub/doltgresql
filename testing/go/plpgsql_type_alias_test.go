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

package _go

import (
	"fmt"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

// typeAliasCase declares a single PL/pgSQL variable of the type named by TypeName and checks what the variable
// actually holds, so that a type name resolving to the wrong type is caught rather than just a type name that
// fails to resolve at all.
type typeAliasCase struct {
	// FuncName names the generated function, and so also names the assertion.
	FuncName string
	// TypeName is written verbatim into the DECLARE, e.g. "pg_catalog.boolean".
	TypeName string
	// Literal is the DECLARE's default value, written verbatim. When empty the variable is declared without one.
	Literal string
	// ValueExpr is the expression the function returns, and defaults to "b::text".
	ValueExpr string
	// ExpectedType is the expected pg_typeof of the declared variable. When empty, pg_typeof is not checked,
	// which is necessary for the types that pg_typeof does not report yet.
	ExpectedType string
	// ExpectedValue is the expected result of ValueExpr.
	ExpectedValue string
}

func (c typeAliasCase) statement() string {
	declaration := fmt.Sprintf("b %s", c.TypeName)
	if c.Literal != "" {
		declaration = fmt.Sprintf("%s := %s", declaration, c.Literal)
	}
	value := c.ValueExpr
	if value == "" {
		value = "b::text"
	}
	if c.ExpectedType != "" {
		value = fmt.Sprintf("pg_typeof(b)::text || '|' || %s", value)
	}
	return fmt.Sprintf(`CREATE FUNCTION %s() RETURNS text AS $$ DECLARE %s; BEGIN RETURN %s; END; $$ LANGUAGE plpgsql;`,
		c.FuncName, declaration, value)
}

func (c typeAliasCase) expected() string {
	if c.ExpectedType != "" {
		return c.ExpectedType + "|" + c.ExpectedValue
	}
	return c.ExpectedValue
}

// typeAliasScript builds a ScriptTest that creates one function per case and asserts on what each one returns.
func typeAliasScript(name string, cases []typeAliasCase) ScriptTest {
	script := ScriptTest{Name: name}
	for _, c := range cases {
		script.SetUpScript = append(script.SetUpScript, c.statement())
		script.Assertions = append(script.Assertions, ScriptTestAssertion{
			Query:    fmt.Sprintf("SELECT %s();", c.FuncName),
			Expected: []sql.Row{{c.expected()}},
		})
	}
	return script
}

// TestPlpgsqlDeclareTypeAliases tests that a PL/pgSQL variable may be declared with any of the SQL standard type
// names that Postgres accepts as aliases for its own type names, qualified with the pg_catalog schema.
//
// The grammar resolves these names when they are written bare, but a schema-qualified name is handed to the
// interpreter as text (pg_query_go reports a declared type as the text that was written), so it can only be
// resolved by looking the name up in the parser's type name table.
func TestPlpgsqlDeclareTypeAliases(t *testing.T) {
	RunScripts(t, []ScriptTest{
		typeAliasScript("schema-qualified numeric type aliases", []typeAliasCase{
			{FuncName: "alias_boolean", TypeName: "pg_catalog.boolean", Literal: "true",
				ExpectedType: "boolean", ExpectedValue: "true"},
			{FuncName: "alias_int", TypeName: "pg_catalog.int", Literal: "11",
				ExpectedType: "integer", ExpectedValue: "11"},
			{FuncName: "alias_integer", TypeName: "pg_catalog.integer", Literal: "12",
				ExpectedType: "integer", ExpectedValue: "12"},
			{FuncName: "alias_bigint", TypeName: "pg_catalog.bigint", Literal: "9223372036854775807",
				ExpectedType: "bigint", ExpectedValue: "9223372036854775807"},
			{FuncName: "alias_smallint", TypeName: "pg_catalog.smallint", Literal: "32767",
				ExpectedType: "smallint", ExpectedValue: "32767"},
			{FuncName: "alias_decimal", TypeName: "pg_catalog.decimal", Literal: "'1.25'",
				ExpectedType: "numeric", ExpectedValue: "1.25"},
			{FuncName: "alias_dec", TypeName: "pg_catalog.dec", Literal: "'2.50'",
				ExpectedType: "numeric", ExpectedValue: "2.50"},
			{FuncName: "alias_real", TypeName: "pg_catalog.real", Literal: "'1.5'",
				ExpectedType: "real", ExpectedValue: "1.5"},
			{FuncName: "alias_double_precision", TypeName: "pg_catalog.double precision", Literal: "'2.5'",
				ExpectedType: "double precision", ExpectedValue: "2.5"},
			// FLOAT with no precision is FLOAT8, the same as DOUBLE PRECISION.
			{FuncName: "alias_float", TypeName: "pg_catalog.float", Literal: "'3.5'",
				ExpectedType: "double precision", ExpectedValue: "3.5"},
		}),
		typeAliasScript("schema-qualified character type aliases", []typeAliasCase{
			{FuncName: "alias_character_varying", TypeName: "pg_catalog.character varying", Literal: "'abc'",
				ExpectedType: "character varying", ExpectedValue: "abc"},
			{FuncName: "alias_char_varying", TypeName: "pg_catalog.char varying", Literal: "'def'",
				ExpectedType: "character varying", ExpectedValue: "def"},
			{FuncName: "alias_national_character_varying", TypeName: "pg_catalog.national character varying", Literal: "'ghi'",
				ExpectedType: "character varying", ExpectedValue: "ghi"},
			{FuncName: "alias_national_char_varying", TypeName: "pg_catalog.national char varying", Literal: "'jkl'",
				ExpectedType: "character varying", ExpectedValue: "jkl"},
			{FuncName: "alias_nchar_varying", TypeName: "pg_catalog.nchar varying", Literal: "'mno'",
				ExpectedType: "character varying", ExpectedValue: "mno"},
			{FuncName: "alias_character", TypeName: "pg_catalog.character", Literal: "'pqr'",
				ExpectedType: "character", ExpectedValue: "pqr"},
			{FuncName: "alias_national_character", TypeName: "pg_catalog.national character", Literal: "'stu'",
				ExpectedType: "character", ExpectedValue: "stu"},
			{FuncName: "alias_nchar", TypeName: "pg_catalog.nchar", Literal: "'vwx'",
				ExpectedType: "character", ExpectedValue: "vwx"},
		}),
		typeAliasScript("schema-qualified date/time type aliases", []typeAliasCase{
			{FuncName: "alias_timestamp_without_tz", TypeName: "pg_catalog.timestamp without time zone",
				Literal: "'2024-01-02 03:04:05'", ExpectedType: "timestamp without time zone",
				ExpectedValue: "2024-01-02 03:04:05"},
			// The value is rendered in UTC rather than in the session's time zone so that the expected result does
			// not depend on where the test runs.
			{FuncName: "alias_timestamp_with_tz", TypeName: "pg_catalog.timestamp with time zone",
				Literal: "'2024-01-02 03:04:05+00'", ValueExpr: "(b AT TIME ZONE 'UTC')::text",
				ExpectedType: "timestamp with time zone", ExpectedValue: "2024-01-02 03:04:05"},
			// pg_typeof does not report the time types yet, so these only check the round-tripped value.
			{FuncName: "alias_time_without_tz", TypeName: "pg_catalog.time without time zone",
				Literal: "'03:04:05'", ExpectedValue: "03:04:05"},
			{FuncName: "alias_time_with_tz", TypeName: "pg_catalog.time with time zone",
				Literal: "'03:04:05+00'", ExpectedValue: "03:04:05+00"},
		}),
		typeAliasScript("schema-qualified bit string type alias", []typeAliasCase{
			// A PL/pgSQL variable of a bit string type cannot hold a value yet (assigning one fails the same way
			// for the canonical `varbit` spelling), so this only checks that the name resolves to a usable type.
			{FuncName: "alias_bit_varying", TypeName: "pg_catalog.bit varying",
				ValueExpr: "(b IS NULL)::text", ExpectedValue: "true"},
		}),
		// pg_query_go reports a declared type as the text that was written, without normalizing the whitespace
		// within a multi-word name, so the lookup has to normalize it.
		typeAliasScript("multi-word type aliases with irregular whitespace", []typeAliasCase{
			{FuncName: "alias_double_precision_spaces", TypeName: "pg_catalog.double   precision", Literal: "'4.5'",
				ExpectedType: "double precision", ExpectedValue: "4.5"},
			{FuncName: "alias_timestamp_tz_newline", TypeName: "pg_catalog.timestamp\nwith\ttime  zone",
				Literal: "'2024-01-02 03:04:05+00'", ValueExpr: "(b AT TIME ZONE 'UTC')::text",
				ExpectedType: "timestamp with time zone", ExpectedValue: "2024-01-02 03:04:05"},
		}),
		// Not aliases, but they resolve through the same lookup, and each of them is a name whose
		// registered spelling differs from the one being declared.
		typeAliasScript("names whose registered spelling differs", []typeAliasCase{
			{FuncName: "canonical_bytea", TypeName: "pg_catalog.bytea", Literal: `'\x0102'`,
				ExpectedValue: `\x0102`},
			{FuncName: "unqualified_bytea", TypeName: "bytea", Literal: `'\x0304'`,
				ExpectedValue: `\x0304`},
			{FuncName: "canonical_bpchar", TypeName: "pg_catalog.bpchar", Literal: "'yz'",
				ExpectedType: "character", ExpectedValue: "yz"},
			{FuncName: "unqualified_character", TypeName: "character(3)", Literal: "'abc'",
				ExpectedType: "character", ExpectedValue: "abc"},
			// Unlike the bare CHARACTER keyword, which names bpchar, `pg_catalog.char` names the internal
			// one-byte "char" type. A PL/pgSQL variable of that type cannot hold a value yet (assigning one fails
			// the same way for the canonical `"char"` spelling), so this only checks that the name resolves.
			{FuncName: "canonical_qchar", TypeName: "pg_catalog.char",
				ValueExpr: "(b IS NULL)::text", ExpectedValue: "true"},
			{FuncName: "canonical_integer_array", TypeName: "pg_catalog.integer[]", Literal: "'{1,2,3}'",
				ExpectedType: "integer[]", ExpectedValue: "{1,2,3}"},
		}),
	})
}

// TestPlpgsqlDeclareUnknownType tests that a schema-qualified type name that names no type is still rejected,
// rather than being resolved by a too-eager alias lookup.
func TestPlpgsqlDeclareUnknownType(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "unknown schema-qualified type names are still rejected",
			SetUpScript: []string{
				`CREATE FUNCTION unknown_type() RETURNS text AS $$ DECLARE b pg_catalog.not_a_type; BEGIN RETURN 'x'; END; $$ LANGUAGE plpgsql;`,
				`CREATE FUNCTION unknown_multiword_type() RETURNS text AS $$ DECLARE b pg_catalog.double imprecision; BEGIN RETURN 'x'; END; $$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT unknown_type();",
					ExpectedErr: `type "pg_catalog.not_a_type" does not exist`,
				},
				{
					Query:       "SELECT unknown_multiword_type();",
					ExpectedErr: `type "pg_catalog.double imprecision" does not exist`,
				},
			},
		},
	})
}
