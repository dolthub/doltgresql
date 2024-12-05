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
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// getDoltConstraintViolationsSchema returns the base schema for the dolt_constraint_violations_* table.
func getDoltConstraintViolationsBaseSqlSchema() sql.Schema {
	return []*sql.Column{
		{Name: "from_root_ish", Type: pgtypes.Text, PrimaryKey: false, Nullable: true},
		{Name: "violation_type", Type: pgtypes.MustCreateNewVarCharType(16), PrimaryKey: true},
	}
}

// mapCVType maps a prolly.ArtifactType to a string.
func mapCVType(artifactType prolly.ArtifactType) any {
	return mapCVTypeString(artifactType)
}

func mapCVTypeString(artifactType prolly.ArtifactType) (outType string) {
	switch artifactType {
	case prolly.ArtifactTypeForeignKeyViol:
		outType = "foreign key"
	case prolly.ArtifactTypeUniqueKeyViol:
		outType = "unique index"
	case prolly.ArtifactTypeChkConsViol:
		outType = "check constraint"
	case prolly.ArtifactTypeNullViol:
		outType = "not null"
	default:
		panic("unhandled cv type")
	}
	return
}

// unmapCVType unmaps a string to a prolly.ArtifactType.
func unmapCVType(in any) (out prolly.ArtifactType) {
	if cv, ok := in.(string); ok {
		return unmapCVTypeString(cv)
	}
	panic("invalid type")
}

func unmapCVTypeString(in string) (out prolly.ArtifactType) {
	switch in {
	case "foreign key":
		out = prolly.ArtifactTypeForeignKeyViol
	case "unique index":
		out = prolly.ArtifactTypeUniqueKeyViol
	case "check constraint":
		out = prolly.ArtifactTypeChkConsViol
	case "not null":
		out = prolly.ArtifactTypeNullViol
	default:
		panic("unhandled cv type")
	}
	return
}
