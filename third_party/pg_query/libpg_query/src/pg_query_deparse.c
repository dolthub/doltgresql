#include "pg_query.h"
#include "pg_query_internal.h"
#include "pg_query_readfuncs.h"

#include "postgres_deparse.h"

#include "postgres.h"
#include "lib/stringinfo.h"
#include "nodes/parsenodes.h"

#include "protobuf/pg_query.pb-c.h"

PgQueryDeparseResult
pg_query_deparse_protobuf(PgQueryProtobuf parse_tree)
{
	PostgresDeparseOpts opts;

	MemSet(&opts, 0, sizeof(PostgresDeparseOpts));
	return pg_query_deparse_protobuf_opts(parse_tree, opts);
}

PgQueryDeparseResult
pg_query_deparse_protobuf_opts(PgQueryProtobuf parse_tree, PostgresDeparseOpts opts)
{
	PgQueryDeparseResult result = {0};
	StringInfoData str;
	MemoryContext ctx;
	List	   *stmts;
	ListCell   *lc;

	ctx = pg_query_enter_memory_context();

	PG_TRY();
	{
		stmts = pg_query_protobuf_to_nodes(parse_tree);

		initStringInfo(&str);

		foreach(lc, stmts)
		{
			deparseRawStmtOpts(&str, castNode(RawStmt, lfirst(lc)), opts);
			if (lnext(stmts, lc))
				appendStringInfoString(&str, "; ");
		}
		result.query = strdup(str.data);
	}
	PG_CATCH();
	{
		ErrorData  *error_data;
		PgQueryError *error;

		MemoryContextSwitchTo(ctx);
		error_data = CopyErrorData();

		/*
		 * Note: This is intentionally malloc so exiting the memory context
		 * doesn't free this
		 */
		error = malloc(sizeof(PgQueryError));
		error->message = strdup(error_data->message);
		error->filename = strdup(error_data->filename);
		error->funcname = strdup(error_data->funcname);
		error->context = NULL;
		error->lineno = error_data->lineno;
		error->cursorpos = error_data->cursorpos;

		result.error = error;
		FlushErrorState();
	}
	PG_END_TRY();

	pg_query_exit_memory_context(ctx);

	return result;
}

void
pg_query_free_deparse_result(PgQueryDeparseResult result)
{
	if (result.error)
	{
		pg_query_free_error(result.error);
	}

	free(result.query);
}

PgQueryDeparseCommentsResult
pg_query_deparse_comments_for_query(const char *query)
{
	PgQueryDeparseCommentsResult result = {0};
	PgQueryScanResult scan_result_raw = pg_query_scan(query);

	if (scan_result_raw.error)
	{
		result.error = scan_result_raw.error;
		return result;
	}

	PgQuery__ScanResult *scan_result = pg_query__scan_result__unpack(NULL, scan_result_raw.pbuf.len, (void *) scan_result_raw.pbuf.data);
	bool		prior_token_was_comment = false;
	int32_t		prior_non_comment_end = 0;
	int32_t		prior_token_end = 0;

	result.comment_count = 0;
	for (int i = 0; i < scan_result->n_tokens; i++)
	{
		PgQuery__ScanToken *token = scan_result->tokens[i];

		if (token->token == PG_QUERY__TOKEN__SQL_COMMENT || token->token == PG_QUERY__TOKEN__C_COMMENT)
			result.comment_count++;
	}

	result.comments = malloc(result.comment_count * sizeof(PostgresDeparseComment *));
	size_t		comment_idx = 0;

	for (int i = 0; i < scan_result->n_tokens; i++)
	{
		PgQuery__ScanToken *token = scan_result->tokens[i];

		if (token->token == PG_QUERY__TOKEN__SQL_COMMENT || token->token == PG_QUERY__TOKEN__C_COMMENT)
		{
			size_t		token_len = token->end - token->start;
			PostgresDeparseComment *comment = malloc(sizeof(PostgresDeparseComment));

			/*
			 * Set to the end of the prior non-comment token.
			 *
			 * We could alternatively set the match location to the start of
			 * the next non-comment token instead (which would be more in line
			 * with emitting that happens before the node once the location is
			 * matched).
			 *
			 * However, that doesn't work well in practice, specifically.
			 * RawStmt start locations are 0, even when there are comments at
			 * the start of the statement.
			 */
			comment->match_location = prior_non_comment_end;

			comment->newlines_before_comment = 0;
			comment->newlines_after_comment = 0;

			/*
			 * Emit newlines before the comment if prior token was not a
			 * comment (otherwise that comment would have emitted them)
			 */
			if (!prior_token_was_comment)
			{
				for (int j = prior_token_end; j < token->start; j++)
				{
					if (query[j] == '\n')
						comment->newlines_before_comment++;
				}
			}

			if (i < scan_result->n_tokens - 1)
			{
				for (int j = token->end; j < scan_result->tokens[i + 1]->start; j++)
				{
					if (query[j] == '\n')
						comment->newlines_after_comment++;
				}
			}

			comment->str = malloc(token_len + 1);
			memcpy(comment->str, &(query[token->start]), token_len);
			comment->str[token_len] = '\0';

			result.comments[comment_idx] = comment;
			comment_idx++;
			prior_token_was_comment = true;
		}
		else
		{
			prior_non_comment_end = token->end;
			prior_token_was_comment = false;
		}
		prior_token_end = token->end;
	}

	pg_query__scan_result__free_unpacked(scan_result, NULL);
	pg_query_free_scan_result(scan_result_raw);

	return result;
}

void
pg_query_free_deparse_comments_result(PgQueryDeparseCommentsResult result)
{
	for (int i = 0; i < result.comment_count; i++)
	{
		free(result.comments[i]->str);
		free(result.comments[i]);
	}
	if (result.comments != NULL)
		free(result.comments);
}
