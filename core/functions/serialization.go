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

package functions

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/plpgsql"
	"github.com/dolthub/doltgresql/utils"
)

// Serialize returns the Function as a byte slice. If the Function is invalid, then this returns a nil slice.
func (function Function) Serialize(ctx context.Context) ([]byte, error) {
	if !function.ID.IsValid() {
		return nil, nil
	}

	// Write all of the functions to the writer
	writer := utils.NewWriter(256)
	writer.VariableUint(1) // Version
	// Write the function data
	writer.Id(function.ID.AsId())
	writer.Id(function.ReturnType.AsId())
	writer.StringSlice(function.ParameterNames)
	writer.IdTypeSlice(function.ParameterTypes)
	writer.Bool(function.Variadic)
	writer.Bool(function.IsNonDeterministic)
	writer.Bool(function.Strict)
	writer.String(function.Definition)
	// Write the operations
	writer.VariableUint(uint64(len(function.Operations)))
	for _, op := range function.Operations {
		writer.Uint16(uint16(op.OpCode))
		writer.String(op.PrimaryData)
		writer.StringSlice(op.SecondaryData)
		writer.String(op.Target)
		writer.Int32(int32(op.Index))
		writer.StringMap(op.Options)
	}
	// Write version 1 data
	writer.String(function.ExtensionName)
	writer.String(function.ExtensionSymbol)
	writer.String(function.SQLDefinition)
	writer.Bool(function.SetOf)
	// Returns the data
	return writer.Data(), nil
}

// DeserializeFunction returns the Function that was serialized in the byte slice. Returns an empty Function (invalid
// ID) if data is nil or empty.
func DeserializeFunction(ctx context.Context, data []byte) (Function, error) {
	if len(data) == 0 {
		return Function{}, nil
	}
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	if version > 1 {
		return Function{}, errors.Errorf("version %d of functions is not supported, please upgrade the server", version)
	}

	// Read from the reader
	f := Function{}
	f.ID = id.Function(reader.Id())
	f.ReturnType = id.Type(reader.Id())
	f.ParameterNames = reader.StringSlice()
	f.ParameterTypes = reader.IdTypeSlice()
	f.Variadic = reader.Bool()
	f.IsNonDeterministic = reader.Bool()
	f.Strict = reader.Bool()
	f.Definition = reader.String()
	// Read the operations
	opCount := reader.VariableUint()
	f.Operations = make([]plpgsql.InterpreterOperation, opCount)
	for opIdx := uint64(0); opIdx < opCount; opIdx++ {
		op := plpgsql.InterpreterOperation{}
		op.OpCode = plpgsql.OpCode(reader.Uint16())
		op.PrimaryData = reader.String()
		op.SecondaryData = reader.StringSlice()
		op.Target = reader.String()
		op.Index = int(reader.Int32())
		op.Options = reader.StringMap()
		f.Operations[opIdx] = op
	}
	if version == 1 {
		f.ExtensionName = reader.String()
		f.ExtensionSymbol = reader.String()
		f.SQLDefinition = reader.String()
		f.SetOf = reader.Bool()
	}
	if !reader.IsEmpty() {
		return Function{}, errors.Errorf("extra data found while deserializing a function")
	}
	// Return the deserialized object
	return f, nil
}
