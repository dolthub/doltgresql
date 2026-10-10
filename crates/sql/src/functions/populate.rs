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

//! json_populate_record, jsonb_populate_record, and their recordset forms: values of a composite type built from JSON
//! objects, converting each field as Postgres' jsonfuncs.c does.

use super::{ANYELEMENT, Function};
use crate::catalog::ColumnType;
use crate::error::{PgError, Result, code};
use crate::expr::typ;
use crate::json::{Json, Raw, RawKind};
use crate::oid::{BOOL, JSON, JSONB};
use crate::query::Ctx;
use crate::types::{CompositeValue, Value};
use crate::usertypes::Kind;

/// f declares a populate function, whose binder passes the composite type last.
const fn f(name: &'static str, args: &'static [u32], implementation: super::Implementation) -> Function {
    Function { name, args, ret: ANYELEMENT, strict: false, variadic: false, implementation }
}

/// to declares a strict function that builds records of the columns of a function scan's column definition list.
const fn to(name: &'static str, args: &'static [u32], implementation: super::Implementation) -> Function {
    Function { name, args, ret: crate::oid::RECORD, strict: true, variadic: false, implementation }
}

/// FUNCTIONS are the populate functions.
pub const FUNCTIONS: &[Function] = &[
    f("json_populate_record", &[ANYELEMENT, JSON], |ctx, args| record(ctx, args, "json_populate_record")),
    f("json_populate_record", &[ANYELEMENT, JSON, BOOL], |ctx, args| record(ctx, args, "json_populate_record")),
    f("jsonb_populate_record", &[ANYELEMENT, JSONB], |ctx, args| record(ctx, args, "jsonb_populate_record")),
    f("json_populate_recordset", &[ANYELEMENT, JSON], |ctx, args| recordset(ctx, args, "json_populate_recordset")),
    f("json_populate_recordset", &[ANYELEMENT, JSON, BOOL], |ctx, args| {
        recordset(ctx, args, "json_populate_recordset")
    }),
    f("jsonb_populate_recordset", &[ANYELEMENT, JSONB], |ctx, args| recordset(ctx, args, "jsonb_populate_recordset")),
    to("json_to_record", &[JSON], |ctx, args| to_record(ctx, args, "json_to_record")),
    to("jsonb_to_record", &[JSONB], |ctx, args| to_record(ctx, args, "jsonb_to_record")),
    to("json_to_recordset", &[JSON], |ctx, args| to_record(ctx, args, "json_to_recordset")),
    to("jsonb_to_recordset", &[JSONB], |ctx, args| to_record(ctx, args, "jsonb_to_recordset")),
];

/// NAMES are the functions whose binder passes the composite type as a last argument.
pub const NAMES: [&str; 4] =
    ["json_populate_record", "jsonb_populate_record", "json_populate_recordset", "jsonb_populate_recordset"];

/// Js is a JSON value of json input, which keeps the text of each part as written, or of jsonb input.
#[derive(Clone, Copy)]
enum Js<'a> {
    Raw(&'a Raw<'a>),
    Jsonb(&'a Json),
}

impl<'a> Js<'a> {
    /// is_null reports whether the value is JSON null.
    fn is_null(self) -> bool {
        matches!(self, Js::Raw(Raw { kind: RawKind::Scalar(Json::Null), .. }) | Js::Jsonb(Json::Null))
    }

    /// is_object reports whether the value is an object.
    fn is_object(self) -> bool {
        matches!(self, Js::Raw(Raw { kind: RawKind::Object(_), .. }) | Js::Jsonb(Json::Object(_)))
    }

    /// elements returns an array's elements, or None for any other value.
    fn elements(self) -> Option<Vec<Js<'a>>> {
        match self {
            Js::Raw(Raw { kind: RawKind::Array(values), .. }) => Some(values.iter().map(Js::Raw).collect()),
            Js::Jsonb(Json::Array(values)) => Some(values.iter().map(Js::Jsonb).collect()),
            _ => None,
        }
    }

    /// get returns an object's value for a key.
    fn get(self, key: &str) -> Option<Js<'a>> {
        match self {
            Js::Raw(raw) => raw.get(key).map(Js::Raw),
            Js::Jsonb(json) => json.get(key).map(Js::Jsonb),
        }
    }

    /// text returns a string's contents, or the value's text otherwise.
    fn text(self) -> String {
        match self {
            Js::Raw(Raw { kind: RawKind::Scalar(Json::String(s)), .. }) | Js::Jsonb(Json::String(s)) => s.clone(),
            Js::Raw(raw) => raw.text.to_string(),
            Js::Jsonb(json) => json.to_text(),
        }
    }

    /// json_text returns the value's JSON text.
    fn json_text(self) -> String {
        match self {
            Js::Raw(raw) => raw.text.to_string(),
            Js::Jsonb(json) => json.to_text(),
        }
    }

    /// kind_error returns the error of calling a function that takes an object on another value.
    fn kind_error(self, function: &str) -> PgError {
        let what = if self.elements().is_some() { "an array" } else { "a scalar" };
        PgError::new(code::INVALID_PARAMETER_VALUE, format!("cannot call {function} on {what}"))
    }
}

/// composite_type returns the attributes of the composite type a populate function builds, which for an anonymous
/// record are its base record's fields, failing as Postgres does for a type that is not a composite one.
fn composite_type(
    ctx: &Ctx<'_>,
    type_oid: u32,
    base_record: &Value,
    function: &str,
) -> Result<Vec<(String, ColumnType)>> {
    let base = crate::usertypes::base_type(typ(type_oid));
    match crate::usertypes::get(base.oid).map(|t| t.kind.clone()) {
        Some(Kind::Composite(attributes)) => Ok(attributes),
        _ if let Value::Record(fields) = base_record => {
            Ok(fields.iter().enumerate().map(|(i, v)| (format!("f{}", i + 1), typ(super::value_type(v)))).collect())
        }
        _ if base.oid == crate::oid::RECORD
            && let Some(columns) = &ctx.session.expected_columns =>
        {
            Ok(columns.clone())
        }
        _ if base.oid == crate::oid::RECORD => Err(PgError {
            hint: Some(
                "Provide a non-null record argument, or call the function in the FROM clause using a column \
                 definition list."
                    .into(),
            ),
            ..PgError::new(
                code::FEATURE_NOT_SUPPORTED,
                format!("could not determine row type for result of {function}"),
            )
        }),
        _ => Err(PgError::new(code::DATATYPE_MISMATCH, format!("first argument of {function} must be a row type"))),
    }
}

/// with_input parses a populate function's JSON argument and runs a function on it.
fn with_input<T>(input: &Value, run: &mut dyn FnMut(Js<'_>) -> Result<T>) -> Result<T> {
    match input {
        Value::Jsonb(json) => run(Js::Jsonb(json)),
        other => {
            let text = other.output().unwrap_or_default();
            let raw = crate::json::parse_raw(&text)?;
            run(Js::Raw(&raw))
        }
    }
}

/// record builds a value of a composite type from a JSON object, taking the base record's fields for missing keys.
fn record(ctx: &mut Ctx<'_>, args: &[Value], function: &str) -> Result<Value> {
    let type_oid = type_argument(args);
    let attributes = composite_type(ctx, type_oid, &args[0], function)?;
    let (base, input) = (&args[0], &args[1]);
    if input.is_null() {
        return Ok(base.clone());
    }
    with_input(input, &mut |js| {
        if js.is_null() {
            return Ok(base.clone());
        }
        if !js.is_object() {
            return Err(js.kind_error("populate_composite"));
        }
        populate_object(ctx, type_oid, &attributes, js, base)
    })
}

/// recordset builds a value of a composite type from each object of a JSON array.
fn recordset(ctx: &mut Ctx<'_>, args: &[Value], function: &str) -> Result<Value> {
    let type_oid = type_argument(args);
    let attributes = composite_type(ctx, type_oid, &args[0], function)?;
    let (base, input) = (&args[0], &args[1]);
    if input.is_null() {
        return Ok(Value::Set(Vec::new()));
    }
    with_input(input, &mut |js| {
        let Some(elements) = js.elements() else {
            let what = if js.is_object() { "an object" } else { "a scalar" };
            return Err(PgError::new(code::INVALID_PARAMETER_VALUE, format!("cannot call {function} on {what}")));
        };
        let mut rows = Vec::with_capacity(elements.len());
        for element in elements {
            if !element.is_object() {
                return Err(PgError::new(
                    code::INVALID_PARAMETER_VALUE,
                    format!("argument of {function} must be an array of objects"),
                ));
            }
            rows.push(populate_object(ctx, type_oid, &attributes, element, base)?);
        }
        Ok(Value::Set(rows))
    })
}

/// to_record builds records of the columns of a function scan's column definition list from a JSON object, or a
/// set of them from an array of objects, as Postgres' json_to_record and json_to_recordset functions do.
fn to_record(ctx: &mut Ctx<'_>, args: &[Value], function: &str) -> Result<Value> {
    let args = [Value::Null, args[0].clone(), Value::Int8(i64::from(crate::oid::RECORD))];
    match function.ends_with("set") {
        true => recordset(ctx, &args, function),
        false => record(ctx, &args, function),
    }
}

/// type_argument returns the composite type that the binder passes last.
fn type_argument(args: &[Value]) -> u32 {
    match args.last() {
        Some(Value::Int8(oid)) => *oid as u32,
        _ => crate::oid::RECORD,
    }
}

/// populate_object builds a value of a composite type from an object, taking each field the object leaves out from a
/// default record.
fn populate_object(
    ctx: &mut Ctx<'_>,
    type_oid: u32,
    attributes: &[(String, ColumnType)],
    object: Js<'_>,
    default: &Value,
) -> Result<Value> {
    let defaults = match default {
        Value::Composite(composite) => Some(composite.fields.clone()),
        Value::Record(fields) => Some(fields.clone()),
        _ => None,
    };
    let base = crate::usertypes::base_type(typ(type_oid)).oid;
    let mut fields = Vec::with_capacity(attributes.len());
    for (i, (name, ty)) in attributes.iter().enumerate() {
        let default = defaults.as_ref().and_then(|d| d.get(i).cloned()).unwrap_or(Value::Null);
        fields.push(match (object.get(name), &defaults) {
            (None, Some(_)) => default,
            (value, _) => populate_field(ctx, *ty, value, name, &default)?,
        });
    }
    if crate::usertypes::get(type_oid).is_none() {
        return Ok(Value::Record(fields));
    }
    let value = Value::Composite(Box::new(CompositeValue { type_oid: base, fields }));
    if base != type_oid {
        ctx.check_domain(&value, typ(type_oid))?;
    }
    Ok(value)
}

/// populate_field converts a JSON value to a value of a field's type, as Postgres' populate_record_field does.
fn populate_field(ctx: &mut Ctx<'_>, ty: ColumnType, js: Option<Js<'_>>, name: &str, default: &Value) -> Result<Value> {
    let kind = crate::usertypes::get(ty.oid).map(|t| t.kind.clone());
    if let Some(Kind::Domain(domain)) = &kind {
        let value = match js.filter(|j| !j.is_null()) {
            Some(js) => populate_field(ctx, domain.base, Some(js), name, default)?,
            None => Value::Null,
        };
        ctx.check_domain(&value, ty)?;
        return Ok(value);
    }
    let Some(js) = js.filter(|j| !j.is_null()) else { return Ok(Value::Null) };
    if let Some(Kind::Composite(attributes)) = &kind {
        if let Js::Raw(Raw { kind: RawKind::Scalar(Json::String(text)), .. }) | Js::Jsonb(Json::String(text)) = js {
            return crate::cast::input(text, ty.oid);
        }
        if !js.is_object() {
            return Err(js.kind_error("populate_composite"));
        }
        return populate_object(ctx, ty.oid, attributes, js, default);
    }
    if crate::array::is_array_type(ty.oid) {
        return populate_array(ctx, ty, js, name);
    }
    match ty.oid {
        JSON => Ok(Value::Json(js.json_text())),
        JSONB => Ok(Value::Jsonb(Box::new(crate::json::parse(&js.json_text(), true)?))),
        _ => {
            let value = crate::cast::input_with_modifier(&js.text(), ty)?;
            crate::cast::cast_value(value, ty, false)
        }
    }
}

/// populate_array converts a JSON array, or an array literal in a string, to a value of an array type.
fn populate_array(ctx: &mut Ctx<'_>, ty: ColumnType, js: Js<'_>, name: &str) -> Result<Value> {
    let element = crate::expr::element_type(ty.oid);
    if let Js::Raw(Raw { kind: RawKind::Scalar(Json::String(text)), .. }) | Js::Jsonb(Json::String(text)) = js {
        return crate::cast::cast_value(crate::cast::input(text, ty.oid)?, ty, false);
    }
    let expected = |indices: &[usize]| {
        let hint = match indices.is_empty() {
            true => format!("See the value of key \"{name}\"."),
            false => {
                let path: String = indices.iter().map(|i| format!("[{i}]")).collect();
                format!("See the array element {path} of key \"{name}\".")
            }
        };
        PgError { hint: Some(hint), ..PgError::new(code::INVALID_TEXT_REPRESENTATION, "expected JSON array") }
    };
    let Some(top) = js.elements() else { return Err(expected(&[])) };
    let mut dims: Vec<usize> = Vec::new();
    let mut probe = top.clone();
    dims.push(probe.len());
    while let Some(first) = probe.first().copied() {
        match first.elements() {
            Some(inner) => {
                dims.push(inner.len());
                probe = inner;
            }
            None => break,
        }
    }
    let malformed = || PgError {
        detail: Some("Multidimensional arrays must have sub-arrays with matching dimensions.".into()),
        ..PgError::new(code::INVALID_TEXT_REPRESENTATION, "malformed JSON array")
    };
    let mut values = Vec::new();
    let element_type = ColumnType { oid: element, modifier: ty.modifier };
    let mut stack: Vec<(Vec<Js<'_>>, usize)> = vec![(top, 0)];
    let mut indices = vec![0usize];
    while let Some((items, position)) = stack.last_mut() {
        let depth = indices.len() - 1;
        if *position == items.len() {
            if items.len() != dims[depth] {
                return Err(malformed());
            }
            stack.pop();
            indices.pop();
            if let Some(last) = indices.last_mut() {
                *last += 1;
            }
            continue;
        }
        let item = items[*position];
        *position += 1;
        if depth + 1 < dims.len() {
            match item.elements() {
                Some(inner) => {
                    stack.push((inner, 0));
                    indices.push(0);
                }
                None => return Err(expected(&indices[..=depth])),
            }
        } else {
            values.push(populate_field(ctx, element_type, Some(item), name, &Value::Null)?);
            if let Some(last) = indices.last_mut() {
                *last += 1;
            }
        }
    }
    let dims = if values.is_empty() { Vec::new() } else { dims.iter().map(|&n| (n as i32, 1)).collect() };
    let array = crate::array::Array { element, dims, values };
    crate::cast::cast_value(Value::Array(Box::new(array)), ty, false)
}
