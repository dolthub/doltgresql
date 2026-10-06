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

//! Postgres' arbitrary precision numeric type.

use std::cmp::Ordering;

use num_bigint::BigUint;
use num_integer::Integer;
use num_traits::{ToPrimitive, Zero};

use crate::error::{PgError, Result, code};

/// MIN_SIG_DIGITS is the least number of significant digits a division result has.
const MIN_SIG_DIGITS: i64 = 16;

/// MAX_DISPLAY_SCALE is the largest scale an operation chooses for its result.
const MAX_DISPLAY_SCALE: i64 = 1000;

/// Numeric is a numeric value: a finite decimal with its display scale, NaN, or an infinity.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Numeric {
    /// The value is `coefficient / 10^scale`, negated when `negative`, and prints `scale` fractional digits.
    Finite {
        negative: bool,
        coefficient: BigUint,
        scale: u32,
    },
    NaN,
    Infinity,
    NegativeInfinity,
}

/// pow10 returns 10 to the power.
fn pow10(power: u32) -> BigUint {
    BigUint::from(10u32).pow(power)
}

/// invalid returns Postgres' error for text that is not a numeric.
fn invalid(text: &str) -> PgError {
    PgError::new(code::INVALID_TEXT_REPRESENTATION, format!("invalid input syntax for type numeric: \"{text}\""))
}

impl Numeric {
    /// zero returns zero with the scale.
    pub fn zero(scale: u32) -> Numeric {
        Numeric::Finite { negative: false, coefficient: BigUint::zero(), scale }
    }

    /// from_i64 returns an integer as a numeric.
    pub fn from_i64(value: i64) -> Numeric {
        Numeric::Finite { negative: value < 0, coefficient: BigUint::from(value.unsigned_abs()), scale: 0 }
    }

    /// parse reads a numeric as Postgres' numeric_in does.
    pub fn parse(text: &str) -> Result<Numeric> {
        let trimmed = text.trim_matches(|c: char| c.is_ascii_whitespace());
        match trimmed.to_ascii_lowercase().as_str() {
            "nan" => return Ok(Numeric::NaN),
            "infinity" | "+infinity" | "inf" | "+inf" => return Ok(Numeric::Infinity),
            "-infinity" | "-inf" => return Ok(Numeric::NegativeInfinity),
            _ => {}
        }
        let (negative, rest) = match trimmed.as_bytes().first() {
            Some(b'-') => (true, &trimmed[1..]),
            Some(b'+') => (false, &trimmed[1..]),
            _ => (false, trimmed),
        };
        let (mantissa, exponent) = match rest.find(['e', 'E']) {
            Some(i) => {
                let exponent: i64 = rest[i + 1..].parse().map_err(|_| invalid(text))?;
                (&rest[..i], exponent)
            }
            None => (rest, 0),
        };
        let (whole, fraction) = mantissa.split_once('.').unwrap_or((mantissa, ""));
        if (whole.is_empty() && fraction.is_empty())
            || !whole.bytes().all(|b| b.is_ascii_digit())
            || !fraction.bytes().all(|b| b.is_ascii_digit())
        {
            return Err(invalid(text));
        }
        if exponent.abs() > 1000 * 1000 {
            return Err(invalid(text));
        }
        let digits = format!("{whole}{fraction}");
        let mut coefficient = digits.parse::<BigUint>().unwrap_or_default();
        let scale = fraction.len() as i64 - exponent;
        let scale = if scale < 0 {
            coefficient *= pow10((-scale) as u32);
            0
        } else {
            scale as u32
        };
        Ok(Numeric::Finite { negative: negative && !coefficient.is_zero(), coefficient, scale })
    }

    /// scale returns the display scale of a finite value, and 0 otherwise.
    pub fn scale(&self) -> u32 {
        match self {
            Numeric::Finite { scale, .. } => *scale,
            _ => 0,
        }
    }

    /// with_scale returns the value rounded half away from zero, or padded, to the scale.
    pub fn with_scale(&self, new_scale: u32) -> Numeric {
        let Numeric::Finite { negative, coefficient, scale } = self else { return self.clone() };
        let coefficient = match new_scale.cmp(scale) {
            Ordering::Equal => coefficient.clone(),
            Ordering::Greater => coefficient * pow10(new_scale - scale),
            Ordering::Less => {
                let divisor = pow10(scale - new_scale);
                let (quotient, remainder) = coefficient.div_rem(&divisor);
                if remainder * 2u32 >= divisor { quotient + 1u32 } else { quotient }
            }
        };
        Numeric::Finite { negative: *negative && !coefficient.is_zero(), coefficient, scale: new_scale }
    }

    /// apply_typmod rounds the value to a numeric(precision, scale) typmod, failing when it has too many digits.
    pub fn apply_typmod(&self, typmod: i32) -> Result<Numeric> {
        if typmod < 4 {
            return Ok(self.clone());
        }
        let typmod = typmod - 4;
        let precision = (typmod >> 16) & 0xffff;
        let scale = ((typmod & 0x7ff) ^ 1024) - 1024;
        if !matches!(self, Numeric::Finite { .. }) {
            if matches!(self, Numeric::NaN) {
                return Ok(Numeric::NaN);
            }
            return Err(PgError {
                detail: Some(format!(
                    "A field with precision {precision}, scale {scale} cannot hold an infinite value."
                )),
                ..PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "numeric field overflow")
            });
        }
        let rounded = if scale >= 0 {
            self.with_scale(scale as u32)
        } else {
            let shifted = self.with_scale(0);
            let Numeric::Finite { negative, coefficient, .. } = shifted else { unreachable!() };
            let unit = pow10((-scale) as u32);
            let (quotient, remainder) = coefficient.div_rem(&unit);
            let quotient = if remainder * 2u32 >= unit { quotient + 1u32 } else { quotient };
            Numeric::Finite { negative, coefficient: quotient * unit, scale: 0 }
        };
        let Numeric::Finite { coefficient, scale: actual, .. } = &rounded else { unreachable!() };
        let integer_digits = {
            let integer = coefficient / pow10(*actual);
            if integer.is_zero() { 0 } else { integer.to_string().len() as i32 }
        };
        if integer_digits > precision - scale {
            return Err(PgError {
                detail: Some(if precision - scale <= 0 {
                    format!(
                        "A field with precision {precision}, scale {scale} must round to an absolute value less than 1."
                    )
                } else {
                    format!(
                        "A field with precision {precision}, scale {scale} must round to an absolute value less than \
                         10^{}.",
                        precision - scale
                    )
                }),
                ..PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "numeric field overflow")
            });
        }
        Ok(rounded)
    }

    /// signed_at returns a finite value's sign and coefficient at the scale.
    fn signed_at(&self, scale: u32) -> (bool, BigUint) {
        match self.with_scale(scale) {
            Numeric::Finite { negative, coefficient, .. } => (negative, coefficient),
            _ => (false, BigUint::zero()),
        }
    }

    /// finite returns a value from a sign and a coefficient at a scale.
    fn finite(negative: bool, coefficient: BigUint, scale: u32) -> Numeric {
        Numeric::Finite { negative: negative && !coefficient.is_zero(), coefficient, scale }
    }

    /// is_negative reports whether the value is below zero.
    fn is_negative(&self) -> bool {
        matches!(self, Numeric::Finite { negative: true, .. } | Numeric::NegativeInfinity)
    }

    /// add returns the sum.
    pub fn add(&self, other: &Numeric) -> Numeric {
        match (self, other) {
            (Numeric::NaN, _) | (_, Numeric::NaN) => Numeric::NaN,
            (Numeric::Infinity, Numeric::NegativeInfinity) | (Numeric::NegativeInfinity, Numeric::Infinity) => {
                Numeric::NaN
            }
            (Numeric::Infinity | Numeric::NegativeInfinity, _) => self.clone(),
            (_, Numeric::Infinity | Numeric::NegativeInfinity) => other.clone(),
            _ => {
                let scale = self.scale().max(other.scale());
                let (ln, l) = self.signed_at(scale);
                let (rn, r) = other.signed_at(scale);
                if ln == rn {
                    Numeric::finite(ln, l + r, scale)
                } else if l >= r {
                    Numeric::finite(ln, l - r, scale)
                } else {
                    Numeric::finite(rn, r - l, scale)
                }
            }
        }
    }

    /// negate returns the value with its sign flipped.
    pub fn negate(&self) -> Numeric {
        match self {
            Numeric::Finite { negative, coefficient, scale } => Numeric::finite(!negative, coefficient.clone(), *scale),
            Numeric::NaN => Numeric::NaN,
            Numeric::Infinity => Numeric::NegativeInfinity,
            Numeric::NegativeInfinity => Numeric::Infinity,
        }
    }

    /// sub returns the difference.
    pub fn sub(&self, other: &Numeric) -> Numeric {
        self.add(&other.negate())
    }

    /// mul returns the product, whose scale is the sum of the operands' scales.
    pub fn mul(&self, other: &Numeric) -> Numeric {
        match (self, other) {
            (Numeric::NaN, _) | (_, Numeric::NaN) => Numeric::NaN,
            (
                Numeric::Finite { negative: ln, coefficient: l, scale: ls },
                Numeric::Finite { negative: rn, coefficient: r, scale: rs },
            ) => {
                let product = Numeric::finite(ln != rn, l * r, ls + rs);
                if (ls + rs) as i64 > MAX_DISPLAY_SCALE {
                    product.with_scale(MAX_DISPLAY_SCALE as u32)
                } else {
                    product
                }
            }
            _ => {
                if self.is_zero() || other.is_zero() {
                    return Numeric::NaN;
                }
                if self.is_negative() != other.is_negative() { Numeric::NegativeInfinity } else { Numeric::Infinity }
            }
        }
    }

    /// is_zero reports whether the value is zero.
    pub fn is_zero(&self) -> bool {
        matches!(self, Numeric::Finite { coefficient, .. } if coefficient.is_zero())
    }

    /// weight returns the base-10000 weight of the first nonzero digit group and that group, as Postgres stores it.
    fn weight(&self) -> (i64, u32) {
        let Numeric::Finite { coefficient, scale, .. } = self else { return (0, 0) };
        if coefficient.is_zero() {
            return (0, 0);
        }
        let digits = coefficient.to_string();
        let integer_digits = digits.len() as i64 - *scale as i64;
        // The digit groups are aligned on the decimal point, so pad the integer part to a multiple of four digits.
        let pad = (4 - integer_digits.rem_euclid(4)) % 4;
        let aligned = format!("{}{digits}", "0".repeat(pad as usize));
        let weight = (integer_digits + pad) / 4 - 1;
        let groups: Vec<u32> = aligned
            .as_bytes()
            .chunks(4)
            .map(|c| format!("{:0<4}", std::str::from_utf8(c).unwrap()).parse().unwrap())
            .collect();
        let first = groups.iter().position(|&g| g != 0).unwrap_or(0);
        (weight - first as i64, groups[first])
    }

    /// div returns the quotient at the scale Postgres chooses, failing on division by zero.
    pub fn div(&self, other: &Numeric) -> Result<Numeric> {
        if other.is_zero() {
            return Err(PgError::new(code::DIVISION_BY_ZERO, "division by zero"));
        }
        Ok(match (self, other) {
            (Numeric::NaN, _) | (_, Numeric::NaN) => Numeric::NaN,
            (Numeric::Finite { .. }, Numeric::Infinity | Numeric::NegativeInfinity) => Numeric::zero(0),
            (Numeric::Infinity | Numeric::NegativeInfinity, Numeric::Infinity | Numeric::NegativeInfinity) => {
                Numeric::NaN
            }
            (Numeric::Infinity | Numeric::NegativeInfinity, _) => {
                if self.is_negative() != other.is_negative() {
                    Numeric::NegativeInfinity
                } else {
                    Numeric::Infinity
                }
            }
            _ => {
                let (w1, d1) = self.weight();
                let (w2, d2) = other.weight();
                let mut quotient_weight = w1 - w2;
                if d1 <= d2 {
                    quotient_weight -= 1;
                }
                let scale = (MIN_SIG_DIGITS - quotient_weight * 4)
                    .max(self.scale() as i64)
                    .max(other.scale() as i64)
                    .clamp(0, MAX_DISPLAY_SCALE) as u32;
                self.div_at(other, scale)
            }
        })
    }

    /// sqrt returns the square root rounded half away from zero to the scale, or NaN for a negative value.
    pub fn sqrt(&self, scale: u32) -> Numeric {
        match self {
            Numeric::Finite { negative: false, coefficient, scale: s } => {
                // sqrt(c / 10^s) * 10^(scale + 1) = sqrt(c * 10^(2 * (scale + 1) - s))
                let exponent = 2 * (scale as i64 + 1) - *s as i64;
                let n = if exponent >= 0 {
                    coefficient * pow10(exponent as u32)
                } else {
                    coefficient / pow10((-exponent) as u32)
                };
                let (quotient, digit) = n.sqrt().div_rem(&BigUint::from(10u32));
                let rounded = if digit >= BigUint::from(5u32) { quotient + 1u32 } else { quotient };
                Numeric::finite(false, rounded, scale)
            }
            Numeric::Infinity => Numeric::Infinity,
            _ => Numeric::NaN,
        }
    }

    /// div_exact returns the value divided by 10 to the power, keeping every digit.
    pub fn div_exact(&self, power: u32) -> Numeric {
        match self {
            Numeric::Finite { negative, coefficient, scale } => {
                Numeric::Finite { negative: *negative, coefficient: coefficient.clone(), scale: scale + power }
            }
            other => other.clone(),
        }
    }

    /// div_at returns the finite quotient rounded half away from zero to the scale.
    fn div_at(&self, other: &Numeric, scale: u32) -> Numeric {
        let (
            Numeric::Finite { negative: ln, coefficient: l, scale: ls },
            Numeric::Finite { negative: rn, coefficient: r, scale: rs },
        ) = (self, other)
        else {
            return Numeric::NaN;
        };
        // l/10^ls / (r/10^rs) * 10^scale = l * 10^(scale + rs) / (r * 10^ls)
        let numerator = l * pow10(scale + rs);
        let denominator = r * pow10(*ls);
        let (quotient, remainder) = numerator.div_rem(&denominator);
        let quotient = if remainder * 2u32 >= denominator { quotient + 1u32 } else { quotient };
        Numeric::finite(ln != rn, quotient, scale)
    }

    /// rem returns the remainder of truncating division, failing on division by zero.
    pub fn rem(&self, other: &Numeric) -> Result<Numeric> {
        if other.is_zero() {
            return Err(PgError::new(code::DIVISION_BY_ZERO, "division by zero"));
        }
        Ok(match (self, other) {
            (Numeric::Finite { .. }, Numeric::Finite { .. }) => {
                let scale = self.scale().max(other.scale());
                let (ln, l) = self.signed_at(scale);
                let (_, r) = other.signed_at(scale);
                Numeric::finite(ln, l % r, scale)
            }
            (Numeric::Finite { .. }, Numeric::Infinity | Numeric::NegativeInfinity) => self.clone(),
            _ => Numeric::NaN,
        })
    }

    /// cmp_numeric orders numerics as Postgres does, with NaN above everything.
    pub fn cmp_numeric(&self, other: &Numeric) -> Ordering {
        let rank = |n: &Numeric| match n {
            Numeric::NegativeInfinity => 0,
            Numeric::Finite { .. } => 1,
            Numeric::Infinity => 2,
            Numeric::NaN => 3,
        };
        match (self, other) {
            (Numeric::Finite { .. }, Numeric::Finite { .. }) => {
                let scale = self.scale().max(other.scale());
                let (ln, l) = self.signed_at(scale);
                let (rn, r) = other.signed_at(scale);
                match (ln, rn) {
                    (false, true) => Ordering::Greater,
                    (true, false) => Ordering::Less,
                    (false, false) => l.cmp(&r),
                    (true, true) => r.cmp(&l),
                }
            }
            _ => rank(self).cmp(&rank(other)),
        }
    }

    /// to_i64 rounds the value half away from zero to an integer, or None when it is not a finite 64-bit value.
    pub fn to_i64(&self) -> Option<i64> {
        let Numeric::Finite { negative, coefficient, .. } = self.with_scale(0) else { return None };
        let magnitude = coefficient.to_i128()?;
        i64::try_from(if negative { -magnitude } else { magnitude }).ok()
    }

    /// to_f64 returns the nearest float.
    pub fn to_f64(&self) -> f64 {
        match self {
            Numeric::NaN => f64::NAN,
            Numeric::Infinity => f64::INFINITY,
            Numeric::NegativeInfinity => f64::NEG_INFINITY,
            finite => finite.to_string().parse().unwrap_or(f64::NAN),
        }
    }

    /// from_f64 converts a float as Postgres does, through its shortest exact text.
    pub fn from_f64(value: f64) -> Numeric {
        if value.is_nan() {
            return Numeric::NaN;
        }
        if value.is_infinite() {
            return if value > 0.0 { Numeric::Infinity } else { Numeric::NegativeInfinity };
        }
        Numeric::parse(&format!("{value:e}")).unwrap_or(Numeric::NaN)
    }

    /// send returns Postgres' binary format: the digit group count, weight, sign, display scale, and base-10000
    /// groups.
    pub fn send(&self) -> Vec<u8> {
        let (sign, groups, weight, scale): (u16, Vec<u16>, i16, u16) = match self {
            Numeric::NaN => (0xC000, Vec::new(), 0, 0),
            Numeric::Infinity => (0xD000, Vec::new(), 0, 0),
            Numeric::NegativeInfinity => (0xF000, Vec::new(), 0, 0),
            Numeric::Finite { negative, coefficient, scale } => {
                if coefficient.is_zero() {
                    (0, Vec::new(), 0, *scale as u16)
                } else {
                    let digits = coefficient.to_string();
                    let integer_digits = digits.len() as i64 - *scale as i64;
                    let pad_left = (4 - integer_digits.rem_euclid(4)) % 4;
                    let mut aligned = format!("{}{digits}", "0".repeat(pad_left as usize));
                    while aligned.len() % 4 != 0 {
                        aligned.push('0');
                    }
                    let mut groups: Vec<u16> = aligned
                        .as_bytes()
                        .chunks(4)
                        .map(|c| std::str::from_utf8(c).unwrap().parse().unwrap())
                        .collect();
                    let mut weight = (integer_digits + pad_left) / 4 - 1;
                    while groups.first() == Some(&0) {
                        groups.remove(0);
                        weight -= 1;
                    }
                    while groups.last() == Some(&0) {
                        groups.pop();
                    }
                    (if *negative { 0x4000 } else { 0 }, groups, weight as i16, *scale as u16)
                }
            }
        };
        let mut out = Vec::with_capacity(8 + groups.len() * 2);
        out.extend_from_slice(&(groups.len() as u16).to_be_bytes());
        out.extend_from_slice(&weight.to_be_bytes());
        out.extend_from_slice(&sign.to_be_bytes());
        out.extend_from_slice(&scale.to_be_bytes());
        for group in groups {
            out.extend_from_slice(&group.to_be_bytes());
        }
        out
    }

    /// receive reads Postgres' binary format.
    pub fn receive(bytes: &[u8]) -> Option<Numeric> {
        let word = |i: usize| bytes.get(i..i + 2).map(|b| u16::from_be_bytes([b[0], b[1]]));
        let (count, weight, sign, scale) = (word(0)?, word(2)? as i16, word(4)?, word(6)?);
        match sign {
            0xC000 => return Some(Numeric::NaN),
            0xD000 => return Some(Numeric::Infinity),
            0xF000 => return Some(Numeric::NegativeInfinity),
            0 | 0x4000 => {}
            _ => return None,
        }
        let mut coefficient = BigUint::zero();
        for i in 0..count as usize {
            coefficient = coefficient * 10000u32 + word(8 + i * 2)?;
        }
        // The groups hold the digits from 10000^weight down to 10000^(weight - count + 1).
        let exponent = (weight as i64 - count as i64 + 1) * 4;
        let value = if exponent >= 0 {
            Numeric::Finite { negative: sign == 0x4000, coefficient: coefficient * pow10(exponent as u32), scale: 0 }
        } else {
            Numeric::Finite { negative: sign == 0x4000, coefficient, scale: (-exponent) as u32 }
        };
        Some(value.with_scale(scale as u32))
    }

    /// encode returns the value in Dolt's decimal encoding: a little-endian exponent, a sign byte, and the
    /// big-endian coefficient padded to whole 64-bit words, or a sentinel for NaN and the infinities.
    pub fn encode(&self) -> Vec<u8> {
        match self {
            Numeric::NaN => 0xc000i32.to_le_bytes().to_vec(),
            Numeric::Infinity => 0xd000i32.to_le_bytes().to_vec(),
            Numeric::NegativeInfinity => 0xf000i32.to_le_bytes().to_vec(),
            Numeric::Finite { negative, coefficient, scale } => {
                let mut out = (-(*scale as i32)).to_le_bytes().to_vec();
                let sign: i8 = if coefficient.is_zero() {
                    0
                } else if *negative {
                    -1
                } else {
                    1
                };
                out.push(sign as u8);
                if !coefficient.is_zero() {
                    let bytes = coefficient.to_bytes_be();
                    let padded = bytes.len().div_ceil(8) * 8;
                    out.extend(std::iter::repeat_n(0, padded - bytes.len()));
                    out.extend_from_slice(&bytes);
                }
                out
            }
        }
    }

    /// decode reads Dolt's decimal encoding.
    pub fn decode(bytes: &[u8]) -> Option<Numeric> {
        let exponent = i32::from_le_bytes(bytes.get(..4)?.try_into().ok()?);
        if bytes.len() == 4 {
            return match exponent {
                0xc000 => Some(Numeric::NaN),
                0xd000 => Some(Numeric::Infinity),
                0xf000 => Some(Numeric::NegativeInfinity),
                _ => None,
            };
        }
        let negative = (*bytes.get(4)? as i8) < 0;
        let coefficient = BigUint::from_bytes_be(&bytes[5..]);
        Some(if exponent >= 0 {
            Numeric::finite(negative, coefficient * pow10(exponent as u32), 0)
        } else {
            Numeric::finite(negative, coefficient, (-exponent) as u32)
        })
    }
}

impl std::fmt::Display for Numeric {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Numeric::NaN => write!(f, "NaN"),
            Numeric::Infinity => write!(f, "Infinity"),
            Numeric::NegativeInfinity => write!(f, "-Infinity"),
            Numeric::Finite { negative, coefficient, scale } => {
                let digits = coefficient.to_string();
                let scale = *scale as usize;
                let digits = if digits.len() <= scale {
                    format!("{}{digits}", "0".repeat(scale + 1 - digits.len()))
                } else {
                    digits
                };
                let (whole, fraction) = digits.split_at(digits.len() - scale);
                let sign = if *negative { "-" } else { "" };
                if scale == 0 { write!(f, "{sign}{whole}") } else { write!(f, "{sign}{whole}.{fraction}") }
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    /// n parses a numeric.
    fn n(text: &str) -> Numeric {
        Numeric::parse(text).unwrap()
    }

    #[test]
    fn numerics_print_with_their_scale() {
        assert_eq!(n("1.50").to_string(), "1.50");
        assert_eq!(n("-0.001").to_string(), "-0.001");
        assert_eq!(n("1e3").to_string(), "1000");
        assert_eq!(n("1.5e-3").to_string(), "0.0015");
        assert_eq!(n(" -0 ").to_string(), "0");
        assert_eq!(n("NaN").to_string(), "NaN");
        assert!(Numeric::parse("1.2.3").is_err());
    }

    #[test]
    fn arithmetic_follows_postgres_scales() {
        assert_eq!(n("1.5").add(&n("2.25")).to_string(), "3.75");
        assert_eq!(n("1.5").mul(&n("2.25")).to_string(), "3.375");
        assert_eq!(n("1").div(&n("3")).unwrap().to_string(), "0.33333333333333333333");
        assert_eq!(n("10").div(&n("4")).unwrap().to_string(), "2.5000000000000000");
        assert_eq!(n("100000").div(&n("3")).unwrap().to_string(), "33333.333333333333");
        assert_eq!(n("7.5").rem(&n("2")).unwrap().to_string(), "1.5");
        assert_eq!(n("12345678901234567890").div(&n("7.0")).unwrap().to_string(), "1763668414462081127.1");
        assert_eq!(n("2.5").to_i64(), Some(3));
        assert_eq!(n("-2.5").to_i64(), Some(-3));
    }

    #[test]
    fn typmods_round_and_overflow() {
        let typmod = ((5 << 16) | 2) + 4;
        assert_eq!(n("123.456").apply_typmod(typmod).unwrap().to_string(), "123.46");
        let err = n("1234.5").apply_typmod(typmod).unwrap_err();
        assert_eq!(
            err.detail.unwrap(),
            "A field with precision 5, scale 2 must round to an absolute value less than 10^3."
        );
    }

    #[test]
    fn formats_round_trip() {
        for text in ["0", "1.50", "-12345.6789", "0.0001", "100000000", "NaN", "-Infinity"] {
            let value = n(text);
            assert_eq!(Numeric::receive(&value.send()).unwrap(), value, "{text}");
            assert_eq!(Numeric::decode(&value.encode()).unwrap(), value, "{text}");
        }
        assert_eq!(n("1.5").encode(), [0xff, 0xff, 0xff, 0xff, 1, 0, 0, 0, 0, 0, 0, 0, 15]);
    }
}
