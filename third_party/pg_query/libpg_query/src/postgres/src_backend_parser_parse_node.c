/*--------------------------------------------------------------------
 * Symbols referenced in this file:
 * - parser_errposition
 * - setup_parser_errposition_callback
 * - pcb_error_callback
 * - cancel_parser_errposition_callback
 *--------------------------------------------------------------------
 */

/*-------------------------------------------------------------------------
 *
 * parse_node.c
 *	  various routines that make nodes for querytrees
 *
 * Portions Copyright (c) 1996-2025, PostgreSQL Global Development Group
 * Portions Copyright (c) 1994, Regents of the University of California
 *
 *
 * IDENTIFICATION
 *	  src/backend/parser/parse_node.c
 *
 *-------------------------------------------------------------------------
 */
#include "postgres.h"

#include "access/htup_details.h"
#include "access/table.h"
#include "catalog/pg_type.h"
#include "mb/pg_wchar.h"
#include "nodes/makefuncs.h"
#include "nodes/miscnodes.h"
#include "nodes/nodeFuncs.h"
#include "nodes/subscripting.h"
#include "parser/parse_node.h"
#include "utils/builtins.h"
#include "utils/lsyscache.h"

static void pcb_error_callback(void *arg);


/*
 * make_parsestate
 *		Allocate and initialize a new ParseState.
 *
 * Caller should eventually release the ParseState via free_parsestate().
 */


/*
 * free_parsestate
 *		Release a ParseState and any subsidiary resources.
 */



/*
 * parser_errposition
 *		Report a parse-analysis-time cursor position, if possible.
 *
 * This is expected to be used within an ereport() call.  The return value
 * is a dummy (always 0, in fact).
 *
 * The locations stored in raw parsetrees are byte offsets into the source
 * string.  We have to convert them to 1-based character indexes for reporting
 * to clients.  (We do things this way to avoid unnecessary overhead in the
 * normal non-error case: computing character indexes would be much more
 * expensive than storing token offsets.)
 */
int
parser_errposition(ParseState *pstate, int location)
{
	int			pos;

	/* No-op if location was not provided */
	if (location < 0)
		return 0;
	/* Can't do anything if source text is not available */
	if (pstate == NULL || pstate->p_sourcetext == NULL)
		return 0;
	/* Convert offset to character number */
	pos = pg_mbstrlen_with_len(pstate->p_sourcetext, location) + 1;
	/* And pass it to the ereport mechanism */
	return errposition(pos);
}


/*
 * setup_parser_errposition_callback
 *		Arrange for non-parser errors to report an error position
 *
 * Sometimes the parser calls functions that aren't part of the parser
 * subsystem and can't reasonably be passed a ParseState; yet we would
 * like any errors thrown in those functions to be tagged with a parse
 * error location.  Use this function to set up an error context stack
 * entry that will accomplish that.  Usage pattern:
 *
 *		declare a local variable "ParseCallbackState pcbstate"
 *		...
 *		setup_parser_errposition_callback(&pcbstate, pstate, location);
 *		call function that might throw error;
 *		cancel_parser_errposition_callback(&pcbstate);
 */
void
setup_parser_errposition_callback(ParseCallbackState *pcbstate,
								  ParseState *pstate, int location)
{
	/* Setup error traceback support for ereport() */
	pcbstate->pstate = pstate;
	pcbstate->location = location;
	pcbstate->errcallback.callback = pcb_error_callback;
	pcbstate->errcallback.arg = pcbstate;
	pcbstate->errcallback.previous = error_context_stack;
	error_context_stack = &pcbstate->errcallback;
}

/*
 * Cancel a previously-set-up errposition callback.
 */
void
cancel_parser_errposition_callback(ParseCallbackState *pcbstate)
{
	/* Pop the error context stack */
	error_context_stack = pcbstate->errcallback.previous;
}

/*
 * Error context callback for inserting parser error location.
 *
 * Note that this will be called for *any* error occurring while the
 * callback is installed.  We avoid inserting an irrelevant error location
 * if the error is a query cancel --- are there any other important cases?
 */
static void
pcb_error_callback(void *arg)
{
	ParseCallbackState *pcbstate = (ParseCallbackState *) arg;

	if (geterrcode() != ERRCODE_QUERY_CANCELED)
		(void) parser_errposition(pcbstate->pstate, pcbstate->location);
}


/*
 * transformContainerType()
 *		Identify the actual container type for a subscripting operation.
 *
 * containerType/containerTypmod are modified if necessary to identify
 * the actual container type and typmod.  This mainly involves smashing
 * any domain to its base type, but there are some special considerations.
 * Note that caller still needs to check if the result type is a container.
 */


/*
 * transformContainerSubscripts()
 *		Transform container (array, etc) subscripting.  This is used for both
 *		container fetch and container assignment.
 *
 * In a container fetch, we are given a source container value and we produce
 * an expression that represents the result of extracting a single container
 * element or a container slice.
 *
 * Container assignments are treated basically the same as container fetches
 * here.  The caller will modify the result node to insert the source value
 * that is to be assigned to the element or slice that a fetch would have
 * retrieved.  The execution result will be a new container value with
 * the source value inserted into the right part of the container.
 *
 * For both cases, if the source is of a domain-over-container type, the
 * result is the same as if it had been of the container type; essentially,
 * we must fold a domain to its base type before applying subscripting.
 * (Note that int2vector and oidvector are treated as domains here.)
 *
 * pstate			Parse state
 * containerBase	Already-transformed expression for the container as a whole
 * containerType	OID of container's datatype (should match type of
 *					containerBase, or be the base type of containerBase's
 *					domain type)
 * containerTypMod	typmod for the container
 * indirection		Untransformed list of subscripts (must not be NIL)
 * isAssignment		True if this will become a container assignment.
 */


/*
 * make_const
 *
 *	Convert an A_Const node (as returned by the grammar) to a Const node
 *	of the "natural" type for the constant.  Note that this routine is
 *	only used when there is no explicit cast for the constant, so we
 *	have to guess what type is wanted.
 *
 *	For string literals we produce a constant of type UNKNOWN ---- whose
 *	representation is the same as cstring, but it indicates to later type
 *	resolution that we're not sure yet what type it should be considered.
 *	Explicit "NULL" constants are also typed as UNKNOWN.
 *
 *	For integers and floats we produce int4, int8, or numeric depending
 *	on the value of the number.  XXX We should produce int2 as well,
 *	but additional cleanup is needed before we can do that; there are
 *	too many examples that fail if we try.
 */

