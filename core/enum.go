// Copyright 2026 Dolthub, Inc.
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
	"maps"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

type enumTypeKey struct {
	database string
	typ      id.Type
}
type enumValueKey struct {
	enumTypeKey
	value string
}

// These maps are immutable journal snapshots, including across savepoints.
type enumTransactionChanges struct {
	added map[enumValueKey]bool
}

// RecordEnumValueAdded tracks values that must not be used until commit.
func RecordEnumValueAdded(ctx *sql.Context, typ id.Type, value string) error {
	cv, err := getContextValues(ctx)
	if err != nil {
		return err
	}
	next := *cv.enumChanges.Current()
	key := enumTypeKey{ctx.GetCurrentDatabase(), typ}
	next.added = maps.Clone(next.added)
	if next.added == nil {
		next.added = make(map[enumValueKey]bool)
	}
	next.added[enumValueKey{key, value}] = true
	cv.enumChanges.SetLocal(next)
	return nil
}

// CheckEnumValueUsage prevents uncommitted new labels from entering stored rows.
func CheckEnumValueUsage(ctx *sql.Context, typ id.Type, value string) error {
	cv, err := getContextValues(ctx)
	if err != nil {
		return err
	}
	if cv.enumChanges.Current().added[enumValueKey{enumTypeKey{ctx.GetCurrentDatabase(), typ}, value}] {
		return pgerror.Newf(pgcode.UnsafeNewEnumValueUsage, "unsafe use of new value %q of enum type %s", value, typ.TypeName())
	}
	return nil
}
