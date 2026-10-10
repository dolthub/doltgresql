/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - transformExpr
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * parse_expr.c
 *	  handle expressions in parser
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 *
 * IDENTIFICATION
 *	  src/backend/parser/parse_expr.c
 *
 *-------------------------------------------------------------------------
 */

#include "postgres.h"

#include "catalog/pg_aggregate.h"
#include "catalog/pg_type.h"
#include "commands/dbcommands.h"
#include "miscadmin.h"
#include "nodes/makefuncs.h"
#include "nodes/nodeFuncs.h"
#include "optimizer/optimizer.h"
#include "parser/analyze.h"
#include "parser/parse_agg.h"
#include "parser/parse_clause.h"
#include "parser/parse_coerce.h"
#include "parser/parse_collate.h"
#include "parser/parse_expr.h"
#include "parser/parse_func.h"
#include "parser/parse_oper.h"
#include "parser/parse_relation.h"
#include "parser/parse_target.h"
#include "parser/parse_type.h"
#include "utils/builtins.h"
#include "utils/date.h"
#include "utils/fmgroids.h"
#include "utils/lsyscache.h"
#include "utils/timestamp.h"
#include "utils/xml.h"

/* GUC parameters */



static Node *transformExprRecurse(ParseState *pstate, Node *expr);
static Node *transformParamRef(ParseState *pstate, ParamRef *pref);
static Node *transformAExprOp(ParseState *pstate, A_Expr *a);
static Node *transformAExprOpAny(ParseState *pstate, A_Expr *a);
static Node *transformAExprOpAll(ParseState *pstate, A_Expr *a);
static Node *transformAExprDistinct(ParseState *pstate, A_Expr *a);
static Node *transformAExprNullIf(ParseState *pstate, A_Expr *a);
static Node *transformAExprIn(ParseState *pstate, A_Expr *a);
static Node *transformAExprBetween(ParseState *pstate, A_Expr *a);
static Node *transformMergeSupportFunc(ParseState *pstate, MergeSupportFunc *f);
static Node *transformBoolExpr(ParseState *pstate, BoolExpr *a);
static Node *transformFuncCall(ParseState *pstate, FuncCall *fn);
static Node *transformMultiAssignRef(ParseState *pstate, MultiAssignRef *maref);
static Node *transformCaseExpr(ParseState *pstate, CaseExpr *c);
static Node *transformSubLink(ParseState *pstate, SubLink *sublink);
static Node *transformArrayExpr(ParseState *pstate, A_ArrayExpr *a,
								Oid array_type, Oid element_type, int32 typmod);
static Node *transformRowExpr(ParseState *pstate, RowExpr *r, bool allowDefault);
static Node *transformCoalesceExpr(ParseState *pstate, CoalesceExpr *c);
static Node *transformMinMaxExpr(ParseState *pstate, MinMaxExpr *m);
static Node *transformSQLValueFunction(ParseState *pstate,
									   SQLValueFunction *svf);
static Node *transformXmlExpr(ParseState *pstate, XmlExpr *x);
static Node *transformXmlSerialize(ParseState *pstate, XmlSerialize *xs);
static Node *transformBooleanTest(ParseState *pstate, BooleanTest *b);
static Node *transformCurrentOfExpr(ParseState *pstate, CurrentOfExpr *cexpr);
static Node *transformColumnRef(ParseState *pstate, ColumnRef *cref);
static Node *transformWholeRowRef(ParseState *pstate,
								  ParseNamespaceItem *nsitem,
								  int sublevels_up, int location);
static Node *transformIndirection(ParseState *pstate, A_Indirection *ind);
static Node *transformTypeCast(ParseState *pstate, TypeCast *tc);
static Node *transformCollateClause(ParseState *pstate, CollateClause *c);
static Node *transformJsonObjectConstructor(ParseState *pstate,
											JsonObjectConstructor *ctor);
static Node *transformJsonArrayConstructor(ParseState *pstate,
										   JsonArrayConstructor *ctor);
static Node *transformJsonArrayQueryConstructor(ParseState *pstate,
												JsonArrayQueryConstructor *ctor);
static Node *transformJsonObjectAgg(ParseState *pstate, JsonObjectAgg *agg);
static Node *transformJsonArrayAgg(ParseState *pstate, JsonArrayAgg *agg);
static Node *transformJsonIsPredicate(ParseState *pstate, JsonIsPredicate *pred);
static Node *transformJsonParseExpr(ParseState *pstate, JsonParseExpr *jsexpr);
static Node *transformJsonScalarExpr(ParseState *pstate, JsonScalarExpr *jsexpr);
static Node *transformJsonSerializeExpr(ParseState *pstate,
										JsonSerializeExpr *expr);
static Node *transformJsonFuncExpr(ParseState *pstate, JsonFuncExpr *func);
static void transformJsonPassingArgs(ParseState *pstate, const char *constructName,
									 JsonFormatType format, List *args,
									 List **passing_values, List **passing_names);
static JsonBehavior *transformJsonBehavior(ParseState *pstate, JsonExpr *jsexpr,
										   JsonBehavior *behavior,
										   JsonBehaviorType default_behavior,
										   JsonReturning *returning);
static Node *GetJsonBehaviorConst(JsonBehaviorType btype, int location);
static Node *make_row_comparison_op(ParseState *pstate, List *opname,
									List *largs, List *rargs, int location);
static Node *make_row_distinct_op(ParseState *pstate, List *opname,
								  RowExpr *lrow, RowExpr *rrow, int location);
static Expr *make_distinct_op(ParseState *pstate, List *opname,
							  Node *ltree, Node *rtree, int location);
static Node *make_nulltest_from_distinct(ParseState *pstate,
										 A_Expr *distincta, Node *arg);


/*
 * transformExpr -
 *	  Analyze and transform expressions. Type checking and type casting is
 *	  done here.  This processing converts the raw grammar output into
 *	  expression trees with fully determined semantics.
 */

Node *
transformExpr(ParseState *pstate, Node *expr, ParseExprKind exprKind)
{
return expr;}



/*
 * helper routine for delivering "column does not exist" error message
 *
 * (Usually we don't have to work this hard, but the general case of field
 * selection from an arbitrary node needs it.)
 */




/*
 * Transform a ColumnRef.
 *
 * If you find yourself changing this code, see also ExpandColumnRefStar.
 */




/* Test whether an a_expr is a plain NULL constant or not */




























/*
 * transformArrayExpr
 *
 * If the caller specifies the target type, the resulting array will
 * be of exactly that type.  Otherwise we try to infer a common type
 * for the elements using select_common_type().
 */


















/*
 * Construct a whole-row reference to represent the notation "relation.*".
 */


/*
 * Handle an explicit CAST construct.
 *
 * Transform the argument, look up the type name, and apply any necessary
 * coercion function(s).
 */


/*
 * Handle an explicit COLLATE clause.
 *
 * Transform the argument, and look up the collation name.
 */


/*
 * Transform a "row compare-op row" construct
 *
 * The inputs are lists of already-transformed expressions.
 * As with coerce_type, pstate may be NULL if no special unknown-Param
 * processing is wanted.
 *
 * The output may be a single OpExpr, an AND or OR combination of OpExprs,
 * or a RowCompareExpr.  In all cases it is guaranteed to return boolean.
 * The AND, OR, and RowCompareExpr cases further imply things about the
 * behavior of the operators (ie, they behave as =, <>, or < <= > >=).
 */


/*
 * Transform a "row IS DISTINCT FROM row" construct
 *
 * The input RowExprs are already transformed
 */


/*
 * make the node for an IS DISTINCT FROM operator
 */


/*
 * Produce a NullTest node from an IS [NOT] DISTINCT FROM NULL construct
 *
 * "arg" is the untransformed other argument
 */


/*
 * Produce a string identifying an expression by kind.
 *
 * Note: when practical, use a simple SQL keyword for the result.  If that
 * doesn't work well, check call sites to see whether custom error message
 * strings are required.
 */


/*
 * Make string Const node from JSON encoding name.
 *
 * UTF8 is default encoding.
 */


/*
 * Make bytea => text conversion using specified JSON format encoding.
 */


/*
 * Transform JSON value expression using specified input JSON format or
 * default format otherwise, coercing to the targettype if needed.
 *
 * Returned expression is either ve->raw_expr coerced to text (if needed) or
 * a JsonValueExpr with formatted_expr set to the coerced copy of raw_expr
 * if the specified format and the targettype requires it.
 */


/*
 * Checks specified output format for its applicability to the target type.
 */


/*
 * Transform JSON output clause.
 *
 * Assigns target type oid and modifier.
 * Assigns default format or checks specified format for its applicability to
 * the target type.
 */


/*
 * Transform JSON output clause of JSON constructor functions.
 *
 * Derive RETURNING type, if not specified, from argument types.
 */


/*
 * Coerce json[b]-valued function expression to the output type.
 */


/*
 * Make a JsonConstructorExpr node.
 */


/*
 * Transform JSON_OBJECT() constructor.
 *
 * JSON_OBJECT() is transformed into a JsonConstructorExpr node of type
 * JSCTOR_JSON_OBJECT.  The result is coerced to the target type given
 * by ctor->output.
 */


/*
 * Transform JSON_ARRAY(query [FORMAT] [RETURNING] [ON NULL]) into
 *  (SELECT  JSON_ARRAYAGG(a  [FORMAT] [RETURNING] [ON NULL]) FROM (query) q(a))
 */


/*
 * Common code for JSON_OBJECTAGG and JSON_ARRAYAGG transformation.
 */


/*
 * Transform JSON_OBJECTAGG() aggregate function.
 *
 * JSON_OBJECTAGG() is transformed into a JsonConstructorExpr node of type
 * JSCTOR_JSON_OBJECTAGG, which at runtime becomes a
 * json[b]_object_agg[_unique][_strict](agg->arg->key, agg->arg->value) call
 * depending on the output JSON format.  The result is coerced to the target
 * type given by agg->constructor->output.
 */


/*
 * Transform JSON_ARRAYAGG() aggregate function.
 *
 * JSON_ARRAYAGG() is transformed into a JsonConstructorExpr node of type
 * JSCTOR_JSON_ARRAYAGG, which at runtime becomes a
 * json[b]_object_agg[_unique][_strict](agg->arg) call depending on the output
 * JSON format.  The result is coerced to the target type given by
 * agg->constructor->output.
 */


/*
 * Transform JSON_ARRAY() constructor.
 *
 * JSON_ARRAY() is transformed into a JsonConstructorExpr node of type
 * JSCTOR_JSON_ARRAY.  The result is coerced to the target type given
 * by ctor->output.
 */




/*
 * Transform IS JSON predicate.
 */


/*
 * Transform the RETURNING clause of a JSON_*() expression if there is one and
 * create one if not.
 */


/*
 * Transform a JSON() expression.
 *
 * JSON() is transformed into a JsonConstructorExpr of type JSCTOR_JSON_PARSE,
 * which validates the input expression value as JSON.
 */


/*
 * Transform a JSON_SCALAR() expression.
 *
 * JSON_SCALAR() is transformed into a JsonConstructorExpr of type
 * JSCTOR_JSON_SCALAR, which converts the input SQL scalar value into
 * a json[b] value.
 */


/*
 * Transform a JSON_SERIALIZE() expression.
 *
 * JSON_SERIALIZE() is transformed into a JsonConstructorExpr of type
 * JSCTOR_JSON_SERIALIZE which converts the input JSON value into a character
 * or bytea string.
 */


/*
 * Transform JSON_VALUE, JSON_QUERY, JSON_EXISTS, JSON_TABLE functions into
 * a JsonExpr node.
 */


/*
 * Transform a SQL/JSON PASSING clause.
 */


/*
 * Recursively checks if the given expression, or its sub-node in some cases,
 * is valid for using as an ON ERROR / ON EMPTY DEFAULT expression.
 */


/*
 * Transform a JSON BEHAVIOR clause.
 */


/*
 * Returns a Const node holding the value for the given non-ERROR
 * JsonBehaviorType.
 */

