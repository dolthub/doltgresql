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
	"sync"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/interpreter"
	"github.com/dolthub/doltgresql/utils"
)

// Serialize returns the Collection as a byte slice. If the Collection is nil, then this returns a nil slice.
func (pgf *Collection) Serialize(ctx context.Context) ([]byte, error) {
	if pgf == nil {
		return nil, nil
	}
	pgf.mutex.Lock()
	defer pgf.mutex.Unlock()

	// Write all of the functions to the writer
	writer := utils.NewWriter(256)
	writer.VariableUint(0) // Version
	funcIDs := utils.GetMapKeysSorted(pgf.funcMap)
	writer.VariableUint(uint64(len(funcIDs)))
	for _, funcID := range funcIDs {
		f := pgf.funcMap[funcID]
		writer.Id(f.ID.AsId())
		writer.Id(f.ReturnType.AsId())
		writer.StringSlice(f.ParameterNames)
		writer.IdTypeSlice(f.ParameterTypes)
		writer.Bool(f.Variadic)
		writer.Bool(f.IsNonDeterministic)
		writer.Bool(f.Strict)
		// Write the operations
		writer.VariableUint(uint64(len(f.Operations)))
		for _, op := range f.Operations {
			writer.Uint16(uint16(op.OpCode))
			writer.String(op.PrimaryData)
			writer.StringSlice(op.SecondaryData)
			writer.String(op.Target)
			writer.Int32(int32(op.Index))
		}
	}

	return writer.Data(), nil
}

// Deserialize returns the Collection that was serialized in the byte slice. Returns an empty Collection if data is nil
// or empty.
func Deserialize(ctx context.Context, data []byte) (*Collection, error) {
	if len(data) == 0 {
		return &Collection{
			funcMap:     make(map[id.Function]*Function),
			overloadMap: make(map[id.Function][]*Function),
			mutex:       &sync.Mutex{},
		}, nil
	}
	funcMap := make(map[id.Function]*Function)
	overloadMap := make(map[id.Function][]*Function)
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	if version != 0 {
		return nil, errors.Errorf("version %d of functions is not supported, please upgrade the server", version)
	}

	// Read from the reader
	numOfFunctions := reader.VariableUint()
	for i := uint64(0); i < numOfFunctions; i++ {
		f := &Function{}
		f.ID = id.Function(reader.Id())
		f.ReturnType = id.Type(reader.Id())
		f.ParameterNames = reader.StringSlice()
		f.ParameterTypes = reader.IdTypeSlice()
		f.Variadic = reader.Bool()
		f.IsNonDeterministic = reader.Bool()
		f.Strict = reader.Bool()
		// Read the operations
		opCount := reader.VariableUint()
		f.Operations = make([]interpreter.InterpreterOperation, opCount)
		for opIdx := uint64(0); opIdx < opCount; opIdx++ {
			op := interpreter.InterpreterOperation{}
			op.OpCode = interpreter.OpCode(reader.Uint16())
			op.PrimaryData = reader.String()
			op.SecondaryData = reader.StringSlice()
			op.Target = reader.String()
			op.Index = int(reader.Int32())
			f.Operations[opIdx] = op
		}
		// Add the function to each map
		funcMap[f.ID] = f
		funcNameOnly := id.NewFunction(f.ID.SchemaName(), f.ID.FunctionName())
		overloadMap[funcNameOnly] = append(overloadMap[funcNameOnly], f)
	}
	if !reader.IsEmpty() {
		return nil, errors.Errorf("extra data found while deserializing functions")
	}

	// Return the deserialized object
	return &Collection{
		funcMap:     funcMap,
		overloadMap: overloadMap,
		mutex:       &sync.Mutex{},
	}, nil
}
