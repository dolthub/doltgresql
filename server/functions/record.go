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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"

	"github.com/dolthub/doltgresql/core"

	"github.com/dolthub/doltgresql/core/id"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initRecord registers the functions to the catalog.
func initRecord() {
	framework.RegisterFunction(record_in)
	framework.RegisterFunction(record_out)
	framework.RegisterFunction(record_recv)
	framework.RegisterFunction(record_send)
	framework.RegisterFunction(btrecordcmp)
	framework.RegisterFunction(btrecordimagecmp)
}

// record_in represents the PostgreSQL function of record type IO input.
var record_in = framework.Function3{
	Name:       "record_in",
	Return:     pgtypes.Record,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Cstring, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		return nil, errors.Errorf("record_in not implemented")
	},
}

// record_out represents the PostgreSQL function of record type IO output.
var record_out = framework.Function1{
	Name:       "record_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Record},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		values, ok := val.([]pgtypes.RecordValue)
		if !ok {
			return nil, errors.Errorf("expected []RecordValue, but got %T", val)
		}
		return pgtypes.RecordToString(ctx, values)
	},
}

// record_recv represents the PostgreSQL function of record type IO receive. The input of this function is expected to
// be the output of record_send.
var record_recv = framework.Function3{
	Name:       "record_recv",
	Return:     pgtypes.Record,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable:   record_recv_callable,
}

// record_recv_callable is the function definition of record_recv.
func record_recv_callable(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
	data := val1.([]byte)
	if data == nil {
		return nil, nil
	}
	typeColl, err := core.GetTypesCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	reader := utils.NewWireReader(data)
	count := reader.ReadInt32()
	recordVals := make([]pgtypes.RecordValue, count)
	for i := range recordVals {
		typeID := id.Type(id.Cache().ToInternal(reader.ReadUint32()))
		recordType, err := typeColl.GetType(ctx, typeID)
		if err != nil {
			return nil, err
		}
		if recordType == nil {
			return nil, pgtypes.ErrTypeDoesNotExist.New(typeID.TypeName())
		}
		valLen := reader.ReadInt32()
		var recordVal any
		if valLen != -1 {
			valBytes := reader.ReadBytes(uint32(valLen))
			recordVal, err = recordType.CallReceive(ctx, valBytes)
			if err != nil {
				return nil, err
			}
		}
		recordVals[i] = pgtypes.RecordValue{
			Value: recordVal,
			Type:  recordType,
		}
	}
	return recordVals, nil
}

// record_send represents the PostgreSQL function of record type IO send. The output of this function is expected to
// be the input of record_recv.
var record_send = framework.Function1{
	Name:       "record_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Record},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		if wrapper, ok := val.(sql.AnyWrapper); ok {
			var err error
			val, err = wrapper.UnwrapAny(ctx)
			if err != nil {
				return nil, err
			}
			if val == nil {
				return nil, nil
			}
		}
		recordVals := val.([]pgtypes.RecordValue)
		writer := utils.NewWireWriter()
		writer.WriteInt32(int32(len(recordVals)))
		for _, recordVal := range recordVals {
			switch recordType := recordVal.Type.(type) {
			case *pgtypes.DoltgresType:
				writer.WriteUint32(id.Cache().ToOID(recordType.ID.AsId()))
				if recordVal.Value != nil {
					valBytes, err := recordType.CallSend(ctx, recordVal.Value)
					if err != nil {
						return nil, err
					}
					writer.WriteInt32(int32(len(valBytes)))
					writer.WriteBytes(valBytes)
				} else {
					writer.WriteInt32(-1)
				}
			default:
				cast := pgexprs.NewGMSCast(expression.NewLiteral(recordVal.Value, recordType))
				writer.WriteUint32(id.Cache().ToOID(cast.DoltgresType(ctx).ID.AsId()))
				if recordVal.Value != nil {
					castVal, err := cast.Eval(ctx, nil)
					if err != nil {
						return nil, err
					}
					valBytes, err := cast.DoltgresType(ctx).CallSend(ctx, castVal)
					if err != nil {
						return nil, err
					}
					writer.WriteInt32(int32(len(valBytes)))
					writer.WriteBytes(valBytes)
				} else {
					writer.WriteInt32(-1)
				}
			}
		}
		return writer.BufferData(), nil
	},
}

// btrecordcmp represents the PostgreSQL function of record type compare.
var btrecordcmp = framework.Function2{
	Name:       "btrecordcmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Record, pgtypes.Record},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		// TODO
		ab := val1.(string)
		bb := val2.(string)
		if ab == bb {
			return int32(0), nil
		} else if ab < bb {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}

// btrecordimagecmp represents the PostgreSQL function of record type compare.
var btrecordimagecmp = framework.Function2{
	Name:       "btrecordimagecmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Record, pgtypes.Record},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		// TODO
		return int32(1), nil
	},
}
