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

func TestXmlFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "xpath",
			SetUpScript: []string{
				"CREATE TABLE t_xml (id INTEGER PRIMARY KEY, v1 XML);",
				"INSERT INTO t_xml VALUES (1, '<note><to>Tove</to></note>'), (2, '<book><title>Introduction to Golang</title><author>John Doe</author></book>'), (3, NULL);",
				"CREATE TABLE t_text (id INTEGER PRIMARY KEY, doc TEXT);",
				"INSERT INTO t_text VALUES (1, '<a>x</a>'), (2, '<b>y</b>');",
				"CREATE VIEW v_xml AS SELECT id, xpath('/book/title/text()', v1) AS x FROM t_xml;",
				"CREATE VIEW v_text AS SELECT id, xpath('/a/text()', doc::xml) AS x FROM t_text;",
				"CREATE VIEW v_cast AS SELECT id, doc::xml AS d FROM t_text;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT xpath('/a/text()', '<a>x</a>'::xml);",
					Expected: []sql.Row{{"{x}"}},
				},
				{
					Query:    "SELECT xpath('/a/text()', '<a>x</a>'), pg_catalog.xpath('/a/text()', '<a>x</a>');",
					Expected: []sql.Row{{"{x}", "{x}"}},
				},
				{
					Query:    "SELECT pg_typeof(xpath('/a', '<a/>'::xml));",
					Expected: []sql.Row{{"xml[]"}},
				},
				{
					Query:    "SELECT xpath('/a', '<a>x</a>'::xml), xpath('/a/b', '<a><b>1</b><b>2</b></a>'::xml), xpath('//b/text()', '<a><b>1</b><b>2</b></a>'::xml), xpath('/a/@id', '<a id=\"7\">x</a>'::xml);",
					Expected: []sql.Row{{"{<a>x</a>}", "{<b>1</b>,<b>2</b>}", "{1,2}", "{7}"}},
				},
				{
					Query:    "SELECT xpath('/a/text()', '<a>&lt;&amp;&gt;\"''</a>'::xml), xpath('/a/@id', '<a id=\"&lt;&amp;&gt;&quot;\">x</a>'::xml);",
					Expected: []sql.Row{{`{"&lt;&amp;&gt;\"'"}`, `{"&lt;&amp;&gt;\""}`}},
				},
				{
					Query:    "SELECT xpath('/a', '<a b=''q\"''>&amp;</a>'::xml), xpath('/a', '<a b=\"&lt;\">&amp;</a>'::xml), xpath('/a', '<a>\"</a>'::xml);",
					Expected: []sql.Row{{`{"<a b=\"q&quot;\">&amp;</a>"}`, `{"<a b=\"&lt;\">&amp;</a>"}`, `{"<a>\"</a>"}`}},
				},
				{
					Query:    "SELECT xpath('/a', '<a><b/><c></c></a>'::xml), xpath('/a', E'<a>\\n  <b/>\\n</a>'::xml);",
					Expected: []sql.Row{{"{<a><b/><c/></a>}", "{\"<a>\n  <b/>\n</a>\"}"}},
				},
				{
					Query:    "SELECT xpath('/a', '<a><![CDATA[<x>]]></a>'::xml), xpath('/a/text()', '<a><![CDATA[<x>]]></a>'::xml), xpath('/a/comment()', '<a><!-- c --></a>'::xml), xpath('/a/node()', '<a>x<b>y</b>z</a>'::xml);",
					Expected: []sql.Row{{"{<a><![CDATA[<x>]]></a>}", "{<![CDATA[<x>]]>}", `{"<!-- c -->"}`, "{x,<b>y</b>,z}"}},
				},
				{
					Query:    "SELECT xpath('/', '<a>x</a>'::xml), xpath('/zzz', '<a>x</a>'::xml), xpath('/a/b/text()', '<a>x</a>'::xml);",
					Expected: []sql.Row{{"{\"<a>x</a>\n\"}", "{}", "{}"}},
				},
				{
					Query:    "SELECT xpath('count(/a/b)', '<a><b/><b/></a>'::xml), xpath('count(/a/b) div 3', '<a><b/><b/></a>'::xml), xpath('1 div 0', '<a/>'::xml), xpath('-1 div 0', '<a/>'::xml), xpath('0 div 0', '<a/>'::xml), xpath('1.5', '<a/>'::xml);",
					Expected: []sql.Row{{"{2}", "{0.6666666666666666}", "{Infinity}", "{-Infinity}", "{NaN}", "{1.5}"}},
				},
				{
					Query:    "SELECT xpath('1=1', '<a/>'::xml), xpath('1=2', '<a/>'::xml), xpath('string(/a)', '<a>x<b>y</b></a>'::xml), xpath('string(/a)', '<a>&lt;</a>'::xml), xpath('concat(\"a\",\"<\")', '<a/>'::xml);",
					Expected: []sql.Row{{"{true}", "{false}", "{xy}", "{&lt;}", "{a&lt;}"}},
				},
				{
					Query:    "SELECT xpath('/a', '<?xml version=\"1.0\"?><a>x</a>'::xml), xpath('/a', E'  \\n <a>x</a>'::xml), xpath('/a', '<!-- c --><a>x</a>'::xml);",
					Expected: []sql.Row{{"{<a>x</a>}", "{<a>x</a>}", "{<a>x</a>}"}},
				},
				{
					Query:    "SELECT xpath('//x:b/text()', '<a xmlns:x=\"urn:x\"><x:b>1</x:b></a>'::xml, ARRAY['x','urn:x']), xpath('//y:b/text()', '<a xmlns:x=\"urn:x\"><x:b>1</x:b></a>'::xml, ARRAY['y','urn:x']);",
					Expected: []sql.Row{{"{1}", "{1}"}},
				},
				{
					Query:    "SELECT xpath('//x:b', '<a xmlns:x=\"urn:x\"><x:b>1</x:b></a>'::xml, ARRAY['x','urn:x']), xpath('/d:a/d:b', '<a xmlns=\"urn:d\"><b>1</b></a>'::xml, ARRAY['d','urn:d']);",
					Expected: []sql.Row{{`{"<x:b xmlns:x=\"urn:x\">1</x:b>"}`, `{"<b xmlns=\"urn:d\">1</b>"}`}},
				},
				{
					Query:    "SELECT xpath('/a', '<a/>'::xml, '{}'::text[]), xpath('/a', '<a/>'::xml, ARRAY['x','urn:x','y','urn:y']);",
					Expected: []sql.Row{{"{<a/>}", "{<a/>}"}},
				},
				{
					Query:       "SELECT xpath('//x:b/text()', '<a xmlns:x=\"urn:x\"><x:b>1</x:b></a>'::xml, ARRAY['x','urn:x','z']);",
					ExpectedErr: "invalid array for XML namespace mapping",
				},
				{
					Query:    "SELECT xpath(NULL, '<a/>'::xml), xpath('/a', NULL), xpath('/a', '<a/>'::xml, NULL);",
					Expected: []sql.Row{{nil, nil, nil}},
				},
				{
					Query:       "SELECT xpath('', '<a>x</a>'::xml);",
					ExpectedErr: "empty XPath expression",
				},
				{
					Query:       "SELECT xpath('/a[', '<a>x</a>'::xml);",
					ExpectedErr: "invalid XPath expression",
				},
				{
					Query:       "SELECT xpath('/a', '<a/><b/>'::xml);",
					ExpectedErr: "could not parse XML document",
				},
				{
					Query:       "SELECT xpath('/a', 'plain'::xml);",
					ExpectedErr: "could not parse XML document",
				},
				{
					Query:       "SELECT xpath('/a', ''::xml);",
					ExpectedErr: "could not parse XML document",
				},
				{
					Query:    "SELECT id, xpath('/book/title/text()', v1) FROM t_xml ORDER BY id;",
					Expected: []sql.Row{{1, "{}"}, {2, `{"Introduction to Golang"}`}, {3, nil}},
				},
				{
					Query:    "SELECT * FROM v_xml ORDER BY id;",
					Expected: []sql.Row{{1, "{}"}, {2, `{"Introduction to Golang"}`}, {3, nil}},
				},
				{
					Query:    "SELECT * FROM v_text ORDER BY id;",
					Expected: []sql.Row{{1, "{x}"}, {2, "{}"}},
				},
				{
					Query:    "SELECT id, d, pg_typeof(d) FROM v_cast ORDER BY id;",
					Expected: []sql.Row{{1, "<a>x</a>", "xml"}, {2, "<b>y</b>", "xml"}},
				},
			},
		},
		{
			Name: "xpath_exists",
			SetUpScript: []string{
				"CREATE TABLE t_xml (id INTEGER PRIMARY KEY, v1 XML);",
				"INSERT INTO t_xml VALUES (1, '<note><to>Tove</to></note>'), (2, '<book><title>Introduction to Golang</title></book>'), (3, NULL);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT xpath_exists('/a', '<a>x</a>'::xml), xpath_exists('/b', '<a>x</a>'::xml), xpath_exists('/a', '<a>x</a>');",
					Expected: []sql.Row{{"t", "f", "t"}},
				},
				{
					Query:    "SELECT xpath_exists('count(/a)', '<a>x</a>'::xml), xpath_exists('1=2', '<a>x</a>'::xml), xpath_exists('string(/b)', '<a>x</a>'::xml), xpath_exists('0', '<a/>'::xml);",
					Expected: []sql.Row{{"t", "t", "t", "t"}},
				},
				{
					Query:    "SELECT xpath_exists('//x:b', '<a xmlns:x=\"urn:x\"><x:b>1</x:b></a>'::xml, ARRAY['x','urn:x']), xpath_exists(NULL, '<a/>'::xml);",
					Expected: []sql.Row{{"t", nil}},
				},
				{
					Query:       "SELECT xpath_exists('/a', '<a/><b/>'::xml);",
					ExpectedErr: "could not parse XML document",
				},
				{
					Query:       "SELECT xpath_exists('/a[', '<a/>'::xml);",
					ExpectedErr: "invalid XPath expression",
				},
				{
					Query:    "SELECT id FROM t_xml WHERE xpath_exists('/book', v1);",
					Expected: []sql.Row{{2}},
				},
			},
		},
		{
			Name: "xml_is_well_formed",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT xml_is_well_formed('<a/>'), xml_is_well_formed('<a>'), xml_is_well_formed('x'), xml_is_well_formed('<a/><b/>'), xml_is_well_formed('');",
					Expected: []sql.Row{{"t", "f", "t", "t", "t"}},
				},
				{
					Query:    "SELECT xml_is_well_formed_document('<a/>'), xml_is_well_formed_document('x'), xml_is_well_formed_document('<a/><b/>'), xml_is_well_formed_document('<?xml version=\"1.0\"?><a/>'), xml_is_well_formed_document('<!-- c --><a/>'), xml_is_well_formed_document(' <a/>'), xml_is_well_formed_document('');",
					Expected: []sql.Row{{"t", "f", "f", "t", "t", "t", "f"}},
				},
				{
					Query:    "SELECT xml_is_well_formed_content('<a/>'), xml_is_well_formed_content('x'), xml_is_well_formed_content('<a/><b/>'), xml_is_well_formed_content('<a>'), xml_is_well_formed_content('');",
					Expected: []sql.Row{{"t", "t", "t", "f", "t"}},
				},
				{
					Query:    "SET xmloption TO document;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT xml_is_well_formed('<a/><b/>'), xml_is_well_formed('<a/>');",
					Expected: []sql.Row{{"f", "t"}},
				},
			},
		},
		{
			Name: "XMLPARSE",
			SetUpScript: []string{
				"CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
				`INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT XMLPARSE(DOCUMENT '<a>x</a>');",
					Expected:         []sql.Row{{"<a>x</a>"}},
					ExpectedColNames: []string{"xmlparse"},
				},
				{
					Query:    "SELECT XMLPARSE(CONTENT 'ab<c/>'), XMLPARSE(CONTENT ''), XMLPARSE(CONTENT NULL), XMLPARSE(CONTENT 1);",
					Expected: []sql.Row{{"ab<c/>", "", nil, "1"}},
				},
				{
					Query:    "SELECT XMLPARSE(DOCUMENT '<a/>' PRESERVE WHITESPACE), XMLPARSE(DOCUMENT '<a> </a>' STRIP WHITESPACE);",
					Expected: []sql.Row{{"<a/>", "<a> </a>"}},
				},
				{
					Query:    `SELECT XMLPARSE(DOCUMENT '<?xml version="1.0"?><a/>'), XMLPARSE(DOCUMENT '<?xml version="1.1"?><a/>');`,
					Expected: []sql.Row{{"<a/>", `<?xml version="1.1"?><a/>`}},
				},
				{
					Query:    "SELECT pg_typeof(XMLPARSE(CONTENT '<a/>'));",
					Expected: []sql.Row{{"xml"}},
				},
				{
					Query:       "SELECT XMLPARSE(DOCUMENT 'a<b/>');",
					ExpectedErr: "invalid XML document",
				},
				{
					Query:       "SELECT XMLPARSE(CONTENT '<a');",
					ExpectedErr: "invalid XML content",
				},
				{
					Query:    "SELECT XMLPARSE(DOCUMENT doc::text) FROM docs WHERE id = 2;",
					Expected: []sql.Row{{"<r/>"}},
				},
			},
		},
		{
			Name: "xmlconcat",
			SetUpScript: []string{
				"CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
				`INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT xmlconcat('<a/>', '<b/>');",
					Expected:         []sql.Row{{"<a/><b/>"}},
					ExpectedColNames: []string{"xmlconcat"},
				},
				{
					Query:    "SELECT xmlconcat('<a/>', NULL, '<b/>'), xmlconcat(NULL, NULL), xmlconcat('a', 'b'), xmlconcat('<a/>');",
					Expected: []sql.Row{{"<a/><b/>", nil, "ab", "<a/>"}},
				},
				{
					Query:    `SELECT xmlconcat('<?xml version="1.0" standalone="yes"?><a/>', '<?xml version="1.0" standalone="no"?><b/>');`,
					Expected: []sql.Row{{`<?xml version="1.0" standalone="no"?><a/><b/>`}},
				},
				{
					Query:    `SELECT xmlconcat('<?xml version="1.1"?><a/>', '<?xml version="1.0"?><b/>'), xmlconcat('<?xml version="1.1"?><a/>', '<?xml version="1.1"?><b/>');`,
					Expected: []sql.Row{{"<a/><b/>", `<?xml version="1.1"?><a/><b/>`}},
				},
				{
					Query:    `SELECT xmlconcat('<?xml version="1.0" standalone="yes"?><a/>', '<b/>'), xmlconcat('<?xml version="1.1" standalone="yes"?><a/>', '<?xml version="1.0"?><b/>');`,
					Expected: []sql.Row{{"<a/><b/>", "<a/><b/>"}},
				},
				{
					Query:    "SELECT xmlconcat(doc, '<z/>') FROM docs ORDER BY id;",
					Expected: []sql.Row{{`<r><i n="a">1</i><i n="b">2</i></r><z/>`}, {"<r/><z/>"}, {"<z/>"}},
				},
				{
					Query:       "SELECT xmlconcat('<a/>'::text, '<b/>');",
					ExpectedErr: "argument of XMLCONCAT must be type xml, not type text",
				},
			},
		},
		{
			Name: "xmlcomment",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT xmlcomment('hi'), xmlcomment(''), xmlcomment(NULL);",
					Expected: []sql.Row{{"<!--hi-->", "<!---->", nil}},
				},
				{
					Query:       "SELECT xmlcomment('a--b');",
					ExpectedErr: "invalid XML comment",
				},
				{
					Query:       "SELECT xmlcomment('a-');",
					ExpectedErr: "invalid XML comment",
				},
			},
		},
		{
			Name: "xmlelement",
			SetUpScript: []string{
				"CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
				`INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT xmlelement(NAME a, 'x'), xmlelement(NAME a), xmlelement(NAME a, NULL);",
					Expected:         []sql.Row{{"<a>x</a>", "<a/>", "<a/>"}},
					ExpectedColNames: []string{"xmlelement", "xmlelement", "xmlelement"},
				},
				{
					Query:    `SELECT xmlelement(NAME a, 'x', 'y', 1, 2.5, true, NULL, '<b/>'::xml, '<c/>', 'q<&>"''');`,
					Expected: []sql.Row{{`<a>xy12.5true<b/>&lt;c/&gt;q&lt;&amp;&gt;"'</a>`}},
				},
				{
					Query:    `SELECT xmlelement(NAME a, xmlattributes('1' AS x, 2 AS y, NULL AS z, true AS b, 'q<&>"''' AS e), 'body');`,
					Expected: []sql.Row{{`<a x="1" y="2" b="true" e="q&lt;&amp;&gt;&quot;'">body</a>`}},
				},
				{
					Query:    "SELECT xmlelement(NAME a, xmlattributes('1' AS x), xmlelement(NAME b, 'inner'));",
					Expected: []sql.Row{{`<a x="1"><b>inner</b></a>`}},
				},
				{
					Query:    "SELECT xmlelement(NAME a, xmlattributes(id AS x), 'y') FROM (SELECT 7 AS id) t;",
					Expected: []sql.Row{{`<a x="7">y</a>`}},
				},
				{
					Query:    `SELECT xmlelement(NAME "a b", 'x'), xmlelement(NAME "A", 'x'), xmlelement(NAME "a:b", 'x'), xmlelement(NAME "1a", 'x'), xmlelement(NAME a, xmlattributes('v' AS "a b"));`,
					Expected: []sql.Row{{"<a_x0020_b>x</a_x0020_b>", "<A>x</A>", "<a:b>x</a:b>", "<_x0031_a>x</_x0031_a>", `<a a_x0020_b="v"/>`}},
				},
				{
					Query:    "SELECT xmlelement(NAME a, '2001-02-03'::date, '2001-02-03 04:05:06'::timestamp, '04:05:06'::time, 'ab'::bytea, 1.50::numeric, ARRAY[1,2]);",
					Expected: []sql.Row{{"<a>2001-02-032001-02-03T04:05:0604:05:06YWI=1.50<element>1</element><element>2</element></a>"}},
				},
				{
					Query:    "SELECT xmlelement(NAME a, 1.0::float8, 'NaN'::float8, 2.5::float4);",
					Expected: []sql.Row{{"<a>1NaN2.5</a>"}},
				},
				{
					Query:    "SELECT xmlelement(NAME a, '<b>'::text), xmlelement(NAME a, 'x'::char(3)), xmlelement(NAME a, xmlattributes('<b/>'::xml AS x));",
					Expected: []sql.Row{{"<a>&lt;b&gt;</a>", "<a>x  </a>", `<a x="&lt;b/&gt;"/>`}},
				},
				{
					Query:    "SELECT xmlelement(NAME a, xmlattributes(E'a\\nb' AS x, E'a\\tb' AS y), E'c\\nd');",
					Expected: []sql.Row{{"<a x=\"a&#10;b\" y=\"a&#9;b\">c\nd</a>"}},
				},
				{
					Query:    `SELECT xmlelement(NAME a, '<?xml version="1.0"?><b/>'::xml, '<?xml version="1.1"?><c/>'::xml);`,
					Expected: []sql.Row{{`<a><b/><?xml version="1.1"?><c/></a>`}},
				},
				{
					Query:    "SELECT xmlelement(NAME d, xmlattributes(id), doc) FROM docs ORDER BY id;",
					Expected: []sql.Row{{`<d id="1"><r><i n="a">1</i><i n="b">2</i></r></d>`}, {`<d id="2"><r/></d>`}, {`<d id="3"/>`}},
				},
				{
					Query:    "SELECT pg_typeof(xmlelement(NAME a));",
					Expected: []sql.Row{{"xml"}},
				},
				{
					Query:    "SET xmlbinary TO hex;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT xmlelement(NAME a, 'ab'::bytea);",
					Expected: []sql.Row{{"<a>6162</a>"}},
				},
				{
					Query:    "SET xmlbinary TO base64;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT xmlelement(NAME a, 'ab'::bytea);",
					Expected: []sql.Row{{"<a>YWI=</a>"}},
				},
				{
					Query:       "SELECT xmlelement(NAME a, xmlattributes(1, 'x' AS q));",
					ExpectedErr: "unnamed XML attribute value must be a column reference",
				},
				{
					Query:       "SELECT xmlelement(NAME a, xmlattributes('1' AS x, '2' AS x));",
					ExpectedErr: `XML attribute name "x" appears more than once`,
				},
			},
		},
		{
			Name: "xmlforest",
			SetUpScript: []string{
				"CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
				`INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT xmlforest(1 AS a, 'x' AS b, NULL AS c);",
					Expected:         []sql.Row{{"<a>1</a><b>x</b>"}},
					ExpectedColNames: []string{"xmlforest"},
				},
				{
					Query:    `SELECT xmlforest(NULL AS c), xmlforest(1 AS a, 2 AS a), xmlforest('a<b' AS a), xmlforest(1 AS "A b");`,
					Expected: []sql.Row{{"", "<a>1</a><a>2</a>", "<a>a&lt;b</a>", "<A_x0020_b>1</A_x0020_b>"}},
				},
				{
					Query:    "SELECT xmlforest('<a/>'::xml AS a), xmlforest(true AS a, '2020-01-01 10:00:00'::timestamp AS b, ARRAY[1,2] AS c);",
					Expected: []sql.Row{{"<a><a/></a>", "<a>true</a><b>2020-01-01T10:00:00</b><c><element>1</element><element>2</element></c>"}},
				},
				{
					Query:    "SELECT xmlforest(id, doc) FROM docs ORDER BY id;",
					Expected: []sql.Row{{`<id>1</id><doc><r><i n="a">1</i><i n="b">2</i></r></doc>`}, {"<id>2</id><doc><r/></doc>"}, {"<id>3</id>"}},
				},
				{
					Query:    "SELECT xmlelement(NAME r, xmlforest(1 AS a, 2 AS b)), xmlforest(1 AS a) IS DOCUMENT;",
					Expected: []sql.Row{{"<r><a>1</a><b>2</b></r>", "t"}},
				},
				{
					Query:       "SELECT xmlforest(1);",
					ExpectedErr: "unnamed XML element value must be a column reference",
				},
			},
		},
		{
			Name: "xmlpi",
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT xmlpi(NAME php, 'echo 1;');",
					Expected:         []sql.Row{{"<?php echo 1;?>"}},
					ExpectedColNames: []string{"xmlpi"},
				},
				{
					Query:    "SELECT xmlpi(NAME php), xmlpi(NAME php, '  x'), xmlpi(NAME php, ''), xmlpi(NAME php, 1), xmlpi(NAME php, NULL), xmlpi(NAME php, 'a?b>');",
					Expected: []sql.Row{{"<?php?>", "<?php x?>", "<?php ?>", "<?php 1?>", nil, "<?php a?b>?>"}},
				},
				{
					Query:    `SELECT xmlpi(NAME "a b", 'x'), xmlpi(NAME "a:b"), xmlelement(NAME a, xmlpi(NAME php, 'x'));`,
					Expected: []sql.Row{{"<?a_x0020_b x?>", "<?a:b?>", "<a><?php x?></a>"}},
				},
				{
					Query:       "SELECT xmlpi(NAME php, 'a?>b');",
					ExpectedErr: "invalid XML processing instruction",
				},
				{
					Query:       "SELECT xmlpi(NAME xml, 'x');",
					ExpectedErr: "invalid XML processing instruction",
				},
				{
					Query:       `SELECT xmlpi(NAME "xMl");`,
					ExpectedErr: "invalid XML processing instruction",
				},
			},
		},
		{
			Name: "xmlroot",
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT xmlroot('<a/>', VERSION '1.1');",
					Expected:         []sql.Row{{`<?xml version="1.1"?><a/>`}},
					ExpectedColNames: []string{"xmlroot"},
				},
				{
					Query:    "SELECT xmlroot('<a/>', VERSION NO VALUE), xmlroot('<a/>', VERSION '1.0', STANDALONE YES), xmlroot('<a/>', VERSION '1.0', STANDALONE NO), xmlroot('<a/>', VERSION '1.0', STANDALONE NO VALUE);",
					Expected: []sql.Row{{"<a/>", `<?xml version="1.0" standalone="yes"?><a/>`, `<?xml version="1.0" standalone="no"?><a/>`, "<a/>"}},
				},
				{
					Query:    `SELECT xmlroot('<?xml version="1.1" standalone="yes"?><a/>', VERSION NO VALUE), xmlroot('<?xml version="1.1" standalone="yes"?><a/>', VERSION '1.0', STANDALONE NO VALUE), xmlroot('<?xml version="1.1" standalone="no"?><a/>', VERSION '1.1');`,
					Expected: []sql.Row{{`<?xml version="1.0" standalone="yes"?><a/>`, "<a/>", `<?xml version="1.1" standalone="no"?><a/>`}},
				},
				{
					Query:    "SELECT xmlroot('<a/>', VERSION NULL), xmlroot('<a/>', VERSION 1), xmlroot(NULL, VERSION '1.0'), xmlroot('a<b/>', VERSION '1.0'), xmlroot('<a/>', VERSION NO VALUE, STANDALONE YES);",
					Expected: []sql.Row{{"<a/>", `<?xml version="1"?><a/>`, nil, "a<b/>", `<?xml version="1.0" standalone="yes"?><a/>`}},
				},
				{
					Query:    "SELECT pg_typeof(xmlroot('<a/>', VERSION '1.0'));",
					Expected: []sql.Row{{"xml"}},
				},
				{
					Query:       "SELECT xmlroot('<a/>'::text, VERSION '1.0');",
					ExpectedErr: "argument of XMLROOT must be type xml, not type text",
				},
			},
		},
		{
			Name: "xmlexists",
			SetUpScript: []string{
				"CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
				`INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT xmlexists('//a' PASSING '<a/>');",
					Expected:         []sql.Row{{"t"}},
					ExpectedColNames: []string{"xmlexists"},
				},
				{
					Query:    "SELECT xmlexists('//a' PASSING BY REF '<b/>'), xmlexists('//a' PASSING BY VALUE '<a/>' BY REF), pg_catalog.xmlexists('//a', '<a/>'::xml), xmlexists('//a' PASSING NULL), xmlexists('count(//a)' PASSING '<a/>');",
					Expected: []sql.Row{{"f", "t", "t", nil, "t"}},
				},
				{
					Query:    `SELECT id FROM docs WHERE xmlexists('/r/i[@n="b"]' PASSING doc);`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:       "SELECT xmlexists('' PASSING '<a/>');",
					ExpectedErr: "empty XPath expression",
				},
				{
					Query:       "SELECT xmlexists('//a' PASSING ('<a'));",
					ExpectedErr: "invalid XML content",
				},
			},
		},
		{
			Name: "XMLSERIALIZE",
			SetUpScript: []string{
				"CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
				`INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT XMLSERIALIZE(DOCUMENT '<a/>' AS text);",
					Expected:         []sql.Row{{"<a/>"}},
					ExpectedColNames: []string{"xmlserialize"},
				},
				{
					Query:    "SELECT XMLSERIALIZE(CONTENT '<a/>' AS char(10)), XMLSERIALIZE(CONTENT 'a<b/>' AS text), XMLSERIALIZE(CONTENT NULL AS text), XMLSERIALIZE(CONTENT '<a/>' AS name), XMLSERIALIZE(CONTENT '<a/>' AS varchar);",
					Expected: []sql.Row{{"<a/>      ", "a<b/>", nil, "<a/>", "<a/>"}},
				},
				{
					Query:    `SELECT XMLSERIALIZE(CONTENT ('<?xml version="1.0"?><a/>'::xml) AS text), pg_typeof(XMLSERIALIZE(CONTENT '<a/>' AS varchar(20)));`,
					Expected: []sql.Row{{`<?xml version="1.0"?><a/>`, "character varying"}},
				},
				{
					Query:    "SELECT XMLSERIALIZE(CONTENT doc AS text) FROM docs ORDER BY id;",
					Expected: []sql.Row{{`<r><i n="a">1</i><i n="b">2</i></r>`}, {"<r/>"}, {nil}},
				},
				{
					Query:       "SELECT XMLSERIALIZE(CONTENT '<a/>' AS varchar(2));",
					ExpectedErr: "value too long for type",
				},
				{
					Query:       "SELECT XMLSERIALIZE(CONTENT '<a/>' AS int);",
					ExpectedErr: "cannot cast XMLSERIALIZE result to integer",
				},
				{
					Query:       "SELECT XMLSERIALIZE(CONTENT '<a/>' AS bytea);",
					ExpectedErr: "cannot cast XMLSERIALIZE result to bytea",
				},
				{
					Query:       "SELECT XMLSERIALIZE(CONTENT '<a/>' AS xml);",
					ExpectedErr: "cannot cast XMLSERIALIZE result to xml",
				},
				{
					Query:       "SELECT XMLSERIALIZE(DOCUMENT 'a<b/>' AS text);",
					ExpectedErr: "not an XML document",
				},
				{
					Query:       "SELECT XMLSERIALIZE(CONTENT '<a' AS text);",
					ExpectedErr: "invalid XML content",
				},
				{
					Query:       "SELECT XMLSERIALIZE(CONTENT '<a/>'::text AS text);",
					ExpectedErr: "argument of XMLSERIALIZE must be type xml, not type text",
				},
			},
		},
		{
			Name: "IS DOCUMENT",
			SetUpScript: []string{
				"CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
				`INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT '<a/>' IS DOCUMENT, '<a/><b/>' IS DOCUMENT, '<a/>' IS NOT DOCUMENT, NULL IS DOCUMENT, 'x' IS DOCUMENT;",
					Expected: []sql.Row{{"t", "f", "f", nil, "f"}},
				},
				{
					Query:    "SELECT id FROM docs WHERE doc IS DOCUMENT ORDER BY id;",
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query:       "SELECT '<a' IS DOCUMENT;",
					ExpectedErr: "invalid XML content",
				},
				{
					Query:       "SELECT '<a/>'::text IS DOCUMENT;",
					ExpectedErr: "argument of IS DOCUMENT must be type xml, not type text",
				},
				{
					Query:       "SELECT 1 IS DOCUMENT;",
					ExpectedErr: "argument of IS DOCUMENT must be type xml, not type integer",
				},
			},
		},
		{
			Name: "xmlagg",
			SetUpScript: []string{
				"CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
				`INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "SELECT xmlagg(x) FROM (VALUES ('<a/>'::xml), ('<b/>'), (NULL)) AS t(x);",
					Expected:         []sql.Row{{"<a/><b/>"}},
					ExpectedColNames: []string{"xmlagg"},
				},
				{
					Query:    "SELECT xmlagg(x) FROM (VALUES (NULL::xml)) AS t(x);",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    "SELECT xmlagg(x) FROM (VALUES ('<a/>'::xml)) AS t(x) WHERE false;",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    "SELECT pg_typeof(xmlagg(x)) FROM (VALUES ('<a/>'::xml)) AS t(x);",
					Expected: []sql.Row{{"xml"}},
				},
				{
					Query:    `SELECT xmlagg(x) FROM (VALUES ('<?xml version="1.0" standalone="yes"?><a/>'::xml), ('<?xml version="1.0" standalone="no"?><b/>')) AS t(x);`,
					Expected: []sql.Row{{`<?xml version="1.0" standalone="no"?><a/><b/>`}},
				},
				{
					Query:    "SELECT x, xmlagg(y) FROM (VALUES (1, '<a/>'::xml), (1, '<b/>'), (2, NULL), (3, '<c/>')) AS t(x, y) GROUP BY x ORDER BY x;",
					Expected: []sql.Row{{1, "<a/><b/>"}, {2, nil}, {3, "<c/>"}},
				},
				{
					Query:    "SELECT xmlagg(y) OVER (ORDER BY x) FROM (VALUES (1, '<a/>'::xml), (2, '<b/>')) AS t(x, y);",
					Expected: []sql.Row{{"<a/>"}, {"<a/><b/>"}},
				},
				{
					Query:    "SELECT xmlagg(xmlelement(NAME d, xmlattributes(id))), xmlagg(doc) FROM docs;",
					Expected: []sql.Row{{`<d id="1"/><d id="2"/><d id="3"/>`, `<r><i n="a">1</i><i n="b">2</i></r><r/>`}},
				},
			},
		},
		{
			Name: "XMLTABLE",
			SetUpScript: []string{
				"CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
				`INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT * FROM XMLTABLE('/a' PASSING '<a>x</a>' COLUMNS v text PATH 'text()');",
					Expected: []sql.Row{{"x"}},
				},
				{
					Query:            `SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row id="1"><n>a</n></row></r>'::xml) COLUMNS id int PATH '@id', n text) AS t(x, y);`,
					Expected:         []sql.Row{{1, "a"}},
					ExpectedColNames: []string{"x", "y"},
				},
				{
					Query:            `SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row id="1"><n>a</n></row></r>'::xml) COLUMNS id int PATH '@id', n text) t;`,
					Expected:         []sql.Row{{1, "a"}},
					ExpectedColNames: []string{"id", "n"},
				},
				{
					Query:    `SELECT t.n, t.id FROM XMLTABLE('/r/row' PASSING ('<r><row id="1"><n>a</n></row></r>'::xml) COLUMNS id int PATH '@id', n text) AS t;`,
					Expected: []sql.Row{{"a", 1}},
				},
				{
					Query:    "SELECT xmltable.v FROM XMLTABLE('/a' PASSING '<a>x</a>' COLUMNS v text PATH 'text()');",
					Expected: []sql.Row{{"x"}},
				},
				{
					Query:    `SELECT * FROM XMLTABLE('/r/row' PASSING BY REF ('<r><row id="1"><n>a</n></row></r>'::xml) BY REF COLUMNS id int PATH '@id');`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM XMLTABLE('/r/row' PASSING BY VALUE ('<r><row id="1"><n>a</n></row></r>'::xml) COLUMNS id int PATH '@id');`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT * FROM XMLTABLE(XMLNAMESPACES('urn:x' AS x, 'urn:y' AS y), '/x:r/x:row' PASSING ('<r xmlns="urn:x"><row><n>a</n></row></r>'::xml) COLUMNS n text PATH 'x:n');`,
					Expected: []sql.Row{{"a"}},
				},
				{
					Query:    "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>a</n><n>b</n></row></r>'::xml) COLUMNS n xml PATH 'n', m xml PATH 'n/text()');",
					Expected: []sql.Row{{"<n>a</n><n>b</n>", "ab"}},
				},
				{
					Query:    "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>&lt;&amp;</n></row></r>'::xml) COLUMNS n text PATH 'n', m xml PATH 'n/text()');",
					Expected: []sql.Row{{"<&", "&lt;&amp;"}},
				},
				{
					Query:    `SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>x</n></row></r>'::xml) COLUMNS n text PATH 'count(n)', c int PATH 'count(n)', b bool PATH 'n = "x"', s text PATH 'string(n)', k text PATH 'count(n) div 0', j text PATH '1.5 + 1', m text PATH '0 div 0', f text PATH 'false()');`,
					Expected: []sql.Row{{"1", 1, "t", "x", "Infinity", "2.5", "NaN", "false"}},
				},
				{
					Query:    "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>x</n></row></r>'::xml) COLUMNS n text PATH 'm', d text PATH 'm' DEFAULT 'd', o int PATH 'm' DEFAULT 5, p int DEFAULT 1 + 1);",
					Expected: []sql.Row{{nil, "d", 5, 2}},
				},
				{
					Query:    "SELECT * FROM XMLTABLE('/r/none' PASSING ('<r><row><n>x</n></row></r>'::xml) COLUMNS n text);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>x</n></row><row/></r>'::xml) COLUMNS n text PATH 'n', d text PATH 'n' DEFAULT 'dd', o FOR ORDINALITY);",
					Expected: []sql.Row{{"x", "x", 1}, {nil, "dd", 2}},
				},
				{
					Query:    `SELECT * FROM XMLTABLE('/r/row/@id' PASSING ('<r><row id="1"/><row id="2"/></r>'::xml) COLUMNS v text PATH '.');`,
					Expected: []sql.Row{{"1"}, {"2"}},
				},
				{
					Query:    `SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row id="1"/></r>'::xml) COLUMNS v text PATH '.', w xml PATH '.', "Id" int PATH '@id', i2 int PATH '@id' DEFAULT 9);`,
					Expected: []sql.Row{{"", `<row id="1"/>`, 1, 1}},
				},
				{
					Query:    `SELECT * FROM XMLTABLE('/r' PASSING ('<r a="1" b="2"/>'::xml) COLUMNS a int, b int, c int PATH '@a + @b');`,
					Expected: []sql.Row{{nil, nil, 3}},
				},
				{
					Query:    "SELECT * FROM XMLTABLE('/r' PASSING ('<r>1</r>'::xml) COLUMNS v text PATH 'text()', w text PATH '.', y int PATH '.');",
					Expected: []sql.Row{{"1", "1", 1}},
				},
				{
					Query:    "SELECT * FROM XMLTABLE('/r' PASSING ('<r><!-- c --></r>'::xml) COLUMNS v text PATH 'comment()', w xml PATH 'comment()');",
					Expected: []sql.Row{{" c ", "<!-- c -->"}},
				},
				{
					Query:    "SELECT * FROM XMLTABLE('/r' PASSING ('<r><![CDATA[<x>]]></r>'::xml) COLUMNS v text PATH 'text()', w xml PATH 'text()', u text PATH '.', z xml PATH '.');",
					Expected: []sql.Row{{"<x>", "<![CDATA[<x>]]>", "<x>", "<r><![CDATA[<x>]]></r>"}},
				},
				{
					Query:    "SELECT * FROM XMLTABLE('/r' PASSING ('<r>  <a/>  </r>'::xml) COLUMNS v text PATH '.', w xml PATH '.');",
					Expected: []sql.Row{{"    ", "<r>  <a/>  </r>"}},
				},
				{
					Query:    `SELECT * FROM XMLTABLE('/r' PASSING '<r x="a&lt;b&amp;c"><t>p&lt;q</t><!--c&lt;--></r>' COLUMNS v text PATH '@x', w xml PATH '@x', t text PATH 't/text()', tx xml PATH 't/text()', c xml PATH 'comment()', ct text PATH 'comment()');`,
					Expected: []sql.Row{{"a<b&c", "a&lt;b&amp;c", "p<q", "p&lt;q", "<!--c&lt;-->", "c&lt;"}},
				},
				{
					Query:    `SELECT * FROM XMLTABLE('/r' PASSING '<r x="a&#10;b" />' COLUMNS v text PATH '@x', w xml PATH '@x');`,
					Expected: []sql.Row{{"a\nb", "a\nb"}},
				},
				{
					Query:    "SELECT * FROM XMLTABLE('/r/row' PASSING '<r><row>a<n>1</n>b</row></r>' COLUMNS n xml PATH 'text()', m xml PATH 'node()');",
					Expected: []sql.Row{{"ab", "a<n>1</n>b"}},
				},
				{
					Query:    `SELECT * FROM XMLTABLE('/r' PASSING '<?xml version="1.0"?><r/>' COLUMNS n xml PATH '.');`,
					Expected: []sql.Row{{"<r/>"}},
				},
				{
					Query:    "SELECT * FROM XMLTABLE('count(/r)' PASSING '<r/>' COLUMNS n xml PATH '.');",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS id int PATH 1, v int PATH 'concat(\"1\", \"\")');",
					Expected: []sql.Row{{1, 1}},
				},
				{
					Query:    `SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS v varchar(2) PATH 'string("ab")', w char(4) PATH 'string("ab")', d date PATH 'string("2001-02-03")', n numeric(5,2) PATH 'string("1.234")', a int[] PATH 'string("{1,2}")');`,
					Expected: []sql.Row{{"ab", "ab  ", "2001-02-03", Numeric("1.23"), "{1,2}"}},
				},
				{
					Query:    "SELECT * FROM XMLTABLE('/r/row' PASSING '<r><row><n>1</n></row></r>' COLUMNS n int NOT NULL PATH 'n', m int DEFAULT 5 PATH 'n', o int PATH 'q' DEFAULT 5 NOT NULL, p int PATH 'q' NOT NULL DEFAULT 5, q int PATH 'a' NULL);",
					Expected: []sql.Row{{1, 1, 5, 5, nil}},
				},
				{
					Query:    "SELECT * FROM XMLTABLE('/r/row' PASSING (NULL::xml) COLUMNS id int PATH '@id');",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM XMLTABLE('/r' PASSING NULL COLUMNS n xml PATH '.');",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT x.* FROM (VALUES ('<r><row><n>1</n></row></r>'), ('<r><row><n>2</n></row><row><n>3</n></row></r>')) AS d(doc), XMLTABLE('/r/row' PASSING (d.doc::xml) COLUMNS n int) AS x ORDER BY n;",
					Expected: []sql.Row{{1}, {2}, {3}},
				},
				{
					Query:    "SELECT id, x.* FROM docs, XMLTABLE('/r/i' PASSING doc COLUMNS n text PATH '@n', v int PATH 'text()') AS x ORDER BY id, n;",
					Expected: []sql.Row{{1, "a", 1}, {1, "b", 2}},
				},
				{
					Query:    "SELECT id, x.n FROM docs LEFT JOIN XMLTABLE('/r/i' PASSING doc COLUMNS n text PATH '@n') AS x ON TRUE ORDER BY id, n;",
					Expected: []sql.Row{{1, "a"}, {1, "b"}, {2, nil}, {3, nil}},
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>a</n><n>b</n></row></r>'::xml) COLUMNS n text PATH 'n');",
					ExpectedErr: "more than one value returned by column XPath expression",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r/row' PASSING '<r><row>a<n>1</n>b</row></r>' COLUMNS n text PATH 'node()');",
					ExpectedErr: "more than one value returned by column XPath expression",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>x</n></row></r>'::xml) COLUMNS n int PATH 'n');",
					ExpectedErr: `invalid input syntax for type int4: "x"`,
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>x</n></row></r>'::xml) COLUMNS n text PATH 'm' NOT NULL);",
					ExpectedErr: `null is not allowed in column "n"`,
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS v int PATH 'q' DEFAULT 'zz');",
					ExpectedErr: `invalid input syntax for type int4: "zz"`,
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS v int PATH 'q' DEFAULT 'a'::text);",
					ExpectedErr: "argument of XMLTABLE must be type integer, not type text",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r/row' PASSING '<r>' COLUMNS id int PATH '@id');",
					ExpectedErr: "invalid XML content",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r/row' PASSING 'x' COLUMNS id int PATH '@id');",
					ExpectedErr: "could not parse XML document",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r' PASSING '<r/><s/>' COLUMNS n xml PATH '.');",
					ExpectedErr: "could not parse XML document",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r' PASSING '' COLUMNS n xml PATH '.');",
					ExpectedErr: "could not parse XML document",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r/[' PASSING '<r/>' COLUMNS id int PATH '@id');",
					ExpectedErr: "invalid XPath expression",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS id int PATH '@[');",
					ExpectedErr: "invalid XPath expression",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS id int PATH NULL);",
					ExpectedErr: "column filter expression must not be null",
				},
				{
					Query:       "SELECT * FROM XMLTABLE(NULL PASSING '<r/>' COLUMNS n xml PATH '.');",
					ExpectedErr: "row filter expression must not be null",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('' PASSING '<r/>' COLUMNS n xml PATH '.');",
					ExpectedErr: "row path filter must not be empty string",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/a' PASSING ('<a>x</a>'::text) COLUMNS v text PATH 'text()');",
					ExpectedErr: "argument of XMLTABLE must be type xml, not type text",
				},
				{
					Query:       `SELECT * FROM XMLTABLE(XMLNAMESPACES(DEFAULT 'urn:x'), '/r/row' PASSING ('<r xmlns="urn:x"><row><n>a</n></row></r>'::xml) COLUMNS n text PATH 'n');`,
					ExpectedErr: "DEFAULT namespace is not supported",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS id int PATH 'a' DEFAULT 1 DEFAULT 2);",
					ExpectedErr: "only one DEFAULT value is allowed",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS id int PATH 'n' PATH 'n');",
					ExpectedErr: "only one PATH value per column is allowed",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS id FOR ORDINALITY, id2 FOR ORDINALITY);",
					ExpectedErr: "only one FOR ORDINALITY column is allowed",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS n int, n int);",
					ExpectedErr: `column name "n" is not unique`,
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS n int PATH 'n' NOT NULL NULL);",
					ExpectedErr: `conflicting or redundant NULL / NOT NULL declarations for column "n"`,
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r' PASSING '<r/>');",
					ExpectedErr: "syntax error",
				},
				{
					Query:       "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>1</n></row></r>'::xml) COLUMNS n int) WITH ORDINALITY;",
					ExpectedErr: "syntax error",
				},
			},
		},
	})
}
