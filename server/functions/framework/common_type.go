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

// FindCommonType returns the common type that given types can convert to.
// https://www.postgresql.org/docs/15/typeconv-union-case.html
func FindCommonType(types []*pgtypes.DoltgresType) (*pgtypes.DoltgresType, error) {
	var candidateType = pgtypes.Unknown
	var fail = false
	for _, typ := range types {
		if typ.ID == candidateType.ID {
			continue
		} else if candidateType.ID == pgtypes.Unknown.ID {
			candidateType = typ
		} else {
			candidateType = pgtypes.Unknown
			fail = true
		}
	}
	if !fail {
		if candidateType.ID == pgtypes.Unknown.ID {
			return pgtypes.Text, nil
		}
		return candidateType, nil
	}
	for _, typ := range types {
		if candidateType.ID == pgtypes.Unknown.ID {
			candidateType = typ
		}
		if typ.ID != pgtypes.Unknown.ID && candidateType.TypCategory != typ.TypCategory {
			return nil, errors.Errorf("types %s and %s cannot be matched", candidateType.String(), typ.String())
		}
	}

	var preferredTypeFound = false
	for _, typ := range types {
		if typ.ID == pgtypes.Unknown.ID {
			continue
		} else if GetImplicitCast(typ, candidateType) != nil {
			continue
		} else if GetImplicitCast(candidateType, typ) == nil {
			return nil, errors.Errorf("cannot find implicit cast function from %s to %s", candidateType.String(), typ.String())
		} else if !preferredTypeFound {
			if candidateType.IsPreferred {
				candidateType = typ
				preferredTypeFound = true
			}
		} else {
			return nil, errors.Errorf("found another preferred candidate type")
		}
	}
	return candidateType, nil
}
