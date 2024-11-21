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
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/val"
)

func getDoltIgnoreSchema() sql.Schema {
	return []*sql.Column{
		{Name: "pattern", Type: pgtypes.Text, Source: doltdb.IgnoreTableName, PrimaryKey: true},
		{Name: "ignored", Type: pgtypes.Bool, Source: doltdb.IgnoreTableName, PrimaryKey: false, Nullable: false},
	}
}

func convertTupleToIgnoreBoolean(valueDesc val.TupleDesc, valueTuple val.Tuple) (bool, error) {
	extendedTuple := val.NewTupleDescriptorWithArgs(
		val.TupleDescriptorArgs{Comparator: valueDesc.Comparator(), Handlers: valueDesc.Handlers},
		val.Type{Enc: val.ExtendedEnc, Nullable: false},
	)
	if !valueDesc.Equals(extendedTuple) {
		return false, fmt.Errorf("dolt_ignore had unexpected value type, this should never happen")
	}
	extended, ok := valueDesc.GetExtended(0, valueTuple)
	if !ok {
		return false, fmt.Errorf("could not read boolean")
	}
	val, err := valueDesc.Handlers[0].DeserializeValue(extended)
	if err != nil {
		return false, err
	}
	ignore, ok := val.(bool)
	if !ok {
		return false, fmt.Errorf("could not read boolean")
	}
	return ignore, nil
}
