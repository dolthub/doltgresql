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

//! The tsvector and tsquery types, read and written as Postgres' tsvector.c and tsquery.c do, and stored as their
//! binary formats.

use crate::error::{PgError, Result, code};
use crate::extensions::BaseType;

/// TSVECTOR is the tsvector type, a sorted list of distinct lexemes with their positions.
pub const TSVECTOR: BaseType = BaseType {
    name: "tsvector",
    input: |text, _| vector_in(text).map(|lexemes| vector_bytes(&lexemes)),
    output: |bytes| vector_out(&vector_lexemes(bytes).unwrap_or_default()),
    receive: |bytes, _| vector_lexemes(bytes).map(|lexemes| vector_bytes(&normalized(lexemes))),
    send: <[u8]>::to_vec,
    typmod_in: |_| Ok(-1),
    typmod: |_, _| Ok(()),
    compare: |l, r| l.len().cmp(&r.len()).then_with(|| l.cmp(r)),
    vector: None,
};

/// TSQUERY is the tsquery type, a tree of lexemes joined by operators.
pub const TSQUERY: BaseType = BaseType {
    name: "tsquery",
    input: |text, _| query_in(text).map(|query| query_bytes(query.as_ref())),
    output: |bytes| query_out(query_tree(bytes).ok().flatten().as_ref()),
    receive: |bytes, _| query_tree(bytes).map(|query| query_bytes(query.as_ref())),
    send: <[u8]>::to_vec,
    typmod_in: |_| Ok(-1),
    typmod: |_, _| Ok(()),
    compare: query_compare,
    vector: None,
};

/// MAX_POSITION is the largest position a lexeme can have, which larger positions become.
pub(crate) const MAX_POSITION: u16 = (1 << 14) - 1;

/// MAX_POSITIONS is how many positions a lexeme keeps.
const MAX_POSITIONS: usize = 256;

/// MAX_LEXEME is the length in bytes of the longest lexeme.
const MAX_LEXEME: usize = 2047;

/// Lexeme is a lexeme with its positions, each a 14-bit position under a 2-bit weight where 3 is A and 0 is D.
pub(crate) type Lexeme = (Vec<u8>, Vec<u16>);

/// syntax_error returns the error for text that is not a valid value of a text search type.
fn syntax_error(kind: &str, text: &str) -> PgError {
    PgError::new(code::SYNTAX_ERROR, format!("syntax error in {kind}: \"{text}\""))
}

/// Reader walks the text of a text search value.
struct Reader<'t> {
    text: &'t str,
    bytes: &'t [u8],
    at: usize,
}

impl Reader<'_> {
    /// skip_space moves past whitespace.
    fn skip_space(&mut self) {
        while self.bytes.get(self.at).is_some_and(u8::is_ascii_whitespace) {
            self.at += 1;
        }
    }

    /// word reads a lexeme, quoted in apostrophes or ended by whitespace or the characters that end an operand,
    /// where a backslash escapes the next character and two apostrophes in a quoted lexeme stand for one.
    fn word(&mut self, kind: &str, ends: &[u8]) -> Result<Vec<u8>> {
        let mut word = Vec::new();
        let quoted = self.bytes.get(self.at) == Some(&b'\'');
        if quoted {
            self.at += 1;
        }
        loop {
            let Some(&byte) = self.bytes.get(self.at) else {
                return match quoted {
                    true => Err(syntax_error(kind, self.text)),
                    false => Ok(word),
                };
            };
            match byte {
                b'\\' => {
                    let Some(&next) = self.bytes.get(self.at + 1) else { return Err(syntax_error(kind, self.text)) };
                    word.push(next);
                    self.at += 2;
                    continue;
                }
                b'\'' if quoted && self.bytes.get(self.at + 1) == Some(&b'\'') => {
                    word.push(b'\'');
                    self.at += 2;
                    continue;
                }
                b'\'' if quoted => {
                    self.at += 1;
                    return Ok(word);
                }
                _ if !quoted && (byte.is_ascii_whitespace() || ends.contains(&byte)) => return Ok(word),
                _ => word.push(byte),
            }
            self.at += 1;
        }
    }
}

/// weight_bits returns the 2-bit weight that a weight letter stands for, where A is 3 and D is 0.
fn weight_bits(letter: u8) -> Option<u16> {
    match letter.to_ascii_uppercase() {
        b'A' => Some(3),
        b'B' => Some(2),
        b'C' => Some(1),
        b'D' => Some(0),
        _ => None,
    }
}

/// vector_in reads a tsvector's text, sorting its lexemes and merging their positions.
fn vector_in(text: &str) -> Result<Vec<Lexeme>> {
    let mut reader = Reader { text, bytes: text.as_bytes(), at: 0 };
    let mut lexemes = Vec::new();
    loop {
        reader.skip_space();
        if reader.at >= reader.bytes.len() {
            break;
        }
        let word = reader.word("tsvector", b":")?;
        if word.len() > MAX_LEXEME {
            return Err(PgError::new(
                code::PROGRAM_LIMIT_EXCEEDED,
                format!("word is too long ({} bytes, max {MAX_LEXEME} bytes)", word.len()),
            ));
        }
        let mut positions = Vec::new();
        if reader.bytes.get(reader.at) == Some(&b':') {
            reader.at += 1;
            loop {
                let start = reader.at;
                while reader.bytes.get(reader.at).is_some_and(u8::is_ascii_digit) {
                    reader.at += 1;
                }
                let position: u32 = text[start..reader.at].parse().map_err(|_| syntax_error("tsvector", text))?;
                if position == 0 {
                    return Err(PgError::new(
                        code::SYNTAX_ERROR,
                        format!("wrong position info in tsvector: \"{text}\""),
                    ));
                }
                let mut entry = position.min(u32::from(MAX_POSITION)) as u16;
                while let Some(weight) =
                    reader.bytes.get(reader.at).and_then(|&b| weight_bits(b).or((b == b'*').then_some(3)))
                {
                    if entry >> 14 != 0 {
                        return Err(syntax_error("tsvector", text));
                    }
                    entry |= weight << 14;
                    reader.at += 1;
                }
                positions.push(entry);
                match reader.bytes.get(reader.at) {
                    Some(b',') => reader.at += 1,
                    Some(b) if !b.is_ascii_whitespace() => return Err(syntax_error("tsvector", text)),
                    _ => break,
                }
            }
        }
        lexemes.push((word, positions));
    }
    Ok(normalized(lexemes))
}

/// normalized sorts lexemes and merges each one's duplicates, sorting its positions and keeping the heaviest weight
/// of each position, as Postgres' uniqueentry and uniquePos do.
pub(crate) fn normalized(mut lexemes: Vec<Lexeme>) -> Vec<Lexeme> {
    lexemes.sort_by(|a, b| a.0.cmp(&b.0));
    let mut merged: Vec<Lexeme> = Vec::with_capacity(lexemes.len());
    for (word, positions) in lexemes {
        match merged.last_mut() {
            Some((last, kept)) if *last == word => kept.extend(positions),
            _ => merged.push((word, positions)),
        }
    }
    for (_, positions) in &mut merged {
        positions.sort_by_key(|p| (p & MAX_POSITION, std::cmp::Reverse(p >> 14)));
        positions.dedup_by_key(|p| *p & MAX_POSITION);
        positions.truncate(MAX_POSITIONS);
    }
    merged
}

/// vector_bytes writes lexemes in tsvector's binary format: their count, then each lexeme ending in a zero byte with
/// its positions' count and positions.
pub(crate) fn vector_bytes(lexemes: &[Lexeme]) -> Vec<u8> {
    let mut out = (lexemes.len() as u32).to_be_bytes().to_vec();
    for (word, positions) in lexemes {
        out.extend_from_slice(word);
        out.push(0);
        out.extend_from_slice(&(positions.len() as u16).to_be_bytes());
        for position in positions {
            out.extend_from_slice(&position.to_be_bytes());
        }
    }
    out
}

/// vector_lexemes reads tsvector's binary format.
pub(crate) fn vector_lexemes(bytes: &[u8]) -> Result<Vec<Lexeme>> {
    let invalid = || PgError::new(code::INVALID_BINARY_REPRESENTATION, "invalid tsvector binary data");
    let count = u32::from_be_bytes(bytes.get(..4).ok_or_else(invalid)?.try_into().map_err(|_| invalid())?);
    let mut at = 4;
    let mut lexemes = Vec::new();
    for _ in 0..count {
        let end = at + bytes.get(at..).ok_or_else(invalid)?.iter().position(|&b| b == 0).ok_or_else(invalid)?;
        let word = bytes[at..end].to_vec();
        at = end + 1;
        let read = |at: usize| bytes.get(at..at + 2).map(|b| u16::from_be_bytes([b[0], b[1]])).ok_or_else(invalid);
        let count = read(at)?;
        at += 2;
        let mut positions = Vec::with_capacity(count as usize);
        for _ in 0..count {
            positions.push(read(at)?);
            at += 2;
        }
        lexemes.push((word, positions));
    }
    Ok(lexemes)
}

/// quoted writes a lexeme in apostrophes, doubling its apostrophes and backslashes.
fn quoted(word: &[u8]) -> String {
    let text = String::from_utf8_lossy(word);
    format!("'{}'", text.replace('\'', "''").replace('\\', "\\\\"))
}

/// vector_out writes a tsvector's text: its lexemes in order, each with its positions and their weights.
fn vector_out(lexemes: &[Lexeme]) -> String {
    let words: Vec<String> = lexemes
        .iter()
        .map(|(word, positions)| {
            let mut text = quoted(word);
            for (i, position) in positions.iter().enumerate() {
                text.push(if i == 0 { ':' } else { ',' });
                text.push_str(&(position & MAX_POSITION).to_string());
                match position >> 14 {
                    3 => text.push('A'),
                    2 => text.push('B'),
                    1 => text.push('C'),
                    _ => {}
                }
            }
            text
        })
        .collect();
    words.join(" ")
}

/// Operator is a tsquery operator, numbered as Postgres stores it.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub(crate) enum Operator {
    Not = 1,
    And = 2,
    Or = 3,
    Phrase = 4,
}

impl Operator {
    /// priority returns how tightly the operator binds, as Postgres' OP_PRIORITY gives it.
    fn priority(self) -> u8 {
        match self {
            Operator::Not => 4,
            Operator::Phrase => 3,
            Operator::And => 2,
            Operator::Or => 1,
        }
    }
}

/// Query is a node of a tsquery: a lexeme with its weights, as bits where A is 8 and D is 1, and whether it matches
/// as a prefix, or an operator with its operands, where the phrase operator has a distance.
#[derive(Clone, Debug, PartialEq)]
pub(crate) enum Query {
    Lexeme { word: Vec<u8>, weights: u8, prefix: bool },
    Not(Box<Query>),
    Binary { operator: Operator, distance: u16, left: Box<Query>, right: Box<Query> },
}

/// Token is a token of a tsquery's text.
enum Token {
    Operand(Query),
    Operator(Operator, u16),
    Open,
    Close,
}

/// query_tokens reads a tsquery's text as tokens, as Postgres' gettoken_query_standard does.
fn query_tokens(text: &str) -> Result<Vec<Token>> {
    let mut reader = Reader { text, bytes: text.as_bytes(), at: 0 };
    let mut tokens = Vec::new();
    let mut operand = true;
    loop {
        reader.skip_space();
        let Some(&byte) = reader.bytes.get(reader.at) else { break };
        match (operand, byte) {
            (true, b'!') => {
                reader.at += 1;
                tokens.push(Token::Operator(Operator::Not, 0));
            }
            (true, b'(') => {
                reader.at += 1;
                tokens.push(Token::Open);
            }
            (true, b':' | b'&' | b'|' | b')') => return Err(syntax_error("tsquery", text)),
            (true, _) => {
                let word = reader.word("tsquery", b":&|!()<")?;
                let (mut weights, mut prefix) = (0u8, false);
                if reader.bytes.get(reader.at) == Some(&b':') {
                    reader.at += 1;
                    while let Some(&b) = reader.bytes.get(reader.at) {
                        match (b, weight_bits(b)) {
                            (b'*', _) => prefix = true,
                            (_, Some(bits)) => weights |= 1 << bits,
                            _ => break,
                        }
                        reader.at += 1;
                    }
                }
                tokens.push(Token::Operand(Query::Lexeme { word, weights, prefix }));
                operand = false;
            }
            (false, b'&') => {
                reader.at += 1;
                tokens.push(Token::Operator(Operator::And, 0));
                operand = true;
            }
            (false, b'|') => {
                reader.at += 1;
                tokens.push(Token::Operator(Operator::Or, 0));
                operand = true;
            }
            (false, b'<') => {
                let rest = &text[reader.at + 1..];
                let close = rest.find('>').ok_or_else(|| syntax_error("tsquery", text))?;
                let distance = match &rest[..close] {
                    "-" => 1,
                    digits => digits.parse::<u16>().map_err(|_| syntax_error("tsquery", text))?,
                };
                if distance > MAX_POSITION {
                    return Err(PgError::new(
                        code::INVALID_PARAMETER_VALUE,
                        format!(
                            "distance in phrase operator must be an integer value between zero and {MAX_POSITION} inclusive"
                        ),
                    ));
                }
                reader.at += close + 2;
                tokens.push(Token::Operator(Operator::Phrase, distance));
                operand = true;
            }
            (false, b')') => {
                reader.at += 1;
                tokens.push(Token::Close);
            }
            (false, _) => return Err(syntax_error("tsquery", text)),
        }
    }
    if operand && !tokens.is_empty() {
        return Err(PgError::new(code::SYNTAX_ERROR, format!("no operand in tsquery: \"{text}\"")));
    }
    Ok(tokens)
}

/// query_in reads a tsquery's text into its tree, or None for a query without lexemes, building it by operator
/// priority as Postgres' makepol and cleanOpStack do.
fn query_in(text: &str) -> Result<Option<Query>> {
    let tokens = query_tokens(text)?;
    let mut operands: Vec<Query> = Vec::new();
    let mut operators: Vec<Option<(Operator, u16)>> = Vec::new();
    let apply = |operands: &mut Vec<Query>, (operator, distance): (Operator, u16)| -> Result<()> {
        let right = operands.pop().ok_or_else(|| syntax_error("tsquery", text))?;
        let node = match operator {
            Operator::Not => Query::Not(Box::new(right)),
            _ => {
                let left = operands.pop().ok_or_else(|| syntax_error("tsquery", text))?;
                Query::Binary { operator, distance, left: Box::new(left), right: Box::new(right) }
            }
        };
        operands.push(node);
        Ok(())
    };
    let clean =
        |operands: &mut Vec<Query>, operators: &mut Vec<Option<(Operator, u16)>>, next: Operator| -> Result<()> {
            while let Some(&Some((top, distance))) = operators.last() {
                let stops = match next {
                    Operator::Not => next.priority() >= top.priority(),
                    _ => next.priority() > top.priority(),
                };
                if stops {
                    break;
                }
                operators.pop();
                apply(operands, (top, distance))?;
            }
            Ok(())
        };
    for token in tokens {
        match token {
            Token::Operand(operand) => operands.push(operand),
            Token::Operator(operator, distance) => {
                clean(&mut operands, &mut operators, operator)?;
                operators.push(Some((operator, distance)));
            }
            Token::Open => operators.push(None),
            Token::Close => {
                clean(&mut operands, &mut operators, Operator::Or)?;
                if operators.pop() != Some(None) {
                    return Err(syntax_error("tsquery", text));
                }
            }
        }
    }
    clean(&mut operands, &mut operators, Operator::Or)?;
    if !operators.is_empty() || operands.len() > 1 {
        return Err(syntax_error("tsquery", text));
    }
    Ok(operands.pop())
}

/// query_bytes writes a tsquery in its binary format: the count of its nodes, then each node with an operator before
/// its right operand and then its left one, as Postgres stores the tree.
pub(crate) fn query_bytes(query: Option<&Query>) -> Vec<u8> {
    let mut out = query.map_or(0, node_count).to_be_bytes().to_vec();
    if let Some(query) = query {
        write_nodes(query, &mut out);
    }
    out
}

/// query_tree reads tsquery's binary format.
pub(crate) fn query_tree(bytes: &[u8]) -> Result<Option<Query>> {
    let invalid = || PgError::new(code::INVALID_BINARY_REPRESENTATION, "invalid tsquery binary data");
    let count = u32::from_be_bytes(bytes.get(..4).ok_or_else(invalid)?.try_into().map_err(|_| invalid())?);
    if count == 0 {
        return Ok(None);
    }
    let mut at = 4;
    read_node(bytes, &mut at, &invalid).map(Some)
}

/// query_out writes a tsquery's text, parenthesizing an operation inside one that binds more tightly and a phrase
/// operation on the right of another, as Postgres' infix does.
pub(crate) fn query_out(query: Option<&Query>) -> String {
    let mut out = String::new();
    if let Some(query) = query {
        infix(query, 0, false, &mut out);
    }
    out
}

/// node_count returns how many nodes a tree has.
pub(crate) fn node_count(query: &Query) -> u32 {
    match query {
        Query::Lexeme { .. } => 1,
        Query::Not(operand) => 1 + node_count(operand),
        Query::Binary { left, right, .. } => 1 + node_count(left) + node_count(right),
    }
}

/// write_nodes writes a tsquery tree's nodes, each operator before its right operand and then its left one.
fn write_nodes(query: &Query, out: &mut Vec<u8>) {
    match query {
        Query::Lexeme { word, weights, prefix } => {
            out.extend_from_slice(&[1, *weights, u8::from(*prefix)]);
            out.extend_from_slice(word);
            out.push(0);
        }
        Query::Not(operand) => {
            out.extend_from_slice(&[2, Operator::Not as u8]);
            write_nodes(operand, out);
        }
        Query::Binary { operator, distance, left, right } => {
            out.extend_from_slice(&[2, *operator as u8]);
            if *operator == Operator::Phrase {
                out.extend_from_slice(&distance.to_be_bytes());
            }
            write_nodes(right, out);
            write_nodes(left, out);
        }
    }
}

/// read_node reads one node of tsquery's binary format and its operands.
fn read_node(bytes: &[u8], at: &mut usize, invalid: &dyn Fn() -> PgError) -> Result<Query> {
    let mut byte = || -> Result<u8> {
        let b = *bytes.get(*at).ok_or_else(invalid)?;
        *at += 1;
        Ok(b)
    };
    match byte()? {
        1 => {
            let (weights, prefix) = (byte()?, byte()? != 0);
            let end = *at + bytes.get(*at..).ok_or_else(invalid)?.iter().position(|&b| b == 0).ok_or_else(invalid)?;
            let word = bytes[*at..end].to_vec();
            *at = end + 1;
            Ok(Query::Lexeme { word, weights, prefix })
        }
        2 => {
            let operator = match byte()? {
                1 => Operator::Not,
                2 => Operator::And,
                3 => Operator::Or,
                4 => Operator::Phrase,
                _ => return Err(invalid()),
            };
            if operator == Operator::Not {
                return Ok(Query::Not(Box::new(read_node(bytes, at, invalid)?)));
            }
            let mut distance = 0;
            if operator == Operator::Phrase {
                distance = u16::from_be_bytes([byte()?, byte()?]);
            }
            let right = read_node(bytes, at, invalid)?;
            let left = read_node(bytes, at, invalid)?;
            Ok(Query::Binary { operator, distance, left: Box::new(left), right: Box::new(right) })
        }
        _ => Err(invalid()),
    }
}

/// infix writes a tsquery node's text under a parent of the priority, which is on the right of a phrase operator
/// when `right_of_phrase` is set.
fn infix(query: &Query, parent: u8, right_of_phrase: bool, out: &mut String) {
    match query {
        Query::Lexeme { word, weights, prefix } => {
            out.push_str(&quoted(word));
            if *weights != 0 || *prefix {
                out.push(':');
                if *prefix {
                    out.push('*');
                }
                for (bit, letter) in [(8, 'A'), (4, 'B'), (2, 'C'), (1, 'D')] {
                    if weights & bit != 0 {
                        out.push(letter);
                    }
                }
            }
        }
        Query::Not(operand) => {
            let priority = Operator::Not.priority();
            let parenthesized = priority < parent;
            if parenthesized {
                out.push_str("( ");
            }
            out.push('!');
            infix(operand, priority, false, out);
            if parenthesized {
                out.push_str(" )");
            }
        }
        Query::Binary { operator, distance, left, right } => {
            let priority = operator.priority();
            let parenthesized = priority < parent || (*operator == Operator::Phrase && right_of_phrase);
            if parenthesized {
                out.push_str("( ");
            }
            infix(left, priority, false, out);
            match (operator, distance) {
                (Operator::Or, _) => out.push_str(" | "),
                (Operator::And, _) => out.push_str(" & "),
                (_, 1) => out.push_str(" <-> "),
                (_, distance) => out.push_str(&format!(" <{distance}> ")),
            }
            infix(right, priority, *operator == Operator::Phrase, out);
            if parenthesized {
                out.push_str(" )");
            }
        }
    }
}

/// query_compare orders tsqueries as Postgres' CompareTSQ does: by node count, then by the size of their lexemes, then
/// node by node.
fn query_compare(l: &[u8], r: &[u8]) -> std::cmp::Ordering {
    let (Ok(l), Ok(r)) = (query_tree(l), query_tree(r)) else { return l.cmp(r) };
    let count = |q: &Option<Query>| q.as_ref().map_or(0, node_count);
    let size = |q: &Option<Query>| q.as_ref().map_or(0, lexeme_bytes);
    count(&l).cmp(&count(&r)).then_with(|| size(&l).cmp(&size(&r))).then_with(|| match (&l, &r) {
        (Some(l), Some(r)) => node_compare(l, r),
        _ => std::cmp::Ordering::Equal,
    })
}

/// lexeme_bytes returns how many bytes a query's lexemes take stored with their terminators.
fn lexeme_bytes(query: &Query) -> usize {
    match query {
        Query::Lexeme { word, .. } => word.len() + 1,
        Query::Not(inner) => lexeme_bytes(inner),
        Query::Binary { left, right, .. } => lexeme_bytes(left) + lexeme_bytes(right),
    }
}

/// node_compare orders two query nodes as Postgres' QTNodeCompare does, putting operators first, comparing lexemes by
/// their CRC before their text, and comparing an operator's right operand before its left.
fn node_compare(l: &Query, r: &Query) -> std::cmp::Ordering {
    use std::cmp::Ordering;
    let operator = |q: &Query| match q {
        Query::Lexeme { .. } => None,
        Query::Not(_) => Some((Operator::Not as u8, 0)),
        Query::Binary { operator, distance, .. } => Some((*operator as u8, *distance)),
    };
    match (l, r) {
        (Query::Lexeme { word: lw, .. }, Query::Lexeme { word: rw, .. }) => {
            legacy_crc32(rw).cmp(&legacy_crc32(lw)).then_with(|| lw.cmp(rw))
        }
        (Query::Lexeme { .. }, _) => Ordering::Greater,
        (_, Query::Lexeme { .. }) => Ordering::Less,
        _ => {
            let ((lo, ld), (ro, rd)) = (operator(l).unwrap_or_default(), operator(r).unwrap_or_default());
            let (lc, rc) = (children(l), children(r));
            ro.cmp(&lo)
                .then_with(|| rc.len().cmp(&lc.len()))
                .then_with(|| {
                    lc.iter().zip(&rc).map(|(a, b)| node_compare(a, b)).find(|o| o.is_ne()).unwrap_or(Ordering::Equal)
                })
                .then_with(|| if lo == Operator::Phrase as u8 { rd.cmp(&ld) } else { Ordering::Equal })
        }
    }
}

/// children returns an operator's operands in the order Postgres stores them, right before left.
fn children(query: &Query) -> Vec<&Query> {
    match query {
        Query::Not(inner) => vec![inner],
        Query::Binary { left, right, .. } => vec![right, left],
        Query::Lexeme { .. } => Vec::new(),
    }
}

/// legacy_crc32 returns the CRC that Postgres keeps for a tsquery lexeme, its LEGACY_CRC32, which feeds bytes from the
/// high end into the reflected CRC-32 table.
fn legacy_crc32(bytes: &[u8]) -> i32 {
    let entry = |index: u32| (0..8).fold(index, |c, _| if c & 1 != 0 { 0xEDB8_8320 ^ (c >> 1) } else { c >> 1 });
    let crc = bytes.iter().fold(u32::MAX, |crc, &b| entry(((crc >> 24) ^ u32::from(b)) & 0xFF) ^ (crc << 8));
    !crc as i32
}
