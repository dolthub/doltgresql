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
)

// initXmlIsWellFormed registers the functions to the catalog.
func initXmlIsWellFormed() {
	framework.RegisterFunction(xml_is_well_formed_text)
	framework.RegisterFunction(xml_is_well_formed_document_text)
	framework.RegisterFunction(xml_is_well_formed_content_text)
}

// xml_is_well_formed_text represents the PostgreSQL function of the same name, taking the same parameters.
var xml_is_well_formed_text = framework.Function1{
	Name:       "xml_is_well_formed",
	Return:     pgtypes.Bool,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		document, err := xmlOptionIsDocument(ctx)
		if err != nil {
			return nil, err
		}
		return xmlIsWellFormed(ctx, val1, document)
	},
}

// xml_is_well_formed_document_text represents the PostgreSQL function of the same name, taking the same parameters.
var xml_is_well_formed_document_text = framework.Function1{
	Name:       "xml_is_well_formed_document",
	Return:     pgtypes.Bool,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		return xmlIsWellFormed(ctx, val1, true)
	},
}

// xml_is_well_formed_content_text represents the PostgreSQL function of the same name, taking the same parameters.
var xml_is_well_formed_content_text = framework.Function1{
	Name:       "xml_is_well_formed_content",
	Return:     pgtypes.Bool,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		return xmlIsWellFormed(ctx, val1, false)
	},
}

// xmlIsWellFormed returns whether `val` is well-formed XML content, or a well-formed XML document when `document` is
// set.
func xmlIsWellFormed(ctx *sql.Context, val any, document bool) (any, error) {
	str, err := framework.UnwrapString(ctx, val)
	if err != nil {
		return nil, err
	}
	return xml.CheckWellFormed(str, document) == nil, nil
}
