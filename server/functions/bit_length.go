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
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initBitLength registers the functions to the catalog.
func initBitLength() {
	framework.RegisterFunction(bit_length_varchar)
}

// bit_length_varchar represents the PostgreSQL function of the same name, taking the same parameters.
var bit_length_varchar = framework.Function1{
	Name:       "bit_length",
	Return:     pgtypes.Int32,
	Parameters: []pgtypes.DoltgresType{pgtypes.VarChar},
	Callable: func(ctx framework.Context, val1 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		result, err := octet_length_varchar.Callable(ctx, val1)
		if err != nil {
			return nil, err
		}
		return result.(int32) * 8, nil
	},
}
