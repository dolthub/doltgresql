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

	pg_query "github.com/pganalyze/pg_query_go/v6"
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
	expression, referencedVariables, err := substituteVariableReferences(stmt.Expression, stack)
	if err != nil {
		// TODO: add an error return param instead of panicing
		panic(err)
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
	condition, referencedVariables, err := substituteVariableReferences(stmt.Condition, stack)
	if err != nil {
		// TODO: add an error return param instead of panicing
		panic(err)
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
	expression, referencedVariables, err := substituteVariableReferences(stmt.Expression, stack)
	if err != nil {
		// TODO: add an error return param instead of panicing
		panic(err)
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

// substituteVariableReferences parses the specified |expression| and replaces
// any token that matches a variable name in the |stack| with "$N", where N
// indicates which variable in the returned |referenceVars| slice is used.
func substituteVariableReferences(expression string, stack *InterpreterStack) (newExpression string, referencedVars []string, err error) {
	scanResult, err := pg_query.Scan(expression)
	if err != nil {
		return "", nil, err
	}

	varMap := stack.ListVariables()
	for _, token := range scanResult.Tokens {
		substring := expression[token.Start:token.End]
		if _, ok := varMap[substring]; ok {
			referencedVars = append(referencedVars, substring)
			newExpression += fmt.Sprintf("$%d ", len(referencedVars))
		} else {
			newExpression += substring + " "
		}
	}

	return newExpression, referencedVars, nil
}
