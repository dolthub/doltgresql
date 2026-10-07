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

//! PL/pgSQL: compiling a body into Go's interpreter operations, which Go stores in the function's root object, and
//! running those operations, with each embedded statement taking the variables it names as parameters.

mod compile;

use std::collections::HashMap;

use objects::Operation;
use pg_query::NodeEnum;

use crate::catalog::{ColumnType, builtin_type_named};
use crate::error::{PgError, Result, code};
use crate::expr::typ;
use crate::query::Ctx;
use crate::routines::{Routine, TRIGGER, VOID};
use crate::types::Value;
use crate::{Outcome, oid};

/// OpCode is what an operation does, numbered as Go stores it.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
#[repr(u16)]
pub enum OpCode {
    Alias = 0,
    Assign = 1,
    Declare = 3,
    Execute = 6,
    Goto = 8,
    If = 9,
    Perform = 11,
    Raise = 12,
    Return = 13,
    ScopeBegin = 14,
    ScopeEnd = 15,
    ReturnQuery = 18,
    ForQueryInit = 19,
    ForQueryNext = 20,
    DeclareRecord = 21,
    ExecuteInto = 22,
}

impl OpCode {
    /// from_stored returns the operation code Go stores as the number, or None for one that does nothing.
    fn from_stored(code: u16) -> Option<OpCode> {
        Some(match code {
            0 => OpCode::Alias,
            1 => OpCode::Assign,
            3 => OpCode::Declare,
            6 => OpCode::Execute,
            8 => OpCode::Goto,
            9 => OpCode::If,
            11 => OpCode::Perform,
            12 => OpCode::Raise,
            13 => OpCode::Return,
            14 => OpCode::ScopeBegin,
            15 => OpCode::ScopeEnd,
            18 => OpCode::ReturnQuery,
            19 => OpCode::ForQueryInit,
            20 => OpCode::ForQueryNext,
            21 => OpCode::DeclareRecord,
            22 => OpCode::ExecuteInto,
            _ => return None,
        })
    }
}

/// The option keys of operations, as Go names them.
const OPTION_SETS_FOUND: &str = "sets_found";
const OPTION_DYNAMIC_EXPRESSION: &str = "dynamic_expression";
const OPTION_DYNAMIC_BINDING: &str = "dynamic_binding_";
const OPTION_DYNAMIC_BINDING_COUNT: &str = "dynamic_binding_count";
const OPTION_DYNAMIC_USING_COUNT: &str = "dynamic_using_count";
const OPTION_DYNAMIC_USING_EXPRESSION: &str = "dynamic_using_expression_";
const OPTION_DYNAMIC_USING_BINDING_COUNT: &str = "dynamic_using_binding_count_";
const OPTION_DYNAMIC_USING_BINDING: &str = "dynamic_using_binding_";
const OPTION_RETYPE_TARGET: &str = "retype_target";
const OPTION_LOOP_CONDITION: &str = "loop_condition";
/// OPTION_STRICT marks an INTO STRICT, which Go ignores.
const OPTION_STRICT: &str = "strict";
/// CONTINUE_TARGET is the option compilation uses for where a loop's CONTINUE jumps, which never reaches storage.
const CONTINUE_TARGET: &str = "continue_target";

/// RAISE's USING options, by the numbers the parser gives them.
const RAISE_ERRCODE: &str = "0";
const RAISE_MESSAGE: &str = "1";
const RAISE_DETAIL: &str = "2";
const RAISE_HINT: &str = "3";

/// FOUND is the name of the variable that reports whether the last statement found a row.
const FOUND: &str = "found";

/// TRIGGER_VARIABLES are the special variables of a trigger function with their types.
const TRIGGER_VARIABLES: [(&str, u32); 10] = [
    ("tg_name", oid::NAME),
    ("tg_when", oid::TEXT),
    ("tg_level", oid::TEXT),
    ("tg_op", oid::TEXT),
    ("tg_relid", oid::OID),
    ("tg_relname", oid::NAME),
    ("tg_table_name", oid::NAME),
    ("tg_table_schema", oid::NAME),
    ("tg_nargs", oid::INT4),
    ("tg_argv", oid::TEXT_ARRAY),
];

/// SQL_TYPE_NAMES are the SQL spellings of built-in types that the parser can give a declaration's type as, with the
/// names the types have in pg_catalog.
const SQL_TYPE_NAMES: [(&str, &str); 9] = [
    ("integer", "int4"),
    ("boolean", "bool"),
    ("smallint", "int2"),
    ("bigint", "int8"),
    ("real", "float4"),
    ("double precision", "float8"),
    ("decimal", "numeric"),
    ("character varying", "varchar"),
    ("character", "bpchar"),
];

/// normalize_identifier folds an identifier as Postgres does: unquoted to lowercase, quoted without its quotes.
fn normalize_identifier(identifier: &str) -> String {
    match identifier.strip_prefix('"').and_then(|i| i.strip_suffix('"')) {
        Some(quoted) if identifier.len() >= 2 => quoted.replace("\"\"", "\""),
        _ => identifier.to_lowercase(),
    }
}

/// normalize_path folds a reference written as `name` or `name.field`, or returns None for any other text.
fn normalize_path(text: &str) -> Option<String> {
    let text = text.trim();
    let is_identifier = |s: &str| {
        if let Some(inner) = s.strip_prefix('"') {
            return inner.strip_suffix('"').is_some_and(|i| !i.contains('"'));
        }
        let mut chars = s.chars();
        chars.next().is_some_and(|c| c == '_' || c.is_ascii_alphabetic())
            && chars.all(|c| c == '_' || c == '$' || c.is_ascii_alphanumeric())
    };
    match text.split_once('.') {
        Some((base, field)) if is_identifier(base) && is_identifier(field) => {
            Some(format!("{}.{field}", normalize_identifier(base)))
        }
        None if is_identifier(text) => Some(normalize_identifier(text)),
        _ => None,
    }
}

/// quote_identifier quotes a name so that folding gives it back unchanged.
fn quote_identifier(name: &str) -> String {
    format!("\"{}\"", name.replace('"', "\"\""))
}

/// declared_type_name returns the name a declaration's type is stored under: `pg_catalog.name` for a built-in type,
/// which Go can resolve, and the text as written for anything else, which is resolved when the declaration runs.
fn declared_type_name(text: &str) -> String {
    let lowered = text.to_lowercase();
    if lowered.contains('%') {
        return lowered;
    }
    match crate::routines::parse_type(text) {
        Ok(ty) => crate::catalog::builtin_type(ty.oid).map_or(lowered, |t| format!("pg_catalog.{}", t.name)),
        Err(_) => lowered,
    }
}

/// compile compiles the body of a CREATE FUNCTION or CREATE PROCEDURE statement into Go's operations, checking the
/// declared types as Postgres does.
pub fn compile(ctx: &mut Ctx<'_>, text: &str, body: &str) -> Result<Vec<Operation>> {
    let locate = |err: PgError| {
        let keyword = if err.message.starts_with("EXIT") { "exit" } else { "continue" };
        match err.message.contains("cannot be used outside a loop") {
            true => PgError { position: statement_position(text, body, keyword, false), ..err },
            false => err,
        }
    };
    let json = pg_query::parse_plpgsql(text).map_err(|err| match err {
        pg_query::Error::Parse(message) => locate(PgError::new(code::SYNTAX_ERROR, message)),
        other => PgError::internal(other),
    })?;
    let functions = json.as_array().map_or(&[][..], Vec::as_slice);
    let [function] = functions else { return Err(PgError::internal("CREATE FUNCTION parsed multiple blocks")) };
    let function = function.get("PLpgSQL_function").ok_or_else(|| PgError::internal("a PL/pgSQL function"))?;
    check_declarations(ctx, function, body, text.find(body).unwrap_or(0))?;
    compile::compile_json(function, body).map_err(locate)
}

/// statement_position returns the position in a statement of the first PL/pgSQL statement in its body that starts with
/// the keyword, or of the token after the keyword when asked, which compilation errors point at.
pub fn statement_position(text: &str, body: &str, keyword: &str, after: bool) -> Option<u32> {
    let offset = text.find(body)?;
    let tokens = pg_query::scan(body).ok()?.tokens;
    let piece = |t: &pg_query::protobuf::ScanToken| &body[t.start as usize..t.end as usize];
    let i = tokens.iter().position(|t| piece(t).eq_ignore_ascii_case(keyword))?;
    let token = if after { tokens.get(i + 1)? } else { &tokens[i] };
    Some((offset + token.start as usize + 1) as u32)
}

/// check_declarations checks that the type of each declared variable parses and exists, as Postgres does when it
/// compiles a body, with error positions in the statement whose body starts at the offset.
fn check_declarations(ctx: &mut Ctx<'_>, function: &serde_json::Value, body: &str, offset: usize) -> Result<()> {
    let Ok(scan) = pg_query::scan(body) else { return Ok(()) };
    let tokens = scan.tokens;
    let piece = |t: &pg_query::protobuf::ScanToken| &body[t.start as usize..t.end as usize];
    let mut sections = Vec::new();
    for (i, token) in tokens.iter().enumerate() {
        if piece(token).eq_ignore_ascii_case("declare") {
            let end =
                tokens[i..].iter().position(|t| piece(t).eq_ignore_ascii_case("begin")).map_or(tokens.len(), |n| i + n);
            sections.push(i..end);
        }
    }
    let datums = function.get("datums").and_then(serde_json::Value::as_array).map_or(&[][..], Vec::as_slice);
    for datum in datums {
        let Some(var) = datum.get("PLpgSQL_var") else { continue };
        let name = var.get("refname").and_then(serde_json::Value::as_str).unwrap_or_default();
        if var.get("lineno").and_then(serde_json::Value::as_i64).unwrap_or(0) == 0
            || name.starts_with("__Case__Variable_")
        {
            continue;
        }
        let found = sections.iter().find_map(|section| {
            section.clone().find(|&i| {
                normalize_identifier(piece(&tokens[i])) == name
                    && i > 0
                    && matches!(piece(&tokens[i - 1]).to_lowercase().as_str(), "declare" | ";")
                    && !tokens.get(i + 1).is_some_and(|t| piece(t).eq_ignore_ascii_case("alias"))
            })
        });
        let Some(found) = found else { continue };
        let mut start = found + 1;
        if tokens.get(start).is_some_and(|t| piece(t).eq_ignore_ascii_case("constant")) {
            start += 1;
        }
        let end = tokens[start..]
            .iter()
            .position(|t| {
                matches!(piece(t), ":=" | "=" | ";")
                    || ["default", "not", "collate"].iter().any(|w| piece(t).eq_ignore_ascii_case(w))
            })
            .map_or(tokens.len(), |n| start + n);
        if end <= start {
            continue;
        }
        let type_text = &body[tokens[start].start as usize..tokens[end - 1].end as usize];
        if !type_text.contains('%') {
            check_type(ctx, type_text, (offset + tokens[start].start as usize + 1) as u32)?;
        }
    }
    Ok(())
}

/// PREFIX turns a type's text into a statement that parses it.
const PREFIX: &str = "SELECT NULL::";

/// check_type checks that a declaration's type text parses and names a type or table, failing with errors at the
/// position the text starts at.
fn check_type(ctx: &mut Ctx<'_>, text: &str, position: u32) -> Result<()> {
    let result = match pg_query::parse_with_cursor(&format!("{PREFIX}{text}")) {
        Ok(result) => result,
        Err((err, cursor, state)) => {
            let at = (cursor as u32).saturating_sub(1 + PREFIX.len() as u32);
            return Err(PgError { position: Some(position + at), ..crate::parse::syntax_error(err, cursor, &state) });
        }
    };
    let target = result.protobuf.stmts.first().and_then(|raw| match raw.stmt.as_ref()?.node.as_ref()? {
        NodeEnum::SelectStmt(select) => match select.target_list.first()?.node.as_ref()? {
            NodeEnum::ResTarget(target) => Some(target.clone()),
            _ => None,
        },
        _ => None,
    });
    let Some(target) = target else { return Ok(()) };
    if !target.name.is_empty() {
        let word = text.trim_end().rsplit(char::is_whitespace).next().unwrap_or_default();
        let at = text.trim_end().len() - word.len();
        return Err(PgError {
            position: Some(position + at as u32),
            ..PgError::new(code::SYNTAX_ERROR, format!("syntax error at or near \"{word}\""))
        });
    }
    let Some(NodeEnum::TypeCast(cast)) = target.val.as_ref().and_then(|v| v.node.as_ref()) else { return Ok(()) };
    let Some(type_name) = &cast.type_name else { return Ok(()) };
    match crate::expr::resolve_type_name(type_name) {
        Ok(_) => Ok(()),
        Err(err) if err.code == code::UNDEFINED_OBJECT => {
            let names: Vec<&str> = type_name.names.iter().filter_map(crate::expr::node_name).collect();
            if table_columns(ctx, &names.join("."))?.is_some() {
                return Ok(());
            }
            let shown = format!("{}{}", names.join("."), "[]".repeat(type_name.array_bounds.len()));
            Err(PgError {
                position: Some(position),
                ..PgError::new(code::UNDEFINED_OBJECT, format!("type \"{shown}\" does not exist"))
            })
        }
        Err(err) => Err(PgError { position: Some(position), ..err }),
    }
}

/// Columns are the names and types of a record's fields.
type Columns = Vec<(String, ColumnType)>;

/// Variable is a PL/pgSQL variable: a value of a type, or a record with the columns it took on, which are None until
/// something is assigned to a RECORD.
#[derive(Clone, Debug)]
struct Variable {
    ty: ColumnType,
    value: Value,
    columns: Option<Vec<(String, ColumnType)>>,
    /// Whether a declaration fixed the record's columns, as a table's row type does, rather than each assignment.
    fixed: bool,
}

impl Variable {
    /// scalar returns a variable holding a value of the type.
    fn scalar(ty: ColumnType, value: Value) -> Variable {
        Variable { ty, value, columns: None, fixed: false }
    }

    /// record returns a record variable with the columns, whose fields are NULL when no row is given.
    fn record(columns: Vec<(String, ColumnType)>, row: Option<Vec<Value>>) -> Variable {
        let row = row.unwrap_or_else(|| vec![Value::Null; columns.len()]);
        Variable { ty: typ(oid::RECORD), value: Value::Record(row), columns: Some(columns), fixed: false }
    }

    /// is_record reports whether the variable holds a row.
    fn is_record(&self) -> bool {
        self.columns.is_some() || self.ty.oid == oid::RECORD
    }
}

/// Cursor is the rows a FOR loop walks.
struct Cursor {
    columns: Vec<(String, ColumnType)>,
    rows: std::vec::IntoIter<Vec<Value>>,
}

/// Scope is a block's variables, by name, and the loop state of a loop's block.
#[derive(Default)]
struct Scope {
    names: HashMap<String, usize>,
    cursor: Option<Cursor>,
    /// Whether leaving the scope sets FOUND, as leaving a FOR loop does, and whether the loop ran its body.
    reports_found: bool,
    iterated: bool,
}

/// QueryResult is what an embedded statement returned.
struct QueryResult {
    columns: Vec<(String, ColumnType)>,
    rows: Vec<Vec<Value>>,
    /// Whether the statement produced or changed a row, which FOUND reports.
    found: bool,
}

/// Frame is one running call of a PL/pgSQL routine.
pub(crate) struct Frame<'r> {
    routine: &'r Routine,
    ops: &'r [Operation],
    variables: Vec<Variable>,
    scopes: Vec<Scope>,
    /// The rows RETURN QUERY added, with the columns of the last.
    returned: Option<Vec<Vec<Value>>>,
    /// The names of the output parameters, whose values a routine without RETURN values returns.
    outputs: Vec<String>,
}

/// call runs a PL/pgSQL routine on arguments already converted to its input types.
pub fn call(ctx: &mut Ctx<'_>, routine: &Routine, ops: &[Operation], args: Vec<Value>) -> Result<Value> {
    let mut frame = Frame::new(routine, ops);
    for param in routine.params.iter().filter(|p| p.mode.is_output()) {
        frame.declare(&param.name, Variable::scalar(param.ty, Value::Null));
        frame.outputs.push(param.name.clone());
    }
    for (i, (param, value)) in routine.inputs().zip(args).enumerate() {
        frame.declare(&param.name, Variable::scalar(param.ty, value));
        let index = frame.variables.len() - 1;
        frame.alias(&format!("${}", i + 1), index);
    }
    frame.run(ctx)
}

/// call_trigger runs a trigger function with its NEW and OLD rows, in the table's columns, and its special
/// variables, returning the row the function returned, or None when it returned NULL.
#[allow(clippy::too_many_arguments)]
pub fn call_trigger(
    ctx: &mut Ctx<'_>,
    routine: &Routine,
    ops: &[Operation],
    row_type: ColumnType,
    columns: &[(String, ColumnType)],
    new: Option<Vec<Value>>,
    old: Option<Vec<Value>>,
    special: Vec<(&str, Value)>,
) -> Result<Option<Vec<Value>>> {
    let mut frame = Frame::new(routine, ops);
    frame.declare_rows(row_type, columns, new, old);
    for (name, value) in special {
        let ty = TRIGGER_VARIABLES.iter().find(|(n, _)| *n == name).map_or(oid::TEXT, |(_, t)| *t);
        frame.declare(name, Variable::scalar(typ(ty), value));
    }
    Ok(record_fields(frame.run(ctx)?))
}

/// call_condition runs a trigger's compiled WHEN condition with the NEW and OLD rows, returning its value.
pub fn call_condition(
    ctx: &mut Ctx<'_>,
    routine: &Routine,
    ops: &[Operation],
    row_type: ColumnType,
    columns: &[(String, ColumnType)],
    new: Option<Vec<Value>>,
    old: Option<Vec<Value>>,
) -> Result<Value> {
    let mut frame = Frame::new(routine, ops);
    frame.declare_rows(row_type, columns, new, old);
    frame.run(ctx)
}

/// text returns an operation's byte string as text.
fn text(bytes: &[u8]) -> String {
    String::from_utf8_lossy(bytes).into_owned()
}

/// option returns an operation's option as text.
fn option(op: &Operation, key: &str) -> Option<String> {
    op.options.get(key.as_bytes()).map(|v| text(v))
}

/// parse parses the one statement of an embedded query.
fn parse(sql: &str) -> Result<NodeEnum> {
    let result = pg_query::parse(sql).map_err(|err| PgError::new(code::SYNTAX_ERROR, err.to_string()))?;
    let mut statements = result.protobuf.stmts.into_iter().filter_map(|raw| raw.stmt.and_then(|s| s.node));
    match (statements.next(), statements.next()) {
        (Some(statement), None) => Ok(statement),
        (None, _) => Err(PgError::new(code::SYNTAX_ERROR, "query is empty")),
        _ => Err(PgError::new(code::SYNTAX_ERROR, "query contains more than one statement")),
    }
}

/// affected returns the row count at the end of a command tag, such as `INSERT 0 3`.
fn affected(tag: &str) -> u64 {
    tag.rsplit(' ').next().and_then(|n| n.parse().ok()).unwrap_or(0)
}

impl<'r> Frame<'r> {
    /// new returns a frame for a call, with the scope its parameters live in.
    fn new(routine: &'r Routine, ops: &'r [Operation]) -> Frame<'r> {
        Frame {
            routine,
            ops,
            variables: Vec::new(),
            scopes: vec![Scope::default()],
            returned: None,
            outputs: Vec::new(),
        }
    }

    /// declare_rows declares a trigger's NEW and OLD records of the table's row type, which are NULL when the event
    /// has no such row.
    fn declare_rows(
        &mut self,
        row_type: ColumnType,
        columns: &[(String, ColumnType)],
        new: Option<Vec<Value>>,
        old: Option<Vec<Value>>,
    ) {
        for (name, row) in [("new", new), ("old", old)] {
            let record = Variable::record(columns.to_vec(), None);
            let value = row.map_or(Value::Null, Value::Record);
            self.declare(name, Variable { ty: row_type, value, fixed: true, ..record });
        }
    }

    /// declare adds a variable to the innermost scope, hiding any of the name in outer scopes.
    fn declare(&mut self, name: &str, variable: Variable) {
        self.variables.push(variable);
        let index = self.variables.len() - 1;
        self.alias(name, index);
    }

    /// alias makes a name in the innermost scope refer to an existing variable.
    fn alias(&mut self, name: &str, index: usize) {
        if let Some(scope) = self.scopes.last_mut() {
            scope.names.insert(name.to_string(), index);
        }
    }

    /// find returns the variable a name refers to, from the innermost scope out.
    fn find(&self, name: &str) -> Option<usize> {
        self.scopes.iter().rev().find_map(|s| s.names.get(name).copied())
    }

    /// variable returns the variable a name refers to, failing when there is none.
    fn variable(&self, name: &str) -> Result<usize> {
        self.find(name).ok_or_else(|| PgError::internal(format!("variable `{name}` could not be found")))
    }

    /// binding returns the value and type of a name a statement binds: a variable, a field of a record, or a whole
    /// record.
    fn binding(&self, name: &str) -> Result<(Value, ColumnType)> {
        let (base, field) = match name.split_once('.') {
            Some((base, field)) if !field.contains('.') => (base, Some(field.trim_matches('"'))),
            _ => (name, None),
        };
        let variable = &self.variables[self.variable(base)?];
        match (field, &variable.columns) {
            (None, None) if variable.ty.oid != oid::RECORD => Ok((variable.value.clone(), variable.ty)),
            (None | Some("*"), _) => {
                self.require_assigned(base, variable)?;
                let value = match (&variable.value, &variable.columns) {
                    (Value::Record(fields), Some(columns)) => {
                        let type_oid = match variable.ty.oid {
                            oid::RECORD => crate::usertypes::transient("record", columns),
                            row_type => row_type,
                        };
                        Value::Composite(Box::new(crate::types::CompositeValue { type_oid, fields: fields.clone() }))
                    }
                    (value, _) => value.clone(),
                };
                Ok((value, variable.ty))
            }
            (Some(field), Some(columns)) => {
                let index = field_index(columns, field).ok_or_else(|| record_has_no_field(base, field))?;
                let value = match &variable.value {
                    Value::Record(fields) => fields.get(index).cloned().unwrap_or(Value::Null),
                    _ => Value::Null,
                };
                Ok((value, columns[index].1))
            }
            (Some(_), None) if variable.ty.oid == oid::RECORD => Err(not_assigned(base)),
            (Some(field), None) => Err(PgError::new(
                code::UNDEFINED_COLUMN,
                format!("could not identify column \"{field}\" in record data type"),
            )),
        }
    }

    /// require_assigned fails for a RECORD that nothing was assigned to yet.
    fn require_assigned(&self, name: &str, variable: &Variable) -> Result<()> {
        if variable.columns.is_none() && variable.ty.oid == oid::RECORD {
            return Err(not_assigned(name));
        }
        Ok(())
    }

    /// assign stores a value in a variable or a record's field, converting it to the type it holds, where a composite
    /// value gives a RECORD its fields.
    fn assign(&mut self, name: &str, value: Value) -> Result<()> {
        if let Some((base, field)) = name.split_once('.') {
            let index = self.variable(base)?;
            let variable = &mut self.variables[index];
            let Some(columns) = &variable.columns else { return Err(not_assigned(base)) };
            let field = field.trim_matches('"');
            let i = field_index(columns, field).ok_or_else(|| record_has_no_field(base, field))?;
            let value = crate::cast::cast_value(value, columns[i].1, false)?;
            if let Value::Record(fields) = &mut variable.value {
                fields[i] = value;
            } else {
                let mut fields = vec![Value::Null; columns.len()];
                fields[i] = value;
                variable.value = Value::Record(fields);
            }
            return Ok(());
        }
        let index = self.variable(name)?;
        let variable = &mut self.variables[index];
        if let Some(columns) = &variable.columns {
            variable.value = match record_fields(value.clone()) {
                Some(fields) => {
                    let mut converted = Vec::with_capacity(columns.len());
                    for (i, (_, ty)) in columns.iter().enumerate() {
                        converted.push(crate::cast::cast_value(
                            fields.get(i).cloned().unwrap_or(Value::Null),
                            *ty,
                            false,
                        )?);
                    }
                    Value::Record(converted)
                }
                None if value.is_null() => Value::Null,
                None => {
                    return Err(PgError::new(code::DATATYPE_MISMATCH, format!("cannot assign {value:?} to a record")));
                }
            };
            return Ok(());
        }
        if variable.ty.oid == oid::RECORD
            && let Value::Composite(composite) = &value
            && let Some(crate::usertypes::Kind::Composite(attributes)) =
                crate::usertypes::get(composite.type_oid).map(|t| t.kind.clone())
        {
            variable.columns = Some(attributes);
            variable.value = Value::Record(composite.fields.clone());
            return Ok(());
        }
        variable.value =
            if variable.ty.oid == oid::RECORD { value } else { crate::cast::cast_value(value, variable.ty, false)? };
        Ok(())
    }

    /// set_found sets FOUND, unless the routine hides it with a variable of another type.
    fn set_found(&mut self, found: bool) {
        if let Some(index) = self.find(FOUND)
            && self.variables[index].ty.oid == oid::BOOL
        {
            self.variables[index].value = Value::Bool(found);
        }
    }

    /// leave_scope leaves the innermost scope, reporting a FOR loop's FOUND.
    fn leave_scope(&mut self) {
        if let Some(scope) = self.scopes.pop()
            && scope.reports_found
        {
            self.set_found(scope.iterated);
        }
    }

    /// mark_loop marks the innermost scope as a FOR loop's, which reports FOUND when left, noting whether it ran.
    fn mark_loop(&mut self, advanced: bool) {
        if let Some(scope) = self.scopes.last_mut() {
            scope.reports_found = true;
            scope.iterated |= advanced;
        }
    }

    /// expand_whole_row returns the bindings of a statement that is only a `name.*` reference to a record of one
    /// field, which binds that field, and fails for a record of several.
    fn expand_whole_row(&self, sql: &str, bindings: Vec<String>, source: &str) -> Result<Vec<String>> {
        let [binding] = bindings.as_slice() else { return Ok(bindings) };
        let Some(base) = binding.strip_suffix(".*") else { return Ok(bindings) };
        let expression = sql.trim().trim_start_matches("SELECT").trim().trim_end_matches(';').trim();
        if expression != "$1" {
            return Ok(bindings);
        }
        let Some(index) = self.find(base) else { return Ok(bindings) };
        let Some(columns) = &self.variables[index].columns else { return Ok(bindings) };
        match columns.as_slice() {
            [(name, _)] => Ok(vec![format!("{base}.{name}")]),
            _ => Err(PgError::new(code::SYNTAX_ERROR, format!("{source} returned {} columns", columns.len()))),
        }
    }

    /// expand_stars returns a statement whose `name.*` parameters for a record bind each of the record's fields in
    /// turn where they make up a whole item of the SELECT list, as Postgres expands a whole-row reference there into
    /// its columns, and leaves them whole elsewhere.
    fn expand_stars(&self, sql: &str, bindings: &[String]) -> Result<(String, Vec<String>)> {
        let fields = |binding: &String| -> Option<Vec<String>> {
            let base = binding.strip_suffix(".*")?;
            let columns = self.variables[self.find(base)?].columns.as_ref()?;
            Some(columns.iter().map(|(name, _)| format!("{base}.{name}")).collect())
        };
        if !bindings.iter().any(|b| fields(b).is_some()) {
            return Ok((sql.to_string(), bindings.to_vec()));
        }
        use pg_query::protobuf::Token;
        let tokens = pg_query::scan(sql).map_err(|err| PgError::new(code::SYNTAX_ERROR, err.to_string()))?.tokens;
        let mut whole = std::collections::HashSet::new();
        let mut depth = 0;
        for (i, token) in tokens.iter().enumerate() {
            match token.token {
                t if t == Token::Ascii40 as i32 => depth += 1,
                t if t == Token::Ascii41 as i32 => depth -= 1,
                t if t == Token::Param as i32 && depth == 0 => {
                    let before = i.checked_sub(1).map(|j| tokens[j].token);
                    let after = tokens.get(i + 1).map(|t| t.token);
                    let starts = matches!(before, Some(t) if t == Token::Select as i32 || t == Token::Ascii44 as i32);
                    let ends = match after {
                        None => true,
                        Some(t) => {
                            [Token::Ascii44, Token::Ascii59, Token::From, Token::Into].iter().any(|e| *e as i32 == t)
                        }
                    };
                    if starts && ends {
                        whole.insert(i);
                    }
                }
                _ => {}
            }
        }
        let mut expanded: Vec<String> = bindings.to_vec();
        let mut numbers: Vec<Vec<usize>> = vec![Vec::new(); bindings.len()];
        for (k, binding) in bindings.iter().enumerate() {
            if let Some(names) = fields(binding) {
                numbers[k] = (expanded.len() + 1..=expanded.len() + names.len()).collect();
                expanded.extend(names);
            }
        }
        let mut out = String::new();
        let mut last = 0;
        for (_, token) in tokens.iter().enumerate().filter(|(i, _)| whole.contains(i)) {
            let (start, end) = (token.start as usize, token.end as usize);
            let Some(list) = sql[start + 1..end].parse::<usize>().ok().and_then(|n| numbers.get(n.wrapping_sub(1)))
            else {
                continue;
            };
            if list.is_empty() {
                continue;
            }
            out.push_str(&sql[last..start]);
            out.push_str(&list.iter().map(|n| format!("${n}")).collect::<Vec<_>>().join(", "));
            last = end;
        }
        out.push_str(&sql[last..]);
        Ok((out, expanded))
    }

    /// query runs an embedded statement with the variables its parameters bind, expanding whole-row references.
    fn query(&self, ctx: &mut Ctx<'_>, sql: &str, bindings: &[String]) -> Result<QueryResult> {
        let (sql, bindings) = self.expand_stars(sql, bindings)?;
        let (sql, bindings) = (sql.as_str(), bindings.as_slice());
        let statement = parse(sql)?;
        let mut values = Vec::with_capacity(bindings.len());
        let mut types = Vec::with_capacity(bindings.len());
        for name in bindings {
            let (value, ty) = self.binding(name)?;
            values.push(value);
            types.push(ty.oid);
        }
        self.run_statement(ctx, &statement, &mut types, &values)
    }

    /// run_statement runs a parsed statement with parameters of the types and values.
    fn run_statement(
        &self,
        ctx: &mut Ctx<'_>,
        statement: &NodeEnum,
        types: &mut Vec<u32>,
        values: &[Value],
    ) -> Result<QueryResult> {
        match ctx.nested(types, values, None, |ctx| ctx.run(statement))? {
            Outcome::Rows { columns, rows, tag } => {
                let columns: Vec<(String, ColumnType)> = columns
                    .into_iter()
                    .map(|c| (c.name, ColumnType { oid: c.type_oid, modifier: c.type_modifier }))
                    .collect();
                let found = !rows.is_empty() || (!tag.starts_with("SELECT") && affected(&tag) > 0);
                Ok(QueryResult { columns, rows, found })
            }
            Outcome::Command { tag } => {
                Ok(QueryResult { columns: Vec::new(), rows: Vec::new(), found: affected(&tag) > 0 })
            }
            Outcome::Empty => Ok(QueryResult { columns: Vec::new(), rows: Vec::new(), found: false }),
            Outcome::CopyIn { .. } | Outcome::CopyOut { .. } => {
                ctx.session.pending_copy = None;
                Err(PgError::new(code::FEATURE_NOT_SUPPORTED, "cannot COPY to/from client in PL/pgSQL"))
            }
        }
    }

    /// single runs a statement that returns one value, which is NULL without rows.
    fn single(&self, ctx: &mut Ctx<'_>, sql: &str, bindings: &[String]) -> Result<(Value, ColumnType)> {
        let result = self.query(ctx, sql, bindings)?;
        if result.columns.len() != 1 {
            return Err(PgError::new(code::SYNTAX_ERROR, format!("query returned {} columns", result.columns.len())));
        }
        if result.rows.len() > 1 {
            return Err(PgError::new(code::CARDINALITY_VIOLATION, "query returned more than one row"));
        }
        let value = result.rows.into_iter().next().and_then(|r| r.into_iter().next()).unwrap_or(Value::Null);
        Ok((value, result.columns[0].1))
    }

    /// evaluate evaluates expression text from the source, such as a RAISE argument, with the variables it names,
    /// leaving a whole-row reference unparenthesized so that it expands into its columns.
    fn evaluate(&self, ctx: &mut Ctx<'_>, expression: &str) -> Result<(Value, ColumnType)> {
        let sql = match expression.trim().ends_with(".*") {
            true => format!("SELECT {expression}"),
            false => format!("SELECT ({expression})"),
        };
        let (sql, bindings) = self.substitute(&sql)?;
        self.single(ctx, &sql, &bindings)
    }

    /// substitute replaces the variables that source text names with parameters, as compilation does.
    fn substitute(&self, sql: &str) -> Result<(String, Vec<String>)> {
        let tokens = pg_query::scan(sql).map_err(|err| PgError::new(code::SYNTAX_ERROR, err.to_string()))?.tokens;
        let mut out = String::new();
        let mut bindings = Vec::new();
        let mut i = 0;
        while i < tokens.len() {
            let piece = &sql[tokens[i].start as usize..tokens[i].end as usize];
            let after_dot = i > 0 && tokens[i - 1].token == pg_query::protobuf::Token::Ascii46 as i32;
            let name = normalize_identifier(piece);
            let is_call = tokens.get(i + 1).is_some_and(|t| t.token == pg_query::protobuf::Token::Ascii40 as i32);
            if !after_dot && !is_call && self.find(&name).is_some() {
                let mut binding = name;
                while i + 2 < tokens.len() && tokens[i + 1].token == pg_query::protobuf::Token::Ascii46 as i32 {
                    binding = format!("{binding}.{}", &sql[tokens[i + 2].start as usize..tokens[i + 2].end as usize]);
                    i += 2;
                }
                bindings.push(binding);
                out.push_str(&format!("${} ", bindings.len()));
            } else {
                out.push_str(piece);
                out.push(' ');
            }
            i += 1;
        }
        Ok((out, bindings))
    }

    /// resolve_type resolves a declaration's type: a built-in type by name, a table's row type, or a `%TYPE` or
    /// `%ROWTYPE` reference, which gives a record's columns for a row type.
    fn resolve_type(&self, ctx: &mut Ctx<'_>, name: &str) -> Result<(ColumnType, Option<Columns>)> {
        let unquoted = name.replace('"', "");
        if let Some(type_name) = unquoted.strip_prefix("pg_catalog.")
            && let Some(t) =
                builtin_type_named(SQL_TYPE_NAMES.iter().find(|(n, _)| *n == type_name).map_or(type_name, |(_, t)| t))
        {
            return Ok((typ(t.oid), None));
        }
        let not_found = || PgError::new(code::UNDEFINED_OBJECT, format!("type \"{name}\" does not exist"));
        if let Some(reference) = unquoted.strip_suffix("%type") {
            if let Some(index) = self.find(reference) {
                let variable = &self.variables[index];
                return Ok((variable.ty, variable.columns.clone()));
            }
            let (relation, column) = reference.rsplit_once('.').ok_or_else(not_found)?;
            let table = table_columns(ctx, relation)?.ok_or_else(not_found)?;
            let (_, ty) = table.into_iter().find(|(c, _)| c == column).ok_or_else(not_found)?;
            return Ok((ty, None));
        }
        let relation = unquoted.strip_suffix("%rowtype").unwrap_or(&unquoted);
        if let Some(columns) = table_columns(ctx, relation)? {
            return Ok((typ(oid::RECORD), Some(columns)));
        }
        let ty = crate::routines::parse_type(name).map_err(|_| not_found())?;
        Ok((ty, None))
    }

    /// run runs the operations from the first, returning the routine's result.
    fn run(&mut self, ctx: &mut Ctx<'_>) -> Result<Value> {
        let mut pc = 0usize;
        while pc < self.ops.len() {
            let op = &self.ops[pc];
            let mut next = pc + 1;
            let primary = text(&op.primary_data);
            let secondary: Vec<String> = op.secondary_data.iter().map(|s| text(s)).collect();
            let target = text(&op.target);
            match OpCode::from_stored(op.op_code) {
                Some(OpCode::Alias) => {
                    let index = self.variable(&primary)?;
                    self.alias(&target, index);
                }
                Some(OpCode::Assign) => {
                    let bindings = self.expand_whole_row(&primary, secondary, "assignment source")?;
                    if option(op, OPTION_RETYPE_TARGET).as_deref() == Some("true") {
                        let (value, ty) = self.single(ctx, &primary, &bindings)?;
                        match self.find(&target) {
                            Some(index) => self.variables[index] = Variable::scalar(ty, value),
                            None => self.declare(&target, Variable::scalar(ty, value)),
                        }
                    } else {
                        let result = self.query(ctx, &primary, &bindings)?;
                        if result.rows.len() > 1 {
                            return Err(PgError::new(code::CARDINALITY_VIOLATION, "query returned more than one row"));
                        }
                        let row = result.rows.into_iter().next();
                        let value = match row {
                            Some(row) if row.len() == 1 => row.into_iter().next().unwrap_or(Value::Null),
                            Some(row) => Value::Record(row),
                            None => Value::Null,
                        };
                        self.assign(&target, value)?;
                    }
                }
                Some(OpCode::Declare) => {
                    let (ty, columns) = self.resolve_type(ctx, &primary)?;
                    let value = match secondary.get(1) {
                        Some(query) => {
                            let bindings = secondary[2..].to_vec();
                            Some(self.single(ctx, query, &bindings)?.0)
                        }
                        None if secondary.len() == 1 => Some(self.evaluate(ctx, &secondary[0])?.0),
                        None => None,
                    };
                    let variable = match columns {
                        Some(columns) => Variable { fixed: true, ..Variable::record(columns, None) },
                        None => Variable::scalar(ty, Value::Null),
                    };
                    self.declare(&target, variable);
                    if let Some(value) = value {
                        self.assign(&target, value)?;
                    }
                }
                Some(OpCode::DeclareRecord) => {
                    self.declare(&target, Variable::scalar(typ(oid::RECORD), Value::Null));
                    if let Some(query) = secondary.get(1) {
                        let result = self.query(ctx, query, &secondary[2..])?;
                        let row = result.rows.into_iter().next();
                        let index = self.variable(&target)?;
                        self.variables[index] = match row {
                            Some(row) if row.len() == 1 && record_fields(row[0].clone()).is_some() => {
                                let fields = row.into_iter().next().and_then(record_fields).unwrap_or_default();
                                let source = secondary[2..].first().and_then(|b| self.find(b));
                                let columns = match source.and_then(|i| self.variables[i].columns.clone()) {
                                    Some(columns) => columns,
                                    None => fields
                                        .iter()
                                        .enumerate()
                                        .map(|(i, v)| (format!("f{}", i + 1), typ(crate::functions::value_type(v))))
                                        .collect(),
                                };
                                Variable::record(columns, Some(fields))
                            }
                            row => Variable::record(result.columns, row),
                        };
                    }
                }
                Some(OpCode::Execute | OpCode::ExecuteInto) => self.execute(ctx, op, &primary, secondary, &target)?,
                Some(OpCode::Goto) => {
                    let index = op.index.max(0) as usize;
                    if pc <= index {
                        let mut i = pc;
                        while i + 1 < index {
                            match OpCode::from_stored(self.ops[i].op_code) {
                                Some(OpCode::ScopeBegin) => self.scopes.push(Scope::default()),
                                Some(OpCode::ScopeEnd) => self.leave_scope(),
                                _ => {}
                            }
                            i += 1;
                        }
                    } else {
                        let mut i = pc;
                        while i >= index {
                            match OpCode::from_stored(self.ops[i].op_code) {
                                Some(OpCode::ScopeBegin) => self.leave_scope(),
                                Some(OpCode::ScopeEnd) => self.scopes.push(Scope::default()),
                                _ => {}
                            }
                            if i == 0 {
                                break;
                            }
                            i -= 1;
                        }
                    }
                    next = index;
                }
                Some(OpCode::If) => {
                    let bindings = self.expand_whole_row(&primary, secondary, "query")?;
                    let (value, _) = self.single(ctx, &primary, &bindings)?;
                    let met = matches!(crate::cast::cast_value(value, typ(oid::BOOL), false)?, Value::Bool(true));
                    if option(op, OPTION_LOOP_CONDITION).as_deref() == Some("true") {
                        self.mark_loop(met);
                    }
                    if met {
                        next = op.index.max(0) as usize;
                    }
                }
                Some(OpCode::Perform) => {
                    let sql = perform_query(&primary);
                    let result = self.query(ctx, &sql, &secondary)?;
                    self.set_found(result.found);
                }
                Some(OpCode::Raise) => self.raise(ctx, op, &primary, &secondary)?,
                Some(OpCode::Return) => return self.return_value(ctx, &primary, secondary),
                Some(OpCode::ReturnQuery) => {
                    let result = self.query(ctx, &primary, &secondary)?;
                    self.check_structure(&result.columns)?;
                    self.set_found(!result.rows.is_empty());
                    self.returned.get_or_insert_with(Vec::new).extend(result.rows);
                }
                Some(OpCode::ForQueryInit) => {
                    let result = self.query(ctx, &primary, &secondary)?;
                    if let Some(scope) = self.scopes.last_mut() {
                        scope.cursor = Some(Cursor { columns: result.columns, rows: result.rows.into_iter() });
                    }
                    self.mark_loop(false);
                }
                Some(OpCode::ForQueryNext) => {
                    let row = self.scopes.last_mut().and_then(|s| s.cursor.as_mut()).and_then(|c| {
                        let columns = c.columns.clone();
                        c.rows.next().map(|row| (columns, row))
                    });
                    match row {
                        None => next = op.index.max(0) as usize,
                        Some((columns, row)) => {
                            self.mark_loop(true);
                            self.store_row(&target, columns, Some(row))?;
                        }
                    }
                }
                Some(OpCode::ScopeBegin) => self.scopes.push(Scope::default()),
                Some(OpCode::ScopeEnd) => self.leave_scope(),
                None => {}
            }
            pc = next;
        }
        self.finish(None)
    }

    /// check_structure fails as Postgres does when RETURN QUERY's columns do not match the routine's result.
    fn check_structure(&self, columns: &[(String, ColumnType)]) -> Result<()> {
        let expected: Vec<ColumnType> = if self.routine.columns.is_empty() {
            vec![self.routine.ret]
        } else {
            self.routine.columns.iter().map(|(_, ty)| *ty).collect()
        };
        let mismatch = |detail: String| PgError {
            detail: Some(detail),
            ..PgError::new(code::DATATYPE_MISMATCH, "structure of query does not match function result type")
        };
        if self.routine.ret.oid == oid::RECORD && self.routine.columns.is_empty() {
            return Ok(());
        }
        if columns.len() != expected.len() {
            return Err(mismatch(format!(
                "Number of returned columns ({}) does not match expected column count ({}).",
                columns.len(),
                expected.len()
            )));
        }
        let text_like = |t: u32| matches!(t, oid::TEXT | oid::VARCHAR);
        for (i, ((_, actual), wanted)) in columns.iter().zip(&expected).enumerate() {
            if actual.oid != wanted.oid && !(text_like(actual.oid) && text_like(wanted.oid)) {
                return Err(mismatch(format!(
                    "Returned type {} does not match expected type {} in column {}.",
                    crate::cast::type_display(actual.oid),
                    crate::cast::type_display(wanted.oid),
                    i + 1
                )));
            }
        }
        Ok(())
    }

    /// store_row stores a query's row in what an INTO clause or FOR loop names: a record, which takes on the row's
    /// columns, a row-typed variable, or a list of variables, which take the row's values in order.
    fn store_row(&mut self, target: &str, columns: Vec<(String, ColumnType)>, row: Option<Vec<Value>>) -> Result<()> {
        let names: Vec<&str> = target.split(',').collect();
        if let [name] = names.as_slice()
            && let Some(index) = self.find(name)
            && self.variables[index].is_record()
        {
            if self.variables[index].fixed {
                self.assign(name, row.map_or(Value::Null, Value::Record))?;
            } else {
                self.variables[index] = Variable::record(columns, row);
            }
            return Ok(());
        }
        let mut values = row.map(Vec::into_iter);
        for name in names {
            let value = values.as_mut().and_then(Iterator::next).unwrap_or(Value::Null);
            self.assign(name, value)?;
        }
        Ok(())
    }

    /// execute runs an embedded statement, or a dynamic EXECUTE, storing its first row in the INTO target.
    fn execute(
        &mut self,
        ctx: &mut Ctx<'_>,
        op: &Operation,
        primary: &str,
        secondary: Vec<String>,
        target: &str,
    ) -> Result<()> {
        let dynamic = option(op, OPTION_DYNAMIC_EXPRESSION).as_deref() == Some("true");
        let result = if dynamic {
            let count: usize = option(op, OPTION_DYNAMIC_BINDING_COUNT).and_then(|c| c.parse().ok()).unwrap_or(0);
            let bindings: Vec<String> =
                (0..count).filter_map(|i| option(op, &format!("{OPTION_DYNAMIC_BINDING}{i}"))).collect();
            let (query, _) = self.single(ctx, &format!("SELECT ({primary})::text"), &bindings)?;
            let Value::Text(query) = query else {
                return Err(PgError::new(code::NULL_VALUE_NOT_ALLOWED, "query string argument of EXECUTE is null"));
            };
            let using: usize = option(op, OPTION_DYNAMIC_USING_COUNT).and_then(|c| c.parse().ok()).unwrap_or(0);
            let mut values = Vec::with_capacity(using);
            let mut types = Vec::with_capacity(using);
            for i in 0..using {
                let expression = option(op, &format!("{OPTION_DYNAMIC_USING_EXPRESSION}{i}")).unwrap_or_default();
                let count: usize = option(op, &format!("{OPTION_DYNAMIC_USING_BINDING_COUNT}{i}"))
                    .and_then(|c| c.parse().ok())
                    .unwrap_or(0);
                let bindings: Vec<String> =
                    (0..count).filter_map(|j| option(op, &format!("{OPTION_DYNAMIC_USING_BINDING}{i}_{j}"))).collect();
                let (value, ty) = self.single(ctx, &format!("SELECT ({expression})"), &bindings)?;
                values.push(value);
                types.push(ty.oid);
            }
            let statement = parse(&query)?;
            self.run_statement(ctx, &statement, &mut types, &values)?
        } else {
            let statement = parse(primary)?;
            if target.is_empty() && returns_rows(&statement) {
                return Err(PgError {
                    hint: Some("If you want to discard the results of a SELECT, use PERFORM instead.".into()),
                    ..PgError::new(code::SYNTAX_ERROR, "query has no destination for result data")
                });
            }
            let mut values = Vec::with_capacity(secondary.len());
            let mut types = Vec::with_capacity(secondary.len());
            for name in &secondary {
                let (value, ty) = self.binding(name)?;
                values.push(value);
                types.push(ty.oid);
            }
            self.run_statement(ctx, &statement, &mut types, &values)?
        };
        if !target.is_empty() {
            if option(op, OPTION_STRICT).as_deref() == Some("true") {
                match result.rows.len() {
                    0 => return Err(PgError::new(code::NO_DATA_FOUND, "query returned no rows")),
                    1 => {}
                    _ => {
                        return Err(PgError {
                            hint: Some("Make sure the query returns a single row, or use LIMIT 1.".into()),
                            ..PgError::new(code::TOO_MANY_ROWS, "query returned more than one row")
                        });
                    }
                }
            }
            let found = !result.rows.is_empty();
            let row = result.rows.into_iter().next();
            self.store_row(target, result.columns, row)?;
            if option(op, OPTION_SETS_FOUND).as_deref() == Some("true") {
                self.set_found(found);
            }
        } else if option(op, OPTION_SETS_FOUND).as_deref() == Some("true") {
            self.set_found(result.found);
        }
        Ok(())
    }

    /// raise runs RAISE: a notice at its level, or an error for EXCEPTION, with the message's `%` placeholders
    /// replaced by the arguments and the USING options applied.
    fn raise(&mut self, ctx: &mut Ctx<'_>, op: &Operation, level: &str, secondary: &[String]) -> Result<()> {
        let (format, params) = secondary.split_first().map_or(("", &[][..]), |(f, p)| (f.as_str(), p));
        let mut message = String::new();
        let mut params = params.iter();
        let mut chars = format.chars().peekable();
        while let Some(c) = chars.next() {
            if c != '%' {
                message.push(c);
                continue;
            }
            if chars.peek() == Some(&'%') {
                chars.next();
                message.push('%');
                continue;
            }
            let Some(param) = params.next() else {
                return Err(PgError::new(code::SYNTAX_ERROR, "too few parameters specified for RAISE"));
            };
            let (value, _) = match self.find(param.split('.').next().unwrap_or_default()) {
                Some(_) => self.binding(param)?,
                None => self.evaluate(ctx, param)?,
            };
            message.push_str(&value.output().unwrap_or_else(|| "<NULL>".into()));
        }
        if params.next().is_some() {
            return Err(PgError::new(code::SYNTAX_ERROR, "too many parameters specified for RAISE"));
        }
        let mut options = HashMap::new();
        for (key, value) in &op.options {
            let key = text(key);
            let raw = text(value);
            let value = if key == RAISE_ERRCODE
                && let Some(state) = sql_state(&raw)
            {
                state
            } else {
                match self.evaluate(ctx, &raw)?.0.output() {
                    Some(text) => text,
                    None => {
                        return Err(PgError::new(
                            code::NULL_VALUE_NOT_ALLOWED,
                            "RAISE statement option cannot be null",
                        ));
                    }
                }
            };
            options.insert(key, value);
        }
        if let Some(text) = options.remove(RAISE_MESSAGE) {
            message = text;
        }
        let code = match options.get(RAISE_ERRCODE) {
            Some(code) => error_code(code)?,
            None if level == "EXCEPTION" => code::RAISE_EXCEPTION,
            None => "00000",
        };
        let error = PgError {
            detail: options.remove(RAISE_DETAIL),
            hint: options.remove(RAISE_HINT),
            ..PgError::new(code, message)
        };
        if level == "EXCEPTION" {
            return Err(error);
        }
        let severity = match level {
            "WARNING" => "WARNING",
            "INFO" => "INFO",
            "LOG" => "LOG",
            "DEBUG" => "DEBUG",
            _ => "NOTICE",
        };
        if visible(ctx, severity) {
            let code = if options.contains_key(RAISE_ERRCODE) {
                code
            } else if severity == "WARNING" {
                "01000"
            } else {
                "00000"
            };
            ctx.session.notice(PgError { severity, code, ..error });
        }
        Ok(())
    }

    /// return_value runs RETURN: the rows RETURN QUERY gathered, the output parameters, or the expression's value.
    fn return_value(&mut self, ctx: &mut Ctx<'_>, primary: &str, secondary: Vec<String>) -> Result<Value> {
        if self.returned.is_some() || self.routine.set_of {
            return self.finish(None);
        }
        if primary.is_empty() {
            return self.finish(None);
        }
        if !self.outputs.is_empty() {
            return self.finish(Some(Value::Null));
        }
        if self.routine.ret.oid == TRIGGER
            && let [binding] = secondary.as_slice()
            && primary.replace(' ', "").eq_ignore_ascii_case("select$1;")
            && let Some(index) = self.find(binding)
            && self.variables[index].columns.is_some()
        {
            return Ok(self.variables[index].value.clone());
        }
        let bindings = self.expand_whole_row(primary, secondary, "query")?;
        let result = self.query(ctx, primary, &bindings)?;
        if result.rows.len() > 1 {
            return Err(PgError::new(code::CARDINALITY_VIOLATION, "query returned more than one row"));
        }
        let row = result.rows.into_iter().next();
        let value = match row {
            Some(row) if row.len() == 1 => row.into_iter().next().unwrap_or(Value::Null),
            Some(row) => Value::Record(row),
            None => Value::Null,
        };
        if self.routine.ret.oid == TRIGGER || self.routine.ret.oid == oid::RECORD {
            return Ok(value);
        }
        crate::routines::result_value(self.routine, vec![vec![value]])
    }

    /// finish returns the result of a routine that returns without a value: its gathered rows, its output parameters,
    /// or NULL, failing as Postgres does for any other routine.
    fn finish(&mut self, returned: Option<Value>) -> Result<Value> {
        if self.routine.set_of {
            let rows = self.returned.take().unwrap_or_default();
            return crate::routines::result_value(self.routine, rows);
        }
        if !self.outputs.is_empty() {
            let mut values = Vec::with_capacity(self.outputs.len());
            for name in &self.outputs {
                values.push(self.variables[self.variable(name)?].value.clone());
            }
            return crate::routines::result_value(self.routine, vec![values]);
        }
        match returned {
            Some(value) => Ok(value),
            None if self.routine.ret.oid == VOID => Ok(Value::Null),
            None if self.routine.ret.oid == TRIGGER => {
                Err(PgError::new("2F005", "control reached end of trigger procedure without RETURN"))
            }
            None => Err(PgError::new("2F005", "control reached end of function without RETURN")),
        }
    }
}

/// record_fields returns the fields of a row value, anonymous or of a composite type, or None for any other value.
fn record_fields(value: Value) -> Option<Vec<Value>> {
    match value {
        Value::Record(fields) => Some(fields),
        Value::Composite(c) => Some(c.fields),
        _ => None,
    }
}

/// perform_query returns the query PERFORM runs, which the parser gives with PERFORM in place of SELECT.
fn perform_query(text: &str) -> String {
    let trimmed = text.trim_start();
    match trimmed.get(..7) {
        Some(word) if word.eq_ignore_ascii_case("perform") => format!("SELECT{}", &trimmed[7..]),
        _ => text.to_string(),
    }
}

/// returns_rows reports whether a statement returns rows that need an INTO clause: a SELECT, or a data-modifying
/// statement with RETURNING.
fn returns_rows(statement: &NodeEnum) -> bool {
    match statement {
        NodeEnum::SelectStmt(select) => select.into_clause.is_none(),
        NodeEnum::InsertStmt(insert) => !insert.returning_list.is_empty(),
        NodeEnum::UpdateStmt(update) => !update.returning_list.is_empty(),
        NodeEnum::DeleteStmt(delete) => !delete.returning_list.is_empty(),
        _ => false,
    }
}

/// field_index returns the position of a record's field by name, ignoring case.
fn field_index(columns: &[(String, ColumnType)], field: &str) -> Option<usize> {
    columns
        .iter()
        .position(|(name, _)| name == field)
        .or_else(|| columns.iter().position(|(name, _)| name.eq_ignore_ascii_case(field)))
}

/// not_assigned returns Postgres' error for reading a field of a RECORD nothing was assigned to.
fn not_assigned(name: &str) -> PgError {
    PgError {
        detail: Some("The tuple structure of a not-yet-assigned record is indeterminate.".into()),
        ..PgError::new(code::OBJECT_NOT_IN_PREREQUISITE_STATE, format!("record \"{name}\" is not assigned yet"))
    }
}

/// record_has_no_field returns Postgres' error for a field a record lacks.
fn record_has_no_field(name: &str, field: &str) -> PgError {
    PgError::new(code::UNDEFINED_COLUMN, format!("record \"{name}\" has no field \"{field}\""))
}

/// table_columns returns the columns of a table or view a name finds on the search path, or None when it finds none.
fn table_columns(ctx: &mut Ctx<'_>, name: &str) -> Result<Option<Vec<(String, ColumnType)>>> {
    let (schema, relation) = match name.split_once('.') {
        Some((schema, relation)) => (schema.to_string(), relation.to_string()),
        None => (String::new(), name.to_string()),
    };
    let schemas = if schema.is_empty() { ctx.session.search_path() } else { vec![schema] };
    for schema in schemas {
        if let Some(table) = ctx.txn.table(ctx.db, &schema, &relation)? {
            return Ok(Some(table.columns.iter().map(|c| (c.name.clone(), c.ty)).collect()));
        }
    }
    Ok(None)
}

/// sql_state returns the SQLSTATE that an ERRCODE option's text names directly, quoted or not.
fn sql_state(text: &str) -> Option<String> {
    let text = text.trim();
    let text = text.strip_prefix('\'').and_then(|t| t.strip_suffix('\'')).unwrap_or(text);
    (text.len() == 5 && text.chars().all(|c| c.is_ascii_digit() || c.is_ascii_uppercase())).then(|| text.to_string())
}

/// CONDITIONS are the condition names that RAISE accepts for ERRCODE, with their SQLSTATEs.
const CONDITIONS: [(&str, &str); 14] = [
    ("raise_exception", "P0001"),
    ("no_data_found", "P0002"),
    ("too_many_rows", "P0003"),
    ("assert_failure", "P0004"),
    ("unique_violation", "23505"),
    ("foreign_key_violation", "23503"),
    ("not_null_violation", "23502"),
    ("check_violation", "23514"),
    ("division_by_zero", "22012"),
    ("invalid_parameter_value", "22023"),
    ("data_exception", "22000"),
    ("undefined_table", "42P01"),
    ("undefined_column", "42703"),
    ("feature_not_supported", "0A000"),
];

/// error_code returns the SQLSTATE an ERRCODE option names, as a code or a condition name, failing as Postgres does
/// for anything else.
fn error_code(text: &str) -> Result<&'static str> {
    let lowered = text.to_lowercase();
    if let Some((_, state)) = CONDITIONS.iter().find(|(name, _)| *name == lowered) {
        return Ok(state);
    }
    if sql_state(text).is_some() {
        return Ok(intern(text));
    }
    Err(PgError::new(code::UNDEFINED_OBJECT, format!("unrecognized exception condition \"{text}\"")))
}

/// CODES are the SQLSTATEs that RAISE has named, kept for the life of the server.
static CODES: std::sync::Mutex<Vec<&'static str>> = std::sync::Mutex::new(Vec::new());

/// intern returns a static copy of a SQLSTATE, keeping one copy of each.
fn intern(state: &str) -> &'static str {
    let mut codes = CODES.lock().unwrap_or_else(|e| e.into_inner());
    if let Some(existing) = codes.iter().find(|c| **c == state) {
        return existing;
    }
    let leaked: &'static str = Box::leak(state.to_string().into_boxed_str());
    codes.push(leaked);
    leaked
}

/// visible reports whether a notice of the severity reaches the client under client_min_messages.
fn visible(ctx: &Ctx<'_>, severity: &str) -> bool {
    if severity == "INFO" {
        return true;
    }
    let rank = |level: &str| match level.to_lowercase().as_str() {
        "debug5" | "debug4" | "debug3" | "debug2" | "debug1" | "debug" => 0,
        "log" => 1,
        "notice" => 2,
        "warning" => 3,
        "error" => 4,
        _ => 2,
    };
    let minimum = ctx.session.settings.get("client_min_messages").unwrap_or_else(|| "notice".into());
    rank(severity) >= rank(&minimum)
}
