#ifndef POSTGRES_DEPARSE_H
#define POSTGRES_DEPARSE_H

#include <stdbool.h>
#include <sys/types.h>

typedef struct PostgresDeparseComment {
    int match_location;          // Insert comment before a node, once we find a node whose location field is equal-or-higher than this location
    int newlines_before_comment; // Insert newlines before inserting the comment (set to non-zero if the source comment was separated from the prior token by at least one newline)
    int newlines_after_comment;  // Insert newlines after inserting the comment (set to non-zero if the source comment was separated from the next token by at least one newline)
    char *str;                   // The actual comment string, including comment start/end tokens, and newline characters in comment (if any)
} PostgresDeparseComment;

typedef struct PostgresDeparseOpts {
    PostgresDeparseComment **comments;
    size_t comment_count;

    // Pretty print options
    bool pretty_print;
    int indent_size;           // Indentation size (Default 4 spaces)
    int max_line_length;       // Restricts the line length of certain lists of items (Default 80 characters)
    bool trailing_newline;     // Whether to add a trailing newline at the end of the output (Default off)
    bool commas_start_of_line; // Place separating commas at start of line (Default off)
} PostgresDeparseOpts;

/* Forward declarations to allow referencing the structs in this include file without needing Postgres includes */
struct StringInfoData;
typedef struct StringInfoData *StringInfo;
struct RawStmt;

extern void deparseRawStmt(StringInfo str, struct RawStmt *raw_stmt);
extern void deparseRawStmtOpts(StringInfo str, struct RawStmt *raw_stmt, PostgresDeparseOpts opts);

#endif
