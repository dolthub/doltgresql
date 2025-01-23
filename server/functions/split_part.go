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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initSplitPart registers the functions to the catalog.
func initSplitPart() {
	framework.RegisterFunction(split_part_text_text_int32)
}

// split_part_text_text_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var split_part_text_text_int32 = framework.Function3{
	Name:       "split_part",
	Return:     pgtypes.Text,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, str any, delimiter any, n any) (any, error) {
		if n.(int32) == 0 {
			return nil, errors.Errorf("field position must not be zero")
		}
		var parts []string
		if len(delimiter.(string)) > 0 {
			parts = strings.Split(str.(string), delimiter.(string))
		} else {
			parts = []string{str.(string)}
		}
		if int(utils.Abs(n.(int32))) > len(parts) {
			return "", nil
		}
		if n.(int32) > 0 {
			return parts[n.(int32)-1], nil
		} else {
			return parts[int32(len(parts))+n.(int32)], nil
		}
	},
}
