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

package analyzer

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"

	"github.com/dolthub/doltgresql/server/extensions"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// validateVectorIndex enforces the operator class rules that require the indexed column's type, resolving the default
// operator class when the statement named none. The distance metric of the resolved class is written back to the node.
func validateVectorIndex(colMap map[string]*sql.Column, ai *plan.AlterIndex) error {
	col, ok := colMap[strings.ToLower(ai.Columns[0].Name)]
	if !ok {
		return sql.ErrKeyColumnDoesNotExist.New(ai.Columns[0].Name)
	}
	colType, ok := col.Type.(*pgtypes.DoltgresType)
	if !ok {
		return nil
	}
	typeName := colType.Name()
	var opclass extdef.OperatorClass
	if ai.VectorOpClass == "" {
		if opclass, ok = extensions.GetDefaultOperatorClass(ai.VectorAccessMethod, typeName); !ok {
			return errors.Errorf(`data type %s has no default operator class for access method "%s"`, typeName, ai.VectorAccessMethod)
		}
	} else {
		if opclass, ok = extensions.GetOperatorClass(ai.VectorOpClass); !ok {
			return errors.Errorf(`operator class "%s" does not exist for access method "%s"`, ai.VectorOpClass, ai.VectorAccessMethod)
		}
		if opclass.Type != typeName {
			return errors.Errorf(`operator class "%s" does not accept data type %s`, ai.VectorOpClass, typeName)
		}
	}
	dims := colType.GetAttTypMod()
	if dims < 1 {
		return errors.Errorf("column does not have dimensions")
	}
	if opclass.MaxDimensions > 0 && dims > opclass.MaxDimensions {
		return errors.Errorf("column cannot have more than %d dimensions for %s index", opclass.MaxDimensions, ai.VectorAccessMethod)
	}
	ai.VectorProperties.DistanceType = opclass.DistanceType
	return nil
}
