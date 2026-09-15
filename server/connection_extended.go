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
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/sirupsen/logrus"
	"net"
)

// handleParse handles a parse message, returning any error that occurs
func (h *ConnectionHandler) handleParse(message *pgproto3.Parse) error {
	h.waitForSync = true

	// TODO: "Named prepared statements must be explicitly closed before they can be redefined by another Parse message, but this is not required for the unnamed statement"
	queries, err := h.convertQuery(message.Query)
	if err != nil {
		if printErrorStackTraces {
			fmt.Printf("Error parsing query: %+v\n", err)
		}
		return err
	}
	if len(queries) != 1 {
		return errors.Errorf("cannot insert multiple commands into a prepared statement")
	}
	query := queries[0]

	if err = h.rejectStatementIfTransactionFailed(query); err != nil {
		return err
	}

	if query.AST == nil {
		// special case: empty query
		h.preparedStatements[message.Name] = PreparedStatementData{
			Query: query,
		}
		return nil
	}

	ctx, err := h.doltgresHandler.sm.NewContextWithQuery(context.Background(), h.mysqlConn, query.String)
	if err != nil {
		return err
	}
	parsedQuery, fields, err := h.doltgresHandler.ComPrepareParsed(ctx, h.mysqlConn, query.String, query.AST)
	if err != nil {
		return err
	}

	analyzedPlan, ok := parsedQuery.(sql.Node)
	if !ok {
		return errors.Errorf("expected a sql.Node, got %T", parsedQuery)
	}

	// A valid Parse message must have ParameterObjectIDs if there are any binding variables.
	bindVarTypes := message.ParameterOIDs

	// Clients can specify an OID of zero for a parameter, or omit trailing parameters from
	// ParameterOIDs entirely (the Postgres protocol allows specifying types for only a prefix
	// of the placeholders), to indicate that a parameter's type should be inferred. We always
	// compute the plan-inferred types (we can't know whether bindVarTypes is missing
	// trailing entries without first inspecting the analyzed plan) but only use an inferred
	// type to fill a position the client left unspecified. An explicit, non-zero OID from the
	// client must never be overwritten by an inferred type, since the client will encode that
	// parameter's Bind value using its own declared type.
	inferredTypes, err := extractBindVarTypes(ctx, analyzedPlan)
	if err != nil {
		return err
	}
	merged := make([]uint32, len(inferredTypes))
	copy(merged, inferredTypes)
	for i, oid := range bindVarTypes {
		if oid != 0 && i < len(merged) {
			merged[i] = oid
		}
	}
	bindVarTypes = merged

	h.preparedStatements[message.Name] = PreparedStatementData{
		Query:        query,
		ReturnFields: fields,
		BindVarTypes: bindVarTypes,
	}
	return h.send(&pgproto3.ParseComplete{})
}

// handleDescribe handles a Describe message, returning any error that occurs
func (h *ConnectionHandler) handleDescribe(message *pgproto3.Describe) error {
	var fields []pgproto3.FieldDescription
	var bindvarTypes []uint32
	var query ConvertedQuery

	h.waitForSync = true
	if message.ObjectType == 'S' {
		preparedStatementData, ok := h.preparedStatements[message.Name]
		if !ok {
			return errors.Errorf("prepared statement %s does not exist", message.Name)
		}

		fields = preparedStatementData.ReturnFields
		bindvarTypes = preparedStatementData.BindVarTypes
		query = preparedStatementData.Query
	} else {
		portalData, ok := h.portals[message.Name]
		if !ok {
			return errors.Errorf("portal %s does not exist", message.Name)
		}

		fields = portalData.Fields
		query = portalData.Query
	}

	return h.sendDescribeResponse(fields, bindvarTypes, query)
}

// handleBind handles a bind message, returning any error that occurs
func (h *ConnectionHandler) handleBind(message *pgproto3.Bind) error {
	h.waitForSync = true

	// TODO: a named portal object lasts till the end of the current transaction, unless explicitly destroyed
	//  we need to destroy the named portal as a side effect of the transaction ending
	logrus.Tracef("binding portal %q to prepared statement %s", message.DestinationPortal, message.PreparedStatement)
	preparedData, ok := h.preparedStatements[message.PreparedStatement]
	if !ok {
		return errors.Errorf("prepared statement %s does not exist", message.PreparedStatement)
	}

	if err := h.rejectStatementIfTransactionFailed(preparedData.Query); err != nil {
		return err
	}

	if preparedData.Query.AST == nil {
		// special case: empty query
		h.portals[message.DestinationPortal] = PortalData{
			Query:        preparedData.Query,
			IsEmptyQuery: true,
		}
		return h.send(&pgproto3.BindComplete{})
	}

	analyzedPlan, fields, err := h.doltgresHandler.ComBind(
		context.Background(),
		h.mysqlConn,
		preparedData.Query.String,
		preparedData.Query.AST,
		BindVariables{
			varTypes:    preparedData.BindVarTypes,
			formatCodes: message.ParameterFormatCodes,
			parameters:  message.Parameters,
		},
		message.ResultFormatCodes)
	if err != nil {
		return err
	}

	boundPlan, ok := analyzedPlan.(sql.Node)
	if !ok {
		return errors.Errorf("expected a sql.Node, got %T", analyzedPlan)
	}

	resultFormatCodes, err := extendFormatCodes(len(fields), message.ResultFormatCodes)
	if err != nil {
		return err
	}
	h.portals[message.DestinationPortal] = PortalData{
		Query:       preparedData.Query,
		Fields:      fields,
		BoundPlan:   boundPlan,
		FormatCodes: resultFormatCodes,
	}
	return h.send(&pgproto3.BindComplete{})
}

// handleExecute handles an execute message, returning any error that occurs
func (h *ConnectionHandler) handleExecute(message *pgproto3.Execute) error {
	h.waitForSync = true

	// TODO: implement the RowMax
	portalData, ok := h.portals[message.Portal]
	if !ok {
		return errors.Errorf("portal %s does not exist", message.Portal)
	}

	logrus.Tracef("executing portal %s with contents %v", message.Portal, portalData)
	query := portalData.Query

	if portalData.IsEmptyQuery {
		return h.send(&pgproto3.EmptyQueryResponse{})
	}

	if err := h.rejectStatementIfTransactionFailed(query); err != nil {
		return err
	}

	// Statements executed via the extended query protocol run in an implicit transaction block, which is
	// closed (committed on success, rolled back on error) by the next Sync message
	if err := h.startImplicitTransaction(query); err != nil {
		return err
	}

	// Certain statement types get handled directly by the handler instead of being passed to the engine
	handled, _, err := h.handleQueryOutsideEngine(query)
	if handled {
		return err
	}

	// |rowsAffected| gets altered by the callback below
	rowsAffected := int32(0)

	callback := h.spoolRowsCallback(query, &rowsAffected, true)
	err = h.doltgresHandler.ComExecuteBound(context.Background(), h.mysqlConn, query.String, portalData.BoundPlan, portalData.FormatCodes, callback)
	if err != nil {
		return err
	}

	return h.send(makeCommandComplete(query.StatementTag, rowsAffected))
}

// deallocatePreparedStatement handles a DEALLOCATE statement by deleting the corresponding prepared statement from the
// handler's prepared statement map, and sending a CommandComplete message back to the client. Pass an empty |name|
// for `ALL`. This matches the behavior in the parser, which doesn't include a separate field for ALL.
func (h *ConnectionHandler) deallocatePreparedStatement(name string, preparedStatements map[string]PreparedStatementData, query ConvertedQuery, conn net.Conn) error {
	if name == "" {
		for name := range preparedStatements {
			delete(preparedStatements, name)
		}
	} else {
		_, ok := preparedStatements[name]
		if !ok {
			return errors.Errorf("prepared statement %s does not exist", name)
		}
		delete(preparedStatements, name)
	}

	return h.send(&pgproto3.CommandComplete{
		CommandTag: []byte(query.StatementTag),
	})
}

// sendDescribeResponse sends a response message for a Describe message
func (h *ConnectionHandler) sendDescribeResponse(fields []pgproto3.FieldDescription, types []uint32, query ConvertedQuery) error {
	// The prepared statement variant of the describe command returns the OIDs of the parameters.
	if types != nil {
		if err := h.send(&pgproto3.ParameterDescription{
			ParameterOIDs: types,
		}); err != nil {
			return err
		}
	}

	if returnsRow(query) {
		// Both variants finish with a row description.
		return h.send(&pgproto3.RowDescription{
			Fields: fields,
		})
	} else {
		return h.send(&pgproto3.NoData{})
	}
}
