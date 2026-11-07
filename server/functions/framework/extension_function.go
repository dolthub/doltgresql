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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression/function/aggregation"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/procedures"
	"github.com/dolthub/doltgresql/server/extensions"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ExtensionFunction is the implementation of functions that an extension provides, looked up in
// server/extensions by extension name and symbol.
type ExtensionFunction struct {
	ID                 id.Function
	ReturnType         *pgtypes.DoltgresType
	ParameterTypes     []*pgtypes.DoltgresType
	AllParams          []procedures.Parameter
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

// ExtensionWindowFunction is an extension function that may only be called as a window function, which hands the
// arguments of every row of a partition to its implementation at once.
type ExtensionWindowFunction struct {
	ExtensionFunction
}

var _ WindowFunctionInterface = ExtensionWindowFunction{}

// NewWindowFunc implements the interface WindowFunctionInterface.
func (extFunc ExtensionWindowFunction) NewWindowFunc() NewWindowFunctionFn {
	return func(args []sql.Expression, _ *sql.WindowDefinition) (sql.WindowFunction, error) {
		impl, ok := extensions.GetWindowFunction(extFunc.ExtensionName, extFunc.ExtensionSymbol)
		if !ok {
			return nil, errors.Errorf(`extension "%s" does not declare the window function "%s"`, extFunc.ExtensionName, extFunc.ExtensionSymbol)
		}
		return &extensionWindow{args: args, impl: impl}, nil
	}
}

// extensionWindow is the sql.WindowFunction of an ExtensionWindowFunction, which computes the results of a whole
// partition when the partition starts.
type extensionWindow struct {
	args    []sql.Expression
	impl    extdef.WindowFunction
	results []any
	next    int
}

var _ sql.WindowFunction = (*extensionWindow)(nil)

// DefaultFramer implements the interface sql.WindowFunction.
func (w *extensionWindow) DefaultFramer() sql.WindowFramer {
	return aggregation.NewPartitionFramer()
}

// StartPartition implements the interface sql.WindowFunction.
func (w *extensionWindow) StartPartition(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) error {
	rows := make([][]any, 0, interval.End-interval.Start)
	for _, row := range buf[interval.Start:interval.End] {
		values := make([]any, len(w.args))
		for i, arg := range w.args {
			value, err := arg.Eval(ctx, row)
			if err != nil {
				return err
			}
			if values[i], err = sql.UnwrapAny(ctx, value); err != nil {
				return err
			}
		}
		rows = append(rows, values)
	}
	results, err := w.impl(ctx, rows)
	if err != nil {
		return err
	}
	w.results, w.next = results, 0
	return nil
}

// Compute implements the interface sql.WindowFunction.
func (w *extensionWindow) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (any, error) {
	result := w.results[w.next]
	w.next++
	return result, nil
}

// Dispose implements the interface sql.WindowFunction.
func (w *extensionWindow) Dispose(ctx *sql.Context) {}
