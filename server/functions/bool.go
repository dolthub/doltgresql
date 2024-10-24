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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initBinaryNotEqual registers the functions to the catalog.
func initUnaryTypeIn() {
	framework.RegisterFunction(boolin)
	framework.RegisterFunction(boolout)
	framework.RegisterFunction(boolrecv)
	framework.RegisterFunction(boolsend)
}

// boolin represents the PostgreSQL function of boolean type ioInput().
var boolin = framework.Function1{
	Name:       "boolin",
	Return:     pgtypes.Bool,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, input any) (any, error) {
		input = strings.TrimSpace(strings.ToLower(input.(string)))
		if input == "true" || input == "t" || input == "yes" || input == "on" || input == "1" {
			return true, nil
		} else if input == "false" || input == "f" || input == "no" || input == "off" || input == "0" {
			return false, nil
		} else {
			return nil, pgtypes.ErrInvalidSyntaxForType.New("boolean", input)
		}
	},
}

// booout represents the PostgreSQL function of boolean type ioOutput().
var boolout = framework.Function1{
	Name:       "boolout",
	Return:     pgtypes.Bool,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Bool},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, input any) (any, error) {
		// TODO: should the input be converted or should be converted here?
		if input.(bool) {
			return "true", nil
		} else {
			return "false", nil
		}
	},
}

// boorecv represents the PostgreSQL function of the same name, taking the same parameters.
var boolrecv = framework.Function1{
	Name:       "boolrecv",
	Return:     pgtypes.Bool,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Bool},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, input any) (any, error) {
		switch v := input.(type) {
		case bool:
			return v, nil
		case nil:
			return nil, nil
		default:
			return nil, pgtypes.ErrUnhandledType.New("boolean", v)
		}
	},
}

// boosend represents the PostgreSQL function of the same name, taking the same parameters.
var boolsend = framework.Function1{
	Name:       "boolsend",
	Return:     pgtypes.Text, // TODO: should it be bytea
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Bool},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, input any) (any, error) {
		// TODO: should the input be result of ioOutput or should be done here?
		valBytes := types.AppendAndSliceBytes(nil, []byte{input.(string)[0]})
		return string(valBytes), nil
	},
}
