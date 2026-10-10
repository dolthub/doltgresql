#include "postgres_deparse.h"

#include "postgres.h"
#include "catalog/index.h"
#include "catalog/pg_am.h"
#include "catalog/pg_attribute.h"
#include "catalog/pg_class.h"
#include "catalog/pg_trigger.h"
#include "commands/trigger.h"
#include "common/keywords.h"
#include "common/kwlookup.h"
#include "lib/stringinfo.h"
#include "nodes/nodes.h"
#include "nodes/parsenodes.h"
#include "nodes/pg_list.h"
#include "utils/builtins.h"
#include "utils/datetime.h"
#include "utils/timestamp.h"
#include "utils/xml.h"
#include "miscadmin.h"

/*
 * Strict structural checking of the input tree (on by default).
 *
 * The deparser receives its parse tree from the caller (in libpg_query, from
 * an arbitrary protobuf), so the tree may not have the shape a given deparse
 * path expects. The stock code documents those expectations with Assert(),
 * which is a no-op in a release build, so a malformed tree instead triggers
 * out-of-bounds reads that can leak adjacent heap memory or crash. Two cases:
 *
 *   - A node's actual type not matching what a path expects. The stock
 *     castNode() is a bare cast with no runtime check, so intVal()/strVal()/
 *     boolVal()/floatVal() read the wrong union member.
 *
 *   - A list being shorter than a path assumes. The stock linitial()/lsecond()/
 *     lthird()/lfourth()/llast() index past the end of the ListCell array.
 *
 * By default we redefine those macros here to check and raise a normal error
 * on violation, caught by libpg_query's surrounding PG_TRY and returned as a
 * PgQueryError. Redefining castNode() covers intVal()/strVal()/etc., since
 * they are all defined in terms of it; redefining the positional list
 * accessors covers every fixed-index access.
 *
 * Define PG_QUERY_DEPARSE_NO_STRICT_CHECKS to opt out and keep the stock
 * unchecked macros -- e.g. if you feed the deparser only trees you have
 * already validated and want to avoid the per-access checks.
 */
#ifndef PG_QUERY_DEPARSE_NO_STRICT_CHECKS
static inline void *
pg_query_strict_cast_node(NodeTag type, void *ptr, const char *file, int line)
{
	if (ptr != NULL && nodeTag(ptr) != type)
		elog(ERROR, "deparse: unexpected node type %d, expected %d (%s:%d)",
			 (int) nodeTag(ptr), (int) type, file, line);
	return ptr;
}
#undef castNode
#define castNode(_type_, nodeptr) \
	((_type_ *) pg_query_strict_cast_node(T_##_type_, (nodeptr), __FILE__, __LINE__))

/*
 * intVal()/strVal()/boolVal()/floatVal() extract a scalar from a value node, so
 * a NULL node is always a bug (a genuinely-absent value is handled by the
 * caller before extraction, e.g. "PASSWORD NULL"). Reject NULL here rather than
 * dereferencing it. castNode()'s own check then covers a wrong node type.
 */
static inline void *
pg_query_require_value_node(void *ptr, const char *file, int line)
{
	if (ptr == NULL)
		elog(ERROR, "deparse: value extraction from NULL node (%s:%d)", file, line);
	return ptr;
}
#undef intVal
#undef floatVal
#undef boolVal
#undef strVal
#define intVal(v)	(castNode(Integer, pg_query_require_value_node((v), __FILE__, __LINE__))->ival)
#define floatVal(v)	atof(castNode(Float, pg_query_require_value_node((v), __FILE__, __LINE__))->fval)
#define boolVal(v)	(castNode(Boolean, pg_query_require_value_node((v), __FILE__, __LINE__))->boolval)
#define strVal(v)	(castNode(String, pg_query_require_value_node((v), __FILE__, __LINE__))->sval)

static inline void *
pg_query_checked_list_nth(const List *l, int n, const char *file, int line)
{
	if (l == NIL || n < 0 || n >= list_length(l))
		elog(ERROR, "deparse: list index %d out of range for length %d (%s:%d)",
			 n, list_length(l), file, line);
	return list_nth(l, n);
}

static inline void *
pg_query_checked_list_last(const List *l, const char *file, int line)
{
	if (l == NIL)
		elog(ERROR, "deparse: llast() on empty list (%s:%d)", file, line);
	return list_nth(l, list_length(l) - 1);
}

#undef linitial
#undef lsecond
#undef lthird
#undef lfourth
#undef llast
#define linitial(l)	pg_query_checked_list_nth((l), 0, __FILE__, __LINE__)
#define lsecond(l)	pg_query_checked_list_nth((l), 1, __FILE__, __LINE__)
#define lthird(l)	pg_query_checked_list_nth((l), 2, __FILE__, __LINE__)
#define lfourth(l)	pg_query_checked_list_nth((l), 3, __FILE__, __LINE__)
#define llast(l)	pg_query_checked_list_last((l), __FILE__, __LINE__)
#endif

/*
 * # Deparser overview
 *
 * The deparser works by walking the input parse tree and emitting into the
 * currently active "part" as pointed to (indirectly) by the state struct.
 *
 * A lot of the structure of the deparser described below is meant to support
 * the optional "pretty print" mode that can be enabled and configured through
 * the deparser options, and inserts whitespace (newlines/spaces) as needed to
 * make the output easier to read.
 *
 * ## Nesting levels
 *
 * Starting at the state struct we have a currently active "nesting level",
 * which indicates the base indentation, as well as contains one or more part
 * groups, each of which has one or more parts.
 *
 * Nesting levels are typically opened, or "increased", when processing a new
 * statement contained within other statements, or for certain statement-like
 * constructs (e.g. CASE clauses) that require custom indendation.
 *
 * Nesting levels are closed, or "decreased", when exiting such statements or
 * statement-like constructs. During this operation all parts and subparts get
 * flattened (turned into a simple list of parts), and are added to the parent
 * nesting level.
 *
 * At the end of the tree walk the parts of the top most nesting level get
 * emitted into the output string, formatted based on the deparse options.
 *
 * ## Part groups
 *
 * Part groups are usually started for each "major" keyword, which is chosen
 * for how the output is intended to be laid out. For example, the "FROM"
 * keyword in a SELECT statement is considered "major" and starts a new part
 * group, vs the "JOIN" keyword is not, and as such both regular tables and
 * JOIN clauses will be within one part group.
 *
 * ## Parts, indentation and merging
 *
 * The parts contained in part groups may be indented one additional level
 * beyond the base indentation of the currently active nesting level, to help
 * differentiate them from the major keywords. This can be turned off with
 * DEPARSE_PART_NO_INDENT, for example as needed to format WITH clauses
 * effectively.
 *
 * Parts may be merged with other parts within the same group, if the indent
 * mode is set to DEPARSE_PART_INDENT_AND_MERGE. This is utilized to put
 * certain items, for example the target list items in SELECT on one or more
 * lines, breaking at separators as needed to respect the maximum line limit.
 *
 * ## Node context
 *
 * Node context is used when the same function in the deparser needs to behave
 * differently depending on the parent node. For example, a SELECT statement
 * needs wrapping parenthesis in certain situations where the parent statement
 * uses a set operation (e.g. UNION).
 *
 * ## Comments
 *
 * Comments do not exist as nodes in a Postgres parse tree, and as such need to
 * be passed in separately through the deparser options. They are placed in the
 * output on a best effort basis, by matching against node locations. Comments
 * will only be output at most once, but may not be output at all in certain
 * cases, for example when the comment's specified match location is never
 * reached by any of the nodes visited.
 */

typedef enum DeparseNodeContext {
	DEPARSE_NODE_CONTEXT_NONE,
	// Parent node type (and sometimes field)
	DEPARSE_NODE_CONTEXT_INSERT_RELATION,
	DEPARSE_NODE_CONTEXT_INSERT_SELECT,
	DEPARSE_NODE_CONTEXT_A_EXPR,
	DEPARSE_NODE_CONTEXT_B_EXPR,
	DEPARSE_NODE_CONTEXT_CREATE_TYPE,
	DEPARSE_NODE_CONTEXT_ALTER_TYPE,
	DEPARSE_NODE_CONTEXT_ALTER_DOMAIN,
	DEPARSE_NODE_CONTEXT_SET_STATEMENT,
	DEPARSE_NODE_CONTEXT_FUNC_EXPR,
	DEPARSE_NODE_CONTEXT_SELECT_SETOP,
	DEPARSE_NODE_CONTEXT_SELECT_SORT_CLAUSE,
	// Identifier vs constant context
	DEPARSE_NODE_CONTEXT_IDENTIFIER,
	DEPARSE_NODE_CONTEXT_CONSTANT
} DeparseNodeContext;

typedef enum DeparsePartIndentMode {
	/* Don't indent parts at all, used in special cases (e.g. WITH clauses) */
	DEPARSE_PART_NO_INDENT,

	/* Indent parts but don't merge them (keep each on their own line) */
	DEPARSE_PART_INDENT,

	/* Indent parts and merge them together up to the max line length */
	DEPARSE_PART_INDENT_AND_MERGE
} DeparsePartIndentMode;

// Each part is typically one line to be emitted in pretty print mode
typedef struct DeparseStatePart
{
	StringInfo str;

	/* If this gets emitted as its own line, number of spaces to be added as indentation */
	int indent;

	/* Allow merging this part with adjacent parts that are marked as mergeable */
	bool mergeable;
} DeparseStatePart;

// A part group are typically all parts associated with a major keyword
typedef struct DeparseStatePartGroup
{
	const char *keyword;
	List *parts;
	DeparsePartIndentMode indent_mode;
} DeparseStatePartGroup;

// Nesting levels may be statements, or complex parts of statements that should be indented (e.g. CASE)
typedef struct DeparseStateNestingLevel
{
	/* List of DeparseStatePartGroup items */
	List *part_groups;

	/* Add this much indentation to every part */
	int base_indent;
} DeparseStateNestingLevel;

typedef struct DeparseState
{
	DeparseStateNestingLevel *current;

	/* List of DeparseStatePart items */
	List *result_parts;

	/* Deparse options originally passed in */
	PostgresDeparseOpts *opts;

	/* Set of indexes of comments already placed in the output query */
	Bitmapset *emitted_comments;
} DeparseState;

static void deparseSelectStmt(DeparseState *state, SelectStmt *stmt, DeparseNodeContext context);
static void deparseIntoClause(DeparseState *state, IntoClause *into_clause);
static void deparseRangeVar(DeparseState *state, RangeVar *range_var, DeparseNodeContext context);
static void deparseResTarget(DeparseState *state, ResTarget *res_target, DeparseNodeContext context);
static void deparseAlias(DeparseState *state, Alias *alias);
static void deparseWindowDef(DeparseState *state, WindowDef* window_def);
static void deparseColumnRef(DeparseState *state, ColumnRef* column_ref);
static void deparseSubLink(DeparseState *state, SubLink* sub_link);
static void deparseAExpr(DeparseState *state, A_Expr* a_expr, DeparseNodeContext context);
static bool needsParensAsBExpr(Node *node);
static bool needsParensAsCExpr(Node *node);
static void deparseBoolExpr(DeparseState *state, BoolExpr *bool_expr);
static void deparseAStar(DeparseState *state, A_Star* a_star);
static void deparseCollateClause(DeparseState *state, CollateClause* collate_clause);
static void deparseSortBy(DeparseState *state, SortBy* sort_by);
static void deparseParamRef(DeparseState *state, ParamRef* param_ref);
static void deparseSQLValueFunction(DeparseState *state, SQLValueFunction* sql_value_function);
static void deparseWithClause(DeparseState *state, WithClause *with_clause);
static void deparseJoinExpr(DeparseState *state, JoinExpr *join_expr);
static void deparseCommonTableExpr(DeparseState *state, CommonTableExpr *cte);
static void deparseRangeSubselect(DeparseState *state, RangeSubselect *range_subselect);
static void deparseRangeFunction(DeparseState *state, RangeFunction *range_func);
static void deparseAArrayExpr(DeparseState *state, A_ArrayExpr * array_expr);
static void deparseRowExpr(DeparseState *state, RowExpr *row_expr);
static void deparseTypeCast(DeparseState *state, TypeCast *type_cast, DeparseNodeContext context);
static void deparseTypeName(DeparseState *state, TypeName *type_name);
static void deparseIntervalTypmods(DeparseState *state, TypeName *type_name);
static void deparseNullTest(DeparseState *state, NullTest *null_test);
static void deparseCaseExpr(DeparseState *state, CaseExpr *case_expr);
static void deparseCaseWhen(DeparseState *state, CaseWhen *case_when);
static void deparseAIndirection(DeparseState *state, A_Indirection *a_indirection);
static void deparseAIndices(DeparseState *state, A_Indices *a_indices);
static void deparseCoalesceExpr(DeparseState *state, CoalesceExpr *coalesce_expr);
static void deparseBooleanTest(DeparseState *state, BooleanTest *boolean_test);
static void deparseColumnDef(DeparseState *state, ColumnDef *column_def);
static void deparseInsertStmt(DeparseState *state, InsertStmt *insert_stmt);
static void deparseOnConflictClause(DeparseState *state, OnConflictClause *on_conflict_clause);
static void deparseIndexElem(DeparseState *state, IndexElem* index_elem);
static void deparseUpdateStmt(DeparseState *state, UpdateStmt *update_stmt);
static void deparseDeleteStmt(DeparseState *state, DeleteStmt *delete_stmt);
static void deparseLockingClause(DeparseState *state, LockingClause *locking_clause);
static void deparseSetToDefault(DeparseState *state, SetToDefault *set_to_default);
static void deparseCreateCastStmt(DeparseState *state, CreateCastStmt *create_cast_stmt);
static void deparseCreateDomainStmt(DeparseState *state, CreateDomainStmt *create_domain_stmt);
static void deparseFunctionParameter(DeparseState *state, FunctionParameter *function_parameter);
static void deparseRoleSpec(DeparseState *state, RoleSpec *role_spec);
static void deparseViewStmt(DeparseState *state, ViewStmt *view_stmt);
static void deparseVariableSetStmt(DeparseState *state, VariableSetStmt* variable_set_stmt);
static void deparseReplicaIdentityStmt(DeparseState *state, ReplicaIdentityStmt *replica_identity_stmt);
static void deparseRangeTableSample(DeparseState *state, RangeTableSample *range_table_sample);
static void deparseRangeTableFunc(DeparseState *state, RangeTableFunc* range_table_func);
static void deparseGroupingSet(DeparseState *state, GroupingSet *grouping_set);
static void deparseFuncCall(DeparseState *state, FuncCall *func_call);
static void deparseMinMaxExpr(DeparseState *state, MinMaxExpr *min_max_expr);
static void deparseXmlExpr(DeparseState *state, XmlExpr* xml_expr, DeparseNodeContext context);
static void deparseXmlSerialize(DeparseState *state, XmlSerialize *xml_serialize);
static void deparseJsonIsPredicate(DeparseState *state, JsonIsPredicate *json_is_predicate);
static void deparseJsonObjectAgg(DeparseState *state, JsonObjectAgg *json_object_agg);
static void deparseJsonArrayAgg(DeparseState *state, JsonArrayAgg *json_array_agg);
static void deparseJsonObjectConstructor(DeparseState *state, JsonObjectConstructor *json_object_constructor);
static void deparseJsonArrayConstructor(DeparseState *state, JsonArrayConstructor *json_array_constructor);
static void deparseJsonArrayQueryConstructor(DeparseState *state, JsonArrayQueryConstructor *json_array_query_constructor);
static void deparseJsonValueExpr(DeparseState *state, JsonValueExpr *json_value_expr);
static void deparseJsonOutput(DeparseState *state, JsonOutput *json_output);
static void deparseJsonParseExpr(DeparseState *state, JsonParseExpr *json_parse_expr);
static void deparseJsonScalarExpr(DeparseState *state, JsonScalarExpr *json_scalar_expr);
static void deparseJsonSerializeExpr(DeparseState *state, JsonSerializeExpr *json_serialize_expr);
static void deparseJsonTable(DeparseState *state, JsonTable *json_table);
static void deparseJsonTableColumn(DeparseState *state, JsonTableColumn *json_table_column);
static void deparseJsonTableColumns(DeparseState *state, List *json_table_columns);
static void deparseJsonTablePathSpec(DeparseState *state, JsonTablePathSpec *json_table_path_spec);
static void deparseJsonBehavior(DeparseState *state, JsonBehavior *json_behavior);
static void deparseJsonFuncExpr(DeparseState *state, JsonFuncExpr *json_func_expr);
static void deparseJsonQuotesClauseOpt(DeparseState *state, JsonQuotes quotes);
static void deparseJsonOnErrorClauseOpt(DeparseState *state, JsonBehavior *behavior);
static void deparseJsonOnEmptyClauseOpt(DeparseState *state, JsonBehavior *behavior);
static void deparseConstraint(DeparseState *state, Constraint *constraint, DeparseNodeContext context);
static void deparseATAlterConstraint(DeparseState *state, ATAlterConstraint *constraint);
static void deparseSchemaStmt(DeparseState *state, Node *node);
static void deparseExecuteStmt(DeparseState *state, ExecuteStmt *execute_stmt);
static void deparseTriggerTransition(DeparseState *state, TriggerTransition *trigger_transition);
static void deparseCreateOpClassItem(DeparseState *state, CreateOpClassItem *create_op_class_item);
static void deparseAConst(DeparseState *state, A_Const *a_const);
static void deparseGroupingFunc(DeparseState *state, GroupingFunc *grouping_func);

static void deparsePreparableStmt(DeparseState *state, Node *node);
static void deparseRuleActionStmt(DeparseState *state, Node *node);
static void deparseExplainableStmt(DeparseState *state, Node *node);
static void deparseStmt(DeparseState *state, Node *node);
static void deparseValue(DeparseState *state, union ValUnion *value, DeparseNodeContext context);

static void
removeTrailingSpaceFromStr(StringInfo str)
{
	if (str->len >= 1 && str->data[str->len - 1] == ' ') {
		str->len -= 1;
		str->data[str->len] = '\0';
	}
}

// Check whether the value is a reserved keyword, to determine escaping for output
//
// Note that since the parser lowercases all keywords, this does *not* match when the
// value is not all-lowercase and a reserved keyword.
static bool
isReservedKeyword(const char *val)
{
	int	kwnum = ScanKeywordLookup(val, &ScanKeywords);
	bool all_lower_case = true;
	const char *cp;

	for (cp = val; *cp; cp++)
	{
		if (!(
			(*cp >= 'a' && *cp <= 'z') ||
			(*cp >= '0' && *cp <= '9') ||
			(*cp == '_')))
		{
			all_lower_case = false;
			break;
		}
	}

	return all_lower_case && kwnum >= 0 && ScanKeywordCategories[kwnum] == RESERVED_KEYWORD;
}

// Returns whether the given value consists only of operator characters
static bool
isOp(const char *val)
{
	const char *cp;

	Assert(strlen(val) > 0);

	for (cp = val; *cp; cp++)
	{
		if (!(
			*cp == '~' ||
			*cp == '!' ||
			*cp == '@' ||
			*cp == '#' ||
			*cp == '^' ||
			*cp == '&' ||
			*cp == '|' ||
			*cp == '`' ||
			*cp == '?' ||
			*cp == '+' ||
			*cp == '-' ||
			*cp == '*' ||
			*cp == '/' ||
			*cp == '%' ||
			*cp == '<' ||
			*cp == '>' ||
			*cp == '='))
			return false;
	}

	return true;
}

static DeparseStatePart *
makeDeparseStatePart(DeparseState *state, DeparseStateNestingLevel *level, DeparsePartIndentMode indent_mode)
{
	DeparseStatePart *part = palloc(sizeof(DeparseStatePart));
	part->str = makeStringInfo();
	part->indent = level->base_indent;
	if (indent_mode != DEPARSE_PART_NO_INDENT)
		part->indent += state->opts->indent_size;
	part->mergeable = indent_mode == DEPARSE_PART_INDENT_AND_MERGE;
	return part;
}

static void
freeDeparseStatePart(DeparseStatePart *part)
{
	pfree(part->str->data);
	pfree(part->str);
	pfree(part);
}

// Returns current active part group, and initializes the first one (including an empty part) if needed
static DeparseStatePartGroup *
deparseGetCurrentPartGroup(DeparseState *state)
{
	DeparseStateNestingLevel *level = state->current;
	if (level->part_groups)
		return (DeparseStatePartGroup *) llast(level->part_groups);

	DeparseStatePartGroup *part_group = palloc0(sizeof(DeparseStatePartGroup));
	part_group->parts = lappend(part_group->parts, makeDeparseStatePart(state, level, part_group->indent_mode));
	level->part_groups = lappend(level->part_groups, part_group);
	return part_group;
}

static DeparseStatePart *
deparseGetCurrentPart(DeparseState *state)
{
	DeparseStatePartGroup *part_group = deparseGetCurrentPartGroup(state);
	return (DeparseStatePart *) llast(part_group->parts);
}

static void
deparseMarkCurrentPartNonMergable(DeparseState *state)
{
	DeparseStatePart *part = deparseGetCurrentPart(state);
	part->mergeable = false;
}

static StringInfo deparseGetCurrentStringInfo(DeparseState *state)
{
	DeparseStatePart *part = deparseGetCurrentPart(state);
	Assert(part != NULL);
	return part->str;
}

static void deparseAppendStringInfo(DeparseState *state, const char *fmt,...)
{
	StringInfo	str = deparseGetCurrentStringInfo(state);
	int			save_errno = errno;

	for (;;)
	{
		va_list		args;
		int			needed;

		/* Try to format the data. */
		errno = save_errno;
		va_start(args, fmt);
		needed = appendStringInfoVA(str, fmt, args);
		va_end(args);

		if (needed == 0)
			break;				/* success */

		/* Increase the buffer size and try again. */
		enlargeStringInfo(str, needed);
	}
}

static void deparseAppendStringInfoString(DeparseState *state, const char *s)
{
	StringInfo str = deparseGetCurrentStringInfo(state);
	appendStringInfoString(str, s);
}

static void deparseAppendStringInfoChar(DeparseState *state, char ch)
{
	StringInfo str = deparseGetCurrentStringInfo(state);
	appendStringInfoChar(str, ch);
}

static void
removeTrailingSpace(DeparseState *state)
{
	StringInfo str = deparseGetCurrentStringInfo(state);
	removeTrailingSpaceFromStr(str);
}

static void
deparseRemoveTrailingEmptyPart(DeparseState *state)
{
	DeparseStatePartGroup *part_group = deparseGetCurrentPartGroup(state);
	DeparseStatePart *last_part = deparseGetCurrentPart(state);

	if (last_part->str->len == 0)
	{
		freeDeparseStatePart(last_part);
		part_group->parts = list_delete_last(part_group->parts);
	}
}

static void
deparseAppendPart(DeparseState *state, bool deduplicate)
{
	DeparseStatePartGroup *part_group = deparseGetCurrentPartGroup(state);

	// Remove previous part if its empty and we deduplicate. We don't keep
	// the existing part since it may have the wrong indent level or mode.
	if (deduplicate)
		deparseRemoveTrailingEmptyPart(state);

	part_group->parts = lappend(part_group->parts, makeDeparseStatePart(state, state->current, part_group->indent_mode));
}

static void
deparseAppendCommaAndPart(DeparseState *state)
{
	if (state->opts->commas_start_of_line)
	{
		deparseAppendPart(state, true);
		deparseAppendStringInfoString(state, ", ");
	}
	else
	{
		deparseAppendStringInfoChar(state, ',');
		deparseAppendPart(state, true);
	}
}

static void
deparseAppendPartGroup(DeparseState *state, const char *keyword, DeparsePartIndentMode indent_mode)
{
	DeparseStateNestingLevel *level = state->current;
	DeparseStatePartGroup *part_group = palloc0(sizeof(DeparseStatePartGroup));

	if (list_length(level->part_groups) > 0)
		removeTrailingSpace(state);

	part_group->keyword = keyword;
	part_group->parts = lappend(part_group->parts, makeDeparseStatePart(state, state->current, indent_mode));
	part_group->indent_mode = indent_mode;

	level->part_groups = lappend(level->part_groups, part_group);
}

static void
deparseAppendCommentsIfNeeded(DeparseState *state, ParseLoc location)
{
	for (int i = 0; i < state->opts->comment_count; i++)
	{
		if (bms_is_member(i, state->emitted_comments))
			continue;

		PostgresDeparseComment *comment = state->opts->comments[i];
		if (comment->match_location > location)
			continue;

		// Emit one less leading newline if we already emitted one for formatting reasons
		int newlines_before_comment = comment->newlines_before_comment;
		if (state->opts->pretty_print && newlines_before_comment > 0 &&
			deparseGetCurrentPartGroup(state)->keyword != NULL &&
			deparseGetCurrentStringInfo(state)->len == 0)
			newlines_before_comment -= 1;

		for (int j = 0; j < newlines_before_comment; j++)
		{
			if (state->opts->pretty_print)
				deparseAppendPart(state, false);
			else
				deparseAppendStringInfoChar(state, '\n');
		}

		deparseAppendStringInfoString(state, comment->str);

		/* never merge comments with other parts */
		deparseMarkCurrentPartNonMergable(state);

		for (int j = 0; j < comment->newlines_after_comment; j++)
		{
			if (state->opts->pretty_print)
				deparseAppendPart(state, false);
			else
				deparseAppendStringInfoChar(state, '\n');
		}

		state->emitted_comments = bms_add_member(state->emitted_comments, i);
	}
}

static void
deparseEmit(DeparseState *state, StringInfo str)
{
	ListCell *lc;

	/* If the last part is empty, drop it, so we don't confuse the newline output */
	DeparseStatePart *last_part = (DeparseStatePart *) llast(state->result_parts);
	if (last_part && last_part->str->len == 0)
	{
		freeDeparseStatePart(last_part);
		state->result_parts = list_delete_last(state->result_parts);
	}

	foreach (lc, state->result_parts)
	{
		DeparseStatePart *part = (DeparseStatePart *) lfirst(lc);
		bool last_part = list_cell_number(state->result_parts, lc) == list_length(state->result_parts) - 1;

		if (!state->opts->pretty_print && part->str->len > 0 && (part->str->data[0] == ')' || part->str->data[0] == ';'))
			removeTrailingSpaceFromStr(str);

		if (state->opts->pretty_print)
		{
			for (int i = 0; i < part->indent; i++)
				appendStringInfoChar(str, ' ');
		}

		appendStringInfoString(str, part->str->data);
		removeTrailingSpaceFromStr(str);

		if (!last_part)
		{
			if (state->opts->pretty_print)
				appendStringInfoChar(str, '\n');
			else if (str->data[str->len - 1] != '(')
				appendStringInfoChar(str, ' ');
		}

		freeDeparseStatePart(part);
	}
	list_free(state->result_parts);
	state->result_parts = NIL;

	if (state->opts->pretty_print && state->opts->trailing_newline)
		appendStringInfoChar(str, '\n');
}

static DeparseStateNestingLevel *
deparseStateIncreaseNestingLevel(DeparseState *state)
{
	DeparseStateNestingLevel *parent = state->current;
	DeparseStateNestingLevel *level = palloc0(sizeof(DeparseStateNestingLevel));
	if (parent)
	{
		DeparseStatePartGroup *part_group = deparseGetCurrentPartGroup(state);
		level->base_indent = parent->base_indent + state->opts->indent_size;
		if (part_group->indent_mode != DEPARSE_PART_NO_INDENT) /* Indent again if parts next to us are also indented */
			level->base_indent += state->opts->indent_size;

		/* Parts with nested elements don't get merged, even if otherwise permitted */
		deparseMarkCurrentPartNonMergable(state);
	}
	state->current = level;
	return parent;
}

static void
deparseStateDecreaseNestingLevel(DeparseState *state, DeparseStateNestingLevel *parent_level)
{
	ListCell *lc;
	ListCell *lc2;
	DeparseStateNestingLevel *level = state->current;
	Assert(level != NULL);

	foreach (lc, level->part_groups)
	{
		DeparseStatePartGroup *part_group = (DeparseStatePartGroup *) lfirst(lc);

		/* Merge parts */
		if (part_group->indent_mode == DEPARSE_PART_INDENT_AND_MERGE && list_length(part_group->parts) > 1)
		{
			DeparseStatePart *target = (DeparseStatePart *) linitial(part_group->parts);
			for_each_from (lc2, part_group->parts, 1)
			{
				DeparseStatePart *part = (DeparseStatePart *) lfirst(lc2);
				removeTrailingSpaceFromStr(target->str);
				if (target->mergeable && part->mergeable &&
					target->indent + target->str->len + 1 + part->str->len <= state->opts->max_line_length)
				{
					if (target->str->len > 0 && target->str->data[target->str->len - 1] != '(')
						appendStringInfoChar(target->str, ' ');
					appendStringInfoString(target->str, part->str->data);
					freeDeparseStatePart(part);
					part_group->parts = foreach_delete_current(part_group->parts, lc2);
				}
				else
				{
					target = part;
				}
			}
		}

		if (part_group->keyword != NULL)
		{
			DeparseStatePart *target = makeDeparseStatePart(state, level, false);
			appendStringInfoString(target->str, part_group->keyword);
			part_group->parts = list_insert_nth(part_group->parts, 0, target);

			if (list_length(part_group->parts) == 2 || part_group->indent_mode == DEPARSE_PART_NO_INDENT)
			{
				DeparseStatePart *part = (DeparseStatePart *) lsecond(part_group->parts);
				if (part->str->len > 0)
					appendStringInfo(target->str, " %s", part->str->data);
				freeDeparseStatePart(part);
				part_group->parts = list_delete_nth_cell(part_group->parts, 1);
			}
		}

		if (parent_level)
		{
			DeparseStatePartGroup *parent_part_group = (DeparseStatePartGroup *) llast(parent_level->part_groups);
			parent_part_group->parts = list_concat(parent_part_group->parts, part_group->parts);
		}
		else
		{
			// If its the top level, save as results instead
			state->result_parts = list_concat(state->result_parts, part_group->parts);
		}

		list_free(part_group->parts);
		pfree(part_group);
	}
	list_free(level->part_groups);
	pfree(level);

	state->current = parent_level;

	if (parent_level)
	{
		/* Make sure parent statement writes that follow are on their own line */
		deparseAppendPart(state, true);

		/* Parts with nested elements don't get merged, even if otherwise permitted */
		deparseMarkCurrentPartNonMergable(state);
	}
}

/*
 * Append a SQL string literal representing "val" to buf.
 *
 * Copied here from postgres_fdw/deparse.c to avoid adding
 * many additional dependencies, and modified to work with deparser
 * state.
 */
static void
deparseStringLiteral(DeparseState *state, const char *val)
{
	const char *valptr;

	/*
	 * Rather than making assumptions about the remote server's value of
	 * standard_conforming_strings, always use E'foo' syntax if there are any
	 * backslashes.  This will fail on remote servers before 8.1, but those
	 * are long out of support.
	 */
	if (strchr(val, '\\') != NULL)
		deparseAppendStringInfoChar(state, ESCAPE_STRING_SYNTAX);
	deparseAppendStringInfoChar(state, '\'');
	for (valptr = val; *valptr; valptr++)
	{
		char		ch = *valptr;

		if (SQL_STR_DOUBLE(ch, true))
			deparseAppendStringInfoChar(state, ch);
		deparseAppendStringInfoChar(state, ch);
	}
	deparseAppendStringInfoChar(state, '\'');
}

// "any_name" in gram.y
static void deparseAnyName(DeparseState *state, List *parts)
{
	ListCell *lc = NULL;

	foreach(lc, parts)
	{
		Assert(IsA(lfirst(lc), String));
		deparseAppendStringInfoString(state, quote_identifier(strVal(lfirst(lc))));
		if (lnext(parts, lc))
			deparseAppendStringInfoChar(state, '.');
	}
}
static void deparseAnyNameSkipFirst(DeparseState *state, List *parts)
{
	ListCell *lc = NULL;

	for_each_from(lc, parts, 1)
	{
		Assert(IsA(lfirst(lc), String));
		deparseAppendStringInfoString(state, quote_identifier(strVal(lfirst(lc))));
		if (lnext(parts, lc))
			deparseAppendStringInfoChar(state, '.');
	}
}
static void deparseAnyNameSkipLast(DeparseState *state, List *parts)
{
	ListCell *lc = NULL;

	foreach (lc, parts)
	{
		if (lnext(parts, lc))
		{
			deparseAppendStringInfoString(state, quote_identifier(strVal(lfirst(lc))));
			if (foreach_current_index(lc) < list_length(parts) - 2)
				deparseAppendStringInfoChar(state, '.');
		}
	}
}

// "func_expr" in gram.y
static void deparseFuncExpr(DeparseState *state, Node *node, DeparseNodeContext context)
{
	if (node == NULL)
		elog(ERROR, "deparse: unexpected NULL node");

	switch (nodeTag(node))
	{
		case T_FuncCall:
			deparseFuncCall(state, castNode(FuncCall, node));
			break;
		case T_SQLValueFunction:
			deparseSQLValueFunction(state, castNode(SQLValueFunction, node));
			break;
		case T_MinMaxExpr:
			deparseMinMaxExpr(state, castNode(MinMaxExpr, node));
			break;
		case T_CoalesceExpr:
			deparseCoalesceExpr(state, castNode(CoalesceExpr, node));
			break;
		case T_XmlExpr:
			deparseXmlExpr(state, castNode(XmlExpr, node), context);
			break;
		case T_XmlSerialize:
			deparseXmlSerialize(state, castNode(XmlSerialize, node));
			break;
		case T_JsonObjectAgg:
			deparseJsonObjectAgg(state, castNode(JsonObjectAgg, node));
			break;
		case T_JsonArrayAgg:
			deparseJsonArrayAgg(state, castNode(JsonArrayAgg, node));
			break;
		case T_JsonObjectConstructor:
			deparseJsonObjectConstructor(state, castNode(JsonObjectConstructor, node));
			break;
		case T_JsonArrayConstructor:
			deparseJsonArrayConstructor(state, castNode(JsonArrayConstructor, node));
			break;
		case T_JsonArrayQueryConstructor:
			deparseJsonArrayQueryConstructor(state, castNode(JsonArrayQueryConstructor, node));
			break;
		default:
			elog(ERROR, "deparse: unpermitted node type in func_expr: %d",
				 (int) nodeTag(node));
			break;
	}
}

static void deparseCExpr(DeparseState *state, Node *node);

// "a_expr" in gram.y
static void deparseExpr(DeparseState *state, Node *node, DeparseNodeContext context)
{
	check_stack_depth();

	if (node == NULL)
		return;

	switch (nodeTag(node))
	{
		case T_ColumnRef:
		case T_A_Const:
		case T_ParamRef:
		case T_A_Indirection:
		case T_CaseExpr:
		case T_A_ArrayExpr:
		case T_RowExpr:
		case T_GroupingFunc:
			deparseCExpr(state, node);
			break;
		case T_SubLink:
			// Not always a "c_expr" (e.g. "x = ANY (...)"), so don't go through deparseCExpr
			deparseSubLink(state, castNode(SubLink, node));
			break;
		case T_TypeCast:
			deparseTypeCast(state, castNode(TypeCast, node), DEPARSE_NODE_CONTEXT_NONE);
			break;
		case T_CollateClause:
			deparseCollateClause(state, castNode(CollateClause, node));
			break;
		case T_A_Expr:
			// Operands of A_Expr nodes in "b_expr" positions must be valid "b_expr" as well
			deparseAExpr(state, castNode(A_Expr, node), context == DEPARSE_NODE_CONTEXT_B_EXPR ? DEPARSE_NODE_CONTEXT_B_EXPR : DEPARSE_NODE_CONTEXT_A_EXPR);
			break;
		case T_BoolExpr:
			deparseBoolExpr(state, castNode(BoolExpr, node));
			break;
		case T_NullTest:
			deparseNullTest(state, castNode(NullTest, node));
			break;
		case T_BooleanTest:
			deparseBooleanTest(state, castNode(BooleanTest, node));
			break;
		case T_JsonIsPredicate:
			deparseJsonIsPredicate(state, castNode(JsonIsPredicate, node));
			break;
		case T_SetToDefault:
			deparseSetToDefault(state, castNode(SetToDefault, node));
			break;
		case T_MergeSupportFunc:
			deparseAppendStringInfoString(state, "merge_action() ");
			break;
		case T_JsonParseExpr:
			deparseJsonParseExpr(state, castNode(JsonParseExpr, node));
			break;
		case T_JsonScalarExpr:
			deparseJsonScalarExpr(state, castNode(JsonScalarExpr, node));
			break;
		case T_JsonSerializeExpr:
			deparseJsonSerializeExpr(state, castNode(JsonSerializeExpr, node));
			break;
		case T_JsonFuncExpr:
			deparseJsonFuncExpr(state, castNode(JsonFuncExpr, node));
			break;
		case T_FuncCall:
		case T_SQLValueFunction:
		case T_MinMaxExpr:
		case T_CoalesceExpr:
		case T_XmlExpr:
		case T_XmlSerialize:
		case T_JsonObjectAgg:
		case T_JsonArrayAgg:
		case T_JsonObjectConstructor:
		case T_JsonArrayConstructor:
		case T_JsonArrayQueryConstructor:
			deparseFuncExpr(state, node, context);
			break;
		default:
			// Note that this is also the fallthrough for deparseBExpr and deparseCExpr
			elog(ERROR, "deparse: unpermitted node type in a_expr/b_expr/c_expr: %d",
				 (int) nodeTag(node));
			break;
	}
}

// "b_expr" in gram.y
static void deparseBExpr(DeparseState *state, Node *node)
{
	if (node == NULL)
		elog(ERROR, "deparse: unexpected NULL node in deparseBExpr");

	if (needsParensAsBExpr(node))
	{
		deparseAppendStringInfoChar(state, '(');
		// Because we wrap this in parenthesis, the expression inside follows "a_expr" parser rules
		deparseExpr(state, node, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoChar(state, ')');
		return;
	}

	deparseExpr(state, node, DEPARSE_NODE_CONTEXT_B_EXPR);
}

// "AexprConst" in gram.y
static void deparseAexprConst(DeparseState *state, Node *node)
{
	if (node == NULL)
		elog(ERROR, "deparse: unexpected NULL node");

	switch (nodeTag(node))
	{
		case T_A_Const:
			deparseAConst(state, castNode(A_Const, node));
			break;
		case T_TypeCast:
			deparseTypeCast(state, castNode(TypeCast, node), DEPARSE_NODE_CONTEXT_NONE);
			break;
		default:
			elog(ERROR, "deparse: unpermitted node type in AexprConst: %d",
				 (int) nodeTag(node));
			break;
	}
}

// "c_expr" in gram.y
static void deparseCExpr(DeparseState *state, Node *node)
{
	if (node == NULL)
		elog(ERROR, "deparse: unexpected NULL node");

	if (needsParensAsCExpr(node))
	{
		deparseAppendStringInfoChar(state, '(');
		deparseExpr(state, node, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoChar(state, ')');
		return;
	}

	switch (nodeTag(node))
	{
		case T_ColumnRef:
			deparseColumnRef(state, castNode(ColumnRef, node));
			break;
		case T_A_Const:
			deparseAConst(state, castNode(A_Const, node));
			break;
		case T_ParamRef:
			deparseParamRef(state, castNode(ParamRef, node));
			break;
		case T_A_Indirection:
			deparseAIndirection(state, castNode(A_Indirection, node));
			break;
		case T_CaseExpr:
			deparseCaseExpr(state, castNode(CaseExpr, node));
			break;
		case T_SubLink:
			deparseSubLink(state, castNode(SubLink, node));
			break;
		case T_A_ArrayExpr:
			deparseAArrayExpr(state, castNode(A_ArrayExpr, node));
			break;
		case T_RowExpr:
			deparseRowExpr(state, castNode(RowExpr, node));
			break;
		case T_GroupingFunc:
			deparseGroupingFunc(state, castNode(GroupingFunc, node));
			break;
		case T_FuncCall:
		case T_SQLValueFunction:
		case T_MinMaxExpr:
		case T_CoalesceExpr:
		case T_XmlExpr:
		case T_XmlSerialize:
		case T_JsonObjectAgg:
		case T_JsonArrayAgg:
		case T_JsonObjectConstructor:
		case T_JsonArrayConstructor:
		case T_JsonArrayQueryConstructor:
			deparseFuncExpr(state, node, DEPARSE_NODE_CONTEXT_NONE);
			break;
		default:
			deparseAppendStringInfoChar(state, '(');
			// Because we wrap this in parenthesis, the expression inside follows "a_expr" parser rules
			deparseExpr(state, node, DEPARSE_NODE_CONTEXT_A_EXPR);
			deparseAppendStringInfoChar(state, ')');
			break;
	}
}

// "expr_list" in gram.y
static void deparseExprList(DeparseState *state, List *exprs)
{
	ListCell *lc;
	foreach(lc, exprs)
	{
		deparseExpr(state, lfirst(lc), DEPARSE_NODE_CONTEXT_A_EXPR);
		if (lnext(exprs, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "ColId", "name", "database_name", "access_method" and "index_name" in gram.y
static void deparseColId(DeparseState *state, char *s)
{
	deparseAppendStringInfoString(state, quote_identifier(s));
}

// "ColLabel", "attr_name"
//
// Note this is kept separate from ColId in case we ever want to be more
// specific on how to handle keywords here
static void deparseColLabel(DeparseState *state, char *s)
{
	deparseAppendStringInfoString(state, quote_identifier(s));
}

// "SignedIconst" and "Iconst" in gram.y
static void deparseSignedIconst(DeparseState *state, Node *node)
{
	deparseAppendStringInfo(state, "%d", intVal(node));
}

// "indirection" and "opt_indirection" in gram.y
static void deparseOptIndirection(DeparseState *state, List *indirection, int N)
{
	ListCell *lc = NULL;

	for_each_from(lc, indirection, N)
	{
		if (IsA(lfirst(lc), String))
		{
			deparseAppendStringInfoChar(state, '.');
			deparseColLabel(state, strVal(lfirst(lc)));
		}
		else if (IsA(lfirst(lc), A_Star))
		{
			deparseAppendStringInfoString(state, ".*");
		}
		else if (IsA(lfirst(lc), A_Indices))
		{
			deparseAIndices(state, castNode(A_Indices, lfirst(lc)));
		}
		else
		{
			// No other nodes should appear here
			Assert(false);
		}
	}
}

// "role_list" in gram.y
static void deparseRoleList(DeparseState *state, List *roles)
{
	ListCell *lc;

	foreach(lc, roles)
	{
		RoleSpec *role_spec = castNode(RoleSpec, lfirst(lc));
		deparseRoleSpec(state, role_spec);
		if (lnext(roles, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "SimpleTypename" in gram.y
static void deparseSimpleTypename(DeparseState *state, Node *node)
{
	deparseTypeName(state, castNode(TypeName, node));
}

// "NumericOnly" in gram.y
static void deparseNumericOnly(DeparseState *state, union ValUnion *value)
{
	if (value == NULL)
		elog(ERROR, "deparse: unexpected NULL value in deparseNumericOnly");

	switch (nodeTag(value))
	{
		case T_Integer:
			deparseAppendStringInfo(state, "%d", value->ival.ival);
			break;
		case T_Float:
			deparseAppendStringInfoString(state, value->sval.sval);
			break;
		default:
			Assert(false);
	}
}

// "NumericOnly_list" in gram.y
static void deparseNumericOnlyList(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	foreach(lc, l)
	{
		deparseNumericOnly(state, (union ValUnion *) lfirst(lc));
		if (lnext(l, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "SeqOptElem" in gram.y
static void deparseSeqOptElem(DeparseState *state, DefElem *def_elem)
{
	ListCell *lc;

	if (strcmp(def_elem->defname, "as") == 0)
	{
		deparseAppendStringInfoString(state, "AS ");
		deparseSimpleTypename(state, def_elem->arg);
	}
	else if (strcmp(def_elem->defname, "cache") == 0)
	{
		deparseAppendStringInfoString(state, "CACHE ");
		deparseNumericOnly(state, (union ValUnion *) def_elem->arg);
	}
	else if (strcmp(def_elem->defname, "cycle") == 0 && boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "CYCLE");
	}
	else if (strcmp(def_elem->defname, "cycle") == 0 && !boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "NO CYCLE");
	}
	else if (strcmp(def_elem->defname, "increment") == 0)
	{
		deparseAppendStringInfoString(state, "INCREMENT ");
		deparseNumericOnly(state, (union ValUnion *) def_elem->arg);
	}
	else if (strcmp(def_elem->defname, "maxvalue") == 0 && def_elem->arg != NULL)
	{
		deparseAppendStringInfoString(state, "MAXVALUE ");
		deparseNumericOnly(state, (union ValUnion *) def_elem->arg);
	}
	else if (strcmp(def_elem->defname, "maxvalue") == 0 && def_elem->arg == NULL)
	{
		deparseAppendStringInfoString(state, "NO MAXVALUE");
	}
	else if (strcmp(def_elem->defname, "minvalue") == 0 && def_elem->arg != NULL)
	{
		deparseAppendStringInfoString(state, "MINVALUE ");
		deparseNumericOnly(state, (union ValUnion *) def_elem->arg);
	}
	else if (strcmp(def_elem->defname, "minvalue") == 0 && def_elem->arg == NULL)
	{
		deparseAppendStringInfoString(state, "NO MINVALUE");
	}
	else if (strcmp(def_elem->defname, "owned_by") == 0)
	{
		deparseAppendStringInfoString(state, "OWNED BY ");
		deparseAnyName(state, castNode(List, def_elem->arg));
	}
	else if (strcmp(def_elem->defname, "sequence_name") == 0)
	{
		deparseAppendStringInfoString(state, "SEQUENCE NAME ");
		deparseAnyName(state, castNode(List, def_elem->arg));
	}
	else if (strcmp(def_elem->defname, "start") == 0)
	{
		deparseAppendStringInfoString(state, "START ");
		deparseNumericOnly(state, (union ValUnion *) def_elem->arg);
	}
	else if (strcmp(def_elem->defname, "restart") == 0 && def_elem->arg == NULL)
	{
		deparseAppendStringInfoString(state, "RESTART");
	}
	else if (strcmp(def_elem->defname, "restart") == 0 && def_elem->arg != NULL)
	{
		deparseAppendStringInfoString(state, "RESTART ");
		deparseNumericOnly(state, (union ValUnion *) def_elem->arg);
	}
	else
	{
		Assert(false);
	}
}

// "SeqOptList" in gram.y
static void deparseSeqOptList(DeparseState *state, List *options)
{
	ListCell *lc;
	Assert(list_length(options) > 0);
	foreach (lc, options)
	{
		deparseSeqOptElem(state, castNode(DefElem, lfirst(lc)));
		deparseAppendStringInfoChar(state, ' ');
	}
}

// "OptSeqOptList" in gram.y
static void deparseOptSeqOptList(DeparseState *state, List *options)
{
	if (list_length(options) > 0)
		deparseSeqOptList(state, options);
}

// "OptParenthesizedSeqOptList" in gram.y
static void deparseOptParenthesizedSeqOptList(DeparseState *state, List *options)
{
	if (list_length(options) > 0)
	{
		deparseAppendStringInfoChar(state, '(');
		deparseSeqOptList(state, options);
		deparseAppendStringInfoChar(state, ')');
	}
}

// "opt_drop_behavior" in gram.y
static void deparseOptDropBehavior(DeparseState *state, DropBehavior behavior)
{
	switch (behavior)
	{
		case DROP_RESTRICT:
			// Default
			break;
		case DROP_CASCADE:
			deparseAppendStringInfoString(state, "CASCADE ");
			break;
	}
}

// "any_operator" in gram.y
static void deparseAnyOperator(DeparseState *state, List *op)
{
	Assert(isOp(strVal(llast(op))));
	if (list_length(op) == 2)
	{
		deparseAppendStringInfoString(state, quote_identifier(strVal(linitial(op))));
		deparseAppendStringInfoChar(state, '.');
		deparseAppendStringInfoString(state, strVal(llast(op)));
	}
	else if (list_length(op) == 1)
	{
		deparseAppendStringInfoString(state, strVal(llast(op)));
	}
	else
	{
		Assert(false);
	}
}

// "qual_Op" and "qual_all_Op" in gram.y
static void deparseQualOp(DeparseState *state, List *op)
{
	if (list_length(op) == 1 && isOp(strVal(linitial(op))))
	{
		deparseAppendStringInfoString(state, strVal(linitial(op)));
	}
	else
	{
		deparseAppendStringInfoString(state, "OPERATOR(");
		deparseAnyOperator(state, op);
		deparseAppendStringInfoString(state, ")");
	}
}

// "subquery_Op" in gram.y
static void deparseSubqueryOp(DeparseState *state, List *op)
{
	if (list_length(op) == 1 && strcmp(strVal(linitial(op)), "~~") == 0)
	{
		deparseAppendStringInfoString(state, "LIKE");
	}
	else if (list_length(op) == 1 && strcmp(strVal(linitial(op)), "!~~") == 0)
	{
		deparseAppendStringInfoString(state, "NOT LIKE");
	}
	else if (list_length(op) == 1 && strcmp(strVal(linitial(op)), "~~*") == 0)
	{
		deparseAppendStringInfoString(state, "ILIKE");
	}
	else if (list_length(op) == 1 && strcmp(strVal(linitial(op)), "!~~*") == 0)
	{
		deparseAppendStringInfoString(state, "NOT ILIKE");
	}
	else if (list_length(op) == 1 && isOp(strVal(linitial(op))))
	{
		deparseAppendStringInfoString(state, strVal(linitial(op)));
	}
	else
	{
		deparseAppendStringInfoString(state, "OPERATOR(");
		deparseAnyOperator(state, op);
		deparseAppendStringInfoString(state, ")");
	}
}

// Precedence of "a_expr" rules, lowest to highest (mirrors %left/%right/%nonassoc in gram.y)
typedef enum DeparseExprPrec
{
	DEPARSE_PREC_OR,         /* OR */
	DEPARSE_PREC_AND,        /* AND */
	DEPARSE_PREC_NOT,        /* NOT (right-associative) */
	DEPARSE_PREC_IS,         /* IS, ISNULL, NOTNULL (non-associative) */
	DEPARSE_PREC_COMPARISON, /* < > = <= >= <> (non-associative) */
	DEPARSE_PREC_BETWEEN,    /* BETWEEN IN LIKE ILIKE SIMILAR NOT_LA (non-associative) */
	DEPARSE_PREC_OP,         /* Op OPERATOR, i.e. all other operators */
	DEPARSE_PREC_ADD,        /* + - */
	DEPARSE_PREC_MUL,        /* * / % */
	DEPARSE_PREC_EXP,        /* ^ */
	DEPARSE_PREC_AT,         /* AT TIME ZONE, AT LOCAL */
	DEPARSE_PREC_COLLATE,    /* COLLATE */
	DEPARSE_PREC_UMINUS,     /* unary + - (right-associative) */
	DEPARSE_PREC_ATOM        /* c_expr, typecasts, and anything else that can't be split apart */
} DeparseExprPrec;

// Precedence of the binary operator tokens handled by "qual_Op" and its special cases in gram.y
static DeparseExprPrec getOperatorPrecedence(List *op)
{
	const char *name;

	// OPERATOR(schema.op) always has the precedence of Op
	if (list_length(op) != 1)
		return DEPARSE_PREC_OP;

	name = strVal(linitial(op));
	if (strcmp(name, "+") == 0 || strcmp(name, "-") == 0)
		return DEPARSE_PREC_ADD;
	if (strcmp(name, "*") == 0 || strcmp(name, "/") == 0 || strcmp(name, "%") == 0)
		return DEPARSE_PREC_MUL;
	if (strcmp(name, "^") == 0)
		return DEPARSE_PREC_EXP;
	if (strcmp(name, "<") == 0 || strcmp(name, ">") == 0 || strcmp(name, "=") == 0 ||
		strcmp(name, "<=") == 0 || strcmp(name, ">=") == 0 || strcmp(name, "<>") == 0 ||
		strcmp(name, "!=") == 0)
		return DEPARSE_PREC_COMPARISON;

	return DEPARSE_PREC_OP;
}

// Precedence of the operator token in "a_expr subquery_Op sub_type ..." (i.e. "x = ANY (...)")
static DeparseExprPrec getSubqueryOperatorPrecedence(List *op)
{
	const char *name;

	if (list_length(op) == 1)
	{
		name = strVal(linitial(op));
		// These get output as [NOT] LIKE / [NOT] ILIKE by deparseSubqueryOp
		if (strcmp(name, "~~") == 0 || strcmp(name, "!~~") == 0 ||
			strcmp(name, "~~*") == 0 || strcmp(name, "!~~*") == 0)
			return DEPARSE_PREC_BETWEEN;
	}

	return getOperatorPrecedence(op);
}

static bool isSqlSyntaxFuncCall(FuncCall *func_call, const char *name)
{
	return func_call->funcformat == COERCE_SQL_SYNTAX &&
		list_length(func_call->funcname) == 2 &&
		strcmp(strVal(linitial(func_call->funcname)), "pg_catalog") == 0 &&
		strcmp(strVal(lsecond(func_call->funcname)), name) == 0;
}

// Precedence of the gram.y rule that a node gets deparsed as (see DeparseExprPrec)
static DeparseExprPrec getExprPrecedence(Node *node)
{
	if (node == NULL)
		return DEPARSE_PREC_ATOM;

	switch (nodeTag(node))
	{
		case T_A_Expr:
			{
				A_Expr *a_expr = castNode(A_Expr, node);
				switch (a_expr->kind)
				{
					case AEXPR_OP:
						// Prefix operators: unary + and - have UMINUS precedence, all others that of Op
						if (a_expr->lexpr == NULL)
							return getOperatorPrecedence(a_expr->name) == DEPARSE_PREC_ADD ? DEPARSE_PREC_UMINUS : DEPARSE_PREC_OP;
						return getOperatorPrecedence(a_expr->name);
					case AEXPR_OP_ANY:
					case AEXPR_OP_ALL:
						// The rule has Op precedence, but as a right operand the operator token may bind looser
						return Min(getSubqueryOperatorPrecedence(a_expr->name), DEPARSE_PREC_OP);
					case AEXPR_DISTINCT:
					case AEXPR_NOT_DISTINCT:
						return DEPARSE_PREC_IS;
					case AEXPR_NULLIF:
						return DEPARSE_PREC_ATOM;
					default: /* IN, LIKE, ILIKE, SIMILAR TO, BETWEEN */
						return DEPARSE_PREC_BETWEEN;
				}
			}
		case T_BoolExpr:
			switch (castNode(BoolExpr, node)->boolop)
			{
				case AND_EXPR:
					return DEPARSE_PREC_AND;
				case NOT_EXPR:
					return DEPARSE_PREC_NOT;
				default:
					return DEPARSE_PREC_OR;
			}
		case T_NullTest:
		case T_BooleanTest:
		case T_JsonIsPredicate:
			return DEPARSE_PREC_IS;
		case T_CollateClause:
			return DEPARSE_PREC_COLLATE;
		case T_XmlExpr:
			if (castNode(XmlExpr, node)->op == IS_DOCUMENT)
				return DEPARSE_PREC_IS;
			return DEPARSE_PREC_ATOM;
		case T_SubLink:
			{
				SubLink *sub_link = castNode(SubLink, node);
				if (sub_link->subLinkType == ANY_SUBLINK && list_length(sub_link->operName) == 0)
					return DEPARSE_PREC_BETWEEN; // IN (...)
				if (sub_link->subLinkType == ANY_SUBLINK || sub_link->subLinkType == ALL_SUBLINK)
					return Min(getSubqueryOperatorPrecedence(sub_link->operName), DEPARSE_PREC_OP);
				return DEPARSE_PREC_ATOM;
			}
		case T_FuncCall:
			{
				FuncCall *func_call = castNode(FuncCall, node);
				if (isSqlSyntaxFuncCall(func_call, "timezone") &&
					list_length(func_call->args) > 0 &&
					list_length(func_call->args) <= 2)
					return DEPARSE_PREC_AT;
				if (isSqlSyntaxFuncCall(func_call, "is_normalized"))
					return DEPARSE_PREC_IS;
				return DEPARSE_PREC_ATOM;
			}
		default:
			return DEPARSE_PREC_ATOM;
	}
}

// Which side of the operator an operand is on (prefix operators only have a right operand, postfix ones a left one)
typedef enum DeparseOperandSide
{
	DEPARSE_OPERAND_LEFT,
	DEPARSE_OPERAND_RIGHT
} DeparseOperandSide;

/*
 * Recognizes subquery comparison forms ("x = ANY (...)", "x < ALL (SELECT ...)",
 * "x IN (...)"). As the left operand of an operator at the same precedence these never
 * need parens (see needsParensForPrecedence); at any other precedence the normal rules
 * apply, since dropping parens there can reassociate the expression or fail to parse:
 * "d = x = ANY(c)" is rejected outright, and "d || x = ANY(c)" reparses with the
 * comparison swallowing the concatenation.
 */
static bool isSubqueryComparisonForm(Node *node)
{
	if (IsA(node, A_Expr))
	{
		switch (castNode(A_Expr, node)->kind)
		{
			case AEXPR_OP_ANY:
			case AEXPR_OP_ALL:
			case AEXPR_IN:
				return true;
			default:
				return false;
		}
	}
	if (IsA(node, SubLink))
	{
		SubLink *sub_link = castNode(SubLink, node);
		return sub_link->subLinkType == ANY_SUBLINK || sub_link->subLinkType == ALL_SUBLINK;
	}
	return false;
}

/*
 * Checks whether a node needs parens as the left or right operand of an "a_expr"
 * rule with the given precedence. At the same precedence, left-associative rules
 * only need them on the right, non-associative rules on both sides.
 */
static bool needsParensForPrecedence(Node *node, DeparseExprPrec parent_prec, DeparseOperandSide side)
{
	DeparseExprPrec prec = getExprPrecedence(node);

	if (side == DEPARSE_OPERAND_LEFT && isSubqueryComparisonForm(node) &&
		prec == parent_prec)
		return false;

	if (prec != parent_prec)
		return prec < parent_prec;

	switch (parent_prec)
	{
		case DEPARSE_PREC_IS:
		case DEPARSE_PREC_COMPARISON:
		case DEPARSE_PREC_BETWEEN:
			return true;
		case DEPARSE_PREC_NOT:
		case DEPARSE_PREC_UMINUS:
			// Right-associative prefix operators, e.g. "NOT NOT a"
			return false;
		default:
			return side == DEPARSE_OPERAND_RIGHT;
	}
}

// Deparses an operand of an "a_expr" (or "b_expr") rule with the given precedence, adding parens if needed
static void deparseExprOperand(DeparseState *state, Node *node, DeparseExprPrec parent_prec, DeparseOperandSide side, DeparseNodeContext context)
{
	bool need_parens = needsParensForPrecedence(node, parent_prec, side) ||
		(context == DEPARSE_NODE_CONTEXT_B_EXPR && needsParensAsBExpr(node));

	if (need_parens)
		deparseAppendStringInfoChar(state, '(');
	deparseExpr(state, node, need_parens ? DEPARSE_NODE_CONTEXT_A_EXPR : context);
	if (need_parens)
		deparseAppendStringInfoChar(state, ')');
}

// Not present directly in gram.y (usually matched by ColLabel)
static void deparseGenericDefElemName(DeparseState *state, const char *in)
{
	Assert(in != NULL);
	char *val = pstrdup(in);
	for (unsigned char *p = (unsigned char *) val; *p; p++)
		*p = pg_toupper(*p);
	deparseAppendStringInfoString(state, val);
	pfree(val);
}

// "def_arg" and "operator_def_arg" in gram.y
static void deparseDefArg(DeparseState *state, Node *arg, bool is_operator_def_arg)
{
	if (IsA(arg, TypeName)) // func_type
	{
		deparseTypeName(state, castNode(TypeName, arg));
	}
	else if (IsA(arg, List)) // qual_all_Op
	{
		List *l = castNode(List, arg);
		Assert(list_length(l) == 1 || list_length(l) == 2);

		// Schema qualified operator
		if (list_length(l) == 2)
		{
			deparseAppendStringInfoString(state, "OPERATOR(");
			deparseAnyOperator(state, l);
			deparseAppendStringInfoChar(state, ')');
		}
		else if (list_length(l) == 1)
		{
			deparseAppendStringInfoString(state, strVal(linitial(l)));
		}
	}
	else if (IsA(arg, Float) || IsA(arg, Integer)) // NumericOnly
	{
		deparseValue(state, (union ValUnion *) arg, DEPARSE_NODE_CONTEXT_NONE);
	}
	else if (IsA(arg, String))
	{
		char *s = strVal(arg);
		if (!is_operator_def_arg && IsA(arg, String) && strcmp(s, "none") == 0) // NONE
		{
			deparseAppendStringInfoString(state, "NONE");
		}
		else if (isReservedKeyword(s)) // reserved_keyword
		{
			deparseAppendStringInfoString(state, s);
		}
		else // Sconst
		{
			deparseStringLiteral(state, s);
		}
	}
	else 
	{
		Assert(false);
	}
}

// "definition" in gram.y
static void deparseDefinition(DeparseState *state, List *options)
{
	ListCell *lc = NULL;

	deparseAppendStringInfoChar(state, '(');
	foreach (lc, options)
	{
		DefElem *def_elem = castNode(DefElem, lfirst(lc));
		deparseAppendStringInfoString(state, quote_identifier(def_elem->defname));
		if (def_elem->arg != NULL) {
			deparseAppendStringInfoString(state, " = ");
			deparseDefArg(state, def_elem->arg, false);
		}

		if (lnext(options, lc))
			deparseAppendStringInfoString(state, ", ");
	}
	deparseAppendStringInfoChar(state, ')');
}

// "opt_definition" in gram.y
//
// Note this method adds a trailing space if a value is output
static void deparseOptDefinition(DeparseState *state, List *options)
{
	if (list_length(options) > 0)
	{
		deparseAppendStringInfoString(state, "WITH ");
		deparseDefinition(state, options);
	}
}

// "create_generic_options" in gram.y
static void deparseCreateGenericOptions(DeparseState *state, List *options)
{
	ListCell *lc = NULL;

	if (options == NULL)
		return;

	deparseAppendStringInfoString(state, "OPTIONS (");
	foreach(lc, options)
	{
		DefElem *def_elem = castNode(DefElem, lfirst(lc));
		deparseAppendStringInfoString(state, quote_identifier(def_elem->defname));
		deparseAppendStringInfoChar(state, ' ');
		deparseStringLiteral(state, strVal(def_elem->arg));
		if (lnext(options, lc))
			deparseAppendStringInfoString(state, ", ");
	}
	deparseAppendStringInfoString(state, ")");
}

// "common_func_opt_item" in gram.y
static void deparseCommonFuncOptItem(DeparseState *state, DefElem *def_elem)
{
	if (strcmp(def_elem->defname, "strict") == 0 && boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "RETURNS NULL ON NULL INPUT");
	}
	else if (strcmp(def_elem->defname, "strict") == 0 && !boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "CALLED ON NULL INPUT");
	}
	else if (strcmp(def_elem->defname, "volatility") == 0 && strcmp(strVal(def_elem->arg), "immutable") == 0)
	{
		deparseAppendStringInfoString(state, "IMMUTABLE");
	}
	else if (strcmp(def_elem->defname, "volatility") == 0 && strcmp(strVal(def_elem->arg), "stable") == 0)
	{
		deparseAppendStringInfoString(state, "STABLE");
	}
	else if (strcmp(def_elem->defname, "volatility") == 0 && strcmp(strVal(def_elem->arg), "volatile") == 0)
	{
		deparseAppendStringInfoString(state, "VOLATILE");
	}
	else if (strcmp(def_elem->defname, "security") == 0 && boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "SECURITY DEFINER");
	}
	else if (strcmp(def_elem->defname, "security") == 0 && !boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "SECURITY INVOKER");
	}
	else if (strcmp(def_elem->defname, "leakproof") == 0 && boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "LEAKPROOF");
	}
	else if (strcmp(def_elem->defname, "leakproof") == 0 && !boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "NOT LEAKPROOF");
	}
	else if (strcmp(def_elem->defname, "cost") == 0)
	{
		deparseAppendStringInfoString(state, "COST ");
		deparseValue(state, (union ValUnion *) def_elem->arg, DEPARSE_NODE_CONTEXT_NONE);
	}
	else if (strcmp(def_elem->defname, "rows") == 0)
	{
		deparseAppendStringInfoString(state, "ROWS ");
		deparseValue(state, (union ValUnion *) def_elem->arg, DEPARSE_NODE_CONTEXT_NONE);
	}
	else if (strcmp(def_elem->defname, "support") == 0)
	{
		deparseAppendStringInfoString(state, "SUPPORT ");
		deparseAnyName(state, castNode(List, def_elem->arg));
	}
	else if (strcmp(def_elem->defname, "set") == 0 && def_elem->arg != NULL && IsA(def_elem->arg, VariableSetStmt)) // FunctionSetResetClause
	{
		deparseVariableSetStmt(state, castNode(VariableSetStmt, def_elem->arg));
	}
	else if (strcmp(def_elem->defname, "parallel") == 0)
	{
		deparseAppendStringInfoString(state, "PARALLEL ");
		deparseAppendStringInfoString(state, quote_identifier(strVal(def_elem->arg)));
	}
	else
	{
		Assert(false);
	}
}

// "NonReservedWord_or_Sconst" in gram.y
//
// Note since both identifiers and string constants are allowed here, we
// currently always return an identifier, except:
//
// 1) when the string is empty (since an empty identifier can't be scanned)
// 2) when the value is equal or larger than NAMEDATALEN (64+ characters)
static void deparseNonReservedWordOrSconst(DeparseState *state, const char *val)
{
	if (strlen(val) == 0)
		deparseAppendStringInfoString(state, "''");
	else if (strlen(val) >= NAMEDATALEN)
		deparseStringLiteral(state, val);
	else
		deparseAppendStringInfoString(state, quote_identifier(val));
}

// "func_as" in gram.y
static void deparseFuncAs(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	foreach(lc, l)
	{
		char *strval = strVal(lfirst(lc));
		if (strstr(strval, "$$") == NULL)
		{
			deparseAppendStringInfoString(state, "$$");
			deparseAppendStringInfoString(state, strval);
			deparseAppendStringInfoString(state, "$$");
		}
		else
		{
			deparseStringLiteral(state, strval);
		}

		if (lnext(l, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "createfunc_opt_item" in gram.y
static void deparseCreateFuncOptItem(DeparseState *state, DefElem *def_elem)
{
	ListCell *lc = NULL;

	if (strcmp(def_elem->defname, "as") == 0)
	{
		deparseAppendStringInfoString(state, "AS ");
		deparseFuncAs(state, castNode(List, def_elem->arg));
	}
	else if (strcmp(def_elem->defname, "language") == 0)
	{
		deparseAppendStringInfoString(state, "LANGUAGE ");
		deparseNonReservedWordOrSconst(state, strVal(def_elem->arg));
	}
	else if (strcmp(def_elem->defname, "transform") == 0)
	{
		List *l = castNode(List, def_elem->arg);
		deparseAppendStringInfoString(state, "TRANSFORM ");
		foreach (lc, l)
		{
			deparseAppendStringInfoString(state, "FOR TYPE ");
			deparseTypeName(state, castNode(TypeName, lfirst(lc)));
			if (lnext(l, lc))
				deparseAppendStringInfoString(state, ", ");
		}
	}
	else if (strcmp(def_elem->defname, "window") == 0)
	{
		deparseAppendStringInfoString(state, "WINDOW");
	}
	else
	{
		deparseCommonFuncOptItem(state, def_elem);
	}
}

// "alter_generic_options" in gram.y
static void deparseAlterGenericOptions(DeparseState *state, List *options)
{
	ListCell *lc = NULL;

	deparseAppendStringInfoString(state, "OPTIONS (");
	foreach(lc, options)
	{
		DefElem *def_elem = castNode(DefElem, lfirst(lc));
		switch (def_elem->defaction)
		{
			case DEFELEM_UNSPEC:
				deparseAppendStringInfoString(state, quote_identifier(def_elem->defname));
				deparseAppendStringInfoChar(state, ' ');
				deparseStringLiteral(state, strVal(def_elem->arg));
				break;
			case DEFELEM_SET:
				deparseAppendStringInfoString(state, "SET ");
				deparseAppendStringInfoString(state, quote_identifier(def_elem->defname));
				deparseAppendStringInfoChar(state, ' ');
				deparseStringLiteral(state, strVal(def_elem->arg));
				break;
			case DEFELEM_ADD:
				deparseAppendStringInfoString(state, "ADD ");
				deparseAppendStringInfoString(state, quote_identifier(def_elem->defname));
				deparseAppendStringInfoChar(state, ' ');
				deparseStringLiteral(state, strVal(def_elem->arg));
				break;
			case DEFELEM_DROP:
				deparseAppendStringInfoString(state, "DROP ");
				deparseAppendStringInfoString(state, quote_identifier(def_elem->defname));
				break;
		}

		if (lnext(options, lc))
			deparseAppendStringInfoString(state, ", ");
	}
	deparseAppendStringInfoString(state, ") ");
}

// "func_name" in gram.y
static void deparseFuncName(DeparseState *state, List *func_name)
{
	ListCell *lc = NULL;

	foreach(lc, func_name)
	{
		deparseAppendStringInfoString(state, quote_identifier(strVal(lfirst(lc))));
		if (lnext(func_name, lc))
			deparseAppendStringInfoChar(state, '.');
	}
}

// "function_with_argtypes" in gram.y
static void deparseFunctionWithArgtypes(DeparseState *state, ObjectWithArgs *object_with_args)
{
	if (object_with_args == NULL)
		elog(ERROR, "deparse: unexpected NULL ObjectWithArgs");

	ListCell *lc;
	deparseFuncName(state, object_with_args->objname);

	if (!object_with_args->args_unspecified)
	{
		deparseAppendStringInfoChar(state, '(');
		List *objargs = object_with_args->objargs;
		if (object_with_args->objfuncargs)
			objargs = object_with_args->objfuncargs;

		foreach(lc, objargs)
		{
			if (IsA(lfirst(lc), FunctionParameter))
				deparseFunctionParameter(state, castNode(FunctionParameter, lfirst(lc)));
			else
				deparseTypeName(state, castNode(TypeName, lfirst(lc)));
			if (lnext(objargs, lc))
				deparseAppendStringInfoString(state, ", ");
		}
		deparseAppendStringInfoChar(state, ')');
	}
}

// "function_with_argtypes_list" in gram.y
static void deparseFunctionWithArgtypesList(DeparseState *state, List *l)
{
	ListCell *lc;

	foreach(lc, l)
	{
		deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, lfirst(lc)));
		if (lnext(l, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "operator_with_argtypes" in gram.y
static void deparseOperatorWithArgtypes(DeparseState *state, ObjectWithArgs *object_with_args)
{
	if (object_with_args == NULL)
		elog(ERROR, "deparse: unexpected NULL ObjectWithArgs");

	deparseAnyOperator(state, object_with_args->objname);

	Assert(list_length(object_with_args->objargs) == 2);
	deparseAppendStringInfoChar(state, '(');
	if (linitial(object_with_args->objargs) == NULL)
		deparseAppendStringInfoString(state, "NONE");
	else
		deparseTypeName(state, castNode(TypeName, linitial(object_with_args->objargs)));
	deparseAppendStringInfoString(state, ", ");
	if (lsecond(object_with_args->objargs) == NULL)
		deparseAppendStringInfoString(state, "NONE");
	else
		deparseTypeName(state, castNode(TypeName, lsecond(object_with_args->objargs)));
	deparseAppendStringInfoChar(state, ')');
}

// "aggr_args" in gram.y
static void deparseAggrArgs(DeparseState *state, List *aggr_args)
{
	Assert(list_length(aggr_args) == 2);

	ListCell *lc = NULL;
	List *args = linitial(aggr_args);
	int order_by_pos = intVal(lsecond(aggr_args));

	deparseAppendStringInfoChar(state, '(');
	if (args == NULL)
	{
		deparseAppendStringInfoChar(state, '*');
	}
	else
	{
		foreach(lc, args)
		{
			if (foreach_current_index(lc) == order_by_pos)
			{
				if (foreach_current_index(lc) > 0)
					deparseAppendStringInfoChar(state, ' ');
				deparseAppendStringInfoString(state, "ORDER BY ");
			}
			else if (foreach_current_index(lc) > 0)
			{
				deparseAppendStringInfoString(state, ", ");
			}

			deparseFunctionParameter(state, castNode(FunctionParameter, lfirst(lc)));
		}

		// Repeat the last direct arg as a ordered arg to handle the
		// simplification done by makeOrderedSetArgs in gram.y
		if (order_by_pos == list_length(args))
		{
			deparseAppendStringInfoString(state, " ORDER BY ");
			deparseFunctionParameter(state, castNode(FunctionParameter, llast(args)));
		}
	}
	deparseAppendStringInfoChar(state, ')');
}

// "aggregate_with_argtypes" in gram.y
static void deparseAggregateWithArgtypes(DeparseState *state, ObjectWithArgs *object_with_args)
{
	if (object_with_args == NULL)
		elog(ERROR, "deparse: unexpected NULL ObjectWithArgs");

	ListCell *lc = NULL;

	deparseFuncName(state, object_with_args->objname);

	deparseAppendStringInfoChar(state, '(');
	if (object_with_args->objargs == NULL && object_with_args->objfuncargs == NULL)
	{
		deparseAppendStringInfoChar(state, '*');
	}
	else
	{
		List *objargs = object_with_args->objargs;
		if (object_with_args->objfuncargs)
			objargs = object_with_args->objfuncargs;

		foreach(lc, objargs)
		{
			if (IsA(lfirst(lc), FunctionParameter))
				deparseFunctionParameter(state, castNode(FunctionParameter, lfirst(lc)));
			else
				deparseTypeName(state, castNode(TypeName, lfirst(lc)));
			if (lnext(objargs, lc))
				deparseAppendStringInfoString(state, ", ");
		}
	}
	deparseAppendStringInfoChar(state, ')');
}

// "columnList" in gram.y
static void deparseColumnList(DeparseState *state, List *columns)
{
	ListCell *lc = NULL;
	foreach(lc, columns)
	{
		deparseAppendStringInfoString(state, quote_identifier(strVal(lfirst(lc))));
		if (lnext(columns, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "opt_column_and_period_list" and "optionalPeriodName" in gram.y
static void deparseColumnListWithPeriod(DeparseState *state, List *columns)
{
	ListCell *lc = NULL;
	foreach(lc, columns)
	{
		bool not_last = lnext(columns, lc);
		if (!not_last)
			deparseAppendStringInfoString(state, "PERIOD ");

		deparseAppendStringInfoString(state, quote_identifier(strVal(lfirst(lc))));

		if (not_last)
			deparseAppendStringInfoString(state, ", ");
	}
}

// "OptTemp" in gram.y
//
// Note this method adds a trailing space if a value is output
static void deparseOptTemp(DeparseState *state, char relpersistence)
{
	switch (relpersistence)
	{
		case RELPERSISTENCE_PERMANENT:
			// Default
			break;
		case RELPERSISTENCE_UNLOGGED:
			deparseAppendStringInfoString(state, "UNLOGGED ");
			break;
		case RELPERSISTENCE_TEMP:
			deparseAppendStringInfoString(state, "TEMPORARY ");
			break;
		default:
			Assert(false);
			break;
	}
}

// "relation_expr_list" in gram.y
static void deparseRelationExprList(DeparseState *state, List *relation_exprs)
{
	ListCell *lc = NULL;
	foreach(lc, relation_exprs)
	{
		deparseRangeVar(state, castNode(RangeVar, lfirst(lc)), DEPARSE_NODE_CONTEXT_NONE);
		if (lnext(relation_exprs, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "handler_name" in gram.y
static void deparseHandlerName(DeparseState *state, List *handler_name)
{
	ListCell *lc = NULL;

	foreach(lc, handler_name)
	{
		deparseAppendStringInfoString(state, quote_identifier(strVal(lfirst(lc))));
		if (lnext(handler_name, lc))
			deparseAppendStringInfoChar(state, '.');
	}
}

// "fdw_options" in gram.y
static void deparseFdwOptions(DeparseState *state, List *fdw_options)
{
	ListCell *lc = NULL;

	foreach (lc, fdw_options)
	{
		DefElem *def_elem = castNode(DefElem, lfirst(lc));
		if (strcmp(def_elem->defname, "handler") == 0 && def_elem->arg != NULL)
		{
			deparseAppendStringInfoString(state, "HANDLER ");
			deparseHandlerName(state, castNode(List, def_elem->arg));
		}
		else if (strcmp(def_elem->defname, "handler") == 0 && def_elem->arg == NULL)
		{
			deparseAppendStringInfoString(state, "NO HANDLER ");
		}
		else if (strcmp(def_elem->defname, "validator") == 0 && def_elem->arg != NULL)
		{
			deparseAppendStringInfoString(state, "VALIDATOR ");
			deparseHandlerName(state, castNode(List, def_elem->arg));
		}
		else if (strcmp(def_elem->defname, "validator") == 0 && def_elem->arg == NULL)
		{
			deparseAppendStringInfoString(state, "NO VALIDATOR ");
		}
		else
		{
			Assert(false);
		}

		if (lnext(fdw_options, lc))
			deparseAppendStringInfoChar(state, ' ');
	}
}

// "type_list" in gram.y
static void deparseTypeList(DeparseState *state, List *type_list)
{
	ListCell *lc = NULL;
	foreach(lc, type_list)
	{
		deparseTypeName(state, castNode(TypeName, lfirst(lc)));
		if (lnext(type_list, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "opt_boolean_or_string" in gram.y
static void deparseOptBooleanOrString(DeparseState *state, char *s)
{
	if (s == NULL)
		return; // No value set
	else if (strcmp(s, "true") == 0)
		deparseAppendStringInfoString(state, "TRUE");
	else if (strcmp(s, "false") == 0)
		deparseAppendStringInfoString(state, "FALSE");
	else if (strcmp(s, "on") == 0)
		deparseAppendStringInfoString(state, "ON");
	else if (strcmp(s, "off") == 0)
		deparseAppendStringInfoString(state, "OFF");
	else
		deparseNonReservedWordOrSconst(state, s);
}

static void deparseOptBoolean(DeparseState *state, Node *node)
{
	if (node == NULL)
	{
		return;
	}

	switch (nodeTag(node))
	{
		case T_String:
			deparseAppendStringInfo(state, " %s", strVal(node));
			break;
		case T_Integer:
			deparseAppendStringInfo(state, " %d", intVal(node));
			break;
		case T_Boolean:
			deparseAppendStringInfo(state, " %s", boolVal(node) ? "TRUE" : "FALSE");
			break;
		default:
			Assert(false);
			break;
	}
}

bool optBooleanValue(Node *node)
{
	if (node == NULL)
	{
		return true;
	}

	switch (nodeTag(node))
	{
		case T_String: {
			// Longest valid string is "off\0"
			char lower[4];
			strlcpy(lower, strVal(node), sizeof(lower));

			if (strcmp(lower, "on") == 0) {
				return true;
			} else if (strcmp(lower, "off") == 0) {
				return false;
			}

			// No sane way to handle this.
			return false;
		}
		case T_Integer:
			return intVal(node) != 0;
		case T_Boolean:
			return boolVal(node);
		default:
			Assert(false);
			return false;
	}
}

// "var_name"
//
// Note this is kept separate from ColId in case we want to improve the
// output of namespaced variable names
static void deparseVarName(DeparseState *state, char *s)
{
	deparseColId(state, s);
}

// "var_list"
static void deparseVarList(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	foreach(lc, l)
	{
		if (IsA(lfirst(lc), ParamRef))
		{
			deparseParamRef(state, castNode(ParamRef, lfirst(lc)));
		}
		else if (IsA(lfirst(lc), A_Const))
		{
			A_Const *a_const = castNode(A_Const, lfirst(lc));
			if (IsA(&a_const->val, Integer) || IsA(&a_const->val, Float))
				deparseNumericOnly(state, (union ValUnion *) &a_const->val);
			else if (IsA(&a_const->val, String))
				deparseOptBooleanOrString(state, strVal(&a_const->val));
			else
				Assert(false);
		}
		else if (IsA(lfirst(lc), TypeCast))
		{
			deparseTypeCast(state, castNode(TypeCast, lfirst(lc)), DEPARSE_NODE_CONTEXT_SET_STATEMENT);
		}
		else
		{
			Assert(false);
		}

		if (lnext(l, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "transaction_mode_list" in gram.y
static void deparseTransactionModeList(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	foreach (lc, l)
	{
		DefElem *def_elem = castNode(DefElem, lfirst(lc));

		if (strcmp(def_elem->defname, "transaction_isolation") == 0)
		{
			char *s = strVal(&castNode(A_Const, def_elem->arg)->val);
			deparseAppendStringInfoString(state, "ISOLATION LEVEL ");
			if (strcmp(s, "read uncommitted") == 0)
				deparseAppendStringInfoString(state, "READ UNCOMMITTED");
			else if (strcmp(s, "read committed") == 0)
				deparseAppendStringInfoString(state, "READ COMMITTED");
			else if (strcmp(s, "repeatable read") == 0)
				deparseAppendStringInfoString(state, "REPEATABLE READ");
			else if (strcmp(s, "serializable") == 0)
				deparseAppendStringInfoString(state, "SERIALIZABLE");
			else
				Assert(false);
		}
		else if (strcmp(def_elem->defname, "transaction_read_only") == 0 && intVal(&castNode(A_Const, def_elem->arg)->val) == 1)
		{
			deparseAppendStringInfoString(state, "READ ONLY");
		}
		else if (strcmp(def_elem->defname, "transaction_read_only") == 0 && intVal(&castNode(A_Const, def_elem->arg)->val) == 0)
		{
			deparseAppendStringInfoString(state, "READ WRITE");
		}
		else if (strcmp(def_elem->defname, "transaction_deferrable") == 0 && intVal(&castNode(A_Const, def_elem->arg)->val) == 1)
		{
			deparseAppendStringInfoString(state, "DEFERRABLE");
		}
		else if (strcmp(def_elem->defname, "transaction_deferrable") == 0 && intVal(&castNode(A_Const, def_elem->arg)->val) == 0)
		{
			deparseAppendStringInfoString(state, "NOT DEFERRABLE");
		}
		else
		{
			Assert(false);
		}

		if (lnext(l, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "alter_identity_column_option_list" in gram.y
static void deparseAlterIdentityColumnOptionList(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	foreach (lc, l)
	{
		DefElem *def_elem = castNode(DefElem, lfirst(lc));
		if (strcmp(def_elem->defname, "restart") == 0 && def_elem->arg == NULL)
		{
			deparseAppendStringInfoString(state, "RESTART");
		}
		else if (strcmp(def_elem->defname, "restart") == 0 && def_elem->arg != NULL)
		{
			deparseAppendStringInfoString(state, "RESTART ");
			deparseNumericOnly(state, (union ValUnion *) def_elem->arg);
		}
		else if (strcmp(def_elem->defname, "generated") == 0)
		{
			deparseAppendStringInfoString(state, "SET GENERATED ");
			if (def_elem->arg == NULL)
				elog(ERROR, "deparse: missing argument for identity generation specification");
			if (intVal(def_elem->arg) == ATTRIBUTE_IDENTITY_ALWAYS)
				deparseAppendStringInfoString(state, "ALWAYS");
			else if (intVal(def_elem->arg) == ATTRIBUTE_IDENTITY_BY_DEFAULT)
				deparseAppendStringInfoString(state, "BY DEFAULT");
			else
				Assert(false);
		}
		else
		{
			deparseAppendStringInfoString(state, "SET ");
			deparseSeqOptElem(state, def_elem);
		}
		if (lnext(l, lc))
			deparseAppendStringInfoChar(state, ' ');
	}
}

// "reloptions" in gram.y
static void deparseRelOptions(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	deparseAppendStringInfoChar(state, '(');
	foreach(lc, l)
	{
		DefElem *def_elem = castNode(DefElem, lfirst(lc));
		if (def_elem->defnamespace != NULL)
		{
			deparseAppendStringInfoString(state, quote_identifier(def_elem->defnamespace));
			deparseAppendStringInfoChar(state, '.');
		}
		if (def_elem->defname != NULL)
			deparseAppendStringInfoString(state, quote_identifier(def_elem->defname));
		if (def_elem->defname != NULL && def_elem->arg != NULL)
			deparseAppendStringInfoChar(state, '=');
		if (def_elem->arg != NULL)
			deparseDefArg(state, def_elem->arg, false);

		if (lnext(l, lc))
			deparseAppendStringInfoString(state, ", ");
	}
	deparseAppendStringInfoChar(state, ')');
}

// "OptWith" and "opt_reloptions" in gram.y
//
// Note this method adds a trailing space if a value is output
static void deparseOptWith(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	if (list_length(l) > 0)
	{
		deparseAppendStringInfoString(state, "WITH ");
		deparseRelOptions(state, l);
		deparseAppendStringInfoChar(state, ' ');
	}
}

// "target_list" and "opt_target_list" in gram.y
static void deparseTargetList(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	foreach(lc, l)
	{
		ResTarget *res_target = castNode(ResTarget, lfirst(lc));

		if (res_target->val == NULL)
			elog(ERROR, "deparse: error in deparseTargetList: ResTarget without val");
		else if (IsA(res_target->val, ColumnRef))
			deparseColumnRef(state, castNode(ColumnRef, res_target->val));
		else
			deparseExpr(state, res_target->val, DEPARSE_NODE_CONTEXT_A_EXPR);

		if (res_target->name != NULL) {
			deparseAppendStringInfoString(state, " AS ");
			deparseAppendStringInfoString(state, quote_identifier(res_target->name));
		}

		if (lnext(l, lc))
			deparseAppendCommaAndPart(state);
	}
}

// "insert_column_list" in gram.y
static void deparseInsertColumnList(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	foreach(lc, l)
	{
		ResTarget *res_target = castNode(ResTarget, lfirst(lc));
		Assert(res_target->name != NULL);
		deparseAppendStringInfoString(state, quote_identifier(res_target->name));
		deparseOptIndirection(state, res_target->indirection, 0);
		if (lnext(l, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "xml_attribute_list" in gram.y
static void deparseXmlAttributeList(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	foreach(lc, l)
	{
		ResTarget *res_target = castNode(ResTarget, lfirst(lc));
		Assert(res_target->val != NULL);

		deparseExpr(state, res_target->val, DEPARSE_NODE_CONTEXT_A_EXPR);

		if (res_target->name != NULL)
		{
			deparseAppendStringInfoString(state, " AS ");
			deparseAppendStringInfoString(state, quote_identifier(res_target->name));
		}

		if (lnext(l, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "xml_namespace_list" in gram.y
static void deparseXmlNamespaceList(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	foreach(lc, l)
	{
		ResTarget *res_target = castNode(ResTarget, lfirst(lc));
		Assert(res_target->val != NULL);

		if (res_target->name == NULL)
			deparseAppendStringInfoString(state, "DEFAULT ");

		deparseBExpr(state, res_target->val);

		if (res_target->name != NULL)
		{
			deparseAppendStringInfoString(state, " AS ");
			deparseAppendStringInfoString(state, quote_identifier(res_target->name));
		}

		if (lnext(l, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "table_ref" in gram.y
static void deparseTableRef(DeparseState *state, Node *node)
{
	if (node == NULL)
		elog(ERROR, "deparse: unexpected NULL node");

	switch (nodeTag(node))
	{
		case T_RangeVar:
			deparseRangeVar(state, castNode(RangeVar, node), DEPARSE_NODE_CONTEXT_NONE);
			break;
		case T_RangeTableSample:
			deparseRangeTableSample(state, castNode(RangeTableSample, node));
			break;
		case T_RangeFunction:
			deparseRangeFunction(state, castNode(RangeFunction, node));
			break;
		case T_RangeTableFunc:
			deparseRangeTableFunc(state, castNode(RangeTableFunc, node));
			break;
		case T_RangeSubselect:
			deparseRangeSubselect(state, castNode(RangeSubselect, node));
			break;
		case T_JoinExpr:
			deparseJoinExpr(state, castNode(JoinExpr, node));
			break;
		case T_JsonTable:
			deparseJsonTable(state, castNode(JsonTable, node));
			break;
		default:
			Assert(false);
	}
}

// "from_list" in gram.y
static void deparseFromList(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	foreach(lc, l)
	{
		deparseTableRef(state, lfirst(lc));
		if (lnext(l, lc))
			deparseAppendCommaAndPart(state);
	}
}

// "from_clause" in gram.y
//
// Note this method adds a trailing space if a value is output
static void deparseFromClause(DeparseState *state, List *l)
{
	if (list_length(l) > 0)
	{
		deparseAppendPartGroup(state, "FROM", DEPARSE_PART_INDENT);
		deparseFromList(state, l);
		deparseAppendStringInfoChar(state, ' ');
	}
}

// "where_clause" in gram.y
//
// Note this method adds a trailing space if a value is output
static void deparseWhereClause(DeparseState *state, Node *node)
{
	if (node != NULL)
	{
		deparseAppendPartGroup(state, "WHERE", DEPARSE_PART_INDENT);
		deparseExpr(state, node, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoChar(state, ' ');
	}
}

// "where_or_current_clause" in gram.y
//
// Note this method adds a trailing space if a value is output
static void deparseWhereOrCurrentClause(DeparseState *state, Node *node)
{
	if (node == NULL)
		return;

	deparseAppendPartGroup(state, "WHERE", DEPARSE_PART_INDENT);

	if (IsA(node, CurrentOfExpr)) {
		CurrentOfExpr *current_of_expr = castNode(CurrentOfExpr, node);
		deparseAppendStringInfoString(state, "CURRENT OF ");
		deparseAppendStringInfoString(state, quote_identifier(current_of_expr->cursor_name));
	} else {
		deparseExpr(state, node, DEPARSE_NODE_CONTEXT_A_EXPR);
	}

	deparseAppendStringInfoChar(state, ' ');
}

// "group_by_list" in gram.y
static void deparseGroupByList(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	foreach(lc, l)
	{
		if (IsA(lfirst(lc), GroupingSet))
			deparseGroupingSet(state, castNode(GroupingSet, lfirst(lc)));
		else
			deparseExpr(state, lfirst(lc), DEPARSE_NODE_CONTEXT_A_EXPR);

		if (lnext(l, lc))
			deparseAppendCommaAndPart(state);
	}
}

// "set_target" in gram.y
static void deparseSetTarget(DeparseState *state, ResTarget *res_target)
{
	Assert(res_target->name != NULL);
	deparseColId(state, res_target->name);
	deparseOptIndirection(state, res_target->indirection, 0);
}

// "any_name_list" in gram.y
static void deparseAnyNameList(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	foreach(lc, l)
	{
		deparseAnyName(state, castNode(List, lfirst(lc)));
		if (lnext(l, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "name_list" in gram.y
static void deparseNameList(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	foreach(lc, l)
	{
		deparseColId(state, strVal(lfirst(lc)));
		if (lnext(l, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "opt_sort_clause" and "json_array_aggregate_order_by_clause_opt" in gram.y
//
// Note this method adds a trailing space if a value is output
static void deparseOptSortClause(DeparseState *state, List *l, DeparseNodeContext context)
{
	ListCell *lc = NULL;

	if (list_length(l) > 0)
	{
		if (context == DEPARSE_NODE_CONTEXT_SELECT_SORT_CLAUSE)
			deparseAppendPartGroup(state, "ORDER BY", DEPARSE_PART_INDENT_AND_MERGE);
		else
			deparseAppendStringInfoString(state, "ORDER BY ");

		foreach(lc, l)
		{
			deparseSortBy(state, castNode(SortBy, lfirst(lc)));
			if (lnext(l, lc))
			{
				if (context == DEPARSE_NODE_CONTEXT_SELECT_SORT_CLAUSE)
					deparseAppendCommaAndPart(state);
				else
					deparseAppendStringInfoString(state, ", ");
			}
		}
		deparseAppendStringInfoChar(state, ' ');
	}
}

// "func_arg_expr" in gram.y
static void deparseFuncArgExpr(DeparseState *state, Node *node)
{
	if (IsA(node, NamedArgExpr))
	{
		NamedArgExpr *named_arg_expr = castNode(NamedArgExpr, node);
		deparseAppendStringInfoString(state, named_arg_expr->name);
		deparseAppendStringInfoString(state, " := ");
		deparseExpr(state, (Node *) named_arg_expr->arg, DEPARSE_NODE_CONTEXT_A_EXPR);
	}
	else
	{
		deparseExpr(state, node, DEPARSE_NODE_CONTEXT_A_EXPR);
	}
}

// "set_clause_list" in gram.y
static void deparseSetClauseList(DeparseState *state, List *target_list)
{
	ListCell *lc;
	ListCell *lc2;
	int skip_next_n_elems = 0;

	Assert(list_length(target_list) > 0);

	foreach(lc, target_list)
	{
		if (skip_next_n_elems > 0)
		{
			skip_next_n_elems--;
			continue;
		}

		if (foreach_current_index(lc) != 0)
			deparseAppendCommaAndPart(state);

		ResTarget *res_target = castNode(ResTarget, lfirst(lc));
		if (res_target->val == NULL)
			elog(ERROR, "deparse: unexpected NULL val in SET clause ResTarget");

		if (IsA(res_target->val, MultiAssignRef))
		{
			MultiAssignRef *r = castNode(MultiAssignRef, res_target->val);
			deparseAppendStringInfoString(state, "(");
			for_each_cell(lc2, target_list, lc)
			{
				deparseSetTarget(state, castNode(ResTarget, lfirst(lc2)));
				if ((foreach_current_index(lc2) - foreach_current_index(lc)) == r->ncolumns - 1) // Last element in this multi-assign
					break;
				else if (lnext(target_list, lc2))
					deparseAppendStringInfoString(state, ", ");
			}
			deparseAppendStringInfoString(state, ") = ");
			deparseExpr(state, r->source, DEPARSE_NODE_CONTEXT_A_EXPR);
			skip_next_n_elems = r->ncolumns - 1;
		}
		else
		{
			deparseSetTarget(state, res_target);
			deparseAppendStringInfoString(state, " = ");
			deparseExpr(state, res_target->val, DEPARSE_NODE_CONTEXT_A_EXPR);
		}
	}
}

// "func_expr_windowless" in gram.y
static void deparseFuncExprWindowless(DeparseState *state, Node* node)
{
	if (node == NULL)
		elog(ERROR, "deparse: unexpected NULL node");

	switch (nodeTag(node))
	{
		case T_FuncCall:
			deparseFuncCall(state, castNode(FuncCall, node));
			break;
		case T_SQLValueFunction:
			deparseSQLValueFunction(state, castNode(SQLValueFunction, node));
			break;
		case T_TypeCast:
			deparseTypeCast(state, castNode(TypeCast, node), DEPARSE_NODE_CONTEXT_FUNC_EXPR);
			break;
		case T_CoalesceExpr:
			deparseCoalesceExpr(state, castNode(CoalesceExpr, node));
			break;
		case T_MinMaxExpr:
			deparseMinMaxExpr(state, castNode(MinMaxExpr, node));
			break;
		case T_XmlExpr:
			deparseXmlExpr(state, castNode(XmlExpr, node), DEPARSE_NODE_CONTEXT_NONE /* we don't know which kind of expression */);
			break;
		case T_XmlSerialize:
			deparseXmlSerialize(state, castNode(XmlSerialize, node));
			break;
		default:
			Assert(false);
	}
}

// "opt_collate" in gram.y
//
// Note this method adds a trailing space if a value is output
static void deparseOptCollate(DeparseState *state, List *l)
{
	if (list_length(l) > 0)
	{
		deparseAppendStringInfoString(state, "COLLATE ");
		deparseAnyName(state, l);
		deparseAppendStringInfoChar(state, ' ');
	}
}

// "index_elem" in gram.y
static void deparseIndexElem(DeparseState *state, IndexElem* index_elem)
{
	if (index_elem->name != NULL)
	{
		deparseColId(state, index_elem->name);
		deparseAppendStringInfoChar(state, ' ');
	}
	else if (index_elem->expr != NULL)
	{
		bool is_func_expr = false;

		switch (nodeTag(index_elem->expr))
		{
			// Simple function calls can be written without wrapping parens
			case T_FuncCall: // func_application
			case T_SQLValueFunction: // func_expr_common_subexpr
			case T_CoalesceExpr: // func_expr_common_subexpr
			case T_MinMaxExpr: // func_expr_common_subexpr
			case T_XmlExpr: // func_expr_common_subexpr
			case T_XmlSerialize: // func_expr_common_subexpr
				is_func_expr = true;
				break;
			default:
				break;
		}

		// Some function call nodes are output using operator syntax (e.g. "x AT TIME ZONE y"), and need parens
		if (is_func_expr && !needsParensAsCExpr(index_elem->expr))
		{
			deparseFuncExprWindowless(state, index_elem->expr);
			deparseAppendStringInfoString(state, " ");
		}
		else
		{
			deparseAppendStringInfoChar(state, '(');
			deparseExpr(state, index_elem->expr, DEPARSE_NODE_CONTEXT_A_EXPR);
			deparseAppendStringInfoString(state, ") ");
		}
	}
	else
	{
		Assert(false);
	}

	deparseOptCollate(state, index_elem->collation);

	if (list_length(index_elem->opclass) > 0)
	{
		deparseAnyName(state, index_elem->opclass);

		if (list_length(index_elem->opclassopts) > 0)
			deparseRelOptions(state, index_elem->opclassopts);

		deparseAppendStringInfoChar(state, ' ');
	}

	switch (index_elem->ordering)
	{
		case SORTBY_DEFAULT:
			// Default
			break;
		case SORTBY_ASC:
			deparseAppendStringInfoString(state, "ASC ");
			break;
		case SORTBY_DESC:
			deparseAppendStringInfoString(state, "DESC ");
			break;
		case SORTBY_USING:
			// Not allowed in CREATE INDEX
			Assert(false);
			break;
	}

	switch (index_elem->nulls_ordering)
	{
		case SORTBY_NULLS_DEFAULT:
			// Default
			break;
		case SORTBY_NULLS_FIRST:
			deparseAppendStringInfoString(state, "NULLS FIRST ");
			break;
		case SORTBY_NULLS_LAST:
			deparseAppendStringInfoString(state, "NULLS LAST ");
			break;
	}

	removeTrailingSpace(state);
}

// "qualified_name_list" in gram.y
static void deparseQualifiedNameList(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	foreach(lc, l)
	{
		deparseRangeVar(state, castNode(RangeVar, lfirst(lc)), DEPARSE_NODE_CONTEXT_NONE);
		if (lnext(l, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "OptInherit" in gram.y
//
// Note this method adds a trailing space if a value is output
static void deparseOptInherit(DeparseState *state, List *l)
{
	if (list_length(l) > 0)
	{
		deparseAppendStringInfoString(state, "INHERITS (");
		deparseQualifiedNameList(state, l);
		deparseAppendStringInfoString(state, ") ");
	}
}

// "privilege_target" in gram.y
static void deparsePrivilegeTarget(DeparseState *state, GrantTargetType targtype, ObjectType	objtype, List *objs)
{
	switch (targtype)
	{
		case ACL_TARGET_OBJECT:
			switch (objtype)
			{
				case OBJECT_TABLE:
					deparseQualifiedNameList(state, objs);
					break;
				case OBJECT_SEQUENCE:
					deparseAppendStringInfoString(state, "SEQUENCE ");
					deparseQualifiedNameList(state, objs);
					break;
				case OBJECT_FDW:
					deparseAppendStringInfoString(state, "FOREIGN DATA WRAPPER ");
					deparseNameList(state, objs);
					break;
				case OBJECT_FOREIGN_SERVER:
					deparseAppendStringInfoString(state, "FOREIGN SERVER ");
					deparseNameList(state, objs);
					break;
				case OBJECT_FUNCTION:
					deparseAppendStringInfoString(state, "FUNCTION ");
					deparseFunctionWithArgtypesList(state, objs);
					break;
				case OBJECT_PROCEDURE:
					deparseAppendStringInfoString(state, "PROCEDURE ");
					deparseFunctionWithArgtypesList(state, objs);
					break;
				case OBJECT_ROUTINE:
					deparseAppendStringInfoString(state, "ROUTINE ");
					deparseFunctionWithArgtypesList(state, objs);
					break;
				case OBJECT_DATABASE:
					deparseAppendStringInfoString(state, "DATABASE ");
					deparseNameList(state, objs);
					break;
				case OBJECT_DOMAIN:
					deparseAppendStringInfoString(state, "DOMAIN ");
					deparseAnyNameList(state, objs);
					break;
				case OBJECT_LANGUAGE:
					deparseAppendStringInfoString(state, "LANGUAGE ");
					deparseNameList(state, objs);
					break;
				case OBJECT_LARGEOBJECT:
					deparseAppendStringInfoString(state, "LARGE OBJECT ");
					deparseNumericOnlyList(state, objs);
					break;
				case OBJECT_SCHEMA:
					deparseAppendStringInfoString(state, "SCHEMA ");
					deparseNameList(state, objs);
					break;
				case OBJECT_TABLESPACE:
					deparseAppendStringInfoString(state, "TABLESPACE ");
					deparseNameList(state, objs);
					break;
				case OBJECT_TYPE:
					deparseAppendStringInfoString(state, "TYPE ");
					deparseAnyNameList(state, objs);
					break;
				default:
					// Other types are not supported here
					Assert(false);
					break;
			}
			break;
		case ACL_TARGET_ALL_IN_SCHEMA:
			switch (objtype)
			{
				case OBJECT_TABLE:
					deparseAppendStringInfoString(state, "ALL TABLES IN SCHEMA ");
					deparseNameList(state, objs);
					break;
				case OBJECT_SEQUENCE:
					deparseAppendStringInfoString(state, "ALL SEQUENCES IN SCHEMA ");
					deparseNameList(state, objs);
					break;
				case OBJECT_FUNCTION:
					deparseAppendStringInfoString(state, "ALL FUNCTIONS IN SCHEMA ");
					deparseNameList(state, objs);
					break;
				case OBJECT_PROCEDURE:
					deparseAppendStringInfoString(state, "ALL PROCEDURES IN SCHEMA ");
					deparseNameList(state, objs);
					break;
				case OBJECT_ROUTINE:
					deparseAppendStringInfoString(state, "ALL ROUTINES IN SCHEMA ");
					deparseNameList(state, objs);
					break;
				default:
					// Other types are not supported here
					Assert(false);
					break;
			}
			break;
		case ACL_TARGET_DEFAULTS: // defacl_privilege_target
			switch (objtype)
			{
				case OBJECT_TABLE:
					deparseAppendStringInfoString(state, "TABLES");
					break;
				case OBJECT_FUNCTION:
					deparseAppendStringInfoString(state, "FUNCTIONS");
					break;
				case OBJECT_SEQUENCE:
					deparseAppendStringInfoString(state, "SEQUENCES");
					break;
				case OBJECT_TYPE:
					deparseAppendStringInfoString(state, "TYPES");
					break;
				case OBJECT_SCHEMA:
					deparseAppendStringInfoString(state, "SCHEMAS");
					break;
				case OBJECT_LARGEOBJECT:
					deparseAppendStringInfoString(state, "LARGE OBJECTS");
					break;
				default:
					// Other types are not supported here
					Assert(false);
					break;
			}
			break;
	}
}

// "opclass_item_list" in gram.y
static void deparseOpclassItemList(DeparseState *state, List *items)
{
	ListCell *lc = NULL;

	foreach (lc, items)
	{
		deparseCreateOpClassItem(state, castNode(CreateOpClassItem, lfirst(lc)));
		if (lnext(items, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

// "createdb_opt_list" in gram.y
static void deparseCreatedbOptList(DeparseState *state, List *l)
{
	ListCell *lc = NULL;

	foreach (lc, l)
	{
		DefElem *def_elem = castNode(DefElem, lfirst(lc));
		if (strcmp(def_elem->defname, "connection_limit") == 0)
			deparseAppendStringInfoString(state, "CONNECTION LIMIT");
		else
			deparseGenericDefElemName(state, def_elem->defname);

		deparseAppendStringInfoChar(state, ' ');

		if (def_elem->arg == NULL)
			deparseAppendStringInfoString(state, "DEFAULT");
		else if (IsA(def_elem->arg, Integer))
			deparseSignedIconst(state, def_elem->arg);
		else if (IsA(def_elem->arg, String))
			deparseOptBooleanOrString(state, strVal(def_elem->arg));

		if (lnext(l, lc))
			deparseAppendStringInfoChar(state, ' ');
	}
}

// "utility_option_list" in gram.y
static void deparseUtilityOptionList(DeparseState *state, List *options)
{
	ListCell *lc = NULL;
	char *defname = NULL;

	if (list_length(options) > 0)
	{
		deparseAppendStringInfoChar(state, '(');
		foreach(lc, options)
		{
			DefElem *def_elem = castNode(DefElem, lfirst(lc));
			deparseGenericDefElemName(state, def_elem->defname);

			if (def_elem->arg != NULL)
			{
				deparseAppendStringInfoChar(state, ' ');
				if (IsA(def_elem->arg, Integer) || IsA(def_elem->arg, Float))
					deparseNumericOnly(state, (union ValUnion *) def_elem->arg);
				else if (IsA(def_elem->arg, String))
					deparseOptBooleanOrString(state, strVal(def_elem->arg));
				else
					Assert(false);
			}

			if (lnext(options, lc))
				deparseAppendStringInfoString(state, ", ");
		}
		deparseAppendStringInfoString(state, ") ");
	}
}

static void deparseSelectStmt(DeparseState *state, SelectStmt *stmt, DeparseNodeContext context)
{
	if (stmt == NULL)
		elog(ERROR, "deparse: unexpected NULL SelectStmt");

	const ListCell *lc = NULL;
	const ListCell *lc2 = NULL;
	bool need_parens = context == DEPARSE_NODE_CONTEXT_SELECT_SETOP && (
		list_length(stmt->sortClause) > 0 ||
		stmt->limitOffset != NULL ||
		stmt->limitCount != NULL ||
		list_length(stmt->lockingClause) > 0 ||
		stmt->withClause != NULL ||
		stmt->op != SETOP_NONE);
	DeparseStateNestingLevel *parent_level = NULL;

	/* Set operations (UNION etc) recurse here directly, without going through deparseExpr */
	check_stack_depth();

	if (need_parens)
	{
		deparseAppendPart(state, true);
		deparseAppendStringInfoChar(state, '(');
	}
	if (need_parens || (context != DEPARSE_NODE_CONTEXT_SELECT_SETOP && context != DEPARSE_NODE_CONTEXT_INSERT_SELECT))
		parent_level = deparseStateIncreaseNestingLevel(state);

	if (stmt->withClause)
	{
		deparseWithClause(state, stmt->withClause);
		deparseAppendStringInfoChar(state, ' ');
	}

	switch (stmt->op) {
		case SETOP_NONE:
			if (list_length(stmt->valuesLists) > 0)
			{
				const ListCell *lc;
				deparseAppendPartGroup(state, "VALUES", DEPARSE_PART_INDENT_AND_MERGE);

				foreach(lc, stmt->valuesLists)
				{
					deparseAppendStringInfoChar(state, '(');
					deparseExprList(state, lfirst(lc));
					deparseAppendStringInfoChar(state, ')');
					if (lnext(stmt->valuesLists, lc))
						deparseAppendCommaAndPart(state);
				}
				deparseAppendStringInfoChar(state, ' ');
				break;
			}

			deparseAppendPartGroup(state, "SELECT", DEPARSE_PART_INDENT_AND_MERGE);

			if (list_length(stmt->targetList) > 0)
			{
				if (stmt->distinctClause != NULL)
				{
					deparseMarkCurrentPartNonMergable(state);
					deparseAppendStringInfoString(state, "DISTINCT ");

					if (list_length(stmt->distinctClause) > 0 && linitial(stmt->distinctClause) != NULL)
					{
						deparseAppendStringInfoString(state, "ON (");
						deparseExprList(state, stmt->distinctClause);
						deparseAppendStringInfoString(state, ") ");
					}
					deparseAppendPart(state, true);
				}

				deparseTargetList(state, stmt->targetList);
				deparseAppendStringInfoChar(state, ' ');
			}

			if (stmt->intoClause != NULL)
			{
				if (stmt->intoClause->rel == NULL)
					elog(ERROR, "deparse: unexpected NULL rel in IntoClause");
				deparseAppendPartGroup(state, "INTO", DEPARSE_PART_INDENT);
				deparseOptTemp(state, stmt->intoClause->rel->relpersistence);
				deparseIntoClause(state, stmt->intoClause);
				deparseAppendStringInfoChar(state, ' ');
			}

			deparseFromClause(state, stmt->fromClause);
			deparseWhereClause(state, stmt->whereClause);

			if (list_length(stmt->groupClause) > 0)
			{
				deparseAppendPartGroup(state, "GROUP BY", DEPARSE_PART_INDENT_AND_MERGE);
				if (stmt->groupDistinct)
					deparseAppendStringInfoString(state, "DISTINCT ");
				deparseGroupByList(state, stmt->groupClause);
				deparseAppendStringInfoChar(state, ' ');
			}

			if (stmt->havingClause != NULL)
			{
				deparseAppendPartGroup(state, "HAVING", DEPARSE_PART_INDENT);
				deparseExpr(state, stmt->havingClause, DEPARSE_NODE_CONTEXT_A_EXPR);
				deparseAppendStringInfoChar(state, ' ');
			}

			if (stmt->windowClause != NULL)
			{
				deparseAppendPartGroup(state, "WINDOW", DEPARSE_PART_INDENT);
				foreach(lc, stmt->windowClause)
				{
					WindowDef *window_def = castNode(WindowDef, lfirst(lc));
					Assert(window_def->name != NULL);
					deparseAppendStringInfoString(state, window_def->name);
					deparseAppendStringInfoString(state, " AS ");
					deparseWindowDef(state, window_def);
					if (lnext(stmt->windowClause, lc))
						deparseAppendStringInfoString(state, ", ");
				}
				deparseAppendStringInfoChar(state, ' ');
			}
			break;
		case SETOP_UNION:
		case SETOP_INTERSECT:
		case SETOP_EXCEPT:
			{
				deparseSelectStmt(state, stmt->larg, DEPARSE_NODE_CONTEXT_SELECT_SETOP);
				switch (stmt->op)
				{
					case SETOP_UNION:
						if (stmt->all)
							deparseAppendPartGroup(state, "UNION ALL", DEPARSE_PART_NO_INDENT);
						else
							deparseAppendPartGroup(state, "UNION", DEPARSE_PART_NO_INDENT);
						break;
					case SETOP_INTERSECT:
						if (stmt->all)
							deparseAppendPartGroup(state, "INTERSECT ALL", DEPARSE_PART_NO_INDENT);
						else
							deparseAppendPartGroup(state, "INTERSECT", DEPARSE_PART_NO_INDENT);
						break;
					case SETOP_EXCEPT:
						if (stmt->all)
							deparseAppendPartGroup(state, "EXCEPT ALL", DEPARSE_PART_NO_INDENT);
						else
							deparseAppendPartGroup(state, "EXCEPT", DEPARSE_PART_NO_INDENT);
						break;
					default:
						Assert(false);
				}
				deparseAppendPart(state, true);
				deparseSelectStmt(state, stmt->rarg, DEPARSE_NODE_CONTEXT_SELECT_SETOP);
				deparseAppendStringInfoChar(state, ' ');
			}
			break;
	}

	deparseOptSortClause(state, stmt->sortClause, DEPARSE_NODE_CONTEXT_SELECT_SORT_CLAUSE);

	if (stmt->limitCount != NULL)
	{
		if (stmt->limitOption == LIMIT_OPTION_COUNT)
			deparseAppendPartGroup(state, "LIMIT", DEPARSE_PART_INDENT);
		else if (stmt->limitOption == LIMIT_OPTION_WITH_TIES)
			deparseAppendStringInfoString(state, "FETCH FIRST ");

		if (IsA(stmt->limitCount, A_Const) && castNode(A_Const, stmt->limitCount)->isnull)
			deparseAppendStringInfoString(state, "ALL");
		else if (stmt->limitOption == LIMIT_OPTION_WITH_TIES)
			deparseCExpr(state, stmt->limitCount);
		else
			deparseExpr(state, stmt->limitCount, DEPARSE_NODE_CONTEXT_A_EXPR);

		deparseAppendStringInfoChar(state, ' ');

		if (stmt->limitOption == LIMIT_OPTION_WITH_TIES)
			deparseAppendStringInfoString(state, "ROWS WITH TIES ");
	}

	if (stmt->limitOffset != NULL)
	{
		deparseAppendPartGroup(state, "OFFSET", DEPARSE_PART_INDENT);
		deparseExpr(state, stmt->limitOffset, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (list_length(stmt->lockingClause) > 0)
	{
		foreach(lc, stmt->lockingClause)
		{
			deparseLockingClause(state, castNode(LockingClause, lfirst(lc)));
			if (lnext(stmt->lockingClause, lc))
				deparseAppendStringInfoString(state, " ");
		}
		deparseAppendStringInfoChar(state, ' ');
	}

	removeTrailingSpace(state);

	/* Note that parent_level can be NULL, so we repeat the full if condition here */
	if (need_parens || (context != DEPARSE_NODE_CONTEXT_SELECT_SETOP && context != DEPARSE_NODE_CONTEXT_INSERT_SELECT))
		deparseStateDecreaseNestingLevel(state, parent_level);
	if (need_parens)
		deparseAppendStringInfoChar(state, ')');
}

static void deparseIntoClause(DeparseState *state, IntoClause *into_clause)
{
	ListCell *lc;

	deparseRangeVar(state, into_clause->rel, DEPARSE_NODE_CONTEXT_NONE); /* target relation name */

	if (list_length(into_clause->colNames) > 0)
	{
		deparseAppendStringInfoChar(state, '(');
		deparseColumnList(state, into_clause->colNames);
		deparseAppendStringInfoChar(state, ')');
	}
	deparseAppendStringInfoChar(state, ' ');

	if (into_clause->accessMethod != NULL)
	{
		deparseAppendStringInfoString(state, "USING ");
		deparseAppendStringInfoString(state, quote_identifier(into_clause->accessMethod));
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseOptWith(state, into_clause->options);

	switch (into_clause->onCommit)
	{
		case ONCOMMIT_NOOP:
			// No clause
			break;
		case ONCOMMIT_PRESERVE_ROWS:
			deparseAppendStringInfoString(state, "ON COMMIT PRESERVE ROWS ");
			break;
		case ONCOMMIT_DELETE_ROWS:
			deparseAppendStringInfoString(state, "ON COMMIT DELETE ROWS ");
			break;
		case ONCOMMIT_DROP:
			deparseAppendStringInfoString(state, "ON COMMIT DROP ");
			break;
	}

	if (into_clause->tableSpaceName != NULL)
	{
		deparseAppendStringInfoString(state, "TABLESPACE ");
		deparseAppendStringInfoString(state, quote_identifier(into_clause->tableSpaceName));
		deparseAppendStringInfoChar(state, ' ');
	}

	removeTrailingSpace(state);
}

static void deparseRangeVar(DeparseState *state, RangeVar *range_var, DeparseNodeContext context)
{
	if (range_var == NULL)
		elog(ERROR, "deparse: unexpected NULL RangeVar");

	if (!range_var->inh && context != DEPARSE_NODE_CONTEXT_CREATE_TYPE && context != DEPARSE_NODE_CONTEXT_ALTER_TYPE)
		deparseAppendStringInfoString(state, "ONLY ");

	if (range_var->catalogname != NULL)
	{
		deparseAppendStringInfoString(state, quote_identifier(range_var->catalogname));
		deparseAppendStringInfoChar(state, '.');
	}

	if (range_var->schemaname != NULL)
	{
		deparseAppendStringInfoString(state, quote_identifier(range_var->schemaname));
		deparseAppendStringInfoChar(state, '.');
	}

	Assert(range_var->relname != NULL);
	deparseAppendStringInfoString(state, quote_identifier(range_var->relname));
	deparseAppendStringInfoChar(state, ' ');

	if (range_var->alias != NULL)
	{
		if (context == DEPARSE_NODE_CONTEXT_INSERT_RELATION)
			deparseAppendStringInfoString(state, "AS ");
		deparseAlias(state, range_var->alias);
		deparseAppendStringInfoChar(state, ' ');
	}

	removeTrailingSpace(state);
}

void deparseRawStmt(StringInfo str, struct RawStmt *raw_stmt)
{
	PostgresDeparseOpts opts;
	MemSet(&opts, 0, sizeof(PostgresDeparseOpts)); // zero initialized means pretty print = false
	deparseRawStmtOpts(str, raw_stmt, &opts);
}

void deparseRawStmtOpts(StringInfo str, struct RawStmt *raw_stmt, PostgresDeparseOpts *opts)
{
	DeparseState *state = NULL;
	if (raw_stmt->stmt == NULL)
		elog(ERROR, "deparse error in deparseRawStmt: RawStmt with empty Stmt");

	state = palloc0(sizeof(DeparseState));
	state->opts = opts;
	if (state->opts->indent_size == 0)
		state->opts->indent_size = 4;
	if (state->opts->max_line_length == 0)
		state->opts->max_line_length = 80;

	/* Check for any comments at the start of the statement */
	if (state->opts->comment_count > 0)
	{
		deparseStateIncreaseNestingLevel(state);
		deparseAppendCommentsIfNeeded(state, raw_stmt->stmt_location);
		deparseRemoveTrailingEmptyPart(state);
		deparseStateDecreaseNestingLevel(state, NULL);
	}

	deparseStmt(state, raw_stmt->stmt);

	/* Check for any comments at the end of the statement */
	if (state->opts->comment_count > 0)
	{
		deparseStateIncreaseNestingLevel(state);
		deparseAppendCommentsIfNeeded(state, INT_MAX);
		deparseRemoveTrailingEmptyPart(state);
		deparseStateDecreaseNestingLevel(state, NULL);
	}

	deparseEmit(state, str);

	bms_free(state->emitted_comments);
	pfree(state);
}

static void deparseAlias(DeparseState *state, Alias *alias)
{
	deparseAppendStringInfoString(state, quote_identifier(alias->aliasname));

	if (list_length(alias->colnames) > 0)
	{
		const ListCell *lc = NULL;
		deparseAppendStringInfoChar(state, '(');
		deparseNameList(state, alias->colnames);
		deparseAppendStringInfoChar(state, ')');
	}
}

static void deparseAConst(DeparseState *state, A_Const *a_const)
{
	union ValUnion *val = a_const->isnull ? NULL : &a_const->val;
	deparseAppendCommentsIfNeeded(state, a_const->location);
	deparseValue(state, val, DEPARSE_NODE_CONTEXT_CONSTANT);
}

static void deparseFuncCall(DeparseState *state, FuncCall *func_call)
{
	if (func_call == NULL)
		elog(ERROR, "deparse: unexpected NULL FuncCall");

	const ListCell *lc = NULL;

	Assert(list_length(func_call->funcname) > 0);

	if (list_length(func_call->funcname) == 2 &&
		strcmp(strVal(linitial(func_call->funcname)), "pg_catalog") == 0 &&
		strcmp(strVal(lsecond(func_call->funcname)), "overlay") == 0 &&
		list_length(func_call->args) == 4)
	{
		/*
		 * Note that this is a bit odd, but "OVERLAY" is a keyword on its own merit, and only accepts the
		 * keyword parameter style when its called as a keyword, not as a regular function (i.e. pg_catalog.overlay)
		 */
		deparseAppendStringInfoString(state, "OVERLAY(");
		deparseExpr(state, linitial(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, " PLACING ");
		deparseExpr(state, lsecond(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, " FROM ");
		deparseExpr(state, lthird(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, " FOR ");
		deparseExpr(state, lfourth(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoChar(state, ')');
		return;
	} else if (func_call->funcformat == COERCE_SQL_SYNTAX &&
		list_length(func_call->funcname) == 2 &&
		strcmp(strVal(linitial(func_call->funcname)), "pg_catalog") == 0 &&
		strcmp(strVal(lsecond(func_call->funcname)), "substring") == 0)
	{
		/*
		 * "SUBSTRING" is a keyword on its own merit, and only accepts the
		 * keyword parameter style when its called as a keyword, not as a regular function (i.e. pg_catalog.substring)
		 */
		Assert(list_length(func_call->args) == 2 || list_length(func_call->args) == 3);
		deparseAppendStringInfoString(state, "SUBSTRING(");
		deparseExpr(state, linitial(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, " FROM ");
		deparseExpr(state, lsecond(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		if (list_length(func_call->args) == 3)
		{
			deparseAppendStringInfoString(state, " FOR ");
			deparseExpr(state, lthird(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		}
		deparseAppendStringInfoChar(state, ')');
		return;
	} else if (func_call->funcformat == COERCE_SQL_SYNTAX &&
		list_length(func_call->funcname) == 2 &&
		strcmp(strVal(linitial(func_call->funcname)), "pg_catalog") == 0 &&
		strcmp(strVal(lsecond(func_call->funcname)), "position") == 0 &&
		list_length(func_call->args) == 2)
	{
		/*
		 * "POSITION" is a keyword on its own merit, and only accepts the
		 * keyword parameter style when its called as a keyword, not as a regular function (i.e. pg_catalog.position)
		 * Note that the first and second arguments are switched in this format
		 */
		deparseAppendStringInfoString(state, "POSITION(");
		deparseBExpr(state, lsecond(func_call->args));
		deparseAppendStringInfoString(state, " IN ");
		deparseBExpr(state, linitial(func_call->args));
		deparseAppendStringInfoChar(state, ')');
		return;
	} else if (func_call->funcformat == COERCE_SQL_SYNTAX &&
		list_length(func_call->funcname) == 2 &&
		strcmp(strVal(linitial(func_call->funcname)), "pg_catalog") == 0 &&
		strcmp(strVal(lsecond(func_call->funcname)), "overlay") == 0 &&
		list_length(func_call->args) == 3)
	{
		/*
		 * "OVERLAY" is a keyword on its own merit, and only accepts the
		 * keyword parameter style when its called as a keyword, not as a regular function (i.e. pg_catalog.overlay)
		 */
		deparseAppendStringInfoString(state, "overlay(");
		deparseExpr(state, linitial(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, " placing ");
		deparseExpr(state, lsecond(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, " from ");
		deparseExpr(state, lthird(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoChar(state, ')');
		return;
	} else if (func_call->funcformat == COERCE_SQL_SYNTAX &&
		list_length(func_call->funcname) == 2 &&
		strcmp(strVal(linitial(func_call->funcname)), "pg_catalog") == 0 &&
		strcmp(strVal(lsecond(func_call->funcname)), "pg_collation_for") == 0 &&
		list_length(func_call->args) == 1)
	{
		/*
		 * "collation for" is a keyword on its own merit, and only accepts the
		 * keyword parameter style when its called as a keyword, not as a regular function (i.e. pg_catalog.overlay)
		 */
		deparseAppendStringInfoString(state, "collation for (");
		deparseExpr(state, linitial(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoChar(state, ')');
		return;
	} else if (func_call->funcformat == COERCE_SQL_SYNTAX &&
		list_length(func_call->funcname) == 2 &&
		strcmp(strVal(linitial(func_call->funcname)), "pg_catalog") == 0 &&
		strcmp(strVal(lsecond(func_call->funcname)), "extract") == 0 &&
		list_length(func_call->args) == 2)
	{
		/*
		 * "EXTRACT" is a keyword on its own merit, and only accepts the
		 * keyword parameter style when its called as a keyword, not as a regular function (i.e. pg_catalog.extract)
		 */
		deparseAppendStringInfoString(state, "extract (");
		deparseExpr(state, linitial(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, " FROM ");
		deparseExpr(state, lsecond(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoChar(state, ')');
		return;
	} else if (func_call->funcformat == COERCE_SQL_SYNTAX &&
		list_length(func_call->funcname) == 2 &&
		strcmp(strVal(linitial(func_call->funcname)), "pg_catalog") == 0 &&
		strcmp(strVal(lsecond(func_call->funcname)), "overlaps") == 0 &&
		list_length(func_call->args) == 4)
	{
		/*
		 * "OVERLAPS" is a keyword on its own merit, and only accepts the
		 * keyword parameter style when its called as a keyword, not as a regular function (i.e. pg_catalog.overlaps)
		 * format: (start_1, end_1) OVERLAPS (start_2, end_2)
		 */
		deparseAppendStringInfoChar(state, '(');
		deparseExpr(state, linitial(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, ", ");
		deparseExpr(state, lsecond(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, ") OVERLAPS (");
		deparseExpr(state, lthird(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, ", ");
		deparseExpr(state, lfourth(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoChar(state, ')');
		return;
	} else if (func_call->funcformat == COERCE_SQL_SYNTAX &&
		list_length(func_call->funcname) == 2 &&
		strcmp(strVal(linitial(func_call->funcname)), "pg_catalog") == 0 &&
		(
			strcmp(strVal(lsecond(func_call->funcname)), "ltrim") == 0 ||
			strcmp(strVal(lsecond(func_call->funcname)), "btrim") == 0 ||
			strcmp(strVal(lsecond(func_call->funcname)), "rtrim") == 0
		))
	{
		/*
		 * "TRIM " is a keyword on its own merit, and only accepts the
		 * keyword parameter style when its called as a keyword, not as a regular function (i.e. pg_catalog.ltrim)
		 * Note that the first and second arguments are switched in this format
		 */
		Assert(list_length(func_call->args) == 1 || list_length(func_call->args) == 2);
		deparseAppendStringInfoString(state, "TRIM (");
		if (strcmp(strVal(lsecond(func_call->funcname)), "ltrim") == 0)
			deparseAppendStringInfoString(state, "LEADING ");
		else if (strcmp(strVal(lsecond(func_call->funcname)), "btrim") == 0)
			deparseAppendStringInfoString(state, "BOTH ");
		else if (strcmp(strVal(lsecond(func_call->funcname)), "rtrim") == 0)
			deparseAppendStringInfoString(state, "TRAILING ");

		if (list_length(func_call->args) == 2)
			deparseExpr(state, lsecond(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, " FROM ");
		deparseExpr(state, linitial(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoChar(state, ')');
		return;
	} else if (func_call->funcformat == COERCE_SQL_SYNTAX &&
		list_length(func_call->funcname) == 2 &&
		strcmp(strVal(linitial(func_call->funcname)), "pg_catalog") == 0 &&
		strcmp(strVal(lsecond(func_call->funcname)), "timezone") == 0 &&
		list_length(func_call->args) > 0 &&
		list_length(func_call->args) <= 2)
	{
		/*
		 * "AT TIME ZONE" is a keyword on its own merit, and only accepts the
		 * keyword parameter style when its called as a keyword, not as a regular function (i.e. pg_catalog.timezone)
		 * Note that the arguments are swapped in this case
		 */
		Node* e;
		bool isLocal = list_length(func_call->args) == 1;

		if (isLocal)
			e = linitial(func_call->args);
		else
			e = lsecond(func_call->args);

		// Note that callers outside of an "a_expr" context add parens around this (see needsParensAsBExpr)
		deparseExprOperand(state, e, DEPARSE_PREC_AT, DEPARSE_OPERAND_LEFT, DEPARSE_NODE_CONTEXT_A_EXPR);

		if (isLocal)
			deparseAppendStringInfoString(state, " AT LOCAL");
		else {
			deparseAppendStringInfoString(state, " AT TIME ZONE ");
			deparseExprOperand(state, linitial(func_call->args), DEPARSE_PREC_AT, DEPARSE_OPERAND_RIGHT, DEPARSE_NODE_CONTEXT_A_EXPR);
		}

		return;
	} else if (func_call->funcformat == COERCE_SQL_SYNTAX &&
		list_length(func_call->funcname) == 2 &&
		strcmp(strVal(linitial(func_call->funcname)), "pg_catalog") == 0 &&
		strcmp(strVal(lsecond(func_call->funcname)), "normalize") == 0)
	{
		/*
		 * "NORMALIZE" is a keyword on its own merit, and only accepts the
		 * keyword parameter style when its called as a keyword, not as a regular function (i.e. pg_catalog.normalize)
		 */
		Assert(list_length(func_call->args) == 1 || list_length(func_call->args) == 2);
		deparseAppendStringInfoString(state, "normalize (");

		deparseExpr(state, linitial(func_call->args), DEPARSE_NODE_CONTEXT_A_EXPR);
		if (list_length(func_call->args) == 2)
		{
			deparseAppendStringInfoString(state, ", ");
			Assert(IsA(lsecond(func_call->args), A_Const));
			A_Const *aconst = lsecond(func_call->args);
			deparseValue(state, &aconst->val, DEPARSE_NODE_CONTEXT_NONE);
		}
		deparseAppendStringInfoChar(state, ')');
		return;
	} else if (func_call->funcformat == COERCE_SQL_SYNTAX &&
		list_length(func_call->funcname) == 2 &&
		strcmp(strVal(linitial(func_call->funcname)), "pg_catalog") == 0 &&
		strcmp(strVal(lsecond(func_call->funcname)), "is_normalized") == 0)
	{
		/*
		 * "IS NORMALIZED" is a keyword on its own merit, and only accepts the
		 * keyword parameter style when its called as a keyword, not as a regular function (i.e. pg_catalog.is_normalized)
		 */
		Assert(list_length(func_call->args) == 1 || list_length(func_call->args) == 2);

		deparseExprOperand(state, linitial(func_call->args), DEPARSE_PREC_IS, DEPARSE_OPERAND_LEFT, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, " IS ");
		if (list_length(func_call->args) == 2)
		{
			Assert(IsA(lsecond(func_call->args), A_Const));
			A_Const *aconst = lsecond(func_call->args);
			deparseValue(state, &aconst->val, DEPARSE_NODE_CONTEXT_NONE);
		}
		deparseAppendStringInfoString(state, " NORMALIZED ");
		return;
	} else if (func_call->funcformat == COERCE_SQL_SYNTAX &&
		list_length(func_call->funcname) == 2 &&
		strcmp(strVal(linitial(func_call->funcname)), "pg_catalog") == 0 &&
		strcmp(strVal(lsecond(func_call->funcname)), "xmlexists") == 0 &&
		list_length(func_call->args) == 2)
	{
		deparseAppendStringInfoString(state, "xmlexists (");
		deparseCExpr(state, linitial(func_call->args));
		deparseAppendStringInfoString(state, " PASSING ");
		deparseCExpr(state, lsecond(func_call->args));
		deparseAppendStringInfoChar(state, ')');
		return;
	} else if (func_call->funcformat == COERCE_SQL_SYNTAX &&
		list_length(func_call->funcname) == 2 &&
		strcmp(strVal(linitial(func_call->funcname)), "pg_catalog") == 0 &&
		strcmp(strVal(lsecond(func_call->funcname)), "system_user") == 0)
	{
		deparseAppendStringInfoString(state, "SYSTEM_USER");
		return;
	}
		
	deparseFuncName(state, func_call->funcname);
	deparseAppendStringInfoChar(state, '(');

	if (func_call->agg_distinct)
		deparseAppendStringInfoString(state, "DISTINCT ");

	if (func_call->agg_star)
	{
		deparseAppendStringInfoChar(state, '*');
	}
	else if (list_length(func_call->args) > 0)
	{
		foreach(lc, func_call->args)
		{
			if (func_call->func_variadic && !lnext(func_call->args, lc))
				deparseAppendStringInfoString(state, "VARIADIC ");
			deparseFuncArgExpr(state, lfirst(lc));
			if (lnext(func_call->args, lc))
				deparseAppendStringInfoString(state, ", ");
		}
	}
	deparseAppendStringInfoChar(state, ' ');

	if (func_call->agg_order != NULL && !func_call->agg_within_group)
	{
		deparseOptSortClause(state, func_call->agg_order, DEPARSE_NODE_CONTEXT_NONE);
	}

	removeTrailingSpace(state);
	deparseAppendStringInfoString(state, ") ");

	if (func_call->agg_order != NULL && func_call->agg_within_group)
	{
		deparseAppendStringInfoString(state, "WITHIN GROUP (");
		deparseOptSortClause(state, func_call->agg_order, DEPARSE_NODE_CONTEXT_NONE);
		removeTrailingSpace(state);
		deparseAppendStringInfoString(state, ") ");
	}

	if (func_call->agg_filter)
	{
		deparseAppendStringInfoString(state, "FILTER (WHERE ");
		deparseExpr(state, func_call->agg_filter, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, ") ");
	}

	if (func_call->over)
	{
		deparseAppendStringInfoString(state, "OVER ");
		if (func_call->over->name)
			deparseAppendStringInfoString(state, func_call->over->name);
		else
			deparseWindowDef(state, func_call->over);
	}

	removeTrailingSpace(state);
}

static void deparseWindowDef(DeparseState *state, WindowDef* window_def)
{
	ListCell *lc;
	DeparseStateNestingLevel *parent_level = NULL;

	// The parent node is responsible for outputting window_def->name

	// Special case completely empty window clauses and return early
	if (!window_def->refname && list_length(window_def->partitionClause) == 0 &&
		list_length(window_def->orderClause) == 0 &&
		!(window_def->frameOptions & FRAMEOPTION_NONDEFAULT))
	{
		deparseAppendStringInfoString(state, "()");
		return;
	}

	deparseAppendStringInfoChar(state, '(');
	parent_level = deparseStateIncreaseNestingLevel(state);

	if (window_def->refname != NULL)
	{
		deparseAppendStringInfoString(state, quote_identifier(window_def->refname));
		deparseAppendStringInfoChar(state, ' ');
	}

	if (list_length(window_def->partitionClause) > 0)
	{
		deparseAppendStringInfoString(state, "PARTITION BY ");
		deparseExprList(state, window_def->partitionClause);
		deparseAppendStringInfoChar(state, ' ');
	}

	removeTrailingSpace(state);
	deparseAppendPart(state, true);
	deparseOptSortClause(state, window_def->orderClause, DEPARSE_NODE_CONTEXT_NONE);

	if (window_def->frameOptions & FRAMEOPTION_NONDEFAULT)
	{
		deparseAppendPartGroup(state, NULL, DEPARSE_PART_NO_INDENT);
		if (window_def->frameOptions & FRAMEOPTION_RANGE)
			deparseAppendStringInfoString(state, "RANGE ");
		else if (window_def->frameOptions & FRAMEOPTION_ROWS)
			deparseAppendStringInfoString(state, "ROWS ");
		else if (window_def->frameOptions & FRAMEOPTION_GROUPS)
			deparseAppendStringInfoString(state, "GROUPS ");
	
		if (window_def->frameOptions & FRAMEOPTION_BETWEEN)
			deparseAppendStringInfoString(state, "BETWEEN ");

		// frame_start
		if (window_def->frameOptions & FRAMEOPTION_START_UNBOUNDED_PRECEDING)
		{
			deparseAppendStringInfoString(state, "UNBOUNDED PRECEDING ");
		}
		else if (window_def->frameOptions & FRAMEOPTION_START_UNBOUNDED_FOLLOWING)
		{
			Assert(false); // disallowed
		}
		else if (window_def->frameOptions & FRAMEOPTION_START_CURRENT_ROW)
		{
			deparseAppendStringInfoString(state, "CURRENT ROW ");
		}
		else if (window_def->frameOptions & FRAMEOPTION_START_OFFSET_PRECEDING)
		{
			Assert(window_def->startOffset != NULL);
			deparseExpr(state, window_def->startOffset, DEPARSE_NODE_CONTEXT_A_EXPR);
			deparseAppendStringInfoString(state, " PRECEDING ");
		}
		else if (window_def->frameOptions & FRAMEOPTION_START_OFFSET_FOLLOWING)
		{
			Assert(window_def->startOffset != NULL);
			deparseExpr(state, window_def->startOffset, DEPARSE_NODE_CONTEXT_A_EXPR);
			deparseAppendStringInfoString(state, " FOLLOWING ");
		}

		if (window_def->frameOptions & FRAMEOPTION_BETWEEN)
		{
			deparseAppendStringInfoString(state, "AND ");

			// frame_end
			if (window_def->frameOptions & FRAMEOPTION_END_UNBOUNDED_PRECEDING)
			{
				Assert(false); // disallowed
			}
			else if (window_def->frameOptions & FRAMEOPTION_END_UNBOUNDED_FOLLOWING)
			{
				deparseAppendStringInfoString(state, "UNBOUNDED FOLLOWING ");
			}
			else if (window_def->frameOptions & FRAMEOPTION_END_CURRENT_ROW)
			{
				deparseAppendStringInfoString(state, "CURRENT ROW ");
			}
			else if (window_def->frameOptions & FRAMEOPTION_END_OFFSET_PRECEDING)
			{
				Assert(window_def->endOffset != NULL);
				deparseExpr(state, window_def->endOffset, DEPARSE_NODE_CONTEXT_A_EXPR);
				deparseAppendStringInfoString(state, " PRECEDING ");
			}
			else if (window_def->frameOptions & FRAMEOPTION_END_OFFSET_FOLLOWING)
			{
				Assert(window_def->endOffset != NULL);
				deparseExpr(state, window_def->endOffset, DEPARSE_NODE_CONTEXT_A_EXPR);
				deparseAppendStringInfoString(state, " FOLLOWING ");
			}
		}

		if (window_def->frameOptions & FRAMEOPTION_EXCLUDE_CURRENT_ROW)
			deparseAppendStringInfoString(state, "EXCLUDE CURRENT ROW ");
		else if (window_def->frameOptions & FRAMEOPTION_EXCLUDE_GROUP)
			deparseAppendStringInfoString(state, "EXCLUDE GROUP ");
		else if (window_def->frameOptions & FRAMEOPTION_EXCLUDE_TIES)
			deparseAppendStringInfoString(state, "EXCLUDE TIES ");
	}

	removeTrailingSpace(state);
	deparseStateDecreaseNestingLevel(state, parent_level);
	deparseAppendStringInfoChar(state, ')');
}

static void deparseColumnRef(DeparseState *state, ColumnRef* column_ref)
{
	Assert(list_length(column_ref->fields) >= 1);

	deparseAppendCommentsIfNeeded(state, column_ref->location);

	if (IsA(linitial(column_ref->fields), A_Star))
		deparseAStar(state, castNode(A_Star, linitial(column_ref->fields)));
	else if (IsA(linitial(column_ref->fields), String))
		deparseColLabel(state, strVal(linitial(column_ref->fields)));

	deparseOptIndirection(state, column_ref->fields, 1);
}

static void deparseSubLink(DeparseState *state, SubLink* sub_link)
{
	switch (sub_link->subLinkType) {
		case EXISTS_SUBLINK:
			deparseAppendStringInfoString(state, "EXISTS (");
			deparseSelectStmt(state, castNode(SelectStmt, sub_link->subselect), DEPARSE_NODE_CONTEXT_NONE);
			deparseAppendStringInfoChar(state, ')');
			return;
		case ALL_SUBLINK:
			// The left operand gets compared against the operator token, not the precedence of the whole rule
			deparseExprOperand(state, sub_link->testexpr, getSubqueryOperatorPrecedence(sub_link->operName), DEPARSE_OPERAND_LEFT, DEPARSE_NODE_CONTEXT_A_EXPR);
			deparseAppendStringInfoChar(state, ' ');
			deparseSubqueryOp(state, sub_link->operName);
			deparseAppendStringInfoString(state, " ALL (");
			deparseSelectStmt(state, castNode(SelectStmt, sub_link->subselect), DEPARSE_NODE_CONTEXT_NONE);
			deparseAppendStringInfoChar(state, ')');
			return;
		case ANY_SUBLINK:
			if (list_length(sub_link->operName) > 0)
				deparseExprOperand(state, sub_link->testexpr, getSubqueryOperatorPrecedence(sub_link->operName), DEPARSE_OPERAND_LEFT, DEPARSE_NODE_CONTEXT_A_EXPR);
			else
				deparseExprOperand(state, sub_link->testexpr, DEPARSE_PREC_BETWEEN /* IN */, DEPARSE_OPERAND_LEFT, DEPARSE_NODE_CONTEXT_A_EXPR);
			if (list_length(sub_link->operName) > 0)
			{
				deparseAppendStringInfoChar(state, ' ');
				deparseSubqueryOp(state, sub_link->operName);
				deparseAppendStringInfoString(state, " ANY ");
			}
			else
			{
				deparseAppendStringInfoString(state, " IN ");
			}
			deparseAppendStringInfoChar(state, '(');
			deparseSelectStmt(state, castNode(SelectStmt, sub_link->subselect), DEPARSE_NODE_CONTEXT_NONE);
			deparseAppendStringInfoChar(state, ')');
			return;
		case ROWCOMPARE_SUBLINK:
			// Not present in raw parse trees
			Assert(false);
			return;
		case EXPR_SUBLINK:
			deparseAppendStringInfoString(state, "(");
			deparseSelectStmt(state, castNode(SelectStmt, sub_link->subselect), DEPARSE_NODE_CONTEXT_NONE);
			deparseAppendStringInfoChar(state, ')');
			return;
		case MULTIEXPR_SUBLINK:
			// Not present in raw parse trees
			Assert(false);
			return;
		case ARRAY_SUBLINK:
			deparseAppendStringInfoString(state, "ARRAY(");
			deparseSelectStmt(state, castNode(SelectStmt, sub_link->subselect), DEPARSE_NODE_CONTEXT_NONE);
			deparseAppendStringInfoChar(state, ')');
			return;
		case CTE_SUBLINK: /* for SubPlans only */
			// Not present in raw parse trees
			Assert(false);
			return;
	}
}

// Checks whether a node needs parens when used in a "b_expr" context (because the node is handled by the "a_expr" rule in gram.y)
static bool
needsParensAsBExpr(Node *node)
{
	if (node == NULL)
		return false;

	// These have their own "b_expr" rules (their operands get deparsed in a "b_expr" context as well)
	if (IsA(node, A_Expr))
	{
		switch (castNode(A_Expr, node)->kind)
		{
			case AEXPR_OP:
			case AEXPR_DISTINCT:
			case AEXPR_NOT_DISTINCT:
			case AEXPR_NULLIF: /* function call syntax */
				return false;
			default:
				return true;
		}
	}
	if (IsA(node, XmlExpr) && castNode(XmlExpr, node)->op == IS_DOCUMENT)
		return false;

	// "row OVERLAPS row" can't be ambiguous (so it has no precedence), but is only handled by "a_expr"
	if (IsA(node, FuncCall) && isSqlSyntaxFuncCall(castNode(FuncCall, node), "overlaps") &&
		list_length(castNode(FuncCall, node)->args) == 4)
		return true;

	// Everything else that isn't a "c_expr" (e.g. NOT, IS NULL, COLLATE, AT TIME ZONE) is only handled by "a_expr"
	return getExprPrecedence(node) != DEPARSE_PREC_ATOM;
}

// Checks whether a node needs parens in a "c_expr" context, i.e. anything that isn't deparsed as a "c_expr" (e.g. "a + b", "NOT a", "x = ANY (...)", "x AT TIME ZONE y" or "x IS DOCUMENT")
static bool
needsParensAsCExpr(Node *node)
{
	return getExprPrecedence(node) != DEPARSE_PREC_ATOM || needsParensAsBExpr(node);
}

// This handles "A_Expr" parse tree objects, which are a subset of the rules in "a_expr" (handled by deparseExpr)
static void deparseAExpr(DeparseState *state, A_Expr* a_expr, DeparseNodeContext context)
{
	if (a_expr->rexpr == NULL)
		elog(ERROR, "deparse: unexpected NULL rexpr in A_Expr");

	char *name;
	DeparseExprPrec prec = getExprPrecedence((Node *) a_expr);

	switch (a_expr->kind) {
		case AEXPR_OP: /* normal operator */
			{
				if (a_expr->lexpr != NULL)
				{
					deparseExprOperand(state, a_expr->lexpr, prec, DEPARSE_OPERAND_LEFT, context);
					deparseAppendStringInfoChar(state, ' ');
				}
				deparseQualOp(state, a_expr->name);
				if (a_expr->rexpr != NULL)
				{
					deparseAppendStringInfoChar(state, ' ');
					deparseExprOperand(state, a_expr->rexpr, prec, DEPARSE_OPERAND_RIGHT, context);
				}
			}
			return;
		case AEXPR_OP_ANY: /* scalar op ANY (array) */
			// The left operand gets compared against the operator token, not the precedence of the whole rule
			deparseExprOperand(state, a_expr->lexpr, getSubqueryOperatorPrecedence(a_expr->name), DEPARSE_OPERAND_LEFT, context);
			deparseAppendStringInfoChar(state, ' ');
			deparseSubqueryOp(state, a_expr->name);
			deparseAppendStringInfoString(state, " ANY(");
			deparseExpr(state, a_expr->rexpr, DEPARSE_NODE_CONTEXT_A_EXPR);
			deparseAppendStringInfoChar(state, ')');
			return;
		case AEXPR_OP_ALL: /* scalar op ALL (array) */
			deparseExprOperand(state, a_expr->lexpr, getSubqueryOperatorPrecedence(a_expr->name), DEPARSE_OPERAND_LEFT, context);
			deparseAppendStringInfoChar(state, ' ');
			deparseSubqueryOp(state, a_expr->name);
			deparseAppendStringInfoString(state, " ALL(");
			deparseExpr(state, a_expr->rexpr, DEPARSE_NODE_CONTEXT_A_EXPR);
			deparseAppendStringInfoChar(state, ')');
			return;
		case AEXPR_DISTINCT: /* IS DISTINCT FROM - name must be "=" */
			Assert(list_length(a_expr->name) == 1);
			Assert(IsA(linitial(a_expr->name), String));
			Assert(strcmp(strVal(linitial(a_expr->name)), "=") == 0);

			deparseExprOperand(state, a_expr->lexpr, prec, DEPARSE_OPERAND_LEFT, context);
			deparseAppendStringInfoString(state, " IS DISTINCT FROM ");
			deparseExprOperand(state, a_expr->rexpr, prec, DEPARSE_OPERAND_RIGHT, context);
			return;
		case AEXPR_NOT_DISTINCT: /* IS NOT DISTINCT FROM - name must be "=" */
			Assert(list_length(a_expr->name) == 1);
			Assert(IsA(linitial(a_expr->name), String));
			Assert(strcmp(strVal(linitial(a_expr->name)), "=") == 0);

			deparseExprOperand(state, a_expr->lexpr, prec, DEPARSE_OPERAND_LEFT, context);
			deparseAppendStringInfoString(state, " IS NOT DISTINCT FROM ");
			deparseExprOperand(state, a_expr->rexpr, prec, DEPARSE_OPERAND_RIGHT, context);
			return;
		case AEXPR_NULLIF: /* NULLIF - name must be "=" */
			Assert(list_length(a_expr->name) == 1);
			Assert(IsA(linitial(a_expr->name), String));
			Assert(strcmp(strVal(linitial(a_expr->name)), "=") == 0);

			deparseAppendStringInfoString(state, "NULLIF(");
			deparseExpr(state, a_expr->lexpr, context);
			deparseAppendStringInfoString(state, ", ");
			deparseExpr(state, a_expr->rexpr, context);
			deparseAppendStringInfoChar(state, ')');
			return;
		case AEXPR_IN: /* [NOT] IN - name must be "=" or "<>" */
			Assert(list_length(a_expr->name) == 1);
			Assert(IsA(linitial(a_expr->name), String));
			Assert(IsA(a_expr->rexpr, List));
			deparseExprOperand(state, a_expr->lexpr, prec, DEPARSE_OPERAND_LEFT, context);
			deparseAppendStringInfoChar(state, ' ');
			name = ((union ValUnion *) linitial(a_expr->name))->sval.sval;
			if (strcmp(name, "=") == 0) {
				deparseAppendStringInfoString(state, "IN ");
			} else if (strcmp(name, "<>") == 0) {
				deparseAppendStringInfoString(state, "NOT IN ");
			} else {
				Assert(false);
			}
			deparseAppendStringInfoChar(state, '(');
			if (IsA(a_expr->rexpr, SubLink))
				deparseSubLink(state, castNode(SubLink, a_expr->rexpr));
			else
				deparseExprList(state, castNode(List, a_expr->rexpr));
			deparseAppendStringInfoChar(state, ')');
			return;
		case AEXPR_LIKE: /* [NOT] LIKE - name must be "~~" or "!~~" */
			Assert(list_length(a_expr->name) == 1);
			Assert(IsA(linitial(a_expr->name), String));
			deparseExprOperand(state, a_expr->lexpr, prec, DEPARSE_OPERAND_LEFT, context);
			deparseAppendStringInfoChar(state, ' ');

			name = ((union ValUnion *) linitial(a_expr->name))->sval.sval;
			if (strcmp(name, "~~") == 0) {
				deparseAppendStringInfoString(state, "LIKE ");
			} else if (strcmp(name, "!~~") == 0) {
				deparseAppendStringInfoString(state, "NOT LIKE ");
			} else {
				Assert(false);
			}

			deparseExprOperand(state, a_expr->rexpr, prec, DEPARSE_OPERAND_RIGHT, context);
			return;
		case AEXPR_ILIKE: /* [NOT] ILIKE - name must be "~~*" or "!~~*" */
			Assert(list_length(a_expr->name) == 1);
			Assert(IsA(linitial(a_expr->name), String));
			deparseExprOperand(state, a_expr->lexpr, prec, DEPARSE_OPERAND_LEFT, context);
			deparseAppendStringInfoChar(state, ' ');

			name = ((union ValUnion *) linitial(a_expr->name))->sval.sval;
			if (strcmp(name, "~~*") == 0) {
				deparseAppendStringInfoString(state, "ILIKE ");
			} else if (strcmp(name, "!~~*") == 0) {
				deparseAppendStringInfoString(state, "NOT ILIKE ");
			} else {
				Assert(false);
			}

			deparseExprOperand(state, a_expr->rexpr, prec, DEPARSE_OPERAND_RIGHT, context);
			return;
		case AEXPR_SIMILAR: /* [NOT] SIMILAR - name must be "~" or "!~" */
			Assert(list_length(a_expr->name) == 1);
			Assert(IsA(linitial(a_expr->name), String));
			deparseExprOperand(state, a_expr->lexpr, prec, DEPARSE_OPERAND_LEFT, context);
			deparseAppendStringInfoChar(state, ' ');

			name = ((union ValUnion *) linitial(a_expr->name))->sval.sval;
			if (strcmp(name, "~") == 0) {
				deparseAppendStringInfoString(state, "SIMILAR TO ");
			} else if (strcmp(name, "!~") == 0) {
				deparseAppendStringInfoString(state, "NOT SIMILAR TO ");
			} else {
				Assert(false);
			}

			FuncCall *n = castNode(FuncCall, a_expr->rexpr);
			Assert(list_length(n->funcname) == 2);
			Assert(strcmp(strVal(linitial(n->funcname)), "pg_catalog") == 0);
			Assert(strcmp(strVal(lsecond(n->funcname)), "similar_to_escape") == 0);
			Assert(list_length(n->args) == 1 || list_length(n->args) == 2);

			deparseExprOperand(state, linitial(n->args), prec, DEPARSE_OPERAND_RIGHT, context);
			if (list_length(n->args) == 2)
			{
				deparseAppendStringInfoString(state, " ESCAPE ");
				deparseExprOperand(state, lsecond(n->args), prec, DEPARSE_OPERAND_RIGHT, context);
			}

			return;
		case AEXPR_BETWEEN: /* name must be "BETWEEN" */
		case AEXPR_NOT_BETWEEN: /* name must be "NOT BETWEEN" */
		case AEXPR_BETWEEN_SYM: /* name must be "BETWEEN SYMMETRIC" */
		case AEXPR_NOT_BETWEEN_SYM: /* name must be "NOT BETWEEN SYMMETRIC" */
			Assert(list_length(a_expr->name) == 1);
			Assert(IsA(linitial(a_expr->name), String));
			Assert(IsA(a_expr->rexpr, List));

			Assert(list_length(castNode(List, a_expr->rexpr)) == 2);

			deparseExprOperand(state, a_expr->lexpr, prec, DEPARSE_OPERAND_LEFT, context);
			deparseAppendStringInfoChar(state, ' ');
			deparseAppendStringInfoString(state, strVal(linitial(a_expr->name)));
			deparseAppendStringInfoChar(state, ' ');

			// The lower bound is a "b_expr" in gram.y, the upper bound an "a_expr"
			deparseExprOperand(state, linitial(castNode(List, a_expr->rexpr)), prec, DEPARSE_OPERAND_RIGHT, DEPARSE_NODE_CONTEXT_B_EXPR);
			deparseAppendStringInfoString(state, " AND ");
			deparseExprOperand(state, lsecond(castNode(List, a_expr->rexpr)), prec, DEPARSE_OPERAND_RIGHT, context);
			return;
	}
}

static void deparseBoolExpr(DeparseState *state, BoolExpr *bool_expr)
{
	const ListCell *lc = NULL;
	switch (bool_expr->boolop)
	{
		case AND_EXPR:
			foreach(lc, bool_expr->args)
			{
				// Put parantheses around AND + OR nodes that are inside
				bool need_parens = IsA(lfirst(lc), BoolExpr) && (castNode(BoolExpr, lfirst(lc))->boolop == AND_EXPR || castNode(BoolExpr, lfirst(lc))->boolop == OR_EXPR);

				if (need_parens)
					deparseAppendStringInfoChar(state, '(');

				deparseExpr(state, lfirst(lc), DEPARSE_NODE_CONTEXT_A_EXPR);

				if (need_parens)
					deparseAppendStringInfoChar(state, ')');

				if (lnext(bool_expr->args, lc))
				{
					deparseAppendPart(state, true);
					deparseAppendStringInfoString(state, "AND ");
				}
			}
			return;
		case OR_EXPR:
			foreach(lc, bool_expr->args)
			{
				// Put parantheses around AND + OR nodes that are inside
				bool need_parens = IsA(lfirst(lc), BoolExpr) && (castNode(BoolExpr, lfirst(lc))->boolop == AND_EXPR || castNode(BoolExpr, lfirst(lc))->boolop == OR_EXPR);

				if (need_parens)
					deparseAppendStringInfoChar(state, '(');

				deparseExpr(state, lfirst(lc), DEPARSE_NODE_CONTEXT_A_EXPR);

				if (need_parens)
					deparseAppendStringInfoChar(state, ')');

				if (lnext(bool_expr->args, lc))
					deparseAppendStringInfoString(state, " OR ");
			}
			return;
		case NOT_EXPR:
			Assert(list_length(bool_expr->args) == 1);
			bool need_parens = IsA(linitial(bool_expr->args), BoolExpr) && (castNode(BoolExpr, linitial(bool_expr->args))->boolop == AND_EXPR || castNode(BoolExpr, linitial(bool_expr->args))->boolop == OR_EXPR);
			deparseAppendStringInfoString(state, "NOT ");
			if (need_parens)
				deparseAppendStringInfoChar(state, '(');
			deparseExpr(state, linitial(bool_expr->args), DEPARSE_NODE_CONTEXT_A_EXPR);
			if (need_parens)
				deparseAppendStringInfoChar(state, ')');
			return;
	}
}

static void deparseAStar(DeparseState *state, A_Star *a_star)
{
	deparseAppendStringInfoChar(state, '*');
}

static void deparseCollateClause(DeparseState *state, CollateClause* collate_clause)
{
	ListCell *lc;
	if (collate_clause->arg != NULL)
	{
		deparseExprOperand(state, collate_clause->arg, DEPARSE_PREC_COLLATE, DEPARSE_OPERAND_LEFT, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoChar(state, ' ');
	}
	deparseAppendStringInfoString(state, "COLLATE ");
	deparseAnyName(state, collate_clause->collname);
}

// "sortby" in gram.y
static void deparseSortBy(DeparseState *state, SortBy* sort_by)
{
	deparseExpr(state, sort_by->node, DEPARSE_NODE_CONTEXT_A_EXPR);
	deparseAppendStringInfoChar(state, ' ');

	switch (sort_by->sortby_dir)
	{
		case SORTBY_DEFAULT:
			break;
		case SORTBY_ASC:
			deparseAppendStringInfoString(state, "ASC ");
			break;
		case SORTBY_DESC:
			deparseAppendStringInfoString(state, "DESC ");
			break;
		case SORTBY_USING:
			deparseAppendStringInfoString(state, "USING ");
			deparseQualOp(state, sort_by->useOp);
			break;
	}

	switch (sort_by->sortby_nulls)
	{
		case SORTBY_NULLS_DEFAULT:
			break;
		case SORTBY_NULLS_FIRST:
			deparseAppendStringInfoString(state, "NULLS FIRST ");
			break;
		case SORTBY_NULLS_LAST:
			deparseAppendStringInfoString(state, "NULLS LAST ");
			break;
	}

	removeTrailingSpace(state);
}

static void deparseParamRef(DeparseState *state, ParamRef* param_ref)
{
	if (param_ref->number == 0) {
		deparseAppendStringInfoChar(state, '?');
	} else {
		deparseAppendStringInfo(state, "$%d", param_ref->number);
	}
}

static void deparseSQLValueFunction(DeparseState *state, SQLValueFunction* sql_value_function)
{
	switch (sql_value_function->op)
	{
		case SVFOP_CURRENT_DATE:
			deparseAppendStringInfoString(state, "current_date");
			break;
		case SVFOP_CURRENT_TIME:
			deparseAppendStringInfoString(state, "current_time");
			break;
		case SVFOP_CURRENT_TIME_N:
			deparseAppendStringInfoString(state, "current_time"); // with precision
			break;
		case SVFOP_CURRENT_TIMESTAMP:
			deparseAppendStringInfoString(state, "current_timestamp");
			break;
		case SVFOP_CURRENT_TIMESTAMP_N:
			deparseAppendStringInfoString(state, "current_timestamp"); // with precision
			break;
		case SVFOP_LOCALTIME:
			deparseAppendStringInfoString(state, "localtime");
			break;
		case SVFOP_LOCALTIME_N:
			deparseAppendStringInfoString(state, "localtime"); // with precision
			break;
		case SVFOP_LOCALTIMESTAMP:
			deparseAppendStringInfoString(state, "localtimestamp");
			break;
		case SVFOP_LOCALTIMESTAMP_N:
			deparseAppendStringInfoString(state, "localtimestamp"); // with precision
			break;
		case SVFOP_CURRENT_ROLE:
			deparseAppendStringInfoString(state, "current_role");
			break;
		case SVFOP_CURRENT_USER:
			deparseAppendStringInfoString(state, "current_user");
			break;
		case SVFOP_USER:
			deparseAppendStringInfoString(state, "user");
			break;
		case SVFOP_SESSION_USER:
			deparseAppendStringInfoString(state, "session_user");
			break;
		case SVFOP_CURRENT_CATALOG:
			deparseAppendStringInfoString(state, "current_catalog");
			break;
		case SVFOP_CURRENT_SCHEMA:
			deparseAppendStringInfoString(state, "current_schema");
			break;
	}

	if (sql_value_function->typmod != -1)
	{
		deparseAppendStringInfo(state, "(%d)", sql_value_function->typmod);
	}
}

static void deparseWithClause(DeparseState *state, WithClause *with_clause)
{
	ListCell *lc;

	deparseAppendPartGroup(state, "WITH", DEPARSE_PART_NO_INDENT);
	if (with_clause->recursive)
		deparseAppendStringInfoString(state, "RECURSIVE ");

	foreach(lc, with_clause->ctes) {
		deparseCommonTableExpr(state, castNode(CommonTableExpr, lfirst(lc)));
		if (lnext(with_clause->ctes, lc))
			deparseAppendStringInfoString(state, ", ");
	}

	removeTrailingSpace(state);
}

// "joined_table" in gram.y
static void deparseJoinExpr(DeparseState *state, JoinExpr *join_expr)
{
	if (join_expr->larg == NULL || join_expr->rarg == NULL)
		elog(ERROR, "deparse: unexpected NULL larg/rarg in JoinExpr");

	ListCell *lc;
	bool need_alias_parens = join_expr->alias != NULL;
	bool need_rarg_parens = IsA(join_expr->rarg, JoinExpr) && castNode(JoinExpr, join_expr->rarg)->alias == NULL;

	if (need_alias_parens)
		deparseAppendStringInfoChar(state, '(');

	deparseTableRef(state, join_expr->larg);
	deparseAppendPart(state, true);

	if (join_expr->isNatural)
		deparseAppendStringInfoString(state, "NATURAL ");

	switch (join_expr->jointype)
	{
		case JOIN_INNER: /* matching tuple pairs only */
			if (!join_expr->isNatural && join_expr->quals == NULL && list_length(join_expr->usingClause) == 0)
				deparseAppendStringInfoString(state, "CROSS ");
			break;
		case JOIN_LEFT: /* pairs + unmatched LHS tuples */
			deparseAppendStringInfoString(state, "LEFT ");
			break;
		case JOIN_FULL: /* pairs + unmatched LHS + unmatched RHS */
			deparseAppendStringInfoString(state, "FULL ");
			break;
		case JOIN_RIGHT: /* pairs + unmatched RHS tuples */
			deparseAppendStringInfoString(state, "RIGHT ");
			break;
		case JOIN_SEMI:
		case JOIN_ANTI:
		case JOIN_RIGHT_SEMI:
		case JOIN_RIGHT_ANTI:
		case JOIN_UNIQUE_OUTER:
		case JOIN_UNIQUE_INNER:
			// Only used by the planner/executor, not seen in parser output
			Assert(false);
			break;
	}

	deparseAppendStringInfoString(state, "JOIN ");

	if (need_rarg_parens)
		deparseAppendStringInfoChar(state, '(');
	deparseTableRef(state, join_expr->rarg);
	if (need_rarg_parens)
		deparseAppendStringInfoChar(state, ')');
	deparseAppendStringInfoChar(state, ' ');

	if (join_expr->quals != NULL)
	{
		deparseAppendStringInfoString(state, "ON ");
		deparseExpr(state, join_expr->quals, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (list_length(join_expr->usingClause) > 0)
	{
		deparseAppendStringInfoString(state, "USING (");
		deparseNameList(state, join_expr->usingClause);
		deparseAppendStringInfoString(state, ") ");

		if (join_expr->join_using_alias)
		{
			deparseAppendStringInfoString(state, "AS ");
			deparseAppendStringInfoString(state, join_expr->join_using_alias->aliasname);
		}
	}

	if (need_alias_parens)
		deparseAppendStringInfoString(state, ") ");

	if (join_expr->alias != NULL)
		deparseAlias(state, join_expr->alias);

	removeTrailingSpace(state);
}

static void deparseCTESearchClause(DeparseState *state, CTESearchClause *search_clause)
{
	deparseAppendStringInfoString(state, " SEARCH ");
	if (search_clause->search_breadth_first)
		deparseAppendStringInfoString(state, "BREADTH ");
	else
		deparseAppendStringInfoString(state, "DEPTH ");

	deparseAppendStringInfoString(state, "FIRST BY ");

	if (search_clause->search_col_list)
		deparseColumnList(state, search_clause->search_col_list);

	deparseAppendStringInfoString(state, " SET ");
	deparseAppendStringInfoString(state, quote_identifier(search_clause->search_seq_column));
}

// "opt_cycle_clause" in gram.y
static void deparseCTECycleClause(DeparseState *state, CTECycleClause *cycle_clause)
{
	deparseAppendStringInfoString(state, " CYCLE ");

	if (cycle_clause->cycle_col_list)
		deparseColumnList(state, cycle_clause->cycle_col_list);

	deparseAppendStringInfoString(state, " SET ");
	deparseAppendStringInfoString(state, quote_identifier(cycle_clause->cycle_mark_column));

	if (cycle_clause->cycle_mark_value)
	{
		deparseAppendStringInfoString(state, " TO ");
		deparseAexprConst(state, cycle_clause->cycle_mark_value);
	}

	if (cycle_clause->cycle_mark_default)
	{
		deparseAppendStringInfoString(state, " DEFAULT ");
		deparseAexprConst(state, cycle_clause->cycle_mark_default);
	}

	deparseAppendStringInfoString(state, " USING ");
	deparseAppendStringInfoString(state, quote_identifier(cycle_clause->cycle_path_column));
}

static void deparseCommonTableExpr(DeparseState *state, CommonTableExpr *cte)
{
	deparseAppendCommentsIfNeeded(state, cte->location);

	deparseColId(state, cte->ctename);

	if (list_length(cte->aliascolnames) > 0)
	{
		deparseAppendStringInfoChar(state, '(');
		deparseNameList(state, cte->aliascolnames);
		deparseAppendStringInfoChar(state, ')');
	}
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "AS ");
	switch (cte->ctematerialized) {
		case CTEMaterializeDefault: /* no option specified */
			break;
		case CTEMaterializeAlways:
			deparseAppendStringInfoString(state, "MATERIALIZED ");
			break;
		case CTEMaterializeNever:
			deparseAppendStringInfoString(state, "NOT MATERIALIZED ");
			break;
	}

	deparseAppendStringInfoChar(state, '(');
	deparsePreparableStmt(state, cte->ctequery);
	deparseAppendStringInfoChar(state, ')');

	if (cte->search_clause)
		deparseCTESearchClause(state, cte->search_clause);
	if (cte->cycle_clause)
		deparseCTECycleClause(state, cte->cycle_clause);
}

static void deparseRangeSubselect(DeparseState *state, RangeSubselect *range_subselect)
{
	if (range_subselect->lateral)
		deparseAppendStringInfoString(state, "LATERAL ");

	deparseAppendStringInfoChar(state, '(');
	deparseSelectStmt(state, castNode(SelectStmt, range_subselect->subquery), DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ')');

	if (range_subselect->alias != NULL)
	{
		deparseAppendStringInfoChar(state, ' ');
		deparseAlias(state, range_subselect->alias);
	}
}

static void deparseRangeFunction(DeparseState *state, RangeFunction *range_func)
{
	ListCell *lc;
	ListCell *lc2;

	if (range_func->lateral)
		deparseAppendStringInfoString(state, "LATERAL ");

	if (range_func->is_rowsfrom)
	{
		deparseAppendStringInfoString(state, "ROWS FROM ");
		deparseAppendStringInfoChar(state, '(');
		foreach(lc, range_func->functions)
		{
			List *lfunc = castNode(List, lfirst(lc));
			Assert(list_length(lfunc) == 2);
			deparseFuncExprWindowless(state, linitial(lfunc));
			deparseAppendStringInfoChar(state, ' ');
			List *coldeflist = castNode(List, lsecond(lfunc));
			if (list_length(coldeflist) > 0)
			{
				deparseAppendStringInfoString(state, "AS (");
				foreach(lc2, coldeflist)
				{
					deparseColumnDef(state, castNode(ColumnDef, lfirst(lc2)));
					if (lnext(coldeflist, lc2))
						deparseAppendStringInfoString(state, ", ");
				}
				deparseAppendStringInfoChar(state, ')');
			}
			if (lnext(range_func->functions, lc))
				deparseAppendStringInfoString(state, ", ");
		}
		deparseAppendStringInfoChar(state, ')');
	}
	else
	{
		Assert(list_length(linitial(range_func->functions)) == 2);
		deparseFuncExprWindowless(state, linitial(linitial(range_func->functions)));
	}
	deparseAppendStringInfoChar(state, ' ');

	if (range_func->ordinality)
		deparseAppendStringInfoString(state, "WITH ORDINALITY ");

	if (range_func->alias != NULL)
	{
		deparseAlias(state, range_func->alias);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (list_length(range_func->coldeflist) > 0)
	{
		if (range_func->alias == NULL)
			deparseAppendStringInfoString(state, "AS ");
		deparseAppendStringInfoChar(state, '(');
		foreach(lc, range_func->coldeflist)
		{
			deparseColumnDef(state, castNode(ColumnDef, lfirst(lc)));
			if (lnext(range_func->coldeflist, lc))
				deparseAppendStringInfoString(state, ", ");
		}
		deparseAppendStringInfoChar(state, ')');
	}

	removeTrailingSpace(state);
}

static void deparseAArrayExpr(DeparseState *state, A_ArrayExpr *array_expr)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "ARRAY[");
	deparseExprList(state, array_expr->elements);
	deparseAppendStringInfoChar(state, ']');
}

static void deparseRowExpr(DeparseState *state, RowExpr *row_expr)
{
	ListCell *lc;

	switch (row_expr->row_format)
	{
		case COERCE_EXPLICIT_CALL:
			deparseAppendStringInfoString(state, "ROW");
			break;
		case COERCE_SQL_SYNTAX:
		case COERCE_EXPLICIT_CAST:
			// Not present in raw parser output
			Assert(false);
			break;
		case COERCE_IMPLICIT_CAST:
			// No prefix
			break;
	}

	deparseAppendStringInfoString(state, "(");
	deparseExprList(state, row_expr->args);
	deparseAppendStringInfoChar(state, ')');
}

static void deparseTypeCast(DeparseState *state, TypeCast *type_cast, DeparseNodeContext context)
{
	bool need_parens;

	if (type_cast->arg == NULL || type_cast->typeName == NULL)
		elog(ERROR, "deparse: unexpected NULL arg/typeName in TypeCast");

	// TYPECAST binds tighter than any other operator, so anything that isn't a "c_expr" needs parens
	need_parens = needsParensAsCExpr(type_cast->arg);

	if (context == DEPARSE_NODE_CONTEXT_FUNC_EXPR)
	{
		deparseAppendStringInfoString(state, "CAST(");
		deparseExpr(state, type_cast->arg, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, " AS ");
		deparseTypeName(state, type_cast->typeName);
		deparseAppendStringInfoChar(state, ')');
		return;
	}

	if (IsA(type_cast->arg, A_Const))
	{
		A_Const *a_const = castNode(A_Const, type_cast->arg);

		if (list_length(type_cast->typeName->names) == 2 &&
			strcmp(strVal(linitial(type_cast->typeName->names)), "pg_catalog") == 0)
		{
			char *typename = strVal(lsecond(type_cast->typeName->names));
			if (strcmp(typename, "bpchar") == 0 && type_cast->typeName->typmods == NULL)
			{
				deparseAppendStringInfoString(state, "char ");
				deparseAConst(state, a_const);
				return;
			}
			else if (strcmp(typename, "bool") == 0 && IsA(&a_const->val, String))
			{
				/*
				* Handle "bool" or "false" in the statement, which is represented as a typecast
				* (other boolean casts should be represented as a cast, i.e. don't need special handling)
				*/
				char *const_val = strVal(&a_const->val);
				if (strcmp(const_val, "t") == 0)
				{
					deparseAppendStringInfoString(state, "true");
					return;
				}
				if (strcmp(const_val, "f") == 0)
				{
					deparseAppendStringInfoString(state, "false");
					return;
				}
			}
			else if (strcmp(typename, "interval") == 0 && context == DEPARSE_NODE_CONTEXT_SET_STATEMENT && IsA(&a_const->val, String))
			{
				deparseAppendStringInfoString(state, "interval ");
				deparseAConst(state, a_const);
				deparseIntervalTypmods(state, type_cast->typeName);
				return;
			}
		}

		// Ensure negative values have wrapping parentheses
		if (IsA(&a_const->val, Float) || (IsA(&a_const->val, Integer) && intVal(&a_const->val) < 0))
		{
			need_parens = true;
		}

		if (list_length(type_cast->typeName->names) == 1 &&
			strcmp(strVal(linitial(type_cast->typeName->names)), "point") == 0 &&
			a_const->location > type_cast->typeName->location)
		{
			deparseAppendStringInfoString(state, " point ");
			deparseAConst(state, a_const);
			return;
		}
	}


	if (need_parens)
		deparseAppendStringInfoChar(state, '(');
	// Inside parens the arg follows "a_expr" rules, otherwise it is a "c_expr" (which is valid in any context)
	deparseExpr(state, type_cast->arg, need_parens ? DEPARSE_NODE_CONTEXT_A_EXPR : DEPARSE_NODE_CONTEXT_NONE);
	if (need_parens)
		deparseAppendStringInfoChar(state, ')');

	deparseAppendStringInfoString(state, "::");
	deparseTypeName(state, type_cast->typeName);
}

static void deparseTypeName(DeparseState *state, TypeName *type_name)
{
	if (type_name == NULL)
		elog(ERROR, "deparse: unexpected NULL TypeName");

	ListCell *lc;
	bool skip_typmods = false;

	if (type_name->setof)
		deparseAppendStringInfoString(state, "SETOF ");

	if (list_length(type_name->names) == 2 && strcmp(strVal(linitial(type_name->names)), "pg_catalog") == 0)
	{
		const char *name = strVal(lsecond(type_name->names));
		if (strcmp(name, "bpchar") == 0)
		{
			deparseAppendStringInfoString(state, "char");
		}
		else if (strcmp(name, "varchar") == 0)
		{
			deparseAppendStringInfoString(state, "varchar");
		}
		else if (strcmp(name, "numeric") == 0)
		{
			deparseAppendStringInfoString(state, "numeric");
		}
		else if (strcmp(name, "bool") == 0)
		{
			deparseAppendStringInfoString(state, "boolean");
		}
		else if (strcmp(name, "int2") == 0)
		{
			deparseAppendStringInfoString(state, "smallint");
		}
		else if (strcmp(name, "int4") == 0)
		{
			deparseAppendStringInfoString(state, "int");
		}
		else if (strcmp(name, "int8") == 0)
		{
			deparseAppendStringInfoString(state, "bigint");
		}
		else if (strcmp(name, "real") == 0 || strcmp(name, "float4") == 0)
		{
			deparseAppendStringInfoString(state, "real");
		}
		else if (strcmp(name, "float8") == 0)
		{
			deparseAppendStringInfoString(state, "double precision");
		}
		else if (strcmp(name, "time") == 0)
		{
			deparseAppendStringInfoString(state, "time");
		}
		else if (strcmp(name, "timetz") == 0)
		{
			deparseAppendStringInfoString(state, "time ");
			if (list_length(type_name->typmods) > 0)
			{
				deparseAppendStringInfoChar(state, '(');
				foreach(lc, type_name->typmods)
				{
					deparseSignedIconst(state, (Node *) &castNode(A_Const, lfirst(lc))->val);
					if (lnext(type_name->typmods, lc))
						deparseAppendStringInfoString(state, ", ");
				}
				deparseAppendStringInfoString(state, ") ");
			}
			deparseAppendStringInfoString(state, "with time zone");
			skip_typmods = true;
		}
		else if (strcmp(name, "timestamp") == 0)
		{
			deparseAppendStringInfoString(state, "timestamp");
		}
		else if (strcmp(name, "timestamptz") == 0)
		{
			deparseAppendStringInfoString(state, "timestamp ");
			if (list_length(type_name->typmods) > 0)
			{
				deparseAppendStringInfoChar(state, '(');
				foreach(lc, type_name->typmods)
				{
					deparseSignedIconst(state, (Node *) &castNode(A_Const, lfirst(lc))->val);
					if (lnext(type_name->typmods, lc))
						deparseAppendStringInfoString(state, ", ");
				}
				deparseAppendStringInfoString(state, ") ");
			}
			deparseAppendStringInfoString(state, "with time zone");
			skip_typmods = true;
		}
		else if (strcmp(name, "interval") == 0 && list_length(type_name->typmods) == 0)
		{
			deparseAppendStringInfoString(state, "interval");
		}
		else if (strcmp(name, "interval") == 0 && list_length(type_name->typmods) >= 1)
		{
			deparseAppendStringInfoString(state, "interval");
			deparseIntervalTypmods(state, type_name);

			skip_typmods = true;
		}
		else
		{
			deparseAppendStringInfoString(state, "pg_catalog.");
			deparseAppendStringInfoString(state, name);
		}
	}
	else
	{
		deparseAnyName(state, type_name->names);
	}

	if (list_length(type_name->typmods) > 0 && !skip_typmods)
	{
		deparseAppendStringInfoChar(state, '(');
		foreach(lc, type_name->typmods)
		{
			if (IsA(lfirst(lc), A_Const))
				deparseAConst(state, lfirst(lc));
			else if (IsA(lfirst(lc), ParamRef))
				deparseParamRef(state, lfirst(lc));
			else if (IsA(lfirst(lc), ColumnRef))
				deparseColumnRef(state, lfirst(lc));
			else
				Assert(false);

			if (lnext(type_name->typmods, lc))
				deparseAppendStringInfoString(state, ", ");
		}
		deparseAppendStringInfoChar(state, ')');
	}

	foreach(lc, type_name->arrayBounds)
	{
		deparseAppendStringInfoChar(state, '[');
		if (IsA(lfirst(lc), Integer) && intVal(lfirst(lc)) != -1)
			deparseSignedIconst(state, lfirst(lc));
		deparseAppendStringInfoChar(state, ']');
	}

	if (type_name->pct_type)
		deparseAppendStringInfoString(state, "%type");
}

// Handle typemods for Interval types separately
// so that they can be applied appropriately for different contexts.
// For example, when using `SET` a query like `INTERVAL 'x' hour TO minute`
// the `INTERVAL` keyword is specified first.
// In all other contexts, intervals use the `'x'::interval` style.
static void deparseIntervalTypmods(DeparseState *state, TypeName *type_name)
{
	const char *name = strVal(lsecond(type_name->names));
	Assert(strcmp(name, "interval") == 0);
	Assert(list_length(type_name->typmods) >= 1);
	Assert(IsA(linitial(type_name->typmods), A_Const));
	Assert(IsA(&castNode(A_Const, linitial(type_name->typmods))->val, Integer));

	int fields = intVal(&castNode(A_Const, linitial(type_name->typmods))->val);

	// This logic is based on intervaltypmodout in timestamp.c
	switch (fields)
	{
		case INTERVAL_MASK(YEAR):
			deparseAppendStringInfoString(state, " year");
			break;
		case INTERVAL_MASK(MONTH):
			deparseAppendStringInfoString(state, " month");
			break;
		case INTERVAL_MASK(DAY):
			deparseAppendStringInfoString(state, " day");
			break;
		case INTERVAL_MASK(HOUR):
			deparseAppendStringInfoString(state, " hour");
			break;
		case INTERVAL_MASK(MINUTE):
			deparseAppendStringInfoString(state, " minute");
			break;
		case INTERVAL_MASK(SECOND):
			deparseAppendStringInfoString(state, " second");
			break;
		case INTERVAL_MASK(YEAR) | INTERVAL_MASK(MONTH):
			deparseAppendStringInfoString(state, " year to month");
			break;
		case INTERVAL_MASK(DAY) | INTERVAL_MASK(HOUR):
			deparseAppendStringInfoString(state, " day to hour");
			break;
		case INTERVAL_MASK(DAY) | INTERVAL_MASK(HOUR) | INTERVAL_MASK(MINUTE):
			deparseAppendStringInfoString(state, " day to minute");
			break;
		case INTERVAL_MASK(DAY) | INTERVAL_MASK(HOUR) | INTERVAL_MASK(MINUTE) | INTERVAL_MASK(SECOND):
			deparseAppendStringInfoString(state, " day to second");
			break;
		case INTERVAL_MASK(HOUR) | INTERVAL_MASK(MINUTE):
			deparseAppendStringInfoString(state, " hour to minute");
			break;
		case INTERVAL_MASK(HOUR) | INTERVAL_MASK(MINUTE) | INTERVAL_MASK(SECOND):
			deparseAppendStringInfoString(state, " hour to second");
			break;
		case INTERVAL_MASK(MINUTE) | INTERVAL_MASK(SECOND):
			deparseAppendStringInfoString(state, " minute to second");
			break;
		case INTERVAL_FULL_RANGE:
			// Nothing
			break;
		default:
			Assert(false);
			break;
	}

	if (list_length(type_name->typmods) == 2)
	{
		int precision = intVal(&castNode(A_Const, lsecond(type_name->typmods))->val);
		if (precision != INTERVAL_FULL_PRECISION)
			deparseAppendStringInfo(state, "(%d)", precision);
	}
}

static void deparseNullTest(DeparseState *state, NullTest *null_test)
{
	// argisrow is always false in raw parser output
	Assert(null_test->argisrow == false);

	deparseExprOperand(state, (Node *) null_test->arg, DEPARSE_PREC_IS, DEPARSE_OPERAND_LEFT, DEPARSE_NODE_CONTEXT_A_EXPR);
	switch (null_test->nulltesttype)
	{
		case IS_NULL:
			deparseAppendStringInfoString(state, " IS NULL");
			break;
		case IS_NOT_NULL:
			deparseAppendStringInfoString(state, " IS NOT NULL");
			break;
	}
}

// "case_expr" in gram.y
static void deparseCaseExpr(DeparseState *state, CaseExpr *case_expr)
{
	ListCell *lc;
	DeparseStateNestingLevel *parent_level = NULL;

	deparseAppendStringInfoString(state, "CASE ");

	if (case_expr->arg != NULL)
	{
		deparseExpr(state, (Node *) case_expr->arg, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoChar(state, ' ');
	}

	parent_level = deparseStateIncreaseNestingLevel(state);

	foreach(lc, case_expr->args)
	{
		deparseCaseWhen(state, castNode(CaseWhen, lfirst(lc)));
	}

	if (case_expr->defresult != NULL)
	{
		deparseAppendPartGroup(state, "ELSE", DEPARSE_PART_INDENT);
		deparseExpr(state, (Node *) case_expr->defresult, DEPARSE_NODE_CONTEXT_A_EXPR);
	}

	deparseAppendStringInfoChar(state, ' ');
	deparseStateDecreaseNestingLevel(state, parent_level);

	deparseAppendStringInfoString(state, "END");
}

// "when_clause" in gram.y
static void deparseCaseWhen(DeparseState *state, CaseWhen *case_when)
{
	deparseAppendPartGroup(state, "WHEN", DEPARSE_PART_INDENT);
	deparseExpr(state, (Node *) case_when->expr, DEPARSE_NODE_CONTEXT_A_EXPR);
	deparseAppendStringInfoString(state, " THEN ");
	deparseExpr(state, (Node *) case_when->result, DEPARSE_NODE_CONTEXT_A_EXPR);
}

static void deparseAIndirection(DeparseState *state, A_Indirection *a_indirection)
{
	if (a_indirection->arg == NULL)
		elog(ERROR, "deparse: unexpected NULL arg in A_Indirection");

	ListCell *lc;
	/*
	 * Only a few "c_expr" rules in gram.y accept indirection without parens:
	 * "PARAM opt_indirection", "columnref" (whose field names are part of the
	 * ColumnRef itself, so only subscripts can follow directly) and
	 * "select_with_parens indirection". Everything else must use
	 * "'(' a_expr ')' opt_indirection".
	 */
	bool need_parens = !(
		IsA(a_indirection->arg, ParamRef) ||
		(IsA(a_indirection->arg, ColumnRef) && IsA(linitial(a_indirection->indirection), A_Indices)) ||
		(IsA(a_indirection->arg, SubLink) && castNode(SubLink, a_indirection->arg)->subLinkType == EXPR_SUBLINK));

	if (need_parens)
		deparseAppendStringInfoChar(state, '(');

	deparseExpr(state, a_indirection->arg, need_parens ? DEPARSE_NODE_CONTEXT_A_EXPR : DEPARSE_NODE_CONTEXT_NONE);

	if (need_parens)
		deparseAppendStringInfoChar(state, ')');

	deparseOptIndirection(state, a_indirection->indirection, 0);
}

static void deparseAIndices(DeparseState *state, A_Indices *a_indices)
{
	deparseAppendStringInfoChar(state, '[');
	if (a_indices->lidx != NULL)
		deparseExpr(state, a_indices->lidx, DEPARSE_NODE_CONTEXT_A_EXPR);
	if (a_indices->is_slice)
		deparseAppendStringInfoChar(state, ':');
	if (a_indices->uidx != NULL)
		deparseExpr(state, a_indices->uidx, DEPARSE_NODE_CONTEXT_A_EXPR);
	deparseAppendStringInfoChar(state, ']');
}

static void deparseCoalesceExpr(DeparseState *state, CoalesceExpr *coalesce_expr)
{
	deparseAppendStringInfoString(state, "COALESCE(");
	deparseExprList(state, coalesce_expr->args);
	deparseAppendStringInfoChar(state, ')');
}

static void deparseMinMaxExpr(DeparseState *state, MinMaxExpr *min_max_expr)
{
	switch (min_max_expr->op)
	{
		case IS_GREATEST:
			deparseAppendStringInfoString(state, "GREATEST(");
			break;
		case IS_LEAST:
			deparseAppendStringInfoString(state, "LEAST(");
			break;
	}
	deparseExprList(state, min_max_expr->args);
	deparseAppendStringInfoChar(state, ')');
}

static void deparseBooleanTest(DeparseState *state, BooleanTest *boolean_test)
{
	if (boolean_test->arg == NULL)
		elog(ERROR, "deparse: unexpected NULL arg in BooleanTest");

	deparseExprOperand(state, (Node *) boolean_test->arg, DEPARSE_PREC_IS, DEPARSE_OPERAND_LEFT, DEPARSE_NODE_CONTEXT_A_EXPR);

	switch (boolean_test->booltesttype)
	{
		case IS_TRUE:
			deparseAppendStringInfoString(state, " IS TRUE");
			break;
		case IS_NOT_TRUE:
			deparseAppendStringInfoString(state, " IS NOT TRUE");
			break;
		case IS_FALSE:
			deparseAppendStringInfoString(state, " IS FALSE");
			break;
		case IS_NOT_FALSE:
			deparseAppendStringInfoString(state, " IS NOT FALSE");
			break;
		case IS_UNKNOWN:
			deparseAppendStringInfoString(state, " IS UNKNOWN");
			break;
		case IS_NOT_UNKNOWN:
			deparseAppendStringInfoString(state, " IS NOT UNKNOWN");
			break;
		default:
			Assert(false);
	}
}

// "columnDef" and "alter_table_cmd" in gram.y
static void deparseColumnDef(DeparseState *state, ColumnDef *column_def)
{
	if (column_def == NULL)
		elog(ERROR, "deparse: unexpected NULL ColumnDef");

	ListCell *lc;

	if (column_def->colname != NULL)
	{
		deparseAppendStringInfoString(state, quote_identifier(column_def->colname));
		deparseAppendStringInfoChar(state, ' ');
	}

	if (column_def->typeName != NULL)
	{
		deparseTypeName(state, column_def->typeName);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (column_def->storage_name)
	{
		deparseAppendStringInfoString(state, "STORAGE ");
		deparseAppendStringInfoString(state, column_def->storage_name);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (column_def->raw_default != NULL)
	{
		deparseAppendStringInfoString(state, "USING ");
		deparseExpr(state, column_def->raw_default, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (column_def->compression != NULL)
	{
		deparseAppendStringInfoString(state, "COMPRESSION ");
		deparseAppendStringInfoString(state, column_def->compression);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (column_def->fdwoptions != NULL)
	{
		deparseCreateGenericOptions(state, column_def->fdwoptions);
		deparseAppendStringInfoChar(state, ' ');
	}

	foreach(lc, column_def->constraints)
	{
		deparseConstraint(state, castNode(Constraint, lfirst(lc)), DEPARSE_NODE_CONTEXT_NONE);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (column_def->collClause != NULL)
	{
		deparseCollateClause(state, column_def->collClause);
	}

	removeTrailingSpace(state);
}

// "returning_clause" and "returning_option" in gram.y
static void deparseReturningClause(DeparseState *state, ReturningClause *returning_clause)
{
	ListCell *lc;

	deparseAppendPartGroup(state, "RETURNING", DEPARSE_PART_INDENT_AND_MERGE);
	if (list_length(returning_clause->options) > 0)
	{
		deparseAppendStringInfoString(state, "WITH (");
		foreach(lc, returning_clause->options)
		{
			ReturningOption *opt = castNode(ReturningOption, lfirst(lc));
			switch (opt->option)
			{
				case RETURNING_OPTION_OLD:
					deparseAppendStringInfoString(state, "OLD AS ");
					break;
				case RETURNING_OPTION_NEW:
					deparseAppendStringInfoString(state, "NEW AS ");
					break;
			}
			deparseColId(state, opt->value);
			if (lnext(returning_clause->options, lc))
				deparseAppendStringInfoString(state, ", ");
		}
		deparseAppendStringInfoString(state, ") ");
	}
	deparseTargetList(state, returning_clause->exprs);
}

static void deparseInsertOverride(DeparseState *state, OverridingKind override)
{
	switch (override)
	{
		case OVERRIDING_NOT_SET:
			// Do nothing
			break;
		case OVERRIDING_USER_VALUE:
			deparseAppendStringInfoString(state, "OVERRIDING USER VALUE ");
			break;
		case OVERRIDING_SYSTEM_VALUE:
			deparseAppendStringInfoString(state, "OVERRIDING SYSTEM VALUE ");
			break;
	}
}

static void deparseInsertStmt(DeparseState *state, InsertStmt *insert_stmt)
{
	ListCell *lc;
	ListCell *lc2;
	DeparseStateNestingLevel *parent_level = deparseStateIncreaseNestingLevel(state);

	if (insert_stmt->withClause != NULL)
	{
		deparseWithClause(state, insert_stmt->withClause);
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseAppendPartGroup(state, "INSERT INTO", DEPARSE_PART_INDENT);
	deparseRangeVar(state, insert_stmt->relation, DEPARSE_NODE_CONTEXT_INSERT_RELATION);
	deparseAppendStringInfoChar(state, ' ');

	if (list_length(insert_stmt->cols) > 0)
	{
		deparseAppendStringInfoChar(state, '(');
		deparseInsertColumnList(state, insert_stmt->cols);
		deparseAppendStringInfoString(state, ") ");
	}

	deparseInsertOverride(state, insert_stmt->override);

	if (insert_stmt->selectStmt != NULL)
	{
		deparseSelectStmt(state, castNode(SelectStmt, insert_stmt->selectStmt), DEPARSE_NODE_CONTEXT_INSERT_SELECT);
		deparseAppendStringInfoChar(state, ' ');
	}
	else
	{
		deparseAppendStringInfoString(state, "DEFAULT VALUES ");
	}

	if (insert_stmt->onConflictClause != NULL)
	{
		deparseOnConflictClause(state, insert_stmt->onConflictClause);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (insert_stmt->returningClause != NULL)
		deparseReturningClause(state, insert_stmt->returningClause);

	removeTrailingSpace(state);
	deparseStateDecreaseNestingLevel(state, parent_level);
}

static void deparseInferClause(DeparseState *state, InferClause *infer_clause)
{
	ListCell *lc;

	if (list_length(infer_clause->indexElems) > 0)
	{
		deparseAppendStringInfoChar(state, '(');
		foreach(lc, infer_clause->indexElems)
		{
			deparseIndexElem(state, lfirst(lc));
			if (lnext(infer_clause->indexElems, lc))
				deparseAppendStringInfoString(state, ", ");
		}
		deparseAppendStringInfoString(state, ") ");
	}

	if (infer_clause->conname != NULL)
	{
		deparseAppendStringInfoString(state, "ON CONSTRAINT ");
		deparseAppendStringInfoString(state, quote_identifier(infer_clause->conname));
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseWhereClause(state, infer_clause->whereClause);

	removeTrailingSpace(state);
}

static void deparseOnConflictClause(DeparseState *state, OnConflictClause *on_conflict_clause)
{
	ListCell *lc;

	deparseAppendPartGroup(state, "ON CONFLICT", DEPARSE_PART_INDENT);

	if (on_conflict_clause->infer != NULL)
	{
		deparseInferClause(state, on_conflict_clause->infer);
		deparseAppendStringInfoChar(state, ' ');
	}

	switch (on_conflict_clause->action)
	{
		case ONCONFLICT_NONE:
			Assert(false);
			break;
		case ONCONFLICT_NOTHING:
			deparseAppendStringInfoString(state, "DO NOTHING ");
			break;
		case ONCONFLICT_UPDATE:
			deparseAppendStringInfoString(state, "DO UPDATE ");
			break;
	}

	if (list_length(on_conflict_clause->targetList) > 0)
	{
		deparseAppendPartGroup(state, "SET", DEPARSE_PART_INDENT);
		deparseSetClauseList(state, on_conflict_clause->targetList);
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseWhereClause(state, on_conflict_clause->whereClause);

	removeTrailingSpace(state);
}

static void deparseUpdateStmt(DeparseState *state, UpdateStmt *update_stmt)
{
	ListCell* lc;
	ListCell* lc2;
	ListCell* lc3;
	DeparseStateNestingLevel *parent_level = deparseStateIncreaseNestingLevel(state);

	if (update_stmt->withClause != NULL)
	{
		deparseWithClause(state, update_stmt->withClause);
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseAppendPartGroup(state, "UPDATE", DEPARSE_PART_INDENT);
	deparseRangeVar(state, update_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ' ');

	if (list_length(update_stmt->targetList) > 0)
	{
		deparseAppendPartGroup(state, "SET", DEPARSE_PART_INDENT);
		deparseSetClauseList(state, update_stmt->targetList);
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseFromClause(state, update_stmt->fromClause);
	deparseWhereOrCurrentClause(state, update_stmt->whereClause);

	if (update_stmt->returningClause != NULL)
		deparseReturningClause(state, update_stmt->returningClause);

	removeTrailingSpace(state);
	deparseStateDecreaseNestingLevel(state, parent_level);
}

// "MergeStmt" in gram.y
static void deparseMergeStmt(DeparseState *state, MergeStmt *merge_stmt)
{
	DeparseStateNestingLevel *parent_level = deparseStateIncreaseNestingLevel(state);

	if (merge_stmt->withClause != NULL)
	{
		deparseWithClause(state, merge_stmt->withClause);
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseAppendPartGroup(state, "MERGE", DEPARSE_PART_INDENT);
	deparseAppendStringInfoString(state, "INTO ");
	deparseRangeVar(state, merge_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendPartGroup(state, "USING", DEPARSE_PART_INDENT);
	deparseTableRef(state, merge_stmt->sourceRelation);
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "ON ");
	deparseExpr(state, merge_stmt->joinCondition, DEPARSE_NODE_CONTEXT_A_EXPR);
	deparseAppendStringInfoChar(state, ' ');

	ListCell *lc;
	foreach (lc, merge_stmt->mergeWhenClauses)
	{
		MergeWhenClause *clause = castNode(MergeWhenClause, lfirst(lc));

		deparseAppendStringInfoString(state, "WHEN ");

		switch (clause->matchKind)
		{
			case MERGE_WHEN_MATCHED:
				deparseAppendStringInfoString(state, "MATCHED ");
				break;
			case MERGE_WHEN_NOT_MATCHED_BY_SOURCE:
				deparseAppendStringInfoString(state, "NOT MATCHED BY SOURCE ");
				break;
			case MERGE_WHEN_NOT_MATCHED_BY_TARGET:
				deparseAppendStringInfoString(state, "NOT MATCHED ");
				break;
		}

		if (clause->condition)
		{
			deparseAppendStringInfoString(state, "AND ");
			deparseExpr(state, clause->condition, DEPARSE_NODE_CONTEXT_A_EXPR);
			deparseAppendStringInfoChar(state, ' ');
		}

		deparseAppendStringInfoString(state, "THEN ");

		switch (clause->commandType) {
			case CMD_INSERT:
				deparseAppendStringInfoString(state, "INSERT ");

				if (clause->targetList) {
					deparseAppendStringInfoChar(state, '(');
					deparseInsertColumnList(state, clause->targetList);
					deparseAppendStringInfoString(state, ") ");
				}

				deparseInsertOverride(state, clause->override);

				if (clause->values) {
					deparseAppendStringInfoString(state, "VALUES (");
					deparseExprList(state, clause->values);
					deparseAppendStringInfoString(state, ")");
				} else {
					deparseAppendStringInfoString(state, "DEFAULT VALUES ");
				}

				break;
			case CMD_UPDATE:
				deparseAppendStringInfoString(state, "UPDATE SET ");
				deparseSetClauseList(state, clause->targetList);
				break;
			case CMD_DELETE:
				deparseAppendStringInfoString(state, "DELETE");
				break;
			case CMD_NOTHING:
				deparseAppendStringInfoString(state, "DO NOTHING");
				break;
			default:
				elog(ERROR, "deparse: unpermitted command type in merge statement: %d", clause->commandType);
				break;
		}

		deparseAppendStringInfoChar(state, ' ');
	}

	if (merge_stmt->returningClause != NULL)
		deparseReturningClause(state, merge_stmt->returningClause);

	removeTrailingSpace(state);
	deparseStateDecreaseNestingLevel(state, parent_level);
}

static void deparseDeleteStmt(DeparseState *state, DeleteStmt *delete_stmt)
{
	DeparseStateNestingLevel *parent_level = deparseStateIncreaseNestingLevel(state);

	if (delete_stmt->withClause != NULL)
	{
		deparseWithClause(state, delete_stmt->withClause);
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseAppendPartGroup(state, "DELETE", DEPARSE_PART_INDENT);
	deparseAppendStringInfoString(state, "FROM ");
	deparseRangeVar(state, delete_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ' ');

	if (delete_stmt->usingClause != NULL)
	{
		deparseAppendPartGroup(state, "USING", DEPARSE_PART_INDENT);
		deparseFromList(state, delete_stmt->usingClause);
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseWhereOrCurrentClause(state, delete_stmt->whereClause);

	if (delete_stmt->returningClause != NULL)
		deparseReturningClause(state, delete_stmt->returningClause);

	removeTrailingSpace(state);
	deparseStateDecreaseNestingLevel(state, parent_level);
}

static void deparseLockingClause(DeparseState *state, LockingClause *locking_clause)
{
	ListCell *lc;

	switch (locking_clause->strength)
	{
		case LCS_NONE:
			/* no such clause - only used in PlanRowMark */
			Assert(false);
			break;
		case LCS_FORKEYSHARE:
			deparseAppendPartGroup(state, "FOR KEY SHARE", DEPARSE_PART_INDENT);
			break;
		case LCS_FORSHARE:
			deparseAppendPartGroup(state, "FOR SHARE", DEPARSE_PART_INDENT);
			break;
		case LCS_FORNOKEYUPDATE:
			deparseAppendPartGroup(state, "FOR NO KEY UPDATE", DEPARSE_PART_INDENT);
			break;
		case LCS_FORUPDATE:
			deparseAppendPartGroup(state, "FOR UPDATE", DEPARSE_PART_INDENT);
			break;
	}

	if (list_length(locking_clause->lockedRels) > 0)
	{
		deparseAppendStringInfoString(state, "OF ");
		deparseQualifiedNameList(state, locking_clause->lockedRels);
		deparseAppendStringInfoChar(state, ' ');
	}

	switch (locking_clause->waitPolicy)
	{
		case LockWaitError:
			deparseAppendStringInfoString(state, "NOWAIT");
			break;
		case LockWaitSkip:
			deparseAppendStringInfoString(state, "SKIP LOCKED");
			break;
		case LockWaitBlock:
			// Default
			break;
	}

	removeTrailingSpace(state);
}

static void deparseSetToDefault(DeparseState *state, SetToDefault *set_to_default)
{
	deparseAppendStringInfoString(state, "DEFAULT");
}

static void deparseCreateCastStmt(DeparseState *state, CreateCastStmt *create_cast_stmt)
{
	ListCell *lc;
	ListCell *lc2;

	deparseAppendStringInfoString(state, "CREATE CAST (");
	deparseTypeName(state, create_cast_stmt->sourcetype);
	deparseAppendStringInfoString(state, " AS ");
	deparseTypeName(state, create_cast_stmt->targettype);
	deparseAppendStringInfoString(state, ") ");

	if (create_cast_stmt->func != NULL)
	{
		deparseAppendStringInfoString(state, "WITH FUNCTION ");
		deparseFunctionWithArgtypes(state, create_cast_stmt->func);
		deparseAppendStringInfoChar(state, ' ');
	}
	else if (create_cast_stmt->inout)
	{
		deparseAppendStringInfoString(state, "WITH INOUT ");
	}
	else
	{
		deparseAppendStringInfoString(state, "WITHOUT FUNCTION ");
	}

	switch (create_cast_stmt->context)
	{
		case COERCION_IMPLICIT:
			deparseAppendStringInfoString(state, "AS IMPLICIT");
			break;
		case COERCION_ASSIGNMENT:
			deparseAppendStringInfoString(state, "AS ASSIGNMENT");
			break;
		case COERCION_PLPGSQL:
			// Not present in raw parser output
			Assert(false);
			break;
		case COERCION_EXPLICIT:
			// Default
			break;
	}
}

static void deparseCreateOpClassStmt(DeparseState *state, CreateOpClassStmt *create_op_class_stmt)
{
	ListCell *lc = NULL;

	deparseAppendStringInfoString(state, "CREATE OPERATOR CLASS ");

	deparseAnyName(state, create_op_class_stmt->opclassname);
	deparseAppendStringInfoChar(state, ' ');

	if (create_op_class_stmt->isDefault)
		deparseAppendStringInfoString(state, "DEFAULT ");

	deparseAppendStringInfoString(state, "FOR TYPE ");
	deparseTypeName(state, create_op_class_stmt->datatype);
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "USING ");
	deparseAppendStringInfoString(state, quote_identifier(create_op_class_stmt->amname));
	deparseAppendStringInfoChar(state, ' ');

	if (create_op_class_stmt->opfamilyname != NULL)
	{
		deparseAppendStringInfoString(state, "FAMILY ");
		deparseAnyName(state, create_op_class_stmt->opfamilyname);
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseAppendStringInfoString(state, "AS ");
	deparseOpclassItemList(state, create_op_class_stmt->items);
}

static void deparseCreateOpFamilyStmt(DeparseState *state, CreateOpFamilyStmt *create_op_family_stmt)
{
	deparseAppendStringInfoString(state, "CREATE OPERATOR FAMILY ");

	deparseAnyName(state, create_op_family_stmt->opfamilyname);
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "USING ");
	deparseAppendStringInfoString(state, quote_identifier(create_op_family_stmt->amname));
}

static void deparseCreateOpClassItem(DeparseState *state, CreateOpClassItem *create_op_class_item)
{
	ListCell *lc = NULL;

	switch (create_op_class_item->itemtype)
	{
		case OPCLASS_ITEM_OPERATOR:
			deparseAppendStringInfoString(state, "OPERATOR ");
			deparseAppendStringInfo(state, "%d ", create_op_class_item->number);

			if (create_op_class_item->name != NULL)
			{
				if (create_op_class_item->name->objargs != NULL)
					deparseOperatorWithArgtypes(state, create_op_class_item->name);
				else
					deparseAnyOperator(state, create_op_class_item->name->objname);
				deparseAppendStringInfoChar(state, ' ');
			}

			if (create_op_class_item->order_family != NULL)
			{
				deparseAppendStringInfoString(state, "FOR ORDER BY ");
				deparseAnyName(state, create_op_class_item->order_family);
			}

			if (create_op_class_item->class_args != NULL)
			{
				deparseAppendStringInfoChar(state, '(');
				deparseTypeList(state, create_op_class_item->class_args);
				deparseAppendStringInfoChar(state, ')');
			}
			removeTrailingSpace(state);
			break;
		case OPCLASS_ITEM_FUNCTION:
			deparseAppendStringInfoString(state, "FUNCTION ");
			deparseAppendStringInfo(state, "%d ", create_op_class_item->number);
			if (create_op_class_item->class_args != NULL)
			{
				deparseAppendStringInfoChar(state, '(');
				deparseTypeList(state, create_op_class_item->class_args);
				deparseAppendStringInfoString(state, ") ");
			}
			if (create_op_class_item->name != NULL)
				deparseFunctionWithArgtypes(state, create_op_class_item->name);
			removeTrailingSpace(state);
			break;
		case OPCLASS_ITEM_STORAGETYPE:
			deparseAppendStringInfoString(state, "STORAGE ");
			deparseTypeName(state, create_op_class_item->storedtype);
			break;
		default:
			Assert(false);
	}
}

static void deparseTableLikeClause(DeparseState *state, TableLikeClause *table_like_clause)
{
	deparseAppendStringInfoString(state, "LIKE ");
	deparseRangeVar(state, table_like_clause->relation, DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ' ');

	if (table_like_clause->options == CREATE_TABLE_LIKE_ALL)
		deparseAppendStringInfoString(state, "INCLUDING ALL ");
	else
	{
		if (table_like_clause->options & CREATE_TABLE_LIKE_COMMENTS)
			deparseAppendStringInfoString(state, "INCLUDING COMMENTS ");
		if (table_like_clause->options & CREATE_TABLE_LIKE_COMPRESSION)
			deparseAppendStringInfoString(state, "INCLUDING COMPRESSION ");
		if (table_like_clause->options & CREATE_TABLE_LIKE_CONSTRAINTS)
			deparseAppendStringInfoString(state, "INCLUDING CONSTRAINTS ");
		if (table_like_clause->options & CREATE_TABLE_LIKE_DEFAULTS)
			deparseAppendStringInfoString(state, "INCLUDING DEFAULTS ");
		if (table_like_clause->options & CREATE_TABLE_LIKE_IDENTITY)
			deparseAppendStringInfoString(state, "INCLUDING IDENTITY ");
		if (table_like_clause->options & CREATE_TABLE_LIKE_GENERATED)
			deparseAppendStringInfoString(state, "INCLUDING GENERATED ");
		if (table_like_clause->options & CREATE_TABLE_LIKE_INDEXES)
			deparseAppendStringInfoString(state, "INCLUDING INDEXES ");
		if (table_like_clause->options & CREATE_TABLE_LIKE_STATISTICS)
			deparseAppendStringInfoString(state, "INCLUDING STATISTICS ");
		if (table_like_clause->options & CREATE_TABLE_LIKE_STORAGE)
			deparseAppendStringInfoString(state, "INCLUDING STORAGE ");
	}
	removeTrailingSpace(state);
}

static void deparseCreateDomainStmt(DeparseState *state, CreateDomainStmt *create_domain_stmt)
{
	ListCell *lc;

	Assert(create_domain_stmt->typeName != NULL);

	deparseAppendStringInfoString(state, "CREATE DOMAIN ");
	deparseAnyName(state, create_domain_stmt->domainname);
	deparseAppendStringInfoString(state, " AS ");

	deparseTypeName(state, create_domain_stmt->typeName);
	deparseAppendStringInfoChar(state, ' ');

	if (create_domain_stmt->collClause != NULL)
	{
		deparseCollateClause(state, create_domain_stmt->collClause);
		deparseAppendStringInfoChar(state, ' ');
	}

	foreach(lc, create_domain_stmt->constraints)
	{
		deparseConstraint(state, castNode(Constraint, lfirst(lc)), DEPARSE_NODE_CONTEXT_NONE);
		deparseAppendStringInfoChar(state, ' ');
	}

	removeTrailingSpace(state);
}

static void deparseCreateExtensionStmt(DeparseState *state, CreateExtensionStmt *create_extension_stmt)
{
	ListCell *lc = NULL;

	deparseAppendStringInfoString(state, "CREATE EXTENSION ");

	if (create_extension_stmt->if_not_exists)
		deparseAppendStringInfoString(state, "IF NOT EXISTS ");

	deparseColId(state, create_extension_stmt->extname);
	deparseAppendStringInfoChar(state, ' ');

	foreach (lc, create_extension_stmt->options)
	{
		DefElem *def_elem = castNode(DefElem, lfirst(lc));

		if (strcmp(def_elem->defname, "schema") == 0)
		{
			deparseAppendStringInfoString(state, "SCHEMA ");
			deparseColId(state, strVal(def_elem->arg));
		}
		else if (strcmp(def_elem->defname, "new_version") == 0)
		{
			deparseAppendStringInfoString(state, "VERSION ");
			deparseNonReservedWordOrSconst(state, strVal(def_elem->arg));
		}
		else if (strcmp(def_elem->defname, "cascade") == 0)
		{
			deparseAppendStringInfoString(state, "CASCADE");
		}
		else
		{
			Assert(false);
		}

		deparseAppendStringInfoChar(state, ' ');
	}

	removeTrailingSpace(state);
}

// "ColConstraintElem" and "ConstraintElem" in gram.y
static void deparseConstraint(DeparseState *state, Constraint *constraint, DeparseNodeContext context)
{
	if (constraint == NULL)
		elog(ERROR, "deparse: unexpected NULL Constraint");

	ListCell *lc;

	if (constraint->conname != NULL)
	{
		deparseAppendStringInfoString(state, "CONSTRAINT ");
		deparseAppendStringInfoString(state, quote_identifier(constraint->conname));
		deparseAppendStringInfoChar(state, ' ');
	}

	switch (constraint->contype) {
		case CONSTR_NULL:
			deparseAppendStringInfoString(state, "NULL ");
			break;
		case CONSTR_NOTNULL:
			deparseAppendStringInfoString(state, "NOT NULL ");
			break;
		case CONSTR_DEFAULT:
			deparseAppendStringInfoString(state, "DEFAULT ");
			deparseBExpr(state, constraint->raw_expr);
			break;
		case CONSTR_IDENTITY:
			deparseAppendStringInfoString(state, "GENERATED ");
			switch (constraint->generated_when)
			{
				case ATTRIBUTE_IDENTITY_ALWAYS:
					deparseAppendStringInfoString(state, "ALWAYS ");
					break;
				case ATTRIBUTE_IDENTITY_BY_DEFAULT:
					deparseAppendStringInfoString(state, "BY DEFAULT ");
					break;
				default:
					Assert(false);
			}
			deparseAppendStringInfoString(state, "AS IDENTITY ");
			deparseOptParenthesizedSeqOptList(state, constraint->options);
			break;
		case CONSTR_GENERATED:
			Assert(constraint->generated_when == ATTRIBUTE_IDENTITY_ALWAYS);
			deparseAppendStringInfoString(state, "GENERATED ALWAYS AS (");
			deparseExpr(state, constraint->raw_expr, DEPARSE_NODE_CONTEXT_A_EXPR);
			deparseAppendStringInfoString(state, ") ");
			if (constraint->generated_kind == ATTRIBUTE_GENERATED_STORED)
				deparseAppendStringInfoString(state, "STORED ");
			else
				deparseAppendStringInfoString(state, "VIRTUAL ");
			break;
		case CONSTR_CHECK:
			deparseAppendStringInfoString(state, "CHECK (");
			deparseExpr(state, constraint->raw_expr, DEPARSE_NODE_CONTEXT_A_EXPR);
			deparseAppendStringInfoString(state, ") ");
			break;
		case CONSTR_PRIMARY:
			deparseAppendStringInfoString(state, "PRIMARY KEY ");
			break;
		case CONSTR_UNIQUE:
			deparseAppendStringInfoString(state, "UNIQUE ");
			if (constraint->nulls_not_distinct)
				deparseAppendStringInfoString(state, "NULLS NOT DISTINCT ");
			break;
		case CONSTR_EXCLUSION:
			deparseAppendStringInfoString(state, "EXCLUDE ");
			if (strcmp(constraint->access_method, DEFAULT_INDEX_TYPE) != 0)
			{
				deparseAppendStringInfoString(state, "USING ");
				deparseAppendStringInfoString(state, quote_identifier(constraint->access_method));
				deparseAppendStringInfoChar(state, ' ');
			}
			deparseAppendStringInfoChar(state, '(');
			foreach(lc, constraint->exclusions)
			{
				List *exclusion = castNode(List, lfirst(lc));
				Assert(list_length(exclusion) == 2);
				deparseIndexElem(state, castNode(IndexElem, linitial(exclusion)));
				deparseAppendStringInfoString(state, " WITH ");
				deparseAnyOperator(state, castNode(List, lsecond(exclusion)));
				if (lnext(constraint->exclusions, lc))
					deparseAppendStringInfoString(state, ", ");
			}
			deparseAppendStringInfoString(state, ") ");
			if (constraint->where_clause != NULL)
			{
				deparseAppendStringInfoString(state, "WHERE (");
				deparseExpr(state, constraint->where_clause, DEPARSE_NODE_CONTEXT_A_EXPR);
				deparseAppendStringInfoString(state, ") ");
			}
			break;
		case CONSTR_FOREIGN:
			if (list_length(constraint->fk_attrs) > 0)
				deparseAppendStringInfoString(state, "FOREIGN KEY ");
			break;
		case CONSTR_ATTR_DEFERRABLE:
			deparseAppendStringInfoString(state, "DEFERRABLE ");
			break;
		case CONSTR_ATTR_NOT_DEFERRABLE:
			deparseAppendStringInfoString(state, "NOT DEFERRABLE ");
			break;
		case CONSTR_ATTR_DEFERRED:
			deparseAppendStringInfoString(state, "INITIALLY DEFERRED ");
			break;
		case CONSTR_ATTR_IMMEDIATE:
			deparseAppendStringInfoString(state, "INITIALLY IMMEDIATE ");
			break;
		case CONSTR_ATTR_ENFORCED:
			deparseAppendStringInfoString(state, "ENFORCED ");
			break;
		case CONSTR_ATTR_NOT_ENFORCED:
			deparseAppendStringInfoString(state, "NOT ENFORCED ");
			break;
	}

	if (list_length(constraint->keys) > 0)
	{
		bool emit_keys = true;
		bool needs_parens = true;

		if (list_length(constraint->keys) == 1)
		{
			Node* key = linitial(constraint->keys);
			if (context == DEPARSE_NODE_CONTEXT_ALTER_DOMAIN)
				emit_keys = !(IsA(key, String) && strcmp(castNode(String, key)->sval, "value") == 0);
			needs_parens = constraint->contype != CONSTR_NULL && constraint->contype != CONSTR_NOTNULL;
		}

		if (needs_parens)
			deparseAppendStringInfoChar(state, '(');

		if (emit_keys)
			deparseColumnList(state, constraint->keys);

		if (constraint->without_overlaps)
			deparseAppendStringInfoString(state, " WITHOUT OVERLAPS");

		if (needs_parens)
			deparseAppendStringInfoChar(state, ')');

		deparseAppendStringInfoChar(state, ' ');
	}

	if (list_length(constraint->fk_attrs) > 0)
	{
		deparseAppendStringInfoChar(state, '(');
		if (constraint->fk_with_period)
			deparseColumnListWithPeriod(state, constraint->fk_attrs);
		else
			deparseColumnList(state, constraint->fk_attrs);
		deparseAppendStringInfoString(state, ") ");
	}

	if (constraint->pktable != NULL)
	{
		deparseAppendStringInfoString(state, "REFERENCES ");
		deparseRangeVar(state, constraint->pktable, DEPARSE_NODE_CONTEXT_NONE);
		deparseAppendStringInfoChar(state, ' ');
		if (list_length(constraint->pk_attrs) > 0)
		{
			deparseAppendStringInfoChar(state, '(');
			if (constraint->pk_with_period)
				deparseColumnListWithPeriod(state, constraint->pk_attrs);
			else
				deparseColumnList(state, constraint->pk_attrs);
			deparseAppendStringInfoString(state, ") ");
		}
	}

	switch (constraint->fk_matchtype)
	{
		case FKCONSTR_MATCH_SIMPLE:
			// Default
			break;
		case FKCONSTR_MATCH_FULL:
			deparseAppendStringInfoString(state, "MATCH FULL ");
			break;
		case FKCONSTR_MATCH_PARTIAL:
			// Not implemented in Postgres
			Assert(false);
			break;
		default:
			// Not specified
			break;
	}

	switch (constraint->fk_upd_action)
	{
		case FKCONSTR_ACTION_NOACTION:
			// Default
			break;
		case FKCONSTR_ACTION_RESTRICT:
			deparseAppendStringInfoString(state, "ON UPDATE RESTRICT ");
			break;
		case FKCONSTR_ACTION_CASCADE:
			deparseAppendStringInfoString(state, "ON UPDATE CASCADE ");
			break;
		case FKCONSTR_ACTION_SETNULL:
			deparseAppendStringInfoString(state, "ON UPDATE SET NULL ");
			break;
		case FKCONSTR_ACTION_SETDEFAULT:
			deparseAppendStringInfoString(state, "ON UPDATE SET DEFAULT ");
			break;
		default:
			// Not specified
			break;
	}

	switch (constraint->fk_del_action)
	{
		case FKCONSTR_ACTION_NOACTION:
			// Default
			break;
		case FKCONSTR_ACTION_RESTRICT:
			deparseAppendStringInfoString(state, "ON DELETE RESTRICT ");
			break;
		case FKCONSTR_ACTION_CASCADE:
			deparseAppendStringInfoString(state, "ON DELETE CASCADE ");
			break;
		case FKCONSTR_ACTION_SETNULL:
		case FKCONSTR_ACTION_SETDEFAULT:
			deparseAppendStringInfoString(state, "ON DELETE SET ");

			switch (constraint->fk_del_action) {
				case FKCONSTR_ACTION_SETDEFAULT: deparseAppendStringInfoString(state, "DEFAULT "); break;
				case FKCONSTR_ACTION_SETNULL:    deparseAppendStringInfoString(state, "NULL "); break;
			}

			if (constraint->fk_del_set_cols) {
				deparseAppendStringInfoString(state, "(");
				ListCell *lc;
				foreach (lc, constraint->fk_del_set_cols) {
					deparseAppendStringInfoString(state, strVal(lfirst(lc)));
					if (lfirst(lc) != llast(constraint->fk_del_set_cols))
						deparseAppendStringInfoString(state, ", ");
				}
				deparseAppendStringInfoString(state, ")");
			}
			break;
		default:
			// Not specified
			break;
	}

	if (list_length(constraint->including) > 0)
	{
		deparseAppendStringInfoString(state, "INCLUDE (");
		deparseColumnList(state, constraint->including);
		deparseAppendStringInfoString(state, ") ");
	}

	switch (constraint->contype)
	{
		case CONSTR_PRIMARY:
		case CONSTR_UNIQUE:
		case CONSTR_EXCLUSION:
			deparseOptWith(state, constraint->options);
			break;
		default:
			break;
	}

	if (constraint->indexname != NULL)
		deparseAppendStringInfo(state, "USING INDEX %s ", quote_identifier(constraint->indexname));

	if (constraint->indexspace != NULL)
		deparseAppendStringInfo(state, "USING INDEX TABLESPACE %s ", quote_identifier(constraint->indexspace));

	if (constraint->deferrable)
		deparseAppendStringInfoString(state, "DEFERRABLE ");

	if (constraint->initdeferred)
		deparseAppendStringInfoString(state, "INITIALLY DEFERRED ");

	if (constraint->is_no_inherit)
		deparseAppendStringInfoString(state, "NO INHERIT ");

	if ((constraint->contype == CONSTR_FOREIGN || constraint->contype == CONSTR_CHECK) && !constraint->is_enforced)
		deparseAppendStringInfoString(state, "NOT ENFORCED ");

	if (constraint->skip_validation)
		deparseAppendStringInfoString(state, "NOT VALID ");
	
	removeTrailingSpace(state);
}

// "ALTER CONSTRAINT name ConstraintAttributeSpec" in gram.y
static void deparseATAlterConstraint(DeparseState *state, ATAlterConstraint *constraint)
{
	ListCell *lc;

	if (constraint->conname != NULL)
	{
		deparseAppendStringInfoString(state, "CONSTRAINT ");
		deparseAppendStringInfoString(state, quote_identifier(constraint->conname));
		deparseAppendStringInfoChar(state, ' ');
	}

	if (constraint->alterEnforceability)
	{
		if (constraint->is_enforced)
			deparseAppendStringInfoString(state, "ENFORCED ");
		else
			deparseAppendStringInfoString(state, "NOT ENFORCED ");
	}

	if (constraint->alterDeferrability)
	{
		if (constraint->initdeferred)
			deparseAppendStringInfoString(state, "INITIALLY DEFERRED ");
		else if (constraint->deferrable)
			deparseAppendStringInfoString(state, "DEFERRABLE ");
		else
			deparseAppendStringInfoString(state, "NOT DEFERRABLE ");
	}

	if (constraint->alterInheritability)
	{
		if (constraint->noinherit)
			deparseAppendStringInfoString(state, "NO INHERIT ");
		else
			deparseAppendStringInfoString(state, "INHERIT ");
	}

	removeTrailingSpace(state);
}

// "ReturnStmt" in gram.y
static void deparseReturnStmt(DeparseState *state, ReturnStmt *return_stmt)
{
	deparseAppendStringInfoString(state, "RETURN ");
	deparseExpr(state, return_stmt->returnval, DEPARSE_NODE_CONTEXT_A_EXPR);
}

static void deparseCreateFunctionStmt(DeparseState *state, CreateFunctionStmt *create_function_stmt)
{
	ListCell *lc;
	bool tableFunc = false;

	deparseAppendStringInfoString(state, "CREATE ");
	if (create_function_stmt->replace)
		deparseAppendStringInfoString(state, "OR REPLACE ");
	if (create_function_stmt->is_procedure)
		deparseAppendStringInfoString(state, "PROCEDURE ");
	else
		deparseAppendStringInfoString(state, "FUNCTION ");

	deparseFuncName(state, create_function_stmt->funcname);

	deparseAppendStringInfoChar(state, '(');
	foreach(lc, create_function_stmt->parameters)
	{
		FunctionParameter *function_parameter = castNode(FunctionParameter, lfirst(lc));
		if (function_parameter->mode != FUNC_PARAM_TABLE)
		{
			deparseFunctionParameter(state, function_parameter);
			if (lnext(create_function_stmt->parameters, lc) && castNode(FunctionParameter, lfirst(lnext(create_function_stmt->parameters, lc)))->mode != FUNC_PARAM_TABLE)
				deparseAppendStringInfoString(state, ", ");
		}
		else
		{
			tableFunc = true;
		}
	}
	deparseAppendStringInfoString(state, ") ");

	if (tableFunc)
	{
		deparseAppendStringInfoString(state, "RETURNS TABLE (");
		foreach(lc, create_function_stmt->parameters)
		{
			FunctionParameter *function_parameter = castNode(FunctionParameter, lfirst(lc));
			if (function_parameter->mode == FUNC_PARAM_TABLE)
			{
				deparseFunctionParameter(state, function_parameter);
				if (lnext(create_function_stmt->parameters, lc))
					deparseAppendStringInfoString(state, ", ");
			}
		}
		deparseAppendStringInfoString(state, ") ");
	}
	else if (create_function_stmt->returnType != NULL)
	{
		deparseAppendStringInfoString(state, "RETURNS ");
		deparseTypeName(state, create_function_stmt->returnType);
		deparseAppendStringInfoChar(state, ' ');
	}

	foreach(lc, create_function_stmt->options)
	{
		deparseCreateFuncOptItem(state, castNode(DefElem, lfirst(lc)));
		deparseAppendStringInfoChar(state, ' ');
	}

	if (create_function_stmt->sql_body)
	{
		/* RETURN or BEGIN ... END
		 */
		if (IsA(create_function_stmt->sql_body, ReturnStmt))
		{
			deparseReturnStmt(state, castNode(ReturnStmt, create_function_stmt->sql_body));
		}
		else
		{
			deparseAppendStringInfoString(state, "BEGIN ATOMIC ");
			if (IsA(create_function_stmt->sql_body, List), linitial((List *) create_function_stmt->sql_body) != NULL)
			{
				List *body_stmt_list = castNode(List, linitial((List *) create_function_stmt->sql_body));
				foreach(lc, body_stmt_list)
				{
					if (IsA(lfirst(lc), ReturnStmt))
					{
						deparseReturnStmt(state, lfirst_node(ReturnStmt, lc));
						deparseAppendStringInfoString(state, "; ");
					}
					else
					{
						deparseStmt(state, lfirst(lc));
						deparseAppendStringInfoString(state, "; ");
					}
				}
			}
			
			deparseAppendStringInfoString(state, "END ");
		}
	}

	removeTrailingSpace(state);
}

// "func_arg", "func_arg_with_default" and other places in gram.y
static void deparseFunctionParameter(DeparseState *state, FunctionParameter *function_parameter)
{
	switch (function_parameter->mode)
	{
		case FUNC_PARAM_IN: /* input only */
			deparseAppendStringInfoString(state, "IN ");
			break;
		case FUNC_PARAM_OUT: /* output only */
			deparseAppendStringInfoString(state, "OUT ");
			break;
		case FUNC_PARAM_INOUT: /* both */
			deparseAppendStringInfoString(state, "INOUT ");
			break;
		case FUNC_PARAM_VARIADIC: /* variadic (always input) */
			deparseAppendStringInfoString(state, "VARIADIC ");
			break;
		case FUNC_PARAM_TABLE: /* table function output column */
			// No special annotation, the caller is expected to correctly put
			// this into the RETURNS part of the CREATE FUNCTION statement
			break;
		case FUNC_PARAM_DEFAULT:
			// Default
			break;
		default:
			Assert(false);
			break;
	}

	if (function_parameter->name != NULL)
	{
		deparseAppendStringInfoString(state, function_parameter->name);
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseTypeName(state, function_parameter->argType);
	deparseAppendStringInfoChar(state, ' ');

	if (function_parameter->defexpr != NULL)
	{
		deparseAppendStringInfoString(state, "= ");
		deparseExpr(state, function_parameter->defexpr, DEPARSE_NODE_CONTEXT_A_EXPR);
	}

	removeTrailingSpace(state);
}

static void deparseCheckPointStmt(DeparseState *state, CheckPointStmt *check_point_stmt)
{
	deparseAppendStringInfoString(state, "CHECKPOINT");
}

static void deparseCreateSchemaStmt(DeparseState *state, CreateSchemaStmt *create_schema_stmt)
{
	ListCell *lc;
	deparseAppendStringInfoString(state, "CREATE SCHEMA ");

	if (create_schema_stmt->if_not_exists)
		deparseAppendStringInfoString(state, "IF NOT EXISTS ");

	if (create_schema_stmt->schemaname)
	{
		deparseColId(state, create_schema_stmt->schemaname);
		deparseAppendStringInfoChar(state, ' ');
	}
		
	if (create_schema_stmt->authrole != NULL)
	{
		deparseAppendStringInfoString(state, "AUTHORIZATION ");
		deparseRoleSpec(state, create_schema_stmt->authrole);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (create_schema_stmt->schemaElts)
	{
		foreach(lc, create_schema_stmt->schemaElts)
		{
			deparseSchemaStmt(state, lfirst(lc));
			if (lnext(create_schema_stmt->schemaElts, lc))
				deparseAppendStringInfoChar(state, ' ');
		}
	}

	removeTrailingSpace(state);
}

static void deparseAlterRoleSetStmt(DeparseState *state, AlterRoleSetStmt *alter_role_set_stmt)
{
	deparseAppendStringInfoString(state, "ALTER ROLE ");

	if (alter_role_set_stmt->role == NULL)
		deparseAppendStringInfoString(state, "ALL");
	else
		deparseRoleSpec(state, alter_role_set_stmt->role);

	deparseAppendStringInfoChar(state, ' ');
	
	if (alter_role_set_stmt->database != NULL)
	{
		deparseAppendStringInfoString(state, "IN DATABASE ");
		deparseAppendStringInfoString(state, quote_identifier(alter_role_set_stmt->database));
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseVariableSetStmt(state, alter_role_set_stmt->setstmt);
}

static void deparseCreateConversionStmt(DeparseState *state, CreateConversionStmt *create_conversion_stmt)
{
	deparseAppendStringInfoString(state, "CREATE ");
	if (create_conversion_stmt->def)
		deparseAppendStringInfoString(state, "DEFAULT ");

	deparseAppendStringInfoString(state, "CONVERSION ");
	deparseAnyName(state, create_conversion_stmt->conversion_name);
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "FOR ");
	deparseStringLiteral(state, create_conversion_stmt->for_encoding_name);
	deparseAppendStringInfoString(state, " TO ");
	deparseStringLiteral(state, create_conversion_stmt->to_encoding_name);

	deparseAppendStringInfoString(state, "FROM ");
	deparseAnyName(state, create_conversion_stmt->func_name);
}

static void deparseRoleSpec(DeparseState *state, RoleSpec *role_spec)
{
	if (role_spec == NULL)
		elog(ERROR, "deparse: unexpected NULL RoleSpec");

	switch (role_spec->roletype)
	{
		case ROLESPEC_CSTRING:
			Assert(role_spec->rolename != NULL);
			deparseAppendStringInfoString(state, quote_identifier(role_spec->rolename));
			break;
		case ROLESPEC_CURRENT_ROLE:
			deparseAppendStringInfoString(state, "CURRENT_ROLE");
			break;
		case ROLESPEC_CURRENT_USER:
			deparseAppendStringInfoString(state, "CURRENT_USER");
			break;
		case ROLESPEC_SESSION_USER:
			deparseAppendStringInfoString(state, "SESSION_USER");
			break;
		case ROLESPEC_PUBLIC:
			deparseAppendStringInfoString(state, "public");
			break;
	}
}

// "part_elem" in gram.y
static void deparsePartitionElem(DeparseState *state, PartitionElem *partition_elem)
{
	ListCell *lc;

	if (partition_elem->name != NULL)
	{
		deparseColId(state, partition_elem->name);
		deparseAppendStringInfoChar(state, ' ');
	}
	else if (partition_elem->expr != NULL)
	{
		deparseAppendStringInfoChar(state, '(');
		deparseExpr(state, partition_elem->expr, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, ") ");
	}

	deparseOptCollate(state, partition_elem->collation);
	deparseAnyName(state, partition_elem->opclass);

	removeTrailingSpace(state);
}

static void deparsePartitionSpec(DeparseState *state, PartitionSpec *partition_spec)
{
	ListCell *lc;

	deparseAppendPartGroup(state, "PARTITION BY", DEPARSE_PART_INDENT);

	switch (partition_spec->strategy)
	{
		case PARTITION_STRATEGY_LIST:
			deparseAppendStringInfoString(state, "LIST");
			break;
		case PARTITION_STRATEGY_HASH:
			deparseAppendStringInfoString(state, "HASH");
			break;
		case PARTITION_STRATEGY_RANGE:
			deparseAppendStringInfoString(state, "RANGE");
			break;
	}

	deparseAppendStringInfoString(state, " (");
	foreach(lc, partition_spec->partParams)
	{
		deparsePartitionElem(state, castNode(PartitionElem, lfirst(lc)));
		if (lnext(partition_spec->partParams, lc))
			deparseAppendStringInfoString(state, ", ");
	}
	deparseAppendStringInfoChar(state, ')');
}

static void deparsePartitionBoundSpec(DeparseState *state, PartitionBoundSpec *partition_bound_spec)
{
	ListCell *lc;

	if (partition_bound_spec->is_default)
	{
		deparseAppendStringInfoString(state, "DEFAULT");
		return;
	}

	deparseAppendStringInfoString(state, "FOR VALUES ");

	switch (partition_bound_spec->strategy)
	{
		case PARTITION_STRATEGY_HASH:
			deparseAppendStringInfo(state, "WITH (MODULUS %d, REMAINDER %d)", partition_bound_spec->modulus, partition_bound_spec->remainder);
			break;
		case PARTITION_STRATEGY_LIST:
			deparseAppendStringInfoString(state, "IN (");
			deparseExprList(state, partition_bound_spec->listdatums);
			deparseAppendStringInfoChar(state, ')');
			break;
		case PARTITION_STRATEGY_RANGE:
			deparseAppendStringInfoString(state, "FROM (");
			deparseExprList(state, partition_bound_spec->lowerdatums);
			deparseAppendStringInfoString(state, ") TO (");
			deparseExprList(state, partition_bound_spec->upperdatums);
			deparseAppendStringInfoChar(state, ')');
			break;
		default:
			Assert(false);
			break;
	}
}

static void deparsePartitionCmd(DeparseState *state, PartitionCmd *partition_cmd)
{
	if (partition_cmd == NULL)
		elog(ERROR, "deparse: unexpected NULL PartitionCmd");

	deparseRangeVar(state, partition_cmd->name, DEPARSE_NODE_CONTEXT_NONE);

	if (partition_cmd->bound != NULL)
	{
		deparseAppendStringInfoChar(state, ' ');
		deparsePartitionBoundSpec(state, partition_cmd->bound);
	}
	if (partition_cmd->concurrent)
		deparseAppendStringInfoString(state, " CONCURRENTLY ");
}

// "TableElement" in gram.y
static void deparseTableElement(DeparseState *state, Node *node)
{
	if (node == NULL)
		elog(ERROR, "deparse: unexpected NULL node");

	switch (nodeTag(node))
	{
		case T_ColumnDef:
			deparseColumnDef(state, castNode(ColumnDef, node));
			break;
		case T_TableLikeClause:
			deparseTableLikeClause(state, castNode(TableLikeClause, node));
			break;
		case T_Constraint:
			deparseConstraint(state, castNode(Constraint, node), DEPARSE_NODE_CONTEXT_NONE);
			break;
		default:
			Assert(false);
	}
}

static void deparseCreateStmt(DeparseState *state, CreateStmt *create_stmt, bool is_foreign_table)
{
	if (create_stmt->relation == NULL)
		elog(ERROR, "deparse: unexpected NULL relation in CreateStmt");

	ListCell *lc;

	deparseAppendStringInfoString(state, "CREATE ");

	if (is_foreign_table)
		deparseAppendStringInfoString(state, "FOREIGN ");

	deparseOptTemp(state, create_stmt->relation->relpersistence);

	deparseAppendStringInfoString(state, "TABLE ");

	if (create_stmt->if_not_exists)
		deparseAppendStringInfoString(state, "IF NOT EXISTS ");

	deparseRangeVar(state, create_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ' ');

	if (create_stmt->ofTypename != NULL)
	{
		deparseAppendStringInfoString(state, "OF ");
		deparseTypeName(state, create_stmt->ofTypename);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (create_stmt->partbound != NULL)
	{
		Assert(list_length(create_stmt->inhRelations) == 1);
		deparseAppendStringInfoString(state, "PARTITION OF ");
		deparseRangeVar(state, castNode(RangeVar, linitial(create_stmt->inhRelations)), DEPARSE_NODE_CONTEXT_NONE);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (list_length(create_stmt->tableElts) > 0)
	{
		DeparseStateNestingLevel *parent_level = NULL;
		// In raw parse output tableElts contains both columns and constraints
		// (and the constraints field is NIL)
		deparseAppendStringInfoChar(state, '(');
		parent_level = deparseStateIncreaseNestingLevel(state);
		foreach(lc, create_stmt->tableElts)
		{
			deparseTableElement(state, lfirst(lc));
			if (lnext(create_stmt->tableElts, lc))
				deparseAppendCommaAndPart(state);
		}
		deparseStateDecreaseNestingLevel(state, parent_level);
		deparseAppendStringInfoString(state, ") ");
	}
	else if (create_stmt->partbound == NULL && create_stmt->ofTypename == NULL)
	{
		deparseAppendStringInfoString(state, "() ");
	}

	if (create_stmt->partbound != NULL)
	{
		deparsePartitionBoundSpec(state, create_stmt->partbound);
		deparseAppendStringInfoChar(state, ' ');
	}
	else
	{
		deparseOptInherit(state, create_stmt->inhRelations);
	}

	if (create_stmt->partspec != NULL)
	{
		deparsePartitionSpec(state, create_stmt->partspec);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (create_stmt->accessMethod != NULL)
	{
		deparseAppendStringInfoString(state, "USING ");
		deparseAppendStringInfoString(state, quote_identifier(create_stmt->accessMethod));
	}

	deparseOptWith(state, create_stmt->options);

	switch (create_stmt->oncommit)
	{
		case ONCOMMIT_NOOP:
			// No ON COMMIT clause
			break;
		case ONCOMMIT_PRESERVE_ROWS:
			deparseAppendStringInfoString(state, "ON COMMIT PRESERVE ROWS ");
			break;
		case ONCOMMIT_DELETE_ROWS:
			deparseAppendStringInfoString(state, "ON COMMIT DELETE ROWS ");
			break;
		case ONCOMMIT_DROP:
			deparseAppendStringInfoString(state, "ON COMMIT DROP ");
			break;
	}

	if (create_stmt->tablespacename != NULL)
	{
		deparseAppendStringInfoString(state, "TABLESPACE ");
		deparseAppendStringInfoString(state, quote_identifier(create_stmt->tablespacename));
	}

	removeTrailingSpace(state);
}

static void deparseCreateFdwStmt(DeparseState *state, CreateFdwStmt *create_fdw_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "CREATE FOREIGN DATA WRAPPER ");
	deparseAppendStringInfoString(state, quote_identifier(create_fdw_stmt->fdwname));
	deparseAppendStringInfoChar(state, ' ');

	if (list_length(create_fdw_stmt->func_options) > 0)
	{
		deparseFdwOptions(state, create_fdw_stmt->func_options);
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseCreateGenericOptions(state, create_fdw_stmt->options);

	removeTrailingSpace(state);
}

static void deparseAlterFdwStmt(DeparseState *state, AlterFdwStmt *alter_fdw_stmt)
{
	deparseAppendStringInfoString(state, "ALTER FOREIGN DATA WRAPPER ");
	deparseAppendStringInfoString(state, quote_identifier(alter_fdw_stmt->fdwname));
	deparseAppendStringInfoChar(state, ' ');

	if (list_length(alter_fdw_stmt->func_options) > 0)
	{
		deparseFdwOptions(state, alter_fdw_stmt->func_options);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (list_length(alter_fdw_stmt->options) > 0)
		deparseAlterGenericOptions(state, alter_fdw_stmt->options);

	removeTrailingSpace(state);
}

static void deparseCreateForeignServerStmt(DeparseState *state, CreateForeignServerStmt *create_foreign_server_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "CREATE SERVER ");
	if (create_foreign_server_stmt->if_not_exists)
		deparseAppendStringInfoString(state, "IF NOT EXISTS ");
	deparseAppendStringInfoString(state, quote_identifier(create_foreign_server_stmt->servername));
	deparseAppendStringInfoChar(state, ' ');

	if (create_foreign_server_stmt->servertype != NULL)
	{
		deparseAppendStringInfoString(state, "TYPE ");
		deparseStringLiteral(state, create_foreign_server_stmt->servertype);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (create_foreign_server_stmt->version != NULL)
	{
		deparseAppendStringInfoString(state, "VERSION ");
		deparseStringLiteral(state, create_foreign_server_stmt->version);
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseAppendStringInfoString(state, "FOREIGN DATA WRAPPER ");
	deparseAppendStringInfoString(state, quote_identifier(create_foreign_server_stmt->fdwname));
	deparseAppendStringInfoChar(state, ' ');

	deparseCreateGenericOptions(state, create_foreign_server_stmt->options);

	removeTrailingSpace(state);
}

static void deparseAlterForeignServerStmt(DeparseState *state, AlterForeignServerStmt *alter_foreign_server_stmt)
{
	deparseAppendStringInfoString(state, "ALTER SERVER ");

	deparseAppendStringInfoString(state, quote_identifier(alter_foreign_server_stmt->servername));
	deparseAppendStringInfoChar(state, ' ');

	if (alter_foreign_server_stmt->has_version)
	{
		deparseAppendStringInfoString(state, "VERSION ");
		if (alter_foreign_server_stmt->version != NULL)
			deparseStringLiteral(state, alter_foreign_server_stmt->version);
		else
			deparseAppendStringInfoString(state, "NULL");
		deparseAppendStringInfoChar(state, ' ');
	}

	if (list_length(alter_foreign_server_stmt->options) > 0)
		deparseAlterGenericOptions(state, alter_foreign_server_stmt->options);

	removeTrailingSpace(state);
}

static void deparseCreateUserMappingStmt(DeparseState *state, CreateUserMappingStmt *create_user_mapping_stmt)
{
	deparseAppendStringInfoString(state, "CREATE USER MAPPING ");
	if (create_user_mapping_stmt->if_not_exists)
		deparseAppendStringInfoString(state, "IF NOT EXISTS ");

	deparseAppendStringInfoString(state, "FOR ");
	deparseRoleSpec(state, create_user_mapping_stmt->user);
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "SERVER ");
	deparseAppendStringInfoString(state, quote_identifier(create_user_mapping_stmt->servername));
	deparseAppendStringInfoChar(state, ' ');

	deparseCreateGenericOptions(state, create_user_mapping_stmt->options);

	removeTrailingSpace(state);
}

static void deparseCreatedbStmt(DeparseState *state, CreatedbStmt *createdb_stmt)
{
	deparseAppendStringInfoString(state, "CREATE DATABASE ");
	deparseColId(state, createdb_stmt->dbname);
	deparseAppendStringInfoChar(state, ' ');
	deparseCreatedbOptList(state, createdb_stmt->options);
	removeTrailingSpace(state);
}

static void deparseAlterUserMappingStmt(DeparseState *state, AlterUserMappingStmt *alter_user_mapping_stmt)
{
	deparseAppendStringInfoString(state, "ALTER USER MAPPING FOR ");
	deparseRoleSpec(state, alter_user_mapping_stmt->user);
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "SERVER ");
	deparseAppendStringInfoString(state, quote_identifier(alter_user_mapping_stmt->servername));
	deparseAppendStringInfoChar(state, ' ');

	deparseAlterGenericOptions(state, alter_user_mapping_stmt->options);

	removeTrailingSpace(state);
}

static void deparseDropUserMappingStmt(DeparseState *state, DropUserMappingStmt *drop_user_mapping_stmt)
{
	deparseAppendStringInfoString(state, "DROP USER MAPPING ");

	if (drop_user_mapping_stmt->missing_ok)
		deparseAppendStringInfoString(state, "IF EXISTS ");

	deparseAppendStringInfoString(state, "FOR ");
	deparseRoleSpec(state, drop_user_mapping_stmt->user);
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "SERVER ");
	deparseAppendStringInfoString(state, quote_identifier(drop_user_mapping_stmt->servername));
}

static void deparseSecLabelStmt(DeparseState *state, SecLabelStmt *sec_label_stmt)
{
	ListCell *lc = NULL;

	deparseAppendStringInfoString(state, "SECURITY LABEL ");

	if (sec_label_stmt->provider != NULL)
	{
		deparseAppendStringInfoString(state, "FOR ");
		deparseAppendStringInfoString(state, quote_identifier(sec_label_stmt->provider));
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseAppendStringInfoString(state, "ON ");

	switch (sec_label_stmt->objtype)
	{
		case OBJECT_COLUMN:
			deparseAppendStringInfoString(state, "COLUMN ");
			deparseAnyName(state, castNode(List, sec_label_stmt->object));
			break;
		case OBJECT_FOREIGN_TABLE:
			deparseAppendStringInfoString(state, "FOREIGN TABLE ");
			deparseAnyName(state, castNode(List, sec_label_stmt->object));
			break;
		case OBJECT_SEQUENCE:
			deparseAppendStringInfoString(state, "SEQUENCE ");
			deparseAnyName(state, castNode(List, sec_label_stmt->object));
			break;
		case OBJECT_TABLE:
			deparseAppendStringInfoString(state, "TABLE ");
			deparseAnyName(state, castNode(List, sec_label_stmt->object));
			break;
		case OBJECT_VIEW:
			deparseAppendStringInfoString(state, "VIEW ");
			deparseAnyName(state, castNode(List, sec_label_stmt->object));
			break;
		case OBJECT_MATVIEW:
			deparseAppendStringInfoString(state, "MATERIALIZED VIEW ");
			deparseAnyName(state, castNode(List, sec_label_stmt->object));
			break;
		case OBJECT_DATABASE:
			deparseAppendStringInfoString(state, "DATABASE ");
			deparseAppendStringInfoString(state, quote_identifier(strVal(sec_label_stmt->object)));
			break;
		case OBJECT_EVENT_TRIGGER:
			deparseAppendStringInfoString(state, "EVENT TRIGGER ");
			deparseAppendStringInfoString(state, quote_identifier(strVal(sec_label_stmt->object)));
			break;
		case OBJECT_LANGUAGE:
			deparseAppendStringInfoString(state, "LANGUAGE ");
			deparseAppendStringInfoString(state, quote_identifier(strVal(sec_label_stmt->object)));
			break;
		case OBJECT_PUBLICATION:
			deparseAppendStringInfoString(state, "PUBLICATION ");
			deparseAppendStringInfoString(state, quote_identifier(strVal(sec_label_stmt->object)));
			break;
		case OBJECT_ROLE:
			deparseAppendStringInfoString(state, "ROLE ");
			deparseAppendStringInfoString(state, quote_identifier(strVal(sec_label_stmt->object)));
			break;
		case OBJECT_SCHEMA:
			deparseAppendStringInfoString(state, "SCHEMA ");
			deparseAppendStringInfoString(state, quote_identifier(strVal(sec_label_stmt->object)));
			break;
		case OBJECT_SUBSCRIPTION:
			deparseAppendStringInfoString(state, "SUBSCRIPTION ");
			deparseAppendStringInfoString(state, quote_identifier(strVal(sec_label_stmt->object)));
			break;
		case OBJECT_TABLESPACE:
			deparseAppendStringInfoString(state, "TABLESPACE ");
			deparseAppendStringInfoString(state, quote_identifier(strVal(sec_label_stmt->object)));
			break;
		case OBJECT_TYPE:
			deparseAppendStringInfoString(state, "TYPE ");
			deparseTypeName(state, castNode(TypeName, sec_label_stmt->object));
			break;
		case OBJECT_DOMAIN:
			deparseAppendStringInfoString(state, "DOMAIN ");
			deparseTypeName(state, castNode(TypeName, sec_label_stmt->object));
			break;
		case OBJECT_AGGREGATE:
			deparseAppendStringInfoString(state, "AGGREGATE ");
			deparseAggregateWithArgtypes(state, castNode(ObjectWithArgs, sec_label_stmt->object));
			break;
		case OBJECT_FUNCTION:
			deparseAppendStringInfoString(state, "FUNCTION ");
			deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, sec_label_stmt->object));
			break;
		case OBJECT_LARGEOBJECT:
			deparseAppendStringInfoString(state, "LARGE OBJECT ");
			deparseValue(state, (union ValUnion *) sec_label_stmt->object, DEPARSE_NODE_CONTEXT_CONSTANT);
			break;
		case OBJECT_PROCEDURE:
			deparseAppendStringInfoString(state, "PROCEDURE ");
			deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, sec_label_stmt->object));
			break;
		case OBJECT_ROUTINE:
			deparseAppendStringInfoString(state, "ROUTINE ");
			deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, sec_label_stmt->object));
			break;
		default:
			// Not supported in the parser
			Assert(false);
			break;
	}

	deparseAppendStringInfoString(state, " IS ");

	if (sec_label_stmt->label != NULL)
		deparseStringLiteral(state, sec_label_stmt->label);
	else
		deparseAppendStringInfoString(state, "NULL");
}

static void deparseCreateForeignTableStmt(DeparseState *state, CreateForeignTableStmt *create_foreign_table_stmt)
{
	ListCell *lc;

	deparseCreateStmt(state, &create_foreign_table_stmt->base, true);

	deparseAppendStringInfoString(state, " SERVER ");
	deparseAppendStringInfoString(state, quote_identifier(create_foreign_table_stmt->servername));
	deparseAppendStringInfoChar(state, ' ');

	if (list_length(create_foreign_table_stmt->options) > 0)
		deparseAlterGenericOptions(state, create_foreign_table_stmt->options);

	removeTrailingSpace(state);
}

static void deparseImportForeignSchemaStmt(DeparseState *state, ImportForeignSchemaStmt *import_foreign_schema_stmt)
{
	deparseAppendStringInfoString(state, "IMPORT FOREIGN SCHEMA ");

	deparseAppendStringInfoString(state, import_foreign_schema_stmt->remote_schema);
	deparseAppendStringInfoChar(state, ' ');

	switch (import_foreign_schema_stmt->list_type)
	{
		case FDW_IMPORT_SCHEMA_ALL:
			// Default
			break;
		case FDW_IMPORT_SCHEMA_LIMIT_TO:
			deparseAppendStringInfoString(state, "LIMIT TO (");
			deparseRelationExprList(state, import_foreign_schema_stmt->table_list);
			deparseAppendStringInfoString(state, ") ");
			break;
		case FDW_IMPORT_SCHEMA_EXCEPT:
			deparseAppendStringInfoString(state, "EXCEPT (");
			deparseRelationExprList(state, import_foreign_schema_stmt->table_list);
			deparseAppendStringInfoString(state, ") ");
			break;
	}

	deparseAppendStringInfoString(state, "FROM SERVER ");
	deparseAppendStringInfoString(state, quote_identifier(import_foreign_schema_stmt->server_name));
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "INTO ");
	deparseAppendStringInfoString(state, quote_identifier(import_foreign_schema_stmt->local_schema));
	deparseAppendStringInfoChar(state, ' ');

	deparseCreateGenericOptions(state, import_foreign_schema_stmt->options);

	removeTrailingSpace(state);
}

static void deparseCreateTableAsStmt(DeparseState *state, CreateTableAsStmt *create_table_as_stmt)
{
	if (create_table_as_stmt->into == NULL || create_table_as_stmt->into->rel == NULL ||
		create_table_as_stmt->query == NULL)
		elog(ERROR, "deparse: unexpected NULL into/query in CreateTableAsStmt");

	ListCell *lc;
	deparseAppendStringInfoString(state, "CREATE ");

	deparseOptTemp(state, create_table_as_stmt->into->rel->relpersistence);

	switch (create_table_as_stmt->objtype)
	{
		case OBJECT_TABLE:
			deparseAppendStringInfoString(state, "TABLE ");
			break;
		case OBJECT_MATVIEW:
			deparseAppendStringInfoString(state, "MATERIALIZED VIEW ");
			break;
		default:
			// Not supported here
			Assert(false);
			break;
	}

	if (create_table_as_stmt->if_not_exists)
		deparseAppendStringInfoString(state, "IF NOT EXISTS ");

	deparseIntoClause(state, create_table_as_stmt->into);
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "AS ");
	if (IsA(create_table_as_stmt->query, ExecuteStmt))
		deparseExecuteStmt(state, castNode(ExecuteStmt, create_table_as_stmt->query));
	else
		deparseSelectStmt(state, castNode(SelectStmt, create_table_as_stmt->query), DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ' ');

	if (create_table_as_stmt->into->skipData)
		deparseAppendStringInfoString(state, "WITH NO DATA ");

	removeTrailingSpace(state);
}

static void deparseViewStmt(DeparseState *state, ViewStmt *view_stmt)
{
	if (view_stmt->view == NULL)
		elog(ERROR, "deparse: unexpected NULL view in ViewStmt");

	ListCell *lc;

	deparseAppendStringInfoString(state, "CREATE ");

	if (view_stmt->replace)
		deparseAppendStringInfoString(state, "OR REPLACE ");

	deparseOptTemp(state, view_stmt->view->relpersistence);

	deparseAppendStringInfoString(state, "VIEW ");
	deparseRangeVar(state, view_stmt->view, DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ' ');

	if (list_length(view_stmt->aliases) > 0)
	{
		deparseAppendStringInfoChar(state, '(');
		deparseColumnList(state, view_stmt->aliases);
		deparseAppendStringInfoString(state, ") ");
	}

	deparseOptWith(state, view_stmt->options);

	deparseAppendStringInfoString(state, "AS ");
	deparseSelectStmt(state, castNode(SelectStmt, view_stmt->query), DEPARSE_NODE_CONTEXT_NONE);

	switch (view_stmt->withCheckOption)
	{
		case NO_CHECK_OPTION:
			// Default
			break;
		case LOCAL_CHECK_OPTION:
			deparseAppendStringInfoString(state, "WITH LOCAL CHECK OPTION ");
			break;
		case CASCADED_CHECK_OPTION:
			deparseAppendStringInfoString(state, "WITH CHECK OPTION ");
			break;
	}

	removeTrailingSpace(state);
}

static void deparseDropStmt(DeparseState *state, DropStmt *drop_stmt)
{
	ListCell *lc;
	List *l;

	deparseAppendStringInfoString(state, "DROP ");

	switch (drop_stmt->removeType)
	{
		case OBJECT_ACCESS_METHOD:
			deparseAppendStringInfoString(state, "ACCESS METHOD ");
			break;
		case OBJECT_AGGREGATE:
			deparseAppendStringInfoString(state, "AGGREGATE ");
			break;
		case OBJECT_CAST:
			deparseAppendStringInfoString(state, "CAST ");
			break;
		case OBJECT_COLLATION:
			deparseAppendStringInfoString(state, "COLLATION ");
			break;
		case OBJECT_CONVERSION:
			deparseAppendStringInfoString(state, "CONVERSION ");
			break;
		case OBJECT_DOMAIN:
			deparseAppendStringInfoString(state, "DOMAIN ");
			break;
		case OBJECT_EVENT_TRIGGER:
			deparseAppendStringInfoString(state, "EVENT TRIGGER ");
			break;
		case OBJECT_EXTENSION:
			deparseAppendStringInfoString(state, "EXTENSION ");
			break;
		case OBJECT_FDW:
			deparseAppendStringInfoString(state, "FOREIGN DATA WRAPPER ");
			break;
		case OBJECT_FOREIGN_SERVER:
			deparseAppendStringInfoString(state, "SERVER ");
			break;
		case OBJECT_FOREIGN_TABLE:
			deparseAppendStringInfoString(state, "FOREIGN TABLE ");
			break;
		case OBJECT_FUNCTION:
			deparseAppendStringInfoString(state, "FUNCTION ");
			break;
		case OBJECT_INDEX:
			deparseAppendStringInfoString(state, "INDEX ");
			break;
		case OBJECT_LANGUAGE:
			deparseAppendStringInfoString(state, "LANGUAGE ");
			break;
		case OBJECT_MATVIEW:
			deparseAppendStringInfoString(state, "MATERIALIZED VIEW ");
			break;
		case OBJECT_OPCLASS:
			deparseAppendStringInfoString(state, "OPERATOR CLASS ");
			break;
		case OBJECT_OPERATOR:
			deparseAppendStringInfoString(state, "OPERATOR ");
			break;
		case OBJECT_OPFAMILY:
			deparseAppendStringInfoString(state, "OPERATOR FAMILY ");
			break;
		case OBJECT_POLICY:
			deparseAppendStringInfoString(state, "POLICY ");
			break;
		case OBJECT_PROCEDURE:
			deparseAppendStringInfoString(state, "PROCEDURE ");
			break;
		case OBJECT_PUBLICATION:
			deparseAppendStringInfoString(state, "PUBLICATION ");
			break;
		case OBJECT_ROUTINE:
			deparseAppendStringInfoString(state, "ROUTINE ");
			break;
		case OBJECT_RULE:
			deparseAppendStringInfoString(state, "RULE ");
			break;
		case OBJECT_SCHEMA:
			deparseAppendStringInfoString(state, "SCHEMA ");
			break;
		case OBJECT_SEQUENCE:
			deparseAppendStringInfoString(state, "SEQUENCE ");
			break;
		case OBJECT_STATISTIC_EXT:
			deparseAppendStringInfoString(state, "STATISTICS ");
			break;
		case OBJECT_TABLE:
			deparseAppendStringInfoString(state, "TABLE ");
			break;
		case OBJECT_TRANSFORM:
			deparseAppendStringInfoString(state, "TRANSFORM ");
			break;
		case OBJECT_TRIGGER:
			deparseAppendStringInfoString(state, "TRIGGER ");
			break;
		case OBJECT_TSCONFIGURATION:
			deparseAppendStringInfoString(state, "TEXT SEARCH CONFIGURATION ");
			break;
		case OBJECT_TSDICTIONARY:
			deparseAppendStringInfoString(state, "TEXT SEARCH DICTIONARY ");
			break;
		case OBJECT_TSPARSER:
			deparseAppendStringInfoString(state, "TEXT SEARCH PARSER ");
			break;
		case OBJECT_TSTEMPLATE:
			deparseAppendStringInfoString(state, "TEXT SEARCH TEMPLATE ");
			break;
		case OBJECT_TYPE:
			deparseAppendStringInfoString(state, "TYPE ");
			break;
		case OBJECT_VIEW:
			deparseAppendStringInfoString(state, "VIEW ");
			break;
		default:
			// Other object types are not supported here in the parser
			Assert(false);
	}

	if (drop_stmt->concurrent)
		deparseAppendStringInfoString(state, "CONCURRENTLY ");

	if (drop_stmt->missing_ok)
		deparseAppendStringInfoString(state, "IF EXISTS ");

	switch (drop_stmt->removeType)
	{
		// drop_type_any_name
		case OBJECT_TABLE:
		case OBJECT_SEQUENCE:
		case OBJECT_VIEW:
		case OBJECT_MATVIEW:
		case OBJECT_INDEX:
		case OBJECT_FOREIGN_TABLE:
		case OBJECT_COLLATION:
		case OBJECT_CONVERSION:
		case OBJECT_STATISTIC_EXT:
		case OBJECT_TSPARSER:
		case OBJECT_TSDICTIONARY:
		case OBJECT_TSTEMPLATE:
		case OBJECT_TSCONFIGURATION:
			deparseAnyNameList(state, drop_stmt->objects);
			deparseAppendStringInfoChar(state, ' ');
			break;
		// drop_type_name
		case OBJECT_ACCESS_METHOD:
		case OBJECT_EVENT_TRIGGER:
		case OBJECT_EXTENSION:
		case OBJECT_FDW:
		case OBJECT_PUBLICATION:
		case OBJECT_SCHEMA:
		case OBJECT_FOREIGN_SERVER:
			deparseNameList(state, drop_stmt->objects);
			deparseAppendStringInfoChar(state, ' ');
			break;
		// drop_type_name_on_any_name
		case OBJECT_POLICY:
		case OBJECT_RULE:
		case OBJECT_TRIGGER:
			Assert(list_length(drop_stmt->objects) == 1);
			l = linitial(drop_stmt->objects);
			deparseColId(state, strVal(llast(l)));
			deparseAppendStringInfoString(state, " ON ");
			deparseAnyNameSkipLast(state, l);
			deparseAppendStringInfoChar(state, ' ');
			break;
		case OBJECT_CAST:
			Assert(list_length(drop_stmt->objects) == 1);
			l = linitial(drop_stmt->objects);
			Assert(list_length(l) == 2);
			deparseAppendStringInfoChar(state, '(');
			deparseTypeName(state, castNode(TypeName, linitial(l)));
			deparseAppendStringInfoString(state, " AS ");
			deparseTypeName(state, castNode(TypeName, lsecond(l)));
			deparseAppendStringInfoChar(state, ')');
			deparseAppendStringInfoChar(state, ' ');
			break;
		case OBJECT_OPFAMILY:
		case OBJECT_OPCLASS:
			Assert(list_length(drop_stmt->objects) == 1);
			l = linitial(drop_stmt->objects);
			deparseAnyNameSkipFirst(state, l);
			deparseAppendStringInfoString(state, " USING ");
			deparseColId(state, strVal(linitial(l)));
			deparseAppendStringInfoChar(state, ' ');
			break;
		case OBJECT_TRANSFORM:
			Assert(list_length(drop_stmt->objects) == 1);
			l = linitial(drop_stmt->objects);
			deparseAppendStringInfoString(state, "FOR ");
			deparseTypeName(state, castNode(TypeName, linitial(l)));
			deparseAppendStringInfoString(state, " LANGUAGE ");
			deparseColId(state, strVal(lsecond(l)));
			deparseAppendStringInfoChar(state, ' ');
			break;
		case OBJECT_LANGUAGE:
			deparseNameList(state, drop_stmt->objects);
			deparseAppendStringInfoChar(state, ' ');
			break;
		case OBJECT_TYPE:
		case OBJECT_DOMAIN:
			foreach(lc, drop_stmt->objects)
			{
				deparseTypeName(state, castNode(TypeName, lfirst(lc)));
				if (lnext(drop_stmt->objects, lc))
					deparseAppendStringInfoString(state, ", ");
			}
			deparseAppendStringInfoChar(state, ' ');
			break;
		case OBJECT_AGGREGATE:
			foreach(lc, drop_stmt->objects)
			{
				deparseAggregateWithArgtypes(state, castNode(ObjectWithArgs, lfirst(lc)));
				if (lnext(drop_stmt->objects, lc))
					deparseAppendStringInfoString(state, ", ");
			}
			deparseAppendStringInfoChar(state, ' ');
			break;
		case OBJECT_FUNCTION:
		case OBJECT_PROCEDURE:
		case OBJECT_ROUTINE:
			foreach(lc, drop_stmt->objects)
			{
				deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, lfirst(lc)));
				if (lnext(drop_stmt->objects, lc))
					deparseAppendStringInfoString(state, ", ");
			}
			deparseAppendStringInfoChar(state, ' ');
			break;
		case OBJECT_OPERATOR:
			foreach(lc, drop_stmt->objects)
			{
				deparseOperatorWithArgtypes(state, castNode(ObjectWithArgs, lfirst(lc)));
				if (lnext(drop_stmt->objects, lc))
					deparseAppendStringInfoString(state, ", ");
			}
			deparseAppendStringInfoChar(state, ' ');
			break;
		default:
			Assert(false);
	}

	deparseOptDropBehavior(state, drop_stmt->behavior);

	removeTrailingSpace(state);
}

static void deparseGroupingSet(DeparseState *state, GroupingSet *grouping_set)
{
	switch(grouping_set->kind)
	{
		case GROUPING_SET_EMPTY:
			deparseAppendStringInfoString(state, "()");
			break;
		case GROUPING_SET_SIMPLE:
			// Not present in raw parse trees
			Assert(false);
			break;
		case GROUPING_SET_ROLLUP:
			deparseAppendStringInfoString(state, "ROLLUP (");
			deparseExprList(state, grouping_set->content);
			deparseAppendStringInfoChar(state, ')');
			break;
		case GROUPING_SET_CUBE:
			deparseAppendStringInfoString(state, "CUBE (");
			deparseExprList(state, grouping_set->content);
			deparseAppendStringInfoChar(state, ')');
			break;
		case GROUPING_SET_SETS:
			deparseAppendStringInfoString(state, "GROUPING SETS (");
			deparseGroupByList(state, grouping_set->content);
			deparseAppendStringInfoChar(state, ')');
			break;
	}
}

static void deparseDropTableSpaceStmt(DeparseState *state, DropTableSpaceStmt *drop_table_space_stmt)
{
	deparseAppendStringInfoString(state, "DROP TABLESPACE ");

	if (drop_table_space_stmt->missing_ok)
		deparseAppendStringInfoString(state, "IF EXISTS ");

	deparseAppendStringInfoString(state, drop_table_space_stmt->tablespacename);
}

static void deparseAlterObjectDependsStmt(DeparseState *state, AlterObjectDependsStmt *alter_object_depends_stmt)
{
	deparseAppendStringInfoString(state, "ALTER ");

	switch (alter_object_depends_stmt->objectType)
	{
		case OBJECT_FUNCTION:
			deparseAppendStringInfoString(state, "FUNCTION ");
			deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, alter_object_depends_stmt->object));
			break;
		case OBJECT_PROCEDURE:
			deparseAppendStringInfoString(state, "PROCEDURE ");
			deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, alter_object_depends_stmt->object));
			break;
		case OBJECT_ROUTINE:
			deparseAppendStringInfoString(state, "ROUTINE ");
			deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, alter_object_depends_stmt->object));
			break;
		case OBJECT_TRIGGER:
			deparseAppendStringInfoString(state, "TRIGGER ");
			deparseColId(state, strVal(linitial(castNode(List, alter_object_depends_stmt->object))));
			deparseAppendStringInfoString(state, " ON ");
			deparseRangeVar(state, alter_object_depends_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
			break;
		case OBJECT_MATVIEW:
			deparseAppendStringInfoString(state, "MATERIALIZED VIEW ");
			deparseRangeVar(state, alter_object_depends_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
			break;
		case OBJECT_INDEX:
			deparseAppendStringInfoString(state, "INDEX ");
			deparseRangeVar(state, alter_object_depends_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
			break;
		default:
			// No other object types supported here
			Assert(false);
	}
	deparseAppendStringInfoChar(state, ' ');

	if (alter_object_depends_stmt->remove)
		deparseAppendStringInfoString(state, "NO ");

	deparseAppendStringInfo(state, "DEPENDS ON EXTENSION %s", alter_object_depends_stmt->extname->sval);
}

static void deparseAlterObjectSchemaStmt(DeparseState *state, AlterObjectSchemaStmt *alter_object_schema_stmt)
{
	List *l = NULL;
	ListCell *lc = NULL;

	deparseAppendStringInfoString(state, "ALTER ");

	switch (alter_object_schema_stmt->objectType)
	{
		case OBJECT_AGGREGATE:
			deparseAppendStringInfoString(state, "AGGREGATE ");
			deparseAggregateWithArgtypes(state, castNode(ObjectWithArgs, alter_object_schema_stmt->object));
			break;
		case OBJECT_COLLATION:
			deparseAppendStringInfoString(state, "COLLATION ");
			deparseAnyName(state, castNode(List, alter_object_schema_stmt->object));
			break;
		case OBJECT_CONVERSION:
			deparseAppendStringInfoString(state, "CONVERSION ");
			deparseAnyName(state, castNode(List, alter_object_schema_stmt->object));
			break;
		case OBJECT_DOMAIN:
			deparseAppendStringInfoString(state, "DOMAIN ");
			deparseAnyName(state, castNode(List, alter_object_schema_stmt->object));
			break;
		case OBJECT_EXTENSION:
			deparseAppendStringInfoString(state, "EXTENSION ");
			deparseAppendStringInfoString(state, quote_identifier(strVal(alter_object_schema_stmt->object)));
			break;
		case OBJECT_FUNCTION:
			deparseAppendStringInfoString(state, "FUNCTION ");
			deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, alter_object_schema_stmt->object));
			break;
		case OBJECT_OPERATOR:
			deparseAppendStringInfoString(state, "OPERATOR ");
			deparseOperatorWithArgtypes(state, castNode(ObjectWithArgs, alter_object_schema_stmt->object));
			break;
		case OBJECT_OPCLASS:
			l = castNode(List, alter_object_schema_stmt->object);
			deparseAppendStringInfoString(state, "OPERATOR CLASS ");
			deparseAnyNameSkipFirst(state, l);
			deparseAppendStringInfoString(state, " USING ");
			deparseAppendStringInfoString(state, quote_identifier(strVal(linitial(l))));
			break;
		case OBJECT_OPFAMILY:
			l = castNode(List, alter_object_schema_stmt->object);
			deparseAppendStringInfoString(state, "OPERATOR FAMILY ");
			deparseAnyNameSkipFirst(state, l);
			deparseAppendStringInfoString(state, " USING ");
			deparseAppendStringInfoString(state, quote_identifier(strVal(linitial(l))));
			break;
		case OBJECT_PROCEDURE:
			deparseAppendStringInfoString(state, "PROCEDURE ");
			deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, alter_object_schema_stmt->object));
			break;
		case OBJECT_ROUTINE:
			deparseAppendStringInfoString(state, "ROUTINE ");
			deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, alter_object_schema_stmt->object));
			break;
		case OBJECT_TABLE:
			deparseAppendStringInfoString(state, "TABLE ");
			if (alter_object_schema_stmt->missing_ok)
				deparseAppendStringInfoString(state, "IF EXISTS ");
			deparseRangeVar(state, alter_object_schema_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
			break;
		case OBJECT_STATISTIC_EXT:
			deparseAppendStringInfoString(state, "STATISTICS ");
			deparseAnyName(state, castNode(List, alter_object_schema_stmt->object));
			break;
		case OBJECT_TSPARSER:
			deparseAppendStringInfoString(state, "TEXT SEARCH PARSER ");
			deparseAnyName(state, castNode(List, alter_object_schema_stmt->object));
			break;
		case OBJECT_TSDICTIONARY:
			deparseAppendStringInfoString(state, "TEXT SEARCH DICTIONARY ");
			deparseAnyName(state, castNode(List, alter_object_schema_stmt->object));
			break;
		case OBJECT_TSTEMPLATE:
			deparseAppendStringInfoString(state, "TEXT SEARCH TEMPLATE ");
			deparseAnyName(state, castNode(List, alter_object_schema_stmt->object));
			break;
		case OBJECT_TSCONFIGURATION:
			deparseAppendStringInfoString(state, "TEXT SEARCH CONFIGURATION ");
			deparseAnyName(state, castNode(List, alter_object_schema_stmt->object));
			break;
		case OBJECT_SEQUENCE:
			deparseAppendStringInfoString(state, "SEQUENCE ");
			if (alter_object_schema_stmt->missing_ok)
				deparseAppendStringInfoString(state, "IF EXISTS ");
			deparseRangeVar(state, alter_object_schema_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
			break;
		case OBJECT_VIEW:
			deparseAppendStringInfoString(state, "VIEW ");
			if (alter_object_schema_stmt->missing_ok)
				deparseAppendStringInfoString(state, "IF EXISTS ");
			deparseRangeVar(state, alter_object_schema_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
			break;
		case OBJECT_MATVIEW:
			deparseAppendStringInfoString(state, "MATERIALIZED VIEW ");
			if (alter_object_schema_stmt->missing_ok)
				deparseAppendStringInfoString(state, "IF EXISTS ");
			deparseRangeVar(state, alter_object_schema_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
			break;
		case OBJECT_FOREIGN_TABLE:
			deparseAppendStringInfoString(state, "FOREIGN TABLE ");
			if (alter_object_schema_stmt->missing_ok)
				deparseAppendStringInfoString(state, "IF EXISTS ");
			deparseRangeVar(state, alter_object_schema_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
			break;
		case OBJECT_TYPE:
			deparseAppendStringInfoString(state, "TYPE ");
			deparseAnyName(state, castNode(List, alter_object_schema_stmt->object));
			break;
		default:
			Assert(false);
			break;
	}

	deparseAppendStringInfoString(state, " SET SCHEMA ");
	deparseAppendStringInfoString(state, quote_identifier(alter_object_schema_stmt->newschema));
}

// "alter_table_cmd" in gram.y
static void deparseAlterTableCmd(DeparseState *state, AlterTableCmd *alter_table_cmd, DeparseNodeContext context)
{
	ListCell *lc = NULL;
	const char *options = NULL;
	bool trailing_missing_ok = false;

	switch (alter_table_cmd->subtype)
	{
		case AT_AddColumn: /* add column */
			if (context == DEPARSE_NODE_CONTEXT_ALTER_TYPE)
				deparseAppendStringInfoString(state, "ADD ATTRIBUTE ");
			else
				deparseAppendStringInfoString(state, "ADD COLUMN ");
			break;
		case AT_AddColumnToView: /* implicitly via CREATE OR REPLACE VIEW */
			// Not present in raw parser output
			Assert(false);
			break;
		case AT_ColumnDefault: /* alter column default */
			deparseAppendStringInfoString(state, "ALTER COLUMN ");
			if (alter_table_cmd->def != NULL)
				options = "SET DEFAULT";
			else
				options = "DROP DEFAULT";
			break;
		case AT_CookedColumnDefault: /* add a pre-cooked column default */
			// Not present in raw parser output
			Assert(false);
			break;
		case AT_DropNotNull: /* alter column drop not null */
			deparseAppendStringInfoString(state, "ALTER COLUMN ");
			options = "DROP NOT NULL";
			break;
		case AT_SetNotNull: /* alter column set not null */
			deparseAppendStringInfoString(state, "ALTER COLUMN ");
			options = "SET NOT NULL";
			break;
		case AT_DropExpression: /* alter column drop expression */
			deparseAppendStringInfoString(state, "ALTER COLUMN ");
			options = "DROP EXPRESSION";
			trailing_missing_ok = true;
			break;
		case AT_SetStatistics: /* alter column set statistics */
			deparseAppendStringInfoString(state, "ALTER COLUMN ");
			options = "SET STATISTICS";
			break;
		case AT_SetOptions: /* alter column set ( options ) */
			deparseAppendStringInfoString(state, "ALTER COLUMN ");
			options = "SET";
			break;
		case AT_ResetOptions: /* alter column reset ( options ) */
			deparseAppendStringInfoString(state, "ALTER COLUMN ");
			options = "RESET";
			break;
		case AT_SetStorage: /* alter column set storage */
			deparseAppendStringInfoString(state, "ALTER COLUMN ");
			options = "SET STORAGE";
			break;
		case AT_SetCompression: /* alter column set compression */
			deparseAppendStringInfoString(state, "ALTER COLUMN ");
			options = "SET COMPRESSION";
			break;
		case AT_DropColumn: /* drop column */
			if (context == DEPARSE_NODE_CONTEXT_ALTER_TYPE)
				deparseAppendStringInfoString(state, "DROP ATTRIBUTE ");
			else
				deparseAppendStringInfoString(state, "DROP ");
			break;
		case AT_AddIndex: /* add index */
			deparseAppendStringInfoString(state, "ADD INDEX ");
			break;
		case AT_ReAddIndex: /* internal to commands/tablecmds.c */
			Assert(false);
			break;
		case AT_AddConstraint: /* add constraint */
			deparseAppendStringInfoString(state, "ADD ");
			break;
		case AT_ReAddConstraint: /* internal to commands/tablecmds.c */
			Assert(false);
			break;
		case AT_ReAddDomainConstraint: /* internal to commands/tablecmds.c */
			Assert(false);
			break;
		case AT_AlterConstraint: /* alter constraint */
			deparseAppendStringInfoString(state, "ALTER "); // CONSTRAINT keyword gets added by the Constraint itself (when deparsing def)
			break;
		case AT_ValidateConstraint: /* validate constraint */
			deparseAppendStringInfoString(state, "VALIDATE CONSTRAINT ");
			break;
		case AT_AddIndexConstraint: /* add constraint using existing index */
			// Not present in raw parser output
			Assert(false);
			break;
		case AT_DropConstraint: /* drop constraint */
			deparseAppendStringInfoString(state, "DROP CONSTRAINT ");
			break;
		case AT_ReAddComment: /* internal to commands/tablecmds.c */
		case AT_ReAddStatistics: /* internal to commands/tablecmds.c */
			Assert(false);
			break;
		case AT_AlterColumnType: /* alter column type */
			if (context == DEPARSE_NODE_CONTEXT_ALTER_TYPE)
				deparseAppendStringInfoString(state, "ALTER ATTRIBUTE ");
			else
				deparseAppendStringInfoString(state, "ALTER COLUMN ");
			options = "TYPE";
			break;
		case AT_AlterColumnGenericOptions: /* alter column OPTIONS (...) */
			deparseAppendStringInfoString(state, "ALTER COLUMN ");
			// Handled via special case in def handling
			break;
		case AT_ChangeOwner: /* change owner */
			deparseAppendStringInfoString(state, "OWNER TO ");
			deparseRoleSpec(state, alter_table_cmd->newowner);
			break;
		case AT_ClusterOn: /* CLUSTER ON */
			deparseAppendStringInfoString(state, "CLUSTER ON ");
			break;
		case AT_DropCluster: /* SET WITHOUT CLUSTER */
			deparseAppendStringInfoString(state, "SET WITHOUT CLUSTER ");
			break;
		case AT_SetLogged: /* SET LOGGED */
			deparseAppendStringInfoString(state, "SET LOGGED ");
			break;
		case AT_SetUnLogged: /* SET UNLOGGED */
			deparseAppendStringInfoString(state, "SET UNLOGGED ");
			break;
		case AT_DropOids: /* SET WITHOUT OIDS */
			deparseAppendStringInfoString(state, "SET WITHOUT OIDS ");
			break;
		case AT_SetTableSpace: /* SET TABLESPACE */
			deparseAppendStringInfoString(state, "SET TABLESPACE ");
			break;
		case AT_SetRelOptions: /* SET (...) -- AM specific parameters */
			deparseAppendStringInfoString(state, "SET ");
			break;
		case AT_SetAccessMethod:
			deparseAppendStringInfo(state, "SET ACCESS METHOD ");
			break;
		case AT_ResetRelOptions: /* RESET (...) -- AM specific parameters */
			deparseAppendStringInfoString(state, "RESET ");
			break;
		case AT_ReplaceRelOptions: /* replace reloption list in its entirety */
			// Not present in raw parser output
			Assert(false);
			break;
		case AT_EnableTrig: /* ENABLE TRIGGER name */
			deparseAppendStringInfoString(state, "ENABLE TRIGGER ");
			break;
		case AT_EnableAlwaysTrig: /* ENABLE ALWAYS TRIGGER name */
			deparseAppendStringInfoString(state, "ENABLE ALWAYS TRIGGER ");
			break;
		case AT_EnableReplicaTrig: /* ENABLE REPLICA TRIGGER name */
			deparseAppendStringInfoString(state, "ENABLE REPLICA TRIGGER ");
			break;
		case AT_DisableTrig: /* DISABLE TRIGGER name */
			deparseAppendStringInfoString(state, "DISABLE TRIGGER ");
			break;
		case AT_EnableTrigAll: /* ENABLE TRIGGER ALL */
			deparseAppendStringInfoString(state, "ENABLE TRIGGER ALL ");
			break;
		case AT_DisableTrigAll: /* DISABLE TRIGGER ALL */
			deparseAppendStringInfoString(state, "DISABLE TRIGGER ALL ");
			break;
		case AT_EnableTrigUser: /* ENABLE TRIGGER USER */
			deparseAppendStringInfoString(state, "ENABLE TRIGGER USER ");
			break;
		case AT_DisableTrigUser: /* DISABLE TRIGGER USER */
			deparseAppendStringInfoString(state, "DISABLE TRIGGER USER ");
			break;
		case AT_EnableRule: /* ENABLE RULE name */
			deparseAppendStringInfoString(state, "ENABLE RULE ");
			break;
		case AT_EnableAlwaysRule: /* ENABLE ALWAYS RULE name */
			deparseAppendStringInfoString(state, "ENABLE ALWAYS RULE ");
			break;
		case AT_EnableReplicaRule: /* ENABLE REPLICA RULE name */
			deparseAppendStringInfoString(state, "ENABLE REPLICA RULE ");
			break;
		case AT_DisableRule: /* DISABLE RULE name */
			deparseAppendStringInfoString(state, "DISABLE RULE ");
			break;
		case AT_AddInherit: /* INHERIT parent */
			deparseAppendStringInfoString(state, "INHERIT ");
			break;
		case AT_DropInherit: /* NO INHERIT parent */
			deparseAppendStringInfoString(state, "NO INHERIT ");
			break;
		case AT_AddOf: /* OF <type_name> */
			deparseAppendStringInfoString(state, "OF ");
			break;
		case AT_DropOf: /* NOT OF */
			deparseAppendStringInfoString(state, "NOT OF ");
			break;
		case AT_ReplicaIdentity: /* REPLICA IDENTITY */
			deparseAppendStringInfoString(state, "REPLICA IDENTITY ");
			break;
		case AT_EnableRowSecurity: /* ENABLE ROW SECURITY */
			deparseAppendStringInfoString(state, "ENABLE ROW LEVEL SECURITY ");
			break;
		case AT_DisableRowSecurity: /* DISABLE ROW SECURITY */
			deparseAppendStringInfoString(state, "DISABLE ROW LEVEL SECURITY ");
			break;
		case AT_ForceRowSecurity: /* FORCE ROW SECURITY */
			deparseAppendStringInfoString(state, "FORCE ROW LEVEL SECURITY ");
			break;
		case AT_NoForceRowSecurity: /* NO FORCE ROW SECURITY */
			deparseAppendStringInfoString(state, "NO FORCE ROW LEVEL SECURITY ");
			break;
		case AT_GenericOptions: /* OPTIONS (...) */
			// Handled in def field handling
			break;
		case AT_AttachPartition: /* ATTACH PARTITION */
			deparseAppendStringInfoString(state, "ATTACH PARTITION ");
			break;
		case AT_DetachPartition: /* DETACH PARTITION */
			deparseAppendStringInfoString(state, "DETACH PARTITION ");
			break;
		case AT_DetachPartitionFinalize: /* DETACH PARTITION FINALIZE */
			deparseAppendStringInfoString(state, "DETACH PARTITION ");
			break;
		case AT_AddIdentity: /* ADD IDENTITY */
			deparseAppendStringInfoString(state, "ALTER ");
			options = "ADD";
			// Other details are output via the constraint node (in def field)
			break;
		case AT_SetIdentity: /* SET identity column options */
			deparseAppendStringInfoString(state, "ALTER ");
			break;
		case AT_DropIdentity: /* DROP IDENTITY */
			deparseAppendStringInfoString(state, "ALTER COLUMN ");
			options = "DROP IDENTITY";
			trailing_missing_ok = true;
			break;
		case AT_SetExpression:
			deparseAppendStringInfoString(state, "ALTER COLUMN ");
			break;
	}

	if (alter_table_cmd->missing_ok && !trailing_missing_ok)
	{
		if (alter_table_cmd->subtype == AT_AddColumn)
			deparseAppendStringInfoString(state, "IF NOT EXISTS ");
		else
			deparseAppendStringInfoString(state, "IF EXISTS ");
	}

	if (alter_table_cmd->name != NULL)
	{
		deparseAppendStringInfoString(state, quote_identifier(alter_table_cmd->name));
		deparseAppendStringInfoChar(state, ' ');
	} else if (alter_table_cmd->subtype == AT_SetAccessMethod)
	{
		deparseAppendStringInfoString(state, " DEFAULT");
	}

	if (alter_table_cmd->num > 0)
		deparseAppendStringInfo(state, "%d ", alter_table_cmd->num);

	if (options != NULL)
	{
		deparseAppendStringInfoString(state, options);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (alter_table_cmd->missing_ok && trailing_missing_ok)
		deparseAppendStringInfoString(state, "IF EXISTS ");

	switch (alter_table_cmd->subtype)
	{
		case AT_AttachPartition:
		case AT_DetachPartition:
			deparsePartitionCmd(state, castNode(PartitionCmd, alter_table_cmd->def));
			deparseAppendStringInfoChar(state, ' ');
			break;
		case AT_DetachPartitionFinalize:
			deparsePartitionCmd(state, castNode(PartitionCmd, alter_table_cmd->def));
			deparseAppendStringInfoString(state, " FINALIZE ");
			break;
		case AT_AddColumn:
		case AT_AlterColumnType:
			deparseColumnDef(state, castNode(ColumnDef, alter_table_cmd->def));
			deparseAppendStringInfoChar(state, ' ');
			break;
		case AT_ColumnDefault:
			if (alter_table_cmd->def != NULL)
			{
				deparseExpr(state, alter_table_cmd->def, DEPARSE_NODE_CONTEXT_A_EXPR);
				deparseAppendStringInfoChar(state, ' ');
			}
			break;
		case AT_SetStatistics:
			if (alter_table_cmd->def != NULL)
				deparseSignedIconst(state, alter_table_cmd->def);
			else
				deparseAppendStringInfoString(state, "DEFAULT");
			deparseAppendStringInfoChar(state, ' ');
			break;
		case AT_SetOptions:
		case AT_ResetOptions:
		case AT_SetRelOptions:
		case AT_ResetRelOptions:
			deparseRelOptions(state, castNode(List, alter_table_cmd->def));
			deparseAppendStringInfoChar(state, ' ');
			break;
		case AT_SetStorage:
			deparseColId(state, strVal(alter_table_cmd->def));
			deparseAppendStringInfoChar(state, ' ');
			break;
		case AT_SetCompression:
			if (strcmp(strVal(alter_table_cmd->def), "default") == 0)
				deparseAppendStringInfoString(state, "DEFAULT");
			else
				deparseColId(state, strVal(alter_table_cmd->def));
			deparseAppendStringInfoChar(state, ' ');
			break;
		case AT_AddIdentity:
		case AT_AddConstraint:
			deparseConstraint(state, castNode(Constraint, alter_table_cmd->def), DEPARSE_NODE_CONTEXT_NONE);
			deparseAppendStringInfoChar(state, ' ');
			break;
		case AT_AlterConstraint:
			deparseATAlterConstraint(state, castNode(ATAlterConstraint, alter_table_cmd->def));
			deparseAppendStringInfoChar(state, ' ');
			break;
		case AT_SetIdentity:
			deparseAlterIdentityColumnOptionList(state, castNode(List, alter_table_cmd->def));
			deparseAppendStringInfoChar(state, ' ');
			break;
		case AT_AlterColumnGenericOptions:
		case AT_GenericOptions:
			deparseAlterGenericOptions(state, castNode(List, alter_table_cmd->def));
			deparseAppendStringInfoChar(state, ' ');
			break;
		case AT_AddInherit:
		case AT_DropInherit:
			deparseRangeVar(state, castNode(RangeVar, alter_table_cmd->def), DEPARSE_NODE_CONTEXT_NONE);
			deparseAppendStringInfoChar(state, ' ');
			break;
		case AT_AddOf:
			deparseTypeName(state, castNode(TypeName, alter_table_cmd->def));
			deparseAppendStringInfoChar(state, ' ');
			break;
		case AT_ReplicaIdentity:
			deparseReplicaIdentityStmt(state, castNode(ReplicaIdentityStmt, alter_table_cmd->def));
			deparseAppendStringInfoChar(state, ' ');
			break;
		case AT_SetExpression:
			deparseAppendStringInfoString(state, "SET EXPRESSION AS (");
			deparseExpr(state, alter_table_cmd->def, DEPARSE_NODE_CONTEXT_A_EXPR);
			deparseAppendStringInfoChar(state, ')');
			break;
		default:
			Assert(alter_table_cmd->def == NULL);
			break;
	}

	deparseOptDropBehavior(state, alter_table_cmd->behavior);

	removeTrailingSpace(state);
}

static DeparseNodeContext deparseAlterTableObjType(DeparseState *state, ObjectType type)
{
	switch (type)
	{
		case OBJECT_TABLE:
			deparseAppendStringInfoString(state, "TABLE ");
			break;
		case OBJECT_FOREIGN_TABLE:
			deparseAppendStringInfoString(state, "FOREIGN TABLE ");
			break;
		case OBJECT_INDEX:
			deparseAppendStringInfoString(state, "INDEX ");
			break;
		case OBJECT_SEQUENCE:
			deparseAppendStringInfoString(state, "SEQUENCE ");
			break;
		case OBJECT_VIEW:
			deparseAppendStringInfoString(state, "VIEW ");
			break;
		case OBJECT_MATVIEW:
			deparseAppendStringInfoString(state, "MATERIALIZED VIEW ");
			break;
		case OBJECT_TYPE:
			deparseAppendStringInfoString(state, "TYPE ");
			return DEPARSE_NODE_CONTEXT_ALTER_TYPE;
			break;
		default:
			Assert(false);
			break;
	}

	return DEPARSE_NODE_CONTEXT_NONE;
}

static void deparseAlterTableMoveAllStmt(DeparseState *state, AlterTableMoveAllStmt *move_all_stmt)
{
	deparseAppendStringInfoString(state, "ALTER ");
	deparseAlterTableObjType(state, move_all_stmt->objtype);

	deparseAppendStringInfoString(state, "ALL IN TABLESPACE ");
	deparseAppendStringInfoString(state, move_all_stmt->orig_tablespacename);
	deparseAppendStringInfoChar(state, ' ');

	if (move_all_stmt->roles)
	{
		deparseAppendStringInfoString(state, "OWNED BY ");
		deparseRoleList(state, move_all_stmt->roles);
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseAppendStringInfoString(state, "SET TABLESPACE ");
	deparseAppendStringInfoString(state, move_all_stmt->new_tablespacename);
	deparseAppendStringInfoChar(state, ' ');

	if (move_all_stmt->nowait)
	{
		deparseAppendStringInfoString(state, "NOWAIT");
	}
}

static void deparseAlterTableStmt(DeparseState *state, AlterTableStmt *alter_table_stmt)
{
	ListCell *lc;
	DeparseStateNestingLevel *parent_level = NULL;

	deparseAppendStringInfoString(state, "ALTER ");
	DeparseNodeContext context = deparseAlterTableObjType(state, alter_table_stmt->objtype);

	if (alter_table_stmt->missing_ok)
		deparseAppendStringInfoString(state, "IF EXISTS ");

	deparseRangeVar(state, alter_table_stmt->relation, context);
	deparseAppendStringInfoChar(state, ' ');

	parent_level = deparseStateIncreaseNestingLevel(state);
	foreach(lc, alter_table_stmt->cmds)
	{
		deparseAlterTableCmd(state, castNode(AlterTableCmd, lfirst(lc)), context);
		if (lnext(alter_table_stmt->cmds, lc))
			deparseAppendCommaAndPart(state);
	}
	deparseStateDecreaseNestingLevel(state, parent_level);
}

static void deparseAlterTableSpaceOptionsStmt(DeparseState *state, AlterTableSpaceOptionsStmt *alter_table_space_options_stmt)
{
	deparseAppendStringInfoString(state, "ALTER TABLESPACE ");
	deparseColId(state, alter_table_space_options_stmt->tablespacename);
	deparseAppendStringInfoChar(state, ' ');

	if (alter_table_space_options_stmt->isReset)
		deparseAppendStringInfoString(state, "RESET ");
	else
		deparseAppendStringInfoString(state, "SET ");

	deparseRelOptions(state, alter_table_space_options_stmt->options);
}

// "AlterDomainStmt" in gram.y
static void deparseAlterDomainStmt(DeparseState *state, AlterDomainStmt *alter_domain_stmt)
{
	deparseAppendStringInfoString(state, "ALTER DOMAIN ");
	deparseAnyName(state, alter_domain_stmt->typeName);
	deparseAppendStringInfoChar(state, ' ');

	switch (alter_domain_stmt->subtype)
	{
		case 'T':
			if (alter_domain_stmt->def != NULL)
			{
				deparseAppendStringInfoString(state, "SET DEFAULT ");
				deparseExpr(state, alter_domain_stmt->def, DEPARSE_NODE_CONTEXT_A_EXPR);
			}
			else
			{
				deparseAppendStringInfoString(state, "DROP DEFAULT");
			}
			break;
		case 'N':
			deparseAppendStringInfoString(state, "DROP NOT NULL");
			break;
		case 'O':
			deparseAppendStringInfoString(state, "SET NOT NULL");
			break;
		case 'C':
			deparseAppendStringInfoString(state, "ADD ");
			deparseConstraint(state, castNode(Constraint, alter_domain_stmt->def), DEPARSE_NODE_CONTEXT_ALTER_DOMAIN);
			break;
		case 'X':
			deparseAppendStringInfoString(state, "DROP CONSTRAINT ");
			if (alter_domain_stmt->missing_ok)
				deparseAppendStringInfoString(state, "IF EXISTS ");
			deparseAppendStringInfoString(state, quote_identifier(alter_domain_stmt->name));
			if (alter_domain_stmt->behavior == DROP_CASCADE)
				deparseAppendStringInfoString(state, " CASCADE");
			break;
		case 'V':
			deparseAppendStringInfoString(state, "VALIDATE CONSTRAINT ");
			deparseAppendStringInfoString(state, quote_identifier(alter_domain_stmt->name));
			break;
		default:
			// No other subtypes supported by the parser
			Assert(false);
	}
}

static void deparseRenameStmt(DeparseState *state, RenameStmt *rename_stmt)
{
	List *l = NULL;

	deparseAppendStringInfoString(state, "ALTER ");

	switch (rename_stmt->renameType)
	{
		case OBJECT_AGGREGATE:
			deparseAppendStringInfoString(state, "AGGREGATE ");
			break;
		case OBJECT_COLLATION:
			deparseAppendStringInfoString(state, "COLLATION ");
			break;
		case OBJECT_CONVERSION:
			deparseAppendStringInfoString(state, "CONVERSION ");
			break;
		case OBJECT_DATABASE:
			deparseAppendStringInfoString(state, "DATABASE ");
			break;
		case OBJECT_DOMAIN:
		case OBJECT_DOMCONSTRAINT:
			deparseAppendStringInfoString(state, "DOMAIN ");
			break;
		case OBJECT_FDW:
			deparseAppendStringInfoString(state, "FOREIGN DATA WRAPPER ");
			break;
		case OBJECT_FUNCTION:
			deparseAppendStringInfoString(state, "FUNCTION ");
			break;
		case OBJECT_ROLE:
			deparseAppendStringInfoString(state, "ROLE ");
			break;
		case OBJECT_LANGUAGE:
			deparseAppendStringInfoString(state, "LANGUAGE ");
			break;
		case OBJECT_OPCLASS:
			deparseAppendStringInfoString(state, "OPERATOR CLASS ");
			break;
		case OBJECT_OPFAMILY:
			deparseAppendStringInfoString(state, "OPERATOR FAMILY ");
			break;
		case OBJECT_POLICY:
			deparseAppendStringInfoString(state, "POLICY ");
			break;
		case OBJECT_PROCEDURE:
			deparseAppendStringInfoString(state, "PROCEDURE ");
			break;
		case OBJECT_PUBLICATION:
			deparseAppendStringInfoString(state, "PUBLICATION ");
			break;
		case OBJECT_ROUTINE:
			deparseAppendStringInfoString(state, "ROUTINE ");
			break;
		case OBJECT_SCHEMA:
			deparseAppendStringInfoString(state, "SCHEMA ");
			break;
		case OBJECT_FOREIGN_SERVER:
			deparseAppendStringInfoString(state, "SERVER ");
			break;
		case OBJECT_SUBSCRIPTION:
			deparseAppendStringInfoString(state, "SUBSCRIPTION ");
			break;
		case OBJECT_TABLE:
		case OBJECT_TABCONSTRAINT:
			deparseAppendStringInfoString(state, "TABLE ");
			break;
		case OBJECT_COLUMN:
			switch (rename_stmt->relationType)
			{
				case OBJECT_TABLE:
					deparseAppendStringInfoString(state, "TABLE ");
					break;
				case OBJECT_FOREIGN_TABLE:
					deparseAppendStringInfoString(state, "FOREIGN TABLE ");
					break;
				case OBJECT_VIEW:
					deparseAppendStringInfoString(state, "VIEW ");
					break;
				case OBJECT_MATVIEW:
					deparseAppendStringInfoString(state, "MATERIALIZED VIEW ");
					break;
				default:
					Assert(false);
			}
			break;
		case OBJECT_SEQUENCE:
			deparseAppendStringInfoString(state, "SEQUENCE ");
			break;
		case OBJECT_VIEW:
			deparseAppendStringInfoString(state, "VIEW ");
			break;
		case OBJECT_MATVIEW:
			deparseAppendStringInfoString(state, "MATERIALIZED VIEW ");
			break;
		case OBJECT_INDEX:
			deparseAppendStringInfoString(state, "INDEX ");
			break;
		case OBJECT_FOREIGN_TABLE:
			deparseAppendStringInfoString(state, "FOREIGN TABLE ");
			break;
		case OBJECT_RULE:
			deparseAppendStringInfoString(state, "RULE ");
			break;
		case OBJECT_TRIGGER:
			deparseAppendStringInfoString(state, "TRIGGER ");
			break;
		case OBJECT_EVENT_TRIGGER:
			deparseAppendStringInfoString(state, "EVENT TRIGGER ");
			break;
		case OBJECT_TABLESPACE:
			deparseAppendStringInfoString(state, "TABLESPACE ");
			break;
		case OBJECT_STATISTIC_EXT:
			deparseAppendStringInfoString(state, "STATISTICS ");
			break;
		case OBJECT_TSPARSER:
			deparseAppendStringInfoString(state, "TEXT SEARCH PARSER ");
			break;
		case OBJECT_TSDICTIONARY:
			deparseAppendStringInfoString(state, "TEXT SEARCH DICTIONARY ");
			break;
		case OBJECT_TSTEMPLATE:
			deparseAppendStringInfoString(state, "TEXT SEARCH TEMPLATE ");
			break;
		case OBJECT_TSCONFIGURATION:
			deparseAppendStringInfoString(state, "TEXT SEARCH CONFIGURATION ");
			break;
		case OBJECT_TYPE:
		case OBJECT_ATTRIBUTE:
			deparseAppendStringInfoString(state, "TYPE ");
			break;
		default:
			Assert(false);
			break;
	}

	if (rename_stmt->missing_ok)
		deparseAppendStringInfoString(state, "IF EXISTS ");

	switch (rename_stmt->renameType)
	{
		case OBJECT_AGGREGATE:
			deparseAggregateWithArgtypes(state, castNode(ObjectWithArgs, rename_stmt->object));
			deparseAppendStringInfoString(state, " RENAME ");
			break;
		case OBJECT_DOMCONSTRAINT:
			deparseAnyName(state, castNode(List, rename_stmt->object));
			deparseAppendStringInfoString(state, " RENAME CONSTRAINT ");
			deparseAppendStringInfoString(state, quote_identifier(rename_stmt->subname));
			deparseAppendStringInfoChar(state, ' ');
			break;
		case OBJECT_OPCLASS:
		case OBJECT_OPFAMILY:
			l = castNode(List, rename_stmt->object);
			deparseAnyNameSkipFirst(state, l);
			deparseAppendStringInfoString(state, " USING ");
			deparseAppendStringInfoString(state, quote_identifier(strVal(linitial(l))));
			deparseAppendStringInfoString(state, " RENAME ");
			break;
		case OBJECT_POLICY:
			deparseAppendStringInfoString(state, quote_identifier(rename_stmt->subname));
			deparseAppendStringInfoString(state, " ON ");
			deparseRangeVar(state, rename_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
			deparseAppendStringInfoString(state, " RENAME ");
			break;
		case OBJECT_FUNCTION:
		case OBJECT_PROCEDURE:
		case OBJECT_ROUTINE:
			deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, rename_stmt->object));
			deparseAppendStringInfoString(state, " RENAME ");
			break;
		case OBJECT_SUBSCRIPTION:
			deparseColId(state, strVal(rename_stmt->object));
			deparseAppendStringInfoString(state, " RENAME ");
			break;
		case OBJECT_TABLE:
		case OBJECT_SEQUENCE:
		case OBJECT_VIEW:
		case OBJECT_MATVIEW:
		case OBJECT_INDEX:
		case OBJECT_FOREIGN_TABLE:
			deparseRangeVar(state, rename_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
			deparseAppendStringInfoString(state, " RENAME ");
			break;
		case OBJECT_COLUMN:
			deparseRangeVar(state, rename_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
			deparseAppendStringInfoString(state, " RENAME COLUMN ");
			deparseAppendStringInfoString(state, quote_identifier(rename_stmt->subname));
			deparseAppendStringInfoChar(state, ' ');
			break;
		case OBJECT_TABCONSTRAINT:
			deparseRangeVar(state, rename_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
			deparseAppendStringInfoString(state, " RENAME CONSTRAINT ");
			deparseAppendStringInfoString(state, quote_identifier(rename_stmt->subname));
			deparseAppendStringInfoChar(state, ' ');
			break;
		case OBJECT_RULE:
		case OBJECT_TRIGGER:
			deparseAppendStringInfoString(state, quote_identifier(rename_stmt->subname));
			deparseAppendStringInfoString(state, " ON ");
			deparseRangeVar(state, rename_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
			deparseAppendStringInfoString(state, " RENAME ");
			break;
		case OBJECT_FDW:
		case OBJECT_LANGUAGE:
		case OBJECT_PUBLICATION:
		case OBJECT_FOREIGN_SERVER:
		case OBJECT_EVENT_TRIGGER:
			deparseAppendStringInfoString(state, quote_identifier(strVal(rename_stmt->object)));
			deparseAppendStringInfoString(state, " RENAME ");
			break;
		case OBJECT_DATABASE:
		case OBJECT_ROLE:
		case OBJECT_SCHEMA:
		case OBJECT_TABLESPACE:
			deparseAppendStringInfoString(state, quote_identifier(rename_stmt->subname));
			deparseAppendStringInfoString(state, " RENAME ");
			break;
		case OBJECT_COLLATION:
		case OBJECT_CONVERSION:
		case OBJECT_DOMAIN:
		case OBJECT_STATISTIC_EXT:
		case OBJECT_TSPARSER:
		case OBJECT_TSDICTIONARY:
		case OBJECT_TSTEMPLATE:
		case OBJECT_TSCONFIGURATION:
		case OBJECT_TYPE:
			deparseAnyName(state, castNode(List, rename_stmt->object));
			deparseAppendStringInfoString(state, " RENAME ");
			break;
		case OBJECT_ATTRIBUTE:
			deparseRangeVar(state, rename_stmt->relation, DEPARSE_NODE_CONTEXT_ALTER_TYPE);
			deparseAppendStringInfoString(state, " RENAME ATTRIBUTE ");
			deparseAppendStringInfoString(state, quote_identifier(rename_stmt->subname));
			deparseAppendStringInfoChar(state, ' ');
			break;
		default:
			Assert(false);
			break;
	}

	deparseAppendStringInfoString(state, "TO ");
	deparseAppendStringInfoString(state, quote_identifier(rename_stmt->newname));
	deparseAppendStringInfoChar(state, ' ');

	deparseOptDropBehavior(state, rename_stmt->behavior);

	removeTrailingSpace(state);
}

static void deparseTransactionStmt(DeparseState *state, TransactionStmt *transaction_stmt)
{
	ListCell *lc;
	switch (transaction_stmt->kind)
	{
		case TRANS_STMT_BEGIN:
			deparseAppendStringInfoString(state, "BEGIN ");
			deparseTransactionModeList(state, transaction_stmt->options);
			break;
		case TRANS_STMT_START:
			deparseAppendStringInfoString(state, "START TRANSACTION ");
			deparseTransactionModeList(state, transaction_stmt->options);
			break;
		case TRANS_STMT_COMMIT:
			deparseAppendStringInfoString(state, "COMMIT ");
			if (transaction_stmt->chain)
				deparseAppendStringInfoString(state, "AND CHAIN ");
			break;
		case TRANS_STMT_ROLLBACK:
			deparseAppendStringInfoString(state, "ROLLBACK ");
			if (transaction_stmt->chain)
				deparseAppendStringInfoString(state, "AND CHAIN ");
			break;
		case TRANS_STMT_SAVEPOINT:
			deparseAppendStringInfoString(state, "SAVEPOINT ");
			deparseAppendStringInfoString(state, quote_identifier(transaction_stmt->savepoint_name));
			break;
		case TRANS_STMT_RELEASE:
			deparseAppendStringInfoString(state, "RELEASE ");
			deparseAppendStringInfoString(state, quote_identifier(transaction_stmt->savepoint_name));
			break;
		case TRANS_STMT_ROLLBACK_TO:
			deparseAppendStringInfoString(state, "ROLLBACK ");
			deparseAppendStringInfoString(state, "TO SAVEPOINT ");
			deparseAppendStringInfoString(state, quote_identifier(transaction_stmt->savepoint_name));
			break;
		case TRANS_STMT_PREPARE:
			deparseAppendStringInfoString(state, "PREPARE TRANSACTION ");
			deparseStringLiteral(state, transaction_stmt->gid);
			break;
		case TRANS_STMT_COMMIT_PREPARED:
			deparseAppendStringInfoString(state, "COMMIT PREPARED ");
			deparseStringLiteral(state, transaction_stmt->gid);
			break;
		case TRANS_STMT_ROLLBACK_PREPARED:
			deparseAppendStringInfoString(state, "ROLLBACK PREPARED ");
			deparseStringLiteral(state, transaction_stmt->gid);
			break;
	}

	removeTrailingSpace(state);
}

static void deparseVariableSetStmt(DeparseState *state, VariableSetStmt* variable_set_stmt)
{
	if (variable_set_stmt == NULL)
		elog(ERROR, "deparse: unexpected NULL VariableSetStmt");

	ListCell *lc;

	switch (variable_set_stmt->kind)
	{
		case VAR_SET_VALUE: /* SET var = value */
			deparseAppendStringInfoString(state, "SET ");
			if (variable_set_stmt->is_local)
				deparseAppendStringInfoString(state, "LOCAL ");
			if (strcmp(variable_set_stmt->name, "timezone") == 0 && variable_set_stmt->jumble_args)
			{
				deparseAppendStringInfoString(state, "TIME ZONE ");
				deparseVarList(state, variable_set_stmt->args);
			}
			else if (strcmp(variable_set_stmt->name, "xmloption") == 0 && variable_set_stmt->jumble_args)
			{
				char *option = linitial_node(A_Const, variable_set_stmt->args)->val.sval.sval;
				deparseAppendStringInfoString(state, "XML OPTION ");
				if (strcmp(option, "DOCUMENT") == 0)
					deparseAppendStringInfoString(state, "DOCUMENT ");
				else if (strcmp(option, "CONTENT") == 0)
					deparseAppendStringInfoString(state, "CONTENT ");
				else
					deparseVarList(state, variable_set_stmt->args);
			}
			else
			{
				deparseVarName(state, variable_set_stmt->name);
				deparseAppendStringInfoString(state, " TO ");
				deparseVarList(state, variable_set_stmt->args);
			}
			break;
		case VAR_SET_DEFAULT: /* SET var TO DEFAULT */
			deparseAppendStringInfoString(state, "SET ");
			if (variable_set_stmt->is_local)
				deparseAppendStringInfoString(state, "LOCAL ");
			if (strcmp(variable_set_stmt->name, "timezone") == 0 && variable_set_stmt->jumble_args)
			{
				deparseAppendStringInfoString(state, "TIME ZONE DEFAULT");
			}
			else
			{
				deparseVarName(state, variable_set_stmt->name);
				deparseAppendStringInfoString(state, " TO DEFAULT");
			}
			break;
		case VAR_SET_CURRENT: /* SET var FROM CURRENT */
			deparseAppendStringInfoString(state, "SET ");
			if (variable_set_stmt->is_local)
				deparseAppendStringInfoString(state, "LOCAL ");
			deparseVarName(state, variable_set_stmt->name);
			deparseAppendStringInfoString(state, " FROM CURRENT");
			break;
		case VAR_SET_MULTI: /* special case for SET TRANSACTION ... */
			Assert(variable_set_stmt->name != NULL);
			deparseAppendStringInfoString(state, "SET ");
			if (variable_set_stmt->is_local)
				deparseAppendStringInfoString(state, "LOCAL ");
			if (strcmp(variable_set_stmt->name, "TRANSACTION") == 0)
			{
				deparseAppendStringInfoString(state, "TRANSACTION ");
				deparseTransactionModeList(state, variable_set_stmt->args);
			}
			else if (strcmp(variable_set_stmt->name, "SESSION CHARACTERISTICS") == 0)
			{
				deparseAppendStringInfoString(state, "SESSION CHARACTERISTICS AS TRANSACTION ");
				deparseTransactionModeList(state, variable_set_stmt->args);
			}
			else if (strcmp(variable_set_stmt->name, "TRANSACTION SNAPSHOT") == 0)
			{
				deparseAppendStringInfoString(state, "TRANSACTION SNAPSHOT ");
				deparseStringLiteral(state, strVal(&castNode(A_Const, linitial(variable_set_stmt->args))->val));
			}
			else
			{
				Assert(false);
			}
			break;
		case VAR_RESET: /* RESET var */
			deparseAppendStringInfoString(state, "RESET ");
			deparseVarName(state, variable_set_stmt->name);
			break;
		case VAR_RESET_ALL: /* RESET ALL */
			deparseAppendStringInfoString(state, "RESET ALL");
			break;
	}
}

static void deparseDropdbStmt(DeparseState *state, DropdbStmt *dropdb_stmt)
{
	ListCell *lc = NULL;

	deparseAppendStringInfoString(state, "DROP DATABASE ");
	if (dropdb_stmt->missing_ok)
		deparseAppendStringInfoString(state, "IF EXISTS ");

	deparseAppendStringInfoString(state, quote_identifier(dropdb_stmt->dbname));
	deparseAppendStringInfoChar(state, ' ');

	if (list_length(dropdb_stmt->options) > 0)
	{
		deparseAppendStringInfoChar(state, '(');
		foreach(lc, dropdb_stmt->options)
		{
			DefElem *def_elem = castNode(DefElem, lfirst(lc));
			if (strcmp(def_elem->defname, "force") == 0)
				deparseAppendStringInfoString(state, "FORCE");
			else
				Assert(false); // Currently there are other supported values

			if (lnext(dropdb_stmt->options, lc))
				deparseAppendStringInfoString(state, ", ");
		}
		deparseAppendStringInfoChar(state, ')');
	}

	removeTrailingSpace(state);
}

static void deparseVacuumStmt(DeparseState *state, VacuumStmt *vacuum_stmt)
{
	ListCell *lc = NULL;
	ListCell *lc2 = NULL;

	if (vacuum_stmt->is_vacuumcmd)
		deparseAppendStringInfoString(state, "VACUUM ");
	else
		deparseAppendStringInfoString(state, "ANALYZE ");

        deparseUtilityOptionList(state, vacuum_stmt->options);

	foreach(lc, vacuum_stmt->rels)
	{
		Assert(IsA(lfirst(lc), VacuumRelation));
		VacuumRelation *rel = castNode(VacuumRelation, lfirst(lc));

		deparseRangeVar(state, rel->relation, DEPARSE_NODE_CONTEXT_NONE);
		if (list_length(rel->va_cols) > 0)
		{
			deparseAppendStringInfoChar(state, '(');
			foreach(lc2, rel->va_cols)
			{
				deparseAppendStringInfoString(state, quote_identifier(strVal(lfirst(lc2))));
				if (lnext(rel->va_cols, lc2))
					deparseAppendStringInfoString(state, ", ");
			}
			deparseAppendStringInfoChar(state, ')');
		}

		if (lnext(vacuum_stmt->rels, lc))
			deparseAppendStringInfoString(state, ", ");
	}

	removeTrailingSpace(state);
}

static void deparseLoadStmt(DeparseState *state, LoadStmt *load_stmt)
{
	deparseAppendStringInfoString(state, "LOAD ");
	deparseStringLiteral(state, load_stmt->filename);
}

static void deparseLockStmt(DeparseState *state, LockStmt *lock_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "LOCK TABLE ");

	deparseRelationExprList(state, lock_stmt->relations);
	deparseAppendStringInfoChar(state, ' ');

	if (lock_stmt->mode != AccessExclusiveLock)
	{
		deparseAppendStringInfoString(state, "IN ");
		switch (lock_stmt->mode)
		{
			case AccessShareLock:
				deparseAppendStringInfoString(state, "ACCESS SHARE ");
				break;
			case RowShareLock:
				deparseAppendStringInfoString(state, "ROW SHARE ");
				break;
			case RowExclusiveLock:
				deparseAppendStringInfoString(state, "ROW EXCLUSIVE ");
				break;
			case ShareUpdateExclusiveLock:
				deparseAppendStringInfoString(state, "SHARE UPDATE EXCLUSIVE ");
				break;
			case ShareLock:
				deparseAppendStringInfoString(state, "SHARE ");
				break;
			case ShareRowExclusiveLock:
				deparseAppendStringInfoString(state, "SHARE ROW EXCLUSIVE ");
				break;
			case ExclusiveLock:
				deparseAppendStringInfoString(state, "EXCLUSIVE ");
				break;
			case AccessExclusiveLock:
				deparseAppendStringInfoString(state, "ACCESS EXCLUSIVE ");
				break;
			default:
				Assert(false);
				break;
		}
		deparseAppendStringInfoString(state, "MODE ");
	}

	if (lock_stmt->nowait)
		deparseAppendStringInfoString(state, "NOWAIT ");

	removeTrailingSpace(state);
}

static void deparseConstraintsSetStmt(DeparseState *state, ConstraintsSetStmt *constraints_set_stmt)
{
	deparseAppendStringInfoString(state, "SET CONSTRAINTS ");

	if (list_length(constraints_set_stmt->constraints) > 0)
	{
		deparseQualifiedNameList(state, constraints_set_stmt->constraints);
		deparseAppendStringInfoChar(state, ' ');
	}
	else
	{
		deparseAppendStringInfoString(state, "ALL ");
	}

	if (constraints_set_stmt->deferred)
		deparseAppendStringInfoString(state, "DEFERRED");
	else
		deparseAppendStringInfoString(state, "IMMEDIATE");
}

static void deparseExplainStmt(DeparseState *state, ExplainStmt *explain_stmt)
{
	ListCell *lc = NULL;
	char *defname = NULL;

	deparseAppendPartGroup(state, "EXPLAIN", DEPARSE_PART_NO_INDENT);

    deparseUtilityOptionList(state, explain_stmt->options);

	deparseExplainableStmt(state, explain_stmt->query);
}

static void deparseCopyStmt(DeparseState *state, CopyStmt *copy_stmt)
{
	ListCell *lc = NULL;
	ListCell *lc2 = NULL;

	deparseAppendPartGroup(state, "COPY", DEPARSE_PART_INDENT);

	if (copy_stmt->relation != NULL)
	{
		deparseRangeVar(state, copy_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
		if (list_length(copy_stmt->attlist) > 0)
		{
			deparseAppendStringInfoChar(state, '(');
			deparseColumnList(state, copy_stmt->attlist);
			deparseAppendStringInfoChar(state, ')');
		}
		deparseAppendStringInfoChar(state, ' ');
	}

	if (copy_stmt->query != NULL)
	{
		deparseAppendStringInfoChar(state, '(');
		deparsePreparableStmt(state, copy_stmt->query);
		deparseAppendStringInfoString(state, ") ");
	}

	if (copy_stmt->is_from)
		deparseAppendStringInfoString(state, "FROM ");
	else
		deparseAppendStringInfoString(state, "TO ");

	if (copy_stmt->is_program)
		deparseAppendStringInfoString(state, "PROGRAM ");

	if (copy_stmt->filename != NULL)
	{
		deparseStringLiteral(state, copy_stmt->filename);
		deparseAppendStringInfoChar(state, ' ');
	}
	else
	{
		if (copy_stmt->is_from)
			deparseAppendStringInfoString(state, "STDIN ");
		else
			deparseAppendStringInfoString(state, "STDOUT ");
	}

	if (list_length(copy_stmt->options) > 0)
	{
		// In some cases, equivalent expressions may have slightly different parse trees for `COPY`
		// statements. For example the following two statements result in different (but equivalent) parse
		// trees:
		//
		//   - COPY foo FROM STDIN CSV FREEZE
		//   - COPY foo FROM STDIN WITH (FORMAT CSV, FREEZE)
		//
		// In order to make sure we deparse to the "correct" version, we always try to deparse to the older
		// compact syntax first.
		//
		// The old syntax can be seen here in the Postgres 8.4 Reference:
		//     https://www.postgresql.org/docs/8.4/sql-copy.html

		bool old_fmt = true;

		// Loop over the options to see if any require the new `WITH (...)` syntax.
		foreach(lc, copy_stmt->options)
		{
			DefElem *def_elem = castNode(DefElem, lfirst(lc));

			if (strcmp(def_elem->defname, "freeze") == 0 && def_elem->arg && optBooleanValue(def_elem->arg))
			{}
			else if (strcmp(def_elem->defname, "header") == 0 && def_elem->arg && optBooleanValue(def_elem->arg))
			{}
			else if (strcmp(def_elem->defname, "format") == 0 && strcmp(strVal(def_elem->arg), "csv") == 0)
			{}
			else if (strcmp(def_elem->defname, "force_quote") == 0 && def_elem->arg && nodeTag(def_elem->arg) == T_List)
			{}
			else
			{
				old_fmt = false;
				break;
			}
		}

		// Branch to differing output modes, depending on if we can use the old syntax.
		if (old_fmt) {
			foreach(lc, copy_stmt->options)
			{
				DefElem *def_elem = castNode(DefElem, lfirst(lc));

				if (strcmp(def_elem->defname, "freeze") == 0 && def_elem->arg && optBooleanValue(def_elem->arg))
				{
					deparseAppendStringInfoString(state, "FREEZE ");
				}
				else if (strcmp(def_elem->defname, "header") == 0 && def_elem->arg && optBooleanValue(def_elem->arg))
				{
					deparseAppendStringInfoString(state, "HEADER ");
				}
				else if (strcmp(def_elem->defname, "format") == 0 && strcmp(strVal(def_elem->arg), "csv") == 0)
				{
					deparseAppendStringInfoString(state, "CSV ");
				}
				else if (strcmp(def_elem->defname, "force_quote") == 0 && def_elem->arg && nodeTag(def_elem->arg) == T_List)
				{
					deparseAppendStringInfoString(state, "FORCE QUOTE ");
					deparseColumnList(state, castNode(List, def_elem->arg));
				}
				else
				{
					// This isn't reachable, the conditions here are exactly the same as the first loop above.
					Assert(false);
				}
			}
		} else {
			deparseAppendStringInfoString(state, "WITH (");
			foreach(lc, copy_stmt->options)
			{
				DefElem *def_elem = castNode(DefElem, lfirst(lc));

				if (strcmp(def_elem->defname, "format") == 0)
				{
					deparseAppendStringInfoString(state, "FORMAT ");

					char *format = strVal(def_elem->arg);
					if (strcmp(format, "binary") == 0)
						deparseAppendStringInfoString(state, "BINARY");
					else if (strcmp(format, "csv") == 0)
						deparseAppendStringInfoString(state, "CSV");
					else if (strcmp(format, "text") == 0)
						deparseAppendStringInfoString(state, "TEXT");
					else
						Assert(false);
				}
				else if (strcmp(def_elem->defname, "freeze") == 0)
				{
					deparseAppendStringInfoString(state, "FREEZE");
					deparseOptBoolean(state, def_elem->arg);
				}
				else if (strcmp(def_elem->defname, "delimiter") == 0)
				{
					deparseAppendStringInfoString(state, "DELIMITER ");
					deparseStringLiteral(state, strVal(def_elem->arg));
				}
				else if (strcmp(def_elem->defname, "null") == 0)
				{
					deparseAppendStringInfoString(state, "NULL ");
					deparseStringLiteral(state, strVal(def_elem->arg));
				}
				else if (strcmp(def_elem->defname, "header") == 0)
				{
					deparseAppendStringInfoString(state, "HEADER");
					deparseOptBoolean(state, def_elem->arg);
				}
				else if (strcmp(def_elem->defname, "quote") == 0)
				{
					deparseAppendStringInfoString(state, "QUOTE ");
					deparseStringLiteral(state, strVal(def_elem->arg));
				}
				else if (strcmp(def_elem->defname, "escape") == 0)
				{
					deparseAppendStringInfoString(state, "ESCAPE ");
					deparseStringLiteral(state, strVal(def_elem->arg));
				}
				else if (strcmp(def_elem->defname, "force_quote") == 0)
				{
					deparseAppendStringInfoString(state, "FORCE_QUOTE ");
					if (def_elem->arg == NULL)
						elog(ERROR, "deparse: unexpected NULL arg for FORCE_QUOTE");
					if (IsA(def_elem->arg, A_Star))
					{
						deparseAppendStringInfoChar(state, '*');
					}
					else if (IsA(def_elem->arg, List))
					{
						deparseAppendStringInfoChar(state, '(');
						deparseColumnList(state, castNode(List, def_elem->arg));
						deparseAppendStringInfoChar(state, ')');
					}
					else
					{
						Assert(false);
					}
				}
				else if (strcmp(def_elem->defname, "force_not_null") == 0)
				{
					deparseAppendStringInfoString(state, "FORCE_NOT_NULL ");

					if (def_elem->arg == NULL)
						elog(ERROR, "deparse: unexpected NULL arg for FORCE_NOT_NULL");
					if (IsA(def_elem->arg, A_Star))
						deparseAStar(state, castNode(A_Star, def_elem->arg));
					else
					{
						deparseAppendStringInfoChar(state, '(');
						deparseColumnList(state, castNode(List, def_elem->arg));
						deparseAppendStringInfoChar(state, ')');
					}
				}
				else if (strcmp(def_elem->defname, "force_null") == 0)
				{
					deparseAppendStringInfoString(state, "FORCE_NULL ");

					if (def_elem->arg == NULL)
						elog(ERROR, "deparse: unexpected NULL arg for FORCE_NULL");
					if (IsA(def_elem->arg, A_Star))
						deparseAStar(state, castNode(A_Star, def_elem->arg));
					else
					{
						deparseAppendStringInfoChar(state, '(');
						deparseColumnList(state, castNode(List, def_elem->arg));
						deparseAppendStringInfoChar(state, ')');
					}
				}
				else if (strcmp(def_elem->defname, "encoding") == 0)
				{
					deparseAppendStringInfoString(state, "ENCODING ");
					deparseStringLiteral(state, strVal(def_elem->arg));
				}
				else
				{
					deparseAppendStringInfoString(state, quote_identifier(def_elem->defname));
					if (def_elem->arg != NULL)
						deparseAppendStringInfoChar(state, ' ');
					
					if (def_elem->arg == NULL)
					{
						// Nothing
					}
					else if (IsA(def_elem->arg, String))
					{
						deparseOptBooleanOrString(state, strVal(def_elem->arg));
					}
					else if (IsA(def_elem->arg, Integer) || IsA(def_elem->arg, Float))
					{
						deparseNumericOnly(state, (union ValUnion *) def_elem->arg);
					}
					else if (IsA(def_elem->arg, A_Star))
					{
						deparseAStar(state, castNode(A_Star, def_elem->arg));
					}
					else if (IsA(def_elem->arg, List))
					{
						List *l = castNode(List, def_elem->arg);
						deparseAppendStringInfoChar(state, '(');
						foreach(lc2, l)
						{
							deparseOptBooleanOrString(state, strVal(lfirst(lc2)));
							if (lnext(l, lc2))
								deparseAppendStringInfoString(state, ", ");
						}
						deparseAppendStringInfoChar(state, ')');
					}
				}

				if (lnext(copy_stmt->options, lc))
					deparseAppendStringInfoString(state, ", ");
			}
			deparseAppendStringInfoString(state, ") ");
		}
	}

	deparseWhereClause(state, copy_stmt->whereClause);

	removeTrailingSpace(state);
}

static void deparseDoStmt(DeparseState *state, DoStmt *do_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "DO ");

	foreach (lc, do_stmt->args)
	{
		DefElem *defel = castNode(DefElem, lfirst(lc));
		if (strcmp(defel->defname, "language") == 0)
		{
			deparseAppendStringInfoString(state, "LANGUAGE ");
			deparseAppendStringInfoString(state, quote_identifier(strVal(defel->arg)));
			deparseAppendStringInfoChar(state, ' ');
		}
		else if (strcmp(defel->defname, "as") == 0)
		{
			char *strval = strVal(defel->arg);
			const char *delim = "$$";
			if (strstr(strval, "$$") != NULL)
				delim = "$outer$";
			deparseAppendStringInfoString(state, delim);
			deparseAppendStringInfoString(state, strval);
			deparseAppendStringInfoString(state, delim);
			deparseAppendStringInfoChar(state, ' ');
		}
	}

	removeTrailingSpace(state);
}

static void deparseDiscardStmt(DeparseState *state, DiscardStmt *discard_stmt)
{
	deparseAppendStringInfoString(state, "DISCARD ");
	switch (discard_stmt->target)
	{
		case DISCARD_ALL:
			deparseAppendStringInfoString(state, "ALL");
			break;
		case DISCARD_PLANS:
			deparseAppendStringInfoString(state, "PLANS");
			break;
		case DISCARD_SEQUENCES:
			deparseAppendStringInfoString(state, "SEQUENCES");
			break;
		case DISCARD_TEMP:
			deparseAppendStringInfoString(state, "TEMP");
			break;
	}
}

static void deparseDefineStmt(DeparseState *state, DefineStmt *define_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "CREATE ");

	if (define_stmt->replace)
		deparseAppendStringInfoString(state, "OR REPLACE ");

	switch (define_stmt->kind)
	{
		case OBJECT_AGGREGATE:
			deparseAppendStringInfoString(state, "AGGREGATE ");
			break;
		case OBJECT_OPERATOR:
			deparseAppendStringInfoString(state, "OPERATOR ");
			break;
		case OBJECT_TYPE:
			deparseAppendStringInfoString(state, "TYPE ");
			break;
		case OBJECT_TSPARSER:
			deparseAppendStringInfoString(state, "TEXT SEARCH PARSER ");
			break;
		case OBJECT_TSDICTIONARY:
			deparseAppendStringInfoString(state, "TEXT SEARCH DICTIONARY ");
			break;
		case OBJECT_TSTEMPLATE:
			deparseAppendStringInfoString(state, "TEXT SEARCH TEMPLATE ");
			break;
		case OBJECT_TSCONFIGURATION:
			deparseAppendStringInfoString(state, "TEXT SEARCH CONFIGURATION ");
			break;
		case OBJECT_COLLATION:
			deparseAppendStringInfoString(state, "COLLATION ");
			break;
		default:
			// This shouldn't happen
			Assert(false);
			break;
	}

	if (define_stmt->if_not_exists)
		deparseAppendStringInfoString(state, "IF NOT EXISTS ");

	switch (define_stmt->kind)
	{
		case OBJECT_AGGREGATE:
			deparseFuncName(state, define_stmt->defnames);
			break;
		case OBJECT_OPERATOR:
			deparseAnyOperator(state, define_stmt->defnames);
			break;
		case OBJECT_TYPE:
		case OBJECT_TSPARSER:
		case OBJECT_TSDICTIONARY:
		case OBJECT_TSTEMPLATE:
		case OBJECT_TSCONFIGURATION:
		case OBJECT_COLLATION:
			deparseAnyName(state, define_stmt->defnames);
			break;
		default:
			Assert(false);
	}
	deparseAppendStringInfoChar(state, ' ');

	if (!define_stmt->oldstyle && define_stmt->kind == OBJECT_AGGREGATE)
	{
		deparseAggrArgs(state, define_stmt->args);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (define_stmt->kind == OBJECT_COLLATION &&
		list_length(define_stmt->definition) == 1 &&
		strcmp(castNode(DefElem, linitial(define_stmt->definition))->defname, "from") == 0)
	{
		deparseAppendStringInfoString(state, "FROM ");
		deparseAnyName(state, castNode(List, castNode(DefElem, linitial(define_stmt->definition))->arg));
	}
	else if (list_length(define_stmt->definition) > 0)
	{
		deparseDefinition(state, define_stmt->definition);
	}

	removeTrailingSpace(state);
}

static void deparseCompositeTypeStmt(DeparseState *state, CompositeTypeStmt *composite_type_stmt)
{
	ListCell *lc;
	RangeVar *typevar;
	DeparseStateNestingLevel *parent_level = NULL;

	deparseAppendStringInfoString(state, "CREATE TYPE ");
	deparseRangeVar(state, composite_type_stmt->typevar, DEPARSE_NODE_CONTEXT_CREATE_TYPE);

	deparseAppendStringInfoString(state, " AS (");
	parent_level = deparseStateIncreaseNestingLevel(state);
	foreach(lc, composite_type_stmt->coldeflist)
	{
		deparseColumnDef(state, castNode(ColumnDef, lfirst(lc)));
		if (lnext(composite_type_stmt->coldeflist, lc))
			deparseAppendCommaAndPart(state);
	}
	deparseStateDecreaseNestingLevel(state, parent_level);
	deparseAppendStringInfoChar(state, ')');
}

static void deparseCreateEnumStmt(DeparseState *state, CreateEnumStmt *create_enum_stmt)
{
	ListCell *lc;
	DeparseStateNestingLevel *parent_level = NULL;

	deparseAppendStringInfoString(state, "CREATE TYPE ");

	deparseAnyName(state, create_enum_stmt->typeName);
	deparseAppendStringInfoString(state, " AS ENUM (");
	parent_level = deparseStateIncreaseNestingLevel(state);
	foreach(lc, create_enum_stmt->vals)
	{
		deparseStringLiteral(state, strVal(lfirst(lc)));
		if (lnext(create_enum_stmt->vals, lc))
			deparseAppendCommaAndPart(state);
	}
	deparseStateDecreaseNestingLevel(state, parent_level);
	deparseAppendStringInfoChar(state, ')');
}

static void deparseCreateRangeStmt(DeparseState *state, CreateRangeStmt *create_range_stmt)
{
	deparseAppendStringInfoString(state, "CREATE TYPE ");
	deparseAnyName(state, create_range_stmt->typeName);
	deparseAppendStringInfoString(state, " AS RANGE ");
	deparseDefinition(state, create_range_stmt->params);
}

static void deparseAlterEnumStmt(DeparseState *state, AlterEnumStmt *alter_enum_stmt)
{
	deparseAppendStringInfoString(state, "ALTER TYPE ");
	deparseAnyName(state, alter_enum_stmt->typeName);
	deparseAppendStringInfoChar(state, ' ');

	if (alter_enum_stmt->oldVal == NULL)
	{
		deparseAppendStringInfoString(state, "ADD VALUE ");
		if (alter_enum_stmt->skipIfNewValExists)
			deparseAppendStringInfoString(state, "IF NOT EXISTS ");

		deparseStringLiteral(state, alter_enum_stmt->newVal);
		deparseAppendStringInfoChar(state, ' ');

		if (alter_enum_stmt->newValNeighbor)
		{
			if (alter_enum_stmt->newValIsAfter)
				deparseAppendStringInfoString(state, "AFTER ");
			else
				deparseAppendStringInfoString(state, "BEFORE ");
			deparseStringLiteral(state, alter_enum_stmt->newValNeighbor);
		}
	}
	else
	{
		deparseAppendStringInfoString(state, "RENAME VALUE ");
		deparseStringLiteral(state, alter_enum_stmt->oldVal);
		deparseAppendStringInfoString(state, " TO ");
		deparseStringLiteral(state, alter_enum_stmt->newVal);
	}

	removeTrailingSpace(state);
}

static void deparseAlterExtensionStmt(DeparseState *state, AlterExtensionStmt *alter_extension_stmt)
{
	ListCell *lc = NULL;

	deparseAppendStringInfoString(state, "ALTER EXTENSION ");
	deparseColId(state, alter_extension_stmt->extname);
	deparseAppendStringInfoString(state, " UPDATE ");
	foreach (lc, alter_extension_stmt->options)
	{
		DefElem *def_elem = castNode(DefElem, lfirst(lc));
		if (strcmp(def_elem->defname, "new_version") == 0)
		{
			deparseAppendStringInfoString(state, "TO ");
			deparseNonReservedWordOrSconst(state, strVal(def_elem->arg));
		}
		else
		{
			Assert(false);
		}
		deparseAppendStringInfoChar(state, ' ');
	}
	removeTrailingSpace(state);
}

static void deparseAlterExtensionContentsStmt(DeparseState *state, AlterExtensionContentsStmt *alter_extension_contents_stmt)
{
	List *l = NULL;

	deparseAppendStringInfoString(state, "ALTER EXTENSION ");
	deparseColId(state, alter_extension_contents_stmt->extname);
	deparseAppendStringInfoChar(state, ' ');

	if (alter_extension_contents_stmt->action == 1)
		deparseAppendStringInfoString(state, "ADD ");
	else if (alter_extension_contents_stmt->action == -1)
		deparseAppendStringInfoString(state, "DROP ");
	else
		Assert(false);

	switch (alter_extension_contents_stmt->objtype)
	{
		case OBJECT_ACCESS_METHOD:
			deparseAppendStringInfoString(state, "ACCESS METHOD ");
			break;
		case OBJECT_AGGREGATE:
			deparseAppendStringInfoString(state, "AGGREGATE ");
			break;
		case OBJECT_CAST:
			deparseAppendStringInfoString(state, "CAST ");
			break;
		case OBJECT_COLLATION:
			deparseAppendStringInfoString(state, "COLLATION ");
			break;
		case OBJECT_CONVERSION:
			deparseAppendStringInfoString(state, "CONVERSION ");
			break;
		case OBJECT_DOMAIN:
			deparseAppendStringInfoString(state, "DOMAIN ");
			break;
		case OBJECT_FUNCTION:
			deparseAppendStringInfoString(state, "FUNCTION ");
			break;
		case OBJECT_LANGUAGE:
			deparseAppendStringInfoString(state, "LANGUAGE ");
			break;
		case OBJECT_OPERATOR:
			deparseAppendStringInfoString(state, "OPERATOR ");
			break;
		case OBJECT_OPCLASS:
			deparseAppendStringInfoString(state, "OPERATOR CLASS ");
			break;
		case OBJECT_OPFAMILY:
			deparseAppendStringInfoString(state, "OPERATOR FAMILY ");
			break;
		case OBJECT_PROCEDURE:
			deparseAppendStringInfoString(state, "PROCEDURE ");
			break;
		case OBJECT_ROUTINE:
			deparseAppendStringInfoString(state, "ROUTINE ");
			break;
		case OBJECT_SCHEMA:
			deparseAppendStringInfoString(state, "SCHEMA ");
			break;
		case OBJECT_EVENT_TRIGGER:
			deparseAppendStringInfoString(state, "EVENT TRIGGER ");
			break;
		case OBJECT_TABLE:
			deparseAppendStringInfoString(state, "TABLE ");
			break;
		case OBJECT_TSPARSER:
			deparseAppendStringInfoString(state, "TEXT SEARCH PARSER ");
			break;
		case OBJECT_TSDICTIONARY:
			deparseAppendStringInfoString(state, "TEXT SEARCH DICTIONARY ");
			break;
		case OBJECT_TSTEMPLATE:
			deparseAppendStringInfoString(state, "TEXT SEARCH TEMPLATE ");
			break;
		case OBJECT_TSCONFIGURATION:
			deparseAppendStringInfoString(state, "TEXT SEARCH CONFIGURATION ");
			break;
		case OBJECT_SEQUENCE:
			deparseAppendStringInfoString(state, "SEQUENCE ");
			break;
		case OBJECT_VIEW:
			deparseAppendStringInfoString(state, "VIEW ");
			break;
		case OBJECT_MATVIEW:
			deparseAppendStringInfoString(state, "MATERIALIZED VIEW ");
			break;
		case OBJECT_FOREIGN_TABLE:
			deparseAppendStringInfoString(state, "FOREIGN TABLE ");
			break;
		case OBJECT_FDW:
			deparseAppendStringInfoString(state, "FOREIGN DATA WRAPPER ");
			break;
		case OBJECT_FOREIGN_SERVER:
			deparseAppendStringInfoString(state, "SERVER ");
			break;
		case OBJECT_TRANSFORM:
			deparseAppendStringInfoString(state, "TRANSFORM ");
			break;
		case OBJECT_TYPE:
			deparseAppendStringInfoString(state, "TYPE ");
			break;
		default:
			// No other object types are supported here in the parser
			Assert(false);
			break;
	}

	switch (alter_extension_contents_stmt->objtype)
	{
		// any_name
		case OBJECT_COLLATION:
		case OBJECT_CONVERSION:
		case OBJECT_TABLE:
		case OBJECT_TSPARSER:
		case OBJECT_TSDICTIONARY:
		case OBJECT_TSTEMPLATE:
		case OBJECT_TSCONFIGURATION:
		case OBJECT_SEQUENCE:
		case OBJECT_VIEW:
		case OBJECT_MATVIEW:
		case OBJECT_FOREIGN_TABLE:
			deparseAnyName(state, castNode(List, alter_extension_contents_stmt->object));
			break;
		// name
		case OBJECT_ACCESS_METHOD:
		case OBJECT_LANGUAGE:
		case OBJECT_SCHEMA:
		case OBJECT_EVENT_TRIGGER:
		case OBJECT_FDW:
		case OBJECT_FOREIGN_SERVER:
			deparseColId(state, strVal(alter_extension_contents_stmt->object));
			break;
		case OBJECT_AGGREGATE:
			deparseAggregateWithArgtypes(state, castNode(ObjectWithArgs, alter_extension_contents_stmt->object));
			break;
		case OBJECT_CAST:
			l = castNode(List, alter_extension_contents_stmt->object);
			Assert(list_length(l) == 2);
			deparseAppendStringInfoChar(state, '(');
			deparseTypeName(state, castNode(TypeName, linitial(l)));
			deparseAppendStringInfoString(state, " AS ");
			deparseTypeName(state, castNode(TypeName, lsecond(l)));
			deparseAppendStringInfoChar(state, ')');
			break;
		case OBJECT_DOMAIN:
		case OBJECT_TYPE:
			deparseTypeName(state, castNode(TypeName, alter_extension_contents_stmt->object));
			break;
		case OBJECT_FUNCTION:
		case OBJECT_PROCEDURE:
		case OBJECT_ROUTINE:
			deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, alter_extension_contents_stmt->object));
			break;
		case OBJECT_OPERATOR:
			deparseOperatorWithArgtypes(state, castNode(ObjectWithArgs, alter_extension_contents_stmt->object));
			break;
		case OBJECT_OPFAMILY:
		case OBJECT_OPCLASS:
			l = castNode(List, alter_extension_contents_stmt->object);
			Assert(list_length(l) == 2);
			deparseAnyNameSkipFirst(state, l);
			deparseAppendStringInfoString(state, " USING ");
			deparseColId(state, strVal(linitial(l)));
			break;
		case OBJECT_TRANSFORM:
			l = castNode(List, alter_extension_contents_stmt->object);
			deparseAppendStringInfoString(state, "FOR ");
			deparseTypeName(state, castNode(TypeName, linitial(l)));
			deparseAppendStringInfoString(state, " LANGUAGE ");
			deparseColId(state, strVal(lsecond(l)));
			break;
		default:
			Assert(false);
			break;
	}
}

static void deparseAccessPriv(DeparseState *state, AccessPriv *access_priv)
{
	ListCell *lc;

	if (access_priv->priv_name != NULL)
	{
		if (strcmp(access_priv->priv_name, "select") == 0)
			deparseAppendStringInfoString(state, "select");
		else if (strcmp(access_priv->priv_name, "references") == 0)
			deparseAppendStringInfoString(state, "references");
		else if (strcmp(access_priv->priv_name, "create") == 0)
			deparseAppendStringInfoString(state, "create");
		else
			deparseAppendStringInfoString(state, quote_identifier(access_priv->priv_name));
	}
	else
	{
		deparseAppendStringInfoString(state, "ALL");
	}
	deparseAppendStringInfoChar(state, ' ');

	if (list_length(access_priv->cols) > 0)
	{
		deparseAppendStringInfoChar(state, '(');
		deparseColumnList(state, access_priv->cols);
		deparseAppendStringInfoChar(state, ')');
	}

	removeTrailingSpace(state);
}

static void deparseGrantStmt(DeparseState *state, GrantStmt *grant_stmt)
{
	if (grant_stmt == NULL)
		elog(ERROR, "deparse: unexpected NULL GrantStmt");

	ListCell *lc;
	if (grant_stmt->is_grant)
		deparseAppendStringInfoString(state, "GRANT ");
	else
		deparseAppendStringInfoString(state, "REVOKE ");

	if (!grant_stmt->is_grant && grant_stmt->grant_option)
		deparseAppendStringInfoString(state, "GRANT OPTION FOR ");

	if (list_length(grant_stmt->privileges) > 0)
	{
		foreach(lc, grant_stmt->privileges)
		{
			deparseAccessPriv(state, castNode(AccessPriv, lfirst(lc)));
			if (lnext(grant_stmt->privileges, lc))
				deparseAppendStringInfoString(state, ", ");
		}
		deparseAppendStringInfoChar(state, ' ');
	}
	else
	{
		deparseAppendStringInfoString(state, "ALL ");
	}
	
	deparseAppendStringInfoString(state, "ON ");

	deparsePrivilegeTarget(state, grant_stmt->targtype, grant_stmt->objtype, grant_stmt->objects);
	deparseAppendStringInfoChar(state, ' ');

	if (grant_stmt->is_grant)
		deparseAppendStringInfoString(state, "TO ");
	else
		deparseAppendStringInfoString(state, "FROM ");

	foreach(lc, grant_stmt->grantees)
	{
		deparseRoleSpec(state, castNode(RoleSpec, lfirst(lc)));
		if (lnext(grant_stmt->grantees, lc))
			deparseAppendStringInfoChar(state, ',');
		deparseAppendStringInfoChar(state, ' ');
	}

	if (grant_stmt->is_grant && grant_stmt->grant_option)
		deparseAppendStringInfoString(state, "WITH GRANT OPTION ");

	deparseOptDropBehavior(state, grant_stmt->behavior);

	if (grant_stmt->grantor)
	{
		deparseAppendStringInfoString(state, "GRANTED BY ");
		deparseRoleSpec(state, castNode(RoleSpec, grant_stmt->grantor));
	}

	removeTrailingSpace(state);
}

static void deparseGrantRoleStmt(DeparseState *state, GrantRoleStmt *grant_role_stmt)
{
	ListCell *lc;

	if (grant_role_stmt->is_grant)
		deparseAppendStringInfoString(state, "GRANT ");
	else
		deparseAppendStringInfoString(state, "REVOKE ");

	if (!grant_role_stmt->is_grant && list_length(grant_role_stmt->opt)) {
		DefElem *defelem = castNode(DefElem, linitial(grant_role_stmt->opt));
		if (defelem->arg == NULL)
			elog(ERROR, "deparse: unexpected NULL arg in GRANT role option");
		Assert(!castNode(Boolean, defelem->arg)->boolval);

		if (strcmp("admin", defelem->defname) == 0) {
			deparseAppendStringInfoString(state, "ADMIN ");
		} else if (strcmp("inherit", defelem->defname) == 0) {
			deparseAppendStringInfoString(state, "INHERIT ");
		} else if (strcmp("set", defelem->defname) == 0) {
			deparseAppendStringInfoString(state, "SET ");
		}

		deparseAppendStringInfoString(state, "OPTION FOR ");
	}

	foreach(lc, grant_role_stmt->granted_roles)
	{
		deparseAccessPriv(state, castNode(AccessPriv, lfirst(lc)));
		if (lnext(grant_role_stmt->granted_roles, lc))
			deparseAppendStringInfoChar(state, ',');
		deparseAppendStringInfoChar(state, ' ');
	}

	if (grant_role_stmt->is_grant)
		deparseAppendStringInfoString(state, "TO ");
	else
		deparseAppendStringInfoString(state, "FROM ");

	deparseRoleList(state, grant_role_stmt->grantee_roles);
	deparseAppendStringInfoChar(state, ' ');

	if (grant_role_stmt->is_grant) {
		if (list_length(grant_role_stmt->opt) > 0) {
			deparseAppendStringInfoString(state, "WITH ");
		}

		foreach(lc, grant_role_stmt->opt) {
			DefElem *defelem = castNode(DefElem, lfirst(lc));
			if (defelem->arg == NULL)
				elog(ERROR, "deparse: unexpected NULL arg in GRANT role option");
			if (strcmp("admin", defelem->defname) == 0) {
				deparseAppendStringInfoString(state, "ADMIN ");
				deparseAppendStringInfoString(state, castNode(Boolean, defelem->arg)->boolval ? "OPTION" : "FALSE");
			} else if (strcmp("inherit", defelem->defname) == 0) {
				deparseAppendStringInfoString(state, "INHERIT ");
				deparseAppendStringInfoString(state, castNode(Boolean, defelem->arg)->boolval ? "OPTION" : "FALSE");
			} else if (strcmp("set", defelem->defname) == 0) {
				deparseAppendStringInfoString(state, "SET ");
				deparseAppendStringInfoString(state, castNode(Boolean, defelem->arg)->boolval ? "OPTION" : "FALSE");
			}

			if (lnext(grant_role_stmt->opt, lc)) {
				deparseAppendStringInfoChar(state, ',');
			}

			deparseAppendStringInfoChar(state, ' ');
		}
	}

	if (grant_role_stmt->grantor)
	{
		deparseAppendStringInfoString(state, "GRANTED BY ");
		deparseRoleSpec(state, castNode(RoleSpec, grant_role_stmt->grantor));
	}

	if (grant_role_stmt->behavior == DROP_CASCADE) {
		deparseAppendStringInfoString(state, "CASCADE ");
	}

	removeTrailingSpace(state);
}

static void deparseDropRoleStmt(DeparseState *state, DropRoleStmt *drop_role_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "DROP ROLE ");

	if (drop_role_stmt->missing_ok)
		deparseAppendStringInfoString(state, "IF EXISTS ");

	deparseRoleList(state, drop_role_stmt->roles);
}

static void deparseIndexStmt(DeparseState *state, IndexStmt *index_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "CREATE ");

	if (index_stmt->unique)
		deparseAppendStringInfoString(state, "UNIQUE ");

	deparseAppendStringInfoString(state, "INDEX ");

	if (index_stmt->concurrent)
		deparseAppendStringInfoString(state, "CONCURRENTLY ");

	if (index_stmt->if_not_exists)
		deparseAppendStringInfoString(state, "IF NOT EXISTS ");

	if (index_stmt->idxname != NULL)
	{
		deparseAppendStringInfoString(state, quote_identifier(index_stmt->idxname));
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseAppendPartGroup(state, "ON", DEPARSE_PART_INDENT);
	deparseRangeVar(state, index_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ' ');

	if (index_stmt->accessMethod != NULL)
	{
		deparseAppendStringInfoString(state, "USING ");
		deparseAppendStringInfoString(state, quote_identifier(index_stmt->accessMethod));
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseAppendStringInfoChar(state, '(');
	foreach (lc, index_stmt->indexParams)
	{
		deparseIndexElem(state, castNode(IndexElem, lfirst(lc)));
		if (lnext(index_stmt->indexParams, lc))
			deparseAppendStringInfoString(state, ", ");
	}
	deparseAppendStringInfoString(state, ") ");

	if (list_length(index_stmt->indexIncludingParams) > 0)
	{
		deparseAppendStringInfoString(state, "INCLUDE (");
		foreach (lc, index_stmt->indexIncludingParams)
		{
			deparseIndexElem(state, castNode(IndexElem, lfirst(lc)));
			if (lnext(index_stmt->indexIncludingParams, lc))
				deparseAppendStringInfoString(state, ", ");
		}
		deparseAppendStringInfoString(state, ") ");
	}

	if (index_stmt->nulls_not_distinct)
	{
		deparseAppendStringInfoString(state, "NULLS NOT DISTINCT ");
	}

	deparseOptWith(state, index_stmt->options);

	if (index_stmt->tableSpace != NULL)
	{
		deparseAppendStringInfoString(state, "TABLESPACE ");
		deparseAppendStringInfoString(state, quote_identifier(index_stmt->tableSpace));
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseWhereClause(state, index_stmt->whereClause);

	removeTrailingSpace(state);
}

static void deparseAlterOpFamilyStmt(DeparseState *state, AlterOpFamilyStmt *alter_op_family_stmt)
{
	deparseAppendStringInfoString(state, "ALTER OPERATOR FAMILY ");
	deparseAnyName(state, alter_op_family_stmt->opfamilyname);
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "USING ");
	deparseAppendStringInfoString(state, quote_identifier(alter_op_family_stmt->amname));
	deparseAppendStringInfoChar(state, ' ');

	if (alter_op_family_stmt->isDrop)
		deparseAppendStringInfoString(state, "DROP ");
	else
		deparseAppendStringInfoString(state, "ADD ");

	deparseOpclassItemList(state, alter_op_family_stmt->items);
}

static void deparsePrepareStmt(DeparseState *state, PrepareStmt *prepare_stmt)
{
	ListCell *lc = NULL;

	deparseAppendStringInfoString(state, "PREPARE ");
	deparseColId(state, prepare_stmt->name);
	if (list_length(prepare_stmt->argtypes) > 0)
	{
		deparseAppendStringInfoChar(state, '(');
		deparseTypeList(state, prepare_stmt->argtypes);
		deparseAppendStringInfoChar(state, ')');
	}
	deparseAppendStringInfoString(state, " AS ");
	deparsePreparableStmt(state, prepare_stmt->query);
}

static void deparseExecuteStmt(DeparseState *state, ExecuteStmt *execute_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "EXECUTE ");
	deparseAppendStringInfoString(state, quote_identifier(execute_stmt->name));
	if (list_length(execute_stmt->params) > 0)
	{
		deparseAppendStringInfoChar(state, '(');
		deparseExprList(state, execute_stmt->params);
		deparseAppendStringInfoChar(state, ')');
	}
}

static void deparseDeallocateStmt(DeparseState *state, DeallocateStmt *deallocate_stmt)
{
	deparseAppendStringInfoString(state, "DEALLOCATE ");
	if (deallocate_stmt->name != NULL)
		deparseAppendStringInfoString(state, quote_identifier(deallocate_stmt->name));
	else
		deparseAppendStringInfoString(state, "ALL");
}

// "AlterOptRoleElem" in gram.y
static void deparseAlterRoleElem(DeparseState *state, DefElem *def_elem)
{
	if (strcmp(def_elem->defname, "password") == 0)
	{
		deparseAppendStringInfoString(state, "PASSWORD ");
		if (def_elem->arg == NULL)
		{
			deparseAppendStringInfoString(state, "NULL");
		}
		else if (IsA(def_elem->arg, ParamRef))
		{
			deparseParamRef(state, castNode(ParamRef, def_elem->arg));
		}
		else if (IsA(def_elem->arg, String))
		{
			deparseStringLiteral(state, strVal(def_elem->arg));
		}
		else
		{
			Assert(false);
		}
	}
	else if (strcmp(def_elem->defname, "connectionlimit") == 0)
	{
		deparseAppendStringInfo(state, "CONNECTION LIMIT %d", intVal(def_elem->arg));
	}
	else if (strcmp(def_elem->defname, "validUntil") == 0)
	{
		deparseAppendStringInfoString(state, "VALID UNTIL ");
		deparseStringLiteral(state, strVal(def_elem->arg));
	}
	else if (strcmp(def_elem->defname, "superuser") == 0 && boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "SUPERUSER");
	}
	else if (strcmp(def_elem->defname, "superuser") == 0 && !boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "NOSUPERUSER");
	}
	else if (strcmp(def_elem->defname, "createrole") == 0 && boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "CREATEROLE");
	}
	else if (strcmp(def_elem->defname, "createrole") == 0 && !boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "NOCREATEROLE");
	}
	else if (strcmp(def_elem->defname, "isreplication") == 0 && boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "REPLICATION");
	}
	else if (strcmp(def_elem->defname, "isreplication") == 0 && !boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "NOREPLICATION");
	}
	else if (strcmp(def_elem->defname, "createdb") == 0 && boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "CREATEDB");
	}
	else if (strcmp(def_elem->defname, "createdb") == 0 && !boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "NOCREATEDB");
	}
	else if (strcmp(def_elem->defname, "canlogin") == 0 && boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "LOGIN");
	}
	else if (strcmp(def_elem->defname, "canlogin") == 0 && !boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "NOLOGIN");
	}
	else if (strcmp(def_elem->defname, "bypassrls") == 0 && boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "BYPASSRLS");
	}
	else if (strcmp(def_elem->defname, "bypassrls") == 0 && !boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "NOBYPASSRLS");
	}
	else if (strcmp(def_elem->defname, "inherit") == 0 && boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "INHERIT");
	}
	else if (strcmp(def_elem->defname, "inherit") == 0 && !boolVal(def_elem->arg))
	{
		deparseAppendStringInfoString(state, "NOINHERIT");
	}
	else
	{
		Assert(false);
	}
}

// "CreateOptRoleElem" in gram.y
static void deparseCreateRoleElem(DeparseState *state, DefElem *def_elem)
{
	if (strcmp(def_elem->defname, "sysid") == 0)
	{
		deparseAppendStringInfo(state, "SYSID %d", intVal(def_elem->arg));
	}
	else if (strcmp(def_elem->defname, "adminmembers") == 0)
	{
		deparseAppendStringInfoString(state, "ADMIN ");
		deparseRoleList(state, castNode(List, def_elem->arg));
	}
	else if (strcmp(def_elem->defname, "rolemembers") == 0)
	{
		deparseAppendStringInfoString(state, "ROLE ");
		deparseRoleList(state, castNode(List, def_elem->arg));
	}
	else if (strcmp(def_elem->defname, "addroleto") == 0)
	{
		deparseAppendStringInfoString(state, "IN ROLE ");
		deparseRoleList(state, castNode(List, def_elem->arg));
	}
	else
	{
		deparseAlterRoleElem(state, def_elem);
	}
}

static void deparseCreatePLangStmt(DeparseState *state, CreatePLangStmt *create_p_lang_stmt)
{
	deparseAppendStringInfoString(state, "CREATE ");

	if (create_p_lang_stmt->replace)
		deparseAppendStringInfoString(state, "OR REPLACE ");

	if (create_p_lang_stmt->pltrusted)
		deparseAppendStringInfoString(state, "TRUSTED ");

	deparseAppendStringInfoString(state, "LANGUAGE ");
	deparseNonReservedWordOrSconst(state, create_p_lang_stmt->plname);
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "HANDLER ");
	deparseHandlerName(state, create_p_lang_stmt->plhandler);
	deparseAppendStringInfoChar(state, ' ');

	if (create_p_lang_stmt->plinline)
	{
		deparseAppendStringInfoString(state, "INLINE ");
		deparseHandlerName(state, create_p_lang_stmt->plinline);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (create_p_lang_stmt->plvalidator)
	{
		deparseAppendStringInfoString(state, "VALIDATOR ");
		deparseHandlerName(state, create_p_lang_stmt->plvalidator);
		deparseAppendStringInfoChar(state, ' ');
	}

	removeTrailingSpace(state);
}

static void deparseCreateRoleStmt(DeparseState *state, CreateRoleStmt *create_role_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "CREATE ");

	switch (create_role_stmt->stmt_type)
	{
		case ROLESTMT_ROLE:
			deparseAppendStringInfoString(state, "ROLE ");
			break;
		case ROLESTMT_USER:
			deparseAppendStringInfoString(state, "USER ");
			break;
		case ROLESTMT_GROUP:
			deparseAppendStringInfoString(state, "GROUP ");
			break;
	}

	deparseAppendStringInfoString(state, quote_identifier(create_role_stmt->role));
	deparseAppendStringInfoChar(state, ' ');

	if (create_role_stmt->options != NULL)
	{
		deparseAppendStringInfoString(state, "WITH ");
		foreach (lc, create_role_stmt->options)
		{
			deparseCreateRoleElem(state, castNode(DefElem, lfirst(lc)));
			deparseAppendStringInfoChar(state, ' ');
		}
	}

	removeTrailingSpace(state);
}

static void deparseAlterRoleStmt(DeparseState *state, AlterRoleStmt *alter_role_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "ALTER ");

	if (list_length(alter_role_stmt->options) == 1 && strcmp(castNode(DefElem, linitial(alter_role_stmt->options))->defname, "rolemembers") == 0)
	{
		deparseAppendStringInfoString(state, "GROUP ");
		deparseRoleSpec(state, alter_role_stmt->role);
		deparseAppendStringInfoChar(state, ' ');

		if (alter_role_stmt->action == 1)
		{
			deparseAppendStringInfoString(state, "ADD USER ");
		}
		else if (alter_role_stmt->action == -1)
		{
			deparseAppendStringInfoString(state, "DROP USER ");
		}
		else
		{
			Assert(false);
		}

		deparseRoleList(state, castNode(List, castNode(DefElem, linitial(alter_role_stmt->options))->arg));
	}
	else
	{
		deparseAppendStringInfoString(state, "ROLE ");
		deparseRoleSpec(state, alter_role_stmt->role);
		deparseAppendStringInfoChar(state, ' ');

		deparseAppendStringInfoString(state, "WITH ");
		foreach (lc, alter_role_stmt->options)
		{
			deparseAlterRoleElem(state, castNode(DefElem, lfirst(lc)));
			deparseAppendStringInfoChar(state, ' ');
		}
	}

	removeTrailingSpace(state);
}

static void deparseDeclareCursorStmt(DeparseState *state, DeclareCursorStmt *declare_cursor_stmt)
{
	deparseAppendStringInfoString(state, "DECLARE ");
	deparseAppendStringInfoString(state, quote_identifier(declare_cursor_stmt->portalname));
	deparseAppendStringInfoChar(state, ' ');

	if (declare_cursor_stmt->options & CURSOR_OPT_BINARY)
		deparseAppendStringInfoString(state, "BINARY ");

	if (declare_cursor_stmt->options & CURSOR_OPT_SCROLL)
		deparseAppendStringInfoString(state, "SCROLL ");

	if (declare_cursor_stmt->options & CURSOR_OPT_NO_SCROLL)
		deparseAppendStringInfoString(state, "NO SCROLL ");

	if (declare_cursor_stmt->options & CURSOR_OPT_INSENSITIVE)
		deparseAppendStringInfoString(state, "INSENSITIVE ");

	deparseAppendStringInfoString(state, "CURSOR ");

	if (declare_cursor_stmt->options & CURSOR_OPT_HOLD)
		deparseAppendStringInfoString(state, "WITH HOLD ");

	deparseAppendStringInfoString(state, "FOR ");

	deparseSelectStmt(state, castNode(SelectStmt, declare_cursor_stmt->query), DEPARSE_NODE_CONTEXT_NONE);
}

static void deparseFetchStmt(DeparseState *state, FetchStmt *fetch_stmt)
{
	if (fetch_stmt->ismove)
		deparseAppendStringInfoString(state, "MOVE ");
	else
		deparseAppendStringInfoString(state, "FETCH ");

	switch (fetch_stmt->direction)
	{
		case FETCH_FORWARD:
			if (fetch_stmt->howMany == 1)
			{
				// Default
			}
			else if (fetch_stmt->howMany == FETCH_ALL)
			{
				deparseAppendStringInfoString(state, "ALL ");
			}
			else
			{
				deparseAppendStringInfo(state, "FORWARD %ld ", fetch_stmt->howMany);
			}
			break;
		case FETCH_BACKWARD:
			if (fetch_stmt->howMany == 1)
			{
				deparseAppendStringInfoString(state, "PRIOR ");
			}
			else if (fetch_stmt->howMany == FETCH_ALL)
			{
				deparseAppendStringInfoString(state, "BACKWARD ALL ");
			}
			else
			{
				deparseAppendStringInfo(state, "BACKWARD %ld ", fetch_stmt->howMany);
			}
			break;
		case FETCH_ABSOLUTE:
			if (fetch_stmt->howMany == 1)
			{
				deparseAppendStringInfoString(state, "FIRST ");
			}
			else if (fetch_stmt->howMany == -1)
			{
				deparseAppendStringInfoString(state, "LAST ");
			}
			else
			{
				deparseAppendStringInfo(state, "ABSOLUTE %ld ", fetch_stmt->howMany);
			}
			break;
		case FETCH_RELATIVE:
			deparseAppendStringInfo(state, "RELATIVE %ld ", fetch_stmt->howMany);
	}

	deparseAppendStringInfoString(state, quote_identifier(fetch_stmt->portalname));
}

static void deparseAlterDefaultPrivilegesStmt(DeparseState *state, AlterDefaultPrivilegesStmt *alter_default_privileges_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "ALTER DEFAULT PRIVILEGES ");

	foreach (lc, alter_default_privileges_stmt->options)
	{
		DefElem *defelem = castNode(DefElem, lfirst(lc));
		if (strcmp(defelem->defname, "schemas") == 0)
		{
			deparseAppendStringInfoString(state, "IN SCHEMA ");
			deparseNameList(state, castNode(List, defelem->arg));
			deparseAppendStringInfoChar(state, ' ');
		}
		else if (strcmp(defelem->defname, "roles") == 0)
		{
			deparseAppendStringInfoString(state, "FOR ROLE ");
			deparseRoleList(state, castNode(List, defelem->arg));
			deparseAppendStringInfoChar(state, ' ');
		}
		else
		{
			// No other DefElems are supported
			Assert(false);
		}
	}

	deparseGrantStmt(state, alter_default_privileges_stmt->action);
}

static void deparseReindexStmt(DeparseState *state, ReindexStmt *reindex_stmt)
{
	deparseAppendStringInfoString(state, "REINDEX ");

        deparseUtilityOptionList(state, reindex_stmt->params);

	switch (reindex_stmt->kind)
	{
		case REINDEX_OBJECT_INDEX:
			deparseAppendStringInfoString(state, "INDEX ");
			break;
		case REINDEX_OBJECT_TABLE:
			deparseAppendStringInfoString(state, "TABLE ");
			break;
		case REINDEX_OBJECT_SCHEMA:
			deparseAppendStringInfoString(state, "SCHEMA ");
			break;
		case REINDEX_OBJECT_SYSTEM:
			deparseAppendStringInfoString(state, "SYSTEM ");
			break;
		case REINDEX_OBJECT_DATABASE:
			deparseAppendStringInfoString(state, "DATABASE ");
			break;
	}

	if (reindex_stmt->relation != NULL)
	{
		deparseRangeVar(state, reindex_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
	}
	else if (reindex_stmt->name != NULL)
	{
		deparseAppendStringInfoString(state, quote_identifier(reindex_stmt->name));
	}
}

static void deparseRuleStmt(DeparseState *state, RuleStmt* rule_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "CREATE ");

	if (rule_stmt->replace)
		deparseAppendStringInfoString(state, "OR REPLACE ");

	deparseAppendStringInfoString(state, "RULE ");
	deparseAppendStringInfoString(state, quote_identifier(rule_stmt->rulename));
	deparseAppendStringInfoString(state, " AS ON ");

	switch (rule_stmt->event)
	{
		case CMD_UNKNOWN:
		case CMD_UTILITY:
		case CMD_NOTHING:
			// Not supported here
			Assert(false);
			break;
		case CMD_SELECT:
			deparseAppendStringInfoString(state, "SELECT ");
			break;
		case CMD_UPDATE:
			deparseAppendStringInfoString(state, "UPDATE ");
			break;
		case CMD_INSERT:
			deparseAppendStringInfoString(state, "INSERT ");
			break;
		case CMD_DELETE:
			deparseAppendStringInfoString(state, "DELETE ");
			break;
		case CMD_MERGE:
			deparseAppendStringInfoString(state, "MERGE ");
			break;
	}

	deparseAppendStringInfoString(state, "TO ");
	deparseRangeVar(state, rule_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ' ');

	deparseWhereClause(state, rule_stmt->whereClause);

	deparseAppendStringInfoString(state, "DO ");

	if (rule_stmt->instead)
		deparseAppendStringInfoString(state, "INSTEAD ");

	if (list_length(rule_stmt->actions) == 0)
	{
		deparseAppendStringInfoString(state, "NOTHING");
	}
	else if (list_length(rule_stmt->actions) == 1)
	{
		deparseRuleActionStmt(state, linitial(rule_stmt->actions));
	}
	else
	{
		deparseAppendStringInfoChar(state, '(');
		foreach (lc, rule_stmt->actions)
		{
			deparseRuleActionStmt(state, lfirst(lc));
			if (lnext(rule_stmt->actions, lc))
				deparseAppendStringInfoString(state, "; ");
		}
		deparseAppendStringInfoChar(state, ')');
	}
}

static void deparseNotifyStmt(DeparseState *state, NotifyStmt *notify_stmt)
{
	deparseAppendStringInfoString(state, "NOTIFY ");
	deparseAppendStringInfoString(state, quote_identifier(notify_stmt->conditionname));

	if (notify_stmt->payload != NULL)
	{
		deparseAppendStringInfoString(state, ", ");
		deparseStringLiteral(state, notify_stmt->payload);
	}
}

static void deparseListenStmt(DeparseState *state, ListenStmt *listen_stmt)
{
	deparseAppendStringInfoString(state, "LISTEN ");
	deparseAppendStringInfoString(state, quote_identifier(listen_stmt->conditionname));
}

static void deparseUnlistenStmt(DeparseState *state, UnlistenStmt *unlisten_stmt)
{
	deparseAppendStringInfoString(state, "UNLISTEN ");
	if (unlisten_stmt->conditionname == NULL)
		deparseAppendStringInfoString(state, "*");
	else
		deparseAppendStringInfoString(state, quote_identifier(unlisten_stmt->conditionname));
}

static void deparseCreateSeqStmt(DeparseState *state, CreateSeqStmt *create_seq_stmt)
{
	if (create_seq_stmt->sequence == NULL)
		elog(ERROR, "deparse: unexpected NULL sequence in CreateSeqStmt");

	ListCell *lc;

	deparseAppendStringInfoString(state, "CREATE ");

	deparseOptTemp(state, create_seq_stmt->sequence->relpersistence);

	deparseAppendStringInfoString(state, "SEQUENCE ");

	if (create_seq_stmt->if_not_exists)
		deparseAppendStringInfoString(state, "IF NOT EXISTS ");

	deparseRangeVar(state, create_seq_stmt->sequence, DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ' ');

	deparseOptSeqOptList(state, create_seq_stmt->options);

	removeTrailingSpace(state);
}

static void deparseAlterFunctionStmt(DeparseState *state, AlterFunctionStmt *alter_function_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "ALTER ");

	switch (alter_function_stmt->objtype)
	{
		case OBJECT_FUNCTION:
			deparseAppendStringInfoString(state, "FUNCTION ");
			break;
		case OBJECT_PROCEDURE:
			deparseAppendStringInfoString(state, "PROCEDURE ");
			break;
		case OBJECT_ROUTINE:
			deparseAppendStringInfoString(state, "ROUTINE ");
			break;
		default:
			// Not supported here
			Assert(false);
			break;
	}

	deparseFunctionWithArgtypes(state, alter_function_stmt->func);
	deparseAppendStringInfoChar(state, ' ');

	foreach (lc, alter_function_stmt->actions)
	{
		deparseCommonFuncOptItem(state, castNode(DefElem, lfirst(lc)));
		if (lnext(alter_function_stmt->actions, lc))
			deparseAppendStringInfoChar(state, ' ');
	}
}

static void deparseTruncateStmt(DeparseState *state, TruncateStmt *truncate_stmt)
{
	deparseAppendStringInfoString(state, "TRUNCATE ");

	deparseRelationExprList(state, truncate_stmt->relations);
	deparseAppendStringInfoChar(state, ' ');

	if (truncate_stmt->restart_seqs)
		deparseAppendStringInfoString(state, "RESTART IDENTITY ");

	deparseOptDropBehavior(state, truncate_stmt->behavior);

	removeTrailingSpace(state);
}

static void deparseCreateEventTrigStmt(DeparseState *state, CreateEventTrigStmt *create_event_trig_stmt)
{
	ListCell *lc = NULL;
	ListCell *lc2 = NULL;

	deparseAppendStringInfoString(state, "CREATE EVENT TRIGGER ");
	deparseAppendStringInfoString(state, quote_identifier(create_event_trig_stmt->trigname));
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "ON ");
	deparseAppendStringInfoString(state, quote_identifier(create_event_trig_stmt->eventname));
	deparseAppendStringInfoChar(state, ' ');

	if (create_event_trig_stmt->whenclause)
	{
		deparseAppendStringInfoString(state, "WHEN ");

		foreach (lc, create_event_trig_stmt->whenclause)
		{
			DefElem *def_elem = castNode(DefElem, lfirst(lc));
			List *l = castNode(List, def_elem->arg);
			deparseAppendStringInfoString(state, quote_identifier(def_elem->defname));
			deparseAppendStringInfoString(state, " IN (");
			foreach (lc2, l)
			{
				deparseStringLiteral(state, strVal(lfirst(lc2)));
				if (lnext(l, lc2))
					deparseAppendStringInfoString(state, ", ");
			}
			deparseAppendStringInfoChar(state, ')');
			if (lnext(create_event_trig_stmt->whenclause, lc))
				deparseAppendStringInfoString(state, " AND ");
		}

		deparseAppendStringInfoChar(state, ' ');
	}

	deparseAppendStringInfoString(state, "EXECUTE FUNCTION ");
	deparseFuncName(state, create_event_trig_stmt->funcname);
	deparseAppendStringInfoString(state, "()");
}

static void deparseAlterEventTrigStmt(DeparseState *state, AlterEventTrigStmt *alter_event_trig_stmt)
{
	deparseAppendStringInfoString(state, "ALTER EVENT TRIGGER ");
	deparseAppendStringInfoString(state, quote_identifier(alter_event_trig_stmt->trigname));
	deparseAppendStringInfoChar(state, ' ');

	switch (alter_event_trig_stmt->tgenabled)
	{
		case TRIGGER_FIRES_ON_ORIGIN:
			deparseAppendStringInfoString(state, "ENABLE");
			break;
		case TRIGGER_FIRES_ON_REPLICA:
			deparseAppendStringInfoString(state, "ENABLE REPLICA");
			break;
		case TRIGGER_FIRES_ALWAYS:
			deparseAppendStringInfoString(state, "ENABLE ALWAYS");
			break;
		case TRIGGER_DISABLED:
			deparseAppendStringInfoString(state, "DISABLE");
			break;
	}
}

static void deparseRefreshMatViewStmt(DeparseState *state, RefreshMatViewStmt *refresh_mat_view_stmt)
{
	deparseAppendStringInfoString(state, "REFRESH MATERIALIZED VIEW ");

	if (refresh_mat_view_stmt->concurrent)
		deparseAppendStringInfoString(state, "CONCURRENTLY ");

	deparseRangeVar(state, refresh_mat_view_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ' ');

	if (refresh_mat_view_stmt->skipData)
		deparseAppendStringInfoString(state, "WITH NO DATA ");

	removeTrailingSpace(state);
}

static void deparseReplicaIdentityStmt(DeparseState *state, ReplicaIdentityStmt *replica_identity_stmt)
{
	if (replica_identity_stmt == NULL)
		elog(ERROR, "deparse: unexpected NULL ReplicaIdentityStmt");

	switch (replica_identity_stmt->identity_type)
	{
		case REPLICA_IDENTITY_NOTHING:
			deparseAppendStringInfoString(state, "NOTHING ");
			break;
		case REPLICA_IDENTITY_FULL:
			deparseAppendStringInfoString(state, "FULL ");
			break;
		case REPLICA_IDENTITY_DEFAULT:
			deparseAppendStringInfoString(state, "DEFAULT ");
			break;
		case REPLICA_IDENTITY_INDEX:
			Assert(replica_identity_stmt->name != NULL);
			deparseAppendStringInfoString(state, "USING INDEX ");
			deparseAppendStringInfoString(state, quote_identifier(replica_identity_stmt->name));
			break;
	}
}

// "CreatePolicyStmt" in gram.y
static void deparseCreatePolicyStmt(DeparseState *state, CreatePolicyStmt *create_policy_stmt)
{
	ListCell *lc = NULL;

	deparseAppendStringInfoString(state, "CREATE POLICY ");
	deparseColId(state, create_policy_stmt->policy_name);
	deparseAppendStringInfoString(state, " ON ");
	deparseRangeVar(state, create_policy_stmt->table, DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ' ');

	if (!create_policy_stmt->permissive)
		deparseAppendStringInfoString(state, "AS RESTRICTIVE ");

	if (strcmp(create_policy_stmt->cmd_name, "all") == 0)
		Assert(true); // Default
	else if (strcmp(create_policy_stmt->cmd_name, "select") == 0)
		deparseAppendStringInfoString(state, "FOR SELECT ");
	else if (strcmp(create_policy_stmt->cmd_name, "insert") == 0)
		deparseAppendStringInfoString(state, "FOR INSERT ");
	else if (strcmp(create_policy_stmt->cmd_name, "update") == 0)
		deparseAppendStringInfoString(state, "FOR UPDATE ");
	else if (strcmp(create_policy_stmt->cmd_name, "delete") == 0)
		deparseAppendStringInfoString(state, "FOR DELETE ");
	else
		Assert(false);

	deparseAppendStringInfoString(state, "TO ");
	deparseRoleList(state, create_policy_stmt->roles);
	deparseAppendStringInfoChar(state, ' ');

	if (create_policy_stmt->qual != NULL)
	{
		deparseAppendStringInfoString(state, "USING (");
		deparseExpr(state, create_policy_stmt->qual, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, ") ");
	}

	if (create_policy_stmt->with_check != NULL)
	{
		deparseAppendStringInfoString(state, "WITH CHECK (");
		deparseExpr(state, create_policy_stmt->with_check, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, ") ");
	}
}

// "AlterPolicyStmt" in gram.y
static void deparseAlterPolicyStmt(DeparseState *state, AlterPolicyStmt *alter_policy_stmt)
{
	deparseAppendStringInfoString(state, "ALTER POLICY ");
	deparseAppendStringInfoString(state, quote_identifier(alter_policy_stmt->policy_name));
	deparseAppendStringInfoString(state, " ON ");
	deparseRangeVar(state, alter_policy_stmt->table, DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ' ');

	if (list_length(alter_policy_stmt->roles) > 0)
	{
		deparseAppendStringInfoString(state, "TO ");
		deparseRoleList(state, alter_policy_stmt->roles);
		deparseAppendStringInfoChar(state, ' ');
	}
	
	if (alter_policy_stmt->qual != NULL)
	{
		deparseAppendStringInfoString(state, "USING (");
		deparseExpr(state, alter_policy_stmt->qual, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, ") ");
	}

	if (alter_policy_stmt->with_check != NULL)
	{
		deparseAppendStringInfoString(state, "WITH CHECK (");
		deparseExpr(state, alter_policy_stmt->with_check, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, ") ");
	}
}

static void deparseCreateTableSpaceStmt(DeparseState *state, CreateTableSpaceStmt *create_table_space_stmt)
{
	deparseAppendStringInfoString(state, "CREATE TABLESPACE ");
	deparseColId(state, create_table_space_stmt->tablespacename);
	deparseAppendStringInfoChar(state, ' ');

	if (create_table_space_stmt->owner != NULL)
	{
		deparseAppendStringInfoString(state, "OWNER ");
		deparseRoleSpec(state, create_table_space_stmt->owner);
		deparseAppendStringInfoChar(state, ' ');
	}

	deparseAppendStringInfoString(state, "LOCATION ");

	if (create_table_space_stmt->location != NULL)
		deparseStringLiteral(state, create_table_space_stmt->location);
	else
		deparseAppendStringInfoString(state, "''");

	deparseAppendStringInfoChar(state, ' ');

	deparseOptWith(state, create_table_space_stmt->options);

	removeTrailingSpace(state);
}

static void deparseCreateTransformStmt(DeparseState *state, CreateTransformStmt *create_transform_stmt)
{
	deparseAppendStringInfoString(state, "CREATE ");
	if (create_transform_stmt->replace)
		deparseAppendStringInfoString(state, "OR REPLACE ");

	deparseAppendStringInfoString(state, "TRANSFORM FOR ");
	deparseTypeName(state, create_transform_stmt->type_name);
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "LANGUAGE ");
	deparseAppendStringInfoString(state, quote_identifier(create_transform_stmt->lang));
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoChar(state, '(');

	if (create_transform_stmt->fromsql)
	{
		deparseAppendStringInfoString(state, "FROM SQL WITH FUNCTION ");
		deparseFunctionWithArgtypes(state, create_transform_stmt->fromsql);
	}

	if (create_transform_stmt->fromsql && create_transform_stmt->tosql)
		deparseAppendStringInfoString(state, ", ");

	if (create_transform_stmt->tosql)
	{
		deparseAppendStringInfoString(state, "TO SQL WITH FUNCTION ");
		deparseFunctionWithArgtypes(state, create_transform_stmt->tosql);
	}

	deparseAppendStringInfoChar(state, ')');
}

static void deparseCreateAmStmt(DeparseState *state, CreateAmStmt *create_am_stmt)
{
	deparseAppendStringInfoString(state, "CREATE ACCESS METHOD ");
	deparseAppendStringInfoString(state, quote_identifier(create_am_stmt->amname));
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "TYPE ");
	switch (create_am_stmt->amtype)
	{
		case AMTYPE_INDEX:
			deparseAppendStringInfoString(state, "INDEX ");
			break;
		case AMTYPE_TABLE:
			deparseAppendStringInfoString(state, "TABLE ");
			break;
	}

	deparseAppendStringInfoString(state, "HANDLER ");
	deparseHandlerName(state, create_am_stmt->handler_name);
}

// "pub_obj_list" in gram.y
static void deparsePublicationObjectList(DeparseState *state, List *pubobjects) {
	const ListCell *lc;
	foreach(lc, pubobjects) {
		PublicationObjSpec *obj = lfirst(lc);

		switch (obj->pubobjtype) {
			case PUBLICATIONOBJ_TABLE:
				if (obj->pubtable == NULL)
					elog(ERROR, "deparse: unexpected NULL pubtable in PublicationObjSpec");
				deparseAppendStringInfoString(state, "TABLE ");
				deparseRangeVar(state, obj->pubtable->relation, DEPARSE_NODE_CONTEXT_NONE);
				
				if (obj->pubtable->columns)
				{
					deparseAppendStringInfoChar(state, '(');
					deparseColumnList(state, obj->pubtable->columns);
					deparseAppendStringInfoChar(state, ')');
				}

				if (obj->pubtable->whereClause)
				{
					deparseAppendStringInfoString(state, " WHERE (");
					deparseExpr(state, obj->pubtable->whereClause, DEPARSE_NODE_CONTEXT_A_EXPR);
					deparseAppendStringInfoString(state, ")");
				}

				break;
			case PUBLICATIONOBJ_TABLES_IN_SCHEMA:
				deparseAppendStringInfoString(state, "TABLES IN SCHEMA ");
				deparseAppendStringInfoString(state, quote_identifier(obj->name));
				break;
			case PUBLICATIONOBJ_TABLES_IN_CUR_SCHEMA:
				deparseAppendStringInfoString(state, "TABLES IN SCHEMA CURRENT_SCHEMA");
				break;
			case PUBLICATIONOBJ_CONTINUATION:
				// This should be unreachable, the parser merges these before we can even get here.
				Assert(false);
				break;
		}
		
		if (lnext(pubobjects, lc)) {
			deparseAppendStringInfoString(state, ", ");
		}
	}
}

static void deparseCreatePublicationStmt(DeparseState *state, CreatePublicationStmt *create_publication_stmt)
{
	ListCell *lc = NULL;

	deparseAppendStringInfoString(state, "CREATE PUBLICATION ");
	deparseAppendStringInfoString(state, quote_identifier(create_publication_stmt->pubname));
	deparseAppendStringInfoChar(state, ' ');

	if (list_length(create_publication_stmt->pubobjects) > 0)
	{
		deparseAppendStringInfoString(state, "FOR ");
		deparsePublicationObjectList(state, create_publication_stmt->pubobjects);
		deparseAppendStringInfoChar(state, ' ');
	}
	else if (create_publication_stmt->for_all_tables)
	{
		deparseAppendStringInfoString(state, "FOR ALL TABLES ");
	}

	deparseOptDefinition(state, create_publication_stmt->options);
	removeTrailingSpace(state);
}

static void deparseAlterPublicationStmt(DeparseState *state, AlterPublicationStmt *alter_publication_stmt)
{
	deparseAppendStringInfoString(state, "ALTER PUBLICATION ");
	deparseColId(state, alter_publication_stmt->pubname);
	deparseAppendStringInfoChar(state, ' ');

	if (list_length(alter_publication_stmt->pubobjects) > 0)
	{
		switch (alter_publication_stmt->action)
		{
			case AP_SetObjects:
				deparseAppendStringInfoString(state, "SET ");
				break;
			case AP_AddObjects:
				deparseAppendStringInfoString(state, "ADD ");
				break;
			case AP_DropObjects:
				deparseAppendStringInfoString(state, "DROP ");
				break;
		}

		deparsePublicationObjectList(state, alter_publication_stmt->pubobjects);
	}
	else if (list_length(alter_publication_stmt->options) > 0)
	{
		deparseAppendStringInfoString(state, "SET ");
		deparseDefinition(state, alter_publication_stmt->options);
	}
	else
	{
		Assert(false);
	}
}

static void deparseAlterSeqStmt(DeparseState *state, AlterSeqStmt *alter_seq_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "ALTER SEQUENCE ");

	if (alter_seq_stmt->missing_ok)
		deparseAppendStringInfoString(state, "IF EXISTS ");

	deparseRangeVar(state, alter_seq_stmt->sequence, DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ' ');

	deparseSeqOptList(state, alter_seq_stmt->options);

	removeTrailingSpace(state);
}

static void deparseAlterSystemStmt(DeparseState *state, AlterSystemStmt *alter_system_stmt)
{
	deparseAppendStringInfoString(state, "ALTER SYSTEM ");
	deparseVariableSetStmt(state, alter_system_stmt->setstmt);
}

static void deparseCommentStmt(DeparseState *state, CommentStmt *comment_stmt)
{
	ListCell *lc;
	List *l;

	deparseAppendStringInfoString(state, "COMMENT ON ");

	switch (comment_stmt->objtype)
	{
		case OBJECT_COLUMN:
			deparseAppendStringInfoString(state, "COLUMN ");
			break;
		case OBJECT_INDEX:
			deparseAppendStringInfoString(state, "INDEX ");
			break;
		case OBJECT_SEQUENCE:
			deparseAppendStringInfoString(state, "SEQUENCE ");
			break;
		case OBJECT_STATISTIC_EXT:
			deparseAppendStringInfoString(state, "STATISTICS ");
			break;
		case OBJECT_TABLE:
			deparseAppendStringInfoString(state, "TABLE ");
			break;
		case OBJECT_VIEW:
			deparseAppendStringInfoString(state, "VIEW ");
			break;
		case OBJECT_MATVIEW:
			deparseAppendStringInfoString(state, "MATERIALIZED VIEW ");
			break;
		case OBJECT_COLLATION:
			deparseAppendStringInfoString(state, "COLLATION ");
			break;
		case OBJECT_CONVERSION:
			deparseAppendStringInfoString(state, "CONVERSION ");
			break;
		case OBJECT_FOREIGN_TABLE:
			deparseAppendStringInfoString(state, "FOREIGN TABLE ");
			break;
		case OBJECT_TSCONFIGURATION:
			deparseAppendStringInfoString(state, "TEXT SEARCH CONFIGURATION ");
			break;
		case OBJECT_TSDICTIONARY:
			deparseAppendStringInfoString(state, "TEXT SEARCH DICTIONARY ");
			break;
		case OBJECT_TSPARSER:
			deparseAppendStringInfoString(state, "TEXT SEARCH PARSER ");
			break;
		case OBJECT_TSTEMPLATE:
			deparseAppendStringInfoString(state, "TEXT SEARCH TEMPLATE ");
			break;
		case OBJECT_ACCESS_METHOD:
			deparseAppendStringInfoString(state, "ACCESS METHOD ");
			break;
		case OBJECT_DATABASE:
			deparseAppendStringInfoString(state, "DATABASE ");
			break;
		case OBJECT_EVENT_TRIGGER:
			deparseAppendStringInfoString(state, "EVENT TRIGGER ");
			break;
		case OBJECT_EXTENSION:
			deparseAppendStringInfoString(state, "EXTENSION ");
			break;
		case OBJECT_FDW:
			deparseAppendStringInfoString(state, "FOREIGN DATA WRAPPER ");
			break;
		case OBJECT_LANGUAGE:
			deparseAppendStringInfoString(state, "LANGUAGE ");
			break;
		case OBJECT_PUBLICATION:
			deparseAppendStringInfoString(state, "PUBLICATION ");
			break;
		case OBJECT_ROLE:
			deparseAppendStringInfoString(state, "ROLE ");
			break;
		case OBJECT_SCHEMA:
			deparseAppendStringInfoString(state, "SCHEMA ");
			break;
		case OBJECT_FOREIGN_SERVER:
			deparseAppendStringInfoString(state, "SERVER ");
			break;
		case OBJECT_SUBSCRIPTION:
			deparseAppendStringInfoString(state, "SUBSCRIPTION ");
			break;
		case OBJECT_TABLESPACE:
			deparseAppendStringInfoString(state, "TABLESPACE ");
			break;
		case OBJECT_TYPE:
			deparseAppendStringInfoString(state, "TYPE ");
			break;
		case OBJECT_DOMAIN:
			deparseAppendStringInfoString(state, "DOMAIN ");
			break;
		case OBJECT_AGGREGATE:
			deparseAppendStringInfoString(state, "AGGREGATE ");
			break;
		case OBJECT_FUNCTION:
			deparseAppendStringInfoString(state, "FUNCTION ");
			break;
		case OBJECT_OPERATOR:
			deparseAppendStringInfoString(state, "OPERATOR ");
			break;
		case OBJECT_TABCONSTRAINT:
			deparseAppendStringInfoString(state, "CONSTRAINT ");
			break;
		case OBJECT_DOMCONSTRAINT:
			deparseAppendStringInfoString(state, "CONSTRAINT ");
			break;
		case OBJECT_POLICY:
			deparseAppendStringInfoString(state, "POLICY ");
			break;
		case OBJECT_PROCEDURE:
			deparseAppendStringInfoString(state, "PROCEDURE ");
			break;
		case OBJECT_ROUTINE:
			deparseAppendStringInfoString(state, "ROUTINE ");
			break;
		case OBJECT_RULE:
			deparseAppendStringInfoString(state, "RULE ");
			break;
		case OBJECT_TRANSFORM:
			deparseAppendStringInfoString(state, "TRANSFORM ");
			break;
		case OBJECT_TRIGGER:
			deparseAppendStringInfoString(state, "TRIGGER ");
			break;
		case OBJECT_OPCLASS:
			deparseAppendStringInfoString(state, "OPERATOR CLASS ");
			break;
		case OBJECT_OPFAMILY:
			deparseAppendStringInfoString(state, "OPERATOR FAMILY ");
			break;
		case OBJECT_LARGEOBJECT:
			deparseAppendStringInfoString(state, "LARGE OBJECT ");
			break;
		case OBJECT_CAST:
			deparseAppendStringInfoString(state, "CAST ");
			break;
		default:
			// No other cases are supported in the parser
			Assert(false);
			break;
	}

	switch (comment_stmt->objtype)
	{
		case OBJECT_COLUMN:
		case OBJECT_INDEX:
		case OBJECT_SEQUENCE:
		case OBJECT_STATISTIC_EXT:
		case OBJECT_TABLE:
		case OBJECT_VIEW:
		case OBJECT_MATVIEW:
		case OBJECT_COLLATION:
		case OBJECT_CONVERSION:
		case OBJECT_FOREIGN_TABLE:
		case OBJECT_TSCONFIGURATION:
		case OBJECT_TSDICTIONARY:
		case OBJECT_TSPARSER:
		case OBJECT_TSTEMPLATE:
			deparseAnyName(state, castNode(List, comment_stmt->object));
			break;
		case OBJECT_ACCESS_METHOD:
		case OBJECT_DATABASE:
		case OBJECT_EVENT_TRIGGER:
		case OBJECT_EXTENSION:
		case OBJECT_FDW:
		case OBJECT_LANGUAGE:
		case OBJECT_PUBLICATION:
		case OBJECT_ROLE:
		case OBJECT_SCHEMA:
		case OBJECT_FOREIGN_SERVER:
		case OBJECT_SUBSCRIPTION:
		case OBJECT_TABLESPACE:
			deparseAppendStringInfoString(state, quote_identifier(strVal(comment_stmt->object)));
			break;
		case OBJECT_TYPE:
		case OBJECT_DOMAIN:
			deparseTypeName(state, castNode(TypeName, comment_stmt->object));
			break;
		case OBJECT_AGGREGATE:
			deparseAggregateWithArgtypes(state, castNode(ObjectWithArgs, comment_stmt->object));
			break;
		case OBJECT_FUNCTION:
		case OBJECT_PROCEDURE:
		case OBJECT_ROUTINE:
			deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, comment_stmt->object));
			break;
		case OBJECT_OPERATOR:
			deparseOperatorWithArgtypes(state, castNode(ObjectWithArgs, comment_stmt->object));
			break;
		case OBJECT_TABCONSTRAINT:
		case OBJECT_POLICY:
		case OBJECT_RULE:
		case OBJECT_TRIGGER:
			l = castNode(List, comment_stmt->object);
			deparseAppendStringInfoString(state, quote_identifier(strVal(llast(l))));
			deparseAppendStringInfoString(state, " ON ");
			deparseAnyNameSkipLast(state, l);
			break;
		case OBJECT_DOMCONSTRAINT:
			l = castNode(List, comment_stmt->object);
			deparseAppendStringInfoString(state, quote_identifier(strVal(llast(l))));
			deparseAppendStringInfoString(state, " ON DOMAIN ");
			deparseTypeName(state, linitial(l));
			break;
		case OBJECT_TRANSFORM:
			l = castNode(List, comment_stmt->object);
			deparseAppendStringInfoString(state, "FOR ");
			deparseTypeName(state, castNode(TypeName, linitial(l)));
			deparseAppendStringInfoString(state, " LANGUAGE ");
			deparseAppendStringInfoString(state, quote_identifier(strVal(lsecond(l))));
			break;
		case OBJECT_OPCLASS:
		case OBJECT_OPFAMILY:
			l = castNode(List, comment_stmt->object);
			deparseAnyNameSkipFirst(state, l);
			deparseAppendStringInfoString(state, " USING ");
			deparseAppendStringInfoString(state, quote_identifier(strVal(linitial(l))));
			break;
		case OBJECT_LARGEOBJECT:
			deparseValue(state, (union ValUnion *) comment_stmt->object, DEPARSE_NODE_CONTEXT_NONE);
			break;
		case OBJECT_CAST:
			l = castNode(List, comment_stmt->object);
			deparseAppendStringInfoChar(state, '(');
			deparseTypeName(state, castNode(TypeName, linitial(l)));
			deparseAppendStringInfoString(state, " AS ");
			deparseTypeName(state, castNode(TypeName, lsecond(l)));
			deparseAppendStringInfoChar(state, ')');
			break;
		default:
			// No other cases are supported in the parser
			Assert(false);
			break;
	}

	deparseAppendStringInfoString(state, " IS ");

	if (comment_stmt->comment != NULL)
		deparseStringLiteral(state, comment_stmt->comment);
	else
		deparseAppendStringInfoString(state, "NULL");
}

// "stats_param" in gram.y
static void deparseStatsElem(DeparseState *state, StatsElem *stats_elem)
{
	// only one of stats_elem->name or stats_elem->expr can be non-null
	if (stats_elem->name)
		deparseAppendStringInfoString(state, stats_elem->name);
	else if (stats_elem->expr)
	{
		deparseAppendStringInfoChar(state, '(');
		deparseExpr(state, stats_elem->expr, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoChar(state, ')');
	}
}

static void deparseCreateStatsStmt(DeparseState *state, CreateStatsStmt *create_stats_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "CREATE STATISTICS ");

	if (create_stats_stmt->if_not_exists)
		deparseAppendStringInfoString(state, "IF NOT EXISTS ");

	deparseAnyName(state, create_stats_stmt->defnames);
	deparseAppendStringInfoChar(state, ' ');

	if (list_length(create_stats_stmt->stat_types) > 0)
	{
		deparseAppendStringInfoChar(state, '(');
		deparseNameList(state, create_stats_stmt->stat_types);
		deparseAppendStringInfoString(state, ") ");
	}

	deparseAppendStringInfoString(state, "ON ");
	foreach (lc, create_stats_stmt->exprs)
	{
		deparseStatsElem(state, lfirst(lc));
		if (lnext(create_stats_stmt->exprs, lc))
			deparseAppendStringInfoString(state, ", ");
	}

	deparseAppendStringInfoString(state, " FROM ");
	deparseFromList(state, create_stats_stmt->relations);
}

static void deparseAlterCollationStmt(DeparseState *state, AlterCollationStmt *alter_collation_stmt)
{
	deparseAppendStringInfoString(state, "ALTER COLLATION ");
	deparseAnyName(state, alter_collation_stmt->collname);
	deparseAppendStringInfoString(state, " REFRESH VERSION");
}

static void deparseAlterDatabaseStmt(DeparseState *state, AlterDatabaseStmt *alter_database_stmt)
{
	deparseAppendStringInfoString(state, "ALTER DATABASE ");
	deparseColId(state, alter_database_stmt->dbname);
	deparseAppendStringInfoChar(state, ' ');
	deparseCreatedbOptList(state, alter_database_stmt->options);
	removeTrailingSpace(state);
}

static void deparseAlterDatabaseSetStmt(DeparseState *state, AlterDatabaseSetStmt *alter_database_set_stmt)
{
	deparseAppendStringInfoString(state, "ALTER DATABASE ");
	deparseColId(state, alter_database_set_stmt->dbname);
	deparseAppendStringInfoChar(state, ' ');
	deparseVariableSetStmt(state, alter_database_set_stmt->setstmt);
}

static void deparseAlterStatsStmt(DeparseState *state, AlterStatsStmt *alter_stats_stmt)
{
	deparseAppendStringInfoString(state, "ALTER STATISTICS ");

	if (alter_stats_stmt->missing_ok)
		deparseAppendStringInfoString(state, "IF EXISTS ");

	deparseAnyName(state, alter_stats_stmt->defnames);
	deparseAppendStringInfoChar(state, ' ');

	if (alter_stats_stmt->stxstattarget)
		deparseAppendStringInfo(state, "SET STATISTICS %d", castNode(Integer, alter_stats_stmt->stxstattarget)->ival);
	else
		deparseAppendStringInfo(state, "SET STATISTICS DEFAULT");
}

static void deparseAlterTSDictionaryStmt(DeparseState *state, AlterTSDictionaryStmt *alter_ts_dictionary_stmt)
{
	deparseAppendStringInfoString(state, "ALTER TEXT SEARCH DICTIONARY ");

	deparseAnyName(state, alter_ts_dictionary_stmt->dictname);
	deparseAppendStringInfoChar(state, ' ');

	deparseDefinition(state, alter_ts_dictionary_stmt->options);
}

static void deparseAlterTSConfigurationStmt(DeparseState *state, AlterTSConfigurationStmt *alter_ts_configuration_stmt)
{
	ListCell *lc = NULL;

	deparseAppendStringInfoString(state, "ALTER TEXT SEARCH CONFIGURATION ");
	deparseAnyName(state, alter_ts_configuration_stmt->cfgname);
	deparseAppendStringInfoChar(state, ' ');

	switch (alter_ts_configuration_stmt->kind)
	{
		case ALTER_TSCONFIG_ADD_MAPPING:
			deparseAppendStringInfoString(state, "ADD MAPPING FOR ");
			deparseNameList(state, alter_ts_configuration_stmt->tokentype);
			deparseAppendStringInfoString(state, " WITH ");
			deparseAnyNameList(state, alter_ts_configuration_stmt->dicts);
			break;
		case ALTER_TSCONFIG_ALTER_MAPPING_FOR_TOKEN:
			deparseAppendStringInfoString(state, "ALTER MAPPING FOR ");
			deparseNameList(state, alter_ts_configuration_stmt->tokentype);
			deparseAppendStringInfoString(state, " WITH ");
			deparseAnyNameList(state, alter_ts_configuration_stmt->dicts);
			break;
		case ALTER_TSCONFIG_REPLACE_DICT:
			deparseAppendStringInfoString(state, "ALTER MAPPING REPLACE ");
			deparseAnyName(state, linitial(alter_ts_configuration_stmt->dicts));
			deparseAppendStringInfoString(state, " WITH ");
			deparseAnyName(state, lsecond(alter_ts_configuration_stmt->dicts));
			break;
		case ALTER_TSCONFIG_REPLACE_DICT_FOR_TOKEN:
			deparseAppendStringInfoString(state, "ALTER MAPPING FOR ");
			deparseNameList(state, alter_ts_configuration_stmt->tokentype);
			deparseAppendStringInfoString(state, " REPLACE ");
			deparseAnyName(state, linitial(alter_ts_configuration_stmt->dicts));
			deparseAppendStringInfoString(state, " WITH ");
			deparseAnyName(state, lsecond(alter_ts_configuration_stmt->dicts));
			break;
		case ALTER_TSCONFIG_DROP_MAPPING:
			deparseAppendStringInfoString(state, "DROP MAPPING ");
			if (alter_ts_configuration_stmt->missing_ok)
				deparseAppendStringInfoString(state, "IF EXISTS ");
			deparseAppendStringInfoString(state, "FOR ");
			deparseNameList(state, alter_ts_configuration_stmt->tokentype);
			break;
	}
}

static void deparseVariableShowStmt(DeparseState *state, VariableShowStmt *variable_show_stmt)
{
	deparseAppendStringInfoString(state, "SHOW ");

	if (strcmp(variable_show_stmt->name, "timezone") == 0)
		deparseAppendStringInfoString(state, "TIME ZONE");
	else if (strcmp(variable_show_stmt->name, "transaction_isolation") == 0)
		deparseAppendStringInfoString(state, "TRANSACTION ISOLATION LEVEL");
	else if (strcmp(variable_show_stmt->name, "session_authorization") == 0)
		deparseAppendStringInfoString(state, "SESSION AUTHORIZATION");
	else if (strcmp(variable_show_stmt->name, "all") == 0)
		deparseAppendStringInfoString(state, "ALL");
	else
		deparseAppendStringInfoString(state, quote_identifier(variable_show_stmt->name));
}

// "tablesample_clause" in gram.y
static void deparseRangeTableSample(DeparseState *state, RangeTableSample *range_table_sample)
{
	deparseRangeVar(state, castNode(RangeVar, range_table_sample->relation), DEPARSE_NODE_CONTEXT_NONE);

	deparseAppendStringInfoString(state, " TABLESAMPLE ");

	deparseFuncName(state, range_table_sample->method);
	deparseAppendStringInfoChar(state, '(');
	deparseExprList(state, range_table_sample->args);
	deparseAppendStringInfoString(state, ") ");

	if (range_table_sample->repeatable != NULL)
	{
		deparseAppendStringInfoString(state, "REPEATABLE (");
		deparseExpr(state, range_table_sample->repeatable, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, ") ");
	}

	removeTrailingSpace(state);
}

static void deparseCreateSubscriptionStmt(DeparseState *state, CreateSubscriptionStmt *create_subscription_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "CREATE SUBSCRIPTION ");
	deparseAppendStringInfoString(state, quote_identifier(create_subscription_stmt->subname));

	deparseAppendStringInfoString(state, " CONNECTION ");
	if (create_subscription_stmt->conninfo != NULL)
		deparseStringLiteral(state, create_subscription_stmt->conninfo);
	else
		deparseAppendStringInfoString(state, "''");

	deparseAppendStringInfoString(state, " PUBLICATION ");

	foreach(lc, create_subscription_stmt->publication)
	{
		deparseColLabel(state, strVal(lfirst(lc)));
		if (lnext(create_subscription_stmt->publication, lc))
			deparseAppendStringInfoString(state, ", ");
	}
	deparseAppendStringInfoChar(state, ' ');

	deparseOptDefinition(state, create_subscription_stmt->options);
	removeTrailingSpace(state);
}

static void deparseAlterSubscriptionStmt(DeparseState *state, AlterSubscriptionStmt *alter_subscription_stmt)
{
	ListCell *lc;

	deparseAppendStringInfoString(state, "ALTER SUBSCRIPTION ");
	deparseAppendStringInfoString(state, quote_identifier(alter_subscription_stmt->subname));
	deparseAppendStringInfoChar(state, ' ');

	switch (alter_subscription_stmt->kind)
	{
		case ALTER_SUBSCRIPTION_OPTIONS:
			deparseAppendStringInfoString(state, "SET ");
			deparseDefinition(state, alter_subscription_stmt->options);
			break;
		case ALTER_SUBSCRIPTION_SKIP:
			deparseAppendStringInfoString(state, "SKIP ");
			deparseDefinition(state, alter_subscription_stmt->options);
			break;
		case ALTER_SUBSCRIPTION_CONNECTION:
			deparseAppendStringInfoString(state, "CONNECTION ");
			deparseStringLiteral(state, alter_subscription_stmt->conninfo);
			deparseAppendStringInfoChar(state, ' ');
			break;
		case ALTER_SUBSCRIPTION_REFRESH:
			deparseAppendStringInfoString(state, "REFRESH PUBLICATION ");
			deparseOptDefinition(state, alter_subscription_stmt->options);
			break;
		case ALTER_SUBSCRIPTION_ADD_PUBLICATION:
			deparseAppendStringInfoString(state, "ADD PUBLICATION ");
			foreach(lc, alter_subscription_stmt->publication)
			{
				deparseColLabel(state, strVal(lfirst(lc)));
				if (lnext(alter_subscription_stmt->publication, lc))
					deparseAppendStringInfoString(state, ", ");
			}
			deparseAppendStringInfoChar(state, ' ');
			deparseOptDefinition(state, alter_subscription_stmt->options);
			break;
		case ALTER_SUBSCRIPTION_DROP_PUBLICATION:
			deparseAppendStringInfoString(state, "DROP PUBLICATION ");
			foreach(lc, alter_subscription_stmt->publication)
			{
				deparseColLabel(state, strVal(lfirst(lc)));
				if (lnext(alter_subscription_stmt->publication, lc))
					deparseAppendStringInfoString(state, ", ");
			}
			deparseAppendStringInfoChar(state, ' ');
			deparseOptDefinition(state, alter_subscription_stmt->options);
			break;
		case ALTER_SUBSCRIPTION_SET_PUBLICATION:
			deparseAppendStringInfoString(state, "SET PUBLICATION ");
			foreach(lc, alter_subscription_stmt->publication)
			{
				deparseColLabel(state, strVal(lfirst(lc)));
				if (lnext(alter_subscription_stmt->publication, lc))
					deparseAppendStringInfoString(state, ", ");
			}
			deparseAppendStringInfoChar(state, ' ');
			deparseOptDefinition(state, alter_subscription_stmt->options);
			break;
		case ALTER_SUBSCRIPTION_ENABLED:
			Assert(list_length(alter_subscription_stmt->options) == 1);
			DefElem *defelem = castNode(DefElem, linitial(alter_subscription_stmt->options));
			Assert(strcmp(defelem->defname, "enabled") == 0);
			if (optBooleanValue(defelem->arg))
			{
				deparseAppendStringInfoString(state, " ENABLE ");
			}
			else
			{
				deparseAppendStringInfoString(state, " DISABLE ");
			}
			break;
	}
	
	removeTrailingSpace(state);
}

static void deparseDropSubscriptionStmt(DeparseState *state, DropSubscriptionStmt *drop_subscription_stmt)
{
	deparseAppendStringInfoString(state, "DROP SUBSCRIPTION ");

	if (drop_subscription_stmt->missing_ok)
		deparseAppendStringInfoString(state, "IF EXISTS ");

	deparseAppendStringInfoString(state, drop_subscription_stmt->subname);
}

static void deparseCallStmt(DeparseState *state, CallStmt *call_stmt)
{
	deparseAppendStringInfoString(state, "CALL ");
	deparseFuncCall(state, call_stmt->funccall);
}

static void deparseAlterOwnerStmt(DeparseState *state, AlterOwnerStmt *alter_owner_stmt)
{
	List *l = NULL;

	deparseAppendStringInfoString(state, "ALTER ");

	switch (alter_owner_stmt->objectType)
	{
		case OBJECT_AGGREGATE:
			deparseAppendStringInfoString(state, "AGGREGATE ");
			deparseAggregateWithArgtypes(state, castNode(ObjectWithArgs, alter_owner_stmt->object));
			break;
		case OBJECT_COLLATION:
			deparseAppendStringInfoString(state, "COLLATION ");
			deparseAnyName(state, castNode(List, alter_owner_stmt->object));
			break;
		case OBJECT_CONVERSION:
			deparseAppendStringInfoString(state, "CONVERSION ");
			deparseAnyName(state, castNode(List, alter_owner_stmt->object));
			break;
		case OBJECT_DATABASE:
			deparseAppendStringInfoString(state, "DATABASE ");
			deparseColId(state, strVal(alter_owner_stmt->object));
			break;
		case OBJECT_DOMAIN:
			deparseAppendStringInfoString(state, "DOMAIN ");
			deparseAnyName(state, castNode(List, alter_owner_stmt->object));
			break;
		case OBJECT_FUNCTION:
			deparseAppendStringInfoString(state, "FUNCTION ");
			deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, alter_owner_stmt->object));
			break;
		case OBJECT_LANGUAGE:
			deparseAppendStringInfoString(state, "LANGUAGE ");
			deparseColId(state, strVal(alter_owner_stmt->object));
			break;
		case OBJECT_LARGEOBJECT:
			deparseAppendStringInfoString(state, "LARGE OBJECT ");
			deparseNumericOnly(state, (union ValUnion *) alter_owner_stmt->object);
			break;
		case OBJECT_OPERATOR:
			deparseAppendStringInfoString(state, "OPERATOR ");
			deparseOperatorWithArgtypes(state, castNode(ObjectWithArgs, alter_owner_stmt->object));
			break;
		case OBJECT_OPCLASS:
			l = castNode(List, alter_owner_stmt->object);
			deparseAppendStringInfoString(state, "OPERATOR CLASS ");
			deparseAnyNameSkipFirst(state, l);
			deparseAppendStringInfoString(state, " USING ");
			deparseColId(state, strVal(linitial(l)));
			break;
		case OBJECT_OPFAMILY:
			l = castNode(List, alter_owner_stmt->object);
			deparseAppendStringInfoString(state, "OPERATOR FAMILY ");
			deparseAnyNameSkipFirst(state, l);
			deparseAppendStringInfoString(state, " USING ");
			deparseColId(state, strVal(linitial(l)));
			break;
		case OBJECT_PROCEDURE:
			deparseAppendStringInfoString(state, "PROCEDURE ");
			deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, alter_owner_stmt->object));
			break;
		case OBJECT_ROUTINE:
			deparseAppendStringInfoString(state, "ROUTINE ");
			deparseFunctionWithArgtypes(state, castNode(ObjectWithArgs, alter_owner_stmt->object));
			break;
		case OBJECT_SCHEMA:
			deparseAppendStringInfoString(state, "SCHEMA ");
			deparseColId(state, strVal(alter_owner_stmt->object));
			break;
		case OBJECT_TYPE:
			deparseAppendStringInfoString(state, "TYPE ");
			deparseAnyName(state, castNode(List, alter_owner_stmt->object));
			break;
		case OBJECT_TABLESPACE:
			deparseAppendStringInfoString(state, "TABLESPACE ");
			deparseColId(state, strVal(alter_owner_stmt->object));
			break;
		case OBJECT_STATISTIC_EXT:
			deparseAppendStringInfoString(state, "STATISTICS ");
			deparseAnyName(state, castNode(List, alter_owner_stmt->object));
			break;
		case OBJECT_TSDICTIONARY:
			deparseAppendStringInfoString(state, "TEXT SEARCH DICTIONARY ");
			deparseAnyName(state, castNode(List, alter_owner_stmt->object));
			break;
		case OBJECT_TSCONFIGURATION:
			deparseAppendStringInfoString(state, "TEXT SEARCH CONFIGURATION ");
			deparseAnyName(state, castNode(List, alter_owner_stmt->object));
			break;
		case OBJECT_FDW:
			deparseAppendStringInfoString(state, "FOREIGN DATA WRAPPER ");
			deparseColId(state, strVal(alter_owner_stmt->object));
			break;
		case OBJECT_FOREIGN_SERVER:
			deparseAppendStringInfoString(state, "SERVER ");
			deparseColId(state, strVal(alter_owner_stmt->object));
			break;
		case OBJECT_EVENT_TRIGGER:
			deparseAppendStringInfoString(state, "EVENT TRIGGER ");
			deparseColId(state, strVal(alter_owner_stmt->object));
			break;
		case OBJECT_PUBLICATION:
			deparseAppendStringInfoString(state, "PUBLICATION ");
			deparseColId(state, strVal(alter_owner_stmt->object));
			break;
		case OBJECT_SUBSCRIPTION:
			deparseAppendStringInfoString(state, "SUBSCRIPTION ");
			deparseColId(state, strVal(alter_owner_stmt->object));
			break;
		default:
			Assert(false);
	}

	deparseAppendStringInfoString(state, " OWNER TO ");
	deparseRoleSpec(state, alter_owner_stmt->newowner);
}

// "operator_def_list" in gram.y
static void deparseOperatorDefList(DeparseState *state, List *defs)
{
	ListCell *lc = NULL;

	foreach (lc, defs)
	{
		DefElem *def_elem = castNode(DefElem, lfirst(lc));
		deparseAppendStringInfoString(state, quote_identifier(def_elem->defname));
		deparseAppendStringInfoString(state, " = ");
		if (def_elem->arg != NULL)
			deparseDefArg(state, def_elem->arg, true);
		else
			deparseAppendStringInfoString(state, "NONE");

		if (lnext(defs, lc))
			deparseAppendStringInfoString(state, ", ");
	}
}

static void deparseAlterOperatorStmt(DeparseState *state, AlterOperatorStmt *alter_operator_stmt)
{
	deparseAppendStringInfoString(state, "ALTER OPERATOR ");
	deparseOperatorWithArgtypes(state, alter_operator_stmt->opername);
	deparseAppendStringInfoString(state, " SET (");
	deparseOperatorDefList(state, alter_operator_stmt->options);
	deparseAppendStringInfoChar(state, ')');
}

static void deparseAlterTypeStmt(DeparseState *state, AlterTypeStmt *alter_type_stmt)
{
	deparseAppendStringInfoString(state, "ALTER TYPE ");
	deparseAnyName(state, alter_type_stmt->typeName);
	deparseAppendStringInfoString(state, " SET (");
	deparseOperatorDefList(state, alter_type_stmt->options);
	deparseAppendStringInfoChar(state, ')');
}

static void deparseDropOwnedStmt(DeparseState *state, DropOwnedStmt *drop_owned_stmt)
{
	deparseAppendStringInfoString(state, "DROP OWNED BY ");
	deparseRoleList(state, drop_owned_stmt->roles);
	deparseAppendStringInfoChar(state, ' ');
	deparseOptDropBehavior(state, drop_owned_stmt->behavior);
	removeTrailingSpace(state);
}

static void deparseReassignOwnedStmt(DeparseState *state, ReassignOwnedStmt *reassigned_owned_stmt)
{
	deparseAppendStringInfoString(state, "REASSIGN OWNED BY ");

	deparseRoleList(state, reassigned_owned_stmt->roles);
	deparseAppendStringInfoChar(state, ' ');

	deparseAppendStringInfoString(state, "TO ");
	deparseRoleSpec(state, reassigned_owned_stmt->newrole);
}

static void deparseClosePortalStmt(DeparseState *state, ClosePortalStmt *close_portal_stmt)
{
	deparseAppendStringInfoString(state, "CLOSE ");
	if (close_portal_stmt->portalname != NULL)
	{
		deparseAppendStringInfoString(state, quote_identifier(close_portal_stmt->portalname));
	}
	else
	{
		deparseAppendStringInfoString(state, "ALL");
	}
}

// "CreateTrigStmt" in gram.y
static void deparseCreateTrigStmt(DeparseState *state, CreateTrigStmt *create_trig_stmt)
{
	ListCell *lc;
	bool skip_events_or = true;

	deparseAppendStringInfoString(state, "CREATE ");
	if (create_trig_stmt->replace)
		deparseAppendStringInfoString(state, "OR REPLACE ");
	if (create_trig_stmt->isconstraint)
		deparseAppendStringInfoString(state, "CONSTRAINT ");
	deparseAppendStringInfoString(state, "TRIGGER ");

	deparseAppendStringInfoString(state, quote_identifier(create_trig_stmt->trigname));
	deparseAppendStringInfoChar(state, ' ');

	switch (create_trig_stmt->timing)
	{
		case TRIGGER_TYPE_BEFORE:
			deparseAppendStringInfoString(state, "BEFORE ");
			break;
		case TRIGGER_TYPE_AFTER:
			deparseAppendStringInfoString(state, "AFTER ");
			break;
		case TRIGGER_TYPE_INSTEAD:
			deparseAppendStringInfoString(state, "INSTEAD OF ");
			break;
		default:
			Assert(false);
	}

	if (TRIGGER_FOR_INSERT(create_trig_stmt->events))
	{
		deparseAppendStringInfoString(state, "INSERT ");
		skip_events_or = false;
	}
	if (TRIGGER_FOR_DELETE(create_trig_stmt->events))
	{
		if (!skip_events_or)
			deparseAppendStringInfoString(state, "OR ");
		deparseAppendStringInfoString(state, "DELETE ");
		skip_events_or = false;
	}
	if (TRIGGER_FOR_UPDATE(create_trig_stmt->events))
	{
		if (!skip_events_or)
			deparseAppendStringInfoString(state, "OR ");
		deparseAppendStringInfoString(state, "UPDATE ");
		if (list_length(create_trig_stmt->columns) > 0)
		{
			deparseAppendStringInfoString(state, "OF ");
			deparseColumnList(state, create_trig_stmt->columns);
			deparseAppendStringInfoChar(state, ' ');
		}
		skip_events_or = false;
	}
	if (TRIGGER_FOR_TRUNCATE(create_trig_stmt->events))
	{
		if (!skip_events_or)
			deparseAppendStringInfoString(state, "OR ");
		deparseAppendStringInfoString(state, "TRUNCATE ");
	}

	deparseAppendStringInfoString(state, "ON ");
	deparseRangeVar(state, create_trig_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
	deparseAppendStringInfoChar(state, ' ');

	if (create_trig_stmt->transitionRels != NULL)
	{
		deparseAppendStringInfoString(state, "REFERENCING ");
		foreach(lc, create_trig_stmt->transitionRels)
		{
			deparseTriggerTransition(state, castNode(TriggerTransition, lfirst(lc)));
			deparseAppendStringInfoChar(state, ' ');
		}
	}

	if (create_trig_stmt->constrrel != NULL)
	{
		deparseAppendStringInfoString(state, "FROM ");
		deparseRangeVar(state, create_trig_stmt->constrrel, DEPARSE_NODE_CONTEXT_NONE);
		deparseAppendStringInfoChar(state, ' ');
	}
	
	if (create_trig_stmt->deferrable)
		deparseAppendStringInfoString(state, "DEFERRABLE ");

	if (create_trig_stmt->initdeferred)
		deparseAppendStringInfoString(state, "INITIALLY DEFERRED ");

	if (create_trig_stmt->row)
		deparseAppendStringInfoString(state, "FOR EACH ROW ");

	if (create_trig_stmt->whenClause)
	{
		deparseAppendStringInfoString(state, "WHEN (");
		deparseExpr(state, create_trig_stmt->whenClause, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, ") ");
	}

	deparseAppendStringInfoString(state, "EXECUTE FUNCTION ");
	deparseFuncName(state, create_trig_stmt->funcname);
	deparseAppendStringInfoChar(state, '(');
	foreach(lc, create_trig_stmt->args)
	{
		deparseStringLiteral(state, strVal(lfirst(lc)));
		if (lnext(create_trig_stmt->args, lc))
			deparseAppendStringInfoString(state, ", ");
	}
	deparseAppendStringInfoChar(state, ')');
}

static void deparseTriggerTransition(DeparseState *state, TriggerTransition *trigger_transition)
{
	if (trigger_transition->isNew)
		deparseAppendStringInfoString(state, "NEW ");
	else
		deparseAppendStringInfoString(state, "OLD ");

	if (trigger_transition->isTable)
		deparseAppendStringInfoString(state, "TABLE ");
	else
		deparseAppendStringInfoString(state, "ROW ");

	deparseAppendStringInfoString(state, quote_identifier(trigger_transition->name));
}

static void deparseXmlExpr(DeparseState *state, XmlExpr* xml_expr, DeparseNodeContext context)
{
	switch (xml_expr->op)
	{
		case IS_XMLCONCAT: /* XMLCONCAT(args) */
			deparseAppendStringInfoString(state, "xmlconcat(");
			deparseExprList(state, xml_expr->args);
			deparseAppendStringInfoChar(state, ')');
			break;
		case IS_XMLELEMENT: /* XMLELEMENT(name, xml_attributes, args) */
			deparseAppendStringInfoString(state, "xmlelement(name ");
			deparseAppendStringInfoString(state, quote_identifier(xml_expr->name));
			if (xml_expr->named_args != NULL)
			{
				deparseAppendStringInfoString(state, ", xmlattributes(");
				deparseXmlAttributeList(state, xml_expr->named_args);
				deparseAppendStringInfoString(state, ")");
			}
			if (xml_expr->args != NULL)
			{
				deparseAppendStringInfoString(state, ", ");
				deparseExprList(state, xml_expr->args);
			}
			deparseAppendStringInfoString(state, ")");
			break;
		case IS_XMLFOREST: /* XMLFOREST(xml_attributes) */
			deparseAppendStringInfoString(state, "xmlforest(");
			deparseXmlAttributeList(state, xml_expr->named_args);
			deparseAppendStringInfoChar(state, ')');
			break;
		case IS_XMLPARSE: /* XMLPARSE(text, is_doc, preserve_ws) */
			Assert(list_length(xml_expr->args) == 2);
			deparseAppendStringInfoString(state, "xmlparse(");
			switch (xml_expr->xmloption)
			{
				case XMLOPTION_DOCUMENT:
					deparseAppendStringInfoString(state, "document ");
					break;
				case XMLOPTION_CONTENT:
					deparseAppendStringInfoString(state, "content ");
					break;
				default:
					Assert(false);
			}
			deparseExpr(state, linitial(xml_expr->args), DEPARSE_NODE_CONTEXT_A_EXPR);
			deparseAppendStringInfoChar(state, ')');
			break;
		case IS_XMLPI: /* XMLPI(name [, args]) */
			deparseAppendStringInfoString(state, "xmlpi(name ");
			deparseAppendStringInfoString(state, quote_identifier(xml_expr->name));
			if (xml_expr->args != NULL)
			{
				deparseAppendStringInfoString(state, ", ");
				deparseExpr(state, linitial(xml_expr->args), DEPARSE_NODE_CONTEXT_A_EXPR);
			}
			deparseAppendStringInfoChar(state, ')');
			break;
		case IS_XMLROOT: /* XMLROOT(xml, version, standalone) */
			deparseAppendStringInfoString(state, "xmlroot(");
			deparseExpr(state, linitial(xml_expr->args), DEPARSE_NODE_CONTEXT_A_EXPR);
			deparseAppendStringInfoString(state, ", version ");
			if (castNode(A_Const, lsecond(xml_expr->args))->isnull)
				deparseAppendStringInfoString(state, "no value");
			else
				deparseExpr(state, lsecond(xml_expr->args), DEPARSE_NODE_CONTEXT_A_EXPR);
			if (intVal(&castNode(A_Const, lthird(xml_expr->args))->val) == XML_STANDALONE_YES)
				deparseAppendStringInfoString(state, ", standalone yes");
			else if (intVal(&castNode(A_Const, lthird(xml_expr->args))->val) == XML_STANDALONE_NO)
				deparseAppendStringInfoString(state, ", standalone no");
			else if (intVal(&castNode(A_Const, lthird(xml_expr->args))->val) == XML_STANDALONE_NO_VALUE)
				deparseAppendStringInfoString(state, ", standalone no value");
			deparseAppendStringInfoChar(state, ')');
			break;
		case IS_XMLSERIALIZE: /* XMLSERIALIZE(is_document, xmlval) */
			// These are represented as XmlSerialize in raw parse trees
			Assert(false);
			break;
		case IS_DOCUMENT: /* xmlval IS DOCUMENT */
			Assert(list_length(xml_expr->args) == 1);
			deparseExprOperand(state, linitial(xml_expr->args), DEPARSE_PREC_IS, DEPARSE_OPERAND_LEFT, context);
			deparseAppendStringInfoString(state, " IS DOCUMENT");
			break;
	}
}

// "xmltable_column_el" in gram.y
static void deparseRangeTableFuncCol(DeparseState *state, RangeTableFuncCol* range_table_func_col)
{
	deparseAppendStringInfoString(state, quote_identifier(range_table_func_col->colname));
	deparseAppendStringInfoChar(state, ' ');

	if (range_table_func_col->for_ordinality)
	{
		deparseAppendStringInfoString(state, "FOR ORDINALITY ");
	}
	else
	{
		deparseTypeName(state, range_table_func_col->typeName);
		deparseAppendStringInfoChar(state, ' ');

		if (range_table_func_col->colexpr)
		{
			deparseAppendStringInfoString(state, "PATH ");
			deparseBExpr(state, range_table_func_col->colexpr);
			deparseAppendStringInfoChar(state, ' ');
		}

		if (range_table_func_col->coldefexpr)
		{
			deparseAppendStringInfoString(state, "DEFAULT ");
			deparseBExpr(state, range_table_func_col->coldefexpr);
			deparseAppendStringInfoChar(state, ' ');
		}

		if (range_table_func_col->is_not_null)
			deparseAppendStringInfoString(state, "NOT NULL ");
	}

	removeTrailingSpace(state);
}

// "table_ref" and "xmltable" in gram.y
static void deparseRangeTableFunc(DeparseState *state, RangeTableFunc* range_table_func)
{
	ListCell *lc;

	if (range_table_func->lateral)
		deparseAppendStringInfoString(state, "LATERAL ");
	
	deparseAppendStringInfoString(state, "xmltable(");
	if (range_table_func->namespaces)
	{
		deparseAppendStringInfoString(state, "xmlnamespaces(");
		deparseXmlNamespaceList(state, range_table_func->namespaces);
		deparseAppendStringInfoString(state, "), ");
	}

	deparseAppendStringInfoChar(state, '(');
	deparseExpr(state, range_table_func->rowexpr, DEPARSE_NODE_CONTEXT_NONE /* c_expr */);
	deparseAppendStringInfoChar(state, ')');

	deparseAppendStringInfoString(state, " PASSING ");
	deparseCExpr(state, range_table_func->docexpr);

	deparseAppendStringInfoString(state, " COLUMNS ");
	foreach(lc, range_table_func->columns)
	{
		deparseRangeTableFuncCol(state, castNode(RangeTableFuncCol, lfirst(lc)));
		if (lnext(range_table_func->columns, lc))
			deparseAppendStringInfoString(state, ", ");
	}

	deparseAppendStringInfoString(state, ") ");

	if (range_table_func->alias)
	{
		deparseAppendStringInfoString(state, "AS ");
		deparseAlias(state, range_table_func->alias);
	}

	removeTrailingSpace(state);
}

static void deparseXmlSerialize(DeparseState *state, XmlSerialize *xml_serialize)
{
	deparseAppendStringInfoString(state, "xmlserialize(");
	switch (xml_serialize->xmloption)
	{
		case XMLOPTION_DOCUMENT:
			deparseAppendStringInfoString(state, "document ");
			break;
		case XMLOPTION_CONTENT:
			deparseAppendStringInfoString(state, "content ");
			break;
		default:
			Assert(false);
	}
	deparseExpr(state, xml_serialize->expr, DEPARSE_NODE_CONTEXT_A_EXPR);
	deparseAppendStringInfoString(state, " AS ");
	deparseTypeName(state, xml_serialize->typeName);

	if (xml_serialize->indent) {
		deparseAppendStringInfoString(state, " INDENT");
	}

	deparseAppendStringInfoString(state, ")");
}

static void deparseJsonFormat(DeparseState *state, JsonFormat *json_format)
{
	if (json_format == NULL || json_format->format_type == JS_FORMAT_DEFAULT)
		return;

	deparseAppendStringInfoString(state, "FORMAT JSON ");

	switch (json_format->encoding)
	{
		case JS_ENC_UTF8:
			deparseAppendStringInfoString(state, "ENCODING utf8 ");
			break;
		case JS_ENC_UTF16:
			deparseAppendStringInfoString(state, "ENCODING utf16 ");
			break;
		case JS_ENC_UTF32:
			deparseAppendStringInfoString(state, "ENCODING utf32 ");
			break;
		case JS_ENC_DEFAULT:
			// no encoding specified
			break;
	}
}

static void deparseJsonIsPredicate(DeparseState *state, JsonIsPredicate *j)
{
	deparseExprOperand(state, j->expr, DEPARSE_PREC_IS, DEPARSE_OPERAND_LEFT, DEPARSE_NODE_CONTEXT_A_EXPR);
	deparseAppendStringInfoChar(state, ' ');

	deparseJsonFormat(state, castNode(JsonFormat, j->format));

	deparseAppendStringInfoString(state, "IS ");

	switch (j->item_type)
	{
		case JS_TYPE_ANY:
			deparseAppendStringInfoString(state, "JSON ");
			break;
		case JS_TYPE_ARRAY:
			deparseAppendStringInfoString(state, "JSON ARRAY ");
			break;
		case JS_TYPE_OBJECT:
			deparseAppendStringInfoString(state, "JSON OBJECT ");
			break;
		case JS_TYPE_SCALAR:
			deparseAppendStringInfoString(state, "JSON SCALAR ");
			break;
	}

	if (j->unique_keys)
		deparseAppendStringInfoString(state, "WITH UNIQUE ");

	removeTrailingSpace(state);
}

// "json_value_expr" in gram.y
static void deparseJsonValueExpr(DeparseState *state, JsonValueExpr *json_value_expr)
{
	if (json_value_expr == NULL)
		elog(ERROR, "deparse: unexpected NULL JsonValueExpr");

	deparseExpr(state, (Node *) json_value_expr->raw_expr, DEPARSE_NODE_CONTEXT_A_EXPR);
	deparseAppendStringInfoChar(state, ' ');
	deparseJsonFormat(state, json_value_expr->format);
}

// "json_value_expr_list" in gram.y
static void deparseJsonValueExprList(DeparseState *state, List *exprs)
{
	ListCell *lc;
	foreach(lc, exprs)
	{
		deparseJsonValueExpr(state, lfirst(lc));
		removeTrailingSpace(state);
		if (lnext(exprs, lc))
			deparseAppendStringInfoString(state, ", ");
	}
	deparseAppendStringInfoChar(state, ' ');
}

// "json_name_and_value" in gram.y
static void deparseJsonKeyValue(DeparseState *state, JsonKeyValue *json_key_value)
{
	deparseExpr(state, (Node *) json_key_value->key, DEPARSE_NODE_CONTEXT_A_EXPR);
	deparseAppendStringInfoString(state, ": ");
	deparseJsonValueExpr(state, json_key_value->value);
}

// "json_name_and_value_list" in gram.y
static void deparseJsonKeyValueList(DeparseState *state, List *exprs)
{
	ListCell *lc;
	foreach(lc, exprs)
	{
		deparseJsonKeyValue(state, lfirst(lc));
		removeTrailingSpace(state);
		if (lnext(exprs, lc))
			deparseAppendStringInfoString(state, ", ");
	}
	deparseAppendStringInfoChar(state, ' ');
}

static void deparseJsonOutput(DeparseState *state, JsonOutput *json_output)
{
	if (json_output == NULL)
		return;

	if (json_output->returning == NULL)
		elog(ERROR, "deparse: unexpected NULL returning in JsonOutput");

	deparseAppendStringInfoString(state, "RETURNING ");
	deparseTypeName(state, json_output->typeName);
	deparseAppendStringInfoChar(state, ' ');
	deparseJsonFormat(state, json_output->returning->format);
}

// "json_aggregate_func" and "func_expr" in gram.y
static void deparseJsonObjectAgg(DeparseState *state, JsonObjectAgg *json_object_agg)
{
	Assert(json_object_agg->constructor != NULL);

	deparseAppendStringInfoString(state, "JSON_OBJECTAGG(");
	deparseJsonKeyValue(state, json_object_agg->arg);

	if (json_object_agg->absent_on_null)
		deparseAppendStringInfoString(state, "ABSENT ON NULL ");

	if (json_object_agg->unique)
		deparseAppendStringInfoString(state, "WITH UNIQUE ");

	deparseJsonOutput(state, json_object_agg->constructor->output);

	removeTrailingSpace(state);
	deparseAppendStringInfoString(state, ") ");

	if (json_object_agg->constructor->agg_filter)
	{
		deparseAppendStringInfoString(state, "FILTER (WHERE ");
		deparseExpr(state, json_object_agg->constructor->agg_filter, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, ") ");
	}

	if (json_object_agg->constructor->over)
	{
		struct WindowDef *over = json_object_agg->constructor->over;
		deparseAppendStringInfoString(state, "OVER ");
		if (over->name)
			deparseAppendStringInfoString(state, over->name);
		else
			deparseWindowDef(state, over);
	}

	removeTrailingSpace(state);
}

// "json_aggregate_func" and "func_expr" in gram.y
static void deparseJsonArrayAgg(DeparseState *state, JsonArrayAgg *json_array_agg)
{
	Assert(json_array_agg->constructor != NULL);

	deparseAppendStringInfoString(state, "JSON_ARRAYAGG(");
	deparseJsonValueExpr(state, json_array_agg->arg);
	deparseOptSortClause(state, json_array_agg->constructor->agg_order, DEPARSE_NODE_CONTEXT_NONE);

	if (!json_array_agg->absent_on_null)
		deparseAppendStringInfoString(state, "NULL ON NULL ");

	deparseJsonOutput(state, json_array_agg->constructor->output);

	removeTrailingSpace(state);
	deparseAppendStringInfoString(state, ") ");

	if (json_array_agg->constructor->agg_filter)
	{
		deparseAppendStringInfoString(state, "FILTER (WHERE ");
		deparseExpr(state, json_array_agg->constructor->agg_filter, DEPARSE_NODE_CONTEXT_A_EXPR);
		deparseAppendStringInfoString(state, ") ");
	}

	if (json_array_agg->constructor->over)
	{
		struct WindowDef *over = json_array_agg->constructor->over;
		deparseAppendStringInfoString(state, "OVER ");
		if (over->name)
			deparseAppendStringInfoString(state, over->name);
		else
			deparseWindowDef(state, over);
	}

	removeTrailingSpace(state);
}

static void deparseJsonObjectConstructor(DeparseState *state, JsonObjectConstructor *json_object_constructor)
{
	deparseAppendStringInfoString(state, "JSON_OBJECT(");
	deparseJsonKeyValueList(state, json_object_constructor->exprs);

	if (json_object_constructor->absent_on_null)
		deparseAppendStringInfoString(state, "ABSENT ON NULL ");

	if (json_object_constructor->unique)
		deparseAppendStringInfoString(state, "WITH UNIQUE ");

	deparseJsonOutput(state, json_object_constructor->output);

	removeTrailingSpace(state);
	deparseAppendStringInfoChar(state, ')');
}

static void deparseJsonArrayConstructor(DeparseState *state, JsonArrayConstructor *json_array_constructor)
{
	deparseAppendStringInfoString(state, "JSON_ARRAY(");
	deparseJsonValueExprList(state, json_array_constructor->exprs);

	if (!json_array_constructor->absent_on_null)
		deparseAppendStringInfoString(state, "NULL ON NULL ");

	deparseJsonOutput(state, json_array_constructor->output);

	removeTrailingSpace(state);
	deparseAppendStringInfoChar(state, ')');
}

static void deparseJsonArrayQueryConstructor(DeparseState *state, JsonArrayQueryConstructor *json_array_query_constructor)
{
	deparseAppendStringInfoString(state, "JSON_ARRAY(");

	deparseSelectStmt(state, castNode(SelectStmt, json_array_query_constructor->query), DEPARSE_NODE_CONTEXT_NONE);
	deparseJsonFormat(state, json_array_query_constructor->format);
	deparseJsonOutput(state, json_array_query_constructor->output);

	removeTrailingSpace(state);
	deparseAppendStringInfoChar(state, ')');
}

static void deparseJsonParseExpr(DeparseState *state, JsonParseExpr *json_parse_expr)
{
	deparseAppendStringInfoString(state, "JSON(");

	deparseJsonValueExpr(state, json_parse_expr->expr);

	if (json_parse_expr->unique_keys)
		deparseAppendStringInfoString(state, " WITH UNIQUE KEYS");

	deparseAppendStringInfoString(state, ")");
}

static void deparseJsonScalarExpr(DeparseState *state, JsonScalarExpr *json_scalar_expr)
{
	deparseAppendStringInfoString(state, "JSON_SCALAR(");
	deparseExpr(state, (Node*) json_scalar_expr->expr, DEPARSE_NODE_CONTEXT_A_EXPR);
	deparseAppendStringInfoString(state, ")");
}

static void deparseJsonSerializeExpr(DeparseState *state, JsonSerializeExpr *json_serialize_expr)
{
	deparseAppendStringInfoString(state, "JSON_SERIALIZE(");

	deparseJsonValueExpr(state, json_serialize_expr->expr);

	if (json_serialize_expr->output)
		deparseJsonOutput(state, json_serialize_expr->output);

	deparseAppendStringInfoString(state, ")");
}

static void deparseJsonQuotesClauseOpt(DeparseState *state, JsonQuotes quotes)
{
	switch (quotes)
	{
		case JS_QUOTES_UNSPEC:
			break;
		case JS_QUOTES_KEEP:
			deparseAppendStringInfoString(state, " KEEP QUOTES");
			break;
		case JS_QUOTES_OMIT:
			deparseAppendStringInfoString(state, " OMIT QUOTES");
			break;
	}
}

static void deparseJsonOnErrorClauseOpt(DeparseState *state, JsonBehavior *behavior)
{
	if (!behavior)
		return;

	deparseAppendStringInfoChar(state, ' ');
	deparseJsonBehavior(state, behavior);
	deparseAppendStringInfoString(state, " ON ERROR");
}

static void deparseJsonOnEmptyClauseOpt(DeparseState *state, JsonBehavior *behavior)
{
	if (behavior)
	{
		deparseAppendStringInfoChar(state, ' ');
		deparseJsonBehavior(state, behavior);
		deparseAppendStringInfoString(state, " ON EMPTY");
	}
}

static void deparseJsonFuncExpr(DeparseState *state, JsonFuncExpr *json_func_expr)
{
	switch (json_func_expr->op)
	{
		case JSON_EXISTS_OP:
			deparseAppendStringInfoString(state, "JSON_EXISTS(");
			break;
		case JSON_QUERY_OP:
			deparseAppendStringInfoString(state, "JSON_QUERY(");
			break;
		case JSON_VALUE_OP:
			deparseAppendStringInfoString(state, "JSON_VALUE(");
			break;
		case JSON_TABLE_OP:
			deparseAppendStringInfoString(state, "JSON_TABLE(");
			break;
	}

	deparseJsonValueExpr(state, json_func_expr->context_item);
	deparseAppendStringInfoString(state, ", ");
	deparseExpr(state, json_func_expr->pathspec, DEPARSE_NODE_CONTEXT_A_EXPR);

	if (json_func_expr->passing)
		deparseAppendStringInfoString(state, " PASSING ");

	ListCell *lc = NULL;
	foreach (lc, json_func_expr->passing)
	{
		JsonArgument *json_argument = castNode(JsonArgument, lfirst(lc));
		deparseJsonValueExpr(state, json_argument->val);
		deparseAppendStringInfoString(state, " AS ");
		deparseColLabel(state, json_argument->name);

		if (lnext(json_func_expr->passing, lc))
			deparseAppendStringInfoString(state, ", ");
	}

	if (json_func_expr->output)
	{
		deparseAppendStringInfoChar(state, ' ');
		deparseJsonOutput(state, json_func_expr->output);
	}

	switch (json_func_expr->wrapper)
	{
		case JSW_UNSPEC:
			break;
		case JSW_NONE:
			deparseAppendStringInfoString(state, " WITHOUT WRAPPER");
			break;
		case JSW_CONDITIONAL:
			deparseAppendStringInfoString(state, " WITH CONDITIONAL WRAPPER");
			break;
		case JSW_UNCONDITIONAL:
			deparseAppendStringInfoString(state, " WITH UNCONDITIONAL WRAPPER");
			break;
	}

	deparseJsonQuotesClauseOpt(state, json_func_expr->quotes);
	deparseJsonOnEmptyClauseOpt(state, json_func_expr->on_empty);
	deparseJsonOnErrorClauseOpt(state, json_func_expr->on_error);

	deparseAppendStringInfoChar(state, ')');
}

static void deparseJsonTablePathSpec(DeparseState *state, JsonTablePathSpec *json_table_path_spec)
{
	deparseStringLiteral(state, castNode(A_Const, json_table_path_spec->string)->val.sval.sval);

	if (json_table_path_spec->name)
	{
		deparseAppendStringInfoString(state, " AS ");
		deparseColLabel(state, json_table_path_spec->name);
	}
}

// "json_behavior" in gram.y
static void deparseJsonBehavior(DeparseState *state, JsonBehavior *json_behavior)
{
	switch (json_behavior->btype)
	{
		case JSON_BEHAVIOR_NULL:
			deparseAppendStringInfoString(state, "NULL");
			break;
		case JSON_BEHAVIOR_ERROR:
			deparseAppendStringInfoString(state, "ERROR");
			break;
		case JSON_BEHAVIOR_EMPTY:
			deparseAppendStringInfoString(state, "EMPTY");
			break;
		case JSON_BEHAVIOR_TRUE:
			deparseAppendStringInfoString(state, "TRUE");
			break;
		case JSON_BEHAVIOR_FALSE:
			deparseAppendStringInfoString(state, "FALSE");
			break;
		case JSON_BEHAVIOR_EMPTY_ARRAY:
			deparseAppendStringInfoString(state, "EMPTY ARRAY");
			break;
		case JSON_BEHAVIOR_EMPTY_OBJECT:
			deparseAppendStringInfoString(state, "EMPTY OBJECT");
			break;
		case JSON_BEHAVIOR_DEFAULT:
			deparseAppendStringInfoString(state, "DEFAULT ");
			deparseExpr(state, (Node*) json_behavior->expr, DEPARSE_NODE_CONTEXT_A_EXPR);
			break;
		case JSON_BEHAVIOR_UNKNOWN:
			deparseAppendStringInfoString(state, "UNKNOWN");
			break;
	}
}

static void deparseJsonTableColumn(DeparseState *state, JsonTableColumn *json_table_column)
{
	if (json_table_column->coltype == JTC_NESTED)
	{
		deparseAppendStringInfoString(state, "NESTED PATH ");
		deparseJsonTablePathSpec(state, json_table_column->pathspec);
		deparseJsonTableColumns(state, json_table_column->columns);
		return;
	}

	deparseColLabel(state, json_table_column->name);
	deparseAppendStringInfoChar(state, ' ');

	switch (json_table_column->coltype)
	{
		case JTC_FOR_ORDINALITY:
			deparseAppendStringInfoString(state, " FOR ORDINALITY");
			break;
		case JTC_EXISTS:
		case JTC_FORMATTED:
		case JTC_REGULAR:
			deparseTypeName(state, json_table_column->typeName);

			if (json_table_column->coltype == JTC_EXISTS)
				deparseAppendStringInfoString(state, " EXISTS ");
			else
				deparseAppendStringInfoChar(state, ' ');

			if (json_table_column->format)
				deparseJsonFormat(state, json_table_column->format);

			if (json_table_column->pathspec)
			{
				deparseAppendStringInfoString(state, "PATH ");
				deparseJsonTablePathSpec(state, json_table_column->pathspec);
			}
			break;
		case JTC_NESTED:
			Assert(false);
	}

	switch (json_table_column->wrapper)
	{
		case JSW_UNSPEC:
			break;
		case JSW_NONE:
			if (json_table_column->coltype == JTC_REGULAR || json_table_column->coltype == JTC_FORMATTED)
				deparseAppendStringInfoString(state, " WITHOUT WRAPPER");
			break;
		case JSW_CONDITIONAL:
			deparseAppendStringInfoString(state, " WITH CONDITIONAL WRAPPER");
			break;
		case JSW_UNCONDITIONAL:
			deparseAppendStringInfoString(state, " WITH UNCONDITIONAL WRAPPER");
			break;
	}

	deparseJsonQuotesClauseOpt(state, json_table_column->quotes);
	deparseJsonOnEmptyClauseOpt(state, json_table_column->on_empty);
	deparseJsonOnErrorClauseOpt(state, json_table_column->on_error);
}

static void deparseJsonTableColumns(DeparseState *state, List *json_table_columns)
{
	deparseAppendStringInfoString(state, " COLUMNS (");

	ListCell *lc = NULL;
	foreach(lc, json_table_columns)
	{
		deparseJsonTableColumn(state, castNode(JsonTableColumn, lfirst(lc)));

		if (lnext(json_table_columns, lc))
			deparseAppendStringInfoString(state, ", ");
	}

	deparseAppendStringInfoChar(state, ')');
}

static void deparseJsonTable(DeparseState *state, JsonTable *json_table)
{
	deparseAppendStringInfoString(state, "JSON_TABLE(");

	deparseJsonValueExpr(state, json_table->context_item);
	deparseAppendStringInfoString(state, ", ");
	deparseJsonTablePathSpec(state, json_table->pathspec);

	if (json_table->passing)
		deparseAppendStringInfoString(state, " PASSING ");

	ListCell *lc = NULL;
	foreach (lc, json_table->passing)
	{
		JsonArgument *json_argument = castNode(JsonArgument, lfirst(lc));
		deparseJsonValueExpr(state, json_argument->val);
		deparseAppendStringInfoString(state, " AS ");
		deparseColLabel(state, json_argument->name);

		if (lnext(json_table->passing, lc))
			deparseAppendStringInfoString(state, ", ");
	}

	deparseJsonTableColumns(state, json_table->columns);

	if (json_table->on_error)
	{
		deparseJsonBehavior(state, json_table->on_error);
		deparseAppendStringInfoString(state, " ON ERROR");
	}

	deparseAppendStringInfoChar(state, ')');

	if (json_table->alias)
	{
		deparseAppendStringInfoChar(state, ' ');
		deparseAlias(state, json_table->alias);
	}
}

static void deparseGroupingFunc(DeparseState *state, GroupingFunc *grouping_func)
{
	deparseAppendStringInfoString(state, "GROUPING(");
	deparseExprList(state, grouping_func->args);
	deparseAppendStringInfoChar(state, ')');
}

static void deparseClusterStmt(DeparseState *state, ClusterStmt *cluster_stmt)
{
	deparseAppendStringInfoString(state, "CLUSTER ");

        deparseUtilityOptionList(state, cluster_stmt->params);

	if (cluster_stmt->relation != NULL)
	{
		deparseRangeVar(state, cluster_stmt->relation, DEPARSE_NODE_CONTEXT_NONE);
		deparseAppendStringInfoChar(state, ' ');
	}

	if (cluster_stmt->indexname != NULL)
	{
		deparseAppendStringInfoString(state, "USING ");
		deparseAppendStringInfoString(state, quote_identifier(cluster_stmt->indexname));
		deparseAppendStringInfoChar(state, ' ');
	}

	removeTrailingSpace(state);
}

static void deparseValue(DeparseState *state, union ValUnion *value, DeparseNodeContext context)
{
	if (!value) {
		deparseAppendStringInfoString(state, "NULL");
		return;
	}

	switch (nodeTag(value))
	{
		case T_Integer:
		case T_Float:
			deparseNumericOnly(state, value);
			break;
		case T_Boolean:
			deparseAppendStringInfoString(state, value->boolval.boolval ? "true" : "false");
			break;
		case T_String:
			if (context == DEPARSE_NODE_CONTEXT_IDENTIFIER) {
				deparseAppendStringInfoString(state, quote_identifier(value->sval.sval));
			} else if (context == DEPARSE_NODE_CONTEXT_CONSTANT) {
				deparseStringLiteral(state, value->sval.sval);
			} else {
				deparseAppendStringInfoString(state, value->sval.sval);
			}
			break;
		case T_BitString:
			if (strlen(value->sval.sval) >= 1 && value->sval.sval[0] == 'x')
			{
				deparseAppendStringInfoChar(state, 'x');
				deparseStringLiteral(state, value->sval.sval + 1);
			}
			else if (strlen(value->sval.sval) >= 1 && value->sval.sval[0] == 'b')
			{
				deparseAppendStringInfoChar(state, 'b');
				deparseStringLiteral(state, value->sval.sval + 1);
			}
			else
			{
				Assert(false);
			}
			break;
		default:
			elog(ERROR, "deparse: unrecognized value node type: %d",
				 (int) nodeTag(value));
			break;
	}
}

// "PrepareableStmt" in gram.y
static void deparsePreparableStmt(DeparseState *state, Node *node)
{
	if (node == NULL)
		elog(ERROR, "deparse: unexpected NULL node");

	switch (nodeTag(node))
	{
		case T_SelectStmt:
			deparseSelectStmt(state, castNode(SelectStmt, node), DEPARSE_NODE_CONTEXT_NONE);
			break;
		case T_InsertStmt:
			deparseInsertStmt(state, castNode(InsertStmt, node));
			break;
		case T_UpdateStmt:
			deparseUpdateStmt(state, castNode(UpdateStmt, node));
			break;
		case T_DeleteStmt:
			deparseDeleteStmt(state, castNode(DeleteStmt, node));
			break;
		case T_MergeStmt:
			deparseMergeStmt(state, castNode(MergeStmt, node));
			break;
		default:
			Assert(false);
	}
}

// "RuleActionStmt" in gram.y
static void deparseRuleActionStmt(DeparseState *state, Node *node)
{
	if (node == NULL)
		elog(ERROR, "deparse: unexpected NULL node");

	switch (nodeTag(node))
	{
		case T_SelectStmt:
			deparseSelectStmt(state, castNode(SelectStmt, node), DEPARSE_NODE_CONTEXT_NONE);
			break;
		case T_InsertStmt:
			deparseInsertStmt(state, castNode(InsertStmt, node));
			break;
		case T_UpdateStmt:
			deparseUpdateStmt(state, castNode(UpdateStmt, node));
			break;
		case T_DeleteStmt:
			deparseDeleteStmt(state, castNode(DeleteStmt, node));
			break;
		case T_NotifyStmt:
			deparseNotifyStmt(state, castNode(NotifyStmt, node));
			break;
		default:
			Assert(false);
	}
}

// "ExplainableStmt" in gram.y
static void deparseExplainableStmt(DeparseState *state, Node *node)
{
	if (node == NULL)
		elog(ERROR, "deparse: unexpected NULL node");

	switch (nodeTag(node))
	{
		case T_SelectStmt:
			deparseSelectStmt(state, castNode(SelectStmt, node), DEPARSE_NODE_CONTEXT_NONE);
			break;
		case T_InsertStmt:
			deparseInsertStmt(state, castNode(InsertStmt, node));
			break;
		case T_UpdateStmt:
			deparseUpdateStmt(state, castNode(UpdateStmt, node));
			break;
		case T_DeleteStmt:
			deparseDeleteStmt(state, castNode(DeleteStmt, node));
			break;
		case T_DeclareCursorStmt:
			deparseDeclareCursorStmt(state, castNode(DeclareCursorStmt, node));
			break;
		case T_CreateTableAsStmt:
			deparseCreateTableAsStmt(state, castNode(CreateTableAsStmt, node));
			break;
		case T_RefreshMatViewStmt:
			deparseRefreshMatViewStmt(state, castNode(RefreshMatViewStmt, node));
			break;
		case T_ExecuteStmt:
			deparseExecuteStmt(state, castNode(ExecuteStmt, node));
			break;
		case T_MergeStmt:
			deparseMergeStmt(state, castNode(MergeStmt, node));
			break;
		default:
			Assert(false);
	}
}

// "schema_stmt" in gram.y
static void deparseSchemaStmt(DeparseState *state, Node *node)
{
	if (node == NULL)
		elog(ERROR, "deparse: unexpected NULL node");

	switch (nodeTag(node))
	{
		case T_CreateStmt:
			deparseCreateStmt(state, castNode(CreateStmt, node), false);
			break;
		case T_IndexStmt:
			deparseIndexStmt(state, castNode(IndexStmt, node));
			break;
		case T_CreateSeqStmt:
			deparseCreateSeqStmt(state, castNode(CreateSeqStmt, node));
			break;
		case T_CreateTrigStmt:
			deparseCreateTrigStmt(state, castNode(CreateTrigStmt, node));
			break;
		case T_GrantStmt:
			deparseGrantStmt(state, castNode(GrantStmt, node));
			break;
		case T_ViewStmt:
			deparseViewStmt(state, castNode(ViewStmt, node));
			break;
		default:
			Assert(false);
	}
}

// "stmt" in gram.y
static void deparseStmt(DeparseState *state, Node *node)
{
	DeparseStateNestingLevel *parent_level = NULL;

	// For statements that can be nested, push/pop is handled directly in the
	// respective deparse...Stmt methods
	bool skip_push_pop = IsA(node, SelectStmt) ||
		IsA(node, InsertStmt) ||
		IsA(node, UpdateStmt) ||
		IsA(node, DeleteStmt) ||
		IsA(node, MergeStmt);

	check_stack_depth();

	if (!skip_push_pop)
		parent_level = deparseStateIncreaseNestingLevel(state);

	// Note the following grammar names are missing in the list, because they
	// get mapped to other node types:
	//
	// - AlterForeignTableStmt (=> AlterTableStmt)
	// - AlterGroupStmt (=> AlterRoleStmt)
	// - AlterCompositeTypeStmt (=> AlterTableStmt)
	// - AnalyzeStmt (=> VacuumStmt)
	// - CreateGroupStmt (=> CreateRoleStmt)
	// - CreateMatViewStmt (=> CreateTableAsStmt)
	// - CreateUserStmt (=> CreateRoleStmt)
	// - DropCastStmt (=> DropStmt)
	// - DropOpClassStmt (=> DropStmt)
	// - DropOpFamilyStmt (=> DropStmt)
	// - DropPLangStmt (=> DropPLangStmt)
	// - DropTransformStmt (=> DropStmt)
	// - RemoveAggrStmt (=> DropStmt)
	// - RemoveFuncStmt (=> DropStmt)
	// - RemoveOperStmt (=> DropStmt)
	// - RevokeStmt (=> GrantStmt)
	// - RevokeRoleStmt (=> GrantRoleStmt)
	// - VariableResetStmt (=> VariableSetStmt)
	//
	// And the following grammar names error out in the parser:
	// - CreateAssertionStmt (not supported yet)
	if (node == NULL)
		elog(ERROR, "deparse: unexpected NULL node");

	switch (nodeTag(node))
	{
		case T_AlterEventTrigStmt:
			deparseAlterEventTrigStmt(state, castNode(AlterEventTrigStmt, node));
			break;
		case T_AlterCollationStmt:
			deparseAlterCollationStmt(state, castNode(AlterCollationStmt, node));
			break;
		case T_AlterDatabaseStmt:
			deparseAlterDatabaseStmt(state, castNode(AlterDatabaseStmt, node));
			break;
		case T_AlterDatabaseSetStmt:
			deparseAlterDatabaseSetStmt(state, castNode(AlterDatabaseSetStmt, node));
			break;
		case T_AlterDefaultPrivilegesStmt:
			deparseAlterDefaultPrivilegesStmt(state, castNode(AlterDefaultPrivilegesStmt, node));
			break;
		case T_AlterDomainStmt:
			deparseAlterDomainStmt(state, castNode(AlterDomainStmt, node));
			break;
		case T_AlterEnumStmt:
			deparseAlterEnumStmt(state, castNode(AlterEnumStmt, node));
			break;
		case T_AlterExtensionStmt:
			deparseAlterExtensionStmt(state, castNode(AlterExtensionStmt, node));
			break;
		case T_AlterExtensionContentsStmt:
			deparseAlterExtensionContentsStmt(state, castNode(AlterExtensionContentsStmt, node));
			break;
		case T_AlterFdwStmt:
			deparseAlterFdwStmt(state, castNode(AlterFdwStmt, node));
			break;
		case T_AlterForeignServerStmt:
			deparseAlterForeignServerStmt(state, castNode(AlterForeignServerStmt, node));
			break;
		case T_AlterFunctionStmt:
			deparseAlterFunctionStmt(state, castNode(AlterFunctionStmt, node));
			break;
		case T_AlterObjectDependsStmt:
			deparseAlterObjectDependsStmt(state, castNode(AlterObjectDependsStmt, node));
			break;
		case T_AlterObjectSchemaStmt:
			deparseAlterObjectSchemaStmt(state, castNode(AlterObjectSchemaStmt, node));
			break;
		case T_AlterOwnerStmt:
			deparseAlterOwnerStmt(state, castNode(AlterOwnerStmt, node));
			break;
		case T_AlterOperatorStmt:
			deparseAlterOperatorStmt(state, castNode(AlterOperatorStmt, node));
			break;
		case T_AlterTypeStmt:
			deparseAlterTypeStmt(state, castNode(AlterTypeStmt, node));
			break;
		case T_AlterPolicyStmt:
			deparseAlterPolicyStmt(state, castNode(AlterPolicyStmt, node));
			break;
		case T_AlterSeqStmt:
			deparseAlterSeqStmt(state, castNode(AlterSeqStmt, node));
			break;
		case T_AlterSystemStmt:
			deparseAlterSystemStmt(state, castNode(AlterSystemStmt, node));
			break;
		case T_AlterTableMoveAllStmt:
			deparseAlterTableMoveAllStmt(state, castNode(AlterTableMoveAllStmt, node));
			break;
		case T_AlterTableStmt:
			deparseAlterTableStmt(state, castNode(AlterTableStmt, node));
			break;
		case T_AlterTableSpaceOptionsStmt: // "AlterTblSpcStmt" in gram.y
			deparseAlterTableSpaceOptionsStmt(state, castNode(AlterTableSpaceOptionsStmt, node));
			break;
		case T_AlterPublicationStmt:
			deparseAlterPublicationStmt(state, castNode(AlterPublicationStmt, node));
			break;
		case T_AlterRoleSetStmt:
			deparseAlterRoleSetStmt(state, castNode(AlterRoleSetStmt, node));
			break;
		case T_AlterRoleStmt:
			deparseAlterRoleStmt(state, castNode(AlterRoleStmt, node));
			break;
		case T_AlterSubscriptionStmt:
			deparseAlterSubscriptionStmt(state, castNode(AlterSubscriptionStmt, node));
			break;
		case T_AlterStatsStmt:
			deparseAlterStatsStmt(state, castNode(AlterStatsStmt, node));
			break;
		case T_AlterTSConfigurationStmt:
			deparseAlterTSConfigurationStmt(state, castNode(AlterTSConfigurationStmt, node));
			break;
		case T_AlterTSDictionaryStmt:
			deparseAlterTSDictionaryStmt(state, castNode(AlterTSDictionaryStmt, node));
			break;
		case T_AlterUserMappingStmt:
			deparseAlterUserMappingStmt(state, castNode(AlterUserMappingStmt, node));
			break;
		case T_CallStmt:
			deparseCallStmt(state, castNode(CallStmt, node));
			break;
		case T_CheckPointStmt:
			deparseCheckPointStmt(state, castNode(CheckPointStmt, node));
			break;
		case T_ClosePortalStmt:
			deparseClosePortalStmt(state, castNode(ClosePortalStmt, node));
			break;
		case T_ClusterStmt:
			deparseClusterStmt(state, castNode(ClusterStmt, node));
			break;
		case T_CommentStmt:
			deparseCommentStmt(state, castNode(CommentStmt, node));
			break;
		case T_ConstraintsSetStmt:
			deparseConstraintsSetStmt(state, castNode(ConstraintsSetStmt, node));
			break;
		case T_CopyStmt:
			deparseCopyStmt(state, castNode(CopyStmt, node));
			break;
		case T_CreateAmStmt:
			deparseCreateAmStmt(state, castNode(CreateAmStmt, node));
			break;
		case T_CreateTableAsStmt: // "CreateAsStmt" in gram.y
			deparseCreateTableAsStmt(state, castNode(CreateTableAsStmt, node));
			break;
		case T_CreateCastStmt:
			deparseCreateCastStmt(state, castNode(CreateCastStmt, node));
			break;
		case T_CreateConversionStmt:
			deparseCreateConversionStmt(state, castNode(CreateConversionStmt, node));
			break;
		case T_CreateDomainStmt:
			deparseCreateDomainStmt(state, castNode(CreateDomainStmt, node));
			break;
		case T_CreateExtensionStmt:
			deparseCreateExtensionStmt(state, castNode(CreateExtensionStmt, node));
			break;
		case T_CreateFdwStmt:
			deparseCreateFdwStmt(state, castNode(CreateFdwStmt, node));
			break;
		case T_CreateForeignServerStmt:
			deparseCreateForeignServerStmt(state, castNode(CreateForeignServerStmt, node));
			break;
		case T_CreateForeignTableStmt:
			deparseCreateForeignTableStmt(state, castNode(CreateForeignTableStmt, node));
			break;
		case T_CreateFunctionStmt:
			deparseCreateFunctionStmt(state, castNode(CreateFunctionStmt, node));
			break;
		case T_CreateOpClassStmt:
			deparseCreateOpClassStmt(state, castNode(CreateOpClassStmt, node));
			break;
		case T_CreateOpFamilyStmt:
			deparseCreateOpFamilyStmt(state, castNode(CreateOpFamilyStmt, node));
			break;
		case T_CreatePublicationStmt:
			deparseCreatePublicationStmt(state, castNode(CreatePublicationStmt, node));
			break;
		case T_AlterOpFamilyStmt:
			deparseAlterOpFamilyStmt(state, castNode(AlterOpFamilyStmt, node));
			break;
		case T_CreatePolicyStmt:
			deparseCreatePolicyStmt(state, castNode(CreatePolicyStmt, node));
			break;
		case T_CreatePLangStmt:
			deparseCreatePLangStmt(state, castNode(CreatePLangStmt, node));
			break;
		case T_CreateSchemaStmt:
			deparseCreateSchemaStmt(state, castNode(CreateSchemaStmt, node));
			break;
		case T_CreateSeqStmt:
			deparseCreateSeqStmt(state, castNode(CreateSeqStmt, node));
			break;
		case T_CreateStmt:
			deparseCreateStmt(state, castNode(CreateStmt, node), false);
			break;
		case T_CreateSubscriptionStmt:
			deparseCreateSubscriptionStmt(state, castNode(CreateSubscriptionStmt, node));
			break;
		case T_CreateStatsStmt:
			deparseCreateStatsStmt(state, castNode(CreateStatsStmt, node));
			break;
		case T_CreateTableSpaceStmt:
			deparseCreateTableSpaceStmt(state, castNode(CreateTableSpaceStmt, node));
			break;
		case T_CreateTransformStmt:
			deparseCreateTransformStmt(state, castNode(CreateTransformStmt, node));
			break;
		case T_CreateTrigStmt:
			deparseCreateTrigStmt(state, castNode(CreateTrigStmt, node));
			break;
		case T_CreateEventTrigStmt:
			deparseCreateEventTrigStmt(state, castNode(CreateEventTrigStmt, node));
			break;
		case T_CreateRoleStmt:
			deparseCreateRoleStmt(state, castNode(CreateRoleStmt, node));
			break;
		case T_CreateUserMappingStmt:
			deparseCreateUserMappingStmt(state, castNode(CreateUserMappingStmt, node));
			break;
		case T_CreatedbStmt:
			deparseCreatedbStmt(state, castNode(CreatedbStmt, node));
			break;
		case T_DeallocateStmt:
			deparseDeallocateStmt(state, castNode(DeallocateStmt, node));
			break;
		case T_DeclareCursorStmt:
			deparseDeclareCursorStmt(state, castNode(DeclareCursorStmt, node));
			break;
		case T_DefineStmt:
			deparseDefineStmt(state, castNode(DefineStmt, node));
			break;
		case T_DeleteStmt:
			deparseDeleteStmt(state, castNode(DeleteStmt, node));
			break;
		case T_DiscardStmt:
			deparseDiscardStmt(state, castNode(DiscardStmt, node));
			break;
		case T_DoStmt:
			deparseDoStmt(state, castNode(DoStmt, node));
			break;
		case T_DropOwnedStmt:
			deparseDropOwnedStmt(state, castNode(DropOwnedStmt, node));
			break;
		case T_DropStmt:
			deparseDropStmt(state, castNode(DropStmt, node));
			break;
		case T_DropSubscriptionStmt:
			deparseDropSubscriptionStmt(state, castNode(DropSubscriptionStmt, node));
			break;
		case T_DropTableSpaceStmt:
			deparseDropTableSpaceStmt(state, castNode(DropTableSpaceStmt, node));
			break;
		case T_DropRoleStmt:
			deparseDropRoleStmt(state, castNode(DropRoleStmt, node));
			break;
		case T_DropUserMappingStmt:
			deparseDropUserMappingStmt(state, castNode(DropUserMappingStmt, node));
			break;
		case T_DropdbStmt:
			deparseDropdbStmt(state, castNode(DropdbStmt, node));
			break;
		case T_ExecuteStmt:
			deparseExecuteStmt(state, castNode(ExecuteStmt, node));
			break;
		case T_ExplainStmt:
			deparseExplainStmt(state, castNode(ExplainStmt, node));
			break;
		case T_FetchStmt:
			deparseFetchStmt(state, castNode(FetchStmt, node));
			break;
		case T_GrantStmt:
			deparseGrantStmt(state, castNode(GrantStmt, node));
			break;
		case T_GrantRoleStmt:
			deparseGrantRoleStmt(state, castNode(GrantRoleStmt, node));
			break;
		case T_ImportForeignSchemaStmt:
			deparseImportForeignSchemaStmt(state, castNode(ImportForeignSchemaStmt, node));
			break;
		case T_IndexStmt:
			deparseIndexStmt(state, castNode(IndexStmt, node));
			break;
		case T_InsertStmt:
			deparseInsertStmt(state, castNode(InsertStmt, node));
			break;
		case T_ListenStmt:
			deparseListenStmt(state, castNode(ListenStmt, node));
			break;
		case T_RefreshMatViewStmt:
			deparseRefreshMatViewStmt(state, castNode(RefreshMatViewStmt, node));
			break;
		case T_LoadStmt:
			deparseLoadStmt(state, castNode(LoadStmt, node));
			break;
		case T_LockStmt:
			deparseLockStmt(state, castNode(LockStmt, node));
			break;
		case T_MergeStmt:
			deparseMergeStmt(state, castNode(MergeStmt, node));
			break;
		case T_NotifyStmt:
			deparseNotifyStmt(state, castNode(NotifyStmt, node));
			break;
		case T_PrepareStmt:
			deparsePrepareStmt(state, castNode(PrepareStmt, node));
			break;
		case T_ReassignOwnedStmt:
			deparseReassignOwnedStmt(state, castNode(ReassignOwnedStmt, node));
			break;
		case T_ReindexStmt:
			deparseReindexStmt(state, castNode(ReindexStmt, node));
			break;
		case T_RenameStmt:
			deparseRenameStmt(state, castNode(RenameStmt, node));
			break;
		case T_RuleStmt:
			deparseRuleStmt(state, castNode(RuleStmt, node));
			break;
		case T_SecLabelStmt:
			deparseSecLabelStmt(state, castNode(SecLabelStmt, node));
			break;
		case T_SelectStmt:
			deparseSelectStmt(state, castNode(SelectStmt, node), DEPARSE_NODE_CONTEXT_NONE);
			break;
		case T_TransactionStmt:
			deparseTransactionStmt(state, castNode(TransactionStmt, node));
			break;
		case T_TruncateStmt:
			deparseTruncateStmt(state, castNode(TruncateStmt, node));
			break;
		case T_UnlistenStmt:
			deparseUnlistenStmt(state, castNode(UnlistenStmt, node));
			break;
		case T_UpdateStmt:
			deparseUpdateStmt(state, castNode(UpdateStmt, node));
			break;
		case T_VacuumStmt:
			deparseVacuumStmt(state, castNode(VacuumStmt, node));
			break;
		case T_VariableSetStmt:
			deparseVariableSetStmt(state, castNode(VariableSetStmt, node));
			break;
		case T_VariableShowStmt:
			deparseVariableShowStmt(state, castNode(VariableShowStmt, node));
			break;
		case T_ViewStmt:
			deparseViewStmt(state, castNode(ViewStmt, node));
			break;
		// These node types are created by DefineStmt grammar for CREATE TYPE in some cases
		case T_CompositeTypeStmt:
			deparseCompositeTypeStmt(state, castNode(CompositeTypeStmt, node));
			break;
		case T_CreateEnumStmt:
			deparseCreateEnumStmt(state, castNode(CreateEnumStmt, node));
			break;
		case T_CreateRangeStmt:
			deparseCreateRangeStmt(state, castNode(CreateRangeStmt, node));
			break;
		default:
			elog(ERROR, "deparse: unsupported top-level node type: %u", nodeTag(node));
	}

	if (!skip_push_pop)
		deparseStateDecreaseNestingLevel(state, parent_level);
}
