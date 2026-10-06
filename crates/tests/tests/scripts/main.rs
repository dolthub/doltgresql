// Copyright 2026 Dolthub, Inc.
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

//! Script tests ported from the Go test suite, whose expectations are what Postgres returns.

#![allow(unused_imports, clippy::octal_escapes)]

mod adaptive_encoding;
mod adaptive_encoding_keys;
mod all_quantifier;
mod alter;
mod alter_table;
mod application_settings;
mod array_dimension_limit;
mod array_regtype;
mod array_slice;
mod array_update;
mod as_of;
mod auth;
mod auth_quick;
mod binding;
mod bit_string_length;
mod bytea_length;
mod coercion;
mod command_tag;
mod composite_unique_index;
mod conflicts_root_object;
mod convert_to;
mod copy_from;
mod copy_to;
mod count_distinct;
mod create_aggregate;
mod create_cast;
mod create_database;
mod create_extension;
mod create_function_plpgsql;
mod create_function_sql;
mod create_operator;
mod create_procedure_plpgsql;
mod create_procedure_sql;
mod create_table;
mod create_view;
mod delete;
#[path = "do.rs"]
mod do_statement;
mod dolt_functions;
mod dolt_procedures_record;
mod dolt_tables;
mod domain;
mod drop_database;
mod drop_function;
mod drop_procedure;
mod drop_table;
mod empty_array_type;
mod explain;
mod expressions;
mod extension_emulation;
mod foreach_slice;
mod foreign_keys;
mod functional_index_multi_expr;
mod functions;
mod getting_started_guide;
mod groupby;
mod hashtext;
mod identity;
mod implicit_transaction;
mod index;
mod information_schema;
mod insert;
mod issues;
mod json_function;
mod jsonb_index;
mod lateral;
mod limit;
mod lock;
mod locking_clause;
mod lookup_join_kvexec;
mod merge;
mod multi_statement;
mod multidimensional_array;
mod operators;
mod parameters;
mod partial_index_join;
mod pgcatalog;
mod pgvector_catalog;
mod pgvector_index;
mod pgvector_knn;
mod pgvector_upstream_index;
mod pgvector_v0_8_6;
mod plpgsql_continue;
mod plpgsql_found;
mod plpgsql_identifiers;
mod plpgsql_into;
mod plpgsql_record;
mod plpgsql_type_alias;
mod prepared_statement;
mod psql;
mod record;
mod regression;
mod result_batch;
mod root_object_collections;
mod scalar_subscript;
mod schemas;
mod select;
mod sequences;
mod server_config_variable;
mod session;
mod set;
mod set_constraints;
mod set_local;
mod set_operations;
mod show;
mod smoke;
mod sqlstate;
mod stats;
mod stats_usage;
mod subqueries;
mod trigger;
mod types;
mod union;
mod update;
mod uuid_ossp_v1_1;
mod values_statement;
mod window;
mod wire;
mod with;
mod with_recursive_cycle;
mod xml;
