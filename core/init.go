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

package core

import (
	"fmt"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/types"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/core/conflicts"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/typecollection"
	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Init initializes this package.
func Init() {
	doltdb.EmptyRootValue = emptyRootValue
	doltdb.NewRootValue = newRootValue
	types.DoltgresRootValueHumanReadableStringAtIndentationLevel = rootValueHumanReadableStringAtIndentationLevel
	types.DoltgresRootValueWalkAddrs = rootValueWalkAddrs
	conflicts.ClearContextValues = ClearContextValues
	plpgsql.GetTypesCollectionFromContext = GetTypesCollectionFromContext
	id.RegisterListener(sequenceIDListener{}, id.Section_Table)
	typecollection.GetSqlTableFromContext = GetSqlTableFromContext
	typecollection.GetSchemaName = GetSchemaName
	pgtypes.GetTypesCollectionFromContext = func(ctx *sql.Context, database string) (pgtypes.TypeCollection, error) {
		return GetTypesCollectionFromContext(ctx, database)
	}
	pgtypes.GetCastFunc = func(ctx *sql.Context, convTyp byte) (func(*pgtypes.DoltgresType, *pgtypes.DoltgresType) (pgtypes.Cast, bool, error), error) {
		castsColl, err := GetCastsCollectionFromContext(ctx, "")
		return func(sourceType *pgtypes.DoltgresType, targetType *pgtypes.DoltgresType) (pgtypes.Cast, bool, error) {
			var cast casts.Cast
			var cErr error
			switch convTyp {
			case 'e':
				cast, cErr = castsColl.GetExplicitCast(ctx, sourceType, targetType)
			case 'a':
				cast, cErr = castsColl.GetAssignmentCast(ctx, sourceType, targetType)
			case 'i':
				cast, cErr = castsColl.GetImplicitCast(ctx, sourceType, targetType)
			default:
				return nil, false, fmt.Errorf("unknown conversion type: %v", convTyp)
			}
			if cErr != nil {
				return nil, false, cErr
			}
			return cast, cast.ID.IsValid(), nil
		}, err
	}
	sql.GetCommonExtendedType = func(ctx *sql.Context, sourceType, targetType sql.ExtendedType) sql.ExtendedType {
		if sourceType.Equals(targetType) {
			return sourceType
		}
		source, ok := sourceType.(*pgtypes.DoltgresType)
		if !ok {
			return source
		}
		target, ok := targetType.(*pgtypes.DoltgresType)
		if !ok {
			return source
		}

		if source.ID == target.ID {
			return source
		}
		if source.IsPreferred {
			return source
		} else if target.IsPreferred {
			return target
		}
		castsColl, err := GetCastsCollectionFromContext(ctx, "")
		if err != nil {
			return source
		}
		cast, err := castsColl.GetImplicitCast(ctx, source, target)
		if err == nil && cast.ID.IsValid() {
			return target
		}
		cast, err = castsColl.GetImplicitCast(ctx, target, source)
		if err == nil && cast.ID.IsValid() {
			return source
		}
		return pgtypes.Text
	}
}
