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

//! SQL cursors: DECLARE, FETCH, MOVE, and CLOSE, over the rows of a query that DECLARE runs.

use pg_query::protobuf::{ClosePortalStmt, DeclareCursorStmt, FetchDirection, FetchStmt};

use crate::error::{PgError, Result, code};
use crate::query::Ctx;
use crate::types::Value;
use crate::{Column, Outcome};

/// NO_SCROLL and HOLD are the bits of a DECLARE's options that forbid fetching backward and keep the cursor open
/// past its transaction.
const NO_SCROLL: i32 = 0x4;
const HOLD: i32 = 0x20;

/// Cursor is an open cursor: its rows, its position among them, where 0 is before the first row and one past the
/// last row is after it, whether it may fetch backward and outlive its transaction, and whether its transaction is
/// still open.
#[derive(Clone, Debug)]
pub struct Cursor {
    name: String,
    columns: Vec<Column>,
    rows: Vec<Vec<Value>>,
    position: usize,
    no_scroll: bool,
    hold: bool,
    new: bool,
}

/// end_transaction closes the cursors that end with a transaction: every cursor without HOLD, and the held ones that
/// a rolled back transaction declared.
pub fn end_transaction(cursors: &mut Vec<Cursor>, committed: bool) {
    cursors.retain(|c| c.hold && (committed || !c.new));
    for cursor in cursors {
        cursor.new = false;
    }
}

/// columns returns the columns that a FETCH returns, which are its cursor's, or None for a MOVE or a missing cursor.
pub fn columns(cursors: &[Cursor], stmt: &FetchStmt) -> Option<Vec<Column>> {
    let cursor = cursors.iter().find(|c| c.name == stmt.portalname).filter(|_| !stmt.ismove)?;
    Some(cursor.columns.clone())
}

impl Ctx<'_> {
    /// declare_cursor runs DECLARE, which runs the cursor's query.
    pub fn declare_cursor(&mut self, stmt: &DeclareCursorStmt) -> Result<Outcome> {
        let hold = stmt.options & HOLD != 0;
        if !hold && !self.session.explicit && !self.session.implicit_block {
            return Err(PgError::new(
                code::NO_ACTIVE_SQL_TRANSACTION,
                "DECLARE CURSOR can only be used in transaction blocks",
            ));
        }
        if self.session.cursors.iter().any(|c| c.name == stmt.portalname) {
            return Err(PgError::new(code::DUPLICATE_CURSOR, format!("cursor \"{}\" already exists", stmt.portalname)));
        }
        let query = stmt.query.as_deref().and_then(|q| q.node.as_ref());
        let Some(query) = query else { return Err(PgError::internal("a cursor without a query")) };
        let (columns, rows) = match self.run(query)? {
            Outcome::Rows { columns, rows, .. } => (columns, rows),
            _ => (Vec::new(), Vec::new()),
        };
        self.session.cursors.push(Cursor {
            name: stmt.portalname.clone(),
            columns,
            rows,
            position: 0,
            no_scroll: stmt.options & NO_SCROLL != 0,
            hold,
            new: true,
        });
        Ok(Outcome::command("DECLARE CURSOR"))
    }

    /// fetch runs FETCH, which returns rows from a cursor's position, or MOVE, which only moves it, as Postgres'
    /// DoPortalRunFetch does.
    pub fn fetch(&mut self, stmt: &FetchStmt) -> Result<Outcome> {
        let Some(cursor) = self.session.cursors.iter_mut().find(|c| c.name == stmt.portalname) else {
            return Err(PgError::new(
                code::INVALID_CURSOR_NAME,
                format!("cursor \"{}\" does not exist", stmt.portalname),
            ));
        };
        let n = cursor.rows.len() as i64;
        let p = cursor.position as i64;
        let count = stmt.how_many;
        let direction = FetchDirection::try_from(stmt.direction).unwrap_or(FetchDirection::FetchForward);
        let (target, step): (i64, Option<i64>) = match direction {
            FetchDirection::FetchBackward => (p.saturating_sub(count).max(0), Some(-1)),
            FetchDirection::FetchAbsolute if count >= 0 => (count.min(n + 1), None),
            FetchDirection::FetchAbsolute => ((n + 1 + count).max(0), None),
            FetchDirection::FetchRelative => (p.saturating_add(count).clamp(0, n + 1), None),
            _ => (p.saturating_add(count).min(n + 1), Some(1)),
        };
        if cursor.no_scroll && (target < p || step == Some(-1) && count > 0) {
            return Err(PgError {
                hint: Some("Declare it with SCROLL option to enable backward scan.".into()),
                ..PgError::new(code::OBJECT_NOT_IN_PREREQUISITE_STATE, "cursor can only scan forward")
            });
        }
        let on_row = |position: i64| (1..=n).contains(&position).then(|| position as usize - 1);
        let picked: Vec<usize> = match step {
            Some(1) => (p + 1..=target.min(n)).filter_map(on_row).collect(),
            Some(_) => (target.max(1)..p).rev().filter_map(on_row).collect(),
            None => on_row(target).into_iter().collect(),
        };
        cursor.position = target as usize;
        let tag = if stmt.ismove { "MOVE" } else { "FETCH" };
        if stmt.ismove {
            return Ok(Outcome::command(format!("{tag} {}", picked.len())));
        }
        let rows: Vec<Vec<Value>> = picked.iter().map(|&i| cursor.rows[i].clone()).collect();
        Ok(Outcome::Rows { columns: cursor.columns.clone(), tag: format!("{tag} {}", rows.len()), rows })
    }

    /// close_cursor runs CLOSE, which closes a cursor, or every cursor for CLOSE ALL.
    pub fn close_cursor(&mut self, stmt: &ClosePortalStmt) -> Result<Outcome> {
        if stmt.portalname.is_empty() {
            self.session.cursors.clear();
            return Ok(Outcome::command("CLOSE CURSOR ALL"));
        }
        let before = self.session.cursors.len();
        self.session.cursors.retain(|c| c.name != stmt.portalname);
        if self.session.cursors.len() == before {
            return Err(PgError::new(
                code::INVALID_CURSOR_NAME,
                format!("cursor \"{}\" does not exist", stmt.portalname),
            ));
        }
        Ok(Outcome::command("CLOSE CURSOR"))
    }
}
