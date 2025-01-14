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
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initEnum registers the functions to the catalog.
func initEnum() {
	framework.RegisterFunction(enum_in)
	framework.RegisterFunction(enum_out)
	framework.RegisterFunction(enum_recv)
	framework.RegisterFunction(enum_send)
	framework.RegisterFunction(enum_cmp)
}

// enum_in represents the PostgreSQL function of enum type IO input.
var enum_in = framework.Function2{
	Name:       "enum_in",
	Return:     pgtypes.AnyEnum,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Cstring, pgtypes.Oid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		typ, err := getDoltgresTypeFromId(ctx, val2.(id.Id))
		if err != nil {
			return nil, err
		}
		if typ.TypCategory != pgtypes.TypeCategory_EnumTypes {
			return nil, fmt.Errorf(`"%s" is not an enum type`, typ.Name())
		}

		value := val1.(string)
		if _, exists := typ.EnumLabels[value]; !exists {
			return nil, pgtypes.ErrInvalidInputValueForEnum.New(typ.Name(), value)
		}
		// TODO: should return the index instead of label?
		return value, nil
	},
}

// enum_out represents the PostgreSQL function of enum type IO output.
var enum_out = framework.Function1{
	Name:       "enum_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyEnum},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// TODO: should receive the index instead of label?
		return val.(string), nil
	},
}

// enum_recv represents the PostgreSQL function of enum type IO receive.
var enum_recv = framework.Function2{
	Name:       "enum_recv",
	Return:     pgtypes.AnyEnum,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		// TODO: should return the index instead of label?
		data := val1.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		reader := utils.NewReader(data)
		value := reader.String()
		if ctx == nil {
			// TODO: currently, in some places we use nil context, should fix it.
			return value, nil
		}
		typ, err := getDoltgresTypeFromId(ctx, val2.(id.Id))
		if err != nil {
			return nil, err
		}
		if typ.TypCategory != pgtypes.TypeCategory_EnumTypes {
			return nil, fmt.Errorf(`"%s" is not an enum type`, typ.Name())
		}
		if _, exists := typ.EnumLabels[value]; !exists {
			return nil, pgtypes.ErrInvalidInputValueForEnum.New(typ.Name(), value)
		}
		return value, nil
	},
}

// enum_send represents the PostgreSQL function of enum type IO send.
var enum_send = framework.Function1{
	Name:       "enum_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyEnum},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// TODO: should return the index instead of label?
		str := val.(string)
		writer := utils.NewWriter(uint64(len(str) + 4))
		writer.String(str)
		return writer.Data(), nil
	},
}

// enum_cmp represents the PostgreSQL function of enum type compare.
var enum_cmp = framework.Function2{
	Name:       "enum_cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyEnum, pgtypes.AnyEnum},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(string)
		bb := val2.(string)
		enumType := t[0]
		if enumType.EnumLabels == nil {
			return nil, fmt.Errorf(`enum label lookup failed for type %s`, enumType.Name())
		}
		abLabel, ok := enumType.EnumLabels[ab]
		if !ok {
			return nil, pgtypes.ErrInvalidInputValueForEnum.New(enumType.Name(), ab)
		}
		bbLabel, ok := enumType.EnumLabels[bb]
		if !ok {
			return nil, pgtypes.ErrInvalidInputValueForEnum.New(enumType.Name(), bb)
		}
		if abLabel.SortOrder == bbLabel.SortOrder {
			return int32(0), nil
		} else if abLabel.SortOrder < bbLabel.SortOrder {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}

// getDoltgresTypeFromId takes internal ID and returns the DoltgresType associated to it. It allows retrieving a
// user-defined type, and it requires a valid sql.Context.
func getDoltgresTypeFromId(ctx *sql.Context, rawId id.Id) (*pgtypes.DoltgresType, error) {
	typCol, err := core.GetTypesCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	typID := id.Type(rawId)

	schName := typID.SchemaName()
	sch, err := core.GetCurrentSchema(ctx)
	if err != nil {
		return nil, err
	}
	if schName == "" {
		schName = sch
	}

	typName := typID.TypeName()
	typ, found := typCol.GetType(id.NewType(schName, typName))
	if !found {
		return nil, pgtypes.ErrTypeDoesNotExist.New(typName)
	}
	return typ, nil
}
