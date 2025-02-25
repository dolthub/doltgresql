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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	parsertypes "github.com/dolthub/doltgresql/postgres/parser/types"
	"github.com/dolthub/doltgresql/server/types"
)

// DropFunction implements DROP FUNCTION.
type DropFunction struct {
	routinesWithArgs []tree.RoutineWithArgs
	ifExists         bool
	cascade          bool
}

var _ sql.ExecSourceRel = (*DropFunction)(nil)
var _ vitess.Injectable = (*DropFunction)(nil)

// NewDropFunction returns a new *DropFunction.
func NewDropFunction(ifExists bool, routinesWithArgs []tree.RoutineWithArgs, cascade bool) *DropFunction {
	return &DropFunction{
		ifExists:         ifExists,
		routinesWithArgs: routinesWithArgs,
		cascade:          cascade,
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
func (d *DropFunction) Schema() sql.Schema {
	return nil
}

// Children implements the interface sql.ExecSourceRel.
func (d *DropFunction) Children() []sql.Node {
	return nil
}

// WithChildren implements the interface sql.ExecSourceRel.
func (d *DropFunction) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(d, children...)
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (d *DropFunction) IsReadOnly() bool {
	return false
}

// RowIter implements the interface sql.ExecSourceRel.
func (d *DropFunction) RowIter(ctx *sql.Context, r sql.Row) (iter sql.RowIter, err error) {
	for _, routineWithArgs := range d.routinesWithArgs {
		unresolvedObjectName := routineWithArgs.Name
		dbName := unresolvedObjectName.Catalog()
		routineName := unresolvedObjectName.Object()

		if dbName != "" && dbName != ctx.GetCurrentDatabase() {
			return nil, fmt.Errorf("DROP FUNCTION is currently only supported for the current database")
		}

		var function *functions.Function
		if len(routineWithArgs.Args) == 0 {
			function, err = d.findFunctionByName(ctx, routineName)
			if err != nil {
				return nil, err
			}
		} else {
			function, err = d.findFunctionBySignature(ctx, routineWithArgs)
			if err != nil {
				return nil, err
			}
		}

		if function == nil {
			if d.ifExists {
				// TODO: issue a notice
				return sql.RowsToRowIter(), nil
			} else {
				return nil, types.ErrFunctionDoesNotExist.New(formatRoutineName(routineWithArgs))
			}
		}

		// TODO: Check to see if this function is used by anything before dropping

		collection, err := core.GetFunctionsCollectionFromContext(ctx)
		if err != nil {
			return nil, err
		}

		err = collection.DropFunction(function.ID)
		if err != nil {
			return nil, err
		}
	}

	return sql.RowsToRowIter(), nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (d *DropFunction) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return d, nil
}

// findFunctionByName searches through the available functions, looking for one named |routineName|.
// If multiple functions with that name are found, then the function overload with no parameters
// will be returned if it exists. If multiple functions match, but they all have parameters, then
// an error message about the name not being unique will be returned.
func (d *DropFunction) findFunctionByName(ctx *sql.Context, routineName string) (*functions.Function, error) {
	collection, err := core.GetFunctionsCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var matchingFunctions []functions.Function
	err = collection.IterateFunctions(func(function *functions.Function) error {
		if function.ID.FunctionName() == routineName {
			matchingFunctions = append(matchingFunctions, *function)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	switch len(matchingFunctions) {
	case 0:
		return nil, nil
	case 1:
		return &matchingFunctions[0], nil
	default:
		for _, function := range matchingFunctions {
			if len(function.ParameterNames) == 0 {
				return &function, nil
			}
		}
		return nil, fmt.Errorf(`function name "%s" is not unique`, routineName)
	}
}

// findFunctionBySignature takes the specified signature of |routineWithArgs| and forms a function
// ID using the optional catalog and schema name, the routine name, and the specified parameter
// types. If a function matching that signature is found, it will be returned.
func (d *DropFunction) findFunctionBySignature(ctx *sql.Context, routineWithArgs tree.RoutineWithArgs) (*functions.Function, error) {
	collection, err := core.GetFunctionsCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	unresolvedObjectName := routineWithArgs.Name
	routineName := unresolvedObjectName.Object()
	// TODO: User defined functions are currently registered in pg_catalog
	schemaName := "pg_catalog"

	typeIds := make([]id.Type, 0, len(routineWithArgs.Args))
	for _, routineArg := range routineWithArgs.Args {
		switch routineArg.Mode {
		case tree.RoutineArgModeIn:
			// This is the default parameter mode
		case tree.RoutineArgModeOut:
			// Skip any out params, since they are not used to disambiguate function overloads
			continue
		case tree.RoutineArgModeVariadic:
			return nil, fmt.Errorf("DROP FUNCTION does not currently support VARIADIC parameters")
		case tree.RoutineArgModeInout:
			return nil, fmt.Errorf("DROP FUNCTION does not currently support INOUT parameters")
		}

		// TODO: This is becoming a common pattern... should extract a helper function
		var typeName string
		switch typ := routineArg.Type.(type) {
		case *parsertypes.T:
			typeName = strings.ToLower(typ.Name())
		default:
			typeName = strings.ToLower(typ.SQLString())
		}

		typeId := id.NewType(schemaName, typeName)

		typeCollection, err := core.GetTypesCollectionFromContext(ctx)
		if err != nil {
			return nil, err
		}
		getType, found := typeCollection.GetType(typeId)
		if !found {
			return nil, types.ErrTypeDoesNotExist.New(typeName)
		}
		typeIds = append(typeIds, getType.ID)
	}

	schema, err := core.GetSchemaName(ctx, nil, schemaName)
	if err != nil {
		return nil, err
	}

	functionId := id.NewFunction(schema, routineName, typeIds...)
	return collection.GetFunction(functionId), nil
}

// formatRoutineName takes the specified |routineWithArgs| and returns a string representing
// it, including the catalog and schema name if they are specified, as well as any type
// information if it is specified.
func formatRoutineName(routineWithArgs tree.RoutineWithArgs) (s string) {
	if routineWithArgs.Name.Catalog() != "" {
		s += routineWithArgs.Name.Catalog() + "."
	}
	if routineWithArgs.Name.Schema() != "" {
		s += routineWithArgs.Name.Schema() + "."
	}
	s += routineWithArgs.Name.Object()

	if len(routineWithArgs.Args) > 0 {
		s += "("
		for i, arg := range routineWithArgs.Args {
			if i > 0 {
				s += ", "
			}
			s += arg.Type.SQLString()
		}
		s += ")"
	}

	return s
}
