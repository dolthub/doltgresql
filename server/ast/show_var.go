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
	"strings"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeShowVar handles *tree.ShowVar nodes.
func nodeShowVar(node *tree.ShowVar) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}

	if strings.ToLower(node.Name) == "is_superuser" {
		return nil, fmt.Errorf("SHOW IS_SUPERUSER is not yet supported")
	} else if strings.ToLower(node.Name) == "all" {
		// TODO: need this soon
		return nil, fmt.Errorf("SHOW ALL is not yet supported")
	}

	// TODO: this is a temporary way to get the param value for the current implementation
	//   need better way to get these info
	s := &vitess.Select{
		SelectExprs: vitess.SelectExprs{
			&vitess.AliasedExpr{
				Expr: &vitess.ColName{
					Name:      vitess.NewColIdent("@@session." + node.Name),
					Qualifier: vitess.TableName{},
				},
				StartParsePos: 7,
				EndParsePos:   17 + len(node.Name),
				As:            vitess.NewColIdent(node.Name),
			},
		},
	}
	return s, nil
}
