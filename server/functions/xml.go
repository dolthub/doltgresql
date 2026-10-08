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

package functions

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/xml"
	"github.com/dolthub/doltgresql/utils"
)

// initXml registers the functions to the catalog.
func initXml() {
	framework.RegisterFunction(xml_in)
	framework.RegisterFunction(xml_out)
	framework.RegisterFunction(xml_recv)
	framework.RegisterFunction(xml_send)
}

// xml_in represents the PostgreSQL function of xml type IO input.
var xml_in = framework.Function1{
	Name:       "xml_in",
	Return:     pgtypes.Xml,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input, err := framework.UnwrapString(ctx, val)
		if err != nil {
			return nil, err
		}
		document, err := xmlOptionIsDocument(ctx)
		if err != nil {
			return nil, err
		}
		if err = xml.CheckWellFormed(input, document); err != nil {
			return nil, err
		}
		return input, nil
	},
}

// xml_out represents the PostgreSQL function of xml type IO output.
var xml_out = framework.Function1{
	Name:       "xml_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Xml},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		str, err := framework.UnwrapString(ctx, val)
		if err != nil {
			return nil, err
		}
		return xml.Output(str), nil
	},
}

// xml_recv represents the PostgreSQL function of xml type IO receive.
var xml_recv = framework.Function1{
	Name:       "xml_recv",
	Return:     pgtypes.Xml,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data, err := framework.UnwrapBytes(ctx, val)
		if err != nil {
			return nil, err
		}
		if data == nil {
			return nil, nil
		}
		return xml_in.Callable(ctx, [2]*pgtypes.DoltgresType{}, string(data))
	},
}

// xml_send represents the PostgreSQL function of xml type IO send.
var xml_send = framework.Function1{
	Name:       "xml_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Xml},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		str, err := framework.UnwrapString(ctx, val)
		if err != nil {
			return nil, err
		}
		writer := utils.NewWireWriter()
		writer.WriteString(xml.Output(str))
		return writer.BufferData(), nil
	},
}

// xmlOptionIsDocument returns whether the `xmloption` session variable is set to `document`.
func xmlOptionIsDocument(ctx *sql.Context) (bool, error) {
	val, err := ctx.GetSessionVariable(ctx, "xmloption")
	if err != nil {
		return false, err
	}
	return val.(string) == "document", nil
}
