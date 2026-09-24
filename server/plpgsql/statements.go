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
	"sort"
	"strconv"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/cockroachdb/errors"
	pg_query "github.com/dolthub/pg_query_go/v6"
)

// Statement represents a PL/pgSQL statement.
type Statement interface {
	// OperationSize reports the number of operations that the statement will convert to.
	OperationSize() int32
	// AppendOperations adds the statement to the operation slice.
	AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) error
}

// Assignment represents an assignment statement.
type Assignment struct {
	VariableName  string
	Expression    string
	VariableIndex int32 // TODO: figure out what this is used for, probably to get around shadowed variables?
	// RetypeTarget sets the target's type from the value assigned. A CASE statement's variable is declared
	// int4 whatever its expression yields, so its type is only known once the expression has run.
	RetypeTarget bool
}

var _ Statement = Assignment{}

// OperationSize implements the interface Statement.
func (Assignment) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (stmt Assignment) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) error {
	expression, referencedVariables, err := substituteVariableReferences(stmt.Expression, stack)
	if err != nil {
		return err
	}

	op := InterpreterOperation{
		OpCode:        OpCode_Assign,
		PrimaryData:   "SELECT " + expression + ";",
		SecondaryData: referencedVariables,
		Target:        stmt.VariableName,
	}
	if stmt.RetypeTarget {
		op.Options = map[string]string{OptionRetypeTarget: "true"}
	}
	*ops = append(*ops, op)
	return nil
}

// Block contains a collection of statements, alongside the variables that were declared for the block. Only the
// top-level block will contain parameter variables.
type Block struct {
	TriggerNew int32 // When non-zero, indicates that the NEW record exists for use with triggers
	TriggerOld int32 // When non-zero, indicates that the OLD record exists for use with triggers
	Variables  []Variable
	Records    []Record
	Body       []Statement
	Label      string
	IsLoop     bool
	// ContinueTargetOffset gives the loop's next-iteration operation, where a CONTINUE for this loop jumps, as
	// an offset from the body's first operation. It applies only when IsLoop is true, and the zero value suits
	// WHILE and plain LOOP, whose bodies begin with that step rather than with loop setup.
	ContinueTargetOffset int32
}

var _ Statement = Block{}

// OperationSize implements the interface Statement.
func (stmt Block) OperationSize() int32 {
	total := int32(2) // We start with 2 since we'll have ScopeBegin and ScopeEnd
	for _, variable := range stmt.Variables {
		if !variable.IsParameter {
			total++
		}
	}
	for _, record := range stmt.Records {
		if record.IsDeclared() {
			total++
		}
	}
	for _, innerStmt := range stmt.Body {
		total += innerStmt.OperationSize()
	}
	return total
}

// AppendOperations implements the interface Statement.
func (stmt Block) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) error {
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
	scopeBeginIndex := len(*ops)
	*ops = append(*ops, InterpreterOperation{
		OpCode:      OpCode_ScopeBegin,
		PrimaryData: stmt.Label,
		Target:      loop,
	})
	// NEW and OLD are supplied by the trigger invocation, so they are in scope for every declaration.
	for _, record := range stmt.Records {
		if !record.IsDeclared() {
			stack.NewRecord(record.Name, record.fakeSchema(), nil)
		}
	}
	// Everything else is declared in the order it was written, so that a default expression sees the
	// parameters and the declarations ahead of it, as `r RECORD := ROW(n)` or `id int := OLD.id` does.
	for _, decl := range stmt.orderedDeclarations() {
		if decl.record != nil {
			record := decl.record
			// The schema here only exists so that field references such as `r.id` are recognized as
			// variable references while the body is compiled. The real schema is not known until the
			// record is assigned.
			stack.NewRecord(record.Name, record.fakeSchema(), nil)
			op := InterpreterOperation{
				OpCode: OpCode_DeclareRecord,
				Target: record.Name,
			}
			if record.Default != "" {
				query, referencedVariables, err := compileRecordDeclareDefault(record.Default, stack)
				if err != nil {
					return err
				}
				op.SecondaryData = append([]string{record.Default, query}, referencedVariables...)
			}
			*ops = append(*ops, op)
			continue
		}
		variable := decl.variable
		op := InterpreterOperation{
			OpCode:      OpCode_Declare,
			PrimaryData: variable.Type,
			Target:      variable.Name,
		}
		if variable.Default != "" {
			// A default is an arbitrary expression, so it compiles like the right-hand side of an
			// assignment. Registering each variable as we go leaves only those declared ahead of
			// this one in scope, matching PostgreSQL's evaluation of defaults in declaration order.
			query, referencedVariables, err := compileDeclareDefault(variable.Default, stack)
			if err != nil {
				return err
			}
			op.SecondaryData = append([]string{variable.Default, query}, referencedVariables...)
		}
		if !variable.IsParameter {
			*ops = append(*ops, op)
		}
		// This stack only resolves names; the variable's type and value are not known until the
		// declaration runs.
		stack.NewVariableWithValue(variable.Name, nil, nil)
	}
	if stmt.IsLoop {
		// Declarations are already appended, so the body starts at the next operation. reconcileLabels
		// resolves this loop's CONTINUE statements through this offset.
		continueTarget := len(*ops) + int(stmt.ContinueTargetOffset)
		(*ops)[scopeBeginIndex].Options = map[string]string{
			continueTargetOption: strconv.Itoa(continueTarget - scopeBeginIndex),
		}
	}
	for _, innerStmt := range stmt.Body {
		if err := innerStmt.AppendOperations(ops, stack); err != nil {
			return err
		}
	}
	*ops = append(*ops, InterpreterOperation{
		OpCode: OpCode_ScopeEnd,
	})
	stack.PopScope()
	return nil
}

// declaration is either a Record or a Variable declared by a Block.
type declaration struct {
	record   *Record
	variable *Variable
}

// orderedDeclarations returns the block's declared records and its variables in declaration order.
func (stmt Block) orderedDeclarations() []declaration {
	decls := make([]declaration, 0, len(stmt.Records)+len(stmt.Variables))
	for i := range stmt.Records {
		if stmt.Records[i].IsDeclared() {
			decls = append(decls, declaration{record: &stmt.Records[i]})
		}
	}
	for i := range stmt.Variables {
		decls = append(decls, declaration{variable: &stmt.Variables[i]})
	}
	// Blocks built by the interpreter rather than parsed leave every DatumNumber at zero, and a stable sort
	// keeps those in the order above.
	sort.SliceStable(decls, func(i, j int) bool {
		return decls[i].datumNumber() < decls[j].datumNumber()
	})
	return decls
}

func (decl declaration) datumNumber() int32 {
	if decl.record != nil {
		return decl.record.DatumNumber
	}
	return decl.variable.DatumNumber
}

// ExecuteSQL represents a standard SQL statement's execution (including the INTO syntax).
type ExecuteSQL struct {
	Statement string
	Target    string
	// TargetIsRecord states that Target names a single RECORD variable that receives the entire result row,
	// rather than a comma-separated list of scalar variables that each receive one column.
	TargetIsRecord bool
	// SetsFound states that the statement updates the built-in FOUND variable. PostgreSQL limits that to
	// statements carrying an INTO clause and to data-modifying statements. Everything else leaves FOUND
	// as it was, a utility statement such as CREATE TABLE in particular.
	SetsFound bool
}

// isDataModifying reports whether |query| is an INSERT, UPDATE, DELETE, or MERGE. Those are the statements
// PostgreSQL treats as data-modifying when deciding whether to update FOUND; its own compiler records the
// same thing as PLpgSQL_stmt_execsql.mod_stmt. A query that does not parse is reported as not data-modifying,
// since leaving FOUND alone is what every statement outside this set does.
func isDataModifying(query string) bool {
	result, err := pg_query.Parse(query)
	if err != nil {
		return false
	}
	for _, rawStmt := range result.GetStmts() {
		switch rawStmt.GetStmt().GetNode().(type) {
		case *pg_query.Node_InsertStmt, *pg_query.Node_UpdateStmt,
			*pg_query.Node_DeleteStmt, *pg_query.Node_MergeStmt:
			return true
		}
	}
	return false
}

var _ Statement = ExecuteSQL{}

// OperationSize implements the interface Statement.
func (ExecuteSQL) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (stmt ExecuteSQL) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) error {
	statementStr, referencedVariables, err := substituteVariableReferences(stmt.Statement, stack)
	if err != nil {
		return err
	}
	op := InterpreterOperation{
		OpCode:        executeOpCode(stmt.TargetIsRecord),
		PrimaryData:   statementStr,
		SecondaryData: referencedVariables,
		Target:        stmt.Target,
	}
	if stmt.SetsFound {
		op.Options = map[string]string{OptionSetsFound: "true"}
	}
	*ops = append(*ops, op)
	return nil
}

// DynamicExecute represents a dynamic SQL statement's execution.
type DynamicExecute struct {
	Query  string
	Params []string
	Target string
	// TargetIsRecord states that Target names a single RECORD variable that receives the entire result row,
	// rather than a comma-separated list of scalar variables that each receive one column.
	TargetIsRecord bool
}

var _ Statement = DynamicExecute{}

// OperationSize implements the interface Statement.
func (DynamicExecute) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (stmt DynamicExecute) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) error {
	query, bindings, err := substituteVariableReferences(stmt.Query, stack)
	if err != nil {
		return err
	}
	options := map[string]string{
		OptionDynamicExpression:   "true",
		OptionDynamicBindingCount: strconv.Itoa(len(bindings)),
	}
	for i, binding := range bindings {
		options[OptionDynamicBindingPrefix+strconv.Itoa(i)] = binding
	}
	options[OptionDynamicUsingCount] = strconv.Itoa(len(stmt.Params))
	for i, param := range stmt.Params {
		expression, paramBindings, err := substituteVariableReferences(param, stack)
		if err != nil {
			return err
		}
		index := strconv.Itoa(i)
		options[OptionDynamicUsingExpressionPrefix+index] = expression
		options[OptionDynamicUsingBindingCountPrefix+index] = strconv.Itoa(len(paramBindings))
		for j, binding := range paramBindings {
			options[OptionDynamicUsingBindingPrefix+index+"_"+strconv.Itoa(j)] = binding
		}
	}
	*ops = append(*ops, InterpreterOperation{
		OpCode:      executeOpCode(stmt.TargetIsRecord),
		PrimaryData: query,
		Target:      stmt.Target,
		Options:     options,
	})
	return nil
}

// executeOpCode returns the opcode used to run a SQL statement whose results are written into the given
// kind of INTO target.
func executeOpCode(targetIsRecord bool) OpCode {
	if targetIsRecord {
		return OpCode_ExecuteInto
	}
	return OpCode_Execute
}

// ForQueryInit executes a SQL query and stores the result set as the cursor of the scope it runs in. The
// rows are a FOR record IN query LOOP's own query, or the elements of a FOREACH's array.
type ForQueryInit struct {
	Query string
}

var _ Statement = ForQueryInit{}

// OperationSize implements the interface Statement.
func (ForQueryInit) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (stmt ForQueryInit) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) error {
	queryStr, referencedVariables, err := substituteVariableReferences(stmt.Query, stack)
	if err != nil {
		return err
	}
	*ops = append(*ops, InterpreterOperation{
		OpCode:        OpCode_ForQueryInit,
		PrimaryData:   queryStr,
		SecondaryData: referencedVariables,
	})
	return nil
}

// ForQueryNext fetches the next row from the cursor of the scope it runs in and assigns it to a record
// variable. When the cursor is exhausted it jumps forward by GotoOffset (like an If), exiting the loop.
type ForQueryNext struct {
	RecordVar  string
	GotoOffset int32
}

var _ Statement = ForQueryNext{}

// OperationSize implements the interface Statement.
func (ForQueryNext) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (stmt ForQueryNext) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) error {
	*ops = append(*ops, InterpreterOperation{
		OpCode: OpCode_ForQueryNext,
		Target: stmt.RecordVar,
		Index:  len(*ops) + int(stmt.GotoOffset),
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
func (stmt Goto) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) error {
	if len(stmt.Label) > 0 {
		*ops = append(*ops, InterpreterOperation{
			OpCode:      OpCode_Goto,
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
		*ops = append(*ops, InterpreterOperation{
			OpCode:      OpCode_Goto,
			PrimaryData: label,
			Index:       int(stmt.Offset),
		})
	} else {
		*ops = append(*ops, InterpreterOperation{
			OpCode: OpCode_Goto,
			Index:  len(*ops) + int(stmt.Offset),
		})
	}
	return nil
}

// If represents an IF condition, alongside its Goto offset if the condition is true.
type If struct {
	Condition  string
	GotoOffset int32
	// IsLoopCondition marks this as the conditional jump that advances an integer FOR loop, whose result
	// is the only record of the loop having run its body.
	IsLoopCondition bool
}

var _ Statement = If{}

// OperationSize implements the interface Statement.
func (If) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (stmt If) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) error {
	condition, referencedVariables, err := substituteVariableReferences(stmt.Condition, stack)
	if err != nil {
		return err
	}

	op := InterpreterOperation{
		OpCode:        OpCode_If,
		PrimaryData:   "SELECT " + condition + ";",
		SecondaryData: referencedVariables,
		Index:         len(*ops) + int(stmt.GotoOffset),
	}
	if stmt.IsLoopCondition {
		op.Options = map[string]string{OptionLoopCondition: "true"}
	}
	*ops = append(*ops, op)
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
func (stmt Perform) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) error {
	statementStr, referencedVariables, err := substituteVariableReferences(stmt.Statement, stack)
	if err != nil {
		return err
	}

	*ops = append(*ops, InterpreterOperation{
		OpCode:        OpCode_Perform,
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
	Options map[string]string
	// SqlState gives the SQLSTATE that an EXCEPTION-level RAISE reports, and is empty for a RAISE that does
	// not name one. A RAISE written in a function body carries its code in Options, put there by the USING
	// clause's ERRCODE option; this is for the RAISE statements the compiler generates itself, which have no
	// source text to carry one.
	SqlState string
}

var _ Statement = Raise{}

// OperationSize implements the interface Statement.
func (r Raise) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (r Raise) AppendOperations(ops *[]InterpreterOperation, _ *InterpreterStack) error {
	options := r.Options
	if len(r.SqlState) > 0 {
		// The statement's own options are left alone, since a Statement may be appended more than once.
		options = make(map[string]string, len(r.Options)+1)
		for key, value := range r.Options {
			options[key] = value
		}
		options[errCodeOptionKey] = r.SqlState
	}
	*ops = append(*ops, InterpreterOperation{
		OpCode:        OpCode_Raise,
		PrimaryData:   r.Level,
		SecondaryData: append([]string{r.Message}, r.Params...),
		Options:       options,
	})
	return nil
}

// Record represents a record (along with known fields for future access). These are exclusively found within Block.
type Record struct {
	Name    string
	Fields  []string
	Default string
	// DatumNumber is the record's position among the function's declarations.
	DatumNumber int32
	// IsTriggerRecord is true for the NEW and OLD records of a trigger function. Those are created by the
	// trigger invocation rather than by the function body, so they are not declared when the block is entered.
	IsTriggerRecord bool
}

// IsDeclared returns whether entering the record's block should declare it. Trigger records are supplied by
// the trigger invocation, and PL/pgSQL leaves a record's name empty when it is only referenced internally.
func (record Record) IsDeclared() bool {
	return !record.IsTriggerRecord && len(record.Name) > 0
}

// fakeSchema returns a schema naming the record's known fields, with no types.
func (record Record) fakeSchema() sql.Schema {
	var fakeSch sql.Schema
	for _, fieldName := range record.Fields {
		fakeSch = append(fakeSch, &sql.Column{Name: fieldName})
	}
	return fakeSch
}

// ReturnQuery represents a RETURN QUERY statement.
type ReturnQuery struct {
	Query string
}

var _ Statement = ReturnQuery{}

// OperationSize implements the interface Statement.
func (r ReturnQuery) OperationSize() int32 {
	return 1
}

// AppendOperations implements the interface Statement.
func (r ReturnQuery) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) error {
	query, referencedVariables, err := substituteVariableReferences(r.Query, stack)
	if err != nil {
		return err
	}

	*ops = append(*ops, InterpreterOperation{
		OpCode:        OpCode_ReturnQuery,
		PrimaryData:   query,
		SecondaryData: referencedVariables,
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
func (stmt Return) AppendOperations(ops *[]InterpreterOperation, stack *InterpreterStack) error {
	expression, referencedVariables, err := substituteVariableReferences(stmt.Expression, stack)
	if err != nil {
		return err
	}
	if len(expression) > 0 {
		expression = "SELECT " + expression + ";"
	}
	*ops = append(*ops, InterpreterOperation{
		OpCode:        OpCode_Return,
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
	Default     string
	// DatumNumber is the variable's position among the function's declarations.
	DatumNumber int32
}

// OperationSizeForStatements returns the sum of OperationSize for every statement.
func OperationSizeForStatements(stmts []Statement) int32 {
	total := int32(0)
	for _, stmt := range stmts {
		total += stmt.OperationSize()
	}
	return total
}

// compileDeclareDefault compiles the source text of a declaration's default into the query that
// evaluates it, along with the names of the variables that query binds. Whatever the |stack| holds is
// in scope for the default.
func compileDeclareDefault(defaultText string, stack *InterpreterStack) (query string, bindings []string, err error) {
	expression, bindings, err := substituteVariableReferences(defaultText, stack)
	if err != nil {
		return "", nil, err
	}
	return "SELECT " + expression + ";", bindings, nil
}

// compileRecordDeclareDefault evaluates a RECORD default like the right-hand side of an assignment.
func compileRecordDeclareDefault(defaultText string, stack *InterpreterStack) (query string, bindings []string, err error) {
	expression, bindings, err := substituteVariableReferences(defaultText, stack)
	if err != nil {
		return "", nil, err
	}
	return "SELECT " + expression + ";", bindings, nil
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
	for i := 0; i < len(scanResult.Tokens); i++ {
		token := scanResult.Tokens[i]
		substring := expression[token.Start:token.End]
		// varMap lowercases everything, so we'll lowercase our substring to enable case-insensitivity
		isAfterDot := i > 0 && scanResult.Tokens[i-1].Token == '.'

		if !isAfterDot {
			// A variable is named by whatever the reference folds to, not by how it happens to be spelled
			// here, so the binding is recorded under the folded name.
			normalized := NormalizeIdentifier(substring)
			if _, ok := varMap[normalized]; ok {
				bindingName := normalized
				// If there's a '.', then we'll assume this is accessing a record's field (`NEW.val1` for example)
				for i+2 < len(scanResult.Tokens) && scanResult.Tokens[i+1].Token == '.' {
					nextFieldSubstring := expression[scanResult.Tokens[i+2].Start:scanResult.Tokens[i+2].End]
					substring += "." + nextFieldSubstring
					bindingName += "." + nextFieldSubstring
					i += 2
				}
				// Variables cannot have a '(' after their name as that would classify them as functions, so we have to
				// explicitly check for that. This is because variables and functions can share names, for example:
				// SELECT COUNT(*) INTO count FROM table_name;
				if i+1 >= len(scanResult.Tokens) || scanResult.Tokens[i+1].Token != '(' {
					referencedVars = append(referencedVars, bindingName)
					newExpression += fmt.Sprintf("$%d ", len(referencedVars))
				} else {
					newExpression += substring + " "
				}
			} else if _, ok := triggerSpecialVariables[normalized]; ok {
				referencedVars = append(referencedVars, normalized)
				newExpression += fmt.Sprintf("$%d ", len(referencedVars))
			} else {
				newExpression += substring + " "
			}
		} else {
			newExpression += substring + " "
		}
	}

	return newExpression, referencedVars, nil
}
