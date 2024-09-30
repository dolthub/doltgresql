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

package _go

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"testing"
	"time"

	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/postgres/parser/timeofday"
	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	dserver "github.com/dolthub/doltgresql/server"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/servercfg"
)

// runOnPostgres is a debug setting to redirect the test framework to a local running postgres server,
// rather than starting a doltgres server.
const runOnPostgres = false

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
	ExpectedErr string

	BindVars []any

	// SkipResultsCheck is used to skip assertions on the expected rows returned from a query. For now, this is
	// included as some messages do not have a full logical implementation. Skipping the results check allows us to
	// force the test client to not send of those messages.
	SkipResultsCheck bool

	// Skip is used to completely skip a test, not execute its query at all, and record it as a skipped test
	// in the test suite results.
	Skip bool

	// ExpectedTag is used to check the command tag returned from the server.
	// This is checked only if no Expected is defined
	ExpectedTag string

	// Cols is used to check the column names returned from the server.
	Cols []string
}

// RunScript runs the given script.
func RunScript(t *testing.T, script ScriptTest, normalizeRows bool) {
	scriptDatabase := script.Database
	if len(scriptDatabase) == 0 {
		scriptDatabase = "postgres"
	}

	var ctx context.Context
	var conn *pgx.Conn

	if runOnPostgres {
		var err error
		ctx = context.Background()
		conn, err = pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/%s?sslmode=disable", 5432, scriptDatabase))
		require.NoError(t, err)
		defer func() {
			_ = conn.Close(ctx)
		}()
	} else {
		var controller *svcs.Controller
		ctx, conn, controller = CreateServer(t, scriptDatabase)
		defer func() {
			_ = conn.Close(ctx)
			controller.Stop()
			err := controller.WaitForStop()
			require.NoError(t, err)
		}()
	}

	t.Run(script.Name, func(t *testing.T) {
		runScript(t, ctx, script, conn, normalizeRows)
	})
}

// runScript runs the script given on the postgres connection provided
func runScript(t *testing.T, ctx context.Context, script ScriptTest, conn *pgx.Conn, normalizeRows bool) {
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
			if assertion.SkipResultsCheck || assertion.ExpectedErr != "" {
				_, err := conn.Exec(ctx, assertion.Query, assertion.BindVars...)
				if assertion.ExpectedErr != "" {
					require.Error(t, err)
					assert.Contains(t, err.Error(), assertion.ExpectedErr)
				} else {
					require.NoError(t, err)
				}
			} else if assertion.ExpectedTag != "" {
				// check for command tag
				commandTag, err := conn.Exec(ctx, assertion.Query)
				require.NoError(t, err)
				assert.Equal(t, assertion.ExpectedTag, commandTag.String())
			} else {
				rows, err := conn.Query(ctx, assertion.Query, assertion.BindVars...)
				require.NoError(t, err)
				readRows, err := ReadRows(rows, normalizeRows)
				require.NoError(t, err)

				if assertion.Cols != nil {
					fields := rows.FieldDescriptions()
					if assert.Len(t, fields, len(assertion.Cols), "expected length of columns") {
						for i, col := range assertion.Cols {
							assert.Equal(t, col, fields[i].Name)
						}
					}
				}

				if normalizeRows {
					assert.Equal(t, NormalizeExpectedRow(rows.FieldDescriptions(), assertion.Expected), readRows)
				} else {
					assert.Equal(t, assertion.Expected, readRows)
				}
			}
		})
	}
}

// RunScripts runs the given collection of scripts. This normalizes all rows before comparing them.
func RunScripts(t *testing.T, scripts []ScriptTest) {
	runScripts(t, scripts, true)
}

// RunScriptsWithoutNormalization runs the given collection of scripts, without normalizing any rows.
func RunScriptsWithoutNormalization(t *testing.T, scripts []ScriptTest) {
	runScripts(t, scripts, false)
}

// runScripts is the implementation of both RunScripts and RunScriptsWithoutNormalization.
func runScripts(t *testing.T, scripts []ScriptTest, normalizeRows bool) {
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
		RunScript(t, script, normalizeRows)
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

	fmt.Printf("port is %d\n", port)

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
		_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s;", database))
		return err
	}()
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/%s", port, database))
	require.NoError(t, err)
	return ctx, conn, controller
}

// ReadRows reads all of the given rows into a slice, then closes the rows. If `normalizeRows` is true, then the rows
// will be normalized such that all integers are int64, etc.
func ReadRows(rows pgx.Rows, normalizeRows bool) (readRows []sql.Row, err error) {
	defer func() {
		err = errors.Join(err, rows.Err())
	}()
	var slice []sql.Row
	for rows.Next() {
		row, err := rows.Values()
		if err != nil {
			return nil, err
		}
		slice = append(slice, row)
	}
	return NormalizeRows(rows.FieldDescriptions(), slice, normalizeRows), nil
}

// NormalizeRows normalizes each value's type within each row, as the tests only want to compare values. Returns a new
// set of rows in the same order.
func NormalizeRows(fds []pgconn.FieldDescription, rows []sql.Row, normalize bool) []sql.Row {
	newRows := make([]sql.Row, len(rows))
	for i := range rows {
		newRows[i] = NormalizeRow(fds, rows[i], normalize)
	}
	return newRows
}

// NormalizeRow normalizes each value's type, as the tests only want to compare values.
// Returns a new row.
func NormalizeRow(fds []pgconn.FieldDescription, row sql.Row, normalize bool) sql.Row {
	if len(row) == 0 {
		return nil
	}
	newRow := make(sql.Row, len(row))
	for i := range row {
		dt, ok := dserver.OidToDoltgresType[fds[i].DataTypeOID]
		if !ok {
			panic(fmt.Sprintf("unhandled oid type: %v", fds[i].DataTypeOID))
		}
		newRow[i] = NormalizeValToString(dt, row[i])
		if normalize {
			newRow[i] = NormalizeIntsAndFloats(newRow[i])
		}
	}
	return newRow
}

// NormalizeExpectedRow normalizes each value's type, as the tests only want to compare values. Returns a new row.
func NormalizeExpectedRow(fds []pgconn.FieldDescription, rows []sql.Row) []sql.Row {
	newRows := make([]sql.Row, len(rows))
	for ri, row := range rows {
		if len(row) == 0 {
			newRows[ri] = nil
		} else if len(row) != len(fds) {
			// Return if the expected row count does not match the field description count, we'll error elsewhere
			return rows
		} else {
			newRow := make(sql.Row, len(row))
			for i := range row {
				dt, ok := dserver.OidToDoltgresType[fds[i].DataTypeOID]
				if !ok {
					panic(fmt.Sprintf("unhandled oid type: %v", fds[i].DataTypeOID))
				}
				if dt == types.Json {
					newRow[i] = UnmarshalAndMarshalJsonString(row[i].(string))
				} else if dta, ok := dt.(types.DoltgresArrayType); ok && dta.BaseType() == types.Json {
					v, err := dta.IoInput(nil, row[i].(string))
					if err != nil {
						panic(err)
					}
					arr := v.([]any)
					newArr := make([]any, len(arr))
					for j, el := range arr {
						newArr[j] = UnmarshalAndMarshalJsonString(el.(string))
					}
					ret, err := dt.IoOutput(nil, newArr)
					if err != nil {
						panic(err)
					}
					newRow[i] = ret
				} else {
					newRow[i] = NormalizeIntsAndFloats(row[i])
				}
			}
			newRows[ri] = newRow
		}
	}
	return newRows
}

// UnmarshalAndMarshalJsonString is used to normalize expected json type value to compare the actual value.
// JSON type value is in string format, and since Postrges JSON type preserves the input string if valid,
// it cannot be compared to the returned map as json.Marshal method space padded key value pair.
// To allow result matching, we unmarshal and marshal the expected string. This causes missing check
// for the identical format as the input of the json string.
func UnmarshalAndMarshalJsonString(val string) string {
	var decoded any
	err := json.Unmarshal([]byte(val), &decoded)
	if err != nil {
		panic(err)
	}
	ret, err := json.Marshal(decoded)
	if err != nil {
		panic(err)
	}
	return string(ret)
}

// NormalizeValToString normalizes values into types that can be compared.
// JSON types, any pg types and time and decimal type values are converted into string value.
// |normalizeNumeric| defines whether to normalize Numeric values into either Numeric type or string type.
// There are an infinite number of ways to represent the same value in-memory,
// so we must at least normalize Numeric values.
func NormalizeValToString(dt types.DoltgresType, v any) any {
	switch t := dt.(type) {
	case types.JsonType:
		str, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		ret, err := t.IoOutput(nil, string(str))
		if err != nil {
			panic(err)
		}
		return ret
	case types.JsonBType:
		jv, err := t.ConvertToJsonDocument(v)
		if err != nil {
			panic(err)
		}
		str, err := t.IoOutput(nil, types.JsonDocument{Value: jv})
		if err != nil {
			panic(err)
		}
		return str
	case types.InternalCharType:
		if v == nil {
			return nil
		}
		var b []byte
		if v.(int32) == 0 {
			b = []byte{}
		} else {
			b = []byte{uint8(v.(int32))}
		}
		val, err := t.IoOutput(nil, string(b))
		if err != nil {
			panic(err)
		}
		return val
	case types.IntervalType, types.UuidType, types.DateType, types.TimeType, types.TimestampType:
		// These values need to be normalized into the appropriate types
		// before being converted to string type using the Doltgres
		// IoOutput method.
		if v == nil {
			return nil
		}
		tVal, err := dt.IoOutput(nil, NormalizeVal(dt, v))
		if err != nil {
			panic(err)
		}
		return tVal
	case types.TimestampTZType:
		// timestamptz returns a value in server timezone
		_, offset := v.(time.Time).Zone()
		if offset%3600 != 0 {
			return v.(time.Time).Format("2006-01-02 15:04:05.999999999-07:00")
		} else {
			return v.(time.Time).Format("2006-01-02 15:04:05.999999999-07")
		}
	}

	switch val := v.(type) {
	case bool:
		if val {
			return "t"
		} else {
			return "f"
		}
	case pgtype.Numeric:
		if val.NaN {
			return math.NaN()
		} else if val.InfinityModifier != pgtype.Finite {
			return math.Inf(int(val.InfinityModifier))
		} else if !val.Valid {
			return nil
		} else {
			decStr := decimal.NewFromBigInt(val.Int, val.Exp).StringFixed(val.Exp * -1)
			return Numeric(decStr)
		}
	case []any:
		if dta, ok := dt.(types.DoltgresArrayType); ok {
			return NormalizeArrayType(dta, val)
		}
	}
	return v
}

// NormalizeArrayType normalizes array types by normalizing its elements first,
// then to a string using the type IoOutput method.
func NormalizeArrayType(dta types.DoltgresArrayType, arr []any) any {
	newVal := make([]any, len(arr))
	for i, el := range arr {
		newVal[i] = NormalizeVal(dta.BaseType(), el)
	}
	baseType := dta.BaseType()
	if baseType == types.Bool {
		sqlVal, err := dta.SQL(nil, nil, newVal)
		if err != nil {
			panic(err)
		}
		return sqlVal.ToString()
	} else {
		ret, err := dta.IoOutput(nil, newVal)
		if err != nil {
			panic(err)
		}
		return ret
	}
}

// NormalizeVal normalizes values to the Doltgres type expects, so it can be used to
// convert the values using the given Doltgres type. This is used to normalize array
// types as the type conversion expects certain type values.
func NormalizeVal(dt types.DoltgresType, v any) any {
	switch t := dt.(type) {
	case types.JsonType:
		str, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		return string(str)
	case types.JsonBType:
		jv, err := t.ConvertToJsonDocument(v)
		if err != nil {
			panic(err)
		}
		return types.JsonDocument{Value: jv}
	}

	switch val := v.(type) {
	case pgtype.Numeric:
		if val.NaN {
			return math.NaN()
		} else if val.InfinityModifier != pgtype.Finite {
			return math.Inf(int(val.InfinityModifier))
		} else if !val.Valid {
			return nil
		} else {
			return decimal.NewFromBigInt(val.Int, val.Exp)
		}
	case pgtype.Time:
		// This value type is used for TIME type.
		return timeofday.FromInt(val.Microseconds).ToTime()
	case pgtype.Interval:
		//This value type is used for INTERVAL type.
		return duration.MakeDuration(val.Microseconds*functions.NanosPerMicro, int64(val.Days), int64(val.Months))
	case [16]byte:
		// This value type is used for UUID type.
		u, err := uuid.FromBytes(val[:])
		if err != nil {
			panic(err)
		}
		return u
	case []any:
		baseType := dt
		if dta, ok := baseType.(types.DoltgresArrayType); ok {
			baseType = dta.BaseType()
		}
		newVal := make([]any, len(val))
		for i, el := range val {
			newVal[i] = NormalizeVal(baseType, el)
		}
		return newVal
	}
	return v
}

// NormalizeIntsAndFloats normalizes all int and float types
// to int64 and float64, respectively.
func NormalizeIntsAndFloats(v any) any {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int8:
		return int64(val)
	case int16:
		return int64(val)
	case int32:
		return int64(val)
	case uint:
		return int64(val)
	case uint8:
		return int64(val)
	case uint16:
		return int64(val)
	case uint32:
		return int64(val)
	case uint64:
		// PostgreSQL does not support an uint64 type, so we can always convert this to an int64 safely.
		return int64(val)
	case float32:
		return float64(val)
	default:
		return val
	}
}

// GetUnusedPort returns an unused port.
func GetUnusedPort(t *testing.T) int {
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	return port
}

// Numeric creates a numeric value from a string.
func Numeric(str string) pgtype.Numeric {
	numeric := pgtype.Numeric{}
	if err := numeric.Scan(str); err != nil {
		panic(err)
	}
	return numeric
}
