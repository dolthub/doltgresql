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

// admin is an offline administration tool for doltgres databases.
//
// It currently supports detecting and repairing a specific kind of corruption: adaptive-encoded
// values (TEXT, BYTEA, JSON/JSONB, and other unbounded types) whose out-of-band chunks were lost
// from the chunk store during garbage collection. Such values are stored in row tuples as a
// content address, and reading them fails when the addressed chunk no longer exists.
//
// Usage:
//
//	admin report -data-dir <dir> [-db <name>] [-branch <name>] [-out report.html]
//	admin repair -data-dir <dir> [-db <name>] [-branch <name>] [-out report.html] [-yes]
//
// Report mode scans the HEAD of every branch of every database and emits an HTML report of how
// many adaptive values are unreadable per table. Repair mode scans every commit reachable from
// every branch, NULLs out every unreadable value (rewriting commit history in place), and
// re-points branch refs and working sets at the repaired commits. The -db and -branch flags
// restrict either mode to a single database and/or branch. The server must be stopped while
// this tool runs.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dolthub/dolt/go/libraries/doltcore/dbfactory"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server"
	"github.com/dolthub/doltgresql/server/initialization"
	"github.com/dolthub/doltgresql/servercfg"
)

// loadMultiEnv loads every doltgres database found under |absDataDir| (one subdirectory per
// database, each containing a .dolt directory).
func loadMultiEnv(ctx context.Context, absDataDir string) (*env.MultiRepoEnv, error) {
	fs, err := filesys.LocalFilesysWithWorkingDir(absDataDir)
	if err != nil {
		return nil, err
	}
	dEnv := env.Load(ctx, env.GetCurrentUserHomeDir, fs, doltdb.LocalDirDoltDB, "doltgres-admin-"+server.Version)
	// Bypass the in-process singleton database cache so that closing and reopening a database
	// (which the tests rely on, and which protects against stale state generally) always yields
	// a fresh chunk store.
	dEnv.DBLoadParams = map[string]interface{}{dbfactory.DisableSingletonCacheParam: true}
	return env.MultiEnvForDirectory(ctx, fs, dEnv)
}

func progress(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 || (os.Args[1] != "report" && os.Args[1] != "repair") {
		fail("usage: admin <report|repair> -data-dir <dir> [-db <name>] [-branch <name>] [-out <file.html>] [-yes]")
	}
	mode := os.Args[1]

	flags := flag.NewFlagSet(mode, flag.ExitOnError)
	dataDir := flags.String("data-dir", ".", "directory containing the doltgres databases (one subdirectory per database)")
	dbFilter := flags.String("db", "", "only process the named database (default: all databases in the data dir)")
	branchFilter := flags.String("branch", "", "only process the named branch (default: all branches)")
	outPath := flags.String("out", "adaptive_corruption_report.html", "path to write the HTML report to")
	yes := flags.Bool("yes", false, "skip the confirmation prompt in repair mode")
	if err := flags.Parse(os.Args[2:]); err != nil {
		fail("%v", err)
	}

	absDataDir, err := filepath.Abs(*dataDir)
	if err != nil {
		fail("resolving data dir: %v", err)
	}

	if mode == "repair" && !*yes {
		fmt.Printf("Repair mode rewrites commit history in place for every impacted database under\n%s\n", absDataDir)
		fmt.Printf("Back up the data directory before proceeding. The doltgres server must be stopped.\n")
		fmt.Printf("Type 'yes' to continue: ")
		var answer string
		_, _ = fmt.Scanln(&answer)
		if strings.TrimSpace(answer) != "yes" {
			fail("aborted")
		}
	}

	ctx := context.Background()

	// Initialize doltgres globals: this swaps in the doltgres RootValue implementation and
	// registers the extended type serializers, both required to read doltgres databases.
	initialization.Initialize(nil, servercfg.DefaultServerConfig())

	mrEnv, err := loadMultiEnv(ctx, absDataDir)
	if err != nil {
		fail("loading databases from %s: %v", absDataDir, err)
	}

	report := &Report{
		Mode:        mode,
		DataDir:     absDataDir,
		GeneratedAt: time.Now(),
	}

	// Extended type deserialization requires a *sql.Context rather than a plain context.Context.
	sqlCtx := sql.NewContext(ctx)

	found := false
	err = mrEnv.Iter(func(name string, dEnv *env.DoltEnv) (bool, error) {
		if *dbFilter != "" && name != *dbFilter {
			return false, nil
		}
		found = true
		progress("database %s:", name)
		ddb := dEnv.DoltDB(ctx)

		var dbRep *DatabaseReport
		var err error
		if mode == "repair" {
			dbRep, err = repairDatabase(sqlCtx, name, ddb, *branchFilter)
		} else {
			dbRep, err = reportDatabase(sqlCtx, name, ddb, *branchFilter)
		}
		if err != nil {
			return false, fmt.Errorf("database %s: %w", name, err)
		}
		report.Databases = append(report.Databases, dbRep)
		return false, nil
	})
	if err != nil {
		fail("%v", err)
	}
	if !found {
		if *dbFilter != "" {
			fail("database %q not found under %s", *dbFilter, absDataDir)
		}
		fail("no databases found under %s", absDataDir)
	}

	out, err := os.Create(*outPath)
	if err != nil {
		fail("creating report file: %v", err)
	}
	if err = report.WriteHTML(out); err != nil {
		fail("writing report: %v", err)
	}
	if err = out.Close(); err != nil {
		fail("writing report: %v", err)
	}

	printSummary(report)
	progress("\nHTML report written to %s", *outPath)
}

// printSummary prints a text summary of the run to stdout.
func printSummary(report *Report) {
	var impacted, errors int
	for _, db := range report.Databases {
		errors += len(db.Errors)
		for _, br := range db.Branches {
			for _, ts := range br.Tables {
				if ts.Impacted() {
					impacted++
					progress("IMPACTED: db %s branch %s table %s: %d missing values (%.2f%% of rows, %.2f%% of adaptive values)%s",
						db.Database, br.Branch, ts.Table, ts.MissingValues+ts.MissingKeyValues,
						ts.PctRowsMissing(), ts.PctValuesMissing(),
						map[bool]string{true: fmt.Sprintf(", %d repaired", ts.RepairedValues), false: ""}[report.Mode == "repair"])
				}
			}
		}
	}
	if impacted == 0 {
		progress("No missing adaptive values found.")
	}
	if errors > 0 {
		progress("%d errors were encountered; see the HTML report for details.", errors)
	}
}
