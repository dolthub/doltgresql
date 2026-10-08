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

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeDropCast handles *tree.DropCast nodes.
func nodeDropCast(ctx *Context, node *tree.DropCast) (vitess.Statement, error) {
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
	return vitess.InjectedStatement{
		Statement: pgnodes.NewDropCast(
			sourceType,
			targetType,
			node.IfExists,
		),
		Auth: vitess.AuthInformation{
			AuthType:    auth.AuthType_DELETE,
			TargetType:  auth.AuthTargetType_TODO,
			TargetNames: []string{},
		},
	}, nil
}
