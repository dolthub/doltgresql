// Copyright 2024 Dolthub, Inc.
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

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dserver "github.com/dolthub/doltgresql/server"
	"github.com/dolthub/doltgresql/servercfg"
	framework "github.com/dolthub/doltgresql/testing/go"
	"github.com/dolthub/doltgresql/utils"
)

var (
	// Inf represent positive infinity as a float64
	Inf = math.Inf(0)
	// EquivalenceThresholdFloat64 represents the allowable delta for values to be considered equivalent.
	// For now, we do not expect all non-integer values to be exactly equivalent due to small computational differences
	// between Go and C/C++, so this threshold allows us to check that non-integer values are close enough. Over time,
	// we may reduce or remove this value as we become more accurate.
	EquivalenceThresholdFloat64 = 0.001
	// EquivalenceThresholdNumeric represents the allowable delta for values to be considered equivalent.
	// This is computed using the float64 variant so that they're equivalent.
	EquivalenceThresholdNumeric = decimal.RequireFromString(strconv.FormatFloat(EquivalenceThresholdFloat64, 'f', -1, 64))
)

// ScriptTest defines a consistent structure for testing queries.
type ScriptTest struct {
	// Name of the script.
	Name string
	// The database to create and use. If not provided, then it defaults to "postgres".
	Database string
	// The SQL statements to execute as setup, in order. Results are not checked, but statements must not error.
	SetUpScript []string
	// The set of assertions to make after setup, in order
	Assertions []ScriptTestAssertion
	// When using RunScripts, setting this on one (or more) tests causes RunScripts to ignore all tests that have this
	// set to false (which is the default value). This allows a developer to easily "focus" on a specific test without
	// having to comment out other tests, pull it into a different function, etc. In addition, CI ensures that this is
	// false before passing, meaning this prevents the commented-out situation where the developer forgets to uncomment
	// their code.
	Focus bool
	// Skip is used to completely skip a test including setup
	Skip bool
}

// ScriptTestAssertion are the assertions upon which the script executes its main "testing" logic.
type ScriptTestAssertion struct {
	Query       string
	Expected    []sql.Row
	ExpectedErr bool

	// Skip is used to completely skip a test, not execute its query at all, and record it as a skipped test
	// in the test suite results.
	Skip bool
}

// RunScript runs the given script.
func RunScript(t *testing.T, script ScriptTest) {
	scriptDatabase := script.Database
	if len(scriptDatabase) == 0 {
		scriptDatabase = "postgres"
	}

	ctx, conn, controller := CreateServer(t, scriptDatabase)
	defer func() {
		conn.Close(ctx)
		controller.Stop()
		err := controller.WaitForStop()
		require.NoError(t, err)
	}()

	t.Run(script.Name, func(t *testing.T) {
		runScript(t, script, conn, ctx)
	})
}

// runScript runs the script given on the postgres connection provided
func runScript(t *testing.T, script ScriptTest, conn *pgx.Conn, ctx context.Context) {
	if script.Skip {
		t.Skip("Skip has been set in the script")
	}

	// Run the setup
	for _, query := range script.SetUpScript {
		_, err := conn.Exec(ctx, query)
		require.NoError(t, err)
	}

	// Run the assertions
	for _, assertion := range script.Assertions {
		t.Run(assertion.Query, func(t *testing.T) {
			if assertion.Skip {
				t.Skip("Skip has been set in the assertion")
			}
			// If we're skipping the results check, then we call Execute, as it uses a simplified message model.
			if assertion.ExpectedErr {
				_, err := conn.Exec(ctx, assertion.Query)
				if assertion.ExpectedErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			} else {
				rows, err := conn.Query(ctx, assertion.Query)
				require.NoError(t, err)
				readRows, _, err := ReadRows(rows)
				require.NoError(t, err)
				if !CompareResults(t, assertion.Expected, readRows) {
					// We know it will fail, but this will print a prettier message
					assert.Equal(t, assertion.Expected, readRows)
				}
			}
		})
	}
}

// RunScripts runs the given collection of scripts.
func RunScripts(t *testing.T, scripts []ScriptTest) {
	// First, we'll run through the scripts to check for the Focus variable. If it's true, then append it to the new slice.
	focusScripts := make([]ScriptTest, 0, len(scripts))
	for _, script := range scripts {
		if script.Focus {
			// If this is running in GitHub Actions, then we'll panic, because someone forgot to disable it before committing
			if _, ok := os.LookupEnv("GITHUB_ACTION"); ok {
				panic(fmt.Sprintf("The script `%s` has Focus set to `true`. GitHub Actions requires that "+
					"all tests are run, which Focus circumvents, leading to this error. Please disable Focus on "+
					"all tests.", script.Name))
			}
			focusScripts = append(focusScripts, script)
		}
	}
	// If we have scripts with Focus set, then we replace the normal script slice with the new slice.
	if len(focusScripts) > 0 {
		scripts = focusScripts
	}

	for _, script := range scripts {
		RunScript(t, script)
	}
}

func ptr[T any](val T) *T {
	return &val
}

// CreateServer creates a server with the given database, returning a connection to the server. The server will close
// when the connection is closed (or loses its connection to the server). The accompanying WaitGroup may be used to wait
// until the server has closed.
func CreateServer(t *testing.T, database string) (context.Context, *pgx.Conn, *svcs.Controller) {
	require.NotEmpty(t, database)
	port := GetUnusedPort(t)
	controller, err := dserver.RunInMemory(&servercfg.DoltgresConfig{
		ListenerConfig: &servercfg.DoltgresListenerConfig{
			PortNumber: &port,
			HostStr:    ptr("127.0.0.1"),
		},
	})
	require.NoError(t, err)

	ctx := context.Background()
	err = func() error {
		// The connection attempt may be made before the server has grabbed the port, so we'll retry the first
		// connection a few times.
		var conn *pgx.Conn
		var err error
		for i := 0; i < 3; i++ {
			conn, err = pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/", port))
			if err == nil {
				break
			} else {
				time.Sleep(time.Second)
			}
		}
		if err != nil {
			return err
		}

		defer conn.Close(ctx)
		if database != "postgres" {
			_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s;", database))
			return err
		}
		return nil
	}()
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/%s", port, database))
	require.NoError(t, err)
	return ctx, conn, controller
}

// ReadRows reads all of the given rows into a slice, then closes the rows.
func ReadRows(rows pgx.Rows) (readRows []sql.Row, oids []uint32, err error) {
	for _, desc := range rows.FieldDescriptions() {
		oids = append(oids, desc.DataTypeOID)
	}
	defer func() {
		err = errors.Join(err, rows.Err())
	}()
	for rows.Next() {
		row, err := rows.Values()
		if err != nil {
			return nil, nil, err
		}
		readRows = append(readRows, row)
	}
	return readRows, oids, nil
}

// Numeric creates a numeric value from a string.
func Numeric(str string) pgtype.Numeric {
	numeric := pgtype.Numeric{}
	if err := numeric.Scan(str); err != nil {
		panic(err)
	}
	return numeric
}

// NumericToDecimal converts a pgtype.Numeric value to a decimal.Decimal value.
func NumericToDecimal(val pgtype.Numeric) decimal.Decimal {
	strVal, err := val.Value()
	if err != nil {
		panic(err)
	}
	return decimal.RequireFromString(strVal.(string))
}

// GetUnusedPort returns an unused port.
func GetUnusedPort(t *testing.T) int {
	return framework.GetUnusedPort(t)
}

// CompareResults compares two sets of results, taking the equivalence thresholds into account when making the
// comparisons.
func CompareResults(t *testing.T, a []sql.Row, b []sql.Row) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !CompareRows(t, a[i], b[i]) {
			return false
		}
	}
	return true
}

// CompareRows compares two rows, taking the equivalence thresholds into account when making the comparisons.
func CompareRows(t *testing.T, a sql.Row, b sql.Row) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		aVal := a[i]
		bVal := b[i]
		if aVal != bVal {
			if reflect.TypeOf(aVal) != reflect.TypeOf(bVal) {
				return false
			}
			switch aVal.(type) {
			case float32:
				if utils.Abs(aVal.(float32)-bVal.(float32)) > float32(EquivalenceThresholdFloat64) {
					return false
				}
			case float64:
				if utils.Abs(aVal.(float64)-bVal.(float64)) > EquivalenceThresholdFloat64 {
					return false
				}
			case pgtype.Numeric:
				delta := NumericToDecimal(aVal.(pgtype.Numeric)).Sub(NumericToDecimal(bVal.(pgtype.Numeric))).Abs()
				if delta.Cmp(EquivalenceThresholdNumeric) == 1 {
					return false
				}
			default:
				return false
			}
			t.Log("values are not equal but fall within equivalence threshold")
		}
	}
	return true
}
