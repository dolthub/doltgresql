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

package operators

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/utils"
)

// Serialize returns the Operator as a byte slice. If the Operator is invalid, then this returns a nil slice.
func (operator Operator) Serialize(ctx context.Context) ([]byte, error) {
	if !operator.ID.IsValid() {
		return nil, nil
	}

	// Initialize the writer and version
	writer := utils.NewWriter(256)
	writer.VariableUint(0) // Version
	// Write the operator data
	writer.Id(operator.ID.AsId())
	writer.Id(operator.Function.AsId())
	writer.Id(operator.ReturnType.AsId())
	writer.String(operator.Commutator)
	writer.String(operator.Negator)
	writer.Bool(operator.Hashes)
	writer.Bool(operator.Merges)
	// Returns the data
	return writer.Data(), nil
}

// DeserializeOperator returns the Operator that was serialized in the byte slice. Returns an empty Operator (invalid
// ID) if data is nil or empty.
func DeserializeOperator(ctx context.Context, data []byte) (Operator, error) {
	if len(data) == 0 {
		return Operator{}, nil
	}
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	if version != 0 {
		return Operator{}, errors.Errorf("version %d of operators is not supported, please upgrade the server", version)
	}

	// Read from the reader
	o := Operator{}
	o.ID = id.Operator(reader.Id())
	o.Function = id.Function(reader.Id())
	o.ReturnType = id.Type(reader.Id())
	o.Commutator = reader.String()
	o.Negator = reader.String()
	o.Hashes = reader.Bool()
	o.Merges = reader.Bool()
	if !reader.IsEmpty() {
		return Operator{}, errors.Errorf("extra data found while deserializing an operator")
	}
	// Return the deserialized object
	return o, nil
}
