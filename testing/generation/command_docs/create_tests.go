// Copyright 2023 Dolthub, Inc.
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

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/server/ast"
)

const TestHeader = `// Copyright %d Dolthub, Inc.
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

package output

import "testing"

func Test%s(t *testing.T) {
	tests := []QueryParses{
`

const TestFooter = `	}
	RunTests(t, tests)
}
`

// GenerateTestsFromSynopses generates a test file in the output directory for each file in the synopses directory.
func GenerateTestsFromSynopses() (err error) {
	parentFolder, err := GetCommandDocsFolder()
	if err != nil {
		return err
	}
	fileInfos, err := os.ReadDir(fmt.Sprintf("%s/synopses", parentFolder))
	if err != nil {
		return err
	}

FileLoop:
	for i, fileInfo := range fileInfos {
		if i != 0 {
			//TODO: this runs a single file to prevent writing all of the files, since some are unbelievably large
			continue FileLoop
		}
		prefix := strings.ToUpper(
			strings.ReplaceAll(
				strings.ReplaceAll(
					fileInfo.Name(),
					".txt",
					"",
				),
				"_",
				" ",
			),
		)
		fmt.Println(SectionMarker(prefix, '+', 80))
		data, nErr := os.ReadFile(fmt.Sprintf("%s/synopses/%s", parentFolder, fileInfo.Name()))
		if nErr != nil {
			err = errors.Join(err, nErr)
			continue FileLoop
		}
		dataStr := strings.TrimSpace(string(data))
		scanner := NewScanner(dataStr)
		tokens, nErr := scanner.Process()
		if nErr != nil {
			err = errors.Join(err, nErr)
			continue FileLoop
		}
		scannerString := scanner.String()
		if dataStr != scannerString {
			sb := strings.Builder{}
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(dataStr, scannerString, true)
			whitespaceOnly := true
			for _, diff := range diffs {
				if diff.Type != diffmatchpatch.DiffEqual && diff.Text != " " && diff.Text != "\n" {
					whitespaceOnly = false
				}
			}
			if whitespaceOnly {
				sb.WriteString(SectionMarker("Whitespace Differences", '-', 80))
			} else {
				sb.WriteString(dmp.DiffPrettyText(diffs))
			}
			err = errors.Join(err, errors.New(sb.String()))
			continue FileLoop
		}
		stmtGen, nErr := ParseTokens(tokens)
		if nErr != nil {
			err = errors.Join(err, nErr)
			continue FileLoop
		}
		sb := strings.Builder{}
		sb.WriteString(fmt.Sprintf(TestHeader, time.Now().Year(), strings.ReplaceAll(strings.Title(strings.ToLower(prefix)), " ", "")))

		result, nErr := GetQueryResult(stmtGen.String())
		if nErr != nil {
			err = errors.Join(err, nErr)
			continue FileLoop
		}
		sb.WriteString(result)
		for stmtGen.Consume() {
			result, nErr = GetQueryResult(stmtGen.String())
			if nErr != nil {
				err = errors.Join(err, nErr)
				continue FileLoop
			}
			sb.WriteString(result)
		}

		sb.WriteString(TestFooter)
		outputFileName := strings.ToLower(strings.ReplaceAll(prefix, " ", "_"))
		if nErr = os.WriteFile(fmt.Sprintf("%s/output/%s_test.go", parentFolder, outputFileName), []byte(sb.String()), 0644); nErr != nil {
			err = errors.Join(err, nErr)
			continue FileLoop
		}
	}
	return err
}

// ParseTokens parses the given tokens into a StatementGenerator.
func ParseTokens(tokens []Token) (StatementGenerator, error) {
	stack := NewStatementGeneratorStack()
	var statements []StatementGenerator
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
			//TODO: implement variable definitions
			break ForLoop
		case TokenType_Or:
			if err := stack.Or(); err != nil {
				return nil, err
			}
		case TokenType_Repeat:
			if err := stack.Repeat(false); err != nil {
				return nil, err
			}
		case TokenType_CommaRepeat:
			if err := stack.Repeat(true); err != nil {
				return nil, err
			}
		case TokenType_OptionalRepeat:
			if err := stack.OptionalRepeat(false); err != nil {
				return nil, err
			}
		case TokenType_OptionalCommaRepeat:
			if err := stack.OptionalRepeat(true); err != nil {
				return nil, err
			}
		case TokenType_ShortSpace, TokenType_MediumSpace:
			return nil, fmt.Errorf("token reader should have removed all short and medium spaces")
		case TokenType_LongSpace, TokenType_EOF:
			newStatement, err := stack.Finish()
			if err != nil {
				return nil, err
			}
			if newStatement == nil {
				return nil, fmt.Errorf("long space encountered before writing to the stack")
			}
			statements = append(statements, newStatement)
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
		return nil, fmt.Errorf("encountered an early EOF, as the stack was still processing")
	}
	if len(statements) == 0 {
		return nil, fmt.Errorf("no statements were generated from the token stream")
	} else if len(statements) == 1 {
		return statements[0], nil
	} else {
		return Or(statements...), nil
	}
}

// GetQueryResult runs the query against a Postgres server to validate that the query is syntactically valid. It then
// tests the query against the Postgres parser and Postgres-Vitess AST converter to check the current level of support.
// It returns a string that may be inserted directly into a test source file (two tabs are prefixed).
func GetQueryResult(query string) (string, error) {
	//TODO: verify the query against a Postgres server
	formattedQuery := strings.ReplaceAll(query, `"`, `\"`)
	statements, err := parser.Parse(query)
	if err != nil || len(statements) == 0 {
		return fmt.Sprintf("\t\tUnimplemented(\"%s\"),\n", formattedQuery), nil
	}
	for _, statement := range statements {
		vitessAST, err := ast.Convert(statement)
		if err != nil || vitessAST == nil {
			return fmt.Sprintf("\t\tParses(\"%s\"),\n", formattedQuery), nil
		}
	}
	return fmt.Sprintf("\t\tConverts(\"%s\"),\n", formattedQuery), nil
}

// SectionMarker returns a marker that may be used to denote sections.
//
// For example, SectionMarker("abc", '-', 21) would return:
//
// -------- abc --------
func SectionMarker(centeredText string, fillerCharacter rune, totalLength int) string {
	fillerStr := string(fillerCharacter)
	remainingLength := totalLength - (len(centeredText) + 2)
	if remainingLength <= 0 {
		return fmt.Sprintf(" %s ", centeredText)
	}
	left := remainingLength / 2
	right := remainingLength - left // Integer division doesn't do fractions, so this will handle odd counts
	return fmt.Sprintf("%s %s %s",
		strings.Repeat(fillerStr, left), centeredText, strings.Repeat(fillerStr, right))
}
