#include "pg_query.h"
#include "pg_query_internal.h"
#include "pg_query_readfuncs.h"

#include "postgres_deparse.h"

#include "postgres.h"
#include "lib/stringinfo.h"
#include "nodes/parsenodes.h"
#include "gramparse.h"

#include <limits.h>

static PostgresDeparseOpts * copy_deparse_opts_for_stmt(RawStmt *raw_stmt, PostgresDeparseOpts * opts, size_t start, size_t end);

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
	MemoryContext ctx = pg_query_enter_memory_context();

	PG_TRY();
	{
		StringInfoData str;
		List	   *stmts = pg_query_protobuf_to_nodes(parse_tree);
		size_t		prev_end = 0;

		initStringInfo(&str);

		foreach_ptr(RawStmt, raw_stmt, stmts)
		{
			PostgresDeparseOpts *stmt_opts = &opts;
			bool		is_last = foreach_current_index(raw_stmt) == (list_length(stmts) - 1);

			/*
			 * If we have more than one statement, and we have comments, we
			 * split them up by statement
			 */
			if (list_length(stmts) > 1 && opts.comment_count > 0)
			{
				size_t		end = is_last ? INT_MAX : (raw_stmt->stmt_location + raw_stmt->stmt_len);

				stmt_opts = copy_deparse_opts_for_stmt(raw_stmt, &opts, prev_end, end);
				prev_end = end;
			}

			deparseRawStmtOpts(&str, raw_stmt, stmt_opts);
			if (!is_last)
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

static PostgresDeparseOpts *
copy_deparse_opts_for_stmt(RawStmt *raw_stmt, PostgresDeparseOpts * opts, size_t start, size_t end)
{
	PostgresDeparseOpts *stmt_opts = palloc(sizeof(PostgresDeparseOpts));

	memcpy(stmt_opts, opts, sizeof(PostgresDeparseOpts));

	stmt_opts->comments = palloc0(sizeof(PostgresDeparseComment *) * opts->comment_count);
	stmt_opts->comment_count = 0;

	for (int i = 0; i < opts->comment_count; i++)
	{
		if (opts->comments[i]->match_location >= start && opts->comments[i]->match_location < end)
		{
			stmt_opts->comments[stmt_opts->comment_count] = opts->comments[i];
			stmt_opts->comment_count++;
		}
	}

	return stmt_opts;
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
	PgQueryScanTokensResult scan_result = pg_query_scan_tokens(query);
	const PgQueryScanToken *tokens = scan_result.tokens;
	int			n_tokens = scan_result.n_tokens;
	bool		prior_token_was_comment = false;
	int32_t		prior_non_comment_end = 0;
	int32_t		prior_token_end = 0;

	if (scan_result.error)
	{
		result.error = scan_result.error;
		return result;
	}

	result.comment_count = 0;
	for (int i = 0; i < n_tokens; i++)
	{
		if (tokens[i].token == SQL_COMMENT || tokens[i].token == C_COMMENT)
			result.comment_count++;
	}

	result.comments = malloc(result.comment_count * sizeof(PostgresDeparseComment *));
	size_t		comment_idx = 0;

	for (int i = 0; i < n_tokens; i++)
	{
		const PgQueryScanToken *token = &tokens[i];

		if (token->token == SQL_COMMENT || token->token == C_COMMENT)
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

			if (i < n_tokens - 1)
			{
				for (int j = token->end; j < tokens[i + 1].start; j++)
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

	pg_query_free_scan_tokens_result(scan_result);

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
