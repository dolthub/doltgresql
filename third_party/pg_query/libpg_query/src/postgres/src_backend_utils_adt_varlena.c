/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - text_to_cstring
 * - cstring_to_text
 * - cstring_to_text_with_len
 * - text_to_cstring
 * - levenshtein_result
 * - rest_of_char_same
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * varlena.c
 *	  Functions for the variable-length built-in types.
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 *
 * IDENTIFICATION
 *	  src/backend/utils/adt/varlena.c
 *
 *-------------------------------------------------------------------------
 */
#include "postgres.h"

#include <ctype.h>
#include <limits.h>

#include "access/detoast.h"
#include "access/toast_compression.h"
#include "catalog/pg_collation.h"
#include "catalog/pg_type.h"
#include "common/hashfn.h"
#include "common/int.h"
#include "common/unicode_category.h"
#include "common/unicode_norm.h"
#include "common/unicode_version.h"
#include "funcapi.h"
#include "lib/hyperloglog.h"
#include "libpq/pqformat.h"
#include "miscadmin.h"
#include "nodes/execnodes.h"
#include "parser/scansup.h"
#include "port/pg_bswap.h"
#include "regex/regex.h"
#include "utils/builtins.h"
#include "utils/bytea.h"
#include "utils/guc.h"
#include "utils/lsyscache.h"
#include "utils/memutils.h"
#include "utils/pg_locale.h"
#include "utils/sortsupport.h"
#include "utils/varlena.h"


/* GUC variable */


typedef struct varlena VarString;

/*
 * State for text_position_* functions.
 */
typedef struct
{
	pg_locale_t locale;			/* collation used for substring matching */
	bool		is_multibyte_char_in_char;	/* need to check char boundaries? */
	bool		greedy;			/* find longest possible substring? */

	char	   *str1;			/* haystack string */
	char	   *str2;			/* needle string */
	int			len1;			/* string lengths in bytes */
	int			len2;

	/* Skip table for Boyer-Moore-Horspool search algorithm: */
	int			skiptablemask;	/* mask for ANDing with skiptable subscripts */
	int			skiptable[256]; /* skip distance for given mismatched char */

	/*
	 * Note that with nondeterministic collations, the length of the last
	 * match is not necessarily equal to the length of the "needle" passed in.
	 */
	char	   *last_match;		/* pointer to last match in 'str1' */
	int			last_match_len; /* length of last match */
	int			last_match_len_tmp; /* same but for internal use */

	/*
	 * Sometimes we need to convert the byte position of a match to a
	 * character position.  These store the last position that was converted,
	 * so that on the next call, we can continue from that point, rather than
	 * count characters from the very beginning.
	 */
	char	   *refpoint;		/* pointer within original haystack string */
	int			refpos;			/* 0-based character offset of the same point */
} TextPositionState;

typedef struct
{
	char	   *buf1;			/* 1st string, or abbreviation original string
								 * buf */
	char	   *buf2;			/* 2nd string, or abbreviation strxfrm() buf */
	int			buflen1;		/* Allocated length of buf1 */
	int			buflen2;		/* Allocated length of buf2 */
	int			last_len1;		/* Length of last buf1 string/strxfrm() input */
	int			last_len2;		/* Length of last buf2 string/strxfrm() blob */
	int			last_returned;	/* Last comparison result (cache) */
	bool		cache_blob;		/* Does buf2 contain strxfrm() blob, etc? */
	bool		collate_c;
	Oid			typid;			/* Actual datatype (text/bpchar/bytea/name) */
	hyperLogLogState abbr_card; /* Abbreviated key cardinality state */
	hyperLogLogState full_card; /* Full key cardinality state */
	double		prop_card;		/* Required cardinality proportion */
	pg_locale_t locale;
} VarStringSortSupport;

/*
 * Output data for split_text(): we output either to an array or a table.
 * tupstore and tupdesc must be set up in advance to output to a table.
 */
typedef struct
{
	ArrayBuildState *astate;
	Tuplestorestate *tupstore;
	TupleDesc	tupdesc;
} SplitTextOutputData;

/*
 * This should be large enough that most strings will fit, but small enough
 * that we feel comfortable putting it on the stack
 */
#define TEXTBUFLEN		1024

#define DatumGetVarStringP(X)		((VarString *) PG_DETOAST_DATUM(X))
#define DatumGetVarStringPP(X)		((VarString *) PG_DETOAST_DATUM_PACKED(X))

static int	varstrfastcmp_c(Datum x, Datum y, SortSupport ssup);
static int	bpcharfastcmp_c(Datum x, Datum y, SortSupport ssup);
static int	namefastcmp_c(Datum x, Datum y, SortSupport ssup);
static int	varlenafastcmp_locale(Datum x, Datum y, SortSupport ssup);
static int	namefastcmp_locale(Datum x, Datum y, SortSupport ssup);
static int	varstrfastcmp_locale(char *a1p, int len1, char *a2p, int len2, SortSupport ssup);
static Datum varstr_abbrev_convert(Datum original, SortSupport ssup);
static bool varstr_abbrev_abort(int memtupcount, SortSupport ssup);
static int32 text_length(Datum str);
static text *text_catenate(text *t1, text *t2);
static text *text_substring(Datum str,
							int32 start,
							int32 length,
							bool length_not_specified);
static int	pg_mbcharcliplen_chars(const char *mbstr, int len, int limit);
static text *text_overlay(text *t1, text *t2, int sp, int sl);
static int	text_position(text *t1, text *t2, Oid collid);
static void text_position_setup(text *t1, text *t2, Oid collid, TextPositionState *state);
static bool text_position_next(TextPositionState *state);
static char *text_position_next_internal(char *start_ptr, TextPositionState *state);
static char *text_position_get_match_ptr(TextPositionState *state);
static int	text_position_get_match_pos(TextPositionState *state);
static void text_position_cleanup(TextPositionState *state);
static void check_collation_set(Oid collid);
static int	text_cmp(text *arg1, text *arg2, Oid collid);
static bytea *bytea_catenate(bytea *t1, bytea *t2);
static bytea *bytea_substring(Datum str,
							  int S,
							  int L,
							  bool length_not_specified);
static bytea *bytea_overlay(bytea *t1, bytea *t2, int sp, int sl);
static void appendStringInfoText(StringInfo str, const text *t);
static bool split_text(FunctionCallInfo fcinfo, SplitTextOutputData *tstate);
static void split_text_accum_result(SplitTextOutputData *tstate,
									text *field_value,
									text *null_string,
									Oid collation);
static text *array_to_text_internal(FunctionCallInfo fcinfo, ArrayType *v,
									const char *fldsep, const char *null_string);
static StringInfo makeStringAggState(FunctionCallInfo fcinfo);
static bool text_format_parse_digits(const char **ptr, const char *end_ptr,
									 int *value);
static const char *text_format_parse_format(const char *start_ptr,
											const char *end_ptr,
											int *argpos, int *widthpos,
											int *flags, int *width);
static void text_format_string_conversion(StringInfo buf, char conversion,
										  FmgrInfo *typOutputInfo,
										  Datum value, bool isNull,
										  int flags, int width);
static void text_format_append_string(StringInfo buf, const char *str,
									  int flags, int width);


/*****************************************************************************
 *	 CONVERSION ROUTINES EXPORTED FOR USE BY C CODE							 *
 *****************************************************************************/

/*
 * cstring_to_text
 *
 * Create a text value from a null-terminated C string.
 *
 * The new text value is freshly palloc'd with a full-size VARHDR.
 */
text *
cstring_to_text(const char *s)
{
	return cstring_to_text_with_len(s, strlen(s));
}

/*
 * cstring_to_text_with_len
 *
 * Same as cstring_to_text except the caller specifies the string length;
 * the string need not be null_terminated.
 */
text *
cstring_to_text_with_len(const char *s, int len)
{
	text	   *result = (text *) palloc(len + VARHDRSZ);

	SET_VARSIZE(result, len + VARHDRSZ);
	memcpy(VARDATA(result), s, len);

	return result;
}

/*
 * text_to_cstring
 *
 * Create a palloc'd, null-terminated C string from a text value.
 *
 * We support being passed a compressed or toasted text value.
 * This is a bit bogus since such values shouldn't really be referred to as
 * "text *", but it seems useful for robustness.  If we didn't handle that
 * case here, we'd need another routine that did, anyway.
 */
char *
text_to_cstring(const text *t)
{
	/* must cast away the const, unfortunately */
	text	   *tunpacked = pg_detoast_datum_packed(unconstify(text *, t));
	int			len = VARSIZE_ANY_EXHDR(tunpacked);
	char	   *result;

	result = (char *) palloc(len + 1);
	memcpy(result, VARDATA_ANY(tunpacked), len);
	result[len] = '\0';

	if (tunpacked != t)
		pfree(tunpacked);

	return result;
}

/*
 * text_to_cstring_buffer
 *
 * Copy a text value into a caller-supplied buffer of size dst_len.
 *
 * The text string is truncated if necessary to fit.  The result is
 * guaranteed null-terminated (unless dst_len == 0).
 *
 * We support being passed a compressed or toasted text value.
 * This is a bit bogus since such values shouldn't really be referred to as
 * "text *", but it seems useful for robustness.  If we didn't handle that
 * case here, we'd need another routine that did, anyway.
 */



/*****************************************************************************
 *	 USER I/O ROUTINES														 *
 *****************************************************************************/


#define VAL(CH)			((CH) - '0')
#define DIG(VAL)		((VAL) + '0')

/*
 *		byteain			- converts from printable representation of byte array
 *
 *		Non-printable characters must be passed as '\nnn' (octal) and are
 *		converted to internal form.  '\' must be passed as '\\'.
 *		ereport(ERROR, ...) if bad form.
 *
 *		BUGS:
 *				The input is scanned twice.
 *				The error checking of input is minimal.
 */


/*
 *		byteaout		- converts to printable representation of byte array
 *
 *		In the traditional escaped format, non-printable characters are
 *		printed as '\nnn' (octal) and '\' as '\\'.
 */


/*
 *		bytearecv			- converts external binary format to bytea
 */


/*
 *		byteasend			- converts bytea to binary format
 *
 * This is a special case: just copy the input...
 */






/*
 *		textin			- converts cstring to internal representation
 */


/*
 *		textout			- converts internal representation to cstring
 */


/*
 *		textrecv			- converts external binary format to text
 */


/*
 *		textsend			- converts text to binary format
 */



/*
 *		unknownin			- converts cstring to internal representation
 */


/*
 *		unknownout			- converts internal representation to cstring
 */


/*
 *		unknownrecv			- converts external binary format to unknown
 */


/*
 *		unknownsend			- converts unknown to binary format
 */



/* ========== PUBLIC ROUTINES ========== */

/*
 * textlen -
 *	  returns the logical length of a text*
 *	   (which is less than the VARSIZE of the text*)
 */


/*
 * text_length -
 *	Does the real work for textlen()
 *
 *	This is broken out so it can be called directly by other string processing
 *	functions.  Note that the argument is passed as a Datum, to indicate that
 *	it may still be in compressed form.  We can avoid decompressing it at all
 *	in some cases.
 */


/*
 * textoctetlen -
 *	  returns the physical length of a text*
 *	   (which is less than the VARSIZE of the text*)
 */


/*
 * textcat -
 *	  takes two text* and returns a text* that is the concatenation of
 *	  the two.
 *
 * Rewritten by Sapa, sapa@hq.icb.chel.su. 8-Jul-96.
 * Updated by Thomas, Thomas.Lockhart@jpl.nasa.gov 1997-07-10.
 * Allocate space for output in all cases.
 * XXX - thomas 1997-07-10
 */


/*
 * text_catenate
 *	Guts of textcat(), broken out so it can be used by other functions
 *
 * Arguments can be in short-header form, but not compressed or out-of-line
 */


/*
 * charlen_to_bytelen()
 *	Compute the number of bytes occupied by n characters starting at *p
 *
 * The caller shall ensure there are n complete characters.  Callers achieve
 * this by deriving "n" from regmatch_t findings from searching a wchar array.
 * pg_mb2wchar_with_len() skips any trailing incomplete character, so regex
 * matches will end no later than the last complete character.  (The string
 * need not be null-terminated.)
 */


/*
 * text_substr()
 * Return a substring starting at the specified position.
 * - thomas 1997-12-31
 *
 * Input:
 *	- string
 *	- starting position (is one-based)
 *	- string length
 *
 * If the starting position is zero or less, then return from the start of the string
 *	adjusting the length to be consistent with the "negative start" per SQL.
 * If the length is less than zero, return the remaining string.
 *
 * Added multibyte support.
 * - Tatsuo Ishii 1998-4-21
 * Changed behavior if starting position is less than one to conform to SQL behavior.
 * Formerly returned the entire string; now returns a portion.
 * - Thomas Lockhart 1998-12-10
 * Now uses faster TOAST-slicing interface
 * - John Gray 2002-02-22
 * Remove "#ifdef MULTIBYTE" and test for encoding_max_length instead. Change
 * behaviors conflicting with SQL to meet SQL (if E = S + L < S throw
 * error; if E < 1, return '', not entire string). Fixed MB related bug when
 * S > LC and < LC + 4 sometimes garbage characters are returned.
 * - Joe Conway 2002-08-10
 */


/*
 * text_substr_no_len -
 *	  Wrapper to avoid opr_sanity failure due to
 *	  one function accepting a different number of args.
 */


/*
 * text_substring -
 *	Does the real work for text_substr() and text_substr_no_len()
 *
 *	This is broken out so it can be called directly by other string processing
 *	functions.  Note that the argument is passed as a Datum, to indicate that
 *	it may still be in compressed/toasted form.  We can avoid detoasting all
 *	of it in some cases.
 *
 *	The result is always a freshly palloc'd datum.
 */


/*
 * pg_mbcharcliplen_chars -
 *	Mirror pg_mbcharcliplen(), except return value unit is chars, not bytes.
 *
 *	This mirrors all the dubious historical behavior, so it's static to
 *	discourage proliferation.  The assertions are specific to the one caller.
 */


/*
 * textoverlay
 *	Replace specified substring of first string with second
 *
 * The SQL standard defines OVERLAY() in terms of substring and concatenation.
 * This code is a direct implementation of what the standard says.
 */






/*
 * textpos -
 *	  Return the position of the specified substring.
 *	  Implements the SQL POSITION() function.
 *	  Ref: A Guide To The SQL Standard, Date & Darwen, 1997
 * - thomas 1997-07-27
 */


/*
 * text_position -
 *	Does the real work for textpos()
 *
 * Inputs:
 *		t1 - string to be searched
 *		t2 - pattern to match within t1
 * Result:
 *		Character index of the first matched char, starting from 1,
 *		or 0 if no match.
 *
 *	This is broken out so it can be called directly by other string processing
 *	functions.
 */



/*
 * text_position_setup, text_position_next, text_position_cleanup -
 *	Component steps of text_position()
 *
 * These are broken out so that a string can be efficiently searched for
 * multiple occurrences of the same pattern.  text_position_next may be
 * called multiple times, and it advances to the next match on each call.
 * text_position_get_match_ptr() and text_position_get_match_pos() return
 * a pointer or 1-based character position of the last match, respectively.
 *
 * The "state" variable is normally just a local variable in the caller.
 *
 * NOTE: text_position_next skips over the matched portion.  For example,
 * searching for "xx" in "xxx" returns only one match, not two.
 */



/*
 * Advance to the next match, starting from the end of the previous match
 * (or the beginning of the string, on first call).  Returns true if a match
 * is found.
 *
 * Note that this refuses to match an empty-string needle.  Most callers
 * will have handled that case specially and we'll never see it here.
 */


/*
 * Subroutine of text_position_next().  This searches for the raw byte
 * sequence, ignoring any multi-byte encoding issues.  Returns the first
 * match starting at 'start_ptr', or NULL if no match is found.
 */


/*
 * Return a pointer to the current match.
 *
 * The returned pointer points into the original haystack string.
 */


/*
 * Return the offset of the current match.
 *
 * The offset is in characters, 1-based.
 */


/*
 * Reset search state to the initial state installed by text_position_setup.
 *
 * The next call to text_position_next will search from the beginning
 * of the string.
 */







/*
 * varstr_cmp()
 *
 * Comparison function for text strings with given lengths, using the
 * appropriate locale. Returns an integer less than, equal to, or greater than
 * zero, indicating whether arg1 is less than, equal to, or greater than arg2.
 *
 * Note: many functions that depend on this are marked leakproof; therefore,
 * avoid reporting the actual contents of the input when throwing errors.
 * All errors herein should be things that can't happen except on corrupt
 * data, anyway; otherwise we will have trouble with indexing strings that
 * would cause them.
 */


/* text_cmp()
 * Internal comparison function for text strings.
 * Returns -1, 0 or 1
 */


/*
 * Comparison functions for text strings.
 *
 * Note: btree indexes need these routines not to leak memory; therefore,
 * be careful to free working copies of toasted datums.  Most places don't
 * need to be so careful.
 */



















/*
 * Generic sortsupport interface for character type's operator classes.
 * Includes locale support, and support for BpChar semantics (i.e. removing
 * trailing spaces before comparison).
 *
 * Relies on the assumption that text, VarChar, BpChar, and bytea all have the
 * same representation.  Callers that always use the C collation (e.g.
 * non-collatable type callers like bytea) may have NUL bytes in their strings;
 * this will not work with any other collation, though.
 */


/*
 * sortsupport comparison func (for C locale case)
 */


/*
 * sortsupport comparison func (for BpChar C locale case)
 *
 * BpChar outsources its sortsupport to this module.  Specialization for the
 * varstr_sortsupport BpChar case, modeled on
 * internal_bpchar_pattern_compare().
 */


/*
 * sortsupport comparison func (for NAME C locale case)
 */


/*
 * sortsupport comparison func (for locale case with all varlena types)
 */


/*
 * sortsupport comparison func (for locale case with NAME type)
 */


/*
 * sortsupport comparison func for locale cases
 */


/*
 * Conversion routine for sortsupport.  Converts original to abbreviated key
 * representation.  Our encoding strategy is simple -- pack the first 8 bytes
 * of a strxfrm() blob into a Datum (on little-endian machines, the 8 bytes are
 * stored in reverse order), and treat it as an unsigned integer.  When the "C"
 * locale is used, or in case of bytea, just memcpy() from original instead.
 */
#if SIZEOF_DATUM == 8
#else							/* SIZEOF_DATUM != 8 */
#endif

/*
 * Callback for estimating effectiveness of abbreviated key optimization, using
 * heuristic rules.  Returns value indicating if the abbreviation optimization
 * should be aborted, based on its projected effectiveness.
 */


/*
 * Generic equalimage support function for character type's operator classes.
 * Disables the use of deduplication with nondeterministic collations.
 */







/*
 * Cross-type comparison functions for types text and name.
 */













#define CmpCall(cmpfunc) \
	DatumGetInt32(DirectFunctionCall2Coll(cmpfunc, \
										  PG_GET_COLLATION(), \
										  PG_GETARG_DATUM(0), \
										  PG_GETARG_DATUM(1)))

















#undef CmpCall


/*
 * The following operators support character-by-character comparison
 * of text datums, to allow building indexes suitable for LIKE clauses.
 * Note that the regular texteq/textne comparison operators, and regular
 * support functions 1 and 2 with "C" collation are assumed to be
 * compatible with these!
 */






















/*-------------------------------------------------------------
 * byteaoctetlen
 *
 * get the number of bytes contained in an instance of type 'bytea'
 *-------------------------------------------------------------
 */


/*
 * byteacat -
 *	  takes two bytea* and returns a bytea* that is the concatenation of
 *	  the two.
 *
 * Cloned from textcat and modified as required.
 */


/*
 * bytea_catenate
 *	Guts of byteacat(), broken out so it can be used by other functions
 *
 * Arguments can be in short-header form, but not compressed or out-of-line
 */


#define PG_STR_GET_BYTEA(str_) \
	DatumGetByteaPP(DirectFunctionCall1(byteain, CStringGetDatum(str_)))

/*
 * bytea_substr()
 * Return a substring starting at the specified position.
 * Cloned from text_substr and modified as required.
 *
 * Input:
 *	- string
 *	- starting position (is one-based)
 *	- string length (optional)
 *
 * If the starting position is zero or less, then return from the start of the string
 * adjusting the length to be consistent with the "negative start" per SQL.
 * If the length is less than zero, an ERROR is thrown. If no third argument
 * (length) is provided, the length to the end of the string is assumed.
 */


/*
 * bytea_substr_no_len -
 *	  Wrapper to avoid opr_sanity failure due to
 *	  one function accepting a different number of args.
 */




/*
 * byteaoverlay
 *	Replace specified substring of first string with second
 *
 * The SQL standard defines OVERLAY() in terms of substring and concatenation.
 * This code is a direct implementation of what the standard says.
 */






/*
 * bit_count
 */


/*
 * byteapos -
 *	  Return the position of the specified substring.
 *	  Implements the SQL POSITION() function.
 * Cloned from textpos and modified as required.
 */


/*-------------------------------------------------------------
 * byteaGetByte
 *
 * this routine treats "bytea" as an array of bytes.
 * It returns the Nth byte (a number between 0 and 255).
 *-------------------------------------------------------------
 */


/*-------------------------------------------------------------
 * byteaGetBit
 *
 * This routine treats a "bytea" type like an array of bits.
 * It returns the value of the Nth bit (0 or 1).
 *
 *-------------------------------------------------------------
 */


/*-------------------------------------------------------------
 * byteaSetByte
 *
 * Given an instance of type 'bytea' creates a new one with
 * the Nth byte set to the given value.
 *
 *-------------------------------------------------------------
 */


/*-------------------------------------------------------------
 * byteaSetBit
 *
 * Given an instance of type 'bytea' creates a new one with
 * the Nth bit set to the given value.
 *
 *-------------------------------------------------------------
 */


/*
 * Return reversed bytea
 */



/* text_name()
 * Converts a text type to a Name type.
 */


/* name_text()
 * Converts a Name type to a text type.
 */



/*
 * textToQualifiedNameList - convert a text object to list of names
 *
 * This implements the input parsing needed by nextval() and other
 * functions that take a text parameter representing a qualified name.
 * We split the name at dots, downcase if not double-quoted, and
 * truncate names if they're too long.
 */


/*
 * SplitIdentifierString --- parse a string containing identifiers
 *
 * This is the guts of textToQualifiedNameList, and is exported for use in
 * other situations such as parsing GUC variables.  In the GUC case, it's
 * important to avoid memory leaks, so the API is designed to minimize the
 * amount of stuff that needs to be allocated and freed.
 *
 * Inputs:
 *	rawstring: the input string; must be overwritable!	On return, it's
 *			   been modified to contain the separated identifiers.
 *	separator: the separator punctuation expected between identifiers
 *			   (typically '.' or ',').  Whitespace may also appear around
 *			   identifiers.
 * Outputs:
 *	namelist: filled with a palloc'd list of pointers to identifiers within
 *			  rawstring.  Caller should list_free() this even on error return.
 *
 * Returns true if okay, false if there is a syntax error in the string.
 *
 * Note that an empty string is considered okay here, though not in
 * textToQualifiedNameList.
 */



/*
 * SplitDirectoriesString --- parse a string containing file/directory names
 *
 * This works fine on file names too; the function name is historical.
 *
 * This is similar to SplitIdentifierString, except that the parsing
 * rules are meant to handle pathnames instead of identifiers: there is
 * no downcasing, embedded spaces are allowed, the max length is MAXPGPATH-1,
 * and we apply canonicalize_path() to each extracted string.  Because of the
 * last, the returned strings are separately palloc'd rather than being
 * pointers into rawstring --- but we still scribble on rawstring.
 *
 * Inputs:
 *	rawstring: the input string; must be modifiable!
 *	separator: the separator punctuation expected between directories
 *			   (typically ',' or ';').  Whitespace may also appear around
 *			   directories.
 * Outputs:
 *	namelist: filled with a palloc'd list of directory names.
 *			  Caller should list_free_deep() this even on error return.
 *
 * Returns true if okay, false if there is a syntax error in the string.
 *
 * Note that an empty string is considered okay here.
 */



/*
 * SplitGUCList --- parse a string containing identifiers or file names
 *
 * This is used to split the value of a GUC_LIST_QUOTE GUC variable, without
 * presuming whether the elements will be taken as identifiers or file names.
 * We assume the input has already been through flatten_set_variable_args(),
 * so that we need never downcase (if appropriate, that was done already).
 * Nor do we ever truncate, since we don't know the correct max length.
 * We disallow embedded whitespace for simplicity (it shouldn't matter,
 * because any embedded whitespace should have led to double-quoting).
 * Otherwise the API is identical to SplitIdentifierString.
 *
 * XXX it's annoying to have so many copies of this string-splitting logic.
 * However, it's not clear that having one function with a bunch of option
 * flags would be much better.
 *
 * XXX there is a version of this function in src/fe_utils/string_utils.c.
 * Be sure to update that if you have to change this.
 *
 * Inputs:
 *	rawstring: the input string; must be overwritable!	On return, it's
 *			   been modified to contain the separated identifiers.
 *	separator: the separator punctuation expected between identifiers
 *			   (typically '.' or ',').  Whitespace may also appear around
 *			   identifiers.
 * Outputs:
 *	namelist: filled with a palloc'd list of pointers to identifiers within
 *			  rawstring.  Caller should list_free() this even on error return.
 *
 * Returns true if okay, false if there is a syntax error in the string.
 */



/*****************************************************************************
 *	Comparison Functions used for bytea
 *
 * Note: btree indexes need these routines not to leak memory; therefore,
 * be careful to free working copies of toasted datums.  Most places don't
 * need to be so careful.
 *****************************************************************************/





















/* Cast bytea -> int2 */


/* Cast bytea -> int4 */


/* Cast bytea -> int8 */


/* Cast int2 -> bytea; can just use int2send() */


/* Cast int4 -> bytea; can just use int4send() */


/* Cast int8 -> bytea; can just use int8send() */


/*
 * appendStringInfoText
 *
 * Append a text to str.
 * Like appendStringInfoString(str, text_to_cstring(t)) but faster.
 */


/*
 * replace_text
 * replace all occurrences of 'old_sub_str' in 'orig_str'
 * with 'new_sub_str' to form 'new_str'
 *
 * returns 'orig_str' if 'old_sub_str' == '' or 'orig_str' == ''
 * otherwise returns 'new_str'
 */


/*
 * check_replace_text_has_escape
 *
 * Returns 0 if text contains no backslashes that need processing.
 * Returns 1 if text contains backslashes, but not regexp submatch specifiers.
 * Returns 2 if text contains regexp submatch specifiers (\1 .. \9).
 */


/*
 * appendStringInfoRegexpSubstr
 *
 * Append replace_text to str, substituting regexp back references for
 * \n escapes.  start_ptr is the start of the match in the source string,
 * at logical character position data_pos.
 */


/*
 * replace_text_regexp
 *
 * replace substring(s) in src_text that match pattern with replace_text.
 * The replace_text can contain backslash markers to substitute
 * (parts of) the matched text.
 *
 * cflags: regexp compile flags.
 * collation: collation to use.
 * search_start: the character (not byte) offset in src_text at which to
 * begin searching.
 * n: if 0, replace all matches; if > 0, replace only the N'th match.
 */


/*
 * split_part
 * parse input string based on provided field separator
 * return N'th item (1 based, negative counts from end)
 */


/*
 * Convenience function to return true when two text params are equal.
 */


/*
 * text_to_array
 * parse input string and return text array of elements,
 * based on provided field separator
 */


/*
 * text_to_array_null
 * parse input string and return text array of elements,
 * based on provided field separator and null string
 *
 * This is a separate entry point only to prevent the regression tests from
 * complaining about different argument sets for the same internal function.
 */


/*
 * text_to_table
 * parse input string and return table of elements,
 * based on provided field separator
 */


/*
 * text_to_table_null
 * parse input string and return table of elements,
 * based on provided field separator and null string
 *
 * This is a separate entry point only to prevent the regression tests from
 * complaining about different argument sets for the same internal function.
 */


/*
 * Common code for text_to_array, text_to_array_null, text_to_table
 * and text_to_table_null functions.
 *
 * These are not strict so we have to test for null inputs explicitly.
 * Returns false if result is to be null, else returns true.
 *
 * Note that if the result is valid but empty (zero elements), we return
 * without changing *tstate --- caller must handle that case, too.
 */


/*
 * Add text item to result set (table or array).
 *
 * This is also responsible for checking to see if the item matches
 * the null_string, in which case we should emit NULL instead.
 */


/*
 * array_to_text
 * concatenate Cstring representation of input array elements
 * using provided field separator
 */


/*
 * array_to_text_null
 * concatenate Cstring representation of input array elements
 * using provided field separator and null string
 *
 * This version is not strict so we have to test for null inputs explicitly.
 */


/*
 * common code for array_to_text and array_to_text_null functions
 */


/*
 * Workhorse for to_bin, to_oct, and to_hex.  Note that base must be > 1 and <=
 * 16.
 */


/*
 * Convert an integer to a string containing a base-2 (binary) representation
 * of the number.
 */



/*
 * Convert an integer to a string containing a base-8 (oct) representation of
 * the number.
 */



/*
 * Convert an integer to a string containing a base-16 (hex) representation of
 * the number.
 */



/*
 * Return the size of a datum, possibly compressed
 *
 * Works on any data type
 */


/*
 * Return the compression method stored in the compressed attribute.  Return
 * NULL for non varlena type or uncompressed data.
 */


/*
 * Return the chunk_id of the on-disk TOASTed value.  Return NULL if the value
 * is un-TOASTed or not on-disk.
 */


/*
 * string_agg - Concatenates values and returns string.
 *
 * Syntax: string_agg(value text, delimiter text) RETURNS text
 *
 * Note: Any NULL values are ignored. The first-call delimiter isn't
 * actually used at all, and on subsequent calls the delimiter precedes
 * the associated value.
 */

/* subroutine to initialize state */




/*
 * string_agg_combine
 *		Aggregate combine function for string_agg(text) and string_agg(bytea)
 */


/*
 * string_agg_serialize
 *		Aggregate serialize function for string_agg(text) and string_agg(bytea)
 *
 * This is strict, so we need not handle NULL input
 */


/*
 * string_agg_deserialize
 *		Aggregate deserial function for string_agg(text) and string_agg(bytea)
 *
 * This is strict, so we need not handle NULL input
 */




/*
 * Prepare cache with fmgr info for the output functions of the datatypes of
 * the arguments of a concat-like function, beginning with argument "argidx".
 * (Arguments before that will have corresponding slots in the resulting
 * FmgrInfo array, but we don't fill those slots.)
 */


/*
 * Implementation of both concat() and concat_ws().
 *
 * sepstr is the separator string to place between values.
 * argidx identifies the first argument to concatenate (counting from zero);
 * note that this must be constant across any one series of calls.
 *
 * Returns NULL if result should be NULL, else text value.
 */


/*
 * Concatenate all arguments. NULL arguments are ignored.
 */


/*
 * Concatenate all but first argument value with separators. The first
 * parameter is used as the separator. NULL arguments are ignored.
 */


/*
 * Return first n characters in the string. When n is negative,
 * return all but last |n| characters.
 */


/*
 * Return last n characters in the string. When n is negative,
 * return all but first |n| characters.
 */


/*
 * Return reversed string
 */



/*
 * Support macros for text_format()
 */
#define TEXT_FORMAT_FLAG_MINUS	0x0001	/* is minus flag present? */

#define ADVANCE_PARSE_POINTER(ptr,end_ptr) \
	do { \
		if (++(ptr) >= (end_ptr)) \
			ereport(ERROR, \
					(errcode(ERRCODE_INVALID_PARAMETER_VALUE), \
					 errmsg("unterminated format() type specifier"), \
					 errhint("For a single \"%%\" use \"%%%%\"."))); \
	} while (0)

/*
 * Returns a formatted string
 */


/*
 * Parse contiguous digits as a decimal number.
 *
 * Returns true if some digits could be parsed.
 * The value is returned into *value, and *ptr is advanced to the next
 * character to be parsed.
 *
 * Note parsing invariant: at least one character is known available before
 * string end (end_ptr) at entry, and this is still true at exit.
 */


/*
 * Parse a format specifier (generally following the SUS printf spec).
 *
 * We have already advanced over the initial '%', and we are looking for
 * [argpos][flags][width]type (but the type character is not consumed here).
 *
 * Inputs are start_ptr (the position after '%') and end_ptr (string end + 1).
 * Output parameters:
 *	argpos: argument position for value to be printed.  -1 means unspecified.
 *	widthpos: argument position for width.  Zero means the argument position
 *			was unspecified (ie, take the next arg) and -1 means no width
 *			argument (width was omitted or specified as a constant).
 *	flags: bitmask of flags.
 *	width: directly-specified width value.  Zero means the width was omitted
 *			(note it's not necessary to distinguish this case from an explicit
 *			zero width value).
 *
 * The function result is the next character position to be parsed, ie, the
 * location where the type character is/should be.
 *
 * Note parsing invariant: at least one character is known available before
 * string end (end_ptr) at entry, and this is still true at exit.
 */


/*
 * Format a %s, %I, or %L conversion
 */


/*
 * Append str to buf, padding as directed by flags/width
 */


/*
 * text_format_nv - nonvariadic wrapper for text_format function.
 *
 * note: this wrapper is necessary to pass the sanity check in opr_sanity,
 * which checks that all built-in functions that share the implementing C
 * function take the same number of arguments.
 */


/*
 * Helper function for Levenshtein distance functions. Faster than memcmp(),
 * for this use case.
 */
static inline bool
rest_of_char_same(const char *s1, const char *s2, int len)
{
	while (len > 0)
	{
		len--;
		if (s1[len] != s2[len])
			return false;
	}
	return true;
}

/*
 * Helper function for checking return value of Levenshtein distance functions.
 * We calculate it as an int64, but the distance functions return an int32.
 */
static inline int
levenshtein_result(int64 res)
{
	if (unlikely(res < PG_INT32_MIN || res > PG_INT32_MAX))
		ereport(ERROR,
				(errcode(ERRCODE_NUMERIC_VALUE_OUT_OF_RANGE),
				 errmsg("levenshtein distance out of range")));
	return res;
}

/* Expand each Levenshtein distance variant */
#include "levenshtein.c"
#define LEVENSHTEIN_LESS_EQUAL
#include "levenshtein.c"


/*
 * The following *ClosestMatch() functions can be used to determine whether a
 * user-provided string resembles any known valid values, which is useful for
 * providing hints in log messages, among other things.  Use these functions
 * like so:
 *
 *		initClosestMatch(&state, source_string, max_distance);
 *
 *		for (int i = 0; i < num_valid_strings; i++)
 *			updateClosestMatch(&state, valid_strings[i]);
 *
 *		closestMatch = getClosestMatch(&state);
 */

/*
 * Initialize the given state with the source string and maximum Levenshtein
 * distance to consider.
 */


/*
 * If the candidate string is a closer match than the current one saved (or
 * there is no match saved), save it as the closest match.
 *
 * If the source or candidate string is NULL, empty, or too long, this function
 * takes no action.  Likewise, if the Levenshtein distance exceeds the maximum
 * allowed or more than half the characters are different, no action is taken.
 */


/*
 * Return the closest match.  If no suitable candidates were provided via
 * updateClosestMatch(), return NULL.
 */



/*
 * Unicode support
 */



/*
 * Returns version of Unicode used by Postgres in "major.minor" format (the
 * same format as the Unicode version reported by ICU). The third component
 * ("update version") never involves additions to the character repertoire and
 * is unimportant for most purposes.
 *
 * See: https://unicode.org/versions/
 */


/*
 * Returns version of Unicode used by ICU, if enabled; otherwise NULL.
 */
#ifdef USE_ICU
#else
#endif

/*
 * Check whether the string contains only assigned Unicode code
 * points. Requires that the database encoding is UTF-8.
 */




/*
 * Check whether the string is in the specified Unicode normalization form.
 *
 * This is done by converting the string to the specified normal form and then
 * comparing that to the original string.  To speed that up, we also apply the
 * "quick check" algorithm specified in UAX #15, which can give a yes or no
 * answer for many strings by just scanning the string once.
 *
 * This function should generally be optimized for the case where the string
 * is in fact normalized.  In that case, we'll end up looking at the entire
 * string, so it's probably not worth doing any incremental conversion etc.
 */


/*
 * Check if first n chars are hexadecimal digits
 */




/*
 * Translate string with hexadecimal digits to number
 */


/*
 * Replaces Unicode escape sequences by Unicode characters
 */

