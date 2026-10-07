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

//! Dolt's smaller procedures: cleaning untracked tables, removing tracked ones, counting commits between two
//! revisions, committing into a session variable, restoring and purging dropped databases, and dumping threads.

use crate::dolt::args::{Kind, Parser, error};
use crate::dolt::history;
use crate::dolt::procedures::{find_object, find_table, strings, table_map};
use crate::error::{PgError, Result};
use crate::query::Ctx;
use crate::txn::read;
use crate::types::Value;

/// CLEAN parses dolt_clean's arguments.
const CLEAN: Parser =
    Parser { command: "clean", options: &[("dry-run", "", Kind::Flag), ("x", "x", Kind::Flag)], max_args: None };

/// dolt_clean deletes the working root's tables and root objects that the staged root lacks, or only the named ones,
/// leaving out tables that dolt_ignore ignores unless asked not to, as Dolt's CleanUntracked does.
pub fn dolt_clean(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    ctx.check_branch_access(crate::dolt::branch_control::WRITE)?;
    let parsed = CLEAN.parse(&strings(args))?;
    let working = ctx.txn.root.clone();
    let mut untracked = Vec::new();
    for name in &parsed.args {
        match find_table(ctx, &[&working], name)? {
            Some(key) => untracked.push(key),
            None if crate::dolt::procedures::find_object(ctx, &[&working], name)?.is_some() => {}
            None => return Err(error(format!("failed to clean; table not found: '{name}'"))),
        }
    }
    if parsed.args.is_empty() {
        for key in table_map(ctx.db, &working)?.into_keys() {
            if !parsed.has("x") {
                let patterns = crate::dolt::ignore::patterns(ctx, &working, &key.0)?;
                if crate::dolt::ignore::is_ignored(&patterns, &key.0, &key.1)? {
                    continue;
                }
            }
            untracked.push(key);
        }
    }
    let staged = table_map(ctx.db, &ctx.txn.staged.clone())?;
    untracked.retain(|key| !staged.contains_key(key));
    let staged_objects = crate::dolt::diff::object_entries(ctx.db, &ctx.txn.staged.clone())?;
    let mut objects = Vec::new();
    for (name, (collection, key, _)) in crate::dolt::diff::object_entries(ctx.db, &working)? {
        let named = parsed.args.is_empty()
            || parsed.args.iter().any(|a| *a == name.1 || *a == format!("{}.{}", name.0, name.1));
        if named && !staged_objects.contains_key(&name) {
            objects.push((collection, key));
        }
    }
    if !parsed.has("dry-run") {
        for (schema, name) in untracked {
            ctx.txn.root.put_table(ctx.db, &schema, &name, None)?;
            ctx.drop_table_triggers(&schema, &name)?;
            ctx.drop_owned_sequences(&schema, &name)?;
        }
        for (collection, key) in objects {
            ctx.txn.root.put_object(ctx.db, collection, &key, None)?;
        }
    }
    Ok(Value::Int8(0))
}

/// COUNT_COMMITS parses dolt_count_commits's arguments.
const COUNT_COMMITS: Parser = Parser {
    command: "count-commits",
    options: &[("from", "f", Kind::Value), ("to", "t", Kind::Value)],
    max_args: None,
};

/// commits_until counts the commits reachable from a commit, in Dolt's log order, before reaching another.
fn commits_until(ctx: &Ctx<'_>, start: store::Hash, target: store::Hash) -> Result<i64> {
    let log = history::log(ctx.db, &[start])?;
    log.iter()
        .position(|c| c.hash == target)
        .map(|n| n as i64)
        .ok_or_else(|| error("no match found to ancestor commit"))
}

/// dolt_count_commits returns how many commits one revision is ahead of and behind another, counted from their merge
/// base, as Dolt's countCommits does.
pub fn dolt_count_commits(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let parsed = COUNT_COMMITS.parse(&strings(args))?;
    let from = parsed.value("from").ok_or_else(|| error("missing from ref"))?.to_string();
    if from.is_empty() {
        return Err(error("empty from ref"));
    }
    let to = parsed.value("to").ok_or_else(|| error("missing to ref"))?.to_string();
    if to.is_empty() {
        return Err(error("empty to ref"));
    }
    let from = history::resolve(ctx.db, ctx.txn.head, &from)?;
    let to = history::resolve(ctx.db, ctx.txn.head, &to)?;
    let base = history::merge_base(ctx.db, from, to)?.ok_or_else(|| error("no common ancestor"))?;
    let (ahead, behind) =
        if from == to { (0, 0) } else { (commits_until(ctx, from, base)?, commits_until(ctx, to, base)?) };
    Ok(Value::Record(vec![Value::Int8(ahead), Value::Int8(behind)]))
}

/// dolt_branch_status counts, for each branch after the first, the commits it has that the first lacks and the
/// commits the first has that it lacks, as Dolt's dolt_branch_status does.
pub fn dolt_branch_status(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let specs = strings(args);
    if specs.is_empty() {
        return Err(crate::dolt::args::argument_count("dolt_branch_status", "at least 1", 0));
    }
    let commits = specs.iter().map(|s| history::resolve(ctx.db, ctx.txn.head, s)).collect::<Result<Vec<_>>>()?;
    let ancestors = |ctx: &mut Ctx<'_>, commit| -> Result<std::collections::HashSet<store::Hash>> {
        Ok(history::log(ctx.db, &[commit])?.into_iter().map(|c| c.hash).collect())
    };
    let base = ancestors(ctx, commits[0])?;
    let mut rows = Vec::new();
    for (spec, &commit) in specs.iter().zip(&commits).skip(1) {
        let (ahead, behind) = if commit == commits[0] {
            (0, 0)
        } else {
            let branch = ancestors(ctx, commit)?;
            (branch.difference(&base).count(), base.difference(&branch).count())
        };
        let count = |n: usize| Value::Numeric(crate::numeric::Numeric::from_i64(n as i64));
        rows.push(Value::Record(vec![Value::Text(spec.clone()), count(ahead), count(behind)]));
    }
    Ok(Value::Set(rows))
}

/// dolt_commit_hash_out commits as dolt_commit does with the remaining arguments and stores the new commit's hash in
/// the session variable that the first argument names.
pub fn dolt_commit_hash_out(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let Some((variable, rest)) = args.split_first() else {
        return Err(error("dolt_commit_hash_out requires the name of a variable for the commit hash"));
    };
    let hash = crate::dolt::procedures::dolt_commit(ctx, rest)?;
    if let Some(text) = hash.output() {
        let name = variable.output().unwrap_or_default();
        ctx.session.settings.set(&name, Some(&text), false, ctx.session.explicit)?;
    }
    Ok(hash)
}

/// RM parses dolt_rm's arguments.
const RM: Parser = Parser { command: "rm", options: &[("cached", "", Kind::Flag)], max_args: None };

/// table_error returns Dolt's error about tables, as its TblError writes it.
fn table_error(tables: &[String], problem: &str) -> PgError {
    error(format!("error: the table(s) {} {problem}", tables.join(", ")))
}

/// dolt_rm removes tables from the staged root, and from the working root too unless only the staged ones are
/// removed, refusing tables with changes that the removal would lose, as Dolt's doDoltRm does.
pub fn dolt_rm(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    ctx.check_branch_access(crate::dolt::branch_control::WRITE)?;
    let parsed = RM.parse(&strings(args))?;
    if parsed.args.is_empty() {
        return Err(error("Nothing specified, nothing removed. Which tables should I remove?"));
    }
    let cached = parsed.has("cached");
    let head = doltdb::root::Root::decode(&read(ctx.db, &ctx.txn.head_root)?)?;
    let (staged, working) = (ctx.txn.staged.clone(), ctx.txn.root.clone());
    let (mut found, mut missing, mut missing_staged, mut unstaged) = (Vec::new(), Vec::new(), Vec::new(), Vec::new());
    for name in &parsed.args {
        let mut in_head = find_table(ctx, &[&head], name)?.map(Target::Table);
        let mut in_staged = find_table(ctx, &[&staged], name)?.map(Target::Table);
        if in_head.is_none() && in_staged.is_none() {
            in_head = find_object(ctx, &[&head], name)?.map(Target::Object);
            in_staged = find_object(ctx, &[&staged], name)?.map(Target::Object);
        }
        if !cached {
            let key = in_staged.clone().or_else(|| in_head.clone());
            let changed = match &key {
                Some(target) => {
                    let current = target.address(ctx, &working)?;
                    let from_staged = target.address(ctx, &staged)?;
                    let from_head = target.address(ctx, &head)?;
                    (in_staged.is_some() || in_head.is_none()) && current != from_staged
                        || (in_head.is_some() || in_staged.is_none()) && current != from_head
                }
                None => false,
            };
            if changed {
                unstaged.push(name.clone());
                continue;
            }
        }
        match (in_staged, in_head) {
            (Some(key), Some(_)) => found.push(key),
            (Some(key), None) if cached => found.push(key),
            (Some(_), None) => missing_staged.push(name.clone()),
            (None, Some(key)) => found.push(key),
            (None, None) => missing.push(name.clone()),
        }
    }
    if !missing.is_empty() {
        return Err(table_error(&missing, "do not exist"));
    }
    if !cached && !unstaged.is_empty() {
        return Err(table_error(&unstaged, "have unstaged changes."));
    }
    if !missing_staged.is_empty() {
        return Err(table_error(&missing_staged, "have changes saved in the index. Use --cached or commit."));
    }
    for target in &found {
        let mut staged = ctx.txn.staged.clone();
        target.remove(ctx, &mut staged)?;
        ctx.txn.staged = staged;
        if !cached {
            let mut working = ctx.txn.root.clone();
            target.remove(ctx, &mut working)?;
            ctx.txn.root = working;
        }
    }
    crate::dolt::procedures::flush(ctx)?;
    Ok(Value::Int8(0))
}

/// Target is what dolt_rm removes: a table by schema and name, or a root object by collection and ID.
#[derive(Clone)]
enum Target {
    Table((String, String)),
    Object((usize, Vec<u8>)),
}

impl Target {
    /// address returns the target's address in a root, or None when the root lacks it.
    fn address(&self, ctx: &mut Ctx<'_>, root: &doltdb::root::Root) -> Result<Option<store::Hash>> {
        Ok(match self {
            Target::Table((schema, table)) => root.table(ctx.db, schema, table)?,
            Target::Object((collection, id)) => {
                root.objects(ctx.db, *collection)?.into_iter().find(|(k, _)| k == id).map(|(_, a)| a)
            }
        })
    }

    /// remove removes the target from a root.
    fn remove(&self, ctx: &mut Ctx<'_>, root: &mut doltdb::root::Root) -> Result<()> {
        match self {
            Target::Table((schema, table)) => root.put_table(ctx.db, schema, table, None)?,
            Target::Object((collection, id)) => root.put_object(ctx.db, *collection, id, None)?,
        }
        Ok(())
    }
}

/// GC parses dolt_gc's arguments.
const GC: Parser = Parser {
    command: "gc",
    options: &[
        ("shallow", "s", Kind::Flag),
        ("full", "f", Kind::Flag),
        ("archive-level", "", Kind::Value),
        ("incremental-file-size", "", Kind::Value),
    ],
    max_args: Some(0),
};

/// dolt_gc removes the chunks that nothing reachable from the database's store root refers to, as Dolt's dolt_gc
/// does, ending the other sessions that use the database with a transaction open.
pub fn dolt_gc(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    ctx.check_branch_access(crate::dolt::branch_control::WRITE)?;
    crate::dolt::procedures::require_admin(ctx)?;
    let parsed = GC.parse(&strings(args))?;
    if parsed.has("shallow") && parsed.has("full") {
        return Err(error("cannot supply both --shallow and --full to dolt_gc: error: invalid usage"));
    }
    if let Some(level) = parsed.value("archive-level") {
        match level.parse::<i64>() {
            Ok(0 | 1) => {}
            Ok(level) => return Err(error(format!("invalid value for archive-level: {level}"))),
            Err(_) => return Err(error(format!("parse error for value for archive-level: {level}"))),
        }
    }
    let mode = match (parsed.has("shallow"), parsed.has("full")) {
        (true, _) => doltdb::database::GcMode::Shallow,
        (_, true) => doltdb::database::GcMode::Full,
        _ => doltdb::database::GcMode::Default,
    };
    crate::dolt::procedures::flush(ctx)?;
    ctx.db.gc(mode)?;
    ctx.session.engine.collected(&ctx.session.database, ctx.session.id);
    Ok(Value::Int8(0))
}

/// dolt_undrop restores a dropped database.
pub fn dolt_undrop(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    crate::dolt::procedures::require_admin(ctx)?;
    match strings(args).as_slice() {
        [] => Err(error(format!(
            "no database name specified. {}",
            crate::engine::undrop_hint(&ctx.session.engine.dropped_databases())
        ))),
        [name] => {
            ctx.session.engine.undrop_database(name)?;
            Ok(Value::Int8(0))
        }
        _ => Err(error(
            "dolt_undrop called with too many arguments: dolt_undrop only accepts one argument - the name of the \
             dropped database to restore",
        )),
    }
}

/// dolt_purge_dropped_databases deletes every dropped database.
pub fn dolt_purge_dropped_databases(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    crate::dolt::procedures::require_admin(ctx)?;
    if !args.is_empty() {
        return Err(error("dolt_purge_dropped_databases does not take any arguments"));
    }
    ctx.session.engine.purge_dropped_databases()?;
    Ok(Value::Int8(0))
}

/// UPDATE_TAG parses dolt_update_column_tag's arguments.
const UPDATE_TAG: Parser = Parser { command: "update-tag", options: &[], max_args: None };

/// dolt_update_column_tag changes a column's tag. Like Dolt under Doltgres, it looks tables up without a schema, so
/// it finds none, since every Doltgres table belongs to one.
pub fn dolt_update_column_tag(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    ctx.check_branch_access(crate::dolt::branch_control::WRITE)?;
    crate::dolt::procedures::require_admin(ctx)?;
    let parsed = UPDATE_TAG.parse(&strings(args))?;
    if parsed.args.len() != 3 {
        return Err(error("incorrect number of arguments: must provide <table> <column> <tag>"));
    }
    if parsed.args[2].parse::<u64>().is_err() {
        return Err(error(format!("failed to parse tag {}", parsed.args[2])));
    }
    Err(error(format!("table {} does not exist", parsed.args[0])))
}

/// dolt_thread_dump returns a dump of the thread running the statement.
pub fn dolt_thread_dump(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    crate::dolt::procedures::require_admin(ctx)?;
    let current = std::thread::current();
    Ok(Value::Text(format!("thread {:?} [running]: {}\n", current.id(), current.name().unwrap_or("connection"))))
}

/// dolt_stats_info returns Dolt's JSON summary of the statistics coordinator, which collects nothing in this server.
pub fn dolt_stats_info(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Text(format!(
        r#"{{"dbCnt":{},"active":false,"storageBucketCnt":0,"cachedBucketCnt":0,"cachedBoundCnt":0,"cachedTemplateCnt":0,"statCnt":0,"backing":"memory","lastUpdate":"0001-01-01T00:00:00Z"}}"#,
        ctx.session.database_names().len()
    )))
}

/// dolt_stats_once returns Dolt's JSON summary of a statistics update, which finds nothing to do in this server.
pub fn dolt_stats_once(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Text(format!(
        r#"{{"dbCnt":{},"bucketWrites":0,"tablesProcessed":0,"tablesSkipped":0,"lastUpdate":"0001-01-01T00:00:00Z"}}"#,
        ctx.session.database_names().len()
    )))
}

/// dolt_stats_ok returns Dolt's `Ok` result for the statistics functions, which have no work to do in this server.
pub fn dolt_stats_ok(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Text("Ok".into()))
}

/// VERIFY_CONSTRAINTS parses dolt_verify_constraints's arguments.
const VERIFY_CONSTRAINTS: Parser = Parser {
    command: "verify-constraints",
    options: &[("all", "a", Kind::Flag), ("output-only", "o", Kind::Flag)],
    max_args: None,
};

/// dolt_verify_constraints records the working root's foreign key violations among the rows changed since the head
/// commit, or among all rows, and returns 1 when a checked table has constraint violations, as Dolt's
/// doDoltConstraintsVerify does.
pub fn dolt_verify_constraints(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    ctx.check_branch_access(crate::dolt::branch_control::WRITE)?;
    let parsed = VERIFY_CONSTRAINTS.parse(&strings(args))?;
    let working = ctx.txn.root.clone();
    let mut checked = Vec::new();
    for name in &parsed.args {
        match find_table(ctx, &[&working], name)? {
            Some(key) => checked.push(key),
            None => return Err(PgError::new("42P01", format!("table not found: {name}"))),
        }
    }
    let comparing = match parsed.has("all") {
        true => doltdb::root::Root::decode(&doltdb::create::empty_root_value(&[]))?,
        false => doltdb::root::Root::decode(&read(ctx.db, &ctx.txn.head_root)?)?,
    };
    let rootish = ctx.db.write_value(working.encode())?;
    let mut verified = working.clone();
    crate::dolt::merge::check_foreign_keys(ctx, &mut verified, &comparing, rootish)?;
    let tables = table_map(ctx.db, &verified)?;
    if !checked.is_empty() {
        let original = table_map(ctx.db, &working)?;
        for (key, address) in &tables {
            if !checked.contains(key) && original.get(key) != Some(address) {
                verified.put_table(ctx.db, &key.0, &key.1, original.get(key).copied())?;
            }
        }
    } else {
        checked = tables.keys().filter(|key| key.0 == "public").cloned().collect();
    }
    let mut violated = false;
    for key in checked {
        let Some(&address) = tables.get(&key) else { continue };
        let table = crate::catalog::table::TableDef::load(ctx.db, &key.0, &key.1, address)?;
        violated |=
            crate::dolt::artifacts::read(ctx.db, &table)?.iter().any(|a| a.kind != crate::dolt::artifacts::CONFLICT);
    }
    if !parsed.has("output-only") {
        ctx.txn.root = verified;
    }
    Ok(Value::Int8(i64::from(violated)))
}

/// dolt_storage_format returns the name of the storage format, which Doltgres databases always use.
pub fn dolt_storage_format(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Text("NEW ( __DOLT__ )".into()))
}
