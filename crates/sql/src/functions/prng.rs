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

//! The pseudorandom functions, ported from Postgres' pseudorandomfuncs.c with the xoroshiro128** generator of
//! pg_prng.c, so that a session that calls setseed draws the same values Postgres draws.

use num_bigint::{BigInt, BigUint, Sign};
use num_traits::{ToPrimitive, Zero};

use super::Function;
use crate::error::{PgError, Result, code};
use crate::numeric::Numeric;
use crate::oid::{FLOAT8, INT4, INT8, NUMERIC};
use crate::query::Ctx;
use crate::types::Value;

/// NBASE is the base of a Postgres numeric's digits.
const NBASE: u64 = 10_000;

/// Prng is the state of a session's generator.
#[derive(Clone, Copy, Debug)]
pub struct Prng {
    s0: u64,
    s1: u64,
}

impl Prng {
    /// strong returns a generator seeded from the system's random bits.
    fn strong() -> Prng {
        let mut prng = Prng { s0: rand::random(), s1: rand::random() };
        prng.check();
        prng
    }

    /// seed seeds a generator from a 64-bit value, as pg_prng_seed does.
    fn seed(seed: u64) -> Prng {
        let mut state = seed;
        let mut prng = Prng { s0: splitmix64(&mut state), s1: splitmix64(&mut state) };
        prng.check();
        prng
    }

    /// check replaces an all-zero state, which the generator cannot leave, as pg_prng_seed_check does.
    fn check(&mut self) {
        if self.s0 == 0 && self.s1 == 0 {
            self.s0 = 0x5851_F42D_4C95_7F2D;
            self.s1 = 0x1405_7B7E_F767_814F;
        }
    }

    /// next returns the generator's next value, as xoroshiro128ss does.
    fn next(&mut self) -> u64 {
        let (s0, sx) = (self.s0, self.s1 ^ self.s0);
        let value = s0.wrapping_mul(5).rotate_left(7).wrapping_mul(9);
        self.s0 = s0.rotate_left(24) ^ sx ^ (sx << 16);
        self.s1 = sx.rotate_left(37);
        value
    }

    /// uint64_range returns a value in [min, max] by bitmask rejection, as pg_prng_uint64_range does.
    fn uint64_range(&mut self, min: u64, max: u64) -> u64 {
        if max <= min {
            return min;
        }
        let range = max - min;
        let shift = range.leading_zeros();
        loop {
            let value = self.next() >> shift;
            if value <= range {
                return min + value;
            }
        }
    }

    /// int64_range returns a value in [min, max], as pg_prng_int64_range does.
    fn int64_range(&mut self, min: i64, max: i64) -> i64 {
        if max <= min {
            return min;
        }
        (min as u64).wrapping_add(self.uint64_range(0, (max as u64).wrapping_sub(min as u64))) as i64
    }

    /// double returns a value in [0, 1), as pg_prng_double does.
    fn double(&mut self) -> f64 {
        (self.next() >> 12) as f64 * 2f64.powi(-52)
    }

    /// double_normal returns a value of the standard normal distribution by the Box-Muller transform, as
    /// pg_prng_double_normal does.
    fn double_normal(&mut self) -> f64 {
        let u1 = 1.0 - self.double();
        let u2 = 1.0 - self.double();
        (-2.0 * u1.ln()).sqrt() * (2.0 * std::f64::consts::PI * u2).sin()
    }
}

/// splitmix64 returns the next value of a splitmix64 generator, which seeds xoroshiro128**.
fn splitmix64(state: &mut u64) -> u64 {
    *state = state.wrapping_add(0x9E37_79B9_7F4A_7C15);
    let mut value = *state;
    value = (value ^ (value >> 30)).wrapping_mul(0xBF58_476D_1CE4_E5B9);
    value = (value ^ (value >> 27)).wrapping_mul(0x94D0_49BB_1331_11EB);
    value ^ (value >> 31)
}

/// FUNCTIONS are the pseudorandom functions.
pub const FUNCTIONS: &[Function] = &[
    f("setseed", &[FLOAT8], crate::routines::VOID, setseed),
    f("random", &[], FLOAT8, random),
    f("random", &[INT4, INT4], INT4, random_int),
    f("random", &[INT8, INT8], INT8, random_int),
    f("random", &[NUMERIC, NUMERIC], NUMERIC, random_numeric),
    f("random_normal", &[], FLOAT8, random_normal),
    f("random_normal", &[FLOAT8], FLOAT8, random_normal),
    f("random_normal", &[FLOAT8, FLOAT8], FLOAT8, random_normal),
];

/// f declares a pseudorandom function, which is volatile and not strict where its arguments have defaults.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// generator returns the session's generator, seeding it from the system the first time.
fn generator<'c>(ctx: &'c mut Ctx<'_>) -> &'c mut Prng {
    ctx.session.prng.get_or_insert_with(Prng::strong)
}

/// invalid returns an invalid parameter value error.
fn invalid(message: impl Into<String>) -> PgError {
    PgError::new(code::INVALID_PARAMETER_VALUE, message)
}

/// setseed seeds the session's generator from a value in [-1, 1], as pg_prng_fseed does.
fn setseed(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let Value::Float8(seed) = args[0] else { return Err(PgError::internal("setseed without a float8")) };
    if !(-1.0..=1.0).contains(&seed) {
        return Err(invalid(format!("setseed parameter {} is out of allowed range [-1,1]", c_g(seed))));
    }
    ctx.session.prng = Some(Prng::seed((((1u64 << 52) - 1) as f64 * seed) as i64 as u64));
    Ok(Value::Text(String::new()))
}

/// c_g writes a float as C's `%g` does: six significant digits, in exponent form for a decimal exponent below -4 or
/// at least 6, without trailing zeros.
fn c_g(value: f64) -> String {
    if value.is_nan() {
        return "nan".into();
    }
    if value.is_infinite() {
        return if value < 0.0 { "-inf".into() } else { "inf".into() };
    }
    let scientific = format!("{value:.5e}");
    let (mantissa, exponent) = scientific.split_once('e').unwrap_or((&scientific, "0"));
    let exponent: i32 = exponent.parse().unwrap_or_default();
    let trim = |text: String| match text.contains('.') {
        true => text.trim_end_matches('0').trim_end_matches('.').to_string(),
        false => text,
    };
    if !(-4..6).contains(&exponent) {
        let sign = if exponent < 0 { '-' } else { '+' };
        return format!("{}e{sign}{:02}", trim(mantissa.to_string()), exponent.abs());
    }
    trim(format!("{value:.*}", (5 - exponent).max(0) as usize))
}

/// random returns a value in [0, 1).
fn random(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Float8(generator(ctx).double()))
}

/// random_normal returns a value of the normal distribution with a mean and standard deviation, 0 and 1 by default.
fn random_normal(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let float = |at: usize, default: f64| match args.get(at) {
        Some(Value::Float8(v)) => *v,
        _ => default,
    };
    let (mean, stddev) = (float(0, 0.0), float(1, 1.0));
    Ok(Value::Float8(stddev * generator(ctx).double_normal() + mean))
}

/// random_int returns an integer in [min, max].
fn random_int(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (min, max) = match (&args[0], &args[1]) {
        (Value::Int4(a), Value::Int4(b)) => (i64::from(*a), i64::from(*b)),
        (Value::Int8(a), Value::Int8(b)) => (*a, *b),
        _ => return Err(PgError::internal("random without integers")),
    };
    if min > max {
        return Err(invalid("lower bound must be less than or equal to upper bound"));
    }
    let value = generator(ctx).int64_range(min, max);
    Ok(match args[0] {
        Value::Int4(_) => Value::Int4(value as i32),
        _ => Value::Int8(value),
    })
}

/// random_numeric returns a numeric in [min, max] with the larger of their scales, choosing its base-10000 digits as
/// Postgres' random_var does.
fn random_numeric(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (Value::Numeric(min), Value::Numeric(max)) = (&args[0], &args[1]) else {
        return Err(PgError::internal("random without numerics"));
    };
    for (value, bound) in [(min, "lower"), (max, "upper")] {
        match value {
            Numeric::NaN => return Err(invalid(format!("{bound} bound cannot be NaN"))),
            Numeric::Infinity | Numeric::NegativeInfinity => {
                return Err(invalid(format!("{bound} bound cannot be infinity")));
            }
            Numeric::Finite { .. } => {}
        }
    }
    let scale = |n: &Numeric| match n {
        Numeric::Finite { scale, .. } => *scale,
        _ => 0,
    };
    let rscale = scale(min).max(scale(max));
    let (min_coefficient, max_coefficient) = (scaled(min, rscale), scaled(max, rscale));
    let range = &max_coefficient - &min_coefficient;
    if range.sign() == Sign::Minus {
        return Err(invalid("lower bound must be less than or equal to upper bound"));
    }
    let range = range.magnitude().clone();
    if range.is_zero() {
        return Ok(Value::Numeric(finite(min_coefficient, rscale)));
    }
    let (digits, weight) = nbase_digits(&range, rscale);
    let fractional = rscale.div_ceil(4) as i64;
    let res_ndigits = (weight + 1 + fractional) as usize;
    let pow10 = 10u64.pow((fractional * 4 - rscale as i64) as u32);
    let mut range64 = digits[0];
    let mut range64_ndigits = 1;
    while range64_ndigits < res_ndigits && range64_ndigits < 4 {
        range64 *= NBASE;
        if range64_ndigits < digits.len() {
            range64 += digits[range64_ndigits];
        }
        range64_ndigits += 1;
    }
    let prng = generator(ctx);
    loop {
        let mut result = vec![0u64; res_ndigits];
        let mut value = match range64_ndigits == res_ndigits && pow10 != 1 {
            true => prng.uint64_range(0, range64 / pow10) * pow10,
            false => prng.uint64_range(0, range64),
        };
        for digit in result[..range64_ndigits].iter_mut().rev() {
            *digit = value % NBASE;
            value /= NBASE;
        }
        let whole = if pow10 != 1 { res_ndigits - 1 } else { res_ndigits };
        let mut i = range64_ndigits;
        while i + 3 < whole {
            let mut value = prng.uint64_range(0, NBASE * NBASE * NBASE * NBASE - 1);
            for _ in 0..3 {
                result[i] = value % NBASE;
                value /= NBASE;
                i += 1;
            }
            result[i] = value;
            i += 1;
        }
        while i < whole {
            result[i] = prng.uint64_range(0, NBASE - 1);
            i += 1;
        }
        if i < res_ndigits {
            result[i] = prng.uint64_range(0, NBASE / pow10 - 1) * pow10;
        }
        let mut number = BigUint::zero();
        for digit in &result {
            number = number * NBASE + digit;
        }
        let digits_after_point = (res_ndigits as i64 - 1 - weight) * 4;
        let coefficient = number / BigUint::from(10u32).pow((digits_after_point - rscale as i64) as u32);
        if coefficient <= range {
            return Ok(Value::Numeric(finite(BigInt::from(coefficient) + min_coefficient, rscale)));
        }
    }
}

/// scaled returns a finite numeric's value times 10^scale, for a scale at least its own.
fn scaled(value: &Numeric, scale: u32) -> BigInt {
    match value {
        Numeric::Finite { negative, coefficient, scale: own } => {
            let magnitude = coefficient * BigUint::from(10u32).pow(scale - own);
            BigInt::from_biguint(if *negative { Sign::Minus } else { Sign::Plus }, magnitude)
        }
        _ => BigInt::zero(),
    }
}

/// finite returns the numeric whose value times 10^scale is a coefficient.
fn finite(coefficient: BigInt, scale: u32) -> Numeric {
    let negative = coefficient.sign() == Sign::Minus;
    Numeric::Finite { negative, coefficient: coefficient.magnitude().clone(), scale }
}

/// nbase_digits returns the base-10000 digits of a positive value times 10^-scale without leading or trailing zero
/// digits, and the weight of the first, as a Postgres NumericVar holds them.
fn nbase_digits(coefficient: &BigUint, scale: u32) -> (Vec<u64>, i64) {
    let padded = scale.div_ceil(4) * 4;
    let mut value = coefficient * BigUint::from(10u32).pow(padded - scale);
    let mut digits = Vec::new();
    while !value.is_zero() {
        digits.push((&value % NBASE).to_u64().unwrap_or_default());
        value /= NBASE;
    }
    digits.reverse();
    let mut weight = digits.len() as i64 - (padded / 4) as i64 - 1;
    while digits.first() == Some(&0) {
        digits.remove(0);
        weight -= 1;
    }
    while digits.last() == Some(&0) {
        digits.pop();
    }
    (digits, weight)
}
