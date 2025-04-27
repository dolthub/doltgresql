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

package plpgsql

import (
	"fmt"
	"strings"

	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
	"github.com/dolthub/go-mysql-server/sql"
)

// interpreterVariable is a variable that lives on the stack. This will hold an actual value, but will not be directly
// interacted with. InterpreterVariableReference are, instead, the avenue of interaction as a variable may be an
// aggregate type (such as a record).
type interpreterVariable struct {
	Record sql.Schema
	Type   *pgtypes.DoltgresType
	Value  any
}

// InterpreterVariableReference is a reference to a variable that lives on the stack. If the type is not null, then it
// is valid to dereference the value for assignment. We make use of references rather than directly interacting with
// the variables as this allows for interacting with sections of aggregate types (such as record) as well as normal
// variable interaction.
type InterpreterVariableReference struct {
	Type  *pgtypes.DoltgresType
	Value *any
}

// InterpreterScopeDetails contains all of the details that are relevant to a particular scope.
type InterpreterScopeDetails struct {
	variables map[string]*interpreterVariable
	label     string
}

// InterpreterStack represents the working information that an interpreter will use during execution. It is not exactly
// the same as a stack in the traditional programming sense, but rather is a loose abstraction that serves the same
// general purpose.
type InterpreterStack struct {
	stack   *utils.Stack[*InterpreterScopeDetails]
	runner  sql.StatementRunner
	labelID int
}

// NewInterpreterStack creates a new InterpreterStack.
func NewInterpreterStack(runner sql.StatementRunner) InterpreterStack {
	stack := utils.NewStack[*InterpreterScopeDetails]()
	// This first push represents the function base, including parameters
	stack.Push(&InterpreterScopeDetails{
		variables: make(map[string]*interpreterVariable),
	})
	return InterpreterStack{
		stack:  stack,
		runner: runner,
	}
}

// Details returns the details for the current scope.
func (is *InterpreterStack) Details() *InterpreterScopeDetails {
	return is.stack.Peek()
}

// Runner returns the runner that is being used for the function's execution.
func (is *InterpreterStack) Runner() sql.StatementRunner {
	return is.runner
}

// GetCurrentLabel traverses the stack (starting from the top) returning the first label found. Returns an empty string
// if no labels were set.
func (is *InterpreterStack) GetCurrentLabel() string {
	for i := 0; i < is.stack.Len(); i++ {
		label := is.stack.PeekDepth(i).label
		if len(label) > 0 {
			return label
		}
	}
	return ""
}

// GetVariable traverses the stack (starting from the top) to find a variable with a matching name. Returns nil if no
// variable was found.
func (is *InterpreterStack) GetVariable(name string) InterpreterVariableReference {
	// TODO: handle nested record access
	fieldName := ""
	if strings.Count(name, ".") == 1 {
		splitName := strings.Split(name, ".")
		name = splitName[0]
		fieldName = splitName[1]
	}
	for i := 0; i < is.stack.Len(); i++ {
		if iv, ok := is.stack.PeekDepth(i).variables[name]; ok {
			if len(fieldName) == 0 {
				return InterpreterVariableReference{
					Type:  iv.Type,
					Value: &iv.Value,
				}
			} else if len(iv.Record) > 0 {
				fieldIdx := iv.Record.IndexOf(fieldName, iv.Record[0].Source)
				if fieldIdx == -1 {
					// TODO: implement this as a proper error for missing record field rather than the generic "variable not found"
					return InterpreterVariableReference{}
				}
				return InterpreterVariableReference{
					Type:  iv.Record[fieldIdx].Type.(*pgtypes.DoltgresType),
					Value: &(iv.Value.(sql.Row)[fieldIdx]),
				}
			} else {
				// Can't access fields on an empty record
				return InterpreterVariableReference{}
			}
		}
	}
	return InterpreterVariableReference{}
}

// ListVariables returns a map with the names of all variables. The attached slice represents field names for records.
// All names are lowercased.
func (is *InterpreterStack) ListVariables() map[string][]string {
	seen := make(map[string][]string)
	for i := 0; i < is.stack.Len(); i++ {
		for varName, iv := range is.stack.PeekDepth(i).variables {
			var fieldNames []string
			if len(iv.Record) > 0 {
				for _, col := range iv.Record {
					fieldNames = append(fieldNames, strings.ToLower(col.Name))
				}
			}
			seen[strings.ToLower(varName)] = fieldNames
		}
	}
	return seen
}

// NewRecord creates a new record in the current scope. If a record with the same name exists in a previous scope, then
// that record will be shadowed until the current scope exits.
func (is *InterpreterStack) NewRecord(name string, sch sql.Schema, val sql.Row) {
	// TODO: this is currently implemented only for the specific record types used in triggers: OLD and NEW
	var newVal sql.Row
	if val != nil {
		newVal = make(sql.Row, len(val))
		copy(newVal, val)
	}
	is.stack.Peek().variables[name] = &interpreterVariable{
		Record: sch,
		Type:   pgtypes.Trigger, // TODO: we need to implement the RECORD pseudotype and replace the TRIGGER type here
		Value:  newVal,
	}
}

// NewVariable creates a new variable in the current scope. If a variable with the same name exists in a previous scope,
// then that variable will be shadowed until the current scope exits.
func (is *InterpreterStack) NewVariable(name string, typ *pgtypes.DoltgresType) {
	is.NewVariableWithValue(name, typ, typ.Zero())
}

// NewVariableWithValue creates a new variable in the current scope, setting its initial value to the one given.
func (is *InterpreterStack) NewVariableWithValue(name string, typ *pgtypes.DoltgresType, val any) {
	is.stack.Peek().variables[name] = &interpreterVariable{
		Type:  typ,
		Value: val,
	}
}

// NewVariableAlias creates a new variable alias, named |alias|, in the current frame of this stack,
// pointing to the specified |variable|.
func (is *InterpreterStack) NewVariableAlias(alias string, target string) {
	for i := 0; i < is.stack.Len(); i++ {
		if iv, ok := is.stack.PeekDepth(i).variables[target]; ok {
			// TODO: this won't work for RECORD types
			is.stack.Peek().variables[alias] = iv
			break
		}
	}
}

// PushScope creates a new scope.
func (is *InterpreterStack) PushScope() {
	is.stack.Push(&InterpreterScopeDetails{
		variables: make(map[string]*interpreterVariable),
	})
}

// PopScope removes the current scope.
func (is *InterpreterStack) PopScope() {
	is.stack.Pop()
}

// SetVariable sets the first variable found, with a matching name, to the value given. This does not ensure that the
// value matches the expectations of the type, so it should be validated before this is called. Returns an error if the
// variable cannot be found.
func (is *InterpreterStack) SetVariable(ctx *sql.Context, name string, val any) error {
	iv := is.GetVariable(name)
	if iv.Type == nil {
		return fmt.Errorf("variable `%s` could not be found", name)
	}
	*iv.Value = val
	return nil
}

// SetLabel sets the label for the current scope.
func (is *InterpreterStack) SetLabel(label string) {
	is.stack.Peek().label = label
}

// SetAnonymousLabel sets the label for the current scope to a guaranteed unique value.
func (is *InterpreterStack) SetAnonymousLabel() {
	// Postgres labels cannot have a tab character, so we can generate a label with one to guarantee it's unique
	is.stack.Peek().label = fmt.Sprintf("\t%d", is.labelID)
	is.labelID++
}
