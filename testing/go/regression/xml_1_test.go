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

func TestXml1(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_xml_1)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_xml_1,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE xmltest (
    id int,
    data xml
);`,
			},
			{
				Statement:   `INSERT INTO xmltest VALUES (1, '<value>one</value>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `INSERT INTO xmltest VALUES (2, '<value>two</value>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `INSERT INTO xmltest VALUES (3, '<wrong');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SELECT * FROM xmltest;`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `SELECT xmlcomment('test');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlcomment('-test');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlcomment('test-');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlcomment('--test');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlcomment('te st');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SELECT xmlconcat(xmlcomment('hello'),
                 xmlelement(NAME qux, 'foo'),
                 xmlcomment('world'));`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlconcat('hello', 'you');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlconcat(1, 2);`,
				ErrorString: `argument of XMLCONCAT must be type xml, not type integer`,
			},
			{
				Statement:   `SELECT xmlconcat('bad', '<syntax');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlconcat('<foo/>', NULL, '<?xml version="1.1" standalone="no"?><bar/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlconcat('<?xml version="1.1"?><foo/>', NULL, '<?xml version="1.1" standalone="no"?><bar/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SELECT xmlconcat(NULL);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT xmlconcat(NULL, NULL);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT xmlelement(name element,
                  xmlattributes (1 as one, 'deuce' as two),
                  'content');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SELECT xmlelement(name element,
                  xmlattributes ('unnamed and wrong'));`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlelement(name element, xmlelement(name nested, 'stuff'));`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlelement(name employee, xmlforest(name, age, salary as pay)) FROM emp;`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlelement(name duplicate, xmlattributes(1 as a, 2 as b, 3 as a));`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlelement(name num, 37);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlelement(name foo, text 'bar');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlelement(name foo, xml 'bar');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlelement(name foo, text 'b<a/>r');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlelement(name foo, xml 'b<a/>r');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlelement(name foo, array[1, 2, 3]);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SET xmlbinary TO base64;`,
			},
			{
				Statement:   `SELECT xmlelement(name foo, bytea 'bar');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SET xmlbinary TO hex;`,
			},
			{
				Statement:   `SELECT xmlelement(name foo, bytea 'bar');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlelement(name foo, xmlattributes(true as bar));`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlelement(name foo, xmlattributes('2009-04-09 00:24:37'::timestamp as bar));`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlelement(name foo, xmlattributes('infinity'::timestamp as bar));`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlelement(name foo, xmlattributes('<>&"''' as funny, xml 'b<a/>r' as funnier));`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(content '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(content '  ');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(content 'abc');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(content '<abc>x</abc>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(content '<invalidentity>&</invalidentity>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(content '<undefinedentity>&idontexist;</undefinedentity>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(content '<invalidns xmlns=''&lt;''/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(content '<relativens xmlns=''relative''/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(content '<twoerrors>&idontexist;</unbalanced>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(content '<nosuchprefix:tag/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(document '   ');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(document 'abc');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(document '<abc>x</abc>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(document '<invalidentity>&</abc>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(document '<undefinedentity>&idontexist;</abc>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(document '<invalidns xmlns=''&lt;''/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(document '<relativens xmlns=''relative''/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(document '<twoerrors>&idontexist;</unbalanced>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlparse(document '<nosuchprefix:tag/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlpi(name foo);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlpi(name xml);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlpi(name xmlstuff);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlpi(name foo, 'bar');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlpi(name foo, 'in?>valid');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlpi(name foo, null);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlpi(name xml, null);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlpi(name xmlstuff, null);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlpi(name "xml-stylesheet", 'href="mystyle.css" type="text/css"');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlpi(name foo, '   bar');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlroot(xml '<foo/>', version no value, standalone no value);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlroot(xml '<foo/>', version '2.0');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlroot(xml '<foo/>', version no value, standalone yes);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlroot(xml '<?xml version="1.1"?><foo/>', version no value, standalone yes);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlroot(xmlroot(xml '<foo/>', version '1.0'), version '1.1', standalone no);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlroot('<?xml version="1.1" standalone="yes"?><foo/>', version no value, standalone no);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlroot('<?xml version="1.1" standalone="yes"?><foo/>', version no value, standalone no value);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlroot('<?xml version="1.1" standalone="yes"?><foo/>', version no value);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SELECT xmlroot (
  xmlelement (
    name gazonk,
    xmlattributes (
      'val' AS name,
      1 + 1 AS num
    ),
    xmlelement (
      NAME qux,
      'foo'
    )
  ),
  version '1.0',
  standalone yes
);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SELECT xmlserialize(content data as character varying(20)) FROM xmltest;`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `SELECT xmlserialize(content 'good' as char(10));`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlserialize(document 'bad' as text);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml '<foo>bar</foo>' IS DOCUMENT;`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml '<foo>bar</foo><bar>foo</bar>' IS DOCUMENT;`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml '<abc/>' IS NOT DOCUMENT;`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml 'abc' IS NOT DOCUMENT;`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT '<>' IS NOT DOCUMENT;`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SELECT xmlagg(data) FROM xmltest;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT xmlagg(data) FROM xmltest WHERE id > 10;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `SELECT xmlelement(name employees, xmlagg(xmlelement(name name, name))) FROM emp;`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlpi(name ":::_xml_abc135.%-&_");`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlpi(name "123");`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `PREPARE foo (xml) AS SELECT xmlconcat('<foo/>', $1);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SET XML OPTION DOCUMENT;`,
			},
			{
				Statement:   `EXECUTE foo ('<bar/>');`,
				ErrorString: `prepared statement "foo" does not exist`,
			},
			{
				Statement:   `EXECUTE foo ('bad');`,
				ErrorString: `prepared statement "foo" does not exist`,
			},
			{
				Statement:   `SELECT xml '<!DOCTYPE a><a/><b/>';`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SET XML OPTION CONTENT;`,
			},
			{
				Statement:   `EXECUTE foo ('<bar/>');`,
				ErrorString: `prepared statement "foo" does not exist`,
			},
			{
				Statement:   `EXECUTE foo ('good');`,
				ErrorString: `prepared statement "foo" does not exist`,
			},
			{
				Statement:   `SELECT xml '<!-- in SQL:2006+ a doc is content too--> <?y z?> <!DOCTYPE a><a/>';`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml '<?xml version="1.0"?> <!-- hi--> <!DOCTYPE a><a/>';`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml '<!DOCTYPE a><a/>';`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml '<!-- hi--> oops <!DOCTYPE a><a/>';`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml '<!-- hi--> <oops/> <!DOCTYPE a><a/>';`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml '<!DOCTYPE a><a/><b/>';`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `CREATE VIEW xmlview1 AS SELECT xmlcomment('test');`,
			},
			{
				Statement:   `CREATE VIEW xmlview2 AS SELECT xmlconcat('hello', 'you');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `CREATE VIEW xmlview3 AS SELECT xmlelement(name element, xmlattributes (1 as ":one:", 'deuce' as two), 'content&');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `CREATE VIEW xmlview4 AS SELECT xmlelement(name employee, xmlforest(name, age, salary as pay)) FROM emp;`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `CREATE VIEW xmlview5 AS SELECT xmlparse(content '<abc>x</abc>');`,
			},
			{
				Statement:   `CREATE VIEW xmlview6 AS SELECT xmlpi(name foo, 'bar');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `CREATE VIEW xmlview7 AS SELECT xmlroot(xml '<foo/>', version no value, standalone yes);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `CREATE VIEW xmlview8 AS SELECT xmlserialize(content 'good' as char(10));`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `CREATE VIEW xmlview9 AS SELECT xmlserialize(content 'good' as text);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SELECT table_name, view_definition FROM information_schema.views
  WHERE table_name LIKE 'xmlview%' ORDER BY 1;`,
				Results: []sql.Row{{`xmlview1`, `SELECT xmlcomment('test'::text) AS xmlcomment;`}, {`xmlview5`, `SELECT XMLPARSE(CONTENT '<abc>x</abc>'::text STRIP WHITESPACE) AS "xmlparse";`}},
			},
			{
				Statement: `SELECT xpath('/value', data) FROM xmltest;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT xpath(NULL, NULL) IS NULL FROM xmltest;`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `SELECT xpath('', '<!-- error -->');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath('//text()', '<local:data xmlns:local="http://127.0.0.1"><local:piece id="1">number one</local:piece><local:piece id="2" /></local:data>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath('//loc:piece/@id', '<local:data xmlns:local="http://127.0.0.1"><local:piece id="1">number one</local:piece><local:piece id="2" /></local:data>', ARRAY[ARRAY['loc', 'http://127.0.0.1']]);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath('//loc:piece', '<local:data xmlns:local="http://127.0.0.1"><local:piece id="1">number one</local:piece><local:piece id="2" /></local:data>', ARRAY[ARRAY['loc', 'http://127.0.0.1']]);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath('//loc:piece', '<local:data xmlns:local="http://127.0.0.1" xmlns="http://127.0.0.2"><local:piece id="1"><internal>number one</internal><internal2/></local:piece><local:piece id="2" /></local:data>', ARRAY[ARRAY['loc', 'http://127.0.0.1']]);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath('//b', '<a>one <b>two</b> three <b>etc</b></a>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath('//text()', '<root>&lt;</root>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath('//@value', '<root value="&lt;"/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath('''<<invalid>>''', '<root/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath('count(//*)', '<root><sub/><sub/></root>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath('count(//*)=0', '<root><sub/><sub/></root>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath('count(//*)=3', '<root><sub/><sub/></root>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath('name(/*)', '<root><sub/><sub/></root>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath('/nosuchtag', '<root/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath('root', '<root/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `DO $$
DECLARE
  xml_declaration text := '<?xml version="1.0" encoding="ISO-8859-1"?>';`,
			},
			{
				Statement: `  degree_symbol text;`,
			},
			{
				Statement: `  res xml[];`,
			},
			{
				Statement: `BEGIN
  -- Per the documentation, except when the server encoding is UTF8, xpath()
  -- may not work on non-ASCII data.  The untranslatable_character and
  -- undefined_function traps below, currently dead code, will become relevant
  -- if we remove this limitation.
  IF current_setting('server_encoding') <> 'UTF8' THEN
    RAISE LOG 'skip: encoding % unsupported for xpath',
      current_setting('server_encoding');`,
			},
			{
				Statement: `    RETURN;`,
			},
			{
				Statement: `  END IF;`,
			},
			{
				Statement: `  degree_symbol := convert_from('\xc2b0', 'UTF8');`,
			},
			{
				Statement: `  res := xpath('text()', (xml_declaration ||
    '<x>' || degree_symbol || '</x>')::xml);`,
			},
			{
				Statement: `  IF degree_symbol <> res[1]::text THEN
    RAISE 'expected % (%), got % (%)',
      degree_symbol, convert_to(degree_symbol, 'UTF8'),
      res[1], convert_to(res[1]::text, 'UTF8');`,
			},
			{
				Statement: `  END IF;`,
			},
			{
				Statement: `EXCEPTION
  -- character with byte sequence 0xc2 0xb0 in encoding "UTF8" has no equivalent in encoding "LATIN8"
  WHEN untranslatable_character
  -- default conversion function for encoding "UTF8" to "MULE_INTERNAL" does not exist
  OR undefined_function
  -- unsupported XML feature
  OR feature_not_supported THEN
    RAISE LOG 'skip: %', SQLERRM;`,
			},
			{
				Statement: `END
$$;`,
			},
			{
				Statement:   `SELECT xmlexists('//town[text() = ''Toronto'']' PASSING BY REF '<towns><town>Bidford-on-Avon</town><town>Cwmbran</town><town>Bristol</town></towns>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlexists('//town[text() = ''Cwmbran'']' PASSING BY REF '<towns><town>Bidford-on-Avon</town><town>Cwmbran</town><town>Bristol</town></towns>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xmlexists('count(/nosuchtag)' PASSING BY REF '<root/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath_exists('//town[text() = ''Toronto'']','<towns><town>Bidford-on-Avon</town><town>Cwmbran</town><town>Bristol</town></towns>'::xml);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath_exists('//town[text() = ''Cwmbran'']','<towns><town>Bidford-on-Avon</town><town>Cwmbran</town><town>Bristol</town></towns>'::xml);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath_exists('count(/nosuchtag)', '<root/>'::xml);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `INSERT INTO xmltest VALUES (4, '<menu><beers><name>Budvar</name><cost>free</cost><name>Carling</name><cost>lots</cost></beers></menu>'::xml);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `INSERT INTO xmltest VALUES (5, '<menu><beers><name>Molson</name><cost>free</cost><name>Carling</name><cost>lots</cost></beers></menu>'::xml);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `INSERT INTO xmltest VALUES (6, '<myns:menu xmlns:myns="http://myns.com"><myns:beers><myns:name>Budvar</myns:name><myns:cost>free</myns:cost><myns:name>Carling</myns:name><myns:cost>lots</myns:cost></myns:beers></myns:menu>'::xml);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `INSERT INTO xmltest VALUES (7, '<myns:menu xmlns:myns="http://myns.com"><myns:beers><myns:name>Molson</myns:name><myns:cost>free</myns:cost><myns:name>Carling</myns:name><myns:cost>lots</myns:cost></myns:beers></myns:menu>'::xml);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SELECT COUNT(id) FROM xmltest WHERE xmlexists('/menu/beer' PASSING data);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT COUNT(id) FROM xmltest WHERE xmlexists('/menu/beer' PASSING BY REF data BY REF);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT COUNT(id) FROM xmltest WHERE xmlexists('/menu/beers' PASSING BY REF data);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT COUNT(id) FROM xmltest WHERE xmlexists('/menu/beers/name[text() = ''Molson'']' PASSING BY REF data);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT COUNT(id) FROM xmltest WHERE xpath_exists('/menu/beer',data);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT COUNT(id) FROM xmltest WHERE xpath_exists('/menu/beers',data);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT COUNT(id) FROM xmltest WHERE xpath_exists('/menu/beers/name[text() = ''Molson'']',data);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT COUNT(id) FROM xmltest WHERE xpath_exists('/myns:menu/myns:beer',data,ARRAY[ARRAY['myns','http://myns.com']]);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT COUNT(id) FROM xmltest WHERE xpath_exists('/myns:menu/myns:beers',data,ARRAY[ARRAY['myns','http://myns.com']]);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT COUNT(id) FROM xmltest WHERE xpath_exists('/myns:menu/myns:beers/myns:name[text() = ''Molson'']',data,ARRAY[ARRAY['myns','http://myns.com']]);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `CREATE TABLE query ( expr TEXT );`,
			},
			{
				Statement: `INSERT INTO query VALUES ('/menu/beers/cost[text() = ''lots'']');`,
			},
			{
				Statement: `SELECT COUNT(id) FROM xmltest, query WHERE xmlexists(expr PASSING BY REF data);`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement:   `SELECT xml_is_well_formed_document('<foo>bar</foo>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed_document('abc');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed_content('<foo>bar</foo>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed_content('abc');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SET xmloption TO DOCUMENT;`,
			},
			{
				Statement:   `SELECT xml_is_well_formed('abc');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed('<>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed('<abc/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed('<foo>bar</foo>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed('<foo>bar</foo');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed('<foo><bar>baz</foo>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed('<local:data xmlns:local="http://127.0.0.1"><local:piece id="1">number one</local:piece><local:piece id="2" /></local:data>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed('<pg:foo xmlns:pg="http://postgresql.org/stuff">bar</my:foo>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed('<pg:foo xmlns:pg="http://postgresql.org/stuff">bar</pg:foo>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed('<invalidentity>&</abc>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed('<undefinedentity>&idontexist;</abc>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed('<invalidns xmlns=''&lt;''/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed('<relativens xmlns=''relative''/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xml_is_well_formed('<twoerrors>&idontexist;</unbalanced>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SET xmloption TO CONTENT;`,
			},
			{
				Statement:   `SELECT xml_is_well_formed('abc');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `\set VERBOSITY terse
SELECT xpath('/*', '<invalidns xmlns=''&lt;''/>');`,
				ErrorString: `unsupported XML feature at character 20`,
			},
			{
				Statement: `\set VERBOSITY default
SELECT xpath('/*', '<nosuchprefix:tag/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT xpath('/*', '<relativens xmlns=''relative''/>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT XMLPARSE(DOCUMENT '<!DOCTYPE foo [<!ENTITY c SYSTEM "/etc/passwd">]><foo>&c;</foo>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT XMLPARSE(DOCUMENT '<!DOCTYPE foo [<!ENTITY c SYSTEM "/etc/no.such.file">]><foo>&c;</foo>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT XMLPARSE(DOCUMENT '<!DOCTYPE chapter PUBLIC "-//OASIS//DTD DocBook XML V4.1.2//EN" "http://www.oasis-open.org/docbook/xml/4.1.2/docbookx.dtd"><chapter>&nbsp;</chapter>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `CREATE TABLE xmldata(data xml);`,
			},
			{
				Statement: `INSERT INTO xmldata VALUES('<ROWS>
<ROW id="1">
  <COUNTRY_ID>AU</COUNTRY_ID>
  <COUNTRY_NAME>Australia</COUNTRY_NAME>
  <REGION_ID>3</REGION_ID>
</ROW>
<ROW id="2">
  <COUNTRY_ID>CN</COUNTRY_ID>
  <COUNTRY_NAME>China</COUNTRY_NAME>
  <REGION_ID>3</REGION_ID>
</ROW>
<ROW id="3">
  <COUNTRY_ID>HK</COUNTRY_ID>
  <COUNTRY_NAME>HongKong</COUNTRY_NAME>
  <REGION_ID>3</REGION_ID>
</ROW>
<ROW id="4">
  <COUNTRY_ID>IN</COUNTRY_ID>
  <COUNTRY_NAME>India</COUNTRY_NAME>
  <REGION_ID>3</REGION_ID>
</ROW>
<ROW id="5">
  <COUNTRY_ID>JP</COUNTRY_ID>
  <COUNTRY_NAME>Japan</COUNTRY_NAME>
  <REGION_ID>3</REGION_ID><PREMIER_NAME>Sinzo Abe</PREMIER_NAME>
</ROW>
<ROW id="6">
  <COUNTRY_ID>SG</COUNTRY_ID>
  <COUNTRY_NAME>Singapore</COUNTRY_NAME>
  <REGION_ID>3</REGION_ID><SIZE unit="km">791</SIZE>
</ROW>
</ROWS>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SELECT  xmltable.*
   FROM (SELECT data FROM xmldata) x,
        LATERAL XMLTABLE('/ROWS/ROW'
                         PASSING data
                         COLUMNS id int PATH '@id',
                                  _id FOR ORDINALITY,
                                  country_name text PATH 'COUNTRY_NAME/text()' NOT NULL,
                                  country_id text PATH 'COUNTRY_ID',
                                  region_id int PATH 'REGION_ID',
                                  size float PATH 'SIZE',
                                  unit text PATH 'SIZE/@unit',
                                  premier_name text PATH 'PREMIER_NAME' DEFAULT 'not specified');`,
				Results: []sql.Row{},
			},
			{
				Statement: `CREATE VIEW xmltableview1 AS SELECT  xmltable.*
   FROM (SELECT data FROM xmldata) x,
        LATERAL XMLTABLE('/ROWS/ROW'
                         PASSING data
                         COLUMNS id int PATH '@id',
                                  _id FOR ORDINALITY,
                                  country_name text PATH 'COUNTRY_NAME/text()' NOT NULL,
                                  country_id text PATH 'COUNTRY_ID',
                                  region_id int PATH 'REGION_ID',
                                  size float PATH 'SIZE',
                                  unit text PATH 'SIZE/@unit',
                                  premier_name text PATH 'PREMIER_NAME' DEFAULT 'not specified');`,
			},
			{
				Statement: `SELECT * FROM xmltableview1;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `\sv xmltableview1
CREATE OR REPLACE VIEW public.xmltableview1 AS
 SELECT "xmltable".id,
    "xmltable"._id,
    "xmltable".country_name,
    "xmltable".country_id,
    "xmltable".region_id,
    "xmltable".size,
    "xmltable".unit,
    "xmltable".premier_name
   FROM ( SELECT xmldata.data
           FROM xmldata) x,
    LATERAL XMLTABLE(('/ROWS/ROW'::text) PASSING (x.data) COLUMNS id integer PATH ('@id'::text), _id FOR ORDINALITY, country_name text PATH ('COUNTRY_NAME/text()'::text) NOT NULL, country_id text PATH ('COUNTRY_ID'::text), region_id integer PATH ('REGION_ID'::text), size double precision PATH ('SIZE'::text), unit text PATH ('SIZE/@unit'::text), premier_name text DEFAULT ('not specified'::text) PATH ('PREMIER_NAME'::text))
EXPLAIN (COSTS OFF) SELECT * FROM xmltableview1;`,
				Results: []sql.Row{{`Nested Loop`}, {`->  Seq Scan on xmldata`}, {`->  Table Function Scan on "xmltable"`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF, VERBOSE) SELECT * FROM xmltableview1;`,
				Results:   []sql.Row{{`Nested Loop`}, {`Output: "xmltable".id, "xmltable"._id, "xmltable".country_name, "xmltable".country_id, "xmltable".region_id, "xmltable".size, "xmltable".unit, "xmltable".premier_name`}, {`->  Seq Scan on public.xmldata`}, {`Output: xmldata.data`}, {`->  Table Function Scan on "xmltable"`}, {`Output: "xmltable".id, "xmltable"._id, "xmltable".country_name, "xmltable".country_id, "xmltable".region_id, "xmltable".size, "xmltable".unit, "xmltable".premier_name`}, {`Table Function Call: XMLTABLE(('/ROWS/ROW'::text) PASSING (xmldata.data) COLUMNS id integer PATH ('@id'::text), _id FOR ORDINALITY, country_name text PATH ('COUNTRY_NAME/text()'::text) NOT NULL, country_id text PATH ('COUNTRY_ID'::text), region_id integer PATH ('REGION_ID'::text), size double precision PATH ('SIZE'::text), unit text PATH ('SIZE/@unit'::text), premier_name text DEFAULT ('not specified'::text) PATH ('PREMIER_NAME'::text))`}},
			},
			{
				Statement:   `SELECT * FROM XMLTABLE (ROW () PASSING null COLUMNS v1 timestamp) AS f (v1, v2);`,
				ErrorString: `XMLTABLE function has 1 columns available but 2 columns specified`,
			},
			{
				Statement: `SELECT * FROM XMLTABLE(XMLNAMESPACES('http://x.y' AS zz),
                      '/zz:rows/zz:row'
                      PASSING '<rows xmlns="http://x.y"><row><a>10</a></row></rows>'
                      COLUMNS a int PATH 'zz:a');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `CREATE VIEW xmltableview2 AS SELECT * FROM XMLTABLE(XMLNAMESPACES('http://x.y' AS zz),
                      '/zz:rows/zz:row'
                      PASSING '<rows xmlns="http://x.y"><row><a>10</a></row></rows>'
                      COLUMNS a int PATH 'zz:a');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT * FROM xmltableview2;`,
				ErrorString: `relation "xmltableview2" does not exist`,
			},
			{
				Statement: `SELECT * FROM XMLTABLE(XMLNAMESPACES(DEFAULT 'http://x.y'),
                      '/rows/row'
                      PASSING '<rows xmlns="http://x.y"><row><a>10</a></row></rows>'
                      COLUMNS a int PATH 'a');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SELECT * FROM XMLTABLE('.'
                       PASSING '<foo/>'
                       COLUMNS a text PATH 'foo/namespace::node()');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `PREPARE pp AS
SELECT  xmltable.*
   FROM (SELECT data FROM xmldata) x,
        LATERAL XMLTABLE('/ROWS/ROW'
                         PASSING data
                         COLUMNS id int PATH '@id',
                                  _id FOR ORDINALITY,
                                  country_name text PATH 'COUNTRY_NAME' NOT NULL,
                                  country_id text PATH 'COUNTRY_ID',
                                  region_id int PATH 'REGION_ID',
                                  size float PATH 'SIZE',
                                  unit text PATH 'SIZE/@unit',
                                  premier_name text PATH 'PREMIER_NAME' DEFAULT 'not specified');`,
			},
			{
				Statement: `EXECUTE pp;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT xmltable.* FROM xmldata, LATERAL xmltable('/ROWS/ROW[COUNTRY_NAME="Japan" or COUNTRY_NAME="India"]' PASSING data COLUMNS "COUNTRY_NAME" text, "REGION_ID" int);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT xmltable.* FROM xmldata, LATERAL xmltable('/ROWS/ROW[COUNTRY_NAME="Japan" or COUNTRY_NAME="India"]' PASSING data COLUMNS id FOR ORDINALITY, "COUNTRY_NAME" text, "REGION_ID" int);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT xmltable.* FROM xmldata, LATERAL xmltable('/ROWS/ROW[COUNTRY_NAME="Japan" or COUNTRY_NAME="India"]' PASSING data COLUMNS id int PATH '@id', "COUNTRY_NAME" text, "REGION_ID" int);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT xmltable.* FROM xmldata, LATERAL xmltable('/ROWS/ROW[COUNTRY_NAME="Japan" or COUNTRY_NAME="India"]' PASSING data COLUMNS id int PATH '@id');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT xmltable.* FROM xmldata, LATERAL xmltable('/ROWS/ROW[COUNTRY_NAME="Japan" or COUNTRY_NAME="India"]' PASSING data COLUMNS id FOR ORDINALITY);`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT xmltable.* FROM xmldata, LATERAL xmltable('/ROWS/ROW[COUNTRY_NAME="Japan" or COUNTRY_NAME="India"]' PASSING data COLUMNS id int PATH '@id', "COUNTRY_NAME" text, "REGION_ID" int, rawdata xml PATH '.');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT xmltable.* FROM xmldata, LATERAL xmltable('/ROWS/ROW[COUNTRY_NAME="Japan" or COUNTRY_NAME="India"]' PASSING data COLUMNS id int PATH '@id', "COUNTRY_NAME" text, "REGION_ID" int, rawdata xml PATH './*');`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `SELECT * FROM xmltable('/root' passing '<root><element>a1a<!-- aaaa -->a2a<?aaaaa?> <!--z-->  bbbb<x>xxx</x>cccc</element></root>' COLUMNS element text);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT * FROM xmltable('/root' passing '<root><element>a1a<!-- aaaa -->a2a<?aaaaa?> <!--z-->  bbbb<x>xxx</x>cccc</element></root>' COLUMNS element text PATH 'element/text()'); -- should fail`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `select * from xmltable('d/r' passing '<d><r><c><![CDATA[<hello> &"<>!<a>foo</a>]]></c></r><r><c>2</c></r></d>' columns c text);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT * FROM xmltable('/x/a' PASSING '<x><a><ent>&apos;</ent></a><a><ent>&quot;</ent></a><a><ent>&amp;</ent></a><a><ent>&lt;</ent></a><a><ent>&gt;</ent></a></x>' COLUMNS ent text);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT * FROM xmltable('/x/a' PASSING '<x><a><ent>&apos;</ent></a><a><ent>&quot;</ent></a><a><ent>&amp;</ent></a><a><ent>&lt;</ent></a><a><ent>&gt;</ent></a></x>' COLUMNS ent xml);`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `EXPLAIN (VERBOSE, COSTS OFF)
SELECT  xmltable.*
   FROM (SELECT data FROM xmldata) x,
        LATERAL XMLTABLE('/ROWS/ROW'
                         PASSING data
                         COLUMNS id int PATH '@id',
                                  _id FOR ORDINALITY,
                                  country_name text PATH 'COUNTRY_NAME' NOT NULL,
                                  country_id text PATH 'COUNTRY_ID',
                                  region_id int PATH 'REGION_ID',
                                  size float PATH 'SIZE',
                                  unit text PATH 'SIZE/@unit',
                                  premier_name text PATH 'PREMIER_NAME' DEFAULT 'not specified');`,
				Results: []sql.Row{{`Nested Loop`}, {`Output: "xmltable".id, "xmltable"._id, "xmltable".country_name, "xmltable".country_id, "xmltable".region_id, "xmltable".size, "xmltable".unit, "xmltable".premier_name`}, {`->  Seq Scan on public.xmldata`}, {`Output: xmldata.data`}, {`->  Table Function Scan on "xmltable"`}, {`Output: "xmltable".id, "xmltable"._id, "xmltable".country_name, "xmltable".country_id, "xmltable".region_id, "xmltable".size, "xmltable".unit, "xmltable".premier_name`}, {`Table Function Call: XMLTABLE(('/ROWS/ROW'::text) PASSING (xmldata.data) COLUMNS id integer PATH ('@id'::text), _id FOR ORDINALITY, country_name text PATH ('COUNTRY_NAME'::text) NOT NULL, country_id text PATH ('COUNTRY_ID'::text), region_id integer PATH ('REGION_ID'::text), size double precision PATH ('SIZE'::text), unit text PATH ('SIZE/@unit'::text), premier_name text DEFAULT ('not specified'::text) PATH ('PREMIER_NAME'::text))`}},
			},
			{
				Statement: `SELECT xmltable.* FROM xmldata, LATERAL xmltable('/ROWS/ROW[COUNTRY_NAME="Japan" or COUNTRY_NAME="India"]' PASSING data COLUMNS "COUNTRY_NAME" text, "REGION_ID" int) WHERE "COUNTRY_NAME" = 'Japan';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `EXPLAIN (VERBOSE, COSTS OFF)
SELECT xmltable.* FROM xmldata, LATERAL xmltable('/ROWS/ROW[COUNTRY_NAME="Japan" or COUNTRY_NAME="India"]' PASSING data COLUMNS "COUNTRY_NAME" text, "REGION_ID" int) WHERE "COUNTRY_NAME" = 'Japan';`,
				Results: []sql.Row{{`Nested Loop`}, {`Output: "xmltable"."COUNTRY_NAME", "xmltable"."REGION_ID"`}, {`->  Seq Scan on public.xmldata`}, {`Output: xmldata.data`}, {`->  Table Function Scan on "xmltable"`}, {`Output: "xmltable"."COUNTRY_NAME", "xmltable"."REGION_ID"`}, {`Table Function Call: XMLTABLE(('/ROWS/ROW[COUNTRY_NAME="Japan" or COUNTRY_NAME="India"]'::text) PASSING (xmldata.data) COLUMNS "COUNTRY_NAME" text, "REGION_ID" integer)`}, {`Filter: ("xmltable"."COUNTRY_NAME" = 'Japan'::text)`}},
			},
			{
				Statement: `INSERT INTO xmldata VALUES('<ROWS>
<ROW id="10">
  <COUNTRY_ID>CZ</COUNTRY_ID>
  <COUNTRY_NAME>Czech Republic</COUNTRY_NAME>
  <REGION_ID>2</REGION_ID><PREMIER_NAME>Milos Zeman</PREMIER_NAME>
</ROW>
<ROW id="11">
  <COUNTRY_ID>DE</COUNTRY_ID>
  <COUNTRY_NAME>Germany</COUNTRY_NAME>
  <REGION_ID>2</REGION_ID>
</ROW>
<ROW id="12">
  <COUNTRY_ID>FR</COUNTRY_ID>
  <COUNTRY_NAME>France</COUNTRY_NAME>
  <REGION_ID>2</REGION_ID>
</ROW>
</ROWS>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `INSERT INTO xmldata VALUES('<ROWS>
<ROW id="20">
  <COUNTRY_ID>EG</COUNTRY_ID>
  <COUNTRY_NAME>Egypt</COUNTRY_NAME>
  <REGION_ID>1</REGION_ID>
</ROW>
<ROW id="21">
  <COUNTRY_ID>SD</COUNTRY_ID>
  <COUNTRY_NAME>Sudan</COUNTRY_NAME>
  <REGION_ID>1</REGION_ID>
</ROW>
</ROWS>');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SELECT  xmltable.*
   FROM (SELECT data FROM xmldata) x,
        LATERAL XMLTABLE('/ROWS/ROW'
                         PASSING data
                         COLUMNS id int PATH '@id',
                                  _id FOR ORDINALITY,
                                  country_name text PATH 'COUNTRY_NAME' NOT NULL,
                                  country_id text PATH 'COUNTRY_ID',
                                  region_id int PATH 'REGION_ID',
                                  size float PATH 'SIZE',
                                  unit text PATH 'SIZE/@unit',
                                  premier_name text PATH 'PREMIER_NAME' DEFAULT 'not specified');`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT  xmltable.*
   FROM (SELECT data FROM xmldata) x,
        LATERAL XMLTABLE('/ROWS/ROW'
                         PASSING data
                         COLUMNS id int PATH '@id',
                                  _id FOR ORDINALITY,
                                  country_name text PATH 'COUNTRY_NAME' NOT NULL,
                                  country_id text PATH 'COUNTRY_ID',
                                  region_id int PATH 'REGION_ID',
                                  size float PATH 'SIZE',
                                  unit text PATH 'SIZE/@unit',
                                  premier_name text PATH 'PREMIER_NAME' DEFAULT 'not specified')
  WHERE region_id = 2;`,
				Results: []sql.Row{},
			},
			{
				Statement: `EXPLAIN (VERBOSE, COSTS OFF)
SELECT  xmltable.*
   FROM (SELECT data FROM xmldata) x,
        LATERAL XMLTABLE('/ROWS/ROW'
                         PASSING data
                         COLUMNS id int PATH '@id',
                                  _id FOR ORDINALITY,
                                  country_name text PATH 'COUNTRY_NAME' NOT NULL,
                                  country_id text PATH 'COUNTRY_ID',
                                  region_id int PATH 'REGION_ID',
                                  size float PATH 'SIZE',
                                  unit text PATH 'SIZE/@unit',
                                  premier_name text PATH 'PREMIER_NAME' DEFAULT 'not specified')
  WHERE region_id = 2;`,
				Results: []sql.Row{{`Nested Loop`}, {`Output: "xmltable".id, "xmltable"._id, "xmltable".country_name, "xmltable".country_id, "xmltable".region_id, "xmltable".size, "xmltable".unit, "xmltable".premier_name`}, {`->  Seq Scan on public.xmldata`}, {`Output: xmldata.data`}, {`->  Table Function Scan on "xmltable"`}, {`Output: "xmltable".id, "xmltable"._id, "xmltable".country_name, "xmltable".country_id, "xmltable".region_id, "xmltable".size, "xmltable".unit, "xmltable".premier_name`}, {`Table Function Call: XMLTABLE(('/ROWS/ROW'::text) PASSING (xmldata.data) COLUMNS id integer PATH ('@id'::text), _id FOR ORDINALITY, country_name text PATH ('COUNTRY_NAME'::text) NOT NULL, country_id text PATH ('COUNTRY_ID'::text), region_id integer PATH ('REGION_ID'::text), size double precision PATH ('SIZE'::text), unit text PATH ('SIZE/@unit'::text), premier_name text DEFAULT ('not specified'::text) PATH ('PREMIER_NAME'::text))`}, {`Filter: ("xmltable".region_id = 2)`}},
			},
			{
				Statement: `SELECT  xmltable.*
   FROM (SELECT data FROM xmldata) x,
        LATERAL XMLTABLE('/ROWS/ROW'
                         PASSING data
                         COLUMNS id int PATH '@id',
                                  _id FOR ORDINALITY,
                                  country_name text PATH 'COUNTRY_NAME' NOT NULL,
                                  country_id text PATH 'COUNTRY_ID',
                                  region_id int PATH 'REGION_ID',
                                  size float PATH 'SIZE' NOT NULL,
                                  unit text PATH 'SIZE/@unit',
                                  premier_name text PATH 'PREMIER_NAME' DEFAULT 'not specified');`,
				Results: []sql.Row{},
			},
			{
				Statement: `WITH
   x AS (SELECT proname, proowner, procost::numeric, pronargs,
                array_to_string(proargnames,',') as proargnames,
                case when proargtypes <> '' then array_to_string(proargtypes::oid[],',') end as proargtypes
           FROM pg_proc WHERE proname = 'f_leak'),
   y AS (SELECT xmlelement(name proc,
                           xmlforest(proname, proowner,
                                     procost, pronargs,
                                     proargnames, proargtypes)) as proc
           FROM x),
   z AS (SELECT xmltable.*
           FROM y,
                LATERAL xmltable('/proc' PASSING proc
                                 COLUMNS proname name,
                                         proowner oid,
                                         procost float,
                                         pronargs int,
                                         proargnames text,
                                         proargtypes text))
   SELECT * FROM z
   EXCEPT SELECT * FROM x;`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `WITH
   x AS (SELECT proname, proowner, procost::numeric, pronargs,
                array_to_string(proargnames,',') as proargnames,
                case when proargtypes <> '' then array_to_string(proargtypes::oid[],',') end as proargtypes
           FROM pg_proc),
   y AS (SELECT xmlelement(name data,
                           xmlagg(xmlelement(name proc,
                                             xmlforest(proname, proowner, procost,
                                                       pronargs, proargnames, proargtypes)))) as doc
           FROM x),
   z AS (SELECT xmltable.*
           FROM y,
                LATERAL xmltable('/data/proc' PASSING doc
                                 COLUMNS proname name,
                                         proowner oid,
                                         procost float,
                                         pronargs int,
                                         proargnames text,
                                         proargtypes text))
   SELECT * FROM z
   EXCEPT SELECT * FROM x;`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `CREATE TABLE xmltest2(x xml, _path text);`,
			},
			{
				Statement:   `INSERT INTO xmltest2 VALUES('<d><r><ac>1</ac></r></d>', 'A');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `INSERT INTO xmltest2 VALUES('<d><r><bc>2</bc></r></d>', 'B');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `INSERT INTO xmltest2 VALUES('<d><r><cc>3</cc></r></d>', 'C');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `INSERT INTO xmltest2 VALUES('<d><r><dc>2</dc></r></d>', 'D');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `SELECT xmltable.* FROM xmltest2, LATERAL xmltable('/d/r' PASSING x COLUMNS a int PATH '' || lower(_path) || 'c');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT xmltable.* FROM xmltest2, LATERAL xmltable(('/d/r/' || lower(_path) || 'c') PASSING x COLUMNS a int PATH '.');`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT xmltable.* FROM xmltest2, LATERAL xmltable(('/d/r/' || lower(_path) || 'c') PASSING x COLUMNS a int PATH 'x' DEFAULT ascii(_path) - 54);`,
				Results:   []sql.Row{},
			},
			{
				Statement:   `SELECT * FROM XMLTABLE('*' PASSING '<a>a</a>' COLUMNS a xml PATH '.', b text PATH '.', c text PATH '"hi"', d boolean PATH '. = "a"', e integer PATH 'string-length(.)');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `\x
SELECT * FROM XMLTABLE('*' PASSING '<e>pre<!--c1--><?pi arg?><![CDATA[&ent1]]><n2>&amp;deep</n2>post</e>' COLUMNS x xml PATH 'node()', y xml PATH '/');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `\x
SELECT * FROM XMLTABLE('.' PASSING XMLELEMENT(NAME a) columns a varchar(20) PATH '"<foo/>"', b xml PATH '"<foo/>"');`,
				ErrorString: `unsupported XML feature`,
			},
		},
	})
}
