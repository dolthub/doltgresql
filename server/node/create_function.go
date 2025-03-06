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
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/functions"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/interpreter"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// CreateFunction implements CREATE FUNCTION.
type CreateFunction struct {
	FunctionName   string
	SchemaName     string
	Replace        bool
	ReturnType     *pgtypes.DoltgresType
	ParameterNames []string
	ParameterTypes []*pgtypes.DoltgresType
	Strict         bool
	Statements     []interpreter.InterpreterOperation
}

var _ sql.ExecSourceRel = (*CreateFunction)(nil)
var _ vitess.Injectable = (*CreateFunction)(nil)

// NewCreateFunction returns a new *CreateFunction.
func NewCreateFunction(
	functionName string,
	schemaName string,
	replace bool,
	retType *pgtypes.DoltgresType,
	paramNames []string,
	paramTypes []*pgtypes.DoltgresType,
	strict bool,
	statements []interpreter.InterpreterOperation) *CreateFunction {
	return &CreateFunction{
		FunctionName:   functionName,
		SchemaName:     schemaName,
		Replace:        replace,
		ReturnType:     retType,
		ParameterNames: paramNames,
		ParameterTypes: paramTypes,
		Strict:         strict,
		Statements:     statements,
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
	idTypes := make([]id.Type, len(c.ParameterTypes))
	for i, typ := range c.ParameterTypes {
		idTypes[i] = typ.ID
	}
	funcCollection, err := core.GetFunctionsCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	paramTypes := make([]id.Type, len(c.ParameterTypes))
	for i, paramType := range c.ParameterTypes {
		paramTypes[i] = paramType.ID
	}
	funcID := id.NewFunction(c.SchemaName, c.FunctionName, idTypes...)
	if c.Replace && funcCollection.HasFunction(funcID) {
		if err = funcCollection.DropFunction(funcID); err != nil {
			return nil, err
		}
	}
	err = funcCollection.AddFunction(&functions.Function{
		ID:                 funcID,
		ReturnType:         c.ReturnType.ID,
		ParameterNames:     c.ParameterNames,
		ParameterTypes:     paramTypes,
		Variadic:           false, // TODO: implement this
		IsNonDeterministic: true,
		Strict:             c.Strict,
		Operations:         c.Statements,
	})
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateFunction) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateFunction) String() string {
	// TODO: fully implement this
	return "CREATE FUNCTION"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateFunction) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *CreateFunction) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
