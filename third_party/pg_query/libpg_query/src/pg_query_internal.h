#ifndef PG_QUERY_INTERNAL_H
#define PG_QUERY_INTERNAL_H

#include "postgres.h"
#include "utils/memutils.h"
#include "nodes/pg_list.h"

#define STDERR_BUFFER_LEN 4096
#define DEBUG

/*
 * Maximum protobuf message nesting depth, used for both encoding
 * (pg_query_nodes_to_protobuf) and decoding (pg_query_protobuf_to_nodes).
 *
 * Every expression level costs two message levels (the Node wrapper plus the
 * node message itself), so this allows roughly 5,000 levels of nested
 * expressions (e.g. "a || b || c ..."), far more than upb's default of 100.
 *
 * This is only a sanity cap: what actually prevents stack overflows is the
 * stack depth limit (max_stack_depth, sized per call to the available stack),
 * which upb's encoder and decoder also check (see
 * vendor/upb/patches/0003-stack-depth-check.patch), just like our recursive
 * tree walkers do via check_stack_depth(). Depending on the available stack
 * and the stack used per level, that limit may be reached first.
 */
#define PG_QUERY_PROTOBUF_MAX_DEPTH 10000

typedef struct
{
	List	   *tree;
	char	   *stderr_buffer;
	PgQueryError *error;
}			PgQueryInternalParsetreeAndError;

PgQueryInternalParsetreeAndError pg_query_raw_parse(const char *input, int parser_options);

void		pg_query_free_error(PgQueryError * error);

MemoryContext pg_query_enter_memory_context();
void		pg_query_exit_memory_context(MemoryContext ctx);

#endif
