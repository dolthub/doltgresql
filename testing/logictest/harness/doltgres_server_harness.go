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
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/dolthub/sqllogictest/go/logictest"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	dsn               = "postgresql://postgres:password@localhost:5432/sqllogictest?sslmode=disable"
	doltgresNoDbDsn   = "postgresql://doltgres:password@127.0.0.1:5432/?sslmode=disable"
	doltgresWithDbDsn = "postgresql://doltgres:password@0.0.0.0:5432/sqllogictest?sslmode=disable"
	doltgresDBDir     = "doltgresDatabases"
	serverLogFile     = "server.log"
	harnessLogFile    = "harness.log"
)

var _ logictest.Harness = &DoltgresHarness{}

// DoltgresHarness is sqllogictest harness for doltgres databases.
type DoltgresHarness struct {
	db               *sql.DB
	doltgresExec     string
	server           *DoltgresServer
	serverDir        string
	timeout          int64 // in seconds
	harnessLog       *os.File
	configFile       string
	stashedLogOutput io.Writer
}

// NewDoltgresHarness returns a new Doltgres test harness for the data source name given.
// It starts doltgres server and handles every connection to it.
func NewDoltgresHarness(doltgresExec string, t int64) *DoltgresHarness {
	serverDir := prepareSqlLogicTestDBAndGetServerDir(context.Background(), doltgresExec)
	hl, err := os.OpenFile(harnessLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	stashLogOutput := log.Writer()
	log.SetOutput(hl)
	logMsg("creating a new DoltgresHarness")

	return &DoltgresHarness{
		doltgresExec:     doltgresExec,
		serverDir:        serverDir,
		timeout:          t,
		harnessLog:       hl,
		stashedLogOutput: stashLogOutput,
	}
}

func (h *DoltgresHarness) EngineStr() string {
	return "postgresql"
}

func (h *DoltgresHarness) Init() error {
	config, err := h.createTempConfigFile()
	if err != nil {
		return err
	}

	h.configFile = config

	h.startNewDoltgresServer(context.Background(), logictest.GetCurrentFileName())
	db, err := sql.Open("pgx", doltgresNoDbDsn)
	if err != nil {
		logErr(err, "opening connection to pgx")
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}

	// create database if not exists
	//_, err = db.ExecContext(context.Background(), "\\c sqllogictest;")
	_, err = db.ExecContext(context.Background(), "CREATE DATABASE sqllogictest")
	if err != nil {
		logErr(err, "creating database")
		return err
	}

	err = db.Close()
	if err != nil {
		logErr(err, "closing database connection")
		return err
	}

	db, err = sql.Open("pgx", doltgresWithDbDsn)
	if err != nil {
		logErr(err, "opening connection to pgx")
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}

	h.db = db

	if err := h.dropAllTables(); err != nil {
		return err
	}

	return h.dropAllViews()
}

func (s *DoltgresHarness) Close() error {
	s.ClearServer()
	s.harnessLog.Close()
	log.SetOutput(s.stashedLogOutput)
	return os.RemoveAll(s.serverDir)
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

	schema, columns, err := columns(rows)
	if err != nil {
		return "", nil, err
	}

	for rows.Next() {
		err := rows.Scan(columns...)
		if err != nil {
			return "", nil, err
		}

		for _, col := range columns {
			results = append(results, stringVal(col))
		}
	}

	if rows.Err() != nil {
		return "", nil, rows.Err()
	}

	return schema, results, nil
}

func (h *DoltgresHarness) GetTimeout() int64 {
	return h.timeout
}

func (h *DoltgresHarness) dropAllTables() error {
	var rows *sql.Rows
	var err error
	rows, err = h.db.QueryContext(context.Background(), "SELECT table_name FROM information_schema.tables WHERE table_schema = 'sqllogictest' AND table_type = 'BASE TABLE';")
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return err
	}

	_, columns, err := columns(rows)
	if err != nil {
		return err
	}

	var tableNames []string
	for rows.Next() {
		err := rows.Scan(columns...)
		if err != nil {
			return err
		}

		tableName := columns[0].(*sql.NullString)
		tableNames = append(tableNames, tableName.String)
	}

	if len(tableNames) > 0 {
		dropTables := "drop table if exists " + strings.Join(tableNames, ",")
		_, err = h.db.Exec(dropTables)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *DoltgresHarness) dropAllViews() error {
	rows, err := h.db.QueryContext(context.Background(), "select table_name from INFORMATION_SCHEMA.views")
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return err
	}

	_, columns, err := columns(rows)
	if err != nil {
		return err
	}

	var viewNames []string
	for rows.Next() {
		err := rows.Scan(columns...)
		if err != nil {
			return err
		}

		viewName := columns[0].(*sql.NullString)
		viewNames = append(viewNames, viewName.String)
	}

	if len(viewNames) > 0 {
		dropView := "drop view if exists " + strings.Join(viewNames, ",")
		_, err = h.db.Exec(dropView)
		if err != nil {
			return err
		}
	}

	return nil
}

var configTemplate = `log_level: info

behavior:
  read_only: false

listener:
  host: %s
  port: %d
  read_timeout_millis: 28800000
  write_timeout_millis: 28800000
`

func (h *DoltgresHarness) createTempConfigFile() (string, error) {
	content := fmt.Sprintf(configTemplate, "127.0.0.1", 5432)
	file, err := os.CreateTemp("", "doltgres_config.yaml")
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = io.Copy(file, strings.NewReader(content))
	return file.Name(), err
}

// startNewDoltgresServer stops the existing server if exists.
// It starts a new server and update the |server| of the harness.
func (h *DoltgresHarness) startNewDoltgresServer(ctx context.Context, newTestFile string) {
	h.ClearServer()

	withKeyCtx, cancel := context.WithCancel(ctx)
	gServer, serverCtx := errgroup.WithContext(withKeyCtx)

	server := exec.CommandContext(serverCtx, h.doltgresExec, "--data-dir=.", fmt.Sprintf("--config=%s", h.configFile))
	server.Dir = h.serverDir

	// open log file for server output
	l, err := os.OpenFile(serverLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logErr(err, fmt.Sprintf("opening %s file", serverLogFile))
	}
	server.Stdout = l
	server.Stderr = l

	// handle user interrupt
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

	doltgresServer := &DoltgresServer{
		dir:       h.serverDir,
		quit:      quit,
		wg:        &wg,
		gServer:   gServer,
		server:    server,
		testFile:  newTestFile,
		serverLog: l,
	}

	h.server = doltgresServer
	h.server.Start()
}

// ClearServer closes the connection to the server and the server if either exists.
func (h *DoltgresHarness) ClearServer() {
	if h.db != nil {
		err := h.db.Close()
		if err != nil {
			logErr(err, "closing connection")
		}
		h.db = nil
	}
	// close
	if h.server != nil {
		h.server.Close()
		h.server = nil
	}
}

func prepareSqlLogicTestDBAndGetServerDir(ctx context.Context, doltgresExec string) string {
	cwd, err := os.Getwd()
	if err != nil {
		logErr(err, "getting cwd")
	}

	serverDir := filepath.Join(cwd, doltgresDBDir)
	// remove this dir to make sure it doesn't exist from previous run
	err = os.RemoveAll(serverDir)
	if err != nil {
		logErr(err, "running `RemoveAll`")
	}

	// todo: this no longer creates the db sqllogictest
	// this creates db named 'sqllogictest'
	logicTestDbDir := filepath.Join(serverDir, "sqllogictest")
	err = os.MkdirAll(logicTestDbDir, os.ModePerm)
	if err != nil {
		logErr(err, "running `MkdirAll`")
	}

	//testInit := exec.CommandContext(ctx, doltgresExec, "init")
	//testInit.Dir = logicTestDbDir
	//err = testInit.Run()
	//if err != nil {
	//	logErr(err, "running `doltgres init`")
	//}

	return serverDir
}

type DoltgresServer struct {
	dir       string
	quit      chan os.Signal
	wg        *sync.WaitGroup
	gServer   *errgroup.Group
	server    *exec.Cmd
	testFile  string
	serverLog *os.File
}

func (s *DoltgresServer) Start() {
	logMsg(fmt.Sprintf("starting doltgres server for: %s", s.testFile))
	var err error
	// launch the dolt server
	s.gServer.Go(func() error {
		err = s.server.Run()
		return err
	})

	// sleep to allow the server to start
	time.Sleep(3 * time.Second)
	if err != nil {
		logErr(err, "from server.Start()")
	}
}

func (s *DoltgresServer) Stop() {
	select {
	case <-s.quit:
		// closed
		return
	default:
	}

	// send signal to dolt server
	s.quit <- syscall.SIGTERM
	//defer s.isRunning.Store(false)
	err := s.gServer.Wait()
	if err != nil {
		// we expect a kill error
		// we only exit in error
		// if this is not the error
		if err.Error() == "signal: killed" {
			logMsg("doltgres server is stopped successfully from SIGTERM")
		} else {
			logErr(err, "from server.Stop()")
		}
	}

	close(s.quit)
	s.wg.Wait()
}

func (s *DoltgresServer) Close() {
	s.Stop()
	err := s.serverLog.Close()
	if err != nil {
		logErr(err, fmt.Sprintf("closing server.log file for server for %s", s.testFile))
	}
}

func logErr(err error, cause string) {
	log.Printf("ERROR: %s received from %s", err.Error(), cause)
}

func logMsg(msg string) {
	log.Println(msg)
}
