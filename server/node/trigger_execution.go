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
	"github.com/dolthub/go-mysql-server/sql/rowexec"

	"github.com/dolthub/doltgresql/core/triggers"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TriggerExecutionRowHandling states how to interpret the source row, or how to return the resulting row.
type TriggerExecutionRowHandling uint8

const (
	TriggerExecutionRowHandling_None TriggerExecutionRowHandling = iota
	TriggerExecutionRowHandling_Old
	TriggerExecutionRowHandling_OldNew
	TriggerExecutionRowHandling_NewOld
	TriggerExecutionRowHandling_New
)

// TriggerExecution handles the execution of a set of triggers on a table.
type TriggerExecution struct {
	Triggers []triggers.Trigger
	Split    TriggerExecutionRowHandling // How the source row should be split
	Return   TriggerExecutionRowHandling // How the returned rows should be combined
	Sch      sql.Schema
	Source   sql.Node
	Runner   pgexprs.StatementRunner
}

var _ sql.ExecSourceRel = (*TriggerExecution)(nil)
var _ sql.Expressioner = (*TriggerExecution)(nil)

func (te *TriggerExecution) Children() []sql.Node {
	return []sql.Node{te.Source}
}

// Expressions implements the interface sql.Expressioner.
func (te *TriggerExecution) Expressions() []sql.Expression {
	return []sql.Expression{te.Runner}
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (te *TriggerExecution) IsReadOnly() bool {
	return te.Source.IsReadOnly()
}

// Resolved implements the interface sql.ExecSourceRel.
func (te *TriggerExecution) Resolved() bool {
	return te.Source.Resolved()
}

// RowIter implements the interface sql.ExecSourceRel.
func (te *TriggerExecution) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	sourceIter, err := rowexec.DefaultBuilder.Build(ctx, te.Source, r)
	if err != nil {
		return nil, err
	}
	// If there are no triggers, then we'll just return the source iter
	if len(te.Triggers) == 0 {
		return sourceIter, nil
	}
	trigFuncs := make([]framework.InterpretedFunction, len(te.Triggers))
	whens := make([]framework.InterpretedFunction, len(te.Triggers))
	for i, trig := range te.Triggers {
		trigFuncs[i], err = te.loadTriggerFunction(ctx, trig)
		if err != nil {
			return nil, err
		}
		// If we have a WHEN expression, then we need to build a "function" to execute the expression
		if len(trig.When) > 0 {
			whens[i] = framework.InterpretedFunction{
				ID:         trigFuncs[i].ID, // Assign the same ID just so we have a valid one for later
				ReturnType: pgtypes.Bool,
				Statements: trig.When,
			}
		}
	}
	return &triggerExecutionIter{
		functions: trigFuncs,
		whens:     whens,
		split:     te.Split,
		treturn:   te.Return,
		runner:    te.Runner.Runner,
		sch:       te.Sch,
		source:    sourceIter,
	}, nil
}

// Schema implements the interface sql.ExecSourceRel.
func (te *TriggerExecution) Schema() sql.Schema {
	return te.Source.Schema()
}

// String implements the interface sql.ExecSourceRel.
func (te *TriggerExecution) String() string {
	return "TRIGGER EXECUTION"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (te *TriggerExecution) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(te, len(children), 1)
	}
	newTe := *te
	newTe.Source = children[0]
	return &newTe, nil
}

// WithExpressions implements the interface sql.Expressioner.
func (te *TriggerExecution) WithExpressions(expressions ...sql.Expression) (sql.Node, error) {
	if len(expressions) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(te, len(expressions), 1)
	}
	newTe := *te
	newTe.Runner = expressions[0].(pgexprs.StatementRunner)
	return &newTe, nil
}

// loadTriggerFunction loads the given trigger's framework.InterpretedFunction.
func (te *TriggerExecution) loadTriggerFunction(ctx *sql.Context, trigger triggers.Trigger) (framework.InterpretedFunction, error) {
	function, err := loadFunction(ctx, nil, trigger.Function)
	if err != nil {
		return framework.InterpretedFunction{}, err
	}
	if !function.ID.IsValid() {
		return framework.InterpretedFunction{}, errors.Errorf("function %s() does not exist", trigger.Function.FunctionName())
	}
	if function.ReturnType != pgtypes.Trigger.ID {
		return framework.InterpretedFunction{}, errors.Errorf(`function %s must return type trigger`, function.ID.FunctionName())
	}
	return framework.InterpretedFunction{
		ID:                 function.ID,
		ReturnType:         pgtypes.Trigger,
		ParameterNames:     nil,
		ParameterTypes:     nil,
		Variadic:           function.Variadic,
		IsNonDeterministic: function.IsNonDeterministic,
		Strict:             function.Strict,
		Statements:         function.Operations,
	}, nil
}

// triggerExecutionIter is the iterator for TriggerExecution.
type triggerExecutionIter struct {
	functions []framework.InterpretedFunction
	whens     []framework.InterpretedFunction
	split     TriggerExecutionRowHandling
	treturn   TriggerExecutionRowHandling
	runner    sql.StatementRunner
	sch       sql.Schema
	source    sql.RowIter
}

var _ sql.RowIter = (*triggerExecutionIter)(nil)

// Next implements the interface sql.RowIter.
func (t *triggerExecutionIter) Next(ctx *sql.Context) (sql.Row, error) {
	nextRow, err := t.source.Next(ctx)
	if err != nil {
		return nextRow, err
	}
	var oldRow sql.Row
	var newRow sql.Row
	switch t.split {
	case TriggerExecutionRowHandling_Old:
		oldRow = nextRow
	case TriggerExecutionRowHandling_OldNew:
		oldRow = nextRow[:len(t.sch)]
		newRow = nextRow[len(t.sch):]
	case TriggerExecutionRowHandling_NewOld:
		newRow = nextRow[:len(t.sch)]
		oldRow = nextRow[len(t.sch):]
	case TriggerExecutionRowHandling_New:
		newRow = nextRow
	}
	for funcIdx, function := range t.functions {
		if t.whens[funcIdx].ID.IsValid() {
			whenValue, err := plpgsql.TriggerCall(ctx, t.whens[funcIdx], t.runner, t.sch, oldRow, newRow)
			if err != nil {
				if strings.Contains(err.Error(), "no valid cast for return value") {
					// TODO: this error should technically be caught during parsing, but interpreted functions don't
					//  have the ability to determine types during parsing yet (also applies to the same error below)
					return nil, fmt.Errorf("argument of WHEN must be type boolean")
				}
				return nil, err
			}
			whenBool, ok := whenValue.(bool)
			if !ok {
				return nil, fmt.Errorf("argument of WHEN must be type boolean")
			}
			if !whenBool {
				continue
			}
		}
		returnedValue, err := plpgsql.TriggerCall(ctx, function, t.runner, t.sch, oldRow, newRow)
		if err != nil {
			return nil, err
		}
		if returnedValue == nil {
			return make(sql.Row, len(nextRow)), nil
		}
		var ok bool
		returnedRow, ok := returnedValue.(sql.Row)
		if !ok {
			return nil, fmt.Errorf("invalid trigger return value")
		}
		switch t.split {
		case TriggerExecutionRowHandling_Old:
			oldRow = returnedRow
		case TriggerExecutionRowHandling_OldNew, TriggerExecutionRowHandling_NewOld, TriggerExecutionRowHandling_New:
			newRow = returnedRow
		}
	}
	switch t.treturn {
	case TriggerExecutionRowHandling_Old:
		return oldRow, nil
	case TriggerExecutionRowHandling_OldNew:
		retRow := make(sql.Row, len(nextRow))
		copy(retRow, oldRow)
		copy(retRow[len(oldRow):], newRow)
		return retRow, nil
	case TriggerExecutionRowHandling_NewOld:
		retRow := make(sql.Row, len(nextRow))
		copy(retRow, newRow)
		copy(retRow[len(newRow):], oldRow)
		return retRow, nil
	case TriggerExecutionRowHandling_New:
		return newRow, nil
	default:
		return nextRow, nil
	}
}

// Close implements the interface sql.RowIter.
func (t *triggerExecutionIter) Close(ctx *sql.Context) error {
	return t.source.Close(ctx)
}
