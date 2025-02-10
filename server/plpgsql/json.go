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

// elsif exists to match the expected JSON format.
type elsif struct {
	ElseIf plpgSQL_if_elsif `json:"PLpgSQL_if_elsif"`
}

// expr exists to match the expected JSON format.
type expr struct {
	Expression plpgSQL_expr `json:"PLpgSQL_expr"`
}

// datum exists to match the expected JSON format.
type datum struct {
	Variable plpgSQL_var `json:"PLpgSQL_var"`
}

// function exists to match the expected JSON format.
type function struct {
	Function plpgSQL_block `json:"PLpgSQL_function"`
}

// plpgSQL_block exists to match the expected JSON format.
type plpgSQL_block struct {
	Datums []datum `json:"datums"`
	Action action  `json:"action"`
}

// plpgSQL_expr exists to match the expected JSON format.
type plpgSQL_expr struct {
	Query     string `json:"query"`
	ParseMode int32  `json:"parseMode"`
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
	LineNumber int32       `json:"lineno"`
}

// plpgSQL_if_elsif exists to match the expected JSON format.
type plpgSQL_if_elsif struct {
	Condition  cond        `json:"cond"`
	Then       []statement `json:"stmts"`
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

// plpgSQL_stmt_return exists to match the expected JSON format.
type plpgSQL_stmt_return struct {
	Expression expr  `json:"expr"`
	LineNumber int32 `json:"lineno"`
}

// plpgSQL_type exists to match the expected JSON format.
type plpgSQL_type struct {
	Name string `json:"typname"`
}

// plpgSQL_var exists to match the expected JSON format.
type plpgSQL_var struct {
	RefName    string   `json:"refname"`
	Type       datatype `json:"datatype"`
	LineNumber int32    `json:"lineno"`
}

// statement exists to match the expected JSON format. Unlike other structs, this is used like a union rather than
// having a singular expected implementation.
type statement struct {
	Assignment *plpgSQL_stmt_assign `json:"PLpgSQL_stmt_assign"`
	If         *plpgSQL_stmt_if     `json:"PLpgSQL_stmt_if"`
	Return     *plpgSQL_stmt_return `json:"PLpgSQL_stmt_return"`
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

// Convert converts the JSON statement into its output form.
func (stmt *plpgSQL_stmt_if) Convert() (Block, error) {
	// We store all GOTOs that will need to go to the end of the block. Since we can't know that ahead of time, we store
	// their indexes and set them at the end of the function.
	var gotoEndIndexes []int32
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
	gotoEndIndexes = append(gotoEndIndexes, int32(len(returnBlock.Body)))
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
		gotoEndIndexes = append(gotoEndIndexes, int32(len(returnBlock.Body)))
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
	for _, gotoEndIndex := range gotoEndIndexes {
		returnBlock.Body[gotoEndIndex] = Goto{Offset: int32(len(returnBlock.Body)) - gotoEndIndex}
	}
	return returnBlock, nil
}

// Convert converts the JSON statement into its output form.
func (stmt *plpgSQL_stmt_return) Convert() Return {
	return Return{
		Expression: stmt.Expression.Expression.Query,
	}
}
