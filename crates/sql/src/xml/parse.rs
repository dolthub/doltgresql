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

//! An XML parser that builds a tree and reports errors where and as libxml2 reports them, which Postgres' messages
//! quote.

/// Node is a node of a parsed XML tree.
#[derive(Clone, Debug, PartialEq)]
pub enum Node {
    Element(Box<Element>),
    Text(String),
    CData(String),
    Comment(String),
    /// A processing instruction's target and data.
    Pi(String, String),
}

/// Name is an element or attribute name with the namespace it is in.
#[derive(Clone, Debug, PartialEq)]
pub struct Name {
    pub prefix: Option<String>,
    pub local: String,
    /// The namespace URI the name is in, or None for no namespace.
    pub uri: Option<String>,
}

impl Name {
    /// qualified returns the name as written, with its prefix.
    pub fn qualified(&self) -> String {
        match &self.prefix {
            Some(prefix) => format!("{prefix}:{}", self.local),
            None => self.local.clone(),
        }
    }
}

/// Attribute is an attribute of an element, with its value after references are replaced.
#[derive(Clone, Debug, PartialEq)]
pub struct Attribute {
    pub name: Name,
    pub value: String,
}

/// Element is an element with the namespaces it declares, its attributes, and its children.
#[derive(Clone, Debug, PartialEq)]
pub struct Element {
    pub name: Name,
    /// The namespaces the element declares, as a prefix (None for the default namespace) and a URI.
    pub namespaces: Vec<(Option<String>, String)>,
    pub attributes: Vec<Attribute>,
    pub children: Vec<Node>,
}

/// Declaration is an XML declaration's version and standalone values.
#[derive(Clone, Debug, Default, PartialEq)]
pub struct Declaration {
    pub version: Option<String>,
    pub standalone: Option<bool>,
}

/// Parsed is the result of parsing: the top-level nodes, and the warnings libxml2 would have printed.
#[derive(Debug)]
pub struct Parsed {
    pub nodes: Vec<Node>,
    pub warnings: Vec<String>,
}

/// Error is a parse error with libxml2's message, line, and context.
#[derive(Debug)]
pub struct Error {
    /// The messages, each with its line number, the line of input, and a pointer to the position.
    pub detail: String,
}

/// Parser reads XML text.
struct Parser<'a> {
    text: &'a str,
    bytes: &'a [u8],
    at: usize,
    /// The in-scope namespace declarations, innermost last.
    scopes: Vec<(Option<String>, String)>,
    warnings: Vec<String>,
    /// The errors found so far, which libxml2 keeps adding to after the errors it recovers from.
    errors: Vec<String>,
}

/// is_name_start reports whether a character can start an XML name.
fn is_name_start(c: char) -> bool {
    c.is_ascii_alphabetic() || c == '_' || c == ':' || (!c.is_ascii() && c.is_alphabetic())
}

/// is_name_char reports whether a character can continue an XML name.
fn is_name_char(c: char) -> bool {
    is_name_start(c)
        || c.is_ascii_digit()
        || c == '-'
        || c == '.'
        || c == '\u{b7}'
        || (!c.is_ascii() && c.is_alphanumeric())
}

/// is_blank reports whether a byte is XML white space.
fn is_blank(b: u8) -> bool {
    matches!(b, b' ' | b'\t' | b'\n' | b'\r')
}

/// context writes the line of the input around a position and a pointer under the position, as libxml2's
/// xmlParserPrintFileContext does.
pub fn context(input: &str, position: usize) -> String {
    let bytes = input.as_bytes();
    let position = position.min(bytes.len());
    let mut start = position;
    while start > 0 && start < bytes.len() && matches!(bytes[start], b'\n' | b'\r') {
        start -= 1;
    }
    let mut n = 0;
    while n < 80 && start > 0 && !matches!(bytes.get(start), Some(b'\n' | b'\r')) {
        start -= 1;
        n += 1;
    }
    if matches!(bytes.get(start), Some(b'\n' | b'\r')) {
        start += 1;
    }
    let column = position.saturating_sub(start);
    let mut end = start;
    while end < bytes.len() && !matches!(bytes[end], b'\n' | b'\r') && end - start < 80 {
        end += 1;
    }
    let line = String::from_utf8_lossy(&bytes[start..end]).into_owned();
    let pointer: String = line.bytes().take(column).map(|b| if b == b'\t' { '\t' } else { ' ' }).collect();
    format!("{line}\n{pointer}^")
}

/// line_of returns the 1-based line of a position.
fn line_of(input: &str, position: usize) -> usize {
    input.as_bytes()[..position.min(input.len())].iter().filter(|&&b| b == b'\n').count() + 1
}

/// message writes a libxml2 message at a position as Postgres quotes it: the line, the message, and the context.
pub fn message(input: &str, position: usize, text: &str) -> String {
    format!("line {}: {text}\n{}", line_of(input, position), context(input, position))
}

impl<'a> Parser<'a> {
    /// new returns a parser of text.
    fn new(text: &'a str) -> Parser<'a> {
        Parser {
            text,
            bytes: text.as_bytes(),
            at: 0,
            scopes: vec![(Some("xml".into()), "http://www.w3.org/XML/1998/namespace".into())],
            warnings: Vec::new(),
            errors: Vec::new(),
        }
    }

    /// error records an error at a position that ends parsing, returning it.
    fn error(&mut self, position: usize, text: &str) -> Error {
        self.note(position, text);
        Error { detail: self.errors.join("\n") }
    }

    /// note records an error at a position that parsing goes on after.
    fn note(&mut self, position: usize, text: &str) {
        self.errors.push(message(self.text, position, text));
    }

    /// finish returns the result of a parse, which fails when any error was recorded.
    fn finish<T>(&self, result: Result<T, Error>) -> Result<T, Error> {
        match result {
            Ok(_) if !self.errors.is_empty() => Err(Error { detail: self.errors.join("\n") }),
            other => other,
        }
    }

    /// peek returns the byte being read, or zero at the end.
    fn peek(&self) -> u8 {
        self.bytes.get(self.at).copied().unwrap_or(0)
    }

    /// rest returns the text left to read.
    fn rest(&self) -> &'a str {
        &self.text[self.at..]
    }

    /// starts reports whether the text left starts with a string.
    fn starts(&self, s: &str) -> bool {
        self.rest().starts_with(s)
    }

    /// skip_blanks skips white space, returning whether there was any.
    fn skip_blanks(&mut self) -> bool {
        let start = self.at;
        while is_blank(self.peek()) {
            self.at += 1;
        }
        self.at > start
    }

    /// name reads an XML name, or returns None when no name starts here.
    fn name(&mut self) -> Option<String> {
        let mut chars = self.rest().char_indices();
        let (_, first) = chars.next()?;
        if !is_name_start(first) {
            return None;
        }
        let end = chars.find(|(_, c)| !is_name_char(*c)).map_or(self.rest().len(), |(i, _)| i);
        let name = self.rest()[..end].to_string();
        self.at += end;
        Some(name)
    }

    /// resolve finds the namespace of a qualified name, for an element when `element` is set, as libxml2 does.
    fn resolve(&self, qualified: &str, element: bool) -> Name {
        let (prefix, local) = match qualified.split_once(':') {
            Some((prefix, local)) if !prefix.is_empty() && !local.is_empty() => (Some(prefix.to_string()), local),
            _ => (None, qualified),
        };
        let uri = match &prefix {
            None if !element => None,
            _ => self
                .scopes
                .iter()
                .rev()
                .find(|(p, _)| *p == prefix)
                .map(|(_, uri)| uri.clone())
                .filter(|u| !u.is_empty()),
        };
        Name { prefix, local: local.to_string(), uri }
    }

    /// reference reads a character or entity reference after its `&`, returning the text it stands for.
    fn reference(&mut self) -> Result<String, Error> {
        self.at += 1;
        if self.peek() == b'#' {
            self.at += 1;
            let hex = self.peek() == b'x';
            if hex {
                self.at += 1;
            }
            let digits_start = self.at;
            while if hex { self.peek().is_ascii_hexdigit() } else { self.peek().is_ascii_digit() } {
                self.at += 1;
            }
            let digits = &self.text[digits_start..self.at];
            if self.peek() != b';' {
                return Err(self.error(self.at, "xmlParseCharRef: invalid xmlChar value 0"));
            }
            self.at += 1;
            let value = u32::from_str_radix(digits, if hex { 16 } else { 10 }).ok();
            return match value.and_then(char::from_u32).filter(|&c| c != '\0') {
                Some(c) => Ok(c.to_string()),
                None => {
                    Err(self.error(self.at, &format!("xmlParseCharRef: invalid xmlChar value {}", value.unwrap_or(0))))
                }
            };
        }
        let Some(name) = self.name() else {
            return Err(self.error(self.at, "xmlParseEntityRef: no name"));
        };
        if self.peek() != b';' {
            return Err(self.error(self.at, "EntityRef: expecting ';'"));
        }
        self.at += 1;
        match name.as_str() {
            "lt" => Ok("<".into()),
            "gt" => Ok(">".into()),
            "amp" => Ok("&".into()),
            "quot" => Ok("\"".into()),
            "apos" => Ok("'".into()),
            _ => Err(self.error(self.at, &format!("Entity '{name}' not defined"))),
        }
    }

    /// attribute_value reads a quoted attribute value, replacing references and normalizing white space.
    fn attribute_value(&mut self) -> Result<String, Error> {
        let quote = self.peek();
        if quote != b'"' && quote != b'\'' {
            return Err(self.error(self.at, "AttValue: \" or ' expected"));
        }
        self.at += 1;
        let mut value = String::new();
        loop {
            match self.peek() {
                0 if self.at >= self.bytes.len() => {
                    return Err(self.error(self.at, "AttValue: ' expected"));
                }
                b if b == quote => {
                    self.at += 1;
                    return Ok(value);
                }
                b'<' => return Err(self.error(self.at, "Unescaped '<' not allowed in attributes values")),
                b'&' => {
                    let text = self.reference()?;
                    value.push_str(&text);
                }
                b'\t' | b'\n' | b'\r' => {
                    value.push(' ');
                    self.at += 1;
                }
                _ => {
                    let c = self.rest().chars().next().unwrap_or('\0');
                    value.push(c);
                    self.at += c.len_utf8();
                }
            }
        }
    }

    /// declaration reads an XML declaration at the start of the text, if there is one.
    fn declaration(&mut self) -> Result<Option<Declaration>, Error> {
        let is_declaration = self.starts("<?xml") && self.bytes.get(5).is_some_and(|&b| is_blank(b));
        if !is_declaration {
            return Ok(None);
        }
        self.at += 5;
        let mut declaration = Declaration::default();
        for (key, required) in [("version", true), ("encoding", false), ("standalone", false)] {
            let before = self.at;
            self.skip_blanks();
            if !self.starts(key) {
                self.at = before;
                if required {
                    return Err(self.error(self.at, "Malformed declaration expecting version"));
                }
                continue;
            }
            self.at += key.len();
            self.skip_blanks();
            if self.peek() != b'=' {
                return Err(self.error(self.at, "Blank needed here"));
            }
            self.at += 1;
            self.skip_blanks();
            let value = self.attribute_value()?;
            match key {
                "version" => {
                    if value != "1.0" {
                        self.warnings.push(message(self.text, self.at, &format!("Unsupported version '{value}'")));
                    }
                    declaration.version = Some(value);
                }
                "standalone" => declaration.standalone = Some(value == "yes"),
                _ => {}
            }
        }
        self.skip_blanks();
        if !self.starts("?>") {
            return Err(self.error(self.at, "parsing XML declaration: '?>' expected"));
        }
        self.at += 2;
        Ok(Some(declaration))
    }

    /// misc reads a comment or processing instruction outside the root element, or returns None at anything else.
    fn misc(&mut self) -> Result<Option<Node>, Error> {
        if self.starts("<!--") {
            return self.comment();
        }
        if self.starts("<?") {
            return self.pi().map(Some);
        }
        Ok(None)
    }

    /// comment reads a comment, or records that it never ends and returns None at the end of the text.
    fn comment(&mut self) -> Result<Option<Node>, Error> {
        self.at += 4;
        match self.rest().find("--") {
            Some(end) if self.rest()[end..].starts_with("-->") => {
                let text = self.rest()[..end].to_string();
                self.at += end + 3;
                Ok(Some(Node::Comment(text)))
            }
            Some(end) => {
                self.at += end;
                Err(self.error(self.at, "Double hyphen within comment"))
            }
            None => {
                self.at = self.bytes.len();
                self.note(self.at, "Comment not terminated");
                Ok(None)
            }
        }
    }

    /// pi reads a processing instruction.
    fn pi(&mut self) -> Result<Node, Error> {
        self.at += 2;
        let Some(target) = self.name() else {
            return Err(self.error(self.at, "xmlParsePI : no target name"));
        };
        if target.eq_ignore_ascii_case("xml") {
            return Err(self.error(self.at, "XML declaration allowed only at the start of the document"));
        }
        if self.starts("?>") {
            self.at += 2;
            return Ok(Node::Pi(target, String::new()));
        }
        if !self.skip_blanks() {
            return Err(self.error(self.at, "ParsePI: PI target space expected"));
        }
        match self.rest().find("?>") {
            Some(end) => {
                let data = self.rest()[..end].to_string();
                self.at += end + 2;
                Ok(Node::Pi(target, data))
            }
            None => {
                self.at = self.bytes.len();
                Err(self.error(self.at, &format!("ParsePI: PI {target} never end ...")))
            }
        }
    }

    /// doctype skips a document type declaration.
    fn doctype(&mut self) -> Result<(), Error> {
        let mut depth = 0;
        while self.at < self.bytes.len() {
            match self.peek() {
                b'[' => depth += 1,
                b']' => depth -= 1,
                b'>' if depth <= 0 => {
                    self.at += 1;
                    return Ok(());
                }
                _ => {}
            }
            self.at += 1;
        }
        Err(self.error(self.at, "DOCTYPE improperly terminated"))
    }

    /// element reads an element at a `<`, or records the errors of a start tag that libxml2 recovers from and
    /// returns None at the position where it goes on reading content.
    fn element(&mut self) -> Result<Option<Element>, Error> {
        let open = self.at;
        self.at += 1;
        let Some(qualified) = self.name() else {
            return Err(self.error(self.at, "StartTag: invalid element name"));
        };
        let line = line_of(self.text, open);
        let mut raw_attributes: Vec<(String, String)> = Vec::new();
        let mut namespaces = Vec::new();
        let unterminated = format!("Couldn't find end of Start Tag {qualified} line {line}");
        let self_closing = loop {
            let blank = self.skip_blanks();
            if self.starts("/>") {
                self.at += 2;
                break true;
            }
            if self.peek() == b'>' {
                self.at += 1;
                break false;
            }
            if self.at >= self.bytes.len() {
                self.note(self.at, &unterminated);
                return Ok(None);
            }
            let attribute_start = self.at;
            let name = self.name().filter(|_| blank);
            let Some(name) = name else {
                self.at = attribute_start;
                self.note(self.at, "attributes construct error");
                self.note(self.at, &unterminated);
                return Ok(None);
            };
            self.skip_blanks();
            if self.peek() != b'=' {
                return Err(self.error(self.at, &format!("Specification mandates value for attribute {name}")));
            }
            self.at += 1;
            self.skip_blanks();
            let value = match self.attribute_value() {
                Ok(value) => value,
                Err(_) if matches!(self.errors.last(), Some(e) if e.contains("AttValue") || e.contains("Unescaped '<'")) =>
                {
                    self.note(self.at, "attributes construct error");
                    self.note(self.at, &unterminated);
                    return Ok(None);
                }
                Err(err) => return Err(err),
            };
            if raw_attributes.iter().any(|(n, _)| *n == name) {
                return Err(self.error(self.at, &format!("Attribute {name} redefined")));
            }
            if name == "xmlns" {
                namespaces.push((None, value.clone()));
            } else if let Some(prefix) = name.strip_prefix("xmlns:") {
                namespaces.push((Some(prefix.to_string()), value.clone()));
            }
            raw_attributes.push((name, value));
        };
        let scope_depth = self.scopes.len();
        self.scopes.extend(namespaces.iter().cloned());
        let element_name = self.resolve(&qualified, true);
        let attributes = raw_attributes
            .into_iter()
            .filter(|(name, _)| name != "xmlns" && !name.starts_with("xmlns:"))
            .map(|(name, value)| Attribute { name: self.resolve(&name, false), value })
            .collect();
        let mut element = Element { name: element_name, namespaces, attributes, children: Vec::new() };
        if !self_closing {
            element.children = self.content(Some((&qualified, line)))?;
        }
        self.scopes.truncate(scope_depth);
        Ok(Some(element))
    }

    /// content reads mixed content until the end tag of an open element, or until the end of the text without one.
    fn content(&mut self, open: Option<(&str, usize)>) -> Result<Vec<Node>, Error> {
        let mut nodes = Vec::new();
        let mut text = String::new();
        let flush = |text: &mut String, nodes: &mut Vec<Node>| {
            if !text.is_empty() {
                nodes.push(Node::Text(std::mem::take(text)));
            }
        };
        loop {
            if self.at >= self.bytes.len() {
                if let Some((name, line)) = open {
                    return Err(self.error(self.at, &format!("Premature end of data in tag {name} line {line}")));
                }
                flush(&mut text, &mut nodes);
                return Ok(nodes);
            }
            match self.peek() {
                b'<' if self.starts("</") => {
                    let Some((name, line)) = open else {
                        flush(&mut text, &mut nodes);
                        return Err(self.error(self.at, "Extra content at the end of the document"));
                    };
                    self.at += 2;
                    let end_name = self.name().unwrap_or_default();
                    self.skip_blanks();
                    if self.peek() != b'>' {
                        return Err(self.error(self.at, "expected '>'"));
                    }
                    self.at += 1;
                    if end_name != name {
                        self.note(
                            self.at,
                            &format!("Opening and ending tag mismatch: {name} line {line} and {end_name}"),
                        );
                    }
                    flush(&mut text, &mut nodes);
                    return Ok(nodes);
                }
                b'<' if self.starts("<!--") => {
                    flush(&mut text, &mut nodes);
                    nodes.extend(self.comment()?);
                }
                b'<' if self.starts("<![CDATA[") => {
                    flush(&mut text, &mut nodes);
                    self.at += 9;
                    match self.rest().find("]]>") {
                        Some(end) => {
                            nodes.push(Node::CData(self.rest()[..end].to_string()));
                            self.at += end + 3;
                        }
                        None => {
                            self.at = self.bytes.len();
                            self.note(self.at, "Unregistered error message");
                        }
                    }
                }
                b'<' if self.starts("<?") => {
                    flush(&mut text, &mut nodes);
                    nodes.push(self.pi()?);
                }
                b'<' if self.starts("<!DOCTYPE") && open.is_none() => {
                    let start = self.at;
                    self.doctype()?;
                    text.push_str(&self.text[start..self.at]);
                }
                b'<' => {
                    flush(&mut text, &mut nodes);
                    if let Some(element) = self.element()? {
                        nodes.push(Node::Element(Box::new(element)));
                    }
                }
                b'&' => {
                    let replaced = self.reference()?;
                    text.push_str(&replaced);
                }
                b']' if self.starts("]]>") => {
                    return Err(self.error(self.at, "Sequence ']]>' not allowed in content"));
                }
                b'\r' => {
                    self.at += 1;
                    if self.peek() != b'\n' {
                        text.push('\n');
                    }
                }
                _ => {
                    let c = self.rest().chars().next().unwrap_or('\0');
                    text.push(c);
                    self.at += c.len_utf8();
                }
            }
        }
    }

    /// document reads a whole document: a declaration, comments and processing instructions, one root element, and
    /// more comments and processing instructions.
    fn document(&mut self) -> Result<Vec<Node>, Error> {
        if self.bytes.is_empty() {
            return Err(self.error(0, "Document is empty"));
        }
        self.declaration()?;
        let mut nodes = Vec::new();
        loop {
            self.skip_blanks();
            if self.starts("<!DOCTYPE") {
                self.doctype()?;
                continue;
            }
            match self.misc()? {
                Some(node) => nodes.push(node),
                None => break,
            }
        }
        if self.at >= self.bytes.len() && nodes.is_empty() && self.text.trim().is_empty() {
            return Err(self.error(self.at, "Document is empty"));
        }
        if self.peek() != b'<' {
            return Err(self.error(self.at, "Start tag expected, '<' not found"));
        }
        if let Some(element) = self.element()? {
            nodes.push(Node::Element(Box::new(element)));
        }
        loop {
            self.skip_blanks();
            match self.misc()? {
                Some(node) => nodes.push(node),
                None => break,
            }
        }
        if self.at < self.bytes.len() {
            return Err(self.error(self.at, "Extra content at the end of the document"));
        }
        Ok(nodes)
    }
}

/// parse_document parses text as a whole XML document.
pub fn parse_document(text: &str) -> Result<Parsed, Error> {
    let mut parser = Parser::new(text);
    let result = parser.document();
    let nodes = parser.finish(result)?;
    Ok(Parsed { nodes, warnings: parser.warnings })
}

/// parse_content parses text as XML content: any mix of text and elements, after an optional declaration.
pub fn parse_content(text: &str) -> Result<Parsed, Error> {
    let mut parser = Parser::new(text);
    let result = parser.declaration().and_then(|_| parser.content(None));
    let nodes = parser.finish(result)?;
    Ok(Parsed { nodes, warnings: parser.warnings })
}

/// split_declaration splits a leading XML declaration off text, returning it and the text after it.
pub fn split_declaration(text: &str) -> (Option<Declaration>, &str) {
    let mut parser = Parser::new(text);
    match parser.declaration() {
        Ok(Some(declaration)) => (Some(declaration), &text[parser.at..]),
        _ => (None, text),
    }
}

/// escape_text escapes text for an XML text node, as libxml2 writes it.
pub fn escape_text(text: &str) -> String {
    let mut out = String::with_capacity(text.len());
    for c in text.chars() {
        match c {
            '&' => out.push_str("&amp;"),
            '<' => out.push_str("&lt;"),
            '>' => out.push_str("&gt;"),
            '\r' => out.push_str("&#13;"),
            c => out.push(c),
        }
    }
    out
}

/// escape_attribute escapes text for a double-quoted attribute value, as libxml2 writes it.
pub fn escape_attribute(text: &str) -> String {
    let mut out = String::with_capacity(text.len());
    for c in text.chars() {
        match c {
            '&' => out.push_str("&amp;"),
            '<' => out.push_str("&lt;"),
            '>' => out.push_str("&gt;"),
            '"' => out.push_str("&quot;"),
            '\n' => out.push_str("&#10;"),
            '\r' => out.push_str("&#13;"),
            '\t' => out.push_str("&#9;"),
            c => out.push(c),
        }
    }
    out
}

/// serialize writes a node as libxml2's xmlNodeDump writes it.
pub fn serialize(node: &Node, out: &mut String) {
    match node {
        Node::Element(element) => serialize_element(element, out),
        Node::Text(text) => out.push_str(&escape_text(text)),
        Node::CData(text) => {
            out.push_str("<![CDATA[");
            out.push_str(text);
            out.push_str("]]>");
        }
        Node::Comment(text) => {
            out.push_str("<!--");
            out.push_str(text);
            out.push_str("-->");
        }
        Node::Pi(target, data) => {
            out.push_str("<?");
            out.push_str(target);
            if !data.is_empty() {
                out.push(' ');
                out.push_str(data);
            }
            out.push_str("?>");
        }
    }
}

/// serialize_element writes an element with its namespace declarations, attributes, and children.
fn serialize_element(element: &Element, out: &mut String) {
    let name = element.name.qualified();
    out.push('<');
    out.push_str(&name);
    for (prefix, uri) in &element.namespaces {
        match prefix {
            Some(prefix) => out.push_str(&format!(" xmlns:{prefix}=\"{}\"", escape_attribute(uri))),
            None => out.push_str(&format!(" xmlns=\"{}\"", escape_attribute(uri))),
        }
    }
    for attribute in &element.attributes {
        out.push_str(&format!(" {}=\"{}\"", attribute.name.qualified(), escape_attribute(&attribute.value)));
    }
    if element.children.is_empty() {
        out.push_str("/>");
        return;
    }
    out.push('>');
    for child in &element.children {
        serialize(child, out);
    }
    out.push_str("</");
    out.push_str(&name);
    out.push('>');
}
