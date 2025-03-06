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

	"github.com/cockroachdb/errors"
	pg_query "github.com/pganalyze/pg_query_go/v6"

	"github.com/dolthub/doltgresql/core/interpreter"
)

// Statement represents a PL/pgSQL statement.
type Statement interface {
	// OperationSize reports the number of operations that the statement will convert to.
	OperationSize() int32
	// AppendOperations adds the statement to the operation slice.
	AppendOperations(ops *[]interpreter.InterpreterOperation, stack *InterpreterStack) error
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
func (stmt Assignment) AppendOperations(ops *[]interpreter.InterpreterOperation, stack *InterpreterStack) error {
	expression, referencedVariables, err := substituteVariableReferences(stmt.Expression, stack)
	if err != nil {
		return err
	}

	*ops = append(*ops, interpreter.InterpreterOperation{
		OpCode:        interpreter.OpCode_Assign,
		PrimaryData:   "SELECT " + expression + ";",
		SecondaryData: referencedVariables,
		Target:        stmt.VariableName,
	})
	return nil
}

// Block contains a collection of statements, alongside the variables that were declared for the block. Only the
// top-level block will contain parameter variables.
type Block struct {
	Variable []Variable
	Body     []Statement
	Label    string
	IsLoop   bool
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
func (stmt Block) AppendOperations(ops *[]interpreter.InterpreterOperation, stack *InterpreterStack) error {
	stack.PushScope()
	stack.SetLabel(stmt.Label) // If the label is empty, then this won't change anything
	var loop string
	if stmt.IsLoop {
		loop = "_"
		// All loops need a label, so we'll make an anonymous one if an explicit one hasn't been given
		if len(stmt.Label) == 0 {
			stack.SetAnonymousLabel()
			stmt.Label = stack.GetCurrentLabel()
		}
	}
	*ops = append(*ops, interpreter.InterpreterOperation{
		OpCode:      interpreter.OpCode_ScopeBegin,
		PrimaryData: stmt.Label,
		Target:      loop,
	})
	for _, variable := range stmt.Variable {
		if !variable.IsParameter {
			*ops = append(*ops, interpreter.InterpreterOperation{
				OpCode:      interpreter.OpCode_Declare,
				PrimaryData: variable.Type,
				Target:      variable.Name,
			})
		}
		stack.NewVariableWithValue(variable.Name, nil, nil)
	}
	for _, innerStmt := range stmt.Body {
		if err := innerStmt.AppendOperations(ops, stack); err != nil {
			return err
		}
	}
	*ops = append(*ops, interpreter.InterpreterOperation{
		OpCode: interpreter.OpCode_ScopeEnd,
	})
	stack.PopScope()
	return nil
}

// ExecuteSQL represents a standard SQL statement's execution (including the INTO syntax).
type ExecuteSQL struct {
	Statement string
	Target    string
}

var _ Statement = ExecuteSQL{}

// OperationSize implements the interface Statement.
func (ExecuteSQL) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (stmt ExecuteSQL) AppendOperations(ops *[]interpreter.InterpreterOperation, stack *InterpreterStack) error {
	statementStr, referencedVariables, err := substituteVariableReferences(stmt.Statement, stack)
	if err != nil {
		return err
	}
	*ops = append(*ops, interpreter.InterpreterOperation{
		OpCode:        interpreter.OpCode_Execute,
		PrimaryData:   statementStr,
		SecondaryData: referencedVariables,
		Target:        stmt.Target,
	})
	return nil
}

// Goto jumps to the counter at the given offset.
type Goto struct {
	Offset         int32
	Label          string
	NearestScopeOp bool
}

var _ Statement = Goto{}

// OperationSize implements the interface Statement.
func (Goto) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (stmt Goto) AppendOperations(ops *[]interpreter.InterpreterOperation, stack *InterpreterStack) error {
	if len(stmt.Label) > 0 {
		*ops = append(*ops, interpreter.InterpreterOperation{
			OpCode:      interpreter.OpCode_Goto,
			PrimaryData: stmt.Label,
			Index:       int(stmt.Offset),
		})
	} else if stmt.NearestScopeOp {
		label := stack.GetCurrentLabel()
		if len(label) == 0 {
			if stmt.Offset > 0 {
				return errors.New("EXIT cannot be used outside a loop, unless it has a label")
			} else {
				return errors.New("CONTINUE cannot be used outside a loop")
			}
		}
		*ops = append(*ops, interpreter.InterpreterOperation{
			OpCode:      interpreter.OpCode_Goto,
			PrimaryData: label,
			Index:       int(stmt.Offset),
		})
	} else {
		*ops = append(*ops, interpreter.InterpreterOperation{
			OpCode: interpreter.OpCode_Goto,
			Index:  len(*ops) + int(stmt.Offset),
		})
	}
	return nil
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
func (stmt If) AppendOperations(ops *[]interpreter.InterpreterOperation, stack *InterpreterStack) error {
	condition, referencedVariables, err := substituteVariableReferences(stmt.Condition, stack)
	if err != nil {
		return err
	}

	*ops = append(*ops, interpreter.InterpreterOperation{
		OpCode:        interpreter.OpCode_If,
		PrimaryData:   "SELECT " + condition + ";",
		SecondaryData: referencedVariables,
		Index:         len(*ops) + int(stmt.GotoOffset),
	})
	return nil
}

// Perform represents a PERFORM statement.
type Perform struct {
	Statement string
}

var _ Statement = Perform{}

// OperationSize implements the interface Statement.
func (Perform) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (stmt Perform) AppendOperations(ops *[]interpreter.InterpreterOperation, stack *InterpreterStack) error {
	statementStr, referencedVariables, err := substituteVariableReferences(stmt.Statement, stack)
	if err != nil {
		return err
	}

	*ops = append(*ops, interpreter.InterpreterOperation{
		OpCode:        interpreter.OpCode_Perform,
		PrimaryData:   statementStr,
		SecondaryData: referencedVariables,
	})
	return nil
}

// Raise represents a RAISE statement
type Raise struct {
	Level   string
	Message string
	Params  []string
	Options map[uint8]string
}

var _ Statement = Raise{}

// OperationSize implements the interface Statement.
func (r Raise) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (r Raise) AppendOperations(ops *[]interpreter.InterpreterOperation, stack *InterpreterStack) error {
	*ops = append(*ops, interpreter.InterpreterOperation{
		OpCode:        interpreter.OpCode_Raise,
		PrimaryData:   r.Level,
		SecondaryData: append([]string{r.Message}, r.Params...),
		Options:       r.Options,
	})
	return nil
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
func (stmt Return) AppendOperations(ops *[]interpreter.InterpreterOperation, stack *InterpreterStack) error {
	expression, referencedVariables, err := substituteVariableReferences(stmt.Expression, stack)
	if err != nil {
		return err
	}
	if len(expression) > 0 {
		expression = "SELECT " + expression + ";"
	}
	*ops = append(*ops, interpreter.InterpreterOperation{
		OpCode:        interpreter.OpCode_Return,
		PrimaryData:   expression,
		SecondaryData: referencedVariables,
	})
	return nil
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
