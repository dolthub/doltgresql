/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - construct_array_builtin
 * - construct_array
 * - construct_md_array
 * - construct_empty_array
 * - CopyArrayEls
 * - ArrayCastAndSet
 * - deconstruct_array_builtin
 * - deconstruct_array
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * arrayfuncs.c
 *	  Support functions for arrays.
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 *
 * IDENTIFICATION
 *	  src/backend/utils/adt/arrayfuncs.c
 *
 *-------------------------------------------------------------------------
 */
#include "postgres.h"

#include <ctype.h>
#include <math.h>

#include "catalog/pg_type.h"
#include "common/int.h"
#include "funcapi.h"
#include "libpq/pqformat.h"
#include "nodes/nodeFuncs.h"
#include "nodes/supportnodes.h"
#include "optimizer/optimizer.h"
#include "parser/scansup.h"
#include "port/pg_bitutils.h"
#include "utils/array.h"
#include "utils/arrayaccess.h"
#include "utils/builtins.h"
#include "utils/datum.h"
#include "utils/fmgroids.h"
#include "utils/lsyscache.h"
#include "utils/memutils.h"
#include "utils/selfuncs.h"
#include "utils/typcache.h"


/*
 * GUC parameter
 */


/*
 * Local definitions
 */
#define ASSGN	 "="

#define AARR_FREE_IF_COPY(array,n) \
	do { \
		if (!VARATT_IS_EXPANDED_HEADER(array)) \
			PG_FREE_IF_COPY(array, n); \
	} while (0)

/* ReadArrayToken return type */
typedef enum
{
	ATOK_LEVEL_START,
	ATOK_LEVEL_END,
	ATOK_DELIM,
	ATOK_ELEM,
	ATOK_ELEM_NULL,
	ATOK_ERROR,
} ArrayToken;

/* Working state for array_iterate() */
typedef struct ArrayIteratorData
{
	/* basic info about the array, set up during array_create_iterator() */
	ArrayType  *arr;			/* array we're iterating through */
	bits8	   *nullbitmap;		/* its null bitmap, if any */
	int			nitems;			/* total number of elements in array */
	int16		typlen;			/* element type's length */
	bool		typbyval;		/* element type's byval property */
	char		typalign;		/* element type's align property */

	/* information about the requested slice size */
	int			slice_ndim;		/* slice dimension, or 0 if not slicing */
	int			slice_len;		/* number of elements per slice */
	int		   *slice_dims;		/* slice dims array */
	int		   *slice_lbound;	/* slice lbound array */
	Datum	   *slice_values;	/* workspace of length slice_len */
	bool	   *slice_nulls;	/* workspace of length slice_len */

	/* current position information, updated on each iteration */
	char	   *data_ptr;		/* our current position in the array */
	int			current_item;	/* the item # we're at in the array */
}			ArrayIteratorData;

static bool ReadArrayDimensions(char **srcptr, int *ndim_p,
								int *dim, int *lBound,
								const char *origStr, Node *escontext);
static bool ReadDimensionInt(char **srcptr, int *result,
							 const char *origStr, Node *escontext);
static bool ReadArrayStr(char **srcptr,
						 FmgrInfo *inputproc, Oid typioparam, int32 typmod,
						 char typdelim,
						 int typlen, bool typbyval, char typalign,
						 int *ndim_p, int *dim,
						 int *nitems_p,
						 Datum **values_p, bool **nulls_p,
						 const char *origStr, Node *escontext);
static ArrayToken ReadArrayToken(char **srcptr, StringInfo elembuf, char typdelim,
								 const char *origStr, Node *escontext);
static void ReadArrayBinary(StringInfo buf, int nitems,
							FmgrInfo *receiveproc, Oid typioparam, int32 typmod,
							int typlen, bool typbyval, char typalign,
							Datum *values, bool *nulls,
							bool *hasnulls, int32 *nbytes);
static Datum array_get_element_expanded(Datum arraydatum,
										int nSubscripts, int *indx,
										int arraytyplen,
										int elmlen, bool elmbyval, char elmalign,
										bool *isNull);
static Datum array_set_element_expanded(Datum arraydatum,
										int nSubscripts, int *indx,
										Datum dataValue, bool isNull,
										int arraytyplen,
										int elmlen, bool elmbyval, char elmalign);
static bool array_get_isnull(const bits8 *nullbitmap, int offset);
static void array_set_isnull(bits8 *nullbitmap, int offset, bool isNull);
static Datum ArrayCast(char *value, bool byval, int len);
static int	ArrayCastAndSet(Datum src,
							int typlen, bool typbyval, char typalign,
							char *dest);
static char *array_seek(char *ptr, int offset, bits8 *nullbitmap, int nitems,
						int typlen, bool typbyval, char typalign);
static int	array_nelems_size(char *ptr, int offset, bits8 *nullbitmap,
							  int nitems, int typlen, bool typbyval, char typalign);
static int	array_copy(char *destptr, int nitems,
					   char *srcptr, int offset, bits8 *nullbitmap,
					   int typlen, bool typbyval, char typalign);
static int	array_slice_size(char *arraydataptr, bits8 *arraynullsptr,
							 int ndim, int *dim, int *lb,
							 int *st, int *endp,
							 int typlen, bool typbyval, char typalign);
static void array_extract_slice(ArrayType *newarray,
								int ndim, int *dim, int *lb,
								char *arraydataptr, bits8 *arraynullsptr,
								int *st, int *endp,
								int typlen, bool typbyval, char typalign);
static void array_insert_slice(ArrayType *destArray, ArrayType *origArray,
							   ArrayType *srcArray,
							   int ndim, int *dim, int *lb,
							   int *st, int *endp,
							   int typlen, bool typbyval, char typalign);
static int	array_cmp(FunctionCallInfo fcinfo);
static ArrayType *create_array_envelope(int ndims, int *dimv, int *lbsv, int nbytes,
										Oid elmtype, int dataoffset);
static ArrayType *array_fill_internal(ArrayType *dims, ArrayType *lbs,
									  Datum value, bool isnull, Oid elmtype,
									  FunctionCallInfo fcinfo);
static ArrayType *array_replace_internal(ArrayType *array,
										 Datum search, bool search_isnull,
										 Datum replace, bool replace_isnull,
										 bool remove, Oid collation,
										 FunctionCallInfo fcinfo);
static int	width_bucket_array_float8(Datum operand, ArrayType *thresholds);
static int	width_bucket_array_fixed(Datum operand,
									 ArrayType *thresholds,
									 Oid collation,
									 TypeCacheEntry *typentry);
static int	width_bucket_array_variable(Datum operand,
										ArrayType *thresholds,
										Oid collation,
										TypeCacheEntry *typentry);


/*
 * array_in :
 *		  converts an array from the external format in "string" to
 *		  its internal format.
 *
 * return value :
 *		  the internal representation of the input array
 */


/*
 * ReadArrayDimensions
 *	 parses the array dimensions part of the input and converts the values
 *	 to internal format.
 *
 * On entry, *srcptr points to the string to parse. It is advanced to point
 * after whitespace (if any) and dimension info (if any).
 *
 * *ndim_p, dim[], and lBound[] are output variables. They are filled with the
 * number of dimensions (<= MAXDIM), the lengths of each dimension, and the
 * lower subscript bounds, respectively.  If no dimension info appears,
 * *ndim_p will be set to zero, and dim[] and lBound[] are unchanged.
 *
 * 'origStr' is the original input string, used only in error messages.
 * If *escontext points to an ErrorSaveContext, details of any error are
 * reported there.
 *
 * Result:
 *	true for success, false for failure (if escontext is provided).
 *
 * Note that dim[] and lBound[] are allocated by the caller, and must have
 * MAXDIM elements.
 */


/*
 * ReadDimensionInt
 *	 parse an integer, for the array dimensions
 *
 * On entry, *srcptr points to the string to parse. It is advanced past the
 * digits of the integer. If there are no digits, returns true and leaves
 * *srcptr unchanged.
 *
 * Result:
 *	true for success, false for failure (if escontext is provided).
 *  On success, the parsed integer is returned in *result.
 */


/*
 * ReadArrayStr :
 *	 parses the array string pointed to by *srcptr and converts the values
 *	 to internal format.  Determines the array dimensions as it goes.
 *
 * On entry, *srcptr points to the string to parse (it must point to a '{').
 * On successful return, it is advanced to point past the closing '}'.
 *
 * If dimensions were specified explicitly, they are passed in *ndim_p and
 * dim[].  This function will check that the array values match the specified
 * dimensions.  If dimensions were not given, caller must pass *ndim_p == 0
 * and initialize all elements of dim[] to -1.  Then this function will
 * deduce the dimensions from the structure of the input and store them in
 * *ndim_p and the dim[] array.
 *
 * Element type information:
 *	inputproc: type-specific input procedure for element datatype.
 *	typioparam, typmod: auxiliary values to pass to inputproc.
 *	typdelim: the value delimiter (type-specific).
 *	typlen, typbyval, typalign: storage parameters of element datatype.
 *
 * Outputs:
 *  *ndim_p, dim: dimensions deduced from the input structure.
 *  *nitems_p: total number of elements.
 *	*values_p[]: palloc'd array, filled with converted data values.
 *	*nulls_p[]: palloc'd array, filled with is-null markers.
 *
 * 'origStr' is the original input string, used only in error messages.
 * If *escontext points to an ErrorSaveContext, details of any error are
 * reported there.
 *
 * Result:
 *	true for success, false for failure (if escontext is provided).
 */


/*
 * ReadArrayToken
 *	 read one token from an array value string
 *
 * Starts scanning from *srcptr.  On non-error return, *srcptr is
 * advanced past the token.
 *
 * If the token is ATOK_ELEM, the de-escaped string is returned in elembuf.
 */


/*
 * Copy data into an array object from a temporary array of Datums.
 *
 * array: array object (with header fields already filled in)
 * values: array of Datums to be copied
 * nulls: array of is-null flags (can be NULL if no nulls)
 * nitems: number of Datums to be copied
 * typbyval, typlen, typalign: info about element datatype
 * freedata: if true and element type is pass-by-ref, pfree data values
 * referenced by Datums after copying them.
 *
 * If the input data is of varlena type, the caller must have ensured that
 * the values are not toasted.  (Doing it here doesn't work since the
 * caller has already allocated space for the array...)
 */
void
CopyArrayEls(ArrayType *array,
			 Datum *values,
			 bool *nulls,
			 int nitems,
			 int typlen,
			 bool typbyval,
			 char typalign,
			 bool freedata)
{
	char	   *p = ARR_DATA_PTR(array);
	bits8	   *bitmap = ARR_NULLBITMAP(array);
	int			bitval = 0;
	int			bitmask = 1;
	int			i;

	if (typbyval)
		freedata = false;

	for (i = 0; i < nitems; i++)
	{
		if (nulls && nulls[i])
		{
			if (!bitmap)		/* shouldn't happen */
				elog(ERROR, "null array element where not supported");
			/* bitmap bit stays 0 */
		}
		else
		{
			bitval |= bitmask;
			p += ArrayCastAndSet(values[i], typlen, typbyval, typalign, p);
			if (freedata)
				pfree(DatumGetPointer(values[i]));
		}
		if (bitmap)
		{
			bitmask <<= 1;
			if (bitmask == 0x100)
			{
				*bitmap++ = bitval;
				bitval = 0;
				bitmask = 1;
			}
		}
	}

	if (bitmap && bitmask != 1)
		*bitmap = bitval;
}

/*
 * array_out :
 *		   takes the internal representation of an array and returns a string
 *		  containing the array in its external format.
 */
#define APPENDSTR(str)	(strcpy(p, (str)), p += strlen(p))
#define APPENDCHAR(ch)	(*p++ = (ch), *p = '\0')
#undef APPENDSTR
#undef APPENDCHAR

/*
 * array_recv :
 *		  converts an array from the external binary format to
 *		  its internal format.
 *
 * return value :
 *		  the internal representation of the input array
 */


/*
 * ReadArrayBinary:
 *	 collect the data elements of an array being read in binary style.
 *
 * Inputs:
 *	buf: the data buffer to read from.
 *	nitems: total number of array elements (already read).
 *	receiveproc: type-specific receive procedure for element datatype.
 *	typioparam, typmod: auxiliary values to pass to receiveproc.
 *	typlen, typbyval, typalign: storage parameters of element datatype.
 *
 * Outputs:
 *	values[]: filled with converted data values.
 *	nulls[]: filled with is-null markers.
 *	*hasnulls: set true iff there are any null elements.
 *	*nbytes: set to total size of data area needed (including alignment
 *		padding but not including array header overhead).
 *
 * Note that values[] and nulls[] are allocated by the caller, and must have
 * nitems elements.
 */



/*
 * array_send :
 *		  takes the internal representation of an array and returns a bytea
 *		  containing the array in its external binary format.
 */


/*
 * array_ndims :
 *		  returns the number of dimensions of the array pointed to by "v"
 */


/*
 * array_dims :
 *		  returns the dimensions of the array pointed to by "v", as a "text"
 */


/*
 * array_lower :
 *		returns the lower dimension, of the DIM requested, for
 *		the array pointed to by "v", as an int4
 */


/*
 * array_upper :
 *		returns the upper dimension, of the DIM requested, for
 *		the array pointed to by "v", as an int4
 */


/*
 * array_length :
 *		returns the length, of the dimension requested, for
 *		the array pointed to by "v", as an int4
 */


/*
 * array_cardinality:
 *		returns the total number of elements in an array
 */



/*
 * array_get_element :
 *	  This routine takes an array datum and a subscript array and returns
 *	  the referenced item as a Datum.  Note that for a pass-by-reference
 *	  datatype, the returned Datum is a pointer into the array object.
 *
 * This handles both ordinary varlena arrays and fixed-length arrays.
 *
 * Inputs:
 *	arraydatum: the array object (mustn't be NULL)
 *	nSubscripts: number of subscripts supplied
 *	indx[]: the subscript values
 *	arraytyplen: pg_type.typlen for the array type
 *	elmlen: pg_type.typlen for the array's element type
 *	elmbyval: pg_type.typbyval for the array's element type
 *	elmalign: pg_type.typalign for the array's element type
 *
 * Outputs:
 *	The return value is the element Datum.
 *	*isNull is set to indicate whether the element is NULL.
 */


/*
 * Implementation of array_get_element() for an expanded array
 */


/*
 * array_get_slice :
 *		   This routine takes an array and a range of indices (upperIndx and
 *		   lowerIndx), creates a new array structure for the referred elements
 *		   and returns a pointer to it.
 *
 * This handles both ordinary varlena arrays and fixed-length arrays.
 *
 * Inputs:
 *	arraydatum: the array object (mustn't be NULL)
 *	nSubscripts: number of subscripts supplied (must be same for upper/lower)
 *	upperIndx[]: the upper subscript values
 *	lowerIndx[]: the lower subscript values
 *	upperProvided[]: true for provided upper subscript values
 *	lowerProvided[]: true for provided lower subscript values
 *	arraytyplen: pg_type.typlen for the array type
 *	elmlen: pg_type.typlen for the array's element type
 *	elmbyval: pg_type.typbyval for the array's element type
 *	elmalign: pg_type.typalign for the array's element type
 *
 * Outputs:
 *	The return value is the new array Datum (it's never NULL)
 *
 * Omitted upper and lower subscript values are replaced by the corresponding
 * array bound.
 *
 * NOTE: we assume it is OK to scribble on the provided subscript arrays
 * lowerIndx[] and upperIndx[]; also, these arrays must be of size MAXDIM
 * even when nSubscripts is less.  These are generally just temporaries.
 */


/*
 * array_set_element :
 *		  This routine sets the value of one array element (specified by
 *		  a subscript array) to a new value specified by "dataValue".
 *
 * This handles both ordinary varlena arrays and fixed-length arrays.
 *
 * Inputs:
 *	arraydatum: the initial array object (mustn't be NULL)
 *	nSubscripts: number of subscripts supplied
 *	indx[]: the subscript values
 *	dataValue: the datum to be inserted at the given position
 *	isNull: whether dataValue is NULL
 *	arraytyplen: pg_type.typlen for the array type
 *	elmlen: pg_type.typlen for the array's element type
 *	elmbyval: pg_type.typbyval for the array's element type
 *	elmalign: pg_type.typalign for the array's element type
 *
 * Result:
 *		  A new array is returned, just like the old except for the one
 *		  modified entry.  The original array object is not changed,
 *		  unless what is passed is a read-write reference to an expanded
 *		  array object; in that case the expanded array is updated in-place.
 *
 * For one-dimensional arrays only, we allow the array to be extended
 * by assigning to a position outside the existing subscript range; any
 * positions between the existing elements and the new one are set to NULLs.
 * (XXX TODO: allow a corresponding behavior for multidimensional arrays)
 *
 * NOTE: For assignments, we throw an error for invalid subscripts etc,
 * rather than returning a NULL as the fetch operations do.
 */


/*
 * Implementation of array_set_element() for an expanded array
 *
 * Note: as with any operation on a read/write expanded object, we must
 * take pains not to leave the object in a corrupt state if we fail partway
 * through.
 */


/*
 * array_set_slice :
 *		  This routine sets the value of a range of array locations (specified
 *		  by upper and lower subscript values) to new values passed as
 *		  another array.
 *
 * This handles both ordinary varlena arrays and fixed-length arrays.
 *
 * Inputs:
 *	arraydatum: the initial array object (mustn't be NULL)
 *	nSubscripts: number of subscripts supplied (must be same for upper/lower)
 *	upperIndx[]: the upper subscript values
 *	lowerIndx[]: the lower subscript values
 *	upperProvided[]: true for provided upper subscript values
 *	lowerProvided[]: true for provided lower subscript values
 *	srcArrayDatum: the source for the inserted values
 *	isNull: indicates whether srcArrayDatum is NULL
 *	arraytyplen: pg_type.typlen for the array type
 *	elmlen: pg_type.typlen for the array's element type
 *	elmbyval: pg_type.typbyval for the array's element type
 *	elmalign: pg_type.typalign for the array's element type
 *
 * Result:
 *		  A new array is returned, just like the old except for the
 *		  modified range.  The original array object is not changed.
 *
 * Omitted upper and lower subscript values are replaced by the corresponding
 * array bound.
 *
 * For one-dimensional arrays only, we allow the array to be extended
 * by assigning to positions outside the existing subscript range; any
 * positions between the existing elements and the new ones are set to NULLs.
 * (XXX TODO: allow a corresponding behavior for multidimensional arrays)
 *
 * NOTE: we assume it is OK to scribble on the provided index arrays
 * lowerIndx[] and upperIndx[]; also, these arrays must be of size MAXDIM
 * even when nSubscripts is less.  These are generally just temporaries.
 *
 * NOTE: For assignments, we throw an error for silly subscripts etc,
 * rather than returning a NULL or empty array as the fetch operations do.
 */


/*
 * array_ref : backwards compatibility wrapper for array_get_element
 *
 * This only works for detoasted/flattened varlena arrays, since the array
 * argument is declared as "ArrayType *".  However there's enough code like
 * that to justify preserving this API.
 */


/*
 * array_set : backwards compatibility wrapper for array_set_element
 *
 * This only works for detoasted/flattened varlena arrays, since the array
 * argument and result are declared as "ArrayType *".  However there's enough
 * code like that to justify preserving this API.
 */


/*
 * array_map()
 *
 * Map an array through an arbitrary expression.  Return a new array with
 * the same dimensions and each source element transformed by the given,
 * already-compiled expression.  Each source element is placed in the
 * innermost_caseval/innermost_casenull fields of the ExprState.
 *
 * Parameters are:
 * * arrayd: Datum representing array argument.
 * * exprstate: ExprState representing the per-element transformation.
 * * econtext: context for expression evaluation.
 * * retType: OID of element type of output array.  This must be the same as,
 *	 or binary-compatible with, the result type of the expression.  It might
 *	 be different from the input array's element type.
 * * amstate: workspace for array_map.  Must be zeroed by caller before
 *	 first call, and not touched after that.
 *
 * It is legitimate to pass a freshly-zeroed ArrayMapState on each call,
 * but better performance can be had if the state can be preserved across
 * a series of calls.
 *
 * NB: caller must assure that input array is not NULL.  NULL elements in
 * the array are OK however.
 * NB: caller should be running in econtext's per-tuple memory context.
 */


/*
 * construct_array	--- simple method for constructing an array object
 *
 * elems: array of Datum items to become the array contents
 *		  (NULL element values are not supported).
 * nelems: number of items
 * elmtype, elmlen, elmbyval, elmalign: info for the datatype of the items
 *
 * A palloc'd 1-D array object is constructed and returned.  Note that
 * elem values will be copied into the object even if pass-by-ref type.
 * Also note the result will be 0-D not 1-D if nelems = 0.
 *
 * NOTE: it would be cleaner to look up the elmlen/elmbval/elmalign info
 * from the system catalogs, given the elmtype.  However, the caller is
 * in a better position to cache this info across multiple uses, or even
 * to hard-wire values if the element type is hard-wired.
 */
ArrayType *
construct_array(Datum *elems, int nelems,
				Oid elmtype,
				int elmlen, bool elmbyval, char elmalign)
{
	int			dims[1];
	int			lbs[1];

	dims[0] = nelems;
	lbs[0] = 1;

	return construct_md_array(elems, NULL, 1, dims, lbs,
							  elmtype, elmlen, elmbyval, elmalign);
}

/*
 * Like construct_array(), where elmtype must be a built-in type, and
 * elmlen/elmbyval/elmalign is looked up from hardcoded data.  This is often
 * useful when manipulating arrays from/for system catalogs.
 */
ArrayType *
construct_array_builtin(Datum *elems, int nelems, Oid elmtype)
{
	int			elmlen;
	bool		elmbyval;
	char		elmalign;

	switch (elmtype)
	{
		case CHAROID:
			elmlen = 1;
			elmbyval = true;
			elmalign = TYPALIGN_CHAR;
			break;

		case CSTRINGOID:
			elmlen = -2;
			elmbyval = false;
			elmalign = TYPALIGN_CHAR;
			break;

		case FLOAT4OID:
			elmlen = sizeof(float4);
			elmbyval = true;
			elmalign = TYPALIGN_INT;
			break;

		case FLOAT8OID:
			elmlen = sizeof(float8);
			elmbyval = FLOAT8PASSBYVAL;
			elmalign = TYPALIGN_DOUBLE;
			break;

		case INT2OID:
			elmlen = sizeof(int16);
			elmbyval = true;
			elmalign = TYPALIGN_SHORT;
			break;

		case INT4OID:
			elmlen = sizeof(int32);
			elmbyval = true;
			elmalign = TYPALIGN_INT;
			break;

		case INT8OID:
			elmlen = sizeof(int64);
			elmbyval = FLOAT8PASSBYVAL;
			elmalign = TYPALIGN_DOUBLE;
			break;

		case NAMEOID:
			elmlen = NAMEDATALEN;
			elmbyval = false;
			elmalign = TYPALIGN_CHAR;
			break;

		case OIDOID:
		case REGTYPEOID:
			elmlen = sizeof(Oid);
			elmbyval = true;
			elmalign = TYPALIGN_INT;
			break;

		case TEXTOID:
			elmlen = -1;
			elmbyval = false;
			elmalign = TYPALIGN_INT;
			break;

		case TIDOID:
			elmlen = sizeof(ItemPointerData);
			elmbyval = false;
			elmalign = TYPALIGN_SHORT;
			break;

		case XIDOID:
			elmlen = sizeof(TransactionId);
			elmbyval = true;
			elmalign = TYPALIGN_INT;
			break;

		default:
			elog(ERROR, "type %u not supported by construct_array_builtin()", elmtype);
			/* keep compiler quiet */
			elmlen = 0;
			elmbyval = false;
			elmalign = 0;
	}

	return construct_array(elems, nelems, elmtype, elmlen, elmbyval, elmalign);
}

/*
 * construct_md_array	--- simple method for constructing an array object
 *							with arbitrary dimensions and possible NULLs
 *
 * elems: array of Datum items to become the array contents
 * nulls: array of is-null flags (can be NULL if no nulls)
 * ndims: number of dimensions
 * dims: integer array with size of each dimension
 * lbs: integer array with lower bound of each dimension
 * elmtype, elmlen, elmbyval, elmalign: info for the datatype of the items
 *
 * A palloc'd ndims-D array object is constructed and returned.  Note that
 * elem values will be copied into the object even if pass-by-ref type.
 * Also note the result will be 0-D not ndims-D if any dims[i] = 0.
 *
 * NOTE: it would be cleaner to look up the elmlen/elmbval/elmalign info
 * from the system catalogs, given the elmtype.  However, the caller is
 * in a better position to cache this info across multiple uses, or even
 * to hard-wire values if the element type is hard-wired.
 */
ArrayType *
construct_md_array(Datum *elems,
				   bool *nulls,
				   int ndims,
				   int *dims,
				   int *lbs,
				   Oid elmtype, int elmlen, bool elmbyval, char elmalign)
{
	ArrayType  *result;
	bool		hasnulls;
	int32		nbytes;
	int32		dataoffset;
	int			i;
	int			nelems;

	if (ndims < 0)				/* we do allow zero-dimension arrays */
		ereport(ERROR,
				(errcode(ERRCODE_INVALID_PARAMETER_VALUE),
				 errmsg("invalid number of dimensions: %d", ndims)));
	if (ndims > MAXDIM)
		ereport(ERROR,
				(errcode(ERRCODE_PROGRAM_LIMIT_EXCEEDED),
				 errmsg("number of array dimensions (%d) exceeds the maximum allowed (%d)",
						ndims, MAXDIM)));

	/* This checks for overflow of the array dimensions */
	nelems = ArrayGetNItems(ndims, dims);
	ArrayCheckBounds(ndims, dims, lbs);

	/* if ndims <= 0 or any dims[i] == 0, return empty array */
	if (nelems <= 0)
		return construct_empty_array(elmtype);

	/* compute required space */
	nbytes = 0;
	hasnulls = false;
	for (i = 0; i < nelems; i++)
	{
		if (nulls && nulls[i])
		{
			hasnulls = true;
			continue;
		}
		/* make sure data is not toasted */
		if (elmlen == -1)
			elems[i] = PointerGetDatum(PG_DETOAST_DATUM(elems[i]));
		nbytes = att_addlength_datum(nbytes, elmlen, elems[i]);
		nbytes = att_align_nominal(nbytes, elmalign);
		/* check for overflow of total request */
		if (!AllocSizeIsValid(nbytes))
			ereport(ERROR,
					(errcode(ERRCODE_PROGRAM_LIMIT_EXCEEDED),
					 errmsg("array size exceeds the maximum allowed (%d)",
							(int) MaxAllocSize)));
	}

	/* Allocate and initialize result array */
	if (hasnulls)
	{
		dataoffset = ARR_OVERHEAD_WITHNULLS(ndims, nelems);
		nbytes += dataoffset;
	}
	else
	{
		dataoffset = 0;			/* marker for no null bitmap */
		nbytes += ARR_OVERHEAD_NONULLS(ndims);
	}
	result = (ArrayType *) palloc0(nbytes);
	SET_VARSIZE(result, nbytes);
	result->ndim = ndims;
	result->dataoffset = dataoffset;
	result->elemtype = elmtype;
	memcpy(ARR_DIMS(result), dims, ndims * sizeof(int));
	memcpy(ARR_LBOUND(result), lbs, ndims * sizeof(int));

	CopyArrayEls(result,
				 elems, nulls, nelems,
				 elmlen, elmbyval, elmalign,
				 false);

	return result;
}

/*
 * construct_empty_array	--- make a zero-dimensional array of given type
 */
ArrayType *
construct_empty_array(Oid elmtype)
{
	ArrayType  *result;

	result = (ArrayType *) palloc0(sizeof(ArrayType));
	SET_VARSIZE(result, sizeof(ArrayType));
	result->ndim = 0;
	result->dataoffset = 0;
	result->elemtype = elmtype;
	return result;
}

/*
 * construct_empty_expanded_array: make an empty expanded array
 * given only type information.  (metacache can be NULL if not needed.)
 */


/*
 * deconstruct_array  --- simple method for extracting data from an array
 *
 * array: array object to examine (must not be NULL)
 * elmtype, elmlen, elmbyval, elmalign: info for the datatype of the items
 * elemsp: return value, set to point to palloc'd array of Datum values
 * nullsp: return value, set to point to palloc'd array of isnull markers
 * nelemsp: return value, set to number of extracted values
 *
 * The caller may pass nullsp == NULL if it does not support NULLs in the
 * array.  Note that this produces a very uninformative error message,
 * so do it only in cases where a NULL is really not expected.
 *
 * If array elements are pass-by-ref data type, the returned Datums will
 * be pointers into the array object.
 *
 * NOTE: it would be cleaner to look up the elmlen/elmbval/elmalign info
 * from the system catalogs, given the elmtype.  However, the caller is
 * in a better position to cache this info across multiple uses, or even
 * to hard-wire values if the element type is hard-wired.
 */
void
deconstruct_array(ArrayType *array,
				  Oid elmtype,
				  int elmlen, bool elmbyval, char elmalign,
				  Datum **elemsp, bool **nullsp, int *nelemsp)
{
	Datum	   *elems;
	bool	   *nulls;
	int			nelems;
	char	   *p;
	bits8	   *bitmap;
	int			bitmask;
	int			i;

	Assert(ARR_ELEMTYPE(array) == elmtype);

	nelems = ArrayGetNItems(ARR_NDIM(array), ARR_DIMS(array));
	*elemsp = elems = (Datum *) palloc(nelems * sizeof(Datum));
	if (nullsp)
		*nullsp = nulls = (bool *) palloc0(nelems * sizeof(bool));
	else
		nulls = NULL;
	*nelemsp = nelems;

	p = ARR_DATA_PTR(array);
	bitmap = ARR_NULLBITMAP(array);
	bitmask = 1;

	for (i = 0; i < nelems; i++)
	{
		/* Get source element, checking for NULL */
		if (bitmap && (*bitmap & bitmask) == 0)
		{
			elems[i] = (Datum) 0;
			if (nulls)
				nulls[i] = true;
			else
				ereport(ERROR,
						(errcode(ERRCODE_NULL_VALUE_NOT_ALLOWED),
						 errmsg("null array element not allowed in this context")));
		}
		else
		{
			elems[i] = fetch_att(p, elmbyval, elmlen);
			p = att_addlength_pointer(p, elmlen, p);
			p = (char *) att_align_nominal(p, elmalign);
		}

		/* advance bitmap pointer if any */
		if (bitmap)
		{
			bitmask <<= 1;
			if (bitmask == 0x100)
			{
				bitmap++;
				bitmask = 1;
			}
		}
	}
}

/*
 * Like deconstruct_array(), where elmtype must be a built-in type, and
 * elmlen/elmbyval/elmalign is looked up from hardcoded data.  This is often
 * useful when manipulating arrays from/for system catalogs.
 */
void
deconstruct_array_builtin(ArrayType *array,
						  Oid elmtype,
						  Datum **elemsp, bool **nullsp, int *nelemsp)
{
	int			elmlen;
	bool		elmbyval;
	char		elmalign;

	switch (elmtype)
	{
		case CHAROID:
			elmlen = 1;
			elmbyval = true;
			elmalign = TYPALIGN_CHAR;
			break;

		case CSTRINGOID:
			elmlen = -2;
			elmbyval = false;
			elmalign = TYPALIGN_CHAR;
			break;

		case FLOAT8OID:
			elmlen = sizeof(float8);
			elmbyval = FLOAT8PASSBYVAL;
			elmalign = TYPALIGN_DOUBLE;
			break;

		case INT2OID:
			elmlen = sizeof(int16);
			elmbyval = true;
			elmalign = TYPALIGN_SHORT;
			break;

		case INT4OID:
			elmlen = sizeof(int32);
			elmbyval = true;
			elmalign = TYPALIGN_INT;
			break;

		case OIDOID:
			elmlen = sizeof(Oid);
			elmbyval = true;
			elmalign = TYPALIGN_INT;
			break;

		case TEXTOID:
			elmlen = -1;
			elmbyval = false;
			elmalign = TYPALIGN_INT;
			break;

		case TIDOID:
			elmlen = sizeof(ItemPointerData);
			elmbyval = false;
			elmalign = TYPALIGN_SHORT;
			break;

		default:
			elog(ERROR, "type %u not supported by deconstruct_array_builtin()", elmtype);
			/* keep compiler quiet */
			elmlen = 0;
			elmbyval = false;
			elmalign = 0;
	}

	deconstruct_array(array, elmtype, elmlen, elmbyval, elmalign, elemsp, nullsp, nelemsp);
}

/*
 * array_contains_nulls --- detect whether an array has any null elements
 *
 * This gives an accurate answer, whereas testing ARR_HASNULL only tells
 * if the array *might* contain a null.
 */



/*
 * array_eq :
 *		  compares two arrays for equality
 * result :
 *		  returns true if the arrays are equal, false otherwise.
 *
 * Note: we do not use array_cmp here, since equality may be meaningful in
 * datatypes that don't have a total ordering (and hence no btree support).
 */



/*-----------------------------------------------------------------------------
 * array-array bool operators:
 *		Given two arrays, iterate comparison operators
 *		over the array. Uses logic similar to text comparison
 *		functions, except element-by-element instead of
 *		character-by-character.
 *----------------------------------------------------------------------------
 */













/*
 * array_cmp()
 * Internal comparison function for arrays.
 *
 * Returns -1, 0 or 1
 */



/*-----------------------------------------------------------------------------
 * array hashing
 *		Hash the elements and combine the results.
 *----------------------------------------------------------------------------
 */



/*
 * Returns 64-bit value by hashing a value to a 64-bit value, with a seed.
 * Otherwise, similar to hash_array.
 */



/*-----------------------------------------------------------------------------
 * array overlap/containment comparisons
 *		These use the same methods of comparing array elements as array_eq.
 *		We consider only the elements of the arrays, ignoring dimensionality.
 *----------------------------------------------------------------------------
 */

/*
 * array_contain_compare :
 *		  compares two arrays for overlap/containment
 *
 * When matchall is true, return true if all members of array1 are in array2.
 * When matchall is false, return true if any members of array1 are in array2.
 */









/*-----------------------------------------------------------------------------
 * Array iteration functions
 *		These functions are used to iterate efficiently through arrays
 *-----------------------------------------------------------------------------
 */

/*
 * array_create_iterator --- set up to iterate through an array
 *
 * If slice_ndim is zero, we will iterate element-by-element; the returned
 * datums are of the array's element type.
 *
 * If slice_ndim is 1..ARR_NDIM(arr), we will iterate by slices: the
 * returned datums are of the same array type as 'arr', but of size
 * equal to the rightmost N dimensions of 'arr'.
 *
 * The passed-in array must remain valid for the lifetime of the iterator.
 */


/*
 * Iterate through the array referenced by 'iterator'.
 *
 * As long as there is another element (or slice), return it into
 * *value / *isnull, and return true.  Return false when no more data.
 */


/*
 * Release an ArrayIterator data structure
 */



/***************************************************************************/
/******************|		  Support  Routines			  |*****************/
/***************************************************************************/

/*
 * Check whether a specific array element is NULL
 *
 * nullbitmap: pointer to array's null bitmap (NULL if none)
 * offset: 0-based linear element number of array element
 */


/*
 * Set a specific array element's null-bitmap entry
 *
 * nullbitmap: pointer to array's null bitmap (mustn't be NULL)
 * offset: 0-based linear element number of array element
 * isNull: null status to set
 */


/*
 * Fetch array element at pointer, converted correctly to a Datum
 *
 * Caller must have handled case of NULL element
 */


/*
 * Copy datum to *dest and return total space used (including align padding)
 *
 * Caller must have handled case of NULL element
 */
static int
ArrayCastAndSet(Datum src,
				int typlen,
				bool typbyval,
				char typalign,
				char *dest)
{
	int			inc;

	if (typlen > 0)
	{
		if (typbyval)
			store_att_byval(dest, src, typlen);
		else
			memmove(dest, DatumGetPointer(src), typlen);
		inc = att_align_nominal(typlen, typalign);
	}
	else
	{
		Assert(!typbyval);
		inc = att_addlength_datum(0, typlen, src);
		memmove(dest, DatumGetPointer(src), inc);
		inc = att_align_nominal(inc, typalign);
	}

	return inc;
}

/*
 * Advance ptr over nitems array elements
 *
 * ptr: starting location in array
 * offset: 0-based linear element number of first element (the one at *ptr)
 * nullbitmap: start of array's null bitmap, or NULL if none
 * nitems: number of array elements to advance over (>= 0)
 * typlen, typbyval, typalign: storage parameters of array element datatype
 *
 * It is caller's responsibility to ensure that nitems is within range
 */


/*
 * Compute total size of the nitems array elements starting at *ptr
 *
 * Parameters same as for array_seek
 */


/*
 * Copy nitems array elements from srcptr to destptr
 *
 * destptr: starting destination location (must be enough room!)
 * nitems: number of array elements to copy (>= 0)
 * srcptr: starting location in source array
 * offset: 0-based linear element number of first element (the one at *srcptr)
 * nullbitmap: start of source array's null bitmap, or NULL if none
 * typlen, typbyval, typalign: storage parameters of array element datatype
 *
 * Returns number of bytes copied
 *
 * NB: this does not take care of setting up the destination's null bitmap!
 */


/*
 * Copy nitems null-bitmap bits from source to destination
 *
 * destbitmap: start of destination array's null bitmap (mustn't be NULL)
 * destoffset: 0-based linear element number of first dest element
 * srcbitmap: start of source array's null bitmap, or NULL if none
 * srcoffset: 0-based linear element number of first source element
 * nitems: number of bits to copy (>= 0)
 *
 * If srcbitmap is NULL then we assume the source is all-non-NULL and
 * fill 1's into the destination bitmap.  Note that only the specified
 * bits in the destination map are changed, not any before or after.
 *
 * Note: this could certainly be optimized using standard bitblt methods.
 * However, it's not clear that the typical Postgres array has enough elements
 * to make it worth worrying too much.  For the moment, KISS.
 */


/*
 * Compute space needed for a slice of an array
 *
 * We assume the caller has verified that the slice coordinates are valid.
 */


/*
 * Extract a slice of an array into consecutive elements in the destination
 * array.
 *
 * We assume the caller has verified that the slice coordinates are valid,
 * allocated enough storage for the result, and initialized the header
 * of the new array.
 */


/*
 * Insert a slice into an array.
 *
 * ndim/dim[]/lb[] are dimensions of the original array.  A new array with
 * those same dimensions is to be constructed.  destArray must already
 * have been allocated and its header initialized.
 *
 * st[]/endp[] identify the slice to be replaced.  Elements within the slice
 * volume are taken from consecutive elements of the srcArray; elements
 * outside it are copied from origArray.
 *
 * We assume the caller has verified that the slice coordinates are valid.
 */


/*
 * initArrayResult - initialize an empty ArrayBuildState
 *
 *	element_type is the array element type (must be a valid array element type)
 *	rcontext is where to keep working state
 *	subcontext is a flag determining whether to use a separate memory context
 *
 * Note: there are two common schemes for using accumArrayResult().
 * In the older scheme, you start with a NULL ArrayBuildState pointer, and
 * call accumArrayResult once per element.  In this scheme you end up with
 * a NULL pointer if there were no elements, which you need to special-case.
 * In the newer scheme, call initArrayResult and then call accumArrayResult
 * once per element.  In this scheme you always end with a non-NULL pointer
 * that you can pass to makeArrayResult; you get an empty array if there
 * were no elements.  This is preferred if an empty array is what you want.
 *
 * It's possible to choose whether to create a separate memory context for the
 * array build state, or whether to allocate it directly within rcontext.
 *
 * When there are many concurrent small states (e.g. array_agg() using hash
 * aggregation of many small groups), using a separate memory context for each
 * one may result in severe memory bloat. In such cases, use the same memory
 * context to initialize all such array build states, and pass
 * subcontext=false.
 *
 * In cases when the array build states have different lifetimes, using a
 * single memory context is impractical. Instead, pass subcontext=true so that
 * the array build states can be freed individually.
 */


/*
 * initArrayResultWithSize
 *		As initArrayResult, but allow the initial size of the allocated arrays
 *		to be specified.
 */


/*
 * accumArrayResult - accumulate one (more) Datum for an array result
 *
 *	astate is working state (can be NULL on first call)
 *	dvalue/disnull represent the new Datum to append to the array
 *	element_type is the Datum's type (must be a valid array element type)
 *	rcontext is where to keep working state
 */


/*
 * makeArrayResult - produce 1-D final result of accumArrayResult
 *
 * Note: only releases astate if it was initialized within a separate memory
 * context (i.e. using subcontext=true when calling initArrayResult).
 *
 *	astate is working state (must not be NULL)
 *	rcontext is where to construct result
 */


/*
 * makeMdArrayResult - produce multi-D final result of accumArrayResult
 *
 * beware: no check that specified dimensions match the number of values
 * accumulated.
 *
 * Note: if the astate was not initialized within a separate memory context
 * (that is, initArrayResult was called with subcontext=false), then using
 * release=true is illegal. Instead, release astate along with the rest of its
 * context when appropriate.
 *
 *	astate is working state (must not be NULL)
 *	rcontext is where to construct result
 *	release is true if okay to release working state
 */


/*
 * The following three functions provide essentially the same API as
 * initArrayResult/accumArrayResult/makeArrayResult, but instead of accepting
 * inputs that are array elements, they accept inputs that are arrays and
 * produce an output array having N+1 dimensions.  The inputs must all have
 * identical dimensionality as well as element type.
 */

/*
 * initArrayResultArr - initialize an empty ArrayBuildStateArr
 *
 *	array_type is the array type (must be a valid varlena array type)
 *	element_type is the type of the array's elements (lookup if InvalidOid)
 *	rcontext is where to keep working state
 *	subcontext is a flag determining whether to use a separate memory context
 */


/*
 * accumArrayResultArr - accumulate one (more) sub-array for an array result
 *
 *	astate is working state (can be NULL on first call)
 *	dvalue/disnull represent the new sub-array to append to the array
 *	array_type is the array type (must be a valid varlena array type)
 *	rcontext is where to keep working state
 */


/*
 * makeArrayResultArr - produce N+1-D final result of accumArrayResultArr
 *
 *	astate is working state (must not be NULL)
 *	rcontext is where to construct result
 *	release is true if okay to release working state
 */


/*
 * The following three functions provide essentially the same API as
 * initArrayResult/accumArrayResult/makeArrayResult, but can accept either
 * scalar or array inputs, invoking the appropriate set of functions above.
 */

/*
 * initArrayResultAny - initialize an empty ArrayBuildStateAny
 *
 *	input_type is the input datatype (either element or array type)
 *	rcontext is where to keep working state
 *	subcontext is a flag determining whether to use a separate memory context
 */


/*
 * accumArrayResultAny - accumulate one (more) input for an array result
 *
 *	astate is working state (can be NULL on first call)
 *	dvalue/disnull represent the new input to append to the array
 *	input_type is the input datatype (either element or array type)
 *	rcontext is where to keep working state
 */


/*
 * makeArrayResultAny - produce final result of accumArrayResultAny
 *
 *	astate is working state (must not be NULL)
 *	rcontext is where to construct result
 *	release is true if okay to release working state
 */








typedef struct generate_subscripts_fctx
{
	int32		lower;
	int32		upper;
	bool		reverse;
} generate_subscripts_fctx;

/*
 * generate_subscripts(array anyarray, dim int [, reverse bool])
 *		Returns all subscripts of the array for any dimension
 */


/*
 * generate_subscripts_nodir
 *		Implements the 2-argument version of generate_subscripts
 */


/*
 * array_fill_with_lower_bounds
 *		Create and fill array with defined lower bounds.
 */


/*
 * array_fill
 *		Create and fill array with default lower bounds.
 */







/*
 * UNNEST
 */


/*
 * Planner support function for array_unnest(anyarray)
 *
 * Note: this is now also used for information_schema._pg_expandarray(),
 * which is simply a wrapper around array_unnest().
 */



/*
 * array_replace/array_remove support
 *
 * Find all array entries matching (not distinct from) search/search_isnull,
 * and delete them if remove is true, else replace them with
 * replace/replace_isnull.  Comparisons are done using the specified
 * collation.  fcinfo is passed only for caching purposes.
 */


/*
 * Remove any occurrences of an element from an array
 *
 * If used on a multi-dimensional array this will raise an error.
 */


/*
 * Replace any occurrences of an element in an array
 */


/*
 * Implements width_bucket(anyelement, anyarray).
 *
 * 'thresholds' is an array containing lower bound values for each bucket;
 * these must be sorted from smallest to largest, or bogus results will be
 * produced.  If N thresholds are supplied, the output is from 0 to N:
 * 0 is for inputs < first threshold, N is for inputs >= last threshold.
 */


/*
 * width_bucket_array for float8 data.
 */


/*
 * width_bucket_array for generic fixed-width data types.
 */


/*
 * width_bucket_array for generic variable-width data types.
 */


/*
 * Trim the last N elements from an array by building an appropriate slice.
 * Only the first dimension is trimmed.
 */

