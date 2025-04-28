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
	})
}
