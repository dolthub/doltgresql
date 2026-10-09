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

//! The catalog entries of Dolt's own tables, which the catalogs list when dolt_show_system_tables is on, with the
//! columns, nullability, and indexes that Go's Doltgres gives them.

use crate::catalog::ColumnType;
use crate::catalog::table::{ColumnDef, TableDef};
use crate::dolt::tables::SystemTable;
use crate::oid::{BOOL, INT4, INT8, JSON, NUMERIC, TEXT, TIMESTAMP, VARCHAR};
use crate::pgcatalog::rows::TableIndex;

/// Columns is where a fixed system table's columns come from: its definition, or a list for the tables that Doltgres
/// does not query.
enum Columns {
    System(SystemTable),
    Listed(&'static [(&'static str, u32)]),
}

/// Fixed is a system table that every schema of one kind has: its name, its columns, the columns that may be NULL,
/// and the ID and column of its unique index when it has one.
struct Fixed {
    name: &'static str,
    columns: Columns,
    nullable: &'static [&'static str],
    index: Option<(&'static str, &'static str, bool)>,
}

/// BRANCH_NULLABLE are the columns of the branch tables that may be NULL.
const BRANCH_NULLABLE: &[&str] = &[
    "latest_committer",
    "latest_committer_email",
    "latest_commit_date",
    "latest_commit_message",
    "remote",
    "branch",
    "dirty",
    "latest_author",
    "latest_author_email",
    "latest_author_date",
];

/// STATUS_IGNORED are the columns of dolt_status_ignored.
const STATUS_IGNORED: &[(&str, u32)] = &[("table_name", TEXT), ("staged", BOOL), ("status", TEXT), ("ignored", BOOL)];

/// fixed declares a fixed system table.
const fn fixed(
    name: &'static str,
    columns: Columns,
    nullable: &'static [&'static str],
    index: Option<(&'static str, &'static str, bool)>,
) -> Fixed {
    Fixed { name, columns, nullable, index }
}

/// BRANCH_INDEX and COMMIT_INDEX are the indexes that Dolt's branch and commit tables have.
const BRANCH_INDEX: Option<(&str, &str, bool)> = Some(("dolt_branches_name_idx", "name", true));
const COMMIT_INDEX: Option<(&str, &str, bool)> = Some(("commit_hash", "commit_hash", true));

/// USER_SCHEMA are the fixed system tables of a user schema.
const USER_SCHEMA: &[Fixed] = &[
    fixed("dolt_branches", Columns::System(SystemTable::Branches), BRANCH_NULLABLE, BRANCH_INDEX),
    fixed("dolt_column_diff", Columns::System(SystemTable::ColumnDiff), &[], None),
    fixed("dolt_commit_ancestors", Columns::System(SystemTable::CommitAncestors), &[], COMMIT_INDEX),
    fixed("dolt_commits", Columns::System(SystemTable::Commits), &[], COMMIT_INDEX),
    fixed("dolt_conflicts", Columns::System(SystemTable::Conflicts), &[], None),
    fixed("dolt_constraint_violations", Columns::System(SystemTable::ConstraintViolations), &[], None),
    fixed("dolt_diff", Columns::System(SystemTable::Diff), &[], Some(("commit_hash", "commit_hash", false))),
    fixed("dolt_log", Columns::System(SystemTable::Log), &["parents", "signature"], COMMIT_INDEX),
    fixed(
        "dolt_merge_status",
        Columns::System(SystemTable::MergeStatus),
        &["source", "source_commit", "target", "unmerged_tables"],
        None,
    ),
    fixed("dolt_remote_branches", Columns::System(SystemTable::RemoteBranches), BRANCH_NULLABLE, BRANCH_INDEX),
    fixed("dolt_remotes", Columns::System(SystemTable::Remotes), &["fetch_specs", "params"], None),
    fixed("dolt_schema_conflicts", Columns::System(SystemTable::SchemaConflicts), &[], None),
    fixed("dolt_status", Columns::System(SystemTable::Status), &[], None),
    fixed("dolt_status_ignored", Columns::Listed(STATUS_IGNORED), &[], None),
    fixed("dolt_tags", Columns::System(SystemTable::Tags), &[], Some(("dolt_tags_name_idx", "tag_name", true))),
];

/// DOLT_SCHEMA are the system tables of the `dolt` schema.
const DOLT_SCHEMA: &[Fixed] = &[
    fixed("branches", Columns::System(SystemTable::Branches), BRANCH_NULLABLE, BRANCH_INDEX),
    fixed("commit_ancestors", Columns::System(SystemTable::CommitAncestors), &[], COMMIT_INDEX),
    fixed("commits", Columns::System(SystemTable::Commits), &[], COMMIT_INDEX),
    fixed("conflicts", Columns::System(SystemTable::Conflicts), &[], None),
    fixed("constraint_violations", Columns::System(SystemTable::ConstraintViolations), &[], None),
    fixed("dolt_backups", Columns::System(SystemTable::Backups), &[], None),
    fixed(
        "dolt_branch_activity",
        Columns::Listed(&[
            ("branch", TEXT),
            ("last_read", TIMESTAMP),
            ("last_write", TIMESTAMP),
            ("active_sessions", INT4),
            ("system_start_time", TIMESTAMP),
        ]),
        &["last_read", "last_write"],
        None,
    ),
    fixed("dolt_help", Columns::System(SystemTable::Help), &[], None),
    fixed(
        "dolt_stashes",
        Columns::Listed(&[
            ("name", TEXT),
            ("stash_id", TEXT),
            ("branch", TEXT),
            ("hash", TEXT),
            ("commit_message", TEXT),
        ]),
        &["commit_message"],
        None,
    ),
    fixed("log", Columns::System(SystemTable::Log), &["parents", "signature"], COMMIT_INDEX),
    fixed("remote_branches", Columns::System(SystemTable::RemoteBranches), BRANCH_NULLABLE, BRANCH_INDEX),
    fixed("remotes", Columns::System(SystemTable::Remotes), &["fetch_specs", "params"], None),
    fixed("status", Columns::System(SystemTable::Status), &[], None),
    fixed("status_ignored", Columns::Listed(STATUS_IGNORED), &[], None),
];

/// Column is a generated column: its name, its type, and whether it may be NULL.
type Column = (String, ColumnType, bool);

/// typ returns a column type without a modifier.
fn typ(oid: u32) -> ColumnType {
    ColumnType { oid, modifier: -1 }
}

/// table makes the definition of a system table with its columns.
fn table(schema: &str, name: &str, columns: Vec<Column>) -> TableDef {
    let columns = columns
        .into_iter()
        .map(|(name, ty, nullable)| ColumnDef {
            name,
            encoding: ty.encoding(),
            ty,
            tag: 0,
            nullable,
            primary_key: false,
            default: String::new(),
            generated: false,
            mysql_type: String::new(),
            comment: String::new(),
            identity: 0,
            legacy_array: false,
        })
        .collect();
    TableDef {
        schema: schema.to_string(),
        name: name.to_string(),
        primary: Default::default(),
        columns,
        hidden: Vec::new(),
        checks: Vec::new(),
        indexes: Vec::new(),
        key_columns: Vec::new(),
        value_columns: Vec::new(),
        comment: String::new(),
        table: doltdb::table::Table {
            schema: Default::default(),
            primary_index: Vec::new(),
            secondary_indexes: Vec::new(),
            auto_increment: 0,
            conflicts: Default::default(),
            violations: Vec::new(),
            artifacts: Vec::new(),
        },
    }
}

/// index makes an index of a system table over columns of the table by name, which Dolt's own indexes order with
/// NULLs first, under the ID that Go gives it and the name it shows.
fn index(
    table: &TableDef,
    id: &str,
    name: String,
    columns: &[String],
    unique: bool,
    virtual_index: bool,
) -> TableIndex {
    let positions: Vec<usize> =
        columns.iter().filter_map(|c| table.columns.iter().position(|t| t.name == *c)).collect();
    TableIndex {
        id: id.to_string(),
        name,
        unique,
        primary: id == "PRIMARY",
        descending: vec![false; positions.len()],
        nulls_first: vec![virtual_index; positions.len()],
        vector: None,
        deferrable: false,
        initially_deferred: false,
        names: columns.to_vec(),
        op_classes: vec![String::new(); positions.len()],
        predicate: String::new(),
        columns: positions,
        plain: false,
    }
}

/// fixed_tables makes the fixed system tables of a schema.
fn fixed_tables(schema: &str, list: &[Fixed]) -> Vec<(TableDef, Vec<TableIndex>)> {
    let mut out = Vec::with_capacity(list.len());
    for fixed in list {
        let columns: Vec<(String, ColumnType)> = match &fixed.columns {
            Columns::System(system) => system.columns(),
            Columns::Listed(columns) => columns.iter().map(|(n, t)| (n.to_string(), typ(*t))).collect(),
        };
        let columns = columns
            .into_iter()
            .map(|(name, ty)| {
                let nullable = fixed.nullable.contains(&name.as_str());
                (name, ty, nullable)
            })
            .collect();
        let def = table(schema, fixed.name, columns);
        let indexes = match fixed.index {
            Some((id, column, unique)) => {
                let name = format!("{}_{id}_key", fixed.name);
                vec![index(&def, id, name, &[column.to_string()], unique, true)]
            }
            None => Vec::new(),
        };
        out.push((def, indexes));
    }
    out
}

/// prefixed returns a table's columns with a prefix on each name, all of which may be NULL.
fn prefixed(user: &TableDef, prefix: &str) -> Vec<Column> {
    user.columns.iter().map(|c| (format!("{prefix}{}", c.name), c.ty, true)).collect()
}

/// key_names returns the names of a table's primary key columns with a prefix, in key order.
fn key_names(user: &TableDef, prefix: &str) -> Vec<String> {
    user.key_columns.iter().map(|&k| format!("{prefix}{}", user.columns[k].name)).collect()
}

/// diff_columns returns the columns of a table's diff and commit diff tables.
fn diff_columns(user: &TableDef) -> Vec<Column> {
    let mut columns = prefixed(user, "to_");
    columns.push(("to_commit".into(), typ(TEXT), true));
    columns.push(("to_commit_date".into(), typ(TIMESTAMP), true));
    columns.extend(prefixed(user, "from_"));
    columns.push(("from_commit".into(), typ(TEXT), true));
    columns.push(("from_commit_date".into(), typ(TIMESTAMP), true));
    columns.push(("diff_type".into(), typ(TEXT), true));
    columns
}

/// per_table makes the system tables that Dolt generates for a stored table.
fn per_table(user: &TableDef) -> Vec<(TableDef, Vec<TableIndex>)> {
    let (schema, name) = (user.schema.as_str(), user.name.as_str());
    let keyed = !user.key_columns.is_empty();
    let is_key = |i: usize| user.key_columns.contains(&i);
    let mut out = Vec::new();

    let commit_diff = table(schema, &format!("dolt_commit_diff_{name}"), diff_columns(user));
    let commits = ["to_commit".to_string(), "from_commit".to_string()];
    let indexes = ["from", "to"]
        .map(|side| {
            let columns = [&commits[..], &key_names(user, &format!("{side}_"))].concat();
            let id = format!("commits_{side}");
            index(&commit_diff, &id, id.clone(), &columns, true, true)
        })
        .into_iter()
        .collect();
    out.push((commit_diff, indexes));

    let mut conflicts = vec![("from_root_ish".to_string(), typ(TEXT), true)];
    conflicts.extend(prefixed(user, "base_"));
    conflicts.extend(user.columns.iter().enumerate().map(|(i, c)| (format!("our_{}", c.name), c.ty, !is_key(i))));
    conflicts.push(("our_diff_type".into(), typ(TEXT), true));
    conflicts.extend(prefixed(user, "their_"));
    conflicts.push(("their_diff_type".into(), typ(TEXT), true));
    conflicts.push(("dolt_conflict_id".into(), typ(TEXT), true));
    if !keyed {
        for side in ["base", "our", "their"] {
            conflicts.push((format!("{side}_cardinality"), typ(NUMERIC), true));
        }
    }
    out.push((table(schema, &format!("dolt_conflicts_{name}"), conflicts), Vec::new()));

    let mut violations = vec![
        ("from_root_ish".to_string(), typ(TEXT), true),
        ("violation_type".to_string(), ColumnType { oid: VARCHAR, modifier: 20 }, false),
    ];
    if !keyed {
        violations.push(("dolt_row_hash".into(), typ(TEXT), false));
    }
    let order = user.key_columns.iter().copied().chain((0..user.columns.len()).filter(|&i| !is_key(i)));
    violations.extend(order.map(|i| (user.columns[i].name.clone(), user.columns[i].ty, !is_key(i))));
    violations.push(("violation_info".into(), typ(JSON), false));
    out.push((table(schema, &format!("dolt_constraint_violations_{name}"), violations), Vec::new()));

    let diff_name = format!("dolt_diff_{name}");
    let diff = table(schema, &diff_name, diff_columns(user));
    let mut indexes = Vec::new();
    if keyed {
        for side in ["from", "to"] {
            let column = format!("{side}_commit");
            indexes.push(index(
                &diff,
                &column,
                format!("{diff_name}_{column}_key"),
                std::slice::from_ref(&column),
                false,
                true,
            ));
        }
        for side in ["from", "to"] {
            let id = format!("{side}_pks");
            indexes.push(index(&diff, &id, id.clone(), &key_names(user, &format!("{side}_")), true, true));
        }
    }
    out.push((diff, indexes));

    let history_name = format!("dolt_history_{name}");
    let mut columns: Vec<Column> = user.columns.iter().map(|c| (c.name.clone(), c.ty, c.nullable)).collect();
    columns.push(("commit_hash".into(), typ(TEXT), false));
    columns.push(("committer".into(), typ(TEXT), false));
    columns.push(("commit_date".into(), typ(TIMESTAMP), false));
    let history = table(schema, &history_name, columns);
    let mut indexes = Vec::new();
    if keyed {
        let key = key_names(user, "");
        indexes.push(index(&history, "PRIMARY", format!("{history_name}_pkey"), &key, false, false));
    }
    let commit_hash = ["commit_hash".to_string()];
    indexes.push(index(&history, "commit_hash", format!("{history_name}_commit_hash_key"), &commit_hash, false, true));
    for secondary in user.indexes.iter().filter(|i| !i.system && i.columns.iter().all(|&c| c < user.columns.len())) {
        let columns: Vec<String> = secondary.columns.iter().map(|&c| user.columns[c].name.clone()).collect();
        indexes.push(index(&history, &secondary.name, secondary.name.clone(), &columns, false, false));
    }
    out.push((history, indexes));

    let mut workspace = vec![
        ("id".to_string(), typ(INT8), false),
        ("staged".to_string(), typ(BOOL), false),
        ("diff_type".to_string(), typ(TEXT), false),
    ];
    workspace.extend(prefixed(user, "to_"));
    workspace.extend(prefixed(user, "from_"));
    out.push((table(schema, &format!("dolt_workspace_{name}"), workspace), Vec::new()));
    out
}

/// generated returns the catalog entries of Dolt's tables in each schema: the `dolt` schema's tables, each user
/// schema's fixed tables, the stored system tables, and the tables that Dolt generates for each stored table.
pub fn generated(
    schemas: &[String],
    stored: &[TableDef],
    stored_system: Vec<TableDef>,
) -> Vec<(TableDef, Vec<TableIndex>)> {
    let mut out = Vec::new();
    for schema in schemas {
        match schema.as_str() {
            "pg_catalog" | "information_schema" => continue,
            "dolt" => out.extend(fixed_tables(schema, DOLT_SCHEMA)),
            _ => out.extend(fixed_tables(schema, USER_SCHEMA)),
        }
    }
    for user in stored.iter().chain(&stored_system) {
        if user.schema != "dolt" {
            out.extend(per_table(user));
        }
    }
    out.extend(stored_system.into_iter().map(|t| (t, Vec::new())));
    out.sort_by(|a, b| (&a.0.schema, &a.0.name).cmp(&(&b.0.schema, &b.0.name)));
    out
}
