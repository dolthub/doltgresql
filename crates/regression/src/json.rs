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

use crate::pgx::Any;

/// unmarshal matches json.Unmarshal into an `any`, returning None where Go returns an error.
pub fn unmarshal(src: &[u8]) -> Option<Any> {
    let mut parser = Parser { s: src, i: 0 };
    parser.space();
    let value = parser.value()?;
    parser.space();
    (parser.i == src.len()).then_some(value)
}

struct Parser<'a> {
    s: &'a [u8],
    i: usize,
}

impl Parser<'_> {
    fn space(&mut self) {
        while self.i < self.s.len() && matches!(self.s[self.i], b' ' | b'\t' | b'\r' | b'\n') {
            self.i += 1;
        }
    }

    fn eat(&mut self, byte: u8) -> Option<()> {
        (self.s.get(self.i) == Some(&byte)).then(|| self.i += 1)
    }

    fn value(&mut self) -> Option<Any> {
        match *self.s.get(self.i)? {
            b'{' => {
                self.i += 1;
                let mut entries: Vec<(String, Any)> = Vec::new();
                self.space();
                if self.eat(b'}').is_some() {
                    return Some(Any::Map(entries));
                }
                loop {
                    self.space();
                    let key = self.string()?;
                    self.space();
                    self.eat(b':')?;
                    self.space();
                    let value = self.value()?;
                    match entries.iter_mut().find(|(k, _)| *k == key) {
                        Some(entry) => entry.1 = value,
                        None => entries.push((key, value)),
                    }
                    self.space();
                    if self.eat(b',').is_some() {
                        continue;
                    }
                    self.eat(b'}')?;
                    entries.sort_by(|a, b| a.0.as_bytes().cmp(b.0.as_bytes()));
                    return Some(Any::Map(entries));
                }
            }
            b'[' => {
                self.i += 1;
                let mut values = Vec::new();
                self.space();
                if self.eat(b']').is_some() {
                    return Some(Any::Slice(values));
                }
                loop {
                    self.space();
                    values.push(self.value()?);
                    self.space();
                    if self.eat(b',').is_some() {
                        continue;
                    }
                    self.eat(b']')?;
                    return Some(Any::Slice(values));
                }
            }
            b'"' => self.string().map(Any::String),
            b't' => self.literal(b"true", Any::Bool(true)),
            b'f' => self.literal(b"false", Any::Bool(false)),
            b'n' => self.literal(b"null", Any::Nil),
            b'-' | b'0'..=b'9' => self.number(),
            _ => None,
        }
    }

    fn literal(&mut self, text: &[u8], value: Any) -> Option<Any> {
        self.s[self.i..].starts_with(text).then(|| {
            self.i += text.len();
            value
        })
    }

    fn digits(&mut self) -> usize {
        let start = self.i;
        while self.i < self.s.len() && self.s[self.i].is_ascii_digit() {
            self.i += 1;
        }
        self.i - start
    }

    fn number(&mut self) -> Option<Any> {
        let start = self.i;
        let _ = self.eat(b'-');
        match self.s.get(self.i)? {
            b'0' => self.i += 1,
            b'1'..=b'9' => {
                self.digits();
            }
            _ => return None,
        }
        if self.eat(b'.').is_some() && self.digits() == 0 {
            return None;
        }
        if matches!(self.s.get(self.i), Some(b'e' | b'E')) {
            self.i += 1;
            if matches!(self.s.get(self.i), Some(b'+' | b'-')) {
                self.i += 1;
            }
            if self.digits() == 0 {
                return None;
            }
        }
        let text = std::str::from_utf8(&self.s[start..self.i]).ok()?;
        let value: f64 = text.parse().ok()?;
        value.is_finite().then_some(Any::Float64(value))
    }

    fn hex4(&self, at: usize) -> Option<u32> {
        let digits = self.s.get(at..at + 4)?;
        let text = std::str::from_utf8(digits).ok()?;
        u32::from_str_radix(text, 16).ok().filter(|_| digits.iter().all(u8::is_ascii_hexdigit))
    }

    /// string reads a quoted string, replacing invalid UTF-8 and unpaired surrogates the way Go does.
    fn string(&mut self) -> Option<String> {
        self.eat(b'"')?;
        let mut out: Vec<u8> = Vec::new();
        loop {
            let c = *self.s.get(self.i)?;
            match c {
                b'"' => {
                    self.i += 1;
                    return Some(go_string(&out));
                }
                b'\\' => {
                    let escape = *self.s.get(self.i + 1)?;
                    self.i += 2;
                    match escape {
                        b'"' | b'\\' | b'/' => out.push(escape),
                        b'b' => out.push(8),
                        b'f' => out.push(12),
                        b'n' => out.push(b'\n'),
                        b'r' => out.push(b'\r'),
                        b't' => out.push(b'\t'),
                        b'u' => {
                            let first = self.hex4(self.i)?;
                            self.i += 4;
                            let mut rune = char::from_u32(first);
                            if (0xD800..0xE000).contains(&first) {
                                rune = Some('\u{FFFD}');
                                if first < 0xDC00
                                    && self.s.get(self.i) == Some(&b'\\')
                                    && self.s.get(self.i + 1) == Some(&b'u')
                                    && let Some(second) = self.hex4(self.i + 2)
                                    && (0xDC00..0xE000).contains(&second)
                                {
                                    self.i += 6;
                                    rune = char::from_u32(0x10000 + ((first - 0xD800) << 10) + (second - 0xDC00));
                                }
                            }
                            let mut buf = [0u8; 4];
                            out.extend_from_slice(rune?.encode_utf8(&mut buf).as_bytes());
                        }
                        _ => return None,
                    }
                }
                0..=0x1f => return None,
                _ => {
                    out.push(c);
                    self.i += 1;
                }
            }
        }
    }
}

/// go_string converts bytes to text the way Go's decoders coerce invalid UTF-8, with one U+FFFD per invalid byte.
pub fn go_string(bytes: &[u8]) -> String {
    let mut out = String::with_capacity(bytes.len());
    for chunk in bytes.utf8_chunks() {
        out.push_str(chunk.valid());
        for _ in chunk.invalid() {
            out.push('\u{FFFD}');
        }
    }
    out
}

/// marshal_map matches json.Marshal of a map[string]any.
pub fn marshal_map(entries: &[(String, Any)]) -> String {
    let mut out = String::new();
    marshal(&Any::Map(entries.to_vec()), &mut out);
    out
}

fn marshal(value: &Any, out: &mut String) {
    match value {
        Any::Nil => out.push_str("null"),
        Any::Bool(v) => out.push_str(if *v { "true" } else { "false" }),
        Any::Float64(v) => out.push_str(&float(*v)),
        Any::String(v) => string(v, out),
        Any::Map(entries) => {
            out.push('{');
            for (i, (key, value)) in entries.iter().enumerate() {
                if i > 0 {
                    out.push(',');
                }
                string(key, out);
                out.push(':');
                marshal(value, out);
            }
            out.push('}');
        }
        Any::Slice(values) => {
            out.push('[');
            for (i, value) in values.iter().enumerate() {
                if i > 0 {
                    out.push(',');
                }
                marshal(value, out);
            }
            out.push(']');
        }
        other => unreachable!("not a JSON value: {other:?}"),
    }
}

/// float matches encoding/json's float64 encoding.
fn float(value: f64) -> String {
    let abs = value.abs();
    if abs != 0.0 && !(1e-6..1e21).contains(&abs) {
        let text = format!("{value:e}");
        let (mantissa, exponent) = text.split_once('e').unwrap();
        let exponent: i32 = exponent.parse().unwrap();
        if exponent < 0 { format!("{mantissa}e-{}", -exponent) } else { format!("{mantissa}e+{exponent:02}") }
    } else {
        format!("{value}")
    }
}

/// string matches encoding/json's HTML-safe string encoding.
fn string(value: &str, out: &mut String) {
    const HEX: &[u8; 16] = b"0123456789abcdef";
    out.push('"');
    for c in value.chars() {
        match c {
            '\\' | '"' => {
                out.push('\\');
                out.push(c);
            }
            '\u{8}' => out.push_str("\\b"),
            '\u{c}' => out.push_str("\\f"),
            '\n' => out.push_str("\\n"),
            '\r' => out.push_str("\\r"),
            '\t' => out.push_str("\\t"),
            '\u{0}'..='\u{1f}' | '<' | '>' | '&' => {
                let b = c as u8;
                out.push_str("\\u00");
                out.push(HEX[(b >> 4) as usize] as char);
                out.push(HEX[(b & 0xf) as usize] as char);
            }
            '\u{2028}' => out.push_str("\\u2028"),
            '\u{2029}' => out.push_str("\\u2029"),
            _ => out.push(c),
        }
    }
    out.push('"');
}
