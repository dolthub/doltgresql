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
	ctx := NewContext()
	switch stmt := postgresStmt.AST.(type) {
	case *tree.AlterAggregate:
		return nodeAlterAggregate(ctx, stmt)
	case *tree.AlterDatabase:
		return nodeAlterDatabase(ctx, stmt)
	case *tree.AlterFunction:
		return nodeAlterFunction(ctx, stmt)
	case *tree.AlterIndex:
		return nodeAlterIndex(ctx, stmt)
	case *tree.AlterProcedure:
		return nodeAlterProcedure(ctx, stmt)
	case *tree.AlterRole:
		return nodeAlterRole(ctx, stmt)
	case *tree.AlterSchema:
		return nodeAlterSchema(ctx, stmt)
	case *tree.AlterSequence:
		return nodeAlterSequence(ctx, stmt)
	case *tree.AlterTable:
		return nodeAlterTable(ctx, stmt)
	case *tree.AlterTablePartition:
		return nodeAlterTablePartition(ctx, stmt)
	case *tree.AlterTableSetSchema:
		return nodeAlterTableSetSchema(ctx, stmt)
	case *tree.AlterType:
		return nodeAlterType(ctx, stmt)
	case *tree.Analyze:
		return nodeAnalyze(ctx, stmt)
	case *tree.Backup:
		return nodeBackup(ctx, stmt)
	case *tree.BeginTransaction:
		return nodeBeginTransaction(ctx, stmt)
	case *tree.Call:
		return nodeCall(ctx, stmt)
	case *tree.CancelQueries:
		return nodeCancelQueries(ctx, stmt)
	case *tree.CancelSessions:
		return nodeCancelSessions(ctx, stmt)
	case *tree.CannedOptPlan:
		return nodeCannedOptPlan(ctx, stmt)
	case *tree.Comment:
		return nodeComment(ctx, stmt)
	case *tree.CommitTransaction:
		return nodeCommitTransaction(ctx, stmt)
	case *tree.ControlJobs:
		return nodeControlJobs(ctx, stmt)
	case *tree.ControlJobsForSchedules:
		return nodeControlJobsForSchedules(ctx, stmt)
	case *tree.ControlSchedules:
		return nodeControlSchedules(ctx, stmt)
	case *tree.CopyFrom:
		return nodeCopyFrom(ctx, stmt)
	case *tree.CreateAggregate:
		return nodeCreateAggregate(ctx, stmt)
	case *tree.CreateChangefeed:
		return nodeCreateChangefeed(ctx, stmt)
	case *tree.CreateDatabase:
		return nodeCreateDatabase(ctx, stmt)
	case *tree.CreateDomain:
		return nodeCreateDomain(ctx, stmt)
	case *tree.CreateFunction:
		return nodeCreateFunction(ctx, stmt)
	case *tree.CreateIndex:
		return nodeCreateIndex(ctx, stmt)
	case *tree.CreateProcedure:
		return nodeCreateProcedure(ctx, stmt)
	case *tree.CreateRole:
		return nodeCreateRole(ctx, stmt)
	case *tree.CreateSchema:
		return nodeCreateSchema(ctx, stmt)
	case *tree.CreateSequence:
		return nodeCreateSequence(ctx, stmt)
	case *tree.CreateStats:
		return nodeCreateStats(ctx, stmt)
	case *tree.CreateTable:
		return nodeCreateTable(ctx, stmt)
	case *tree.CreateTrigger:
		return nodeCreateTrigger(ctx, stmt)
	case *tree.CreateType:
		return nodeCreateType(ctx, stmt)
	case *tree.CreateView:
		return nodeCreateView(ctx, stmt)
	case *tree.Deallocate:
		return nodeDeallocate(ctx, stmt)
	case *tree.Delete:
		return nodeDelete(ctx, stmt)
	case *tree.Discard:
		return nodeDiscard(ctx, stmt)
	case *tree.DropAggregate:
		return nodeDropAggregate(ctx, stmt)
	case *tree.DropDatabase:
		return nodeDropDatabase(ctx, stmt)
	case *tree.DropDomain:
		return nodeDropDomain(ctx, stmt)
	case *tree.DropIndex:
		return nodeDropIndex(ctx, stmt)
	case *tree.DropRole:
		return nodeDropRole(ctx, stmt)
	case *tree.DropSchema:
		return nodeDropSchema(ctx, stmt)
	case *tree.DropSequence:
		return nodeDropSequence(ctx, stmt)
	case *tree.DropTable:
		return nodeDropTable(ctx, stmt)
	case *tree.DropTrigger:
		return nodeDropTrigger(ctx, stmt)
	case *tree.DropType:
		return nodeDropType(ctx, stmt)
	case *tree.DropView:
		return nodeDropView(ctx, stmt)
	case *tree.Execute:
		return nodeExecute(ctx, stmt)
	case *tree.Explain:
		return nodeExplain(ctx, stmt)
	case *tree.ExplainAnalyzeDebug:
		return nodeExplainAnalyzeDebug(ctx, stmt)
	case *tree.Export:
		return nodeExport(ctx, stmt)
	case *tree.Grant:
		return nodeGrant(ctx, stmt)
	case *tree.GrantRole:
		return nodeGrantRole(ctx, stmt)
	case *tree.Import:
		return nodeImport(ctx, stmt)
	case *tree.Insert:
		return nodeInsert(ctx, stmt)
	case *tree.ParenSelect:
		return nodeParenSelect(ctx, stmt)
	case *tree.Prepare:
		return nodePrepare(ctx, stmt)
	case *tree.RefreshMaterializedView:
		return nodeRefreshMaterializedView(ctx, stmt)
	case *tree.ReleaseSavepoint:
		return nodeReleaseSavepoint(ctx, stmt)
	case *tree.Relocate:
		return nodeRelocate(ctx, stmt)
	case *tree.RenameColumn:
		return nodeRenameColumn(ctx, stmt)
	case *tree.RenameDatabase:
		return nodeRenameDatabase(ctx, stmt)
	case *tree.RenameIndex:
		return nodeRenameIndex(ctx, stmt)
	case *tree.RenameTable:
		return nodeRenameTable(ctx, stmt)
	case *tree.ReparentDatabase:
		return nodeReparentDatabase(ctx, stmt)
	case *tree.Restore:
		return nodeRestore(ctx, stmt)
	case *tree.Revoke:
		return nodeRevoke(ctx, stmt)
	case *tree.RevokeRole:
		return nodeRevokeRole(ctx, stmt)
	case *tree.RollbackToSavepoint:
		return nodeRollbackToSavepoint(ctx, stmt)
	case *tree.RollbackTransaction:
		return nodeRollbackTransaction(ctx, stmt)
	case *tree.Savepoint:
		return nodeSavepoint(ctx, stmt)
	case *tree.Scatter:
		return nodeScatter(ctx, stmt)
	case *tree.ScheduledBackup:
		return nodeScheduledBackup(ctx, stmt)
	case *tree.Scrub:
		return nodeScrub(ctx, stmt)
	case *tree.Select:
		return nodeSelect(ctx, stmt)
	case *tree.SelectClause:
		return nodeSelectClause(ctx, stmt)
	case *tree.SetSessionAuthorization:
		return nodeSetSessionAuthorization(ctx, stmt)
	case *tree.SetSessionCharacteristics:
		return nodeSetSessionCharacteristics(ctx, stmt)
	case *tree.SetTransaction:
		return nodeSetTransaction(ctx, stmt)
	case *tree.SetVar:
		return nodeSetVar(ctx, stmt)
	case *tree.ShowBackup:
		return nodeShowBackup(ctx, stmt)
	case *tree.ShowColumns:
		return nodeShowColumns(ctx, stmt)
	case *tree.ShowConstraints:
		return nodeShowConstraints(ctx, stmt)
	case *tree.ShowCreate:
		return nodeShowCreate(ctx, stmt)
	case *tree.ShowDatabaseIndexes:
		return nodeShowDatabaseIndexes(ctx, stmt)
	case *tree.ShowDatabases:
		return nodeShowDatabases(ctx, stmt)
	case *tree.ShowEnums:
		return nodeShowEnums(ctx, stmt)
	case *tree.ShowFingerprints:
		return nodeShowFingerprints(ctx, stmt)
	case *tree.ShowGrants:
		return nodeShowGrants(ctx, stmt)
	case *tree.ShowHistogram:
		return nodeShowHistogram(ctx, stmt)
	case *tree.ShowIndexes:
		return nodeShowIndexes(ctx, stmt)
	case *tree.ShowJobs:
		return nodeShowJobs(ctx, stmt)
	case *tree.ShowLastQueryStatistics:
		return nodeShowLastQueryStatistics(ctx, stmt)
	case *tree.ShowPartitions:
		return nodeShowPartitions(ctx, stmt)
	case *tree.ShowQueries:
		return nodeShowQueries(ctx, stmt)
	case *tree.ShowRoleGrants:
		return nodeShowRoleGrants(ctx, stmt)
	case *tree.ShowRoles:
		return nodeShowRoles(ctx, stmt)
	case *tree.ShowSavepointStatus:
		return nodeShowSavepointStatus(ctx, stmt)
	case *tree.ShowSchedules:
		return nodeShowSchedules(ctx, stmt)
	case *tree.ShowSchemas:
		return nodeShowSchemas(ctx, stmt)
	case *tree.ShowSequences:
		return nodeShowSequences(ctx, stmt)
	case *tree.ShowSessions:
		return nodeShowSessions(ctx, stmt)
	case *tree.ShowSyntax:
		return nodeShowSyntax(ctx, stmt)
	case *tree.ShowTableStats:
		return nodeShowTableStats(ctx, stmt)
	case *tree.ShowTables:
		return nodeShowTables(ctx, stmt)
	case *tree.ShowTraceForSession:
		return nodeShowTraceForSession(ctx, stmt)
	case *tree.ShowTransactionStatus:
		return nodeShowTransactionStatus(ctx, stmt)
	case *tree.ShowTransactions:
		return nodeShowTransactions(ctx, stmt)
	case *tree.ShowTypes:
		return nodeShowTypes(ctx, stmt)
	case *tree.ShowUsers:
		return nodeShowUsers(ctx, stmt)
	case *tree.ShowVar:
		return nodeShowVar(ctx, stmt)
	case *tree.Split:
		return nodeSplit(ctx, stmt)
	case *tree.Truncate:
		return nodeTruncate(ctx, stmt)
	case *tree.UnionClause:
		return nodeUnionClause(ctx, stmt)
	case *tree.Unsplit:
		return nodeUnsplit(ctx, stmt)
	case *tree.Update:
		return nodeUpdate(ctx, stmt)
	case *tree.ValuesClause:
		return nodeValuesClause(ctx, stmt)
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown statement type encountered: `%T`", stmt)
	}
}
