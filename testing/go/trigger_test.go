// Copyright 2025 Dolthub, Inc.
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

func TestCreateTrigger(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "BEFORE INSERT",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, "hi_1"},
						{2, "there_2"},
					},
				},
			},
		},
		{
			Name: "BEFORE UPDATE",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				"INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
			BEGIN
				NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
				RETURN NEW;
			END;
			$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "UPDATE test SET v1 = v1 || '|' WHERE pk IN (1, 2);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, "hi|_1"},
						{2, "there|_2"},
					},
				},
			},
		},
		{
			Name: "BEFORE DELETE",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				"CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
				"INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
			BEGIN
				INSERT INTO test2 VALUES (OLD.pk, OLD.v1);
				RETURN OLD;
			END;
			$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE DELETE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, "hi"},
						{2, "there"},
					},
				},
				{
					Query:    "SELECT * FROM test2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DELETE FROM test WHERE pk = 1;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{2, "there"},
					},
				},
				{
					Query: "SELECT * FROM test2;",
					Expected: []sql.Row{
						{1, "hi"},
					},
				},
			},
		},
		{
			Name: "BEFORE INSERT returning NULL",
			Skip: true, // TODO: returning a NULL-filled row isn't quite valid for this
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				"CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
					INSERT INTO test2 VALUES (NEW.pk, NEW.v1);
					RETURN NULL;
				END;
				$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, "hi"},
						{2, "there"},
					},
				},
				{
					Query: "SELECT * FROM test2;",
					Expected: []sql.Row{
						{1, "hi_1"},
						{2, "there_2"},
					},
				},
			},
		},
		{
			Name: "BEFORE UPDATE returning NULL",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				"CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
				"INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
			BEGIN
				NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
				INSERT INTO test2 VALUES (NEW.pk, NEW.v1);
				RETURN NULL;
			END;
			$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "UPDATE test SET v1 = v1 || '|' WHERE pk IN (1, 2);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, "hi"},
						{2, "there"},
					},
				},
				{
					Query: "SELECT * FROM test2;",
					Expected: []sql.Row{
						{1, "hi|_1"},
						{2, "there|_2"},
					},
				},
			},
		},
		{
			Name: "BEFORE DELETE returning NULL",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				"CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
				"INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
			BEGIN
				INSERT INTO test2 VALUES (OLD.pk, OLD.v1);
				RETURN NULL;
			END;
			$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE DELETE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, "hi"},
						{2, "there"},
					},
				},
				{
					Query:    "SELECT * FROM test2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DELETE FROM test WHERE pk = 1;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, "hi"},
						{2, "there"},
					},
				},
				{
					Query: "SELECT * FROM test2;",
					Expected: []sql.Row{
						{1, "hi"},
					},
				},
			},
		},
		{
			Name: "BEFORE UPDATE with DELETE DML",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				"CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
				"INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
			BEGIN
				INSERT INTO test2 VALUES (OLD.pk, OLD.v1);
				RETURN OLD;
			END;
			$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, "hi"},
						{2, "there"},
					},
				},
				{
					Query:    "SELECT * FROM test2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DELETE FROM test WHERE pk = 1;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{2, "there"},
					},
				},
				{
					Query:    "SELECT * FROM test2;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "AFTER INSERT",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				"CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
					INSERT INTO test2 VALUES (NEW.pk, NEW.v1);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger AFTER INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, "hi"},
						{2, "there"},
					},
				},
				{
					Query: "SELECT * FROM test2;",
					Expected: []sql.Row{
						{1, "hi_1"},
						{2, "there_2"},
					},
				},
			},
		},
		{
			Name: "AFTER UPDATE",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				"CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
				"INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
			BEGIN
				NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
				INSERT INTO test2 VALUES (NEW.pk, NEW.v1);
				RETURN NEW;
			END;
			$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger AFTER UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "UPDATE test SET v1 = v1 || '|' WHERE pk IN (1, 2);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, "hi|"},
						{2, "there|"},
					},
				},
				{
					Query: "SELECT * FROM test2;",
					Expected: []sql.Row{
						{1, "hi|_1"},
						{2, "there|_2"},
					},
				},
			},
		},
		{
			Name: "AFTER DELETE returning NULL",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				"CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
				"INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
			BEGIN
				INSERT INTO test2 VALUES (OLD.pk, OLD.v1);
				RETURN NULL;
			END;
			$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger AFTER DELETE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, "hi"},
						{2, "there"},
					},
				},
				{
					Query:    "SELECT * FROM test2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DELETE FROM test WHERE pk = 1;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{2, "there"},
					},
				},
				{
					Query: "SELECT * FROM test2;",
					Expected: []sql.Row{
						{1, "hi"},
					},
				},
			},
		},
		{
			Name: "Cascading DELETE into INSERT, different tables",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				"CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
				"INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
BEGIN
	INSERT INTO test2 VALUES (OLD.pk, OLD.v1);
	RETURN OLD;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE FUNCTION trigger_func2() RETURNS TRIGGER AS $$
BEGIN
	NEW.pk := NEW.pk + 100;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE DELETE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
				`CREATE TRIGGER test_trigger BEFORE INSERT ON test2 FOR EACH ROW EXECUTE FUNCTION trigger_func2();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, "hi"},
						{2, "there"},
					},
				},
				{
					Query:    "SELECT * FROM test2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DELETE FROM test WHERE pk = 1;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{2, "there"},
					},
				},
				{
					Query: "SELECT * FROM test2;",
					Expected: []sql.Row{
						{101, "hi"},
					},
				},
			},
		},
		{
			Name: "Cascading INSERT into UPDATE, same table",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
BEGIN
	UPDATE test SET v1 = v1 || NEW.pk::text;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE FUNCTION trigger_func2() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.v1 || '_u';
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
				`CREATE TRIGGER test_trigger2 BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func2();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test ORDER BY pk;",
					Skip:  true, // TODO: the UPDATE cannot see the table's contents until the INSERT has completely finished
					Expected: []sql.Row{
						{1, "hi2_u"},
						{2, "there"},
					},
				},
			},
		},
		{
			Name: "Multiple triggers on same table",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				`CREATE FUNCTION trigger_func_a() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.v1 || 'a';
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE FUNCTION trigger_func_c() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.v1 || 'c';
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE FUNCTION trigger_func_b() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.v1 || 'b';
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger_b BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func_b();`,
				`CREATE TRIGGER test_trigger_a BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func_a();`,
				`CREATE TRIGGER test_trigger_c BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func_c();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test ORDER BY pk;",
					Expected: []sql.Row{
						{1, "hiabc"},
						{2, "thereabc"},
					},
				},
			},
		},
		{
			Name: "Stack depth limit exceeded",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				"CREATE TABLE test2 (pk INT PRIMARY KEY, v1 TEXT);",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
BEGIN
	INSERT INTO test2 VALUES (NEW.pk+2, NEW.v1 || '_');
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE FUNCTION trigger_func2() RETURNS TRIGGER AS $$
BEGIN
	INSERT INTO test VALUES (NEW.pk+4, NEW.v1 || '|');
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
				`CREATE TRIGGER test_trigger BEFORE INSERT ON test2 FOR EACH ROW EXECUTE FUNCTION trigger_func2();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
					Skip:        true, // TODO: currently we'll just run until we run out of memory, need to abort before that
					ExpectedErr: "stack depth limit exceeded",
				},
			},
		},
		{
			Name: "DELETE TABLE deletes attached triggers",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.v1 || '_';
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
				`CREATE TRIGGER test_trigger2 BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
					ExpectedErr: "already exists",
				},
				{
					Query:       "CREATE TRIGGER test_trigger2 BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
					ExpectedErr: "already exists",
				},
				{
					Query:    "DROP TABLE test;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TRIGGER test_trigger2 BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "WHEN on BEFORE INSERT",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				`CREATE FUNCTION trigger_func1() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.pk::text || '_' || NEW.v1;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE FUNCTION trigger_func2() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger1 BEFORE INSERT ON test FOR EACH ROW WHEN (NEW.pk < 1) EXECUTE FUNCTION trigger_func1();`,
				`CREATE TRIGGER test_trigger2 BEFORE INSERT ON test FOR EACH ROW WHEN (NEW.pk > 1) EXECUTE FUNCTION trigger_func2();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO test VALUES (0, 'hi'), (1, 'there'), (2, 'dude');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{0, "0_hi"},
						{1, "there"},
						{2, "dude_2"},
					},
				},
			},
		},
		{
			Name: "WHEN with non-boolean expression",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
BEGIN
	NEW.v1 := NEW.pk::text || '_' || NEW.v1;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW WHEN (NEW.pk + 1) EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
					ExpectedErr: "argument of WHEN must be type boolean",
				},
			},
		},
		{
			Name: "Table as type",
			Skip: true, // TODO: figure out why this is not recognizing rec.qty as valid
			SetUpScript: []string{
				`CREATE TABLE test (id INT4 PRIMARY KEY, name TEXT NOT NULL, qty INT4 NOT NULL, price REAL NOT NULL);`,
				`CREATE FUNCTION trigger_func() RETURNS trigger AS $$
DECLARE
	rec test;
BEGIN
	rec := NEW;
	IF rec.qty < 0 THEN
		rec.qty := -rec.qty;
	END IF;
	NEW := rec;
	RETURN NEW;
END; $$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `INSERT INTO test VALUES (1, 'apple', 3, 2.5), (2, 'banana', -5, -1.2);`,
					Expected: []sql.Row{},
				}, {
					Query: `SELECT * FROM test;`,
					Expected: []sql.Row{
						{1, "apple", 3, 2.5},
						{2, "banana", 5, -1.2},
					},
				},
			},
		},
		{
			Name: "DECLARE default referencing the trigger records",
			SetUpScript: []string{
				`CREATE TABLE test (id INT4 PRIMARY KEY, val TEXT);`,
				`CREATE TABLE log (msg TEXT);`,
				`INSERT INTO test VALUES (7, 'a');`,
				`CREATE FUNCTION trigger_func() RETURNS trigger AS $$
DECLARE
	old_id INT4 := OLD.id;
	changed BOOLEAN := OLD.val <> NEW.val;
BEGIN
	INSERT INTO log VALUES ('id=' || old_id || ' changed=' || changed);
	RETURN NEW;
END; $$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `UPDATE test SET val = 'b' WHERE id = 7;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT msg FROM log;`,
					Expected: []sql.Row{{"id=7 changed=true"}},
				},
			},
		},
		{
			Name: "trigger to call procedure that updates another table using dynamic execute",
			SetUpScript: []string{
				`create table public."Collections"(
				 id uuid PRIMARY KEY NOT NULL,
				 name text not null,
				 username varchar(28) not null,
				 total_tracks integer DEFAULT 0);`,
				`INSERT INTO public."Collections" (id, name, username, total_tracks) VALUES ('550e8400-e29b-41d4-a716-446655440000', 'My Custom Playlist', 'user_alpha', 10);`,
				`create table public."CollectionItems"(
			collection_id uuid not null,
			track_id integer not null);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE OR REPLACE FUNCTION update_collections()
  RETURNS trigger AS $$
  DECLARE
    BEGIN
    IF TG_OP = 'INSERT' THEN
      EXECUTE 'update public."Collections" set total_tracks=total_tracks+1 where id = $1;'
      USING NEW.collection_id;
    END IF;

    IF TG_OP = 'DELETE' THEN 
      EXECUTE 'update public."Collections" set total_tracks=total_tracks-1 where id = $1;'
      USING OLD.collection_id;
    END IF;
    
    RETURN NEW;
    END;
$$ LANGUAGE plpgsql;`,
					Expected: []sql.Row{},
				},
				{
					Query: `CREATE TRIGGER update_collection
				AFTER INSERT OR DELETE ON public."CollectionItems"
				FOR EACH ROW EXECUTE PROCEDURE update_collections();`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO public."CollectionItems" (collection_id, track_id) VALUES ('550e8400-e29b-41d4-a716-446655440000', 101);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT total_tracks FROM public."Collections"`,
					Expected: []sql.Row{{11}},
				},
			},
		},
		{
			Name: "DROP TRIGGER",
			SetUpScript: []string{
				`CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);`,
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					NEW.v1 := NEW.v1 || '_' || NEW.pk::text;
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "DROP TRIGGER test_trigger ON test;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP TRIGGER IF EXISTS test_trigger ON test;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "OLD.* IS DISTINCT FROM NEW.* in a trigger",
			SetUpScript: []string{
				"CREATE TABLE t3336_issue (a INT);",
				"INSERT INTO t3336_issue VALUES (1);",
				"CREATE FUNCTION f3336_issue() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE NOTICE 'the trigger ran'; RETURN NEW; END $$;",
				"CREATE TRIGGER tr BEFORE UPDATE ON t3336_issue FOR EACH ROW WHEN (old.* IS DISTINCT FROM new.*) EXECUTE FUNCTION f3336_issue();",
				"CREATE TABLE t3336 (a INT PRIMARY KEY, b TEXT);",
				"CREATE TABLE t3336_log (a INT, src TEXT);",
				"CREATE FUNCTION f3336_when() RETURNS trigger AS $$ BEGIN INSERT INTO t3336_log VALUES (NEW.a, 'when'); RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE FUNCTION f3336_body() RETURNS trigger AS $$ BEGIN IF OLD.* IS DISTINCT FROM NEW.* THEN INSERT INTO t3336_log VALUES (NEW.a, 'body'); END IF; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE TRIGGER tr3336_when AFTER UPDATE ON t3336 FOR EACH ROW WHEN (OLD.* IS DISTINCT FROM NEW.*) EXECUTE FUNCTION f3336_when();",
				"CREATE TRIGGER tr3336_body AFTER UPDATE ON t3336 FOR EACH ROW EXECUTE FUNCTION f3336_body();",
				"INSERT INTO t3336 VALUES (1, 'x'), (2, NULL);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "UPDATE t3336_issue SET a = 2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM t3336_issue;",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT ROW(2, NULL::TEXT) IS DISTINCT FROM ROW(2, 'y'::TEXT), ROW(2, NULL::TEXT) IS DISTINCT FROM ROW(2, NULL::TEXT), ROW(2, NULL::TEXT) IS NOT DISTINCT FROM ROW(2, NULL::TEXT), ROW(1, 2) IS DISTINCT FROM ROW(1, 3), ROW(1, 2) IS NOT DISTINCT FROM ROW(1, 2);",
					Expected: []sql.Row{{"t", "f", "t", "t", "t"}},
				},
				{
					Query:    "UPDATE t3336 SET b = 'x' WHERE a = 1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT COUNT(*) FROM t3336_log;",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "UPDATE t3336 SET b = 'y' WHERE a = 2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "UPDATE t3336 SET b = NULL WHERE a = 2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "UPDATE t3336 SET b = NULL WHERE a = 2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "UPDATE t3336 SET b = 'z' WHERE a = 1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM t3336_log ORDER BY a, src;",
					Expected: []sql.Row{{1, "body"}, {1, "when"}, {2, "body"}, {2, "body"}, {2, "when"}, {2, "when"}},
				},
			},
		},
		{
			Name: "Whole-row references outside of a record comparison are rejected",
			SetUpScript: []string{
				"CREATE TABLE t3336 (a INT PRIMARY KEY, b TEXT);",
				"CREATE TABLE t3336_one (a INT);",
				"CREATE TABLE t3336_bool (b BOOLEAN);",
				"CREATE TABLE t3336_log2 (v TEXT);",
				"INSERT INTO t3336 VALUES (1, 'x');",
				"INSERT INTO t3336_one VALUES (1);",
				"INSERT INTO t3336_bool VALUES (true);",
				"CREATE FUNCTION f3336() RETURNS TRIGGER AS $$ BEGIN RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE FUNCTION f3336_assign() RETURNS TRIGGER AS $$ DECLARE v INT; BEGIN v := NEW.*; INSERT INTO t3336_log2 VALUES ('assign ' || v); RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE FUNCTION f3336_if() RETURNS TRIGGER AS $$ BEGIN IF NEW.* THEN INSERT INTO t3336_log2 VALUES ('if'); END IF; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE FUNCTION f3336_eq() RETURNS TRIGGER AS $$ BEGIN IF NEW.* = 1 THEN RETURN NEW; END IF; RETURN NEW; END; $$ LANGUAGE plpgsql;",
				"CREATE FUNCTION f3336_raise() RETURNS TRIGGER AS $$ BEGIN RAISE EXCEPTION 'val %', NEW.*; END; $$ LANGUAGE plpgsql;",
				"CREATE FUNCTION f3336_ret() RETURNS TRIGGER AS $$ BEGIN RETURN NEW.*; END; $$ LANGUAGE plpgsql;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE TRIGGER tr3336_bad BEFORE UPDATE ON t3336 FOR EACH ROW WHEN (OLD.*) EXECUTE FUNCTION f3336();",
					Expected: []sql.Row{},
				},
				{
					Query:       "UPDATE t3336 SET b = 'q' WHERE a = 1;",
					ExpectedErr: "argument of WHEN must be type boolean",
				},
				{
					Query:    "DROP TRIGGER tr3336_bad ON t3336;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TRIGGER tr3336_bad2 BEFORE UPDATE ON t3336 FOR EACH ROW WHEN (OLD.* = 1) EXECUTE FUNCTION f3336();",
					Expected: []sql.Row{},
				},
				{
					Query:       "UPDATE t3336 SET b = 'q' WHERE a = 1;",
					ExpectedErr: "operator does not exist",
				},
				{
					Query:    "DROP TRIGGER tr3336_bad2 ON t3336;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TRIGGER tr3336_assign BEFORE UPDATE ON t3336 FOR EACH ROW EXECUTE FUNCTION f3336_assign();",
					Expected: []sql.Row{},
				},
				{
					Query:       "UPDATE t3336 SET b = 'q' WHERE a = 1;",
					ExpectedErr: "assignment source returned 2 columns",
				},
				{
					Query:    "DROP TRIGGER tr3336_assign ON t3336;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TRIGGER tr3336_if BEFORE UPDATE ON t3336 FOR EACH ROW EXECUTE FUNCTION f3336_if();",
					Expected: []sql.Row{},
				},
				{
					Query:       "UPDATE t3336 SET b = 'q' WHERE a = 1;",
					ExpectedErr: "query returned 2 columns",
				},
				{
					Query:    "DROP TRIGGER tr3336_if ON t3336;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TRIGGER tr3336_eq BEFORE UPDATE ON t3336 FOR EACH ROW EXECUTE FUNCTION f3336_eq();",
					Expected: []sql.Row{},
				},
				{
					Query:       "UPDATE t3336 SET b = 'q' WHERE a = 1;",
					ExpectedErr: "operator does not exist",
				},
				{
					Query:    "DROP TRIGGER tr3336_eq ON t3336;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TRIGGER tr3336_raise BEFORE UPDATE ON t3336 FOR EACH ROW EXECUTE FUNCTION f3336_raise();",
					Expected: []sql.Row{},
				},
				{
					Query:       "UPDATE t3336 SET b = 'q' WHERE a = 1;",
					ExpectedErr: "query returned 2 columns",
				},
				{
					Query:    "DROP TRIGGER tr3336_raise ON t3336;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TRIGGER tr3336_ret BEFORE UPDATE ON t3336 FOR EACH ROW EXECUTE FUNCTION f3336_ret();",
					Expected: []sql.Row{},
				},
				{
					Query:       "UPDATE t3336 SET b = 'q' WHERE a = 1;",
					ExpectedErr: "query returned 2 columns",
				},
				{
					Query:    "DROP TRIGGER tr3336_ret ON t3336;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM t3336;",
					Expected: []sql.Row{{1, "x"}},
				},
				{
					Query:    "CREATE TRIGGER tr3336_assign1 BEFORE UPDATE ON t3336_one FOR EACH ROW EXECUTE FUNCTION f3336_assign();",
					Expected: []sql.Row{},
				},
				{
					Query:    "UPDATE t3336_one SET a = 2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP TRIGGER tr3336_assign1 ON t3336_one;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TRIGGER tr3336_raise1 BEFORE UPDATE ON t3336_one FOR EACH ROW EXECUTE FUNCTION f3336_raise();",
					Expected: []sql.Row{},
				},
				{
					Query:       "UPDATE t3336_one SET a = 3;",
					ExpectedErr: "val 3",
				},
				{
					Query:    "DROP TRIGGER tr3336_raise1 ON t3336_one;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE TRIGGER tr3336_if1 BEFORE UPDATE ON t3336_bool FOR EACH ROW EXECUTE FUNCTION f3336_if();",
					Expected: []sql.Row{},
				},
				{
					Query:    "UPDATE t3336_bool SET b = true;",
					Expected: []sql.Row{},
				},
				{
					Query:    "UPDATE t3336_bool SET b = false;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DROP TRIGGER tr3336_if1 ON t3336_bool;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM t3336_log2 ORDER BY v;",
					Expected: []sql.Row{{"assign 2"}, {"if"}},
				},
				{
					Query:    "SELECT * FROM t3336_one;",
					Expected: []sql.Row{{2}},
				},
			},
		},
	})
}

// TestTriggerWholeRecordReference covers a trigger body that references NEW or OLD as a whole rather than a
// field of it, which is what passing the row to a function such as to_jsonb() does.
func TestTriggerWholeRecordReference(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "whole record passed to a function",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				"CREATE TABLE log (id SERIAL PRIMARY KEY, which TEXT, j JSONB);",
				"INSERT INTO test VALUES (1, 'hi');",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				DECLARE
					old_v jsonb;
					new_v jsonb;
				BEGIN
					old_v := to_jsonb(OLD);
					new_v := to_jsonb(NEW);
					INSERT INTO log (which, j) VALUES ('old', old_v), ('new', new_v);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "UPDATE test SET v1 = 'bye' WHERE pk = 1;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT which, j::text FROM log ORDER BY id;",
					Expected: []sql.Row{
						{"old", `{"pk": 1, "v1": "hi"}`},
						{"new", `{"pk": 1, "v1": "bye"}`},
					},
				},
			},
		},
		{
			Name: "whole record as text",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT, b BOOL);",
				"CREATE TABLE log (id SERIAL PRIMARY KEY, t TEXT);",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log (t) VALUES (NEW::text);
					RAISE NOTICE 'row: %', NEW;
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           `INSERT INTO test VALUES (1, 'a,b"c', true);`,
					Expected:        []sql.Row{},
					ExpectedNotices: []ExpectedNotice{{Severity: "NOTICE", Message: `row: (1,"a,b\"c",t)`}},
				},
				{
					Query:    "SELECT t FROM log ORDER BY id;",
					Expected: []sql.Row{{`(1,"a,b\"c",t)`}},
				},
			},
		},
		{
			Name: "whole record with a NULL field",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				"CREATE TABLE log (id SERIAL PRIMARY KEY, j TEXT, t TEXT);",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log (j, t) VALUES (to_jsonb(NEW)::text, NEW::text);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO test VALUES (1, NULL);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT j, t FROM log ORDER BY id;",
					Expected: []sql.Row{{`{"pk": 1, "v1": null}`, "(1,)"}},
				},
			},
		},
		{
			Name: "whole record whose field is named like the record",
			SetUpScript: []string{
				// A field named `record` collides with the alias the row is rendered under.
				"CREATE TABLE test (pk INT PRIMARY KEY, record TEXT);",
				"CREATE TABLE log (id SERIAL PRIMARY KEY, j TEXT);",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log (j) VALUES (to_jsonb(NEW)::text);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO test VALUES (1, 'hi');",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT j FROM log ORDER BY id;",
					Expected: []sql.Row{{`{"pk": 1, "record": "hi"}`}},
				},
			},
		},
		{
			Name: "records compared as a whole",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				"INSERT INTO test VALUES (1, 'hi'), (2, 'there');",
				"CREATE TABLE log (id SERIAL PRIMARY KEY, msg TEXT);",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					IF NEW IS DISTINCT FROM OLD THEN
						INSERT INTO log (msg) VALUES ('changed ' || NEW.pk::text);
					ELSE
						INSERT INTO log (msg) VALUES ('same ' || NEW.pk::text);
					END IF;
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "UPDATE test SET v1 = 'hi' WHERE pk IN (1, 2);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT msg FROM log ORDER BY id;",
					Expected: []sql.Row{{"same 1"}, {"changed 2"}},
				},
			},
		},
		{
			Name: "records compared as a whole when a field leaves NULL",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT, v2 TEXT);",
				"INSERT INTO test VALUES (1, 'hi', NULL);",
				"CREATE TABLE log (id SERIAL PRIMARY KEY, msg TEXT);",
				`CREATE FUNCTION trigger_func() RETURNS TRIGGER AS $$
				BEGIN
					IF OLD IS DISTINCT FROM NEW THEN
						INSERT INTO log (msg) VALUES ('changed ' || NEW.pk::text);
					ELSE
						INSERT INTO log (msg) VALUES ('same ' || NEW.pk::text);
					END IF;
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER test_trigger BEFORE UPDATE ON test FOR EACH ROW EXECUTE FUNCTION trigger_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "UPDATE test SET v1 = 'bye' WHERE pk = 1;",
					Expected: []sql.Row{},
				},
				{
					// A field going from NULL to a value is still a difference between the two records.
					Query:    "UPDATE test SET v2 = 'now set' WHERE pk = 1;",
					Expected: []sql.Row{},
				},
				{
					// And going back to NULL.
					Query:    "UPDATE test SET v2 = NULL WHERE pk = 1;",
					Expected: []sql.Row{},
				},
				{
					// Both records having NULL in the same field is not a difference.
					Query:    "UPDATE test SET v2 = NULL WHERE pk = 1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT msg FROM log ORDER BY id;",
					Expected: []sql.Row{{"changed 1"}, {"changed 1"}, {"changed 1"}, {"same 1"}},
				},
			},
		},
		{
			Name: "the record an operation does not supply",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT PRIMARY KEY, v1 TEXT);",
				"CREATE TABLE log (id SERIAL PRIMARY KEY, o TEXT, n TEXT);",
				`CREATE FUNCTION insert_func() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log (o, n) VALUES (to_jsonb(OLD)::text, to_jsonb(NEW)::text);
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;`,
				`CREATE FUNCTION delete_func() RETURNS TRIGGER AS $$
				BEGIN
					INSERT INTO log (o, n) VALUES (to_jsonb(OLD)::text, to_jsonb(NEW)::text);
					RETURN OLD;
				END;
				$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER t1 BEFORE INSERT ON test FOR EACH ROW EXECUTE FUNCTION insert_func();`,
				`CREATE TRIGGER t2 BEFORE DELETE ON test FOR EACH ROW EXECUTE FUNCTION delete_func();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO test VALUES (1, 'hi');",
					Expected: []sql.Row{},
				},
				{
					Query:    "DELETE FROM test WHERE pk = 1;",
					Expected: []sql.Row{},
				},
				{
					// TODO: PostgreSQL leaves OLD unassigned for an INSERT and NEW for a DELETE, so
					//  neither yields an object of NULL fields there. Triggers here give both records
					//  the table's shape whatever the operation; see TriggerCall.
					Query: "SELECT o, n FROM log ORDER BY id;",
					Expected: []sql.Row{
						{`{"pk": null, "v1": null}`, `{"pk": 1, "v1": "hi"}`},
						{`{"pk": 1, "v1": "hi"}`, `{"pk": null, "v1": null}`},
					},
				},
			},
		},
	})
}
