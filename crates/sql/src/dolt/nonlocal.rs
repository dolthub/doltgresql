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

//! Dolt's dolt_nonlocal_tables table: rules that make table names resolve to the tables of another branch or revision,
//! as Dolt's getNonlocalTableEntry reads them.

use doltdb::root::Root;
use pg_query::protobuf::RangeVar;

use crate::catalog::table::{ColumnDef, Primary, TableDef};
use crate::dolt::args::error;
use crate::error::Result;
use crate::query::Ctx;
use crate::types::Value;

/// SCHEMA is the schema that holds the rules table.
pub const SCHEMA: &str = "dolt";

/// TABLE is the name of the table that holds the rules.
pub const TABLE: &str = "dolt_nonlocal_tables";

/// COLUMNS are the names of the rules table's columns, the first of which is its primary key.
const COLUMNS: [&str; 4] = ["table_name", "target_ref", "ref_table", "options"];

/// COLUMN_TYPE is the MySQL type that Dolt gives each column of the rules table.
const COLUMN_TYPE: &str = "varchar(65535)";

/// Rule is a rule of the table: a pattern of table names, the branch or revision they resolve on, the name of the
/// table they resolve to there, and the rule's options.
pub struct Rule {
    pub pattern: String,
    pub target: String,
    pub table: String,
    pub options: String,
}

/// rules returns the rules in a root in the table's key order, which has none without a rules table.
pub fn rules(ctx: &mut Ctx<'_>, root: &Root) -> Result<Vec<Rule>> {
    let Some(address) = root.table(ctx.db, SCHEMA, TABLE)? else { return Ok(Vec::new()) };
    let table = TableDef::load(ctx.db, SCHEMA, TABLE, address)?;
    let text = |value: Option<&Value>| match value {
        Some(Value::Text(text)) => text.clone(),
        _ => String::new(),
    };
    Ok(crate::query::scan(ctx.db, &table)?
        .into_iter()
        .map(|row| Rule {
            pattern: text(row.first()),
            target: text(row.get(1)),
            table: text(row.get(2)),
            options: text(row.get(3)),
        })
        .collect())
}

/// table returns the rules table, creating it with Dolt's MySQL columns when a write needs it, as Dolt does.
pub fn table(ctx: &mut Ctx<'_>) -> Result<TableDef> {
    if let Some(table) = ctx.txn.table(ctx.db, SCHEMA, TABLE)? {
        return Ok(table);
    }
    let columns = COLUMNS
        .iter()
        .enumerate()
        .map(|(i, name)| ColumnDef {
            name: name.to_string(),
            ty: crate::expr::typ(crate::oid::TEXT),
            tag: 0,
            encoding: prolly::val::encoding::STRING,
            nullable: i > 0,
            primary_key: i == 0,
            default: String::new(),
            generated: false,
            mysql_type: COLUMN_TYPE.into(),
        })
        .collect();
    ctx.write_new_table(SCHEMA, TABLE, columns, (vec![0], Primary::default()), Vec::new(), Vec::new())?;
    ctx.txn
        .table(ctx.db, SCHEMA, TABLE)?
        .ok_or_else(|| crate::error::PgError::internal("the new table dolt_nonlocal_tables"))
}

impl Ctx<'_> {
    /// nonlocal_rule returns the first rule whose pattern matches a table name, ignoring case, with an empty target
    /// and table filled in as the session's branch and the name itself.
    pub fn nonlocal_rule(&mut self, name: &str) -> Result<Option<Rule>> {
        let name = name.to_lowercase();
        let root = self.txn.root.clone();
        let Some(mut rule) = rules(self, &root)?
            .into_iter()
            .find(|r| crate::dolt::ignore::matches(&r.pattern.to_lowercase(), &name, false))
        else {
            return Ok(None);
        };
        if rule.target.is_empty() {
            rule.target = self.txn.branch.clone();
        }
        if rule.table.is_empty() {
            rule.table = name;
        }
        Ok(Some(rule))
    }

    /// nonlocal_target returns the branch or revision and the relation that a relation without a database names
    /// under a rule, failing for a rule whose options are not `immediate`, as Dolt's getNonlocalTable does.
    pub fn nonlocal_target(&mut self, relation: &RangeVar) -> Result<Option<(String, RangeVar)>> {
        if !relation.catalogname.is_empty() {
            return Ok(None);
        }
        let Some(rule) = self.nonlocal_rule(&relation.relname)? else { return Ok(None) };
        if rule.options != "immediate" {
            return Err(error(format!(
                "Invalid nonlocal table options {}: only valid value is 'immediate'.",
                rule.options
            )));
        }
        Ok(Some((rule.target, RangeVar { relname: rule.table, ..relation.clone() })))
    }

    /// nonlocal_table loads the table that a relation names under a rule: from the working root of a branch, which the
    /// session's transaction on it holds, or from a revision, or None when no rule matches or that table is missing.
    pub fn nonlocal_table(&mut self, relation: &RangeVar) -> Result<Option<TableDef>> {
        let Some((target, renamed)) = self.nonlocal_target(relation)? else { return Ok(None) };
        let root = match target == self.txn.branch {
            true => Some(self.txn.root.clone()),
            false => match self.branch_root(&target)? {
                Some(root) => Some(root),
                None => self.revision_root(&target).ok().flatten(),
            },
        };
        match root {
            Some(root) => Ok(self.resolve_table_in(&renamed, &root).ok()),
            None => Ok(None),
        }
    }

    /// nonlocal_tables returns the tables that rules give names to in a schema, under those names, leaving out names
    /// whose rule fails, as Dolt's getNonlocalTableNames finds them.
    pub fn nonlocal_tables(&mut self, schema: &str) -> Result<Vec<TableDef>> {
        let root = self.txn.root.clone();
        let mut names = Vec::new();
        for rule in rules(self, &root)? {
            let pattern = rule.pattern.to_lowercase();
            let target = match rule.target.as_str() {
                "" => Some(root.clone()),
                target if target == self.txn.branch => Some(root.clone()),
                target => match self.branch_root(target)? {
                    Some(root) => Some(root),
                    None => self.revision_root(target).ok().flatten(),
                },
            };
            let Some(target) = target else { continue };
            if !rule.table.is_empty() {
                names.push(pattern);
                continue;
            }
            for (key, _) in target.tables(self.db)? {
                let text = String::from_utf8_lossy(&key).into_owned();
                let mut parts = text.splitn(3, '\0').skip(1);
                if parts.next() == Some(schema)
                    && let Some(name) = parts.next()
                    && crate::dolt::ignore::matches(&pattern, name, false)
                {
                    names.push(name.to_string());
                }
            }
        }
        let mut tables = Vec::new();
        for name in names {
            let relation = RangeVar { schemaname: schema.to_string(), relname: name.clone(), ..RangeVar::default() };
            if let Ok(Some(table)) = self.nonlocal_table(&relation) {
                tables.push(TableDef { name, ..table });
            }
        }
        Ok(tables)
    }

    /// check_nonlocal_name fails when a new table's name matches a rule that resolves it to a table of another branch
    /// or name, as Dolt's checkNonlocalTableName does.
    pub fn check_nonlocal_name(&mut self, name: &str) -> Result<()> {
        match self.nonlocal_rule(name)? {
            Some(rule) if rule.target != self.txn.branch || !rule.table.eq_ignore_ascii_case(name) => Err(error(
                format!("Cannot create table name {name} because it matches a name present in dolt_nonlocal_tables."),
            )),
            _ => Ok(()),
        }
    }
}
