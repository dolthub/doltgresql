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
	"github.com/dolthub/go-mysql-server/sql"
	gmstypes "github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/vt/proto/query"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema/typeinfo"
)

const SchemasTablesSchemaNameCol = "schema_name"

func getSchemasSchema() schema.Schema {
	var schemasTableCols = schema.NewColCollection(
		mustNewColWithTypeInfo(doltdb.SchemasTablesTypeCol, schema.DoltSchemasTypeTag, typeinfo.CreateVarStringTypeFromSqlType(mustCreateStringType(query.Type_VARCHAR, 64, sql.Collation_utf8mb4_0900_ai_ci)), true, "", false, ""),
		mustNewColWithTypeInfo(doltdb.SchemasTablesNameCol, schema.DoltSchemasNameTag, typeinfo.CreateVarStringTypeFromSqlType(mustCreateStringType(query.Type_VARCHAR, 64, sql.Collation_utf8mb4_0900_ai_ci)), true, "", false, ""),
		mustNewColWithTypeInfo(doltdb.SchemasTablesSchemaNameCol, schema.DoltSchemasSchemaNameTag, typeinfo.CreateVarStringTypeFromSqlType(mustCreateStringType(query.Type_VARCHAR, 64, sql.Collation_utf8mb4_0900_ai_ci)), true, "", false, ""),
		mustNewColWithTypeInfo(doltdb.SchemasTablesFragmentCol, schema.DoltSchemasFragmentTag, typeinfo.CreateVarStringTypeFromSqlType(gmstypes.LongText), false, "", false, ""),
		mustNewColWithTypeInfo(doltdb.SchemasTablesExtraCol, schema.DoltSchemasExtraTag, typeinfo.JSONType, false, "", false, ""),
		mustNewColWithTypeInfo(doltdb.SchemasTablesSqlModeCol, schema.DoltSchemasSqlModeTag, typeinfo.CreateVarStringTypeFromSqlType(mustCreateStringType(query.Type_VARCHAR, 256, sql.Collation_utf8mb4_0900_ai_ci)), false, "", false, ""),
	)

	return schema.MustSchemaFromCols(schemasTableCols)
}

func mustCreateStringType(baseType query.Type, length int64, collation sql.CollationID) sql.StringType {
	ti, err := gmstypes.CreateString(baseType, length, collation)
	if err != nil {
		panic(err)
	}
	return ti
}

func mustNewColWithTypeInfo(name string, tag uint64, typeInfo typeinfo.TypeInfo, partOfPK bool, defaultVal string, autoIncrement bool, comment string, constraints ...schema.ColConstraint) schema.Column {
	col, err := schema.NewColumnWithTypeInfo(name, tag, typeInfo, partOfPK, defaultVal, autoIncrement, comment, constraints...)
	if err != nil {
		panic(err)
	}
	return col
}

// getStatusTableName returns the name of the status table.
func getSchemasTableName() string {
	return "schemas"
}
