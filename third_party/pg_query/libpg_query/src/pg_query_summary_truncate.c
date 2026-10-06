#include "pg_query.h"
#include "pg_query_internal.h"
#include "pg_query_summary.h"

#include "nodes/pg_list.h"
#include "nodes/nodeFuncs.h"

#include "mb/pg_wchar.h"

enum TruncationAttr
{
	TRUNCATION_TARGET_LIST,
	TRUNCATION_WHERE_CLAUSE,
	TRUNCATION_VALUES_LISTS,
	TRUNCATION_COLS,
	TRUNCATION_CTE_QUERY,
};

typedef struct
{
	List	   *truncations;
	int32_t		depth;
}			TruncationState;

typedef struct
{
	enum TruncationAttr attr;
	Node	   *node;
	int32_t		depth;
	int32_t		length;
}			PossibleTruncation;

static bool generate_possible_truncations(Node *tree, TruncationState * state);
static void apply_truncations(Summary * summary, Node *tree, TruncationState * state, int truncate_limit);
static int	cmp_possible_truncations(const ListCell *a, const ListCell *b);

static int32_t select_target_list_len(List *nodes);
static int32_t select_values_lists_len(List *nodes);
static int32_t update_target_list_len(List *nodes);
static int32_t where_clause_len(Node *node);
static int32_t cols_len(List *nodes);

static ColumnRef *dummy_column(void);
static ResTarget *dummy_target(void);
static Node *dummy_select(List *targetList, Node *whereClause, List *valuesLists);
static Node *dummy_insert(List *cols);
static Node *dummy_update(List *targetList);

static size_t deparse_stmt_len(Node *node);
static StringInfo deparse_raw_stmt_list(List *stmts);

/*  Given a walked parse tree and summary, store the truncated version in `summary`. */
void
pg_query_summary_truncate(Summary * summary, Node *tree, int truncate_limit)
{
	TruncationState state = {NULL, 0};

	StringInfo	output = deparse_raw_stmt_list((List *) tree);

	if (output->len <= truncate_limit)
	{
		summary->truncated_query = output->data;
		return;
	}

	destroyStringInfo(output);

	generate_possible_truncations(tree, &state);

	list_sort(state.truncations, cmp_possible_truncations);
	apply_truncations(summary, tree, &state, truncate_limit);
}

static void
truncate_mbstr(char *mbstr, size_t max_chars)
{
	/* Determine the number of characters in mbstr. */
	int			n_chars = pg_mbstrlen(mbstr);

	/* If we don't need to truncate the string, return immediately. */
	if (n_chars <= max_chars)
		return;

	/* Determine how many bytes hold `max_chars - 3`. */
	int			n_bytes = pg_mbcharcliplen(mbstr, n_chars, max_chars - 3);

	/* Actually truncate it. */
	strncpy(mbstr + n_bytes, "...", 4);
	mbstr[n_bytes + 3] = '\0';
}

static void
add_truncation(TruncationState * state, enum TruncationAttr attr,
			   Node *node, int32_t length)
{
	/* Don't bother truncating if it won't become shorter. */
	if (length <= 3)
		return;

	PossibleTruncation *truncation = palloc(sizeof(PossibleTruncation));

	truncation->attr = attr;
	truncation->node = node;
	truncation->depth = state->depth;
	truncation->length = length;

	state->truncations = lappend(state->truncations, truncation);
}

static void
add_truncation_where_clause(TruncationState * state, Node *node, Node *whereClause)
{
	if (whereClause == NULL)
		return;

	add_truncation(state,
				   TRUNCATION_WHERE_CLAUSE,
				   node,
				   where_clause_len(whereClause));
}

static bool
generate_possible_truncations(Node *node, TruncationState * state)
{
	if (node == NULL)
		return false;

	switch (nodeTag(node))
	{
		case T_RawStmt:
			return generate_possible_truncations(castNode(RawStmt, node)->stmt, state);

		case T_SelectStmt:
			{
				SelectStmt *stmt = castNode(SelectStmt, node);

				if (stmt->targetList != NULL)
					add_truncation(state,
								   TRUNCATION_TARGET_LIST,
								   node,
								   select_target_list_len(stmt->targetList));

				add_truncation_where_clause(state, node, stmt->whereClause);

				if (stmt->valuesLists != NULL)
					add_truncation(state,
								   TRUNCATION_VALUES_LISTS,
								   node,
								   select_values_lists_len(stmt->valuesLists));

				break;
			}

		case T_InsertStmt:
			{
				InsertStmt *stmt = castNode(InsertStmt, node);

				if (stmt->cols != NULL)
					add_truncation(state, TRUNCATION_COLS, node, cols_len(stmt->cols));

				break;
			}

		case T_UpdateStmt:
			{
				UpdateStmt *stmt = castNode(UpdateStmt, node);

				if (stmt->targetList != NULL)
					add_truncation(state,
								   TRUNCATION_TARGET_LIST,
								   node,
								   update_target_list_len(stmt->targetList));

				add_truncation_where_clause(state, node, stmt->whereClause);

				break;
			}

		case T_DeleteStmt:
			{
				DeleteStmt *stmt = castNode(DeleteStmt, node);

				add_truncation_where_clause(state, node, stmt->whereClause);
				break;
			}

		case T_CopyStmt:
			{
				CopyStmt   *stmt = castNode(CopyStmt, node);

				add_truncation_where_clause(state, node, stmt->whereClause);
				break;
			}

		case T_IndexStmt:
			{
				IndexStmt  *stmt = castNode(IndexStmt, node);

				add_truncation_where_clause(state, node, stmt->whereClause);
				break;
			}

		case T_RuleStmt:
			{
				RuleStmt   *stmt = castNode(RuleStmt, node);

				add_truncation_where_clause(state, node, stmt->whereClause);
				break;
			}

		case T_CommonTableExpr:
			{
				CommonTableExpr *stmt = castNode(CommonTableExpr, node);

				if (stmt->ctequery != NULL)
				{
					size_t		query_len = deparse_stmt_len((Node *) stmt->ctequery);

					add_truncation(state,
								   TRUNCATION_CTE_QUERY,
								   node,
								   query_len);
				}

				break;
			}

		case T_InferClause:
			{
				InferClause *stmt = castNode(InferClause, node);

				add_truncation_where_clause(state, node, stmt->whereClause);
				break;
			}

		case T_OnConflictClause:
			{
				OnConflictClause *stmt = castNode(OnConflictClause, node);

				if (stmt->targetList != NULL)
					add_truncation(state,
								   TRUNCATION_TARGET_LIST,
								   node,
								   update_target_list_len(stmt->targetList));

				add_truncation_where_clause(state, node, stmt->whereClause);

				break;
			}

		default:
			break;
	}

	if (!pg_query_raw_tree_walker_supports(node))
		return false;

	int			old_depth = state->depth;

	state->depth++;

	bool		result = raw_expression_tree_walker(node, generate_possible_truncations, (void *) state);

	/* Restore old depth value, since the current node (or its parents) may */
	/* have sibling elements. */
	state->depth = old_depth;

	return result;
}

static int
cmp_possible_truncations(const ListCell *a, const ListCell *b)
{
	const		PossibleTruncation *pt_a = lfirst(a);
	const		PossibleTruncation *pt_b = lfirst(b);

	int			depth_cmp = pt_b->depth - pt_a->depth;

	if (depth_cmp != 0)
		return depth_cmp;

	return pt_b->length - pt_a->length;
}

static void
global_replace(char *str, char *pattern, char *replacement)
{
	size_t		plen = strlen(pattern);
	size_t		rlen = strlen(replacement);
	char	   *s = str;

	Assert(plen >= rlen);

	while ((s = strstr(s, pattern)) != NULL)
	{
		memcpy(s, replacement, rlen);

		/* Shift remainder of the string if needed */
		if (plen > rlen)
			memmove(s + rlen, s + plen, strlen(s + plen) + 1);

		s += rlen;
	}
}

static void
apply_truncations(Summary * summary, Node *tree, TruncationState * state, int truncation_limit)
{
	List	   *truncations = state->truncations;
	StringInfo	output = NULL;
	ListCell   *lc;

	foreach(lc, state->truncations)
	{
		PossibleTruncation *truncation = lfirst(lc);

		Node	   *node = truncation->node;
		enum TruncationAttr attr = truncation->attr;

		if (IsA(node, SelectStmt) && attr == TRUNCATION_TARGET_LIST)
			castNode(SelectStmt, node)->targetList = list_make1(dummy_target());
		else if (IsA(node, SelectStmt) && attr == TRUNCATION_WHERE_CLAUSE)
			castNode(SelectStmt, node)->whereClause = (Node *) dummy_column();
		else if (IsA(node, SelectStmt) && attr == TRUNCATION_VALUES_LISTS)
			castNode(SelectStmt, node)->valuesLists = list_make1(list_make1(dummy_column()));
		else if (IsA(node, UpdateStmt) && attr == TRUNCATION_TARGET_LIST)
			castNode(UpdateStmt, node)->targetList = list_make1(dummy_target());
		else if (IsA(node, InsertStmt) && attr == TRUNCATION_COLS)
			castNode(InsertStmt, node)->cols = list_make1(dummy_target());
		else if (IsA(node, UpdateStmt) && attr == TRUNCATION_WHERE_CLAUSE)
			castNode(UpdateStmt, node)->whereClause = (Node *) dummy_column();
		else if (IsA(node, DeleteStmt) && attr == TRUNCATION_WHERE_CLAUSE)
			castNode(DeleteStmt, node)->whereClause = (Node *) dummy_column();
		else if (IsA(node, CopyStmt) && attr == TRUNCATION_WHERE_CLAUSE)
			castNode(CopyStmt, node)->whereClause = (Node *) dummy_column();
		else if (IsA(node, IndexStmt) && attr == TRUNCATION_WHERE_CLAUSE)
			castNode(IndexStmt, node)->whereClause = (Node *) dummy_column();
		else if (IsA(node, RuleStmt) && attr == TRUNCATION_WHERE_CLAUSE)
			castNode(RuleStmt, node)->whereClause = (Node *) dummy_column();
		else if (IsA(node, CommonTableExpr) && attr == TRUNCATION_CTE_QUERY)
			castNode(CommonTableExpr, node)->ctequery = dummy_select(NULL, (Node *) dummy_column(), NULL);
		else if (IsA(node, InferClause) && attr == TRUNCATION_WHERE_CLAUSE)
			castNode(InferClause, node)->whereClause = (Node *) dummy_column();
		else if (IsA(node, OnConflictClause) && attr == TRUNCATION_TARGET_LIST)
			castNode(OnConflictClause, node)->targetList = list_make1(dummy_target());
		else if (IsA(node, OnConflictClause) && attr == TRUNCATION_WHERE_CLAUSE)
			castNode(OnConflictClause, node)->whereClause = (Node *) dummy_column();
		else
			elog(ERROR, "apply_truncations() got unknown truncation type");

		if (output)
			destroyStringInfo(output);
		output = deparse_raw_stmt_list((List *) tree);

		global_replace(output->data, "SELECT \"…\" AS \"…\"", "SELECT \"…\"");
		global_replace(output->data, "SELECT WHERE \"…\"", "\"…\"");
		global_replace(output->data, "\"…\"", "...");

		if (strlen(output->data) <= truncation_limit)
		{
			summary->truncated_query = output->data;
			return;
		}
	}

	if (!output)
		output = deparse_raw_stmt_list((List *) tree);

	truncate_mbstr(output->data, truncation_limit);
	summary->truncated_query = output->data;
}

static ColumnRef *
dummy_column(void)
{
	ColumnRef  *colref = makeNode(ColumnRef);

	colref->fields = list_make1(makeString(pstrdup("…")));
	colref->location = 0;

	return colref;
}

static ResTarget *
dummy_target(void)
{
	ResTarget  *target = makeNode(ResTarget);

	target->name = pstrdup("…");
	target->val = (Node *) dummy_column();

	/*
	 * TODO:docs for ResTarget say "-1 if unknown"-- would that be more
	 * correct ? (see also, dummy_column())
	 */
	target->location = 0;

	return target;
}

static Node *
dummy_select(List *targetList, Node *whereClause, List *valuesLists)
{
	SelectStmt *stmt = makeNode(SelectStmt);

	stmt->targetList = targetList;
	stmt->whereClause = whereClause;
	stmt->valuesLists = valuesLists;

	return (Node *) stmt;
}

static Node *
dummy_insert(List *cols)
{
	RangeVar   *rv = makeNode(RangeVar);

	rv->relname = pstrdup("x");
	rv->inh = true;
	rv->relpersistence = 'p';
	rv->location = 0;

	InsertStmt *stmt = makeNode(InsertStmt);

	stmt->relation = rv;
	stmt->cols = cols;
	stmt->override = 1;

	return (Node *) stmt;
}

static Node *
dummy_update(List *targetList)
{
	RangeVar   *rv = makeNode(RangeVar);

	rv->relname = pstrdup("x");
	rv->inh = true;
	rv->relpersistence = 'p';
	rv->location = 0;

	UpdateStmt *stmt = makeNode(UpdateStmt);

	stmt->relation = rv;
	stmt->targetList = targetList;

	return (Node *) stmt;
}

static int32_t
select_target_list_len(List *nodes)
{
	size_t		result_len = deparse_stmt_len(dummy_select(nodes, NULL, NULL));
	size_t		dummy_len = 7;	/* "SELECT " */

	return (int32_t) result_len - dummy_len;
}

static int32_t
select_values_lists_len(List *nodes)
{
	size_t		result_len = deparse_stmt_len(dummy_select(NULL, NULL, nodes));
	size_t		dummy_len = 9;	/* "VALUES ()" */

	return (int32_t) result_len - dummy_len;
}

static int32_t
update_target_list_len(List *nodes)
{
	size_t		result_len = deparse_stmt_len(dummy_update(nodes));
	size_t		dummy_len = 13; /* "UPDATE x SET " */

	return (int32_t) result_len - dummy_len;
}

static int32_t
where_clause_len(Node *node)
{
	size_t		result_len = deparse_stmt_len(dummy_select(NULL, node, NULL));
	size_t		dummy_len = 13; /* "SELECT WHERE " */

	return (int32_t) result_len - dummy_len;
}

static int32_t
cols_len(List *nodes)
{
	size_t		result_len = deparse_stmt_len(dummy_insert(nodes));
	size_t		dummy_len = 31; /* "INSERT INTO x () DEFAULT VALUES" */

	return (int32_t) result_len - dummy_len;
}

static size_t
deparse_stmt_len(Node *node)
{
	if (!IsA(node, RawStmt))
	{
		RawStmt    *raw = makeNode(RawStmt);

		raw->stmt = node;
		raw->stmt_location = -1;
		raw->stmt_len = 0;
		node = (Node *) raw;
	}

	StringInfo	str = deparse_raw_stmt_list(list_make1(node));
	size_t		len = str->len;

	destroyStringInfo(str);
	return len;
}

static StringInfo
deparse_raw_stmt_list(List *stmts)
{
	PostgresDeparseOpts opts = {0};
	StringInfo	str = makeStringInfo();
	ListCell   *lc;

	foreach(lc, stmts)
	{
		deparseRawStmtOpts(str, castNode(RawStmt, lfirst(lc)), opts);
		if (lnext(stmts, lc))
			appendStringInfoString(str, "; ");
	}

	return str;
}
