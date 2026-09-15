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
)

// portalData records one bound portal belonging to a connection.
type portalData struct {
	Query        ConvertedQuery
	IsEmptyQuery bool
	Fields       []pgproto3.FieldDescription
	BoundPlan    sql.Node
	FormatCodes  []int16
}

// preparedStatementData records one parsed prepared statement belonging to a connection.
type preparedStatementData struct {
	Query        ConvertedQuery
	ReturnFields []pgproto3.FieldDescription
	BindVarTypes []uint32
}

// extendedQueryState owns the prepared statements and portals belonging to one connection.
type extendedQueryState struct {
	preparedStatements map[string]preparedStatementData
	portals            map[string]portalData
}

// newExtendedQueryState returns empty extended-query state for a new connection.
func newExtendedQueryState() extendedQueryState {
	return extendedQueryState{
		preparedStatements: make(map[string]preparedStatementData),
		portals:            make(map[string]portalData),
	}
}

// clearUnnamed removes the unnamed statement and portal replaced by a simple Query message.
func (s *extendedQueryState) clearUnnamed() {
	delete(s.preparedStatements, "")
	delete(s.portals, "")
}

// close removes the named statement or portal identified by a Close message.
func (s *extendedQueryState) close(objectType byte, name string) {
	if objectType == 'S' {
		delete(s.preparedStatements, name)
	} else {
		delete(s.portals, name)
	}
}

// deallocate removes one prepared statement, or all prepared statements when name is empty.
func (s *extendedQueryState) deallocate(name string) error {
	if name == "" {
		clear(s.preparedStatements)
		return nil
	}
	if _, ok := s.preparedStatements[name]; !ok {
		return errors.Errorf("prepared statement %s does not exist", name)
	}
	delete(s.preparedStatements, name)
	return nil
}

// handleParse handles a Parse message.
func (h *ConnectionHandler) handleParse(message *pgproto3.Parse) error {
	// TODO: Named prepared statements must be explicitly closed before they can be redefined by another Parse
	// message, but this is not required for the unnamed statement.
	queries, err := convertQuery(message.Query)
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
		h.extended.preparedStatements[message.Name] = preparedStatementData{Query: query}
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

	// Clients may omit parameter OIDs or use zero to request inference. Explicit non-zero OIDs must be preserved.
	inferredTypes, err := extractBindVarTypes(ctx, analyzedPlan)
	if err != nil {
		return err
	}
	bindVarTypes := make([]uint32, len(inferredTypes))
	copy(bindVarTypes, inferredTypes)
	for i, oid := range message.ParameterOIDs {
		if oid != 0 && i < len(bindVarTypes) {
			bindVarTypes[i] = oid
		}
	}

	h.extended.preparedStatements[message.Name] = preparedStatementData{
		Query:        query,
		ReturnFields: fields,
		BindVarTypes: bindVarTypes,
	}
	return h.send(&pgproto3.ParseComplete{})
}

// handleDescribe handles a Describe message.
func (h *ConnectionHandler) handleDescribe(message *pgproto3.Describe) error {
	var fields []pgproto3.FieldDescription
	var bindVarTypes []uint32
	var query ConvertedQuery

	if message.ObjectType == 'S' {
		preparedStatement, ok := h.extended.preparedStatements[message.Name]
		if !ok {
			return errors.Errorf("prepared statement %s does not exist", message.Name)
		}
		fields = preparedStatement.ReturnFields
		bindVarTypes = preparedStatement.BindVarTypes
		query = preparedStatement.Query
	} else {
		portal, ok := h.extended.portals[message.Name]
		if !ok {
			return errors.Errorf("portal %s does not exist", message.Name)
		}
		fields = portal.Fields
		query = portal.Query
	}

	return h.sendDescribeResponse(fields, bindVarTypes, query)
}

// handleBind handles a Bind message.
func (h *ConnectionHandler) handleBind(message *pgproto3.Bind) error {
	// TODO: A named portal lasts until the end of the current transaction unless explicitly destroyed.
	logrus.Tracef("binding portal %q to prepared statement %s", message.DestinationPortal, message.PreparedStatement)
	preparedData, ok := h.extended.preparedStatements[message.PreparedStatement]
	if !ok {
		return errors.Errorf("prepared statement %s does not exist", message.PreparedStatement)
	}

	if err := h.rejectStatementIfTransactionFailed(preparedData.Query); err != nil {
		return err
	}

	if preparedData.Query.AST == nil {
		h.extended.portals[message.DestinationPortal] = portalData{Query: preparedData.Query, IsEmptyQuery: true}
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
	h.extended.portals[message.DestinationPortal] = portalData{
		Query:       preparedData.Query,
		Fields:      fields,
		BoundPlan:   boundPlan,
		FormatCodes: resultFormatCodes,
	}
	return h.send(&pgproto3.BindComplete{})
}

// handleExecute handles an Execute message.
func (h *ConnectionHandler) handleExecute(message *pgproto3.Execute) error {
	// TODO: Implement RowMax.
	portalData, ok := h.extended.portals[message.Portal]
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
	if err := h.startImplicitTransaction(query); err != nil {
		return err
	}

	handled, _, err := h.handleQueryOutsideEngine(query, newExtendedQueryCopyContinuation())
	if handled {
		return err
	}

	rowsAffected := int32(0)
	callback := h.spoolRowsCallback(query, &rowsAffected, true)
	err = h.doltgresHandler.ComExecuteBound(context.Background(), h.mysqlConn, query.String, portalData.BoundPlan, portalData.FormatCodes, callback)
	if err != nil {
		return err
	}
	return h.send(makeCommandComplete(query.StatementTag, rowsAffected))
}

// deallocatePreparedStatement implements DEALLOCATE for this connection.
func (h *ConnectionHandler) deallocatePreparedStatement(name string, query ConvertedQuery) error {
	if err := h.extended.deallocate(name); err != nil {
		return err
	}
	return h.send(&pgproto3.CommandComplete{CommandTag: []byte(query.StatementTag)})
}

// sendDescribeResponse sends the appropriate response to a Describe message.
func (h *ConnectionHandler) sendDescribeResponse(fields []pgproto3.FieldDescription, types []uint32, query ConvertedQuery) error {
	if types != nil {
		if err := h.send(&pgproto3.ParameterDescription{ParameterOIDs: types}); err != nil {
			return err
		}
	}
	if returnsRow(query) {
		return h.send(&pgproto3.RowDescription{Fields: fields})
	}
	return h.send(&pgproto3.NoData{})
}
