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
	"strings"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeAnalyze handles *tree.Analyze nodes.
func nodeAnalyze(ctx *Context, node *tree.Analyze) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}

	// None of the options affect how statistics are collected here, so they are validated and then ignored.
	for _, option := range node.Options {
		switch option.Option {
		case "VERBOSE", "SKIP_LOCKED", "BUFFER_USAGE_LIMIT":
		default:
			return nil, pgerror.Newf(pgcode.Syntax, `unrecognized ANALYZE option "%s"`, strings.ToLower(option.Option))
		}
	}

	// If no tables were specified, return an empty Analyze statement, and the analyzer will populate all tables to be
	// analyzed. A column list still analyzes the whole table.
	tables := make(vitess.TableNames, len(node.TablesAndCols))
	for i, table := range node.TablesAndCols {
		tables[i] = vitess.TableName{
			Name:            vitess.NewTableIdent(table.Name.Object()),
			SchemaQualifier: vitess.NewTableIdent(table.Name.Schema()),
			DbQualifier:     vitess.NewTableIdent(table.Name.Catalog()),
		}
	}

	return &vitess.Analyze{Tables: tables}, nil
}
