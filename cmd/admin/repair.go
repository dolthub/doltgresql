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
	"bytes"
	"log"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/gen/fb/serial"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb/durable"
	"github.com/dolthub/dolt/go/libraries/doltcore/ref"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/pool"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/message"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/dolt/go/store/val"
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

// repairer rewrites prolly tree nodes whose value_address_offsets field omits references to out-of-band
// values. Corrupt leaf nodes are re-serialized with the current serializer using their original
// key and value tuple bytes, which recomputes the correct offsets; ancestor nodes are then rewritten to
// reference the repaired children. Results are cached per chunk, so identical subtrees shared between
// branches, commits, and structurally identical tables are only repaired once.
type repairer struct {
	db *database
	// treeCache maps old node chunk hashes to their repaired equivalents (identity for clean subtrees).
	treeCache     map[integrity.CacheKey]hash.Hash
	summary       repairSummary
	verbose       bool
	transformLeaf leafTransform
}

// leafTransform inspects a leaf node and returns a replacement message, or false if the node is
// unchanged.
type leafTransform func(r *repairer, pm *serial.ProllyTreeNode, kd, vd *val.TupleDesc) (serial.Message, bool, error)

func newRepairer(db *database, verbose bool) *repairer {
	return &repairer{
		db:            db,
		treeCache:     make(map[integrity.CacheKey]hash.Hash),
		verbose:       verbose,
		transformLeaf: repairLeafTransform,
	}
}

// repairDatabase rewrites every commit reachable from every branch and tag of the database, repairing
// all corrupted table data, then updates branch, tag, and working set refs to the rewritten commits.
// Commits whose content (and ancestry) is unaffected keep their original hashes.
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
		if !ti.Impacted() && !ti.KeyImpacted() {
			continue
		}
		m, err := ti.RowMap(sctx)
		if err != nil {
			return nil, false, err
		}
		oldAddr := m.Node().HashOf()
		newAddr, err := r.repairTree(sctx, oldAddr, ti.KeyDesc, ti.ValDesc)
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

// repairTree repairs the subtree rooted at |addr|, returning the (possibly identical) new root address.
func (r *repairer) repairTree(sctx *sql.Context, addr hash.Hash, kd, vd *val.TupleDesc) (hash.Hash, error) {
	key := integrity.CacheKey{Addr: addr, Desc: integrity.DescFingerprint(kd, vd)}
	if newAddr, ok := r.treeCache[key]; ok {
		return newAddr, nil
	}

	msg, err := integrity.GetTreeNodeMessage(sctx, r.db.cs, addr)
	if err != nil {
		return hash.Hash{}, err
	}
	var pm serial.ProllyTreeNode
	err = serial.InitProllyTreeNodeRoot(&pm, msg, serial.MessagePrefixSz)
	if err != nil {
		return hash.Hash{}, err
	}

	newAddr := addr
	if pm.TreeLevel() == 0 {
		newMsg, changed, err := r.transformLeaf(r, &pm, kd, vd)
		if err != nil {
			return hash.Hash{}, errors.Wrapf(err, "failed to rewrite leaf node %s", addr.String())
		}
		if changed {
			newAddr, err = r.writeNode(sctx, newMsg)
			if err != nil {
				return hash.Hash{}, errors.Wrapf(err, "failed to write rewritten leaf node %s", addr.String())
			}
			r.summary.LeafChunksRewritten++
		}
	} else {
		children := integrity.ChildAddresses(&pm)
		newChildren := make([]hash.Hash, len(children))
		childrenChanged := false
		for i, child := range children {
			newChild, err := r.repairTree(sctx, child, kd, vd)
			if err != nil {
				return hash.Hash{}, err
			}
			newChildren[i] = newChild
			childrenChanged = childrenChanged || newChild != child
		}
		if childrenChanged {
			newAddr, err = r.rewriteInternal(sctx, msg, &pm, newChildren, kd, vd)
			if err != nil {
				return hash.Hash{}, errors.Wrapf(err, "failed to rewrite internal node %s", addr.String())
			}
			r.summary.InternalChunksRewritten++
		}
	}

	r.treeCache[key] = newAddr
	return newAddr, nil
}

// repairLeafTransform re-serializes a corrupt leaf node with the current serializer, using the node's
// original key and value tuple bytes. The serializer recomputes value_address_offsets from the tuple
// contents, which repairs the omitted entries. Leaf nodes with correct offsets are left unchanged.
func repairLeafTransform(r *repairer, pm *serial.ProllyTreeNode, kd, vd *val.TupleDesc) (serial.Message, bool, error) {
	la, err := integrity.AnalyzeLeaf(pm, kd, vd)
	if err != nil {
		return nil, false, err
	}
	if !la.Corrupt {
		return nil, false, nil
	}

	newMsg, err := reserializeLeaf(pm, kd, vd, r.db.ns.Pool())
	if err != nil {
		return nil, false, err
	}

	// The rewritten node's recomputed offsets must be complete.
	var npm serial.ProllyTreeNode
	err = serial.InitProllyTreeNodeRoot(&npm, newMsg, serial.MessagePrefixSz)
	if err != nil {
		return nil, false, err
	}
	newLa, err := integrity.AnalyzeLeaf(&npm, kd, vd)
	if err != nil {
		return nil, false, err
	}
	if newLa.Corrupt || newLa.Stats.UnexpectedOffsets != 0 {
		return nil, false, errors.New("rewritten leaf node still has incorrect value_address_offsets")
	}
	return newMsg, true, nil
}

// reserializeLeaf re-serializes a leaf node's original key and value tuple bytes with the given
// descriptor and verifies that the result is tuple-for-tuple identical to the original.
func reserializeLeaf(pm *serial.ProllyTreeNode, kd, vd *val.TupleDesc, pool pool.BuffPool) (serial.Message, error) {
	keys := extractItems(pm.KeyItemsBytes(), pm.KeyOffsetsLength(), pm.KeyOffsets)
	values := extractItems(pm.ValueItemsBytes(), pm.ValueOffsetsLength(), pm.ValueOffsets)

	serializer := message.NewProllyMapSerializer(kd, vd, pool)
	newMsg := serializer.Serialize(keys, values, nil, 0)

	// Sanity-check the rewritten node: tuples must be byte-identical to the original.
	var npm serial.ProllyTreeNode
	err := serial.InitProllyTreeNodeRoot(&npm, newMsg, serial.MessagePrefixSz)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(npm.KeyItemsBytes(), pm.KeyItemsBytes()) || !bytes.Equal(npm.ValueItemsBytes(), pm.ValueItemsBytes()) {
		return nil, errors.New("rewritten leaf node has different tuple bytes than the original")
	}
	if npm.TreeCount() != pm.TreeCount() {
		return nil, errors.Errorf("rewritten leaf node has tree count %d, expected %d", npm.TreeCount(), pm.TreeCount())
	}
	return newMsg, nil
}

// rewriteInternal re-serializes an internal node, replacing its child addresses with the repaired ones.
// Keys, subtree counts, and level are preserved from the original node.
func (r *repairer) rewriteInternal(sctx *sql.Context, oldMsg serial.Message, pm *serial.ProllyTreeNode, newChildren []hash.Hash, kd, vd *val.TupleDesc) (hash.Hash, error) {
	keys := extractItems(pm.KeyItemsBytes(), pm.KeyOffsetsLength(), pm.KeyOffsets)
	if len(keys) != len(newChildren) {
		return hash.Hash{}, errors.Errorf("internal node has %d keys but %d children", len(keys), len(newChildren))
	}

	values := make([][]byte, len(newChildren))
	for i := range newChildren {
		addr := newChildren[i]
		values[i] = addr[:]
	}

	subtrees, err := subtreeCounts(oldMsg, len(newChildren))
	if err != nil {
		return hash.Hash{}, err
	}

	serializer := message.NewProllyMapSerializer(kd, vd, r.db.ns.Pool())
	newMsg := serializer.Serialize(keys, values, subtrees, int(pm.TreeLevel()))

	// Sanity-check the rewritten node before writing it.
	var npm serial.ProllyTreeNode
	err = serial.InitProllyTreeNodeRoot(&npm, newMsg, serial.MessagePrefixSz)
	if err != nil {
		return hash.Hash{}, err
	}
	if !bytes.Equal(npm.KeyItemsBytes(), pm.KeyItemsBytes()) {
		return hash.Hash{}, errors.New("rewritten internal node has different key bytes than the original")
	}
	if npm.TreeLevel() != pm.TreeLevel() || npm.TreeCount() != pm.TreeCount() {
		return hash.Hash{}, errors.New("rewritten internal node has different level or tree count than the original")
	}
	newAddrs := integrity.ChildAddresses(&npm)
	if len(newAddrs) != len(newChildren) {
		return hash.Hash{}, errors.New("rewritten internal node has wrong child count")
	}
	for i := range newAddrs {
		if newAddrs[i] != newChildren[i] {
			return hash.Hash{}, errors.New("rewritten internal node has wrong child addresses")
		}
	}

	return r.writeNode(sctx, newMsg)
}

// writeNode writes a serialized tree node message to the node store and returns its address.
func (r *repairer) writeNode(sctx *sql.Context, msg serial.Message) (hash.Hash, error) {
	node, fileID, err := tree.NodeFromBytes(msg)
	if err != nil {
		return hash.Hash{}, err
	}
	if fileID != serial.ProllyTreeNodeFileID {
		return hash.Hash{}, errors.Errorf("rewritten node has unexpected file ID %s", fileID)
	}
	return r.db.ns.Write(sctx, node)
}

// extractItems splits a flatbuffer item buffer into individual items using its offsets vector.
// |offsetsLen| is the length of the offsets vector, which contains one more entry than there are items.
func extractItems(buf []byte, offsetsLen int, offsetAt func(int) uint16) [][]byte {
	if offsetsLen <= 1 {
		return nil
	}
	items := make([][]byte, offsetsLen-1)
	for i := 0; i < offsetsLen-1; i++ {
		items[i] = buf[offsetAt(i):offsetAt(i+1)]
	}
	return items
}

// subtreeCounts decodes the subtree count array of an internal node message.
func subtreeCounts(msg serial.Message, count int) ([]uint64, error) {
	node, fileID, err := tree.NodeFromBytes(msg)
	if err != nil {
		return nil, err
	}
	if fileID != serial.ProllyTreeNodeFileID {
		return nil, errors.Errorf("unexpected file ID %s", fileID)
	}
	node, err = node.LoadSubtrees()
	if err != nil {
		return nil, err
	}
	counts := make([]uint64, count)
	for i := 0; i < count; i++ {
		counts[i] = node.GetSubtreeCount(i)
	}
	return counts, nil
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
