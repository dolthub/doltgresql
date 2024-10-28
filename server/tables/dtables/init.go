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
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dtables"
)

// Init handles initialization of all Postgres-specific and Doltgres-specific Dolt system tables.
func Init() {
	dtables.GetDocsSchema = getDocsSchema
	doltdb.GetDocTableName = getDocTableName
	doltdb.GetBranchesTableName = getBranchesTableName
	doltdb.GetLogTableName = getLogTableName
	doltdb.GetStatusTableName = getStatusTableName
	doltdb.GetTagsTableName = getTagsTableName
}

// getBranchesTableName returns the name of the branches table.
func getBranchesTableName() string {
	return "branches"
}

// getLogTableName returns the name of the branches table.
func getLogTableName() string {
	return "log"
}

// getStatusTableName returns the name of the status table.
func getStatusTableName() string {
	return "status"
}

// getTagsTableName returns the name of the tags table.
func getTagsTableName() string {
	return "tags"
}
