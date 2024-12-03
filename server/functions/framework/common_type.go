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

	"github.com/lib/pq/oid"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// FindCommonType returns the common type that given types can convert to.
// https://www.postgresql.org/docs/15/typeconv-union-case.html
func FindCommonType(typOids []uint32) (uint32, error) {
	var candidateTypeOid = pgtypes.Unknown.OID
	var fail = false
	for _, typOid := range typOids {
		if typOid == candidateTypeOid {
			continue
		} else if candidateTypeOid == uint32(oid.T_unknown) {
			candidateTypeOid = typOid
		} else {
			candidateTypeOid = pgtypes.Unknown.OID
			fail = true
		}
	}
	if !fail {
		if candidateTypeOid == uint32(oid.T_unknown) {
			return pgtypes.Text.OID, nil
		}
		return candidateTypeOid, nil
	}
	for _, typOid := range typOids {
		if candidateTypeOid == uint32(oid.T_unknown) {
			candidateTypeOid = typOid
		}
		candidateType := pgtypes.OidToBuiltInDoltgresType[candidateTypeOid]
		typ := pgtypes.OidToBuiltInDoltgresType[typOid]
		if typOid != uint32(oid.T_unknown) && candidateType.TypCategory != typ.TypCategory {
			return 0, fmt.Errorf("types %s and %s cannot be matched", candidateType.String(), typ.String())
		}
	}

	var preferredTypeFound = false
	for _, typOid := range typOids {
		if typOid == uint32(oid.T_unknown) {
			continue
		} else if GetImplicitCast(typOid, candidateTypeOid) != nil {
			continue
		} else if GetImplicitCast(candidateTypeOid, typOid) == nil {
			candidateType := pgtypes.OidToBuiltInDoltgresType[candidateTypeOid]
			typ := pgtypes.OidToBuiltInDoltgresType[typOid]
			return 0, fmt.Errorf("cannot find implicit cast function from %s to %s", candidateType.String(), typ.String())
		} else if !preferredTypeFound {
			candidateType := pgtypes.OidToBuiltInDoltgresType[candidateTypeOid]
			if candidateType.IsPreferred {
				candidateTypeOid = typOid
				preferredTypeFound = true
			}
		} else {
			return 0, fmt.Errorf("found another preferred candidate type")
		}
	}
	return candidateTypeOid, nil
}
