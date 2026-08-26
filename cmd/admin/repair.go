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
	"log"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb/durable"
	"github.com/dolthub/dolt/go/libraries/doltcore/ref"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/integrity"
)

// repairSummary records the work performed while repairing one database.
type repairSummary struct {
	CommitsExamined         int
	CommitsRewritten        int
	BranchesUpdated         int
	TagsUpdated             int
	WorkingSetsFixed        int
	LeafChunksRewritten     uint64
	InternalChunksRewritten uint64
}

// repairer repairs the corruption found by the integrity scanner across a database's entire commit
// graph. The storage-level detection and node rewriting live in the integrity package (Scanner and
// TreeRewriter, which share behavior and caches with the reporting scan and the server's startup
// check); the repairer orchestrates them across every commit, branch, tag, and working set, updating
// refs to the rewritten history. Commits whose content (and ancestry) is unaffected keep their
// original hashes.
type repairer struct {
	db       *database
	rewriter *integrity.TreeRewriter
	summary  repairSummary
	verbose  bool
}

// newRepairer returns a repairer for the given database. The rewriter is built on |sc|, so a scan
// phase that already ran with it (e.g. the report) shares its per-chunk results with the repair.
func newRepairer(db *database, sc *integrity.Scanner, verbose bool) *repairer {
	return &repairer{
		db:       db,
		rewriter: integrity.NewTreeRewriter(sc, db.ns),
		verbose:  verbose,
	}
}

// repairDatabase rewrites every commit reachable from every branch and tag of the database, repairing
// all corrupted table data, then updates branch, tag, and working set refs to the rewritten commits.
func (r *repairer) repairDatabase(sctx *sql.Context) (*repairSummary, error) {
	ddb := r.db.ddb
	branches, err := ddb.GetBranches(sctx)
	if err != nil {
		return nil, err
	}
	tags, err := ddb.GetTags(sctx)
	if err != nil {
		return nil, err
	}

	// visited maps original commit hashes to their repaired commits (possibly the original commit
	// itself), shared across all refs so common history is only processed once.
	visited := make(map[hash.Hash]*doltdb.Commit)

	for _, branchRef := range branches {
		head, err := ddb.ResolveCommitRef(sctx, branchRef)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to resolve branch %s", branchRef.String())
		}

		// Capture the working set before updating the branch ref: NewBranchAtCommit resets the
		// working set to the new head, which would discard uncommitted changes.
		wsRef, err := ref.WorkingSetRefForHead(branchRef)
		if err != nil {
			return nil, err
		}
		var origWorking, origStaged doltdb.RootValue
		hasWorkingSet := false
		if ws, wsErr := ddb.ResolveWorkingSet(sctx, wsRef); wsErr == nil {
			// Not all branches have working sets (e.g. branches never checked out).
			hasWorkingSet = true
			origWorking = ws.WorkingRoot()
			origStaged = ws.StagedRoot()
		}

		newHead, err := r.repairCommit(sctx, head, visited)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to repair history of branch %s", branchRef.String())
		}

		headChanged, err := commitsDiffer(head, newHead)
		if err != nil {
			return nil, err
		}
		if headChanged {
			err = ddb.NewBranchAtCommitAllowCaseConflict(sctx, branchRef, newHead, nil)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to update branch %s", branchRef.String())
			}
			r.summary.BranchesUpdated++
			if r.verbose {
				log.Printf("updated branch %s to repaired commit", branchRef.String())
			}
		}

		if hasWorkingSet {
			err = r.restoreWorkingSet(sctx, wsRef, origWorking, origStaged)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to repair working set of branch %s", branchRef.String())
			}
		}
	}

	for _, tagRef := range tags {
		tr, ok := tagRef.(ref.TagRef)
		if !ok {
			continue
		}
		tag, err := ddb.ResolveTag(sctx, tr)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to resolve tag %s", tagRef.String())
		}
		newCm, err := r.repairCommit(sctx, tag.Commit, visited)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to repair history of tag %s", tagRef.String())
		}
		changed, err := commitsDiffer(tag.Commit, newCm)
		if err != nil {
			return nil, err
		}
		if changed {
			if err = ddb.DeleteTag(sctx, tr); err != nil {
				return nil, err
			}
			if err = ddb.NewTagAtCommit(sctx, tr, newCm, tag.Meta); err != nil {
				return nil, err
			}
			r.summary.TagsUpdated++
			if r.verbose {
				log.Printf("updated tag %s to repaired commit", tagRef.String())
			}
		}
	}

	r.summary.LeafChunksRewritten = r.rewriter.LeafChunksRewritten
	r.summary.InternalChunksRewritten = r.rewriter.InternalChunksRewritten
	return &r.summary, nil
}

// restoreWorkingSet repairs the original (pre-branch-update) working and staged roots and stores them
// in the branch's working set if they differ from what is currently stored there. This both repairs
// corrupted uncommitted data and restores uncommitted changes that the branch ref update reset.
func (r *repairer) restoreWorkingSet(sctx *sql.Context, wsRef ref.WorkingSetRef, origWorking, origStaged doltdb.RootValue) error {
	ws, err := r.db.ddb.ResolveWorkingSet(sctx, wsRef)
	if err != nil {
		return err
	}

	changed := false
	if origWorking != nil {
		newRoot, _, err := r.repairRootValue(sctx, origWorking)
		if err != nil {
			return err
		}
		differs, err := rootsDiffer(ws.WorkingRoot(), newRoot)
		if err != nil {
			return err
		}
		if differs {
			ws = ws.WithWorkingRoot(newRoot)
			changed = true
		}
	}
	if origStaged != nil {
		newRoot, _, err := r.repairRootValue(sctx, origStaged)
		if err != nil {
			return err
		}
		differs, err := rootsDiffer(ws.StagedRoot(), newRoot)
		if err != nil {
			return err
		}
		if differs {
			ws = ws.WithStagedRoot(newRoot)
			changed = true
		}
	}
	if !changed {
		return nil
	}

	// WorkingSet.HashOf returns the hash the working set had when it was loaded, which is what
	// UpdateWorkingSet requires for its optimistic locking check.
	currHash, err := ws.HashOf()
	if err != nil {
		return err
	}
	err = r.db.ddb.UpdateWorkingSet(sctx, wsRef, ws, currHash, ws.Meta(), nil)
	if err != nil {
		return err
	}
	r.summary.WorkingSetsFixed++
	if r.verbose {
		log.Printf("repaired working set %s", wsRef.String())
	}
	return nil
}

// rootsDiffer returns whether two root values have different hashes. A nil root differs from any
// non-nil root.
func rootsDiffer(a, b doltdb.RootValue) (bool, error) {
	if a == nil || b == nil {
		return a != nil || b != nil, nil
	}
	ah, err := a.HashOf()
	if err != nil {
		return false, err
	}
	bh, err := b.HashOf()
	if err != nil {
		return false, err
	}
	return ah != bh, nil
}

// repairCommit recursively repairs the given commit and all its ancestors (parents first), returning
// the repaired commit. If neither the commit's root value nor any ancestor was modified, the original
// commit is returned and its hash is preserved.
func (r *repairer) repairCommit(sctx *sql.Context, cm *doltdb.Commit, visited map[hash.Hash]*doltdb.Commit) (*doltdb.Commit, error) {
	h, err := cm.HashOf()
	if err != nil {
		return nil, err
	}
	if v, ok := visited[h]; ok {
		return v, nil
	}
	r.summary.CommitsExamined++

	optParents, err := r.db.ddb.ResolveAllParents(sctx, cm)
	if err != nil {
		return nil, err
	}
	parents := make([]*doltdb.Commit, len(optParents))
	for i, opt := range optParents {
		parent, ok := opt.ToCommit()
		if !ok {
			return nil, errors.Errorf("commit %s has a ghost parent: cannot repair shallow clones", h.String())
		}
		parents[i] = parent
	}

	parentsChanged := false
	newParents := make([]*doltdb.Commit, len(parents))
	for i, parent := range parents {
		newParent, err := r.repairCommit(sctx, parent, visited)
		if err != nil {
			return nil, err
		}
		newParents[i] = newParent
		differ, err := commitsDiffer(parent, newParent)
		if err != nil {
			return nil, err
		}
		parentsChanged = parentsChanged || differ
	}

	root, err := cm.GetRootValue(sctx)
	if err != nil {
		return nil, err
	}
	newRoot, rootChanged, err := r.repairRootValue(sctx, root)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to repair root value of commit %s", h.String())
	}

	if !rootChanged && !parentsChanged {
		visited[h] = cm
		return cm, nil
	}

	_, valueHash, err := r.db.ddb.WriteRootValue(sctx, newRoot)
	if err != nil {
		return nil, err
	}
	meta, err := cm.GetCommitMeta(sctx)
	if err != nil {
		return nil, err
	}
	newCm, err := r.db.ddb.CommitDanglingWithParentCommits(sctx, valueHash, newParents, meta)
	if err != nil {
		return nil, err
	}
	r.summary.CommitsRewritten++
	if r.verbose {
		log.Printf("rewrote commit %s (root changed: %v)", h.String(), rootChanged)
	}
	visited[h] = newCm
	return newCm, nil
}

// repairRootValue repairs the row data of every impacted table in the given root value, returning the
// updated root and whether anything changed.
func (r *repairer) repairRootValue(sctx *sql.Context, root doltdb.RootValue) (doltdb.RootValue, bool, error) {
	tables, err := integrity.TablesForRoot(sctx, root, r.db.ns)
	if err != nil {
		return nil, false, err
	}

	changed := false
	for _, ti := range tables {
		if !ti.ValColsImpacted() && !ti.KeyColsImpacted() {
			continue
		}
		m, err := ti.RowMap(sctx)
		if err != nil {
			return nil, false, err
		}
		oldAddr := m.Node().HashOf()
		newAddr, err := r.rewriter.RewriteTree(sctx, oldAddr, ti.KeyDesc, ti.ValDesc)
		if err != nil {
			return nil, false, errors.Wrapf(err, "failed to repair table %s", ti.Name.String())
		}
		if newAddr == oldAddr {
			continue
		}

		newRootNode, err := r.db.ns.Read(sctx, newAddr)
		if err != nil {
			return nil, false, err
		}
		newMap := prolly.NewMap(newRootNode, r.db.ns, ti.KeyDesc, ti.ValDesc)
		newTbl, err := ti.Tbl.UpdateRows(sctx, durable.IndexFromProllyMap(newMap))
		if err != nil {
			return nil, false, err
		}
		root, err = root.PutTable(sctx, ti.Name, newTbl)
		if err != nil {
			return nil, false, err
		}
		changed = true
	}
	return root, changed, nil
}

// commitsDiffer returns whether two commits have different hashes.
func commitsDiffer(a, b *doltdb.Commit) (bool, error) {
	ah, err := a.HashOf()
	if err != nil {
		return false, err
	}
	bh, err := b.HashOf()
	if err != nil {
		return false, err
	}
	return ah != bh, nil
}
