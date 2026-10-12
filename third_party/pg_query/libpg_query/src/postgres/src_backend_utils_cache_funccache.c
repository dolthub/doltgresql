/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - cfunc_resolve_polymorphic_argtypes
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * funccache.c
 *	  Function cache management.
 *
 * funccache.c manages a cache of function execution data.  The cache
 * is used by SQL-language and PL/pgSQL functions, and could be used by
 * other function languages.  Each cache entry is specific to the execution
 * of a particular function (identified by OID) with specific input data
 * types; so a polymorphic function could have many associated cache entries.
 * Trigger functions similarly have a cache entry per trigger.  These rules
 * allow the cached data to be specific to the particular data types the
 * function call will be dealing with.
 *
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 * IDENTIFICATION
 *	  src/backend/utils/cache/funccache.c
 *
 *-------------------------------------------------------------------------
 */
#include "postgres.h"

#include "catalog/pg_proc.h"
#include "commands/event_trigger.h"
#include "commands/trigger.h"
#include "common/hashfn.h"
#include "funcapi.h"
#include "utils/funccache.h"
#include "utils/hsearch.h"
#include "utils/syscache.h"


/*
 * Hash table for cached functions
 */


typedef struct CachedFunctionHashEntry
{
	CachedFunctionHashKey key;	/* hash key, must be first */
	CachedFunction *function;	/* points to data of language-specific size */
} CachedFunctionHashEntry;

#define FUNCS_PER_USER		128 /* initial table size */

static uint32 cfunc_hash(const void *key, Size keysize);
static int	cfunc_match(const void *key1, const void *key2, Size keysize);


/*
 * Initialize the hash table on first use.
 *
 * The hash table will be in TopMemoryContext regardless of caller's context.
 */


/*
 * cfunc_hash: hash function for cfunc hash table
 *
 * We need special hash and match functions to deal with the optional
 * presence of a TupleDesc in the hash keys.  As long as we have to do
 * that, we might as well also be smart about not comparing unused
 * elements of the argtypes arrays.
 */


/*
 * cfunc_match: match function to use with cfunc_hash
 */


/*
 * Look up the CachedFunction for the given hash key.
 * Returns NULL if not present.
 */


/*
 * Insert a hash table entry.
 */


/*
 * Delete a hash table entry.
 */


/*
 * Compute the hashkey for a given function invocation
 *
 * The hashkey is returned into the caller-provided storage at *hashkey.
 * Note however that if a callResultType is incorporated, we've not done
 * anything about copying that.
 */


/*
 * This is the same as the standard resolve_polymorphic_argtypes() function,
 * except that:
 * 1. We go ahead and report the error if we can't resolve the types.
 * 2. We treat RECORD-type input arguments (not output arguments) as if
 *    they were polymorphic, replacing their types with the actual input
 *    types if we can determine those.  This allows us to create a separate
 *    function cache entry for each named composite type passed to such an
 *    argument.
 * 3. In validation mode, we have no inputs to look at, so assume that
 *    polymorphic arguments are integer, integer-array or integer-range.
 */
/*
 * libpg_query never has a real fn_expr to consult, and only invokes the
 * plpgsql compile path in validator mode. Implement the validator branch
 * of upstream's cfunc_resolve_polymorphic_argtypes (mapping polymorphic
 * types to INT4OID family) and bail out for anything else.
 */
void
cfunc_resolve_polymorphic_argtypes(int numargs, Oid *argtypes, char *argmodes,
								   Node *call_expr, bool forValidator,
								   const char *proname)
{
	int			i;

	if (!forValidator)
		elog(ERROR, "Not implemented (cfunc_resolve_polymorphic_argtypes outside validator mode)");

	for (i = 0; i < numargs; i++)
	{
		switch (argtypes[i])
		{
			case ANYELEMENTOID:
			case ANYNONARRAYOID:
			case ANYENUMOID:
			case ANYCOMPATIBLEOID:
			case ANYCOMPATIBLENONARRAYOID:
				argtypes[i] = INT4OID;
				break;
			case ANYARRAYOID:
			case ANYCOMPATIBLEARRAYOID:
				argtypes[i] = INT4ARRAYOID;
				break;
			case ANYRANGEOID:
			case ANYCOMPATIBLERANGEOID:
				argtypes[i] = INT4RANGEOID;
				break;
			case ANYMULTIRANGEOID:
				argtypes[i] = INT4MULTIRANGEOID;
				break;
			default:
				break;
		}
	}
}



/*
 * delete_function - clean up as much as possible of a stale function cache
 *
 * We can't release the CachedFunction struct itself, because of the
 * possibility that there are fn_extra pointers to it.  We can release
 * the subsidiary storage, but only if there are no active evaluations
 * in progress.  Otherwise we'll just leak that storage.  Since the
 * case would only occur if a pg_proc update is detected during a nested
 * recursive call on the function, a leak seems acceptable.
 *
 * Note that this can be called more than once if there are multiple fn_extra
 * pointers to the same function cache.  Hence be careful not to do things
 * twice.
 */


/*
 * Compile a cached function, if no existing cache entry is suitable.
 *
 * fcinfo is the current call information.
 *
 * function should be NULL or the result of a previous call of
 * cached_function_compile() for the same fcinfo.  The caller will
 * typically save the result in fcinfo->flinfo->fn_extra, or in a
 * field of a struct pointed to by fn_extra, to re-use in later
 * calls within the same query.
 *
 * ccallback and dcallback are function-language-specific callbacks to
 * compile and delete a cached function entry.  dcallback can be NULL
 * if there's nothing for it to do.
 *
 * cacheEntrySize is the function-language-specific size of the cache entry
 * (which embeds a CachedFunction struct and typically has many more fields
 * after that).
 *
 * If includeResultType is true and the function returns composite,
 * include the actual result descriptor in the cache lookup key.
 *
 * If forValidator is true, we're only compiling for validation purposes,
 * and so some checks are skipped.
 *
 * Note: it's important for this to fall through quickly if the function
 * has already been compiled.
 *
 * Note: this function leaves the "use_count" field as zero.  The caller
 * is expected to increment the use_count and decrement it when done with
 * the cache entry.
 */

