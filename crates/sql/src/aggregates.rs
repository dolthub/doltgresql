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

//! The aggregates of a root value's aggregate collection, which emulated extensions create.

use std::cell::RefCell;
use std::sync::Arc;

use crate::catalog::{ColumnType, id};
use crate::error::Result;
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
}

/// AggregateCache is the aggregates of a root value, with the addresses of the aggregate and function collections.
pub type AggregateCache = ((Option<store::Hash>, Option<store::Hash>), Arc<Vec<Arc<UserAggregate>>>);

thread_local! {
    /// INSTALLED is the aggregates of the running statement's root value.
    static INSTALLED: RefCell<Arc<Vec<Arc<UserAggregate>>>> = RefCell::new(Arc::new(Vec::new()));
}

/// exists reports whether an installed aggregate has the name.
pub fn exists(name: &str) -> bool {
    INSTALLED.with(|aggregates| aggregates.borrow().iter().any(|a| a.name == name))
}

/// find returns the installed aggregate of the name for arguments of the types, where an untyped argument matches any
/// parameter.
pub fn find(name: &str, types: &[u32]) -> Option<Arc<UserAggregate>> {
    INSTALLED.with(|aggregates| {
        aggregates
            .borrow()
            .iter()
            .find(|a| {
                a.name == name
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
