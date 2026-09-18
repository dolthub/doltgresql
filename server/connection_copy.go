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

// copyContinuationKind identifies which frontend protocol operation resumes after COPY finishes.
type copyContinuationKind byte

const (
	invalidCopyContinuation copyContinuationKind = iota
	simpleQueryCopyContinuation
	extendedQueryCopyContinuation
)

// copyContinuation records the frontend protocol operation that resumes after COPY finishes.
type copyContinuation struct {
	kind      copyContinuationKind
	execution *simpleQueryExecution
}

// newSimpleQueryCopyContinuation returns a continuation for the remaining statements in a simple Query message.
func newSimpleQueryCopyContinuation(execution *simpleQueryExecution) copyContinuation {
	return copyContinuation{kind: simpleQueryCopyContinuation, execution: execution}
}

// newExtendedQueryCopyContinuation returns a continuation for the extended-query batch that initiated COPY.
func newExtendedQueryCopyContinuation() copyContinuation {
	return copyContinuation{kind: extendedQueryCopyContinuation}
}

// transactionOwnership records whether COPY created the engine transaction it is using.
type transactionOwnership byte

const (
	borrowedTransaction transactionOwnership = iota
	copyOwnedTransaction
)

// copyInState owns all transfer and continuation data for one COPY FROM STDIN operation.
type copyInState struct {
	copyFromStdinNode *node.CopyFrom
	insertNode        sql.Node
	dataLoader        dataloader.DataLoader
	transaction       transactionOwnership
	continuation      copyContinuation
}

// newCopyInState returns the initial runtime state for a COPY FROM STDIN operation.
func newCopyInState(copyFrom *node.CopyFrom, continuation copyContinuation) *copyInState {
	return &copyInState{copyFromStdinNode: copyFrom, continuation: continuation}
}

// handleCopyMessage enforces the restricted frontend message set accepted during COPY FROM STDIN.
func (h *ConnectionHandler) handleCopyMessage(copyState *copyInState, message pgproto3.Message) messageResult {
	switch message := message.(type) {
	case *pgproto3.CopyData:
		if err := h.handleCopyDataHelper(copyState, bytes.NewReader(message.Data)); err != nil {
			return h.finishCopy(copyState, err)
		}
		return continueResult()
	case *pgproto3.CopyDone:
		return h.handleCopyDone(copyState)
	case *pgproto3.CopyFail:
		return h.finishCopy(copyState, pgerror.Newf(pgcode.QueryCanceled, "COPY from stdin failed: %s", message.Message))
	case *pgproto3.Flush, *pgproto3.Sync:
		// PostgreSQL ignores Flush and Sync while COPY owns the connection.
		return continueResult()
	default:
		return h.closeForCopyProtocolViolation(copyState)
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
func (h *ConnectionHandler) handleCopyFromStdinQuery(copyFrom *node.CopyFrom, continuation copyContinuation) error {
	if !h.state.beginCopyMode(newCopyInState(copyFrom, continuation)) {
		return errors.New("cannot begin COPY FROM STDIN with invalid state")
	}
	return h.send(&pgproto3.CopyInResponse{
		OverallFormat: 0,
	})
}

// handleCopyDone finalizes an in-progress COPY, including a transfer containing no CopyData messages.
func (h *ConnectionHandler) handleCopyDone(copyState *copyInState) messageResult {
	if copyState.dataLoader == nil {
		if err := h.handleCopyDataHelper(copyState, bytes.NewReader(nil)); err != nil {
			return h.finishCopy(copyState, err)
		}
	}

	sqlCtx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
	if err != nil {
		return h.finishCopy(copyState, err)
	}

	loadDataResults, err := copyState.dataLoader.Finish(sqlCtx)
	if err != nil {
		return h.finishCopy(copyState, err)
	}

	if copyState.transaction == copyOwnedTransaction {
		txSession, ok := sqlCtx.Session.(sql.TransactionSession)
		if !ok {
			return h.finishCopy(copyState, errors.Errorf("session does not implement sql.TransactionSession"))
		}
		if err = txSession.CommitTransaction(sqlCtx, txSession.GetTransaction()); err != nil {
			return h.finishCopy(copyState, err)
		}
		sqlCtx.SetIgnoreAutoCommit(false)
	}

	if err = h.send(&pgproto3.CommandComplete{
		CommandTag: []byte(fmt.Sprintf("COPY %d", loadDataResults.RowsLoaded)),
	}); err != nil {
		return h.finishCopy(copyState, err)
	}
	return h.finishCopy(copyState, nil)
}

// handleCopyDataHelper initializes a COPY transfer as needed and loads one input chunk.
func (h *ConnectionHandler) handleCopyDataHelper(copyState *copyInState, copyFromData io.Reader) (err error) {
	if copyState == nil {
		return errors.Errorf("COPY DATA message received without a COPY FROM STDIN operation in progress")
	}

	// Grab a sql.Context and ensure the session has a transaction started, otherwise the copied data
	// won't get committed correctly.
	sqlCtx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "COPY FROM STDIN")
	if err != nil {
		return err
	}
	if sqlCtx.GetTransaction() == nil && h.state.txState == idleTransactionState {
		copyState.transaction = copyOwnedTransaction
	}
	if h.state.txState != idleTransactionState {
		sqlCtx.SetIgnoreAutoCommit(true)
	}
	if err = startTransactionIfNecessary(sqlCtx); err != nil {
		return err
	}

	if copyState.copyFromStdinNode.TableName.Schema != "" {
		originalSchema, err := sqlCtx.GetSessionVariable(sqlCtx, "search_path")
		if err != nil {
			return err
		}
		err = sqlCtx.SetSessionVariable(sqlCtx, "search_path", copyState.copyFromStdinNode.TableName.Schema)
		if err != nil {
			return err
		}
		defer func() {
			_ = sqlCtx.SetSessionVariable(sqlCtx, "search_path", originalSchema)
		}()
	}

	dataLoader := copyState.dataLoader
	if dataLoader == nil {
		copyFromStdinNode := copyState.copyFromStdinNode
		if copyFromStdinNode == nil {
			return errors.Errorf("no COPY FROM STDIN node found")
		}

		// we build an insert node to use for the full insert plan, for which the copy from node will be the row source
		builder := planbuilder.New(sqlCtx, h.doltgresHandler.e.Analyzer.Catalog, nil)
		node, flags, err := builder.BindOnly(copyFromStdinNode.InsertStub, "", nil)
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

		switch copyFromStdinNode.CopyOptions.CopyFormat {
		case tree.CopyFormatText:
			dataLoader, err = dataloader.NewTabularDataLoader(insertNode.ColumnNames, tbl.Schema(sqlCtx), copyFromStdinNode.CopyOptions.Delimiter, "", copyFromStdinNode.CopyOptions.Header)
		case tree.CopyFormatCsv:
			dataLoader, err = dataloader.NewCsvDataLoader(insertNode.ColumnNames, tbl.Schema(sqlCtx), copyFromStdinNode.CopyOptions.Delimiter, copyFromStdinNode.CopyOptions.Header)
		case tree.CopyFormatBinary:
			dataLoader, err = dataloader.NewBinaryDataLoader(insertNode.ColumnNames, tbl.Schema(sqlCtx))
		default:
			err = errors.Errorf("unknown format specified for COPY FROM: %v",
				copyFromStdinNode.CopyOptions.CopyFormat)
		}

		if err != nil {
			return err
		}

		// we have to set the data loader on the copyFrom node before we analyze it, because we need the loader's
		// schema to analyze
		copyState.copyFromStdinNode.DataLoader = dataLoader

		// After building out stub insert node, swap out the source node with the COPY node, then analyze the entire thing
		node = insertNode.WithSource(copyFromStdinNode)
		analyzedNode, err := h.doltgresHandler.e.Analyzer.Analyze(sqlCtx, node, nil, flags)
		if err != nil {
			return err
		}

		copyState.insertNode = analyzedNode
		copyState.dataLoader = dataLoader
	}

	reader := bufio.NewReader(copyFromData)
	if err = dataLoader.SetNextDataChunk(sqlCtx, reader); err != nil {
		return err
	}

	callback := func(_ *sql.Context, _ *Result) error { return nil }
	err = h.doltgresHandler.ComExecuteBound(sqlCtx, h.mysqlConn, "COPY FROM", copyState.insertNode, nil, callback)
	if err != nil {
		return err
	}
	return nil
}

// finishCopy releases COPY state, rolls back failed work, and resumes the initiating protocol operation.
func (h *ConnectionHandler) finishCopy(copyState *copyInState, copyErr error) messageResult {
	if copyState == nil || !h.state.finishCopyMode(copyState) {
		h.state.closeConnection()
		return messageResult{
			action: closeConnection,
			err:    errors.New("cannot finish inactive COPY FROM STDIN operation"),
		}
	}
	continuation := copyState.continuation
	if copyErr != nil {
		h.rollbackCopyTransaction(copyState.transaction)
	}
	switch continuation.kind {
	case simpleQueryCopyContinuation:
		if continuation.execution == nil {
			h.state.closeConnection()
			return messageResult{
				action: closeConnection,
				err:    errors.New("simple-query COPY continuation has no execution"),
			}
		}
		if copyErr != nil {
			return readyResult(copyErr)
		}
		complete, err := h.resumeSimpleQuery(continuation.execution)
		return messageResultForCompletion(complete, err)
	case extendedQueryCopyContinuation:
		if copyErr != nil {
			h.state.discardUntilSync()
			return messageResult{err: copyErr}
		}
		h.state.beginExtendedQueryMode()
		return continueResult()
	default:
		h.state.closeConnection()
		return messageResult{
			action: closeConnection,
			err:    errors.New("COPY FROM STDIN has an invalid protocol continuation"),
		}
	}
}

// closeForCopyProtocolViolation reports COPY protocol desynchronization and terminates the connection.
func (h *ConnectionHandler) closeForCopyProtocolViolation(copyState *copyInState) messageResult {
	h.rollbackCopyTransaction(copyState.transaction)
	h.failActiveTransaction()
	h.state.closeConnection()
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

// copyFromFileQuery handles a COPY FROM message that is reading from a file, returning any error that occurs
func (h *ConnectionHandler) copyFromFileQuery(stmt *node.CopyFrom) error {
	copyState := &copyInState{
		copyFromStdinNode: stmt,
	}

	// TODO: security check for file path
	// TODO: Privilege Checking: https://www.postgresql.org/docs/15/sql-copy.html
	f, err := os.Open(stmt.File)
	if err != nil {
		return err
	}
	defer f.Close()

	err = h.handleCopyDataHelper(copyState, f)
	if err != nil {
		return err
	}

	sqlCtx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
	if err != nil {
		return err
	}

	loadDataResults, err := copyState.dataLoader.Finish(sqlCtx)
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

// rollbackCopyTransaction rolls back the transaction started on behalf of a failed or aborted COPY FROM STDIN
// operation, so that rows loaded by any chunks that were processed successfully don't linger in an open
// transaction that a later statement would commit. If the COPY did not start the transaction itself (e.g. it was
// run inside a transaction block), this does nothing: the normal statement-failure handling takes care of it,
// matching Postgres.
func (h *ConnectionHandler) rollbackCopyTransaction(ownership transactionOwnership) {
	if ownership != copyOwnedTransaction || h.state.txState != idleTransactionState {
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
