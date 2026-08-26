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

// Package integrity detects a form of storage corruption introduced by earlier DoltgreSQL releases:
// prolly tree node messages whose value_address_offsets field omits entries for adaptive-encoded
// values stored out of band. Nodes written this way are missing chunk references, which causes push
// and clone to omit the out-of-band chunks, and garbage collection to delete them, resulting in data
// loss. See cmd/admin for the offline tool that reports and repairs this corruption.
package integrity

import (
	"context"
	"hash/fnv"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/gen/fb/serial"
	"github.com/dolthub/dolt/go/store/chunks"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/val"
)

// Stats records corruption statistics aggregated over a prolly tree node and all its children.
type Stats struct {
	// Chunks is the total number of tree node chunks in the subtree.
	Chunks uint64
	// LeafChunks is the number of leaf node chunks in the subtree.
	LeafChunks uint64
	// CorruptChunks is the number of leaf chunks with at least one missing key or value address offset.
	CorruptChunks uint64
	// Rows is the total number of rows in the subtree.
	Rows uint64
	// CorruptRows is the number of rows with at least one out-of-band value missing from the node's
	// recorded key or value address offsets.
	CorruptRows uint64
	// AdaptiveValues is the number of non-NULL adaptive-encoded values in value tuples.
	AdaptiveValues uint64
	// OutOfBandValues is the number of adaptive-encoded values stored out-of-band in value tuples.
	OutOfBandValues uint64
	// CorruptValues is the number of expected chunk address references (out-of-band adaptive values and
	// non-empty address-encoded fields) missing from the node's value_address_offsets field.
	CorruptValues uint64
	// UnexpectedOffsets is the number of recorded value_address_offsets entries that do not correspond to
	// any address in the node's value tuples. Always zero for both healthy and known-corrupt databases.
	UnexpectedOffsets uint64
	// KeyAdaptiveValues is the number of non-NULL adaptive-encoded values in key tuples.
	KeyAdaptiveValues uint64
	// KeyOutOfBandValues is the number of adaptive-encoded values stored out-of-band in key tuples.
	KeyOutOfBandValues uint64
	// KeyCorruptValues is the number of expected chunk address references missing from the node's
	// key_address_offsets field. Nodes written before that field existed record no key addresses at
	// all, so any out-of-band key value they hold counts here.
	KeyCorruptValues uint64
	// MissingChunks is the number of out-of-band values (key or value side) whose chunks are absent from
	// the chunk store. These values are already lost (e.g. to a previous GC) and cannot be repaired by
	// rewriting nodes.
	MissingChunks uint64
}

// Add accumulates another node's statistics into this one.
func (s *Stats) Add(o *Stats) {
	s.Chunks += o.Chunks
	s.LeafChunks += o.LeafChunks
	s.CorruptChunks += o.CorruptChunks
	s.Rows += o.Rows
	s.CorruptRows += o.CorruptRows
	s.AdaptiveValues += o.AdaptiveValues
	s.OutOfBandValues += o.OutOfBandValues
	s.CorruptValues += o.CorruptValues
	s.UnexpectedOffsets += o.UnexpectedOffsets
	s.KeyAdaptiveValues += o.KeyAdaptiveValues
	s.KeyOutOfBandValues += o.KeyOutOfBandValues
	s.KeyCorruptValues += o.KeyCorruptValues
	s.MissingChunks += o.MissingChunks
}

// CacheKey identifies a scanned subtree: the node chunk hash plus a fingerprint of the tuple descriptors
// it was interpreted with. Identical chunks shared between branches, commits, and structurally identical
// tables are only ever processed once.
type CacheKey struct {
	Addr hash.Hash
	Desc uint64
}

// DescFingerprint returns a cheap fingerprint of the key and value tuple descriptors' encodings.
func DescFingerprint(kd, vd *val.TupleDesc) uint64 {
	h := fnv.New64a()
	for _, t := range kd.Types {
		_, _ = h.Write([]byte{byte(t.Enc)})
	}
	_, _ = h.Write([]byte{0xff})
	for _, t := range vd.Types {
		_, _ = h.Write([]byte{byte(t.Enc)})
	}
	return h.Sum64()
}

// Scanner walks prolly trees at the chunk level and detects leaf nodes whose value_address_offsets field
// omits references to out-of-band values. Results are cached per chunk, so re-scanning the same table
// on another branch or commit is nearly free.
type Scanner struct {
	cs    chunks.ChunkStore
	cache map[CacheKey]*Stats
	// presence caches chunk-existence lookups for out-of-band value chunks.
	presence map[hash.Hash]bool

	// CacheHits counts subtree scans satisfied from the cache.
	CacheHits uint64
}

func NewScanner(cs chunks.ChunkStore) *Scanner {
	return &Scanner{
		cs:       cs,
		cache:    make(map[CacheKey]*Stats),
		presence: make(map[hash.Hash]bool),
	}
}

// ScanTable scans the primary row storage of the given table.
func (s *Scanner) ScanTable(ctx context.Context, ti *TableInfo) (*Stats, error) {
	m, err := ti.RowMap(ctx)
	if err != nil {
		return nil, err
	}
	return s.ScanTree(ctx, m.Node().HashOf(), ti.KeyDesc, ti.ValDesc)
}

// ScanTree scans the subtree rooted at |addr| and returns aggregated statistics.
func (s *Scanner) ScanTree(ctx context.Context, addr hash.Hash, kd, vd *val.TupleDesc) (*Stats, error) {
	key := CacheKey{Addr: addr, Desc: DescFingerprint(kd, vd)}
	if cached, ok := s.cache[key]; ok {
		s.CacheHits++
		return cached, nil
	}

	msg, err := GetTreeNodeMessage(ctx, s.cs, addr)
	if err != nil {
		return nil, err
	}
	var pm serial.ProllyTreeNode
	err = serial.InitProllyTreeNodeRoot(&pm, msg, serial.MessagePrefixSz)
	if err != nil {
		return nil, err
	}

	stats := &Stats{Chunks: 1}
	if pm.TreeLevel() > 0 {
		for _, childAddr := range ChildAddresses(&pm) {
			childStats, err := s.ScanTree(ctx, childAddr, kd, vd)
			if err != nil {
				return nil, err
			}
			stats.Add(childStats)
		}
	} else {
		la, err := AnalyzeLeaf(&pm, kd, vd)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to analyze leaf node %s", addr.String())
		}
		stats.Add(&la.Stats)
		stats.LeafChunks = 1
		if la.Corrupt {
			stats.CorruptChunks = 1
		}

		missing, err := s.countMissing(ctx, append(la.ValueOOBAddrs, la.KeyOOBAddrs...))
		if err != nil {
			return nil, err
		}
		stats.MissingChunks = missing
	}

	s.cache[key] = stats
	return stats, nil
}

// countMissing returns how many of the given chunk addresses are absent from the chunk store.
func (s *Scanner) countMissing(ctx context.Context, addrs []hash.Hash) (uint64, error) {
	toCheck := hash.NewHashSet()
	for _, a := range addrs {
		if _, ok := s.presence[a]; !ok {
			toCheck.Insert(a)
		}
	}
	if toCheck.Size() > 0 {
		absent, err := s.cs.HasMany(ctx, toCheck)
		if err != nil {
			return 0, err
		}
		for a := range toCheck {
			s.presence[a] = !absent.Has(a)
		}
	}
	var missing uint64
	for _, a := range addrs {
		if !s.presence[a] {
			missing++
		}
	}
	return missing, nil
}

// GetTreeNodeMessage reads the chunk at |addr| and validates that it is a ProllyTreeNode message.
func GetTreeNodeMessage(ctx context.Context, cs chunks.ChunkStore, addr hash.Hash) (serial.Message, error) {
	c, err := cs.Get(ctx, addr)
	if err != nil {
		return nil, err
	}
	if c.IsEmpty() {
		return nil, errors.Errorf("chunk %s not found in chunk store", addr.String())
	}
	msg := serial.Message(c.Data())
	if id := serial.GetFileID(msg); id != serial.ProllyTreeNodeFileID {
		return nil, errors.Errorf("chunk %s: expected a %s message, found %s",
			addr.String(), serial.ProllyTreeNodeFileID, id)
	}
	return msg, nil
}

// ChildAddresses returns the child chunk addresses of an internal tree node.
func ChildAddresses(pm *serial.ProllyTreeNode) []hash.Hash {
	arr := pm.AddressArrayBytes()
	addrs := make([]hash.Hash, 0, len(arr)/hash.ByteLen)
	for i := 0; i+hash.ByteLen <= len(arr); i += hash.ByteLen {
		addrs = append(addrs, hash.New(arr[i:i+hash.ByteLen]))
	}
	return addrs
}

// LeafAnalysis is the result of examining a single leaf node's tuples against its recorded
// value_address_offsets field.
type LeafAnalysis struct {
	Stats Stats
	// Corrupt is true when at least one expected offset is missing from value_address_offsets or
	// key_address_offsets; such nodes must be rewritten to be repaired.
	Corrupt bool
	// ValueOOBAddrs are the chunk addresses referenced from value tuples (out-of-band adaptive values
	// and address-encoded fields).
	ValueOOBAddrs []hash.Hash
	// KeyOOBAddrs are the chunk addresses referenced by out-of-band adaptive values in key tuples.
	KeyOOBAddrs []hash.Hash
}

// AnalyzeLeaf recomputes the expected value_address_offsets and key_address_offsets for a leaf node
// from its tuples, mirroring the (fixed) serializer logic in dolt's go/store/prolly/message package,
// and compares the result against the offsets actually recorded in the message.
func AnalyzeLeaf(pm *serial.ProllyTreeNode, kd, vd *val.TupleDesc) (*LeafAnalysis, error) {
	la := &LeafAnalysis{}

	n := pm.KeyOffsetsLength() - 1
	if n <= 0 {
		return la, nil
	}
	if pm.ValueOffsetsLength()-1 != n {
		return nil, errors.Errorf("leaf node has %d keys but %d values", n, pm.ValueOffsetsLength()-1)
	}

	valueSide := analyzeTupleAddresses(vd, pm.ValueItemsBytes(), pm.ValueOffsets, n,
		recordedOffsets(pm.ValueAddressOffsetsLength(), pm.ValueAddressOffsets))
	keySide := analyzeTupleAddresses(kd, pm.KeyItemsBytes(), pm.KeyOffsets, n,
		recordedOffsets(pm.KeyAddressOffsetsLength(), pm.KeyAddressOffsets))

	la.Stats.AdaptiveValues = valueSide.adaptiveValues
	la.Stats.OutOfBandValues = valueSide.outOfBandValues
	la.Stats.CorruptValues = valueSide.corruptValues
	la.ValueOOBAddrs = valueSide.oobAddrs
	la.Stats.KeyAdaptiveValues = keySide.adaptiveValues
	la.Stats.KeyOutOfBandValues = keySide.outOfBandValues
	la.Stats.KeyCorruptValues = keySide.corruptValues
	la.KeyOOBAddrs = keySide.oobAddrs

	la.Stats.Rows = uint64(n)
	for i := 0; i < n; i++ {
		if valueSide.corruptRows[i] || keySide.corruptRows[i] {
			la.Stats.CorruptRows++
			la.Corrupt = true
		}
	}

	la.Stats.UnexpectedOffsets = uint64(pm.ValueAddressOffsetsLength()-valueSide.matched) +
		uint64(pm.KeyAddressOffsetsLength()-keySide.matched)

	return la, nil
}

// sideAnalysis is the result of checking the tuples of one side (keys or values) of a leaf node
// against the address offsets recorded for that side.
type sideAnalysis struct {
	adaptiveValues  uint64
	outOfBandValues uint64
	corruptValues   uint64
	matched         int
	corruptRows     []bool
	oobAddrs        []hash.Hash
}

// recordedOffsets collects one side's recorded address offsets into a multiset. Offsets are unique in
// practice, but count matching makes the comparison exact.
func recordedOffsets(length int, offsetAt func(int) uint16) map[uint16]int {
	recorded := make(map[uint16]int, length)
	for i := 0; i < length; i++ {
		recorded[offsetAt(i)]++
	}
	return recorded
}

// analyzeTupleAddresses computes the expected address offsets for |n| tuples of one side of a leaf node
// and matches them against the |recorded| offsets, consuming matches from the multiset.
func analyzeTupleAddresses(desc *val.TupleDesc, items []byte, offsetAt func(int) uint16, n int, recorded map[uint16]int) sideAnalysis {
	sa := sideAnalysis{corruptRows: make([]bool, n)}
	match := func(expected uint16, row int) {
		if recorded[expected] > 0 {
			recorded[expected]--
			sa.matched++
		} else {
			sa.corruptValues++
			sa.corruptRows[row] = true
		}
	}

	for i := 0; i < n; i++ {
		start := int(offsetAt(i))
		end := int(offsetAt(i + 1))
		tup := val.Tuple(items[start:end])

		val.IterAddressFields(desc, func(j int, t val.Type) {
			field := tup.GetField(j)
			if len(field) == 0 || hash.New(field).IsEmpty() {
				return
			}
			sa.oobAddrs = append(sa.oobAddrs, hash.New(field))
			off, _ := tup.GetOffset(j)
			match(uint16(start+off), i)
		})

		val.IterAdaptiveFields(desc, func(j int, t val.Type) {
			field := tup.GetField(j)
			av := val.AdaptiveValue(field)
			if av.IsNull() {
				return
			}
			sa.adaptiveValues++
			if !av.IsOutOfBand() {
				return
			}
			sa.outOfBandValues++
			if addr, err := av.OutOfBandAddr(); err == nil {
				sa.oobAddrs = append(sa.oobAddrs, addr)
			}
			// Out-of-band adaptive values end in an address, so the expected offset is
			// hash.ByteLen bytes before the end of the field.
			off, _ := tup.GetOffset(j)
			match(uint16(start+off+len(field)-hash.ByteLen), i)
		})
	}
	return sa
}
