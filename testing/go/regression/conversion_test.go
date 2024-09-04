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

func TestConversion(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_conversion)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_conversion,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv libdir PG_LIBDIR
\getenv dlsuffix PG_DLSUFFIX
\set regresslib :libdir '/regress' :dlsuffix
CREATE FUNCTION test_enc_conversion(bytea, name, name, bool, validlen OUT int, result OUT bytea)
    AS :'regresslib', 'test_enc_conversion'
    LANGUAGE C STRICT;`,
			},
			{
				Statement: `CREATE USER regress_conversion_user WITH NOCREATEDB NOCREATEROLE;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_conversion_user;`,
			},
			{
				Statement: `CREATE CONVERSION myconv FOR 'LATIN1' TO 'UTF8' FROM iso8859_1_to_utf8;`,
			},
			{
				Statement:   `CREATE CONVERSION myconv FOR 'LATIN1' TO 'UTF8' FROM iso8859_1_to_utf8;`,
				ErrorString: `conversion "myconv" already exists`,
			},
			{
				Statement: `CREATE DEFAULT CONVERSION public.mydef FOR 'LATIN1' TO 'UTF8' FROM iso8859_1_to_utf8;`,
			},
			{
				Statement:   `CREATE DEFAULT CONVERSION public.mydef2 FOR 'LATIN1' TO 'UTF8' FROM iso8859_1_to_utf8;`,
				ErrorString: `default conversion for LATIN1 to UTF8 already exists`,
			},
			{
				Statement:   `COMMENT ON CONVERSION myconv_bad IS 'foo';`,
				ErrorString: `conversion "myconv_bad" does not exist`,
			},
			{
				Statement: `COMMENT ON CONVERSION myconv IS 'bar';`,
			},
			{
				Statement: `COMMENT ON CONVERSION myconv IS NULL;`,
			},
			{
				Statement: `DROP CONVERSION myconv;`,
			},
			{
				Statement: `DROP CONVERSION mydef;`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `DROP USER regress_conversion_user;`,
			},
			{
				Statement: `create or replace function test_conv(
  input IN bytea,
  src_encoding IN text,
  dst_encoding IN text,
  result OUT bytea,
  errorat OUT bytea,
  error OUT text)
language plpgsql as
$$
declare
  validlen int;`,
			},
			{
				Statement: `begin
  -- First try to perform the conversion with noError = false. If that errors out,
  -- capture the error message, and try again with noError = true. The second call
  -- should succeed and return the position of the error, return that too.
  begin
    select * into validlen, result from test_enc_conversion(input, src_encoding, dst_encoding, false);`,
			},
			{
				Statement: `    errorat = NULL;`,
			},
			{
				Statement: `    error := NULL;`,
			},
			{
				Statement: `  exception when others then
    error := sqlerrm;`,
			},
			{
				Statement: `    select * into validlen, result from test_enc_conversion(input, src_encoding, dst_encoding, true);`,
			},
			{
				Statement: `    errorat = substr(input, validlen + 1);`,
			},
			{
				Statement: `  end;`,
			},
			{
				Statement: `  return;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `CREATE TABLE utf8_verification_inputs (inbytes bytea, description text PRIMARY KEY);`,
			},
			{
				Statement: `insert into utf8_verification_inputs  values
  ('\x66006f',	'NUL byte'),
  ('\xaf',		'bare continuation'),
  ('\xc5',		'missing second byte in 2-byte char'),
  ('\xc080',	'smallest 2-byte overlong'),
  ('\xc1bf',	'largest 2-byte overlong'),
  ('\xc280',	'next 2-byte after overlongs'),
  ('\xdfbf',	'largest 2-byte'),
  ('\xe9af',	'missing third byte in 3-byte char'),
  ('\xe08080',	'smallest 3-byte overlong'),
  ('\xe09fbf',	'largest 3-byte overlong'),
  ('\xe0a080',	'next 3-byte after overlong'),
  ('\xed9fbf',	'last before surrogates'),
  ('\xeda080',	'smallest surrogate'),
  ('\xedbfbf',	'largest surrogate'),
  ('\xee8080',	'next after surrogates'),
  ('\xefbfbf',	'largest 3-byte'),
  ('\xf1afbf',	'missing fourth byte in 4-byte char'),
  ('\xf0808080',	'smallest 4-byte overlong'),
  ('\xf08fbfbf',	'largest 4-byte overlong'),
  ('\xf0908080',	'next 4-byte after overlong'),
  ('\xf48fbfbf',	'largest 4-byte'),
  ('\xf4908080',	'smallest too large'),
  ('\xfa9a9a8a8a',	'5-byte');`,
			},
			{
				Statement: `select description, (test_conv(inbytes, 'utf8', 'utf8')).* from utf8_verification_inputs;`,
				Results:   []sql.Row{{`NUL byte`, `\x66`, `\x006f`, `invalid byte sequence for encoding "UTF8": 0x00`}, {`bare continuation`, `\x`, `\xaf`, `invalid byte sequence for encoding "UTF8": 0xaf`}, {`missing second byte in 2-byte char`, `\x`, `\xc5`, `invalid byte sequence for encoding "UTF8": 0xc5`}, {`smallest 2-byte overlong`, `\x`, `\xc080`, `invalid byte sequence for encoding "UTF8": 0xc0 0x80`}, {`largest 2-byte overlong`, `\x`, `\xc1bf`, `invalid byte sequence for encoding "UTF8": 0xc1 0xbf`}, {`next 2-byte after overlongs`, `\xc280`, ``, ``}, {`largest 2-byte`, `\xdfbf`, ``, ``}, {`missing third byte in 3-byte char`, `\x`, `\xe9af`, `invalid byte sequence for encoding "UTF8": 0xe9 0xaf`}, {`smallest 3-byte overlong`, `\x`, `\xe08080`, `invalid byte sequence for encoding "UTF8": 0xe0 0x80 0x80`}, {`largest 3-byte overlong`, `\x`, `\xe09fbf`, `invalid byte sequence for encoding "UTF8": 0xe0 0x9f 0xbf`}, {`next 3-byte after overlong`, `\xe0a080`, ``, ``}, {`last before surrogates`, `\xed9fbf`, ``, ``}, {`smallest surrogate`, `\x`, `\xeda080`, `invalid byte sequence for encoding "UTF8": 0xed 0xa0 0x80`}, {`largest surrogate`, `\x`, `\xedbfbf`, `invalid byte sequence for encoding "UTF8": 0xed 0xbf 0xbf`}, {`next after surrogates`, `\xee8080`, ``, ``}, {`largest 3-byte`, `\xefbfbf`, ``, ``}, {`missing fourth byte in 4-byte char`, `\x`, `\xf1afbf`, `invalid byte sequence for encoding "UTF8": 0xf1 0xaf 0xbf`}, {`smallest 4-byte overlong`, `\x`, `\xf0808080`, `invalid byte sequence for encoding "UTF8": 0xf0 0x80 0x80 0x80`}, {`largest 4-byte overlong`, `\x`, `\xf08fbfbf`, `invalid byte sequence for encoding "UTF8": 0xf0 0x8f 0xbf 0xbf`}, {`next 4-byte after overlong`, `\xf0908080`, ``, ``}, {`largest 4-byte`, `\xf48fbfbf`, ``, ``}, {`smallest too large`, `\x`, `\xf4908080`, `invalid byte sequence for encoding "UTF8": 0xf4 0x90 0x80 0x80`}, {`5-byte`, `\x`, `\xfa9a9a8a8a`, `invalid byte sequence for encoding "UTF8": 0xfa`}},
			},
			{
				Statement: `with test_bytes as (
  select
    inbytes,
    description,
    (test_conv(inbytes || repeat('.', 3)::bytea, 'utf8', 'utf8')).error
  from utf8_verification_inputs
), test_padded as (
  select
    description,
    (test_conv(inbytes || repeat('.', 64)::bytea, 'utf8', 'utf8')).error
  from test_bytes
)
select
  description,
  b.error as orig_error,
  p.error as error_after_padding
from test_padded p
join test_bytes b
using (description)
where p.error is distinct from b.error
order by description;`,
				Results: []sql.Row{},
			},
			{
				Statement: `with test_bytes as (
  select
    inbytes,
    description,
    (test_conv(inbytes || repeat('.', 3)::bytea, 'utf8', 'utf8')).error
  from utf8_verification_inputs
), test_padded as (
  select
    description,
    (test_conv(repeat('.', 64 - length(inbytes))::bytea || inbytes || repeat('.', 64)::bytea, 'utf8', 'utf8')).error
  from test_bytes
)
select
  description,
  b.error as orig_error,
  p.error as error_after_padding
from test_padded p
join test_bytes b
using (description)
where p.error is distinct from b.error
order by description;`,
				Results: []sql.Row{},
			},
			{
				Statement: `with test_bytes as (
  select
    inbytes,
    description,
    (test_conv(inbytes || repeat('.', 3)::bytea, 'utf8', 'utf8')).error
  from utf8_verification_inputs
), test_padded as (
  select
    description,
    (test_conv(repeat('.', 64)::bytea || inbytes || repeat('.', 3)::bytea, 'utf8', 'utf8')).error
  from test_bytes
)
select
  description,
  b.error as orig_error,
  p.error as error_after_padding
from test_padded p
join test_bytes b
using (description)
where p.error is distinct from b.error
order by description;`,
				Results: []sql.Row{},
			},
			{
				Statement: `with test_bytes as (
  select
    inbytes,
    description,
    (test_conv(inbytes || repeat('.', 3)::bytea, 'utf8', 'utf8')).error
  from utf8_verification_inputs
), test_padded as (
  select
    description,
    (test_conv(repeat('.', 64 - length(inbytes))::bytea || inbytes || repeat('.', 3)::bytea, 'utf8', 'utf8')).error
  from test_bytes
)
select
  description,
  b.error as orig_error,
  p.error as error_after_padding
from test_padded p
join test_bytes b
using (description)
where p.error is distinct from b.error
order by description;`,
				Results: []sql.Row{},
			},
			{
				Statement: `CREATE TABLE utf8_inputs (inbytes bytea, description text);`,
			},
			{
				Statement: `insert into utf8_inputs  values
  ('\x666f6f',		'valid, pure ASCII'),
  ('\xc3a4c3b6',	'valid, extra latin chars'),
  ('\xd184d0bed0be',	'valid, cyrillic'),
  ('\x666f6fe8b1a1',	'valid, kanji/Chinese'),
  ('\xe382abe3829a',	'valid, two chars that combine to one in EUC_JIS_2004'),
  ('\xe382ab',		'only first half of combined char in EUC_JIS_2004'),
  ('\xe382abe382',	'incomplete combination when converted EUC_JIS_2004'),
  ('\xecbd94eb81bceba6ac', 'valid, Hangul, Korean'),
  ('\x666f6fefa8aa',	'valid, needs mapping function to convert to GB18030'),
  ('\x66e8b1ff6f6f',	'invalid byte sequence'),
  ('\x66006f',		'invalid, NUL byte'),
  ('\x666f6fe8b100',	'invalid, NUL byte'),
  ('\x666f6fe8b1',	'incomplete character at end');`,
			},
			{
				Statement: `select description, (test_conv(inbytes, 'utf8', 'utf8')).* from utf8_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, ``, ``}, {`valid, extra latin chars`, `\xc3a4c3b6`, ``, ``}, {`valid, cyrillic`, `\xd184d0bed0be`, ``, ``}, {`valid, kanji/Chinese`, `\x666f6fe8b1a1`, ``, ``}, {`valid, two chars that combine to one in EUC_JIS_2004`, `\xe382abe3829a`, ``, ``}, {`only first half of combined char in EUC_JIS_2004`, `\xe382ab`, ``, ``}, {`incomplete combination when converted EUC_JIS_2004`, `\xe382ab`, `\xe382`, `invalid byte sequence for encoding "UTF8": 0xe3 0x82`}, {`valid, Hangul, Korean`, `\xecbd94eb81bceba6ac`, ``, ``}, {`valid, needs mapping function to convert to GB18030`, `\x666f6fefa8aa`, ``, ``}, {`invalid byte sequence`, `\x66`, `\xe8b1ff6f6f`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1 0xff`}, {`invalid, NUL byte`, `\x66`, `\x006f`, `invalid byte sequence for encoding "UTF8": 0x00`}, {`invalid, NUL byte`, `\x666f6f`, `\xe8b100`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1 0x00`}, {`incomplete character at end`, `\x666f6f`, `\xe8b1`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'utf8', 'euc_jis_2004')).* from utf8_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid, extra latin chars`, `\xc3a4c3b6`, `\xa9daa9ec`, ``, ``}, {`valid, cyrillic`, `\xd184d0bed0be`, `\xa7e6a7e0a7e0`, ``, ``}, {`valid, kanji/Chinese`, `\x666f6fe8b1a1`, `\x666f6fbedd`, ``, ``}, {`valid, two chars that combine to one in EUC_JIS_2004`, `\xe382abe3829a`, `\xa5f7`, ``, ``}, {`only first half of combined char in EUC_JIS_2004`, `\xe382ab`, `\xa5ab`, ``, ``}, {`incomplete combination when converted EUC_JIS_2004`, `\xe382abe382`, `\x`, `\xe382abe382`, `invalid byte sequence for encoding "UTF8": 0xe3 0x82`}, {`valid, Hangul, Korean`, `\xecbd94eb81bceba6ac`, `\x`, `\xecbd94eb81bceba6ac`, `character with byte sequence 0xec 0xbd 0x94 in encoding "UTF8" has no equivalent in encoding "EUC_JIS_2004"`}, {`valid, needs mapping function to convert to GB18030`, `\x666f6fefa8aa`, `\x666f6f`, `\xefa8aa`, `character with byte sequence 0xef 0xa8 0xaa in encoding "UTF8" has no equivalent in encoding "EUC_JIS_2004"`}, {`invalid byte sequence`, `\x66e8b1ff6f6f`, `\x66`, `\xe8b1ff6f6f`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1 0xff`}, {`invalid, NUL byte`, `\x66006f`, `\x66`, `\x006f`, `invalid byte sequence for encoding "UTF8": 0x00`}, {`invalid, NUL byte`, `\x666f6fe8b100`, `\x666f6f`, `\xe8b100`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1 0x00`}, {`incomplete character at end`, `\x666f6fe8b1`, `\x666f6f`, `\xe8b1`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'utf8', 'latin1')).* from utf8_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid, extra latin chars`, `\xc3a4c3b6`, `\xe4f6`, ``, ``}, {`valid, cyrillic`, `\xd184d0bed0be`, `\x`, `\xd184d0bed0be`, `character with byte sequence 0xd1 0x84 in encoding "UTF8" has no equivalent in encoding "LATIN1"`}, {`valid, kanji/Chinese`, `\x666f6fe8b1a1`, `\x666f6f`, `\xe8b1a1`, `character with byte sequence 0xe8 0xb1 0xa1 in encoding "UTF8" has no equivalent in encoding "LATIN1"`}, {`valid, two chars that combine to one in EUC_JIS_2004`, `\xe382abe3829a`, `\x`, `\xe382abe3829a`, `character with byte sequence 0xe3 0x82 0xab in encoding "UTF8" has no equivalent in encoding "LATIN1"`}, {`only first half of combined char in EUC_JIS_2004`, `\xe382ab`, `\x`, `\xe382ab`, `character with byte sequence 0xe3 0x82 0xab in encoding "UTF8" has no equivalent in encoding "LATIN1"`}, {`incomplete combination when converted EUC_JIS_2004`, `\xe382abe382`, `\x`, `\xe382abe382`, `character with byte sequence 0xe3 0x82 0xab in encoding "UTF8" has no equivalent in encoding "LATIN1"`}, {`valid, Hangul, Korean`, `\xecbd94eb81bceba6ac`, `\x`, `\xecbd94eb81bceba6ac`, `character with byte sequence 0xec 0xbd 0x94 in encoding "UTF8" has no equivalent in encoding "LATIN1"`}, {`valid, needs mapping function to convert to GB18030`, `\x666f6fefa8aa`, `\x666f6f`, `\xefa8aa`, `character with byte sequence 0xef 0xa8 0xaa in encoding "UTF8" has no equivalent in encoding "LATIN1"`}, {`invalid byte sequence`, `\x66e8b1ff6f6f`, `\x66`, `\xe8b1ff6f6f`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1 0xff`}, {`invalid, NUL byte`, `\x66006f`, `\x66`, `\x006f`, `invalid byte sequence for encoding "UTF8": 0x00`}, {`invalid, NUL byte`, `\x666f6fe8b100`, `\x666f6f`, `\xe8b100`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1 0x00`}, {`incomplete character at end`, `\x666f6fe8b1`, `\x666f6f`, `\xe8b1`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'utf8', 'latin2')).* from utf8_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid, extra latin chars`, `\xc3a4c3b6`, `\xe4f6`, ``, ``}, {`valid, cyrillic`, `\xd184d0bed0be`, `\x`, `\xd184d0bed0be`, `character with byte sequence 0xd1 0x84 in encoding "UTF8" has no equivalent in encoding "LATIN2"`}, {`valid, kanji/Chinese`, `\x666f6fe8b1a1`, `\x666f6f`, `\xe8b1a1`, `character with byte sequence 0xe8 0xb1 0xa1 in encoding "UTF8" has no equivalent in encoding "LATIN2"`}, {`valid, two chars that combine to one in EUC_JIS_2004`, `\xe382abe3829a`, `\x`, `\xe382abe3829a`, `character with byte sequence 0xe3 0x82 0xab in encoding "UTF8" has no equivalent in encoding "LATIN2"`}, {`only first half of combined char in EUC_JIS_2004`, `\xe382ab`, `\x`, `\xe382ab`, `character with byte sequence 0xe3 0x82 0xab in encoding "UTF8" has no equivalent in encoding "LATIN2"`}, {`incomplete combination when converted EUC_JIS_2004`, `\xe382abe382`, `\x`, `\xe382abe382`, `character with byte sequence 0xe3 0x82 0xab in encoding "UTF8" has no equivalent in encoding "LATIN2"`}, {`valid, Hangul, Korean`, `\xecbd94eb81bceba6ac`, `\x`, `\xecbd94eb81bceba6ac`, `character with byte sequence 0xec 0xbd 0x94 in encoding "UTF8" has no equivalent in encoding "LATIN2"`}, {`valid, needs mapping function to convert to GB18030`, `\x666f6fefa8aa`, `\x666f6f`, `\xefa8aa`, `character with byte sequence 0xef 0xa8 0xaa in encoding "UTF8" has no equivalent in encoding "LATIN2"`}, {`invalid byte sequence`, `\x66e8b1ff6f6f`, `\x66`, `\xe8b1ff6f6f`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1 0xff`}, {`invalid, NUL byte`, `\x66006f`, `\x66`, `\x006f`, `invalid byte sequence for encoding "UTF8": 0x00`}, {`invalid, NUL byte`, `\x666f6fe8b100`, `\x666f6f`, `\xe8b100`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1 0x00`}, {`incomplete character at end`, `\x666f6fe8b1`, `\x666f6f`, `\xe8b1`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'utf8', 'latin5')).* from utf8_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid, extra latin chars`, `\xc3a4c3b6`, `\xe4f6`, ``, ``}, {`valid, cyrillic`, `\xd184d0bed0be`, `\x`, `\xd184d0bed0be`, `character with byte sequence 0xd1 0x84 in encoding "UTF8" has no equivalent in encoding "LATIN5"`}, {`valid, kanji/Chinese`, `\x666f6fe8b1a1`, `\x666f6f`, `\xe8b1a1`, `character with byte sequence 0xe8 0xb1 0xa1 in encoding "UTF8" has no equivalent in encoding "LATIN5"`}, {`valid, two chars that combine to one in EUC_JIS_2004`, `\xe382abe3829a`, `\x`, `\xe382abe3829a`, `character with byte sequence 0xe3 0x82 0xab in encoding "UTF8" has no equivalent in encoding "LATIN5"`}, {`only first half of combined char in EUC_JIS_2004`, `\xe382ab`, `\x`, `\xe382ab`, `character with byte sequence 0xe3 0x82 0xab in encoding "UTF8" has no equivalent in encoding "LATIN5"`}, {`incomplete combination when converted EUC_JIS_2004`, `\xe382abe382`, `\x`, `\xe382abe382`, `character with byte sequence 0xe3 0x82 0xab in encoding "UTF8" has no equivalent in encoding "LATIN5"`}, {`valid, Hangul, Korean`, `\xecbd94eb81bceba6ac`, `\x`, `\xecbd94eb81bceba6ac`, `character with byte sequence 0xec 0xbd 0x94 in encoding "UTF8" has no equivalent in encoding "LATIN5"`}, {`valid, needs mapping function to convert to GB18030`, `\x666f6fefa8aa`, `\x666f6f`, `\xefa8aa`, `character with byte sequence 0xef 0xa8 0xaa in encoding "UTF8" has no equivalent in encoding "LATIN5"`}, {`invalid byte sequence`, `\x66e8b1ff6f6f`, `\x66`, `\xe8b1ff6f6f`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1 0xff`}, {`invalid, NUL byte`, `\x66006f`, `\x66`, `\x006f`, `invalid byte sequence for encoding "UTF8": 0x00`}, {`invalid, NUL byte`, `\x666f6fe8b100`, `\x666f6f`, `\xe8b100`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1 0x00`}, {`incomplete character at end`, `\x666f6fe8b1`, `\x666f6f`, `\xe8b1`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'utf8', 'koi8r')).* from utf8_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid, extra latin chars`, `\xc3a4c3b6`, `\x`, `\xc3a4c3b6`, `character with byte sequence 0xc3 0xa4 in encoding "UTF8" has no equivalent in encoding "KOI8R"`}, {`valid, cyrillic`, `\xd184d0bed0be`, `\xc6cfcf`, ``, ``}, {`valid, kanji/Chinese`, `\x666f6fe8b1a1`, `\x666f6f`, `\xe8b1a1`, `character with byte sequence 0xe8 0xb1 0xa1 in encoding "UTF8" has no equivalent in encoding "KOI8R"`}, {`valid, two chars that combine to one in EUC_JIS_2004`, `\xe382abe3829a`, `\x`, `\xe382abe3829a`, `character with byte sequence 0xe3 0x82 0xab in encoding "UTF8" has no equivalent in encoding "KOI8R"`}, {`only first half of combined char in EUC_JIS_2004`, `\xe382ab`, `\x`, `\xe382ab`, `character with byte sequence 0xe3 0x82 0xab in encoding "UTF8" has no equivalent in encoding "KOI8R"`}, {`incomplete combination when converted EUC_JIS_2004`, `\xe382abe382`, `\x`, `\xe382abe382`, `character with byte sequence 0xe3 0x82 0xab in encoding "UTF8" has no equivalent in encoding "KOI8R"`}, {`valid, Hangul, Korean`, `\xecbd94eb81bceba6ac`, `\x`, `\xecbd94eb81bceba6ac`, `character with byte sequence 0xec 0xbd 0x94 in encoding "UTF8" has no equivalent in encoding "KOI8R"`}, {`valid, needs mapping function to convert to GB18030`, `\x666f6fefa8aa`, `\x666f6f`, `\xefa8aa`, `character with byte sequence 0xef 0xa8 0xaa in encoding "UTF8" has no equivalent in encoding "KOI8R"`}, {`invalid byte sequence`, `\x66e8b1ff6f6f`, `\x66`, `\xe8b1ff6f6f`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1 0xff`}, {`invalid, NUL byte`, `\x66006f`, `\x66`, `\x006f`, `invalid byte sequence for encoding "UTF8": 0x00`}, {`invalid, NUL byte`, `\x666f6fe8b100`, `\x666f6f`, `\xe8b100`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1 0x00`}, {`incomplete character at end`, `\x666f6fe8b1`, `\x666f6f`, `\xe8b1`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'utf8', 'gb18030')).* from utf8_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid, extra latin chars`, `\xc3a4c3b6`, `\x81308a3181308b32`, ``, ``}, {`valid, cyrillic`, `\xd184d0bed0be`, `\xa7e6a7e0a7e0`, ``, ``}, {`valid, kanji/Chinese`, `\x666f6fe8b1a1`, `\x666f6fcff3`, ``, ``}, {`valid, two chars that combine to one in EUC_JIS_2004`, `\xe382abe3829a`, `\xa5ab8139a732`, ``, ``}, {`only first half of combined char in EUC_JIS_2004`, `\xe382ab`, `\xa5ab`, ``, ``}, {`incomplete combination when converted EUC_JIS_2004`, `\xe382abe382`, `\xa5ab`, `\xe382`, `invalid byte sequence for encoding "UTF8": 0xe3 0x82`}, {`valid, Hangul, Korean`, `\xecbd94eb81bceba6ac`, `\x8334e5398238c4338330b335`, ``, ``}, {`valid, needs mapping function to convert to GB18030`, `\x666f6fefa8aa`, `\x666f6f84309c38`, ``, ``}, {`invalid byte sequence`, `\x66e8b1ff6f6f`, `\x66`, `\xe8b1ff6f6f`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1 0xff`}, {`invalid, NUL byte`, `\x66006f`, `\x66`, `\x006f`, `invalid byte sequence for encoding "UTF8": 0x00`}, {`invalid, NUL byte`, `\x666f6fe8b100`, `\x666f6f`, `\xe8b100`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1 0x00`}, {`incomplete character at end`, `\x666f6fe8b1`, `\x666f6f`, `\xe8b1`, `invalid byte sequence for encoding "UTF8": 0xe8 0xb1`}},
			},
			{
				Statement: `CREATE TABLE euc_jis_2004_inputs (inbytes bytea, description text);`,
			},
			{
				Statement: `insert into euc_jis_2004_inputs  values
  ('\x666f6f',		'valid, pure ASCII'),
  ('\x666f6fbedd',	'valid'),
  ('\xa5f7',		'valid, translates to two UTF-8 chars '),
  ('\xbeddbe',		'incomplete char '),
  ('\x666f6f00bedd',	'invalid, NUL byte'),
  ('\x666f6fbe00dd',	'invalid, NUL byte'),
  ('\x666f6fbedd00',	'invalid, NUL byte'),
  ('\xbe04',		'invalid byte sequence');`,
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'euc_jis_2004', 'euc_jis_2004')).* from euc_jis_2004_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid`, `\x666f6fbedd`, `\x666f6fbedd`, ``, ``}, {`valid, translates to two UTF-8 chars`, `\xa5f7`, `\xa5f7`, ``, ``}, {`incomplete char`, `\xbeddbe`, `\xbedd`, `\xbe`, `invalid byte sequence for encoding "EUC_JIS_2004": 0xbe`}, {`invalid, NUL byte`, `\x666f6f00bedd`, `\x666f6f`, `\x00bedd`, `invalid byte sequence for encoding "EUC_JIS_2004": 0x00`}, {`invalid, NUL byte`, `\x666f6fbe00dd`, `\x666f6f`, `\xbe00dd`, `invalid byte sequence for encoding "EUC_JIS_2004": 0xbe 0x00`}, {`invalid, NUL byte`, `\x666f6fbedd00`, `\x666f6fbedd`, `\x00`, `invalid byte sequence for encoding "EUC_JIS_2004": 0x00`}, {`invalid byte sequence`, `\xbe04`, `\x`, `\xbe04`, `invalid byte sequence for encoding "EUC_JIS_2004": 0xbe 0x04`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'euc_jis_2004', 'utf8')).* from euc_jis_2004_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid`, `\x666f6fbedd`, `\x666f6fe8b1a1`, ``, ``}, {`valid, translates to two UTF-8 chars`, `\xa5f7`, `\xe382abe3829a`, ``, ``}, {`incomplete char`, `\xbeddbe`, `\xe8b1a1`, `\xbe`, `invalid byte sequence for encoding "EUC_JIS_2004": 0xbe`}, {`invalid, NUL byte`, `\x666f6f00bedd`, `\x666f6f`, `\x00bedd`, `invalid byte sequence for encoding "EUC_JIS_2004": 0x00`}, {`invalid, NUL byte`, `\x666f6fbe00dd`, `\x666f6f`, `\xbe00dd`, `invalid byte sequence for encoding "EUC_JIS_2004": 0xbe 0x00`}, {`invalid, NUL byte`, `\x666f6fbedd00`, `\x666f6fe8b1a1`, `\x00`, `invalid byte sequence for encoding "EUC_JIS_2004": 0x00`}, {`invalid byte sequence`, `\xbe04`, `\x`, `\xbe04`, `invalid byte sequence for encoding "EUC_JIS_2004": 0xbe 0x04`}},
			},
			{
				Statement: `CREATE TABLE shiftjis2004_inputs (inbytes bytea, description text);`,
			},
			{
				Statement: `insert into shiftjis2004_inputs  values
  ('\x666f6f',		'valid, pure ASCII'),
  ('\x666f6f8fdb',	'valid'),
  ('\x666f6f81c0',	'valid, no translation to UTF-8'),
  ('\x666f6f82f5',	'valid, translates to two UTF-8 chars '),
  ('\x666f6f8fdb8f',	'incomplete char '),
  ('\x666f6f820a',	'incomplete char, followed by newline '),
  ('\x666f6f008fdb',	'invalid, NUL byte'),
  ('\x666f6f8f00db',	'invalid, NUL byte'),
  ('\x666f6f8fdb00',	'invalid, NUL byte');`,
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'shiftjis2004', 'shiftjis2004')).* from shiftjis2004_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid`, `\x666f6f8fdb`, `\x666f6f8fdb`, ``, ``}, {`valid, no translation to UTF-8`, `\x666f6f81c0`, `\x666f6f81c0`, ``, ``}, {`valid, translates to two UTF-8 chars`, `\x666f6f82f5`, `\x666f6f82f5`, ``, ``}, {`incomplete char`, `\x666f6f8fdb8f`, `\x666f6f8fdb`, `\x8f`, `invalid byte sequence for encoding "SHIFT_JIS_2004": 0x8f`}, {`incomplete char, followed by newline`, `\x666f6f820a`, `\x666f6f`, `\x820a`, `invalid byte sequence for encoding "SHIFT_JIS_2004": 0x82 0x0a`}, {`invalid, NUL byte`, `\x666f6f008fdb`, `\x666f6f`, `\x008fdb`, `invalid byte sequence for encoding "SHIFT_JIS_2004": 0x00`}, {`invalid, NUL byte`, `\x666f6f8f00db`, `\x666f6f`, `\x8f00db`, `invalid byte sequence for encoding "SHIFT_JIS_2004": 0x8f 0x00`}, {`invalid, NUL byte`, `\x666f6f8fdb00`, `\x666f6f8fdb`, `\x00`, `invalid byte sequence for encoding "SHIFT_JIS_2004": 0x00`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'shiftjis2004', 'utf8')).* from shiftjis2004_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid`, `\x666f6f8fdb`, `\x666f6fe8b1a1`, ``, ``}, {`valid, no translation to UTF-8`, `\x666f6f81c0`, `\x666f6fe28a84`, ``, ``}, {`valid, translates to two UTF-8 chars`, `\x666f6f82f5`, `\x666f6fe3818be3829a`, ``, ``}, {`incomplete char`, `\x666f6f8fdb8f`, `\x666f6fe8b1a1`, `\x8f`, `invalid byte sequence for encoding "SHIFT_JIS_2004": 0x8f`}, {`incomplete char, followed by newline`, `\x666f6f820a`, `\x666f6f`, `\x820a`, `invalid byte sequence for encoding "SHIFT_JIS_2004": 0x82 0x0a`}, {`invalid, NUL byte`, `\x666f6f008fdb`, `\x666f6f`, `\x008fdb`, `invalid byte sequence for encoding "SHIFT_JIS_2004": 0x00`}, {`invalid, NUL byte`, `\x666f6f8f00db`, `\x666f6f`, `\x8f00db`, `invalid byte sequence for encoding "SHIFT_JIS_2004": 0x8f 0x00`}, {`invalid, NUL byte`, `\x666f6f8fdb00`, `\x666f6fe8b1a1`, `\x00`, `invalid byte sequence for encoding "SHIFT_JIS_2004": 0x00`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'shiftjis2004', 'euc_jis_2004')).* from shiftjis2004_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid`, `\x666f6f8fdb`, `\x666f6fbedd`, ``, ``}, {`valid, no translation to UTF-8`, `\x666f6f81c0`, `\x666f6fa2c2`, ``, ``}, {`valid, translates to two UTF-8 chars`, `\x666f6f82f5`, `\x666f6fa4f7`, ``, ``}, {`incomplete char`, `\x666f6f8fdb8f`, `\x666f6fbedd`, `\x8f`, `invalid byte sequence for encoding "SHIFT_JIS_2004": 0x8f`}, {`incomplete char, followed by newline`, `\x666f6f820a`, `\x666f6f`, `\x820a`, `invalid byte sequence for encoding "SHIFT_JIS_2004": 0x82 0x0a`}, {`invalid, NUL byte`, `\x666f6f008fdb`, `\x666f6f`, `\x008fdb`, `invalid byte sequence for encoding "SHIFT_JIS_2004": 0x00`}, {`invalid, NUL byte`, `\x666f6f8f00db`, `\x666f6f`, `\x8f00db`, `invalid byte sequence for encoding "SHIFT_JIS_2004": 0x8f 0x00`}, {`invalid, NUL byte`, `\x666f6f8fdb00`, `\x666f6fbedd`, `\x00`, `invalid byte sequence for encoding "SHIFT_JIS_2004": 0x00`}},
			},
			{
				Statement: `CREATE TABLE gb18030_inputs (inbytes bytea, description text);`,
			},
			{
				Statement: `insert into gb18030_inputs  values
  ('\x666f6f',		'valid, pure ASCII'),
  ('\x666f6fcff3',	'valid'),
  ('\x666f6f8431a530',	'valid, no translation to UTF-8'),
  ('\x666f6f84309c38',	'valid, translates to UTF-8 by mapping function'),
  ('\x666f6f84309c',	'incomplete char '),
  ('\x666f6f84309c0a',	'incomplete char, followed by newline '),
  ('\x666f6f84309c3800', 'invalid, NUL byte'),
  ('\x666f6f84309c0038', 'invalid, NUL byte');`,
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'gb18030', 'gb18030')).* from gb18030_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid`, `\x666f6fcff3`, `\x666f6fcff3`, ``, ``}, {`valid, no translation to UTF-8`, `\x666f6f8431a530`, `\x666f6f8431a530`, ``, ``}, {`valid, translates to UTF-8 by mapping function`, `\x666f6f84309c38`, `\x666f6f84309c38`, ``, ``}, {`incomplete char`, `\x666f6f84309c`, `\x666f6f`, `\x84309c`, `invalid byte sequence for encoding "GB18030": 0x84 0x30 0x9c`}, {`incomplete char, followed by newline`, `\x666f6f84309c0a`, `\x666f6f`, `\x84309c0a`, `invalid byte sequence for encoding "GB18030": 0x84 0x30 0x9c 0x0a`}, {`invalid, NUL byte`, `\x666f6f84309c3800`, `\x666f6f84309c38`, `\x00`, `invalid byte sequence for encoding "GB18030": 0x00`}, {`invalid, NUL byte`, `\x666f6f84309c0038`, `\x666f6f`, `\x84309c0038`, `invalid byte sequence for encoding "GB18030": 0x84 0x30 0x9c 0x00`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'gb18030', 'utf8')).* from gb18030_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid`, `\x666f6fcff3`, `\x666f6fe8b1a1`, ``, ``}, {`valid, no translation to UTF-8`, `\x666f6f8431a530`, `\x666f6f`, `\x8431a530`, `character with byte sequence 0x84 0x31 0xa5 0x30 in encoding "GB18030" has no equivalent in encoding "UTF8"`}, {`valid, translates to UTF-8 by mapping function`, `\x666f6f84309c38`, `\x666f6fefa8aa`, ``, ``}, {`incomplete char`, `\x666f6f84309c`, `\x666f6f`, `\x84309c`, `invalid byte sequence for encoding "GB18030": 0x84 0x30 0x9c`}, {`incomplete char, followed by newline`, `\x666f6f84309c0a`, `\x666f6f`, `\x84309c0a`, `invalid byte sequence for encoding "GB18030": 0x84 0x30 0x9c 0x0a`}, {`invalid, NUL byte`, `\x666f6f84309c3800`, `\x666f6fefa8aa`, `\x00`, `invalid byte sequence for encoding "GB18030": 0x00`}, {`invalid, NUL byte`, `\x666f6f84309c0038`, `\x666f6f`, `\x84309c0038`, `invalid byte sequence for encoding "GB18030": 0x84 0x30 0x9c 0x00`}},
			},
			{
				Statement: `CREATE TABLE iso8859_5_inputs (inbytes bytea, description text);`,
			},
			{
				Statement: `insert into iso8859_5_inputs  values
  ('\x666f6f',		'valid, pure ASCII'),
  ('\xe4dede',		'valid'),
  ('\x00',		'invalid, NUL byte'),
  ('\xe400dede',	'invalid, NUL byte'),
  ('\xe4dede00',	'invalid, NUL byte');`,
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'iso8859-5', 'iso8859-5')).* from iso8859_5_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid`, `\xe4dede`, `\xe4dede`, ``, ``}, {`invalid, NUL byte`, `\x00`, `\x`, `\x00`, `invalid byte sequence for encoding "ISO_8859_5": 0x00`}, {`invalid, NUL byte`, `\xe400dede`, `\xe4`, `\x00dede`, `invalid byte sequence for encoding "ISO_8859_5": 0x00`}, {`invalid, NUL byte`, `\xe4dede00`, `\xe4dede`, `\x00`, `invalid byte sequence for encoding "ISO_8859_5": 0x00`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'iso8859-5', 'utf8')).* from iso8859_5_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid`, `\xe4dede`, `\xd184d0bed0be`, ``, ``}, {`invalid, NUL byte`, `\x00`, `\x`, `\x00`, `invalid byte sequence for encoding "ISO_8859_5": 0x00`}, {`invalid, NUL byte`, `\xe400dede`, `\xd184`, `\x00dede`, `invalid byte sequence for encoding "ISO_8859_5": 0x00`}, {`invalid, NUL byte`, `\xe4dede00`, `\xd184d0bed0be`, `\x00`, `invalid byte sequence for encoding "ISO_8859_5": 0x00`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'iso8859-5', 'koi8r')).* from iso8859_5_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid`, `\xe4dede`, `\xc6cfcf`, ``, ``}, {`invalid, NUL byte`, `\x00`, `\x`, `\x00`, `invalid byte sequence for encoding "ISO_8859_5": 0x00`}, {`invalid, NUL byte`, `\xe400dede`, `\xc6`, `\x00dede`, `invalid byte sequence for encoding "ISO_8859_5": 0x00`}, {`invalid, NUL byte`, `\xe4dede00`, `\xc6cfcf`, `\x00`, `invalid byte sequence for encoding "ISO_8859_5": 0x00`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'iso8859_5', 'mule_internal')).* from iso8859_5_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid`, `\xe4dede`, `\x8bc68bcf8bcf`, ``, ``}, {`invalid, NUL byte`, `\x00`, `\x`, `\x00`, `invalid byte sequence for encoding "ISO_8859_5": 0x00`}, {`invalid, NUL byte`, `\xe400dede`, `\x8bc6`, `\x00dede`, `invalid byte sequence for encoding "ISO_8859_5": 0x00`}, {`invalid, NUL byte`, `\xe4dede00`, `\x8bc68bcf8bcf`, `\x00`, `invalid byte sequence for encoding "ISO_8859_5": 0x00`}},
			},
			{
				Statement: `CREATE TABLE big5_inputs (inbytes bytea, description text);`,
			},
			{
				Statement: `insert into big5_inputs  values
  ('\x666f6f',		'valid, pure ASCII'),
  ('\x666f6fb648',	'valid'),
  ('\x666f6fa27f',	'valid, no translation to UTF-8'),
  ('\x666f6fb60048',	'invalid, NUL byte'),
  ('\x666f6fb64800',	'invalid, NUL byte');`,
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'big5', 'big5')).* from big5_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid`, `\x666f6fb648`, `\x666f6fb648`, ``, ``}, {`valid, no translation to UTF-8`, `\x666f6fa27f`, `\x666f6fa27f`, ``, ``}, {`invalid, NUL byte`, `\x666f6fb60048`, `\x666f6f`, `\xb60048`, `invalid byte sequence for encoding "BIG5": 0xb6 0x00`}, {`invalid, NUL byte`, `\x666f6fb64800`, `\x666f6fb648`, `\x00`, `invalid byte sequence for encoding "BIG5": 0x00`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'big5', 'utf8')).* from big5_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid`, `\x666f6fb648`, `\x666f6fe8b1a1`, ``, ``}, {`valid, no translation to UTF-8`, `\x666f6fa27f`, `\x666f6f`, `\xa27f`, `character with byte sequence 0xa2 0x7f in encoding "BIG5" has no equivalent in encoding "UTF8"`}, {`invalid, NUL byte`, `\x666f6fb60048`, `\x666f6f`, `\xb60048`, `invalid byte sequence for encoding "BIG5": 0xb6 0x00`}, {`invalid, NUL byte`, `\x666f6fb64800`, `\x666f6fe8b1a1`, `\x00`, `invalid byte sequence for encoding "BIG5": 0x00`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'big5', 'mule_internal')).* from big5_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid`, `\x666f6fb648`, `\x666f6f95e2af`, ``, ``}, {`valid, no translation to UTF-8`, `\x666f6fa27f`, `\x666f6f95a3c1`, ``, ``}, {`invalid, NUL byte`, `\x666f6fb60048`, `\x666f6f`, `\xb60048`, `invalid byte sequence for encoding "BIG5": 0xb6 0x00`}, {`invalid, NUL byte`, `\x666f6fb64800`, `\x666f6f95e2af`, `\x00`, `invalid byte sequence for encoding "BIG5": 0x00`}},
			},
			{
				Statement: `CREATE TABLE mic_inputs (inbytes bytea, description text);`,
			},
			{
				Statement: `insert into mic_inputs  values
  ('\x666f6f',		'valid, pure ASCII'),
  ('\x8bc68bcf8bcf',	'valid (in KOI8R)'),
  ('\x8bc68bcf8b',	'invalid,incomplete char'),
  ('\x92bedd',		'valid (in SHIFT_JIS)'),
  ('\x92be',		'invalid, incomplete char)'),
  ('\x666f6f95a3c1',	'valid (in Big5)'),
  ('\x666f6f95a3',	'invalid, incomplete char'),
  ('\x9200bedd',	'invalid, NUL byte'),
  ('\x92bedd00',	'invalid, NUL byte'),
  ('\x8b00c68bcf8bcf',	'invalid, NUL byte');`,
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'mule_internal', 'mule_internal')).* from mic_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid (in KOI8R)`, `\x8bc68bcf8bcf`, `\x8bc68bcf8bcf`, ``, ``}, {`invalid,incomplete char`, `\x8bc68bcf8b`, `\x8bc68bcf`, `\x8b`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x8b`}, {`valid (in SHIFT_JIS)`, `\x92bedd`, `\x92bedd`, ``, ``}, {`invalid, incomplete char)`, `\x92be`, `\x`, `\x92be`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x92 0xbe`}, {`valid (in Big5)`, `\x666f6f95a3c1`, `\x666f6f95a3c1`, ``, ``}, {`invalid, incomplete char`, `\x666f6f95a3`, `\x666f6f`, `\x95a3`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x95 0xa3`}, {`invalid, NUL byte`, `\x9200bedd`, `\x`, `\x9200bedd`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x92 0x00 0xbe`}, {`invalid, NUL byte`, `\x92bedd00`, `\x92bedd`, `\x00`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x00`}, {`invalid, NUL byte`, `\x8b00c68bcf8bcf`, `\x`, `\x8b00c68bcf8bcf`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x8b 0x00`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'mule_internal', 'koi8r')).* from mic_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid (in KOI8R)`, `\x8bc68bcf8bcf`, `\xc6cfcf`, ``, ``}, {`invalid,incomplete char`, `\x8bc68bcf8b`, `\xc6cf`, `\x8b`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x8b`}, {`valid (in SHIFT_JIS)`, `\x92bedd`, `\x`, `\x92bedd`, `character with byte sequence 0x92 0xbe 0xdd in encoding "MULE_INTERNAL" has no equivalent in encoding "KOI8R"`}, {`invalid, incomplete char)`, `\x92be`, `\x`, `\x92be`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x92 0xbe`}, {`valid (in Big5)`, `\x666f6f95a3c1`, `\x666f6f`, `\x95a3c1`, `character with byte sequence 0x95 0xa3 0xc1 in encoding "MULE_INTERNAL" has no equivalent in encoding "KOI8R"`}, {`invalid, incomplete char`, `\x666f6f95a3`, `\x666f6f`, `\x95a3`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x95 0xa3`}, {`invalid, NUL byte`, `\x9200bedd`, `\x`, `\x9200bedd`, `character with byte sequence 0x92 0x00 0xbe in encoding "MULE_INTERNAL" has no equivalent in encoding "KOI8R"`}, {`invalid, NUL byte`, `\x92bedd00`, `\x`, `\x92bedd00`, `character with byte sequence 0x92 0xbe 0xdd in encoding "MULE_INTERNAL" has no equivalent in encoding "KOI8R"`}, {`invalid, NUL byte`, `\x8b00c68bcf8bcf`, `\x`, `\x8b00c68bcf8bcf`, `character with byte sequence 0x8b 0x00 in encoding "MULE_INTERNAL" has no equivalent in encoding "KOI8R"`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'mule_internal', 'iso8859-5')).* from mic_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid (in KOI8R)`, `\x8bc68bcf8bcf`, `\xe4dede`, ``, ``}, {`invalid,incomplete char`, `\x8bc68bcf8b`, `\xe4de`, `\x8b`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x8b`}, {`valid (in SHIFT_JIS)`, `\x92bedd`, `\x`, `\x92bedd`, `character with byte sequence 0x92 0xbe 0xdd in encoding "MULE_INTERNAL" has no equivalent in encoding "ISO_8859_5"`}, {`invalid, incomplete char)`, `\x92be`, `\x`, `\x92be`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x92 0xbe`}, {`valid (in Big5)`, `\x666f6f95a3c1`, `\x666f6f`, `\x95a3c1`, `character with byte sequence 0x95 0xa3 0xc1 in encoding "MULE_INTERNAL" has no equivalent in encoding "ISO_8859_5"`}, {`invalid, incomplete char`, `\x666f6f95a3`, `\x666f6f`, `\x95a3`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x95 0xa3`}, {`invalid, NUL byte`, `\x9200bedd`, `\x`, `\x9200bedd`, `character with byte sequence 0x92 0x00 0xbe in encoding "MULE_INTERNAL" has no equivalent in encoding "ISO_8859_5"`}, {`invalid, NUL byte`, `\x92bedd00`, `\x`, `\x92bedd00`, `character with byte sequence 0x92 0xbe 0xdd in encoding "MULE_INTERNAL" has no equivalent in encoding "ISO_8859_5"`}, {`invalid, NUL byte`, `\x8b00c68bcf8bcf`, `\x`, `\x8b00c68bcf8bcf`, `character with byte sequence 0x8b 0x00 in encoding "MULE_INTERNAL" has no equivalent in encoding "ISO_8859_5"`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'mule_internal', 'sjis')).* from mic_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid (in KOI8R)`, `\x8bc68bcf8bcf`, `\x`, `\x8bc68bcf8bcf`, `character with byte sequence 0x8b 0xc6 in encoding "MULE_INTERNAL" has no equivalent in encoding "SJIS"`}, {`invalid,incomplete char`, `\x8bc68bcf8b`, `\x`, `\x8bc68bcf8b`, `character with byte sequence 0x8b 0xc6 in encoding "MULE_INTERNAL" has no equivalent in encoding "SJIS"`}, {`valid (in SHIFT_JIS)`, `\x92bedd`, `\x8fdb`, ``, ``}, {`invalid, incomplete char)`, `\x92be`, `\x`, `\x92be`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x92 0xbe`}, {`valid (in Big5)`, `\x666f6f95a3c1`, `\x666f6f`, `\x95a3c1`, `character with byte sequence 0x95 0xa3 0xc1 in encoding "MULE_INTERNAL" has no equivalent in encoding "SJIS"`}, {`invalid, incomplete char`, `\x666f6f95a3`, `\x666f6f`, `\x95a3`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x95 0xa3`}, {`invalid, NUL byte`, `\x9200bedd`, `\x`, `\x9200bedd`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x92 0x00 0xbe`}, {`invalid, NUL byte`, `\x92bedd00`, `\x8fdb`, `\x00`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x00`}, {`invalid, NUL byte`, `\x8b00c68bcf8bcf`, `\x`, `\x8b00c68bcf8bcf`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x8b 0x00`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'mule_internal', 'big5')).* from mic_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid (in KOI8R)`, `\x8bc68bcf8bcf`, `\x`, `\x8bc68bcf8bcf`, `character with byte sequence 0x8b 0xc6 in encoding "MULE_INTERNAL" has no equivalent in encoding "BIG5"`}, {`invalid,incomplete char`, `\x8bc68bcf8b`, `\x`, `\x8bc68bcf8b`, `character with byte sequence 0x8b 0xc6 in encoding "MULE_INTERNAL" has no equivalent in encoding "BIG5"`}, {`valid (in SHIFT_JIS)`, `\x92bedd`, `\x`, `\x92bedd`, `character with byte sequence 0x92 0xbe 0xdd in encoding "MULE_INTERNAL" has no equivalent in encoding "BIG5"`}, {`invalid, incomplete char)`, `\x92be`, `\x`, `\x92be`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x92 0xbe`}, {`valid (in Big5)`, `\x666f6f95a3c1`, `\x666f6fa2a1`, ``, ``}, {`invalid, incomplete char`, `\x666f6f95a3`, `\x666f6f`, `\x95a3`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x95 0xa3`}, {`invalid, NUL byte`, `\x9200bedd`, `\x`, `\x9200bedd`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x92 0x00 0xbe`}, {`invalid, NUL byte`, `\x92bedd00`, `\x`, `\x92bedd00`, `character with byte sequence 0x92 0xbe 0xdd in encoding "MULE_INTERNAL" has no equivalent in encoding "BIG5"`}, {`invalid, NUL byte`, `\x8b00c68bcf8bcf`, `\x`, `\x8b00c68bcf8bcf`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x8b 0x00`}},
			},
			{
				Statement: `select description, inbytes, (test_conv(inbytes, 'mule_internal', 'euc_jp')).* from mic_inputs;`,
				Results:   []sql.Row{{`valid, pure ASCII`, `\x666f6f`, `\x666f6f`, ``, ``}, {`valid (in KOI8R)`, `\x8bc68bcf8bcf`, `\x`, `\x8bc68bcf8bcf`, `character with byte sequence 0x8b 0xc6 in encoding "MULE_INTERNAL" has no equivalent in encoding "EUC_JP"`}, {`invalid,incomplete char`, `\x8bc68bcf8b`, `\x`, `\x8bc68bcf8b`, `character with byte sequence 0x8b 0xc6 in encoding "MULE_INTERNAL" has no equivalent in encoding "EUC_JP"`}, {`valid (in SHIFT_JIS)`, `\x92bedd`, `\xbedd`, ``, ``}, {`invalid, incomplete char)`, `\x92be`, `\x`, `\x92be`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x92 0xbe`}, {`valid (in Big5)`, `\x666f6f95a3c1`, `\x666f6f`, `\x95a3c1`, `character with byte sequence 0x95 0xa3 0xc1 in encoding "MULE_INTERNAL" has no equivalent in encoding "EUC_JP"`}, {`invalid, incomplete char`, `\x666f6f95a3`, `\x666f6f`, `\x95a3`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x95 0xa3`}, {`invalid, NUL byte`, `\x9200bedd`, `\x`, `\x9200bedd`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x92 0x00 0xbe`}, {`invalid, NUL byte`, `\x92bedd00`, `\xbedd`, `\x00`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x00`}, {`invalid, NUL byte`, `\x8b00c68bcf8bcf`, `\x`, `\x8b00c68bcf8bcf`, `invalid byte sequence for encoding "MULE_INTERNAL": 0x8b 0x00`}},
			},
		},
	})
}
