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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initDoltRecordTrim registers the functions to the catalog.
func initDoltRecordTrim() {
	framework.RegisterFunction(dolt_recordTrim)
}

// dolt_recordTrim is used to remove a specific column within a composite type. This will generally lead to an invalid
// value for the composite type, however this is used within the DROP COLUMN table hook to fix data for all columns that
// reference the type, as that is the only time when the data is invalid. This is why this is a "dolt_" function as
// well, as it's not intended for general use.
var dolt_recordTrim = framework.Function2{
	Name:       "dolt_recordtrim",
	Return:     pgtypes.AnyElement,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyElement, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, types [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if !types[0].IsCompositeType() {
			return val1, nil
		}
		trimVal := val2.(int32)
		recordVals := val1.([]pgtypes.RecordValue)
		if trimVal < 0 || int(trimVal) >= len(recordVals) {
			return nil, errors.New("record trim index is out of bounds")
		}
		newRecordVals := make([]pgtypes.RecordValue, len(recordVals)-1)
		copy(newRecordVals, recordVals[:trimVal])
		copy(newRecordVals[trimVal:], recordVals[trimVal+1:])
		return newRecordVals, nil
	},
}
