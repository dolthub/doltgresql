// Copyright 2024 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

// AllTestResultFilesNames contains the names of all of the test result files that should be created.
var AllTestResultFilesNames = []string{
	"results/test_setup.results",
	"results/tablespace.results",
	"results/boolean.results",
	"results/char.results",
	"results/name.results",
	"results/varchar.results",
	"results/text.results",
	"results/int2.results",
	"results/int4.results",
	"results/int8.results",
	"results/oid.results",
	"results/float4.results",
	"results/float8.results",
	"results/bit.results",
	"results/numeric.results",
	"results/txid.results",
	"results/uuid.results",
	"results/enum.results",
	"results/money.results",
	"results/rangetypes.results",
	"results/pg_lsn.results",
	"results/regproc.results",
	"results/strings.results",
	"results/numerology.results",
	"results/point.results",
	"results/lseg.results",
	"results/line.results",
	"results/box.results",
	"results/path.results",
	"results/polygon.results",
	"results/circle.results",
	"results/date.results",
	"results/time.results",
	"results/timetz.results",
	"results/timestamp.results",
	"results/timestamptz.results",
	"results/interval.results",
	"results/inet.results",
	"results/macaddr.results",
	"results/macaddr8.results",
	"results/multirangetypes.results",
	"results/geometry.results",
	"results/horology.results",
	"results/tstypes.results",
	"results/regex.results",
	"results/type_sanity.results",
	"results/opr_sanity.results",
	"results/misc_sanity.results",
	"results/comments.results",
	"results/expressions.results",
	"results/unicode.results",
	"results/xid.results",
	"results/mvcc.results",
	"results/copy.results",
	"results/copyselect.results",
	"results/copydml.results",
	"results/insert.results",
	"results/insert_conflict.results",
	"results/create_misc.results",
	"results/create_operator.results",
	"results/create_procedure.results",
	"results/create_table.results",
	"results/create_type.results",
	"results/create_schema.results",
	"results/create_index.results",
	"results/create_index_spgist.results",
	"results/create_view.results",
	"results/index_including.results",
	"results/index_including_gist.results",
	"results/create_aggregate.results",
	"results/create_function_sql.results",
	"results/create_cast.results",
	"results/constraints.results",
	"results/triggers.results",
	"results/select.results",
	"results/inherit.results",
	"results/typed_table.results",
	"results/vacuum.results",
	"results/drop_if_exists.results",
	"results/updatable_views.results",
	"results/roleattributes.results",
	"results/create_am.results",
	"results/hash_func.results",
	"results/errors.results",
	"results/infinite_recurse.results",
	"results/sanity_check.results",
	"results/select_into.results",
	"results/select_distinct.results",
	"results/select_distinct_on.results",
	"results/select_implicit.results",
	"results/select_having.results",
	"results/subselect.results",
	"results/union.results",
	"results/case.results",
	"results/join.results",
	"results/aggregates.results",
	"results/transactions.results",
	"results/random.results",
	"results/portals.results",
	"results/arrays.results",
	"results/btree_index.results",
	"results/hash_index.results",
	"results/update.results",
	"results/delete.results",
	"results/namespace.results",
	"results/brin.results",
	"results/gin.results",
	"results/gist.results",
	"results/spgist.results",
	"results/init_privs.results",
	"results/security_label.results",
	"results/collate.results",
	"results/matview.results",
	"results/lock.results",
	"results/replica_identity.results",
	"results/rowsecurity.results",
	"results/object_address.results",
	"results/tablesample.results",
	"results/groupingsets.results",
	"results/drop_operator.results",
	"results/password.results",
	"results/identity.results",
	"results/generated.results",
	"results/join_hash.results",
	"results/brin_bloom.results",
	"results/brin_multi.results",
	"results/create_table_like.results",
	"results/alter_generic.results",
	"results/alter_operator.results",
	"results/misc.results",
	"results/async.results",
	"results/dbsize.results",
	"results/merge.results",
	"results/misc_functions.results",
	"results/sysviews.results",
	"results/tsrf.results",
	"results/tid.results",
	"results/tidscan.results",
	"results/tidrangescan.results",
	"results/collate.icu.utf8.results",
	"results/incremental_sort.results",
	"results/create_role.results",
	"results/rules.results",
	"results/psql.results",
	"results/psql_crosstab.results",
	"results/amutils.results",
	"results/stats_ext.results",
	"results/select_parallel.results",
	"results/write_parallel.results",
	"results/vacuum_parallel.results",
	"results/publication.results",
	"results/subscription.results",
	"results/select_views.results",
	"results/portals_p2.results",
	"results/foreign_key.results",
	"results/cluster.results",
	"results/dependency.results",
	"results/guc.results",
	"results/bitmapops.results",
	"results/combocid.results",
	"results/tsearch.results",
	"results/tsdicts.results",
	"results/window.results",
	"results/xmlmap.results",
	"results/functional_deps.results",
	"results/advisory_lock.results",
	"results/indirect_toast.results",
	"results/equivclass.results",
	"results/json.results",
	"results/jsonb.results",
	"results/json_encoding.results",
	"results/jsonpath.results",
	"results/jsonpath_encoding.results",
	"results/jsonb_jsonpath.results",
	"results/plancache.results",
	"results/limit.results",
	"results/plpgsql.results",
	"results/copy2.results",
	"results/domain.results",
	"results/rangefuncs.results",
	"results/prepare.results",
	"results/conversion.results",
	"results/truncate.results",
	"results/alter_table.results",
	"results/sequence.results",
	"results/polymorphism.results",
	"results/rowtypes.results",
	"results/returning.results",
	"results/largeobject.results",
	"results/with.results",
	"results/xml.results",
	"results/partition_join.results",
	"results/partition_prune.results",
	"results/reloptions.results",
	"results/hash_part.results",
	"results/indexing.results",
	"results/partition_aggregate.results",
	"results/partition_info.results",
	"results/tuplesort.results",
	"results/explain.results",
	"results/compression.results",
	"results/memoize.results",
	"results/event_trigger.results",
	"results/oidjoins.results",
	"results/fast_default.results",
}