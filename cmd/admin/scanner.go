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
	"fmt"
	"io"

	"github.com/dolthub/dolt/go/store/chunks"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/pool"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/val"
)

// presenceBatchSize is the maximum number of unresolved chunk addresses buffered before
// issuing a HasMany call to the chunk store.
const presenceBatchSize = 4096

// rowBatchSize is the maximum number of rows buffered while awaiting presence resolution.
const rowBatchSize = 65536

// filterBatchSize is the maximum number of entries buffered per batch when rebuilding a map
// with corrupted keys. Smaller than rowBatchSize because every entry in a batch is retained
// until its addresses resolve, not just the corrupted ones.
const filterBatchSize = 16384

// scanResult records the outcome of scanning a single prolly map for adaptive-encoded values
// whose out-of-band chunks are missing. Field indexes are positions in the map's key/value
// tuple descriptors; they are translated to column names by the caller, which knows the schema.
type scanResult struct {
	RowsScanned     uint64
	RowsWithMissing uint64
	AdaptiveValues  uint64
	OutOfBandValues uint64
	MissingByValIdx map[int]uint64
	MissingByKeyIdx map[int]uint64
}

func (sr *scanResult) missingValues() uint64 {
	var n uint64
	for _, c := range sr.MissingByValIdx {
		n += c
	}
	return n
}

func (sr *scanResult) missingKeyValues() uint64 {
	var n uint64
	for _, c := range sr.MissingByKeyIdx {
		n += c
	}
	return n
}

func (sr *scanResult) impacted() bool {
	return len(sr.MissingByValIdx) > 0 || len(sr.MissingByKeyIdx) > 0
}

// repairResult is a cached table repair: the rewritten map along with the scan stats that
// produced it. If keyCorruption is true, the map was not repaired because one or more primary
// key fields reference missing chunks, which cannot be NULLed out.
type repairResult struct {
	repaired      prolly.Map
	changed       bool
	keyCorruption bool
	stats         *scanResult
}

// dbScanner scans and repairs the tables of a single database. It memoizes work at two levels:
//   - presence: chunk address -> whether the chunk exists in this database's chunk store
//   - scanCache/repairCache: prolly map root hash -> scan/repair results, so that structurally
//     identical tables encountered on other branches or commits are processed only once
//
// Caches are per-database because chunk presence depends on the database's chunk store.
type dbScanner struct {
	cs           chunks.ChunkStore
	presence     map[hash.Hash]bool
	scanCache    map[hash.Hash]*scanResult
	repairCache  map[hash.Hash]*repairResult
	rebuildCache map[hash.Hash]*rebuildResult
}

func newDbScanner(cs chunks.ChunkStore) *dbScanner {
	return &dbScanner{
		cs:           cs,
		presence:     make(map[hash.Hash]bool),
		scanCache:    make(map[hash.Hash]*scanResult),
		repairCache:  make(map[hash.Hash]*repairResult),
		rebuildCache: make(map[hash.Hash]*rebuildResult),
	}
}

// adaptiveFieldIndexes returns the field indexes in |td| that use an adaptive encoding.
func adaptiveFieldIndexes(td *val.TupleDesc) []int {
	var idxs []int
	for i := 0; i < td.Count(); i++ {
		if val.IsAdaptiveEncoding(td.Types[i].Enc) {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

// scanMap scans every row of |m|, checking that every out-of-band adaptive value's chunk is
// present in the chunk store. Results are memoized by the map's root hash.
func (s *dbScanner) scanMap(ctx context.Context, m prolly.Map) (*scanResult, error) {
	rootHash := m.HashOf()
	if cached, ok := s.scanCache[rootHash]; ok {
		return cached, nil
	}

	kd, vd := m.Descriptors()
	keyIdxs, valIdxs := adaptiveFieldIndexes(kd), adaptiveFieldIndexes(vd)

	res := &scanResult{
		MissingByValIdx: make(map[int]uint64),
		MissingByKeyIdx: make(map[int]uint64),
	}
	iter, err := m.IterAll(ctx)
	if err != nil {
		return nil, err
	}

	batch := newRowBatch()
	flush := func() error {
		if err := s.resolvePresence(ctx, batch.unknown); err != nil {
			return err
		}
		for _, row := range batch.rows {
			s.tallyRow(row, res)
		}
		batch.reset()
		return nil
	}

	for {
		k, v, err := iter.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		res.RowsScanned++
		row := s.examineRow(k, v, keyIdxs, valIdxs, res, batch)
		if row != nil {
			batch.rows = append(batch.rows, *row)
		}
		if len(batch.unknown) >= presenceBatchSize || len(batch.rows) >= rowBatchSize {
			if err := flush(); err != nil {
				return nil, err
			}
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}

	s.scanCache[rootHash] = res
	return res, nil
}

// rowRef records the out-of-band addresses found in a single row, pending presence resolution.
type rowRef struct {
	key     val.Tuple
	value   val.Tuple
	keyRefs []fieldRef
	valRefs []fieldRef
}

type fieldRef struct {
	fieldIdx int
	addr     hash.Hash
}

type rowBatch struct {
	rows    []rowRef
	unknown hash.HashSet
}

func newRowBatch() *rowBatch {
	return &rowBatch{unknown: hash.NewHashSet()}
}

func (b *rowBatch) reset() {
	b.rows = b.rows[:0]
	b.unknown = hash.NewHashSet()
}

// examineRow inspects the adaptive fields of a row, updating inline-value counters in |res|
// immediately and returning a rowRef if the row contains out-of-band values that are either
// unresolved or already known to be missing. Rows whose addresses are all already known to be
// present are fully accounted for here and are not buffered, keeping memory bounded when
// scanning large healthy tables. Unresolved addresses are added to |batch.unknown|.
func (s *dbScanner) examineRow(k, v val.Tuple, keyIdxs, valIdxs []int, res *scanResult, batch *rowBatch) *rowRef {
	var row *rowRef
	needed := false
	examine := func(tup val.Tuple, idxs []int, isKey bool) {
		for _, i := range idxs {
			field := val.AdaptiveValue(tup.GetField(i))
			if field.IsNull() {
				continue
			}
			res.AdaptiveValues++
			if !field.IsOutOfBand() {
				continue
			}
			res.OutOfBandValues++
			addr, err := field.OutOfBandAddr()
			if err != nil {
				// Cannot happen: we already checked for NULL and inline.
				continue
			}
			if row == nil {
				row = &rowRef{key: k, value: v}
			}
			ref := fieldRef{fieldIdx: i, addr: addr}
			if isKey {
				row.keyRefs = append(row.keyRefs, ref)
			} else {
				row.valRefs = append(row.valRefs, ref)
			}
			present, known := s.presence[addr]
			if !known {
				batch.unknown.Insert(addr)
				needed = true
			} else if !present {
				needed = true
			}
		}
	}
	examine(k, keyIdxs, true)
	examine(v, valIdxs, false)
	if !needed {
		return nil
	}
	return row
}

// resolvePresence issues a HasMany call for the given addresses and records the results in the
// presence cache.
func (s *dbScanner) resolvePresence(ctx context.Context, addrs hash.HashSet) error {
	if len(addrs) == 0 {
		return nil
	}
	absent, err := s.cs.HasMany(ctx, addrs)
	if err != nil {
		return fmt.Errorf("checking chunk presence: %w", err)
	}
	for h := range addrs {
		s.presence[h] = !absent.Has(h)
	}
	return nil
}

// tallyRow updates |res| with the missing-value counts for a row whose addresses have all been
// resolved in the presence cache.
func (s *dbScanner) tallyRow(row rowRef, res *scanResult) (missingKey, missingVal []int) {
	for _, ref := range row.keyRefs {
		if !s.presence[ref.addr] {
			res.MissingByKeyIdx[ref.fieldIdx]++
			missingKey = append(missingKey, ref.fieldIdx)
		}
	}
	for _, ref := range row.valRefs {
		if !s.presence[ref.addr] {
			res.MissingByValIdx[ref.fieldIdx]++
			missingVal = append(missingVal, ref.fieldIdx)
		}
	}
	if len(missingKey) > 0 || len(missingVal) > 0 {
		res.RowsWithMissing++
	}
	return missingKey, missingVal
}

// repairMap produces a new version of |m| in which every value field referencing a missing
// chunk has been replaced with NULL. Results are memoized by the map's root hash.
//
// The map is scanned (via scanMap) before anything is mutated. This matters for correctness:
// if any primary KEY field references a missing chunk, the map is left unmodified and
// keyCorruption is set — key fields cannot be NULLed, and mutating a map whose key tuples
// cannot all be compared (adaptive key comparison dereferences chunks) would panic. The scan
// also fully populates the presence cache, so the mutation pass needs no batched lookups.
func (s *dbScanner) repairMap(ctx context.Context, m prolly.Map) (*repairResult, error) {
	rootHash := m.HashOf()
	if cached, ok := s.repairCache[rootHash]; ok {
		return cached, nil
	}

	sr, err := s.scanMap(ctx, m)
	if err != nil {
		return nil, err
	}

	var result *repairResult
	switch {
	case sr.missingKeyValues() > 0:
		result = &repairResult{repaired: m, keyCorruption: true, stats: sr}
	case sr.missingValues() == 0:
		result = &repairResult{repaired: m, stats: sr}
	default:
		_, vd := m.Descriptors()
		iter, err := m.IterAll(ctx)
		if err != nil {
			return nil, err
		}
		mutIter := &mutationTupleIter{
			ctx:     ctx,
			scanner: s,
			src:     iter,
			valIdxs: adaptiveFieldIndexes(vd),
			vd:      vd,
			pool:    m.Pool(),
		}
		repaired, err := prolly.MutateMapWithTupleIter(ctx, m, mutIter)
		if err != nil {
			return nil, err
		}
		if mutIter.err != nil {
			return nil, mutIter.err
		}
		result = &repairResult{repaired: repaired, changed: mutIter.mutations > 0, stats: sr}
	}

	s.repairCache[rootHash] = result
	return result, nil
}

type repairedRow struct {
	key   val.Tuple
	value val.Tuple
}

// mutationTupleIter adapts a table scan into a prolly.TupleIter of mutations: it yields
// (key, repairedValue) pairs for rows containing missing out-of-band value fields, in key
// order, terminating with a nil key. It relies on the presence cache having been fully
// populated by a prior scanMap of the same map. Because prolly.TupleIter's Next cannot return
// an error, errors are recorded in |err| and terminate the stream; callers must check it.
type mutationTupleIter struct {
	ctx     context.Context
	scanner *dbScanner
	src     prolly.MapIter
	valIdxs []int
	vd      *val.TupleDesc
	pool    pool.BuffPool

	mutations uint64
	err       error
}

func (it *mutationTupleIter) Next(ctx context.Context) (val.Tuple, val.Tuple) {
	for {
		k, v, err := it.src.Next(it.ctx)
		if err == io.EOF {
			return nil, nil
		}
		if err != nil {
			it.err = err
			return nil, nil
		}
		missing := it.scanner.missingFields(collectOutOfBandRefs(v, it.valIdxs))
		if len(missing) == 0 {
			continue
		}
		it.mutations++
		return k, nullOutFields(it.pool, it.vd, v, missing)
	}
}

// rebuildResult is a cached map rebuild: the rewritten map (which may have had corrupted KEY
// fields NULLed) along with whether anything changed.
type rebuildResult struct {
	rebuilt prolly.Map
	changed bool
}

// rebuildMapWithCorruptKeys repairs a map that may contain dangling out-of-band addresses in
// its KEY fields, such as a secondary index map. Corrupted keys cannot be located with ordinary
// map edits, because navigating to them compares adaptive key fields, which dereferences their
// (missing) chunks. Instead the map is rebuilt with a linear scan: healthy entries are streamed
// in their original order into a new map, and corrupted entries are then re-inserted with their
// dangling fields NULLed out. Results are memoized by the map's root hash.
func (s *dbScanner) rebuildMapWithCorruptKeys(ctx context.Context, m prolly.Map, keyIdxs, valIdxs []int) (prolly.Map, bool, error) {
	rootHash := m.HashOf()
	if cached, ok := s.rebuildCache[rootHash]; ok {
		return cached.rebuilt, cached.changed, nil
	}

	// A plain scan (cached, and cheap relative to a rebuild) determines whether this map has
	// any missing chunks at all before we commit to rewriting it.
	sr, err := s.scanMap(ctx, m)
	if err != nil {
		return prolly.Map{}, false, err
	}
	if !sr.impacted() {
		s.rebuildCache[rootHash] = &rebuildResult{rebuilt: m, changed: false}
		return m, false, nil
	}

	srcIter, err := m.IterAll(ctx)
	if err != nil {
		return prolly.Map{}, false, err
	}
	kd, vd := m.Descriptors()
	iter := &filterTupleIter{
		ctx:     ctx,
		scanner: s,
		src:     srcIter,
		keyIdxs: keyIdxs,
		valIdxs: valIdxs,
	}
	rebuilt, err := prolly.NewMapFromTupleIter(ctx, m.NodeStore(), kd, vd, iter)
	if err != nil {
		return prolly.Map{}, false, err
	}
	if iter.err != nil {
		return prolly.Map{}, false, iter.err
	}

	// Re-insert the corrupted entries with their dangling fields NULLed. Their nulled keys only
	// ever compare against healthy keys (and NULL fields compare without chunk reads), so this
	// is safe.
	mut := rebuilt.Mutate()
	for _, entry := range iter.corrupt {
		newKey := entry.key
		if len(entry.missingKey) > 0 {
			newKey = nullOutFields(m.Pool(), kd, entry.key, entry.missingKey)
		}
		newVal := entry.value
		if len(entry.missingVal) > 0 {
			newVal = nullOutFields(m.Pool(), vd, entry.value, entry.missingVal)
		}
		if err = mut.Put(ctx, newKey, newVal); err != nil {
			return prolly.Map{}, false, err
		}
	}
	fixed, err := mut.Map(ctx)
	if err != nil {
		return prolly.Map{}, false, err
	}

	s.rebuildCache[rootHash] = &rebuildResult{rebuilt: fixed, changed: true}
	return fixed, true, nil
}

// corruptEntry is a map entry containing dangling out-of-band addresses, along with the field
// indexes that are dangling.
type corruptEntry struct {
	key, value val.Tuple
	missingKey []int
	missingVal []int
}

// filterTupleIter streams every entry of a map in order, except entries containing dangling
// addresses, which are recorded in |corrupt| instead of being emitted. Errors are recorded in
// |err| and terminate the stream.
type filterTupleIter struct {
	ctx     context.Context
	scanner *dbScanner
	src     prolly.MapIter
	keyIdxs []int
	valIdxs []int

	emit      []repairedRow
	emitIdx   int
	corrupt   []corruptEntry
	exhausted bool
	err       error
}

func (it *filterTupleIter) Next(ctx context.Context) (val.Tuple, val.Tuple) {
	for {
		if it.emitIdx < len(it.emit) {
			pair := it.emit[it.emitIdx]
			it.emitIdx++
			return pair.key, pair.value
		}
		if it.exhausted || it.err != nil {
			return nil, nil
		}
		it.fillBatch()
	}
}

func (it *filterTupleIter) fillBatch() {
	it.emit = it.emit[:0]
	it.emitIdx = 0

	type pendingEntry struct {
		key, value val.Tuple
		keyRefs    []fieldRef
		valRefs    []fieldRef
	}
	var entries []pendingEntry
	unknown := hash.NewHashSet()

	for !it.exhausted && len(entries) < filterBatchSize && len(unknown) < presenceBatchSize {
		k, v, err := it.src.Next(it.ctx)
		if err == io.EOF {
			it.exhausted = true
			break
		}
		if err != nil {
			it.err = err
			return
		}
		entry := pendingEntry{
			key:     k,
			value:   v,
			keyRefs: collectOutOfBandRefs(k, it.keyIdxs),
			valRefs: collectOutOfBandRefs(v, it.valIdxs),
		}
		for _, ref := range entry.keyRefs {
			if _, known := it.scanner.presence[ref.addr]; !known {
				unknown.Insert(ref.addr)
			}
		}
		for _, ref := range entry.valRefs {
			if _, known := it.scanner.presence[ref.addr]; !known {
				unknown.Insert(ref.addr)
			}
		}
		entries = append(entries, entry)
	}

	if err := it.scanner.resolvePresence(it.ctx, unknown); err != nil {
		it.err = err
		return
	}

	for _, entry := range entries {
		missingKey := it.scanner.missingFields(entry.keyRefs)
		missingVal := it.scanner.missingFields(entry.valRefs)
		if len(missingKey) == 0 && len(missingVal) == 0 {
			it.emit = append(it.emit, repairedRow{key: entry.key, value: entry.value})
		} else {
			it.corrupt = append(it.corrupt, corruptEntry{
				key:        entry.key,
				value:      entry.value,
				missingKey: missingKey,
				missingVal: missingVal,
			})
		}
	}
}

// collectOutOfBandRefs returns the out-of-band address references in the given fields of a tuple.
func collectOutOfBandRefs(tup val.Tuple, idxs []int) []fieldRef {
	var refs []fieldRef
	for _, i := range idxs {
		field := val.AdaptiveValue(tup.GetField(i))
		if field.IsNull() || !field.IsOutOfBand() {
			continue
		}
		addr, err := field.OutOfBandAddr()
		if err != nil {
			continue
		}
		refs = append(refs, fieldRef{fieldIdx: i, addr: addr})
	}
	return refs
}

// missingFields returns the field indexes of refs whose chunks are known to be absent, per the
// presence cache. Addresses that have not been resolved are conservatively treated as present:
// callers are expected to have resolved every address first (via scanMap or resolvePresence),
// and an unresolved address must never cause a value to be NULLed.
func (s *dbScanner) missingFields(refs []fieldRef) []int {
	var missing []int
	for _, ref := range refs {
		if present, known := s.presence[ref.addr]; known && !present {
			missing = append(missing, ref.fieldIdx)
		}
	}
	return missing
}

// nullOutFields returns a copy of |tup| with the given fields set to NULL. The copy is built
// from raw field bytes so that no adaptive value is ever dereferenced or re-encoded.
func nullOutFields(pool pool.BuffPool, td *val.TupleDesc, tup val.Tuple, nullIdxs []int) val.Tuple {
	fields := make([][]byte, td.Count())
	for i := 0; i < td.Count(); i++ {
		fields[i] = tup.GetField(i)
	}
	for _, i := range nullIdxs {
		fields[i] = nil
	}
	return val.NewTuple(pool, fields...)
}
