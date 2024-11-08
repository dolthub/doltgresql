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
	"github.com/dolthub/doltgresql/server/auth"
)

// nodeTruncate handles *tree.Truncate nodes.
func nodeTruncate(ctx *Context, node *tree.Truncate) (*vitess.DDL, error) {
	if node == nil || len(node.Tables) == 0 {
		return nil, nil
	}
	switch node.DropBehavior {
	case tree.DropDefault:
		// Default behavior, nothing to do
	case tree.DropRestrict:
		return nil, fmt.Errorf("RESTRICT is not yet supported")
	case tree.DropCascade:
		return nil, fmt.Errorf("CASCADE is not yet supported")
	}
	if len(node.Tables) > 1 {
		return nil, fmt.Errorf("truncating multiple tables at once is not yet supported")
	}
	tableName, err := nodeTableName(ctx, &node.Tables[0])
	if err != nil {
		return nil, err
	}
	return &vitess.DDL{
		Action: vitess.TruncateStr,
		Table:  tableName,
		Auth: vitess.AuthInformation{
			AuthType:    auth.AuthType_TRUNCATE,
			TargetType:  auth.AuthTargetType_SingleTableIdentifier,
			TargetNames: []string{tableName.SchemaQualifier.String(), tableName.Name.String()},
		},
	}, nil
}
