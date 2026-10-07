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

//! Triggers: Go's trigger root objects, CREATE and DROP TRIGGER, and firing a table's triggers around the rows a
//! statement writes, as Postgres does.

use std::sync::Arc;

use doltdb::database::Database;
use doltdb::root::Root;
use objects::{Function, Trigger, TriggerEvent};
use pg_query::NodeEnum;
use pg_query::protobuf::{CreateTrigStmt, DropStmt};
use store::Hash;

use crate::auth::Object;
use crate::catalog::ColumnType;
use crate::catalog::id::{self, SECTION_FUNCTION, SECTION_TRIGGER};
use crate::catalog::table::TableDef;
use crate::error::{PgError, Result, code};
use crate::expr::{node_name, position};
use crate::query::Ctx;
use crate::routines::{Body, Routine, TRIGGER};
use crate::types::Value;
use crate::{Outcome, oid};

/// COLLECTION is the position of the trigger collection among a root value's root object collections.
pub const COLLECTION: usize = 3;

/// The timings Go stores.
pub const BEFORE: u8 = 0;
pub const AFTER: u8 = 1;
pub const INSTEAD_OF: u8 = 2;

/// Event is a kind of change that fires triggers, numbered as Go stores it.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Event {
    Insert = 0,
    Update = 1,
    Delete = 2,
    Truncate = 3,
}

impl Event {
    /// name returns the event's name, which TG_OP holds.
    fn name(self) -> &'static str {
        match self {
            Event::Insert => "INSERT",
            Event::Update => "UPDATE",
            Event::Delete => "DELETE",
            Event::Truncate => "TRUNCATE",
        }
    }
}

/// The trigger type bits of Postgres' parse tree.
const TYPE_BEFORE: i32 = 1 << 1;
const TYPE_INSERT: i32 = 1 << 2;
const TYPE_DELETE: i32 = 1 << 3;
const TYPE_UPDATE: i32 = 1 << 4;
const TYPE_TRUNCATE: i32 = 1 << 5;
const TYPE_INSTEAD: i32 = 1 << 6;

/// trigger_id returns the ID of a trigger on a table.
fn trigger_id(schema: &str, table: &str, name: &str) -> Vec<u8> {
    id::new(SECTION_TRIGGER, &[schema, table, name])
}

/// all returns every trigger of a root value.
fn all(db: &mut Database, root: &Root) -> Result<Vec<Arc<Trigger>>> {
    let mut triggers = Vec::new();
    for (_, address) in root.objects(db, COLLECTION)? {
        triggers.push(Arc::new(Trigger::deserialize(&prolly::read_blob(db, &address)?)?));
    }
    Ok(triggers)
}

/// store writes a trigger into a root value.
fn store(db: &mut Database, root: &mut Root, trigger: &Trigger) -> Result<()> {
    let data = trigger.serialize();
    let mut sink = |_: Hash, bytes: &[u8]| {
        db.write_value(bytes.to_vec()).map(|_| ()).map_err(|e| store::Error::Corrupt(e.to_string()))
    };
    let (address, _) = prolly::write_blob(&data, &mut sink)?.ok_or_else(|| PgError::internal("an empty trigger"))?;
    root.put_object(db, COLLECTION, &trigger.id, Some(address))?;
    Ok(())
}

/// names returns a trigger's schema, table, and name.
pub(crate) fn names(trigger: &Trigger) -> (String, String, String) {
    let mut segments = id::segments(&trigger.id).into_iter();
    let schema = segments.next().unwrap_or_default();
    let table = segments.next().unwrap_or_default();
    (schema, table, segments.next().unwrap_or_default())
}

/// Fired is a trigger of a table, ready to fire, with its function, WHEN condition, and the columns of an UPDATE OF.
pub struct Fired {
    trigger: Arc<Trigger>,
    name: String,
    function: Arc<Routine>,
    when: Option<Routine>,
    update_columns: Vec<usize>,
}

/// Triggers are the triggers of a table that a statement may fire, in name order, as Postgres fires them.
pub struct Triggers {
    table: TableDef,
    row_type: ColumnType,
    columns: Vec<(String, ColumnType)>,
    list: Vec<Fired>,
}

impl Ctx<'_> {
    /// triggers returns every trigger of the working root, reusing them while the trigger collection is unchanged.
    pub(crate) fn triggers(&mut self) -> Result<Arc<Vec<Arc<Trigger>>>> {
        let address = self.txn.root.root_objects[COLLECTION];
        if let Some((cached, triggers)) = &self.session.triggers
            && *cached == address
        {
            return Ok(triggers.clone());
        }
        let triggers = Arc::new(all(self.db, &self.txn.root)?);
        self.session.triggers = Some((address, triggers.clone()));
        Ok(triggers)
    }

    /// table_triggers returns the triggers of a table, with their functions resolved.
    pub fn table_triggers(&mut self, table: &TableDef) -> Result<Triggers> {
        let mut list = Vec::new();
        for trigger in self.triggers()?.iter() {
            let (schema, relation, name) = names(trigger);
            if schema != table.schema || relation != table.name {
                continue;
            }
            let function = self.trigger_function(&trigger.function)?;
            let when = (!trigger.when.is_empty()).then(|| Routine::condition(trigger.when.clone()));
            let update_columns = trigger
                .events
                .iter()
                .filter(|e| e.event_type == Event::Update as u8)
                .flat_map(|e| &e.column_names)
                .filter_map(|name| table.columns.iter().position(|c| c.name.as_bytes() == name.as_slice()))
                .collect();
            list.push(Fired { trigger: trigger.clone(), name, function, when, update_columns });
        }
        list.sort_by(|a, b| a.name.cmp(&b.name));
        let columns = table.columns.iter().map(|c| (c.name.clone(), c.ty)).collect();
        let row_type = crate::expr::typ(crate::usertypes::register_row_type(table));
        Ok(Triggers { table: table.clone(), row_type, columns, list })
    }

    /// trigger_function returns the function a trigger runs, which Go names by schema and name, finding an
    /// unqualified one on the search path.
    fn trigger_function(&mut self, function_id: &[u8]) -> Result<Arc<Routine>> {
        let mut segments = id::segments(function_id).into_iter();
        let schema = segments.next().unwrap_or_default();
        let name = segments.next().unwrap_or_default();
        let schema = (!schema.is_empty()).then_some(schema);
        let found = self.routines_named(schema.as_deref(), &name)?.into_iter().find(|r| r.inputs().next().is_none());
        found.ok_or_else(|| PgError::new(code::UNDEFINED_FUNCTION, format!("function {name}() does not exist")))
    }

    /// create_trigger runs CREATE TRIGGER.
    pub fn create_trigger(&mut self, stmt: &CreateTrigStmt, text: &str) -> Result<Outcome> {
        if stmt.isconstraint {
            return Err(PgError::unsupported("CREATE CONSTRAINT TRIGGER"));
        }
        if !stmt.transition_rels.is_empty() {
            return Err(PgError::unsupported("REFERENCING in CREATE TRIGGER"));
        }
        let relation = stmt.relation.as_ref().ok_or_else(|| PgError::internal("CREATE TRIGGER without a table"))?;
        let table = self.resolve_table(relation).map_err(|err| PgError { position: None, ..err })?;
        self.require_owner(&Object::Table(table.schema.clone(), table.name.clone()))?;
        let timing = if stmt.timing & TYPE_INSTEAD != 0 {
            INSTEAD_OF
        } else if stmt.timing & TYPE_BEFORE != 0 {
            BEFORE
        } else {
            AFTER
        };
        if timing == INSTEAD_OF {
            return Err(PgError {
                detail: Some("Tables cannot have INSTEAD OF triggers.".into()),
                ..PgError::new(code::WRONG_OBJECT_TYPE, format!("\"{}\" is a table", table.name))
            });
        }
        let names: Vec<&str> = stmt.funcname.iter().filter_map(node_name).collect();
        let (function_schema, function_name) = match names.as_slice() {
            [name] => (String::new(), name.to_string()),
            [.., schema, name] => (schema.to_string(), name.to_string()),
            [] => return Err(PgError::internal("CREATE TRIGGER without a function")),
        };
        let function = self.trigger_function(&id::new(SECTION_FUNCTION, &[&function_schema, &function_name]))?;
        if function.ret.oid != TRIGGER {
            return Err(PgError::new(
                code::INVALID_OBJECT_DEFINITION,
                format!("function {function_name} must return type trigger"),
            ));
        }
        let mut columns = Vec::new();
        for column in stmt.columns.iter().filter_map(node_name) {
            if !table.columns.iter().any(|c| c.name == column) {
                return Err(PgError::new(
                    code::UNDEFINED_COLUMN,
                    format!("column \"{column}\" of relation \"{}\" does not exist", table.name),
                ));
            }
            if columns.iter().any(|c| c == column) {
                return Err(PgError::new(
                    code::DUPLICATE_COLUMN,
                    format!("column \"{column}\" specified more than once"),
                ));
            }
            columns.push(column.to_string());
        }
        let mut events = Vec::new();
        for (bit, event) in [(TYPE_INSERT, Event::Insert), (TYPE_UPDATE, Event::Update), (TYPE_DELETE, Event::Delete)] {
            if stmt.events & bit != 0 {
                let column_names = if event == Event::Update {
                    columns.iter().map(|c| c.clone().into_bytes()).collect()
                } else {
                    Vec::new()
                };
                events.push(TriggerEvent { event_type: event as u8, column_names });
            }
        }
        if stmt.events & TYPE_TRUNCATE != 0 {
            if stmt.row {
                return Err(PgError::new(
                    code::FEATURE_NOT_SUPPORTED,
                    "TRUNCATE FOR EACH ROW triggers are not supported",
                ));
            }
            events.push(TriggerEvent { event_type: Event::Truncate as u8, column_names: Vec::new() });
        }
        let when = match &stmt.when_clause {
            Some(clause) => self.compile_when(stmt, &table, clause)?,
            None => Vec::new(),
        };
        let mut arguments = Vec::new();
        for arg in &stmt.args {
            arguments.push(match arg.node.as_ref() {
                Some(NodeEnum::String(s)) => s.sval.clone(),
                Some(NodeEnum::Integer(i)) => i.ival.to_string(),
                Some(NodeEnum::Float(f)) => f.fval.clone(),
                _ => String::new(),
            });
        }
        let trigger = Trigger {
            id: trigger_id(&table.schema, &table.name, &stmt.trigname),
            function: id::new(SECTION_FUNCTION, &[&function_schema, &function_name]),
            timing,
            events,
            for_each_row: stmt.row,
            when,
            arguments: arguments.into_iter().map(String::into_bytes).collect(),
            definition: text.as_bytes().to_vec(),
            ..Trigger::default()
        };
        if self.triggers()?.iter().any(|t| t.id == trigger.id) && !stmt.replace {
            return Err(PgError::new(
                code::DUPLICATE_OBJECT,
                format!("trigger \"{}\" for relation \"{}\" already exists", stmt.trigname, table.name),
            ));
        }
        store(self.db, &mut self.txn.root, &trigger)?;
        Ok(Outcome::command("CREATE TRIGGER"))
    }

    /// compile_when checks a trigger's WHEN condition as Postgres does, and compiles it as the body of a trigger
    /// function returning it, as Go stores it.
    fn compile_when(
        &mut self,
        stmt: &CreateTrigStmt,
        table: &TableDef,
        clause: &pg_query::Node,
    ) -> Result<Vec<objects::Operation>> {
        let mut scope = crate::expr::Scope::default();
        let table_oid = crate::pgcatalog::snapshot::table_oid(&table.schema, &table.name);
        crate::usertypes::register_row_type(table);
        for record in ["new", "old"] {
            for (i, column) in table.columns.iter().enumerate() {
                scope.columns.push(crate::expr::ScopeColumn {
                    table: record.into(),
                    name: column.name.clone(),
                    ty: column.ty,
                    hidden: true,
                    origin: (table_oid, i as u16 + 1),
                });
            }
        }
        let mut binder = crate::expr::Binder::new(self, scope);
        binder.clause = "WHEN expressions";
        binder.definition = true;
        let bound = binder.bind(clause);
        let referenced: Vec<(usize, i32)> = binder.columns.clone();
        match bound {
            Ok((_, ty)) if ty.oid != oid::BOOL => {
                return Err(PgError {
                    position: position(leftmost(clause)),
                    ..PgError::new(
                        code::DATATYPE_MISMATCH,
                        format!(
                            "argument of WHEN must be type boolean, not type {}",
                            crate::cast::type_display(ty.oid)
                        ),
                    )
                });
            }
            Ok(_) => {}
            Err(err) if err.code == code::FEATURE_NOT_SUPPORTED => {}
            Err(err) => return Err(err),
        }
        let width = table.columns.len();
        let uses_new = referenced.iter().find(|(i, _)| *i < width).map(|(_, l)| *l);
        let uses_old = referenced.iter().find(|(i, _)| *i >= width).map(|(_, l)| *l);
        let fail = |message: &str, location: i32| {
            Err(PgError {
                position: position(location),
                ..PgError::new(code::INVALID_OBJECT_DEFINITION, message.to_string())
            })
        };
        if let (false, Some(location)) = (stmt.row, uses_new.or(uses_old)) {
            return fail("statement trigger's WHEN condition cannot reference column values", location);
        }
        if let Some(location) = uses_new.filter(|_| stmt.events & TYPE_DELETE != 0) {
            return fail("DELETE trigger's WHEN condition cannot reference NEW values", location);
        }
        if let Some(location) = uses_old.filter(|_| stmt.events & TYPE_INSERT != 0) {
            return fail("INSERT trigger's WHEN condition cannot reference OLD values", location);
        }
        let condition = crate::ddl::expression_text(clause)?;
        let wrapper = format!(
            "CREATE FUNCTION when_wrapper() RETURNS TRIGGER AS $$\nBEGIN\n\tRETURN ({condition});\nEND;\n$$ LANGUAGE plpgsql;"
        );
        let body = format!("\nBEGIN\n\tRETURN ({condition});\nEND;\n");
        crate::plpgsql::compile(self, &wrapper, &body)
    }

    /// drop_triggers runs DROP TRIGGER.
    pub fn drop_triggers(&mut self, drop: &DropStmt) -> Result<Outcome> {
        for object in &drop.objects {
            let Some(NodeEnum::List(list)) = object.node.as_ref() else { continue };
            let names: Vec<&str> = list.items.iter().filter_map(node_name).collect();
            let (schema, table, name) = match names.as_slice() {
                [table, name] => (String::new(), table.to_string(), name.to_string()),
                [.., schema, table, name] => (schema.to_string(), table.to_string(), name.to_string()),
                _ => continue,
            };
            let relation = pg_query::protobuf::RangeVar {
                schemaname: schema,
                relname: table.clone(),
                inh: true,
                location: -1,
                ..Default::default()
            };
            let table = match self.resolve_table(&relation) {
                Ok(table) => table,
                Err(err) if drop.missing_ok && err.code == code::UNDEFINED_TABLE => {
                    self.session
                        .notice(PgError::notice("00000", format!("relation \"{table}\" does not exist, skipping")));
                    continue;
                }
                Err(err) => return Err(err),
            };
            let id = trigger_id(&table.schema, &table.name, &name);
            if !self.triggers()?.iter().any(|t| t.id == id) {
                if drop.missing_ok {
                    self.session.notice(PgError::notice(
                        "00000",
                        format!("trigger \"{name}\" for relation \"{}\" does not exist, skipping", table.name),
                    ));
                    continue;
                }
                return Err(PgError::new(
                    code::UNDEFINED_OBJECT,
                    format!("trigger \"{name}\" for table \"{}\" does not exist", table.name),
                ));
            }
            self.require_owner(&Object::Table(table.schema.clone(), table.name.clone()))?;
            self.txn.root.put_object(self.db, COLLECTION, &id, None)?;
        }
        Ok(Outcome::command("DROP TRIGGER"))
    }

    /// drop_table_triggers drops the triggers of a table that is being dropped.
    pub fn drop_table_triggers(&mut self, schema: &str, table: &str) -> Result<()> {
        for trigger in self.triggers()?.iter() {
            let (s, t, _) = names(trigger);
            if s == schema && t == table {
                self.txn.root.put_object(self.db, COLLECTION, &trigger.id, None)?;
            }
        }
        Ok(())
    }

    /// rename_trigger_column follows a column rename in the UPDATE OF columns of the table's triggers.
    pub fn rename_trigger_column(&mut self, schema: &str, table: &str, old: &str, new: &str) -> Result<()> {
        for trigger in self.triggers()?.iter() {
            let (s, t, _) = names(trigger);
            let names_column = |e: &TriggerEvent| e.column_names.iter().any(|c| c == old.as_bytes());
            if s != schema || t != table || !trigger.events.iter().any(names_column) {
                continue;
            }
            let mut renamed = (**trigger).clone();
            for column in renamed.events.iter_mut().flat_map(|e| &mut e.column_names) {
                if column == old.as_bytes() {
                    *column = new.as_bytes().to_vec();
                }
            }
            store(self.db, &mut self.txn.root, &renamed)?;
        }
        Ok(())
    }

    /// drop_trigger_column fails as Postgres does when triggers of the table name a dropped column in UPDATE OF, or
    /// drops those triggers with a notice for CASCADE.
    pub fn drop_trigger_column(&mut self, schema: &str, table: &str, column: &str, cascade: bool) -> Result<()> {
        let mut dependents = Vec::new();
        for trigger in self.triggers()?.iter() {
            let (s, t, name) = names(trigger);
            let names_column = |e: &TriggerEvent| e.column_names.iter().any(|c| c == column.as_bytes());
            if s == schema && t == table && trigger.events.iter().any(names_column) {
                dependents.push((trigger.id.clone(), format!("trigger {name} on table {table}")));
            }
        }
        if dependents.is_empty() {
            return Ok(());
        }
        if !cascade {
            let detail = dependents.iter().map(|(_, d)| format!("{d} depends on column {column} of table {table}"));
            return Err(PgError {
                detail: Some(detail.collect::<Vec<_>>().join("\n")),
                hint: Some("Use DROP ... CASCADE to drop the dependent objects too.".into()),
                ..PgError::new(
                    code::DEPENDENT_OBJECTS_STILL_EXIST,
                    format!("cannot drop column {column} of table {table} because other objects depend on it"),
                )
            });
        }
        for (id, dependent) in dependents {
            self.session.notice(PgError::notice("00000", format!("drop cascades to {dependent}")));
            self.txn.root.put_object(self.db, COLLECTION, &id, None)?;
        }
        Ok(())
    }

    /// trigger_dependents returns the triggers that run a function, as `trigger t on table x`.
    pub fn trigger_dependents(&mut self, routine: &Routine) -> Result<Vec<String>> {
        let mut dependents = Vec::new();
        for trigger in self.triggers()?.iter() {
            let segments = id::segments(&trigger.function);
            let (schema, name) =
                (segments.first().cloned().unwrap_or_default(), segments.get(1).cloned().unwrap_or_default());
            if name == routine.name && (schema.is_empty() || schema == routine.schema) {
                let (_, table, trigger_name) = names(trigger);
                dependents.push(format!("trigger {trigger_name} on table {table}"));
            }
        }
        Ok(dependents)
    }
}

/// leftmost returns the location of an expression's leftmost token, where Postgres reports errors about it.
fn leftmost(node: &pg_query::Node) -> i32 {
    let location = crate::expr::arg_location(node);
    let inner = match node.node.as_ref() {
        Some(NodeEnum::AExpr(e)) => e.lexpr.as_deref().map(leftmost),
        Some(NodeEnum::BoolExpr(b)) => b.args.first().map(leftmost),
        _ => None,
    };
    match inner {
        Some(inner) if inner >= 0 && (location < 0 || inner < location) => inner,
        _ => location,
    }
}

impl Triggers {
    /// fires reports whether any trigger fires for the event at the timing and level.
    pub fn fires(&self, event: Event, timing: u8, row: bool) -> bool {
        self.list.iter().any(|f| f.matches(event, timing, row, None))
    }

    /// before_row runs the BEFORE ROW triggers of an event on a row, the old row for a DELETE, returning the row to
    /// write, which a trigger may change, or None when a trigger returned NULL to skip it.
    pub fn before_row(
        &self,
        ctx: &mut Ctx<'_>,
        event: Event,
        old: Option<&[Value]>,
        new: Option<Vec<Value>>,
        updated: &[usize],
    ) -> Result<Option<Vec<Value>>> {
        let mut new = new;
        let mut old_row = old.map(<[Value]>::to_vec);
        for fired in &self.list {
            if !fired.matches(event, BEFORE, true, Some(updated)) {
                continue;
            }
            if !self.condition(ctx, fired, old, new.as_deref())? {
                continue;
            }
            let returned = self.call(ctx, fired, event, BEFORE, "ROW", old, new.clone())?;
            match returned {
                None => return Ok(None),
                Some(row) if event == Event::Delete => old_row = Some(row),
                Some(row) => new = Some(row),
            }
        }
        Ok(if event == Event::Delete { old_row } else { new })
    }

    /// after_row runs the AFTER ROW triggers of an event on a row.
    pub fn after_row(
        &self,
        ctx: &mut Ctx<'_>,
        event: Event,
        old: Option<&[Value]>,
        new: Option<&[Value]>,
        updated: &[usize],
    ) -> Result<()> {
        for fired in &self.list {
            if !fired.matches(event, AFTER, true, Some(updated)) {
                continue;
            }
            if !self.condition(ctx, fired, old, new)? {
                continue;
            }
            self.call(ctx, fired, event, AFTER, "ROW", old, new.map(<[Value]>::to_vec))?;
        }
        Ok(())
    }

    /// statement runs the statement-level triggers of an event at the timing.
    pub fn statement(&self, ctx: &mut Ctx<'_>, event: Event, timing: u8, updated: &[usize]) -> Result<()> {
        for fired in &self.list {
            if !fired.matches(event, timing, false, Some(updated)) {
                continue;
            }
            if !self.condition(ctx, fired, None, None)? {
                continue;
            }
            self.call(ctx, fired, event, timing, "STATEMENT", None, None)?;
        }
        Ok(())
    }

    /// condition evaluates a trigger's WHEN condition on the rows, which holds without one.
    fn condition(
        &self,
        ctx: &mut Ctx<'_>,
        fired: &Fired,
        old: Option<&[Value]>,
        new: Option<&[Value]>,
    ) -> Result<bool> {
        let Some(when) = &fired.when else { return Ok(true) };
        let Body::PlPgSql(ops) = &when.body else { return Ok(true) };
        let value = crate::plpgsql::call_condition(
            ctx,
            when,
            ops,
            self.row_type,
            &self.columns,
            new.map(<[Value]>::to_vec),
            old.map(<[Value]>::to_vec),
        )?;
        Ok(matches!(value, Value::Bool(true)))
    }

    /// call runs a trigger's function with its rows and special variables.
    #[allow(clippy::too_many_arguments)]
    fn call(
        &self,
        ctx: &mut Ctx<'_>,
        fired: &Fired,
        event: Event,
        timing: u8,
        level: &str,
        old: Option<&[Value]>,
        new: Option<Vec<Value>>,
    ) -> Result<Option<Vec<Value>>> {
        let Body::PlPgSql(ops) = &fired.function.body else {
            return Err(PgError::unsupported(format!("the trigger function {}", fired.function.signature())));
        };
        let table_id = id::new(crate::catalog::id::SECTION_TABLE, &[&self.table.schema, &self.table.name]);
        let arguments: Vec<Value> =
            fired.trigger.arguments.iter().map(|a| Value::Text(String::from_utf8_lossy(a).into_owned())).collect();
        let argv = crate::array::Array {
            element: oid::TEXT,
            dims: if arguments.is_empty() { Vec::new() } else { vec![(arguments.len() as i32, 0)] },
            values: arguments.clone(),
        };
        let special = vec![
            ("tg_name", Value::Text(fired.name.clone())),
            ("tg_when", Value::Text(if timing == BEFORE { "BEFORE" } else { "AFTER" }.into())),
            ("tg_level", Value::Text(level.into())),
            ("tg_op", Value::Text(event.name().into())),
            ("tg_relid", Value::Oid(crate::catalog::oids::oid(&table_id))),
            ("tg_relname", Value::Text(self.table.name.clone())),
            ("tg_table_name", Value::Text(self.table.name.clone())),
            ("tg_table_schema", Value::Text(self.table.schema.clone())),
            ("tg_nargs", Value::Int4(arguments.len() as i32)),
            ("tg_argv", Value::Array(Box::new(argv))),
        ];
        let (old, new) = match event {
            Event::Insert => (None, new),
            Event::Delete => (old.map(<[Value]>::to_vec), None),
            _ => (old.map(<[Value]>::to_vec), new),
        };
        crate::routines::check_depth(ctx)?;
        ctx.session.call_depth += 1;
        let result =
            crate::plpgsql::call_trigger(ctx, &fired.function, ops, self.row_type, &self.columns, new, old, special);
        ctx.session.call_depth -= 1;
        result.map_err(|err| PgError { position: None, ..err })
    }
}

impl Fired {
    /// matches reports whether the trigger fires for an event at a timing and level, given the columns an UPDATE
    /// sets, one of which an UPDATE OF trigger needs.
    fn matches(&self, event: Event, timing: u8, row: bool, updated: Option<&[usize]>) -> bool {
        self.trigger.timing == timing
            && self.trigger.for_each_row == row
            && self.trigger.events.iter().any(|e| e.event_type == event as u8)
            && (event != Event::Update
                || self.update_columns.is_empty()
                || updated.is_none_or(|u| self.update_columns.iter().any(|c| u.contains(c))))
    }
}

impl Routine {
    /// condition returns a routine that runs a trigger's compiled WHEN condition.
    pub fn condition(operations: Vec<objects::Operation>) -> Routine {
        let object = Function { operations, ..Function::default() };
        Routine::internal(object, crate::expr::typ(oid::BOOL))
    }
}
