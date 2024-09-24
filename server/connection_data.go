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
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
	"github.com/dolthub/vitess/go/vt/proto/query"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/core/dataloader"
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
	extractBindVars := func(expr sql.Expression) bool {
		if err != nil {
			return false
		}
		switch e := expr.(type) {
		case *expression.BindVar:
			var typOid uint32
			if doltgresType, ok := e.Type().(pgtypes.DoltgresType); ok {
				typOid = doltgresType.OID()
			} else {
				// TODO: should remove usage non doltgres type
				typOid, err = VitessTypeToObjectID(e.Type().Type())
				if err != nil {
					err = fmt.Errorf("could not determine OID for placeholder %s: %w", e.Name, err)
					return false
				}
			}
			types = append(types, typOid)
		case *pgexprs.ExplicitCast:
			if bindVar, ok := e.Child().(*expression.BindVar); ok {
				var typOid uint32
				if doltgresType, ok := bindVar.Type().(pgtypes.DoltgresType); ok {
					typOid = doltgresType.OID()
				} else {
					typOid, err = VitessTypeToObjectID(e.Type().Type())
					if err != nil {
						err = fmt.Errorf("could not determine OID for placeholder %s: %w", bindVar.Name, err)
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
					err = fmt.Errorf("could not determine OID for placeholder %s: %w", bindVar.Name, err)
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
		return 0, fmt.Errorf("unsupported type: %s", typ)
	}
}

// OidToDoltgresType is map of oid to Doltgres type.
var OidToDoltgresType = map[uint32]pgtypes.DoltgresType{
	uint32(oid.T_bool):             pgtypes.Bool,
	uint32(oid.T_bytea):            pgtypes.Bytea,
	uint32(oid.T_char):             pgtypes.InternalChar,
	uint32(oid.T_name):             pgtypes.Name,
	uint32(oid.T_int8):             pgtypes.Int64,
	uint32(oid.T_int2):             pgtypes.Int16,
	uint32(oid.T_int2vector):       pgtypes.Unknown,
	uint32(oid.T_int4):             pgtypes.Int32,
	uint32(oid.T_regproc):          pgtypes.Regproc,
	uint32(oid.T_text):             pgtypes.Text,
	uint32(oid.T_oid):              pgtypes.Oid,
	uint32(oid.T_tid):              pgtypes.Unknown,
	uint32(oid.T_xid):              pgtypes.Xid,
	uint32(oid.T_cid):              pgtypes.Unknown,
	uint32(oid.T_oidvector):        pgtypes.Unknown,
	uint32(oid.T_pg_ddl_command):   pgtypes.Unknown,
	uint32(oid.T_pg_type):          pgtypes.Unknown,
	uint32(oid.T_pg_attribute):     pgtypes.Unknown,
	uint32(oid.T_pg_proc):          pgtypes.Unknown,
	uint32(oid.T_pg_class):         pgtypes.Unknown,
	uint32(oid.T_json):             pgtypes.Json,
	uint32(oid.T_xml):              pgtypes.Unknown,
	uint32(oid.T__xml):             pgtypes.Unknown,
	uint32(oid.T_pg_node_tree):     pgtypes.Unknown,
	uint32(oid.T__json):            pgtypes.JsonArray,
	uint32(oid.T_smgr):             pgtypes.Unknown,
	uint32(oid.T_index_am_handler): pgtypes.Unknown,
	uint32(oid.T_point):            pgtypes.Unknown,
	uint32(oid.T_lseg):             pgtypes.Unknown,
	uint32(oid.T_path):             pgtypes.Unknown,
	uint32(oid.T_box):              pgtypes.Unknown,
	uint32(oid.T_polygon):          pgtypes.Unknown,
	uint32(oid.T_line):             pgtypes.Unknown,
	uint32(oid.T__line):            pgtypes.Unknown,
	uint32(oid.T_cidr):             pgtypes.Unknown,
	uint32(oid.T__cidr):            pgtypes.Unknown,
	uint32(oid.T_float4):           pgtypes.Float32,
	uint32(oid.T_float8):           pgtypes.Float64,
	uint32(oid.T_abstime):          pgtypes.Unknown,
	uint32(oid.T_reltime):          pgtypes.Unknown,
	uint32(oid.T_tinterval):        pgtypes.Unknown,
	uint32(oid.T_unknown):          pgtypes.Unknown,
	uint32(oid.T_circle):           pgtypes.Unknown,
	uint32(oid.T__circle):          pgtypes.Unknown,
	uint32(oid.T_money):            pgtypes.Unknown,
	uint32(oid.T__money):           pgtypes.Unknown,
	uint32(oid.T_macaddr):          pgtypes.Unknown,
	uint32(oid.T_inet):             pgtypes.Unknown,
	uint32(oid.T__bool):            pgtypes.BoolArray,
	uint32(oid.T__bytea):           pgtypes.ByteaArray,
	uint32(oid.T__char):            pgtypes.InternalCharArray,
	uint32(oid.T__name):            pgtypes.NameArray,
	uint32(oid.T__int2):            pgtypes.Int16Array,
	uint32(oid.T__int2vector):      pgtypes.Unknown,
	uint32(oid.T__int4):            pgtypes.Int32Array,
	uint32(oid.T__regproc):         pgtypes.RegprocArray,
	uint32(oid.T__text):            pgtypes.TextArray,
	uint32(oid.T__tid):             pgtypes.Unknown,
	uint32(oid.T__xid):             pgtypes.XidArray,
	uint32(oid.T__cid):             pgtypes.Unknown,
	uint32(oid.T__oidvector):       pgtypes.Unknown,
	uint32(oid.T__bpchar):          pgtypes.BpCharArray,
	uint32(oid.T__varchar):         pgtypes.VarCharArray,
	uint32(oid.T__int8):            pgtypes.Int64Array,
	uint32(oid.T__point):           pgtypes.Unknown,
	uint32(oid.T__lseg):            pgtypes.Unknown,
	uint32(oid.T__path):            pgtypes.Unknown,
	uint32(oid.T__box):             pgtypes.Unknown,
	uint32(oid.T__float4):          pgtypes.Float32Array,
	uint32(oid.T__float8):          pgtypes.Float64Array,
	uint32(oid.T__abstime):         pgtypes.Unknown,
	uint32(oid.T__reltime):         pgtypes.Unknown,
	uint32(oid.T__tinterval):       pgtypes.Unknown,
	uint32(oid.T__polygon):         pgtypes.Unknown,
	uint32(oid.T__oid):             pgtypes.OidArray,
	uint32(oid.T_aclitem):          pgtypes.Unknown,
	uint32(oid.T__aclitem):         pgtypes.Unknown,
	uint32(oid.T__macaddr):         pgtypes.Unknown,
	uint32(oid.T__inet):            pgtypes.Unknown,
	uint32(oid.T_bpchar):           pgtypes.BpChar,
	uint32(oid.T_varchar):          pgtypes.VarChar,
	uint32(oid.T_date):             pgtypes.Date,
	uint32(oid.T_time):             pgtypes.Time,
	uint32(oid.T_timestamp):        pgtypes.Timestamp,
	uint32(oid.T__timestamp):       pgtypes.TimestampArray,
	uint32(oid.T__date):            pgtypes.DateArray,
	uint32(oid.T__time):            pgtypes.TimeArray,
	uint32(oid.T_timestamptz):      pgtypes.TimestampTZ,
	uint32(oid.T__timestamptz):     pgtypes.TimestampTZArray,
	uint32(oid.T_interval):         pgtypes.Interval,
	uint32(oid.T__interval):        pgtypes.IntervalArray,
	uint32(oid.T__numeric):         pgtypes.NumericArray,
	uint32(oid.T_pg_database):      pgtypes.Unknown,
	uint32(oid.T__cstring):         pgtypes.Unknown,
	uint32(oid.T_timetz):           pgtypes.TimeTZ,
	uint32(oid.T__timetz):          pgtypes.TimeTZArray,
	uint32(oid.T_bit):              pgtypes.Unknown,
	uint32(oid.T__bit):             pgtypes.Unknown,
	uint32(oid.T_varbit):           pgtypes.Unknown,
	uint32(oid.T__varbit):          pgtypes.Unknown,
	uint32(oid.T_numeric):          pgtypes.Numeric,
	uint32(oid.T_refcursor):        pgtypes.Unknown,
	uint32(oid.T__refcursor):       pgtypes.Unknown,
	uint32(oid.T_regprocedure):     pgtypes.Unknown,
	uint32(oid.T_regoper):          pgtypes.Unknown,
	uint32(oid.T_regoperator):      pgtypes.Unknown,
	uint32(oid.T_regclass):         pgtypes.Regclass,
	uint32(oid.T_regtype):          pgtypes.Regtype,
	uint32(oid.T__regprocedure):    pgtypes.Unknown,
	uint32(oid.T__regoper):         pgtypes.Unknown,
	uint32(oid.T__regoperator):     pgtypes.Unknown,
	uint32(oid.T__regclass):        pgtypes.RegclassArray,
	uint32(oid.T__regtype):         pgtypes.RegtypeArray,
	uint32(oid.T_record):           pgtypes.Unknown,
	uint32(oid.T_cstring):          pgtypes.Unknown,
	uint32(oid.T_any):              pgtypes.Unknown,
	uint32(oid.T_anyarray):         pgtypes.AnyArray,
	uint32(oid.T_void):             pgtypes.Unknown,
	uint32(oid.T_trigger):          pgtypes.Unknown,
	uint32(oid.T_language_handler): pgtypes.Unknown,
	uint32(oid.T_internal):         pgtypes.Unknown,
	uint32(oid.T_opaque):           pgtypes.Unknown,
	uint32(oid.T_anyelement):       pgtypes.AnyElement,
	uint32(oid.T__record):          pgtypes.Unknown,
	uint32(oid.T_anynonarray):      pgtypes.AnyNonArray,
	uint32(oid.T_pg_authid):        pgtypes.Unknown,
	uint32(oid.T_pg_auth_members):  pgtypes.Unknown,
	uint32(oid.T__txid_snapshot):   pgtypes.Unknown,
	uint32(oid.T_uuid):             pgtypes.Uuid,
	uint32(oid.T__uuid):            pgtypes.UuidArray,
	uint32(oid.T_txid_snapshot):    pgtypes.Unknown,
	uint32(oid.T_fdw_handler):      pgtypes.Unknown,
	uint32(oid.T_pg_lsn):           pgtypes.Unknown,
	uint32(oid.T__pg_lsn):          pgtypes.Unknown,
	uint32(oid.T_tsm_handler):      pgtypes.Unknown,
	uint32(oid.T_anyenum):          pgtypes.Unknown,
	uint32(oid.T_tsvector):         pgtypes.Unknown,
	uint32(oid.T_tsquery):          pgtypes.Unknown,
	uint32(oid.T_gtsvector):        pgtypes.Unknown,
	uint32(oid.T__tsvector):        pgtypes.Unknown,
	uint32(oid.T__gtsvector):       pgtypes.Unknown,
	uint32(oid.T__tsquery):         pgtypes.Unknown,
	uint32(oid.T_regconfig):        pgtypes.Unknown,
	uint32(oid.T__regconfig):       pgtypes.Unknown,
	uint32(oid.T_regdictionary):    pgtypes.Unknown,
	uint32(oid.T__regdictionary):   pgtypes.Unknown,
	uint32(oid.T_jsonb):            pgtypes.JsonB,
	uint32(oid.T__jsonb):           pgtypes.JsonBArray,
	uint32(oid.T_anyrange):         pgtypes.Unknown,
	uint32(oid.T_event_trigger):    pgtypes.Unknown,
	uint32(oid.T_int4range):        pgtypes.Unknown,
	uint32(oid.T__int4range):       pgtypes.Unknown,
	uint32(oid.T_numrange):         pgtypes.Unknown,
	uint32(oid.T__numrange):        pgtypes.Unknown,
	uint32(oid.T_tsrange):          pgtypes.Unknown,
	uint32(oid.T__tsrange):         pgtypes.Unknown,
	uint32(oid.T_tstzrange):        pgtypes.Unknown,
	uint32(oid.T__tstzrange):       pgtypes.Unknown,
	uint32(oid.T_daterange):        pgtypes.Unknown,
	uint32(oid.T__daterange):       pgtypes.Unknown,
	uint32(oid.T_int8range):        pgtypes.Unknown,
	uint32(oid.T__int8range):       pgtypes.Unknown,
	uint32(oid.T_pg_shseclabel):    pgtypes.Unknown,
	uint32(oid.T_regnamespace):     pgtypes.Unknown,
	uint32(oid.T__regnamespace):    pgtypes.Unknown,
	uint32(oid.T_regrole):          pgtypes.Unknown,
	uint32(oid.T__regrole):         pgtypes.Unknown,
}
