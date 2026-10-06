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

//! Dolt's stored procedures, which Doltgres calls as functions taking text arguments.

use std::collections::BTreeMap;

use doltdb::create::{branch_ref, working_set_ref};
use doltdb::database::{CommitMeta, Database};
use doltdb::root::Root;
use serial::write::{Meta, WorkingSetFields, write_tag};
use store::Hash;

use crate::dolt::args::{Kind, Parsed, Parser, error};
use crate::dolt::history::{self, branch_not_found, valid_branch_name, valid_tag_name};
use crate::error::Result;
use crate::functions::Function;
use crate::oid::{BOOL, INT8, TEXT};
use crate::query::Ctx;
use crate::txn::read;
use crate::types::Value;

/// RECORD is the OID of the record pseudo-type.
pub const RECORD: u32 = 2249;

/// v declares a procedure taking any number of text arguments.
const fn v(name: &'static str, ret: u32, implementation: crate::functions::Implementation) -> Function {
    Function { name, args: &[TEXT], ret, strict: true, variadic: true, implementation }
}

/// FUNCTIONS are Dolt's procedures.
pub const FUNCTIONS: &[Function] = &[
    v("dolt_add", INT8, dolt_add),
    v("dolt_commit", TEXT, dolt_commit),
    v("dolt_branch", INT8, dolt_branch),
    v("dolt_checkout", RECORD, dolt_checkout),
    v("dolt_tag", INT8, dolt_tag),
    v("dolt_reset", INT8, dolt_reset),
    v("dolt_merge", RECORD, dolt_merge),
    f("active_branch", &[], TEXT, active_branch),
    f("hashof", &[TEXT], TEXT, hashof),
    f("dolt_hashof", &[TEXT], TEXT, hashof),
    f("dolt_merge_base", &[TEXT, TEXT], TEXT, dolt_merge_base),
    f("has_ancestor", &[TEXT, TEXT], BOOL, has_ancestor),
    f("dolt_version", &[], TEXT, dolt_version),
    Function {
        name: "dolt_log",
        args: &[TEXT],
        ret: RECORD,
        strict: false,
        variadic: true,
        implementation: crate::dolt::tables::dolt_log,
    },
    Function {
        name: "dolt_diff_summary",
        args: &[TEXT],
        ret: RECORD,
        strict: false,
        variadic: true,
        implementation: crate::dolt::diff::dolt_diff_summary,
    },
    Function {
        name: "dolt_diff_stat",
        args: &[TEXT],
        ret: RECORD,
        strict: false,
        variadic: true,
        implementation: crate::dolt::diff::dolt_diff_stat,
    },
];

/// f declares a strict function with fixed parameters.
const fn f(
    name: &'static str,
    args: &'static [u32],
    ret: u32,
    implementation: crate::functions::Implementation,
) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// active_branch returns the session's branch.
fn active_branch(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Text(ctx.session.branch.clone()))
}

/// hashof returns the hash of the commit a spec names.
fn hashof(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let spec = args[0].output().unwrap_or_default();
    Ok(Value::Text(history::resolve(ctx.db, ctx.txn.head, &spec)?.to_string()))
}

/// dolt_merge_base returns the best common ancestor of two commits.
fn dolt_merge_base(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let left = history::resolve(ctx.db, ctx.txn.head, &args[0].output().unwrap_or_default())?;
    let right = history::resolve(ctx.db, ctx.txn.head, &args[1].output().unwrap_or_default())?;
    let base = history::merge_base(ctx.db, left, right)?.ok_or_else(|| error("no common ancestor"))?;
    Ok(Value::Text(base.to_string()))
}

/// has_ancestor reports whether a commit has another as an ancestor.
fn has_ancestor(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let commit = history::resolve(ctx.db, ctx.txn.head, &args[0].output().unwrap_or_default())?;
    let ancestor = history::resolve(ctx.db, ctx.txn.head, &args[1].output().unwrap_or_default())?;
    Ok(Value::Bool(history::is_ancestor(ctx.db, ancestor, commit)?))
}

/// dolt_version returns the server's version.
fn dolt_version(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Text(crate::DOLTGRES_VERSION.to_string()))
}

/// OUT_COLUMNS are the result columns of the procedures, which name their columns in FROM.
pub const OUT_COLUMNS: &[(&str, &[(&str, u32)])] = &[
    ("dolt_add", &[("status", INT8)]),
    ("dolt_commit", &[("hash", TEXT)]),
    ("dolt_branch", &[("status", INT8)]),
    ("dolt_tag", &[("status", INT8)]),
    ("dolt_reset", &[("status", INT8)]),
    ("dolt_checkout", &[("status", INT8), ("message", TEXT)]),
    ("dolt_merge", &[("hash", TEXT), ("fast_forward", INT8), ("conflicts", INT8), ("message", TEXT)]),
    (
        "dolt_log",
        &[
            ("commit_hash", TEXT),
            ("committer", TEXT),
            ("email", TEXT),
            ("date", crate::oid::TIMESTAMP),
            ("message", TEXT),
            ("commit_order", crate::oid::NUMERIC),
            ("parents", TEXT),
            ("refs", TEXT),
            ("signature", TEXT),
            ("author", TEXT),
            ("author_email", TEXT),
            ("author_date", crate::oid::TIMESTAMP),
        ],
    ),
    (
        "dolt_diff_summary",
        &[
            ("from_table_name", TEXT),
            ("to_table_name", TEXT),
            ("diff_type", TEXT),
            ("data_change", BOOL),
            ("schema_change", BOOL),
        ],
    ),
    (
        "dolt_diff_stat",
        &[
            ("table_name", TEXT),
            ("rows_unmodified", INT8),
            ("rows_added", INT8),
            ("rows_deleted", INT8),
            ("rows_modified", INT8),
            ("cells_added", INT8),
            ("cells_deleted", INT8),
            ("cells_modified", INT8),
            ("old_row_count", INT8),
            ("new_row_count", INT8),
            ("old_cell_count", INT8),
            ("new_cell_count", INT8),
        ],
    ),
];

/// strings returns the text arguments.
fn strings(args: &[Value]) -> Vec<String> {
    args.iter().map(|a| a.output().unwrap_or_default()).collect()
}

/// table_map returns a root's tables by schema and name.
pub fn table_map(db: &mut Database, root: &Root) -> Result<BTreeMap<(String, String), Hash>> {
    let mut tables = BTreeMap::new();
    for (key, address) in root.tables(db)? {
        let text = String::from_utf8_lossy(&key).into_owned();
        let mut parts = text.splitn(3, '\0').skip(1);
        let schema = parts.next().unwrap_or_default().to_string();
        let name = parts.next().unwrap_or_default().to_string();
        tables.insert((schema, name), address);
    }
    Ok(tables)
}

/// find_table finds a table by name in the roots, searching the session's schemas for an unqualified name.
fn find_table(ctx: &mut Ctx<'_>, roots: &[&Root], name: &str) -> Result<Option<(String, String)>> {
    let (schemas, table) = match name.split_once('.') {
        Some((schema, table)) => (vec![schema.to_string()], table),
        None => (ctx.session.search_path(), name),
    };
    for root in roots {
        let tables = table_map(ctx.db, root)?;
        for schema in &schemas {
            if tables.contains_key(&(schema.clone(), table.to_string())) {
                return Ok(Some((schema.clone(), table.to_string())));
            }
        }
    }
    Ok(None)
}

/// stage_all copies every table, schema, and root object of the working root to the staged root.
fn stage_all(ctx: &mut Ctx<'_>) -> Result<()> {
    let working = table_map(ctx.db, &ctx.txn.root)?;
    let staged = table_map(ctx.db, &ctx.txn.staged)?;
    for key in staged.keys().filter(|k| !working.contains_key(*k)) {
        ctx.txn.staged.put_table(ctx.db, &key.0, &key.1, None)?;
    }
    for (key, address) in &working {
        if staged.get(key) != Some(address) {
            ctx.txn.staged.put_table(ctx.db, &key.0, &key.1, Some(*address))?;
        }
    }
    stage_database(ctx);
    Ok(())
}

/// stage_database copies the working root's schemas, foreign keys, and root objects to the staged root.
fn stage_database(ctx: &mut Ctx<'_>) {
    for schema in &ctx.txn.root.schemas {
        if !ctx.txn.staged.schemas.contains(schema) {
            ctx.txn.staged.schemas.push(schema.clone());
        }
    }
    ctx.txn.staged.schemas.sort();
    ctx.txn.staged.foreign_keys = ctx.txn.root.foreign_keys.clone();
    ctx.txn.staged.root_objects = ctx.txn.root.root_objects;
    ctx.txn.staged.collation = ctx.txn.root.collation;
}

/// stage_tables copies the named tables of the working root to the staged root.
fn stage_tables(ctx: &mut Ctx<'_>, names: &[String]) -> Result<()> {
    let mut found = Vec::new();
    let mut missing = Vec::new();
    let (working, staged) = (ctx.txn.root.clone(), ctx.txn.staged.clone());
    for name in names {
        match find_table(ctx, &[&working, &staged], name)? {
            Some(key) => found.push(key),
            None => missing.push(name.clone()),
        }
    }
    if !missing.is_empty() {
        return Err(error(format!("error: the table(s) {} do not exist", missing.join(", "))));
    }
    for (schema, table) in found {
        let address = working.table(ctx.db, &schema, &table)?;
        ctx.txn.staged.put_table(ctx.db, &schema, &table, address)?;
    }
    Ok(())
}

/// ADD parses dolt_add's arguments.
const ADD: Parser = Parser {
    command: "add",
    options: &[("all", "A", Kind::Flag), ("force", "f", Kind::Flag), ("branch", "", Kind::Value)],
    max_args: None,
};

/// dolt_add stages tables.
fn dolt_add(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let parsed = ADD.parse(&strings(args))?;
    if parsed.args.is_empty() && !parsed.has("all") {
        return Err(error("Nothing specified, nothing added. Maybe you wanted to say 'dolt add .'?"));
    }
    if parsed.has("all") || parsed.args == ["."] {
        stage_all(ctx)?;
    } else {
        stage_tables(ctx, &parsed.args)?;
    }
    Ok(Value::Int8(0))
}

/// stage_modified stages the tables of the working root that the staged root also has.
fn stage_modified(ctx: &mut Ctx<'_>) -> Result<()> {
    let working = table_map(ctx.db, &ctx.txn.root)?;
    let staged = table_map(ctx.db, &ctx.txn.staged)?;
    for (key, address) in &staged {
        match working.get(key) {
            Some(w) if w != address => ctx.txn.staged.put_table(ctx.db, &key.0, &key.1, Some(*w))?,
            None => ctx.txn.staged.put_table(ctx.db, &key.0, &key.1, None)?,
            _ => {}
        }
    }
    stage_database(ctx);
    Ok(())
}

/// COMMIT parses dolt_commit's arguments.
const COMMIT: Parser = Parser {
    command: "commit",
    options: &[
        ("message", "m", Kind::Value),
        ("allow-empty", "", Kind::Flag),
        ("skip-empty", "", Kind::Flag),
        ("date", "", Kind::Value),
        ("force", "f", Kind::Flag),
        ("author", "", Kind::Value),
        ("all", "a", Kind::Flag),
        ("ALL", "A", Kind::Flag),
        ("amend", "", Kind::Flag),
        ("gpg-sign", "S", Kind::OptionalValue),
        ("skip-verification", "", Kind::Flag),
        ("branch", "", Kind::Value),
    ],
    max_args: Some(0),
};

/// parse_author reads an author written as `Name <email>`.
pub fn parse_author(text: &str) -> Result<(String, String)> {
    let malformed = || error("Author not formatted correctly. Use 'Name <author@example.com>' format");
    let open = text.find('<').ok_or_else(malformed)?;
    let close = text.rfind('>').filter(|&c| c > open).ok_or_else(malformed)?;
    let name = text[..open].trim();
    if name.is_empty() {
        return Err(malformed());
    }
    Ok((name.to_string(), text[open + 1..close].trim().to_string()))
}

/// parse_date reads a commit date in one of the formats Dolt accepts, returning Unix milliseconds.
fn parse_date(text: &str) -> Result<i64> {
    let format = crate::datetime::Format::from_settings("ISO, MDY", "postgres", "UTC");
    let ts = crate::datetime::parse_timestamp(text, true, &format, crate::datetime::now())
        .map_err(|_| error(format!("error: '{text}' is not in a supported format.")))?;
    Ok((ts - crate::datetime::UNIX_EPOCH_DAYS * crate::datetime::USECS_PER_DAY).div_euclid(1000))
}

/// now_millis returns the current time in Unix milliseconds.
pub fn now_millis() -> i64 {
    std::time::SystemTime::now().duration_since(std::time::UNIX_EPOCH).map_or(0, |d| d.as_millis() as i64)
}

/// commit_meta returns the metadata of a new commit by the session's user, now.
pub fn commit_meta(ctx: &Ctx<'_>, description: &str) -> CommitMeta {
    let millis = now_millis();
    CommitMeta {
        name: ctx.session.user.clone(),
        email: format!("{}@{}", ctx.session.user, ctx.session.host),
        description: description.to_string(),
        author_millis: millis,
        committer_millis: millis as u64,
        signature: String::new(),
        committer_name: None,
        committer_email: None,
    }
}

/// dolt_commit commits the staged tables.
fn dolt_commit(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let parsed = COMMIT.parse(&strings(args))?;
    if parsed.has("allow-empty") && parsed.has("skip-empty") {
        return Err(error("error: cannot use both --allow-empty and --skip-empty"));
    }
    if parsed.has("ALL") {
        stage_all(ctx)?;
    } else if parsed.has("all") {
        stage_modified(ctx)?;
    }
    let amend = parsed.has("amend");
    let head = history::load(ctx.db, ctx.txn.head)?;
    let message = match parsed.value("message") {
        Some(m) => m.to_string(),
        None if amend => head.description.clone(),
        None => return Err(error("Must provide commit message.")),
    };
    let mut meta = commit_meta(ctx, &message);
    if let Some(author) = parsed.value("author") {
        let (name, email) = parse_author(author)?;
        meta.name = name;
        meta.email = email;
    }
    if let Some(date) = parsed.value("date") {
        let millis = parse_date(date)?;
        meta.author_millis = millis;
        meta.committer_millis = millis as u64;
    }
    let merging = ctx.txn.merge.as_ref().map(|m| m.from_commit);
    let empty = Hash::of(&ctx.txn.staged.encode()) == ctx.txn.head_root;
    if empty && !parsed.has("allow-empty") && merging.is_none() && !amend {
        if parsed.has("skip-empty") {
            return Ok(Value::Null);
        }
        return Err(error("nothing to commit"));
    }
    let (user, host) = (ctx.session.user.clone(), ctx.session.host.clone());
    let hash = if amend {
        amend_commit(ctx, &head, meta)?
    } else {
        let parents = merging.map(|m| vec![ctx.txn.head, m]).unwrap_or_default();
        ctx.txn.dolt_commit(ctx.db, &user, &host, parents, meta)?
    };
    if ctx.session.explicit {
        ctx.session.explicit = false;
        ctx.session.end_transaction(true);
    }
    Ok(Value::Text(hash.to_string()))
}

/// amend_commit replaces the head commit with one of the staged root on the head's parents.
fn amend_commit(ctx: &mut Ctx<'_>, head: &history::CommitInfo, meta: CommitMeta) -> Result<Hash> {
    let commit = ctx.db.build_commit(None, ctx.txn.staged.encode(), head.parents.clone(), &meta)?;
    ctx.db.write_value(commit.bytes.clone())?;
    ctx.db.set_head(&branch_ref(&ctx.txn.branch), commit.hash)?;
    ctx.txn.head = commit.hash;
    ctx.txn.head_root = history::load(ctx.db, commit.hash)?.root;
    flush(ctx)?;
    Ok(commit.hash)
}

/// flush writes the current branch's working set now, as Dolt does when a procedure commits the transaction.
pub fn flush(ctx: &mut Ctx<'_>) -> Result<()> {
    let (user, host) = (ctx.session.user.clone(), ctx.session.host.clone());
    ctx.txn.flush(ctx.db, &user, &host)
}

/// new_branch points a branch at a commit with a working set of the commit's root, as Dolt's NewBranchAtCommit
/// does.
pub fn new_branch(db: &mut Database, name: &str, commit: Hash) -> Result<()> {
    let root = history::load(db, commit)?.root;
    db.set_head(&branch_ref(name), commit)?;
    let previous = db.head(&working_set_ref(name))?.unwrap_or_default();
    let fields = WorkingSetFields {
        working_root: root,
        staged_root: Some(root),
        merge_state: None,
        rebase_state: None,
        meta: Some(Meta {
            name: Vec::new(),
            email: Vec::new(),
            description: b"updated from dolt environment".to_vec(),
            timestamp_millis: now_millis() as u64 / 1000,
            user_timestamp_millis: 0,
        }),
    };
    db.update_working_set(&working_set_ref(name), &fields, previous)?;
    Ok(())
}

/// delete_branch removes a branch and its working set.
pub fn delete_branch(db: &mut Database, name: &str) -> Result<()> {
    db.delete_heads(&[branch_ref(name), working_set_ref(name)])?;
    Ok(())
}

/// branch_exists reports whether the branch exists.
pub fn branch_exists(db: &mut Database, name: &str) -> Result<bool> {
    Ok(db.head(&branch_ref(name))?.is_some())
}

/// BRANCH parses dolt_branch's arguments.
const BRANCH: Parser = Parser {
    command: "branch",
    options: &[
        ("force", "f", Kind::Flag),
        ("copy", "c", Kind::Flag),
        ("move", "m", Kind::Flag),
        ("delete", "d", Kind::Flag),
        ("D", "", Kind::Flag),
        ("track", "t", Kind::Flag),
        ("set-upstream-to", "u", Kind::Value),
        ("remote", "r", Kind::Flag),
    ],
    max_args: None,
};

/// dolt_branch creates, copies, renames, or deletes branches.
fn dolt_branch(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let parsed = BRANCH.parse(&strings(args))?;
    let force = parsed.has("force");
    if parsed.has("copy") {
        copy_branch(ctx, &parsed, force)?;
    } else if parsed.has("move") {
        rename_branch(ctx, &parsed, force)?;
    } else if parsed.has("delete") || parsed.has("D") {
        delete_branches(ctx, &parsed, force || parsed.has("D"))?;
    } else {
        create_branch(ctx, &parsed, force)?;
    }
    flush(ctx)?;
    Ok(Value::Int8(0))
}

/// invalid_usage returns Dolt's error for wrong arguments.
fn invalid_usage() -> crate::error::PgError {
    error("error: invalid usage")
}

/// empty_branch_name returns Dolt's error for an empty branch name.
fn empty_branch_name() -> crate::error::PgError {
    error("error: cannot branch empty string")
}

/// create_branch creates a branch at a start point, the session's head by default.
fn create_branch(ctx: &mut Ctx<'_>, parsed: &Parsed, force: bool) -> Result<()> {
    let name = match parsed.args.as_slice() {
        [name] | [name, _] => name.clone(),
        _ => return Err(invalid_usage()),
    };
    if name.is_empty() {
        return Err(empty_branch_name());
    }
    let start = parsed.args.get(1).map_or("head", String::as_str);
    if start.is_empty() {
        return Err(invalid_usage());
    }
    create_branch_at(ctx, &name, start, force)
}

/// create_branch_at creates a branch at the commit a spec names, as Dolt's CreateBranchWithStartPt does.
pub fn create_branch_at(ctx: &mut Ctx<'_>, name: &str, start: &str, force: bool) -> Result<()> {
    if !force && branch_exists(ctx.db, name)? {
        return Err(error(format!("fatal: A branch named '{name}' already exists.")));
    }
    if !valid_branch_name(name) || name.eq_ignore_ascii_case("head") {
        return Err(error(format!("fatal: '{name}' is an invalid branch name.")));
    }
    let commit = history::resolve(ctx.db, ctx.txn.head, start)
        .map_err(|e| error(format!("fatal: Unexpected error creating branch '{name}' : {}", e.message)))?;
    new_branch(ctx.db, name, commit)
}

/// copy_branch copies a branch to a new name.
fn copy_branch(ctx: &mut Ctx<'_>, parsed: &Parsed, force: bool) -> Result<()> {
    let [source, dest] = parsed.args.as_slice() else { return Err(invalid_usage()) };
    if source.is_empty() || dest.is_empty() {
        return Err(empty_branch_name());
    }
    let Some(commit) = ctx.db.head(&branch_ref(source))? else {
        return Err(error(format!("fatal: A branch named '{source}' not found")));
    };
    if !force && branch_exists(ctx.db, dest)? {
        return Err(error(format!("fatal: A branch named '{dest}' already exists.")));
    }
    if !valid_branch_name(dest) {
        return Err(error(format!("fatal: '{dest}' is not a valid branch name.")));
    }
    new_branch(ctx.db, dest, commit)
}

/// rename_branch renames a branch with its working set, following it when it is the session's branch.
fn rename_branch(ctx: &mut Ctx<'_>, parsed: &Parsed, force: bool) -> Result<()> {
    let [old, new] = parsed.args.as_slice() else { return Err(invalid_usage()) };
    if old.is_empty() || new.is_empty() {
        return Err(empty_branch_name());
    }
    if old == new {
        return if branch_exists(ctx.db, old)? { Ok(()) } else { Err(error("branch not found")) };
    }
    flush(ctx)?;
    let Some(commit) = ctx.db.head(&branch_ref(old))? else { return Err(error("branch not found")) };
    if !force && branch_exists(ctx.db, new)? {
        return Err(error(format!("fatal: A branch named '{new}' already exists.")));
    }
    if !valid_branch_name(new) {
        return Err(error("not a valid user branch name"));
    }
    ctx.db.set_head(&branch_ref(new), commit)?;
    if let Some(ws) = ctx.db.head(&working_set_ref(old))? {
        ctx.db.set_head(&working_set_ref(new), ws)?;
    }
    delete_branch(ctx.db, old)?;
    if ctx.session.branch == *old {
        ctx.session.branch = new.clone();
        ctx.txn.branch = new.clone();
    }
    Ok(())
}

/// delete_branches deletes branches, refusing unmerged ones without force.
fn delete_branches(ctx: &mut Ctx<'_>, parsed: &Parsed, force: bool) -> Result<()> {
    if parsed.args.is_empty() {
        return Err(invalid_usage());
    }
    for name in &parsed.args {
        if name.is_empty() {
            return Err(empty_branch_name());
        }
        if *name == ctx.session.branch && !force {
            return Err(error(format!("Cannot delete checked out branch '{name}'")));
        }
        let Some(head) = ctx.db.head(&branch_ref(name))? else { return Err(error("branch not found")) };
        if !force && !history::is_ancestor(ctx.db, head, ctx.txn.head)? {
            return Err(error(format!("branch '{name}' is not fully merged")));
        }
        delete_branch(ctx.db, name)?;
        if *name == ctx.session.branch {
            ctx.session.branch = crate::DEFAULT_BRANCH.to_string();
        }
    }
    Ok(())
}

/// CHECKOUT parses dolt_checkout's arguments.
const CHECKOUT: Parser = Parser {
    command: "checkout",
    options: &[
        ("b", "", Kind::Value),
        ("B", "", Kind::Value),
        ("force", "f", Kind::Flag),
        ("track", "t", Kind::Value),
        ("overwrite-ignore", "", Kind::Flag),
        ("no-overwrite-ignore", "", Kind::Flag),
        ("move", "m", Kind::Flag),
    ],
    max_args: None,
};

/// record returns a record of a status and a message.
fn record(status: i64, message: impl Into<String>) -> Value {
    Value::Record(vec![Value::Int8(status), Value::Text(message.into())])
}

/// dolt_checkout switches the session's branch, creates a branch and switches to it, or restores tables.
fn dolt_checkout(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let parsed = CHECKOUT.parse(&strings(args))?;
    if parsed.has("b") && parsed.has("B") {
        return Err(error("Improper usage. Cannot use both -b and -B."));
    }
    let new_branch = parsed.value("b").or(parsed.value("B"));
    if new_branch == Some("") {
        return Err(error("error: cannot checkout empty string"));
    }
    if let Some(separator) = parsed.separator
        && separator >= 2
    {
        return Err(error(format!("only one reference expected, {separator} given")));
    }
    if let Some(name) = new_branch {
        let start = parsed.args.first().map_or("head", String::as_str);
        create_branch_at(ctx, name, start, parsed.has("B"))?;
        flush(ctx)?;
        ctx.session.branch = name.to_string();
        return Ok(record(0, format!("Switched to branch '{name}'")));
    }
    let Some(first) = parsed.args.first().cloned() else { return Err(error("Improper usage.")) };
    if first.is_empty() {
        return Err(error("error: cannot checkout empty string"));
    }
    if parsed.separator == Some(0) {
        checkout_tables(ctx, &parsed.args)?;
        return Ok(record(0, ""));
    }
    if parsed.args.len() == 1 && first == ctx.session.branch {
        return Ok(record(0, format!("Already on branch '{first}'")));
    }
    let is_branch = branch_exists(ctx.db, &first)?;
    if parsed.args.len() == 1 && !is_branch {
        let tags = history::refs(ctx.db, "refs/tags/")?;
        if history::is_hash(&first) || tags.iter().any(|(t, _)| *t == first) {
            return Err(error(format!(
                "dolt does not support a detached head state. To create a branch at this ref, run:\n\tdolt checkout {first} -b {{new_branch_name}}"
            )));
        }
    }
    if is_branch && parsed.args.len() == 1 {
        ctx.session.branch = first.clone();
        return Ok(record(0, format!("Switched to branch '{first}'")));
    }
    if parsed.args.len() > 1 && is_branch {
        checkout_tables_from(ctx, &first, &parsed.args[1..])?;
        return Ok(record(0, ""));
    }
    checkout_tables(ctx, &parsed.args)?;
    Ok(record(0, ""))
}

/// checkout_tables restores tables of the working root from the staged root, or from the head for tables the
/// staged root lacks, as Dolt's checkoutTablesFromHead does.
fn checkout_tables(ctx: &mut Ctx<'_>, names: &[String]) -> Result<()> {
    let head = Root::decode(&read(ctx.db, &ctx.txn.head_root)?)?;
    let (working, staged) = (ctx.txn.root.clone(), ctx.txn.staged.clone());
    let mut keys = Vec::new();
    let mut warnings = Vec::new();
    if names == ["."] {
        let mut all: Vec<(String, String)> = table_map(ctx.db, &head)?.into_keys().collect();
        all.extend(table_map(ctx.db, &staged)?.into_keys());
        all.sort();
        all.dedup();
        keys = all;
    } else {
        for name in names {
            match find_table(ctx, &[&working, &staged, &head], name)? {
                Some(key) => keys.push(key),
                None => warnings.push(format!("error: tablespec '{name}' did not match any table(s) known to dolt")),
            }
        }
    }
    let mut unknown = false;
    for (schema, table) in &keys {
        let address = match staged.table(ctx.db, schema, table)? {
            Some(address) => Some(address),
            None => head.table(ctx.db, schema, table)?,
        };
        match address {
            Some(address) => ctx.txn.root.put_table(ctx.db, schema, table, Some(address))?,
            None => unknown = true,
        }
    }
    if unknown {
        return Err(error("error: given tables do not exist"));
    }
    if !keys.is_empty() {
        flush(ctx)?;
    }
    if !warnings.is_empty() {
        return Err(error(warnings.join("\n")));
    }
    Ok(())
}

/// checkout_tables_from restores tables of the working and staged roots from a commit.
fn checkout_tables_from(ctx: &mut Ctx<'_>, spec: &str, names: &[String]) -> Result<()> {
    let commit = history::resolve(ctx.db, ctx.txn.head, spec)?;
    let root = Root::decode(&read(ctx.db, &history::load(ctx.db, commit)?.root)?)?;
    for name in names {
        let Some((schema, table)) = find_table(ctx, &[&root], name)? else {
            return Err(error(format!("error: the table(s) {name} do not exist")));
        };
        let address = root.table(ctx.db, &schema, &table)?;
        ctx.txn.root.put_table(ctx.db, &schema, &table, address)?;
        ctx.txn.staged.put_table(ctx.db, &schema, &table, address)?;
    }
    Ok(())
}

/// TAG parses dolt_tag's arguments.
const TAG: Parser = Parser {
    command: "tag",
    options: &[
        ("message", "m", Kind::Value),
        ("verbose", "v", Kind::Flag),
        ("delete", "d", Kind::Flag),
        ("author", "", Kind::Value),
    ],
    max_args: None,
};

/// dolt_tag creates or deletes tags.
fn dolt_tag(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let parsed = TAG.parse(&strings(args))?;
    if parsed.has("delete") {
        if parsed.args.is_empty() {
            return Err(error("error: tag name must be provided."));
        }
        for name in &parsed.args {
            if ctx.db.head(&format!("refs/tags/{name}"))?.is_none() {
                return Err(error(format!("tag not found: {name}")));
            }
            ctx.db.delete_heads(&[format!("refs/tags/{name}")])?;
        }
        return Ok(Value::Int8(0));
    }
    let (name, start) = match parsed.args.as_slice() {
        [name] => (name.clone(), "head".to_string()),
        [name, start] => (name.clone(), start.clone()),
        [] => return Err(error("error: tag name must be provided.")),
        _ => return Err(error("create tag takes at most two args")),
    };
    if !valid_tag_name(&name) {
        return Err(error("not a valid user tag name"));
    }
    if ctx.db.head(&format!("refs/tags/{name}"))?.is_some() {
        return Err(error(format!("tag '{name}' already exists")));
    }
    let commit = history::resolve(ctx.db, ctx.txn.head, &start)?;
    let mut meta = commit_meta(ctx, parsed.value("message").unwrap_or_default());
    if let Some(author) = parsed.value("author") {
        let (name, email) = parse_author(author)?;
        meta.name = name;
        meta.email = email;
    }
    let tag = write_tag(
        commit,
        Some(&Meta {
            name: meta.name.into_bytes(),
            email: meta.email.into_bytes(),
            description: meta.description.into_bytes(),
            timestamp_millis: meta.committer_millis,
            user_timestamp_millis: meta.author_millis,
        }),
    );
    let address = ctx.db.write_value(tag)?;
    ctx.db.set_head(&format!("refs/tags/{name}"), address)?;
    Ok(Value::Int8(0))
}

/// RESET parses dolt_reset's arguments.
const RESET: Parser =
    Parser { command: "reset", options: &[("hard", "", Kind::Flag), ("soft", "", Kind::Flag)], max_args: None };

/// dolt_reset resets the staged root, and with --hard the working root, to a commit, moving the branch to it.
fn dolt_reset(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let parsed = RESET.parse(&strings(args))?;
    if parsed.has("hard") && parsed.has("soft") {
        return Err(error("error: --hard and --soft are mutually exclusive options."));
    }
    if parsed.has("hard") || parsed.has("soft") {
        let spec = parsed.args.first().map_or("head", String::as_str);
        let commit = history::resolve(ctx.db, ctx.txn.head, spec)?;
        if parsed.has("hard") {
            let root = Root::decode(&read(ctx.db, &history::load(ctx.db, commit)?.root)?)?;
            let head = Root::decode(&read(ctx.db, &ctx.txn.head_root)?)?;
            let tracked: Vec<(String, String)> =
                table_map(ctx.db, &head)?.into_keys().chain(table_map(ctx.db, &ctx.txn.staged)?.into_keys()).collect();
            let untracked: Vec<((String, String), Hash)> =
                table_map(ctx.db, &ctx.txn.root)?.into_iter().filter(|(k, _)| !tracked.contains(k)).collect();
            let mut working = root.clone();
            for ((schema, name), address) in untracked {
                if working.table(ctx.db, &schema, &name)?.is_none() {
                    working.put_table(ctx.db, &schema, &name, Some(address))?;
                }
            }
            ctx.txn.root = working;
            ctx.txn.staged = root;
            ctx.txn.merge = None;
        }
        move_head(ctx, commit)?;
    } else if parsed.args.is_empty() || parsed.args == ["."] {
        reset_staged_tables(ctx, None)?;
    } else {
        let head = Root::decode(&read(ctx.db, &ctx.txn.head_root)?)?;
        let (staged, working) = (ctx.txn.staged.clone(), ctx.txn.root.clone());
        match find_table(ctx, &[&head, &staged, &working], &parsed.args[0])? {
            Some(key) => reset_staged_tables(ctx, Some(vec![key]))?,
            None => {
                let commit = history::resolve(ctx.db, ctx.txn.head, &parsed.args[0])?;
                ctx.txn.staged = Root::decode(&read(ctx.db, &history::load(ctx.db, commit)?.root)?)?;
                ctx.txn.merge = None;
                move_head(ctx, commit)?;
            }
        }
    }
    flush(ctx)?;
    Ok(Value::Int8(0))
}

/// reset_staged_tables restores tables of the staged root from the head, all of them without names.
fn reset_staged_tables(ctx: &mut Ctx<'_>, keys: Option<Vec<(String, String)>>) -> Result<()> {
    let head = Root::decode(&read(ctx.db, &ctx.txn.head_root)?)?;
    let keys = match keys {
        Some(keys) => keys,
        None => {
            ctx.txn.staged.schemas = head.schemas.clone();
            ctx.txn.staged.foreign_keys = head.foreign_keys.clone();
            ctx.txn.staged.root_objects = head.root_objects;
            let mut all: Vec<(String, String)> = table_map(ctx.db, &head)?.into_keys().collect();
            all.extend(table_map(ctx.db, &ctx.txn.staged)?.into_keys());
            all
        }
    };
    for (schema, name) in keys {
        let address = head.table(ctx.db, &schema, &name)?;
        ctx.txn.staged.put_table(ctx.db, &schema, &name, address)?;
    }
    Ok(())
}

/// move_head points the session's branch at a commit.
fn move_head(ctx: &mut Ctx<'_>, commit: Hash) -> Result<()> {
    if commit != ctx.txn.head {
        ctx.db.set_head(&branch_ref(&ctx.txn.branch), commit)?;
        ctx.txn.head = commit;
        ctx.txn.head_root = history::load(ctx.db, commit)?.root;
    }
    Ok(())
}

/// MERGE parses dolt_merge's arguments.
const MERGE: Parser = Parser {
    command: "merge",
    options: &[
        ("no-ff", "", Kind::Flag),
        ("ff-only", "", Kind::Flag),
        ("squash", "", Kind::Flag),
        ("message", "m", Kind::Value),
        ("abort", "", Kind::Flag),
        ("commit", "", Kind::Flag),
        ("no-commit", "", Kind::Flag),
        ("no-edit", "", Kind::Flag),
        ("author", "", Kind::Value),
        ("skip-verification", "", Kind::Flag),
    ],
    max_args: Some(1),
};

/// merge_record returns dolt_merge's result.
fn merge_record(hash: &str, fast_forward: bool, conflicts: i64, message: &str) -> Value {
    Value::Record(vec![
        Value::Text(hash.to_string()),
        Value::Int8(fast_forward as i64),
        Value::Int8(conflicts),
        Value::Text(message.to_string()),
    ])
}

/// dolt_merge merges a branch or commit into the session's branch.
fn dolt_merge(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let parsed = MERGE.parse(&strings(args)).map_err(|e| {
        if e.message.contains("too many positional arguments") {
            error("Error: Dolt does not support merging from multiple commits. You probably meant to checkout one and then merge from the other.")
        } else {
            e
        }
    })?;
    if parsed.has("abort") {
        let Some(merge) = ctx.txn.merge.take() else { return Err(error("fatal: There is no merge to abort")) };
        ctx.txn.root = Root::decode(&read(ctx.db, &merge.pre_working_root)?)?;
        ctx.txn.staged = Root::decode(&read(ctx.db, &ctx.txn.head_root)?)?;
        flush(ctx)?;
        return Ok(merge_record("", false, 0, "merge aborted"));
    }
    let Some(spec) = parsed.args.first().cloned() else {
        return Err(error("Error: No commit specified"));
    };
    if ctx.txn.merge.is_some() {
        return Err(error("merging is not possible because you have not committed an active merge"));
    }
    let theirs = history::resolve(ctx.db, ctx.txn.head, &spec)
        .map_err(|e| if e.message.starts_with("branch not found") { branch_not_found(&spec) } else { e })?;
    if history::is_ancestor(ctx.db, theirs, ctx.txn.head)? {
        return Ok(merge_record(
            &ctx.txn.head.to_string(),
            false,
            0,
            "cannot fast forward from a to b. a is ahead of b already",
        ));
    }
    let fast_forward = history::is_ancestor(ctx.db, ctx.txn.head, theirs)? && !parsed.has("no-ff");
    if fast_forward {
        if ctx.txn.changed() || Hash::of(&ctx.txn.staged.encode()) != ctx.txn.head_root {
            return Err(error(
                "error: Your local changes would be overwritten by merge.\nPlease commit your changes before you merge.",
            ));
        }
        let root = Root::decode(&read(ctx.db, &history::load(ctx.db, theirs)?.root)?)?;
        ctx.txn.root = root.clone();
        ctx.txn.staged = root;
        move_head(ctx, theirs)?;
        flush(ctx)?;
        return Ok(merge_record(&theirs.to_string(), true, 0, "merge successful"));
    }
    if parsed.has("ff-only") {
        return Err(error("fatal: Not possible to fast-forward, aborting."));
    }
    Err(crate::error::PgError::unsupported("merges that are not fast-forwards"))
}
