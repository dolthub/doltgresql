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
	"bytes"
	"fmt"
	"time"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/merge"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/dolt/go/store/types"
	"github.com/dolthub/dolt/go/store/val"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/typecollection"
	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// rowLockWaitTimeout returns how long a statement waits for a row lock before giving up, which the lock_timeout
// setting gives in milliseconds. Zero waits indefinitely.
func rowLockWaitTimeout(ctx *sql.Context) time.Duration {
	val, err := ctx.GetSessionVariable(ctx, "lock_timeout")
	if err != nil {
		return 0
	}

	millis, ok := val.(int64)
	if !ok || millis <= 0 {
		return 0
	}

	return time.Duration(millis) * time.Millisecond
}

// Init initializes this package.
func Init() {
	doltdb.EmptyRootValue = emptyRootValue
	doltdb.NewRootValue = newRootValue
	dsess.TransactionRowMergePolicy = transactionRowMergePolicy
	dsess.RowLockWaitTimeout = rowLockWaitTimeout
	types.DoltgresRootValueHumanReadableStringAtIndentationLevel = rootValueHumanReadableStringAtIndentationLevel
	types.DoltgresRootValueWalkAddrs = rootValueWalkAddrs
	plpgsql.GetTypesCollectionFromContext = GetTypesCollectionFromContext
	id.RegisterListener(sequenceIDListener{}, id.Section_Table)
	typecollection.GetSqlTableFromContext = GetSqlTableFromContext
	typecollection.LookupSchemaName = LookupSchemaName
	pgtypes.GetTypesCollectionFromContext = func(ctx *sql.Context, database string) (pgtypes.TypeCollection, error) {
		return GetTypesCollectionFromContext(ctx, database)
	}
	pgtypes.GetCastFunc = func(ctx *sql.Context, convTyp byte) (func(*pgtypes.DoltgresType, *pgtypes.DoltgresType) (pgtypes.Cast, bool, error), error) {
		castsColl, err := GetCastsCollectionFromContext(ctx, "")
		return func(sourceType *pgtypes.DoltgresType, targetType *pgtypes.DoltgresType) (pgtypes.Cast, bool, error) {
			var cast casts.Cast
			var castType casts.CastType
			var cErr error
			switch convTyp {
			case 'e':
				castType = casts.CastType_Explicit
				cast, cErr = castsColl.GetExplicitCast(ctx, sourceType, targetType)
			case 'a':
				castType = casts.CastType_Assignment
				cast, cErr = castsColl.GetAssignmentCast(ctx, sourceType, targetType)
			case 'i':
				castType = casts.CastType_Implicit
				cast, cErr = castsColl.GetImplicitCast(ctx, sourceType, targetType)
			default:
				return nil, false, fmt.Errorf("unknown conversion type: %v", convTyp)
			}
			if cErr != nil {
				return nil, false, cErr
			}
			// As in PostgreSQL, there are no casts to the unknown type, whose values are the text of a literal. Values
			// still convert to it here through their output function, so that callers can check that a value converted
			// from an unknown literal converts back to the same text.
			if !cast.ID.IsValid() && targetType.ID == pgtypes.Unknown.ID && sourceType.ID != pgtypes.Unknown.ID {
				cast = casts.Cast{
					ID:       id.NewCast(sourceType.ID, targetType.ID),
					CastType: castType,
					Function: id.NullFunction,
					UseInOut: true,
				}
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

// transactionRowMergePolicy decides each row that both a committing transaction and a concurrently committed
// transaction changed. If both changed the same column of the row, the later commit fails with a serialization failure,
// even when both wrote the same value: each may have computed its value from the row as it was before the other's
// change, as `SET n = n + 1` does, so accepting both would silently lose one of the changes. Inserting or deleting a row
// changes all of its columns. Changes to different columns of the same row merge as usual.
func transactionRowMergePolicy(_ *sql.Context, in merge.RowMergeInput) (val.Tuple, tree.RowMergeStatus, error) {
	if in.Base == nil || in.Left == nil || in.Right == nil {
		return nil, tree.RowMergeConflict, nil
	}

	for i := 0; i < in.ValueDesc.Count(); i++ {
		base := in.Base.GetField(i)
		if !bytes.Equal(in.Left.GetField(i), base) && !bytes.Equal(in.Right.GetField(i), base) {
			return nil, tree.RowMergeConflict, nil
		}
	}
	return nil, tree.RowMergeDefer, nil
}
