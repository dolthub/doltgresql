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

package aggregates

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/utils"
)

// Serialize returns the Aggregate as a byte slice. If the Aggregate is invalid, then this returns a nil slice.
func (aggregate Aggregate) Serialize(ctx context.Context) ([]byte, error) {
	if !aggregate.ID.IsValid() {
		return nil, nil
	}

	// Initialize the writer and version
	writer := utils.NewWriter(256)
	writer.VariableUint(0) // Version
	// Write the aggregate data
	writer.Id(aggregate.ID.AsId())
	writer.Id(aggregate.ReturnType.AsId())
	writer.Id(aggregate.SFunc.AsId())
	writer.Id(aggregate.SType.AsId())
	writer.Id(aggregate.FinalFunc.AsId())
	writer.Id(aggregate.CombineFunc.AsId())
	writer.String(aggregate.InitCond)
	writer.Bool(aggregate.HasInitCond)
	// Returns the data
	return writer.Data(), nil
}

// DeserializeAggregate returns the Aggregate that was serialized in the byte slice. Returns an empty Aggregate
// (invalid ID) if data is nil or empty.
func DeserializeAggregate(ctx context.Context, data []byte) (Aggregate, error) {
	if len(data) == 0 {
		return Aggregate{}, nil
	}
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	if version != 0 {
		return Aggregate{}, errors.Errorf("version %d of aggregates is not supported, please upgrade the server", version)
	}

	// Read from the reader
	a := Aggregate{}
	a.ID = id.Function(reader.Id())
	a.ReturnType = id.Type(reader.Id())
	a.SFunc = id.Function(reader.Id())
	a.SType = id.Type(reader.Id())
	a.FinalFunc = id.Function(reader.Id())
	a.CombineFunc = id.Function(reader.Id())
	a.InitCond = reader.String()
	a.HasInitCond = reader.Bool()
	if !reader.IsEmpty() {
		return Aggregate{}, errors.Errorf("extra data found while deserializing an aggregate")
	}
	// Return the deserialized object
	return a, nil
}
