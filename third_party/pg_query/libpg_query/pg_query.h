#ifndef PG_QUERY_H
#define PG_QUERY_H

#include <stdbool.h>
#include <stdint.h>
#include <sys/types.h>

#include "postgres_deparse.h"
#include "pg_query_scan_tokens.h"

typedef struct {
	char* message; // exception message
	char* funcname; // source function of exception (e.g. SearchSysCache)
	char* filename; // source of exception (e.g. parse.l)
	int lineno; // source of exception (e.g. 104)
	int cursorpos; // char in query at which exception occurred
	char* context; // additional context (optional, can be NULL)
	int sqlerrcode; // encoded SQLSTATE of the exception, set by pg_query_parse_protobuf (added for Doltgres)
} PgQueryError;

typedef struct {
	int length;
	bool *items;
	PgQueryError* error;
} PgQueryIsUtilityResult;

typedef struct {
  size_t len;
  char* data;
} PgQueryProtobuf;

typedef struct {
  PgQueryProtobuf pbuf;
  char* stderr_buffer;
  PgQueryError* error;
} PgQueryScanResult;

typedef struct {
  int start; // byte offset of the token in the input
  int end; // byte offset just past the token
  PgQueryToken token; // use pg_query_token_name to get the name as a string
  PgQueryKeywordKind keyword_kind; // use pg_query_keyword_kind_name to get the name as a string
} PgQueryScanToken;

typedef struct {
  PgQueryScanToken* tokens;
  int n_tokens;
  PgQueryError* error;
} PgQueryScanTokensResult;

typedef struct {
  char* parse_tree;
  char* stderr_buffer;
  PgQueryError* error;
} PgQueryParseResult;

typedef struct {
  PgQueryProtobuf parse_tree;
  char* stderr_buffer;
  PgQueryError* error;
} PgQueryProtobufParseResult;

typedef struct {
  int stmt_location;
  int stmt_len;
} PgQuerySplitStmt;

typedef struct {
  PgQuerySplitStmt **stmts;
  int n_stmts;
  char* stderr_buffer;
  PgQueryError* error;
} PgQuerySplitResult;

typedef struct {
  char* query;
  PgQueryError* error;
} PgQueryDeparseResult;

typedef struct {
  PostgresDeparseComment **comments;
  size_t comment_count;
  PgQueryError* error;
} PgQueryDeparseCommentsResult;

typedef struct {
  char* plpgsql_funcs;
  PgQueryError* error;
} PgQueryPlpgsqlParseResult;

typedef struct {
  uint64_t fingerprint;
  char* fingerprint_str;
  char* stderr_buffer;
  PgQueryError* error;
} PgQueryFingerprintResult;

typedef struct {
  char* normalized_query;
  PgQueryError* error;
} PgQueryNormalizeResult;

typedef struct {
	PgQueryProtobuf summary;
	char* stderr_buffer;
	PgQueryError* error;
} PgQuerySummaryParseResult;

// Postgres parser options (parse mode and GUCs that affect parsing)

typedef enum
{
	PG_QUERY_PARSE_DEFAULT = 0,
	PG_QUERY_PARSE_TYPE_NAME,
	PG_QUERY_PARSE_PLPGSQL_EXPR,
	PG_QUERY_PARSE_PLPGSQL_ASSIGN1,
	PG_QUERY_PARSE_PLPGSQL_ASSIGN2,
	PG_QUERY_PARSE_PLPGSQL_ASSIGN3
} PgQueryParseMode;

// We technically only need 3 bits to store parse mode, but
// having 4 bits avoids API breaks if another one gets added.
#define PG_QUERY_PARSE_MODE_BITS 4
#define PG_QUERY_PARSE_MODE_BITMASK ((1 << PG_QUERY_PARSE_MODE_BITS) - 1)

#define PG_QUERY_DISABLE_BACKSLASH_QUOTE 16 // backslash_quote = off (default is safe_encoding, which is effectively on)
#define PG_QUERY_DISABLE_STANDARD_CONFORMING_STRINGS 32 // standard_conforming_strings = off (default is on)
#define PG_QUERY_DISABLE_ESCAPE_STRING_WARNING 64 // escape_string_warning = off (default is on)

// Fingerprint options (flags that control how fingerprints are calculated)

#define PG_QUERY_FINGERPRINT_DEFAULT 0

// RangeVar handling (bits 0-3)
//
// By default, relation references are fingerprinted following the Postgres 18+
// query ID behavior (see Postgres commit 787514b30bb): in SELECT/DML statements
// the alias name is fingerprinted, the relation name is ignored when an alias
// is present, and schema names are ignored.
#define PG_QUERY_FINGERPRINT_RANGEVAR_IGNORE_ALIASES (1 << 0) // Relation names are always fingerprinted, aliases are ignored
#define PG_QUERY_FINGERPRINT_RANGEVAR_INCLUDE_SCHEMA (1 << 1) // Schema names are also fingerprinted in SELECT/DML statements (they are always fingerprinted in utility statements)

// Convenience combination that matches how Postgres 17 and earlier calculate
// query IDs, and how libpg_query 17 and earlier calculated fingerprints
#define PG_QUERY_FINGERPRINT_RANGEVAR_PG17_COMPAT (PG_QUERY_FINGERPRINT_RANGEVAR_IGNORE_ALIASES | PG_QUERY_FINGERPRINT_RANGEVAR_INCLUDE_SCHEMA)

// Relation name handling (bits 4-7)
//
// By default, sequences of two or more digits in relation names are ignored
// when fingerprinting, so that queries on date/number-suffixed tables (e.g.
// partitions like "orders_2024_01") get the same fingerprint. Set this flag
// to fingerprint the full relation name instead.
#define PG_QUERY_FINGERPRINT_FULL_RELNAME (1 << 4)

#ifdef __cplusplus
extern "C" {
#endif

PgQueryNormalizeResult pg_query_normalize(const char* input);
PgQueryNormalizeResult pg_query_normalize_utility(const char* input);

// pg_query_scan returns the tokens as a serialized ScanResult protobuf (for
// decoding in another language), pg_query_scan_tokens returns them as a plain
// C array (for use from C).
PgQueryScanResult pg_query_scan(const char* input);
PgQueryScanTokensResult pg_query_scan_tokens(const char* input);
const char* pg_query_token_name(PgQueryToken token);
const char* pg_query_keyword_kind_name(PgQueryKeywordKind keyword_kind);

PgQueryParseResult pg_query_parse(const char* input);
PgQueryParseResult pg_query_parse_opts(const char* input, int parser_options);
PgQueryProtobufParseResult pg_query_parse_protobuf(const char* input);
PgQueryProtobufParseResult pg_query_parse_protobuf_opts(const char* input, int parser_options);
PgQueryPlpgsqlParseResult pg_query_parse_plpgsql(const char* input);

PgQueryFingerprintResult pg_query_fingerprint(const char* input);
PgQueryFingerprintResult pg_query_fingerprint_opts(const char* input, int parser_options, int fingerprint_options);

// Use pg_query_split_with_scanner when you need to split statements that may
// contain parse errors, otherwise pg_query_split_with_parser is recommended
// for improved accuracy due the parser adding additional token handling.
//
// Note that we try to support special cases like comments, strings containing
// ";" on both, as well as oddities like "CREATE RULE .. (SELECT 1; SELECT 2);"
// which is treated as as single statement.
PgQuerySplitResult pg_query_split_with_scanner(const char *input);
PgQuerySplitResult pg_query_split_with_parser(const char *input);

PgQueryDeparseResult pg_query_deparse_protobuf(PgQueryProtobuf parse_tree);
PgQueryDeparseResult pg_query_deparse_protobuf_opts(PgQueryProtobuf parse_tree, struct PostgresDeparseOpts opts);
PgQueryDeparseCommentsResult pg_query_deparse_comments_for_query(const char *query);

PgQueryIsUtilityResult pg_query_is_utility_stmt(const char *query);

PgQuerySummaryParseResult pg_query_summary(const char* input, int parser_options, int truncate_limit);

void pg_query_free_normalize_result(PgQueryNormalizeResult result);
void pg_query_free_scan_result(PgQueryScanResult result);
void pg_query_free_scan_tokens_result(PgQueryScanTokensResult result);
void pg_query_free_parse_result(PgQueryParseResult result);
void pg_query_free_split_result(PgQuerySplitResult result);
void pg_query_free_deparse_result(PgQueryDeparseResult result);
void pg_query_free_deparse_comments_result(PgQueryDeparseCommentsResult result);
void pg_query_free_protobuf_parse_result(PgQueryProtobufParseResult result);
void pg_query_free_plpgsql_parse_result(PgQueryPlpgsqlParseResult result);
void pg_query_free_fingerprint_result(PgQueryFingerprintResult result);
void pg_query_free_is_utility_result(PgQueryIsUtilityResult result);
void pg_query_free_summary_parse_result(PgQuerySummaryParseResult result);

// Optional, cleans up the top-level memory context (automatically done for threads that exit)
void pg_query_exit(void);

// Postgres version information
#define PG_MAJORVERSION "18"
#define PG_VERSION "18.6"
#define PG_VERSION_NUM 180006

// Deprecated APIs below

void pg_query_init(void); // Deprecated as of 9.5-1.4.1, this is now run automatically as needed

#ifdef __cplusplus
}
#endif

#endif
