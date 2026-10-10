/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - interpret_function_parameter_list
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * functioncmds.c
 *
 *	  Routines for CREATE and DROP FUNCTION commands and CREATE and DROP
 *	  CAST commands.
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 *
 * IDENTIFICATION
 *	  src/backend/commands/functioncmds.c
 *
 * DESCRIPTION
 *	  These routines take the parse tree and pick out the
 *	  appropriate arguments/flags, and pass the results to the
 *	  corresponding "FooCreate" routines (in src/backend/catalog) that do
 *	  the actual catalog-munging.  These routines also verify permission
 *	  of the user to execute the command.
 *
 * NOTES
 *	  These things must be defined and committed in the following order:
 *		"create function":
 *				input/output, recv/send procedures
 *		"create type":
 *				type
 *		"create operator":
 *				operators
 *
 *-------------------------------------------------------------------------
 */
#include "postgres.h"

#include "access/htup_details.h"
#include "access/table.h"
#include "catalog/catalog.h"
#include "catalog/dependency.h"
#include "catalog/indexing.h"
#include "catalog/objectaccess.h"
#include "catalog/pg_aggregate.h"
#include "catalog/pg_cast.h"
#include "catalog/pg_language.h"
#include "catalog/pg_namespace.h"
#include "catalog/pg_proc.h"
#include "catalog/pg_transform.h"
#include "catalog/pg_type.h"
#include "commands/defrem.h"
#include "commands/extension.h"
#include "commands/proclang.h"
#include "executor/executor.h"
#include "executor/functions.h"
#include "funcapi.h"
#include "miscadmin.h"
#include "nodes/nodeFuncs.h"
#include "optimizer/optimizer.h"
#include "parser/analyze.h"
#include "parser/parse_coerce.h"
#include "parser/parse_collate.h"
#include "parser/parse_expr.h"
#include "parser/parse_func.h"
#include "parser/parse_type.h"
#include "pgstat.h"
#include "tcop/pquery.h"
#include "tcop/utility.h"
#include "utils/acl.h"
#include "utils/builtins.h"
#include "utils/guc.h"
#include "utils/lsyscache.h"
#include "utils/rel.h"
#include "utils/snapmgr.h"
#include "utils/syscache.h"
#include "utils/typcache.h"

/*
 *	 Examine the RETURNS clause of the CREATE FUNCTION statement
 *	 and return information about it as *prorettype_p and *returnsSet_p.
 *
 * This is more complex than the average typename lookup because we want to
 * allow a shell type to be used, or even created if the specified return type
 * doesn't exist yet.  (Without this, there's no way to define the I/O procs
 * for a new type.)  But SQL function creation won't cope, so error out if
 * the target language is SQL.  (We do this here, not in the SQL-function
 * validator, so as not to produce a NOTICE and then an ERROR for the same
 * condition.)
 */


/*
 * Interpret the function parameter list of a CREATE FUNCTION,
 * CREATE PROCEDURE, or CREATE AGGREGATE statement.
 *
 * Input parameters:
 * parameters: list of FunctionParameter structs
 * languageOid: OID of function language (InvalidOid if it's CREATE AGGREGATE)
 * objtype: identifies type of object being created
 *
 * Results are stored into output parameters.  parameterTypes must always
 * be created, but the other arrays/lists can be NULL pointers if not needed.
 * variadicArgType is set to the variadic array type if there's a VARIADIC
 * parameter (there can be only one); or to InvalidOid if not.
 * requiredResultType is set to InvalidOid if there are no OUT parameters,
 * else it is set to the OID of the implied result type.
 */
void
interpret_function_parameter_list(ParseState *pstate,
								  List *parameters,
								  Oid languageOid,
								  ObjectType objtype,
								  oidvector **parameterTypes,
								  List **parameterTypes_list,
								  ArrayType **allParameterTypes,
								  ArrayType **parameterModes,
								  ArrayType **parameterNames,
								  List **inParameterNames_list,
								  List **parameterDefaults,
								  Oid *variadicArgType,
								  Oid *requiredResultType)
{
	int			parameterCount = list_length(parameters);
	Oid		   *inTypes;
	int			inCount = 0;
	Datum	   *allTypes;
	Datum	   *paramModes;
	Datum	   *paramNames;
	int			outCount = 0;
	int			varCount = 0;
	bool		have_names = false;
	bool		have_defaults = false;
	ListCell   *x;
	int			i;

	*variadicArgType = InvalidOid;	/* default result */
	*requiredResultType = InvalidOid;	/* default result */

	inTypes = (Oid *) palloc(parameterCount * sizeof(Oid));
	allTypes = (Datum *) palloc(parameterCount * sizeof(Datum));
	paramModes = (Datum *) palloc(parameterCount * sizeof(Datum));
	paramNames = (Datum *) palloc0(parameterCount * sizeof(Datum));
	*parameterDefaults = NIL;

	/* Scan the list and extract data into work arrays */
	i = 0;
	foreach(x, parameters)
	{
		FunctionParameter *fp = (FunctionParameter *) lfirst(x);
		TypeName   *t = fp->argType;
		FunctionParameterMode fpmode = fp->mode;
		bool		isinput = false;
		Oid			toid;
		Type		typtup;

		/* For our purposes here, a defaulted mode spec is identical to IN */
		if (fpmode == FUNC_PARAM_DEFAULT)
			fpmode = FUNC_PARAM_IN;

		typtup = LookupTypeName(pstate, t, NULL, false);
		if (typtup)
		{
			if (!((Form_pg_type) GETSTRUCT(typtup))->typisdefined)
			{
				/* As above, hard error if language is SQL */
				if (languageOid == SQLlanguageId)
					ereport(ERROR,
							(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
							 errmsg("SQL function cannot accept shell type %s",
									TypeNameToString(t)),
							 parser_errposition(pstate, t->location)));
				/* We don't allow creating aggregates on shell types either */
				else if (objtype == OBJECT_AGGREGATE)
					ereport(ERROR,
							(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
							 errmsg("aggregate cannot accept shell type %s",
									TypeNameToString(t)),
							 parser_errposition(pstate, t->location)));
				else
					ereport(NOTICE,
							(errcode(ERRCODE_WRONG_OBJECT_TYPE),
							 errmsg("argument type %s is only a shell",
									TypeNameToString(t)),
							 parser_errposition(pstate, t->location)));
			}
			toid = typeTypeId(typtup);
			ReleaseSysCache(typtup);
		}
		else
		{
			ereport(ERROR,
					(errcode(ERRCODE_UNDEFINED_OBJECT),
					 errmsg("type %s does not exist",
							TypeNameToString(t)),
					 parser_errposition(pstate, t->location)));
			toid = InvalidOid;	/* keep compiler quiet */
		}

		if (t->setof)
		{
			if (objtype == OBJECT_AGGREGATE)
				ereport(ERROR,
						(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
						 errmsg("aggregates cannot accept set arguments"),
						 parser_errposition(pstate, fp->location)));
			else if (objtype == OBJECT_PROCEDURE)
				ereport(ERROR,
						(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
						 errmsg("procedures cannot accept set arguments"),
						 parser_errposition(pstate, fp->location)));
			else
				ereport(ERROR,
						(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
						 errmsg("functions cannot accept set arguments"),
						 parser_errposition(pstate, fp->location)));
		}

		/* handle input parameters */
		if (fpmode != FUNC_PARAM_OUT && fpmode != FUNC_PARAM_TABLE)
		{
			/* other input parameters can't follow a VARIADIC parameter */
			if (varCount > 0)
				ereport(ERROR,
						(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
						 errmsg("VARIADIC parameter must be the last input parameter"),
						 parser_errposition(pstate, fp->location)));
			inTypes[inCount++] = toid;
			isinput = true;
			if (parameterTypes_list)
				*parameterTypes_list = lappend_oid(*parameterTypes_list, toid);
		}

		/* handle output parameters */
		if (fpmode != FUNC_PARAM_IN && fpmode != FUNC_PARAM_VARIADIC)
		{
			if (objtype == OBJECT_PROCEDURE)
			{
				/*
				 * We disallow OUT-after-VARIADIC only for procedures.  While
				 * such a case causes no confusion in ordinary function calls,
				 * it would cause confusion in a CALL statement.
				 */
				if (varCount > 0)
					ereport(ERROR,
							(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
							 errmsg("VARIADIC parameter must be the last parameter"),
							 parser_errposition(pstate, fp->location)));
				/* Procedures with output parameters always return RECORD */
				*requiredResultType = RECORDOID;
			}
			else if (outCount == 0) /* save first output param's type */
				*requiredResultType = toid;
			outCount++;
		}

		if (fpmode == FUNC_PARAM_VARIADIC)
		{
			*variadicArgType = toid;
			varCount++;
			/* validate variadic parameter type */
			switch (toid)
			{
				case ANYARRAYOID:
				case ANYCOMPATIBLEARRAYOID:
				case ANYOID:
					/* okay */
					break;
				default:
					if (!OidIsValid(get_element_type(toid)))
						ereport(ERROR,
								(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
								 errmsg("VARIADIC parameter must be an array"),
								 parser_errposition(pstate, fp->location)));
					break;
			}
		}

		allTypes[i] = ObjectIdGetDatum(toid);

		paramModes[i] = CharGetDatum(fpmode);

		if (fp->name && fp->name[0])
		{
			ListCell   *px;

			/*
			 * As of Postgres 9.0 we disallow using the same name for two
			 * input or two output function parameters.  Depending on the
			 * function's language, conflicting input and output names might
			 * be bad too, but we leave it to the PL to complain if so.
			 */
			foreach(px, parameters)
			{
				FunctionParameter *prevfp = (FunctionParameter *) lfirst(px);
				FunctionParameterMode prevfpmode;

				if (prevfp == fp)
					break;
				/* as above, default mode is IN */
				prevfpmode = prevfp->mode;
				if (prevfpmode == FUNC_PARAM_DEFAULT)
					prevfpmode = FUNC_PARAM_IN;
				/* pure in doesn't conflict with pure out */
				if ((fpmode == FUNC_PARAM_IN ||
					 fpmode == FUNC_PARAM_VARIADIC) &&
					(prevfpmode == FUNC_PARAM_OUT ||
					 prevfpmode == FUNC_PARAM_TABLE))
					continue;
				if ((prevfpmode == FUNC_PARAM_IN ||
					 prevfpmode == FUNC_PARAM_VARIADIC) &&
					(fpmode == FUNC_PARAM_OUT ||
					 fpmode == FUNC_PARAM_TABLE))
					continue;
				if (prevfp->name && prevfp->name[0] &&
					strcmp(prevfp->name, fp->name) == 0)
					ereport(ERROR,
							(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
							 errmsg("parameter name \"%s\" used more than once",
									fp->name),
							 parser_errposition(pstate, fp->location)));
			}

			paramNames[i] = CStringGetTextDatum(fp->name);
			have_names = true;
		}

		if (inParameterNames_list)
			*inParameterNames_list = lappend(*inParameterNames_list, makeString(fp->name ? fp->name : pstrdup("")));

		if (fp->defexpr)
		{
			Node	   *def;

			if (!isinput)
				ereport(ERROR,
						(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
						 errmsg("only input parameters can have default values"),
						 parser_errposition(pstate, fp->location)));

			def = transformExpr(pstate, fp->defexpr,
								EXPR_KIND_FUNCTION_DEFAULT);
			def = coerce_to_specific_type(pstate, def, toid, "DEFAULT");
			assign_expr_collations(pstate, def);

			/*
			 * Make sure no variables are referred to (this is probably dead
			 * code now that add_missing_from is history).
			 */
			if (pstate->p_rtable != NIL ||
				contain_var_clause(def))
				ereport(ERROR,
						(errcode(ERRCODE_INVALID_COLUMN_REFERENCE),
						 errmsg("cannot use table references in parameter default value"),
						 parser_errposition(pstate, fp->location)));

			/*
			 * transformExpr() should have already rejected subqueries,
			 * aggregates, and window functions, based on the EXPR_KIND_ for a
			 * default expression.
			 *
			 * It can't return a set either --- but coerce_to_specific_type
			 * already checked that for us.
			 *
			 * Note: the point of these restrictions is to ensure that an
			 * expression that, on its face, hasn't got subplans, aggregates,
			 * etc cannot suddenly have them after function default arguments
			 * are inserted.
			 */

			*parameterDefaults = lappend(*parameterDefaults, def);
			have_defaults = true;
		}
		else
		{
			if (isinput && have_defaults)
				ereport(ERROR,
						(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
						 errmsg("input parameters after one with a default value must also have defaults"),
						 parser_errposition(pstate, fp->location)));

			/*
			 * For procedures, we also can't allow OUT parameters after one
			 * with a default, because the same sort of confusion arises in a
			 * CALL statement.
			 */
			if (objtype == OBJECT_PROCEDURE && have_defaults)
				ereport(ERROR,
						(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
						 errmsg("procedure OUT parameters cannot appear after one with a default value"),
						 parser_errposition(pstate, fp->location)));
		}

		i++;
	}

	/* Now construct the proper outputs as needed */
	*parameterTypes = buildoidvector(inTypes, inCount);

	if (outCount > 0 || varCount > 0)
	{
		*allParameterTypes = construct_array_builtin(allTypes, parameterCount, OIDOID);
		*parameterModes = construct_array_builtin(paramModes, parameterCount, CHAROID);
		if (outCount > 1)
			*requiredResultType = RECORDOID;
		/* otherwise we set requiredResultType correctly above */
	}
	else
	{
		*allParameterTypes = NULL;
		*parameterModes = NULL;
	}

	if (have_names)
	{
		for (i = 0; i < parameterCount; i++)
		{
			if (paramNames[i] == PointerGetDatum(NULL))
				paramNames[i] = CStringGetTextDatum("");
		}
		*parameterNames = construct_array_builtin(paramNames, parameterCount, TEXTOID);
	}
	else
		*parameterNames = NULL;
}


/*
 * Recognize one of the options that can be passed to both CREATE
 * FUNCTION and ALTER FUNCTION and return it via one of the out
 * parameters. Returns true if the passed option was recognized. If
 * the out parameter we were going to assign to points to non-NULL,
 * raise a duplicate-clause error.  (We don't try to detect duplicate
 * SET parameters though --- if you're redundant, the last one wins.)
 */






/*
 * Update a proconfig value according to a list of VariableSetStmt items.
 *
 * The input and result may be NULL to signify a null entry.
 */





/*
 * Dissect the list of options assembled in gram.y into function
 * attributes.
 */



/*
 * For a dynamically linked C language object, the form of the clause is
 *
 *	   AS <object file name> [, <link symbol name> ]
 *
 * In all other cases
 *
 *	   AS <object reference, or sql code>
 */



/*
 * CreateFunction
 *	 Execute a CREATE FUNCTION (or CREATE PROCEDURE) utility statement.
 */


/*
 * Guts of function deletion.
 *
 * Note: this is also used for aggregate deletion, since the OIDs of
 * both functions and aggregates point to pg_proc.
 */


/*
 * Implements the ALTER FUNCTION utility command (except for the
 * RENAME and OWNER clauses, which are handled as part of the generic
 * ALTER framework).
 */



/*
 * CREATE CAST
 */
#ifdef NOT_USED
#endif





/*
 * CREATE TRANSFORM
 */



/*
 * get_transform_oid - given type OID and language OID, look up a transform OID
 *
 * If missing_ok is false, throw an error if the transform is not found.  If
 * true, just return InvalidOid.
 */



/*
 * Subroutine for ALTER FUNCTION/AGGREGATE SET SCHEMA/RENAME
 *
 * Is there a function with the given name and signature already in the given
 * namespace?  If so, raise an appropriate error message.
 */


/*
 * ExecuteDoStmt
 *		Execute inline procedural-language code
 *
 * See at ExecuteCallStmt() about the atomic argument.
 */


/*
 * Execute CALL statement
 *
 * Inside a top-level CALL statement, transaction-terminating commands such as
 * COMMIT or a PL-specific equivalent are allowed.  The terminology in the SQL
 * standard is that CALL establishes a non-atomic execution context.  Most
 * other commands establish an atomic execution context, in which transaction
 * control actions are not allowed.  If there are nested executions of CALL,
 * we want to track the execution context recursively, so that the nested
 * CALLs can also do transaction control.  Note, however, that for example in
 * CALL -> SELECT -> CALL, the second call cannot do transaction control,
 * because the SELECT in between establishes an atomic execution context.
 *
 * So when ExecuteCallStmt() is called from the top level, we pass in atomic =
 * false (recall that that means transactions = yes).  We then create a
 * CallContext node with content atomic = false, which is passed in the
 * fcinfo->context field to the procedure invocation.  The language
 * implementation should then take appropriate measures to allow or prevent
 * transaction commands based on that information, e.g., call
 * SPI_connect_ext(SPI_OPT_NONATOMIC).  The language should also pass on the
 * atomic flag to any nested invocations to CALL.
 *
 * The expression data structures and execution context that we create
 * within this function are children of the portalContext of the Portal
 * that the CALL utility statement runs in.  Therefore, any pass-by-ref
 * values that we're passing to the procedure will survive transaction
 * commits that might occur inside the procedure.
 */


/*
 * Construct the tuple descriptor for a CALL statement return
 */

