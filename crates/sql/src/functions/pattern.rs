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

//! Pattern matching: LIKE, SIMILAR TO, and Postgres' POSIX regular expressions.

use std::collections::HashMap;
use std::rc::Rc;

use fancy_regex::{Regex, RegexBuilder};

use super::{ANY, Function, text};
use crate::array::Array;
use crate::error::{PgError, Result, code};
use crate::oid::{BOOL, INT4, TEXT, TEXT_ARRAY};
use crate::query::Ctx;
use crate::types::Value;

/// f declares a strict pattern function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// FUNCTIONS are the pattern functions, including the ones behind the LIKE and regular expression operators.
pub const FUNCTIONS: &[Function] = &[
    f("textlike", &[TEXT, TEXT], BOOL, like),
    f("textnlike", &[TEXT, TEXT], BOOL, not_like),
    f("texticlike", &[TEXT, TEXT], BOOL, ilike),
    f("texticnlike", &[TEXT, TEXT], BOOL, not_ilike),
    f("textregexeq", &[TEXT, TEXT], BOOL, regex_match),
    f("textregexne", &[TEXT, TEXT], BOOL, regex_no_match),
    f("texticregexeq", &[TEXT, TEXT], BOOL, regex_imatch),
    f("texticregexne", &[TEXT, TEXT], BOOL, regex_no_imatch),
    f("like_escape", &[TEXT, TEXT], TEXT, like_escape),
    f("similar_to_escape", &[TEXT], TEXT, similar_to_escape),
    f("similar_to_escape", &[TEXT, TEXT], TEXT, similar_to_escape),
    f("regexp_replace", &[TEXT, TEXT, TEXT], TEXT, regexp_replace),
    f("regexp_replace", &[TEXT, TEXT, TEXT, TEXT], TEXT, regexp_replace),
    f("regexp_replace", &[TEXT, TEXT, TEXT, INT4], TEXT, regexp_replace_from),
    f("regexp_replace", &[TEXT, TEXT, TEXT, INT4, INT4], TEXT, regexp_replace_from),
    f("regexp_replace", &[TEXT, TEXT, TEXT, INT4, INT4, TEXT], TEXT, regexp_replace_from),
    f("regexp_match", &[TEXT, TEXT], TEXT_ARRAY, regexp_match),
    f("regexp_match", &[TEXT, TEXT, TEXT], TEXT_ARRAY, regexp_match),
    f("regexp_matches", &[TEXT, TEXT], TEXT_ARRAY, regexp_matches),
    f("regexp_matches", &[TEXT, TEXT, TEXT], TEXT_ARRAY, regexp_matches),
    f("regexp_like", &[TEXT, TEXT], BOOL, regexp_like),
    f("regexp_like", &[TEXT, TEXT, TEXT], BOOL, regexp_like),
    f("regexp_count", &[TEXT, TEXT], INT4, regexp_count),
    f("regexp_count", &[TEXT, TEXT, INT4], INT4, regexp_count),
    f("regexp_count", &[TEXT, TEXT, INT4, TEXT], INT4, regexp_count),
    f("regexp_instr", &[TEXT, TEXT], INT4, regexp_instr),
    f("regexp_instr", &[TEXT, TEXT, INT4], INT4, regexp_instr),
    f("regexp_instr", &[TEXT, TEXT, INT4, INT4], INT4, regexp_instr),
    f("regexp_instr", &[TEXT, TEXT, INT4, INT4, INT4], INT4, regexp_instr),
    f("regexp_instr", &[TEXT, TEXT, INT4, INT4, INT4, TEXT], INT4, regexp_instr),
    f("regexp_instr", &[TEXT, TEXT, INT4, INT4, INT4, TEXT, INT4], INT4, regexp_instr),
    f("regexp_substr", &[TEXT, TEXT], TEXT, regexp_substr),
    f("regexp_substr", &[TEXT, TEXT, INT4], TEXT, regexp_substr),
    f("regexp_substr", &[TEXT, TEXT, INT4, INT4], TEXT, regexp_substr),
    f("regexp_substr", &[TEXT, TEXT, INT4, INT4, TEXT], TEXT, regexp_substr),
    f("regexp_substr", &[TEXT, TEXT, INT4, INT4, TEXT, INT4], TEXT, regexp_substr),
    f("regexp_split_to_array", &[TEXT, TEXT], TEXT_ARRAY, regexp_split_to_array),
    f("regexp_split_to_array", &[TEXT, TEXT, TEXT], TEXT_ARRAY, regexp_split_to_array),
    f("regexp_split_to_table", &[TEXT, TEXT], TEXT, regexp_split_to_table),
    f("regexp_split_to_table", &[TEXT, TEXT, TEXT], TEXT, regexp_split_to_table),
    f("substring", &[TEXT, TEXT], TEXT, substring_regex),
    f("substring", &[TEXT, TEXT, TEXT], TEXT, substring_similar),
    f("translate", &[TEXT, TEXT, TEXT], TEXT, translate),
    Function { name: "format", args: &[TEXT, ANY], ret: TEXT, strict: false, variadic: true, implementation: format },
    Function { name: "format", args: &[TEXT], ret: TEXT, strict: false, variadic: false, implementation: format },
];

/// like_matches reports whether text matches a LIKE pattern whose escape character is a backslash, as UTF-8 where `_`
/// matches one character, or byte by byte.
pub(crate) fn like_matches(text: &[u8], pattern: &[u8], utf8: bool) -> Result<bool> {
    // The length of the character that starts with a byte, which keeps every position at a character's start.
    let width = |b: u8| match b {
        _ if !utf8 => 1,
        0xF0.. => 4,
        0xE0.. => 3,
        0xC0.. => 2,
        _ => 1,
    };
    let (mut t, mut p) = (0, 0);
    let (mut star_p, mut star_t) = (None, 0);
    while t < text.len() {
        if p < pattern.len() {
            match pattern[p] {
                b'%' => {
                    star_p = Some(p);
                    star_t = t;
                    p += 1;
                    continue;
                }
                b'_' => {
                    t += width(text[t]);
                    p += 1;
                    continue;
                }
                b'\\' => {
                    let Some(&c) = pattern.get(p + 1) else {
                        return Err(PgError::new(
                            code::INVALID_ESCAPE_SEQUENCE,
                            "LIKE pattern must not end with escape character",
                        ));
                    };
                    let n = width(c);
                    if text.get(t..t + n) == pattern.get(p + 1..p + 1 + n) {
                        t += n;
                        p += 1 + n;
                        continue;
                    }
                }
                c if c == text[t] => {
                    t += 1;
                    p += 1;
                    continue;
                }
                _ => {}
            }
        }
        match star_p {
            Some(s) => {
                star_t += width(text[star_t]);
                t = star_t;
                p = s + 1;
            }
            None => return Ok(false),
        }
    }
    while p < pattern.len() && pattern[p] == b'%' {
        p += 1;
    }
    if p + 1 == pattern.len() && pattern[p] == b'\\' {
        return Err(PgError::new(code::INVALID_ESCAPE_SEQUENCE, "LIKE pattern must not end with escape character"));
    }
    Ok(p == pattern.len())
}

/// plain_like returns whether a function is NOT LIKE when it is LIKE or NOT LIKE of text, which an expression can
/// match against a column's value without copying it.
pub(crate) fn plain_like(index: usize) -> Option<bool> {
    match super::function(index).name {
        "textlike" => Some(false),
        "textnlike" => Some(true),
        _ => None,
    }
}

/// like_value matches LIKE or ILIKE.
fn like_value(args: &[Value], fold: bool) -> Result<bool> {
    match fold {
        true => like_matches(text(&args[0]).to_lowercase().as_bytes(), text(&args[1]).to_lowercase().as_bytes(), true),
        false => like_matches(text(&args[0]).as_bytes(), text(&args[1]).as_bytes(), true),
    }
}

/// like implements LIKE.
fn like(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(like_value(args, false)?))
}

/// not_like implements NOT LIKE.
fn not_like(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(!like_value(args, false)?))
}

/// ilike implements ILIKE.
fn ilike(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(like_value(args, true)?))
}

/// not_ilike implements NOT ILIKE.
fn not_ilike(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(!like_value(args, true)?))
}

/// like_escape rewrites a LIKE pattern with an escape character into one escaped with backslashes.
fn like_escape(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (pattern, escape) = (text(&args[0]), text(&args[1]));
    let mut escape_chars = escape.chars();
    let escape = match (escape_chars.next(), escape_chars.next()) {
        (None, _) => None,
        (Some(c), None) => Some(c),
        _ => {
            return Err(PgError {
                hint: Some("Escape string must be empty or one character.".into()),
                ..PgError::new(code::INVALID_ESCAPE_SEQUENCE, "invalid escape string")
            });
        }
    };
    let mut out = String::new();
    let mut chars = pattern.chars();
    while let Some(c) = chars.next() {
        match escape {
            Some(e) if c == e => match chars.next() {
                Some(next) => {
                    out.push('\\');
                    out.push(next);
                }
                None => {
                    return Err(PgError::new(
                        code::INVALID_ESCAPE_SEQUENCE,
                        "LIKE pattern must not end with escape character",
                    ));
                }
            },
            _ if c == '\\' => out.push_str("\\\\"),
            _ => out.push(c),
        }
    }
    Ok(Value::Text(out))
}

/// similar_to_escape converts a SIMILAR TO pattern into a POSIX regular expression, as Postgres' similar_escape does.
fn similar_to_escape(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let pattern = text(&args[0]);
    let escape = match args.get(1).map(text) {
        None => Some('\\'),
        Some("") => None,
        Some(e) if e.chars().count() == 1 => e.chars().next(),
        Some(_) => {
            return Err(PgError {
                hint: Some("Escape string must be empty or one character.".into()),
                ..PgError::new(code::INVALID_ESCAPE_SEQUENCE, "invalid escape string")
            });
        }
    };
    let mut out = String::from("^(?:");
    let mut chars = pattern.chars().peekable();
    let mut in_bracket = false;
    while let Some(c) = chars.next() {
        if Some(c) == escape {
            match chars.next() {
                Some(next) => {
                    out.push('\\');
                    out.push(next);
                }
                None => {
                    return Err(PgError::new(code::INVALID_ESCAPE_SEQUENCE, "invalid escape string"));
                }
            }
            continue;
        }
        if in_bracket {
            if c == ']' {
                in_bracket = false;
            }
            out.push(c);
            continue;
        }
        match c {
            '[' => {
                in_bracket = true;
                out.push(c);
            }
            '%' => out.push_str(".*"),
            '_' => out.push('.'),
            '(' => out.push_str("(?:"),
            '\\' | '.' | '^' | '$' => {
                out.push('\\');
                out.push(c);
            }
            _ => out.push(c),
        }
    }
    out.push_str(")$");
    Ok(Value::Text(out))
}

/// REGEX_CACHE_SIZE is how many compiled regular expressions a thread keeps, as Postgres' RE cache keeps 32.
const REGEX_CACHE_SIZE: usize = 32;

thread_local! {
    /// REGEXES are the regular expressions this thread compiled lately, by pattern and flags.
    static REGEXES: std::cell::RefCell<HashMap<(String, String), Rc<Regex>>> = std::cell::RefCell::new(HashMap::new());
}

/// compile returns a Postgres regular expression with flags compiled, reusing one compiled lately.
fn compile(pattern: &str, flags: &str) -> Result<Rc<Regex>> {
    let key = (pattern.to_string(), flags.to_string());
    if let Some(regex) = REGEXES.with(|r| r.borrow().get(&key).cloned()) {
        return Ok(regex);
    }
    let regex = Rc::new(build(pattern, flags)?);
    REGEXES.with(|r| {
        let mut regexes = r.borrow_mut();
        if regexes.len() >= REGEX_CACHE_SIZE {
            regexes.clear();
        }
        regexes.insert(key, regex.clone());
    });
    Ok(regex)
}

/// build compiles a Postgres regular expression with flags, as an advanced regular expression where `.` matches
/// newlines.
fn build(pattern: &str, flags: &str) -> Result<Regex> {
    let mut builder = RegexBuilder::new(&translate_regex(pattern));
    builder.dot_matches_new_line(true);
    for flag in flags.chars() {
        match flag {
            'i' => {
                builder.case_insensitive(true);
            }
            'c' => {
                builder.case_insensitive(false);
            }
            'n' | 'm' => {
                builder.dot_matches_new_line(false).multi_line(true);
            }
            's' => {
                builder.dot_matches_new_line(true).multi_line(false);
            }
            'x' => {
                builder.ignore_whitespace(true);
            }
            'g' => {}
            other => {
                return Err(PgError::new(
                    code::INVALID_PARAMETER_VALUE,
                    format!("invalid regular expression option: \"{other}\""),
                ));
            }
        }
    }
    builder.build().map_err(|e| {
        PgError::new(code::INVALID_REGULAR_EXPRESSION, format!("invalid regular expression: {}", regex_error(&e)))
    })
}

/// matching converts the error of running a regular expression, such as running out of backtracking, to Postgres'.
fn matching<T>(result: std::result::Result<T, fancy_regex::Error>) -> Result<T> {
    result.map_err(|e| {
        PgError::new(code::INVALID_REGULAR_EXPRESSION, format!("invalid regular expression: {}", regex_error(&e)))
    })
}

/// regex_error summarizes a regular expression error as briefly as Postgres' messages do.
pub(crate) fn regex_error(err: &dyn std::fmt::Display) -> String {
    let text = err.to_string();
    let lower = text.to_lowercase();
    if lower.contains("unclosed group")
        || lower.contains("unopened group")
        || lower.contains("unclosed open paren")
        || lower.contains("parenthesis")
    {
        "parentheses () not balanced".into()
    } else if lower.contains("unclosed character class") || lower.contains("invalid character class") {
        "brackets [] not balanced".into()
    } else if lower.contains("repetition") || lower.contains("target of repeat") {
        "quantifier operand invalid".into()
    } else {
        text.lines().last().unwrap_or_default().trim().to_string()
    }
}

/// translate_regex rewrites Postgres' regular expression escapes that the regex crate spells differently.
pub(crate) fn translate_regex(pattern: &str) -> String {
    let mut out = String::with_capacity(pattern.len());
    let mut chars = pattern.chars();
    while let Some(c) = chars.next() {
        if c == '\\' {
            match chars.next() {
                Some('m') => out.push_str(r"\b{start}"),
                Some('M') => out.push_str(r"\b{end}"),
                Some('y') => out.push_str(r"\b"),
                Some('Y') => out.push_str(r"\B"),
                Some(next) => {
                    out.push('\\');
                    out.push(next);
                }
                None => out.push('\\'),
            }
        } else {
            out.push(c);
        }
    }
    out
}

/// regex_value matches text against a regular expression, optionally ignoring case.
fn regex_value(args: &[Value], fold: bool) -> Result<bool> {
    matching(compile(text(&args[1]), if fold { "i" } else { "" })?.is_match(text(&args[0])))
}

/// regex_match implements ~.
fn regex_match(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(regex_value(args, false)?))
}

/// regex_no_match implements !~.
fn regex_no_match(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(!regex_value(args, false)?))
}

/// regex_imatch implements ~*.
fn regex_imatch(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(regex_value(args, true)?))
}

/// regex_no_imatch implements !~*.
fn regex_no_imatch(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(!regex_value(args, true)?))
}

/// flags returns an optional flags argument.
fn flags(args: &[Value], index: usize) -> &str {
    args.get(index).map_or("", text)
}

/// replacement rewrites a Postgres replacement string, with \1 through \9 and \&, into the regex crate's form.
fn replacement(text: &str) -> String {
    let mut out = String::new();
    let mut chars = text.chars();
    while let Some(c) = chars.next() {
        match c {
            '\\' => match chars.next() {
                Some(d @ '1'..='9') => out.push_str(&format!("${{{d}}}")),
                Some('&') => out.push_str("${0}"),
                Some('\\') => out.push('\\'),
                Some(other) => {
                    out.push('\\');
                    out.push(other);
                }
                None => out.push('\\'),
            },
            '$' => out.push_str("$$"),
            _ => out.push(c),
        }
    }
    out
}

/// regexp_replace replaces the first match, or every match with the g flag.
fn regexp_replace(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let flags = flags(args, 3);
    let regex = compile(text(&args[1]), flags)?;
    let replace = replacement(text(&args[2]));
    let source = text(&args[0]);
    let limit = if flags.contains('g') { 0 } else { 1 };
    Ok(Value::Text(matching(regex.try_replacen(source, limit, replace.as_str()))?.into_owned()))
}

/// captures returns a match's capture groups as a text array, or the whole match without groups.
fn captures(regex: &Regex, caps: &fancy_regex::Captures<'_, str>) -> Value {
    let values = if regex.captures_len() > 1 {
        (1..regex.captures_len()).map(|i| caps.get(i).map_or(Value::Null, |m| Value::Text(m.as_str().into()))).collect()
    } else {
        vec![Value::Text(caps[0].to_string())]
    };
    Value::Array(Box::new(Array::one_dimensional(TEXT, values)))
}

/// regexp_match returns the capture groups of the first match.
fn regexp_match(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let flags = flags(args, 2);
    if flags.contains('g') {
        return Err(PgError {
            hint: Some("Use the regexp_matches function instead.".into()),
            ..PgError::new(code::INVALID_PARAMETER_VALUE, "regexp_match() does not support the \"global\" option")
        });
    }
    let regex = compile(text(&args[1]), flags)?;
    Ok(matching(regex.captures(text(&args[0])))?.map_or(Value::Null, |c| captures(&regex, &c)))
}

/// regexp_matches returns the capture groups of the first match, or of every match with the g flag, as rows.
fn regexp_matches(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let flags = flags(args, 2);
    let regex = compile(text(&args[1]), flags)?;
    let source = text(&args[0]);
    let rows: Vec<Value> = if flags.contains('g') {
        regex.captures_iter(source).map(|c| matching(c).map(|c| captures(&regex, &c))).collect::<Result<_>>()?
    } else {
        matching(regex.captures(source))?.map(|c| captures(&regex, &c)).into_iter().collect()
    };
    Ok(Value::Set(rows))
}

/// regexp_like reports whether text matches.
fn regexp_like(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(matching(compile(text(&args[1]), flags(args, 2))?.is_match(text(&args[0])))?))
}

/// regexp_count counts the matches at or after a starting character.
fn regexp_count(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let start = parameter(args, 2, "start", 1)?;
    let regex = compile(text(&args[1]), single_flags(args, 3, "regexp_count")?)?;
    Ok(Value::Int4(matches_from(&regex, text(&args[0]), start)?.len() as i32))
}

/// regexp_instr returns the character position of the start, or end, of a match or of one of its groups, as Postgres'
/// regexp_instr does, or 0 without one.
fn regexp_instr(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (start, n) = (parameter(args, 2, "start", 1)?, parameter(args, 3, "n", 1)?);
    let end = match args.get(4) {
        Some(Value::Int4(e)) if !matches!(e, 0 | 1) => {
            return Err(PgError::new(
                code::INVALID_PARAMETER_VALUE,
                format!("invalid value for parameter \"endoption\": {e}"),
            ));
        }
        Some(Value::Int4(e)) => *e == 1,
        _ => false,
    };
    let group = parameter(args, 6, "subexpr", 0)?;
    let regex = compile(text(&args[1]), single_flags(args, 5, "regexp_instr")?)?;
    let source = text(&args[0]);
    let Some(caps) = matches_from(&regex, source, start)?.into_iter().nth(n as usize - 1) else {
        return Ok(Value::Int4(0));
    };
    Ok(Value::Int4(match caps.get(group as usize) {
        Some(m) => source[..if end { m.end() } else { m.start() }].chars().count() as i32 + 1,
        None => 0,
    }))
}

/// regexp_substr returns the text of a match or of one of its groups, or NULL without one.
fn regexp_substr(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (start, n) = (parameter(args, 2, "start", 1)?, parameter(args, 3, "n", 1)?);
    let group = parameter(args, 5, "subexpr", 0)?;
    let regex = compile(text(&args[1]), single_flags(args, 4, "regexp_substr")?)?;
    let found = matches_from(&regex, text(&args[0]), start)?.into_iter().nth(n as usize - 1);
    Ok(found
        .and_then(|caps| caps.get(group as usize).map(|m| Value::Text(m.as_str().to_string())))
        .unwrap_or(Value::Null))
}

/// regexp_replace_from replaces the matches at or after a starting character: every one when N is 0 or the g flag is
/// given, and otherwise the Nth, which is the first without N.
fn regexp_replace_from(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let start = parameter(args, 3, "start", 1)?;
    let n = match args.get(4) {
        Some(Value::Int4(n)) if *n < 0 => {
            return Err(PgError::new(code::INVALID_PARAMETER_VALUE, format!("invalid value for parameter \"n\": {n}")));
        }
        Some(Value::Int4(n)) => *n,
        _ => 1,
    };
    let flags = flags(args, 5);
    let regex = compile(text(&args[1]), flags)?;
    let replace = replacement(text(&args[2]));
    let source = text(&args[0]);
    let (mut out, mut last) = (String::new(), 0);
    for (i, caps) in matches_from(&regex, source, start)?.into_iter().enumerate() {
        if n != 0 && !flags.contains('g') && i as i32 + 1 != n {
            continue;
        }
        let whole = caps.get(0).expect("a match has its whole text");
        out.push_str(&source[last..whole.start()]);
        caps.expand(&replace, &mut out);
        last = whole.end();
    }
    out.push_str(&source[last..]);
    Ok(Value::Text(out))
}

/// parameter reads an integer argument that must be positive, or for subexpr not negative, failing as Postgres does,
/// with a default for an argument the call leaves out.
fn parameter(args: &[Value], index: usize, name: &str, default: i32) -> Result<i32> {
    let minimum = if name == "subexpr" { 0 } else { 1 };
    match args.get(index) {
        Some(Value::Int4(n)) if *n < minimum => {
            Err(PgError::new(code::INVALID_PARAMETER_VALUE, format!("invalid value for parameter \"{name}\": {n}")))
        }
        Some(Value::Int4(n)) => Ok(*n),
        _ => Ok(default),
    }
}

/// single_flags returns an optional flags argument, failing as Postgres does for the g flag, which the function
/// cannot take.
fn single_flags<'a>(args: &'a [Value], index: usize, function: &str) -> Result<&'a str> {
    let flags = flags(args, index);
    if flags.contains('g') {
        return Err(PgError::new(
            code::INVALID_PARAMETER_VALUE,
            format!("{function}() does not support the \"global\" option"),
        ));
    }
    Ok(flags)
}

/// matches_from returns the matches of a regular expression in text that start at or after a 1-based character,
/// where the text before it still counts for anchors.
fn matches_from<'t>(regex: &Regex, source: &'t str, start: i32) -> Result<Vec<fancy_regex::Captures<'t, str>>> {
    let mut position = source.char_indices().nth(start as usize - 1).map_or(source.len(), |(i, _)| i);
    let mut found = Vec::new();
    while position <= source.len() {
        let Some(caps) = matching(regex.captures_from_pos(source, position))? else { break };
        let whole = caps.get(0).expect("a match has its whole text");
        position = match whole.end() == whole.start() {
            true => whole.end() + source[whole.end()..].chars().next().map_or(1, char::len_utf8),
            false => whole.end(),
        };
        found.push(caps);
    }
    Ok(found)
}

/// split splits text at the matches, as Postgres does, ignoring empty matches at the ends of the text or right after
/// another match.
fn split(args: &[Value]) -> Result<Vec<Value>> {
    let regex = compile(text(&args[1]), flags(args, 2))?;
    let source = text(&args[0]);
    let mut parts = Vec::new();
    let mut start = 0;
    for m in regex.find_iter(source) {
        let m = matching(m)?;
        if m.start() == m.end() && (m.start() == start || m.start() == source.len()) {
            continue;
        }
        parts.push(Value::Text(source[start..m.start()].to_string()));
        start = m.end();
    }
    parts.push(Value::Text(source[start..].to_string()));
    Ok(parts)
}

/// regexp_split_to_array splits text at the matches into an array.
fn regexp_split_to_array(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Array(Box::new(Array::one_dimensional(TEXT, split(args)?))))
}

/// regexp_split_to_table splits text at the matches into rows.
fn regexp_split_to_table(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Set(split(args)?))
}

/// substring_regex returns the first parenthesized group of the first match, or the whole match without groups.
fn substring_regex(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let regex = compile(text(&args[1]), "")?;
    Ok(match matching(regex.captures(text(&args[0])))? {
        Some(caps) if regex.captures_len() > 1 => caps.get(1).map_or(Value::Null, |m| Value::Text(m.as_str().into())),
        Some(caps) => Value::Text(caps[0].to_string()),
        None => Value::Null,
    })
}

/// substring_similar returns the part of text that a SIMILAR TO pattern's escaped double quotes surround.
fn substring_similar(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let escape = text(&args[2]);
    let quote = format!("{escape}\"");
    let pattern = text(&args[1]);
    let parts: Vec<&str> = pattern.split(quote.as_str()).collect();
    let regex = match parts.as_slice() {
        [whole] => similar_part(ctx, whole, escape, true)?,
        [before, middle, after] => format!(
            "^(?:{})({})(?:{})$",
            similar_part(ctx, before, escape, false)?,
            similar_part(ctx, middle, escape, false)?,
            similar_part(ctx, after, escape, false)?
        ),
        _ => {
            return Err(PgError::new(
                code::INVALID_ESCAPE_SEQUENCE,
                "SQL regular expression may not contain more than two escape-double-quote separators",
            ));
        }
    };
    let compiled = compile(&regex, "")?;
    Ok(match matching(compiled.captures(text(&args[0])))? {
        Some(caps) if parts.len() == 3 => caps.get(1).map_or(Value::Null, |m| Value::Text(m.as_str().into())),
        Some(caps) => Value::Text(caps[0].to_string()),
        None => Value::Null,
    })
}

/// similar_part converts part of a SIMILAR TO pattern to a regular expression, anchored when it is the whole pattern.
fn similar_part(ctx: &mut Ctx<'_>, part: &str, escape: &str, anchored: bool) -> Result<String> {
    let Value::Text(converted) = similar_to_escape(ctx, &[Value::Text(part.into()), Value::Text(escape.into())])?
    else {
        return Ok(String::new());
    };
    if anchored {
        return Ok(converted);
    }
    Ok(converted.trim_start_matches("^(?:").trim_end_matches(")$").to_string())
}

/// translate replaces each character of text found in one string with the character at the same place in another,
/// dropping it when the other string is shorter.
fn translate(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let from: Vec<char> = text(&args[1]).chars().collect();
    let to: Vec<char> = text(&args[2]).chars().collect();
    Ok(Value::Text(
        text(&args[0])
            .chars()
            .filter_map(|c| match from.iter().position(|&f| f == c) {
                Some(i) => to.get(i).copied(),
                None => Some(c),
            })
            .collect(),
    ))
}

/// format formats text with %s, %I, %L, and %%, and their positions and widths, as Postgres' format does.
fn format(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let Value::Text(spec) = &args[0] else { return Ok(Value::Null) };
    let values = &args[1..];
    let mut out = String::new();
    let mut chars = spec.chars().peekable();
    let mut next = 0;
    let unterminated = || PgError::new(code::INVALID_PARAMETER_VALUE, "unterminated format() type specifier");
    while let Some(c) = chars.next() {
        if c != '%' {
            out.push(c);
            continue;
        }
        if chars.peek() == Some(&'%') {
            chars.next();
            out.push('%');
            continue;
        }
        let mut digits = String::new();
        while chars.peek().is_some_and(|c| c.is_ascii_digit()) {
            digits.push(chars.next().unwrap_or_default());
        }
        let mut position = None;
        if chars.peek() == Some(&'$') && !digits.is_empty() {
            chars.next();
            position = Some(digits.parse::<usize>().unwrap_or(0));
            digits.clear();
        }
        let mut left = false;
        if chars.peek() == Some(&'-') {
            chars.next();
            left = true;
        }
        let mut width = None;
        if chars.peek() == Some(&'*') {
            chars.next();
            let w = values.get(next).and_then(|v| v.output()).and_then(|v| v.parse::<i64>().ok()).unwrap_or(0);
            next += 1;
            if w < 0 {
                left = true;
            }
            width = Some(w.unsigned_abs() as usize);
        } else {
            while chars.peek().is_some_and(|c| c.is_ascii_digit()) {
                digits.push(chars.next().unwrap_or_default());
            }
            if !digits.is_empty() {
                width = digits.parse().ok();
            }
        }
        let kind = chars.next().ok_or_else(unterminated)?;
        let index = match position {
            Some(0) => {
                return Err(PgError::new(
                    code::INVALID_PARAMETER_VALUE,
                    "format specifies argument 0, but arguments are numbered from 1",
                ));
            }
            Some(p) => {
                next = p;
                p - 1
            }
            None => {
                next += 1;
                next - 1
            }
        };
        let value = values
            .get(index)
            .ok_or_else(|| PgError::new(code::INVALID_PARAMETER_VALUE, "too few arguments for format()"))?;
        let rendered = match kind {
            's' => value.output().unwrap_or_default(),
            'I' => match value.output() {
                Some(v) => crate::engine::quote_identifier(&v),
                None => {
                    return Err(PgError::new(
                        code::NULL_VALUE_NOT_ALLOWED,
                        "null values cannot be formatted as an SQL identifier",
                    ));
                }
            },
            'L' => match value.output() {
                Some(v) if v.contains('\\') => format!("E'{}'", v.replace('\\', "\\\\").replace('\'', "''")),
                Some(v) => format!("'{}'", v.replace('\'', "''")),
                None => "NULL".to_string(),
            },
            other => {
                return Err(PgError::new(
                    code::INVALID_PARAMETER_VALUE,
                    format!("unrecognized format() type specifier \"{other}\""),
                ));
            }
        };
        let pad = width.unwrap_or(0).saturating_sub(rendered.chars().count());
        if left {
            out.push_str(&rendered);
            out.extend(std::iter::repeat_n(' ', pad));
        } else {
            out.extend(std::iter::repeat_n(' ', pad));
            out.push_str(&rendered);
        }
    }
    Ok(Value::Text(out))
}
