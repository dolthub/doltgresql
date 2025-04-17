// Copyright 2024 Dolthub, Inc.
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
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
	"github.com/dolthub/vitess/go/vt/proto/query"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/core/dataloader"
	"github.com/dolthub/doltgresql/core/id"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/node"
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

// ReadyForQueryTransactionIndicator indicates the state of the transaction related to the query.
type ReadyForQueryTransactionIndicator byte

const (
	ReadyForQueryTransactionIndicator_Idle                   ReadyForQueryTransactionIndicator = 'I'
	ReadyForQueryTransactionIndicator_TransactionBlock       ReadyForQueryTransactionIndicator = 'T'
	ReadyForQueryTransactionIndicator_FailedTransactionBlock ReadyForQueryTransactionIndicator = 'E'
)

// ConvertedQuery represents a query that has been converted from the Postgres representation to the Vitess
// representation. String may contain the string version of the converted query. AST will contain the tree
// version of the converted query, and is the recommended form to use. If AST is nil, then use the String version,
// otherwise always prefer to AST.
type ConvertedQuery struct {
	String       string
	AST          vitess.Statement
	StatementTag string
}

// copyFromStdinState tracks the metadata for an import of data into a table using a COPY FROM STDIN statement. When
// this statement is processed, the server accepts COPY DATA messages from the client with chunks of data to load
// into a table.
type copyFromStdinState struct {
	// copyFromStdinNode stores the original CopyFrom statement that initiated the CopyData message sequence. This
	// node is used to look at what parameters were specified, such as which table to load data into, file format,
	// delimiters, etc.
	copyFromStdinNode *node.CopyFrom
	// insertNode stores the analyzed insert node that will be used to load the data into the target table. This node
	// only needs to be built once, and can be reused with updates to its underlying data loader.
	insertNode sql.Node
	// dataLoader is the implementation of DataLoader that is used to load each individual CopyData chunk into the
	// target table.
	dataLoader dataloader.DataLoader
	// copyErr stores any error that was returned while processing a CopyData message and loading a chunk of data
	// to the target table. The server needs to keep track of any errors that were encountered while processing chunks
	// so that it can avoid sending a CommandComplete message if an error was encountered after the client already
	// sent a CopyDone message to the server.
	copyErr error
}

type PortalData struct {
	Query        ConvertedQuery
	IsEmptyQuery bool
	Fields       []pgproto3.FieldDescription
	BoundPlan    sql.Node
}

type PreparedStatementData struct {
	Query        ConvertedQuery
	ReturnFields []pgproto3.FieldDescription
	BindVarTypes []uint32
}

// extractBindVarTypes returns types based on the given query plan.
// This function is used to get bind var types for running our prepared
// tests only. A valid prepared query and execution messages must have
// the types defined.
func extractBindVarTypes(queryPlan sql.Node) ([]uint32, error) {
	inspectNode := queryPlan
	switch queryPlan := queryPlan.(type) {
	case *plan.InsertInto:
		inspectNode = queryPlan.Source
	}

	types := make([]uint32, 0)
	var err error
	var extractBindVars func(expr sql.Expression) bool
	extractBindVars = func(expr sql.Expression) bool {
		if err != nil {
			return false
		}
		
		switch e := expr.(type) {
		// Subquery doesn't walk its Node child via Expressions, so we must walk it separately here
		case *plan.Subquery:
			transform.InspectExpressions(e.Query, extractBindVars)
		case *expression.BindVar:
			var typOid uint32
			if doltgresType, ok := e.Type().(*pgtypes.DoltgresType); ok {
				typOid = id.Cache().ToOID(doltgresType.ID.AsId())
			} else {
				// TODO: should remove usage non doltgres type
				typOid, err = VitessTypeToObjectID(e.Type().Type())
				if err != nil {
					err = errors.Errorf("could not determine OID for placeholder %s: %w", e.Name, err)
					return false
				}
			}
			types = append(types, typOid)
		case *pgexprs.ExplicitCast:
			if bindVar, ok := e.Child().(*expression.BindVar); ok {
				var typOid uint32
				if doltgresType, ok := bindVar.Type().(*pgtypes.DoltgresType); ok {
					typOid = id.Cache().ToOID(doltgresType.ID.AsId())
				} else {
					typOid, err = VitessTypeToObjectID(e.Type().Type())
					if err != nil {
						err = errors.Errorf("could not determine OID for placeholder %s: %w", bindVar.Name, err)
						return false
					}
				}
				types = append(types, typOid)
				return false
			}
		// $1::text and similar get converted to a Convert expression wrapping the bindvar
		case *expression.Convert:
			if bindVar, ok := e.Child.(*expression.BindVar); ok {
				var typOid uint32
				typOid, err = VitessTypeToObjectID(e.Type().Type())
				if err != nil {
					err = errors.Errorf("could not determine OID for placeholder %s: %w", bindVar.Name, err)
					return false
				}
				types = append(types, typOid)
				return false
			}
		}

		return true
	}

	transform.InspectExpressions(inspectNode, extractBindVars)
	return types, err
}

// VitessTypeToObjectID returns a type, as defined by Vitess, into a type as defined by Postgres.
// OIDs can be obtained with the following query: `SELECT oid, typname FROM pg_type ORDER BY 1;`
func VitessTypeToObjectID(typ query.Type) (uint32, error) {
	switch typ {
	case query.Type_INT8:
		// Postgres doesn't make use of a small integer type for integer returns, which presents a bit of a conundrum.
		// GMS defines boolean operations as the smallest integer type, while Postgres has an explicit bool type.
		// We can't always assume that `INT8` means bool, since it could just be a small integer. As a result, we'll
		// always return this as though it's an `INT16`, which also means that we can't support bools right now.
		// OIDs 16 (bool) and 18 (char, ASCII only?) are the only single-byte types as far as I'm aware.
		return uint32(oid.T_int2), nil
	case query.Type_INT16:
		// The technically correct OID is 21 (2-byte integer), however it seems like some clients don't actually expect
		// this, so I'm not sure when it's actually used by Postgres. Because of this, we'll just pretend it's an `INT32`.
		return uint32(oid.T_int2), nil
	case query.Type_INT24:
		// Postgres doesn't have a 3-byte integer type, so just pretend it's `INT32`.
		return uint32(oid.T_int4), nil
	case query.Type_INT32:
		return uint32(oid.T_int4), nil
	case query.Type_INT64:
		return uint32(oid.T_int8), nil
	case query.Type_UINT8:
		return uint32(oid.T_int4), nil
	case query.Type_UINT16:
		return uint32(oid.T_int4), nil
	case query.Type_UINT24:
		return uint32(oid.T_int4), nil
	case query.Type_UINT32:
		// Since this has an upperbound greater than `INT32`, we'll treat it as `INT64`
		return uint32(oid.T_oid), nil
	case query.Type_UINT64:
		// Since this has an upperbound greater than `INT64`, we'll treat it as `NUMERIC`
		return uint32(oid.T_numeric), nil
	case query.Type_FLOAT32:
		return uint32(oid.T_float4), nil
	case query.Type_FLOAT64:
		return uint32(oid.T_float8), nil
	case query.Type_DECIMAL:
		return uint32(oid.T_numeric), nil
	case query.Type_CHAR:
		return uint32(oid.T_char), nil
	case query.Type_VARCHAR:
		return uint32(oid.T_varchar), nil
	case query.Type_TEXT:
		return uint32(oid.T_text), nil
	case query.Type_BLOB:
		return uint32(oid.T_bytea), nil
	case query.Type_JSON:
		return uint32(oid.T_json), nil
	case query.Type_TIMESTAMP, query.Type_DATETIME:
		return uint32(oid.T_timestamp), nil
	case query.Type_DATE:
		return uint32(oid.T_date), nil
	case query.Type_NULL_TYPE:
		return uint32(oid.T_text), nil // NULL is treated as TEXT on the wire
	case query.Type_ENUM:
		return uint32(oid.T_text), nil // TODO: temporary solution until we support CREATE TYPE
	default:
		return 0, errors.Errorf("unsupported type: %s", typ)
	}
}
