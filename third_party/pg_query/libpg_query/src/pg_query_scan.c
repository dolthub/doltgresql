#include "pg_query.h"
#include "pg_query_internal.h"

#include "gramparse.h"
#include "lib/stringinfo.h"

#include "protobuf/pg_query.upb.h"

#include <unistd.h>
#include <fcntl.h>

/* This is ugly. We need to access yyleng outside of scan.l, and casting yyscanner
   to this internal struct seemed like one way to do it... */
struct yyguts_t
{
  void *yyextra_r;
  FILE *yyin_r, *yyout_r;
  size_t yy_buffer_stack_top; /**< index of top of stack. */
  size_t yy_buffer_stack_max; /**< capacity of stack. */
  struct yy_buffer_state *yy_buffer_stack; /**< Stack as an array. */
  char yy_hold_char;
  size_t yy_n_chars;
  size_t yyleng_r;
};

/*
 * Runs the scanner over the input and returns the tokens as a palloc'd array
 * in the current memory context. Scanner errors are thrown (via elog), so this
 * must be called inside PG_TRY.
 */
static PgQueryScanToken *
scan_tokens(const char *input, int *n_tokens_out)
{
  core_yyscan_t yyscanner;
  core_yy_extra_type yyextra;
  core_YYSTYPE yylval;
  YYLTYPE    yylloc;
  int capacity = 64;
  int n_tokens = 0;
  PgQueryScanToken *tokens = palloc(sizeof(PgQueryScanToken) * capacity);

  /* initialize the flex scanner --- should match raw_parser() */
  yyscanner = scanner_init(input, &yyextra, &ScanKeywords, ScanKeywordTokens);

  /* return comments as tokens, unlike raw_parser() */
  yyextra.return_comments = true;

  /* Lex tokens  */
  for (;;)
  {
    int tok;
    PgQueryScanToken *token;

    tok = core_yylex(&yylval, &yylloc, yyscanner);
    if (tok == 0) break;

    if (n_tokens == capacity)
    {
      capacity *= 2;
      tokens = repalloc(tokens, sizeof(PgQueryScanToken) * capacity);
    }
    token = &tokens[n_tokens++];

    token->start = yylloc;
    if (tok == SCONST || tok == USCONST || tok == BCONST || tok == XCONST || tok == IDENT || tok == UIDENT || tok == C_COMMENT) {
      token->end = yyextra.yyllocend;
    } else {
      token->end = yylloc + ((struct yyguts_t*) yyscanner)->yyleng_r;
    }
    token->token = tok;

    switch (tok) {
    #define PG_KEYWORD(a,b,c,d) case b: token->keyword_kind = c + 1; break;
    #include "parser/kwlist.h"
    #undef PG_KEYWORD
    default: token->keyword_kind = 0;
    }
  }

  scanner_finish(yyscanner);

  *n_tokens_out = n_tokens;
  return tokens;
}

PgQueryScanResult pg_query_scan(const char* input)
{
  MemoryContext ctx = NULL;
  PgQueryScanResult result = {0};
  /* volatile: assigned inside PG_TRY and read in PG_CATCH (after longjmp) */
  upb_Arena *volatile arena = NULL;
  pg_query_ScanResult *scan_result;

  ctx = pg_query_enter_memory_context();

  MemoryContext parse_context = CurrentMemoryContext;

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
    size_t packed_len;
    char *packed_data;
    int n_tokens;
    PgQueryScanToken *tokens = scan_tokens(input, &n_tokens);

    arena = upb_Arena_New();
    scan_result = pg_query_ScanResult_new(arena);
    pg_query_ScanResult_set_version(scan_result, PG_VERSION_NUM);

    for (int i = 0; i < n_tokens; i++)
    {
      pg_query_ScanToken *scan_token = pg_query_ScanResult_add_tokens(scan_result, arena);

      pg_query_ScanToken_set_start(scan_token, tokens[i].start);
      pg_query_ScanToken_set_end(scan_token, tokens[i].end);
      pg_query_ScanToken_set_token(scan_token, tokens[i].token);
      pg_query_ScanToken_set_keyword_kind(scan_token, tokens[i].keyword_kind);
    }

    /* Serialize into the arena, then copy out into a malloc'd buffer that
     * survives exiting the memory context (freed in pg_query_free_scan_result).
     * Serialization fails (returns NULL) e.g. when the arena runs out of memory;
     * PG_CATCH below frees the arena. */
    packed_data = pg_query_ScanResult_serialize(scan_result, arena, &packed_len);
    if (packed_data == NULL)
      elog(ERROR, "could not serialize scan result to protobuf");

    result.pbuf.len = packed_len;
    result.pbuf.data = malloc(packed_len);
    memcpy(result.pbuf.data, packed_data, packed_len);

    upb_Arena_Free(arena);
    arena = NULL;

#ifndef DEBUG
    // Save stderr for result
    read(stderr_pipe[0], stderr_buffer, STDERR_BUFFER_LEN);
#endif

    result.stderr_buffer = strdup(stderr_buffer);
  }
  PG_CATCH();
  {
    ErrorData* error_data;
    PgQueryError* error;

    MemoryContextSwitchTo(parse_context);
    error_data = CopyErrorData();

    // The arena is malloc-backed, so it is not freed with the memory context
    if (arena != NULL) {
      upb_Arena_Free(arena);
      arena = NULL;
    }

    // Note: This is intentionally malloc so exiting the memory context doesn't free this
    error = malloc(sizeof(PgQueryError));
    error->message   = strdup(error_data->message);
    error->filename  = strdup(error_data->filename);
    error->funcname  = strdup(error_data->funcname);
    error->context   = NULL;
    error->lineno    = error_data->lineno;
    error->cursorpos = error_data->cursorpos;

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

  pg_query_exit_memory_context(ctx);

  return result;
}

void pg_query_free_scan_result(PgQueryScanResult result)
{
  if (result.error) {
    pg_query_free_error(result.error);
  }

  free(result.pbuf.data);
  free(result.stderr_buffer);
}

/*
 * Like pg_query_scan, but returns the tokens as a plain C array instead of a
 * serialized protobuf, so C callers don't need to decode the result themselves.
 */
PgQueryScanTokensResult pg_query_scan_tokens(const char* input)
{
  MemoryContext ctx = NULL;
  PgQueryScanTokensResult result = {0};

  ctx = pg_query_enter_memory_context();

  PG_TRY();
  {
    int n_tokens;
    PgQueryScanToken *tokens = scan_tokens(input, &n_tokens);

    result.n_tokens = n_tokens;
    if (n_tokens > 0)
    {
      // Note: This is intentionally malloc so exiting the memory context doesn't free this
      result.tokens = malloc(sizeof(PgQueryScanToken) * n_tokens);
      memcpy(result.tokens, tokens, sizeof(PgQueryScanToken) * n_tokens);
    }
  }
  PG_CATCH();
  {
    ErrorData* error_data;
    PgQueryError* error;

    MemoryContextSwitchTo(ctx);
    error_data = CopyErrorData();

    // Note: This is intentionally malloc so exiting the memory context doesn't free this
    error = malloc(sizeof(PgQueryError));
    error->message   = strdup(error_data->message);
    error->filename  = strdup(error_data->filename);
    error->funcname  = strdup(error_data->funcname);
    error->context   = NULL;
    error->lineno    = error_data->lineno;
    error->cursorpos = error_data->cursorpos;

    result.error = error;
    FlushErrorState();
  }
  PG_END_TRY();

  pg_query_exit_memory_context(ctx);

  return result;
}

void pg_query_free_scan_tokens_result(PgQueryScanTokensResult result)
{
  if (result.error) {
    pg_query_free_error(result.error);
  }

  free(result.tokens);
}

#include "pg_query_scan_defs.c"
