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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestSetStatements(t *testing.T) {
	RunScripts(t, setStmts)
}

var setStmts = []ScriptTest{
	{
		Name:        "special cases",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SET TIME ZONE LOCAL;",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET TIME ZONE DEFAULT;",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET SCHEMA 'postgres';",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET NAMES 'UTF8';",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET SEED 1;",
				Expected: []sql.Row{{}},
			},
		},
	},
	{
		Name:        "all configuration parameters",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SET allow_in_place_tablespaces TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET allow_system_table_mods TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET application_name TO 'psql'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET archive_cleanup_command TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET archive_command TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET archive_library TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET archive_mode TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET archive_timeout TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET array_nulls TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET authentication_timeout TO '120'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET autovacuum TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET autovacuum_analyze_scale_factor TO '0.1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET autovacuum_analyze_threshold TO '50'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET autovacuum_freeze_max_age TO '200000000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET autovacuum_max_workers TO '3'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET autovacuum_multixact_freeze_max_age TO '400000000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET autovacuum_naptime TO '60'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET autovacuum_vacuum_cost_delay TO '2'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET autovacuum_vacuum_cost_limit TO '-1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET autovacuum_vacuum_insert_scale_factor TO '0.2'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET autovacuum_vacuum_insert_threshold TO '1000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET autovacuum_vacuum_scale_factor TO '0.2'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET autovacuum_vacuum_threshold TO '50'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET autovacuum_work_mem TO '-1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET backend_flush_after TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET backslash_quote TO 'safe_encoding'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET backtrace_functions TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET bgwriter_delay TO '200'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET bgwriter_flush_after TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET bgwriter_lru_maxpages TO '100'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET bgwriter_lru_multiplier TO '2'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET block_size TO '8192'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET bonjour TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET bonjour_name TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET bytea_output TO 'hex'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET check_function_bodies TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET checkpoint_completion_target TO '0.9'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET checkpoint_flush_after TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET checkpoint_timeout TO '300'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET checkpoint_warning TO '30'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET client_connection_check_interval TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET client_encoding TO 'UTF8'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET client_min_messages TO 'notice'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET cluster_name TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET commit_delay TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET commit_siblings TO '5'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET compute_query_id TO 'auto'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET config_file TO '/Users/postgres/postgresql.conf'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET constraint_exclusion TO 'partition'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET cpu_index_tuple_cost TO '0.005'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET cpu_operator_cost TO '0.0025'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET cpu_tuple_cost TO '0.01'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET createrole_self_grant TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET cursor_tuple_fraction TO '0.1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET data_checksums TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET data_directory TO '/Users/postgres'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET data_directory_mode TO '448'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET data_sync_retry TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET DateStyle TO 'ISO, MDY'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET db_user_namespace TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET deadlock_timeout TO '1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET debug_assertions TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET debug_discard_caches TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET debug_io_direct TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET debug_logical_replication_streaming TO 'buffered'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET debug_parallel_query TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET debug_pretty_print TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET debug_print_parse TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET debug_print_plan TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET debug_print_rewritten TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET default_statistics_target TO '100'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET default_table_access_method TO 'heap'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET default_tablespace TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET default_text_search_config TO 'pg_catalog.english'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET default_toast_compression TO 'pglz'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET default_transaction_deferrable TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET default_transaction_isolation TO 'read committed'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET default_transaction_read_only TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET dynamic_library_path TO '$libdir'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET dynamic_shared_memory_type TO 'posix'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET effective_cache_size TO '400000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET effective_io_concurrency TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_async_append TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_bitmapscan TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_gathermerge TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_hashagg TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_hashjoin TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_incremental_sort TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_indexonlyscan TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_indexscan TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_material TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_memoize TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_mergejoin TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_nestloop TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_parallel_append TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_parallel_hash TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_partition_pruning TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_partitionwise_aggregate TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_partitionwise_join TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_presorted_aggregate TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_seqscan TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_sort TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET enable_tidscan TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET escape_string_warning TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET event_source TO 'PostgreSQL'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET exit_on_error TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET external_pid_file TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET extra_float_digits TO '1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET from_collapse_limit TO '8'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET fsync TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET full_page_writes TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET geqo TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET geqo_effort TO '5'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET geqo_generations TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET geqo_pool_size TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET geqo_seed TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET geqo_selection_bias TO '2'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET geqo_threshold TO '12'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET gin_fuzzy_search_limit TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET gin_pending_list_limit TO '4000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET gss_accept_delegation TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET hash_mem_multiplier TO '2'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET hba_file TO '/Users/postgres/pg_hba.conf'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET hot_standby TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET hot_standby_feedback TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET huge_page_size TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET huge_pages TO 'try'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET icu_validation_level TO 'warning'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ident_file TO '/Users/postgres/pg_ident.conf'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET idle_in_transaction_session_timeout TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET idle_session_timeout TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ignore_checksum_failure TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ignore_invalid_pages TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ignore_system_indexes TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET in_hot_standby TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET integer_datetimes TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET IntervalStyle TO 'postgres'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET jit TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET jit_above_cost TO '100000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET jit_debugging_support TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET jit_dump_bitcode TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET jit_expressions TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET jit_inline_above_cost TO '500000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET jit_optimize_above_cost TO '500000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET jit_profiling_support TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET jit_provider TO 'llvmjit'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET jit_tuple_deforming TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET join_collapse_limit TO '8'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET krb_caseins_users TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET krb_server_keyfile TO 'FILE:'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET lc_messages TO 'en_US.UTF-8'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET lc_monetary TO 'en_US.UTF-8'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET lc_numeric TO 'en_US.UTF-8'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET lc_time TO 'en_US.UTF-8'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET listen_addresses TO 'localhost'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET lo_compat_privileges TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET local_preload_libraries TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET lock_timeout TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_autovacuum_min_duration TO '600'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_checkpoints TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_connections TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_destination TO 'stderr'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_directory TO 'log'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_disconnections TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_duration TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_error_verbosity TO 'default'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_executor_stats TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_file_mode TO '384'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_filename TO 'postgresql-%Y-%m-%d_%H%M%S.log'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_hostname TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_line_prefix TO '%m [%p]'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_lock_waits TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_min_duration_sample TO '-1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_min_duration_statement TO '-1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_min_error_statement TO 'error'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_min_messages TO 'warning'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_parameter_max_length TO '-1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_parameter_max_length_on_error TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_parser_stats TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_planner_stats TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_recovery_conflict_waits TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_replication_commands TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_rotation_age TO '1440'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_rotation_size TO '10240'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_startup_progress_interval TO '10'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_statement TO 'none'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_statement_sample_rate TO '1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_statement_stats TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_temp_files TO '-1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_timezone TO 'America/Los_Angeles'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_transaction_sample_rate TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET log_truncate_on_rotation TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET logging_collector TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET logical_decoding_work_mem TO '64000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET maintenance_io_concurrency TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET maintenance_work_mem TO '64000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_connections TO '100'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_files_per_process TO '1000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_function_args TO '100'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_identifier_length TO '63'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_index_keys TO '32'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_locks_per_transaction TO '64'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_logical_replication_workers TO '4'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_parallel_apply_workers_per_subscription TO '2'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_parallel_maintenance_workers TO '2'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_parallel_workers TO '8'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_parallel_workers_per_gather TO '2'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_pred_locks_per_page TO '2'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_pred_locks_per_relation TO '-2'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_pred_locks_per_transaction TO '64'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_prepared_transactions TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_replication_slots TO '10'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_slot_wal_keep_size TO '-1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_stack_depth TO '2000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_standby_archive_delay TO '30'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_standby_streaming_delay TO '30'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_sync_workers_per_subscription TO '2'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_wal_senders TO '10'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_wal_size TO '1000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET max_worker_processes TO '8'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET min_dynamic_shared_memory TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET min_parallel_index_scan_size TO '512'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET min_parallel_table_scan_size TO '800'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET min_wal_size TO '8000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET old_snapshot_threshold TO '-1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET parallel_leader_participation TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET parallel_setup_cost TO '1000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET parallel_tuple_cost TO '0.1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET password_encryption TO 'scram-sha-256'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET plan_cache_mode TO 'auto'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET port TO '5432'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET post_auth_delay TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET pre_auth_delay TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET primary_conninfo TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET primary_slot_name TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET quote_all_identifiers TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET random_page_cost TO '4'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET recovery_end_command TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET recovery_init_sync_method TO 'fsync'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET recovery_min_apply_delay TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET recovery_prefetch TO 'try'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET recovery_target TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET recovery_target_action TO 'pause'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET recovery_target_inclusive TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET recovery_target_lsn TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET recovery_target_name TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET recovery_target_time TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET recovery_target_timeline TO 'latest'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET recovery_target_xid TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET recursive_worktable_factor TO '10'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET remove_temp_files_after_crash TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET reserved_connections TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET restart_after_crash TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET restore_command TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET row_security TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET scram_iterations TO '4096'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET search_path TO '\"$user\", public'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET segment_size TO '131072'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET send_abort_for_crash TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET send_abort_for_kill TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET seq_page_cost TO '1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET server_encoding TO 'UTF8'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET server_version TO '16.1 (Homebrew)'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET server_version_num TO '160001'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET session_preload_libraries TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET session_replication_role TO 'origin'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET shared_buffers TO '128000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET shared_memory_size TO '143000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET shared_memory_size_in_huge_pages TO '-1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET shared_memory_type TO 'mmap'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET shared_preload_libraries TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ssl TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ssl_ca_file TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ssl_cert_file TO 'server.crt'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ssl_ciphers TO 'HIGH:MEDIUM:'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ssl_crl_dir TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ssl_crl_file TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ssl_dh_params_file TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ssl_ecdh_curve TO 'prime256v1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ssl_key_file TO 'server.key'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ssl_library TO 'OpenSSL'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ssl_max_protocol_version TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ssl_min_protocol_version TO 'TLSv1.2'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ssl_passphrase_command TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ssl_passphrase_command_supports_reload TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET ssl_prefer_server_ciphers TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET standard_conforming_strings TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET statement_timeout TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET stats_fetch_consistency TO 'cache'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET superuser_reserved_connections TO '3'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET synchronize_seqscans TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET synchronous_commit TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET synchronous_standby_names TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET syslog_facility TO 'local0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET syslog_ident TO 'postgres'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET syslog_sequence_numbers TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET syslog_split_messages TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET tcp_keepalives_count TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET tcp_keepalives_idle TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET tcp_keepalives_interval TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET tcp_user_timeout TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET temp_buffers TO '8000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET temp_file_limit TO '-1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET temp_tablespaces TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET TimeZone TO 'America/Los_Angeles'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET timezone_abbreviations TO 'Default'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET trace_notify TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET trace_recovery_messages TO 'log'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET trace_sort TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET track_activities TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET track_activity_query_size TO '1024'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET track_commit_timestamp TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET track_counts TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET track_functions TO 'none'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET track_io_timing TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET track_wal_io_timing TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET transaction_deferrable TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET transaction_isolation TO 'read committed'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET transaction_read_only TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET transform_null_equals TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET unix_socket_directories TO '/tmp'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET unix_socket_group TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET unix_socket_permissions TO '511'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET update_process_title TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET vacuum_buffer_usage_limit TO '256'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET vacuum_cost_delay TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET vacuum_cost_limit TO '200'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET vacuum_cost_page_dirty TO '20'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET vacuum_cost_page_hit TO '1'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET vacuum_cost_page_miss TO '2'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET vacuum_failsafe_age TO '1600000000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET vacuum_freeze_min_age TO '50000000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET vacuum_freeze_table_age TO '150000000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET vacuum_multixact_failsafe_age TO '1600000000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET vacuum_multixact_freeze_min_age TO '5000000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET vacuum_multixact_freeze_table_age TO '150000000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_block_size TO '8192'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_buffers TO '4000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_compression TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_consistency_checking TO ''",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_decode_buffer_size TO '524288'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_init_zero TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_keep_size TO '0'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_level TO 'replica'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_log_hints TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_receiver_create_temp_slot TO 'off'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_receiver_status_interval TO '10'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_receiver_timeout TO '60'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_recycle TO 'on'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_retrieve_retry_interval TO '5'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_segment_size TO '16777216'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_sender_timeout TO '60'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_skip_threshold TO '2000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_sync_method TO 'open_datasync'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_writer_delay TO '200'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET wal_writer_flush_after TO '1000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET work_mem TO '4000'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET xmlbinary TO 'base64'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET xmloption TO 'content'",
				Expected: []sql.Row{{}},
			},
			{
				Query:    "SET zero_damaged_pages TO 'off'",
				Expected: []sql.Row{{}},
			},
		},
	},
}
