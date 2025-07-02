// Copyright 2025 Dolthub, Inc.
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
	"github.com/dolthub/doltgresql/core/extensions/pg_extension"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

var cConversionToDatumMap = map[id.Type]func(val any) (pg_extension.NullableDatum, error){
	pgtypes.Text.ID: textToDatum,
	pgtypes.Uuid.ID: uuidToDatum,
}
var cConversionFromDatumMap = map[id.Type]func(datum pg_extension.Datum) (any, error){
	pgtypes.Text.ID: textFromDatum,
	pgtypes.Uuid.ID: uuidFromDatum,
}

// textFromDatum converts from a Datum to a TEXT value.
func textFromDatum(datum pg_extension.Datum) (any, error) {
	convertedVal := pg_extension.FromDatumGoString(datum)
	pg_extension.FreeDatum(datum)
	return convertedVal, nil
}

// textToDatum converts from a TEXT value to a NullableDatum.
func textToDatum(val any) (pg_extension.NullableDatum, error) {
	if val == nil {
		return pg_extension.NullableDatum{
			Value:  0,
			IsNull: true,
		}, nil
	}
	return pg_extension.NullableDatum{
		Value:  pg_extension.ToDatumGoString(val.(string)),
		IsNull: false,
	}, nil
}

// uuidFromDatum converts from a Datum to a UUID value.
func uuidFromDatum(datum pg_extension.Datum) (any, error) {
	convertedVal := pg_extension.FromDatumGoBytes(datum, 16)
	pg_extension.FreeDatum(datum)
	return uuid.FromBytes(convertedVal)
}

// uuidToDatum converts from a UUID value to a NullableDatum.
func uuidToDatum(val any) (pg_extension.NullableDatum, error) {
	if val == nil {
		return pg_extension.NullableDatum{
			Value:  0,
			IsNull: true,
		}, nil
	}
	return pg_extension.NullableDatum{
		Value:  pg_extension.ToDatumGoBytes(val.(uuid.UUID).GetBytes()),
		IsNull: false,
	}, nil
}
