/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - get_namespace_name_or_temp
 * - get_namespace_name
 * - get_array_type
 * - type_is_rowtype
 * - get_typtype
 * - getBaseType
 * - getBaseTypeAndTypmod
 * - get_element_type
 * - get_base_element_type
 * - type_is_rowtype
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * lsyscache.c
 *	  Convenience routines for common queries in the system catalog cache.
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 * IDENTIFICATION
 *	  src/backend/utils/cache/lsyscache.c
 *
 * NOTES
 *	  Eventually, the index information should go through here, too.
 *-------------------------------------------------------------------------
 */
#include "postgres.h"

#include "access/hash.h"
#include "access/htup_details.h"
#include "bootstrap/bootstrap.h"
#include "catalog/namespace.h"
#include "catalog/pg_am.h"
#include "catalog/pg_amop.h"
#include "catalog/pg_amproc.h"
#include "catalog/pg_cast.h"
#include "catalog/pg_class.h"
#include "catalog/pg_collation.h"
#include "catalog/pg_constraint.h"
#include "catalog/pg_index.h"
#include "catalog/pg_language.h"
#include "catalog/pg_namespace.h"
#include "catalog/pg_opclass.h"
#include "catalog/pg_opfamily.h"
#include "catalog/pg_operator.h"
#include "catalog/pg_proc.h"
#include "catalog/pg_publication.h"
#include "catalog/pg_range.h"
#include "catalog/pg_statistic.h"
#include "catalog/pg_subscription.h"
#include "catalog/pg_transform.h"
#include "catalog/pg_type.h"
#include "miscadmin.h"
#include "nodes/makefuncs.h"
#include "utils/array.h"
#include "utils/builtins.h"
#include "utils/catcache.h"
#include "utils/datum.h"
#include "utils/fmgroids.h"
#include "utils/lsyscache.h"
#include "utils/syscache.h"
#include "utils/typcache.h"

/* Hook for plugins to get control in get_attavgwidth() */



/*				---------- AMOP CACHES ----------						 */

/*
 * op_in_opfamily
 *
 *		Return t iff operator 'opno' is in operator family 'opfamily'.
 *
 * This function only considers search operators, not ordering operators.
 */


/*
 * get_op_opfamily_strategy
 *
 *		Get the operator's strategy number within the specified opfamily,
 *		or 0 if it's not a member of the opfamily.
 *
 * This function only considers search operators, not ordering operators.
 */


/*
 * get_op_opfamily_sortfamily
 *
 *		If the operator is an ordering operator within the specified opfamily,
 *		return its amopsortfamily OID; else return InvalidOid.
 */


/*
 * get_op_opfamily_properties
 *
 *		Get the operator's strategy number and declared input data types
 *		within the specified opfamily.
 *
 * Caller should already have verified that opno is a member of opfamily,
 * therefore we raise an error if the tuple is not found.
 */


/*
 * get_opfamily_member
 *		Get the OID of the operator that implements the specified strategy
 *		with the specified datatypes for the specified opfamily.
 *
 * Returns InvalidOid if there is no pg_amop entry for the given keys.
 */


/*
 * get_opfamily_member_for_cmptype
 *		Get the OID of the operator that implements the specified comparison
 *		type with the specified datatypes for the specified opfamily.
 *
 * Returns InvalidOid if there is no mapping for the comparison type or no
 * pg_amop entry for the given keys.
 */


/*
 * get_opmethod_canorder
 *		Return amcanorder field for given index AM.
 *
 * To speed things up in the common cases, we're hardcoding the results from
 * the built-in index types.  Note that we also need to hardcode the negative
 * results from the built-in non-btree index types, since you'll usually get a
 * few hits for those as well.  It would be nice to organize and cache this a
 * bit differently to avoid the hardcoding.
 */


/*
 * get_ordering_op_properties
 *		Given the OID of an ordering operator (a "<" or ">" operator),
 *		determine its opfamily, its declared input datatype, and its
 *		comparison type.
 *
 * Returns true if successful, false if no matching pg_amop entry exists.
 * (This indicates that the operator is not a valid ordering operator.)
 *
 * Note: the operator could be registered in multiple families, for example
 * if someone were to build a "reverse sort" opfamily.  This would result in
 * uncertainty as to whether "ORDER BY USING op" would default to NULLS FIRST
 * or NULLS LAST, as well as inefficient planning due to failure to match up
 * pathkeys that should be the same.  So we want a determinate result here.
 * Because of the way the syscache search works, we'll use the interpretation
 * associated with the opfamily with smallest OID, which is probably
 * determinate enough.  Since there is no longer any particularly good reason
 * to build reverse-sort opfamilies, it doesn't seem worth expending any
 * additional effort on ensuring consistency.
 */


/*
 * get_equality_op_for_ordering_op
 *		Get the OID of the datatype-specific equality operator
 *		associated with an ordering operator (a "<" or ">" operator).
 *
 * If "reverse" isn't NULL, also set *reverse to false if the operator is "<",
 * true if it's ">"
 *
 * Returns InvalidOid if no matching equality operator can be found.
 * (This indicates that the operator is not a valid ordering operator.)
 */


/*
 * get_ordering_op_for_equality_op
 *		Get the OID of a datatype-specific "less than" ordering operator
 *		associated with an equality operator.  (If there are multiple
 *		possibilities, assume any one will do.)
 *
 * This function is used when we have to sort data before unique-ifying,
 * and don't much care which sorting op is used as long as it's compatible
 * with the intended equality operator.  Since we need a sorting operator,
 * it should be single-data-type even if the given operator is cross-type.
 * The caller specifies whether to find an op for the LHS or RHS data type.
 *
 * Returns InvalidOid if no matching ordering operator can be found.
 */


/*
 * get_mergejoin_opfamilies
 *		Given a putatively mergejoinable operator, return a list of the OIDs
 *		of the amcanorder opfamilies in which it represents equality.
 *
 * It is possible (though at present unusual) for an operator to be equality
 * in more than one opfamily, hence the result is a list.  This also lets us
 * return NIL if the operator is not found in any opfamilies.
 *
 * The planner currently uses simple equal() tests to compare the lists
 * returned by this function, which makes the list order relevant, though
 * strictly speaking it should not be.  Because of the way syscache list
 * searches are handled, in normal operation the result will be sorted by OID
 * so everything works fine.  If running with system index usage disabled,
 * the result ordering is unspecified and hence the planner might fail to
 * recognize optimization opportunities ... but that's hardly a scenario in
 * which performance is good anyway, so there's no point in expending code
 * or cycles here to guarantee the ordering in that case.
 */


/*
 * get_compatible_hash_operators
 *		Get the OID(s) of hash equality operator(s) compatible with the given
 *		operator, but operating on its LHS and/or RHS datatype.
 *
 * An operator for the LHS type is sought and returned into *lhs_opno if
 * lhs_opno isn't NULL.  Similarly, an operator for the RHS type is sought
 * and returned into *rhs_opno if rhs_opno isn't NULL.
 *
 * If the given operator is not cross-type, the results should be the same
 * operator, but in cross-type situations they will be different.
 *
 * Returns true if able to find the requested operator(s), false if not.
 * (This indicates that the operator should not have been marked oprcanhash.)
 *
 * Callers must beware that for container types (arrays, records, ranges)
 * this function will succeed for array_eq etc, but the hash function could
 * fail at runtime if the contained type(s) are not hashable.  If it is
 * possible that the operator is one of these, precheck with op_hashjoinable
 * or get_op_hash_functions_ext.
 */


/*
 * get_op_hash_functions
 *		Get the OID(s) of the standard hash support function(s) compatible with
 *		the given operator, operating on its LHS and/or RHS datatype as required.
 *
 * A function for the LHS type is sought and returned into *lhs_procno if
 * lhs_procno isn't NULL.  Similarly, a function for the RHS type is sought
 * and returned into *rhs_procno if rhs_procno isn't NULL.
 *
 * If the given operator is not cross-type, the results should be the same
 * function, but in cross-type situations they will be different.
 *
 * Returns true if able to find the requested function(s), false if not.
 * (This indicates that the operator should not have been marked oprcanhash.)
 *
 * Callers must beware that for container types (arrays, records, ranges)
 * this function will succeed for array_eq etc, but the hash function could
 * fail at runtime if the contained type(s) are not hashable.  If it is
 * possible that the operator is one of these, use get_op_hash_functions_ext
 * or precheck with op_hashjoinable.
 */


/*
 * get_op_hash_functions_ext
 *		As above, but verify hashability in container-type cases.
 *
 * As with op_hashjoinable, assume the left input type is sufficient
 * to disambiguate container-type cases.
 */


/*
 * get_op_index_interpretation
 *		Given an operator's OID, find out which amcanorder opfamilies it belongs to,
 *		and what properties it has within each one.  The results are returned
 *		as a palloc'd list of OpIndexInterpretation structs.
 *
 * In addition to the normal btree operators, we consider a <> operator to be
 * a "member" of an opfamily if its negator is an equality operator of the
 * opfamily.  COMPARE_NE is returned as the strategy number for this case.
 */


/*
 * equality_ops_are_compatible
 *		Return true if the two given operators have compatible equality
 *		semantics.
 *
 * This is trivially true if they are the same operator.  Otherwise,
 * we look to see if they both belong to an opfamily that guarantees
 * compatible semantics for equality.  Either finding allows us to assume
 * that they have compatible notions of equality.
 *
 * The typical use is to compare two equality operators (for instance the
 * cross-type operators int24eq vs int4eq), but the test is meaningful for
 * any pair of operators in a btree/hash opfamily.  Btree marks its
 * opfamilies as amconsistentequality, which guarantees that every member
 * of the family (=, <, <=, >, >=) agrees on the equivalence relation
 * defined by the family's "=".  So a non-equality operator and an
 * equality operator from the same opfamily are also "compatible" in this
 * sense.
 */


/*
 * comparison_ops_are_compatible
 *		Return true if the two given comparison operators have compatible
 *		semantics.
 *
 * This is trivially true if they are the same operator.  Otherwise, we look
 * to see if they both belong to an opfamily that guarantees compatible
 * semantics for ordering.  (For example, for btree, '<' and '>=' ops match if
 * they belong to the same family.)
 *
 * (This is identical to equality_ops_are_compatible(), except that we check
 * amconsistentordering instead of amconsistentequality.)
 */


/*
 * collations_agree_on_equality
 *		Return true if the two collations have equivalent notions of equality,
 *		so that a uniqueness or equality proof established under one side
 *		carries over to a comparison performed under the other side.
 *
 * Note: this is equality compatibility only.  Do NOT use this to reason
 * about ordering.
 *
 * An InvalidOid on either side denotes the absence of a collation -- that
 * side's operation is not collation-sensitive (e.g. a non-collatable column
 * type).  Absence of a collation cannot conflict with the other side's
 * collation, so we treat such pairs as agreeing on equality.  This generalizes
 * the asymmetric treatment in IndexCollMatchesExprColl().
 *
 * Otherwise the collations have equivalent equality if they match, or if both
 * are deterministic: by definition a deterministic collation treats two
 * strings as equal iff they are byte-wise equal (see CREATE COLLATION), so any
 * two deterministic collations share the same equality relation.  A mismatch
 * involving a nondeterministic collation, however, may mean the two equality
 * relations disagree, and the proof is unsound.
 */


/*
 * op_is_safe_index_member
 *		Check if the operator is a member of a B-tree or Hash operator family.
 *
 * Membership in such an opfamily has several useful implications: the operator
 * returns non-null for non-null inputs (i.e. "null-safety", required so that
 * the operator doesn't break index integrity), and it agrees with other
 * members of the same opfamily on equality semantics.  Callers use this check
 * as a proxy for any of those properties.
 */



/*				---------- AMPROC CACHES ----------						 */

/*
 * get_opfamily_proc
 *		Get the OID of the specified support function
 *		for the specified opfamily and datatypes.
 *
 * Returns InvalidOid if there is no pg_amproc entry for the given keys.
 */



/*				---------- ATTRIBUTE CACHES ----------					 */

/*
 * get_attname
 *		Given the relation id and the attribute number, return the "attname"
 *		field from the attribute relation as a palloc'ed string.
 *
 * If no such attribute exists and missing_ok is true, NULL is returned;
 * otherwise a not-intended-for-user-consumption error is thrown.
 */


/*
 * get_attnum
 *
 *		Given the relation id and the attribute name,
 *		return the "attnum" field from the attribute relation.
 *
 *		Returns InvalidAttrNumber if the attr doesn't exist (or is dropped).
 */


/*
 * get_attgenerated
 *
 *		Given the relation id and the attribute number,
 *		return the "attgenerated" field from the attribute relation.
 *
 *		Errors if not found.
 *
 *		Since not generated is represented by '\0', this can also be used as a
 *		Boolean test.
 */


/*
 * get_atttype
 *
 *		Given the relation OID and the attribute number with the relation,
 *		return the attribute type OID.
 */


/*
 * get_atttypetypmodcoll
 *
 *		A three-fer: given the relation id and the attribute number,
 *		fetch atttypid, atttypmod, and attcollation in a single cache lookup.
 *
 * Unlike the otherwise-similar get_atttype, this routine
 * raises an error if it can't obtain the information.
 */


/*
 * get_attoptions
 *
 *		Given the relation id and the attribute number,
 *		return the attribute options text[] datum, if any.
 */


/*				---------- PG_CAST CACHE ----------					 */

/*
 * get_cast_oid - given two type OIDs, look up a cast OID
 *
 * If missing_ok is false, throw an error if the cast is not found.  If
 * true, just return InvalidOid.
 */


/*				---------- COLLATION CACHE ----------					 */

/*
 * get_collation_name
 *		Returns the name of a given pg_collation entry.
 *
 * Returns a palloc'd copy of the string, or NULL if no such collation.
 *
 * NOTE: since collation name is not unique, be wary of code that uses this
 * for anything except preparing error messages.
 */




/*				---------- CONSTRAINT CACHE ----------					 */

/*
 * get_constraint_name
 *		Returns the name of a given pg_constraint entry.
 *
 * Returns a palloc'd copy of the string, or NULL if no such constraint.
 *
 * NOTE: since constraint name is not unique, be wary of code that uses this
 * for anything except preparing error messages.
 */


/*
 * get_constraint_index
 *		Given the OID of a unique, primary-key, or exclusion constraint,
 *		return the OID of the underlying index.
 *
 * Returns InvalidOid if the constraint could not be found or is of
 * the wrong type.
 *
 * The intent of this function is to return the index "owned" by the
 * specified constraint.  Therefore we must check contype, since some
 * pg_constraint entries (e.g. for foreign-key constraints) store the
 * OID of an index that is referenced but not owned by the constraint.
 */


/*
 * get_constraint_type
 *		Return the pg_constraint.contype value for the given constraint.
 *
 * No frills.
 */


/*				---------- LANGUAGE CACHE ----------					 */



/*				---------- OPCLASS CACHE ----------						 */

/*
 * get_opclass_family
 *
 *		Returns the OID of the operator family the opclass belongs to.
 */


/*
 * get_opclass_input_type
 *
 *		Returns the OID of the datatype the opclass indexes.
 */


/*
 * get_opclass_opfamily_and_input_type
 *
 *		Returns the OID of the operator family the opclass belongs to,
 *				the OID of the datatype the opclass indexes
 */


/*
 * get_opclass_method
 *
 *		Returns the OID of the index access method the opclass belongs to.
 */


/*				---------- OPFAMILY CACHE ----------					 */

/*
 * get_opfamily_method
 *
 *		Returns the OID of the index access method the opfamily is for.
 */




/*				---------- OPERATOR CACHE ----------					 */

/*
 * get_opcode
 *
 *		Returns the regproc id of the routine used to implement an
 *		operator given the operator oid.
 */


/*
 * get_opname
 *	  returns the name of the operator with the given opno
 *
 * Note: returns a palloc'd copy of the string, or NULL if no such operator.
 */


/*
 * get_op_rettype
 *		Given operator oid, return the operator's result type.
 */


/*
 * op_input_types
 *
 *		Returns the left and right input datatypes for an operator
 *		(InvalidOid if not relevant).
 */


/*
 * op_mergejoinable
 *
 * Returns true if the operator is potentially mergejoinable.  (The planner
 * will fail to find any mergejoin plans unless there are suitable btree
 * opfamily entries for this operator and associated sortops.  The pg_operator
 * flag is just a hint to tell the planner whether to bother looking.)
 *
 * In some cases (currently only array_eq and record_eq), mergejoinability
 * depends on the specific input data type the operator is invoked for, so
 * that must be passed as well. We currently assume that only one input's type
 * is needed to check this --- by convention, pass the left input's data type.
 */


/*
 * op_hashjoinable
 *
 * Returns true if the operator is hashjoinable.  (There must be a suitable
 * hash opfamily entry for this operator if it is so marked.)
 *
 * In some cases (currently array_eq, record_eq, range_eq, multirange_eq),
 * hashjoinability depends on the specific input data type the operator is
 * invoked for, so that must be passed as well.  We currently assume that only
 * one input's type is needed to check this --- by convention, pass the left
 * input's data type.
 */


/*
 * op_strict
 *
 * Get the proisstrict flag for the operator's underlying function.
 */


/*
 * op_volatile
 *
 * Get the provolatile flag for the operator's underlying function.
 */


/*
 * get_commutator
 *
 *		Returns the corresponding commutator of an operator.
 */


/*
 * get_negator
 *
 *		Returns the corresponding negator of an operator.
 */


/*
 * get_oprrest
 *
 *		Returns procedure id for computing selectivity of an operator.
 */


/*
 * get_oprjoin
 *
 *		Returns procedure id for computing selectivity of a join.
 */


/*				---------- FUNCTION CACHE ----------					 */

/*
 * get_func_name
 *	  returns the name of the function with the given funcid
 *
 * Note: returns a palloc'd copy of the string, or NULL if no such function.
 */


/*
 * get_func_namespace
 *
 *		Returns the pg_namespace OID associated with a given function.
 */


/*
 * get_func_rettype
 *		Given procedure id, return the function's result type.
 */


/*
 * get_func_nargs
 *		Given procedure id, return the number of arguments.
 */


/*
 * get_func_signature
 *		Given procedure id, return the function's argument and result types.
 *		(The return value is the result type.)
 *
 * The arguments are returned as a palloc'd array.
 */


/*
 * get_func_variadictype
 *		Given procedure id, return the function's provariadic field.
 */


/*
 * get_func_retset
 *		Given procedure id, return the function's proretset flag.
 */


/*
 * func_strict
 *		Given procedure id, return the function's proisstrict flag.
 */


/*
 * func_volatile
 *		Given procedure id, return the function's provolatile flag.
 */


/*
 * func_parallel
 *		Given procedure id, return the function's proparallel flag.
 */


/*
 * get_func_prokind
 *	   Given procedure id, return the routine kind.
 */


/*
 * get_func_leakproof
 *	   Given procedure id, return the function's leakproof field.
 */


/*
 * get_func_support
 *
 *		Returns the support function OID associated with a given function,
 *		or InvalidOid if there is none.
 */


/*				---------- RELATION CACHE ----------					 */

/*
 * get_relname_relid
 *		Given name and namespace of a relation, look up the OID.
 *
 * Returns InvalidOid if there is no such relation.
 */


#ifdef NOT_USED
/*
 * get_relnatts
 *
 *		Returns the number of attributes for a given relation.
 */
int
get_relnatts(Oid relid)
{
	HeapTuple	tp;

	tp = SearchSysCache1(RELOID, ObjectIdGetDatum(relid));
	if (HeapTupleIsValid(tp))
	{
		Form_pg_class reltup = (Form_pg_class) GETSTRUCT(tp);
		int			result;

		result = reltup->relnatts;
		ReleaseSysCache(tp);
		return result;
	}
	else
		return InvalidAttrNumber;
}
#endif

/*
 * get_rel_name
 *		Returns the name of a given relation.
 *
 * Returns a palloc'd copy of the string, or NULL if no such relation.
 *
 * NOTE: since relation name is not unique, be wary of code that uses this
 * for anything except preparing error messages.
 */


/*
 * get_rel_namespace
 *
 *		Returns the pg_namespace OID associated with a given relation.
 */


/*
 * get_rel_type_id
 *
 *		Returns the pg_type OID associated with a given relation.
 *
 * Note: not all pg_class entries have associated pg_type OIDs; so be
 * careful to check for InvalidOid result.
 */


/*
 * get_rel_relkind
 *
 *		Returns the relkind associated with a given relation.
 */


/*
 * get_rel_relispartition
 *
 *		Returns the relispartition flag associated with a given relation.
 */


/*
 * get_rel_tablespace
 *
 *		Returns the pg_tablespace OID associated with a given relation.
 *
 * Note: InvalidOid might mean either that we couldn't find the relation,
 * or that it is in the database's default tablespace.
 */


/*
 * get_rel_persistence
 *
 *		Returns the relpersistence associated with a given relation.
 */


/*
 * get_rel_relam
 *
 *		Returns the relam associated with a given relation.
 */



/*				---------- TRANSFORM CACHE ----------						 */






/*				---------- TYPE CACHE ----------						 */

/*
 * get_typisdefined
 *
 *		Given the type OID, determine whether the type is defined
 *		(if not, it's only a shell).
 */


/*
 * get_typlen
 *
 *		Given the type OID, return the length of the type.
 */


/*
 * get_typbyval
 *
 *		Given the type OID, determine whether the type is returned by value or
 *		not.  Returns true if by value, false if by reference.
 */


/*
 * get_typlenbyval
 *
 *		A two-fer: given the type OID, return both typlen and typbyval.
 *
 *		Since both pieces of info are needed to know how to copy a Datum,
 *		many places need both.  Might as well get them with one cache lookup
 *		instead of two.  Also, this routine raises an error instead of
 *		returning a bogus value when given a bad type OID.
 */


/*
 * get_typlenbyvalalign
 *
 *		A three-fer: given the type OID, return typlen, typbyval, typalign.
 */


/*
 * getTypeIOParam
 *		Given a pg_type row, select the type OID to pass to I/O functions
 *
 * Formerly, all I/O functions were passed pg_type.typelem as their second
 * parameter, but we now have a more complex rule about what to pass.
 * This knowledge is intended to be centralized here --- direct references
 * to typelem elsewhere in the code are wrong, if they are associated with
 * I/O calls and not with actual subscripting operations!  (But see
 * bootstrap.c's boot_get_type_io_data() if you need to change this.)
 *
 * As of PostgreSQL 8.1, output functions receive only the value itself
 * and not any auxiliary parameters, so the name of this routine is now
 * a bit of a misnomer ... it should be getTypeInputParam.
 */


/*
 * get_type_io_data
 *
 *		A six-fer:	given the type OID, return typlen, typbyval, typalign,
 *					typdelim, typioparam, and IO function OID. The IO function
 *					returned is controlled by IOFuncSelector
 */


#ifdef NOT_USED
char
get_typalign(Oid typid)
{
	HeapTuple	tp;

	tp = SearchSysCache1(TYPEOID, ObjectIdGetDatum(typid));
	if (HeapTupleIsValid(tp))
	{
		Form_pg_type typtup = (Form_pg_type) GETSTRUCT(tp);
		char		result;

		result = typtup->typalign;
		ReleaseSysCache(tp);
		return result;
	}
	else
		return TYPALIGN_INT;
}
#endif



/*
 * get_typdefault
 *	  Given a type OID, return the type's default value, if any.
 *
 *	  The result is a palloc'd expression node tree, or NULL if there
 *	  is no defined default for the datatype.
 *
 * NB: caller should be prepared to coerce result to correct datatype;
 * the returned expression tree might produce something of the wrong type.
 */


/*
 * getBaseType
 *		If the given type is a domain, return its base type;
 *		otherwise return the type's own OID.
 */
Oid
getBaseType(Oid typid)
{
	int32		typmod = -1;

	return getBaseTypeAndTypmod(typid, &typmod);
}

/*
 * getBaseTypeAndTypmod
 *		If the given type is a domain, return its base type and typmod;
 *		otherwise return the type's own OID, and leave *typmod unchanged.
 *
 * Note that the "applied typmod" should be -1 for every domain level
 * above the bottommost; therefore, if the passed-in typid is indeed
 * a domain, *typmod should be -1.
 */
Oid
getBaseTypeAndTypmod(Oid typid, int32 *typmod)
{
	/*
	 * We loop to find the bottom base type in a stack of domains.
	 */
	for (;;)
	{
		HeapTuple	tup;
		Form_pg_type typTup;

		tup = SearchSysCache1(TYPEOID, ObjectIdGetDatum(typid));
		if (!HeapTupleIsValid(tup))
			elog(ERROR, "cache lookup failed for type %u", typid);
		typTup = (Form_pg_type) GETSTRUCT(tup);
		if (typTup->typtype != TYPTYPE_DOMAIN)
		{
			/* Not a domain, so done */
			ReleaseSysCache(tup);
			break;
		}

		Assert(*typmod == -1);
		typid = typTup->typbasetype;
		*typmod = typTup->typtypmod;

		ReleaseSysCache(tup);
	}

	return typid;
}

/*
 * get_typavgwidth
 *
 *	  Given a type OID and a typmod value (pass -1 if typmod is unknown),
 *	  estimate the average width of values of the type.  This is used by
 *	  the planner, which doesn't require absolutely correct results;
 *	  it's OK (and expected) to guess if we don't know for sure.
 */


/*
 * get_typtype
 *
 *		Given the type OID, find if it is a basic type, a complex type, etc.
 *		It returns the null char if the cache lookup fails...
 */
char
get_typtype(Oid typid)
{
	HeapTuple	tp;

	tp = SearchSysCache1(TYPEOID, ObjectIdGetDatum(typid));
	if (HeapTupleIsValid(tp))
	{
		Form_pg_type typtup = (Form_pg_type) GETSTRUCT(tp);
		char		result;

		result = typtup->typtype;
		ReleaseSysCache(tp);
		return result;
	}
	else
		return '\0';
}

/*
 * type_is_rowtype
 *
 *		Convenience function to determine whether a type OID represents
 *		a "rowtype" type --- either RECORD or a named composite type
 *		(including a domain over a named composite type).
 */
bool
type_is_rowtype(Oid typid)
{
	if (typid == RECORDOID)
		return true;			/* easy case */
	switch (get_typtype(typid))
	{
		case TYPTYPE_COMPOSITE:
			return true;
		case TYPTYPE_DOMAIN:
			if (get_typtype(getBaseType(typid)) == TYPTYPE_COMPOSITE)
				return true;
			break;
		default:
			break;
	}
	return false;
}

/*
 * type_is_enum
 *	  Returns true if the given type is an enum type.
 */


/*
 * type_is_range
 *	  Returns true if the given type is a range type.
 */


/*
 * type_is_multirange
 *	  Returns true if the given type is a multirange type.
 */


/*
 * get_type_category_preferred
 *
 *		Given the type OID, fetch its category and preferred-type status.
 *		Throws error on failure.
 */


/*
 * get_typ_typrelid
 *
 *		Given the type OID, get the typrelid (InvalidOid if not a complex
 *		type).
 */


/*
 * get_element_type
 *
 *		Given the type OID, get the typelem (InvalidOid if not an array type).
 *
 * NB: this only succeeds for "true" arrays having array_subscript_handler
 * as typsubscript.  For other types, InvalidOid is returned independently
 * of whether they have typelem or typsubscript set.
 */
Oid
get_element_type(Oid typid)
{
	HeapTuple	tp;

	tp = SearchSysCache1(TYPEOID, ObjectIdGetDatum(typid));
	if (HeapTupleIsValid(tp))
	{
		Form_pg_type typtup = (Form_pg_type) GETSTRUCT(tp);
		Oid			result;

		if (IsTrueArrayType(typtup))
			result = typtup->typelem;
		else
			result = InvalidOid;
		ReleaseSysCache(tp);
		return result;
	}
	else
		return InvalidOid;
}

/*
 * get_array_type
 *
 *		Given the type OID, get the corresponding "true" array type.
 *		Returns InvalidOid if no array type can be found.
 */
Oid
get_array_type(Oid typid)
{
	HeapTuple	tp;
	Oid			result = InvalidOid;

	tp = SearchSysCache1(TYPEOID, ObjectIdGetDatum(typid));
	if (HeapTupleIsValid(tp))
	{
		result = ((Form_pg_type) GETSTRUCT(tp))->typarray;
		ReleaseSysCache(tp);
	}
	return result;
}

/*
 * get_promoted_array_type
 *
 *		The "promoted" type is what you'd get from an ARRAY(SELECT ...)
 *		construct, that is, either the corresponding "true" array type
 *		if the input is a scalar type that has such an array type,
 *		or the same type if the input is already a "true" array type.
 *		Returns InvalidOid if neither rule is satisfied.
 */


/*
 * get_base_element_type
 *		Given the type OID, get the typelem, looking "through" any domain
 *		to its underlying array type.
 *
 * This is equivalent to get_element_type(getBaseType(typid)), but avoids
 * an extra cache lookup.  Note that it fails to provide any information
 * about the typmod of the array.
 */
Oid
get_base_element_type(Oid typid)
{
	/*
	 * We loop to find the bottom base type in a stack of domains.
	 */
	for (;;)
	{
		HeapTuple	tup;
		Form_pg_type typTup;

		tup = SearchSysCache1(TYPEOID, ObjectIdGetDatum(typid));
		if (!HeapTupleIsValid(tup))
			break;
		typTup = (Form_pg_type) GETSTRUCT(tup);
		if (typTup->typtype != TYPTYPE_DOMAIN)
		{
			/* Not a domain, so stop descending */
			Oid			result;

			/* This test must match get_element_type */
			if (IsTrueArrayType(typTup))
				result = typTup->typelem;
			else
				result = InvalidOid;
			ReleaseSysCache(tup);
			return result;
		}

		typid = typTup->typbasetype;
		ReleaseSysCache(tup);
	}

	/* Like get_element_type, silently return InvalidOid for bogus input */
	return InvalidOid;
}

/*
 * getTypeInputInfo
 *
 *		Get info needed for converting values of a type to internal form
 */


/*
 * getTypeOutputInfo
 *
 *		Get info needed for printing values of a type
 */


/*
 * getTypeBinaryInputInfo
 *
 *		Get info needed for binary input of values of a type
 */


/*
 * getTypeBinaryOutputInfo
 *
 *		Get info needed for binary output of values of a type
 */


/*
 * get_typmodin
 *
 *		Given the type OID, return the type's typmodin procedure, if any.
 */


#ifdef NOT_USED
/*
 * get_typmodout
 *
 *		Given the type OID, return the type's typmodout procedure, if any.
 */
Oid
get_typmodout(Oid typid)
{
	HeapTuple	tp;

	tp = SearchSysCache1(TYPEOID, ObjectIdGetDatum(typid));
	if (HeapTupleIsValid(tp))
	{
		Form_pg_type typtup = (Form_pg_type) GETSTRUCT(tp);
		Oid			result;

		result = typtup->typmodout;
		ReleaseSysCache(tp);
		return result;
	}
	else
		return InvalidOid;
}
#endif							/* NOT_USED */

/*
 * get_typcollation
 *
 *		Given the type OID, return the type's typcollation attribute.
 */



/*
 * type_is_collatable
 *
 *		Return whether the type cares about collations
 */



/*
 * get_typsubscript
 *
 *		Given the type OID, return the type's subscripting handler's OID,
 *		if it has one.
 *
 * If typelemp isn't NULL, we also store the type's typelem value there.
 * This saves some callers an extra catalog lookup.
 */


/*
 * getSubscriptingRoutines
 *
 *		Given the type OID, fetch the type's subscripting methods struct.
 *		Return NULL if type is not subscriptable.
 *
 * If typelemp isn't NULL, we also store the type's typelem value there.
 * This saves some callers an extra catalog lookup.
 */



/*				---------- STATISTICS CACHE ----------					 */

/*
 * get_attavgwidth
 *
 *	  Given the table and attribute number of a column, get the average
 *	  width of entries in the column.  Return zero if no data available.
 *
 * Currently this is only consulted for individual tables, not for inheritance
 * trees, so we don't need an "inh" parameter.
 *
 * Calling a hook at this point looks somewhat strange, but is required
 * because the optimizer calls this function without any other way for
 * plug-ins to control the result.
 */


/*
 * get_attstatsslot
 *
 *		Extract the contents of a "slot" of a pg_statistic tuple.
 *		Returns true if requested slot type was found, else false.
 *
 * Unlike other routines in this file, this takes a pointer to an
 * already-looked-up tuple in the pg_statistic cache.  We do this since
 * most callers will want to extract more than one value from the cache
 * entry, and we don't want to repeat the cache lookup unnecessarily.
 * Also, this API allows this routine to be used with statistics tuples
 * that have been provided by a stats hook and didn't really come from
 * pg_statistic.
 *
 * sslot: pointer to output area (typically, a local variable in the caller).
 * statstuple: pg_statistic tuple to be examined.
 * reqkind: STAKIND code for desired statistics slot kind.
 * reqop: STAOP value wanted, or InvalidOid if don't care.
 * flags: bitmask of ATTSTATSSLOT_VALUES and/or ATTSTATSSLOT_NUMBERS.
 *
 * If a matching slot is found, true is returned, and *sslot is filled thus:
 * staop: receives the actual STAOP value.
 * stacoll: receives the actual STACOLL value.
 * valuetype: receives actual datatype of the elements of stavalues.
 * values: receives pointer to an array of the slot's stavalues.
 * nvalues: receives number of stavalues.
 * numbers: receives pointer to an array of the slot's stanumbers (as float4).
 * nnumbers: receives number of stanumbers.
 *
 * valuetype/values/nvalues are InvalidOid/NULL/0 if ATTSTATSSLOT_VALUES
 * wasn't specified.  Likewise, numbers/nnumbers are NULL/0 if
 * ATTSTATSSLOT_NUMBERS wasn't specified.
 *
 * If no matching slot is found, false is returned, and *sslot is zeroed.
 *
 * Note that the current API doesn't allow for searching for a slot with
 * a particular collation.  If we ever actually support recording more than
 * one collation, we'll have to extend the API, but for now simple is good.
 *
 * The data referred to by the fields of sslot is locally palloc'd and
 * is independent of the original pg_statistic tuple.  When the caller
 * is done with it, call free_attstatsslot to release the palloc'd data.
 *
 * If it's desirable to call free_attstatsslot when get_attstatsslot might
 * not have been called, memset'ing sslot to zeroes will allow that.
 *
 * Passing flags=0 can be useful to quickly check if the requested slot type
 * exists.  In this case no arrays are extracted, so free_attstatsslot need
 * not be called.
 */


/*
 * free_attstatsslot
 *		Free data allocated by get_attstatsslot
 */


/*				---------- PG_NAMESPACE CACHE ----------				 */

/*
 * get_namespace_name
 *		Returns the name of a given namespace
 *
 * Returns a palloc'd copy of the string, or NULL if no such namespace.
 */
char *
get_namespace_name(Oid nspid)
{
	HeapTuple	tp;

	tp = SearchSysCache1(NAMESPACEOID, ObjectIdGetDatum(nspid));
	if (HeapTupleIsValid(tp))
	{
		Form_pg_namespace nsptup = (Form_pg_namespace) GETSTRUCT(tp);
		char	   *result;

		result = pstrdup(NameStr(nsptup->nspname));
		ReleaseSysCache(tp);
		return result;
	}
	else
		return NULL;
}

/*
 * get_namespace_name_or_temp
 *		As above, but if it is this backend's temporary namespace, return
 *		"pg_temp" instead.
 */
char *
get_namespace_name_or_temp(Oid nspid)
{
	if (isTempNamespace(nspid))
		return pstrdup("pg_temp");
	else
		return get_namespace_name(nspid);
}

/*				---------- PG_RANGE CACHES ----------				 */

/*
 * get_range_subtype
 *		Returns the subtype of a given range type
 *
 * Returns InvalidOid if the type is not a range type.
 */


/*
 * get_range_collation
 *		Returns the collation of a given range type
 *
 * Returns InvalidOid if the type is not a range type,
 * or if its subtype is not collatable.
 */


/*
 * get_range_multirange
 *		Returns the multirange type of a given range type
 *
 * Returns InvalidOid if the type is not a range type.
 */


/*
 * get_multirange_range
 *		Returns the range type of a given multirange
 *
 * Returns InvalidOid if the type is not a multirange.
 */


/*				---------- PG_INDEX CACHE ----------				 */

/*
 * get_index_column_opclass
 *
 *		Given the index OID and column number,
 *		return opclass of the index column
 *			or InvalidOid if the index was not found
 *				or column is non-key one.
 */


/*
 * get_index_isreplident
 *
 *		Given the index OID, return pg_index.indisreplident.
 */


/*
 * get_index_isvalid
 *
 *		Given the index OID, return pg_index.indisvalid.
 */


/*
 * get_index_isclustered
 *
 *		Given the index OID, return pg_index.indisclustered.
 */


/*
 * get_publication_oid - given a publication name, look up the OID
 *
 * If missing_ok is false, throw an error if name not found.  If true, just
 * return InvalidOid.
 */


/*
 * get_publication_name - given a publication Oid, look up the name
 *
 * If missing_ok is false, throw an error if name not found.  If true, just
 * return NULL.
 */


/*
 * get_subscription_oid - given a subscription name, look up the OID
 *
 * If missing_ok is false, throw an error if name not found.  If true, just
 * return InvalidOid.
 */


/*
 * get_subscription_name - given a subscription OID, look up the name
 *
 * If missing_ok is false, throw an error if name not found.  If true, just
 * return NULL.
 */

