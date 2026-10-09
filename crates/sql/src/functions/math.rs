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
    f("&", &[INT2, INT2], INT2, int_and),
    f("&", &[INT4, INT4], INT4, int_and),
    f("&", &[INT8, INT8], INT8, int_and),
    f("|", &[INT2, INT2], INT2, int_or),
    f("|", &[INT4, INT4], INT4, int_or),
    f("|", &[INT8, INT8], INT8, int_or),
    f("#", &[INT2, INT2], INT2, int_xor),
    f("#", &[INT4, INT4], INT4, int_xor),
    f("#", &[INT8, INT8], INT8, int_xor),
    f("~", &[INT2], INT2, int_not),
    f("~", &[INT4], INT4, int_not),
    f("~", &[INT8], INT8, int_not),
    f("<<", &[INT2, INT4], INT2, int_shift_left),
    f("<<", &[INT4, INT4], INT4, int_shift_left),
    f("<<", &[INT8, INT4], INT8, int_shift_left),
    f(">>", &[INT2, INT4], INT2, int_shift_right),
    f(">>", &[INT4, INT4], INT4, int_shift_right),
    f(">>", &[INT8, INT4], INT8, int_shift_right),
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
    f("sqrt", &[NUMERIC], NUMERIC, sqrt_numeric),
    f("cbrt", &[FLOAT8], FLOAT8, cbrt),
    f("exp", &[FLOAT8], FLOAT8, exp),
    f("exp", &[NUMERIC], NUMERIC, exp_numeric),
    f("ln", &[FLOAT8], FLOAT8, ln),
    f("ln", &[NUMERIC], NUMERIC, ln_numeric),
    f("log", &[FLOAT8], FLOAT8, log10),
    f("log", &[NUMERIC], NUMERIC, log10_numeric),
    f("log", &[NUMERIC, NUMERIC], NUMERIC, log_numeric),
    f("log10", &[FLOAT8], FLOAT8, log10),
    f("log10", &[NUMERIC], NUMERIC, log10_numeric),
    f("power", &[FLOAT8, FLOAT8], FLOAT8, power),
    f("power", &[NUMERIC, NUMERIC], NUMERIC, power_numeric),
    f("pow", &[FLOAT8, FLOAT8], FLOAT8, power),
    f("pow", &[NUMERIC, NUMERIC], NUMERIC, power_numeric),
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
    f("acos", &[FLOAT8], FLOAT8, acos),
    f("asin", &[FLOAT8], FLOAT8, asin),
    f("atan", &[FLOAT8], FLOAT8, atan),
    f("atan2", &[FLOAT8, FLOAT8], FLOAT8, atan2),
    f("cot", &[FLOAT8], FLOAT8, cot),
    f("sind", &[FLOAT8], FLOAT8, sind),
    f("cosd", &[FLOAT8], FLOAT8, cosd),
    f("tand", &[FLOAT8], FLOAT8, tand),
    f("cotd", &[FLOAT8], FLOAT8, cotd),
    f("asind", &[FLOAT8], FLOAT8, asind),
    f("acosd", &[FLOAT8], FLOAT8, acosd),
    f("atand", &[FLOAT8], FLOAT8, atand),
    f("atan2d", &[FLOAT8, FLOAT8], FLOAT8, atan2d),
    f("sinh", &[FLOAT8], FLOAT8, sinh),
    f("cosh", &[FLOAT8], FLOAT8, cosh),
    f("tanh", &[FLOAT8], FLOAT8, tanh),
    f("asinh", &[FLOAT8], FLOAT8, asinh),
    f("acosh", &[FLOAT8], FLOAT8, acosh),
    f("atanh", &[FLOAT8], FLOAT8, atanh),
    f("factorial", &[INT8], NUMERIC, factorial),
    f("scale", &[NUMERIC], INT4, scale),
    f("trim_scale", &[NUMERIC], NUMERIC, trim_scale),
    f("width_bucket", &[FLOAT8, FLOAT8, FLOAT8, INT4], INT4, width_bucket_float),
    f("width_bucket", &[NUMERIC, NUMERIC, NUMERIC, INT4], INT4, width_bucket_numeric),
    Function { name: "random", args: &[], ret: FLOAT8, strict: false, variadic: false, implementation: random },
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

/// sqrt_numeric returns the square root of a numeric.
fn sqrt_numeric(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Numeric(crate::numeric_math::sqrt(&numeric(&args[0]))?))
}

/// exp_numeric returns e raised to a numeric.
fn exp_numeric(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Numeric(crate::numeric_math::exp(&numeric(&args[0]))?))
}

/// ln_numeric returns the natural logarithm of a numeric.
fn ln_numeric(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Numeric(crate::numeric_math::ln(&numeric(&args[0]))?))
}

/// log10_numeric returns the base 10 logarithm of a numeric.
fn log10_numeric(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Numeric(crate::numeric_math::log(&Numeric::from_i64(10), &numeric(&args[0]))?))
}

/// log_numeric returns the logarithm of a numeric to a base.
fn log_numeric(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Numeric(crate::numeric_math::log(&numeric(&args[0]), &numeric(&args[1]))?))
}

/// power_numeric returns a numeric raised to a numeric power.
fn power_numeric(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Numeric(crate::numeric_math::power(&numeric(&args[0]), &numeric(&args[1]))?))
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

/// int_and returns the bitwise AND of two integers.
fn int_and(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    same_int(&args[0], Some(int(&args[0]) & int(&args[1])))
}

/// int_or returns the bitwise OR of two integers.
fn int_or(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    same_int(&args[0], Some(int(&args[0]) | int(&args[1])))
}

/// int_xor returns the bitwise exclusive OR of two integers.
fn int_xor(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    same_int(&args[0], Some(int(&args[0]) ^ int(&args[1])))
}

/// int_not returns the bitwise complement of an integer.
fn int_not(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    same_int(&args[0], Some(!int(&args[0])))
}

/// int_shift shifts an integer left by a count, or right for `right`, wrapping the count to the width as C's shifts
/// do and keeping the integer's type.
fn int_shift(value: &Value, count: i64, right: bool) -> Value {
    let count = count as u32;
    match value {
        Value::Int8(i) => Value::Int8(if right { i.wrapping_shr(count) } else { i.wrapping_shl(count) }),
        other => {
            let i = int(other) as i32;
            let shifted = if right { i.wrapping_shr(count) } else { i.wrapping_shl(count) };
            if let Value::Int2(_) = other { Value::Int2(shifted as i16) } else { Value::Int4(shifted) }
        }
    }
}

/// int_shift_left shifts an integer left.
fn int_shift_left(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(int_shift(&args[0], int(&args[1]), false))
}

/// int_shift_right shifts an integer right, keeping its sign.
fn int_shift_right(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(int_shift(&args[0], int(&args[1]), true))
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

/// trig_input fails as Postgres does for an infinite trigonometric input.
fn trig_input(value: f64) -> Result<f64> {
    if value.is_infinite() {
        return Err(PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "input is out of range"));
    }
    Ok(value)
}

/// inverse_input fails as Postgres does for an inverse sine or cosine input outside -1 to 1.
fn inverse_input(value: f64) -> Result<f64> {
    if !(-1.0..=1.0).contains(&value) {
        return Err(PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "input is out of range"));
    }
    Ok(value)
}

/// acos returns the inverse cosine in radians.
fn acos(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let f = float(&args[0]);
    if f.is_nan() {
        return Ok(Value::Float8(f64::NAN));
    }
    Ok(Value::Float8(inverse_input(f)?.acos()))
}

/// asin returns the inverse sine in radians.
fn asin(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let f = float(&args[0]);
    if f.is_nan() {
        return Ok(Value::Float8(f64::NAN));
    }
    Ok(Value::Float8(inverse_input(f)?.asin()))
}

/// atan returns the inverse tangent in radians.
fn atan(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).atan()))
}

/// atan2 returns the inverse tangent of `y/x` in radians.
fn atan2(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).atan2(float(&args[1]))))
}

/// cot returns the cotangent.
fn cot(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let f = float(&args[0]);
    if f.is_nan() {
        return Ok(Value::Float8(f64::NAN));
    }
    Ok(Value::Float8(1.0 / trig_input(f)?.tan()))
}

/// sind_0_to_30 returns the sine of degrees from 0 to 30, exact at 30, as Postgres computes it.
fn sind_0_to_30(x: f64) -> f64 {
    (x.to_radians().sin() / 30f64.to_radians().sin()) / 2.0
}

/// cosd_0_to_60 returns the cosine of degrees from 0 to 60, exact at 60, as Postgres computes it.
fn cosd_0_to_60(x: f64) -> f64 {
    1.0 - ((1.0 - x.to_radians().cos()) / (1.0 - 60f64.to_radians().cos())) / 2.0
}

/// sind_q1 returns the sine of degrees from 0 to 90.
fn sind_q1(x: f64) -> f64 {
    if x <= 30.0 { sind_0_to_30(x) } else { cosd_0_to_60(90.0 - x) }
}

/// cosd_q1 returns the cosine of degrees from 0 to 90.
fn cosd_q1(x: f64) -> f64 {
    if x <= 60.0 { cosd_0_to_60(x) } else { sind_0_to_30(90.0 - x) }
}

/// first_quadrant reduces degrees to 0 through 90, returning them with the signs that the sine and cosine of the
/// original angle take.
fn first_quadrant(degrees: f64) -> (f64, f64, f64) {
    let (mut x, mut sin_sign, mut cos_sign) = (degrees % 360.0, 1.0, 1.0);
    if x < 0.0 {
        x = -x;
        sin_sign = -sin_sign;
    }
    if x > 180.0 {
        x = 360.0 - x;
        sin_sign = -sin_sign;
    }
    if x > 90.0 {
        x = 180.0 - x;
        cos_sign = -cos_sign;
    }
    (x, sin_sign, cos_sign)
}

/// degree_input returns a degree argument, or the NaN result it gives, failing for an infinity.
fn degree_input(value: &Value) -> Result<std::result::Result<f64, Value>> {
    let f = float(value);
    if f.is_nan() {
        return Ok(Err(Value::Float8(f64::NAN)));
    }
    trig_input(f).map(Ok)
}

/// sind returns the sine of degrees, exact at multiples of 30 and 90.
fn sind(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let f = match degree_input(&args[0])? {
        Ok(f) => f,
        Err(nan) => return Ok(nan),
    };
    let (x, sin_sign, _) = first_quadrant(f);
    Ok(Value::Float8(sin_sign * sind_q1(x)))
}

/// cosd returns the cosine of degrees, exact at multiples of 60 and 90.
fn cosd(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let f = match degree_input(&args[0])? {
        Ok(f) => f,
        Err(nan) => return Ok(nan),
    };
    let (x, _, cos_sign) = first_quadrant(f);
    Ok(Value::Float8(cos_sign * cosd_q1(x)))
}

/// tand returns the tangent of degrees, exact at multiples of 45, without negative zero.
fn tand(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let f = match degree_input(&args[0])? {
        Ok(f) => f,
        Err(nan) => return Ok(nan),
    };
    let (x, sin_sign, cos_sign) = first_quadrant(f);
    let result = sin_sign * cos_sign * ((sind_q1(x) / cosd_q1(x)) / (sind_q1(45.0) / cosd_q1(45.0)));
    Ok(Value::Float8(if result == 0.0 { 0.0 } else { result }))
}

/// cotd returns the cotangent of degrees, exact at multiples of 45, without negative zero.
fn cotd(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let f = match degree_input(&args[0])? {
        Ok(f) => f,
        Err(nan) => return Ok(nan),
    };
    let (x, sin_sign, cos_sign) = first_quadrant(f);
    let result = sin_sign * cos_sign * ((cosd_q1(x) / sind_q1(x)) / (cosd_q1(45.0) / sind_q1(45.0)));
    Ok(Value::Float8(if result == 0.0 { 0.0 } else { result }))
}

/// asind_q1 returns the inverse sine in degrees of a value from 0 to 1, exact at 0.5 and 1.
fn asind_q1(x: f64) -> f64 {
    if x <= 0.5 { (x.asin() / 0.5f64.asin()) * 30.0 } else { 90.0 - (x.acos() / 0.5f64.acos()) * 60.0 }
}

/// acosd_q1 returns the inverse cosine in degrees of a value from 0 to 1, exact at 0.5 and 1.
fn acosd_q1(x: f64) -> f64 {
    if x <= 0.5 { 90.0 - (x.asin() / 0.5f64.asin()) * 30.0 } else { (x.acos() / 0.5f64.acos()) * 60.0 }
}

/// asind returns the inverse sine in degrees.
fn asind(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let f = float(&args[0]);
    if f.is_nan() {
        return Ok(Value::Float8(f64::NAN));
    }
    let f = inverse_input(f)?;
    Ok(Value::Float8(if f >= 0.0 { asind_q1(f) } else { -asind_q1(-f) }))
}

/// acosd returns the inverse cosine in degrees.
fn acosd(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let f = float(&args[0]);
    if f.is_nan() {
        return Ok(Value::Float8(f64::NAN));
    }
    let f = inverse_input(f)?;
    Ok(Value::Float8(if f >= 0.0 { acosd_q1(f) } else { 90.0 + asind_q1(-f) }))
}

/// atand returns the inverse tangent in degrees, exact at 1.
fn atand(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8((float(&args[0]).atan() / 1f64.atan()) * 45.0))
}

/// atan2d returns the inverse tangent of `y/x` in degrees.
fn atan2d(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8((float(&args[0]).atan2(float(&args[1])) / 1f64.atan()) * 45.0))
}

/// sinh returns the hyperbolic sine.
fn sinh(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).sinh()))
}

/// cosh returns the hyperbolic cosine.
fn cosh(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).cosh()))
}

/// tanh returns the hyperbolic tangent.
fn tanh(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).tanh()))
}

/// asinh returns the inverse hyperbolic sine.
fn asinh(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Float8(float(&args[0]).asinh()))
}

/// acosh returns the inverse hyperbolic cosine, which only inputs of at least 1 have.
fn acosh(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let f = float(&args[0]);
    if f < 1.0 {
        return Err(PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "input is out of range"));
    }
    Ok(Value::Float8(f.acosh()))
}

/// atanh returns the inverse hyperbolic tangent, which only inputs from -1 to 1 have.
fn atanh(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let f = float(&args[0]);
    if f.is_nan() {
        return Ok(Value::Float8(f64::NAN));
    }
    Ok(Value::Float8(inverse_input(f)?.atanh()))
}

/// factorial returns the product of the integers from 1 to the argument.
fn factorial(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let n = int(&args[0]);
    if n < 0 {
        return Err(PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "factorial of a negative number is undefined"));
    }
    if n > 32177 {
        return Err(PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "value overflows numeric format"));
    }
    let product = (2..=n.max(1)).fold(num_bigint::BigUint::from(1u32), |acc, i| acc * i as u64);
    Ok(Value::Numeric(Numeric::Finite { negative: false, coefficient: product, scale: 0 }))
}

/// scale returns the display scale of a finite numeric, and NULL for NaN and the infinities.
fn scale(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(match numeric(&args[0]) {
        Numeric::Finite { scale, .. } => Value::Int4(scale as i32),
        _ => Value::Null,
    })
}

/// trim_scale removes a numeric's trailing fractional zeroes.
fn trim_scale(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Numeric(numeric(&args[0]).trimmed()))
}

/// bucket_error returns Postgres' error for invalid width_bucket arguments.
fn bucket_error(message: &str) -> PgError {
    PgError::new(code::INVALID_ARGUMENT_FOR_WIDTH_BUCKET_FUNCTION, message)
}

/// last_bucket returns the bucket past the last, for operands beyond the upper bound.
fn last_bucket(count: i32) -> Result<Value> {
    let bucket =
        count.checked_add(1).ok_or_else(|| PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "integer out of range"))?;
    Ok(Value::Int4(bucket))
}

/// width_bucket_float returns the bucket of an operand among `count` equal-width buckets between the bounds.
fn width_bucket_float(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (operand, bound1, bound2, count) = (float(&args[0]), float(&args[1]), float(&args[2]), int(&args[3]) as i32);
    if count <= 0 {
        return Err(bucket_error("count must be greater than zero"));
    }
    if operand.is_nan() || bound1.is_nan() || bound2.is_nan() {
        return Err(bucket_error("operand, lower bound, and upper bound cannot be NaN"));
    }
    if bound1.is_infinite() || bound2.is_infinite() {
        return Err(bucket_error("lower and upper bounds must be finite"));
    }
    if bound1 == bound2 {
        return Err(bucket_error("lower bound cannot equal upper bound"));
    }
    let (operand, low, high) = if bound1 < bound2 { (operand, bound1, bound2) } else { (-operand, -bound1, -bound2) };
    if operand < low {
        return Ok(Value::Int4(0));
    }
    if operand >= high {
        return last_bucket(count);
    }
    let fraction = if (high - low).is_finite() {
        (operand - low) / (high - low)
    } else {
        (operand / 2.0 - low / 2.0) / (high / 2.0 - low / 2.0)
    };
    let bucket = ((count as f64 * fraction) as i32).min(count - 1);
    Ok(Value::Int4(bucket + 1))
}

/// width_bucket_numeric returns the bucket of an operand among `count` equal-width buckets between the bounds.
fn width_bucket_numeric(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (operand, bound1, bound2, count) =
        (numeric(&args[0]), numeric(&args[1]), numeric(&args[2]), int(&args[3]) as i32);
    if count <= 0 {
        return Err(bucket_error("count must be greater than zero"));
    }
    if [&operand, &bound1, &bound2].iter().any(|n| matches!(n, Numeric::NaN)) {
        return Err(bucket_error("operand, lower bound, and upper bound cannot be NaN"));
    }
    if [&bound1, &bound2].iter().any(|n| !matches!(n, Numeric::Finite { .. })) {
        return Err(bucket_error("lower and upper bounds must be finite"));
    }
    let (operand, low, high) = match bound1.cmp_numeric(&bound2) {
        std::cmp::Ordering::Equal => return Err(bucket_error("lower bound cannot equal upper bound")),
        std::cmp::Ordering::Less => (operand, bound1, bound2),
        std::cmp::Ordering::Greater => (operand.negate(), bound1.negate(), bound2.negate()),
    };
    if operand.cmp_numeric(&low).is_lt() {
        return Ok(Value::Int4(0));
    }
    if operand.cmp_numeric(&high).is_ge() {
        return last_bucket(count);
    }
    let (
        Numeric::Finite { coefficient: a, scale: a_scale, .. },
        Numeric::Finite { coefficient: b, scale: b_scale, .. },
    ) = (operand.sub(&low), high.sub(&low))
    else {
        return Ok(Value::Null);
    };
    let ten = num_bigint::BigUint::from(10u32);
    let bucket = (a * count as u64 * ten.pow(b_scale)) / (b * ten.pow(a_scale));
    let bucket = i32::try_from(bucket).unwrap_or(count).min(count - 1);
    Ok(Value::Int4(bucket + 1))
}

/// random returns a uniformly random value from 0 up to 1.
fn random(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Float8(rand::random::<f64>()))
}
