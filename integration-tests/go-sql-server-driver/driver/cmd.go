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

package driver

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

var DoltgresPath string
var DelvePath string

const TestUserName = "Bats Tests"
const TestEmailAddress = "bats@email.fake"

const ConnectAttempts = 50
const RetrySleepDuration = 50 * time.Millisecond

// EnvDoltgresBinPath is the environment variable used to locate the doltgres
// binary used by these tests. If unset, "doltgres" on the PATH is used.
const EnvDoltgresBinPath = "DOLTGRES_BIN_PATH"

func init() {
	path := os.Getenv(EnvDoltgresBinPath)
	if path == "" {
		path = "doltgres"
	}
	path = filepath.Clean(path)
	var err error

	DoltgresPath, err = exec.LookPath(path)
	if err != nil {
		log.Printf("did not find doltgres binary: %v\n", err.Error())
	}

	DelvePath, _ = exec.LookPath("dlv")
}

// DoltUser is an abstraction for a user account that calls the `doltgres`
// server. All of our doltgres binary invocations are done through DoltUser.
//
// For our purposes, it does the following:
// * owns a tmpdir, to which it sets DOLT_ROOT_PATH when invoking doltgres.
// * writes some initial dolt global config (user.name, user.email,
//   metrics.disabled = true) into that root path. Doltgres has no `config`
//   CLI command, so unlike Dolt we write the config file directly.
//
// * can create repo stores, which will be a tmpdir to store a data-dir
//   containing one or more databases.
type DoltUser struct {
	tmpdir string
}

var _ DoltCmdable = DoltUser{}
var _ DoltDebuggable = DoltUser{}

func NewDoltUser() (DoltUser, error) {
	tmpdir, err := os.MkdirTemp("", "go-sql-server-driver-")
	if err != nil {
		return DoltUser{}, err
	}
	res := DoltUser{tmpdir}
	// Doltgres has no `config --global` CLI command, so we write the global
	// config file directly. Doltgres uses Dolt's env loading, which reads the
	// global config from $DOLT_ROOT_PATH/.dolt/config_global.json.
	cfgDir := filepath.Join(tmpdir, ".dolt")
	if err := os.MkdirAll(cfgDir, 0750); err != nil {
		return DoltUser{}, err
	}
	contents := fmt.Sprintf(`{"metrics.disabled":"true","user.name":%q,"user.email":%q}`+"\n", TestUserName, TestEmailAddress)
	if err := os.WriteFile(filepath.Join(cfgDir, "config_global.json"), []byte(contents), 0640); err != nil {
		return DoltUser{}, err
	}
	return res, nil
}

func (u DoltUser) DoltCmd(args ...string) *exec.Cmd {
	cmd := exec.Command(DoltgresPath, args...)
	cmd.Dir = u.tmpdir
	cmd.Env = append(os.Environ(), "DOLT_ROOT_PATH="+u.tmpdir)
	ApplyCmdAttributes(cmd)
	return cmd
}

func (u DoltUser) DoltDebug(debuggerPort int, args ...string) *exec.Cmd {
	if DelvePath != "" {
		dlvArgs := []string{
			fmt.Sprintf("--listen=:%d", debuggerPort),
			"--headless",
			"--api-version=2",
			"--accept-multiclient",
			"exec",
			DoltgresPath,
			"--",
		}
		cmd := exec.Command(DelvePath, append(dlvArgs, args...)...)
		cmd.Dir = u.tmpdir
		cmd.Env = append(os.Environ(), "DOLT_ROOT_PATH="+u.tmpdir)
		ApplyCmdAttributes(cmd)
		return cmd
	} else {
		panic("dlv not found")
	}
}

func (u DoltUser) MakeRepoStore() (RepoStore, error) {
	tmpdir, err := os.MkdirTemp(u.tmpdir, "repo-store-")
	if err != nil {
		return RepoStore{}, err
	}
	return RepoStore{u, tmpdir}, nil
}

func (u DoltUser) Cleanup() error {
	return os.RemoveAll(u.tmpdir)
}

type RepoStore struct {
	user DoltUser
	Dir  string
}

var _ DoltCmdable = RepoStore{}
var _ DoltDebuggable = RepoStore{}

// MakeRepo creates a new doltgres database named |name| within this store's
// data-dir. Because doltgres has no `init` CLI command, the database is
// initialized by briefly running a doltgres server with DOLTGRES_DB set to the
// database name and waiting for the on-disk database to be created.
func (rs RepoStore) MakeRepo(name string) (Repo, error) {
	ret := Repo{rs.user, rs, filepath.Join(rs.Dir, name), name}
	err := rs.initDatabase(name, nil)
	if err != nil {
		return Repo{}, err
	}
	return ret, nil
}

func (rs RepoStore) DoltCmd(args ...string) *exec.Cmd {
	cmd := rs.user.DoltCmd(args...)
	cmd.Dir = rs.Dir
	return cmd
}

func (rs RepoStore) DoltDebug(debuggerPort int, args ...string) *exec.Cmd {
	cmd := rs.user.DoltDebug(debuggerPort, args...)
	cmd.Dir = rs.Dir
	return cmd
}

// initDatabase starts a short-lived doltgres server bound to a free port with
// the data-dir set to this store, creates the database |name| if it does not
// already exist, runs the optional |fn| against a connection to that database,
// and then shuts the server down. This is the doltgres equivalent of
// `dolt init` plus any pre-server setup such as adding remotes.
func (rs RepoStore) initDatabase(name string, fn func(db *sql.DB) error) error {
	port, err := freePort()
	if err != nil {
		return err
	}
	cfgPath := filepath.Join(rs.Dir, fmt.Sprintf(".init-%s-config.yaml", sanitize(name)))
	// The unix socket path must stay well under the OS limit (~108 chars), so
	// it is placed directly in the temp dir keyed by the (unique) port rather
	// than inside the potentially-long data-dir path.
	sockPath := filepath.Join(os.TempDir(), fmt.Sprintf("dg-init-%d.sock", port))
	cfgContents := fmt.Sprintf("log_level: warn\nlistener:\n  host: 127.0.0.1\n  port: %d\n  socket: %s\n", port, sockPath)
	if err := os.WriteFile(cfgPath, []byte(cfgContents), 0640); err != nil {
		return err
	}
	defer os.Remove(cfgPath)

	cmd := rs.DoltCmd("--data-dir=.", "--config="+cfgPath)
	cmd.Env = append(cmd.Env, "DOLTGRES_DB="+name)
	output := new(bytes.Buffer)
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Start(); err != nil {
		return err
	}
	defer func() {
		_ = cmd.Process.Signal(os.Interrupt)
		_, _ = cmd.Process.Wait()
	}()

	db, err := ConnectDB("postgres", "password", name, "127.0.0.1", port, nil)
	if err != nil {
		return fmt.Errorf("could not connect to init server for %s: %w (output: %s)", name, err, output.String())
	}
	defer db.Close()

	if _, err := db.Exec(fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s`, name)); err != nil {
		return err
	}
	if fn != nil {
		if err := fn(db); err != nil {
			return err
		}
	}
	return nil
}

func sanitize(s string) string {
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

// freePort asks the OS for an available TCP port on the loopback interface.
func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

type Repo struct {
	user DoltUser
	rs   RepoStore
	Dir  string
	Name string
}

func (r Repo) DoltCmd(args ...string) *exec.Cmd {
	cmd := r.user.DoltCmd(args...)
	cmd.Dir = r.Dir
	return cmd
}

// CreateRemote adds a remote named |name| with the given |url| to this
// database. Doltgres has no `remote` CLI command, so this is done by briefly
// running a server and calling the dolt_remote() stored function.
func (r Repo) CreateRemote(name, url string) error {
	return r.rs.initDatabase(r.Name, func(db *sql.DB) error {
		_, err := db.Exec(`SELECT dolt_remote('add', $1, $2)`, name, url)
		return err
	})
}

type SqlServer struct {
	Name        string
	Done        chan struct{}
	Cmd         *exec.Cmd
	CmdWaitErr  error
	Port        int
	DebugPort   int
	Output      *bytes.Buffer
	DBName      string
	RecreateCmd func(args ...string) *exec.Cmd

	// Where to write server log output for display. If nil,
	// defaults to os.Stdout.
	LogWriter io.Writer

	// If non-nil, called with each complete line of server output.
	OutputVisitor func(string)
}

type SqlServerOpt func(s *SqlServer)

func WithArgs(args ...string) SqlServerOpt {
	return func(s *SqlServer) {
		s.Cmd.Args = append(s.Cmd.Args, args...)
	}
}

func WithName(name string) SqlServerOpt {
	return func(s *SqlServer) {
		s.Name = name
	}
}

func WithEnvs(envs ...string) SqlServerOpt {
	return func(s *SqlServer) {
		s.Cmd.Env = append(s.Cmd.Env, envs...)
	}
}

func WithPort(port int) SqlServerOpt {
	return func(s *SqlServer) {
		s.Port = port
	}
}

func WithOutputVisitor(f func(string)) SqlServerOpt {
	return func(s *SqlServer) {
		s.OutputVisitor = f
	}
}

func WithLogWriter(w io.Writer) SqlServerOpt {
	return func(s *SqlServer) {
		s.LogWriter = w
	}
}

func WithDebugPort(port int) SqlServerOpt {
	return func(s *SqlServer) {
		s.DebugPort = port
	}
}

type DoltCmdable interface {
	DoltCmd(args ...string) *exec.Cmd
}

type DoltDebuggable interface {
	DoltDebug(debuggerPort int, args ...string) *exec.Cmd
}

// StartSqlServer starts a doltgres server. Unlike Dolt, doltgres has no
// `sql-server` subcommand; running the binary itself starts the server.
func StartSqlServer(dc DoltCmdable, opts ...SqlServerOpt) (*SqlServer, error) {
	cmd := dc.DoltCmd()
	return runSqlServerCommand(dc, opts, cmd)
}

func DebugSqlServer(dc DoltCmdable, debuggerPort int, opts ...SqlServerOpt) (*SqlServer, error) {
	ddb, ok := dc.(DoltDebuggable)
	if !ok {
		return nil, fmt.Errorf("%T does not implement DoltDebuggable", dc)
	}

	cmd := ddb.DoltDebug(debuggerPort)
	return runSqlServerCommand(dc, append(opts, WithDebugPort(debuggerPort)), cmd)
}

func runSqlServerCommand(dc DoltCmdable, opts []SqlServerOpt, cmd *exec.Cmd) (*SqlServer, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = cmd.Stdout
	output := new(bytes.Buffer)
	done := make(chan struct{})

	server := &SqlServer{
		Done:   done,
		Cmd:    cmd,
		Port:   5432,
		Output: output,
	}
	for _, o := range opts {
		o(server)
	}

	go func() {
		defer func() {
			server.CmdWaitErr = server.Cmd.Wait()
			close(done)
		}()
		logw := server.LogWriter
		if logw == nil {
			logw = os.Stdout
		}
		multiCopyWithNamePrefix(logw, output, stdout, server.Name, server.OutputVisitor)
	}()

	server.RecreateCmd = func(args ...string) *exec.Cmd {
		if server.DebugPort > 0 {
			ddb, ok := dc.(DoltDebuggable)
			if !ok {
				panic(fmt.Sprintf("%T does not implement DoltDebuggable", dc))
			}
			return ddb.DoltDebug(server.DebugPort, args...)
		} else {
			return dc.DoltCmd(args...)
		}
	}

	err = server.Cmd.Start()
	if err != nil {
		return nil, err
	}
	return server, nil
}

func (s *SqlServer) ErrorStop() error {
	<-s.Done
	return s.CmdWaitErr
}

func multiCopyWithNamePrefix(stdout, captured io.Writer, in io.Reader, name string, visitor func(string)) {
	reader := bufio.NewReader(in)
	multiOut := io.MultiWriter(stdout, captured)
	var lineBuf []byte
	wantsPrefix := true
	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			return
		}
		if wantsPrefix && name != "" {
			stdout.Write([]byte("["))
			stdout.Write([]byte(name))
			stdout.Write([]byte("] "))
		}
		multiOut.Write(line)
		if isPrefix {
			if visitor != nil {
				lineBuf = append(lineBuf, line...)
			}
			wantsPrefix = false
		} else {
			multiOut.Write([]byte("\n"))
			if visitor != nil {
				if lineBuf != nil {
					lineBuf = append(lineBuf, line...)
					visitor(string(lineBuf))
					lineBuf = nil
				} else {
					visitor(string(line))
				}
			}
			wantsPrefix = true
		}
	}
}

func (s *SqlServer) Restart(newargs *[]string, newenvs *[]string) error {
	err := s.GracefulStop()
	if err != nil {
		return err
	}
	args := s.Cmd.Args[1:]
	if newargs != nil {
		args = *newargs
	}
	s.Cmd = s.RecreateCmd(args...)
	if newenvs != nil {
		s.Cmd.Env = append(s.Cmd.Env, (*newenvs)...)
	}
	stdout, err := s.Cmd.StdoutPipe()
	if err != nil {
		return err
	}
	s.CmdWaitErr = nil
	s.Cmd.Stderr = s.Cmd.Stdout
	s.Done = make(chan struct{})
	go func() {
		defer func() {
			s.CmdWaitErr = s.Cmd.Wait()
			close(s.Done)
		}()
		logw := s.LogWriter
		if logw == nil {
			logw = os.Stdout
		}
		multiCopyWithNamePrefix(logw, s.Output, stdout, s.Name, s.OutputVisitor)
	}()
	return s.Cmd.Start()
}

func (s *SqlServer) DB(c Connection) (*sql.DB, error) {
	connector, err := s.Connector(c)
	if err != nil {
		return nil, err
	}
	return OpenDB(connector)
}

// Connector returns a database/sql driver.Connector for a connection to this
// server using the pgx stdlib driver.
func (s *SqlServer) Connector(c Connection) (driver.Connector, error) {
	pass, err := c.Password()
	if err != nil {
		return nil, err
	}
	user := c.User
	if user == "" {
		user = "postgres"
	}
	dsn := GetDSN(user, pass, s.DBName, "127.0.0.1", s.Port, c.DriverParams)
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	return stdlib.GetConnector(*cfg), nil
}

// GetDSN builds a postgres connection URL.
func GetDSN(user, password, name, host string, port int, driverParams map[string]string) string {
	params := make(url.Values)
	params.Set("sslmode", "prefer")
	for k, v := range driverParams {
		params.Set(k, v)
	}
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     fmt.Sprintf("%s:%d", host, port),
		Path:     "/" + name,
		RawQuery: params.Encode(),
	}
	return u.String()
}

func OpenDB(connector driver.Connector) (*sql.DB, error) {
	db := sql.OpenDB(connector)
	var err error
	for i := 0; i < ConnectAttempts; i++ {
		err = db.Ping()
		if err == nil {
			return db, nil
		}
		time.Sleep(RetrySleepDuration)
	}
	if err != nil {
		return nil, err
	}
	return db, nil
}

func ConnectDB(user, password, name, host string, port int, params map[string]string) (*sql.DB, error) {
	dsn := GetDSN(user, password, name, host, port, params)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	for i := 0; i < ConnectAttempts; i++ {
		err = db.Ping()
		if err == nil {
			return db, nil
		}
		time.Sleep(RetrySleepDuration)
	}
	if err != nil {
		return nil, err
	}
	return db, nil
}

// WithConnectRetriesDisabled is retained for API-compatibility with the Dolt
// test driver. The mysql driver performed automatic connection retries that
// some tests needed to circumvent; the pgx stdlib driver does not perform the
// same retries, so this currently returns the context unchanged.
//
// TODO: If we add tests that require precise control over connection-retry
// behavior against doltgres (e.g. max-connection tests), implement an
// equivalent hook here.
func WithConnectRetriesDisabled(ctx context.Context) context.Context {
	return ctx
}
