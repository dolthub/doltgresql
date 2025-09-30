// Copyright 2025 Dolthub, Inc.
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

package analyzer

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"
	"github.com/dolthub/go-mysql-server/sql/transform"

	psql "github.com/dolthub/doltgresql/postgres/parser/parser/sql"
	"github.com/dolthub/doltgresql/server/node"
)

// ValidateCreateFunction validates that a function can be created as specified. It validates functions defined
// with SQL language.
func ValidateCreateFunction(ctx *sql.Context, a *analyzer.Analyzer, n sql.Node, scope *plan.Scope, sel analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	ct, ok := n.(*node.CreateFunction)
	if !ok {
		return n, transform.SameTree, nil
	}

	if len(ct.SqlDef) == 0 {
		return n, transform.SameTree, nil
	}

	parser := psql.NewPostgresParser()
	builder := planbuilder.New(ctx, a.Catalog, nil, parser)
	_, _, err := builder.BindOnly(ct.SqlDefParsed, ct.SqlDef, nil)
	if err != nil {
		return nil, transform.SameTree, err
	}

	return n, transform.SameTree, nil
}
