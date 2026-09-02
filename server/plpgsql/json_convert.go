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
	block := Block{
		TriggerNew: jsonBlock.NewVariableNumber,
		TriggerOld: jsonBlock.OldVariableNumber,
		Label:      jsonBlock.Action.StmtBlock.Label,
	}
	lowestRecordNumber := int32(2147483647)
	// We do a first loop to determine the offset for the first record
	for _, v := range jsonBlock.Datums {
		switch {
		case v.Record != nil:
			if v.Record.DatumNumber < lowestRecordNumber {
				lowestRecordNumber = v.Record.DatumNumber
			}
		}
	}
	offset := int32(0) - lowestRecordNumber
	// PL/pgSQL creates the built-in FOUND variable itself, immediately after the function's parameters, so
	// it arrives looking like a parameter (no line number). We track where it lands so it can be turned
	// back into a declaration below. A parameter may also be named `found`, in which case the built-in is
	// the later of the two and shadows it, so the last match is the one that counts.
	foundVariableIndex := -1
	// Then we do a second loop that actually adds all of the datums to the block
	for _, v := range jsonBlock.Datums {
		switch {
		case v.Record != nil:
			// TODO: support normal record types
			datumNumber := v.Record.DatumNumber + offset
			if int(datumNumber) >= len(block.Records) {
				oldRecords := block.Records
				block.Records = make([]Record, datumNumber+1)
				copy(block.Records, oldRecords)
			}

			if v.Record.DatumNumber > 0 {
				block.Records[datumNumber].Name = v.Record.RefName
			}
		case v.RecordField != nil:
			recordParentNumber := v.RecordField.RecordParentNumber + offset
			if int(recordParentNumber) >= len(block.Records) {
				return Block{}, errors.New("invalid record parent number")
			}
			block.Records[recordParentNumber].Fields = append(
				block.Records[recordParentNumber].Fields, v.RecordField.FieldName)
		case v.Row != nil:
		case v.Variable != nil:
			if v.Variable.LineNumber == 0 && strings.EqualFold(v.Variable.RefName, FoundVariableName) {
				foundVariableIndex = len(block.Variables)
			}
			block.Variables = append(block.Variables, Variable{
				Name:        v.Variable.RefName,
				Type:        strings.ToLower(v.Variable.Type.Type.Name),
				IsParameter: v.Variable.LineNumber == 0,
				Default:     v.Variable.Default.Var.Query,
			})
		default:
			// The datum struct is a union of pointers, so printing its type here would only ever say
			// "plpgsql.datum". Name the arms we do handle instead.
			return Block{}, errors.New("unhandled declared variable: expected a record, record field, row, or variable")
		}
	}
	// FOUND is not passed in by the caller, so it has to be declared and initialized like any other local.
	// Postgres starts it out false.
	if foundVariableIndex >= 0 {
		block.Variables[foundVariableIndex].IsParameter = false
		block.Variables[foundVariableIndex].Default = "false"
		// The datum names the type as `pg_catalog."boolean"`, which OpCode_Declare cannot resolve: it maps
		// a qualified pg_catalog name through TypeForNonKeywordTypeName, which has no entry for the
		// `boolean` spelling, so the lookup is left asking for a type named `boolean` and fails. Name the
		// type by its canonical name instead. The same gap makes a hand-written `DECLARE b
		// pg_catalog.boolean` fail, so teaching that alias table about `boolean` would let this go away.
		block.Variables[foundVariableIndex].Type = "pg_catalog.bool"
	}
	// The NEW and OLD records of a trigger appear in the datum list like any other record, but they are
	// supplied by the trigger invocation rather than declared by the function, so they must not be
	// redeclared (which would shadow the supplied values with an empty record).
	for _, triggerRecordNumber := range []int32{jsonBlock.NewVariableNumber, jsonBlock.OldVariableNumber} {
		if triggerRecordNumber == 0 {
			continue
		}
		if idx := triggerRecordNumber + offset; idx >= 0 && int(idx) < len(block.Records) {
			block.Records[idx].IsTriggerRecord = true
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
	case stmt.Block != nil:
		stmts, err := jsonConvertStatements(stmt.Block.Body)
		if err != nil {
			return Block{}, err
		}
		return Block{
			Body: stmts,
		}, nil
	case stmt.Call != nil:
		return stmt.Call.Convert()
	case stmt.Case != nil:
		return stmt.Case.Convert()
	case stmt.DynExec != nil:
		return stmt.DynExec.Convert()
	case stmt.ExecSQL != nil:
		return stmt.ExecSQL.Convert()
	case stmt.Exit != nil:
		return stmt.Exit.Convert(), nil
	case stmt.ForILoop != nil:
		return stmt.ForILoop.Convert()
	case stmt.ForSLoop != nil:
		return stmt.ForSLoop.Convert()
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
	case stmt.ReturnQuery != nil:
		return stmt.ReturnQuery.Convert(), nil
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
