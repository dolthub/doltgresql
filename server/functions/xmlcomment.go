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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initXmlcomment registers the functions to the catalog.
func initXmlcomment() {
	framework.RegisterFunction(xmlcomment_text)
}

// xmlcomment_text represents the PostgreSQL function of the same name, taking the same parameters.
var xmlcomment_text = framework.Function1{
	Name:       "xmlcomment",
	Return:     pgtypes.Xml,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		str, err := framework.UnwrapString(ctx, val1)
		if err != nil {
			return nil, err
		}
		if strings.Contains(str, "--") || strings.HasSuffix(str, "-") {
			return nil, pgerror.New(pgcode.InvalidXMLComment, "invalid XML comment")
		}
		return "<!--" + str + "-->", nil
	},
}
