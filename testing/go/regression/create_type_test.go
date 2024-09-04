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

package regression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestCreateType(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_create_type)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_create_type,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv libdir PG_LIBDIR
\getenv dlsuffix PG_DLSUFFIX
\set regresslib :libdir '/regress' :dlsuffix
CREATE FUNCTION widget_in(cstring)
   RETURNS widget
   AS :'regresslib'
   LANGUAGE C STRICT IMMUTABLE;`,
			},
			{
				Statement: `DETAIL:  Creating a shell type definition.
CREATE FUNCTION widget_out(widget)
   RETURNS cstring
   AS :'regresslib'
   LANGUAGE C STRICT IMMUTABLE;`,
			},
			{
				Statement: `CREATE FUNCTION int44in(cstring)
   RETURNS city_budget
   AS :'regresslib'
   LANGUAGE C STRICT IMMUTABLE;`,
			},
			{
				Statement: `DETAIL:  Creating a shell type definition.
CREATE FUNCTION int44out(city_budget)
   RETURNS cstring
   AS :'regresslib'
   LANGUAGE C STRICT IMMUTABLE;`,
			},
			{
				Statement: `CREATE TYPE widget (
   internallength = 24,
   input = widget_in,
   output = widget_out,
   typmod_in = numerictypmodin,
   typmod_out = numerictypmodout,
   alignment = double
);`,
			},
			{
				Statement: `CREATE TYPE city_budget (
   internallength = 16,
   input = int44in,
   output = int44out,
   element = int4,
   category = 'x',   -- just to verify the system will take it
   preferred = true  -- ditto
);`,
			},
			{
				Statement: `CREATE TYPE shell;`,
			},
			{
				Statement:   `CREATE TYPE shell;   -- fail, type already present`,
				ErrorString: `type "shell" already exists`,
			},
			{
				Statement: `DROP TYPE shell;`,
			},
			{
				Statement:   `DROP TYPE shell;     -- fail, type not exist`,
				ErrorString: `type "shell" does not exist`,
			},
			{
				Statement: `CREATE TYPE myshell;`,
			},
			{
				Statement: `CREATE TYPE int42;`,
			},
			{
				Statement: `CREATE TYPE text_w_default;`,
			},
			{
				Statement: `CREATE FUNCTION int42_in(cstring)
   RETURNS int42
   AS 'int4in'
   LANGUAGE internal STRICT IMMUTABLE;`,
			},
			{
				Statement: `CREATE FUNCTION int42_out(int42)
   RETURNS cstring
   AS 'int4out'
   LANGUAGE internal STRICT IMMUTABLE;`,
			},
			{
				Statement: `CREATE FUNCTION text_w_default_in(cstring)
   RETURNS text_w_default
   AS 'textin'
   LANGUAGE internal STRICT IMMUTABLE;`,
			},
			{
				Statement: `CREATE FUNCTION text_w_default_out(text_w_default)
   RETURNS cstring
   AS 'textout'
   LANGUAGE internal STRICT IMMUTABLE;`,
			},
			{
				Statement: `CREATE TYPE int42 (
   internallength = 4,
   input = int42_in,
   output = int42_out,
   alignment = int4,
   default = 42,
   passedbyvalue
);`,
			},
			{
				Statement: `CREATE TYPE text_w_default (
   internallength = variable,
   input = text_w_default_in,
   output = text_w_default_out,
   alignment = int4,
   default = 'zippo'
);`,
			},
			{
				Statement: `CREATE TABLE default_test (f1 text_w_default, f2 int42);`,
			},
			{
				Statement: `INSERT INTO default_test DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT * FROM default_test;`,
				Results:   []sql.Row{{`zippo`, 42}},
			},
			{
				Statement: `CREATE TYPE bogus_type;`,
			},
			{
				Statement: `CREATE TYPE bogus_type (
	"Internallength" = 4,
	"Input" = int42_in,
	"Output" = int42_out,
	"Alignment" = int4,
	"Default" = 42,
	"Passedbyvalue"
);`,
			},
			{
				Statement: `LINE 2:  "Internallength" = 4,
         ^
LINE 3:  "Input" = int42_in,
         ^
LINE 4:  "Output" = int42_out,
         ^
LINE 5:  "Alignment" = int4,
         ^
LINE 6:  "Default" = 42,
         ^
LINE 7:  "Passedbyvalue"
         ^
ERROR:  type input function must be specified
CREATE TYPE bogus_type (INPUT = array_in,
    OUTPUT = array_out,
    ELEMENT = int,
    INTERNALLENGTH = 32);`,
				ErrorString: `type input function array_in must return type bogus_type`,
			},
			{
				Statement: `DROP TYPE bogus_type;`,
			},
			{
				Statement: `CREATE TYPE bogus_type (INPUT = array_in,
    OUTPUT = array_out,
    ELEMENT = int,
    INTERNALLENGTH = 32);`,
				ErrorString: `type "bogus_type" does not exist`,
			},
			{
				Statement: `CREATE TYPE default_test_row AS (f1 text_w_default, f2 int42);`,
			},
			{
				Statement: `CREATE FUNCTION get_default_test() RETURNS SETOF default_test_row AS '
  SELECT * FROM default_test;`,
			},
			{
				Statement: `' LANGUAGE SQL;`,
			},
			{
				Statement: `SELECT * FROM get_default_test();`,
				Results:   []sql.Row{{`zippo`, 42}},
			},
			{
				Statement:   `COMMENT ON TYPE bad IS 'bad comment';`,
				ErrorString: `type "bad" does not exist`,
			},
			{
				Statement: `COMMENT ON TYPE default_test_row IS 'good comment';`,
			},
			{
				Statement: `COMMENT ON TYPE default_test_row IS NULL;`,
			},
			{
				Statement:   `COMMENT ON COLUMN default_test_row.nope IS 'bad comment';`,
				ErrorString: `column "nope" of relation "default_test_row" does not exist`,
			},
			{
				Statement: `COMMENT ON COLUMN default_test_row.f1 IS 'good comment';`,
			},
			{
				Statement: `COMMENT ON COLUMN default_test_row.f1 IS NULL;`,
			},
			{
				Statement:   `CREATE TYPE text_w_default;		-- should fail`,
				ErrorString: `type "text_w_default" already exists`,
			},
			{
				Statement: `DROP TYPE default_test_row CASCADE;`,
			},
			{
				Statement: `DROP TABLE default_test;`,
			},
			{
				Statement: `CREATE TYPE base_type;`,
			},
			{
				Statement: `CREATE FUNCTION base_fn_in(cstring) RETURNS base_type AS 'boolin'
    LANGUAGE internal IMMUTABLE STRICT;`,
			},
			{
				Statement: `CREATE FUNCTION base_fn_out(base_type) RETURNS cstring AS 'boolout'
    LANGUAGE internal IMMUTABLE STRICT;`,
			},
			{
				Statement: `CREATE TYPE base_type(INPUT = base_fn_in, OUTPUT = base_fn_out);`,
			},
			{
				Statement:   `DROP FUNCTION base_fn_in(cstring); -- error`,
				ErrorString: `cannot drop function base_fn_in(cstring) because other objects depend on it`,
			},
			{
				Statement: `DETAIL:  type base_type depends on function base_fn_in(cstring)
function base_fn_out(base_type) depends on type base_type
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DROP FUNCTION base_fn_out(base_type); -- error`,
				ErrorString: `cannot drop function base_fn_out(base_type) because other objects depend on it`,
			},
			{
				Statement: `DETAIL:  type base_type depends on function base_fn_out(base_type)
function base_fn_in(cstring) depends on type base_type
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DROP TYPE base_type; -- error`,
				ErrorString: `cannot drop type base_type because other objects depend on it`,
			},
			{
				Statement: `DETAIL:  function base_fn_in(cstring) depends on type base_type
function base_fn_out(base_type) depends on type base_type
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DROP TYPE base_type CASCADE;`,
			},
			{
				Statement: `DETAIL:  drop cascades to function base_fn_in(cstring)
drop cascades to function base_fn_out(base_type)
CREATE TEMP TABLE mytab (foo widget(42,13,7));     -- should fail`,
				ErrorString: `invalid NUMERIC type modifier`,
			},
			{
				Statement: `CREATE TEMP TABLE mytab (foo widget(42,13));`,
			},
			{
				Statement: `SELECT format_type(atttypid,atttypmod) FROM pg_attribute
WHERE attrelid = 'mytab'::regclass AND attnum > 0;`,
				Results: []sql.Row{{`widget(42,13)`}},
			},
			{
				Statement: `INSERT INTO mytab VALUES ('(1,2,3)'), ('(-44,5.5,12)');`,
			},
			{
				Statement: `TABLE mytab;`,
				Results:   []sql.Row{{`(1,2,3)`}, {`(-44,5.5,12)`}},
			},
			{
				Statement: `select format_type('varchar'::regtype, 42);`,
				Results:   []sql.Row{{`character varying(38)`}},
			},
			{
				Statement: `select format_type('bpchar'::regtype, null);`,
				Results:   []sql.Row{{`character`}},
			},
			{
				Statement: `select format_type('bpchar'::regtype, -1);`,
				Results:   []sql.Row{{`bpchar`}},
			},
			{
				Statement: `CREATE FUNCTION pt_in_widget(point, widget)
   RETURNS bool
   AS :'regresslib'
   LANGUAGE C STRICT;`,
			},
			{
				Statement: `CREATE OPERATOR <% (
   leftarg = point,
   rightarg = widget,
   procedure = pt_in_widget,
   commutator = >% ,
   negator = >=%
);`,
			},
			{
				Statement: `SELECT point '(1,2)' <% widget '(0,0,3)' AS t,
       point '(1,2)' <% widget '(0,0,1)' AS f;`,
				Results: []sql.Row{{true, false}},
			},
			{
				Statement: `CREATE TABLE city (
	name		name,
	location 	box,
	budget 		city_budget
);`,
			},
			{
				Statement: `INSERT INTO city VALUES
('Podunk', '(1,2),(3,4)', '100,127,1000'),
('Gotham', '(1000,34),(1100,334)', '123456,127,-1000,6789');`,
			},
			{
				Statement: `TABLE city;`,
				Results:   []sql.Row{{`Podunk`, `(3,4),(1,2)`, `100,127,1000,0`}, {`Gotham`, `(1100,334),(1000,34)`, `123456,127,-1000,6789`}},
			},
			{
				Statement: `CREATE TYPE myvarchar;`,
			},
			{
				Statement: `CREATE FUNCTION myvarcharin(cstring, oid, integer) RETURNS myvarchar
LANGUAGE internal IMMUTABLE PARALLEL SAFE STRICT AS 'varcharin';`,
			},
			{
				Statement: `CREATE FUNCTION myvarcharout(myvarchar) RETURNS cstring
LANGUAGE internal IMMUTABLE PARALLEL SAFE STRICT AS 'varcharout';`,
			},
			{
				Statement: `CREATE FUNCTION myvarcharsend(myvarchar) RETURNS bytea
LANGUAGE internal STABLE PARALLEL SAFE STRICT AS 'varcharsend';`,
			},
			{
				Statement: `CREATE FUNCTION myvarcharrecv(internal, oid, integer) RETURNS myvarchar
LANGUAGE internal STABLE PARALLEL SAFE STRICT AS 'varcharrecv';`,
			},
			{
				Statement:   `ALTER TYPE myvarchar SET (storage = extended);`,
				ErrorString: `type "myvarchar" is only a shell`,
			},
			{
				Statement: `CREATE TYPE myvarchar (
    input = myvarcharin,
    output = myvarcharout,
    alignment = integer,
    storage = main
);`,
			},
			{
				Statement: `CREATE DOMAIN myvarchardom AS myvarchar;`,
			},
			{
				Statement:   `ALTER TYPE myvarchar SET (storage = plain);  -- not allowed`,
				ErrorString: `cannot change type's storage to PLAIN`,
			},
			{
				Statement: `ALTER TYPE myvarchar SET (storage = extended);`,
			},
			{
				Statement: `ALTER TYPE myvarchar SET (
    send = myvarcharsend,
    receive = myvarcharrecv,
    typmod_in = varchartypmodin,
    typmod_out = varchartypmodout,
    -- these are bogus, but it's safe as long as we don't use the type:
    analyze = ts_typanalyze,
    subscript = raw_array_subscript_handler
);`,
			},
			{
				Statement: `SELECT typinput, typoutput, typreceive, typsend, typmodin, typmodout,
       typanalyze, typsubscript, typstorage
FROM pg_type WHERE typname = 'myvarchar';`,
				Results: []sql.Row{{`myvarcharin`, `myvarcharout`, `myvarcharrecv`, `myvarcharsend`, `varchartypmodin`, `varchartypmodout`, `ts_typanalyze`, `raw_array_subscript_handler`, `x`}},
			},
			{
				Statement: `SELECT typinput, typoutput, typreceive, typsend, typmodin, typmodout,
       typanalyze, typsubscript, typstorage
FROM pg_type WHERE typname = '_myvarchar';`,
				Results: []sql.Row{{`array_in`, `array_out`, `array_recv`, `array_send`, `varchartypmodin`, `varchartypmodout`, `array_typanalyze`, `array_subscript_handler`, `x`}},
			},
			{
				Statement: `SELECT typinput, typoutput, typreceive, typsend, typmodin, typmodout,
       typanalyze, typsubscript, typstorage
FROM pg_type WHERE typname = 'myvarchardom';`,
				Results: []sql.Row{{`domain_in`, `myvarcharout`, `domain_recv`, `myvarcharsend`, `-`, `-`, `ts_typanalyze`, `-`, `x`}},
			},
			{
				Statement: `SELECT typinput, typoutput, typreceive, typsend, typmodin, typmodout,
       typanalyze, typsubscript, typstorage
FROM pg_type WHERE typname = '_myvarchardom';`,
				Results: []sql.Row{{`array_in`, `array_out`, `array_recv`, `array_send`, `-`, `-`, `array_typanalyze`, `array_subscript_handler`, `x`}},
			},
			{
				Statement:   `DROP FUNCTION myvarcharsend(myvarchar);  -- fail`,
				ErrorString: `cannot drop function myvarcharsend(myvarchar) because other objects depend on it`,
			},
			{
				Statement: `DETAIL:  type myvarchar depends on function myvarcharsend(myvarchar)
function myvarcharin(cstring,oid,integer) depends on type myvarchar
function myvarcharout(myvarchar) depends on type myvarchar
function myvarcharrecv(internal,oid,integer) depends on type myvarchar
type myvarchardom depends on function myvarcharsend(myvarchar)
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DROP TYPE myvarchar;  -- fail`,
				ErrorString: `cannot drop type myvarchar because other objects depend on it`,
			},
			{
				Statement: `DETAIL:  function myvarcharin(cstring,oid,integer) depends on type myvarchar
function myvarcharout(myvarchar) depends on type myvarchar
function myvarcharsend(myvarchar) depends on type myvarchar
function myvarcharrecv(internal,oid,integer) depends on type myvarchar
type myvarchardom depends on type myvarchar
HINT:  Use DROP ... CASCADE to drop the dependent objects too.
DROP TYPE myvarchar CASCADE;`,
			},
			{
				Statement: `DETAIL:  drop cascades to function myvarcharin(cstring,oid,integer)
drop cascades to function myvarcharout(myvarchar)
drop cascades to function myvarcharsend(myvarchar)
drop cascades to function myvarcharrecv(internal,oid,integer)
drop cascades to type myvarchardom`,
			},
		},
	})
}
