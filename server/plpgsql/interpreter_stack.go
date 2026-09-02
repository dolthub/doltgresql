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
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"gopkg.in/src-d/go-errors.v1"

	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// ErrRecordNotAssigned is returned when a field is read from a RECORD variable that has not been assigned yet,
// and therefore has no shape to read a field from.
var ErrRecordNotAssigned = errors.NewKind(`record "%s" is not assigned yet`)

// ErrRecordHasNoField is returned when a field is read from a RECORD variable that has no such field.
var ErrRecordHasNoField = errors.NewKind(`record "%s" has no field "%s"`)

// ErrVariableNotFound is returned when a referenced variable does not exist on the stack.
var ErrVariableNotFound = errors.NewKind("variable `%s` could not be found")

// NormalizeIdentifier folds a PL/pgSQL identifier the way Postgres does, so that the name a reference is
// written with and the name a variable was declared with can be compared directly. An unquoted identifier
// folds to lowercase; a quoted one keeps its case and loses its quotes, with a doubled quote inside standing
// for a literal one. `MyVar` and `MYVAR` therefore name the same variable, and `"MyVar"` names a different
// one. pg_query already stores declared names in this form, so normalizing every name we take from raw
// source text is what puts both sides in the same alphabet.
func NormalizeIdentifier(ident string) string {
	if len(ident) >= 2 && strings.HasPrefix(ident, `"`) && strings.HasSuffix(ident, `"`) {
		return strings.ReplaceAll(ident[1:len(ident)-1], `""`, `"`)
	}
	return strings.ToLower(ident)
}

// NormalizeIdentifierPath folds a bare reference of the form `name` or `name.field` and reports whether the
// text was such a reference at all. Some statements, RAISE among them, carry their arguments as raw source
// text rather than through the expression rewriter, and each one may be either a variable reference or an
// expression to evaluate. Anything that is not a plain reference is left alone for the caller to evaluate.
func NormalizeIdentifierPath(text string) (string, bool) {
	text = strings.TrimSpace(text)
	// GetVariable resolves at most one level of field access, so anything deeper is not a reference it
	// could answer anyway.
	base, field, hasField := strings.Cut(text, ".")
	if !isIdentifier(base) || (hasField && !isIdentifier(field)) {
		return "", false
	}
	// Only the base names a variable; the field is matched against the record's columns separately.
	if hasField {
		return NormalizeIdentifier(base) + "." + field, true
	}
	return NormalizeIdentifier(base), true
}

// isIdentifier reports whether |s| is a single SQL identifier, quoted or otherwise.
func isIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	if strings.HasPrefix(s, `"`) {
		return len(s) >= 2 && strings.HasSuffix(s, `"`) && !strings.Contains(s[1:len(s)-1], `"`)
	}
	for i, r := range s {
		switch {
		case r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
		case i > 0 && (r == '$' || (r >= '0' && r <= '9')):
		default:
			return false
		}
	}
	return true
}

// QuoteIdentifier renders |name| so that it survives NormalizeIdentifier unchanged. Use it when building SQL
// that refers to a variable whose declared name is already known, since that name may hold capitals that an
// unquoted mention would fold away.
func QuoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// TriggerNewRecordName and TriggerOldRecordName are the names a trigger invocation supplies its row records
// under. They are the folded form of NEW and OLD, matching how pg_query names them in a parsed function.
const (
	TriggerNewRecordName = "new"
	TriggerOldRecordName = "old"
)

// FoundVariableName is the name of the built-in FOUND variable, which PL/pgSQL creates for every function
// and which reports whether the most recent statement that is defined to set it produced any row.
// https://www.postgresql.org/docs/15/plpgsql-statements.html#PLPGSQL-STATEMENTS-DIAGNOSTICS
const FoundVariableName = "found"

// cursorState holds the result set for a FOR record IN query LOOP cursor.
type cursorState struct {
	Schema sql.Schema
	Rows   []sql.Row
	Index  int
}

// interpreterVariable is a variable that lives on the stack. This will hold an actual value, but will not be directly
// interacted with. InterpreterVariableReference are, instead, the avenue of interaction as a variable may be an
// aggregate type (such as a record).
type interpreterVariable struct {
	Record sql.Schema // TODO: all records carry their type information alongside the value, so this is redundant
	Type   *pgtypes.DoltgresType
	Value  any
	// IsRecord marks a variable declared as RECORD (or a trigger's NEW/OLD). Such a variable has no shape
	// until something is assigned to it, at which point Record and Value are set together.
	IsRecord bool
}

// InterpreterVariableReference is a reference to a variable that lives on the stack. If the type is not null, then it
// is valid to dereference the value for assignment. We make use of references rather than directly interacting with
// the variables as this allows for interacting with sections of aggregate types (such as record) as well as normal
// variable interaction.
type InterpreterVariableReference struct {
	Type  *pgtypes.DoltgresType
	Value *any
}

// InterpreterScopeDetails contains all of the details that are relevant to a particular scope.
type InterpreterScopeDetails struct {
	variables map[string]*interpreterVariable
	label     string
	// cursor names the FOR..IN..SELECT cursor this scope owns, if it is such a loop's scope. The scope
	// owning it is what lets the cursor be torn down wherever the loop is left, rather than only where
	// the cursor runs out.
	cursor string
	// reportsFound marks the scope of a loop that sets FOUND when it is left, and iterated records
	// whether that loop ever advanced into its body.
	reportsFound bool
	iterated     bool
}

// InterpreterStack represents the working information that an interpreter will use during execution. It is not exactly
// the same as a stack in the traditional programming sense, but rather is a loose abstraction that serves the same
// general purpose.
type InterpreterStack struct {
	outParams []string
	stack     *utils.Stack[*InterpreterScopeDetails]
	runner    sql.StatementRunner
	labelID   int

	// returnQueryBuffer buffers results from RETURN QUERY statements
	returnQueryBuffer [][]pgtypes.RecordValue
	// cursors holds the active FOR record IN query LOOP result sets
	cursors map[string]*cursorState
}

// NewInterpreterStack creates a new InterpreterStack.
func NewInterpreterStack(runner sql.StatementRunner) InterpreterStack {
	stack := utils.NewStack[*InterpreterScopeDetails]()
	// This first push represents the function base, including parameters
	stack.Push(&InterpreterScopeDetails{
		variables: make(map[string]*interpreterVariable),
	})
	return InterpreterStack{
		outParams: make([]string, 0),
		stack:     stack,
		runner:    runner,
		cursors:   make(map[string]*cursorState),
	}
}

// Details returns the details for the current scope.
func (is *InterpreterStack) Details() *InterpreterScopeDetails {
	return is.stack.Peek()
}

// Runner returns the runner that is being used for the function's execution.
func (is *InterpreterStack) Runner() sql.StatementRunner {
	return is.runner
}

// GetCurrentLabel traverses the stack (starting from the top) returning the first label found. Returns an empty string
// if no labels were set.
func (is *InterpreterStack) GetCurrentLabel() string {
	for i := 0; i < is.stack.Len(); i++ {
		label := is.stack.PeekDepth(i).label
		if len(label) > 0 {
			return label
		}
	}
	return ""
}

// GetVariable traverses the stack (starting from the top) to find a variable with a matching name. Returns a
// reference with a nil type if no variable was found. Use GetVariableWithError when the reason for the failure
// matters.
func (is *InterpreterStack) GetVariable(name string) InterpreterVariableReference {
	ref, _ := is.GetVariableWithError(name)
	return ref
}

// GetVariableWithError behaves like GetVariable, but explains why the reference could not be resolved rather
// than returning an empty reference.
func (is *InterpreterStack) GetVariableWithError(name string) (InterpreterVariableReference, error) {
	// TODO: handle nested record access
	fullName := name
	fieldName := ""
	if strings.Count(name, ".") == 1 {
		splitName := strings.Split(name, ".")
		name = splitName[0]
		// A field may be written as a quoted identifier (`blocker."id"`), which reaches us with its quotes
		// still attached. Field lookup is case-insensitive here, so the quotes are all that need removing.
		fieldName = strings.Trim(splitName[1], `"`)
	}
	for i := 0; i < is.stack.Len(); i++ {
		if iv, ok := is.stack.PeekDepth(i).variables[name]; ok {
			if len(fieldName) == 0 {
				return InterpreterVariableReference{
					Type:  iv.Type,
					Value: &iv.Value,
				}, nil
			} else if len(iv.Record) > 0 {
				fieldIdx := recordFieldIndex(iv.Record, fieldName)
				if fieldIdx == -1 {
					return InterpreterVariableReference{}, ErrRecordHasNoField.New(name, fieldName)
				}
				fieldType, ok := iv.Record[fieldIdx].Type.(*pgtypes.DoltgresType)
				if !ok {
					return InterpreterVariableReference{}, fmt.Errorf(
						"field `%s` of record `%s` does not have a Postgres type", fieldName, name)
				}
				return InterpreterVariableReference{
					Type:  fieldType,
					Value: &(iv.Value.(sql.Row)[fieldIdx]),
				}, nil
			} else if iv.IsRecord {
				// A record that has never been assigned has no shape, so there is no field to read.
				return InterpreterVariableReference{}, ErrRecordNotAssigned.New(name)
			} else if iv.Type != nil && iv.Type.IsCompositeType() {
				for fieldIdx := range iv.Type.CompositeAttrs {
					if iv.Type.CompositeAttrs[fieldIdx].Name == fieldName {
						vals := iv.Value.([]pgtypes.RecordValue)
						return InterpreterVariableReference{
							Type:  vals[fieldIdx].Type.(*pgtypes.DoltgresType),
							Value: &(vals[fieldIdx].Value),
						}, nil
					}
				}
				return InterpreterVariableReference{}, ErrRecordHasNoField.New(name, fieldName)
			} else {
				return InterpreterVariableReference{}, fmt.Errorf(
					"could not identify column `%s` in variable `%s`", fieldName, name)
			}
		}
	}
	return InterpreterVariableReference{}, ErrVariableNotFound.New(fullName)
}

// recordFieldIndex returns the index of the field named |fieldName| within |sch|, or -1 if there is no such
// field. Unlike sql.Schema.IndexOf, the column's source is ignored, since a record's shape comes from the
// output columns of whatever query was assigned to it and those may originate in different tables.
func recordFieldIndex(sch sql.Schema, fieldName string) int {
	for i, col := range sch {
		if strings.EqualFold(col.Name, fieldName) {
			return i
		}
	}
	return -1
}

// ListVariables returns a map with the names of all variables. The attached slice represents field names for
// records. Names are keyed exactly as they were declared, which is already the folded form, so a reference
// matches by being folded the same way rather than by being compared case-insensitively.
func (is *InterpreterStack) ListVariables() map[string][]string {
	seen := make(map[string][]string)
	for i := 0; i < is.stack.Len(); i++ {
		for varName, iv := range is.stack.PeekDepth(i).variables {
			var fieldNames []string
			if len(iv.Record) > 0 {
				for _, col := range iv.Record {
					fieldNames = append(fieldNames, strings.ToLower(col.Name))
				}
			}
			seen[varName] = fieldNames
		}
	}
	return seen
}

// NewRecord creates a new record in the current scope. If a record with the same name exists in a previous scope, then
// that record will be shadowed until the current scope exits. A nil |sch| declares a record that has no shape yet,
// which is what a `DECLARE r RECORD` produces until something is assigned to it. When |sch| is non-nil but |val| is
// nil, the record has a shape whose every field is NULL, which is what a trigger's OLD record is on an INSERT.
func (is *InterpreterStack) NewRecord(name string, sch sql.Schema, val sql.Row) {
	is.stack.Peek().variables[name] = &interpreterVariable{
		Record:   sch,
		Type:     pgtypes.Trigger, // TODO: we need to implement the RECORD pseudotype and replace the TRIGGER type here
		Value:    copyRecordRow(sch, val),
		IsRecord: true,
	}
}

// copyRecordRow returns a copy of |val| for storage in a record. A nil |val| becomes an all-NULL row matching
// |sch|, so that every field of a record with a known shape is addressable.
func copyRecordRow(sch sql.Schema, val sql.Row) sql.Row {
	if val == nil {
		if len(sch) == 0 {
			return nil
		}
		return make(sql.Row, len(sch))
	}
	newVal := make(sql.Row, len(val))
	copy(newVal, val)
	return newVal
}

// NewVariable creates a new variable in the current scope. If a variable with the same name exists in a previous scope,
// then that variable will be shadowed until the current scope exits.
func (is *InterpreterStack) NewVariable(name string, typ *pgtypes.DoltgresType) {
	is.NewVariableWithValue(name, typ, typ.Zero())
}

// NewVariableWithValue creates a new variable in the current scope, setting its initial value to the one given.
func (is *InterpreterStack) NewVariableWithValue(name string, typ *pgtypes.DoltgresType, val any) {
	is.stack.Peek().variables[name] = &interpreterVariable{
		Type:  typ,
		Value: val,
	}
}

// NewVariableAlias creates a new variable alias, named |alias|, in the current frame of this stack,
// pointing to the specified |variable|.
func (is *InterpreterStack) NewVariableAlias(alias string, target string) {
	for i := 0; i < is.stack.Len(); i++ {
		if iv, ok := is.stack.PeekDepth(i).variables[target]; ok {
			// TODO: this won't work for RECORD types
			is.stack.Peek().variables[alias] = iv
			break
		}
	}
}

// PushScope creates a new scope.
func (is *InterpreterStack) PushScope() {
	is.stack.Push(&InterpreterScopeDetails{
		variables: make(map[string]*interpreterVariable),
	})
}

// PopScope removes the current scope.
func (is *InterpreterStack) PopScope() {
	is.stack.Pop()
}

// SetOutParams sets the output parameter names slice for the stack.
func (is *InterpreterStack) SetOutParams(name []string) {
	is.outParams = name
}

// SetVariable sets the first variable found, with a matching name, to the value given. This does not ensure that the
// value matches the expectations of the type, so it should be validated before this is called. Returns an error if the
// variable cannot be found.
func (is *InterpreterStack) SetVariable(ctx *sql.Context, name string, val any) error {
	iv := is.GetVariable(name)
	if iv.Type == nil {
		return fmt.Errorf("variable `%s` could not be found", name)
	}
	*iv.Value = val
	return nil
}

// SetLabel sets the label for the current scope.
func (is *InterpreterStack) SetLabel(label string) {
	is.stack.Peek().label = label
}

// SetAnonymousLabel sets the label for the current scope to a guaranteed unique value.
func (is *InterpreterStack) SetAnonymousLabel() {
	// Postgres labels cannot have a tab character, so we can generate a label with one to guarantee it's unique
	is.stack.Peek().label = fmt.Sprintf("\t%d", is.labelID)
	is.labelID++
}

// BufferReturnQueryResults buffers |results| from a RETURN QUERY statement so that they can be returned when
// the function exits. If results from a previous RETURN QUERY call have already been buffered, |results| will
// be appended.
func (is *InterpreterStack) BufferReturnQueryResults(results [][]pgtypes.RecordValue) {
	is.returnQueryBuffer = append(is.returnQueryBuffer, results...)
}

// ReturnQueryResults returns the buffered results from a RETURN QUERY statement.
func (is *InterpreterStack) ReturnQueryResults() [][]pgtypes.RecordValue {
	return is.returnQueryBuffer
}

// ReturnOutParamResults returns the results in output parameters if there is any from a RETURN QUERY statement.
func (is *InterpreterStack) ReturnOutParamResults() any {
	// single OUT parameter is not record type
	if len(is.outParams) == 1 {
		v := is.GetVariable(is.outParams[0])
		return *v.Value
	}
	// multiple OUT parameters are record type
	record := make([]pgtypes.RecordValue, len(is.outParams))
	for i, iv := range is.outParams {
		v := is.GetVariable(iv)
		record[i] = pgtypes.RecordValue{
			Value: *v.Value,
			Type:  v.Type,
		}
	}
	return record
}

// InitCursor stores the result set for a FOR record IN query LOOP cursor. The cursor is opened in the
// loop's own scope, which takes ownership of it.
func (is *InterpreterStack) InitCursor(name string, schema sql.Schema, rows []sql.Row) {
	is.cursors[name] = &cursorState{
		Schema: schema,
		Rows:   rows,
		Index:  0,
	}
	is.stack.Peek().cursor = name
}

// ScopeCursor returns the name of the cursor the current scope owns, or an empty string when the scope is
// not that of a FOR..IN..SELECT loop.
func (is *InterpreterStack) ScopeCursor() string {
	return is.stack.Peek().cursor
}

// AdvanceCursor returns the next row for the named cursor and advances its index.
// Returns (schema, row, true) if a row is available, or (nil, nil, false) when exhausted.
func (is *InterpreterStack) AdvanceCursor(name string) (sql.Schema, sql.Row, bool) {
	cs, ok := is.cursors[name]
	if !ok || cs.Index >= len(cs.Rows) {
		return nil, nil, false
	}
	row := cs.Rows[cs.Index]
	cs.Index++
	return cs.Schema, row, true
}

// CloseCursor removes the named cursor from the stack.
func (is *InterpreterStack) CloseCursor(name string) {
	delete(is.cursors, name)
}

// MarkScopeLoop marks the current scope as a loop's, whose exit reports FOUND, and records whether the loop
// has just advanced into its body. Only the loop itself knows that it advanced, and it cannot record it in
// FOUND, which PostgreSQL leaves alone until the loop is left.
func (is *InterpreterStack) MarkScopeLoop(advanced bool) {
	details := is.stack.Peek()
	details.reportsFound = true
	details.iterated = details.iterated || advanced
}

// ScopeLoop reports whether leaving the current scope sets FOUND, and if so whether its loop body ran.
func (is *InterpreterStack) ScopeLoop() (reportsFound bool, iterated bool) {
	details := is.stack.Peek()
	return details.reportsFound, details.iterated
}

// SetFound updates the built-in FOUND variable. Functions compiled before FOUND was supported do not declare
// it, and a function may shadow the name with a variable of its own, so anything other than the built-in
// boolean is left alone rather than being overwritten.
func (is *InterpreterStack) SetFound(ctx *sql.Context, found bool) error {
	ref := is.GetVariable(FoundVariableName)
	if ref.Type == nil || ref.Type.ID != pgtypes.Bool.ID {
		return nil
	}
	return is.SetVariable(ctx, FoundVariableName, found)
}

// UpdateRecord finds the named variable and sets its schema and row value. A nil |val| gives the record the
// shape of |schema| with every field NULL, which is what PostgreSQL does when an INTO query matches no rows.
func (is *InterpreterStack) UpdateRecord(name string, schema sql.Schema, val sql.Row) error {
	normalizedSchema, err := normalizeRecordSchema(schema)
	if err != nil {
		return err
	}
	for i := 0; i < is.stack.Len(); i++ {
		if iv, ok := is.stack.PeekDepth(i).variables[name]; ok {
			iv.Record = normalizedSchema
			iv.Value = copyRecordRow(normalizedSchema, val)
			iv.IsRecord = true
			return nil
		}
	}
	return fmt.Errorf("record variable `%s` could not be found", name)
}

// normalizeRecordSchema returns |schema| with every column type converted to a DoltgresType. Query results can
// carry plain GMS types (an aggregate such as `count(*)`, for example), but a record's fields are read back as
// Doltgres values, so the types have to be converted before the schema is stored.
func normalizeRecordSchema(schema sql.Schema) (sql.Schema, error) {
	normalized := make(sql.Schema, len(schema))
	for i, col := range schema {
		if _, ok := col.Type.(*pgtypes.DoltgresType); ok {
			normalized[i] = col
			continue
		}
		// TODO: this only converts the type, not the value. See the same TODO in convertRowsToRecords.
		doltgresType, err := pgtypes.FromGmsTypeToDoltgresType(col.Type)
		if err != nil {
			return nil, err
		}
		newCol := *col
		newCol.Type = doltgresType
		normalized[i] = &newCol
	}
	return normalized, nil
}
