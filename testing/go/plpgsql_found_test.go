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
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

// TestPlpgsqlFound covers the built-in FOUND variable. The set of statements that update it, and the ones
// that deliberately leave it alone, were verified against PostgreSQL 16.
func TestPlpgsqlFound(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "FOUND starts out false",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_init() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_init();`,
					Expected: []sql.Row{{"f"}},
				},
			},
		},
		{
			Name: "SELECT INTO sets FOUND",
			SetUpScript: []string{
				`CREATE TABLE k (id int, nm text);`,
				`INSERT INTO k VALUES (1, 'a'), (2, 'b');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_hit() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN SELECT id INTO v FROM k WHERE id = 1; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_hit();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query: `CREATE FUNCTION f_miss() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN SELECT id INTO v FROM k WHERE id = 99; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_miss();`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// The lowercase spelling resolves to the same variable.
					Query: `CREATE FUNCTION f_lower() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN SELECT id INTO v FROM k WHERE id = 99; RETURN found; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_lower();`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// A RECORD target sets FOUND the same way a scalar one does.
					Query: `CREATE FUNCTION f_rec_miss() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN SELECT id INTO r FROM k WHERE id = 99; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_rec_miss();`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query: `CREATE FUNCTION f_rec_hit() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN SELECT id INTO r FROM k WHERE id = 1; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_rec_hit();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// The real-world shape from the dump: guard the rest of the body on FOUND.
					Query: `CREATE FUNCTION f_guard(p int) RETURNS text LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN
	SELECT id, nm INTO r FROM k WHERE id = p;
	IF NOT FOUND THEN RETURN 'absent'; END IF;
	RETURN r.nm;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_guard(2);`,
					Expected: []sql.Row{{"b"}},
				},
				{
					Query:    `SELECT f_guard(99);`,
					Expected: []sql.Row{{"absent"}},
				},
			},
		},
		{
			Name: "statements that leave FOUND alone",
			SetUpScript: []string{
				`CREATE TABLE k (id int);`,
				`INSERT INTO k VALUES (1);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// An assignment is not one of the statements that sets FOUND.
					Query: `CREATE FUNCTION f_assign() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN SELECT id INTO v FROM k WHERE id = 1; v := 42; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_assign();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// Dynamic EXECUTE deliberately does not set FOUND, even with an INTO clause.
					Query: `CREATE FUNCTION f_exec_hit() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN EXECUTE 'SELECT id FROM k WHERE id = 1' INTO v; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_exec_hit();`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// A utility statement is not data-modifying, so it leaves FOUND as the INTO set it.
					Query: `CREATE FUNCTION f_ddl() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN
	SELECT id INTO v FROM k WHERE id = 1;
	CREATE TABLE f_ddl_scratch (a int);
	RETURN FOUND;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_ddl();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// The same the other way round: a utility statement cannot turn FOUND on either.
					Query: `CREATE FUNCTION f_ddl_miss() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN
	SELECT id INTO v FROM k WHERE id = 99;
	CREATE TABLE f_ddl_scratch2 (a int);
	RETURN FOUND;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_ddl_miss();`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// A data-modifying statement wrapped in a CTE still counts as data-modifying.
					Query: `CREATE FUNCTION f_cte_dml() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN
	WITH src AS (SELECT 99 AS id) INSERT INTO k SELECT id FROM src;
	RETURN FOUND;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_cte_dml();`,
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "PERFORM and data-modifying statements set FOUND",
			SetUpScript: []string{
				`CREATE TABLE k (id int);`,
				`INSERT INTO k VALUES (1), (2), (3);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_perf_hit() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN PERFORM 1 FROM k WHERE id = 1; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_perf_hit();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query: `CREATE FUNCTION f_perf_miss() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN PERFORM 1 FROM k WHERE id = 99; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_perf_miss();`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// DML reports whether it affected any row, not whether it returned one.
					Query: `CREATE FUNCTION f_del_miss() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN DELETE FROM k WHERE id = 99; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_del_miss();`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query: `CREATE FUNCTION f_del_hit() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN DELETE FROM k WHERE id = 3; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_del_hit();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query: `CREATE FUNCTION f_upd_miss() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN UPDATE k SET id = id WHERE id = 99; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_upd_miss();`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query: `CREATE FUNCTION f_ins_plain() RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN INSERT INTO k VALUES (10); RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_ins_plain();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query: `CREATE FUNCTION f_ins_returning() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN INSERT INTO k VALUES (11) RETURNING id INTO v; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_ins_returning();`,
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "FOR..IN..SELECT sets FOUND when the loop exits",
			SetUpScript: []string{
				`CREATE TABLE k (id int);`,
				`INSERT INTO k VALUES (1), (2);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_for_hit() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN FOR r IN SELECT id FROM k LOOP END LOOP; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_for_hit();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query: `CREATE FUNCTION f_for_miss() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE r RECORD;
BEGIN FOR r IN SELECT id FROM k WHERE id = 99 LOOP END LOOP; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_for_miss();`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// The loop does not touch FOUND while it is running, only when it exits.
					Query: `CREATE FUNCTION f_in_loop() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE r RECORD; res boolean;
BEGIN FOR r IN SELECT id FROM k LOOP res := FOUND; END LOOP; RETURN res; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_in_loop();`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// Leaving through EXIT is still leaving, so the loop reports that it ran.
					Query: `CREATE FUNCTION f_for_exit() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE r RECORD; n int := 0;
BEGIN
	FOR r IN SELECT id FROM k ORDER BY id LOOP n := n + 1; EXIT; END LOOP;
	RETURN FOUND;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_for_exit();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// EXIT WHEN takes the same path out.
					Query: `CREATE FUNCTION f_for_exit_when() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; n int := 0;
BEGIN
	FOR r IN SELECT id FROM k ORDER BY id LOOP
		n := n + 1;
		EXIT WHEN r.id = 1;
	END LOOP;
	IF NOT FOUND THEN RETURN -1; END IF;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_for_exit_when();`,
					Expected: []sql.Row{{1}},
				},
				{
					// A labelled EXIT from inside the inner loop jumps over the inner loop's ScopeEnd on
					// its way to the outer one's, so the inner loop is left by the Goto rather than by
					// reaching its own ScopeEnd.
					Query: `CREATE FUNCTION f_labelled_exit() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; s RECORD; n int := 0;
BEGIN
	<<outer>>
	FOR r IN SELECT id FROM k ORDER BY id LOOP
		FOR s IN SELECT id FROM k ORDER BY id LOOP
			n := n + 1;
			EXIT outer;
		END LOOP;
	END LOOP;
	IF NOT FOUND THEN RETURN -1; END IF;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_labelled_exit();`,
					Expected: []sql.Row{{1}},
				},
				{
					// A loop entered twice re-runs its query each time rather than resuming where the
					// first pass left the cursor.
					Query: `CREATE FUNCTION f_reentered_loop() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; i int; n int := 0;
BEGIN
	FOR i IN 1..2 LOOP
		FOR r IN SELECT id FROM k ORDER BY id LOOP n := n + 1; END LOOP;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_reentered_loop();`,
					Expected: []sql.Row{{4}},
				},
			},
		},
		{
			Name: "integer FOR loops set FOUND when they exit",
			SetUpScript: []string{
				`CREATE TABLE k (id int);`,
				`INSERT INTO k VALUES (1);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_fori_hit() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE i int;
BEGIN FOR i IN 1..3 LOOP END LOOP; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_fori_hit();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// A range that is empty never runs the body, so the loop reports that.
					Query: `CREATE FUNCTION f_fori_miss() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE i int;
BEGIN FOR i IN 1..0 LOOP END LOOP; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_fori_miss();`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// An empty range does not leave an earlier statement's FOUND alone, it overwrites it.
					Query: `CREATE FUNCTION f_fori_overwrites() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE i int; v int;
BEGIN
	SELECT id INTO v FROM k WHERE id = 1;
	FOR i IN 1..0 LOOP END LOOP;
	RETURN FOUND;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_fori_overwrites();`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query: `CREATE FUNCTION f_fori_reverse() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE i int;
BEGIN FOR i IN REVERSE 3..1 LOOP END LOOP; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_fori_reverse();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// Leaving through EXIT still reports that the body ran.
					Query: `CREATE FUNCTION f_fori_exit() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE i int; n int := 0;
BEGIN FOR i IN 1..3 LOOP n := n + 1; EXIT; END LOOP; RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_fori_exit();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// The loop does not touch FOUND while it is running, only when it exits.
					Query: `CREATE FUNCTION f_fori_in_loop() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE i int; res boolean;
BEGIN FOR i IN 1..2 LOOP res := FOUND; END LOOP; RETURN res; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_fori_in_loop();`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// A FOR loop nested in a FOR loop reports its own exit, and the outer one then
					// reports over it.
					Query: `CREATE FUNCTION f_fori_nested() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE i int; j int;
BEGIN
	FOR i IN 1..2 LOOP
		FOR j IN 1..0 LOOP END LOOP;
	END LOOP;
	RETURN FOUND;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_fori_nested();`,
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "WHILE and plain LOOP leave FOUND alone",
			SetUpScript: []string{
				`CREATE TABLE k (id int);`,
				`INSERT INTO k VALUES (1);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// A WHILE loop is not a FOR loop, so it does not report anything on exit.
					Query: `CREATE FUNCTION f_while_keeps() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int; i int := 0;
BEGIN
	SELECT id INTO v FROM k WHERE id = 1;
	WHILE i < 3 LOOP i := i + 1; END LOOP;
	RETURN FOUND;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_while_keeps();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// A WHILE whose body never runs cannot turn FOUND on either.
					Query: `CREATE FUNCTION f_while_no_body() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN
	SELECT id INTO v FROM k WHERE id = 99;
	WHILE false LOOP NULL; END LOOP;
	RETURN FOUND;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_while_no_body();`,
					Expected: []sql.Row{{"f"}},
				},
				{
					// The same for an unconditional LOOP.
					Query: `CREATE FUNCTION f_loop_keeps() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN
	SELECT id INTO v FROM k WHERE id = 1;
	LOOP EXIT; END LOOP;
	RETURN FOUND;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_loop_keeps();`,
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "RETURN QUERY sets FOUND",
			SetUpScript: []string{
				`CREATE TABLE k (id int);`,
				`INSERT INTO k VALUES (1), (2);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// FOUND is reported through an error rather than the result set, because returning
					// a set of a scalar type from RETURN QUERY is separately broken.
					Query: `CREATE FUNCTION f_rq_miss() RETURNS SETOF int LANGUAGE plpgsql AS $$
BEGIN
	RETURN QUERY SELECT id FROM k WHERE id = 99;
	IF NOT FOUND THEN RAISE EXCEPTION 'no rows'; END IF;
	RAISE EXCEPTION 'had rows';
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM f_rq_miss();`,
					ExpectedErr: `no rows`,
				},
				{
					Query: `CREATE FUNCTION f_rq_hit() RETURNS SETOF int LANGUAGE plpgsql AS $$
BEGIN
	RETURN QUERY SELECT id FROM k ORDER BY id;
	IF NOT FOUND THEN RAISE EXCEPTION 'no rows'; END IF;
	RAISE EXCEPTION 'had rows';
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM f_rq_hit();`,
					ExpectedErr: `had rows`,
				},
			},
		},
		{
			// Modelled on reporting_exception_supersession_consistency from the dump that motivated this
			// work: a RECORD target, a NOT FOUND guard, then field comparisons against NEW.
			Name: "trigger combining a RECORD target with a NOT FOUND guard",
			SetUpScript: []string{
				`CREATE TABLE reporting_exception (
	id int PRIMARY KEY, org_id int, register_key text, entry_key text,
	version int, superseded_by_id int);`,
				`CREATE FUNCTION resc() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
	successor RECORD;
BEGIN
	IF NEW."superseded_by_id" IS NULL THEN
		RETURN NULL;
	END IF;

	SELECT "register_key", "entry_key", "version"
	  INTO successor
	  FROM "reporting_exception"
	 WHERE "id" = NEW."superseded_by_id"
	   AND "org_id" = NEW."org_id";

	IF NOT FOUND THEN
		RETURN NULL;
	END IF;

	IF successor."version" <= NEW."version" THEN
		RAISE EXCEPTION 'reporting_exception %: successor must carry a HIGHER version (this row is version %, successor is version %)',
			NEW."id", NEW."version", successor."version";
	END IF;

	IF successor."register_key" IS DISTINCT FROM NEW."register_key"
	   OR successor."entry_key" IS DISTINCT FROM NEW."entry_key" THEN
		RAISE EXCEPTION 'reporting_exception %: successor must supersede the SAME entry', NEW."id";
	END IF;

	RETURN NULL;
END; $$;`,
				`CREATE TRIGGER trg AFTER INSERT ON reporting_exception FOR EACH ROW EXECUTE FUNCTION resc();`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// No successor pointer at all, so the trigger returns before the query.
					Query:    `INSERT INTO reporting_exception VALUES (1, 7, 'rk', 'ek', 1, NULL);`,
					Expected: []sql.Row{},
				},
				{
					// The successor row does not exist: the NOT FOUND guard accepts the insert.
					Query:    `INSERT INTO reporting_exception VALUES (2, 7, 'rk', 'ek', 2, 999);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO reporting_exception VALUES (3, 7, 'rk', 'ek', 9, NULL);`,
					Expected: []sql.Row{},
				},
				{
					// A successor with a higher version and matching keys is accepted.
					Query:    `INSERT INTO reporting_exception VALUES (4, 7, 'rk', 'ek', 5, 3);`,
					Expected: []sql.Row{},
				},
				{
					Query:       `INSERT INTO reporting_exception VALUES (5, 7, 'rk', 'ek', 50, 3);`,
					ExpectedErr: `successor must carry a HIGHER version`,
				},
				{
					Query:       `INSERT INTO reporting_exception VALUES (6, 7, 'other', 'ek', 1, 3);`,
					ExpectedErr: `must supersede the SAME entry`,
				},
				{
					Query:    `SELECT id FROM reporting_exception ORDER BY id;`,
					Expected: []sql.Row{{1}, {2}, {3}, {4}},
				},
			},
		},
		{
			Name: "FOUND survives a nested block, and can be shadowed",
			SetUpScript: []string{
				`CREATE TABLE k (id int);`,
				`INSERT INTO k VALUES (1);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_nested() RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN
	BEGIN SELECT id INTO v FROM k WHERE id = 1; END;
	RETURN FOUND;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_nested();`,
					Expected: []sql.Row{{"t"}},
				},
				{
					// A variable the user declares as `found` shadows the built-in.
					Query: `CREATE FUNCTION f_shadow() RETURNS int LANGUAGE plpgsql AS $$
DECLARE found int := 5;
BEGIN RETURN found; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_shadow();`,
					Expected: []sql.Row{{5}},
				},
				{
					// A parameter named `found` is itself shadowed by the built-in, which PL/pgSQL
					// creates after the parameters.
					Query: `CREATE FUNCTION f_param(found boolean) RETURNS boolean LANGUAGE plpgsql AS $$
BEGIN RETURN FOUND; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_param(true);`,
					Expected: []sql.Row{{"f"}},
				},
			},
		},
		{
			// A labelled CONTINUE of an outer loop terminates the inner loop, and PostgreSQL treats that
			// as the inner loop exiting: it reports FOUND on the way out, like any other way of leaving
			// it. The value is observable at the top of the outer loop's next iteration, which is where
			// the CONTINUE lands.
			Name: "a labelled CONTINUE reports the inner loop's FOUND",
			SetUpScript: []string{
				`CREATE TABLE cf (id int);`,
				`INSERT INTO cf VALUES (1), (2), (3);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_cont_fori() RETURNS text LANGUAGE plpgsql AS $$
DECLARE i int; j int; acc text := '';
BEGIN
	<<outer_loop>>
	FOR i IN 1..3 LOOP
		acc := acc || CASE WHEN FOUND THEN 't' ELSE 'f' END;
		FOR j IN 1..3 LOOP
			CONTINUE outer_loop WHEN j = 1;
		END LOOP;
	END LOOP;
	RETURN acc;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// FOUND is false at the top of the first iteration and true afterwards, since by then
					// the inner loop has run its body once and been left by the CONTINUE.
					Query:    `SELECT f_cont_fori();`,
					Expected: []sql.Row{{"ftt"}},
				},
				{
					Query: `CREATE FUNCTION f_cont_fors() RETURNS text LANGUAGE plpgsql AS $$
DECLARE i int; r RECORD; acc text := '';
BEGIN
	<<outer_loop>>
	FOR i IN 1..3 LOOP
		acc := acc || CASE WHEN FOUND THEN 't' ELSE 'f' END;
		FOR r IN SELECT id FROM cf ORDER BY id LOOP
			CONTINUE outer_loop WHEN r.id = 1;
		END LOOP;
	END LOOP;
	RETURN acc;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// The same holds for a FOR..IN..SELECT loop, whose cursor the CONTINUE also closes.
					Query:    `SELECT f_cont_fors();`,
					Expected: []sql.Row{{"ftt"}},
				},
				{
					Query: `CREATE FUNCTION f_cont_empty() RETURNS text LANGUAGE plpgsql AS $$
DECLARE i int; r RECORD; acc text := '';
BEGIN
	<<outer_loop>>
	FOR i IN 1..3 LOOP
		acc := acc || CASE WHEN FOUND THEN 't' ELSE 'f' END;
		FOR r IN SELECT id FROM cf WHERE id < 0 LOOP
			NULL;
		END LOOP;
		CONTINUE outer_loop;
	END LOOP;
	RETURN acc;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// What the inner loop reports is whether its body ran, not that it was left, so an
					// inner loop matching nothing leaves FOUND false on every iteration.
					Query:    `SELECT f_cont_empty();`,
					Expected: []sql.Row{{"fff"}},
				},
			},
		},
	})
}
