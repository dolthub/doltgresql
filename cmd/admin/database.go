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
	"fmt"
	"sort"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb/durable"
	"github.com/dolthub/dolt/go/libraries/doltcore/ref"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
	"github.com/dolthub/dolt/go/store/datas"
	"github.com/dolthub/go-mysql-server/sql"
)

// fieldInfo maps tuple descriptor field indexes to column names and tags for a table schema.
type fieldInfo struct {
	valNames []string
	valTags  []uint64
	keyNames []string
}

// fieldInfoForSchema builds the field index -> column mapping for a schema. Value tuple fields
// are the non-PK, non-virtual columns in schema order, with a leading cardinality field for
// keyless tables. Key tuple fields are the PK columns in key order; keyless tables use a
// synthetic hash id key.
func fieldInfoForSchema(sch schema.Schema) fieldInfo {
	var info fieldInfo
	if schema.IsKeyless(sch) {
		info.valNames = append(info.valNames, "(cardinality)")
		info.valTags = append(info.valTags, 0)
		info.keyNames = append(info.keyNames, "(row id)")
	} else {
		for _, col := range sch.GetPKCols().GetColumns() {
			info.keyNames = append(info.keyNames, col.Name)
		}
	}
	for _, col := range sch.GetNonPKCols().GetColumns() {
		if col.Virtual {
			continue
		}
		info.valNames = append(info.valNames, col.Name)
		info.valTags = append(info.valTags, col.Tag)
	}
	return info
}

func (fi fieldInfo) valName(i int) string {
	if i < len(fi.valNames) {
		return fi.valNames[i]
	}
	return fmt.Sprintf("(value field %d)", i)
}

func (fi fieldInfo) keyName(i int) string {
	if i < len(fi.keyNames) {
		return fi.keyNames[i]
	}
	return fmt.Sprintf("(key field %d)", i)
}

// tableStatsFromScan translates a scanResult (keyed by tuple field indexes) into a TableStats
// (keyed by column names) using the table's schema.
func tableStatsFromScan(tableName string, sch schema.Schema, sr *scanResult, keyIdxs, valIdxs []int) *TableStats {
	info := fieldInfoForSchema(sch)
	ts := &TableStats{
		Table:           tableName,
		RowsScanned:     sr.RowsScanned,
		RowsWithMissing: sr.RowsWithMissing,
		AdaptiveValues:  sr.AdaptiveValues,
		OutOfBandValues: sr.OutOfBandValues,
		MissingValues:   sr.missingValues(),
		MissingKeyValues: func() uint64 {
			var n uint64
			for _, c := range sr.MissingByKeyIdx {
				n += c
			}
			return n
		}(),
	}
	for _, i := range valIdxs {
		ts.AdaptiveColumns = append(ts.AdaptiveColumns, info.valName(i))
	}
	for _, i := range keyIdxs {
		ts.AdaptiveKeyColumns = append(ts.AdaptiveKeyColumns, info.keyName(i))
	}
	if sr.impacted() {
		ts.MissingByColumn = make(map[string]uint64)
		for i, n := range sr.MissingByValIdx {
			ts.MissingByColumn[info.valName(i)] += n
		}
		for i, n := range sr.MissingByKeyIdx {
			ts.MissingByColumn[info.keyName(i)+" (key)"] += n
		}
	}
	return ts
}

// selectBranches returns the branches to process, sorted by name. When |branchFilter| is
// non-empty, only the branch with that name is returned; if the database has no such branch,
// an error is recorded in |dbRep| and the returned slice is empty.
func selectBranches(sqlCtx *sql.Context, ddb *doltdb.DoltDB, branchFilter string, dbRep *DatabaseReport) ([]ref.DoltRef, error) {
	branches, err := ddb.GetBranches(sqlCtx)
	if err != nil {
		return nil, err
	}
	if branchFilter != "" {
		for _, branch := range branches {
			if branch.GetPath() == branchFilter {
				return []ref.DoltRef{branch}, nil
			}
		}
		dbRep.Errors = append(dbRep.Errors, fmt.Sprintf("branch %q not found in this database", branchFilter))
		return nil, nil
	}
	sort.Slice(branches, func(i, j int) bool { return branches[i].GetPath() < branches[j].GetPath() })
	return branches, nil
}

// reportDatabase runs report mode against a single database: it scans the HEAD of every branch
// (or just |branchFilter|, if non-empty) for tables with adaptive-encoded columns and checks
// every out-of-band value against the chunk store, without modifying anything.
func reportDatabase(sqlCtx *sql.Context, dbName string, ddb *doltdb.DoltDB, branchFilter string) (*DatabaseReport, error) {
	dbRep := &DatabaseReport{Database: dbName}
	cs := datas.ChunkStoreFromDatabase(doltdb.ExposeDatabaseFromDoltDB(ddb))
	scanner := newDbScanner(cs)

	branches, err := selectBranches(sqlCtx, ddb, branchFilter, dbRep)
	if err != nil {
		return nil, err
	}

	for _, branch := range branches {
		br := &BranchReport{Branch: branch.GetPath(), CommitsScanned: 1}
		dbRep.Branches = append(dbRep.Branches, br)

		cm, err := ddb.ResolveCommitRef(sqlCtx, branch)
		if err != nil {
			dbRep.Errors = append(dbRep.Errors, fmt.Sprintf("branch %s: resolving head: %v", branch.GetPath(), err))
			continue
		}
		root, err := cm.GetRootValue(sqlCtx)
		if err != nil {
			dbRep.Errors = append(dbRep.Errors, fmt.Sprintf("branch %s: reading root: %v", branch.GetPath(), err))
			continue
		}

		err = root.IterTables(sqlCtx, func(name doltdb.TableName, tbl *doltdb.Table, sch schema.Schema) (bool, error) {
			ts, impactedSchema, err := scanTableForReport(sqlCtx, scanner, name, tbl, sch)
			if err != nil {
				dbRep.Errors = append(dbRep.Errors, fmt.Sprintf("branch %s: table %s: %v", branch.GetPath(), name.String(), err))
				return false, nil
			}
			if !impactedSchema {
				br.SkippedTables++
				return false, nil
			}
			br.Tables = append(br.Tables, ts)
			return false, nil
		})
		if err != nil {
			dbRep.Errors = append(dbRep.Errors, fmt.Sprintf("branch %s: iterating tables: %v", branch.GetPath(), err))
		}
		sortTables(br.Tables)
		progress("  branch %s: %d tables with adaptive columns scanned, %d skipped", br.Branch, len(br.Tables), br.SkippedTables)
	}
	return dbRep, nil
}

// scanTableForReport scans a single table's row data. It returns impactedSchema=false when the
// table has no adaptive-encoded columns and was skipped. Panics from the storage layer (e.g.
// dangling addresses reached through unexpected paths) are converted to errors.
func scanTableForReport(sqlCtx *sql.Context, scanner *dbScanner, name doltdb.TableName, tbl *doltdb.Table, sch schema.Schema) (ts *TableStats, impactedSchema bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic while scanning: %v", r)
		}
	}()

	rows, err := tbl.GetRowData(sqlCtx)
	if err != nil {
		return nil, false, err
	}
	m, err := durable.ProllyMapFromIndex(rows)
	if err != nil {
		return nil, false, err
	}
	kd, vd := m.Descriptors()
	keyIdxs, valIdxs := adaptiveFieldIndexes(kd), adaptiveFieldIndexes(vd)
	if len(keyIdxs) == 0 && len(valIdxs) == 0 {
		return nil, false, nil
	}
	sr, err := scanner.scanMap(sqlCtx, m)
	if err != nil {
		return nil, true, err
	}
	return tableStatsFromScan(name.String(), sch, sr, keyIdxs, valIdxs), true, nil
}

func sortTables(tables []*TableStats) {
	sort.Slice(tables, func(i, j int) bool { return tables[i].Table < tables[j].Table })
}

// aggregateTable finds or creates the aggregate TableStats for |name| in |br| and accumulates
// |ts| into it.
func aggregateTable(br *BranchReport, ts *TableStats) {
	for _, existing := range br.Tables {
		if existing.Table == ts.Table {
			existing.add(ts)
			return
		}
	}
	fresh := &TableStats{
		Table:              ts.Table,
		AdaptiveColumns:    ts.AdaptiveColumns,
		AdaptiveKeyColumns: ts.AdaptiveKeyColumns,
	}
	fresh.add(ts)
	br.Tables = append(br.Tables, fresh)
}
