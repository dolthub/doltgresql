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

//! Splits a regression test script into the units whose output a run compares. Each unit ends with a line on which
//! psql sends statements or runs a meta-command, found with the rules of psql's lexer (fe_utils/psqlscan.l).

/// Unit is the part of a script that ends with a line on which psql sends statements or runs a meta-command.
#[derive(Clone, Debug, PartialEq)]
pub struct Unit {
    /// The unit's lines that psql echoes before running it.
    pub echo: Vec<String>,
}

impl Unit {
    /// query returns the unit's echoed text without its leading comments and blank lines, which names it in reports.
    pub fn query(&self) -> String {
        let start = self.echo.iter().position(|l| !l.trim().is_empty() && !l.trim_start().starts_with("--"));
        self.echo[start.unwrap_or(0)..].join("\n")
    }
}

/// units splits a script into units, as psql run with `-a` reads it. Lines that psql never echoes, as the data of a
/// COPY FROM STDIN, empty lines outside quotes and comments, or lines read while the ECHO variable is not `all`,
/// belong to no unit, and a unit with no echoed lines joins the one before it.
pub fn units(script: &str) -> Vec<Unit> {
    let mut lexer = Lexer::default();
    let mut units = Vec::new();
    let mut echo = Vec::new();
    let (mut echoing, mut copying) = (true, false);
    for line in script.lines() {
        if copying {
            copying = line != "\\.";
            continue;
        }
        if line.is_empty() && lexer.state == State::Normal {
            continue;
        }
        if echoing {
            echo.push(line.to_string());
        }
        let end = lexer.line(line);
        if let Some(meta) = end.meta {
            let words: Vec<&str> = meta.split_whitespace().collect();
            match words.as_slice() {
                ["\\set", "ECHO", value, ..] => echoing = *value == "all",
                ["\\set", "ECHO"] | ["\\unset", "ECHO", ..] => echoing = false,
                _ => {}
            }
        }
        if end.sends {
            if !echo.is_empty() {
                units.push(Unit { echo: std::mem::take(&mut echo) });
            }
            copying = end.copies;
        }
    }
    if let Some(last) = units.last_mut() {
        last.echo.extend(echo);
    }
    units
}

/// State is where the lexer is within a line.
#[derive(Clone, Debug, Default, PartialEq)]
enum State {
    #[default]
    Normal,
    /// A quoted string, in which backslashes escape when it is an escape string.
    Quote {
        escapes: bool,
    },
    DoubleQuote,
    /// A dollar-quoted string, ended by its delimiter.
    Dollar(String),
    /// A block comment, nested to the depth.
    Comment(usize),
}

/// End is what a line does when psql reads it.
#[derive(Default)]
struct End<'l> {
    /// Whether it sends statements or runs a meta-command.
    sends: bool,
    /// Whether what it sends is a COPY FROM STDIN, whose data follows it.
    copies: bool,
    /// The meta-command that it runs.
    meta: Option<&'l str>,
}

/// Lexer holds psql's lexing state between lines.
#[derive(Default)]
struct Lexer {
    state: State,
    paren_depth: usize,
    begin_depth: usize,
    /// The first keywords of the statement, as psqlscan_record_initial_keyword records them.
    idents: Vec<u8>,
}

impl Lexer {
    /// line lexes a line and returns what it does.
    fn line<'l>(&mut self, line: &'l str) -> End<'l> {
        let b = line.as_bytes();
        let mut end = End::default();
        let mut i = 0;
        while i < b.len() {
            match &mut self.state {
                State::Comment(depth) => {
                    if b[i..].starts_with(b"/*") {
                        *depth += 1;
                        i += 2;
                    } else if b[i..].starts_with(b"*/") {
                        *depth -= 1;
                        if *depth == 0 {
                            self.state = State::Normal;
                        }
                        i += 2;
                    } else {
                        i += 1;
                    }
                }
                State::Quote { escapes } => match b[i] {
                    b'\\' if *escapes => i += 2,
                    b'\'' if b.get(i + 1) == Some(&b'\'') => i += 2,
                    b'\'' => {
                        self.state = State::Normal;
                        i += 1;
                    }
                    _ => i += 1,
                },
                State::DoubleQuote => match b[i] {
                    b'"' if b.get(i + 1) == Some(&b'"') => i += 2,
                    b'"' => {
                        self.state = State::Normal;
                        i += 1;
                    }
                    _ => i += 1,
                },
                State::Dollar(delimiter) => match b[i..].starts_with(delimiter.as_bytes()) {
                    true => {
                        i += delimiter.len();
                        self.state = State::Normal;
                    }
                    false => i += 1,
                },
                State::Normal => match b[i] {
                    b'-' if b.get(i + 1) == Some(&b'-') => break,
                    b'/' if b.get(i + 1) == Some(&b'*') => {
                        self.state = State::Comment(1);
                        i += 2;
                    }
                    b'\'' => {
                        self.state = State::Quote { escapes: false };
                        i += 1;
                    }
                    b'"' => {
                        self.state = State::DoubleQuote;
                        i += 1;
                    }
                    b'$' => match dollar_delimiter(&line[i..]) {
                        Some(delimiter) => {
                            i += delimiter.len();
                            self.state = State::Dollar(delimiter.to_string());
                        }
                        None => i += 1,
                    },
                    b'(' => {
                        self.paren_depth += 1;
                        i += 1;
                    }
                    b')' => {
                        self.paren_depth = self.paren_depth.saturating_sub(1);
                        i += 1;
                    }
                    b';' => {
                        if self.paren_depth == 0 && self.begin_depth == 0 {
                            end.sends = true;
                            end.copies = self.copies_from_stdin();
                            self.idents.clear();
                        }
                        i += 1;
                    }
                    b'\\' if matches!(b.get(i + 1), Some(b';' | b':')) => {
                        if b[i + 1] == b';' && self.paren_depth == 0 && self.begin_depth == 0 {
                            self.idents.clear();
                        }
                        i += 2;
                    }
                    b'\\' => {
                        let meta = &line[i..];
                        end.sends = true;
                        end.copies = meta.to_ascii_lowercase().starts_with("\\copy") && is_from_stdin(meta);
                        end.meta = Some(meta);
                        break;
                    }
                    c if c.is_ascii_digit() => {
                        i += b[i..]
                            .iter()
                            .take_while(|c| c.is_ascii_alphanumeric() || **c == b'_' || **c == b'.')
                            .count();
                    }
                    c if is_ident_start(c) => {
                        let len = b[i..]
                            .iter()
                            .take_while(|c| is_ident_start(**c) || c.is_ascii_digit() || **c == b'$')
                            .count();
                        let word = &line[i..i + len];
                        i += len;
                        if word.eq_ignore_ascii_case("e") && b.get(i) == Some(&b'\'') {
                            self.state = State::Quote { escapes: true };
                            i += 1;
                        } else {
                            self.track_identifier(word);
                        }
                    }
                    _ => i += 1,
                },
            }
        }
        end
    }

    /// track_identifier records a statement's first keywords and counts the BEGIN and END pairs of a routine's body,
    /// as psqlscan_track_identifier does.
    fn track_identifier(&mut self, word: &str) {
        if self.paren_depth != 0 {
            return;
        }
        if self.idents.len() < 8 {
            let first = word.as_bytes()[0];
            let lower = word.to_ascii_lowercase();
            self.idents.push(match lower.as_str() {
                "create" | "function" | "procedure" | "or" | "replace" => first.to_ascii_lowercase(),
                "copy" | "from" | "stdin" | "stdout" => first.to_ascii_uppercase(),
                _ => 0,
            });
        }
        let ident = |i: usize| self.idents.get(i).copied().unwrap_or(0);
        let creates_routine = ident(0) == b'c'
            && (matches!(ident(1), b'f' | b'p')
                || (ident(1) == b'o' && ident(2) == b'r' && matches!(ident(3), b'f' | b'p')));
        if !creates_routine {
            return;
        }
        if word.eq_ignore_ascii_case("begin") {
            self.begin_depth += 1;
        } else if word.eq_ignore_ascii_case("case") {
            if self.begin_depth >= 1 {
                self.begin_depth += 1;
            }
        } else if word.eq_ignore_ascii_case("end") {
            self.begin_depth = self.begin_depth.saturating_sub(1);
        }
    }

    /// copies_from_stdin reports whether the statement is a COPY FROM STDIN, as psqlscan_is_copy_from_stdin does.
    fn copies_from_stdin(&self) -> bool {
        let ident = |i: usize| self.idents.get(i).copied().unwrap_or(0);
        if ident(0) != b'C' {
            return false;
        }
        (1..7).find(|&i| ident(i) == b'F').is_some_and(|i| ident(i + 1) == b'S')
    }
}

/// is_ident_start reports whether a byte starts an identifier.
fn is_ident_start(c: u8) -> bool {
    c.is_ascii_alphabetic() || c == b'_' || c >= 0x80
}

/// is_from_stdin reports whether a `\copy` meta-command reads from psql's input.
fn is_from_stdin(meta: &str) -> bool {
    let words: Vec<String> = meta.split_whitespace().map(str::to_ascii_lowercase).collect();
    words.windows(2).any(|w| w[0] == "from" && w[1] == "stdin")
}

/// dollar_delimiter returns the delimiter of a dollar-quoted string that the text starts with.
fn dollar_delimiter(text: &str) -> Option<&str> {
    let b = text.as_bytes();
    let tag =
        b[1..].iter().enumerate().take_while(|(j, c)| is_ident_start(**c) || (*j > 0 && c.is_ascii_digit())).count();
    (b.get(1 + tag) == Some(&b'$')).then(|| &text[..tag + 2])
}
