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
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/procedures"
)

// DropProcedure implements DROP PROCEDURE.
type DropProcedure struct {
	RoutinesWithArgs []*RoutineWithParams
	IfExists         bool
	Cascade          bool
}

var _ sql.ExecSourceRel = (*DropProcedure)(nil)
var _ vitess.Injectable = (*DropProcedure)(nil)

// NewDropProcedure returns a new *DropProcedure.
func NewDropProcedure(ifExists bool, routinesWithArgs []*RoutineWithParams, cascade bool) *DropProcedure {
	return &DropProcedure{
		IfExists:         ifExists,
		RoutinesWithArgs: routinesWithArgs,
		Cascade:          cascade,
	}
}

// Resolved implements the interface sql.ExecSourceRel.
func (d *DropProcedure) Resolved() bool {
	return true
}

// String implements the interface sql.ExecSourceRel.
func (d *DropProcedure) String() string {
	return "DROP PROCEDURE"
}

// Schema implements the interface sql.ExecSourceRel.
func (d *DropProcedure) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// Children implements the interface sql.ExecSourceRel.
func (d *DropProcedure) Children() []sql.Node {
	return nil
}

// WithChildren implements the interface sql.ExecSourceRel.
func (d *DropProcedure) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(d, children...)
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (d *DropProcedure) IsReadOnly() bool {
	return false
}

// RowIter implements the interface sql.ExecSourceRel.
func (d *DropProcedure) RowIter(ctx *sql.Context, r sql.Row) (iter sql.RowIter, err error) {
	procColl, err := core.GetProceduresCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	for _, routineWithArgs := range d.RoutinesWithArgs {
		err = dropProcedure(ctx, procColl, routineWithArgs, d.IfExists)
		if err != nil {
			return nil, err
		}
	}

	return sql.RowsToRowIter(), nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (d *DropProcedure) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return d, nil
}

func dropProcedure(ctx *sql.Context, procColl *procedures.Collection, fn *RoutineWithParams, ifExists bool) error {
	// TODO: provide db
	schema, err := core.GetSchemaName(ctx, nil, fn.SchemaName)
	if err != nil {
		return err
	}

	var procId = id.NewProcedure(schema, fn.RoutineName)
	if len(fn.Args) == 0 {
		procs, err := procColl.GetProcedureOverloads(ctx, procId)
		if err != nil {
			return err
		}
		if len(procs) == 1 {
			procId = procs[0].ID
		} else if len(procs) > 1 {
			procExists := procColl.HasProcedure(ctx, procId)
			if !procExists {
				return errors.Errorf(`procedure name "%s" is not unique`, fn.RoutineName)
			}
		}
	} else {
		var argTypes = make([]id.Type, len(fn.Args))
		for i, arg := range fn.Args {
			argTypes[i] = arg.Type.ID
		}
		procId = id.NewProcedure(schema, fn.RoutineName, argTypes...)
	}
	procExists := procColl.HasProcedure(ctx, procId)
	if !procExists && ifExists {
		return nil
	}
	return procColl.DropProcedure(ctx, procId)
}
