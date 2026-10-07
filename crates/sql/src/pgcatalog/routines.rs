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

//! Catalog rows for user-defined functions, procedures, and triggers.

use pg_query::NodeEnum;

use crate::array::Array;
use crate::catalog::id::{self, SECTION_FUNCTION};
use crate::catalog::oids;
use crate::error::Result;
use crate::oid as types;
use crate::pgcatalog::rows::regproc;
use crate::pgcatalog::snapshot::{namespace_oid, table_oid};
use crate::pgcatalog::{Rows, boolean, int2, int4, oid, text};
use crate::query::Ctx;
use crate::routines::{Body, Mode, Routine};
use crate::triggers::{AFTER, BEFORE, INSTEAD_OF, names};
use crate::types::Value;

/// The OIDs of the languages in the catalog template.
const SQL_LANGUAGE: u32 = 14;
const PLPGSQL_LANGUAGE: u32 = 14035;
const C_LANGUAGE: u32 = 13;
const INTERNAL_LANGUAGE: u32 = 12;

/// The trigger type bits that pg_trigger's tgtype holds.
const TYPE_ROW: i16 = 1;
const TYPE_BEFORE: i16 = 1 << 1;
const TYPE_INSERT: i16 = 1 << 2;
const TYPE_DELETE: i16 = 1 << 3;
const TYPE_UPDATE: i16 = 1 << 4;
const TYPE_TRUNCATE: i16 = 1 << 5;
const TYPE_INSTEAD: i16 = 1 << 6;

/// routine_oid returns the OID of a routine.
pub fn routine_oid(routine: &Routine) -> u32 {
    oids::oid(&routine.object.id)
}

/// array returns a one-dimensional array of the element type.
fn array(element: u32, values: Vec<Value>) -> Value {
    Value::Array(Box::new(Array::one_dimensional(element, values)))
}

/// create_statement returns the CREATE FUNCTION or CREATE PROCEDURE statement a routine was defined by.
pub(super) fn create_statement(routine: &Routine) -> Option<pg_query::protobuf::CreateFunctionStmt> {
    let definition = String::from_utf8_lossy(&routine.object.definition);
    let raw = pg_query::parse(&definition).ok()?.protobuf.stmts.into_iter().next()?;
    match raw.stmt?.node? {
        NodeEnum::CreateFunctionStmt(create) => Some(*create),
        _ => None,
    }
}

/// source returns a routine's body as pg_proc's prosrc shows it: the SQL or PL/pgSQL text after AS.
pub(super) fn source(routine: &Routine) -> String {
    let body = create_statement(routine).and_then(|create| {
        create.options.iter().find_map(|option| {
            let Some(NodeEnum::DefElem(def)) = option.node.as_ref() else { return None };
            let Some(NodeEnum::List(list)) = def.arg.as_deref()?.node.as_ref() else { return None };
            match list.items.first()?.node.as_ref()? {
                NodeEnum::String(s) if def.defname == "as" => Some(s.sval.clone()),
                _ => None,
            }
        })
    });
    match (&routine.body, body) {
        (_, Some(body)) => body,
        (Body::Sql(sql), None) => sql.clone(),
        _ => String::new(),
    }
}

impl Ctx<'_> {
    /// pg_proc adds the user routines to pg_proc.
    pub(super) fn pg_proc(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        for routine in self.routines()?.iter() {
            let inputs: Vec<u32> = routine.inputs().map(|p| p.ty.oid).collect();
            let all_inputs = routine.params.iter().all(|p| p.mode == Mode::In);
            let modes = routine
                .params
                .iter()
                .map(|p| {
                    Value::Text(
                        match p.mode {
                            Mode::In => "i",
                            Mode::Out => "o",
                            Mode::InOut => "b",
                            Mode::Variadic => "v",
                        }
                        .into(),
                    )
                })
                .collect();
            let named = routine.params.iter().any(|p| !p.name.is_empty());
            let language = match routine.body {
                Body::Sql(_) => SQL_LANGUAGE,
                Body::PlPgSql(_) => PLPGSQL_LANGUAGE,
                Body::External => C_LANGUAGE,
            };
            let ret = if routine.procedure { 0 } else { routine.ret.oid };
            let ret = if routine.procedure && !routine.columns.is_empty() { types::RECORD } else { ret };
            rows.push(vec![
                ("oid", oid(routine_oid(routine))),
                ("proname", text(routine.name.clone())),
                ("pronamespace", oid(namespace_oid(&routine.schema))),
                ("proowner", oid(10)),
                ("prolang", oid(language)),
                ("procost", Value::Float4(100.0)),
                ("prorows", Value::Float4(if routine.set_of { 1000.0 } else { 0.0 })),
                ("provariadic", oid(0)),
                (
                    "prosupport",
                    Value::Reg(Box::new(crate::types::Reg { type_oid: types::REGPROC, oid: 0, name: "-".into() })),
                ),
                ("prokind", text(if routine.procedure { "p" } else { "f" })),
                ("prosecdef", boolean(false)),
                ("proleakproof", boolean(false)),
                ("proisstrict", boolean(routine.strict)),
                ("proretset", boolean(routine.set_of)),
                ("provolatile", text("v")),
                ("proparallel", text("u")),
                ("pronargs", int2(inputs.len() as i16)),
                ("pronargdefaults", int2(routine.params.iter().filter(|p| p.default.is_some()).count() as i16)),
                ("prorettype", oid(ret)),
                ("proargtypes", text(inputs.iter().map(u32::to_string).collect::<Vec<_>>().join(" "))),
                (
                    "proallargtypes",
                    if all_inputs {
                        Value::Null
                    } else {
                        array(types::OID, routine.params.iter().map(|p| Value::Oid(p.ty.oid)).collect())
                    },
                ),
                ("proargmodes", if all_inputs { Value::Null } else { array(types::CHAR, modes) }),
                (
                    "proargnames",
                    if named {
                        array(types::TEXT, routine.params.iter().map(|p| Value::Text(p.name.clone())).collect())
                    } else {
                        Value::Null
                    },
                ),
                ("prosrc", text(source(routine))),
            ]);
        }
        for aggregate in self.user_aggregates()?.iter() {
            rows.push(vec![
                ("oid", oid(oids::oid(&aggregate.stored.id))),
                ("proname", text(aggregate.name.clone())),
                ("pronamespace", oid(namespace_oid(&aggregate.schema))),
                ("proowner", oid(10)),
                ("prolang", oid(INTERNAL_LANGUAGE)),
                ("procost", Value::Float4(1.0)),
                ("prorows", Value::Float4(0.0)),
                ("provariadic", oid(0)),
                ("prosupport", regproc(&[])),
                ("prokind", text("a")),
                ("prosecdef", boolean(false)),
                ("proleakproof", boolean(false)),
                ("proisstrict", boolean(false)),
                ("proretset", boolean(false)),
                ("provolatile", text("i")),
                ("proparallel", text("u")),
                ("pronargs", int2(aggregate.params.len() as i16)),
                ("pronargdefaults", int2(0)),
                ("prorettype", oid(aggregate.ret.oid)),
                ("proargtypes", text(aggregate.params.iter().map(u32::to_string).collect::<Vec<_>>().join(" "))),
                ("prosrc", text("aggregate_dummy")),
            ]);
        }
        Ok(())
    }

    /// pg_operator lists the operators of the operator collection, as Go's pg_operator does.
    pub(super) fn pg_operator(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let operators = self.user_operators()?;
        let linked = |name: &[u8], left: u32, right: u32, schema: &str| {
            operators
                .iter()
                .find(|o| o.name.as_bytes() == name && o.left == left && o.right == right && o.schema == schema)
                .map_or(0, |o| oids::oid(&o.stored.id))
        };
        for operator in operators.iter() {
            let stored = &operator.stored;
            rows.push(vec![
                ("oid", oid(oids::oid(&stored.id))),
                ("oprname", text(operator.name.clone())),
                ("oprnamespace", oid(namespace_oid(&operator.schema))),
                ("oprowner", oid(10)),
                ("oprkind", text(if operator.left == 0 { "l" } else { "b" })),
                ("oprcanmerge", boolean(stored.merges)),
                ("oprcanhash", boolean(stored.hashes)),
                ("oprleft", oid(operator.left)),
                ("oprright", oid(operator.right)),
                ("oprresult", oid(operator.routine.ret.oid)),
                ("oprcom", oid(linked(&stored.commutator, operator.right, operator.left, &operator.schema))),
                ("oprnegate", oid(linked(&stored.negator, operator.left, operator.right, &operator.schema))),
                ("oprcode", regproc(&stored.function)),
                ("oprrest", regproc(&[])),
                ("oprjoin", regproc(&[])),
            ]);
        }
        Ok(())
    }

    /// pg_aggregate lists the aggregates of the aggregate collection, as Go's pg_aggregate does.
    pub(super) fn pg_aggregate(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        for aggregate in self.user_aggregates()?.iter() {
            let stored = &aggregate.stored;
            rows.push(vec![
                ("aggfnoid", regproc(&stored.id)),
                ("aggkind", text("n")),
                ("aggnumdirectargs", int2(0)),
                ("aggtransfn", regproc(&stored.s_func)),
                ("aggfinalfn", regproc(&stored.final_func)),
                ("aggcombinefn", regproc(&stored.combine_func)),
                ("aggserialfn", regproc(&[])),
                ("aggdeserialfn", regproc(&[])),
                ("aggmtransfn", regproc(&[])),
                ("aggminvtransfn", regproc(&[])),
                ("aggmfinalfn", regproc(&[])),
                ("aggfinalextra", boolean(false)),
                ("aggmfinalextra", boolean(false)),
                ("aggfinalmodify", text("r")),
                ("aggmfinalmodify", text("r")),
                ("aggsortop", oid(0)),
                ("aggtranstype", oid(aggregate.state_type.oid)),
                ("aggtransspace", int4(0)),
                ("aggmtranstype", oid(0)),
                ("aggmtransspace", int4(0)),
                ("agginitval", aggregate.init_cond.clone().map_or(Value::Null, text)),
            ]);
        }
        Ok(())
    }

    /// pg_trigger lists the user triggers.
    pub(super) fn pg_trigger(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        for trigger in self.triggers()?.iter() {
            let (schema, table, name) = names(trigger);
            let function = self.trigger_routine_oid(&trigger.function)?;
            let columns = self.txn.table(self.db, &schema, &table)?.map(|t| t.columns).unwrap_or_default();
            let attributes: Vec<String> = trigger
                .events
                .iter()
                .flat_map(|e| &e.column_names)
                .filter_map(|c| columns.iter().position(|col| col.name.as_bytes() == c.as_slice()))
                .map(|i| (i + 1).to_string())
                .collect();
            let mut arguments = Vec::new();
            for argument in &trigger.arguments {
                arguments.extend_from_slice(argument);
                arguments.push(0);
            }
            rows.push(vec![
                ("oid", oid(oids::oid(&trigger.id))),
                ("tgrelid", oid(table_oid(&schema, &table))),
                ("tgparentid", oid(0)),
                ("tgname", text(name)),
                ("tgfoid", oid(function)),
                ("tgtype", int2(trigger_type(trigger))),
                ("tgenabled", text("O")),
                ("tgisinternal", boolean(false)),
                ("tgconstrrelid", oid(0)),
                ("tgconstrindid", oid(0)),
                ("tgconstraint", oid(0)),
                ("tgdeferrable", boolean(false)),
                ("tginitdeferred", boolean(false)),
                ("tgnargs", int2(trigger.arguments.len() as i16)),
                ("tgattr", text(attributes.join(" "))),
                ("tgargs", crate::cast::input(&bytes_text(&arguments), 17).unwrap_or(Value::Null)),
            ]);
        }
        Ok(())
    }

    /// trigger_routine_oid returns the OID of the function a trigger runs, or 0 when it is missing.
    fn trigger_routine_oid(&mut self, function: &[u8]) -> Result<u32> {
        let segments = id::segments(function);
        let (schema, name) =
            (segments.first().cloned().unwrap_or_default(), segments.get(1).cloned().unwrap_or_default());
        let schema = (!schema.is_empty()).then_some(schema);
        let found = self.routines_named(schema.as_deref(), &name)?.into_iter().find(|r| r.inputs().next().is_none());
        Ok(found.map_or_else(|| oids::oid(&id::new(SECTION_FUNCTION, &[&name])), |r| routine_oid(&r)))
    }

    /// information_schema_triggers lists each event of each user trigger.
    pub(super) fn information_schema_triggers(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let database = self.session.display.clone();
        let triggers = self.triggers()?;
        let mut entries: Vec<(String, String, String, u8, u8, &objects::Trigger)> = Vec::new();
        for trigger in triggers.iter() {
            let (schema, table, name) = names(trigger);
            for event in &trigger.events {
                entries.push((schema.clone(), table.clone(), name.clone(), event.event_type, trigger.timing, trigger));
            }
        }
        entries.sort_by(|a, b| (&a.0, &a.1, a.3, a.4, &a.2).cmp(&(&b.0, &b.1, b.3, b.4, &b.2)));
        let mut previous: Option<(String, String, u8, u8)> = None;
        let mut order = 0;
        for (schema, table, name, event, timing, trigger) in entries {
            let key = (schema.clone(), table.clone(), event, timing);
            order = if previous.as_ref() == Some(&key) { order + 1 } else { 1 };
            previous = Some(key);
            let (condition, statement) = trigger_clauses(trigger);
            rows.push(vec![
                ("trigger_catalog", text(database.clone())),
                ("trigger_schema", text(schema.clone())),
                ("trigger_name", text(name)),
                ("event_manipulation", text(["INSERT", "UPDATE", "DELETE", "TRUNCATE"][event.min(3) as usize])),
                ("event_object_catalog", text(database.clone())),
                ("event_object_schema", text(schema)),
                ("event_object_table", text(table)),
                ("action_order", int4(order)),
                ("action_condition", condition.map_or(Value::Null, text)),
                ("action_statement", text(statement)),
                ("action_orientation", text(if trigger.for_each_row { "ROW" } else { "STATEMENT" })),
                (
                    "action_timing",
                    text(match timing {
                        BEFORE => "BEFORE",
                        AFTER => "AFTER",
                        INSTEAD_OF => "INSTEAD OF",
                        _ => "",
                    }),
                ),
            ]);
        }
        Ok(())
    }

    /// pg_cast lists the casts of the cast collection, as Go's pg_cast does.
    pub(super) fn pg_cast(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        for cast in self.user_casts()?.iter() {
            let method = match (&cast.routine, cast.stored.use_in_out) {
                (Some(_), _) => "f",
                (None, true) => "i",
                (None, false) => "b",
            };
            rows.push(vec![
                ("oid", oid(oids::oid(&cast.stored.id))),
                ("castsource", oid(cast.source)),
                ("casttarget", oid(cast.target)),
                ("castfunc", oid(cast.routine.as_deref().map_or(0, routine_oid))),
                ("castcontext", text(crate::casts::context_code(cast.context))),
                ("castmethod", text(method)),
            ]);
        }
        Ok(())
    }
}

/// trigger_type returns the type bits pg_trigger's tgtype holds for a trigger.
fn trigger_type(trigger: &objects::Trigger) -> i16 {
    let mut bits = if trigger.for_each_row { TYPE_ROW } else { 0 };
    bits |= match trigger.timing {
        BEFORE => TYPE_BEFORE,
        INSTEAD_OF => TYPE_INSTEAD,
        _ => 0,
    };
    for event in &trigger.events {
        bits |= [TYPE_INSERT, TYPE_UPDATE, TYPE_DELETE, TYPE_TRUNCATE][event.event_type.min(3) as usize];
    }
    bits
}

/// bytes_text returns bytes in bytea's escape input format.
fn bytes_text(bytes: &[u8]) -> String {
    let mut out = String::from("\\x");
    for b in bytes {
        out.push_str(&format!("{b:02x}"));
    }
    out
}

/// trigger_clauses returns a trigger's WHEN condition and its EXECUTE FUNCTION clause as information_schema shows
/// them, from the CREATE TRIGGER statement that defined it.
fn trigger_clauses(trigger: &objects::Trigger) -> (Option<String>, String) {
    let definition = String::from_utf8_lossy(&trigger.definition);
    let create = pg_query::parse(&definition).ok().and_then(|result| {
        let raw = result.protobuf.stmts.into_iter().next()?;
        match raw.stmt?.node? {
            NodeEnum::CreateTrigStmt(create) => Some(create),
            _ => None,
        }
    });
    let segments = id::segments(&trigger.function);
    let name = segments.get(1).cloned().unwrap_or_default();
    let arguments: Vec<String> =
        trigger.arguments.iter().map(|a| format!("'{}'", String::from_utf8_lossy(a).replace('\'', "''"))).collect();
    let statement = format!("EXECUTE FUNCTION {name}({})", arguments.join(", "));
    let condition =
        create.and_then(|c| c.when_clause).and_then(|w| crate::ddl::expression_text(&w).ok()).map(|w| format!("({w})"));
    (condition, statement)
}
