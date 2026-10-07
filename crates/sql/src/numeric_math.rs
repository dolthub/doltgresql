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

//! The square roots, logarithms, exponentials, and powers of numerics, with the result scales that Postgres 15's
//! numeric.c chooses, computed with extra digits and rounded half away from zero.

use std::cmp::Ordering;

use num_bigint::{BigInt, BigUint, Sign};
use num_integer::Integer;
use num_traits::{Signed, ToPrimitive, Zero};

use crate::error::{PgError, Result, code};
use crate::numeric::Numeric;

/// MIN_SIG_DIGITS is the least number of significant digits a result has, as Postgres' NUMERIC_MIN_SIG_DIGITS.
const MIN_SIG_DIGITS: i64 = 16;

/// MAX_DISPLAY_SCALE is the largest scale a result has, as Postgres' NUMERIC_MAX_DISPLAY_SCALE.
const MAX_DISPLAY_SCALE: i64 = 1000;

/// MAX_RESULT_SCALE bounds the exponent of an exponential, as Postgres' NUMERIC_MAX_RESULT_SCALE.
const MAX_RESULT_SCALE: f64 = 2000.0;

/// DEC_DIGITS is the number of decimal digits in each of Postgres' base-10000 digits.
const DEC_DIGITS: i64 = 4;

/// RESULT_WEIGHT_FACTOR is the approximation of log10(e) that Postgres' power_var estimates a result's decimal weight
/// with, which decides the result's scale.
#[allow(clippy::approx_constant)]
const RESULT_WEIGHT_FACTOR: f64 = 0.434294481903252;

/// GUARD is how many digits beyond those a result needs the computations keep.
const GUARD: i64 = 12;

/// Var is a finite numeric as Postgres' NumericVar holds it: base-10000 digits without trailing zero digits, the
/// weight of the first, and the display scale.
struct Var {
    negative: bool,
    weight: i64,
    digits: Vec<u32>,
    dscale: i64,
}

/// var returns a finite numeric as a Var, or None for NaN and the infinities.
fn var(n: &Numeric) -> Option<Var> {
    let Numeric::Finite { negative, coefficient, scale } = n else { return None };
    let dscale = *scale as i64;
    if coefficient.is_zero() {
        return Some(Var { negative: false, weight: 0, digits: Vec::new(), dscale });
    }
    let pad = (DEC_DIGITS - dscale % DEC_DIGITS) % DEC_DIGITS;
    let text = (coefficient * pow10(pad)).to_string();
    let text = format!("{}{text}", "0".repeat(((DEC_DIGITS - text.len() as i64 % DEC_DIGITS) % DEC_DIGITS) as usize));
    let mut digits: Vec<u32> = text
        .as_bytes()
        .chunks(DEC_DIGITS as usize)
        .map(|c| std::str::from_utf8(c).unwrap_or("0").parse().unwrap_or(0))
        .collect();
    let weight = digits.len() as i64 - 1 - (dscale + pad) / DEC_DIGITS;
    while digits.last() == Some(&0) {
        digits.pop();
    }
    Some(Var { negative: *negative, weight, digits, dscale })
}

/// pow10 returns 10 to a power.
fn pow10(power: i64) -> BigUint {
    BigUint::from(10u32).pow(power.max(0) as u32)
}

/// signed returns an unsigned integer with a sign.
fn signed(negative: bool, magnitude: BigUint) -> BigInt {
    BigInt::from_biguint(if negative { Sign::Minus } else { Sign::Plus }, magnitude)
}

/// parts returns the sign, coefficient, and scale of a finite numeric.
fn parts(n: &Numeric) -> (bool, BigUint, i64) {
    match n {
        Numeric::Finite { negative, coefficient, scale } => (*negative, coefficient.clone(), *scale as i64),
        _ => (false, BigUint::zero(), 0),
    }
}

/// fixed returns a finite numeric times 10 to the scale, truncated.
fn fixed(n: &Numeric, scale: i64) -> BigInt {
    let (negative, coefficient, s) = parts(n);
    let magnitude = if scale >= s { coefficient * pow10(scale - s) } else { coefficient / pow10(s - scale) };
    signed(negative, magnitude)
}

/// rounded returns a value given times 10 to the `from` scale as a numeric of the `to` scale, rounded half away from
/// zero.
fn rounded(value: &BigInt, from: i64, to: i64) -> Numeric {
    let negative = value.is_negative();
    let magnitude = value.magnitude().clone();
    if from <= to {
        return Numeric::finite(negative, magnitude * pow10(to - from), to as u32);
    }
    ratio(negative, magnitude, pow10(from), to)
}

/// ratio returns a fraction as a numeric of the scale, rounded half away from zero.
fn ratio(negative: bool, numerator: BigUint, denominator: BigUint, scale: i64) -> Numeric {
    let (quotient, remainder) = (numerator * pow10(scale)).div_rem(&denominator);
    let quotient = if remainder * 2u32 >= denominator { quotient + 1u32 } else { quotient };
    Numeric::finite(negative, quotient, scale as u32)
}

/// to_float returns a value given times 10 to the scale as a float.
fn to_float(value: &BigInt, scale: i64) -> f64 {
    let shift = (scale - 17).max(0);
    (value / BigInt::from(pow10(shift))).to_f64().unwrap_or(f64::INFINITY) / 10f64.powi((scale - shift) as i32)
}

/// clamp_scale bounds a result scale to the display scales.
fn clamp_scale(scale: i64) -> i64 {
    scale.clamp(0, MAX_DISPLAY_SCALE)
}

/// overflow returns Postgres' error for a result too large for a numeric.
fn overflow() -> PgError {
    PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "value overflows numeric format")
}

/// log_error returns Postgres' error for the logarithm of zero or of a negative number.
fn log_error(zero: bool) -> PgError {
    let message = if zero { "cannot take logarithm of zero" } else { "cannot take logarithm of a negative number" };
    PgError::new(code::INVALID_ARGUMENT_FOR_LOG, message)
}

/// complex_error returns Postgres' error for raising a negative number to a fractional power.
fn complex_error() -> PgError {
    PgError::new(
        code::INVALID_ARGUMENT_FOR_POWER,
        "a negative number raised to a non-integer power yields a complex result",
    )
}

/// digit_count returns the number of decimal digits of an integer.
fn digit_count(n: &BigUint) -> i64 {
    if n.is_zero() { 1 } else { n.to_string().len() as i64 }
}

/// is_integral reports whether a Var holds an integer.
fn is_integral(v: &Var) -> bool {
    v.digits.is_empty() || v.digits.len() as i64 <= v.weight + 1
}

/// is_odd reports whether a Var holds an odd integer.
fn is_odd(v: &Var) -> bool {
    !v.digits.is_empty() && v.digits.len() as i64 == v.weight + 1 && v.digits[v.digits.len() - 1] & 1 == 1
}

/// atanh_inverse returns atanh(1/q) times 10 to the scale.
fn atanh_inverse(q: u32, scale: i64) -> BigInt {
    let squared = BigInt::from(q) * q;
    let mut power = BigInt::from(pow10(scale)) / q;
    let mut sum = power.clone();
    let mut n = 1u32;
    loop {
        power /= &squared;
        if power.is_zero() {
            return sum;
        }
        sum += &power / (2 * n + 1);
        n += 1;
    }
}

/// ln10 returns ln(10) times 10 to the scale, from ln(10) = 6 atanh(1/3) + 2 atanh(1/9).
fn ln10(scale: i64) -> BigInt {
    let g = scale + 5;
    (atanh_inverse(3, g) * 6 + atanh_inverse(9, g) * 2) / BigInt::from(pow10(5))
}

/// ln_fixed returns the natural logarithm of a positive finite numeric times 10 to the scale, within a few units of
/// its last digit.
fn ln_fixed(x: &Numeric, scale: i64) -> BigInt {
    const HALVINGS: u32 = 6;
    let (_, coefficient, s) = parts(x);
    let digits = digit_count(&coefficient);
    let exponent = digits - 1 - s;
    let g = scale + GUARD + digit_count(&BigUint::from(exponent.unsigned_abs()));
    let one = BigInt::from(pow10(g));
    let mut m = BigInt::from(coefficient * pow10(g)) / BigInt::from(pow10(digits - 1));
    for _ in 0..HALVINGS {
        m = (m * &one).sqrt();
    }
    let z = (&m - &one) * &one / (&m + &one);
    let z2 = &z * &z / &one;
    let (mut term, mut sum, mut n) = (z.clone(), z, 1u32);
    loop {
        term = term * &z2 / &one;
        if term.is_zero() {
            break;
        }
        sum += &term / (2 * n + 1);
        n += 1;
    }
    let ln = sum * (2u32 << HALVINGS) + ln10(g) * exponent;
    ln / BigInt::from(pow10(g - scale))
}

/// exp_fixed returns e raised to a value given times 10 to the scale `xs`, times 10 to the scale, within a few units
/// of its last digit when the value has enough digits.
fn exp_fixed(x: &BigInt, xs: i64, scale: i64) -> BigInt {
    let negative = x.is_negative();
    let magnitude = x.magnitude();
    let whole = magnitude / pow10(xs);
    let halvings = whole.bits() as i64 + 7;
    let weight = (whole.to_f64().unwrap_or(0.0) * std::f64::consts::LOG10_E).ceil() as i64 + 1;
    let g = scale + weight + halvings / 3 + GUARD;
    let one = BigInt::from(pow10(g));
    let r = if g >= xs { BigInt::from(magnitude * pow10(g - xs)) } else { BigInt::from(magnitude / pow10(xs - g)) };
    let r = r >> halvings as usize;
    let (mut term, mut sum, mut n) = (one.clone(), one.clone(), 1u32);
    loop {
        term = term * &r / &one / n;
        if term.is_zero() {
            break;
        }
        sum += &term;
        n += 1;
    }
    for _ in 0..halvings {
        sum = &sum * &sum / &one;
    }
    if negative {
        sum = &one * &one / sum;
    }
    sum / BigInt::from(pow10(g - scale))
}

/// estimate_ln_dweight estimates the decimal weight of a value's natural logarithm, as Postgres'
/// estimate_ln_dweight does.
fn estimate_ln_dweight(x: &Numeric) -> i64 {
    let Some(v) = var(x) else { return 0 };
    if v.negative {
        return 0;
    }
    let (low, high) = (Numeric::parse("0.9").unwrap_or(Numeric::NaN), Numeric::parse("1.1").unwrap_or(Numeric::NaN));
    if x.cmp_numeric(&low) != Ordering::Less && x.cmp_numeric(&high) != Ordering::Greater {
        let difference = x.sub(&Numeric::from_i64(1));
        return match var(&difference) {
            Some(d) if !d.digits.is_empty() => d.weight * DEC_DIGITS + (d.digits[0] as f64).log10() as i64,
            _ => 0,
        };
    }
    if v.digits.is_empty() {
        return 0;
    }
    let mut digits = v.digits[0] as f64;
    let mut dweight = v.weight * DEC_DIGITS;
    if v.digits.len() > 1 {
        digits = digits * 10000.0 + v.digits[1] as f64;
        dweight -= DEC_DIGITS;
    }
    let ln = digits.ln() + dweight as f64 * std::f64::consts::LN_10;
    ln.abs().log10() as i64
}

/// sqrt returns the square root of a numeric, as Postgres' numeric_sqrt does.
pub fn sqrt(x: &Numeric) -> Result<Numeric> {
    let negative = || PgError::new(code::INVALID_ARGUMENT_FOR_POWER, "cannot take square root of a negative number");
    match x {
        Numeric::NaN | Numeric::Infinity => return Ok(x.clone()),
        Numeric::NegativeInfinity => return Err(negative()),
        _ if x.is_negative() => return Err(negative()),
        _ => {}
    }
    let v = var(x).ok_or_else(negative)?;
    let sweight = (v.weight + 1) * DEC_DIGITS / 2 - 1;
    let rscale = clamp_scale((MIN_SIG_DIGITS - sweight).max(v.dscale));
    Ok(x.sqrt(rscale as u32))
}

/// ln returns the natural logarithm of a numeric, as Postgres' numeric_ln does.
pub fn ln(x: &Numeric) -> Result<Numeric> {
    match x {
        Numeric::NaN | Numeric::Infinity => return Ok(x.clone()),
        Numeric::NegativeInfinity => return Err(log_error(false)),
        _ if x.is_zero() => return Err(log_error(true)),
        _ if x.is_negative() => return Err(log_error(false)),
        _ => {}
    }
    let dscale = parts(x).2;
    let rscale = clamp_scale((MIN_SIG_DIGITS - estimate_ln_dweight(x)).max(dscale));
    Ok(rounded(&ln_fixed(x, rscale + GUARD), rscale + GUARD, rscale))
}

/// log returns the logarithm of a numeric to a base, as Postgres' numeric_log does.
pub fn log(base: &Numeric, x: &Numeric) -> Result<Numeric> {
    if !matches!(base, Numeric::Finite { .. }) || !matches!(x, Numeric::Finite { .. }) {
        if matches!(base, Numeric::NaN) || matches!(x, Numeric::NaN) {
            return Ok(Numeric::NaN);
        }
        if base.is_negative() || x.is_negative() {
            return Err(log_error(false));
        }
        if base.is_zero() || x.is_zero() {
            return Err(log_error(true));
        }
        return Ok(match (base, x) {
            (Numeric::Infinity, Numeric::Infinity) => Numeric::NaN,
            (Numeric::Infinity, _) => Numeric::zero(0),
            _ => Numeric::Infinity,
        });
    }
    for n in [base, x] {
        if n.is_zero() {
            return Err(log_error(true));
        }
        if n.is_negative() {
            return Err(log_error(false));
        }
    }
    let (ln_base_dweight, ln_x_dweight) = (estimate_ln_dweight(base), estimate_ln_dweight(x));
    let result_dweight = ln_x_dweight - ln_base_dweight;
    let rscale = clamp_scale((MIN_SIG_DIGITS - result_dweight).max(parts(base).2).max(parts(x).2));
    let base_scale = (rscale + result_dweight - ln_base_dweight + GUARD).max(0) + GUARD;
    let x_scale = (rscale + result_dweight - ln_x_dweight + GUARD).max(0) + GUARD;
    let ln_base = ln_fixed(base, base_scale);
    if ln_base.is_zero() {
        return Err(PgError::new(code::DIVISION_BY_ZERO, "division by zero"));
    }
    let ln_x = ln_fixed(x, x_scale);
    let numerator = ln_x.magnitude() * pow10(base_scale);
    let denominator = ln_base.magnitude() * pow10(x_scale);
    Ok(ratio(ln_x.is_negative() != ln_base.is_negative(), numerator, denominator, rscale))
}

/// exp returns e raised to a numeric, as Postgres' numeric_exp does.
pub fn exp(x: &Numeric) -> Result<Numeric> {
    match x {
        Numeric::NaN | Numeric::Infinity => return Ok(x.clone()),
        Numeric::NegativeInfinity => return Ok(Numeric::zero(0)),
        _ => {}
    }
    let val = (x.to_f64() * std::f64::consts::LOG10_E).clamp(-MAX_RESULT_SCALE, MAX_RESULT_SCALE);
    let rscale = clamp_scale((MIN_SIG_DIGITS - val as i64).max(parts(x).2));
    let xs = parts(x).2.max(rscale + (val as i64).max(0) + GUARD);
    exp_scaled(&fixed(x, xs), xs, rscale)
}

/// exp_scaled returns e raised to a value given times 10 to the scale `xs` as a numeric of the scale, failing as
/// Postgres' exp_var does when it overflows.
fn exp_scaled(x: &BigInt, xs: i64, rscale: i64) -> Result<Numeric> {
    let val = to_float(x, xs);
    if val.abs() >= MAX_RESULT_SCALE * 3.0 {
        if val > 0.0 {
            return Err(overflow());
        }
        return Ok(Numeric::zero(rscale as u32));
    }
    Ok(rounded(&exp_fixed(x, xs, rscale + GUARD), rscale + GUARD, rscale))
}

/// power returns a numeric raised to a numeric power, as Postgres' numeric_power does.
pub fn power(base: &Numeric, exponent: &Numeric) -> Result<Numeric> {
    let one = Numeric::from_i64(1);
    let sign = |n: &Numeric| {
        if n.is_zero() {
            0
        } else if n.is_negative() {
            -1
        } else {
            1
        }
    };
    let zero_error = || PgError::new(code::INVALID_ARGUMENT_FOR_POWER, "zero raised to a negative power is undefined");
    let finite = |n: &Numeric| matches!(n, Numeric::Finite { .. });
    if !finite(base) || !finite(exponent) {
        if matches!(base, Numeric::NaN) {
            return Ok(if finite(exponent) && exponent.is_zero() { one } else { Numeric::NaN });
        }
        if matches!(exponent, Numeric::NaN) {
            return Ok(if finite(base) && base.cmp_numeric(&one) == Ordering::Equal { one } else { Numeric::NaN });
        }
        let (sign_base, sign_exponent) = (sign(base), sign(exponent));
        let integral = !finite(exponent) || var(exponent).is_some_and(|v| is_integral(&v));
        if sign_base == 0 && sign_exponent < 0 {
            return Err(zero_error());
        }
        if sign_base < 0 && !integral {
            return Err(complex_error());
        }
        if finite(base) && base.cmp_numeric(&one) == Ordering::Equal {
            return Ok(one);
        }
        if sign_exponent == 0 {
            return Ok(one);
        }
        if sign_base == 0 && sign_exponent > 0 {
            return Ok(Numeric::zero(0));
        }
        if !finite(exponent) {
            let above_one = match base {
                Numeric::Finite { .. } => {
                    if base.cmp_numeric(&Numeric::from_i64(-1)) == Ordering::Equal {
                        return Ok(one);
                    }
                    let absolute = if base.is_negative() { base.negate() } else { base.clone() };
                    absolute.cmp_numeric(&one) == Ordering::Greater
                }
                _ => true,
            };
            return Ok(if above_one == (sign_exponent > 0) { Numeric::Infinity } else { Numeric::zero(0) });
        }
        if matches!(base, Numeric::Infinity) {
            return Ok(if sign_exponent > 0 { Numeric::Infinity } else { Numeric::zero(0) });
        }
        if sign_exponent < 0 {
            return Ok(Numeric::zero(0));
        }
        let odd = var(exponent).is_some_and(|v| is_odd(&v));
        return Ok(if odd { Numeric::NegativeInfinity } else { Numeric::Infinity });
    }
    if base.is_zero() && exponent.is_negative() {
        return Err(zero_error());
    }
    power_var(base, exponent)
}

/// power_var raises a finite numeric to a finite power, as Postgres' power_var does.
fn power_var(base: &Numeric, exponent: &Numeric) -> Result<Numeric> {
    let (Some(b), Some(e)) = (var(base), var(exponent)) else { return Ok(Numeric::NaN) };
    if is_integral(&e)
        && let Some(n) = exponent.to_i64()
        && let Ok(n) = i32::try_from(n)
    {
        let rscale = clamp_scale(MIN_SIG_DIGITS.max(b.dscale));
        return power_var_int(base, &b, n, rscale);
    }
    if base.is_zero() {
        return Ok(Numeric::zero(MIN_SIG_DIGITS as u32));
    }
    let mut negative = false;
    let mut absolute = base.clone();
    if b.negative {
        if !is_integral(&e) {
            return Err(complex_error());
        }
        negative = is_odd(&e);
        absolute = base.negate();
    }
    let ln_dweight = estimate_ln_dweight(&absolute);
    let (exponent_negative, exponent_coefficient, exponent_scale) = parts(exponent);
    let estimate_scale = (8 - ln_dweight).max(0) + GUARD;
    let estimate = ln_fixed(&absolute, estimate_scale) * signed(exponent_negative, exponent_coefficient.clone());
    let val = to_float(&estimate, estimate_scale + exponent_scale);
    if val.abs() > MAX_RESULT_SCALE * 3.01 {
        if val > 0.0 {
            return Err(overflow());
        }
        return Ok(Numeric::zero(MAX_DISPLAY_SCALE as u32));
    }
    let val = val * RESULT_WEIGHT_FACTOR;
    let rscale = clamp_scale((MIN_SIG_DIGITS - val as i64).max(b.dscale).max(e.dscale));
    let xs = rscale + (val as i64).max(0) + GUARD;
    let whole = digit_count(&(&exponent_coefficient / pow10(exponent_scale)));
    let ln_scale = xs + whole + 2;
    let x = ln_fixed(&absolute, ln_scale) * signed(exponent_negative, exponent_coefficient)
        / BigInt::from(pow10(exponent_scale + ln_scale - xs));
    let result = exp_scaled(&x, xs, rscale)?;
    Ok(if negative { result.negate() } else { result })
}

/// power_var_int raises a finite numeric to an integer power at a result scale, as Postgres 15's power_var_int does.
fn power_var_int(base: &Numeric, b: &Var, n: i32, rscale: i64) -> Result<Numeric> {
    let (negative, coefficient, scale) = parts(base);
    match n {
        0 => return Ok(Numeric::finite(false, pow10(rscale), rscale as u32)),
        1 => return Ok(rounded(&signed(negative, coefficient), scale, rscale)),
        -1 if coefficient.is_zero() => return Err(PgError::new(code::DIVISION_BY_ZERO, "division by zero")),
        -1 => return Ok(ratio(negative, pow10(scale), coefficient, rscale)),
        2 => return Ok(rounded(&BigInt::from(&coefficient * &coefficient), scale * 2, rscale)),
        _ => {}
    }
    if coefficient.is_zero() {
        if n < 0 {
            return Err(PgError::new(code::DIVISION_BY_ZERO, "division by zero"));
        }
        return Ok(Numeric::zero(rscale as u32));
    }
    let mut f = b.digits[0] as f64;
    let mut p = b.weight * DEC_DIGITS;
    for (i, &digit) in b.digits.iter().enumerate().skip(1) {
        if i as i64 * DEC_DIGITS >= 16 {
            break;
        }
        f = f * 10000.0 + digit as f64;
        p -= DEC_DIGITS;
    }
    let f = n as f64 * (f.log10() + p as f64);
    if f > (3 * i16::MAX as i64 * DEC_DIGITS) as f64 {
        return Err(overflow());
    }
    if f + 1.0 < -(rscale as f64) || f + 1.0 < -(MAX_DISPLAY_SCALE as f64) {
        return Ok(Numeric::zero(rscale as u32));
    }
    let digits = 1 + rscale + f as i64 + (n.unsigned_abs() as f64).ln() as i64 + 8 + GUARD;
    let truncate = |mantissa: BigUint, exponent: i64| {
        let extra = digit_count(&mantissa) - digits;
        if extra > 0 { (mantissa / pow10(extra), exponent + extra) } else { (mantissa, exponent) }
    };
    let (mut square, mut product) = ((coefficient.clone(), -scale), (BigUint::from(1u32), 0i64));
    let mut mask = n.unsigned_abs();
    let limit = i16::MAX as i64 * DEC_DIGITS;
    loop {
        if mask & 1 == 1 {
            product = truncate(&product.0 * &square.0, product.1 + square.1);
        }
        mask >>= 1;
        if mask == 0 {
            break;
        }
        square = truncate(&square.0 * &square.0, square.1 * 2);
        if digit_count(&square.0) + square.1 > limit || digit_count(&product.0) + product.1 > limit {
            if n > 0 {
                return Err(overflow());
            }
            return Ok(Numeric::zero(rscale as u32));
        }
    }
    let odd = negative && n % 2 != 0;
    let (mantissa, exponent) = product;
    Ok(if n < 0 {
        if exponent >= 0 {
            ratio(odd, BigUint::from(1u32), mantissa * pow10(exponent), rscale)
        } else {
            ratio(odd, pow10(-exponent), mantissa, rscale)
        }
    } else {
        rounded(&signed(odd, mantissa), -exponent, rscale)
    })
}
