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

	paramNames := make([]string, len(procedure.AllParams))
	paramTypes := make([]id.Type, len(procedure.AllParams))
	paramModes := make([]ParameterMode, len(procedure.AllParams))
	paramDefaults := make([]string, len(procedure.AllParams))
	for i, param := range procedure.AllParams {
		paramNames[i] = param.Name
		paramTypes[i] = param.Type
		paramDefaults[i] = param.Default
		paramModes[i] = param.Mode
	}

	// Write all of the procedures to the writer
	writer := utils.NewWriter(256)
	writer.VariableUint(1) // Version
	// Write the procedure data
	writer.Id(procedure.ID.AsId())
	writer.StringSlice(paramNames)
	writer.IdTypeSlice(paramTypes)
	writer.String(procedure.Definition)
	writer.String(procedure.ExtensionName)
	writer.String(procedure.ExtensionSymbol)
	writer.String(procedure.SQLDefinition)
	//Write the parameter modes
	writer.VariableUint(uint64(len(paramModes)))
	for _, mode := range paramModes {
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
	// Write version 1 data
	writer.StringSlice(paramDefaults)
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
	if version > 1 {
		return Procedure{}, errors.Errorf("version %d of procedures is not supported, please upgrade the server", version)
	}

	// Read from the reader
	p := Procedure{}
	p.ID = id.Procedure(reader.Id())
	paramNames := reader.StringSlice()
	paramTypes := reader.IdTypeSlice()
	p.Definition = reader.String()
	p.ExtensionName = reader.String()
	p.ExtensionSymbol = reader.String()
	p.SQLDefinition = reader.String()
	// Read the parameter modes
	modeCount := reader.VariableUint()
	paramModes := make([]ParameterMode, modeCount)
	for modeIdx := uint64(0); modeIdx < modeCount; modeIdx++ {
		paramModes[modeIdx] = ParameterMode(reader.Uint8())
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
	var paramDefaults []string
	if version >= 1 {
		paramDefaults = reader.StringSlice()
	}
	p.AllParams = make([]Parameter, len(paramNames))
	for i, paramName := range paramNames {
		p.AllParams[i] = Parameter{
			Name: paramName,
			Type: paramTypes[i],
			Mode: paramModes[i],
		}
		if paramDefaults != nil {
			p.AllParams[i].Default = paramDefaults[i]
		}
	}
	if !reader.IsEmpty() {
		return Procedure{}, errors.New("extra data found while deserializing a procedure")
	}
	// Return the deserialized object
	return p, nil
}
