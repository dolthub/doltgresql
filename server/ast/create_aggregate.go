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

package ast

import (
	"github.com/cockroachdb/errors"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// nodeCreateAggregate handles *tree.CreateAggregate nodes.
func nodeCreateAggregate(ctx *Context, node *tree.CreateAggregate) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	if err := validateAggArgMode(ctx, node.Args, node.OrderByArgs); err != nil {
		return nil, err
	}
	if len(node.OrderByArgs) > 0 {
		return NotYetSupportedError("ordered-set aggregates are not yet supported")
	}
	var argTypeRefs []tree.ResolvableTypeReference
	if node.Args != nil {
		for _, arg := range node.Args {
			if arg.Mode == tree.RoutineArgModeVariadic {
				return NotYetSupportedError("VARIADIC aggregates are not yet supported")
			}
			argTypeRefs = append(argTypeRefs, arg.Type)
		}
	} else {
		argTypeRefs = append(argTypeRefs, node.BaseType)
	}
	argTypes := make([]*pgtypes.DoltgresType, len(argTypeRefs))
	for i, argTypeRef := range argTypeRefs {
		_, argType, err := nodeResolvableTypeReference(ctx, argTypeRef, false)
		if err != nil {
			return nil, err
		}
		argTypes[i] = argType
	}
	_, sType, err := nodeResolvableTypeReference(ctx, node.SType, false)
	if err != nil {
		return nil, err
	}
	var finalFunc, combineFunc *tree.UnresolvedObjectName
	var initCond string
	var hasInitCond bool
	for _, option := range node.AggOptions {
		switch option.Option {
		case tree.AggOptTypeFinalFunc:
			finalFunc = option.FuncName
		case tree.AggOptTypeCombineFunc:
			combineFunc = option.FuncName
		case tree.AggOptTypeInitCond:
			if strVal, ok := option.CondVal.(*tree.StrVal); ok {
				initCond = strVal.RawString()
			} else {
				initCond = tree.AsString(option.CondVal)
			}
			hasInitCond = true
		default:
			return NotYetSupportedError("the given aggregate option is not yet supported")
		}
	}
	var finalFuncSchema, finalFuncName string
	if finalFunc != nil {
		finalFuncTableName := finalFunc.ToTableName()
		finalFuncSchema = finalFuncTableName.Schema()
		finalFuncName = finalFuncTableName.Table()
	}
	var combineFuncSchema, combineFuncName string
	if combineFunc != nil {
		combineFuncTableName := combineFunc.ToTableName()
		combineFuncSchema = combineFuncTableName.Schema()
		combineFuncName = combineFuncTableName.Table()
	}
	name := node.Name.ToTableName()
	sFuncName := node.SFunc.ToTableName()
	return vitess.InjectedStatement{
		Statement: pgnodes.NewCreateAggregate(
			name.Schema(),
			name.Table(),
			node.Replace,
			argTypes,
			sType,
			sFuncName.Schema(),
			sFuncName.Table(),
			finalFuncSchema,
			finalFuncName,
			combineFuncSchema,
			combineFuncName,
			initCond,
			hasInitCond,
		),
		Auth: vitess.AuthInformation{
			AuthType:    auth.AuthType_CREATE,
			TargetType:  auth.AuthTargetType_TODO,
			TargetNames: []string{},
		},
	}, nil
}

// validateAggArgMode checks routine arguments for `OUT` and `INOUT` modes,
// which cannot be used for AGGREGATE arguments.
func validateAggArgMode(ctx *Context, args, orderByArgs tree.RoutineArgs) error {
	for _, sig := range args {
		if sig.Mode == tree.RoutineArgModeOut || sig.Mode == tree.RoutineArgModeInout {
			return errors.Errorf("aggregates cannot have output arguments")
		}
	}
	for _, sig := range orderByArgs {
		if sig.Mode == tree.RoutineArgModeOut || sig.Mode == tree.RoutineArgModeInout {
			return errors.Errorf("aggregates cannot have output arguments")
		}
	}
	return nil
}
