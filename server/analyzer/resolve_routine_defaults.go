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

package analyzer

import (
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/extensions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ResolveProcedureDefaults resolves default expressions of routines that are in string format by parsing it into sql.Expression.
// This function retrieves the procedure overloads and sets CompiledFunction in the Call node.
func ResolveProcedureDefaults(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	switch n := node.(type) {
	case *pgnodes.Call:
		procCollection, err := core.GetProceduresCollectionFromContext(ctx)
		if err != nil {
			return nil, transform.SameTree, err
		}
		typesCollection, err := core.GetTypesCollectionFromContext(ctx)
		if err != nil {
			return nil, transform.SameTree, err
		}
		schemaName, err := core.GetSchemaName(ctx, nil, n.SchemaName)
		if err != nil {
			return nil, transform.SameTree, err
		}
		procName := id.NewProcedure(schemaName, n.ProcedureName)
		overloads, err := procCollection.GetProcedureOverloads(ctx, procName)
		if err != nil {
			return nil, transform.SameTree, err
		}
		if len(overloads) == 0 {
			if strings.HasPrefix(n.ProcedureName, "dolt_") {
				return nil, transform.SameTree, functions.ErrDoltProcedureSelectOnly
			}
			return nil, transform.SameTree, sql.ErrStoredProcedureDoesNotExist.New(n.ProcedureName)
		}

		same := transform.SameTree
		overloadTree := framework.NewOverloads()
		for _, overload := range overloads {
			paramTypes := make([]*pgtypes.DoltgresType, len(overload.ParameterTypes))
			for i, paramType := range overload.ParameterTypes {
				paramTypes[i], err = typesCollection.GetType(ctx, paramType)
				if err != nil || paramTypes[i] == nil {
					return nil, transform.SameTree, err
				}
			}
			// TODO: we should probably have procedure equivalents instead of converting these to functions
			//  probably fine for now since we don't implement/support the differing functionality between the two just yet
			if len(overload.ExtensionName) > 0 {
				if err = overloadTree.Add(framework.CFunction{
					ID:                 id.Function(overload.ID),
					ReturnType:         pgtypes.Void,
					ParameterTypes:     paramTypes,
					Variadic:           false,
					IsNonDeterministic: true,
					Strict:             false,
					ExtensionName:      extensions.LibraryIdentifier(overload.ExtensionName),
					ExtensionSymbol:    overload.ExtensionSymbol,
				}); err != nil {
					return nil, transform.SameTree, err
				}
			} else if len(overload.SQLDefinition) > 0 {
				if err = overloadTree.Add(framework.SQLFunction{
					ID:                 id.Function(overload.ID),
					ReturnType:         pgtypes.Void,
					ParameterNames:     overload.ParameterNames,
					ParameterTypes:     paramTypes,
					ParameterDefaults:  overload.ParameterDefaults,
					Variadic:           false,
					IsNonDeterministic: true,
					Strict:             false,
					SqlStatement:       overload.SQLDefinition,
					SetOf:              false,
				}); err != nil {
					return nil, transform.SameTree, err
				}
			} else {
				if err = overloadTree.Add(framework.InterpretedFunction{
					ID:                 id.Function(overload.ID),
					ReturnType:         pgtypes.Void,
					ParameterNames:     overload.ParameterNames,
					ParameterTypes:     paramTypes,
					Variadic:           false,
					IsNonDeterministic: true,
					Strict:             false,
					Statements:         overload.Operations,
				}); err != nil {
					return nil, transform.SameTree, err
				}
			}
		}
		compiledFunction := framework.NewCompiledFunction(ctx, n.ProcedureName, n.Exprs, overloadTree, false)
		// fill in default exprs if applicable
		if err := compiledFunction.ResolveDefaultValues(ctx, func(defExpr string) (sql.Expression, error) {
			return getDefaultExpr(ctx, a.Catalog, defExpr)
		}); err != nil {
			return nil, transform.SameTree, err
		}
		n.CompiledFunc = compiledFunction
		return node, same, nil
	default:
		return node, transform.SameTree, nil
	}
}
