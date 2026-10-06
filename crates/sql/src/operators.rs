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

//! The operators of a root value's operator collection, which emulated extensions create.

use std::sync::Arc;

use crate::catalog::id;
use crate::error::Result;
use crate::query::Ctx;
use crate::routines::Routine;

/// COLLECTION is the position of the operators in a root value's root object collections.
pub const COLLECTION: usize = 8;

/// UserOperator is an operator of the operator collection, with the routine that computes it.
pub struct UserOperator {
    pub schema: String,
    pub name: String,
    pub left: u32,
    pub right: u32,
    pub stored: objects::Operator,
    pub routine: Arc<Routine>,
}

/// OperatorCache is the operators of a root value, with the addresses of the operator and function collections.
pub type OperatorCache = ((Option<store::Hash>, Option<store::Hash>), Arc<Vec<Arc<UserOperator>>>);

impl Ctx<'_> {
    /// user_operators returns the operators of the working root, reusing them while their collections are unchanged.
    pub fn user_operators(&mut self) -> Result<Arc<Vec<Arc<UserOperator>>>> {
        let address = (self.txn.root.root_objects[COLLECTION], self.txn.root.root_objects[crate::routines::COLLECTION]);
        if let Some((cached, operators)) = &self.session.operators
            && *cached == address
        {
            return Ok(operators.clone());
        }
        let routines = self.routines()?;
        let mut operators = Vec::new();
        for (_, address) in self.txn.root.objects(self.db, COLLECTION)? {
            let stored = objects::Operator::deserialize(&prolly::read_blob(self.db, &address)?)?;
            let Some(routine) = routines.iter().find(|r| r.object.id == stored.function) else { continue };
            let mut segments = id::segments(&stored.id).into_iter();
            let (schema, name) = (segments.next().unwrap_or_default(), segments.next().unwrap_or_default());
            let left = segments.next().map_or(0, |t| crate::usertypes::type_oid(t.as_bytes()));
            let right = segments.next().map_or(0, |t| crate::usertypes::type_oid(t.as_bytes()));
            operators.push(Arc::new(UserOperator { schema, name, left, right, stored, routine: routine.clone() }));
        }
        let operators = Arc::new(operators);
        self.session.operators = Some((address, operators.clone()));
        Ok(operators)
    }
}
