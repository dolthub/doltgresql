/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - typeStringToTypeName
 * - pts_error_callback
 * - typenameTypeIdAndMod
 * - typenameType
 * - LookupTypeName
 * - LookupTypeNameExtended
 * - TypeNameToString
 * - appendTypeNameToBuffer
 * - typeTypeId
 * - LookupTypeName
 * - typeTypeId
 * - TypeNameToString
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * parse_type.c
 *		handle type operations for parser
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 *
 * IDENTIFICATION
 *	  src/backend/parser/parse_type.c
 *
 *-------------------------------------------------------------------------
 */
#include "postgres.h"

#include "access/htup_details.h"
#include "catalog/namespace.h"
#include "catalog/pg_type.h"
#include "lib/stringinfo.h"
#include "nodes/makefuncs.h"
#include "parser/parse_type.h"
#include "parser/parser.h"
#include "utils/array.h"
#include "utils/builtins.h"
#include "utils/lsyscache.h"
#include "utils/syscache.h"

static int32 typenameTypeMod(ParseState *pstate, const TypeName *typeName,
							 Type typ);


/*
 * LookupTypeName
 *		Wrapper for typical case.
 */
Type
LookupTypeName(ParseState *pstate, const TypeName *typeName,
			   int32 *typmod_p, bool missing_ok)
{
	return LookupTypeNameExtended(pstate,
								  typeName, typmod_p, true, missing_ok);
}

/*
 * LookupTypeNameExtended
 *		Given a TypeName object, lookup the pg_type syscache entry of the type.
 *		Returns NULL if no such type can be found.  If the type is found,
 *		the typmod value represented in the TypeName struct is computed and
 *		stored into *typmod_p.
 *
 * NB: on success, the caller must ReleaseSysCache the type tuple when done
 * with it.
 *
 * NB: direct callers of this function MUST check typisdefined before assuming
 * that the type is fully valid.  Most code should go through typenameType
 * or typenameTypeId instead.
 *
 * typmod_p can be passed as NULL if the caller does not care to know the
 * typmod value, but the typmod decoration (if any) will be validated anyway,
 * except in the case where the type is not found.  Note that if the type is
 * found but is a shell, and there is typmod decoration, an error will be
 * thrown --- this is intentional.
 *
 * If temp_ok is false, ignore types in the temporary namespace.  Pass false
 * when the caller will decide, using goodness of fit criteria, whether the
 * typeName is actually a type or something else.  If typeName always denotes
 * a type (or denotes nothing), pass true.
 *
 * pstate is only used for error location info, and may be NULL.
 */
Type
LookupTypeNameExtended(ParseState *pstate,
					   const TypeName *typeName, int32 *typmod_p,
					   bool temp_ok, bool missing_ok)
{
	Oid			typoid;
	HeapTuple	tup;
	int32		typmod;

	if (typeName->names == NIL)
	{
		/* We have the OID already if it's an internally generated TypeName */
		typoid = typeName->typeOid;
	}
	else if (typeName->pct_type)
	{
        // CHANGED: Not currently implemented, requires us to get type mappings from caller
        elog(ERROR, "Not implemented");
		/* Handle %TYPE reference to type of an existing field */
		//RangeVar   *rel = makeRangeVar(NULL, NULL, typeName->location);
		//char	   *field = NULL;
		//Oid			relid;
		//AttrNumber	attnum;

		/* deconstruct the name list */
		//switch (list_length(typeName->names))
		//{
		//	case 1:
		//		ereport(ERROR,
		//				(errcode(ERRCODE_SYNTAX_ERROR),
		//				 errmsg("improper %%TYPE reference (too few dotted names): %s",
		//						NameListToString(typeName->names)),
		//				 parser_errposition(pstate, typeName->location)));
		//		break;
		//	case 2:
		//		rel->relname = strVal(linitial(typeName->names));
		//		field = strVal(lsecond(typeName->names));
		//		break;
		//	case 3:
		//		rel->schemaname = strVal(linitial(typeName->names));
		//		rel->relname = strVal(lsecond(typeName->names));
		//		field = strVal(lthird(typeName->names));
		//		break;
		//	case 4:
		//		rel->catalogname = strVal(linitial(typeName->names));
		//		rel->schemaname = strVal(lsecond(typeName->names));
		//		rel->relname = strVal(lthird(typeName->names));
		//		field = strVal(lfourth(typeName->names));
		//		break;
		//	default:
		//		ereport(ERROR,
		//				(errcode(ERRCODE_SYNTAX_ERROR),
		//				 errmsg("improper %%TYPE reference (too many dotted names): %s",
		//						NameListToString(typeName->names)),
		//				 parser_errposition(pstate, typeName->location)));
		//		break;
		//}

		/*
		 * Look up the field.
		 *
		 * XXX: As no lock is taken here, this might fail in the presence of
		 * concurrent DDL.  But taking a lock would carry a performance
		 * penalty and would also require a permissions check.
		 */
		//relid = RangeVarGetRelid(rel, NoLock, missing_ok);
		//attnum = get_attnum(relid, field);
		//if (attnum == InvalidAttrNumber)
		//{
		//	if (missing_ok)
		//		typoid = InvalidOid;
		//	else
		//		ereport(ERROR,
		//				(errcode(ERRCODE_UNDEFINED_COLUMN),
		//				 errmsg("column \"%s\" of relation \"%s\" does not exist",
		//						field, rel->relname),
		//				 parser_errposition(pstate, typeName->location)));
		//}
		//else
		//{
		//	typoid = get_atttype(relid, attnum);
        //
		//	/* this construct should never have an array indicator */
		//	Assert(typeName->arrayBounds == NIL);
        //
		//	/* emit nuisance notice (intentionally not errposition'd) */
		//	ereport(NOTICE,
		//			(errmsg("type reference %s converted to %s",
		//					TypeNameToString(typeName),
		//					format_type_be(typoid))));
		//}
	}
	else
	{
		/* Normal reference to a type name */
		char	   *schemaname;
		char	   *typname;

		/* deconstruct the name list */
		DeconstructQualifiedName(typeName->names, &schemaname, &typname);

		if (schemaname)
		{
			/* Look in specific schema only */
			Oid			namespaceId;
			ParseCallbackState pcbstate;

			setup_parser_errposition_callback(&pcbstate, pstate, typeName->location);

			namespaceId = LookupExplicitNamespace(schemaname, missing_ok);
			if (OidIsValid(namespaceId))
				typoid = GetSysCacheOid2(TYPENAMENSP, Anum_pg_type_oid,
										 PointerGetDatum(typname),
										 ObjectIdGetDatum(namespaceId));
			else
				typoid = InvalidOid;

			cancel_parser_errposition_callback(&pcbstate);
		}
		else
		{
			/* Unqualified type name, so search the search path */
			typoid = TypenameGetTypidExtended(typname, temp_ok);
		}

		/* If an array reference, return the array type instead */
		if (typeName->arrayBounds != NIL)
			typoid = get_array_type(typoid);
	}

	if (!OidIsValid(typoid))
	{
		if (typmod_p)
			*typmod_p = -1;
		return NULL;
	}

	tup = SearchSysCache1(TYPEOID, ObjectIdGetDatum(typoid));
	if (!HeapTupleIsValid(tup)) /* should not happen */
		elog(ERROR, "cache lookup failed for type %u", typoid);

	typmod = typenameTypeMod(pstate, typeName, (Type) tup);

	if (typmod_p)
		*typmod_p = typmod;

	return (Type) tup;
}


/*
 * LookupTypeNameOid
 *		Given a TypeName object, lookup the pg_type syscache entry of the type.
 *		Returns InvalidOid if no such type can be found.  If the type is found,
 *		return its Oid.
 *
 * NB: direct callers of this function need to be aware that the type OID
 * returned may correspond to a shell type.  Most code should go through
 * typenameTypeId instead.
 *
 * pstate is only used for error location info, and may be NULL.
 */


/*
 * typenameType - given a TypeName, return a Type structure and typmod
 *
 * This is equivalent to LookupTypeName, except that this will report
 * a suitable error message if the type cannot be found or is not defined.
 * Callers of this can therefore assume the result is a fully valid type.
 */
Type
typenameType(ParseState *pstate, const TypeName *typeName, int32 *typmod_p)
{
	Type		tup;

	tup = LookupTypeName(pstate, typeName, typmod_p, false);
	if (tup == NULL)
		ereport(ERROR,
				(errcode(ERRCODE_UNDEFINED_OBJECT),
				 errmsg("type \"%s\" does not exist",
						TypeNameToString(typeName)),
				 parser_errposition(pstate, typeName->location)));
	if (!((Form_pg_type) GETSTRUCT(tup))->typisdefined)
		ereport(ERROR,
				(errcode(ERRCODE_UNDEFINED_OBJECT),
				 errmsg("type \"%s\" is only a shell",
						TypeNameToString(typeName)),
				 parser_errposition(pstate, typeName->location)));
	return tup;
}

/*
 * typenameTypeId - given a TypeName, return the type's OID
 *
 * This is similar to typenameType, but we only hand back the type OID
 * not the syscache entry.
 */


/*
 * typenameTypeIdAndMod - given a TypeName, return the type's OID and typmod
 *
 * This is equivalent to typenameType, but we only hand back the type OID
 * and typmod, not the syscache entry.
 */
void
typenameTypeIdAndMod(ParseState *pstate, const TypeName *typeName,
					 Oid *typeid_p, int32 *typmod_p)
{
	Type		tup;

	tup = typenameType(pstate, typeName, typmod_p);
	*typeid_p = ((Form_pg_type) GETSTRUCT(tup))->oid;
	ReleaseSysCache(tup);
}

/*
 * typenameTypeMod - given a TypeName, return the internal typmod value
 *
 * This will throw an error if the TypeName includes type modifiers that are
 * illegal for the data type.
 *
 * The actual type OID represented by the TypeName must already have been
 * looked up, and is passed as "typ".
 *
 * pstate is only used for error location info, and may be NULL.
 */

static int32
typenameTypeMod(ParseState *pstate, const TypeName *typeName, Type typ)
{
return -1;}

/*
 * appendTypeNameToBuffer
 *		Append a string representing the name of a TypeName to a StringInfo.
 *		This is the shared guts of TypeNameToString and TypeNameListToString.
 *
 * NB: this must work on TypeNames that do not describe any actual type;
 * it is mostly used for reporting lookup errors.
 */
static void
appendTypeNameToBuffer(const TypeName *typeName, StringInfo string)
{
	if (typeName->names != NIL)
	{
		/* Emit possibly-qualified name as-is */
		ListCell   *l;

		foreach(l, typeName->names)
		{
			if (l != list_head(typeName->names))
				appendStringInfoChar(string, '.');
			appendStringInfoString(string, strVal(lfirst(l)));
		}
	}
	else
	{
		/* Look up internally-specified type */
		appendStringInfoString(string, format_type_be(typeName->typeOid));
	}

	/*
	 * Add decoration as needed, but only for fields considered by
	 * LookupTypeName
	 */
	if (typeName->pct_type)
		appendStringInfoString(string, "%TYPE");

	if (typeName->arrayBounds != NIL)
		appendStringInfoString(string, "[]");
}

/*
 * TypeNameToString
 *		Produce a string representing the name of a TypeName.
 *
 * NB: this must work on TypeNames that do not describe any actual type;
 * it is mostly used for reporting lookup errors.
 */
char *
TypeNameToString(const TypeName *typeName)
{
	StringInfoData string;

	initStringInfo(&string);
	appendTypeNameToBuffer(typeName, &string);
	return string.data;
}

/*
 * TypeNameListToString
 *		Produce a string representing the name(s) of a List of TypeNames
 */


/*
 * LookupCollation
 *
 * Look up collation by name, return OID, with support for error location.
 */


/*
 * GetColumnDefCollation
 *
 * Get the collation to be used for a column being defined, given the
 * ColumnDef node and the previously-determined column type OID.
 *
 * pstate is only used for error location purposes, and can be NULL.
 */


/* return a Type structure, given a type id */
/* NB: caller must ReleaseSysCache the type tuple when done with it */


/* given type (as type struct), return the type OID */
Oid
typeTypeId(Type tp)
{
	if (tp == NULL)				/* probably useless */
		elog(ERROR, "typeTypeId() called with NULL type struct");
	return ((Form_pg_type) GETSTRUCT(tp))->oid;
}

/* given type (as type struct), return the length of type */


/* given type (as type struct), return its 'byval' attribute */


/* given type (as type struct), return the type's name */


/* given type (as type struct), return its 'typrelid' attribute */


/* given type (as type struct), return its 'typcollation' attribute */


/*
 * Given a type structure and a string, returns the internal representation
 * of that string.  The "string" can be NULL to perform conversion of a NULL
 * (which might result in failure, if the input function rejects NULLs).
 */


/*
 * Given a typeid, return the type's typrelid (associated relation), if any.
 * Returns InvalidOid if type is not a composite type.
 */


/*
 * Given a typeid, return the type's typrelid (associated relation), if any.
 * Returns InvalidOid if type is not a composite type or a domain over one.
 * This is the same as typeidTypeRelid(getBaseType(type_id)), but faster.
 */


/*
 * error context callback for parse failure during parseTypeString()
 */
static void
pts_error_callback(void *arg)
{
	const char *str = (const char *) arg;

	errcontext("invalid type name \"%s\"", str);
}

/*
 * Given a string that is supposed to be a SQL-compatible type declaration,
 * such as "int4" or "integer" or "character varying(32)", parse
 * the string and return the result as a TypeName.
 *
 * If the string cannot be parsed as a type, an error is raised,
 * unless escontext is an ErrorSaveContext node, in which case we may
 * fill that and return NULL.  But note that the ErrorSaveContext option
 * is mostly aspirational at present: errors detected by the main
 * grammar, rather than here, will still be thrown.
 */
TypeName *
typeStringToTypeName(const char *str, Node *escontext)
{
	List	   *raw_parsetree_list;
	TypeName   *typeName;
	ErrorContextCallback ptserrcontext;

	/* make sure we give useful error for empty input */
	if (strspn(str, " \t\n\r\f\v") == strlen(str))
		goto fail;

	/*
	 * Setup error traceback support in case of ereport() during parse
	 */
	ptserrcontext.callback = pts_error_callback;
	ptserrcontext.arg = unconstify(char *, str);
	ptserrcontext.previous = error_context_stack;
	error_context_stack = &ptserrcontext;

	raw_parsetree_list = raw_parser(str, RAW_PARSE_TYPE_NAME);

	error_context_stack = ptserrcontext.previous;

	/* We should get back exactly one TypeName node. */
	Assert(list_length(raw_parsetree_list) == 1);
	typeName = linitial_node(TypeName, raw_parsetree_list);

	/* The grammar allows SETOF in TypeName, but we don't want that here. */
	if (typeName->setof)
		goto fail;

	return typeName;

fail:
	ereturn(escontext, NULL,
			(errcode(ERRCODE_SYNTAX_ERROR),
			 errmsg("invalid type name \"%s\"", str)));
}

/*
 * Given a string that is supposed to be a SQL-compatible type declaration,
 * such as "int4" or "integer" or "character varying(32)", parse
 * the string and convert it to a type OID and type modifier.
 *
 * If escontext is an ErrorSaveContext node, then errors are reported by
 * filling escontext and returning false, instead of throwing them.
 */

