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

package regression

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"testing"
	"time"

	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dserver "github.com/dolthub/doltgresql/server"
	"github.com/dolthub/doltgresql/servercfg"
	"github.com/dolthub/doltgresql/testing/generation/utils"
)

// RegressionFileName represents
type RegressionFileName uint8

const (
	RegressionFileName_invalid                 RegressionFileName = iota // Causes an error, since it indicates that the name was not set
	RegressionFileName_advisory_lock                                     // advisory_lock
	RegressionFileName_aggregates                                        // aggregates
	RegressionFileName_alter_generic                                     // alter_generic
	RegressionFileName_alter_operator                                    // alter_operator
	RegressionFileName_alter_table                                       // alter_table
	RegressionFileName_amutils                                           // amutils
	RegressionFileName_arrays                                            // arrays
	RegressionFileName_async                                             // async
	RegressionFileName_bit                                               // bit
	RegressionFileName_bitmapops                                         // bitmapops
	RegressionFileName_boolean                                           // boolean
	RegressionFileName_box                                               // box
	RegressionFileName_brin                                              // brin
	RegressionFileName_brin_bloom                                        // brin_bloom
	RegressionFileName_brin_multi                                        // brin_multi
	RegressionFileName_btree_index                                       // btree_index
	RegressionFileName_case                                              // case
	RegressionFileName_char                                              // char
	RegressionFileName_char_1                                            // char_1
	RegressionFileName_char_2                                            // char_2
	RegressionFileName_circle                                            // circle
	RegressionFileName_cluster                                           // cluster
	RegressionFileName_collate                                           // collate
	RegressionFileName_combocid                                          // combocid
	RegressionFileName_comments                                          // comments
	RegressionFileName_compression                                       // compression
	RegressionFileName_compression_1                                     // compression_1
	RegressionFileName_constraints                                       // constraints
	RegressionFileName_conversion                                        // conversion
	RegressionFileName_copy                                              // copy
	RegressionFileName_copy2                                             // copy2
	RegressionFileName_copydml                                           // copydml
	RegressionFileName_copyselect                                        // copyselect
	RegressionFileName_create_aggregate                                  // create_aggregate
	RegressionFileName_create_am                                         // create_am
	RegressionFileName_create_cast                                       // create_cast
	RegressionFileName_create_function_c                                 // create_function_c
	RegressionFileName_create_function_sql                               // create_function_sql
	RegressionFileName_create_index                                      // create_index
	RegressionFileName_create_index_spgist                               // create_index_spgist
	RegressionFileName_create_misc                                       // create_misc
	RegressionFileName_create_operator                                   // create_operator
	RegressionFileName_create_procedure                                  // create_procedure
	RegressionFileName_create_role                                       // create_role
	RegressionFileName_create_schema                                     // create_schema
	RegressionFileName_create_table                                      // create_table
	RegressionFileName_create_table_like                                 // create_table_like
	RegressionFileName_create_type                                       // create_type
	RegressionFileName_create_view                                       // create_view
	RegressionFileName_date                                              // date
	RegressionFileName_dbsize                                            // dbsize
	RegressionFileName_delete                                            // delete
	RegressionFileName_dependency                                        // dependency
	RegressionFileName_domain                                            // domain
	RegressionFileName_drop_if_exists                                    // drop_if_exists
	RegressionFileName_drop_operator                                     // drop_operator
	RegressionFileName_enum                                              // enum
	RegressionFileName_equivclass                                        // equivclass
	RegressionFileName_errors                                            // errors
	RegressionFileName_event_trigger                                     // event_trigger
	RegressionFileName_explain                                           // explain
	RegressionFileName_expressions                                       // expressions
	RegressionFileName_fast_default                                      // fast_default
	RegressionFileName_float4_misrounded_input                           // float4-misrounded-input
	RegressionFileName_float4                                            // float4
	RegressionFileName_float8                                            // float8
	RegressionFileName_foreign_data                                      // foreign_data
	RegressionFileName_foreign_key                                       // foreign_key
	RegressionFileName_functional_deps                                   // functional_deps
	RegressionFileName_generated                                         // generated
	RegressionFileName_geometry                                          // geometry
	RegressionFileName_gin                                               // gin
	RegressionFileName_gist                                              // gist
	RegressionFileName_groupingsets                                      // groupingsets
	RegressionFileName_guc                                               // guc
	RegressionFileName_hash_func                                         // hash_func
	RegressionFileName_hash_index                                        // hash_index
	RegressionFileName_hash_part                                         // hash_part
	RegressionFileName_horology                                          // horology
	RegressionFileName_identity                                          // identity
	RegressionFileName_incremental_sort                                  // incremental_sort
	RegressionFileName_index_including                                   // index_including
	RegressionFileName_index_including_gist                              // index_including_gist
	RegressionFileName_indexing                                          // indexing
	RegressionFileName_indirect_toast                                    // indirect_toast
	RegressionFileName_inet                                              // inet
	RegressionFileName_infinite_recurse                                  // infinite_recurse
	RegressionFileName_infinite_recurse_1                                // infinite_recurse_1
	RegressionFileName_inherit                                           // inherit
	RegressionFileName_init_privs                                        // init_privs
	RegressionFileName_insert                                            // insert
	RegressionFileName_insert_conflict                                   // insert_conflict
	RegressionFileName_int2                                              // int2
	RegressionFileName_int4                                              // int4
	RegressionFileName_int8                                              // int8
	RegressionFileName_interval                                          // interval
	RegressionFileName_join                                              // join
	RegressionFileName_join_hash                                         // join_hash
	RegressionFileName_json                                              // json
	RegressionFileName_json_encoding                                     // json_encoding
	RegressionFileName_json_encoding_1                                   // json_encoding_1
	RegressionFileName_json_encoding_2                                   // json_encoding_2
	RegressionFileName_jsonb                                             // jsonb
	RegressionFileName_jsonb_jsonpath                                    // jsonb_jsonpath
	RegressionFileName_jsonpath                                          // jsonpath
	RegressionFileName_jsonpath_encoding                                 // jsonpath_encoding
	RegressionFileName_jsonpath_encoding_1                               // jsonpath_encoding_1
	RegressionFileName_jsonpath_encoding_2                               // jsonpath_encoding_2
	RegressionFileName_largeobject                                       // largeobject
	RegressionFileName_largeobject_1                                     // largeobject_1
	RegressionFileName_limit                                             // limit
	RegressionFileName_line                                              // line
	RegressionFileName_lock                                              // lock
	RegressionFileName_lseg                                              // lseg
	RegressionFileName_macaddr                                           // macaddr
	RegressionFileName_macaddr8                                          // macaddr8
	RegressionFileName_matview                                           // matview
	RegressionFileName_memoize                                           // memoize
	RegressionFileName_merge                                             // merge
	RegressionFileName_misc                                              // misc
	RegressionFileName_misc_functions                                    // misc_functions
	RegressionFileName_misc_sanity                                       // misc_sanity
	RegressionFileName_money                                             // money
	RegressionFileName_multirangetypes                                   // multirangetypes
	RegressionFileName_mvcc                                              // mvcc
	RegressionFileName_name                                              // name
	RegressionFileName_namespace                                         // namespace
	RegressionFileName_numeric                                           // numeric
	RegressionFileName_numeric_big                                       // numeric_big
	RegressionFileName_numerology                                        // numerology
	RegressionFileName_object_address                                    // object_address
	RegressionFileName_oid                                               // oid
	RegressionFileName_oidjoins                                          // oidjoins
	RegressionFileName_opr_sanity                                        // opr_sanity
	RegressionFileName_partition_aggregate                               // partition_aggregate
	RegressionFileName_partition_info                                    // partition_info
	RegressionFileName_partition_join                                    // partition_join
	RegressionFileName_partition_prune                                   // partition_prune
	RegressionFileName_password                                          // password
	RegressionFileName_path                                              // path
	RegressionFileName_pg_lsn                                            // pg_lsn
	RegressionFileName_plancache                                         // plancache
	RegressionFileName_plpgsql                                           // plpgsql
	RegressionFileName_point                                             // point
	RegressionFileName_polygon                                           // polygon
	RegressionFileName_polymorphism                                      // polymorphism
	RegressionFileName_portals                                           // portals
	RegressionFileName_portals_p2                                        // portals_p2
	RegressionFileName_prepare                                           // prepare
	RegressionFileName_prepared_xacts                                    // prepared_xacts
	RegressionFileName_prepared_xacts_1                                  // prepared_xacts_1
	RegressionFileName_privileges                                        // privileges
	RegressionFileName_psql                                              // psql
	RegressionFileName_psql_crosstab                                     // psql_crosstab
	RegressionFileName_publication                                       // publication
	RegressionFileName_random                                            // random
	RegressionFileName_rangefuncs                                        // rangefuncs
	RegressionFileName_rangetypes                                        // rangetypes
	RegressionFileName_regex                                             // regex
	RegressionFileName_regproc                                           // regproc
	RegressionFileName_reindex_catalog                                   // reindex_catalog
	RegressionFileName_reloptions                                        // reloptions
	RegressionFileName_replica_identity                                  // replica_identity
	RegressionFileName_returning                                         // returning
	RegressionFileName_roleattributes                                    // roleattributes
	RegressionFileName_rowsecurity                                       // rowsecurity
	RegressionFileName_rowtypes                                          // rowtypes
	RegressionFileName_rules                                             // rules
	RegressionFileName_sanity_check                                      // sanity_check
	RegressionFileName_security_label                                    // security_label
	RegressionFileName_select                                            // select
	RegressionFileName_select_distinct                                   // select_distinct
	RegressionFileName_select_distinct_on                                // select_distinct_on
	RegressionFileName_select_having                                     // select_having
	RegressionFileName_select_having_1                                   // select_having_1
	RegressionFileName_select_having_2                                   // select_having_2
	RegressionFileName_select_implicit                                   // select_implicit
	RegressionFileName_select_implicit_1                                 // select_implicit_1
	RegressionFileName_select_implicit_2                                 // select_implicit_2
	RegressionFileName_select_into                                       // select_into
	RegressionFileName_select_parallel                                   // select_parallel
	RegressionFileName_select_views                                      // select_views
	RegressionFileName_sequence                                          // sequence
	RegressionFileName_spgist                                            // spgist
	RegressionFileName_stats                                             // stats
	RegressionFileName_stats_ext                                         // stats_ext
	RegressionFileName_strings                                           // strings
	RegressionFileName_subscription                                      // subscription
	RegressionFileName_subselect                                         // subselect
	RegressionFileName_sysviews                                          // sysviews
	RegressionFileName_tablesample                                       // tablesample
	RegressionFileName_tablespace                                        // tablespace
	RegressionFileName_temp                                              // temp
	RegressionFileName_test_setup                                        // test_setup
	RegressionFileName_text                                              // text
	RegressionFileName_tid                                               // tid
	RegressionFileName_tidrangescan                                      // tidrangescan
	RegressionFileName_tidscan                                           // tidscan
	RegressionFileName_time                                              // time
	RegressionFileName_timestamp                                         // timestamp
	RegressionFileName_timestamptz                                       // timestamptz
	RegressionFileName_timetz                                            // timetz
	RegressionFileName_transactions                                      // transactions
	RegressionFileName_triggers                                          // triggers
	RegressionFileName_truncate                                          // truncate
	RegressionFileName_tsdicts                                           // tsdicts
	RegressionFileName_tsearch                                           // tsearch
	RegressionFileName_tsrf                                              // tsrf
	RegressionFileName_tstypes                                           // tstypes
	RegressionFileName_tuplesort                                         // tuplesort
	RegressionFileName_txid                                              // txid
	RegressionFileName_type_sanity                                       // type_sanity
	RegressionFileName_typed_table                                       // typed_table
	RegressionFileName_unicode                                           // unicode
	RegressionFileName_unicode_1                                         // unicode_1
	RegressionFileName_union                                             // union
	RegressionFileName_updatable_views                                   // updatable_views
	RegressionFileName_update                                            // update
	RegressionFileName_uuid                                              // uuid
	RegressionFileName_vacuum                                            // vacuum
	RegressionFileName_vacuum_parallel                                   // vacuum_parallel
	RegressionFileName_varchar                                           // varchar
	RegressionFileName_varchar_1                                         // varchar_1
	RegressionFileName_varchar_2                                         // varchar_2
	RegressionFileName_window                                            // window
	RegressionFileName_with                                              // with
	RegressionFileName_write_parallel                                    // write_parallel
	RegressionFileName_xid                                               // xid
	RegressionFileName_xml                                               // xml
	RegressionFileName_xml_1                                             // xml_1
	RegressionFileName_xml_2                                             // xml_2
	RegressionFileName_xmlmap                                            // xmlmap
	RegressionFileName_xmlmap_1                                          // xmlmap_1
)

// regressionFileNames is a slice that maps from a RegressionFileName to its string
var regressionFileNames = []string{
	"invalid",
	"advisory_lock",
	"aggregates",
	"alter_generic",
	"alter_operator",
	"alter_table",
	"amutils",
	"arrays",
	"async",
	"bit",
	"bitmapops",
	"boolean",
	"box",
	"brin",
	"brin_bloom",
	"brin_multi",
	"btree_index",
	"case",
	"char",
	"char_1",
	"char_2",
	"circle",
	"cluster",
	"collate",
	"combocid",
	"comments",
	"compression",
	"compression_1",
	"constraints",
	"conversion",
	"copy",
	"copy2",
	"copydml",
	"copyselect",
	"create_aggregate",
	"create_am",
	"create_cast",
	"create_function_c",
	"create_function_sql",
	"create_index",
	"create_index_spgist",
	"create_misc",
	"create_operator",
	"create_procedure",
	"create_role",
	"create_schema",
	"create_table",
	"create_table_like",
	"create_type",
	"create_view",
	"date",
	"dbsize",
	"delete",
	"dependency",
	"domain",
	"drop_if_exists",
	"drop_operator",
	"enum",
	"equivclass",
	"errors",
	"event_trigger",
	"explain",
	"expressions",
	"fast_default",
	"float4-misrounded-input",
	"float4",
	"float8",
	"foreign_data",
	"foreign_key",
	"functional_deps",
	"generated",
	"geometry",
	"gin",
	"gist",
	"groupingsets",
	"guc",
	"hash_func",
	"hash_index",
	"hash_part",
	"horology",
	"identity",
	"incremental_sort",
	"index_including",
	"index_including_gist",
	"indexing",
	"indirect_toast",
	"inet",
	"infinite_recurse",
	"infinite_recurse_1",
	"inherit",
	"init_privs",
	"insert",
	"insert_conflict",
	"int2",
	"int4",
	"int8",
	"interval",
	"join",
	"join_hash",
	"json",
	"json_encoding",
	"json_encoding_1",
	"json_encoding_2",
	"jsonb",
	"jsonb_jsonpath",
	"jsonpath",
	"jsonpath_encoding",
	"jsonpath_encoding_1",
	"jsonpath_encoding_2",
	"largeobject",
	"largeobject_1",
	"limit",
	"line",
	"lock",
	"lseg",
	"macaddr",
	"macaddr8",
	"matview",
	"memoize",
	"merge",
	"misc",
	"misc_functions",
	"misc_sanity",
	"money",
	"multirangetypes",
	"mvcc",
	"name",
	"namespace",
	"numeric",
	"numeric_big",
	"numerology",
	"object_address",
	"oid",
	"oidjoins",
	"opr_sanity",
	"partition_aggregate",
	"partition_info",
	"partition_join",
	"partition_prune",
	"password",
	"path",
	"pg_lsn",
	"plancache",
	"plpgsql",
	"point",
	"polygon",
	"polymorphism",
	"portals",
	"portals_p2",
	"prepare",
	"prepared_xacts",
	"prepared_xacts_1",
	"privileges",
	"psql",
	"psql_crosstab",
	"publication",
	"random",
	"rangefuncs",
	"rangetypes",
	"regex",
	"regproc",
	"reindex_catalog",
	"reloptions",
	"replica_identity",
	"returning",
	"roleattributes",
	"rowsecurity",
	"rowtypes",
	"rules",
	"sanity_check",
	"security_label",
	"select",
	"select_distinct",
	"select_distinct_on",
	"select_having",
	"select_having_1",
	"select_having_2",
	"select_implicit",
	"select_implicit_1",
	"select_implicit_2",
	"select_into",
	"select_parallel",
	"select_views",
	"sequence",
	"spgist",
	"stats",
	"stats_ext",
	"strings",
	"subscription",
	"subselect",
	"sysviews",
	"tablesample",
	"tablespace",
	"temp",
	"test_setup",
	"text",
	"tid",
	"tidrangescan",
	"tidscan",
	"time",
	"timestamp",
	"timestamptz",
	"timetz",
	"transactions",
	"triggers",
	"truncate",
	"tsdicts",
	"tsearch",
	"tsrf",
	"tstypes",
	"tuplesort",
	"txid",
	"type_sanity",
	"typed_table",
	"unicode",
	"unicode_1",
	"union",
	"updatable_views",
	"update",
	"uuid",
	"vacuum",
	"vacuum_parallel",
	"varchar",
	"varchar_1",
	"varchar_2",
	"window",
	"with",
	"write_parallel",
	"xid",
	"xml",
	"xml_1",
	"xml_2",
	"xmlmap",
	"xmlmap_1",
}

// String returns the name of the RegressionFileName.
func (name RegressionFileName) String() string {
	return regressionFileNames[name]
}

// RegressionFile represents the contents of an "expected" file from the Postgres regression tests.
// https://www.postgresql.org/docs/15/regress.html
type RegressionFile struct {
	RegressionFileName
	DependsOn  []RegressionFileName // Ensures that the named files are always run before this one, however they are not reported in this test's results
	Statements []RegressionFileStatement
}

// RegressionFileStatement is a statement within the RegressionFile.
type RegressionFileStatement struct {
	Statement            string    // This is the statement from the file
	Results              []sql.Row // This is the collection of rows that are expected as a result
	ErrorString          string    // When non-empty, checks that the statement returns an error containing this string
	DisableNormalization bool      // When true, normalization of values is disabled
	OrderBy              bool      // When true, results are required to match in their returned order.
	Skip                 bool
}

// allFiles is a map from the RegressionFileName to the RegressionFile.
var allFiles = make(map[RegressionFileName]RegressionFile)

// RegisterRegressionFile is called from within an init() function in each file. This registers the regression file, in
// addition to ensuring that there are no duplicate files.
func RegisterRegressionFile(f RegressionFile) {
	if f.RegressionFileName == RegressionFileName_invalid {
		panic("invalid regression file defined")
	}
	if _, ok := allFiles[f.RegressionFileName]; ok {
		panic("duplicate regression file")
	}
	allFiles[f.RegressionFileName] = f
}

// RunTests runs all of the files given (in addition to their dependencies, which are not tracked). Returns a Tracker,
// which contains all of the run information for the given files.
func RunTests(t *testing.T, files ...RegressionFileName) *Tracker {
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() {
		_ = conn.Close(ctx)
		controller.Stop()
		err := controller.WaitForStop()
		require.NoError(t, err)
	}()

	tracker := &Tracker{make(map[RegressionFileName]*TrackerTest)}
	for _, fileName := range files {
		runFile(t, ctx, tracker, fileName, conn, false)
	}
	return tracker
}

// GetDataFolder returns the directory of the data folder that contains the files that will be read into tables via
// COPY FROM.
func GetDataFolder() utils.RootFolderLocation {
	root, err := utils.GetRootFolder()
	if err != nil {
		panic(err)
	}
	return root.MoveRoot("testing/go/regression/data")
}

// Tracker contains the pass or fail states of all tests that have been run.
type Tracker struct {
	tests map[RegressionFileName]*TrackerTest
}

// TrackerTest represents the state of a test file.
type TrackerTest struct {
	Name    string `json:"Name"`
	PassIDs []int  `json:"PassIDs"`
	SkipIDs []int  `json:"SkipIDs"`
	FailIDs []int  `json:"FailIDs"`
}

// HasRun returns whether the tracker contains an entry for the given RegressionFileName.
func (tracker *Tracker) HasRun(rfs RegressionFileName) bool {
	_, ok := tracker.tests[rfs]
	return ok
}

// GetTotals returns the total number of passed, skipped, and failed tests.
func (tracker *Tracker) GetTotals() (pass int, skip int, fail int) {
	for _, test := range tracker.tests {
		if test == nil {
			continue
		}
		pass += len(test.PassIDs)
		skip += len(test.SkipIDs)
		fail += len(test.FailIDs)
	}
	return pass, skip, fail
}

// runFile runs the given file, in addition to running all dependencies that the file requires.
func runFile(t *testing.T, ctx context.Context, tracker *Tracker, fileName RegressionFileName, conn *pgx.Conn, isDependency bool) {
	file, ok := allFiles[fileName]
	if !ok {
		// If the file doesn't exist, then we'll pretend that it ran anyway.
		// This is due to some files being unable to be generated for a number a reasons.
		// Rather than error for every test (since they're likely not dependencies), we just skip them.
		tracker.tests[file.RegressionFileName] = nil
	}
	for _, dependency := range file.DependsOn {
		if !tracker.HasRun(dependency) {
			runFile(t, ctx, tracker, dependency, conn, true)
		}
	}
	// We don't want dependencies to count toward the tracker, error states, etc.
	// This means that we'll simply run the statements and ignore everything else.
	if isDependency {
		// Dependencies create a nil entry, so that we know to ignore it since we only want the side effects
		tracker.tests[file.RegressionFileName] = nil
		for _, rfs := range file.Statements {
			_, _ = conn.Exec(ctx, rfs.Statement)
			continue
		}
		return
	}

	t.Run(file.RegressionFileName.String(), func(t *testing.T) {
		if tracker.HasRun(file.RegressionFileName) {
			t.Fatal("Test has already run")
		}
		fileTracker := &TrackerTest{
			Name:    file.RegressionFileName.String(),
			PassIDs: nil,
			SkipIDs: nil,
			FailIDs: nil,
		}
		tracker.tests[file.RegressionFileName] = fileTracker

		for statementIndex, rfs := range file.Statements {
			t.Run(rfs.Statement, func(t *testing.T) {
				if rfs.Skip {
					fileTracker.SkipIDs = append(fileTracker.SkipIDs, statementIndex)
					t.Skip()
					return
				}
				if rfs.ErrorString != "" {
					_, err := conn.Exec(ctx, rfs.Statement)
					if assert.Error(t, err) && assert.Contains(t, err.Error(), rfs.ErrorString) {
						fileTracker.PassIDs = append(fileTracker.PassIDs, statementIndex)
					} else {
						fileTracker.FailIDs = append(fileTracker.FailIDs, statementIndex)
					}
				} else {
					success := true
					rows, err := conn.Query(ctx, rfs.Statement)
					if success = assert.NoError(t, err); success {
						readRows, err := ReadRows(rows, !rfs.DisableNormalization)
						if success = assert.NoError(t, err); success {
							if rfs.Results == nil {
								rfs.Results = []sql.Row{}
							}
							results := rfs.Results
							if !rfs.DisableNormalization {
								results = NormalizeRows(results)
							}
							if rfs.OrderBy {
								success = assert.Equal(t, results, readRows)
							} else {
								success = assert.ElementsMatch(t, results, readRows)
							}
						}
					}
					if success {
						fileTracker.PassIDs = append(fileTracker.PassIDs, statementIndex)
					} else {
						fileTracker.FailIDs = append(fileTracker.FailIDs, statementIndex)
					}
				}
			})
		}
	})
}

// CreateServer creates a server with the given database, returning a connection to the server. The server will close
// when the connection is closed (or loses its connection to the server). The accompanying WaitGroup may be used to wait
// until the server has closed.
func CreateServer(t *testing.T, database string) (context.Context, *pgx.Conn, *svcs.Controller) {
	require.NotEmpty(t, database)
	port := GetUnusedPort(t)
	host := "127.0.0.1"
	controller, err := dserver.RunInMemory(&servercfg.DoltgresConfig{
		ListenerConfig: &servercfg.DoltgresListenerConfig{
			PortNumber: &port,
			HostStr:    &host,
		},
	})
	require.NoError(t, err)

	ctx := context.Background()
	err = func() error {
		// The connection attempt may be made before the server has grabbed the port, so we'll retry the first
		// connection a few times.
		var conn *pgx.Conn
		var err error
		for i := 0; i < 3; i++ {
			conn, err = pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/", port))
			if err == nil {
				break
			} else {
				time.Sleep(time.Second)
			}
		}
		if err != nil {
			return err
		}

		defer conn.Close(ctx)
		_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s;", database))
		return err
	}()
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/%s", port, database))
	require.NoError(t, err)
	return ctx, conn, controller
}

// ReadRows reads all of the given rows into a slice, then closes the rows. If `normalizeRows` is true, then the rows
// will be normalized such that all integers are int64, etc.
func ReadRows(rows pgx.Rows, normalizeRows bool) (readRows []sql.Row, err error) {
	defer func() {
		err = errors.Join(err, rows.Err())
	}()
	var slice []sql.Row
	for rows.Next() {
		row, err := rows.Values()
		if err != nil {
			return nil, err
		}
		slice = append(slice, row)
	}
	if normalizeRows {
		return NormalizeRows(slice), nil
	} else {
		// We must always normalize Numeric values, as they have an infinite number of ways to represent the same value
		return NormalizeRowsOnlyNumeric(slice), nil
	}
}

// NormalizeRow normalizes each value's type, as the tests only want to compare values. Returns a new row.
func NormalizeRow(row sql.Row) sql.Row {
	if len(row) == 0 {
		return nil
	}
	newRow := make(sql.Row, len(row))
	for i := range row {
		switch val := row[i].(type) {
		case int:
			newRow[i] = int64(val)
		case int8:
			newRow[i] = int64(val)
		case int16:
			newRow[i] = int64(val)
		case int32:
			newRow[i] = int64(val)
		case uint:
			newRow[i] = int64(val)
		case uint8:
			newRow[i] = int64(val)
		case uint16:
			newRow[i] = int64(val)
		case uint32:
			newRow[i] = int64(val)
		case uint64:
			// PostgreSQL does not support an uint64 type, so we can always convert this to an int64 safely.
			newRow[i] = int64(val)
		case float32:
			newRow[i] = float64(val)
		case pgtype.Numeric:
			if val.NaN {
				newRow[i] = math.NaN()
			} else if val.InfinityModifier != pgtype.Finite {
				newRow[i] = math.Inf(int(val.InfinityModifier))
			} else if !val.Valid {
				newRow[i] = nil
			} else {
				fVal, err := val.Float64Value()
				if err != nil {
					panic(err)
				}
				if !fVal.Valid {
					panic("no idea why the numeric float value is invalid")
				}
				newRow[i] = fVal.Float64
			}
		case time.Time:
			newRow[i] = val.Format("2006-01-02 15:04:05")
		case map[string]interface{}:
			str, err := json.Marshal(val)
			if err != nil {
				panic(err)
			}
			newRow[i] = string(str)
		default:
			newRow[i] = val
		}
	}
	return newRow
}

// NormalizeRows normalizes each value's type within each row, as the tests only want to compare values. Returns a new
// set of rows in the same order.
func NormalizeRows(rows []sql.Row) []sql.Row {
	newRows := make([]sql.Row, len(rows))
	for i := range rows {
		newRows[i] = NormalizeRow(rows[i])
	}
	return newRows
}

// NormalizeRowsOnlyNumeric normalizes Numeric values only. There are an infinite number of ways to represent the same
// value in-memory, so we must at least normalize Numeric values.
func NormalizeRowsOnlyNumeric(rows []sql.Row) []sql.Row {
	newRows := make([]sql.Row, len(rows))
	for rowIdx, row := range rows {
		newRow := make(sql.Row, len(row))
		copy(newRow, row)
		for colIdx := range newRow {
			if numericValue, ok := newRow[colIdx].(pgtype.Numeric); ok {
				val, err := numericValue.Value()
				if err != nil {
					panic(err) // Should never happen
				}
				// Using decimal as an intermediate value will remove all differences between the string formatting
				d := decimal.RequireFromString(val.(string))
				newRow[colIdx] = Numeric(d.String())
			}
		}
		newRows[rowIdx] = newRow
	}
	return newRows
}

// GetUnusedPort returns an unused port.
func GetUnusedPort(t *testing.T) int {
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	return port
}

// Numeric creates a numeric value from a string.
func Numeric(str string) pgtype.Numeric {
	numeric := pgtype.Numeric{}
	if err := numeric.Scan(str); err != nil {
		panic(err)
	}
	return numeric
}
