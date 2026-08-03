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

package functions

import (
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dprocedures"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/types"

	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

var (
	ErrDoltProcedurePermissionDenied = errors.New("permission denied for Dolt procedure")
	ErrDoltProcedureSelectOnly       = errors.New("Dolt stored procedure may only be invoked using SELECT")
)

func initDoltProcedures() {
	for _, procDef := range dprocedures.DoltProcedures {
		p, err := resolveExternalStoredProcedure(nil, procDef)
		if err != nil {
			panic(err)
		}

		outParams, returnType := doltProcedureOutParams(procDef)
		funcVal := reflect.ValueOf(procDef.Function)
		varArgCallable := varArgCallableForDoltProcedure(p, funcVal, procDef.Schema)
		noArgCallable := noArgCallableForDoltProcedure(p, funcVal, procDef.Schema)
		framework.RegisterFunction(framework.Function1{
			Name:       procDef.Name,
			Return:     returnType,
			Parameters: [1]*pgtypes.DoltgresType{pgtypes.TextArray},
			Variadic:   true,
			Callable:   varArgCallable,
			OutParams:  outParams,
		})
		framework.RegisterFunction(framework.Function0{
			Name:      procDef.Name,
			Return:    returnType,
			Callable:  noArgCallable,
			OutParams: outParams,
		})
	}
	initDynamicDoltProcedures()
}

// doltProcedureOutParams converts the GMS output schema of a Dolt stored procedure into the OUT parameter schema and
// return type of the equivalent Postgres function. A procedure with a single result column behaves like a function
// with a single OUT parameter (its return type is the column's type), while procedures with multiple result columns
// behave like functions with multiple OUT parameters (their return type is RECORD).
func doltProcedureOutParams(procDef sql.ExternalStoredProcedureDetails) (sql.Schema, *pgtypes.DoltgresType) {
	outParams := make(sql.Schema, len(procDef.Schema))
	for i, col := range procDef.Schema {
		outParams[i] = &sql.Column{
			Name:     col.Name,
			Type:     pgtypes.FromGmsType(col.Type),
			Nullable: col.Nullable,
			Source:   procDef.Name,
		}
	}
	returnType := pgtypes.Record
	if len(outParams) == 1 {
		returnType = outParams[0].Type.(*pgtypes.DoltgresType)
	}
	return outParams, returnType
}

// dynamicDoltProcedures are Dolt stored procedures that are registered with the database provider dynamically
// during engine construction (e.g. by the cluster replication controller), rather than appearing in the static
// dprocedures.DoltProcedures list. The schemas here must mirror the ones the dynamic registrations declare.
// Procedures marked srf return multiple rows and are registered as set-returning functions.
var dynamicDoltProcedures = []struct {
	name   string
	params [2]*pgtypes.DoltgresType
	schema sql.Schema
	srf    bool
}{
	{
		name:   "dolt_assume_cluster_role",
		params: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Int64},
		schema: sql.Schema{
			{Name: "status", Type: types.Int64, Nullable: false},
		},
	},
	{
		name:   "dolt_cluster_transition_to_standby",
		params: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
		schema: sql.Schema{
			{Name: "caught_up", Type: types.Int8, Nullable: false},
			{Name: "database", Type: types.LongText, Nullable: false},
			{Name: "remote", Type: types.LongText, Nullable: false},
			{Name: "remote_url", Type: types.LongText, Nullable: false},
		},
		srf: true, // returns one row per replicated database
	},
}

func initDynamicDoltProcedures() {
	for _, def := range dynamicDoltProcedures {
		outParams, returnType := doltProcedureOutParams(sql.ExternalStoredProcedureDetails{
			Name:   def.name,
			Schema: def.schema,
		})
		schema := def.schema
		name := def.name
		callableFor := varArgCallableForDoltProcedure
		if def.srf {
			// SETOF functions are declared with a row return type wrapping the type of a single row
			returnType = pgtypes.RowTypeWithReturnType(returnType)
			callableFor = srfVarArgCallableForDoltProcedure
		}
		framework.RegisterFunction(framework.Function2{
			Name:       def.name,
			Return:     returnType,
			Parameters: def.params,
			OutParams:  outParams,
			SRF:        def.srf,
			Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
				p, funcVal, err := resolveDynamicDoltProcedure(ctx, name)
				if err != nil {
					return nil, err
				}
				return callableFor(p, funcVal, schema)(ctx, [2]*pgtypes.DoltgresType{}, []any{val1, val2})
			},
		})
	}
}

// resolveDynamicDoltProcedure resolves a dynamically-registered Dolt stored procedure by name through the session
// provider's external stored procedure registry. Returns an error if no procedure is registered under the name,
// which is the case when the feature that registers it (e.g. cluster replication) isn't configured.
func resolveDynamicDoltProcedure(ctx *sql.Context, name string) (*plan.ExternalProcedure, reflect.Value, error) {
	session := dsess.DSessFromSess(ctx.Session)
	espp, ok := session.Provider().(sql.ExternalStoredProcedureProvider)
	if !ok {
		return nil, reflect.Value{}, errors.Errorf("function: '%s' not found", name)
	}
	details, err := espp.ExternalStoredProcedures(ctx, name)
	if err != nil {
		return nil, reflect.Value{}, err
	}
	if len(details) == 0 {
		return nil, reflect.Value{}, errors.Errorf("function: '%s' not found", name)
	}
	procDef := details[0]
	p, err := resolveExternalStoredProcedure(ctx, procDef)
	if err != nil {
		return nil, reflect.Value{}, err
	}
	return p, reflect.ValueOf(procDef.Function), nil
}

// varArgCallableForDoltProcedure creates a callable function that takes in a variadic number of parameters. This is
// equivalent to calling "DOLT_PROC_NAME('abc', ...)".
func varArgCallableForDoltProcedure(p *plan.ExternalProcedure, funcVal reflect.Value, outSchema sql.Schema) func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val1 any) (any, error) {
	return func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		rowIter, err := invokeDoltProcedure(ctx, p, funcVal, val1)
		if err != nil {
			return nil, err
		}
		return drainRowIter(ctx, rowIter, outSchema)
	}
}

// srfVarArgCallableForDoltProcedure is the set-returning equivalent of varArgCallableForDoltProcedure, used for the
// Dolt stored procedures that return more than one row (e.g. dolt_cluster_transition_to_standby, which returns one
// row per replicated database). It returns a row iterator over all result rows rather than the first row's value.
func srfVarArgCallableForDoltProcedure(p *plan.ExternalProcedure, funcVal reflect.Value, outSchema sql.Schema) func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val1 any) (any, error) {
	return func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		rowIter, err := invokeDoltProcedure(ctx, p, funcVal, val1)
		if err != nil {
			return nil, err
		}
		return pgtypes.NewSetReturningFunctionRowIter(func(ctx *sql.Context) (sql.Row, error) {
			row, err := rowIter.Next(ctx)
			if err != nil {
				closeErr := rowIter.Close(ctx)
				if err == io.EOF {
					if closeErr != nil {
						return nil, closeErr
					}
					return nil, io.EOF
				}
				return nil, err
			}
			return convertProcedureRow(ctx, row, outSchema)
		}), nil
	}
}

// invokeDoltProcedure invokes the given Dolt stored procedure with the given variadic argument values, returning its
// result row iterator.
func invokeDoltProcedure(ctx *sql.Context, p *plan.ExternalProcedure, funcVal reflect.Value, val1 any) (sql.RowIter, error) {
	funcType := funcVal.Type()

	err := checkDoltProcedureAccess(ctx, p)
	if err != nil {
		return nil, err
	}

	values, ok := val1.([]any)
	if !ok {
		return nil, sql.ErrExternalProcedureInvalidParamType.New(reflect.TypeOf(val1).String())
	}

	funcParams := make([]reflect.Value, len(values)+1)
	funcParams[0] = reflect.ValueOf(ctx)

	if !funcType.IsVariadic() && len(values) != len(p.ParamDefinitions) {
		return nil, errors.Errorf("function '%s' expects %d parameters, %d were provided",
			p.Name, len(p.ParamDefinitions), len(values))
	}

	for i := range values {
		var paramDefinition plan.ProcedureParam
		var funcParamType reflect.Type
		if funcType.IsVariadic() {
			paramDefinition = p.ParamDefinitions[0]
			funcParamType = funcType.In(funcType.NumIn() - 1).Elem()
		} else {
			paramDefinition = p.ParamDefinitions[i]
			funcParamType = funcType.In(i + 1)
		}

		// Grab the passed-in variable and convert it to the type we expect
		exprParamVal, _, err := paramDefinition.Type.Convert(ctx, values[i])
		if err != nil {
			return nil, err
		}

		funcParams[i+1], err = p.ProcessParam(ctx, funcParamType, exprParamVal)
		if err != nil {
			return nil, err
		}
	}

	out := funcVal.Call(funcParams)
	if err, ok := out[1].Interface().(error); ok { // Only evaluates to true when error is not nil
		return nil, err
	}

	if iter, ok := out[0].Interface().(sql.RowIter); ok {
		return iter, nil
	}
	return sql.RowsToRowIter(), nil
}

// noArgCallableForDoltProcedure creates a callable function that does not take any parameters. This is equivalent to
// calling "DOLT_PROC_NAME()".
func noArgCallableForDoltProcedure(p *plan.ExternalProcedure, funcVal reflect.Value, outSchema sql.Schema) func(ctx *sql.Context) (any, error) {
	return func(ctx *sql.Context) (any, error) {
		err := checkDoltProcedureAccess(ctx, p)
		if err != nil {
			return nil, err
		}

		funcParams := []reflect.Value{reflect.ValueOf(ctx)}
		out := funcVal.Call(funcParams)
		if err, ok := out[1].Interface().(error); ok { // Only evaluates to true when error is not nil
			return nil, err
		}

		var rowIter sql.RowIter
		if iter, ok := out[0].Interface().(sql.RowIter); ok {
			rowIter = iter
		} else {
			rowIter = sql.RowsToRowIter()
		}
		return drainRowIter(ctx, rowIter, outSchema)
	}
}

// checkDoltProcedureAccess ensures the current user is authorized as a SUPERUSER if the given |procedure| requires
// admin, and that the server is not read-only if the procedure writes.
func checkDoltProcedureAccess(ctx *sql.Context, procedure *plan.ExternalProcedure) error {
	if !procedure.ReadOnly {
		if _, readOnly, ok := sql.SystemVariables.GetGlobal("read_only"); ok {
			ro, err := sql.ConvertToBool(ctx, readOnly)
			if err != nil {
				return err
			}
			if ro {
				return sql.ErrReadOnly.New()
			}
		}
	}

	if !procedure.AdminOnly {
		return nil
	}

	var userRole auth.Role
	auth.LockRead(func() {
		userRole = auth.GetRole(ctx.Client().User)
	})

	if !userRole.IsValid() || !userRole.IsSuperUser {
		return ErrDoltProcedurePermissionDenied
	}

	return nil
}

// drainRowIter reads the single result row of a Dolt stored procedure and converts it to the value the equivalent
// Postgres function would return: the bare value for a single-column schema, or a record value for a multi-column
// schema (matching a function with multiple OUT parameters). Procedures that return multiple rows must be registered
// as set-returning functions using srfVarArgCallableForDoltProcedure instead.
func drainRowIter(ctx *sql.Context, rowIter sql.RowIter, outSchema sql.Schema) (any, error) {
	defer rowIter.Close(ctx)

	row, err := rowIter.Next(ctx)
	if err == io.EOF {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	converted, err := convertProcedureRow(ctx, row, outSchema)
	if err != nil {
		return nil, err
	}
	if len(converted) == 1 {
		return converted[0], nil
	}

	values := make([]pgtypes.RecordValue, len(converted))
	for i := range converted {
		values[i] = pgtypes.RecordValue{
			Value: converted[i],
			Type:  pgtypes.FromGmsType(outSchema[i].Type),
		}
	}
	return values, nil
}

// convertProcedureRow converts a result row from a Dolt stored procedure into the Go values that the corresponding
// Postgres OUT parameter types expect, per the procedure's declared GMS output schema.
func convertProcedureRow(ctx *sql.Context, row sql.Row, outSchema sql.Schema) (sql.Row, error) {
	if len(row) != len(outSchema) {
		return nil, errors.Errorf("dolt_procedures: expected %d result columns, got %d", len(outSchema), len(row))
	}
	converted := make(sql.Row, len(row))
	for i := range row {
		val := row[i]
		if val != nil {
			var err error
			val, _, err = outSchema[i].Type.Convert(ctx, val)
			if err != nil {
				return nil, err
			}
			val, err = coerceToPostgresValue(val, pgtypes.FromGmsType(outSchema[i].Type))
			if err != nil {
				return nil, err
			}
		}
		converted[i] = val
	}
	return converted, nil
}

// coerceToPostgresValue converts a value normalized by a GMS type into the Go representation that the given Doltgres
// type expects. The GMS integer types are narrower than their Doltgres equivalents (e.g. a GMS TINYINT normalizes to
// an int8, while its Doltgres equivalent INT4 expects an int32), so integer values must be widened to match.
func coerceToPostgresValue(val any, pgType *pgtypes.DoltgresType) (any, error) {
	switch pgType.ID {
	case pgtypes.Int16.ID, pgtypes.Int32.ID, pgtypes.Int64.ID:
		var intVal int64
		switch v := val.(type) {
		case int8:
			intVal = int64(v)
		case int16:
			intVal = int64(v)
		case int32:
			intVal = int64(v)
		case int64:
			intVal = v
		case uint8:
			intVal = int64(v)
		case uint16:
			intVal = int64(v)
		case uint32:
			intVal = int64(v)
		default:
			return nil, errors.Errorf("dolt_procedures: unsupported value %T for type %s", val, pgType.String())
		}
		switch pgType.ID {
		case pgtypes.Int16.ID:
			return int16(intVal), nil
		case pgtypes.Int32.ID:
			return int32(intVal), nil
		default:
			return intVal, nil
		}
	case pgtypes.Bool.ID:
		switch v := val.(type) {
		case bool:
			return v, nil
		case int8:
			return v != 0, nil
		default:
			return nil, errors.Errorf("dolt_procedures: unsupported value %T for type %s", val, pgType.String())
		}
	default:
		return val, nil
	}
}

var (
	// ctxType is the reflect.Type of a *sql.Context.
	ctxType = reflect.TypeOf((*sql.Context)(nil))
	// ctxType is the reflect.Type of a sql.RowIter.
	rowIterType = reflect.TypeOf((*sql.RowIter)(nil)).Elem()
	// ctxType is the reflect.Type of an error.
	errorType = reflect.TypeOf((*error)(nil)).Elem()
	// externalStoredProcedurePointerTypes maps a non-pointer type to a sql.Type for external stored procedures.
	externalStoredProcedureTypes = map[reflect.Type]sql.Type{
		reflect.TypeOf(int(0)):      types.Int64,
		reflect.TypeOf(int8(0)):     types.Int8,
		reflect.TypeOf(int16(0)):    types.Int16,
		reflect.TypeOf(int32(0)):    types.Int32,
		reflect.TypeOf(int64(0)):    types.Int64,
		reflect.TypeOf(uint(0)):     types.Uint64,
		reflect.TypeOf(uint8(0)):    types.Uint8,
		reflect.TypeOf(uint16(0)):   types.Uint16,
		reflect.TypeOf(uint32(0)):   types.Uint32,
		reflect.TypeOf(uint64(0)):   types.Uint64,
		reflect.TypeOf(float32(0)):  types.Float32,
		reflect.TypeOf(float64(0)):  types.Float64,
		reflect.TypeOf(bool(false)): types.Int8,
		reflect.TypeOf(string("")):  types.LongText,
		reflect.TypeOf([]byte{}):    types.LongBlob,
		reflect.TypeOf(time.Time{}): types.DatetimeMaxPrecision,
	}
	// externalStoredProcedurePointerTypes maps a pointer type to a sql.Type for external stored procedures.
	externalStoredProcedurePointerTypes = map[reflect.Type]sql.Type{
		reflect.TypeOf((*int)(nil)):       types.Int64,
		reflect.TypeOf((*int8)(nil)):      types.Int8,
		reflect.TypeOf((*int16)(nil)):     types.Int16,
		reflect.TypeOf((*int32)(nil)):     types.Int32,
		reflect.TypeOf((*int64)(nil)):     types.Int64,
		reflect.TypeOf((*uint)(nil)):      types.Uint64,
		reflect.TypeOf((*uint8)(nil)):     types.Uint8,
		reflect.TypeOf((*uint16)(nil)):    types.Uint16,
		reflect.TypeOf((*uint32)(nil)):    types.Uint32,
		reflect.TypeOf((*uint64)(nil)):    types.Uint64,
		reflect.TypeOf((*float32)(nil)):   types.Float32,
		reflect.TypeOf((*float64)(nil)):   types.Float64,
		reflect.TypeOf((*bool)(nil)):      types.Int8,
		reflect.TypeOf((*string)(nil)):    types.LongText,
		reflect.TypeOf((*[]byte)(nil)):    types.LongBlob,
		reflect.TypeOf((*time.Time)(nil)): types.DatetimeMaxPrecision,
	}
)

func init() {
	if strconv.IntSize == 32 {
		externalStoredProcedureTypes[reflect.TypeOf(int(0))] = types.Int32
		externalStoredProcedureTypes[reflect.TypeOf(uint(0))] = types.Uint32
		externalStoredProcedurePointerTypes[reflect.TypeOf((*int)(nil))] = types.Int32
		externalStoredProcedurePointerTypes[reflect.TypeOf((*uint)(nil))] = types.Uint32
	}
}

func resolveExternalStoredProcedure(_ *sql.Context, externalProcedure sql.ExternalStoredProcedureDetails) (*plan.ExternalProcedure, error) {
	funcVal := reflect.ValueOf(externalProcedure.Function)
	funcType := funcVal.Type()
	if funcType.Kind() != reflect.Func {
		return nil, sql.ErrExternalProcedureNonFunction.New(externalProcedure.Function)
	}
	if funcType.NumIn() == 0 {
		return nil, sql.ErrExternalProcedureMissingContextParam.New()
	}
	if funcType.NumOut() != 2 {
		return nil, sql.ErrExternalProcedureReturnTypes.New()
	}
	if funcType.In(0) != ctxType {
		return nil, sql.ErrExternalProcedureMissingContextParam.New()
	}
	if funcType.Out(0) != rowIterType {
		return nil, sql.ErrExternalProcedureFirstReturn.New()
	}
	if funcType.Out(1) != errorType {
		return nil, sql.ErrExternalProcedureSecondReturn.New()
	}
	funcIsVariadic := funcType.IsVariadic()

	paramDefinitions := make([]plan.ProcedureParam, funcType.NumIn()-1)
	paramReferences := make([]*expression.ProcedureParam, len(paramDefinitions))
	for i := 0; i < len(paramDefinitions); i++ {
		funcParamType := funcType.In(i + 1)
		paramName := "A" + strconv.FormatInt(int64(i), 10)
		paramIsVariadic := false
		if funcIsVariadic && i == len(paramDefinitions)-1 {
			paramIsVariadic = true
			funcParamType = funcParamType.Elem()
			if funcParamType.Kind() == reflect.Ptr {
				return nil, sql.ErrExternalProcedurePointerVariadic.New()
			}
		}

		if sqlType, ok := externalStoredProcedureTypes[funcParamType]; ok {
			paramDefinitions[i] = plan.ProcedureParam{
				Direction: plan.ProcedureParamDirection_In,
				Name:      paramName,
				Type:      sqlType,
				Variadic:  paramIsVariadic,
			}
			paramReferences[i] = expression.NewProcedureParam(paramName, sqlType)
		} else if sqlType, ok = externalStoredProcedurePointerTypes[funcParamType]; ok {
			paramDefinitions[i] = plan.ProcedureParam{
				Direction: plan.ProcedureParamDirection_Inout,
				Name:      paramName,
				Type:      sqlType,
				Variadic:  paramIsVariadic,
			}
			paramReferences[i] = expression.NewProcedureParam(paramName, sqlType)
		} else {
			return nil, sql.ErrExternalProcedureInvalidParamType.New(funcParamType.String())
		}
	}

	return &plan.ExternalProcedure{
		ExternalStoredProcedureDetails: externalProcedure,
		ParamDefinitions:               paramDefinitions,
		Params:                         paramReferences,
	}, nil
}
