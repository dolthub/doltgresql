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

//! The xml type: checking, printing, and querying XML as Postgres does with libxml2.

pub mod parse;
pub mod sql;
pub mod table;
pub mod xpath;

use std::cell::{Cell, RefCell};
use std::collections::HashMap;

use crate::error::{PgError, Result, code};

thread_local! {
    /// OPTIONS are the running session's xmloption and xmlbinary settings: whether implicit parsing reads documents,
    /// and whether binary values write as hexadecimal.
    static OPTIONS: Cell<(bool, bool)> = const { Cell::new((false, false)) };
    /// CLIENT_ENCODING is the name of the client's encoding when it is not UTF8, which binary xml values declare.
    static CLIENT_ENCODING: Cell<Option<&'static str>> = const { Cell::new(None) };
    /// WARNINGS are the warnings that parsing raised in the running statement, which become notices.
    static WARNINGS: RefCell<Vec<String>> = const { RefCell::new(Vec::new()) };
}

/// install_options makes the xmloption, xmlbinary, and client_encoding settings the ones that xml values on this
/// thread follow.
pub fn install_options(xmloption: &str, xmlbinary: &str, client_encoding: &str) {
    OPTIONS.with(|o| o.set((xmloption.eq_ignore_ascii_case("document"), xmlbinary.eq_ignore_ascii_case("hex"))));
    let encoding = crate::encodings::Encoding::lookup(client_encoding).filter(|e| *e != crate::encodings::UTF8);
    CLIENT_ENCODING.with(|c| c.set(encoding.map(|e| e.name())));
}

/// send returns the binary format of an xml value, which declares the client's encoding when it is not UTF8, as
/// xml_send does.
pub fn send(text: &str) -> String {
    let Some(encoding) = CLIENT_ENCODING.with(Cell::get) else { return output(text) };
    let (declaration, rest) = parse::split_declaration(text);
    let declaration = declaration.unwrap_or_default();
    let standalone = match declaration.standalone {
        Some(true) => " standalone=\"yes\"",
        Some(false) => " standalone=\"no\"",
        None => "",
    };
    format!(
        "<?xml version=\"{}\" encoding=\"{encoding}\"{standalone}?>{rest}",
        declaration.version.as_deref().unwrap_or("1.0")
    )
}

/// document_option reports whether implicit parsing reads documents rather than content.
pub fn document_option() -> bool {
    OPTIONS.with(|o| o.get().0)
}

/// hex_binary reports whether binary values write as hexadecimal rather than base64.
pub fn hex_binary() -> bool {
    OPTIONS.with(|o| o.get().1)
}

/// warn records warnings that parsing raised.
pub fn warn(warnings: Vec<String>) {
    WARNINGS.with(|w| w.borrow_mut().extend(warnings));
}

/// take_warnings returns and clears the warnings that parsing raised.
pub fn take_warnings() -> Vec<String> {
    WARNINGS.with(|w| std::mem::take(&mut *w.borrow_mut()))
}

/// with_detail returns an error with a detail.
fn with_detail(code: &'static str, message: &str, detail: String) -> PgError {
    PgError { detail: Some(detail), ..PgError::new(code, message) }
}

/// check checks that text is well-formed XML content, or a well-formed document when `document` is set, as xml_in
/// does, returning the warnings libxml2 printed.
pub fn check(text: &str, document: bool) -> Result<Vec<String>> {
    if document {
        if text.is_empty() {
            let detail = format!(
                "{}\n{}",
                parse::message(text, 0, "switching encoding : no input"),
                parse::message(text, 0, "Document is empty")
            );
            return Err(with_detail(code::INVALID_XML_DOCUMENT, "invalid XML document", detail));
        }
        return parse::parse_document(text)
            .map(|parsed| parsed.warnings)
            .map_err(|err| with_detail(code::INVALID_XML_DOCUMENT, "invalid XML document", err.detail));
    }
    let (_, rest) = parse::split_declaration(text);
    parse::parse_content(rest)
        .map(|parsed| parsed.warnings)
        .map_err(|err| with_detail(code::INVALID_XML_CONTENT, "invalid XML content", err.detail))
}

/// is_document reports whether text is a well-formed XML document.
pub fn is_document(text: &str) -> bool {
    !text.is_empty() && parse::parse_document(text).is_ok()
}

/// is_well_formed reports whether text is well-formed XML content, or a well-formed document when `document` is set.
pub fn is_well_formed(text: &str, document: bool) -> bool {
    check(text, document).is_ok()
}

/// declaration_text writes the XML declaration that Postgres prints for a version and standalone value, which is
/// empty for version 1.0 without a standalone value.
fn declaration_text(version: Option<&str>, standalone: Option<bool>) -> String {
    let version = version.unwrap_or("1.0");
    match standalone {
        None if version == "1.0" => String::new(),
        None => format!("<?xml version=\"{version}\"?>"),
        Some(standalone) => {
            format!("<?xml version=\"{version}\" standalone=\"{}\"?>", if standalone { "yes" } else { "no" })
        }
    }
}

/// output returns the text form of an xml value, dropping or rewriting its declaration as xml_out does.
pub fn output(text: &str) -> String {
    let (declaration, rest) = parse::split_declaration(text);
    let Some(declaration) = declaration else { return text.to_string() };
    let written = declaration_text(declaration.version.as_deref(), declaration.standalone);
    if written.is_empty() { rest.strip_prefix('\n').unwrap_or(rest).to_string() } else { written + rest }
}

/// concat concatenates xml values, merging their declarations as xmlconcat does.
pub fn concat(values: &[&str]) -> String {
    let mut out = String::new();
    let mut version: Option<String> = None;
    let mut versions_differ = false;
    let mut standalone = Some(true);
    for value in values {
        let (declaration, rest) = parse::split_declaration(value);
        let declaration = declaration.unwrap_or_default();
        match declaration.standalone {
            None => standalone = None,
            Some(false) if standalone == Some(true) => standalone = Some(false),
            _ => {}
        }
        match (&declaration.version, &version) {
            (None, _) => versions_differ = true,
            (Some(v), Some(global)) if v != global => versions_differ = true,
            (Some(v), _) => version = Some(v.clone()),
        }
        out.push_str(rest);
    }
    if versions_differ {
        version = None;
        if standalone.is_none() {
            return out;
        }
    }
    declaration_text(version.as_deref(), standalone) + &out
}

/// root sets the version and standalone value of an xml value's declaration, as xmlroot does.
pub fn root(text: &str, version: Option<&str>, standalone: Option<Option<bool>>) -> String {
    let (declaration, rest) = parse::split_declaration(text);
    let mut standalone_value = declaration.and_then(|d| d.standalone);
    if let Some(value) = standalone {
        standalone_value = value;
    }
    let version = version.unwrap_or("1.0");
    let written = match standalone_value {
        None if version == "1.0" => String::new(),
        None => format!("<?xml version=\"{version}\"?>"),
        Some(s) => format!("<?xml version=\"{version}\" standalone=\"{}\"?>", if s { "yes" } else { "no" }),
    };
    written + rest
}

/// escape escapes text for XML as Postgres' escape_xml does.
pub fn escape(text: &str) -> String {
    let mut out = String::with_capacity(text.len());
    for c in text.chars() {
        match c {
            '&' => out.push_str("&amp;"),
            '<' => out.push_str("&lt;"),
            '>' => out.push_str("&gt;"),
            '\r' => out.push_str("&#x0d;"),
            c => out.push(c),
        }
    }
    out
}

/// Query is a compiled XPath expression with the namespaces its prefixes name.
pub struct Query {
    pub expr: xpath::Expr,
    pub namespaces: HashMap<String, String>,
}

/// parse_for_xpath parses an xml value as a document for XPath, as Postgres' xpath does, failing with a code.
pub fn parse_for_xpath(text: &str) -> Result<xpath::Document> {
    let (_, rest) = parse::split_declaration(text);
    match parse::parse_document(rest) {
        Ok(parsed) => Ok(xpath::Document::new(parsed.nodes)),
        Err(err) => Err(with_detail(code::INVALID_XML_DOCUMENT, "could not parse XML document", err.detail)),
    }
}

/// compile compiles an XPath expression, failing as Postgres does for one that is empty or invalid.
pub fn compile(text: &str) -> Result<xpath::Expr> {
    if text.is_empty() {
        return Err(PgError::new(code::INVALID_ARGUMENT_FOR_XQUERY, "empty XPath expression"));
    }
    xpath::compile(text)
        .map_err(|err| with_detail(code::INVALID_ARGUMENT_FOR_XQUERY, "invalid XPath expression", err.0))
}

/// node_text writes a node of an XPath result as an xml value, as Postgres' xml_xmlnodetoxmltype does.
pub fn node_text(document: &xpath::Document, node: usize) -> String {
    match &document.nodes[node].kind {
        xpath::Kind::Text(_) | xpath::Kind::Attribute(_) => escape(&document.string_value(node)),
        _ => document.serialize_copy(node),
    }
}

/// evaluate evaluates a query at the root of a document, failing as Postgres does when evaluation fails.
pub fn evaluate(query: &Query, document: &xpath::Document) -> Result<xpath::Value> {
    xpath::evaluate(&query.expr, document, 0, &query.namespaces)
        .map_err(|err| with_detail(code::INVALID_ARGUMENT_FOR_XQUERY, "could not create XPath object", err.0))
}

/// xpath evaluates an XPath expression over an xml value, returning the xml values of its result, as Postgres' xpath
/// does.
pub fn xpath(query: &Query, text: &str) -> Result<Vec<String>> {
    let document = parse_for_xpath(text)?;
    Ok(match evaluate(query, &document)? {
        xpath::Value::Nodes(nodes) => nodes.iter().map(|&n| node_text(&document, n)).collect(),
        other => vec![escape(&other.to_text(&document))],
    })
}

/// xpath_exists reports whether an XPath expression over an xml value finds anything, as Postgres' xpath_exists
/// does.
pub fn xpath_exists(query: &Query, text: &str) -> Result<bool> {
    let document = parse_for_xpath(text)?;
    Ok(match evaluate(query, &document)? {
        xpath::Value::Nodes(nodes) => !nodes.is_empty(),
        _ => true,
    })
}
