#include <stdbool.h>

#include "pg_query.h"
#include "pg_query_internal.h"
#include "pg_query_outfuncs.h"

#include "parser/parser.h"
#include "parser/scanner.h"
#include "parser/scansup.h"

#include "nodes/execnodes.h"
#include "nodes/nodeFuncs.h"
#include "nodes/pathnodes.h"

#include "utils/builtins.h"

/*
 * Returns true if raw_expression_tree_walker() handles it,
 * otherwise, returns false.
 *
 * Up-to-date with postgres 17' implementation as of Aug 19 2025.
 * https://github.com/postgres/postgres/blob/REL_17_STABLE/src/backend/nodes/nodeFuncs.c#L3964
 */
bool
pg_query_raw_tree_walker_supports(Node *node)
{
	switch (nodeTag(node))
	{
		case T_JsonFormat:
		case T_SetToDefault:
		case T_CurrentOfExpr:
		case T_SQLValueFunction:
		case T_Integer:
		case T_Float:
		case T_Boolean:
		case T_String:
		case T_BitString:
		case T_ParamRef:
		case T_A_Const:
		case T_A_Star:
		case T_MergeSupportFunc:
		case T_Alias:
		case T_RangeVar:
		case T_GroupingFunc:
		case T_SubLink:
		case T_CaseExpr:
		case T_RowExpr:
		case T_CoalesceExpr:
		case T_MinMaxExpr:
		case T_XmlExpr:
		case T_JsonReturning:
		case T_JsonValueExpr:
		case T_JsonParseExpr:
		case T_JsonScalarExpr:
		case T_JsonSerializeExpr:
		case T_JsonConstructorExpr:
		case T_JsonIsPredicate:
		case T_JsonArgument:
		case T_JsonFuncExpr:
		case T_JsonBehavior:
		case T_JsonTable:
		case T_JsonTableColumn:
		case T_JsonTablePathSpec:
		case T_NullTest:
		case T_BooleanTest:
		case T_JoinExpr:
		case T_IntoClause:
		case T_List:
		case T_InsertStmt:
		case T_DeleteStmt:
		case T_UpdateStmt:
		case T_MergeStmt:
		case T_MergeWhenClause:
		case T_SelectStmt:
		case T_PLAssignStmt:
		case T_A_Expr:
		case T_BoolExpr:
		case T_ColumnRef:
		case T_FuncCall:
		case T_NamedArgExpr:
		case T_A_Indices:
		case T_A_Indirection:
		case T_A_ArrayExpr:
		case T_ResTarget:
		case T_MultiAssignRef:
		case T_TypeCast:
		case T_CollateClause:
		case T_SortBy:
		case T_WindowDef:
		case T_RangeSubselect:
		case T_RangeFunction:
		case T_RangeTableSample:
		case T_RangeTableFunc:
		case T_RangeTableFuncCol:
		case T_TypeName:
		case T_ColumnDef:
		case T_IndexElem:
		case T_GroupingSet:
		case T_LockingClause:
		case T_XmlSerialize:
		case T_WithClause:
		case T_InferClause:
		case T_OnConflictClause:
		case T_CommonTableExpr:
		case T_JsonOutput:
		case T_JsonKeyValue:
		case T_JsonObjectConstructor:
		case T_JsonArrayConstructor:
		case T_JsonAggConstructor:
		case T_JsonObjectAgg:
		case T_JsonArrayAgg:
		case T_JsonArrayQueryConstructor:
			return true;
		default:
			return false;
	}
}
