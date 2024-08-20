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
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// FindCommonType returns the common type that given types can convert to.
// https://www.postgresql.org/docs/15/typeconv-union-case.html
func FindCommonType(types []pgtypes.DoltgresTypeBaseID) (pgtypes.DoltgresTypeBaseID, error) {
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
			return pgtypes.DoltgresTypeBaseID_Text, nil
		}
		return candidateType, nil
	}
	for _, typBaseID := range types {
		if candidateType == pgtypes.DoltgresTypeBaseID_Unknown {
			candidateType = typBaseID
		}
		if typBaseID != pgtypes.DoltgresTypeBaseID_Unknown && candidateType.GetTypeCategory() != typBaseID.GetTypeCategory() {
			return 0, fmt.Errorf("types %s and %s cannot be matched", candidateType.GetRepresentativeType().String(), typBaseID.GetRepresentativeType().String())
		}
	}

	var preferredTypeFound = false
	for _, typBaseID := range types {
		if typBaseID == pgtypes.DoltgresTypeBaseID_Unknown {
			continue
		} else if GetImplicitCast(candidateType, typBaseID) == nil {
			continue
		} else if preferredTypeFound && GetImplicitCast(typBaseID, candidateType) == nil {
			return 0, fmt.Errorf("cannot find implicit cast function from %s to %s", typBaseID.String(), candidateType.String())
		} else if !preferredTypeFound {
			candidateType = typBaseID
			if candidateType.GetRepresentativeType().IsPreferredType() {
				preferredTypeFound = true
			}
		}
	}
	return candidateType, nil
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
