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

use crate::decode::reader::Reader;

/// float8_text renders a float8 the way Postgres does with extra_float_digits at its default of 1: the shortest
/// digits that round-trip, in fixed notation for decimal exponents from -4 up to 14 and scientific otherwise.
pub(crate) fn float8_text(value: f64) -> String {
    float_text(value.is_nan(), value.is_infinite(), value.is_sign_negative(), value == 0.0, &format!("{value:e}"), 15)
}

/// float4_text renders a float4 like float8_text, using fixed notation for decimal exponents from -4 up to 5.
pub(crate) fn float4_text(value: f32) -> String {
    float_text(value.is_nan(), value.is_infinite(), value.is_sign_negative(), value == 0.0, &format!("{value:e}"), 6)
}

/// float_text formats a float from Rust's shortest scientific rendering, following Postgres' Ryu-based output.
fn float_text(nan: bool, infinite: bool, negative: bool, zero: bool, scientific: &str, fixed_limit: i32) -> String {
    if nan {
        return "NaN".to_string();
    }
    if infinite {
        return if negative { "-Infinity".to_string() } else { "Infinity".to_string() };
    }
    if zero {
        return if negative { "-0".to_string() } else { "0".to_string() };
    }
    let unsigned = scientific.trim_start_matches('-');
    let (mantissa, exponent) = unsigned.split_once('e').expect("scientific notation has an exponent");
    let exponent: i32 = exponent.parse().expect("exponent is an integer");
    let digits: String = mantissa.chars().filter(|c| *c != '.').collect();
    let mut text = String::new();
    if negative {
        text.push('-');
    }
    if (-4..fixed_limit).contains(&exponent) {
        if exponent < 0 {
            text.push_str("0.");
            text.push_str(&"0".repeat((-exponent - 1) as usize));
            text.push_str(&digits);
        } else {
            let whole_len = exponent as usize + 1;
            if digits.len() <= whole_len {
                text.push_str(&digits);
                text.push_str(&"0".repeat(whole_len - digits.len()));
            } else {
                text.push_str(&digits[..whole_len]);
                text.push('.');
                text.push_str(&digits[whole_len..]);
            }
        }
    } else {
        text.push_str(&digits[..1]);
        if digits.len() > 1 {
            text.push('.');
            text.push_str(&digits[1..]);
        }
        text.push('e');
        text.push(if exponent < 0 { '-' } else { '+' });
        text.push_str(&format!("{:02}", exponent.abs()));
    }
    text
}

/// numeric_text renders a binary numeric the way Postgres' numeric_out does.
pub(crate) fn numeric_text(r: &mut Reader<'_>) -> Result<String, String> {
    let ndigits = r.i16()?;
    let weight = r.i16()? as i32;
    let sign = r.u16()?;
    let dscale = r.u16()? as i32;
    if ndigits < 0 {
        return Err(format!("invalid numeric digit count {ndigits}"));
    }
    let mut digits = Vec::with_capacity(ndigits as usize);
    for _ in 0..ndigits {
        let digit = r.i16()?;
        if !(0..10000).contains(&digit) {
            return Err(format!("invalid numeric digit {digit}"));
        }
        digits.push(digit as i32);
    }
    match sign {
        0x0000 | 0x4000 => {}
        0xC000 => return Ok("NaN".to_string()),
        0xD000 => return Ok("Infinity".to_string()),
        0xF000 => return Ok("-Infinity".to_string()),
        _ => return Err(format!("invalid numeric sign {sign:#x}")),
    }
    let digit_at =
        |index: i32| -> i32 { if index >= 0 && (index as usize) < digits.len() { digits[index as usize] } else { 0 } };
    let mut text = String::new();
    if sign == 0x4000 {
        text.push('-');
    }
    if weight < 0 {
        text.push('0');
    } else {
        for group in 0..=weight {
            let digit = digit_at(group);
            if group == 0 {
                text.push_str(&digit.to_string());
            } else {
                text.push_str(&format!("{digit:04}"));
            }
        }
    }
    if dscale > 0 {
        text.push('.');
        let mut fraction = String::new();
        let mut group = weight + 1;
        while (fraction.len() as i32) < dscale {
            fraction.push_str(&format!("{:04}", digit_at(group)));
            group += 1;
        }
        fraction.truncate(dscale as usize);
        text.push_str(&fraction);
    }
    Ok(text)
}
