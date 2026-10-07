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

//! The aggregates of a root value's aggregate collection, which CREATE AGGREGATE and emulated extensions create.

use std::cell::RefCell;
use std::sync::Arc;

use pg_query::NodeEnum;
use pg_query::protobuf::{DefineStmt, DropStmt, ObjectWithArgs};

use crate::Outcome;
use crate::catalog::{ColumnType, id};
use crate::error::{PgError, Result, code};
use crate::query::Ctx;
use crate::routines::Routine;
use crate::types::Value;

/// COLLECTION is the position of the aggregates in a root value's root object collections.
pub const COLLECTION: usize = 9;

/// UserAggregate is an aggregate of the aggregate collection, with the routines that compute it.
#[derive(Debug, PartialEq)]
pub struct UserAggregate {
    pub schema: String,
    pub name: String,
    pub params: Vec<u32>,
    pub ret: ColumnType,
    pub state_type: ColumnType,
    pub transition: Arc<Routine>,
    pub final_routine: Option<Arc<Routine>>,
    /// The state's initial value as text, or None to start from the first input.
    pub init_cond: Option<String>,
    pub stored: objects::Aggregate,
}

/// AggregateCache is the aggregates of a root value, with the addresses of the aggregate and function collections.
pub type AggregateCache = ((Option<store::Hash>, Option<store::Hash>), Arc<Vec<Arc<UserAggregate>>>);

thread_local! {
    /// INSTALLED is the aggregates of the running statement's root value.
    static INSTALLED: RefCell<Arc<Vec<Arc<UserAggregate>>>> = RefCell::new(Arc::new(Vec::new()));
}

/// named reports whether a call names an aggregate, by its schema when the call gives one, or else by the search path.
fn named(aggregate: &UserAggregate, schema: Option<&str>, name: &str) -> bool {
    aggregate.name == name
        && match schema {
            Some(schema) => aggregate.schema == schema,
            None => crate::usertypes::in_search_path(&aggregate.schema),
        }
}

/// exists reports whether a call names an installed aggregate.
pub fn exists(schema: Option<&str>, name: &str) -> bool {
    INSTALLED.with(|aggregates| aggregates.borrow().iter().any(|a| named(a, schema, name)))
}

/// find returns the installed aggregate that a call names for arguments of the types, where an untyped argument
/// matches any parameter.
pub fn find(schema: Option<&str>, name: &str, types: &[u32]) -> Option<Arc<UserAggregate>> {
    INSTALLED.with(|aggregates| {
        aggregates
            .borrow()
            .iter()
            .find(|a| {
                named(a, schema, name)
                    && a.params.len() == types.len()
                    && a.params.iter().zip(types).all(|(p, t)| p == t || *t == crate::oid::UNKNOWN)
            })
            .cloned()
    })
}

/// run folds the arguments of a group's rows through an aggregate's routines.
pub fn run(ctx: &mut Ctx<'_>, aggregate: &UserAggregate, rows: Vec<Vec<Value>>) -> Result<Value> {
    let mut state = match &aggregate.init_cond {
        Some(text) => crate::cast::cast_value(Value::Text(text.clone()), aggregate.state_type, false)?,
        None => Value::Null,
    };
    for row in rows {
        if aggregate.transition.strict && row.iter().any(Value::is_null) {
            continue;
        }
        if aggregate.transition.strict && state.is_null() {
            state = row.into_iter().next().unwrap_or(Value::Null);
            continue;
        }
        let mut args = vec![state];
        args.extend(row);
        state = crate::routines::call(ctx, &aggregate.transition, args)?;
    }
    match &aggregate.final_routine {
        Some(routine) => crate::routines::call(ctx, routine, vec![state]),
        None => Ok(state),
    }
}

impl Ctx<'_> {
    /// user_aggregates returns the aggregates of the working root, reusing them while their collections are unchanged.
    pub fn user_aggregates(&mut self) -> Result<Arc<Vec<Arc<UserAggregate>>>> {
        let address = (self.txn.root.root_objects[COLLECTION], self.txn.root.root_objects[crate::routines::COLLECTION]);
        if let Some((cached, aggregates)) = &self.session.aggregates
            && *cached == address
        {
            return Ok(aggregates.clone());
        }
        let mut aggregates = Vec::new();
        if address.0.is_some() {
            let routines = self.routines()?;
            let routine = |function: &[u8]| routines.iter().find(|r| r.object.id == function).cloned();
            for (_, address) in self.txn.root.objects(self.db, COLLECTION)? {
                let stored = objects::Aggregate::deserialize(&prolly::read_blob(self.db, &address)?)?;
                let Some(transition) = routine(&stored.s_func) else { continue };
                let mut segments = id::segments(&stored.id).into_iter();
                let (schema, name) = (segments.next().unwrap_or_default(), segments.next().unwrap_or_default());
                let type_of = |type_id: &[u8]| ColumnType { oid: crate::usertypes::type_oid(type_id), modifier: -1 };
                aggregates.push(Arc::new(UserAggregate {
                    schema,
                    name,
                    params: segments.map(|t| crate::usertypes::type_oid(t.as_bytes())).collect(),
                    ret: type_of(&stored.return_type),
                    state_type: type_of(&stored.s_type),
                    transition,
                    final_routine: routine(&stored.final_func),
                    init_cond: stored.has_init_cond.then(|| String::from_utf8_lossy(&stored.init_cond).into_owned()),
                    stored,
                }));
            }
        }
        let aggregates = Arc::new(aggregates);
        self.session.aggregates = Some((address, aggregates.clone()));
        Ok(aggregates)
    }

    /// install_aggregates makes the working root's aggregates known to this thread's statement.
    pub fn install_aggregates(&mut self) -> Result<()> {
        let aggregates = self.user_aggregates()?;
        INSTALLED.with(|installed| *installed.borrow_mut() = aggregates);
        Ok(())
    }
}

/// AggregateFunctions are the names CREATE AGGREGATE gives for an aggregate's routines.
#[derive(Default)]
struct AggregateFunctions {
    transition: Vec<String>,
    final_function: Vec<String>,
    combine: Vec<String>,
}

/// def_names returns the names of a CREATE AGGREGATE option's value, written as a type name or a string.
fn def_names(arg: Option<&pg_query::protobuf::Node>) -> Vec<String> {
    match arg.and_then(|a| a.node.as_ref()) {
        Some(NodeEnum::TypeName(name)) => {
            name.names.iter().filter_map(crate::expr::node_name).map(str::to_string).collect()
        }
        Some(NodeEnum::String(s)) => vec![s.sval.clone()],
        _ => Vec::new(),
    }
}

/// def_text returns the text of a CREATE AGGREGATE option's value.
fn def_text(arg: Option<&pg_query::protobuf::Node>) -> String {
    match arg.and_then(|a| a.node.as_ref()) {
        Some(NodeEnum::String(s)) => s.sval.clone(),
        Some(NodeEnum::Integer(i)) => i.ival.to_string(),
        Some(NodeEnum::Float(f)) => f.fval.clone(),
        _ => def_names(arg).join("."),
    }
}

impl Ctx<'_> {
    /// aggregate_function returns the routine of an aggregate's support function, which takes exactly the types and,
    /// when asked, returns the state type, as Postgres' lookup_agg_function checks.
    fn aggregate_function(
        &mut self,
        names: &[String],
        inputs: &[ColumnType],
        returns: Option<(&str, ColumnType)>,
    ) -> Result<Arc<Routine>> {
        let (schema, name) = match names {
            [.., schema, name] => (Some(schema.as_str()), name.as_str()),
            [name] => (None, name.as_str()),
            [] => return Err(PgError::internal("an aggregate function without a name")),
        };
        let found = self
            .routines_named(schema, name)?
            .into_iter()
            .find(|r| !r.procedure && r.inputs().map(|p| p.ty.oid).eq(inputs.iter().map(|t| t.oid)));
        let shown: Vec<_> = inputs.iter().map(|t| crate::cast::type_display(t.oid)).collect();
        let routine = found.ok_or_else(|| {
            PgError::new(
                code::UNDEFINED_FUNCTION,
                format!("function {}({}) does not exist", names.join("."), shown.join(", ")),
            )
        })?;
        if let Some((kind, state)) = returns
            && routine.ret.oid != state.oid
        {
            return Err(PgError::new(
                code::DATATYPE_MISMATCH,
                format!(
                    "return type of {kind} function {} is not {}",
                    names.join("."),
                    crate::cast::type_display(state.oid)
                ),
            ));
        }
        Ok(routine)
    }

    /// create_aggregate runs CREATE AGGREGATE, checking the aggregate as Postgres' DefineAggregate does, and stores
    /// it as Go stores an aggregate.
    pub fn create_aggregate(&mut self, define: &DefineStmt) -> Result<Outcome> {
        let (named_schema, name) = crate::routines::function_names(&define.defnames);
        let schema = self.target_schema(&named_schema, -1)?;
        let mut params = Vec::new();
        if let Some(NodeEnum::List(list)) = define.args.first().and_then(|a| a.node.as_ref()) {
            for item in &list.items {
                if let Some(NodeEnum::FunctionParameter(param)) = item.node.as_ref()
                    && let Some(type_name) = &param.arg_type
                {
                    self.prepare_type(type_name)?;
                    params.push(crate::expr::resolve_type_name(type_name)?);
                }
            }
        }
        let mut functions = AggregateFunctions::default();
        let mut state = None;
        let mut init_cond = None;
        for item in &define.definition {
            let Some(NodeEnum::DefElem(def)) = item.node.as_ref() else { continue };
            let arg = def.arg.as_deref();
            match def.defname.to_lowercase().as_str() {
                "sfunc" | "sfunc1" => functions.transition = def_names(arg),
                "finalfunc" | "finalfunc1" => functions.final_function = def_names(arg),
                "combinefunc" => functions.combine = def_names(arg),
                "stype" | "stype1" => {
                    if let Some(NodeEnum::TypeName(type_name)) = arg.and_then(|a| a.node.as_ref()) {
                        self.prepare_type(type_name)?;
                        state = Some(crate::expr::resolve_type_name(type_name)?);
                    }
                }
                "initcond" | "initcond1" => init_cond = Some(def_text(arg)),
                other => self.session.notice(PgError {
                    severity: "WARNING",
                    ..PgError::new(code::SYNTAX_ERROR, format!("aggregate attribute \"{other}\" not recognized"))
                }),
            }
        }
        let state = state
            .ok_or_else(|| PgError::new(code::INVALID_FUNCTION_DEFINITION, "aggregate stype must be specified"))?;
        if functions.transition.is_empty() {
            return Err(PgError::new(code::INVALID_FUNCTION_DEFINITION, "aggregate sfunc must be specified"));
        }
        let mut inputs = vec![state];
        inputs.extend(&params);
        let transition = self.aggregate_function(&functions.transition, &inputs, Some(("transition", state)))?;
        if transition.strict && init_cond.is_none() && (params.len() != 1 || params[0].oid != state.oid) {
            return Err(PgError::new(
                code::INVALID_FUNCTION_DEFINITION,
                "must not omit initial value when transition function is strict and transition type is not compatible with input type",
            ));
        }
        let mut ret = state;
        let mut final_function = Vec::new();
        if !functions.final_function.is_empty() {
            let routine = self.aggregate_function(&functions.final_function, &[state], None)?;
            ret = routine.ret;
            final_function = routine.object.id.clone();
        }
        let mut combine = Vec::new();
        if !functions.combine.is_empty() {
            combine = self
                .aggregate_function(&functions.combine, &[state, state], Some(("combine", state)))?
                .object
                .id
                .clone();
        }
        let aggregate_id = crate::routines::function_id(&schema, &name, &params, false);
        let duplicate = || {
            PgError::new(
                code::DUPLICATE_FUNCTION,
                format!("function \"{name}\" already exists with same argument types"),
            )
        };
        if self.routines()?.iter().any(|r| r.object.id == aggregate_id) {
            if define.replace {
                return Err(PgError {
                    detail: Some(format!("\"{name}\" is a function.")),
                    ..PgError::new(code::WRONG_OBJECT_TYPE, "cannot change routine kind")
                });
            }
            return Err(duplicate());
        }
        if self.txn.root.objects(self.db, COLLECTION)?.iter().any(|(key, _)| *key == aggregate_id) && !define.replace {
            return Err(duplicate());
        }
        let aggregate = objects::Aggregate {
            id: aggregate_id,
            return_type: crate::usertypes::type_id(ret.oid),
            s_func: transition.object.id.clone(),
            s_type: crate::usertypes::type_id(state.oid),
            final_func: final_function,
            combine_func: combine,
            has_init_cond: init_cond.is_some(),
            init_cond: init_cond.unwrap_or_default().into_bytes(),
        };
        let data = aggregate.serialize();
        let db = &mut *self.db;
        let mut sink = |_: store::Hash, bytes: &[u8]| {
            db.write_value(bytes.to_vec()).map(|_| ()).map_err(|e| store::Error::Corrupt(e.to_string()))
        };
        let (address, _) =
            prolly::write_blob(&data, &mut sink)?.ok_or_else(|| PgError::internal("an empty aggregate"))?;
        self.txn.root.put_object(self.db, COLLECTION, &aggregate.id, Some(address))?;
        Ok(Outcome::command("CREATE AGGREGATE"))
    }

    /// drop_aggregates runs DROP AGGREGATE.
    pub fn drop_aggregates(&mut self, drop: &DropStmt) -> Result<Outcome> {
        for object in &drop.objects {
            let Some(NodeEnum::ObjectWithArgs(target)) = object.node.as_ref() else { continue };
            let (schema, name, types) = self.aggregate_signature(target)?;
            let found = self
                .user_aggregates()?
                .iter()
                .find(|a| a.name == name && schema.as_ref().is_none_or(|s| *s == a.schema) && a.params == types)
                .map(|a| a.stored.id.clone());
            match found {
                Some(key) => self.txn.root.put_object(self.db, COLLECTION, &key, None)?,
                None => {
                    let shown = match &schema {
                        Some(schema) => format!("{schema}.{name}"),
                        None => name,
                    };
                    if !drop.missing_ok {
                        let types: Vec<_> = types.iter().map(|&t| crate::cast::type_display(t)).collect();
                        return Err(PgError::new(
                            code::UNDEFINED_FUNCTION,
                            format!("aggregate {shown}({}) does not exist", types.join(", ")),
                        ));
                    }
                    let written: Vec<String> = target
                        .objargs
                        .iter()
                        .filter_map(|a| match a.node.as_ref() {
                            Some(NodeEnum::TypeName(t)) => {
                                Some(t.names.iter().filter_map(crate::expr::node_name).collect::<Vec<_>>().join("."))
                            }
                            _ => None,
                        })
                        .collect();
                    self.session.notice(PgError::notice(
                        "00000",
                        format!("aggregate {shown}({}) does not exist, skipping", written.join(",")),
                    ));
                }
            }
        }
        Ok(Outcome::command("DROP AGGREGATE"))
    }

    /// aggregate_signature returns the schema, name, and argument types that DROP AGGREGATE names.
    fn aggregate_signature(&mut self, target: &ObjectWithArgs) -> Result<(Option<String>, String, Vec<u32>)> {
        let (schema, name) = crate::routines::function_names(&target.objname);
        let mut types = Vec::new();
        for arg in &target.objargs {
            if let Some(NodeEnum::TypeName(type_name)) = arg.node.as_ref() {
                self.prepare_type(type_name)?;
                types.push(crate::expr::resolve_type_name(type_name)?.oid);
            }
        }
        Ok(((!schema.is_empty()).then_some(schema), name, types))
    }
}
