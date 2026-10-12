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

//! Postgres' SQL/JSON constructors and IS JSON predicate: JSON_OBJECT, JSON_ARRAY, JSON_OBJECTAGG, JSON_ARRAYAGG,
//! JSON(), JSON_SCALAR, and JSON_SERIALIZE, planned as parse_expr.c's transformJson functions plan them and computed
//! as ExecEvalJsonConstructor and ExecEvalJsonIsPredicate compute them.

use pg_query::protobuf::{
    CaseTestExpr, FuncCall, JsonArrayAgg, JsonArrayConstructor, JsonArrayQueryConstructor, JsonEncoding, JsonFormat,
    JsonFormatType, JsonIsPredicate, JsonObjectAgg, JsonObjectConstructor, JsonOutput, JsonParseExpr, JsonScalarExpr,
    JsonSerializeExpr, JsonValueExpr, JsonValueType,
};
use pg_query::{Node, NodeEnum};

use crate::catalog::ColumnType;
use crate::error::{PgError, Result, code};
use crate::expr::{Binder, Bound, EnclosingLevel, Expr, arg_location, coerce, is_string, typ};
use crate::functions;
use crate::json::{self, Json};
use crate::jsontable::{check_encoding, located};
use crate::oid;
use crate::types::Value;

/// Op is a SQL/JSON constructor or IS JSON predicate over its arguments.
#[derive(Clone, Debug, PartialEq)]
pub enum Op {
    /// JSON_OBJECT over alternating keys and values, building jsonb when set.
    Object { jsonb: bool, absent_on_null: bool, unique: bool },
    /// JSON_ARRAY over its elements.
    Array { jsonb: bool, absent_on_null: bool },
    /// JSON_SCALAR of a value.
    Scalar { jsonb: bool },
    /// JSON() WITH UNIQUE KEYS of text, which it parses, rejecting an object that repeats a key.
    Parse { jsonb: bool },
    /// IS JSON of a value of a type, testing the kind of its item and, when set, that no object repeats a key.
    Is { ty: u32, item: JsonValueType, unique: bool },
}

/// Returning is the type and format of a RETURNING clause, without a type when the clause is missing.
struct Returning {
    ty: Option<ColumnType>,
    format: JsonFormatType,
}

/// format_of returns a JSON format clause's type, treating a missing clause as the default format.
fn format_of(format: Option<&JsonFormat>) -> JsonFormatType {
    match format.map(|f| JsonFormatType::try_from(f.format_type)) {
        Some(Ok(JsonFormatType::JsFormatJson)) => JsonFormatType::JsFormatJson,
        Some(Ok(JsonFormatType::JsFormatJsonb)) => JsonFormatType::JsFormatJsonb,
        _ => JsonFormatType::JsFormatDefault,
    }
}

/// has_encoding reports whether a JSON format clause names an encoding.
fn has_encoding(format: Option<&JsonFormat>) -> bool {
    format.is_some_and(|f| {
        !matches!(JsonEncoding::try_from(f.encoding), Ok(JsonEncoding::JsEncDefault | JsonEncoding::Undefined))
    })
}

/// json_type returns json, or jsonb when set.
fn json_type(jsonb: bool) -> ColumnType {
    typ(if jsonb { oid::JSONB } else { oid::JSON })
}

/// call returns a call of a built-in function of the argument types.
fn call(name: &str, types: &[u32], args: Vec<Expr>, location: i32) -> Result<Expr> {
    Ok(Expr::Func(functions::resolve(name, types, location)?.index, args))
}

/// from_bytea decodes UTF8 JSON text from bytea, as makeJsonByteaToTextConversion does.
fn from_bytea(expr: Expr, location: i32) -> Result<Bound> {
    let encoding = Expr::Const(Value::Text("UTF8".into()));
    Ok((call("convert_from", &[oid::BYTEA, oid::NAME], vec![expr, encoding], location)?, typ(oid::TEXT)))
}

/// value_expr binds a JSON value expression as transformJsonValueExpr does: text read as JSON in the format its
/// FORMAT clause or the default names, bytea decoded first, and json and jsonb values as they are, each converted to
/// the target type when one is given.
fn value_expr(
    binder: &mut Binder<'_, '_>,
    construct: &str,
    value: &JsonValueExpr,
    default: JsonFormatType,
    target: Option<u32>,
) -> Result<Bound> {
    let node = value.raw_expr.as_deref().ok_or_else(|| PgError::internal(format!("{construct} without a value")))?;
    let location = arg_location(node);
    let mut bound = binder.bind(node)?;
    if bound.1.oid == oid::UNKNOWN {
        bound = coerce(bound, typ(oid::TEXT), false, location)?;
    }
    let ty = bound.1.oid;
    let format = value.format.as_ref();
    let format_location = format.map_or(-1, |f| f.location);
    let json = matches!(ty, oid::JSON | oid::JSONB);
    let format = match format_of(format) {
        JsonFormatType::JsFormatDefault => match json {
            true => JsonFormatType::JsFormatDefault,
            false => default,
        },
        _ if has_encoding(format) && ty != oid::BYTEA => {
            return Err(located(
                code::DATATYPE_MISMATCH,
                "JSON ENCODING clause is only allowed for bytea input type",
                format_location,
            ));
        }
        _ if json => JsonFormatType::JsFormatDefault,
        explicit => explicit,
    };
    if format == JsonFormatType::JsFormatDefault && target.is_none_or(|t| t == ty) {
        return Ok(bound);
    }
    if target.is_none() && ty != oid::BYTEA && !is_string(ty) {
        let message = match format_of(value.format.as_ref()) {
            JsonFormatType::JsFormatDefault => "cannot use non-string types with implicit FORMAT JSON clause",
            _ => "cannot use non-string types with explicit FORMAT JSON clause",
        };
        return Err(located(
            code::DATATYPE_MISMATCH,
            message,
            if format_location >= 0 { format_location } else { location },
        ));
    }
    if format == JsonFormatType::JsFormatJson && ty == oid::BYTEA {
        bound = from_bytea(bound.0, location)?;
    }
    let to = target.unwrap_or(if format == JsonFormatType::JsFormatJsonb { oid::JSONB } else { oid::JSON });
    match coerce(bound.clone(), typ(to), true, location) {
        Ok(coerced) => Ok(coerced),
        Err(_) if target.is_some() => Err(located(
            code::CANNOT_COERCE,
            format!("cannot cast type {} to {}", crate::cast::type_display(bound.1.oid), crate::cast::type_display(to)),
            location,
        )),
        Err(_) => {
            let name = if to == oid::JSON { "to_json" } else { "to_jsonb" };
            Ok((call(name, &[bound.1.oid], vec![bound.0], location)?, typ(to)))
        }
    }
}

/// output binds a RETURNING clause as transformJsonOutput does, giving a clause without a FORMAT the format of its
/// type and checking an explicit FORMAT against the type.
fn output(binder: &mut Binder<'_, '_>, output: Option<&JsonOutput>, allow_format: bool) -> Result<Returning> {
    let Some(type_name) = output.and_then(|o| o.type_name.as_ref()) else {
        return Ok(Returning { ty: None, format: JsonFormatType::JsFormatDefault });
    };
    binder.ctx.prepare_type(type_name)?;
    let ty = crate::expr::resolve_type_name(type_name)?;
    if type_name.setof {
        return Err(PgError::new(
            code::FEATURE_NOT_SUPPORTED,
            "returning SETOF types is not supported in SQL/JSON functions",
        ));
    }
    if crate::catalog::builtin_type(ty.oid).is_some_and(|t| t.definition.typ_type == b"p") {
        return Err(PgError::new(
            code::FEATURE_NOT_SUPPORTED,
            "returning pseudo-types is not supported in SQL/JSON functions",
        ));
    }
    let format = output.and_then(|o| o.returning.as_ref()).and_then(|r| r.format.as_ref());
    let format = match (format_of(format), format) {
        (JsonFormatType::JsFormatDefault, _) | (_, None) => match ty.oid {
            oid::JSONB => JsonFormatType::JsFormatJsonb,
            _ => JsonFormatType::JsFormatJson,
        },
        (explicit, Some(clause)) => {
            if !allow_format && !matches!(ty.oid, oid::BYTEA | oid::JSON | oid::JSONB) && !is_string(ty.oid) {
                return Err(located(
                    code::FEATURE_NOT_SUPPORTED,
                    "cannot use JSON format with non-string output types",
                    clause.location,
                ));
            }
            if explicit == JsonFormatType::JsFormatJson {
                check_encoding(clause, ty.oid, false)?;
            }
            explicit
        }
    };
    Ok(Returning { ty: Some(ty), format })
}

/// constructor_output binds a constructor's RETURNING clause as transformJsonConstructorOutput does, returning jsonb
/// without a clause when an argument is jsonb, and json otherwise.
fn constructor_output(binder: &mut Binder<'_, '_>, clause: Option<&JsonOutput>, types: &[u32]) -> Result<Returning> {
    let mut returning = output(binder, clause, true)?;
    if returning.ty.is_none() {
        let jsonb = types.contains(&oid::JSONB);
        returning.ty = Some(json_type(jsonb));
        returning.format = if jsonb { JsonFormatType::JsFormatJsonb } else { JsonFormatType::JsFormatJson };
    }
    Ok(returning)
}

/// query_returning binds the RETURNING clause of JSON(), JSON_SCALAR(), or JSON_SERIALIZE() as transformJsonReturning
/// does, which must name json or jsonb and defaults to json.
fn query_returning(binder: &mut Binder<'_, '_>, clause: Option<&JsonOutput>, name: &str) -> Result<Returning> {
    let Some(clause) = clause else {
        return Ok(Returning { ty: Some(typ(oid::JSON)), format: JsonFormatType::JsFormatJson });
    };
    let returning = output(binder, Some(clause), false)?;
    let ty = returning.ty.map_or(oid::JSON, |t| t.oid);
    if !matches!(ty, oid::JSON | oid::JSONB) {
        return Err(PgError {
            hint: Some("Try returning json or jsonb.".into()),
            ..located(
                code::DATATYPE_MISMATCH,
                format!("cannot use type {} in RETURNING clause of {name}", crate::cast::type_display(ty)),
                clause.type_name.as_ref().map_or(-1, |t| t.location),
            )
        });
    }
    Ok(returning)
}

/// coerce_output converts a constructor's json or jsonb result to its RETURNING type as coerceJsonFuncExpr does,
/// writing bytea in UTF8 when the clause asks for JSON format.
fn coerce_output(bound: Bound, returning: &Returning, location: i32) -> Result<Bound> {
    let Some(ty) = returning.ty else { return Ok(bound) };
    if ty.oid == bound.1.oid {
        return Ok(bound);
    }
    if returning.format == JsonFormatType::JsFormatJson && ty.oid == oid::BYTEA {
        let text = crate::expr::assign(bound, typ(oid::TEXT), "", location)?;
        let encoding = Expr::Const(Value::Text("UTF8".into()));
        return Ok((call("convert_to", &[oid::TEXT, oid::NAME], vec![text.0, encoding], location)?, ty));
    }
    if !crate::expr::assignable(bound.1.oid, crate::usertypes::base_type(ty).oid) {
        return Err(located(
            code::CANNOT_COERCE,
            format!(
                "cannot cast type {} to {}",
                crate::cast::type_display(bound.1.oid),
                crate::cast::type_display(ty.oid)
            ),
            location,
        ));
    }
    let base = crate::usertypes::base_type(ty);
    let converted = crate::expr::assign(bound, base, "", location)?;
    match base == ty {
        true => Ok(converted),
        false => Ok((Expr::Cast(Box::new(converted.0), ty, false), ty)),
    }
}

/// placeholder returns a node that stands for an expression already bound, which a call built from raw nodes binds.
fn placeholder(binder: &mut Binder<'_, '_>, bound: Bound) -> Node {
    binder.placeholders.push(bound);
    let index = binder.placeholders.len() - 1;
    Node { node: Some(NodeEnum::CaseTestExpr(Box::new(CaseTestExpr { type_mod: index as i32, ..Default::default() }))) }
}

/// aggregate binds a call of a built-in JSON aggregate over bound arguments, with a constructor's ORDER BY, FILTER,
/// and OVER clauses, and converts its result to the RETURNING type.
fn aggregate(
    binder: &mut Binder<'_, '_>,
    name: &str,
    args: Vec<Bound>,
    ctor: &pg_query::protobuf::JsonAggConstructor,
    returning: &Returning,
) -> Result<Bound> {
    let string = |s: &str| Node { node: Some(NodeEnum::String(pg_query::protobuf::String { sval: s.into() })) };
    let args = args.into_iter().map(|a| placeholder(binder, a)).collect();
    let call = FuncCall {
        funcname: vec![string("pg_catalog"), string(name)],
        args,
        agg_order: ctor.agg_order.clone(),
        agg_filter: ctor.agg_filter.clone(),
        over: ctor.over.clone(),
        location: ctor.location,
        ..Default::default()
    };
    let bound = binder.bind(&Node { node: Some(NodeEnum::FuncCall(Box::new(call))) })?;
    coerce_output(bound, returning, ctor.location)
}

/// plan_object binds JSON_OBJECT as transformJsonObjectConstructor does.
pub fn plan_object(binder: &mut Binder<'_, '_>, ctor: &JsonObjectConstructor) -> Result<Bound> {
    let (mut args, mut types) = (Vec::new(), Vec::new());
    for node in &ctor.exprs {
        let Some(NodeEnum::JsonKeyValue(pair)) = node.node.as_ref() else { continue };
        let key = pair.key.as_deref().ok_or_else(|| PgError::internal("a JSON_OBJECT key without an expression"))?;
        let key = binder.bind(key)?;
        let value = pair.value.as_deref().ok_or_else(|| PgError::internal("a JSON_OBJECT key without a value"))?;
        let value = value_expr(binder, "JSON_OBJECT()", value, JsonFormatType::JsFormatDefault, None)?;
        types.extend([key.1.oid, value.1.oid]);
        args.extend([key.0, value.0]);
    }
    let returning = constructor_output(binder, ctor.output.as_ref(), &types)?;
    let jsonb = returning.format == JsonFormatType::JsFormatJsonb;
    let op = Op::Object { jsonb, absent_on_null: ctor.absent_on_null, unique: ctor.unique };
    coerce_output((Expr::SqlJson(op, args), json_type(jsonb)), &returning, ctor.location)
}

/// plan_array binds JSON_ARRAY of values as transformJsonArrayConstructor does.
pub fn plan_array(binder: &mut Binder<'_, '_>, ctor: &JsonArrayConstructor) -> Result<Bound> {
    let (mut args, mut types) = (Vec::new(), Vec::new());
    for node in &ctor.exprs {
        let Some(NodeEnum::JsonValueExpr(value)) = node.node.as_ref() else { continue };
        let (expr, ty) = value_expr(binder, "JSON_ARRAY()", value, JsonFormatType::JsFormatDefault, None)?;
        types.push(ty.oid);
        args.push(expr);
    }
    let returning = constructor_output(binder, ctor.output.as_ref(), &types)?;
    let jsonb = returning.format == JsonFormatType::JsFormatJsonb;
    let op = Op::Array { jsonb, absent_on_null: ctor.absent_on_null };
    coerce_output((Expr::SqlJson(op, args), json_type(jsonb)), &returning, ctor.location)
}

/// plan_array_query binds JSON_ARRAY of a query as transformJsonArrayQueryConstructor does, as a subquery
/// `(SELECT JSON_ARRAYAGG(a) FROM (query) q(a))` of the query's only column.
pub fn plan_array_query(binder: &mut Binder<'_, '_>, ctor: &JsonArrayQueryConstructor) -> Result<Bound> {
    use pg_query::protobuf::{
        Alias, ColumnRef, JsonAggConstructor, RangeSubselect, ResTarget, SelectStmt, SubLink, SubLinkType,
    };
    let query = ctor.query.as_deref().ok_or_else(|| PgError::internal("JSON_ARRAY without a query"))?;
    let Some(NodeEnum::SelectStmt(select)) = query.node.as_ref() else {
        return Err(PgError::internal("JSON_ARRAY of a query that is not a SELECT"));
    };
    let reach = binder.ctx.outer_reach;
    let level = EnclosingLevel {
        level: binder.scopes.len() - 1,
        aggregates: None,
        clause: binder.clause,
        outer_exprs: Vec::new(),
    };
    binder.ctx.enclosing_levels.push(level);
    let planned = crate::plan::Planner { ctx: &mut *binder.ctx, outer: binder.scopes.clone() }.plan_query(select);
    binder.ctx.enclosing_levels.pop();
    binder.ctx.outer_reach = reach;
    if planned?.columns.len() != 1 {
        return Err(located(code::SYNTAX_ERROR, "subquery must return only one column", ctor.location));
    }
    let string = |s: &str| Node { node: Some(NodeEnum::String(pg_query::protobuf::String { sval: s.into() })) };
    let column = ColumnRef { fields: vec![string("q"), string("a")], location: ctor.location };
    let column = Node { node: Some(NodeEnum::ColumnRef(column)) };
    let constructor = JsonAggConstructor { output: ctor.output.clone(), location: ctor.location, ..Default::default() };
    let agg = JsonArrayAgg {
        constructor: Some(Box::new(constructor)),
        arg: Some(Box::new(JsonValueExpr {
            raw_expr: Some(Box::new(column.clone())),
            formatted_expr: Some(Box::new(column)),
            format: ctor.format,
        })),
        absent_on_null: ctor.absent_on_null,
    };
    let target = ResTarget {
        val: Some(Box::new(Node { node: Some(NodeEnum::JsonArrayAgg(Box::new(agg))) })),
        location: ctor.location,
        ..Default::default()
    };
    let range = RangeSubselect {
        subquery: ctor.query.clone(),
        alias: Some(Alias { aliasname: "q".into(), colnames: vec![string("a")] }),
        ..Default::default()
    };
    let select = SelectStmt {
        target_list: vec![Node { node: Some(NodeEnum::ResTarget(Box::new(target))) }],
        from_clause: vec![Node { node: Some(NodeEnum::RangeSubselect(Box::new(range))) }],
        ..Default::default()
    };
    let link = SubLink {
        sub_link_type: SubLinkType::ExprSublink as i32,
        subselect: Some(Box::new(Node { node: Some(NodeEnum::SelectStmt(Box::new(select))) })),
        location: ctor.location,
        ..Default::default()
    };
    binder.bind(&Node { node: Some(NodeEnum::SubLink(Box::new(link))) })
}

/// plan_object_agg binds JSON_OBJECTAGG as transformJsonObjectAgg does, as a call of json_object_agg or one of its
/// strict, unique, or jsonb variants.
pub fn plan_object_agg(binder: &mut Binder<'_, '_>, agg: &JsonObjectAgg) -> Result<Bound> {
    let ctor = agg.constructor.as_deref().ok_or_else(|| PgError::internal("JSON_OBJECTAGG without a constructor"))?;
    let pair = agg.arg.as_deref().ok_or_else(|| PgError::internal("JSON_OBJECTAGG without a key and value"))?;
    let key = pair.key.as_deref().ok_or_else(|| PgError::internal("a JSON_OBJECTAGG key without an expression"))?;
    let key = binder.bind(key)?;
    let value = pair.value.as_deref().ok_or_else(|| PgError::internal("a JSON_OBJECTAGG key without a value"))?;
    let value = value_expr(binder, "JSON_OBJECTAGG()", value, JsonFormatType::JsFormatDefault, None)?;
    let returning = constructor_output(binder, ctor.output.as_ref(), &[key.1.oid, value.1.oid])?;
    let mut name = String::from(match returning.format {
        JsonFormatType::JsFormatJsonb => "jsonb_object_agg",
        _ => "json_object_agg",
    });
    if agg.unique {
        name.push_str("_unique");
    }
    if agg.absent_on_null {
        name.push_str("_strict");
    }
    aggregate(binder, &name, vec![key, value], ctor, &returning)
}

/// plan_array_agg binds JSON_ARRAYAGG as transformJsonArrayAgg does, as a call of json_agg or one of its strict or
/// jsonb variants.
pub fn plan_array_agg(binder: &mut Binder<'_, '_>, agg: &JsonArrayAgg) -> Result<Bound> {
    let ctor = agg.constructor.as_deref().ok_or_else(|| PgError::internal("JSON_ARRAYAGG without a constructor"))?;
    let value = agg.arg.as_deref().ok_or_else(|| PgError::internal("JSON_ARRAYAGG without a value"))?;
    let value = value_expr(binder, "JSON_ARRAYAGG()", value, JsonFormatType::JsFormatDefault, None)?;
    let returning = constructor_output(binder, ctor.output.as_ref(), &[value.1.oid])?;
    let name = match (returning.format == JsonFormatType::JsFormatJsonb, agg.absent_on_null) {
        (true, true) => "jsonb_agg_strict",
        (true, false) => "jsonb_agg",
        (false, true) => "json_agg_strict",
        (false, false) => "json_agg",
    };
    aggregate(binder, name, vec![value], ctor, &returning)
}

/// parse_arg binds the input of JSON() WITH UNIQUE KEYS or IS JSON as transformJsonParseArg does: bytea decoded as
/// UTF8 text, and string types converted to text.
fn parse_arg(binder: &mut Binder<'_, '_>, node: &Node, format: Option<&JsonFormat>) -> Result<Bound> {
    let location = arg_location(node);
    let bound = binder.bind(node)?;
    let ty = bound.1.oid;
    if ty == oid::BYTEA {
        return from_bytea(bound.0, location);
    }
    let bound = match ty == oid::UNKNOWN || is_string(ty) {
        true => coerce(bound, typ(oid::TEXT), false, -1).map_err(|_| {
            located(
                code::CANNOT_COERCE,
                format!("cannot cast type {} to text", crate::cast::type_display(ty)),
                location,
            )
        })?,
        false => bound,
    };
    if has_encoding(format) {
        return Err(located(
            code::FEATURE_NOT_SUPPORTED,
            "cannot use JSON FORMAT ENCODING clause for non-bytea input types",
            format.map_or(-1, |f| f.location),
        ));
    }
    Ok(bound)
}

/// plan_parse binds JSON() as transformJsonParseExpr does: text read as json or jsonb by a cast, or parsed checking
/// keys WITH UNIQUE KEYS.
pub fn plan_parse(binder: &mut Binder<'_, '_>, expr: &JsonParseExpr) -> Result<Bound> {
    let returning = query_returning(binder, expr.output.as_ref(), "JSON()")?;
    let ty = returning.ty.unwrap_or(typ(oid::JSON));
    let value = expr.expr.as_deref().ok_or_else(|| PgError::internal("JSON() without a value"))?;
    if !expr.unique_keys {
        return value_expr(binder, "JSON()", value, JsonFormatType::JsFormatJson, Some(ty.oid));
    }
    let node = value.raw_expr.as_deref().ok_or_else(|| PgError::internal("JSON() without a value"))?;
    let (arg, arg_type) = parse_arg(binder, node, value.format.as_ref())?;
    if arg_type.oid != oid::TEXT {
        return Err(located(
            code::DATATYPE_MISMATCH,
            "cannot use non-string types with WITH UNIQUE KEYS clause",
            expr.location,
        ));
    }
    Ok((Expr::SqlJson(Op::Parse { jsonb: ty.oid == oid::JSONB }, vec![arg]), ty))
}

/// plan_scalar binds JSON_SCALAR as transformJsonScalarExpr does.
pub fn plan_scalar(binder: &mut Binder<'_, '_>, expr: &JsonScalarExpr) -> Result<Bound> {
    let node = expr.expr.as_deref().ok_or_else(|| PgError::internal("JSON_SCALAR without a value"))?;
    let mut arg = binder.bind(node)?;
    let returning = query_returning(binder, expr.output.as_ref(), "JSON_SCALAR()")?;
    if arg.1.oid == oid::UNKNOWN {
        arg = coerce(arg, typ(oid::TEXT), false, arg_location(node))?;
    }
    let ty = returning.ty.unwrap_or(typ(oid::JSON));
    Ok((Expr::SqlJson(Op::Scalar { jsonb: ty.oid == oid::JSONB }, vec![arg.0]), ty))
}

/// plan_serialize binds JSON_SERIALIZE as transformJsonSerializeExpr does, as its JSON value converted to the
/// RETURNING type, a string type or bytea that defaults to text.
pub fn plan_serialize(binder: &mut Binder<'_, '_>, expr: &JsonSerializeExpr) -> Result<Bound> {
    let value = expr.expr.as_deref().ok_or_else(|| PgError::internal("JSON_SERIALIZE without a value"))?;
    let arg = value_expr(binder, "JSON_SERIALIZE()", value, JsonFormatType::JsFormatJson, None)?;
    let returning = match expr.output.as_ref() {
        Some(clause) => {
            let returning = output(binder, Some(clause), true)?;
            let ty = returning.ty.unwrap_or(typ(oid::TEXT));
            if ty.oid != oid::BYTEA && !is_string(crate::usertypes::base_type(ty).oid) {
                return Err(PgError {
                    hint: Some("Try returning a string type or bytea.".into()),
                    ..PgError::new(
                        code::DATATYPE_MISMATCH,
                        format!(
                            "cannot use type {} in RETURNING clause of JSON_SERIALIZE()",
                            crate::cast::type_display(ty.oid)
                        ),
                    )
                });
            }
            returning
        }
        None => Returning { ty: Some(typ(oid::TEXT)), format: JsonFormatType::JsFormatJson },
    };
    coerce_output(arg, &returning, expr.location)
}

/// plan_is binds IS JSON as transformJsonIsPredicate does.
pub fn plan_is(binder: &mut Binder<'_, '_>, pred: &JsonIsPredicate) -> Result<Bound> {
    let node = pred.expr.as_deref().ok_or_else(|| PgError::internal("IS JSON without a value"))?;
    let (arg, ty) = parse_arg(binder, node, pred.format.as_ref())?;
    if !matches!(ty.oid, oid::TEXT | oid::JSON | oid::JSONB) {
        return Err(PgError::new(
            code::DATATYPE_MISMATCH,
            format!("cannot use type {} in IS JSON predicate", crate::cast::type_display(ty.oid)),
        ));
    }
    let item = JsonValueType::try_from(pred.item_type).unwrap_or(JsonValueType::JsTypeAny);
    Ok((Expr::SqlJson(Op::Is { ty: ty.oid, item, unique: pred.unique_keys }, vec![arg]), typ(oid::BOOL)))
}

/// repeats_key reports whether an object of a JSON value, at any depth, repeats a key.
fn repeats_key(value: &Json) -> bool {
    match value {
        Json::Object(items) => {
            let mut keys = std::collections::HashSet::new();
            items.iter().any(|(key, value)| !keys.insert(key.as_str()) || repeats_key(value))
        }
        Json::Array(values) => values.iter().any(repeats_key),
        _ => false,
    }
}

/// eval computes a SQL/JSON constructor or IS JSON predicate over the values of its arguments.
pub fn eval(op: &Op, args: Vec<Value>) -> Result<Value> {
    use crate::functions::json::{build_array_worker, build_object_worker, datum_json_text, datum_to_json};
    let first = args.first().cloned().unwrap_or(Value::Null);
    Ok(match op {
        Op::Object { jsonb, absent_on_null, unique } => build_object_worker(&args, *jsonb, *absent_on_null, *unique)?,
        Op::Array { jsonb, absent_on_null } => build_array_worker(&args, *jsonb, *absent_on_null)?,
        _ if first.is_null() => Value::Null,
        Op::Scalar { jsonb: true } => Value::Jsonb(Box::new(json::normalize(datum_to_json(&first)?))),
        Op::Scalar { jsonb: false } => Value::Json(datum_json_text(&first)?),
        Op::Parse { jsonb } => {
            let text = first.output().unwrap_or_default();
            let parsed = json::parse(&text, *jsonb)?;
            if repeats_key(&parsed) {
                return Err(PgError::new(code::DUPLICATE_JSON_OBJECT_KEY_VALUE, "duplicate JSON object key value"));
            }
            match jsonb {
                true => Value::Jsonb(Box::new(json::normalize(parsed))),
                false => Value::Json(text),
            }
        }
        Op::Is { item, unique, .. } => {
            let parsed = match first {
                Value::Jsonb(value) => Some(*value),
                other => json::parse(&other.output().unwrap_or_default(), false).ok(),
            };
            Value::Bool(parsed.is_some_and(|value| {
                let kind = match &value {
                    Json::Object(_) => JsonValueType::JsTypeObject,
                    Json::Array(_) => JsonValueType::JsTypeArray,
                    _ => JsonValueType::JsTypeScalar,
                };
                (*item == JsonValueType::JsTypeAny || *item == kind) && !(*unique && repeats_key(&value))
            }))
        }
    })
}
