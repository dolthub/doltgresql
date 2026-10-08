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

//! JSON functions and the functions behind the json and jsonb operators.

use super::{ANY, ANYARRAY, ANYELEMENT, Function};
use crate::array::Array;
use crate::error::{PgError, Result, code};
use crate::json::{self, Json};
use crate::numeric::Numeric;
use crate::oid::{BOOL, INT4, JSON, JSONB, RECORD, TEXT, TEXT_ARRAY};
use crate::query::Ctx;
use crate::types::Value;

/// f declares a strict json function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// v declares a json function that takes one or more arguments of any type.
const fn v(name: &'static str, ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args: &[ANY], ret, strict: false, variadic: true, implementation }
}

/// FUNCTIONS are the json functions, including ones named after the operators they implement.
pub const FUNCTIONS: &[Function] = &[
    f("->", &[JSON, TEXT], JSON, field),
    f("->", &[JSON, INT4], JSON, element),
    f("->", &[JSONB, TEXT], JSONB, field),
    f("->", &[JSONB, INT4], JSONB, element),
    f("->>", &[JSON, TEXT], TEXT, field_text),
    f("->>", &[JSON, INT4], TEXT, element_text),
    f("->>", &[JSONB, TEXT], TEXT, field_text),
    f("->>", &[JSONB, INT4], TEXT, element_text),
    f("#>", &[JSON, TEXT_ARRAY], JSON, path),
    f("#>", &[JSONB, TEXT_ARRAY], JSONB, path),
    f("#>>", &[JSON, TEXT_ARRAY], TEXT, path_text),
    f("#>>", &[JSONB, TEXT_ARRAY], TEXT, path_text),
    f("@>", &[JSONB, JSONB], BOOL, contains),
    f("<@", &[JSONB, JSONB], BOOL, contained),
    f("?", &[JSONB, TEXT], BOOL, exists),
    f("?|", &[JSONB, TEXT_ARRAY], BOOL, exists_any),
    f("?&", &[JSONB, TEXT_ARRAY], BOOL, exists_all),
    f("||", &[JSONB, JSONB], JSONB, concat),
    f("-", &[JSONB, TEXT], JSONB, delete_key),
    f("-", &[JSONB, INT4], JSONB, delete_index),
    f("-", &[JSONB, TEXT_ARRAY], JSONB, delete_keys),
    f("#-", &[JSONB, TEXT_ARRAY], JSONB, delete_path),
    f("json_typeof", &[JSON], TEXT, typeof_),
    f("jsonb_typeof", &[JSONB], TEXT, typeof_),
    f("json_array_length", &[JSON], INT4, array_length),
    f("jsonb_array_length", &[JSONB], INT4, array_length),
    f("jsonb_object_keys", &[JSONB], TEXT, object_keys),
    f("json_object_keys", &[JSON], TEXT, object_keys),
    f("jsonb_array_elements", &[JSONB], JSONB, array_elements),
    f("json_array_elements", &[JSON], JSON, array_elements),
    f("jsonb_array_elements_text", &[JSONB], TEXT, array_elements_text),
    f("json_array_elements_text", &[JSON], TEXT, array_elements_text),
    f("jsonb_each", &[JSONB], crate::dolt::procedures::RECORD, each),
    f("json_each", &[JSON], crate::dolt::procedures::RECORD, each),
    f("jsonb_each_text", &[JSONB], crate::dolt::procedures::RECORD, each_text),
    f("json_each_text", &[JSON], crate::dolt::procedures::RECORD, each_text),
    f("jsonb_strip_nulls", &[JSONB], JSONB, strip_nulls),
    f("json_strip_nulls", &[JSON], JSON, strip_nulls),
    f("jsonb_pretty", &[JSONB], TEXT, pretty),
    f("jsonb_set", &[JSONB, TEXT_ARRAY, JSONB], JSONB, set),
    f("jsonb_set", &[JSONB, TEXT_ARRAY, JSONB, BOOL], JSONB, set),
    f("jsonb_insert", &[JSONB, TEXT_ARRAY, JSONB], JSONB, insert),
    f("jsonb_insert", &[JSONB, TEXT_ARRAY, JSONB, BOOL], JSONB, insert),
    Function {
        name: "jsonb_extract_path",
        args: &[JSONB, TEXT],
        ret: JSONB,
        strict: true,
        variadic: true,
        implementation: extract_path,
    },
    Function {
        name: "json_extract_path",
        args: &[JSON, TEXT],
        ret: JSON,
        strict: true,
        variadic: true,
        implementation: extract_path,
    },
    Function {
        name: "jsonb_extract_path_text",
        args: &[JSONB, TEXT],
        ret: TEXT,
        strict: true,
        variadic: true,
        implementation: extract_path_text,
    },
    Function {
        name: "json_extract_path_text",
        args: &[JSON, TEXT],
        ret: TEXT,
        strict: true,
        variadic: true,
        implementation: extract_path_text,
    },
    f("jsonb_exists", &[JSONB, TEXT], BOOL, exists),
    f("to_json", &[ANYELEMENT], JSON, to_json),
    f("to_jsonb", &[ANYELEMENT], JSONB, to_jsonb),
    f("row_to_json", &[ANYELEMENT], JSON, to_json),
    f("row_to_json", &[RECORD, BOOL], JSON, to_json_pretty),
    f("array_to_json", &[ANYARRAY], JSON, to_json),
    f("array_to_json", &[ANYARRAY, BOOL], JSON, to_json_pretty),
    v("json_build_object", JSON, build_object),
    v("jsonb_build_object", JSONB, build_object_b),
    v("json_build_array", JSON, build_array),
    v("jsonb_build_array", JSONB, build_array_b),
    f("json_build_object", &[], JSON, build_object),
    f("jsonb_build_object", &[], JSONB, build_object_b),
    f("json_build_array", &[], JSON, build_array),
    f("jsonb_build_array", &[], JSONB, build_array_b),
    f("json_object", &[TEXT_ARRAY], JSON, json_object),
    f("jsonb_object", &[TEXT_ARRAY], JSONB, jsonb_object),
];

/// OUT_COLUMNS are the result columns of the json functions that return records.
pub const OUT_COLUMNS: &[(&str, &[(&str, u32)])] = &[
    ("jsonb_array_elements", &[("value", JSONB)]),
    ("json_array_elements", &[("value", JSON)]),
    ("jsonb_array_elements_text", &[("value", TEXT)]),
    ("json_array_elements_text", &[("value", TEXT)]),
    ("jsonb_each", &[("key", TEXT), ("value", JSONB)]),
    ("json_each", &[("key", TEXT), ("value", JSON)]),
    ("jsonb_each_text", &[("key", TEXT), ("value", TEXT)]),
    ("json_each_text", &[("key", TEXT), ("value", TEXT)]),
];

/// document returns a json or jsonb argument as a value, keeping json objects' keys in their written order.
fn document(value: &Value) -> Result<Json> {
    match value {
        Value::Jsonb(json) => Ok((**json).clone()),
        Value::Json(text) => json::parse(text, false),
        Value::Text(text) => json::parse(text, false),
        _ => Ok(Json::Null),
    }
}

/// wrap returns a result as the json or jsonb that the input was.
fn wrap(input: &Value, json: Json) -> Value {
    match input {
        Value::Jsonb(_) => Value::Jsonb(Box::new(json)),
        _ => Value::Json(json_text(&json)),
    }
}

/// json_text prints a value as json functions write it, with no spaces inside arrays and objects of parsed json.
fn json_text(json: &Json) -> String {
    json.to_text()
}

/// text_arg returns a text argument.
fn text_arg(value: &Value) -> String {
    value.output().unwrap_or_default()
}

/// int_arg returns an integer argument.
fn int_arg(value: &Value) -> i64 {
    match value {
        Value::Int4(i) => *i as i64,
        Value::Int8(i) => *i,
        _ => 0,
    }
}

/// raw_part returns part of a json or jsonb value, keeping a json part's text as written.
fn raw_part(
    input: &Value,
    pick: &dyn Fn(&Json) -> Option<Json>,
    pick_raw: &dyn for<'r> Fn(&'r json::Raw<'r>) -> Option<&'r json::Raw<'r>>,
) -> Result<Value> {
    if let Value::Json(text) = input {
        let raw = json::parse_raw(text)?;
        return Ok(pick_raw(&raw).map_or(Value::Null, |r| Value::Json(r.text.to_string())));
    }
    let json = document(input)?;
    Ok(pick(&json).map_or(Value::Null, |j| wrap(input, j)))
}

/// raw_text returns part of a json or jsonb value as ->> does.
fn raw_text(
    input: &Value,
    pick: &dyn Fn(&Json) -> Option<Json>,
    pick_raw: &dyn for<'r> Fn(&'r json::Raw<'r>) -> Option<&'r json::Raw<'r>>,
) -> Result<Value> {
    if let Value::Json(text) = input {
        let raw = json::parse_raw(text)?;
        return Ok(pick_raw(&raw).and_then(json::Raw::scalar_text).map_or(Value::Null, Value::Text));
    }
    Ok(scalar(pick(&document(input)?).as_ref()))
}

/// field implements -> with a key.
fn field(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let key = text_arg(&args[1]);
    raw_part(&args[0], &|j| j.get(&key).cloned(), &|r| r.get(&key))
}

/// element implements -> with an index.
fn element(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let index = int_arg(&args[1]);
    raw_part(&args[0], &|j| scalar_element(j, index).cloned(), &|r| r.index(index))
}

/// scalar_element returns the element at an index of a jsonb array, where a scalar is the one element of an array, as
/// Postgres stores a scalar.
fn scalar_element(json: &Json, index: i64) -> Option<&Json> {
    match json {
        Json::Array(_) | Json::Object(_) => json.index(index),
        scalar => (index == 0 || index == -1).then_some(scalar),
    }
}

/// scalar returns a value as ->> returns it.
fn scalar(json: Option<&Json>) -> Value {
    json.and_then(Json::scalar_text).map_or(Value::Null, Value::Text)
}

/// field_text implements ->> with a key.
fn field_text(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let key = text_arg(&args[1]);
    raw_text(&args[0], &|j| j.get(&key).cloned(), &|r| r.get(&key))
}

/// element_text implements ->> with an index.
fn element_text(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let index = int_arg(&args[1]);
    raw_text(&args[0], &|j| scalar_element(j, index).cloned(), &|r| r.index(index))
}

/// path_steps returns a text array argument's elements.
fn path_steps(value: &Value) -> Vec<Option<String>> {
    match value {
        Value::Array(a) => a.values.iter().map(Value::output).collect(),
        _ => Vec::new(),
    }
}

/// path implements #>.
fn path(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let steps = path_steps(&args[1]);
    raw_part(&args[0], &|j| j.path(&steps).cloned(), &|r| r.path(&steps))
}

/// path_text implements #>>.
fn path_text(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let steps = path_steps(&args[1]);
    raw_text(&args[0], &|j| j.path(&steps).cloned(), &|r| r.path(&steps))
}

/// extract_path implements json(b)_extract_path.
fn extract_path(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let steps: Vec<Option<String>> = args[1..].iter().map(Value::output).collect();
    raw_part(&args[0], &|j| j.path(&steps).cloned(), &|r| r.path(&steps))
}

/// extract_path_text implements json(b)_extract_path_text.
fn extract_path_text(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let steps: Vec<Option<String>> = args[1..].iter().map(Value::output).collect();
    raw_text(&args[0], &|j| j.path(&steps).cloned(), &|r| r.path(&steps))
}

/// contains implements @>.
fn contains(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(document(&args[0])?.contains(&document(&args[1])?)))
}

/// contained implements <@.
fn contained(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(document(&args[1])?.contains(&document(&args[0])?)))
}

/// has_key reports whether a key is a top-level key or string array element.
fn has_key(json: &Json, key: &str) -> bool {
    match json {
        Json::Object(items) => items.iter().any(|(k, _)| k == key),
        Json::Array(values) => values.iter().any(|v| matches!(v, Json::String(s) if s == key)),
        Json::String(s) => s == key,
        _ => false,
    }
}

/// exists implements ?.
fn exists(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Bool(has_key(&document(&args[0])?, &text_arg(&args[1]))))
}

/// exists_any implements ?|.
fn exists_any(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let json = document(&args[0])?;
    Ok(Value::Bool(path_steps(&args[1]).iter().flatten().any(|k| has_key(&json, k))))
}

/// exists_all implements ?&.
fn exists_all(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let json = document(&args[0])?;
    Ok(Value::Bool(path_steps(&args[1]).iter().flatten().all(|k| has_key(&json, k))))
}

/// concat implements jsonb ||.
fn concat(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (a, b) = (document(&args[0])?, document(&args[1])?);
    let result = match (a, b) {
        (Json::Object(mut x), Json::Object(y)) => {
            x.extend(y);
            json::normalize(Json::Object(x))
        }
        (Json::Array(mut x), Json::Array(y)) => {
            x.extend(y);
            Json::Array(x)
        }
        (Json::Array(mut x), other) => {
            x.push(other);
            Json::Array(x)
        }
        (other, Json::Array(y)) => {
            let mut values = vec![other];
            values.extend(y);
            Json::Array(values)
        }
        (x, y) => Json::Array(vec![x, y]),
    };
    Ok(Value::Jsonb(Box::new(result)))
}

/// scalar_error returns Postgres' error for removing something from a jsonb scalar.
fn scalar_error() -> PgError {
    PgError::new(code::INVALID_PARAMETER_VALUE, "cannot delete from scalar")
}

/// delete_key implements jsonb - text.
fn delete_key(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let key = text_arg(&args[1]);
    let result = match document(&args[0])? {
        Json::Object(items) => Json::Object(items.into_iter().filter(|(k, _)| *k != key).collect()),
        Json::Array(values) => {
            Json::Array(values.into_iter().filter(|v| !matches!(v, Json::String(s) if *s == key)).collect())
        }
        _ => return Err(scalar_error()),
    };
    Ok(Value::Jsonb(Box::new(result)))
}

/// delete_keys implements jsonb - text[].
fn delete_keys(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let keys: Vec<String> = path_steps(&args[1]).into_iter().flatten().collect();
    let result = match document(&args[0])? {
        Json::Object(items) => Json::Object(items.into_iter().filter(|(k, _)| !keys.contains(k)).collect()),
        Json::Array(values) => {
            Json::Array(values.into_iter().filter(|v| !matches!(v, Json::String(s) if keys.contains(s))).collect())
        }
        _ => return Err(scalar_error()),
    };
    Ok(Value::Jsonb(Box::new(result)))
}

/// delete_index implements jsonb - integer.
fn delete_index(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let index = int_arg(&args[1]);
    let result = match document(&args[0])? {
        Json::Array(mut values) => {
            let i = if index < 0 { values.len() as i64 + index } else { index };
            if i >= 0 && (i as usize) < values.len() {
                values.remove(i as usize);
            }
            Json::Array(values)
        }
        Json::Object(_) => {
            return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "cannot delete from object using integer index"));
        }
        _ => return Err(scalar_error()),
    };
    Ok(Value::Jsonb(Box::new(result)))
}

/// delete_at removes the value at a path.
fn delete_at(json: &mut Json, steps: &[String]) {
    let Some((first, rest)) = steps.split_first() else { return };
    match json {
        Json::Object(items) => {
            if rest.is_empty() {
                items.retain(|(k, _)| k != first);
            } else if let Some((_, v)) = items.iter_mut().find(|(k, _)| k == first) {
                delete_at(v, rest);
            }
        }
        Json::Array(values) => {
            let Ok(index) = first.parse::<i64>() else { return };
            let i = if index < 0 { values.len() as i64 + index } else { index };
            if i < 0 || i as usize >= values.len() {
                return;
            }
            if rest.is_empty() {
                values.remove(i as usize);
            } else {
                delete_at(&mut values[i as usize], rest);
            }
        }
        _ => {}
    }
}

/// delete_path implements #-.
fn delete_path(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let mut json = document(&args[0])?;
    if !matches!(json, Json::Object(_) | Json::Array(_)) {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "cannot delete path in scalar"));
    }
    let steps: Vec<String> = path_steps(&args[1]).into_iter().flatten().collect();
    delete_at(&mut json, &steps);
    Ok(Value::Jsonb(Box::new(json)))
}

/// typeof_ returns a value's type name.
fn typeof_(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    if let Value::Json(text) = &args[0] {
        return Ok(Value::Text(json::parse_raw(text)?.type_name().into()));
    }
    Ok(Value::Text(document(&args[0])?.type_name().into()))
}

/// array_length returns an array's length.
fn array_length(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    match document(&args[0])? {
        Json::Array(values) => Ok(Value::Int4(values.len() as i32)),
        other => Err(PgError::new(
            code::INVALID_PARAMETER_VALUE,
            format!(
                "cannot get array length of a {}",
                if matches!(other, Json::Object(_)) { "non-array" } else { "scalar" }
            ),
        )),
    }
}

/// object_keys returns an object's keys as rows.
fn object_keys(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    if let Value::Json(text) = &args[0]
        && let json::RawKind::Object(items) = json::parse_raw(text)?.kind
    {
        return Ok(Value::Set(items.into_iter().map(|(k, _)| Value::Text(k)).collect()));
    }
    let name = if matches!(args[0], Value::Json(_)) { "json_object_keys" } else { "jsonb_object_keys" };
    match document(&args[0])? {
        Json::Object(items) => Ok(Value::Set(items.into_iter().map(|(k, _)| Value::Text(k)).collect())),
        other => Err(PgError::new(
            code::INVALID_PARAMETER_VALUE,
            format!("cannot call {name} on {}", if matches!(other, Json::Array(_)) { "an array" } else { "a scalar" }),
        )),
    }
}

/// elements returns an array's elements, failing for anything else.
fn elements(value: &Value) -> Result<Vec<Json>> {
    match document(value)? {
        Json::Array(values) => Ok(values),
        Json::Object(_) => Err(PgError::new(code::INVALID_PARAMETER_VALUE, "cannot extract elements from an object")),
        _ => Err(PgError::new(code::INVALID_PARAMETER_VALUE, "cannot extract elements from a scalar")),
    }
}

/// raw_elements returns a json array's elements as written.
fn raw_elements(text: &str) -> Result<Vec<json::Raw<'_>>> {
    match json::parse_raw(text)?.kind {
        json::RawKind::Array(values) => Ok(values),
        json::RawKind::Object(_) => {
            Err(PgError::new(code::INVALID_PARAMETER_VALUE, "cannot extract elements from an object"))
        }
        json::RawKind::Scalar(_) => {
            Err(PgError::new(code::INVALID_PARAMETER_VALUE, "cannot extract elements from a scalar"))
        }
    }
}

/// array_elements returns an array's elements as rows.
fn array_elements(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    if let Value::Json(text) = &args[0] {
        return Ok(Value::Set(raw_elements(text)?.iter().map(|r| Value::Json(r.text.to_string())).collect()));
    }
    Ok(Value::Set(elements(&args[0])?.into_iter().map(|j| wrap(&args[0], j)).collect()))
}

/// array_elements_text returns an array's elements as text rows.
fn array_elements_text(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    if let Value::Json(text) = &args[0] {
        let values = raw_elements(text)?.iter().map(|r| r.scalar_text().map_or(Value::Null, Value::Text)).collect();
        return Ok(Value::Set(values));
    }
    Ok(Value::Set(elements(&args[0])?.iter().map(|j| scalar(Some(j))).collect()))
}

/// raw_members returns a json object's keys and values as written.
fn raw_members(text: &str) -> Result<Vec<(String, json::Raw<'_>)>> {
    match json::parse_raw(text)?.kind {
        json::RawKind::Object(items) => Ok(items),
        _ => Err(PgError::new(code::INVALID_PARAMETER_VALUE, "cannot deconstruct an array as an object")),
    }
}

/// members returns an object's keys and values, failing for anything else.
fn members(value: &Value) -> Result<Vec<(String, Json)>> {
    match document(value)? {
        Json::Object(items) => Ok(items),
        _ => Err(PgError::new(code::INVALID_PARAMETER_VALUE, "cannot deconstruct an array as an object")),
    }
}

/// each returns an object's keys and values as rows.
fn each(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    if let Value::Json(text) = &args[0] {
        let rows = raw_members(text)?
            .into_iter()
            .map(|(k, r)| Value::Record(vec![Value::Text(k), Value::Json(r.text.to_string())]))
            .collect();
        return Ok(Value::Set(rows));
    }
    let rows =
        members(&args[0])?.into_iter().map(|(k, j)| Value::Record(vec![Value::Text(k), wrap(&args[0], j)])).collect();
    Ok(Value::Set(rows))
}

/// each_text returns an object's keys and values as text rows.
fn each_text(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    if let Value::Json(text) = &args[0] {
        let rows = raw_members(text)?
            .into_iter()
            .map(|(k, r)| Value::Record(vec![Value::Text(k), r.scalar_text().map_or(Value::Null, Value::Text)]))
            .collect();
        return Ok(Value::Set(rows));
    }
    let rows =
        members(&args[0])?.into_iter().map(|(k, j)| Value::Record(vec![Value::Text(k), scalar(Some(&j))])).collect();
    Ok(Value::Set(rows))
}

/// strip removes object fields whose values are null.
fn strip(json: Json) -> Json {
    match json {
        Json::Object(items) => {
            Json::Object(items.into_iter().filter(|(_, v)| *v != Json::Null).map(|(k, v)| (k, strip(v))).collect())
        }
        Json::Array(values) => Json::Array(values.into_iter().map(strip).collect()),
        other => other,
    }
}

/// strip_nulls removes object fields whose values are null, recursively.
fn strip_nulls(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let stripped = strip(document(&args[0])?);
    Ok(match args[0] {
        Value::Json(_) => Value::Json(stripped.compact()),
        _ => wrap(&args[0], stripped),
    })
}

/// write_pretty prints a value as jsonb_pretty does, indenting nested values by four spaces.
fn write_pretty(json: &Json, depth: usize, out: &mut String) {
    let indent = |out: &mut String, depth: usize| out.push_str(&"    ".repeat(depth));
    match json {
        Json::Array(values) if !values.is_empty() => {
            out.push_str("[\n");
            for (i, v) in values.iter().enumerate() {
                indent(out, depth + 1);
                write_pretty(v, depth + 1, out);
                out.push_str(if i + 1 < values.len() { ",\n" } else { "\n" });
            }
            indent(out, depth);
            out.push(']');
        }
        Json::Object(items) if !items.is_empty() => {
            out.push_str("{\n");
            for (i, (k, v)) in items.iter().enumerate() {
                indent(out, depth + 1);
                json::escape(out, k);
                out.push_str(": ");
                write_pretty(v, depth + 1, out);
                out.push_str(if i + 1 < items.len() { ",\n" } else { "\n" });
            }
            indent(out, depth);
            out.push('}');
        }
        other => other.write(out),
    }
}

/// pretty prints a value with indentation.
fn pretty(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let mut out = String::new();
    write_pretty(&document(&args[0])?, 0, &mut out);
    Ok(Value::Text(out))
}

/// set_at replaces or adds the value at a path, or inserts it before or after an array position.
fn set_at(json: &mut Json, steps: &[String], value: &Json, create: bool, insert: Option<bool>) -> Result<()> {
    let Some((first, rest)) = steps.split_first() else { return Ok(()) };
    match json {
        Json::Object(items) => match items.iter_mut().position(|(k, _)| k == first) {
            Some(i) if rest.is_empty() => {
                if insert.is_some() {
                    return Err(PgError {
                        hint: Some("Try using the function jsonb_set to replace key value.".into()),
                        ..PgError::new(code::INVALID_PARAMETER_VALUE, "cannot replace existing key")
                    });
                }
                items[i].1 = value.clone();
            }
            Some(i) => set_at(&mut items[i].1, rest, value, create, insert)?,
            None if rest.is_empty() && (create || insert.is_some()) => {
                items.push((first.clone(), value.clone()));
                *json = json::normalize(std::mem::replace(json, Json::Null));
            }
            None => {}
        },
        Json::Array(values) => {
            let Ok(index) = first.parse::<i64>() else {
                return Err(PgError::new(
                    code::INVALID_TEXT_REPRESENTATION,
                    format!("path element at position 1 is not an integer: \"{first}\""),
                ));
            };
            let len = values.len() as i64;
            let i = if index < 0 { len + index } else { index };
            if rest.is_empty() {
                match insert {
                    Some(after) => {
                        let at = if i < 0 {
                            0
                        } else if i > len {
                            len
                        } else {
                            i + after as i64
                        };
                        values.insert(at.clamp(0, len) as usize, value.clone());
                    }
                    None if i >= 0 && i < len => values[i as usize] = value.clone(),
                    None if create && i >= len => values.push(value.clone()),
                    None if create && i < 0 => values.insert(0, value.clone()),
                    None => {}
                }
            } else if i >= 0 && i < len {
                set_at(&mut values[i as usize], rest, value, create, insert)?;
            }
        }
        _ => {}
    }
    Ok(())
}

/// set implements jsonb_set.
fn set(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let mut json = document(&args[0])?;
    if !matches!(json, Json::Object(_) | Json::Array(_)) {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "cannot set path in scalar"));
    }
    let steps: Vec<String> = path_steps(&args[1]).into_iter().flatten().collect();
    let create = !matches!(args.get(3), Some(Value::Bool(false)));
    set_at(&mut json, &steps, &document(&args[2])?, create, None)?;
    Ok(Value::Jsonb(Box::new(json)))
}

/// insert implements jsonb_insert.
fn insert(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let mut json = document(&args[0])?;
    if !matches!(json, Json::Object(_) | Json::Array(_)) {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "cannot set path in scalar"));
    }
    let steps: Vec<String> = path_steps(&args[1]).into_iter().flatten().collect();
    let after = matches!(args.get(3), Some(Value::Bool(true)));
    set_at(&mut json, &steps, &document(&args[2])?, false, Some(after))?;
    Ok(Value::Jsonb(Box::new(json)))
}

/// datum_to_json converts a value to JSON as Postgres' to_json does, writing dates and times in ISO 8601 form.
pub fn datum_to_json(value: &Value) -> Result<Json> {
    use crate::datetime as dt;
    Ok(match value {
        Value::Null => Json::Null,
        Value::Bool(b) => Json::Bool(*b),
        Value::Int2(i) => Json::Number(Numeric::from_i64(*i as i64)),
        Value::Int4(i) => Json::Number(Numeric::from_i64(*i as i64)),
        Value::Int8(i) => Json::Number(Numeric::from_i64(*i)),
        Value::Float4(f) if f.is_finite() => Json::Number(Numeric::parse(&value.output().unwrap_or_default())?),
        Value::Float8(f) if f.is_finite() => Json::Number(Numeric::parse(&value.output().unwrap_or_default())?),
        Value::Numeric(n @ Numeric::Finite { .. }) => Json::Number(n.clone()),
        Value::Json(text) => json::parse(text, false)?,
        Value::Jsonb(json) => (**json).clone(),
        Value::Array(a) => array_to_json(a)?,
        Value::Record(fields) => Json::Object(
            fields
                .iter()
                .enumerate()
                .map(|(i, f)| Ok((format!("f{}", i + 1), datum_to_json(f)?)))
                .collect::<Result<_>>()?,
        ),
        Value::Composite(c) => Json::Object(
            field_names(c)
                .into_iter()
                .zip(&c.fields)
                .map(|(name, f)| Ok((name, datum_to_json(f)?)))
                .collect::<Result<_>>()?,
        ),
        Value::Date(_) | Value::Timestamp(_) | Value::TimestampTz(_) => {
            let text = dt::with_format(|current| {
                let format = dt::Format { style: dt::Style::Iso, ..current.clone() };
                match value {
                    Value::Date(d) => dt::format_date(*d, &format),
                    Value::Timestamp(ts) => dt::format_timestamp(*ts, None, &format),
                    Value::TimestampTz(ts) => {
                        let (offset, name) = format.zone.offset_at(*ts);
                        dt::format_timestamp(*ts, Some((offset, &name)), &format)
                    }
                    _ => String::new(),
                }
            });
            Json::String(iso_8601(value, text))
        }
        other => Json::String(other.output().unwrap_or_default()),
    })
}

/// field_names returns the names of a composite value's fields, which are f1, f2, and so on for an unknown type.
fn field_names(value: &crate::types::CompositeValue) -> Vec<String> {
    match crate::usertypes::get(value.type_oid).map(|t| t.kind.clone()) {
        Some(crate::usertypes::Kind::Composite(attributes)) => attributes.into_iter().map(|(name, _)| name).collect(),
        _ => (1..=value.fields.len()).map(|i| format!("f{i}")).collect(),
    }
}

/// iso_8601 rewrites an ISO-style timestamp as JSON writes it, with a T between the date and time and a full time
/// zone offset.
fn iso_8601(value: &Value, text: String) -> String {
    match value {
        Value::Timestamp(_) => text.replacen(' ', "T", 1),
        Value::TimestampTz(_) => {
            let text = text.replacen(' ', "T", 1);
            let (text, era) = match text.strip_suffix(" BC") {
                Some(text) => (text.to_string(), " BC"),
                None => (text, ""),
            };
            match text.rfind(['+', '-']) {
                Some(i) if i > 10 && text.len() - i == 3 => format!("{text}:00{era}"),
                _ => format!("{text}{era}"),
            }
        }
        _ => text,
    }
}

/// array_to_json converts an array, keeping its dimensions as nested arrays.
fn array_to_json(array: &Array) -> Result<Json> {
    let values = array.values.iter().map(datum_to_json).collect::<Result<Vec<_>>>()?;
    if array.dims.is_empty() {
        return Ok(Json::Array(Vec::new()));
    }
    let mut level: Vec<Json> = values;
    for &(length, _) in array.dims.iter().skip(1).rev() {
        level = level.chunks(length.max(1) as usize).map(|c| Json::Array(c.to_vec())).collect();
    }
    Ok(Json::Array(level))
}

/// datum_json_text returns a value as json text, as to_json and the json builders write each value: json as written,
/// jsonb as it prints, and arrays and records without spaces.
pub fn datum_json_text(value: &Value) -> Result<String> {
    separated_json_text(value, ",")
}

/// separated_json_text converts a value to json text, separating the elements of an array's outer dimension and a row's fields
/// with the separator, as array_to_json and row_to_json do with line feeds.
fn separated_json_text(value: &Value, separator: &str) -> Result<String> {
    Ok(match value {
        Value::Json(text) => text.clone(),
        Value::Jsonb(json) => json.to_text(),
        Value::Array(array) => {
            if array.dims.is_empty() {
                return Ok("[]".into());
            }
            let parts = array.values.iter().map(datum_json_text).collect::<Result<Vec<_>>>()?;
            let mut level: Vec<String> = parts;
            for &(length, _) in array.dims.iter().skip(1).rev() {
                level = level.chunks(length.max(1) as usize).map(|c| format!("[{}]", c.join(","))).collect();
            }
            format!("[{}]", level.join(separator))
        }
        Value::Record(fields) => {
            let mut out = String::from("{");
            for (i, field) in fields.iter().enumerate() {
                if i > 0 {
                    out.push_str(separator);
                }
                out.push_str(&format!("\"f{}\":", i + 1));
                out.push_str(&datum_json_text(field)?);
            }
            out.push('}');
            out
        }
        Value::Composite(c) => {
            let mut out = String::from("{");
            for (i, (name, field)) in field_names(c).into_iter().zip(&c.fields).enumerate() {
                if i > 0 {
                    out.push_str(separator);
                }
                out.push_str(&Json::String(name).plain());
                out.push(':');
                out.push_str(&datum_json_text(field)?);
            }
            out.push('}');
            out
        }
        Value::Float4(_) | Value::Float8(_) => {
            let text = value.output().unwrap_or_default();
            match text.parse::<f64>().is_ok_and(f64::is_finite) {
                true => text,
                false => Json::String(text).plain(),
            }
        }
        other => datum_to_json(other)?.plain(),
    })
}

/// to_json converts a value to json.
fn to_json(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Json(datum_json_text(&args[0])?))
}

/// to_json_pretty converts an array or row to json, with line feeds between its outer elements when asked.
fn to_json_pretty(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let separator = if args[1] == Value::Bool(true) { ",\n " } else { "," };
    Ok(Value::Json(separated_json_text(&args[0], separator)?))
}

/// to_jsonb converts a value to jsonb.
fn to_jsonb(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Jsonb(Box::new(json::normalize(datum_to_json(&args[0])?))))
}

/// object_texts reads json_build_object's alternating keys and values, with each value as json text.
fn object_texts(args: &[Value]) -> Result<Vec<(String, String)>> {
    let pairs = object_pairs(args, "json_build_object")?;
    pairs.into_iter().zip(args.chunks(2)).map(|((key, _), pair)| Ok((key, datum_json_text(&pair[1])?))).collect()
}

/// object_pairs reads build_object's alternating keys and values.
fn object_pairs(args: &[Value], function: &str) -> Result<Vec<(String, Json)>> {
    if !args.len().is_multiple_of(2) {
        return Err(PgError {
            hint: Some(format!("The arguments of {function}() must consist of alternating keys and values.")),
            ..PgError::new(code::INVALID_PARAMETER_VALUE, "argument list must have even number of elements")
        });
    }
    args.chunks(2)
        .enumerate()
        .map(|(i, pair)| {
            let key = pair[0].output().ok_or_else(|| {
                PgError::new(code::NULL_VALUE_NOT_ALLOWED, format!("argument {}: key must not be null", i * 2 + 1))
            })?;
            Ok((key, datum_to_json(&pair[1])?))
        })
        .collect()
}

/// build_object builds a json object, written with spaces around each colon as json_build_object writes it.
fn build_object(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let mut out = String::from("{");
    for (i, (key, value)) in object_texts(args)?.into_iter().enumerate() {
        if i > 0 {
            out.push_str(", ");
        }
        json::escape(&mut out, &key);
        out.push_str(" : ");
        out.push_str(&value);
    }
    out.push('}');
    Ok(Value::Json(out))
}

/// build_object_b builds a jsonb object.
fn build_object_b(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Jsonb(Box::new(json::normalize(Json::Object(object_pairs(args, "jsonb_build_object")?)))))
}

/// build_array builds a json array.
fn build_array(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let parts = args.iter().map(datum_json_text).collect::<Result<Vec<_>>>()?;
    Ok(Value::Json(format!("[{}]", parts.join(", "))))
}

/// build_array_b builds a jsonb array.
fn build_array_b(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let values = args.iter().map(datum_to_json).collect::<Result<Vec<_>>>()?;
    Ok(Value::Jsonb(Box::new(json::normalize(Json::Array(values)))))
}

/// object_from_array reads json_object's array of alternating keys and values, or of key and value pairs.
fn object_from_array(value: &Value) -> Result<Vec<(String, Json)>> {
    let Value::Array(array) = value else { return Ok(Vec::new()) };
    let texts: Vec<Option<String>> = array.values.iter().map(Value::output).collect();
    let pairs: Vec<(Option<String>, Option<String>)> = match array.dims.as_slice() {
        [] => Vec::new(),
        [(n, _)] if n % 2 == 0 => texts.chunks(2).map(|c| (c[0].clone(), c[1].clone())).collect(),
        [_] => return Err(PgError::new(code::ARRAY_SUBSCRIPT_ERROR, "array must have even number of elements")),
        [_, (2, _)] => texts.chunks(2).map(|c| (c[0].clone(), c[1].clone())).collect(),
        _ => return Err(PgError::new(code::ARRAY_SUBSCRIPT_ERROR, "array must have two columns")),
    };
    pairs
        .into_iter()
        .map(|(k, v)| {
            let key =
                k.ok_or_else(|| PgError::new(code::NULL_VALUE_NOT_ALLOWED, "null value not allowed for object key"))?;
            Ok((key, v.map_or(Json::Null, Json::String)))
        })
        .collect()
}

/// json_object builds a json object from text pairs.
fn json_object(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let mut out = String::from("{");
    for (i, (key, value)) in object_from_array(&args[0])?.into_iter().enumerate() {
        if i > 0 {
            out.push_str(", ");
        }
        json::escape(&mut out, &key);
        out.push_str(" : ");
        out.push_str(&value.to_text());
    }
    out.push('}');
    Ok(Value::Json(out))
}

/// jsonb_object builds a jsonb object from text pairs.
fn jsonb_object(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Jsonb(Box::new(json::normalize(Json::Object(object_from_array(&args[0])?)))))
}
