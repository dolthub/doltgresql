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
	"github.com/antchfx/xpath"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initXpathExists registers the functions to the catalog.
func initXpathExists() {
	framework.RegisterFunction(xpath_exists_text_xml)
	framework.RegisterFunction(xpath_exists_text_xml_textarray)
}

// xpath_exists_text_xml represents the PostgreSQL function of the same name, taking the same parameters.
var xpath_exists_text_xml = framework.Function2{
	Name:       "xpath_exists",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Xml},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return xpath_exists_text_xml_textarray.Callable(ctx, [4]*pgtypes.DoltgresType{}, val1, val2, []any{})
	},
}

// xpath_exists_text_xml_textarray represents the PostgreSQL function of the same name, taking the same parameters.
var xpath_exists_text_xml_textarray = framework.Function3{
	Name:       "xpath_exists",
	Return:     pgtypes.Bool,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Xml, pgtypes.TextArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1 any, val2 any, val3 any) (any, error) {
		result, err := evaluateXpath(ctx, val1, val2, val3)
		if err != nil {
			return nil, err
		}
		if nodes, ok := result.(*xpath.NodeIterator); ok {
			return nodes.MoveNext(), nil
		}
		return true, nil
	},
}
