#ifndef PG_QUERY_SUMMARY_H
#define PG_QUERY_SUMMARY_H

#include "pg_query.h"
#include "pg_query_internal.h"

#include "nodes/primnodes.h"

#include <unistd.h>

typedef enum
{
	CONTEXT_NONE = 0,
	CONTEXT_SELECT = 1,
	CONTEXT_DML = 2,
	CONTEXT_DDL = 3,
	CONTEXT_CALL = 4,
}			ContextType;

typedef struct
{
	char	   *name;
	char	   *schema_name;
	char	   *table_name;
	ContextType context;
}			SummaryTable;

typedef struct
{
	char	   *key;
	char	   *value;
}			SummaryAlias;

typedef struct
{
	char	   *name;
	char	   *function_name;
	char	   *schema_name;	/* optional */


	ContextType context;
}			SummaryFunction;

typedef struct
{
	char	   *schema_name;
	char	   *table_name;
	char	   *column;
}			FilterColumn;

typedef struct
{
	List	   *tables;			/* List of SummaryTable */

	List	   *aliases;		/* List of SummaryAlias */
	List	   *cte_names;		/* List of char */
	List	   *functions;		/* List of SummaryFunction */
	List	   *filter_columns; /* List of FilterColumn */
	List	   *statement_types;	/* List of char * */

	char	   *truncated_query;
}			Summary;

typedef struct
{
	List	   *tables;			/* List of Table * */
	List	   *aliases;		/* List of SummaryAlias * */
	List	   *functions;		/* List of Function * */
	List	   *range_vars;		/* List of RangeVarWithContext * */
	List	   *cte_names;		/* list of CommonTableExpr */
	List	   *filter_columns; /* List of FilterColumn * */
	bool		save_filter_columns;	/* If true, we should be adding to
										 * filter columns. */
}			WalkState;

typedef struct
{
	RangeVar   *node;
	ContextType context;
}			RangeVarWithContext;

/*
 * This is like PgQuerySummaryParseResult, but the summary is not wrapped
 * up in protobuf.
 */
typedef struct
{
	Summary		summary;
	char	   *stderr_buffer;
	PgQueryError *error;
}			PgQuerySummaryParseResultInternal;

extern bool pg_query_raw_tree_walker_supports(Node *node);

extern void pg_query_summary_statement_walk(Summary * summary, Node *node);

extern void pg_query_summary_truncate(Summary * summary, Node *node, int truncate_limit);

/*
 * This is like pg_query_summary(), but returns a (non-protobuf)
 * PgQuerySummaryParseResultInternal instead of the usual
 * PgQuerySummaryParseResult which converts the summary to protobuf.
 *
 * This is used for tests, to avoid having to process protobuf in C.
 */
extern PgQuerySummaryParseResultInternal pg_query_summary_internal(const char *input, int parser_options, int truncate_limit);
extern void pg_query_free_summary_parse_result_internal(PgQuerySummaryParseResultInternal result);

#endif
