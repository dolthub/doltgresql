package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestPgAggregate(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_aggregate",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_aggregate";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_aggregate";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_aggregate";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT aggfnoid FROM PG_catalog.pg_AGGREGATE ORDER BY aggfnoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgAm(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_am",
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM "pg_catalog"."pg_am";`,
					Expected: []sql.Row{
						{2, "heap", "heap_tableam_handler", "t"},
						{403, "btree", "bthandler", "i"},
						{405, "hash", "hashhandler", "i"},
						{783, "gist", "gisthandler", "i"},
						{2742, "gin", "ginhandler", "i"},
						{4000, "spgist", "spghandler", "i"},
						{3580, "brin", "brinhandler", "i"},
					},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_am";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_am";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT amname FROM PG_catalog.pg_AM ORDER BY amname;",
					Expected: []sql.Row{{"brin"}, {"btree"}, {"gin"}, {"gist"}, {"hash"}, {"heap"}, {"spgist"}},
				},
			},
		},
	})
}

func TestPgAmop(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_amop",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_amop";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_amop";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_amop";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_AMOP ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgAmproc(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_amproc",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_amproc";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_amproc";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_amproc";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_AMPROC ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgAttribute(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_attribute",
			SetUpScript: []string{
				`CREATE SCHEMA testschema;`,
				`SET search_path TO testschema;`,
				`CREATE TABLE test (pk INT primary key, v1 TEXT DEFAULT 'hey');`,
				`CREATE TABLE test2 (pk INT primary key, pktesting INT REFERENCES test(pk), v1 TEXT);`,

				// Should show attributes for all schemas
				`CREATE SCHEMA testschema2;`,
				`SET search_path TO testschema2;`,
			},
			Assertions: []ScriptTestAssertion{
				// {
				// 	Query:    `SELECT * FROM "pg_catalog"."pg_attribute" WHERE attname='pk' AND attrelid='testschema.test'::regclass;`,
				// 	Expected: []sql.Row{{2502341994, "pk", 23, 0, 1, -1, -1, 0, "f", "i", "p", "", "t", "f", "f", "", "", "f", "t", 0, -1, 0, nil, nil, nil, nil}},
				// },
				// {
				// 	Query:    `SELECT * FROM "pg_catalog"."pg_attribute" WHERE attname='v1' AND attrelid='testschema.test'::regclass;`,
				// 	Expected: []sql.Row{{2502341994, "v1", 25, 0, 2, -1, -1, 0, "f", "i", "p", "", "f", "t", "f", "", "", "f", "t", 0, -1, 0, nil, nil, nil, nil}},
				// },
				// { // Different cases and quoted, so it fails
				// 	Query:       `SELECT * FROM "PG_catalog"."pg_attribute";`,
				// 	ExpectedErr: "not",
				// },
				// { // Different cases and quoted, so it fails
				// 	Query:       `SELECT * FROM "pg_catalog"."PG_attribute";`,
				// 	ExpectedErr: "not",
				// },
				// { // Different cases but non-quoted, so it works
				// 	Query: "SELECT attname FROM PG_catalog.pg_ATTRIBUTE ORDER BY attname LIMIT 3;",
				// 	Expected: []sql.Row{
				// 		{"ACTION_CONDITION"},
				// 		{"ACTION_ORDER"},
				// 		{"ACTION_ORIENTATION"},
				// 	},
				// },
				// 		{
				// 			Query: `EXPLAIN SELECT attname FROM "pg_catalog"."pg_attribute" a
				// JOIN "pg_catalog"."pg_class" c ON a.attrelid = c.oid
				//            WHERE c.relname = 'test';`,
				// 			Expected: []sql.Row{
				// 				{"pk"},
				// 				{"v1"},
				// 			},
				// 		},
				{
					Query: `SELECT attname FROM "pg_catalog"."pg_attribute" a
    JOIN "pg_catalog"."pg_class" c ON a.attrelid = c.oid
               WHERE c.relname = 'test';`,
					Expected: []sql.Row{
						{"pk"},
						{"v1"},
					},
				},
				{
					Query: `SELECT count(*) FROM pg_attribute as a1
				WHERE a1.attrelid = 0 OR a1.atttypid = 0 OR a1.attnum = 0 OR
				a1.attcacheoff != -1 OR a1.attinhcount < 0 OR
    		(a1.attinhcount = 0 AND NOT a1.attislocal);`,
					Expected: []sql.Row{{0}},
				},
				{
					// TODO: Even with the caching added to prevent having to regenerate pg_catalog table data
					//       multiple times within the same query, this massive query still times out. The problem
					//       is that this query joins over 7 tables and without a way to do index lookups into the
					//       table data, we end up iterating over the results over and over.
					Skip: true,
					Query: `SELECT "con"."conname" AS "constraint_name", 
       "con"."nspname" AS "table_schema", 
       "con"."relname" AS "table_name", 
       "att2"."attname" AS "column_name", 
       "ns"."nspname" AS "referenced_table_schema", 
       "cl"."relname" AS "referenced_table_name", 
       "att"."attname" AS "referenced_column_name", 
       "con"."confdeltype" AS "on_delete", 
       "con"."confupdtype" AS "on_update", 
       "con"."condeferrable" AS "deferrable", 
       "con"."condeferred" AS "deferred"
FROM 
    ( SELECT UNNEST ("con1"."conkey") AS "parent", 
              UNNEST ("con1"."confkey") AS "child", 
              "con1"."confrelid", 
              "con1"."conrelid", 
              "con1"."conname", 
              "con1"."contype", 
              "ns"."nspname", 
              "cl"."relname", 
              "con1"."condeferrable", 
              CASE 
                  WHEN "con1"."condeferred" THEN 'INITIALLY DEFERRED' 
                  ELSE 'INITIALLY IMMEDIATE' 
                  END as condeferred, 
           CASE "con1"."confdeltype" 
               WHEN 'a' THEN 'NO ACTION' 
               WHEN 'r' THEN 'RESTRICT' 
               WHEN 'c' THEN 'CASCADE' 
               WHEN 'n' THEN 'SET NULL' 
               WHEN 'd' THEN 'SET DEFAULT' 
               END as "confdeltype", 
           CASE "con1"."confupdtype" 
               WHEN 'a' THEN 'NO ACTION' 
               WHEN 'r' THEN 'RESTRICT' 
               WHEN 'c' THEN 'CASCADE' 
               WHEN 'n' THEN 'SET NULL' 
               WHEN 'd' THEN 'SET DEFAULT' 
               END as "confupdtype" 
       FROM "pg_class" "cl" 
           INNER JOIN "pg_namespace" "ns" ON "cl"."relnamespace" = "ns"."oid" 
           INNER JOIN "pg_constraint" "con1" ON "con1"."conrelid" = "cl"."oid" 
       WHERE "con1"."contype" = 'f' 
         AND (("ns"."nspname" = 'testschema' AND "cl"."relname" = 'test2')) ) "con" 
    INNER JOIN "pg_attribute" "att" ON "att"."attrelid" = "con"."confrelid" AND "att"."attnum" = "con"."child"
    INNER JOIN "pg_class" "cl" ON "cl"."oid" = "con"."confrelid"  AND "cl"."relispartition" = 'f'
    INNER JOIN "pg_namespace" "ns" ON "cl"."relnamespace" = "ns"."oid" 
    INNER JOIN "pg_attribute" "att2" ON "att2"."attrelid" = "con"."conrelid" AND "att2"."attnum" = "con"."parent";`,
					Expected: []sql.Row{
						{"test2_pktesting_fkey", "testschema", "test2", "pktesting", "testschema", "test", "pk", "NO ACTION", "NO ACTION", "f", "INITIALLY IMMEDIATE"},
					},
				},
			},
		},
	})
}

func TestPgAttrdef(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_attrdef",
			SetUpScript: []string{
				`CREATE SCHEMA testschema;`,
				`SET search_path TO testschema;`,
				`CREATE TABLE test (pk INT primary key, v1 TEXT DEFAULT 'hey');`,

				// Should show attributes for all schemas
				`CREATE SCHEMA testschema2;`,
				`SET search_path TO testschema2;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM "pg_catalog"."pg_attrdef" WHERE adrelid='testschema.test'::regclass;`,
					Expected: []sql.Row{
						{597021512, 2502341994, 2, nil},
					},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_attrdef";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_attrdef";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query: "SELECT oid FROM PG_catalog.pg_ATTRDEF ORDER BY oid;",
					Expected: []sql.Row{
						{597021512},
					},
				},
			},
		},
	})
}

func TestPgAuthMembers(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_auth_members",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_auth_members";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_auth_members";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_auth_members";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT member FROM PG_catalog.pg_AUTH_MEMBERS ORDER BY member;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgAuthid(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_authid",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_authid";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_authid";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_authid";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT rolname FROM PG_catalog.pg_AUTHID ORDER BY rolname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgAvailableExtensionVersions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_available_extension_versions",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_available_extension_versions";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_available_extension_versions";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_available_extension_versions";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_AVAILABLE_EXTENSION_VERSIONS ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgAvailableExtensions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_available_extensions",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_available_extensions";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_available_extensions";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_available_extensions";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_AVAILABLE_EXTENSIONS ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgBackendMemoryContexts(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_backend_memory_contexts",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_backend_memory_contexts";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_backend_memory_contexts";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_backend_memory_contexts";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_BACKEND_MEMORY_CONTEXTS ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgCast(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_cast",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_cast";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_cast";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_cast";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_CAST ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgClass(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_class",
			SetUpScript: []string{
				`CREATE SCHEMA testschema;`,
				`SET search_path TO testschema;`,
				`CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);`,
				`CREATE VIEW testview AS SELECT * FROM testing LIMIT 1;`,

				// Should show classes for all schemas
				`CREATE SCHEMA testschema2;`,
				`SET search_path TO testschema2;`,
			},
			Assertions: []ScriptTestAssertion{
				// Table
				{
					Query: `SELECT * FROM "pg_catalog"."pg_class" WHERE relname='testing' order by 1;`,
					Expected: []sql.Row{
						{3120782595, "testing", 2638679668, 0, 0, 0, 2, 0, 0, 0, float32(0), 0, 0, "t", "f", "p", "r", 0, 0, "f", "f", "f", "f", "f", "t", "d", "f", 0, 0, 0, nil, nil, nil},
					},
				},
				// Index
				{
					Query: `SELECT * FROM "pg_catalog"."pg_class" WHERE relname='testing_pkey';`,
					Expected: []sql.Row{
						{1067629180, "testing_pkey", 2638679668, 0, 0, 0, 403, 0, 0, 0, float32(0), 0, 0, "f", "f", "p", "i", 0, 0, "f", "f", "f", "f", "f", "t", "d", "f", 0, 0, 0, nil, nil, nil},
					},
				},
				// View
				{
					Query: `SELECT * FROM "pg_catalog"."pg_class" WHERE relname='testview';`,
					Expected: []sql.Row{
						{887295443, "testview", 2638679668, 0, 0, 0, 0, 0, 0, 0, float32(0), 0, 0, "f", "f", "p", "v", 0, 0, "f", "f", "f", "f", "f", "t", "d", "f", 0, 0, 0, nil, nil, nil},
					},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_class";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_class";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query: "SELECT relname FROM PG_catalog.pg_CLASS where relnamespace not in (select oid from pg_namespace where nspname = 'dolt') ORDER BY relname ASC LIMIT 3;",
					Expected: []sql.Row{
						{"administrable_role_authorizations"},
						{"applicable_roles"},
						{"character_sets"},
					},
				},
				{
					Query: "SELECT relname from pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid  WHERE n.nspname = 'testschema' and left(relname, 5) <> 'dolt_' ORDER BY relname;",
					Expected: []sql.Row{
						{"testing"},
						{"testing_pkey"},
						{"testview"},
						{"v1"},
					},
				},
				{
					Query: "SELECT relname from pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid  WHERE n.nspname = 'pg_catalog' LIMIT 3;",
					Expected: []sql.Row{
						{"pg_aggregate"},
						{"pg_am"},
						{"pg_amop"},
					},
				},
				{
					Query: `SELECT relname FROM "pg_class" WHERE relname='testing';`,
					Expected: []sql.Row{
						{"testing"},
					},
				},
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_class" WHERE oid=1234`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "pg_class with regclass",
			SetUpScript: []string{
				`CREATE SCHEMA testschema;`,
				`SET search_path TO testschema;`,
				`CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);`,
				`CREATE VIEW testview AS SELECT * FROM testing LIMIT 1;`,
				`CREATE SCHEMA testschema2;`,
				`SET search_path TO testschema2;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `SELECT * FROM "pg_catalog"."pg_class" WHERE oid='testing'::regclass;`,
					ExpectedErr: "does not exist",
				},
				{
					Query: `SELECT * FROM "pg_catalog"."pg_class" WHERE oid='testschema.testing'::regclass;`,
					Expected: []sql.Row{
						{3120782595, "testing", 2638679668, 0, 0, 0, 2, 0, 0, 0, float32(0), 0, 0, "t", "f", "p", "r", 0, 0, "f", "f", "f", "f", "f", "t", "d", "f", 0, 0, 0, nil, nil, nil},
					},
				},
				{
					Query: `SELECT * FROM "pg_catalog"."pg_class" WHERE oid='testschema.testing_pkey'::regclass;`,
					Expected: []sql.Row{
						{1067629180, "testing_pkey", 2638679668, 0, 0, 0, 403, 0, 0, 0, float32(0), 0, 0, "f", "f", "p", "i", 0, 0, "f", "f", "f", "f", "f", "t", "d", "f", 0, 0, 0, nil, nil, nil},
					},
				},
				{
					Query: `SELECT * FROM "pg_catalog"."pg_class" WHERE oid='testschema.testview'::regclass;`,
					Expected: []sql.Row{
						{887295443, "testview", 2638679668, 0, 0, 0, 0, 0, 0, 0, float32(0), 0, 0, "f", "f", "p", "v", 0, 0, "f", "f", "f", "f", "f", "t", "d", "f", 0, 0, 0, nil, nil, nil},
					},
				},
			},
		},
		{
			Name: "pg_class joined with other pg_catalog tables to retrieve indexes",
			SetUpScript: []string{
				`CREATE TABLE foo (a INTEGER NOT NULL PRIMARY KEY, b INTEGER NULL);`,
				`CREATE INDEX ON foo ( b ASC ) NULLS NOT DISTINCT;`,
				`CREATE INDEX ON foo ( b ASC , a DESC ) NULLS NOT DISTINCT;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// TODO: Now that catalog data is cached for each query, this query no longer iterates the database
					//       100k times, and this query executes in a couple seconds. This is still slow and should
					//       be improved with lookup index support now that we have cached data available.
					Query: `SELECT ix.relname AS index_name, upper(am.amname) AS index_algorithm FROM pg_index i 
JOIN pg_class t ON t.oid = i.indrelid 
JOIN pg_class ix ON ix.oid = i.indexrelid 
JOIN pg_namespace n ON t.relnamespace = n.oid 
JOIN pg_am AS am ON ix.relam = am.oid WHERE t.relname = 'foo' AND n.nspname = 'public';`,
					Expected: []sql.Row{{"foo_pkey", "BTREE"}, {"b", "BTREE"}, {"b_2", "BTREE"}}, // TODO: should follow Postgres index naming convention: "foo_pkey", "foo_b_idx", "foo_b_a_idx"
				},
			},
		},
	})
}

func TestPgCollation(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_collation",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_collation";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_collation";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_collation";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT collname FROM PG_catalog.pg_COLLATION ORDER BY collname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgConfig(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_config",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_config";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_config";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_config";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_CONFIG ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgConstraint(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_constraint",
			SetUpScript: []string{
				`CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);`,
				`CREATE TABLE testing2 (pk INT primary key, pktesting INT REFERENCES testing(pk), v1 TEXT);`,
				// TODO: Uncomment when check constraints supported
				// `ALTER TABLE testing2 ADD CONSTRAINT v1_check CHECK (v1 != '')`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM "pg_catalog"."pg_constraint" WHERE conrelid='testing2'::regclass OR conrelid='testing'::regclass;`,
					Expected: []sql.Row{
						{3757635986, "testing_pkey", 2200, "p", "f", "f", "t", 2147906242, 0, 3757635986, 0, 0, "", "", "", "t", 0, "t", "{1}", nil, nil, nil, nil, nil, nil, nil},
						// TODO: postgres names this index testing_v1_key
						{3050361446, "v1", 2200, "u", "f", "f", "t", 2147906242, 0, 3050361446, 0, 0, "", "", "", "t", 0, "t", "{2}", nil, nil, nil, nil, nil, nil, nil},
						{1719906648, "testing2_pktesting_fkey", 2200, "f", "f", "f", "t", 2694106299, 0, 1719906648, 0, 2147906242, "a", "a", "s", "t", 0, "t", "{0}", "{1}", nil, nil, nil, nil, nil, nil},
						{2068729390, "testing2_pkey", 2200, "p", "f", "f", "t", 2694106299, 0, 2068729390, 0, 0, "", "", "", "t", 0, "t", "{1}", nil, nil, nil, nil, nil, nil, nil},
					},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_constraint";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_constraint";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query: "SELECT conname FROM PG_catalog.pg_CONSTRAINT ORDER BY conname;",
					Expected: []sql.Row{
						{"testing2_pkey"},
						{"testing2_pktesting_fkey"},
						{"testing_pkey"},
						{"v1"},
					},
				},
				{
					Query: "SELECT co.oid, co.conname, co.conrelid, cl.relname FROM pg_catalog.pg_constraint co JOIN pg_catalog.pg_class cl ON co.conrelid = cl.oid WHERE cl.relname = 'testing2';",
					Expected: []sql.Row{
						{2068729390, "testing2_pkey", 2694106299, "testing2"},
						{1719906648, "testing2_pktesting_fkey", 2694106299, "testing2"},
					},
				},
			},
		},
	})
}

func TestPgConversion(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_conversion",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_conversion";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_conversion";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_conversion";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT conname FROM PG_catalog.pg_CONVERSION ORDER BY conname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgCursors(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_cursors",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_cursors";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_cursors";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_cursors";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_CURSORS ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgDatabase(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_database",
			SetUpScript: []string{
				`CREATE DATABASE test;`,
				`CREATE TABLE test (pk INT primary key, v1 INT);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT datname FROM "pg_catalog"."pg_database";`,
					Expected: []sql.Row{
						{"postgres"},
						{"test"},
					},
				},
				{
					Query: `SELECT oid, datname FROM "pg_catalog"."pg_database" ORDER BY datname DESC;`,
					Expected: []sql.Row{
						{258611842, "test"},
						{5, "postgres"},
					},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_database";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_database";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query: "SELECT oid, datname FROM PG_catalog.pg_DATABASE ORDER BY datname ASC;",
					Expected: []sql.Row{
						{5, "postgres"},
						{258611842, "test"},
					},
				},
				{
					Query: "SELECT * FROM pg_catalog.pg_database WHERE datname='test';",
					Expected: []sql.Row{
						{258611842, "test", 0, 6, "i", "f", "t", -1, 0, 0, 0, "", "", nil, "", nil, nil},
					},
				},
			},
		},
	})
}

func TestPgDbRoleSetting(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_db_role_setting",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_db_role_setting";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_db_role_setting";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_db_role_setting";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT setdatabase FROM PG_catalog.pg_DB_ROLE_SETTING ORDER BY setdatabase;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgDefaultAcl(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_default_acl",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_default_acl";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_default_acl";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_default_acl";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_DEFAULT_ACL ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgDepend(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_depend",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_depend";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_depend";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_depend";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT classid FROM PG_catalog.pg_DEPEND ORDER BY classid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgDescription(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_description",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_description";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_description";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_description";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT objoid FROM PG_catalog.pg_DESCRIPTION ORDER BY objoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgEnum(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_enum",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_enum";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_enum";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_enum";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT enumlabel FROM PG_catalog.pg_ENUM ORDER BY enumlabel;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgEventTrigger(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_event_trigger",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_event_trigger";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_event_trigger";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_event_trigger";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT evtname FROM PG_catalog.pg_EVENT_TRIGGER ORDER BY evtname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgExtension(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_extension",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_extension";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_extension";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_extension";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT extname FROM PG_catalog.pg_EXTENSION ORDER BY extname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgFileSettings(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_file_settings",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_file_settings";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_file_settings";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_file_settings";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_FILE_SETTINGS ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgForeignDataWrapper(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_foreign_data_wrapper",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_foreign_data_wrapper";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_foreign_data_wrapper";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_foreign_data_wrapper";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT fdwname FROM PG_catalog.pg_FOREIGN_DATA_WRAPPER ORDER BY fdwname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgForeignServer(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_foreign_server",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_foreign_server";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_foreign_server";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_foreign_server";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT srvname FROM PG_catalog.pg_FOREIGN_SERVER ORDER BY srvname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgForeignTable(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_foreign_table",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_foreign_table";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_foreign_table";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_foreign_table";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT ftrelid FROM PG_catalog.pg_FOREIGN_TABLE ORDER BY ftrelid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgGroup(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_group",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_group";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_group";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_group";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT groname FROM PG_catalog.pg_GROUP ORDER BY groname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgHbaFileRules(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_hba_file_rules",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_hba_file_rules";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_hba_file_rules";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_hba_file_rules";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT line_number FROM PG_catalog.pg_HBA_FILE_RULES ORDER BY line_number;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgIdentFileMappings(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_ident_file_mappings",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_ident_file_mappings";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_ident_file_mappings";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_ident_file_mappings";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT line_number FROM PG_catalog.pg_IDENT_FILE_MAPPINGS ORDER BY line_number;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgIndex(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_index",
			SetUpScript: []string{
				`CREATE SCHEMA testschema;`,
				`SET search_path TO testschema;`,
				`CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);`,
				`CREATE TABLE testing2 (pk INT, v1 INT, PRIMARY KEY (pk, v1));`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT i.* from pg_class c " +
						"JOIN pg_index i ON c.oid = i.indexrelid " +
						"JOIN pg_namespace n ON c.relnamespace = n.oid " +
						"WHERE n.nspname = 'testschema' and left(c.relname, 5) <> 'dolt_' " +
						"ORDER BY 1;",
					Expected: []sql.Row{
						{1067629180, 3120782595, 1, 0, "t", "f", "t", "f", "f", "f", "t", "f", "t", "t", "f", "{1}", "{}", "{}", "0", nil, nil},
						{1322775662, 3120782595, 1, 0, "t", "f", "f", "f", "f", "f", "t", "f", "t", "t", "f", "{2}", "{}", "{}", "0", nil, nil},
						{3185790121, 1784425749, 2, 0, "t", "f", "t", "f", "f", "f", "t", "f", "t", "t", "f", "{1,2}", "{}", "{}", "0", nil, nil},
					},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_index";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_index";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query: "SELECT i.indexrelid from pg_class c " +
						"JOIN PG_catalog.pg_INDEX i ON c.oid = i.indexrelid " +
						"JOIN pg_namespace n ON c.relnamespace = n.oid " +
						"WHERE n.nspname = 'testschema' and left(c.relname, 5) <> 'dolt_' " +
						"ORDER BY 1;",
					Expected: []sql.Row{{1067629180}, {1322775662}, {3185790121}},
				},
				{
					Query: "SELECT i.indexrelid, i.indrelid, c.relname, t.relname  FROM pg_catalog.pg_index i " +
						"JOIN pg_catalog.pg_class c ON i.indexrelid = c.oid " +
						"JOIN pg_catalog.pg_class t ON i.indrelid = t.oid " +
						"JOIN pg_namespace n ON t.relnamespace = n.oid " +
						"WHERE n.nspname = 'testschema' and left(c.relname, 5) <> 'dolt_'",
					Expected: []sql.Row{
						{1067629180, 3120782595, "testing_pkey", "testing"},
						{1322775662, 3120782595, "v1", "testing"},
						{3185790121, 1784425749, "testing2_pkey", "testing2"},
					},
				},
			},
		},
	})
}

func TestPgIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_indexes",
			SetUpScript: []string{
				"CREATE SCHEMA testschema;",
				"SET search_path TO testschema;",
				`CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);`,
				`CREATE TABLE testing2 (pk INT, v1 INT, PRIMARY KEY (pk, v1));`,
				"CREATE INDEX my_index ON testing2(v1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM "pg_catalog"."pg_indexes" where schemaname = 'testschema';`,
					Expected: []sql.Row{
						{"testschema", "testing", "testing_pkey", "", "CREATE UNIQUE INDEX testing_pkey ON testschema.testing USING btree (pk)"},
						{"testschema", "testing", "v1", "", "CREATE UNIQUE INDEX v1 ON testschema.testing USING btree (v1)"},
						{"testschema", "testing2", "testing2_pkey", "", "CREATE UNIQUE INDEX testing2_pkey ON testschema.testing2 USING btree (pk, v1)"},
						{"testschema", "testing2", "my_index", "", "CREATE INDEX my_index ON testschema.testing2 USING btree (v1)"},
					},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT indexname FROM PG_catalog.pg_INDEXES where schemaname='testschema' ORDER BY indexname;",
					Expected: []sql.Row{{"my_index"}, {"testing2_pkey"}, {"testing_pkey"}, {"v1"}},
				},
			},
		},
	})
}

func TestPgInherits(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_inherits",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_inherits";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_inherits";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_inherits";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT inhrelid FROM PG_catalog.pg_INHERITS ORDER BY inhrelid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgInitPrivs(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_init_privs",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_init_privs";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_init_privs";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_init_privs";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT objoid FROM PG_catalog.pg_INIT_PRIVS ORDER BY objoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgLanguage(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_language",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_language";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_language";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_language";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT lanname FROM PG_catalog.pg_LANGUAGE ORDER BY lanname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgLargeobject(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_largeobject",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_largeobject";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_largeobject";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_largeobject";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT loid FROM PG_catalog.pg_LARGEOBJECT ORDER BY loid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgLargeobjectMetadata(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_largeobject_metadata",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_largeobject_metadata";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_largeobject_metadata";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_largeobject_metadata";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_LARGEOBJECT_METADATA ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgLocks(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_locks",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_locks";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_locks";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_locks";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT objid FROM PG_catalog.pg_LOCKS ORDER BY objid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgMatviews(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_matviews",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_matviews";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_matviews";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_matviews";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT matviewname FROM PG_catalog.pg_MATVIEWS ORDER BY matviewname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgNamespace(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_namespace",
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM "pg_catalog"."pg_namespace" ORDER BY nspname;`,
					Expected: []sql.Row{
						{1882653564, "dolt", 0, nil},
						{13183, "information_schema", 0, nil},
						{11, "pg_catalog", 0, nil},
						{2200, "public", 0, nil},
					},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_namespace";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_namespace";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query: "SELECT nspname FROM PG_catalog.pg_NAMESPACE ORDER BY nspname;",
					Expected: []sql.Row{
						{"dolt"},
						{"information_schema"},
						{"pg_catalog"},
						{"public"},
					},
				},
				{
					Query:    "CREATE SCHEMA testschema;",
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT * FROM "pg_catalog"."pg_namespace" ORDER BY nspname;`,
					Expected: []sql.Row{
						{1882653564, "dolt", 0, nil},
						{13183, "information_schema", 0, nil},
						{11, "pg_catalog", 0, nil},
						{2200, "public", 0, nil},
						{2638679668, "testschema", 0, nil},
					},
				},
			},
		},
	})
}

func TestPgOpclass(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_opclass",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_opclass";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_opclass";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_opclass";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT opcname FROM PG_catalog.pg_OPCLASS ORDER BY opcname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgOperator(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_operator",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_operator";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_operator";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_operator";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oprname FROM PG_catalog.pg_OPERATOR ORDER BY oprname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgOpfamily(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_opfamily",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_opfamily";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_opfamily";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_opfamily";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT opfname FROM PG_catalog.pg_OPFAMILY ORDER BY opfname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgParameterAcl(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_parameter_acl",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_parameter_acl";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_parameter_acl";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_parameter_acl";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT parname FROM PG_catalog.pg_PARAMETER_ACL ORDER BY parname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPartitionedTable(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_partitioned_table",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_partitioned_table";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_partitioned_table";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_partitioned_table";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT partrelid FROM PG_catalog.pg_PARTITIONED_TABLE ORDER BY partrelid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPolicies(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_policies",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_policies";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_policies";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_policies";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT policyname FROM PG_catalog.pg_POLICIES ORDER BY policyname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPolicy(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_policy",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_policy";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_policy";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_policy";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT polname FROM PG_catalog.pg_POLICY ORDER BY polname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPreparedStatements(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_prepared_statements",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_prepared_statements";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_prepared_statements";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_prepared_statements";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_PREPARED_STATEMENTS ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPreparedXacts(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_prepared_xacts",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_prepared_xacts";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_prepared_xacts";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_prepared_xacts";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT gid FROM PG_catalog.pg_PREPARED_XACTS ORDER BY gid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgProc(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_proc",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_proc";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_proc";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_proc";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT proname FROM PG_catalog.pg_PROC ORDER BY proname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPublication(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_publication",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_publication";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_publication";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_publication";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pubname FROM PG_catalog.pg_PUBLICATION ORDER BY pubname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPublicationNamespace(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_publication_namespace",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_publication_namespace";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_publication_namespace";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_publication_namespace";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_PUBLICATION_NAMESPACE ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPublicationRel(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_publication_rel",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_publication_rel";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_publication_rel";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_publication_rel";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_PUBLICATION_REL ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPublicationTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_publication_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_publication_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_publication_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_publication_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pubname FROM PG_catalog.pg_PUBLICATION_TABLES ORDER BY pubname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgRange(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_range",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_range";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_range";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_range";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT rngtypid FROM PG_catalog.pg_RANGE ORDER BY rngtypid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgReplicationOrigin(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_replication_origin",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_replication_origin";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_replication_origin";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_replication_origin";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT roname FROM PG_catalog.pg_REPLICATION_ORIGIN ORDER BY roname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgReplicationOriginStatus(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_replication_origin_status",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_replication_origin_status";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_replication_origin_status";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_replication_origin_status";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT local_id FROM PG_catalog.pg_REPLICATION_ORIGIN_STATUS ORDER BY local_id;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgReplicationSlots(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_replication_slot",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_replication_slots";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_replication_slots";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_replication_slots";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT slot_name FROM PG_catalog.pg_REPLICATION_SLOTS ORDER BY slot_name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgRewrite(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_rewrite",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_rewrite";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_rewrite";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_rewrite";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_REWRITE ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgRoles(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_roles",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_roles";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_roles";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_roles";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT rolname FROM PG_catalog.pg_ROLES ORDER BY rolname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgRules(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_rules",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_rules";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_rules";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_rules";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT rulename FROM PG_catalog.pg_RULES ORDER BY rulename;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgSeclabel(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_seclabel",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_seclabel";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_seclabel";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_seclabel";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT objoid FROM PG_catalog.pg_SECLABEL ORDER BY objoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgSeclabels(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_seclabels",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_seclabels";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_seclabels";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_seclabels";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT objoid FROM PG_catalog.pg_SECLABELS ORDER BY objoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgSequences(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_sequences",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_sequences";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT sequencename FROM PG_catalog.pg_SEQUENCES ORDER BY sequencename;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgSettings(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_settings",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_settings";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_settings";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_settings";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_SETTINGS ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgShadow(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_shadow",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_shadow";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_shadow";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_shadow";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT usename FROM PG_catalog.pg_SHADOW ORDER BY usename;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgShdepend(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_shdepend",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_shdepend";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_shdepend";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_shdepend";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT dbid FROM PG_catalog.pg_SHDEPEND ORDER BY dbid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgShdescription(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_shdescription",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_shdescription";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_shdescription";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_shdescription";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT objoid FROM PG_catalog.pg_SHDESCRIPTION ORDER BY objoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgShmemAllocations(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_shmem_allocations",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_shmem_allocations";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_shmem_allocations";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_shmem_allocations";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_SHMEM_ALLOCATIONS ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgShseclabel(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_shseclabel",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_shseclabel";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_shseclabel";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_shseclabel";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT objoid FROM PG_catalog.pg_SHSECLABEL ORDER BY objoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatActivity(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_activity",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_activity";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_activity";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_activity";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT datname FROM PG_catalog.pg_STAT_ACTIVITY ORDER BY datname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatAllIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_all_indexes",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_all_indexes";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_all_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_all_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relname FROM PG_catalog.pg_STAT_ALL_INDEXES ORDER BY relname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatAllTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_all_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_all_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_all_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_all_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relname FROM PG_catalog.pg_STAT_ALL_TABLES ORDER BY relname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatArchiver(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_archiver",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_archiver";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_archiver";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_archiver";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT archived_count FROM PG_catalog.pg_STAT_ARCHIVER ORDER BY archived_count;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatBgwriter(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_bgwriter",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_bgwriter";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_bgwriter";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_bgwriter";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT checkpoints_timed FROM PG_catalog.pg_STAT_BGWRITER ORDER BY checkpoints_timed;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatDatabase(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_database",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_database";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_database";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_database";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT datname FROM PG_catalog.pg_STAT_DATABASE ORDER BY datname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatDatabaseConflicts(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_database_conflicts",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_database_conflicts";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_database_conflicts";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_database_conflicts";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT datname FROM PG_catalog.pg_STAT_DATABASE_CONFLICTS ORDER BY datname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatGssapi(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_gssapi",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_gssapi";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_gssapi";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_gssapi";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_GSSAPI ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatProgressAnalyze(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_progress_analyze",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_progress_analyze";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_progress_analyze";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_progress_analyze";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT datname FROM PG_catalog.pg_STAT_PROGRESS_ANALYZE ORDER BY datname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatProgressBasebackup(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_progress_basebackup",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_progress_basebackup";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_progress_basebackup";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_progress_basebackup";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_PROGRESS_BASEBACKUP ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatProgressCluster(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_progress_cluster",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_progress_cluster";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_progress_cluster";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_progress_cluster";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_PROGRESS_CLUSTER ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatProgressCopy(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_progress_copy",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_progress_copy";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_progress_copy";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_progress_copy";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_PROGRESS_COPY ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatProgressCreateIndex(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_progress_create_index",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_progress_create_index";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_progress_create_index";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_progress_create_index";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_PROGRESS_CREATE_INDEX ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatProgressVacuum(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_progress_vacuum",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_progress_vacuum";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_progress_vacuum";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_progress_vacuum";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_PROGRESS_VACUUM ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatRecoveryPrefetch(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_recovery_prefetch",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_recovery_prefetch";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_recovery_prefetch";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_recovery_prefetch";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT hit FROM PG_catalog.pg_STAT_RECOVERY_PREFETCH ORDER BY hit;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatReplication(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_replication",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_replication";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_replication";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_replication";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_REPLICATION ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatReplicationSlots(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_replication_slots",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_replication_slots";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_replication_slots";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_replication_slots";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT slot_name FROM PG_catalog.pg_STAT_REPLICATION_SLOTS ORDER BY slot_name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatSlru(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_slru",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_slru";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_slru";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_slru";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_STAT_SLRU ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatSsl(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_ssl",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_ssl";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_ssl";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_ssl";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_SSL ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatSubscription(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_subscription",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_subscription";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_subscription";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_subscription";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT subid FROM PG_catalog.pg_STAT_SUBSCRIPTION ORDER BY subid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatSubscriptionStats(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_subscription_stats",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_subscription_stats";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_subscription_stats";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_subscription_stats";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT subid FROM PG_catalog.pg_STAT_SUBSCRIPTION_STATS ORDER BY subid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatSysIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_sys_indexes",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_sys_indexes";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_sys_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_sys_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STAT_SYS_INDEXES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatSysTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_sys_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_sys_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_sys_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_sys_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STAT_SYS_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatUserFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_user_functions",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_user_functions";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_user_functions";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_user_functions";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT funcid FROM PG_catalog.pg_STAT_USER_FUNCTIONS ORDER BY funcid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatUserIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_user_indexes",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_user_indexes";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_user_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_user_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STAT_USER_INDEXES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatUserTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_user_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_user_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_user_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_user_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STAT_USER_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatWal(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_wal",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_wal";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_wal";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_wal";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT wal_records FROM PG_catalog.pg_STAT_WAL ORDER BY wal_records;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatWalReceiver(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_wal_receiver",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_wal_receiver";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_wal_receiver";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_wal_receiver";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT pid FROM PG_catalog.pg_STAT_WAL_RECEIVER ORDER BY pid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatXactAllTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_xact_all_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_xact_all_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_xact_all_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_xact_all_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STAT_XACT_ALL_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatXactSysTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_xact_sys_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_xact_sys_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_xact_sys_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_xact_sys_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STAT_XACT_SYS_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatXactUserFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_xact_user_functions",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_xact_user_functions";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_xact_user_functions";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_xact_user_functions";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT funcid FROM PG_catalog.pg_STAT_XACT_USER_FUNCTIONS ORDER BY funcid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatXactUserTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stat_xact_user_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stat_xact_user_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stat_xact_user_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stat_xact_user_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STAT_XACT_USER_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioAllIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_all_indexes",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_all_indexes";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_all_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_all_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_ALL_INDEXES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioAllSequences(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_all_sequences",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_all_sequences";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_all_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_all_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_ALL_SEQUENCES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioAllTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_all_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_all_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_all_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_all_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_ALL_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioSysIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_sys_indexes",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_sys_indexes";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_sys_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_sys_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_SYS_INDEXES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioSysSequences(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_sys_sequences",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_sys_sequences";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_sys_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_sys_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_SYS_SEQUENCES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioSysTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_sys_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_sys_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_sys_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_sys_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_SYS_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioUserIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_user_indexes",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_user_indexes";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_user_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_user_indexes";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_USER_INDEXES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioUserSequences(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_user_sequences",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_user_sequences";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_user_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_user_sequences";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_USER_SEQUENCES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatioUserTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statio_user_tables",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statio_user_tables";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statio_user_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statio_user_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT relid FROM PG_catalog.pg_STATIO_USER_TABLES ORDER BY relid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatistic(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statistic",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statistic";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statistic";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statistic";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT starelid FROM PG_catalog.pg_STATISTIC ORDER BY starelid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatisticExt(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statistic_ext",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statistic_ext";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statistic_ext";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statistic_ext";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT stxname FROM PG_catalog.pg_STATISTIC_EXT ORDER BY stxname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatisticExtData(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_statistic_ext_data",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_statistic_ext_data";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_statistic_ext_data";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_statistic_ext_data";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT stxoid FROM PG_catalog.pg_STATISTIC_EXT_DATA ORDER BY stxoid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStats(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stats",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stats";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stats";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stats";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT attname FROM PG_catalog.pg_STATS ORDER BY attname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatsExt(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stats_ext",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stats_ext";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stats_ext";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stats_ext";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT statistics_name FROM PG_catalog.pg_STATS_EXT ORDER BY statistics_name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgStatsExtExprs(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_stats_ext_exprs",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_stats_ext_exprs";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_stats_ext_exprs";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_stats_ext_exprs";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT statistics_name FROM PG_catalog.pg_STATS_EXT_EXPRS ORDER BY statistics_name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgSubscription(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_subscription",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_subscription";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_subscription";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_subscription";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT subname FROM PG_catalog.pg_SUBSCRIPTION ORDER BY subname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgSubscriptionRel(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_subscription_rel",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_subscription_rel";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_subscription_rel";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_subscription_rel";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT srsubid FROM PG_catalog.pg_SUBSCRIPTION_REL ORDER BY srsubid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_tables",
			SetUpScript: []string{
				`create table t1 (pk int primary key, v1 int);`,
				`create table t2 (pk int primary key, v1 int);`,
				`CREATE SCHEMA testschema;`,
				`SET search_path TO testschema;`,
				`CREATE TABLE testing (pk INT primary key, v1 INT);`,

				// Should show classes for all schemas
				`CREATE SCHEMA testschema2;`,
				`SET search_path TO testschema2;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_tables" WHERE tablename='testing' order by 1;`,
					Expected: []sql.Row{{"testschema", "testing", "postgres", nil, "t", "f", "f", "f"}},
				},
				{
					Query:    `SELECT count(*) FROM "pg_catalog"."pg_tables" WHERE schemaname='pg_catalog';`,
					Expected: []sql.Row{{139}},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_tables";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query: "SELECT schemaname, tablename FROM PG_catalog.pg_TABLES WHERE schemaname not in ('information_schema', 'dolt', 'public') ORDER BY tablename DESC LIMIT 3;",
					Expected: []sql.Row{
						{"testschema", "testing"},
						{"pg_catalog", "pg_views"},
						{"pg_catalog", "pg_user_mappings"},
					},
				},
				{
					Query: "SELECT schemaname, tablename FROM PG_catalog.pg_TABLES WHERE schemaname  ='public' ORDER BY tablename;",
					Expected: []sql.Row{
						{"public", "dolt_branches"},
						{"public", "dolt_column_diff"},
						{"public", "dolt_commit_ancestors"},
						{"public", "dolt_commit_diff_t1"},
						{"public", "dolt_commit_diff_t2"},
						{"public", "dolt_commits"},
						{"public", "dolt_conflicts"},
						{"public", "dolt_conflicts_t1"},
						{"public", "dolt_conflicts_t2"},
						{"public", "dolt_constraint_violations"},
						{"public", "dolt_constraint_violations_t1"},
						{"public", "dolt_constraint_violations_t2"},
						{"public", "dolt_diff"},
						{"public", "dolt_diff_t1"},
						{"public", "dolt_diff_t2"},
						{"public", "dolt_history_t1"},
						{"public", "dolt_history_t2"},
						{"public", "dolt_log"},
						{"public", "dolt_merge_status"},
						{"public", "dolt_remote_branches"},
						{"public", "dolt_remotes"},
						{"public", "dolt_schema_conflicts"},
						{"public", "dolt_status"},
						{"public", "dolt_tags"},
						{"public", "dolt_workspace_t1"},
						{"public", "dolt_workspace_t2"},
						{"public", "t1"},
						{"public", "t2"},
					},
				},
				{
					Query: "SELECT schemaname, tablename FROM PG_catalog.pg_TABLES WHERE schemaname  ='dolt' ORDER BY tablename;",
					Expected: []sql.Row{
						{"dolt", "branches"},
						{"dolt", "commit_ancestors"},
						{"dolt", "commits"},
						{"dolt", "conflicts"},
						{"dolt", "constraint_violations"},
						{"dolt", "dolt_backups"},
						{"dolt", "dolt_help"},
						{"dolt", "dolt_stashes"},
						{"dolt", "log"},
						{"dolt", "remote_branches"},
						{"dolt", "remotes"},
						{"dolt", "status"},
					},
				},
			},
		},
	})
}

func TestPgTablespace(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_tablespace",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_tablespace";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_tablespace";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_tablespace";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT spcname FROM PG_catalog.pg_TABLESPACE ORDER BY spcname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTimezoneAbbrevs(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_timezone_abbrevs",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_timezone_abbrevs";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_timezone_abbrevs";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_timezone_abbrevs";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT abbrev FROM PG_catalog.pg_TIMEZONE_ABBREVS ORDER BY abbrev;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgPgTimezoneNames(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_timezone_names",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_timezone_names";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_timezone_names";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_timezone_names";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT name FROM PG_catalog.pg_TIMEZONE_NAMES ORDER BY name;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTransform(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_transform",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_transform";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_transform";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_transform";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT oid FROM PG_catalog.pg_TRANSFORM ORDER BY oid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTrigger(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_trigger",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_trigger";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_trigger";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_trigger";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT tgname FROM PG_catalog.pg_TRIGGER ORDER BY tgname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTsConfig(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_ts_config",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_ts_config";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_ts_config";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_ts_config";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT cfgname FROM PG_catalog.pg_TS_CONFIG ORDER BY cfgname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTsConfigMap(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_ts_config_map",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_ts_config_map";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_ts_config_map";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_ts_config_map";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT mapcfg FROM PG_catalog.pg_TS_CONFIG_MAP ORDER BY mapcfg;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTsDict(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_ts_dict",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_ts_dict";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_ts_dict";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_ts_dict";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT dictname FROM PG_catalog.pg_TS_DICT ORDER BY dictname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTsParser(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_ts_parser",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_ts_parser";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_ts_parser";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_ts_parser";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT prsname FROM PG_catalog.pg_TS_PARSER ORDER BY prsname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgTsTemplate(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_ts_template",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_ts_template";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_ts_template";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_ts_template";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT tmplname FROM PG_catalog.pg_TS_TEMPLATE ORDER BY tmplname;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgType(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_type",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_type" WHERE typname = 'float8' order by 1;`,
					Expected: []sql.Row{{701, "float8", 11, 0, 8, "t", "b", "N", "t", "t", ",", 0, "-", 0, 1022, "float8in", "float8out", "float8recv", "float8send", "-", "-", "-", "d", "p", "f", 0, -1, 0, 0, "", "", "{}"}},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_type";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_type";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query: "SELECT typname FROM PG_catalog.pg_TYPE WHERE typname LIKE '%char' ORDER BY typname;",
					Expected: []sql.Row{
						{"_bpchar"},
						{"_char"},
						{"_varchar"},
						{"bpchar"},
						{"char"},
						{"varchar"},
					},
				},
				{
					Query: `SELECT t1.oid, t1.typname as basetype, t2.typname as arraytype, t2.typsubscript
					FROM   pg_type t1 LEFT JOIN pg_type t2 ON (t1.typarray = t2.oid)
					WHERE  t1.typarray <> 0 AND (t2.oid IS NULL OR t2.typsubscript::regproc <> 'array_subscript_handler'::regproc);`,
					Expected: []sql.Row{},
				},
				{
					Skip: true, // TODO: ERROR: function internal_binary_operator_func_<>(text, regproc) does not exist
					Query: `SELECT t1.oid, t1.typname as basetype, t2.typname as arraytype, t2.typsubscript
					FROM   pg_type t1 LEFT JOIN pg_type t2 ON (t1.typarray = t2.oid)
					WHERE  t1.typarray <> 0 AND (t2.oid IS NULL OR t2.typsubscript <> 'array_subscript_handler'::regproc);`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "pg_type with regtype",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_type" WHERE oid='float8'::regtype;`,
					Expected: []sql.Row{{701, "float8", 11, 0, 8, "t", "b", "N", "t", "t", ",", 0, "-", 0, 1022, "float8in", "float8out", "float8recv", "float8send", "-", "-", "-", "d", "p", "f", 0, -1, 0, 0, "", "", "{}"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='double precision'::regtype;`,
					Expected: []sql.Row{{701, "float8"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='DOUBLE PRECISION'::regtype;`,
					Expected: []sql.Row{{701, "float8"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='pg_catalog.float8'::regtype;`,
					Expected: []sql.Row{{701, "float8"}},
				},
				{
					Query:       `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='public.float8'::regtype;`,
					ExpectedErr: `type "public.float8" does not exist`,
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='VARCHAR'::regtype;`,
					Expected: []sql.Row{{1043, "varchar"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='1043'::regtype;`,
					Expected: []sql.Row{{1043, "varchar"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='VARCHAR(10)'::regtype;`,
					Expected: []sql.Row{{1043, "varchar"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='character varying'::regtype;`,
					Expected: []sql.Row{{1043, "varchar"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='timestamptz'::regtype;`,
					Expected: []sql.Row{{1184, "timestamptz"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='timestamp with time zone'::regtype;`,
					Expected: []sql.Row{{1184, "timestamptz"}},
				},
				{
					Query:    `SELECT oid, typname FROM "pg_catalog"."pg_type" WHERE oid='regtype'::regtype;`,
					Expected: []sql.Row{{2206, "regtype"}},
				},
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_type" WHERE oid='integer[]'::regtype;`,
					Expected: []sql.Row{{1007, "_int4", 11, 0, -1, "f", "b", "A", "f", "t", ",", 0, "array_subscript_handler", 23, 0, "array_in", "array_out", "array_recv", "array_send", "-", "-", "array_typanalyze", "i", "x", "f", 0, -1, 0, 0, "", "", "{}"}},
				},
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_type" WHERE oid='anyarray'::regtype;`,
					Expected: []sql.Row{{2277, "anyarray", 11, 0, -1, "f", "p", "P", "f", "t", ",", 0, "-", 0, 0, "anyarray_in", "anyarray_out", "anyarray_recv", "anyarray_send", "-", "-", "-", "d", "x", "f", 0, -1, 0, 0, "", "", "{}"}},
				},
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_type" WHERE oid='anyelement'::regtype;`,
					Expected: []sql.Row{{2283, "anyelement", 11, 0, 4, "t", "p", "P", "f", "t", ",", 0, "-", 0, 0, "anyelement_in", "anyelement_out", "-", "-", "-", "-", "-", "i", "p", "f", 0, -1, 0, 0, "", "", "{}"}},
				},
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_type" WHERE oid='json'::regtype;`,
					Expected: []sql.Row{{114, "json", 11, 0, -1, "f", "b", "U", "f", "t", ",", 0, "-", 0, 199, "json_in", "json_out", "json_recv", "json_send", "-", "-", "-", "i", "x", "f", 0, -1, 0, 0, "", "", "{}"}},
				},
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_type" WHERE oid='char'::regtype;`,
					Expected: []sql.Row{{1042, "bpchar", 11, 0, -1, "f", "b", "S", "f", "t", ",", 0, "-", 0, 1014, "bpcharin", "bpcharout", "bpcharrecv", "bpcharsend", "bpchartypmodin", "bpchartypmodout", "-", "i", "x", "f", 0, -1, 0, 100, "", "", "{}"}},
				},
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_type" WHERE oid='"char"'::regtype;`,
					Expected: []sql.Row{{18, "char", 11, 0, 1, "t", "b", "Z", "f", "t", ",", 0, "-", 0, 1002, "charin", "charout", "charrecv", "charsend", "-", "-", "-", "c", "p", "f", 0, -1, 0, 0, "", "", "{}"}},
				},
			},
		},
		{
			Name: "user defined type",
			SetUpScript: []string{
				`CREATE DOMAIN domain_type AS INTEGER NOT NULL;`,
				`CREATE TYPE enum_type AS ENUM ('1','2','3')`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_type" WHERE typname = 'domain_type' order by 1;`,
					Expected: []sql.Row{{2382076519, "domain_type", 2200, 0, 4, "t", "d", "N", "f", "t", ",", 0, "-", 0, 1297970968, "domain_in", "int4out", "domain_recv", "int4send", "-", "-", "-", "i", "p", "t", 23, -1, 0, 0, "", "", "{}"}},
				},
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_type" WHERE typname = '_domain_type' order by 1;`,
					Expected: []sql.Row{{1297970968, "_domain_type", 2200, 0, -1, "f", "b", "A", "f", "t", ",", 0, "array_subscript_handler", 2382076519, 0, "array_in", "array_out", "array_recv", "array_send", "-", "-", "array_typanalyze", "i", "x", "f", 0, -1, 0, 0, "", "", "{}"}},
				},
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_type" WHERE typname = 'enum_type' order by 1;`,
					Expected: []sql.Row{{2310414518, "enum_type", 2200, 0, 4, "t", "e", "E", "f", "t", ",", 0, "-", 0, 4245115549, "enum_in", "enum_out", "enum_recv", "enum_send", "-", "-", "-", "i", "p", "f", 0, -1, 0, 0, "", "", "{}"}},
				},
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_type" WHERE typname = '_enum_type' order by 1;`,
					Expected: []sql.Row{{4245115549, "_enum_type", 2200, 0, -1, "f", "b", "A", "f", "t", ",", 0, "array_subscript_handler", 2310414518, 0, "array_in", "array_out", "array_recv", "array_send", "-", "-", "array_typanalyze", "i", "x", "f", 0, -1, 0, 0, "", "", "{}"}},
				},
			},
		},
	})
}

func TestPgUser(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_user",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_user";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_user";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_user";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT usename FROM PG_catalog.pg_USER ORDER BY usename;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgUserMapping(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_user_mapping",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_user_mapping";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_user_mapping";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_user_mapping";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT umuser FROM PG_catalog.pg_USER_MAPPING ORDER BY umuser;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgUserMappings(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_user_mappings",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_user_mappings";`,
					Expected: []sql.Row{},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_user_mappings";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_user_mappings";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT umid FROM PG_catalog.pg_USER_MAPPINGS ORDER BY umid;",
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestPgViews(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_views",
			SetUpScript: []string{
				`CREATE SCHEMA testschema;`,
				`SET search_path TO testschema;`,
				"CREATE TABLE testing (pk INT primary key, v1 INT);",
				`CREATE VIEW testview AS SELECT * FROM testing LIMIT 1;`,
				`CREATE VIEW testview2 AS SELECT * FROM testing LIMIT 2;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM "pg_catalog"."pg_views" WHERE viewname='testview';`,
					Expected: []sql.Row{{"testschema", "testview", "", "SELECT * FROM testing LIMIT 1"}},
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "PG_catalog"."pg_views";`,
					ExpectedErr: "not",
				},
				{ // Different cases and quoted, so it fails
					Query:       `SELECT * FROM "pg_catalog"."PG_views";`,
					ExpectedErr: "not",
				},
				{ // Different cases but non-quoted, so it works
					Query:    "SELECT viewname FROM PG_catalog.pg_VIEWS ORDER BY viewname;",
					Expected: []sql.Row{{"testview"}, {"testview2"}},
				},
			},
		},
	})
}

func TestPgClassIndexes(t *testing.T) {
	sharedSetupScript := []string{
		`create table t1 (a int primary key, b int not null)`,
		`create table t2 (c int primary key, d int not null)`,
		`create index on t2 (d)`,
	}

	RunScripts(t, []ScriptTest{
		{
			Name:        "pg_class index lookup",
			SetUpScript: sharedSetupScript,
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT c.oid
FROM pg_catalog.pg_class c 
WHERE c.relname = 't2' and c.relnamespace = 2200 -- public
ORDER BY 1;`,
					Expected: []sql.Row{
						{1496157034},
					},
				},
				{
					Query: `SELECT c.relname
FROM pg_catalog.pg_class c 
WHERE c.relname > 't' AND c.relname < 't2' AND c.relnamespace = 2200 -- public
AND relkind = 'r'
ORDER BY 1;`,
					Expected: []sql.Row{
						{"t1"},
					},
				},
				{
					Query: `SELECT c.relname
FROM pg_catalog.pg_class c 
WHERE c.relname >= 't1' AND c.relname <= 't2' AND c.relnamespace = 2200 -- public
AND relkind = 'r'
ORDER BY 1;`,
					Expected: []sql.Row{
						{"t1"},
						{"t2"},
					},
				},
				{
					Query: `SELECT c.relname
FROM pg_catalog.pg_class c 
WHERE c.relname >= 't1' AND c.relname < 't2' AND c.relnamespace = 2200 -- public
AND relkind = 'r'
ORDER BY 1;`,
					Expected: []sql.Row{
						{"t1"},
					},
				},
				{
					Query: `SELECT c.relname
FROM pg_catalog.pg_class c 
WHERE c.relname > 't1' AND c.relname <= 't2' AND c.relnamespace = 2200 -- public
AND relkind = 'r'
ORDER BY 1;`,
					Expected: []sql.Row{
						{"t2"},
					},
				},
				{
					Query: `SELECT c.relname
FROM pg_catalog.pg_class c 
WHERE c.relname > 't1' AND c.relname <= 't2' AND c.relnamespace > 2199 AND c.relnamespace < 2201 -- public
AND relkind = 'r'
ORDER BY 1;`,
					Expected: []sql.Row{
						{"t2"},
					},
				},
				{
					Query: `SELECT c.relname
FROM pg_catalog.pg_class c 
WHERE c.oid = 1496157034
ORDER BY 1;`,
					Expected: []sql.Row{
						{"t2"},
					},
				},
				{
					Query: `SELECT c.relname
FROM pg_catalog.pg_class c 
WHERE c.oid IN (1496157034, 1496157035) 
ORDER BY 1;`,
					Expected: []sql.Row{
						{"t2"},
					},
				},
				{
					Query: `SELECT c.relname
FROM pg_catalog.pg_class c 
WHERE c.oid > 1496157033 AND c.oid < 1496157035
ORDER BY 1;`,
					Expected: []sql.Row{
						{"t2"},
					},
				},
				{
					// This is to make sure a full range scan works (we don't support a full range scan on the index yet)
					Query:    `SELECT relname from pg_catalog.pg_class order by oid limit 1;`,
					Expected: []sql.Row{sql.Row{"pg_publication_namespace"}},
				},
				{
					Query: `EXPLAIN SELECT c.oid
FROM pg_catalog.pg_class c 
WHERE c.relname = 't2' and c.relnamespace = 2200
ORDER BY 1;`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [c.oid]"},
						{" └─ Sort(c.oid ASC)"},
						{"     └─ Filter"},
						{"         ├─ (c.relname = 't2' AND c.relnamespace = 2200)"},
						{"         └─ TableAlias(c)"},
						{"             └─ IndexedTableAccess(pg_class)"},
						{"                 ├─ index: [pg_class.relname,pg_class.relnamespace]"},
						{"                 └─ filters: [{[t2, t2], [{Namespace:[\"public\"]}, {Namespace:[\"public\"]}]}]"},
					},
				},
				{
					Query: `EXPLAIN SELECT c.relname
FROM pg_catalog.pg_class c 
WHERE c.relname > 't' AND c.relname < 't2' AND c.relnamespace = 2200 -- public
AND relkind = 'r'
ORDER BY 1;`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [c.relname]"},
						{" └─ Filter"},
						{"     ├─ (((c.relname > 't' AND c.relname < 't2') AND c.relnamespace = 2200) AND c.relkind = 'r')"},
						{"     └─ TableAlias(c)"},
						{"         └─ IndexedTableAccess(pg_class)"},
						{"             ├─ index: [pg_class.relname,pg_class.relnamespace]"},
						{"             └─ filters: [{(t, t2), [{Namespace:[\"public\"]}, {Namespace:[\"public\"]}]}]"},
					},
				},
				{
					Query: `EXPLAIN SELECT c.relname
FROM pg_catalog.pg_class c 
WHERE c.oid = 1496157034
ORDER BY 1;`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [c.relname]"},
						{" └─ Sort(c.relname ASC)"},
						{"     └─ Filter"},
						{"         ├─ c.oid = 1496157034"},
						{"         └─ TableAlias(c)"},
						{"             └─ IndexedTableAccess(pg_class)"},
						{"                 ├─ index: [pg_class.oid]"},
						{"                 └─ filters: [{[{Table:[\"public\",\"t2\"]}, {Table:[\"public\",\"t2\"]}]}]"},
					},
				},
				{
					Query: `EXPLAIN SELECT c.relname
FROM pg_catalog.pg_class c 
WHERE c.oid > 1496157033 AND c.oid < 1496157035
ORDER BY 1;`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [c.relname]"},
						{" └─ Sort(c.relname ASC)"},
						{"     └─ Filter"},
						{"         ├─ (c.oid > 1496157033 AND c.oid < 1496157035)"},
						{"         └─ TableAlias(c)"},
						{"             └─ IndexedTableAccess(pg_class)"},
						{"                 ├─ index: [pg_class.oid]"},
						{"                 └─ filters: [{({OID:[\"1496157033\"]}, {OID:[\"1496157035\"]})}]"},
					},
				},
				{
					Query: `EXPLAIN SELECT c.relname
FROM pg_catalog.pg_class c 
WHERE c.oid IN (1496157034, 1496157035) 
ORDER BY 1;`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [c.relname]"},
						{" └─ Sort(c.relname ASC)"},
						{"     └─ Filter"},
						{"         ├─ c.oid IN (1496157034, 1496157035)"},
						{"         └─ TableAlias(c)"},
						{"             └─ IndexedTableAccess(pg_class)"},
						{"                 ├─ index: [pg_class.oid]"},
						{"                 └─ filters: [{[{Table:[\"public\",\"t2\"]}, {Table:[\"public\",\"t2\"]}]}, {[{OID:[\"1496157035\"]}, {OID:[\"1496157035\"]}]}]"},
					},
				},
			},
		},
		{
			Name:        "join on pg_class",
			SetUpScript: sharedSetupScript,
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT c.relname, a.attname 
FROM pg_catalog.pg_class c 
    JOIN pg_catalog.pg_attribute a 
        ON c.oid = a.attrelid 
WHERE c.relkind = 'r' AND a.attnum > 0 
  AND NOT a.attisdropped
  AND c.relname = 't2'
ORDER BY 1,2;`,
					Expected: []sql.Row{
						{"t2", "c"},
						{"t2", "d"},
					},
				},
				{
					Query: `EXPLAIN SELECT c.relname, a.attname 
FROM pg_catalog.pg_class c 
    JOIN pg_catalog.pg_attribute a 
        ON c.oid = a.attrelid 
WHERE c.relkind = 'r' AND a.attnum > 0 
  AND NOT a.attisdropped
  AND c.relname = 't2'
ORDER BY 1,2;`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [c.relname, a.attname]"},
						{" └─ Sort(c.relname ASC, a.attname ASC)"},
						{"     └─ Filter"},
						{"         ├─ (((c.relkind = 'r' AND a.attnum > 0) AND (NOT(a.attisdropped))) AND c.relname = 't2')"},
						{"         └─ LookupJoin"},
						{"             ├─ TableAlias(a)"},
						{"             │   └─ Table"},
						{"             │       └─ name: pg_attribute"},
						{"             └─ TableAlias(c)"},
						{"                 └─ IndexedTableAccess(pg_class)"},
						{"                     ├─ index: [pg_class.oid]"},
						{"                     └─ keys: a.attrelid"},
					},
				},
			},
		},
		{
			Name:        "left join with nil left result",
			SetUpScript: sharedSetupScript,
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT n.nspname as "Schema",
  c.relname as "Name",
  pg_catalog.pg_get_userbyid(c.relowner) as "Owner",
 c2.oid::pg_catalog.regclass as "Table"
FROM pg_catalog.pg_class c
     LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
     LEFT JOIN pg_catalog.pg_index i ON i.indexrelid = c.oid
     LEFT JOIN pg_catalog.pg_class c2 ON i.indrelid = c2.oid
WHERE c.relkind IN ('I','')
 AND NOT c.relispartition
      AND n.nspname <> 'pg_catalog'
      AND n.nspname !~ '^pg_toast'
      AND n.nspname <> 'information_schema'
  AND pg_catalog.pg_table_is_visible(c.oid)
ORDER BY "Schema", "Name"`,
				},
			},
		},
		{
			Name: "tables in multiple schemas",
			SetUpScript: []string{
				`CREATE SCHEMA s1;`,
				`CREATE SCHEMA s2;`,
				`create schema s3;`,
				`CREATE TABLE s2.t (a INT);`,
				`CREATE TABLE s1.t (b INT);`,
				`CREATE TABLE s3.t (c INT);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `select relname, nspname FROM pg_catalog.pg_class c 
join pg_catalog.pg_namespace n on c.relnamespace = n.oid
where c.relname = 't' and c.relkind = 'r'
order by 1,2`,
					Expected: []sql.Row{
						{"t", "s1"},
						{"t", "s2"},
						{"t", "s3"},
					},
				},
				{
					Query: `select relname, relnamespace FROM pg_catalog.pg_class c 
where c.relname = 't' and c.relkind = 'r'
order by 1,2`,
					Expected: []sql.Row{
						{"t", 1634633383},
						{"t", 1916695891},
						{"t", 2153117264},
					},
				},
				{
					// TODO: this is missing a pushdown index lookup on relnamespace, not sure why
					Query: `explain select relname, nspname FROM pg_catalog.pg_class c 
join pg_catalog.pg_namespace n on c.relnamespace = n.oid
where c.relname = 't' and c.relkind = 'r'
order by 1,2`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [c.relname, n.nspname]"},
						{" └─ Sort(c.relname ASC, n.nspname ASC)"},
						{"     └─ InnerJoin"},
						{"         ├─ c.relnamespace = n.oid"},
						{"         ├─ TableAlias(n)"},
						{"         │   └─ Table"},
						{"         │       └─ name: pg_namespace"},
						{"         └─ Filter"},
						{"             ├─ (c.relname = 't' AND c.relkind = 'r')"},
						{"             └─ TableAlias(c)"},
						{"                 └─ Table"},
						{"                     └─ name: pg_class"},
					},
				},
				{
					Query: `explain select relname, relnamespace FROM pg_catalog.pg_class c 
where c.relname = 't' and c.relkind = 'r'
order by 1,2`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [c.relname, c.relnamespace]"},
						{" └─ Filter"},
						{"     ├─ (c.relname = 't' AND c.relkind = 'r')"},
						{"     └─ TableAlias(c)"},
						{"         └─ IndexedTableAccess(pg_class)"},
						{"             ├─ index: [pg_class.relname,pg_class.relnamespace]"},
						{"             └─ filters: [{[t, t], [NULL, ∞)}]"},
					},
				},
			},
		},
	})
}

func TestPgIndexIndexes(t *testing.T) {
	sharedSetupScript := []string{
		`create table t1 (a int primary key, b int not null)`,
		`create table t2 (c int primary key, d int not null)`,
		`create index on t2 (d)`,
	}

	RunScripts(t, []ScriptTest{
		{
			Name:        "pg_index index lookup",
			SetUpScript: sharedSetupScript,
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM pg_catalog.pg_index i 
WHERE i.indrelid = 1496157034 order by 1`,
					Expected: []sql.Row{
						{3674955271, 1496157034, 1, 0, "f", "f", "f", "f", "f", "f", "t", "f", "t", "t", "f", "{2}", "{}", "{}", "0", nil, nil},
						{3992679530, 1496157034, 1, 0, "t", "f", "t", "f", "f", "f", "t", "f", "t", "t", "f", "{1}", "{}", "{}", "0", nil, nil},
					},
				},
				{
					Query: `SELECT c.relname, c2.relname FROM pg_catalog.pg_index i
         join pg_class c on i.indrelid = c.oid
         join pg_class c2 on i.indexrelid = c2.oid
WHERE c.relname = 't2' order by 1,2`,
					Expected: []sql.Row{
						{"t2", "d"},
						{"t2", "t2_pkey"},
					},
				},
				{
					Query: `SELECT i.indrelid FROM pg_catalog.pg_index i 
WHERE i.indexrelid = (SELECT c.oid FROM pg_catalog.pg_class c WHERE c.relname = 't2_pkey')
ORDER BY 1;`,
					Expected: []sql.Row{
						{1496157034},
					},
				},
				{
					Query: `SELECT count(*) FROM pg_catalog.pg_index i 
WHERE i.indrelid = 1496157034`,
					Expected: []sql.Row{
						{2},
					},
				},
				{
					Query: `SELECT i.indisprimary FROM pg_catalog.pg_index i 
WHERE i.indrelid = 1496157034 AND i.indisprimary = true
ORDER BY 1;`,
					Expected: []sql.Row{
						{"t"},
					},
				},
				{
					Query: `SELECT COUNT(*) FROM pg_catalog.pg_index i 
WHERE i.indrelid IN (1496157033, 1496157034)`,
					Expected: []sql.Row{
						{2},
					},
				},
				{
					// TODO: this uses an index but the plan doesn't show it because of prepared statements
					Query: `EXPLAIN SELECT i.indrelid FROM pg_catalog.pg_index i 
WHERE i.indexrelid = (SELECT c.oid FROM pg_catalog.pg_class c WHERE c.relname = 't1_pkey')
ORDER BY 1;`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [i.indrelid]"},
						{" └─ Sort(i.indrelid ASC)"},
						{"     └─ Filter"},
						{"         ├─ i.indexrelid = Subquery((select c.oid from pg_class as c where ? = ?))"},
						{"         └─ TableAlias(i)"},
						{"             └─ Table"},
						{"                 └─ name: pg_index"},
					},
				},
				{
					Query: `EXPLAIN SELECT COUNT(*) FROM pg_catalog.pg_index i 
WHERE i.indrelid = 1496157034 ORDER BY 1`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [count(1) as count]"},
						{" └─ Sort(count(1) as count ASC)"},
						{"     └─ GroupBy"},
						{"         ├─ SelectDeps(COUNT(1))"},
						{"         ├─ Grouping()"},
						{"         └─ Filter"},
						{"             ├─ i.indrelid = 1496157034"},
						{"             └─ TableAlias(i)"},
						{"                 └─ IndexedTableAccess(pg_index)"},
						{"                     ├─ index: [pg_index.indrelid]"},
						{"                     └─ filters: [{[{Table:[\"public\",\"t2\"]}, {Table:[\"public\",\"t2\"]}]}]"},
					},
				},
				{
					Query: `EXPLAIN SELECT COUNT(*) FROM pg_catalog.pg_index i 
WHERE i.indrelid IN (1496157033, 1496157034) ORDER BY 1`,
					Expected: []sql.Row{
						{"Project"},
						{" ├─ columns: [count(1) as count]"},
						{" └─ Sort(count(1) as count ASC)"},
						{"     └─ GroupBy"},
						{"         ├─ SelectDeps(COUNT(1))"},
						{"         ├─ Grouping()"},
						{"         └─ Filter"},
						{"             ├─ i.indrelid IN (1496157033, 1496157034)"},
						{"             └─ TableAlias(i)"},
						{"                 └─ IndexedTableAccess(pg_index)"},
						{"                     ├─ index: [pg_index.indrelid]"},
						{"                     └─ filters: [{[{OID:[\"1496157033\"]}, {OID:[\"1496157033\"]}]}, {[{Table:[\"public\",\"t2\"]}, {Table:[\"public\",\"t2\"]}]}]"},
					},
				},
			},
		},
	})
}

func TestSqlAlchemyQueries(t *testing.T) {
	sharedSetupScript := []string{
		`create table t1 (a int primary key, b int not null)`,
		`create table t2 (a int primary key, b int not null)`,
		`create index on t2 (b)`,
	}

	RunScripts(t, []ScriptTest{
		{
			Name:        "schema for dolt_log",
			SetUpScript: sharedSetupScript,
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT pg_catalog.pg_attribute.attname AS name, pg_catalog.format_type(pg_catalog.pg_attribute.atttypid, pg_catalog.pg_attribute.atttypmod) AS format_type, (SELECT pg_catalog.pg_get_expr(pg_catalog.pg_attrdef.adbin, pg_catalog.pg_attrdef.adrelid) AS pg_get_expr_1 
FROM pg_catalog.pg_attrdef 
WHERE pg_catalog.pg_attrdef.adrelid = pg_catalog.pg_attribute.attrelid AND pg_catalog.pg_attrdef.adnum = pg_catalog.pg_attribute.attnum AND pg_catalog.pg_attribute.atthasdef) AS "default", pg_catalog.pg_attribute.attnotnull AS not_null, pg_catalog.pg_class.relname AS table_name, pg_catalog.pg_description.description AS comment, pg_catalog.pg_attribute.attgenerated AS generated, (SELECT json_build_object('always', pg_catalog.pg_attribute.attidentity = 'a', 'start', pg_catalog.pg_sequence.seqstart, 'increment', pg_catalog.pg_sequence.seqincrement, 'minvalue', pg_catalog.pg_sequence.seqmin, 'maxvalue', pg_catalog.pg_sequence.seqmax, 'cache', pg_catalog.pg_sequence.seqcache, 'cycle', pg_catalog.pg_sequence.seqcycle) AS json_build_object_1 
FROM pg_catalog.pg_sequence 
WHERE pg_catalog.pg_attribute.attidentity != '' AND pg_catalog.pg_sequence.seqrelid = CAST(CAST(pg_catalog.pg_get_serial_sequence(CAST(CAST(pg_catalog.pg_attribute.attrelid AS REGCLASS) AS TEXT), pg_catalog.pg_attribute.attname) AS REGCLASS) AS OID)) AS identity_options 
FROM pg_catalog.pg_class LEFT OUTER JOIN pg_catalog.pg_attribute ON pg_catalog.pg_class.oid = pg_catalog.pg_attribute.attrelid AND pg_catalog.pg_attribute.attnum > 0 AND NOT pg_catalog.pg_attribute.attisdropped LEFT OUTER JOIN pg_catalog.pg_description ON pg_catalog.pg_description.objoid = pg_catalog.pg_attribute.attrelid AND pg_catalog.pg_description.objsubid = pg_catalog.pg_attribute.attnum JOIN pg_catalog.pg_namespace ON pg_catalog.pg_namespace.oid = pg_catalog.pg_class.relnamespace 
WHERE pg_catalog.pg_class.relkind = ANY (ARRAY['r', 'p', 'f', 'v', 'm']) AND pg_catalog.pg_table_is_visible(pg_catalog.pg_class.oid) AND pg_catalog.pg_namespace.nspname != 'pg_catalog' AND pg_catalog.pg_class.relname IN ('dolt_log') ORDER BY pg_catalog.pg_class.relname, pg_catalog.pg_attribute.attnum`,
					Expected: []sql.Row{
						{"commit_hash", "text", nil, "t", "dolt_log", nil, "", nil},
						{"committer", "text", nil, "t", "dolt_log", nil, "", nil},
						{"email", "text", nil, "t", "dolt_log", nil, "", nil},
						{"date", "timestamp without time zone", nil, "t", "dolt_log", nil, "", nil},
						{"message", "text", nil, "t", "dolt_log", nil, "", nil},
						{"commit_order", "bigint", nil, "t", "dolt_log", nil, "", nil},
					},
				},
			},
		},
		{
			Name:        "type queries",
			SetUpScript: sharedSetupScript,
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT pg_catalog.pg_type.typname AS name,
       pg_catalog.pg_type_is_visible(pg_catalog.pg_type.oid) AS visible,
       pg_catalog.pg_namespace.nspname AS schema,
       lbl_agg.labels AS labels
FROM pg_catalog.pg_type
JOIN pg_catalog.pg_namespace ON pg_catalog.pg_namespace.oid = pg_catalog.pg_type.typnamespace
    LEFT OUTER JOIN 
    (SELECT pg_catalog.pg_enum.enumtypid AS enumtypid, 
    array_agg(CAST(pg_catalog.pg_enum.enumlabel AS TEXT) ORDER BY pg_catalog.pg_enum.enumsortorder) 
    AS labels FROM pg_catalog.pg_enum GROUP BY pg_catalog.pg_enum.enumtypid) AS lbl_agg
    ON pg_catalog.pg_type.oid = lbl_agg.enumtypid WHERE pg_catalog.pg_type.typtype = 'e'
    ORDER BY pg_catalog.pg_namespace.nspname, pg_catalog.pg_type.typname`,
				},
			},
		},
		{
			Name:        "dolt_log schema 2",
			SetUpScript: sharedSetupScript,
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT pg_catalog.pg_attribute.attname AS name,
    pg_catalog.format_type(pg_catalog.pg_attribute.atttypid,
    pg_catalog.pg_attribute.atttypmod) AS format_type,
    (SELECT pg_catalog.pg_get_expr(pg_catalog  .pg_attrdef.adbin, pg_catalog.pg_attrdef.adrelid) AS pg_get_expr_1
			 FROM pg_catalog.pg_attrdef 
			 WHERE pg_catalog.pg_attrdef.adrelid = pg_catalog.pg_attribute.attrelid
				 AND pg_catalog.pg_attrdef.adnum = pg_catalog.pg_attribute.attnum
				 AND pg_catalog.pg_attribute.atthasdef) AS "default",
    pg_catalog.pg_attribute.attnotnull AS not_null,
    pg_catalog.pg_class.relname AS table_name,
    pg_catalog.pg_description.description AS comment,
    pg_catalog.pg_attribute.attgenerated AS generated,
    (SELECT json_build_object('always', pg_catalog.pg_attribute.attidentity = 'a',
                              'start', pg_catalog.pg_sequence.seqstart,
                              'increment', pg_catalog.pg_sequence.seqincrement,
                              'minvalue', pg_catalog.pg_sequence.seqmin,
                              'maxvalue', pg_catalog.pg_sequence.seqmax,
                              'cache', pg_catalog.pg_sequence.seqcache,
                              'cycle', pg_catalog.pg_sequence.seqcycle) AS json_build_object_1
    			FROM pg_catalog.pg_sequence
       		WHERE pg_catalog.pg_attribute.attidentity != ''
       		AND pg_catalog.pg_sequence.seqrelid = CAST(CAST(pg_catalog.pg_get_serial_sequence(CAST(CAST(pg_catalog.pg_attribute.attrelid AS REGCLASS) AS TEXT), pg_catalog.pg_attribute.attname) AS REGCLASS) AS OID)
       ) AS identity_options
   FROM pg_catalog.pg_class
   LEFT OUTER JOIN pg_catalog.pg_attribute ON pg_catalog.pg_class.oid = pg_catalog.pg_attribute.attrelid 
       AND pg_catalog.pg_attribute.attnum > 0 AND NOT pg_catalog.pg_attribute.attisdropped 
       LEFT OUTER JOIN pg_catalog.pg_description ON pg_catalog.pg_description.objoid = pg_catalog.pg_attribute.attrelid 
       AND pg_catalog.pg_description.objsubid = pg_catalog.pg_attribute.attnum
       JOIN pg_catalog.pg_namespace ON pg_catalog.pg_namespace.oid = pg_catalog.pg_class.relnamespace
       WHERE pg_catalog.pg_class.relkind = ANY (ARRAY['r', 'p', 'f  ', 'v', 'm']) 
       AND pg_catalog.pg_table_is_visible(pg_catalog.pg_class.oid) 
       AND pg_catalog.pg_namespace.nspname != 'pg_catalog' 
       AND pg_catalog.pg_class.relname IN ('dolt_log') 
       ORDER BY pg_catalog.pg_class.relname, pg_catalog.pg_attribute.attnum`,
					Expected: []sql.Row{
						{"commit_hash", "text", nil, "t", "dolt_log", nil, "", nil},
						{"committer", "text", nil, "t", "dolt_log", nil, "", nil},
						{"email", "text", nil, "t", "dolt_log", nil, "", nil},
						{"date", "timestamp without time zone", nil, "t", "dolt_log", nil, "", nil},
						{"message", "text", nil, "t", "dolt_log", nil, "", nil},
						{"commit_order", "bigint", nil, "t", "dolt_log", nil, "", nil},
					},
				},
			},
		},
		{
			Name:        "constraints",
			SetUpScript: sharedSetupScript,
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT attr.conrelid, 
       array_agg(CAST(attr.attname AS TEXT) ORDER BY attr.ord) AS cols,
       attr.conname,
       min(attr.description) AS description,
       NULL AS extra FROM 
				(SELECT con.conrelid AS conrelid,
				        con.conname AS conname,
				        con.conindid AS conindid,
				        con.description AS description,
				        con.ord AS ord,
				        pg_catalog.pg_attribute.attname AS attname
				 FROM pg_catalog.pg_attribute JOIN 
				     (SELECT pg_catalog.pg_constraint.conrelid AS conrelid,
				             pg_catalog.pg_constraint.conname AS conname,
				             pg_catalog.pg_constraint.conindid AS conindid,
				             unnest(pg_catalog.pg_constraint.conkey) AS attnum,
				             generate_subscripts(pg_catalog.pg_constraint.conkey, 1) AS ord,
				             pg_catalog.pg_description.description AS description 
				      FROM pg_catalog.pg_constraint 
				          LEFT OUTER JOIN pg_catalog.pg_description 
				              ON pg_catalog.pg_description.objoid = pg_catalog.pg_constraint.oid
				      WHERE pg_catalog.pg_constraint.contype = 'p'
				        AND pg_catalog.pg_constraint.conrelid IN (3491847678)) AS con
				     ON pg_catalog.pg_attribute.attnum = con.attnum 
				            AND pg_catalog.pg_attribute.attrelid = con.conrelid
				 WHERE con.conrelid IN (3491847678)) AS attr 
            GROUP BY attr.conrelid, attr.conname ORDER BY attr.conrelid, attr.conname`,
				},
			},
		},
		{
			Name:        "has constraints",
			SetUpScript: sharedSetupScript,
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT pg_catalog.pg_index.indrelid,
       cls_idx.relname AS relname_index,
       pg_catalog.pg_index.indisunique,
       pg_catalog.pg_constraint.conrelid IS NOT NULL AS has_constraint,
       pg_catalog.pg_index.indoption,
       cls_idx.reloptions,
       pg_catalog.pg_am.amname,
       CASE WHEN (pg_catalog.pg_index.indpred IS NOT NULL) 
           THEN pg_catalog.pg_get_expr(pg_catalog.pg_index.indpred, pg_catalog.pg_index.indrelid) 
           END AS filter_definition,
    	 pg_catalog.pg_index.indnkeyatts,
    	 pg_catalog.pg_index.indnullsnotdistinct,
    	 idx_cols.elements,
    	 idx_cols.elements_is_expr 
FROM pg_catalog.pg_index 
    JOIN pg_catalog.pg_class AS cls_idx 
        ON pg_catalog.pg_index.indexrelid = cls_idx.oid 
    JOIN pg_catalog.pg_am 
        ON cls_idx.relam = pg_catalog.pg_am.oid 
    LEFT OUTER JOIN (SELECT idx_attr.indexrelid AS indexrelid, min(idx_attr.indrelid) AS min_1,
                            array_agg(idx_attr.element ORDER BY idx_attr.ord) AS elements,
                            array_agg(idx_attr.is_expr ORDER BY idx_attr.ord) AS elements_is_expr
                     FROM (SELECT idx.indexrelid AS indexrelid,
                                  idx.indrelid AS indrelid,
                                  idx.ord AS ord,
                                  CASE WHEN (idx.attnum = 0) THEN pg_catalog.pg_get_indexdef(idx.indexrelid, idx.ord + 1, true)
                                      ELSE CAST(pg_catalog.pg_attribute.attname AS TEXT) 
                                      END AS element,
                                  idx.attnum = 0 AS is_expr
                           FROM (SELECT pg_catalog.pg_index.indexrelid AS indexrelid,
                                        pg_catalog.pg_index.indrelid AS indrelid,
                                        unnest(pg_catalog.pg_index.indkey) AS attnum,
                                        generate_subscripts(pg_catalog.pg_index.indkey, 1) AS ord
                                 FROM pg_catalog.pg_index
                                 WHERE NOT pg_catalog.pg_index.indisprimary
                                   AND pg_catalog.pg_index.indrelid IN (3491847678)) AS idx
                           LEFT OUTER JOIN pg_catalog.pg_attribute
                               ON pg_catalog.pg_attribute.attnum = idx.attnum
                                      AND pg_catalog.pg_attribute.attrelid = idx.indrelid
                           WHERE idx.indrelid IN (3491847678)) AS idx_attr
                     GROUP BY idx_attr.indexrelid) AS idx_cols
        ON pg_catalog.pg_index.indexrelid = idx_cols.indexrelid
    LEFT OUTER JOIN pg_catalog.pg_constraint
        ON pg_catalog.pg_index.indrelid = pg_catalog.pg_constraint.conrelid
               AND pg_catalog.pg_index.indexrelid = pg_catalog.pg_constraint.conindid
               AND pg_catalog.pg_constraint.contype = ANY (ARRAY['p', 'u', 'x'])
WHERE pg_catalog.pg_index.indrelid IN (3491847678)
  AND NOT pg_catalog.pg_index.indisprimary
ORDER BY pg_catalog.pg_index.indrelid, cls_idx.relname`,
				},
			},
		},
		{
			Name:        "attributes",
			SetUpScript: sharedSetupScript,
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT attr.conrelid,
       array_agg(CAST(attr.attname AS TEXT) ORDER BY attr.ord) AS cols,
       attr.conname,
       min(attr.description) AS description,
       bool_and(pg_catalog.pg_index.indnullsnotdistinct) AS indnullsnotdistinct
FROM (SELECT con.conrelid AS conrelid,
             con.conname AS conname,
             con.conindid AS conindid,
             con.description AS description,
             con.ord AS ord, pg_catalog.pg_attribute.attname AS attname
      FROM pg_catalog.pg_attribute 
          JOIN (SELECT pg_catalog.pg_constraint.conrelid AS conrelid,
                       pg_catalog.pg_constraint.conname AS conname,
                       pg_catalog.pg_constraint.conindid AS conindid,
                       unnest(pg_catalog.pg_constraint.conkey) AS attnum,
                       generate_subscripts(pg_catalog.pg_constraint.conkey, 1) AS ord,
                       pg_catalog.pg_description.description AS description
                FROM pg_catalog.pg_constraint 
                    LEFT OUTER JOIN pg_catalog.pg_description 
                        ON pg_catalog.pg_description.objoid = pg_catalog.pg_constraint.oid
                WHERE pg_catalog.pg_constraint.contype = 'u'
                  AND pg_catalog.pg_constraint.conrelid IN (3491847678)) AS con
              ON pg_catalog.pg_attribute.attnum = con.attnum
                     AND pg_catalog.pg_attribute.attrelid = con.conrelid
      WHERE con.conrelid IN (3491847678)) AS attr
    JOIN pg_catalog.pg_index 
        ON attr.conindid = pg_catalog.pg_index.indexrelid
GROUP BY attr.conrelid, attr.conname
ORDER BY attr.conrelid, attr.conname`,
				},
			},
		},
		{
			Name:        "key constraints",
			SetUpScript: sharedSetupScript,
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT attr.conrelid,
       array_agg(CAST(attr.attname AS TEXT) ORDER BY attr.ord) AS cols,
       attr.conname,
       min(attr.description) AS description,
       NULL AS extra FROM
                         (SELECT con.conrelid AS conrelid,
                                 con.conname AS conname,
                                 con.conindid AS conindid,
                                 con.description AS description,
                                 con.ord AS ord,
                                 pg_catalog.pg_attribute.attname AS attname
                          FROM pg_catalog.pg_attribute 
                              JOIN (SELECT pg_catalog.pg_constraint.conrelid AS conrelid,
                                           pg_catalog.pg_constraint.conname AS conname,
                                           pg_catalog.pg_constraint.conindid AS conindid,
                                           unnest(pg_catalog.pg_constraint.conkey) AS attnum,
                                           generate_subscripts(pg_catalog.pg_constraint.conkey, 1) AS ord,
                                           pg_catalog.pg_description.description AS description
                                    FROM pg_catalog.pg_constraint
                                        LEFT OUTER JOIN pg_catalog.pg_description
                                            ON pg_catalog.pg_description.objoid = pg_catalog.pg_constraint.oid
                                    WHERE pg_catalog.pg_constraint.contype = 'p'
                                      AND pg_catalog.pg_constraint.conrelid IN (select oid from pg_class where relname='t1'))
                                  AS con
                                  ON pg_catalog.pg_attribute.attnum = con.attnum
                                         AND pg_catalog.pg_attribute.attrelid = con.conrelid
                          WHERE con.conrelid IN (select oid from pg_class where relname='t1')) AS attr
                     GROUP BY attr.conrelid, attr.conname
                     ORDER BY attr.conrelid, attr.conname`,
					Expected: []sql.Row{
						{1249736862, "{a}", "t1_pkey", nil, nil},
					},
				},
			},
		},
		{
			Name:        "index queries",
			SetUpScript: sharedSetupScript,
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT pg_catalog.pg_index.indrelid,
       cls_idx.relname AS relname_index,
       pg_catalog.pg_index.indisunique,
       pg_catalog.pg_constraint.conrelid IS NOT NULL AS has_constraint,
       pg_catalog.pg_index.indoption,
       cls_idx.reloptions,
       pg_catalog.pg_am.amname,
       CASE WHEN (pg_catalog.pg_index.indpred IS NOT NULL)
           THEN pg_catalog.pg_get_expr(pg_catalog.pg_index.indpred, pg_catalog.pg_index.indrelid)
           END AS filter_definition,
       pg_catalog.pg_index.indnkeyatts,
       pg_catalog.pg_index.indnullsnotdistinct,
       idx_cols.elements,
       idx_cols.elements_is_expr
FROM pg_catalog.pg_index
    JOIN pg_catalog.pg_class AS cls_idx 
        ON pg_catalog.pg_index.indexrelid = cls_idx.oid
    JOIN pg_catalog.pg_am ON cls_idx.relam = pg_catalog.pg_am.oid
    LEFT OUTER JOIN 
    (SELECT idx_attr.indexrelid AS indexrelid,
            min(idx_attr.indrelid) AS min_1,
            array_agg(idx_attr.element ORDER BY idx_attr.ord) AS elements,
            array_agg(idx_attr.is_expr ORDER BY idx_attr.ord) AS elements_is_expr
     FROM (SELECT idx.indexrelid AS indexrelid,
                  idx.indrelid AS indrelid,
                  idx.ord AS ord,
                  CASE WHEN (idx.attnum = 0)
                      THEN pg_catalog.pg_get_indexdef(idx.indexrelid, idx.ord + 1, true)
                      ELSE CAST(pg_catalog.pg_attribute.attname AS TEXT)
                      END AS element,
               idx.attnum = 0 AS is_expr 
           FROM (SELECT pg_catalog.pg_index.indexrelid AS indexrelid,
                        pg_catalog.pg_index.indrelid AS indrelid,
                        unnest(pg_catalog.pg_index.indkey) AS attnum,
                        generate_subscripts(pg_catalog.pg_index.indkey, 1) AS ord
                 FROM pg_catalog.pg_index 
                 WHERE NOT pg_catalog.pg_index.indisprimary 
                   AND pg_catalog.pg_index.indrelid IN (select oid from pg_class where relname='t2')) AS idx
               LEFT OUTER JOIN pg_catalog.pg_attribute
                   ON pg_catalog.pg_attribute.attnum = idx.attnum
                          AND pg_catalog.pg_attribute.attrelid = idx.indrelid
           WHERE idx.indrelid IN (select oid from pg_class where relname='t2')) AS idx_attr GROUP BY idx_attr.indexrelid) AS idx_cols
        ON pg_catalog.pg_index.indexrelid = idx_cols.indexrelid
    LEFT OUTER JOIN pg_catalog.pg_constraint
        ON pg_catalog.pg_index.indrelid = pg_catalog.pg_constraint.conrelid
               AND pg_catalog.pg_index.indexrelid = pg_catalog.pg_constraint.conindid
               AND pg_catalog.pg_constraint.contype = ANY (ARRAY['p', 'u', 'x']) 
WHERE pg_catalog.pg_index.indrelid IN (select oid from pg_class where relname='t2')
  AND NOT pg_catalog.pg_index.indisprimary ORDER BY pg_catalog.pg_index.indrelid, cls_idx.relname`,
					Expected: []sql.Row{
						{1496157034, "b", "f", "f", "0", nil, "btree", nil, 0, "f", "{b}", "{f}"},
					},
				},
			},
		},
	})
}

func TestSystemTablesInPgcatalog(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_tables",
			SetUpScript: []string{
				`CREATE SCHEMA s1;`,
				`CREATE TABLE s1.t1 (pk INT primary key, v1 INT);`,
			},
			// TODO: some of these dolt_ table names are wrong, see https://github.com/dolthub/doltgresql/issues/1560
			Assertions: []ScriptTestAssertion{
				{
					Query: "select * from pg_catalog.pg_tables where schemaname not in ('information_schema', 'pg_catalog') order by schemaname, tablename;",
					Expected: []sql.Row{
						{"dolt", "branches", "postgres", nil, "f", "f", "f", "f"},
						{"dolt", "commit_ancestors", "postgres", nil, "f", "f", "f", "f"},
						{"dolt", "commits", "postgres", nil, "f", "f", "f", "f"},
						{"dolt", "conflicts", "postgres", nil, "f", "f", "f", "f"},
						{"dolt", "constraint_violations", "postgres", nil, "f", "f", "f", "f"},
						{"dolt", "dolt_backups", "postgres", nil, "f", "f", "f", "f"},
						{"dolt", "dolt_help", "postgres", nil, "f", "f", "f", "f"},
						{"dolt", "dolt_stashes", "postgres", nil, "f", "f", "f", "f"},
						{"dolt", "log", "postgres", nil, "f", "f", "f", "f"},
						{"dolt", "remote_branches", "postgres", nil, "f", "f", "f", "f"},
						{"dolt", "remotes", "postgres", nil, "f", "f", "f", "f"},
						{"dolt", "status", "postgres", nil, "f", "f", "f", "f"},
						{"public", "dolt_branches", "postgres", nil, "f", "f", "f", "f"},
						{"public", "dolt_column_diff", "postgres", nil, "f", "f", "f", "f"},
						{"public", "dolt_commit_ancestors", "postgres", nil, "f", "f", "f", "f"},
						{"public", "dolt_commits", "postgres", nil, "f", "f", "f", "f"},
						{"public", "dolt_conflicts", "postgres", nil, "f", "f", "f", "f"},
						{"public", "dolt_constraint_violations", "postgres", nil, "f", "f", "f", "f"},
						{"public", "dolt_diff", "postgres", nil, "f", "f", "f", "f"},
						{"public", "dolt_log", "postgres", nil, "f", "f", "f", "f"},
						{"public", "dolt_merge_status", "postgres", nil, "f", "f", "f", "f"},
						{"public", "dolt_remote_branches", "postgres", nil, "f", "f", "f", "f"},
						{"public", "dolt_remotes", "postgres", nil, "f", "f", "f", "f"},
						{"public", "dolt_schema_conflicts", "postgres", nil, "f", "f", "f", "f"},
						{"public", "dolt_status", "postgres", nil, "f", "f", "f", "f"},
						{"public", "dolt_tags", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_branches", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_column_diff", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_commit_ancestors", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_commit_diff_t1", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_commits", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_conflicts", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_conflicts_t1", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_constraint_violations", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_constraint_violations_t1", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_diff", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_diff_t1", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_history_t1", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_log", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_merge_status", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_remote_branches", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_remotes", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_schema_conflicts", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_status", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_tags", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "dolt_workspace_t1", "postgres", nil, "f", "f", "f", "f"},
						{"s1", "t1", "postgres", nil, "t", "f", "f", "f"},
					},
				},
			},
		},
		{
			Name: "pg_class",
			SetUpScript: []string{
				`CREATE SCHEMA s1;`,
				`CREATE TABLE s1.t1 (pk INT primary key, v1 INT);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// TODO: some of these dolt_ table names are wrong, see https://github.com/dolthub/doltgresql/issues/1560
					Query: `select oid, relname, relnamespace, relkind from pg_class where relnamespace not in (select oid from pg_namespace where nspname in ('information_schema', 'pg_catalog')) order by relnamespace, relname;`,
					Expected: []sql.Row{
						{458530874, "dolt_branches", 2200, "r"},
						{2056815203, "dolt_column_diff", 2200, "r"},
						{1555944102, "dolt_commit_ancestors", 2200, "r"},
						{3152041833, "dolt_commits", 2200, "r"},
						{245736992, "dolt_conflicts", 2200, "r"},
						{1932298159, "dolt_constraint_violations", 2200, "r"},
						{2357712556, "dolt_diff", 2200, "r"},
						{3491847678, "dolt_log", 2200, "r"},
						{604995978, "dolt_merge_status", 2200, "r"},
						{887648921, "dolt_remote_branches", 2200, "r"},
						{341706375, "dolt_remotes", 2200, "r"},
						{3210116770, "dolt_schema_conflicts", 2200, "r"},
						{1060579466, "dolt_status", 2200, "r"},
						{1807684176, "dolt_tags", 2200, "r"},
						{1763579892, "dolt_branches", 1634633383, "r"},
						{1212681264, "dolt_column_diff", 1634633383, "r"},
						{4001633963, "dolt_commit_ancestors", 1634633383, "r"},
						{115796810, "dolt_commit_diff_t1", 1634633383, "r"},
						{3112353516, "dolt_commits", 1634633383, "r"},
						{2517735330, "dolt_conflicts", 1634633383, "r"},
						{2419641880, "dolt_conflicts_t1", 1634633383, "r"},
						{1322753784, "dolt_constraint_violations", 1634633383, "r"},
						{3390577184, "dolt_constraint_violations_t1", 1634633383, "r"},
						{649632770, "dolt_diff", 1634633383, "r"},
						{876336553, "dolt_diff_t1", 1634633383, "r"},
						{3422698383, "dolt_history_t1", 1634633383, "r"},
						{2067982358, "dolt_log", 1634633383, "r"},
						{3947121936, "dolt_merge_status", 1634633383, "r"},
						{867423409, "dolt_remote_branches", 1634633383, "r"},
						{373092098, "dolt_remotes", 1634633383, "r"},
						{225426095, "dolt_schema_conflicts", 1634633383, "r"},
						{3554775706, "dolt_status", 1634633383, "r"},
						{3246414078, "dolt_tags", 1634633383, "r"},
						{1640933374, "dolt_workspace_t1", 1634633383, "r"},
						{2849341124, "t1", 1634633383, "r"},
						{512149063, "t1_pkey", 1634633383, "i"},
						{398111247, "branches", 1882653564, "r"},
						{4126412490, "commit_ancestors", 1882653564, "r"},
						{3425483043, "commits", 1882653564, "r"},
						{1218627310, "conflicts", 1882653564, "r"},
						{1967026500, "constraint_violations", 1882653564, "r"},
						{1167248682, "dolt_backups", 1882653564, "r"},
						{629684363, "dolt_help", 1882653564, "r"},
						{1384122262, "dolt_stashes", 1882653564, "r"},
						{909123395, "log", 1882653564, "r"},
						{148630507, "remote_branches", 1882653564, "r"},
						{1670572237, "remotes", 1882653564, "r"},
						{3431637196, "status", 1882653564, "r"},
					},
				},
			},
		},
		{
			Name: "pg_attribute",
			SetUpScript: []string{
				`CREATE SCHEMA s1;`,
				`CREATE TABLE s1.t1 (pk INT primary key, v1 INT);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `select attrelid, attname, atttypid, attnum, attnotnull, atthasdef, attisdropped from pg_catalog.pg_attribute where attrelid in (select oid from pg_catalog.pg_class where relnamespace not in (select oid from pg_namespace where nspname in ('information_schema', 'pg_catalog'))) order by attrelid, attnum;`,
					Expected: []sql.Row{
						{115796810, "to_pk", 23, 1, "f", "f", "f"},
						{115796810, "to_v1", 23, 2, "f", "f", "f"},
						{115796810, "to_commit", 25, 3, "f", "f", "f"},
						{115796810, "to_commit_date", 1114, 4, "f", "f", "f"},
						{115796810, "from_pk", 23, 5, "f", "f", "f"},
						{115796810, "from_v1", 23, 6, "f", "f", "f"},
						{115796810, "from_commit", 25, 7, "f", "f", "f"},
						{115796810, "from_commit_date", 1114, 8, "f", "f", "f"},
						{115796810, "diff_type", 25, 9, "f", "f", "f"},
						{148630507, "name", 25, 1, "t", "f", "f"},
						{148630507, "hash", 25, 2, "t", "f", "f"},
						{148630507, "latest_committer", 25, 3, "f", "f", "f"},
						{148630507, "latest_committer_email", 25, 4, "f", "f", "f"},
						{148630507, "latest_commit_date", 1114, 5, "f", "f", "f"},
						{148630507, "latest_commit_message", 25, 6, "f", "f", "f"},
						{225426095, "table_name", 25, 1, "t", "f", "f"},
						{225426095, "base_schema", 25, 2, "t", "f", "f"},
						{225426095, "our_schema", 25, 3, "t", "f", "f"},
						{225426095, "their_schema", 25, 4, "t", "f", "f"},
						{225426095, "description", 25, 5, "t", "f", "f"},
						{245736992, "table", 25, 1, "t", "f", "f"},
						{245736992, "num_conflicts", 20, 2, "t", "f", "f"},
						{341706375, "name", 25, 1, "t", "f", "f"},
						{341706375, "url", 25, 2, "t", "f", "f"},
						{341706375, "fetch_specs", 114, 3, "f", "f", "f"},
						{341706375, "params", 114, 4, "f", "f", "f"},
						{373092098, "name", 25, 1, "t", "f", "f"},
						{373092098, "url", 25, 2, "t", "f", "f"},
						{373092098, "fetch_specs", 114, 3, "f", "f", "f"},
						{373092098, "params", 114, 4, "f", "f", "f"},
						{398111247, "name", 25, 1, "t", "f", "f"},
						{398111247, "hash", 25, 2, "t", "f", "f"},
						{398111247, "latest_committer", 25, 3, "f", "f", "f"},
						{398111247, "latest_committer_email", 25, 4, "f", "f", "f"},
						{398111247, "latest_commit_date", 1114, 5, "f", "f", "f"},
						{398111247, "latest_commit_message", 25, 6, "f", "f", "f"},
						{398111247, "remote", 25, 7, "f", "f", "f"},
						{398111247, "branch", 25, 8, "f", "f", "f"},
						{398111247, "dirty", 16, 9, "f", "f", "f"},
						{458530874, "name", 25, 1, "t", "f", "f"},
						{458530874, "hash", 25, 2, "t", "f", "f"},
						{458530874, "latest_committer", 25, 3, "f", "f", "f"},
						{458530874, "latest_committer_email", 25, 4, "f", "f", "f"},
						{458530874, "latest_commit_date", 1114, 5, "f", "f", "f"},
						{458530874, "latest_commit_message", 25, 6, "f", "f", "f"},
						{458530874, "remote", 25, 7, "f", "f", "f"},
						{458530874, "branch", 25, 8, "f", "f", "f"},
						{458530874, "dirty", 16, 9, "f", "f", "f"},
						{604995978, "is_merging", 16, 1, "t", "f", "f"},
						{604995978, "source", 25, 2, "f", "f", "f"},
						{604995978, "source_commit", 25, 3, "f", "f", "f"},
						{604995978, "target", 25, 4, "f", "f", "f"},
						{604995978, "unmerged_tables", 25, 5, "f", "f", "f"},
						{629684363, "name", 25, 1, "t", "f", "f"},
						{629684363, "type", 21, 2, "t", "f", "f"},
						{629684363, "synopsis", 25, 3, "t", "f", "f"},
						{629684363, "short_description", 25, 4, "t", "f", "f"},
						{629684363, "long_description", 25, 5, "t", "f", "f"},
						{629684363, "arguments", 114, 6, "t", "f", "f"},
						{649632770, "commit_hash", 25, 1, "t", "f", "f"},
						{649632770, "table_name", 25, 2, "t", "f", "f"},
						{649632770, "committer", 25, 3, "t", "f", "f"},
						{649632770, "email", 25, 4, "t", "f", "f"},
						{649632770, "date", 1114, 5, "t", "f", "f"},
						{649632770, "message", 25, 6, "t", "f", "f"},
						{649632770, "data_change", 16, 7, "t", "f", "f"},
						{649632770, "schema_change", 16, 8, "t", "f", "f"},
						{867423409, "name", 25, 1, "t", "f", "f"},
						{867423409, "hash", 25, 2, "t", "f", "f"},
						{867423409, "latest_committer", 25, 3, "f", "f", "f"},
						{867423409, "latest_committer_email", 25, 4, "f", "f", "f"},
						{867423409, "latest_commit_date", 1114, 5, "f", "f", "f"},
						{867423409, "latest_commit_message", 25, 6, "f", "f", "f"},
						{876336553, "to_pk", 23, 1, "f", "f", "f"},
						{876336553, "to_v1", 23, 2, "f", "f", "f"},
						{876336553, "to_commit", 25, 3, "f", "f", "f"},
						{876336553, "to_commit_date", 1114, 4, "f", "f", "f"},
						{876336553, "from_pk", 23, 5, "f", "f", "f"},
						{876336553, "from_v1", 23, 6, "f", "f", "f"},
						{876336553, "from_commit", 25, 7, "f", "f", "f"},
						{876336553, "from_commit_date", 1114, 8, "f", "f", "f"},
						{876336553, "diff_type", 25, 9, "f", "f", "f"},
						{887648921, "name", 25, 1, "t", "f", "f"},
						{887648921, "hash", 25, 2, "t", "f", "f"},
						{887648921, "latest_committer", 25, 3, "f", "f", "f"},
						{887648921, "latest_committer_email", 25, 4, "f", "f", "f"},
						{887648921, "latest_commit_date", 1114, 5, "f", "f", "f"},
						{887648921, "latest_commit_message", 25, 6, "f", "f", "f"},
						{909123395, "commit_hash", 25, 1, "t", "f", "f"},
						{909123395, "committer", 25, 2, "t", "f", "f"},
						{909123395, "email", 25, 3, "t", "f", "f"},
						{909123395, "date", 1114, 4, "t", "f", "f"},
						{909123395, "message", 25, 5, "t", "f", "f"},
						{909123395, "commit_order", 20, 6, "t", "f", "f"},
						{1060579466, "table_name", 25, 1, "t", "f", "f"},
						{1060579466, "staged", 16, 2, "t", "f", "f"},
						{1060579466, "status", 25, 3, "t", "f", "f"},
						{1167248682, "name", 25, 1, "t", "f", "f"},
						{1167248682, "url", 25, 2, "t", "f", "f"},
						{1212681264, "commit_hash", 25, 1, "t", "f", "f"},
						{1212681264, "table_name", 25, 2, "t", "f", "f"},
						{1212681264, "column_name", 25, 3, "t", "f", "f"},
						{1212681264, "committer", 25, 4, "t", "f", "f"},
						{1212681264, "email", 25, 5, "t", "f", "f"},
						{1212681264, "date", 1114, 6, "t", "f", "f"},
						{1212681264, "message", 25, 7, "t", "f", "f"},
						{1212681264, "diff_type", 25, 8, "t", "f", "f"},
						{1218627310, "table", 25, 1, "t", "f", "f"},
						{1218627310, "num_conflicts", 20, 2, "t", "f", "f"},
						{1322753784, "table", 25, 1, "t", "f", "f"},
						{1322753784, "num_violations", 20, 2, "t", "f", "f"},
						{1384122262, "name", 25, 1, "t", "f", "f"},
						{1384122262, "stash_id", 25, 2, "t", "f", "f"},
						{1384122262, "branch", 25, 3, "t", "f", "f"},
						{1384122262, "hash", 25, 4, "t", "f", "f"},
						{1384122262, "commit_message", 25, 5, "f", "f", "f"},
						{1555944102, "commit_hash", 25, 1, "t", "f", "f"},
						{1555944102, "parent_hash", 25, 2, "t", "f", "f"},
						{1555944102, "parent_index", 23, 3, "t", "f", "f"},
						{1640933374, "id", 20, 1, "t", "f", "f"},
						{1640933374, "staged", 16, 2, "t", "f", "f"},
						{1640933374, "diff_type", 25, 3, "t", "f", "f"},
						{1640933374, "to_pk", 23, 4, "f", "f", "f"},
						{1640933374, "to_v1", 23, 5, "f", "f", "f"},
						{1640933374, "from_pk", 23, 6, "f", "f", "f"},
						{1640933374, "from_v1", 23, 7, "f", "f", "f"},
						{1670572237, "name", 25, 1, "t", "f", "f"},
						{1670572237, "url", 25, 2, "t", "f", "f"},
						{1670572237, "fetch_specs", 114, 3, "f", "f", "f"},
						{1670572237, "params", 114, 4, "f", "f", "f"},
						{1763579892, "name", 25, 1, "t", "f", "f"},
						{1763579892, "hash", 25, 2, "t", "f", "f"},
						{1763579892, "latest_committer", 25, 3, "f", "f", "f"},
						{1763579892, "latest_committer_email", 25, 4, "f", "f", "f"},
						{1763579892, "latest_commit_date", 1114, 5, "f", "f", "f"},
						{1763579892, "latest_commit_message", 25, 6, "f", "f", "f"},
						{1763579892, "remote", 25, 7, "f", "f", "f"},
						{1763579892, "branch", 25, 8, "f", "f", "f"},
						{1763579892, "dirty", 16, 9, "f", "f", "f"},
						{1807684176, "tag_name", 25, 1, "t", "f", "f"},
						{1807684176, "tag_hash", 25, 2, "t", "f", "f"},
						{1807684176, "tagger", 25, 3, "t", "f", "f"},
						{1807684176, "email", 25, 4, "t", "f", "f"},
						{1807684176, "date", 1114, 5, "t", "f", "f"},
						{1807684176, "message", 25, 6, "t", "f", "f"},
						{1932298159, "table", 25, 1, "t", "f", "f"},
						{1932298159, "num_violations", 20, 2, "t", "f", "f"},
						{1967026500, "table", 25, 1, "t", "f", "f"},
						{1967026500, "num_violations", 20, 2, "t", "f", "f"},
						{2056815203, "commit_hash", 25, 1, "t", "f", "f"},
						{2056815203, "table_name", 25, 2, "t", "f", "f"},
						{2056815203, "column_name", 25, 3, "t", "f", "f"},
						{2056815203, "committer", 25, 4, "t", "f", "f"},
						{2056815203, "email", 25, 5, "t", "f", "f"},
						{2056815203, "date", 1114, 6, "t", "f", "f"},
						{2056815203, "message", 25, 7, "t", "f", "f"},
						{2056815203, "diff_type", 25, 8, "t", "f", "f"},
						{2067982358, "commit_hash", 25, 1, "t", "f", "f"},
						{2067982358, "committer", 25, 2, "t", "f", "f"},
						{2067982358, "email", 25, 3, "t", "f", "f"},
						{2067982358, "date", 1114, 4, "t", "f", "f"},
						{2067982358, "message", 25, 5, "t", "f", "f"},
						{2067982358, "commit_order", 20, 6, "t", "f", "f"},
						{2357712556, "commit_hash", 25, 1, "t", "f", "f"},
						{2357712556, "table_name", 25, 2, "t", "f", "f"},
						{2357712556, "committer", 25, 3, "t", "f", "f"},
						{2357712556, "email", 25, 4, "t", "f", "f"},
						{2357712556, "date", 1114, 5, "t", "f", "f"},
						{2357712556, "message", 25, 6, "t", "f", "f"},
						{2357712556, "data_change", 16, 7, "t", "f", "f"},
						{2357712556, "schema_change", 16, 8, "t", "f", "f"},
						{2419641880, "from_root_ish", 25, 1, "f", "f", "f"},
						{2419641880, "base_pk", 23, 2, "f", "f", "f"},
						{2419641880, "base_v1", 23, 3, "f", "f", "f"},
						{2419641880, "our_pk", 23, 4, "t", "f", "f"},
						{2419641880, "our_v1", 23, 5, "f", "f", "f"},
						{2419641880, "our_diff_type", 25, 6, "f", "f", "f"},
						{2419641880, "their_pk", 23, 7, "f", "f", "f"},
						{2419641880, "their_v1", 23, 8, "f", "f", "f"},
						{2419641880, "their_diff_type", 25, 9, "f", "f", "f"},
						{2419641880, "dolt_conflict_id", 25, 10, "f", "f", "f"},
						{2517735330, "table", 25, 1, "t", "f", "f"},
						{2517735330, "num_conflicts", 20, 2, "t", "f", "f"},
						{2849341124, "pk", 23, 1, "t", "f", "f"},
						{2849341124, "v1", 23, 2, "f", "f", "f"},
						{3112353516, "commit_hash", 25, 1, "t", "f", "f"},
						{3112353516, "committer", 25, 2, "t", "f", "f"},
						{3112353516, "email", 25, 3, "t", "f", "f"},
						{3112353516, "date", 1114, 4, "t", "f", "f"},
						{3112353516, "message", 25, 5, "t", "f", "f"},
						{3152041833, "commit_hash", 25, 1, "t", "f", "f"},
						{3152041833, "committer", 25, 2, "t", "f", "f"},
						{3152041833, "email", 25, 3, "t", "f", "f"},
						{3152041833, "date", 1114, 4, "t", "f", "f"},
						{3152041833, "message", 25, 5, "t", "f", "f"},
						{3210116770, "table_name", 25, 1, "t", "f", "f"},
						{3210116770, "base_schema", 25, 2, "t", "f", "f"},
						{3210116770, "our_schema", 25, 3, "t", "f", "f"},
						{3210116770, "their_schema", 25, 4, "t", "f", "f"},
						{3210116770, "description", 25, 5, "t", "f", "f"},
						{3246414078, "tag_name", 25, 1, "t", "f", "f"},
						{3246414078, "tag_hash", 25, 2, "t", "f", "f"},
						{3246414078, "tagger", 25, 3, "t", "f", "f"},
						{3246414078, "email", 25, 4, "t", "f", "f"},
						{3246414078, "date", 1114, 5, "t", "f", "f"},
						{3246414078, "message", 25, 6, "t", "f", "f"},
						{3390577184, "from_root_ish", 25, 1, "f", "f", "f"},
						{3390577184, "violation_type", 1043, 2, "t", "f", "f"},
						{3390577184, "pk", 23, 3, "t", "f", "f"},
						{3390577184, "v1", 23, 4, "f", "f", "f"},
						{3390577184, "violation_info", 114, 5, "f", "f", "f"},
						{3422698383, "pk", 23, 1, "t", "f", "f"},
						{3422698383, "v1", 23, 2, "f", "f", "f"},
						{3422698383, "commit_hash", 25, 3, "t", "f", "f"},
						{3422698383, "committer", 25, 4, "t", "f", "f"},
						{3422698383, "commit_date", 1114, 5, "t", "f", "f"},
						{3425483043, "commit_hash", 25, 1, "t", "f", "f"},
						{3425483043, "committer", 25, 2, "t", "f", "f"},
						{3425483043, "email", 25, 3, "t", "f", "f"},
						{3425483043, "date", 1114, 4, "t", "f", "f"},
						{3425483043, "message", 25, 5, "t", "f", "f"},
						{3431637196, "table_name", 25, 1, "t", "f", "f"},
						{3431637196, "staged", 16, 2, "t", "f", "f"},
						{3431637196, "status", 25, 3, "t", "f", "f"},
						{3491847678, "commit_hash", 25, 1, "t", "f", "f"},
						{3491847678, "committer", 25, 2, "t", "f", "f"},
						{3491847678, "email", 25, 3, "t", "f", "f"},
						{3491847678, "date", 1114, 4, "t", "f", "f"},
						{3491847678, "message", 25, 5, "t", "f", "f"},
						{3491847678, "commit_order", 20, 6, "t", "f", "f"},
						{3554775706, "table_name", 25, 1, "t", "f", "f"},
						{3554775706, "staged", 16, 2, "t", "f", "f"},
						{3554775706, "status", 25, 3, "t", "f", "f"},
						{3947121936, "is_merging", 16, 1, "t", "f", "f"},
						{3947121936, "source", 25, 2, "f", "f", "f"},
						{3947121936, "source_commit", 25, 3, "f", "f", "f"},
						{3947121936, "target", 25, 4, "f", "f", "f"},
						{3947121936, "unmerged_tables", 25, 5, "f", "f", "f"},
						{4001633963, "commit_hash", 25, 1, "t", "f", "f"},
						{4001633963, "parent_hash", 25, 2, "t", "f", "f"},
						{4001633963, "parent_index", 23, 3, "t", "f", "f"},
						{4126412490, "commit_hash", 25, 1, "t", "f", "f"},
						{4126412490, "parent_hash", 25, 2, "t", "f", "f"},
						{4126412490, "parent_index", 23, 3, "t", "f", "f"},
					},
				},
			},
		},
	})
}

func TestPgAttributeIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "pg_attribute indexes",
			SetUpScript: []string{
				`CREATE SCHEMA test_schema;`,
				`SET search_path TO test_schema;`,
				`CREATE TABLE test_table (
					id INT PRIMARY KEY,
					name TEXT NOT NULL,
					description VARCHAR(255),
					created_at TIMESTAMP DEFAULT NOW()
				);`,
				`CREATE TABLE another_table (
					pk BIGINT PRIMARY KEY,
					value TEXT
				);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// Test index on attrelid (non-unique index) using JOIN instead of regclass
					Query: `SELECT a.attname, a.attnum FROM pg_catalog.pg_attribute a
							JOIN pg_catalog.pg_class c ON a.attrelid = c.oid 
							WHERE c.relname = 'test_table'
							ORDER BY a.attnum;`,
					Expected: []sql.Row{
						{"id", int16(1)},
						{"name", int16(2)},
						{"description", int16(3)},
						{"created_at", int16(4)},
					},
				},
				{
					// Test unique index on attrelid + attname (using string values for boolean fields)
					Query: `SELECT a.attnum, a.attnotnull, a.atthasdef FROM pg_catalog.pg_attribute a
							JOIN pg_catalog.pg_class c ON a.attrelid = c.oid
							WHERE c.relname = 'test_table' 
							AND a.attname = 'name';`,
					Expected: []sql.Row{
						{int16(2), "t", "f"},
					},
				},
				{
					// Test another unique index lookup
					Query: `SELECT a.attnum FROM pg_catalog.pg_attribute a
							JOIN pg_catalog.pg_class c ON a.attrelid = c.oid
							WHERE c.relname = 'another_table' 
							AND a.attname = 'pk';`,
					Expected: []sql.Row{
						{int16(1)},
					},
				},
				{
					// Test range lookup on attrelid index
					Query: `SELECT COUNT(*) FROM pg_catalog.pg_attribute a
							WHERE a.attrelid IN (
								SELECT oid FROM pg_catalog.pg_class 
								WHERE relname IN ('test_table', 'another_table')
							);`,
					Expected: []sql.Row{
						{6},
					},
				},
				{
					// Test JOIN using the indexes
					Query: `SELECT c.relname, a.attname, a.attnum 
							FROM pg_catalog.pg_class c 
							JOIN pg_catalog.pg_attribute a ON c.oid = a.attrelid 
							WHERE c.relname IN ('test_table', 'another_table') 
							ORDER BY c.relname, a.attnum;`,
					Expected: []sql.Row{
						{"another_table", "pk", int16(1)},
						{"another_table", "value", int16(2)},
						{"test_table", "id", int16(1)},
						{"test_table", "name", int16(2)},
						{"test_table", "description", int16(3)},
						{"test_table", "created_at", int16(4)},
					},
				},
			},
		},
	})
}
