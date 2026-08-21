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

import "time"

// TableStats records the result of scanning a single table's row data for adaptive-encoded
// values whose out-of-band chunks are missing from the chunk store.
type TableStats struct {
	Table string
	// AdaptiveColumns are the names of the non-key columns in this table with adaptive encodings.
	AdaptiveColumns []string
	// AdaptiveKeyColumns are the names of the primary key columns with adaptive encodings. Missing
	// values in key columns cannot be repaired by NULLing them out, so they are tracked separately.
	AdaptiveKeyColumns []string
	// RowsScanned is the total number of rows examined.
	RowsScanned uint64
	// RowsWithMissing is the number of rows containing at least one missing out-of-band value.
	RowsWithMissing uint64
	// AdaptiveValues is the total number of non-NULL adaptive values examined.
	AdaptiveValues uint64
	// OutOfBandValues is the number of adaptive values stored out-of-band (as chunk addresses).
	OutOfBandValues uint64
	// MissingValues is the number of out-of-band values whose chunks are absent from the chunk store.
	MissingValues uint64
	// MissingKeyValues is the number of missing out-of-band values found in primary key columns.
	// These are reported but never repaired.
	MissingKeyValues uint64
	// RepairedValues is the number of values NULLed out (repair mode only).
	RepairedValues uint64
	// MissingByColumn breaks down MissingValues (and MissingKeyValues) by column name.
	MissingByColumn map[string]uint64
}

// Impacted returns whether the scan found any missing values in this table.
func (ts *TableStats) Impacted() bool {
	return ts.MissingValues > 0 || ts.MissingKeyValues > 0
}

// PctRowsMissing returns the percentage of scanned rows that contain at least one missing value.
func (ts *TableStats) PctRowsMissing() float64 {
	return pct(ts.RowsWithMissing, ts.RowsScanned)
}

// PctValuesMissing returns the percentage of examined adaptive values that are missing.
func (ts *TableStats) PctValuesMissing() float64 {
	return pct(ts.MissingValues+ts.MissingKeyValues, ts.AdaptiveValues)
}

// add accumulates |other| into |ts|. Used to aggregate stats across the commits of a branch.
func (ts *TableStats) add(other *TableStats) {
	ts.RowsScanned += other.RowsScanned
	ts.RowsWithMissing += other.RowsWithMissing
	ts.AdaptiveValues += other.AdaptiveValues
	ts.OutOfBandValues += other.OutOfBandValues
	ts.MissingValues += other.MissingValues
	ts.MissingKeyValues += other.MissingKeyValues
	ts.RepairedValues += other.RepairedValues
	for col, n := range other.MissingByColumn {
		if ts.MissingByColumn == nil {
			ts.MissingByColumn = make(map[string]uint64)
		}
		ts.MissingByColumn[col] += n
	}
}

func pct(num, denom uint64) float64 {
	if denom == 0 {
		return 0
	}
	return 100 * float64(num) / float64(denom)
}

// ImpactedTables returns only the tables with missing values; clean tables are collected for
// bookkeeping but omitted from the report.
func (br *BranchReport) ImpactedTables() []*TableStats {
	var impacted []*TableStats
	for _, ts := range br.Tables {
		if ts.Impacted() {
			impacted = append(impacted, ts)
		}
	}
	return impacted
}

// CleanTableCount returns the number of scanned tables with adaptive columns that had no
// missing values.
func (br *BranchReport) CleanTableCount() int {
	return len(br.Tables) - len(br.ImpactedTables())
}

// BranchReport aggregates the per-table scan results for a single branch.
type BranchReport struct {
	Branch string
	// CommitsScanned is the number of commits examined on this branch. In report mode this is 1
	// (the branch HEAD); in repair mode it is every commit reachable from the branch head that
	// had not already been processed via another branch.
	CommitsScanned uint64
	Tables         []*TableStats
	// SkippedTables is the number of tables on this branch with no adaptive-encoded columns.
	SkippedTables uint64
}

// DatabaseReport aggregates the per-branch scan results for a single database.
type DatabaseReport struct {
	Database string
	Branches []*BranchReport
	// Errors records non-fatal problems encountered while scanning this database.
	Errors []string
}

// Report is the top-level result of a report or repair run.
type Report struct {
	Mode        string
	DataDir     string
	GeneratedAt time.Time
	Databases   []*DatabaseReport
}
