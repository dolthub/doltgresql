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
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/lib/pq/oid"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
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

// extendedQueryObjects owns the prepared statements and portals belonging to one connection.
type extendedQueryObjects struct {
	preparedStatements map[string]preparedStatementData
	portals            map[string]portalData
}

// newExtendedQueryObjects returns empty extended-query objects for a new connection.
func newExtendedQueryObjects() extendedQueryObjects {
	return extendedQueryObjects{
		preparedStatements: make(map[string]preparedStatementData),
		portals:            make(map[string]portalData),
	}
}

// clearUnnamed removes the unnamed statement and portal replaced by a simple Query message.
func (s *extendedQueryObjects) clearUnnamed() {
	delete(s.preparedStatements, "")
	delete(s.portals, "")
}

// close removes the named statement or portal identified by a Close message.
func (s *extendedQueryObjects) close(objectType byte, name string) {
	if objectType == 'S' {
		delete(s.preparedStatements, name)
	} else {
		delete(s.portals, name)
	}
}

// deallocate removes one prepared statement, or all prepared statements when name is empty.
func (s *extendedQueryObjects) deallocate(name string) error {
	if name == "" {
		clear(s.preparedStatements)
		return nil
	}
	if _, ok := s.preparedStatements[name]; !ok {
		return pgerror.Newf(pgcode.InvalidSQLStatementName, "prepared statement %q does not exist", name)
	}
	delete(s.preparedStatements, name)
	return nil
}

// handleParse handles a Parse message.
func (h *ConnectionHandler) handleParse(message *pgproto3.Parse) error {
	if message.Name != "" {
		if _, ok := h.state.extendedQueryObjects.preparedStatements[message.Name]; ok {
			return pgerror.Newf(pgcode.DuplicatePreparedStatement,
				"prepared statement %q already exists", message.Name)
		}
	}
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
		h.state.extendedQueryObjects.preparedStatements[message.Name] = preparedStatementData{Query: query}
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

	h.state.extendedQueryObjects.preparedStatements[message.Name] = preparedStatementData{
		Query:        query,
		ReturnFields: fields,
		BindVarTypes: bindVarTypes,
	}
	return h.send(&pgproto3.ParseComplete{})
}

// handleDescribe handles a Describe message.
func (h *ConnectionHandler) handleDescribe(message *pgproto3.Describe) error {
	if message.ObjectType != 'S' && message.ObjectType != 'P' {
		return pgerror.Newf(pgcode.ProtocolViolation, "invalid DESCRIBE message subtype %d", message.ObjectType)
	}
	var fields []pgproto3.FieldDescription
	var bindVarTypes []uint32
	var query ConvertedQuery

	if message.ObjectType == 'S' {
		preparedStatement, ok := h.state.extendedQueryObjects.preparedStatements[message.Name]
		if !ok {
			return pgerror.Newf(pgcode.InvalidSQLStatementName,
				"prepared statement %q does not exist", message.Name)
		}
		fields = preparedStatement.ReturnFields
		bindVarTypes = preparedStatement.BindVarTypes
		query = preparedStatement.Query
	} else {
		portal, ok := h.state.extendedQueryObjects.portals[message.Name]
		if !ok {
			return pgerror.Newf(pgcode.InvalidCursorName, "portal %q does not exist", message.Name)
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
	if message.DestinationPortal != "" {
		if _, ok := h.state.extendedQueryObjects.portals[message.DestinationPortal]; ok {
			return pgerror.Newf(pgcode.DuplicateCursor, "cursor %q already exists", message.DestinationPortal)
		}
	}
	preparedData, ok := h.state.extendedQueryObjects.preparedStatements[message.PreparedStatement]
	if !ok {
		return pgerror.Newf(pgcode.InvalidSQLStatementName,
			"prepared statement %q does not exist", message.PreparedStatement)
	}

	if err := h.rejectStatementIfTransactionFailed(preparedData.Query); err != nil {
		return err
	}

	if preparedData.Query.AST == nil {
		h.state.extendedQueryObjects.portals[message.DestinationPortal] = portalData{Query: preparedData.Query, IsEmptyQuery: true}
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
	h.state.extendedQueryObjects.portals[message.DestinationPortal] = portalData{
		Query:       preparedData.Query,
		Fields:      fields,
		BoundPlan:   boundPlan,
		FormatCodes: resultFormatCodes,
	}
	return h.send(&pgproto3.BindComplete{})
}

// handleClose removes the named statement or portal identified by a valid Close message.
func (h *ConnectionHandler) handleClose(message *pgproto3.Close) error {
	if message.ObjectType != 'S' && message.ObjectType != 'P' {
		return pgerror.Newf(pgcode.ProtocolViolation, "invalid CLOSE message subtype %d", message.ObjectType)
	}
	h.state.extendedQueryObjects.close(message.ObjectType, message.Name)
	return h.send(&pgproto3.CloseComplete{})
}

// handleExecute handles an Execute message.
func (h *ConnectionHandler) handleExecute(message *pgproto3.Execute) error {
	// TODO: Implement RowMax.
	portalData, ok := h.state.extendedQueryObjects.portals[message.Portal]
	if !ok {
		return pgerror.Newf(pgcode.InvalidCursorName, "portal %q does not exist", message.Portal)
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

	handled, _, err := h.handleQueryOutsideEngine(query, nil)
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
	if err := h.state.extendedQueryObjects.deallocate(name); err != nil {
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

// extractBindVarTypes infers parameter OIDs from an analyzed extended-query plan.
func extractBindVarTypes(ctx *sql.Context, queryPlan sql.Node) ([]uint32, error) {
	types := make(map[string]uint32)
	var err error
	var extractBindVars func(ctx *sql.Context, n sql.Node, expr sql.Expression) bool
	extractBindVars = func(ctx *sql.Context, n sql.Node, expr sql.Expression) bool {
		if err != nil {
			return false
		}

		switch e := expr.(type) {
		// Subquery doesn't walk its Node child via Expressions, so we must walk it separately here.
		case *plan.Subquery:
			transform.InspectExpressionsWithNode(ctx, e.Query, extractBindVars)
		case *expression.BindVar:
			var typOID uint32
			if doltgresType, ok := e.Type(ctx).(*pgtypes.DoltgresType); ok {
				typOID = id.Cache().ToOID(doltgresType.ID.AsId())
			} else if _, ok := e.Type(ctx).(sql.DeferredType); ok {
				// Deferred LIMIT and OFFSET parameters have the PostgreSQL integer type expected by those clauses.
				switch n.(type) {
				case *plan.Limit, *plan.Offset:
					typOID = uint32(oid.T_int4)
				default:
					typOID, err = VitessTypeToObjectID(e.Type(ctx))
				}
			} else {
				// TODO: Remove uses of non-Doltgres types.
				typOID, err = VitessTypeToObjectID(e.Type(ctx))
			}
			if err != nil {
				err = errors.Wrapf(err, "could not determine OID for placeholder %s", e.Name)
				return false
			}
			err = recordBindVarType(ctx, types, e.Name, typOID)
		case *pgexprs.ExplicitCast:
			if bindVar, ok := e.Child().(*expression.BindVar); ok {
				var typOID uint32
				if doltgresType, ok := e.Type(ctx).(*pgtypes.DoltgresType); ok {
					typOID = id.Cache().ToOID(doltgresType.ID.AsId())
				} else {
					typOID, err = VitessTypeToObjectID(e.Type(ctx))
				}
				if err != nil {
					err = errors.Wrapf(err, "could not determine OID for placeholder %s", bindVar.Name)
					return false
				}
				err = recordBindVarType(ctx, types, bindVar.Name, typOID)
				return false
			}
		// $1::text and similar get converted to a Convert expression wrapping the bind variable.
		case *expression.Convert:
			if bindVar, ok := e.Child.(*expression.BindVar); ok {
				typOID, typeErr := VitessTypeToObjectID(e.Type(ctx))
				if typeErr != nil {
					err = errors.Wrapf(typeErr, "could not determine OID for placeholder %s", bindVar.Name)
					return false
				}
				err = recordBindVarType(ctx, types, bindVar.Name, typOID)
				return false
			}
		}
		return true
	}

	transform.InspectExpressionsWithNode(ctx, queryPlan, extractBindVars)

	// Insert nodes are special, as their source expressions are not returned by Expressions().
	if insert, ok := queryPlan.(*plan.InsertInto); ok {
		transform.InspectExpressionsWithNode(ctx, insert.Source, extractBindVars)
		bindInsertSelect(ctx, insert, types)
	}

	typesArr := make([]uint32, len(types))
	for name, typOID := range types {
		idx, parseErr := strconv.ParseInt(strings.TrimPrefix(name, "v"), 10, 32)
		if parseErr != nil {
			return nil, errors.Wrapf(parseErr, "could not determine the index of placeholder %s", name)
		}
		if int(idx-1) >= len(types) {
			return nil, errors.Errorf("could not determine the index of placeholder %s in slice of %d elements", name, len(types))
		}
		typesArr[idx-1] = typOID
	}
	return typesArr, err
}

// bindInsertSelect infers direct SELECT bind variables from their corresponding INSERT destination columns.
func bindInsertSelect(ctx *sql.Context, insert *plan.InsertInto, types map[string]uint32) {
	project, ok := insert.Source.(*plan.Project)
	if !ok {
		return
	}
	destinationTypes := make(map[string]sql.Type)
	for _, col := range insert.Destination.Schema(ctx) {
		destinationTypes[strings.ToLower(col.Name)] = col.Type
	}
	unknownOID := id.Cache().ToOID(pgtypes.Unknown.ID.AsId())
	for i, projection := range project.Projections {
		bindVar := pgexprs.UnwrapBindVar(projection)
		if bindVar == nil || i >= len(insert.ColumnNames) || types[bindVar.Name] != unknownOID {
			continue
		}
		destinationType, ok := destinationTypes[strings.ToLower(insert.ColumnNames[i])]
		if !ok {
			continue
		}
		if doltgresType, ok := destinationType.(*pgtypes.DoltgresType); ok {
			types[bindVar.Name] = id.Cache().ToOID(doltgresType.ID.AsId())
		} else if typOID, typeErr := VitessTypeToObjectID(destinationType); typeErr == nil {
			types[bindVar.Name] = typOID
		}
	}
}

// recordBindVarType records one inferred parameter type after checking repeated uses for compatibility.
func recordBindVarType(ctx *sql.Context, types map[string]uint32, name string, typOID uint32) error {
	if existingOID, ok := types[name]; ok {
		if err := checkCompatibleTypes(ctx, existingOID, typOID, name); err != nil {
			return err
		}
	}
	types[name] = typOID
	return nil
}

// checkCompatibleTypes checks whether the types inferred for repeated uses of a parameter are compatible.
func checkCompatibleTypes(ctx *sql.Context, existingOID, newOID uint32, name string) error {
	existing := pgtypes.GetTypeByID(id.Type(id.Cache().ToInternal(existingOID)))
	newType := pgtypes.GetTypeByID(id.Type(id.Cache().ToInternal(newOID)))
	if existing == nil || newType == nil {
		// TODO: User-defined types are not in the built-in map, so their compatibility is not checked.
		return nil
	}
	if _, _, err := framework.FindCommonType(ctx, []*pgtypes.DoltgresType{existing, newType}); err != nil {
		return errors.Errorf("parameter %s is used for incompatible types: %s and %s", name, existing.String(), newType.String())
	}
	return nil
}
