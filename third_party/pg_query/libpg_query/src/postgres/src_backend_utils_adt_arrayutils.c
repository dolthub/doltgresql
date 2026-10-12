/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - ArrayGetNItems
 * - ArrayGetNItemsSafe
 * - ArrayCheckBounds
 * - ArrayCheckBoundsSafe
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * arrayutils.c
 *	  This file contains some support routines required for array functions.
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 *
 * IDENTIFICATION
 *	  src/backend/utils/adt/arrayutils.c
 *
 *-------------------------------------------------------------------------
 */

#include "postgres.h"

#include "catalog/pg_type.h"
#include "common/int.h"
#include "utils/array.h"
#include "utils/builtins.h"
#include "utils/memutils.h"


/*
 * Convert subscript list into linear element number (from 0)
 *
 * We assume caller has already range-checked the dimensions and subscripts,
 * so no overflow is possible.
 */


/*
 * Convert array dimensions into number of elements
 *
 * This must do overflow checking, since it is used to validate that a user
 * dimensionality request doesn't overflow what we can handle.
 *
 * The multiplication overflow check only works on machines that have int64
 * arithmetic, but that is nearly all platforms these days, and doing check
 * divides for those that don't seems way too expensive.
 */
int
ArrayGetNItems(int ndim, const int *dims)
{
	return ArrayGetNItemsSafe(ndim, dims, NULL);
}

/*
 * This entry point can return the error into an ErrorSaveContext
 * instead of throwing an exception.  -1 is returned after an error.
 */
int
ArrayGetNItemsSafe(int ndim, const int *dims, struct Node *escontext)
{
	int32		ret;
	int			i;

	if (ndim <= 0)
		return 0;
	ret = 1;
	for (i = 0; i < ndim; i++)
	{
		int64		prod;

		/* A negative dimension implies that UB-LB overflowed ... */
		if (dims[i] < 0)
			ereturn(escontext, -1,
					(errcode(ERRCODE_PROGRAM_LIMIT_EXCEEDED),
					 errmsg("array size exceeds the maximum allowed (%d)",
							(int) MaxArraySize)));

		prod = (int64) ret * (int64) dims[i];

		ret = (int32) prod;
		if ((int64) ret != prod)
			ereturn(escontext, -1,
					(errcode(ERRCODE_PROGRAM_LIMIT_EXCEEDED),
					 errmsg("array size exceeds the maximum allowed (%d)",
							(int) MaxArraySize)));
	}
	Assert(ret >= 0);
	if ((Size) ret > MaxArraySize)
		ereturn(escontext, -1,
				(errcode(ERRCODE_PROGRAM_LIMIT_EXCEEDED),
				 errmsg("array size exceeds the maximum allowed (%d)",
						(int) MaxArraySize)));
	return (int) ret;
}

/*
 * Verify sanity of proposed lower-bound values for an array
 *
 * The lower-bound values must not be so large as to cause overflow when
 * calculating subscripts, e.g. lower bound 2147483640 with length 10
 * must be disallowed.  We actually insist that dims[i] + lb[i] be
 * computable without overflow, meaning that an array with last subscript
 * equal to INT_MAX will be disallowed.
 *
 * It is assumed that the caller already called ArrayGetNItems, so that
 * overflowed (negative) dims[] values have been eliminated.
 */
void
ArrayCheckBounds(int ndim, const int *dims, const int *lb)
{
	(void) ArrayCheckBoundsSafe(ndim, dims, lb, NULL);
}

/*
 * This entry point can return the error into an ErrorSaveContext
 * instead of throwing an exception.
 */
bool
ArrayCheckBoundsSafe(int ndim, const int *dims, const int *lb,
					 struct Node *escontext)
{
	int			i;

	for (i = 0; i < ndim; i++)
	{
		/* PG_USED_FOR_ASSERTS_ONLY prevents variable-isn't-read warnings */
		int32		sum PG_USED_FOR_ASSERTS_ONLY;

		if (pg_add_s32_overflow(dims[i], lb[i], &sum))
			ereturn(escontext, false,
					(errcode(ERRCODE_PROGRAM_LIMIT_EXCEEDED),
					 errmsg("array lower bound is too large: %d",
							lb[i])));
	}

	return true;
}

/*
 * Compute ranges (sub-array dimensions) for an array slice
 *
 * We assume caller has validated slice endpoints, so overflow is impossible
 */


/*
 * Compute products of array dimensions, ie, scale factors for subscripts
 *
 * We assume caller has validated dimensions, so overflow is impossible
 */


/*
 * From products of whole-array dimensions and spans of a sub-array,
 * compute offset distances needed to step through subarray within array
 *
 * We assume caller has validated dimensions, so overflow is impossible
 */


/*
 * Generates the tuple that is lexicographically one greater than the current
 * n-tuple in "curr", with the restriction that the i-th element of "curr" is
 * less than the i-th element of "span".
 *
 * Returns -1 if no next tuple exists, else the subscript position (0..n-1)
 * corresponding to the dimension to advance along.
 *
 * We assume caller has validated dimensions, so overflow is impossible
 */


/*
 * ArrayGetIntegerTypmods: verify that argument is a 1-D cstring array,
 * and get the contents converted to integers.  Returns a palloc'd array
 * and places the length at *n.
 */

