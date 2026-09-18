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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ErrorResponseSeverity represents the severity of an ErrorResponse message.
type ErrorResponseSeverity string

const (
	ErrorResponseSeverity_Error   ErrorResponseSeverity = "ERROR"
	ErrorResponseSeverity_Fatal   ErrorResponseSeverity = "FATAL"
	ErrorResponseSeverity_Panic   ErrorResponseSeverity = "PANIC"
	ErrorResponseSeverity_Warning ErrorResponseSeverity = "WARNING"
	ErrorResponseSeverity_Notice  ErrorResponseSeverity = "NOTICE"
	ErrorResponseSeverity_Debug   ErrorResponseSeverity = "DEBUG"
	ErrorResponseSeverity_Info    ErrorResponseSeverity = "INFO"
	ErrorResponseSeverity_Log     ErrorResponseSeverity = "LOG"
)

// endOfMessages completes an operation with an optional error and sends ReadyForQuery.
func (h *ConnectionHandler) endOfMessages(err error) {
	if err != nil {
		h.handleOperationError(err)
	}
	if sendErr := h.send(&pgproto3.ReadyForQuery{TxStatus: h.state.txState.readyStatus()}); sendErr != nil {
		panic(sendErr)
	}
}

// handleMessageError applies transaction failure semantics and selects protocol recovery.
func (h *ConnectionHandler) handleMessageError(err error) {
	switch h.state.mode {
	case copyInConnectionMode:
		copyFrom := h.state.activeCopyFrom
		if copyFrom == nil {
			h.state.beginCloseConnectionMode()
			h.handleOperationError(errors.Wrap(err, "COPY mode has no active operation"))
			return
		}
		h.rollbackCopyInTransaction(copyFrom.txOwnership)
		if !h.state.finishCopyInMode(copyFrom) {
			h.state.beginCloseConnectionMode()
			h.handleOperationError(errors.Wrap(err, "COPY FROM STDIN state changed during error recovery"))
			return
		}
		if copyFrom.suspendedSimpleQuery != nil {
			h.endOfMessages(err)
			return
		}
		h.state.beginDiscardUntilSyncMode()
		h.handleOperationError(err)
	case extendedQueryConnectionMode:
		h.state.beginDiscardUntilSyncMode()
		h.handleOperationError(err)
	case discardUntilSyncConnectionMode, closingConnectionMode:
		h.handleOperationError(err)
	case readyConnectionMode:
		h.endOfMessages(err)
	default:
		h.state.beginCloseConnectionMode()
		h.handleOperationError(errors.Wrap(err, "invalid connection mode"))
	}
}

// handleOperationError updates transaction state and sends ErrorResponse without selecting recovery.
func (h *ConnectionHandler) handleOperationError(err error) {
	h.failActiveTransaction()
	h.sendError(err)
}

// sendError sends one PostgreSQL ErrorResponse to the client.
func (h *ConnectionHandler) sendError(err error) {
	pgErr := castSQLError(err)
	if sendErr := h.send(&pgproto3.ErrorResponse{
		Severity: pgErr.Severity,
		Code:     pgErr.Code,
		Message:  pgErr.Message,
	}); sendErr != nil {
		panic(sendErr)
	}
}

// castSQLError returns a *pgconn.PgError with the error SQL state code, populated for the specified error object.
// Many tools (e.g. ORMs, SQL workbenches) rely on this error metadata to work correctly. If the specified error is nil,
// nil will be returned. If the error is already of type *pgconn.PgError, the error will be returned as is.
func castSQLError(err error) *pgconn.PgError {
	if err == nil {
		return nil
	}
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr
	}

	if w, ok := err.(sql.WrappedInsertError); ok {
		return castSQLError(w.Cause)
	}

	if wm, ok := err.(sql.WrappedTypeConversionError); ok {
		return castSQLError(wm.Err)
	}

	// Errors originating in our Postgres-derived parser already carry a candidate SQLSTATE (e.g. 42601
	// for syntax errors, 0A000 for unimplemented syntax); report that code directly when present.
	if pgerror.HasCandidateCode(err) {
		if code := pgerror.GetPGCode(err); code != pgcode.Uncategorized {
			return &pgconn.PgError{
				Severity: string(ErrorResponseSeverity_Error),
				Code:     code.String(),
				Message:  err.Error(),
			}
		}
	}

	// Errors that reach a client with an XX-class (internal error) code are more than cosmetic: some
	// clients (e.g. Npgsql) treat XX-class errors as critical failures and close the connection, so any
	// error a client can provoke should be mapped to its proper SQLSTATE here.
	// TODO: should update the error message to match Postgres
	var code pgcode.Code
	switch {
	// Class 42 — Syntax Error or Access Rule Violation
	case sql.ErrInsertConflictTarget.Is(err):
		code = pgcode.InvalidColumnReference
	// Class 0A — Feature Not Supported
	case sql.ErrUnsupportedFeature.Is(err), sql.ErrUnsupportedSyntax.Is(err):
		code = pgcode.FeatureNotSupported
	// Class 21 — Cardinality Violation
	case sql.ErrExpectedSingleRow.Is(err), sql.ErrMoreThanOneRow.Is(err):
		code = pgcode.CardinalityViolation
	// Class 22 — Data Exception
	case pgtypes.ErrDivisionByZero.Is(err):
		code = pgcode.DivisionByZero
	case sql.ErrValueOutOfRange.Is(err), sql.ErrIntegerOutOfRange.Is(err), pgtypes.ErrValueIsOutOfRangeForType.Is(err),
		pgtypes.ErrOutOfRange.Is(err), pgtypes.ErrInputOutOfRange.Is(err),
		errors.Is(err, pgtypes.ErrCastOutOfRange):
		code = pgcode.NumericValueOutOfRange
	case pgtypes.ErrInvalidSyntaxForType.Is(err), sql.ErrInvalidValue.Is(err):
		code = pgcode.InvalidTextRepresentation
	case pgtypes.ErrWrongLengthBit.Is(err), pgtypes.ErrVarBitLengthExceeded.Is(err):
		code = pgcode.StringDataLengthMismatch
	case sql.ErrInvalidTimeZone.Is(err), sql.ErrInvalidArgument.Is(err), sql.ErrInvalidArgumentDetails.Is(err):
		code = pgcode.InvalidParameterValue
	// Class 23 — Integrity Constraint Violation
	case sql.ErrPrimaryKeyViolation.Is(err), sql.ErrUniqueKeyViolation.Is(err),
		sql.ErrDuplicateEntry.Is(err), sql.ErrDuplicateEntrySet.Is(err):
		code = pgcode.UniqueViolation
	case sql.ErrForeignKeyChildViolation.Is(err), sql.ErrForeignKeyParentViolation.Is(err),
		sql.ErrForeignKeyNotResolved.Is(err):
		code = pgcode.ForeignKeyViolation
	case sql.ErrCheckConstraintViolated.Is(err), pgtypes.ErrDomainValueViolatesCheckConstraint.Is(err):
		code = pgcode.CheckViolation
	case sql.ErrInsertIntoNonNullableProvidedNull.Is(err), sql.ErrFieldNoDefaultValue.Is(err),
		sql.ErrColumnDefaultReturnedNull.Is(err), pgtypes.ErrDomainDoesNotAllowNullValues.Is(err):
		code = pgcode.NotNullViolation
	// Class 25 — Invalid Transaction State
	case sql.ErrReadOnly.Is(err), sql.ErrReadOnlyTransaction.Is(err):
		code = pgcode.ReadOnlySQLTransaction
	// Classes 26, 34, 3B — statement, cursor, and savepoint names
	case sql.ErrUnknownPreparedStatement.Is(err):
		code = pgcode.InvalidSQLStatementName
	case sql.ErrCursorNotFound.Is(err):
		code = pgcode.InvalidCursorName
	case sql.ErrCursorAlreadyOpen.Is(err):
		code = pgcode.DuplicateCursor
	case sql.ErrSavepointDoesNotExist.Is(err):
		code = pgcode.InvalidSavepointSpecification
	// Classes 3D, 3F — catalog and schema names
	case sql.ErrDatabaseExists.Is(err):
		code = pgcode.DuplicateDatabase
	case sql.ErrDatabaseNotFound.Is(err):
		code = pgcode.UndefinedDatabase
	case sql.ErrDatabaseSchemaExists.Is(err):
		code = pgcode.DuplicateSchema
	case sql.ErrDatabaseSchemaNotFound.Is(err):
		code = pgcode.UndefinedSchema
	// Class 40 — Transaction Rollback. Dolt reports commit-time transaction conflicts as ErrLockDeadlock.
	case sql.ErrLockDeadlock.Is(err):
		code = pgcode.SerializationFailure
	// Class 42 — Syntax or Access Rule Violation
	case sql.ErrSyntaxError.Is(err), sql.ErrInvalidSyntax.Is(err), sql.ErrColValCountMismatch.Is(err),
		sql.ErrInsertIntoMismatchValueCount.Is(err), sql.ErrColumnNumberDoesNotMatch.Is(err):
		code = pgcode.Syntax
	case sql.ErrPrivilegeCheckFailed.Is(err), sql.ErrDatabaseAccessDeniedForUser.Is(err),
		sql.ErrTableAccessDeniedForUser.Is(err):
		code = pgcode.InsufficientPrivilege
	case sql.ErrNonAggregatedColumnWithoutGroupBy.Is(err):
		code = pgcode.Grouping
	case sql.ErrTableNotFound.Is(err), sql.ErrUnknownTable.Is(err), sql.ErrViewDoesNotExist.Is(err):
		code = pgcode.UndefinedTable
	case sql.ErrTableAlreadyExists.Is(err), sql.ErrExistingView.Is(err):
		code = pgcode.DuplicateRelation
	case sql.ErrColumnNotFound.Is(err), sql.ErrTableColumnNotFound.Is(err), sql.ErrUnknownColumn.Is(err),
		sql.ErrKeyColumnDoesNotExist.Is(err):
		code = pgcode.UndefinedColumn
	case sql.ErrColumnExists.Is(err), sql.ErrDuplicateColumn.Is(err), sql.ErrColumnSpecifiedTwice.Is(err):
		code = pgcode.DuplicateColumn
	case sql.ErrAmbiguousColumnName.Is(err), sql.ErrAmbiguousColumnOrAliasName.Is(err),
		sql.ErrAmbiguousColumnInOrderBy.Is(err):
		code = pgcode.AmbiguousColumn
	case sql.ErrFunctionNotFound.Is(err), sql.ErrTableFunctionNotFound.Is(err),
		sql.ErrInvalidArgumentNumber.Is(err), framework.ErrFunctionDoesNotExist.Is(err):
		code = pgcode.UndefinedFunction
	case sql.ErrForeignKeyDuplicateName.Is(err), pgtypes.ErrTypeAlreadyExists.Is(err):
		code = pgcode.DuplicateObject
	case sql.ErrUnknownSystemVariable.Is(err), sql.ErrUnknownConstraint.Is(err), pgtypes.ErrTypeDoesNotExist.Is(err):
		code = pgcode.UndefinedObject
	default:
		code = pgcode.Internal
	}

	return &pgconn.PgError{
		Severity: string(ErrorResponseSeverity_Error),
		Code:     code.String(),
		Message:  err.Error(),
	}
}
