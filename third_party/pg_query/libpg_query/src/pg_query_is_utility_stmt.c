#include <stdbool.h>
#include <stddef.h>
#include <stdlib.h>

#include "pg_query.h"
#include "pg_query_internal.h"
#include <nodes/nodes.h>
#include <nodes/nodeFuncs.h>

static
bool
is_utility_stmt_actual(RawStmt *raw_stmt)
{
	switch (nodeTag(raw_stmt->stmt))
	{
		case T_SelectStmt:
		case T_InsertStmt:
		case T_UpdateStmt:
		case T_DeleteStmt:
		case T_MergeStmt:
			return false;

		default:
			return true;
	}
}

PgQueryIsUtilityResult
pg_query_is_utility_stmt(const char *query)
{
	PgQueryIsUtilityResult result = {0};
	MemoryContext ctx = pg_query_enter_memory_context();

	PgQueryInternalParsetreeAndError parsetree_and_error = pg_query_raw_parse(query, 0);

	if (parsetree_and_error.error)
	{
		result.error = parsetree_and_error.error;
	}
	else
	{
		ListCell   *lc;

		result.length = list_length(parsetree_and_error.tree);
		result.items = malloc(sizeof(bool) * result.length);

		foreach(lc, parsetree_and_error.tree)
		{
			RawStmt    *raw_stmt = lfirst_node(RawStmt, lc);

			result.items[foreach_current_index(lc)] = is_utility_stmt_actual(raw_stmt);
		}
	}

	if (parsetree_and_error.stderr_buffer)
		free(parsetree_and_error.stderr_buffer);

	pg_query_exit_memory_context(ctx);

	return result;
}

void
pg_query_free_is_utility_result(PgQueryIsUtilityResult result)
{
	if (result.error)
		pg_query_free_error(result.error);

	free(result.items);
}
