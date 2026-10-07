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

//! The function, procedure, and trigger definitions that pg_get_functiondef, pg_get_function_arguments,
//! pg_get_function_result, and pg_get_triggerdef print, as Postgres' ruleutils.c prints them.

use std::collections::HashMap;
use std::sync::OnceLock;

use pg_query::NodeEnum;
use pg_query::protobuf::FunctionParameterMode;

use crate::engine::quote_identifier;
use crate::error::{PgError, Result, code};
use crate::expr::node_name;
use crate::pgcatalog::routines::{create_statement, routine_oid, source};
use crate::pgcatalog::{builtin, lookup};
use crate::query::Ctx;
use crate::routines::{Mode, Routine};
use crate::triggers::{AFTER, BEFORE, Event, names};
use crate::types::Value;

/// BUILTIN_DEFAULTS are the defaults of the built-in functions that have them, by OID, for their last input
/// parameters.
const BUILTIN_DEFAULTS: &[(u32, &[&str])] = &[
    (1177, &["'{}'::jsonb", "false"]),
    (1179, &["'{}'::jsonb", "false"]),
    (1180, &["'{}'::jsonb", "false"]),
    (1268, &["true"]),
    (2023, &["'{}'::jsonb", "false"]),
    (2030, &["'{}'::jsonb", "false"]),
    (2096, &["0"]),
    (2172, &["false"]),
    (2739, &["true"]),
    (3305, &["true"]),
    (3436, &["true", "60"]),
    (3464, &["0", "0", "0", "0", "0", "0", "0.0"]),
    (3579, &["false"]),
    (3779, &["false", "false"]),
    (3782, &["'{}'::text[]"]),
    (3783, &["'{}'::text[]"]),
    (3784, &["'{}'::text[]"]),
    (3785, &["'{}'::text[]"]),
    (3786, &["false", "false"]),
    (3960, &["false"]),
    (3961, &["false"]),
    (4005, &["'{}'::jsonb", "false"]),
    (4006, &["'{}'::jsonb", "false"]),
    (4007, &["'{}'::jsonb", "false"]),
    (4008, &["'{}'::jsonb", "false"]),
    (4009, &["'{}'::jsonb", "false"]),
    (4350, &["'NFC'::text"]),
    (4351, &["'NFC'::text"]),
    (5054, &["true", "'use_json_null'::text"]),
];

/// LIST_SETTINGS are the settings whose values SET clauses print as one quoted literal for each list item.
const LIST_SETTINGS: &[&str] = &[
    "search_path",
    "temp_tablespaces",
    "session_preload_libraries",
    "shared_preload_libraries",
    "local_preload_libraries",
];

/// ProcDef is what the pg_get_function family prints about a function or procedure.
pub struct ProcDef {
    schema: String,
    name: String,
    /// The kind letter of pg_proc's prokind.
    kind: char,
    /// The parameters, each as its mode letter of pg_proc's proargmodes, name, type name, and default.
    params: Vec<(char, String, String, Option<String>)>,
    /// The result type's name, after SETOF for a set.
    result: String,
    language: String,
    /// The options that print on one line, such as `IMMUTABLE STRICT`.
    options: Vec<String>,
    /// The SET clauses, one for each line.
    settings: Vec<String>,
    /// The body, as `AS` with the quoted source or as a SQL-standard body.
    body: String,
    /// The SQL-standard body, which pg_get_function_sqlbody returns.
    sql_body: Option<String>,
}

/// type_name returns a type's name as format_type_be prints it.
fn type_name(type_oid: u32) -> String {
    crate::cast::format_type(type_oid, None).unwrap_or_else(|| "???".into())
}

/// float_text prints a cost or row estimate as C's `%g` does for the values they take.
fn float_text(value: f64) -> String {
    if value.fract() == 0.0 && value.abs() < 1e6 { format!("{}", value as i64) } else { format!("{value}") }
}

/// quote_literal quotes text as a string literal.
fn quote_literal(text: &str) -> String {
    format!("'{}'", text.replace('\'', "''"))
}

/// quoted_source returns the AS clause of a body's source, quoted in dollars with a tag the source does not hold.
fn quoted_source(source: &str, binary: Option<&str>, procedure: bool) -> String {
    let mut tag = format!("${}", if procedure { "procedure" } else { "function" });
    while source.contains(&tag) {
        tag.push('x');
    }
    tag.push('$');
    let binary = binary.map(|b| format!("{}, ", quote_literal(b))).unwrap_or_default();
    format!("AS {binary}{tag}{source}{tag}")
}

/// mode_letter returns the letter of pg_proc's proargmodes for a routine's parameter mode.
fn mode_letter(mode: Mode) -> char {
    match mode {
        Mode::In => 'i',
        Mode::Out => 'o',
        Mode::InOut => 'b',
        Mode::Variadic => 'v',
    }
}

/// texts returns the text of each element of an array value, or the words of a vector's text.
fn texts(value: &Value) -> Vec<String> {
    match value {
        Value::Array(array) => array.values.iter().map(|v| v.output().unwrap_or_default()).collect(),
        Value::Null => Vec::new(),
        other => other.output().unwrap_or_default().split_whitespace().map(str::to_string).collect(),
    }
}

/// BuiltinProcs are the rows of the built-in pg_proc by OID, with the positions of its columns by name.
type BuiltinProcs = (HashMap<u32, Vec<Value>>, HashMap<&'static str, usize>);

/// builtin_procs returns the rows of the built-in pg_proc by OID, with the positions of its columns by name.
fn builtin_procs() -> &'static BuiltinProcs {
    static PROCS: OnceLock<BuiltinProcs> = OnceLock::new();
    PROCS.get_or_init(|| {
        let Some(table) = lookup("pg_catalog", "pg_proc") else { return Default::default() };
        let columns: HashMap<&'static str, usize> =
            table.columns.iter().enumerate().map(|(i, c)| (c.name, i)).collect();
        let oid = columns.get("oid").copied().unwrap_or(0);
        let rows = builtin::rows(table)
            .into_iter()
            .filter_map(|row| match row[oid] {
                Value::Oid(o) => Some((o, row)),
                _ => None,
            })
            .collect();
        (rows, columns)
    })
}

/// builtin_def returns the definition of the built-in function or procedure with the OID.
fn builtin_def(oid: u32) -> Option<ProcDef> {
    let (rows, columns) = builtin_procs();
    let row = rows.get(&oid)?;
    let get = |name: &str| columns.get(name).map_or(&Value::Null, |&i| &row[i]);
    let text = |name: &str| get(name).output().unwrap_or_default();
    let flag = |name: &str| matches!(get(name), Value::Bool(true));
    let float = |name: &str| text(name).parse::<f64>().unwrap_or(0.0);
    let all_types: Vec<u32> = match texts(get("proallargtypes")) {
        types if !types.is_empty() => types,
        _ => texts(get("proargtypes")),
    }
    .iter()
    .filter_map(|t| t.parse().ok())
    .collect();
    let modes = texts(get("proargmodes"));
    let names = texts(get("proargnames"));
    let mut params: Vec<(char, String, String, Option<String>)> = all_types
        .iter()
        .enumerate()
        .map(|(i, &t)| {
            let mode = modes.get(i).and_then(|m| m.chars().next()).unwrap_or('i');
            (mode, names.get(i).cloned().unwrap_or_default(), type_name(t), None)
        })
        .collect();
    if let Some((_, defaults)) = BUILTIN_DEFAULTS.iter().find(|(o, _)| *o == oid) {
        let inputs: Vec<usize> = (0..params.len()).filter(|&i| matches!(params[i].0, 'i' | 'b' | 'v')).collect();
        for (&i, default) in inputs[inputs.len() - defaults.len()..].iter().zip(defaults.iter()) {
            params[i].3 = Some(default.to_string());
        }
    }
    let kind = text("prokind").chars().next().unwrap_or('f');
    let language = match text("prolang").as_str() {
        "12" => "internal",
        "13" => "c",
        _ => "sql",
    };
    let default_cost = if matches!(language, "internal" | "c") { 1.0 } else { 100.0 };
    let mut options = Vec::new();
    if kind == 'w' {
        options.push("WINDOW".to_string());
    }
    match text("provolatile").as_str() {
        "i" => options.push("IMMUTABLE".into()),
        "s" => options.push("STABLE".into()),
        _ => {}
    }
    match text("proparallel").as_str() {
        "s" => options.push("PARALLEL SAFE".into()),
        "r" => options.push("PARALLEL RESTRICTED".into()),
        _ => {}
    }
    for (column, option) in
        [("proisstrict", "STRICT"), ("prosecdef", "SECURITY DEFINER"), ("proleakproof", "LEAKPROOF")]
    {
        if flag(column) {
            options.push(option.into());
        }
    }
    if float("procost") != default_cost {
        options.push(format!("COST {}", float_text(float("procost"))));
    }
    if float("prorows") > 0.0 && float("prorows") != 1000.0 {
        options.push(format!("ROWS {}", float_text(float("prorows"))));
    }
    if let Value::Reg(support) = get("prosupport")
        && support.oid != 0
    {
        options.push(format!("SUPPORT {}", support.name));
    }
    let set_of = if flag("proretset") { "SETOF " } else { "" };
    let probin = get("probin").output();
    Some(ProcDef {
        schema: "pg_catalog".into(),
        name: text("proname"),
        kind,
        params,
        result: format!("{set_of}{}", type_name(text("prorettype").parse().unwrap_or(0))),
        language: language.into(),
        options,
        settings: Vec::new(),
        body: quoted_source(&text("prosrc"), probin.as_deref(), kind == 'p'),
        sql_body: None,
    })
}

impl ProcDef {
    /// arguments prints the parameters as pg_get_function_arguments does, with their defaults when asked.
    pub fn arguments(&self, defaults: bool) -> String {
        let printed: Vec<String> = self
            .params
            .iter()
            .filter(|p| p.0 != 't')
            .map(|(mode, name, ty, default)| {
                let mode = match mode {
                    'o' => "OUT ",
                    'b' => "INOUT ",
                    'v' => "VARIADIC ",
                    _ if self.kind == 'p' => "IN ",
                    _ => "",
                };
                let name = if name.is_empty() { String::new() } else { format!("{} ", quote_identifier(name)) };
                let default =
                    default.as_ref().filter(|_| defaults).map(|d| format!(" DEFAULT {d}")).unwrap_or_default();
                format!("{mode}{name}{ty}{default}")
            })
            .collect();
        printed.join(", ")
    }

    /// result prints the result type as pg_get_function_result does, or None for a procedure.
    pub fn result(&self) -> Option<String> {
        if self.kind == 'p' {
            return None;
        }
        let table: Vec<String> = self
            .params
            .iter()
            .filter(|p| p.0 == 't')
            .map(|(_, name, ty, _)| format!("{} {ty}", quote_identifier(name)))
            .collect();
        Some(if table.is_empty() { self.result.clone() } else { format!("TABLE({})", table.join(", ")) })
    }

    /// sql_body returns the SQL-standard body as pg_get_function_sqlbody prints it.
    pub fn sql_body(&self) -> Option<String> {
        self.sql_body.clone()
    }

    /// definition prints the CREATE OR REPLACE statement as pg_get_functiondef does, failing for an aggregate.
    pub fn definition(&self) -> Result<String> {
        if self.kind == 'a' {
            return Err(PgError::new(code::WRONG_OBJECT_TYPE, format!("\"{}\" is an aggregate function", self.name)));
        }
        let kind = if self.kind == 'p' { "PROCEDURE" } else { "FUNCTION" };
        let mut out = format!(
            "CREATE OR REPLACE {kind} {}.{}({})\n",
            quote_identifier(&self.schema),
            quote_identifier(&self.name),
            self.arguments(true)
        );
        if let Some(result) = self.result() {
            out.push_str(&format!(" RETURNS {result}\n"));
        }
        out.push_str(&format!(" LANGUAGE {}\n", quote_identifier(&self.language)));
        if !self.options.is_empty() {
            out.push_str(&format!(" {}\n", self.options.join(" ")));
        }
        for setting in &self.settings {
            out.push_str(&format!(" {setting}\n"));
        }
        out.push_str(&self.body);
        out.push('\n');
        Ok(out)
    }
}

/// setting returns the SET clause of a routine's SET option.
fn setting(set: &pg_query::protobuf::VariableSetStmt) -> String {
    let values: Vec<String> = set
        .args
        .iter()
        .filter_map(|arg| match arg.node.as_ref()? {
            NodeEnum::AConst(c) => match c.val.as_ref()? {
                pg_query::protobuf::a_const::Val::Sval(s) => Some(s.sval.clone()),
                pg_query::protobuf::a_const::Val::Ival(i) => Some(i.ival.to_string()),
                pg_query::protobuf::a_const::Val::Fval(f) => Some(f.fval.clone()),
                pg_query::protobuf::a_const::Val::Boolval(b) => Some(b.boolval.to_string()),
                pg_query::protobuf::a_const::Val::Bsval(b) => Some(b.bsval.clone()),
            },
            other => node_name(&pg_query::Node { node: Some(other.clone()) }).map(str::to_string),
        })
        .collect();
    let value = if LIST_SETTINGS.contains(&set.name.as_str()) {
        values.iter().map(|v| quote_literal(v)).collect::<Vec<_>>().join(", ")
    } else {
        quote_literal(&values.join(", "))
    };
    format!("SET {} TO {value}", quote_identifier(&set.name))
}

impl Ctx<'_> {
    /// routine_def returns the definition of the function or procedure with the OID, a user routine or a built-in
    /// one, or None when nothing has the OID.
    pub fn routine_def(&mut self, oid: u32) -> Result<Option<ProcDef>> {
        let routines = self.routines()?;
        match routines.iter().find(|r| routine_oid(r) == oid) {
            Some(routine) => self.user_routine_def(routine).map(Some),
            None => Ok(builtin_def(oid)),
        }
    }

    /// user_routine_def returns the definition of a user routine, reading its options from the statement that
    /// created it.
    fn user_routine_def(&mut self, routine: &Routine) -> Result<ProcDef> {
        let create = create_statement(routine).unwrap_or_default();
        let mut inputs = routine.params.iter();
        let mut columns = routine.columns.iter();
        let mut params = Vec::new();
        for node in &create.parameters {
            let Some(NodeEnum::FunctionParameter(param)) = node.node.as_ref() else { continue };
            if param.mode == FunctionParameterMode::FuncParamTable as i32 {
                if let Some((name, ty)) = columns.next() {
                    params.push(('t', name.clone(), type_name(ty.oid), None));
                }
                continue;
            }
            let Some(p) = inputs.next() else { continue };
            let default = match &p.default {
                Some(default) => {
                    let mut analyzer = crate::ruleutils::Analyzer::new(self, Vec::new());
                    Some(analyzer.deparse(default, Some(crate::catalog::ColumnType { modifier: -1, ..p.ty }), false)?)
                }
                None => None,
            };
            params.push((mode_letter(p.mode), p.name.clone(), type_name(p.ty.oid), default));
        }
        let declared = create.return_type.as_ref();
        let result = match declared {
            Some(t) if routine.ret.oid == crate::oid::RECORD => {
                let names: Vec<&str> = t.names.iter().filter_map(node_name).collect();
                match names.last() {
                    Some(&"record") | None => type_name(routine.ret.oid),
                    Some(name) => quote_identifier(name),
                }
            }
            _ => type_name(routine.ret.oid),
        };
        let set_of = if routine.set_of && declared.is_some() { "SETOF " } else { "" };
        let mut language = String::new();
        let mut options = Vec::new();
        let mut settings = Vec::new();
        let (mut volatility, mut parallel, mut security, mut leakproof) = ("", "", false, false);
        let (mut cost, mut rows) = (None, None);
        for node in &create.options {
            let Some(NodeEnum::DefElem(def)) = node.node.as_ref() else { continue };
            let arg = def.arg.as_deref().and_then(|a| a.node.as_ref());
            let number = || match arg {
                Some(NodeEnum::Integer(i)) => Some(i.ival as f64),
                Some(NodeEnum::Float(f)) => f.fval.parse().ok(),
                _ => None,
            };
            match (def.defname.as_str(), arg) {
                ("language", Some(NodeEnum::String(s))) => language = s.sval.to_lowercase(),
                ("volatility", Some(NodeEnum::String(s))) if s.sval == "immutable" => volatility = "IMMUTABLE",
                ("volatility", Some(NodeEnum::String(s))) if s.sval == "stable" => volatility = "STABLE",
                ("parallel", Some(NodeEnum::String(s))) if s.sval == "safe" => parallel = "PARALLEL SAFE",
                ("parallel", Some(NodeEnum::String(s))) if s.sval == "restricted" => parallel = "PARALLEL RESTRICTED",
                ("security", Some(NodeEnum::Boolean(b))) => security = b.boolval,
                ("leakproof", Some(NodeEnum::Boolean(b))) => leakproof = b.boolval,
                ("cost", _) => cost = number(),
                ("rows", _) => rows = number(),
                ("set", Some(NodeEnum::VariableSetStmt(set))) => settings.push(setting(set)),
                _ => {}
            }
        }
        let sql_body = match create.sql_body.as_deref().and_then(|b| b.node.as_ref()) {
            Some(NodeEnum::ReturnStmt(ret)) => {
                let columns = routine.params.iter().map(|p| (p.name.clone(), p.ty)).collect();
                let mut analyzer = crate::ruleutils::Analyzer::new(self, columns);
                let text = crate::ddl::expression_text(ret.returnval.as_deref().unwrap_or(&Default::default()))?;
                Some(format!("RETURN {}", analyzer.deparse(&text, None, false)?))
            }
            Some(_) => Some(source(routine)),
            None => None,
        };
        if language.is_empty() {
            language = "sql".into();
        }
        options.extend([volatility, parallel].into_iter().filter(|o| !o.is_empty()).map(str::to_string));
        if routine.strict {
            options.push("STRICT".into());
        }
        if security {
            options.push("SECURITY DEFINER".into());
        }
        if leakproof {
            options.push("LEAKPROOF".into());
        }
        if let Some(cost) = cost.filter(|&c| c != 100.0) {
            options.push(format!("COST {}", float_text(cost)));
        }
        if let Some(rows) = rows.filter(|&r| routine.set_of && r > 0.0 && r != 1000.0) {
            options.push(format!("ROWS {}", float_text(rows)));
        }
        let body = match &sql_body {
            Some(body) => body.clone(),
            None => quoted_source(&source(routine), None, routine.procedure),
        };
        Ok(ProcDef {
            schema: routine.schema.clone(),
            name: routine.name.clone(),
            kind: if routine.procedure { 'p' } else { 'f' },
            params,
            result: format!("{set_of}{result}"),
            language,
            options,
            settings,
            body,
            sql_body,
        })
    }
}

impl Ctx<'_> {
    /// trigger_definition_of returns the CREATE TRIGGER statement of the trigger with the OID as pg_get_triggerdef
    /// prints it, naming the table with its schema unless printing prettily and the table is visible, or None when
    /// no trigger has the OID.
    pub fn trigger_definition_of(&mut self, oid: u32, pretty: bool) -> Result<Option<String>> {
        let triggers = self.triggers()?;
        let Some(trigger) = triggers.iter().find(|t| crate::catalog::oids::oid(&t.id) == oid) else { return Ok(None) };
        let (schema, table, name) = names(trigger);
        let columns = self.txn.table(self.db, &schema, &table)?.map(|t| t.columns).unwrap_or_default();
        let mut events = Vec::new();
        for event in [Event::Insert, Event::Delete, Event::Update, Event::Truncate] {
            let Some(found) = trigger.events.iter().find(|e| e.event_type == event as u8) else { continue };
            let mut text = match event {
                Event::Insert => "INSERT".to_string(),
                Event::Delete => "DELETE".to_string(),
                Event::Update => "UPDATE".to_string(),
                Event::Truncate => "TRUNCATE".to_string(),
            };
            if !found.column_names.is_empty() {
                let names: Vec<String> =
                    found.column_names.iter().map(|c| quote_identifier(&String::from_utf8_lossy(c))).collect();
                text.push_str(&format!(" OF {}", names.join(", ")));
            }
            events.push(text);
        }
        let timing = match trigger.timing {
            BEFORE => "BEFORE",
            AFTER => "AFTER",
            _ => "INSTEAD OF",
        };
        let relation = if pretty && self.session.search_path().contains(&schema) {
            quote_identifier(&table)
        } else {
            format!("{}.{}", quote_identifier(&schema), quote_identifier(&table))
        };
        let mut out =
            format!("CREATE TRIGGER {} {timing} {} ON {relation} ", quote_identifier(&name), events.join(" OR "));
        let (old, new) = (&trigger.old_transition_name, &trigger.new_transition_name);
        if !old.is_empty() || !new.is_empty() {
            out.push_str("REFERENCING ");
            if !old.is_empty() {
                out.push_str(&format!("OLD TABLE AS {} ", quote_identifier(&String::from_utf8_lossy(old))));
            }
            if !new.is_empty() {
                out.push_str(&format!("NEW TABLE AS {} ", quote_identifier(&String::from_utf8_lossy(new))));
            }
        }
        out.push_str(if trigger.for_each_row { "FOR EACH ROW " } else { "FOR EACH STATEMENT " });
        let definition = String::from_utf8_lossy(&trigger.definition);
        let condition = pg_query::parse(&definition).ok().and_then(|result| {
            match result.protobuf.stmts.into_iter().next()?.stmt?.node? {
                NodeEnum::CreateTrigStmt(create) => create.when_clause,
                _ => None,
            }
        });
        if let Some(condition) = condition {
            let text = crate::ddl::expression_text(&condition)?;
            let columns = columns.iter().map(|c| (c.name.clone(), c.ty)).collect();
            let printed = crate::ruleutils::Analyzer::trigger(self, columns).deparse(&text, None, pretty)?;
            out.push_str(&format!("WHEN ({printed}) "));
        }
        let segments = crate::catalog::id::segments(&trigger.function);
        let (function_schema, function) =
            (segments.first().cloned().unwrap_or_default(), segments.get(1).cloned().unwrap_or_default());
        let function = if function_schema.is_empty() || self.session.search_path().contains(&function_schema) {
            quote_identifier(&function)
        } else {
            format!("{}.{}", quote_identifier(&function_schema), quote_identifier(&function))
        };
        let arguments: Vec<String> =
            trigger.arguments.iter().map(|a| quote_literal(&String::from_utf8_lossy(a))).collect();
        out.push_str(&format!("EXECUTE FUNCTION {function}({})", arguments.join(", ")));
        Ok(Some(out))
    }
}
