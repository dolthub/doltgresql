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

package ast

import (
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeCreateCast handles *tree.CreateCast nodes.
func nodeCreateCast(ctx *Context, node *tree.CreateCast) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	_, sourceType, err := nodeResolvableTypeReference(ctx, node.Source, false)
	if err != nil {
		return nil, err
	}
	_, targetType, err := nodeResolvableTypeReference(ctx, node.Target, false)
	if err != nil {
		return nil, err
	}
	var funcName tree.TableName
	var funcParams []pgnodes.RoutineParam
	switch node.Type {
	case tree.CreateCastType_WithFunction:
		funcName = node.FuncName.ToTableName()
		funcParams = make([]pgnodes.RoutineParam, len(node.FuncArgs))
		for i, arg := range node.FuncArgs {
			funcParams[i].Name = arg.Name.String()
			_, funcParams[i].Type, err = nodeResolvableTypeReference(ctx, arg.Type, false)
			if err != nil {
				return nil, err
			}
		}
	case tree.CreateCastType_WithoutFunction:
		// TODO: restrict to superusers only with error: "must be superuser to create a cast WITHOUT FUNCTION"
	case tree.CreateCastType_Inout:
		// Nothing to do
	default:
		panic("unhandled case")
	}
	var scope casts.CastType
	switch node.Scope {
	case tree.CreateCastScope_Explicit:
		scope = casts.CastType_Explicit
	case tree.CreateCastScope_Assignment:
		scope = casts.CastType_Assignment
	case tree.CreateCastScope_Implicit:
		scope = casts.CastType_Implicit
	}
	return vitess.InjectedStatement{
		Statement: pgnodes.NewCreateCast(
			sourceType,
			targetType,
			scope,
			node.Type,
			funcName.Schema(),
			funcName.Table(),
			funcParams,
		),
		Auth: vitess.AuthInformation{
			AuthType:    auth.AuthType_CREATE,
			TargetType:  auth.AuthTargetType_TODO,
			TargetNames: []string{},
		},
	}, nil
}
