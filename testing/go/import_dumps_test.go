// Copyright 2025 Dolthub, Inc.
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
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/testing/dumps"
)

// TestImportingDumps are regression tests against dumps taken from various sources.
func TestImportingDumps(t *testing.T) {
	RunImportTests(t, []ImportTest{
		{
			Name: "Scrubbed-1",
			SetUpScript: []string{
				"CREATE USER behfjgnf WITH SUPERUSER PASSWORD 'password';",
			},
			SkipQueries: []string{"CREATE UNIQUE INDEX dawkmezfehakyikllr"},
			SQLFilename: "scrubbed-1.sql",
		},
	})
}

// TriggerImportBreakpoint exists so that a breakpoint may be set within the function, on the unused Sprintf. This
// function is called whenever a query matches one of the breakpoint queries defined in the import test. This enables us
// to simulate some kind of breakpoint functionality on import queries, which isn't normally possible.
func TriggerImportBreakpoint(breakpointQuery string) {
	// It doesn't actually matter what this function is. It's just here so we can set a breakpoint on something.
	_ = fmt.Sprintf("__%s", breakpointQuery)
}

// ImportTest is a test for importing SQL dumps.
type ImportTest struct {
	Name        string
	SetUpScript []string
	Focus       bool
	Skip        bool
	SQLFilename string
	// Breakpoints allow for triggering breakpoints when any matching queries are given. A breakpoint must be set within
	// TriggerImportBreakpoint for this to work.
	Breakpoints []string
	// SkipQueries
	SkipQueries []string
}

// RunImportTests runs the given ImportTest scripts.
func RunImportTests(t *testing.T, scripts []ImportTest) {
	if _, ok := os.LookupEnv("GITHUB_ACTION"); ok {
		if _, ok = os.LookupEnv("GITHUB_ACTION_IMPORT_DUMPS"); !ok {
			t.Skip("These tests are run in their own dedicated action")
		}
	}
	var psqlCommand string
	switch runtime.GOOS {
	case "windows":
		psqlCommand = "psql.exe"
	default:
		psqlCommand = "psql"
	}
	// Check if PSQL runs directly
	var outBuffer bytes.Buffer
	cmd := exec.Command(psqlCommand, "--version")
	cmd.Stdout = &outBuffer
	if !assert.NoError(t, cmd.Run()) || !strings.Contains(outBuffer.String(), "PostgreSQL") {
		// We could not run PSQL and get the version, so it must not be in the path.
		// We'll check if pg_config is in the path and reference the binary directly.
		outBuffer.Reset()
		cmd = exec.Command("pg_config", "--bindir")
		cmd.Stdout = &outBuffer
		if !assert.NoError(t, cmd.Run()) {
			require.Fail(t, "Postgres is not installed, cannot run tests")
		}
		psqlCommand = filepath.Join(strings.TrimSpace(outBuffer.String()), psqlCommand)
		// pg_config is in the path, so we'll try and run PSQL by directly referencing the binary
		outBuffer.Reset()
		cmd = exec.Command(psqlCommand, "--version")
		cmd.Stdout = &outBuffer
		if !assert.NoError(t, cmd.Run()) || !strings.Contains(outBuffer.String(), "PostgreSQL") {
			t.Fatalf("PSQL cannot be found at: `%s`", psqlCommand)
		}
	}
	// Grab the folder with the files to import
	_, currentFileLocation, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Unable to find the folder where the files to import are located")
	}
	dumpsFolder := filepath.Clean(filepath.Join(filepath.Dir(currentFileLocation), "../dumps/"))
	// Set whether we're checking Focus-only scripts or not
	useFocus := false
	for _, script := range scripts {
		if script.Focus {
			// If this is running in GitHub Actions, then we'll panic, because someone forgot to disable it before committing
			if _, ok := os.LookupEnv("GITHUB_ACTION"); ok {
				panic(fmt.Sprintf("The script `%s` has Focus set to `true`. GitHub Actions requires that "+
					"all tests are run, which Focus circumvents, leading to this error. Please disable Focus on "+
					"all tests.", script.Name))
			}
			useFocus = true
			break
		}
	}
	for _, script := range scripts {
		if useFocus != script.Focus {
			continue
		}
		RunImportTest(t, script, psqlCommand, dumpsFolder)
	}
}

// RunImportTest runs the given ImportTest script.
func RunImportTest(t *testing.T, script ImportTest, psqlCommand string, dumpsFolder string) {
	// TODO: handle other dump types, such as those that require pg_restore
	t.Run(script.Name, func(t *testing.T) {
		// Mark this test as skipped if we have it set
		if script.Skip {
			t.Skip()
		}
		// Create the in-memory server that we'll test against
		port, err := sql.GetEmptyPort()
		require.NoError(t, err)
		ctx, conn, controller := CreateServerWithPort(t, "postgres", port)
		func() {
			defer conn.Close(ctx)
			for _, query := range script.SetUpScript {
				_, err = conn.Exec(ctx, query)
				require.NoError(t, err)
			}
		}()
		defer func() {
			controller.Stop()
			err := controller.WaitForStop()
			require.NoError(t, err)
		}()
		// Create the message interceptor
		var qeChan chan dumps.ImportQueryError
		port, qeChan = dumps.InterceptImportMessages(t, dumps.InterceptArgs{
			DoltgresPort:      port,
			SkippedQueries:    script.SkipQueries,
			BreakpointQueries: script.Breakpoints,
			TriggerBreakpoint: TriggerImportBreakpoint,
		})
		defer close(qeChan)
		var allErrors []dumps.ImportQueryError
		go func() {
			for chanErr := range qeChan {
				allErrors = append(allErrors, chanErr)
			}
		}()
		// Run the import
		var outBuffer bytes.Buffer
		var errBuffer bytes.Buffer
		cmd := exec.Command(psqlCommand, fmt.Sprintf("postgresql://postgres:password@localhost:%d/postgres?sslmode=disable", port))
		cmd.Stdout = &outBuffer
		cmd.Stderr = &errBuffer
		targetFile, err := os.Open(filepath.Join(dumpsFolder, "sql", script.SQLFilename))
		require.NoError(t, err)
		cmd.Stdin = targetFile
		require.NoError(t, cmd.Run())
		if len(allErrors) > 0 {
			t.Logf("COUNT: %d", len(allErrors))
			// If we have more than some threshold, then we'll only show the first few for ease of consumption
			for i := 0; i < len(allErrors) && i < 10; i++ {
				t.Logf("QUERY: %s\nERROR: %s", allErrors[i].Query, allErrors[i].Error)
			}
			t.FailNow()
		}
	})
}
