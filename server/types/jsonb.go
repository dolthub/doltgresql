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

package types

import (
	"github.com/dolthub/go-mysql-server/sql"
	gmstypes "github.com/dolthub/go-mysql-server/sql/types"
	"github.com/goccy/go-json"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/utils"
)

// JsonB is the deserialized and structured version of JSON that deals with JsonDocument.
var JsonB = &DoltgresType{
	ID:                  toInternal("jsonb"),
	TypLength:           int16(-1),
	PassedByVal:         false,
	TypType:             TypeType_Base,
	TypCategory:         TypeCategory_UserDefinedTypes,
	IsPreferred:         false,
	IsDefined:           true,
	Delimiter:           ",",
	RelID:               id.Null,
	SubscriptFunc:       toFuncID("jsonb_subscript_handler", toInternal("internal")),
	Elem:                id.NullType,
	Array:               toInternal("_jsonb"),
	InputFunc:           toFuncID("jsonb_in", toInternal("cstring")),
	OutputFunc:          toFuncID("jsonb_out", toInternal("jsonb")),
	ReceiveFunc:         toFuncID("jsonb_recv", toInternal("internal")),
	SendFunc:            toFuncID("jsonb_send", toInternal("jsonb")),
	ModInFunc:           toFuncID("-"),
	ModOutFunc:          toFuncID("-"),
	AnalyzeFunc:         toFuncID("-"),
	Align:               TypeAlignment_Int,
	Storage:             TypeStorage_Extended,
	NotNull:             false,
	BaseTypeID:          id.NullType,
	TypMod:              -1,
	NDims:               0,
	TypCollation:        id.NullCollation,
	DefaulBin:           "",
	Default:             "",
	Acl:                 nil,
	Checks:              nil,
	attTypMod:           -1,
	CompareFunc:         toFuncID("jsonb_cmp", toInternal("jsonb"), toInternal("jsonb")),
	SerializationFunc:   serializeTypeJsonB,
	DeserializationFunc: deserializeTypeJsonB,
}

// serializeTypeJsonB handles serialization from the standard representation to our serialized representation that is
// written in Dolt. This is used for the legacy ExtendedEnc storage path.
func serializeTypeJsonB(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	res, err := sql.UnwrapAny(ctx, val)
	if err != nil {
		return nil, err
	}
	switch v := res.(type) {
	case sql.JSONWrapper:
		j, err := v.ToInterface(ctx)
		if err != nil {
			return nil, err
		}
		return json.Marshal(j)
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(v)
	}
}

// deserializeTypeJsonB handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes. This is used for the legacy ExtendedEnc storage path.
func deserializeTypeJsonB(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		// Fall back to the legacy deserialization format for old data
		reader := utils.NewReader(data)
		jsonValue, err2 := JsonValueDeserialize(reader)
		if err2 != nil {
			return nil, err
		}
		// Convert from legacy JsonValue format to a types.JSONDocument
		return gmstypes.JSONDocument{Val: jsonValueToInterface(jsonValue)}, nil
	}
	return gmstypes.JSONDocument{Val: v}, nil
}

// jsonValueToInterface converts a legacy JsonValue to a native Go interface value.
func jsonValueToInterface(value JsonValue) any {
	switch v := value.(type) {
	case JsonValueObject:
		obj := make(map[string]any, len(v.Items))
		for _, item := range v.Items {
			obj[item.Key] = jsonValueToInterface(item.Value)
		}
		return obj
	case JsonValueArray:
		arr := make([]any, len(v))
		for i, item := range v {
			arr[i] = jsonValueToInterface(item)
		}
		return arr
	case JsonValueString:
		return string(v)
	case JsonValueNumber:
		f, _ := decimal.Decimal(v).Float64()
		return f
	case JsonValueBoolean:
		return bool(v)
	case JsonValueNull:
		return nil
	default:
		return nil
	}
}
