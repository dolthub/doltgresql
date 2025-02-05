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
)

// Statement represents a PL/pgSQL statement.
type Statement interface {
	// OperationSize reports the number of operations that the statement will convert to.
	OperationSize() int32
	// AppendOperations adds the statement to the operation slice.
	AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack)
}

// Assignment represents an assignment statement.
type Assignment struct {
	VariableName  string
	Expression    string
	VariableIndex int32 // TODO: figure out what this is used for, probably to get around shadowed variables?
}

var _ Statement = Assignment{}

// OperationSize implements the interface Statement.
func (Assignment) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (stmt Assignment) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) {
	// TODO: figure out how I'm supposed to actually do this, rather than a search and replace
	expression := stmt.Expression
	var referencedVariables []string
	for varName := range stack.ListVariables() {
		if strings.Contains(expression, varName) {
			referencedVariables = append(referencedVariables, varName)
		}
		expression = strings.Replace(expression, varName, fmt.Sprintf("$%d", len(referencedVariables)), 1)
	}
	*ops = append(*ops, InterpreterOperation{
		OpCode:        OpCode_Assign,
		PrimaryData:   "SELECT " + expression + ";",
		SecondaryData: referencedVariables,
		Target:        stmt.VariableName,
	})
}

// Block contains a collection of statements, alongside the variables that were declared for the block. Only the
// top-level block will contain parameter variables.
type Block struct {
	Variable []Variable
	Body     []Statement
}

var _ Statement = Block{}

// OperationSize implements the interface Statement.
func (stmt Block) OperationSize() int32 {
	total := int32(2) // We start with 2 since we'll have ScopeBegin and ScopeEnd
	for _, variable := range stmt.Variable {
		if !variable.IsParameter {
			total++
		}
	}
	for _, innerStmt := range stmt.Body {
		total += innerStmt.OperationSize()
	}
	return total
}

// AppendOperations implements the interface Statement.
func (stmt Block) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) {
	stack.PushScope()
	*ops = append(*ops, InterpreterOperation{
		OpCode: OpCode_ScopeBegin,
	})
	for _, variable := range stmt.Variable {
		if !variable.IsParameter {
			*ops = append(*ops, InterpreterOperation{
				OpCode:      OpCode_Declare,
				PrimaryData: variable.Type,
				Target:      variable.Name,
			})
		}
		stack.NewVariableWithValue(variable.Name, nil, nil)
	}
	for _, innerStmt := range stmt.Body {
		innerStmt.AppendOperations(ops, stack)
	}
	*ops = append(*ops, InterpreterOperation{
		OpCode: OpCode_ScopeEnd,
	})
	stack.PopScope()
}

// Goto jumps to the counter at the given offset.
type Goto struct {
	Offset int32
}

var _ Statement = Goto{}

// OperationSize implements the interface Statement.
func (Goto) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (stmt Goto) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) {
	*ops = append(*ops, InterpreterOperation{
		OpCode: OpCode_Goto,
		Index:  len(*ops) + int(stmt.Offset),
	})
}

// If represents an IF condition, alongside its Goto offset if the condition is true.
type If struct {
	Condition  string
	GotoOffset int32
}

var _ Statement = If{}

// OperationSize implements the interface Statement.
func (If) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (stmt If) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) {
	condition := stmt.Condition
	var referencedVariables []string
	for varName := range stack.ListVariables() {
		if strings.Contains(condition, varName) {
			referencedVariables = append(referencedVariables, varName)
		}
		condition = strings.Replace(condition, varName, fmt.Sprintf("$%d", len(referencedVariables)), 1)
	}
	*ops = append(*ops, InterpreterOperation{
		OpCode:        OpCode_If,
		PrimaryData:   "SELECT " + condition + ";",
		SecondaryData: referencedVariables,
		Index:         len(*ops) + int(stmt.GotoOffset),
	})
}

// Return represents a RETURN statement.
type Return struct {
	Expression string
}

var _ Statement = Return{}

// OperationSize implements the interface Statement.
func (Return) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (stmt Return) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) {
	// TODO: figure out how I'm supposed to actually do this, rather than a search and replace
	expression := stmt.Expression
	var referencedVariables []string
	for varName := range stack.ListVariables() {
		if strings.Contains(expression, varName) {
			referencedVariables = append(referencedVariables, varName)
		}
		expression = strings.Replace(expression, varName, fmt.Sprintf("$%d", len(referencedVariables)), 1)
	}
	*ops = append(*ops, InterpreterOperation{
		OpCode:        OpCode_Return,
		PrimaryData:   "SELECT " + expression + ";",
		SecondaryData: referencedVariables,
	})
}

// Variable represents a variable. These are exclusively found within Block.
type Variable struct {
	Name        string
	Type        string
	IsParameter bool
}

// OperationSizeForStatements returns the sum of OperationSize for every statement.
func OperationSizeForStatements(stmts []Statement) int32 {
	total := int32(0)
	for _, stmt := range stmts {
		total += stmt.OperationSize()
	}
	return total
}
