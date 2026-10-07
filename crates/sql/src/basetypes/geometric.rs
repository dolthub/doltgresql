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

//! The geometric types, read and written as Postgres' geo_ops.c does and stored as their binary formats: point, lseg,
//! box, path, polygon, line, and circle.

use std::cmp::Ordering;

use crate::error::{PgError, Result, code};
use crate::extensions::BaseType;
use crate::types::Value;

/// EPSILON is the tolerance of Postgres' fuzzy geometric comparisons.
const EPSILON: f64 = 1.0e-06;

/// geometric declares a geometric type with its text and binary formats.
const fn geometric(
    name: &'static str,
    input: fn(&str, i32) -> Result<Vec<u8>>,
    output: fn(&[u8]) -> String,
    receive: fn(&[u8], i32) -> Result<Vec<u8>>,
) -> BaseType {
    BaseType {
        name,
        input,
        output,
        receive,
        send: <[u8]>::to_vec,
        typmod_in: |_| Ok(-1),
        typmod: |_, _| Ok(()),
        compare: <[u8]>::cmp,
        vector: None,
    }
}

/// POINT is the point type.
pub const POINT: BaseType = geometric("point", point_in, point_out, |b, _| fixed(b, 2));

/// LSEG is the lseg type.
pub const LSEG: BaseType = geometric("lseg", lseg_in, lseg_out, |b, _| fixed(b, 4));

/// BOX is the box type.
pub const BOX: BaseType = geometric("box", box_in, box_out, box_recv);

/// PATH is the path type.
pub const PATH: BaseType = geometric("path", path_in, path_out, path_recv);

/// POLYGON is the polygon type.
pub const POLYGON: BaseType = geometric("polygon", polygon_in, polygon_out, polygon_recv);

/// LINE is the line type.
pub const LINE: BaseType = geometric("line", line_in, line_out, line_recv);

/// CIRCLE is the circle type.
pub const CIRCLE: BaseType = geometric("circle", circle_in, circle_out, circle_recv);

/// Scanner reads the text format of a geometric value, keeping the whole text for errors.
struct Scanner<'a> {
    text: &'a str,
    position: usize,
    type_name: &'static str,
}

impl<'a> Scanner<'a> {
    /// new starts reading a value's text.
    fn new(text: &'a str, type_name: &'static str) -> Scanner<'a> {
        Scanner { text, position: 0, type_name }
    }

    /// invalid returns Postgres' error for text that is not a value of the type.
    fn invalid(&self) -> PgError {
        PgError::new(
            code::INVALID_TEXT_REPRESENTATION,
            format!("invalid input syntax for type {}: \"{}\"", self.type_name, self.text),
        )
    }

    /// peek returns the next byte, or 0 at the end.
    fn peek(&self) -> u8 {
        self.text.as_bytes().get(self.position).copied().unwrap_or(0)
    }

    /// skip_space moves past whitespace.
    fn skip_space(&mut self) {
        while self.peek().is_ascii_whitespace() {
            self.position += 1;
        }
    }

    /// eat moves past the byte when it is next, reporting whether it was.
    fn eat(&mut self, byte: u8) -> bool {
        let found = self.peek() == byte;
        if found {
            self.position += 1;
        }
        found
    }

    /// expect moves past the byte, failing when another comes next.
    fn expect(&mut self, byte: u8) -> Result<()> {
        if self.peek() != byte {
            return Err(self.invalid());
        }
        self.position += 1;
        Ok(())
    }

    /// finish fails unless the text has ended.
    fn finish(&self) -> Result<()> {
        if self.position < self.text.len() {
            return Err(self.invalid());
        }
        Ok(())
    }

    /// float reads a float8 with the whitespace around it, as float8in_internal does.
    fn float(&mut self) -> Result<f64> {
        self.skip_space();
        let rest = &self.text[self.position..];
        let length = float_prefix(rest);
        if length == 0 {
            return Err(self.invalid());
        }
        let value = crate::cast::cast_value(
            Value::Text(rest[..length].to_string()),
            crate::expr::typ(crate::oid::FLOAT8),
            true,
        )
        .map_err(|_| self.invalid())?;
        self.position += length;
        self.skip_space();
        match value {
            Value::Float8(value) => Ok(value),
            _ => Err(self.invalid()),
        }
    }

    /// pair reads a point, written with or without parentheses, as pair_decode does.
    fn pair(&mut self) -> Result<[f64; 2]> {
        self.skip_space();
        let enclosed = self.eat(b'(');
        let x = self.float()?;
        self.expect(b',')?;
        let y = self.float()?;
        if enclosed {
            self.expect(b')')?;
            self.skip_space();
        }
        Ok([x, y])
    }

    /// path reads a count of points, enclosed in brackets for an open path when `open` allows it or in parentheses,
    /// as path_decode does, returning the points and whether the path is open.
    fn path(&mut self, open: bool, count: usize) -> Result<(Vec<[f64; 2]>, bool)> {
        self.skip_space();
        let mut depth = 0;
        let is_open = self.peek() == b'[';
        if is_open {
            if !open {
                return Err(self.invalid());
            }
            depth += 1;
            self.position += 1;
        } else if self.peek() == b'(' {
            let after = self.position + 1;
            let next = after + self.text[after..].len() - self.text[after..].trim_start().len();
            let only = self.text.rfind('(') == Some(self.position);
            if self.text.as_bytes().get(next) == Some(&b'(') || only {
                depth += 1;
                self.position = next;
            }
        }
        let mut points = Vec::with_capacity(count);
        for _ in 0..count {
            points.push(self.pair()?);
            self.eat(b',');
        }
        while depth > 0 {
            if self.peek() == b')' || (self.peek() == b']' && is_open && depth == 1) {
                depth -= 1;
                self.position += 1;
                self.skip_space();
            } else {
                return Err(self.invalid());
            }
        }
        Ok((points, is_open))
    }
}

/// float_prefix returns the length of the float that starts the text, as strtod reads it, or 0 without one.
fn float_prefix(text: &str) -> usize {
    let bytes = text.as_bytes();
    let mut i = usize::from(matches!(bytes.first(), Some(b'+' | b'-')));
    let lower = text[i..].to_ascii_lowercase();
    for word in ["infinity", "inf", "nan"] {
        if lower.starts_with(word) {
            return i + word.len();
        }
    }
    let start = i;
    while bytes.get(i).is_some_and(u8::is_ascii_digit) {
        i += 1;
    }
    if bytes.get(i) == Some(&b'.') {
        i += 1;
        while bytes.get(i).is_some_and(u8::is_ascii_digit) {
            i += 1;
        }
    }
    if i == start || (i == start + 1 && bytes[start] == b'.') {
        return 0;
    }
    if matches!(bytes.get(i), Some(b'e' | b'E')) {
        let mut j = i + 1 + usize::from(matches!(bytes.get(i + 1), Some(b'+' | b'-')));
        let digits = j;
        while bytes.get(j).is_some_and(u8::is_ascii_digit) {
            j += 1;
        }
        if j > digits {
            i = j;
        }
    }
    i
}

/// pair_count returns how many points a text holds from its commas, as pair_count does, or None for an odd count.
fn pair_count(text: &str) -> Option<usize> {
    let commas = text.bytes().filter(|&b| b == b',').count();
    (commas % 2 == 1).then_some(commas.div_ceil(2))
}

/// encode writes the floats of a value in network order.
fn encode(values: &[f64]) -> Vec<u8> {
    values.iter().flat_map(|v| v.to_be_bytes()).collect()
}

/// decode reads the floats of a value in network order.
fn decode(bytes: &[u8]) -> Vec<f64> {
    bytes.as_chunks::<8>().0.iter().map(|c| f64::from_be_bytes(*c)).collect()
}

/// float_text writes a float as float8out does.
fn float_text(value: f64) -> String {
    Value::Float8(value).output().unwrap_or_default()
}

/// points_text writes points as path_encode does, inside the delimiters.
fn points_text(points: &[f64], open: &str, close: &str) -> String {
    let pairs: Vec<String> =
        points.as_chunks::<2>().0.iter().map(|[x, y]| format!("({},{})", float_text(*x), float_text(*y))).collect();
    format!("{open}{}{close}", pairs.join(","))
}

/// insufficient returns Postgres' error for binary data shorter than a value needs.
fn insufficient() -> PgError {
    PgError::new(code::PROTOCOL_VIOLATION, "insufficient data left in message")
}

/// fixed reads the binary format of a value of a count of floats.
fn fixed(bytes: &[u8], count: usize) -> Result<Vec<u8>> {
    match bytes.len().cmp(&(count * 8)) {
        Ordering::Less => Err(insufficient()),
        Ordering::Greater => Err(PgError::new(code::INVALID_BINARY_REPRESENTATION, "incorrect binary data format")),
        Ordering::Equal => Ok(bytes.to_vec()),
    }
}

/// point_in reads a point.
fn point_in(text: &str, _: i32) -> Result<Vec<u8>> {
    let mut scanner = Scanner::new(text, "point");
    let point = scanner.pair()?;
    scanner.finish()?;
    Ok(encode(&point))
}

/// point_out writes a point.
fn point_out(bytes: &[u8]) -> String {
    points_text(&decode(bytes), "", "")
}

/// lseg_in reads a line segment.
fn lseg_in(text: &str, _: i32) -> Result<Vec<u8>> {
    let mut scanner = Scanner::new(text, "lseg");
    let (points, _) = scanner.path(true, 2)?;
    scanner.finish()?;
    Ok(encode(&points.concat()))
}

/// lseg_out writes a line segment.
fn lseg_out(bytes: &[u8]) -> String {
    points_text(&decode(bytes), "[", "]")
}

/// normalized_box orders a box's corners as box_construct does: the upper right corner first.
fn normalized_box(values: &[f64]) -> Vec<u8> {
    let (x1, y1, x2, y2) = (values[0], values[1], values[2], values[3]);
    encode(&[x1.max(x2), y1.max(y2), x1.min(x2), y1.min(y2)])
}

/// box_in reads a box from two of its corners.
fn box_in(text: &str, _: i32) -> Result<Vec<u8>> {
    let mut scanner = Scanner::new(text, "box");
    let (points, _) = scanner.path(false, 2)?;
    scanner.finish()?;
    Ok(normalized_box(&points.concat()))
}

/// box_out writes a box.
fn box_out(bytes: &[u8]) -> String {
    points_text(&decode(bytes), "", "")
}

/// box_recv reads the binary format of a box, ordering its corners.
fn box_recv(bytes: &[u8], _: i32) -> Result<Vec<u8>> {
    Ok(normalized_box(&decode(&fixed(bytes, 4)?)))
}

/// path_in reads a path, open when written in brackets.
fn path_in(text: &str, _: i32) -> Result<Vec<u8>> {
    let mut scanner = Scanner::new(text, "path");
    let count = pair_count(text).filter(|&n| n > 0).ok_or_else(|| scanner.invalid())?;
    scanner.skip_space();
    let mut depth = 0;
    if scanner.peek() == b'(' && text.rfind('(') == Some(scanner.position) {
        scanner.position += 1;
        depth += 1;
    }
    let (points, open) = scanner.path(true, count)?;
    if depth > 0 {
        scanner.expect(b')')?;
        scanner.skip_space();
    }
    scanner.finish()?;
    let mut out = vec![u8::from(!open)];
    out.extend((count as i32).to_be_bytes());
    out.extend(encode(&points.concat()));
    Ok(out)
}

/// path_out writes a path, in parentheses when closed and brackets when open.
fn path_out(bytes: &[u8]) -> String {
    let points = decode(&bytes[5..]);
    if bytes[0] != 0 { points_text(&points, "(", ")") } else { points_text(&points, "[", "]") }
}

/// counted reads the point count of a binary path or polygon at an offset, checking that the points follow.
fn counted(bytes: &[u8], offset: usize, type_name: &str) -> Result<usize> {
    let count = bytes.get(offset..offset + 4).ok_or_else(insufficient)?;
    let count = i32::from_be_bytes(count.try_into().unwrap_or_default());
    if count <= 0 || count as usize > (i32::MAX as usize - offset) / 16 {
        return Err(PgError::new(
            code::INVALID_BINARY_REPRESENTATION,
            format!("invalid number of points in external \"{type_name}\" value"),
        ));
    }
    if bytes.len() < offset + 4 + count as usize * 16 {
        return Err(insufficient());
    }
    Ok(count as usize)
}

/// path_recv reads the binary format of a path: whether it is closed, then its points.
fn path_recv(bytes: &[u8], _: i32) -> Result<Vec<u8>> {
    let closed = *bytes.first().ok_or_else(insufficient)?;
    let count = counted(bytes, 1, "path")?;
    let mut out = vec![u8::from(closed != 0)];
    out.extend_from_slice(&bytes[1..5 + count * 16]);
    Ok(out)
}

/// polygon_in reads a polygon.
fn polygon_in(text: &str, _: i32) -> Result<Vec<u8>> {
    let mut scanner = Scanner::new(text, "polygon");
    let count = pair_count(text).filter(|&n| n > 0).ok_or_else(|| scanner.invalid())?;
    let (points, _) = scanner.path(false, count)?;
    scanner.finish()?;
    let mut out = (count as i32).to_be_bytes().to_vec();
    out.extend(encode(&points.concat()));
    Ok(out)
}

/// polygon_out writes a polygon.
fn polygon_out(bytes: &[u8]) -> String {
    points_text(&decode(&bytes[4..]), "(", ")")
}

/// polygon_recv reads the binary format of a polygon: its point count and points.
fn polygon_recv(bytes: &[u8], _: i32) -> Result<Vec<u8>> {
    let count = counted(bytes, 0, "polygon")?;
    Ok(bytes[..4 + count * 16].to_vec())
}

/// fp_eq compares floats as Postgres' FPeq does, within its tolerance.
fn fp_eq(a: f64, b: f64) -> bool {
    a == b || (a - b).abs() <= EPSILON
}

/// both_zero returns Postgres' error, with the code given, for a line whose A and B are both zero.
fn both_zero(code: &'static str) -> PgError {
    PgError::new(code, "invalid line specification: A and B cannot both be zero")
}

/// line_in reads a line as its equation's coefficients `{A,B,C}`, or as two points it passes through.
fn line_in(text: &str, _: i32) -> Result<Vec<u8>> {
    let mut scanner = Scanner::new(text, "line");
    scanner.skip_space();
    if scanner.eat(b'{') {
        let a = scanner.float()?;
        scanner.expect(b',')?;
        let b = scanner.float()?;
        scanner.expect(b',')?;
        let c = scanner.float()?;
        scanner.expect(b'}')?;
        scanner.skip_space();
        scanner.finish()?;
        if a.abs() <= EPSILON && b.abs() <= EPSILON {
            return Err(both_zero(code::INVALID_TEXT_REPRESENTATION));
        }
        return Ok(encode(&[a, b, c]));
    }
    let (points, _) = scanner.path(true, 2)?;
    scanner.finish()?;
    let ([x1, y1], [x2, y2]) = (points[0], points[1]);
    if fp_eq(x1, x2) && fp_eq(y1, y2) {
        return Err(PgError::new(
            code::INVALID_TEXT_REPRESENTATION,
            "invalid line specification: must be two distinct points",
        ));
    }
    let slope = match (fp_eq(x1, x2), fp_eq(y1, y2)) {
        (true, _) => f64::INFINITY,
        (_, true) => 0.0,
        _ => (y1 - y2) / (x1 - x2),
    };
    let line = if slope.is_infinite() {
        [-1.0, 0.0, x1]
    } else if slope == 0.0 {
        [0.0, -1.0, y1]
    } else {
        let c = y1 - slope * x1;
        [slope, -1.0, if c == 0.0 { 0.0 } else { c }]
    };
    Ok(encode(&line))
}

/// line_out writes a line as its equation's coefficients.
fn line_out(bytes: &[u8]) -> String {
    let values: Vec<String> = decode(bytes).into_iter().map(float_text).collect();
    format!("{{{}}}", values.join(","))
}

/// line_recv reads the binary format of a line.
fn line_recv(bytes: &[u8], _: i32) -> Result<Vec<u8>> {
    let bytes = fixed(bytes, 3)?;
    let values = decode(&bytes);
    if values[0].abs() <= EPSILON && values[1].abs() <= EPSILON {
        return Err(both_zero(code::INVALID_BINARY_REPRESENTATION));
    }
    Ok(bytes)
}

/// circle_in reads a circle as its center and radius.
fn circle_in(text: &str, _: i32) -> Result<Vec<u8>> {
    let mut scanner = Scanner::new(text, "circle");
    scanner.skip_space();
    let mut depth = 0;
    if scanner.eat(b'<') {
        depth += 1;
    } else if scanner.peek() == b'(' {
        let after = scanner.position + 1;
        let next = after + text[after..].len() - text[after..].trim_start().len();
        if text.as_bytes().get(next) == Some(&b'(') {
            depth += 1;
            scanner.position = next;
        }
    }
    let [x, y] = scanner.pair()?;
    scanner.eat(b',');
    let radius = scanner.float()?;
    if radius < 0.0 {
        return Err(scanner.invalid());
    }
    while depth > 0 {
        if scanner.peek() == b')' || (scanner.peek() == b'>' && depth == 1) {
            depth -= 1;
            scanner.position += 1;
            scanner.skip_space();
        } else {
            return Err(scanner.invalid());
        }
    }
    scanner.finish()?;
    Ok(encode(&[x, y, radius]))
}

/// circle_out writes a circle as `<(x,y),r>`.
fn circle_out(bytes: &[u8]) -> String {
    let values = decode(bytes);
    format!("<({},{}),{}>", float_text(values[0]), float_text(values[1]), float_text(values[2]))
}

/// circle_recv reads the binary format of a circle.
fn circle_recv(bytes: &[u8], _: i32) -> Result<Vec<u8>> {
    let bytes = fixed(bytes, 3)?;
    if decode(&bytes)[2] < 0.0 {
        return Err(PgError::new(code::INVALID_BINARY_REPRESENTATION, "invalid radius in external \"circle\" value"));
    }
    Ok(bytes)
}
