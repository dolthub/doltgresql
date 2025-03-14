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

package dtables

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/val"
	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// getDoltIgnoreSchema returns the schema for the dolt_ignore table.
func getDoltIgnoreSchema() sql.Schema {
	return []*sql.Column{
		{Name: "pattern", Type: pgtypes.Text, Source: doltdb.IgnoreTableName, PrimaryKey: true},
		{Name: "ignored", Type: pgtypes.Bool, Source: doltdb.IgnoreTableName, PrimaryKey: false, Nullable: false},
	}
}

// convertTupleToIgnoreBoolean reads a boolean from a tuple and returns it.
func convertTupleToIgnoreBoolean(ctx context.Context, valueDesc val.TupleDesc, valueTuple val.Tuple) (bool, error) {
	extendedTuple := val.NewTupleDescriptorWithArgs(
		val.TupleDescriptorArgs{Comparator: valueDesc.Comparator(), Handlers: valueDesc.Handlers},
		val.Type{Enc: val.ExtendedEnc, Nullable: false},
	)
	if !valueDesc.Equals(extendedTuple) {
		return false, errors.Errorf("dolt_ignore had unexpected value type, this should never happen")
	}
	extended, ok := valueDesc.GetExtended(0, valueTuple)
	if !ok {
		return false, errors.Errorf("could not read boolean")
	}
	val, err := valueDesc.Handlers[0].DeserializeValue(ctx, extended)
	if err != nil {
		return false, err
	}
	ignore, ok := val.(bool)
	if !ok {
		return false, errors.Errorf("could not read boolean")
	}
	return ignore, nil
}

// getIgnoreTablePatternKey reads the pattern key from a tuple and returns it.
func getIgnoreTablePatternKey(ctx context.Context, keyDesc val.TupleDesc, keyTuple val.Tuple) (string, error) {
	extendedTuple := val.NewTupleDescriptorWithArgs(
		val.TupleDescriptorArgs{Comparator: keyDesc.Comparator(), Handlers: keyDesc.Handlers},
		val.Type{Enc: val.ExtendedAddrEnc, Nullable: false},
	)
	if !keyDesc.Equals(extendedTuple) {
		return "", fmt.Errorf("dolt_ignore had unexpected key type, this should never happen")
	}

	keyAddr, ok := keyDesc.GetExtendedAddr(0, keyTuple)
	if !ok {
		return "", fmt.Errorf("could not read pattern")
	}

	key, err := keyDesc.Handlers[0].DeserializeValue(ctx, keyAddr[:])
	if err != nil {
		return "", err
	}

	return key.(string), nil
}
