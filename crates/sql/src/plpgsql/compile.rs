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

//! Compiling a PL/pgSQL body into Go's interpreter operations: the parser's JSON becomes statements, and the
//! statements become a flat list of operations whose jumps are resolved against their labels.

use std::collections::{BTreeMap, HashMap, HashSet};

use objects::Operation;
use pg_query::protobuf::Token;
use serde_json::Value as Json;

use super::{
    CONTINUE_TARGET, FOUND, OPTION_CONTEXT, OPTION_DYNAMIC_BINDING, OPTION_DYNAMIC_BINDING_COUNT,
    OPTION_DYNAMIC_EXPRESSION, OPTION_DYNAMIC_USING_BINDING, OPTION_DYNAMIC_USING_BINDING_COUNT,
    OPTION_DYNAMIC_USING_COUNT, OPTION_DYNAMIC_USING_EXPRESSION, OPTION_LOOP_CONDITION, OPTION_RETYPE_TARGET,
    OPTION_SETS_FOUND, OPTION_STRICT, OpCode, TRIGGER_VARIABLES, normalize_identifier, normalize_path,
    quote_identifier,
};
use crate::error::{PgError, Result, code};

/// Statement is a PL/pgSQL statement on its way to becoming operations.
#[derive(Clone, Debug)]
enum Statement {
    Assignment {
        variable: String,
        expression: String,
        retype: bool,
    },
    Block(Block),
    ExecuteSql {
        statement: String,
        target: String,
        record: bool,
        sets_found: bool,
        strict: bool,
    },
    DynamicExecute {
        query: String,
        params: Vec<String>,
        target: String,
        record: bool,
    },
    ForQueryInit {
        query: String,
        dynamic: Option<Vec<String>>,
    },
    Unsupported {
        name: String,
    },
    ReturnNext {
        expression: String,
    },
    ForQueryNext {
        record: String,
        offset: i32,
    },
    Goto {
        offset: i32,
        label: String,
        nearest: bool,
    },
    If {
        condition: String,
        offset: i32,
        loop_condition: bool,
    },
    Perform {
        statement: String,
    },
    Raise {
        level: String,
        message: String,
        params: Vec<String>,
        options: BTreeMap<String, String>,
    },
    ReturnQuery {
        query: String,
    },
    Return {
        expression: String,
    },
    /// A statement with where it is in the body, such as `line 3 at RAISE`, which errors it raises report.
    At {
        context: String,
        statement: Box<Statement>,
    },
}

/// Block is a scope of statements with the variables and records it declares.
#[derive(Clone, Debug, Default)]
struct Block {
    variables: Vec<Variable>,
    records: Vec<Record>,
    body: Vec<Statement>,
    label: String,
    is_loop: bool,
    /// The offset from the body's first operation to where a CONTINUE of the loop jumps.
    continue_target: i32,
}

/// Variable is a declared variable or a parameter.
#[derive(Clone, Debug, Default)]
struct Variable {
    name: String,
    type_name: String,
    is_parameter: bool,
    default: String,
    datum: i32,
}

/// Record is a declared record, or a trigger's NEW or OLD, with the fields the body refers to.
#[derive(Clone, Debug, Default)]
struct Record {
    name: String,
    fields: Vec<String>,
    default: String,
    datum: i32,
    trigger: bool,
}

impl Record {
    /// is_declared reports whether entering the block declares the record, which the trigger records and the
    /// parser's unnamed internal records are not.
    fn is_declared(&self) -> bool {
        !self.trigger && !self.name.is_empty()
    }
}

impl Statement {
    /// size returns how many operations the statement becomes.
    fn size(&self) -> i32 {
        match self {
            Statement::Block(block) => {
                2 + block.variables.iter().filter(|v| !v.is_parameter).count() as i32
                    + block.records.iter().filter(|r| r.is_declared()).count() as i32
                    + size(&block.body)
            }
            Statement::At { statement, .. } => statement.size(),
            _ => 1,
        }
    }
}

/// dynamic_operation returns an operation that runs the query text an expression evaluates to with the values of the
/// USING expressions, as EXECUTE and FOR ... IN EXECUTE do.
fn dynamic_operation(code: OpCode, query: &str, params: &[String], names: &Names) -> Result<Operation> {
    let (query, bindings) = substitute(query, names)?;
    let mut items = vec![
        (OPTION_DYNAMIC_EXPRESSION.to_string(), "true".to_string()),
        (OPTION_DYNAMIC_BINDING_COUNT.to_string(), bindings.len().to_string()),
    ];
    for (i, binding) in bindings.into_iter().enumerate() {
        items.push((format!("{OPTION_DYNAMIC_BINDING}{i}"), binding));
    }
    items.push((OPTION_DYNAMIC_USING_COUNT.to_string(), params.len().to_string()));
    for (i, param) in params.iter().enumerate() {
        let (expression, bindings) = substitute(param, names)?;
        items.push((format!("{OPTION_DYNAMIC_USING_EXPRESSION}{i}"), expression));
        items.push((format!("{OPTION_DYNAMIC_USING_BINDING_COUNT}{i}"), bindings.len().to_string()));
        for (j, binding) in bindings.into_iter().enumerate() {
            items.push((format!("{OPTION_DYNAMIC_USING_BINDING}{i}_{j}"), binding));
        }
    }
    let mut operation = op(code, query);
    operation.options = items.into_iter().map(|(k, v)| (k.into_bytes(), v.into_bytes())).collect();
    Ok(operation)
}

/// size returns how many operations the statements become.
fn size(statements: &[Statement]) -> i32 {
    statements.iter().map(Statement::size).sum()
}

/// Scope is a compile-time scope: the names it declares and its label.
#[derive(Default)]
struct Scope {
    names: HashSet<String>,
    label: String,
}

/// Names tracks the names in scope and the labels while operations are appended.
struct Names {
    scopes: Vec<Scope>,
    label_id: usize,
    aliases: Aliases,
}

/// Aliases maps each name that `ALIAS FOR` declares to the name it stands for, which compilation uses in its place.
#[derive(Default)]
struct Aliases(HashMap<String, String>);

impl Aliases {
    /// find returns the aliases a body declares.
    fn find(body: &str) -> Aliases {
        let Ok(scan) = pg_query::scan(body) else { return Aliases::default() };
        let tokens = scan.tokens;
        let piece = |i: usize| tokens.get(i).map_or("", |t| &body[t.start as usize..t.end as usize]);
        let mut aliases = HashMap::new();
        for i in 0..tokens.len() {
            if piece(i + 1).eq_ignore_ascii_case("alias") && piece(i + 2).eq_ignore_ascii_case("for") {
                aliases.insert(normalize_identifier(piece(i)), normalize_identifier(piece(i + 3)));
            }
        }
        Aliases(aliases)
    }

    /// with_parameters returns the aliases a body declares, with `$N` standing for the routine's Nth parameter, as
    /// Postgres names every parameter.
    fn with_parameters(body: &str, variables: &[Variable]) -> Aliases {
        let mut aliases = Aliases::find(body);
        for (i, variable) in variables.iter().filter(|v| v.is_parameter && v.name != FOUND).enumerate() {
            aliases.0.entry(format!("${}", i + 1)).or_insert_with(|| variable.name.clone());
        }
        aliases
    }

    /// resolve returns the name a name stands for, following aliases of aliases.
    fn resolve(&self, name: &str) -> String {
        let mut name = name.to_string();
        for _ in 0..self.0.len() {
            match self.0.get(&name) {
                Some(target) => name = target.clone(),
                None => break,
            }
        }
        name
    }

    /// resolve_path resolves the variable of a `name` or `name.field` reference.
    fn resolve_path(&self, path: &str) -> String {
        match path.split_once('.') {
            Some((base, field)) => format!("{}.{field}", self.resolve(base)),
            None => self.resolve(path),
        }
    }
}

impl Names {
    /// current_label returns the innermost label.
    fn current_label(&self) -> String {
        self.scopes.iter().rev().map(|s| s.label.clone()).find(|l| !l.is_empty()).unwrap_or_default()
    }

    /// contains reports whether a name is declared in any scope.
    fn contains(&self, name: &str) -> bool {
        self.scopes.iter().any(|s| s.names.contains(name))
    }

    /// declare adds a name to the innermost scope.
    fn declare(&mut self, name: &str) {
        if let Some(scope) = self.scopes.last_mut() {
            scope.names.insert(name.to_string());
        }
    }
}

/// op returns an operation with the code and primary data.
fn op(code: OpCode, primary: impl Into<String>) -> Operation {
    Operation { op_code: code as u16, primary_data: primary.into().into_bytes(), ..Operation::default() }
}

/// strings converts strings to the byte strings an operation stores.
fn strings(items: Vec<String>) -> Vec<Vec<u8>> {
    items.into_iter().map(String::into_bytes).collect()
}

/// options converts options to the byte strings an operation stores.
fn options(items: &[(&str, String)]) -> BTreeMap<Vec<u8>, Vec<u8>> {
    items.iter().map(|(k, v)| (k.as_bytes().to_vec(), v.clone().into_bytes())).collect()
}

impl Statement {
    /// append adds the statement's operations.
    fn append(&self, ops: &mut Vec<Operation>, names: &mut Names) -> Result<()> {
        match self {
            Statement::Assignment { variable, expression, retype } => {
                let (expression, bindings) = substitute(expression, names)?;
                let mut operation = op(OpCode::Assign, format!("SELECT {expression};"));
                operation.secondary_data = strings(bindings);
                operation.target = variable.clone().into_bytes();
                if *retype {
                    operation.options = options(&[(OPTION_RETYPE_TARGET, "true".into())]);
                }
                ops.push(operation);
            }
            Statement::Block(block) => block.append(ops, names)?,
            Statement::At { context, statement } => {
                let start = ops.len();
                statement.append(ops, names)?;
                for operation in &mut ops[start..] {
                    operation
                        .options
                        .entry(OPTION_CONTEXT.as_bytes().to_vec())
                        .or_insert_with(|| context.clone().into_bytes());
                }
            }
            Statement::ExecuteSql { statement, target, record, sets_found, strict } => {
                let (statement, bindings) = substitute(statement, names)?;
                let mut operation = op(if *record { OpCode::ExecuteInto } else { OpCode::Execute }, statement);
                operation.secondary_data = strings(bindings);
                operation.target = target.clone().into_bytes();
                if *sets_found {
                    operation.options.insert(OPTION_SETS_FOUND.as_bytes().to_vec(), b"true".to_vec());
                }
                if *strict {
                    operation.options.insert(OPTION_STRICT.as_bytes().to_vec(), b"true".to_vec());
                }
                ops.push(operation);
            }
            Statement::DynamicExecute { query, params, target, record } => {
                let mut operation = dynamic_operation(
                    if *record { OpCode::ExecuteInto } else { OpCode::Execute },
                    query,
                    params,
                    names,
                )?;
                operation.target = target.clone().into_bytes();
                ops.push(operation);
            }
            Statement::ForQueryInit { query, dynamic: Some(params) } => {
                ops.push(dynamic_operation(OpCode::ForQueryInit, query, params, names)?);
            }
            Statement::ForQueryInit { query, dynamic: None } => {
                let (query, bindings) = substitute(query, names)?;
                let mut operation = op(OpCode::ForQueryInit, query);
                operation.secondary_data = strings(bindings);
                ops.push(operation);
            }
            Statement::Unsupported { name } => ops.push(op(OpCode::Unsupported, name)),
            Statement::ForQueryNext { record, offset } => {
                let mut operation = op(OpCode::ForQueryNext, "");
                operation.target = record.clone().into_bytes();
                operation.index = ops.len() as i32 + offset;
                ops.push(operation);
            }
            Statement::Goto { offset, label, nearest } => {
                let mut operation = op(OpCode::Goto, "");
                if !label.is_empty() {
                    operation.primary_data = label.clone().into_bytes();
                    operation.index = *offset;
                } else if *nearest {
                    let label = names.current_label();
                    if label.is_empty() {
                        return Err(PgError::new(
                            code::SYNTAX_ERROR,
                            if *offset > 0 {
                                "EXIT cannot be used outside a loop, unless it has a label"
                            } else {
                                "CONTINUE cannot be used outside a loop"
                            },
                        ));
                    }
                    operation.primary_data = label.into_bytes();
                    operation.index = *offset;
                } else {
                    operation.index = ops.len() as i32 + offset;
                }
                ops.push(operation);
            }
            Statement::If { condition, offset, loop_condition } => {
                let (condition, bindings) = substitute(condition, names)?;
                let mut operation = op(OpCode::If, format!("SELECT {condition};"));
                operation.secondary_data = strings(bindings);
                operation.index = ops.len() as i32 + offset;
                if *loop_condition {
                    operation.options = options(&[(OPTION_LOOP_CONDITION, "true".into())]);
                }
                ops.push(operation);
            }
            Statement::Perform { statement } => {
                let (statement, bindings) = substitute(statement, names)?;
                let mut operation = op(OpCode::Perform, statement);
                operation.secondary_data = strings(bindings);
                ops.push(operation);
            }
            Statement::Raise { level, message, params, options } => {
                let mut operation = op(OpCode::Raise, level.clone());
                operation.secondary_data = strings(std::iter::once(message.clone()).chain(params.clone()).collect());
                operation.options =
                    options.iter().map(|(k, v)| (k.clone().into_bytes(), v.clone().into_bytes())).collect();
                ops.push(operation);
            }
            Statement::ReturnQuery { query } => {
                let (query, bindings) = substitute(query, names)?;
                let mut operation = op(OpCode::ReturnQuery, query);
                operation.secondary_data = strings(bindings);
                ops.push(operation);
            }
            Statement::Return { expression } | Statement::ReturnNext { expression } => {
                let (expression, bindings) = substitute(expression, names)?;
                let expression = if expression.is_empty() { expression } else { format!("SELECT {expression};") };
                let code = if matches!(self, Statement::Return { .. }) { OpCode::Return } else { OpCode::ReturnNext };
                let mut operation = op(code, expression);
                operation.secondary_data = strings(bindings);
                ops.push(operation);
            }
        }
        Ok(())
    }
}

/// Declaration is a record or variable that a block declares.
enum Declaration<'b> {
    Record(&'b Record),
    Variable(&'b Variable),
}

impl Block {
    /// append adds the block's operations: its scope, its declarations in the order written, and its body.
    fn append(&self, ops: &mut Vec<Operation>, names: &mut Names) -> Result<()> {
        names.scopes.push(Scope { label: self.label.clone(), ..Scope::default() });
        let mut label = self.label.clone();
        if self.is_loop && label.is_empty() {
            label = format!("\t{}", names.label_id);
            names.label_id += 1;
            if let Some(scope) = names.scopes.last_mut() {
                scope.label = label.clone();
            }
        }
        let begin = ops.len();
        let mut scope_begin = op(OpCode::ScopeBegin, label);
        if self.is_loop {
            scope_begin.target = b"_".to_vec();
        }
        ops.push(scope_begin);
        for record in &self.records {
            if !record.is_declared() {
                names.declare(&record.name);
            }
        }
        let mut declarations: Vec<(i32, Declaration<'_>)> =
            self.records.iter().filter(|r| r.is_declared()).map(|r| (r.datum, Declaration::Record(r))).collect();
        declarations.extend(self.variables.iter().map(|v| (v.datum, Declaration::Variable(v))));
        declarations.sort_by_key(|(datum, _)| *datum);
        for (_, declaration) in declarations {
            match declaration {
                Declaration::Record(record) => {
                    names.declare(&record.name);
                    let mut operation = op(OpCode::DeclareRecord, "");
                    operation.target = record.name.clone().into_bytes();
                    if !record.default.is_empty() {
                        let (expression, bindings) = substitute(&record.default, names)?;
                        operation.secondary_data = strings(
                            [record.default.clone(), format!("SELECT {expression};")]
                                .into_iter()
                                .chain(bindings)
                                .collect(),
                        );
                    }
                    ops.push(operation);
                }
                Declaration::Variable(variable) => {
                    let mut operation = op(OpCode::Declare, variable.type_name.clone());
                    operation.target = variable.name.clone().into_bytes();
                    if !variable.default.is_empty() {
                        let (expression, bindings) = substitute(&variable.default, names)?;
                        operation.secondary_data = strings(
                            [variable.default.clone(), format!("SELECT {expression};")]
                                .into_iter()
                                .chain(bindings)
                                .collect(),
                        );
                    }
                    if !variable.is_parameter {
                        ops.push(operation);
                    }
                    names.declare(&variable.name);
                }
            }
        }
        if self.is_loop {
            let target = ops.len() as i32 + self.continue_target;
            ops[begin].options = options(&[(CONTINUE_TARGET, (target - begin as i32).to_string())]);
        }
        for statement in &self.body {
            statement.append(ops, names)?;
        }
        ops.push(op(OpCode::ScopeEnd, ""));
        names.scopes.pop();
        Ok(())
    }
}

/// substitute replaces each variable an expression names, or field of one, with `$N`, returning the expression, with
/// the text between its tokens kept, and the names the parameters bind, where a name before `(` is a function
/// and the columns that an INSERT lists or an UPDATE sets stay column names.
fn substitute(expression: &str, names: &Names) -> Result<(String, Vec<String>)> {
    let tokens = pg_query::scan(expression).map_err(|err| PgError::new(code::SYNTAX_ERROR, err.to_string()))?.tokens;
    let targets = assignment_targets(expression);
    let text = |i: usize| &expression[tokens[i].start as usize..tokens[i].end as usize];
    let is = |i: usize, token: Token| tokens.get(i).is_some_and(|t| t.token == token as i32);
    let mut out = String::new();
    let mut bindings = Vec::new();
    let mut written = 0;
    let mut i = 0;
    while i < tokens.len() {
        let start = tokens[i].start as usize;
        out.push_str(&expression[written..start]);
        let after_dot = i > 0 && is(i - 1, Token::Ascii46);
        let normalized = names.aliases.resolve(&normalize_identifier(text(i)));
        if !after_dot && !targets.contains(&tokens[i].start) && names.contains(&normalized) {
            let mut binding = normalized;
            while i + 2 < tokens.len() && is(i + 1, Token::Ascii46) {
                binding = format!("{binding}.{}", text(i + 2));
                i += 2;
            }
            let end = tokens[i].end as usize;
            if is(i + 1, Token::Ascii40) {
                out.push_str(&expression[start..end]);
            } else {
                bindings.push(binding);
                out.push_str(&format!("${}", bindings.len()));
            }
        } else if !after_dot && TRIGGER_VARIABLES.iter().any(|(name, _)| *name == normalized) {
            bindings.push(normalized);
            out.push_str(&format!("${}", bindings.len()));
        } else {
            out.push_str(text(i));
        }
        written = tokens[i].end as usize;
        i += 1;
    }
    out.push_str(&expression[written..]);
    Ok((out, bindings))
}

/// assignment_targets returns where the column names that an INSERT lists or an UPDATE sets start, which name the
/// table's columns rather than variables, as PL/pgSQL's parser hooks leave them.
fn assignment_targets(text: &str) -> std::collections::HashSet<i32> {
    let mut targets = std::collections::HashSet::new();
    let Ok(parsed) = pg_query::parse(text, 0) else { return targets };
    let names = |list: &[pg_query::Node]| -> Vec<i32> {
        list.iter()
            .filter_map(|n| match n.node.as_ref() {
                Some(pg_query::NodeEnum::ResTarget(t)) => Some(t.location),
                _ => None,
            })
            .collect()
    };
    for statement in parsed.protobuf.stmts.iter().filter_map(|s| s.stmt.as_ref()).filter_map(|s| s.node.as_ref()) {
        match statement {
            pg_query::NodeEnum::InsertStmt(insert) => {
                targets.extend(names(&insert.cols));
                if let Some(clause) = insert.on_conflict_clause.as_ref() {
                    targets.extend(names(&clause.target_list));
                }
            }
            pg_query::NodeEnum::UpdateStmt(update) => targets.extend(names(&update.target_list)),
            _ => {}
        }
    }
    targets
}

/// get returns a field of a JSON object.
fn get<'j>(json: &'j Json, key: &str) -> Option<&'j Json> {
    json.as_object().and_then(|o| o.get(key))
}

/// text returns a JSON string field, or an empty string.
fn text(json: &Json, key: &str) -> String {
    get(json, key).and_then(Json::as_str).unwrap_or_default().to_string()
}

/// int returns a JSON integer field, or zero.
fn int(json: &Json, key: &str) -> i32 {
    get(json, key).and_then(Json::as_i64).unwrap_or(0) as i32
}

/// flag returns a JSON boolean field, or false.
fn flag(json: &Json, key: &str) -> bool {
    get(json, key).and_then(Json::as_bool).unwrap_or(false)
}

/// list returns a JSON array field, or an empty slice.
fn list<'j>(json: &'j Json, key: &str) -> &'j [Json] {
    get(json, key).and_then(Json::as_array).map_or(&[], Vec::as_slice)
}

/// query returns the text of the expression in a field, which wraps it as `{"PLpgSQL_expr": {"query": ...}}`.
fn query(json: &Json, key: &str) -> String {
    get(json, key).and_then(|e| get(e, "PLpgSQL_expr")).map(|e| text(e, "query")).unwrap_or_default()
}

/// variant returns the one key of a union object, such as a statement or datum, with its value.
fn variant(json: &Json) -> Option<(&str, &Json)> {
    json.as_object().and_then(|o| o.iter().next()).map(|(k, v)| (k.as_str(), v))
}

/// Datums names each of a function's datums by its number, and knows the body's aliases.
struct Datums(Vec<String>, Aliases);

impl Datums {
    /// name returns the name of a datum by its number.
    fn name(&self, number: i32) -> Result<String> {
        match self.0.get(number as usize) {
            Some(name) if !name.is_empty() => Ok(name.clone()),
            _ => Err(PgError::internal(format!("PL/pgSQL datum {number} does not name a declared variable"))),
        }
    }
}

/// compile_json turns the parser's JSON for one function into Go's operations.
pub fn compile_json(function: &Json, body: &str) -> Result<Vec<Operation>> {
    let mut block = convert_function(function, body)?;
    for variable in &mut block.variables {
        if !variable.is_parameter || variable.name == FOUND {
            variable.type_name = super::declared_type_name(&variable.type_name);
        }
    }
    let mut ops = Vec::new();
    let aliases = Aliases::with_parameters(body, &block.variables);
    let mut names = Names { scopes: vec![Scope::default()], label_id: 0, aliases };
    Statement::Block(block).append(&mut ops, &mut names)?;
    reconcile_labels(&mut ops)?;
    Ok(ops)
}

/// convert_function converts a function's datums and body into its top-level block, where a record datum without a
/// line is a parameter of a composite type when it has no number or comes before the FOUND variable.
fn convert_function(function: &Json, body: &str) -> Result<Block> {
    let datums = list(function, "datums");
    let action = get(function, "action").and_then(|a| get(a, "PLpgSQL_stmt_block")).cloned().unwrap_or(Json::Null);
    let new_number = int(function, "new_varno");
    let old_number = int(function, "old_varno");
    let mut block = Block { label: text(&action, "label"), ..Block::default() };
    let lowest = datums.iter().filter_map(|d| get(d, "PLpgSQL_rec")).map(|r| int(r, "dno")).min().unwrap_or(i32::MAX);
    let offset = 0i32.wrapping_sub(lowest);
    let parameters = datums
        .iter()
        .position(|d| get(d, "PLpgSQL_var").is_some_and(|v| get(v, "lineno").is_none() && text(v, "refname") == FOUND))
        .unwrap_or(datums.len());
    let mut found = None;
    for (index, datum) in datums.iter().enumerate() {
        let Some((kind, value)) = variant(datum) else { continue };
        match kind {
            "PLpgSQL_rec" => {
                let number = (int(value, "dno") + offset) as usize;
                if number >= block.records.len() {
                    block.records.resize(number + 1, Record::default());
                }
                let parameter = get(value, "lineno").is_none()
                    && ![new_number, old_number].iter().any(|&n| n != 0 && n == int(value, "dno"))
                    && (int(value, "dno") <= 0 || index < parameters);
                if parameter {
                    block.variables.push(Variable {
                        name: text(value, "refname"),
                        type_name: String::new(),
                        is_parameter: true,
                        default: String::new(),
                        datum: index as i32,
                    });
                } else if int(value, "dno") > 0 {
                    block.records[number].name = text(value, "refname");
                    block.records[number].default = query(value, "default_val");
                    block.records[number].datum = index as i32;
                }
            }
            "PLpgSQL_recfield" => {
                let parent = (int(value, "recparentno") + offset) as usize;
                let record =
                    block.records.get_mut(parent).ok_or_else(|| PgError::internal("invalid record parent number"))?;
                record.fields.push(text(value, "fieldname"));
            }
            "PLpgSQL_row" | "PLpgSQL_promise" => {}
            "PLpgSQL_var" => {
                let name = text(value, "refname");
                let line = int(value, "lineno");
                if line == 0 && name.eq_ignore_ascii_case(FOUND) {
                    found = Some(block.variables.len());
                }
                let type_name = get(value, "datatype")
                    .and_then(|d| get(d, "PLpgSQL_type"))
                    .map(|t| text(t, "typname").to_lowercase())
                    .unwrap_or_default();
                block.variables.push(Variable {
                    name,
                    type_name,
                    is_parameter: line == 0,
                    default: query(value, "default_val"),
                    datum: index as i32,
                });
            }
            other => return Err(PgError::unsupported(format!("the PL/pgSQL declaration {other}"))),
        }
    }
    if let Some(found) = found {
        let variable = &mut block.variables[found];
        variable.is_parameter = false;
        variable.default = "false".into();
        variable.type_name = "pg_catalog.bool".into();
    }
    for number in [new_number, old_number] {
        if number == 0 {
            continue;
        }
        if let Some(record) = usize::try_from(number.wrapping_add(offset)).ok().and_then(|i| block.records.get_mut(i)) {
            record.trigger = true;
        }
    }
    let names = Datums(
        datums
            .iter()
            .enumerate()
            .map(|(i, datum)| match variant(datum) {
                Some(("PLpgSQL_rec" | "PLpgSQL_var" | "PLpgSQL_promise", v)) => text(v, "refname"),
                Some(("PLpgSQL_recfield", v)) => {
                    let parent = int(v, "recparentno");
                    match datums.get(parent as usize).and_then(variant) {
                        Some((_, p)) if (parent as usize) < i => {
                            format!("{}.{}", text(p, "refname"), text(v, "fieldname"))
                        }
                        _ => String::new(),
                    }
                }
                _ => String::new(),
            })
            .collect(),
        Aliases::with_parameters(body, &block.variables),
    );
    block.body = convert_statements(list(&action, "body"), &names)?;
    Ok(block)
}

/// convert_statements converts a list of statements.
fn convert_statements(statements: &[Json], datums: &Datums) -> Result<Vec<Statement>> {
    statements
        .iter()
        .map(|s| {
            let statement = Box::new(convert_statement(s, datums)?);
            Ok(match variant(s) {
                Some((kind, fields)) => Statement::At {
                    context: format!("line {} at {}", int(fields, "lineno"), kind_name(kind, fields)),
                    statement,
                },
                None => *statement,
            })
        })
        .collect()
}

/// kind_name returns the name that Postgres' error context gives a kind of PL/pgSQL statement.
fn kind_name(kind: &str, statement: &Json) -> &'static str {
    match kind.trim_start_matches("PLpgSQL_stmt_") {
        "block" => "statement block",
        "assign" => "assignment",
        "if" => "IF",
        "case" => "CASE",
        "loop" => "LOOP",
        "while" => "WHILE",
        "fori" => "FOR with integer loop variable",
        "fors" => "FOR over SELECT rows",
        "forc" => "FOR over cursor",
        "foreach_a" => "FOREACH over array",
        "exit" if flag(statement, "is_exit") => "EXIT",
        "exit" => "CONTINUE",
        "return" => "RETURN",
        "return_next" => "RETURN NEXT",
        "return_query" => "RETURN QUERY",
        "raise" => "RAISE",
        "assert" => "ASSERT",
        "execsql" => "SQL statement",
        "dynexecute" => "EXECUTE",
        "dynfors" => "FOR over EXECUTE statement",
        "getdiag" => "GET DIAGNOSTICS",
        "open" => "OPEN",
        "fetch" => "FETCH",
        "close" => "CLOSE",
        "perform" => "PERFORM",
        "call" if flag(statement, "is_call") => "CALL",
        "call" => "DO",
        "commit" => "COMMIT",
        "rollback" => "ROLLBACK",
        _ => "statement",
    }
}

/// into_target returns what an INTO clause writes: a record by name, or the comma-separated names of a row's
/// variables or a single variable.
fn into_target(target: &Json) -> Result<(String, bool)> {
    match variant(target) {
        Some(("PLpgSQL_row", row)) => {
            let fields: Vec<String> = list(row, "fields").iter().map(|f| text(f, "name")).collect();
            Ok((fields.join(","), false))
        }
        Some(("PLpgSQL_var", var)) => Ok((text(var, "refname"), false)),
        Some(("PLpgSQL_rec", rec)) => Ok((text(rec, "refname"), true)),
        _ => Err(PgError::internal("unhandled INTO target: expected a record, row, or variable")),
    }
}

/// is_data_modifying reports whether a query is an INSERT, UPDATE, DELETE, or MERGE, the statements that set FOUND
/// without an INTO clause.
fn is_data_modifying(query: &str) -> bool {
    use pg_query::NodeEnum;
    pg_query::parse(query, 0).is_ok_and(|result| {
        result.protobuf.stmts.iter().any(|raw| {
            matches!(
                raw.stmt.as_ref().and_then(|s| s.node.as_ref()),
                Some(
                    NodeEnum::InsertStmt(_)
                        | NodeEnum::UpdateStmt(_)
                        | NodeEnum::DeleteStmt(_)
                        | NodeEnum::MergeStmt(_)
                )
            )
        })
    })
}

/// normalized_params folds the arguments of RAISE and USING that are plain references, which arrive as source text.
fn normalized_params(params: &[Json]) -> Vec<String> {
    params
        .iter()
        .map(|p| {
            let text = get(p, "PLpgSQL_expr").map(|e| text(e, "query")).unwrap_or_default();
            normalize_path(&text).unwrap_or(text)
        })
        .collect()
}

/// NOTICE_LEVELS names RAISE's levels by the numbers the parser gives them.
const NOTICE_LEVELS: [(i32, &str); 6] =
    [(14, "DEBUG"), (15, "LOG"), (17, "INFO"), (18, "NOTICE"), (19, "WARNING"), (21, "EXCEPTION")];

/// convert_statement converts one statement.
fn convert_statement(statement: &Json, datums: &Datums) -> Result<Statement> {
    let Some((kind, s)) = variant(statement) else { return Err(PgError::internal("an empty PL/pgSQL statement")) };
    Ok(match kind {
        "PLpgSQL_stmt_assign" => {
            let query = query(s, "expr");
            let (variable, expression) = match query.find(":=") {
                Some(i) if i > 0 => (query[..i].trim(), query[i + 2..].trim()),
                _ => match query.find('=') {
                    Some(i) if i > 0 => (query[..i].trim(), query[i + 1..].trim()),
                    _ => return Err(PgError::internal("PL/pgSQL assignment cannot find `:=` sign")),
                },
            };
            Statement::Assignment {
                variable: datums.1.resolve_path(&normalize_identifier(variable)),
                expression: expression.to_string(),
                retype: false,
            }
        }
        "PLpgSQL_stmt_block" => {
            Statement::Block(Block { body: convert_statements(list(s, "body"), datums)?, ..Block::default() })
        }
        "PLpgSQL_stmt_call" => {
            let target = get(s, "target").cloned().unwrap_or(Json::Null);
            let (target, record) = if !flag(s, "is_call") && variant(&target).is_some() {
                into_target(&target)?
            } else {
                (String::new(), false)
            };
            Statement::ExecuteSql { statement: query(s, "expr"), target, record, sets_found: false, strict: false }
        }
        "PLpgSQL_stmt_case" => convert_case(s, datums)?,
        "PLpgSQL_stmt_dynexecute" => {
            let (target, record) = if flag(s, "into") {
                into_target(get(s, "target").unwrap_or(&Json::Null))?
            } else {
                (String::new(), false)
            };
            Statement::DynamicExecute {
                query: query(s, "query"),
                params: normalized_params(list(s, "params")),
                target,
                record,
            }
        }
        "PLpgSQL_stmt_execsql" => {
            let into = flag(s, "into");
            let (target, record) =
                if into { into_target(get(s, "target").unwrap_or(&Json::Null))? } else { (String::new(), false) };
            let statement = query(s, "sqlstmt");
            let sets_found = into || is_data_modifying(&statement);
            Statement::ExecuteSql { statement, target, record, sets_found, strict: into && flag(s, "strict") }
        }
        "PLpgSQL_stmt_exit" => {
            let offset = if flag(s, "is_exit") { 1 } else { -1 };
            let label = text(s, "label");
            let jump = Statement::Goto { offset, nearest: label.is_empty(), label };
            match get(s, "cond") {
                Some(_) => Statement::Block(Block {
                    body: vec![
                        Statement::If { condition: query(s, "cond"), offset: 2, loop_condition: false },
                        Statement::Goto { offset: 2, label: String::new(), nearest: false },
                        jump,
                    ],
                    ..Block::default()
                }),
                None => jump,
            }
        }
        "PLpgSQL_stmt_foreach_a" => convert_foreach(s, datums)?,
        "PLpgSQL_stmt_fori" => convert_fori(s, datums)?,
        "PLpgSQL_stmt_fors" | "PLpgSQL_stmt_dynfors" => {
            let var = get(s, "var").unwrap_or(&Json::Null);
            let name = match variant(var) {
                Some(("PLpgSQL_rec" | "PLpgSQL_var", v)) => text(v, "refname"),
                Some(("PLpgSQL_row", _)) => into_target(var)?.0,
                _ => return Err(PgError::internal("FOR..IN..SELECT loop variable must be a record, row, or variable")),
            };
            let body = convert_statements(list(s, "body"), datums)?;
            let n = size(&body);
            let dynamic = (kind == "PLpgSQL_stmt_dynfors").then(|| normalized_params(list(s, "params")));
            let mut statements = vec![
                Statement::ForQueryInit { query: query(s, "query"), dynamic },
                Statement::ForQueryNext { record: name, offset: n + 2 },
            ];
            statements.extend(body);
            statements.push(Statement::Goto { offset: -(1 + n), label: String::new(), nearest: false });
            Statement::Block(Block {
                label: text(s, "label"),
                is_loop: true,
                continue_target: 1,
                body: statements,
                ..Block::default()
            })
        }
        "PLpgSQL_stmt_if" => convert_if(s, datums)?,
        "PLpgSQL_stmt_loop" => {
            let mut body = convert_statements(list(s, "body"), datums)?;
            let n = size(&body);
            body.push(Statement::Goto { offset: -n, label: String::new(), nearest: false });
            Statement::Block(Block { label: text(s, "label"), is_loop: true, body, ..Block::default() })
        }
        "PLpgSQL_stmt_perform" => Statement::Perform { statement: query(s, "expr") },
        "PLpgSQL_stmt_raise" => {
            let level = int(s, "elog_level");
            let options = list(s, "options")
                .iter()
                .filter_map(|o| get(o, "PLpgSQL_raise_option"))
                .map(|o| (int(o, "opt_type").to_string(), query(o, "expr")))
                .collect();
            Statement::Raise {
                level: NOTICE_LEVELS.iter().find(|(n, _)| *n == level).map_or("UNKNOWN", |(_, l)| l).to_string(),
                message: text(s, "message"),
                params: normalized_params(list(s, "params")),
                options,
            }
        }
        "PLpgSQL_stmt_return" => Statement::Return { expression: returned_expression(s, datums)? },
        "PLpgSQL_stmt_return_query" => Statement::ReturnQuery { query: query(s, "query") },
        "PLpgSQL_stmt_return_next" => Statement::ReturnNext { expression: returned_expression(s, datums)? },
        "PLpgSQL_stmt_while" => {
            let body = convert_statements(list(s, "body"), datums)?;
            let n = size(&body);
            let mut statements = vec![
                Statement::If { condition: query(s, "cond"), offset: 2, loop_condition: false },
                Statement::Goto { offset: 1 + n + 1, label: String::new(), nearest: false },
            ];
            statements.extend(body);
            statements.push(Statement::Goto { offset: -(n + 2), label: String::new(), nearest: false });
            Statement::Block(Block { label: text(s, "label"), is_loop: true, body: statements, ..Block::default() })
        }
        other => {
            let name = other.strip_prefix("PLpgSQL_stmt_").unwrap_or(other).replace('_', " ").to_uppercase();
            Statement::Unsupported { name }
        }
    })
}

/// returned_expression returns what RETURN or RETURN NEXT returns: its expression, or the variable that the parser
/// gives by number, which is empty for the row of OUT parameters that a bare RETURN returns.
fn returned_expression(s: &Json, datums: &Datums) -> Result<String> {
    if get(s, "expr").is_some() {
        return Ok(query(s, "expr"));
    }
    match get(s, "retvarno").and_then(Json::as_i64).and_then(|n| usize::try_from(n).ok()) {
        Some(number) => match datums.0.get(number) {
            Some(name) if name.is_empty() => Ok(String::new()),
            Some(name) => Ok(quote_identifier(name)),
            None => Err(PgError::internal("invalid PL/pgSQL datum number")),
        },
        None => Ok(String::new()),
    }
}

/// convert_case converts CASE into an assignment of its expression, when it has one, and a chain of conditions that
/// fails as Postgres does when no branch matches and there is no ELSE.
fn convert_case(s: &Json, datums: &Datums) -> Result<Statement> {
    let mut body = Vec::new();
    let expression = query(s, "t_expr");
    if !expression.is_empty() {
        body.push(Statement::Assignment {
            variable: format!("__Case__Variable_{}__", int(s, "t_varno")),
            expression,
            retype: true,
        });
    }
    let mut ends = Vec::new();
    for when in list(s, "case_when_list") {
        let Some(when) = get(when, "PLpgSQL_case_when") else {
            return Err(PgError::internal("case statement WHEN clause is nil"));
        };
        let statements = convert_statements(list(when, "stmts"), datums)?;
        body.push(Statement::If { condition: query(when, "expr"), offset: 2, loop_condition: false });
        body.push(Statement::Goto { offset: size(&statements) + 2, label: String::new(), nearest: false });
        body.extend(statements);
        ends.push((body.len(), size(&body)));
        body.push(Statement::Goto { offset: 0, label: String::new(), nearest: false });
    }
    if flag(s, "have_else") {
        body.extend(convert_statements(list(s, "else_stmts"), datums)?);
    } else {
        body.push(Statement::Raise {
            level: "EXCEPTION".into(),
            message: "case not found".into(),
            params: Vec::new(),
            options: [
                ("0".to_string(), format!("'{}'", code::CASE_NOT_FOUND)),
                ("3".to_string(), "'CASE statement is missing ELSE part.'".to_string()),
            ]
            .into_iter()
            .collect(),
        });
    }
    let total = size(&body);
    for (index, position) in ends {
        body[index] = Statement::Goto { offset: total - position, label: String::new(), nearest: false };
    }
    Ok(Statement::Block(Block { body, ..Block::default() }))
}

/// convert_if converts IF, ELSIF, and ELSE into conditions and jumps.
fn convert_if(s: &Json, datums: &Datums) -> Result<Statement> {
    let mut body = Vec::new();
    let mut ends = Vec::new();
    let mut branch = |body: &mut Vec<Statement>, condition: String, statements: Vec<Statement>| {
        body.push(Statement::If { condition, offset: 2, loop_condition: false });
        body.push(Statement::Goto { offset: size(&statements) + 2, label: String::new(), nearest: false });
        body.extend(statements);
        ends.push((body.len(), size(body)));
        body.push(Statement::Goto { offset: 0, label: String::new(), nearest: false });
    };
    branch(&mut body, query(s, "cond"), convert_statements(list(s, "then_body"), datums)?);
    for elsif in list(s, "elsif_list") {
        let elsif = get(elsif, "PLpgSQL_if_elsif").unwrap_or(&Json::Null);
        branch(&mut body, query(elsif, "cond"), convert_statements(list(elsif, "stmts"), datums)?);
    }
    body.extend(convert_statements(list(s, "else_body"), datums)?);
    let total = size(&body);
    for (index, position) in ends {
        body[index] = Statement::Goto { offset: total - position, label: String::new(), nearest: false };
    }
    Ok(Statement::Block(Block { body, ..Block::default() }))
}

/// convert_fori converts an integer FOR loop.
fn convert_fori(s: &Json, datums: &Datums) -> Result<Statement> {
    let var =
        get(s, "var").and_then(|v| get(v, "PLpgSQL_var")).ok_or_else(|| PgError::internal("for loop variable"))?;
    let name = text(var, "refname");
    let quoted = quote_identifier(&name);
    let bound = |key: &str| {
        let q = query(s, key);
        if q.is_empty() { "1".to_string() } else { q }
    };
    let (lower, upper, step) = (bound("lower"), bound("upper"), bound("step"));
    let (condition, increment) = if flag(s, "reverse") {
        (format!("{quoted} >= ({upper})"), format!("{quoted} - ({step})"))
    } else {
        (format!("{quoted} <= ({upper})"), format!("{quoted} + ({step})"))
    };
    let body = convert_statements(list(s, "body"), datums)?;
    let n = size(&body);
    let mut statements = vec![
        Statement::Assignment { variable: name.clone(), expression: lower, retype: false },
        Statement::Goto { offset: 2, label: String::new(), nearest: false },
        Statement::Assignment { variable: name, expression: increment, retype: false },
        Statement::If { condition, offset: 2, loop_condition: true },
        Statement::Goto { offset: 2 + n, label: String::new(), nearest: false },
    ];
    statements.extend(body);
    statements.push(Statement::Goto { offset: -(3 + n), label: String::new(), nearest: false });
    Ok(Statement::Block(Block {
        label: text(s, "label"),
        is_loop: true,
        continue_target: 2,
        body: statements,
        ..Block::default()
    }))
}

/// The fields of the records a FOREACH loop works through.
const FOREACH_ARRAY: &str = "__foreach_array__";
const FOREACH_TYPE: &str = "__foreach_type__";
const FOREACH_ELEMENT: &str = "__foreach_element__";

/// convert_foreach converts FOREACH over an array, which evaluates the array once into a record, checks it, and then
/// walks its elements, or its slices, as a query's rows.
fn convert_foreach(s: &Json, datums: &Datums) -> Result<Statement> {
    let name = datums.name(int(s, "varno"))?;
    let line = int(s, "lineno");
    let slice = int(s, "slice");
    let source = format!("__foreach_src_{line}__");
    let row = format!("__foreach_row_{line}__");
    let array = format!("{source}.{FOREACH_ARRAY}");
    let type_ref = format!("{source}.{FOREACH_TYPE}");
    let evaluate = format!(
        "SELECT __foreach_value__ AS {FOREACH_ARRAY}, CAST(pg_typeof(__foreach_value__) AS TEXT) AS {FOREACH_TYPE} \
         FROM (SELECT ({}) AS __foreach_value__) AS __foreach_input__",
        query(s, "expr")
    );
    let iterate = if slice > 0 {
        format!(
            "SELECT __foreach_source__ AS {FOREACH_ELEMENT} FROM __doltgres_foreach_slice({array},{slice}) AS __foreach_source__"
        )
    } else {
        format!("SELECT __foreach_source__ AS {FOREACH_ELEMENT} FROM unnest({array}) AS __foreach_source__")
    };
    let body = convert_statements(list(s, "body"), datums)?;
    let n = size(&body);
    let raise = |message: &str, params: Vec<String>, state: &str| Statement::Raise {
        level: "EXCEPTION".into(),
        message: message.into(),
        params,
        options: [("0".to_string(), state.to_string())].into_iter().collect(),
    };
    let mut statements = vec![
        Statement::ExecuteSql {
            statement: evaluate,
            target: source.clone(),
            record: true,
            sets_found: false,
            strict: false,
        },
        Statement::If { condition: format!("{array} IS NOT NULL"), offset: 2, loop_condition: false },
        raise("FOREACH expression must not be null", Vec::new(), code::NULL_VALUE_NOT_ALLOWED),
        Statement::If { condition: format!("{type_ref} LIKE '%[]'"), offset: 2, loop_condition: false },
        raise("FOREACH expression must yield an array, not type %", vec![type_ref.clone()], code::DATATYPE_MISMATCH),
        Statement::ForQueryInit { query: iterate, dynamic: None },
        Statement::ForQueryNext { record: row.clone(), offset: n + 3 },
        Statement::Assignment { variable: name.clone(), expression: format!("{row}.{FOREACH_ELEMENT}"), retype: false },
    ];
    let mut continue_target = 6;
    if slice > 0 {
        let checks = vec![
            Statement::If {
                condition: format!("CAST(pg_typeof({}) AS TEXT) LIKE '%[]'", quote_identifier(&name)),
                offset: 2,
                loop_condition: false,
            },
            raise("FOREACH ... SLICE loop variable must be of an array type", Vec::new(), code::DATATYPE_MISMATCH),
        ];
        continue_target += checks.len() as i32;
        statements.splice(5..5, checks);
    }
    statements.extend(body);
    statements.push(Statement::Goto { offset: -(2 + n), label: String::new(), nearest: false });
    Ok(Statement::Block(Block {
        label: text(s, "label"),
        is_loop: true,
        continue_target,
        records: vec![
            Record { name: source, fields: vec![FOREACH_ARRAY.into(), FOREACH_TYPE.into()], ..Record::default() },
            Record { name: row, fields: vec![FOREACH_ELEMENT.into()], ..Record::default() },
        ],
        body: statements,
        ..Block::default()
    }))
}

/// reconcile_labels resolves the jumps that name a label, which EXIT and CONTINUE make, against the scopes enclosing
/// them, and clears the labels and loop markers that only resolution needs.
fn reconcile_labels(ops: &mut [Operation]) -> Result<()> {
    let undefined = |label: &[u8]| {
        PgError::new(
            code::SYNTAX_ERROR,
            format!(
                "there is no label \"{}\" attached to any block or loop enclosing this statement",
                String::from_utf8_lossy(label)
            ),
        )
    };
    let mut labels: Vec<(Vec<u8>, i32, bool)> = Vec::new();
    let mut exits: Vec<usize> = Vec::new();
    for index in 0..ops.len() {
        match ops[index].op_code {
            c if c == OpCode::Goto as u16 && !ops[index].primary_data.is_empty() => {
                if ops[index].index < 0 {
                    let label = ops[index].primary_data.clone();
                    let (_, start, is_loop) =
                        labels.iter().rev().find(|(l, ..)| *l == label).cloned().ok_or_else(|| undefined(&label))?;
                    if !is_loop {
                        return Err(PgError::new(code::SYNTAX_ERROR, "CONTINUE cannot be used outside a loop"));
                    }
                    ops[index].index = start;
                    ops[index].primary_data.clear();
                } else {
                    exits.push(index);
                }
            }
            c if c == OpCode::ScopeBegin as u16 => {
                let start = match ops[index].options.get(CONTINUE_TARGET.as_bytes()) {
                    Some(offset) => {
                        let offset: i32 = String::from_utf8_lossy(offset)
                            .parse()
                            .map_err(|_| PgError::internal(format!("invalid CONTINUE target for scope at {index}")))?;
                        index as i32 + offset
                    }
                    None => index as i32 + 1,
                };
                labels.push((ops[index].primary_data.clone(), start, !ops[index].target.is_empty()));
                ops[index].primary_data.clear();
                ops[index].target.clear();
                ops[index].options.remove(CONTINUE_TARGET.as_bytes());
            }
            c if c == OpCode::ScopeEnd as u16 => {
                let (label, ..) = labels.pop().unwrap_or_default();
                exits.retain(|&exit| {
                    if ops[exit].primary_data == label {
                        ops[exit].index = index as i32;
                        ops[exit].primary_data.clear();
                        false
                    } else {
                        true
                    }
                });
            }
            _ => {}
        }
    }
    match exits.first() {
        Some(&exit) => Err(undefined(&ops[exit].primary_data)),
        None => Ok(()),
    }
}
