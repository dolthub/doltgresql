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
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/gen/fb/serial"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/pool"
	"github.com/dolthub/dolt/go/store/prolly/message"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/dolt/go/store/val"
	"github.com/dolthub/doltgresql/core/integrity"
)

// LeafTransform inspects a leaf node and returns a replacement message, or false if the node needs
// no rewrite.
type LeafTransform func(rw *TreeRewriter, pm *serial.ProllyTreeNode, kd, vd *val.TupleDesc) (serial.Message, bool, error)

// TreeRewriter rewrites prolly tree nodes whose address offset fields omit references to out-of-band
// values. It is built on a Scanner, which is the single authority on which subtrees contain corruption:
// subtrees the Scanner reports as clean are never visited, so a rewrite can never disagree with the
// scan (or the startup integrity check) about what needs repairing. Sharing one Scanner between a
// reporting scan and a rewrite also shares its per-chunk cache between the two phases.
//
// Corrupt leaf nodes are re-serialized with the current serializer using their original key and value
// tuple bytes, which recomputes the correct offsets; ancestor nodes are then rewritten to reference
// the repaired children. Results are cached per chunk, so identical subtrees shared between branches,
// commits, and structurally identical tables are only rewritten once.
type TreeRewriter struct {
	sc *integrity.Scanner
	ns tree.NodeStore

	// TransformLeaf produces the replacement message for a leaf node, or reports that the leaf needs
	// no rewrite. It defaults to RepairLeafTransform; tests substitute a corrupting transform to
	// simulate the serialization bugs in old releases.
	TransformLeaf LeafTransform
	// ShouldRewrite decides, from the Scanner's aggregated statistics for a subtree, whether the
	// subtree contains anything TransformLeaf would change; subtrees it rejects are skipped without
	// visiting their nodes. It defaults to descending into subtrees with corrupt chunks.
	ShouldRewrite func(*integrity.Stats) bool

	// cache maps old node chunk hashes to their rewritten equivalents (identity for skipped subtrees).
	cache map[integrity.CacheKey]hash.Hash

	// LeafChunksRewritten and InternalChunksRewritten count the nodes rewritten so far.
	LeafChunksRewritten     uint64
	InternalChunksRewritten uint64
}

// NewTreeRewriter returns a TreeRewriter that repairs the corruption detected by |sc|, writing
// rewritten nodes to |ns|.
func NewTreeRewriter(sc *integrity.Scanner, ns tree.NodeStore) *TreeRewriter {
	return &TreeRewriter{
		sc:            sc,
		ns:            ns,
		TransformLeaf: RepairLeafTransform,
		ShouldRewrite: func(stats *integrity.Stats) bool { return stats.CorruptChunks > 0 },
		cache:         make(map[integrity.CacheKey]hash.Hash),
	}
}

// Pool returns the buffer pool of the rewriter's node store, for use by leaf transforms.
func (rw *TreeRewriter) Pool() pool.BuffPool {
	return rw.ns.Pool()
}

// RewriteTree rewrites the subtree rooted at |addr|, returning the (possibly identical) new root
// address.
func (rw *TreeRewriter) RewriteTree(ctx context.Context, addr hash.Hash, kd, vd *val.TupleDesc) (hash.Hash, error) {
	key := integrity.CacheKey{Addr: addr, Desc: integrity.DescFingerprint(kd, vd)}
	if newAddr, ok := rw.cache[key]; ok {
		return newAddr, nil
	}

	// The Scanner decides whether this subtree contains anything worth rewriting.
	stats, err := rw.sc.ScanTree(ctx, addr, kd, vd)
	if err != nil {
		return hash.Hash{}, err
	}
	if !rw.ShouldRewrite(stats) {
		rw.cache[key] = addr
		return addr, nil
	}

	msg, err := integrity.GetTreeNodeMessage(ctx, rw.sc.Cs, addr)
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
		newMsg, changed, err := rw.TransformLeaf(rw, &pm, kd, vd)
		if err != nil {
			return hash.Hash{}, errors.Wrapf(err, "failed to rewrite leaf node %s", addr.String())
		}
		if changed {
			newAddr, err = rw.writeNode(ctx, newMsg)
			if err != nil {
				return hash.Hash{}, errors.Wrapf(err, "failed to write rewritten leaf node %s", addr.String())
			}
			rw.LeafChunksRewritten++
		}
	} else {
		children := integrity.ChildAddresses(&pm)
		newChildren := make([]hash.Hash, len(children))
		childrenChanged := false
		for i, child := range children {
			newChild, err := rw.RewriteTree(ctx, child, kd, vd)
			if err != nil {
				return hash.Hash{}, err
			}
			newChildren[i] = newChild
			childrenChanged = childrenChanged || newChild != child
		}
		if childrenChanged {
			newAddr, err = rw.rewriteInternal(ctx, msg, &pm, newChildren, kd, vd)
			if err != nil {
				return hash.Hash{}, errors.Wrapf(err, "failed to rewrite internal node %s", addr.String())
			}
			rw.InternalChunksRewritten++
		}
	}

	rw.cache[key] = newAddr
	return newAddr, nil
}

// RepairLeafTransform re-serializes a corrupt leaf node with the current serializer, using the node's
// original key and value tuple bytes. The serializer recomputes the address offset fields from the
// tuple contents, which repairs the omitted entries. Leaf nodes with correct offsets are left unchanged.
func RepairLeafTransform(rw *TreeRewriter, pm *serial.ProllyTreeNode, kd, vd *val.TupleDesc) (serial.Message, bool, error) {
	la, err := integrity.AnalyzeLeaf(pm, kd, vd)
	if err != nil {
		return nil, false, err
	}
	if !la.Corrupt {
		return nil, false, nil
	}

	newMsg, err := ReserializeLeaf(pm, kd, vd, rw.Pool())
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
		return nil, false, errors.New("rewritten leaf node still has incorrect address offsets")
	}
	return newMsg, true, nil
}

// ReserializeLeaf re-serializes a leaf node's original key and value tuple bytes with the given
// descriptors and verifies that the result is tuple-for-tuple identical to the original.
func ReserializeLeaf(pm *serial.ProllyTreeNode, kd, vd *val.TupleDesc, pool pool.BuffPool) (serial.Message, error) {
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

// rewriteInternal re-serializes an internal node, replacing its child addresses with the rewritten
// ones. Keys, subtree counts, and level are preserved from the original node.
func (rw *TreeRewriter) rewriteInternal(ctx context.Context, oldMsg serial.Message, pm *serial.ProllyTreeNode, newChildren []hash.Hash, kd, vd *val.TupleDesc) (hash.Hash, error) {
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

	serializer := message.NewProllyMapSerializer(kd, vd, rw.Pool())
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

	return rw.writeNode(ctx, newMsg)
}

// writeNode writes a serialized tree node message to the node store and returns its address.
func (rw *TreeRewriter) writeNode(ctx context.Context, msg serial.Message) (hash.Hash, error) {
	node, fileID, err := tree.NodeFromBytes(msg)
	if err != nil {
		return hash.Hash{}, err
	}
	if fileID != serial.ProllyTreeNodeFileID {
		return hash.Hash{}, errors.Errorf("rewritten node has unexpected file ID %s", fileID)
	}
	return rw.ns.Write(ctx, node)
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
