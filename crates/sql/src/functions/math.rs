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

//! Mathematical functions.

use super::Function;
use crate::error::{PgError, Result, code};
use crate::numeric::Numeric;
use crate::oid::{FLOAT4, FLOAT8, INT2, INT4, INT8, NUMERIC};
use crate::query::Ctx;
use crate::types::Value;

/// f declares a strict math function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// FUNCTIONS are the math functions.
pub const FUNCTIONS: &[Function] = &[
    f("abs", &[INT2], INT2, abs),
    f("abs", &[INT4], INT4, abs),
    f("abs", &[INT8], INT8, abs),
    f("abs", &[FLOAT4], FLOAT4, abs),
    f("abs", &[FLOAT8], FLOAT8, abs),
    f("abs", &[NUMERIC], NUMERIC, abs),
    f("round", &[FLOAT8], FLOAT8, round_float),
    f("round", &[NUMERIC], NUMERIC, round_numeric),
    f("round", &[NUMERIC, INT4], NUMERIC, round_numeric_to),
    f("trunc", &[FLOAT8], FLOAT8, trunc_float),
    f("trunc", &[NUMERIC], NUMERIC, trunc_numeric),
    f("trunc", &[NUMERIC, INT4], NUMERIC, trunc_numeric_to),
    f("ceil", &[FLOAT8], FLOAT8, ceil_float),
    f("ceil", &[NUMERIC], NUMERIC, ceil_numeric),
    f("ceiling", &[FLOAT8], FLOAT8, ceil_float),
    f("ceiling", &[NUMERIC], NUMERIC, ceil_numeric),
    f("floor", &[FLOAT8], FLOAT8, floor_float),
    f("floor", &[NUMERIC], NUMERIC, floor_numeric),
    f("sign", &[FLOAT8], FLOAT8, sign_float),
    f("sign", &[NUMERIC], NUMERIC, sign_numeric),
    f("sqrt", &[FLOAT8], FLOAT8, sqrt),
    f("cbrt", &[FLOAT8], FLOAT8, cbrt),
    f("exp", &[FLOAT8], FLOAT8, exp),
    f("ln", &[FLOAT8], FLOAT8, ln),
    f("log", &[FLOAT8], FLOAT8, log10),
    f("log10", &[FLOAT8], FLOAT8, log10),
    f("power", &[FLOAT8, FLOAT8], FLOAT8, power),
    f("pow", &[FLOAT8, FLOAT8], FLOAT8, power),
    f("pi", &[], FLOAT8, pi),
    f("degrees", &[FLOAT8], FLOAT8, degrees),
    f("radians", &[FLOAT8], FLOAT8, radians),
    f("sin", &[FLOAT8], FLOAT8, sin),
    f("cos", &[FLOAT8], FLOAT8, cos),
    f("tan", &[FLOAT8], FLOAT8, tan),
    f("mod", &[INT2, INT2], INT2, modulo),
    f("mod", &[INT4, INT4], INT4, modulo),
    f("mod", &[INT8, INT8], INT8, modulo),
    f("mod", &[NUMERIC, NUMERIC], NUMERIC, modulo),
    f("gcd", &[INT4, INT4], INT4, gcd),
    f("gcd", &[INT8, INT8], INT8, gcd),
    f("lcm", &[INT4, INT4], INT4, lcm),
    f("lcm", &[INT8, INT8], INT8, lcm),
    f("div", &[NUMERIC, NUMERIC], NUMERIC, div),
];

/// float returns a float8 argument.
fn float(value: &Value) -> f64 {
    match value {
        Value::Float8(f) => *f,
        Value::Float4(f) => *f as f64,
        _ => 0.0,
    }
}

/// numeric returns a numeric argument.
fn numeric(value: &Value) -> Numeric {
    match value {
        Value::Numeric(n) => n.clone(),
        _ => Numeric::zero(0),
    }
}

/// out_of_range returns Postgres' error for an integer result outside its type.
fn out_of_range(value: &Value) -> PgError {
    let name = match value {
        Value::Int2(_) => "smallint",
        Value::Int4(_) => "integer",
        _ => "bigint",
    };
    PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, format!("{name} out of range"))
}

/// abs returns the absolute value.
fn abs(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let value = &args[0];
    Ok(match value {
        Value::Int2(i) => Value::Int2(i.checked_abs().ok_or_else(|| out_of_range(value))?),
        Value::Int4(i) => Value::Int4(i.checked_abs().ok_or_else(|| out_of_range(value))?),
        Value::Int8(i) => Value::Int8(i.checked_abs().ok_or_else(|| out_of_range(value))?),
        Value::Float4(f) => Value::Float4(f.abs()),
        Value::Float8(f) => Value::Float8(f.abs()),
        Value::Numeric(n) if n.cmp_numeric(&Numeric::zero(0)).is_lt() => Value::Numeric(n.negate()),
        other => other.clone(),
    })
}

/// round_float rounds half to even, as the C library's rint does.
fn round_float(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).round_ties_even()))
}

/// round_numeric rounds half away from zero to an integer.
fn round_numeric(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Numeric(numeric(&args[0]).with_scale(0)))
}

/// round_to returns a numeric rounded half away from zero, or truncated, to a number of decimal places, which may
/// be negative.
fn round_to(n: &Numeric, places: i32, truncate: bool) -> Numeric {
    if places >= 0 {
        let places = places as u32;
        if truncate {
            let shifted = n.mul(&Numeric::parse(&format!("1e{places}")).unwrap_or(Numeric::from_i64(1)));
            let whole = truncate_to_integer(&shifted);
            return whole.div_exact(places);
        }
        return n.with_scale(places);
    }
    let unit = Numeric::parse(&format!("1e{}", -places)).unwrap_or(Numeric::from_i64(1));
    let scaled = n.div(&unit).unwrap_or(Numeric::NaN);
    let rounded = if truncate { truncate_to_integer(&scaled) } else { scaled.with_scale(0) };
    rounded.mul(&unit).with_scale(0)
}

/// truncate_to_integer drops a numeric's fraction.
fn truncate_to_integer(n: &Numeric) -> Numeric {
    match n {
        Numeric::Finite { negative, coefficient, scale } => {
            let divisor = num_bigint::BigUint::from(10u32).pow(*scale);
            let whole = Numeric::Finite { negative: *negative, coefficient: coefficient / divisor, scale: 0 };
            if whole.is_zero() { Numeric::zero(0) } else { whole }
        }
        other => other.clone(),
    }
}

/// round_numeric_to rounds to a number of decimal places.
fn round_numeric_to(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let places = if let Value::Int4(p) = args[1] { p } else { 0 };
    Ok(Value::Numeric(round_to(&numeric(&args[0]), places, false)))
}

/// trunc_float drops the fraction.
fn trunc_float(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).trunc()))
}

/// trunc_numeric drops the fraction.
fn trunc_numeric(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Numeric(truncate_to_integer(&numeric(&args[0]))))
}

/// trunc_numeric_to truncates to a number of decimal places.
fn trunc_numeric_to(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let places = if let Value::Int4(p) = args[1] { p } else { 0 };
    Ok(Value::Numeric(round_to(&numeric(&args[0]), places, true)))
}

/// ceil_float returns the smallest integer not less than the value.
fn ceil_float(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).ceil()))
}

/// ceil_numeric returns the smallest integer not less than the value.
fn ceil_numeric(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let n = numeric(&args[0]);
    let whole = truncate_to_integer(&n);
    let bumped = if whole.cmp_numeric(&n).is_lt() { whole.add(&Numeric::from_i64(1)) } else { whole };
    Ok(Value::Numeric(bumped))
}

/// floor_float returns the largest integer not greater than the value.
fn floor_float(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).floor()))
}

/// floor_numeric returns the largest integer not greater than the value.
fn floor_numeric(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let n = numeric(&args[0]);
    let whole = truncate_to_integer(&n);
    let lowered = if whole.cmp_numeric(&n).is_gt() { whole.sub(&Numeric::from_i64(1)) } else { whole };
    Ok(Value::Numeric(lowered))
}

/// sign_float returns -1, 0, or 1.
fn sign_float(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let f = float(&args[0]);
    Ok(Value::Float8(if f > 0.0 {
        1.0
    } else if f < 0.0 {
        -1.0
    } else {
        0.0
    }))
}

/// sign_numeric returns -1, 0, or 1.
fn sign_numeric(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let n = numeric(&args[0]);
    Ok(Value::Numeric(Numeric::from_i64(match n.cmp_numeric(&Numeric::zero(0)) {
        std::cmp::Ordering::Less => -1,
        std::cmp::Ordering::Equal => 0,
        std::cmp::Ordering::Greater => 1,
    })))
}

/// sqrt returns the square root, failing for a negative value.
fn sqrt(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let f = float(&args[0]);
    if f < 0.0 {
        return Err(PgError::new(code::INVALID_ARGUMENT_FOR_POWER, "cannot take square root of a negative number"));
    }
    Ok(Value::Float8(f.sqrt()))
}

/// cbrt returns the cube root.
fn cbrt(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).cbrt()))
}

/// exp returns e to the power.
fn exp(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let result = float(&args[0]).exp();
    if result.is_infinite() {
        return Err(PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "value out of range: overflow"));
    }
    Ok(Value::Float8(result))
}

/// ln returns the natural logarithm.
fn ln(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let f = float(&args[0]);
    if f == 0.0 {
        return Err(PgError::new(code::INVALID_ARGUMENT_FOR_LOG, "cannot take logarithm of zero"));
    }
    if f < 0.0 {
        return Err(PgError::new(code::INVALID_ARGUMENT_FOR_LOG, "cannot take logarithm of a negative number"));
    }
    Ok(Value::Float8(f.ln()))
}

/// log10 returns the base 10 logarithm.
fn log10(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let f = float(&args[0]);
    if f == 0.0 {
        return Err(PgError::new(code::INVALID_ARGUMENT_FOR_LOG, "cannot take logarithm of zero"));
    }
    if f < 0.0 {
        return Err(PgError::new(code::INVALID_ARGUMENT_FOR_LOG, "cannot take logarithm of a negative number"));
    }
    Ok(Value::Float8(f.log10()))
}

/// power raises the first value to the second.
fn power(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (base, exponent) = (float(&args[0]), float(&args[1]));
    if base == 0.0 && exponent < 0.0 {
        return Err(PgError::new(code::INVALID_ARGUMENT_FOR_POWER, "zero raised to a negative power is undefined"));
    }
    if base < 0.0 && exponent.fract() != 0.0 {
        return Err(PgError::new(
            code::INVALID_ARGUMENT_FOR_POWER,
            "a negative number raised to a non-integer power yields a complex result",
        ));
    }
    let result = base.powf(exponent);
    if result.is_infinite() && base.is_finite() && exponent.is_finite() {
        return Err(PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "value out of range: overflow"));
    }
    Ok(Value::Float8(result))
}

/// pi returns π.
fn pi(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Float8(std::f64::consts::PI))
}

/// degrees converts radians to degrees.
fn degrees(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).to_degrees()))
}

/// radians converts degrees to radians.
fn radians(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).to_radians()))
}

/// sin returns the sine.
fn sin(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).sin()))
}

/// cos returns the cosine.
fn cos(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).cos()))
}

/// tan returns the tangent.
fn tan(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).tan()))
}

/// int returns an integer argument widened.
fn int(value: &Value) -> i64 {
    match value {
        Value::Int2(i) => *i as i64,
        Value::Int4(i) => *i as i64,
        Value::Int8(i) => *i,
        _ => 0,
    }
}

/// same_int returns an integer of the same type as the template, failing when it is out of that type's range.
fn same_int(template: &Value, value: Option<i64>) -> Result<Value> {
    let value = value.ok_or_else(|| out_of_range(template))?;
    match template {
        Value::Int2(_) => i16::try_from(value).map(Value::Int2).map_err(|_| out_of_range(template)),
        Value::Int4(_) => i32::try_from(value).map(Value::Int4).map_err(|_| out_of_range(template)),
        _ => Ok(Value::Int8(value)),
    }
}

/// modulo returns the remainder of truncating division.
fn modulo(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    if let (Value::Numeric(l), Value::Numeric(r)) = (&args[0], &args[1]) {
        return Ok(Value::Numeric(l.rem(r)?));
    }
    let (l, r) = (int(&args[0]), int(&args[1]));
    if r == 0 {
        return Err(PgError::new(code::DIVISION_BY_ZERO, "division by zero"));
    }
    same_int(&args[0], Some(l.checked_rem(r).unwrap_or(0)))
}

/// gcd_of returns the greatest common divisor of two magnitudes.
fn gcd_of(mut a: i128, mut b: i128) -> i128 {
    a = a.abs();
    b = b.abs();
    while b != 0 {
        (a, b) = (b, a % b);
    }
    a
}

/// gcd returns the greatest common divisor.
fn gcd(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let g = gcd_of(int(&args[0]) as i128, int(&args[1]) as i128);
    same_int(&args[0], i64::try_from(g).ok())
}

/// lcm returns the least common multiple.
fn lcm(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (a, b) = (int(&args[0]) as i128, int(&args[1]) as i128);
    if a == 0 || b == 0 {
        return same_int(&args[0], Some(0));
    }
    let l = (a / gcd_of(a, b) * b).abs();
    same_int(&args[0], i64::try_from(l).ok())
}

/// div returns the integer quotient, truncated toward zero.
fn div(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (l, r) = (numeric(&args[0]), numeric(&args[1]));
    let quotient = l.div(&r)?;
    Ok(Value::Numeric(truncate_to_integer(&quotient)))
}
