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

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

const filepath = "/path/to/expected/"  // This should point to the "expected" folder, and end with a forward slash
const outtestpath = "/path/to/output/" // This should point to wherever test files should be saved, and end with a forward slash

func main() {
	names := []string{
		"advisory_lock.out",
		"aggregates.out",
		"alter_generic.out",
		"alter_operator.out",
		"alter_table.out",
		"amutils.out",
		"arrays.out",
		"async.out",
		"bit.out",
		"bitmapops.out",
		"boolean.out",
		"box.out",
		"brin.out",
		"brin_bloom.out",
		"brin_multi.out",
		"btree_index.out",
		"case.out",
		"char.out",
		"char_1.out",
		"char_2.out",
		"circle.out",
		"cluster.out",
		"collate.out",
		"combocid.out",
		"comments.out",
		"compression.out",
		"compression_1.out",
		"constraints.out",
		"conversion.out",
		"copy.out",
		"copy2.out",
		"copydml.out",
		"copyselect.out",
		"create_aggregate.out",
		"create_am.out",
		"create_cast.out",
		"create_function_c.out",
		"create_function_sql.out",
		"create_index.out",
		"create_index_spgist.out",
		"create_misc.out",
		"create_operator.out",
		"create_procedure.out",
		"create_role.out",
		"create_schema.out",
		"create_table.out",
		"create_table_like.out",
		"create_type.out",
		"create_view.out",
		"date.out",
		"dbsize.out",
		"delete.out",
		"dependency.out",
		"domain.out",
		"drop_if_exists.out",
		"drop_operator.out",
		"enum.out",
		"equivclass.out",
		"errors.out",
		"event_trigger.out",
		"explain.out",
		"expressions.out",
		"fast_default.out",
		"float4-misrounded-input.out",
		"float4.out",
		"float8.out",
		"foreign_data.out",
		"foreign_key.out",
		"functional_deps.out",
		"generated.out",
		"geometry.out",
		"gin.out",
		"gist.out",
		"groupingsets.out",
		"guc.out",
		"hash_func.out",
		"hash_index.out",
		"hash_part.out",
		"horology.out",
		"identity.out",
		"incremental_sort.out",
		"index_including.out",
		"index_including_gist.out",
		"indexing.out",
		"indirect_toast.out",
		"inet.out",
		"infinite_recurse.out",
		"infinite_recurse_1.out",
		"inherit.out",
		"init_privs.out",
		"insert.out",
		"insert_conflict.out",
		"int2.out",
		"int4.out",
		"int8.out",
		"interval.out",
		"join.out",
		"join_hash.out",
		"json.out",
		"json_encoding.out",
		"json_encoding_1.out",
		"json_encoding_2.out",
		"jsonb.out",
		"jsonb_jsonpath.out",
		"jsonpath.out",
		"jsonpath_encoding.out",
		"jsonpath_encoding_1.out",
		"jsonpath_encoding_2.out",
		"largeobject.out",
		"largeobject_1.out",
		"limit.out",
		"line.out",
		"lock.out",
		"lseg.out",
		"macaddr.out",
		"macaddr8.out",
		"matview.out",
		"memoize.out",
		"merge.out",
		"misc.out",
		"misc_functions.out",
		"misc_sanity.out",
		"money.out",
		"multirangetypes.out",
		"mvcc.out",
		"name.out",
		"namespace.out",
		"numeric.out",
		"numeric_big.out",
		"numerology.out",
		"object_address.out",
		"oid.out",
		"oidjoins.out",
		"opr_sanity.out",
		"partition_aggregate.out",
		"partition_info.out",
		"partition_join.out",
		"partition_prune.out",
		"password.out",
		"path.out",
		"pg_lsn.out",
		"plancache.out",
		"plpgsql.out",
		"point.out",
		"polygon.out",
		"polymorphism.out",
		"portals.out",
		"portals_p2.out",
		"prepare.out",
		"prepared_xacts.out",
		"prepared_xacts_1.out",
		"privileges.out",
		"psql.out",
		"psql_crosstab.out",
		"publication.out",
		"random.out",
		"rangefuncs.out",
		"rangetypes.out",
		"regex.out",
		"regproc.out",
		"reindex_catalog.out",
		"reloptions.out",
		"replica_identity.out",
		"returning.out",
		"roleattributes.out",
		"rowsecurity.out",
		"rowtypes.out",
		"rules.out",
		"sanity_check.out",
		"security_label.out",
		"select.out",
		"select_distinct.out",
		"select_distinct_on.out",
		"select_having.out",
		"select_having_1.out",
		"select_having_2.out",
		"select_implicit.out",
		"select_implicit_1.out",
		"select_implicit_2.out",
		"select_into.out",
		"select_parallel.out",
		"select_views.out",
		"sequence.out",
		"spgist.out",
		"stats.out",
		"stats_ext.out",
		"strings.out",
		"subscription.out",
		"subselect.out",
		"sysviews.out",
		"tablesample.out",
		"tablespace.out",
		"temp.out",
		//"test_setup.out",
		"text.out",
		"tid.out",
		"tidrangescan.out",
		"tidscan.out",
		"time.out",
		"timestamp.out",
		"timestamptz.out",
		"timetz.out",
		"transactions.out",
		"triggers.out",
		"truncate.out",
		"tsdicts.out",
		"tsearch.out",
		"tsrf.out",
		"tstypes.out",
		"tuplesort.out",
		"txid.out",
		"type_sanity.out",
		"typed_table.out",
		"unicode.out",
		"unicode_1.out",
		"union.out",
		"updatable_views.out",
		"update.out",
		"uuid.out",
		"vacuum.out",
		"vacuum_parallel.out",
		"varchar.out",
		"varchar_1.out",
		"varchar_2.out",
		"window.out",
		"with.out",
		"write_parallel.out",
		"xid.out",
		"xml.out",
		"xml_1.out",
		"xml_2.out",
		"xmlmap.out",
		"xmlmap_1.out",
	}
	for _, n := range names {
		createTest(n, true)
	}
}

func createTest(filename string, write bool) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Error encountered for `%s`: %v\n", filename, err)
		}
	}()
	testFile := NewTestFile(filepath + filename)
	testName := strings.ReplaceAll(filename[:len(filename)-4], "-", "_")
	properTestName := toProperName(testName)

	sb := strings.Builder{}
	sb.WriteString(`// Copyright 2024 Dolthub, Inc.
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

package regression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func Test`)
	sb.WriteString(properTestName)
	sb.WriteString(`(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_`)
	sb.WriteString(testName)
	sb.WriteString(`)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_`)
	sb.WriteString(testName)
	sb.WriteString(`,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements:         []RegressionFileStatement{
`)
	// Parse the file for statements
	for statement, ok := testFile.ReadStatement(); ok; statement, ok = testFile.ReadStatement() {
		sb.WriteString("			{\n")
		sb.WriteString("				Statement: `")
		sb.WriteString(statement)
		sb.WriteString("`,\n")
		if errString, ok := testFile.GetError(); ok {
			sb.WriteString("				ErrorString: `")
			sb.WriteString(errString)
			sb.WriteString("`,\n")
		} else if results, _, ok := testFile.GetResult(); ok {
			sb.WriteString("				Results: ")
			sb.WriteString(results)
			sb.WriteString(",\n")
		}
		sb.WriteString("			},\n")
	}
	sb.WriteString(`		},
	})
}
`)
	if write {
		if err := os.WriteFile(outtestpath+testName+"_test.go", []byte(sb.String()), 0644); err != nil {
			fmt.Print(err)
		}
	} else {
		fmt.Print(sb.String())
	}
}

// toProperName returns a name that should be used appended to the test function's name.
func toProperName(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			runes := []rune(part)
			runes[0] = unicode.ToUpper(runes[0])
			parts[i] = string(runes)
		}
	}
	return strings.Join(parts, "")
}
