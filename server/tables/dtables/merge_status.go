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
)

// getDoltMergeStatusSchema returns the schema for the merge_status table.
func getDoltMergeStatusSchema(dbName, tableName string) sql.Schema {
	return []*sql.Column{
		{Name: "is_merging", Type: pgtypes.Bool, Source: tableName, PrimaryKey: false, Nullable: false, DatabaseSource: dbName},
		{Name: "source", Type: pgtypes.Text, Source: tableName, PrimaryKey: false, Nullable: true, DatabaseSource: dbName},
		{Name: "source_commit", Type: pgtypes.Text, Source: tableName, PrimaryKey: false, Nullable: true, DatabaseSource: dbName},
		{Name: "target", Type: pgtypes.Text, Source: tableName, PrimaryKey: false, Nullable: true, DatabaseSource: dbName},
		{Name: "unmerged_tables", Type: pgtypes.Text, Source: tableName, PrimaryKey: false, Nullable: true, DatabaseSource: dbName},
	}
}

// getMergeStatusTableName returns the name of the merge status table.
func getMergeStatusTableName() string {
	return "merge_status"
}
