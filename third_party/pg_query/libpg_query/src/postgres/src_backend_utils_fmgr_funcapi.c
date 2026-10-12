/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - get_func_arg_info
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * funcapi.c
 *	  Utility and convenience functions for fmgr functions that return
 *	  sets and/or composite types, or deal with VARIADIC inputs.
 *
 * Copyright (c) 2002-2025, PostgreSQL Global Development Group
 *
 * IDENTIFICATION
 *	  src/backend/utils/fmgr/funcapi.c
 *
 *-------------------------------------------------------------------------
 */
#include "postgres.h"

#include "access/htup_details.h"
#include "access/relation.h"
#include "catalog/namespace.h"
#include "catalog/pg_proc.h"
#include "catalog/pg_type.h"
#include "funcapi.h"
#include "miscadmin.h"
#include "nodes/nodeFuncs.h"
#include "utils/array.h"
#include "utils/builtins.h"
#include "utils/lsyscache.h"
#include "utils/memutils.h"
#include "utils/regproc.h"
#include "utils/rel.h"
#include "utils/syscache.h"
#include "utils/tuplestore.h"
#include "utils/typcache.h"


typedef struct polymorphic_actuals
{
	Oid			anyelement_type;	/* anyelement mapping, if known */
	Oid			anyarray_type;	/* anyarray mapping, if known */
	Oid			anyrange_type;	/* anyrange mapping, if known */
	Oid			anymultirange_type; /* anymultirange mapping, if known */
} polymorphic_actuals;

static void shutdown_MultiFuncCall(Datum arg);
static TypeFuncClass internal_get_result_type(Oid funcid,
											  Node *call_expr,
											  ReturnSetInfo *rsinfo,
											  Oid *resultTypeId,
											  TupleDesc *resultTupleDesc);
static void resolve_anyelement_from_others(polymorphic_actuals *actuals);
static void resolve_anyarray_from_others(polymorphic_actuals *actuals);
static void resolve_anyrange_from_others(polymorphic_actuals *actuals);
static void resolve_anymultirange_from_others(polymorphic_actuals *actuals);
static bool resolve_polymorphic_tupdesc(TupleDesc tupdesc,
										oidvector *declared_args,
										Node *call_expr);
static TypeFuncClass get_type_func_class(Oid typid, Oid *base_typeid);


/*
 * InitMaterializedSRF
 *
 * Helper function to build the state of a set-returning function used
 * in the context of a single call with materialize mode.  This code
 * includes sanity checks on ReturnSetInfo, creates the Tuplestore and
 * the TupleDesc used with the function and stores them into the
 * function's ReturnSetInfo.
 *
 * "flags" can be set to MAT_SRF_USE_EXPECTED_DESC, to use the tuple
 * descriptor coming from expectedDesc, which is the tuple descriptor
 * expected by the caller.  MAT_SRF_BLESS can be set to complete the
 * information associated to the tuple descriptor, which is necessary
 * in some cases where the tuple descriptor comes from a transient
 * RECORD datatype.
 */



/*
 * init_MultiFuncCall
 * Create an empty FuncCallContext data structure
 * and do some other basic Multi-function call setup
 * and error checking
 */


/*
 * per_MultiFuncCall
 *
 * Do Multi-function per-call setup
 */


/*
 * end_MultiFuncCall
 * Clean up after init_MultiFuncCall
 */


/*
 * shutdown_MultiFuncCall
 * Shutdown function to clean up after init_MultiFuncCall
 */



/*
 * get_call_result_type
 *		Given a function's call info record, determine the kind of datatype
 *		it is supposed to return.  If resultTypeId isn't NULL, *resultTypeId
 *		receives the actual datatype OID (this is mainly useful for scalar
 *		result types).  If resultTupleDesc isn't NULL, *resultTupleDesc
 *		receives a pointer to a TupleDesc when the result is of a composite
 *		type, or NULL when it's a scalar result.
 *
 * One hard case that this handles is resolution of actual rowtypes for
 * functions returning RECORD (from either the function's OUT parameter
 * list, or a ReturnSetInfo context node).  TYPEFUNC_RECORD is returned
 * only when we couldn't resolve the actual rowtype for lack of information.
 *
 * The other hard case that this handles is resolution of polymorphism.
 * We will never return polymorphic pseudotypes (ANYELEMENT etc), either
 * as a scalar result type or as a component of a rowtype.
 *
 * This function is relatively expensive --- in a function returning set,
 * try to call it only the first time through.
 */


/*
 * get_expr_result_type
 *		As above, but work from a calling expression node tree
 *
 * Beware of using this on the funcexpr of a RTE that has a coldeflist.
 * The correct conclusion in such cases is always that the function returns
 * RECORD with the columns defined by the coldeflist fields (funccolnames etc).
 * If it does not, it's the executor's responsibility to catch the discrepancy
 * at runtime; but code processing the query in advance of that point might
 * come to inconsistent conclusions if it checks the actual expression.
 */


/*
 * get_func_result_type
 *		As above, but work from a function's OID only
 *
 * This will not be able to resolve pure-RECORD results nor polymorphism.
 */


/*
 * internal_get_result_type -- workhorse code implementing all the above
 *
 * funcid must always be supplied.  call_expr and rsinfo can be NULL if not
 * available.  We will return TYPEFUNC_RECORD, and store NULL into
 * *resultTupleDesc, if we cannot deduce the complete result rowtype from
 * the available information.
 */


/*
 * get_expr_result_tupdesc
 *		Get a tupdesc describing the result of a composite-valued expression
 *
 * If expression is not composite or rowtype can't be determined, returns NULL
 * if noError is true, else throws error.
 *
 * This is a simpler version of get_expr_result_type() for use when the caller
 * is only interested in determinate rowtype results.  As with that function,
 * beware of using this on the funcexpr of a RTE that has a coldeflist.
 */


/*
 * Resolve actual type of ANYELEMENT from other polymorphic inputs
 *
 * Note: the error cases here and in the sibling functions below are not
 * really user-facing; they could only occur if the function signature is
 * incorrect or the parser failed to enforce consistency of the actual
 * argument types.  Hence, we don't sweat too much over the error messages.
 */


/*
 * Resolve actual type of ANYARRAY from other polymorphic inputs
 */


/*
 * Resolve actual type of ANYRANGE from other polymorphic inputs
 */


/*
 * Resolve actual type of ANYMULTIRANGE from other polymorphic inputs
 */


/*
 * Given the result tuple descriptor for a function with OUT parameters,
 * replace any polymorphic column types (ANYELEMENT etc) in the tupdesc
 * with concrete data types deduced from the input arguments.
 * declared_args is an oidvector of the function's declared input arg types
 * (showing which are polymorphic), and call_expr is the call expression.
 *
 * Returns true if able to deduce all types, false if necessary information
 * is not provided (call_expr is NULL or arg types aren't identifiable).
 */


/*
 * Given the declared argument types and modes for a function, replace any
 * polymorphic types (ANYELEMENT etc) in argtypes[] with concrete data types
 * deduced from the input arguments found in call_expr.
 *
 * Returns true if able to deduce all types, false if necessary information
 * is not provided (call_expr is NULL or arg types aren't identifiable).
 *
 * This is the same logic as resolve_polymorphic_tupdesc, but with a different
 * argument representation, and slightly different output responsibilities.
 *
 * argmodes may be NULL, in which case all arguments are assumed to be IN mode.
 */


/*
 * get_type_func_class
 *		Given the type OID, obtain its TYPEFUNC classification.
 *		Also, if it's a domain, return the base type OID.
 *
 * This is intended to centralize a bunch of formerly ad-hoc code for
 * classifying types.  The categories used here are useful for deciding
 * how to handle functions returning the datatype.
 */



/*
 * get_func_arg_info
 *
 * Fetch info about the argument types, names, and IN/OUT modes from the
 * pg_proc tuple.  Return value is the total number of arguments.
 * Other results are palloc'd.  *p_argtypes is always filled in, but
 * *p_argnames and *p_argmodes will be set NULL in the default cases
 * (no names, and all IN arguments, respectively).
 *
 * Note that this function simply fetches what is in the pg_proc tuple;
 * it doesn't do any interpretation of polymorphic types.
 */
/*
 * libpg_query version of get_func_arg_info that pulls argument info from the
 * libpg_query-forged pg_proc tuple via the SysCacheGetAttr mock. The body
 * matches upstream's funcapi.c logic; only the surrounding context (tuple
 * source) differs.
 */
int
get_func_arg_info(HeapTuple procTup,
				  Oid **p_argtypes, char ***p_argnames, char **p_argmodes)
{
	Form_pg_proc procStruct = (Form_pg_proc) GETSTRUCT(procTup);
	Datum		proallargtypes;
	Datum		proargmodes;
	Datum		proargnames;
	bool		isNull;
	ArrayType  *arr;
	int			numargs;
	Datum	   *elems;
	int			nelems;
	int			i;

	proallargtypes = SysCacheGetAttr(PROCOID, procTup,
									 Anum_pg_proc_proallargtypes,
									 &isNull);
	if (!isNull)
	{
		arr = DatumGetArrayTypeP(proallargtypes);
		numargs = ARR_DIMS(arr)[0];
		if (ARR_NDIM(arr) != 1 ||
			numargs < 0 ||
			ARR_HASNULL(arr) ||
			ARR_ELEMTYPE(arr) != OIDOID)
			elog(ERROR, "proallargtypes is not a 1-D Oid array or it contains nulls");
		*p_argtypes = (Oid *) palloc(numargs * sizeof(Oid));
		memcpy(*p_argtypes, ARR_DATA_PTR(arr),
			   numargs * sizeof(Oid));
	}
	else
	{
		numargs = procStruct->proargtypes.dim1;
		*p_argtypes = (Oid *) palloc(numargs * sizeof(Oid));
		memcpy(*p_argtypes, procStruct->proargtypes.values,
			   numargs * sizeof(Oid));
	}

	proargnames = SysCacheGetAttr(PROCOID, procTup,
								  Anum_pg_proc_proargnames,
								  &isNull);
	if (isNull)
		*p_argnames = NULL;
	else
	{
		deconstruct_array_builtin(DatumGetArrayTypeP(proargnames), TEXTOID,
								  &elems, NULL, &nelems);
		if (nelems != numargs)
			elog(ERROR, "proargnames must have the same number of elements as the function has arguments");
		*p_argnames = (char **) palloc(sizeof(char *) * numargs);
		for (i = 0; i < numargs; i++)
			(*p_argnames)[i] = TextDatumGetCString(elems[i]);
	}

	proargmodes = SysCacheGetAttr(PROCOID, procTup,
								  Anum_pg_proc_proargmodes,
								  &isNull);
	if (isNull)
		*p_argmodes = NULL;
	else
	{
		arr = DatumGetArrayTypeP(proargmodes);
		if (ARR_NDIM(arr) != 1 ||
			ARR_DIMS(arr)[0] != numargs ||
			ARR_HASNULL(arr) ||
			ARR_ELEMTYPE(arr) != CHAROID)
			elog(ERROR, "proargmodes is not a 1-D char array of length %d or it contains nulls",
				 numargs);
		*p_argmodes = (char *) palloc(numargs * sizeof(char));
		memcpy(*p_argmodes, ARR_DATA_PTR(arr),
			   numargs * sizeof(char));
	}

	return numargs;
}



/*
 * get_func_trftypes
 *
 * Returns the number of transformed types used by the function.
 * If there are any, a palloc'd array of the type OIDs is returned
 * into *p_trftypes.
 */


/*
 * get_func_input_arg_names
 *
 * Extract the names of input arguments only, given a function's
 * proargnames and proargmodes entries in Datum form.
 *
 * Returns the number of input arguments, which is the length of the
 * palloc'd array returned to *arg_names.  Entries for unnamed args
 * are set to NULL.  You don't get anything if proargnames is NULL.
 */



/*
 * get_func_result_name
 *
 * If the function has exactly one output parameter, and that parameter
 * is named, return the name (as a palloc'd string).  Else return NULL.
 *
 * This is used to determine the default output column name for functions
 * returning scalar types.
 */



/*
 * build_function_result_tupdesc_t
 *
 * Given a pg_proc row for a function, return a tuple descriptor for the
 * result rowtype, or NULL if the function does not have OUT parameters.
 *
 * Note that this does not handle resolution of polymorphic types;
 * that is deliberate.
 */


/*
 * build_function_result_tupdesc_d
 *
 * Build a RECORD function's tupledesc from the pg_proc proallargtypes,
 * proargmodes, and proargnames arrays.  This is split out for the
 * convenience of ProcedureCreate, which needs to be able to compute the
 * tupledesc before actually creating the function.
 *
 * For functions (but not for procedures), returns NULL if there are not at
 * least two OUT or INOUT arguments.
 */



/*
 * RelationNameGetTupleDesc
 *
 * Given a (possibly qualified) relation name, build a TupleDesc.
 *
 * Note: while this works as advertised, it's seldom the best way to
 * build a tupdesc for a function's result type.  It's kept around
 * only for backwards compatibility with existing user-written code.
 */


/*
 * TypeGetTupleDesc
 *
 * Given a type Oid, build a TupleDesc.  (In most cases you should be
 * using get_call_result_type or one of its siblings instead of this
 * routine, so that you can handle OUT parameters, RECORD result type,
 * and polymorphic results.)
 *
 * If the type is composite, *and* a colaliases List is provided, *and*
 * the List is of natts length, use the aliases instead of the relation
 * attnames.  (NB: this usage is deprecated since it may result in
 * creation of unnecessary transient record types.)
 *
 * If the type is a base type, a single item alias List is required.
 */


/*
 * extract_variadic_args
 *
 * Extract a set of argument values, types and NULL markers for a given
 * input function which makes use of a VARIADIC input whose argument list
 * depends on the caller context. When doing a VARIADIC call, the caller
 * has provided one argument made of an array of values, so deconstruct the
 * array data before using it for the next processing. If no VARIADIC call
 * is used, just fill in the status data based on all the arguments given
 * by the caller.
 *
 * This function returns the number of arguments generated, or -1 in the
 * case of "VARIADIC NULL".
 */

