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
func FindCommonType(types []pgtypes.DoltgresTypeBaseID) (pgtypes.DoltgresTypeBaseID, bool) {
	var candidateType = pgtypes.DoltgresTypeBaseID_Any
	for _, typBaseID := range types {
		if candidateType == pgtypes.DoltgresTypeBaseID_Any {
			candidateType = typBaseID
		} else if typBaseID == pgtypes.DoltgresTypeBaseID_Unknown {
			continue
		} else if candidateType != typBaseID {
			if candidateType == pgtypes.DoltgresTypeBaseID_Unknown {
				candidateType = typBaseID
				continue
			} else if candidateType.GetTypeCategory() != typBaseID.GetTypeCategory() {
				return 0, false
			} else if typCastFunction := GetImplicitCast(candidateType, typBaseID); typCastFunction == nil {
				continue
			}
			candidateType = typBaseID
			if candidateType.GetRepresentativeType().IsPreferredType() {
				return candidateType, true
			}
		}
	}
	if candidateType == pgtypes.DoltgresTypeBaseID_Unknown {
		return pgtypes.DoltgresTypeBaseID_Text, true
	}
	return candidateType, true
}

// ConvertValToCommonType returns an input converted to the final candidate/common type.
// Fail if there is not an implicit conversion from a given input type to the candidate type.
func ConvertValToCommonType(ctx *sql.Context, val any, valTyp, resultTyp pgtypes.DoltgresType) (any, error) {
	if val == nil {
		return nil, nil
	}

	if valTyp.BaseID() == pgtypes.DoltgresTypeBaseID_Unknown {
		valTyp = pgtypes.Text
	}

	// We always cast the element, as there may be parameter restrictions in place
	castFunc := GetImplicitCast(valTyp.BaseID(), resultTyp.BaseID())
	if castFunc == nil {
		return nil, fmt.Errorf("cannot find implicit cast function from %s to %s", valTyp.String(), resultTyp.String())
	}

	return castFunc(ctx, val, resultTyp)
}
