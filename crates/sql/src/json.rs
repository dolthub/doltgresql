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

//! JSON: parsing and printing json and jsonb as Postgres does, and jsonb's values, ordering, and operators.

use std::cmp::Ordering;

use crate::error::{ErrorObjects, PgError, Result, code};
use crate::numeric::Numeric;

/// Json is a parsed JSON value, whose objects keep their keys in jsonb's order without duplicates.
#[derive(Clone, Debug, PartialEq)]
pub enum Json {
    Null,
    Bool(bool),
    Number(Numeric),
    String(String),
    Array(Vec<Json>),
    Object(Vec<(String, Json)>),
}

/// compare_keys orders object keys as jsonb stores them: shorter keys first, then by their bytes.
pub fn compare_keys(a: &str, b: &str) -> Ordering {
    a.len().cmp(&b.len()).then_with(|| a.as_bytes().cmp(b.as_bytes()))
}

/// Token is a JSON token, with where it starts and ends in the input.
#[derive(Clone, Debug, PartialEq)]
enum Token {
    ObjectStart,
    ObjectEnd,
    ArrayStart,
    ArrayEnd,
    Comma,
    Colon,
    String(String),
    Number(String),
    True,
    False,
    Null,
    End,
}

/// Lexer reads JSON tokens as Postgres' json lexer does.
struct Lexer<'a> {
    input: &'a str,
    position: usize,
    /// Where the current token starts and ends, and where the previous token ended.
    start: usize,
    end: usize,
    previous_end: usize,
    line: usize,
    line_start: usize,
    /// Whether NUL characters, which jsonb cannot hold, are rejected.
    jsonb: bool,
}

impl<'a> Lexer<'a> {
    /// error returns Postgres' invalid json error with a detail and the context of the current token.
    fn error(&self, detail: String) -> PgError {
        let mut context_start = self.line_start;
        while self.end.saturating_sub(context_start) >= 50 {
            context_start += self.input[context_start..].chars().next().map_or(1, char::len_utf8);
        }
        let prefix = if context_start > self.line_start { "..." } else { "" };
        let rest = &self.input[self.end.min(self.input.len())..];
        let suffix = if self.end < self.input.len() && !rest.starts_with(['\n', '\r']) { "..." } else { "" };
        let context = format!(
            "JSON data, line {}: {prefix}{}{suffix}",
            self.line,
            &self.input[context_start..self.end.min(self.input.len())]
        );
        PgError {
            detail: Some(detail),
            objects: Some(Box::new(ErrorObjects { where_: Some(context), ..ErrorObjects::default() })),
            ..PgError::new(code::INVALID_TEXT_REPRESENTATION, "invalid input syntax for type json")
        }
    }

    /// word returns the text from the token's start through the following letters, digits, and underscores, as
    /// Postgres reports an invalid token.
    fn invalid_word(&mut self, from: usize) -> PgError {
        let bytes = self.input.as_bytes();
        let mut end = from;
        while end < bytes.len() && (bytes[end].is_ascii_alphanumeric() || bytes[end] == b'_' || bytes[end] >= 0x80) {
            end += 1;
        }
        if end == from {
            end += self.input[from..].chars().next().map_or(1, char::len_utf8);
        }
        while !self.input.is_char_boundary(end) {
            end += 1;
        }
        self.end = end;
        self.error(format!("Token \"{}\" is invalid.", &self.input[self.start..end]))
    }

    /// next reads the next token.
    fn next(&mut self) -> Result<Token> {
        self.previous_end = self.end;
        let bytes = self.input.as_bytes();
        while self.position < bytes.len() && matches!(bytes[self.position], b' ' | b'\t' | b'\n' | b'\r') {
            if bytes[self.position] == b'\n' {
                self.line += 1;
                self.line_start = self.position + 1;
            }
            self.position += 1;
        }
        self.start = self.position;
        if self.position >= bytes.len() {
            self.end = self.position;
            return Ok(Token::End);
        }
        let c = bytes[self.position];
        let single = |lexer: &mut Self, token: Token| {
            lexer.position += 1;
            lexer.end = lexer.position;
            Ok(token)
        };
        match c {
            b'{' => single(self, Token::ObjectStart),
            b'}' => single(self, Token::ObjectEnd),
            b'[' => single(self, Token::ArrayStart),
            b']' => single(self, Token::ArrayEnd),
            b',' => single(self, Token::Comma),
            b':' => single(self, Token::Colon),
            b'"' => self.string(),
            b'-' | b'0'..=b'9' => self.number(),
            _ => {
                let mut end = self.position;
                while end < bytes.len()
                    && (bytes[end].is_ascii_alphanumeric() || bytes[end] == b'_' || bytes[end] >= 0x80)
                {
                    end += 1;
                }
                let word = &self.input[self.position..end];
                let token = match word {
                    "true" => Token::True,
                    "false" => Token::False,
                    "null" => Token::Null,
                    _ => return Err(self.invalid_word(self.position)),
                };
                self.position = end;
                self.end = end;
                Ok(token)
            }
        }
    }

    /// number reads a number, which must follow JSON's grammar.
    fn number(&mut self) -> Result<Token> {
        let bytes = self.input.as_bytes();
        let start = self.position;
        let mut p = start;
        if bytes[p] == b'-' {
            p += 1;
        }
        let digits = |p: &mut usize| {
            let s = *p;
            while *p < bytes.len() && bytes[*p].is_ascii_digit() {
                *p += 1;
            }
            *p > s
        };
        let mut valid = true;
        if p < bytes.len() && bytes[p] == b'0' {
            p += 1;
        } else if !digits(&mut p) {
            valid = false;
        }
        if valid && p < bytes.len() && bytes[p] == b'.' {
            p += 1;
            valid = digits(&mut p);
        }
        if valid && p < bytes.len() && (bytes[p] == b'e' || bytes[p] == b'E') {
            p += 1;
            if p < bytes.len() && (bytes[p] == b'+' || bytes[p] == b'-') {
                p += 1;
            }
            valid = digits(&mut p);
        }
        if !valid || (p < bytes.len() && (bytes[p].is_ascii_alphanumeric() || bytes[p] == b'_' || bytes[p] >= 0x80)) {
            return Err(self.invalid_word(start));
        }
        self.position = p;
        self.end = p;
        Ok(Token::Number(self.input[start..p].to_string()))
    }

    /// string reads a string, decoding its escapes.
    fn string(&mut self) -> Result<Token> {
        let mut out = String::new();
        let mut chars = self.input[self.position + 1..].char_indices();
        let base = self.position + 1;
        loop {
            let Some((i, c)) = chars.next() else {
                self.end = self.input.len();
                return Err(self.error("The input string ended unexpectedly.".into()));
            };
            match c {
                '"' => {
                    self.position = base + i + 1;
                    self.end = self.position;
                    return Ok(Token::String(out));
                }
                '\\' => {
                    let Some((j, e)) = chars.next() else {
                        self.end = self.input.len();
                        return Err(self.error("The input string ended unexpectedly.".into()));
                    };
                    match e {
                        '"' => out.push('"'),
                        '\\' => out.push('\\'),
                        '/' => out.push('/'),
                        'b' => out.push('\u{8}'),
                        'f' => out.push('\u{c}'),
                        'n' => out.push('\n'),
                        'r' => out.push('\r'),
                        't' => out.push('\t'),
                        'u' => {
                            let code = self.hex4(base + j + 1)?;
                            for _ in 0..4 {
                                chars.next();
                            }
                            if (0xD800..0xDC00).contains(&code) {
                                let rest = &self.input[base + j + 5..];
                                if !rest.starts_with("\\u") {
                                    self.end = base + j + 5;
                                    return Err(
                                        self.error("Unicode high surrogate must not follow a high surrogate.".into())
                                    );
                                }
                                let low = self.hex4(base + j + 7)?;
                                if !(0xDC00..0xE000).contains(&low) {
                                    self.end = base + j + 11;
                                    return Err(
                                        self.error("Unicode low surrogate must follow a high surrogate.".into())
                                    );
                                }
                                for _ in 0..6 {
                                    chars.next();
                                }
                                let combined = 0x10000 + ((code - 0xD800) << 10) + (low - 0xDC00);
                                out.push(char::from_u32(combined).unwrap_or('\u{FFFD}'));
                            } else if (0xDC00..0xE000).contains(&code) {
                                self.end = base + j + 5;
                                return Err(self.error("Unicode low surrogate must follow a high surrogate.".into()));
                            } else if code == 0 && self.jsonb {
                                self.end = base + j + 5;
                                return Err(PgError {
                                    detail: Some("\\u0000 cannot be converted to text.".into()),
                                    ..PgError::new(
                                        code::UNTRANSLATABLE_CHARACTER,
                                        "unsupported Unicode escape sequence",
                                    )
                                });
                            } else {
                                out.push(char::from_u32(code).unwrap_or('\u{FFFD}'));
                            }
                        }
                        other => {
                            self.end = base + j + other.len_utf8();
                            return Err(self.error(format!("Escape sequence \"\\{other}\" is invalid.")));
                        }
                    }
                }
                c if (c as u32) < 0x20 => {
                    self.end = base + i + 1;
                    return Err(self.error(format!("Character with value 0x{:02x} must be escaped.", c as u32)));
                }
                c => out.push(c),
            }
        }
    }

    /// hex4 reads the four hex digits of a \u escape.
    fn hex4(&mut self, at: usize) -> Result<u32> {
        let digits = self.input.get(at..at + 4).filter(|d| d.bytes().all(|b| b.is_ascii_hexdigit()));
        match digits {
            Some(d) => Ok(u32::from_str_radix(d, 16).unwrap_or(0)),
            None => {
                let mut end = at;
                let bytes = self.input.as_bytes();
                while end < bytes.len() && end < at + 4 && bytes[end].is_ascii_hexdigit() {
                    end += 1;
                }
                self.end = end;
                Err(self.error("\"\\u\" must be followed by four hexadecimal digits.".into()))
            }
        }
    }
}

/// describe returns how an error message shows a token.
fn describe(lexer: &Lexer<'_>, token: &Token) -> String {
    match token {
        Token::End => String::new(),
        _ => lexer.input[lexer.start..lexer.end].to_string(),
    }
}

/// Parser builds JSON values from tokens.
struct Parser<'a> {
    lexer: Lexer<'a>,
    token: Token,
}

impl Parser<'_> {
    /// advance reads the next token.
    fn advance(&mut self) -> Result<()> {
        self.token = self.lexer.next()?;
        Ok(())
    }

    /// unexpected returns the error for an unexpected token, or for the input ending.
    fn unexpected(&self, expected: &str) -> PgError {
        if self.token == Token::End {
            return self.lexer.error("The input string ended unexpectedly.".into());
        }
        self.lexer.error(format!("Expected {expected}, but found \"{}\".", describe(&self.lexer, &self.token)))
    }

    /// value parses a value.
    fn value(&mut self) -> Result<Json> {
        let value = match std::mem::replace(&mut self.token, Token::End) {
            Token::ObjectStart => {
                self.advance()?;
                let mut items: Vec<(String, Json)> = Vec::new();
                if self.token == Token::ObjectEnd {
                    self.advance()?;
                    return Ok(Json::Object(items));
                }
                loop {
                    let Token::String(key) = std::mem::replace(&mut self.token, Token::End) else {
                        self.token = self.lexer.next_again()?;
                        return Err(self.unexpected("string"));
                    };
                    self.advance()?;
                    if self.token != Token::Colon {
                        return Err(self.unexpected("\":\""));
                    }
                    self.advance()?;
                    let value = self.value()?;
                    items.push((key, value));
                    match self.token {
                        Token::Comma => self.advance()?,
                        Token::ObjectEnd => {
                            self.advance()?;
                            break;
                        }
                        _ => return Err(self.unexpected("\",\" or \"}\"")),
                    }
                }
                return Ok(Json::Object(items));
            }
            Token::ArrayStart => {
                self.advance()?;
                let mut values = Vec::new();
                if self.token == Token::ArrayEnd {
                    self.advance()?;
                    return Ok(Json::Array(values));
                }
                loop {
                    values.push(self.value()?);
                    match self.token {
                        Token::Comma => self.advance()?,
                        Token::ArrayEnd => {
                            self.advance()?;
                            break;
                        }
                        _ => return Err(self.unexpected("\",\" or \"]\"")),
                    }
                }
                return Ok(Json::Array(values));
            }
            Token::String(s) => Json::String(s),
            Token::Number(n) => Json::Number(Numeric::parse(&n)?),
            Token::True => Json::Bool(true),
            Token::False => Json::Bool(false),
            Token::Null => Json::Null,
            other => {
                self.token = other;
                return Err(self.unexpected("JSON value"));
            }
        };
        self.advance()?;
        Ok(value)
    }
}

impl Lexer<'_> {
    /// next_again re-reads the token that a parser took, so that an error can describe it.
    fn next_again(&mut self) -> Result<Token> {
        self.position = self.start;
        self.next()
    }
}

/// Raw is a parsed json value that keeps the text of each part as written, which json's operators return.
#[derive(Clone, Debug, PartialEq)]
pub struct Raw<'a> {
    pub text: &'a str,
    pub kind: RawKind<'a>,
}

/// RawKind is what a Raw value is.
#[derive(Clone, Debug, PartialEq)]
pub enum RawKind<'a> {
    Object(Vec<(String, Raw<'a>)>),
    Array(Vec<Raw<'a>>),
    Scalar(Json),
}

impl<'a> Parser<'a> {
    /// raw parses a value, keeping its text.
    fn raw(&mut self) -> Result<Raw<'a>> {
        let input = self.lexer.input;
        let start = self.lexer.start;
        let kind = match self.token {
            Token::ObjectStart => {
                self.advance()?;
                let mut items = Vec::new();
                if self.token != Token::ObjectEnd {
                    loop {
                        let Token::String(key) = std::mem::replace(&mut self.token, Token::End) else {
                            self.token = self.lexer.next_again()?;
                            return Err(self.unexpected("string"));
                        };
                        self.advance()?;
                        if self.token != Token::Colon {
                            return Err(self.unexpected("\":\""));
                        }
                        self.advance()?;
                        items.push((key, self.raw()?));
                        match self.token {
                            Token::Comma => self.advance()?,
                            Token::ObjectEnd => break,
                            _ => return Err(self.unexpected("\",\" or \"}\"")),
                        }
                    }
                }
                RawKind::Object(items)
            }
            Token::ArrayStart => {
                self.advance()?;
                let mut values = Vec::new();
                if self.token != Token::ArrayEnd {
                    loop {
                        values.push(self.raw()?);
                        match self.token {
                            Token::Comma => self.advance()?,
                            Token::ArrayEnd => break,
                            _ => return Err(self.unexpected("\",\" or \"]\"")),
                        }
                    }
                }
                RawKind::Array(values)
            }
            _ => {
                let value = self.value()?;
                let end = self.lexer.previous_end;
                return Ok(Raw { text: &input[start..end], kind: RawKind::Scalar(value) });
            }
        };
        let end = self.lexer.end;
        self.advance()?;
        Ok(Raw { text: &input[start..end], kind })
    }
}

/// parse_raw parses json text, keeping the text of each part.
pub fn parse_raw(text: &str) -> Result<Raw<'_>> {
    let mut lexer =
        Lexer { input: text, position: 0, start: 0, end: 0, previous_end: 0, line: 1, line_start: 0, jsonb: false };
    let token = lexer.next()?;
    let mut parser = Parser { lexer, token };
    let value = parser.raw()?;
    if parser.token != Token::End {
        return Err(parser.unexpected("end of input"));
    }
    Ok(value)
}

impl<'a> Raw<'a> {
    /// get returns an object's last value for a key.
    pub fn get(&self, key: &str) -> Option<&Raw<'a>> {
        match &self.kind {
            RawKind::Object(items) => items.iter().rev().find(|(k, _)| k == key).map(|(_, v)| v),
            _ => None,
        }
    }

    /// index returns an array's element, counting back from the end for a negative index.
    pub fn index(&self, index: i64) -> Option<&Raw<'a>> {
        match &self.kind {
            RawKind::Array(values) => {
                let i = if index < 0 { values.len() as i64 + index } else { index };
                usize::try_from(i).ok().and_then(|i| values.get(i))
            }
            _ => None,
        }
    }

    /// path follows a path of keys and array indexes.
    pub fn path(&self, path: &[Option<String>]) -> Option<&Raw<'a>> {
        let mut current = self;
        for step in path {
            let step = step.as_ref()?;
            current = match current.kind {
                RawKind::Object(_) => current.get(step)?,
                RawKind::Array(_) => current.index(step.trim().parse::<i64>().ok()?)?,
                RawKind::Scalar(_) => return None,
            };
        }
        Some(current)
    }

    /// scalar_text returns the value as ->> returns it: a string's contents, null as None, and others as written.
    pub fn scalar_text(&self) -> Option<String> {
        match &self.kind {
            RawKind::Scalar(Json::Null) => None,
            RawKind::Scalar(Json::String(s)) => Some(s.clone()),
            _ => Some(self.text.to_string()),
        }
    }

    /// type_name returns the value's type as json_typeof names it.
    pub fn type_name(&self) -> &'static str {
        match &self.kind {
            RawKind::Object(_) => "object",
            RawKind::Array(_) => "array",
            RawKind::Scalar(json) => json.type_name(),
        }
    }
}

/// parse parses JSON text, as json input checks it and jsonb input reads it, with jsonb's key order and the last of
/// any duplicate keys when asked.
pub fn parse(text: &str, jsonb: bool) -> Result<Json> {
    let mut lexer =
        Lexer { input: text, position: 0, start: 0, end: 0, previous_end: 0, line: 1, line_start: 0, jsonb };
    let token = lexer.next()?;
    let mut parser = Parser { lexer, token };
    let value = parser.value()?;
    if parser.token != Token::End {
        return Err(parser.unexpected("end of input"));
    }
    Ok(if jsonb { normalize(value) } else { value })
}

/// normalize sorts objects' keys in jsonb's order, keeping the last of duplicate keys.
pub fn normalize(value: Json) -> Json {
    match value {
        Json::Array(values) => Json::Array(values.into_iter().map(normalize).collect()),
        Json::Object(items) => {
            let mut sorted: Vec<(String, Json)> = Vec::with_capacity(items.len());
            for (key, value) in items {
                let value = normalize(value);
                match sorted.binary_search_by(|(k, _)| compare_keys(k, &key)) {
                    Ok(i) => sorted[i].1 = value,
                    Err(i) => sorted.insert(i, (key, value)),
                }
            }
            Json::Object(sorted)
        }
        other => other,
    }
}

/// escape writes a JSON string literal, escaping characters as Postgres' escape_json does.
pub fn escape(out: &mut String, text: &str) {
    out.push('"');
    for c in text.chars() {
        match c {
            '"' => out.push_str("\\\""),
            '\\' => out.push_str("\\\\"),
            '\u{8}' => out.push_str("\\b"),
            '\u{c}' => out.push_str("\\f"),
            '\n' => out.push_str("\\n"),
            '\r' => out.push_str("\\r"),
            '\t' => out.push_str("\\t"),
            c if (c as u32) < 0x20 => out.push_str(&format!("\\u{:04x}", c as u32)),
            c => out.push(c),
        }
    }
    out.push('"');
}

impl Json {
    /// write prints the value as jsonb's output does, with a space after each comma and colon.
    pub fn write(&self, out: &mut String) {
        match self {
            Json::Null => out.push_str("null"),
            Json::Bool(b) => out.push_str(if *b { "true" } else { "false" }),
            Json::Number(n) => out.push_str(&n.to_string()),
            Json::String(s) => escape(out, s),
            Json::Array(values) => {
                out.push('[');
                for (i, v) in values.iter().enumerate() {
                    if i > 0 {
                        out.push_str(", ");
                    }
                    v.write(out);
                }
                out.push(']');
            }
            Json::Object(items) => {
                out.push('{');
                for (i, (k, v)) in items.iter().enumerate() {
                    if i > 0 {
                        out.push_str(", ");
                    }
                    escape(out, k);
                    out.push_str(": ");
                    v.write(out);
                }
                out.push('}');
            }
        }
    }

    /// to_text returns the value as jsonb prints it.
    pub fn to_text(&self) -> String {
        let mut out = String::new();
        self.write(&mut out);
        out
    }

    /// plain prints the value without spaces, with object keys in their order, as json functions such as to_json
    /// write arrays and records.
    pub fn plain(&self) -> String {
        let mut out = String::new();
        self.write_plain(&mut out);
        out
    }

    /// write_plain writes the value without spaces, keeping object keys in their order.
    fn write_plain(&self, out: &mut String) {
        match self {
            Json::Array(values) => {
                out.push('[');
                for (i, v) in values.iter().enumerate() {
                    if i > 0 {
                        out.push(',');
                    }
                    v.write_plain(out);
                }
                out.push(']');
            }
            Json::Object(items) => {
                out.push('{');
                for (i, (k, v)) in items.iter().enumerate() {
                    if i > 0 {
                        out.push(',');
                    }
                    escape(out, k);
                    out.push(':');
                    v.write_plain(out);
                }
                out.push('}');
            }
            other => other.write(out),
        }
    }

    /// compact prints the value without spaces, with object keys in byte order, as Doltgres stores jsonb.
    pub fn compact(&self) -> String {
        let mut out = String::new();
        self.write_compact(&mut out);
        out
    }

    /// write_compact writes the value without spaces, with object keys in byte order.
    fn write_compact(&self, out: &mut String) {
        match self {
            Json::Array(values) => {
                out.push('[');
                for (i, v) in values.iter().enumerate() {
                    if i > 0 {
                        out.push(',');
                    }
                    v.write_compact(out);
                }
                out.push(']');
            }
            Json::Object(items) => {
                let mut sorted: Vec<&(String, Json)> = items.iter().collect();
                sorted.sort_by(|a, b| a.0.as_bytes().cmp(b.0.as_bytes()));
                out.push('{');
                for (i, (k, v)) in sorted.into_iter().enumerate() {
                    if i > 0 {
                        out.push(',');
                    }
                    escape(out, k);
                    out.push(':');
                    v.write_compact(out);
                }
                out.push('}');
            }
            other => other.write(out),
        }
    }

    /// type_name returns the value's type as jsonb_typeof names it.
    pub fn type_name(&self) -> &'static str {
        match self {
            Json::Null => "null",
            Json::Bool(_) => "boolean",
            Json::Number(_) => "number",
            Json::String(_) => "string",
            Json::Array(_) => "array",
            Json::Object(_) => "object",
        }
    }

    /// rank orders the kinds of values as jsonb's ordering does.
    fn rank(&self) -> u8 {
        match self {
            Json::Null => 0,
            Json::String(_) => 1,
            Json::Number(_) => 2,
            Json::Bool(_) => 3,
            Json::Array(_) => 4,
            Json::Object(_) => 5,
        }
    }

    /// get returns an object's value for a key.
    pub fn get(&self, key: &str) -> Option<&Json> {
        match self {
            Json::Object(items) => items.iter().find(|(k, _)| k == key).map(|(_, v)| v),
            _ => None,
        }
    }

    /// index returns an array's element at an index, counting back from the end for a negative one.
    pub fn index(&self, index: i64) -> Option<&Json> {
        match self {
            Json::Array(values) => {
                let i = if index < 0 { values.len() as i64 + index } else { index };
                usize::try_from(i).ok().and_then(|i| values.get(i))
            }
            _ => None,
        }
    }

    /// path follows a path of keys and array indexes.
    pub fn path(&self, path: &[Option<String>]) -> Option<&Json> {
        let mut current = self;
        for step in path {
            let step = step.as_ref()?;
            current = match current {
                Json::Object(_) => current.get(step)?,
                Json::Array(_) => current.index(step.trim().parse::<i64>().ok()?)?,
                _ => return None,
            };
        }
        Some(current)
    }

    /// scalar_text returns the value as ->> returns it: a string's contents, null as None, and others as JSON text.
    pub fn scalar_text(&self) -> Option<String> {
        match self {
            Json::Null => None,
            Json::String(s) => Some(s.clone()),
            other => Some(other.to_text()),
        }
    }

    /// contains reports whether the value contains another, as jsonb's @> does.
    pub fn contains(&self, other: &Json) -> bool {
        match (self, other) {
            (Json::Object(items), Json::Object(wanted)) => {
                wanted.iter().all(|(k, v)| items.iter().any(|(k2, v2)| k2 == k && v2.contains(v)))
            }
            (Json::Array(values), Json::Array(wanted)) => wanted.iter().all(|w| match w {
                Json::Array(_) | Json::Object(_) => values.iter().any(|v| v.contains(w)),
                scalar => values.iter().any(|v| compare(v, scalar) == Ordering::Equal),
            }),
            (Json::Array(values), scalar) if !matches!(scalar, Json::Object(_) | Json::Array(_)) => {
                values.iter().any(|v| compare(v, scalar) == Ordering::Equal)
            }
            (a, b) => compare(a, b) == Ordering::Equal,
        }
    }
}

/// compare orders jsonb values as Postgres' jsonb btree ordering does.
pub fn compare(a: &Json, b: &Json) -> Ordering {
    match (a, b) {
        (Json::Array(x), Json::Array(y)) => x
            .len()
            .cmp(&y.len())
            .then_with(|| x.iter().zip(y).map(|(p, q)| compare(p, q)).find(|o| o.is_ne()).unwrap_or(Ordering::Equal)),
        (Json::Object(x), Json::Object(y)) => x.len().cmp(&y.len()).then_with(|| {
            for ((k1, v1), (k2, v2)) in x.iter().zip(y) {
                let o = compare_keys(k1, k2).then_with(|| compare(v1, v2));
                if o.is_ne() {
                    return o;
                }
            }
            Ordering::Equal
        }),
        (Json::String(x), Json::String(y)) => x.as_bytes().cmp(y.as_bytes()),
        (Json::Number(x), Json::Number(y)) => x.cmp_numeric(y),
        (Json::Bool(x), Json::Bool(y)) => x.cmp(y),
        (Json::Null, Json::Null) => Ordering::Equal,
        _ => a.rank().cmp(&b.rank()),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn jsonb_parses_and_prints_as_postgres_does() {
        let value = parse(r#"{"z": 1,   "a": [1.50, "x\n", null, true], "bb": {}, "a": 2}"#, true).unwrap();
        assert_eq!(value.to_text(), r#"{"a": 2, "z": 1, "bb": {}}"#);
        assert_eq!(value.compact(), r#"{"a":2,"bb":{},"z":1}"#);
        assert_eq!(parse("[1.50, -0, 1e3]", true).unwrap().to_text(), "[1.50, 0, 1000]");
        let err = parse("{\"a\":1,}", false).unwrap_err();
        assert_eq!(err.detail.as_deref(), Some("Expected string, but found \"}\"."));
        assert_eq!(err.objects.unwrap().where_.as_deref(), Some("JSON data, line 1: {\"a\":1,}"));
        let err = parse("01", false).unwrap_err();
        assert_eq!(err.detail.as_deref(), Some("Token \"01\" is invalid."));
        let err = parse("[1, 2", false).unwrap_err();
        assert_eq!(err.detail.as_deref(), Some("The input string ended unexpectedly."));
    }
}
