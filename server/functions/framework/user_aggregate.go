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

// UserAggregate is the implementation of a user-defined aggregate.
type UserAggregate struct {
	ID             id.Function
	ReturnType     *pgtypes.DoltgresType
	ParameterTypes []*pgtypes.DoltgresType
	StateType      *pgtypes.DoltgresType
	SFunc          id.Function
	FinalFunc      id.Function
	InitCond       string
	HasInitCond    bool
}

var _ AggregateFunctionInterface = UserAggregate{}

// GetExpectedParameterCount implements the interface FunctionInterface.
func (agg UserAggregate) GetExpectedParameterCount() int {
	return len(agg.ParameterTypes)
}

// GetName implements the interface FunctionInterface.
func (agg UserAggregate) GetName() string {
	return agg.ID.FunctionName()
}

// GetOutParameters implements the interface FunctionInterface.
func (agg UserAggregate) GetOutParameters() sql.Schema {
	return nil
}

// GetInputParameterTypes implements the interface FunctionInterface.
func (agg UserAggregate) GetInputParameterTypes() []*pgtypes.DoltgresType {
	return agg.ParameterTypes
}

// GetReturn implements the interface FunctionInterface.
func (agg UserAggregate) GetReturn() *pgtypes.DoltgresType {
	return agg.ReturnType
}

// InternalID implements the interface FunctionInterface.
func (agg UserAggregate) InternalID() id.Id {
	return agg.ID.AsId()
}

// IsCVariadic implements the interface FunctionInterface.
func (agg UserAggregate) IsCVariadic() bool {
	return false
}

// IsSRF implements the interface FunctionInterface.
func (agg UserAggregate) IsSRF() bool {
	return false
}

// IsStrict implements the interface FunctionInterface.
func (agg UserAggregate) IsStrict() bool {
	return false
}

// NonDeterministic implements the interface FunctionInterface.
func (agg UserAggregate) NonDeterministic() bool {
	return false
}

// VariadicIndex implements the interface FunctionInterface.
func (agg UserAggregate) VariadicIndex() int {
	return -1
}

// NewBuffer implements the interface AggregateFunctionInterface.
func (agg UserAggregate) NewBuffer(exprs []sql.Expression) (sql.AggregationBuffer, error) {
	return &userAggregateBuffer{aggregate: agg, arguments: exprs}, nil
}

// NewWindowFunc implements the interface AggregateFunctionInterface.
func (agg UserAggregate) NewWindowFunc() NewWindowFunctionFn {
	//TODO: support user-defined aggregates within an OVER(...) clause
	return nil
}

// enforceInterfaceInheritance implements the interface FunctionInterface.
func (agg UserAggregate) enforceInterfaceInheritance(error) {}

// userAggregateBuffer accumulates the transition state of a UserAggregate over the rows of a group.
type userAggregateBuffer struct {
	aggregate   UserAggregate
	arguments   []sql.Expression
	sFunc       *UserFunctionCall
	finalFunc   *UserFunctionCall
	state       any
	stateExists bool
}

var _ sql.AggregationBuffer = (*userAggregateBuffer)(nil)

// Dispose implements the interface sql.AggregationBuffer.
func (b *userAggregateBuffer) Dispose(ctx *sql.Context) {}

// Eval implements the interface sql.AggregationBuffer.
func (b *userAggregateBuffer) Eval(ctx *sql.Context) (interface{}, error) {
	if err := b.resolve(ctx); err != nil {
		return nil, err
	}
	if b.finalFunc == nil {
		return b.state, nil
	}
	return b.finalFunc.Call(ctx, b.state)
}

// Update implements the interface sql.AggregationBuffer.
func (b *userAggregateBuffer) Update(ctx *sql.Context, row sql.Row) error {
	if err := b.resolve(ctx); err != nil {
		return err
	}
	args := make([]any, len(b.arguments)+1)
	for i, argument := range b.arguments {
		val, err := argument.Eval(ctx, row)
		if err != nil {
			return err
		}
		if val == nil && b.sFunc.strict {
			return nil
		}
		args[i+1] = val
	}
	if b.sFunc.strict && !b.stateExists {
		b.state = args[1]
		b.stateExists = true
		return nil
	}
	args[0] = b.state
	newState, err := b.sFunc.Call(ctx, args...)
	if err != nil {
		return err
	}
	b.state = newState
	b.stateExists = true
	return nil
}

// resolve loads the transition and final functions.
func (b *userAggregateBuffer) resolve(ctx *sql.Context) error {
	if b.sFunc != nil {
		return nil
	}
	sFuncParams := append([]*pgtypes.DoltgresType{b.aggregate.StateType}, b.aggregate.ParameterTypes...)
	b.sFunc = NewUserFunctionCall(ctx, b.aggregate.SFunc.SchemaName(), b.aggregate.SFunc.FunctionName(), sFuncParams)
	if b.sFunc == nil {
		return ErrFunctionDoesNotExist.New(b.aggregate.SFunc.DisplayString())
	}
	if b.aggregate.FinalFunc.IsValid() {
		b.finalFunc = NewUserFunctionCall(ctx, b.aggregate.FinalFunc.SchemaName(), b.aggregate.FinalFunc.FunctionName(),
			[]*pgtypes.DoltgresType{b.aggregate.StateType})
		if b.finalFunc == nil {
			return ErrFunctionDoesNotExist.New(b.aggregate.FinalFunc.DisplayString())
		}
	}
	if b.aggregate.HasInitCond {
		initCond, err := b.aggregate.StateType.IoInput(ctx, b.aggregate.InitCond)
		if err != nil {
			return err
		}
		b.state = initCond
		b.stateExists = true
	}
	return nil
}
