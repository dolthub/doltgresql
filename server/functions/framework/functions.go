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
	GetReturn() *pgtypes.DoltgresType
	// GetParameters returns the parameter types for the function.
	GetParameters() []*pgtypes.DoltgresType
	// VariadicIndex returns the index of the variadic parameter, if it exists, or -1 otherwise
	VariadicIndex() int
	// GetExpectedParameterCount returns the number of parameters that are valid for this function.
	GetExpectedParameterCount() int
	// NonDeterministic returns whether the function is non-deterministic.
	NonDeterministic() bool
	// IsStrict returns whether the function is STRICT, which means if any parameter is NULL, then it returns NULL.
	// Otherwise, if it's not, the NULL input must be handled by user.
	IsStrict() bool
	// enforceInterfaceInheritance is a special function that ensures only the expected types inherit this interface.
	enforceInterfaceInheritance(error)
}

// Function0 is a function that does not take any parameters.
type Function0 struct {
	Name               string
	Return             *pgtypes.DoltgresType
	IsNonDeterministic bool
	Strict             bool
	Callable           func(ctx *sql.Context) (any, error)
}

// Function1 is a function that takes one parameter. The parameter and return type is passed into the Callable function
// when the parameter (and possibly return type) is a polymorphic type. The return type is the last type in the array.
type Function1 struct {
	Name               string
	Return             *pgtypes.DoltgresType
	Parameters         [1]*pgtypes.DoltgresType
	Variadic           bool
	IsNonDeterministic bool
	Strict             bool
	Callable           func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val1 any) (any, error)
}

// Function2 is a function that takes two parameters. The parameter and return types are passed into the Callable
// function when the parameters (and possibly return type) have at least one polymorphic type. The return type is the
// last type in the array.
type Function2 struct {
	Name               string
	Return             *pgtypes.DoltgresType
	Parameters         [2]*pgtypes.DoltgresType
	Variadic           bool
	IsNonDeterministic bool
	Strict             bool
	Callable           func(ctx *sql.Context, paramsAndReturn [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error)
}

// Function3 is a function that takes three parameters. The parameter and return types are passed into the Callable
// function when the parameters (and possibly return type) have at least one polymorphic type. The return type is the
// last type in the array.
type Function3 struct {
	Name               string
	Return             *pgtypes.DoltgresType
	Parameters         [3]*pgtypes.DoltgresType
	Variadic           bool
	IsNonDeterministic bool
	Strict             bool
	Callable           func(ctx *sql.Context, paramsAndReturn [4]*pgtypes.DoltgresType, val1 any, val2 any, val3 any) (any, error)
}

// Function4 is a function that takes four parameters. The parameter and return types are passed into the Callable
// function when the parameters (and possibly return type) have at least one polymorphic type. The return type is the
// last type in the array.
type Function4 struct {
	Name               string
	Return             *pgtypes.DoltgresType
	Parameters         [4]*pgtypes.DoltgresType
	Variadic           bool
	IsNonDeterministic bool
	Strict             bool
	Callable           func(ctx *sql.Context, paramsAndReturn [5]*pgtypes.DoltgresType, val1 any, val2 any, val3 any, val4 any) (any, error)
}

var _ FunctionInterface = Function0{}
var _ FunctionInterface = Function1{}
var _ FunctionInterface = Function2{}
var _ FunctionInterface = Function3{}
var _ FunctionInterface = Function4{}

// GetName implements the FunctionInterface interface.
func (f Function0) GetName() string { return f.Name }

// GetReturn implements the FunctionInterface interface.
func (f Function0) GetReturn() *pgtypes.DoltgresType { return f.Return }

// GetParameters implements the FunctionInterface interface.
func (f Function0) GetParameters() []*pgtypes.DoltgresType { return nil }

func (f Function0) VariadicIndex() int {
	return -1
}

// GetExpectedParameterCount implements the FunctionInterface interface.
func (f Function0) GetExpectedParameterCount() int { return 0 }

// NonDeterministic implements the FunctionInterface interface.
func (f Function0) NonDeterministic() bool { return f.IsNonDeterministic }

// IsStrict implements the FunctionInterface interface.
func (f Function0) IsStrict() bool { return f.Strict }

// enforceInterfaceInheritance implements the FunctionInterface interface.
func (f Function0) enforceInterfaceInheritance(error) {}

// GetName implements the FunctionInterface interface.
func (f Function1) GetName() string { return f.Name }

// GetReturn implements the FunctionInterface interface.
func (f Function1) GetReturn() *pgtypes.DoltgresType { return f.Return }

// GetParameters implements the FunctionInterface interface.
func (f Function1) GetParameters() []*pgtypes.DoltgresType { return f.Parameters[:] }

// VariadicIndex implements the FunctionInterface interface.
func (f Function1) VariadicIndex() int {
	if f.Variadic {
		return 0
	} else {
		return -1
	}
}

// GetExpectedParameterCount implements the FunctionInterface interface.
func (f Function1) GetExpectedParameterCount() int { return 1 }

// NonDeterministic implements the FunctionInterface interface.
func (f Function1) NonDeterministic() bool { return f.IsNonDeterministic }

// IsStrict implements the FunctionInterface interface.
func (f Function1) IsStrict() bool { return f.Strict }

// enforceInterfaceInheritance implements the FunctionInterface interface.
func (f Function1) enforceInterfaceInheritance(error) {}

// GetName implements the FunctionInterface interface.
func (f Function2) GetName() string { return f.Name }

// GetReturn implements the FunctionInterface interface.
func (f Function2) GetReturn() *pgtypes.DoltgresType { return f.Return }

// GetParameters implements the FunctionInterface interface.
func (f Function2) GetParameters() []*pgtypes.DoltgresType { return f.Parameters[:] }

// VariadicIndex implements the FunctionInterface interface.
func (f Function2) VariadicIndex() int {
	if f.Variadic {
		return 1
	} else {
		return -1
	}
}

// GetExpectedParameterCount implements the FunctionInterface interface.
func (f Function2) GetExpectedParameterCount() int { return 2 }

// NonDeterministic implements the FunctionInterface interface.
func (f Function2) NonDeterministic() bool { return f.IsNonDeterministic }

// IsStrict implements the FunctionInterface interface.
func (f Function2) IsStrict() bool { return f.Strict }

// enforceInterfaceInheritance implements the FunctionInterface interface.
func (f Function2) enforceInterfaceInheritance(error) {}

// GetName implements the FunctionInterface interface.
func (f Function3) GetName() string { return f.Name }

// GetReturn implements the FunctionInterface interface.
func (f Function3) GetReturn() *pgtypes.DoltgresType { return f.Return }

// GetParameters implements the FunctionInterface interface.
func (f Function3) GetParameters() []*pgtypes.DoltgresType { return f.Parameters[:] }

// VariadicIndex implements the FunctionInterface interface.
func (f Function3) VariadicIndex() int {
	if f.Variadic {
		return 2
	} else {
		return -1
	}
}

// GetExpectedParameterCount implements the FunctionInterface interface.
func (f Function3) GetExpectedParameterCount() int { return 3 }

// NonDeterministic implements the FunctionInterface interface.
func (f Function3) NonDeterministic() bool { return f.IsNonDeterministic }

// IsStrict implements the FunctionInterface interface.
func (f Function3) IsStrict() bool { return f.Strict }

// enforceInterfaceInheritance implements the FunctionInterface interface.
func (f Function3) enforceInterfaceInheritance(error) {}

// GetName implements the FunctionInterface interface.
func (f Function4) GetName() string { return f.Name }

// GetReturn implements the FunctionInterface interface.
func (f Function4) GetReturn() *pgtypes.DoltgresType { return f.Return }

// GetParameters implements the FunctionInterface interface.
func (f Function4) GetParameters() []*pgtypes.DoltgresType { return f.Parameters[:] }

// VariadicIndex implements the FunctionInterface interface.
func (f Function4) VariadicIndex() int {
	if f.Variadic {
		return 3
	} else {
		return -1
	}
}

// GetExpectedParameterCount implements the FunctionInterface interface.
func (f Function4) GetExpectedParameterCount() int { return 4 }

// NonDeterministic implements the FunctionInterface interface.
func (f Function4) NonDeterministic() bool { return f.IsNonDeterministic }

// IsStrict implements the FunctionInterface interface.
func (f Function4) IsStrict() bool { return f.Strict }

// enforceInterfaceInheritance implements the FunctionInterface interface.
func (f Function4) enforceInterfaceInheritance(error) {}
