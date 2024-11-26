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
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dprocedures"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dtables"
)

// Init handles initialization of all Postgres-specific and Doltgres-specific Dolt system tables.
func Init() {
	// Table names
	doltdb.GetBranchesTableName = getBranchesTableName
	doltdb.GetDocTableName = getDocTableName
	doltdb.GetColumnDiffTableName = getColumnDiffTableName
	doltdb.GetCommitAncestorsTableName = getCommitAncestorsTableName
	doltdb.GetCommitsTableName = getCommitsTableName
	doltdb.GetDiffTableName = getDiffTableName
	doltdb.GetLogTableName = getLogTableName
	doltdb.GetMergeStatusTableName = getMergeStatusTableName
	doltdb.GetRebaseTableName = getRebaseTableName
	doltdb.GetRemoteBranchesTableName = getRemoteBranchesTableName
	doltdb.GetRemotesTableName = getRemotesTableName
	doltdb.GetSchemaConflictsTableName = getSchemaConflictsTableName
	doltdb.GetStatusTableName = getStatusTableName
	doltdb.GetTableOfTablesInConflictName = getTableOfTablesInConflictName
	doltdb.GetTableOfTablesWithViolationsName = getTableOfTablesWithViolationsName
	doltdb.GetTagsTableName = getTagsTableName

	// Schemas
	dtables.GetDocsSchema = getDocsSchema
	dtables.GetDoltIgnoreSchema = getDoltIgnoreSchema
	dprocedures.GetDoltRebaseSystemTableSchema = getRebaseSchema

	// Conversions
	doltdb.ConvertTupleToIgnoreBoolean = convertTupleToIgnoreBoolean
	sqle.ConvertRebasePlanStepToRow = convertRebasePlanStepToRow
	sqle.ConvertRowToRebasePlanStep = convertRowToRebasePlanStep
}

// getBranchesTableName returns the name of the branches table.
func getBranchesTableName() string {
	return "branches"
}

// getColumnDiffTableName returns the name of the column diff table.
func getColumnDiffTableName() string {
	return "column_diff"
}

// getCommitAncestorsTableName returns the name of the commit ancestors table.
func getCommitAncestorsTableName() string {
	return "commit_ancestors"
}

// getCommitsTableName returns the name of the commits table.
func getCommitsTableName() string {
	return "commits"
}

// getDiffTableName returns the name of the diff table.
func getDiffTableName() string {
	return "diff"
}

// getLogTableName returns the name of the branches table.
func getLogTableName() string {
	return "log"
}

// getMergeStatusTableName returns the name of the merge status table.
func getMergeStatusTableName() string {
	return "merge_status"
}

// getRemoteBranchesTableName returns the name of the remote branches table.
func getRemoteBranchesTableName() string {
	return "remote_branches"
}

// getRemotesTableName returns the name of the remotes table.
func getRemotesTableName() string {
	return "remotes"
}

// getSchemaConflictsTableName returns the name of the schema conflicts table.
func getSchemaConflictsTableName() string {
	return "schema_conflicts"
}

// getStatusTableName returns the name of the status table.
func getStatusTableName() string {
	return "status"
}

// getTableOfTablesInConflictName returns the name of the conflicts table.
func getTableOfTablesInConflictName() string {
	return "conflicts"
}

// getTableOfTablesWithViolationsName returns the name of the constraint violations table.
func getTableOfTablesWithViolationsName() string {
	return "constraint_violations"
}

// getTagsTableName returns the name of the tags table.
func getTagsTableName() string {
	return "tags"
}
