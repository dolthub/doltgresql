/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - assign_expr_collations
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * parse_collate.c
 *		Routines for assigning collation information.
 *
 * We choose to handle collation analysis in a post-pass over the output
 * of expression parse analysis.  This is because we need more state to
 * perform this processing than is needed in the finished tree.  If we
 * did it on-the-fly while building the tree, all that state would have
 * to be kept in expression node trees permanently.  This way, the extra
 * storage is just local variables in this recursive routine.
 *
 * The info that is actually saved in the finished tree is:
 * 1. The output collation of each expression node, or InvalidOid if it
 * returns a noncollatable data type.  This can also be InvalidOid if the
 * result type is collatable but the collation is indeterminate.
 * 2. The collation to be used in executing each function.  InvalidOid means
 * that there are no collatable inputs or their collation is indeterminate.
 * This value is only stored in node types that might call collation-using
 * functions.
 *
 * You might think we could get away with storing only one collation per
 * node, but the two concepts really need to be kept distinct.  Otherwise
 * it's too confusing when a function produces a collatable output type but
 * has no collatable inputs or produces noncollatable output from collatable
 * inputs.
 *
 * Cases with indeterminate collation might result in an error being thrown
 * at runtime.  If we knew exactly which functions require collation
 * information, we could throw those errors at parse time instead.
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 *
 * IDENTIFICATION
 *	  src/backend/parser/parse_collate.c
 *
 *-------------------------------------------------------------------------
 */
#include "postgres.h"

#include "catalog/pg_aggregate.h"
#include "catalog/pg_collation.h"
#include "nodes/makefuncs.h"
#include "nodes/nodeFuncs.h"
#include "parser/parse_collate.h"
#include "utils/lsyscache.h"


/*
 * Collation strength (the SQL standard calls this "derivation").  Order is
 * chosen to allow comparisons to work usefully.  Note: the standard doesn't
 * seem to distinguish between NONE and CONFLICT.
 */
typedef enum
{
	COLLATE_NONE,				/* expression is of a noncollatable datatype */
	COLLATE_IMPLICIT,			/* collation was derived implicitly */
	COLLATE_CONFLICT,			/* we had a conflict of implicit collations */
	COLLATE_EXPLICIT,			/* collation was derived explicitly */
} CollateStrength;

typedef struct
{
	ParseState *pstate;			/* parse state (for error reporting) */
	Oid			collation;		/* OID of current collation, if any */
	CollateStrength strength;	/* strength of current collation choice */
	int			location;		/* location of expr that set collation */
	/* Remaining fields are only valid when strength == COLLATE_CONFLICT */
	Oid			collation2;		/* OID of conflicting collation */
	int			location2;		/* location of expr that set collation2 */
} assign_collations_context;

static bool assign_query_collations_walker(Node *node, ParseState *pstate);
static bool assign_collations_walker(Node *node,
									 assign_collations_context *context);
static void merge_collation_state(Oid collation,
								  CollateStrength strength,
								  int location,
								  Oid collation2,
								  int location2,
								  assign_collations_context *context);
static void assign_aggregate_collations(Aggref *aggref,
										assign_collations_context *loccontext);
static void assign_ordered_set_collations(Aggref *aggref,
										  assign_collations_context *loccontext);
static void assign_hypothetical_collations(Aggref *aggref,
										   assign_collations_context *loccontext);


/*
 * assign_query_collations()
 *		Mark all expressions in the given Query with collation information.
 *
 * This should be applied to each Query after completion of parse analysis
 * for expressions.  Note that we do not recurse into sub-Queries, since
 * those should have been processed when built.
 */


/*
 * Walker for assign_query_collations
 *
 * Each expression found by query_tree_walker is processed independently.
 * Note that query_tree_walker may pass us a whole List, such as the
 * targetlist, in which case each subexpression must be processed
 * independently --- we don't want to bleat if two different targetentries
 * have different collations.
 */


/*
 * assign_list_collations()
 *		Mark all nodes in the list of expressions with collation information.
 *
 * The list member expressions are processed independently; they do not have
 * to share a common collation.
 */


/*
 * assign_expr_collations()
 *		Mark all nodes in the given expression tree with collation information.
 *
 * This is exported for the benefit of various utility commands that process
 * expressions without building a complete Query.  It should be applied after
 * calling transformExpr() plus any expression-modifying operations such as
 * coerce_to_boolean().
 */

void
assign_expr_collations(ParseState *pstate, Node *expr)
{
	/* Do nothing */
}

/*
 * select_common_collation()
 *		Identify a common collation for a list of expressions.
 *
 * The expressions should all return the same datatype, else this is not
 * terribly meaningful.
 *
 * none_ok means that it is permitted to return InvalidOid, indicating that
 * no common collation could be identified, even for collatable datatypes.
 * Otherwise, an error is thrown for conflict of implicit collations.
 *
 * In theory, none_ok = true reflects the rules of SQL standard clause "Result
 * of data type combinations", none_ok = false reflects the rules of clause
 * "Collation determination" (in some cases invoked via "Grouping
 * operations").
 */


/*
 * assign_collations_walker()
 *		Recursive guts of collation processing.
 *
 * Nodes with no children (eg, Vars, Consts, Params) must have been marked
 * when built.  All upper-level nodes are marked here.
 *
 * Note: if this is invoked directly on a List, it will attempt to infer a
 * common collation for all the list members.  In particular, it will throw
 * error if there are conflicting explicit collations for different members.
 */


/*
 * Merge collation state of a subexpression into the context for its parent.
 */


/*
 * Aggref is a special case because expressions used only for ordering
 * shouldn't be taken to conflict with each other or with regular args,
 * indeed shouldn't affect the aggregate's result collation at all.
 * We handle this by applying assign_expr_collations() to them rather than
 * passing down our loccontext.
 *
 * Note that we recurse to each TargetEntry, not directly to its contained
 * expression, so that the case above for T_TargetEntry will complain if we
 * can't resolve a collation for an ORDER BY item (whether or not it is also
 * a normal aggregate arg).
 *
 * We need not recurse into the aggorder or aggdistinct lists, because those
 * contain only SortGroupClause nodes which we need not process.
 */


/*
 * For ordered-set aggregates, it's somewhat unclear how best to proceed.
 * The spec-defined inverse distribution functions have only one sort column
 * and don't return collatable types, but this is clearly too restrictive in
 * the general case.  Our solution is to consider that the aggregate's direct
 * arguments contribute normally to determination of the aggregate's own
 * collation, while aggregated arguments contribute only when the aggregate
 * is designed to have exactly one aggregated argument (i.e., it has a single
 * aggregated argument and is non-variadic).  If it can have more than one
 * aggregated argument, we process the aggregated arguments as independent
 * sort columns.  This avoids throwing error for something like
 *		agg(...) within group (order by x collate "foo", y collate "bar")
 * while also guaranteeing that variadic aggregates don't change in behavior
 * depending on how many sort columns a particular call happens to have.
 *
 * Otherwise this is much like the plain-aggregate case.
 */


/*
 * Hypothetical-set aggregates are even more special: per spec, we need to
 * unify the collations of each pair of hypothetical and aggregated args.
 * And we need to force the choice of collation down into the sort column
 * to ensure that the sort happens with the chosen collation.  Other than
 * that, the behavior is like regular ordered-set aggregates.  Note that
 * hypothetical direct arguments contribute to the aggregate collation
 * only when their partner aggregated arguments do.
 */

