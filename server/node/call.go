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

package node

import (
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/extensions"
	"github.com/dolthub/doltgresql/core/id"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Call is used to call stored procedures.
type Call struct {
	SchemaName    string
	ProcedureName string
	Exprs         []sql.Expression
	Runner        pgexprs.StatementRunner
	cachedSch     sql.Schema
	originalExprs vitess.Exprs
}

var _ sql.ExecSourceRel = (*Call)(nil)
var _ sql.Expressioner = (*Call)(nil)
var _ vitess.Injectable = (*Call)(nil)

// NewCall returns a new *Call.
func NewCall(schema string, name string, originalExprs vitess.Exprs) *Call {
	return &Call{
		SchemaName:    schema,
		ProcedureName: name,
		Exprs:         nil,
		originalExprs: originalExprs,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *Call) Children() []sql.Node {
	return nil
}

// Expressions implements the interface sql.Expressioner.
func (c *Call) Expressions() []sql.Expression {
	exprs := make([]sql.Expression, len(c.Exprs)+1)
	exprs[0] = c.Runner
	copy(exprs[1:], c.Exprs)
	return exprs
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *Call) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *Call) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *Call) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	if !core.IsContextValid(ctx) {
		return nil, errors.New("invalid context while attempting to call a procedure")
	}
	procCollection, err := core.GetProceduresCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	typesCollection, err := core.GetTypesCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	schemaName, err := core.GetSchemaName(ctx, nil, c.SchemaName)
	if err != nil {
		return nil, err
	}
	procName := id.NewProcedure(schemaName, c.ProcedureName)
	overloads, err := procCollection.GetProcedureOverloads(ctx, procName)
	if err != nil {
		return nil, err
	}
	if len(overloads) == 0 {
		// We're going to assume that this is calling one of the few remaining Dolt stored procedures
		sch, rowIter, _, err := c.Runner.Runner.QueryWithBindings(ctx, "", &vitess.Call{
			ProcName: vitess.ProcedureName{
				Name:      vitess.NewColIdent(c.ProcedureName),
				Qualifier: vitess.NewTableIdent(c.SchemaName),
			},
			Params: c.originalExprs,
		}, nil, nil)
		c.cachedSch = sch
		return rowIter, err
	}

	overloadTree := framework.NewOverloads()
	for _, overload := range overloads {
		paramTypes := make([]*pgtypes.DoltgresType, len(overload.ParameterTypes))
		for i, paramType := range overload.ParameterTypes {
			paramTypes[i], err = typesCollection.GetType(ctx, paramType)
			if err != nil || paramTypes[i] == nil {
				return nil, err
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
				return nil, err
			}
		} else if len(overload.SQLDefinition) > 0 {
			if err = overloadTree.Add(framework.SQLFunction{
				ID:                 id.Function(overload.ID),
				ReturnType:         pgtypes.Void,
				ParameterNames:     overload.ParameterNames,
				ParameterTypes:     paramTypes,
				Variadic:           false,
				IsNonDeterministic: true,
				Strict:             false,
				SqlStatement:       overload.SQLDefinition,
				SetOf:              false,
			}); err != nil {
				return nil, err
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
				return nil, err
			}
		}
	}
	compiledFunc := framework.NewCompiledFunction(c.ProcedureName, c.Exprs, overloadTree, false)
	compiledFunc = compiledFunc.SetStatementRunner(ctx, c.Runner.Runner).(*framework.CompiledFunction)
	_, err = compiledFunc.Eval(ctx, nil)
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *Call) Schema() sql.Schema {
	// TODO: this should be the INOUT and OUT parameters of the target procedure assuming we're not using the cached schema
	return c.cachedSch
}

// String implements the interface sql.ExecSourceRel.
func (c *Call) String() string {
	return "CALL"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *Call) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithExpressions implements the interface sql.Expressioner.
func (c *Call) WithExpressions(exprs ...sql.Expression) (sql.Node, error) {
	if len(c.Exprs)+1 != len(exprs) {
		return nil, errors.Errorf("expected `%d` child expressions but received `%d`", len(c.Exprs), len(exprs))
	}
	nc := *c
	nc.Runner = exprs[0].(pgexprs.StatementRunner)
	nc.Exprs = exprs[1:]
	return &nc, nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *Call) WithResolvedChildren(children []any) (any, error) {
	resolvedChildren := make([]sql.Expression, len(children))
	for i, child := range children {
		var ok bool
		resolvedChildren[i], ok = child.(sql.Expression)
		if !ok {
			return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", child)
		}
	}
	nc := *c
	nc.Exprs = resolvedChildren
	return &nc, nil
}
