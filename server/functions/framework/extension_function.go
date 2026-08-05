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

package framework

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ExtensionFunction is the implementation of functions that an extension provides, looked up in
// server/extensions by extension name and symbol.
type ExtensionFunction struct {
	ID                 id.Function
	ReturnType         *pgtypes.DoltgresType
	ParameterTypes     []*pgtypes.DoltgresType
	Variadic           bool
	IsNonDeterministic bool
	Strict             bool
	SetOf              bool
	ExtensionName      string
	ExtensionSymbol    string
}

var _ FunctionInterface = ExtensionFunction{}

// GetExpectedParameterCount implements the interface FunctionInterface.
func (extFunc ExtensionFunction) GetExpectedParameterCount() int {
	return len(extFunc.ParameterTypes)
}

// GetName implements the interface FunctionInterface.
func (extFunc ExtensionFunction) GetName() string {
	return extFunc.ID.FunctionName()
}

// GetOutParameters implements the interface FunctionInterface.
func (extFunc ExtensionFunction) GetOutParameters() sql.Schema {
	return nil
}

// GetInputParameterTypes implements the interface FunctionInterface.
func (extFunc ExtensionFunction) GetInputParameterTypes() []*pgtypes.DoltgresType {
	return extFunc.ParameterTypes
}

// GetReturn implements the interface FunctionInterface.
func (extFunc ExtensionFunction) GetReturn() *pgtypes.DoltgresType {
	return extFunc.ReturnType
}

// InternalID implements the interface FunctionInterface.
func (extFunc ExtensionFunction) InternalID() id.Id {
	return extFunc.ID.AsId()
}

// IsStrict implements the interface FunctionInterface.
func (extFunc ExtensionFunction) IsStrict() bool {
	return extFunc.Strict
}

// NonDeterministic implements the interface FunctionInterface.
func (extFunc ExtensionFunction) NonDeterministic() bool {
	return extFunc.IsNonDeterministic
}

// IsCVariadic implements the FunctionInterface interface.
func (extFunc ExtensionFunction) IsCVariadic() bool {
	// TODO: implement c-language variadic
	return false
}

// VariadicIndex implements the interface FunctionInterface.
func (extFunc ExtensionFunction) VariadicIndex() int {
	// TODO: implement variadic
	return -1
}

// IsSRF implements the interface FunctionInterface.
func (extFunc ExtensionFunction) IsSRF() bool {
	return extFunc.SetOf
}

// enforceInterfaceInheritance implements the interface FunctionInterface.
func (extFunc ExtensionFunction) enforceInterfaceInheritance(error) {}
