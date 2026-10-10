#include "pg_query_outfuncs.h"
#include "pg_query_internal.h"

#include "postgres.h"
#include <ctype.h>
#include "access/relation.h"
#include "nodes/parsenodes.h"
#include "nodes/plannodes.h"
#include "nodes/value.h"
#include "utils/datum.h"
#include "miscadmin.h"

#include "protobuf/pg_query.upb.h"

/*
 * The arena that all messages for the current pg_query_nodes_to_protobuf() call
 * are built in. It is thread-local (libpg_query runs per-thread) and set once at
 * the top of that function; the WRITE_* macros and _out* helpers below read it
 * instead of threading an arena parameter through every function. Keeping it out
 * of the function signatures matters because the generated _out* declarations in
 * pg_query_outfuncs_defs.c are shared with the JSON backend, which has no arena.
 */
static __thread upb_Arena *out_arena;

/*
 * upb stores string fields as a (pointer, length) view without copying, so the
 * bytes must outlive the message. The parse tree is palloc'd in the memory
 * context of the current libpg_query call, which is only exited after the
 * message has been serialized and copied out, so its strings can be referenced
 * directly instead of being duplicated into the arena.
 */
static inline upb_StringView
_strview(const char *s)
{
	return upb_StringView_FromString(s);
}

static upb_StringView
_strview_char(char c)
{
	char	   *buf = upb_Arena_Malloc(out_arena, 1);

	buf[0] = c;
	return upb_StringView_FromDataAndSize(buf, 1);
}

/*
 * The upb message type is named pg_query_<MessageName> (the proto message name,
 * underscores preserved). Every WRITE_ and READ_ macro therefore takes the
 * enclosing message type as its first argument so it can name the accessor
 * pg_query_<MsgType>_set_<field>(). The enclosing message pointer is always
 * `out`; the arena is the file-scope out_arena.
 */
#define OUT_TYPE(typename, typename_c) pg_query_##typename*

/*
 * The Node oneof field for Float is named "float" in the proto. The C++ and
 * protobuf-c backends mangle that C/C++ keyword to "float_" (and the shared
 * OUT_NODE call site passes "float_"), but upb leaves it unmangled as part of
 * the longer accessor token. Alias the mangled name to upb's real accessor so
 * the shared OUT_NODE works for all backends. (Float is the only such case;
 * every other oneof field name matches across backends.)
 */
#define pg_query_Node_mutable_float_ pg_query_Node_mutable_float

/*
 * Similarly, upb appends "_" to a field whose name collides with another
 * field's generated accessor: AlterForeignServerStmt.has_version collides with
 * the presence accessor of its `version` field, so upb names the setter
 * set_has_version_ (C++/JSON keep the plain name). Alias the plain accessor the
 * shared WRITE_BOOL_FIELD emits to upb's real one.
 */
#define pg_query_AlterForeignServerStmt_set_has_version pg_query_AlterForeignServerStmt_set_has_version_

#define OUT_NODE(typename, typename_c, typename_underscore, typename_underscore_upcase, typename_cast, fldname) \
  { \
    pg_query_##typename *__node = pg_query_Node_mutable_##fldname(out, out_arena); \
    _out##typename_c(__node, (const typename_cast *) obj); \
  }

#define WRITE_INT_FIELD(msgtype, outname, outname_json, fldname)    pg_query_##msgtype##_set_##outname(out, node->fldname);
#define WRITE_UINT_FIELD(msgtype, outname, outname_json, fldname)   pg_query_##msgtype##_set_##outname(out, node->fldname);
#define WRITE_UINT64_FIELD(msgtype, outname, outname_json, fldname) pg_query_##msgtype##_set_##outname(out, node->fldname);
#define WRITE_LONG_FIELD(msgtype, outname, outname_json, fldname)   pg_query_##msgtype##_set_##outname(out, node->fldname);
#define WRITE_FLOAT_FIELD(msgtype, outname, outname_json, fldname)  pg_query_##msgtype##_set_##outname(out, node->fldname);
#define WRITE_BOOL_FIELD(msgtype, outname, outname_json, fldname)   pg_query_##msgtype##_set_##outname(out, node->fldname);

#define WRITE_CHAR_FIELD(msgtype, outname, outname_json, fldname) \
	if (node->fldname != 0) { \
		pg_query_##msgtype##_set_##outname(out, _strview_char(node->fldname)); \
	}

#define WRITE_STRING_FIELD(msgtype, outname, outname_json, fldname) \
	if (node->fldname != NULL) { \
		pg_query_##msgtype##_set_##outname(out, _strview(node->fldname)); \
	}

#define WRITE_ENUM_FIELD(msgtype, enumtype, outname, outname_json, fldname) \
	pg_query_##msgtype##_set_##outname(out, _enumToInt##enumtype(node->fldname));

#define WRITE_LIST_FIELD(msgtype, outname, outname_json, fldname) \
	if (node->fldname != NULL) { \
		const ListCell *__lc; \
		foreach(__lc, node->fldname) { \
			pg_query_Node *__n = pg_query_##msgtype##_add_##outname(out, out_arena); \
			_outNode(__n, lfirst(__lc)); \
		} \
	}

#define WRITE_BITMAPSET_FIELD(msgtype, outname, outname_json, fldname) \
	if (!bms_is_empty(node->fldname)) { \
		int __x = -1; \
		while ((__x = bms_next_member(node->fldname, __x)) >= 0) \
			pg_query_##msgtype##_add_##outname(out, (uint64_t) __x, out_arena); \
	}

#define WRITE_NODE_FIELD(msgtype, outname, outname_json, fldname) \
	{ \
		pg_query_Node *__n = pg_query_##msgtype##_mutable_##outname(out, out_arena); \
		_outNode(__n, &node->fldname); \
	}

#define WRITE_NODE_PTR_FIELD(msgtype, outname, outname_json, fldname) \
	if (node->fldname != NULL) { \
		pg_query_Node *__n = pg_query_##msgtype##_mutable_##outname(out, out_arena); \
		_outNode(__n, node->fldname); \
	}

#define WRITE_SPECIFIC_NODE_FIELD(msgtype, typename, typename_underscore, outname, outname_json, fldname) \
	{ \
		pg_query_##typename *__n = pg_query_##msgtype##_mutable_##outname(out, out_arena); \
		_out##typename(__n, &node->fldname); \
	}

/*
 * This recurses into _out##typename directly, bypassing the stack depth check
 * in _outNode, so check here (e.g. a long UNION chain nests SelectStmt in
 * SelectStmt without ever going through _outNode).
 */
#define WRITE_SPECIFIC_NODE_PTR_FIELD(msgtype, typename, typename_underscore, outname, outname_json, fldname) \
	if (node->fldname != NULL) { \
		pg_query_##typename *__n; \
		check_stack_depth(); \
		__n = pg_query_##msgtype##_mutable_##outname(out, out_arena); \
		_out##typename(__n, node->fldname); \
	}

static void _outNode(pg_query_Node* out, const void *obj);

static void
_outList(pg_query_List* out, const List *node)
{
	const ListCell *lc;

	foreach(lc, node)
	{
		pg_query_Node *__n = pg_query_List_add_items(out, out_arena);
		_outNode(__n, lfirst(lc));
	}
}

static void
_outIntList(pg_query_IntList* out, const List *node)
{
	const ListCell *lc;

	foreach(lc, node)
	{
		pg_query_Node *__n = pg_query_IntList_add_items(out, out_arena);
		_outNode(__n, lfirst(lc));
	}
}

static void
_outOidList(pg_query_OidList* out, const List *node)
{
	const ListCell *lc;

	foreach(lc, node)
	{
		pg_query_Node *__n = pg_query_OidList_add_items(out, out_arena);
		_outNode(__n, lfirst(lc));
	}
}

static void
_outInteger(pg_query_Integer* out, const Integer *node)
{
	pg_query_Integer_set_ival(out, node->ival);
}

static void
_outFloat(pg_query_Float* out, const Float *node)
{
	pg_query_Float_set_fval(out, _strview(node->fval));
}

static void
_outBoolean(pg_query_Boolean* out, const Boolean *node)
{
	pg_query_Boolean_set_boolval(out, node->boolval);
}

static void
_outString(pg_query_String* out, const String *node)
{
	pg_query_String_set_sval(out, _strview(node->sval));
}

static void
_outBitString(pg_query_BitString* out, const BitString *node)
{
	pg_query_BitString_set_bsval(out, _strview(node->bsval));
}

static void
_outAConst(pg_query_A_Const* out, const A_Const *node)
{
	pg_query_A_Const_set_isnull(out, node->isnull);
	pg_query_A_Const_set_location(out, node->location);

	if (!node->isnull) {
		switch (nodeTag(&node->val.node)) {
			case T_Integer:
				pg_query_Integer_set_ival(pg_query_A_Const_mutable_ival(out, out_arena), node->val.ival.ival);
				break;
			case T_Float:
				pg_query_Float_set_fval(pg_query_A_Const_mutable_fval(out, out_arena), _strview(node->val.fval.fval));
				break;
			case T_Boolean:
				pg_query_Boolean_set_boolval(pg_query_A_Const_mutable_boolval(out, out_arena), node->val.boolval.boolval);
				break;
			case T_String:
				pg_query_String_set_sval(pg_query_A_Const_mutable_sval(out, out_arena), _strview(node->val.sval.sval));
				break;
			case T_BitString:
				pg_query_BitString_set_bsval(pg_query_A_Const_mutable_bsval(out, out_arena), _strview(node->val.bsval.bsval));
				break;
			default:
				// Unreachable
				Assert(false);
		}
	}
}

#include "pg_query_enum_defs.c"
#include "pg_query_outfuncs_defs.c"

static void
_outNode(pg_query_Node* out, const void *obj)
{
	check_stack_depth();

	if (obj == NULL)
		return; // Keep out as NULL

	switch (nodeTag(obj))
	{
		#include "pg_query_outfuncs_conds.c"

		default:
			elog(WARNING, "could not dump unrecognized node type: %d",
					(int) nodeTag(obj));

			return;
	}
}

PgQueryProtobuf
pg_query_nodes_to_protobuf(const void *obj)
{
	PgQueryProtobuf protobuf;
	pg_query_ParseResult *parse_result;
	char	   *data;
	size_t		len;

	out_arena = upb_Arena_New();
	parse_result = pg_query_ParseResult_new(out_arena);

	pg_query_ParseResult_set_version(parse_result, PG_VERSION_NUM);

	/*
	 * Building the message may throw (e.g. "stack depth limit exceeded" from
	 * _outNode), and the arena is malloc-backed, so it needs to be freed here
	 * before passing on the error.
	 */
	PG_TRY();
	{
		if (obj != NULL)
		{
			const ListCell *lc;

			foreach(lc, (const List *) obj)
			{
				pg_query_RawStmt *stmt = pg_query_ParseResult_add_stmts(parse_result, out_arena);
				_outRawStmt(stmt, lfirst(lc));
			}
		}
	}
	PG_CATCH();
	{
		upb_Arena_Free(out_arena);
		out_arena = NULL;
		PG_RE_THROW();
	}
	PG_END_TRY();

	/*
	 * Encode with the same depth limit the decoder uses, so anything we emit
	 * can be read back. Serialization fails (returns NULL) for deeper trees, or
	 * when the stack depth limit is reached while encoding.
	 *
	 * Note: upb also returns NULL when the arena runs out of memory
	 * (kUpb_EncodeStatus_OutOfMemory), which we report with the same message
	 * since the generated serialize_ex wrapper discards the encode status. That
	 * is unlikely enough in practice that we don't call upb_Encode directly to
	 * tell the two apart.
	 */
	data = pg_query_ParseResult_serialize_ex(parse_result, upb_EncodeOptions_MaxDepth(PG_QUERY_PROTOBUF_MAX_DEPTH), out_arena, &len);
	if (data == NULL)
	{
		upb_Arena_Free(out_arena);
		out_arena = NULL;
		elog(ERROR, "parse tree is nested too deeply to serialize to protobuf");
	}

	protobuf.len = len;
	// Note: This is intentionally malloc so exiting the memory context doesn't free this
	protobuf.data = malloc(len);
	memcpy(protobuf.data, data, len);

	upb_Arena_Free(out_arena);
	out_arena = NULL;

	return protobuf;
}
