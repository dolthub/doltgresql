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
	"os"
	"regexp"
	"strconv"
	"strings"
)

// TestFile contains the lines of a test file, along with the current line being read.
type TestFile struct {
	lines []string
	line  int
	isEOF bool
}

var lineRegex, _ = regexp.Compile(`^LINE \d+:`)
var linePointerRegex, _ = regexp.Compile(`^\s+\^\s*$`)
var endsWithSemicolonRegex, _ = regexp.Compile(`;\s*(--\s+.*)?$`)
var resultLineRegex, _ = regexp.Compile(`^-+(\+-+)*$`)
var resultRowRegex, _ = regexp.Compile(`^\((\d+) rows?\)$`)
var isNumberRegex, _ = regexp.Compile(`^[+\-]?(\d+\.\d*|\d*\.\d+|\d+)([eE][+\-]?\d+)?$`)

// NewTestFile creates a new *TestFile using the file located at the given path.
func NewTestFile(path string) *TestFile {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	lines := strings.Split(
		strings.ReplaceAll(
			strings.ReplaceAll(string(data), "\r\n", "\n"),
			"\n\n",
			"\n"),
		"\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if (strings.HasPrefix(lines[i], "--") && (len(lines[i]) == 2 || !strings.HasSuffix(lines[i], "-"))) ||
			len(lines[i]) == 0 ||
			strings.HasPrefix(lines[i], "WARNING: ") ||
			strings.HasPrefix(lines[i], "NOTICE: ") {
			copy(lines[i:], lines[i+1:])
			lines = lines[:len(lines)-1]
		}
	}
	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "DETAIL: ") {
			isCascading := false
			if strings.Contains(lines[i], "drop cascades to") {
				isCascading = true
			}
			copy(lines[i:], lines[i+1:])
			lines = lines[:len(lines)-1]
			if isCascading {
				for strings.HasPrefix(lines[i], "drop cascades to") {
					copy(lines[i:], lines[i+1:])
					lines = lines[:len(lines)-1]
				}
			}
			// Since we've changed the current line by deleting it, we need to decrement to counteract "for"'s increment
			i--
		}
	}
	return &TestFile{
		lines: lines,
		line:  0,
		isEOF: false,
	}
}

// ReadStatement reads the next statement.
func (tf *TestFile) ReadStatement() (string, bool) {
	if tf.isEOF {
		return "", false
	}
	statement := tf.lines[tf.line]
	tf.advance()
	for !tf.isEOF && !endsWithSemicolonRegex.MatchString(statement) {
		statement += "\n" + tf.lines[tf.line]
		tf.advance()
	}
	return statement, true
}

// GetError returns true when the next part is an error, in addition to the error message that it contains.
func (tf *TestFile) GetError() (string, bool) {
	if tf.isEOF {
		return "", false
	}
	if !strings.HasPrefix(tf.lines[tf.line], "ERROR:") {
		return "", false
	}
	errorMessage := strings.TrimSpace(tf.lines[tf.line][6:])
	tf.advance()
	for !tf.isEOF {
		if strings.HasPrefix(tf.lines[tf.line], "HINT:") {
			tf.advance()
			continue
		} else if strings.HasPrefix(tf.lines[tf.line], "LINE") && lineRegex.MatchString(tf.lines[tf.line]) {
			tf.advance()
			continue
		} else if strings.HasPrefix(tf.lines[tf.line], "     ") && linePointerRegex.MatchString(tf.lines[tf.line]) {
			tf.advance()
			continue
		} else {
			break
		}
	}
	return errorMessage, true
}

// GetResult returns true when the next part is a result set, in addition to the column names and results that it parsed.
func (tf *TestFile) GetResult() (results string, columns []string, ok bool) {
	if tf.isEOF {
		return "", nil, false
	}
	if tf.line+2 >= len(tf.lines) {
		return "", nil, false
	}
	if !strings.HasPrefix(tf.lines[tf.line], " ") || !strings.HasSuffix(tf.lines[tf.line], " ") {
		return "", nil, false
	}
	if !strings.HasPrefix(tf.lines[tf.line+1], "---") || !resultLineRegex.MatchString(tf.lines[tf.line+1]) {
		return "", nil, false
	}
	numOfResults := -1
	for i := 2; tf.line+i < len(tf.lines); i++ {
		if matches := resultRowRegex.FindStringSubmatch(tf.lines[tf.line+i]); len(matches) > 0 {
			if len(matches) != 2 {
				panic(fmt.Errorf("unexpected match length: %d\nline %d: %s", len(matches), tf.line+i+1, tf.lines[tf.line+i]))
			}
			var err error
			numOfResults, err = strconv.Atoi(matches[1])
			if err != nil {
				panic(fmt.Errorf("line %d\n%s", tf.line+i+1, err.Error()))
			}
			if numOfResults != i-2 {
				panic(fmt.Errorf("line %d: row count differs from line count (%d)", tf.line+i+1, i-2))
			}
			break
		}
	}
	if numOfResults == -1 {
		return "", nil, false
	}
	// Verified that this is a result, so now we can read it
	// We'll start by reading the column list
	columns = strings.Split(tf.lines[tf.line], " | ")
	for i := range columns {
		columns[i] = strings.TrimSpace(columns[i])
	}
	tf.advance()
	// Skip the column/row delimiter
	tf.advance()
	// Start reading the result rows
	resultBuffer := strings.Builder{}
	resultBuffer.WriteString("[]sql.Row{")
	for i := 0; i < numOfResults; i++ {
		if i > 0 {
			resultBuffer.WriteString(", ")
		}

		resultBuffer.WriteRune('{')
		values := strings.Split(tf.lines[tf.line], " | ")
		if len(values) != len(columns) {
			panic(fmt.Errorf("line %d: value count (%d) differs from column count (%d)", tf.line+1, len(values), len(columns)))
		}
		tf.advance()
		// Trimming spaces will change the result of strings that rely on spaces, so we need to find these differences
		for valIdx := range values {
			value := strings.TrimSpace(values[valIdx])
			if valIdx > 0 {
				resultBuffer.WriteString(", ")
			}
			// This is probably wrong and will mess up some results, but we'll check if it's a number to determine
			// whether it should be a string or not, in addition to boolean values.
			if isNumberRegex.MatchString(value) {
				resultBuffer.WriteString(value)
			} else if value == "t" {
				resultBuffer.WriteString("true")
			} else if value == "f" {
				resultBuffer.WriteString("false")
			} else {
				// If it has parenthesis, then we'll just use those
				if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
					resultBuffer.WriteString(value)
				} else {
					resultBuffer.WriteRune('`')
					resultBuffer.WriteString(value)
					resultBuffer.WriteRune('`')
				}
			}
		}
		resultBuffer.WriteRune('}')
	}
	resultBuffer.WriteRune('}')
	// Skip the (2 rows) part with a final advance
	tf.advance()
	return resultBuffer.String(), columns, true
}

func (tf *TestFile) advance() {
	if tf.isEOF {
		return
	}
	tf.line++
	if tf.line == len(tf.lines) {
		tf.isEOF = true
	}
}
