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
)

// nodeCreateAggregate handles *tree.CreateAggregate nodes.
func nodeCreateAggregate(ctx *Context, node *tree.CreateAggregate) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	if err := validateAggArgMode(ctx, node.Args, node.OrderByArgs); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("CREATE AGGREGATE is not yet supported")
}

// validateAggArgMode checks routine arguments for `OUT` and `INOUT` modes,
// which cannot be used for AGGREGATE arguments.
func validateAggArgMode(ctx *Context, args, orderByArgs tree.RoutineArgs) error {
	for _, sig := range args {
		if sig.Mode == tree.RoutineArgModeOut || sig.Mode == tree.RoutineArgModeInout {
			return fmt.Errorf("aggregate functions do not support OUT or INOUT arguments")
		}
	}
	for _, sig := range orderByArgs {
		if sig.Mode == tree.RoutineArgModeOut || sig.Mode == tree.RoutineArgModeInout {
			return fmt.Errorf("aggregate functions do not support OUT or INOUT arguments")
		}
	}
	return nil
}
