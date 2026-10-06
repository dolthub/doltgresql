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

//! Array functions.

use std::cmp::Ordering;

use super::{ANYARRAY, ANYELEMENT, Function};
use crate::array::{self, Array};
use crate::error::{PgError, Result, code};
use crate::expr::compare_values;
use crate::oid::{BOOL, INT4, RECORD, TEXT, TEXT_ARRAY};
use crate::query::Ctx;
use crate::types::Value;

/// INT4_ARRAY is the OID of integer[].
const INT4_ARRAY: u32 = 1007;

/// f declares a strict array function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// n declares an array function that runs on NULL arguments.
const fn n(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: false, variadic: false, implementation }
}

/// FUNCTIONS are the array functions.
pub const FUNCTIONS: &[Function] = &[
    f("record_eq", &[RECORD, RECORD], BOOL, |_, a| Ok(Value::Bool(compare_values(&a[0], &a[1]).is_eq()))),
    f("record_ne", &[RECORD, RECORD], BOOL, |_, a| Ok(Value::Bool(compare_values(&a[0], &a[1]).is_ne()))),
    f("record_lt", &[RECORD, RECORD], BOOL, |_, a| Ok(Value::Bool(compare_values(&a[0], &a[1]).is_lt()))),
    f("record_le", &[RECORD, RECORD], BOOL, |_, a| Ok(Value::Bool(compare_values(&a[0], &a[1]).is_le()))),
    f("record_gt", &[RECORD, RECORD], BOOL, |_, a| Ok(Value::Bool(compare_values(&a[0], &a[1]).is_gt()))),
    f("record_ge", &[RECORD, RECORD], BOOL, |_, a| Ok(Value::Bool(compare_values(&a[0], &a[1]).is_ge()))),
    f("array_length", &[ANYARRAY, INT4], INT4, array_length),
    f("array_lower", &[ANYARRAY, INT4], INT4, array_lower),
    f("array_upper", &[ANYARRAY, INT4], INT4, array_upper),
    f("array_ndims", &[ANYARRAY], INT4, array_ndims),
    f("array_dims", &[ANYARRAY], TEXT, array_dims),
    f("cardinality", &[ANYARRAY], INT4, cardinality),
    n("array_append", &[ANYARRAY, ANYELEMENT], ANYARRAY, array_append),
    n("array_prepend", &[ANYELEMENT, ANYARRAY], ANYARRAY, array_prepend),
    n("array_cat", &[ANYARRAY, ANYARRAY], ANYARRAY, array_cat),
    n("array_remove", &[ANYARRAY, ANYELEMENT], ANYARRAY, array_remove),
    n("array_replace", &[ANYARRAY, ANYELEMENT, ANYELEMENT], ANYARRAY, array_replace),
    n("array_position", &[ANYARRAY, ANYELEMENT], INT4, array_position),
    n("array_position", &[ANYARRAY, ANYELEMENT, INT4], INT4, array_position),
    n("array_positions", &[ANYARRAY, ANYELEMENT], INT4_ARRAY, array_positions),
    f("array_to_string", &[ANYARRAY, TEXT], TEXT, array_to_string),
    n("array_to_string", &[ANYARRAY, TEXT, TEXT], TEXT, array_to_string),
    n("string_to_array", &[TEXT, TEXT], TEXT_ARRAY, string_to_array),
    n("string_to_array", &[TEXT, TEXT, TEXT], TEXT_ARRAY, string_to_array),
    n("array_fill", &[ANYELEMENT, INT4_ARRAY], ANYARRAY, array_fill),
    n("array_fill", &[ANYELEMENT, INT4_ARRAY, INT4_ARRAY], ANYARRAY, array_fill),
    f("trim_array", &[ANYARRAY, INT4], ANYARRAY, trim_array),
    f("unnest", &[ANYARRAY], ANYELEMENT, unnest),
    f("generate_subscripts", &[ANYARRAY, INT4], INT4, generate_subscripts),
    f("generate_subscripts", &[ANYARRAY, INT4, BOOL], INT4, generate_subscripts),
];

/// arr returns an array argument.
fn arr(value: &Value) -> Option<&Array> {
    match value {
        Value::Array(a) => Some(a),
        _ => None,
    }
}

/// int returns an integer argument.
fn int(value: &Value) -> i32 {
    match value {
        Value::Int4(i) => *i,
        _ => 0,
    }
}

/// dimension returns the length and lower bound of an array's dimension, counted from 1.
fn dimension(args: &[Value]) -> Option<(i32, i32)> {
    let dim = int(&args[1]);
    let a = arr(&args[0])?;
    usize::try_from(dim).ok().filter(|&d| d >= 1).and_then(|d| a.dims.get(d - 1)).copied()
}

/// or_null returns an integer, or NULL when it is missing.
fn or_null(value: Option<i32>) -> Value {
    value.map_or(Value::Null, Value::Int4)
}

/// array_length returns the length of a dimension.
fn array_length(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(or_null(dimension(args).map(|(n, _)| n)))
}

/// array_lower returns the lower bound of a dimension.
fn array_lower(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(or_null(dimension(args).map(|(_, l)| l)))
}

/// array_upper returns the upper bound of a dimension.
fn array_upper(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(or_null(dimension(args).map(|(n, l)| l + n - 1)))
}

/// array_ndims returns the number of dimensions.
fn array_ndims(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(or_null(arr(&args[0]).map(|a| a.dims.len() as i32).filter(|&d| d > 0)))
}

/// array_dims returns the bounds of every dimension, as `[1:2][1:3]`.
fn array_dims(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let Some(a) = arr(&args[0]).filter(|a| !a.dims.is_empty()) else { return Ok(Value::Null) };
    Ok(Value::Text(a.dims.iter().map(|(n, l)| format!("[{l}:{}]", l + n - 1)).collect()))
}

/// cardinality returns the number of elements.
fn cardinality(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(arr(&args[0]).map_or(0, |a| a.values.len() as i32)))
}

/// or_empty returns an array argument, or an empty array of the element type for NULL.
fn or_empty(value: &Value, element: u32) -> Array {
    arr(value).cloned().unwrap_or(Array { element, dims: Vec::new(), values: Vec::new() })
}

/// array_append adds an element to the end of an array.
fn array_append(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let a = or_empty(&args[0], value_type(&args[1]));
    Ok(Value::Array(Box::new(array::append(a, args[1].clone(), false)?)))
}

/// array_prepend adds an element to the start of an array.
fn array_prepend(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let a = or_empty(&args[1], value_type(&args[0]));
    Ok(Value::Array(Box::new(array::append(a, args[0].clone(), true)?)))
}

/// array_cat joins two arrays.
fn array_cat(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    array::operate(crate::expr::ArrayOp::Concat, args[0].clone(), args[1].clone())
}

/// same reports whether two values are equal, counting two NULLs as equal.
fn same(left: &Value, right: &Value) -> bool {
    match (left.is_null(), right.is_null()) {
        (true, true) => true,
        (false, false) => compare_values(left, right) == Ordering::Equal,
        _ => false,
    }
}

/// one_dimensional returns an array argument, failing for a multidimensional one as the searching functions do.
fn one_dimensional<'v>(value: &'v Value, action: &str) -> Result<Option<&'v Array>> {
    match arr(value) {
        Some(a) if a.dims.len() > 1 => Err(PgError::new(
            code::FEATURE_NOT_SUPPORTED,
            format!("{action} in multidimensional arrays is not supported"),
        )),
        a => Ok(a),
    }
}

/// array_remove removes every element equal to the value.
fn array_remove(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let Some(a) = arr(&args[0]) else { return Ok(Value::Null) };
    if a.dims.len() > 1 {
        return Err(PgError::new(
            code::FEATURE_NOT_SUPPORTED,
            "removing elements from multidimensional arrays is not supported",
        ));
    }
    let values: Vec<Value> = a.values.iter().filter(|v| !same(v, &args[1])).cloned().collect();
    let lower = a.dims.first().map_or(1, |d| d.1);
    let mut result = Array::one_dimensional(a.element, values);
    if let Some(d) = result.dims.first_mut() {
        d.1 = lower;
    }
    Ok(Value::Array(Box::new(result)))
}

/// array_replace replaces every element equal to the value.
fn array_replace(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let Some(a) = arr(&args[0]) else { return Ok(Value::Null) };
    let values = a.values.iter().map(|v| if same(v, &args[1]) { args[2].clone() } else { v.clone() }).collect();
    Ok(Value::Array(Box::new(Array { values, ..a.clone() })))
}

/// array_position returns the subscript of the first element equal to the value, from an optional start.
fn array_position(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let Some(a) = one_dimensional(&args[0], "searching for elements")? else { return Ok(Value::Null) };
    let lower = a.dims.first().map_or(1, |d| d.1);
    let start = match args.get(2) {
        Some(Value::Null) => {
            return Err(PgError::new(code::NULL_VALUE_NOT_ALLOWED, "initial position must not be null"));
        }
        Some(v) => int(v),
        None => lower,
    };
    Ok(or_null(
        a.values
            .iter()
            .enumerate()
            .map(|(i, v)| (lower + i as i32, v))
            .find(|(i, v)| *i >= start && same(v, &args[1]))
            .map(|(i, _)| i),
    ))
}

/// array_positions returns the subscripts of every element equal to the value.
fn array_positions(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let Some(a) = one_dimensional(&args[0], "searching for elements")? else { return Ok(Value::Null) };
    let lower = a.dims.first().map_or(1, |d| d.1);
    let positions = a
        .values
        .iter()
        .enumerate()
        .filter(|(_, v)| same(v, &args[1]))
        .map(|(i, _)| Value::Int4(lower + i as i32))
        .collect::<Vec<_>>();
    Ok(Value::Array(Box::new(Array::one_dimensional(INT4, positions))))
}

/// array_to_string joins the elements with a delimiter, writing NULLs as the optional string or skipping them.
fn array_to_string(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (Some(a), Value::Text(delimiter)) = (arr(&args[0]), &args[1]) else { return Ok(Value::Null) };
    let null_string = args.get(2).and_then(Value::output);
    let parts: Vec<String> = a.values.iter().filter_map(|v| v.output().or_else(|| null_string.clone())).collect();
    Ok(Value::Text(parts.join(delimiter)))
}

/// string_to_array splits a string at a delimiter, or into characters for a NULL delimiter, turning fields equal to
/// the optional string into NULLs.
fn string_to_array(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let Value::Text(input) = &args[0] else { return Ok(Value::Null) };
    if input.is_empty() {
        return Ok(Value::Array(Box::new(Array::one_dimensional(TEXT, Vec::new()))));
    }
    let fields: Vec<String> = match &args[1] {
        Value::Null => input.chars().map(String::from).collect(),
        Value::Text(d) if d.is_empty() => vec![input.clone()],
        Value::Text(d) => input.split(d.as_str()).map(String::from).collect(),
        _ => Vec::new(),
    };
    let null_string = args.get(2).and_then(Value::output);
    let values = fields
        .into_iter()
        .map(|f| if null_string.as_ref() == Some(&f) { Value::Null } else { Value::Text(f) })
        .collect();
    Ok(Value::Array(Box::new(Array::one_dimensional(TEXT, values))))
}

/// array_fill returns an array of the dimensions, and optional lower bounds, with every element set to the value.
fn array_fill(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let null_bounds =
        || PgError::new(code::NULL_VALUE_NOT_ALLOWED, "dimension array or low bound array cannot be null");
    let ints = |v: &Value| -> Result<Vec<i32>> {
        let a = arr(v).ok_or_else(null_bounds)?;
        if a.dims.len() > 1 {
            return Err(PgError {
                detail: Some("Dimension array must be one dimensional.".into()),
                ..PgError::new(code::ARRAY_SUBSCRIPT_ERROR, "wrong number of array subscripts")
            });
        }
        a.values
            .iter()
            .map(|v| match v {
                Value::Int4(i) => Ok(*i),
                _ => Err(PgError::new(code::NULL_VALUE_NOT_ALLOWED, "dimension values cannot be null")),
            })
            .collect()
    };
    let lengths = ints(&args[1])?;
    let lowers = match args.get(2) {
        Some(v) => ints(v)?,
        None => vec![1; lengths.len()],
    };
    if lengths.len() != lowers.len() {
        return Err(PgError {
            detail: Some("Low bound array has different size than dimensions array.".into()),
            ..PgError::new(code::ARRAY_SUBSCRIPT_ERROR, "wrong number of array subscripts")
        });
    }
    let element = value_type(&args[0]);
    if lengths.is_empty() || lengths.contains(&0) {
        return Ok(Value::Array(Box::new(Array { element, dims: Vec::new(), values: Vec::new() })));
    }
    if lengths.iter().any(|&n| n < 0) {
        return Err(PgError::new(code::ARRAY_SUBSCRIPT_ERROR, "array size exceeds the maximum allowed (134217727)"));
    }
    let count = lengths.iter().map(|&n| n as usize).product();
    let dims = lengths.into_iter().zip(lowers).collect();
    Ok(Value::Array(Box::new(Array { element, dims, values: vec![args[0].clone(); count] })))
}

/// trim_array removes elements from the end of an array.
fn trim_array(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let Some(a) = arr(&args[0]) else { return Ok(Value::Null) };
    let length = a.dims.first().map_or(0, |d| d.0);
    let trim = int(&args[1]);
    if trim < 0 || trim > length {
        return Err(PgError::new(
            code::ARRAY_SUBSCRIPT_ERROR,
            format!("number of elements to trim must be between 0 and {length}"),
        ));
    }
    let lower = a.dims.first().map_or(1, |d| d.1);
    Ok(Value::Array(Box::new(array::slice(a, &[(Some(lower), Some(lower + length - trim - 1))]))))
}

/// unnest returns an array's elements as rows.
fn unnest(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Set(arr(&args[0]).map_or_else(Vec::new, |a| a.values.clone())))
}

/// generate_subscripts returns the subscripts of a dimension as rows, in reverse when asked.
fn generate_subscripts(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let Some((length, lower)) = dimension(args) else { return Ok(Value::Set(Vec::new())) };
    let mut subscripts: Vec<Value> = (lower..lower + length).map(Value::Int4).collect();
    if args.get(2) == Some(&Value::Bool(true)) {
        subscripts.reverse();
    }
    Ok(Value::Set(subscripts))
}

/// value_type returns the type of a value, taking text for NULL and strings.
pub fn value_type(value: &Value) -> u32 {
    use crate::oid;
    match value {
        Value::Bool(_) => oid::BOOL,
        Value::Int2(_) => oid::INT2,
        Value::Int4(_) => oid::INT4,
        Value::Int8(_) => oid::INT8,
        Value::Float4(_) => oid::FLOAT4,
        Value::Float8(_) => oid::FLOAT8,
        Value::Numeric(_) => oid::NUMERIC,
        Value::Date(_) => oid::DATE,
        Value::Time(_) => oid::TIME,
        Value::TimeTz(..) => oid::TIMETZ,
        Value::Timestamp(_) => oid::TIMESTAMP,
        Value::TimestampTz(_) => oid::TIMESTAMPTZ,
        Value::Interval(_) => oid::INTERVAL,
        Value::Array(a) => a.array_type(),
        Value::Json(_) => oid::JSON,
        Value::Jsonb(_) => oid::JSONB,
        Value::Record(_) => oid::RECORD,
        Value::Oid(_) => oid::OID,
        Value::Reg(reg) => reg.type_oid,
        Value::Enum(e) => e.type_oid,
        Value::Composite(c) => c.type_oid,
        _ => oid::TEXT,
    }
}
