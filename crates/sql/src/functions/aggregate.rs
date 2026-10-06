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

//! Aggregate functions, which fold the rows of a group into one value.

use std::cmp::Ordering;
use std::collections::HashSet;

use crate::cast::type_display;
use crate::error::{PgError, Result, code};
use crate::expr::{Expr, compare_values, position};
use crate::numeric::Numeric;
use crate::oid::{BOOL, FLOAT4, FLOAT8, INT2, INT4, INT8, NUMERIC, TEXT};
use crate::query::Ctx;
use crate::types::Value;

use super::{ANYELEMENT, implicitly_castable};

/// Kind is what an aggregate computes.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Kind {
    CountStar,
    Count,
    Sum,
    Avg,
    Min,
    Max,
    BoolAnd,
    BoolOr,
    StringAgg,
    VarPop,
    VarSamp,
    StddevPop,
    StddevSamp,
}

/// Aggregate is one overload of an aggregate function.
#[derive(Debug)]
pub struct Aggregate {
    pub name: &'static str,
    pub args: &'static [u32],
    pub ret: u32,
    pub kind: Kind,
}

/// a declares an aggregate overload.
const fn a(name: &'static str, args: &'static [u32], ret: u32, kind: Kind) -> Aggregate {
    Aggregate { name, args, ret, kind }
}

/// AGGREGATES are the built-in aggregates.
pub const AGGREGATES: &[Aggregate] = &[
    a("count", &[], INT8, Kind::CountStar),
    a("count", &[super::ANY], INT8, Kind::Count),
    a("sum", &[INT2], INT8, Kind::Sum),
    a("sum", &[INT4], INT8, Kind::Sum),
    a("sum", &[INT8], NUMERIC, Kind::Sum),
    a("sum", &[NUMERIC], NUMERIC, Kind::Sum),
    a("sum", &[FLOAT4], FLOAT4, Kind::Sum),
    a("sum", &[FLOAT8], FLOAT8, Kind::Sum),
    a("avg", &[INT2], NUMERIC, Kind::Avg),
    a("avg", &[INT4], NUMERIC, Kind::Avg),
    a("avg", &[INT8], NUMERIC, Kind::Avg),
    a("avg", &[NUMERIC], NUMERIC, Kind::Avg),
    a("avg", &[FLOAT4], FLOAT8, Kind::Avg),
    a("avg", &[FLOAT8], FLOAT8, Kind::Avg),
    a("min", &[ANYELEMENT], ANYELEMENT, Kind::Min),
    a("max", &[ANYELEMENT], ANYELEMENT, Kind::Max),
    a("bool_and", &[BOOL], BOOL, Kind::BoolAnd),
    a("every", &[BOOL], BOOL, Kind::BoolAnd),
    a("bool_or", &[BOOL], BOOL, Kind::BoolOr),
    a("string_agg", &[TEXT, TEXT], TEXT, Kind::StringAgg),
    a("var_pop", &[NUMERIC], NUMERIC, Kind::VarPop),
    a("var_pop", &[FLOAT8], FLOAT8, Kind::VarPop),
    a("var_samp", &[NUMERIC], NUMERIC, Kind::VarSamp),
    a("var_samp", &[FLOAT8], FLOAT8, Kind::VarSamp),
    a("variance", &[NUMERIC], NUMERIC, Kind::VarSamp),
    a("variance", &[FLOAT8], FLOAT8, Kind::VarSamp),
    a("stddev_pop", &[NUMERIC], NUMERIC, Kind::StddevPop),
    a("stddev_pop", &[FLOAT8], FLOAT8, Kind::StddevPop),
    a("stddev_samp", &[NUMERIC], NUMERIC, Kind::StddevSamp),
    a("stddev_samp", &[FLOAT8], FLOAT8, Kind::StddevSamp),
    a("stddev", &[NUMERIC], NUMERIC, Kind::StddevSamp),
    a("stddev", &[FLOAT8], FLOAT8, Kind::StddevSamp),
];

/// exists reports whether an aggregate of the name exists.
pub fn exists(name: &str) -> bool {
    AGGREGATES.iter().any(|a| a.name == name)
}

/// AggCall is a call of an aggregate in a grouped query, over the input rows.
#[derive(Clone, Debug, PartialEq)]
pub struct AggCall {
    pub index: usize,
    pub args: Vec<Expr>,
    pub distinct: bool,
    pub filter: Option<Expr>,
    /// The aggregate's ORDER BY keys, with their directions and NULLS placement.
    pub order: Vec<(Expr, bool, bool)>,
    /// The result type.
    pub ret: u32,
}

/// resolve chooses the aggregate overload for arguments of the types, preferring exact matches, then integer and
/// numeric parameters for integer arguments as Postgres' preferred numeric promotions do.
pub fn resolve(name: &str, types: &[u32], location: i32) -> Result<(usize, Vec<u32>, u32)> {
    let mut candidates: Vec<usize> = AGGREGATES
        .iter()
        .enumerate()
        .filter(|(_, a)| a.name == name && a.args.len() == types.len())
        .filter(|(_, a)| a.args.iter().zip(types).all(|(&p, &t)| implicitly_castable(t, p)))
        .map(|(i, _)| i)
        .collect();
    let exact = |i: &usize| AGGREGATES[*i].args.iter().zip(types).filter(|(p, t)| *p == *t).count();
    let best = candidates.iter().map(exact).max().unwrap_or(0);
    candidates.retain(|i| exact(i) == best);
    // Among conversions, Postgres prefers the preferred type of the category, float8 for numbers.
    if candidates.len() > 1 {
        let preferred = |i: &usize| AGGREGATES[*i].args.iter().filter(|&&p| p == FLOAT8 || p == TEXT).count();
        let most = candidates.iter().map(preferred).max().unwrap_or(0);
        candidates.retain(|i| preferred(i) == most);
    }
    let Some(&index) = candidates.first() else {
        return Err(PgError {
            position: position(location),
            hint: Some(
                "No function matches the given name and argument types. You might need to add explicit type casts."
                    .into(),
            ),
            ..PgError::new(
                code::UNDEFINED_FUNCTION,
                format!(
                    "function {name}({}) does not exist",
                    types.iter().map(|&t| type_display(t)).collect::<Vec<_>>().join(", ")
                ),
            )
        });
    };
    let aggregate = &AGGREGATES[index];
    let element = types.first().copied().filter(|&t| t != crate::oid::UNKNOWN).unwrap_or(TEXT);
    let arg_types = aggregate
        .args
        .iter()
        .zip(types)
        .map(
            |(&p, &t)| {
                if p == ANYELEMENT || p == super::ANY { if t == crate::oid::UNKNOWN { TEXT } else { t } } else { p }
            },
        )
        .collect();
    let ret = if aggregate.ret == ANYELEMENT { element } else { aggregate.ret };
    Ok((index, arg_types, ret))
}

/// Accumulator collects a group's argument rows for one aggregate call.
pub struct Accumulator {
    rows: Vec<Vec<Value>>,
    /// The ORDER BY key values of each row.
    keys: Vec<Vec<Value>>,
}

impl Accumulator {
    /// new starts an empty accumulator for the call.
    pub fn new(_: &AggCall) -> Accumulator {
        Accumulator { rows: Vec::new(), keys: Vec::new() }
    }

    /// add adds an input row to the group.
    pub fn add(&mut self, ctx: &mut Ctx<'_>, call: &AggCall, row: &[Value]) -> Result<()> {
        if let Some(filter) = &call.filter
            && !filter.is_true(ctx, row)?
        {
            return Ok(());
        }
        let args = call.args.iter().map(|a| a.eval(ctx, row)).collect::<Result<Vec<_>>>()?;
        let keys = call.order.iter().map(|k| k.0.eval(ctx, row)).collect::<Result<Vec<_>>>()?;
        self.rows.push(args);
        self.keys.push(keys);
        Ok(())
    }

    /// finish computes the aggregate over the group.
    pub fn finish(self, _: &mut Ctx<'_>, call: &AggCall) -> Result<Value> {
        let aggregate = &AGGREGATES[call.index];
        let mut rows: Vec<(Vec<Value>, Vec<Value>)> = self.keys.into_iter().zip(self.rows).collect();
        if !call.order.is_empty() {
            rows.sort_by(|a, b| {
                for (i, (_, descending, nulls_first)) in call.order.iter().enumerate() {
                    let ordering = match (&a.0[i], &b.0[i]) {
                        (Value::Null, Value::Null) => Ordering::Equal,
                        (Value::Null, _) => {
                            if *nulls_first {
                                Ordering::Less
                            } else {
                                Ordering::Greater
                            }
                        }
                        (_, Value::Null) => {
                            if *nulls_first {
                                Ordering::Greater
                            } else {
                                Ordering::Less
                            }
                        }
                        (l, r) => {
                            let o = compare_values(l, r);
                            if *descending { o.reverse() } else { o }
                        }
                    };
                    if ordering != Ordering::Equal {
                        return ordering;
                    }
                }
                Ordering::Equal
            });
        }
        let mut args: Vec<Vec<Value>> = rows.into_iter().map(|(_, args)| args).collect();
        if call.distinct {
            let mut seen = HashSet::new();
            args.retain(|row| {
                seen.insert(
                    row.iter().map(|v| v.output().unwrap_or_else(|| "\u{0}".into())).collect::<Vec<_>>().join("\u{1}"),
                )
            });
        }
        if aggregate.kind == Kind::CountStar {
            return Ok(Value::Int8(args.len() as i64));
        }
        // Every other aggregate skips NULL inputs.
        let values: Vec<Value> = match aggregate.kind {
            Kind::StringAgg => {
                let parts: Vec<(String, Option<String>)> = args
                    .iter()
                    .filter(|r| !r[0].is_null())
                    .map(|r| (r[0].output().unwrap_or_default(), r[1].output()))
                    .collect();
                if parts.is_empty() {
                    return Ok(Value::Null);
                }
                let mut out = String::new();
                for (i, (part, separator)) in parts.into_iter().enumerate() {
                    if i > 0 {
                        out.push_str(separator.as_deref().unwrap_or(""));
                    }
                    out.push_str(&part);
                }
                return Ok(Value::Text(out));
            }
            _ => args.into_iter().filter_map(|r| r.into_iter().next()).filter(|v| !v.is_null()).collect(),
        };
        match aggregate.kind {
            Kind::Count => Ok(Value::Int8(values.len() as i64)),
            _ if values.is_empty() => Ok(Value::Null),
            Kind::Sum => sum(&values, call.ret),
            Kind::Avg => avg(&values, call.ret),
            Kind::Min => Ok(values
                .into_iter()
                .reduce(|a, b| if compare_values(&b, &a) == Ordering::Less { b } else { a })
                .unwrap_or(Value::Null)),
            Kind::Max => Ok(values
                .into_iter()
                .reduce(|a, b| if compare_values(&b, &a) == Ordering::Greater { b } else { a })
                .unwrap_or(Value::Null)),
            Kind::BoolAnd => Ok(Value::Bool(values.iter().all(|v| *v == Value::Bool(true)))),
            Kind::BoolOr => Ok(Value::Bool(values.contains(&Value::Bool(true)))),
            Kind::VarPop | Kind::VarSamp | Kind::StddevPop | Kind::StddevSamp => {
                variance(&values, aggregate.kind, call.ret)
            }
            Kind::CountStar | Kind::StringAgg => unreachable!("handled above"),
        }
    }
}

/// numeric_of converts an integer or numeric value to numeric.
fn numeric_of(value: &Value) -> Numeric {
    match value {
        Value::Int2(i) => Numeric::from_i64(*i as i64),
        Value::Int4(i) => Numeric::from_i64(*i as i64),
        Value::Int8(i) => Numeric::from_i64(*i),
        Value::Numeric(n) => n.clone(),
        _ => Numeric::NaN,
    }
}

/// float_of converts a float value to f64.
fn float_of(value: &Value) -> f64 {
    match value {
        Value::Float4(f) => *f as f64,
        Value::Float8(f) => *f,
        other => numeric_of(other).to_f64(),
    }
}

/// sum adds the values in the result type.
fn sum(values: &[Value], ret: u32) -> Result<Value> {
    Ok(match ret {
        INT8 => {
            let mut total: i64 = 0;
            for v in values {
                let n = match v {
                    Value::Int2(i) => *i as i64,
                    Value::Int4(i) => *i as i64,
                    Value::Int8(i) => *i,
                    _ => 0,
                };
                total = total
                    .checked_add(n)
                    .ok_or_else(|| PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "bigint out of range"))?;
            }
            Value::Int8(total)
        }
        FLOAT4 => Value::Float4(values.iter().map(|v| float_of(v) as f32).sum()),
        FLOAT8 => Value::Float8(values.iter().map(float_of).sum()),
        _ => Value::Numeric(values.iter().fold(Numeric::zero(0), |total, v| total.add(&numeric_of(v)))),
    })
}

/// avg returns the mean, as numeric for integers and numerics and as float8 for floats.
fn avg(values: &[Value], ret: u32) -> Result<Value> {
    if ret == FLOAT8 {
        let total: f64 = values.iter().map(float_of).sum();
        return Ok(Value::Float8(total / values.len() as f64));
    }
    let total = values.iter().fold(Numeric::zero(0), |total, v| total.add(&numeric_of(v)));
    Ok(Value::Numeric(total.div(&Numeric::from_i64(values.len() as i64))?))
}

/// variance computes a variance or standard deviation as Postgres does: by the Youngs-Cramer algorithm for floats,
/// and from the sums of the values and their squares for numerics.
fn variance(values: &[Value], kind: Kind, ret: u32) -> Result<Value> {
    let sample = matches!(kind, Kind::VarSamp | Kind::StddevSamp);
    let root = matches!(kind, Kind::StddevPop | Kind::StddevSamp);
    let n = values.len();
    if sample && n < 2 {
        return Ok(Value::Null);
    }
    if ret == FLOAT8 {
        let (mut count, mut sx, mut sxx) = (0.0f64, 0.0f64, 0.0f64);
        for v in values {
            let x = float_of(v);
            let previous = count;
            count += 1.0;
            sx += x;
            if previous > 0.0 {
                let tmp = x * count - sx;
                sxx += tmp * tmp / (count * previous);
            }
        }
        let var = if sample { sxx / (count - 1.0) } else { sxx / count };
        return Ok(Value::Float8(if root { var.sqrt() } else { var }));
    }
    let numbers: Vec<Numeric> = values.iter().map(numeric_of).collect();
    let sum_x = numbers.iter().fold(Numeric::zero(0), |t, x| t.add(x));
    let sum_x2 = numbers.iter().fold(Numeric::zero(0), |t, x| t.add(&x.mul(x)));
    let count = Numeric::from_i64(n as i64);
    let numerator = count.mul(&sum_x2).sub(&sum_x.mul(&sum_x));
    if numerator.cmp_numeric(&Numeric::zero(0)) != Ordering::Greater {
        return Ok(Value::Numeric(Numeric::zero(0)));
    }
    let denominator = if sample { count.mul(&Numeric::from_i64(n as i64 - 1)) } else { count.mul(&count) };
    let var = numerator.div(&denominator)?;
    Ok(Value::Numeric(if root { var.sqrt(var.scale()) } else { var }))
}
