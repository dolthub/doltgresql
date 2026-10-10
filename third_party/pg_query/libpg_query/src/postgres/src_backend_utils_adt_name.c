/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - namestrcpy
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * name.c
 *	  Functions for the built-in type "name".
 *
 * name replaces char16 and is carefully implemented so that it
 * is a string of physical length NAMEDATALEN.
 * DO NOT use hard-coded constants anywhere
 * always use NAMEDATALEN as the symbolic constant!   - jolly 8/21/95
 *
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 *
 * IDENTIFICATION
 *	  src/backend/utils/adt/name.c
 *
 *-------------------------------------------------------------------------
 */
#include "postgres.h"

#include "catalog/namespace.h"
#include "catalog/pg_collation.h"
#include "catalog/pg_type.h"
#include "libpq/pqformat.h"
#include "mb/pg_wchar.h"
#include "miscadmin.h"
#include "utils/array.h"
#include "utils/builtins.h"
#include "utils/lsyscache.h"
#include "utils/varlena.h"


/*****************************************************************************
 *	 USER I/O ROUTINES (none)												 *
 *****************************************************************************/


/*
 *		namein	- converts cstring to internal representation
 *
 *		Note:
 *				[Old] Currently if strlen(s) < NAMEDATALEN, the extra chars are nulls
 *				Now, always NULL terminated
 */


/*
 *		nameout - converts internal representation to cstring
 */


/*
 *		namerecv			- converts external binary format to name
 */


/*
 *		namesend			- converts name to binary format
 */



/*****************************************************************************
 *	 COMPARISON/SORTING ROUTINES											 *
 *****************************************************************************/

/*
 *		nameeq	- returns 1 iff arguments are equal
 *		namene	- returns 1 iff arguments are not equal
 *		namelt	- returns 1 iff a < b
 *		namele	- returns 1 iff a <= b
 *		namegt	- returns 1 iff a > b
 *		namege	- returns 1 iff a >= b
 *
 * Note that the use of strncmp with NAMEDATALEN limit is mostly historical;
 * strcmp would do as well, because we do not allow NAME values that don't
 * have a '\0' terminator.  Whatever might be past the terminator is not
 * considered relevant to comparisons.
 */



















/*****************************************************************************
 *	 MISCELLANEOUS PUBLIC ROUTINES											 *
 *****************************************************************************/

void
namestrcpy(Name name, const char *str)
{
	/* NB: We need to zero-pad the destination. */
	strncpy(NameStr(*name), str, NAMEDATALEN);
	NameStr(*name)[NAMEDATALEN - 1] = '\0';
}

/*
 * Compare a NAME to a C string
 *
 * Assumes C collation always; be careful when using this for
 * anything but equality checks!
 */



/*
 * SQL-functions CURRENT_USER, SESSION_USER
 */





/*
 * SQL-functions CURRENT_SCHEMA, CURRENT_SCHEMAS
 */




/*
 * SQL-function nameconcatoid(name, oid) returns name
 *
 * This is used in the information_schema to produce specific_name columns,
 * which are supposed to be unique per schema.  We achieve that (in an ugly
 * way) by appending the object's OID.  The result is the same as
 *		($1::text || '_' || $2::text)::name
 * except that, if it would not fit in NAMEDATALEN, we make it do so by
 * truncating the name input (not the oid).
 */

