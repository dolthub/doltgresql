package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestInfoSchemaColumns(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "information_schema.columns",
			SetUpScript: []string{
				"create table test_table (id int primary key, col1 varchar(255));",
				"create view test_view as select * from test_table;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT DISTINCT table_schema FROM information_schema.columns ORDER BY table_schema;`,
					Expected: []sql.Row{
						{"information_schema"}, {"pg_catalog"}, {"public"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema, table_name, column_name FROM information_schema.columns WHERE table_schema='public' ORDER BY table_name;`,
					Expected: []sql.Row{
						{"postgres", "public", "test_table", "id"},
						{"postgres", "public", "test_table", "col1"},
						{"postgres", "public", "test_view", ""},
					},
				},
				{
					Query: `SELECT 
						columns.column_name, 
						pg_catalog.col_description(('"' || table_catalog || '"."' || table_schema || '"."' || table_name || '"')::regclass::oid, ordinal_position) AS description, 
						('"' || "udt_schema" || '"."' || "udt_name" || '"')::"regtype" AS "regtype", 
						pg_catalog.format_type("col_attr"."atttypid", "col_attr"."atttypmod") AS "format_type" 
						FROM "information_schema"."columns" 
						LEFT JOIN "pg_catalog"."pg_attribute" AS "col_attr" 
						ON "col_attr"."attname" = "columns"."column_name" AND "col_attr"."attrelid" = ( 
							SELECT "cls"."oid" FROM "pg_catalog"."pg_class" AS "cls" 
							LEFT JOIN "pg_catalog"."pg_namespace" AS "ns" ON "ns"."oid" = "cls"."relnamespace" 
							WHERE "cls"."relname" = "columns"."table_name" AND "ns"."nspname" = "columns"."table_schema" 
						) WHERE ("table_schema" = 'public' AND "table_name" = 'test_table');`,
					Expected: []sql.Row{
						{"id", "", "integer", "integer"},
						{"col1", "", "character varying", "character varying"},
					},
				},
				{
					Query:    `CREATE SCHEMA test_schema;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SET SEARCH_PATH TO test_schema;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE test_table2 (id2 INT);`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT DISTINCT table_schema FROM information_schema.columns order by table_schema;`,
					Expected: []sql.Row{
						{"information_schema"}, {"pg_catalog"}, {"public"}, {"test_schema"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema, table_name, column_name FROM information_schema.columns WHERE table_schema='test_schema';`,
					Expected: []sql.Row{
						{"postgres", "test_schema", "test_table2", "id2"},
					},
				},
				{
					Query: "SELECT * FROM information_schema.columns WHERE table_name='test_table';",
					Expected: []sql.Row{
						{"postgres", "public", "test_table", "id", 1, nil, "NO", "integer", nil, nil, 32, 2, 0, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "postgres", "pg_catalog", "int4", nil, nil, nil, nil, nil, "NO", "NO", nil, nil, nil, nil, nil, "NO", "NEVER", nil, "YES"},
						{"postgres", "public", "test_table", "col1", 2, nil, "YES", "character varying", 255, 1020, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "postgres", "pg_catalog", "varchar", nil, nil, nil, nil, nil, "NO", "NO", nil, nil, nil, nil, nil, "NO", "NEVER", nil, "YES"},
					},
				},
				{
					Skip:  true, // TODO: Don't have complete view information to fill out these rows
					Query: "SELECT * FROM information_schema.columns WHERE table_name='test_view';",
					Expected: []sql.Row{
						{"postgres", "public", "test_view", "id", 1, nil, "YES", "integer", nil, nil, 32, 2, 0, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "postgres", "pg_catalog", "int4", nil, nil, nil, nil, 1, "NO", "NO", nil, nil, nil, nil, nil, "NO", "NEVER", nil, "YES"},
						{"postgres", "public", "test_view", "col1", 2, nil, "YES", "character varying", 255, 1020, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "postgres", "pg_catalog", "varchar", nil, nil, nil, nil, 2, "NO", "NO", nil, nil, nil, nil, nil, "NO", "NEVER", nil, "YES"},
					},
				},
				{
					Query: `SELECT columns.table_name, columns.column_name from "information_schema"."columns" WHERE table_name='test_table';`,
					Expected: []sql.Row{
						{"test_table", "id"},
						{"test_table", "col1"},
					},
				},
				{
					Query:    `CREATE TABLE testnumtypes (id INT PRIMARY KEY, col1 SMALLINT, col2 BIGINT, col3 REAL, col4 DOUBLE PRECISION, col5 NUMERIC, col6 DECIMAL(10, 2), col7 OID, col8 XID);`,
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT column_name, ordinal_position, data_type, udt_name, numeric_precision, numeric_precision_radix, numeric_scale FROM information_schema.columns WHERE table_name='testnumtypes' ORDER BY ordinal_position ASC;",
					Expected: []sql.Row{
						{"id", 1, "integer", "int4", 32, 2, 0},
						{"col1", 2, "smallint", "int2", 16, 2, 0},
						{"col2", 3, "bigint", "int8", 64, 2, 0},
						{"col3", 4, "real", "float4", 24, 2, nil},
						{"col4", 5, "double precision", "float8", 53, 2, nil},
						{"col5", 6, "numeric", "numeric", nil, 10, nil},
						{"col6", 7, "numeric", "numeric", 10, 10, 2},
						{"col7", 8, "oid", "oid", nil, nil, nil},
						{"col8", 9, "xid", "xid", nil, nil, nil},
					},
				},
				{
					Query:    `CREATE TABLE teststringtypes (id INT PRIMARY KEY, col1 CHAR(10), col2 VARCHAR(10), col3 TEXT, col4 "char", col5 CHARACTER, col6 VARCHAR, col7 UUID);`,
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT column_name, ordinal_position, data_type, udt_name, character_maximum_length, character_octet_length FROM information_schema.columns WHERE table_name='teststringtypes' ORDER BY ordinal_position ASC;",
					Expected: []sql.Row{
						{"id", 1, "integer", "int4", nil, nil},
						{"col1", 2, "character", "bpchar", 10, 40},
						{"col2", 3, "character varying", "varchar", 10, 40},
						{"col3", 4, "text", "text", nil, 1073741824},
						{"col4", 5, `"char"`, "char", nil, nil},
						{"col5", 6, "character", "bpchar", 1, 4},
						{"col6", 7, "character varying", "varchar", nil, 1073741824},
						{"col7", 8, "uuid", "uuid", nil, nil},
					},
				},
				{
					Query:    `CREATE TABLE testtimetypes (id INT PRIMARY KEY, col1 DATE, col2 TIME, col3 TIMESTAMP, col4 TIMESTAMPTZ,  col5 TIMETZ);`,
					Expected: []sql.Row{},
				},
				{
					// TODO: Test timestamps with precision when it is implemented
					Query: "SELECT column_name, ordinal_position, data_type, datetime_precision FROM information_schema.columns WHERE table_name='testtimetypes' ORDER BY ordinal_position ASC;",
					Expected: []sql.Row{
						{"id", 1, "integer", nil},
						{"col1", 2, "date", 0},
						{"col2", 3, "time without time zone", 6},
						{"col3", 4, "timestamp without time zone", 6},
						{"col4", 5, "timestamp with time zone", 6},
						{"col5", 6, "time with time zone", 6},
					},
				},
				{
					Query:    `SELECT p.oid AS oid, p.relname AS table_name, n.nspname as table_schema FROM pg_class AS p JOIN pg_namespace AS n ON p.relnamespace=n.oid WHERE n.nspname='public' AND p.relkind='r';`,
					Expected: []sql.Row{{2685403136, "test_table", "public"}},
				},
				{
					Query:    `select col_description(2685403136,ordinal_position) as comment from information_schema.columns limit 1;`,
					Expected: []sql.Row{{""}},
				},
			},
		},
	})
}

func TestInfoSchemaSchemata(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:     "information_schema.schemata",
			Database: "newdb",
			SetUpScript: []string{
				"create schema test_schema",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT catalog_name, schema_name FROM information_schema.schemata order by schema_name;`,
					Expected: []sql.Row{
						{"newdb", "information_schema"},
						{"newdb", "pg_catalog"},
						{"newdb", "public"},
						{"newdb", "test_schema"},
					},
				},
				{
					Query: `SELECT * FROM information_schema.schemata order by schema_name;`,
					Expected: []sql.Row{
						{"newdb", "information_schema", "", nil, nil, nil, nil},
						{"newdb", "pg_catalog", "", nil, nil, nil, nil},
						{"newdb", "public", "", nil, nil, nil, nil},
						{"newdb", "test_schema", "", nil, nil, nil, nil},
					},
				},
			},
		},
	})
}

func TestInfoSchemaTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "information_schema.tables",
			SetUpScript: []string{
				"create table test_table (id int)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM information_schema.tables WHERE table_name='test_table';`,
					Expected: []sql.Row{
						{"postgres", "public", "test_table", "BASE TABLE", nil, nil, nil, nil, nil, "YES", "NO", nil},
					},
				},
				{
					Query: `SELECT DISTINCT table_schema FROM information_schema.tables order by table_schema;`,
					Expected: []sql.Row{
						{"information_schema"}, {"pg_catalog"}, {"public"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema FROM information_schema.tables group by table_catalog, table_schema order by table_schema;`,
					Expected: []sql.Row{
						{"postgres", "information_schema"},
						{"postgres", "pg_catalog"},
						{"postgres", "public"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema, table_name FROM information_schema.tables WHERE table_schema='public';`,
					Expected: []sql.Row{
						{"postgres", "public", "test_table"},
					},
				},
				{
					Query:    `CREATE SCHEMA test_schema;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SET SEARCH_PATH TO test_schema;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE test_table2 (id INT);`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT DISTINCT table_schema FROM information_schema.tables order by table_schema;`,
					Expected: []sql.Row{
						{"information_schema"}, {"pg_catalog"}, {"public"}, {"test_schema"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema FROM information_schema.tables group by table_catalog, table_schema order by table_schema;`,
					Expected: []sql.Row{
						{"postgres", "information_schema"},
						{"postgres", "pg_catalog"},
						{"postgres", "public"},
						{"postgres", "test_schema"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema, table_name FROM information_schema.tables WHERE table_schema='test_schema';`,
					Expected: []sql.Row{
						{"postgres", "test_schema", "test_table2"},
					},
				},
				{
					Query: "SELECT table_catalog, table_schema, table_name, table_type FROM information_schema.tables WHERE table_schema = 'test_schema' ORDER BY table_name;",
					Expected: []sql.Row{
						{"postgres", "test_schema", "test_table2", "BASE TABLE"},
					},
				},
				{
					Query:    `SELECT "table_schema", "table_name", obj_description(('"' || "table_schema" || '"."' || "table_name" || '"')::regclass, 'pg_class') AS table_comment FROM "information_schema"."tables" WHERE ("table_schema" = 'test_schema' AND "table_name" = 'test_table2')`,
					Expected: []sql.Row{{"test_schema", "test_table2", ""}},
				},
				{
					Query:    `CREATE VIEW test_view AS SELECT * FROM test_table2;`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT table_catalog, table_schema, table_name, table_type  FROM information_schema.tables WHERE table_schema='test_schema' OR table_schema='public';`,
					Expected: []sql.Row{
						{"postgres", "public", "test_table", "BASE TABLE"},
						{"postgres", "test_schema", "test_view", "VIEW"},
						{"postgres", "test_schema", "test_table2", "BASE TABLE"},
					},
				},
			},
		},
	})
}

func TestInfoSchemaViews(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "information_schema.views",
			SetUpScript: []string{
				"create table test_table (id int)",
				"create view test_view as select * from test_table",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM information_schema.views order by table_schema;`,
					Expected: []sql.Row{
						{"postgres", "public", "test_view", "SELECT * FROM test_table", "NONE", nil, nil, nil, nil, nil},
					},
				},
				{
					Query: `SELECT DISTINCT table_schema FROM information_schema.views order by table_schema;`,
					Expected: []sql.Row{
						{"public"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema FROM information_schema.views group by table_catalog, table_schema order by table_schema;`,
					Expected: []sql.Row{
						{"postgres", "public"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema, table_name FROM information_schema.views WHERE table_schema='public';`,
					Expected: []sql.Row{
						{"postgres", "public", "test_view"},
					},
				},
				{
					Query:    `CREATE SCHEMA test_schema;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SET SEARCH_PATH TO test_schema;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TABLE test_table2 (id int);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE VIEW test_view2 as select * from test_table2;`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT DISTINCT table_schema FROM information_schema.views order by table_schema;`,
					Expected: []sql.Row{
						{"public"},
						{"test_schema"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema FROM information_schema.views group by table_catalog, table_schema order by table_schema;`,
					Expected: []sql.Row{
						{"postgres", "public"},
						{"postgres", "test_schema"},
					},
				},
				{
					Query: `SELECT table_catalog, table_schema, table_name FROM information_schema.views WHERE table_schema='test_schema';`,
					Expected: []sql.Row{
						{"postgres", "test_schema", "test_view2"},
					},
				},
				{
					Query: "SELECT table_catalog, table_schema, table_name, view_definition FROM information_schema.views WHERE table_schema = 'test_schema' ORDER BY table_name;",
					Expected: []sql.Row{
						{"postgres", "test_schema", "test_view2", "SELECT * FROM test_table2"},
					},
				},
			},
		},
	})
}
