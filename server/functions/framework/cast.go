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
	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/core/id"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TODO: no need to use these functions, should instead add everything directly to built-in.
//  For now, this just makes the transition easier since it's less to rewrite

// TypeCast is used to cast from one type to another.
type TypeCast struct {
	FromType *pgtypes.DoltgresType
	ToType   *pgtypes.DoltgresType
	Function pgtypes.TypeCastFunction
}

// MustAddExplicitTypeCast registers the given explicit type cast. Panics if an error occurs.
func MustAddExplicitTypeCast(builtInCasts map[id.Cast]casts.Cast, cast TypeCast) {
	castID := id.NewCast(cast.FromType.ID, cast.ToType.ID)
	if _, ok := builtInCasts[castID]; ok {
		panic("duplicate built-in cast")
	}
	builtInCasts[castID] = casts.Cast{
		ID:       castID,
		CastType: casts.CastType_Explicit,
		Function: id.NullFunction,
		BuiltIn:  cast.Function,
		UseInOut: false,
	}
}

// MustAddAssignmentTypeCast registers the given assignment type cast. Panics if an error occurs.
func MustAddAssignmentTypeCast(builtInCasts map[id.Cast]casts.Cast, cast TypeCast) {
	castID := id.NewCast(cast.FromType.ID, cast.ToType.ID)
	if _, ok := builtInCasts[castID]; ok {
		panic("duplicate built-in cast")
	}
	builtInCasts[castID] = casts.Cast{
		ID:       castID,
		CastType: casts.CastType_Assignment,
		Function: id.NullFunction,
		BuiltIn:  cast.Function,
		UseInOut: false,
	}
}

// MustAddImplicitTypeCast registers the given implicit type cast. Panics if an error occurs.
func MustAddImplicitTypeCast(builtInCasts map[id.Cast]casts.Cast, cast TypeCast) {
	castID := id.NewCast(cast.FromType.ID, cast.ToType.ID)
	if _, ok := builtInCasts[castID]; ok {
		panic("duplicate built-in cast")
	}
	builtInCasts[castID] = casts.Cast{
		ID:       castID,
		CastType: casts.CastType_Implicit,
		Function: id.NullFunction,
		BuiltIn:  cast.Function,
		UseInOut: false,
	}
}
