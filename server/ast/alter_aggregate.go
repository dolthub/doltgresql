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

// nodeAlterAggregate handles *tree.AlterAggregate nodes.
func nodeAlterAggregate(ctx *Context, node *tree.AlterAggregate) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	if err := validateAggArgMode(ctx, node.AggSig.Args, node.AggSig.OrderByArgs); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("ALTER AGGREGATE is not yet supported")
}
