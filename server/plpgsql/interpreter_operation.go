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
	OpCode_Alias      OpCode = iota // https://www.postgresql.org/docs/15/plpgsql-declarations.html#PLPGSQL-DECLARATION-ALIAS
	OpCode_Assign                   // https://www.postgresql.org/docs/15/plpgsql-statements.html#PLPGSQL-STATEMENTS-ASSIGNMENT
	OpCode_Case                     // https://www.postgresql.org/docs/15/plpgsql-control-structures.html#PLPGSQL-CONDITIONALS
	OpCode_Declare                  // https://www.postgresql.org/docs/15/plpgsql-declarations.html
	OpCode_DeleteInto               // https://www.postgresql.org/docs/15/plpgsql-statements.html
	OpCode_Exception                // https://www.postgresql.org/docs/15/plpgsql-control-structures.html#PLPGSQL-ERROR-TRAPPING
	OpCode_Execute                  // Executing a standard SQL statement (expects no rows returned unless Target is specified)
	OpCode_Get                      // https://www.postgresql.org/docs/15/plpgsql-statements.html#PLPGSQL-STATEMENTS-DIAGNOSTICS
	OpCode_Goto                     // All control-flow structures can be represented using Goto
	OpCode_If                       // https://www.postgresql.org/docs/15/plpgsql-control-structures.html#PLPGSQL-CONDITIONALS
	OpCode_InsertInto               // https://www.postgresql.org/docs/15/plpgsql-statements.html
	OpCode_Perform                  // https://www.postgresql.org/docs/15/plpgsql-statements.html
	OpCode_Raise                    // https://www.postgresql.org/docs/15/plpgsql-errors-and-messages.html
	OpCode_Return                   // https://www.postgresql.org/docs/15/plpgsql-control-structures.html#PLPGSQL-STATEMENTS-RETURNING
	OpCode_ScopeBegin               // This is used for scope control, specific to Doltgres
	OpCode_ScopeEnd                 // This is used for scope control, specific to Doltgres
	OpCode_SelectInto               // https://www.postgresql.org/docs/15/plpgsql-statements.html
	OpCode_UpdateInto               // https://www.postgresql.org/docs/15/plpgsql-statements.html
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
