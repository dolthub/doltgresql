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
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/doltgresql/core/dataloader"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/node"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"
	"github.com/dolthub/go-mysql-server/sql/transform"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"os"
)

// rollbackCopyTransaction rolls back the transaction started on behalf of a failed or aborted COPY FROM STDIN
// operation, so that rows loaded by any chunks that were processed successfully don't linger in an open
// transaction that a later statement would commit. If the COPY did not start the transaction itself (e.g. it was
// run inside a transaction block), this does nothing: the normal statement-failure handling takes care of it,
// matching Postgres.
func (h *ConnectionHandler) rollbackCopyTransaction(startedTransaction bool) {
	if !startedTransaction || h.transactionState != idleTransactionState {
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

// copyFromFileQuery handles a COPY FROM message that is reading from a file, returning any error that occurs
func (h *ConnectionHandler) copyFromFileQuery(stmt *node.CopyFrom) error {
	copyState := &copyFromStdinState{
		copyFromStdinNode: stmt,
	}

	// TODO: security check for file path
	// TODO: Privilege Checking: https://www.postgresql.org/docs/15/sql-copy.html
	f, err := os.Open(stmt.File)
	if err != nil {
		return err
	}
	defer f.Close()

	_, _, err = h.handleCopyDataHelper(copyState, f)
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

// handleCopyDataHelper is a helper function that should only be invoked by handleCopyData. handleCopyData wraps this
// function so that it can capture any returned error message and store it in the saved state.
func (h *ConnectionHandler) handleCopyDataHelper(copyState *copyFromStdinState, copyFromData io.Reader) (stop bool, endOfMessages bool, err error) {
	if copyState == nil {
		return false, true, errors.Errorf("COPY DATA message received without a COPY FROM STDIN operation in progress")
	}

	// Grab a sql.Context and ensure the session has a transaction started, otherwise the copied data
	// won't get committed correctly.
	sqlCtx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "COPY FROM STDIN")
	if err != nil {
		return false, false, err
	}
	if sqlCtx.GetTransaction() == nil && h.transactionState == idleTransactionState {
		copyState.startedTransaction = true
	}
	if h.transactionState != idleTransactionState {
		sqlCtx.SetIgnoreAutoCommit(true)
	}
	if err = startTransactionIfNecessary(sqlCtx); err != nil {
		return false, false, err
	}

	if copyState.copyFromStdinNode.TableName.Schema != "" {
		originalSchema, err := sqlCtx.GetSessionVariable(sqlCtx, "search_path")
		if err != nil {
			return false, false, err
		}
		err = sqlCtx.SetSessionVariable(sqlCtx, "search_path", copyState.copyFromStdinNode.TableName.Schema)
		if err != nil {
			return false, false, err
		}
		defer func() {
			_ = sqlCtx.SetSessionVariable(sqlCtx, "search_path", originalSchema)
		}()
	}

	dataLoader := copyState.dataLoader
	if dataLoader == nil {
		copyFromStdinNode := copyState.copyFromStdinNode
		if copyFromStdinNode == nil {
			return false, false, errors.Errorf("no COPY FROM STDIN node found")
		}

		// we build an insert node to use for the full insert plan, for which the copy from node will be the row source
		builder := planbuilder.New(sqlCtx, h.doltgresHandler.e.Analyzer.Catalog, nil)
		node, flags, err := builder.BindOnly(copyFromStdinNode.InsertStub, "", nil)
		if err != nil {
			return false, false, err
		}

		insertNode, ok := node.(*plan.InsertInto)
		if !ok {
			return false, false, errors.Errorf("expected plan.InsertInto, got %T", node)
		}

		// now that we have our insert node, we can build the data loader
		tbl := getInsertableTable(insertNode.Destination)
		if tbl == nil {
			// this should be impossible, enforced by analyzer above
			return false, false, errors.Errorf("no insertable table found in %v", insertNode.Destination)
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
			return false, false, err
		}

		// we have to set the data loader on the copyFrom node before we analyze it, because we need the loader's
		// schema to analyze
		copyState.copyFromStdinNode.DataLoader = dataLoader

		// After building out stub insert node, swap out the source node with the COPY node, then analyze the entire thing
		node = insertNode.WithSource(copyFromStdinNode)
		analyzedNode, err := h.doltgresHandler.e.Analyzer.Analyze(sqlCtx, node, nil, flags)
		if err != nil {
			return false, false, err
		}

		copyState.insertNode = analyzedNode
		copyState.dataLoader = dataLoader
	}

	reader := bufio.NewReader(copyFromData)
	if err = dataLoader.SetNextDataChunk(sqlCtx, reader); err != nil {
		return false, false, err
	}

	callback := func(_ *sql.Context, _ *Result) error { return nil }
	err = h.doltgresHandler.ComExecuteBound(sqlCtx, h.mysqlConn, "COPY FROM", copyState.insertNode, nil, callback)
	if err != nil {
		return false, false, err
	}

	// We expect to see more CopyData messages until we see either a CopyDone or CopyFail message, so
	// return false for endOfMessages
	return false, false, nil
}

// Returns the first sql.InsertableTable node found in the tree provided, or nil if none is found.
func getInsertableTable(node sql.Node) sql.InsertableTable {
	var tbl sql.InsertableTable
	transform.Inspect(node, func(node sql.Node) bool {
		if rt, ok := node.(*plan.ResolvedTable); ok {
			if insertable, ok := rt.Table.(sql.InsertableTable); ok {
				tbl = insertable
				return false
			}
		}
		return true
	})

	return tbl
}

// handleCopyDone handles a COPY DONE message by finalizing the in-progress COPY DATA operation. A transaction started
// solely for this COPY is committed here; an enclosing transaction remains open. The |stop| response parameter is
// true if the connection handler should shut down the connection,
// |endOfMessages| is true if no more COPY DATA messages are expected, and the server should tell the client that it is
// ready for the next query, and |err| contains any error that occurred while processing the COPY DATA message.
func (h *ConnectionHandler) handleCopyDone(_ *pgproto3.CopyDone) (stop bool, endOfMessages bool, err error) {
	if h.copyFromStdinState == nil {
		return false, true,
			errors.Errorf("COPY DONE message received without a COPY FROM STDIN operation in progress")
	}

	// The COPY DONE message ends the COPY operation whether it succeeds or fails below, so always clear the COPY
	// state, leaving the connection ready for its next query. If finalizing the operation fails, its work must
	// also be rolled back.
	startedTransaction := h.copyFromStdinState.startedTransaction
	defer func() {
		h.copyFromStdinState = nil
		if err != nil {
			h.rollbackCopyTransaction(startedTransaction)
		}
	}()

	// If there was a previous error returned from processing a CopyData message, then don't return an error here
	// and don't send endOfMessage=true, since the CopyData error already sent endOfMessage=true. If we do send
	// endOfMessage=true here, then the client gets confused about the unexpected/extra Idle message since the
	// server has already reported it was idle in the last message after the returned error.
	if h.copyFromStdinState.copyErr != nil {
		return false, false, nil
	}

	dataLoader := h.copyFromStdinState.dataLoader
	if dataLoader == nil {
		return false, true,
			errors.Errorf("no data loader found for COPY FROM STDIN operation")
	}

	sqlCtx, err := h.doltgresHandler.NewContext(context.Background(), h.mysqlConn, "")
	if err != nil {
		return false, false, err
	}

	loadDataResults, err := dataLoader.Finish(sqlCtx)
	if err != nil {
		return false, false, err
	}

	if startedTransaction {
		txSession, ok := sqlCtx.Session.(sql.TransactionSession)
		if !ok {
			return false, false, errors.Errorf("session does not implement sql.TransactionSession")
		}
		if err = txSession.CommitTransaction(sqlCtx, txSession.GetTransaction()); err != nil {
			return false, false, err
		}
		sqlCtx.SetIgnoreAutoCommit(false)
	}

	// We send back endOfMessage=true, since the COPY DONE message ends the COPY DATA flow and the server is ready
	// to accept the next query now.
	return false, true, h.send(&pgproto3.CommandComplete{
		CommandTag: []byte(fmt.Sprintf("COPY %d", loadDataResults.RowsLoaded)),
	})
}

// handleCopyFromStdinQuery handles the COPY FROM STDIN query at the Doltgres layer, without passing it to the engine.
// COPY FROM STDIN can't be handled directly by the GMS engine, since COPY FROM STDIN relies on multiple messages sent
// over the wire.
func (h *ConnectionHandler) handleCopyFromStdinQuery(copyFrom *node.CopyFrom, conn net.Conn) error {
	h.copyFromStdinState = &copyFromStdinState{
		copyFromStdinNode: copyFrom,
	}

	return h.send(&pgproto3.CopyInResponse{
		OverallFormat: 0,
	})
}
