// Copyright 2026 Dolthub, Inc.
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

package server

import (
	"bytes"
	"fmt"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/go-mysql-server/sql"
)

// pgError is the error structure returned from calling a db library function for postgres
type pgError struct {
	State   pgcode.Code
	Message string
}

// Error implements the error interface.
func (se *pgError) Error() string {
	buf := &bytes.Buffer{}
	buf.WriteString(se.Message)
	return buf.String()
}

// CastSQLError returns a *pgError with the error SQL state code, populated for the specified error object.
// Many tools (e.g. ORMs, SQL workbenches) rely on this error metadata to work correctly. If the specified error is nil,
// nil will be returned. If the error is already of type *pgError, the error will be returned as is.
func CastSQLError(err error) *pgError {
	if err == nil {
		return nil
	}
	if mysqlErr, ok := err.(*pgError); ok {
		return mysqlErr
	}

	if w, ok := err.(sql.WrappedInsertError); ok {
		return CastSQLError(w.Cause)
	}

	if wm, ok := err.(sql.WrappedTypeConversionError); ok {
		return CastSQLError(wm.Err)
	}

	// TODO: map more errors case
	// TODO: should update the error message to match Postgres
	var code pgcode.Code
	switch {
	case sql.ErrCheckConstraintViolated.Is(err):
		code = pgcode.CheckViolation
	case sql.ErrDatabaseExists.Is(err):
		code = pgcode.DuplicateDatabase
	case sql.ErrDatabaseNotFound.Is(err):
		code = pgcode.UndefinedDatabase
	case sql.ErrDatabaseSchemaExists.Is(err):
		code = pgcode.DuplicateSchema
	case sql.ErrDatabaseSchemaNotFound.Is(err):
		code = pgcode.UndefinedSchema
	case sql.ErrForeignKeyChildViolation.Is(err):
		code = pgcode.ForeignKeyViolation
	case sql.ErrForeignKeyDuplicateName.Is(err):
		code = pgcode.DuplicateObject
	case sql.ErrTableNotFound.Is(err):
		code = pgcode.UndefinedTable
	case sql.ErrUniqueKeyViolation.Is(err):
		code = pgcode.UniqueViolation
	default:
		code = pgcode.Internal
	}

	return &pgError{
		State:   code,
		Message: fmt.Sprintf("%s", err.Error()),
	}
}
