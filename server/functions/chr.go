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
	"unicode/utf8"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initChr registers the functions to the catalog.
func initChr() {
	framework.RegisterFunction(chr_int32)
}

// chr_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var chr_int32 = framework.Function1{
	Name:       "chr",
	Return:     pgtypes.Text,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val1Interface any) (any, error) {
		val1 := val1Interface.(int32)
		if val1 == 0 {
			return nil, fmt.Errorf("null character not permitted")
		} else if val1 < 0 {
			return nil, fmt.Errorf("character number must be positive")
		}
		r := rune(val1)
		if !utf8.ValidRune(r) {
			return nil, fmt.Errorf("requested character too large for encoding")
		}
		return string(r), nil
	},
}
