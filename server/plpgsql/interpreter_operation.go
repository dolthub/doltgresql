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

// OpCode states the operation to be performed. Most operations have a direct analogue to a Pl/pgSQL operation, however
// some exist only in Doltgres (specific to our interpreter implementation).
type OpCode uint16

const (
	// New OpCode values MUST be added to the END of this list!
	// Function OpCodes are persisted to disk, so these values MUST be stable across Doltgres versions.
	OpCode_Alias       OpCode = 0  // https://www.postgresql.org/docs/15/plpgsql-declarations.html#PLPGSQL-DECLARATION-ALIAS
	OpCode_Assign      OpCode = 1  // https://www.postgresql.org/docs/15/plpgsql-statements.html#PLPGSQL-STATEMENTS-ASSIGNMENT
	OpCode_Case        OpCode = 2  // https://www.postgresql.org/docs/15/plpgsql-control-structures.html#PLPGSQL-CONDITIONALS
	OpCode_Declare     OpCode = 3  // https://www.postgresql.org/docs/15/plpgsql-declarations.html
	OpCode_DeleteInto  OpCode = 4  // https://www.postgresql.org/docs/15/plpgsql-statements.html
	OpCode_Exception   OpCode = 5  // https://www.postgresql.org/docs/15/plpgsql-control-structures.html#PLPGSQL-ERROR-TRAPPING
	OpCode_Execute     OpCode = 6  // Executing a standard SQL statement (expects no rows returned unless Target is specified)
	OpCode_Get         OpCode = 7  // https://www.postgresql.org/docs/15/plpgsql-statements.html#PLPGSQL-STATEMENTS-DIAGNOSTICS
	OpCode_Goto        OpCode = 8  // All control-flow structures can be represented using Goto
	OpCode_If          OpCode = 9  // https://www.postgresql.org/docs/15/plpgsql-control-structures.html#PLPGSQL-CONDITIONALS
	OpCode_InsertInto  OpCode = 10 // https://www.postgresql.org/docs/15/plpgsql-statements.html
	OpCode_Perform     OpCode = 11 // https://www.postgresql.org/docs/15/plpgsql-statements.html
	OpCode_Raise       OpCode = 12 // https://www.postgresql.org/docs/15/plpgsql-errors-and-messages.html
	OpCode_Return      OpCode = 13 // https://www.postgresql.org/docs/15/plpgsql-control-structures.html#PLPGSQL-STATEMENTS-RETURNING
	OpCode_ScopeBegin  OpCode = 14 // This is used for scope control, specific to Doltgres
	OpCode_ScopeEnd    OpCode = 15 // This is used for scope control, specific to Doltgres
	OpCode_SelectInto  OpCode = 16 // https://www.postgresql.org/docs/15/plpgsql-statements.html
	OpCode_UpdateInto  OpCode = 17 // https://www.postgresql.org/docs/15/plpgsql-statements.html
	OpCode_ReturnQuery OpCode = 18 // https://www.postgresql.org/docs/current/plpgsql-control-structures.html#PLPGSQL-STATEMENTS-RETURNING-RETURN-NEXT
	// New OpCode values MUST be added to the END of this list!
	// Function OpCodes are persisted to disk, so these values MUST be stable across Doltgres versions.
)

// InterpreterOperation is an operation that will be performed by the interpreter.
type InterpreterOperation struct {
	OpCode        OpCode
	PrimaryData   string            // This will represent the "main" data, such as the query for PERFORM, expression for IF, etc.
	SecondaryData []string          // This represents auxiliary data, such as bindings, strictness, etc.
	Target        string            // This is the variable that will store the results (if applicable)
	Index         int               // This is the index that should be set for operations that move the function counter
	Options       map[string]string // This is extra data for operations that need it
}
