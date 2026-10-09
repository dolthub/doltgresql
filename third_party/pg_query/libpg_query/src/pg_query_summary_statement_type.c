#include "pg_query.h"
#include "pg_query_internal.h"
#include "pg_query_outfuncs.h"

#include "parser/parser.h"
#include "parser/scanner.h"
#include "parser/scansup.h"

#include "nodes/nodeFuncs.h"
#include "utils/builtins.h"

#include <unistd.h>
#include <fcntl.h>
#include <string.h>

#include "pg_query_summary.h"

/* Declare the huge function early so we can put it at the end of the file. */
static bool pg_query_summary_stmt_walk_impl(Node *node, Summary * summary);

/*
 * Adds new_stmt_type to summary->statement_types, if it's not already in it.
 * What we really want is a set, but I'd rather not implement one in C.
 */
void
add_stmt_type_if_needed(Summary * summary, char *new_stmt_type)
{
	ListCell   *lc = NULL;

	foreach(lc, summary->statement_types)
	{
		char	   *stmt_type = lfirst(lc);

		if (strcmp(new_stmt_type, stmt_type) == 0)
			return;
	}

	summary->statement_types = lappend(summary->statement_types, new_stmt_type);
}


void
pg_query_summary_statement_walk(Summary * summary, Node *tree)
{
	pg_query_summary_stmt_walk_impl(tree, summary);
}

static bool
pg_query_summary_stmt_walk_impl(Node *node, Summary * summary)
{
	if (node == NULL)
		return false;

	switch (nodeTag(node))
	{
		case T_AlterCollationStmt:
			add_stmt_type_if_needed(summary, "AlterCollationStmt");
			break;
		case T_AlterDatabaseSetStmt:
			add_stmt_type_if_needed(summary, "AlterDatabaseSetStmt");
			break;
		case T_AlterDatabaseStmt:
			add_stmt_type_if_needed(summary, "AlterDatabaseStmt");
			break;
		case T_AlterDefaultPrivilegesStmt:
			add_stmt_type_if_needed(summary, "AlterDefaultPrivilegesStmt");
			break;
		case T_AlterDomainStmt:
			add_stmt_type_if_needed(summary, "AlterDomainStmt");
			break;
		case T_AlterEnumStmt:
			add_stmt_type_if_needed(summary, "AlterEnumStmt");
			break;
		case T_AlterEventTrigStmt:
			add_stmt_type_if_needed(summary, "AlterEventTrigStmt");
			break;
		case T_AlterExtensionContentsStmt:
			add_stmt_type_if_needed(summary, "AlterExtensionContentsStmt");
			break;
		case T_AlterExtensionStmt:
			add_stmt_type_if_needed(summary, "AlterExtensionStmt");
			break;
		case T_AlterFdwStmt:
			add_stmt_type_if_needed(summary, "AlterFdwStmt");
			break;
		case T_AlterForeignServerStmt:
			add_stmt_type_if_needed(summary, "AlterForeignServerStmt");
			break;
		case T_AlterFunctionStmt:
			add_stmt_type_if_needed(summary, "AlterFunctionStmt");
			break;
		case T_AlterObjectDependsStmt:
			add_stmt_type_if_needed(summary, "AlterObjectDependsStmt");
			break;
		case T_AlterObjectSchemaStmt:
			add_stmt_type_if_needed(summary, "AlterObjectSchemaStmt");
			break;
		case T_AlterOpFamilyStmt:
			add_stmt_type_if_needed(summary, "AlterOpFamilyStmt");
			break;
		case T_AlterOperatorStmt:
			add_stmt_type_if_needed(summary, "AlterOperatorStmt");
			break;
		case T_AlterOwnerStmt:
			add_stmt_type_if_needed(summary, "AlterOwnerStmt");
			break;
		case T_AlterPolicyStmt:
			add_stmt_type_if_needed(summary, "AlterPolicyStmt");
			break;
		case T_AlterPublicationStmt:
			add_stmt_type_if_needed(summary, "AlterPublicationStmt");
			break;
		case T_AlterRoleSetStmt:
			add_stmt_type_if_needed(summary, "AlterRoleSetStmt");
			break;
		case T_AlterRoleStmt:
			add_stmt_type_if_needed(summary, "AlterRoleStmt");
			break;
		case T_AlterSeqStmt:
			add_stmt_type_if_needed(summary, "AlterSeqStmt");
			break;
		case T_AlterStatsStmt:
			add_stmt_type_if_needed(summary, "AlterStatsStmt");
			break;
		case T_AlterSubscriptionStmt:
			add_stmt_type_if_needed(summary, "AlterSubscriptionStmt");
			break;
		case T_AlterSystemStmt:
			add_stmt_type_if_needed(summary, "AlterSystemStmt");
			break;
		case T_AlterTableCmd:
			add_stmt_type_if_needed(summary, "AlterTableCmd");
			break;
		case T_AlterTableMoveAllStmt:
			add_stmt_type_if_needed(summary, "AlterTableMoveAllStmt");
			break;
		case T_AlterTableSpaceOptionsStmt:
			add_stmt_type_if_needed(summary, "AlterTableSpaceOptionsStmt");
			break;
		case T_AlterTableStmt:
			add_stmt_type_if_needed(summary, "AlterTableStmt");
			break;
		case T_AlterTSConfigurationStmt:
			add_stmt_type_if_needed(summary, "AlterTSConfigurationStmt");
			break;
		case T_AlterTSDictionaryStmt:
			add_stmt_type_if_needed(summary, "AlterTSDictionaryStmt");
			break;
		case T_AlterTypeStmt:
			add_stmt_type_if_needed(summary, "AlterTypeStmt");
			break;
		case T_AlterUserMappingStmt:
			add_stmt_type_if_needed(summary, "AlterUserMappingStmt");
			break;
		case T_CallStmt:
			add_stmt_type_if_needed(summary, "CallStmt");
			break;
		case T_CheckPointStmt:
			add_stmt_type_if_needed(summary, "CheckPointStmt");
			break;
		case T_ClosePortalStmt:
			add_stmt_type_if_needed(summary, "ClosePortalStmt");
			break;
		case T_ClusterStmt:
			add_stmt_type_if_needed(summary, "ClusterStmt");
			break;
		case T_CommentStmt:
			add_stmt_type_if_needed(summary, "CommentStmt");
			break;
		case T_CompositeTypeStmt:
			add_stmt_type_if_needed(summary, "CompositeTypeStmt");
			break;
		case T_ConstraintsSetStmt:
			add_stmt_type_if_needed(summary, "ConstraintsSetStmt");
			break;
		case T_CopyStmt:
			add_stmt_type_if_needed(summary, "CopyStmt");
			break;
		case T_CreateAmStmt:
			add_stmt_type_if_needed(summary, "CreateAmStmt");
			break;
		case T_CreateCastStmt:
			add_stmt_type_if_needed(summary, "CreateCastStmt");
			break;
		case T_CreateConversionStmt:
			add_stmt_type_if_needed(summary, "CreateConversionStmt");
			break;
		case T_CreateDomainStmt:
			add_stmt_type_if_needed(summary, "CreateDomainStmt");
			break;
		case T_CreateEnumStmt:
			add_stmt_type_if_needed(summary, "CreateEnumStmt");
			break;
		case T_CreateEventTrigStmt:
			add_stmt_type_if_needed(summary, "CreateEventTrigStmt");
			break;
		case T_CreateExtensionStmt:
			add_stmt_type_if_needed(summary, "CreateExtensionStmt");
			break;
		case T_CreateFdwStmt:
			add_stmt_type_if_needed(summary, "CreateFdwStmt");
			break;
		case T_CreateForeignServerStmt:
			add_stmt_type_if_needed(summary, "CreateForeignServerStmt");
			break;
		case T_CreateForeignTableStmt:
			add_stmt_type_if_needed(summary, "CreateForeignTableStmt");
			break;
		case T_CreateFunctionStmt:
			add_stmt_type_if_needed(summary, "CreateFunctionStmt");
			break;
		case T_CreateOpClassStmt:
			add_stmt_type_if_needed(summary, "CreateOpClassStmt");
			break;
		case T_CreateOpFamilyStmt:
			add_stmt_type_if_needed(summary, "CreateOpFamilyStmt");
			break;
		case T_CreatePLangStmt:
			add_stmt_type_if_needed(summary, "CreatePLangStmt");
			break;
		case T_CreatePolicyStmt:
			add_stmt_type_if_needed(summary, "CreatePolicyStmt");
			break;
		case T_CreatePublicationStmt:
			add_stmt_type_if_needed(summary, "CreatePublicationStmt");
			break;
		case T_CreateRangeStmt:
			add_stmt_type_if_needed(summary, "CreateRangeStmt");
			break;
		case T_CreateRoleStmt:
			add_stmt_type_if_needed(summary, "CreateRoleStmt");
			break;
		case T_CreateSchemaStmt:
			add_stmt_type_if_needed(summary, "CreateSchemaStmt");
			break;
		case T_CreateSeqStmt:
			add_stmt_type_if_needed(summary, "CreateSeqStmt");
			break;
		case T_CreateStatsStmt:
			add_stmt_type_if_needed(summary, "CreateStatsStmt");
			break;
		case T_CreateStmt:
			add_stmt_type_if_needed(summary, "CreateStmt");
			break;
		case T_CreateSubscriptionStmt:
			add_stmt_type_if_needed(summary, "CreateSubscriptionStmt");
			break;
		case T_CreateTableAsStmt:
			{
				add_stmt_type_if_needed(summary, "CreateTableAsStmt");
				CreateTableAsStmt *stmt = (CreateTableAsStmt *) node;

				if (stmt->query)
					pg_query_summary_stmt_walk_impl(stmt->query, summary);
				break;
			}
		case T_CreateTableSpaceStmt:
			add_stmt_type_if_needed(summary, "CreateTableSpaceStmt");
			break;
		case T_CreateTransformStmt:
			add_stmt_type_if_needed(summary, "CreateTransformStmt");
			break;
		case T_CreateTrigStmt:
			add_stmt_type_if_needed(summary, "CreateTrigStmt");
			break;
		case T_CreateUserMappingStmt:
			add_stmt_type_if_needed(summary, "CreateUserMappingStmt");
			break;
		case T_CreatedbStmt:
			add_stmt_type_if_needed(summary, "CreatedbStmt");
			break;
		case T_DeallocateStmt:
			add_stmt_type_if_needed(summary, "DeallocateStmt");
			break;
		case T_DeclareCursorStmt:
			add_stmt_type_if_needed(summary, "DeclareCursorStmt");
			break;
		case T_DefineStmt:
			add_stmt_type_if_needed(summary, "DefineStmt");
			break;
		case T_DeleteStmt:
			add_stmt_type_if_needed(summary, "DeleteStmt");
			break;
		case T_DiscardStmt:
			add_stmt_type_if_needed(summary, "DiscardStmt");
			break;
		case T_DoStmt:
			add_stmt_type_if_needed(summary, "DoStmt");
			break;
		case T_DropOwnedStmt:
			add_stmt_type_if_needed(summary, "DropOwnedStmt");
			break;
		case T_DropRoleStmt:
			add_stmt_type_if_needed(summary, "DropRoleStmt");
			break;
		case T_DropStmt:
			add_stmt_type_if_needed(summary, "DropStmt");
			break;
		case T_DropSubscriptionStmt:
			add_stmt_type_if_needed(summary, "DropSubscriptionStmt");
			break;
		case T_DropTableSpaceStmt:
			add_stmt_type_if_needed(summary, "DropTableSpaceStmt");
			break;
		case T_DropUserMappingStmt:
			add_stmt_type_if_needed(summary, "DropUserMappingStmt");
			break;
		case T_DropdbStmt:
			add_stmt_type_if_needed(summary, "DropdbStmt");
			break;
		case T_ExecuteStmt:
			add_stmt_type_if_needed(summary, "ExecuteStmt");
			break;
		case T_ExplainStmt:
			{
				add_stmt_type_if_needed(summary, "ExplainStmt");
				ExplainStmt *stmt = (ExplainStmt *) node;

				if (stmt->query)
					pg_query_summary_stmt_walk_impl(stmt->query, summary);
				break;
			}
		case T_FetchStmt:
			add_stmt_type_if_needed(summary, "FetchStmt");
			break;
		case T_GrantRoleStmt:
			add_stmt_type_if_needed(summary, "GrantRoleStmt");
			break;
		case T_GrantStmt:
			add_stmt_type_if_needed(summary, "GrantStmt");
			break;
		case T_ImportForeignSchemaStmt:
			add_stmt_type_if_needed(summary, "ImportForeignSchemaStmt");
			break;
		case T_IndexStmt:
			add_stmt_type_if_needed(summary, "IndexStmt");
			break;
		case T_InsertStmt:
			add_stmt_type_if_needed(summary, "InsertStmt");
			break;
		case T_ListenStmt:
			add_stmt_type_if_needed(summary, "ListenStmt");
			break;
		case T_LoadStmt:
			add_stmt_type_if_needed(summary, "LoadStmt");
			break;
		case T_LockStmt:
			add_stmt_type_if_needed(summary, "LockStmt");
			break;
		case T_MergeStmt:
			add_stmt_type_if_needed(summary, "MergeStmt");
			break;
		case T_NotifyStmt:
			add_stmt_type_if_needed(summary, "NotifyStmt");
			break;
		case T_PrepareStmt:
			{
				add_stmt_type_if_needed(summary, "PrepareStmt");
				PrepareStmt *stmt = (PrepareStmt *) node;

				if (stmt->query)
					pg_query_summary_stmt_walk_impl(stmt->query, summary);
				break;
			}
		case T_ReassignOwnedStmt:
			add_stmt_type_if_needed(summary, "ReassignOwnedStmt");
			break;
		case T_RefreshMatViewStmt:
			add_stmt_type_if_needed(summary, "RefreshMatViewStmt");
			break;
		case T_ReindexStmt:
			add_stmt_type_if_needed(summary, "ReindexStmt");
			break;
		case T_RenameStmt:
			add_stmt_type_if_needed(summary, "RenameStmt");
			break;
		case T_ReplicaIdentityStmt:
			add_stmt_type_if_needed(summary, "ReplicaIdentityStmt");
			break;
		case T_RuleStmt:
			add_stmt_type_if_needed(summary, "RuleStmt");
			break;
		case T_SecLabelStmt:
			add_stmt_type_if_needed(summary, "SecLabelStmt");
			break;
		case T_SelectStmt:
			add_stmt_type_if_needed(summary, "SelectStmt");
			break;
		case T_SetOperationStmt:
			add_stmt_type_if_needed(summary, "SetOperationStmt");
			break;
		case T_TransactionStmt:
			add_stmt_type_if_needed(summary, "TransactionStmt");
			break;
		case T_TruncateStmt:
			add_stmt_type_if_needed(summary, "TruncateStmt");
			break;
		case T_UnlistenStmt:
			add_stmt_type_if_needed(summary, "UnlistenStmt");
			break;
		case T_UpdateStmt:
			add_stmt_type_if_needed(summary, "UpdateStmt");
			break;
		case T_VacuumStmt:
			add_stmt_type_if_needed(summary, "VacuumStmt");
			break;
		case T_VariableSetStmt:
			add_stmt_type_if_needed(summary, "VariableSetStmt");
			break;
		case T_VariableShowStmt:
			add_stmt_type_if_needed(summary, "VariableShowStmt");
			break;
		case T_ViewStmt:
			{
				add_stmt_type_if_needed(summary, "ViewStmt");
				ViewStmt   *stmt = (ViewStmt *) node;

				if (stmt->query)
					pg_query_summary_stmt_walk_impl(stmt->query, summary);
				break;
			}
		case T_RawStmt:
			return pg_query_summary_stmt_walk_impl(((RawStmt *) node)->stmt, summary);

			/*
			 * All the cases below here are not handled in the Rust version.
			 * They're added here so, if a new statement type (like MergeStmt)
			 * is added, the C compiler will generate a warning.
			 */
		case T_Invalid:
		case T_List:
		case T_Alias:
		case T_RangeVar:
		case T_TableFunc:
		case T_IntoClause:
		case T_Var:
		case T_Const:
		case T_Param:
		case T_Aggref:
		case T_GroupingFunc:
		case T_WindowFunc:
		case T_WindowFuncRunCondition:
		case T_MergeSupportFunc:
		case T_SubscriptingRef:
		case T_FuncExpr:
		case T_NamedArgExpr:
		case T_OpExpr:
		case T_DistinctExpr:
		case T_NullIfExpr:
		case T_ScalarArrayOpExpr:
		case T_BoolExpr:
		case T_SubLink:
		case T_SubPlan:
		case T_AlternativeSubPlan:
		case T_FieldSelect:
		case T_FieldStore:
		case T_RelabelType:
		case T_CoerceViaIO:
		case T_ArrayCoerceExpr:
		case T_ConvertRowtypeExpr:
		case T_CollateExpr:
		case T_CaseExpr:
		case T_CaseWhen:
		case T_CaseTestExpr:
		case T_ArrayExpr:
		case T_RowExpr:
		case T_RowCompareExpr:
		case T_CoalesceExpr:
		case T_MinMaxExpr:
		case T_SQLValueFunction:
		case T_XmlExpr:
		case T_JsonFormat:
		case T_JsonReturning:
		case T_JsonValueExpr:
		case T_JsonConstructorExpr:
		case T_JsonIsPredicate:
		case T_JsonBehavior:
		case T_JsonExpr:
		case T_JsonTablePath:
		case T_JsonTablePathScan:
		case T_JsonTableSiblingJoin:
		case T_NullTest:
		case T_BooleanTest:
		case T_MergeAction:
		case T_CoerceToDomain:
		case T_CoerceToDomainValue:
		case T_SetToDefault:
		case T_CurrentOfExpr:
		case T_NextValueExpr:
		case T_InferenceElem:
		case T_TargetEntry:
		case T_RangeTblRef:
		case T_JoinExpr:
		case T_FromExpr:
		case T_OnConflictExpr:
		case T_Query:
		case T_TypeName:
		case T_ColumnRef:
		case T_ParamRef:
		case T_A_Expr:
		case T_A_Const:
		case T_TypeCast:
		case T_CollateClause:
		case T_RoleSpec:
		case T_FuncCall:
		case T_A_Star:
		case T_A_Indices:
		case T_A_Indirection:
		case T_A_ArrayExpr:
		case T_ResTarget:
		case T_MultiAssignRef:
		case T_SortBy:
		case T_WindowDef:
		case T_RangeSubselect:
		case T_RangeFunction:
		case T_RangeTableFunc:
		case T_RangeTableFuncCol:
		case T_RangeTableSample:
		case T_ColumnDef:
		case T_TableLikeClause:
		case T_IndexElem:
		case T_DefElem:
		case T_LockingClause:
		case T_XmlSerialize:
		case T_PartitionElem:
		case T_PartitionSpec:
		case T_PartitionBoundSpec:
		case T_PartitionRangeDatum:
		case T_SinglePartitionSpec:
		case T_PartitionCmd:
		case T_RangeTblEntry:
		case T_RTEPermissionInfo:
		case T_RangeTblFunction:
		case T_TableSampleClause:
		case T_WithCheckOption:
		case T_SortGroupClause:
		case T_GroupingSet:
		case T_WindowClause:
		case T_RowMarkClause:
		case T_WithClause:
		case T_InferClause:
		case T_OnConflictClause:
		case T_CTESearchClause:
		case T_CTECycleClause:
		case T_CommonTableExpr:
		case T_MergeWhenClause:
		case T_TriggerTransition:
		case T_JsonOutput:
		case T_JsonArgument:
		case T_JsonFuncExpr:
		case T_JsonTablePathSpec:
		case T_JsonTable:
		case T_JsonTableColumn:
		case T_JsonKeyValue:
		case T_JsonParseExpr:
		case T_JsonScalarExpr:
		case T_JsonSerializeExpr:
		case T_JsonObjectConstructor:
		case T_JsonArrayConstructor:
		case T_JsonArrayQueryConstructor:
		case T_JsonAggConstructor:
		case T_JsonObjectAgg:
		case T_JsonArrayAgg:
		case T_ReturnStmt:
		case T_PLAssignStmt:
		case T_ObjectWithArgs:
		case T_AccessPriv:
		case T_Constraint:
		case T_CreateOpClassItem:
		case T_StatsElem:
		case T_FunctionParameter:
		case T_InlineCodeBlock:
		case T_CallContext:
		case T_AlterDatabaseRefreshCollStmt:
		case T_VacuumRelation:
		case T_PublicationTable:
		case T_PublicationObjSpec:
		case T_PlannerGlobal:
		case T_PlannerInfo:
		case T_RelOptInfo:
		case T_IndexOptInfo:
		case T_ForeignKeyOptInfo:
		case T_StatisticExtInfo:
		case T_JoinDomain:
		case T_EquivalenceClass:
		case T_EquivalenceMember:
		case T_PathKey:
		case T_GroupByOrdering:
		case T_PathTarget:
		case T_ParamPathInfo:
		case T_Path:
		case T_IndexPath:
		case T_IndexClause:
		case T_BitmapHeapPath:
		case T_BitmapAndPath:
		case T_BitmapOrPath:
		case T_TidPath:
		case T_TidRangePath:
		case T_SubqueryScanPath:
		case T_ForeignPath:
		case T_CustomPath:
		case T_AppendPath:
		case T_MergeAppendPath:
		case T_GroupResultPath:
		case T_MaterialPath:
		case T_MemoizePath:
		case T_UniquePath:
		case T_GatherPath:
		case T_GatherMergePath:
		case T_NestPath:
		case T_MergePath:
		case T_HashPath:
		case T_ProjectionPath:
		case T_ProjectSetPath:
		case T_SortPath:
		case T_IncrementalSortPath:
		case T_GroupPath:
		case T_UpperUniquePath:
		case T_AggPath:
		case T_GroupingSetData:
		case T_RollupData:
		case T_GroupingSetsPath:
		case T_MinMaxAggPath:
		case T_WindowAggPath:
		case T_SetOpPath:
		case T_RecursiveUnionPath:
		case T_LockRowsPath:
		case T_ModifyTablePath:
		case T_LimitPath:
		case T_RestrictInfo:
		case T_PlaceHolderVar:
		case T_SpecialJoinInfo:
		case T_OuterJoinClauseInfo:
		case T_AppendRelInfo:
		case T_RowIdentityVarInfo:
		case T_PlaceHolderInfo:
		case T_MinMaxAggInfo:
		case T_PlannerParamItem:
		case T_AggInfo:
		case T_AggTransInfo:
		case T_PlannedStmt:
		case T_Result:
		case T_ProjectSet:
		case T_ModifyTable:
		case T_Append:
		case T_MergeAppend:
		case T_RecursiveUnion:
		case T_BitmapAnd:
		case T_BitmapOr:
		case T_SeqScan:
		case T_SampleScan:
		case T_IndexScan:
		case T_IndexOnlyScan:
		case T_BitmapIndexScan:
		case T_BitmapHeapScan:
		case T_TidScan:
		case T_TidRangeScan:
		case T_SubqueryScan:
		case T_FunctionScan:
		case T_ValuesScan:
		case T_TableFuncScan:
		case T_CteScan:
		case T_NamedTuplestoreScan:
		case T_WorkTableScan:
		case T_ForeignScan:
		case T_CustomScan:
		case T_NestLoop:
		case T_NestLoopParam:
		case T_MergeJoin:
		case T_HashJoin:
		case T_Material:
		case T_Memoize:
		case T_Sort:
		case T_IncrementalSort:
		case T_Group:
		case T_Agg:
		case T_WindowAgg:
		case T_Unique:
		case T_Gather:
		case T_GatherMerge:
		case T_Hash:
		case T_SetOp:
		case T_LockRows:
		case T_Limit:
		case T_PlanRowMark:
		case T_PartitionPruneInfo:
		case T_PartitionedRelPruneInfo:
		case T_PartitionPruneStepOp:
		case T_PartitionPruneStepCombine:
		case T_PlanInvalItem:
		case T_ExprState:
		case T_IndexInfo:
		case T_ExprContext:
		case T_ReturnSetInfo:
		case T_ProjectionInfo:
		case T_JunkFilter:
		case T_OnConflictSetState:
		case T_MergeActionState:
		case T_ResultRelInfo:
		case T_EState:
		case T_WindowFuncExprState:
		case T_SetExprState:
		case T_SubPlanState:
		case T_DomainConstraintState:
		case T_ResultState:
		case T_ProjectSetState:
		case T_ModifyTableState:
		case T_AppendState:
		case T_MergeAppendState:
		case T_RecursiveUnionState:
		case T_BitmapAndState:
		case T_BitmapOrState:
		case T_ScanState:
		case T_SeqScanState:
		case T_SampleScanState:
		case T_IndexScanState:
		case T_IndexOnlyScanState:
		case T_BitmapIndexScanState:
		case T_BitmapHeapScanState:
		case T_TidScanState:
		case T_TidRangeScanState:
		case T_SubqueryScanState:
		case T_FunctionScanState:
		case T_ValuesScanState:
		case T_TableFuncScanState:
		case T_CteScanState:
		case T_NamedTuplestoreScanState:
		case T_WorkTableScanState:
		case T_ForeignScanState:
		case T_CustomScanState:
		case T_JoinState:
		case T_NestLoopState:
		case T_MergeJoinState:
		case T_HashJoinState:
		case T_MaterialState:
		case T_MemoizeState:
		case T_SortState:
		case T_IncrementalSortState:
		case T_GroupState:
		case T_AggState:
		case T_WindowAggState:
		case T_UniqueState:
		case T_GatherState:
		case T_GatherMergeState:
		case T_HashState:
		case T_SetOpState:
		case T_LockRowsState:
		case T_LimitState:
		case T_IndexAmRoutine:
		case T_TableAmRoutine:
		case T_TsmRoutine:
		case T_EventTriggerData:
		case T_TriggerData:
		case T_TupleTableSlot:
		case T_FdwRoutine:
		case T_Bitmapset:
		case T_ExtensibleNode:
		case T_ErrorSaveContext:
		case T_IdentifySystemCmd:
		case T_BaseBackupCmd:
		case T_CreateReplicationSlotCmd:
		case T_DropReplicationSlotCmd:
		case T_AlterReplicationSlotCmd:
		case T_StartReplicationCmd:
		case T_ReadReplicationSlotCmd:
		case T_TimeLineHistoryCmd:
		case T_UploadManifestCmd:
		case T_SupportRequestSimplify:
		case T_SupportRequestSelectivity:
		case T_SupportRequestCost:
		case T_SupportRequestRows:
		case T_SupportRequestIndexCondition:
		case T_SupportRequestWFuncMonotonic:
		case T_SupportRequestOptimizeWindowClause:
		case T_Integer:
		case T_Float:
		case T_Boolean:
		case T_String:
		case T_BitString:
		case T_ForeignKeyCacheInfo:
		case T_IntList:
		case T_OidList:
		case T_XidList:
		case T_AllocSetContext:
		case T_GenerationContext:
		case T_SlabContext:
		case T_BumpContext:
		case T_TIDBitmap:
		case T_WindowObjectData:
			break;
	}

	if (!pg_query_raw_tree_walker_supports(node))
		return false;

	return raw_expression_tree_walker(node, pg_query_summary_stmt_walk_impl, (void *) summary);
}
