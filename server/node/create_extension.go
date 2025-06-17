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
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/extensions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
)

// CreateExtension implements CREATE EXTENSION.
type CreateExtension struct {
	Name        string
	IfNotExists bool
	SchemaName  string
	Version     string
	Cascade     bool
	Runner      pgexprs.StatementRunner
}

var _ sql.ExecSourceRel = (*CreateExtension)(nil)
var _ sql.Expressioner = (*CreateExtension)(nil)
var _ vitess.Injectable = (*CreateExtension)(nil)

// NewCreateExtension returns a new *CreateExtension.
func NewCreateExtension(name string, ifNotExists bool, schemaName string, version string, cascade bool) *CreateExtension {
	return &CreateExtension{
		Name:        name,
		IfNotExists: ifNotExists,
		SchemaName:  schemaName,
		Version:     version,
		Cascade:     cascade,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *CreateExtension) Children() []sql.Node {
	return nil
}

// Expressions implements the interface sql.Expressioner.
func (c *CreateExtension) Expressions() []sql.Expression {
	return []sql.Expression{c.Runner}
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CreateExtension) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CreateExtension) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CreateExtension) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	extCollection, err := core.GetExtensionsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	if extCollection.HasLoadedExtension(ctx, id.NewExtension(c.Name)) {
		if c.IfNotExists {
			return sql.RowsToRowIter(), nil
		}
		return nil, errors.Errorf(`extension "%s" already exists`, c.Name)
	}
	ext, err := extensions.GetExtension(c.Name)
	if err != nil {
		return nil, err
	}
	// The returned files are in their proper order of execution, so we can iterate and execute
	sqlFiles, err := ext.LoadSQLFiles()
	if err != nil {
		return nil, err
	}
	for _, sqlFile := range sqlFiles {
		// Remove echo PSQL control statements
		for {
			echoStartIdx := strings.Index(sqlFile, `\echo`)
			if echoStartIdx == -1 {
				break
			}
			echoEndIdx := strings.Index(sqlFile[echoStartIdx:], "\n")
			if echoEndIdx != -1 {
				// Set the correct absolute position if there is a newline
				echoEndIdx += echoStartIdx
			} else {
				// Set the position at the end of the file if there's no newline (comment appears before EOF)
				echoEndIdx = len(sqlFile)
			}
			sqlFile = strings.Replace(sqlFile, sqlFile[echoStartIdx:echoEndIdx], "", 1)
		}
		statements, err := parser.Parse(sqlFile)
		if err != nil {
			return nil, err
		}
		for _, statement := range statements {
			statementSQL := statement.SQL
			if _, ok := statement.AST.(*tree.CreateFunction); ok {
				statementSQL = strings.ReplaceAll(statementSQL, `'MODULE_PATHNAME'`, fmt.Sprintf(`'%s'`, c.Name))
			}
			_, err = sql.RunInterpreted(ctx, func(subCtx *sql.Context) ([]sql.Row, error) {
				_, rowIter, _, err := c.Runner.Runner.QueryWithBindings(subCtx, statementSQL, nil, nil, nil)
				if err != nil {
					return nil, err
				}
				return sql.RowIterToRows(subCtx, rowIter)
			})
			if err != nil {
				return nil, err
			}
		}
	}
	namespace := id.NullNamespace
	if len(ext.Control.Schema) > 0 {
		namespace = id.NewNamespace(ext.Control.Schema)
	}
	err = extCollection.AddLoadedExtension(ctx, extensions.Extension{
		ExtName:       id.NewExtension(c.Name),
		Namespace:     namespace,
		Relocatable:   ext.Control.Relocatable,
		LibIdentifier: extensions.CreateLibraryIdentifier(c.Name, ext.Control.DefaultVersion),
	})
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateExtension) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateExtension) String() string {
	return "CREATE EXTENSION"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateExtension) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithExpressions implements the interface sql.Expressioner.
func (c *CreateExtension) WithExpressions(expressions ...sql.Expression) (sql.Node, error) {
	if len(expressions) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(c, len(expressions), 1)
	}
	newC := *c
	newC.Runner = expressions[0].(pgexprs.StatementRunner)
	return &newC, nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *CreateExtension) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
