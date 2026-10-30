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

package _go

import (
	"github.com/dolthub/go-mysql-server/sql"
	"testing"
)

// PostgreSQL exposes SHOW-style object metadata through catalog queries.
// Check definitions, visibility and identity without fixed OIDs or AM names.
func testNativeShowTables(t *testing.T) {
	RunScripts(t, []ScriptTest{{Name: "native table inventory and search path", SetUpScript: []string{
		"CREATE SEQUENCE seq1", "CREATE TABLE t1 (a int PRIMARY KEY, name text)", "CREATE TABLE t2 (b int PRIMARY KEY, name text)", "CREATE SCHEMA schema2", "CREATE TABLE schema2.t3 (a int)", "CREATE TABLE schema2.t4 (b int)",
	}, Assertions: []ScriptTestAssertion{
		{Query: "SELECT tablename FROM pg_tables WHERE schemaname=current_schema() ORDER BY tablename", Expected: []sql.Row{{"t1"}, {"t2"}}},
		{Query: "SELECT tablename FROM pg_tables WHERE schemaname='schema2' ORDER BY tablename", Expected: []sql.Row{{"t3"}, {"t4"}}},
		{Query: "SELECT 'schema3'::regnamespace", ExpectedErrCode: "3F000"},
		{Query: "SET search_path TO schema2", Expected: []sql.Row{}},
		{Query: "SELECT tablename FROM pg_tables WHERE schemaname=current_schema() ORDER BY tablename", Expected: []sql.Row{{"t3"}, {"t4"}}},
		{Query: "SET search_path TO schema3", Expected: []sql.Row{}},
		{Query: "SELECT tablename FROM pg_tables WHERE schemaname=current_schema()", Expected: []sql.Row{}},
	}}})
}

func testNativeShowCreateTable(t *testing.T) {
	RunScripts(t, []ScriptTest{{Name: "native definition introspection", SetUpScript: []string{
		"CREATE TABLE t1 (a int PRIMARY KEY, name text)", "CREATE TABLE t2 (b serial PRIMARY KEY, time timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP)", "CREATE TABLE t3 (a timestamp PRIMARY KEY, name varchar(100))",
	}, Assertions: []ScriptTestAssertion{
		{Query: "SELECT table_name,column_name,data_type,is_nullable,character_maximum_length FROM information_schema.columns WHERE table_schema='public' AND table_name IN ('t1','t2','t3') ORDER BY table_name,ordinal_position", Expected: []sql.Row{{"t1", "a", "integer", "NO", nil}, {"t1", "name", "text", "YES", nil}, {"t2", "b", "integer", "NO", nil}, {"t2", "time", "timestamp without time zone", "NO", nil}, {"t3", "a", "timestamp without time zone", "NO", nil}, {"t3", "name", "character varying", "YES", 100}}},
		{Query: "SELECT conrelid::regclass::text, pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid IN ('t1'::regclass,'t2'::regclass,'t3'::regclass) AND contype='p' ORDER BY conrelid::regclass::text", Expected: []sql.Row{{"t1", "PRIMARY KEY (a)"}, {"t2", "PRIMARY KEY (b)"}, {"t3", "PRIMARY KEY (a)"}}},
		{Query: "SELECT column_name, column_default LIKE CASE column_name WHEN 'b' THEN 'nextval(%' ELSE 'CURRENT_TIMESTAMP' END FROM information_schema.columns WHERE table_schema='public' AND table_name='t2' ORDER BY ordinal_position", Expected: []sql.Row{{"b", "t"}, {"time", "t"}}},
		{Query: "SELECT 'dne'::regclass", ExpectedErrCode: "42P01"},
	}}})
}

func testNativeShowIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{{Name: "native index definitions", SetUpScript: []string{"CREATE TABLE t1(a int PRIMARY KEY, name text, value int)", "CREATE INDEX idx_name ON t1(name)", "CREATE INDEX idx_value ON t1(value)"}, Assertions: []ScriptTestAssertion{
		{Query: "SELECT c.relname,i.indisunique,i.indisprimary,pg_get_indexdef(i.indexrelid,1,true) FROM pg_index i JOIN pg_class c ON c.oid=i.indexrelid WHERE i.indrelid='t1'::regclass ORDER BY c.relname", Expected: []sql.Row{{"idx_name", "f", "f", "name"}, {"idx_value", "f", "f", "value"}, {"t1_pkey", "t", "t", "a"}}},
		{Query: "SELECT 'dne'::regclass", ExpectedErrCode: "42P01"},
	}}})
}

func testNativeShowDatabasesAndSchemas(t *testing.T) {
	RunScripts(t, []ScriptTest{{Name: "native database and namespace inventory", SetUpScript: []string{"CREATE DATABASE db1", "CREATE DATABASE db2", "CREATE SCHEMA schema1", "CREATE SCHEMA schema2"}, Assertions: []ScriptTestAssertion{
		{Query: "SELECT datname FROM pg_database WHERE NOT datistemplate ORDER BY datname", Expected: []sql.Row{{"db1"}, {"db2"}, {"postgres"}}},
		{Query: "SELECT nspname FROM pg_namespace WHERE nspname IN ('information_schema','pg_catalog','public','schema1','schema2') ORDER BY nspname", Expected: []sql.Row{{"information_schema"}, {"pg_catalog"}, {"public"}, {"schema1"}, {"schema2"}}},
		{Query: "SELECT count(*) FROM pg_database WHERE datname='dne'", Expected: []sql.Row{{0}}},
	}}})
}

func testNativeShowSequences(t *testing.T) {
	// Each database is connected explicitly; PostgreSQL has no USE statement.
	RunScripts(t, []ScriptTest{
		{Name: "default database sequences", SetUpScript: []string{"CREATE SEQUENCE seq1", "CREATE SEQUENCE seq2", "CREATE SCHEMA schema1", "CREATE SEQUENCE schema1.seq3"}, Assertions: []ScriptTestAssertion{{Query: "SELECT schemaname,sequencename FROM pg_sequences WHERE schemaname IN ('public','schema1') ORDER BY schemaname,sequencename", Expected: []sql.Row{{"public", "seq1"}, {"public", "seq2"}, {"schema1", "seq3"}}}}},
		{Name: "explicit database sequences", Database: "db1", SetUpScript: []string{"CREATE SEQUENCE seq4", "CREATE SCHEMA schema2", "CREATE SEQUENCE schema2.seq5"}, Assertions: []ScriptTestAssertion{{Query: "SELECT current_database(),schemaname,sequencename FROM pg_sequences WHERE schemaname IN ('public','schema2') ORDER BY schemaname,sequencename", Expected: []sql.Row{{"db1", "public", "seq4"}, {"db1", "schema2", "seq5"}}}}},
	})
}
