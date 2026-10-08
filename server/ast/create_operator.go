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
	"github.com/cockroachdb/errors"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// nodeCreateOperator handles *tree.CreateOperator nodes.
func nodeCreateOperator(ctx *Context, node *tree.CreateOperator) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	var function *tree.UnresolvedObjectName
	var leftArg, rightArg tree.ResolvableTypeReference
	var commutator, negator string
	var hashes, merges bool
	for _, option := range node.Options {
		switch option.Option {
		case tree.OperatorOptTypeFunction:
			function = option.FuncName
		case tree.OperatorOptTypeLeftArg:
			leftArg = option.TypeVal
		case tree.OperatorOptTypeRightArg:
			rightArg = option.TypeVal
		case tree.OperatorOptTypeCommutator:
			commutator = tree.OperatorSymbol(option.OpVal)
		case tree.OperatorOptTypeNegator:
			negator = tree.OperatorSymbol(option.OpVal)
		case tree.OperatorOptTypeRestrict:
			return NotYetSupportedError("RESTRICT is not yet supported")
		case tree.OperatorOptTypeJoin:
			return NotYetSupportedError("JOIN is not yet supported")
		case tree.OperatorOptTypeHashes:
			hashes = true
		case tree.OperatorOptTypeMerges:
			merges = true
		}
	}
	if function == nil {
		return nil, errors.New("operator function must be specified")
	}
	if leftArg == nil && rightArg == nil {
		return nil, errors.New("operator argument types must be specified")
	}
	if rightArg == nil {
		return nil, errors.New("operator right argument type must be specified")
	}
	var err error
	var leftType *pgtypes.DoltgresType
	if leftArg != nil {
		_, leftType, err = nodeResolvableTypeReference(ctx, leftArg, false)
		if err != nil {
			return nil, err
		}
	}
	_, rightType, err := nodeResolvableTypeReference(ctx, rightArg, false)
	if err != nil {
		return nil, err
	}
	funcName := function.ToTableName()
	return vitess.InjectedStatement{
		Statement: pgnodes.NewCreateOperator(
			tree.OperatorSymbol(node.Name),
			leftType,
			rightType,
			funcName.Schema(),
			funcName.Table(),
			commutator,
			negator,
			hashes,
			merges,
		),
		Auth: vitess.AuthInformation{
			AuthType:    auth.AuthType_CREATE,
			TargetType:  auth.AuthTargetType_TODO,
			TargetNames: []string{},
		},
	}, nil
}
