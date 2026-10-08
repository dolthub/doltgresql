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
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/core/dataloader"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/node"
)

// transactionOwnership records whether COPY created the engine transaction it is using.
type transactionOwnership byte

const (
	borrowedTransaction transactionOwnership = iota
	copyOwnedTransaction
)

// copyFromState owns the execution state shared by all COPY FROM operations.
type copyFromState struct {
	statement            *node.CopyFrom
	insertPlan           sql.Node
	dataLoader           dataloader.DataLoader
	txOwnership          transactionOwnership
	suspendedSimpleQuery *simpleQueryExecution
}

// newCopyFromState returns the initial execution state for a COPY FROM statement.
func newCopyFromState(statement *node.CopyFrom, suspendedSimpleQuery *simpleQueryExecution) *copyFromState {
	return &copyFromState{
		statement:            statement,
		suspendedSimpleQuery: suspendedSimpleQuery,
	}
}

// handleCopyInMessage enforces the restricted frontend message set accepted during copy-in mode.
func (h *ConnectionHandler) handleCopyInMessage(copyFrom *copyFromState, message pgproto3.Message) messageResult {
	switch message := message.(type) {
	case *pgproto3.CopyData:
		if err := h.loadCopyFromData(copyFrom, bytes.NewReader(message.Data)); err != nil {
			return h.finishCopyIn(copyFrom, err)
		}
		return continueResult()
	case *pgproto3.CopyDone:
		return h.handleCopyInDone(copyFrom)
	case *pgproto3.CopyFail:
		return h.finishCopyIn(copyFrom, pgerror.Newf(pgcode.QueryCanceled, "COPY from stdin failed: %s", message.Message))
	case *pgproto3.Flush, *pgproto3.Sync:
		// PostgreSQL ignores Flush and Sync while COPY owns the connection.
		return continueResult()
	default:
		return h.closeForCopyInProtocolViolation(copyFrom)
	}
}

// handleCopyTo handles a COPY ... TO STDOUT statement, streaming the results of the underlying SELECT statement
// back to the client as COPY DATA messages. Copying to a server-side file is rejected during AST conversion, as a
// security measure. Returns any error that occurs.
func (h *ConnectionHandler) handleCopyTo(copyTo *node.CopyTo) (err error) {
	sqlCtx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "COPY TO")
	if err != nil {
		return err
	}

	if copyTo.TableName.Schema != "" {
		originalSchema, err := sqlCtx.GetSessionVariable(sqlCtx, "search_path")
		if err != nil {
			return err
		}
		err = sqlCtx.SetSessionVariable(sqlCtx, "search_path", copyTo.TableName.Schema)
		if err != nil {
			return err
		}
		defer func() {
			_ = sqlCtx.SetSessionVariable(sqlCtx, "search_path", originalSchema)
		}()
	}

	// Build and analyze the plan for the SELECT statement underlying this COPY TO statement.
	builder := planbuilder.New(sqlCtx, h.doltgresHandler.e.Analyzer.Catalog, nil)
	boundNode, flags, err := builder.BindOnly(copyTo.SelectStub, "", nil)
	if err != nil {
		return err
	}
	analyzedNode, err := h.doltgresHandler.e.Analyzer.Analyze(sqlCtx, boundNode, nil, flags)
	if err != nil {
		return err
	}

	sch := analyzedNode.Schema(sqlCtx)
	colNames := make([]string, len(sch))
	for i, col := range sch {
		colNames[i] = col.Name
	}

	// For the binary format, we ask the engine for values in each type's binary send format by setting the format
	// code for every column to 1 (binary). The text and CSV formats use the default text encoding (format code 0).
	var dataWriter dataloader.DataWriter
	var formatCodes []int16
	switch copyTo.CopyOptions.CopyFormat {
	case tree.CopyFormatText:
		dataWriter = dataloader.NewTabularDataWriter(colNames, copyTo.CopyOptions.Delimiter, "", copyTo.CopyOptions.Header)
	case tree.CopyFormatCsv:
		dataWriter = dataloader.NewCsvDataWriter(colNames, copyTo.CopyOptions.Delimiter, copyTo.CopyOptions.Header)
	case tree.CopyFormatBinary:
		dataWriter = dataloader.NewBinaryDataWriter()
		formatCodes = make([]int16, len(sch))
		for i := range formatCodes {
			formatCodes[i] = 1
		}
	default:
		return errors.Errorf("unknown format specified for COPY TO: %v", copyTo.CopyOptions.CopyFormat)
	}

	var overallFormat byte
	columnFormatCodes := make([]uint16, len(sch))
	if copyTo.CopyOptions.CopyFormat == tree.CopyFormatBinary {
		overallFormat = 1
		for i := range columnFormatCodes {
			columnFormatCodes[i] = 1
		}
	}
	if err = h.send(&pgproto3.CopyOutResponse{
		OverallFormat:     overallFormat,
		ColumnFormatCodes: columnFormatCodes,
	}); err != nil {
		return err
	}

	// The header is not sent on its own, but is instead prepended to the next chunk of data written (the first row,
	// or the footer when there are no rows). This matches Postgres's message framing for COPY TO STDOUT, which some
	// clients rely on
	pendingHeader, err := dataWriter.WriteHeader()
	if err != nil {
		return err
	}
	writeChunk := func(data []byte) {
		if pendingHeader != nil {
			data = append(pendingHeader, data...)
			pendingHeader = nil
		}
		h.backend.Send(&pgproto3.CopyData{Data: data})
	}

	numRows := 0
	callback := func(_ *sql.Context, res *Result) error {
		for _, row := range res.Rows {
			data, wErr := dataWriter.WriteRow(row.val)
			if wErr != nil {
				return wErr
			}
			writeChunk(data)
		}
		numRows += len(res.Rows)
		return h.backend.Flush()
	}

	if err = h.doltgresHandler.ComExecuteBound(sqlCtx, h.mysqlConn, "COPY TO", analyzedNode, formatCodes, callback); err != nil {
		return err
	}

	footerData, err := dataWriter.WriteFooter()
	if err != nil {
		return err
	}
	if footerData != nil {
		writeChunk(footerData)
	}

	// If nothing was written at all (no rows and no footer), the header must still be sent on its own.
	if pendingHeader != nil {
		h.backend.Send(&pgproto3.CopyData{Data: pendingHeader})
	}

	if err = h.backend.Flush(); err != nil {
		return err
	}

	h.backend.Send(&pgproto3.CopyDone{})
	return h.send(makeCommandComplete("COPY", int32(numRows)))
}

// handleCopyFromStdinQuery handles the COPY FROM STDIN query at the Doltgres layer, without passing it to the engine.
// COPY FROM STDIN can't be handled directly by the GMS engine, since COPY FROM STDIN relies on multiple messages sent
// over the wire.
func (h *ConnectionHandler) handleCopyFromStdinQuery(statement *node.CopyFrom, simpleQuery *simpleQueryExecution) error {
	if !h.state.beginCopyInMode(newCopyFromState(statement, simpleQuery)) {
		return errors.New("cannot begin COPY FROM STDIN with invalid state")
	}
	return h.send(&pgproto3.CopyInResponse{
		OverallFormat: 0,
	})
}

// handleCopyInDone finalizes an in-progress copy-in operation, including one with no CopyData messages.
func (h *ConnectionHandler) handleCopyInDone(copyFrom *copyFromState) messageResult {
	if copyFrom.dataLoader == nil {
		if err := h.loadCopyFromData(copyFrom, bytes.NewReader(nil)); err != nil {
			return h.finishCopyIn(copyFrom, err)
		}
	}

	sqlCtx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
	if err != nil {
		return h.finishCopyIn(copyFrom, err)
	}

	loadDataResults, err := copyFrom.dataLoader.Finish(sqlCtx)
	if err != nil {
		return h.finishCopyIn(copyFrom, err)
	}

	if copyFrom.txOwnership == copyOwnedTransaction {
		txSession, ok := sqlCtx.Session.(sql.TransactionSession)
		if !ok {
			return h.finishCopyIn(copyFrom, errors.Errorf("session does not implement sql.TransactionSession"))
		}
		if err = txSession.CommitTransaction(sqlCtx, txSession.GetTransaction()); err != nil {
			return h.finishCopyIn(copyFrom, err)
		}
		sqlCtx.SetIgnoreAutoCommit(false)
	}

	if err = h.send(&pgproto3.CommandComplete{
		CommandTag: []byte(fmt.Sprintf("COPY %d", loadDataResults.RowsLoaded)),
	}); err != nil {
		return h.finishCopyIn(copyFrom, err)
	}
	return h.finishCopyIn(copyFrom, nil)
}

// loadCopyFromData initializes a COPY FROM execution as needed and loads data from one input reader.
func (h *ConnectionHandler) loadCopyFromData(state *copyFromState, data io.Reader) (err error) {
	if state == nil {
		return errors.New("COPY FROM execution state is missing")
	}
	if state.statement == nil {
		return errors.New("COPY FROM statement is missing")
	}

	// Grab a sql.Context and ensure the session has a transaction started, otherwise the copied data
	// won't get committed correctly.
	sqlCtx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "COPY FROM STDIN")
	if err != nil {
		return err
	}
	if sqlCtx.GetTransaction() == nil && h.state.txState == idleTransactionState {
		state.txOwnership = copyOwnedTransaction
	}
	if h.state.txState != idleTransactionState {
		sqlCtx.SetIgnoreAutoCommit(true)
	}
	if err = startTransactionIfNecessary(sqlCtx); err != nil {
		return err
	}

	if state.statement.TableName.Schema != "" {
		originalSchema, err := sqlCtx.GetSessionVariable(sqlCtx, "search_path")
		if err != nil {
			return err
		}
		err = sqlCtx.SetSessionVariable(sqlCtx, "search_path", state.statement.TableName.Schema)
		if err != nil {
			return err
		}
		defer func() {
			_ = sqlCtx.SetSessionVariable(sqlCtx, "search_path", originalSchema)
		}()
	}

	dataLoader := state.dataLoader
	if dataLoader == nil {
		statement := state.statement

		// we build an insert node to use for the full insert plan, for which the copy from node will be the row source
		builder := planbuilder.New(sqlCtx, h.doltgresHandler.e.Analyzer.Catalog, nil)
		node, flags, err := builder.BindOnly(statement.InsertStub, "", nil)
		if err != nil {
			return err
		}

		insertNode, ok := node.(*plan.InsertInto)
		if !ok {
			return errors.Errorf("expected plan.InsertInto, got %T", node)
		}

		// now that we have our insert node, we can build the data loader
		tbl, err := plan.GetInsertable(insertNode.Destination)
		if err != nil {
			return errors.Wrap(err, "COPY destination is not insertable")
		}

		switch statement.CopyOptions.CopyFormat {
		case tree.CopyFormatText:
			dataLoader, err = dataloader.NewTabularDataLoader(insertNode.ColumnNames, tbl.Schema(sqlCtx), statement.CopyOptions.Delimiter, "", statement.CopyOptions.Header)
		case tree.CopyFormatCsv:
			dataLoader, err = dataloader.NewCsvDataLoader(insertNode.ColumnNames, tbl.Schema(sqlCtx), statement.CopyOptions.Delimiter, statement.CopyOptions.Header)
		case tree.CopyFormatBinary:
			dataLoader, err = dataloader.NewBinaryDataLoader(insertNode.ColumnNames, tbl.Schema(sqlCtx))
		default:
			err = errors.Errorf("unknown format specified for COPY FROM: %v",
				statement.CopyOptions.CopyFormat)
		}

		if err != nil {
			return err
		}

		// we have to set the data loader on the copyFrom node before we analyze it, because we need the loader's
		// schema to analyze
		statement.DataLoader = dataLoader

		// After building out stub insert node, swap out the source node with the COPY node, then analyze the entire thing
		node = insertNode.WithSource(statement)
		analyzedNode, err := h.doltgresHandler.e.Analyzer.Analyze(sqlCtx, node, nil, flags)
		if err != nil {
			return err
		}

		state.insertPlan = analyzedNode
		state.dataLoader = dataLoader
	}

	reader := bufio.NewReader(data)
	if err = dataLoader.SetNextDataChunk(sqlCtx, reader); err != nil {
		return err
	}

	callback := func(_ *sql.Context, _ *Result) error { return nil }
	err = h.doltgresHandler.ComExecuteBound(sqlCtx, h.mysqlConn, "COPY FROM", state.insertPlan, nil, callback)
	if err != nil {
		return err
	}
	return nil
}

// finishCopyIn releases copy-in state, rolls back failed work, and resumes the initiating protocol operation.
func (h *ConnectionHandler) finishCopyIn(copyFrom *copyFromState, copyErr error) messageResult {
	if copyFrom == nil || !h.state.finishCopyInMode(copyFrom) {
		h.state.beginCloseConnectionMode()
		return messageResult{
			action: closeConnection,
			err:    errors.New("cannot finish inactive COPY FROM STDIN operation"),
		}
	}
	if copyErr != nil {
		h.rollbackCopyInTransaction(copyFrom.txOwnership)
	}
	if copyFrom.suspendedSimpleQuery != nil {
		if copyErr != nil {
			return readyResult(copyErr)
		}
		complete, err := h.resumeSimpleQuery(copyFrom.suspendedSimpleQuery)
		return messageResultForCompletion(complete, err)
	}
	if copyErr != nil {
		h.state.beginDiscardUntilSyncMode()
		return messageResult{err: copyErr}
	}
	h.state.beginExtendedQueryMode()
	return continueResult()
}

// closeForCopyInProtocolViolation reports copy-in protocol desynchronization and terminates the connection.
func (h *ConnectionHandler) closeForCopyInProtocolViolation(copyFrom *copyFromState) messageResult {
	h.rollbackCopyInTransaction(copyFrom.txOwnership)
	h.failActiveTransaction()
	h.state.beginCloseConnectionMode()
	responses := []*pgproto3.ErrorResponse{
		{
			Severity: string(ErrorResponseSeverity_Error),
			Code:     pgcode.ProtocolViolation.String(),
			Message:  "unexpected message during COPY from stdin",
		},
		{
			Severity: string(ErrorResponseSeverity_Fatal),
			Code:     pgcode.ProtocolViolation.String(),
			Message:  "terminating connection because protocol synchronization was lost",
		},
	}
	for _, response := range responses {
		if sendErr := h.send(response); sendErr != nil {
			return closeResult()
		}
	}
	return closeResult()
}

// handleCopyFromFileQuery handles a COPY FROM message that is reading from a file, returning any error that occurs
func (h *ConnectionHandler) handleCopyFromFileQuery(stmt *node.CopyFrom) error {
	state := newCopyFromState(stmt, nil)

	// TODO: security check for file path
	// TODO: Privilege Checking: https://www.postgresql.org/docs/15/sql-copy.html
	f, err := os.Open(stmt.File)
	if err != nil {
		return err
	}
	defer f.Close()

	err = h.loadCopyFromData(state, f)
	if err != nil {
		return err
	}

	sqlCtx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
	if err != nil {
		return err
	}

	loadDataResults, err := state.dataLoader.Finish(sqlCtx)
	if err != nil {
		return err
	}

	if sqlCtx.GetTransaction() != nil && sqlCtx.GetIgnoreAutoCommit() {
		txSession, ok := sqlCtx.Session.(sql.TransactionSession)
		if !ok {
			return errors.Errorf("session does not implement sql.TransactionSession")
		}
		if err = txSession.CommitTransaction(sqlCtx, txSession.GetTransaction()); err != nil {
			return err
		}
		sqlCtx.SetIgnoreAutoCommit(false)
	}

	return h.send(&pgproto3.CommandComplete{
		CommandTag: []byte(fmt.Sprintf("COPY %d", loadDataResults.RowsLoaded)),
	})
}

// rollbackCopyInTransaction rolls back the transaction started on behalf of a failed or aborted COPY FROM STDIN
// operation, so that rows loaded by any chunks that were processed successfully don't linger in an open
// transaction that a later statement would commit. If the COPY did not start the transaction itself (e.g. it was
// run inside a transaction block), this does nothing: the normal statement-failure handling takes care of it,
// matching Postgres.
func (h *ConnectionHandler) rollbackCopyInTransaction(txOwnership transactionOwnership) {
	if txOwnership != copyOwnedTransaction || h.state.txState != idleTransactionState {
		return
	}
	if h.restoredAutoCommitWithoutTransaction() {
		return
	}
	if err := h.runEngineTransactionControl("ROLLBACK"); err != nil {
		logrus.Warnf("error rolling back COPY FROM STDIN transaction: %s", err)
	}
	h.restoredAutoCommitWithoutTransaction()
}
