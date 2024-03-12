// Copyright 2024 Dolthub, Inc.
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
	"fmt"
	"sort"

	"github.com/dolthub/go-mysql-server/sql"

	_ "github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/testing/generation/utils"
)

// Assertion represents a single SELECT assertion.
type Assertion struct {
	Stmt  string
	Rows  []sql.Row
	Error bool
}

func main() {
	framework.Initialize()
	rootFolder, err := utils.GetRootFolder()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}
	rootFolder = rootFolder.MoveRoot("testing/generation/function_coverage")

	// Sort the strings for determinism, since range iteration on maps is non-deterministic
	functionNames := make([]string, 0, len(framework.Catalog))
	for functionName := range framework.Catalog {
		functionNames = append(functionNames, functionName)
	}
	sort.Strings(functionNames)

	for _, functionName := range functionNames {
		var assertions []Assertion

		utils.PrintSectionMarker(functionName, '+', 80)

	FunctionLoop:
		for _, function := range framework.Catalog[functionName] {
			var literalGeneratorParams []utils.StatementGenerator
			literalGeneratorParams = append(literalGeneratorParams, utils.Text(functionName+"("))
			for i, paramType := range function.GetParameters() {
				if i > 0 {
					literalGeneratorParams = append(literalGeneratorParams, utils.Text(", "))
				}
				if generator, ok := valueMappings[paramType.BaseID()]; ok {
					literalGeneratorParams = append(literalGeneratorParams, generator)
				} else {
					fmt.Printf("missing support for functions with the parameter type: `%s`\n", paramType.String())
					continue FunctionLoop
				}
			}
			literalGeneratorParams = append(literalGeneratorParams, utils.Text(")"))
			literalGenerator := utils.Collection(
				utils.Text("SELECT"),
				utils.Collection(literalGeneratorParams...),
				utils.Text(";"),
			)

			randomInts, err := utils.GenerateRandomInts(500, literalGenerator.Permutations())
			if err != nil {
				fmt.Println(err.Error())
				continue FunctionLoop
			}
			for _, randomInt := range randomInts {
				literalGenerator.SetConsumeIterations(randomInt)
				assertion := Assertion{
					Stmt:  literalGenerator.String(),
					Error: false,
				}
				result, _, err := QueryPostgres(assertion.Stmt)
				if err != nil {
					assertion.Error = true
				} else {
					// Some tests may generate enormous strings, so we'll skip those
					roughRowTotal := uint64(0)
					for _, v1 := range result {
						for _, v2 := range v1 {
							if stringVal, ok := v2.(string); ok {
								roughRowTotal += uint64(len(stringVal))
							}
						}
					}
					if roughRowTotal > 300 {
						continue
					}
					assertion.Rows = result
				}
				assertions = append(assertions, assertion)
			}
		}

		if len(assertions) > 0 {
			if err := GenerateTests(rootFolder, functionName, assertions); err != nil {
				fmt.Printf("%s\n", err.Error())
				return
			}
		}
	}
}
