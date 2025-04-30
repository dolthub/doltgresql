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
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/triggers"
	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// CreateTrigger implements CREATE TRIGGER.
type CreateTrigger struct {
	Name       id.Trigger
	Function   id.Function
	Replace    bool
	Timing     triggers.TriggerTiming
	Events     []triggers.TriggerEvent
	ForEachRow bool
	When       []plpgsql.InterpreterOperation
	Arguments  []string
	Definition string
}

var _ sql.ExecSourceRel = (*CreateTrigger)(nil)
var _ vitess.Injectable = (*CreateTrigger)(nil)

// NewCreateTrigger returns a new *CreateTrigger.
func NewCreateTrigger(
	triggerName id.Trigger,
	functionName id.Function,
	replace bool,
	timing triggers.TriggerTiming,
	events []triggers.TriggerEvent,
	forEachRow bool,
	when []plpgsql.InterpreterOperation,
	arguments []string,
	definition string) *CreateTrigger {
	return &CreateTrigger{
		Name:       triggerName,
		Function:   functionName,
		Replace:    replace,
		Timing:     timing,
		Events:     events,
		ForEachRow: forEachRow,
		When:       when,
		Arguments:  arguments,
		Definition: definition,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *CreateTrigger) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CreateTrigger) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CreateTrigger) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CreateTrigger) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	schema, err := core.GetSchemaName(ctx, nil, c.Name.SchemaName())
	if err != nil {
		return nil, err
	}
	triggerID := id.NewTrigger(schema, c.Name.TableName(), c.Name.TriggerName())
	relationType, err := core.GetRelationType(ctx, schema, c.Name.TableName())
	if err != nil {
		return nil, err
	}
	if relationType == core.RelationType_DoesNotExist {
		return nil, errors.Errorf(`relation "%s" does not exist`, c.Name.TableName())
	} else if relationType != core.RelationType_Table {
		return nil, errors.Errorf(`"%s" is not a table or view`, c.Name.TableName())
	}
	function, err := loadFunction(ctx, nil, c.Function)
	if err != nil {
		return nil, err
	}
	if !function.ID.IsValid() {
		return nil, errors.Errorf("function %s() does not exist", c.Function.FunctionName())
	}
	if function.ReturnType != pgtypes.Trigger.ID {
		return nil, errors.Errorf(`function %s must return type trigger`, function.ID.FunctionName())
	}
	trigCollection, err := core.GetTriggersCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if c.Replace && trigCollection.HasTrigger(ctx, triggerID) {
		if err = trigCollection.DropTrigger(ctx, triggerID); err != nil {
			return nil, err
		}
	}
	err = trigCollection.AddTrigger(ctx, triggers.Trigger{
		ID:                  triggerID,
		Function:            function.ID,
		Timing:              c.Timing,
		Events:              c.Events,
		ForEachRow:          c.ForEachRow,
		When:                c.When,
		Deferrable:          triggers.TriggerDeferrable_NotDeferrable,
		ReferencedTableName: "",
		Constraint:          false,
		OldTransitionName:   "",
		NewTransitionName:   "",
		Arguments:           c.Arguments,
		Definition:          c.Definition,
	})
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateTrigger) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateTrigger) String() string {
	return "CREATE TRIGGER"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateTrigger) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *CreateTrigger) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}

// loadFunction loads the function with the given ID from the given collection. If the collection is nil, then this also
// loads the collection from the context.
func loadFunction(ctx *sql.Context, funcCollection *functions.Collection, funcID id.Function) (functions.Function, error) {
	var err error
	if funcCollection == nil {
		funcCollection, err = core.GetFunctionsCollectionFromContext(ctx)
		if err != nil {
			return functions.Function{}, err
		}
	}
	var function functions.Function
	if len(funcID.SchemaName()) > 0 {
		function, err = funcCollection.GetFunction(ctx, funcID)
		if err != nil {
			return functions.Function{}, err
		}
	} else {
		searchPaths, err := resolve.SearchPath(ctx)
		if err != nil {
			return functions.Function{}, err
		}
		searchPaths = append(searchPaths, "pg_catalog") // This isn't included in the search path but functions use it
		for _, searchPath := range searchPaths {
			function, err = funcCollection.GetFunction(ctx, id.NewFunction(searchPath, funcID.FunctionName()))
			if err != nil {
				return functions.Function{}, err
			}
			if function.ID.IsValid() {
				break
			}
		}
	}
	return function, nil
}
