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

package harness

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/dolthub/sqllogictest/go/logictest"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	doltgresDBDir  = "doltgresDatabases"
	serverLogFile  = "server.log"
	harnessLogFile = "harness.log"
	DefaultPort    = 5432
	defaultDbName  = "sqllogictest"
)

func noDatabaseDSN(port int) string {
	return fmt.Sprintf("postgresql://postgres:password@127.0.0.1:%d/?sslmode=disable", port)
}

func withDatabaseDSN(port int, dbName string) string {
	return fmt.Sprintf("postgresql://postgres:password@127.0.0.1:%d/%s?sslmode=disable", port, dbName)
}

var _ logictest.Harness = &DoltgresHarness{}

// DoltgresHarness is a sqllogictest harness for doltgres databases.
type DoltgresHarness struct {
	db               *sql.DB
	doltgresExec     string
	server           *DoltgresServer
	serverDir        string
	timeout          int64
	port             int
	dbName           string
	harnessLog       *os.File
	stashedLogOutput io.Writer
}

// NewDoltgresHarness returns a harness that manages its own server lifecycle. The server is
// started once at construction; Init resets the database without restarting the process.
func NewDoltgresHarness(doltgresExec string, t int64) *DoltgresHarness {
	cwd, err := os.Getwd()
	if err != nil {
		logErr(err, "getting cwd")
	}
	serverDir := filepath.Join(cwd, doltgresDBDir)
	err = os.RemoveAll(serverDir)
	if err != nil {
		logErr(err, fmt.Sprintf("running `RemoveAll` for '%s'", serverDir))
	}
	err = os.MkdirAll(serverDir, os.ModePerm)
	if err != nil {
		logErr(err, fmt.Sprintf("running `MkdirAll` for '%s'", serverDir))
	}
	hl, err := os.OpenFile(harnessLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	stashLogOutput := log.Writer()
	log.SetOutput(hl)
	logMsg("creating a new DoltgresHarness")

	h := &DoltgresHarness{
		doltgresExec:     doltgresExec,
		serverDir:        serverDir,
		timeout:          t,
		port:             DefaultPort,
		dbName:           defaultDbName,
		harnessLog:       hl,
		stashedLogOutput: stashLogOutput,
	}
	h.server = startServerProcess(doltgresExec, serverDir)
	return h
}

// NewDoltgresWorkerHarness returns a harness for use in a concurrent worker. It connects to an
// already-running server on the given port using dbName as its isolated database. The caller is
// responsible for starting and stopping the server (see StartSharedServer).
func NewDoltgresWorkerHarness(port int, dbName string, t int64) *DoltgresHarness {
	return &DoltgresHarness{
		timeout: t,
		port:    port,
		dbName:  dbName,
	}
}

// StartSharedServer starts a single doltgres server process for use by multiple concurrent
// workers. The returned *DoltgresServer must be closed when testing is complete.
func StartSharedServer(doltgresExec string) *DoltgresServer {
	cwd, err := os.Getwd()
	if err != nil {
		logErr(err, "getting cwd")
	}
	serverDir := filepath.Join(cwd, doltgresDBDir)
	err = os.RemoveAll(serverDir)
	if err != nil {
		logErr(err, fmt.Sprintf("running `RemoveAll` for '%s'", serverDir))
	}
	err = os.MkdirAll(serverDir, os.ModePerm)
	if err != nil {
		logErr(err, fmt.Sprintf("running `MkdirAll` for '%s'", serverDir))
	}
	return startServerProcess(doltgresExec, serverDir)
}

// startServerProcess starts a doltgres subprocess in serverDir and waits for it to be ready.
func startServerProcess(doltgresExec string, serverDir string) *DoltgresServer {
	withKeyCtx, cancel := context.WithCancel(context.Background())
	gServer, serverCtx := errgroup.WithContext(withKeyCtx)

	server := exec.CommandContext(serverCtx, doltgresExec, "--data-dir=.")
	server.Dir = serverDir

	l, err := os.OpenFile(serverLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logErr(err, fmt.Sprintf("opening %s file", serverLogFile))
	}
	server.Stdout = l
	server.Stderr = l

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		<-quit
		defer wg.Done()
		signal.Stop(quit)
		cancel()
	}()

	ds := &DoltgresServer{
		dir:       serverDir,
		quit:      quit,
		wg:        &wg,
		gServer:   gServer,
		server:    server,
		serverLog: l,
	}
	ds.Start()
	return ds
}

func (h *DoltgresHarness) EngineStr() string {
	return "postgresql"
}

// Init resets the harness to a clean state for the next test file.
// For worker harnesses (no server process), it keeps a persistent connection and resets the
// public schema via DROP+CREATE — avoiding database-level connection races entirely.
// For server-owning harnesses, it drops and recreates the database (safe because only one worker
// uses each database in the non-concurrent path).
func (h *DoltgresHarness) Init() error {
	if h.server == nil {
		return h.initWorker()
	}
	return h.initServer()
}

// initWorker resets a worker harness between test files.
//
// When useDropSchema is true (PostgreSQL), it wipes and recreates the public schema on the
// persistent connection — no reconnect needed, no races.
func (h *DoltgresHarness) initWorker() error {
	ctx := context.Background()

	if h.db != nil {
		h.db.Close()
		h.db = nil
	}
	noDB, err := sql.Open("pgx", noDatabaseDSN(h.port))
	if err != nil {
		return err
	}
	_, err = noDB.ExecContext(ctx, "DROP DATABASE IF EXISTS "+h.dbName)
	if err == nil {
		_, err = noDB.ExecContext(ctx, "CREATE DATABASE "+h.dbName)
	}
	noDB.Close()
	if err != nil {
		return fmt.Errorf("resetting database %s: %w", h.dbName, err)
	}
	return h.openWorkerDB(ctx)
}

// openWorkerDB establishes a connection to the worker's database, creating it if necessary.
func (h *DoltgresHarness) openWorkerDB(ctx context.Context) error {
	db, err := sql.Open("pgx", withDatabaseDSN(h.port, h.dbName))
	if err != nil {
		return err
	}
	if pingErr := db.Ping(); pingErr == nil {
		h.db = db
		return nil
	}
	db.Close()

	// Database does not exist yet; create it.
	noDB, err := sql.Open("pgx", noDatabaseDSN(h.port))
	if err != nil {
		return err
	}
	_, createErr := noDB.ExecContext(ctx, "CREATE DATABASE "+h.dbName)
	noDB.Close()
	if createErr != nil {
		logErr(createErr, "creating database "+h.dbName)
		return createErr
	}

	db, err = sql.Open("pgx", withDatabaseDSN(h.port, h.dbName))
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return err
	}
	h.db = db
	return nil
}

// initServer resets the server-owning harness by dropping and recreating the database.
// This is only used in the non-concurrent (single-worker) path.
func (h *DoltgresHarness) initServer() error {
	ctx := context.Background()

	if h.db != nil {
		h.db.Close()
		h.db = nil
	}

	noDB, openErr := sql.Open("pgx", noDatabaseDSN(h.port))
	if openErr != nil {
		logErr(openErr, "opening no-db connection")
		return openErr
	}
	_, err := noDB.ExecContext(ctx, "DROP DATABASE IF EXISTS "+h.dbName)
	if err == nil {
		_, err = noDB.ExecContext(ctx, "CREATE DATABASE "+h.dbName)
	}
	noDB.Close()
	if err != nil {
		logErr(err, "resetting database "+h.dbName)
		return err
	}

	db, err := sql.Open("pgx", withDatabaseDSN(h.port, h.dbName))
	if err != nil {
		logErr(err, "opening connection to "+h.dbName)
		return err
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return err
	}
	h.db = db
	return nil
}

func (h *DoltgresHarness) Close() error {
	if h.db != nil {
		h.db.Close()
		h.db = nil
	}
	if h.server != nil {
		h.server.Close()
		h.server = nil
	}
	if h.harnessLog != nil {
		h.harnessLog.Close()
		log.SetOutput(h.stashedLogOutput)
	}
	return os.RemoveAll(h.serverDir)
}

func (h *DoltgresHarness) ExecuteStatement(statement string) error {
	_, err := h.db.Exec(statement)
	return err
}

func (h *DoltgresHarness) ExecuteQuery(statement string) (schema string, results []string, err error) {
	rows, err := h.db.Query(statement)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return "", nil, err
	}
	return h.getSchemaAndResults(rows)
}

func (h *DoltgresHarness) ExecuteStatementContext(ctx context.Context, statement string) error {
	_, err := h.db.ExecContext(ctx, statement)
	return err
}

func (h *DoltgresHarness) ExecuteQueryContext(ctx context.Context, statement string) (schema string, results []string, err error) {
	rows, err := h.db.QueryContext(ctx, statement)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return "", nil, err
	}
	return h.getSchemaAndResults(rows)
}

func (h *DoltgresHarness) GetTimeout() int64 {
	return h.timeout
}

func (h *DoltgresHarness) getSchemaAndResults(rows *sql.Rows) (schema string, results []string, err error) {
	schema, cols, err := columns(rows)
	if err != nil {
		return "", nil, err
	}
	for rows.Next() {
		if err = rows.Scan(cols...); err != nil {
			return "", nil, err
		}
		for _, col := range cols {
			results = append(results, stringVal(col))
		}
	}
	if rows.Err() != nil {
		return "", nil, rows.Err()
	}
	return schema, results, nil
}

type DoltgresServer struct {
	dir       string
	quit      chan os.Signal
	wg        *sync.WaitGroup
	gServer   *errgroup.Group
	server    *exec.Cmd
	serverLog *os.File
}

func (s *DoltgresServer) Start() {
	logMsg("starting doltgres server")
	var err error
	s.gServer.Go(func() error {
		err = s.server.Run()
		return err
	})
	// Allow the server time to start accepting connections.
	time.Sleep(3 * time.Second)
	if err != nil {
		logErr(err, "from server.Start()")
	}
}

func (s *DoltgresServer) Stop() {
	select {
	case <-s.quit:
		return
	default:
	}
	s.quit <- syscall.SIGTERM
	err := s.gServer.Wait()
	if err != nil {
		if err.Error() == "signal: killed" {
			logMsg("doltgres server stopped successfully")
		} else {
			logErr(err, "from server.Stop()")
		}
	}
	close(s.quit)
	s.wg.Wait()
}

func (s *DoltgresServer) Close() {
	s.Stop()
	if err := s.serverLog.Close(); err != nil {
		logErr(err, "closing server.log")
	}
}

func logErr(err error, cause string) {
	log.Printf("ERROR: %s received from %s", err.Error(), cause)
}

func logMsg(msg string) {
	log.Println(msg)
}
