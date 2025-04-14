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
	"github.com/cockroachdb/errors"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeAnalyze handles *tree.Analyze nodes.
func nodeAnalyze(ctx *Context, node *tree.Analyze) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}

	// If no tables were specified, return an empty Analyze statement, and the analyzer
	// will populate all tables to be analyzed.
	if node.Table == nil {
		return &vitess.Analyze{}, nil
	}

	objectName, ok := node.Table.(*tree.UnresolvedObjectName)
	if !ok {
		return nil, errors.Errorf("unsupported table type in Analyze node: %T", node.Table)
	}

	return &vitess.Analyze{Tables: []vitess.TableName{
		{
			Name:            vitess.NewTableIdent(objectName.Object()),
			SchemaQualifier: vitess.NewTableIdent(objectName.Schema()),
			DbQualifier:     vitess.NewTableIdent(objectName.Catalog()),
		},
	}}, nil
}
