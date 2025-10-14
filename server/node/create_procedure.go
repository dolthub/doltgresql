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

	"github.com/dolthub/doltgresql/core/procedures"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/extensions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// CreateProcedure implements CREATE PROCEDURE.
type CreateProcedure struct {
	ProcedureName   string
	SchemaName      string
	Replace         bool
	ParameterNames  []string
	ParameterTypes  []*pgtypes.DoltgresType
	ParameterModes  []procedures.ParameterMode
	Statements      []plpgsql.InterpreterOperation
	ExtensionName   string
	ExtensionSymbol string
	Definition      string
	SqlDef          string
	SqlDefParsed    vitess.Statement
}

var _ sql.ExecSourceRel = (*CreateProcedure)(nil)
var _ vitess.Injectable = (*CreateProcedure)(nil)

// NewCreateProcedure returns a new *CreateProcedure.
func NewCreateProcedure(
	procedureName string,
	schemaName string,
	replace bool,
	paramNames []string,
	paramTypes []*pgtypes.DoltgresType,
	paramModes []procedures.ParameterMode,
	definition string,
	extensionName string,
	extensionSymbol string,
	statements []plpgsql.InterpreterOperation,
	sqlDef string,
	sqlDefParsed vitess.Statement) *CreateProcedure {
	return &CreateProcedure{
		ProcedureName:   procedureName,
		SchemaName:      schemaName,
		Replace:         replace,
		ParameterNames:  paramNames,
		ParameterTypes:  paramTypes,
		ParameterModes:  paramModes,
		Statements:      statements,
		ExtensionName:   extensionName,
		ExtensionSymbol: extensionSymbol,
		Definition:      definition,
		SqlDef:          sqlDef,
		SqlDefParsed:    sqlDefParsed,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *CreateProcedure) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CreateProcedure) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CreateProcedure) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CreateProcedure) RowIter(ctx *sql.Context, _ sql.Row) (sql.RowIter, error) {
	idTypes := make([]id.Type, len(c.ParameterTypes))
	for i, typ := range c.ParameterTypes {
		idTypes[i] = typ.ID
	}
	procCollection, err := core.GetProceduresCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	paramTypes := make([]id.Type, len(c.ParameterTypes))
	for i, paramType := range c.ParameterTypes {
		paramTypes[i] = paramType.ID
	}
	schemaName, err := core.GetSchemaName(ctx, nil, c.SchemaName)
	if err != nil {
		return nil, err
	}
	procID := id.NewProcedure(schemaName, c.ProcedureName, idTypes...)
	if c.Replace && procCollection.HasProcedure(ctx, procID) {
		if err = procCollection.DropProcedure(ctx, procID); err != nil {
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
	err = procCollection.AddProcedure(ctx, procedures.Procedure{
		ID:              procID,
		ParameterNames:  c.ParameterNames,
		ParameterTypes:  paramTypes,
		ParameterModes:  c.ParameterModes,
		Definition:      c.Definition,
		ExtensionName:   extName,
		ExtensionSymbol: c.ExtensionSymbol,
		Operations:      c.Statements,
		SQLDefinition:   c.SqlDef,
	})
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateProcedure) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateProcedure) String() string {
	return "CREATE PROCEDURE"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateProcedure) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *CreateProcedure) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
