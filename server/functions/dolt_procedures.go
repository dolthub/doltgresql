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
	"reflect"
	"strconv"
	"time"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dprocedures"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/types"
)

func initDoltProcedures() {
	for _, procDef := range dprocedures.DoltProcedures {
		funcVal := reflect.ValueOf(procDef.Function)
		funcType := funcVal.Type()
		p, err := resolveExternalStoredProcedure(nil, procDef)
		if err != nil {
			panic(err)
		}

		callable := func(ctx *sql.Context, values ...any) (any, error) {
			var funcParams []reflect.Value
			funcParams = append(funcParams, reflect.ValueOf(ctx))

			for i := range p.Params {
				paramDefinition := p.ParamDefinitions[i]
				var funcParamType reflect.Type
				if paramDefinition.Variadic {
					funcParamType = funcType.In(funcType.NumIn() - 1).Elem()
				} else {
					funcParamType = funcType.In(i + 1)
				}

				// Grab the passed-in variable and convert it to the type we expect
				exprParamVal, err := p.Params[i].Eval(ctx, nil)
				if err != nil {
					return nil, err
				}
				exprParamVal, _, err = paramDefinition.Type.Convert(exprParamVal)
				if err != nil {
					return nil, err
				}

				funcParams[i+1], err = p.ProcessParam(ctx, funcParamType, exprParamVal)
				if err != nil {
					return nil, err
				}
			}
			rowIter := funcVal.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(values)})
			// TODO: drain the iter, don't return it
			return rowIter, nil
		}

		framework.RegisterFunction(framework.FunctionN{
			Name:        procDef.Name,
			Return:      pgtypes.Int64, // TODO: real type
			Parameters:  make([]pgtypes.DoltgresType, 0),
			VarargsType: pgtypes.TextType{},
			Callable:    callable,
		})
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