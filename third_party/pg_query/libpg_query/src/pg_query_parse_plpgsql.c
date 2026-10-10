#include <stdio.h>

#include "pg_query.h"
#include "pg_query_internal.h"
#include "pg_query_json_plpgsql.h"
#include "pg_query_proctup_attrs.h"

/* the text of the statement whose PL/pgSQL function is compiling (added for Doltgres) */
extern __thread const char *pg_query_plpgsql_source;

#include <assert.h>

#include <catalog/pg_type.h>
#include <catalog/objectaddress.h>
#include <catalog/pg_language_d.h>
#include <catalog/pg_proc.h>
#include <commands/defrem.h>
#include <nodes/parsenodes.h>
#include <nodes/nodeFuncs.h>
#include <parser/parse_type.h>
#include <utils/builtins.h>
#include <utils/lsyscache.h>
#include <utils/syscache.h>

typedef struct {
	PLpgSQL_function *func;
	PgQueryError* error;
} PgQueryInternalPlpgsqlFuncAndError;

static PgQueryInternalPlpgsqlFuncAndError pg_query_raw_parse_plpgsql(Node* stmt);

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
static void
compute_return_type(TypeName *returnType, Oid languageOid,
					Oid *prorettype_p, bool *returnsSet_p)
{
	Oid			rettype;
	Type		typtup;
	//AclResult	aclresult;

	typtup = LookupTypeName(NULL, returnType, NULL, false);

	if (typtup)
	{
		if (!((Form_pg_type) GETSTRUCT(typtup))->typisdefined)
		{
			if (languageOid == SQLlanguageId)
				ereport(ERROR,
						(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
						 errmsg("SQL function cannot return shell type %s",
								TypeNameToString(returnType))));
			else
				ereport(NOTICE,
						(errcode(ERRCODE_WRONG_OBJECT_TYPE),
						 errmsg("return type %s is only a shell",
								TypeNameToString(returnType))));
		}
		rettype = typeTypeId(typtup);
		ReleaseSysCache(typtup);
	}
	else
	{
		char	   *typnam = TypeNameToString(returnType);
		Oid			namespaceId;
		char	   *typname;
		ObjectAddress address;

		/*
		 * Only C-coded functions can be I/O functions.  We enforce this
		 * restriction here mainly to prevent littering the catalogs with
		 * shell types due to simple typos in user-defined function
		 * definitions.
		 */
		if (languageOid != INTERNALlanguageId &&
			languageOid != ClanguageId)
			ereport(ERROR,
					(errcode(ERRCODE_UNDEFINED_OBJECT),
					 errmsg("type \"%s\" does not exist", typnam)));

		/* Reject if there's typmod decoration, too */
		if (returnType->typmods != NIL)
			ereport(ERROR,
					(errcode(ERRCODE_SYNTAX_ERROR),
					 errmsg("type modifier cannot be specified for shell type \"%s\"",
							typnam)));

		elog(ERROR, "Not implemented");

		/* Otherwise, go ahead and make a shell type */
		/*ereport(NOTICE,
				(errcode(ERRCODE_UNDEFINED_OBJECT),
				 errmsg("type \"%s\" is not yet defined", typnam),
				 errdetail("Creating a shell type definition.")));
		namespaceId = QualifiedNameGetCreationNamespace(returnType->names,
														&typname);
		aclresult = object_aclcheck(NamespaceRelationId, namespaceId, GetUserId(),
									ACL_CREATE);
		if (aclresult != ACLCHECK_OK)
			aclcheck_error(aclresult, OBJECT_SCHEMA,
						   get_namespace_name(namespaceId));
		address = TypeShellMake(typname, namespaceId, GetUserId());
		rettype = address.objectId;
		Assert(OidIsValid(rettype));*/
	}

	//aclresult = object_aclcheck(TypeRelationId, rettype, GetUserId(), ACL_USAGE);
	//if (aclresult != ACLCHECK_OK)
	//	aclcheck_error_type(aclresult, rettype);

	*prorettype_p = rettype;
	*returnsSet_p = returnType->setof;
}

static PLpgSQL_function *compile_do_stmt(DoStmt* stmt)
{
	char *proc_source = NULL;
	const ListCell *lc;
	char *language = "plpgsql";

	assert(IsA(stmt, DoStmt));

	foreach(lc, stmt->args)
	{
		DefElem* elem = (DefElem*) lfirst(lc);

		if (strcmp(elem->defname, "as") == 0) {

			assert(IsA(elem->arg, String));
			proc_source = strVal(elem->arg);
		} else if (strcmp(elem->defname, "language") == 0) {
			language = strVal(elem->arg);
		}
	}

	if (proc_source == NULL)
		ereport(ERROR,
				(errcode(ERRCODE_SYNTAX_ERROR),
				 errmsg("no inline code specified")));

	if(strcmp(language, "plpgsql") != 0) {
		return (PLpgSQL_function *) palloc0(sizeof(PLpgSQL_function));
	}
	return plpgsql_compile_inline(proc_source);

}

/*
 * Allocate a wrapper holding a HeapTupleData and the Anum-indexed
 * values[]/nulls[] arrays, with the pg_proc header+Form_pg_proc payload
 * appended after it so t_data points into the same chunk. We embed the
 * arrays in the wrapper struct so SysCacheGetAttr can recover them via
 * container_of from the HeapTuple pointer without any thread-local state.
 */
static ProcTupWithAttrs *
build_fake_proc_tuple(Form_pg_proc procStruct)
{
	ProcTupWithAttrs *wrapper;
	HeapTupleHeader td;
	Size		len,
				data_len;
	int			hoff;

	len = offsetof(HeapTupleHeaderData, t_bits);
	hoff = len = MAXALIGN(len);
	data_len = MAXALIGN(sizeof(FormData_pg_proc));
	len += data_len;

	wrapper = (ProcTupWithAttrs *) palloc0(sizeof(ProcTupWithAttrs) + len);
	wrapper->tup.t_data = td = (HeapTupleHeader) ((char *) wrapper + sizeof(ProcTupWithAttrs));
	wrapper->tup.t_len = len;
	ItemPointerSetInvalid(&(wrapper->tup.t_self));
	wrapper->tup.t_tableOid = InvalidOid;

	HeapTupleHeaderSetDatumLength(td, len);
	ItemPointerSetInvalid(&(td->t_ctid));
	HeapTupleHeaderSetNatts(td, Natts_pg_proc);
	td->t_hoff = hoff;

	memcpy((char *) td + hoff, procStruct, sizeof(FormData_pg_proc));

	return wrapper;
}

/*
 * Stripped version of upstream ProcedureCreate. Builds the Anum_pg_proc_*
 * indexed values[]/nulls[] arrays (as upstream does on its way to
 * heap_form_tuple), populates the Form_pg_proc fields plpgsql_compile_callback
 * reads via GETSTRUCT, and returns a wrapper carrying both the HeapTuple and
 * the attribute arrays. Catalog insertion, syscache duplicate checks, ACL
 * handling, and dependency tracking are all dropped.
 */
static ProcTupWithAttrs *
pg_query_procedure_create(const char *procedureName,
						  bool returnsSet,
						  Oid returnType,
						  char prokind,
						  char volatility,
						  oidvector *parameterTypes,
						  Datum allParameterTypes,
						  Datum parameterModes,
						  Datum parameterNames,
						  const char *prosrc)
{
	ProcTupWithAttrs *wrapper;
	FormData_pg_proc procStruct = {0};
	Datum	   *values;
	bool	   *nulls;
	int			parameterCount;

	Assert(PointerIsValid(prosrc));

	parameterCount = parameterTypes->dim1;

	/* Fixed-size attributes that GETSTRUCT will read back via procStruct. */
	namestrcpy(&procStruct.proname, procedureName);
	procStruct.prokind = prokind;
	procStruct.provolatile = volatility;
	procStruct.pronargs = parameterCount;
	procStruct.prorettype = returnType;
	procStruct.proretset = returnsSet;

	wrapper = build_fake_proc_tuple(&procStruct);
	values = wrapper->values;
	nulls = wrapper->nulls;

	/*
	 * Anum-indexed attributes that get_func_arg_info reads via
	 * SysCacheGetAttr. These mirror the assignments in upstream
	 * ProcedureCreate.
	 */
	values[Anum_pg_proc_proname - 1] = NameGetDatum(&procStruct.proname);
	values[Anum_pg_proc_prokind - 1] = CharGetDatum(prokind);
	values[Anum_pg_proc_provolatile - 1] = CharGetDatum(volatility);
	values[Anum_pg_proc_pronargs - 1] = UInt16GetDatum(parameterCount);
	values[Anum_pg_proc_prorettype - 1] = ObjectIdGetDatum(returnType);
	values[Anum_pg_proc_proretset - 1] = BoolGetDatum(returnsSet);
	values[Anum_pg_proc_proargtypes - 1] = PointerGetDatum(parameterTypes);

	if (allParameterTypes != PointerGetDatum(NULL))
		values[Anum_pg_proc_proallargtypes - 1] = allParameterTypes;
	else
		nulls[Anum_pg_proc_proallargtypes - 1] = true;

	if (parameterModes != PointerGetDatum(NULL))
		values[Anum_pg_proc_proargmodes - 1] = parameterModes;
	else
		nulls[Anum_pg_proc_proargmodes - 1] = true;

	if (parameterNames != PointerGetDatum(NULL))
		values[Anum_pg_proc_proargnames - 1] = parameterNames;
	else
		nulls[Anum_pg_proc_proargnames - 1] = true;

	values[Anum_pg_proc_prosrc - 1] = CStringGetTextDatum(prosrc);

	return wrapper;
}

/*
 * Stripped version of upstream CreateFunction. Picks the language and source
 * out of stmt->options, runs interpret_function_parameter_list and
 * compute_return_type the same way upstream does, then calls
 * pg_query_procedure_create to obtain a pg_proc HeapTuple.
 *
 * is_dml_trigger / is_event_trigger are inferred from the RETURNS clause and
 * returned so the caller can attach a TriggerData / EventTriggerData node to
 * fcinfo->context, which is what plpgsql_compile_callback inspects via
 * CALLED_AS_TRIGGER / CALLED_AS_EVENT_TRIGGER.
 */
static ProcTupWithAttrs *
pg_query_create_function(CreateFunctionStmt *stmt,
						 const char *language,
						 const char *proc_source,
						 bool *is_dml_trigger,
						 bool *is_event_trigger)
{
	char	   *funcname;
	Oid			prorettype;
	bool		returnsSet;
	char		prokind;
	oidvector  *parameterTypes;
	List	   *parameterTypes_list = NIL;
	ArrayType  *allParameterTypes;
	ArrayType  *parameterModes;
	ArrayType  *parameterNames;
	List	   *inParameterNames_list = NIL;
	List	   *parameterDefaults;
	Oid			variadicArgType;
	Oid			requiredResultType;

	funcname = strVal(linitial(stmt->funcname));

	*is_dml_trigger = false;
	*is_event_trigger = false;
	if (stmt->returnType != NULL)
	{
		foreach_ptr(String, val, stmt->returnType->names)
		{
			if (strcmp(val->sval, "trigger") == 0)
				*is_dml_trigger = true;
			else if (strcmp(val->sval, "event_trigger") == 0)
				*is_event_trigger = true;
		}
	}

	interpret_function_parameter_list(palloc0(sizeof(ParseState)),
									  stmt->parameters,
									  InvalidOid,
									  stmt->is_procedure ? OBJECT_PROCEDURE : OBJECT_FUNCTION,
									  &parameterTypes,
									  &parameterTypes_list,
									  &allParameterTypes,
									  &parameterModes,
									  &parameterNames,
									  &inParameterNames_list,
									  &parameterDefaults,
									  &variadicArgType,
									  &requiredResultType);

	if (stmt->is_procedure)
	{
		Assert(!stmt->returnType);
		prorettype = requiredResultType ? requiredResultType : VOIDOID;
		returnsSet = false;
	}
	else if (stmt->returnType)
	{
		compute_return_type(stmt->returnType, InvalidOid,
							&prorettype, &returnsSet);
		if (OidIsValid(requiredResultType) && prorettype != requiredResultType)
			ereport(ERROR,
					(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
					 errmsg("function result type must be %s because of OUT parameters",
							format_type_be(requiredResultType))));
	}
	else if (OidIsValid(requiredResultType))
	{
		prorettype = requiredResultType;
		returnsSet = false;
	}
	else
	{
		ereport(ERROR,
				(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
				 errmsg("function result type must be specified")));
	}

	prokind = stmt->is_procedure ? PROKIND_PROCEDURE : PROKIND_FUNCTION;

	/*
	 * When there are no OUT args, interpret_function_parameter_list returns
	 * allParameterTypes=NULL and upstream get_func_arg_info would then read
	 * the inline procStruct->proargtypes oidvector. Our fake Form_pg_proc has
	 * an empty oidvector, so synthesise an allParameterTypes array from the
	 * IN args to keep get_func_arg_info on the array path.
	 */
	if (allParameterTypes == NULL && parameterTypes->dim1 > 0)
	{
		Datum	   *typeDatums = palloc(parameterTypes->dim1 * sizeof(Datum));
		int			i;

		for (i = 0; i < parameterTypes->dim1; i++)
			typeDatums[i] = ObjectIdGetDatum(parameterTypes->values[i]);
		allParameterTypes = construct_array_builtin(typeDatums,
													parameterTypes->dim1,
													OIDOID);
	}

	return pg_query_procedure_create(funcname,
									 returnsSet,
									 prorettype,
									 prokind,
									 PROVOLATILE_VOLATILE,
									 parameterTypes,
									 allParameterTypes ? PointerGetDatum(allParameterTypes) : PointerGetDatum(NULL),
									 parameterModes ? PointerGetDatum(parameterModes) : PointerGetDatum(NULL),
									 parameterNames ? PointerGetDatum(parameterNames) : PointerGetDatum(NULL),
									 proc_source);
}

/*
 * Alternate path: rather than duplicate plpgsql_compile_callback's body in
 * compile_create_function_stmt, build the inputs the callback expects via
 * pg_query_create_function and invoke the real callback. The pg_proc tuple's
 * variable-length attributes are exposed to the callback's SysCacheGetAttr
 * calls through the values[]/nulls[] arrays mirrored from upstream
 * ProcedureCreate.
 */
static PLpgSQL_function *
compile_create_function_stmt_via_callback(CreateFunctionStmt *stmt)
{
	PLpgSQL_function *function;
	ProcTupWithAttrs *wrapper;
	FmgrInfo	flinfo = {0};
	LOCAL_FCINFO(fcinfo, 0);
	TriggerData triggerData = {0};
	EventTriggerData eventTriggerData = {0};
	char	   *proc_source = NULL;
	char	   *language = "plpgsql";
	bool		is_dml_trigger;
	bool		is_event_trigger;

	assert(IsA(stmt, CreateFunctionStmt));

	foreach_ptr(DefElem, elem, stmt->options)
	{
		if (strcmp(elem->defname, "as") == 0)
		{
			assert(IsA(elem->arg, List));
			foreach_ptr(String, proc_source_str, (List *) elem->arg)
				proc_source = proc_source_str->sval;
		}
		else if (strcmp(elem->defname, "language") == 0)
		{
			language = strVal(elem->arg);
		}
	}

	if (strcmp(language, "plpgsql") != 0)
		return (PLpgSQL_function *) palloc0(sizeof(PLpgSQL_function));

	/* Mirrors the checks in upstream interpret_AS_clause */
	if (stmt->sql_body == NULL && proc_source == NULL)
		ereport(ERROR,
				(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
				 errmsg("no function body specified")));

	if (stmt->sql_body != NULL && proc_source != NULL)
		ereport(ERROR,
				(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
				 errmsg("duplicate function body specified")));

	if (stmt->sql_body != NULL)
		ereport(ERROR,
				(errcode(ERRCODE_INVALID_FUNCTION_DEFINITION),
				 errmsg("inline SQL function body only valid for language SQL")));

	wrapper = pg_query_create_function(stmt, language, proc_source,
									   &is_dml_trigger, &is_event_trigger);

	flinfo.fn_oid = InvalidOid;
	flinfo.fn_expr = NULL;
	fcinfo->flinfo = &flinfo;
	fcinfo->fncollation = InvalidOid;
	fcinfo->context = NULL;
	fcinfo->resultinfo = NULL;
	fcinfo->isnull = false;
	fcinfo->nargs = 0;

	if (is_dml_trigger)
	{
		triggerData.type = T_TriggerData;
		fcinfo->context = (Node *) &triggerData;
	}
	else if (is_event_trigger)
	{
		eventTriggerData.type = T_EventTriggerData;
		fcinfo->context = (Node *) &eventTriggerData;
	}

	function = (PLpgSQL_function *) palloc0(sizeof(PLpgSQL_function));

	plpgsql_compile_callback(fcinfo, &wrapper->tup, NULL,
							 (CachedFunction *) function, true);

	return function;
}

PgQueryInternalPlpgsqlFuncAndError pg_query_raw_parse_plpgsql(Node* stmt)
{
	PgQueryInternalPlpgsqlFuncAndError result = {0};
	MemoryContext cctx = CurrentMemoryContext;

	char stderr_buffer[STDERR_BUFFER_LEN + 1] = {0};
#ifndef DEBUG
	int stderr_global;
	int stderr_pipe[2];
#endif

#ifndef DEBUG
	// Setup pipe for stderr redirection
	if (pipe(stderr_pipe) != 0) {
		PgQueryError* error = malloc(sizeof(PgQueryError));

		error->message = strdup("Failed to open pipe, too many open file descriptors")

		result.error = error;

		return result;
	}

	fcntl(stderr_pipe[0], F_SETFL, fcntl(stderr_pipe[0], F_GETFL) | O_NONBLOCK);

	// Redirect stderr to the pipe
	stderr_global = dup(STDERR_FILENO);
	dup2(stderr_pipe[1], STDERR_FILENO);
	close(stderr_pipe[1]);
#endif

	PG_TRY();
	{
		if (IsA(stmt, CreateFunctionStmt)) {
			result.func = compile_create_function_stmt_via_callback((CreateFunctionStmt *) stmt);
		} else if (IsA(stmt, DoStmt)){
			result.func = compile_do_stmt((DoStmt *) stmt);
		} else {
			elog(ERROR, "Unexpected node type for PL/pgSQL parsing: %d", nodeTag(stmt));
		}

#ifndef DEBUG
		// Save stderr for result
		read(stderr_pipe[0], stderr_buffer, STDERR_BUFFER_LEN);
#endif

		if (strlen(stderr_buffer) > 0) {
			/* zeroed, so that the cursor position and SQLSTATE read as unset (changed for Doltgres) */
			PgQueryError* error = calloc(1, sizeof(PgQueryError));
			error->message = strdup(stderr_buffer);
			error->filename = "";
			error->funcname = "";
			error->context  = "";
			result.error = error;
		}
	}
	PG_CATCH();
	{
		ErrorData* error_data;
		PgQueryError* error;

		MemoryContextSwitchTo(cctx);
		error_data = CopyErrorData();

		// Note: This is intentionally malloc so exiting the memory context doesn't free this
		error = malloc(sizeof(PgQueryError));
		error->message   = strdup(error_data->message);
		error->filename  = strdup(error_data->filename);
		error->funcname  = strdup(error_data->funcname);
		error->context   = error_data->context ? strdup(error_data->context) : NULL;
		error->lineno    = error_data->lineno;
		error->cursorpos = error_data->cursorpos;
		error->sqlerrcode = error_data->sqlerrcode;

		result.error = error;
		FlushErrorState();
	}
	PG_END_TRY();

#ifndef DEBUG
	// Restore stderr, close pipe
	dup2(stderr_global, STDERR_FILENO);
	close(stderr_pipe[0]);
	close(stderr_global);
#endif

	return result;
}

typedef struct plStmts
{
	Node **stmts;
	int stmts_buf_size;
	int stmts_count;
} plStmts;

static bool stmts_walker(Node *node, plStmts *state)
{
	bool result;
	MemoryContext ccxt = CurrentMemoryContext;

	if (node == NULL) return false;

	if (IsA(node, CreateFunctionStmt) || IsA(node, DoStmt))
	{
		if (state->stmts_count >= state->stmts_buf_size)
		{
			state->stmts_buf_size *= 2;
			state->stmts = (Node**) repalloc(state->stmts, state->stmts_buf_size * sizeof(Node*));
		}
		state->stmts[state->stmts_count] = (Node *) node;
		state->stmts_count++;
	} else if (IsA(node, RawStmt)) {
		return stmts_walker((Node *) ((RawStmt *) node)->stmt, state);
	}

	PG_TRY();
	{
		result = raw_expression_tree_walker(node, stmts_walker, (void*) state);
	}
	PG_CATCH();
	{
		MemoryContextSwitchTo(ccxt);
		FlushErrorState();
		result = false;
	}
	PG_END_TRY();

	return result;
}

PgQueryPlpgsqlParseResult pg_query_parse_plpgsql(const char* input)
{
	MemoryContext ctx = NULL;
	PgQueryPlpgsqlParseResult result = {0};
	PgQueryInternalParsetreeAndError parse_result;
	plStmts statements;
	size_t i;

	ctx = pg_query_enter_memory_context();

	parse_result = pg_query_raw_parse(input, PG_QUERY_PARSE_DEFAULT);
	result.error = parse_result.error;
	if (result.error != NULL) {
		pg_query_exit_memory_context(ctx);
		return result;
	}

	statements.stmts_buf_size = 100;
	statements.stmts = (Node**) palloc(statements.stmts_buf_size * sizeof(Node*));
	statements.stmts_count = 0;

	stmts_walker((Node*) parse_result.tree, &statements);

	if (statements.stmts_count == 0) {
		result.plpgsql_funcs = strdup("[]");
		free(parse_result.stderr_buffer);
		pg_query_exit_memory_context(ctx);
		return result;
	}

	result.plpgsql_funcs = strdup("[\n");

	for (i = 0; i < statements.stmts_count; i++) {
		PgQueryInternalPlpgsqlFuncAndError func_and_error;

		/* lets errors in the function body point into the statement (added for Doltgres) */
		pg_query_plpgsql_source = input;
		func_and_error = pg_query_raw_parse_plpgsql(statements.stmts[i]);
		pg_query_plpgsql_source = NULL;

		// These are all malloc-ed and will survive exiting the memory context, the caller is responsible to free them now
		result.error = func_and_error.error;

		if (result.error != NULL) {
			free(parse_result.stderr_buffer);
			pg_query_exit_memory_context(ctx);
			return result;
		}

		if (func_and_error.func != NULL) {
			char *func_json;
			char *new_out;
			size_t new_out_len;

			func_json = plpgsqlToJSON(func_and_error.func);
			plpgsql_free_function_memory(func_and_error.func);

			new_out_len = strlen(result.plpgsql_funcs) + strlen(func_json) + 3;
			new_out = malloc(new_out_len);
			int n = snprintf(new_out, new_out_len, "%s%s,\n", result.plpgsql_funcs, func_json);
			if (n < 0 || n >= new_out_len) {
				PgQueryError* error = malloc(sizeof(PgQueryError));
				error->message = strdup("Failed to output PL/pgSQL functions due to snprintf failure");
				result.error = error;
			} else {
				free(result.plpgsql_funcs);
				result.plpgsql_funcs = new_out;
			}

			pfree(func_json);
		}
	}

	result.plpgsql_funcs[strlen(result.plpgsql_funcs) - 2] = '\n';
	result.plpgsql_funcs[strlen(result.plpgsql_funcs) - 1] = ']';

	free(parse_result.stderr_buffer);
	pg_query_exit_memory_context(ctx);

	return result;
}

void pg_query_free_plpgsql_parse_result(PgQueryPlpgsqlParseResult result)
{
	if (result.error) {
		pg_query_free_error(result.error);
	}

	free(result.plpgsql_funcs);
}
