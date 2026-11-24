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

package node

import (
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/procedures"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

// validateFunctionReplacement returns an error if CREATE OR REPLACE FUNCTION would change a property of the existing
// function that Postgres does not allow to be changed in place. This mirrors the checks in Postgres' ProcedureCreate.
func validateFunctionReplacement(oldRetType id.Type, oldSetOf bool, oldParams []procedures.Parameter, newRetType id.Type, newSetOf bool, newParams []procedures.Parameter) error {
	if oldRetType != newRetType || oldSetOf != newSetOf {
		return pgerror.Newf(pgcode.InvalidFunctionDefinition, "cannot change return type of existing function")
	}
	if !equalOutputParams(oldParams, newParams) {
		return pgerror.Newf(pgcode.InvalidFunctionDefinition, "cannot change return type of existing function")
	}
	return validateInputParamReplacement(oldParams, newParams)
}

// validateProcedureReplacement returns an error if CREATE OR REPLACE PROCEDURE would change a property of the existing
// procedure that Postgres does not allow to be changed in place. This mirrors the checks in Postgres' ProcedureCreate.
func validateProcedureReplacement(oldParams []procedures.Parameter, newParams []procedures.Parameter) error {
	if hasOutputParams(oldParams) != hasOutputParams(newParams) {
		return pgerror.Newf(pgcode.InvalidFunctionDefinition, "cannot change whether a procedure has output parameters")
	}
	if !equalOutputParams(oldParams, newParams) {
		return pgerror.Newf(pgcode.InvalidFunctionDefinition, "cannot change return type of existing function")
	}
	return validateInputParamReplacement(oldParams, newParams)
}

// validateInputParamReplacement checks that input parameter names are unchanged and that no defaults are removed.
// The input parameter types are guaranteed to match, since they're part of the routine's identity.
func validateInputParamReplacement(oldParams []procedures.Parameter, newParams []procedures.Parameter) error {
	oldInputs := inputParams(oldParams)
	newInputs := inputParams(newParams)
	for i := range oldInputs {
		if i >= len(newInputs) {
			break
		}
		if len(oldInputs[i].Name) > 0 && oldInputs[i].Name != newInputs[i].Name {
			return pgerror.Newf(pgcode.InvalidFunctionDefinition, `cannot change name of input parameter "%s"`, oldInputs[i].Name)
		}
	}
	if countDefaults(oldInputs) > countDefaults(newInputs) {
		return pgerror.Newf(pgcode.InvalidFunctionDefinition, "cannot remove parameter defaults from existing function")
	}
	return nil
}

// inputParams returns the parameters that are passed in by the caller.
func inputParams(params []procedures.Parameter) []procedures.Parameter {
	var inputs []procedures.Parameter
	for _, param := range params {
		if param.Mode != procedures.ParameterMode_OUT {
			inputs = append(inputs, param)
		}
	}
	return inputs
}

// outputParams returns the parameters that are returned to the caller.
func outputParams(params []procedures.Parameter) []procedures.Parameter {
	var outputs []procedures.Parameter
	for _, param := range params {
		if param.Mode == procedures.ParameterMode_OUT || param.Mode == procedures.ParameterMode_INOUT {
			outputs = append(outputs, param)
		}
	}
	return outputs
}

// hasOutputParams returns whether any parameter is returned to the caller.
func hasOutputParams(params []procedures.Parameter) bool {
	return len(outputParams(params)) > 0
}

// equalOutputParams returns whether the row types defined by the output parameters have the same names and types.
func equalOutputParams(oldParams []procedures.Parameter, newParams []procedures.Parameter) bool {
	oldOutputs := outputParams(oldParams)
	newOutputs := outputParams(newParams)
	if len(oldOutputs) != len(newOutputs) {
		return false
	}
	for i := range oldOutputs {
		if oldOutputs[i].Name != newOutputs[i].Name || oldOutputs[i].Type != newOutputs[i].Type {
			return false
		}
	}
	return true
}

// countDefaults returns the number of parameters that have a default value.
func countDefaults(params []procedures.Parameter) int {
	count := 0
	for _, param := range params {
		if len(param.Default) > 0 {
			count++
		}
	}
	return count
}
