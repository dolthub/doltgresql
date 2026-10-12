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

//! LOCK, ported from Postgres' LockTableCommand: it checks the transaction, the relations, and the privileges that
//! each lock mode needs.

use pg_query::NodeEnum;
use pg_query::protobuf::LockStmt;

use crate::Outcome;
use crate::auth::Object;
use crate::error::{PgError, Result, code};
use crate::query::Ctx;

/// ACCESS_SHARE and ROW_EXCLUSIVE are the lock modes up to which SELECT and INSERT privileges permit a lock.
const ACCESS_SHARE: i32 = 1;
const ROW_EXCLUSIVE: i32 = 3;

impl Ctx<'_> {
    /// lock_tables runs LOCK.
    pub fn lock_tables(&mut self, stmt: &LockStmt) -> Result<Outcome> {
        if !self.session.explicit {
            return Err(PgError::new(
                code::NO_ACTIVE_SQL_TRANSACTION,
                "LOCK TABLE can only be used in transaction blocks",
            ));
        }
        for node in &stmt.relations {
            let Some(NodeEnum::RangeVar(relation)) = node.node.as_ref() else { continue };
            let Some(found) = self.find_relation(relation)? else {
                let shown = match relation.schemaname.as_str() {
                    "" => relation.relname.clone(),
                    schema => format!("{schema}.{}", relation.relname),
                };
                return Err(PgError::new(code::UNDEFINED_TABLE, format!("relation \"{shown}\" does not exist")));
            };
            let unsupported = match found.kind {
                'r' | 'p' | 'v' => None,
                'i' => Some("indexes"),
                'S' => Some("sequences"),
                't' => Some("TOAST tables"),
                _ => Some("relations of this kind"),
            };
            if let Some(kind) = unsupported {
                return Err(PgError {
                    detail: Some(format!("This operation is not supported for {kind}.")),
                    ..PgError::new(code::WRONG_OBJECT_TYPE, format!("cannot lock relation \"{}\"", relation.relname))
                });
            }
            let mut privileges = vec!["m", "w", "d", "D"];
            if stmt.mode <= ACCESS_SHARE {
                privileges.push("r");
            }
            if stmt.mode <= ROW_EXCLUSIVE {
                privileges.push("a");
            }
            let object = Object::Table(found.schema, found.name);
            let mut denied = None;
            for privilege in privileges {
                match self.require(&object, privilege, -1) {
                    Ok(()) => {
                        denied = None;
                        break;
                    }
                    Err(err) => denied = Some(err),
                }
            }
            if let Some(err) = denied {
                return Err(err);
            }
        }
        //TODO: hold the lock until the transaction ends, so that other sessions wait for it.
        Ok(Outcome::command("LOCK TABLE"))
    }
}
