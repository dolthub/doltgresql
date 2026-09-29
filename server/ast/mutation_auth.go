// Copyright 2026 Dolthub, Inc.
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
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/auth"
)

// mutationReadsTarget checks expressions that consume target values, excluding
// assignment destinations. Doltgres has table-level rather than column-level ACLs.
// TODO: Resolve unqualified columns in subqueries against their local scopes;
// currently they conservatively require SELECT on the mutation target as well.
func mutationReadsTarget(name, alias string, nodes ...vitess.SQLNode) bool {
	reads := false
	_ = vitess.Walk(func(node vitess.SQLNode) (bool, error) {
		var qualifier string
		switch node := node.(type) {
		case *vitess.ColName:
			qualifier = node.Qualifier.Name.String()
		case *vitess.StarExpr:
			qualifier = node.TableName.Name.String()
		case *vitess.AssignmentExpr:
			if mutationReadsTarget(name, alias, node.Expr) {
				reads = true
			}
			return false, nil
		default:
			return true, nil
		}
		if qualifier == "" || qualifier == name || qualifier == alias || qualifier == "excluded" {
			reads = true
		}
		return false, nil
	}, nodes...)
	return reads
}

// authorizeMutationReads adds SELECT to the mutation target's required rights.
func authorizeMutationReads(table vitess.TableExpr, nodes ...vitess.SQLNode) {
	target, ok := table.(*vitess.AliasedTableExpr)
	if !ok {
		return
	}
	name, ok := target.Expr.(vitess.TableName)
	if ok && mutationReadsTarget(name.Name.String(), target.As.String(), nodes...) {
		target.Auth.Extra = auth.AdditionalTablePrivileges{auth.Privilege_SELECT}
	}
}
