/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - CreateTemplateTupleDesc
 * - TupleDescInitEntry
 * - populate_compact_attribute
 * - populate_compact_attribute_internal
 * - TupleDescInitEntryCollation
 * - CreateTemplateTupleDesc
 * - TupleDescInitEntry
 * - TupleDescInitEntryCollation
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * tupdesc.c
 *	  POSTGRES tuple descriptor support code
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 *
 * IDENTIFICATION
 *	  src/backend/access/common/tupdesc.c
 *
 * NOTES
 *	  some of the executor utility code such as "ExecTypeFromTL" should be
 *	  moved here.
 *
 *-------------------------------------------------------------------------
 */

#include "postgres.h"

#include "access/htup_details.h"
#include "access/toast_compression.h"
#include "access/tupdesc_details.h"
#include "catalog/catalog.h"
#include "catalog/pg_collation.h"
#include "catalog/pg_type.h"
#include "common/hashfn.h"
#include "utils/builtins.h"
#include "utils/datum.h"
#include "utils/resowner.h"
#include "utils/syscache.h"

/* ResourceOwner callbacks to hold tupledesc references  */
static void ResOwnerReleaseTupleDesc(Datum res);
static char *ResOwnerPrintTupleDesc(Datum res);



/* Convenience wrappers over ResourceOwnerRemember/Forget */




/*
 * populate_compact_attribute_internal
 *		Helper function for populate_compact_attribute()
 */
static inline void
populate_compact_attribute_internal(Form_pg_attribute src,
									CompactAttribute *dst)
{
	memset(dst, 0, sizeof(CompactAttribute));

	dst->attcacheoff = -1;
	dst->attlen = src->attlen;

	dst->attbyval = src->attbyval;
	dst->attispackable = (src->attstorage != TYPSTORAGE_PLAIN);
	dst->atthasmissing = src->atthasmissing;
	dst->attisdropped = src->attisdropped;
	dst->attgenerated = (src->attgenerated != '\0');

	/*
	 * Assign nullability status for this column.  Assuming that a constraint
	 * exists, at this point we don't know if a not-null constraint is valid,
	 * so we assign UNKNOWN unless the table is a catalog, in which case we
	 * know it's valid.
	 */
	dst->attnullability = !src->attnotnull ? ATTNULLABLE_UNRESTRICTED :
		IsCatalogRelationOid(src->attrelid) ? ATTNULLABLE_VALID :
		ATTNULLABLE_UNKNOWN;

	switch (src->attalign)
	{
		case TYPALIGN_INT:
			dst->attalignby = ALIGNOF_INT;
			break;
		case TYPALIGN_CHAR:
			dst->attalignby = sizeof(char);
			break;
		case TYPALIGN_DOUBLE:
			dst->attalignby = ALIGNOF_DOUBLE;
			break;
		case TYPALIGN_SHORT:
			dst->attalignby = ALIGNOF_SHORT;
			break;
		default:
			dst->attalignby = 0;
			elog(ERROR, "invalid attalign value: %c", src->attalign);
			break;
	}
}

/*
 * populate_compact_attribute
 *		Fill in the corresponding CompactAttribute element from the
 *		Form_pg_attribute for the given attribute number.  This must be called
 *		whenever a change is made to a Form_pg_attribute in the TupleDesc.
 */
void
populate_compact_attribute(TupleDesc tupdesc, int attnum)
{
	Form_pg_attribute src = TupleDescAttr(tupdesc, attnum);
	CompactAttribute *dst;

	/*
	 * Don't use TupleDescCompactAttr to prevent infinite recursion in assert
	 * builds.
	 */
	dst = &tupdesc->compact_attrs[attnum];

	populate_compact_attribute_internal(src, dst);
}

/*
 * verify_compact_attribute
 *		In Assert enabled builds, we verify that the CompactAttribute is
 *		populated correctly.  This helps find bugs in places such as ALTER
 *		TABLE where code makes changes to the FormData_pg_attribute but
 *		forgets to call populate_compact_attribute().
 *
 * This is used in TupleDescCompactAttr(), but declared here to allow access
 * to populate_compact_attribute_internal().
 */
#ifdef USE_ASSERT_CHECKING
#endif

/*
 * CreateTemplateTupleDesc
 *		This function allocates an empty tuple descriptor structure.
 *
 * Tuple type ID information is initially set for an anonymous record type;
 * caller can overwrite this if needed.
 */
TupleDesc
CreateTemplateTupleDesc(int natts)
{
	TupleDesc	desc;

	/*
	 * sanity checks
	 */
	Assert(natts >= 0);

	/*
	 * Allocate enough memory for the tuple descriptor, the CompactAttribute
	 * array and also an array of FormData_pg_attribute.
	 *
	 * Note: the FormData_pg_attribute array stride is
	 * sizeof(FormData_pg_attribute), since we declare the array elements as
	 * FormData_pg_attribute for notational convenience.  However, we only
	 * guarantee that the first ATTRIBUTE_FIXED_PART_SIZE bytes of each entry
	 * are valid; most code that copies tupdesc entries around copies just
	 * that much.  In principle that could be less due to trailing padding,
	 * although with the current definition of pg_attribute there probably
	 * isn't any padding.
	 */
	desc = (TupleDesc) palloc(offsetof(struct TupleDescData, compact_attrs) +
							  natts * sizeof(CompactAttribute) +
							  natts * sizeof(FormData_pg_attribute));

	/*
	 * Initialize other fields of the tupdesc.
	 */
	desc->natts = natts;
	desc->constr = NULL;
	desc->tdtypeid = RECORDOID;
	desc->tdtypmod = -1;
	desc->tdrefcount = -1;		/* assume not reference-counted */

	return desc;
}

/*
 * CreateTupleDesc
 *		This function allocates a new TupleDesc by copying a given
 *		Form_pg_attribute array.
 *
 * Tuple type ID information is initially set for an anonymous record type;
 * caller can overwrite this if needed.
 */


/*
 * CreateTupleDescCopy
 *		This function creates a new TupleDesc by copying from an existing
 *		TupleDesc.
 *
 * !!! Constraints and defaults are not copied !!!
 */


/*
 * CreateTupleDescTruncatedCopy
 *		This function creates a new TupleDesc with only the first 'natts'
 *		attributes from an existing TupleDesc
 *
 * !!! Constraints and defaults are not copied !!!
 */


/*
 * CreateTupleDescCopyConstr
 *		This function creates a new TupleDesc by copying from an existing
 *		TupleDesc (including its constraints and defaults).
 */


/*
 * TupleDescCopy
 *		Copy a tuple descriptor into caller-supplied memory.
 *		The memory may be shared memory mapped at any address, and must
 *		be sufficient to hold TupleDescSize(src) bytes.
 *
 * !!! Constraints and defaults are not copied !!!
 */


/*
 * TupleDescCopyEntry
 *		This function copies a single attribute structure from one tuple
 *		descriptor to another.
 *
 * !!! Constraints and defaults are not copied !!!
 */


/*
 * Free a TupleDesc including all substructure
 */


/*
 * Increment the reference count of a tupdesc, and log the reference in
 * CurrentResourceOwner.
 *
 * Do not apply this to tupdescs that are not being refcounted.  (Use the
 * macro PinTupleDesc for tupdescs of uncertain status.)
 */


/*
 * Decrement the reference count of a tupdesc, remove the corresponding
 * reference from CurrentResourceOwner, and free the tupdesc if no more
 * references remain.
 *
 * Do not apply this to tupdescs that are not being refcounted.  (Use the
 * macro ReleaseTupleDesc for tupdescs of uncertain status.)
 */


/*
 * Compare two TupleDesc structures for logical equality
 */


/*
 * equalRowTypes
 *
 * This determines whether two tuple descriptors have equal row types.  This
 * only checks those fields in pg_attribute that are applicable for row types,
 * while ignoring those fields that define the physical row storage or those
 * that define table column metadata.
 *
 * Specifically, this checks:
 *
 * - same number of attributes
 * - same composite type ID (but could both be zero)
 * - corresponding attributes (in order) have same the name, type, typmod,
 *   collation
 *
 * This is used to check whether two record types are compatible, whether
 * function return row types are the same, and other similar situations.
 *
 * (XXX There was some discussion whether attndims should be checked here, but
 * for now it has been decided not to.)
 *
 * Note: We deliberately do not check the tdtypmod field.  This allows
 * typcache.c to use this routine to see if a cached record type matches a
 * requested type.
 */


/*
 * hashRowType
 *
 * If two tuple descriptors would be considered equal by equalRowTypes()
 * then their hash value will be equal according to this function.
 */


/*
 * TupleDescInitEntry
 *		This function initializes a single attribute structure in
 *		a previously allocated tuple descriptor.
 *
 * If attributeName is NULL, the attname field is set to an empty string
 * (this is for cases where we don't know or need a name for the field).
 * Also, some callers use this function to change the datatype-related fields
 * in an existing tupdesc; they pass attributeName = NameStr(att->attname)
 * to indicate that the attname field shouldn't be modified.
 *
 * Note that attcollation is set to the default for the specified datatype.
 * If a nondefault collation is needed, insert it afterwards using
 * TupleDescInitEntryCollation.
 */
void
TupleDescInitEntry(TupleDesc desc,
				   AttrNumber attributeNumber,
				   const char *attributeName,
				   Oid oidtypeid,
				   int32 typmod,
				   int attdim)
{
	HeapTuple	tuple;
	Form_pg_type typeForm;
	Form_pg_attribute att;

	/*
	 * sanity checks
	 */
	Assert(PointerIsValid(desc));
	Assert(attributeNumber >= 1);
	Assert(attributeNumber <= desc->natts);
	Assert(attdim >= 0);
	Assert(attdim <= PG_INT16_MAX);

	/*
	 * initialize the attribute fields
	 */
	att = TupleDescAttr(desc, attributeNumber - 1);

	att->attrelid = 0;			/* dummy value */

	/*
	 * Note: attributeName can be NULL, because the planner doesn't always
	 * fill in valid resname values in targetlists, particularly for resjunk
	 * attributes. Also, do nothing if caller wants to re-use the old attname.
	 */
	if (attributeName == NULL)
		MemSet(NameStr(att->attname), 0, NAMEDATALEN);
	else if (attributeName != NameStr(att->attname))
		namestrcpy(&(att->attname), attributeName);

	att->atttypmod = typmod;

	att->attnum = attributeNumber;
	att->attndims = attdim;

	att->attnotnull = false;
	att->atthasdef = false;
	att->atthasmissing = false;
	att->attidentity = '\0';
	att->attgenerated = '\0';
	att->attisdropped = false;
	att->attislocal = true;
	att->attinhcount = 0;
	/* variable-length fields are not present in tupledescs */

	tuple = SearchSysCache1(TYPEOID, ObjectIdGetDatum(oidtypeid));
	if (!HeapTupleIsValid(tuple))
		elog(ERROR, "cache lookup failed for type %u", oidtypeid);
	typeForm = (Form_pg_type) GETSTRUCT(tuple);

	att->atttypid = oidtypeid;
	att->attlen = typeForm->typlen;
	att->attbyval = typeForm->typbyval;
	att->attalign = typeForm->typalign;
	att->attstorage = typeForm->typstorage;
	att->attcompression = InvalidCompressionMethod;
	att->attcollation = typeForm->typcollation;

	populate_compact_attribute(desc, attributeNumber - 1);

	ReleaseSysCache(tuple);
}

/*
 * TupleDescInitBuiltinEntry
 *		Initialize a tuple descriptor without catalog access.  Only
 *		a limited range of builtin types are supported.
 */


/*
 * TupleDescInitEntryCollation
 *
 * Assign a nondefault collation to a previously initialized tuple descriptor
 * entry.
 */
void
TupleDescInitEntryCollation(TupleDesc desc,
							AttrNumber attributeNumber,
							Oid collationid)
{
	/*
	 * sanity checks
	 */
	Assert(PointerIsValid(desc));
	Assert(attributeNumber >= 1);
	Assert(attributeNumber <= desc->natts);

	TupleDescAttr(desc, attributeNumber - 1)->attcollation = collationid;
}

/*
 * BuildDescFromLists
 *
 * Build a TupleDesc given lists of column names (as String nodes),
 * column type OIDs, typmods, and collation OIDs.
 *
 * No constraints are generated.
 *
 * This is for use with functions returning RECORD.
 */


/*
 * Get default expression (or NULL if none) for the given attribute number.
 */


/* ResourceOwner callbacks */




