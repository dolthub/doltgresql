// Copyright 2024 Dolthub, Inc.
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

package sequences

import (
	"context"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/tree"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
)

// Collection contains a collection of sequences.
type Collection struct {
	accessedMap   map[id.Sequence]*Sequence // Whenever a sequence is accessed, it is added to the access map for faster retrieval
	underlyingMap prolly.AddressMap
	ns            tree.NodeStore
}

// Persistence controls the persistence of a Sequence.
type Persistence uint8

const (
	Persistence_Permanent Persistence = 0
	Persistence_Temporary Persistence = 1
	Persistence_Unlogged  Persistence = 2
)

// Sequence represents a single sequence within the pg_sequence table.
type Sequence struct {
	Id          id.Sequence
	DataTypeID  id.Type
	Persistence Persistence
	Start       int64
	Current     int64
	Increment   int64
	Minimum     int64
	Maximum     int64
	Cache       int64
	Cycle       bool
	IsAtEnd     bool
	OwnerTable  id.Table
	OwnerColumn string
}

var _ objinterface.Collection = (*Collection)(nil)
var _ objinterface.RootObject = (*Sequence)(nil)
var _ doltdb.RootObject = (*Sequence)(nil)

// GetSequence returns the sequence with the given schema and name. Returns nil if the sequence cannot be found.
func (pgs *Collection) GetSequence(ctx context.Context, name id.Sequence) (*Sequence, error) {
	return pgs.getSequence(ctx, name)
}

// GetSequencesWithTable returns all sequences with the given table as the owner.
func (pgs *Collection) GetSequencesWithTable(ctx context.Context, name doltdb.TableName) ([]*Sequence, error) {
	// For now, this function isn't used in a critical path, so we're not too worried about performance
	if err := pgs.cacheAllSequences(ctx); err != nil {
		return nil, err
	}
	var seqs []*Sequence
	nameID := id.NewTable(name.Schema, name.Name)
	for _, seq := range pgs.accessedMap {
		if seq.OwnerTable == nameID {
			seqs = append(seqs, seq)
		}
	}
	return seqs, nil
}

// GetAllSequences returns a map containing all sequences in the collection, grouped by the schema they're contained in.
// Each sequence array is also sorted by the sequence name.
func (pgs *Collection) GetAllSequences(ctx context.Context) (sequences map[string][]*Sequence, schemaNames []string, totalCount int, err error) {
	// For now, this function is only used by the "reg" types, so we're not too worried about performance
	if err = pgs.cacheAllSequences(ctx); err != nil {
		return nil, nil, 0, err
	}

	totalCount = len(pgs.accessedMap)
	schemaNamesMap := make(map[string]struct{})
	sequences = make(map[string][]*Sequence)
	for seqID, seq := range pgs.accessedMap {
		schemaNamesMap[seqID.SchemaName()] = struct{}{}
		sequences[seqID.SchemaName()] = append(sequences[seqID.SchemaName()], seq)
	}
	// Sort the sequences in the sequence map
	for _, seqs := range sequences {
		sort.Slice(seqs, func(i, j int) bool {
			return seqs[i].Id < seqs[j].Id
		})
	}
	// Create and sort the schema names
	schemaNames = make([]string, 0, len(schemaNamesMap))
	for name := range schemaNamesMap {
		schemaNames = append(schemaNames, name)
	}
	sort.Slice(schemaNames, func(i, j int) bool {
		return schemaNames[i] < schemaNames[j]
	})
	return
}

// HasSequence returns whether the sequence is present.
func (pgs *Collection) HasSequence(ctx context.Context, name id.Sequence) bool {
	// Subsequent loads are cached
	if _, ok := pgs.accessedMap[name]; ok {
		return true
	}
	// The initial load is from the internal map
	ok, err := pgs.underlyingMap.Has(ctx, string(name))
	if err == nil && ok {
		return true
	}
	return false
}

// CreateSequence creates a new sequence.
func (pgs *Collection) CreateSequence(ctx context.Context, seq *Sequence) error {
	// Ensure that the sequence does not already exist
	if _, ok := pgs.accessedMap[seq.Id]; ok {
		return errors.Errorf(`relation "%s" already exists`, seq.Id)
	}
	if ok, err := pgs.underlyingMap.Has(ctx, string(seq.Id)); err != nil {
		return err
	} else if ok {
		return errors.Errorf(`relation "%s" already exists`, seq.Id)
	}
	// Add it to our cache, which will be emptied when we do anything permanent
	pgs.accessedMap[seq.Id] = seq
	return nil
}

// DropSequence drops existing sequences.
func (pgs *Collection) DropSequence(ctx context.Context, names ...id.Sequence) (err error) {
	// We need to clear the cache so that we only need to worry about the underlying map
	if err = pgs.writeCache(ctx); err != nil {
		return err
	}
	for _, name := range names {
		if ok, err := pgs.underlyingMap.Has(ctx, string(name)); err != nil {
			return err
		} else if !ok {
			return errors.Errorf(`sequence "%s" does not exist`, name.SequenceName())
		}
	}
	// Now we'll remove the sequences from the underlying map
	mapEditor := pgs.underlyingMap.Editor()
	for _, name := range names {
		if err = mapEditor.Delete(ctx, string(name)); err != nil {
			return err
		}
	}
	pgs.underlyingMap, err = mapEditor.Flush(ctx)
	return err
}

// resolveName returns the fully resolved name of the given sequence. Returns an error if the name is ambiguous.
func (pgs *Collection) resolveName(ctx context.Context, schemaName string, sequenceName string) (id.Sequence, error) {
	if err := pgs.writeCache(ctx); err != nil {
		return id.NullSequence, err
	}
	count, err := pgs.underlyingMap.Count()
	if err != nil || count == 0 {
		return id.NullSequence, err
	}

	// First check for an exact match
	inputID := id.NewSequence(schemaName, sequenceName)
	ok, err := pgs.underlyingMap.Has(ctx, string(inputID))
	if err != nil {
		return id.NullSequence, err
	} else if ok {
		return inputID, nil
	}

	// Now we'll iterate over all the names
	var resolvedID id.Sequence
	if len(schemaName) > 0 {
		err = pgs.underlyingMap.IterAll(ctx, func(k string, _ hash.Hash) error {
			seqID := id.Sequence(k)
			if strings.EqualFold(sequenceName, seqID.SequenceName()) &&
				strings.EqualFold(schemaName, seqID.SchemaName()) {
				if resolvedID.IsValid() {
					return fmt.Errorf("`%s.%s` is ambiguous, matches `%s.%s` and `%s.%s`",
						schemaName, sequenceName, seqID.SchemaName(), seqID.SequenceName(), resolvedID.SchemaName(), resolvedID.SequenceName())
				}
				resolvedID = seqID
			}
			return nil
		})
		if err != nil {
			return id.NullSequence, err
		}
	} else {
		err = pgs.underlyingMap.IterAll(ctx, func(k string, _ hash.Hash) error {
			seqID := id.Sequence(k)
			if strings.EqualFold(sequenceName, seqID.SequenceName()) {
				if resolvedID.IsValid() {
					return fmt.Errorf("`%s` is ambiguous, matches `%s.%s` and `%s.%s`",
						sequenceName, seqID.SchemaName(), seqID.SequenceName(), resolvedID.SchemaName(), resolvedID.SequenceName())
				}
				resolvedID = seqID
			}
			return nil
		})
		if err != nil {
			return id.NullSequence, err
		}
	}
	return resolvedID, nil
}

// iterateIDs iterates over all sequence IDs in the collection.
func (pgs *Collection) iterateIDs(ctx context.Context, f func(seqID id.Sequence) (stop bool, err error)) (err error) {
	if err = pgs.writeCache(ctx); err != nil {
		return err
	}
	return pgs.underlyingMap.IterAll(ctx, func(k string, _ hash.Hash) error {
		seqID := id.Sequence(k)
		stop, err := f(seqID)
		if err != nil {
			return err
		} else if stop {
			return io.EOF
		} else {
			return nil
		}
	})
}

// IterateSequences iterates over all sequences in the collection.
func (pgs *Collection) IterateSequences(ctx context.Context, f func(seq *Sequence) (stop bool, err error)) (err error) {
	// For now, this function isn't used in a critical path, so we're not too worried about performance
	if err = pgs.cacheAllSequences(ctx); err != nil {
		return err
	}
	for _, seq := range pgs.accessedMap {
		if stop, err := f(seq); err != nil {
			return err
		} else if stop {
			break
		}
	}
	return nil
}

// NextVal returns the next value in the sequence.
func (pgs *Collection) NextVal(ctx context.Context, name id.Sequence) (int64, error) {
	seq, err := pgs.getSequence(ctx, name)
	if err != nil {
		return 0, err
	}
	if seq == nil {
		return 0, errors.Errorf(`relation "%s" does not exist`, name.SequenceName())
	}
	return seq.nextValForSequence()
}

// SetVal sets the sequence to the
func (pgs *Collection) SetVal(ctx context.Context, name id.Sequence, newValue int64, autoAdvance bool) error {
	seq, err := pgs.getSequence(ctx, name)
	if err != nil {
		return err
	}
	if seq == nil {
		return errors.Errorf(`relation "%s" does not exist`, name.SequenceName())
	}
	if newValue < seq.Minimum || newValue > seq.Maximum {
		return errors.Errorf(`setval: value %d is out of bounds for sequence "%s" (%d..%d)`,
			newValue, name, seq.Minimum, seq.Maximum)
	}
	seq.Current = newValue
	seq.IsAtEnd = false
	if autoAdvance {
		_, err := seq.nextValForSequence()
		return err
	}
	return nil
}

// Clone returns a new *Collection with the same contents as the original.
func (pgs *Collection) Clone(ctx context.Context) *Collection {
	newCollection := &Collection{
		accessedMap:   make(map[id.Sequence]*Sequence),
		underlyingMap: pgs.underlyingMap,
		ns:            pgs.ns,
	}
	for seqID, seq := range pgs.accessedMap {
		newCollection.accessedMap[seqID] = seq
	}
	return newCollection
}

// Map writes any cached sequences to the underlying map, and then returns the underlying map.
func (pgs *Collection) Map(ctx context.Context) (prolly.AddressMap, error) {
	if err := pgs.writeCache(ctx); err != nil {
		return prolly.AddressMap{}, err
	}
	return pgs.underlyingMap, nil
}

// GetID implements the interface objinterface.RootObject.
func (sequence *Sequence) GetID() id.Id {
	return sequence.Id.AsId()
}

// GetRootObjectID implements the interface objinterface.RootObject.
func (sequence *Sequence) GetRootObjectID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Sequences
}

// HashOf implements the interface rootobject.RootObject.
func (sequence *Sequence) HashOf(ctx context.Context) (hash.Hash, error) {
	data, err := sequence.Serialize(ctx)
	if err != nil {
		return hash.Hash{}, err
	}
	return hash.Of(data), nil
}

// Name implements the interface rootobject.RootObject.
func (sequence *Sequence) Name() doltdb.TableName {
	return doltdb.TableName{
		Name:   sequence.Id.SequenceName(),
		Schema: sequence.Id.SchemaName(),
	}
}

// cacheAllSequences loads every sequence from the Dolt map into our local map. This exists to simplify any iteration
// logic, and shouldn't be used on a performance-critical path.
func (pgs *Collection) cacheAllSequences(ctx context.Context) error {
	found := make(map[id.Sequence]struct{})
	for seqID := range pgs.accessedMap {
		found[seqID] = struct{}{}
	}
	return pgs.underlyingMap.IterAll(ctx, func(k string, v hash.Hash) error {
		seqID := id.Sequence(k)
		if _, ok := found[seqID]; ok {
			return nil
		}
		found[seqID] = struct{}{}
		data, err := pgs.ns.ReadBytes(ctx, v)
		if err != nil {
			return err
		}
		seq, err := DeserializeSequence(ctx, data)
		if err != nil {
			return err
		}
		pgs.accessedMap[seq.Id] = seq
		return nil
	})
}

// getSequence gets the sequence matching the given name.
func (pgs *Collection) getSequence(ctx context.Context, name id.Sequence) (*Sequence, error) {
	// Subsequent loads are cached
	if seq, ok := pgs.accessedMap[name]; ok {
		return seq, nil
	}
	// The initial load is from the internal map
	h, err := pgs.underlyingMap.Get(ctx, string(name))
	if err != nil || h.IsEmpty() {
		return nil, err
	}
	data, err := pgs.ns.ReadBytes(ctx, h)
	if err != nil {
		return nil, err
	}
	seq, err := DeserializeSequence(ctx, data)
	if err != nil {
		return nil, err
	}
	pgs.accessedMap[seq.Id] = seq
	return seq, nil
}

// writeCache writes every Sequence in the cache to the underlying map.
func (pgs *Collection) writeCache(ctx context.Context) (err error) {
	if len(pgs.accessedMap) == 0 {
		return nil
	}
	mapEditor := pgs.underlyingMap.Editor()
	for _, seq := range pgs.accessedMap {
		data, err := seq.Serialize(ctx)
		if err != nil {
			return err
		}
		h, err := pgs.ns.WriteBytes(ctx, data)
		if err != nil {
			return err
		}
		if err = mapEditor.Update(ctx, string(seq.Id), h); err != nil {
			return err
		}
	}
	pgs.underlyingMap, err = mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	clear(pgs.accessedMap)
	return nil
}

// nextValForSequence increments the calling sequence.
func (sequence *Sequence) nextValForSequence() (int64, error) {
	// First we'll check if we've reached the end, and cycle or error as necessary
	if sequence.IsAtEnd {
		if !sequence.Cycle {
			if sequence.Increment > 0 {
				return 0, errors.Errorf(`nextval: reached maximum value of sequence "%s" (%d)`, sequence.Id, sequence.Maximum)
			} else {
				return 0, errors.Errorf(`nextval: reached minimum value of sequence "%s" (%d)`, sequence.Id, sequence.Minimum)
			}
		}
		sequence.IsAtEnd = false
		if sequence.Increment > 0 {
			sequence.Current = sequence.Minimum
		} else {
			sequence.Current = sequence.Maximum
		}
	}
	// We'll return the current value, so everything after this sets the value for the next call
	valueToReturn := sequence.Current
	// Increment the current value
	if sequence.Increment > 0 {
		// Check for overflow or crossing the maximum, meaning we're at the end
		if sequence.Current > math.MaxInt64-sequence.Increment || sequence.Current+sequence.Increment > sequence.Maximum {
			sequence.IsAtEnd = true
		} else {
			sequence.Current += sequence.Increment
		}
	} else {
		// Check for underflow or crossing the minimum, meaning we're at the end
		if sequence.Current < math.MinInt64-sequence.Increment || sequence.Current+sequence.Increment < sequence.Minimum {
			sequence.IsAtEnd = true
		} else {
			sequence.Current += sequence.Increment
		}
	}
	return valueToReturn, nil
}
