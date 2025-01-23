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
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

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

import (
	"testing"
	"github.com/dolthub/go-mysql-server/sql"
)

func Test_%s(t *testing.T) {
	RunScripts(t, []ScriptTest{
`

const TestFooter = `	})
}
`

// GenerateTests generates and writes a test file for the given function.
func GenerateTests(parentFolder utils.RootFolderLocation, functionName string, assertions []Assertion) error {
	// We want to skip some functions since we know they'll always return a different result
	switch functionName {
	case "random":
		return nil
	}
	// For now we'll skip some functions that will take a long time to calculate
	switch functionName {
	case "exp", "factorial", "lpad", "power", "rpad", "repeat", "round":
		return nil
	}

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf(TestHeader, time.Now().Year(),
		strings.ReplaceAll(cases.Title(language.English).String(strings.ReplaceAll(functionName, "_", " ")), " ", "")))
	sb.WriteString("\t\t{\n")
	sb.WriteString(fmt.Sprintf("\t\t\tName: \"%s\",\n", functionName))
	sb.WriteString("\t\t\tAssertions: []ScriptTestAssertion{\n")
	for _, assertion := range assertions {
		sb.WriteString("\t\t\t\t{\n")
		if !assertion.Error {
			sb.WriteString(fmt.Sprintf("\t\t\t\t\tQuery:    \"%s\",\n", assertion.Stmt))
			sb.WriteString("\t\t\t\t\tExpected: []sql.Row{")
			for i, results := range assertion.Rows {
				if i > 0 {
					sb.WriteRune(',')
				}
				sb.WriteRune('{')
				for j, result := range results {
					if j > 0 {
						sb.WriteRune(',')
					}
					switch result := result.(type) {
					case int:
						sb.WriteString(fmt.Sprintf("int64(%d)", result))
					case int8:
						sb.WriteString(fmt.Sprintf("int8(%d)", result))
					case int16:
						sb.WriteString(fmt.Sprintf("int16(%d)", result))
					case int32:
						sb.WriteString(fmt.Sprintf("int32(%d)", result))
					case int64:
						sb.WriteString(fmt.Sprintf("int64(%d)", result))
					case uint:
						sb.WriteString(fmt.Sprintf("uint64(%d)", result))
					case uint8:
						sb.WriteString(fmt.Sprintf("uint8(%d)", result))
					case uint16:
						sb.WriteString(fmt.Sprintf("uint16(%d)", result))
					case uint32:
						sb.WriteString(fmt.Sprintf("uint32(%d)", result))
					case uint64:
						sb.WriteString(fmt.Sprintf("uint64(%d)", result))
					case float32:
						sb.WriteString(fmt.Sprintf("float32(%f)", result))
					case float64:
						sb.WriteString(fmt.Sprintf("float64(%f)", result))
					case string:
						sb.WriteString(fmt.Sprintf("%q", result))
					case pgtype.Numeric:
						sb.WriteString(fmt.Sprintf(`Numeric("%s")`, NumericToString(result)))
					case nil:
						sb.WriteString("nil")
					default:
						return errors.Errorf("%T does not have a switch case", result)
					}
				}
				sb.WriteRune('}')
			}
			sb.WriteString("},\n")
		} else {
			sb.WriteString(fmt.Sprintf("\t\t\t\t\tQuery:       \"%s\",\n", assertion.Stmt))
			sb.WriteString("\t\t\t\t\tExpectedErr: true,\n")
		}
		sb.WriteString("\t\t\t\t},\n")
	}
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t},\n")
	sb.WriteString(TestFooter)
	outputFileName := strings.ToLower(strings.ReplaceAll(functionName, " ", "_"))
	return parentFolder.WriteFileToDirectory("output", outputFileName+"_test.go", []byte(sb.String()), 0644)
}

// NumericToString converts a numeric to a string.
func NumericToString(numeric pgtype.Numeric) string {
	str, err := numeric.Value()
	if err != nil {
		panic(err)
	}
	return str.(string)
}
