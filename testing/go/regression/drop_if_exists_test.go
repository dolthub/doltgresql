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
	"testing"
)

func TestDropIfExists(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_drop_if_exists)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_drop_if_exists,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement:   `DROP TABLE test_exists;`,
				ErrorString: `table "test_exists" does not exist`,
			},
			{
				Statement: `DROP TABLE IF EXISTS test_exists;`,
			},
			{
				Statement: `CREATE TABLE test_exists (a int, b text);`,
			},
			{
				Statement:   `DROP VIEW test_view_exists;`,
				ErrorString: `view "test_view_exists" does not exist`,
			},
			{
				Statement: `DROP VIEW IF EXISTS test_view_exists;`,
			},
			{
				Statement: `CREATE VIEW test_view_exists AS select * from test_exists;`,
			},
			{
				Statement: `DROP VIEW IF EXISTS test_view_exists;`,
			},
			{
				Statement:   `DROP VIEW test_view_exists;`,
				ErrorString: `view "test_view_exists" does not exist`,
			},
			{
				Statement:   `DROP INDEX test_index_exists;`,
				ErrorString: `index "test_index_exists" does not exist`,
			},
			{
				Statement: `DROP INDEX IF EXISTS test_index_exists;`,
			},
			{
				Statement: `CREATE INDEX test_index_exists on test_exists(a);`,
			},
			{
				Statement: `DROP INDEX IF EXISTS test_index_exists;`,
			},
			{
				Statement:   `DROP INDEX test_index_exists;`,
				ErrorString: `index "test_index_exists" does not exist`,
			},
			{
				Statement:   `DROP SEQUENCE test_sequence_exists;`,
				ErrorString: `sequence "test_sequence_exists" does not exist`,
			},
			{
				Statement: `DROP SEQUENCE IF EXISTS test_sequence_exists;`,
			},
			{
				Statement: `CREATE SEQUENCE test_sequence_exists;`,
			},
			{
				Statement: `DROP SEQUENCE IF EXISTS test_sequence_exists;`,
			},
			{
				Statement:   `DROP SEQUENCE test_sequence_exists;`,
				ErrorString: `sequence "test_sequence_exists" does not exist`,
			},
			{
				Statement:   `DROP SCHEMA test_schema_exists;`,
				ErrorString: `schema "test_schema_exists" does not exist`,
			},
			{
				Statement: `DROP SCHEMA IF EXISTS test_schema_exists;`,
			},
			{
				Statement: `CREATE SCHEMA test_schema_exists;`,
			},
			{
				Statement: `DROP SCHEMA IF EXISTS test_schema_exists;`,
			},
			{
				Statement:   `DROP SCHEMA test_schema_exists;`,
				ErrorString: `schema "test_schema_exists" does not exist`,
			},
			{
				Statement:   `DROP TYPE test_type_exists;`,
				ErrorString: `type "test_type_exists" does not exist`,
			},
			{
				Statement: `DROP TYPE IF EXISTS test_type_exists;`,
			},
			{
				Statement: `CREATE type test_type_exists as (a int, b text);`,
			},
			{
				Statement: `DROP TYPE IF EXISTS test_type_exists;`,
			},
			{
				Statement:   `DROP TYPE test_type_exists;`,
				ErrorString: `type "test_type_exists" does not exist`,
			},
			{
				Statement:   `DROP DOMAIN test_domain_exists;`,
				ErrorString: `type "test_domain_exists" does not exist`,
			},
			{
				Statement: `DROP DOMAIN IF EXISTS test_domain_exists;`,
			},
			{
				Statement: `CREATE domain test_domain_exists as int not null check (value > 0);`,
			},
			{
				Statement: `DROP DOMAIN IF EXISTS test_domain_exists;`,
			},
			{
				Statement:   `DROP DOMAIN test_domain_exists;`,
				ErrorString: `type "test_domain_exists" does not exist`,
			},
			{
				Statement: `---
---
CREATE USER regress_test_u1;`,
			},
			{
				Statement: `CREATE ROLE regress_test_r1;`,
			},
			{
				Statement: `CREATE GROUP regress_test_g1;`,
			},
			{
				Statement:   `DROP USER regress_test_u2;`,
				ErrorString: `role "regress_test_u2" does not exist`,
			},
			{
				Statement: `DROP USER IF EXISTS regress_test_u1, regress_test_u2;`,
			},
			{
				Statement:   `DROP USER regress_test_u1;`,
				ErrorString: `role "regress_test_u1" does not exist`,
			},
			{
				Statement:   `DROP ROLE regress_test_r2;`,
				ErrorString: `role "regress_test_r2" does not exist`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_test_r1, regress_test_r2;`,
			},
			{
				Statement:   `DROP ROLE regress_test_r1;`,
				ErrorString: `role "regress_test_r1" does not exist`,
			},
			{
				Statement:   `DROP GROUP regress_test_g2;`,
				ErrorString: `role "regress_test_g2" does not exist`,
			},
			{
				Statement: `DROP GROUP IF EXISTS regress_test_g1, regress_test_g2;`,
			},
			{
				Statement:   `DROP GROUP regress_test_g1;`,
				ErrorString: `role "regress_test_g1" does not exist`,
			},
			{
				Statement: `DROP COLLATION IF EXISTS test_collation_exists;`,
			},
			{
				Statement:   `DROP CONVERSION test_conversion_exists;`,
				ErrorString: `conversion "test_conversion_exists" does not exist`,
			},
			{
				Statement: `DROP CONVERSION IF EXISTS test_conversion_exists;`,
			},
			{
				Statement: `CREATE CONVERSION test_conversion_exists
    FOR 'LATIN1' TO 'UTF8' FROM iso8859_1_to_utf8;`,
			},
			{
				Statement: `DROP CONVERSION test_conversion_exists;`,
			},
			{
				Statement:   `DROP TEXT SEARCH PARSER test_tsparser_exists;`,
				ErrorString: `text search parser "test_tsparser_exists" does not exist`,
			},
			{
				Statement: `DROP TEXT SEARCH PARSER IF EXISTS test_tsparser_exists;`,
			},
			{
				Statement:   `DROP TEXT SEARCH DICTIONARY test_tsdict_exists;`,
				ErrorString: `text search dictionary "test_tsdict_exists" does not exist`,
			},
			{
				Statement: `DROP TEXT SEARCH DICTIONARY IF EXISTS test_tsdict_exists;`,
			},
			{
				Statement: `CREATE TEXT SEARCH DICTIONARY test_tsdict_exists (
        Template=ispell,
        DictFile=ispell_sample,
        AffFile=ispell_sample
);`,
			},
			{
				Statement: `DROP TEXT SEARCH DICTIONARY test_tsdict_exists;`,
			},
			{
				Statement:   `DROP TEXT SEARCH TEMPLATE test_tstemplate_exists;`,
				ErrorString: `text search template "test_tstemplate_exists" does not exist`,
			},
			{
				Statement: `DROP TEXT SEARCH TEMPLATE IF EXISTS test_tstemplate_exists;`,
			},
			{
				Statement:   `DROP TEXT SEARCH CONFIGURATION test_tsconfig_exists;`,
				ErrorString: `text search configuration "test_tsconfig_exists" does not exist`,
			},
			{
				Statement: `DROP TEXT SEARCH CONFIGURATION IF EXISTS test_tsconfig_exists;`,
			},
			{
				Statement: `CREATE TEXT SEARCH CONFIGURATION test_tsconfig_exists (COPY=english);`,
			},
			{
				Statement: `DROP TEXT SEARCH CONFIGURATION test_tsconfig_exists;`,
			},
			{
				Statement:   `DROP EXTENSION test_extension_exists;`,
				ErrorString: `extension "test_extension_exists" does not exist`,
			},
			{
				Statement: `DROP EXTENSION IF EXISTS test_extension_exists;`,
			},
			{
				Statement:   `DROP FUNCTION test_function_exists();`,
				ErrorString: `function test_function_exists() does not exist`,
			},
			{
				Statement: `DROP FUNCTION IF EXISTS test_function_exists();`,
			},
			{
				Statement:   `DROP FUNCTION test_function_exists(int, text, int[]);`,
				ErrorString: `function test_function_exists(integer, text, integer[]) does not exist`,
			},
			{
				Statement: `DROP FUNCTION IF EXISTS test_function_exists(int, text, int[]);`,
			},
			{
				Statement:   `DROP AGGREGATE test_aggregate_exists(*);`,
				ErrorString: `aggregate test_aggregate_exists(*) does not exist`,
			},
			{
				Statement: `DROP AGGREGATE IF EXISTS test_aggregate_exists(*);`,
			},
			{
				Statement:   `DROP AGGREGATE test_aggregate_exists(int);`,
				ErrorString: `aggregate test_aggregate_exists(integer) does not exist`,
			},
			{
				Statement: `DROP AGGREGATE IF EXISTS test_aggregate_exists(int);`,
			},
			{
				Statement:   `DROP OPERATOR @#@ (int, int);`,
				ErrorString: `operator does not exist: integer @#@ integer`,
			},
			{
				Statement: `DROP OPERATOR IF EXISTS @#@ (int, int);`,
			},
			{
				Statement: `CREATE OPERATOR @#@
        (leftarg = int8, rightarg = int8, procedure = int8xor);`,
			},
			{
				Statement: `DROP OPERATOR @#@ (int8, int8);`,
			},
			{
				Statement:   `DROP LANGUAGE test_language_exists;`,
				ErrorString: `language "test_language_exists" does not exist`,
			},
			{
				Statement: `DROP LANGUAGE IF EXISTS test_language_exists;`,
			},
			{
				Statement:   `DROP CAST (text AS text);`,
				ErrorString: `cast from type text to type text does not exist`,
			},
			{
				Statement: `DROP CAST IF EXISTS (text AS text);`,
			},
			{
				Statement:   `DROP TRIGGER test_trigger_exists ON test_exists;`,
				ErrorString: `trigger "test_trigger_exists" for table "test_exists" does not exist`,
			},
			{
				Statement: `DROP TRIGGER IF EXISTS test_trigger_exists ON test_exists;`,
			},
			{
				Statement:   `DROP TRIGGER test_trigger_exists ON no_such_table;`,
				ErrorString: `relation "no_such_table" does not exist`,
			},
			{
				Statement: `DROP TRIGGER IF EXISTS test_trigger_exists ON no_such_table;`,
			},
			{
				Statement:   `DROP TRIGGER test_trigger_exists ON no_such_schema.no_such_table;`,
				ErrorString: `schema "no_such_schema" does not exist`,
			},
			{
				Statement: `DROP TRIGGER IF EXISTS test_trigger_exists ON no_such_schema.no_such_table;`,
			},
			{
				Statement: `CREATE TRIGGER test_trigger_exists
    BEFORE UPDATE ON test_exists
    FOR EACH ROW EXECUTE PROCEDURE suppress_redundant_updates_trigger();`,
			},
			{
				Statement: `DROP TRIGGER test_trigger_exists ON test_exists;`,
			},
			{
				Statement:   `DROP RULE test_rule_exists ON test_exists;`,
				ErrorString: `rule "test_rule_exists" for relation "test_exists" does not exist`,
			},
			{
				Statement: `DROP RULE IF EXISTS test_rule_exists ON test_exists;`,
			},
			{
				Statement:   `DROP RULE test_rule_exists ON no_such_table;`,
				ErrorString: `relation "no_such_table" does not exist`,
			},
			{
				Statement: `DROP RULE IF EXISTS test_rule_exists ON no_such_table;`,
			},
			{
				Statement:   `DROP RULE test_rule_exists ON no_such_schema.no_such_table;`,
				ErrorString: `schema "no_such_schema" does not exist`,
			},
			{
				Statement: `DROP RULE IF EXISTS test_rule_exists ON no_such_schema.no_such_table;`,
			},
			{
				Statement: `CREATE RULE test_rule_exists AS ON INSERT TO test_exists
    DO INSTEAD
    INSERT INTO test_exists VALUES (NEW.a, NEW.b || NEW.a::text);`,
			},
			{
				Statement: `DROP RULE test_rule_exists ON test_exists;`,
			},
			{
				Statement:   `DROP FOREIGN DATA WRAPPER test_fdw_exists;`,
				ErrorString: `foreign-data wrapper "test_fdw_exists" does not exist`,
			},
			{
				Statement: `DROP FOREIGN DATA WRAPPER IF EXISTS test_fdw_exists;`,
			},
			{
				Statement:   `DROP SERVER test_server_exists;`,
				ErrorString: `server "test_server_exists" does not exist`,
			},
			{
				Statement: `DROP SERVER IF EXISTS test_server_exists;`,
			},
			{
				Statement:   `DROP OPERATOR CLASS test_operator_class USING btree;`,
				ErrorString: `operator class "test_operator_class" does not exist for access method "btree"`,
			},
			{
				Statement: `DROP OPERATOR CLASS IF EXISTS test_operator_class USING btree;`,
			},
			{
				Statement:   `DROP OPERATOR CLASS test_operator_class USING no_such_am;`,
				ErrorString: `access method "no_such_am" does not exist`,
			},
			{
				Statement:   `DROP OPERATOR CLASS IF EXISTS test_operator_class USING no_such_am;`,
				ErrorString: `access method "no_such_am" does not exist`,
			},
			{
				Statement:   `DROP OPERATOR FAMILY test_operator_family USING btree;`,
				ErrorString: `operator family "test_operator_family" does not exist for access method "btree"`,
			},
			{
				Statement: `DROP OPERATOR FAMILY IF EXISTS test_operator_family USING btree;`,
			},
			{
				Statement:   `DROP OPERATOR FAMILY test_operator_family USING no_such_am;`,
				ErrorString: `access method "no_such_am" does not exist`,
			},
			{
				Statement:   `DROP OPERATOR FAMILY IF EXISTS test_operator_family USING no_such_am;`,
				ErrorString: `access method "no_such_am" does not exist`,
			},
			{
				Statement:   `DROP ACCESS METHOD no_such_am;`,
				ErrorString: `access method "no_such_am" does not exist`,
			},
			{
				Statement: `DROP ACCESS METHOD IF EXISTS no_such_am;`,
			},
			{
				Statement: `DROP TABLE IF EXISTS test_exists;`,
			},
			{
				Statement:   `DROP TABLE test_exists;`,
				ErrorString: `table "test_exists" does not exist`,
			},
			{
				Statement: `DROP AGGREGATE IF EXISTS no_such_schema.foo(int);`,
			},
			{
				Statement: `DROP AGGREGATE IF EXISTS foo(no_such_type);`,
			},
			{
				Statement: `DROP AGGREGATE IF EXISTS foo(no_such_schema.no_such_type);`,
			},
			{
				Statement: `DROP CAST IF EXISTS (INTEGER AS no_such_type2);`,
			},
			{
				Statement: `DROP CAST IF EXISTS (no_such_type1 AS INTEGER);`,
			},
			{
				Statement: `DROP CAST IF EXISTS (INTEGER AS no_such_schema.bar);`,
			},
			{
				Statement: `DROP CAST IF EXISTS (no_such_schema.foo AS INTEGER);`,
			},
			{
				Statement: `DROP COLLATION IF EXISTS no_such_schema.foo;`,
			},
			{
				Statement: `DROP CONVERSION IF EXISTS no_such_schema.foo;`,
			},
			{
				Statement: `DROP DOMAIN IF EXISTS no_such_schema.foo;`,
			},
			{
				Statement: `DROP FOREIGN TABLE IF EXISTS no_such_schema.foo;`,
			},
			{
				Statement: `DROP FUNCTION IF EXISTS no_such_schema.foo();`,
			},
			{
				Statement: `DROP FUNCTION IF EXISTS foo(no_such_type);`,
			},
			{
				Statement: `DROP FUNCTION IF EXISTS foo(no_such_schema.no_such_type);`,
			},
			{
				Statement: `DROP INDEX IF EXISTS no_such_schema.foo;`,
			},
			{
				Statement: `DROP MATERIALIZED VIEW IF EXISTS no_such_schema.foo;`,
			},
			{
				Statement: `DROP OPERATOR IF EXISTS no_such_schema.+ (int, int);`,
			},
			{
				Statement: `DROP OPERATOR IF EXISTS + (no_such_type, no_such_type);`,
			},
			{
				Statement: `DROP OPERATOR IF EXISTS + (no_such_schema.no_such_type, no_such_schema.no_such_type);`,
			},
			{
				Statement: `DROP OPERATOR IF EXISTS # (NONE, no_such_schema.no_such_type);`,
			},
			{
				Statement: `DROP OPERATOR CLASS IF EXISTS no_such_schema.widget_ops USING btree;`,
			},
			{
				Statement: `DROP OPERATOR FAMILY IF EXISTS no_such_schema.float_ops USING btree;`,
			},
			{
				Statement: `DROP RULE IF EXISTS foo ON no_such_schema.bar;`,
			},
			{
				Statement: `DROP SEQUENCE IF EXISTS no_such_schema.foo;`,
			},
			{
				Statement: `DROP TABLE IF EXISTS no_such_schema.foo;`,
			},
			{
				Statement: `DROP TEXT SEARCH CONFIGURATION IF EXISTS no_such_schema.foo;`,
			},
			{
				Statement: `DROP TEXT SEARCH DICTIONARY IF EXISTS no_such_schema.foo;`,
			},
			{
				Statement: `DROP TEXT SEARCH PARSER IF EXISTS no_such_schema.foo;`,
			},
			{
				Statement: `DROP TEXT SEARCH TEMPLATE IF EXISTS no_such_schema.foo;`,
			},
			{
				Statement: `DROP TRIGGER IF EXISTS foo ON no_such_schema.bar;`,
			},
			{
				Statement: `DROP TYPE IF EXISTS no_such_schema.foo;`,
			},
			{
				Statement: `DROP VIEW IF EXISTS no_such_schema.foo;`,
			},
			{
				Statement: `CREATE FUNCTION test_ambiguous_funcname(int) returns int as $$ select $1; $$ language sql;`,
			},
			{
				Statement: `CREATE FUNCTION test_ambiguous_funcname(text) returns text as $$ select $1; $$ language sql;`,
			},
			{
				Statement:   `DROP FUNCTION test_ambiguous_funcname;`,
				ErrorString: `function name "test_ambiguous_funcname" is not unique`,
			},
			{
				Statement:   `DROP FUNCTION IF EXISTS test_ambiguous_funcname;`,
				ErrorString: `function name "test_ambiguous_funcname" is not unique`,
			},
			{
				Statement: `DROP FUNCTION test_ambiguous_funcname(int);`,
			},
			{
				Statement: `DROP FUNCTION test_ambiguous_funcname(text);`,
			},
			{
				Statement: `CREATE PROCEDURE test_ambiguous_procname(int) as $$ begin end; $$ language plpgsql;`,
			},
			{
				Statement: `CREATE PROCEDURE test_ambiguous_procname(text) as $$ begin end; $$ language plpgsql;`,
			},
			{
				Statement:   `DROP PROCEDURE test_ambiguous_procname;`,
				ErrorString: `procedure name "test_ambiguous_procname" is not unique`,
			},
			{
				Statement:   `DROP PROCEDURE IF EXISTS test_ambiguous_procname;`,
				ErrorString: `procedure name "test_ambiguous_procname" is not unique`,
			},
			{
				Statement:   `DROP ROUTINE IF EXISTS test_ambiguous_procname;`,
				ErrorString: `routine name "test_ambiguous_procname" is not unique`,
			},
			{
				Statement: `DROP PROCEDURE test_ambiguous_procname(int);`,
			},
			{
				Statement: `DROP PROCEDURE test_ambiguous_procname(text);`,
			},
			{
				Statement:   `drop database test_database_exists (force);`,
				ErrorString: `database "test_database_exists" does not exist`,
			},
			{
				Statement:   `drop database test_database_exists with (force);`,
				ErrorString: `database "test_database_exists" does not exist`,
			},
			{
				Statement: `drop database if exists test_database_exists (force);`,
			},
			{
				Statement: `drop database if exists test_database_exists with (force);`,
			},
		},
	})
}
