// Copyright 2023 Dolthub, Inc.
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

// Copyright 2015 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tree

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/postgres/parser/geo"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/roleoption"
	"github.com/dolthub/doltgresql/postgres/parser/types"
)

var (
	// ErrDivByZero is reported on a division by zero.
	ErrDivByZero = pgerror.New(pgcode.DivisionByZero, "division by zero")
)

// UnaryOp is a unary operator.
type UnaryOp struct {
	Typ        *types.T
	ReturnType *types.T
	Volatility Volatility

	types   TypeList
	retType ReturnTyper
}

func (op *UnaryOp) params() TypeList {
	return op.types
}

func (op *UnaryOp) returnType() ReturnTyper {
	return op.retType
}

func (*UnaryOp) preferred() bool {
	return false
}

func unaryOpFixups(ops map[UnaryOperator]unaryOpOverload) map[UnaryOperator]unaryOpOverload {
	for op, overload := range ops {
		for i, impl := range overload {
			casted := impl.(*UnaryOp)
			casted.types = ArgTypes{{"arg", casted.Typ}}
			casted.retType = FixedReturnType(casted.ReturnType)
			ops[op][i] = casted
		}
	}
	return ops
}

// unaryOpOverload is an overloaded set of unary operator implementations.
type unaryOpOverload []overloadImpl

// UnaryOps contains the unary operations indexed by operation type.
var UnaryOps = unaryOpFixups(map[UnaryOperator]unaryOpOverload{
	UnaryMinus: {
		&UnaryOp{
			Typ:        types.Int,
			ReturnType: types.Int,
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.Float,
			ReturnType: types.Float,
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.Interval,
			ReturnType: types.Interval,
			Volatility: VolatilityImmutable,
		},
	},

	UnaryComplement: {
		&UnaryOp{
			Typ:        types.Int,
			ReturnType: types.Int,
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.VarBit,
			ReturnType: types.VarBit,
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.INet,
			ReturnType: types.INet,
			Volatility: VolatilityImmutable,
		},
	},

	UnarySqrt: {
		&UnaryOp{
			Typ:        types.Float,
			ReturnType: types.Float,
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
	},

	UnaryCbrt: {
		&UnaryOp{
			Typ:        types.Float,
			ReturnType: types.Float,
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
	},
})

// BinOp is a binary operator.
type BinOp struct {
	LeftType     *types.T
	RightType    *types.T
	ReturnType   *types.T
	NullableArgs bool
	Volatility   Volatility

	types   TypeList
	retType ReturnTyper
}

func (op *BinOp) params() TypeList {
	return op.types
}

func (op *BinOp) matchParams(l, r *types.T) bool {
	return op.params().MatchAt(l, 0) && op.params().MatchAt(r, 1)
}

func (op *BinOp) returnType() ReturnTyper {
	return op.retType
}

func (*BinOp) preferred() bool {
	return false
}

// TODO(justin): these might be improved by making arrays into an interface and
// then introducing a ConcatenatedArray implementation which just references two
// existing arrays. This would optimize the common case of appending an element
// (or array) to an array from O(n) to O(1).
func initArrayElementConcatenation() {
	for _, t := range types.Scalar {
		typ := t
		BinOps[Concat] = append(BinOps[Concat], &BinOp{
			LeftType:     types.MakeArray(typ),
			RightType:    typ,
			ReturnType:   types.MakeArray(typ),
			NullableArgs: true,
			Volatility:   VolatilityImmutable,
		})

		BinOps[Concat] = append(BinOps[Concat], &BinOp{
			LeftType:     typ,
			RightType:    types.MakeArray(typ),
			ReturnType:   types.MakeArray(typ),
			NullableArgs: true,
			Volatility:   VolatilityImmutable,
		})
	}
}

func initArrayToArrayConcatenation() {
	for _, t := range types.Scalar {
		typ := t
		BinOps[Concat] = append(BinOps[Concat], &BinOp{
			LeftType:     types.MakeArray(typ),
			RightType:    types.MakeArray(typ),
			ReturnType:   types.MakeArray(typ),
			NullableArgs: true,
			Volatility:   VolatilityImmutable,
		})
	}
}

func init() {
	initArrayElementConcatenation()
	initArrayToArrayConcatenation()
}

func init() {
	for op, overload := range BinOps {
		for i, impl := range overload {
			casted := impl.(*BinOp)
			casted.types = ArgTypes{{"left", casted.LeftType}, {"right", casted.RightType}}
			casted.retType = FixedReturnType(casted.ReturnType)
			BinOps[op][i] = casted
		}
	}
}

// binOpOverload is an overloaded set of binary operator implementations.
type binOpOverload []overloadImpl

func (o binOpOverload) lookupImpl(left, right *types.T) (*BinOp, bool) {
	for _, fn := range o {
		casted := fn.(*BinOp)
		if casted.matchParams(left, right) {
			return casted, true
		}
	}
	return nil, false
}

// BinOps contains the binary operations indexed by operation type.
var BinOps = map[BinaryOperator]binOpOverload{
	Bitand: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.VarBit,
			RightType:  types.VarBit,
			ReturnType: types.VarBit,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.INet,
			RightType:  types.INet,
			ReturnType: types.INet,
			Volatility: VolatilityImmutable,
		},
	},

	Bitor: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.VarBit,
			RightType:  types.VarBit,
			ReturnType: types.VarBit,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.INet,
			RightType:  types.INet,
			ReturnType: types.INet,
			Volatility: VolatilityImmutable,
		},
	},

	Bitxor: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.VarBit,
			RightType:  types.VarBit,
			ReturnType: types.VarBit,
			Volatility: VolatilityImmutable,
		},
	},

	Plus: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Float,
			ReturnType: types.Float,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.Int,
			ReturnType: types.Date,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Date,
			ReturnType: types.Date,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.Time,
			ReturnType: types.Timestamp,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Time,
			RightType:  types.Date,
			ReturnType: types.Timestamp,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.TimeTZ,
			ReturnType: types.TimestampTZ,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.TimeTZ,
			RightType:  types.Date,
			ReturnType: types.TimestampTZ,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Time,
			RightType:  types.Interval,
			ReturnType: types.Time,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Time,
			ReturnType: types.Time,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.TimeTZ,
			RightType:  types.Interval,
			ReturnType: types.TimeTZ,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.TimeTZ,
			ReturnType: types.TimeTZ,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Timestamp,
			RightType:  types.Interval,
			ReturnType: types.Timestamp,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Timestamp,
			ReturnType: types.Timestamp,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.TimestampTZ,
			RightType:  types.Interval,
			ReturnType: types.TimestampTZ,
			Volatility: VolatilityStable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.TimestampTZ,
			ReturnType: types.TimestampTZ,
			Volatility: VolatilityStable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Interval,
			ReturnType: types.Interval,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.Interval,
			ReturnType: types.Timestamp,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Date,
			ReturnType: types.Timestamp,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.INet,
			RightType:  types.Int,
			ReturnType: types.INet,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.INet,
			ReturnType: types.INet,
			Volatility: VolatilityImmutable,
		},
	},

	Minus: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Float,
			ReturnType: types.Float,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.Int,
			ReturnType: types.Date,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.Date,
			ReturnType: types.Int,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.Time,
			ReturnType: types.Timestamp,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Time,
			RightType:  types.Time,
			ReturnType: types.Interval,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Timestamp,
			RightType:  types.Timestamp,
			ReturnType: types.Interval,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.TimestampTZ,
			RightType:  types.TimestampTZ,
			ReturnType: types.Interval,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Timestamp,
			RightType:  types.TimestampTZ,
			ReturnType: types.Interval,
			Volatility: VolatilityStable,
		},
		&BinOp{
			LeftType:   types.TimestampTZ,
			RightType:  types.Timestamp,
			ReturnType: types.Interval,
			Volatility: VolatilityStable,
		},
		&BinOp{
			LeftType:   types.Time,
			RightType:  types.Interval,
			ReturnType: types.Time,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.TimeTZ,
			RightType:  types.Interval,
			ReturnType: types.TimeTZ,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Timestamp,
			RightType:  types.Interval,
			ReturnType: types.Timestamp,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.TimestampTZ,
			RightType:  types.Interval,
			ReturnType: types.TimestampTZ,
			Volatility: VolatilityStable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.Interval,
			ReturnType: types.Timestamp,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Interval,
			ReturnType: types.Interval,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.String,
			ReturnType: types.Jsonb,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.Int,
			ReturnType: types.Jsonb,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.MakeArray(types.String),
			ReturnType: types.Jsonb,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.INet,
			RightType:  types.INet,
			ReturnType: types.Int,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			// Note: postgres ver 10 does NOT have Int - INet. Throws ERROR: 42883.
			LeftType:   types.INet,
			RightType:  types.Int,
			ReturnType: types.INet,
			Volatility: VolatilityImmutable,
		},
	},

	Mult: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Float,
			ReturnType: types.Float,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		// The following two overloads are needed because DInt/DInt = DDecimal. Due
		// to this operation, normalization may sometimes create a DInt * DDecimal
		// operation.
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Interval,
			ReturnType: types.Interval,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Int,
			ReturnType: types.Interval,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Float,
			ReturnType: types.Interval,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Interval,
			ReturnType: types.Interval,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Interval,
			ReturnType: types.Interval,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Decimal,
			ReturnType: types.Interval,
			Volatility: VolatilityImmutable,
		},
	},

	Div: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Float,
			ReturnType: types.Float,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Int,
			ReturnType: types.Interval,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Float,
			ReturnType: types.Interval,
			Volatility: VolatilityImmutable,
		},
	},

	FloorDiv: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Float,
			ReturnType: types.Float,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
	},

	Mod: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Float,
			ReturnType: types.Float,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
	},

	Concat: {
		&BinOp{
			LeftType:   types.String,
			RightType:  types.String,
			ReturnType: types.String,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Bytes,
			RightType:  types.Bytes,
			ReturnType: types.Bytes,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.VarBit,
			RightType:  types.VarBit,
			ReturnType: types.VarBit,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.Jsonb,
			ReturnType: types.Jsonb,
			Volatility: VolatilityImmutable,
		},
	},

	// TODO(pmattis): Check that the shift is valid.
	LShift: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.VarBit,
			RightType:  types.Int,
			ReturnType: types.VarBit,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.INet,
			RightType:  types.INet,
			ReturnType: types.Bool,
			Volatility: VolatilityImmutable,
		},
	},

	RShift: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.VarBit,
			RightType:  types.Int,
			ReturnType: types.VarBit,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.INet,
			RightType:  types.INet,
			ReturnType: types.Bool,
			Volatility: VolatilityImmutable,
		},
	},

	Pow: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Float,
			ReturnType: types.Float,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Volatility: VolatilityImmutable,
		},
	},

	JSONFetchVal: {
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.String,
			ReturnType: types.Jsonb,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.Int,
			ReturnType: types.Jsonb,
			Volatility: VolatilityImmutable,
		},
	},

	JSONFetchValPath: {
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.MakeArray(types.String),
			ReturnType: types.Jsonb,
			Volatility: VolatilityImmutable,
		},
	},

	JSONFetchText: {
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.String,
			ReturnType: types.String,
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.Int,
			ReturnType: types.String,
			Volatility: VolatilityImmutable,
		},
	},

	JSONFetchTextPath: {
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.MakeArray(types.String),
			ReturnType: types.String,
			Volatility: VolatilityImmutable,
		},
	},
}

// CmpOp is a comparison operator.
type CmpOp struct {
	LeftType  *types.T
	RightType *types.T

	// If NullableArgs is false, the operator returns NULL
	// whenever either argument is NULL.
	NullableArgs bool

	Volatility Volatility

	types       TypeList
	isPreferred bool
}

func (op *CmpOp) params() TypeList {
	return op.types
}

func (op *CmpOp) matchParams(l, r *types.T) bool {
	return op.params().MatchAt(l, 0) && op.params().MatchAt(r, 1)
}

var cmpOpReturnType = FixedReturnType(types.Bool)

func (op *CmpOp) returnType() ReturnTyper {
	return cmpOpReturnType
}

func (op *CmpOp) preferred() bool {
	return op.isPreferred
}

func cmpOpFixups(cmpOps map[ComparisonOperator]cmpOpOverload) map[ComparisonOperator]cmpOpOverload {
	findVolatility := func(op ComparisonOperator, t *types.T) Volatility {
		for _, impl := range cmpOps[EQ] {
			o := impl.(*CmpOp)
			if o.LeftType.Equivalent(t) && o.RightType.Equivalent(t) {
				return o.Volatility
			}
		}
		panic(errors.AssertionFailedf("could not find cmp op %s(%s,%s)", op, t, t))
	}

	// Array equality comparisons.
	for _, t := range types.Scalar {
		cmpOps[EQ] = append(cmpOps[EQ], &CmpOp{
			LeftType:   types.MakeArray(t),
			RightType:  types.MakeArray(t),
			Volatility: findVolatility(EQ, t),
		})
		cmpOps[LE] = append(cmpOps[LE], &CmpOp{
			LeftType:   types.MakeArray(t),
			RightType:  types.MakeArray(t),
			Volatility: findVolatility(LE, t),
		})
		cmpOps[LT] = append(cmpOps[LT], &CmpOp{
			LeftType:   types.MakeArray(t),
			RightType:  types.MakeArray(t),
			Volatility: findVolatility(LT, t),
		})

		cmpOps[IsNotDistinctFrom] = append(cmpOps[IsNotDistinctFrom], &CmpOp{
			LeftType:     types.MakeArray(t),
			RightType:    types.MakeArray(t),
			NullableArgs: true,
			Volatility:   findVolatility(IsNotDistinctFrom, t),
		})
	}

	for op, overload := range cmpOps {
		for i, impl := range overload {
			casted := impl.(*CmpOp)
			casted.types = ArgTypes{{"left", casted.LeftType}, {"right", casted.RightType}}
			cmpOps[op][i] = casted
		}
	}

	return cmpOps
}

// cmpOpOverload is an overloaded set of comparison operator implementations.
type cmpOpOverload []overloadImpl

func (o cmpOpOverload) LookupImpl(left, right *types.T) (*CmpOp, bool) {
	for _, fn := range o {
		casted := fn.(*CmpOp)
		if casted.matchParams(left, right) {
			return casted, true
		}
	}
	return nil, false
}

func makeCmpOpOverload(
	a, b *types.T,
	nullableArgs bool,
	v Volatility,
) *CmpOp {
	return &CmpOp{
		LeftType:     a,
		RightType:    b,
		NullableArgs: nullableArgs,
		Volatility:   v,
	}
}

func makeEqFn(a, b *types.T, v Volatility) *CmpOp {
	return makeCmpOpOverload(a, b, false /* NullableArgs */, v)
}
func makeLtFn(a, b *types.T, v Volatility) *CmpOp {
	return makeCmpOpOverload(a, b, false /* NullableArgs */, v)
}
func makeLeFn(a, b *types.T, v Volatility) *CmpOp {
	return makeCmpOpOverload(a, b, false /* NullableArgs */, v)
}
func makeIsFn(a, b *types.T, v Volatility) *CmpOp {
	return makeCmpOpOverload(a, b, true /* NullableArgs */, v)
}

// CmpOps contains the comparison operations indexed by operation type.
var CmpOps = cmpOpFixups(map[ComparisonOperator]cmpOpOverload{
	EQ: {
		// Single-type comparisons.
		makeEqFn(types.AnyEnum, types.AnyEnum, VolatilityImmutable),
		makeEqFn(types.Bool, types.Bool, VolatilityLeakProof),
		makeEqFn(types.Bytes, types.Bytes, VolatilityLeakProof),
		makeEqFn(types.Date, types.Date, VolatilityLeakProof),
		makeEqFn(types.Decimal, types.Decimal, VolatilityImmutable),
		// Note: it is an error to compare two strings with different collations;
		// the operator is leak proof under the assumption that these cases will be
		// detected during type checking.
		makeEqFn(types.AnyCollatedString, types.AnyCollatedString, VolatilityLeakProof),
		makeEqFn(types.Float, types.Float, VolatilityLeakProof),
		makeEqFn(types.Box2D, types.Box2D, VolatilityLeakProof),
		makeEqFn(types.Geography, types.Geography, VolatilityLeakProof),
		makeEqFn(types.Geometry, types.Geometry, VolatilityLeakProof),
		makeEqFn(types.INet, types.INet, VolatilityLeakProof),
		makeEqFn(types.Int, types.Int, VolatilityLeakProof),
		makeEqFn(types.Interval, types.Interval, VolatilityLeakProof),
		makeEqFn(types.Jsonb, types.Jsonb, VolatilityImmutable),
		makeEqFn(types.Oid, types.Oid, VolatilityLeakProof),
		makeEqFn(types.String, types.String, VolatilityLeakProof),
		makeEqFn(types.Time, types.Time, VolatilityLeakProof),
		makeEqFn(types.TimeTZ, types.TimeTZ, VolatilityLeakProof),
		makeEqFn(types.Timestamp, types.Timestamp, VolatilityLeakProof),
		makeEqFn(types.TimestampTZ, types.TimestampTZ, VolatilityLeakProof),
		makeEqFn(types.Uuid, types.Uuid, VolatilityLeakProof),
		makeEqFn(types.VarBit, types.VarBit, VolatilityLeakProof),

		// Mixed-type comparisons.
		makeEqFn(types.Date, types.Timestamp, VolatilityImmutable),
		makeEqFn(types.Date, types.TimestampTZ, VolatilityStable),
		makeEqFn(types.Decimal, types.Float, VolatilityLeakProof),
		makeEqFn(types.Decimal, types.Int, VolatilityLeakProof),
		makeEqFn(types.Float, types.Decimal, VolatilityLeakProof),
		makeEqFn(types.Float, types.Int, VolatilityLeakProof),
		makeEqFn(types.Int, types.Decimal, VolatilityLeakProof),
		makeEqFn(types.Int, types.Float, VolatilityLeakProof),
		makeEqFn(types.Int, types.Oid, VolatilityLeakProof),
		makeEqFn(types.Oid, types.Int, VolatilityLeakProof),
		makeEqFn(types.Timestamp, types.Date, VolatilityImmutable),
		makeEqFn(types.Timestamp, types.TimestampTZ, VolatilityStable),
		makeEqFn(types.TimestampTZ, types.Date, VolatilityStable),
		makeEqFn(types.TimestampTZ, types.Timestamp, VolatilityStable),
		makeEqFn(types.Time, types.TimeTZ, VolatilityStable),
		makeEqFn(types.TimeTZ, types.Time, VolatilityStable),

		// Tuple comparison.
		&CmpOp{
			LeftType:   types.AnyTuple,
			RightType:  types.AnyTuple,
			Volatility: VolatilityImmutable,
		},
	},

	LT: {
		// Single-type comparisons.
		makeLtFn(types.AnyEnum, types.AnyEnum, VolatilityImmutable),
		makeLtFn(types.Bool, types.Bool, VolatilityLeakProof),
		makeLtFn(types.Bytes, types.Bytes, VolatilityLeakProof),
		makeLtFn(types.Date, types.Date, VolatilityLeakProof),
		makeLtFn(types.Decimal, types.Decimal, VolatilityImmutable),
		makeLtFn(types.AnyCollatedString, types.AnyCollatedString, VolatilityLeakProof),
		// Note: it is an error to compare two strings with different collations;
		// the operator is leak proof under the assumption that these cases will be
		// detected during type checking.
		makeLtFn(types.Float, types.Float, VolatilityLeakProof),
		makeLtFn(types.Box2D, types.Box2D, VolatilityLeakProof),
		makeLtFn(types.Geography, types.Geography, VolatilityLeakProof),
		makeLtFn(types.Geometry, types.Geometry, VolatilityLeakProof),
		makeLtFn(types.INet, types.INet, VolatilityLeakProof),
		makeLtFn(types.Int, types.Int, VolatilityLeakProof),
		makeLtFn(types.Interval, types.Interval, VolatilityLeakProof),
		makeLtFn(types.Oid, types.Oid, VolatilityLeakProof),
		makeLtFn(types.String, types.String, VolatilityLeakProof),
		makeLtFn(types.Time, types.Time, VolatilityLeakProof),
		makeLtFn(types.TimeTZ, types.TimeTZ, VolatilityLeakProof),
		makeLtFn(types.Timestamp, types.Timestamp, VolatilityLeakProof),
		makeLtFn(types.TimestampTZ, types.TimestampTZ, VolatilityLeakProof),
		makeLtFn(types.Uuid, types.Uuid, VolatilityLeakProof),
		makeLtFn(types.VarBit, types.VarBit, VolatilityLeakProof),

		// Mixed-type comparisons.
		makeLtFn(types.Date, types.Timestamp, VolatilityImmutable),
		makeLtFn(types.Date, types.TimestampTZ, VolatilityStable),
		makeLtFn(types.Decimal, types.Float, VolatilityLeakProof),
		makeLtFn(types.Decimal, types.Int, VolatilityLeakProof),
		makeLtFn(types.Float, types.Decimal, VolatilityLeakProof),
		makeLtFn(types.Float, types.Int, VolatilityLeakProof),
		makeLtFn(types.Int, types.Decimal, VolatilityLeakProof),
		makeLtFn(types.Int, types.Float, VolatilityLeakProof),
		makeLtFn(types.Int, types.Oid, VolatilityLeakProof),
		makeLtFn(types.Oid, types.Int, VolatilityLeakProof),
		makeLtFn(types.Timestamp, types.Date, VolatilityImmutable),
		makeLtFn(types.Timestamp, types.TimestampTZ, VolatilityStable),
		makeLtFn(types.TimestampTZ, types.Date, VolatilityStable),
		makeLtFn(types.TimestampTZ, types.Timestamp, VolatilityStable),
		makeLtFn(types.Time, types.TimeTZ, VolatilityStable),
		makeLtFn(types.TimeTZ, types.Time, VolatilityStable),

		// Tuple comparison.
		&CmpOp{
			LeftType:   types.AnyTuple,
			RightType:  types.AnyTuple,
			Volatility: VolatilityImmutable,
		},
	},

	LE: {
		// Single-type comparisons.
		makeLeFn(types.AnyEnum, types.AnyEnum, VolatilityImmutable),
		makeLeFn(types.Bool, types.Bool, VolatilityLeakProof),
		makeLeFn(types.Bytes, types.Bytes, VolatilityLeakProof),
		makeLeFn(types.Date, types.Date, VolatilityLeakProof),
		makeLeFn(types.Decimal, types.Decimal, VolatilityImmutable),
		// Note: it is an error to compare two strings with different collations;
		// the operator is leak proof under the assumption that these cases will be
		// detected during type checking.
		makeLeFn(types.AnyCollatedString, types.AnyCollatedString, VolatilityLeakProof),
		makeLeFn(types.Float, types.Float, VolatilityLeakProof),
		makeLeFn(types.Box2D, types.Box2D, VolatilityLeakProof),
		makeLeFn(types.Geography, types.Geography, VolatilityLeakProof),
		makeLeFn(types.Geometry, types.Geometry, VolatilityLeakProof),
		makeLeFn(types.INet, types.INet, VolatilityLeakProof),
		makeLeFn(types.Int, types.Int, VolatilityLeakProof),
		makeLeFn(types.Interval, types.Interval, VolatilityLeakProof),
		makeLeFn(types.Oid, types.Oid, VolatilityLeakProof),
		makeLeFn(types.String, types.String, VolatilityLeakProof),
		makeLeFn(types.Time, types.Time, VolatilityLeakProof),
		makeLeFn(types.TimeTZ, types.TimeTZ, VolatilityLeakProof),
		makeLeFn(types.Timestamp, types.Timestamp, VolatilityLeakProof),
		makeLeFn(types.TimestampTZ, types.TimestampTZ, VolatilityLeakProof),
		makeLeFn(types.Uuid, types.Uuid, VolatilityLeakProof),
		makeLeFn(types.VarBit, types.VarBit, VolatilityLeakProof),

		// Mixed-type comparisons.
		makeLeFn(types.Date, types.Timestamp, VolatilityImmutable),
		makeLeFn(types.Date, types.TimestampTZ, VolatilityStable),
		makeLeFn(types.Decimal, types.Float, VolatilityLeakProof),
		makeLeFn(types.Decimal, types.Int, VolatilityLeakProof),
		makeLeFn(types.Float, types.Decimal, VolatilityLeakProof),
		makeLeFn(types.Float, types.Int, VolatilityLeakProof),
		makeLeFn(types.Int, types.Decimal, VolatilityLeakProof),
		makeLeFn(types.Int, types.Float, VolatilityLeakProof),
		makeLeFn(types.Int, types.Oid, VolatilityLeakProof),
		makeLeFn(types.Oid, types.Int, VolatilityLeakProof),
		makeLeFn(types.Timestamp, types.Date, VolatilityImmutable),
		makeLeFn(types.Timestamp, types.TimestampTZ, VolatilityStable),
		makeLeFn(types.TimestampTZ, types.Date, VolatilityStable),
		makeLeFn(types.TimestampTZ, types.Timestamp, VolatilityStable),
		makeLeFn(types.Time, types.TimeTZ, VolatilityStable),
		makeLeFn(types.TimeTZ, types.Time, VolatilityStable),

		// Tuple comparison.
		&CmpOp{
			LeftType:   types.AnyTuple,
			RightType:  types.AnyTuple,
			Volatility: VolatilityImmutable,
		},
	},

	IsNotDistinctFrom: {
		&CmpOp{
			LeftType:     types.Unknown,
			RightType:    types.Unknown,
			NullableArgs: true,
			// Avoids ambiguous comparison error for NULL IS NOT DISTINCT FROM NULL>
			isPreferred: true,
			Volatility:  VolatilityLeakProof,
		},
		// Single-type comparisons.
		makeIsFn(types.AnyEnum, types.AnyEnum, VolatilityImmutable),
		makeIsFn(types.Bool, types.Bool, VolatilityLeakProof),
		makeIsFn(types.Bytes, types.Bytes, VolatilityLeakProof),
		makeIsFn(types.Date, types.Date, VolatilityLeakProof),
		makeIsFn(types.Decimal, types.Decimal, VolatilityImmutable),
		// Note: it is an error to compare two strings with different collations;
		// the operator is leak proof under the assumption that these cases will be
		// detected during type checking.
		makeIsFn(types.AnyCollatedString, types.AnyCollatedString, VolatilityLeakProof),
		makeIsFn(types.Float, types.Float, VolatilityLeakProof),
		makeIsFn(types.Box2D, types.Box2D, VolatilityLeakProof),
		makeIsFn(types.Geography, types.Geography, VolatilityLeakProof),
		makeIsFn(types.Geometry, types.Geometry, VolatilityLeakProof),
		makeIsFn(types.INet, types.INet, VolatilityLeakProof),
		makeIsFn(types.Int, types.Int, VolatilityLeakProof),
		makeIsFn(types.Interval, types.Interval, VolatilityLeakProof),
		makeIsFn(types.Jsonb, types.Jsonb, VolatilityImmutable),
		makeIsFn(types.Oid, types.Oid, VolatilityLeakProof),
		makeIsFn(types.String, types.String, VolatilityLeakProof),
		makeIsFn(types.Time, types.Time, VolatilityLeakProof),
		makeIsFn(types.TimeTZ, types.TimeTZ, VolatilityLeakProof),
		makeIsFn(types.Timestamp, types.Timestamp, VolatilityLeakProof),
		makeIsFn(types.TimestampTZ, types.TimestampTZ, VolatilityLeakProof),
		makeIsFn(types.Uuid, types.Uuid, VolatilityLeakProof),
		makeIsFn(types.VarBit, types.VarBit, VolatilityLeakProof),

		// Mixed-type comparisons.
		makeIsFn(types.Date, types.Timestamp, VolatilityImmutable),
		makeIsFn(types.Date, types.TimestampTZ, VolatilityStable),
		makeIsFn(types.Decimal, types.Float, VolatilityLeakProof),
		makeIsFn(types.Decimal, types.Int, VolatilityLeakProof),
		makeIsFn(types.Float, types.Decimal, VolatilityLeakProof),
		makeIsFn(types.Float, types.Int, VolatilityLeakProof),
		makeIsFn(types.Int, types.Decimal, VolatilityLeakProof),
		makeIsFn(types.Int, types.Float, VolatilityLeakProof),
		makeIsFn(types.Int, types.Oid, VolatilityLeakProof),
		makeIsFn(types.Oid, types.Int, VolatilityLeakProof),
		makeIsFn(types.Timestamp, types.Date, VolatilityImmutable),
		makeIsFn(types.Timestamp, types.TimestampTZ, VolatilityStable),
		makeIsFn(types.TimestampTZ, types.Date, VolatilityStable),
		makeIsFn(types.TimestampTZ, types.Timestamp, VolatilityStable),
		makeIsFn(types.Time, types.TimeTZ, VolatilityStable),
		makeIsFn(types.TimeTZ, types.Time, VolatilityStable),

		// Tuple comparison.
		&CmpOp{
			LeftType:     types.AnyTuple,
			RightType:    types.AnyTuple,
			NullableArgs: true,
			Volatility:   VolatilityImmutable,
		},
	},

	In: {
		makeEvalTupleIn(types.AnyEnum, VolatilityLeakProof),
		makeEvalTupleIn(types.Bool, VolatilityLeakProof),
		makeEvalTupleIn(types.Bytes, VolatilityLeakProof),
		makeEvalTupleIn(types.Date, VolatilityLeakProof),
		makeEvalTupleIn(types.Decimal, VolatilityLeakProof),
		makeEvalTupleIn(types.AnyCollatedString, VolatilityLeakProof),
		makeEvalTupleIn(types.AnyTuple, VolatilityLeakProof),
		makeEvalTupleIn(types.Float, VolatilityLeakProof),
		makeEvalTupleIn(types.Box2D, VolatilityLeakProof),
		makeEvalTupleIn(types.Geography, VolatilityLeakProof),
		makeEvalTupleIn(types.Geometry, VolatilityLeakProof),
		makeEvalTupleIn(types.INet, VolatilityLeakProof),
		makeEvalTupleIn(types.Int, VolatilityLeakProof),
		makeEvalTupleIn(types.Interval, VolatilityLeakProof),
		makeEvalTupleIn(types.Jsonb, VolatilityLeakProof),
		makeEvalTupleIn(types.Oid, VolatilityLeakProof),
		makeEvalTupleIn(types.String, VolatilityLeakProof),
		makeEvalTupleIn(types.Time, VolatilityLeakProof),
		makeEvalTupleIn(types.TimeTZ, VolatilityLeakProof),
		makeEvalTupleIn(types.Timestamp, VolatilityLeakProof),
		makeEvalTupleIn(types.TimestampTZ, VolatilityLeakProof),
		makeEvalTupleIn(types.Uuid, VolatilityLeakProof),
		makeEvalTupleIn(types.VarBit, VolatilityLeakProof),
	},

	Like: {
		&CmpOp{
			LeftType:   types.String,
			RightType:  types.String,
			Volatility: VolatilityLeakProof,
		},
	},

	ILike: {
		&CmpOp{
			LeftType:   types.String,
			RightType:  types.String,
			Volatility: VolatilityLeakProof,
		},
	},

	SimilarTo: {
		&CmpOp{
			LeftType:   types.String,
			RightType:  types.String,
			Volatility: VolatilityLeakProof,
		},
	},

	RegMatch: append(
		cmpOpOverload{
			&CmpOp{
				LeftType:   types.String,
				RightType:  types.String,
				Volatility: VolatilityImmutable,
			},
		},
		makeBox2DComparisonOperators(
			func(lhs, rhs *geo.CartesianBoundingBox) bool {
				return lhs.Covers(rhs)
			},
		)...,
	),

	RegIMatch: {
		&CmpOp{
			LeftType:   types.String,
			RightType:  types.String,
			Volatility: VolatilityImmutable,
		},
	},

	JSONExists: {
		&CmpOp{
			LeftType:   types.Jsonb,
			RightType:  types.String,
			Volatility: VolatilityImmutable,
		},
	},

	JSONSomeExists: {
		&CmpOp{
			LeftType:   types.Jsonb,
			RightType:  types.StringArray,
			Volatility: VolatilityImmutable,
		},
	},

	JSONAllExists: {
		&CmpOp{
			LeftType:   types.Jsonb,
			RightType:  types.StringArray,
			Volatility: VolatilityImmutable,
		},
	},

	Contains: {
		&CmpOp{
			LeftType:   types.AnyArray,
			RightType:  types.AnyArray,
			Volatility: VolatilityImmutable,
		},
		&CmpOp{
			LeftType:   types.Jsonb,
			RightType:  types.Jsonb,
			Volatility: VolatilityImmutable,
		},
	},

	ContainedBy: {
		&CmpOp{
			LeftType:   types.AnyArray,
			RightType:  types.AnyArray,
			Volatility: VolatilityImmutable,
		},
		&CmpOp{
			LeftType:   types.Jsonb,
			RightType:  types.Jsonb,
			Volatility: VolatilityImmutable,
		},
	},
	Overlaps: append(
		cmpOpOverload{
			&CmpOp{
				LeftType:   types.AnyArray,
				RightType:  types.AnyArray,
				Volatility: VolatilityImmutable,
			},
			&CmpOp{
				LeftType:   types.INet,
				RightType:  types.INet,
				Volatility: VolatilityImmutable,
			},
		},
		makeBox2DComparisonOperators(
			func(lhs, rhs *geo.CartesianBoundingBox) bool {
				return lhs.Intersects(rhs)
			},
		)...,
	),
})

func makeBox2DComparisonOperators(op func(lhs, rhs *geo.CartesianBoundingBox) bool) cmpOpOverload {
	return cmpOpOverload{
		&CmpOp{
			LeftType:   types.Box2D,
			RightType:  types.Box2D,
			Volatility: VolatilityImmutable,
		},
		&CmpOp{
			LeftType:   types.Box2D,
			RightType:  types.Geometry,
			Volatility: VolatilityImmutable,
		},
		&CmpOp{
			LeftType:   types.Geometry,
			RightType:  types.Box2D,
			Volatility: VolatilityImmutable,
		},
		&CmpOp{
			LeftType:   types.Geometry,
			RightType:  types.Geometry,
			Volatility: VolatilityImmutable,
		},
	}
}

// This map contains the inverses for operators in the CmpOps map that have
// inverses.
var cmpOpsInverse map[ComparisonOperator]ComparisonOperator

func init() {
	cmpOpsInverse = make(map[ComparisonOperator]ComparisonOperator)
	for cmpOpIdx := range comparisonOpName {
		cmpOp := ComparisonOperator(cmpOpIdx)
		newOp, _, _, _, _ := FoldComparisonExpr(cmpOp, DNull, DNull)
		if newOp != cmpOp {
			cmpOpsInverse[newOp] = cmpOp
			cmpOpsInverse[cmpOp] = newOp
		}
	}
}

func makeEvalTupleIn(typ *types.T, v Volatility) *CmpOp {
	return &CmpOp{
		LeftType:     typ,
		RightType:    types.AnyTuple,
		NullableArgs: true,
		Volatility:   v,
	}
}

// MultipleResultsError is returned by QueryRow when more than one result is
// encountered.
type MultipleResultsError struct {
	SQL string // the query that produced this error
}

func (e *MultipleResultsError) Error() string {
	return fmt.Sprintf("%s: unexpected multiple results", e.SQL)
}

// EvalDatabase consists of functions that reference the session database
// and is to be used from EvalContext.
type EvalDatabase interface {
	// ParseQualifiedTableName parses a SQL string of the form
	// `[ database_name . ] [ schema_name . ] table_name`.
	// NB: this is deprecated! Use parser.ParseQualifiedTableName when possible.
	ParseQualifiedTableName(sql string) (*TableName, error)

	// ResolveTableName expands the given table name and
	// makes it point to a valid object.
	// If the database name is not given, it uses the search path to find it, and
	// sets it on the returned TableName.
	// It returns the ID of the resolved table, and an error if the table doesn't exist.
	ResolveTableName(ctx context.Context, tn *TableName) (ID, error)

	// LookupSchema looks up the schema with the given name in the given
	// database.
	LookupSchema(ctx context.Context, dbName, scName string) (found bool, scMeta SchemaMeta, err error)
}

// EvalPlanner is a limited planner that can be used from EvalContext.
type EvalPlanner interface {
	EvalDatabase
	TypeReferenceResolver
	// ParseType parses a column type.
	ParseType(sql string) (*types.T, error)

	// EvalSubquery returns the Datum for the given subquery node.
	EvalSubquery(expr *Subquery) (Datum, error)
}

// EvalSessionAccessor is a limited interface to access session variables.
type EvalSessionAccessor interface {
	// SetConfig sets a session variable to a new value.
	//
	// This interface only supports strings as this is sufficient for
	// pg_catalog.set_config().
	SetSessionVar(ctx context.Context, settingName, newValue string) error

	// GetSessionVar retrieves the current value of a session variable.
	GetSessionVar(ctx context.Context, settingName string, missingOk bool) (bool, string, error)

	// HasAdminRole returns true iff the current session user has the admin role.
	HasAdminRole(ctx context.Context) (bool, error)

	// HasAdminRole returns nil iff the current session user has the specified
	// role option.
	HasRoleOption(ctx context.Context, roleOption roleoption.Option) (bool, error)
}

// ClientNoticeSender is a limited interface to send notices to the
// client.
//
// TODO(knz): as of this writing, the implementations of this
// interface only work on the gateway node (i.e. not from
// distributed processors).
type ClientNoticeSender interface {
	// SendClientNotice sends a notice out-of-band to the client.
	SendClientNotice(ctx context.Context, notice error)
}

// PrivilegedAccessor gives access to certain queries that would otherwise
// require someone with RootUser access to query a given data source.
// It is defined independently to prevent a circular dependency on sql, tree and sqlbase.
type PrivilegedAccessor interface {
	// LookupNamespaceID returns the id of the namespace given it's parent id and name.
	// It is meant as a replacement for looking up the system.namespace directly.
	// Returns the id, a bool representing whether the namespace exists, and an error
	// if there is one.
	LookupNamespaceID(
		ctx context.Context, parentID int64, name string,
	) (DInt, bool, error)

	// LookupZoneConfig returns the zone config given a namespace id.
	// It is meant as a replacement for looking up system.zones directly.
	// Returns the config byte array, a bool representing whether the namespace exists,
	// and an error if there is one.
	LookupZoneConfigByNamespaceID(ctx context.Context, id int64) (DBytes, bool, error)
}

// SequenceOperators is used for various sql related functions that can
// be used from EvalContext.
type SequenceOperators interface {
	EvalDatabase

	// GetSerialSequenceNameFromColumn returns the sequence name for a given table and column
	// provided it is part of a SERIAL sequence.
	// Returns an empty string if the sequence name does not exist.
	GetSerialSequenceNameFromColumn(ctx context.Context, tableName *TableName, columnName Name) (*TableName, error)

	// IncrementSequence increments the given sequence and returns the result.
	// It returns an error if the given name is not a sequence.
	// The caller must ensure that seqName is fully qualified already.
	IncrementSequence(ctx context.Context, seqName *TableName) (int64, error)

	// GetLatestValueInSessionForSequence returns the value most recently obtained by
	// nextval() for the given sequence in this session.
	GetLatestValueInSessionForSequence(ctx context.Context, seqName *TableName) (int64, error)

	// SetSequenceValue sets the sequence's value.
	// If isCalled is false, the sequence is set such that the next time nextval() is called,
	// `newVal` is returned. Otherwise, the next call to nextval will return
	// `newVal + seqOpts.Increment`.
	SetSequenceValue(ctx context.Context, seqName *TableName, newVal int64, isCalled bool) error
}

// TenantOperator is capable of interacting with tenant state, allowing SQL
// builtin functions to create and destroy tenants. The methods will return
// errors when run by any tenant other than the system tenant.
type TenantOperator interface {
	// CreateTenant attempts to install a new tenant in the system. It returns
	// an error if the tenant already exists.
	CreateTenant(ctx context.Context, tenantID uint64, tenantInfo []byte) error

	// DestroyTenant attempts to uninstall an existing tenant from the system.
	// It returns an error if the tenant does not exist.
	DestroyTenant(ctx context.Context, tenantID uint64) error
}

// EvalContextTestingKnobs contains test knobs.
type EvalContextTestingKnobs struct {
	// AssertFuncExprReturnTypes indicates whether FuncExpr evaluations
	// should assert that the returned Datum matches the expected
	// ReturnType of the function.
	AssertFuncExprReturnTypes bool
	// AssertUnaryExprReturnTypes indicates whether UnaryExpr evaluations
	// should assert that the returned Datum matches the expected
	// ReturnType of the function.
	AssertUnaryExprReturnTypes bool
	// AssertBinaryExprReturnTypes indicates whether BinaryExpr
	// evaluations should assert that the returned Datum matches the
	// expected ReturnType of the function.
	AssertBinaryExprReturnTypes bool
	// DisableOptimizerRuleProbability is the probability that any given
	// transformation rule in the optimizer is disabled.
	DisableOptimizerRuleProbability float64
	// OptimizerCostPerturbation is used to randomly perturb the estimated
	// cost of each expression in the query tree for the purpose of creating
	// alternate query plans in the optimizer.
	OptimizerCostPerturbation float64
}

// FoldComparisonExpr folds a given comparison operation and its expressions
// into an equivalent operation that will hit in the CmpOps map, returning
// this new operation, along with potentially flipped operands and "flipped"
// and "not" flags.
func FoldComparisonExpr(
	op ComparisonOperator, left, right Expr,
) (newOp ComparisonOperator, newLeft Expr, newRight Expr, flipped bool, not bool) {
	switch op {
	case NE:
		// NE(left, right) is implemented as !EQ(left, right).
		return EQ, left, right, false, true
	case GT:
		// GT(left, right) is implemented as LT(right, left)
		return LT, right, left, true, false
	case GE:
		// GE(left, right) is implemented as LE(right, left)
		return LE, right, left, true, false
	case NotIn:
		// NotIn(left, right) is implemented as !IN(left, right)
		return In, left, right, false, true
	case NotLike:
		// NotLike(left, right) is implemented as !Like(left, right)
		return Like, left, right, false, true
	case NotILike:
		// NotILike(left, right) is implemented as !ILike(left, right)
		return ILike, left, right, false, true
	case NotSimilarTo:
		// NotSimilarTo(left, right) is implemented as !SimilarTo(left, right)
		return SimilarTo, left, right, false, true
	case NotRegMatch:
		// NotRegMatch(left, right) is implemented as !RegMatch(left, right)
		return RegMatch, left, right, false, true
	case NotRegIMatch:
		// NotRegIMatch(left, right) is implemented as !RegIMatch(left, right)
		return RegIMatch, left, right, false, true
	case IsDistinctFrom:
		// IsDistinctFrom(left, right) is implemented as !IsNotDistinctFrom(left, right)
		// Note: this seems backwards, but IS NOT DISTINCT FROM is an extended
		// version of IS and IS DISTINCT FROM is an extended version of IS NOT.
		return IsNotDistinctFrom, left, right, false, true
	}
	return op, left, right, false, false
}
