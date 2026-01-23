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
	"github.com/cockroachdb/errors"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// FindCommonType returns the common type that given types can convert to. Returns false if no implicit casts are needed
// to resolve the given types as the returned common type.
// https://www.postgresql.org/docs/15/typeconv-union-case.html
func FindCommonType(types []*pgtypes.DoltgresType) (_ *pgtypes.DoltgresType, requiresCasts bool, err error) {
	candidateType := pgtypes.Unknown
	differentTypes := false
	for _, typ := range types {
		if typ.ID == candidateType.ID {
			continue
		} else if candidateType.ID == pgtypes.Unknown.ID {
			candidateType = typ
		} else {
			candidateType = pgtypes.Unknown
			differentTypes = true
		}
	}
	if !differentTypes {
		if candidateType.ID == pgtypes.Unknown.ID {
			// We require implicit casts from `unknown` to `text`
			return pgtypes.Text, true, nil
		}
		return candidateType, false, nil
	}
	// We have different types if we've made it this far, so we're guaranteed to require implicit casts
	requiresCasts = true
	for _, typ := range types {
		if candidateType.ID == pgtypes.Unknown.ID {
			candidateType = typ
		}
		if typ.ID != pgtypes.Unknown.ID && candidateType.TypCategory != typ.TypCategory {
			return nil, false, errors.Errorf("types %s and %s cannot be matched", candidateType.String(), typ.String())
		}
	}
	// Attempt to find the most general type (or the preferred type in the type category)
	for _, typ := range types {
		if typ.ID == pgtypes.Unknown.ID || typ.ID == candidateType.ID {
			continue
		} else if GetImplicitCast(typ, candidateType) != nil {
			// typ can convert to the candidate type, so the candidate type is at least as general
			continue
		} else if GetImplicitCast(candidateType, typ) != nil {
			// the candidate type can convert to typ, but not vice versa, so typ is likely more general
			candidateType = typ
			if candidateType.IsPreferred {
				// We stop considering more types once we've found a preferred type
				break
			}
		}
	}
	// Verify that all types have an implicit conversion to the candidate type
	for _, typ := range types {
		if typ.ID == pgtypes.Unknown.ID || typ.ID == candidateType.ID {
			continue
		} else if GetImplicitCast(typ, candidateType) == nil {
			return nil, false, errors.Errorf("cannot find implicit cast function from %s to %s", candidateType.String(), typ.String())
		}
	}
	return candidateType, requiresCasts, nil
}
