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
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/integrity"
)

const (
	modeReport = "report"
	modeRepair = "repair"
)

// admin is an offline administration tool for doltgres databases. It identifies and repairs a form of
// corruption created by release before 0.56.3 where rows containing adaptive value columns (TEXT, JSON, etc.)
// may not be correctly pushed or retained during a garbage collection, resulting in data loss.
//
// Usage:
//
//	admin report [-dir <path>] [-out <report.html>] [-verbose]
//	admin repair [-dir <path>] [-out <report.html>] [-verbose]
//
// Report mode scans the head of every branch of every database found in -dir and writes an HTML report
// summarizing the corruption found. Repair mode does the same, then rewrites every corrupted tree node
// in every commit reachable from any branch or tag, updating refs to the repaired history, and appends
// a post-repair verification scan to the report. Repair preserves the hashes of unaffected commits.
//
// Repair mode must be run while the database is offline (no server running against it).
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 1 || (args[0] != modeReport && args[0] != modeRepair) {
		return fmt.Errorf("usage: admin <report|repair> [-dir <path>] [-out <report.html>] [-verbose]")
	}
	mode := args[0]

	flags := flag.NewFlagSet(mode, flag.ExitOnError)
	dir := flags.String("dir", ".", "database directory, or a data directory containing databases")
	out := flags.String("out", "adaptive-corruption-report.html", "path of the HTML report to write")
	verbose := flags.Bool("verbose", false, "log progress to stderr")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	ctx := context.Background()
	dbs, cleanup, err := openDatabases(ctx, *dir)
	if err != nil {
		return err
	}
	defer cleanup()

	reports := make([]*databaseReport, 0, len(dbs))
	for _, db := range dbs {
		report, err := processDatabase(ctx, db, mode, *verbose)
		if err != nil {
			return fmt.Errorf("database %s: %w", db.name, err)
		}
		reports = append(reports, report)
	}

	f, err := os.Create(*out)
	if err != nil {
		return err
	}
	defer f.Close()
	if err = writeHTMLReport(f, mode, reports); err != nil {
		return err
	}

	printSummary(os.Stdout, mode, *out, reports)
	return nil
}

// processDatabase scans (and in repair mode, repairs) a single database.
func processDatabase(ctx context.Context, db *database, mode string, verbose bool) (*databaseReport, error) {
	sctx := db.sctx
	sc := integrity.NewScanner(db.cs)

	report := &databaseReport{Name: db.name}

	if verbose {
		log.Printf("scanning database %s", db.name)
	}
	branches, err := scanBranchHeads(sctx, db, sc, verbose)
	if err != nil {
		return nil, err
	}
	report.Branches = branches

	if mode == modeRepair {
		if verbose {
			log.Printf("repairing database %s", db.name)
		}
		rep := newRepairer(db, sc, verbose)
		summary, err := rep.repairDatabase(sctx)
		if err != nil {
			return nil, err
		}
		report.Repair = summary

		if verbose {
			log.Printf("verifying database %s post-repair", db.name)
		}
		postRepair, err := scanBranchHeads(sctx, db, sc, verbose)
		if err != nil {
			return nil, err
		}
		report.PostRepair = postRepair
	}

	if verbose {
		log.Printf("database %s: %d subtree scans satisfied from cache", db.name, sc.CacheHits)
	}
	return report, nil
}

// scanBranchHeads scans every table at the head of every branch of the database.
func scanBranchHeads(sctx *sql.Context, db *database, sc *integrity.Scanner, verbose bool) ([]*branchReport, error) {
	branches, err := db.ddb.GetBranches(sctx)
	if err != nil {
		return nil, err
	}

	var reports []*branchReport
	for _, branchRef := range branches {
		head, err := db.ddb.ResolveCommitRef(sctx, branchRef)
		if err != nil {
			return nil, err
		}
		root, err := head.GetRootValue(sctx)
		if err != nil {
			return nil, err
		}
		tables, err := integrity.TablesForRoot(sctx, root, db.ns)
		if err != nil {
			return nil, err
		}

		br := &branchReport{Branch: branchRef.GetPath()}
		for _, ti := range tables {
			tr := &tableReport{
				Schema:            ti.Name.Schema,
				Table:             ti.Name.Name,
				AdaptiveValueCols: ti.AdaptiveValueCols,
				AdaptiveKeyCols:   ti.AdaptiveKeyCols,
				ValColsImpacted:   ti.ValColsImpacted(),
				KeyColsImpacted:   ti.KeyColsImpacted(),
			}
			if ti.ValColsImpacted() || ti.KeyColsImpacted() {
				stats, err := sc.ScanTable(sctx, ti)
				if err != nil {
					tr.Err = err.Error()
				} else {
					tr.Stats = stats
					if verbose {
						log.Printf("branch %s table %s: %d/%d rows corrupt, %d/%d adaptive values corrupt",
							br.Branch, ti.Name.String(), stats.CorruptRows, stats.Rows, stats.CorruptValues, stats.AdaptiveValues)
					}
				}
			}
			br.Tables = append(br.Tables, tr)
		}
		reports = append(reports, br)
	}
	return reports, nil
}

// printSummary writes a short text summary of the results to |w|.
func printSummary(w *os.File, mode, outPath string, reports []*databaseReport) {
	for _, dbr := range reports {
		var rows, corruptRows, values, corruptValues, missing uint64
		for _, br := range dbr.Branches {
			for _, tr := range br.Tables {
				if tr.Stats == nil {
					continue
				}
				rows += tr.Stats.Rows
				corruptRows += tr.Stats.CorruptRows
				values += tr.Stats.AdaptiveValues + tr.Stats.KeyAdaptiveValues
				corruptValues += tr.Stats.CorruptValues + tr.Stats.KeyCorruptValues
				missing += tr.Stats.MissingChunks
			}
		}
		fmt.Fprintf(w, "database %s: %d corrupt rows (of %d scanned), %d corrupt adaptive values (of %d scanned) across all branch heads\n",
			dbr.Name, corruptRows, rows, corruptValues, values)
		if missing > 0 {
			fmt.Fprintf(w, "database %s: WARNING: %d out-of-band chunks are already missing from the chunk store; the values referencing them are unrecoverable\n",
				dbr.Name, missing)
		}
		if dbr.Repair != nil {
			fmt.Fprintf(w, "database %s: repaired %d leaf chunks, rewrote %d of %d commits, updated %d branches, %d tags, %d working sets\n",
				dbr.Name, dbr.Repair.LeafChunksRewritten, dbr.Repair.CommitsRewritten, dbr.Repair.CommitsExamined,
				dbr.Repair.BranchesUpdated, dbr.Repair.TagsUpdated, dbr.Repair.WorkingSetsFixed)
		}
	}
	fmt.Fprintf(w, "full report written to %s\n", outPath)
}
