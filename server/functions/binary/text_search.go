// Copyright 2025 Dolthub, Inc.
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

package binary

import (
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initBinaryTextSearch registers the functions to the catalog.
func initBinaryTextSearch() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryTextSearch, text_ts_match)
}

// text_ts_match represents a simplified PostgreSQL @@ text search operator.
// In a full implementation, this would work with tsvector and tsquery types,
// but for now we implement basic text matching.
var text_ts_match = framework.Function2{
	Name:       "text_ts_match",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		// Basic text search: check if val2 (search term) is contained in val1 (text)
		// In PostgreSQL, this would be much more sophisticated with tsvector and tsquery
		text := strings.ToLower(val1.(string))
		searchTerm := strings.ToLower(val2.(string))
		return strings.Contains(text, searchTerm), nil
	},
}
