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
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime/trace"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

var printErrorStackTraces = false

const PrintErrorStackTracesEnvKey = "DOLTGRES_PRINT_ERROR_STACK_TRACES"

func init() {
	if _, ok := os.LookupEnv(PrintErrorStackTracesEnvKey); ok {
		printErrorStackTraces = true
	}
}

// BindVariables represents arrays of types, format codes and parameters
// used to convert given parameters to binding variables map.
type BindVariables struct {
	varTypes    []uint32
	formatCodes []int16
	parameters  [][]byte
}

// Result represents a query result.
type Result struct {
	Fields       []pgproto3.FieldDescription `json:"fields"`
	Rows         []Row                       `json:"rows"`
	RowsAffected uint64                      `json:"rows_affected"`
}

// Row represents a single row value in bytes format.
// |val| represents array of a single row elements,
// which each element value is in byte array format.
type Row struct {
	val [][]byte
}

const rowsBatch = 128

// DoltgresHandler is a handler uses SQLe engine directly
// running Doltgres specific queries.
type DoltgresHandler struct {
	e                 *sqle.Engine
	sm                *server.SessionManager
	readTimeout       time.Duration
	encodeLoggedQuery bool
	pgTypeMap         *pgtype.Map
	sel               server.ServerEventListener
	backend           *pgproto3.Backend
}

var _ Handler = &DoltgresHandler{}

// ComBind implements the Handler interface.
func (h *DoltgresHandler) ComBind(ctx context.Context, c *mysql.Conn, query string, parsedQuery mysql.ParsedQuery, bindVars BindVariables) (mysql.BoundQuery, []pgproto3.FieldDescription, error) {
	sqlCtx, err := h.sm.NewContextWithQuery(ctx, c, query)
	if err != nil {
		return nil, nil, err
	}

	stmt, ok := parsedQuery.(sqlparser.Statement)
	if !ok {
		return nil, nil, errors.Errorf("parsedQuery must be a sqlparser.Statement, but got %T", parsedQuery)
	}

	bvs, err := h.convertBindParameters(sqlCtx, bindVars.varTypes, bindVars.formatCodes, bindVars.parameters)
	if err != nil {
		if printErrorStackTraces {
			fmt.Printf("unable to convert bind params: %+v\n", err)
		}
		return nil, nil, err
	}

	queryPlan, err := h.e.BoundQueryPlan(sqlCtx, query, stmt, bvs)
	if err != nil {
		if printErrorStackTraces {
			fmt.Printf("unable to bind query plan: %+v\n", err)
		}
		return nil, nil, err
	}

	return queryPlan, schemaToFieldDescriptions(sqlCtx, queryPlan.Schema()), nil
}

// ComExecuteBound implements the Handler interface.
func (h *DoltgresHandler) ComExecuteBound(ctx context.Context, conn *mysql.Conn, query string, boundQuery mysql.BoundQuery, callback func(*Result) error) error {
	analyzedPlan, ok := boundQuery.(sql.Node)
	if !ok {
		return errors.Errorf("boundQuery must be a sql.Node, but got %T", boundQuery)
	}

	// TODO: This technically isn't query start and underestimates query execution time
	start := time.Now()
	if h.sel != nil {
		h.sel.QueryStarted()
	}

	err := h.doQuery(ctx, conn, query, nil, analyzedPlan, h.executeBoundPlan, callback)
	if err != nil {
		err = sql.CastSQLError(err)
	}

	if h.sel != nil {
		h.sel.QueryCompleted(err == nil, time.Since(start))
	}

	return err
}

// ComPrepareParsed implements the Handler interface.
func (h *DoltgresHandler) ComPrepareParsed(ctx context.Context, c *mysql.Conn, query string, parsed sqlparser.Statement) (mysql.ParsedQuery, []pgproto3.FieldDescription, error) {
	sqlCtx, err := h.sm.NewContextWithQuery(ctx, c, query)
	if err != nil {
		return nil, nil, err
	}

	analyzed, err := h.e.PrepareParsedQuery(sqlCtx, query, query, parsed)
	if err != nil {
		if printErrorStackTraces {
			fmt.Printf("unable to prepare query: %+v\n", err)
		}
		logrus.WithField("query", query).Errorf("unable to prepare query: %s", err.Error())
		err := sql.CastSQLError(err)
		return nil, nil, err
	}

	var fields []pgproto3.FieldDescription
	// The query is not a SELECT statement if it corresponds to an OK result.
	if nodeReturnsOkResultSchema(analyzed) {
		fields = []pgproto3.FieldDescription{
			{
				Name:         []byte("Rows"),
				DataTypeOID:  id.Cache().ToOID(pgtypes.Int32.ID.AsId()),
				DataTypeSize: int16(pgtypes.Int32.MaxTextResponseByteLength(nil)),
			},
		}
	} else {
		fields = schemaToFieldDescriptions(sqlCtx, analyzed.Schema())
	}
	return analyzed, fields, nil
}

// ComQuery implements the Handler interface.
func (h *DoltgresHandler) ComQuery(ctx context.Context, c *mysql.Conn, query string, parsed sqlparser.Statement, callback func(*Result) error) error {
	// TODO: This technically isn't query start and underestimates query execution time
	start := time.Now()
	if h.sel != nil {
		h.sel.QueryStarted()
	}

	err := h.doQuery(ctx, c, query, parsed, nil, h.executeQuery, callback)
	if err != nil {
		err = sql.CastSQLError(err)
	}

	if h.sel != nil {
		h.sel.QueryCompleted(err == nil, time.Since(start))
	}

	return err
}

// ComResetConnection implements the Handler interface.
func (h *DoltgresHandler) ComResetConnection(c *mysql.Conn) error {
	logrus.WithField("connectionId", c.ConnectionID).Debug("COM_RESET_CONNECTION command received")

	// Grab the currently selected database name
	db := h.sm.GetCurrentDB(c)

	// Dispose of the connection's current session
	h.maybeReleaseAllLocks(c)
	h.e.CloseSession(c.ConnectionID)

	ctx := context.Background()

	// Create a new session and set the current database
	err := h.sm.NewSession(ctx, c)
	if err != nil {
		return err
	}
	return h.sm.SetDB(ctx, c, db)
}

// ConnectionClosed implements the Handler interface.
func (h *DoltgresHandler) ConnectionClosed(c *mysql.Conn) {
	defer func() {
		if h.sel != nil {
			h.sel.ClientDisconnected()
		}
	}()

	defer h.sm.RemoveConn(c)
	defer h.e.CloseSession(c.ConnectionID)

	h.maybeReleaseAllLocks(c)

	logrus.WithField(sql.ConnectionIdLogField, c.ConnectionID).Infof("ConnectionClosed")
}

// NewConnection implements the Handler interface.
func (h *DoltgresHandler) NewConnection(c *mysql.Conn) {
	if h.sel != nil {
		h.sel.ClientConnected()
	}

	h.sm.AddConn(c)
	sql.StatusVariables.IncrementGlobal("Connections", 1)

	c.DisableClientMultiStatements = true // TODO: h.disableMultiStmts
	logrus.WithField(sql.ConnectionIdLogField, c.ConnectionID).WithField("DisableClientMultiStatements", c.DisableClientMultiStatements).Infof("NewConnection")
}

// NewContext implements the Handler interface.
func (h *DoltgresHandler) NewContext(ctx context.Context, c *mysql.Conn, query string) (*sql.Context, error) {
	return h.sm.NewContextWithQuery(ctx, c, query)
}

// convertBindParameters handles the conversion from bind parameters to variable values.
func (h *DoltgresHandler) convertBindParameters(ctx *sql.Context, types []uint32, formatCodes []int16, values [][]byte) (map[string]sqlparser.Expr, error) {
	bindings := make(map[string]sqlparser.Expr, len(values))
	for i := range values {
		typ := types[i]
		var bindVarString string
		// We'll rely on a library to decode each format, which will deal with text and binary representations for us
		if err := h.pgTypeMap.Scan(typ, formatCodes[i], values[i], &bindVarString); err != nil {
			return nil, err
		}

		pgTyp, ok := pgtypes.IDToBuiltInDoltgresType[id.Type(id.Cache().ToInternal(typ))]
		if !ok {
			return nil, errors.Errorf("unhandled oid type: %v", typ)
		}
		v, err := pgTyp.IoInput(ctx, bindVarString)
		if err != nil {
			return nil, err
		}
		bindings[fmt.Sprintf("v%d", i+1)] = sqlparser.InjectedExpr{Expression: pgexprs.NewUnsafeLiteral(v, pgTyp)}
	}
	return bindings, nil
}

var queryLoggingRegex = regexp.MustCompile(`[\r\n\t ]+`)

func (h *DoltgresHandler) doQuery(ctx context.Context, c *mysql.Conn, query string, parsed sqlparser.Statement, analyzedPlan sql.Node, queryExec QueryExecutor, callback func(*Result) error) error {
	sqlCtx, err := h.sm.NewContextWithQuery(ctx, c, query)
	if err != nil {
		return err
	}

	if err = core.SetBackend(sqlCtx, h.backend); err != nil {
		return err
	}

	start := time.Now()
	var queryStrToLog string
	if h.encodeLoggedQuery {
		queryStrToLog = base64.StdEncoding.EncodeToString([]byte(query))
	} else if logrus.IsLevelEnabled(logrus.DebugLevel) {
		// this is expensive, so skip this unless we're logging at DEBUG level
		queryStrToLog = string(queryLoggingRegex.ReplaceAll([]byte(query), []byte(" ")))
	}

	if queryStrToLog != "" {
		sqlCtx.SetLogger(sqlCtx.GetLogger().WithField("query", queryStrToLog))
	}
	sqlCtx.GetLogger().Debugf("Starting query")
	sqlCtx.GetLogger().Tracef("beginning execution")

	oCtx := ctx

	// TODO: it would be nice to put this logic in the engine, not the handler, but we don't want the process to be
	//  marked done until we're done spooling rows over the wire
	ctx, err = sqlCtx.ProcessList.BeginQuery(sqlCtx, query)
	defer func() {
		if err != nil && ctx != nil {
			sqlCtx.ProcessList.EndQuery(sqlCtx)
		}
	}()

	schema, rowIter, qFlags, err := queryExec(sqlCtx, query, parsed, analyzedPlan)
	if err != nil {
		if printErrorStackTraces {
			fmt.Printf("error running query: %+v\n", err)
		}
		sqlCtx.GetLogger().WithError(err).Warn("error running query")
		return err
	}

	// create result before goroutines to avoid |ctx| racing
	var r *Result
	var processedAtLeastOneBatch bool

	// zero/single return schema use spooling shortcut
	if types.IsOkResultSchema(schema) {
		r, err = resultForOkIter(sqlCtx, rowIter)
	} else if schema == nil {
		r, err = resultForEmptyIter(sqlCtx, rowIter)
	} else if analyzer.FlagIsSet(qFlags, sql.QFlagMax1Row) {
		resultFields := schemaToFieldDescriptions(sqlCtx, schema)
		r, err = resultForMax1RowIter(sqlCtx, schema, rowIter, resultFields)
	} else {
		resultFields := schemaToFieldDescriptions(sqlCtx, schema)
		r, processedAtLeastOneBatch, err = h.resultForDefaultIter(sqlCtx, schema, rowIter, callback, resultFields)
	}
	if err != nil {
		return err
	}

	// errGroup context is now canceled
	ctx = oCtx

	sqlCtx.GetLogger().Debugf("Query finished in %d ms", time.Since(start).Milliseconds())

	// processedAtLeastOneBatch means we already called callback() at least
	// once, so no need to call it if RowsAffected == 0.
	if r != nil && (r.RowsAffected == 0 && processedAtLeastOneBatch) {
		return nil
	}

	return callback(r)
}

// QueryExecutor is a function that executes a query and returns the result as a schema and iterator. Either of
// |parsed| or |analyzed| can be nil depending on the use case
type QueryExecutor func(ctx *sql.Context, query string, parsed sqlparser.Statement, analyzed sql.Node) (sql.Schema, sql.RowIter, *sql.QueryFlags, error)

// executeQuery is a QueryExecutor that calls QueryWithBindings on the given engine using the given query and parsed
// statement, which may be nil.
func (h *DoltgresHandler) executeQuery(ctx *sql.Context, query string, parsed sqlparser.Statement, _ sql.Node) (sql.Schema, sql.RowIter, *sql.QueryFlags, error) {
	return h.e.QueryWithBindings(ctx, query, parsed, nil, nil)
}

// executeBoundPlan is a QueryExecutor that calls QueryWithBindings on the given engine using the given query and parsed
// statement, which may be nil.
func (h *DoltgresHandler) executeBoundPlan(ctx *sql.Context, query string, _ sqlparser.Statement, plan sql.Node) (sql.Schema, sql.RowIter, *sql.QueryFlags, error) {
	return h.e.PrepQueryPlanForExecution(ctx, query, plan, nil)
}

// maybeReleaseAllLocks makes a best effort attempt to release all locks on the given connection. If the attempt fails,
// an error is logged but not returned.
func (h *DoltgresHandler) maybeReleaseAllLocks(c *mysql.Conn) {
	if ctx, err := h.sm.NewContextWithQuery(context.Background(), c, ""); err != nil {
		logrus.Errorf("unable to release all locks on session close: %s", err)
		logrus.Errorf("unable to unlock tables on session close: %s", err)
	} else {
		_, err = h.e.LS.ReleaseAll(ctx)
		if err != nil {
			logrus.Errorf("unable to release all locks on session close: %s", err)
		}
		if err = h.e.Analyzer.Catalog.UnlockTables(ctx, c.ConnectionID); err != nil {
			logrus.Errorf("unable to unlock tables on session close: %s", err)
		}
	}
}

// nodeReturnsOkResultSchema returns whether the node returns OK result or the schema is OK result schema.
// These nodes will eventually return an OK result, but their intermediate forms here return a different schema
// than they will at execution time.
func nodeReturnsOkResultSchema(node sql.Node) bool {
	switch node.(type) {
	case *plan.InsertInto, *plan.Update, *plan.UpdateJoin, *plan.DeleteFrom:
		return true
	}
	return types.IsOkResultSchema(node.Schema())
}

func schemaToFieldDescriptions(ctx *sql.Context, s sql.Schema) []pgproto3.FieldDescription {
	fields := make([]pgproto3.FieldDescription, len(s))
	for i, c := range s {
		var oid uint32
		var typmod = int32(-1)
		var err error
		if doltgresType, ok := c.Type.(*pgtypes.DoltgresType); ok {
			if doltgresType.TypType == pgtypes.TypeType_Domain {
				oid = id.Cache().ToOID(doltgresType.BaseTypeID.AsId())
			} else {
				oid = id.Cache().ToOID(doltgresType.ID.AsId())
			}
			typmod = doltgresType.GetAttTypMod() // pg_attribute.atttypmod
		} else {
			oid, err = VitessTypeToObjectID(c.Type.Type())
			if err != nil {
				panic(err)
			}
		}

		// "Format" field: The format code being used for the field.
		// Currently, will be zero (text) or one (binary).
		// In a RowDescription returned from the statement variant of Describe,
		// the format code is not yet known and will always be zero.

		fields[i] = pgproto3.FieldDescription{
			Name:                 []byte(c.Name),
			TableOID:             uint32(0),
			TableAttributeNumber: uint16(0),
			DataTypeOID:          oid,
			DataTypeSize:         int16(c.Type.MaxTextResponseByteLength(ctx)),
			TypeModifier:         typmod,
			Format:               int16(0),
		}
	}

	return fields
}

// resultForOkIter reads a maximum of one result row from a result iterator.
func resultForOkIter(ctx *sql.Context, iter sql.RowIter) (*Result, error) {
	defer trace.StartRegion(ctx, "DoltgresHandler.resultForOkIter").End()

	row, err := iter.Next(ctx)
	if err != nil {
		if printErrorStackTraces {
			fmt.Printf("row: %+v\n", err)
		}
		return nil, err
	}
	_, err = iter.Next(ctx)
	if err != io.EOF {
		return nil, errors.Errorf("result schema iterator returned more than one row")
	}
	if err := iter.Close(ctx); err != nil {
		return nil, err
	}

	return &Result{
		RowsAffected: row[0].(types.OkResult).RowsAffected,
	}, nil
}

// resultForEmptyIter ensures that an expected empty iterator returns no rows.
func resultForEmptyIter(ctx *sql.Context, iter sql.RowIter) (*Result, error) {
	defer trace.StartRegion(ctx, "DoltgresHandler.resultForEmptyIter").End()
	if _, err := iter.Next(ctx); err != io.EOF {
		return nil, errors.Errorf("result schema iterator returned more than zero rows")
	}
	if err := iter.Close(ctx); err != nil {
		return nil, err
	}
	return &Result{Fields: nil}, nil
}

// resultForMax1RowIter ensures that an empty iterator returns at most one row
func resultForMax1RowIter(ctx *sql.Context, schema sql.Schema, iter sql.RowIter, resultFields []pgproto3.FieldDescription) (*Result, error) {
	defer trace.StartRegion(ctx, "DoltgresHandler.resultForMax1RowIter").End()
	row, err := iter.Next(ctx)
	if err == io.EOF {
		return &Result{Fields: resultFields}, nil
	} else if err != nil {
		return nil, err
	}

	if _, err = iter.Next(ctx); err != io.EOF {
		return nil, errors.Errorf("result max1Row iterator returned more than one row")
	}
	if err := iter.Close(ctx); err != nil {
		return nil, err
	}

	outputRow, err := rowToBytes(ctx, schema, row)
	if err != nil {
		return nil, err
	}

	ctx.GetLogger().Tracef("spooling result row %s", outputRow)

	return &Result{Fields: resultFields, Rows: []Row{{outputRow}}, RowsAffected: 1}, nil
}

// resultForDefaultIter reads batches of rows from the iterator
// and writes results into the callback function.
func (h *DoltgresHandler) resultForDefaultIter(ctx *sql.Context, schema sql.Schema, iter sql.RowIter, callback func(*Result) error, resultFields []pgproto3.FieldDescription) (r *Result, processedAtLeastOneBatch bool, returnErr error) {
	defer trace.StartRegion(ctx, "DoltgresHandler.resultForDefaultIter").End()

	eg, ctx := ctx.NewErrgroup()

	var rowChan = make(chan sql.Row, 512)

	pan2err := func() {
		if HandlePanics {
			if recoveredPanic := recover(); recoveredPanic != nil {
				returnErr = errors.Errorf("DoltgresHandler caught panic: %v", recoveredPanic)
			}
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	// Read rows off the row iterator and send them to the row channel.
	eg.Go(func() error {
		defer pan2err()
		defer wg.Done()
		defer close(rowChan)
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				row, err := iter.Next(ctx)
				if err == io.EOF {
					return nil
				}
				if err != nil {
					return err
				}
				select {
				case rowChan <- row:
				case <-ctx.Done():
					return nil
				}
			}
		}
	})

	// Default waitTime is one minute if there is no timeout configured, in which case
	// it will loop to iterate again unless the socket died by the OS timeout or other problems.
	// If there is a timeout, it will be enforced to ensure that Vitess has a chance to
	// call DoltgresHandler.CloseConnection()
	waitTime := 1 * time.Minute
	if h.readTimeout > 0 {
		waitTime = h.readTimeout
	}
	timer := time.NewTimer(waitTime)
	defer timer.Stop()

	// reads rows from the channel, converts them to wire format,
	// and calls |callback| to give them to vitess.
	eg.Go(func() error {
		defer pan2err()
		// defer cancelF()
		defer wg.Done()
		for {
			if r == nil {
				r = &Result{Fields: resultFields}
			}
			if r.RowsAffected == rowsBatch {
				if err := callback(r); err != nil {
					return err
				}
				r = nil
				processedAtLeastOneBatch = true
				continue
			}

			select {
			case <-ctx.Done():
				return nil
			case row, ok := <-rowChan:
				if !ok {
					return nil
				}
				if types.IsOkResult(row) {
					if len(r.Rows) > 0 {
						panic("Got OkResult mixed with RowResult")
					}
					result := row[0].(types.OkResult)
					r = &Result{
						RowsAffected: result.RowsAffected,
					}
					continue
				}

				outputRow, err := rowToBytes(ctx, schema, row)
				if err != nil {
					return err
				}

				ctx.GetLogger().Tracef("spooling result row %s", outputRow)
				r.Rows = append(r.Rows, Row{outputRow})
				r.RowsAffected++
			case <-timer.C:
				if h.readTimeout != 0 {
					// Cancel and return so Vitess can call the CloseConnection callback
					ctx.GetLogger().Tracef("connection timeout")
					return errors.Errorf("row read wait bigger than connection timeout")
				}
			}
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(waitTime)
		}
	})

	// Close() kills this PID in the process list,
	// wait until all rows have be sent over the wire
	eg.Go(func() error {
		defer pan2err()
		wg.Wait()
		return iter.Close(ctx)
	})

	err := eg.Wait()
	if err != nil {
		ctx.GetLogger().WithError(err).Warn("error running query")
		returnErr = err
	}

	return
}

func rowToBytes(ctx *sql.Context, s sql.Schema, row sql.Row) ([][]byte, error) {
	if len(row) == 0 {
		return nil, nil
	}
	if len(s) == 0 {
		// should not happen
		return nil, errors.Errorf("received empty schema")
	}
	o := make([][]byte, len(row))
	for i, v := range row {
		if v == nil {
			o[i] = nil
		} else {
			val, err := s[i].Type.SQL(ctx, []byte{}, v)
			if err != nil {
				return nil, err
			}
			o[i] = val.ToBytes()
		}
	}
	return o, nil
}
