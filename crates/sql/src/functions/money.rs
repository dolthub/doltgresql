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

//! The money type's operators, functions, and casts, as Postgres' cash.c has them in the C locale, where an amount is
//! a count of cents.

use num_bigint::BigUint;

use super::Function;
use crate::error::{PgError, Result, code};
use crate::numeric::Numeric;
use crate::oid::{FLOAT4, FLOAT8, INT2, INT4, INT8, NUMERIC, TEXT};
use crate::types::{BaseValue, Value};

/// MONEY is the OID of the money type.
pub const MONEY: u32 = 790;

/// f declares a strict money function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// FUNCTIONS are money's operators and functions.
pub const FUNCTIONS: &[Function] = &[
    f("+", &[MONEY, MONEY], MONEY, |_, a| money(cents(&a[0]).checked_add(cents(&a[1])))),
    f("cash_pl", &[MONEY, MONEY], MONEY, |_, a| money(cents(&a[0]).checked_add(cents(&a[1])))),
    f("-", &[MONEY, MONEY], MONEY, |_, a| money(cents(&a[0]).checked_sub(cents(&a[1])))),
    f("cash_mi", &[MONEY, MONEY], MONEY, |_, a| money(cents(&a[0]).checked_sub(cents(&a[1])))),
    f("/", &[MONEY, MONEY], FLOAT8, |_, a| divide_money(&a[0], &a[1])),
    f("cash_div_cash", &[MONEY, MONEY], FLOAT8, |_, a| divide_money(&a[0], &a[1])),
    f("*", &[MONEY, INT8], MONEY, |_, a| money(cents(&a[0]).checked_mul(integer(&a[1])))),
    f("*", &[MONEY, INT4], MONEY, |_, a| money(cents(&a[0]).checked_mul(integer(&a[1])))),
    f("*", &[MONEY, INT2], MONEY, |_, a| money(cents(&a[0]).checked_mul(integer(&a[1])))),
    f("*", &[INT8, MONEY], MONEY, |_, a| money(cents(&a[1]).checked_mul(integer(&a[0])))),
    f("*", &[INT4, MONEY], MONEY, |_, a| money(cents(&a[1]).checked_mul(integer(&a[0])))),
    f("*", &[INT2, MONEY], MONEY, |_, a| money(cents(&a[1]).checked_mul(integer(&a[0])))),
    f("*", &[MONEY, FLOAT8], MONEY, |_, a| scale(&a[0], float(&a[1]), false)),
    f("*", &[MONEY, FLOAT4], MONEY, |_, a| scale(&a[0], float(&a[1]), false)),
    f("*", &[FLOAT8, MONEY], MONEY, |_, a| scale(&a[1], float(&a[0]), false)),
    f("*", &[FLOAT4, MONEY], MONEY, |_, a| scale(&a[1], float(&a[0]), false)),
    f("/", &[MONEY, INT8], MONEY, |_, a| divide(&a[0], integer(&a[1]))),
    f("/", &[MONEY, INT4], MONEY, |_, a| divide(&a[0], integer(&a[1]))),
    f("/", &[MONEY, INT2], MONEY, |_, a| divide(&a[0], integer(&a[1]))),
    f("/", &[MONEY, FLOAT8], MONEY, |_, a| scale(&a[0], float(&a[1]), true)),
    f("/", &[MONEY, FLOAT4], MONEY, |_, a| scale(&a[0], float(&a[1]), true)),
    f("cash_mul_int8", &[MONEY, INT8], MONEY, |_, a| money(cents(&a[0]).checked_mul(integer(&a[1])))),
    f("cash_mul_int4", &[MONEY, INT4], MONEY, |_, a| money(cents(&a[0]).checked_mul(integer(&a[1])))),
    f("cash_mul_int2", &[MONEY, INT2], MONEY, |_, a| money(cents(&a[0]).checked_mul(integer(&a[1])))),
    f("int8_mul_cash", &[INT8, MONEY], MONEY, |_, a| money(cents(&a[1]).checked_mul(integer(&a[0])))),
    f("int4_mul_cash", &[INT4, MONEY], MONEY, |_, a| money(cents(&a[1]).checked_mul(integer(&a[0])))),
    f("int2_mul_cash", &[INT2, MONEY], MONEY, |_, a| money(cents(&a[1]).checked_mul(integer(&a[0])))),
    f("cash_mul_flt8", &[MONEY, FLOAT8], MONEY, |_, a| scale(&a[0], float(&a[1]), false)),
    f("cash_mul_flt4", &[MONEY, FLOAT4], MONEY, |_, a| scale(&a[0], float(&a[1]), false)),
    f("flt8_mul_cash", &[FLOAT8, MONEY], MONEY, |_, a| scale(&a[1], float(&a[0]), false)),
    f("flt4_mul_cash", &[FLOAT4, MONEY], MONEY, |_, a| scale(&a[1], float(&a[0]), false)),
    f("cash_div_int8", &[MONEY, INT8], MONEY, |_, a| divide(&a[0], integer(&a[1]))),
    f("cash_div_int4", &[MONEY, INT4], MONEY, |_, a| divide(&a[0], integer(&a[1]))),
    f("cash_div_int2", &[MONEY, INT2], MONEY, |_, a| divide(&a[0], integer(&a[1]))),
    f("cash_div_flt8", &[MONEY, FLOAT8], MONEY, |_, a| scale(&a[0], float(&a[1]), true)),
    f("cash_div_flt4", &[MONEY, FLOAT4], MONEY, |_, a| scale(&a[0], float(&a[1]), true)),
    f("cashlarger", &[MONEY, MONEY], MONEY, |_, a| Ok(money_value(cents(&a[0]).max(cents(&a[1]))))),
    f("cashsmaller", &[MONEY, MONEY], MONEY, |_, a| Ok(money_value(cents(&a[0]).min(cents(&a[1]))))),
    f("cash_words", &[MONEY], TEXT, |_, a| Ok(Value::Text(words(cents(&a[0]))))),
    f("money", &[NUMERIC], MONEY, |_, a| of_numeric(&a[0])),
    f("money", &[INT4], MONEY, |_, a| of_integer(integer(&a[0]))),
    f("money", &[INT8], MONEY, |_, a| of_integer(integer(&a[0]))),
    f("numeric", &[MONEY], NUMERIC, |_, a| Ok(to_numeric(cents(&a[0])))),
];

/// cents returns the amount of a money value.
pub fn cents(value: &Value) -> i64 {
    match value {
        Value::Base(base) => i64::from_be_bytes(base.data.as_slice().try_into().unwrap_or_default()),
        _ => 0,
    }
}

/// money_value returns a money value of an amount.
pub fn money_value(cents: i64) -> Value {
    Value::Base(Box::new(BaseValue { type_oid: MONEY, data: cents.to_be_bytes().to_vec() }))
}

/// out_of_range is the error of a money result that does not fit.
pub fn out_of_range() -> PgError {
    PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "money out of range")
}

/// money returns a money value of an amount that an overflow check produced.
fn money(cents: Option<i64>) -> Result<Value> {
    cents.map(money_value).ok_or_else(out_of_range)
}

/// integer returns an integer argument as a bigint.
fn integer(value: &Value) -> i64 {
    match value {
        Value::Int2(i) => *i as i64,
        Value::Int4(i) => *i as i64,
        Value::Int8(i) => *i,
        _ => 0,
    }
}

/// float returns a float argument as a double.
fn float(value: &Value) -> f64 {
    match value {
        Value::Float4(f) => *f as f64,
        Value::Float8(f) => *f,
        _ => 0.0,
    }
}

/// division_by_zero is the error of dividing by zero.
fn division_by_zero() -> PgError {
    PgError::new(code::DIVISION_BY_ZERO, "division by zero")
}

/// divide_money divides two amounts into a double, as cash_div_cash does.
fn divide_money(dividend: &Value, divisor: &Value) -> Result<Value> {
    match cents(divisor) {
        0 => Err(division_by_zero()),
        divisor => Ok(Value::Float8(cents(dividend) as f64 / divisor as f64)),
    }
}

/// divide divides an amount by an integer, truncating toward zero, as cash_div_int64 does.
fn divide(value: &Value, divisor: i64) -> Result<Value> {
    match divisor {
        0 => Err(division_by_zero()),
        -1 => money(cents(value).checked_neg()),
        divisor => Ok(money_value(cents(value) / divisor)),
    }
}

/// scale multiplies or divides an amount by a double, rounding to the nearest cent, as cash_mul_float8 and
/// cash_div_float8 do with float8's overflow checks.
fn scale(value: &Value, factor: f64, divide: bool) -> Result<Value> {
    let amount = cents(value) as f64;
    let result = match divide {
        true if factor == 0.0 => return Err(division_by_zero()),
        true => amount / factor,
        false => amount * factor,
    };
    if result.is_infinite() && factor.is_finite() {
        return Err(PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "value out of range: overflow"));
    }
    if result == 0.0 && amount != 0.0 && (if divide { factor.is_finite() } else { factor != 0.0 }) {
        return Err(PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "value out of range: underflow"));
    }
    let result = result.round_ties_even();
    if result.is_nan() || !(i64::MIN as f64..-(i64::MIN as f64)).contains(&result) {
        return Err(out_of_range());
    }
    Ok(money_value(result as i64))
}

/// to_numeric returns an amount in dollars, with two decimal places, as cash_numeric does.
fn to_numeric(cents: i64) -> Value {
    Value::Numeric(Numeric::Finite { negative: cents < 0, coefficient: BigUint::from(cents.unsigned_abs()), scale: 2 })
}

/// of_numeric returns the amount of a number of dollars, rounded to the nearest cent, as numeric_cash does.
fn of_numeric(value: &Value) -> Result<Value> {
    let Value::Numeric(n) = value else { return Err(PgError::internal("a money amount that is not numeric")) };
    let scaled = n.mul(&Numeric::from_i64(100));
    match scaled {
        Numeric::NaN => Err(PgError::new(code::FEATURE_NOT_SUPPORTED, "cannot convert NaN to bigint")),
        Numeric::Infinity | Numeric::NegativeInfinity => {
            Err(PgError::new(code::FEATURE_NOT_SUPPORTED, "cannot convert infinity to bigint"))
        }
        finite => finite
            .to_i64()
            .map(money_value)
            .ok_or_else(|| PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "bigint out of range")),
    }
}

/// of_integer returns the amount of a whole number of dollars, as int4_cash and int8_cash do.
fn of_integer(dollars: i64) -> Result<Value> {
    dollars
        .checked_mul(100)
        .map(money_value)
        .ok_or_else(|| PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "bigint out of range"))
}

/// castable reports whether money has a cast to or from another type, each an assignment cast.
pub(crate) fn castable(from: u32, to: u32) -> bool {
    matches!((from, to), (MONEY, NUMERIC) | (NUMERIC | INT4 | INT8, MONEY))
}

/// cast converts a value to or from money as Postgres' casts between them do, returning None when there is no such
/// cast.
pub(crate) fn cast(value: &Value, to: u32) -> Option<Result<Value>> {
    match (value, to) {
        (Value::Base(base), NUMERIC) if base.type_oid == MONEY => Some(Ok(to_numeric(cents(value)))),
        (Value::Numeric(_), MONEY) => Some(of_numeric(value)),
        (Value::Int4(_) | Value::Int8(_), MONEY) => Some(of_integer(integer(value))),
        _ => None,
    }
}

/// number_words spells out a number below a thousand, as cash.c's num_word does.
fn number_words(value: i64) -> String {
    const SMALL: [&str; 28] = [
        "zero",
        "one",
        "two",
        "three",
        "four",
        "five",
        "six",
        "seven",
        "eight",
        "nine",
        "ten",
        "eleven",
        "twelve",
        "thirteen",
        "fourteen",
        "fifteen",
        "sixteen",
        "seventeen",
        "eighteen",
        "nineteen",
        "twenty",
        "thirty",
        "forty",
        "fifty",
        "sixty",
        "seventy",
        "eighty",
        "ninety",
    ];
    let small = |i: i64| SMALL[i as usize];
    let big = |i: i64| SMALL[i as usize + 18];
    let tu = value % 100;
    if value <= 20 {
        return small(value).to_string();
    }
    if tu == 0 {
        return format!("{} hundred", small(value / 100));
    }
    let tens = if value % 10 == 0 && tu > 10 {
        big(tu / 10).to_string()
    } else if tu < 20 {
        small(tu).to_string()
    } else {
        format!("{} {}", big(tu / 10), small(tu % 10))
    };
    match value > 99 {
        true if tu < 20 && !(value % 10 == 0 && tu > 10) => format!("{} hundred and {tens}", small(value / 100)),
        true => format!("{} hundred {tens}", small(value / 100)),
        false => tens,
    }
}

/// words spells out an amount in dollars and cents, as cash_words does.
fn words(value: i64) -> String {
    let mut out = if value < 0 { "minus ".to_string() } else { String::new() };
    let prefix = out.len();
    let val = value.unsigned_abs();
    let group = |divisor: u64| (val / divisor % 1000) as i64;
    for (divisor, name) in [
        (100_000_000_000_000_000, " quadrillion "),
        (100_000_000_000_000, " trillion "),
        (100_000_000_000, " billion "),
        (100_000_000, " million "),
        (100_000, " thousand "),
    ] {
        if group(divisor) != 0 {
            out.push_str(&number_words(group(divisor)));
            out.push_str(name);
        }
    }
    if group(100) != 0 {
        out.push_str(&number_words(group(100)));
    }
    if out.len() == prefix {
        out.push_str("zero");
    }
    out.push_str(if val / 100 == 1 { " dollar and " } else { " dollars and " });
    let cents = (val % 100) as i64;
    out.push_str(&number_words(cents));
    out.push_str(if cents == 1 { " cent" } else { " cents" });
    out[..1].to_ascii_uppercase() + &out[1..]
}
