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

package framework

import (
	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// FunctionInterface is an interface for PostgreSQL functions.
type FunctionInterface interface {
	// GetName returns the name of the function. The name is case-insensitive, so the casing does not matter.
	GetName() string
	// GetReturn returns the return type.
	GetReturn() pgtypes.DoltgresType
	// GetParameters returns the parameter types for the function.
	GetParameters() []pgtypes.DoltgresType
	// GetExpectedParameterCount returns the number of parameters that are valid for this function.
	GetExpectedParameterCount() int
	// GetIsNonDeterministic returns whether the function is non-deterministic.
	GetIsNonDeterministic() bool
	// GetIsStrict returns whether the function is STRICT, which means if any parameter is NULL, then it returns NULL.
	// Otherwise, if it's not, the NULL input must be handled by user.
	GetIsStrict() bool
	// enforceInterfaceInheritance is a special function that ensures only the expected types inherit this interface.
	enforceInterfaceInheritance(error)
}

// Function0 is a function that does not take any parameters.
type Function0 struct {
	Name               string
	Return             pgtypes.DoltgresType
	Parameters         []pgtypes.DoltgresType
	IsNonDeterministic bool
	Strict             bool
	Callable           func(ctx *sql.Context) (any, error)
}

// Function1 is a function that takes one parameter.
type Function1 struct {
	Name               string
	Return             pgtypes.DoltgresType
	Parameters         []pgtypes.DoltgresType
	IsNonDeterministic bool
	Strict             bool
	Callable           func(ctx *sql.Context, val1 any) (any, error)
}

// Function2 is a function that takes two parameters.
type Function2 struct {
	Name               string
	Return             pgtypes.DoltgresType
	Parameters         []pgtypes.DoltgresType
	IsNonDeterministic bool
	Strict             bool
	Callable           func(ctx *sql.Context, val1 any, val2 any) (any, error)
}

// Function3 is a function that takes three parameters.
type Function3 struct {
	Name               string
	Return             pgtypes.DoltgresType
	Parameters         []pgtypes.DoltgresType
	IsNonDeterministic bool
	Strict             bool
	Callable           func(ctx *sql.Context, val1 any, val2 any, val3 any) (any, error)
}

// Function4 is a function that takes four parameters.
type Function4 struct {
	Name               string
	Return             pgtypes.DoltgresType
	Parameters         []pgtypes.DoltgresType
	IsNonDeterministic bool
	Strict             bool
	Callable           func(ctx *sql.Context, val1 any, val2 any, val3 any, val4 any) (any, error)
}

var _ FunctionInterface = Function0{}
var _ FunctionInterface = Function1{}
var _ FunctionInterface = Function2{}
var _ FunctionInterface = Function3{}
var _ FunctionInterface = Function4{}

// GetName implements the FunctionInterface interface.
func (f Function0) GetName() string { return f.Name }

// GetReturn implements the FunctionInterface interface.
func (f Function0) GetReturn() pgtypes.DoltgresType { return f.Return }

// GetParameters implements the FunctionInterface interface.
func (f Function0) GetParameters() []pgtypes.DoltgresType { return f.Parameters }

// GetExpectedParameterCount implements the FunctionInterface interface.
func (f Function0) GetExpectedParameterCount() int { return 0 }

// GetIsNonDeterministic implements the FunctionInterface interface.
func (f Function0) GetIsNonDeterministic() bool { return f.IsNonDeterministic }

// GetIsStrict implements the FunctionInterface interface.
func (f Function0) GetIsStrict() bool { return f.Strict }

// enforceInterfaceInheritance implements the FunctionInterface interface.
func (f Function0) enforceInterfaceInheritance(error) {}

// GetName implements the FunctionInterface interface.
func (f Function1) GetName() string { return f.Name }

// GetReturn implements the FunctionInterface interface.
func (f Function1) GetReturn() pgtypes.DoltgresType { return f.Return }

// GetParameters implements the FunctionInterface interface.
func (f Function1) GetParameters() []pgtypes.DoltgresType { return f.Parameters }

// GetExpectedParameterCount implements the FunctionInterface interface.
func (f Function1) GetExpectedParameterCount() int { return 1 }

// GetIsNonDeterministic implements the FunctionInterface interface.
func (f Function1) GetIsNonDeterministic() bool { return f.IsNonDeterministic }

// GetIsStrict implements the FunctionInterface interface.
func (f Function1) GetIsStrict() bool { return f.Strict }

// enforceInterfaceInheritance implements the FunctionInterface interface.
func (f Function1) enforceInterfaceInheritance(error) {}

// GetName implements the FunctionInterface interface.
func (f Function2) GetName() string { return f.Name }

// GetReturn implements the FunctionInterface interface.
func (f Function2) GetReturn() pgtypes.DoltgresType { return f.Return }

// GetParameters implements the FunctionInterface interface.
func (f Function2) GetParameters() []pgtypes.DoltgresType { return f.Parameters }

// GetExpectedParameterCount implements the FunctionInterface interface.
func (f Function2) GetExpectedParameterCount() int { return 2 }

// GetIsNonDeterministic implements the FunctionInterface interface.
func (f Function2) GetIsNonDeterministic() bool { return f.IsNonDeterministic }

// GetIsStrict implements the FunctionInterface interface.
func (f Function2) GetIsStrict() bool { return f.Strict }

// enforceInterfaceInheritance implements the FunctionInterface interface.
func (f Function2) enforceInterfaceInheritance(error) {}

// GetName implements the FunctionInterface interface.
func (f Function3) GetName() string { return f.Name }

// GetReturn implements the FunctionInterface interface.
func (f Function3) GetReturn() pgtypes.DoltgresType { return f.Return }

// GetParameters implements the FunctionInterface interface.
func (f Function3) GetParameters() []pgtypes.DoltgresType { return f.Parameters }

// GetExpectedParameterCount implements the FunctionInterface interface.
func (f Function3) GetExpectedParameterCount() int { return 3 }

// GetIsNonDeterministic implements the FunctionInterface interface.
func (f Function3) GetIsNonDeterministic() bool { return f.IsNonDeterministic }

// GetIsStrict implements the FunctionInterface interface.
func (f Function3) GetIsStrict() bool { return f.Strict }

// enforceInterfaceInheritance implements the FunctionInterface interface.
func (f Function3) enforceInterfaceInheritance(error) {}

// GetName implements the FunctionInterface interface.
func (f Function4) GetName() string { return f.Name }

// GetReturn implements the FunctionInterface interface.
func (f Function4) GetReturn() pgtypes.DoltgresType { return f.Return }

// GetParameters implements the FunctionInterface interface.
func (f Function4) GetParameters() []pgtypes.DoltgresType { return f.Parameters }

// GetExpectedParameterCount implements the FunctionInterface interface.
func (f Function4) GetExpectedParameterCount() int { return 4 }

// GetIsNonDeterministic implements the FunctionInterface interface.
func (f Function4) GetIsNonDeterministic() bool { return f.IsNonDeterministic }

// GetIsStrict implements the FunctionInterface interface.
func (f Function4) GetIsStrict() bool { return f.Strict }

// enforceInterfaceInheritance implements the FunctionInterface interface.
func (f Function4) enforceInterfaceInheritance(error) {}
