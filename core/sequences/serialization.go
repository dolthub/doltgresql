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
	"sync"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/utils"
)

// Serialize returns the Collection as a byte slice. If the Collection is nil, then this returns a nil slice.
func (pgs *Collection) Serialize(ctx context.Context) ([]byte, error) {
	if pgs == nil {
		return nil, nil
	}
	pgs.mutex.Lock()
	defer pgs.mutex.Unlock()

	// Write all of the sequences to the writer
	writer := utils.NewWriter(256)
	writer.VariableUint(0) // Version
	schemaMapKeys := utils.GetMapKeysSorted(pgs.schemaMap)
	writer.VariableUint(uint64(len(schemaMapKeys)))
	for _, schemaMapKey := range schemaMapKeys {
		nameMap := pgs.schemaMap[schemaMapKey]
		writer.String(schemaMapKey)
		nameMapKeys := utils.GetMapKeysSorted(nameMap)
		writer.VariableUint(uint64(len(nameMapKeys)))
		for _, nameMapKey := range nameMapKeys {
			sequence := nameMap[nameMapKey]
			writer.Internal(sequence.Name.Internal())
			writer.Internal(sequence.DataTypeID.Internal())
			writer.Uint8(uint8(sequence.Persistence))
			writer.Int64(sequence.Start)
			writer.Int64(sequence.Current)
			writer.Int64(sequence.Increment)
			writer.Int64(sequence.Minimum)
			writer.Int64(sequence.Maximum)
			writer.Int64(sequence.Cache)
			writer.Bool(sequence.Cycle)
			writer.Bool(sequence.IsAtEnd)
			writer.Internal(sequence.OwnerTable.Internal())
			writer.String(sequence.OwnerColumn)
		}
	}

	return writer.Data(), nil
}

// Deserialize returns the Collection that was serialized in the byte slice. Returns an empty Collection if data is nil
// or empty.
func Deserialize(ctx context.Context, data []byte) (*Collection, error) {
	if len(data) == 0 {
		return &Collection{
			schemaMap: make(map[string]map[string]*Sequence),
			mutex:     &sync.Mutex{},
		}, nil
	}
	schemaMap := make(map[string]map[string]*Sequence)
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	if version != 0 {
		return nil, fmt.Errorf("version %d of sequences is not supported, please upgrade the server", version)
	}

	// Read from the reader
	numOfSchemas := reader.VariableUint()
	for i := uint64(0); i < numOfSchemas; i++ {
		schemaName := reader.String()
		numOfSequences := reader.VariableUint()
		nameMap := make(map[string]*Sequence)
		for j := uint64(0); j < numOfSequences; j++ {
			sequence := &Sequence{}
			sequence.Name = id.InternalSequence(reader.Internal())
			sequence.DataTypeID = id.InternalType(reader.Internal())
			sequence.Persistence = Persistence(reader.Uint8())
			sequence.Start = reader.Int64()
			sequence.Current = reader.Int64()
			sequence.Increment = reader.Int64()
			sequence.Minimum = reader.Int64()
			sequence.Maximum = reader.Int64()
			sequence.Cache = reader.Int64()
			sequence.Cycle = reader.Bool()
			sequence.IsAtEnd = reader.Bool()
			sequence.OwnerTable = id.InternalTable(reader.Internal())
			sequence.OwnerColumn = reader.String()
			nameMap[sequence.Name.SequenceName()] = sequence
		}
		schemaMap[schemaName] = nameMap
	}
	if !reader.IsEmpty() {
		return nil, fmt.Errorf("extra data found while deserializing sequences")
	}

	// Return the deserialized object
	return &Collection{
		schemaMap: schemaMap,
		mutex:     &sync.Mutex{},
	}, nil
}
