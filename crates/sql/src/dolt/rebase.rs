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

//! dolt_rebase: replaying a branch's commits onto another commit on a working branch named `dolt_rebase_<branch>`,
//! following a plan in the `dolt.rebase` table that an interactive rebase lets the user edit, as Dolt's
//! dolt_rebase.go does.

use doltdb::root::Root;
use serial::write::RebaseStateFields;
use store::Hash;

use crate::catalog::table::TableDef;
use crate::dolt::args::{Kind, Parser, error};
use crate::dolt::history::{self, CommitInfo};
use crate::dolt::procedures::{create_branch_at, delete_branch, flush, new_branch, strings, table_map};
use crate::dolt::revert::{PickOptions, Picked};
use crate::error::Result;
use crate::query::Ctx;
use crate::txn::{Txn, read};
use crate::types::Value;

/// REBASE parses dolt_rebase's arguments.
const REBASE: Parser = Parser {
    command: "rebase",
    options: &[
        ("empty", "", Kind::Value),
        ("abort", "", Kind::Flag),
        ("continue", "", Kind::Flag),
        ("interactive", "i", Kind::Flag),
        ("skip-verification", "", Kind::Flag),
    ],
    max_args: Some(1),
};

/// DROP_EMPTY and KEEP_EMPTY are Dolt's EmptyCommitHandling values that a rebase stores.
const DROP_EMPTY: u8 = 1;
const KEEP_EMPTY: u8 = 2;

/// SCHEMA and TABLE are where the rebase plan lives on the working branch.
const SCHEMA: &str = "dolt";
const TABLE: &str = "rebase";

/// DEFINITION is the rebase plan table's columns, as Doltgres defines them.
const DEFINITION: &str = "(rebase_order real PRIMARY KEY, action varchar(6) NOT NULL, commit_hash text NOT NULL, commit_message text NOT NULL)";

/// ACTIONS are the actions a plan step may take.
const ACTIONS: [&str; 6] = ["drop", "pick", "reword", "squash", "fixup", "edit"];

/// Step is a step of a rebase plan.
#[derive(Clone, Debug)]
struct Step {
    order: f32,
    action: String,
    commit_hash: String,
    commit_message: String,
}

/// outcome returns dolt_rebase's result of a status and a message.
fn outcome(status: i64, message: impl Into<String>) -> Value {
    Value::Record(vec![Value::Int8(status), Value::Text(message.into())])
}

/// dolt_rebase starts, continues, or aborts a rebase of the session's branch, as Dolt's doDoltRebase does.
pub fn dolt_rebase(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    ctx.check_branch_access(crate::dolt::branch_control::WRITE)?;
    let parsed =
        REBASE.parse(&strings(args)).map_err(|e| match e.message.contains("too many positional arguments") {
            true => error("rebase takes at most one positional argument."),
            false => e,
        })?;
    if parsed.has("abort") {
        abort(ctx)?;
        return Ok(outcome(0, "Interactive rebase aborted"));
    }
    if parsed.has("continue") {
        return continue_rebase(ctx);
    }
    let becomes_empty = match parsed.value("empty") {
        None => DROP_EMPTY,
        Some(v) if v.eq_ignore_ascii_case("keep") => KEEP_EMPTY,
        Some(v) if v.eq_ignore_ascii_case("drop") => DROP_EMPTY,
        Some(v) => {
            return Err(error(format!(
                "unsupported option for the empty flag ({v}); only 'keep' or 'drop' are allowed"
            )));
        }
    };
    let upstream = match parsed.args.as_slice() {
        [] => return Err(error("not enough args")),
        [upstream] => upstream.clone(),
        _ => return Err(error("too many args")),
    };
    start(ctx, &upstream, becomes_empty, parsed.has("skip-verification"))?;
    if !parsed.has("interactive") {
        return continue_rebase(ctx);
    }
    Ok(outcome(
        0,
        format!(
            "interactive rebase started on branch {}; adjust the rebase plan in the dolt_rebase table, then continue \
             rebasing by calling dolt_rebase('--continue')",
            ctx.session.branch
        ),
    ))
}

/// switch_branch writes the session's working set and moves the session's transaction to another branch, as Dolt's
/// SwitchWorkingSet does.
fn switch_branch(ctx: &mut Ctx<'_>, branch: &str) -> Result<()> {
    flush(ctx)?;
    let (handle, sequences, database) = (ctx.txn.handle.clone(), ctx.txn.sequences.clone(), ctx.txn.database.clone());
    let mut txn = Txn::begin_locked(ctx.db, handle, sequences, &database, branch)?;
    txn.started = ctx.txn.started;
    *ctx.txn = txn;
    ctx.session.branch = branch.to_string();
    Ok(())
}

/// head_root returns the root value of the session's head commit.
fn head_root(ctx: &mut Ctx<'_>) -> Result<Root> {
    Ok(Root::decode(&read(ctx.db, &ctx.txn.head_root)?)?)
}

/// has_changes returns whether the session's branch has staged changes and whether it has unstaged ones, leaving out
/// new tables that dolt_ignore ignores, as Dolt's workingSetStatus does.
fn has_changes(ctx: &mut Ctx<'_>) -> Result<(bool, bool)> {
    let (head, staged, working) = (head_root(ctx)?, ctx.txn.staged.clone(), ctx.txn.root.clone());
    let staged_changes = !crate::dolt::diff::deltas(ctx.db, &head, &staged)?.is_empty();
    let unstaged_changes = !crate::dolt::revert::changed(ctx, &staged, &working)?.is_empty();
    Ok((staged_changes, unstaged_changes))
}

/// start starts a rebase onto an upstream commit: it creates the working branch there, records the rebase in its
/// working set, and writes the default plan of picking each of the branch's commits, as Dolt's startRebase does.
fn start(ctx: &mut Ctx<'_>, upstream: &str, becomes_empty: u8, skip_verification: bool) -> Result<()> {
    if upstream.is_empty() {
        return Err(error("no upstream branch specified"));
    }
    if ctx.txn.merge.is_some() {
        return Err(error(
            "unable to start rebase while a merge is in progress – abort the current merge before proceeding",
        ));
    }
    if ctx.txn.rebase.is_some() {
        return Err(error(
            "unable to start rebase while another rebase is in progress – abort the current rebase before proceeding",
        ));
    }
    if has_changes(ctx)? != (false, false) {
        return Err(error("cannot start a rebase with uncommitted changes"));
    }
    let branch = ctx.session.branch.clone();
    let start_commit = ctx.txn.head;
    let upstream_commit = history::resolve(ctx.db, ctx.txn.head, upstream)?;
    let working_branch = format!("dolt_rebase_{branch}");
    create_branch_at(ctx, &working_branch, upstream, false)?;
    flush(ctx)?;
    let pre_working_root = ctx.db.write_value(ctx.txn.root.encode())?;
    switch_branch(ctx, &working_branch)?;
    let onto = Root::decode(&read(ctx.db, &history::load(ctx.db, upstream_commit)?.root)?)?;
    ctx.txn.root = onto.clone();
    ctx.txn.staged = onto;
    ctx.txn.rebase = Some(RebaseStateFields {
        pre_working_root,
        onto_commit: upstream_commit,
        branch: branch.into_bytes(),
        commit_becomes_empty_handling: becomes_empty,
        empty_commit_handling: KEEP_EMPTY,
        last_attempted_step: 0.0,
        rebasing_started: false,
        skip_verification,
    });
    let plan = default_plan(ctx, start_commit, upstream_commit)?;
    if plan.is_empty() {
        abort(ctx)?;
        return Err(error("didn't identify any commits!"));
    }
    save_plan(ctx, &plan)
}

/// default_plan returns a plan that picks each single-parent commit reachable from the branch's head but not from
/// the upstream commit, oldest first, as Dolt's CreateDefaultRebasePlan does.
fn default_plan(ctx: &mut Ctx<'_>, start: Hash, upstream: Hash) -> Result<Vec<Step>> {
    let excluded: std::collections::HashSet<Hash> =
        history::log(ctx.db, &[upstream])?.into_iter().map(|c| c.hash).collect();
    let commits: Vec<CommitInfo> = history::log(ctx.db, &[start])?
        .into_iter()
        .filter(|c| !excluded.contains(&c.hash) && c.parents.len() == 1)
        .collect();
    Ok(commits
        .into_iter()
        .rev()
        .enumerate()
        .map(|(i, c)| Step {
            order: (i + 1) as f32,
            action: "pick".into(),
            commit_hash: c.hash.to_string(),
            commit_message: c.description,
        })
        .collect())
}

/// save_plan creates the plan table on the working branch and writes the plan's steps to it.
fn save_plan(ctx: &mut Ctx<'_>, plan: &[Step]) -> Result<()> {
    let table = crate::dolt::tables::create_backing(ctx, SCHEMA, TABLE, DEFINITION)?;
    let rows: Vec<Vec<Value>> = plan
        .iter()
        .map(|s| {
            vec![
                Value::Float4(s.order),
                Value::Text(s.action.clone()),
                Value::Text(s.commit_hash.clone()),
                Value::Text(s.commit_message.clone()),
            ]
        })
        .collect();
    crate::dml::write_rows(ctx, &table, &rows)
}

/// load_plan reads the plan table in step order and checks the plan as Dolt's ValidateRebasePlan does.
fn load_plan(ctx: &mut Ctx<'_>) -> Result<Vec<Step>> {
    let table: TableDef =
        ctx.txn.table(ctx.db, SCHEMA, TABLE)?.ok_or_else(|| error("unable to find dolt_rebase table"))?;
    let mut plan = Vec::new();
    for row in crate::query::scan(ctx.db, &table)? {
        let text = |v: &Value| v.output().unwrap_or_default();
        let order = match &row[0] {
            Value::Float4(f) => *f,
            other => return Err(error(format!("invalid order value in rebase plan: {}", text(other)))),
        };
        let action = text(&row[1]);
        if !ACTIONS.contains(&action.as_str()) {
            return Err(error(format!("invalid enum value in rebase plan: {action} (string)")));
        }
        plan.push(Step { order, action, commit_hash: text(&row[2]), commit_message: text(&row[3]) });
    }
    plan.sort_by(|a, b| a.order.total_cmp(&b.order));
    let (mut seen_pick, mut seen_reword) = (false, false);
    for (i, step) in plan.iter().enumerate() {
        if i > 0 && plan[i - 1].order >= step.order {
            return Err(error("invalid rebase plan: rebase order must be ascending"));
        }
        match step.action.as_str() {
            "pick" => seen_pick = true,
            "reword" => seen_reword = true,
            "fixup" | "squash" if !seen_pick && !seen_reword => {
                return Err(error(
                    "invalid rebase plan: squash and fixup actions must appear after a pick or reword action",
                ));
            }
            _ => {}
        }
        if !history::is_hash(&step.commit_hash) {
            return Err(error(format!("invalid commit hash: {}", step.commit_hash)));
        }
        if let Err(err) = history::resolve(ctx.db, ctx.txn.head, &step.commit_hash) {
            return Err(error(format!("unable to resolve commit hash {}: {}", step.commit_hash, err.message)));
        }
    }
    Ok(plan)
}

/// abort ends the rebase in progress: it deletes the working branch and returns the session to the rebased branch,
/// as Dolt's abortRebase does.
fn abort(ctx: &mut Ctx<'_>) -> Result<()> {
    let Some(state) = ctx.txn.rebase.take() else { return Err(error("no rebase in progress")) };
    let pre_working = Root::decode(&read(ctx.db, &state.pre_working_root)?)?;
    ctx.txn.root = pre_working.clone();
    ctx.txn.staged = pre_working;
    flush(ctx)?;
    let working_branch = ctx.session.branch.clone();
    delete_branch(ctx.db, &working_branch)?;
    let branch = String::from_utf8_lossy(&state.branch).into_owned();
    let (handle, sequences, database) = (ctx.txn.handle.clone(), ctx.txn.sequences.clone(), ctx.txn.database.clone());
    let mut txn = Txn::begin_locked(ctx.db, handle, sequences, &database, &branch)?;
    txn.started = ctx.txn.started;
    *ctx.txn = txn;
    ctx.session.branch = branch;
    Ok(())
}

/// conflicted_tables returns the names of the working root's tables with data conflicts.
fn conflicted_tables(ctx: &mut Ctx<'_>) -> Result<Vec<String>> {
    let mut out = Vec::new();
    for ((schema, name), address) in table_map(ctx.db, &ctx.txn.root.clone())? {
        let table = TableDef::load(ctx.db, &schema, &name, address)?;
        let found = crate::dolt::artifacts::read(ctx.db, &table)?;
        if found.iter().any(|a| a.kind == crate::dolt::artifacts::CONFLICT) {
            out.push(crate::dolt::diff::full_name(&(schema, name)));
        }
    }
    Ok(out)
}

/// record_step records the plan step that the rebase is attempting.
fn record_step(ctx: &mut Ctx<'_>, order: f32) -> Result<()> {
    if let Some(state) = ctx.txn.rebase.as_mut() {
        state.last_attempted_step = order;
        state.rebasing_started = true;
    }
    flush(ctx)
}

/// options returns how a plan step picks its commit, as Dolt's createCherryPickOptionsForRebaseStep does.
fn options(ctx: &mut Ctx<'_>, step: &Step, state: &RebaseStateFields) -> Result<PickOptions> {
    let drop_empty = state.commit_becomes_empty_handling == DROP_EMPTY;
    Ok(match step.action.as_str() {
        "reword" => PickOptions { amend: false, message: Some(step.commit_message.clone()), drop_empty },
        "squash" => {
            let head = history::load(ctx.db, ctx.txn.head)?;
            let next = history::resolve(ctx.db, ctx.txn.head, &step.commit_hash)?;
            let next = history::load(ctx.db, next)?;
            let message = format!("{}\n\n{}", head.description, next.description);
            PickOptions { amend: true, message: Some(message), drop_empty }
        }
        "fixup" => PickOptions { amend: true, message: None, drop_empty },
        _ => PickOptions { amend: false, message: None, drop_empty },
    })
}

/// commit_staged_step commits the changes a user staged for the plan step the rebase stopped at, as Dolt's
/// commitManuallyStagedChangesForStep does.
fn commit_staged_step(ctx: &mut Ctx<'_>, step: &Step, state: &RebaseStateFields) -> Result<()> {
    let options = options(ctx, step, state)?;
    let original = history::resolve(ctx.db, ctx.txn.head, &step.commit_hash)?;
    let original = history::load(ctx.db, original)?;
    let head = history::load(ctx.db, ctx.txn.head)?;
    let message = match &options.message {
        Some(message) if !message.is_empty() => message.clone(),
        _ if step.action != "fixup" => step.commit_message.clone(),
        _ => head.description.clone(),
    };
    ctx.txn.merge = None;
    crate::dolt::revert::commit_rebase_step(ctx, &original, &head, message, &options)
}

/// setting_on reports whether a boolean setting is on.
fn setting_on(ctx: &Ctx<'_>, name: &str) -> bool {
    ctx.session.settings.get(name).is_some_and(|v| matches!(v.to_ascii_lowercase().as_str(), "on" | "1" | "true"))
}

/// process_step runs one plan step, returning the message to pause with for an edit step.
fn process_step(ctx: &mut Ctx<'_>, step: &Step, state: &RebaseStateFields) -> Result<Option<String>> {
    if step.action == "drop" {
        return Ok(None);
    }
    let options = options(ctx, step, state)?;
    match crate::dolt::revert::pick_for_rebase(ctx, &step.commit_hash, &options)? {
        Picked::SchemaConflicts => {
            abort(ctx)?;
            Err(error(format!(
                "schema conflict detected while rebasing commit {}. the rebase has been automatically aborted",
                step.commit_hash
            )))
        }
        Picked::Conflicts => {
            let allow = setting_on(ctx, "dolt_allow_commit_conflicts");
            if !ctx.session.explicit && !allow {
                abort(ctx)?;
                return Err(error(
                    "data conflicts from rebase, but session settings do not allow preserving conflicts, so they \
                     cannot be resolved. The rebase has been aborted. Set @@autocommit to 0 or set \
                     @@dolt_allow_commit_conflicts to 1 and try the rebase again to resolve the conflicts.",
                ));
            }
            if allow {
                flush(ctx)?;
            }
            Err(error(format!(
                "data conflict detected while rebasing commit {} ({}). \n\nResolve the conflicts and remove them from \
                 the dolt_conflicts_<table> tables, then continue the rebase by calling dolt_rebase('--continue')",
                step.commit_hash, step.commit_message
            )))
        }
        Picked::Done if step.action == "edit" => Ok(Some(format!(
            "edit action paused at commit {} ({}). \n\nYou can now modify the working directory and stage changes. \
             When ready, continue the rebase by calling dolt_rebase('--continue')",
            step.commit_hash, step.commit_message
        ))),
        Picked::Done => Ok(None),
    }
}

/// continue_rebase runs the plan's remaining steps and, when they finish, moves the rebased branch to the working
/// branch's head, as Dolt's continueRebase does.
fn continue_rebase(ctx: &mut Ctx<'_>) -> Result<Value> {
    if ctx.txn.rebase.is_none() {
        return Err(error("no rebase in progress"));
    }
    let conflicted = conflicted_tables(ctx)?;
    if !conflicted.is_empty() {
        return Err(error(format!(
            "conflicts detected in tables {}; resolve conflicts before continuing the rebase",
            conflicted.join(", ")
        )));
    }
    flush(ctx)?;
    let plan = load_plan(ctx)?;
    for step in &plan {
        let Some(state) = ctx.txn.rebase.clone() else { return Err(error("no rebase in progress")) };
        let (staged, unstaged) = has_changes(ctx)?;
        if !state.rebasing_started && (staged || unstaged) {
            return Err(error("cannot start a rebase with uncommitted changes"));
        }
        if state.rebasing_started && step.order < state.last_attempted_step {
            continue;
        }
        if unstaged {
            return Err(error(
                "cannot continue a rebase with unstaged changes. Use dolt_add() to stage tables and then continue \
                 the rebase",
            ));
        }
        if state.rebasing_started && step.order == state.last_attempted_step && staged {
            commit_staged_step(ctx, step, &state)?;
            continue;
        }
        if !state.rebasing_started || step.order > state.last_attempted_step {
            record_step(ctx, step.order)?;
            if let Some(message) = process_step(ctx, step, &state)? {
                flush(ctx)?;
                return Ok(outcome(0, message));
            }
        }
    }
    finish(ctx)
}

/// finish moves the rebased branch to the working branch's head, keeping the branch's untracked tables, returns the
/// session to it, and deletes the working branch.
fn finish(ctx: &mut Ctx<'_>) -> Result<Value> {
    let Some(state) = ctx.txn.rebase.clone() else { return Err(error("no rebase in progress")) };
    let branch = String::from_utf8_lossy(&state.branch).into_owned();
    let working_branch = format!("dolt_rebase_{branch}");
    let pre_working = Root::decode(&read(ctx.db, &state.pre_working_root)?)?;
    let current = ctx.branch_root(&branch)?.ok_or_else(|| error(format!("branch not found: {branch}")))?;
    for name in crate::dolt::revert::changed(ctx, &pre_working, &current)? {
        let patterns = crate::dolt::ignore::patterns(ctx, &current, &name.0)?;
        if !crate::dolt::ignore::is_ignored(&patterns, &name.0, &name.1)? {
            return Err(error(format!("rebase aborted due to changes in branch {branch}")));
        }
    }
    let staged_before = match ctx.db.head(&doltdb::create::working_set_ref(&branch))? {
        Some(address) => {
            let data = read(ctx.db, &address)?;
            let ws = serial::WorkingSet::new(serial::Message(&data))?;
            let staged = ws.staged_root()?.unwrap_or(ws.working_root()?);
            Root::decode(&read(ctx.db, &staged)?)?
        }
        None => current.clone(),
    };
    ctx.can_create_branch(&branch)?;
    ctx.can_delete_branch(&branch)?;
    ctx.txn.rebase = None;
    flush(ctx)?;
    new_branch(ctx.db, &branch, ctx.txn.head)?;
    switch_branch(ctx, &branch)?;
    let restored = move_untracked_tables(ctx, &pre_working, &staged_before, ctx.txn.root.clone())?;
    ctx.txn.root = restored;
    flush(ctx)?;
    delete_branch(ctx.db, &working_branch)?;
    Ok(outcome(0, format!("Successfully rebased and updated refs/heads/{branch}")))
}

/// move_untracked_tables copies the tables of a working root that its staged root lacks into a target root, leaving
/// out those whose names or column tags the target already has, as Dolt's MoveUntrackedTables does.
fn move_untracked_tables(ctx: &mut Ctx<'_>, working: &Root, staged: &Root, mut target: Root) -> Result<Root> {
    let staged_tables = table_map(ctx.db, staged)?;
    let target_tables = table_map(ctx.db, &target)?;
    let mut target_tags = std::collections::HashSet::new();
    for ((schema, name), address) in &target_tables {
        target_tags.extend(TableDef::load(ctx.db, schema, name, *address)?.columns.iter().map(|c| c.tag));
    }
    for ((schema, name), address) in table_map(ctx.db, working)? {
        if staged_tables.contains_key(&(schema.clone(), name.clone()))
            || target_tables.contains_key(&(schema.clone(), name.clone()))
        {
            continue;
        }
        let table = TableDef::load(ctx.db, &schema, &name, address)?;
        if table.columns.iter().any(|c| target_tags.contains(&c.tag)) {
            continue;
        }
        target.put_table(ctx.db, &schema, &name, Some(address))?;
    }
    Ok(target)
}
