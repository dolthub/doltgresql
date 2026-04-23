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
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/extensions"
	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/procedures"
	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// RoutineParam represents a routine parameter with parameter name, type and default value if exists.
type RoutineParam struct {
	Mode       procedures.ParameterMode
	Name       string
	Type       *pgtypes.DoltgresType
	HasDefault bool
	Default    sql.Expression
}

// CreateFunction implements CREATE FUNCTION.
type CreateFunction struct {
	FunctionName      string
	SchemaName        string
	Replace           bool
	ReturnType        *pgtypes.DoltgresType
	Parameters        []RoutineParam
	Strict            bool
	Statements        []plpgsql.InterpreterOperation
	ExtensionName     string
	ExtensionSymbol   string
	Definition        string
	SqlDef            string
	SqlDefParsedStmts []vitess.Statement
	SetOf             bool
}

var _ sql.ExecSourceRel = (*CreateFunction)(nil)
var _ vitess.Injectable = (*CreateFunction)(nil)

// NewCreateFunction returns a new *CreateFunction.
func NewCreateFunction(
	functionName string,
	schemaName string,
	replace bool,
	retType *pgtypes.DoltgresType,
	params []RoutineParam,
	strict bool,
	definition string,
	extensionName string,
	extensionSymbol string,
	statements []plpgsql.InterpreterOperation,
	sqlDef string,
	sqlDefParsedStmts []vitess.Statement,
	setOf bool) *CreateFunction {
	return &CreateFunction{
		FunctionName:      functionName,
		SchemaName:        schemaName,
		Replace:           replace,
		ReturnType:        retType,
		Parameters:        params,
		Strict:            strict,
		Statements:        statements,
		ExtensionName:     extensionName,
		ExtensionSymbol:   extensionSymbol,
		Definition:        definition,
		SqlDef:            sqlDef,
		SqlDefParsedStmts: sqlDefParsedStmts,
		SetOf:             setOf,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *CreateFunction) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CreateFunction) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CreateFunction) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CreateFunction) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	funcCollection, err := core.GetFunctionsCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	paramNames := make([]string, len(c.Parameters))
	paramTypes := make([]id.Type, len(c.Parameters))
	paramDefaults := make([]string, len(c.Parameters))
	for i, param := range c.Parameters {
		paramNames[i] = param.Name
		paramTypes[i] = param.Type.ID
		if param.Default != nil {
			paramDefaults[i] = param.Default.String()
		}
	}

	schemaName, err := core.GetSchemaName(ctx, nil, c.SchemaName)
	if err != nil {
		return nil, err
	}
	funcID := id.NewFunction(schemaName, c.FunctionName, paramTypes...)
	if c.Replace && funcCollection.HasFunction(ctx, funcID) {
		if err = funcCollection.DropFunction(ctx, funcID); err != nil {
			return nil, err
		}
	}
	var extName string
	if len(c.ExtensionName) > 0 {
		ext, err := extensions.GetExtension(c.ExtensionName)
		if err != nil {
			return nil, err
		}
		ident := extensions.CreateLibraryIdentifier(c.ExtensionName, ext.Control.DefaultVersion)
		_, err = extensions.GetExtensionFunction(extensions.CreateLibraryIdentifier(c.ExtensionName, ext.Control.DefaultVersion), c.ExtensionSymbol)
		if err != nil {
			return nil, err
		}
		extName = string(ident)
	}
	err = funcCollection.AddFunction(ctx, functions.Function{
		ID:                 funcID,
		ReturnType:         c.ReturnType.ID,
		ParameterNames:     paramNames,
		ParameterTypes:     paramTypes,
		ParameterDefaults:  paramDefaults,
		Variadic:           false, // TODO: implement this
		IsNonDeterministic: true,
		Strict:             c.Strict,
		Definition:         c.Definition,
		ExtensionName:      extName,
		ExtensionSymbol:    c.ExtensionSymbol,
		Operations:         c.Statements,
		SQLDefinition:      c.SqlDef,
		SetOf:              c.SetOf,
	})
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateFunction) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateFunction) String() string {
	// TODO: fully implement this
	return "CREATE FUNCTION"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateFunction) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *CreateFunction) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) > len(c.Parameters) {
		// the number of default values can be fewer but cannot be more.
		return nil, ErrVitessChildCount.New(len(c.Parameters), len(children))
	}
	newParams := make([]RoutineParam, len(c.Parameters))
	childIdx := 0
	for i, param := range c.Parameters {
		newParams[i] = param
		if c.Parameters[i].HasDefault && childIdx < len(children) {
			expr, ok := children[childIdx].(sql.Expression)
			if !ok {
				return nil, errors.Errorf("invalid vitess child, expected sql.Expression for Default value but got %t", children[i])
			}
			newParams[i].Default = expr
			childIdx++
		}
	}
	ncf := *c
	ncf.Parameters = newParams
	return &ncf, nil
}

// FunctionColumn represents the deferred column used in functions.
// It is a placeholder column reference later used for function calls.
type FunctionColumn struct {
	Name string
	Typ  *pgtypes.DoltgresType
	Idx  uint16
}

var _ vitess.Injectable = (*FunctionColumn)(nil)
var _ sql.Expression = (*FunctionColumn)(nil)

// Resolved implements the interface sql.Expression.
func (f *FunctionColumn) Resolved() bool {
	return !f.Typ.IsEmptyType()
}

// String implements the interface sql.Expression.
func (f *FunctionColumn) String() string {
	if f.Name != "" {
		return fmt.Sprintf(`$%v`, f.Idx+1)
	}
	return f.Name
}

// Type implements the interface sql.Expression.
func (f *FunctionColumn) Type(ctx *sql.Context) sql.Type {
	if f.Typ.IsEmptyType() {
		return pgtypes.Unknown
	}
	return f.Typ
}

// IsNullable implements the interface sql.Expression.
func (f *FunctionColumn) IsNullable(ctx *sql.Context) bool {
	return false
}

// Eval implements the interface sql.Expression.
func (f *FunctionColumn) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	panic("FunctionColumn is a placeholder expression, but Eval() was called")
}

// Children implements the interface sql.Expression.
func (f *FunctionColumn) Children() []sql.Expression {
	return nil
}

// WithChildren implements the interface sql.Expression.
func (f *FunctionColumn) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(f, len(children), 0)
	}
	return f, nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (f *FunctionColumn) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, errors.Errorf("invalid FunctionColumn child count, expected `0` but got `%d`", len(children))
	}
	return f, nil
}
