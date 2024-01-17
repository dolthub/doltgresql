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
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dolthub/sqllogictest/go/logictest"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	dsn               = "postgresql://postgres:password@localhost:5432/sqllogictest?sslmode=disable"
	doltgresResultDir = "doltgresBin"
	serverLog         = "doltgres_server.log"
	errorLog          = "error.log"
)

var _ logictest.Harness = &DoltgresHarness{}

// DoltgresHarness is sqllogictest harness for doltgres databases.
type DoltgresHarness struct {
	db           *sql.DB
	doltgresExec string
	server       *DoltgresServer
	serverDir    string
}

// NewDoltgresHarness returns a new Doltgres test harness for the data source name given.
// It starts doltgres server and handles every connection to it.
func NewDoltgresHarness(doltgresExec string) *DoltgresHarness {
	serverDir := prepareSqlLogicTestDBAndGetServerDir(context.Background(), doltgresExec)
	errLog, err := os.OpenFile(errorLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(errLog)

	return &DoltgresHarness{
		doltgresExec: doltgresExec,
		serverDir:    serverDir,
	}
}

func (h *DoltgresHarness) EngineStr() string {
	return "postgresql"
}

func (h *DoltgresHarness) Init() error {
	h.startNewDoltgresServer(context.Background())
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}
	h.db = db

	if err := h.dropAllTables(); err != nil {
		return err
	}

	return h.dropAllViews()
}

func (s *DoltgresHarness) Close() error {
	s.server.Close()
	s.server = nil
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

// startNewDoltgresServer stops the existing server if exists.
// It starts a new server and update the |server| of the harness.
func (h *DoltgresHarness) startNewDoltgresServer(ctx context.Context) {
	if h.server != nil {
		h.server.Stop()
		h.server = nil
	}

	withKeyCtx, cancel := context.WithCancel(ctx)
	gServer, serverCtx := errgroup.WithContext(withKeyCtx)

	server := exec.CommandContext(serverCtx, h.doltgresExec, "--data-dir=.")
	server.Dir = h.serverDir

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
		dir:     h.serverDir,
		quit:    quit,
		wg:      &wg,
		gServer: gServer,
		server:  server,
	}

	h.server = doltgresServer

	err := h.server.Start()
	if err != nil {
		log.Printf("got error from server.Start: %s", err.Error())
	}
}

func prepareSqlLogicTestDBAndGetServerDir(ctx context.Context, doltgresExec string) string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("got error from cwd: %s", err.Error())
	}

	serverDir := filepath.Join(cwd, doltgresResultDir)
	// remove this dir to make sure it doesn't exist from previous run
	err = os.RemoveAll(serverDir)
	if err != nil {
		log.Printf("got error from RemoveAll: %s", err.Error())
	}

	// this creates db named 'sqllogictest'
	logicTestDbDir := filepath.Join(serverDir, "sqllogictest")
	err = os.MkdirAll(logicTestDbDir, os.ModePerm)
	if err != nil {
		log.Printf("got error from MkdirAll: %s", err.Error())
	}

	testInit := exec.CommandContext(ctx, doltgresExec, "init")
	testInit.Dir = logicTestDbDir
	err = testInit.Run()
	if err != nil {
		log.Printf("got error from running doltgres init: %s", err.Error())
	}

	return serverDir
}

type DoltgresServer struct {
	dir     string
	quit    chan os.Signal
	wg      *sync.WaitGroup
	gServer *errgroup.Group
	server  *exec.Cmd
}

func (s *DoltgresServer) Start() error {
	// launch the dolt server
	s.gServer.Go(func() error {
		return s.server.Run()
	})

	// sleep to allow the server to start
	time.Sleep(3 * time.Second)
	return nil
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
		if err.Error() != "signal: killed" {
			log.Printf("server is stopped from potential panic: %s", err.Error())
		} else {
			log.Printf("server is stopped from SIGTERM sent: %s", err.Error())
		}
	}

	close(s.quit)
	s.wg.Wait()
}

func (s *DoltgresServer) Close() {
	s.Stop()
}
