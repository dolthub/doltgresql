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
	"bufio"
	"context"
	"encoding/json"
	goerrors "errors"
	"fmt"
	"math"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/postgres/parser/timeofday"
	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	dserver "github.com/dolthub/doltgresql/server"
	"github.com/dolthub/doltgresql/server/auth"
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

// ExpectedNotice specifies what notices are expected during a script test assertion.
type ExpectedNotice struct {
	Severity string
	Message  string
}

// ScriptTestAssertion are the assertions upon which the script executes its main "testing" logic.
type ScriptTestAssertion struct {
	Query           string
	Expected        []sql.Row
	ExpectedErr     string
	ExpectedNotices []ExpectedNotice

	BindVars []any

	// SkipResultsCheck is used to skip assertions on the expected rows returned from a query. For now, this is
	// included as some messages do not have a full logical implementation. Skipping the results check allows us to
	// force the test client to not send of those messages.
	SkipResultsCheck bool

	// Skip is used to completely skip a test, not execute its query at all, and record it as a skipped test
	// in the test suite results.
	Skip bool

	// Username specifies the user's name to use for the command. This creates a new connection, using the given name.
	// By default (when the string is empty), the `postgres` superuser account is used. Any consecutive queries that
	// have the same username and password will reuse the same connection. The `postgres` superuser account will always
	// reuse the same connection. Do note that specifying the `postgres` account manually will create a connection
	// that is different from the primary one.
	Username string
	// Password specifies the password that will be used alongside the given username. This field is essentially ignored
	// when no username is given. If a username is given and the password is empty, then it is assumed that the password
	// is the empty string.
	Password string

	// ExpectedTag is used to check the command tag returned from the server.
	// This is checked only if no Expected is defined
	ExpectedTag string

	// Cols is used to check the column names returned from the server.
	Cols []string

	// CopyFromSTDIN is used to test the COPY FROM STDIN command.
	CopyFromStdInFile string
}

// Connection contains the default and current connections.
type Connection struct {
	Default  *pgx.Conn
	Current  *pgx.Conn
	Username string
	Password string
}

// receivedNotices tracks the NOTICE messages received over the connection to the Doltgres server, so that tests
// can assert what notices are expected to be sent to the client.
var receivedNotices []*pgconn.Notice

// RunScript runs the given script.
func RunScript(t *testing.T, script ScriptTest, normalizeRows bool) {
	scriptDatabase := script.Database
	if len(scriptDatabase) == 0 {
		scriptDatabase = "postgres"
	}

	var ctx context.Context
	var conn *Connection

	if runOnPostgres {
		ctx = context.Background()
		pgxConn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/%s?sslmode=disable", 5432, scriptDatabase))
		require.NoError(t, err)
		conn = &Connection{
			Default: pgxConn,
			Current: pgxConn,
		}
		require.NoError(t, pgxConn.Ping(ctx))
		defer func() {
			conn.Close(ctx)
		}()
	} else {
		var controller *svcs.Controller
		ctx, conn, controller = CreateServer(t, scriptDatabase)
		defer func() {
			conn.Close(ctx)
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
func runScript(t *testing.T, ctx context.Context, script ScriptTest, conn *Connection, normalizeRows bool) {
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

			// Clear out any previously received notices
			receivedNotices = nil

			// Use the provided username and password to create a new connection (if a username has been specified).
			// This will automatically handle connection reuse, using the default connection is no user is specified, etc.
			if err := conn.Connect(ctx, assertion.Username, assertion.Password); err != nil {
				if assertion.ExpectedErr != "" {
					require.Error(t, err)
					assert.Contains(t, err.Error(), assertion.ExpectedErr)
				} else {
					require.NoError(t, err)
				}
				return
			}
			// If we're skipping the results check, then we call Execute, as it uses a simplified message model.
			if assertion.CopyFromStdInFile != "" {
				copyFromStdin(t, conn.Current, assertion.Query, assertion.CopyFromStdInFile)
			} else if assertion.SkipResultsCheck || assertion.ExpectedErr != "" {
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

				// not an exact match but works well enough for our tests
				orderBy := strings.Contains(strings.ToLower(assertion.Query), "order by")

				if normalizeRows {
					if orderBy {
						assert.Equal(t, NormalizeExpectedRow(rows.FieldDescriptions(), assertion.Expected), readRows)
					} else {
						assert.ElementsMatch(t, NormalizeExpectedRow(rows.FieldDescriptions(), assertion.Expected), readRows)
					}
				} else {
					if orderBy {
						assert.Equal(t, assertion.Expected, readRows)
					} else {
						assert.ElementsMatch(t, assertion.Expected, readRows)
					}
				}

				if len(assertion.ExpectedNotices) > 0 {
					if len(assertion.ExpectedNotices) == len(receivedNotices) {
						for i, notice := range receivedNotices {
							assert.Equal(t, assertion.ExpectedNotices[i].Severity, notice.Severity)
							assert.Equal(t, assertion.ExpectedNotices[i].Message, notice.Message)
						}
					} else {
						if len(receivedNotices) == 0 {
							receivedNotices = []*pgconn.Notice{}
						}
						assert.Fail(t, "Received notices do not match expected notices",
							"Expected %d notices, but received %d. Expected: %v, Received: %v",
							len(assertion.ExpectedNotices), len(receivedNotices), assertion.ExpectedNotices, receivedNotices)
					}
				}
			}
		})
	}
}

func copyFromStdin(t *testing.T, conn *pgx.Conn, query string, filename string) {
	filePath := filepath.Join("testdata", filename)

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	_, err = conn.PgConn().CopyFrom(context.Background(), reader, query)
	require.NoError(t, err)
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

var testServerLogLevel = "info"

func init() {
	if logLevel, ok := os.LookupEnv("TEST_SERVER_LOG_LEVEL"); ok {
		testServerLogLevel = logLevel
	}
}

// CreateServer creates a server with the given database, returning a connection to the server. The server will close
// when the connection is closed (or loses its connection to the server). The accompanying WaitGroup may be used to wait
// until the server has closed.
func CreateServer(t *testing.T, database string) (context.Context, *Connection, *svcs.Controller) {
	require.NotEmpty(t, database)
	port := GetUnusedPort(t)
	controller, err := dserver.RunInMemory(&servercfg.DoltgresConfig{
		ListenerConfig: &servercfg.DoltgresListenerConfig{
			PortNumber: &port,
			HostStr:    ptr("127.0.0.1"),
		},
		LogLevelStr: ptr(testServerLogLevel),
	})
	require.NoError(t, err)
	auth.ClearDatabase()

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
		if database != "postgres" {
			_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s;", database))
			return err
		}
		return nil
	}()
	require.NoError(t, err)

	connectionString := fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/%s", port, database)
	config, err := pgx.ParseConfig(connectionString)
	require.NoError(t, err)
	config.OnNotice = func(c *pgconn.PgConn, notice *pgconn.Notice) {
		receivedNotices = append(receivedNotices, notice)
	}

	conn, err := pgx.ConnectConfig(ctx, config)
	require.NoError(t, err)

	// Ping the connection to test that it's alive, and also to test that Doltgres handles empty queries correctly
	require.NoError(t, conn.Ping(ctx))
	return ctx, &Connection{
		Default: conn,
		Current: conn,
	}, controller
}

// ReadRows reads all of the given rows into a slice, then closes the rows. If `normalizeRows` is true, then the rows
// will be normalized such that all integers are int64, etc.
func ReadRows(rows pgx.Rows, normalizeRows bool) (readRows []sql.Row, err error) {
	defer func() {
		err = goerrors.Join(err, rows.Err())
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
		dt, ok := types.IDToBuiltInDoltgresType[id.Type(id.Cache().ToInternal(fds[i].DataTypeOID))]
		if !ok {
			// try using text type
			dt = types.Text
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
				dt, ok := types.IDToBuiltInDoltgresType[id.Type(id.Cache().ToInternal(fds[i].DataTypeOID))]
				if !ok {
					// try using text type
					dt = types.Text
				}
				if dt.ID == types.Json.ID {
					newRow[i] = UnmarshalAndMarshalJsonString(row[i].(string))
				} else if dt.IsArrayType() && dt.ArrayBaseType().ID == types.Json.ID {
					// TODO: need to have valid sql.Context
					v, err := dt.IoInput(nil, row[i].(string))
					if err != nil {
						panic(err)
					}
					arr := v.([]any)
					newArr := make([]any, len(arr))
					for j, el := range arr {
						newArr[j] = UnmarshalAndMarshalJsonString(el.(string))
					}
					ret, err := dt.FormatValue(newArr)
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
func NormalizeValToString(dt *types.DoltgresType, v any) any {
	switch dt.ID {
	case types.Json.ID:
		str, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		ret, err := dt.FormatValue(string(str))
		if err != nil {
			panic(err)
		}
		return ret
	case types.JsonB.ID:
		jv, err := types.ConvertToJsonDocument(v)
		if err != nil {
			panic(err)
		}
		str, err := dt.FormatValue(types.JsonDocument{Value: jv})
		if err != nil {
			panic(err)
		}
		return str
	case types.InternalChar.ID:
		if v == nil {
			return nil
		}
		var b []byte
		if v.(int32) == 0 {
			b = []byte{}
		} else {
			b = []byte{uint8(v.(int32))}
		}
		val, err := dt.FormatValue(string(b))
		if err != nil {
			panic(err)
		}
		return val
	case types.Interval.ID, types.Uuid.ID, types.Date.ID, types.Time.ID, types.Timestamp.ID:
		// These values need to be normalized into the appropriate types
		// before being converted to string type using the Doltgres
		// IoOutput method.
		if v == nil {
			return nil
		}
		tVal, err := dt.FormatValue(NormalizeVal(dt, v))
		if err != nil {
			panic(err)
		}
		return tVal
	case types.TimestampTZ.ID:
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
		if dt.IsArrayType() {
			return NormalizeArrayType(dt, val)
		}
	}
	return v
}

// NormalizeArrayType normalizes array types by normalizing its elements first,
// then to a string using the type IoOutput method.
func NormalizeArrayType(dt *types.DoltgresType, arr []any) any {
	newVal := make([]any, len(arr))
	for i, el := range arr {
		newVal[i] = NormalizeVal(dt.ArrayBaseType(), el)
	}
	baseType := dt.ArrayBaseType()
	if baseType.ID == types.Bool.ID {
		sqlVal, err := dt.SQL(sql.NewEmptyContext(), nil, newVal)
		if err != nil {
			panic(err)
		}
		return sqlVal.ToString()
	} else {
		ret, err := dt.FormatValue(newVal)
		if err != nil {
			panic(err)
		}
		return ret
	}
}

// NormalizeVal normalizes values to the Doltgres type expects, so it can be used to
// convert the values using the given Doltgres type. This is used to normalize array
// types as the type conversion expects certain type values.
func NormalizeVal(dt *types.DoltgresType, v any) any {
	switch dt.ID {
	case types.Json.ID:
		str, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		return string(str)
	case types.JsonB.ID:
		jv, err := types.ConvertToJsonDocument(v)
		if err != nil {
			panic(err)
		}
		return types.JsonDocument{Value: jv}
	case types.Oid.ID, types.Regclass.ID, types.Regproc.ID, types.Regtype.ID:
		if uval, ok := v.(uint32); ok {
			if internalID := id.Cache().ToInternal(uval); internalID.IsValid() {
				return internalID
			}
			return id.NewOID(uval).AsId()
		}
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
		if baseType.IsArrayType() {
			baseType = baseType.ArrayBaseType()
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

// Connect replaces the Current connection with a new one, using the given username and password. If the username is
// empty, then the default connection is used. If the username and password match the existing connection, then no new
// connection is made.
func (conn *Connection) Connect(ctx context.Context, username string, password string) error {
	// Reuse the existing connection if it's the same username and password
	if username == conn.Username && password == conn.Password && conn.Current != nil {
		return nil
	}
	// Username or password has changed, so we'll close the Current connection only if it's not the default
	if conn.Username != "" && conn.Current != nil {
		_ = conn.Current.Close(ctx)
	}
	conn.Username = username
	conn.Password = password
	if username == "" {
		conn.Current = conn.Default
		return nil
	} else {
		var err error
		config := conn.Default.Config()
		conn.Current, err = pgx.Connect(ctx, fmt.Sprintf("postgres://%s:%s@127.0.0.1:%d/%s", username, password, config.Port, config.Database))
		return err
	}
}

// Exec calls Exec on the current connection.
func (conn *Connection) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if conn.Current == nil {
		return pgconn.CommandTag{}, errors.New("EXEC: current connection is nil")
	}
	return conn.Current.Exec(ctx, sql, args...)
}

// Query calls Query on the current connection.
func (conn *Connection) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if conn.Current == nil {
		return nil, errors.New("QUERY: current connection is nil")
	}
	return conn.Current.Query(ctx, sql, args...)
}

// Close closes the connections.
func (conn *Connection) Close(ctx context.Context) {
	if conn.Default != nil {
		_ = conn.Default.Close(ctx)
	}
	if conn.Current != nil && conn.Current != conn.Default {
		_ = conn.Current.Close(ctx)
	}
	conn.Default = nil
	conn.Current = nil
}
