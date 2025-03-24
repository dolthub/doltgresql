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

package output

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/server/ast"
)

// QueryParses determines whether a query parses, and then whether it has an AST conversion.
type QueryParses interface {
	// ShouldParse returns whether the query successfully creates a Postgres AST.
	ShouldParse() bool
	// ShouldConvert returns whether the query successfully converts from a Postgres AST to a Vitess AST.
	ShouldConvert() bool
	// String returns the query to test.
	String() string
}

// Unimplemented is used when a query has not yet been implemented in the parser.
type Unimplemented string

var _ QueryParses = Unimplemented("")

// ShouldParse implements the interface QueryParses.
func (Unimplemented) ShouldParse() bool {
	return false
}

// ShouldConvert implements the interface QueryParses.
func (Unimplemented) ShouldConvert() bool {
	return false
}

// String implements the interface QueryParses.
func (u Unimplemented) String() string {
	return string(u)
}

// Parses is used when a query parses into a Postgres AST, but cannot yet convert to a Vitess AST.
type Parses string

var _ QueryParses = Parses("")

// ShouldParse implements the interface QueryParses.
func (Parses) ShouldParse() bool {
	return true
}

// ShouldConvert implements the interface QueryParses.
func (Parses) ShouldConvert() bool {
	return false
}

// String implements the interface QueryParses.
func (p Parses) String() string {
	return string(p)
}

// Converts is used when a query parses into a Postgres AST and converts to a Vitess AST.
type Converts string

var _ QueryParses = Converts("")

// ShouldParse implements the interface QueryParses.
func (Converts) ShouldParse() bool {
	return true
}

// ShouldConvert implements the interface QueryParses.
func (Converts) ShouldConvert() bool {
	return true
}

// String implements the interface QueryParses.
func (c Converts) String() string {
	return string(c)
}

// RunTests runs the given collection of QueryParses tests.

func RunTests(t *testing.T, tests []QueryParses) {
	for _, test := range tests {
		t.Run(test.String(), func(t *testing.T) {
			statements, err := parser.Parse(test.String())
			if !test.ShouldParse() {
				if err == nil && len(statements) > 0 {
					t.Fatal("Query now parses, please upgrade the type to `Parses`")
				}
				return
			}
			require.NoError(t, err, "Regression, query previously parsed")
			require.Truef(t, len(statements) > 0, "Regression, query previously produced a Postgres AST")
			for _, statement := range statements {
				vitessAST, err := func() (vitessAST sqlparser.Statement, err error) {
					defer func() {
						if recoverVal := recover(); recoverVal != nil {
							vitessAST = nil
						}
					}()
					return ast.Convert(statement)
				}()
				if !test.ShouldConvert() {
					if err == nil && vitessAST != nil {
						t.Fatalf("Query %s now converts, please upgrade the type to `Converts`", test.String())
					}
					return
				}
				assert.NoError(t, err, "Regression, query previously converted from a Postgres AST to a Vitess AST")
				assert.NotNil(t, vitessAST, "Regression, query now returns a nil Vitess AST")
			}
		})
	}
}

// Regex which checks for the line beginning with Unimplemented, Parses, or Converts
var testStatementRegex = regexp.MustCompile(`^(\s*)(Unimplemented|Parses|Converts)\(`)

// RewriteTests rewrites the given test file with the given tests to have the new results. This is a utility to avoid
// having to tediously update tests after implementing some parser / engine functionality.
func RewriteTests(t *testing.T, tests []QueryParses, file string) {
	newPath := file + ".new"
	f, err := os.Open(file)
	require.NoError(t, err)
	fNew, err := os.Create(newPath)
	require.NoError(t, err)

	abs, err := filepath.Abs(newPath)
	require.NoError(t, err)
	fmt.Println("Rewriting tests to", abs)

	bufferedReader := bufio.NewReader(f)
	bufferedWriter := bufio.NewWriter(fNew)

	// copy the source file until we find the first test statement
	var line []byte
	for {
		line, _, err = bufferedReader.ReadLine()
		if err == io.EOF {
			break
		} else {
			require.NoError(t, err)
		}

		if testStatementRegex.Match(line) {
			break
		}

		_, err = bufferedWriter.Write(line)
		require.NoError(t, err)
		_, err = bufferedWriter.Write([]byte("\n"))
		require.NoError(t, err)
	}

	for _, test := range tests {
		t.Run(test.String(), func(t *testing.T) {
			statements, err := parser.Parse(test.String())
			if !test.ShouldParse() {
				if err == nil && len(statements) > 0 {
					line = testStatementRegex.ReplaceAll(line, []byte("${1}Parses("))
				}
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

				if !test.ShouldConvert() {
					if err == nil && vitessAST != nil {
						line = testStatementRegex.ReplaceAll(line, []byte("${1}Converts("))
					}
				}
			}

			_, err = bufferedWriter.Write(line)
			require.NoError(t, err)
			_, err = bufferedWriter.Write([]byte("\n"))
			require.NoError(t, err)

			line, _, err = bufferedReader.ReadLine()
			require.NoError(t, err)
		})
	}

	// copy the rest of the file
	for {
		_, err = bufferedWriter.Write(line)
		require.NoError(t, err)
		_, err = bufferedWriter.Write([]byte("\n"))
		require.NoError(t, err)

		line, _, err = bufferedReader.ReadLine()
		if err == io.EOF {
			break
		} else {
			require.NoError(t, err)
		}
	}

	bufferedWriter.Flush()
	f.Close()
	fNew.Close()

	// move the new file to the original file
	err = os.Rename(newPath, file)
	require.NoError(t, err)
}
