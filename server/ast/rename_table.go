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
)

// nodeRenameTable handles *tree.RenameTable nodes.
func nodeRenameTable(ctx *Context, node *tree.RenameTable) (*vitess.DDL, error) {
	if node == nil {
		return nil, nil
	}
	if node.IsSequence {
		return nil, fmt.Errorf("RENAME SEQUENCE is not yet supported")
	}
	if node.IsMaterialized {
		return nil, fmt.Errorf("RENAME MATERIALIZED VIEW is not yet supported")
	}
	fromName, err := nodeUnresolvedObjectName(ctx, node.Name)
	if err != nil {
		return nil, err
	}
	toName, err := nodeUnresolvedObjectName(ctx, node.NewName)
	if err != nil {
		return nil, err
	}
	return &vitess.DDL{
		Action:     vitess.RenameStr,
		FromTables: vitess.TableNames{fromName},
		ToTables:   vitess.TableNames{toName},
		IfExists:   node.IfExists,
	}, nil
}
