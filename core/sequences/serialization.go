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

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/utils"
)

// Serialize returns the Sequence as a byte slice. If the Sequence is nil, then this returns a nil slice.
func (sequence *Sequence) Serialize(ctx context.Context) ([]byte, error) {
	if sequence == nil {
		return nil, nil
	}

	// Create the writer
	writer := utils.NewWriter(256)
	writer.VariableUint(0) // Version
	// Write the sequence data
	writer.Id(sequence.Id.AsId())
	writer.Id(sequence.DataTypeID.AsId())
	writer.Uint8(uint8(sequence.Persistence))
	writer.Int64(sequence.Start)
	writer.Int64(sequence.Current)
	writer.Int64(sequence.Increment)
	writer.Int64(sequence.Minimum)
	writer.Int64(sequence.Maximum)
	writer.Int64(sequence.Cache)
	writer.Bool(sequence.Cycle)
	writer.Bool(sequence.IsAtEnd)
	writer.Id(sequence.OwnerTable.AsId())
	writer.String(sequence.OwnerColumn)
	// Returns the data
	return writer.Data(), nil
}

// DeserializeSequence returns the Sequence that was serialized in the byte slice. Returns an empty Sequence if data is
// nil or empty.
func DeserializeSequence(ctx context.Context, data []byte) (*Sequence, error) {
	if len(data) == 0 {
		return nil, nil
	}
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	if version != 0 {
		return nil, errors.Errorf("version %d of sequences is not supported, please upgrade the server", version)
	}

	// Read from the reader
	sequence := &Sequence{}
	sequence.Id = id.Sequence(reader.Id())
	sequence.DataTypeID = id.Type(reader.Id())
	sequence.Persistence = Persistence(reader.Uint8())
	sequence.Start = reader.Int64()
	sequence.Current = reader.Int64()
	sequence.Increment = reader.Int64()
	sequence.Minimum = reader.Int64()
	sequence.Maximum = reader.Int64()
	sequence.Cache = reader.Int64()
	sequence.Cycle = reader.Bool()
	sequence.IsAtEnd = reader.Bool()
	sequence.OwnerTable = id.Table(reader.Id())
	sequence.OwnerColumn = reader.String()
	if !reader.IsEmpty() {
		return nil, errors.Errorf("extra data found while deserializing a sequence")
	}
	// Return the deserialized object
	return sequence, nil
}
