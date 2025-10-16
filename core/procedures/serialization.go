// Copyright 2025 Dolthub, Inc.
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

package procedures

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/plpgsql"
	"github.com/dolthub/doltgresql/utils"
)

// Serialize returns the Procedure as a byte slice. If the Procedure is invalid, then this returns a nil slice.
func (procedure Procedure) Serialize(ctx context.Context) ([]byte, error) {
	if !procedure.ID.IsValid() {
		return nil, nil
	}

	// Write all of the procedures to the writer
	writer := utils.NewWriter(256)
	writer.VariableUint(0) // Version
	// Write the procedure data
	writer.Id(procedure.ID.AsId())
	writer.StringSlice(procedure.ParameterNames)
	writer.IdTypeSlice(procedure.ParameterTypes)
	writer.String(procedure.Definition)
	writer.String(procedure.ExtensionName)
	writer.String(procedure.ExtensionSymbol)
	writer.String(procedure.SQLDefinition)
	// Write the parameter modes
	writer.VariableUint(uint64(len(procedure.ParameterModes)))
	for _, mode := range procedure.ParameterModes {
		writer.Uint8(uint8(mode))
	}
	// Write the operations
	writer.VariableUint(uint64(len(procedure.Operations)))
	for _, op := range procedure.Operations {
		writer.Uint16(uint16(op.OpCode))
		writer.String(op.PrimaryData)
		writer.StringSlice(op.SecondaryData)
		writer.String(op.Target)
		writer.Int32(int32(op.Index))
		writer.StringMap(op.Options)
	}
	// Returns the data
	return writer.Data(), nil
}

// DeserializeProcedure returns the Procedure that was serialized in the byte slice. Returns an empty Procedure (invalid
// ID) if data is nil or empty.
func DeserializeProcedure(ctx context.Context, data []byte) (Procedure, error) {
	if len(data) == 0 {
		return Procedure{}, nil
	}
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	if version > 0 {
		return Procedure{}, errors.Errorf("version %d of functions is not supported, please upgrade the server", version)
	}

	// Read from the reader
	p := Procedure{}
	p.ID = id.Procedure(reader.Id())
	p.ParameterNames = reader.StringSlice()
	p.ParameterTypes = reader.IdTypeSlice()
	p.Definition = reader.String()
	p.ExtensionName = reader.String()
	p.ExtensionSymbol = reader.String()
	p.SQLDefinition = reader.String()
	// Read the parameter modes
	modeCount := reader.VariableUint()
	p.ParameterModes = make([]ParameterMode, modeCount)
	for modeIdx := uint64(0); modeIdx < modeCount; modeIdx++ {
		p.ParameterModes[modeIdx] = ParameterMode(reader.Uint8())
	}
	// Read the operations
	opCount := reader.VariableUint()
	p.Operations = make([]plpgsql.InterpreterOperation, opCount)
	for opIdx := uint64(0); opIdx < opCount; opIdx++ {
		op := plpgsql.InterpreterOperation{}
		op.OpCode = plpgsql.OpCode(reader.Uint16())
		op.PrimaryData = reader.String()
		op.SecondaryData = reader.StringSlice()
		op.Target = reader.String()
		op.Index = int(reader.Int32())
		op.Options = reader.StringMap()
		p.Operations[opIdx] = op
	}
	if !reader.IsEmpty() {
		return Procedure{}, errors.New("extra data found while deserializing a procedure")
	}
	// Return the deserialized object
	return p, nil
}
