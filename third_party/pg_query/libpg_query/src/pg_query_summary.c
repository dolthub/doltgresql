#include "pg_query.h"
#include "pg_query_internal.h"
#include "pg_query_outfuncs.h"

#include "catalog/namespace.h"

#include "parser/parser.h"
#include "parser/scanner.h"
#include "parser/scansup.h"

#include "nodes/makefuncs.h"
#include "nodes/nodeFuncs.h"
#include "utils/builtins.h"

#include "protobuf/pg_query.upb.h"

#include <unistd.h>
#include <fcntl.h>
#include <string.h>

#include "pg_query_summary.h"

static char *range_var_to_name(RangeVar *rv);
static bool cte_names_include(WalkState * state, char *relname);
static void handle_with_clause(WithClause *with_clause, WalkState * state);


static PgQueryError * make_pg_query_error(const char *message)
{
	PgQueryError *error = malloc(sizeof(PgQueryError));

	error->message = strdup(message);
	return error;
}

/*
 * Given a Node (node) and WalkState (state), wraps node in a RangeVarWithContext
 * and adds it to the state->range_vars list.
 */
static void
add_range_var(Node *node, ContextType context, WalkState * state)
{
	if (!node)
		return;

	switch (node->type)
	{
		case T_JoinExpr:
			{
				JoinExpr   *je = castNode(JoinExpr, node);

				add_range_var(je->larg, context, state);
				add_range_var(je->rarg, context, state);
				/* je->quals is handled by the regular tree walk. */
				break;
			}

		case T_RangeTableSample:
			{
				RangeTableSample *rts = castNode(RangeTableSample, node);

				add_range_var(rts->relation, context, state);
				break;
			}

		case T_JsonTable:
		case T_RangeTableFunc:
		case T_RangeFunction:
		case T_RangeSubselect:
			/* These are handled by the tree walker. */
			break;

		case T_RangeVar:
			{
				RangeVarWithContext *rvwc = palloc(sizeof(RangeVarWithContext));

				rvwc->node = castNode(RangeVar, node);
				rvwc->context = context;

				state->range_vars = lappend(state->range_vars, rvwc);
				break;
			}

		default:
			elog(ERROR, "unexpected node type: %d", nodeTag(node));
	}
}

/*
 * Given a List*-of-Nodes (clause) and WalkState (state), wrap them all
 * in RangeVarWithContext structs and add them to the state->range_vars list.
 */
static void
add_range_var_list(List *clause, ContextType context, WalkState * state)
{
	if (!clause)
		return;

	ListCell   *lc = NULL;

	foreach(lc, clause)
	{
		add_range_var(lfirst(lc), context, state);
	}
}

typedef struct
{
	char	   *ctename;
	WalkState  *state;
}			CteSelfReferenceContext;

/*
 * Records the range vars inside a CTE's own definition that carry the CTE's own
 * name.
 */
static bool
cte_self_reference_walker(Node *node, CteSelfReferenceContext * context)
{
	if (node == NULL)
		return false;

	if (IsA(node, RangeVar))
	{
		RangeVar   *rv = castNode(RangeVar, node);

		/* Schema-qualified names never refer to a CTE. */
		if (rv->schemaname == NULL && rv->relname != NULL &&
			strcmp(rv->relname, context->ctename) == 0)
			context->state->cte_self_reference_range_vars =
				lappend(context->state->cte_self_reference_range_vars, rv);
	}

	if (!pg_query_raw_tree_walker_supports(node))
		return false;

	return raw_expression_tree_walker(node, cte_self_reference_walker, (void *) context);
}

/*
 * Records the CTE names a WITH clause defines, plus the range vars inside a
 * CTE's own definition that carry the CTE's own name: a non-recursive CTE is not
 * visible inside its own definition, so those name a real relation, like the
 * first "users" in
 *
 *   WITH users AS (SELECT * FROM users) SELECT * FROM users
 *
 * Runs before descending into the statement, so that the CTE names of enclosing
 * scopes are known when a nested WITH clause is reached.
 */
static void
handle_with_clause(WithClause *with_clause, WalkState * state)
{
	if (with_clause == NULL)
		return;

	ListCell   *lc = NULL;

	foreach(lc, with_clause->ctes)
	{
		if (!IsA(lfirst(lc), CommonTableExpr))
			continue;

		CommonTableExpr *cte = castNode(CommonTableExpr, lfirst(lc));

		if (cte->ctename == NULL)
			continue;

		/*
		 * WITH RECURSIVE does make a CTE visible to itself, and a name that
		 * already is a visible CTE from an enclosing scope keeps resolving to
		 * that CTE inside this definition, like the inner "a" in
		 *
		 *   WITH a AS (...) SELECT * FROM (WITH a AS (SELECT * FROM a) ...) s
		 */
		if (!with_clause->recursive && cte->ctequery != NULL &&
			!cte_names_include(state, cte->ctename))
		{
			CteSelfReferenceContext context = {cte->ctename, state};

			cte_self_reference_walker(cte->ctequery, &context);
		}

		state->cte_names = lappend(state->cte_names, pstrdup(cte->ctename));
	}
}

static void
add_function(List *funcname, ContextType context, WalkState * state)
{
	SummaryFunction *func = palloc(sizeof(SummaryFunction));

	func->name = NameListToString(funcname);

	if (list_length(funcname) == 3)
		func->schema_name = lsecond_node(String, funcname)->sval;
	else if (list_length(funcname) == 2)
		func->schema_name = linitial_node(String, funcname)->sval;
	else
		func->schema_name = NULL;

	func->function_name = llast_node(String, funcname)->sval;

	func->context = context;

	state->functions = lappend(state->functions, func);
}

bool
pg_query_summary_walk_impl(Node *node, WalkState * state)
{
	if (node == NULL)
		return false;

	switch (nodeTag(node))
	{
		case T_CallStmt:
			{
				CallStmt   *stmt = castNode(CallStmt, node);

				if (stmt->funccall)
					return pg_query_summary_walk_impl((Node *) stmt->funccall, state);
			}

		case T_SelectStmt:
			{
				/*
				 * Don't descend into sub-SELECTs in the WHERE clause when
				 * saving filter columns.
				 */
				if (state->save_filter_columns)
					return true;

				SelectStmt *stmt = castNode(SelectStmt, node);

				handle_with_clause(stmt->withClause, state);

				if (stmt->op == SETOP_NONE)
					add_range_var_list(stmt->fromClause, CONTEXT_SELECT, state);

				/*
				 * For SETOP_UNION, SETOP_EXCEPT, and SETOP_INTERSECT, the
				 * raw_expression_tree_walker() call at the end of this
				 * function will handle stmt->larg/stmt->rarg.
				 */

				if (stmt->intoClause)
					add_range_var((Node *) stmt->intoClause->rel, CONTEXT_DDL, state);

				if (stmt->whereClause)
				{
					/*
					 * If we have a WHERE clause, that means we have to check
					 * for filter columns.
					 *
					 * NOTE: the WHERE clause is intentionally walked twice,
					 * but when saving filter columns we don't descend into
					 * sub-SELECTs.
					 */
					state->save_filter_columns = true;
					pg_query_summary_walk_impl((Node *) stmt->whereClause, state);
					state->save_filter_columns = false;
				}

				break;
			}

		case T_InsertStmt:
			{
				InsertStmt *stmt = castNode(InsertStmt, node);

				handle_with_clause(stmt->withClause, state);
				add_range_var((Node *) stmt->relation, CONTEXT_DML, state);
				break;
			}

		case T_UpdateStmt:
			{
				UpdateStmt *stmt = castNode(UpdateStmt, node);

				handle_with_clause(stmt->withClause, state);
				add_range_var((Node *) stmt->relation, CONTEXT_DML, state);
				add_range_var_list(stmt->fromClause, CONTEXT_SELECT, state);
				break;
			}

		case T_MergeStmt:
			{
				MergeStmt  *stmt = castNode(MergeStmt, node);

				handle_with_clause(stmt->withClause, state);
				add_range_var((Node *) stmt->relation, CONTEXT_DML, state);
				break;
			}

		case T_DeleteStmt:
			{
				DeleteStmt *stmt = castNode(DeleteStmt, node);

				handle_with_clause(stmt->withClause, state);
				add_range_var((Node *) stmt->relation, CONTEXT_DML, state);
				add_range_var_list(stmt->usingClause, CONTEXT_SELECT, state);
				break;
			}

		case T_CopyStmt:
			{
				CopyStmt   *stmt = castNode(CopyStmt, node);

				add_range_var((Node *) stmt->relation, CONTEXT_SELECT, state);
				pg_query_summary_walk_impl((Node *) stmt->query, state);
				break;
			}

		case T_AlterTableStmt:
			{
				AlterTableStmt *stmt = castNode(AlterTableStmt, node);

				/* `ALTER INDEX index_name` is ignored. */
				if (stmt->objtype != OBJECT_INDEX)
					add_range_var((Node *) stmt->relation, CONTEXT_DDL, state);

				break;
			}

		case T_CreateStmt:
			add_range_var((Node *) castNode(CreateStmt, node)->relation, CONTEXT_DDL, state);
			break;

		case T_CreateTableAsStmt:
			{
				CreateTableAsStmt *stmt = castNode(CreateTableAsStmt, node);

				if (stmt->into && stmt->into->rel)
					add_range_var((Node *) stmt->into->rel, CONTEXT_DDL, state);

				pg_query_summary_walk_impl(stmt->query, state);

				break;
			}

		case T_TruncateStmt:
			add_range_var_list(castNode(TruncateStmt, node)->relations, CONTEXT_DDL, state);
			break;

		case T_ViewStmt:
			{
				ViewStmt   *stmt = castNode(ViewStmt, node);

				add_range_var((Node *) stmt->view, CONTEXT_DDL, state);
				pg_query_summary_walk_impl(stmt->query, state);
				break;
			}

		case T_IndexStmt:
			{
				IndexStmt  *stmt = castNode(IndexStmt, node);

				add_range_var((Node *) stmt->relation, CONTEXT_DDL, state);

				ListCell   *lc = NULL;

				foreach(lc, stmt->indexParams)
				{
					IndexElem  *index_elem = lfirst(lc);

					pg_query_summary_walk_impl(index_elem->expr, state);
				}

				pg_query_summary_walk_impl(stmt->whereClause, state);

				break;
			}

		case T_CreateTrigStmt:
			add_range_var((Node *) castNode(CreateTrigStmt, node)->relation, CONTEXT_DDL, state);
			break;

		case T_RuleStmt:
			add_range_var((Node *) castNode(RuleStmt, node)->relation, CONTEXT_DDL, state);
			break;

		case T_VacuumStmt:
			{
				VacuumStmt *stmt = castNode(VacuumStmt, node);

				ListCell   *lc = NULL;

				foreach(lc, stmt->rels)
				{
					VacuumRelation *rel = lfirst(lc);

					add_range_var((Node *) rel->relation, CONTEXT_DDL, state);
				}

				break;
			}

		case T_RefreshMatViewStmt:
			add_range_var((Node *) castNode(RefreshMatViewStmt, node)->relation, CONTEXT_DDL, state);
			break;

		case T_DropStmt:
			{
				DropStmt   *stmt = castNode(DropStmt, node);

				switch (stmt->removeType)
				{
					case OBJECT_TABLE:
						{
							ListCell   *lc = NULL;

							foreach(lc, stmt->objects)
							{
								Node	   *obj = lfirst(lc);

								if (obj->type == T_List)
								{
									List	   *list = castNode(List, obj);

									if (list_length(list) == 0)
										continue;

									RangeVar   *range_var = makeRangeVarFromNameList(list);

									add_range_var((Node *) range_var, CONTEXT_DDL, state);
								}
							}
							break;
						}
					case OBJECT_RULE:
					case OBJECT_TRIGGER:
						{
							ListCell   *lc = NULL;

							foreach(lc, stmt->objects)
							{
								Node	   *obj = lfirst(lc);

								if (obj->type == T_List)
								{
									List	   *olist = castNode(List, obj);

									/*
									 * Ignore the last string (the
									 * rule/trigger name)
									 */
									List	   *list = list_copy(olist);

									list = list_truncate(list, list_length(list) - 1);

									if (list_length(list) == 0)
										continue;

									RangeVar   *range_var = makeRangeVarFromNameList(list);

									add_range_var((Node *) range_var, CONTEXT_DDL, state);
								}
							}
							break;
						}

					case OBJECT_FUNCTION:
						{
							Node	   *obj_node = linitial(stmt->objects);

							if (obj_node != NULL && obj_node->type == T_ObjectWithArgs)
							{
								ObjectWithArgs *obj = castNode(ObjectWithArgs, obj_node);

								if (obj->objname)
									add_function(obj->objname, CONTEXT_DDL, state);
							}

							break;
						}

					default:
						break;
				}
				break;
			}

		case T_GrantStmt:
			{
				GrantStmt  *stmt = castNode(GrantStmt, node);
				List	   *objects = stmt->objects;

				switch (stmt->objtype)
				{
					case OBJECT_TABLE:
						{
							ListCell   *lc = NULL;

							foreach(lc, stmt->objects)
							{
								if (IsA(lfirst(lc), RangeVar))
									add_range_var(lfirst(lc), CONTEXT_DDL, state);
							}
							break;
						}
					case OBJECT_FUNCTION:
					case OBJECT_PROCEDURE:
					case OBJECT_ROUTINE:
						{
							ListCell   *lc = NULL;

							foreach(lc, stmt->objects)
							{
								Node	   *obj_node = lfirst(lc);

								if (IsA(obj_node, ObjectWithArgs))
								{
									ObjectWithArgs *obj = castNode(ObjectWithArgs, obj_node);

									if (obj->objname)
										add_function(obj->objname, CONTEXT_DDL, state);
								}
							}
						}
					default:

						/*
						 * Do nothing.
						 *
						 * Other grant types don't contain table/function
						 * references.
						 */
						break;
				}
				break;
			}

		case T_LockStmt:
			add_range_var_list(castNode(LockStmt, node)->relations, CONTEXT_DDL, state);
			break;

		case T_ExplainStmt:
			return pg_query_summary_walk_impl(castNode(ExplainStmt, node)->query, state);

		case T_CreateFunctionStmt:
			add_function(castNode(CreateFunctionStmt, node)->funcname, CONTEXT_DDL, state);
			break;

		case T_RenameStmt:
			{
				RenameStmt *stmt = castNode(RenameStmt, node);

				switch (stmt->renameType)
				{
					case OBJECT_TABLE:
						add_range_var((Node *) stmt->relation, CONTEXT_DDL, state);
						add_range_var((Node *) makeRangeVar(stmt->relation->schemaname, stmt->newname, -1), CONTEXT_DDL, state);
						break;
					case OBJECT_FUNCTION:
					case OBJECT_PROCEDURE:
					case OBJECT_ROUTINE:
						{
							ObjectWithArgs *obj = castNode(ObjectWithArgs, stmt->object);

							add_function(obj->objname, CONTEXT_DDL, state);
							add_function(list_make1(makeString(stmt->newname)), CONTEXT_DDL, state);
							break;
						}
					default:

						/*
						 * Do nothing.
						 *
						 * Other rename types don't contain table/function
						 * references.
						 */
						break;
				}
				break;
			}

		case T_PrepareStmt:
			return pg_query_summary_walk_impl(castNode(PrepareStmt, node)->query, state);

		case T_RawStmt:
			return pg_query_summary_walk_impl(castNode(RawStmt, node)->stmt, state);

		case T_FuncCall:
			add_function(castNode(FuncCall, node)->funcname, CONTEXT_CALL, state);
			break;

			/*
			 * The following are added to avoid needing subselect_items,
			 *
			 * like is used in Rust and Ruby.
			 */
		case T_ColumnRef:
			{
				if (!state->save_filter_columns)
					break;

				ColumnRef  *ref = castNode(ColumnRef, node);

				Node	   *column_node = llast(ref->fields);

				/*
				 * Check that we are not in the A_Star case here. (Columns can
				 * be stars, but nothing else can.)
				 */
				if (!IsA(column_node, String))
					break;

				String	   *schema_name = NULL;
				String	   *table_name = NULL;
				String	   *column = castNode(String, column_node);

				size_t		num_fields = list_length(ref->fields);

				if (num_fields == 2)
				{
					table_name = linitial(ref->fields);
				}
				else if (num_fields == 3)
				{
					schema_name = linitial(ref->fields);
					table_name = lsecond_node(String, ref->fields);
				}

				FilterColumn *fc = palloc(sizeof(FilterColumn));

				fc->schema_name = schema_name ? schema_name->sval : NULL;
				fc->table_name = table_name ? table_name->sval : NULL;
				fc->column = column ? column->sval : NULL;

				state->filter_columns = lappend(state->filter_columns, fc);

				break;
			}

		default:
			break;
	}

	if (!pg_query_raw_tree_walker_supports(node))
		return false;

	return raw_expression_tree_walker(node, pg_query_summary_walk_impl, (void *) state);
}

static char *
range_var_to_name(RangeVar *rv)
{
	StringInfoData buf;

	initStringInfo(&buf);
	if (rv->schemaname != NULL)
	{
		appendStringInfoString(&buf, quote_identifier(rv->schemaname));
		appendStringInfoChar(&buf, '.');
	}
	appendStringInfoString(&buf, quote_identifier(rv->relname));
	return buf.data;
}

/*
 * Walk through all collected cte_names, and compares them to relname.
 *
 * Returns true if any are an exact match, false otherwise.
 */
static bool
cte_names_include(WalkState * state, char *relname)
{
	if (relname == NULL)
		return false;

	ListCell   *lc = NULL;

	foreach(lc, state->cte_names)
	{
		char	   *ctename = lfirst(lc);

		if (strcmp(ctename, relname) == 0)
			return true;
	}
	return false;
}

/*
 * Returns true if the given range var was recorded as a reference to the CTE it
 * is defined in - see handle_with_clause().
 */
static bool
cte_self_reference_range_vars_include(WalkState * state, RangeVar *rv)
{
	ListCell   *lc = NULL;

	foreach(lc, state->cte_self_reference_range_vars)
	{
		if (lfirst(lc) == rv)
			return true;
	}
	return false;
}

/*
 * Determines whether a range var refers to a CTE rather than to a relation.
 *
 * Only plain (SELECT-style) references can resolve to a CTE: a CTE is not a
 * valid target for DML (INSERT/UPDATE/DELETE/MERGE) or DDL, so those always name
 * a real relation, even when a CTE in the same statement shares the name.
 */
static bool
is_cte_reference(WalkState * state, RangeVar *rv, ContextType context)
{
	if (context != CONTEXT_SELECT)
		return false;

	/* Schema-qualified names always refer to a relation. */
	if (rv->schemaname != NULL)
		return false;

	if (!cte_names_include(state, rv->relname))
		return false;

	return !cte_self_reference_range_vars_include(state, rv);
}

static void
handle_range_var(RangeVar *node, ContextType context, WalkState * state)
{
	if (node == NULL)
		return;

	RangeVar   *rv = node;

	if (!is_cte_reference(state, rv, context))
	{
		SummaryTable *table = palloc(sizeof(SummaryTable));

		table->name = range_var_to_name(rv);
		table->schema_name = pstrdup(rv->schemaname ? rv->schemaname : "");
		table->table_name = pstrdup(rv->relname ? rv->relname : "");
		table->context = context;

		state->tables = lappend(state->tables, table);
	}

	if (rv->alias)
	{
		SummaryAlias *alias = palloc(sizeof(SummaryAlias));

		alias->key = rv->alias->aliasname;
		alias->value = range_var_to_name(rv);
		state->aliases = lappend(state->aliases, alias);
	}
}

/* Walk the parse tree, storing the results in `summary`. */
static void
pg_query_summary_walk(Summary * summary, Node *tree)
{
	WalkState	state = {0};

	/*
	 * NOTE regarding cte_names and range_vars:
	 * - We process cte_names as we iterate through the tree, from the WITH
	 *   clause of a statement, before descending into that statement.
	 * - We only add items to tables if they are *not* a CTE reference.
	 * - CTEs can be defined *after* they're referenced as names, since
	 *   raw_expression_tree_walker() walks the WITH clause of a statement last.
	 * - Due to this, we store range_vars and process it after the tree walk.
	 *
	 * If we try to do it entirely on the initial pass, we wind up with CTEs
	 * being included in state->tables, and thus in summary->tables.
	 */

	pg_query_summary_walk_impl(tree, &state);

	ListCell   *lc = NULL;

	foreach(lc, state.range_vars)
	{
		RangeVarWithContext *rvwc = lfirst(lc);

		handle_range_var(rvwc->node, rvwc->context, &state);

		pfree(rvwc);
	}

	if (state.tables != NULL)
		summary->tables = state.tables;

	if (state.aliases != NULL)
		summary->aliases = state.aliases;

	if (state.cte_names != NULL)
		summary->cte_names = state.cte_names;

	if (state.functions != NULL)
		summary->functions = state.functions;

	if (state.filter_columns != NULL)
		summary->filter_columns = state.filter_columns;
}

static int32_t ctx_to_protobuf_ctx(ContextType ctx)
{
	/*
	 * We're converting from an enum we control (ContextType) to the enum
	 * generated for the protobuf SummaryResult.Context.
	 *
	 * While these should always be equivalent, we can't guarantee correctness
	 * with a simple cast. (It is theoretically possible for the internal
	 * representations to diverge while being functionally equivalent.)
	 *
	 * So, instead, we do a switch/case and manually convert it.
	 */
	switch (ctx)
	{
		case CONTEXT_NONE:
			return pg_query_SummaryResult_None;
		case CONTEXT_SELECT:
			return pg_query_SummaryResult_Select;
		case CONTEXT_DML:
			return pg_query_SummaryResult_DML;
		case CONTEXT_DDL:
			return pg_query_SummaryResult_DDL;
		case CONTEXT_CALL:
			return pg_query_SummaryResult_Call;
		default:
			return pg_query_SummaryResult_None;
	}
}

/* upb string fields hold a (pointer, length) view; the source strings here are
 * palloc'd and live until after we serialize, so we can reference them directly
 * (NULL becomes the empty string, matching the previous protobuf-c behavior). */
static inline upb_StringView sv(const char *s)
{
	return upb_StringView_FromString(s != NULL ? s : "");
}

/* Given the summarized query, convert it to protobuf, and store it in `result`. */
static void
summary_to_protobuf(PgQuerySummaryParseResult * result, Summary * summary)
{
	upb_Arena  *a = upb_Arena_New();
	pg_query_SummaryResult *sr = pg_query_SummaryResult_new(a);
	ListCell   *lc = NULL;
	char	   *data;
	size_t		len;

	foreach(lc, summary->tables)
	{
		SummaryTable *table = lfirst(lc);
		pg_query_SummaryResult_Table *t = pg_query_SummaryResult_add_tables(sr, a);

		pg_query_SummaryResult_Table_set_name(t, sv(table->name));
		pg_query_SummaryResult_Table_set_schema_name(t, sv(table->schema_name));
		pg_query_SummaryResult_Table_set_table_name(t, sv(table->table_name));
		pg_query_SummaryResult_Table_set_context(t, ctx_to_protobuf_ctx(table->context));
	}

	foreach(lc, summary->aliases)
	{
		SummaryAlias *alias = lfirst(lc);

		pg_query_SummaryResult_aliases_set(sr, sv(alias->key), sv(alias->value), a);
	}

	foreach(lc, summary->cte_names)
		pg_query_SummaryResult_add_cte_names(sr, sv(lfirst(lc)), a);

	foreach(lc, summary->functions)
	{
		SummaryFunction *fn = lfirst(lc);
		pg_query_SummaryResult_Function *f = pg_query_SummaryResult_add_functions(sr, a);

		pg_query_SummaryResult_Function_set_name(f, sv(fn->name));
		pg_query_SummaryResult_Function_set_function_name(f, sv(fn->function_name));
		pg_query_SummaryResult_Function_set_schema_name(f, sv(fn->schema_name));
		pg_query_SummaryResult_Function_set_context(f, ctx_to_protobuf_ctx(fn->context));
	}

	foreach(lc, summary->filter_columns)
	{
		FilterColumn *fc = lfirst(lc);
		pg_query_SummaryResult_FilterColumn *c = pg_query_SummaryResult_add_filter_columns(sr, a);

		pg_query_SummaryResult_FilterColumn_set_schema_name(c, sv(fc->schema_name));
		pg_query_SummaryResult_FilterColumn_set_table_name(c, sv(fc->table_name));
		pg_query_SummaryResult_FilterColumn_set_column(c, sv(fc->column));
	}

	foreach(lc, summary->statement_types)
		pg_query_SummaryResult_add_statement_types(sr, sv(lfirst(lc)), a);

	pg_query_SummaryResult_set_truncated_query(sr, sv(summary->truncated_query));

	/*
	 * Serialization fails (returns NULL) e.g. when the arena runs out of
	 * memory. The arena is malloc-backed, so free it before throwing.
	 */
	data = pg_query_SummaryResult_serialize(sr, a, &len);
	if (data == NULL)
	{
		upb_Arena_Free(a);
		elog(ERROR, "could not serialize summary to protobuf");
	}

	/*
	 * Note: This is intentionally malloc so exiting the memory context
	 * doesn't free this
	 */
	result->summary.len = len;
	result->summary.data = malloc(len);
	memcpy(result->summary.data, data, len);

	upb_Arena_Free(a);
}

/*
 * Given a query, parser options, and truncate limit, provide a summary of the query.
 *
 * If `truncate_limit` is -1, no truncation is performed.
 */
PgQuerySummaryParseResultInternal
pg_query_summary_internal(const char *input, int parser_options, int truncate_limit)
{
	PgQueryInternalParsetreeAndError parsetree_and_error;
	PgQuerySummaryParseResultInternal result = {0};
	Summary		summary = {0};
	PgQueryError *error = NULL;
	bool		should_truncate = (truncate_limit != -1);

	parsetree_and_error = pg_query_raw_parse(input, parser_options);
	result.stderr_buffer = parsetree_and_error.stderr_buffer;
	result.error = parsetree_and_error.error;

	if (result.error == NULL)
	{
		/*
		 * The walks may throw (e.g. "stack depth limit exceeded"), and the
		 * malloc-ed stderr buffer is only known to this frame at that point, so
		 * free it before passing on the error.
		 */
		PG_TRY();
		{
			pg_query_summary_walk(&summary, (Node *) parsetree_and_error.tree);
			pg_query_summary_statement_walk(&summary, (Node *) parsetree_and_error.tree);

			if (should_truncate)
				pg_query_summary_truncate(&summary, (Node *) parsetree_and_error.tree, truncate_limit);
		}
		PG_CATCH();
		{
			if (result.stderr_buffer)
				free(result.stderr_buffer);
			PG_RE_THROW();
		}
		PG_END_TRY();
	}

	result.summary = summary;

	return result;
}

PgQuerySummaryParseResult
pg_query_summary(const char *input, int parser_options, int truncate_limit)
{
	PgQuerySummaryParseResult result = {0};
	MemoryContext ctx = pg_query_enter_memory_context();

	PG_TRY();
	{
		PgQuerySummaryParseResultInternal internal_result =
			pg_query_summary_internal(input, parser_options, truncate_limit);

		result.stderr_buffer = internal_result.stderr_buffer;
		result.error = internal_result.error;

		if (result.error == NULL)
			summary_to_protobuf(&result, &internal_result.summary);
	}
	PG_CATCH();
	{
		ErrorData  *error_data;
		PgQueryError *error;

		MemoryContextSwitchTo(ctx);
		error_data = CopyErrorData();

		/*
		 * Note: This is intentionally malloc so exiting the memory context
		 * doesn't free this
		 */
		error = malloc(sizeof(PgQueryError));
		error->message = strdup(error_data->message);
		error->filename = strdup(error_data->filename);
		error->funcname = strdup(error_data->funcname);
		error->context = NULL;
		error->lineno = error_data->lineno;
		error->cursorpos = error_data->cursorpos;

		result.error = error;
		FlushErrorState();
	}
	PG_END_TRY();

	pg_query_exit_memory_context(ctx);

	return result;
}

/* Free all of the various parts of PgQuerySummaryParseResultInternal. */
void
pg_query_free_summary_parse_result_internal(PgQuerySummaryParseResultInternal result)
{
	if (result.error)
		pg_query_free_error(result.error);

	if (result.stderr_buffer)
		free(result.stderr_buffer);
}

/* Free all of the various parts of a PgQuerySummaryParseResult. */
void
pg_query_free_summary_parse_result(PgQuerySummaryParseResult result)
{
	if (result.error)
		pg_query_free_error(result.error);

	if (result.summary.data)
		free(result.summary.data);

	/* This is allocated by strdup(), so it uses malloc() instead of palloc(). */
	if (result.stderr_buffer)
		free(result.stderr_buffer);
}
