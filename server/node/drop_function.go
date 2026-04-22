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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
)

// RoutineWithParams represent a function or a procedure with schema name, routine name and its parameters.
type RoutineWithParams struct {
	SchemaName  string
	RoutineName string
	Args        []RoutineParam
}

// DropFunction implements DROP FUNCTION.
type DropFunction struct {
	RoutinesWithArgs []*RoutineWithParams
	IfExists         bool
	Cascade          bool
}

var _ sql.ExecSourceRel = (*DropFunction)(nil)
var _ vitess.Injectable = (*DropFunction)(nil)

// NewDropFunction returns a new *DropFunction.
func NewDropFunction(ifExists bool, routinesWithArgs []*RoutineWithParams, cascade bool) *DropFunction {
	return &DropFunction{
		IfExists:         ifExists,
		RoutinesWithArgs: routinesWithArgs,
		Cascade:          cascade,
	}
}

// Resolved implements the interface sql.ExecSourceRel.
func (d *DropFunction) Resolved() bool {
	return true
}

// String implements the interface sql.ExecSourceRel.
func (d *DropFunction) String() string {
	return "DROP FUNCTION"
}

// Schema implements the interface sql.ExecSourceRel.
func (d *DropFunction) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// Children implements the interface sql.ExecSourceRel.
func (d *DropFunction) Children() []sql.Node {
	return nil
}

// WithChildren implements the interface sql.ExecSourceRel.
func (d *DropFunction) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(d, children...)
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (d *DropFunction) IsReadOnly() bool {
	return false
}

// RowIter implements the interface sql.ExecSourceRel.
func (d *DropFunction) RowIter(ctx *sql.Context, r sql.Row) (iter sql.RowIter, err error) {
	funcColl, err := core.GetFunctionsCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	for _, routineWithArgs := range d.RoutinesWithArgs {
		err = dropFunction(ctx, funcColl, routineWithArgs, d.IfExists)
		if err != nil {
			return nil, err
		}
	}

	return sql.RowsToRowIter(), nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (d *DropFunction) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return d, nil
}

func dropFunction(ctx *sql.Context, funcColl *functions.Collection, fn *RoutineWithParams, ifExists bool) error {
	// TODO: provide db
	schema, err := core.GetSchemaName(ctx, nil, fn.SchemaName)
	if err != nil {
		return err
	}
	var funcId = id.NewFunction(schema, fn.RoutineName)
	if len(fn.Args) == 0 {
		funcs, err := funcColl.GetFunctionOverloads(ctx, funcId)
		if err != nil {
			return err
		}
		if len(funcs) == 1 {
			funcId = funcs[0].ID
		} else if len(funcs) > 1 {
			funcExists := funcColl.HasFunction(ctx, funcId)
			if !funcExists {
				return errors.Errorf(`function name "%s" is not unique`, fn.RoutineName)
			}
		}
	} else {
		var argTypes = make([]id.Type, len(fn.Args))
		for i, arg := range fn.Args {
			argTypes[i] = arg.Type.ID
		}
		funcId = id.NewFunction(schema, fn.RoutineName, argTypes...)
	}
	funcExists := funcColl.HasFunction(ctx, funcId)
	if !funcExists && ifExists {
		return nil
	}
	return funcColl.DropFunction(ctx, funcId)
}
