#include "pg_query_readfuncs.h"
#include "pg_query_internal.h"

#include "postgres.h"
#include "miscadmin.h"
#include "nodes/nodes.h"
#include "nodes/parsenodes.h"
#include "nodes/pg_list.h"
#include "nodes/value.h"

#include "protobuf/pg_query.upb.h"

#define OUT_TYPE(typename, typename_c) pg_query_##typename*

/*
 * upb appends "_" to a field whose name collides with another field's accessor:
 * AlterForeignServerStmt.has_version collides with the presence accessor of its
 * `version` field, so upb names the getter has_version_. Alias the plain
 * accessor the shared READ_BOOL_FIELD emits to upb's real one. (See the
 * matching setter alias in pg_query_outfuncs_protobuf.c.)
 */
#define pg_query_AlterForeignServerStmt_has_version pg_query_AlterForeignServerStmt_has_version_

/* upb returns string fields as a (pointer, length) view into the parse arena;
 * copy them into palloc'd, NUL-terminated C strings for the Postgres nodes. */
static char *
_strview_to_cstring(upb_StringView sv)
{
	char	   *s = palloc(sv.size + 1);

	memcpy(s, sv.data, sv.size);
	s[sv.size] = '\0';
	return s;
}

/*
 * As in the outfuncs, every READ_ and READ_COND macro takes the enclosing upb
 * message type first so it can name the accessor pg_query_<MsgType>_<field>().
 * The message pointer is always `msg`.
 */
#define READ_COND(typename, typename_c, typename_underscore, typename_underscore_upcase, typename_cast, outname) \
	case pg_query_Node_node_##outname: \
		return (Node *) _read##typename_c(pg_query_Node_##outname(msg));

#define READ_INT_FIELD(msgtype, outname, outname_json, fldname)    node->fldname = pg_query_##msgtype##_##outname(msg);
#define READ_UINT_FIELD(msgtype, outname, outname_json, fldname)   node->fldname = pg_query_##msgtype##_##outname(msg);
#define READ_UINT64_FIELD(msgtype, outname, outname_json, fldname) node->fldname = pg_query_##msgtype##_##outname(msg);
#define READ_LONG_FIELD(msgtype, outname, outname_json, fldname)   node->fldname = pg_query_##msgtype##_##outname(msg);
#define READ_FLOAT_FIELD(msgtype, outname, outname_json, fldname)  node->fldname = pg_query_##msgtype##_##outname(msg);
#define READ_BOOL_FIELD(msgtype, outname, outname_json, fldname)   node->fldname = pg_query_##msgtype##_##outname(msg);

#define READ_CHAR_FIELD(msgtype, outname, outname_json, fldname) \
	{ \
		upb_StringView __sv = pg_query_##msgtype##_##outname(msg); \
		if (__sv.size > 0) \
			node->fldname = __sv.data[0]; \
	}

#define READ_STRING_FIELD(msgtype, outname, outname_json, fldname) \
	{ \
		upb_StringView __sv = pg_query_##msgtype##_##outname(msg); \
		if (__sv.size > 0) \
			node->fldname = _strview_to_cstring(__sv); \
	}

#define READ_ENUM_FIELD(msgtype, enumtype, outname, outname_json, fldname) \
	node->fldname = _intToEnum##enumtype(pg_query_##msgtype##_##outname(msg));

#define READ_LIST_FIELD(msgtype, outname, outname_json, fldname) \
	{ \
		size_t __n; \
		const pg_query_Node *const *__items = pg_query_##msgtype##_##outname(msg, &__n); \
		for (size_t __i = 0; __i < __n; __i++) \
			node->fldname = lappend(node->fldname, _readNode(__items[__i])); \
	}

#define READ_BITMAPSET_FIELD(msgtype, outname, outname_json, fldname) // FIXME

#define READ_NODE_FIELD(msgtype, outname, outname_json, fldname) \
	if (pg_query_##msgtype##_has_##outname(msg)) \
		node->fldname = *_readNode(pg_query_##msgtype##_##outname(msg));

#define READ_NODE_PTR_FIELD(msgtype, outname, outname_json, fldname) \
	if (pg_query_##msgtype##_has_##outname(msg)) \
		node->fldname = _readNode(pg_query_##msgtype##_##outname(msg));

#define READ_ABSTRACT_PTR_FIELD(msgtype, outname, outname_json, fldname, fldtype) \
	if (pg_query_##msgtype##_has_##outname(msg)) \
		node->fldname = (fldtype) _readNode(pg_query_##msgtype##_##outname(msg));

#define READ_VALUE_FIELD(msgtype, outname, outname_json, fldname) \
	if (pg_query_##msgtype##_has_##outname(msg)) \
		node->fldname = *((Value *) _readNode(pg_query_##msgtype##_##outname(msg)));

#define READ_VALUE_PTR_FIELD(msgtype, outname, outname_json, fldname) \
	if (pg_query_##msgtype##_has_##outname(msg)) \
		node->fldname = (Value *) _readNode(pg_query_##msgtype##_##outname(msg));

#define READ_SPECIFIC_NODE_FIELD(msgtype, typename, typename_underscore, outname, outname_json, fldname) \
	if (pg_query_##msgtype##_has_##outname(msg)) \
		node->fldname = *_read##typename(pg_query_##msgtype##_##outname(msg));

/*
 * This recurses into _read##typename directly, bypassing the stack depth check
 * in _readNode, so check here (e.g. a long UNION chain nests SelectStmt in
 * SelectStmt without ever going through _readNode).
 */
#define READ_SPECIFIC_NODE_PTR_FIELD(msgtype, typename, typename_underscore, outname, outname_json, fldname) \
	if (pg_query_##msgtype##_has_##outname(msg)) { \
		check_stack_depth(); \
		node->fldname = _read##typename(pg_query_##msgtype##_##outname(msg)); \
	}

static Node * _readNode(const pg_query_Node *msg);

static String *
_readString(const pg_query_String* msg)
{
	return makeString(_strview_to_cstring(pg_query_String_sval(msg)));
}

#include "pg_query_enum_defs.c"
#include "pg_query_readfuncs_defs.c"

static List *
_readList(const pg_query_List *msg)
{
	List	   *node = NULL;
	size_t		n;
	const pg_query_Node *const *items = pg_query_List_items(msg, &n);

	for (size_t i = 0; i < n; i++)
		node = lappend(node, _readNode(items[i]));
	return node;
}

static Node *
_readNode(const pg_query_Node *msg)
{
	check_stack_depth();

	switch (pg_query_Node_node_case(msg))
	{
		#include "pg_query_readfuncs_conds.c"

		case pg_query_Node_node_integer:
			return (Node *) makeInteger(pg_query_Integer_ival(pg_query_Node_integer(msg)));
		case pg_query_Node_node_float:
			return (Node *) makeFloat(_strview_to_cstring(pg_query_Float_fval(pg_query_Node_float(msg))));
		case pg_query_Node_node_boolean:
			return (Node *) makeBoolean(pg_query_Boolean_boolval(pg_query_Node_boolean(msg)));
		case pg_query_Node_node_string:
			return (Node *) makeString(_strview_to_cstring(pg_query_String_sval(pg_query_Node_string(msg))));
		case pg_query_Node_node_bit_string:
			return (Node *) makeBitString(_strview_to_cstring(pg_query_BitString_bsval(pg_query_Node_bit_string(msg))));
		case pg_query_Node_node_a_const: {
			const pg_query_A_Const *ac_msg = pg_query_Node_a_const(msg);
			A_Const *ac = makeNode(A_Const);
			ac->location = pg_query_A_Const_location(ac_msg);

			if (pg_query_A_Const_isnull(ac_msg)) {
				ac->isnull = true;
			} else {
				switch (pg_query_A_Const_val_case(ac_msg)) {
					case pg_query_A_Const_val_ival:
						ac->val.ival = *makeInteger(pg_query_Integer_ival(pg_query_A_Const_ival(ac_msg)));
						break;
					case pg_query_A_Const_val_fval:
						ac->val.fval = *makeFloat(_strview_to_cstring(pg_query_Float_fval(pg_query_A_Const_fval(ac_msg))));
						break;
					case pg_query_A_Const_val_boolval:
						ac->val.boolval = *makeBoolean(pg_query_Boolean_boolval(pg_query_A_Const_boolval(ac_msg)));
						break;
					case pg_query_A_Const_val_sval:
						ac->val.sval = *makeString(_strview_to_cstring(pg_query_String_sval(pg_query_A_Const_sval(ac_msg))));
						break;
					case pg_query_A_Const_val_bsval:
						ac->val.bsval = *makeBitString(_strview_to_cstring(pg_query_BitString_bsval(pg_query_A_Const_bsval(ac_msg))));
						break;
					case pg_query_A_Const_val_NOT_SET:
						Assert(false);
						break;
				}
			}

			return (Node *) ac;
		}
		case pg_query_Node_node_list:
			return (Node *) _readList(pg_query_Node_list(msg));
		case pg_query_Node_node_NOT_SET:
			return NULL;
		default:
			elog(ERROR, "unsupported protobuf node type: %d",
				 (int) pg_query_Node_node_case(msg));
	}
}

List * pg_query_protobuf_to_nodes(PgQueryProtobuf protobuf)
{
	upb_Arena  *arena = upb_Arena_New();
	pg_query_ParseResult *result;
	List	   *list = NULL;
	size_t		n_stmts = 0;
	const pg_query_RawStmt *const *stmts;
	size_t		i;

	/*
	 * upb's default depth limit (100) is far too low for real queries, use the
	 * same limit as the encoder so we accept any tree we produce ourselves.
	 * Decoding also fails once the stack depth limit is reached (see
	 * vendor/upb/patches/0003-stack-depth-check.patch).
	 */
	result = pg_query_ParseResult_parse_ex(protobuf.data, protobuf.len, NULL,
										   upb_DecodeOptions_MaxDepth(PG_QUERY_PROTOBUF_MAX_DEPTH),
										   arena);

	if (result == NULL)
	{
		upb_Arena_Free(arena);
		elog(ERROR, "could not parse protobuf");
	}

	/*
	 * Note: The version field is not checked. The protobuf format is only
	 * expected to be read by the same libpg_query version, and mismatches show
	 * up as unknown fields (ignored) or missing ones (zero values).
	 */

	/*
	 * Reading the nodes may throw (e.g. "stack depth limit exceeded" from
	 * _readNode), and the arena is malloc-backed, so it needs to be freed here
	 * before passing on the error.
	 */
	PG_TRY();
	{
		stmts = pg_query_ParseResult_stmts(result, &n_stmts);
		for (i = 0; i < n_stmts; i++)
			list = lappend(list, _readRawStmt(stmts[i]));
	}
	PG_CATCH();
	{
		upb_Arena_Free(arena);
		PG_RE_THROW();
	}
	PG_END_TRY();

	upb_Arena_Free(arena);

	return list;
}
