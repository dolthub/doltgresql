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

func TestUpdate(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "simple update",
			SetUpScript: []string{
				"CREATE TABLE t1 (a INT PRIMARY KEY, b INT)",
				"INSERT INTO t1 VALUES (1, 2), (2, 3), (3, 4)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "UPDATE t1 SET b = 5 WHERE a = 2",
				},
				{
					Query: "SELECT * FROM t1 where a =  2",
					Expected: []sql.Row{
						{2, 5},
					},
				},
			},
		},
		{
			Name: "update to default",
			SetUpScript: []string{
				"create table t (i int default 10, j varchar(128) default (concat('abc', 'def')));",
				"insert into t values (100, 'a'), (200, 'b');",
				"create table t2 (i int);",
				"insert into t2 values (1), (2), (3);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "update t set i = default where i = 100;",
				},
				{
					Query: "select * from t order by i",
					Expected: []sql.Row{
						{10, "a"},
						{200, "b"},
					},
				},
				{
					Query: "update t set j = default where i = 200;",
				},
				{
					Query: "select * from t order by i",
					Expected: []sql.Row{
						{10, "a"},
						{200, "abcdef"},
					},
				},
				{
					Query: "update t set i = default, j = default;",
				},
				{
					Query: "select * from t order by i",
					Expected: []sql.Row{
						{10, "abcdef"},
						{10, "abcdef"},
					},
				},
				{
					Query: "update t2 set i = default",
					Skip:  true, // UPDATE: non-Doltgres type found in source
				},
				{
					Query: "select * from t2",
					Skip:  true, // skipped because of above
					Expected: []sql.Row{
						{nil},
						{nil},
						{nil},
					},
				},
			},
		},
		{
			Name: "UPDATE ... RETURNING",
			SetUpScript: []string{
				"CREATE TABLE t (pk INT PRIMARY KEY, c1 TEXT);",
				"INSERT INTO t VALUES (1, 'one');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "UPDATE t SET pk = pk+1, c1 = '42' RETURNING c1, pk, pk * 2;",
					Expected: []sql.Row{{"42", 2, 4}},
				},
				{
					Query:    "UPDATE t SET c1 = '43' RETURNING *;",
					Expected: []sql.Row{{2, "43"}},
				},
			},
		},
		{
			Name: "UPDATE ... RETURNING with join",
			SetUpScript: []string{
				"CREATE TABLE employees (id SERIAL PRIMARY KEY, name TEXT, department_id INT, salary INT);",
				"CREATE TABLE departments (id SERIAL PRIMARY KEY, name TEXT, bonus INT);",
				"INSERT INTO employees (name, department_id, salary) VALUES ('Alice', 1, 50000), ('Bob', 2, 60000);",
				"INSERT INTO departments (name, bonus) VALUES ('Engineering', 5000), ('Marketing', 3000);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "UPDATE employees e SET salary = salary + d.bonus FROM departments d WHERE e.department_id = d.id RETURNING e.id, e.name, e.salary;",
					Expected: []sql.Row{
						{1, "Alice", 55000},
						{2, "Bob", 63000},
					},
				},
			},
		},
		{
			Name: "UPDATE with join on subquery",
			SetUpScript: []string{
				"CREATE TABLE employees (id INT PRIMARY KEY, name TEXT, department_id INT, salary INT);",
				"CREATE TABLE departments (id INT PRIMARY KEY, name TEXT, bonus INT);",
				"INSERT INTO employees VALUES (1, 'Alice', 10, 50000), (2, 'Bob', 20, 60000);",
				"INSERT INTO departments VALUES (10, 'Engineering', 5000), (20, 'HR', 3000);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `
			UPDATE employees SET salary = salary + dept_bonus.bonus
			FROM ( SELECT id, bonus FROM departments ) AS dept_bonus
			WHERE employees.department_id = dept_bonus.id AND employees.name = 'Alice';`,
				},
				{
					Query: "SELECT salary FROM employees WHERE name = 'Alice';",
					Expected: []sql.Row{
						{55000},
					},
				},
			},
		},
		{
			Name: "UPDATE with join on one table",
			SetUpScript: []string{
				"CREATE TABLE products (id SERIAL PRIMARY KEY, name TEXT, price INT, category_id INT);",
				"CREATE TABLE categories (id SERIAL PRIMARY KEY, name TEXT, discount INT);",
				"INSERT INTO products (name, price, category_id) VALUES ('Laptop', 1000, 1), ('Phone', 800, 2);",
				"INSERT INTO categories (name, discount) VALUES ('Electronics', 100), ('Mobiles', 50);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "UPDATE products p SET price = price - c.discount FROM categories c WHERE p.category_id = c.id;",
				},
				{
					Query: "SELECT id, name, price FROM products ORDER BY id;",
					Expected: []sql.Row{
						{1, "Laptop", 900},
						{2, "Phone", 750},
					},
				},
			},
		},
		{
			Name: "UPDATE with join on two tables",
			SetUpScript: []string{
				"CREATE TABLE books (id SERIAL PRIMARY KEY, title TEXT, price INT, author_id INT, publisher_id INT);",
				"CREATE TABLE authors (id SERIAL PRIMARY KEY, royalty INT);",
				"CREATE TABLE publishers (id SERIAL PRIMARY KEY, markup INT);",
				"INSERT INTO books (title, price, author_id, publisher_id) VALUES ('Book A', 100, 1, 1), ('Book B', 120, 2, 2);",
				"INSERT INTO authors (royalty) VALUES (10), (20);",
				"INSERT INTO publishers (markup) VALUES (15), (25);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "UPDATE books b SET price = price + a.royalty - p.markup FROM authors a, publishers p WHERE b.author_id = a.id AND b.publisher_id = p.id;",
				},
				{
					Query: "SELECT id, title, price FROM books ORDER BY id;",
					Expected: []sql.Row{
						{1, "Book A", 95},  // 100 + 10 - 15
						{2, "Book B", 115}, // 120 + 20 - 25
					},
				},
			},
		},
		{
			Name: "UPDATE with join on three tables",
			SetUpScript: []string{
				"CREATE TABLE orders (id SERIAL PRIMARY KEY, customer_id INT, product_id INT, total INT);",
				"CREATE TABLE customers (id SERIAL PRIMARY KEY, loyalty_discount INT);",
				"CREATE TABLE products (id SERIAL PRIMARY KEY, base_price INT, tax_id INT);",
				"CREATE TABLE taxes (id SERIAL PRIMARY KEY, rate INT);",
				"INSERT INTO orders (customer_id, product_id, total) VALUES (1, 1, 0), (2, 2, 0);",
				"INSERT INTO customers (loyalty_discount) VALUES (5), (10);",
				"INSERT INTO products (base_price, tax_id) VALUES (100, 1), (200, 2);",
				"INSERT INTO taxes (rate) VALUES (10), (20);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `
				UPDATE orders o
				SET total = p.base_price + (p.base_price * t.rate / 100) - c.loyalty_discount
				FROM customers c, products p, taxes t
				WHERE o.customer_id = c.id AND o.product_id = p.id AND p.tax_id = t.id;
			`,
				},
				{
					Query: "SELECT id, total FROM orders ORDER BY id;",
					Expected: []sql.Row{
						{1, 105}, // 100 + 10 - 5
						{2, 230}, // 200 + 40 - 10
					},
				},
			},
		},
		{
			Name: "UPDATE with join on four tables",
			SetUpScript: []string{
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
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `
				UPDATE rentals r
				SET total_cost = v.base_rate + f.surcharge - m.discount
				FROM vehicles v, fuel_types f, users u, membership_levels m
				WHERE r.vehicle_id = v.id AND v.fuel_type_id = f.id AND r.user_id = u.id AND u.membership_level_id = m.id;
			`,
				},
				{
					Query: "SELECT id, total_cost FROM rentals ORDER BY id;",
					Expected: []sql.Row{
						{1, 270}, // 300 + 20 - 50
						{2, 360}, // 400 + 40 - 80
					},
				},
			},
		},
		{
			Name: "UPDATE with join on table with trigger",
			SetUpScript: []string{
				"CREATE TABLE departments (id SERIAL PRIMARY KEY, name TEXT, bonus INT\n);",
				"CREATE TABLE employees (id SERIAL PRIMARY KEY, name TEXT, department_id INT REFERENCES departments(id), salary INT);",
				"INSERT INTO departments (name, bonus) VALUES ('Engineering', 1000), ('HR', 500);",
				"INSERT INTO employees (name, department_id, salary) VALUES ('Alice', 1, 50000), ('Bob', 2, 45000);",
				"CREATE TABLE salary_log (employee_id INT, old_salary INT, new_salary INT);",
				`CREATE OR REPLACE FUNCTION log_salary_change()
					RETURNS TRIGGER AS $$
					BEGIN
						IF NEW.salary != OLD.salary THEN
							INSERT INTO salary_log VALUES (OLD.id, OLD.salary, NEW.salary);
						END IF;
						RETURN NEW;
					END;
					$$ LANGUAGE plpgsql;`,
				`CREATE TRIGGER trg_log_salary_change
					AFTER UPDATE ON employees
					FOR EACH ROW
					EXECUTE FUNCTION log_salary_change();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "UPDATE employees e SET salary = salary + d.bonus FROM departments d WHERE e.department_id = d.id;",
				},
				{
					// TODO: Triggers do not currently work for UPDATE ... FROM, because updatableJoinTable
					//       doesn't implement sql.DatabaseSchemaTable
					Skip:  true,
					Query: "SELECT * FROM salary_log;",
					Expected: []sql.Row{
						{1, 50000, 51000},
						{2, 45000, 45500},
					},
				},
			},
		},
	})
}

// TestUpdateAssignmentSemantics covers pre-update-row evaluation, verified against PostgreSQL 18.6.
// https://github.com/dolthub/doltgresql/issues/3092
func TestUpdateAssignmentSemantics(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "customer CASE",
			SetUpScript: []string{
				"CREATE TABLE t_seq (a int, b int)",
				"INSERT INTO t_seq VALUES (1, 0)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_seq SET a = 2, b = CASE WHEN a = 1 THEN 100 ELSE -1 END"},
				{
					Query:    "SELECT a, b FROM t_seq",
					Expected: []sql.Row{{2, 100}},
				},
			},
		},
		{
			Name: "reversed CASE",
			SetUpScript: []string{
				"CREATE TABLE t_seq (a int, b int)",
				"INSERT INTO t_seq VALUES (1, 0)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_seq SET b = CASE WHEN a = 1 THEN 100 ELSE -1 END, a = 2"},
				{
					Query:    "SELECT a, b FROM t_seq",
					Expected: []sql.Row{{2, 100}},
				},
			},
		},
		{
			Name: "swap",
			SetUpScript: []string{
				"CREATE TABLE t_seq (a int, b int)",
				"INSERT INTO t_seq VALUES (1, 0)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_seq SET a = b, b = a"},
				{
					Query:    "SELECT a, b FROM t_seq",
					Expected: []sql.Row{{0, 1}},
				},
			},
		},
		{
			Name: "reversed swap",
			SetUpScript: []string{
				"CREATE TABLE t_seq (a int, b int)",
				"INSERT INTO t_seq VALUES (1, 0)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_seq SET b = a, a = b"},
				{
					Query:    "SELECT a, b FROM t_seq",
					Expected: []sql.Row{{0, 1}},
				},
			},
		},
		{
			Name: "arithmetic chain",
			SetUpScript: []string{
				"CREATE TABLE t_seq (a int, b int)",
				"INSERT INTO t_seq VALUES (1, 0)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_seq SET a = a + 1, b = a + 10"},
				{
					Query:    "SELECT a, b FROM t_seq",
					Expected: []sql.Row{{2, 11}},
				},
			},
		},
		{
			Name: "NULL propagation",
			SetUpScript: []string{
				"CREATE TABLE t_seq (a int, b int)",
				"INSERT INTO t_seq VALUES (1, 0)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_seq SET a = NULL, b = CASE WHEN a IS NULL THEN 100 ELSE -1 END"},
				{
					Query:    "SELECT a, b FROM t_seq",
					Expected: []sql.Row{{nil, -1}},
				},
			},
		},
		{
			Name: "multiple rows",
			SetUpScript: []string{
				"CREATE TABLE t_seq (a int, b int)",
				"INSERT INTO t_seq VALUES (1, 0)",
				"INSERT INTO t_seq VALUES (3, 9)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_seq SET a = a + 1, b = a"},
				{
					Query:    "SELECT a, b FROM t_seq ORDER BY a",
					Expected: []sql.Row{{2, 1}, {4, 3}},
				},
			},
		},
		{
			Name: "scalar correlated subquery",
			SetUpScript: []string{
				"CREATE TABLE t_seq (a int, b int)",
				"INSERT INTO t_seq VALUES (1, 0)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_seq SET a = 2, b = (SELECT a + 10)"},
				{
					Query:    "SELECT a, b FROM t_seq",
					Expected: []sql.Row{{2, 11}},
				},
			},
		},
		{
			Name: "WHERE subquery",
			SetUpScript: []string{
				"CREATE TABLE t_seq (a int, b int)",
				"INSERT INTO t_seq VALUES (1, 0)",
				"CREATE TABLE src (x int PRIMARY KEY)",
				"INSERT INTO src VALUES (1)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_seq SET a = 2, b = a WHERE a IN (SELECT x FROM src)"},
				{
					Query:    "SELECT a, b FROM t_seq",
					Expected: []sql.Row{{2, 1}},
				},
			},
		},
		{
			Name: "assignment conversion",
			SetUpScript: []string{
				"CREATE TABLE t_seq (a int, b int)",
				"INSERT INTO t_seq VALUES (1, 0)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_seq SET a = 1.6, b = a"},
				{
					Query:    "SELECT a, b FROM t_seq",
					Expected: []sql.Row{{2, 1}},
				},
			},
		},
		{
			Name: "generated stored column",
			SetUpScript: []string{
				"CREATE TABLE t_seq (a int, b int, c int GENERATED ALWAYS AS (a+b) STORED)",
				"INSERT INTO t_seq (a,b) VALUES (1,0)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_seq SET a = 2, b = a"},
				{
					Query:    "SELECT a,b,c FROM t_seq",
					Expected: []sql.Row{{2, 1, 3}},
				},
			},
		},
		{
			Name: "join same target",
			SetUpScript: []string{
				"CREATE TABLE t_seq (id int PRIMARY KEY, a int, b int)",
				"INSERT INTO t_seq VALUES (1,1,0)",
				"CREATE TABLE src (id int PRIMARY KEY, x int)",
				"INSERT INTO src VALUES (1,10)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_seq SET a = 2, b = a FROM src WHERE t_seq.id = src.id"},
				{
					Query:    "SELECT a, b FROM t_seq",
					Expected: []sql.Row{{2, 1}},
				},
			},
		},
		{
			Name: "join swap",
			SetUpScript: []string{
				"CREATE TABLE t_seq (id int PRIMARY KEY, a int, b int)",
				"INSERT INTO t_seq VALUES (1,1,0)",
				"CREATE TABLE src (id int PRIMARY KEY, x int)",
				"INSERT INTO src VALUES (1,10)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_seq SET a = b, b = a FROM src WHERE t_seq.id = src.id"},
				{
					Query:    "SELECT a, b FROM t_seq",
					Expected: []sql.Row{{0, 1}},
				},
			},
		},
		{
			Name: "repeated target is rejected",
			Skip: true, // TODO: Reject duplicate targets (https://github.com/dolthub/doltgresql/issues/3092).
			SetUpScript: []string{
				"CREATE TABLE t_seq (a int, b int)",
				"INSERT INTO t_seq VALUES (1, 0)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "UPDATE t_seq SET a = a + 1, a = a + 10, b = a",
					ExpectedErr: `multiple assignments to same column "a"`,
				},
				{Query: "SELECT a, b FROM t_seq", Expected: []sql.Row{{1, 0}}},
			},
		},
		{
			Name: "assignments through foreign key and check handlers",
			SetUpScript: []string{
				"CREATE TABLE parent (id int PRIMARY KEY)",
				"INSERT INTO parent VALUES (1), (2)",
				"CREATE TABLE t_seq (id int PRIMARY KEY, a int REFERENCES parent(id), b int, CHECK (b < a))",
				"INSERT INTO t_seq VALUES (9, 1, 0)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_seq SET a = 2, b = a RETURNING id, a, b", Expected: []sql.Row{{9, 2, 1}}},
				{Query: "SELECT id, a, b FROM t_seq", Expected: []sql.Row{{9, 2, 1}}},
			},
		},
		{
			Name: "RETURNING reads completed new row",
			SetUpScript: []string{
				"CREATE TABLE t_seq (a int, b int)",
				"INSERT INTO t_seq VALUES (1, 0)",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "UPDATE t_seq SET a = b, b = a RETURNING a, b", Expected: []sql.Row{{0, 1}}},
				{Query: "SELECT a, b FROM t_seq", Expected: []sql.Row{{0, 1}}},
			},
		},
	})
}
