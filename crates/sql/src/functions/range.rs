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

//! Range and multirange functions, including the constructors and the functions behind the range operators.

use super::{ANY, ANYELEMENT, Function};
use crate::error::{PgError, Result, code};
use crate::oid::{BOOL, DATE, INT4, INT8, NUMERIC, TEXT, TIMESTAMP, TIMESTAMPTZ};
use crate::query::Ctx;
use crate::rangetypes::{self as rt, ANYMULTIRANGE as M, ANYRANGE as R, Bound, Multirange, Range};
use crate::types::Value;

/// f declares a strict range function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// c declares a range constructor, which reads a NULL bound as infinite.
const fn c(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: false, variadic: false, implementation }
}

/// v declares a multirange constructor that takes any number of ranges, including none.
const fn v(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: true, implementation }
}

/// FUNCTIONS are the range and multirange functions.
pub const FUNCTIONS: &[Function] = &[
    c("int4range", &[INT4, INT4], 3904, |_, args| construct(3904, args)),
    c("int4range", &[INT4, INT4, TEXT], 3904, |_, args| construct(3904, args)),
    c("int8range", &[INT8, INT8], 3926, |_, args| construct(3926, args)),
    c("int8range", &[INT8, INT8, TEXT], 3926, |_, args| construct(3926, args)),
    c("numrange", &[NUMERIC, NUMERIC], 3906, |_, args| construct(3906, args)),
    c("numrange", &[NUMERIC, NUMERIC, TEXT], 3906, |_, args| construct(3906, args)),
    c("tsrange", &[TIMESTAMP, TIMESTAMP], 3908, |_, args| construct(3908, args)),
    c("tsrange", &[TIMESTAMP, TIMESTAMP, TEXT], 3908, |_, args| construct(3908, args)),
    c("tstzrange", &[TIMESTAMPTZ, TIMESTAMPTZ], 3910, |_, args| construct(3910, args)),
    c("tstzrange", &[TIMESTAMPTZ, TIMESTAMPTZ, TEXT], 3910, |_, args| construct(3910, args)),
    c("daterange", &[DATE, DATE], 3912, |_, args| construct(3912, args)),
    c("daterange", &[DATE, DATE, TEXT], 3912, |_, args| construct(3912, args)),
    v("int4multirange", &[3904], 4451, |_, args| construct_multirange(4451, args)),
    v("int8multirange", &[3926], 4536, |_, args| construct_multirange(4536, args)),
    v("nummultirange", &[3906], 4532, |_, args| construct_multirange(4532, args)),
    v("tsmultirange", &[3908], 4533, |_, args| construct_multirange(4533, args)),
    v("tstzmultirange", &[3910], 4534, |_, args| construct_multirange(4534, args)),
    v("datemultirange", &[3912], 4535, |_, args| construct_multirange(4535, args)),
    c("__doltgres_range", &[INT8, ANYELEMENT, ANYELEMENT], R, |_, args| construct(oid_of(&args[0]), &args[1..])),
    c("__doltgres_range", &[INT8, ANYELEMENT, ANYELEMENT, TEXT], R, |_, args| construct(oid_of(&args[0]), &args[1..])),
    f("__doltgres_multirange", &[INT8], M, |_, args| construct_multirange(oid_of(&args[0]), &[])),
    v("__doltgres_multirange", &[INT8, ANY], M, |_, args| construct_multirange(oid_of(&args[0]), &args[1..])),
    f("multirange", &[R], M, to_multirange),
    f("lower", &[R], ANYELEMENT, lower),
    f("upper", &[R], ANYELEMENT, upper),
    f("isempty", &[R], BOOL, is_empty),
    f("lower_inc", &[R], BOOL, lower_inc),
    f("upper_inc", &[R], BOOL, upper_inc),
    f("lower_inf", &[R], BOOL, lower_inf),
    f("upper_inf", &[R], BOOL, upper_inf),
    f("lower", &[M], ANYELEMENT, lower),
    f("upper", &[M], ANYELEMENT, upper),
    f("isempty", &[M], BOOL, is_empty),
    f("lower_inc", &[M], BOOL, lower_inc),
    f("upper_inc", &[M], BOOL, upper_inc),
    f("lower_inf", &[M], BOOL, lower_inf),
    f("upper_inf", &[M], BOOL, upper_inf),
    f("range_merge", &[R, R], R, range_merge),
    f("range_merge", &[M], R, range_merge),
    f("unnest", &[M], R, unnest),
    f("@>", &[R, R], BOOL, contains),
    f("@>", &[R, ANYELEMENT], BOOL, contains),
    f("@>", &[R, M], BOOL, contains),
    f("@>", &[M, M], BOOL, contains),
    f("@>", &[M, R], BOOL, contains),
    f("@>", &[M, ANYELEMENT], BOOL, contains),
    f("<@", &[R, R], BOOL, contained),
    f("<@", &[ANYELEMENT, R], BOOL, contained),
    f("<@", &[M, R], BOOL, contained),
    f("<@", &[M, M], BOOL, contained),
    f("<@", &[R, M], BOOL, contained),
    f("<@", &[ANYELEMENT, M], BOOL, contained),
    f("&&", &[R, R], BOOL, overlaps),
    f("&&", &[R, M], BOOL, overlaps),
    f("&&", &[M, R], BOOL, overlaps),
    f("&&", &[M, M], BOOL, overlaps),
    f("<<", &[R, R], BOOL, before),
    f("<<", &[R, M], BOOL, before),
    f("<<", &[M, R], BOOL, before),
    f("<<", &[M, M], BOOL, before),
    f(">>", &[R, R], BOOL, after),
    f(">>", &[R, M], BOOL, after),
    f(">>", &[M, R], BOOL, after),
    f(">>", &[M, M], BOOL, after),
    f("&<", &[R, R], BOOL, over_left),
    f("&<", &[R, M], BOOL, over_left),
    f("&<", &[M, R], BOOL, over_left),
    f("&<", &[M, M], BOOL, over_left),
    f("&>", &[R, R], BOOL, over_right),
    f("&>", &[R, M], BOOL, over_right),
    f("&>", &[M, R], BOOL, over_right),
    f("&>", &[M, M], BOOL, over_right),
    f("-|-", &[R, R], BOOL, adjacent),
    f("-|-", &[R, M], BOOL, adjacent),
    f("-|-", &[M, R], BOOL, adjacent),
    f("-|-", &[M, M], BOOL, adjacent),
    f("+", &[R, R], R, union),
    f("+", &[M, M], M, union),
    f("-", &[R, R], R, difference),
    f("-", &[M, M], M, difference),
    f("*", &[R, R], R, intersection),
    f("*", &[M, M], M, intersection),
    f("elem_contained_by_multirange", &[ANYELEMENT, M], BOOL, contained),
    f("elem_contained_by_range", &[ANYELEMENT, R], BOOL, contained),
    f("multirange_adjacent_multirange", &[M, M], BOOL, adjacent),
    f("multirange_adjacent_range", &[M, R], BOOL, adjacent),
    f("multirange_after_multirange", &[M, M], BOOL, after),
    f("multirange_after_range", &[M, R], BOOL, after),
    f("multirange_before_multirange", &[M, M], BOOL, before),
    f("multirange_before_range", &[M, R], BOOL, before),
    f("multirange_contained_by_multirange", &[M, M], BOOL, contained),
    f("multirange_contained_by_range", &[M, R], BOOL, contained),
    f("multirange_contains_elem", &[M, ANYELEMENT], BOOL, contains),
    f("multirange_contains_multirange", &[M, M], BOOL, contains),
    f("multirange_contains_range", &[M, R], BOOL, contains),
    f("multirange_eq", &[M, M], BOOL, equal),
    f("multirange_ge", &[M, M], BOOL, greater_equal),
    f("multirange_gt", &[M, M], BOOL, greater),
    f("multirange_intersect", &[M, M], M, intersection),
    f("multirange_le", &[M, M], BOOL, less_equal),
    f("multirange_lt", &[M, M], BOOL, less),
    f("multirange_minus", &[M, M], M, difference),
    f("multirange_ne", &[M, M], BOOL, not_equal),
    f("multirange_overlaps_multirange", &[M, M], BOOL, overlaps),
    f("multirange_overlaps_range", &[M, R], BOOL, overlaps),
    f("multirange_overleft_multirange", &[M, M], BOOL, over_left),
    f("multirange_overleft_range", &[M, R], BOOL, over_left),
    f("multirange_overright_multirange", &[M, M], BOOL, over_right),
    f("multirange_overright_range", &[M, R], BOOL, over_right),
    f("multirange_union", &[M, M], M, union),
    f("range_adjacent", &[R, R], BOOL, adjacent),
    f("range_adjacent_multirange", &[R, M], BOOL, adjacent),
    f("range_after", &[R, R], BOOL, after),
    f("range_after_multirange", &[R, M], BOOL, after),
    f("range_before", &[R, R], BOOL, before),
    f("range_before_multirange", &[R, M], BOOL, before),
    f("range_contained_by", &[R, R], BOOL, contained),
    f("range_contained_by_multirange", &[R, M], BOOL, contained),
    f("range_contains", &[R, R], BOOL, contains),
    f("range_contains_elem", &[R, ANYELEMENT], BOOL, contains),
    f("range_contains_multirange", &[R, M], BOOL, contains),
    f("range_eq", &[R, R], BOOL, equal),
    f("range_ge", &[R, R], BOOL, greater_equal),
    f("range_gt", &[R, R], BOOL, greater),
    f("range_intersect", &[R, R], R, intersection),
    f("range_le", &[R, R], BOOL, less_equal),
    f("range_lt", &[R, R], BOOL, less),
    f("range_minus", &[R, R], R, difference),
    f("range_ne", &[R, R], BOOL, not_equal),
    f("range_overlaps", &[R, R], BOOL, overlaps),
    f("range_overlaps_multirange", &[R, M], BOOL, overlaps),
    f("range_overleft", &[R, R], BOOL, over_left),
    f("range_overleft_multirange", &[R, M], BOOL, over_left),
    f("range_overright", &[R, R], BOOL, over_right),
    f("range_overright_multirange", &[R, M], BOOL, over_right),
    f("range_union", &[R, R], R, union),
    f("range_cmp", &[R, R], INT4, compare),
    f("multirange_cmp", &[M, M], INT4, compare),
];

/// construct builds a range of a type from its bounds, which are infinite when NULL, and its bound flags, which default to `[)`.
fn construct(type_oid: u32, args: &[Value]) -> Result<Value> {
    let flags = match args.get(2) {
        None => "[)".to_string(),
        Some(Value::Null) => {
            return Err(PgError::new(
                code::NULL_VALUE_NOT_ALLOWED,
                "range constructor flags argument must not be null",
            ));
        }
        Some(flags) => flags.output().unwrap_or_default(),
    };
    let (lower_inclusive, upper_inclusive) = match flags.as_str() {
        "[]" => (true, true),
        "[)" => (true, false),
        "(]" => (false, true),
        "()" => (false, false),
        _ => {
            return Err(PgError {
                hint: Some("Valid values are \"[]\", \"[)\", \"(]\", and \"()\".".into()),
                ..PgError::new(code::SYNTAX_ERROR, "invalid range bound flags")
            });
        }
    };
    let bound = |value: &Value, inclusive: bool| Bound { value: (!value.is_null()).then(|| value.clone()), inclusive };
    let range = rt::make(type_oid, bound(&args[0], lower_inclusive), bound(&args[1], upper_inclusive))?;
    Ok(Value::Range(Box::new(range)))
}

/// oid_of returns the type OID that a user-defined range type's constructor passes first.
fn oid_of(value: &Value) -> u32 {
    match value {
        Value::Int8(oid) => *oid as u32,
        _ => 0,
    }
}

/// construct_multirange builds a multirange of a type from no ranges, one range, or an array of ranges.
fn construct_multirange(type_oid: u32, args: &[Value]) -> Result<Value> {
    let mut ranges = Vec::new();
    for arg in args {
        let items = match arg {
            Value::Array(array) => array.values.clone(),
            other => vec![other.clone()],
        };
        for item in items {
            match item {
                Value::Range(range) => ranges.push(*range),
                _ => {
                    return Err(PgError::new(
                        code::NULL_VALUE_NOT_ALLOWED,
                        "multirange values cannot contain null members",
                    ));
                }
            }
        }
    }
    Ok(Value::Multirange(Box::new(rt::normalize(type_oid, ranges)?)))
}

/// as_range returns a range argument, or the span of a multirange argument.
fn as_range(value: &Value) -> Range {
    match value {
        Value::Range(range) => (**range).clone(),
        Value::Multirange(multirange) => rt::span(multirange),
        _ => rt::empty(0),
    }
}

/// as_multirange returns a multirange argument, or a range argument as a multirange of it.
fn as_multirange(value: &Value) -> Result<Multirange> {
    match value {
        Value::Multirange(multirange) => Ok((**multirange).clone()),
        Value::Range(range) => {
            let multirange_oid = rt::range_type(range.type_oid).map_or(0, |t| t.multirange);
            rt::normalize(multirange_oid, vec![(**range).clone()])
        }
        _ => Ok(Multirange { type_oid: 0, ranges: Vec::new() }),
    }
}

/// to_multirange returns a multirange of one range.
fn to_multirange(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Multirange(Box::new(as_multirange(&args[0])?)))
}

/// lower returns a range's lower bound, which is NULL when it is empty or infinite.
fn lower(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let range = as_range(&args[0]);
    Ok(if range.empty { Value::Null } else { range.lower.value.unwrap_or(Value::Null) })
}

/// upper returns a range's upper bound, which is NULL when it is empty or infinite.
fn upper(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let range = as_range(&args[0]);
    Ok(if range.empty { Value::Null } else { range.upper.value.unwrap_or(Value::Null) })
}

/// is_empty reports whether a range or multirange holds no values.
fn is_empty(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(as_range(&args[0]).empty))
}

/// lower_inc reports whether a range includes its lower bound.
fn lower_inc(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let range = as_range(&args[0]);
    Ok(Value::Bool(!range.empty && range.lower.inclusive))
}

/// upper_inc reports whether a range includes its upper bound.
fn upper_inc(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let range = as_range(&args[0]);
    Ok(Value::Bool(!range.empty && range.upper.inclusive))
}

/// lower_inf reports whether a range has no lower bound.
fn lower_inf(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let range = as_range(&args[0]);
    Ok(Value::Bool(!range.empty && range.lower.value.is_none()))
}

/// upper_inf reports whether a range has no upper bound.
fn upper_inf(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let range = as_range(&args[0]);
    Ok(Value::Bool(!range.empty && range.upper.value.is_none()))
}

/// range_merge returns the smallest range that covers two ranges or a multirange.
fn range_merge(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let merged = match args {
        [left, right] => rt::merge(&as_range(left), &as_range(right), false)?,
        [multirange] => as_range(multirange),
        _ => return Err(PgError::internal("range_merge takes one or two arguments")),
    };
    Ok(Value::Range(Box::new(merged)))
}

/// unnest returns a multirange's ranges as rows.
fn unnest(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let ranges = as_multirange(&args[0])?.ranges;
    Ok(Value::Set(ranges.into_iter().map(|r| Value::Range(Box::new(r))).collect()))
}

/// contains reports whether the left argument contains the right one: a range, a multirange, or a value.
fn contains(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(contains_value(&args[0], &args[1])?))
}

/// contained reports whether the right argument contains the left one.
fn contained(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(contains_value(&args[1], &args[0])?))
}

/// contains_value reports whether a range or multirange contains a range, a multirange, or a value.
fn contains_value(outer: &Value, inner: &Value) -> Result<bool> {
    Ok(match (outer, inner) {
        (Value::Range(outer), Value::Range(inner)) => rt::contains(outer, inner),
        (Value::Range(outer), Value::Multirange(inner)) => rt::contains(outer, &rt::span(inner)),
        (Value::Multirange(outer), Value::Range(inner)) => {
            inner.empty || outer.ranges.iter().any(|r| rt::contains(r, inner))
        }
        (Value::Multirange(outer), Value::Multirange(inner)) => {
            inner.ranges.iter().all(|i| outer.ranges.iter().any(|r| rt::contains(r, i)))
        }
        (Value::Range(outer), element) => rt::contains_element(outer, element),
        (Value::Multirange(outer), element) => outer.ranges.iter().any(|r| rt::contains_element(r, element)),
        _ => false,
    })
}

/// pairs returns the ranges of two arguments, each a range or a multirange, for comparing them range by range.
fn pairs(args: &[Value]) -> Result<(Multirange, Multirange)> {
    Ok((as_multirange(&args[0])?, as_multirange(&args[1])?))
}

/// ordering compares two ranges or two multiranges.
fn ordering(args: &[Value]) -> std::cmp::Ordering {
    crate::expr::compare_values(&args[0], &args[1])
}

/// compare returns -1, 0, or 1 as the left range or multirange sorts before, with, or after the right one.
fn compare(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(ordering(args) as i32))
}

/// less reports whether the left argument sorts before the right one.
fn less(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(ordering(args).is_lt()))
}

/// less_equal reports whether the left argument sorts no later than the right one.
fn less_equal(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(ordering(args).is_le()))
}

/// equal reports whether two ranges or multiranges hold the same values.
fn equal(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(ordering(args).is_eq()))
}

/// not_equal reports whether two ranges or multiranges hold different values.
fn not_equal(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(ordering(args).is_ne()))
}

/// greater reports whether the left argument sorts after the right one.
fn greater(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(ordering(args).is_gt()))
}

/// greater_equal reports whether the left argument sorts no earlier than the right one.
fn greater_equal(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(ordering(args).is_ge()))
}

/// overlaps reports whether two ranges or multiranges share a value.
fn overlaps(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (left, right) = pairs(args)?;
    Ok(Value::Bool(left.ranges.iter().any(|l| right.ranges.iter().any(|r| rt::overlaps(l, r)))))
}

/// before reports whether the left argument ends before the right one starts.
fn before(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(rt::before(&as_range(&args[0]), &as_range(&args[1]))))
}

/// after reports whether the left argument starts after the right one ends.
fn after(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(rt::after(&as_range(&args[0]), &as_range(&args[1]))))
}

/// over_left reports whether the left argument ends no later than the right one.
fn over_left(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(rt::over_left(&as_range(&args[0]), &as_range(&args[1]))))
}

/// over_right reports whether the left argument starts no earlier than the right one.
fn over_right(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(rt::over_right(&as_range(&args[0]), &as_range(&args[1]))))
}

/// adjacent reports whether two ranges or multiranges meet with no value between them.
fn adjacent(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (left, right) = pairs(args)?;
    let adjacent = match (left.ranges.first(), left.ranges.last(), right.ranges.first(), right.ranges.last()) {
        (Some(left_first), Some(left_last), Some(right_first), Some(right_last)) => {
            rt::bounds_adjacent(left_first.type_oid, &right_last.upper, &left_first.lower)
                || rt::bounds_adjacent(left_first.type_oid, &left_last.upper, &right_first.lower)
        }
        _ => false,
    };
    Ok(Value::Bool(adjacent))
}

/// union returns the values of two ranges, which must overlap or meet, or of two multiranges.
fn union(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(match (&args[0], &args[1]) {
        (Value::Range(left), Value::Range(right)) => Value::Range(Box::new(rt::merge(left, right, true)?)),
        _ => {
            let (left, right) = pairs(args)?;
            let ranges = left.ranges.into_iter().chain(right.ranges).collect();
            Value::Multirange(Box::new(rt::normalize(left.type_oid, ranges)?))
        }
    })
}

/// difference returns the values of the left argument that the right one leaves.
fn difference(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(match (&args[0], &args[1]) {
        (Value::Range(left), Value::Range(right)) => Value::Range(Box::new(rt::minus(left, right)?)),
        _ => {
            let (left, right) = pairs(args)?;
            Value::Multirange(Box::new(rt::multirange_minus(&left, &right)?))
        }
    })
}

/// intersection returns the values two ranges or multiranges share.
fn intersection(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(match (&args[0], &args[1]) {
        (Value::Range(left), Value::Range(right)) => Value::Range(Box::new(rt::intersect(left, right)?)),
        _ => {
            let (left, right) = pairs(args)?;
            Value::Multirange(Box::new(rt::multirange_intersect(&left, &right)?))
        }
    })
}
