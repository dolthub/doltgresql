/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - SearchSysCache1
 * - ReleaseSysCache
 * - SysCacheGetAttrNotNull
 * - GetSysCacheOid
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * syscache.c
 *	  System cache management routines
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 *
 * IDENTIFICATION
 *	  src/backend/utils/cache/syscache.c
 *
 * NOTES
 *	  These routines allow the parser/planner/executor to perform
 *	  rapid lookups on the contents of the system catalogs.
 *
 *	  see utils/syscache.h for a list of the cache IDs
 *
 *-------------------------------------------------------------------------
 */
#include "postgres.h"

#include "access/htup_details.h"
#include "catalog/pg_db_role_setting_d.h"
#include "catalog/pg_depend_d.h"
#include "catalog/pg_description_d.h"
#include "catalog/pg_seclabel_d.h"
#include "catalog/pg_shdepend_d.h"
#include "catalog/pg_shdescription_d.h"
#include "catalog/pg_shseclabel_d.h"
#include "common/int.h"
#include "lib/qunique.h"
#include "miscadmin.h"
#include "storage/lmgr.h"
#include "utils/catcache.h"
#include "utils/inval.h"
#include "utils/lsyscache.h"
#include "utils/rel.h"
#include "utils/syscache.h"

/*---------------------------------------------------------------------------

	Adding system caches:

	There must be a unique index underlying each syscache (ie, an index
	whose key is the same as that of the cache).  If there is not one
	already, add the definition for it to include/catalog/pg_*.h using
	DECLARE_UNIQUE_INDEX.
	(Adding an index requires a catversion.h update, while simply
	adding/deleting caches only requires a recompile.)

	Add a MAKE_SYSCACHE call to the same pg_*.h file specifying the name of
	your cache, the underlying index, and the initial number of hash buckets.

	The number of hash buckets must be a power of 2.  It's reasonable to
	set this to the number of entries that might be in the particular cache
	in a medium-size database.

	Finally, any place your relation gets heap_insert() or
	heap_update() calls, use CatalogTupleInsert() or CatalogTupleUpdate()
	instead, which also update indexes.  The heap_* calls do not do that.

*---------------------------------------------------------------------------
*/

/*
 *		struct cachedesc: information defining a single syscache
 */
struct cachedesc
{
	Oid			reloid;			/* OID of the relation being cached */
	Oid			indoid;			/* OID of index relation for this cache */
	int			nkeys;			/* # of keys needed for cache lookup */
	int			key[4];			/* attribute numbers of key attrs */
	int			nbuckets;		/* number of hash buckets for this cache */
};

/* Macro to provide nkeys and key array with convenient syntax. */
#define KEY(...) VA_ARGS_NARGS(__VA_ARGS__), { __VA_ARGS__ }

#include "catalog/syscache_info.h"

StaticAssertDecl(lengthof(cacheinfo) == SysCacheSize,
				 "SysCacheSize does not match syscache.c's array");





/* Sorted array of OIDs of tables that have caches on them */



/* Sorted array of OIDs of tables and indexes used by caches */



static int	oid_compare(const void *a, const void *b);


/*
 * InitCatalogCache - initialize the caches
 *
 * Note that no database access is done here; we only allocate memory
 * and initialize the cache structure.  Interrogation of the database
 * to complete initialization of a cache happens upon first use
 * of that cache.
 */


/*
 * InitCatalogCachePhase2 - finish initializing the caches
 *
 * Finish initializing all the caches, including necessary database
 * access.
 *
 * This is *not* essential; normally we allow syscaches to be initialized
 * on first use.  However, it is useful as a mechanism to preload the
 * relcache with entries for the most-commonly-used system catalogs.
 * Therefore, we invoke this routine when we need to write a new relcache
 * init file.
 */



/*
 * SearchSysCache
 *
 *	A layer on top of SearchCatCache that does the initialization and
 *	key-setting for you.
 *
 *	Returns the cache copy of the tuple if one is found, NULL if not.
 *	The tuple is the 'cache' copy and must NOT be modified!
 *
 *	When the caller is done using the tuple, call ReleaseSysCache()
 *	to release the reference count grabbed by SearchSysCache().  If this
 *	is not done, the tuple will remain locked in cache until end of
 *	transaction, which is tolerable but not desirable.
 *
 *	CAUTION: The tuple that is returned must NOT be freed by the caller!
 */


#include <catalog/pg_type.h>
#include <catalog/pg_collation_d.h>
#include "pg_query_pg_type.c"

HeapTuple
SearchSysCache1(int cacheId,
				Datum key1)
{
	/*Assert(cacheId >= 0 && cacheId < SysCacheSize &&
		   PointerIsValid(SysCache[cacheId]));
	Assert(SysCache[cacheId]->cc_nkeys == 1);

	return SearchCatCache1(SysCache[cacheId], key1);*/

	HeapTuple tuple;
	HeapTupleHeader td;
	Form_pg_type t = palloc0(sizeof(FormData_pg_type));
	const PgQueryBuiltinType *bt;
	Size		len,
				data_len;
	int			hoff;

    if (cacheId != TYPEOID)
        elog(ERROR, "Not implemented (SearchSysCache1 only supports TYPEOID cache (%d), got cache %d)", TYPEOID, cacheId);

    bt = pg_query_builtin_type_by_oid(DatumGetObjectId(key1));
    if (bt == NULL)
        elog(ERROR, "Not implemented (SearchSysCache1 got TYPEOID cache request for type OID %d)", DatumGetObjectId(key1));

    strlcpy(NameStr(t->typname), bt->typname, NAMEDATALEN);
    t->typlen = bt->typlen;
    t->typbyval = bt->typbyval;
    t->typtype = bt->typtype;
    t->typcategory = bt->typcategory;
    t->typalign = bt->typalign;
    t->typarray = bt->typarray;
    t->typcollation = bt->typcollation;

    t->oid = DatumGetObjectId(key1);
    t->typisdefined = true;

	// The following logic is copied from heap_form_tuple, but pretends there are no nulls, and copies t_data directly

	/*
	 * Determine total space needed
	 */
	len = offsetof(HeapTupleHeaderData, t_bits);

	//if (hasnull)
	//	len += BITMAPLEN(numberOfAttributes);

	hoff = len = MAXALIGN(len); /* align user data safely */

	//data_len = heap_compute_data_size(tupleDescriptor, values, isnull);
	data_len = MAXALIGN(sizeof(FormData_pg_type));

	len += data_len;

	/*
	 * Allocate and zero the space needed.  Note that the tuple body and
	 * HeapTupleData management structure are allocated in one chunk.
	 */
	tuple = (HeapTuple) palloc0(HEAPTUPLESIZE + len);
	tuple->t_data = td = (HeapTupleHeader) ((char *) tuple + HEAPTUPLESIZE);

	/*
	 * And fill in the information.  Note we fill the Datum fields even though
	 * this tuple may never become a Datum.  This lets HeapTupleHeaderGetDatum
	 * identify the tuple type if needed.
	 */
	tuple->t_len = len;
	ItemPointerSetInvalid(&(tuple->t_self));
	tuple->t_tableOid = InvalidOid;

	HeapTupleHeaderSetDatumLength(td, len);
	//HeapTupleHeaderSetTypeId(td, tupleDescriptor->tdtypeid);
	//HeapTupleHeaderSetTypMod(td, tupleDescriptor->tdtypmod);
	/* We also make sure that t_ctid is invalid unless explicitly set */
	ItemPointerSetInvalid(&(td->t_ctid));

	HeapTupleHeaderSetNatts(td, Natts_pg_type);
	td->t_hoff = hoff;

	/*heap_fill_tuple(tupleDescriptor,
					values,
					isnull,
					(char *) td + hoff,
					data_len,
					&td->t_infomask,
					(hasnull ? td->t_bits : NULL));*/
	memcpy((char *) td + hoff, t, sizeof(FormData_pg_type));

	return tuple;
}









/*
 * ReleaseSysCache
 *		Release previously grabbed reference count on a tuple
 */

void
ReleaseSysCache(HeapTuple tuple)
{
	/* Do nothing */
}

/*
 * SearchSysCacheLocked1
 *
 * Combine SearchSysCache1() with acquiring a LOCKTAG_TUPLE at mode
 * InplaceUpdateTupleLock.  This is a tool for complying with the
 * README.tuplock section "Locking to write inplace-updated tables".  After
 * the caller's heap_update(), it should UnlockTuple(InplaceUpdateTupleLock)
 * and ReleaseSysCache().
 *
 * The returned tuple may be the subject of an uncommitted update, so this
 * doesn't prevent the "tuple concurrently updated" error.
 */


/*
 * SearchSysCacheCopy
 *
 * A convenience routine that does SearchSysCache and (if successful)
 * returns a modifiable copy of the syscache entry.  The original
 * syscache entry is released before returning.  The caller should
 * heap_freetuple() the result when done with it.
 */


/*
 * SearchSysCacheLockedCopy1
 *
 * Meld SearchSysCacheLocked1 with SearchSysCacheCopy().  After the
 * caller's heap_update(), it should UnlockTuple(InplaceUpdateTupleLock) and
 * heap_freetuple().
 */


/*
 * SearchSysCacheExists
 *
 * A convenience routine that just probes to see if a tuple can be found.
 * No lock is retained on the syscache entry.
 */


/*
 * GetSysCacheOid
 *
 * A convenience routine that does SearchSysCache and returns the OID in the
 * oidcol column of the found tuple, or InvalidOid if no tuple could be found.
 * No lock is retained on the syscache entry.
 */
Oid
GetSysCacheOid(int cacheId,
			   AttrNumber oidcol,
			   Datum key1,
			   Datum key2,
			   Datum key3,
			   Datum key4)
{
	if (cacheId != TYPENAMENSP)
        elog(ERROR, "Not implemented (GetSysCacheOid only supports TYPENAMENSP cache (%d), got cache %d)", TYPENAMENSP, cacheId);

    if (oidcol != Anum_pg_type_oid)
        elog(ERROR, "Not implemented (GetSysCacheOid oidcol not as expected)");

    if (key3 != 0 || key4 != 0)
        elog(ERROR, "Not implemented (GetSysCacheOid key 3 and key 4 must be zero)");

    if (IsCatalogNamespace(DatumGetObjectId(key2)))
    {
        return pg_query_builtin_type_oid_by_name(DatumGetPointer(key1));
    }
    else if (DatumGetObjectId(key2) == PG_PUBLIC_NAMESPACE)
    {
        /*
         * For now, we assume that any unknown type in the public namespace
         * is a row type. That is not correct for extensions or custom types,
         * but requires a more invasive fix (e.g. by having the user pass in
         * custom type data) that is not yet implemented.
         */
        return RECORDOID;
    }
    else
    {
        elog(ERROR, "Not implemented (GetSysCacheOid only supported for built-in catalog types or custom pseudotypes in public namespace)");
    }
}




/*
 * SearchSysCacheAttName
 *
 * This routine is equivalent to SearchSysCache on the ATTNAME cache,
 * except that it will return NULL if the found attribute is marked
 * attisdropped.  This is convenient for callers that want to act as
 * though dropped attributes don't exist.
 */


/*
 * SearchSysCacheCopyAttName
 *
 * As above, an attisdropped-aware version of SearchSysCacheCopy.
 */


/*
 * SearchSysCacheExistsAttName
 *
 * As above, an attisdropped-aware version of SearchSysCacheExists.
 */



/*
 * SearchSysCacheAttNum
 *
 * This routine is equivalent to SearchSysCache on the ATTNUM cache,
 * except that it will return NULL if the found attribute is marked
 * attisdropped.  This is convenient for callers that want to act as
 * though dropped attributes don't exist.
 */


/*
 * SearchSysCacheCopyAttNum
 *
 * As above, an attisdropped-aware version of SearchSysCacheCopy.
 */



/*
 * SysCacheGetAttr
 *
 *		Given a tuple previously fetched by SearchSysCache(),
 *		extract a specific attribute.
 *
 * This is equivalent to using heap_getattr() on a tuple fetched
 * from a non-cached relation.  Usually, this is only used for attributes
 * that could be NULL or variable length; the fixed-size attributes in
 * a system table are accessed just by mapping the tuple onto the C struct
 * declarations from include/catalog/.
 *
 * As with heap_getattr(), if the attribute is of a pass-by-reference type
 * then a pointer into the tuple data area is returned --- the caller must
 * not modify or pfree the datum!
 *
 * Note: it is legal to use SysCacheGetAttr() with a cacheId referencing
 * a different cache for the same catalog the tuple was fetched from.
 */
#include "pg_query_proctup_attrs.h"

/*
 * libpg_query only ever invokes SysCacheGetAttr from inside
 * plpgsql_compile_callback (transitively via get_func_arg_info), and the
 * tuple it sees is the libpg_query-forged pg_proc HeapTuple embedded in a
 * ProcTupWithAttrs wrapper. We recover the wrapper from the tuple pointer
 * and read the Anum-indexed value the caller asked for.
 */
Datum
SysCacheGetAttr(int cacheId, HeapTuple tup, AttrNumber attributeNumber, bool *isNull)
{
	ProcTupWithAttrs *w;

	if (cacheId != PROCOID)
		elog(ERROR, "Not implemented (SysCacheGetAttr only supports PROCOID, got cache %d)", cacheId);

	if (attributeNumber < 1 || attributeNumber > Natts_pg_proc)
		elog(ERROR, "SysCacheGetAttr: invalid pg_proc attribute number %d", attributeNumber);

	w = (ProcTupWithAttrs *) ((char *) tup - offsetof(ProcTupWithAttrs, tup));

	*isNull = w->nulls[attributeNumber - 1];
	return w->values[attributeNumber - 1];
}



/*
 * SysCacheGetAttrNotNull
 *
 * As above, a version of SysCacheGetAttr which knows that the attr cannot
 * be NULL.
 */
/*
 * Thin wrapper over the libpg_query SysCacheGetAttr mock: assert the value
 * isn't NULL and return its Datum.
 */
Datum
SysCacheGetAttrNotNull(int cacheId, HeapTuple tup, AttrNumber attributeNumber)
{
	bool		isNull;
	Datum		d = SysCacheGetAttr(cacheId, tup, attributeNumber, &isNull);

	if (isNull)
		elog(ERROR, "unexpected null attribute %d from pg_proc tuple", attributeNumber);
	return d;
}



/*
 * GetSysCacheHashValue
 *
 * Get the hash value that would be used for a tuple in the specified cache
 * with the given search keys.
 *
 * The reason for exposing this as part of the API is that the hash value is
 * exposed in cache invalidation operations, so there are places outside the
 * catcache code that need to be able to compute the hash values.
 */


/*
 * List-search interface
 */


/*
 * SysCacheInvalidate
 *
 *	Invalidate entries in the specified cache, given a hash value.
 *	See CatCacheInvalidate() for more info.
 *
 *	This routine is only quasi-public: it should only be used by inval.c.
 */


/*
 * Certain relations that do not have system caches send snapshot invalidation
 * messages in lieu of catcache messages.  This is for the benefit of
 * GetCatalogSnapshot(), which can then reuse its existing MVCC snapshot
 * for scanning one of those catalogs, rather than taking a new one, if no
 * invalidation has been received.
 *
 * Relations that have syscaches need not (and must not) be listed here.  The
 * catcache invalidation messages will also flush the snapshot.  If you add a
 * syscache for one of these relations, remove it from this list.
 */


/*
 * Test whether a relation has a system cache.
 */


/*
 * Test whether a relation supports a system cache, ie it is either a
 * cached table or the index used for a cache.
 */



/*
 * OID comparator for qsort
 */

