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

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb/durable"
	"github.com/dolthub/dolt/go/libraries/doltcore/ref"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
	"github.com/dolthub/dolt/go/store/datas"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/go-mysql-server/sql"
)

// tableOutcome is the (cached) result of repairing a single table. Tables are cached by their
// table struct hash, so structurally identical tables encountered in other commits or branches
// are repaired only once.
type tableOutcome struct {
	name doltdb.TableName
	// newTable is the repaired table; nil when the table was not changed.
	newTable *doltdb.Table
	changed  bool
	// keyCorruption indicates the table contains missing out-of-band values in primary key
	// fields and was left unmodified: key fields cannot be NULLed out.
	keyCorruption bool
	// err records a non-fatal error that prevented repairing this table; it was left unmodified.
	err   error
	stats *TableStats
}

// rootOutcome is the (cached) result of repairing a root value.
type rootOutcome struct {
	newRoot       doltdb.RootValue
	changed       bool
	tables        []*tableOutcome
	skippedTables uint64
}

type dbRepairer struct {
	ddb     *doltdb.DoltDB
	scanner *dbScanner
	sqlCtx  *sql.Context

	// visited maps original commit hashes to their replacement commits (which may be the
	// original commit if nothing in its history changed).
	visited map[hash.Hash]*doltdb.Commit
	// rootCache maps root value hashes to repair outcomes.
	rootCache map[hash.Hash]*rootOutcome
	// tableCache maps table struct hashes to repair outcomes.
	tableCache map[hash.Hash]*tableOutcome
}

// repairDatabase runs repair mode against a single database: for every commit reachable from
// every branch (or just |branchFilter|, if non-empty), it NULLs out adaptive-encoded values
// whose out-of-band chunks are missing, rewrites the commit DAG with the repaired roots, and
// re-points branch refs and working sets. Commits whose history contains no corruption are
// left untouched (their hashes are preserved).
func repairDatabase(sqlCtx *sql.Context, dbName string, ddb *doltdb.DoltDB, branchFilter string) (*DatabaseReport, error) {
	dbRep := &DatabaseReport{Database: dbName}
	cs := datas.ChunkStoreFromDatabase(doltdb.ExposeDatabaseFromDoltDB(ddb))
	r := &dbRepairer{
		ddb:        ddb,
		scanner:    newDbScanner(cs),
		sqlCtx:     sqlCtx,
		visited:    make(map[hash.Hash]*doltdb.Commit),
		rootCache:  make(map[hash.Hash]*rootOutcome),
		tableCache: make(map[hash.Hash]*tableOutcome),
	}

	branches, err := selectBranches(sqlCtx, ddb, branchFilter, dbRep)
	if err != nil {
		return nil, err
	}

	for _, branch := range branches {
		br := &BranchReport{Branch: branch.GetPath()}
		dbRep.Branches = append(dbRep.Branches, br)

		head, err := ddb.ResolveCommitRef(sqlCtx, branch)
		if err != nil {
			dbRep.Errors = append(dbRep.Errors, fmt.Sprintf("branch %s: resolving head: %v", br.Branch, err))
			continue
		}

		newHead, err := r.rewriteHistory(head, br, dbRep)
		if err != nil {
			dbRep.Errors = append(dbRep.Errors, fmt.Sprintf("branch %s: rewriting history: %v", br.Branch, err))
			continue
		}

		oldHash, err := head.HashOf()
		if err != nil {
			return nil, err
		}
		newHash, err := newHead.HashOf()
		if err != nil {
			return nil, err
		}
		if oldHash != newHash {
			if err = ddb.NewBranchAtCommitAllowCaseConflict(sqlCtx, branch, newHead, nil); err != nil {
				dbRep.Errors = append(dbRep.Errors, fmt.Sprintf("branch %s: updating branch ref: %v", br.Branch, err))
				continue
			}
			progress("  branch %s: head rewritten %s -> %s", br.Branch, oldHash.String(), newHash.String())
		} else {
			progress("  branch %s: no corruption in history, unchanged", br.Branch)
		}

		if err = r.repairWorkingSet(branch, dbRep); err != nil {
			dbRep.Errors = append(dbRep.Errors, fmt.Sprintf("branch %s: repairing working set: %v", br.Branch, err))
		}

		sortTables(br.Tables)
	}
	return dbRep, nil
}

// commitFrame is a stack frame in the iterative post-order traversal of the commit DAG.
type commitFrame struct {
	commit   *doltdb.Commit
	hash     hash.Hash
	parents  []*doltdb.Commit
	expanded bool
}

// rewriteHistory rewrites every commit reachable from |head| (skipping commits already visited
// via another branch), returning the replacement head. The traversal is an iterative post-order
// DFS so that arbitrarily deep histories do not overflow the stack. A commit is replaced only
// if its root value changed or any of its parents were replaced; otherwise the original commit
// (and hash) is preserved.
func (r *dbRepairer) rewriteHistory(head *doltdb.Commit, br *BranchReport, dbRep *DatabaseReport) (*doltdb.Commit, error) {
	headHash, err := head.HashOf()
	if err != nil {
		return nil, err
	}

	stack := []*commitFrame{{commit: head, hash: headHash}}
	for len(stack) > 0 {
		frame := stack[len(stack)-1]
		if _, done := r.visited[frame.hash]; done {
			stack = stack[:len(stack)-1]
			continue
		}

		if !frame.expanded {
			frame.expanded = true
			optParents, err := r.ddb.ResolveAllParents(r.sqlCtx, frame.commit)
			if err != nil {
				return nil, err
			}
			for _, optParent := range optParents {
				parent, ok := optParent.ToCommit()
				if !ok {
					return nil, doltdb.ErrGhostCommitEncountered
				}
				frame.parents = append(frame.parents, parent)
			}
			for _, parent := range frame.parents {
				ph, err := parent.HashOf()
				if err != nil {
					return nil, err
				}
				if _, done := r.visited[ph]; !done {
					stack = append(stack, &commitFrame{commit: parent, hash: ph})
				}
			}
			continue
		}

		// All parents have been processed; rewrite this commit.
		if err := r.rewriteCommit(frame, br, dbRep); err != nil {
			return nil, err
		}
		br.CommitsScanned++
		stack = stack[:len(stack)-1]
	}

	return r.visited[headHash], nil
}

// rewriteCommit repairs a single commit's root value and records its replacement commit in the
// visited map. The replacement is the original commit when neither the root nor any parent
// changed.
func (r *dbRepairer) rewriteCommit(frame *commitFrame, br *BranchReport, dbRep *DatabaseReport) error {
	newParents := make([]*doltdb.Commit, len(frame.parents))
	parentsChanged := false
	for i, parent := range frame.parents {
		ph, err := parent.HashOf()
		if err != nil {
			return err
		}
		newParent, ok := r.visited[ph]
		if !ok {
			return fmt.Errorf("parent commit %s processed out of order", ph.String())
		}
		newParents[i] = newParent
		nph, err := newParent.HashOf()
		if err != nil {
			return err
		}
		if nph != ph {
			parentsChanged = true
		}
	}

	root, err := frame.commit.GetRootValue(r.sqlCtx)
	if err != nil {
		return err
	}
	outcome, err := r.repairRoot(root)
	if err != nil {
		return err
	}
	r.recordOutcome(outcome, frame.hash, br, dbRep)

	if !outcome.changed && !parentsChanged {
		r.visited[frame.hash] = frame.commit
		return nil
	}

	var valueHash hash.Hash
	if outcome.changed {
		_, valueHash, err = r.ddb.WriteRootValue(r.sqlCtx, outcome.newRoot)
		if err != nil {
			return err
		}
	} else {
		valueHash, err = outcome.newRoot.HashOf()
		if err != nil {
			return err
		}
	}

	meta, err := frame.commit.GetCommitMeta(r.sqlCtx)
	if err != nil {
		return err
	}
	newCommit, err := r.ddb.CommitDanglingWithParentCommits(r.sqlCtx, valueHash, newParents, meta)
	if err != nil {
		return err
	}
	r.visited[frame.hash] = newCommit
	return nil
}

// recordOutcome accumulates a root repair outcome into the branch report and surfaces
// per-table problems as database errors.
func (r *dbRepairer) recordOutcome(outcome *rootOutcome, commitHash hash.Hash, br *BranchReport, dbRep *DatabaseReport) {
	for _, to := range outcome.tables {
		if to.stats != nil {
			aggregateTable(br, to.stats)
		}
		if to.keyCorruption {
			dbRep.Errors = append(dbRep.Errors, fmt.Sprintf(
				"branch %s: commit %s: table %s has missing out-of-band values in primary key columns; "+
					"key values cannot be NULLed, table left unrepaired at this commit",
				br.Branch, commitHash.String(), to.name.String()))
		}
		if to.err != nil {
			dbRep.Errors = append(dbRep.Errors, fmt.Sprintf(
				"branch %s: commit %s: table %s could not be repaired: %v",
				br.Branch, commitHash.String(), to.name.String(), to.err))
		}
	}
}

// repairWorkingSet repairs the working and staged roots of a branch's working set, if present.
func (r *dbRepairer) repairWorkingSet(branch ref.DoltRef, dbRep *DatabaseReport) error {
	wsRef, err := ref.WorkingSetRefForHead(branch)
	if err != nil {
		return err
	}
	ws, err := r.ddb.ResolveWorkingSet(r.sqlCtx, wsRef)
	if err != nil {
		// Not every branch has a working set (e.g. databases only ever accessed remotely).
		return nil
	}

	changed := false
	if ws.WorkingRoot() != nil {
		outcome, err := r.repairRoot(ws.WorkingRoot())
		if err != nil {
			return err
		}
		if outcome.changed {
			ws = ws.WithWorkingRoot(outcome.newRoot)
			changed = true
		}
	}
	if ws.StagedRoot() != nil {
		outcome, err := r.repairRoot(ws.StagedRoot())
		if err != nil {
			return err
		}
		if outcome.changed {
			ws = ws.WithStagedRoot(outcome.newRoot)
			changed = true
		}
	}
	if !changed {
		return nil
	}

	prevHash, err := ws.HashOf()
	if err != nil {
		return err
	}
	return r.ddb.UpdateWorkingSet(r.sqlCtx, wsRef, ws, prevHash, ws.Meta(), nil)
}

// repairRoot repairs every impacted table in a root value, returning a new root when any table
// changed. Outcomes are memoized by root value hash.
func (r *dbRepairer) repairRoot(root doltdb.RootValue) (*rootOutcome, error) {
	rootHash, err := root.HashOf()
	if err != nil {
		return nil, err
	}
	if cached, ok := r.rootCache[rootHash]; ok {
		return cached, nil
	}

	outcome := &rootOutcome{newRoot: root}
	err = root.IterTables(r.sqlCtx, func(name doltdb.TableName, tbl *doltdb.Table, sch schema.Schema) (bool, error) {
		to := r.repairTable(name, tbl, sch)
		if to == nil {
			outcome.skippedTables++
			return false, nil
		}
		outcome.tables = append(outcome.tables, to)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	newRoot := root
	for _, to := range outcome.tables {
		if !to.changed {
			continue
		}
		newRoot, err = newRoot.PutTable(r.sqlCtx, to.name, to.newTable)
		if err != nil {
			return nil, err
		}
		outcome.changed = true
	}
	outcome.newRoot = newRoot

	r.rootCache[rootHash] = outcome
	return outcome, nil
}

// repairTable scans and repairs a single table, returning nil when the table has no
// adaptive-encoded columns. Outcomes are memoized by table struct hash. Any panic from the
// storage layer is converted into a per-table error, leaving the table unmodified.
func (r *dbRepairer) repairTable(name doltdb.TableName, tbl *doltdb.Table, sch schema.Schema) (out *tableOutcome) {
	tblHash, err := tbl.HashOf()
	if err != nil {
		return &tableOutcome{name: name, err: err}
	}
	if cached, ok := r.tableCache[tblHash]; ok {
		return cached
	}

	defer func() {
		if p := recover(); p != nil {
			out = &tableOutcome{name: name, err: fmt.Errorf("panic while repairing: %v", p)}
		}
		// A nil outcome (no adaptive columns) is cached too, so the schema check runs once.
		r.tableCache[tblHash] = out
	}()

	out = r.repairTableInner(name, tbl, sch)
	return out
}

func (r *dbRepairer) repairTableInner(name doltdb.TableName, tbl *doltdb.Table, sch schema.Schema) *tableOutcome {
	rows, err := tbl.GetRowData(r.sqlCtx)
	if err != nil {
		return &tableOutcome{name: name, err: err}
	}
	m, err := durable.ProllyMapFromIndex(rows)
	if err != nil {
		return &tableOutcome{name: name, err: err}
	}
	kd, vd := m.Descriptors()
	keyIdxs, valIdxs := adaptiveFieldIndexes(kd), adaptiveFieldIndexes(vd)
	if len(keyIdxs) == 0 && len(valIdxs) == 0 {
		return nil
	}

	rr, err := r.scanner.repairMap(r.sqlCtx, m)
	if err != nil {
		return &tableOutcome{name: name, err: err}
	}
	stats := tableStatsFromScan(name.String(), sch, rr.stats, keyIdxs, valIdxs)
	outcome := &tableOutcome{name: name, stats: stats, keyCorruption: rr.keyCorruption}
	if rr.keyCorruption || !rr.changed {
		return outcome
	}

	newTbl, err := tbl.UpdateRows(r.sqlCtx, durable.IndexFromProllyMap(rr.repaired))
	if err != nil {
		outcome.err = err
		return outcome
	}

	// Secondary indexes copy adaptive fields into their keys as raw bytes, so any index over a
	// corrupted column embeds the same dangling addresses and must be repaired as well.
	for _, idx := range sch.Indexes().AllIndexes() {
		newTbl, err = r.fixSecondaryIndex(newTbl, idx)
		if err != nil {
			// Repairing rows without fixing the index would leave the table inconsistent, so
			// abandon the repair of this table entirely.
			return &tableOutcome{name: name, stats: stats, err: fmt.Errorf("updating index %s: %w", idx.Name(), err)}
		}
	}

	stats.RepairedValues = rr.stats.missingValues()
	outcome.newTable = newTbl
	outcome.changed = true
	return outcome
}

// fixSecondaryIndex repairs a secondary index whose keys embed dangling out-of-band addresses.
//
// Index entries cannot be repaired with ordinary map edits: locating an entry whose key embeds
// a missing address requires comparing that key against neighboring keys, and comparing
// adaptive fields dereferences their chunks, which panics for missing chunks. Instead, the
// index is rebuilt with a linear scan that never compares corrupted keys: every healthy entry
// is streamed in order into a new map, and each corrupted entry is re-inserted afterwards with
// its dangling key fields NULLed out, matching the NULLs written to the primary rows.
func (r *dbRepairer) fixSecondaryIndex(tbl *doltdb.Table, idx schema.Index) (*doltdb.Table, error) {
	idxData, err := tbl.GetIndexRowData(r.sqlCtx, idx.Name())
	if err != nil {
		return nil, err
	}
	idxMap, err := durable.ProllyMapFromIndex(idxData)
	if err != nil {
		return nil, err
	}
	kd, vd := idxMap.Descriptors()
	keyIdxs := adaptiveFieldIndexes(kd)
	if len(keyIdxs) == 0 {
		return tbl, nil
	}

	fixed, changed, err := r.scanner.rebuildMapWithCorruptKeys(r.sqlCtx, idxMap, keyIdxs, adaptiveFieldIndexes(vd))
	if err != nil {
		return nil, err
	}
	if !changed {
		return tbl, nil
	}
	return tbl.SetIndexRows(r.sqlCtx, idx.Name(), durable.IndexFromProllyMap(fixed))
}
