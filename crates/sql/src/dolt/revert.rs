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

//! Dolt's revert and cherry-pick, which merge the undoing of a commit, or a commit's own changes, into the session's
//! branch and commit the result unless the merge leaves conflicts.

use std::collections::BTreeMap;

use doltdb::root::Root;
use serial::write::MergeStateFields;
use store::Hash;

use crate::catalog::table::TableDef;
use crate::dolt::args::{Kind, Parser, error};
use crate::dolt::artifacts;
use crate::dolt::history::{self, CommitInfo};
use crate::dolt::merge::Commits;
use crate::dolt::procedures::{commit_meta, commit_staged, flush, move_head, parse_author, strings, table_map};
use crate::error::Result;
use crate::query::Ctx;
use crate::txn::read;
use crate::types::Value;

/// Name is a table by schema and name.
type Name = (String, String);

/// Counts are the numbers of tables that a merge left with data conflicts, schema conflicts, and constraint
/// violations.
#[derive(Clone, Copy, Debug, Default)]
struct Counts {
    data: i64,
    schema: i64,
    violations: i64,
}

impl Counts {
    /// any reports whether any table has conflicts or constraint violations.
    fn any(&self) -> bool {
        self.data > 0 || self.schema > 0 || self.violations > 0
    }
}

/// REVERT parses dolt_revert's arguments.
const REVERT: Parser = Parser {
    command: "revert",
    options: &[("author", "", Kind::Value), ("abort", "", Kind::Flag), ("continue", "", Kind::Flag)],
    max_args: None,
};

/// CHERRY_PICK parses dolt_cherry_pick's arguments.
const CHERRY_PICK: Parser = Parser {
    command: "cherrypick",
    options: &[
        ("abort", "", Kind::Flag),
        ("continue", "", Kind::Flag),
        ("allow-empty", "", Kind::Flag),
        ("skip-verification", "", Kind::Flag),
    ],
    max_args: Some(1),
};

/// outcome returns the result of dolt_revert and dolt_cherry_pick: the new commit's hash and the counts of tables
/// left with conflicts.
fn outcome(hash: &str, counts: Counts) -> Value {
    Value::Record(vec![
        Value::Text(hash.to_string()),
        Value::Int8(counts.data),
        Value::Int8(counts.schema),
        Value::Int8(counts.violations),
    ])
}

/// root_of loads the root value of a commit.
fn root_of(ctx: &mut Ctx<'_>, commit: &CommitInfo) -> Result<Root> {
    Ok(Root::decode(&read(ctx.db, &commit.root)?)?)
}

/// head_root loads the root value of the session's head commit.
fn head_root(ctx: &mut Ctx<'_>) -> Result<Root> {
    Ok(Root::decode(&read(ctx.db, &ctx.txn.head_root)?)?)
}

/// go_quote quotes text as Go's `%q` verb does.
fn go_quote(text: &str) -> String {
    let mut out = String::from("\"");
    for c in text.chars() {
        match c {
            '"' => out.push_str("\\\""),
            '\\' => out.push_str("\\\\"),
            '\n' => out.push_str("\\n"),
            '\r' => out.push_str("\\r"),
            '\t' => out.push_str("\\t"),
            '\u{7}' => out.push_str("\\a"),
            '\u{8}' => out.push_str("\\b"),
            '\u{c}' => out.push_str("\\f"),
            '\u{b}' => out.push_str("\\v"),
            c if (c as u32) < 0x20 || c == '\u{7f}' => out.push_str(&format!("\\x{:02x}", c as u32)),
            c => out.push(c),
        }
    }
    out.push('"');
    out
}

/// artifact_tables returns the working root's tables that have conflicts or constraint violations, with whether
/// each has conflicts and whether it has violations.
fn artifact_tables(ctx: &mut Ctx<'_>) -> Result<BTreeMap<Name, (bool, bool)>> {
    let mut out = BTreeMap::new();
    for ((schema, name), address) in table_map(ctx.db, &ctx.txn.root.clone())? {
        let stored = doltdb::table::Table::decode(&read(ctx.db, &address)?)?;
        if stored.artifacts.iter().all(|&b| b == 0) {
            continue;
        }
        let table = TableDef::load(ctx.db, &schema, &name, address)?;
        let found = artifacts::read(ctx.db, &table)?;
        let conflicts = found.iter().any(|a| a.kind == artifacts::CONFLICT);
        let violations = found.iter().any(|a| a.kind != artifacts::CONFLICT);
        if conflicts || violations {
            out.insert((schema, name), (conflicts, violations));
        }
    }
    Ok(out)
}

/// counts returns the counts of the working root's tables with conflicts or constraint violations, with the merge's
/// schema conflicts.
fn counts(ctx: &mut Ctx<'_>, schema_conflicts: usize) -> Result<Counts> {
    let tables = artifact_tables(ctx)?;
    Ok(Counts {
        data: tables.values().filter(|t| t.0).count() as i64,
        schema: schema_conflicts as i64,
        violations: tables.values().filter(|t| t.1).count() as i64,
    })
}

/// changed returns the tables and root objects that differ between two roots, leaving out new tables that
/// dolt_ignore ignores.
pub fn changed(ctx: &mut Ctx<'_>, from: &Root, to: &Root) -> Result<Vec<Name>> {
    let mut out = Vec::new();
    for delta in crate::dolt::diff::deltas(ctx.db, from, to)? {
        if let (None, Some((name, _)), false) = (&delta.from, &delta.to, delta.object) {
            let patterns = crate::dolt::ignore::patterns(ctx, to, &name.0)?;
            if crate::dolt::ignore::is_ignored(&patterns, &name.1)? {
                continue;
            }
        }
        out.extend(delta.from.into_iter().chain(delta.to).map(|(name, _)| name));
    }
    Ok(out)
}

/// apply merges their root into the working root over a merge base, stages the tables and root objects the merge
/// changed without conflicts, and returns the counts of tables it left with conflicts, as Dolt's revert and
/// cherry-pick do with their merge results.
fn apply(ctx: &mut Ctx<'_>, theirs: &Root, base: &Root, commits: Commits) -> Result<(Counts, Vec<Name>)> {
    let ours = ctx.txn.root.clone();
    let mut merged = crate::dolt::merge::merge_roots(ctx, &ours, theirs, base, commits)?;
    crate::dolt::merge::check_foreign_keys(ctx, &mut merged.root, base, commits.theirs)?;
    ctx.txn.root = merged.root.clone();
    let unmerged = artifact_tables(ctx)?;
    let before = table_map(ctx.db, &ours)?;
    let after = table_map(ctx.db, &merged.root)?;
    for name in before.keys().chain(after.keys()) {
        if before.get(name) != after.get(name)
            && !unmerged.contains_key(name)
            && !merged.schema_conflicts.contains(name)
        {
            ctx.txn.staged.put_table(ctx.db, &name.0, &name.1, after.get(name).copied())?;
        }
    }
    for schema in &merged.root.schemas {
        if !ctx.txn.staged.schemas.contains(schema) {
            ctx.txn.staged.schemas.push(schema.clone());
        }
    }
    ctx.txn.staged.schemas.sort();
    if merged.root.foreign_keys != ours.foreign_keys {
        ctx.txn.staged.foreign_keys = merged.root.foreign_keys.clone();
    }
    for collection in 0..merged.root.root_objects.len() {
        if merged.root.root_objects[collection] != ours.root_objects[collection] {
            ctx.txn.staged.root_objects[collection] = merged.root.root_objects[collection];
        }
    }
    let counts = Counts {
        data: unmerged.values().filter(|t| t.0).count() as i64,
        schema: merged.schema_conflicts.len() as i64,
        violations: unmerged.values().filter(|t| t.1).count() as i64,
    };
    Ok((counts, merged.schema_conflicts))
}

/// start_merge records an interrupted revert or cherry-pick in the working set, so that `--continue` and `--abort`
/// can finish it.
fn start_merge(ctx: &mut Ctx<'_>, pre_working: &Root, from: Hash, spec: &str, unmergable: &[Name]) -> Result<()> {
    let pre_working_root = ctx.db.write_value(pre_working.encode())?;
    ctx.txn.merge = Some(MergeStateFields {
        pre_working_root,
        from_commit: from,
        from_commit_spec: spec.as_bytes().to_vec(),
        unmergable_tables: unmergable.iter().map(|n| n.1.as_bytes().to_vec()).collect(),
        is_cherry_pick: false,
        is_revert: false,
        pre_merge_head_commit: Some(ctx.txn.head),
        pending_commit_hashes: Vec::new(),
    });
    Ok(())
}

/// abort_merge restores the working set from before an interrupted merge, as Dolt's AbortMerge does.
fn abort_merge(ctx: &mut Ctx<'_>, merge: &MergeStateFields) -> Result<()> {
    ctx.txn.root = Root::decode(&read(ctx.db, &merge.pre_working_root)?)?;
    ctx.txn.staged = head_root(ctx)?;
    ctx.txn.merge = None;
    Ok(())
}

/// dolt_revert commits the undoing of each named commit in turn, stopping at the first that leaves conflicts.
pub fn dolt_revert(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let parsed = REVERT.parse(&strings(args))?;
    if parsed.has("abort") && parsed.has("continue") {
        return Err(error("error: --continue and --abort are mutually exclusive"));
    }
    if parsed.has("abort") {
        abort_revert(ctx)?;
        return Ok(outcome("", Counts::default()));
    }
    let author = parsed.value("author").map(parse_author).transpose()?;
    if parsed.has("continue") {
        return continue_revert(ctx, author);
    }
    if parsed.args.is_empty() {
        return Err(error("error: nothing specified to revert"));
    }
    let mut commits = Vec::with_capacity(parsed.args.len());
    for spec in &parsed.args {
        commits.push(history::resolve(ctx.db, ctx.txn.head, spec)?);
    }
    let series_head = ctx.txn.head;
    let mut last = String::new();
    for (i, &commit) in commits.iter().enumerate() {
        let info = history::load(ctx.db, commit)?;
        if blocks_revert(ctx, &info)? {
            return Err(error(
                "error: Your local changes would be overwritten by revert.\nhint: Please commit your changes before you \
                 revert.",
            ));
        }
        match revert_one(ctx, &info, series_head, &commits[i + 1..], author.clone())? {
            Ok(hash) => last = hash.to_string(),
            Err(counts) => return Ok(outcome("", counts)),
        }
    }
    Ok(outcome(&last, Counts::default()))
}

/// blocks_revert reports whether the working set's changes keep a commit from being reverted: any staged change
/// does, and so does an unstaged change to a table that the commit changed, as Dolt's
/// dirtyTablesConflictWithRevert decides.
fn blocks_revert(ctx: &mut Ctx<'_>, commit: &CommitInfo) -> Result<bool> {
    let head = head_root(ctx)?;
    let (staged, working) = (ctx.txn.staged.clone(), ctx.txn.root.clone());
    if !crate::dolt::diff::deltas(ctx.db, &head, &staged)?.is_empty() {
        return Ok(true);
    }
    let dirty = changed(ctx, &staged, &working)?;
    if dirty.is_empty() {
        return Ok(false);
    }
    let Some(&parent) = commit.parents.first() else { return Ok(true) };
    let parent_root = root_of(ctx, &history::load(ctx.db, parent)?)?;
    let commit_root = root_of(ctx, commit)?;
    let touched: Vec<Name> = crate::dolt::diff::deltas(ctx.db, &parent_root, &commit_root)?
        .into_iter()
        .flat_map(|d| d.from.into_iter().chain(d.to).map(|(name, _)| name))
        .collect();
    Ok(touched.iter().any(|name| dirty.contains(name)))
}

/// revert_one merges the undoing of a commit into the working root and commits it, or records the revert in the
/// working set and returns the counts of tables with conflicts when the merge leaves any, as Dolt's
/// applySingleRevert does.
fn revert_one(
    ctx: &mut Ctx<'_>,
    commit: &CommitInfo,
    series_head: Hash,
    pending: &[Hash],
    author: Option<(String, String)>,
) -> Result<std::result::Result<Hash, Counts>> {
    let Some(&parent) = commit.parents.first() else {
        return Err(error(format!("cannot revert commit with no parents ({})", commit.hash)));
    };
    let base = root_of(ctx, commit)?;
    let theirs = root_of(ctx, &history::load(ctx.db, parent)?)?;
    let pre_working = ctx.txn.root.clone();
    let commits = Commits { ours: ctx.txn.head, theirs: parent, base: commit.hash };
    let (counts, unmergable) = apply(ctx, &theirs, &base, commits)?;
    if counts.any() {
        start_merge(ctx, &pre_working, commit.hash, &commit.hash.to_string(), &unmergable)?;
        if let Some(merge) = ctx.txn.merge.as_mut() {
            merge.is_revert = true;
            merge.pre_merge_head_commit = Some(series_head);
            merge.pending_commit_hashes = pending.iter().map(|h| h.to_string().into_bytes()).collect();
        }
        return Ok(Err(counts));
    }
    if Hash::of(&ctx.txn.staged.encode()) == ctx.txn.head_root {
        return Err(error("nothing to commit"));
    }
    let mut meta = commit_meta(ctx, &format!("Revert {}", go_quote(&commit.description)));
    if let Some((name, email)) = author {
        meta.name = name;
        meta.email = email;
    }
    Ok(Ok(commit_staged(ctx, Vec::new(), meta)?))
}

/// continue_revert commits a revert whose conflicts have been resolved and then reverts the rest of its series, as
/// Dolt's ContinueRevert does.
fn continue_revert(ctx: &mut Ctx<'_>, author: Option<(String, String)>) -> Result<Value> {
    let Some(merge) = ctx.txn.merge.clone().filter(|m| m.is_revert) else {
        return Err(error("error: There is no revert in progress"));
    };
    let counts = counts(ctx, merge.unmergable_tables.len())?;
    if counts.any() {
        return Ok(outcome("", counts));
    }
    let (staged, working) = (ctx.txn.staged.clone(), ctx.txn.root.clone());
    if !changed(ctx, &staged, &working)?.is_empty() {
        return Err(error("error: cannot continue revert with unstaged changes"));
    }
    let reverted = history::load(ctx.db, merge.from_commit)?;
    let series_head = merge.pre_merge_head_commit.unwrap_or(ctx.txn.head);
    ctx.txn.merge = None;
    let mut meta = commit_meta(ctx, &format!("Revert {}", go_quote(&reverted.description)));
    if let Some((_, email)) = author.as_ref().filter(|(name, email)| !name.is_empty() && !email.is_empty()) {
        meta.committer_name = Some(email.clone());
        meta.committer_email = Some(meta.email.clone());
        meta.name = email.clone();
        meta.email = email.clone();
    }
    let mut last = commit_staged(ctx, Vec::new(), meta)?;
    let pending: Vec<Hash> = merge
        .pending_commit_hashes
        .iter()
        .map(|h| history::resolve(ctx.db, ctx.txn.head, &String::from_utf8_lossy(h)))
        .collect::<Result<_>>()?;
    for (i, &commit) in pending.iter().enumerate() {
        let info = history::load(ctx.db, commit)?;
        match revert_one(ctx, &info, series_head, &pending[i + 1..], author.clone())? {
            Ok(hash) => last = hash,
            Err(counts) => return Ok(outcome("", counts)),
        }
    }
    Ok(outcome(&last.to_string(), Counts::default()))
}

/// abort_revert restores the working set and the branch's head from before a revert series, as Dolt's AbortRevert
/// does.
fn abort_revert(ctx: &mut Ctx<'_>) -> Result<()> {
    let Some(merge) = ctx.txn.merge.clone().filter(|m| m.is_revert) else {
        return Err(error("error: There is no revert in progress"));
    };
    abort_merge(ctx, &merge)?;
    if let Some(head) = merge.pre_merge_head_commit {
        move_head(ctx, head)?;
        ctx.txn.root = head_root(ctx)?;
        ctx.txn.staged = ctx.txn.root.clone();
        flush(ctx)?;
    }
    Ok(())
}

/// dolt_cherry_pick commits a commit's changes onto the session's branch.
pub fn dolt_cherry_pick(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let parsed = CHERRY_PICK.parse(&strings(args)).map_err(|e| {
        if e.message.contains("too many positional arguments") {
            error("cherry-picking multiple commits is not supported yet.")
        } else {
            e
        }
    })?;
    if parsed.has("abort") && parsed.has("continue") {
        return Err(error("error: --continue and --abort are mutually exclusive"));
    }
    if parsed.has("abort") {
        let Some(merge) = ctx.txn.merge.clone() else {
            return Err(error("error: There is no cherry-pick merge to abort"));
        };
        abort_merge(ctx, &merge)?;
        return Ok(outcome("", Counts::default()));
    }
    if parsed.has("continue") {
        return continue_cherry_pick(ctx);
    }
    let Some(spec) = parsed.args.first().filter(|s| !s.is_empty()) else {
        return Err(error("cannot cherry-pick empty string"));
    };
    let head = head_root(ctx)?;
    let (staged, working) = (ctx.txn.staged.clone(), ctx.txn.root.clone());
    if !crate::dolt::diff::deltas(ctx.db, &head, &staged)?.is_empty() || !changed(ctx, &staged, &working)?.is_empty() {
        return Err(error("cannot cherry-pick with uncommitted changes"));
    }
    let hash = history::resolve(ctx.db, ctx.txn.head, spec)?;
    let commit = history::load(ctx.db, hash)?;
    let parent = match commit.parents.as_slice() {
        [parent] => *parent,
        [] => return Err(error("cherry-picking a commit without parents is not supported")),
        _ => return Err(error("cherry-picking a merge commit is not supported")),
    };
    let parent_info = history::load(ctx.db, parent)?;
    let empty = commit.root == parent_info.root;
    if empty && !parsed.has("allow-empty") {
        return Err(error("The previous cherry-pick commit is empty. Use --allow-empty to cherry-pick empty commits."));
    }
    let (theirs, base) = (root_of(ctx, &commit)?, root_of(ctx, &parent_info)?);
    let commits = Commits { ours: ctx.txn.head, theirs: commit.hash, base: parent };
    let (counts, unmergable) = apply(ctx, &theirs, &base, commits)?;
    if Hash::of(&ctx.txn.root.encode()) == ctx.txn.head_root && !empty {
        ctx.txn.root = working;
        ctx.txn.staged = staged;
        return Err(error("no changes were made, nothing to commit"));
    }
    if counts.any() {
        start_merge(ctx, &working, commit.hash, spec, &unmergable)?;
        if let Some(merge) = ctx.txn.merge.as_mut() {
            merge.is_cherry_pick = true;
        }
        return Ok(outcome("", counts));
    }
    if !empty && Hash::of(&ctx.txn.staged.encode()) == ctx.txn.head_root {
        return Err(error("nothing to commit"));
    }
    let hash = commit_staged(ctx, Vec::new(), picked_meta(ctx, &commit))?;
    Ok(outcome(&hash.to_string(), Counts::default()))
}

/// picked_meta returns the metadata of a cherry-picked commit, which keeps the original's author, date, and
/// message, with the session's user as its committer.
fn picked_meta(ctx: &Ctx<'_>, original: &CommitInfo) -> doltdb::database::CommitMeta {
    let mut meta = commit_meta(ctx, &original.description);
    if (original.name.as_str(), original.email.as_str()) != (meta.name.as_str(), meta.email.as_str()) {
        meta.committer_name = Some(std::mem::replace(&mut meta.name, original.name.clone()));
        meta.committer_email = Some(std::mem::replace(&mut meta.email, original.email.clone()));
    }
    meta.author_millis = original.author_millis;
    meta
}

/// continue_cherry_pick commits a cherry-pick whose conflicts have been resolved, as Dolt's ContinueCherryPick does.
fn continue_cherry_pick(ctx: &mut Ctx<'_>) -> Result<Value> {
    let Some(merge) = ctx.txn.merge.clone() else {
        return Err(error("error: There is no cherry-pick merge to continue"));
    };
    let counts = counts(ctx, merge.unmergable_tables.len())?;
    if counts.any() {
        return Ok(outcome("", counts));
    }
    if ctx.txn.staged.encode() != ctx.txn.root.encode() {
        return Err(error("error: cannot continue cherry-pick with unstaged changes"));
    }
    let original = history::load(ctx.db, merge.from_commit)?;
    if Hash::of(&ctx.txn.staged.encode()) == ctx.txn.head_root {
        return Err(error("error: no changes to commit"));
    }
    ctx.txn.merge = None;
    let hash = commit_staged(ctx, Vec::new(), picked_meta(ctx, &original))?;
    Ok(outcome(&hash.to_string(), Counts::default()))
}
