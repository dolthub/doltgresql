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
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"

	"github.com/dolthub/doltgresql/testing/generation/utils"
	framework "github.com/dolthub/doltgresql/testing/go"
)

// GenerateRowTests uses the given StatementGenerator to return a random set of queries along with their results that
// were retrieved from a Postgres instance. Uses the MaxTestCount to determine the number of tests to generate. The
// returned map uses the query as the key, with the results being the value.
func GenerateRowTests(stmtGen utils.StatementGenerator) (map[string][]sql.Row, error) {
	randomInts, err := utils.GenerateRandomInts(MaxTestCount, stmtGen.Permutations())
	if err != nil {
		return nil, err
	}
	allResults := make(map[string][]sql.Row)
	for _, randomInt := range randomInts {
		stmtGen.SetConsumeIterations(randomInt)
		generatedString := stmtGen.String()
		result, err := GetRowResults(generatedString)
		if err != nil {
			return nil, err
		}
		allResults[generatedString] = result
	}
	return allResults, nil
}

// GetSynopsisStatementGenerator returns the StatementGenerator for the given synopsis. The synopsisData should be the
// string that has been loaded from the synopses directory. includeRepetition states whether repetition is included in
// the returned StatementGenerator.
func GetSynopsisStatementGenerator(synopsisData string, prefix string, includeRepetition bool) (utils.StatementGenerator, error) {
	scanner := NewScanner(synopsisData)
	tokens, err := scanner.Process()
	if err != nil {
		return nil, err
	}
	stmtGen, err := utils.ParseTokens(tokens, includeRepetition)
	if err != nil {
		return nil, err
	}
	// Not all variables have their definitions set in the synopsis, so we'll handle them here
	unsetVariables, err := utils.UnsetVariables(stmtGen)
	if err != nil {
		return nil, err
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
	if err = utils.ApplyVariableDefinition(stmtGen, customVariableDefinitions); err != nil {
		return nil, err
	}
	return stmtGen, nil
}

// LoadSynopsis loads the given synopsis from the synopses directory. The synopsis name may be the file name, or the
// name of the statement (also known as the prefix). Case-insensitive.
func LoadSynopsis(synopsis string) (data string, prefix string, err error) {
	prefix = strings.ToUpper(
		strings.ReplaceAll(
			strings.ReplaceAll(
				synopsis,
				".txt",
				"",
			),
			"_",
			" ",
		),
	)
	fileName := strings.ToLower(strings.ReplaceAll(prefix+".txt", " ", "_"))
	parentFolder, err := GetCommandDocsFolder()
	if err != nil {
		return "", "", err
	}
	dataBytes, err := parentFolder.ReadFileFromDirectory("synopses", fileName)
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(string(dataBytes)), prefix, nil
}

// GetRowResults runs the query against a Postgres server and returns the resulting rows.
func GetRowResults(query string) ([]sql.Row, error) {
	var err error
	ctx := context.Background()
	if postgresVerificationConnection == nil {
		connectionString := fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/", 5432)
		postgresVerificationConnection, err = pgx.Connect(ctx, connectionString)
		if err != nil {
			return nil, err
		}
	}
	pgxRows, err := postgresVerificationConnection.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	return framework.ReadRows(pgxRows, true)
}
