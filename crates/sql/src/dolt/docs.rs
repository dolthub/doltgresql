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

//! Dolt's docs table: documents versioned with the database, held in the `docs` table of the `dolt` schema, with
//! Dolt's guide for agents when the database has no AGENT.md of its own.

use crate::catalog::table::TableDef;
use crate::error::Result;
use crate::query::Ctx;
use crate::types::Value;

/// SCHEMA is the schema that holds the docs table.
pub const SCHEMA: &str = "dolt";

/// TABLE is the name of the table that holds the docs.
pub const TABLE: &str = "docs";

/// DEFINITION creates the docs table with the columns Doltgres gives it.
const DEFINITION: &str = "(doc_name text PRIMARY KEY, doc_text text)";

/// AGENT_DOC is the name of the guide for agents.
const AGENT_DOC: &str = "AGENT.md";

/// AGENT_TEXT is Dolt's guide for agents, which the docs show when the database has no AGENT.md.
const AGENT_TEXT: &str = include_str!("AGENT.md");

/// is_docs reports whether a schema and name refer to the docs table: `dolt.docs`, or `dolt_docs` elsewhere.
pub fn is_docs(schema: &str, name: &str) -> bool {
    (schema == SCHEMA && name == TABLE) || name == "dolt_docs"
}

/// table returns the docs table, creating it when a write needs it, as Dolt does.
pub fn table(ctx: &mut Ctx<'_>) -> Result<TableDef> {
    match ctx.txn.table(ctx.db, SCHEMA, TABLE)? {
        Some(table) => Ok(table),
        None => crate::dolt::tables::create_backing(ctx, SCHEMA, TABLE, DEFINITION),
    }
}

/// rows returns the docs, followed by Dolt's guide for agents when the docs lack an AGENT.md.
pub fn rows(ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
    let mut rows = match ctx.txn.table(ctx.db, SCHEMA, TABLE)? {
        Some(table) => crate::query::scan(ctx.db, &table)?,
        None => Vec::new(),
    };
    if !rows.iter().any(|row| row.first() == Some(&Value::Text(AGENT_DOC.into()))) {
        rows.push(vec![Value::Text(AGENT_DOC.into()), Value::Text(AGENT_TEXT.into())]);
    }
    Ok(rows)
}

impl Ctx<'_> {
    /// update_docs runs an UPDATE of the docs table, which can change Dolt's guide for agents while the docs lack an
    /// AGENT.md, storing it as Dolt's docs table does, or returns None when the UPDATE is of another table.
    pub fn update_docs(&mut self, update: &pg_query::protobuf::UpdateStmt) -> Result<Option<crate::Outcome>> {
        let Some(relation) = update.relation.as_ref() else { return Ok(None) };
        if !is_docs(&relation.schemaname, &relation.relname) || self.resolve_table(relation).is_ok() {
            return Ok(None);
        }
        let agent = vec![Value::Text(AGENT_DOC.into()), Value::Text(AGENT_TEXT.into())];
        let docs = table(self)?;
        let stored = crate::query::scan(self.db, &docs)?.iter().any(|row| row.first() == agent.first());
        if !stored {
            crate::dml::write_rows(self, &docs, std::slice::from_ref(&agent))?;
        }
        let outcome = self.plan_update(update)?.run(self)?;
        let docs = table(self)?;
        let untouched = crate::query::scan(self.db, &docs)?.into_iter().find(|row| {
            row.first() == agent.first() && row.get(1).and_then(Value::output).as_deref() == Some(AGENT_TEXT)
        });
        if let Some(row) = untouched.filter(|_| !stored) {
            crate::dml::delete_rows(self, &docs, &[row])?;
        }
        Ok(Some(outcome))
    }
}
