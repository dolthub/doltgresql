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

//! Dolt's stashes: named lists of saved working changes, each a root value with the commit it was made on, kept in a
//! `refs/stashes/<name>` dataset and applied back with a merge.

use doltdb::root::Root;
use serial::write::{StashFields, write_stash, write_stash_list};
use serial::{Message, Stash, StashList};
use store::Hash;

use crate::dolt::args::{Kind, Parser, error};
use crate::dolt::diff::Delta;
use crate::dolt::history;
use crate::dolt::procedures::{flush, strings};
use crate::error::Result;
use crate::query::Ctx;
use crate::txn::read;
use crate::types::Value;

/// Name is a table or root object by schema and name.
type Name = (String, String);

/// STASH parses dolt_stash's arguments.
const STASH: Parser = Parser {
    command: "stash",
    options: &[("include-untracked", "u", Kind::Flag), ("all", "a", Kind::Flag)],
    max_args: Some(3),
};

/// dataset returns the dataset of a named stash list.
fn dataset(name: &str) -> String {
    format!("refs/stashes/{name}")
}

/// list returns the stashes of a named list, newest first, each with its key, as Dolt's getStashListOrdered orders
/// them by their keys as text.
fn list(ctx: &mut Ctx<'_>, name: &str) -> Result<Option<Vec<(String, Hash)>>> {
    let Some(address) = ctx.db.head(&dataset(name))? else { return Ok(None) };
    let data = read(ctx.db, &address)?;
    let map = StashList::new(Message(&data))?.address_map()?.to_vec();
    let mut entries = ctx.db.address_map_entries(&map)?;
    entries.reverse();
    Ok(Some(entries))
}

/// save writes a named stash list, or deletes it when it has no stashes.
fn save(ctx: &mut Ctx<'_>, name: &str, entries: &[(String, Hash)]) -> Result<()> {
    if entries.is_empty() {
        return Ok(ctx.db.set_heads(&[(dataset(name), None)])?);
    }
    let map = ctx.db.address_map(entries)?;
    let address = ctx.db.write_value(write_stash_list(&map))?;
    Ok(ctx.db.set_heads(&[(dataset(name), Some(address))])?)
}

/// entry returns a named list's stashes, newest first, with the address of the one at an index.
fn entry(ctx: &mut Ctx<'_>, name: &str, index: usize) -> Result<(Vec<(String, Hash)>, Hash)> {
    let entries = list(ctx, name)?.ok_or_else(|| error("No stash entries found."))?;
    if entries.len() <= index {
        return Err(error(format!("fatal: log for 'stash' only has {} entries", entries.len())));
    }
    let address = entries[index].1;
    Ok((entries, address))
}

/// changed_names returns the names of the tables and root objects that deltas change, and of those they add.
fn changed_names(deltas: &[Delta]) -> (Vec<(Name, bool)>, Vec<String>) {
    let mut all = Vec::new();
    let mut added = Vec::new();
    for delta in deltas {
        let Some((name, _)) = delta.to.as_ref().or(delta.from.as_ref()) else { continue };
        if delta.from.is_none() {
            added.push(name.1.clone());
        }
        all.push((name.clone(), delta.object));
    }
    (all, added)
}

/// has_local_changes reports whether the working set has changes to stash, as Dolt's hasLocalChanges decides.
fn has_local_changes(ctx: &mut Ctx<'_>, include_untracked: bool, all: bool) -> Result<bool> {
    let head = Hash::of(&read(ctx.db, &ctx.txn.head_root)?);
    if Hash::of(&ctx.txn.staged.encode()) != head {
        return Ok(true);
    }
    if Hash::of(&ctx.txn.root.encode()) == head {
        return Ok(false);
    }
    if all {
        return Ok(true);
    }
    let (staged, working) = (ctx.txn.staged.clone(), ctx.txn.root.clone());
    if crate::dolt::revert::changed(ctx, &staged, &working)?.is_empty() {
        return Ok(false);
    }
    if include_untracked {
        return Ok(true);
    }
    Ok(crate::dolt::diff::deltas(ctx.db, &staged, &working)?.iter().any(|d| d.from.is_some()))
}

/// push saves the working set's changes as a new stash and resets the working set to the head, as Dolt's
/// doStashPush does, which moves only tables back from the head and so deletes changed root objects from the
/// working root.
fn push(ctx: &mut Ctx<'_>, name: &str, include_untracked: bool, all: bool) -> Result<()> {
    if !has_local_changes(ctx, include_untracked, all)? {
        return Err(error("No local changes to save"));
    }
    crate::dolt::procedures::stage_modified(ctx)?;
    let head = Root::decode(&read(ctx.db, &ctx.txn.head_root)?)?;
    let staged = ctx.txn.staged.clone();
    let (mut stashed, added) = changed_names(&crate::dolt::diff::deltas(ctx.db, &head, &staged)?);
    if include_untracked || all {
        let working = ctx.txn.root.clone();
        let mut names: Vec<(Name, bool)> = Vec::new();
        for root in [&staged, &working] {
            for key in crate::dolt::procedures::table_map(ctx.db, root)?.into_keys() {
                names.push((key, false));
            }
            for key in crate::dolt::diff::object_map(ctx.db, root)?.into_keys() {
                names.push((key, true));
            }
        }
        names.sort();
        names.dedup();
        let full: Vec<String> = names.iter().map(|(n, _)| crate::dolt::diff::full_name(n)).collect();
        crate::dolt::procedures::stage_tables(ctx, &full, !all)?;
        stashed = names;
    }
    let commit = history::load(ctx.db, ctx.txn.head)?;
    let root = ctx.db.write_value(ctx.txn.staged.encode())?;
    let stash = write_stash(&StashFields {
        root,
        head_commit: ctx.txn.head,
        branch_name: doltdb::create::branch_ref(&ctx.txn.branch).into_bytes(),
        description: commit.description.into_bytes(),
        tables_to_stage: Some(added.into_iter().map(String::into_bytes).collect()),
    });
    let address = ctx.db.write_value(stash)?;
    let mut entries = list(ctx, name)?.unwrap_or_default();
    let next = entries.iter().filter_map(|(k, _)| k.parse::<i64>().ok()).next().map_or(0, |last| last + 1);
    entries.push((next.to_string(), address));
    save(ctx, name, &entries)?;
    ctx.txn.staged = head.clone();
    for ((schema, table), object) in stashed {
        match object {
            false => {
                let address = head.table(ctx.db, &schema, &table)?;
                ctx.txn.root.put_table(ctx.db, &schema, &table, address)?;
            }
            true => {
                let found = crate::dolt::diff::object_entries(ctx.db, &ctx.txn.root.clone())?;
                if let Some((collection, key, _)) = found.get(&(schema, table)) {
                    ctx.txn.root.put_object(ctx.db, *collection, key, None)?;
                }
            }
        }
    }
    flush(ctx)
}

/// apply merges a stash into the working root and stages the tables it added, refusing a stash whose changes
/// conflict with the working set, as Dolt's doStashApply does, and returns the list's stashes, newest first.
fn apply(ctx: &mut Ctx<'_>, name: &str, index: usize) -> Result<Vec<(String, Hash)>> {
    let (entries, address) = entry(ctx, name, index)?;
    let data = read(ctx.db, &address)?;
    let stash = Stash::new(Message(&data))?;
    let stash_root = Root::decode(&read(ctx.db, &serial::hash(stash.root()?)?)?)?;
    let head_commit = serial::hash(stash.head_commit()?)?;
    let to_stage: Vec<String> =
        stash.tables_to_stage()?.iter().map(|t| String::from_utf8_lossy(t).into_owned()).collect();
    let parent = Root::decode(&read(ctx.db, &history::load(ctx.db, head_commit)?.root)?)?;
    let working = ctx.txn.root.clone();
    let commits = crate::dolt::merge::Commits { ours: ctx.txn.head, theirs: head_commit, base: head_commit };
    let merged = crate::dolt::merge::merge_roots(ctx, &working, &stash_root, &parent, commits)?;
    let conflicted = conflict_names(ctx, &merged.root, &working, &merged.schema_conflicts)?;
    if !conflicted.is_empty() {
        return Err(error(format!(
            "error: Your local changes to the following tables would be overwritten by applying stash {index}:\n\t{{'{}'}}\nPlease commit your changes or stash them before you merge.\nAborting\nThe stash entry is kept in case you need it again.\n",
            conflicted.join("', '")
        )));
    }
    ctx.txn.root = merged.root;
    if !to_stage.is_empty() {
        crate::dolt::procedures::stage_tables(ctx, &to_stage, false)?;
    }
    flush(ctx)?;
    Ok(entries)
}

/// conflict_names returns the names of the tables and root objects that a merge left in conflict, which it did not
/// have before.
fn conflict_names(ctx: &mut Ctx<'_>, merged: &Root, before: &Root, schema_conflicts: &[Name]) -> Result<Vec<String>> {
    let mut names: Vec<String> = schema_conflicts.iter().map(|n| n.1.clone()).collect();
    let earlier = crate::dolt::procedures::table_map(ctx.db, before)?;
    for (name, address) in crate::dolt::procedures::table_map(ctx.db, merged)? {
        if earlier.get(&name) == Some(&address) {
            continue;
        }
        let table = doltdb::table::Table::decode(&read(ctx.db, &address)?)?;
        if table.artifacts.iter().any(|&b| b != 0) {
            names.push(name.1);
        }
    }
    names.extend(crate::dolt::objmerge::conflict_names(ctx.db, merged)?.into_iter().map(|n| n.1));
    Ok(names)
}

/// index reads the stash index argument, written as a number or as `stash@{n}`.
fn index(arg: Option<&String>) -> Result<usize> {
    let Some(arg) = arg else { return Ok(0) };
    let text = arg.trim_start_matches("stash@{").trim_end_matches('}');
    text.parse().map_err(|_| error(format!("error: {text} is not a valid reference")))
}

/// dolt_stash pushes, pops, applies, drops, and clears stashes.
pub fn dolt_stash(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    ctx.check_branch_access(crate::dolt::branch_control::WRITE)?;
    let parsed = STASH.parse(&strings(args))?;
    if parsed.args.len() < 2 {
        return Err(error("error: invalid arguments. Must provide valid subcommand and stash name"));
    }
    let name = parsed.args[1].clone();
    let index = index(parsed.args.get(2))?;
    match parsed.args[0].as_str() {
        "push" => {
            if parsed.args.len() > 2 {
                return Err(error("error: invalid arguments. Push takes only subcommand and stash name"));
            }
            push(ctx, &name, parsed.has("include-untracked"), parsed.has("all"))?;
        }
        "pop" => {
            let mut entries = apply(ctx, &name, index)?;
            entries.remove(index);
            save(ctx, &name, &entries)?;
        }
        "apply" => _ = apply(ctx, &name, index)?,
        "drop" => {
            let (mut entries, _) = entry(ctx, &name, index)?;
            entries.remove(index);
            save(ctx, &name, &entries)?;
        }
        "clear" => {
            if parsed.args.len() > 2 {
                return Err(error("error: invalid arguments. Clear takes only subcommand and stash name"));
            }
            ctx.db.set_heads(&[(dataset(&name), None)])?;
        }
        other => return Err(error(format!("unknown stash subcommand {other}"))),
    }
    Ok(Value::Int8(0))
}
