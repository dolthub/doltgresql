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

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/procedures"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	parsertypes "github.com/dolthub/doltgresql/postgres/parser/types"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/types"
)

// DropProcedure implements DROP PROCEDURE.
type DropProcedure struct {
	routinesWithArgs []tree.RoutineWithArgs
	ifExists         bool
	cascade          bool
}

var _ sql.ExecSourceRel = (*DropProcedure)(nil)
var _ vitess.Injectable = (*DropProcedure)(nil)

// NewDropProcedure returns a new *DropProcedure.
func NewDropProcedure(ifExists bool, routinesWithArgs []tree.RoutineWithArgs, cascade bool) *DropProcedure {
	return &DropProcedure{
		ifExists:         ifExists,
		routinesWithArgs: routinesWithArgs,
		cascade:          cascade,
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
func (d *DropProcedure) Schema() sql.Schema {
	return nil
}

// Children implements the interface sql.ExecSourceRel.
func (d *DropProcedure) Children() []sql.Node {
	return nil
}

// WithChildren implements the interface sql.ExecSourceRel.
func (d *DropProcedure) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(d, children...)
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (d *DropProcedure) IsReadOnly() bool {
	return false
}

// RowIter implements the interface sql.ExecSourceRel.
func (d *DropProcedure) RowIter(ctx *sql.Context, r sql.Row) (iter sql.RowIter, err error) {
	for _, routineWithArgs := range d.routinesWithArgs {
		unresolvedObjectName := routineWithArgs.Name
		dbName := unresolvedObjectName.Catalog()
		routineName := unresolvedObjectName.Object()

		if dbName != "" && dbName != ctx.GetCurrentDatabase() {
			return nil, fmt.Errorf("DROP PROCEDURE is currently only supported for the current database")
		}

		var procedure procedures.Procedure
		if len(routineWithArgs.Args) == 0 {
			procedure, err = d.findProcedureByName(ctx, routineName)
			if err != nil {
				return nil, err
			}
		} else {
			procedure, err = d.findProcedureBySignature(ctx, routineWithArgs)
			if err != nil {
				return nil, err
			}
		}

		if !procedure.ID.IsValid() {
			if d.ifExists {
				noticeResponse := &pgproto3.NoticeResponse{
					Severity: "WARNING",
					Message:  fmt.Sprintf("function %s() does not exist, skipping", routineName),
				}
				sess := dsess.DSessFromSess(ctx.Session)
				sess.Notice(noticeResponse)
				return sql.RowsToRowIter(), nil
			} else {
				return nil, framework.ErrFunctionDoesNotExist.New(formatRoutineName(routineWithArgs))
			}
		}

		// TODO: Check to see if this procedure is used by anything before dropping
		collection, err := core.GetProceduresCollectionFromContext(ctx)
		if err != nil {
			return nil, err
		}

		err = collection.DropProcedure(ctx, procedure.ID)
		if err != nil {
			return nil, err
		}
	}

	return sql.RowsToRowIter(), nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (d *DropProcedure) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return d, nil
}

// findProcedureByName searches through the available procedures, looking for one matching `routineName`. If multiple
// procedures with that name are found, then the procedure overload with no parameters will be returned if it exists. If
// multiple procedures match, but they all have parameters, then an error message about the name not being unique will
// be returned.
func (d *DropProcedure) findProcedureByName(ctx *sql.Context, routineName string) (procedures.Procedure, error) {
	collection, err := core.GetProceduresCollectionFromContext(ctx)
	if err != nil {
		return procedures.Procedure{}, err
	}

	var matchingProcedures []procedures.Procedure
	err = collection.IterateProcedures(ctx, func(proc procedures.Procedure) (bool, error) {
		if proc.ID.ProcedureName() == routineName {
			matchingProcedures = append(matchingProcedures, proc)
		}
		return false, nil
	})
	if err != nil {
		return procedures.Procedure{}, err
	}

	switch len(matchingProcedures) {
	case 0:
		return procedures.Procedure{}, nil
	case 1:
		return matchingProcedures[0], nil
	default:
		for _, proc := range matchingProcedures {
			if len(proc.ParameterNames) == 0 {
				return proc, nil
			}
		}
		return procedures.Procedure{}, fmt.Errorf(`function name "%s" is not unique`, routineName)
	}
}

// findProcedureBySignature takes the specified signature of `routineWithArgs` finds a matching procedure. If a
// procedure matching that signature is found, it will be returned.
func (d *DropProcedure) findProcedureBySignature(ctx *sql.Context, routineWithArgs tree.RoutineWithArgs) (procedures.Procedure, error) {
	collection, err := core.GetProceduresCollectionFromContext(ctx)
	if err != nil {
		return procedures.Procedure{}, err
	}

	unresolvedObjectName := routineWithArgs.Name
	routineName := unresolvedObjectName.Object()
	// TODO: procedures need to use the search path for matches
	schemaName, err := core.GetSchemaName(ctx, nil, "")
	if err != nil {
		return procedures.Procedure{}, err
	}

	typeIds := make([]id.Type, 0, len(routineWithArgs.Args))
	for _, routineArg := range routineWithArgs.Args {
		switch routineArg.Mode {
		case tree.RoutineArgModeIn:
			// This is the default parameter mode
		case tree.RoutineArgModeOut:
			// Skip any out params, since they are not used to disambiguate procedure overloads
			continue
		case tree.RoutineArgModeVariadic:
			return procedures.Procedure{}, fmt.Errorf("DROP PROCEDURE does not currently support VARIADIC parameters")
		case tree.RoutineArgModeInout:
			return procedures.Procedure{}, fmt.Errorf("DROP PROCEDURE does not currently support INOUT parameters")
		}

		var typeName string
		switch typ := routineArg.Type.(type) {
		case *parsertypes.T:
			typeName = strings.ToLower(typ.Name())
		default:
			typeName = strings.ToLower(typ.SQLString())
		}

		// TODO: we need to add a way to search for a matching type along the search path, rather than hardcoding the current schema
		typeId := id.NewType("pg_catalog", typeName)
		typeCollection, err := core.GetTypesCollectionFromContext(ctx)
		if err != nil {
			return procedures.Procedure{}, err
		}
		getType, err := typeCollection.GetType(ctx, typeId)
		if err != nil {
			return procedures.Procedure{}, err
		}
		if getType == nil {
			// TODO: we're doing a second check on the current schema, but this should use the search path instead
			typeId = id.NewType(schemaName, typeName)
			getType, err = typeCollection.GetType(ctx, typeId)
			if err != nil {
				return procedures.Procedure{}, err
			}
			if getType == nil {
				return procedures.Procedure{}, types.ErrTypeDoesNotExist.New(typeName)
			}
		}
		typeIds = append(typeIds, getType.ID)
	}

	procId := id.NewProcedure(schemaName, routineName, typeIds...)
	return collection.GetProcedure(ctx, procId)
}
