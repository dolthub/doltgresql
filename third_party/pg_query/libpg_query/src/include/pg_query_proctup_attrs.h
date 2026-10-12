/*
 * Shared definition of the pg_proc tuple wrapper used by libpg_query when
 * driving plpgsql_compile_callback from outside a real backend.
 *
 * The libpg_query path forges a pg_proc HeapTuple plus Anum-indexed
 * values[]/nulls[] arrays (mirroring upstream ProcedureCreate). We embed
 * the arrays in this wrapper alongside the HeapTupleData so the mocks for
 * SysCacheGetAttr / SysCacheGetAttrNotNull can recover them from the
 * HeapTuple pointer via container_of, without any thread-local state.
 */

#ifndef PG_QUERY_PROCTUP_ATTRS_H
#define PG_QUERY_PROCTUP_ATTRS_H

#include "postgres.h"
#include "access/htup.h"
#include "catalog/pg_proc.h"

typedef struct {
	HeapTupleData tup;
	Datum	values[Natts_pg_proc];
	bool	nulls[Natts_pg_proc];
} ProcTupWithAttrs;

#endif							/* PG_QUERY_PROCTUP_ATTRS_H */
