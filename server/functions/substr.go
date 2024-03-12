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

// init registers the functions to the catalog.
func init() {
	framework.RegisterFunction(substr_varchar_int32)
	framework.RegisterFunction(substr_varchar_int32_int32)
}

// substr_varchar_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var substr_varchar_int32 = framework.Function2{
	Name:       "substr",
	Return:     pgtypes.VarCharMax,
	Parameters: []pgtypes.DoltgresType{pgtypes.VarCharMax, pgtypes.Int32},
	Callable: func(ctx framework.Context, str any, start any) (any, error) {
		if str == nil || start == nil {
			return nil, nil
		}
		runes := []rune(str.(string))
		if int(start.(int32)) > len(runes) {
			return "", nil
		}
		return string(runes[start.(int32):]), nil
	},
}

// substr_varchar_int32_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var substr_varchar_int32_int32 = framework.Function3{
	Name:       "substr",
	Return:     pgtypes.VarCharMax,
	Parameters: []pgtypes.DoltgresType{pgtypes.VarCharMax, pgtypes.Int32, pgtypes.Int32},
	Callable: func(ctx framework.Context, str any, startInt any, countInt any) (any, error) {
		if str == nil || startInt == nil || countInt == nil {
			return nil, nil
		}
		start := startInt.(int32)
		count := countInt.(int32)
		runes := []rune(str.(string))
		if int(start) > len(runes) {
			return "", nil
		} else if int64(start)+int64(count) > int64(len(runes)) {
			return string(runes[start:]), nil
		}
		return string(runes[start : start+count]), nil
	},
}
