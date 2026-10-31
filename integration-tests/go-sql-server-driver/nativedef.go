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

package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/dolthub/doltgresql/integration-tests/go-sql-server-driver/driver"
)

// The native cluster tests run the semantics of the Doltgres cluster tests
// against the Postgres fork with prolly storage, using Postgres's own
// replication: streaming hot standbys made with pg_basebackup, promotion with
// pg_promote, and a former primary rejoining with pg_rewind. They run only when
// NativeBinDirEnvKey names the fork's bin directory.
const NativeBinDirEnvKey = "DOLTGRES_NATIVE_BINDIR"

// NativeTableAMEnvKey overrides the default table access method of native
// servers, e.g. "heap" for a control run. Prolly storage by default.
const NativeTableAMEnvKey = "DOLTGRES_NATIVE_TABLE_AM"

// NativeTestDef is a file of native cluster tests.
type NativeTestDef struct {
	Tests    []NativeTest `yaml:"tests"`
	Parallel bool         `yaml:"parallel"`
}

// NativeTest starts |Servers| and runs |Steps| in order.
type NativeTest struct {
	Name    string         `yaml:"name"`
	Skip    string         `yaml:"skip"`
	Servers []NativeServer `yaml:"servers"`
	Steps   []NativeStep   `yaml:"steps"`
}

// NativeServer is a Postgres server. One that |Follows| another is a hot
// standby, copied from it after the primaries are set up.
type NativeServer struct {
	Name    string `yaml:"name"`
	Follows string `yaml:"follows"`

	// Databases created on a primary before its standbys are copied.
	Databases []string `yaml:"databases"`

	// Lines appended to the server's postgresql.conf.
	Config []string `yaml:"config"`

	// Set up, but not started.
	Stopped bool `yaml:"stopped"`
}

// NativeStep is either a connection running queries against a server, as in
// the Doltgres tests, or an action on the cluster.
type NativeStep struct {
	On            string         `yaml:"on"`
	Database      string         `yaml:"database"`
	User          string         `yaml:"user"`
	Queries       []driver.Query `yaml:"queries"`
	RetryAttempts int            `yaml:"retry_attempts"`
	Action        *NativeAction  `yaml:"action"`
}

// NativeAction changes a server's role or state. Kinds:
//
//   - promote: the standby |Server| becomes a primary (pg_promote).
//   - demote: |Server| stops and rejoins as a standby of |Follows|, rewound
//     to its timeline (pg_rewind).
//   - switchover: the graceful transition of the Doltgres tests. Primary
//     |Server| stops cleanly, so that its standby |To| has every write; |To|
//     is promoted; |Server| rejoins as a standby of |To|.
//   - stop, kill, start, restart: stop cleanly, stop immediately, start,
//     stop cleanly and start.
//   - wait_for_replay: the standby |Server| has replayed everything its
//     primary has written so far.
type NativeAction struct {
	Kind    string `yaml:"kind"`
	Server  string `yaml:"server"`
	Follows string `yaml:"follows"`
	To      string `yaml:"to"`
}

// nativeServer is a running or stopped server of a native test.
type nativeServer struct {
	name    string
	dir     string
	sockets string // short, as socket paths are limited in length
	port    int
	follows string
	running bool
}

type nativeCluster struct {
	t       *testing.T
	root    string // holds the servers' data directories and logs
	bindir  string
	servers map[string]*nativeServer
	order   []string
	ports   *DynamicResources
}

func ParseNativeTestsFile(path string) (NativeTestDef, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return NativeTestDef{}, err
	}

	var res NativeTestDef
	dec := yaml.NewDecoder(strings.NewReader(string(contents)))
	dec.KnownFields(true)
	err = dec.Decode(&res)
	return res, err
}

// RunNativeTestsFile runs the native tests of |path|, if NativeBinDirEnvKey is
// set.
func RunNativeTestsFile(t *testing.T, path string) {
	bindir := os.Getenv(NativeBinDirEnvKey)
	if bindir == "" {
		t.Skipf("set %s to the Postgres fork's bin directory to run native cluster tests", NativeBinDirEnvKey)
	}

	def, err := ParseNativeTestsFile(path)
	require.NoError(t, err)
	for _, test := range def.Tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			if def.Parallel {
				t.Parallel()
			}
			test.Run(t, bindir)
		})
	}
}

func (test NativeTest) Run(t *testing.T, bindir string) {
	if test.Skip != "" {
		t.Skip(test.Skip)
	}

	ports := &DynamicResources{
		global: &GlobalPorts,
		t:      t,
	}
	// Made before stopAll is registered, so that it outlives stopAll, which
	// reports the servers' logs: cleanups run last registered first.
	c := &nativeCluster{
		t:       t,
		root:    t.TempDir(),
		bindir:  bindir,
		servers: make(map[string]*nativeServer),
		ports:   ports,
	}
	t.Cleanup(c.stopAll)
	c.setup(test.Servers)
	for i, step := range test.Steps {
		if step.Action != nil {
			c.act(*step.Action)
			continue
		}

		server := c.servers[step.On]
		require.NotNilf(t, server, "error in test spec: could not find server %s for step %d", step.On, i)
		run := func(t TestingT) {
			db, err := c.connect(server, step.Database, step.User)
			require.NoError(t, err)
			defer db.Close()

			conn, err := db.Conn(context.Background())
			require.NoError(t, err)
			defer conn.Close()

			for _, q := range step.Queries {
				if q.RetryAttempts > 1 {
					RetryTestRun(c.t, q.RetryAttempts, func(t TestingT) {
						RunQueryAttempt(t, conn, q, ports)
					})
				} else {
					RunQueryAttempt(t, conn, q, ports)
				}
			}
		}
		if step.RetryAttempts > 1 {
			RetryTestRun(t, step.RetryAttempts, run)
		} else {
			run(t)
		}
	}
}

// setup makes and starts the primaries, creates their databases, and then
// copies each standby from the server it follows.
func (c *nativeCluster) setup(servers []NativeServer) {
	for _, s := range servers {
		sockets, err := os.MkdirTemp("", "pgs")
		require.NoError(c.t, err)
		c.t.Cleanup(func() {
			os.RemoveAll(sockets)
		})
		server := &nativeServer{
			name:    s.Name,
			dir:     filepath.Join(c.root, s.Name),
			sockets: sockets,
			port:    c.freePort(s.Name),
			follows: s.Follows,
		}
		c.servers[s.Name] = server
		c.order = append(c.order, s.Name)
	}

	for _, s := range servers {
		if s.Follows != "" {
			continue
		}

		server := c.servers[s.Name]
		// Dynamic shared memory in files: /dev/shm is small in containers,
		// and many servers run at once.
		c.run("initdb", "-D", server.dir, "-U", "postgres", "--auth=trust", "--encoding=UTF8", "--locale=C",
			"-c", "dynamic_shared_memory_type=mmap")
		c.configure(server, s.Config, true)
		c.start(server)
		for _, name := range s.Databases {
			c.exec(server, "postgres", fmt.Sprintf("CREATE DATABASE %s", name))
		}

		if s.Stopped {
			c.stop(server, "fast")
		}
	}

	// A standby may follow another standby; copy them in dependency order.
	for done := false; !done; {
		done = true
		for _, s := range servers {
			server := c.servers[s.Name]
			if s.Follows == "" || fileExists(filepath.Join(server.dir, "PG_VERSION")) {
				continue
			}

			upstream := c.servers[s.Follows]
			require.NotNilf(c.t, upstream, "error in test spec: %s follows unknown server %s", s.Name, s.Follows)
			if !upstream.running {
				done = false
				continue
			}

			c.run("pg_basebackup", "-h", "127.0.0.1", "-p", fmt.Sprint(upstream.port), "-U", "postgres",
				"-D", server.dir, "-R", "-X", "stream", "-c", "fast")
			c.configure(server, s.Config, false)
			if !s.Stopped {
				c.start(server)
			}
		}
	}
}

// configure appends the settings of |server| to its postgresql.conf. A copied
// standby inherits its primary's settings, so it only overrides the port and
// socket directory.
func (c *nativeCluster) configure(server *nativeServer, extra []string, primary bool) {
	lines := []string{
		fmt.Sprintf("port = %d", server.port),
		fmt.Sprintf("unix_socket_directories = '%s'", server.sockets),
	}
	if primary {
		am := os.Getenv(NativeTableAMEnvKey)
		if am == "" {
			am = "dolt_prolly_table"
		}

		lines = append(lines,
			"listen_addresses = '127.0.0.1'",
			"wal_level = replica",
			"max_wal_senders = 10",
			"hot_standby = on",
			"wal_log_hints = on",
			"max_connections = 50",
			"dynamic_shared_memory_type = mmap",
			fmt.Sprintf("default_table_access_method = '%s'", am),
		)
	}

	lines = append(lines, extra...)
	appendFile(c.t, filepath.Join(server.dir, "postgresql.conf"), strings.Join(lines, "\n")+"\n")
	if primary {
		appendFile(c.t, filepath.Join(server.dir, "pg_hba.conf"), "host replication all 127.0.0.1/32 trust\n")
	}
}

// freePort allocates a port for server |name| that nothing else is listening
// on: other processes on the host may hold ports in the shared pool.
func (c *nativeCluster) freePort(name string) int {
	for attempt := 0; ; attempt++ {
		key := name
		if attempt > 0 {
			key = fmt.Sprintf("%s#%d", name, attempt)
		}

		port := c.ports.GetOrAllocatePort(key)
		l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			l.Close()
			return port
		}

		require.Lessf(c.t, attempt, 32, "no free port for %s", name)
	}
}

func (c *nativeCluster) act(a NativeAction) {
	server := c.servers[a.Server]
	require.NotNilf(c.t, server, "error in test spec: action %s on unknown server %s", a.Kind, a.Server)
	switch a.Kind {
	case "promote":
		c.promote(server)
	case "demote":
		upstream := c.servers[a.Follows]
		require.NotNilf(c.t, upstream, "error in test spec: demote %s to follow unknown server %s", a.Server, a.Follows)
		c.demote(server, upstream)
	case "switchover":
		to := c.servers[a.To]
		require.NotNilf(c.t, to, "error in test spec: switchover from %s to unknown server %s", a.Server, a.To)
		require.Equalf(c.t, server.name, to.follows, "error in test spec: %s does not follow %s", a.To, a.Server)

		// A clean shutdown waits until connected standbys have received
		// all of the primary's WAL, so |to| has every write.
		c.stop(server, "fast")
		c.promote(to)
		c.demote(server, to)
	case "stop":
		c.stop(server, "fast")
	case "kill":
		c.stop(server, "immediate")
	case "start":
		c.start(server)
	case "restart":
		c.stop(server, "fast")
		c.start(server)
	case "wait_for_replay":
		c.waitForReplay(server)
	default:
		c.t.Fatalf("error in test spec: unknown action %s", a.Kind)
	}
}

func (c *nativeCluster) promote(server *nativeServer) {
	db, err := c.connect(server, "postgres", "")
	require.NoError(c.t, err)
	defer db.Close()

	var promoted bool
	require.NoError(c.t, db.QueryRow("SELECT pg_promote(true, 60)").Scan(&promoted))
	require.Truef(c.t, promoted, "%s was not promoted", server.name)
	server.follows = ""
}

// demote stops |server| and starts it again as a standby of |upstream|,
// rewound first if its history diverged from |upstream|'s.
func (c *nativeCluster) demote(server, upstream *nativeServer) {
	if server.running {
		c.stop(server, "fast")
	}

	c.run("pg_rewind", "--target-pgdata="+server.dir,
		fmt.Sprintf("--source-server=host=127.0.0.1 port=%d user=postgres dbname=postgres", upstream.port))
	require.NoError(c.t, os.WriteFile(filepath.Join(server.dir, "standby.signal"), nil, 0600))
	appendFile(c.t, filepath.Join(server.dir, "postgresql.auto.conf"),
		fmt.Sprintf("primary_conninfo = 'host=127.0.0.1 port=%d user=postgres'\n", upstream.port))
	server.follows = upstream.name
	c.start(server)
}

// waitForReplay waits until the standby |server| has replayed the WAL its
// primary has written so far.
func (c *nativeCluster) waitForReplay(server *nativeServer) {
	upstream := c.servers[server.follows]
	require.NotNilf(c.t, upstream, "%s is not a standby", server.name)

	var target string
	c.queryRow(upstream, "SELECT pg_current_wal_lsn()::text", &target)
	deadline := time.Now().Add(60 * time.Second)
	for {
		var caughtUp bool
		c.queryRow(server, fmt.Sprintf("SELECT pg_last_wal_replay_lsn() >= '%s'::pg_lsn", target), &caughtUp)
		if caughtUp {
			return
		}

		require.Truef(c.t, time.Now().Before(deadline), "%s did not replay to %s", server.name, target)
		time.Sleep(50 * time.Millisecond)
	}
}

func (c *nativeCluster) start(server *nativeServer) {
	c.run("pg_ctl", "-D", server.dir, "-l", server.dir+".log", "-w", "-t", "60", "start")
	server.running = true
}

func (c *nativeCluster) stop(server *nativeServer, mode string) {
	c.run("pg_ctl", "-D", server.dir, "-m", mode, "-w", "-t", "60", "stop")
	server.running = false
}

func (c *nativeCluster) stopAll() {
	for _, name := range c.order {
		server := c.servers[name]
		if server.running {
			_ = exec.Command(filepath.Join(c.bindir, "pg_ctl"), "-D", server.dir, "-m", "immediate", "-w", "stop").Run()
			server.running = false
		}

		if log, err := os.ReadFile(server.dir + ".log"); err == nil && c.t.Failed() {
			c.t.Logf("%s log:\n%s", server.name, log)
		}
	}
}

func (c *nativeCluster) connect(server *nativeServer, database, user string) (*sql.DB, error) {
	if database == "" {
		database = "postgres"
	}

	if user == "" {
		user = "postgres"
	}

	return driver.ConnectDB(user, "password", database, "127.0.0.1", server.port, map[string]string{"sslmode": "disable"})
}

func (c *nativeCluster) exec(server *nativeServer, database, statement string) {
	db, err := c.connect(server, database, "")
	require.NoError(c.t, err)
	defer db.Close()

	_, err = db.Exec(statement)
	require.NoError(c.t, err)
}

func (c *nativeCluster) queryRow(server *nativeServer, query string, dest any) {
	db, err := c.connect(server, "postgres", "")
	require.NoError(c.t, err)
	defer db.Close()

	require.NoError(c.t, db.QueryRow(query).Scan(dest))
}

// run runs one of the fork's programs, failing the test with its output if it
// fails.
func (c *nativeCluster) run(program string, args ...string) {
	cmd := exec.Command(filepath.Join(c.bindir, program), args...)
	out, err := cmd.CombinedOutput()
	require.NoErrorf(c.t, err, "%s %s:\n%s", program, strings.Join(args, " "), out)
}

func appendFile(t *testing.T, path, contents string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	require.NoError(t, err)
	defer f.Close()

	_, err = f.WriteString(contents)
	require.NoError(t, err)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
