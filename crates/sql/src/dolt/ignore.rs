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

//! Dolt's dolt_ignore table: patterns of table names that staging leaves out, as Dolt's ignore rules match them.

use doltdb::root::Root;

use crate::catalog::table::TableDef;
use crate::dolt::args::error;
use crate::error::Result;
use crate::query::Ctx;
use crate::types::Value;

/// TABLE is the name of the table that holds a schema's ignore patterns.
pub const TABLE: &str = "dolt_ignore";

/// DEFINITION creates the ignore table with the columns Doltgres gives it.
const DEFINITION: &str = "(pattern text PRIMARY KEY, ignored boolean NOT NULL)";

/// Patterns are a schema's ignore patterns, each with whether it ignores the tables it matches.
pub type Patterns = Vec<(String, bool)>;

/// patterns returns the ignore patterns of a schema in a root, which has none without an ignore table.
pub fn patterns(ctx: &mut Ctx<'_>, root: &Root, schema: &str) -> Result<Patterns> {
    let Some(address) = root.table(ctx.db, schema, TABLE)? else { return Ok(Vec::new()) };
    let table = TableDef::load(ctx.db, schema, TABLE, address)?;
    Ok(crate::query::scan(ctx.db, &table)?
        .into_iter()
        .filter_map(|row| match (row.first(), row.get(1)) {
            (Some(Value::Text(pattern)), Some(Value::Bool(ignored))) => Some((pattern.clone(), *ignored)),
            _ => None,
        })
        .collect())
}

/// create creates a schema's ignore table, which Dolt creates on the first write to it.
pub fn create(ctx: &mut Ctx<'_>, schema: &str) -> Result<TableDef> {
    crate::dolt::tables::create_backing(ctx, schema, TABLE, DEFINITION)
}

/// glob reports whether text matches a pattern, where `*` and `%` match any run of characters and `?` matches one
/// character, which may not be `*` or `%` when `strict` is set, as Dolt compares one pattern with another.
fn glob(pattern: &[char], text: &[char], strict: bool) -> bool {
    match pattern.split_first() {
        None => text.is_empty(),
        Some(('*' | '%', rest)) => (0..=text.len()).any(|i| glob(rest, &text[i..], strict)),
        Some(('?', rest)) => {
            text.first().is_some_and(|c| !strict || !matches!(c, '*' | '%')) && glob(rest, &text[1..], strict)
        }
        Some((c, rest)) => text.first() == Some(c) && glob(rest, &text[1..], strict),
    }
}

/// matches reports whether a pattern matches a name, or a more specific pattern when `strict` is set.
fn matches(pattern: &str, name: &str, strict: bool) -> bool {
    glob(&pattern.chars().collect::<Vec<_>>(), &name.chars().collect::<Vec<_>>(), strict)
}

/// normalize writes a pattern so that equivalent patterns are equal, with `%` for `*` and no repeated `%`.
fn normalize(pattern: &str) -> String {
    let mut out = String::new();
    for c in pattern.chars().map(|c| if c == '*' { '%' } else { c }) {
        if !(c == '%' && out.ends_with('%')) {
            out.push(c);
        }
    }
    out
}

/// conflict returns Dolt's error for a table that patterns both ignore and keep.
fn conflict(name: &str, ignoring: &[&str], keeping: &[&str]) -> crate::error::PgError {
    let mut message = format!("the table {name} matches conflicting patterns in dolt_ignore:");
    for pattern in ignoring {
        message.push_str(&format!("\nignored:     {pattern}"));
    }
    for pattern in keeping {
        message.push_str(&format!("\nnot ignored: {pattern}"));
    }
    error(message)
}

/// is_ignored reports whether patterns ignore a table, where a more specific pattern overrides a less specific one
/// and patterns that neither overrides conflict, as Dolt's IsTableNameIgnored decides, and the rebase plan table,
/// `dolt.rebase` or any table named dolt_rebase, is always ignored.
pub fn is_ignored(patterns: &Patterns, schema: &str, name: &str) -> Result<bool> {
    if name.eq_ignore_ascii_case("dolt_rebase") || (schema == "dolt" && name == "rebase") {
        return Ok(true);
    }
    let matching = |ignored: bool| -> Vec<&str> {
        patterns.iter().filter(|(p, i)| *i == ignored && matches(p, name, false)).map(|(p, _)| p.as_str()).collect()
    };
    let (ignoring, keeping) = (matching(true), matching(false));
    if ignoring.is_empty() {
        return Ok(false);
    }
    if keeping.is_empty() {
        return Ok(true);
    }
    for t in &ignoring {
        if let Some(k) = keeping.iter().find(|k| normalize(t) == normalize(k)) {
            return Err(conflict(name, &[t], &[k]));
        }
    }
    let overridden_ignoring: Vec<&str> =
        ignoring.iter().filter(|t| keeping.iter().any(|k| matches(t, k, true))).copied().collect();
    let overridden_keeping = keeping.iter().filter(|k| ignoring.iter().any(|t| matches(k, t, true))).count();
    if overridden_ignoring.len() == ignoring.len() {
        return Ok(false);
    }
    if overridden_keeping == keeping.len() {
        return Ok(true);
    }
    let remaining_ignoring: Vec<&str> = ignoring.iter().filter(|t| !overridden_ignoring.contains(t)).copied().collect();
    let remaining_keeping: Vec<&str> = keeping.iter().filter(|k| !overridden_ignoring.contains(k)).copied().collect();
    Err(conflict(name, &remaining_ignoring, &remaining_keeping))
}
