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

#[test]
fn test_updates_keeping_primary_keys() {
    run_scripts(&[
        ScriptTest {
            name: "updates that keep primary keys",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE up (id INT PRIMARY KEY, u INT UNIQUE, v INT, w TEXT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX up_v ON up (v);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX up_w ON up (lower(w));",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO up SELECT i, i, i % 3, 'W' || i FROM generate_series(1, 20) i;",
                    expected: Expected::Tag("INSERT 0 20"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE up SET w = w || 'x' WHERE id <= 5;",
                    expected: Expected::Tag("UPDATE 5"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE up SET v = v + 10 WHERE id BETWEEN 3 AND 8;",
                    expected: Expected::Tag("UPDATE 6"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE up SET u = u + 100, v = v WHERE id > 15;",
                    expected: Expected::Tag("UPDATE 5"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE up SET u = u + 1000 WHERE id <= 20;",
                    expected: Expected::Tag("UPDATE 20"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, u, v, w FROM up ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("u", INT4), Column("v", INT4), Column("w", TEXT)],
                        rows: &[
                            &[T("1"), T("1001"), T("1"), T("W1x")],
                            &[T("2"), T("1002"), T("2"), T("W2x")],
                            &[T("3"), T("1003"), T("10"), T("W3x")],
                            &[T("4"), T("1004"), T("11"), T("W4x")],
                            &[T("5"), T("1005"), T("12"), T("W5x")],
                            &[T("6"), T("1006"), T("10"), T("W6")],
                            &[T("7"), T("1007"), T("11"), T("W7")],
                            &[T("8"), T("1008"), T("12"), T("W8")],
                            &[T("9"), T("1009"), T("0"), T("W9")],
                            &[T("10"), T("1010"), T("1"), T("W10")],
                            &[T("11"), T("1011"), T("2"), T("W11")],
                            &[T("12"), T("1012"), T("0"), T("W12")],
                            &[T("13"), T("1013"), T("1"), T("W13")],
                            &[T("14"), T("1014"), T("2"), T("W14")],
                            &[T("15"), T("1015"), T("0"), T("W15")],
                            &[T("16"), T("1116"), T("1"), T("W16")],
                            &[T("17"), T("1117"), T("2"), T("W17")],
                            &[T("18"), T("1118"), T("0"), T("W18")],
                            &[T("19"), T("1119"), T("1"), T("W19")],
                            &[T("20"), T("1120"), T("2"), T("W20")],
                        ],
                        tag: "SELECT 20",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM up WHERE v = 11 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("4")],
                            &[T("7")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM up WHERE lower(w) = 'w2x' ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM up WHERE u = 1003;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE up SET u = 1003 WHERE id = 1;",
                    expected: Expected::Error(Diagnostic { code: "23505", message: r#"duplicate key value violates unique constraint "up_u_key""#, detail: "Key (u)=(1003) already exists.", schema: "public", table: "up", constraint: "up_u_key", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE up SET id = id + 100 WHERE id = 2;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, u FROM up WHERE id > 100 OR u = 1001 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("u", INT4)],
                        rows: &[
                            &[T("1"), T("1001")],
                            &[T("102"), T("1002")],
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
fn test_update_from_joins() {
    run_scripts(&[
        ScriptTest {
            name: "UPDATE FROM and DELETE USING joins",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE TABLE uf (id INT PRIMARY KEY, k INT, v INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX uf_k ON uf (k);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE ufs (id INT PRIMARY KEY, k INT, d INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE ufl (k INT, v INT);",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO uf SELECT i, i % 10, 0 FROM generate_series(1, 200) i;",
                    expected: Expected::Tag("INSERT 0 200"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO ufs SELECT i, i % 20, i FROM generate_series(1, 100) i;",
                    expected: Expected::Tag("INSERT 0 100"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO ufl SELECT i % 5, i FROM generate_series(1, 30) i;",
                    expected: Expected::Tag("INSERT 0 30"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE uf SET v = s.k FROM ufs s WHERE s.k = uf.k AND s.id <= 10 AND uf.id > 150;",
                    expected: Expected::Tag("UPDATE 45"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT k, count(*), sum(v) FROM uf GROUP BY k ORDER BY k;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4), Column("count", INT8), Column("sum", INT8)],
                        rows: &[
                            &[T("0"), T("20"), T("0")],
                            &[T("1"), T("20"), T("5")],
                            &[T("2"), T("20"), T("10")],
                            &[T("3"), T("20"), T("15")],
                            &[T("4"), T("20"), T("20")],
                            &[T("5"), T("20"), T("25")],
                            &[T("6"), T("20"), T("30")],
                            &[T("7"), T("20"), T("35")],
                            &[T("8"), T("20"), T("40")],
                            &[T("9"), T("20"), T("45")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE uf SET v = t.n FROM (SELECT k, count(*) AS n FROM ufs GROUP BY k) t WHERE t.k = uf.k + 10;",
                    expected: Expected::Tag("UPDATE 200"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT k, count(*), sum(v) FROM uf GROUP BY k ORDER BY k;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4), Column("count", INT8), Column("sum", INT8)],
                        rows: &[
                            &[T("0"), T("20"), T("100")],
                            &[T("1"), T("20"), T("100")],
                            &[T("2"), T("20"), T("100")],
                            &[T("3"), T("20"), T("100")],
                            &[T("4"), T("20"), T("100")],
                            &[T("5"), T("20"), T("100")],
                            &[T("6"), T("20"), T("100")],
                            &[T("7"), T("20"), T("100")],
                            &[T("8"), T("20"), T("100")],
                            &[T("9"), T("20"), T("100")],
                        ],
                        tag: "SELECT 10",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE uf SET v = uf.v + 1 FROM ufs s, ufl l WHERE s.id = uf.id AND l.v = s.d AND l.k = 2;",
                    expected: Expected::Tag("UPDATE 6"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, v FROM uf WHERE id <= 30 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", INT4)],
                        rows: &[
                            &[T("1"), T("5")],
                            &[T("2"), T("6")],
                            &[T("3"), T("5")],
                            &[T("4"), T("5")],
                            &[T("5"), T("5")],
                            &[T("6"), T("5")],
                            &[T("7"), T("6")],
                            &[T("8"), T("5")],
                            &[T("9"), T("5")],
                            &[T("10"), T("5")],
                            &[T("11"), T("5")],
                            &[T("12"), T("6")],
                            &[T("13"), T("5")],
                            &[T("14"), T("5")],
                            &[T("15"), T("5")],
                            &[T("16"), T("5")],
                            &[T("17"), T("6")],
                            &[T("18"), T("5")],
                            &[T("19"), T("5")],
                            &[T("20"), T("5")],
                            &[T("21"), T("5")],
                            &[T("22"), T("6")],
                            &[T("23"), T("5")],
                            &[T("24"), T("5")],
                            &[T("25"), T("5")],
                            &[T("26"), T("5")],
                            &[T("27"), T("6")],
                            &[T("28"), T("5")],
                            &[T("29"), T("5")],
                            &[T("30"), T("5")],
                        ],
                        tag: "SELECT 30",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE uf AS a SET v = b.id FROM uf AS b WHERE b.id = a.id + 100 AND a.k = 3;",
                    expected: Expected::Tag("UPDATE 10"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, v FROM uf WHERE k = 3 ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", INT4)],
                        rows: &[
                            &[T("3"), T("103")],
                            &[T("13"), T("113")],
                            &[T("23"), T("123")],
                            &[T("33"), T("133")],
                            &[T("43"), T("143")],
                            &[T("53"), T("153")],
                            &[T("63"), T("163")],
                            &[T("73"), T("173")],
                            &[T("83"), T("183")],
                            &[T("93"), T("193")],
                            &[T("103"), T("5")],
                            &[T("113"), T("5")],
                            &[T("123"), T("5")],
                            &[T("133"), T("5")],
                            &[T("143"), T("5")],
                            &[T("153"), T("5")],
                            &[T("163"), T("5")],
                            &[T("173"), T("5")],
                            &[T("183"), T("5")],
                            &[T("193"), T("5")],
                        ],
                        tag: "SELECT 20",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM uf USING ufs s WHERE s.d = uf.id AND s.k > 15;",
                    expected: Expected::Tag("DELETE 20"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*), sum(id) FROM uf;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("sum", INT8)],
                        rows: &[
                            &[T("180"), T("18950")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM ufl USING uf WHERE uf.k = ufl.k AND uf.id = 7;",
                    expected: Expected::Tag("DELETE 0"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT k, count(*) FROM ufl GROUP BY k ORDER BY k;",
                    expected: Expected::Rows {
                        columns: &[Column("k", INT4), Column("count", INT8)],
                        rows: &[
                            &[T("0"), T("6")],
                            &[T("1"), T("6")],
                            &[T("2"), T("6")],
                            &[T("3"), T("6")],
                            &[T("4"), T("6")],
                        ],
                        tag: "SELECT 5",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE ufl SET v = -ufl.v FROM ufs s WHERE s.id = ufl.v AND s.k = 4;",
                    expected: Expected::Tag("UPDATE 2"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v FROM ufl ORDER BY v;",
                    expected: Expected::Rows {
                        columns: &[Column("v", INT4)],
                        rows: &[
                            &[T("-24")],
                            &[T("-4")],
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                            &[T("5")],
                            &[T("6")],
                            &[T("7")],
                            &[T("8")],
                            &[T("9")],
                            &[T("10")],
                            &[T("11")],
                            &[T("12")],
                            &[T("13")],
                            &[T("14")],
                            &[T("15")],
                            &[T("16")],
                            &[T("17")],
                            &[T("18")],
                            &[T("19")],
                            &[T("20")],
                            &[T("21")],
                            &[T("22")],
                            &[T("23")],
                            &[T("25")],
                            &[T("26")],
                            &[T("27")],
                            &[T("28")],
                            &[T("29")],
                            &[T("30")],
                        ],
                        tag: "SELECT 30",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
