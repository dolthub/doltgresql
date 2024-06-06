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
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
)

// Collection contains a collection of sequences.
type Collection struct {
	schemaMap map[string]map[string]*Sequence
	mutex     *sync.Mutex
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
	Name        string
	DataTypeOID uint32
	Persistence Persistence
	Start       int64
	Current     int64
	Increment   int64
	Minimum     int64
	Maximum     int64
	Cache       int64
	Cycle       bool
	IsAtEnd     bool
	OwnerUser   string
	OwnerTable  string
	OwnerColumn string
}

// GetSequence returns the sequence with the given schema and name. Returns nil if the sequence cannot be found.
func (pgs *Collection) GetSequence(name doltdb.TableName) *Sequence {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	if nameMap, ok := pgs.schemaMap[name.Schema]; ok {
		if seq, ok := nameMap[name.Name]; ok {
			return seq
		}
	}
	return nil
}

// GetSequencesWithTable returns all sequences with the given table as the owner.
func (pgs *Collection) GetSequencesWithTable(name doltdb.TableName) []*Sequence {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	if nameMap, ok := pgs.schemaMap[name.Schema]; ok {
		var seqs []*Sequence
		for _, seq := range nameMap {
			if seq.OwnerTable == name.Name {
				seqs = append(seqs, seq)
			}
		}
		return seqs
	}
	return nil
}

// GetAllSequences returns a map containing all sequences in the collection, grouped by the schema they're contained in.
// Each sequence array is also sorted by the sequence name.
func (pgs *Collection) GetAllSequences() (sequences map[string][]*Sequence, schemaNames []string, totalCount int) {
	sequences = make(map[string][]*Sequence)
	for schemaName, nameMap := range pgs.schemaMap {
		schemaNames = append(schemaNames, schemaName)
		seqs := make([]*Sequence, 0, len(nameMap))
		for _, seq := range nameMap {
			seqs = append(seqs, seq)
		}
		totalCount += len(seqs)
		sort.Slice(seqs, func(i, j int) bool {
			return seqs[i].Name < seqs[j].Name
		})
		sequences[schemaName] = seqs
	}
	sort.Slice(schemaNames, func(i, j int) bool {
		return schemaNames[i] < schemaNames[j]
	})
	return
}

// HasSequence returns whether the sequence is present.
func (pgs *Collection) HasSequence(name doltdb.TableName) bool {
	return pgs.GetSequence(name) != nil
}

// CreateSequence creates a new sequence.
func (pgs *Collection) CreateSequence(schema string, seq *Sequence) error {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	nameMap, ok := pgs.schemaMap[schema]
	if !ok {
		nameMap = make(map[string]*Sequence)
		pgs.schemaMap[schema] = nameMap
	}
	if _, ok = nameMap[seq.Name]; ok {
		return fmt.Errorf(`relation "%s" already exists`, seq.Name)
	}
	nameMap[seq.Name] = seq
	return nil
}

// DropSequence drops an existing sequence.
func (pgs *Collection) DropSequence(name doltdb.TableName) error {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	if nameMap, ok := pgs.schemaMap[name.Schema]; ok {
		if _, ok = nameMap[name.Name]; ok {
			delete(nameMap, name.Name)
			return nil
		}
	}
	return fmt.Errorf(`sequence "%s" does not exist`, name)
}

// IterateSequences iterates over all sequences in the collection.
func (pgs *Collection) IterateSequences(f func(schema string, seq *Sequence) error) error {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	for schema, nameMap := range pgs.schemaMap {
		for _, seq := range nameMap {
			if err := f(schema, seq); err != nil {
				return err
			}
		}
	}
	return nil
}

// NextVal returns the next value in the sequence.
func (pgs *Collection) NextVal(schema, name string) (int64, error) {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	if nameMap, ok := pgs.schemaMap[schema]; ok {
		if seq, ok := nameMap[name]; ok {
			return seq.nextValForSequence()
		}
	}
	return 0, fmt.Errorf(`relation "%s" does not exist`, name)
}

// SetVal sets the sequence to the
func (pgs *Collection) SetVal(schema, name string, newValue int64, autoAdvance bool) error {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	if nameMap, ok := pgs.schemaMap[schema]; ok {
		if seq, ok := nameMap[name]; ok {
			if newValue < seq.Minimum || newValue > seq.Maximum {
				return fmt.Errorf(`setval: value %d is out of bounds for sequence "%s" (%d..%d)`,
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
	}
	return fmt.Errorf(`relation "%s" does not exist`, name)
}

// Clone returns a new *Collection with the same contents as the original.
func (pgs *Collection) Clone() *Collection {
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	newCollection := &Collection{
		schemaMap: make(map[string]map[string]*Sequence),
		mutex:     &sync.Mutex{},
	}
	for schema, nameMap := range pgs.schemaMap {
		if len(nameMap) == 0 {
			continue
		}
		clonedNameMap := make(map[string]*Sequence)
		for key, seq := range nameMap {
			newSeq := *seq
			clonedNameMap[key] = &newSeq
		}
		newCollection.schemaMap[schema] = clonedNameMap
	}
	return newCollection
}

// nextValForSequence increments the calling sequence. Called from other functions that hold locks.
func (sequence *Sequence) nextValForSequence() (int64, error) {
	// First we'll check if we've reached the end, and cycle or error as necessary
	if sequence.IsAtEnd {
		if !sequence.Cycle {
			if sequence.Increment > 0 {
				return 0, fmt.Errorf(`nextval: reached maximum value of sequence "%s" (%d)`, sequence.Name, sequence.Maximum)
			} else {
				return 0, fmt.Errorf(`nextval: reached minimum value of sequence "%s" (%d)`, sequence.Name, sequence.Minimum)
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
