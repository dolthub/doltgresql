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

//! SQL/JSON path expressions: parsing them as Postgres' jsonpath grammar does, and evaluating them over JSON values.

pub mod exec;

use crate::error::{PgError, Result, code};
use crate::numeric::Numeric;

/// JsonPath is a parsed SQL/JSON path expression, in lax or strict mode.
#[derive(Clone, Debug, PartialEq)]
pub struct JsonPath {
    pub lax: bool,
    pub expr: Node,
}

/// ArithOp is a binary arithmetic operator of a path expression.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum ArithOp {
    Add,
    Sub,
    Mul,
    Div,
    Mod,
}

/// CmpOp is a comparison operator of a path predicate.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum CmpOp {
    Eq,
    Ne,
    Lt,
    Le,
    Gt,
    Ge,
}

/// Method is an item method that takes no arguments.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Method {
    Abs,
    Size,
    Type,
    Floor,
    Double,
    Ceiling,
    KeyValue,
    Bigint,
    Boolean,
    Date,
    Integer,
    Number,
    String,
}

/// TimeMethod is an item method that converts to a time or timestamp type with an optional precision.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum TimeMethod {
    Time,
    TimeTz,
    Timestamp,
    TimestampTz,
}

/// ANY_LAST is the level bound of a `.**` accessor that stands for `last`.
pub const ANY_LAST: u32 = u32::MAX;

/// Accessor is a step that follows a primary in an accessor chain.
#[derive(Clone, Debug, PartialEq)]
pub enum Accessor {
    Key(String),
    AnyKey,
    AnyArray,
    /// Array subscripts, each a single index or a range.
    Index(Vec<(Node, Option<Node>)>),
    /// The `.**` accessor over the levels from the first to the last.
    Any(u32, u32),
    Filter(Box<Node>),
    Method(Method),
    /// The `.datetime()` method with its optional template.
    Datetime(Option<String>),
    /// The `.decimal()` method with its optional precision and scale.
    Decimal(Option<Box<Node>>, Option<Box<Node>>),
    Time(TimeMethod, Option<i32>),
}

/// Node is an expression or predicate of a path.
#[derive(Clone, Debug, PartialEq)]
pub enum Node {
    Root,
    Current,
    Last,
    Null,
    Bool(bool),
    Number(Numeric),
    String(String),
    Variable(String),
    Chain(Box<Node>, Vec<Accessor>),
    /// A unary plus, or a minus when the flag is set.
    Unary(bool, Box<Node>),
    Arith(ArithOp, Box<Node>, Box<Node>),
    Compare(CmpOp, Box<Node>, Box<Node>),
    And(Box<Node>, Box<Node>),
    Or(Box<Node>, Box<Node>),
    Not(Box<Node>),
    IsUnknown(Box<Node>),
    StartsWith(Box<Node>, Box<Node>),
    LikeRegex(Box<Node>, String, String),
    Exists(Box<Node>),
}

impl Node {
    /// is_predicate reports whether the node is a predicate, which yields a boolean or unknown.
    pub fn is_predicate(&self) -> bool {
        matches!(
            self,
            Node::Compare(..)
                | Node::And(..)
                | Node::Or(..)
                | Node::Not(..)
                | Node::IsUnknown(..)
                | Node::StartsWith(..)
                | Node::LikeRegex(..)
                | Node::Exists(..)
        )
    }
}

impl JsonPath {
    /// text returns the path as the jsonpath type prints it, as Postgres' jsonPathToCstring does.
    pub fn text(&self) -> String {
        let mut out = String::new();
        if !self.lax {
            out.push_str("strict ");
        }
        print(&mut out, &self.expr, true, false);
        out
    }
}

/// priority returns how tightly a node's operator binds, as Postgres' operationPriority ranks it.
fn priority(node: &Node) -> u8 {
    match node {
        Node::Or(..) => 0,
        Node::And(..) => 1,
        Node::Compare(..) | Node::StartsWith(..) => 2,
        Node::Arith(ArithOp::Add | ArithOp::Sub, ..) => 3,
        Node::Arith(..) => 4,
        Node::Unary(..) if folded(node).is_none() => 5,
        Node::Chain(base, _) => priority(base),
        _ => 6,
    }
}

/// folded returns the number that signs applied to a number literal make, as Postgres' parser folds them.
fn folded(node: &Node) -> Option<Numeric> {
    match node {
        Node::Number(n) => Some(n.clone()),
        Node::Unary(true, operand) => folded(operand).map(|n| n.negate()),
        Node::Unary(false, operand) => folded(operand),
        _ => None,
    }
}

/// print writes a node as Postgres' printJsonPathItem does, in parentheses when `brackets` asks for them around an
/// operator, and with a number in parentheses when accessors follow it.
fn print(out: &mut String, node: &Node, brackets: bool, followed: bool) {
    let binary = |out: &mut String, left: &Node, op: &str, right: &Node| {
        if brackets {
            out.push('(');
        }
        print(out, left, priority(left) <= priority(node), false);
        out.push(' ');
        out.push_str(op);
        out.push(' ');
        print(out, right, priority(right) <= priority(node), false);
        if brackets {
            out.push(')');
        }
    };
    match node {
        Node::Root => out.push('$'),
        Node::Current => out.push('@'),
        Node::Last => out.push_str("last"),
        Node::Null => out.push_str("null"),
        Node::Bool(b) => out.push_str(if *b { "true" } else { "false" }),
        Node::Number(n) if followed => out.push_str(&format!("({n})")),
        Node::Number(n) => out.push_str(&n.to_string()),
        Node::String(s) => crate::json::escape(out, s),
        Node::Variable(name) => {
            out.push('$');
            crate::json::escape(out, name);
        }
        Node::Chain(base, accessors) => {
            print(out, base, brackets, !accessors.is_empty());
            for accessor in accessors {
                print_accessor(out, accessor);
            }
        }
        Node::Unary(..) if let Some(n) = folded(node) => print(out, &Node::Number(n), brackets, followed),
        Node::Unary(negate, operand) => {
            if brackets {
                out.push('(');
            }
            out.push(if *negate { '-' } else { '+' });
            print(out, operand, priority(operand) <= priority(node), false);
            if brackets {
                out.push(')');
            }
        }
        Node::Arith(op, left, right) => {
            let op = match op {
                ArithOp::Add => "+",
                ArithOp::Sub => "-",
                ArithOp::Mul => "*",
                ArithOp::Div => "/",
                ArithOp::Mod => "%",
            };
            binary(out, left, op, right);
        }
        Node::Compare(op, left, right) => {
            let op = match op {
                CmpOp::Eq => "==",
                CmpOp::Ne => "!=",
                CmpOp::Lt => "<",
                CmpOp::Le => "<=",
                CmpOp::Gt => ">",
                CmpOp::Ge => ">=",
            };
            binary(out, left, op, right);
        }
        Node::And(left, right) => binary(out, left, "&&", right),
        Node::Or(left, right) => binary(out, left, "||", right),
        Node::StartsWith(left, right) => binary(out, left, "starts with", right),
        Node::Not(operand) => {
            out.push_str("!(");
            print(out, operand, false, false);
            out.push(')');
        }
        Node::IsUnknown(operand) => {
            out.push('(');
            print(out, operand, false, false);
            out.push_str(") is unknown");
        }
        Node::Exists(operand) => {
            out.push_str("exists (");
            print(out, operand, false, false);
            out.push(')');
        }
        Node::LikeRegex(operand, pattern, flags) => {
            if brackets {
                out.push('(');
            }
            print(out, operand, priority(operand) <= priority(node), false);
            out.push_str(" like_regex ");
            crate::json::escape(out, pattern);
            let flags: String = "ismxq".chars().filter(|&f| flags.contains(f)).collect();
            if !flags.is_empty() {
                out.push_str(&format!(" flag \"{flags}\""));
            }
            if brackets {
                out.push(')');
            }
        }
    }
}

/// print_accessor writes an accessor that follows a primary, as printJsonPathItem prints an item's next item.
fn print_accessor(out: &mut String, accessor: &Accessor) {
    match accessor {
        Accessor::Key(key) => {
            out.push('.');
            crate::json::escape(out, key);
        }
        Accessor::AnyKey => out.push_str(".*"),
        Accessor::AnyArray => out.push_str("[*]"),
        Accessor::Index(subscripts) => {
            out.push('[');
            for (i, (from, to)) in subscripts.iter().enumerate() {
                if i > 0 {
                    out.push(',');
                }
                print(out, from, false, false);
                if let Some(to) = to {
                    out.push_str(" to ");
                    print(out, to, false, false);
                }
            }
            out.push(']');
        }
        Accessor::Any(first, last) => {
            let level = |n: u32| if n == ANY_LAST { "last".to_string() } else { n.to_string() };
            match (*first, *last) {
                (0, ANY_LAST) => out.push_str(".**"),
                (first, last) if first == last => out.push_str(&format!(".**{{{}}}", level(first))),
                (first, last) => out.push_str(&format!(".**{{{} to {}}}", level(first), level(last))),
            }
        }
        Accessor::Filter(predicate) => {
            out.push_str("?(");
            print(out, predicate, false, false);
            out.push(')');
        }
        Accessor::Method(method) => {
            let name = match method {
                Method::Abs => "abs",
                Method::Size => "size",
                Method::Type => "type",
                Method::Floor => "floor",
                Method::Double => "double",
                Method::Ceiling => "ceiling",
                Method::KeyValue => "keyvalue",
                Method::Bigint => "bigint",
                Method::Boolean => "boolean",
                Method::Date => "date",
                Method::Integer => "integer",
                Method::Number => "number",
                Method::String => "string",
            };
            out.push_str(&format!(".{name}()"));
        }
        Accessor::Datetime(template) => {
            out.push_str(".datetime(");
            if let Some(template) = template {
                crate::json::escape(out, template);
            }
            out.push(')');
        }
        Accessor::Decimal(precision, scale) => {
            out.push_str(".decimal(");
            if let Some(precision) = precision {
                print(out, precision, false, false);
            }
            if let Some(scale) = scale {
                out.push(',');
                print(out, scale, false, false);
            }
            out.push(')');
        }
        Accessor::Time(kind, precision) => {
            let name = match kind {
                TimeMethod::Time => "time",
                TimeMethod::TimeTz => "time_tz",
                TimeMethod::Timestamp => "timestamp",
                TimeMethod::TimestampTz => "timestamp_tz",
            };
            out.push_str(&format!(".{name}("));
            if let Some(precision) = precision {
                out.push_str(&precision.to_string());
            }
            out.push(')');
        }
    }
}

/// Token is a lexical token of a path, with the source text that errors quote.
#[derive(Clone, Debug, PartialEq)]
enum Token {
    End,
    String(String),
    Number(Numeric),
    Variable(String),
    Identifier(String),
    Punct(&'static str),
}

/// PUNCTUATION lists the path's operators and punctuation, longest first.
const PUNCTUATION: &[&str] = &[
    "==", "!=", "<>", "<=", ">=", "&&", "||", "**", ".", "[", "]", "(", ")", "{", "}", ",", "?", "@", "*", "+", "-",
    "/", "%", "<", ">", "!", ":", "$",
];

/// is_other reports whether a character can be part of an unquoted identifier, as the jsonpath scanner's `other`
/// class defines.
fn is_other(c: char) -> bool {
    !matches!(
        c,
        '?' | '%'
            | '$'
            | '.'
            | '['
            | ']'
            | '{'
            | '}'
            | '('
            | ')'
            | '|'
            | '&'
            | '!'
            | '='
            | '<'
            | '>'
            | '@'
            | '#'
            | ','
            | '*'
            | ':'
            | '-'
            | '+'
            | '/'
            | '\\'
            | '"'
            | ' '
            | '\t'
            | '\n'
            | '\r'
            | '\u{c}'
    )
}

/// syntax_error returns Postgres' error for a path that does not parse at the token with the text.
fn syntax_error(text: Option<&str>) -> PgError {
    match text {
        None => PgError::new(code::SYNTAX_ERROR, "syntax error at end of jsonpath input"),
        Some(text) => PgError::new(code::SYNTAX_ERROR, format!("syntax error at or near \"{text}\" of jsonpath input")),
    }
}

/// lex splits a path into tokens, each with its source text.
fn lex(text: &str) -> Result<Vec<(Token, String)>> {
    let chars: Vec<char> = text.chars().collect();
    let mut tokens = Vec::new();
    let mut i = 0;
    while i < chars.len() {
        let c = chars[i];
        if matches!(c, ' ' | '\t' | '\n' | '\r' | '\u{c}') {
            i += 1;
            continue;
        }
        let start = i;
        if c == '"' {
            let (value, end) = quoted(&chars, i)?;
            i = end;
            tokens.push((Token::String(value), chars[start..i].iter().collect()));
            continue;
        }
        if c == '$' && i + 1 < chars.len() && (chars[i + 1] == '"' || is_other(chars[i + 1])) {
            let (name, end) = match chars[i + 1] {
                '"' => quoted(&chars, i + 1)?,
                _ => name(&chars, i + 1)?,
            };
            i = end;
            tokens.push((Token::Variable(name), chars[start..i].iter().collect()));
            continue;
        }
        if c.is_ascii_digit() || c == '.' && chars.get(i + 1).is_some_and(|c| c.is_ascii_digit()) {
            let (value, end) = number(&chars, i)?;
            i = end;
            tokens.push((Token::Number(value), chars[start..i].iter().collect()));
            continue;
        }
        if let Some(punct) = PUNCTUATION.iter().find(|p| {
            let p: Vec<char> = p.chars().collect();
            chars[i..].starts_with(&p)
        }) {
            i += punct.chars().count();
            tokens.push((Token::Punct(punct), punct.to_string()));
            continue;
        }
        if is_other(c) || c == '\\' {
            let (word, end) = name(&chars, i)?;
            i = end;
            tokens.push((Token::Identifier(word), chars[start..i].iter().collect()));
            continue;
        }
        return Err(syntax_error(Some(&c.to_string())));
    }
    tokens.push((Token::End, String::new()));
    Ok(tokens)
}

/// quoted reads the double-quoted string that starts at a position, returning its value and the position after it.
fn quoted(chars: &[char], start: usize) -> Result<(String, usize)> {
    let mut value = String::new();
    let mut i = start + 1;
    while i < chars.len() {
        match chars[i] {
            '"' => return Ok((value, i + 1)),
            '\\' => i = escape(chars, i, &mut value)?,
            other => {
                value.push(other);
                i += 1;
            }
        }
    }
    Err(PgError::new(code::SYNTAX_ERROR, "unexpected end of quoted string at end of jsonpath input"))
}

/// name reads the unquoted name that starts at a position, with its escapes, returning its value and the position
/// after it.
fn name(chars: &[char], start: usize) -> Result<(String, usize)> {
    let mut value = String::new();
    let mut i = start;
    while i < chars.len() {
        match chars[i] {
            '\\' => i = escape(chars, i, &mut value)?,
            c if is_other(c) => {
                value.push(c);
                i += 1;
            }
            _ => break,
        }
    }
    Ok((value, i))
}

/// escape reads the backslash escape at a position into a value, returning the position after it, as the jsonpath
/// scanner does: a run of `\u` escapes is read together so that surrogate pairs combine.
fn escape(chars: &[char], start: usize, value: &mut String) -> Result<usize> {
    let near = |end: usize, message: &str| {
        let text: String = chars[start..end.min(chars.len())].iter().collect();
        PgError::new(code::SYNTAX_ERROR, format!("{message} at or near \"{text}\" of jsonpath input"))
    };
    let hex = |from: usize, max: usize| {
        chars[from.min(chars.len())..].iter().take(max).take_while(|c| c.is_ascii_hexdigit()).count()
    };
    let Some(&escaped) = chars.get(start + 1) else {
        return Err(PgError::new(code::SYNTAX_ERROR, "unexpected end after backslash at end of jsonpath input"));
    };
    let simple = match escaped {
        'b' => Some('\u{8}'),
        'f' => Some('\u{c}'),
        'n' => Some('\n'),
        'r' => Some('\r'),
        't' => Some('\t'),
        'v' => Some('\u{b}'),
        'x' | 'u' => None,
        other => Some(other),
    };
    if let Some(c) = simple {
        value.push(c);
        return Ok(start + 2);
    }
    if escaped == 'x' {
        if hex(start + 2, 2) < 2 {
            return Err(near(start + 2 + hex(start + 2, 2), "invalid hexadecimal character sequence"));
        }
        let code: String = chars[start + 2..start + 4].iter().collect();
        push_code(value, u32::from_str_radix(&code, 16).unwrap_or(0))?;
        return Ok(start + 4);
    }
    let mut codes = Vec::new();
    let mut i = start;
    while chars.get(i) == Some(&'\\') && chars.get(i + 1) == Some(&'u') {
        let (digits, end) = match chars.get(i + 2) {
            Some('{') => {
                let n = hex(i + 3, 7);
                if n == 0 || n > 6 || chars.get(i + 3 + n) != Some(&'}') {
                    return Err(near(i + 3 + n.min(6), "invalid unicode sequence"));
                }
                (chars[i + 3..i + 3 + n].iter().collect::<String>(), i + 4 + n)
            }
            _ => {
                let n = hex(i + 2, 4);
                if n < 4 {
                    return Err(near(i + 2 + n, "invalid unicode sequence"));
                }
                (chars[i + 2..i + 6].iter().collect::<String>(), i + 6)
            }
        };
        codes.push(u32::from_str_radix(&digits, 16).unwrap_or(0));
        i = end;
    }
    let surrogate = |detail: &str| PgError {
        detail: Some(detail.into()),
        ..PgError::new(code::SYNTAX_ERROR, "invalid input syntax for type jsonpath")
    };
    let mut high: Option<u32> = None;
    for code in codes {
        let code = match code {
            0xD800..=0xDBFF => {
                if high.is_some() {
                    return Err(surrogate("Unicode high surrogate must not follow a high surrogate."));
                }
                high = Some(code);
                continue;
            }
            0xDC00..=0xDFFF => match high.take() {
                Some(high) => 0x10000 + ((high - 0xD800) << 10) + (code - 0xDC00),
                None => return Err(surrogate("Unicode low surrogate must follow a high surrogate.")),
            },
            _ if high.is_some() => return Err(surrogate("Unicode low surrogate must follow a high surrogate.")),
            code => code,
        };
        push_code(value, code)?;
    }
    if high.is_some() {
        return Err(surrogate("Unicode low surrogate must follow a high surrogate."));
    }
    Ok(i)
}

/// push_code adds the character with a code point to a value, refusing the zero character, which text cannot hold.
fn push_code(value: &mut String, code: u32) -> Result<()> {
    if code == 0 {
        return Err(PgError {
            detail: Some("\\u0000 cannot be converted to text.".into()),
            ..PgError::new(code::UNTRANSLATABLE_CHARACTER, "unsupported Unicode escape sequence")
        });
    }
    let c = char::from_u32(code).ok_or_else(|| PgError::new(code::SYNTAX_ERROR, "invalid Unicode code point"))?;
    value.push(c);
    Ok(())
}

/// number reads the numeric literal that starts at a position, returning its value and the position after it.
fn number(chars: &[char], start: usize) -> Result<(Numeric, usize)> {
    let digits_from = |mut i: usize, radix: u32| {
        while i < chars.len()
            && (chars[i].is_digit(radix) || chars[i] == '_' && i + 1 < chars.len() && chars[i + 1].is_digit(radix))
        {
            i += 1;
        }
        i
    };
    let radix = match chars.get(start + 1) {
        Some('x' | 'X') if chars[start] == '0' => Some(16),
        Some('o' | 'O') if chars[start] == '0' => Some(8),
        Some('b' | 'B') if chars[start] == '0' => Some(2),
        _ => None,
    };
    let (value, i) = if let Some(radix) = radix {
        let i = digits_from(start + 2, radix);
        let digits: String = chars[start + 2..i].iter().filter(|c| **c != '_').collect();
        let parsed = num_bigint::BigUint::parse_bytes(digits.as_bytes(), radix);
        (parsed.map(|n| Numeric::Finite { negative: false, coefficient: n, scale: 0 }), i)
    } else {
        let mut i = if chars[start] == '0' { start + 1 } else { digits_from(start, 10) };
        if chars.get(i) == Some(&'.') {
            i = digits_from(i + 1, 10);
        }
        if matches!(chars.get(i), Some('e' | 'E')) {
            let mut j = i + 1;
            if matches!(chars.get(j), Some('+' | '-')) {
                j += 1;
            }
            if chars.get(j).is_some_and(|c| c.is_ascii_digit()) {
                i = digits_from(j, 10);
            }
        }
        let text: String = chars[start..i].iter().filter(|c| **c != '_').collect();
        (Numeric::parse(&text).ok(), i)
    };
    if i < chars.len() && is_other(chars[i]) {
        let junk: String = chars[start..=i].iter().collect();
        return Err(PgError::new(
            code::SYNTAX_ERROR,
            format!("trailing junk after numeric literal at or near \"{junk}\" of jsonpath input"),
        ));
    }
    let value = value.ok_or_else(|| syntax_error(Some(&chars[start..i].iter().collect::<String>())))?;
    Ok((value, i))
}

/// Parser reads a path's tokens by recursive descent, following the precedence of Postgres' jsonpath grammar.
struct Parser {
    tokens: Vec<(Token, String)>,
    position: usize,
}

impl Parser {
    /// peek returns the current token.
    fn peek(&self) -> &Token {
        &self.tokens[self.position].0
    }

    /// peek_at returns the token after the current one by an offset.
    fn peek_at(&self, offset: usize) -> &Token {
        &self.tokens[(self.position + offset).min(self.tokens.len() - 1)].0
    }

    /// advance moves past the current token, returning it.
    fn advance(&mut self) -> Token {
        let token = self.tokens[self.position].0.clone();
        if self.position + 1 < self.tokens.len() {
            self.position += 1;
        }
        token
    }

    /// error returns the syntax error at the current token.
    fn error(&self) -> PgError {
        match &self.tokens[self.position] {
            (Token::End, _) => syntax_error(None),
            (_, text) => syntax_error(Some(text)),
        }
    }

    /// is_punct reports whether the current token is the punctuation.
    fn is_punct(&self, punct: &str) -> bool {
        matches!(self.peek(), Token::Punct(p) if *p == punct)
    }

    /// is_word reports whether the current token is the identifier, which keywords are.
    fn is_word(&self, word: &str) -> bool {
        matches!(self.peek(), Token::Identifier(w) if w == word)
    }

    /// expect moves past the punctuation, failing when it is something else.
    fn expect(&mut self, punct: &str) -> Result<()> {
        if !self.is_punct(punct) {
            return Err(self.error());
        }
        self.advance();
        Ok(())
    }

    /// expression returns a node that must be an expression rather than a predicate.
    fn expression(&mut self, node: Node) -> Result<Node> {
        if node.is_predicate() { Err(self.error()) } else { Ok(node) }
    }

    /// predicate returns a node that must be a predicate rather than an expression.
    fn predicate(&mut self, node: Node) -> Result<Node> {
        if node.is_predicate() { Ok(node) } else { Err(self.error()) }
    }

    /// or parses predicates joined by `||`.
    fn or(&mut self) -> Result<Node> {
        let mut left = self.and()?;
        while self.is_punct("||") {
            left = self.predicate(left)?;
            self.advance();
            let right = self.and()?;
            let right = self.predicate(right)?;
            left = Node::Or(Box::new(left), Box::new(right));
        }
        Ok(left)
    }

    /// and parses predicates joined by `&&`.
    fn and(&mut self) -> Result<Node> {
        let mut left = self.not()?;
        while self.is_punct("&&") {
            left = self.predicate(left)?;
            self.advance();
            let right = self.not()?;
            let right = self.predicate(right)?;
            left = Node::And(Box::new(left), Box::new(right));
        }
        Ok(left)
    }

    /// not parses a negated delimited predicate, or a comparison.
    fn not(&mut self) -> Result<Node> {
        if !self.is_punct("!") {
            return self.comparison();
        }
        self.advance();
        let delimited = self.is_punct("(") || self.is_word("exists") && matches!(self.peek_at(1), Token::Punct("("));
        if !delimited {
            return Err(self.error());
        }
        let inner = self.accessor_expr()?;
        let inner = self.predicate(inner)?;
        Ok(Node::Not(Box::new(inner)))
    }

    /// comparison parses an arithmetic expression and the comparison, `starts with`, or `like_regex` predicate that
    /// it begins.
    fn comparison(&mut self) -> Result<Node> {
        let left = self.additive()?;
        let op = match self.peek() {
            Token::Punct("==") => Some(CmpOp::Eq),
            Token::Punct("!=" | "<>") => Some(CmpOp::Ne),
            Token::Punct("<") => Some(CmpOp::Lt),
            Token::Punct("<=") => Some(CmpOp::Le),
            Token::Punct(">") => Some(CmpOp::Gt),
            Token::Punct(">=") => Some(CmpOp::Ge),
            _ => None,
        };
        if let Some(op) = op {
            let left = self.expression(left)?;
            self.advance();
            let right = self.additive()?;
            let right = self.expression(right)?;
            return Ok(Node::Compare(op, Box::new(left), Box::new(right)));
        }
        if self.is_word("starts") && matches!(self.peek_at(1), Token::Identifier(w) if w == "with") {
            let left = self.expression(left)?;
            self.advance();
            self.advance();
            let initial = match self.advance() {
                Token::String(s) => Node::String(s),
                Token::Variable(v) => Node::Variable(v),
                _ => {
                    self.position -= 1;
                    return Err(self.error());
                }
            };
            return Ok(Node::StartsWith(Box::new(left), Box::new(initial)));
        }
        if self.is_word("like_regex") {
            let left = self.expression(left)?;
            self.advance();
            let Token::String(pattern) = self.advance() else {
                self.position -= 1;
                return Err(self.error());
            };
            let mut flags = String::new();
            if self.is_word("flag") {
                self.advance();
                let Token::String(f) = self.advance() else {
                    self.position -= 1;
                    return Err(self.error());
                };
                flags = f;
            }
            check_flags(&pattern, &flags)?;
            return Ok(Node::LikeRegex(Box::new(left), pattern, flags));
        }
        Ok(left)
    }

    /// additive parses expressions joined by `+` and `-`.
    fn additive(&mut self) -> Result<Node> {
        let mut left = self.multiplicative()?;
        loop {
            let op = match self.peek() {
                Token::Punct("+") => ArithOp::Add,
                Token::Punct("-") => ArithOp::Sub,
                _ => return Ok(left),
            };
            left = self.expression(left)?;
            self.advance();
            let right = self.multiplicative()?;
            let right = self.expression(right)?;
            left = Node::Arith(op, Box::new(left), Box::new(right));
        }
    }

    /// multiplicative parses expressions joined by `*`, `/`, and `%`.
    fn multiplicative(&mut self) -> Result<Node> {
        let mut left = self.unary()?;
        loop {
            let op = match self.peek() {
                Token::Punct("*") => ArithOp::Mul,
                Token::Punct("/") => ArithOp::Div,
                Token::Punct("%") => ArithOp::Mod,
                _ => return Ok(left),
            };
            left = self.expression(left)?;
            self.advance();
            let right = self.unary()?;
            let right = self.expression(right)?;
            left = Node::Arith(op, Box::new(left), Box::new(right));
        }
    }

    /// unary parses a unary plus or minus, or an accessor expression.
    fn unary(&mut self) -> Result<Node> {
        let negate = match self.peek() {
            Token::Punct("+") => false,
            Token::Punct("-") => true,
            _ => return self.accessor_expr(),
        };
        self.advance();
        let operand = self.unary()?;
        let operand = self.expression(operand)?;
        Ok(Node::Unary(negate, Box::new(operand)))
    }

    /// accessor_expr parses a primary, a parenthesized expression or predicate, or an `exists` predicate, with the
    /// accessors that follow it.
    fn accessor_expr(&mut self) -> Result<Node> {
        let primary = match self.advance() {
            Token::Punct("(") => {
                let inner = self.or()?;
                self.expect(")")?;
                if inner.is_predicate() && self.is_word("is") {
                    self.advance();
                    if !self.is_word("unknown") {
                        return Err(self.error());
                    }
                    self.advance();
                    return Ok(Node::IsUnknown(Box::new(inner)));
                }
                if !self.at_accessor() {
                    return Ok(inner);
                }
                inner
            }
            Token::Identifier(w) if w == "exists" && self.is_punct("(") => {
                self.advance();
                let inner = self.or()?;
                let inner = self.expression(inner)?;
                self.expect(")")?;
                return Ok(Node::Exists(Box::new(inner)));
            }
            Token::Punct("$") => Node::Root,
            Token::Punct("@") => Node::Current,
            Token::Identifier(w) if w == "last" => Node::Last,
            Token::Identifier(w) if w == "null" => Node::Null,
            Token::Identifier(w) if w == "true" => Node::Bool(true),
            Token::Identifier(w) if w == "false" => Node::Bool(false),
            Token::String(s) => Node::String(s),
            Token::Number(n) => Node::Number(n),
            Token::Variable(v) => Node::Variable(v),
            _ => {
                self.position = self.position.saturating_sub(1);
                if matches!(self.peek(), Token::End) || self.position + 1 == self.tokens.len() {
                    return Err(syntax_error(None));
                }
                return Err(self.error());
            }
        };
        let mut accessors = Vec::new();
        while self.at_accessor() {
            accessors.push(self.accessor()?);
        }
        Ok(if accessors.is_empty() { primary } else { Node::Chain(Box::new(primary), accessors) })
    }

    /// at_accessor reports whether an accessor starts at the current token.
    fn at_accessor(&self) -> bool {
        self.is_punct(".") || self.is_punct("[") || self.is_punct("?")
    }

    /// accessor parses one accessor.
    fn accessor(&mut self) -> Result<Accessor> {
        match self.advance() {
            Token::Punct("[") => self.subscripts(),
            Token::Punct("?") => {
                self.expect("(")?;
                let inner = self.or()?;
                let inner = self.predicate(inner)?;
                self.expect(")")?;
                Ok(Accessor::Filter(Box::new(inner)))
            }
            _ => self.member(),
        }
    }

    /// subscripts parses the subscripts of an array accessor after its `[`.
    fn subscripts(&mut self) -> Result<Accessor> {
        if self.is_punct("*") {
            self.advance();
            self.expect("]")?;
            return Ok(Accessor::AnyArray);
        }
        let mut list = Vec::new();
        loop {
            let from = self.or()?;
            let from = self.expression(from)?;
            let to = if self.is_word("to") {
                self.advance();
                let to = self.or()?;
                Some(self.expression(to)?)
            } else {
                None
            };
            list.push((from, to));
            if self.is_punct(",") {
                self.advance();
                continue;
            }
            self.expect("]")?;
            return Ok(Accessor::Index(list));
        }
    }

    /// member parses the accessor after a `.`: a key, a wildcard, `**`, or an item method.
    fn member(&mut self) -> Result<Accessor> {
        if self.is_punct("*") {
            self.advance();
            return Ok(Accessor::AnyKey);
        }
        if self.is_punct("**") {
            self.advance();
            if !self.is_punct("{") {
                return Ok(Accessor::Any(0, ANY_LAST));
            }
            self.advance();
            let first = self.level()?;
            let last = if self.is_word("to") {
                self.advance();
                self.level()?
            } else {
                first
            };
            self.expect("}")?;
            return Ok(Accessor::Any(first, last));
        }
        let name = match self.advance() {
            Token::String(s) => return Ok(Accessor::Key(s)),
            Token::Identifier(w) => w,
            _ => {
                self.position -= 1;
                return Err(self.error());
            }
        };
        if !self.is_punct("(") {
            return Ok(Accessor::Key(name));
        }
        let method = match name.as_str() {
            "abs" => Method::Abs,
            "size" => Method::Size,
            "type" => Method::Type,
            "floor" => Method::Floor,
            "double" => Method::Double,
            "ceiling" => Method::Ceiling,
            "keyvalue" => Method::KeyValue,
            "bigint" => Method::Bigint,
            "boolean" => Method::Boolean,
            "date" => Method::Date,
            "integer" => Method::Integer,
            "number" => Method::Number,
            "string" => Method::String,
            "datetime" => {
                self.advance();
                let template = match self.peek().clone() {
                    Token::String(s) => {
                        self.advance();
                        Some(s)
                    }
                    _ => None,
                };
                self.expect(")")?;
                return Ok(Accessor::Datetime(template));
            }
            "decimal" => {
                self.advance();
                let mut args = Vec::new();
                while !self.is_punct(")") {
                    let negate = self.is_punct("-");
                    if negate || self.is_punct("+") {
                        self.advance();
                    }
                    let Token::Number(n) = self.advance() else {
                        self.position -= 1;
                        return Err(self.error());
                    };
                    let n = Node::Number(n);
                    args.push(if negate { Node::Unary(true, Box::new(n)) } else { n });
                    if self.is_punct(",") && args.len() < 2 {
                        self.advance();
                    } else if !self.is_punct(")") {
                        return Err(self.error());
                    }
                }
                self.advance();
                let mut args = args.into_iter().map(Box::new);
                return Ok(Accessor::Decimal(args.next(), args.next()));
            }
            "time" | "time_tz" | "timestamp" | "timestamp_tz" => {
                let kind = match name.as_str() {
                    "time" => TimeMethod::Time,
                    "time_tz" => TimeMethod::TimeTz,
                    "timestamp" => TimeMethod::Timestamp,
                    _ => TimeMethod::TimestampTz,
                };
                self.advance();
                let precision = match self.peek().clone() {
                    Token::Number(n) => {
                        self.advance();
                        let precision = n.to_i64().filter(|_| n.scale() == 0).and_then(|p| i32::try_from(p).ok());
                        Some(precision.ok_or_else(|| {
                            PgError::new(code::SYNTAX_ERROR, "invalid input syntax for type jsonpath")
                        })?)
                    }
                    _ => None,
                };
                self.expect(")")?;
                return Ok(Accessor::Time(kind, precision));
            }
            _ => return Err(self.error()),
        };
        self.advance();
        self.expect(")")?;
        Ok(Accessor::Method(method))
    }

    /// level parses a level bound of a `.**` accessor: a number or `last`.
    fn level(&mut self) -> Result<u32> {
        match self.advance() {
            Token::Identifier(w) if w == "last" => Ok(ANY_LAST),
            Token::Number(n) if n.scale() == 0 => {
                n.to_i64().and_then(|n| u32::try_from(n).ok()).ok_or_else(|| syntax_error(Some(&n.to_string())))
            }
            _ => {
                self.position -= 1;
                Err(self.error())
            }
        }
    }
}

/// check_flags checks the flags of a `like_regex` predicate and compiles its pattern, as Postgres does when it parses
/// one.
fn check_flags(pattern: &str, flags: &str) -> Result<()> {
    if let Some(other) = flags.chars().find(|f| !matches!(f, 'i' | 's' | 'm' | 'x' | 'q')) {
        return Err(PgError {
            detail: Some(format!("Unrecognized flag character \"{other}\" in LIKE_REGEX predicate.")),
            ..PgError::new(code::SYNTAX_ERROR, "invalid input syntax for type jsonpath")
        });
    }
    if flags.contains('x') && !flags.contains('q') {
        return Err(PgError::new(
            code::FEATURE_NOT_SUPPORTED,
            "XQuery \"x\" flag (expanded regular expressions) is not implemented",
        ));
    }
    exec::regex(pattern, flags).map(|_| ())
}

/// check checks where the `@` and `last` items appear, as Postgres does when it flattens a parsed path.
fn check(node: &Node, filters: usize, subscripts: usize) -> Result<()> {
    match node {
        Node::Current if filters == 0 => Err(PgError::new(code::SYNTAX_ERROR, "@ is not allowed in root expressions")),
        Node::Last if subscripts == 0 => {
            Err(PgError::new(code::SYNTAX_ERROR, "LAST is allowed only in array subscripts"))
        }
        Node::Chain(primary, accessors) => {
            check(primary, filters, subscripts)?;
            for accessor in accessors {
                match accessor {
                    Accessor::Index(list) => {
                        for (from, to) in list {
                            check(from, filters, subscripts + 1)?;
                            if let Some(to) = to {
                                check(to, filters, subscripts + 1)?;
                            }
                        }
                    }
                    Accessor::Filter(predicate) => check(predicate, filters + 1, subscripts)?,
                    _ => {}
                }
            }
            Ok(())
        }
        Node::Unary(_, a) | Node::Not(a) | Node::IsUnknown(a) | Node::Exists(a) | Node::LikeRegex(a, ..) => {
            check(a, filters, subscripts)
        }
        Node::Arith(_, a, b) | Node::Compare(_, a, b) | Node::And(a, b) | Node::Or(a, b) | Node::StartsWith(a, b) => {
            check(a, filters, subscripts)?;
            check(b, filters, subscripts)
        }
        _ => Ok(()),
    }
}

/// parse parses a SQL/JSON path, as the jsonpath type's input does.
pub fn parse(text: &str) -> Result<JsonPath> {
    let tokens = lex(text)?;
    let mut parser = Parser { tokens, position: 0 };
    let lax = match parser.peek() {
        Token::Identifier(w) if w == "strict" => {
            parser.advance();
            false
        }
        Token::Identifier(w) if w == "lax" => {
            parser.advance();
            true
        }
        _ => true,
    };
    if matches!(parser.peek(), Token::End) {
        return Err(PgError::new(
            code::INVALID_TEXT_REPRESENTATION,
            format!("invalid input syntax for type jsonpath: \"{text}\""),
        ));
    }
    let expr = parser.or()?;
    if !matches!(parser.peek(), Token::End) {
        return Err(parser.error());
    }
    check(&expr, 0, 0)?;
    Ok(JsonPath { lax, expr })
}
