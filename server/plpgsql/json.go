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
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
)

// action exists to match the expected JSON format.
type action struct {
	StmtBlock plpgSQL_stmt_block `json:"PLpgSQL_stmt_block"`
}

// cond exists to match the expected JSON format.
type cond struct {
	Expression plpgSQL_expr `json:"PLpgSQL_expr"`
}

// datatype exists to match the expected JSON format.
type datatype struct {
	Type plpgSQL_type `json:"PLpgSQL_type"`
}

// default_val exists to match the expected JSON format.
type default_val struct {
	Var plpgSQL_expr `json:"PLpgSQL_expr"`
}

// datum exists to match the expected JSON format.
type datum struct {
	Record      *plpgSQL_rec      `json:"PLpgSQL_rec"`
	RecordField *plpgSQL_recfield `json:"PLpgSQL_recfield"`
	Row         *plpgSQL_row      `json:"PLpgSQL_row"`
	Variable    *plpgSQL_var      `json:"PLpgSQL_var"`
}

// elsif exists to match the expected JSON format.
type elsif struct {
	ElseIf plpgSQL_if_elsif `json:"PLpgSQL_if_elsif"`
}

// expr exists to match the expected JSON format.
type expr struct {
	Expression plpgSQL_expr `json:"PLpgSQL_expr"`
}

// field exists to match the expected JSON format.
type field struct {
	Name           string `json:"name"`
	VariableNumber int32  `json:"varno"`
}

// function exists to match the expected JSON format.
type function struct {
	Function plpgSQL_block `json:"PLpgSQL_function"`
}

// plpgSQL_block exists to match the expected JSON format.
type plpgSQL_block struct {
	NewVariableNumber int32   `json:"new_varno"`
	OldVariableNumber int32   `json:"old_varno"`
	Datums            []datum `json:"datums"`
	Action            action  `json:"action"`
}

// plpgSQL_expr exists to match the expected JSON format.
type plpgSQL_expr struct {
	Query     string `json:"query"`
	ParseMode int32  `json:"parseMode"`
}

// plpgSQL_if_elsif exists to match the expected JSON format.
type plpgSQL_if_elsif struct {
	Condition  cond        `json:"cond"`
	Then       []statement `json:"stmts"`
	LineNumber int32       `json:"lineno"`
}

// plpgSQL_rec exists to match the expected JSON format.
type plpgSQL_rec struct {
	RefName     string `json:"refname"`
	DatumNumber int32  `json:"dno"`
}

// plpgSQL_recfield exists to match the expected JSON format.
type plpgSQL_recfield struct {
	FieldName          string `json:"fieldname"`
	RecordParentNumber int32  `json:"recparentno"`
}

// plpgSQL_row exists to match the expected JSON format.
type plpgSQL_row struct {
	RefName    string  `json:"refname"`
	Fields     []field `json:"fields"`
	LineNumber int32   `json:"lineno"`
}

// plpgSQL_stmt_assign exists to match the expected JSON format.
type plpgSQL_stmt_assign struct {
	Expression     expr  `json:"expr"`
	VariableNumber int32 `json:"varno"`
	LineNumber     int32 `json:"lineno"`
}

// plpgSQL_stmt_block exists to match the expected JSON format.
type plpgSQL_stmt_block struct {
	Body       []statement `json:"body"`
	Label      string      `json:"label"`
	LineNumber int32       `json:"lineno"`
}

// plpgSQL_stmt_case exists to match the expected JSON format.
type plpgSQL_stmt_case struct {
	LineNumber int32 `json:"lineno"`
	Expression expr  `json:"t_expr"`
	// VarNo indicates the ID for the __Case__Variable_N__ variable that holds the evaluated
	// value of the case expression.
	VarNo    int32       `json:"t_varno"`
	WhenList []statement `json:"case_when_list"`
	HasElse  bool        `json:"have_else"`
	Else     []statement `json:"else_stmts"`
}

// plpgSQL_stmt_dynexecute exists to match the expected JSON format.
type plpgSQL_stmt_dynexecute struct {
	LineNumber int32     `json:"lineno"`
	Into       bool      `json:"into"`
	Query      expr      `json:"query"`
	Target     datum     `json:"target"`
	Params     []sqlstmt `json:"params"`
}

// plpgSQL_case_when exists to match the expected JSON format.
type plpgSQL_case_when struct {
	LineNumber int32       `json:"lineno"`
	Expression expr        `json:"expr"`
	Body       []statement `json:"stmts"`
}

// plpgSQL_stmt_execsql exists to match the expected JSON format.
type plpgSQL_stmt_execsql struct {
	SQLStmt    sqlstmt `json:"sqlstmt"`
	LineNumber int32   `json:"lineno"`
	Into       bool    `json:"into"`
	Target     datum   `json:"target"`
}

// plpgSQL_stmt_exit exists to match the expected JSON format.
type plpgSQL_stmt_exit struct {
	Label      string `json:"label"`
	IsExit     bool   `json:"is_exit"`
	Condition  *expr  `json:"cond"`
	LineNumber int32  `json:"lineno"`
}

// plpgSQL_stmt_fori exists to match the expected JSON format.
type plpgSQL_stmt_fori struct {
	Label      string      `json:"label"`
	Var        datum       `json:"var"`
	Lower      *expr       `json:"lower"`
	Upper      *expr       `json:"upper"`
	Step       *expr       `json:"step"`
	Reverse    bool        `json:"reverse"`
	Body       []statement `json:"body"`
	LineNumber int32       `json:"lineno"`
}

// plpgSQL_stmt_fors exists to match the expected JSON format.
type plpgSQL_stmt_fors struct {
	Label      string      `json:"label"`
	Var        datum       `json:"var"`
	Body       []statement `json:"body"`
	Query      *expr       `json:"query"`
	LineNumber int32       `json:"lineno"`
}

// plpgSQL_stmt_if exists to match the expected JSON format.
type plpgSQL_stmt_if struct {
	Condition  cond        `json:"cond"`
	Then       []statement `json:"then_body"`
	ElseIf     []elsif     `json:"elsif_list"`
	Else       []statement `json:"else_body"`
	LineNumber int32       `json:"lineno"`
}

// plpgSQL_stmt_loop exists to match the expected JSON format.
type plpgSQL_stmt_loop struct {
	Body       []statement `json:"body"`
	Label      string      `json:"label"`
	LineNumber int32       `json:"lineno"`
}

// plpgSQL_stmt_perform exists to match the expected JSON format.
type plpgSQL_stmt_perform struct {
	Expression expr  `json:"expr"`
	LineNumber int32 `json:"lineno"`
}

// plpgSQL_stmt_raise exists to match the expected JSON format.
type plpgSQL_stmt_raise struct {
	LineNumber int32                          `json:"lineno"`
	ELogLevel  int32                          `json:"elog_level"`
	Message    string                         `json:"message"`
	Params     []sqlstmt                      `json:"params"`
	Options    []plpgSQL_raise_option_wrapper `json:"options"`
}

// plpgSQL_raise_option_wrapper exists to match the expected JSON format.
type plpgSQL_raise_option_wrapper struct {
	Option plpgSQL_raise_option `json:"PLpgSQL_raise_option"`
}

// plpgSQL_raise_option exists to match the expected JSON format.
type plpgSQL_raise_option struct {
	OptionType int32   `json:"opt_type"`
	Expression sqlstmt `json:"expr"`
}

// plpgSQL_stmt_return exists to match the expected JSON format.
type plpgSQL_stmt_return struct {
	Expression expr  `json:"expr"`
	LineNumber int32 `json:"lineno"`
}

// plpgSQL_stmt_return_query exists to match the expected JSON format.
type plpgSQL_stmt_return_query struct {
	Query      expr  `json:"query"`
	LineNumber int32 `json:"lineno"`
}

// plpgSQL_stmt_while exists to match the expected JSON format.
type plpgSQL_stmt_while struct {
	Condition  cond        `json:"cond"`
	Body       []statement `json:"body"`
	Label      string      `json:"label"`
	LineNumber int32       `json:"lineno"`
}

// plpgSQL_type exists to match the expected JSON format.
type plpgSQL_type struct {
	Name string `json:"typname"`
}

// plpgSQL_var exists to match the expected JSON format.
type plpgSQL_var struct {
	RefName    string      `json:"refname"`
	Type       datatype    `json:"datatype"`
	LineNumber int32       `json:"lineno"`
	Default    default_val `json:"default_val"`
}

// sqlstmt exists to match the expected JSON format.
type sqlstmt struct {
	Expr plpgSQL_expr `json:"PLpgSQL_expr"`
}

// statement exists to match the expected JSON format. Unlike other structs, this is used like a union rather than
// having a singular expected implementation.
type statement struct {
	Assignment  *plpgSQL_stmt_assign       `json:"PLpgSQL_stmt_assign"`
	Case        *plpgSQL_stmt_case         `json:"PLpgSQL_stmt_case"`
	DynExec     *plpgSQL_stmt_dynexecute   `json:"PLpgSQL_stmt_dynexecute"`
	ExecSQL     *plpgSQL_stmt_execsql      `json:"PLpgSQL_stmt_execsql"`
	Exit        *plpgSQL_stmt_exit         `json:"PLpgSQL_stmt_exit"`
	ForILoop    *plpgSQL_stmt_fori         `json:"PLpgSQL_stmt_fori"`
	ForSLoop    *plpgSQL_stmt_fors         `json:"PLpgSQL_stmt_fors"`
	If          *plpgSQL_stmt_if           `json:"PLpgSQL_stmt_if"`
	Loop        *plpgSQL_stmt_loop         `json:"PLpgSQL_stmt_loop"`
	Perform     *plpgSQL_stmt_perform      `json:"PLpgSQL_stmt_perform"`
	Raise       *plpgSQL_stmt_raise        `json:"PLpgSQL_stmt_raise"`
	Return      *plpgSQL_stmt_return       `json:"PLpgSQL_stmt_return"`
	ReturnQuery *plpgSQL_stmt_return_query `json:"PLpgSQL_stmt_return_query"`
	When        *plpgSQL_case_when         `json:"PLpgSQL_case_when"`
	While       *plpgSQL_stmt_while        `json:"PLpgSQL_stmt_while"`
}

// Convert converts the JSON statement into its output form.
func (stmt *plpgSQL_stmt_assign) Convert() (Assignment, error) {
	query := stmt.Expression.Expression.Query
	varName := ""
	if equalsIdx := strings.Index(query, ":="); equalsIdx > 0 {
		varName = strings.TrimSpace(query[:equalsIdx])
		query = strings.TrimSpace(query[equalsIdx+2:])
	} else if equalsIdx = strings.Index(query, "="); equalsIdx > 0 {
		varName = strings.TrimSpace(query[:equalsIdx])
		query = strings.TrimSpace(query[equalsIdx+1:])
	} else {
		return Assignment{}, errors.New("PL/pgSQL assignment cannot find `:=` sign")
	}
	return Assignment{
		VariableName:  varName,
		Expression:    query,
		VariableIndex: stmt.VariableNumber,
	}, nil
}

func (stmt *plpgSQL_stmt_case) Convert() (block Block, err error) {
	// If the CASE statement has a main expression, start by assigning it to a variable so
	// we can evaluate it once and only once.
	if stmt.Expression.Expression.Query != "" {
		// TODO: pg_query_go creates the definitions for these variables, and
		//       ideally users shouldn't be able to reference them. We could
		//       update all the references to them (i.e. declaration, assignment,
		//       and WHEN block exprs) to change the name to include a \0 char to
		//       prevent users from referencing them or colliding with them.
		block.Body = append(block.Body, Assignment{
			VariableName: fmt.Sprintf("__Case__Variable_%d__", stmt.VarNo),
			Expression:   stmt.Expression.Expression.Query,
		})
	}

	// Record indexes of all the GOTO ops that jump to the very end of the case block so we
	// can update them later and plug in the correct offsets after we know the final size.
	var gotoEndOpsIndexes []int

	// Add operations for each WHEN statement...
	for _, stmt := range stmt.WhenList {
		when := stmt.When
		if when == nil {
			return Block{}, fmt.Errorf("case statement WHEN clause is nil")
		}

		// TODO: The generated expressions from pg_query_go uses double quotes
		//       around the variable name, which is valid for Postgres, but
		//       our engine doesn't currently resolve double-quoted strings to
		//       variables, so for now, we just extract the double quotes.
		expressionString := when.Expression.Expression.Query
		expressionString = strings.ReplaceAll(expressionString, `"`, "")

		convertedWhenBodyStatements, err := jsonConvertStatements(when.Body)
		if err != nil {
			return Block{}, err
		}

		block.Body = append(block.Body,
			If{
				Condition:  expressionString,
				GotoOffset: 2,
			},
			Goto{
				// This GOTO jumps to the next WHEN block, so step over all the statements
				// from this WHEN block, plus 1 for the GOTO op we add at the end of each
				// block, and plus 1 more to move to the next statement.
				Offset: int32(len(convertedWhenBodyStatements) + 1 + 1),
			})
		block.Body = append(block.Body, convertedWhenBodyStatements...)

		// Add a GOTO op to jump to the end of the entire CASE block, and record its position
		// in the statement block so we can update it later.
		block.Body = append(block.Body, Goto{})
		gotoEndOpsIndexes = append(gotoEndOpsIndexes, len(block.Body)-1)
	}

	if stmt.HasElse {
		convertElseBodyStatements, err := jsonConvertStatements(stmt.Else)
		if err != nil {
			return Block{}, err
		}
		block.Body = append(block.Body, convertElseBodyStatements...)
		// TODO: If no cases match and there is no ELSE block, then add a RAISE statement
		//       to return an error.
		//} else {
		// Sample PostgreSQL error response:
		//	     ERROR:  case not found
		//	     HINT:  CASE statement is missing ELSE part.
		//	     CONTEXT:  PL/pgSQL function interpreted_case(integer) line 5 at CASE
	}

	// Update all the GOTO ops that jump to the very end of the case block.
	for _, gotoEndOpIndex := range gotoEndOpsIndexes {
		// Sanity check that we are looking at a GOTO statement
		if _, ok := block.Body[gotoEndOpIndex].(Goto); !ok {
			return Block{}, fmt.Errorf("expected Goto statement, got %T", block.Body[gotoEndOpIndex])
		}

		block.Body[gotoEndOpIndex] = Goto{
			Offset: int32(len(block.Body) - gotoEndOpIndex),
		}
	}

	return block, nil
}

// Convert converts the JSON statement into its output form.
func (stmt *plpgSQL_stmt_dynexecute) Convert() (DynamicExecute, error) {
	var params []string
	for _, param := range stmt.Params {
		params = append(params, param.Expr.Query)
	}
	var target string
	if stmt.Into {
		switch {
		case stmt.Target.Row != nil:
			names := make([]string, len(stmt.Target.Row.Fields))
			for i, rowField := range stmt.Target.Row.Fields {
				names[i] = rowField.Name
			}
			target = strings.Join(names, ",")
		case stmt.Target.Variable != nil:
			target = stmt.Target.Variable.RefName
		default:
			return DynamicExecute{}, errors.Errorf("unhandled datum type: %T", stmt.Target)
		}
	}
	query := strings.TrimSuffix(strings.TrimPrefix(stmt.Query.Expression.Query, "'"), "'")
	return DynamicExecute{
		Query:  query,
		Params: params,
		Target: target,
	}, nil
}

// Convert converts the JSON statement into its output form.
func (stmt *plpgSQL_stmt_execsql) Convert() (ExecuteSQL, error) {
	var target string
	if stmt.Into {
		switch {
		case stmt.Target.Row != nil:
			names := make([]string, len(stmt.Target.Row.Fields))
			for i, rowField := range stmt.Target.Row.Fields {
				names[i] = rowField.Name
			}
			target = strings.Join(names, ",")
		case stmt.Target.Variable != nil:
			target = stmt.Target.Variable.RefName
		default:
			return ExecuteSQL{}, errors.Errorf("unhandled datum type: %T", stmt.Target)
		}
	}
	return ExecuteSQL{
		Statement: stmt.SQLStmt.Expr.Query,
		Target:    target,
	}, nil
}

// Convert converts the JSON statement into its output form.
func (stmt *plpgSQL_stmt_exit) Convert() Statement {
	offset := int32(-1)
	if stmt.IsExit {
		offset = 1
	}
	var gotoStmt Goto
	if len(stmt.Label) > 0 {
		gotoStmt = Goto{
			Offset: offset,
			Label:  stmt.Label,
		}
	} else {
		gotoStmt = Goto{
			Offset:         offset,
			NearestScopeOp: true,
		}
	}
	if stmt.Condition == nil {
		return gotoStmt
	} else {
		return Block{
			Body: []Statement{
				If{
					Condition:  stmt.Condition.Expression.Query,
					GotoOffset: 2,
				},
				Goto{Offset: 2},
				gotoStmt,
			},
		}
	}
}

// Convert converts the JSON statement into its output form.
func (stmt *plpgSQL_stmt_fori) Convert() (block Block, err error) {
	block.Label = stmt.Label
	block.IsLoop = true

	if stmt.Var.Variable == nil {
		return Block{}, errors.New("for loop variable cannot be nil")
	}
	varName := stmt.Var.Variable.RefName

	// Extract bound and step expressions
	lowerExpr := "1"
	if stmt.Lower != nil {
		lowerExpr = stmt.Lower.Expression.Query
	}
	upperExpr := "1"
	if stmt.Upper != nil {
		upperExpr = stmt.Upper.Expression.Query
	}
	stepExpr := "1"
	if stmt.Step != nil {
		stepExpr = stmt.Step.Expression.Query
	}

	// Determine init value, loop condition, and increment expression based on direction.
	// In the JSON, Lower is always the starting value and Upper is the ending bound.
	var condition, incrExpr string
	if stmt.Reverse {
		condition = fmt.Sprintf("%s >= (%s)", varName, upperExpr)
		incrExpr = fmt.Sprintf("%s - (%s)", varName, stepExpr)
	} else {
		condition = fmt.Sprintf("%s <= (%s)", varName, upperExpr)
		incrExpr = fmt.Sprintf("%s + (%s)", varName, stepExpr)
	}

	// Convert the loop body.
	convertedBody, err := jsonConvertStatements(stmt.Body)
	if err != nil {
		return Block{}, err
	}
	bodySize := OperationSizeForStatements(convertedBody)

	// Build the loop body:
	//   [0]           InitAssign: varName := lower
	//   [1]           If(condition, GotoOffset:2) → jumps to [3] (first body stmt) when true
	//   [2]           ExitGoto → offset=3+bodySize → jumps to ScopeEnd
	//   [3..3+N-1]    body statements (N = bodySize)
	//   [3+N]         IncrAssign: varName := varName +/- step
	//   [3+N+1]       BackGoto → offset=-(3+bodySize) → jumps back to If at [1]
	//
	// Because no variables are declared in this block (the loop variable is already
	// declared by the caller's DECLARE section), ScopeBegin is at M and the
	// InitAssign is at M+1, so all offsets are consistent.
	block.Body = []Statement{
		Assignment{
			VariableName: varName,
			Expression:   lowerExpr,
		},
		If{
			Condition:  condition,
			GotoOffset: 2,
		},
		Goto{
			Offset: 3 + bodySize,
		},
	}
	block.Body = append(block.Body, convertedBody...)
	block.Body = append(block.Body,
		Assignment{
			VariableName: varName,
			Expression:   incrExpr,
		},
		Goto{
			Offset: -(3 + bodySize),
		},
	)
	return block, nil
}

// Convert converts the JSON statement into its output form.
func (stmt *plpgSQL_stmt_fors) Convert() (block Block, err error) {
	block.Label = stmt.Label
	block.IsLoop = true

	if stmt.Query == nil {
		return Block{}, errors.New("FOR..IN..SELECT loop must have a query")
	}

	var varName string
	switch {
	case stmt.Var.Record != nil:
		varName = stmt.Var.Record.RefName
	case stmt.Var.Variable != nil:
		varName = stmt.Var.Variable.RefName
	case stmt.Var.Row != nil:
		varName = stmt.Var.Row.RefName
	default:
		return Block{}, errors.New("FOR..IN..SELECT loop variable must be a record, row, or variable")
	}

	// Use the line number to keep cursor names unique across multiple ForS loops
	// that might use the same variable name.
	cursorName := fmt.Sprintf("__cursor_%s_%d__", varName, stmt.LineNumber)
	query := stmt.Query.Expression.Query

	convertedBody, err := jsonConvertStatements(stmt.Body)
	if err != nil {
		return Block{}, err
	}
	bodySize := OperationSizeForStatements(convertedBody)

	// Layout inside the block (ScopeBegin/ScopeEnd are added by Block.AppendOperations):
	//   [0] ForQueryInit  – execute query, store rows in cursor
	//   [1] ForQueryNext  – fetch next row into varName, or jump forward by (bodySize+2) to ScopeEnd
	//   [2..2+bodySize-1] body statements
	//   [2+bodySize]      Goto back to ForQueryNext: offset = -(1 + bodySize)
	block.Body = []Statement{
		ForQueryInit{CursorName: cursorName, Query: query},
		ForQueryNext{CursorName: cursorName, RecordVar: varName, GotoOffset: bodySize + 2},
	}
	block.Body = append(block.Body, convertedBody...)
	block.Body = append(block.Body, Goto{Offset: -(1 + bodySize)})
	return block, nil
}

// Convert converts the JSON statement into its output form.
func (stmt *plpgSQL_stmt_if) Convert() (Block, error) {
	// We store all GOTOs that will need to go to the end of the block. Since we can't know that ahead of time, we store
	// their indexes and set them at the end of the function.
	type gotoEndIndex struct {
		BodyIndex int
		GotoIndex int32
	}
	var gotoEndIndexes []gotoEndIndex
	returnBlock := Block{
		Body: []Statement{
			If{
				Condition:  stmt.Condition.Expression.Query,
				GotoOffset: 2, // The operation following the conditional skips the THEN statements, so we're skipping that
			},
		},
	}
	// We'll parse our THEN statements, but we won't add them to the block just yet as we need their operation sizes
	thenStmts, err := jsonConvertStatements(stmt.Then)
	if err != nil {
		return Block{}, err
	}
	// When the condition is false, we want to skip our THEN block, so we do that (plus the GOTO which finishes the THEN block)
	returnBlock.Body = append(returnBlock.Body, Goto{Offset: OperationSizeForStatements(thenStmts) + 2})
	// Then we'll append our THEN block
	returnBlock.Body = append(returnBlock.Body, thenStmts...)
	// Then we want to append the GOTO that finishes the THEN block, but we don't know the end just yet, so we'll save
	// its index and fill it in later
	gotoEndIndexes = append(gotoEndIndexes, gotoEndIndex{
		BodyIndex: len(returnBlock.Body),
		GotoIndex: OperationSizeForStatements(returnBlock.Body),
	})
	returnBlock.Body = append(returnBlock.Body, Goto{})
	// We repeat the same process for each ELSIF statement (refer to the comments above)
	for _, elseIf := range stmt.ElseIf {
		returnBlock.Body = append(returnBlock.Body, If{
			Condition:  elseIf.ElseIf.Condition.Expression.Query,
			GotoOffset: 2, // Same rules as skipping our THEN statement above
		})
		elseIfStmts, err := jsonConvertStatements(elseIf.ElseIf.Then)
		if err != nil {
			return Block{}, err
		}
		returnBlock.Body = append(returnBlock.Body, Goto{Offset: OperationSizeForStatements(elseIfStmts) + 2})
		returnBlock.Body = append(returnBlock.Body, elseIfStmts...)
		gotoEndIndexes = append(gotoEndIndexes, gotoEndIndex{
			BodyIndex: len(returnBlock.Body),
			GotoIndex: OperationSizeForStatements(returnBlock.Body),
		})
		returnBlock.Body = append(returnBlock.Body, Goto{})
	}
	// Finally we handle our ELSE statements. We don't have a condition to check, so we don't have to append any
	// additional GOTOs.
	elseStmts, err := jsonConvertStatements(stmt.Else)
	if err != nil {
		return Block{}, err
	}
	returnBlock.Body = append(returnBlock.Body, elseStmts...)
	// Now we'll set all of our GOTOs so that they skip to the end of the block.
	// We have to take their index position into account, since we want to skip to the end from their relative position.
	for _, idx := range gotoEndIndexes {
		returnBlock.Body[idx.BodyIndex] = Goto{Offset: OperationSizeForStatements(returnBlock.Body) - idx.GotoIndex}
	}
	return returnBlock, nil
}

// Convert converts the JSON statement into its output form.
func (stmt *plpgSQL_stmt_loop) Convert() (block Block, err error) {
	// Set the block's label if one was provided
	block.Label = stmt.Label
	block.IsLoop = true
	// Convert the body of the loop first so we can determine the GOTO offset
	block.Body, err = jsonConvertStatements(stmt.Body)
	if err != nil {
		return Block{}, err
	}
	// The loop returns to the beginning of the loop, skipping the body
	block.Body = append(block.Body, Goto{Offset: -OperationSizeForStatements(block.Body)})
	return block, nil
}

// Convert converts the JSON statement into its output form.
func (stmt *plpgSQL_stmt_perform) Convert() Perform {
	return Perform{
		Statement: stmt.Expression.Expression.Query,
	}
}

// Convert converts the JSON statement into its output form.
func (stmt *plpgSQL_stmt_raise) Convert() Raise {
	var params []string
	for _, param := range stmt.Params {
		params = append(params, param.Expr.Query)
	}

	options := make(map[string]string)
	for _, option := range stmt.Options {
		options[strconv.Itoa(int(option.Option.OptionType))] = option.Option.Expression.Expr.Query
	}

	return Raise{
		Level:   NoticeLevel(uint8(stmt.ELogLevel)).String(),
		Message: stmt.Message,
		Params:  params,
		Options: options,
	}
}

// Convert converts the JSON statement into its output form.
func (stmt *plpgSQL_stmt_return) Convert() Return {
	return Return{
		Expression: stmt.Expression.Expression.Query,
	}
}

// Convert converts the JSON statement into its output form.
func (stmt *plpgSQL_stmt_return_query) Convert() ReturnQuery {
	return ReturnQuery{
		Query: stmt.Query.Expression.Query,
	}
}

// Convert converts the JSON statement into its output form.
func (stmt *plpgSQL_stmt_while) Convert() (block Block, err error) {
	// Convert the body of the loop first so we can determine the GOTO offsets
	convertedLoopBodyStmts, err := jsonConvertStatements(stmt.Body)
	if err != nil {
		return Block{}, err
	}

	block = Block{
		Body: []Statement{
			If{
				Condition: stmt.Condition.Expression.Query,
				// Jump forward two statements, so we skip over the GOTO below that exits the WHILE loop.
				GotoOffset: 2,
			},
			Goto{
				// Jump forward 1 statement to get to the loop body, then jump over the loop body and the
				// GOTO statement that jumps to the start of the WHILE loop.
				Offset: 1 + OperationSizeForStatements(convertedLoopBodyStmts) + 1,
			},
		},
		Label:  stmt.Label,
		IsLoop: true,
	}

	// Add the converted body of the WHILE loop, and a GOTO statement that jumps backwards past the current
	// GOTO statement, and past all the body statements, and past the GOTO statement at the start of the loop.
	block.Body = append(block.Body, convertedLoopBodyStmts...)
	block.Body = append(block.Body, Goto{Offset: -1 * (OperationSizeForStatements(convertedLoopBodyStmts) + 2)})
	return block, nil
}
