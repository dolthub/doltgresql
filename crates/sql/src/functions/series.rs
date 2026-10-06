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

//! Set-returning functions that generate series.

use super::Function;
use crate::error::{PgError, Result, code};
use crate::numeric::Numeric;
use crate::oid::{INT4, INT8, NUMERIC};
use crate::query::Ctx;
use crate::types::Value;

/// f declares a strict series function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// FUNCTIONS are the series functions.
pub const FUNCTIONS: &[Function] = &[
    f("generate_series", &[INT4, INT4], INT4, series_int),
    f("generate_series", &[INT4, INT4, INT4], INT4, series_int),
    f("generate_series", &[INT8, INT8], INT8, series_int),
    f("generate_series", &[INT8, INT8, INT8], INT8, series_int),
    f("generate_series", &[NUMERIC, NUMERIC], NUMERIC, series_numeric),
    f("generate_series", &[NUMERIC, NUMERIC, NUMERIC], NUMERIC, series_numeric),
];

/// MAX_ROWS limits a generated series, to fail instead of exhausting memory.
const MAX_ROWS: usize = 100_000_000;

/// series_int generates integers from the start to the stop by the step.
fn series_int(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let int = |v: &Value| match v {
        Value::Int4(i) => *i as i64,
        Value::Int8(i) => *i,
        _ => 0,
    };
    let (start, stop) = (int(&args[0]), int(&args[1]));
    let step = args.get(2).map_or(1, int);
    if step == 0 {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "step size cannot equal zero"));
    }
    let wrap = |i: i64| if matches!(args[0], Value::Int4(_)) { Value::Int4(i as i32) } else { Value::Int8(i) };
    let mut out = Vec::new();
    let mut i = start;
    while (step > 0 && i <= stop) || (step < 0 && i >= stop) {
        out.push(wrap(i));
        if out.len() > MAX_ROWS {
            return Err(PgError::new(code::PROGRAM_LIMIT_EXCEEDED, "generate_series produced too many rows"));
        }
        match i.checked_add(step) {
            Some(next) => i = next,
            None => break,
        }
    }
    Ok(Value::Set(out))
}

/// series_numeric generates numerics from the start to the stop by the step.
fn series_numeric(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let num = |v: &Value| match v {
        Value::Numeric(n) => n.clone(),
        _ => Numeric::zero(0),
    };
    let (start, stop) = (num(&args[0]), num(&args[1]));
    let step = args.get(2).map_or(Numeric::from_i64(1), num);
    if step.is_zero() {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "step size cannot equal zero"));
    }
    let ascending = step.cmp_numeric(&Numeric::zero(0)).is_gt();
    let mut out = Vec::new();
    let mut i = start;
    while (ascending && i.cmp_numeric(&stop).is_le()) || (!ascending && i.cmp_numeric(&stop).is_ge()) {
        out.push(Value::Numeric(i.clone()));
        if out.len() > MAX_ROWS {
            return Err(PgError::new(code::PROGRAM_LIMIT_EXCEEDED, "generate_series produced too many rows"));
        }
        i = i.add(&step);
    }
    Ok(Value::Set(out))
}
