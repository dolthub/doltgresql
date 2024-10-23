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

	pgtypes "github.com/dolthub/doltgresql/server/types"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
)

// getDocsSchema returns the schema for the docs table.
func getDocsSchema() sql.Schema {
	return []*sql.Column{
		{Name: doltdb.DocPkColumnName, Type: pgtypes.Text, Source: getDocTableName(), PrimaryKey: true, Nullable: false},
		{Name: doltdb.DocTextColumnName, Type: pgtypes.Text, Source: getDocTableName(), PrimaryKey: false},
	}
}

// getDocTableName returns the name of the docs table.
func getDocTableName() string {
	return "docs"
}
