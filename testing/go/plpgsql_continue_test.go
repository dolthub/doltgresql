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

// TestPlpgsqlContinue covers CONTINUE (bare, WHEN, and labelled) in every kind of PL/pgSQL loop. All
// expectations were verified against PostgreSQL 16. Every function carries a guard counter that returns a
// sentinel once the loop exceeds its iteration count, so a loop that restarts instead of advancing fails an
// assertion rather than hanging the test.
func TestPlpgsqlContinue(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "CONTINUE WHEN in an integer FOR loop",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_fori_when() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; n int := 0; guard int := 0;
BEGIN
	FOR i IN 1..5 LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		CONTINUE WHEN i % 2 = 0;
		n := n + i;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// The even values are skipped, so this is 1 + 3 + 5.
					Query:    `SELECT f_fori_when();`,
					Expected: []sql.Row{{9}},
				},
			},
		},
		{
			Name: "bare CONTINUE in an integer FOR loop",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_fori_bare() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; n int := 0; guard int := 0;
BEGIN
	FOR i IN 1..5 LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		IF i = 3 THEN
			CONTINUE;
		END IF;
		n := n + i;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// Only i = 3 is skipped, so this is 1 + 2 + 4 + 5.
					Query:    `SELECT f_fori_bare();`,
					Expected: []sql.Row{{12}},
				},
			},
		},
		{
			Name: "CONTINUE preserves the integer FOR loop variable",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_fori_var() RETURNS text LANGUAGE plpgsql AS $$
DECLARE i int; acc text := ''; guard int := 0;
BEGIN
	FOR i IN 1..4 LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN 'guard'; END IF;
		CONTINUE WHEN i = 2;
		acc := acc || i::text;
	END LOOP;
	RETURN acc;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// i = 2 is skipped and the loop variable keeps counting, so this is 1, 3, 4.
					Query:    `SELECT f_fori_var();`,
					Expected: []sql.Row{{"134"}},
				},
			},
		},
		{
			Name: "CONTINUE as the last statement of an integer FOR loop body",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_fori_last() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; n int := 0; guard int := 0;
BEGIN
	FOR i IN 1..4 LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		n := n + i;
		CONTINUE WHEN i > 0;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// The CONTINUE always fires, but nothing follows it, so every value is summed.
					Query:    `SELECT f_fori_last();`,
					Expected: []sql.Row{{10}},
				},
			},
		},
		{
			Name: "CONTINUE in an integer FOR loop with BY",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_fori_by() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; n int := 0; guard int := 0;
BEGIN
	FOR i IN 1..10 BY 3 LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		CONTINUE WHEN i = 4;
		n := n + i;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// The loop visits 1, 4, 7, 10 and skips 4, so this is 1 + 7 + 10.
					Query:    `SELECT f_fori_by();`,
					Expected: []sql.Row{{18}},
				},
			},
		},
		{
			Name: "CONTINUE in a REVERSE integer FOR loop with BY",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_fori_reverse() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; n int := 0; guard int := 0;
BEGIN
	FOR i IN REVERSE 10..1 BY 3 LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		CONTINUE WHEN i = 7;
		n := n + i;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// The loop counts down 10, 7, 4, 1 and skips 7, so this is 10 + 4 + 1.
					Query:    `SELECT f_fori_reverse();`,
					Expected: []sql.Row{{15}},
				},
			},
		},
		{
			Name: "CONTINUE and EXIT in the same integer FOR loop",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_fori_exit() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; n int := 0; guard int := 0;
BEGIN
	FOR i IN 1..10 LOOP
		guard := guard + 1;
		IF guard > 25 THEN RETURN -99; END IF;
		CONTINUE WHEN i % 2 = 0;
		EXIT WHEN i > 6;
		n := n + i;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// Even values are skipped and the loop exits at i = 7, so this is 1 + 3 + 5.
					Query:    `SELECT f_fori_exit();`,
					Expected: []sql.Row{{9}},
				},
			},
		},
		{
			Name: "bare CONTINUE in a nested integer FOR loop",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_fori_nested() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; j int; n int := 0; guard int := 0;
BEGIN
	FOR i IN 1..3 LOOP
		FOR j IN 1..3 LOOP
			guard := guard + 1;
			IF guard > 30 THEN RETURN -99; END IF;
			CONTINUE WHEN j = 2;
			n := n + 1;
		END LOOP;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// The inner CONTINUE only affects the inner loop: 3 outer * 2 counted inner.
					Query:    `SELECT f_fori_nested();`,
					Expected: []sql.Row{{6}},
				},
			},
		},
		{
			Name: "labelled CONTINUE of an outer integer FOR loop",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_fori_label() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; j int; n int := 0; guard int := 0;
BEGIN
	<<outer_loop>>
	FOR i IN 1..3 LOOP
		FOR j IN 1..3 LOOP
			guard := guard + 1;
			IF guard > 30 THEN RETURN -99; END IF;
			CONTINUE outer_loop WHEN j = 2;
			n := n + (i * 10 + j);
		END LOOP;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// Each outer iteration counts j = 1 and then advances i, so this is 11 + 21 + 31.
					Query:    `SELECT f_fori_label();`,
					Expected: []sql.Row{{63}},
				},
			},
		},
		{
			Name: "CONTINUE in a WHILE loop",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_while() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int := 0; n int := 0; guard int := 0;
BEGIN
	WHILE i < 5 LOOP
		i := i + 1;
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		CONTINUE WHEN i = 3;
		n := n + i;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// Only i = 3 is skipped, so this is 1 + 2 + 4 + 5.
					Query:    `SELECT f_while();`,
					Expected: []sql.Row{{12}},
				},
			},
		},
		{
			Name: "labelled CONTINUE of an outer WHILE loop",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_while_label() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int := 0; j int; n int := 0; guard int := 0;
BEGIN
	<<outer_loop>>
	WHILE i < 3 LOOP
		i := i + 1;
		FOR j IN 1..3 LOOP
			guard := guard + 1;
			IF guard > 30 THEN RETURN -99; END IF;
			CONTINUE outer_loop WHEN j = 2;
			n := n + (i * 10 + j);
		END LOOP;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// The labelled CONTINUE re-tests the WHILE condition, so this is 11 + 21 + 31.
					Query:    `SELECT f_while_label();`,
					Expected: []sql.Row{{63}},
				},
			},
		},
		{
			Name: "CONTINUE in a plain LOOP",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_loop() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int := 0; n int := 0; guard int := 0;
BEGIN
	LOOP
		i := i + 1;
		EXIT WHEN i > 5;
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		CONTINUE WHEN i = 3;
		n := n + i;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// Only i = 3 is skipped, so this is 1 + 2 + 4 + 5.
					Query:    `SELECT f_loop();`,
					Expected: []sql.Row{{12}},
				},
			},
		},
		{
			Name: "CONTINUE WHEN in a FOR..IN..SELECT loop",
			SetUpScript: []string{
				`CREATE TABLE c1 (id int, val int);`,
				`INSERT INTO c1 VALUES (1, 10), (2, 20), (3, 30);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_fors_when() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; n int := 0; guard int := 0;
BEGIN
	FOR r IN SELECT id, val FROM c1 ORDER BY id LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		CONTINUE WHEN r.id = 2;
		n := n + r.val;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// The second row is skipped, so this is 10 + 30.
					Query:    `SELECT f_fors_when();`,
					Expected: []sql.Row{{40}},
				},
			},
		},
		{
			Name: "bare CONTINUE in a FOR..IN..SELECT loop",
			SetUpScript: []string{
				`CREATE TABLE c2 (id int, val int);`,
				`INSERT INTO c2 VALUES (1, 10), (2, 20), (3, 30);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_fors_bare() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; n int := 0; guard int := 0;
BEGIN
	FOR r IN SELECT id, val FROM c2 ORDER BY id LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		IF r.id = 3 THEN
			CONTINUE;
		END IF;
		n := n + r.val;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// The third row is skipped, so this is 10 + 20.
					Query:    `SELECT f_fors_bare();`,
					Expected: []sql.Row{{30}},
				},
			},
		},
		{
			Name: "CONTINUE and EXIT in the same FOR..IN..SELECT loop",
			SetUpScript: []string{
				`CREATE TABLE c3 (id int, val int);`,
				`INSERT INTO c3 VALUES (1, 10), (2, 20), (3, 30), (4, 40), (5, 50);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_fors_exit() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; n int := 0; guard int := 0;
BEGIN
	FOR r IN SELECT id, val FROM c3 ORDER BY id LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		CONTINUE WHEN r.id = 2;
		EXIT WHEN r.id = 4;
		n := n + r.val;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// The second row is skipped and the loop exits on the fourth, so this is 10 + 30.
					Query:    `SELECT f_fors_exit();`,
					Expected: []sql.Row{{40}},
				},
			},
		},
		{
			Name: "labelled CONTINUE of an outer FOR..IN..SELECT loop",
			SetUpScript: []string{
				`CREATE TABLE c4 (id int, val int);`,
				`INSERT INTO c4 VALUES (1, 10), (2, 20), (3, 30);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_fors_label() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; s RECORD; n int := 0; guard int := 0;
BEGIN
	<<outer_loop>>
	FOR r IN SELECT id, val FROM c4 ORDER BY id LOOP
		FOR s IN SELECT id FROM c4 ORDER BY id LOOP
			guard := guard + 1;
			IF guard > 30 THEN RETURN -99; END IF;
			CONTINUE outer_loop WHEN s.id = 2;
			n := n + (r.id * 10 + s.id);
		END LOOP;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// Each outer row counts its first inner row, so this is 11 + 21 + 31.
					Query:    `SELECT f_fors_label();`,
					Expected: []sql.Row{{63}},
				},
			},
		},
		{
			Name: "labelled CONTINUE of an outer integer FOR loop from a FOR..IN..SELECT loop",
			SetUpScript: []string{
				`CREATE TABLE c5 (id int);`,
				`INSERT INTO c5 VALUES (1), (2), (3);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_mixed_label() RETURNS int LANGUAGE plpgsql AS $$
DECLARE i int; r RECORD; n int := 0; guard int := 0;
BEGIN
	<<outer_loop>>
	FOR i IN 1..3 LOOP
		FOR r IN SELECT id FROM c5 ORDER BY id LOOP
			guard := guard + 1;
			IF guard > 30 THEN RETURN -99; END IF;
			CONTINUE outer_loop WHEN r.id = 2;
			n := n + (i * 10 + r.id);
		END LOOP;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// The labelled CONTINUE advances the integer FOR loop, so this is 11 + 21 + 31.
					Query:    `SELECT f_mixed_label();`,
					Expected: []sql.Row{{63}},
				},
			},
		},
		{
			Name: "CONTINUE as the last statement of a FOR..IN..SELECT loop body",
			SetUpScript: []string{
				`CREATE TABLE c6 (id int, val int);`,
				`INSERT INTO c6 VALUES (1, 10), (2, 20), (3, 30);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_fors_last() RETURNS int LANGUAGE plpgsql AS $$
DECLARE r RECORD; n int := 0; guard int := 0;
BEGIN
	FOR r IN SELECT id, val FROM c6 ORDER BY id LOOP
		guard := guard + 1;
		IF guard > 20 THEN RETURN -99; END IF;
		n := n + r.val;
		CONTINUE WHEN r.id > 0;
	END LOOP;
	RETURN n;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					// The CONTINUE always fires, but nothing follows it, so every row is summed.
					Query:    `SELECT f_fors_last();`,
					Expected: []sql.Row{{60}},
				},
			},
		},
	})
}
