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
	"slices"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests cover #3551. Schema USAGE only permits looking up the sequence;
// nextval requires sequence USAGE or UPDATE, and setval requires sequence UPDATE.
func TestSequenceAuthorizationPrivileges(t *testing.T) {
	operations := []struct {
		name              string
		query             string
		allowedPrivileges []string
		result            int64
		nextValue         int64
	}{
		{"nextval", `SELECT nextval('seq_auth.counter');`, []string{"USAGE", "UPDATE"}, 1, 2},
		{"nextval_regclass", `SELECT nextval('seq_auth.counter'::regclass);`, []string{"USAGE", "UPDATE"}, 1, 2},
		{"setval_two_arguments", `SELECT setval('seq_auth.counter', 100);`, []string{"UPDATE"}, 100, 101},
		{"setval_called", `SELECT setval('seq_auth.counter', 100, true);`, []string{"UPDATE"}, 100, 101},
		{"setval_not_called", `SELECT setval('seq_auth.counter', 100, false);`, []string{"UPDATE"}, 100, 100},
	}
	var scripts []ScriptTest
	for _, privilege := range []string{"", "SELECT", "USAGE", "UPDATE"} {
		privilegeName := privilege
		if privilegeName == "" {
			privilegeName = "schema_usage_only"
		}
		for _, operation := range operations {
			assertion := ScriptTestAssertion{
				Query:    operation.query,
				Username: "seq_reader",
				Password: "reader",
			}
			nextValue := int64(1)
			if slices.Contains(operation.allowedPrivileges, privilege) {
				assertion.Expected = []sql.Row{{operation.result}}
				nextValue = operation.nextValue
			} else {
				assertion.ExpectedErr = "permission denied for sequence counter"
				assertion.ExpectedErrCode = "42501"
			}
			scripts = append(scripts, ScriptTest{
				Name:        privilegeName + "_" + operation.name,
				SetUpScript: sequenceAuthorizationSetup(privilege),
				Assertions: []ScriptTestAssertion{
					assertion,
					{
						// Rejected calls must leave the sequence unchanged. Read it as
						// superuser using nextval, since direct sequence reads are unsupported.
						Query:    `SELECT nextval('seq_auth.counter');`,
						Expected: []sql.Row{{nextValue}},
					},
				},
			})
		}
	}
	RunScripts(t, scripts)
}

// Sequence names supplied by expressions must be checked against the sequence
// actually used, rather than the expression's SQL text.
func TestSequenceAuthorizationArguments(t *testing.T) {
	operations := []struct {
		name      string
		format    string
		privilege string
		result    int64
		nextValue int64
	}{
		{"nextval", "SELECT nextval(%s)", "USAGE", 1, 2},
		{"setval", "SELECT setval(%s, 100, false)", "UPDATE", 100, 100},
	}
	arguments := []struct {
		name     string
		expr     string
		suffix   string
		bindVars []any
	}{
		{"column", "sequence_name", " FROM seq_auth.targets;", nil},
		{"bound_parameter", "$1::text", ";", []any{"seq_auth.counter"}},
	}
	var scripts []ScriptTest
	for _, operation := range operations {
		for _, argument := range arguments {
			for _, allowed := range []bool{false, true} {
				privilege := "SELECT"
				if allowed {
					privilege = operation.privilege
				}
				setup := sequenceAuthorizationSetup(privilege)
				setup = append(setup,
					`CREATE TABLE seq_auth.targets (sequence_name TEXT);`,
					`INSERT INTO seq_auth.targets VALUES ('seq_auth.counter');`,
					`GRANT SELECT ON seq_auth.targets TO seq_reader;`,
				)
				assertion := ScriptTestAssertion{
					Query:    fmt.Sprintf(operation.format, argument.expr) + argument.suffix,
					BindVars: argument.bindVars,
					Username: "seq_reader",
					Password: "reader",
				}
				nextValue := int64(1)
				if allowed {
					assertion.Expected = []sql.Row{{operation.result}}
					nextValue = operation.nextValue
				} else {
					assertion.ExpectedErr = "permission denied for sequence counter"
					assertion.ExpectedErrCode = "42501"
				}
				scripts = append(scripts, ScriptTest{
					Name:        operation.name + "_" + argument.name + "_" + privilege,
					SetUpScript: setup,
					Assertions: []ScriptTestAssertion{
						assertion,
						{Query: `SELECT nextval('seq_auth.counter');`, Expected: []sql.Row{{nextValue}}},
					},
				})
			}
		}
	}
	RunScripts(t, scripts)
}

func TestSequenceAuthorizationDefaults(t *testing.T) {
	var scripts []ScriptTest
	for _, privilege := range []string{"", "USAGE", "UPDATE"} {
		setup := sequenceAuthorizationSetup(privilege)
		setup = append(setup,
			`CREATE TABLE seq_auth.items (id BIGINT DEFAULT nextval('seq_auth.counter'));`,
			`GRANT INSERT ON seq_auth.items TO seq_reader;`,
		)
		assertion := ScriptTestAssertion{
			Query:    `INSERT INTO seq_auth.items DEFAULT VALUES;`,
			Username: "seq_reader",
			Password: "reader",
		}
		var rows []sql.Row
		nextValue := int64(1)
		name := privilege
		if privilege == "" {
			name = "schema_usage_only"
			assertion.ExpectedErr = "permission denied for sequence counter"
			assertion.ExpectedErrCode = "42501"
			rows = []sql.Row{}
		} else {
			assertion.Expected = []sql.Row{}
			rows = []sql.Row{{int64(1)}}
			nextValue = 2
		}
		scripts = append(scripts, ScriptTest{
			Name:        name,
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				assertion,
				{Query: `SELECT id FROM seq_auth.items;`, Expected: rows},
				{Query: `SELECT nextval('seq_auth.counter');`, Expected: []sql.Row{{nextValue}}},
			},
		})
	}
	RunScripts(t, scripts)
}

func TestSequenceAuthorizationGrantSources(t *testing.T) {
	var scripts []ScriptTest
	for _, source := range []string{"PUBLIC", "inherited_role", "all_sequences_in_schema"} {
		for _, privilege := range []string{"USAGE", "UPDATE"} {
			setup := sequenceAuthorizationSetup("")
			switch source {
			case "PUBLIC":
				setup = append(setup, fmt.Sprintf("GRANT %s ON SEQUENCE seq_auth.counter TO PUBLIC;", privilege))
			case "inherited_role":
				setup = append(setup,
					`CREATE ROLE seq_group;`,
					`GRANT seq_group TO seq_reader;`,
					fmt.Sprintf("GRANT %s ON SEQUENCE seq_auth.counter TO seq_group;", privilege),
				)
			case "all_sequences_in_schema":
				setup = append(setup, fmt.Sprintf("GRANT %s ON ALL SEQUENCES IN SCHEMA seq_auth TO seq_reader;", privilege))
			}
			setval := ScriptTestAssertion{
				Query:    `SELECT setval('seq_auth.counter', 100, false);`,
				Username: "seq_reader",
				Password: "reader",
			}
			nextValue := int64(2)
			if privilege == "UPDATE" {
				setval.Expected = []sql.Row{{int64(100)}}
				nextValue = 100
			} else {
				setval.ExpectedErr = "permission denied for sequence counter"
				setval.ExpectedErrCode = "42501"
			}
			scripts = append(scripts, ScriptTest{
				Name:        source + "_" + privilege,
				SetUpScript: setup,
				Assertions: []ScriptTestAssertion{
					{Query: `SELECT nextval('seq_auth.counter');`, Username: "seq_reader", Password: "reader", Expected: []sql.Row{{int64(1)}}},
					setval,
					{Query: `SELECT nextval('seq_auth.counter');`, Expected: []sql.Row{{nextValue}}},
				},
			})
		}
	}
	RunScripts(t, scripts)
}

func TestSequenceAuthorizationSearchPath(t *testing.T) {
	var scripts []ScriptTest
	for _, operation := range []struct {
		name      string
		query     string
		privilege string
		result    int64
		nextValue int64
	}{
		{"nextval", `SELECT nextval('counter');`, "USAGE", 1, 2},
		{"setval", `SELECT setval('counter', 100, false);`, "UPDATE", 100, 100},
	} {
		setup := append(sequenceAuthorizationSetup(operation.privilege),
			`CREATE SCHEMA seq_other;`,
			`CREATE SEQUENCE seq_other.counter;`,
			`GRANT USAGE ON SCHEMA seq_other TO seq_reader;`,
		)
		scripts = append(scripts, ScriptTest{
			Name:        operation.name,
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{Query: `SET search_path TO seq_other, seq_auth;`, Username: "seq_reader", Password: "reader", Expected: []sql.Row{}},
				{Query: operation.query, Username: "seq_reader", Password: "reader", ExpectedErr: "permission denied for sequence counter", ExpectedErrCode: "42501"},
				{Query: `SELECT nextval('seq_other.counter');`, Expected: []sql.Row{{int64(1)}}},
				{Query: `SET search_path TO seq_auth;`, Username: "seq_reader", Password: "reader", Expected: []sql.Row{}},
				{Query: operation.query, Username: "seq_reader", Password: "reader", Expected: []sql.Row{{operation.result}}},
				{Query: `SELECT nextval('seq_auth.counter');`, Expected: []sql.Row{{operation.nextValue}}},
			},
		})
	}
	RunScripts(t, scripts)
}

func TestSequenceAuthorizationAlter(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "alter_sequence_still_uses_update_privilege",
		SetUpScript: append(sequenceAuthorizationSetup("USAGE"),
			`CREATE TABLE seq_auth.items (id BIGINT);`,
		),
		Assertions: []ScriptTestAssertion{
			{Query: `ALTER SEQUENCE seq_auth.counter OWNED BY seq_auth.items.id;`, Username: "seq_reader", Password: "reader", ExpectedErr: "permission denied for sequence counter", ExpectedErrCode: "42501"},
			{Query: `GRANT UPDATE ON SEQUENCE seq_auth.counter TO seq_reader;`, Expected: []sql.Row{}},
			{Query: `ALTER SEQUENCE seq_auth.counter OWNED BY seq_auth.items.id;`, Username: "seq_reader", Password: "reader", Expected: []sql.Row{}},
			{Query: `SELECT nextval('seq_auth.counter');`, Expected: []sql.Row{{int64(1)}}},
		},
	}})
}

// SERIAL creation must not require privileges on a sequence that does not yet
// exist. Once created, its default must enforce the sequence's privileges.
func TestSequenceAuthorizationSerial(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "create_then_use_serial_sequence",
		SetUpScript: append(sequenceAuthorizationSetup(""),
			`GRANT CREATE ON SCHEMA seq_auth TO seq_reader;`,
		),
		Assertions: []ScriptTestAssertion{
			{Query: `CREATE TABLE seq_auth.items (id SERIAL);`, Username: "seq_reader", Password: "reader", Expected: []sql.Row{}},
			{Query: `GRANT INSERT ON seq_auth.items TO seq_reader;`, Expected: []sql.Row{}},
			{Query: `INSERT INTO seq_auth.items DEFAULT VALUES;`, Username: "seq_reader", Password: "reader", ExpectedErr: "permission denied for sequence items_id_seq", ExpectedErrCode: "42501"},
			{Query: `SELECT nextval('seq_auth.items_id_seq');`, Expected: []sql.Row{{int64(1)}}},
			{Query: `GRANT USAGE ON SEQUENCE seq_auth.items_id_seq TO seq_reader;`, Expected: []sql.Row{}},
			{Query: `INSERT INTO seq_auth.items DEFAULT VALUES;`, Username: "seq_reader", Password: "reader", Expected: []sql.Row{}},
			{Query: `SELECT id FROM seq_auth.items;`, Expected: []sql.Row{{int32(2)}}},
		},
	}})
}

func TestSequenceAuthorizationPreparedRoleChange(t *testing.T) {
	for _, query := range []string{`SELECT nextval('seq_auth.counter');`, `SELECT setval('seq_auth.counter', 100);`} {
		t.Run(query, func(t *testing.T) {
			ctx, connection, controller := CreateServer(t, "postgres")
			defer func() {
				connection.Close(ctx)
				controller.Stop()
				require.NoError(t, controller.WaitForStop())
			}()
			setup := append(sequenceAuthorizationSetup(""),
				`ALTER ROLE seq_reader NOINHERIT;`,
				`CREATE ROLE seq_actor;`,
				`GRANT seq_actor TO seq_reader;`,
				`GRANT USAGE ON SCHEMA seq_auth TO seq_actor;`,
				`GRANT UPDATE ON SEQUENCE seq_auth.counter TO seq_actor;`,
			)
			for _, statement := range setup {
				_, err := connection.Default.Exec(ctx, statement)
				require.NoError(t, err)
			}
			require.NoError(t, connection.Connect(ctx, "seq_reader", "reader"))
			reader := connection.Current
			_, err := reader.Prepare(ctx, "sequence_mutation", query)
			require.NoError(t, err)
			assertDenied := func() {
				t.Helper()
				var value int64
				err := reader.QueryRow(ctx, "sequence_mutation").Scan(&value)
				var pgErr *pgconn.PgError
				require.ErrorAs(t, err, &pgErr)
				assert.Equal(t, "42501", pgErr.Code)
				assert.Contains(t, pgErr.Message, "permission denied for sequence counter")
			}
			assertDenied()
			_, err = reader.Exec(ctx, "SET ROLE seq_actor")
			require.NoError(t, err)
			var value int64
			require.NoError(t, reader.QueryRow(ctx, "sequence_mutation").Scan(&value))
			nextValue := int64(2)
			if query == `SELECT setval('seq_auth.counter', 100);` {
				assert.Equal(t, int64(100), value)
				nextValue = 101
			} else {
				assert.Equal(t, int64(1), value)
			}
			// Advance before RESET so an unauthorized setval repeat is observable.
			require.NoError(t, connection.Default.QueryRow(ctx, `SELECT nextval('seq_auth.counter');`).Scan(&value))
			require.Equal(t, nextValue, value)
			_, err = reader.Exec(ctx, "RESET ROLE")
			require.NoError(t, err)
			assertDenied()
			require.NoError(t, connection.Default.QueryRow(ctx, `SELECT nextval('seq_auth.counter');`).Scan(&value))
			assert.Equal(t, nextValue+1, value, "denied execution must leave sequence state unchanged")
		})
	}
}

// Preparing or successfully executing a statement must not preserve permission
// to use a sequence after that permission has been revoked.
func TestSequenceAuthorizationPreparedRevoke(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		privilege string
		result    int64
		nextValue int64
	}{
		{"nextval", `SELECT nextval('seq_auth.counter');`, "USAGE", 1, 2},
		{"setval", `SELECT setval('seq_auth.counter', 100);`, "UPDATE", 100, 101},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, connection, controller := CreateServer(t, "postgres")
			defer func() {
				connection.Close(ctx)
				controller.Stop()
				require.NoError(t, controller.WaitForStop())
			}()
			for _, query := range sequenceAuthorizationSetup(test.privilege) {
				_, err := connection.Default.Exec(ctx, query)
				require.NoError(t, err, "setup query: %s", query)
			}
			require.NoError(t, connection.Connect(ctx, "seq_reader", "reader"))
			reader := connection.Current
			_, err := reader.Prepare(ctx, "sequence_mutation", test.query)
			require.NoError(t, err)
			var value int64
			require.NoError(t, reader.QueryRow(ctx, "sequence_mutation").Scan(&value))
			require.Equal(t, test.result, value)

			_, err = connection.Default.Exec(ctx, fmt.Sprintf("REVOKE %s ON SEQUENCE seq_auth.counter FROM seq_reader;", test.privilege))
			require.NoError(t, err)
			// Advance once as superuser so an unauthorized repeat of setval(100)
			// would visibly reset the sequence, rather than writing the same value.
			require.NoError(t, connection.Default.QueryRow(ctx, `SELECT nextval('seq_auth.counter');`).Scan(&value))
			require.Equal(t, test.nextValue, value)
			err = reader.QueryRow(ctx, "sequence_mutation").Scan(&value)
			var pgErr *pgconn.PgError
			if assert.ErrorAs(t, err, &pgErr, "revoked sequence privilege must take effect on a prepared statement") {
				assert.Equal(t, "42501", pgErr.Code)
				assert.Contains(t, pgErr.Message, "permission denied for sequence counter")
			}
			require.NoError(t, connection.Default.QueryRow(ctx, `SELECT nextval('seq_auth.counter');`).Scan(&value))
			assert.Equal(t, test.nextValue+1, value, "the rejected execution must not change sequence state")
		})
	}
}

func sequenceAuthorizationSetup(privilege string) []string {
	queries := []string{
		`CREATE ROLE seq_reader LOGIN PASSWORD 'reader';`,
		`CREATE SCHEMA seq_auth;`,
		`CREATE SEQUENCE seq_auth.counter;`,
		`REVOKE ALL ON SEQUENCE seq_auth.counter FROM PUBLIC;`,
		`GRANT USAGE ON SCHEMA seq_auth TO seq_reader;`,
	}
	if privilege != "" {
		queries = append(queries, fmt.Sprintf("GRANT %s ON SEQUENCE seq_auth.counter TO seq_reader;", privilege))
	}
	return queries
}
