/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - coerce_to_specific_type
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * parse_coerce.c
 *		handle type coercions/conversions for parser
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 *
 * IDENTIFICATION
 *	  src/backend/parser/parse_coerce.c
 *
 *-------------------------------------------------------------------------
 */
#include "postgres.h"

#include "catalog/pg_cast.h"
#include "catalog/pg_class.h"
#include "catalog/pg_inherits.h"
#include "catalog/pg_proc.h"
#include "catalog/pg_type.h"
#include "nodes/makefuncs.h"
#include "nodes/nodeFuncs.h"
#include "parser/parse_coerce.h"
#include "parser/parse_relation.h"
#include "parser/parse_type.h"
#include "utils/builtins.h"
#include "utils/datum.h"		/* needed for datumIsEqual() */
#include "utils/fmgroids.h"
#include "utils/lsyscache.h"
#include "utils/syscache.h"
#include "utils/typcache.h"


static Node *coerce_type_typmod(Node *node,
								Oid targetTypeId, int32 targetTypMod,
								CoercionContext ccontext, CoercionForm cformat,
								int location,
								bool hideInputCoercion);
static void hide_coercion_node(Node *node);
static Node *build_coercion_expression(Node *node,
									   CoercionPathType pathtype,
									   Oid funcId,
									   Oid targetTypeId, int32 targetTypMod,
									   CoercionContext ccontext, CoercionForm cformat,
									   int location);
static Node *coerce_record_to_complex(ParseState *pstate, Node *node,
									  Oid targetTypeId,
									  CoercionContext ccontext,
									  CoercionForm cformat,
									  int location);
static bool is_complex_array(Oid typid);
static bool typeIsOfTypedTable(Oid reltypeId, Oid reloftypeId);


/*
 * coerce_to_target_type()
 *		Convert an expression to a target type and typmod.
 *
 * This is the general-purpose entry point for arbitrary type coercion
 * operations.  Direct use of the component operations can_coerce_type,
 * coerce_type, and coerce_type_typmod should be restricted to special
 * cases (eg, when the conversion is expected to succeed).
 *
 * Returns the possibly-transformed expression tree, or NULL if the type
 * conversion is not possible.  (We do this, rather than ereport'ing directly,
 * so that callers can generate custom error messages indicating context.)
 *
 * pstate - parse state (can be NULL, see coerce_type)
 * expr - input expression tree (already transformed by transformExpr)
 * exprtype - result type of expr
 * targettype - desired result type
 * targettypmod - desired result typmod
 * ccontext, cformat - context indicators to control coercions
 * location - parse location of the coercion request, or -1 if unknown/implicit
 */



/*
 * coerce_type()
 *		Convert an expression to a different type.
 *
 * The caller should already have determined that the coercion is possible;
 * see can_coerce_type.
 *
 * Normally, no coercion to a typmod (length) is performed here.  The caller
 * must call coerce_type_typmod as well, if a typmod constraint is wanted.
 * (But if the target type is a domain, it may internally contain a
 * typmod constraint, which will be applied inside coerce_to_domain.)
 * In some cases pg_cast specifies a type coercion function that also
 * applies length conversion, and in those cases only, the result will
 * already be properly coerced to the specified typmod.
 *
 * pstate is only used in the case that we are able to resolve the type of
 * a previously UNKNOWN Param.  It is okay to pass pstate = NULL if the
 * caller does not want type information updated for Params.
 *
 * Note: this function must not modify the given expression tree, only add
 * decoration on top of it.  See transformSetOperationTree, for example.
 */
#ifdef RANDOMIZE_ALLOCATED_MEMORY
#endif
#ifdef NOT_USED
#endif


/*
 * can_coerce_type()
 *		Can input_typeids be coerced to target_typeids?
 *
 * We must be told the context (CAST construct, assignment, implicit coercion)
 * as this determines the set of available casts.
 */
#ifdef NOT_USED					/* not implemented yet */
#endif


/*
 * Create an expression tree to represent coercion to a domain type.
 *
 * 'arg': input expression
 * 'baseTypeId': base type of domain
 * 'baseTypeMod': base type typmod of domain
 * 'typeId': target type to coerce to
 * 'ccontext': context indicator to control coercions
 * 'cformat': coercion display format
 * 'location': coercion request location
 * 'hideInputCoercion': if true, hide the input coercion under this one.
 *
 * If the target type isn't a domain, the given 'arg' is returned as-is.
 */



/*
 * coerce_type_typmod()
 *		Force a value to a particular typmod, if meaningful and possible.
 *
 * This is applied to values that are going to be stored in a relation
 * (where we have an atttypmod for the column) as well as values being
 * explicitly CASTed (where the typmod comes from the target type spec).
 *
 * The caller must have already ensured that the value is of the correct
 * type, typically by applying coerce_type.
 *
 * ccontext may affect semantics, depending on whether the length coercion
 * function pays attention to the isExplicit flag it's passed.
 *
 * cformat determines the display properties of the generated node (if any).
 *
 * If hideInputCoercion is true *and* we generate a node, the input node is
 * forced to IMPLICIT display form, so that only the typmod coercion node will
 * be visible when displaying the expression.
 *
 * NOTE: this does not need to work on domain types, because any typmod
 * coercion for a domain is considered to be part of the type coercion
 * needed to produce the domain value in the first place.  So, no getBaseType.
 */


/*
 * Mark a coercion node as IMPLICIT so it will never be displayed by
 * ruleutils.c.  We use this when we generate a nest of coercion nodes
 * to implement what is logically one conversion; the inner nodes are
 * forced to IMPLICIT_CAST format.  This does not change their semantics,
 * only display behavior.
 *
 * It is caller error to call this on something that doesn't have a
 * CoercionForm field.
 */


/*
 * build_coercion_expression()
 *		Construct an expression tree for applying a pg_cast entry.
 *
 * This is used for both type-coercion and length-coercion operations,
 * since there is no difference in terms of the calling convention.
 */



/*
 * coerce_record_to_complex
 *		Coerce a RECORD to a specific composite type.
 *
 * Currently we only support this for inputs that are RowExprs or whole-row
 * Vars.
 */


/*
 * coerce_to_boolean()
 *		Coerce an argument of a construct that requires boolean input
 *		(AND, OR, NOT, etc).  Also check that input is not a set.
 *
 * Returns the possibly-transformed node tree.
 *
 * As with coerce_type, pstate may be NULL if no special unknown-Param
 * processing is wanted.
 */


/*
 * coerce_to_specific_type_typmod()
 *		Coerce an argument of a construct that requires a specific data type,
 *		with a specific typmod.  Also check that input is not a set.
 *
 * Returns the possibly-transformed node tree.
 *
 * As with coerce_type, pstate may be NULL if no special unknown-Param
 * processing is wanted.
 */


/*
 * coerce_to_specific_type()
 *		Coerce an argument of a construct that requires a specific data type.
 *		Also check that input is not a set.
 *
 * Returns the possibly-transformed node tree.
 *
 * As with coerce_type, pstate may be NULL if no special unknown-Param
 * processing is wanted.
 */

Node *
coerce_to_specific_type(ParseState *pstate, Node *node,
						Oid targetTypeId,
						const char *constructName)
{
return node;}

/*
 * coerce_null_to_domain()
 *		Build a NULL constant, then wrap it in CoerceToDomain
 *		if the desired type is a domain type.  This allows any
 *		NOT NULL domain constraint to be enforced at runtime.
 */


/*
 * parser_coercion_errposition - report coercion error location, if possible
 *
 * We prefer to point at the coercion request (CAST, ::, etc) if possible;
 * but there may be no such location in the case of an implicit coercion.
 * In that case point at the input expression.
 *
 * XXX possibly this is more generally useful than coercion errors;
 * if so, should rename and place with parser_errposition.
 */



/*
 * select_common_type()
 *		Determine the common supertype of a list of input expressions.
 *		This is used for determining the output type of CASE, UNION,
 *		and similar constructs.
 *
 * 'exprs' is a *nonempty* list of expressions.  Note that earlier items
 * in the list will be preferred if there is doubt.
 * 'context' is a phrase to use in the error message if we fail to select
 * a usable type.  Pass NULL to have the routine return InvalidOid
 * rather than throwing an error on failure.
 * 'which_expr': if not NULL, receives a pointer to the particular input
 * expression from which the result type was taken.
 *
 * Caution: "failure" just means that there were inputs of different type
 * categories.  It is not guaranteed that all the inputs are coercible to the
 * selected type; caller must check that (see verify_common_type).
 */


/*
 * select_common_type_from_oids()
 *		Determine the common supertype of an array of type OIDs.
 *
 * This is the same logic as select_common_type(), but working from
 * an array of type OIDs not a list of expressions.  As in that function,
 * earlier entries in the array have some preference over later ones.
 * On failure, return InvalidOid if noerror is true, else throw an error.
 *
 * Caution: "failure" just means that there were inputs of different type
 * categories.  It is not guaranteed that all the inputs are coercible to the
 * selected type; caller must check that (see verify_common_type_from_oids).
 *
 * Note: neither caller will pass any UNKNOWNOID entries, so the tests
 * for that in this function are dead code.  However, they don't cost much,
 * and it seems better to keep this logic as close to select_common_type()
 * as possible.
 */


/*
 * coerce_to_common_type()
 *		Coerce an expression to the given type.
 *
 * This is used following select_common_type() to coerce the individual
 * expressions to the desired type.  'context' is a phrase to use in the
 * error message if we fail to coerce.
 *
 * As with coerce_type, pstate may be NULL if no special unknown-Param
 * processing is wanted.
 */


/*
 * verify_common_type()
 *		Verify that all input types can be coerced to a proposed common type.
 *		Return true if so, false if not all coercions are possible.
 *
 * Most callers of select_common_type() don't need to do this explicitly
 * because the checks will happen while trying to convert input expressions
 * to the right type, e.g. in coerce_to_common_type().  However, if a separate
 * check step is needed to validate the applicability of the common type, call
 * this.
 */


/*
 * verify_common_type_from_oids()
 *		As above, but work from an array of type OIDs.
 */


/*
 * select_common_typmod()
 *		Determine the common typmod of a list of input expressions.
 *
 * common_type is the selected common type of the expressions, typically
 * computed using select_common_type().
 */


/*
 * check_generic_type_consistency()
 *		Are the actual arguments potentially compatible with a
 *		polymorphic function?
 *
 * The argument consistency rules are:
 *
 * 1) All arguments declared ANYELEMENT must have the same datatype.
 * 2) All arguments declared ANYARRAY must have the same datatype,
 *	  which must be a varlena array type.
 * 3) All arguments declared ANYRANGE must be the same range type.
 *	  Similarly, all arguments declared ANYMULTIRANGE must be the same
 *	  multirange type; and if both of these appear, the ANYRANGE type
 *	  must be the element type of the ANYMULTIRANGE type.
 * 4) If there are arguments of more than one of these polymorphic types,
 *	  the array element type and/or range subtype must be the same as each
 *	  other and the same as the ANYELEMENT type.
 * 5) ANYENUM is treated the same as ANYELEMENT except that if it is used
 *	  (alone or in combination with plain ANYELEMENT), we add the extra
 *	  condition that the ANYELEMENT type must be an enum.
 * 6) ANYNONARRAY is treated the same as ANYELEMENT except that if it is used,
 *	  we add the extra condition that the ANYELEMENT type must not be an array.
 *	  (This is a no-op if used in combination with ANYARRAY or ANYENUM, but
 *	  is an extra restriction if not.)
 * 7) All arguments declared ANYCOMPATIBLE must be implicitly castable
 *	  to a common supertype (chosen as per select_common_type's rules).
 *	  ANYCOMPATIBLENONARRAY works like ANYCOMPATIBLE but also requires the
 *	  common supertype to not be an array.  If there are ANYCOMPATIBLEARRAY
 *	  or ANYCOMPATIBLERANGE or ANYCOMPATIBLEMULTIRANGE arguments, their element
 *	  types or subtypes are included while making the choice of common supertype.
 * 8) The resolved type of ANYCOMPATIBLEARRAY arguments will be the array
 *	  type over the common supertype (which might not be the same array type
 *	  as any of the original arrays).
 * 9) All ANYCOMPATIBLERANGE arguments must be the exact same range type
 *	  (after domain flattening), since we have no preference rule that would
 *	  let us choose one over another.  Furthermore, that range's subtype
 *	  must exactly match the common supertype chosen by rule 7.
 * 10) All ANYCOMPATIBLEMULTIRANGE arguments must be the exact same multirange
 *	  type (after domain flattening), since we have no preference rule that
 *	  would let us choose one over another.  Furthermore, if ANYCOMPATIBLERANGE
 *	  also appears, that range type must be the multirange's element type;
 *	  otherwise, the multirange's range's subtype must exactly match the
 *	  common supertype chosen by rule 7.
 *
 * Domains over arrays match ANYARRAY, and are immediately flattened to their
 * base type.  (Thus, for example, we will consider it a match if one ANYARRAY
 * argument is a domain over int4[] while another one is just int4[].)	Also
 * notice that such a domain does *not* match ANYNONARRAY.  The same goes
 * for ANYCOMPATIBLEARRAY and ANYCOMPATIBLENONARRAY.
 *
 * Similarly, domains over ranges match ANYRANGE or ANYCOMPATIBLERANGE,
 * and are immediately flattened to their base type.  Likewise, domains
 * over multiranges match ANYMULTIRANGE or ANYCOMPATIBLEMULTIRANGE and are
 * immediately flattened to their base type.
 *
 * Note that domains aren't currently considered to match ANYENUM,
 * even if their base type would match.
 *
 * If we have UNKNOWN input (ie, an untyped literal) for any polymorphic
 * argument, assume it is okay.
 *
 * We do not ereport here, but just return false if a rule is violated.
 */


/*
 * enforce_generic_type_consistency()
 *		Make sure a polymorphic function is legally callable, and
 *		deduce actual argument and result types.
 *
 * If any polymorphic pseudotype is used in a function's arguments or
 * return type, we make sure the actual data types are consistent with
 * each other.  The argument consistency rules are shown above for
 * check_generic_type_consistency().
 *
 * If we have UNKNOWN input (ie, an untyped literal) for any polymorphic
 * argument, we attempt to deduce the actual type it should have.  If
 * successful, we alter that position of declared_arg_types[] so that
 * make_fn_arguments will coerce the literal to the right thing.
 *
 * If we have polymorphic arguments of the ANYCOMPATIBLE family,
 * we similarly alter declared_arg_types[] entries to show the resolved
 * common supertype, so that make_fn_arguments will coerce the actual
 * arguments to the proper type.
 *
 * Rules are applied to the function's return type (possibly altering it)
 * if it is declared as a polymorphic type and there is at least one
 * polymorphic argument type:
 *
 * 1) If return type is ANYELEMENT, and any argument is ANYELEMENT, use the
 *	  argument's actual type as the function's return type.
 * 2) If return type is ANYARRAY, and any argument is ANYARRAY, use the
 *	  argument's actual type as the function's return type.
 * 3) Similarly, if return type is ANYRANGE or ANYMULTIRANGE, and any
 *	  argument is ANYRANGE or ANYMULTIRANGE, use that argument's actual type
 *	  (or the corresponding range or multirange type) as the function's return
 *	  type.
 * 4) Otherwise, if return type is ANYELEMENT or ANYARRAY, and there is
 *	  at least one ANYELEMENT, ANYARRAY, ANYRANGE, or ANYMULTIRANGE input,
 *	  deduce the return type from those inputs, or throw error if we can't.
 * 5) Otherwise, if return type is ANYRANGE or ANYMULTIRANGE, throw error.
 *	  (We have no way to select a specific range type if the arguments don't
 *	  include ANYRANGE or ANYMULTIRANGE.)
 * 6) ANYENUM is treated the same as ANYELEMENT except that if it is used
 *	  (alone or in combination with plain ANYELEMENT), we add the extra
 *	  condition that the ANYELEMENT type must be an enum.
 * 7) ANYNONARRAY is treated the same as ANYELEMENT except that if it is used,
 *	  we add the extra condition that the ANYELEMENT type must not be an array.
 *	  (This is a no-op if used in combination with ANYARRAY or ANYENUM, but
 *	  is an extra restriction if not.)
 * 8) ANYCOMPATIBLE, ANYCOMPATIBLEARRAY, and ANYCOMPATIBLENONARRAY are handled
 *	  by resolving the common supertype of those arguments (or their element
 *	  types, for array inputs), and then coercing all those arguments to the
 *	  common supertype, or the array type over the common supertype for
 *	  ANYCOMPATIBLEARRAY.
 * 9) For ANYCOMPATIBLERANGE and ANYCOMPATIBLEMULTIRANGE, there must be at
 *	  least one non-UNKNOWN input matching those arguments, and all such
 *	  inputs must be the same range type (or its multirange type, as
 *	  appropriate), since we cannot deduce a range type from non-range types.
 *	  Furthermore, the range type's subtype is included while choosing the
 *	  common supertype for ANYCOMPATIBLE et al, and it must exactly match
 *	  that common supertype.
 *
 * Domains over arrays or ranges match ANYARRAY or ANYRANGE arguments,
 * respectively, and are immediately flattened to their base type.  (In
 * particular, if the return type is also ANYARRAY or ANYRANGE, we'll set
 * it to the base type not the domain type.)  The same is true for
 * ANYMULTIRANGE, ANYCOMPATIBLEARRAY, ANYCOMPATIBLERANGE, and
 * ANYCOMPATIBLEMULTIRANGE.
 *
 * When allow_poly is false, we are not expecting any of the actual_arg_types
 * to be polymorphic, and we should not return a polymorphic result type
 * either.  When allow_poly is true, it is okay to have polymorphic "actual"
 * arg types, and we can return a matching polymorphic type as the result.
 * (This case is currently used only to check compatibility of an aggregate's
 * declaration with the underlying transfn.)
 *
 * A special case is that we could see ANYARRAY as an actual_arg_type even
 * when allow_poly is false (this is possible only because pg_statistic has
 * columns shown as anyarray in the catalogs).  We allow this to match a
 * declared ANYARRAY argument, but only if there is no other polymorphic
 * argument that we would need to match it with, and no need to determine
 * the element type to infer the result type.  Note this means that functions
 * taking ANYARRAY had better behave sanely if applied to the pg_statistic
 * columns; they can't just assume that successive inputs are of the same
 * actual element type.  There is no similar logic for ANYCOMPATIBLEARRAY;
 * there isn't a need for it since there are no catalog columns of that type,
 * so we won't see it as input.  We could consider matching an actual ANYARRAY
 * input to an ANYCOMPATIBLEARRAY argument, but at present that seems useless
 * as well, since there's no value in using ANYCOMPATIBLEARRAY unless there's
 * at least one other ANYCOMPATIBLE-family argument or result.
 *
 * Also, if there are no arguments declared to be of polymorphic types,
 * we'll return the rettype unmodified even if it's polymorphic.  This should
 * never occur for user-declared functions, because CREATE FUNCTION prevents
 * it.  But it does happen for some built-in functions, such as array_in().
 */


/*
 * check_valid_polymorphic_signature()
 *		Is a proposed function signature valid per polymorphism rules?
 *
 * Returns NULL if the signature is valid (either ret_type is not polymorphic,
 * or it can be deduced from the given declared argument types).  Otherwise,
 * returns a palloc'd, already translated errdetail string saying why not.
 */


/*
 * check_valid_internal_signature()
 *		Is a proposed function signature valid per INTERNAL safety rules?
 *
 * Returns NULL if OK, or a suitable error message if ret_type is INTERNAL but
 * none of the declared arg types are.  (It's unsafe to create such a function
 * since it would allow invocation of INTERNAL-consuming functions directly
 * from SQL.)  It's overkill to return the error detail message, since there
 * is only one possibility, but we do it like this to keep the API similar to
 * check_valid_polymorphic_signature().
 */



/* TypeCategory()
 *		Assign a category to the specified type OID.
 *
 * NB: this must not return TYPCATEGORY_INVALID.
 */



/* IsPreferredType()
 *		Check if this type is a preferred type for the given category.
 *
 * If category is TYPCATEGORY_INVALID, then we'll return true for preferred
 * types of any category; otherwise, only for preferred types of that
 * category.
 */



/* IsBinaryCoercible()
 *		Check if srctype is binary-coercible to targettype.
 *
 * This notion allows us to cheat and directly exchange values without
 * going through the trouble of calling a conversion function.  Note that
 * in general, this should only be an implementation shortcut.  Before 7.4,
 * this was also used as a heuristic for resolving overloaded functions and
 * operators, but that's basically a bad idea.
 *
 * As of 7.3, binary coercibility isn't hardwired into the code anymore.
 * We consider two types binary-coercible if there is an implicitly
 * invokable, no-function-needed pg_cast entry.  Also, a domain is always
 * binary-coercible to its base type, though *not* vice versa (in the other
 * direction, one must apply domain constraint checks before accepting the
 * value as legitimate).  We also need to special-case various polymorphic
 * types.
 *
 * This function replaces IsBinaryCompatible(), which was an inherently
 * symmetric test.  Since the pg_cast entries aren't necessarily symmetric,
 * the order of the operands is now significant.
 */


/* IsBinaryCoercibleWithCast()
 *		Check if srctype is binary-coercible to targettype.
 *
 * This variant also returns the OID of the pg_cast entry if one is involved.
 * *castoid is set to InvalidOid if no binary-coercible cast exists, or if
 * there is a hard-wired rule for it rather than a pg_cast entry.
 */



/*
 * find_coercion_pathway
 *		Look for a coercion pathway between two types.
 *
 * Currently, this deals only with scalar-type cases; it does not consider
 * polymorphic types nor casts between composite types.  (Perhaps fold
 * those in someday?)
 *
 * ccontext determines the set of available casts.
 *
 * The possible result codes are:
 *	COERCION_PATH_NONE: failed to find any coercion pathway
 *				*funcid is set to InvalidOid
 *	COERCION_PATH_FUNC: apply the coercion function returned in *funcid
 *	COERCION_PATH_RELABELTYPE: binary-compatible cast, no function needed
 *				*funcid is set to InvalidOid
 *	COERCION_PATH_ARRAYCOERCE: need an ArrayCoerceExpr node
 *				*funcid is set to InvalidOid
 *	COERCION_PATH_COERCEVIAIO: need a CoerceViaIO node
 *				*funcid is set to InvalidOid
 *
 * Note: COERCION_PATH_RELABELTYPE does not necessarily mean that no work is
 * needed to do the coercion; if the target is a domain then we may need to
 * apply domain constraint checking.  If you want to check for a zero-effort
 * conversion then use IsBinaryCoercible().
 */



/*
 * find_typmod_coercion_function -- does the given type need length coercion?
 *
 * If the target type possesses a pg_cast function from itself to itself,
 * it must need length coercion.
 *
 * "bpchar" (ie, char(N)) and "numeric" are examples of such types.
 *
 * If the given type is a varlena array type, we do not look for a coercion
 * function associated directly with the array type, but instead look for
 * one associated with the element type.  An ArrayCoerceExpr node must be
 * used to apply such a function.  (Note: currently, it's pointless to
 * return the funcid in this case, because it'll just get looked up again
 * in the recursive construction of the ArrayCoerceExpr's elemexpr.)
 *
 * We use the same result enum as find_coercion_pathway, but the only possible
 * result codes are:
 *	COERCION_PATH_NONE: no length coercion needed
 *	COERCION_PATH_FUNC: apply the function returned in *funcid
 *	COERCION_PATH_ARRAYCOERCE: apply the function using ArrayCoerceExpr
 */


/*
 * is_complex_array
 *		Is this type an array of composite?
 *
 * Note: this will not return true for record[]; check for RECORDARRAYOID
 * separately if needed.
 */



/*
 * Check whether reltypeId is the row type of a typed table of type
 * reloftypeId, or is a domain over such a row type.  (This is conceptually
 * similar to the subtype relationship checked by typeInheritsFrom().)
 */

