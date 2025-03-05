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

// jsonConvert handles the conversion from the JSON format into a format that is easier to work with.
func jsonConvert(jsonBlock plpgSQL_block) (Block, error) {
	block := Block{Label: jsonBlock.Action.StmtBlock.Label}
	for _, v := range jsonBlock.Datums {
		switch {
		case v.Row != nil:
			if len(v.Row.Fields) != 1 {
				return Block{}, errors.New("record types are not yet supported")
			}
		case v.Variable != nil:
			block.Variable = append(block.Variable, Variable{
				Name:        v.Variable.RefName,
				Type:        strings.ToLower(v.Variable.Type.Type.Name),
				IsParameter: v.Variable.LineNumber == 0,
			})
		default:
			return Block{}, errors.Errorf("unhandled datum type: %T", v)
		}
	}
	var err error
	block.Body, err = jsonConvertStatements(jsonBlock.Action.StmtBlock.Body)
	if err != nil {
		return Block{}, err
	}
	return block, nil
}

// jsonConvertStatement converts a statement in JSON form to the output form.
func jsonConvertStatement(stmt statement) (Statement, error) {
	switch {
	case stmt.Assignment != nil:
		return stmt.Assignment.Convert()
	case stmt.Case != nil:
		return stmt.Case.Convert()
	case stmt.ExecSQL != nil:
		return stmt.ExecSQL.Convert()
	case stmt.Exit != nil:
		return stmt.Exit.Convert(), nil
	case stmt.If != nil:
		return stmt.If.Convert()
	case stmt.Loop != nil:
		return stmt.Loop.Convert()
	case stmt.Perform != nil:
		return stmt.Perform.Convert(), nil
	case stmt.Raise != nil:
		return stmt.Raise.Convert(), nil
	case stmt.Return != nil:
		return stmt.Return.Convert(), nil
	case stmt.While != nil:
		return stmt.While.Convert()
	default:
		return Block{}, errors.Errorf("unhandled statement type: %T", stmt)
	}
}

// jsonConvertStatements converts a collection of statements in JSON form to their output form.
func jsonConvertStatements(stmts []statement) ([]Statement, error) {
	newStmts := make([]Statement, len(stmts))
	for i, stmt := range stmts {
		var err error
		newStmts[i], err = jsonConvertStatement(stmt)
		if err != nil {
			return nil, err
		}
	}
	return newStmts, nil
}
