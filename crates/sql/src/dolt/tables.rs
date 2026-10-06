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

//! Dolt's system tables, which show a database's history, branches, and pending changes.

use std::collections::{BTreeMap, HashMap};

use doltdb::root::Root;
use store::Hash;

use crate::dolt::args::{Kind, Parser, error};
use crate::dolt::history::{self, CommitInfo};
use crate::dolt::procedures::table_map;
use crate::error::Result;
use crate::numeric::Numeric;
use crate::oid::{BOOL, INT4, NUMERIC, TEXT, TIMESTAMP};
use crate::query::Ctx;
use crate::txn::read;
use crate::types::Value;

/// JSON is the OID of the json type.
const JSON: u32 = 114;

/// SystemTable is one of Dolt's system tables.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum SystemTable {
    Log,
    Branches,
    RemoteBranches,
    Tags,
    Commits,
    CommitAncestors,
    Status,
    Remotes,
    MergeStatus,
    Conflicts,
    ConstraintViolations,
    SchemaConflicts,
}

/// TABLES are the system tables by their names in the `dolt` schema.
const TABLES: &[(&str, SystemTable)] = &[
    ("log", SystemTable::Log),
    ("branches", SystemTable::Branches),
    ("remote_branches", SystemTable::RemoteBranches),
    ("tags", SystemTable::Tags),
    ("commits", SystemTable::Commits),
    ("commit_ancestors", SystemTable::CommitAncestors),
    ("status", SystemTable::Status),
    ("remotes", SystemTable::Remotes),
    ("merge_status", SystemTable::MergeStatus),
    ("conflicts", SystemTable::Conflicts),
    ("constraint_violations", SystemTable::ConstraintViolations),
    ("schema_conflicts", SystemTable::SchemaConflicts),
];

/// lookup returns the system table that a schema and name refer to: a name in the `dolt` schema, or the name with a
/// `dolt_` prefix elsewhere.
pub fn lookup(schema: &str, name: &str) -> Option<SystemTable> {
    let short = if schema == "dolt" { name } else { name.strip_prefix("dolt_")? };
    TABLES.iter().find(|(n, _)| *n == short).map(|(_, t)| *t)
}

/// BRANCH_COLUMNS are the columns of the branches table.
const BRANCH_COLUMNS: &[(&str, u32)] = &[
    ("name", TEXT),
    ("hash", TEXT),
    ("latest_committer", TEXT),
    ("latest_committer_email", TEXT),
    ("latest_commit_date", TIMESTAMP),
    ("latest_commit_message", TEXT),
    ("remote", TEXT),
    ("branch", TEXT),
    ("dirty", BOOL),
    ("latest_author", TEXT),
    ("latest_author_email", TEXT),
    ("latest_author_date", TIMESTAMP),
];

impl SystemTable {
    /// columns returns the table's column names and types.
    pub fn columns(self) -> Vec<(&'static str, u32)> {
        match self {
            SystemTable::Log => vec![
                ("commit_hash", TEXT),
                ("committer", TEXT),
                ("email", TEXT),
                ("date", TIMESTAMP),
                ("message", TEXT),
                ("commit_order", NUMERIC),
                ("parents", TEXT),
                ("refs", TEXT),
                ("signature", TEXT),
                ("author", TEXT),
                ("author_email", TEXT),
                ("author_date", TIMESTAMP),
            ],
            SystemTable::Branches => BRANCH_COLUMNS.to_vec(),
            SystemTable::RemoteBranches => {
                BRANCH_COLUMNS.iter().filter(|(n, _)| !matches!(*n, "remote" | "branch" | "dirty")).copied().collect()
            }
            SystemTable::Tags => vec![
                ("tag_name", TEXT),
                ("tag_hash", TEXT),
                ("tagger", TEXT),
                ("email", TEXT),
                ("date", TIMESTAMP),
                ("message", TEXT),
            ],
            SystemTable::Commits => vec![
                ("commit_hash", TEXT),
                ("committer", TEXT),
                ("email", TEXT),
                ("date", TIMESTAMP),
                ("message", TEXT),
                ("author", TEXT),
                ("author_email", TEXT),
                ("author_date", TIMESTAMP),
            ],
            SystemTable::CommitAncestors => vec![("commit_hash", TEXT), ("parent_hash", TEXT), ("parent_index", INT4)],
            SystemTable::Status => vec![("table_name", TEXT), ("staged", BOOL), ("status", TEXT)],
            SystemTable::Remotes => vec![("name", TEXT), ("url", TEXT), ("fetch_specs", JSON), ("params", JSON)],
            SystemTable::MergeStatus => vec![
                ("is_merging", BOOL),
                ("source", TEXT),
                ("source_commit", TEXT),
                ("target", TEXT),
                ("unmerged_tables", TEXT),
            ],
            SystemTable::Conflicts => vec![("table", TEXT), ("num_conflicts", NUMERIC)],
            SystemTable::ConstraintViolations => vec![("table", TEXT), ("num_violations", NUMERIC)],
            SystemTable::SchemaConflicts => vec![
                ("table_name", TEXT),
                ("base_schema", TEXT),
                ("our_schema", TEXT),
                ("their_schema", TEXT),
                ("description", TEXT),
            ],
        }
    }

    /// rows returns the table's rows.
    pub fn rows(self, ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
        match self {
            SystemTable::Log => log_rows(ctx, &[ctx.txn.head]),
            SystemTable::Branches => branch_rows(ctx, "refs/heads/", true),
            SystemTable::RemoteBranches => branch_rows(ctx, "refs/remotes/", false),
            SystemTable::Tags => tag_rows(ctx),
            SystemTable::Commits => commit_rows(ctx),
            SystemTable::CommitAncestors => ancestor_rows(ctx),
            SystemTable::Status => status_rows(ctx),
            SystemTable::MergeStatus => merge_status_rows(ctx),
            SystemTable::Remotes
            | SystemTable::Conflicts
            | SystemTable::ConstraintViolations
            | SystemTable::SchemaConflicts => Ok(Vec::new()),
        }
    }
}

/// timestamp converts Unix milliseconds to a timestamp.
fn timestamp(millis: i64) -> Value {
    Value::Timestamp(millis * 1000 + crate::datetime::UNIX_EPOCH_DAYS * crate::datetime::USECS_PER_DAY)
}

/// text returns a text value.
fn text(s: impl Into<String>) -> Value {
    Value::Text(s.into())
}

/// ref_labels returns the labels of the branches, remote branches, and tags at each commit, as the log shows them.
fn ref_labels(ctx: &mut Ctx<'_>) -> Result<HashMap<Hash, Vec<String>>> {
    let mut labels: HashMap<Hash, Vec<String>> = HashMap::new();
    for (name, hash) in history::refs(ctx.db, "refs/heads/")? {
        labels.entry(hash).or_default().push(name);
    }
    for (name, hash) in history::refs(ctx.db, "refs/remotes/")? {
        labels.entry(hash).or_default().push(name);
    }
    for (name, address) in history::refs(ctx.db, "refs/tags/")? {
        let commit = history::commit_of(ctx.db, address)?;
        labels.entry(commit).or_default().push(format!("tag: {name}"));
    }
    Ok(labels)
}

/// log_rows returns the log of the commits reachable from the starting commits.
pub fn log_rows(ctx: &mut Ctx<'_>, starts: &[Hash]) -> Result<Vec<Vec<Value>>> {
    let labels = ref_labels(ctx)?;
    let commits = history::log(ctx.db, starts)?;
    Ok(log_table(ctx, commits, &labels, false))
}

/// log_table returns log rows for the commits, labeled with their refs, with their parents when asked.
fn log_table(
    ctx: &Ctx<'_>,
    commits: Vec<CommitInfo>,
    labels: &HashMap<Hash, Vec<String>>,
    parents: bool,
) -> Vec<Vec<Value>> {
    let head = ctx.txn.head;
    commits
        .into_iter()
        .map(|c| {
            let refs = match labels.get(&c.hash) {
                Some(names) if c.hash == head => format!("HEAD -> {}", names.join(", ")),
                Some(names) => names.join(", "),
                None => String::new(),
            };
            vec![
                text(c.hash.to_string()),
                text(c.committer_name.clone()),
                text(c.committer_email.clone()),
                timestamp(c.committer_millis as i64),
                text(c.description.clone()),
                Value::Numeric(Numeric::from_i64(c.height as i64)),
                if parents {
                    text(c.parents.iter().map(Hash::to_string).collect::<Vec<_>>().join(", "))
                } else {
                    Value::Null
                },
                text(refs),
                Value::Null,
                text(c.name.clone()),
                text(c.email.clone()),
                timestamp(c.author_millis),
            ]
        })
        .collect()
}

/// LOG parses the arguments of the dolt_log table function.
const LOG: Parser = Parser {
    command: "log",
    options: &[
        ("number", "n", Kind::Value),
        ("min-parents", "", Kind::Value),
        ("merges", "", Kind::Flag),
        ("parents", "", Kind::Flag),
        ("decorate", "", Kind::Value),
        ("not", "", Kind::Value),
        ("all", "", Kind::Flag),
        ("show-signature", "", Kind::Flag),
        ("tables", "t", Kind::Value),
    ],
    max_args: None,
};

/// dolt_log returns the log of the commits that the arguments name, as rows of records.
pub fn dolt_log(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let parsed = LOG.parse(&args.iter().map(|a| a.output().unwrap_or_default()).collect::<Vec<_>>())?;
    let mut included = Vec::new();
    let mut excluded = Vec::new();
    for arg in &parsed.args {
        if let Some((from, to)) = arg.split_once("..") {
            if to.starts_with('.') {
                return Err(error("three-dot revision ranges are not yet supported"));
            }
            excluded.push(history::resolve(ctx.db, ctx.txn.head, from)?);
            included.push(history::resolve(ctx.db, ctx.txn.head, to)?);
        } else if let Some(revision) = arg.strip_prefix('^') {
            excluded.push(history::resolve(ctx.db, ctx.txn.head, revision)?);
        } else {
            included.push(history::resolve(ctx.db, ctx.txn.head, arg)?);
        }
    }
    if let Some(not) = parsed.value("not") {
        excluded.push(history::resolve(ctx.db, ctx.txn.head, not)?);
    }
    if parsed.has("all") {
        included.extend(all_heads(ctx)?);
    }
    if included.is_empty() {
        included.push(ctx.txn.head);
    }
    let min_parents = if parsed.has("merges") {
        2
    } else {
        match parsed.value("min-parents") {
            Some(n) => n.parse::<usize>().map_err(|_| error(format!("invalid value for min-parents: {n}")))?,
            None => 0,
        }
    };
    let decoration = parsed.value("decorate").unwrap_or("short");
    if !matches!(decoration, "short" | "full" | "no" | "auto") {
        return Err(error(format!("invalid --decorate option: \"{decoration}\"")));
    }
    let hidden: std::collections::HashSet<Hash> =
        history::log(ctx.db, &excluded)?.into_iter().map(|c| c.hash).collect();
    let mut commits: Vec<CommitInfo> = history::log(ctx.db, &included)?
        .into_iter()
        .filter(|c| !hidden.contains(&c.hash) && c.parents.len() >= min_parents)
        .collect();
    if let Some(n) = parsed.value("number") {
        commits.truncate(n.parse::<usize>().map_err(|_| error(format!("invalid value for number: {n}")))?);
    }
    let labels = match decoration {
        "no" => HashMap::new(),
        "full" => full_ref_labels(ctx)?,
        _ => ref_labels(ctx)?,
    };
    let rows = log_table(ctx, commits, &labels, parsed.has("parents"));
    Ok(Value::Set(rows.into_iter().map(Value::Record).collect()))
}

/// full_ref_labels returns the labels of each commit's refs with their full names.
fn full_ref_labels(ctx: &mut Ctx<'_>) -> Result<HashMap<Hash, Vec<String>>> {
    let mut labels: HashMap<Hash, Vec<String>> = HashMap::new();
    for prefix in ["refs/heads/", "refs/remotes/"] {
        for (name, hash) in history::refs(ctx.db, prefix)? {
            labels.entry(hash).or_default().push(format!("{prefix}{name}"));
        }
    }
    for (name, address) in history::refs(ctx.db, "refs/tags/")? {
        let commit = history::commit_of(ctx.db, address)?;
        labels.entry(commit).or_default().push(format!("tag: refs/tags/{name}"));
    }
    Ok(labels)
}

/// commit_row returns a commits table row.
fn commit_row(c: &CommitInfo) -> Vec<Value> {
    vec![
        text(c.hash.to_string()),
        text(c.committer_name.clone()),
        text(c.committer_email.clone()),
        timestamp(c.committer_millis as i64),
        text(c.description.clone()),
        text(c.name.clone()),
        text(c.email.clone()),
        timestamp(c.author_millis),
    ]
}

/// all_heads returns the commits of every branch, remote branch, and tag.
fn all_heads(ctx: &mut Ctx<'_>) -> Result<Vec<Hash>> {
    let mut heads = Vec::new();
    for prefix in ["refs/heads/", "refs/remotes/", "refs/tags/"] {
        for (_, address) in history::refs(ctx.db, prefix)? {
            heads.push(history::commit_of(ctx.db, address)?);
        }
    }
    heads.push(ctx.txn.head);
    Ok(heads)
}

/// commit_rows returns every commit reachable from any ref.
fn commit_rows(ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
    let heads = all_heads(ctx)?;
    Ok(history::log(ctx.db, &heads)?.iter().map(commit_row).collect())
}

/// ancestor_rows returns each commit's parents.
fn ancestor_rows(ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
    let heads = all_heads(ctx)?;
    let mut rows = Vec::new();
    for c in history::log(ctx.db, &heads)? {
        if c.parents.is_empty() {
            rows.push(vec![text(c.hash.to_string()), Value::Null, Value::Int4(0)]);
        }
        for (i, parent) in c.parents.iter().enumerate() {
            rows.push(vec![text(c.hash.to_string()), text(parent.to_string()), Value::Int4(i as i32)]);
        }
    }
    Ok(rows)
}

/// branch_rows returns the branches, or the remote branches, under the prefix.
fn branch_rows(ctx: &mut Ctx<'_>, prefix: &str, local: bool) -> Result<Vec<Vec<Value>>> {
    let mut rows = Vec::new();
    for (name, hash) in history::refs(ctx.db, prefix)? {
        let c = history::load(ctx.db, hash)?;
        let mut row = vec![
            text(name.clone()),
            text(hash.to_string()),
            text(c.committer_name.clone()),
            text(c.committer_email.clone()),
            timestamp(c.committer_millis as i64),
            text(c.description.clone()),
        ];
        if local {
            let dirty = branch_dirty(ctx, &name)?;
            row.extend([text(""), text(""), Value::Bool(dirty)]);
        }
        row.extend([text(c.name.clone()), text(c.email.clone()), timestamp(c.author_millis)]);
        rows.push(row);
    }
    Ok(rows)
}

/// branch_dirty reports whether a branch's working or staged root differs from its head.
fn branch_dirty(ctx: &mut Ctx<'_>, name: &str) -> Result<bool> {
    if name == ctx.txn.branch {
        let head = ctx.txn.head_root;
        return Ok(Hash::of(&ctx.txn.root.encode()) != head || Hash::of(&ctx.txn.staged.encode()) != head);
    }
    let Some(address) = ctx.db.head(&doltdb::create::working_set_ref(name))? else { return Ok(false) };
    let data = read(ctx.db, &address)?;
    let ws = serial::WorkingSet::new(serial::Message(&data))?;
    let working = ws.working_root()?;
    let staged = ws.staged_root()?.unwrap_or(working);
    let head = ctx.db.head(&doltdb::create::branch_ref(name))?.map(|h| history::load(ctx.db, h)).transpose()?;
    Ok(head.is_some_and(|h| h.root != working || h.root != staged))
}

/// tag_rows returns the tags.
fn tag_rows(ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
    let mut rows = Vec::new();
    for (name, address) in history::refs(ctx.db, "refs/tags/")? {
        let data = read(ctx.db, &address)?;
        let tag = serial::Tag::new(serial::Message(&data))?;
        let lossy = |b: &[u8]| String::from_utf8_lossy(b).into_owned();
        rows.push(vec![
            text(name),
            text(tag.commit()?.to_string()),
            text(lossy(tag.name()?)),
            text(lossy(tag.email()?)),
            timestamp(tag.user_timestamp_millis()?),
            text(lossy(tag.description()?)),
        ]);
    }
    Ok(rows)
}

/// delta_status returns how a table changed between two roots, or None when it did not.
fn delta_status(from: Option<&Hash>, to: Option<&Hash>) -> Option<&'static str> {
    match (from, to) {
        (None, Some(_)) => Some("new table"),
        (Some(_), None) => Some("deleted"),
        (Some(f), Some(t)) if f != t => Some("modified"),
        _ => None,
    }
}

/// table_deltas returns the tables that changed between two roots, by name.
fn table_deltas(
    from: &BTreeMap<(String, String), Hash>,
    to: &BTreeMap<(String, String), Hash>,
) -> Vec<(String, &'static str)> {
    let mut names: Vec<&(String, String)> = from.keys().chain(to.keys()).collect();
    names.sort();
    names.dedup();
    names
        .into_iter()
        .filter_map(|key| delta_status(from.get(key), to.get(key)).map(|s| (format!("{}.{}", key.0, key.1), s)))
        .collect()
}

/// object_map returns a root's sequences by schema and name, which the status shows alongside tables.
fn object_map(ctx: &mut Ctx<'_>, root: &Root) -> Result<BTreeMap<(String, String), Hash>> {
    let mut objects = BTreeMap::new();
    for (key, address) in root.objects(ctx.db, crate::sequences::COLLECTION)? {
        let mut parts = crate::catalog::id::segments(&key).into_iter();
        objects.insert((parts.next().unwrap_or_default(), parts.next().unwrap_or_default()), address);
    }
    Ok(objects)
}

/// schema_deltas returns the schemas added to or dropped from a root.
fn schema_deltas(from: &Root, to: &Root) -> Vec<(String, &'static str)> {
    let name = |s: &Vec<u8>| String::from_utf8_lossy(s).into_owned();
    let mut deltas: Vec<(String, &'static str)> =
        to.schemas.iter().filter(|s| !from.schemas.contains(s)).map(|s| (name(s), "new schema")).collect();
    deltas.extend(from.schemas.iter().filter(|s| !to.schemas.contains(s)).map(|s| (name(s), "deleted schema")));
    deltas
}

/// status_rows returns the changes staged for the next commit and the changes not yet staged.
fn status_rows(ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
    let head = Root::decode(&read(ctx.db, &ctx.txn.head_root)?)?;
    let (staged, working) = (ctx.txn.staged.clone(), ctx.txn.root.clone());
    let head_tables = table_map(ctx.db, &head)?;
    let staged_tables = table_map(ctx.db, &staged)?;
    let working_tables = table_map(ctx.db, &working)?;
    let mut rows = Vec::new();
    let head_objects = object_map(ctx, &head)?;
    let staged_objects = object_map(ctx, &staged)?;
    let working_objects = object_map(ctx, &working)?;
    let mut staged_deltas = table_deltas(&head_tables, &staged_tables);
    staged_deltas.extend(table_deltas(&head_objects, &staged_objects));
    staged_deltas.sort();
    let mut unstaged_deltas = table_deltas(&staged_tables, &working_tables);
    unstaged_deltas.extend(table_deltas(&staged_objects, &working_objects));
    unstaged_deltas.sort();
    for (name, status) in staged_deltas {
        rows.push(vec![text(name), Value::Bool(true), text(status)]);
    }
    for (name, status) in unstaged_deltas {
        rows.push(vec![text(name), Value::Bool(false), text(status)]);
    }
    for (name, status) in schema_deltas(&head, &staged) {
        rows.push(vec![text(name), Value::Bool(true), text(status)]);
    }
    for (name, status) in schema_deltas(&staged, &working) {
        rows.push(vec![text(name), Value::Bool(false), text(status)]);
    }
    Ok(rows)
}

/// merge_status_rows returns whether a merge is in progress and what it merges.
fn merge_status_rows(ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
    Ok(vec![match &ctx.txn.merge {
        Some(merge) => vec![
            Value::Bool(true),
            text(String::from_utf8_lossy(&merge.from_commit_spec).into_owned()),
            text(merge.from_commit.to_string()),
            text(format!("refs/heads/{}", ctx.txn.branch)),
            text(
                merge
                    .unmergable_tables
                    .iter()
                    .map(|t| String::from_utf8_lossy(t).into_owned())
                    .collect::<Vec<_>>()
                    .join(", "),
            ),
        ],
        None => vec![Value::Bool(false), Value::Null, Value::Null, Value::Null, Value::Null],
    }])
}
