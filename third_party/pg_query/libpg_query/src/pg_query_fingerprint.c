#include <stdio.h>

#include "pg_query.h"
#include "pg_query_internal.h"
#include "pg_query_fingerprint.h"

#include "postgres.h"
#include "xxhash/xxhash.h"
#include "lib/ilist.h"

#include "parser/parser.h"
#include "parser/scanner.h"
#include "parser/scansup.h"

#include "nodes/parsenodes.h"
#include "nodes/value.h"

#include "common/hashfn.h"

#include <unistd.h>
#include <fcntl.h>

// Definitions

typedef struct FingerprintContext
{
	XXH3_state_t *xxh_state;

	struct listsort_cache_hash *listsort_cache;

	int fingerprint_options;

	bool write_tokens;
	dlist_head tokens;
} FingerprintContext;

typedef struct FingerprintListsortItem
{
	XXH64_hash_t hash;
	size_t list_pos;
} FingerprintListsortItem;

typedef struct FingerprintListsortItemCacheEntry
{
	/* List node this cache entry is for */
	uintptr_t node;

	/* Hashes of all list items -- this is expensive to calculate */
	FingerprintListsortItem **listsort_items;
	size_t listsort_items_size;

	/* hash entry status */
	char status;
} FingerprintListsortItemCacheEntry;

#define SH_PREFIX listsort_cache
#define SH_ELEMENT_TYPE FingerprintListsortItemCacheEntry
#define SH_KEY_TYPE uintptr_t
#define SH_KEY node
#define SH_HASH_KEY(tb, key) hash_bytes((const unsigned char *) &key, sizeof(uintptr_t))
#define SH_EQUAL(tb, a, b) a == b
#define SH_SCOPE static inline
#define SH_DEFINE
#define SH_DECLARE
#include "lib/simplehash.h"

typedef struct FingerprintToken
{
	char *str;
	dlist_node list_node;
} FingerprintToken;

static void _fingerprintNode(FingerprintContext *ctx, const void *obj, const void *parent, char *parent_field_name, unsigned int depth);
static void _fingerprintInitContext(FingerprintContext *ctx, FingerprintContext *parent, int fingerprint_options, bool write_tokens);
static void _fingerprintFreeContext(FingerprintContext *ctx);

#define PG_QUERY_FINGERPRINT_VERSION 3

// Implementations

static void
_fingerprintString(FingerprintContext *ctx, const char *str)
{
	if (ctx->xxh_state != NULL) {
		XXH3_64bits_update(ctx->xxh_state, str, strlen(str));
	}

	if (ctx->write_tokens) {
		FingerprintToken *token = palloc0(sizeof(FingerprintToken));
		token->str = pstrdup(str);
		dlist_push_tail(&ctx->tokens, &token->list_node);
	}
}

static void
_fingerprintInteger(FingerprintContext *ctx, const union ValUnion *value)
{
	if (value->ival.ival != 0) {
		_fingerprintString(ctx, "Integer");
		_fingerprintString(ctx, "ival");
		char buffer[50];
		sprintf(buffer, "%d", value->ival.ival);
		_fingerprintString(ctx, buffer);
	}
}

static void
_fingerprintFloat(FingerprintContext *ctx, const union ValUnion *value)
{
	if (value->fval.fval != NULL) {
		// NB: We output `str` here intentionally, to match the output format from libpg_query 14
		// and below. This results in stable fingerprints, despite the field name being changed in
		// PG15 to `fval`.
		_fingerprintString(ctx, "Float");
		_fingerprintString(ctx, "str");
		_fingerprintString(ctx, value->fval.fval);
	}
}

static void
_fingerprintBoolean(FingerprintContext *ctx, const union ValUnion *value)
{
	_fingerprintString(ctx, "Boolean");
	_fingerprintString(ctx, "boolval");
	_fingerprintString(ctx, value->boolval.boolval ? "true" : "false");
}

static void
_fingerprintBitString(FingerprintContext *ctx, const union ValUnion *value)
{
	if (value->bsval.bsval != NULL) {
		// NB: We output `str` here intentionally, to match the output format from libpg_query 14
		// and below. This results in stable fingerprints, despite the field name being changed in
		// PG15 to `bsval`.
		_fingerprintString(ctx, "BitString");
		_fingerprintString(ctx, "str");
		_fingerprintString(ctx, value->bsval.bsval);
	}
}

static int compareFingerprintListsortItem(const void *a, const void *b)
{
	FingerprintListsortItem *ca = *(FingerprintListsortItem**) a;
	FingerprintListsortItem *cb = *(FingerprintListsortItem**) b;
	if (ca->hash > cb->hash)
		return 1;
	else if (ca->hash < cb->hash)
		return -1;
	return 0;
}

static void
_fingerprintList(FingerprintContext *ctx, const List *node, const void *parent, char *field_name, unsigned int depth)
{
	if (field_name != NULL && (strcmp(field_name, "fromClause") == 0 || strcmp(field_name, "targetList") == 0 ||
		strcmp(field_name, "cols") == 0 || strcmp(field_name, "rexpr") == 0 || strcmp(field_name, "valuesLists") == 0 ||
		strcmp(field_name, "args") == 0))
	{
		/*
		 * Check for cached values for the hashes of subnodes
		 *
		 * Note this cache is important so we avoid exponential runtime behavior,
		 * which would be the case if we fingerprinted each node twice, which
		 * then would also again have to fingerprint each of its subnodes twice,
		 * etc., leading to deep nodes to be fingerprinted many many times over.
		 *
		 * We have seen real-world problems with this logic here without
		 * a cache in place.
		 */
		FingerprintListsortItem** listsort_items = NULL;
		size_t listsort_items_size = 0;
		FingerprintListsortItemCacheEntry *entry = listsort_cache_lookup(ctx->listsort_cache, (uintptr_t) node);
		if (entry != NULL)
		{
			listsort_items = entry->listsort_items;
			listsort_items_size = entry->listsort_items_size;
		}
		else
		{
			listsort_items = palloc0(node->length * sizeof(FingerprintListsortItem*));
			listsort_items_size = 0;
			ListCell *lc;
			bool found;

			foreach(lc, node)
			{
				FingerprintContext fctx;
				FingerprintListsortItem* lctx = palloc0(sizeof(FingerprintListsortItem));

				_fingerprintInitContext(&fctx, ctx, ctx->fingerprint_options, false);
				_fingerprintNode(&fctx, lfirst(lc), parent, field_name, depth + 1);
				lctx->hash = XXH3_64bits_digest(fctx.xxh_state);
				lctx->list_pos = listsort_items_size;
				_fingerprintFreeContext(&fctx);

				listsort_items[listsort_items_size] = lctx;
				listsort_items_size += 1;
			}

			pg_qsort(listsort_items, listsort_items_size, sizeof(FingerprintListsortItem*), compareFingerprintListsortItem);

			FingerprintListsortItemCacheEntry *entry = listsort_cache_insert(ctx->listsort_cache, (uintptr_t) node, &found);
			Assert(!found);

			entry->listsort_items = listsort_items;
			entry->listsort_items_size = listsort_items_size;
		}

		for (size_t i = 0; i < listsort_items_size; i++)
		{
			if (i > 0 && listsort_items[i - 1]->hash == listsort_items[i]->hash)
				continue; // Ignore duplicates

			_fingerprintNode(ctx, lfirst(list_nth_cell(node, listsort_items[i]->list_pos)), parent, field_name, depth + 1);
		}
	}
	else
	{
		const ListCell *lc;

		foreach(lc, node)
		{
			_fingerprintNode(ctx, lfirst(lc), parent, field_name, depth + 1);

			lnext(node, lc);
		}
	}
}

// Note: fingerprint_options only takes effect for root contexts (parent is
// NULL), child contexts always inherit the options of their parent
static void
_fingerprintInitContext(FingerprintContext *ctx, FingerprintContext *parent, int fingerprint_options, bool write_tokens)
{
	ctx->xxh_state = XXH3_createState();
	if (ctx->xxh_state == NULL) abort();
	if (XXH3_64bits_reset_withSeed(ctx->xxh_state, PG_QUERY_FINGERPRINT_VERSION) == XXH_ERROR) abort();

	if (parent != NULL)
	{
		ctx->listsort_cache = parent->listsort_cache;
		ctx->fingerprint_options = parent->fingerprint_options;
	}
	else
	{
		ctx->listsort_cache = listsort_cache_create(CurrentMemoryContext, 128, NULL);
		ctx->fingerprint_options = fingerprint_options;
	}

	if (write_tokens)
	{
		ctx->write_tokens = true;
		dlist_init(&ctx->tokens);
	}
	else
	{
		ctx->write_tokens = false;
	}
}

static void
_fingerprintFreeContext(FingerprintContext *ctx) {
	XXH3_freeState(ctx->xxh_state);
}

#include "pg_query_enum_defs.c"
#include "pg_query_fingerprint_defs.c"

/*
 * Fingerprint the relation name.
 *
 * By default, sequences of 2 or more digits are ignored, so that queries
 * on date/number-suffixed tables (e.g. partitions like "orders_2024_01")
 * get the same fingerprint. With PG_QUERY_FINGERPRINT_FULL_RELNAME set,
 * the relation name is fingerprinted as-is.
 */
static void
_fingerprintRelname(FingerprintContext *ctx, const char *relname)
{
	if (ctx->fingerprint_options & PG_QUERY_FINGERPRINT_FULL_RELNAME)
	{
		_fingerprintString(ctx, "relname");
		_fingerprintString(ctx, relname);
	}
	else
	{
		int len = strlen(relname);
		char *r = palloc0((len + 1) * sizeof(char));
		char *p = r;
		for (int i = 0; i < len; i++)
		{
			if (relname[i] >= '0' && relname[i] <= '9' &&
				((i + 1 < len && relname[i + 1] >= '0' && relname[i + 1] <= '9') ||
				 (i > 0 && relname[i - 1] >= '0' && relname[i - 1] <= '9')))
			{
				// Skip
			}
			else
			{
				*p = relname[i];
				p++;
			}
		}
		*p = 0;
		_fingerprintString(ctx, "relname");
		_fingerprintString(ctx, r);
		pfree(r);
	}
}

/*
 * Fingerprint a RangeVar (relation reference).
 *
 * This is a custom implementation (the generator only emits the declaration
 * and the dispatch case) since the behavior depends on the parent node
 * context and is controlled by fingerprint options.
 *
 * By default, this mirrors Postgres 18+ query jumbling for relation
 * references after parse analysis (RangeTblEntry.eref custom_query_jumble,
 * RangeTblEntry.relid query_jumble_ignore, see Postgres commit 787514b30bb),
 * narrowly: in SELECT/DML contexts we add the user alias name when present,
 * skip the relation name when an alias is present (matches eref.aliasname),
 * and drop the schema name (Postgres jumbles only the eref string;
 * schema-qualified and unqualified references that resolve to the same table
 * share a query ID). catalogname/inh/relpersistence are kept in all contexts
 * to minimize diff against pre-existing fingerprints; their values rarely
 * diverge from defaults in DML parse trees.
 *
 * Two fingerprint options adjust this behavior:
 * PG_QUERY_FINGERPRINT_RANGEVAR_IGNORE_ALIASES always fingerprints the
 * relation name and ignores aliases, and
 * PG_QUERY_FINGERPRINT_RANGEVAR_INCLUDE_SCHEMA also fingerprints schema
 * names in SELECT/DML contexts. Setting both mirrors the Postgres 17 and
 * earlier query jumble, and the fingerprints of earlier libpg_query
 * releases (see PG_QUERY_FINGERPRINT_RANGEVAR_PG17_COMPAT).
 */
static void
_fingerprintRangeVar(FingerprintContext *ctx, const RangeVar *node, const void *parent, const char *field_name, unsigned int depth)
{
	bool ignore_aliases = (ctx->fingerprint_options & PG_QUERY_FINGERPRINT_RANGEVAR_IGNORE_ALIASES) != 0;
	bool include_schema = (ctx->fingerprint_options & PG_QUERY_FINGERPRINT_RANGEVAR_INCLUDE_SCHEMA) != 0;

	bool is_dml_context = false;
	if (parent != NULL && field_name != NULL)
	{
		if (IsA(parent, SelectStmt) && strcmp(field_name, "fromClause") == 0)
			is_dml_context = true;
		else if (IsA(parent, InsertStmt) && strcmp(field_name, "relation") == 0)
			is_dml_context = true;
		else if (IsA(parent, UpdateStmt) && (strcmp(field_name, "relation") == 0 || strcmp(field_name, "fromClause") == 0))
			is_dml_context = true;
		else if (IsA(parent, DeleteStmt) && (strcmp(field_name, "relation") == 0 || strcmp(field_name, "usingClause") == 0))
			is_dml_context = true;
		else if (IsA(parent, MergeStmt) && (strcmp(field_name, "relation") == 0 || strcmp(field_name, "sourceRelation") == 0))
			is_dml_context = true;
		else if (IsA(parent, JoinExpr) || IsA(parent, RangeTableSample) || IsA(parent, LockingClause))
			is_dml_context = true;
	}

	if (!ignore_aliases && node->alias != NULL && node->alias->aliasname != NULL)
	{
		_fingerprintString(ctx, "aliasname");
		_fingerprintString(ctx, node->alias->aliasname);
	}

	if (node->catalogname != NULL)
	{
		_fingerprintString(ctx, "catalogname");
		_fingerprintString(ctx, node->catalogname);
	}

	if (node->inh)
	{
		_fingerprintString(ctx, "inh");
		_fingerprintString(ctx, "true");
	}

	// Intentionally ignoring node->location for fingerprinting

	// In DML/SELECT context, the relation name only contributes to the jumble
	// when there is no user alias (matches eref.aliasname). In utility context,
	// the relation name is always part of the jumble.
	if (node->relname != NULL && node->relpersistence != 't' && !(is_dml_context && !ignore_aliases && node->alias != NULL))
		_fingerprintRelname(ctx, node->relname);

	if (node->relpersistence != 0)
	{
		char buffer[2] = {node->relpersistence, '\0'};
		_fingerprintString(ctx, "relpersistence");
		_fingerprintString(ctx, buffer);
	}

	if (node->schemaname != NULL && (!is_dml_context || include_schema))
	{
		_fingerprintString(ctx, "schemaname");
		_fingerprintString(ctx, node->schemaname);
	}
}

void
_fingerprintNode(FingerprintContext *ctx, const void *obj, const void *parent, char *field_name, unsigned int depth)
{
	// Some queries are overly complex in their parsetree - lets consistently cut them off at 100 nodes deep
	if (depth >= 100) {
		return;
	}

	if (obj == NULL)
	{
		return; // Ignore
	}

	switch (nodeTag(obj))
	{
		case T_List:
			_fingerprintList(ctx, obj, parent, field_name, depth);
			break;
		case T_Integer:
			_fingerprintInteger(ctx, obj);
			break;
		case T_Float:
			_fingerprintFloat(ctx, obj);
			break;
		case T_Boolean:
			_fingerprintBoolean(ctx, obj);
			break;
		case T_String:
			// NB: We output `str` here intentionally, to match the output format from libpg_query
			// 14 and below. This results in stable fingerprints, despite the field name being
			// changed in PG15 to `sval`.
			_fingerprintString(ctx, "String");
			_fingerprintString(ctx, "str");
			_fingerprintString(ctx, ((union ValUnion*) obj)->sval.sval);
			break;
		case T_BitString:
			_fingerprintBitString(ctx, obj);
			break;

		#include "pg_query_fingerprint_conds.c"

		default:
			elog(WARNING, "could not fingerprint unrecognized node type: %d",
					(int) nodeTag(obj));

			return;
	}
}

uint64_t pg_query_fingerprint_node(const void *node)
{
	FingerprintContext ctx;
	uint64 result;

	_fingerprintInitContext(&ctx, NULL, PG_QUERY_FINGERPRINT_DEFAULT, false);
	_fingerprintNode(&ctx, node, NULL, NULL, 0);

	result = XXH3_64bits_digest(ctx.xxh_state);

	_fingerprintFreeContext(&ctx);

	return result;
}

PgQueryFingerprintResult pg_query_fingerprint_with_opts(const char* input, int parser_options, int fingerprint_options, bool printTokens)
{
	MemoryContext ctx = NULL;
	PgQueryInternalParsetreeAndError parsetree_and_error;
	PgQueryFingerprintResult result = {0};

	ctx = pg_query_enter_memory_context();

	parsetree_and_error = pg_query_raw_parse(input, parser_options);

	// These are all malloc-ed and will survive exiting the memory context, the caller is responsible to free them now
	result.stderr_buffer = parsetree_and_error.stderr_buffer;
	result.error = parsetree_and_error.error;

	if (parsetree_and_error.tree != NULL || result.error == NULL) {
		FingerprintContext ctx;
		XXH64_canonical_t chash;

		_fingerprintInitContext(&ctx, NULL, fingerprint_options, printTokens);

		if (parsetree_and_error.tree != NULL) {
			_fingerprintNode(&ctx, parsetree_and_error.tree, NULL, NULL, 0);
		}

		if (printTokens) {
			dlist_iter iter;

			printf("[");

			dlist_foreach(iter, &ctx.tokens)
			{
				FingerprintToken *token = dlist_container(FingerprintToken, list_node, iter.cur);

				printf("\"%s\", ", token->str);
			}

			printf("]\n");
		}

		result.fingerprint = XXH3_64bits_digest(ctx.xxh_state);
		_fingerprintFreeContext(&ctx);

		XXH64_canonicalFromHash(&chash, result.fingerprint);
		result.fingerprint_str = malloc(17 * sizeof(char));
		int n = snprintf(result.fingerprint_str, 17, "%02x%02x%02x%02x%02x%02x%02x%02x",
						   chash.digest[0], chash.digest[1], chash.digest[2], chash.digest[3],
						   chash.digest[4], chash.digest[5], chash.digest[6], chash.digest[7]);
		if (n < 0 || n >= 17) {
			PgQueryError* error = malloc(sizeof(PgQueryError));
			error->message = strdup("Failed to output fingerprint string due to snprintf failure");
			result.error = error;
		}
	}

	pg_query_exit_memory_context(ctx);

	return result;
}

PgQueryFingerprintResult pg_query_fingerprint(const char* input)
{
	return pg_query_fingerprint_with_opts(input, PG_QUERY_PARSE_DEFAULT, PG_QUERY_FINGERPRINT_DEFAULT, false);
}

PgQueryFingerprintResult pg_query_fingerprint_opts(const char* input, int parser_options, int fingerprint_options)
{
	return pg_query_fingerprint_with_opts(input, parser_options, fingerprint_options, false);
}

void pg_query_free_fingerprint_result(PgQueryFingerprintResult result)
{
	if (result.error) {
		free(result.error->message);
		free(result.error->filename);
		free(result.error->funcname);
		free(result.error);
	}

	free(result.fingerprint_str);
	free(result.stderr_buffer);
}
