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

use crate::cast::type_display;
use crate::error::{PgError, Result, code};
use crate::expr::{Expr, compare_values, position};
use crate::numeric::Numeric;
use crate::oid::{BOOL, FLOAT4, FLOAT8, INT2, INT4, INT8, INTERVAL, NUMERIC, TEXT};
use crate::query::Ctx;
use crate::types::Value;

use super::{ANYARRAY, ANYELEMENT, ANYNONARRAY, implicitly_castable};
use crate::rangetypes::{ANYMULTIRANGE, ANYRANGE};

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
    ArrayAgg,
    JsonAgg,
    JsonbAgg,
    JsonObjectAgg,
    JsonbObjectAgg,
    XmlAgg,
    RangeAgg,
    RangeIntersectAgg,
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
    a("sum", &[INTERVAL], INTERVAL, Kind::Sum),
    a("avg", &[INT2], NUMERIC, Kind::Avg),
    a("avg", &[INT4], NUMERIC, Kind::Avg),
    a("avg", &[INT8], NUMERIC, Kind::Avg),
    a("avg", &[NUMERIC], NUMERIC, Kind::Avg),
    a("avg", &[FLOAT4], FLOAT8, Kind::Avg),
    a("avg", &[FLOAT8], FLOAT8, Kind::Avg),
    a("avg", &[INTERVAL], INTERVAL, Kind::Avg),
    a("min", &[ANYELEMENT], ANYELEMENT, Kind::Min),
    a("max", &[ANYELEMENT], ANYELEMENT, Kind::Max),
    a("bool_and", &[BOOL], BOOL, Kind::BoolAnd),
    a("every", &[BOOL], BOOL, Kind::BoolAnd),
    a("bool_or", &[BOOL], BOOL, Kind::BoolOr),
    a("string_agg", &[TEXT, TEXT], TEXT, Kind::StringAgg),
    a("xmlagg", &[crate::oid::XML], crate::oid::XML, Kind::XmlAgg),
    a("var_pop", &[INT2], NUMERIC, Kind::VarPop),
    a("var_pop", &[INT4], NUMERIC, Kind::VarPop),
    a("var_pop", &[INT8], NUMERIC, Kind::VarPop),
    a("var_pop", &[NUMERIC], NUMERIC, Kind::VarPop),
    a("var_pop", &[FLOAT4], FLOAT8, Kind::VarPop),
    a("var_pop", &[FLOAT8], FLOAT8, Kind::VarPop),
    a("var_samp", &[INT2], NUMERIC, Kind::VarSamp),
    a("var_samp", &[INT4], NUMERIC, Kind::VarSamp),
    a("var_samp", &[INT8], NUMERIC, Kind::VarSamp),
    a("var_samp", &[NUMERIC], NUMERIC, Kind::VarSamp),
    a("var_samp", &[FLOAT4], FLOAT8, Kind::VarSamp),
    a("var_samp", &[FLOAT8], FLOAT8, Kind::VarSamp),
    a("variance", &[INT2], NUMERIC, Kind::VarSamp),
    a("variance", &[INT4], NUMERIC, Kind::VarSamp),
    a("variance", &[INT8], NUMERIC, Kind::VarSamp),
    a("variance", &[NUMERIC], NUMERIC, Kind::VarSamp),
    a("variance", &[FLOAT4], FLOAT8, Kind::VarSamp),
    a("variance", &[FLOAT8], FLOAT8, Kind::VarSamp),
    a("stddev_pop", &[INT2], NUMERIC, Kind::StddevPop),
    a("stddev_pop", &[INT4], NUMERIC, Kind::StddevPop),
    a("stddev_pop", &[INT8], NUMERIC, Kind::StddevPop),
    a("stddev_pop", &[NUMERIC], NUMERIC, Kind::StddevPop),
    a("stddev_pop", &[FLOAT4], FLOAT8, Kind::StddevPop),
    a("stddev_pop", &[FLOAT8], FLOAT8, Kind::StddevPop),
    a("stddev_samp", &[INT2], NUMERIC, Kind::StddevSamp),
    a("stddev_samp", &[INT4], NUMERIC, Kind::StddevSamp),
    a("stddev_samp", &[INT8], NUMERIC, Kind::StddevSamp),
    a("stddev_samp", &[NUMERIC], NUMERIC, Kind::StddevSamp),
    a("stddev_samp", &[FLOAT4], FLOAT8, Kind::StddevSamp),
    a("stddev_samp", &[FLOAT8], FLOAT8, Kind::StddevSamp),
    a("stddev", &[INT2], NUMERIC, Kind::StddevSamp),
    a("stddev", &[INT4], NUMERIC, Kind::StddevSamp),
    a("stddev", &[INT8], NUMERIC, Kind::StddevSamp),
    a("stddev", &[NUMERIC], NUMERIC, Kind::StddevSamp),
    a("stddev", &[FLOAT4], FLOAT8, Kind::StddevSamp),
    a("stddev", &[FLOAT8], FLOAT8, Kind::StddevSamp),
    a("array_agg", &[ANYNONARRAY], ANYARRAY, Kind::ArrayAgg),
    a("array_agg", &[ANYARRAY], ANYARRAY, Kind::ArrayAgg),
    a("json_agg", &[ANYELEMENT], crate::oid::JSON, Kind::JsonAgg),
    a("jsonb_agg", &[ANYELEMENT], crate::oid::JSONB, Kind::JsonbAgg),
    a("range_agg", &[ANYRANGE], ANYMULTIRANGE, Kind::RangeAgg),
    a("range_agg", &[ANYMULTIRANGE], ANYMULTIRANGE, Kind::RangeAgg),
    a("range_intersect_agg", &[ANYRANGE], ANYRANGE, Kind::RangeIntersectAgg),
    a("range_intersect_agg", &[ANYMULTIRANGE], ANYMULTIRANGE, Kind::RangeIntersectAgg),
    a("json_object_agg", &[super::ANY, super::ANY], crate::oid::JSON, Kind::JsonObjectAgg),
    a("jsonb_object_agg", &[super::ANY, super::ANY], crate::oid::JSONB, Kind::JsonbObjectAgg),
];

/// exists reports whether a call names an aggregate, built in or user-defined.
pub fn exists(schema: Option<&str>, name: &str) -> bool {
    schema.is_none_or(|s| s == "pg_catalog") && AGGREGATES.iter().any(|a| a.name == name)
        || crate::aggregates::exists(schema, name)
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
    /// The stored aggregate that the call runs instead of the built-in one at `index`.
    pub user: Option<std::sync::Arc<crate::aggregates::UserAggregate>>,
}

impl AggCall {
    /// counts_set_column reports whether the call is a plain COUNT of a column of the table that is NOT NULL, which
    /// counts every row as COUNT(*) does.
    pub fn counts_set_column(&self, table: &crate::catalog::table::TableDef) -> bool {
        AGGREGATES[self.index].kind == Kind::Count
            && matches!(self.args.as_slice(), [Expr::Column(c)] if table.columns.get(*c).is_some_and(|c| !c.nullable))
            && self.filter.is_none()
            && !self.distinct
            && self.order.is_empty()
            && self.user.is_none()
    }

    /// counts_rows reports whether the call is a plain COUNT(*), which counts every row of its group.
    pub fn counts_rows(&self) -> bool {
        AGGREGATES[self.index].kind == Kind::CountStar
            && self.filter.is_none()
            && !self.distinct
            && self.order.is_empty()
            && self.user.is_none()
    }
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
        .map(|(&p, &t)| {
            if matches!(p, ANYELEMENT | ANYNONARRAY | ANYARRAY | ANYRANGE | ANYMULTIRANGE | super::ANY) {
                if t == crate::oid::UNKNOWN { TEXT } else { t }
            } else {
                p
            }
        })
        .collect();
    let ret = match aggregate.ret {
        ANYELEMENT => element,
        ANYARRAY if aggregate.args == [ANYARRAY] => element,
        ANYARRAY => crate::expr::array_of(element),
        ANYRANGE => element,
        ANYMULTIRANGE => crate::rangetypes::range_type(element).map_or(element, |r| r.multirange),
        ret => ret,
    };
    Ok((index, arg_types, ret))
}

/// Accumulator collects a group's input for one aggregate call: a running result for the common aggregates, and the
/// argument rows for the others.
pub struct Accumulator {
    state: State,
}

/// State is what an accumulator keeps of the rows it has seen.
#[derive(Clone)]
enum State {
    /// The argument rows and the ORDER BY key values of each row, with the distinct argument rows so far for a
    /// DISTINCT aggregate, which keeps only the first of equal rows.
    Rows { rows: Vec<Vec<Value>>, keys: Vec<Vec<Value>>, seen: Option<crate::exec::Groups> },
    /// The number of rows, or of non-NULL values for count of a value.
    Count(i64),
    /// The running sum of integers into a bigint and their count, or None before the first value.
    SumInt(Option<(i64, i64)>),
    /// The running sum and count of floats, or None before the first value.
    SumFloat(Option<(f64, i64)>),
    /// The running float4 sum, or None before the first value.
    SumFloat4(Option<f32>),
    /// The running numeric sum, the count, and how many of the values have the sum's scale, or None before the first
    /// value.
    SumNumeric(Option<(Numeric, i64, i64)>),
    /// The least or greatest value so far.
    Extreme(Option<Value>),
    /// Whether every or any value so far was true, or None before the first value.
    Bool(Option<bool>),
}

impl Accumulator {
    /// new starts an empty accumulator for the call.
    pub fn new(call: &AggCall) -> Accumulator {
        let rows =
            || State::Rows { rows: Vec::new(), keys: Vec::new(), seen: call.distinct.then(crate::exec::Groups::new) };
        if call.user.is_some() || call.distinct || !call.order.is_empty() {
            return Accumulator { state: rows() };
        }
        let state = match (AGGREGATES[call.index].kind, call.ret) {
            (Kind::CountStar | Kind::Count, _) => State::Count(0),
            (Kind::Sum, INT8) => State::SumInt(None),
            (Kind::Sum, FLOAT4) => State::SumFloat4(None),
            (Kind::Sum | Kind::Avg, FLOAT8) => State::SumFloat(None),
            (Kind::Sum | Kind::Avg, INTERVAL) => rows(),
            (Kind::Sum | Kind::Avg, _) => State::SumNumeric(None),
            (Kind::Min | Kind::Max, _) => State::Extreme(None),
            (Kind::BoolAnd | Kind::BoolOr, _) => State::Bool(None),
            _ => rows(),
        };
        Accumulator { state }
    }

    /// add adds an input row to the group.
    pub fn add(&mut self, ctx: &mut Ctx<'_>, call: &AggCall, row: &[Value]) -> Result<()> {
        if let Some(filter) = &call.filter
            && !filter.is_true(ctx, row)?
        {
            return Ok(());
        }
        let kind = AGGREGATES[call.index].kind;
        let value = match &mut self.state {
            State::Rows { rows, keys, seen } => {
                let args = call.args.iter().map(|a| a.eval(ctx, row)).collect::<Result<Vec<_>>>()?;
                if seen.as_mut().is_some_and(|seen| !seen.insert(&args).1) {
                    return Ok(());
                }
                rows.push(args);
                keys.push(call.order.iter().map(|k| k.0.eval(ctx, row)).collect::<Result<Vec<_>>>()?);
                return Ok(());
            }
            State::Count(n) if kind == Kind::CountStar => {
                *n += 1;
                return Ok(());
            }
            State::Count(n) if matches!(call.args.as_slice(), [Expr::Column(_)]) => {
                if let [Expr::Column(i)] = call.args.as_slice()
                    && !row[*i].is_null()
                {
                    *n += 1;
                }
                return Ok(());
            }
            _ => match call.args.first() {
                Some(arg) => arg.eval(ctx, row)?,
                None => Value::Null,
            },
        };
        if value.is_null() {
            return Ok(());
        }
        match &mut self.state {
            State::Count(n) => *n += 1,
            State::SumInt(total) => {
                let (sum, count) = total.get_or_insert((0, 0));
                *sum = sum
                    .checked_add(int_of(&value))
                    .ok_or_else(|| PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "bigint out of range"))?;
                *count += 1;
            }
            State::SumFloat(total) => {
                let (sum, count) = total.get_or_insert((std::iter::empty::<f64>().sum(), 0));
                *sum += float_of(&value);
                *count += 1;
            }
            State::SumFloat4(total) => *total.get_or_insert(std::iter::empty::<f32>().sum()) += float_of(&value) as f32,
            State::SumNumeric(total) => {
                let (sum, count, top) = total.get_or_insert((Numeric::zero(0), 0, 0));
                let value = numeric_of(&value);
                match value.scale().cmp(&sum.scale()) {
                    Ordering::Greater => *top = 1,
                    Ordering::Equal => *top += 1,
                    Ordering::Less => {}
                }
                *sum = sum.add(&value);
                *count += 1;
            }
            State::Extreme(extreme) => {
                let wanted = if kind == Kind::Min { Ordering::Less } else { Ordering::Greater };
                if extreme.as_ref().is_none_or(|e| compare_values(&value, e) == wanted) {
                    *extreme = Some(value);
                }
            }
            State::Bool(result) => {
                let truth = value == Value::Bool(true);
                *result = Some(match (kind, *result) {
                    (Kind::BoolAnd, previous) => previous.unwrap_or(true) && truth,
                    (_, previous) => previous.unwrap_or(false) || truth,
                });
            }
            State::Rows { .. } => unreachable!("handled above"),
        }
        Ok(())
    }

    /// invertible reports whether `remove` can take rows back out of the accumulator.
    pub fn invertible(&self) -> bool {
        matches!(self.state, State::Count(_) | State::SumInt(_) | State::SumNumeric(_))
    }

    /// remove takes a row that `add` added back out of the group, as Postgres' inverse transition functions do,
    /// reporting false when it can't, which leaves the aggregate to be computed again.
    pub fn remove(&mut self, ctx: &mut Ctx<'_>, call: &AggCall, row: &[Value]) -> Result<bool> {
        if let Some(filter) = &call.filter
            && !filter.is_true(ctx, row)?
        {
            return Ok(true);
        }
        if !self.invertible() {
            return Ok(false);
        }
        if let State::Count(n) = &mut self.state
            && AGGREGATES[call.index].kind == Kind::CountStar
        {
            *n -= 1;
            return Ok(true);
        }
        let value = match call.args.first() {
            Some(arg) => arg.eval(ctx, row)?,
            None => Value::Null,
        };
        if value.is_null() {
            return Ok(true);
        }
        match &mut self.state {
            State::Count(n) => *n -= 1,
            State::SumInt(total) => {
                let Some((sum, count)) = total else { return Ok(false) };
                *sum -= int_of(&value);
                *count -= 1;
                if *count == 0 {
                    *total = None;
                }
            }
            State::SumNumeric(total) => {
                let Some((sum, count, top)) = total else { return Ok(false) };
                let value = numeric_of(&value);
                if !matches!(sum, Numeric::Finite { .. }) || !matches!(value, Numeric::Finite { .. }) {
                    return Ok(false);
                }
                if value.scale() == sum.scale() {
                    if *top <= 1 {
                        return Ok(false);
                    }
                    *top -= 1;
                }
                *sum = sum.sub(&value);
                *count -= 1;
                if *count == 0 {
                    *total = None;
                }
            }
            _ => return Ok(false),
        }
        Ok(true)
    }

    /// peek computes the aggregate over the rows added so far, keeping them for more.
    pub fn peek(&self, ctx: &mut Ctx<'_>, call: &AggCall) -> Result<Value> {
        Accumulator { state: self.state.clone() }.finish(ctx, call)
    }

    /// finish computes the aggregate over the group.
    pub fn finish(self, ctx: &mut Ctx<'_>, call: &AggCall) -> Result<Value> {
        let aggregate = &AGGREGATES[call.index];
        let (rows, keys) = match self.state {
            State::Rows { rows, keys, .. } => (rows, keys),
            State::Count(n) => return Ok(Value::Int8(n)),
            State::SumInt(total) => return Ok(total.map_or(Value::Null, |(sum, _)| Value::Int8(sum))),
            State::SumFloat(None) | State::SumFloat4(None) | State::SumNumeric(None) => return Ok(Value::Null),
            State::SumFloat(Some((sum, count))) if aggregate.kind == Kind::Avg => {
                return Ok(Value::Float8(sum / count as f64));
            }
            State::SumFloat(Some((sum, _))) => return Ok(Value::Float8(sum)),
            State::SumFloat4(Some(sum)) => return Ok(Value::Float4(sum)),
            State::SumNumeric(Some((sum, count, _))) if aggregate.kind == Kind::Avg => {
                return Ok(Value::Numeric(sum.div(&Numeric::from_i64(count))?));
            }
            State::SumNumeric(Some((sum, ..))) => return Ok(Value::Numeric(sum)),
            State::Extreme(extreme) => return Ok(extreme.unwrap_or(Value::Null)),
            State::Bool(result) => return Ok(result.map_or(Value::Null, Value::Bool)),
        };
        let mut rows: Vec<(Vec<Value>, Vec<Value>)> = keys.into_iter().zip(rows).collect();
        if !call.order.is_empty() || call.distinct {
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
                match call.distinct {
                    true => compare_rows(&a.1, &b.1),
                    false => Ordering::Equal,
                }
            });
        }
        let mut args: Vec<Vec<Value>> = rows.into_iter().map(|(_, args)| args).collect();
        if call.distinct {
            args.dedup_by(|a, b| compare_rows(a, b) == Ordering::Equal);
        }
        if let Some(user) = &call.user {
            return crate::aggregates::run(ctx, user, args);
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
            Kind::JsonAgg | Kind::JsonbAgg | Kind::JsonObjectAgg | Kind::JsonbObjectAgg => {
                return json_aggregate(aggregate.kind, args);
            }
            Kind::ArrayAgg => {
                let values: Vec<Value> = args.into_iter().filter_map(|r| r.into_iter().next()).collect();
                if values.is_empty() {
                    return Ok(Value::Null);
                }
                let element = crate::expr::element_type(call.ret);
                if aggregate.args[0] == ANYARRAY {
                    return array_agg_arrays(element, values);
                }
                return Ok(Value::Array(Box::new(crate::array::Array::one_dimensional(element, values))));
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
            Kind::XmlAgg => Ok(values
                .into_iter()
                .reduce(|a, b| match (a, b) {
                    (Value::Xml(a), Value::Xml(b)) => Value::Xml(crate::xml::concat(&[&a, &b])),
                    (a, _) => a,
                })
                .unwrap_or(Value::Null)),
            Kind::VarPop | Kind::VarSamp | Kind::StddevPop | Kind::StddevSamp => {
                variance(&values, aggregate.kind, call.ret)
            }
            Kind::RangeAgg => {
                let mut ranges = Vec::new();
                for value in values {
                    match value {
                        Value::Range(range) => ranges.push(*range),
                        Value::Multirange(multirange) => ranges.extend(multirange.ranges),
                        _ => {}
                    }
                }
                Ok(Value::Multirange(Box::new(crate::rangetypes::normalize(call.ret, ranges)?)))
            }
            Kind::RangeIntersectAgg => {
                let mut values = values.into_iter();
                let mut result = values.next().unwrap_or(Value::Null);
                for value in values {
                    result = match (result, value) {
                        (Value::Range(l), Value::Range(r)) => {
                            Value::Range(Box::new(crate::rangetypes::intersect(&l, &r)?))
                        }
                        (Value::Multirange(l), Value::Multirange(r)) => {
                            Value::Multirange(Box::new(crate::rangetypes::multirange_intersect(&l, &r)?))
                        }
                        (other, _) => other,
                    };
                }
                Ok(result)
            }
            Kind::CountStar
            | Kind::StringAgg
            | Kind::ArrayAgg
            | Kind::JsonAgg
            | Kind::JsonbAgg
            | Kind::JsonObjectAgg
            | Kind::JsonbObjectAgg => unreachable!("handled above"),
        }
    }
}

/// json_aggregate builds json_agg, jsonb_agg, json_object_agg, or jsonb_object_agg's result, which is NULL without
/// rows.
fn json_aggregate(kind: Kind, rows: Vec<Vec<Value>>) -> Result<Value> {
    use crate::functions::json::{datum_json_text, datum_to_json};
    use crate::json::{Json, escape, normalize};
    if rows.is_empty() {
        return Ok(Value::Null);
    }
    Ok(match kind {
        Kind::JsonAgg => {
            let mut out = String::from("[");
            for (i, row) in rows.iter().enumerate() {
                if i > 0 {
                    out.push_str(", ");
                    if matches!(row[0], Value::Array(_) | Value::Record(_) | Value::Composite(_)) {
                        out.push_str("\n ");
                    }
                }
                out.push_str(&datum_json_text(&row[0])?);
            }
            out.push(']');
            Value::Json(out)
        }
        Kind::JsonbAgg => {
            let values = rows.iter().map(|r| datum_to_json(&r[0])).collect::<Result<Vec<_>>>()?;
            Value::Jsonb(Box::new(normalize(Json::Array(values))))
        }
        _ => {
            let mut items = Vec::new();
            for row in &rows {
                let key = row[0]
                    .output()
                    .ok_or_else(|| PgError::new(code::NULL_VALUE_NOT_ALLOWED, "field name must not be null"))?;
                items.push((key, row[1].clone()));
            }
            if kind == Kind::JsonbObjectAgg {
                let items =
                    items.into_iter().map(|(k, v)| datum_to_json(&v).map(|j| (k, j))).collect::<Result<Vec<_>>>()?;
                Value::Jsonb(Box::new(normalize(Json::Object(items))))
            } else {
                let mut out = String::from("{ ");
                for (i, (key, value)) in items.iter().enumerate() {
                    if i > 0 {
                        out.push_str(", ");
                    }
                    escape(&mut out, key);
                    out.push_str(" : ");
                    out.push_str(&datum_json_text(value)?);
                }
                out.push_str(" }");
                Value::Json(out)
            }
        }
    })
}

/// array_agg_arrays stacks arrays of matching dimensions into an array with one more dimension, checking them in
/// order against the first as Postgres' accumArrayResultArr does.
pub(crate) fn array_agg_arrays(element: u32, values: Vec<Value>) -> Result<Value> {
    let error = |message: &str| PgError::new(code::ARRAY_SUBSCRIPT_ERROR, message);
    let mut first = None;
    for value in &values {
        let Value::Array(a) = value else {
            return Err(PgError::new(code::NULL_VALUE_NOT_ALLOWED, "cannot accumulate null arrays"));
        };
        match first {
            None if a.dims.is_empty() => return Err(error("cannot accumulate empty arrays")),
            None => first = Some(&a.dims),
            Some(dims) if *dims != a.dims => return Err(error("cannot accumulate arrays of different dimensionality")),
            Some(_) => {}
        }
    }
    crate::array::nest(element, values)
        .map(|a| Value::Array(Box::new(a)))
        .map_err(|_| error("cannot accumulate arrays of different dimensionality"))
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

/// int_of converts an integer value to i64.
fn int_of(value: &Value) -> i64 {
    match value {
        Value::Int2(i) => *i as i64,
        Value::Int4(i) => *i as i64,
        Value::Int8(i) => *i,
        _ => 0,
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
        INTERVAL => Value::Interval(interval_sum(values)?),
        _ => Value::Numeric(values.iter().fold(Numeric::zero(0), |total, v| total.add(&numeric_of(v)))),
    })
}

/// interval_sum adds intervals field by field, failing as Postgres' interval_pl does when a field overflows.
fn interval_sum(values: &[Value]) -> Result<crate::datetime::Interval> {
    let overflow = || PgError::new(code::DATETIME_FIELD_OVERFLOW, "interval out of range");
    let mut total = crate::datetime::Interval::default();
    for value in values {
        if let Value::Interval(iv) = value {
            total.months = total.months.checked_add(iv.months).ok_or_else(overflow)?;
            total.days = total.days.checked_add(iv.days).ok_or_else(overflow)?;
            total.micros = total.micros.checked_add(iv.micros).ok_or_else(overflow)?;
        }
    }
    Ok(total)
}

/// avg returns the mean, as numeric for integers and numerics, as float8 for floats, and as an interval for
/// intervals.
fn avg(values: &[Value], ret: u32) -> Result<Value> {
    if ret == INTERVAL {
        let total = interval_sum(values)?;
        return Ok(Value::Interval(super::datetime::interval_multiply(total, values.len() as f64, true)?));
    }
    if ret == FLOAT8 {
        let total: f64 = values.iter().map(float_of).sum();
        return Ok(Value::Float8(total / values.len() as f64));
    }
    let total = values.iter().fold(Numeric::zero(0), |total, v| total.add(&numeric_of(v)));
    Ok(Value::Numeric(total.div(&Numeric::from_i64(values.len() as i64))?))
}

/// compare_rows orders two rows of aggregate arguments as a DISTINCT aggregate sorts them: ascending, with NULLs last.
fn compare_rows(left: &[Value], right: &[Value]) -> Ordering {
    for (l, r) in left.iter().zip(right) {
        let ordering = match (l, r) {
            (Value::Null, Value::Null) => Ordering::Equal,
            (Value::Null, _) => Ordering::Greater,
            (_, Value::Null) => Ordering::Less,
            (l, r) => compare_values(l, r),
        };
        if ordering != Ordering::Equal {
            return ordering;
        }
    }
    Ordering::Equal
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
            let (previous, previous_sx) = (count, sx);
            count += 1.0;
            sx += x;
            if previous > 0.0 {
                let tmp = x * count - sx;
                sxx += tmp * tmp / (count * previous);
                if sx.is_infinite() || sxx.is_infinite() {
                    if !previous_sx.is_infinite() && !x.is_infinite() {
                        return Err(PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "value out of range: overflow"));
                    }
                    sxx = f64::NAN;
                }
            } else if x.is_infinite() || x.is_nan() {
                sxx = f64::NAN;
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
