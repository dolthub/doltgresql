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

//! JSON_TABLE: checking its paths, columns, and behaviors as Postgres 17 does, and producing one row per item that its
//! row path matches, joined with the rows of its nested paths.

use std::collections::{HashMap, HashSet};

use pg_query::NodeEnum;
use pg_query::protobuf::{
    JsonBehavior, JsonBehaviorType, JsonEncoding, JsonExprOp, JsonFormat, JsonFuncExpr, JsonQuotes,
    JsonTableColumnType, JsonWrapper,
};

use crate::catalog::ColumnType;
use crate::error::{PgError, Result, code};
use crate::expr::{Binder, Expr, arg_location, position};
use crate::json::Json;
use crate::jsonpath::exec::{Item, Options};
use crate::jsonpath::{self, JsonPath};
use crate::oid;
use crate::query::Ctx;
use crate::types::Value;

/// Behavior is what an ON EMPTY or ON ERROR clause makes a column, or the table, produce.
#[derive(Clone, Debug, PartialEq)]
pub enum Behavior {
    Null,
    Error,
    True,
    False,
    Unknown,
    EmptyArray,
    EmptyObject,
    Default(Expr),
}

/// Kind is what a column holds: a row number, a scalar item, a JSON item, or whether its path finds any item.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Kind {
    Ordinality,
    Value,
    Query,
    Exists,
}

/// Wrapper is whether a JSON column wraps the items its path finds in an array.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Wrapper {
    None,
    Conditional,
    Unconditional,
}

/// TableColumn is an output column of a JSON_TABLE.
#[derive(Clone, Debug, PartialEq)]
pub struct TableColumn {
    pub name: String,
    pub ty: ColumnType,
    pub kind: Kind,
    pub path: Option<JsonPath>,
    pub wrapper: Wrapper,
    pub omit_quotes: bool,
    pub on_empty: Behavior,
    pub on_error: Behavior,
}

/// TablePath is a row path of a JSON_TABLE: the columns it fills, by position, and the nested paths under it.
#[derive(Clone, Debug, PartialEq)]
pub struct TablePath {
    pub path: JsonPath,
    pub columns: Vec<usize>,
    pub nested: Vec<TablePath>,
}

/// Input is the context item or a PASSING value: its expression, its type, and whether its text reads as JSON, as the
/// context item's and a FORMAT JSON value's do.
#[derive(Clone, Debug, PartialEq)]
pub struct Input {
    pub expr: Expr,
    pub ty: u32,
    pub formatted: bool,
}

/// JsonTable is a planned JSON_TABLE.
#[derive(Clone, Debug, PartialEq)]
pub struct JsonTable {
    pub context: Input,
    pub passing: Vec<(String, Input)>,
    pub root: TablePath,
    pub columns: Vec<TableColumn>,
    pub error_on_error: bool,
}

/// JsonFunc is a planned JSON_EXISTS, JSON_QUERY, or JSON_VALUE, whose expression's arguments are its context item,
/// its path, and its PASSING values, in that order: the type and format of the context item and of each PASSING value,
/// the PASSING values' names, and the column that holds what the function returns, which has no name or path.
#[derive(Clone, Debug, PartialEq)]
pub struct JsonFunc {
    pub context: (u32, bool),
    pub passing: Vec<(String, u32, bool)>,
    pub column: TableColumn,
}

/// located returns an error at a node location.
pub(crate) fn located(code: &'static str, message: impl Into<String>, location: i32) -> PgError {
    PgError { position: position(location), ..PgError::new(code, message) }
}

/// behavior_type returns the type of an ON EMPTY or ON ERROR clause.
fn behavior_type(behavior: &JsonBehavior) -> JsonBehaviorType {
    JsonBehaviorType::try_from(behavior.btype).unwrap_or(JsonBehaviorType::JsonBehaviorNull)
}

/// check_encoding fails as Postgres does for a FORMAT JSON clause with an ENCODING it cannot read, or with one on a
/// type other than bytea.
pub(crate) fn check_encoding(format: &JsonFormat, ty: u32, input: bool) -> Result<()> {
    let encoding = JsonEncoding::try_from(format.encoding).unwrap_or(JsonEncoding::JsEncDefault);
    if matches!(encoding, JsonEncoding::JsEncDefault | JsonEncoding::Undefined) {
        return Ok(());
    }
    if ty != oid::BYTEA {
        return Err(if input {
            located(
                code::DATATYPE_MISMATCH,
                "JSON ENCODING clause is only allowed for bytea input type",
                format.location,
            )
        } else {
            located(code::FEATURE_NOT_SUPPORTED, "cannot set JSON encoding for non-bytea output types", format.location)
        });
    }
    if encoding != JsonEncoding::JsEncUtf8 {
        return Err(PgError {
            hint: Some("Only UTF8 JSON encoding is supported.".into()),
            ..located(code::FEATURE_NOT_SUPPORTED, "unsupported JSON encoding", format.location)
        });
    }
    Ok(())
}

/// is_formatted_type reports whether a column type makes a plain column hold JSON items rather than scalars: a JSON,
/// array, or composite type, or a domain over one.
fn is_formatted_type(ty: u32) -> bool {
    let base = crate::usertypes::base_type(ColumnType { oid: ty, modifier: -1 }).oid;
    matches!(base, oid::JSON | oid::JSONB) || crate::array::is_array_type(base) || crate::expr::is_composite(base)
}

/// exists_value converts whether an EXISTS column's path found an item to the column's type.
fn exists_value(exists: bool, ty: ColumnType) -> Result<Value> {
    match crate::usertypes::base_type(ty).oid {
        oid::BOOL => Ok(Value::Bool(exists)),
        oid::INT4 => Ok(Value::Int4(i32::from(exists))),
        _ => crate::cast::input(if exists { "true" } else { "false" }, ty.oid),
    }
}

/// Planning is the state of planning a JSON_TABLE's columns: the names seen so far and the columns planned.
struct Planning {
    names: HashSet<String>,
    columns: Vec<TableColumn>,
}

impl Planning {
    /// claim records a column or path name, failing when the JSON_TABLE already has it.
    fn claim(&mut self, name: &str, location: i32) -> Result<()> {
        if !self.names.insert(name.to_string()) {
            return Err(located(
                code::DUPLICATE_ALIAS,
                format!("duplicate JSON_TABLE column or path name: {name}"),
                location,
            ));
        }
        Ok(())
    }
}

/// path_of parses a path specification, which must be a string constant.
fn path_of(spec: Option<&pg_query::protobuf::JsonTablePathSpec>) -> Result<JsonPath> {
    let spec = spec.ok_or_else(|| PgError::internal("a JSON_TABLE path without a specification"))?;
    let node = spec.string.as_deref().ok_or_else(|| PgError::internal("a JSON_TABLE path without a string"))?;
    let Some(NodeEnum::AConst(constant)) = node.node.as_ref() else {
        return Err(located(
            code::FEATURE_NOT_SUPPORTED,
            "only string constants are supported in JSON_TABLE path specification",
            arg_location(node),
        ));
    };
    let Some(pg_query::protobuf::a_const::Val::Sval(text)) = &constant.val else {
        return Err(located(
            code::FEATURE_NOT_SUPPORTED,
            "only string constants are supported in JSON_TABLE path specification",
            constant.location,
        ));
    };
    jsonpath::parse(&text.sval).map_err(|err| PgError { position: position(constant.location), ..err })
}

/// bind_input binds the context item or a PASSING value.
fn bind_input(binder: &mut Binder<'_, '_>, value: &pg_query::protobuf::JsonValueExpr) -> Result<(Input, i32)> {
    let node = value.raw_expr.as_deref().ok_or_else(|| PgError::internal("a JSON value without an expression"))?;
    let (expr, ty) = binder.bind(node)?;
    let ty = crate::usertypes::base_type(ty).oid;
    let format = value.format.as_ref().filter(|f| {
        !matches!(
            pg_query::protobuf::JsonFormatType::try_from(f.format_type),
            Ok(pg_query::protobuf::JsonFormatType::JsFormatDefault | pg_query::protobuf::JsonFormatType::Undefined)
        )
    });
    if let Some(format) = format {
        check_encoding(format, ty, true)?;
    }
    Ok((Input { expr, ty, formatted: format.is_some() }, arg_location(node)))
}

/// behavior binds an ON EMPTY or ON ERROR clause of a column, checking that the column's kind allows it.
fn behavior(
    binder: &mut Binder<'_, '_>,
    clause: Option<&JsonBehavior>,
    column: &TableColumn,
    on_error: bool,
) -> Result<Behavior> {
    let Some(clause) = clause else {
        return Ok(if column.kind == Kind::Exists && on_error { Behavior::False } else { Behavior::Null });
    };
    let name = if on_error { "ERROR" } else { "EMPTY" };
    let kind = behavior_type(clause);
    let allowed = match column.kind {
        Kind::Exists => matches!(
            kind,
            JsonBehaviorType::JsonBehaviorError
                | JsonBehaviorType::JsonBehaviorTrue
                | JsonBehaviorType::JsonBehaviorFalse
                | JsonBehaviorType::JsonBehaviorUnknown
        ),
        Kind::Query => matches!(
            kind,
            JsonBehaviorType::JsonBehaviorError
                | JsonBehaviorType::JsonBehaviorNull
                | JsonBehaviorType::JsonBehaviorEmpty
                | JsonBehaviorType::JsonBehaviorEmptyArray
                | JsonBehaviorType::JsonBehaviorEmptyObject
                | JsonBehaviorType::JsonBehaviorDefault
        ),
        _ => matches!(
            kind,
            JsonBehaviorType::JsonBehaviorError
                | JsonBehaviorType::JsonBehaviorNull
                | JsonBehaviorType::JsonBehaviorDefault
        ),
    };
    if !allowed {
        let (allowed, of) = match (column.kind, column.name.is_empty()) {
            (Kind::Exists, function) => {
                ("ERROR, TRUE, FALSE, or UNKNOWN", if function { "JSON_EXISTS()" } else { "EXISTS columns" })
            }
            (Kind::Query, function) => (
                "ERROR, NULL, EMPTY ARRAY, EMPTY OBJECT, or DEFAULT expression",
                if function { "JSON_QUERY()" } else { "formatted columns" },
            ),
            (_, function) => {
                ("ERROR, NULL, or DEFAULT expression", if function { "JSON_VALUE()" } else { "scalar columns" })
            }
        };
        let message = match column.name.is_empty() {
            true => format!("invalid ON {name} behavior"),
            false => format!("invalid ON {name} behavior for column \"{}\"", column.name),
        };
        return Err(PgError {
            detail: Some(format!("Only {allowed} is allowed in ON {name} for {of}.")),
            ..located(code::SYNTAX_ERROR, message, clause.location)
        });
    }
    Ok(match kind {
        JsonBehaviorType::JsonBehaviorError => Behavior::Error,
        JsonBehaviorType::JsonBehaviorTrue => Behavior::True,
        JsonBehaviorType::JsonBehaviorFalse => Behavior::False,
        JsonBehaviorType::JsonBehaviorUnknown => Behavior::Unknown,
        JsonBehaviorType::JsonBehaviorEmpty | JsonBehaviorType::JsonBehaviorEmptyArray => Behavior::EmptyArray,
        JsonBehaviorType::JsonBehaviorEmptyObject => Behavior::EmptyObject,
        JsonBehaviorType::JsonBehaviorDefault => {
            let node = clause.expr.as_deref().ok_or_else(|| PgError::internal("a DEFAULT behavior without a value"))?;
            let location = arg_location(node);
            let (expr, ty) = binder.bind(node)?;
            if !valid_default(&expr) {
                return Err(located(
                    code::DATATYPE_MISMATCH,
                    "can only specify a constant, non-aggregate function, or operator expression for DEFAULT",
                    location,
                ));
            }
            let (mut reads_columns, mut returns_set) = (false, false);
            expr.visit(&mut |e| match e {
                Expr::Column(_) | Expr::Outer(..) => reads_columns = true,
                Expr::Func(f, _) => returns_set |= crate::functions::returns_set(crate::functions::function(*f).name),
                Expr::Routine(routine, _) => returns_set |= routine.set_of,
                _ => {}
            });
            if reads_columns {
                return Err(located(
                    code::DATATYPE_MISMATCH,
                    "DEFAULT expression must not contain column references",
                    location,
                ));
            }
            if returns_set {
                return Err(located(code::DATATYPE_MISMATCH, "DEFAULT expression must not return a set", location));
            }
            let base = crate::usertypes::base_type(column.ty).oid;
            let coerced = match crate::expr::is_string(base) || matches!(base, oid::BIT | oid::VARBIT) {
                true => crate::expr::assign((expr, ty), column.ty, &column.name, location),
                false => crate::expr::coerce((expr, ty), column.ty, true, location),
            };
            let shown = crate::cast::type_display(column.ty.oid);
            let expr = coerced.map_err(|err| match err.code {
                code::DATATYPE_MISMATCH | code::CANNOT_COERCE => PgError {
                    hint: Some(format!("You will need to explicitly cast the expression to type {shown}.")),
                    ..located(
                        code::CANNOT_COERCE,
                        format!(
                            "cannot cast behavior expression of type {} to {shown}",
                            crate::cast::type_display(ty.oid)
                        ),
                        location,
                    )
                },
                _ => err,
            })?;
            Behavior::Default(expr.0)
        }
        _ => Behavior::Null,
    })
}

/// columns plans the columns of one row path, returning the positions of the columns it fills and its nested paths.
fn columns(
    binder: &mut Binder<'_, '_>,
    planning: &mut Planning,
    nodes: &[pg_query::Node],
) -> Result<(Vec<usize>, Vec<TablePath>)> {
    let (mut own, mut nested) = (Vec::new(), Vec::new());
    let mut ordinality = false;
    for node in nodes {
        let Some(NodeEnum::JsonTableColumn(raw)) = node.node.as_ref() else { continue };
        let coltype = JsonTableColumnType::try_from(raw.coltype).unwrap_or(JsonTableColumnType::JtcRegular);
        if coltype == JsonTableColumnType::JtcNested {
            let spec = raw.pathspec.as_deref();
            if let Some(spec) = spec
                && !spec.name.is_empty()
            {
                planning.claim(&spec.name, spec.name_location)?;
            }
            let path = path_of(spec)?;
            let (columns, children) = columns(binder, planning, &raw.columns)?;
            nested.push(TablePath { path, columns, nested: children });
            continue;
        }
        planning.claim(&raw.name, raw.location)?;
        if coltype == JsonTableColumnType::JtcForOrdinality {
            if ordinality {
                return Err(located(code::SYNTAX_ERROR, "only one FOR ORDINALITY column is allowed", raw.location));
            }
            ordinality = true;
            own.push(planning.columns.len());
            planning.columns.push(TableColumn {
                name: raw.name.clone(),
                ty: ColumnType { oid: oid::INT4, modifier: -1 },
                kind: Kind::Ordinality,
                path: None,
                wrapper: Wrapper::None,
                omit_quotes: false,
                on_empty: Behavior::Null,
                on_error: Behavior::Null,
            });
            continue;
        }
        let type_name =
            raw.type_name.as_ref().ok_or_else(|| PgError::internal("a JSON_TABLE column without a type"))?;
        binder.ctx.prepare_type(type_name)?;
        let ty = crate::expr::resolve_type_name(type_name)?;
        let base = crate::usertypes::base_type(ty).oid;
        let wrapper = match JsonWrapper::try_from(raw.wrapper) {
            Ok(JsonWrapper::JswConditional) => Some(Wrapper::Conditional),
            Ok(JsonWrapper::JswUnconditional) => Some(Wrapper::Unconditional),
            Ok(JsonWrapper::JswNone) => Some(Wrapper::None),
            _ => None,
        };
        let quotes = JsonQuotes::try_from(raw.quotes).unwrap_or(JsonQuotes::JsQuotesUnspec);
        let quotes_given = !matches!(quotes, JsonQuotes::JsQuotesUnspec | JsonQuotes::Undefined);
        if quotes_given && matches!(wrapper, Some(Wrapper::Conditional | Wrapper::Unconditional)) {
            return Err(located(
                code::SYNTAX_ERROR,
                "SQL/JSON QUOTES behavior must not be specified when WITH WRAPPER is used",
                raw.location,
            ));
        }
        let format = raw.format.as_ref().filter(|f| {
            !matches!(
                pg_query::protobuf::JsonFormatType::try_from(f.format_type),
                Ok(pg_query::protobuf::JsonFormatType::JsFormatDefault | pg_query::protobuf::JsonFormatType::Undefined)
            )
        });
        if let Some(format) = format {
            if !matches!(base, oid::JSON | oid::JSONB | oid::BYTEA) && !crate::expr::is_string(base) {
                return Err(located(
                    code::FEATURE_NOT_SUPPORTED,
                    "cannot use JSON format with non-string output types",
                    format.location,
                ));
            }
            check_encoding(format, base, false)?;
        }
        let kind = if coltype == JsonTableColumnType::JtcExists {
            Kind::Exists
        } else if format.is_some() || wrapper.is_some() || quotes_given || is_formatted_type(ty.oid) {
            Kind::Query
        } else {
            Kind::Value
        };
        let path = match raw.pathspec.as_deref() {
            Some(spec) => path_of(Some(spec))?,
            None => {
                let mut quoted = String::new();
                crate::json::escape(&mut quoted, &raw.name);
                jsonpath::parse(&format!("$.{quoted}"))?
            }
        };
        let mut column = TableColumn {
            name: raw.name.clone(),
            ty,
            kind,
            path: Some(path),
            wrapper: wrapper.unwrap_or(Wrapper::None),
            omit_quotes: quotes == JsonQuotes::JsQuotesOmit,
            on_empty: Behavior::Null,
            on_error: Behavior::Null,
        };
        column.on_empty = behavior(binder, raw.on_empty.as_deref(), &column, false)?;
        column.on_error = behavior(binder, raw.on_error.as_deref(), &column, true)?;
        check_exists_behavior(&column)?;
        own.push(planning.columns.len());
        planning.columns.push(column);
    }
    Ok((own, nested))
}

/// valid_default reports whether a DEFAULT behavior's expression is a constant, a function call, or an operator, under
/// any casts, as Postgres' ValidJsonBehaviorDefaultExpr allows.
fn valid_default(expr: &Expr) -> bool {
    match expr {
        Expr::Cast(inner, ..) => valid_default(inner),
        Expr::Const(_)
        | Expr::Func(..)
        | Expr::Routine(..)
        | Expr::Operator(..)
        | Expr::Arith(..)
        | Expr::Neg(..)
        | Expr::Compare(..)
        | Expr::Concat(..)
        | Expr::DateTime(..)
        | Expr::ArrayOp(..)
        | Expr::AnyArray(..) => true,
        _ => false,
    }
}

/// check_exists_behavior fails as Postgres does when an EXISTS column's TRUE or FALSE ON ERROR behavior cannot be its
/// type.
fn check_exists_behavior(column: &TableColumn) -> Result<()> {
    if column.kind == Kind::Exists
        && let Behavior::True | Behavior::False = column.on_error
    {
        let value = column.on_error == Behavior::True;
        if let Err(err) = exists_value(value, column.ty) {
            return Err(PgError {
                detail: Some(err.message),
                ..PgError::new(
                    code::DATATYPE_MISMATCH,
                    format!(
                        "could not coerce ON ERROR expression ({}) to the RETURNING type",
                        if value { "TRUE" } else { "FALSE" }
                    ),
                )
            });
        }
    }
    Ok(())
}

/// bind_context binds the context item of a JSON_TABLE or SQL/JSON query function, which must read as JSON.
fn bind_context(binder: &mut Binder<'_, '_>, value: &pg_query::protobuf::JsonValueExpr) -> Result<Input> {
    let (mut context, location) = bind_input(binder, value)?;
    let readable = matches!(context.ty, oid::JSON | oid::JSONB | oid::UNKNOWN)
        || crate::expr::is_string(context.ty)
        || context.ty == oid::BYTEA && context.formatted;
    if !readable {
        return Err(located(
            code::CANNOT_COERCE,
            format!("cannot cast type {} to jsonb", crate::cast::type_display(context.ty)),
            location,
        ));
    }
    context.formatted = true;
    Ok(context)
}

/// bind_passing binds the PASSING values of a JSON_TABLE or SQL/JSON query function.
fn bind_passing(binder: &mut Binder<'_, '_>, nodes: &[pg_query::Node]) -> Result<Vec<(String, Input)>> {
    let mut passing = Vec::new();
    for node in nodes {
        let Some(NodeEnum::JsonArgument(argument)) = node.node.as_ref() else { continue };
        let value = argument.val.as_deref().ok_or_else(|| PgError::internal("a PASSING value without a value"))?;
        let (input, _) = bind_input(binder, value)?;
        let plain = matches!(
            input.ty,
            oid::BOOL
                | oid::INT2
                | oid::INT4
                | oid::INT8
                | oid::FLOAT4
                | oid::FLOAT8
                | oid::NUMERIC
                | oid::TEXT
                | oid::VARCHAR
                | oid::UNKNOWN
                | oid::DATE
                | oid::TIME
                | oid::TIMETZ
                | oid::TIMESTAMP
                | oid::TIMESTAMPTZ
                | oid::JSON
                | oid::JSONB
        );
        if !input.formatted && !plain && crate::expr::is_string(input.ty) {
            return Err(PgError::new(
                code::INVALID_PARAMETER_VALUE,
                format!("could not convert value of type {} to jsonpath", crate::cast::type_display(input.ty)),
            ));
        }
        passing.push((argument.name.clone(), input));
    }
    Ok(passing)
}

/// plan plans a JSON_TABLE, binding its expressions with the binder.
pub fn plan(binder: &mut Binder<'_, '_>, table: &pg_query::protobuf::JsonTable) -> Result<JsonTable> {
    let context = table.context_item.as_deref().ok_or_else(|| PgError::internal("JSON_TABLE without a document"))?;
    let context = bind_context(binder, context)?;
    let root_path = path_of(table.pathspec.as_deref())?;
    let passing = bind_passing(binder, &table.passing)?;
    let mut planning = Planning { names: HashSet::new(), columns: Vec::new() };
    if let Some(spec) = table.pathspec.as_deref()
        && !spec.name.is_empty()
    {
        planning.claim(&spec.name, spec.name_location)?;
    }
    let (own, nested) = columns(binder, &mut planning, &table.columns)?;
    let error_on_error = match table.on_error.as_deref() {
        None => false,
        Some(clause) => match behavior_type(clause) {
            JsonBehaviorType::JsonBehaviorError => true,
            JsonBehaviorType::JsonBehaviorEmpty | JsonBehaviorType::JsonBehaviorEmptyArray => false,
            _ => {
                return Err(PgError {
                    detail: Some("Only EMPTY [ ARRAY ] or ERROR is allowed in the top-level ON ERROR clause.".into()),
                    ..located(code::SYNTAX_ERROR, "invalid ON ERROR behavior", clause.location)
                });
            }
        },
    };
    Ok(JsonTable {
        context,
        passing,
        root: TablePath { path: root_path, columns: own, nested },
        columns: planning.columns,
        error_on_error,
    })
}

/// input_item evaluates the context item or a PASSING value as a SQL/JSON item, or returns None for NULL.
fn input_item(ctx: &mut Ctx<'_>, input: &Input) -> Result<Option<Item>> {
    value_item(input.expr.eval(ctx, &[])?, input.ty, input.formatted)
}

/// value_item converts the value of the context item or a PASSING value, of a type and read as JSON or not, to a
/// SQL/JSON item, or returns None for NULL.
fn value_item(value: Value, ty: u32, formatted: bool) -> Result<Option<Item>> {
    let parse = |text: &str| crate::json::parse(text, true).map(|json| Some(Item::Json(json)));
    match value {
        Value::Null => Ok(None),
        Value::Jsonb(json) => Ok(Some(Item::Json(*json))),
        Value::Json(text) => parse(&text),
        Value::Bytea(bytes) if formatted => parse(&String::from_utf8_lossy(&bytes)),
        value if formatted || matches!(ty, oid::JSON | oid::JSONB) => parse(&value.output().unwrap_or_default()),
        Value::Bool(b) => Ok(Some(Item::Json(Json::Bool(b)))),
        value @ (Value::Int2(_)
        | Value::Int4(_)
        | Value::Int8(_)
        | Value::Float4(_)
        | Value::Float8(_)
        | Value::Numeric(_)) => {
            let text = value.output().unwrap_or_default();
            Ok(Some(Item::Json(Json::Number(crate::numeric::Numeric::parse(&text)?))))
        }
        value @ (Value::Date(_) | Value::Time(_) | Value::TimeTz(..) | Value::Timestamp(_) | Value::TimestampTz(_)) => {
            Ok(Some(Item::DateTime(value, None)))
        }
        Value::Text(text) => Ok(Some(Item::Json(Json::String(text)))),
        other => Ok(Some(Item::Json(crate::functions::json::datum_to_json(&other)?))),
    }
}

/// rows returns the rows of a JSON_TABLE, as Postgres' JSON_TABLE execution computes them.
pub fn rows(ctx: &mut Ctx<'_>, table: &JsonTable) -> Result<Vec<Vec<Value>>> {
    let Some(document) = input_item(ctx, &table.context)? else { return Ok(Vec::new()) };
    let mut vars = HashMap::new();
    for (name, input) in &table.passing {
        if vars.contains_key(name) {
            continue;
        }
        let value = input_item(ctx, input)?.unwrap_or(Item::Json(Json::Null));
        vars.insert(name.clone(), value);
    }
    path_rows(ctx, table, &table.root, &document, &vars)
}

/// plan_function plans a JSON_EXISTS, JSON_QUERY, or JSON_VALUE, binding its expressions with the binder, as Postgres'
/// transformJsonFuncExpr does.
pub fn plan_function(binder: &mut Binder<'_, '_>, func: &JsonFuncExpr) -> Result<(Expr, ColumnType)> {
    let (kind, name) = match JsonExprOp::try_from(func.op) {
        Ok(JsonExprOp::JsonExistsOp) => (Kind::Exists, "JSON_EXISTS"),
        Ok(JsonExprOp::JsonQueryOp) => (Kind::Query, "JSON_QUERY"),
        _ => (Kind::Value, "JSON_VALUE"),
    };
    let output = func.output.as_ref();
    if kind != Kind::Query
        && let Some(format) = output.and_then(|o| o.returning.as_ref()).and_then(|r| r.format.as_ref())
        && (!matches!(
            pg_query::protobuf::JsonFormatType::try_from(format.format_type),
            Ok(pg_query::protobuf::JsonFormatType::JsFormatDefault | pg_query::protobuf::JsonFormatType::Undefined)
        ) || !matches!(
            JsonEncoding::try_from(format.encoding),
            Ok(JsonEncoding::JsEncDefault | JsonEncoding::Undefined)
        ))
    {
        return Err(located(
            code::SYNTAX_ERROR,
            format!("cannot specify FORMAT JSON in RETURNING clause of {name}()"),
            format.location,
        ));
    }
    let wrapper = match JsonWrapper::try_from(func.wrapper) {
        Ok(JsonWrapper::JswConditional) => Wrapper::Conditional,
        Ok(JsonWrapper::JswUnconditional) => Wrapper::Unconditional,
        _ => Wrapper::None,
    };
    let omit_quotes = JsonQuotes::try_from(func.quotes) == Ok(JsonQuotes::JsQuotesOmit);
    if kind == Kind::Query && omit_quotes && wrapper != Wrapper::None {
        return Err(located(
            code::SYNTAX_ERROR,
            "SQL/JSON QUOTES behavior must not be specified when WITH WRAPPER is used",
            func.location,
        ));
    }
    let context =
        func.context_item.as_deref().ok_or_else(|| PgError::internal("a SQL/JSON function without a document"))?;
    let context = bind_context(binder, context)?;
    let spec = func.pathspec.as_deref().ok_or_else(|| PgError::internal("a SQL/JSON function without a path"))?;
    let (path, path_type) = binder.bind(spec)?;
    let location = arg_location(spec);
    let path = crate::expr::coerce((path, path_type), ColumnType { oid: oid::JSONPATH, modifier: -1 }, true, location)
        .map_err(|_| {
            located(
                code::DATATYPE_MISMATCH,
                format!(
                    "JSON path expression must be of type jsonpath, not of type {}",
                    crate::cast::type_display(path_type.oid)
                ),
                location,
            )
        })?
        .0;
    let passing = bind_passing(binder, &func.passing)?;
    let ty = match output.and_then(|o| o.type_name.as_ref()) {
        Some(type_name) => {
            binder.ctx.prepare_type(type_name)?;
            let ty = crate::expr::resolve_type_name(type_name)?;
            if crate::catalog::builtin_type(ty.oid).is_some_and(|t| t.definition.typ_type == b"p") {
                return Err(PgError::new(
                    code::FEATURE_NOT_SUPPORTED,
                    "returning pseudo-types is not supported in SQL/JSON functions",
                ));
            }
            ty
        }
        None => ColumnType {
            oid: match kind {
                Kind::Exists => oid::BOOL,
                Kind::Query => oid::JSONB,
                _ => oid::TEXT,
            },
            modifier: -1,
        },
    };
    let mut column = TableColumn {
        name: String::new(),
        ty,
        kind,
        path: None,
        wrapper,
        omit_quotes: omit_quotes || kind == Kind::Value,
        on_empty: Behavior::Null,
        on_error: Behavior::Null,
    };
    column.on_empty = behavior(binder, func.on_empty.as_deref(), &column, false)?;
    column.on_error = behavior(binder, func.on_error.as_deref(), &column, true)?;
    check_exists_behavior(&column)?;
    let mut args = vec![context.expr, path];
    let mut names = Vec::with_capacity(passing.len());
    for (name, input) in passing {
        names.push((name, input.ty, input.formatted));
        args.push(input.expr);
    }
    let func = JsonFunc { context: (context.ty, context.formatted), passing: names, column };
    Ok((Expr::JsonFunc(Box::new(func), args), ty))
}

/// function_value returns what a SQL/JSON query function returns for the values of its arguments.
pub fn function_value(ctx: &mut Ctx<'_>, func: &JsonFunc, args: Vec<Value>) -> Result<Value> {
    let mut args = args.into_iter();
    let (Some(context), Some(path)) = (args.next(), args.next()) else {
        return Err(PgError::internal("a SQL/JSON function without its document and path"));
    };
    let Some(document) = value_item(context, func.context.0, func.context.1)? else { return Ok(Value::Null) };
    if path == Value::Null {
        return Ok(Value::Null);
    }
    let path = jsonpath::parse(&path.output().unwrap_or_default())?;
    let mut vars = HashMap::new();
    for ((name, ty, formatted), value) in func.passing.iter().zip(args) {
        if !vars.contains_key(name) {
            vars.insert(name.clone(), value_item(value, *ty, *formatted)?.unwrap_or(Item::Json(Json::Null)));
        }
    }
    path_value(ctx, &func.column, &path, &document, &vars)
}

/// path_rows returns the rows of a row path for an item: one per item its path matches, each joined with the rows
/// of its nested paths, which leave the other paths' columns NULL.
fn path_rows(
    ctx: &mut Ctx<'_>,
    table: &JsonTable,
    path: &TablePath,
    item: &Item,
    vars: &HashMap<String, Item>,
) -> Result<Vec<Vec<Value>>> {
    let options = Options { vars, throw: table.error_on_error, use_tz: false };
    let items = jsonpath::exec::query(&path.path, item, &options)?.unwrap_or_default();
    let mut out = Vec::new();
    for (ordinal, item) in items.iter().enumerate() {
        let mut row = vec![Value::Null; table.columns.len()];
        for &index in &path.columns {
            row[index] = column_value(ctx, &table.columns[index], item, ordinal, vars)?;
        }
        let mut nested_rows = Vec::new();
        for nested in &path.nested {
            for mut nested_row in path_rows(ctx, table, nested, item, vars)? {
                for &index in &path.columns {
                    nested_row[index] = row[index].clone();
                }
                nested_rows.push(nested_row);
            }
        }
        if nested_rows.is_empty() {
            out.push(row);
        } else {
            out.extend(nested_rows);
        }
    }
    Ok(out)
}

/// column_value returns a column's value for a row's item, applying its ON EMPTY and ON ERROR behaviors.
fn column_value(
    ctx: &mut Ctx<'_>,
    column: &TableColumn,
    item: &Item,
    ordinal: usize,
    vars: &HashMap<String, Item>,
) -> Result<Value> {
    let Some(path) = &column.path else { return Ok(Value::Int4(ordinal as i32 + 1)) };
    path_value(ctx, column, path, item, vars)
}

/// path_value returns what a column's path makes of an item, applying the column's ON EMPTY and ON ERROR behaviors.
fn path_value(
    ctx: &mut Ctx<'_>,
    column: &TableColumn,
    path: &JsonPath,
    item: &Item,
    vars: &HashMap<String, Item>,
) -> Result<Value> {
    let options = Options { vars, throw: column.on_error == Behavior::Error, use_tz: false };
    if column.kind == Kind::Exists {
        return match jsonpath::exec::exists(path, item, &options)? {
            Some(exists) => exists_value(exists, column.ty),
            None => behavior_value(ctx, column, true, None),
        };
    }
    let Some(items) = jsonpath::exec::query(path, item, &options)? else {
        return behavior_value(ctx, column, true, None);
    };
    if items.is_empty() {
        let message = match column.name.is_empty() {
            true => "no SQL/JSON item found for specified path".to_string(),
            false => format!("no SQL/JSON item found for specified path of column \"{}\"", column.name),
        };
        let err = PgError::new(code::NO_SQL_JSON_ITEM, message);
        return behavior_value(ctx, column, false, Some(err));
    }
    match items_value(column, items) {
        Ok(value) => Ok(value),
        Err(err) => behavior_value(ctx, column, true, Some(err)),
    }
}

/// behavior_value returns the value that a column's ON EMPTY behavior, or its ON ERROR behavior when `on_error`,
/// produces, failing with the error for ERROR, and as Postgres does when the value cannot be the column's type.
fn behavior_value(ctx: &mut Ctx<'_>, column: &TableColumn, on_error: bool, err: Option<PgError>) -> Result<Value> {
    let behavior = if on_error { &column.on_error } else { &column.on_empty };
    let (value, name) = match behavior {
        Behavior::Error => return Err(err.unwrap_or_else(|| PgError::internal("a JSON_TABLE error without a message"))),
        Behavior::Null | Behavior::Unknown => return Ok(Value::Null),
        Behavior::True => (exists_value(true, column.ty), "TRUE"),
        Behavior::False => (exists_value(false, column.ty), "FALSE"),
        Behavior::EmptyArray => (json_value(column, &Json::Array(Vec::new())), "EMPTY ARRAY"),
        Behavior::EmptyObject => (json_value(column, &Json::Object(Vec::new())), "EMPTY OBJECT"),
        Behavior::Default(expr) => (expr.eval(ctx, &[]), "DEFAULT"),
    };
    value.map_err(|err| PgError {
        detail: Some(err.message),
        ..PgError::new(
            code::DATATYPE_MISMATCH,
            format!(
                "could not coerce {} expression ({name}) to the RETURNING type",
                if on_error { "ON ERROR" } else { "ON EMPTY" }
            ),
        )
    })
}

/// scalar_required returns Postgres' error for a scalar column whose path found other than one scalar item.
fn scalar_required(code: &'static str, column: &TableColumn) -> PgError {
    match column.name.is_empty() {
        true => PgError::new(code, "JSON path expression in JSON_VALUE must return single scalar item"),
        false => PgError::new(
            code,
            format!("JSON path expression for column \"{}\" must return single scalar item", column.name),
        ),
    }
}

/// typed reads text as a column's type, with its modifier.
fn typed(text: &str, ty: ColumnType) -> Result<Value> {
    let value = crate::cast::input(text, ty.oid)?;
    crate::cast::cast_value(value, ty, false)
}

/// items_value returns a column's value for the items its path found.
fn items_value(column: &TableColumn, items: Vec<Item>) -> Result<Value> {
    if column.kind == Kind::Value {
        let [item] = items.as_slice() else { return Err(scalar_required(code::MORE_THAN_ONE_SQL_JSON_ITEM, column)) };
        let text = match item {
            Item::Json(Json::Null) => return Ok(Value::Null),
            Item::Json(Json::Array(_) | Json::Object(_)) => {
                return Err(scalar_required(code::SQL_JSON_SCALAR_REQUIRED, column));
            }
            Item::Json(Json::String(s)) => s.clone(),
            Item::Json(Json::Bool(b)) => (if *b { "t" } else { "f" }).to_string(),
            Item::Json(Json::Number(n)) => n.to_string(),
            Item::DateTime(value, _) => value.output().unwrap_or_default(),
        };
        if matches!(crate::usertypes::base_type(column.ty).oid, oid::JSON | oid::JSONB) {
            return typed(&item.to_json().to_text(), column.ty);
        }
        return typed(&text, column.ty);
    }
    let wrap = match column.wrapper {
        Wrapper::Unconditional => true,
        Wrapper::Conditional => items.len() > 1,
        Wrapper::None => false,
    };
    let result = if wrap {
        Json::Array(items.iter().map(Item::to_json).collect())
    } else {
        let [item] = items.as_slice() else {
            let message = match column.name.is_empty() {
                true => {
                    "JSON path expression in JSON_QUERY must return single item when no wrapper is requested".into()
                }
                false => format!(
                    "JSON path expression for column \"{}\" must return single item when no wrapper is requested",
                    column.name
                ),
            };
            return Err(PgError {
                hint: Some("Use the WITH WRAPPER clause to wrap SQL/JSON items into an array.".into()),
                ..PgError::new(code::MORE_THAN_ONE_SQL_JSON_ITEM, message)
            });
        };
        item.to_json()
    };
    json_value(column, &result)
}

/// json_value converts a JSON item to a formatted column's type.
fn json_value(column: &TableColumn, json: &Json) -> Result<Value> {
    if let (Json::String(s), true) = (json, column.omit_quotes) {
        return typed(s, column.ty);
    }
    let base = crate::usertypes::base_type(column.ty).oid;
    if crate::array::is_array_type(base) || crate::expr::is_composite(base) {
        if *json == Json::Null {
            return Ok(Value::Null);
        }
        let literal = literal(base, json, &mut Vec::new())?;
        return typed(&literal, column.ty);
    }
    typed(&json.to_text(), column.ty)
}

/// element_text returns the text an array element or composite field reads from: a string's value, or the JSON text
/// of any other value.
fn element_text(json: &Json, ty: u32, path: &mut Vec<usize>) -> Result<String> {
    let base = crate::usertypes::base_type(ColumnType { oid: ty, modifier: -1 }).oid;
    match json {
        Json::String(s) => Ok(s.clone()),
        _ if crate::array::is_array_type(base) || crate::expr::is_composite(base) => literal(base, json, path),
        other => Ok(other.to_text()),
    }
}

/// quote writes text as a quoted array element or composite field.
fn quote(out: &mut String, text: &str) {
    out.push('"');
    for c in text.chars() {
        if c == '"' || c == '\\' {
            out.push('\\');
        }
        out.push(c);
    }
    out.push('"');
}

/// array_depth returns the dimensions of a JSON array, as its first elements nest.
fn array_depth(json: &Json) -> usize {
    match json {
        Json::Array(elements) => 1 + elements.first().map_or(0, array_depth),
        _ => 0,
    }
}

/// literal converts a JSON value to the input text of an array or composite type, as Postgres' populate_record
/// functions read one, failing with their errors.
fn literal(ty: u32, json: &Json, path: &mut Vec<usize>) -> Result<String> {
    if crate::array::is_array_type(ty) {
        let element = crate::expr::element_type(ty);
        let depth = array_depth(json);
        if depth == 0 {
            return Err(expected_array(path));
        }
        let mut out = String::new();
        array_literal(json, element, depth, path, &mut out)?;
        return Ok(out);
    }
    let Json::Object(pairs) = json else {
        let kind = if matches!(json, Json::Array(_)) { "an array" } else { "a scalar" };
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, format!("cannot call populate_composite on {kind}")));
    };
    let Some(user_type) = crate::usertypes::get(ty) else {
        return Err(PgError::internal("a composite type without a definition"));
    };
    let crate::usertypes::Kind::Composite(attributes) = &user_type.kind else {
        return Err(PgError::internal("a composite type without attributes"));
    };
    let mut out = String::from("(");
    for (i, (name, attribute)) in attributes.iter().enumerate() {
        if i > 0 {
            out.push(',');
        }
        if let Some((_, field)) = pairs.iter().find(|(k, _)| k == name)
            && *field != Json::Null
        {
            let text = element_text(field, attribute.oid, path)?;
            quote(&mut out, &text);
        }
    }
    out.push(')');
    Ok(out)
}

/// expected_array returns Postgres' error for a JSON value that is not the array an array type needs.
fn expected_array(path: &[usize]) -> PgError {
    let hint = (!path.is_empty()).then(|| {
        let indexes: String = path.iter().map(|i| format!("[{i}]")).collect();
        format!("See the array element {indexes}.")
    });
    PgError { hint, ..PgError::new(code::INVALID_TEXT_REPRESENTATION, "expected JSON array") }
}

/// array_literal writes a JSON array of the dimensions as an array literal.
fn array_literal(json: &Json, element: u32, depth: usize, path: &mut Vec<usize>, out: &mut String) -> Result<()> {
    let Json::Array(elements) = json else { return Err(expected_array(path)) };
    out.push('{');
    for (i, value) in elements.iter().enumerate() {
        if i > 0 {
            out.push(',');
        }
        path.push(i);
        if depth > 1 {
            array_literal(value, element, depth - 1, path, out)?;
        } else if *value == Json::Null {
            out.push_str("NULL");
        } else {
            let text = element_text(value, element, path)?;
            quote(out, &text);
        }
        path.pop();
    }
    out.push('}');
    Ok(())
}
