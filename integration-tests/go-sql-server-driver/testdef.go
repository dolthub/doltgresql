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

package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"

	driver "github.com/dolthub/doltgresql/integration-tests/go-sql-server-driver/driver"
)

var GlobalPorts GlobalDynamicResources

type TestDef struct {
	Tests []Test `yaml:"tests"`

	// If true, RunTestfile will run each subtest in parallel.
	Parallel bool `yaml:"parallel"`
}

// Test is a single test to run. The Repos and MultiRepos will be created, and
// any Servers defined within them will be started. The interactions and
// assertions defined in Conns will be run.
type Test struct {
	Name       string              `yaml:"name"`
	Repos      []driver.TestRepo   `yaml:"repos"`
	MultiRepos []driver.MultiRepo  `yaml:"multi_repos"`
	Conns      []driver.Connection `yaml:"connections"`

	// Skip the entire test with this reason.
	Skip string `yaml:"skip"`
}

// Set this environment variable to effectively disable timeouts for debugging.
const debugEnvKey = "DOLTGRES_SQL_SERVER_TEST_DEBUG"

var timeout = 20 * time.Second

func init() {
	_, ok := os.LookupEnv(debugEnvKey)
	if ok {
		timeout = 1000 * time.Hour
	}
}

func ParseTestsFile(path string) (TestDef, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return TestDef{}, err
	}
	dec := yaml.NewDecoder(bytes.NewReader(contents))
	dec.KnownFields(true)
	var res TestDef
	err = dec.Decode(&res)
	return res, err
}

// MakeRepo creates the database for |r| within the store |rs| and writes any
// with_files for the repo. Unlike Dolt, the doltgres server runs from the
// data-dir (the store) and each repo is a database subdirectory, so with_files
// are written relative to the store directory.
func MakeRepo(t *testing.T, rs driver.RepoStore, r driver.TestRepo, ports *DynamicResources) driver.Repo {
	repo, err := rs.MakeRepo(r.Name)
	require.NoError(t, err)
	for _, f := range r.WithFiles {
		f.Template = func(s string) string {
			return ports.ApplyTemplate(s)
		}
		require.NoError(t, f.WriteAtDir(rs.Dir))
	}
	for _, remote := range r.WithRemotes {
		url := remote.URL
		url = ports.ApplyTemplate(url)
		require.NoError(t, repo.CreateRemote(remote.Name, url))
	}
	return repo
}

// Simple interface for wrapping *testing.T. Used for retryingT.
type TestingT interface {
	Fatal(...any)

	FailNow()

	Errorf(string, ...any)

	Cleanup(func())

	TempDir() string
}

// Globally available dynamic ports, backs every instance of
// DynamicPorts and hands them out in a thread-safe manner.
//
// XXX: This structure and its initialization does not currently look
// for "available" ports on the running host. It simply avoids handing
// out the same port to two separate tests that are running at the
// same time. It recycles ports as tests complete.
type GlobalDynamicResources struct {
	mu        sync.Mutex
	available []int
}

func (g *GlobalDynamicResources) GetPort(t TestingT) int {
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(g.available) == 0 {
		t.Fatal("cannot get a port; we are all out.")
	}
	next := g.available[len(g.available)-1]
	g.available = g.available[:len(g.available)-1]
	return next
}

func (g *GlobalDynamicResources) Return(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.available = append(g.available, n)
}

// Tracks dynamic resources available for expansion in test
// definitions, for example through `{{get_port ...}}` templates for
// server args and config files.
type DynamicResources struct {
	// Where we go when we need a new one.
	global *GlobalDynamicResources

	t TestingT

	// Where we put allocated ports. For a given test, the first
	// use will get a new unused port from
	// GlobalDynamicPorts. Then that same port will be returned
	// from here for all uses. When the test finishes, its Cleanup
	// returns the port to GlobalDynamicResources.
	allocatedPorts map[string]int

	// Where we put allocated temp directories. These get removed
	// on cleanup.
	allocatedTempDirs map[string]string
}

func (d *DynamicResources) GetPort(name string) (int, bool) {
	if d.allocatedPorts != nil {
		v, ok := d.allocatedPorts[name]
		return v, ok
	} else {
		return 0, false
	}
}

func (d *DynamicResources) GetTempDir(name string) (string, bool) {
	if d.allocatedTempDirs != nil {
		v, ok := d.allocatedTempDirs[name]
		return v, ok
	} else {
		return "", false
	}
}

func (d *DynamicResources) GetOrAllocatePort(name string) int {
	v, ok := d.GetPort(name)
	if ok {
		return v
	}
	v = d.global.GetPort(d.t)
	if d.allocatedPorts == nil {
		d.allocatedPorts = make(map[string]int)
		// We register one cleanup function for the entire
		// DynamicPorts and we return them all at once.
		//
		// In cases where there are two dependent servers, we
		// want to return all ports after both servers have
		// been shut down. If we return them as we allocated
		// them, it's possible that we allocated them to
		// render the entire config for the first server, some
		// referring to the second server, for example. If
		// testing.T runs cleanups in FIFO order, and the
		// Cleanup for running the second server is
		// responsible for shutting it down, it is possible we
		// would return the second server's ports before it is
		// shut down if we didn't return them all at once.
		d.t.Cleanup(func() {
			for _, p := range d.allocatedPorts {
				d.global.Return(p)
			}
		})
	}
	d.allocatedPorts[name] = v
	return v
}

func (d *DynamicResources) GetOrAllocateTempDir(name string) string {
	v, ok := d.GetTempDir(name)
	if ok {
		return v
	}
	v = d.t.TempDir()
	if d.allocatedTempDirs == nil {
		d.allocatedTempDirs = make(map[string]string)
		d.t.Cleanup(func() {
			d.allocatedTempDirs = nil
		})
	}
	d.allocatedTempDirs[name] = v
	return v
}

func (d *DynamicResources) ApplyTemplate(s string) string {
	tmpl, err := template.New("sql").Funcs(map[string]any{
		"get_port":    d.GetOrAllocatePort,
		"get_tempdir": d.GetOrAllocateTempDir,
	}).Parse(s)
	require.NoError(d.t, err)
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, nil)
	require.NoError(d.t, err)
	return buf.String()
}

// testLogWriter is an io.Writer that sends each line to t.Log.
// It buffers partial lines until a newline is received.
type testLogWriter struct {
	t   *testing.T
	buf []byte
}

func (w *testLogWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		w.t.Helper()
		w.t.Log(string(w.buf[:i]))
		w.buf = w.buf[i+1:]
	}
	return len(p), nil
}

// Flush any remaining partial line.
func (w *testLogWriter) Flush() {
	if len(w.buf) > 0 {
		w.t.Helper()
		w.t.Log(string(w.buf))
		w.buf = w.buf[:0]
	}
}

func newTestLogWriter(t *testing.T) io.Writer {
	w := &testLogWriter{t: t}
	t.Cleanup(w.Flush)
	return w
}

// prepareDoltgresServerArgs translates the server |args| from a test
// definition into arguments valid for the doltgres binary, generating a config
// file in |cwd| that fixes the listener to |port| (and a unique unix socket so
// parallel servers do not collide). doltgres has a minimal CLI (no -P, -l,
// --max-connections, etc.), so the listener port can only be set via the
// config file.
//
// If the incoming args reference a config file via `--config <path>` (or
// `--config=<path>`), that file is loaded as the base config and the listener
// settings are overlaid onto it; otherwise a fresh config is generated. The
// `--data-dir` argument is added (as `.`) if not already present.
func prepareDoltgresServerArgs(t *testing.T, cwd, name string, port int, args []string) []string {
	base := map[string]any{}
	var passthrough []string
	configPath := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--config" || a == "-config":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			}
		case len(a) > 9 && a[:9] == "--config=":
			configPath = a[9:]
		case len(a) > 8 && a[:8] == "-config=":
			configPath = a[8:]
		default:
			passthrough = append(passthrough, a)
		}
	}

	if configPath != "" {
		p := configPath
		if !filepath.IsAbs(p) {
			p = filepath.Join(cwd, p)
		}
		if contents, err := os.ReadFile(p); err == nil {
			_ = yaml.Unmarshal(contents, &base)
		} else {
			require.NoError(t, err)
		}
	}

	listener, _ := base["listener"].(map[string]any)
	if listener == nil {
		listener = map[string]any{}
	}
	listener["host"] = "127.0.0.1"
	listener["port"] = port
	listener["socket"] = filepath.Join(os.TempDir(), fmt.Sprintf("dg-%d.sock", port))
	base["listener"] = listener

	out, err := yaml.Marshal(base)
	require.NoError(t, err)
	genPath := filepath.Join(cwd, fmt.Sprintf(".generated-%s-config.yaml", sanitizeName(name)))
	require.NoError(t, os.WriteFile(genPath, out, 0640))

	hasDataDir := false
	for _, a := range passthrough {
		if a == "--data-dir" || a == "-data-dir" || (len(a) > 11 && a[:11] == "--data-dir=") || (len(a) > 10 && a[:10] == "-data-dir=") {
			hasDataDir = true
		}
	}
	result := append([]string{}, passthrough...)
	if !hasDataDir {
		result = append(result, "--data-dir=.")
	}
	result = append(result, "--config="+genPath)
	return result
}

func sanitizeName(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '/' || r == '\\' || r == ' ' {
			out = append(out, '_')
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}

func MakeServer(t *testing.T, dc driver.DoltCmdable, cwd string, s *driver.Server, resources *DynamicResources, manualOps ...driver.SqlServerOpt) *driver.SqlServer {
	if s == nil {
		return nil
	}
	if s.Port != 0 {
		t.Fatal("cannot specify s.Port on these tests; please use {{get_port ...}} and dynamic_port: to specify a dynamic port.")
	}
	if s.DynamicPort == "" {
		t.Fatal("you must specify s.DynamicPort on these tests; please use {{get_port ...}} and dynamic_port: to specify a dynamic port.")
	}
	port := resources.GetOrAllocatePort(s.DynamicPort)

	rawArgs := make([]string, len(s.Args))
	for i := range rawArgs {
		rawArgs[i] = resources.ApplyTemplate(s.Args[i])
	}
	args := prepareDoltgresServerArgs(t, cwd, s.Name, port, rawArgs)

	opts := []driver.SqlServerOpt{
		driver.WithArgs(args...),
		driver.WithName(s.Name),
		driver.WithLogWriter(newTestLogWriter(t)),
		driver.WithPort(port),
	}
	if len(s.Envs) > 0 {
		opts = append(opts, driver.WithEnvs(s.Envs...))
	}
	opts = append(opts, manualOps...)

	var server *driver.SqlServer
	var err error
	if s.DebugPort != 0 {
		server, err = driver.DebugSqlServer(dc, s.DebugPort, opts...)
	} else {
		server, err = driver.StartSqlServer(dc, opts...)
	}

	require.NoError(t, err)
	if len(s.ErrorMatches) > 0 {
		err := server.ErrorStop()
		require.Error(t, err)
		output := server.Output.String()
		for _, a := range s.ErrorMatches {
			require.Regexp(t, a, output)
		}
		return nil
	} else {
		t.Cleanup(func() {
			// We use assert, not require here, since FailNow() in
			// a Cleanup does not make sense.
			err := server.GracefulStop()
			if assert.NoError(t, err) {
				output := server.Output.String()
				for _, a := range s.LogMatches {
					assert.Regexp(t, a, output)
				}
				for _, a := range s.LogNotMatches {
					assert.NotRegexp(t, a, output)
				}
			}
		})

		return server
	}
}

// Runs the defined test, applying its asserts.
func (test Test) Run(t *testing.T) {
	if test.Skip != "" {
		t.Skip(test.Skip)
	}

	var ports DynamicResources
	ports.global = &GlobalPorts
	ports.t = t

	u, err := driver.NewDoltUser()
	require.NoError(t, err)
	t.Cleanup(func() {
		u.Cleanup()
	})

	servers := make(map[string]*driver.SqlServer)

	for _, r := range test.Repos {
		// Each repo with a server gets its own store (data-dir) so that
		// independent doltgres server processes do not collide on the same
		// on-disk databases.
		rs, err := u.MakeRepoStore()
		require.NoError(t, err)

		MakeRepo(t, rs, r, &ports)

		if r.Server != nil {
			if r.Server.Name == "" {
				r.Server.Name = r.Name
			}
			server := MakeServer(t, rs, rs.Dir, r.Server, &ports)
			if server != nil {
				server.DBName = r.Name
				servers[r.Name] = server
			}
		}
	}

	for _, mr := range test.MultiRepos {
		// Each MultiRepo gets its own data-dir.
		rs, err := u.MakeRepoStore()
		require.NoError(t, err)
		for _, r := range mr.Repos {
			MakeRepo(t, rs, r, &ports)
		}
		for _, f := range mr.WithFiles {
			f.Template = func(s string) string {
				return ports.ApplyTemplate(s)
			}
			require.NoError(t, f.WriteAtDir(rs.Dir))
		}
		if mr.Server != nil {
			if mr.Server.Name == "" {
				mr.Server.Name = mr.Name
			}
			server := MakeServer(t, rs, rs.Dir, mr.Server, &ports)
			if server != nil {
				servers[mr.Name] = server
			}
		}
	}

	for i, c := range test.Conns {
		server := servers[c.On]
		require.NotNilf(t, server, "error in test spec: could not find server %s for connection %d", c.On, i)
		if c.RetryAttempts > 1 {
			RetryTestRun(t, c.RetryAttempts, func(t TestingT) {
				db, err := server.DB(c)
				require.NoError(t, err)
				defer db.Close()

				conn, err := db.Conn(context.Background())
				require.NoError(t, err)
				defer conn.Close()

				for _, q := range c.Queries {
					RunQueryAttempt(t, conn, q, &ports)
				}
			})
		} else {
			func() {
				db, err := server.DB(c)
				require.NoError(t, err)
				defer db.Close()

				conn, err := db.Conn(context.Background())
				require.NoError(t, err)
				defer conn.Close()

				for _, q := range c.Queries {
					RunQuery(t, conn, q, &ports)
				}
			}()
		}
		if c.RestartServer != nil {
			args := c.RestartServer.Args
			if args != nil {
				tmplArgs := make([]string, len(*args))
				for i := range tmplArgs {
					tmplArgs[i] = ports.ApplyTemplate((*args)[i])
				}
				prepared := prepareDoltgresServerArgs(t, server.Cmd.Dir, server.Name, server.Port, tmplArgs)
				args = &prepared
			}
			err := server.Restart(args, c.RestartServer.Envs)
			require.NoError(t, err)
		}
	}
}

func RunTestsFile(t *testing.T, path string) {
	def, err := ParseTestsFile(path)
	require.NoError(t, err)
	parallel := def.Parallel
	for _, test := range def.Tests {
		t.Run(test.Name, func(t *testing.T) {
			if parallel {
				t.Parallel()
			}
			test.Run(t)
		})
	}
}

func RunSingleTest(t *testing.T, path string, testName string) {
	def, err := ParseTestsFile(path)
	require.NoError(t, err)
	for _, test := range def.Tests {
		if test.Name == testName {
			t.Run(test.Name, test.Run)
		}
	}
}

type retryTestingT struct {
	*testing.T
	errorfStrings []string
	errorfArgs    [][]interface{}
	failNow       bool
}

func (r *retryTestingT) Errorf(format string, args ...interface{}) {
	r.T.Helper()
	r.errorfStrings = append(r.errorfStrings, format)
	r.errorfArgs = append(r.errorfArgs, args)
}

func (r *retryTestingT) FailNow() {
	r.T.Helper()
	r.failNow = true
	panic(r)
}

func (r *retryTestingT) try(attempts int, test func(TestingT)) {
	for i := 0; i < attempts; i++ {
		r.errorfStrings = nil
		r.errorfArgs = nil
		r.failNow = false
		if i != 0 {
			time.Sleep(driver.RetrySleepDuration)
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(*retryTestingT); ok {
					} else {
						panic(r)
					}
				}
			}()
			test(r)
		}()
		if !r.failNow && len(r.errorfStrings) == 0 {
			return
		}
	}
	for i := range r.errorfStrings {
		r.T.Errorf(r.errorfStrings[i], r.errorfArgs[i]...)
	}
	if r.failNow {
		r.T.FailNow()
	}
}

func RetryTestRun(t *testing.T, attempts int, test func(TestingT)) {
	if attempts == 0 {
		attempts = 1
	}
	rtt := &retryTestingT{T: t}
	rtt.try(attempts, test)
}

func RunQuery(t *testing.T, conn *sql.Conn, q driver.Query, ports *DynamicResources) {
	RetryTestRun(t, q.RetryAttempts, func(t TestingT) {
		RunQueryAttempt(t, conn, q, ports)
	})
}

func RunQueryAttempt(t TestingT, conn *sql.Conn, q driver.Query, ports *DynamicResources) {
	args := make([]any, len(q.Args))
	for i := range q.Args {
		args[i] = q.Args[i]
	}
	if q.Query != "" {
		ctx, c := context.WithTimeout(context.Background(), timeout)
		defer c()
		rows, err := conn.QueryContext(ctx, q.Query, args...)
		if err == nil {
			defer rows.Close()
		}
		if q.ErrorMatch != "" {
			require.Error(t, err, "expected error running query %s", q.Query)
			require.Regexp(t, q.ErrorMatch, err.Error())
			return
		}
		require.NoError(t, err)

		cols, err := rows.Columns()
		require.NoError(t, err)
		require.Equal(t, q.Result.Columns, cols)

		rowstrings, err := RowsToStrings(len(cols), rows)
		require.NoError(t, err)
		if q.Result.Rows.Or != nil {
			match := *q.Result.Rows.Or
			for i := range match {
				for j := range match[i] {
					for k := range match[i][j] {
						match[i][j][k] = ports.ApplyTemplate(match[i][j][k])
					}
				}
			}
			require.Contains(t, match, rowstrings)
		}
	} else if q.Exec != "" {
		ctx, c := context.WithTimeout(context.Background(), timeout)
		defer c()
		exec := q.Exec
		exec = ports.ApplyTemplate(exec)
		_, err := conn.ExecContext(ctx, exec, args...)
		if q.ErrorMatch == "" {
			require.NoError(t, err, "error running query %s: %v", q.Exec, err)
		} else {
			require.Error(t, err)
			require.Regexp(t, q.ErrorMatch, err.Error())
		}
	}
}

func RowsToStrings(cols int, rows *sql.Rows) ([][]string, error) {
	ret := make([][]string, 0)
	for rows.Next() {
		scanned := make([]any, cols)
		for j := range scanned {
			scanned[j] = new(sql.NullString)
		}
		err := rows.Scan(scanned...)
		if err != nil {
			return nil, err
		}
		printed := make([]string, cols)
		for j := range scanned {
			s := scanned[j].(*sql.NullString)
			if !s.Valid {
				printed[j] = "NULL"
			} else {
				printed[j] = s.String
			}
		}
		ret = append(ret, printed)
	}
	return ret, rows.Err()
}
