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

package framework

import (
	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// FindCommonType returns the common type that given types can convert to.
// https://www.postgresql.org/docs/15/typeconv-union-case.html
func FindCommonType(types []pgtypes.DoltgresTypeBaseID) (pgtypes.DoltgresTypeBaseID, bool) {
	var candidateType = pgtypes.DoltgresTypeBaseID_Unknown
	var fail = false
	for _, typBaseID := range types {
		if typBaseID == candidateType {
			continue
		} else if candidateType == pgtypes.DoltgresTypeBaseID_Unknown {
			candidateType = typBaseID
		} else {
			candidateType = pgtypes.DoltgresTypeBaseID_Unknown
			fail = true
		}
	}
	if !fail {
		if candidateType == pgtypes.DoltgresTypeBaseID_Unknown {
			return pgtypes.DoltgresTypeBaseID_Text, true
		}
		return candidateType, true
	}
	for _, typBaseID := range types {
		if candidateType == pgtypes.DoltgresTypeBaseID_Unknown {
			candidateType = typBaseID
		}
		if typBaseID != pgtypes.DoltgresTypeBaseID_Unknown && candidateType.GetTypeCategory() != typBaseID.GetTypeCategory() {
			return 0, false
		}
	}
	for _, typBaseID := range types {
		if typCastFunction := GetImplicitCast(candidateType, typBaseID); typCastFunction == nil {
			continue
		}
		candidateType = typBaseID
		if candidateType.GetRepresentativeType().IsPreferredType() {
			return candidateType, true
		}
	}
	return candidateType, true
}

// CastFromUnknownType if a type cast function that uses the unknown type output
// to get string value passed to the target type as input.
var CastFromUnknownType TypeCastFunction = func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	if val == nil {
		return nil, nil
	}
	str, err := pgtypes.Unknown.IoOutput(ctx, val)
	if err != nil {
		return nil, err
	}
	return targetType.IoInput(ctx, str)
}
