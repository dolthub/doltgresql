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

//! Numeric formatting with templates, ported from the to_char and to_number half of Postgres' formatting.c, in the C
//! locale, whose signs, decimal point, thousands separator, and currency symbol are Postgres' defaults.

use crate::error::{PgError, Result, code};
use crate::numeric::Numeric;

/// Key is a template pattern of a number, as Postgres' NUM_ ids name them, where the lowercase variants are the ones
/// that print differently.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
enum Key {
    Comma,
    Dec,
    Zero,
    Nine,
    B,
    C,
    D,
    E,
    Fm,
    G,
    L,
    Mi,
    Pl,
    Pr,
    Rn,
    Sg,
    Sp,
    S,
    Th,
    V,
    RnLower,
    ThLower,
}

/// KEYWORDS are the patterns in the order Postgres' NUM_keywords tries them, each with the key it stands for.
const KEYWORDS: &[(&str, Key)] = &[
    (",", Key::Comma),
    (".", Key::Dec),
    ("0", Key::Zero),
    ("9", Key::Nine),
    ("B", Key::B),
    ("C", Key::C),
    ("D", Key::D),
    ("EEEE", Key::E),
    ("FM", Key::Fm),
    ("G", Key::G),
    ("L", Key::L),
    ("MI", Key::Mi),
    ("PL", Key::Pl),
    ("PR", Key::Pr),
    ("RN", Key::Rn),
    ("SG", Key::Sg),
    ("SP", Key::Sp),
    ("S", Key::S),
    ("TH", Key::Th),
    ("V", Key::V),
    ("b", Key::B),
    ("c", Key::C),
    ("d", Key::D),
    ("eeee", Key::E),
    ("fm", Key::Fm),
    ("g", Key::G),
    ("l", Key::L),
    ("mi", Key::Mi),
    ("pl", Key::Pl),
    ("pr", Key::Pr),
    ("rn", Key::RnLower),
    ("sg", Key::Sg),
    ("sp", Key::Sp),
    ("s", Key::S),
    ("th", Key::ThLower),
    ("v", Key::V),
];

/// Node is a part of a template: a pattern, or a character printed as it is.
#[derive(Clone, Debug)]
enum Node {
    Action(Key),
    Char(String),
}

/// The flags of a template, as Postgres' NUM_F_ flags name them.
const DECIMAL: u32 = 1 << 1;
const LDECIMAL: u32 = 1 << 2;
const ZERO: u32 = 1 << 3;
const BLANK: u32 = 1 << 4;
const FILLMODE: u32 = 1 << 5;
const LSIGN: u32 = 1 << 6;
const BRACKET: u32 = 1 << 7;
const MINUS: u32 = 1 << 8;
const PLUS: u32 = 1 << 9;
const ROMAN: u32 = 1 << 10;
const MULTI: u32 = 1 << 11;
const PLUS_POST: u32 = 1 << 12;
const MINUS_POST: u32 = 1 << 13;
const EEEE: u32 = 1 << 14;

/// The places of a locale sign, as Postgres' NUM_LSIGN_ values name them.
const LSIGN_PRE: i32 = -1;
const LSIGN_POST: i32 = 1;
const LSIGN_NONE: i32 = 0;

/// Desc is what a template says about the number, as Postgres' NUMDesc holds it: its digits before and after the
/// decimal point, where the locale sign goes, its flags, how many digits come before the sign, the digits V shifts
/// by, and where zero padding starts and ends.
#[derive(Clone, Default, Debug)]
struct Desc {
    pre: i32,
    post: i32,
    lsign: i32,
    flag: u32,
    pre_lsign_num: i32,
    multi: i32,
    zero_start: i32,
    zero_end: i32,
}

impl Desc {
    /// is reports whether the template has a flag.
    fn is(&self, flag: u32) -> bool {
        self.flag & flag != 0
    }

    /// prepare records what a pattern of a template says about the number, failing as Postgres' NUMDesc_prepare does
    /// for patterns that conflict.
    fn prepare(&mut self, key: Key) -> Result<()> {
        let syntax = |message: &str| Err(PgError::new(code::SYNTAX_ERROR, message.to_string()));
        if self.is(EEEE) && key != Key::E {
            return syntax("\"EEEE\" must be the last pattern used");
        }
        match key {
            Key::Nine => {
                if self.is(BRACKET) {
                    return syntax("\"9\" must be ahead of \"PR\"");
                }
                if self.is(MULTI) {
                    self.multi += 1;
                } else if self.is(DECIMAL) {
                    self.post += 1;
                } else {
                    self.pre += 1;
                }
            }
            Key::Zero => {
                if self.is(BRACKET) {
                    return syntax("\"0\" must be ahead of \"PR\"");
                }
                if !self.is(ZERO) && !self.is(DECIMAL) {
                    self.flag |= ZERO;
                    self.zero_start = self.pre + 1;
                }
                if !self.is(DECIMAL) {
                    self.pre += 1;
                } else {
                    self.post += 1;
                }
                self.zero_end = self.pre + self.post;
            }
            Key::B => {
                if self.pre == 0 && self.post == 0 && !self.is(ZERO) {
                    self.flag |= BLANK;
                }
            }
            Key::D | Key::Dec => {
                if key == Key::D {
                    self.flag |= LDECIMAL;
                }
                if self.is(DECIMAL) {
                    return syntax("multiple decimal points");
                }
                if self.is(MULTI) {
                    return syntax("cannot use \"V\" and decimal point together");
                }
                self.flag |= DECIMAL;
            }
            Key::Fm => self.flag |= FILLMODE,
            Key::S => {
                if self.is(LSIGN) {
                    return syntax("cannot use \"S\" twice");
                }
                if self.is(PLUS) || self.is(MINUS) || self.is(BRACKET) {
                    return syntax("cannot use \"S\" and \"PL\"/\"MI\"/\"SG\"/\"PR\" together");
                }
                if !self.is(DECIMAL) {
                    self.lsign = LSIGN_PRE;
                    self.pre_lsign_num = self.pre;
                    self.flag |= LSIGN;
                } else if self.lsign == LSIGN_NONE {
                    self.lsign = LSIGN_POST;
                    self.flag |= LSIGN;
                }
            }
            Key::Mi => {
                if self.is(LSIGN) {
                    return syntax("cannot use \"S\" and \"MI\" together");
                }
                self.flag |= MINUS;
                if self.is(DECIMAL) {
                    self.flag |= MINUS_POST;
                }
            }
            Key::Pl => {
                if self.is(LSIGN) {
                    return syntax("cannot use \"S\" and \"PL\" together");
                }
                self.flag |= PLUS;
                if self.is(DECIMAL) {
                    self.flag |= PLUS_POST;
                }
            }
            Key::Sg => {
                if self.is(LSIGN) {
                    return syntax("cannot use \"S\" and \"SG\" together");
                }
                self.flag |= MINUS | PLUS;
            }
            Key::Pr => {
                if self.is(LSIGN) || self.is(PLUS) || self.is(MINUS) {
                    return syntax("cannot use \"PR\" and \"S\"/\"PL\"/\"MI\"/\"SG\" together");
                }
                self.flag |= BRACKET;
            }
            Key::Rn | Key::RnLower => self.flag |= ROMAN,
            Key::V => {
                if self.is(DECIMAL) {
                    return syntax("cannot use \"V\" and decimal point together");
                }
                self.flag |= MULTI;
            }
            Key::E => {
                if self.is(EEEE) {
                    return syntax("cannot use \"EEEE\" twice");
                }
                if [BLANK, FILLMODE, LSIGN, BRACKET, MINUS, PLUS, ROMAN, MULTI].iter().any(|&f| self.is(f)) {
                    return Err(PgError {
                        detail: Some(
                            "\"EEEE\" may only be used together with digit and decimal point patterns.".into(),
                        ),
                        ..PgError::new(code::SYNTAX_ERROR, "\"EEEE\" is incompatible with other formats")
                    });
                }
                self.flag |= EEEE;
            }
            _ => {}
        }
        Ok(())
    }
}

/// parse_template reads a numeric template into its parts and what it says about the number, as Postgres'
/// parse_format does for NUM_FLAG templates.
fn parse_template(template: &str) -> Result<(Vec<Node>, Desc)> {
    let mut desc = Desc::default();
    let mut nodes = Vec::new();
    let mut rest = template;
    while let Some(c) = rest.chars().next() {
        if let Some((name, key)) = KEYWORDS.iter().find(|(name, _)| rest.starts_with(name)) {
            nodes.push(Node::Action(*key));
            desc.prepare(*key)?;
            rest = &rest[name.len()..];
            continue;
        }
        if c == '"' {
            rest = &rest[1..];
            while let Some(c) = rest.chars().next() {
                if c == '"' {
                    rest = &rest[1..];
                    break;
                }
                if c == '\\' && rest.len() > 1 {
                    rest = &rest[1..];
                }
                let c = rest.chars().next().unwrap_or_default();
                nodes.push(Node::Char(c.to_string()));
                rest = &rest[c.len_utf8()..];
            }
            continue;
        }
        if c == '\\' && rest[1..].starts_with('"') {
            rest = &rest[1..];
        }
        let c = rest.chars().next().unwrap_or_default();
        nodes.push(Node::Char(c.to_string()));
        rest = &rest[c.len_utf8()..];
    }
    Ok((nodes, desc))
}

/// roman returns a number in Roman numerals, or fifteen # signs outside 1 to 3999, as Postgres' int_to_roman does.
fn roman(number: i64) -> String {
    if !(1..=3999).contains(&number) {
        return "#".repeat(15);
    }
    const ONES: [&str; 9] = ["I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX"];
    const TENS: [&str; 9] = ["X", "XX", "XXX", "XL", "L", "LX", "LXX", "LXXX", "XC"];
    const HUNDREDS: [&str; 9] = ["C", "CC", "CCC", "CD", "D", "DC", "DCC", "DCCC", "CM"];
    let n = number as usize;
    let mut out = "M".repeat(n / 1000);
    for (digit, table) in [((n / 100) % 10, HUNDREDS), ((n / 10) % 10, TENS), (n % 10, ONES)] {
        if digit > 0 {
            out.push_str(table[digit - 1]);
        }
    }
    out
}

/// ordinal_suffix returns the English ordinal suffix of a number's digits, as Postgres' get_th does.
fn ordinal_suffix(number: &[u8], upper: bool) -> Result<&'static str> {
    let Some(&last) = number.last().filter(|b| b.is_ascii_digit()) else {
        return Err(PgError::new(
            code::INVALID_TEXT_REPRESENTATION,
            format!("\"{}\" is not a number", String::from_utf8_lossy(number)),
        ));
    };
    let teen = number.len() > 1 && number[number.len() - 2] == b'1';
    let index = match last {
        b'1' if !teen => 0,
        b'2' if !teen => 1,
        b'3' if !teen => 2,
        _ => 3,
    };
    Ok(if upper { ["ST", "ND", "RD", "TH"][index] } else { ["st", "nd", "rd", "th"][index] })
}

/// hashes returns the # signs that stand for a number too wide for a template, with its decimal point.
fn hashes(desc: &Desc) -> String {
    let mut out = "#".repeat((desc.pre + desc.post + 1) as usize);
    out.replace_range(desc.pre as usize..desc.pre as usize + 1, ".");
    out
}

/// c_exponential writes a float as C's `%+.*e` does, with a sign and at least two exponent digits.
fn c_exponential(value: f64, precision: usize) -> String {
    let text = format!("{:.*e}", precision, value.abs());
    let (mantissa, exponent) = text.split_once('e').unwrap_or((&text, "0"));
    let exponent: i32 = exponent.parse().unwrap_or(0);
    let sign = if value.is_sign_negative() { '-' } else { '+' };
    format!("{sign}{mantissa}e{}{:02}", if exponent < 0 { '-' } else { '+' }, exponent.unsigned_abs())
}

/// fixed writes a float as Postgres' snprintf writes `%.*f`, which spells out the infinities and NaN.
fn fixed(value: f64, precision: usize) -> String {
    match value {
        v if v.is_nan() => "NaN".into(),
        v if v.is_infinite() => (if v < 0.0 { "-Infinity" } else { "Infinity" }).into(),
        v => format!("{v:.precision$}"),
    }
}

/// scientific writes a numeric in scientific notation with a number of fraction digits, as Postgres' numeric_out_sci
/// does.
fn scientific(value: &Numeric, scale: i32) -> String {
    let text = value.to_string();
    if matches!(text.as_str(), "NaN" | "Infinity" | "-Infinity") {
        return text;
    }
    let (negative, digits) = match text.strip_prefix('-') {
        Some(rest) => (true, rest),
        None => (false, text.as_str()),
    };
    let (whole, fraction) = digits.split_once('.').unwrap_or((digits, ""));
    let all: String = format!("{whole}{fraction}");
    let leading = all.bytes().take_while(|&b| b == b'0').count();
    let (exponent, significant) = match leading == all.len() {
        true => (0, "0".to_string()),
        false => (whole.len() as i32 - 1 - leading as i32, all[leading..].to_string()),
    };
    let significand = format!("{}.{}", &significant[..1], &significant[1..]);
    let rounded = Numeric::parse(&significand).map(|n| n.with_scale(scale.max(0) as u32)).unwrap_or(Numeric::NaN);
    let sign = if negative { "-" } else { "" };
    format!("{sign}{rounded}e{}{:02}", if exponent < 0 { '-' } else { '+' }, exponent.unsigned_abs())
}

/// Proc is the state of Postgres' NUM_processor as it writes a number for to_char: the template's description, the
/// sign, how far it is through the digits, where it is in the number, and the text so far.
struct Proc {
    desc: Desc,
    sign: u8,
    sign_wrote: bool,
    num_count: i32,
    num_in: bool,
    num_curr: i32,
    out_pre_spaces: i32,
    number: Vec<u8>,
    number_p: usize,
    last_relevant: Option<usize>,
    out: Vec<u8>,
}

impl Proc {
    /// at returns the byte of the number at a position, or 0 past its end.
    fn at(&self, i: usize) -> u8 {
        self.number.get(i).copied().unwrap_or(0)
    }

    /// predec_space reports whether the digit before the decimal point of a number below one prints as a space, as
    /// Postgres' IS_PREDEC_SPACE decides.
    fn predec_space(&self) -> bool {
        !self.desc.is(ZERO) && self.number_p == 0 && self.at(0) == b'0' && self.desc.post != 0
    }

    /// last_is_point reports whether the last relevant digit is the decimal point.
    fn last_is_point(&self) -> bool {
        self.last_relevant.is_some_and(|i| self.at(i) == b'.')
    }

    /// digit_to_char writes the part of the number that a digit or decimal point pattern stands for, as Postgres'
    /// NUM_numpart_to_char does.
    fn digit_to_char(&mut self, key: Key) {
        if self.desc.is(ROMAN) {
            return;
        }
        self.num_in = false;
        let zero_here = self.desc.is(ZERO) && self.desc.zero_start == self.num_curr;
        if !self.sign_wrote
            && (self.num_curr >= self.out_pre_spaces || zero_here)
            && (!self.predec_space() || self.last_is_point())
        {
            if self.desc.is(LSIGN) {
                if self.desc.lsign == LSIGN_PRE {
                    self.out.push(if self.sign == b'-' { b'-' } else { b'+' });
                    self.sign_wrote = true;
                }
            } else if self.desc.is(BRACKET) {
                self.out.push(if self.sign == b'+' { b' ' } else { b'<' });
                self.sign_wrote = true;
            } else if self.sign == b'+' {
                if !self.desc.is(FILLMODE) {
                    self.out.push(b' ');
                }
                self.sign_wrote = true;
            } else if self.sign == b'-' {
                self.out.push(b'-');
                self.sign_wrote = true;
            }
        }
        if matches!(key, Key::Nine | Key::Zero | Key::D | Key::Dec) {
            if self.num_curr < self.out_pre_spaces && (self.desc.zero_start > self.num_curr || !self.desc.is(ZERO)) {
                if !self.desc.is(FILLMODE) {
                    self.out.push(b' ');
                }
            } else if self.desc.is(ZERO) && self.num_curr < self.out_pre_spaces && self.desc.zero_start <= self.num_curr
            {
                self.out.push(b'0');
                self.num_in = true;
            } else {
                if self.at(self.number_p) == b'.' {
                    if !self.last_is_point() || self.desc.is(FILLMODE) {
                        self.out.push(b'.');
                    }
                } else if self.last_relevant.is_some_and(|l| self.number_p > l) && key != Key::Zero {
                } else if self.predec_space() {
                    if !self.desc.is(FILLMODE) {
                        self.out.push(b' ');
                    } else if self.last_is_point() {
                        self.out.push(b'0');
                    }
                } else {
                    self.out.push(self.at(self.number_p));
                    self.num_in = true;
                }
                if self.at(self.number_p) != 0 {
                    self.number_p += 1;
                }
            }
            let mut end = self.num_count + i32::from(self.out_pre_spaces != 0) + i32::from(self.desc.is(DECIMAL));
            if self.last_relevant == Some(self.number_p) {
                end = self.num_curr;
            }
            if self.num_curr + 1 == end {
                if self.sign_wrote && self.desc.is(BRACKET) {
                    self.out.push(if self.sign == b'+' { b' ' } else { b'>' });
                } else if self.desc.is(LSIGN) && self.desc.lsign == LSIGN_POST {
                    self.out.push(if self.sign == b'-' { b'-' } else { b'+' });
                }
            }
        }
        self.num_curr += 1;
    }
}

/// `format` formats a number's text, already rounded and signed as Postgres' to_char functions prepare it, by a
/// template, as Postgres' NUM_processor does for to_char, ending the result at its first NUL byte like a C string.
fn format(nodes: &[Node], mut desc: Desc, number: String, out_pre_spaces: i32, sign: u8) -> Result<String> {
    if desc.zero_start > 0 {
        desc.zero_start -= 1;
    }
    if desc.is(EEEE) {
        return Ok(number);
    }
    let mut out_pre_spaces = out_pre_spaces;
    let mut sign = sign;
    if desc.is(ROMAN) {
        desc.lsign = 0;
        desc.pre_lsign_num = 0;
        desc.post = 0;
        desc.pre = 0;
        out_pre_spaces = 0;
        sign = 0;
        desc.flag = ROMAN | (desc.flag & FILLMODE);
    }
    let sign_wrote;
    if desc.is(PLUS) || desc.is(MINUS) {
        sign_wrote = !(desc.is(PLUS) && !desc.is(MINUS));
    } else {
        if sign != b'-' {
            if desc.is(BRACKET) && desc.is(FILLMODE) {
                desc.flag &= !BRACKET;
            }
            if desc.is(MINUS) {
                desc.flag &= !MINUS;
            }
        } else if sign != b'+' && desc.is(PLUS) {
            desc.flag &= !PLUS;
        }
        sign_wrote = sign == b'+' && desc.is(FILLMODE) && !desc.is(LSIGN);
        if desc.lsign == LSIGN_PRE && desc.pre == desc.pre_lsign_num {
            desc.lsign = LSIGN_POST;
        }
    }
    let number = number.into_bytes();
    let mut proc = Proc {
        num_count: desc.post + desc.pre - 1,
        desc,
        sign,
        sign_wrote,
        num_in: false,
        num_curr: 0,
        out_pre_spaces,
        number,
        number_p: 0,
        last_relevant: None,
        out: Vec::new(),
    };
    if proc.desc.is(FILLMODE) && proc.desc.is(DECIMAL) {
        let point = proc.number.iter().position(|&b| b == b'.');
        proc.last_relevant = point.map(|p| (p + 1..proc.number.len()).rfind(|&i| proc.number[i] != b'0').unwrap_or(p));
        if let Some(last) = proc.last_relevant
            && proc.desc.zero_end > proc.out_pre_spaces
        {
            let last_zero = ((proc.number.len() as i32 - 1).min(proc.desc.zero_end - proc.out_pre_spaces)) as usize;
            if last < last_zero {
                proc.last_relevant = Some(last_zero);
            }
        }
    }
    if !proc.sign_wrote && proc.out_pre_spaces == 0 {
        proc.num_count += 1;
    }
    for node in nodes {
        let key = match node {
            Node::Char(text) => {
                proc.out.extend_from_slice(text.as_bytes());
                continue;
            }
            Node::Action(key) => *key,
        };
        let fill = proc.desc.is(FILLMODE);
        match key {
            Key::Nine | Key::Zero | Key::Dec | Key::D => proc.digit_to_char(key),
            Key::Comma | Key::G => match (proc.num_in, fill) {
                (false, true) => {}
                (false, false) => proc.out.push(b' '),
                (true, _) => proc.out.push(b','),
            },
            Key::L => proc.out.push(b' '),
            Key::Rn | Key::RnLower => {
                let text = String::from_utf8_lossy(&proc.number[proc.number_p..]).into_owned();
                let text = if key == Key::RnLower { text.to_ascii_lowercase() } else { text };
                match fill {
                    true => proc.out.extend_from_slice(text.as_bytes()),
                    false => proc.out.extend_from_slice(format!("{text:>15}").as_bytes()),
                }
            }
            Key::Th | Key::ThLower => {
                if proc.desc.is(ROMAN) || proc.at(0) == b'#' || proc.sign == b'-' || proc.desc.is(DECIMAL) {
                    continue;
                }
                let suffix = ordinal_suffix(&proc.number, key == Key::Th)?;
                proc.out.extend_from_slice(suffix.as_bytes());
            }
            Key::Mi => match (proc.sign, fill) {
                (b'-', _) => proc.out.push(b'-'),
                (_, true) => {}
                _ => proc.out.push(b' '),
            },
            Key::Pl => match (proc.sign, fill) {
                (b'+', _) => proc.out.push(b'+'),
                (_, true) => {}
                _ => proc.out.push(b' '),
            },
            Key::Sg => proc.out.push(proc.sign),
            _ => {}
        }
    }
    let end = proc.out.iter().position(|&b| b == 0).unwrap_or(proc.out.len());
    Ok(String::from_utf8_lossy(&proc.out[..end]).into_owned())
}

/// Reader is the state of Postgres' NUM_processor as it reads a number for to_number: the input and where it is, and
/// the number read so far, whose first byte is its sign or a space.
struct Reader<'a> {
    input: &'a [u8],
    at: usize,
    number: Vec<u8>,
    read_dec: bool,
    read_pre: i32,
    read_post: i32,
}

impl Reader<'_> {
    /// byte returns the input byte at a position, or 0 past the end.
    fn byte(&self, i: usize) -> u8 {
        self.input.get(i).copied().unwrap_or(0)
    }

    /// done reports whether the reader has reached the end of the input.
    fn done(&self) -> bool {
        self.at >= self.input.len()
    }

    /// skip moves past one character.
    fn skip(&mut self) {
        self.at += 1 + self.input[self.at..].iter().skip(1).take_while(|&&b| b & 0xC0 == 0x80).count();
    }

    /// eat skips up to n characters that cannot be part of a number, as Postgres' NUM_eat_non_data_chars does.
    fn eat(&mut self, n: usize) {
        for _ in 0..n {
            if self.done() || b"0123456789.,+-".contains(&self.byte(self.at)) {
                break;
            }
            self.skip();
        }
    }

    /// digit reads what a digit or decimal point pattern stands for, as Postgres' NUM_numpart_from_char does.
    fn digit(&mut self, desc: &Desc, key: Key) {
        if self.done() {
            return;
        }
        if self.byte(self.at) == b' ' {
            self.at += 1;
        }
        if self.done() {
            return;
        }
        if self.number[0] == b' ' && matches!(key, Key::Nine | Key::Zero) && self.read_pre + self.read_post == 0 {
            let c = self.byte(self.at);
            if desc.is(LSIGN) && desc.lsign == LSIGN_PRE {
                if matches!(c, b'-' | b'+') {
                    self.number[0] = c;
                    self.at += 1;
                }
            } else if c == b'-' || (desc.is(BRACKET) && c == b'<') {
                self.number[0] = b'-';
                self.at += 1;
            } else if c == b'+' {
                self.number[0] = b'+';
                self.at += 1;
            }
        }
        if self.done() {
            return;
        }
        let mut read = false;
        let c = self.byte(self.at);
        if c.is_ascii_digit() {
            if self.read_dec && self.read_post == desc.post {
                return;
            }
            self.number.push(c);
            if self.read_dec {
                self.read_post += 1;
            } else {
                self.read_pre += 1;
            }
            read = true;
        } else if desc.is(DECIMAL) && !self.read_dec && c == b'.' {
            self.number.push(b'.');
            self.read_dec = true;
            read = true;
        }
        if self.done() || self.number[0] != b' ' || self.read_pre + self.read_post == 0 {
            return;
        }
        let next = self.byte(self.at + 1);
        if desc.is(LSIGN) && read && self.at + 1 < self.input.len() && !next.is_ascii_digit() {
            if matches!(next, b'-' | b'+') {
                self.number[0] = next;
                self.at += 1;
            }
        } else if !read && !desc.is(LSIGN) && (desc.is(PLUS) || desc.is(MINUS)) && matches!(c, b'-' | b'+') {
            self.number[0] = c;
        }
    }
}

/// parse_number reads the digits, decimal point, and sign of text by a template, as Postgres' NUM_processor does for
/// to_number, returning the number's text and how many digits followed its decimal point.
fn parse_number(nodes: &[Node], desc: &Desc, input: &[u8]) -> Result<(String, i32)> {
    if desc.is(EEEE) {
        return Err(PgError::new(code::FEATURE_NOT_SUPPORTED, "\"EEEE\" not supported for input"));
    }
    if desc.is(ROMAN) {
        return Err(PgError::new(code::FEATURE_NOT_SUPPORTED, "\"RN\" not supported for input"));
    }
    let mut reader = Reader { input, at: 0, number: vec![b' '], read_dec: false, read_pre: 0, read_post: 0 };
    for node in nodes {
        if reader.done() {
            break;
        }
        let key = match node {
            Node::Char(_) => {
                reader.skip();
                continue;
            }
            Node::Action(key) => *key,
        };
        let c = reader.byte(reader.at);
        match key {
            Key::Nine | Key::Zero | Key::Dec | Key::D => reader.digit(desc, key),
            Key::Comma | Key::G if desc.is(FILLMODE) || c != b',' => continue,
            Key::Comma | Key::G => {}
            Key::L => {
                reader.eat(1);
                continue;
            }
            Key::Th | Key::ThLower => {
                reader.eat(2);
                continue;
            }
            Key::Mi | Key::Pl | Key::Sg => {
                let wanted = match key {
                    Key::Mi => c == b'-',
                    Key::Pl => c == b'+',
                    _ => matches!(c, b'-' | b'+'),
                };
                if !wanted {
                    reader.eat(1);
                    continue;
                }
                reader.number[0] = c;
            }
            _ => continue,
        }
        reader.at += 1;
    }
    let mut number = reader.number;
    if number.last() == Some(&b'.') {
        number.pop();
    }
    Ok((String::from_utf8_lossy(&number).into_owned(), reader.read_post))
}

/// TEMPLATE_LIMIT is the length past which Postgres treats a template as empty.
const TEMPLATE_LIMIT: usize = i32::MAX as usize / 8;

/// numeric_to_char formats a numeric by a template, as Postgres' numeric_to_char does.
pub fn numeric_to_char(value: &Numeric, template: &str) -> Result<String> {
    if template.is_empty() || template.len() >= TEMPLATE_LIMIT {
        return Ok(String::new());
    }
    let (nodes, mut desc) = parse_template(template)?;
    let (mut out_pre_spaces, mut sign) = (0, 0u8);
    let number = if desc.is(ROMAN) {
        let rounded = value.with_scale(0);
        let int = rounded
            .to_i64()
            .filter(|n| i32::try_from(*n).is_ok())
            .ok_or_else(|| PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "integer out of range"))?;
        roman(int)
    } else if desc.is(EEEE) {
        let text = scientific(value, desc.post);
        if matches!(text.as_str(), "NaN" | "Infinity" | "-Infinity") {
            let mut out = "#".repeat((desc.pre + desc.post + 6) as usize);
            out.replace_range(0..1, " ");
            out.replace_range(desc.pre as usize + 1..desc.pre as usize + 2, ".");
            out
        } else if !text.starts_with('-') {
            format!(" {text}")
        } else {
            text
        }
    } else {
        let mut value = value.clone();
        if desc.is(MULTI) {
            value = value.mul(&Numeric::parse(&format!("1e{}", desc.multi))?);
            desc.pre += desc.multi;
        }
        let text = value.with_scale(desc.post.max(0) as u32).to_string();
        let digits = match text.strip_prefix('-') {
            Some(rest) => {
                sign = b'-';
                rest.to_string()
            }
            None => {
                sign = b'+';
                text
            }
        };
        let pre_len = digits.find('.').unwrap_or(digits.len()) as i32;
        if pre_len < desc.pre {
            out_pre_spaces = desc.pre - pre_len;
            digits
        } else if pre_len > desc.pre {
            hashes(&desc)
        } else {
            digits
        }
    };
    format(&nodes, desc, number, out_pre_spaces, sign)
}

/// integer_to_char formats an integer by a template, as Postgres' int4_to_char and int8_to_char do.
pub fn integer_to_char(value: i64, wide: bool, template: &str) -> Result<String> {
    if template.is_empty() || template.len() >= TEMPLATE_LIMIT {
        return Ok(String::new());
    }
    let (nodes, mut desc) = parse_template(template)?;
    let out_of_range = || {
        PgError::new(
            code::NUMERIC_VALUE_OUT_OF_RANGE,
            format!("{} out of range", if wide { "bigint" } else { "integer" }),
        )
    };
    let (mut out_pre_spaces, mut sign) = (0, 0u8);
    let number = if desc.is(ROMAN) {
        let value =
            i32::try_from(value).map_err(|_| PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "integer out of range"))?;
        roman(i64::from(value))
    } else if desc.is(EEEE) {
        match wide {
            true => {
                let text = scientific(&Numeric::from_i64(value), desc.post);
                if text.starts_with('-') { text } else { format!(" {text}") }
            }
            false => {
                let text = c_exponential(value as f64, desc.post.max(0) as usize);
                text.strip_prefix('+').map_or(text.clone(), |rest| format!(" {rest}"))
            }
        }
    } else {
        let mut value = value;
        if desc.is(MULTI) {
            let multi = 10f64.powi(desc.multi).round();
            let limit = if wide { i64::MAX as f64 } else { f64::from(i32::MAX) };
            if multi > limit {
                return Err(out_of_range());
            }
            value = value.checked_mul(multi as i64).ok_or_else(out_of_range)?;
            if !wide && i32::try_from(value).is_err() {
                return Err(out_of_range());
            }
            desc.pre += desc.multi;
        }
        sign = if value < 0 { b'-' } else { b'+' };
        let digits = value.unsigned_abs().to_string();
        let pre_len = digits.len() as i32;
        let digits = match desc.post {
            0 => digits,
            post => format!("{digits}.{}", "0".repeat(post as usize)),
        };
        if pre_len < desc.pre {
            out_pre_spaces = desc.pre - pre_len;
            digits
        } else if pre_len > desc.pre {
            hashes(&desc)
        } else {
            digits
        }
    };
    format(&nodes, desc, number, out_pre_spaces, sign)
}

/// float_to_char formats a float by a template, as Postgres' float4_to_char and float8_to_char do, keeping only the
/// digits the type holds.
pub fn float_to_char(value: f64, digits: i32, template: &str) -> Result<String> {
    if template.is_empty() || template.len() >= TEMPLATE_LIMIT {
        return Ok(String::new());
    }
    let (nodes, mut desc) = parse_template(template)?;
    let (mut out_pre_spaces, mut sign) = (0, 0u8);
    let number = if desc.is(ROMAN) {
        roman(value.round_ties_even() as i32 as i64)
    } else if desc.is(EEEE) {
        if value.is_nan() || value.is_infinite() {
            let mut out = "#".repeat((desc.pre + desc.post + 6) as usize);
            out.replace_range(0..1, " ");
            out.replace_range(desc.pre as usize + 1..desc.pre as usize + 2, ".");
            out
        } else {
            let text = c_exponential(value, desc.post.max(0) as usize);
            text.strip_prefix('+').map_or(text.clone(), |rest| format!(" {rest}"))
        }
    } else {
        let mut value = value;
        if desc.is(MULTI) {
            value *= 10f64.powi(desc.multi);
            desc.pre += desc.multi;
        }
        let whole = fixed(value.abs(), 0).len() as i32;
        if whole >= digits {
            desc.post = 0;
        } else if whole + desc.post > digits {
            desc.post = digits - whole;
        }
        let text = fixed(value, desc.post.max(0) as usize);
        let number = match text.strip_prefix('-') {
            Some(rest) => {
                sign = b'-';
                rest.to_string()
            }
            None => {
                sign = b'+';
                text
            }
        };
        let pre_len = number.find('.').unwrap_or(number.len()) as i32;
        if pre_len < desc.pre {
            out_pre_spaces = desc.pre - pre_len;
            number
        } else if pre_len > desc.pre {
            hashes(&desc)
        } else {
            number
        }
    };
    format(&nodes, desc, number, out_pre_spaces, sign)
}

/// to_number reads a numeric from text by a template, as Postgres' numeric_to_number does.
pub fn to_number(text: &str, template: &str) -> Result<Option<Numeric>> {
    if template.is_empty() || template.len() >= TEMPLATE_LIMIT {
        return Ok(None);
    }
    let (nodes, desc) = parse_template(template)?;
    let (number, scale) = parse_number(&nodes, &desc, text.as_bytes())?;
    let precision = desc.pre + desc.multi + scale;
    let parsed = Numeric::parse(&number).map_err(|_| {
        PgError::new(code::INVALID_TEXT_REPRESENTATION, format!("invalid input syntax for type numeric: \"{number}\""))
    })?;
    let mut result = parsed.apply_typmod(((precision << 16) | scale) + 4)?;
    if desc.is(MULTI) {
        result = result.mul(&Numeric::parse(&format!("1e-{}", desc.multi))?);
    }
    Ok(Some(result))
}
