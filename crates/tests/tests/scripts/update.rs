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

use harness::oid::*;
use harness::pgx::Time;
use harness::plan::PlanFact;
use harness::script::Cell::{Any, Null, Text as T};
use harness::script::{A, BindVar, Column, Diagnostic, E, Expected, Flow, N, S, ScriptTest, ScriptTestAssertion, USER_DEFINED, run_scripts, run_scripts_repeated};
use harness::wire::{Datum, F, Field, Fields, PGX_STARTUP, Receive, Send, Step, W, WireTest, run_wire_tests};

#[test]
fn test_update() {
    run_scripts(&[
        ScriptTest {
            name: "simple update",
            set_up_script: &[
                "CREATE TABLE t1 (a INT PRIMARY KEY, b INT)",
                "INSERT INTO t1 VALUES (1, 2), (2, 3), (3, 4)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t1 SET b = 5 WHERE a = 2",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t1 where a =  2",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("2"), T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "update to default",
            set_up_script: &[
                "create table t (i int default 10, j varchar(128) default (concat('abc', 'def')));",
                "insert into t values (100, 'a'), (200, 'b');",
                "create table t2 (i int);",
                "insert into t2 values (1), (2), (3);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "update t set i = default where i = 100;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from t order by i",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4), Column("j", VARCHAR)],
                        rows: &[
                            &[T("10"), T("a")],
                            &[T("200"), T("b")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "update t set j = default where i = 200;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from t order by i",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4), Column("j", VARCHAR)],
                        rows: &[
                            &[T("10"), T("a")],
                            &[T("200"), T("abcdef")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "update t set i = default, j = default;",
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from t order by i",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4), Column("j", VARCHAR)],
                        rows: &[
                            &[T("10"), T("abcdef")],
                            &[T("10"), T("abcdef")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "update t2 set i = default",
                    expected: Expected::Tag("UPDATE 3"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "select * from t2",
                    expected: Expected::Rows {
                        columns: &[Column("i", INT4)],
                        rows: &[
                            &[Null],
                            &[Null],
                            &[Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UPDATE ... RETURNING",
            set_up_script: &[
                "CREATE TABLE t (pk INT PRIMARY KEY, c1 TEXT);",
                "INSERT INTO t VALUES (1, 'one');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t SET pk = pk+1, c1 = '42' RETURNING c1, pk, pk * 2;",
                    expected: Expected::Rows {
                        columns: &[Column("c1", TEXT), Column("pk", INT4), Column("?column?", INT4)],
                        rows: &[
                            &[T("42"), T("2"), T("4")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET c1 = '43' RETURNING *;",
                    expected: Expected::Rows {
                        columns: &[Column("pk", INT4), Column("c1", TEXT)],
                        rows: &[
                            &[T("2"), T("43")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UPDATE ... RETURNING with join",
            set_up_script: &[
                "CREATE TABLE employees (id SERIAL PRIMARY KEY, name TEXT, department_id INT, salary INT);",
                "CREATE TABLE departments (id SERIAL PRIMARY KEY, name TEXT, bonus INT);",
                "INSERT INTO employees (name, department_id, salary) VALUES ('Alice', 1, 50000), ('Bob', 2, 60000);",
                "INSERT INTO departments (name, bonus) VALUES ('Engineering', 5000), ('Marketing', 3000);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE employees e SET salary = salary + d.bonus FROM departments d WHERE e.department_id = d.id RETURNING e.id, e.name, e.salary;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT), Column("salary", INT4)],
                        rows: &[
                            &[T("1"), T("Alice"), T("55000")],
                            &[T("2"), T("Bob"), T("63000")],
                        ],
                        tag: "UPDATE 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UPDATE with join on subquery",
            set_up_script: &[
                "CREATE TABLE employees (id INT PRIMARY KEY, name TEXT, department_id INT, salary INT);",
                "CREATE TABLE departments (id INT PRIMARY KEY, name TEXT, bonus INT);",
                "INSERT INTO employees VALUES (1, 'Alice', 10, 50000), (2, 'Bob', 20, 60000);",
                "INSERT INTO departments VALUES (10, 'Engineering', 5000), (20, 'HR', 3000);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
			UPDATE employees SET salary = salary + dept_bonus.bonus
			FROM ( SELECT id, bonus FROM departments ) AS dept_bonus
			WHERE employees.department_id = dept_bonus.id AND employees.name = 'Alice';"#,
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT salary FROM employees WHERE name = 'Alice';",
                    expected: Expected::Rows {
                        columns: &[Column("salary", INT4)],
                        rows: &[
                            &[T("55000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UPDATE with join on one table",
            set_up_script: &[
                "CREATE TABLE products (id SERIAL PRIMARY KEY, name TEXT, price INT, category_id INT);",
                "CREATE TABLE categories (id SERIAL PRIMARY KEY, name TEXT, discount INT);",
                "INSERT INTO products (name, price, category_id) VALUES ('Laptop', 1000, 1), ('Phone', 800, 2);",
                "INSERT INTO categories (name, discount) VALUES ('Electronics', 100), ('Mobiles', 50);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE products p SET price = price - c.discount FROM categories c WHERE p.category_id = c.id;",
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, name, price FROM products ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("name", TEXT), Column("price", INT4)],
                        rows: &[
                            &[T("1"), T("Laptop"), T("900")],
                            &[T("2"), T("Phone"), T("750")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UPDATE with join on two tables",
            set_up_script: &[
                "CREATE TABLE books (id SERIAL PRIMARY KEY, title TEXT, price INT, author_id INT, publisher_id INT);",
                "CREATE TABLE authors (id SERIAL PRIMARY KEY, royalty INT);",
                "CREATE TABLE publishers (id SERIAL PRIMARY KEY, markup INT);",
                "INSERT INTO books (title, price, author_id, publisher_id) VALUES ('Book A', 100, 1, 1), ('Book B', 120, 2, 2);",
                "INSERT INTO authors (royalty) VALUES (10), (20);",
                "INSERT INTO publishers (markup) VALUES (15), (25);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE books b SET price = price + a.royalty - p.markup FROM authors a, publishers p WHERE b.author_id = a.id AND b.publisher_id = p.id;",
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, title, price FROM books ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("title", TEXT), Column("price", INT4)],
                        rows: &[
                            &[T("1"), T("Book A"), T("95")],
                            &[T("2"), T("Book B"), T("115")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UPDATE with join on three tables",
            set_up_script: &[
                "CREATE TABLE orders (id SERIAL PRIMARY KEY, customer_id INT, product_id INT, total INT);",
                "CREATE TABLE customers (id SERIAL PRIMARY KEY, loyalty_discount INT);",
                "CREATE TABLE products (id SERIAL PRIMARY KEY, base_price INT, tax_id INT);",
                "CREATE TABLE taxes (id SERIAL PRIMARY KEY, rate INT);",
                "INSERT INTO orders (customer_id, product_id, total) VALUES (1, 1, 0), (2, 2, 0);",
                "INSERT INTO customers (loyalty_discount) VALUES (5), (10);",
                "INSERT INTO products (base_price, tax_id) VALUES (100, 1), (200, 2);",
                "INSERT INTO taxes (rate) VALUES (10), (20);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
				UPDATE orders o
				SET total = p.base_price + (p.base_price * t.rate / 100) - c.loyalty_discount
				FROM customers c, products p, taxes t
				WHERE o.customer_id = c.id AND o.product_id = p.id AND p.tax_id = t.id;
			"#,
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, total FROM orders ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("total", INT4)],
                        rows: &[
                            &[T("1"), T("105")],
                            &[T("2"), T("230")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UPDATE with join on four tables",
            set_up_script: &[
                "CREATE TABLE rentals (id SERIAL PRIMARY KEY, vehicle_id INT, user_id INT, total_cost INT);",
                "CREATE TABLE vehicles (id SERIAL PRIMARY KEY, base_rate INT, fuel_type_id INT);",
                "CREATE TABLE users (id SERIAL PRIMARY KEY, membership_level_id INT);",
                "CREATE TABLE fuel_types (id SERIAL PRIMARY KEY, surcharge INT);",
                "CREATE TABLE membership_levels (id SERIAL PRIMARY KEY, discount INT);",
                "INSERT INTO rentals (vehicle_id, user_id, total_cost) VALUES (1, 1, 0), (2, 2, 0);",
                "INSERT INTO vehicles (base_rate, fuel_type_id) VALUES (300, 1), (400, 2);",
                "INSERT INTO fuel_types (surcharge) VALUES (20), (40);",
                "INSERT INTO users (membership_level_id) VALUES (1), (2);",
                "INSERT INTO membership_levels (discount) VALUES (50), (80);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"
				UPDATE rentals r
				SET total_cost = v.base_rate + f.surcharge - m.discount
				FROM vehicles v, fuel_types f, users u, membership_levels m
				WHERE r.vehicle_id = v.id AND v.fuel_type_id = f.id AND r.user_id = u.id AND u.membership_level_id = m.id;
			"#,
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, total_cost FROM rentals ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("total_cost", INT4)],
                        rows: &[
                            &[T("1"), T("270")],
                            &[T("2"), T("360")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UPDATE with join on table with trigger",
            set_up_script: &[
                r#"CREATE TABLE departments (id SERIAL PRIMARY KEY, name TEXT, bonus INT
);"#,
                "CREATE TABLE employees (id SERIAL PRIMARY KEY, name TEXT, department_id INT REFERENCES departments(id), salary INT);",
                "INSERT INTO departments (name, bonus) VALUES ('Engineering', 1000), ('HR', 500);",
                "INSERT INTO employees (name, department_id, salary) VALUES ('Alice', 1, 50000), ('Bob', 2, 45000);",
                "CREATE TABLE salary_log (employee_id INT, old_salary INT, new_salary INT);",
                r#"CREATE OR REPLACE FUNCTION log_salary_change()
					RETURNS TRIGGER AS $$
					BEGIN
						IF NEW.salary != OLD.salary THEN
							INSERT INTO salary_log VALUES (OLD.id, OLD.salary, NEW.salary);
						END IF;
						RETURN NEW;
					END;
					$$ LANGUAGE plpgsql;"#,
                r#"CREATE TRIGGER trg_log_salary_change
					AFTER UPDATE ON employees
					FOR EACH ROW
					EXECUTE FUNCTION log_salary_change();"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE employees e SET salary = salary + d.bonus FROM departments d WHERE e.department_id = d.id;",
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM salary_log;",
                    expected: Expected::Rows {
                        columns: &[Column("employee_id", INT4), Column("old_salary", INT4), Column("new_salary", INT4)],
                        rows: &[
                            &[T("1"), T("50000"), T("51000")],
                            &[T("2"), T("45000"), T("45500")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_update_affected_rows() {
    run_scripts(&[
        ScriptTest {
            name: "UPDATE and ON CONFLICT command counts",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY, a INT)",
                "INSERT INTO t VALUES (1, 10), (2, 20)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t SET a = a WHERE id = 1",
                    expected: Expected::Tag("UPDATE 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a = 10",
                    expected: Expected::Tag("UPDATE 2"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a = 10 WHERE id = 999",
                    expected: Expected::Tag("UPDATE 0"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a = $1 WHERE id = $2",
                    bind_vars: &[BindVar::Int(10), BindVar::Int(1)],
                    expected: Expected::Tag("UPDATE 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a = a WHERE id = 1 RETURNING id, a",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a = a WHERE id = 1 RETURNING id, a",
                    expected: Expected::Tag("UPDATE 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, 10) ON CONFLICT (id) DO UPDATE SET a = EXCLUDED.a",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, 11) ON CONFLICT (id) DO UPDATE SET a = EXCLUDED.a",
                    expected: Expected::Tag("INSERT 0 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, 12) ON CONFLICT (id) DO UPDATE SET a = EXCLUDED.a WHERE false",
                    expected: Expected::Tag("INSERT 0 0"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "mixed insert and conflict update counts",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY, a INT)",
                "INSERT INTO t VALUES (1, 10), (2, 20), (3, 30)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "INSERT INTO t VALUES (1, 10), (2, 22), (3, 33), (4, 40) ON CONFLICT (id) DO UPDATE SET a = EXCLUDED.a WHERE t.id <> 3",
                    expected: Expected::Tag("INSERT 0 3"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, a FROM t ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("22")],
                            &[T("3"), T("30")],
                            &[T("4"), T("40")],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "PL/pgSQL FOUND after unchanged UPDATE",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY, a INT)",
                "INSERT INTO t VALUES (1, 10)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: r#"CREATE FUNCTION update_found(target_id INT) RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN UPDATE t SET a = a WHERE id = target_id; RETURN FOUND; END; $$;"#,
                    expected: Expected::Tag("CREATE FUNCTION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT update_found(1)",
                    expected: Expected::Rows {
                        columns: &[Column("update_found", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT update_found(999)",
                    expected: Expected::Rows {
                        columns: &[Column("update_found", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "NULL and transaction counts without a primary key",
            set_up_script: &[
                "CREATE TABLE t (a INT)",
                "INSERT INTO t VALUES (NULL), (1)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "BEGIN",
                    expected: Expected::Tag("BEGIN"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a = NULL WHERE a IS NULL",
                    expected: Expected::Tag("UPDATE 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a = a",
                    expected: Expected::Tag("UPDATE 2"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "COMMIT",
                    expected: Expected::Tag("COMMIT"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "BEFORE UPDATE trigger skips a row",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY, a INT)",
                "INSERT INTO t VALUES (1, 10)",
                r#"CREATE FUNCTION skip_update() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RETURN NULL; END; $$;"#,
                "CREATE TRIGGER skip_update BEFORE UPDATE ON t FOR EACH ROW EXECUTE FUNCTION skip_update()",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t SET a = a WHERE id = 1",
                    expected: Expected::Tag("UPDATE 0"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a = a WHERE id = 1 RETURNING id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[],
                        tag: "UPDATE 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a = a WHERE id = 1 RETURNING id",
                    expected: Expected::Tag("UPDATE 0"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "BEFORE UPDATE trigger skips one of several rows",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY, a INT)",
                "INSERT INTO t VALUES (1, 10), (2, 20), (3, 30)",
                r#"CREATE FUNCTION skip_one() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN IF OLD.id = 1 THEN RETURN NULL; END IF; RETURN NEW; END; $$;"#,
                "CREATE TRIGGER skip_one BEFORE UPDATE ON t FOR EACH ROW EXECUTE FUNCTION skip_one()",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t SET a = CASE WHEN id = 3 THEN 31 ELSE a END",
                    expected: Expected::Tag("UPDATE 2"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, a FROM t ORDER BY id",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("20")],
                            &[T("3"), T("31")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "UPDATE FROM counts and RETURNING",
            set_up_script: &[
                "CREATE TABLE t (id INT PRIMARY KEY, a INT)",
                "CREATE TABLE u (id INT)",
                "INSERT INTO t VALUES (1, 10), (2, 20)",
                "INSERT INTO u VALUES (1), (1), (2)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t SET a = t.a FROM u WHERE t.id = u.id",
                    expected: Expected::Tag("UPDATE 2"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a = 10 FROM u WHERE t.id = u.id",
                    expected: Expected::Tag("UPDATE 2"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a = t.a FROM u WHERE t.id = u.id RETURNING t.id, t.a",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT4)],
                        rows: &[
                            &[T("1"), T("10")],
                            &[T("2"), T("10")],
                        ],
                        tag: "UPDATE 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE t SET a = t.a FROM u WHERE t.id = u.id RETURNING t.id, t.a",
                    expected: Expected::Tag("UPDATE 2"),
                    flow: Flow::Exec,
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_update_assignment_semantics() {
    run_scripts(&[
        ScriptTest {
            name: "customer CASE",
            set_up_script: &[
                "CREATE TABLE t_seq (a int, b int)",
                "INSERT INTO t_seq VALUES (1, 0)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET a = 2, b = CASE WHEN a = 1 THEN 100 ELSE -1 END",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM t_seq",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("2"), T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "reversed CASE",
            set_up_script: &[
                "CREATE TABLE t_seq (a int, b int)",
                "INSERT INTO t_seq VALUES (1, 0)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET b = CASE WHEN a = 1 THEN 100 ELSE -1 END, a = 2",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM t_seq",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("2"), T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "swap",
            set_up_script: &[
                "CREATE TABLE t_seq (a int, b int)",
                "INSERT INTO t_seq VALUES (1, 0)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET a = b, b = a",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM t_seq",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("0"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "reversed swap",
            set_up_script: &[
                "CREATE TABLE t_seq (a int, b int)",
                "INSERT INTO t_seq VALUES (1, 0)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET b = a, a = b",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM t_seq",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("0"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "arithmetic chain",
            set_up_script: &[
                "CREATE TABLE t_seq (a int, b int)",
                "INSERT INTO t_seq VALUES (1, 0)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET a = a + 1, b = a + 10",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM t_seq",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("2"), T("11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "NULL propagation",
            set_up_script: &[
                "CREATE TABLE t_seq (a int, b int)",
                "INSERT INTO t_seq VALUES (1, 0)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET a = NULL, b = CASE WHEN a IS NULL THEN 100 ELSE -1 END",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM t_seq",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[Null, T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "multiple rows",
            set_up_script: &[
                "CREATE TABLE t_seq (a int, b int)",
                "INSERT INTO t_seq VALUES (1, 0)",
                "INSERT INTO t_seq VALUES (3, 9)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET a = a + 1, b = a",
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM t_seq ORDER BY a",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("2"), T("1")],
                            &[T("4"), T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "scalar correlated subquery",
            set_up_script: &[
                "CREATE TABLE t_seq (a int, b int)",
                "INSERT INTO t_seq VALUES (1, 0)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET a = 2, b = (SELECT a + 10)",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM t_seq",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("2"), T("11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "WHERE subquery",
            set_up_script: &[
                "CREATE TABLE t_seq (a int, b int)",
                "INSERT INTO t_seq VALUES (1, 0)",
                "CREATE TABLE src (x int PRIMARY KEY)",
                "INSERT INTO src VALUES (1)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET a = 2, b = a WHERE a IN (SELECT x FROM src)",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM t_seq",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("2"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "assignment conversion",
            set_up_script: &[
                "CREATE TABLE t_seq (a int, b int)",
                "INSERT INTO t_seq VALUES (1, 0)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET a = 1.6, b = a",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM t_seq",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("2"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "generated stored column",
            set_up_script: &[
                "CREATE TABLE t_seq (a int, b int, c int GENERATED ALWAYS AS (a+b) STORED)",
                "INSERT INTO t_seq (a,b) VALUES (1,0)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET a = 2, b = a",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a,b,c FROM t_seq",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4)],
                        rows: &[
                            &[T("2"), T("1"), T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "join same target",
            set_up_script: &[
                "CREATE TABLE t_seq (id int PRIMARY KEY, a int, b int)",
                "INSERT INTO t_seq VALUES (1,1,0)",
                "CREATE TABLE src (id int PRIMARY KEY, x int)",
                "INSERT INTO src VALUES (1,10)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET a = 2, b = a FROM src WHERE t_seq.id = src.id",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM t_seq",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("2"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "join swap",
            set_up_script: &[
                "CREATE TABLE t_seq (id int PRIMARY KEY, a int, b int)",
                "INSERT INTO t_seq VALUES (1,1,0)",
                "CREATE TABLE src (id int PRIMARY KEY, x int)",
                "INSERT INTO src VALUES (1,10)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET a = b, b = a FROM src WHERE t_seq.id = src.id",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM t_seq",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("0"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "repeated target is rejected",
            set_up_script: &[
                "CREATE TABLE t_seq (a int, b int)",
                "INSERT INTO t_seq VALUES (1, 0)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET a = a + 1, a = a + 10, b = a",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"multiple assignments to same column "a""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM t_seq",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("1"), T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "assignments through foreign key and check handlers",
            set_up_script: &[
                "CREATE TABLE parent (id int PRIMARY KEY)",
                "INSERT INTO parent VALUES (1), (2)",
                "CREATE TABLE t_seq (id int PRIMARY KEY, a int REFERENCES parent(id), b int, CHECK (b < a))",
                "INSERT INTO t_seq VALUES (9, 1, 0)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET a = 2, b = a RETURNING id, a, b",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("9"), T("2"), T("1")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, a, b FROM t_seq",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("9"), T("2"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "RETURNING reads completed new row",
            set_up_script: &[
                "CREATE TABLE t_seq (a int, b int)",
                "INSERT INTO t_seq VALUES (1, 0)",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "UPDATE t_seq SET a = b, b = a RETURNING a, b",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("0"), T("1")],
                        ],
                        tag: "UPDATE 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT a, b FROM t_seq",
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4)],
                        rows: &[
                            &[T("0"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
