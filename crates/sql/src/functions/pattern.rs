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

use regex::{Regex, RegexBuilder};

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
    f("regexp_match", &[TEXT, TEXT], TEXT_ARRAY, regexp_match),
    f("regexp_match", &[TEXT, TEXT, TEXT], TEXT_ARRAY, regexp_match),
    f("regexp_matches", &[TEXT, TEXT], TEXT_ARRAY, regexp_matches),
    f("regexp_matches", &[TEXT, TEXT, TEXT], TEXT_ARRAY, regexp_matches),
    f("regexp_like", &[TEXT, TEXT], BOOL, regexp_like),
    f("regexp_like", &[TEXT, TEXT, TEXT], BOOL, regexp_like),
    f("regexp_count", &[TEXT, TEXT], INT4, regexp_count),
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

/// like_matches reports whether text matches a LIKE pattern whose escape character is a backslash.
pub(super) fn like_matches(text: &[char], pattern: &[char]) -> Result<bool> {
    let (mut t, mut p) = (0, 0);
    let (mut star_p, mut star_t) = (None, 0);
    while t < text.len() {
        if p < pattern.len() {
            match pattern[p] {
                '%' => {
                    star_p = Some(p);
                    star_t = t;
                    p += 1;
                    continue;
                }
                '_' => {
                    t += 1;
                    p += 1;
                    continue;
                }
                '\\' => {
                    let Some(&c) = pattern.get(p + 1) else {
                        return Err(PgError::new(
                            code::INVALID_ESCAPE_SEQUENCE,
                            "LIKE pattern must not end with escape character",
                        ));
                    };
                    if c == text[t] {
                        t += 1;
                        p += 2;
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
                star_t += 1;
                t = star_t;
                p = s + 1;
            }
            None => return Ok(false),
        }
    }
    while p < pattern.len() && pattern[p] == '%' {
        p += 1;
    }
    if p + 1 == pattern.len() && pattern[p] == '\\' {
        return Err(PgError::new(code::INVALID_ESCAPE_SEQUENCE, "LIKE pattern must not end with escape character"));
    }
    Ok(p == pattern.len())
}

/// like_value matches LIKE or ILIKE.
fn like_value(args: &[Value], fold: bool) -> Result<bool> {
    let fold_chars =
        |s: &str| -> Vec<char> { if fold { s.to_lowercase().chars().collect() } else { s.chars().collect() } };
    like_matches(&fold_chars(text(&args[0])), &fold_chars(text(&args[1])))
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

/// compile compiles a Postgres regular expression with flags, as an advanced regular expression where `.` matches
/// newlines.
fn compile(pattern: &str, flags: &str) -> Result<Regex> {
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

/// regex_error summarizes a regular expression error as briefly as Postgres' messages do.
fn regex_error(err: &regex::Error) -> String {
    let text = err.to_string();
    let lower = text.to_lowercase();
    if lower.contains("unclosed group") || lower.contains("unopened group") {
        "parentheses () not balanced".into()
    } else if lower.contains("unclosed character class") {
        "brackets [] not balanced".into()
    } else if lower.contains("repetition") {
        "quantifier operand invalid".into()
    } else {
        text.lines().last().unwrap_or_default().trim().to_string()
    }
}

/// translate_regex rewrites Postgres' regular expression escapes that the regex crate spells differently.
fn translate_regex(pattern: &str) -> String {
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
    Ok(compile(text(&args[1]), if fold { "i" } else { "" })?.is_match(text(&args[0])))
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
    Ok(Value::Text(if flags.contains('g') {
        regex.replace_all(source, replace.as_str()).into_owned()
    } else {
        regex.replace(source, replace.as_str()).into_owned()
    }))
}

/// captures returns a match's capture groups as a text array, or the whole match without groups.
fn captures(regex: &Regex, caps: &regex::Captures<'_>) -> Value {
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
    Ok(regex.captures(text(&args[0])).map_or(Value::Null, |c| captures(&regex, &c)))
}

/// regexp_matches returns the capture groups of the first match, or of every match with the g flag, as rows.
fn regexp_matches(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let flags = flags(args, 2);
    let regex = compile(text(&args[1]), flags)?;
    let source = text(&args[0]);
    let rows: Vec<Value> = if flags.contains('g') {
        regex.captures_iter(source).map(|c| captures(&regex, &c)).collect()
    } else {
        regex.captures(source).map(|c| captures(&regex, &c)).into_iter().collect()
    };
    Ok(Value::Set(rows))
}

/// regexp_like reports whether text matches.
fn regexp_like(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(compile(text(&args[1]), flags(args, 2))?.is_match(text(&args[0]))))
}

/// regexp_count counts the matches.
fn regexp_count(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(compile(text(&args[1]), "")?.find_iter(text(&args[0])).count() as i32))
}

/// split splits text at the matches, as Postgres does, ignoring empty matches at the ends of the text or right after
/// another match.
fn split(args: &[Value]) -> Result<Vec<Value>> {
    let regex = compile(text(&args[1]), flags(args, 2))?;
    let source = text(&args[0]);
    let mut parts = Vec::new();
    let mut start = 0;
    for m in regex.find_iter(source) {
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
    Ok(match regex.captures(text(&args[0])) {
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
    Ok(match compiled.captures(text(&args[0])) {
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
