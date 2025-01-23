// Copyright 2023-2024 Dolthub, Inc.
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

package utils

import "github.com/cockroachdb/errors"

// ParseTokens parses the given tokens into a StatementGenerator.
func ParseTokens(tokens []Token, includeRepetition bool) (StatementGenerator, error) {
	stack := NewStatementGeneratorStack()
	var statements []StatementGenerator
	variables := make(map[string]StatementGenerator)
	currentVariable := ""
	tokenReader := NewTokenReader(tokens)
ForLoop:
	for {
		token, ok := tokenReader.Next()
		if !ok {
			break ForLoop
		}
		switch token.Type {
		case TokenType_Text:
			stack.AddText(token.Literal)
		case TokenType_Variable:
			stack.AddVariable(token.Literal)
		case TokenType_VariableDefinition:
			currentVariable = token.Literal
			if token, _ = tokenReader.Next(); token.Type != TokenType_LongSpace {
				return nil, errors.Errorf("expected a long space after a variable definition declaration")
			}
		case TokenType_Or:
			if err := stack.Or(); err != nil {
				return nil, err
			}
		case TokenType_Repeat:
			if includeRepetition {
				if err := stack.Repeat(); err != nil {
					return nil, err
				}
			}
		case TokenType_OptionalRepeat:
			if includeRepetition {
				if err := stack.OptionalRepeat(token.Literal); err != nil {
					return nil, err
				}
			}
		case TokenType_ShortSpace, TokenType_MediumSpace:
			return nil, errors.Errorf("token reader should have removed all short and medium spaces")
		case TokenType_LongSpace, TokenType_EOF:
			newStatement, err := stack.Finish()
			if err != nil {
				return nil, err
			}
			if newStatement == nil {
				return nil, errors.Errorf("long space encountered before writing to the stack")
			}
			if len(currentVariable) > 0 {
				if _, ok = variables[currentVariable]; ok {
					return nil, errors.Errorf("multiple definitions for the same variable: %s", currentVariable)
				}
				variables[currentVariable] = newStatement
				currentVariable = ""
			} else {
				statements = append(statements, newStatement)
			}
			if token.Type == TokenType_EOF {
				break ForLoop
			} else {
				stack = NewStatementGeneratorStack()
			}
		case TokenType_ParenOpen:
			stack.NewParenScope()
		case TokenType_ParenClose:
			if err := stack.ExitParenScope(); err != nil {
				return nil, err
			}
		case TokenType_OptionalOpen:
			stack.NewOptionalScope()
		case TokenType_OptionalClose:
			if err := stack.ExitOptionalScope(); err != nil {
				return nil, err
			}
		case TokenType_OneOfOpen:
			stack.NewScope()
		case TokenType_OneOfClose:
			if err := stack.ExitScope(); err != nil {
				return nil, err
			}
		default:
			panic("unhandled token type")
		}
	}
	finalStackContents, err := stack.Finish()
	if err != nil {
		return nil, err
	}
	if finalStackContents != nil {
		return nil, errors.Errorf("encountered an early EOF, as the stack was still processing")
	}
	if len(statements) == 0 {
		return nil, errors.Errorf("no statements were generated from the token stream")
	}
	var finalStatementGenerator StatementGenerator
	if len(statements) == 1 {
		finalStatementGenerator = statements[0]
	} else {
		finalStatementGenerator = Or(statements...)
	}
	if err = ApplyVariableDefinition(finalStatementGenerator, variables); err != nil {
		return nil, err
	}
	return finalStatementGenerator, nil
}
