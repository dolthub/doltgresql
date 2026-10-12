/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - contain_var_clause
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * var.c
 *	  Var node manipulation routines
 *
 * Note: for most purposes, PlaceHolderVar is considered a Var too,
 * even if its contained expression is variable-free.  Also, CurrentOfExpr
 * is treated as a Var for purposes of determining whether an expression
 * contains variables.
 *
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 *
 * IDENTIFICATION
 *	  src/backend/optimizer/util/var.c
 *
 *-------------------------------------------------------------------------
 */
#include "postgres.h"

#include "access/sysattr.h"
#include "nodes/nodeFuncs.h"
#include "optimizer/clauses.h"
#include "optimizer/optimizer.h"
#include "optimizer/placeholder.h"
#include "optimizer/prep.h"
#include "parser/parsetree.h"
#include "rewrite/rewriteManip.h"


typedef struct
{
	Relids		varnos;
	PlannerInfo *root;
	int			sublevels_up;
} pull_varnos_context;

typedef struct
{
	Bitmapset  *varattnos;
	Index		varno;
} pull_varattnos_context;

typedef struct
{
	List	   *vars;
	int			sublevels_up;
} pull_vars_context;

typedef struct
{
	int			var_location;
	int			sublevels_up;
} locate_var_of_level_context;

typedef struct
{
	List	   *varlist;
	int			flags;
} pull_var_clause_context;

typedef struct
{
	PlannerInfo *root;			/* could be NULL! */
	Query	   *query;			/* outer Query */
	int			sublevels_up;
	bool		possible_sublink;	/* could aliases include a SubLink? */
	bool		inserted_sublink;	/* have we inserted a SubLink? */
} flatten_join_alias_vars_context;

static bool pull_varnos_walker(Node *node,
							   pull_varnos_context *context);
static bool pull_varattnos_walker(Node *node, pull_varattnos_context *context);
static bool pull_vars_walker(Node *node, pull_vars_context *context);
static bool contain_var_clause_walker(Node *node, void *context);
static bool contain_vars_of_level_walker(Node *node, int *sublevels_up);
static bool contain_vars_returning_old_or_new_walker(Node *node, void *context);
static bool locate_var_of_level_walker(Node *node,
									   locate_var_of_level_context *context);
static bool pull_var_clause_walker(Node *node,
								   pull_var_clause_context *context);
static Node *flatten_join_alias_vars_mutator(Node *node,
											 flatten_join_alias_vars_context *context);
static Node *flatten_group_exprs_mutator(Node *node,
										 flatten_join_alias_vars_context *context);
static Node *mark_nullable_by_grouping(PlannerInfo *root, Node *newnode,
									   Var *oldvar);
static Node *add_nullingrels_if_needed(PlannerInfo *root, Node *newnode,
									   Var *oldvar);
static bool is_standard_join_alias_expression(Node *newnode, Var *oldvar);
static void adjust_standard_join_alias_expression(Node *newnode, Var *oldvar);
static Relids alias_relid_set(Query *query, Relids relids);


/*
 * pull_varnos
 *		Create a set of all the distinct varnos present in a parsetree.
 *		Only varnos that reference level-zero rtable entries are considered.
 *
 * The result includes outer-join relids mentioned in Var.varnullingrels and
 * PlaceHolderVar.phnullingrels fields in the parsetree.
 *
 * "root" can be passed as NULL if it is not necessary to process
 * PlaceHolderVars.
 *
 * NOTE: this is used on not-yet-planned expressions.  It may therefore find
 * bare SubLinks, and if so it needs to recurse into them to look for uplevel
 * references to the desired rtable level!	But when we find a completed
 * SubPlan, we only need to look at the parameters passed to the subplan.
 */


/*
 * pull_varnos_of_level
 *		Create a set of all the distinct varnos present in a parsetree.
 *		Only Vars of the specified level are considered.
 */





/*
 * pull_varattnos
 *		Find all the distinct attribute numbers present in an expression tree,
 *		and add them to the initial contents of *varattnos.
 *		Only Vars of the given varno and rtable level zero are considered.
 *
 * Attribute numbers are offset by FirstLowInvalidHeapAttributeNumber so that
 * we can include system attributes (e.g., OID) in the bitmap representation.
 *
 * Currently, this does not support unplanned subqueries; that is not needed
 * for current uses.  It will handle already-planned SubPlan nodes, though,
 * looking into only the "testexpr" and the "args" list.  (The subplan cannot
 * contain any other references to Vars of the current level.)
 */





/*
 * pull_vars_of_level
 *		Create a list of all Vars (and PlaceHolderVars) referencing the
 *		specified query level in the given parsetree.
 *
 * Caution: the Vars are not copied, only linked into the list.
 */





/*
 * contain_var_clause
 *	  Recursively scan a clause to discover whether it contains any Var nodes
 *	  (of the current query level).
 *
 *	  Returns true if any varnode found.
 *
 * Does not examine subqueries, therefore must only be used after reduction
 * of sublinks to subplans!
 */

bool
contain_var_clause(Node *node)
{
return false;}




/*
 * contain_vars_of_level
 *	  Recursively scan a clause to discover whether it contains any Var nodes
 *	  of the specified query level.
 *
 *	  Returns true if any such Var found.
 *
 * Will recurse into sublinks.  Also, may be invoked directly on a Query.
 */





/*
 * contain_vars_returning_old_or_new
 *	  Recursively scan a clause to discover whether it contains any Var nodes
 *	  (of the current query level) whose varreturningtype is VAR_RETURNING_OLD
 *	  or VAR_RETURNING_NEW.
 *
 *	  Returns true if any found.
 *
 * Any ReturningExprs are also detected --- if an OLD/NEW Var was rewritten,
 * we still regard this as a clause that returns OLD/NEW values.
 *
 * Does not examine subqueries, therefore must only be used after reduction
 * of sublinks to subplans!
 */





/*
 * locate_var_of_level
 *	  Find the parse location of any Var of the specified query level.
 *
 * Returns -1 if no such Var is in the querytree, or if they all have
 * unknown parse location.  (The former case is probably caller error,
 * but we don't bother to distinguish it from the latter case.)
 *
 * Will recurse into sublinks.  Also, may be invoked directly on a Query.
 *
 * Note: it might seem appropriate to merge this functionality into
 * contain_vars_of_level, but that would complicate that function's API.
 * Currently, the only uses of this function are for error reporting,
 * and so shaving cycles probably isn't very important.
 */





/*
 * pull_var_clause
 *	  Recursively pulls all Var nodes from an expression clause.
 *
 *	  Aggrefs are handled according to these bits in 'flags':
 *		PVC_INCLUDE_AGGREGATES		include Aggrefs in output list
 *		PVC_RECURSE_AGGREGATES		recurse into Aggref arguments
 *		neither flag				throw error if Aggref found
 *	  Vars within an Aggref's expression are included in the result only
 *	  when PVC_RECURSE_AGGREGATES is specified.
 *
 *	  WindowFuncs are handled according to these bits in 'flags':
 *		PVC_INCLUDE_WINDOWFUNCS		include WindowFuncs in output list
 *		PVC_RECURSE_WINDOWFUNCS		recurse into WindowFunc arguments
 *		neither flag				throw error if WindowFunc found
 *	  Vars within a WindowFunc's expression are included in the result only
 *	  when PVC_RECURSE_WINDOWFUNCS is specified.
 *
 *	  PlaceHolderVars are handled according to these bits in 'flags':
 *		PVC_INCLUDE_PLACEHOLDERS	include PlaceHolderVars in output list
 *		PVC_RECURSE_PLACEHOLDERS	recurse into PlaceHolderVar arguments
 *		neither flag				throw error if PlaceHolderVar found
 *	  Vars within a PHV's expression are included in the result only
 *	  when PVC_RECURSE_PLACEHOLDERS is specified.
 *
 *	  GroupingFuncs are treated exactly like Aggrefs, and so do not need
 *	  their own flag bits.
 *
 *	  CurrentOfExpr nodes are ignored in all cases.
 *
 *	  Upper-level vars (with varlevelsup > 0) should not be seen here,
 *	  likewise for upper-level Aggrefs and PlaceHolderVars.
 *
 *	  Returns list of nodes found.  Note the nodes themselves are not
 *	  copied, only referenced.
 *
 * Does not examine subqueries, therefore must only be used after reduction
 * of sublinks to subplans!
 */





/*
 * flatten_join_alias_vars
 *	  Replace Vars that reference JOIN outputs with references to the original
 *	  relation variables instead.  This allows quals involving such vars to be
 *	  pushed down.  Whole-row Vars that reference JOIN relations are expanded
 *	  into RowExpr constructs that name the individual output Vars.  This
 *	  is necessary since we will not scan the JOIN as a base relation, which
 *	  is the only way that the executor can directly handle whole-row Vars.
 *
 * This also adjusts relid sets found in some expression node types to
 * substitute the contained base+OJ rels for any join relid.
 *
 * If a JOIN contains sub-selects that have been flattened, its join alias
 * entries might now be arbitrary expressions, not just Vars.  This affects
 * this function in two important ways.  First, we might find ourselves
 * inserting SubLink expressions into subqueries, and we must make sure that
 * their Query.hasSubLinks fields get set to true if so.  If there are any
 * SubLinks in the join alias lists, the outer Query should already have
 * hasSubLinks = true, so this is only relevant to un-flattened subqueries.
 * Second, we have to preserve any varnullingrels info attached to the
 * alias Vars we're replacing.  If the replacement expression is a Var or
 * PlaceHolderVar or constructed from those, we can just add the
 * varnullingrels bits to the existing nullingrels field(s); otherwise
 * we have to add a PlaceHolderVar wrapper.
 *
 * NOTE: this is also used by the parser, to expand join alias Vars before
 * checking GROUP BY validity.  For that use-case, root will be NULL, which
 * is why we have to pass the Query separately.  We need the root itself only
 * for making PlaceHolderVars.  We can avoid making PlaceHolderVars in the
 * parser's usage because it won't be dealing with arbitrary expressions:
 * so long as adjust_standard_join_alias_expression can handle everything
 * the parser would make as a join alias expression, we're OK.
 */




/*
 * flatten_group_exprs
 *	  Replace Vars that reference GROUP outputs with the underlying grouping
 *	  expressions.
 *
 * We have to preserve any varnullingrels info attached to the group Vars we're
 * replacing.  If the replacement expression is a Var or PlaceHolderVar or
 * constructed from those, we can just add the varnullingrels bits to the
 * existing nullingrels field(s); otherwise we have to add a PlaceHolderVar
 * wrapper.
 *
 * NOTE: this is also used by ruleutils.c, to deparse one query parsetree back
 * to source text, and by check_output_expressions() to check for unsafe
 * pushdowns.  For these use-cases, root will be NULL, which is why we have to
 * pass the Query separately.  We need the root itself only for preserving
 * varnullingrels.  We can avoid preserving varnullingrels in the ruleutils.c's
 * usage because it does not make any difference to the deparsed source text.
 * We can also avoid it in check_output_expressions() because that function
 * uses the expanded expressions solely to check for volatile or set-returning
 * functions, which is independent of the Vars' nullingrels.
 */




/*
 * Add oldvar's varnullingrels, if any, to a flattened grouping expression.
 * The newnode has been copied, so we can modify it freely.
 */


/*
 * Add oldvar's varnullingrels, if any, to a flattened join alias expression.
 * The newnode has been copied, so we can modify it freely.
 */


/*
 * Check to see if we can insert nullingrels into this join alias expression
 * without use of a separate PlaceHolderVar.
 *
 * This will handle Vars, PlaceHolderVars, and implicit-coercion and COALESCE
 * expressions built from those.  This coverage needs to handle anything
 * that the parser would put into joinaliasvars.
 */


/*
 * Insert nullingrels into an expression accepted by
 * is_standard_join_alias_expression.
 */


/*
 * alias_relid_set: in a set of RT indexes, replace joins by their
 * underlying base+OJ relids
 */

