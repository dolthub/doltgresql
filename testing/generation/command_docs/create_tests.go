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
	"context"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5"
	"github.com/sergi/go-diff/diffmatchpatch"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/server/ast"
	"github.com/dolthub/doltgresql/testing/generation/utils"
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

const MaxTestCount = 10000

// GenerateTestsFromSynopses generates a test file in the output directory for each file in the synopses directory.
func GenerateTestsFromSynopses(repetitionDisabled ...string) (err error) {
	parentFolder, err := GetCommandDocsFolder()
	if err != nil {
		return err
	}
	fileInfos, err := parentFolder.ReadDir("synopses")
	if err != nil {
		return err
	}
	removeComments := regexp.MustCompile(`\/\/[^\r\n]*\r?\n?`)

FileLoop:
	for _, fileInfo := range fileInfos {
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
		utils.PrintSectionMarker(prefix, '+', 80)
		data, nErr := parentFolder.ReadFileFromDirectory("synopses", fileInfo.Name())
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
		if removeComments.ReplaceAllString(dataStr, "") != scannerString {
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
				sb.WriteString(utils.SectionMarker("Whitespace Differences", '-', 80))
			} else {
				sb.WriteString(dmp.DiffPrettyText(diffs))
			}
			fmt.Println(sb.String())
			err = errors.Join(err, errors.New(sb.String()))
			continue FileLoop
		}
		includeRepetition := len(repetitionDisabled) == 0 || repetitionDisabled[0] != "*"
		for _, bans := range repetitionDisabled {
			if strings.EqualFold(bans, prefix) {
				includeRepetition = false
				break
			}
		}
		stmtGen, nErr := utils.ParseTokens(tokens, includeRepetition)
		if nErr != nil {
			err = errors.Join(err, nErr)
			continue FileLoop
		}
		// Not all variables have their definitions set in the synopsis, so we'll handle them here
		unsetVariables, nErr := utils.UnsetVariables(stmtGen)
		if nErr != nil {
			err = errors.Join(err, nErr)
			continue FileLoop
		}
		customVariableDefinitions := make(map[string]utils.StatementGenerator)
		for _, unsetVariable := range unsetVariables {
			// Check for a specific definition first
			if prefixVariables, ok := PrefixCustomVariables[prefix]; ok {
				if variableDefinition, ok := prefixVariables[unsetVariable]; ok {
					customVariableDefinitions[unsetVariable] = variableDefinition
					continue
				}
			}
			// Check the global definitions if there isn't a specific definition
			if variableDefinition, ok := GlobalCustomVariables[unsetVariable]; ok {
				customVariableDefinitions[unsetVariable] = variableDefinition
				continue
			}
		}
		if nErr = utils.ApplyVariableDefinition(stmtGen, customVariableDefinitions); nErr != nil {
			err = errors.Join(err, nErr)
			continue FileLoop
		}
		sb := strings.Builder{}
		sb.WriteString(fmt.Sprintf(TestHeader, time.Now().Year(),
			strings.ReplaceAll(cases.Title(language.English).String(strings.ToLower(prefix)), " ", "")))

		permutations := stmtGen.Permutations()
		if permutations.Cmp(big.NewInt(MaxTestCount)) <= 0 {
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
		} else {
			randomInts, nErr := utils.GenerateRandomInts(MaxTestCount, permutations)
			if nErr != nil {
				err = errors.Join(err, nErr)
			}
			for _, randomInt := range randomInts {
				stmtGen.SetConsumeIterations(randomInt)
				result, nErr := GetQueryResult(stmtGen.String())
				if nErr != nil {
					err = errors.Join(err, nErr)
					continue FileLoop
				}
				sb.WriteString(result)
			}
		}

		sb.WriteString(TestFooter)
		outputFileName := strings.ToLower(strings.ReplaceAll(prefix, " ", "_"))
		if nErr = parentFolder.WriteFileToDirectory("output", outputFileName+"_test.go", []byte(sb.String()), 0644); nErr != nil {
			err = errors.Join(err, nErr)
			continue FileLoop
		}
	}
	return err
}

var postgresVerificationConnection *pgx.Conn

// GetQueryResult runs the query against a Postgres server to validate that the query is syntactically valid. It then
// tests the query against the Postgres parser and Postgres-Vitess AST converter to check the current level of support.
// It returns a string that may be inserted directly into a test source file (two tabs are prefixed).
func GetQueryResult(query string) (string, error) {
	var err error
	ctx := context.Background()
	if postgresVerificationConnection == nil {
		connectionString := fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/", 5432)
		postgresVerificationConnection, err = pgx.Connect(ctx, connectionString)
		if err != nil {
			return "", err
		}
	}
	testQuery := fmt.Sprintf("DO $SYNTAX_CHECK$ BEGIN RETURN; %s; END; $SYNTAX_CHECK$;", query)
	_, err = postgresVerificationConnection.Exec(ctx, testQuery)
	if err != nil && strings.Contains(err.Error(), "syntax error") {
		// We only care about syntax errors, as statements may rely on internal state, which is not what we're testing
		// There are statements that will not execute inside our DO block due to how Postgres handles some queries, so
		// to confirm that they're syntax errors, we'll run them outside the block. All such queries should be
		// non-destructive, so this should be safe. All other queries will still return a syntax error.
		_, err = postgresVerificationConnection.Exec(ctx, query)
		// Run a ROLLBACK as some commands may put the connection (not the database) in a bad state
		_, _ = postgresVerificationConnection.Exec(ctx, "ROLLBACK;")
		if err != nil && strings.Contains(err.Error(), "syntax error") {
			return "", fmt.Errorf("%s\n%s", err, query)
		}
	}
	formattedQuery := strings.ReplaceAll(query, `"`, `\"`)
	statements, err := parser.Parse(query)
	if err != nil || len(statements) == 0 {
		return fmt.Sprintf("\t\tUnimplemented(\"%s\"),\n", formattedQuery), nil
	}
	for _, statement := range statements {
		vitessAST, err := func() (vitessAST sqlparser.Statement, err error) {
			defer func() {
				if recoverVal := recover(); recoverVal != nil {
					vitessAST = nil
				}
			}()
			return ast.Convert(statement)
		}()
		if err != nil || vitessAST == nil {
			return fmt.Sprintf("\t\tParses(\"%s\"),\n", formattedQuery), nil
		}
	}
	return fmt.Sprintf("\t\tConverts(\"%s\"),\n", formattedQuery), nil
}
