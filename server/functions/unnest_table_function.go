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

package functions

import (
	"fmt"
	"io"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	dtablefunctions "github.com/dolthub/go-mysql-server/sql/expression/tablefunction"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// UnnestTableFunction is the FROM-clause form of unnest, returning one column per array and padding shorter arrays
// with NULL.
type UnnestTableFunction struct {
	database sql.Database
	arrays   []sql.Expression
}

var _ sql.TableFunction = (*UnnestTableFunction)(nil)
var _ sql.ExecSourceRel = (*UnnestTableFunction)(nil)

// NewInstance implements the interface sql.TableFunction.
func (u *UnnestTableFunction) NewInstance(ctx *sql.Context, database sql.Database, args []sql.Expression) (sql.Node, error) {
	if len(args) == 1 {
		unnest := sql.NewFunctionN(u.Name(), func(ctx *sql.Context, args ...sql.Expression) (sql.Expression, error) {
			compiledFunction, _, err := framework.GetFunction(ctx, u.Name(), args...)
			return compiledFunction, err
		})
		return dtablefunctions.NewTableFunctionWrapper(unnest).NewInstance(ctx, database, args)
	}
	return (&UnnestTableFunction{database: database}).WithExpressions(ctx, args...)
}

// Name implements the interface sql.TableFunction.
func (u *UnnestTableFunction) Name() string {
	return "unnest"
}

// Database implements the interface sql.Databaser.
func (u *UnnestTableFunction) Database() sql.Database {
	return u.database
}

// WithDatabase implements the interface sql.Databaser.
func (u *UnnestTableFunction) WithDatabase(database sql.Database) (sql.Node, error) {
	nu := *u
	nu.database = database
	return &nu, nil
}

// Expressions implements the interface sql.Expressioner.
func (u *UnnestTableFunction) Expressions() []sql.Expression {
	return u.arrays
}

// WithExpressions implements the interface sql.Expressioner.
func (u *UnnestTableFunction) WithExpressions(ctx *sql.Context, exprs ...sql.Expression) (sql.Node, error) {
	for _, expr := range exprs {
		typ, ok := expr.Type(ctx).(*pgtypes.DoltgresType)
		if !ok {
			typ = pgtypes.FromGmsType(expr.Type(ctx))
		}
		if !typ.IsArrayType() {
			return nil, framework.ErrFunctionDoesNotExist.New(fmt.Sprintf("pg_catalog.unnest(%s)", typ))
		}
	}
	nu := *u
	nu.arrays = exprs
	return &nu, nil
}

// Schema implements the interface sql.Node.
func (u *UnnestTableFunction) Schema(ctx *sql.Context) sql.Schema {
	schema := make(sql.Schema, len(u.arrays))
	for i, array := range u.arrays {
		schema[i] = &sql.Column{Name: u.Name(), Type: array.Type(ctx).(*pgtypes.DoltgresType).ArrayBaseType(), Nullable: true}
	}
	return schema
}

// Children implements the interface sql.Node.
func (u *UnnestTableFunction) Children() []sql.Node {
	return nil
}

// WithChildren implements the interface sql.Node.
func (u *UnnestTableFunction) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, errors.Errorf("unexpected children")
	}
	return u, nil
}

// Resolved implements the interface sql.Resolvable.
func (u *UnnestTableFunction) Resolved() bool {
	for _, array := range u.arrays {
		if !array.Resolved() {
			return false
		}
	}
	return true
}

// IsReadOnly implements the interface sql.Node.
func (u *UnnestTableFunction) IsReadOnly() bool {
	return true
}

// String implements the interface fmt.Stringer.
func (u *UnnestTableFunction) String() string {
	arrays := make([]string, len(u.arrays))
	for i, array := range u.arrays {
		arrays[i] = array.String()
	}
	return fmt.Sprintf("unnest(%s)", strings.Join(arrays, ", "))
}

// RowIter implements the interface sql.ExecSourceRel.
func (u *UnnestTableFunction) RowIter(ctx *sql.Context, row sql.Row) (sql.RowIter, error) {
	arrays := make([][]any, len(u.arrays))
	rowCount := 0
	for i, array := range u.arrays {
		val, err := array.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		arrays[i], _ = val.([]any)
		rowCount = max(rowCount, len(arrays[i]))
	}

	var i = 0
	return pgtypes.NewSetReturningFunctionRowIter(func(ctx *sql.Context) (sql.Row, error) {
		defer func() { i++ }()

		if i >= rowCount {
			return nil, io.EOF
		}
		result := make(sql.Row, len(arrays))
		for j, array := range arrays {
			if i < len(array) {
				result[j] = array[i]
			}
		}
		return result, nil
	}), nil
}
