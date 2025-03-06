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
	"encoding/json"

	"github.com/cockroachdb/errors"
	pg_query "github.com/pganalyze/pg_query_go/v6"

	"github.com/dolthub/doltgresql/core/interpreter"
)

// Parse parses the given CREATE FUNCTION string (which must be the entire string, not just the body) into a Block
// containing the contents of the body.
func Parse(fullCreateFunctionString string) ([]interpreter.InterpreterOperation, error) {
	var functions []function
	parsedBody, err := pg_query.ParsePlPgSqlToJSON(fullCreateFunctionString)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(parsedBody), &functions)
	if err != nil {
		return nil, err
	}
	if len(functions) != 1 {
		return nil, errors.New("CREATE FUNCTION parsed multiple blocks")
	}
	block, err := jsonConvert(functions[0].Function)
	if err != nil {
		return nil, err
	}
	ops := make([]interpreter.InterpreterOperation, 0, len(block.Body)+len(block.Variable))
	stack := NewInterpreterStack(nil)
	if err = block.AppendOperations(&ops, &stack); err != nil {
		return nil, err
	}
	if err = reconcileLabels(ops); err != nil {
		return nil, err
	}
	return ops, nil
}
