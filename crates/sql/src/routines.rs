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

//! User-defined functions: Go's function root objects, CREATE and DROP FUNCTION, choosing among the user and built-in
//! overloads of a name, and running SQL-language functions.

use std::sync::{Arc, OnceLock};

use doltdb::database::Database;
use doltdb::root::Root;
use objects::{Function, Operation, Parameter, Procedure};
use pg_query::protobuf::{
    CreateFunctionStmt, DropStmt, FunctionParameter, FunctionParameterMode, ObjectWithArgs, TypeName,
};
use pg_query::{Node, NodeEnum};
use store::Hash;

use crate::auth::Object;
use crate::cast::type_display;
use crate::catalog::id::{self, SECTION_FUNCTION, SECTION_PROCEDURE};
use crate::catalog::{ColumnType, builtin_type, builtin_type_by_id};
use crate::error::{PgError, Result, code};
use crate::expr::{Binder, Bound, Expr, Scope, arg_location, assignable, coerce, node_name, position, typ};
use crate::query::Ctx;
use crate::types::Value;
use crate::{Outcome, functions, oid};

/// COLLECTION is the position of the function collection among a root value's root object collections.
pub const COLLECTION: usize = 2;

/// PROCEDURES is the position of the procedure collection among a root value's root object collections.
const PROCEDURES: usize = 6;

/// VOID is the type of a function that returns nothing.
pub const VOID: u32 = 2278;

/// TRIGGER is the type of a trigger function's result.
pub const TRIGGER: u32 = 2279;

/// MAX_DEPTH is how deeply function calls may nest before failing as Postgres does when it runs out of stack.
const MAX_DEPTH: usize = 400;

/// Mode is how a parameter passes its value, numbered as Go stores it.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Mode {
    In,
    Out,
    InOut,
    Variadic,
}

impl Mode {
    /// from_stored returns the mode Go stores as the number.
    fn from_stored(mode: u8) -> Mode {
        match mode {
            1 => Mode::Out,
            2 => Mode::InOut,
            3 => Mode::Variadic,
            _ => Mode::In,
        }
    }

    /// stored returns the number Go stores for the mode.
    fn stored(self) -> u8 {
        match self {
            Mode::In => 0,
            Mode::Out => 1,
            Mode::InOut => 2,
            Mode::Variadic => 3,
        }
    }

    /// is_input reports whether a caller passes the parameter.
    pub fn is_input(self) -> bool {
        self != Mode::Out
    }

    /// is_output reports whether the parameter is part of the result.
    pub fn is_output(self) -> bool {
        matches!(self, Mode::Out | Mode::InOut)
    }
}

/// Param is a parameter of a routine.
#[derive(Clone, Debug, PartialEq)]
pub struct Param {
    pub name: String,
    pub ty: ColumnType,
    pub mode: Mode,
    /// The default's expression text, for an input parameter that has one.
    pub default: Option<String>,
}

/// Body is what a routine runs.
#[derive(Clone, Debug, PartialEq)]
pub enum Body {
    /// SQL statements, separated by semicolons.
    Sql(String),
    /// PL/pgSQL compiled into Go's interpreter operations.
    PlPgSql(Vec<Operation>),
    /// A function of an extension library, which only emulated extensions provide.
    External,
}

/// Routine is a user-defined function as Go stores it, with its types resolved.
#[derive(Debug)]
pub struct Routine {
    pub object: Function,
    /// Whether the routine is a procedure, which only CALL runs.
    pub procedure: bool,
    pub schema: String,
    pub name: String,
    pub params: Vec<Param>,
    /// The result type, which is record for a result of several columns.
    pub ret: ColumnType,
    /// The result's columns when it is a record whose columns the routine names.
    pub columns: Vec<(String, ColumnType)>,
    pub set_of: bool,
    pub strict: bool,
    pub body: Body,
    /// The row types of the tables whose rows the routine takes or returns, which each statement that uses it
    /// registers.
    pub row_types: Vec<objects::SerializedType>,
    /// The statements of a SQL body, parsed on first use.
    statements: OnceLock<Result<Vec<NodeEnum>>>,
}

impl PartialEq for Routine {
    fn eq(&self, other: &Routine) -> bool {
        self.object.id == other.object.id
    }
}

impl Routine {
    /// load resolves the types of a stored function, finding the columns of a table's row type with the function.
    pub fn load(object: Function, row_type: &mut RowTypes<'_>) -> Result<Routine> {
        Routine::load_routine(object, false, row_type)
    }

    /// load_procedure resolves the types of a stored procedure, which returns its output parameters.
    fn load_procedure(procedure: Procedure, row_type: &mut RowTypes<'_>) -> Result<Routine> {
        let object = Function {
            id: procedure.id,
            all_params: procedure.all_params,
            definition: procedure.definition,
            extension_name: procedure.extension_name,
            extension_symbol: procedure.extension_symbol,
            operations: procedure.operations,
            sql_definition: procedure.sql_definition,
            ..Function::default()
        };
        Routine::load_routine(object, true, row_type)
    }

    /// load_routine resolves the types of a stored function or procedure.
    fn load_routine(object: Function, procedure: bool, row_type: &mut RowTypes<'_>) -> Result<Routine> {
        let mut segments = id::segments(&object.id).into_iter();
        let schema = segments.next().unwrap_or_default();
        let name = segments.next().unwrap_or_default();
        let mut row_types = Vec::new();
        let mut params = Vec::with_capacity(object.all_params.len());
        for param in &object.all_params {
            let ty = match type_from_id(&param.type_id) {
                Ok(ty) => ty,
                Err(err) => {
                    row_types.push(row_type(&param.type_id)?.ok_or(err)?);
                    typ(crate::usertypes::type_oid(&param.type_id))
                }
            };
            params.push(Param {
                name: String::from_utf8_lossy(&param.name).into_owned(),
                ty,
                mode: Mode::from_stored(param.mode),
                default: (!param.default.is_empty()).then(|| String::from_utf8_lossy(&param.default).into_owned()),
            });
        }
        let outputs: Vec<(String, ColumnType)> =
            params.iter().filter(|p| p.mode.is_output()).map(|p| (p.name.clone(), p.ty)).collect();
        let (ret, columns) = match table_columns(&object.return_type)? {
            _ if procedure => (typ(VOID), outputs),
            Some(columns) if columns.len() == 1 => {
                match crate::usertypes::get(columns[0].1.oid).map(|t| t.kind.clone()) {
                    Some(crate::usertypes::Kind::Composite(attributes)) => (columns[0].1, attributes),
                    _ => (columns[0].1, columns),
                }
            }
            Some(columns) => (typ(oid::RECORD), columns),
            None => match type_from_id(&object.return_type) {
                Ok(ty) => match crate::usertypes::get(ty.oid).map(|t| t.kind.clone()) {
                    Some(crate::usertypes::Kind::Composite(attributes)) => (ty, attributes),
                    _ => (ty, outputs),
                },
                Err(err) => {
                    let definition = row_type(&object.return_type)?.ok_or(err)?;
                    let columns = match crate::usertypes::UserType::from_definition(definition.clone()).kind {
                        crate::usertypes::Kind::Composite(columns) => columns,
                        _ => Vec::new(),
                    };
                    row_types.push(definition);
                    (typ(crate::usertypes::type_oid(&object.return_type)), columns)
                }
            },
        };
        let body = if !object.extension_name.is_empty() {
            Body::External
        } else if !object.sql_definition.is_empty() {
            Body::Sql(String::from_utf8_lossy(&object.sql_definition).into_owned())
        } else {
            Body::PlPgSql(object.operations.clone())
        };
        Ok(Routine {
            row_types,
            procedure,
            schema,
            name,
            params,
            ret,
            columns,
            set_of: object.set_of,
            strict: object.strict,
            body,
            statements: OnceLock::new(),
            object,
        })
    }

    /// internal returns a routine that only the server runs, such as a trigger's WHEN condition, with the result type.
    pub fn internal(object: Function, ret: ColumnType) -> Routine {
        Routine {
            row_types: Vec::new(),
            procedure: false,
            schema: String::new(),
            name: String::new(),
            params: Vec::new(),
            ret,
            columns: Vec::new(),
            set_of: false,
            strict: false,
            body: Body::PlPgSql(object.operations.clone()),
            statements: OnceLock::new(),
            object,
        }
    }

    /// inputs returns the parameters a caller passes.
    pub fn inputs(&self) -> impl Iterator<Item = &Param> {
        self.params.iter().filter(|p| p.mode.is_input())
    }

    /// signature returns the routine's name and input types as Postgres shows them, such as `f(integer, text)`.
    pub fn signature(&self) -> String {
        let types: Vec<_> = self.inputs().map(|p| type_display(p.ty.oid)).collect();
        format!("{}({})", self.name, types.join(", "))
    }

    /// kind returns what Postgres calls the routine in messages.
    pub fn kind(&self) -> &'static str {
        if self.procedure { "procedure" } else { "function" }
    }

    /// sql_statements returns the parsed statements of a SQL body.
    fn sql_statements(&self) -> Result<&[NodeEnum]> {
        let Body::Sql(text) = &self.body else { return Err(PgError::internal("a routine without a SQL body")) };
        let parsed = self.statements.get_or_init(|| parse_body(text));
        parsed.as_deref().map_err(Clone::clone)
    }
}

/// parse_body parses the statements of a SQL body, reading a body that is one RETURN as the SELECT of its value.
fn parse_body(text: &str) -> Result<Vec<NodeEnum>> {
    let trimmed = text.trim_start();
    let text = match trimmed.get(..6) {
        Some(word) if word.eq_ignore_ascii_case("return") => format!("SELECT{}", &trimmed[6..]),
        _ => text.to_string(),
    };
    let result = pg_query::parse(&text).map_err(|err| PgError::new(code::SYNTAX_ERROR, err.to_string()))?;
    Ok(result.protobuf.stmts.into_iter().filter_map(|raw| raw.stmt.and_then(|stmt| stmt.node)).collect())
}

/// type_from_id returns the type a stored type ID names.
fn type_from_id(type_id: &[u8]) -> Result<ColumnType> {
    match builtin_type_by_id(type_id) {
        Some(t) => Ok(typ(t.oid)),
        None => match crate::usertypes::get(crate::usertypes::type_oid(type_id)) {
            Some(t) => Ok(typ(t.oid)),
            None => Err(PgError::unsupported(format!("the type {}", id::segments(type_id).join(".")))),
        },
    }
}

/// type_id returns the stored ID of a type.
fn type_id(ty: ColumnType) -> Vec<u8> {
    crate::usertypes::type_id(ty.oid)
}

/// TABLE_PREFIX starts the name of the anonymous type that Go stores as the result of a function returning a table.
const TABLE_PREFIX: &str = "table(";

/// table_columns returns the columns of the anonymous type of a function returning a table, which Go names
/// `table(name:TYPE,...)` with each type spelled as its SQL syntax, or None for any other type.
fn table_columns(type_id: &[u8]) -> Result<Option<Vec<(String, ColumnType)>>> {
    let segments = id::segments(type_id);
    let [schema, name] = segments.as_slice() else { return Ok(None) };
    let Some(list) = name.strip_prefix(TABLE_PREFIX).and_then(|l| l.strip_suffix(')')) else { return Ok(None) };
    if !schema.is_empty() {
        return Ok(None);
    }
    let mut columns = Vec::new();
    let (mut depth, mut start) = (0, 0);
    let bytes = list.as_bytes();
    for i in 0..=bytes.len() {
        match bytes.get(i) {
            Some(b'(') => depth += 1,
            Some(b')') => depth -= 1,
            Some(b',') if depth > 0 => {}
            Some(b',') | None => {
                let item = &list[start..i];
                let (column, spelling) = item.split_once(':').ok_or_else(|| PgError::internal("a table column"))?;
                let column = match column.strip_prefix('"').and_then(|c| c.strip_suffix('"')) {
                    Some(quoted) => quoted.replace("\"\"", "\""),
                    None => column.to_string(),
                };
                columns.push((column, parse_type(spelling)?));
                start = i + 1;
            }
            _ => {}
        }
    }
    Ok(Some(columns))
}

/// parse_type returns the type that SQL syntax spells, where `BYTES` is Go's spelling of bytea.
pub(crate) fn parse_type(spelling: &str) -> Result<ColumnType> {
    let (base, arrays) = spelling.split_at(spelling.find('[').unwrap_or(spelling.len()));
    let spelling = if base.eq_ignore_ascii_case("bytes") { format!("bytea{arrays}") } else { spelling.to_string() };
    let node = crate::parse::expression_node(&format!("NULL::{spelling}"))?;
    let Some(NodeEnum::TypeCast(cast)) = node.node else { return Err(PgError::internal("a type name")) };
    let ty = crate::expr::resolve_type_name(&cast.type_name.ok_or_else(|| PgError::internal("a type name"))?)?;
    Ok(ColumnType { modifier: -1, ..ty })
}

/// go_spelling returns the SQL syntax Go writes for a type in the name of a table's anonymous type.
fn go_spelling(type_oid: u32) -> String {
    if let Some(t) = builtin_type(type_oid)
        && t.elem != 0
        && t.definition.typ_category == b"A"
    {
        return format!("{}[]", go_spelling(t.elem));
    }
    match type_oid {
        oid::INT8 => "BIGINT".into(),
        oid::INT2 => "SMALLINT".into(),
        oid::INT4 => "INTEGER".into(),
        oid::BOOL => "BOOL".into(),
        oid::NUMERIC => "DECIMAL".into(),
        oid::FLOAT4 => "REAL".into(),
        oid::FLOAT8 => "DOUBLE PRECISION".into(),
        oid::BPCHAR => "CHAR".into(),
        oid::CHAR => "\"char\"".into(),
        17 => "BYTES".into(),
        other => builtin_type(other).map_or("TEXT".into(), |t| t.name.to_uppercase()),
    }
}

/// table_type_id returns the ID of the anonymous type Go stores for a function returning a table of the columns.
fn table_type_id(columns: &[(String, ColumnType)]) -> Vec<u8> {
    let items: Vec<String> = columns
        .iter()
        .map(|(name, ty)| format!("{}:{}", crate::engine::quote_identifier(name), go_spelling(ty.oid)))
        .collect();
    id::new(crate::catalog::id::SECTION_TYPE, &["", &format!("{TABLE_PREFIX}{})", items.join(","))])
}

/// RowTypes finds the row type of the table that a stored type ID names, or None for no such table.
pub type RowTypes<'r> = dyn FnMut(&[u8]) -> Result<Option<objects::SerializedType>> + 'r;

/// all returns every function and procedure of a root value that this server can run.
fn all(db: &mut Database, root: &Root) -> Result<Vec<Arc<Routine>>> {
    let mut functions = Vec::new();
    for (_, address) in root.objects(db, COLLECTION)? {
        functions.push(Function::deserialize(&prolly::read_blob(db, &address)?)?);
    }
    let mut procedures = Vec::new();
    for (_, address) in root.objects(db, PROCEDURES)? {
        procedures.push(Procedure::deserialize(&prolly::read_blob(db, &address)?)?);
    }
    let mut row_type = |type_id: &[u8]| -> Result<Option<objects::SerializedType>> {
        let segments = id::segments(type_id);
        let [schema, name] = segments.as_slice() else { return Ok(None) };
        match root.table(db, schema, name)? {
            Some(address) => {
                let table = crate::catalog::table::TableDef::load(db, schema, name, address)?;
                Ok(Some(crate::usertypes::row_type(&table)))
            }
            None => Ok(None),
        }
    };
    let mut routines = Vec::new();
    for function in functions {
        if let Ok(routine) = Routine::load(function, &mut row_type) {
            routines.push(Arc::new(routine));
        }
    }
    for procedure in procedures {
        if let Ok(routine) = Routine::load_procedure(procedure, &mut row_type) {
            routines.push(Arc::new(routine));
        }
    }
    Ok(routines)
}

/// store writes a function, or a procedure, into a root value.
pub(crate) fn store(db: &mut Database, root: &mut Root, function: &Function, procedure: bool) -> Result<()> {
    let data = if procedure {
        Procedure {
            id: function.id.clone(),
            all_params: function.all_params.clone(),
            definition: function.definition.clone(),
            extension_name: function.extension_name.clone(),
            extension_symbol: function.extension_symbol.clone(),
            operations: function.operations.clone(),
            sql_definition: function.sql_definition.clone(),
        }
        .serialize()
    } else {
        function.serialize()
    };
    let mut sink = |_: Hash, bytes: &[u8]| {
        db.write_value(bytes.to_vec()).map(|_| ()).map_err(|e| store::Error::Corrupt(e.to_string()))
    };
    let (address, _) = prolly::write_blob(&data, &mut sink)?.ok_or_else(|| PgError::internal("an empty function"))?;
    root.put_object(db, if procedure { PROCEDURES } else { COLLECTION }, &function.id, Some(address))?;
    Ok(())
}

/// function_id returns the ID of a function, or a procedure, from its schema, name, and input types.
fn function_id(schema: &str, name: &str, inputs: &[ColumnType], procedure: bool) -> Vec<u8> {
    let types: Vec<String> = inputs.iter().map(|t| String::from_utf8_lossy(&type_id(*t)).into_owned()).collect();
    let mut segments = vec![schema, name];
    segments.extend(types.iter().map(String::as_str));
    id::new(if procedure { SECTION_PROCEDURE } else { SECTION_FUNCTION }, &segments)
}

/// invalid_definition returns Postgres' error for a function definition it rejects.
fn invalid_definition(message: impl Into<String>) -> PgError {
    PgError::new(code::INVALID_FUNCTION_DEFINITION, message)
}

/// Options are the options of CREATE FUNCTION.
#[derive(Default)]
struct Options {
    language: Option<String>,
    body: Option<String>,
    link: Option<(String, String)>,
    strict: bool,
}

/// options reads the options of CREATE FUNCTION, failing for one given twice.
fn options(nodes: &[Node]) -> Result<Options> {
    let mut options = Options::default();
    let mut seen = Vec::new();
    for node in nodes {
        let Some(NodeEnum::DefElem(def)) = node.node.as_ref() else { continue };
        if seen.contains(&def.defname) {
            return Err(PgError {
                position: position(def.location),
                ..PgError::new(code::SYNTAX_ERROR, "conflicting or redundant options")
            });
        }
        seen.push(def.defname.clone());
        let arg = def.arg.as_deref().and_then(|a| a.node.as_ref());
        match (def.defname.as_str(), arg) {
            ("language", Some(NodeEnum::String(s))) => options.language = Some(s.sval.to_lowercase()),
            ("as", Some(NodeEnum::List(list))) => {
                let items: Vec<String> = list.items.iter().filter_map(node_name).map(str::to_string).collect();
                match items.as_slice() {
                    [body] => options.body = Some(body.clone()),
                    [file, symbol] => options.link = Some((file.clone(), symbol.clone())),
                    _ => return Err(invalid_definition("only one AS item needed for language")),
                }
            }
            ("strict", Some(NodeEnum::Boolean(b))) => options.strict = b.boolval,
            _ => {}
        }
    }
    Ok(options)
}

/// function_names returns the schema, which is empty when unqualified, and name of a qualified function name.
fn function_names(nodes: &[Node]) -> (String, String) {
    let names: Vec<&str> = nodes.iter().filter_map(node_name).collect();
    match names.as_slice() {
        [name] => (String::new(), name.to_string()),
        [.., schema, name] => (schema.to_string(), name.to_string()),
        [] => (String::new(), String::new()),
    }
}

impl Ctx<'_> {
    /// routines returns every function and procedure of the working root, reusing them while their collections are
    /// unchanged.
    pub fn routines(&mut self) -> Result<Arc<Vec<Arc<Routine>>>> {
        let address = (
            self.txn.root.root_objects[COLLECTION],
            self.txn.root.root_objects[PROCEDURES],
            self.txn.root.root_objects[crate::usertypes::COLLECTION],
        );
        if let Some((cached, routines)) = &self.session.routines
            && *cached == address
        {
            let routines = routines.clone();
            for row_type in routines.iter().flat_map(|r| r.row_types.iter().cloned()) {
                crate::usertypes::register(row_type);
            }
            return Ok(routines);
        }
        let routines = Arc::new(all(self.db, &self.txn.root)?);
        for row_type in routines.iter().flat_map(|r| r.row_types.iter().cloned()) {
            crate::usertypes::register(row_type);
        }
        self.session.routines = Some((address, routines.clone()));
        Ok(routines)
    }

    /// routines_named returns the functions of the name in the schema, or in the search path's schemas in its order.
    pub fn routines_named(&mut self, schema: Option<&str>, name: &str) -> Result<Vec<Arc<Routine>>> {
        let all = self.routines()?;
        if all.iter().all(|r| r.name != name) {
            return Ok(Vec::new());
        }
        let schemas = match schema {
            Some(schema) => vec![schema.to_string()],
            None => self.session.search_path(),
        };
        let mut found = Vec::new();
        for schema in schemas {
            found.extend(all.iter().filter(|r| r.name == name && r.schema == schema).cloned());
        }
        Ok(found)
    }

    /// create_function runs CREATE FUNCTION, whose source text a PL/pgSQL body is compiled from.
    pub fn create_function(&mut self, stmt: &CreateFunctionStmt, text: &str) -> Result<Outcome> {
        let (named_schema, name) = function_names(&stmt.funcname);
        let schema = self.target_schema(&named_schema, -1)?;
        let options = options(&stmt.options)?;
        let mut params = Vec::new();
        let mut table = Vec::new();
        let mut seen_default = false;
        let mut seen_variadic = false;
        for node in &stmt.parameters {
            let Some(NodeEnum::FunctionParameter(param)) = node.node.as_ref() else { continue };
            if let Some(type_name) = &param.arg_type {
                self.prepare_type(type_name)?;
            }
            let ty = parameter_type(param)?;
            let mode = match FunctionParameterMode::try_from(param.mode) {
                Ok(FunctionParameterMode::FuncParamOut) => Mode::Out,
                Ok(FunctionParameterMode::FuncParamInout) => Mode::InOut,
                Ok(FunctionParameterMode::FuncParamVariadic) => Mode::Variadic,
                Ok(FunctionParameterMode::FuncParamTable) => {
                    table.push((param.name.clone(), ty));
                    continue;
                }
                _ => Mode::In,
            };
            if mode.is_input() && seen_variadic {
                return Err(invalid_definition("VARIADIC parameter must be the last input parameter"));
            }
            if mode == Mode::Variadic {
                if !crate::array::is_array_type(ty.oid) {
                    return Err(invalid_definition("VARIADIC parameter must be an array"));
                }
                seen_variadic = true;
            }
            let default = match &param.defexpr {
                Some(expr) => {
                    if !mode.is_input() {
                        return Err(invalid_definition("only input parameters can have default values"));
                    }
                    let mut binder = Binder::new(self, Scope::default());
                    binder.clause = "DEFAULT expressions";
                    binder.definition = true;
                    let bound = binder.bind(expr)?;
                    coerce(bound, ty, false, arg_location(expr))?;
                    seen_default = true;
                    Some(crate::ddl::expression_text(expr)?)
                }
                None if mode.is_input() && seen_default => {
                    return Err(invalid_definition(
                        "input parameters after one with a default value must also have defaults",
                    ));
                }
                None => None,
            };
            params.push(Param { name: param.name.clone(), ty, mode, default });
        }
        let outputs: Vec<&Param> = params.iter().filter(|p| p.mode.is_output()).collect();
        let mut row_type = None;
        if let Some(type_name) = &stmt.return_type {
            self.prepare_type(type_name)?;
        }
        let (mut ret, mut set_of) = match &stmt.return_type {
            Some(type_name) => match return_type(type_name) {
                Ok(ret) => (ret, type_name.setof),
                Err(err) => {
                    let names: Vec<&str> = type_name.names.iter().filter_map(node_name).collect();
                    let relation = pg_query::protobuf::RangeVar {
                        schemaname: if names.len() > 1 { names[names.len() - 2].to_string() } else { String::new() },
                        relname: names.last().map(|n| n.to_string()).unwrap_or_default(),
                        inh: true,
                        location: type_name.location,
                        ..Default::default()
                    };
                    let table = self.resolve_table(&relation).map_err(|_| err)?;
                    let columns: Vec<(String, ColumnType)> =
                        table.columns.iter().map(|c| (c.name.clone(), c.ty)).collect();
                    row_type =
                        Some((id::new(crate::catalog::id::SECTION_TYPE, &[&table.schema, &table.name]), columns));
                    (typ(oid::RECORD), type_name.setof)
                }
            },
            None if stmt.is_procedure => (typ(VOID), false),
            None => match outputs.as_slice() {
                [] => return Err(invalid_definition("function result type must be specified")),
                [only] => (only.ty, false),
                _ => (typ(oid::RECORD), false),
            },
        };
        if !table.is_empty() {
            ret = typ(oid::RECORD);
            set_of = true;
        } else if stmt.return_type.is_some() {
            let required = match outputs.as_slice() {
                [] => None,
                [only] => Some(only.ty),
                _ => Some(typ(oid::RECORD)),
            };
            if let Some(required) = required
                && required.oid != ret.oid
            {
                return Err(invalid_definition(format!(
                    "function result type must be {} because of OUT parameters",
                    type_display(required.oid)
                )));
            }
        }
        let language = match (&options.language, &stmt.sql_body) {
            (Some(language), _) => language.clone(),
            (None, Some(_)) => "sql".into(),
            (None, None) => return Err(invalid_definition("no language specified")),
        };
        let mut function = Function {
            return_type: match &row_type {
                Some((type_id, _)) => type_id.clone(),
                None if table.is_empty() => type_id(ret),
                None => table_type_id(&table),
            },
            all_params: params
                .iter()
                .map(|p| Parameter {
                    mode: p.mode.stored(),
                    name: p.name.clone().into_bytes(),
                    type_id: type_id(p.ty),
                    default: p.default.clone().unwrap_or_default().into_bytes(),
                })
                .collect(),
            is_non_deterministic: true,
            strict: options.strict,
            definition: text.as_bytes().to_vec(),
            set_of,
            ..Function::default()
        };
        match language.as_str() {
            "sql" => {
                let body = match (&options.body, &stmt.sql_body) {
                    (Some(body), _) => body.trim().trim_end_matches(';').trim_end().to_string(),
                    (None, Some(body)) => atomic_body(body)?,
                    (None, None) => return Err(invalid_definition("no function body specified")),
                };
                let offset = text.find(body.as_str()).unwrap_or(0) as u32;
                let composite = match crate::usertypes::get(ret.oid).map(|t| t.kind.clone()) {
                    Some(crate::usertypes::Kind::Composite(attributes)) => Some(attributes),
                    _ => None,
                };
                let columns = if let Some((_, columns)) = &row_type {
                    columns.clone()
                } else if let Some(attributes) = composite {
                    attributes
                } else if table.is_empty() && ret.oid == oid::RECORD && outputs.len() > 1 {
                    outputs.iter().map(|p| (p.name.clone(), p.ty)).collect()
                } else if table.len() > 1 {
                    table.clone()
                } else {
                    Vec::new()
                };
                let ret = if table.len() == 1 { table[0].1 } else { ret };
                self.check_sql_body(&name, &params, &body, ret, &columns)
                    .map_err(|err| PgError { position: err.position.map(|p| p + offset), ..err })?;
                function.sql_definition = body.into_bytes();
            }
            "plpgsql" => {
                let body = options.body.as_deref().ok_or_else(|| invalid_definition("no function body specified"))?;
                function.operations = crate::plpgsql::compile(self, text, body)?;
                let returns_value =
                    |op: &Operation| op.op_code == crate::plpgsql::OpCode::Return as u16 && !op.primary_data.is_empty();
                if stmt.is_procedure && function.operations.iter().any(returns_value) {
                    return Err(PgError {
                        position: crate::plpgsql::statement_position(text, body, "return", true),
                        ..PgError::new(code::SYNTAX_ERROR, "RETURN cannot have a parameter in a procedure")
                    });
                }
            }
            "c" | "internal" => return Err(PgError::unsupported(format!("LANGUAGE {language}"))),
            other => return Err(PgError::new(code::UNDEFINED_OBJECT, format!("language \"{other}\" does not exist"))),
        }
        let inputs: Vec<ColumnType> = params.iter().filter(|p| p.mode.is_input()).map(|p| p.ty).collect();
        function.id = function_id(&schema, &name, &inputs, stmt.is_procedure);
        let existing = self
            .routines()?
            .iter()
            .find(|r| {
                r.schema == schema && r.name == name && r.inputs().map(|p| p.ty.oid).eq(inputs.iter().map(|t| t.oid))
            })
            .cloned();
        if let Some(existing) = existing {
            if !stmt.replace {
                return Err(PgError::new(
                    code::DUPLICATE_FUNCTION,
                    format!("function \"{name}\" already exists with same argument types"),
                ));
            }
            if existing.procedure != stmt.is_procedure {
                return Err(PgError {
                    detail: Some(format!("\"{name}\" is a {}.", existing.kind())),
                    ..PgError::new(code::WRONG_OBJECT_TYPE, "cannot change routine kind")
                });
            }
            if !stmt.is_procedure && (existing.object.return_type != function.return_type || existing.set_of != set_of)
            {
                return Err(PgError {
                    hint: Some(format!("Use DROP FUNCTION {} first.", existing.signature())),
                    ..invalid_definition("cannot change return type of existing function")
                });
            }
            self.require_owner(&Object::Routine(schema.clone(), name.clone(), String::new()))?;
        }
        store(self.db, &mut self.txn.root, &function, stmt.is_procedure)?;
        self.own(Object::Routine(schema, name, String::new()))?;
        Ok(Outcome::command(if stmt.is_procedure { "CREATE PROCEDURE" } else { "CREATE FUNCTION" }))
    }

    /// check_sql_body plans a SQL body's statements up to the first utility statement, as Postgres does with
    /// check_function_bodies, and checks that the last one returns the function's result type.
    fn check_sql_body(
        &mut self,
        name: &str,
        params: &[Param],
        body: &str,
        ret: ColumnType,
        columns: &[(String, ColumnType)],
    ) -> Result<()> {
        let statements = parse_body(body)?;
        let inputs: Vec<&Param> = params.iter().filter(|p| p.mode.is_input()).collect();
        let names = (name.to_string(), inputs.iter().map(|p| p.name.clone()).collect::<Vec<_>>());
        let mut types: Vec<u32> = inputs.iter().map(|p| p.ty.oid).collect();
        let mismatch = |detail: String| PgError {
            detail: Some(detail),
            ..invalid_definition(format!(
                "return type mismatch in function declared to return {}",
                type_display(ret.oid)
            ))
        };
        let final_statement =
            || mismatch("Function's final statement must be SELECT or INSERT/UPDATE/DELETE RETURNING.".into());
        if ret.oid != VOID && statements.last().is_none_or(|s| !crate::engine::describable(s)) {
            return Err(final_statement());
        }
        let mut last = None;
        for statement in &statements {
            if !crate::engine::describable(statement) {
                return Ok(());
            }
            last = self.nested(&mut types, &[], Some(names.clone()), |ctx| ctx.describe(statement))?;
        }
        if ret.oid == VOID {
            return Ok(());
        }
        let Some(result) = last else { return Err(final_statement()) };
        if let [column] = result.as_slice()
            && column.type_oid == ret.oid
        {
            return Ok(());
        }
        if !columns.is_empty() {
            if result.len() > columns.len() {
                return Err(mismatch("Final statement returns too many columns.".into()));
            }
            if result.len() < columns.len() {
                return Err(mismatch("Final statement returns too few columns.".into()));
            }
            for (i, (column, (_, ty))) in result.iter().zip(columns).enumerate() {
                if column.type_oid != ty.oid && !assignable(column.type_oid, ty.oid) {
                    return Err(mismatch(format!(
                        "Final statement returns {} instead of {} at column {}.",
                        type_display(column.type_oid),
                        type_display(ty.oid),
                        i + 1
                    )));
                }
            }
            return Ok(());
        }
        if ret.oid == oid::RECORD {
            return Ok(());
        }
        let [column] = result.as_slice() else {
            return Err(mismatch("Final statement must return exactly one column.".into()));
        };
        if column.type_oid != ret.oid && !assignable(column.type_oid, ret.oid) {
            return Err(mismatch(format!("Actual return type is {}.", type_display(column.type_oid))));
        }
        Ok(())
    }

    /// nested runs a function with a context for a statement inside a routine, with its own parameters and the
    /// names a SQL body refers to them by.
    pub fn nested<T>(
        &mut self,
        parameters: &mut Vec<u32>,
        params: &[Value],
        named_params: Option<(String, Vec<String>)>,
        f: impl FnOnce(&mut Ctx<'_>) -> Result<T>,
    ) -> Result<T> {
        let mut ctx = Ctx {
            db: &mut *self.db,
            txn: &mut *self.txn,
            session: &mut *self.session,
            parameters,
            params,
            outer: Vec::new(),
            subquery_value: Value::Null,
            ctes: Vec::new(),
            work_tables: std::collections::HashMap::new(),
            named_params,
        };
        f(&mut ctx)
    }

    /// drop_routines runs DROP FUNCTION, DROP PROCEDURE, or DROP ROUTINE, whose kind is the procedure flag or None for
    /// either, resolving every routine first so that a missing one drops none.
    pub fn drop_routines(&mut self, drop: &DropStmt, procedures: Option<bool>) -> Result<Outcome> {
        let mut doomed = Vec::new();
        for object in &drop.objects {
            let Some(NodeEnum::ObjectWithArgs(target)) = object.node.as_ref() else { continue };
            if let Some(routine) = self.find_routine(target, drop.missing_ok, procedures)? {
                doomed.push(routine);
            }
        }
        for routine in doomed {
            self.require_owner(&Object::Routine(routine.schema.clone(), routine.name.clone(), String::new()))?;
            let dependents = self.trigger_dependents(&routine)?;
            if !dependents.is_empty() {
                let detail: Vec<String> =
                    dependents.iter().map(|d| format!("{d} depends on function {}", routine.signature())).collect();
                return Err(PgError {
                    detail: Some(detail.join("\n")),
                    hint: Some("Use DROP ... CASCADE to drop the dependent objects too.".into()),
                    ..PgError::new(
                        code::DEPENDENT_OBJECTS_STILL_EXIST,
                        format!("cannot drop function {} because other objects depend on it", routine.signature()),
                    )
                });
            }
            let collection = if routine.procedure { PROCEDURES } else { COLLECTION };
            self.txn.root.put_object(self.db, collection, &routine.object.id, None)?;
        }
        Ok(Outcome::command(match procedures {
            Some(true) => "DROP PROCEDURE",
            Some(false) => "DROP FUNCTION",
            None => "DROP ROUTINE",
        }))
    }

    /// find_routine returns the routine of the kind that DROP or ALTER names, with or without its argument types, or
    /// None when it is missing and that is allowed, after a notice.
    pub(crate) fn find_routine(
        &mut self,
        target: &ObjectWithArgs,
        missing_ok: bool,
        procedures: Option<bool>,
    ) -> Result<Option<Arc<Routine>>> {
        let (schema, name) = function_names(&target.objname);
        let schema = (!schema.is_empty()).then_some(schema);
        let all = self.routines_named(schema.as_deref(), &name)?;
        let kind = match procedures {
            Some(true) => "procedure",
            Some(false) => "function",
            None => "routine",
        };
        let shown = match &schema {
            Some(schema) => format!("{schema}.{name}"),
            None => name.clone(),
        };
        let wrong_kind = |routine: &Routine| {
            PgError::new(code::WRONG_OBJECT_TYPE, format!("{} is not a {kind}", routine.signature()))
        };
        if target.args_unspecified {
            let candidates: Vec<&Arc<Routine>> =
                all.iter().filter(|r| procedures.is_none_or(|p| r.procedure == p)).collect();
            return match candidates.as_slice() {
                [only] => Ok(Some((*only).clone())),
                [] if all.len() == 1 => Err(wrong_kind(&all[0])),
                [] if missing_ok => {
                    self.session.notice(PgError::notice("00000", format!("{kind} {shown}() does not exist, skipping")));
                    Ok(None)
                }
                [] => Err(PgError::new(code::UNDEFINED_FUNCTION, format!("could not find a {kind} named \"{shown}\""))),
                _ => Err(PgError {
                    hint: Some(format!("Specify the argument list to select the {kind} unambiguously.")),
                    ..PgError::new(code::AMBIGUOUS_FUNCTION, format!("{kind} name \"{shown}\" is not unique"))
                }),
            };
        }
        let mut types = Vec::new();
        for (i, arg) in target.objargs.iter().enumerate() {
            let Some(NodeEnum::TypeName(type_name)) = arg.node.as_ref() else { continue };
            let output = target.objfuncargs.get(i).and_then(|a| match a.node.as_ref() {
                Some(NodeEnum::FunctionParameter(p)) => Some(p.mode == FunctionParameterMode::FuncParamOut as i32),
                _ => None,
            });
            if output != Some(true) {
                self.prepare_type(type_name)?;
                types.push(crate::expr::resolve_type_name(type_name)?.oid);
            }
        }
        let found = all.into_iter().find(|r| r.inputs().map(|p| p.ty.oid).eq(types.iter().copied()));
        match found {
            Some(routine) if procedures.is_some_and(|p| routine.procedure != p) => Err(wrong_kind(&routine)),
            Some(routine) => Ok(Some(routine)),
            None if missing_ok => {
                let qualified: Vec<String> = types
                    .iter()
                    .map(|&t| {
                        builtin_type(t).map_or(type_display(t).into_owned(), |b| format!("pg_catalog.{}", b.name))
                    })
                    .collect();
                self.session.notice(PgError::notice(
                    "00000",
                    format!("{kind} {shown}({}) does not exist, skipping", qualified.join(",")),
                ));
                Ok(None)
            }
            None => {
                let shown_types: Vec<_> = types.iter().map(|&t| type_display(t)).collect();
                Err(PgError::new(
                    code::UNDEFINED_FUNCTION,
                    format!("{kind} {shown}({}) does not exist", shown_types.join(", ")),
                ))
            }
        }
    }
}

impl Ctx<'_> {
    /// bind_call chooses the procedure a CALL runs and binds its arguments.
    fn bind_call(&mut self, stmt: &pg_query::protobuf::CallStmt) -> Result<(Arc<Routine>, Vec<Expr>)> {
        let invocation = stmt.funccall.as_ref().ok_or_else(|| PgError::internal("CALL without a call"))?;
        let names: Vec<&str> = invocation.funcname.iter().filter_map(node_name).collect();
        let (schema, name) = match names.as_slice() {
            [name] => (None, *name),
            [.., schema, name] => (Some(*schema), *name),
            [] => return Err(PgError::internal("CALL without a name")),
        };
        let routines = self.routines_named(schema, name)?;
        let bound = Binder::new(self, Scope::default()).routine_call(invocation, schema, name, routines, true)?;
        match bound {
            Some((Expr::Routine(routine, args), _)) => Ok((routine, args)),
            _ => Err(PgError::internal("CALL of something other than a procedure")),
        }
    }

    /// call_columns returns the columns of the row a CALL returns, which a procedure without output parameters lacks.
    pub fn call_columns(&mut self, stmt: &pg_query::protobuf::CallStmt) -> Result<Option<Vec<crate::Column>>> {
        let (routine, _) = self.bind_call(stmt)?;
        Ok((!routine.columns.is_empty())
            .then(|| routine.columns.iter().map(|(name, ty)| crate::query::column(name.clone(), *ty)).collect()))
    }

    /// call_procedure runs CALL, returning the values of the procedure's output parameters as a row when it has any.
    pub fn call_procedure(&mut self, stmt: &pg_query::protobuf::CallStmt) -> Result<Outcome> {
        let (routine, args) = self.bind_call(stmt)?;
        let values = args.iter().map(|a| a.eval(self, &[])).collect::<Result<Vec<_>>>()?;
        let result = call(self, &routine, values)?;
        if routine.columns.is_empty() {
            return Ok(Outcome::command("CALL"));
        }
        let row = match result {
            Value::Record(fields) => fields,
            _ => vec![Value::Null; routine.columns.len()],
        };
        let columns = routine.columns.iter().map(|(name, ty)| crate::query::column(name.clone(), *ty)).collect();
        Ok(Outcome::Rows { columns, rows: vec![row], tag: "CALL".into() })
    }

    /// do_block runs DO, which runs an anonymous PL/pgSQL block compiled from the statement's text.
    pub fn do_block(&mut self, stmt: &pg_query::protobuf::DoStmt, text: &str) -> Result<Outcome> {
        let mut language = "plpgsql".to_string();
        let mut body = String::new();
        for arg in &stmt.args {
            let Some(NodeEnum::DefElem(def)) = arg.node.as_ref() else { continue };
            let Some(NodeEnum::String(value)) = def.arg.as_deref().and_then(|a| a.node.as_ref()) else { continue };
            match def.defname.as_str() {
                "language" => language = value.sval.to_lowercase(),
                "as" => body = value.sval.clone(),
                _ => {}
            }
        }
        match language.as_str() {
            "plpgsql" => {}
            "sql" | "c" | "internal" => {
                return Err(PgError::new(
                    code::FEATURE_NOT_SUPPORTED,
                    format!("language \"{language}\" does not support inline code execution"),
                ));
            }
            other => return Err(PgError::new(code::UNDEFINED_OBJECT, format!("language \"{other}\" does not exist"))),
        }
        let operations = crate::plpgsql::compile(self, text, &body)?;
        let routine = Routine::internal(Function { operations, ..Function::default() }, typ(VOID));
        call(self, &routine, Vec::new())?;
        Ok(Outcome::command("DO"))
    }
}

/// parameter_type resolves a parameter's type, which Postgres stores without its modifier.
fn parameter_type(param: &FunctionParameter) -> Result<ColumnType> {
    let type_name = param.arg_type.as_ref().ok_or_else(|| PgError::internal("a parameter without a type"))?;
    let ty = crate::expr::resolve_type_name(type_name)?;
    Ok(ColumnType { modifier: -1, ..ty })
}

/// return_type resolves a function's declared result type, where `trigger` and `void` are pseudo-types.
fn return_type(type_name: &TypeName) -> Result<ColumnType> {
    let names: Vec<&str> = type_name.names.iter().filter_map(node_name).collect();
    match names.as_slice() {
        [name] | ["pg_catalog", name] if *name == "void" => Ok(typ(VOID)),
        [name] | ["pg_catalog", name] if *name == "trigger" => Ok(typ(TRIGGER)),
        [name] | ["pg_catalog", name] if *name == "record" => Ok(typ(oid::RECORD)),
        _ => Ok(ColumnType { modifier: -1, ..crate::expr::resolve_type_name(type_name)? }),
    }
}

/// atomic_body returns the statements of a `BEGIN ATOMIC ... END` or `RETURN` body as text, separated by semicolons.
fn atomic_body(body: &Node) -> Result<String> {
    let statements: Vec<Node> = match body.node.as_ref() {
        Some(NodeEnum::List(list)) => match list.items.first().and_then(|n| n.node.as_ref()) {
            Some(NodeEnum::List(inner)) => inner.items.clone(),
            _ => list.items.clone(),
        },
        Some(NodeEnum::ReturnStmt(ret)) => {
            let value = ret.returnval.as_deref().ok_or_else(|| PgError::internal("RETURN without a value"))?;
            return Ok(format!("RETURN {}", crate::ddl::expression_text(value)?));
        }
        _ => vec![body.clone()],
    };
    let mut texts = Vec::with_capacity(statements.len());
    for statement in &statements {
        let node = statement.node.as_ref().ok_or_else(|| PgError::internal("an empty statement"))?;
        texts.push(match node {
            NodeEnum::ReturnStmt(ret) => {
                let value = ret.returnval.as_deref().ok_or_else(|| PgError::internal("RETURN without a value"))?;
                format!("SELECT {}", crate::ddl::expression_text(value)?)
            }
            _ => node.to_ref().deparse().map_err(PgError::internal)?,
        });
    }
    Ok(texts.join(";"))
}

/// Candidate is an overload that a call may choose, with the parameter types its arguments convert to.
#[derive(Clone)]
enum Candidate {
    Builtin,
    /// A user routine, with the input parameter each argument goes to.
    User(Arc<Routine>, Vec<usize>),
}

impl Binder<'_, '_> {
    /// routine_call binds a call of a name that user routines have, choosing as Postgres does among them, or among the
    /// procedures for a CALL, and the built-in overloads, and returns None when a built-in function wins.
    pub fn routine_call(
        &mut self,
        call: &pg_query::protobuf::FuncCall,
        schema: Option<&str>,
        name: &str,
        routines: Vec<Arc<Routine>>,
        procedures: bool,
    ) -> Result<Option<Bound>> {
        let mut bound = Vec::with_capacity(call.args.len());
        let mut names = Vec::with_capacity(call.args.len());
        for arg in &call.args {
            match arg.node.as_ref() {
                Some(NodeEnum::NamedArgExpr(named)) => {
                    let value = named.arg.as_deref().ok_or_else(|| PgError::internal("a named argument"))?;
                    bound.push(self.bind(value)?);
                    names.push(Some(named.name.clone()));
                }
                _ => {
                    bound.push(self.bind(arg)?);
                    names.push(None);
                }
            }
        }
        let types: Vec<u32> = bound.iter().map(|(_, t)| t.oid).collect();
        let mut candidates: Vec<(Candidate, Vec<u32>)> = Vec::new();
        if !procedures && names.iter().all(Option::is_none) && schema.is_none_or(|s| s == "pg_catalog") {
            candidates
                .extend(functions::overload_types(name, types.len()).into_iter().map(|p| (Candidate::Builtin, p)));
        }
        for routine in routines.iter().filter(|r| r.procedure == procedures) {
            let params = call_params(routine);
            if let Some(slots) = argument_slots(&params, &names) {
                let types = slots.iter().map(|&s| params[s].ty.oid).collect();
                candidates.push((Candidate::User(routine.clone(), slots), types));
            }
        }
        let shown = || {
            let qualified = schema.map_or(name.to_string(), |s| format!("{s}.{name}"));
            let args: Vec<String> = bound
                .iter()
                .zip(&names)
                .map(|((_, t), n)| match n {
                    Some(n) => format!("{n} => {}", type_display(t.oid)),
                    None => type_display(t.oid).into_owned(),
                })
                .collect();
            format!("{qualified}({})", args.join(", "))
        };
        let mut best = functions::best_candidates(&types, candidates);
        let mut seen: Vec<(String, Vec<u32>)> = Vec::new();
        best.retain(|(candidate, params)| {
            let schema = match candidate {
                Candidate::Builtin => "pg_catalog",
                Candidate::User(routine, _) => routine.schema.as_str(),
            };
            let hidden = seen.iter().any(|(s, p)| s != schema && p == params);
            seen.push((schema.to_string(), params.clone()));
            !hidden
        });
        let kind = if procedures { "procedure" } else { "function" };
        let other_kind = routines.iter().any(|r| r.procedure != procedures && takes(r, &names, &types))
            || (procedures
                && schema.is_none_or(|s| s == "pg_catalog")
                && functions::exists(name)
                && functions::resolve(name, &types, call.location).is_ok());
        match best.as_slice() {
            [] if other_kind => Err(PgError {
                position: position(call.location),
                hint: Some(
                    if procedures { "To call a function, use SELECT." } else { "To call a procedure, use CALL." }
                        .into(),
                ),
                ..PgError::new(
                    code::WRONG_OBJECT_TYPE,
                    format!("{} {}", shown(), if procedures { "is not a procedure" } else { "is a procedure" }),
                )
            }),
            [] => Err(PgError {
                position: position(call.location),
                hint: Some(format!(
                    "No {kind} matches the given name and argument types. You might need to add explicit type casts."
                )),
                ..PgError::new(code::UNDEFINED_FUNCTION, format!("{kind} {} does not exist", shown()))
            }),
            [(Candidate::Builtin, _)] => Ok(None),
            [(Candidate::User(routine, slots), _)] => {
                let routine = routine.clone();
                let slots = slots.clone();
                Ok(Some(self.bind_routine_call(call, routine, bound, &slots)?))
            }
            _ => Err(PgError {
                position: position(call.location),
                hint: Some(format!(
                    "Could not choose a best candidate {kind}. You might need to add explicit type casts."
                )),
                ..PgError::new(code::AMBIGUOUS_FUNCTION, format!("{kind} {} is not unique", shown()))
            }),
        }
    }

    /// bind_routine_call converts a chosen routine's arguments to its parameter types, filling in the defaults of the
    /// parameters the call leaves out, and leaving out the placeholders a CALL passes for output parameters.
    fn bind_routine_call(
        &mut self,
        call: &pg_query::protobuf::FuncCall,
        routine: Arc<Routine>,
        bound: Vec<Bound>,
        slots: &[usize],
    ) -> Result<Bound> {
        let params: Vec<Param> = call_params(&routine).into_iter().cloned().collect();
        let mut args: Vec<Option<Expr>> = vec![None; params.len()];
        for ((bound, &slot), node) in bound.into_iter().zip(slots).zip(&call.args) {
            if let Expr::Param(i) = bound.0
                && self.ctx.parameters[i] == 0
            {
                self.ctx.parameters[i] = params[slot].ty.oid;
            }
            args[slot] = Some(coerce(bound, params[slot].ty, false, arg_location(node))?.0);
        }
        let mut exprs = Vec::with_capacity(params.len());
        for (arg, param) in args.into_iter().zip(&params) {
            if !param.mode.is_input() {
                continue;
            }
            exprs.push(match arg {
                Some(expr) => expr,
                None => {
                    let text = param.default.as_deref().ok_or_else(|| PgError::internal("a missing argument"))?;
                    let node = crate::parse::expression_node(text)?;
                    let bound = Binder::new(self.ctx, Scope::default()).bind(&node)?;
                    coerce(bound, param.ty, false, -1)?.0
                }
            });
        }
        let ret = routine.ret;
        let set_of = routine.set_of;
        let call_expr = Expr::Routine(routine, exprs);
        if set_of {
            let Some(set_functions) = self.set_functions.as_mut() else {
                return Err(PgError {
                    position: position(call.location),
                    ..PgError::new(
                        code::FEATURE_NOT_SUPPORTED,
                        format!("set-returning functions are not allowed in {}", self.clause),
                    )
                });
            };
            set_functions.push(call_expr);
            return Ok((Expr::SetRef(set_functions.len() - 1), ret));
        }
        Ok((call_expr, ret))
    }
}

/// argument_slots returns the input parameter that each argument of a call goes to, with the positional arguments
/// first and then the named ones, or None when the routine cannot take the arguments: a name matches no parameter or
/// names one a positional argument fills, or a parameter left out has no default.
fn argument_slots(inputs: &[&Param], names: &[Option<String>]) -> Option<Vec<usize>> {
    let mut slots = Vec::with_capacity(names.len());
    for (i, name) in names.iter().enumerate() {
        let slot = match name {
            None if i < inputs.len() => i,
            None => return None,
            Some(name) => inputs.iter().position(|p| p.name == *name)?,
        };
        if slots.contains(&slot) {
            return None;
        }
        slots.push(slot);
    }
    let filled = (0..inputs.len()).all(|i| slots.contains(&i) || inputs[i].default.is_some());
    filled.then_some(slots)
}

/// call_params returns the parameters that a call's arguments go to: every parameter of a procedure, whose output
/// parameters CALL passes placeholders for, and the input parameters of a function.
fn call_params(routine: &Routine) -> Vec<&Param> {
    if routine.procedure { routine.params.iter().collect() } else { routine.inputs().collect() }
}

/// takes reports whether a routine's input parameters can take arguments of the types with the names.
fn takes(routine: &Routine, names: &[Option<String>], types: &[u32]) -> bool {
    let params: Vec<&Param> = routine.inputs().collect();
    argument_slots(&params, names).is_some_and(|slots| {
        slots.iter().zip(types).all(|(&slot, &t)| functions::implicitly_castable(t, params[slot].ty.oid))
    })
}

/// call runs a routine on arguments already converted to its input types.
pub fn call(ctx: &mut Ctx<'_>, routine: &Routine, args: Vec<Value>) -> Result<Value> {
    if routine.strict && args.iter().any(Value::is_null) {
        return Ok(if routine.set_of { Value::Set(Vec::new()) } else { Value::Null });
    }
    if ctx.session.call_depth >= MAX_DEPTH {
        return Err(PgError {
            hint: Some(
                "Increase the configuration parameter \"max_stack_depth\" (currently 2048kB), after ensuring the \
                 platform's stack depth limit is adequate."
                    .into(),
            ),
            ..PgError::new(code::STATEMENT_TOO_COMPLEX, "stack depth limit exceeded")
        });
    }
    ctx.session.call_depth += 1;
    let result = match &routine.body {
        Body::Sql(_) => run_sql(ctx, routine, &args),
        Body::PlPgSql(operations) => crate::plpgsql::call(ctx, routine, operations, args),
        Body::External => {
            let (extension, symbol) = (&routine.object.extension_name, &routine.object.extension_symbol);
            match crate::extensions::implementation(
                &String::from_utf8_lossy(extension),
                &String::from_utf8_lossy(symbol),
            ) {
                Some(implementation) => implementation(ctx, &args, routine.ret),
                None => Err(PgError::unsupported(format!("the function {}", routine.signature()))),
            }
        }
    };
    ctx.session.call_depth -= 1;
    match result {
        Ok(Value::Null)
            if routine.ret.oid == VOID
                && routine.columns.is_empty()
                && !routine.procedure
                && matches!(routine.body, Body::PlPgSql(_)) =>
        {
            Ok(Value::Text(String::new()))
        }
        Err(err) => Err(PgError { position: None, ..err }),
        result => result,
    }
}

/// run_sql runs a SQL body's statements in order, with the arguments as their parameters, and returns what the last
/// one returns.
fn run_sql(ctx: &mut Ctx<'_>, routine: &Routine, args: &[Value]) -> Result<Value> {
    let statements = routine.sql_statements()?;
    let names = (routine.name.clone(), routine.inputs().map(|p| p.name.clone()).collect::<Vec<_>>());
    let mut types: Vec<u32> = routine.inputs().map(|p| p.ty.oid).collect();
    let mut last = None;
    for statement in statements {
        last = Some(ctx.nested(&mut types, args, Some(names.clone()), |ctx| ctx.run(statement))?);
    }
    let rows = match last {
        Some(Outcome::Rows { rows, .. }) => rows,
        _ => Vec::new(),
    };
    result_value(routine, rows)
}

/// result_value turns the rows a routine's body produced into its result: all of them for a set-returning routine
/// and otherwise the first, as a record for a routine returning several columns.
pub fn result_value(routine: &Routine, rows: Vec<Vec<Value>>) -> Result<Value> {
    if routine.ret.oid == VOID && routine.columns.is_empty() {
        return Ok(Value::Null);
    }
    let mut values = Vec::with_capacity(rows.len());
    for row in rows {
        values.push(row_value(routine, row)?);
        if !routine.set_of {
            break;
        }
    }
    if routine.set_of {
        return Ok(Value::Set(values));
    }
    Ok(values.pop().unwrap_or(Value::Null))
}

/// row_value converts one row a routine produced to its result type.
fn row_value(routine: &Routine, row: Vec<Value>) -> Result<Value> {
    let composite = routine.ret.oid != oid::RECORD && !routine.procedure && crate::expr::is_composite(routine.ret.oid);
    if routine.columns.len() > 1 || composite || routine.procedure && !routine.columns.is_empty() {
        let row = match row.as_slice() {
            [Value::Composite(_) | Value::Record(_)] if routine.columns.len() > 1 => match row.into_iter().next() {
                Some(Value::Composite(c)) => c.fields,
                Some(Value::Record(fields)) => fields,
                _ => Vec::new(),
            },
            _ => row,
        };
        let mut fields = Vec::with_capacity(row.len());
        for (value, (_, ty)) in row.into_iter().zip(&routine.columns) {
            fields.push(crate::cast::cast_value(value, *ty, false)?);
        }
        if composite {
            return Ok(Value::Composite(Box::new(crate::types::CompositeValue { type_oid: routine.ret.oid, fields })));
        }
        return Ok(Value::Record(fields));
    }
    if routine.ret.oid == oid::RECORD {
        return Ok(Value::Record(row));
    }
    let value = row.into_iter().next().unwrap_or(Value::Null);
    crate::cast::cast_value(value, routine.ret, false)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn table_types_round_trip() {
        let columns = vec![("x".to_string(), typ(oid::INT4)), ("y".to_string(), typ(oid::TEXT))];
        let id = table_type_id(&columns);
        assert_eq!(table_columns(&id).unwrap(), Some(columns));
    }
}
