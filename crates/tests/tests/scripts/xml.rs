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
fn test_xml_functions() {
    run_scripts(&[
        ScriptTest {
            name: "xpath",
            set_up_script: &[
                "CREATE TABLE t_xml (id INTEGER PRIMARY KEY, v1 XML);",
                "INSERT INTO t_xml VALUES (1, '<note><to>Tove</to></note>'), (2, '<book><title>Introduction to Golang</title><author>John Doe</author></book>'), (3, NULL);",
                "CREATE TABLE t_text (id INTEGER PRIMARY KEY, doc TEXT);",
                "INSERT INTO t_text VALUES (1, '<a>x</a>'), (2, '<b>y</b>');",
                "CREATE VIEW v_xml AS SELECT id, xpath('/book/title/text()', v1) AS x FROM t_xml;",
                "CREATE VIEW v_text AS SELECT id, xpath('/a/text()', doc::xml) AS x FROM t_text;",
                "CREATE VIEW v_cast AS SELECT id, doc::xml AS d FROM t_text;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT xpath('/a/text()', '<a>x</a>'::xml);",
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{x}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath('/a/text()', '<a>x</a>'), pg_catalog.xpath('/a/text()', '<a>x</a>');",
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{x}"), T("{x}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(xpath('/a', '<a/>'::xml));",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("xml[]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('/a', '<a>x</a>'::xml), xpath('/a/b', '<a><b>1</b><b>2</b></a>'::xml), xpath('//b/text()', '<a><b>1</b><b>2</b></a>'::xml), xpath('/a/@id', '<a id="7">x</a>'::xml);"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{<a>x</a>}"), T("{<b>1</b>,<b>2</b>}"), T("{1,2}"), T("{7}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('/a/text()', '<a>&lt;&amp;&gt;"''</a>'::xml), xpath('/a/@id', '<a id="&lt;&amp;&gt;&quot;">x</a>'::xml);"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T(r#"{"&lt;&amp;&gt;\"'"}"#), T(r#"{"&lt;&amp;&gt;\""}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('/a', '<a b=''q"''>&amp;</a>'::xml), xpath('/a', '<a b="&lt;">&amp;</a>'::xml), xpath('/a', '<a>"</a>'::xml);"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T(r#"{"<a b=\"q&quot;\">&amp;</a>"}"#), T(r#"{"<a b=\"&lt;\">&amp;</a>"}"#), T(r#"{"<a>\"</a>"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('/a', '<a><b/><c></c></a>'::xml), xpath('/a', E'<a>\n  <b/>\n</a>'::xml);"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{<a><b/><c/></a>}"), T(r#"{"<a>
  <b/>
</a>"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath('/a', '<a><![CDATA[<x>]]></a>'::xml), xpath('/a/text()', '<a><![CDATA[<x>]]></a>'::xml), xpath('/a/comment()', '<a><!-- c --></a>'::xml), xpath('/a/node()', '<a>x<b>y</b>z</a>'::xml);",
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{<a><![CDATA[<x>]]></a>}"), T("{<![CDATA[<x>]]>}"), T(r#"{"<!-- c -->"}"#), T("{x,<b>y</b>,z}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath('/', '<a>x</a>'::xml), xpath('/zzz', '<a>x</a>'::xml), xpath('/a/b/text()', '<a>x</a>'::xml);",
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T(r#"{"<a>x</a>
"}"#), T("{}"), T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath('count(/a/b)', '<a><b/><b/></a>'::xml), xpath('count(/a/b) div 3', '<a><b/><b/></a>'::xml), xpath('1 div 0', '<a/>'::xml), xpath('-1 div 0', '<a/>'::xml), xpath('0 div 0', '<a/>'::xml), xpath('1.5', '<a/>'::xml);",
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{2}"), T("{0.6666666666666666}"), T("{Infinity}"), T("{-Infinity}"), T("{NaN}"), T("{1.5}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('1=1', '<a/>'::xml), xpath('1=2', '<a/>'::xml), xpath('string(/a)', '<a>x<b>y</b></a>'::xml), xpath('string(/a)', '<a>&lt;</a>'::xml), xpath('concat("a","<")', '<a/>'::xml);"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{true}"), T("{false}"), T("{xy}"), T("{&lt;}"), T("{a&lt;}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('/a', '<?xml version="1.0"?><a>x</a>'::xml), xpath('/a', E'  \n <a>x</a>'::xml), xpath('/a', '<!-- c --><a>x</a>'::xml);"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{<a>x</a>}"), T("{<a>x</a>}"), T("{<a>x</a>}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('//x:b/text()', '<a xmlns:x="urn:x"><x:b>1</x:b></a>'::xml, ARRAY[ARRAY['x','urn:x']]), xpath('//y:b/text()', '<a xmlns:x="urn:x"><x:b>1</x:b></a>'::xml, ARRAY[ARRAY['y','urn:x']]);"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{1}"), T("{1}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('//x:b', '<a xmlns:x="urn:x"><x:b>1</x:b></a>'::xml, ARRAY[ARRAY['x','urn:x']]), xpath('/d:a/d:b', '<a xmlns="urn:d"><b>1</b></a>'::xml, ARRAY[ARRAY['d','urn:d']]);"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T(r#"{"<x:b xmlns:x=\"urn:x\">1</x:b>"}"#), T(r#"{"<b xmlns=\"urn:d\">1</b>"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath('/a', '<a/>'::xml, '{}'::text[]), xpath('/a', '<a/>'::xml, ARRAY[ARRAY['x','urn:x'],ARRAY['y','urn:y']]);",
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{<a/>}"), T("{<a/>}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('//x:b/text()', '<a xmlns:x="urn:x"><x:b>1</x:b></a>'::xml, ARRAY['x','urn:x','z']);"#,
                    expected: Expected::Error(Diagnostic { code: "22000", message: "invalid array for XML namespace mapping", detail: "The array must be two-dimensional with length of the second axis equal to 2.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath(NULL, '<a/>'::xml), xpath('/a', NULL), xpath('/a', '<a/>'::xml, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[Null, Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath('', '<a>x</a>'::xml);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "empty XPath expression", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath('/a[', '<a>x</a>'::xml);",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "invalid XPath expression", detail: "Invalid expression", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath('/a', '<a/><b/>'::xml);",
                    expected: Expected::Error(Diagnostic { code: "2200M", message: "could not parse XML document", detail: r#"line 1: Extra content at the end of the document
<a/><b/>
    ^"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath('/a', 'plain'::xml);",
                    expected: Expected::Error(Diagnostic { code: "2200M", message: "could not parse XML document", detail: r#"line 1: Start tag expected, '<' not found
plain
^"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath('/a', ''::xml);",
                    expected: Expected::Error(Diagnostic { code: "2200M", message: "could not parse XML document", detail: r#"line 1: Document is empty

^"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, xpath('/book/title/text()', v1) FROM t_xml ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("1"), T("{}")],
                            &[T("2"), T(r#"{"Introduction to Golang"}"#)],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM v_xml ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("x", XML_ARRAY)],
                        rows: &[
                            &[T("1"), T("{}")],
                            &[T("2"), T(r#"{"Introduction to Golang"}"#)],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM v_text ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("x", XML_ARRAY)],
                        rows: &[
                            &[T("1"), T("{x}")],
                            &[T("2"), T("{}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, d, pg_typeof(d) FROM v_cast ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("d", XML), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("1"), T("<a>x</a>"), T("xml")],
                            &[T("2"), T("<b>y</b>"), T("xml")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('//x:b/text()', '<a xmlns:x="urn:x"><x:b>1</x:b></a>'::xml, ARRAY[ARRAY['x',NULL]]);"#,
                    expected: Expected::Error(Diagnostic { code: "22004", message: "neither namespace name nor URI may be null", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "xpath_exists",
            set_up_script: &[
                "CREATE TABLE t_xml (id INTEGER PRIMARY KEY, v1 XML);",
                "INSERT INTO t_xml VALUES (1, '<note><to>Tove</to></note>'), (2, '<book><title>Introduction to Golang</title></book>'), (3, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT xpath_exists('/a', '<a>x</a>'::xml), xpath_exists('/b', '<a>x</a>'::xml), xpath_exists('/a', '<a>x</a>');",
                    expected: Expected::Rows {
                        columns: &[Column("xpath_exists", BOOL), Column("xpath_exists", BOOL), Column("xpath_exists", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath_exists('count(/a)', '<a>x</a>'::xml), xpath_exists('1=2', '<a>x</a>'::xml), xpath_exists('string(/b)', '<a>x</a>'::xml), xpath_exists('0', '<a/>'::xml);",
                    expected: Expected::Rows {
                        columns: &[Column("xpath_exists", BOOL), Column("xpath_exists", BOOL), Column("xpath_exists", BOOL), Column("xpath_exists", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath_exists('//x:b', '<a xmlns:x="urn:x"><x:b>1</x:b></a>'::xml, ARRAY[ARRAY['x','urn:x']]), xpath_exists(NULL, '<a/>'::xml);"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath_exists", BOOL), Column("xpath_exists", BOOL)],
                        rows: &[
                            &[T("t"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath_exists('/a', '<a/><b/>'::xml);",
                    expected: Expected::Error(Diagnostic { code: "2200M", message: "could not parse XML document", detail: r#"line 1: Extra content at the end of the document
<a/><b/>
    ^"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath_exists('/a[', '<a/>'::xml);",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "invalid XPath expression", detail: "Invalid expression", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM t_xml WHERE xpath_exists('/book', v1);",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "xml_is_well_formed",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT xml_is_well_formed('<a/>'), xml_is_well_formed('<a>'), xml_is_well_formed('x'), xml_is_well_formed('<a/><b/>'), xml_is_well_formed('');",
                    expected: Expected::Rows {
                        columns: &[Column("xml_is_well_formed", BOOL), Column("xml_is_well_formed", BOOL), Column("xml_is_well_formed", BOOL), Column("xml_is_well_formed", BOOL), Column("xml_is_well_formed", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("t"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xml_is_well_formed_document('<a/>'), xml_is_well_formed_document('x'), xml_is_well_formed_document('<a/><b/>'), xml_is_well_formed_document('<?xml version="1.0"?><a/>'), xml_is_well_formed_document('<!-- c --><a/>'), xml_is_well_formed_document(' <a/>'), xml_is_well_formed_document('');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xml_is_well_formed_document", BOOL), Column("xml_is_well_formed_document", BOOL), Column("xml_is_well_formed_document", BOOL), Column("xml_is_well_formed_document", BOOL), Column("xml_is_well_formed_document", BOOL), Column("xml_is_well_formed_document", BOOL), Column("xml_is_well_formed_document", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("f"), T("t"), T("t"), T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xml_is_well_formed_content('<a/>'), xml_is_well_formed_content('x'), xml_is_well_formed_content('<a/><b/>'), xml_is_well_formed_content('<a>'), xml_is_well_formed_content('');",
                    expected: Expected::Rows {
                        columns: &[Column("xml_is_well_formed_content", BOOL), Column("xml_is_well_formed_content", BOOL), Column("xml_is_well_formed_content", BOOL), Column("xml_is_well_formed_content", BOOL), Column("xml_is_well_formed_content", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("t"), T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET xmloption TO document;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xml_is_well_formed('<a/><b/>'), xml_is_well_formed('<a/>');",
                    expected: Expected::Rows {
                        columns: &[Column("xml_is_well_formed", BOOL), Column("xml_is_well_formed", BOOL)],
                        rows: &[
                            &[T("f"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "XMLPARSE",
            set_up_script: &[
                "CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
                r#"INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT XMLPARSE(DOCUMENT '<a>x</a>');",
                    expected: Expected::Rows {
                        columns: &[Column("xmlparse", XML)],
                        rows: &[
                            &[T("<a>x</a>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT XMLPARSE(CONTENT 'ab<c/>'), XMLPARSE(CONTENT ''), XMLPARSE(CONTENT NULL), XMLPARSE(CONTENT 1);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlparse", XML), Column("xmlparse", XML), Column("xmlparse", XML), Column("xmlparse", XML)],
                        rows: &[
                            &[T("ab<c/>"), T(""), Null, T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT XMLPARSE(DOCUMENT '<a/>' PRESERVE WHITESPACE), XMLPARSE(DOCUMENT '<a> </a>' STRIP WHITESPACE);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlparse", XML), Column("xmlparse", XML)],
                        rows: &[
                            &[T("<a/>"), T("<a> </a>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT XMLPARSE(DOCUMENT '<?xml version="1.0"?><a/>'), XMLPARSE(DOCUMENT '<?xml version="1.1"?><a/>');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xmlparse", XML), Column("xmlparse", XML)],
                        rows: &[
                            &[T("<a/>"), T(r#"<?xml version="1.1"?><a/>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    notices: &[Diagnostic { severity: "WARNING", code: "01000", message: r#"line 1: Unsupported version '1.1'
<?xml version="1.1"?><a/>
                   ^"#, ..E }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(XMLPARSE(CONTENT '<a/>'));",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("xml")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT XMLPARSE(DOCUMENT 'a<b/>');",
                    expected: Expected::Error(Diagnostic { code: "2200M", message: "invalid XML document", detail: r#"line 1: Start tag expected, '<' not found
a<b/>
^"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT XMLPARSE(CONTENT '<a');",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Couldn't find end of Start Tag a line 1
<a
  ^"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT XMLPARSE(DOCUMENT doc::text) FROM docs WHERE id = 2;",
                    expected: Expected::Rows {
                        columns: &[Column("xmlparse", XML)],
                        rows: &[
                            &[T("<r/>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "xmlconcat",
            set_up_script: &[
                "CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
                r#"INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT xmlconcat('<a/>', '<b/>');",
                    expected: Expected::Rows {
                        columns: &[Column("xmlconcat", XML)],
                        rows: &[
                            &[T("<a/><b/>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlconcat('<a/>', NULL, '<b/>'), xmlconcat(NULL, NULL), xmlconcat('a', 'b'), xmlconcat('<a/>');",
                    expected: Expected::Rows {
                        columns: &[Column("xmlconcat", XML), Column("xmlconcat", XML), Column("xmlconcat", XML), Column("xmlconcat", XML)],
                        rows: &[
                            &[T("<a/><b/>"), Null, T("ab"), T("<a/>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xmlconcat('<?xml version="1.0" standalone="yes"?><a/>', '<?xml version="1.0" standalone="no"?><b/>');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xmlconcat", XML)],
                        rows: &[
                            &[T(r#"<?xml version="1.0" standalone="no"?><a/><b/>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xmlconcat('<?xml version="1.1"?><a/>', '<?xml version="1.0"?><b/>'), xmlconcat('<?xml version="1.1"?><a/>', '<?xml version="1.1"?><b/>');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xmlconcat", XML), Column("xmlconcat", XML)],
                        rows: &[
                            &[T("<a/><b/>"), T(r#"<?xml version="1.1"?><a/><b/>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xmlconcat('<?xml version="1.0" standalone="yes"?><a/>', '<b/>'), xmlconcat('<?xml version="1.1" standalone="yes"?><a/>', '<?xml version="1.0"?><b/>');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xmlconcat", XML), Column("xmlconcat", XML)],
                        rows: &[
                            &[T("<a/><b/>"), T("<a/><b/>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlconcat(doc, '<z/>') FROM docs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("xmlconcat", XML)],
                        rows: &[
                            &[T(r#"<r><i n="a">1</i><i n="b">2</i></r><z/>"#)],
                            &[T("<r/><z/>")],
                            &[T("<z/>")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlconcat('<a/>'::text, '<b/>');",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "argument of XMLCONCAT must be type xml, not type text", position: 18, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "xmlcomment",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT xmlcomment('hi'), xmlcomment(''), xmlcomment(NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlcomment", XML), Column("xmlcomment", XML), Column("xmlcomment", XML)],
                        rows: &[
                            &[T("<!--hi-->"), T("<!---->"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlcomment('a--b');",
                    expected: Expected::Error(Diagnostic { code: "2200S", message: "invalid XML comment", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlcomment('a-');",
                    expected: Expected::Error(Diagnostic { code: "2200S", message: "invalid XML comment", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "xmlelement",
            set_up_script: &[
                "CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
                r#"INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT xmlelement(NAME a, 'x'), xmlelement(NAME a), xmlelement(NAME a, NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlelement", XML), Column("xmlelement", XML), Column("xmlelement", XML)],
                        rows: &[
                            &[T("<a>x</a>"), T("<a/>"), T("<a/>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xmlelement(NAME a, 'x', 'y', 1, 2.5, true, NULL, '<b/>'::xml, '<c/>', 'q<&>"''');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xmlelement", XML)],
                        rows: &[
                            &[T(r#"<a>xy12.5true<b/>&lt;c/&gt;q&lt;&amp;&gt;"'</a>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xmlelement(NAME a, xmlattributes('1' AS x, 2 AS y, NULL AS z, true AS b, 'q<&>"''' AS e), 'body');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xmlelement", XML)],
                        rows: &[
                            &[T(r#"<a x="1" y="2" b="true" e="q&lt;&amp;&gt;&quot;'">body</a>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlelement(NAME a, xmlattributes('1' AS x), xmlelement(NAME b, 'inner'));",
                    expected: Expected::Rows {
                        columns: &[Column("xmlelement", XML)],
                        rows: &[
                            &[T(r#"<a x="1"><b>inner</b></a>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlelement(NAME a, xmlattributes(id AS x), 'y') FROM (SELECT 7 AS id) t;",
                    expected: Expected::Rows {
                        columns: &[Column("xmlelement", XML)],
                        rows: &[
                            &[T(r#"<a x="7">y</a>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xmlelement(NAME "a b", 'x'), xmlelement(NAME "A", 'x'), xmlelement(NAME "a:b", 'x'), xmlelement(NAME "1a", 'x'), xmlelement(NAME a, xmlattributes('v' AS "a b"));"#,
                    expected: Expected::Rows {
                        columns: &[Column("xmlelement", XML), Column("xmlelement", XML), Column("xmlelement", XML), Column("xmlelement", XML), Column("xmlelement", XML)],
                        rows: &[
                            &[T("<a_x0020_b>x</a_x0020_b>"), T("<A>x</A>"), T("<a:b>x</a:b>"), T("<_x0031_a>x</_x0031_a>"), T(r#"<a a_x0020_b="v"/>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlelement(NAME a, '2001-02-03'::date, '2001-02-03 04:05:06'::timestamp, '04:05:06'::time, 'ab'::bytea, 1.50::numeric, ARRAY[1,2]);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlelement", XML)],
                        rows: &[
                            &[T("<a>2001-02-032001-02-03T04:05:0604:05:06YWI=1.50<element>1</element><element>2</element></a>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlelement(NAME a, 1.0::float8, 'NaN'::float8, 2.5::float4);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlelement", XML)],
                        rows: &[
                            &[T("<a>1NaN2.5</a>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlelement(NAME a, '<b>'::text), xmlelement(NAME a, 'x'::char(3)), xmlelement(NAME a, xmlattributes('<b/>'::xml AS x));",
                    expected: Expected::Rows {
                        columns: &[Column("xmlelement", XML), Column("xmlelement", XML), Column("xmlelement", XML)],
                        rows: &[
                            &[T("<a>&lt;b&gt;</a>"), T("<a>x  </a>"), T(r#"<a x="&lt;b/&gt;"/>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xmlelement(NAME a, xmlattributes(E'a\nb' AS x, E'a\tb' AS y), E'c\nd');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xmlelement", XML)],
                        rows: &[
                            &[T(r#"<a x="a&#10;b" y="a&#9;b">c
d</a>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xmlelement(NAME a, '<?xml version="1.0"?><b/>'::xml, '<?xml version="1.1"?><c/>'::xml);"#,
                    expected: Expected::Rows {
                        columns: &[Column("xmlelement", XML)],
                        rows: &[
                            &[T(r#"<a><b/><?xml version="1.1"?><c/></a>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlelement(NAME d, xmlattributes(id), doc) FROM docs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("xmlelement", XML)],
                        rows: &[
                            &[T(r#"<d id="1"><r><i n="a">1</i><i n="b">2</i></r></d>"#)],
                            &[T(r#"<d id="2"><r/></d>"#)],
                            &[T(r#"<d id="3"/>"#)],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(xmlelement(NAME a));",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("xml")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET xmlbinary TO hex;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlelement(NAME a, 'ab'::bytea);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlelement", XML)],
                        rows: &[
                            &[T("<a>6162</a>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET xmlbinary TO base64;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlelement(NAME a, 'ab'::bytea);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlelement", XML)],
                        rows: &[
                            &[T("<a>YWI=</a>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlelement(NAME a, xmlattributes(1, 'x' AS q));",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "unnamed XML attribute value must be a column reference", position: 41, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlelement(NAME a, xmlattributes('1' AS x, '2' AS x));",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"XML attribute name "x" appears more than once"#, position: 51, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "xmlforest",
            set_up_script: &[
                "CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
                r#"INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT xmlforest(1 AS a, 'x' AS b, NULL AS c);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlforest", XML)],
                        rows: &[
                            &[T("<a>1</a><b>x</b>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xmlforest(NULL AS c), xmlforest(1 AS a, 2 AS a), xmlforest('a<b' AS a), xmlforest(1 AS "A b");"#,
                    expected: Expected::Rows {
                        columns: &[Column("xmlforest", XML), Column("xmlforest", XML), Column("xmlforest", XML), Column("xmlforest", XML)],
                        rows: &[
                            &[Null, T("<a>1</a><a>2</a>"), T("<a>a&lt;b</a>"), T("<A_x0020_b>1</A_x0020_b>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlforest('<a/>'::xml AS a), xmlforest(true AS a, '2020-01-01 10:00:00'::timestamp AS b, ARRAY[1,2] AS c);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlforest", XML), Column("xmlforest", XML)],
                        rows: &[
                            &[T("<a><a/></a>"), T("<a>true</a><b>2020-01-01T10:00:00</b><c><element>1</element><element>2</element></c>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlforest(id, doc) FROM docs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("xmlforest", XML)],
                        rows: &[
                            &[T(r#"<id>1</id><doc><r><i n="a">1</i><i n="b">2</i></r></doc>"#)],
                            &[T("<id>2</id><doc><r/></doc>")],
                            &[T("<id>3</id>")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlelement(NAME r, xmlforest(1 AS a, 2 AS b)), xmlforest(1 AS a) IS DOCUMENT;",
                    expected: Expected::Rows {
                        columns: &[Column("xmlelement", XML), Column("?column?", BOOL)],
                        rows: &[
                            &[T("<r><a>1</a><b>2</b></r>"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlforest(1);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "unnamed XML element value must be a column reference", position: 18, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "xmlpi",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT xmlpi(NAME php, 'echo 1;');",
                    expected: Expected::Rows {
                        columns: &[Column("xmlpi", XML)],
                        rows: &[
                            &[T("<?php echo 1;?>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlpi(NAME php), xmlpi(NAME php, '  x'), xmlpi(NAME php, ''), xmlpi(NAME php, 1), xmlpi(NAME php, NULL), xmlpi(NAME php, 'a?b>');",
                    expected: Expected::Rows {
                        columns: &[Column("xmlpi", XML), Column("xmlpi", XML), Column("xmlpi", XML), Column("xmlpi", XML), Column("xmlpi", XML), Column("xmlpi", XML)],
                        rows: &[
                            &[T("<?php?>"), T("<?php x?>"), T("<?php ?>"), T("<?php 1?>"), Null, T("<?php a?b>?>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xmlpi(NAME "a b", 'x'), xmlpi(NAME "a:b"), xmlelement(NAME a, xmlpi(NAME php, 'x'));"#,
                    expected: Expected::Rows {
                        columns: &[Column("xmlpi", XML), Column("xmlpi", XML), Column("xmlelement", XML)],
                        rows: &[
                            &[T("<?a_x0020_b x?>"), T("<?a:b?>"), T("<a><?php x?></a>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlpi(NAME php, 'a?>b');",
                    expected: Expected::Error(Diagnostic { code: "2200T", message: "invalid XML processing instruction", detail: r#"XML processing instruction cannot contain "?>"."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlpi(NAME xml, 'x');",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "invalid XML processing instruction", detail: r#"XML processing instruction target name cannot be "xml"."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xmlpi(NAME "xMl");"#,
                    expected: Expected::Error(Diagnostic { code: "42601", message: "invalid XML processing instruction", detail: r#"XML processing instruction target name cannot be "xMl"."#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "xmlroot",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT xmlroot('<a/>', VERSION '1.1');",
                    expected: Expected::Rows {
                        columns: &[Column("xmlroot", XML)],
                        rows: &[
                            &[T(r#"<?xml version="1.1"?><a/>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlroot('<a/>', VERSION NO VALUE), xmlroot('<a/>', VERSION '1.0', STANDALONE YES), xmlroot('<a/>', VERSION '1.0', STANDALONE NO), xmlroot('<a/>', VERSION '1.0', STANDALONE NO VALUE);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlroot", XML), Column("xmlroot", XML), Column("xmlroot", XML), Column("xmlroot", XML)],
                        rows: &[
                            &[T("<a/>"), T(r#"<?xml version="1.0" standalone="yes"?><a/>"#), T(r#"<?xml version="1.0" standalone="no"?><a/>"#), T("<a/>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xmlroot('<?xml version="1.1" standalone="yes"?><a/>', VERSION NO VALUE), xmlroot('<?xml version="1.1" standalone="yes"?><a/>', VERSION '1.0', STANDALONE NO VALUE), xmlroot('<?xml version="1.1" standalone="no"?><a/>', VERSION '1.1');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xmlroot", XML), Column("xmlroot", XML), Column("xmlroot", XML)],
                        rows: &[
                            &[T(r#"<?xml version="1.0" standalone="yes"?><a/>"#), T("<a/>"), T(r#"<?xml version="1.1" standalone="no"?><a/>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlroot('<a/>', VERSION NULL), xmlroot('<a/>', VERSION 1), xmlroot(NULL, VERSION '1.0'), xmlroot('a<b/>', VERSION '1.0'), xmlroot('<a/>', VERSION NO VALUE, STANDALONE YES);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlroot", XML), Column("xmlroot", XML), Column("xmlroot", XML), Column("xmlroot", XML), Column("xmlroot", XML)],
                        rows: &[
                            &[T("<a/>"), T(r#"<?xml version="1"?><a/>"#), Null, T("a<b/>"), T(r#"<?xml version="1.0" standalone="yes"?><a/>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(xmlroot('<a/>', VERSION '1.0'));",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("xml")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlroot('<a/>'::text, VERSION '1.0');",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "argument of XMLROOT must be type xml, not type text", position: 16, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "xmlexists",
            set_up_script: &[
                "CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
                r#"INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT xmlexists('//a' PASSING '<a/>');",
                    expected: Expected::Rows {
                        columns: &[Column("xmlexists", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlexists('//a' PASSING BY REF '<b/>'), xmlexists('//a' PASSING BY VALUE '<a/>' BY REF), pg_catalog.xmlexists('//a', '<a/>'::xml), xmlexists('//a' PASSING NULL), xmlexists('count(//a)' PASSING '<a/>');",
                    expected: Expected::Rows {
                        columns: &[Column("xmlexists", BOOL), Column("xmlexists", BOOL), Column("xmlexists", BOOL), Column("xmlexists", BOOL), Column("xmlexists", BOOL)],
                        rows: &[
                            &[T("f"), T("t"), T("t"), Null, T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT id FROM docs WHERE xmlexists('/r/i[@n="b"]' PASSING doc);"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlexists('' PASSING '<a/>');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "empty XPath expression", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlexists('//a' PASSING ('<a'));",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Couldn't find end of Start Tag a line 1
<a
  ^"#, position: 33, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "XMLSERIALIZE",
            set_up_script: &[
                "CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
                r#"INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT XMLSERIALIZE(DOCUMENT '<a/>' AS text);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlserialize", TEXT)],
                        rows: &[
                            &[T("<a/>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT XMLSERIALIZE(CONTENT '<a/>' AS char(10)), XMLSERIALIZE(CONTENT 'a<b/>' AS text), XMLSERIALIZE(CONTENT NULL AS text), XMLSERIALIZE(CONTENT '<a/>' AS name), XMLSERIALIZE(CONTENT '<a/>' AS varchar);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlserialize", BPCHAR), Column("xmlserialize", TEXT), Column("xmlserialize", TEXT), Column("xmlserialize", NAME), Column("xmlserialize", VARCHAR)],
                        rows: &[
                            &[T("<a/>      "), T("a<b/>"), Null, T("<a/>"), T("<a/>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT XMLSERIALIZE(CONTENT ('<?xml version="1.0"?><a/>'::xml) AS text), pg_typeof(XMLSERIALIZE(CONTENT '<a/>' AS varchar(20)));"#,
                    expected: Expected::Rows {
                        columns: &[Column("xmlserialize", TEXT), Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T(r#"<?xml version="1.0"?><a/>"#), T("character varying")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT XMLSERIALIZE(CONTENT doc AS text) FROM docs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("xmlserialize", TEXT)],
                        rows: &[
                            &[T(r#"<r><i n="a">1</i><i n="b">2</i></r>"#)],
                            &[T("<r/>")],
                            &[Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT XMLSERIALIZE(CONTENT '<a/>' AS varchar(2));",
                    expected: Expected::Error(Diagnostic { code: "22001", message: "value too long for type character varying(2)", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT XMLSERIALIZE(CONTENT '<a/>' AS int);",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast XMLSERIALIZE result to integer", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT XMLSERIALIZE(CONTENT '<a/>' AS bytea);",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast XMLSERIALIZE result to bytea", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT XMLSERIALIZE(CONTENT '<a/>' AS xml);",
                    expected: Expected::Error(Diagnostic { code: "42846", message: "cannot cast XMLSERIALIZE result to xml", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT XMLSERIALIZE(DOCUMENT 'a<b/>' AS text);",
                    expected: Expected::Error(Diagnostic { code: "2200L", message: "not an XML document", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT XMLSERIALIZE(CONTENT '<a' AS text);",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Couldn't find end of Start Tag a line 1
<a
  ^"#, position: 29, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT XMLSERIALIZE(CONTENT '<a/>'::text AS text);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "argument of XMLSERIALIZE must be type xml, not type text", position: 29, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "IS DOCUMENT",
            set_up_script: &[
                "CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
                r#"INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '<a/>' IS DOCUMENT, '<a/><b/>' IS DOCUMENT, '<a/>' IS NOT DOCUMENT, NULL IS DOCUMENT, 'x' IS DOCUMENT;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("f"), Null, T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM docs WHERE doc IS DOCUMENT ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a' IS DOCUMENT;",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Couldn't find end of Start Tag a line 1
<a
  ^"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a/>'::text IS DOCUMENT;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "argument of IS DOCUMENT must be type xml, not type text", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 1 IS DOCUMENT;",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "argument of IS DOCUMENT must be type xml, not type integer", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "xmlagg",
            set_up_script: &[
                "CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
                r#"INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT xmlagg(x) FROM (VALUES ('<a/>'::xml), ('<b/>'), (NULL)) AS t(x);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlagg", XML)],
                        rows: &[
                            &[T("<a/><b/>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlagg(x) FROM (VALUES (NULL::xml)) AS t(x);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlagg", XML)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlagg(x) FROM (VALUES ('<a/>'::xml)) AS t(x) WHERE false;",
                    expected: Expected::Rows {
                        columns: &[Column("xmlagg", XML)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_typeof(xmlagg(x)) FROM (VALUES ('<a/>'::xml)) AS t(x);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_typeof", REGTYPE)],
                        rows: &[
                            &[T("xml")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xmlagg(x) FROM (VALUES ('<?xml version="1.0" standalone="yes"?><a/>'::xml), ('<?xml version="1.0" standalone="no"?><b/>')) AS t(x);"#,
                    expected: Expected::Rows {
                        columns: &[Column("xmlagg", XML)],
                        rows: &[
                            &[T(r#"<?xml version="1.0" standalone="no"?><a/><b/>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT x, xmlagg(y) FROM (VALUES (1, '<a/>'::xml), (1, '<b/>'), (2, NULL), (3, '<c/>')) AS t(x, y) GROUP BY x ORDER BY x;",
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4), Column("xmlagg", XML)],
                        rows: &[
                            &[T("1"), T("<a/><b/>")],
                            &[T("2"), Null],
                            &[T("3"), T("<c/>")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlagg(y) OVER (ORDER BY x) FROM (VALUES (1, '<a/>'::xml), (2, '<b/>')) AS t(x, y);",
                    expected: Expected::Rows {
                        columns: &[Column("xmlagg", XML)],
                        rows: &[
                            &[T("<a/>")],
                            &[T("<a/><b/>")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmlagg(xmlelement(NAME d, xmlattributes(id))), xmlagg(doc) FROM docs;",
                    expected: Expected::Rows {
                        columns: &[Column("xmlagg", XML), Column("xmlagg", XML)],
                        rows: &[
                            &[T(r#"<d id="1"/><d id="2"/><d id="3"/>"#), T(r#"<r><i n="a">1</i><i n="b">2</i></r><r/>"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "XMLTABLE",
            set_up_script: &[
                "CREATE TABLE docs (id INT PRIMARY KEY, doc XML);",
                r#"INSERT INTO docs VALUES (1, '<r><i n="a">1</i><i n="b">2</i></r>'), (2, '<r/>'), (3, NULL);"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/a' PASSING '<a>x</a>' COLUMNS v text PATH 'text()');",
                    expected: Expected::Rows {
                        columns: &[Column("v", TEXT)],
                        rows: &[
                            &[T("x")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row id="1"><n>a</n></row></r>'::xml) COLUMNS id int PATH '@id', n text) AS t(x, y);"#,
                    expected: Expected::Rows {
                        columns: &[Column("x", INT4), Column("y", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row id="1"><n>a</n></row></r>'::xml) COLUMNS id int PATH '@id', n text) t;"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("n", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT t.n, t.id FROM XMLTABLE('/r/row' PASSING ('<r><row id="1"><n>a</n></row></r>'::xml) COLUMNS id int PATH '@id', n text) AS t;"#,
                    expected: Expected::Rows {
                        columns: &[Column("n", TEXT), Column("id", INT4)],
                        rows: &[
                            &[T("a"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xmltable.v FROM XMLTABLE('/a' PASSING '<a>x</a>' COLUMNS v text PATH 'text()');",
                    expected: Expected::Rows {
                        columns: &[Column("v", TEXT)],
                        rows: &[
                            &[T("x")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM XMLTABLE('/r/row' PASSING BY REF ('<r><row id="1"><n>a</n></row></r>'::xml) BY REF COLUMNS id int PATH '@id');"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM XMLTABLE('/r/row' PASSING BY VALUE ('<r><row id="1"><n>a</n></row></r>'::xml) COLUMNS id int PATH '@id');"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM XMLTABLE(XMLNAMESPACES('urn:x' AS x, 'urn:y' AS y), '/x:r/x:row' PASSING ('<r xmlns="urn:x"><row><n>a</n></row></r>'::xml) COLUMNS n text PATH 'x:n');"#,
                    expected: Expected::Rows {
                        columns: &[Column("n", TEXT)],
                        rows: &[
                            &[T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>a</n><n>b</n></row></r>'::xml) COLUMNS n xml PATH 'n', m xml PATH 'n/text()');",
                    expected: Expected::Rows {
                        columns: &[Column("n", XML), Column("m", XML)],
                        rows: &[
                            &[T("<n>a</n><n>b</n>"), T("ab")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>&lt;&amp;</n></row></r>'::xml) COLUMNS n text PATH 'n', m xml PATH 'n/text()');",
                    expected: Expected::Rows {
                        columns: &[Column("n", TEXT), Column("m", XML)],
                        rows: &[
                            &[T("<&"), T("&lt;&amp;")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>x</n></row></r>'::xml) COLUMNS n text PATH 'count(n)', c int PATH 'count(n)', b bool PATH 'n = "x"', s text PATH 'string(n)', k text PATH 'count(n) div 0', j text PATH '1.5 + 1', m text PATH '0 div 0', f text PATH 'false()');"#,
                    expected: Expected::Rows {
                        columns: &[Column("n", TEXT), Column("c", INT4), Column("b", BOOL), Column("s", TEXT), Column("k", TEXT), Column("j", TEXT), Column("m", TEXT), Column("f", TEXT)],
                        rows: &[
                            &[T("1"), T("1"), T("t"), T("x"), T("Infinity"), T("2.5"), T("NaN"), T("false")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>x</n></row></r>'::xml) COLUMNS n text PATH 'm', d text PATH 'm' DEFAULT 'd', o int PATH 'm' DEFAULT 5, p int DEFAULT 1 + 1);",
                    expected: Expected::Rows {
                        columns: &[Column("n", TEXT), Column("d", TEXT), Column("o", INT4), Column("p", INT4)],
                        rows: &[
                            &[Null, T("d"), T("5"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/none' PASSING ('<r><row><n>x</n></row></r>'::xml) COLUMNS n text);",
                    expected: Expected::Rows {
                        columns: &[Column("n", TEXT)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>x</n></row><row/></r>'::xml) COLUMNS n text PATH 'n', d text PATH 'n' DEFAULT 'dd', o FOR ORDINALITY);",
                    expected: Expected::Rows {
                        columns: &[Column("n", TEXT), Column("d", TEXT), Column("o", INT4)],
                        rows: &[
                            &[T("x"), T("x"), T("1")],
                            &[Null, T("dd"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM XMLTABLE('/r/row/@id' PASSING ('<r><row id="1"/><row id="2"/></r>'::xml) COLUMNS v text PATH '.');"#,
                    expected: Expected::Rows {
                        columns: &[Column("v", TEXT)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row id="1"/></r>'::xml) COLUMNS v text PATH '.', w xml PATH '.', "Id" int PATH '@id', i2 int PATH '@id' DEFAULT 9);"#,
                    expected: Expected::Rows {
                        columns: &[Column("v", TEXT), Column("w", XML), Column("Id", INT4), Column("i2", INT4)],
                        rows: &[
                            &[T(""), T(r#"<row id="1"/>"#), T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM XMLTABLE('/r' PASSING ('<r a="1" b="2"/>'::xml) COLUMNS a int, b int, c int PATH '@a + @b');"#,
                    expected: Expected::Rows {
                        columns: &[Column("a", INT4), Column("b", INT4), Column("c", INT4)],
                        rows: &[
                            &[Null, Null, T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING ('<r>1</r>'::xml) COLUMNS v text PATH 'text()', w text PATH '.', y int PATH '.');",
                    expected: Expected::Rows {
                        columns: &[Column("v", TEXT), Column("w", TEXT), Column("y", INT4)],
                        rows: &[
                            &[T("1"), T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING ('<r><!-- c --></r>'::xml) COLUMNS v text PATH 'comment()', w xml PATH 'comment()');",
                    expected: Expected::Rows {
                        columns: &[Column("v", TEXT), Column("w", XML)],
                        rows: &[
                            &[T(" c "), T("<!-- c -->")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING ('<r><![CDATA[<x>]]></r>'::xml) COLUMNS v text PATH 'text()', w xml PATH 'text()', u text PATH '.', z xml PATH '.');",
                    expected: Expected::Rows {
                        columns: &[Column("v", TEXT), Column("w", XML), Column("u", TEXT), Column("z", XML)],
                        rows: &[
                            &[T("<x>"), T("<![CDATA[<x>]]>"), T("<x>"), T("<r><![CDATA[<x>]]></r>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING ('<r>  <a/>  </r>'::xml) COLUMNS v text PATH '.', w xml PATH '.');",
                    expected: Expected::Rows {
                        columns: &[Column("v", TEXT), Column("w", XML)],
                        rows: &[
                            &[T("    "), T("<r>  <a/>  </r>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM XMLTABLE('/r' PASSING '<r x="a&lt;b&amp;c"><t>p&lt;q</t><!--c&lt;--></r>' COLUMNS v text PATH '@x', w xml PATH '@x', t text PATH 't/text()', tx xml PATH 't/text()', c xml PATH 'comment()', ct text PATH 'comment()');"#,
                    expected: Expected::Rows {
                        columns: &[Column("v", TEXT), Column("w", XML), Column("t", TEXT), Column("tx", XML), Column("c", XML), Column("ct", TEXT)],
                        rows: &[
                            &[T("a<b&c"), T("a&lt;b&amp;c"), T("p<q"), T("p&lt;q"), T("<!--c&lt;-->"), T("c&lt;")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM XMLTABLE('/r' PASSING '<r x="a&#10;b" />' COLUMNS v text PATH '@x', w xml PATH '@x');"#,
                    expected: Expected::Rows {
                        columns: &[Column("v", TEXT), Column("w", XML)],
                        rows: &[
                            &[T(r#"a
b"#), T(r#"a
b"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/row' PASSING '<r><row>a<n>1</n>b</row></r>' COLUMNS n xml PATH 'text()', m xml PATH 'node()');",
                    expected: Expected::Rows {
                        columns: &[Column("n", XML), Column("m", XML)],
                        rows: &[
                            &[T("ab"), T("a<n>1</n>b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM XMLTABLE('/r' PASSING '<?xml version="1.0"?><r/>' COLUMNS n xml PATH '.');"#,
                    expected: Expected::Rows {
                        columns: &[Column("n", XML)],
                        rows: &[
                            &[T("<r/>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('count(/r)' PASSING '<r/>' COLUMNS n xml PATH '.');",
                    expected: Expected::Rows {
                        columns: &[Column("n", XML)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS id int PATH 1, v int PATH 'concat("1", "")');"#,
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", INT4)],
                        rows: &[
                            &[T("1"), T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS v varchar(2) PATH 'string("ab")', w char(4) PATH 'string("ab")', d date PATH 'string("2001-02-03")', n numeric(5,2) PATH 'string("1.234")', a int[] PATH 'string("{1,2}")');"#,
                    expected: Expected::Rows {
                        columns: &[Column("v", VARCHAR), Column("w", BPCHAR), Column("d", DATE), Column("n", NUMERIC), Column("a", INT4_ARRAY)],
                        rows: &[
                            &[T("ab"), T("ab  "), T("2001-02-03"), T("1.23"), T("{1,2}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/row' PASSING '<r><row><n>1</n></row></r>' COLUMNS n int NOT NULL PATH 'n', m int DEFAULT 5 PATH 'n', o int PATH 'q' DEFAULT 5 NOT NULL, p int PATH 'q' NOT NULL DEFAULT 5, q int PATH 'a' NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4), Column("m", INT4), Column("o", INT4), Column("p", INT4), Column("q", INT4)],
                        rows: &[
                            &[T("1"), T("1"), T("5"), T("5"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/row' PASSING (NULL::xml) COLUMNS id int PATH '@id');",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING NULL COLUMNS n xml PATH '.');",
                    expected: Expected::Rows {
                        columns: &[Column("n", XML)],
                        rows: &[],
                        tag: "SELECT 0",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT x.* FROM (VALUES ('<r><row><n>1</n></row></r>'), ('<r><row><n>2</n></row><row><n>3</n></row></r>')) AS d(doc), XMLTABLE('/r/row' PASSING (d.doc::xml) COLUMNS n int) AS x ORDER BY n;",
                    expected: Expected::Rows {
                        columns: &[Column("n", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, x.* FROM docs, XMLTABLE('/r/i' PASSING doc COLUMNS n text PATH '@n', v int PATH 'text()') AS x ORDER BY id, n;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("n", TEXT), Column("v", INT4)],
                        rows: &[
                            &[T("1"), T("a"), T("1")],
                            &[T("1"), T("b"), T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, x.n FROM docs LEFT JOIN XMLTABLE('/r/i' PASSING doc COLUMNS n text PATH '@n') AS x ON TRUE ORDER BY id, n;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("n", TEXT)],
                        rows: &[
                            &[T("1"), T("a")],
                            &[T("1"), T("b")],
                            &[T("2"), Null],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>a</n><n>b</n></row></r>'::xml) COLUMNS n text PATH 'n');",
                    expected: Expected::Error(Diagnostic { code: "21000", message: "more than one value returned by column XPath expression", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/row' PASSING '<r><row>a<n>1</n>b</row></r>' COLUMNS n text PATH 'node()');",
                    expected: Expected::Error(Diagnostic { code: "21000", message: "more than one value returned by column XPath expression", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>x</n></row></r>'::xml) COLUMNS n int PATH 'n');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "x""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>x</n></row></r>'::xml) COLUMNS n text PATH 'm' NOT NULL);",
                    expected: Expected::Error(Diagnostic { code: "22004", message: r#"null is not allowed in column "n""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS v int PATH 'q' DEFAULT 'zz');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "zz""#, position: 75, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS v int PATH 'q' DEFAULT 'a'::text);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "argument of XMLTABLE must be type integer, not type text", position: 75, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/row' PASSING '<r>' COLUMNS id int PATH '@id');",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Premature end of data in tag r line 1
<r>
   ^"#, position: 41, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/row' PASSING 'x' COLUMNS id int PATH '@id');",
                    expected: Expected::Error(Diagnostic { code: "2200M", message: "could not parse XML document", detail: r#"line 1: Start tag expected, '<' not found
x
^"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING '<r/><s/>' COLUMNS n xml PATH '.');",
                    expected: Expected::Error(Diagnostic { code: "2200M", message: "could not parse XML document", detail: r#"line 1: Extra content at the end of the document
<r/><s/>
    ^"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING '' COLUMNS n xml PATH '.');",
                    expected: Expected::Error(Diagnostic { code: "2200M", message: "could not parse XML document", detail: r#"line 1: Document is empty

^"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/[' PASSING '<r/>' COLUMNS id int PATH '@id');",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "invalid XPath expression", detail: "Invalid expression", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS id int PATH '@[');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "invalid XPath expression", detail: "Invalid expression", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS id int PATH NULL);",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "column filter expression must not be null", detail: r#"Filter for column "id" is null."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE(NULL PASSING '<r/>' COLUMNS n xml PATH '.');",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "row filter expression must not be null", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('' PASSING '<r/>' COLUMNS n xml PATH '.');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "row path filter must not be empty string", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/a' PASSING ('<a>x</a>'::text) COLUMNS v text PATH 'text()');",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "argument of XMLTABLE must be type xml, not type text", position: 38, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT * FROM XMLTABLE(XMLNAMESPACES(DEFAULT 'urn:x'), '/r/row' PASSING ('<r xmlns="urn:x"><row><n>a</n></row></r>'::xml) COLUMNS n text PATH 'n');"#,
                    expected: Expected::Error(Diagnostic { code: "0A000", message: "DEFAULT namespace is not supported", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS id int PATH 'a' DEFAULT 1 DEFAULT 2);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "only one DEFAULT value is allowed", position: 78, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS id int PATH 'n' PATH 'n');",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "only one PATH value per column is allowed", position: 68, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS id FOR ORDINALITY, id2 FOR ORDINALITY);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: "only one FOR ORDINALITY column is allowed", position: 71, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS n int, n int);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"column name "n" is not unique"#, position: 59, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS n int PATH 'n' NOT NULL NULL);",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"conflicting or redundant NULL / NOT NULL declarations for column "n""#, position: 76, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING '<r/>');",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near ")""#, position: 43, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r/row' PASSING ('<r><row><n>1</n></row></r>'::xml) COLUMNS n int) WITH ORDINALITY;",
                    expected: Expected::Error(Diagnostic { code: "42601", message: r#"syntax error at or near "WITH""#, position: 92, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE(XMLNAMESPACES(NULL AS x), '/x:r' PASSING '<r/>' COLUMNS n text PATH '.');",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "namespace URI must not be null", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM XMLTABLE('/r' PASSING '<r/>' COLUMNS n text PATH '');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "column path filter must not be empty string", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "Large documents",
            set_up_script: &[
                "CREATE TABLE big_docs (id INT PRIMARY KEY, doc XML, src TEXT);",
                r#"INSERT INTO big_docs VALUES (1, ('<r>' || repeat('<i n="a">1</i>', 2000) || '</r>')::xml, '<r>' || repeat('<i n="b">2</i>', 2000) || '</r>'), (2, repeat('<i>x</i>', 3000)::xml, repeat('c', 20000));"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, length(doc::text), md5(doc::text), length(src), md5(src) FROM big_docs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("length", INT4), Column("md5", TEXT), Column("length", INT4), Column("md5", TEXT)],
                        rows: &[
                            &[T("1"), T("28007"), T("5926b51458cff20722ee4a900d911e63"), T("28007"), T("ec757e5d3da6bcdbad50e56e8905a8a8")],
                            &[T("2"), T("24000"), T("3910aa29146c00ddc042df788598acc3"), T("20000"), T("3fb3fedd0620db447a450f9b4cda4bdd")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_length(xpath('/r/i/text()', doc), 1), (xpath('count(//i)', doc))[1]::text, (xpath('/r/i[2000]/@n', doc))[1]::text FROM big_docs WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("array_length", INT4), Column("xpath", TEXT), Column("xpath", TEXT)],
                        rows: &[
                            &[T("2000"), T("2000"), T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath_exists('/r/i[2000]', doc), xpath_exists('/r/i[2001]', doc), xmlexists('/r/i[@n="a"]' PASSING doc), xmlexists('/r/i[@n="b"]' PASSING BY REF doc) FROM big_docs WHERE id = 1;"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath_exists", BOOL), Column("xpath_exists", BOOL), Column("xmlexists", BOOL), Column("xmlexists", BOOL)],
                        rows: &[
                            &[T("t"), T("f"), T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, xml_is_well_formed(src), xml_is_well_formed_document(src), xml_is_well_formed_content(src), xml_is_well_formed_document(src || '<') FROM big_docs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("xml_is_well_formed", BOOL), Column("xml_is_well_formed_document", BOOL), Column("xml_is_well_formed_content", BOOL), Column("xml_is_well_formed_document", BOOL)],
                        rows: &[
                            &[T("1"), T("t"), T("t"), T("t"), T("f")],
                            &[T("2"), T("t"), T("f"), T("t"), T("f")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(XMLPARSE(DOCUMENT src)::text), md5(XMLPARSE(CONTENT src)::text) FROM big_docs WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("md5", TEXT)],
                        rows: &[
                            &[T("28007"), T("ec757e5d3da6bcdbad50e56e8905a8a8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(xmlconcat(doc, doc)::text), length(xmlconcat(doc, XMLPARSE(CONTENT src))::text) FROM big_docs WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("length", INT4)],
                        rows: &[
                            &[T("56014"), T("56014")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, length(xmlcomment(src)::text) FROM big_docs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("length", INT4)],
                        rows: &[
                            &[T("1"), T("28014")],
                            &[T("2"), T("20007")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(xmlelement(NAME w, doc)::text), md5(xmlelement(NAME w, xmlattributes(src AS s), src)::text) FROM big_docs WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("md5", TEXT)],
                        rows: &[
                            &[T("28014"), T("36443f21d208f1bec96b8fb283ed5bc4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(xmlforest(doc AS d, src AS s)::text) FROM big_docs WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("80040")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, length(xmlpi(NAME p, src)::text) FROM big_docs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("length", INT4)],
                        rows: &[
                            &[T("1"), T("28013")],
                            &[T("2"), T("20006")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(xmlroot(doc, VERSION '1.0', STANDALONE YES)::text), md5(xmlroot(doc, VERSION '1.0', STANDALONE YES)::text) FROM big_docs WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4), Column("md5", TEXT)],
                        rows: &[
                            &[T("28045"), T("efecfb6b601cf31ddb3c6d02ae1b3112")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, XMLSERIALIZE(CONTENT doc AS text) = doc::text, length(XMLSERIALIZE(CONTENT doc AS varchar)) FROM big_docs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("?column?", BOOL), Column("length", INT4)],
                        rows: &[
                            &[T("1"), T("t"), T("28007")],
                            &[T("2"), T("t"), T("24000")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, doc IS DOCUMENT, doc IS NOT DOCUMENT FROM big_docs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("1"), T("t"), T("f")],
                            &[T("2"), T("f"), T("t")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(xmlagg(doc)::text) FROM big_docs;",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("52007")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*), sum(x.v), max(x.o), min(x.n) FROM big_docs, XMLTABLE('/r/i' PASSING doc COLUMNS o FOR ORDINALITY, v int PATH '.', n text PATH '@n') AS x WHERE big_docs.id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("sum", INT8), Column("max", INT4), Column("min", TEXT)],
                        rows: &[
                            &[T("2000"), T("2000"), T("2000"), T("a")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT length(x.d::text) FROM big_docs, XMLTABLE('/r' PASSING doc COLUMNS d xml PATH '.') AS x WHERE big_docs.id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("length", INT4)],
                        rows: &[
                            &[T("28007")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, length(doc::varchar), length(src::xml::text) FROM big_docs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("length", INT4), Column("length", INT4)],
                        rows: &[
                            &[T("1"), T("28007"), T("28007")],
                            &[T("2"), T("24000"), T("20000")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE big_docs SET doc = xmlconcat(doc, '<z/>') WHERE id = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, length(doc::text), doc IS DOCUMENT FROM big_docs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("length", INT4), Column("?column?", BOOL)],
                        rows: &[
                            &[T("1"), T("28011"), T("f")],
                            &[T("2"), T("24000"), T("f")],
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
fn test_xml_rules() {
    run_scripts(&[
        ScriptTest {
            name: "xml parse errors",
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '<a></b>'::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Opening and ending tag mismatch: a line 1 and b
<a></b>
       ^"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '<a b="1" b="2"/>'::xml;"#,
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Attribute b redefined
<a b="1" b="2"/>
              ^"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a b=1/>'::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: AttValue: " or ' expected
<a b=1/>
     ^
line 1: attributes construct error
<a b=1/>
     ^
line 1: Couldn't find end of Start Tag a line 1
<a b=1/>
     ^"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '<a b="1"c="2"/>'::xml;"#,
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: attributes construct error
<a b="1"c="2"/>
        ^
line 1: Couldn't find end of Start Tag a line 1
<a b="1"c="2"/>
        ^"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a b/>'::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Specification mandates value for attribute b
<a b/>
    ^"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a><!-- x'::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Comment not terminated
<a><!-- x
         ^
line 1: Premature end of data in tag a line 1
<a><!-- x
         ^"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a><![CDATA[x'::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Unregistered error message
<a><![CDATA[x
             ^
line 1: Premature end of data in tag a line 1
<a><![CDATA[x
             ^"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a>]]></a>'::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Sequence ']]>' not allowed in content
<a>]]></a>
   ^"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a>&#0;</a>'::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: xmlParseCharRef: invalid xmlChar value 0
<a>&#0;</a>
       ^"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a>&amp</a>'::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: EntityRef: expecting ';'
<a>&amp</a>
       ^"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT E'<a>\n<b>\n</a>'::xml;"#,
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 3: Opening and ending tag mismatch: b line 2 and a
</a>
    ^
line 3: Premature end of data in tag a line 1
</a>
    ^"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '<a b="<"/>'::xml;"#,
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: Unescaped '<' not allowed in attributes values
<a b="<"/>
      ^
line 1: attributes construct error
<a b="<"/>
      ^
line 1: Couldn't find end of Start Tag a line 1
<a b="<"/>
      ^
line 1: StartTag: invalid element name
<a b="<"/>
       ^"#, position: 8, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<?pi x?><a/>'::xml, '<a><?xml x?></a>'::xml;",
                    expected: Expected::Error(Diagnostic { code: "2200N", message: "invalid XML content", detail: r#"line 1: XML declaration allowed only at the start of the document
<a><?xml x?></a>
        ^"#, position: 29, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<!DOCTYPE a><a/>'::xml;",
                    expected: Expected::Rows {
                        columns: &[Column("xml", XML)],
                        rows: &[
                            &[T("<!DOCTYPE a><a/>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '<a>&#65;&#x42;&lt;&gt;&amp;&quot;&apos;</a>'::xml;",
                    expected: Expected::Rows {
                        columns: &[Column("xml", XML)],
                        rows: &[
                            &[T("<a>&#65;&#x42;&lt;&gt;&amp;&quot;&apos;</a>")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xml_is_well_formed('<a xmlns:x="u"><x:b/></a>'), xml_is_well_formed('<a><b></a></b>');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xml_is_well_formed", BOOL), Column("xml_is_well_formed", BOOL)],
                        rows: &[
                            &[T("t"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "xml serialization through xpath",
            assertions: &[
                ScriptTestAssertion {
                    query: r#"SELECT xpath('/a/*', '<a><b x="1" y="2">t</b><c/><!--k--><?p d?></a>');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T(r#"{"<b x=\"1\" y=\"2\">t</b>",<c/>}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('//@*', '<a x="1"><b y="&amp;"/></a>');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{1,&amp;}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('/a/b', E'<a><b>x\r\ny</b></a>');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T(r#"{"<b>x
y</b>"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath('/a/b/text()', '<a><b>&#13;</b></a>');",
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{&#x0d;}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('//c', '<a xmlns:p="urn:p" xmlns="urn:d"><b><p:c/></b></a>', ARRAY[ARRAY['p','urn:p']]);"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('//p:c', '<a xmlns:p="urn:p"><b><p:c p:z="1"><d/></p:c></b></a>', ARRAY[ARRAY['p','urn:p']]);"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T(r#"{"<p:c xmlns:p=\"urn:p\" p:z=\"1\"><d/></p:c>"}"#)],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath('/a/b[2]', '<a><b>1</b><b>2</b><b>3</b></a>'), xpath('/a/b[last()]', '<a><b>1</b><b>2</b><b>3</b></a>');",
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{<b>2</b>}"), T("{<b>3</b>}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('/a/b[@x > 1]/@x', '<a><b x="1"/><b x="2"/><b x="3"/></a>');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('sum(/a/b) + string-length("abc") * 2', '<a><b>1</b><b>2</b></a>');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{9}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('normalize-space(" a  b ")', '<a/>'), xpath('translate("abc", "ab", "B")', '<a/>'), xpath('substring("12345", 2, 3)', '<a/>');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T(r#"{"a b"}"#), T("{Bc}"), T("{234}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('name(/a/*[1])', '<a><x:b xmlns:x="u"/></a>'), xpath('local-name(/a/*[1])', '<a><x:b xmlns:x="u"/></a>');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{x:b}"), T("{b}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath('/a/b/ancestor::*', '<a><b/></a>'), xpath('/a/b/following-sibling::*', '<a><b/><c/><d/></a>'), xpath('/a/d/preceding-sibling::*[1]', '<a><b/><c/><d/></a>');",
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{<a><b/></a>}"), T("{<c/>,<d/>}"), T("{<c/>}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT xpath('count(//b | //c)', '<a><b/><c/><b/></a>'), xpath('/a/b = "x"', '<a><b>y</b><b>x</b></a>'), xpath('not(/a/z)', '<a/>');"#,
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{3}"), T("{true}"), T("{true}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT xpath('1 mod 0.3', '<a/>'), xpath('-(2)', '<a/>'), xpath('10 div 4', '<a/>'), xpath('round(2.5)', '<a/>'), xpath('floor(-1.5)', '<a/>');",
                    expected: Expected::Rows {
                        columns: &[Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY), Column("xpath", XML_ARRAY)],
                        rows: &[
                            &[T("{0.10000000000000003}"), T("{-2}"), T("{2.5}"), T("{3}"), T("{-2}")],
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
