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

//! The casts of a root value's cast collection, which emulated extensions create.

use std::cell::RefCell;
use std::sync::Arc;

use crate::catalog::id;
use crate::error::Result;
use crate::query::Ctx;
use crate::routines::Routine;

/// COLLECTION is the position of the casts in a root value's root object collections.
pub const COLLECTION: usize = 7;

/// The cast contexts, as Go numbers them.
pub const EXPLICIT: u8 = 0;
pub const ASSIGNMENT: u8 = 1;
pub const IMPLICIT: u8 = 2;

/// UserCast is a cast of the cast collection, with the routine that performs it, or None for a cast through text.
pub struct UserCast {
    pub source: u32,
    pub target: u32,
    pub context: u8,
    pub stored: objects::Cast,
    pub routine: Option<Arc<Routine>>,
}

/// CastCache is the casts of a root value, with the addresses of the cast and function collections.
pub type CastCache = ((Option<store::Hash>, Option<store::Hash>), Arc<Vec<UserCast>>);

thread_local! {
    /// INSTALLED is the casts of the running statement's root value.
    static INSTALLED: RefCell<Arc<Vec<UserCast>>> = RefCell::new(Arc::new(Vec::new()));
}

/// find returns the installed cast from one type to another.
pub fn find(source: u32, target: u32) -> Option<(u8, Option<Arc<Routine>>)> {
    INSTALLED.with(|casts| {
        casts.borrow().iter().find(|c| c.source == source && c.target == target).map(|c| (c.context, c.routine.clone()))
    })
}

/// context returns the context of the installed cast from one type to another.
pub fn context(source: u32, target: u32) -> Option<u8> {
    find(source, target).map(|(context, _)| context)
}

impl Ctx<'_> {
    /// user_casts returns the casts of the working root, reusing them while their collections are unchanged.
    pub fn user_casts(&mut self) -> Result<Arc<Vec<UserCast>>> {
        let address = (self.txn.root.root_objects[COLLECTION], self.txn.root.root_objects[crate::routines::COLLECTION]);
        if let Some((cached, casts)) = &self.session.casts
            && *cached == address
        {
            return Ok(casts.clone());
        }
        let mut casts = Vec::new();
        if address.0.is_some() {
            let routines = self.routines()?;
            for (_, address) in self.txn.root.objects(self.db, COLLECTION)? {
                let stored = objects::Cast::deserialize(&prolly::read_blob(self.db, &address)?)?;
                let segments = id::segments(&stored.id);
                let [source, target] = segments.as_slice() else { continue };
                casts.push(UserCast {
                    source: crate::usertypes::type_oid(source.as_bytes()),
                    target: crate::usertypes::type_oid(target.as_bytes()),
                    context: stored.cast_type,
                    routine: routines.iter().find(|r| r.object.id == stored.function).cloned(),
                    stored,
                });
            }
        }
        let casts = Arc::new(casts);
        self.session.casts = Some((address, casts.clone()));
        Ok(casts)
    }

    /// install_casts makes the working root's casts known to this thread's statement.
    pub fn install_casts(&mut self) -> Result<()> {
        let casts = self.user_casts()?;
        INSTALLED.with(|installed| *installed.borrow_mut() = casts);
        Ok(())
    }
}
