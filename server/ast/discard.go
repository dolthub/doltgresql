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

package ast

import (
	"fmt"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/node"
)

// nodeDiscard handles *tree.Discard nodes.
func nodeDiscard(ctx *Context, discard *tree.Discard) (vitess.Statement, error) {
	if discard == nil {
		return nil, nil
	}
	if discard.Mode != tree.DiscardModeAll {
		return nil, fmt.Errorf("unhandled DISCARD mode: %v", discard.Mode)
	}

	return vitess.InjectedStatement{
		Statement: node.DiscardStatement{},
	}, nil
}
