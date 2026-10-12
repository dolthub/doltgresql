/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - format_procedure
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * regproc.c
 *	  Functions for the built-in types regproc, regclass, regtype, etc.
 *
 * These types are all binary-compatible with type Oid, and rely on Oid
 * for comparison and so forth.  Their only interesting behavior is in
 * special I/O conversion routines.
 *
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 *
 * IDENTIFICATION
 *	  src/backend/utils/adt/regproc.c
 *
 *-------------------------------------------------------------------------
 */
#include "postgres.h"

#include <ctype.h>

#include "access/htup_details.h"
#include "catalog/namespace.h"
#include "catalog/pg_class.h"
#include "catalog/pg_collation.h"
#include "catalog/pg_operator.h"
#include "catalog/pg_proc.h"
#include "catalog/pg_ts_config.h"
#include "catalog/pg_ts_dict.h"
#include "catalog/pg_type.h"
#include "lib/stringinfo.h"
#include "mb/pg_wchar.h"
#include "miscadmin.h"
#include "nodes/miscnodes.h"
#include "parser/parse_type.h"
#include "parser/scansup.h"
#include "utils/acl.h"
#include "utils/builtins.h"
#include "utils/lsyscache.h"
#include "utils/regproc.h"
#include "utils/syscache.h"
#include "utils/varlena.h"

static bool parseNumericOid(char *string, Oid *result, Node *escontext);
static bool parseDashOrOid(char *string, Oid *result, Node *escontext);
static bool parseNameAndArgTypes(const char *string, bool allowNone,
								 List **names, int *nargs, Oid *argtypes,
								 Node *escontext);


/*****************************************************************************
 *	 USER I/O ROUTINES														 *
 *****************************************************************************/

/*
 * regprocin		- converts "proname" to proc OID
 *
 * We also accept a numeric OID, for symmetry with the output routine.
 *
 * '-' signifies unknown (OID 0).  In all other cases, the input must
 * match an existing pg_proc entry.
 */


/*
 * to_regproc	- converts "proname" to proc OID
 *
 * If the name is not found, we return NULL.
 */


/*
 * regprocout		- converts proc OID to "pro_name"
 */


/*
 *		regprocrecv			- converts external binary format to regproc
 */


/*
 *		regprocsend			- converts regproc to binary format
 */



/*
 * regprocedurein		- converts "proname(args)" to proc OID
 *
 * We also accept a numeric OID, for symmetry with the output routine.
 *
 * '-' signifies unknown (OID 0).  In all other cases, the input must
 * match an existing pg_proc entry.
 */


/*
 * to_regprocedure	- converts "proname(args)" to proc OID
 *
 * If the name is not found, we return NULL.
 */


/*
 * format_procedure		- converts proc OID to "pro_name(args)"
 *
 * This exports the useful functionality of regprocedureout for use
 * in other backend modules.  The result is a palloc'd string.
 */

char *
format_procedure(Oid procedure_oid)
{
return pstrdup("plpgsql_function");}



/*
 * format_procedure_extended - converts procedure OID to "pro_name(args)"
 *
 * This exports the useful functionality of regprocedureout for use
 * in other backend modules.  The result is a palloc'd string, or NULL.
 *
 * Routine to produce regprocedure names; see format_procedure above.
 *
 * The following bits in 'flags' modify the behavior:
 * - FORMAT_PROC_INVALID_AS_NULL
 *			if the procedure OID is invalid or unknown, return NULL instead
 *			of the numeric OID.
 * - FORMAT_PROC_FORCE_QUALIFY
 *			always schema-qualify procedure names, regardless of search_path
 */


/*
 * Output an objname/objargs representation for the procedure with the
 * given OID.  If it doesn't exist, an error is thrown.
 *
 * This can be used to feed get_object_address.
 */


/*
 * regprocedureout		- converts proc OID to "pro_name(args)"
 */


/*
 *		regprocedurerecv			- converts external binary format to regprocedure
 */


/*
 *		regproceduresend			- converts regprocedure to binary format
 */



/*
 * regoperin		- converts "oprname" to operator OID
 *
 * We also accept a numeric OID, for symmetry with the output routine.
 *
 * '0' signifies unknown (OID 0).  In all other cases, the input must
 * match an existing pg_operator entry.
 */


/*
 * to_regoper		- converts "oprname" to operator OID
 *
 * If the name is not found, we return NULL.
 */


/*
 * regoperout		- converts operator OID to "opr_name"
 */


/*
 *		regoperrecv			- converts external binary format to regoper
 */


/*
 *		regopersend			- converts regoper to binary format
 */



/*
 * regoperatorin		- converts "oprname(args)" to operator OID
 *
 * We also accept a numeric OID, for symmetry with the output routine.
 *
 * '0' signifies unknown (OID 0).  In all other cases, the input must
 * match an existing pg_operator entry.
 */


/*
 * to_regoperator	- converts "oprname(args)" to operator OID
 *
 * If the name is not found, we return NULL.
 */


/*
 * format_operator_extended - converts operator OID to "opr_name(args)"
 *
 * This exports the useful functionality of regoperatorout for use
 * in other backend modules.  The result is a palloc'd string, or NULL.
 *
 * The following bits in 'flags' modify the behavior:
 * - FORMAT_OPERATOR_INVALID_AS_NULL
 *			if the operator OID is invalid or unknown, return NULL instead
 *			of the numeric OID.
 * - FORMAT_OPERATOR_FORCE_QUALIFY
 *			always schema-qualify operator names, regardless of search_path
 */








/*
 * regoperatorout		- converts operator OID to "opr_name(args)"
 */


/*
 *		regoperatorrecv			- converts external binary format to regoperator
 */


/*
 *		regoperatorsend			- converts regoperator to binary format
 */



/*
 * regclassin		- converts "classname" to class OID
 *
 * We also accept a numeric OID, for symmetry with the output routine.
 *
 * '-' signifies unknown (OID 0).  In all other cases, the input must
 * match an existing pg_class entry.
 */


/*
 * to_regclass		- converts "classname" to class OID
 *
 * If the name is not found, we return NULL.
 */


/*
 * regclassout		- converts class OID to "class_name"
 */


/*
 *		regclassrecv			- converts external binary format to regclass
 */


/*
 *		regclasssend			- converts regclass to binary format
 */



/*
 * regcollationin		- converts "collationname" to collation OID
 *
 * We also accept a numeric OID, for symmetry with the output routine.
 *
 * '-' signifies unknown (OID 0).  In all other cases, the input must
 * match an existing pg_collation entry.
 */


/*
 * to_regcollation		- converts "collationname" to collation OID
 *
 * If the name is not found, we return NULL.
 */


/*
 * regcollationout		- converts collation OID to "collation_name"
 */


/*
 *		regcollationrecv			- converts external binary format to regcollation
 */


/*
 *		regcollationsend			- converts regcollation to binary format
 */



/*
 * regtypein		- converts "typename" to type OID
 *
 * The type name can be specified using the full type syntax recognized by
 * the parser; for example, DOUBLE PRECISION and INTEGER[] will work and be
 * translated to the correct type names.  (We ignore any typmod info
 * generated by the parser, however.)
 *
 * We also accept a numeric OID, for symmetry with the output routine,
 * and for possible use in bootstrap mode.
 *
 * '-' signifies unknown (OID 0).  In all other cases, the input must
 * match an existing pg_type entry.
 */


/*
 * to_regtype		- converts "typename" to type OID
 *
 * If the name is not found, we return NULL.
 */


/*
 * to_regtypemod	- converts "typename" to type modifier
 *
 * If the name is not found, we return NULL.
 */


/*
 * regtypeout		- converts type OID to "typ_name"
 */


/*
 *		regtyperecv			- converts external binary format to regtype
 */


/*
 *		regtypesend			- converts regtype to binary format
 */



/*
 * regconfigin		- converts "tsconfigname" to tsconfig OID
 *
 * We also accept a numeric OID, for symmetry with the output routine.
 *
 * '-' signifies unknown (OID 0).  In all other cases, the input must
 * match an existing pg_ts_config entry.
 */


/*
 * regconfigout		- converts tsconfig OID to "tsconfigname"
 */


/*
 *		regconfigrecv			- converts external binary format to regconfig
 */


/*
 *		regconfigsend			- converts regconfig to binary format
 */



/*
 * regdictionaryin		- converts "tsdictionaryname" to tsdictionary OID
 *
 * We also accept a numeric OID, for symmetry with the output routine.
 *
 * '-' signifies unknown (OID 0).  In all other cases, the input must
 * match an existing pg_ts_dict entry.
 */


/*
 * regdictionaryout		- converts tsdictionary OID to "tsdictionaryname"
 */


/*
 *		regdictionaryrecv	- converts external binary format to regdictionary
 */


/*
 *		regdictionarysend	- converts regdictionary to binary format
 */


/*
 * regrolein	- converts "rolename" to role OID
 *
 * We also accept a numeric OID, for symmetry with the output routine.
 *
 * '-' signifies unknown (OID 0).  In all other cases, the input must
 * match an existing pg_authid entry.
 */


/*
 * to_regrole		- converts "rolename" to role OID
 *
 * If the name is not found, we return NULL.
 */


/*
 * regroleout		- converts role OID to "role_name"
 */


/*
 *		regrolerecv - converts external binary format to regrole
 */


/*
 *		regrolesend - converts regrole to binary format
 */


/*
 * regnamespacein		- converts "nspname" to namespace OID
 *
 * We also accept a numeric OID, for symmetry with the output routine.
 *
 * '-' signifies unknown (OID 0).  In all other cases, the input must
 * match an existing pg_namespace entry.
 */


/*
 * to_regnamespace		- converts "nspname" to namespace OID
 *
 * If the name is not found, we return NULL.
 */


/*
 * regnamespaceout		- converts namespace OID to "nsp_name"
 */


/*
 *		regnamespacerecv	- converts external binary format to regnamespace
 */


/*
 *		regnamespacesend		- converts regnamespace to binary format
 */


/*
 * text_regclass: convert text to regclass
 *
 * This could be replaced by CoerceViaIO, except that we need to treat
 * text-to-regclass as an implicit cast to support legacy forms of nextval()
 * and related functions.
 */



/*
 * Given a C string, parse it into a qualified-name list.
 *
 * If escontext is an ErrorSaveContext node, invalid input will be
 * reported there instead of being thrown, and we return NIL.
 * (NIL is not possible as a success return, since empty-input is an error.)
 */


/*****************************************************************************
 *	 SUPPORT ROUTINES														 *
 *****************************************************************************/

/*
 * Given a C string, see if it is all-digits (and not empty).
 * If so, convert directly to OID and return true.
 * If it is not all-digits, return false.
 *
 * If escontext is an ErrorSaveContext node, any error in oidin() will be
 * reported there instead of being thrown (but we still return true).
 */


/*
 * As above, but also accept "-" as meaning 0 (InvalidOid).
 */


/*
 * Given a C string, parse it into a qualified function or operator name
 * followed by a parenthesized list of type names.  Reduce the
 * type names to an array of OIDs (returned into *nargs and *argtypes;
 * the argtypes array should be of size FUNC_MAX_ARGS).  The function or
 * operator name is returned to *names as a List of Strings.
 *
 * If allowNone is true, accept "NONE" and return it as InvalidOid (this is
 * for unary operators).
 *
 * Returns true on success, false on failure (the latter only possible
 * if escontext is an ErrorSaveContext node).
 */

