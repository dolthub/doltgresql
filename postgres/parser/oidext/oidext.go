// Copyright 2023 Dolthub, Inc.
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

// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package oidext contains oids that are not in `github.com/lib/pq/oid`
// as they are not shipped by default with postgres.
// As CRDB does not support extensions, we'll need to automatically assign
// a few OIDs of our own.
package oidext

import "github.com/lib/pq/oid"

// CockroachPredefinedOIDMax defines the maximum OID allowed for use by
// non user defined types. OIDs for user defined types will start at
// CockroachPrefixedOIDMax and increase as new types are created.
// User defined type descriptors have a cluster-wide unique stable ID.
// CockroachPredefinedOIDMax defines the mapping from this stable ID to
// a type OID. In particular, stable ID + CockroachPredefinedOIDMax = type OID.
// types.StableTypeIDToOID and types.UserDefinedTypeOIDToID should be used when
// converting between stable ID's and type OIDs.
const CockroachPredefinedOIDMax = 100000

// OIDs in this block are extensions of postgres, thus having no official OID.
const (
	T_geometry   = oid.Oid(90000)
	T__geometry  = oid.Oid(90001)
	T_geography  = oid.Oid(90002)
	T__geography = oid.Oid(90003)
	T_box2d      = oid.Oid(90004)
	T__box2d     = oid.Oid(90005)
)

// ExtensionTypeName returns a mapping from extension oids
// to their type name.
var ExtensionTypeName = map[oid.Oid]string{
	T_geometry:   "GEOMETRY",
	T__geometry:  "_GEOMETRY",
	T_geography:  "GEOGRAPHY",
	T__geography: "_GEOGRAPHY",
	T_box2d:      "BOX2D",
	T__box2d:     "_BOX2D",
}

// TypeName checks the name for a given type by first looking up oid.TypeName
// before falling back to looking at the oid extension ExtensionTypeName.
func TypeName(o oid.Oid) (string, bool) {
	name, ok := oid.TypeName[o]
	if ok {
		return name, ok
	}
	name, ok = ExtensionTypeName[o]
	return name, ok
}
