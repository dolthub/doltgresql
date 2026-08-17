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
	"fmt"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// nodeDropAggregate handles *tree.DropAggregate nodes.
func nodeDropAggregate(ctx *Context, node *tree.DropAggregate) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	if node.DropBehavior == tree.DropCascade {
		return nil, fmt.Errorf("DROP AGGREGATE with CASCADE is not supported yet")
	}
	aggs := make([]*pgnodes.AggregateToDrop, len(node.Aggregates))
	for i, agg := range node.Aggregates {
		if err := validateAggArgMode(ctx, agg.AggSig.Args, agg.AggSig.OrderByArgs); err != nil {
			return nil, err
		}
		if agg.AggSig.All {
			return NotYetSupportedError("DROP AGGREGATE with a * signature is not yet supported")
		}
		if len(agg.AggSig.OrderByArgs) > 0 {
			return NotYetSupportedError("ordered-set aggregates are not yet supported")
		}
		argTypes := make([]*pgtypes.DoltgresType, len(agg.AggSig.Args))
		for j, arg := range agg.AggSig.Args {
			_, argType, err := nodeResolvableTypeReference(ctx, arg.Type, false)
			if err != nil {
				return nil, err
			}
			argTypes[j] = argType
		}
		name := agg.Name.ToTableName()
		aggs[i] = &pgnodes.AggregateToDrop{
			SchemaName: name.Schema(),
			Name:       name.Table(),
			ArgTypes:   argTypes,
		}
	}
	return vitess.InjectedStatement{
		Statement: pgnodes.NewDropAggregate(node.IfExists, aggs),
		Children:  nil,
	}, nil
}
