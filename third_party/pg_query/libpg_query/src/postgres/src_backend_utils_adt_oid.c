/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - buildoidvector
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * oid.c
 *	  Functions for the built-in type Oid ... also oidvector.
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 *
 * IDENTIFICATION
 *	  src/backend/utils/adt/oid.c
 *
 *-------------------------------------------------------------------------
 */
#include "postgres.h"

#include <ctype.h>
#include <limits.h>

#include "catalog/pg_type.h"
#include "common/int.h"
#include "libpq/pqformat.h"
#include "nodes/miscnodes.h"
#include "nodes/value.h"
#include "utils/array.h"
#include "utils/builtins.h"


#define OidVectorSize(n)	(offsetof(oidvector, values) + (n) * sizeof(Oid))


/*****************************************************************************
 *	 USER I/O ROUTINES														 *
 *****************************************************************************/





/*
 *		oidrecv			- converts external binary format to oid
 */


/*
 *		oidsend			- converts oid to binary format
 */


/*
 * construct oidvector given a raw array of Oids
 *
 * If oids is NULL then caller must fill values[] afterward
 */
oidvector *
buildoidvector(const Oid *oids, int n)
{
	oidvector  *result;

	result = (oidvector *) palloc0(OidVectorSize(n));

	if (n > 0 && oids)
		memcpy(result->values, oids, n * sizeof(Oid));

	/*
	 * Attach standard array header.  For historical reasons, we set the index
	 * lower bound to 0 not 1.
	 */
	SET_VARSIZE(result, OidVectorSize(n));
	result->ndim = 1;
	result->dataoffset = 0;		/* never any nulls */
	result->elemtype = OIDOID;
	result->dim1 = n;
	result->lbound1 = 0;

	return result;
}

/*
 * validate that an array object meets the restrictions of oidvector
 *
 * We need this because there are pathways by which a general oid[] array can
 * be cast to oidvector, allowing the type's restrictions to be violated.
 * All code that receives an oidvector as a SQL parameter should check this.
 */


/*
 *		oidvectorin			- converts "num num ..." to internal form
 */


/*
 *		oidvectorout - converts internal form to "num num ..."
 */


/*
 *		oidvectorrecv			- converts external binary format to oidvector
 */


/*
 *		oidvectorsend			- converts oidvector to binary format
 */


/*
 *		oidparse				- get OID from ICONST/FCONST node
 */


/* qsort comparison function for Oids */



/*****************************************************************************
 *	 PUBLIC ROUTINES														 *
 *****************************************************************************/




























