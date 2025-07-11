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

package framework

import (
	"github.com/dolthub/doltgresql/core/extensions"
	"github.com/dolthub/doltgresql/core/id"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// CFunction is the implementation of functions that host their logic in a shared library.
type CFunction struct {
	ID                 id.Function
	ReturnType         *pgtypes.DoltgresType
	ParameterTypes     []*pgtypes.DoltgresType
	Variadic           bool
	IsNonDeterministic bool
	Strict             bool
	ExtensionName      extensions.LibraryIdentifier
	ExtensionSymbol    string
}

var _ FunctionInterface = CFunction{}

// GetExpectedParameterCount implements the interface FunctionInterface.
func (cFunc CFunction) GetExpectedParameterCount() int {
	return len(cFunc.ParameterTypes)
}

// GetName implements the interface FunctionInterface.
func (cFunc CFunction) GetName() string {
	return cFunc.ID.FunctionName()
}

// GetParameters implements the interface FunctionInterface.
func (cFunc CFunction) GetParameters() []*pgtypes.DoltgresType {
	return cFunc.ParameterTypes
}

// GetReturn implements the interface FunctionInterface.
func (cFunc CFunction) GetReturn() *pgtypes.DoltgresType {
	return cFunc.ReturnType
}

// InternalID implements the interface FunctionInterface.
func (cFunc CFunction) InternalID() id.Id {
	return cFunc.ID.AsId()
}

// IsStrict implements the interface FunctionInterface.
func (cFunc CFunction) IsStrict() bool {
	return cFunc.Strict
}

// NonDeterministic implements the interface FunctionInterface.
func (cFunc CFunction) NonDeterministic() bool {
	return cFunc.IsNonDeterministic
}

// VariadicIndex implements the interface FunctionInterface.
func (cFunc CFunction) VariadicIndex() int {
	// TODO: implement variadic
	return -1
}

// ISRF implements the interface FunctionInterface.
func (cFunc CFunction) IsSRF() bool {
	return false
}

// enforceInterfaceInheritance implements the interface FunctionInterface.
func (cFunc CFunction) enforceInterfaceInheritance(error) {}
