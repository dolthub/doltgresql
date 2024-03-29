// Copyright 2023 Dolthub, Inc.
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

package ast

import (
	"fmt"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// Convert converts a Postgres AST into a Vitess AST.
func Convert(postgresStmt parser.Statement) (vitess.Statement, error) {
	switch stmt := postgresStmt.AST.(type) {
	case *tree.AlterDatabase:
		return nodeAlterDatabase(stmt)
	case *tree.AlterFunction:
		return nodeAlterFunction(stmt)
	case *tree.AlterIndex:
		return nodeAlterIndex(stmt)
	case *tree.AlterProcedure:
		return nodeAlterProcedure(stmt)
	case *tree.AlterRole:
		return nodeAlterRole(stmt)
	case *tree.AlterSchema:
		return nodeAlterSchema(stmt)
	case *tree.AlterSequence:
		return nodeAlterSequence(stmt)
	case *tree.AlterTable:
		return nodeAlterTable(stmt)
	case *tree.AlterTableSetSchema:
		return nodeAlterTableSetSchema(stmt)
	case *tree.AlterType:
		return nodeAlterType(stmt)
	case *tree.Analyze:
		return nodeAnalyze(stmt)
	case *tree.Backup:
		return nodeBackup(stmt)
	case *tree.BeginTransaction:
		return nodeBeginTransaction(stmt)
	case *tree.Call:
		return nodeCall(stmt)
	case *tree.CancelQueries:
		return nodeCancelQueries(stmt)
	case *tree.CancelSessions:
		return nodeCancelSessions(stmt)
	case *tree.CannedOptPlan:
		return nodeCannedOptPlan(stmt)
	case *tree.Comment:
		return nodeComment(stmt)
	case *tree.CommitTransaction:
		return nodeCommitTransaction(stmt)
	case *tree.ControlJobs:
		return nodeControlJobs(stmt)
	case *tree.ControlJobsForSchedules:
		return nodeControlJobsForSchedules(stmt)
	case *tree.ControlSchedules:
		return nodeControlSchedules(stmt)
	case *tree.CopyFrom:
		return nodeCopyFrom(stmt)
	case *tree.CreateChangefeed:
		return nodeCreateChangefeed(stmt)
	case *tree.CreateDatabase:
		return nodeCreateDatabase(stmt)
	case *tree.CreateFunction:
		return nodeCreateFunction(stmt)
	case *tree.CreateIndex:
		return nodeCreateIndex(stmt)
	case *tree.CreateProcedure:
		return nodeCreateProcedure(stmt)
	case *tree.CreateRole:
		return nodeCreateRole(stmt)
	case *tree.CreateSchema:
		return nodeCreateSchema(stmt)
	case *tree.CreateSequence:
		return nodeCreateSequence(stmt)
	case *tree.CreateStats:
		return nodeCreateStats(stmt)
	case *tree.CreateTable:
		return nodeCreateTable(stmt)
	case *tree.CreateTrigger:
		return nodeCreateTrigger(stmt)
	case *tree.CreateType:
		return nodeCreateType(stmt)
	case *tree.CreateView:
		return nodeCreateView(stmt)
	case *tree.Deallocate:
		return nodeDeallocate(stmt)
	case *tree.Delete:
		return nodeDelete(stmt)
	case *tree.Discard:
		return nodeDiscard(stmt)
	case *tree.DropDatabase:
		return nodeDropDatabase(stmt)
	case *tree.DropIndex:
		return nodeDropIndex(stmt)
	case *tree.DropRole:
		return nodeDropRole(stmt)
	case *tree.DropSchema:
		return nodeDropSchema(stmt)
	case *tree.DropSequence:
		return nodeDropSequence(stmt)
	case *tree.DropTable:
		return nodeDropTable(stmt)
	case *tree.DropTrigger:
		return nodeDropTrigger(stmt)
	case *tree.DropType:
		return nodeDropType(stmt)
	case *tree.DropView:
		return nodeDropView(stmt)
	case *tree.Execute:
		return nodeExecute(stmt)
	case *tree.Explain:
		return nodeExplain(stmt)
	case *tree.ExplainAnalyzeDebug:
		return nodeExplainAnalyzeDebug(stmt)
	case *tree.Export:
		return nodeExport(stmt)
	case *tree.Grant:
		return nodeGrant(stmt)
	case *tree.GrantRole:
		return nodeGrantRole(stmt)
	case *tree.Import:
		return nodeImport(stmt)
	case *tree.Insert:
		return nodeInsert(stmt)
	case *tree.ParenSelect:
		return nodeParenSelect(stmt)
	case *tree.Prepare:
		return nodePrepare(stmt)
	case *tree.RefreshMaterializedView:
		return nodeRefreshMaterializedView(stmt)
	case *tree.ReleaseSavepoint:
		return nodeReleaseSavepoint(stmt)
	case *tree.Relocate:
		return nodeRelocate(stmt)
	case *tree.RenameColumn:
		return nodeRenameColumn(stmt)
	case *tree.RenameDatabase:
		return nodeRenameDatabase(stmt)
	case *tree.RenameIndex:
		return nodeRenameIndex(stmt)
	case *tree.RenameTable:
		return nodeRenameTable(stmt)
	case *tree.ReparentDatabase:
		return nodeReparentDatabase(stmt)
	case *tree.Restore:
		return nodeRestore(stmt)
	case *tree.Revoke:
		return nodeRevoke(stmt)
	case *tree.RevokeRole:
		return nodeRevokeRole(stmt)
	case *tree.RollbackToSavepoint:
		return nodeRollbackToSavepoint(stmt)
	case *tree.RollbackTransaction:
		return nodeRollbackTransaction(stmt)
	case *tree.Savepoint:
		return nodeSavepoint(stmt)
	case *tree.Scatter:
		return nodeScatter(stmt)
	case *tree.ScheduledBackup:
		return nodeScheduledBackup(stmt)
	case *tree.Scrub:
		return nodeScrub(stmt)
	case *tree.Select:
		return nodeSelect(stmt)
	case *tree.SelectClause:
		return nodeSelectClause(stmt)
	case *tree.SetSessionAuthorization:
		return nodeSetSessionAuthorization(stmt)
	case *tree.SetSessionCharacteristics:
		return nodeSetSessionCharacteristics(stmt)
	case *tree.SetTransaction:
		return nodeSetTransaction(stmt)
	case *tree.SetVar:
		return nodeSetVar(stmt)
	case *tree.ShowBackup:
		return nodeShowBackup(stmt)
	case *tree.ShowColumns:
		return nodeShowColumns(stmt)
	case *tree.ShowConstraints:
		return nodeShowConstraints(stmt)
	case *tree.ShowCreate:
		return nodeShowCreate(stmt)
	case *tree.ShowDatabaseIndexes:
		return nodeShowDatabaseIndexes(stmt)
	case *tree.ShowDatabases:
		return nodeShowDatabases(stmt)
	case *tree.ShowEnums:
		return nodeShowEnums(stmt)
	case *tree.ShowFingerprints:
		return nodeShowFingerprints(stmt)
	case *tree.ShowGrants:
		return nodeShowGrants(stmt)
	case *tree.ShowHistogram:
		return nodeShowHistogram(stmt)
	case *tree.ShowIndexes:
		return nodeShowIndexes(stmt)
	case *tree.ShowJobs:
		return nodeShowJobs(stmt)
	case *tree.ShowLastQueryStatistics:
		return nodeShowLastQueryStatistics(stmt)
	case *tree.ShowPartitions:
		return nodeShowPartitions(stmt)
	case *tree.ShowQueries:
		return nodeShowQueries(stmt)
	case *tree.ShowRoleGrants:
		return nodeShowRoleGrants(stmt)
	case *tree.ShowRoles:
		return nodeShowRoles(stmt)
	case *tree.ShowSavepointStatus:
		return nodeShowSavepointStatus(stmt)
	case *tree.ShowSchedules:
		return nodeShowSchedules(stmt)
	case *tree.ShowSchemas:
		return nodeShowSchemas(stmt)
	case *tree.ShowSequences:
		return nodeShowSequences(stmt)
	case *tree.ShowSessions:
		return nodeShowSessions(stmt)
	case *tree.ShowSyntax:
		return nodeShowSyntax(stmt)
	case *tree.ShowTableStats:
		return nodeShowTableStats(stmt)
	case *tree.ShowTables:
		return nodeShowTables(stmt)
	case *tree.ShowTraceForSession:
		return nodeShowTraceForSession(stmt)
	case *tree.ShowTransactionStatus:
		return nodeShowTransactionStatus(stmt)
	case *tree.ShowTransactions:
		return nodeShowTransactions(stmt)
	case *tree.ShowTypes:
		return nodeShowTypes(stmt)
	case *tree.ShowUsers:
		return nodeShowUsers(stmt)
	case *tree.ShowVar:
		return nodeShowVar(stmt)
	case *tree.Split:
		return nodeSplit(stmt)
	case *tree.Truncate:
		return nodeTruncate(stmt)
	case *tree.UnionClause:
		return nodeUnionClause(stmt)
	case *tree.Unsplit:
		return nodeUnsplit(stmt)
	case *tree.Update:
		return nodeUpdate(stmt)
	case *tree.ValuesClause:
		return nodeValuesClause(stmt)
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown statement type encountered: `%T`", stmt)
	}
}
