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
	"bytes"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/lex"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/config"
)

// nodeSetVar handles *tree.SetVar nodes.
func nodeSetVar(ctx *Context, node *tree.SetVar) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	// USE statement alias
	if node.Name == "database" {
		// strip off all quotes from the database name
		dbName := strings.TrimPrefix(strings.TrimSuffix(node.Values[0].String(), "'"), "'")
		dbName = strings.TrimPrefix(strings.TrimSuffix(dbName, "\""), "\"")
		dbName = strings.TrimPrefix(strings.TrimSuffix(dbName, "`"), "`")
		return &vitess.Use{DBName: vitess.NewTableIdent(dbName)}, nil
	}
	if node.Namespace == "" && !config.IsValidPostgresConfigParameter(node.Name) && !config.IsValidDoltConfigParameter(node.Name) {
		return nil, errors.Errorf(`ERROR: unrecognized configuration parameter "%s"`, node.Name)
	}
	if node.IsLocal && node.Namespace != "" {
		// TODO: support transaction-local values for custom (namespaced) parameters, which are session user vars
		return nil, errors.Errorf("SET LOCAL is not yet supported for customized options")
	}
	var expr vitess.Expr
	var err error
	if len(node.Values) == 0 {
		// sanity check
		return nil, errors.Errorf(`ERROR: syntax error at or near ";"'`)
	} else if flattened, ok := flattenIdentifierList(node.Name, node.Values); ok {
		expr = vitess.NewStrVal([]byte(flattened))
	} else if len(node.Values) > 1 {
		vals := make([]string, len(node.Values))
		for i, val := range node.Values {
			vals[i] = val.String()
		}
		expr = &vitess.ColName{
			Name: vitess.NewColIdent(strings.Join(vals, ", ")),
		}
	} else {
		expr, err = nodeExpr(ctx, node.Values[0])
		if err != nil {
			return nil, err
		}
	}

	if node.Namespace == "" {
		// Postgres's SET has no GLOBAL scope syntax, so system variables that have no session scope (e.g.
		// Dolt's cluster replication variables) are routed to their declared scope directly, symmetric with
		// current_setting() reading them from the global scope.
		scope := vitess.SetScope_Session
		if node.IsLocal {
			// SET LOCAL only applies for the duration of the current transaction; the connection handler restores
			// the session values when the transaction ends. Global-only variables are rejected by the engine.
			scope = planbuilder.SetScope_TransactionLocal
		} else if svScope, ok := config.GlobalOnlySystemVariableScope(node.Name); ok {
			switch svScope {
			case sql.SystemVariableScope_Persist:
				scope = vitess.SetScope_Persist
			case sql.SystemVariableScope_PersistOnly:
				scope = vitess.SetScope_PersistOnly
			default:
				scope = vitess.SetScope_Global
			}
		}
		return &vitess.Set{
			Exprs: vitess.SetVarExprs{&vitess.SetVarExpr{
				Scope: scope,
				Name: &vitess.ColName{
					Name: vitess.NewColIdent(node.Name),
				},
				Expr: expr,
			}},
		}, nil
	} else {
		return &vitess.Set{
			Exprs: vitess.SetVarExprs{&vitess.SetVarExpr{
				Scope: vitess.SetScope_User,
				Name: &vitess.ColName{
					Name: vitess.NewColIdent(fmt.Sprintf("%s.%s", node.Namespace, node.Name)),
				},
				Expr: expr,
			}},
		}, nil
	}
}

// flattenIdentifierList renders the values assigned to a parameter whose value is a comma separated list of
// identifiers. It matches the behavior of Postgres's flatten_set_variable_args with a GUC_LIST_QUOTE parameter. Every element
// is written out with quote_identifier. An element that needs no quoting loses any quotes it was typed with and one
// that does need it keeps them.
//
// The second return is false when the parameter takes a plain string rather than a list of identifiers, or when a
// value is something other than a name or a literal. The caller is expected to interpret |values| appropriately
// in such a case.
func flattenIdentifierList(configParameterName string, values tree.Exprs) (string, bool) {
	if !config.IsListQuoteConfigParameter(configParameterName) {
		return "", false
	}
	var buf bytes.Buffer
	for i, value := range values {
		if i > 0 {
			buf.WriteString(", ")
		}
		switch value := value.(type) {
		case *tree.UnresolvedName:
			if value.NumParts != 1 || value.Star {
				return "", false
			}
			lex.EncodeRestrictedSQLIdent(&buf, value.Parts[0], 0)
		case *tree.StrVal:
			lex.EncodeRestrictedSQLIdent(&buf, value.RawString(), 0)
		case *tree.NumVal:
			// A number is written out as itself: Postgres does not quote one.
			buf.WriteString(value.FormattedString())
		default:
			return "", false
		}
	}
	return buf.String(), true
}
