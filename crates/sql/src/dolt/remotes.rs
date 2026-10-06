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

//! Dolt's remotes and backups: their configuration in a database's repository state file, and pushing to, fetching
//! from, pulling from, and cloning file remotes, and syncing and restoring backups.

use std::collections::BTreeMap;
use std::path::{Path, PathBuf};

use doltdb::create::{branch_ref, working_set_ref};
use doltdb::database::Database;
use serde_json::{Map, Value as Json};
use serial::write::{WorkingSetFields, write_working_set};
use store::Hash;

use crate::dolt::args::{Kind, Parser, error};
use crate::dolt::history;
use crate::dolt::procedures::strings;
use crate::error::{PgError, Result};
use crate::query::Ctx;
use crate::types::Value;

/// REPO_STATE is the repository state file in a database's `.dolt` directory.
const REPO_STATE: &str = ".dolt/repo_state.json";

/// Remote is a remote or backup as the repository state file records it.
#[derive(Clone, Debug, PartialEq)]
pub struct Remote {
    pub name: String,
    pub url: String,
    pub fetch_specs: Vec<String>,
    pub params: BTreeMap<String, String>,
}

impl Remote {
    /// new returns a remote whose fetch spec maps every branch to a tracking branch, as Dolt's NewRemote does.
    fn new(name: &str, url: &str) -> Remote {
        Remote {
            name: name.to_string(),
            url: url.to_string(),
            fetch_specs: vec![format!("refs/heads/*:refs/remotes/{name}/*")],
            params: BTreeMap::new(),
        }
    }

    /// json returns the remote as the repository state file writes it.
    fn json(&self) -> Json {
        let mut object = Map::new();
        object.insert("name".into(), Json::String(self.name.clone()));
        object.insert("url".into(), Json::String(self.url.clone()));
        object.insert("fetch_specs".into(), self.fetch_specs.iter().cloned().map(Json::String).collect());
        object.insert(
            "params".into(),
            Json::Object(self.params.iter().map(|(k, v)| (k.clone(), Json::String(v.clone()))).collect()),
        );
        Json::Object(object)
    }

    /// from_json reads a remote as the repository state file writes it.
    fn from_json(value: &Json) -> Remote {
        let text = |key: &str| value.get(key).and_then(Json::as_str).unwrap_or_default().to_string();
        Remote {
            name: text("name"),
            url: text("url"),
            fetch_specs: value
                .get("fetch_specs")
                .and_then(Json::as_array)
                .map(|specs| specs.iter().filter_map(|s| s.as_str().map(str::to_string)).collect())
                .unwrap_or_default(),
            params: value
                .get("params")
                .and_then(Json::as_object)
                .map(|p| p.iter().map(|(k, v)| (k.clone(), v.as_str().unwrap_or_default().to_string())).collect())
                .unwrap_or_default(),
        }
    }

    /// tracking_ref returns the remote-tracking ref that the remote's fetch specs map a branch to, if any does.
    fn tracking_ref(&self, branch: &str) -> Option<Ref> {
        self.fetch_specs.iter().find_map(|spec| match RefSpec::parse(&self.name, spec) {
            Ok(spec) => spec.dest(&Ref::Branch(branch.to_string())),
            Err(_) => None,
        })
    }
}

/// Upstream is the remote branch a local branch tracks: the branch's ref on the remote and the remote's name.
#[derive(Clone, Debug, PartialEq)]
pub struct Upstream {
    pub merge: String,
    pub remote: String,
}

/// RepoState is a database's repository state file: its checked out branch, remotes, backups, and upstreams.
#[derive(Clone, Debug, Default, PartialEq)]
pub struct RepoState {
    pub head: String,
    pub remotes: BTreeMap<String, Remote>,
    pub backups: BTreeMap<String, Remote>,
    pub branches: BTreeMap<String, Upstream>,
}

impl RepoState {
    /// load reads the repository state file of the database in a directory.
    pub fn load(dir: &Path) -> Result<RepoState> {
        let text = std::fs::read_to_string(dir.join(REPO_STATE)).map_err(PgError::internal)?;
        let value: Json = serde_json::from_str(&text).map_err(PgError::internal)?;
        let remotes = |key: &str| -> BTreeMap<String, Remote> {
            value
                .get(key)
                .and_then(Json::as_object)
                .map(|m| m.iter().map(|(k, v)| (k.clone(), Remote::from_json(v))).collect())
                .unwrap_or_default()
        };
        let branches = value
            .get("branches")
            .and_then(Json::as_object)
            .map(|m| {
                m.iter()
                    .map(|(k, v)| {
                        let text = |key: &str| v.get(key).and_then(Json::as_str).unwrap_or_default().to_string();
                        (k.clone(), Upstream { merge: text("head"), remote: text("remote") })
                    })
                    .collect()
            })
            .unwrap_or_default();
        Ok(RepoState {
            head: value.get("head").and_then(Json::as_str).unwrap_or("refs/heads/main").to_string(),
            remotes: remotes("remotes"),
            backups: remotes("backups"),
            branches,
        })
    }

    /// save writes the repository state file of the database in a directory.
    pub fn save(&self, dir: &Path) -> Result<()> {
        let remotes =
            |m: &BTreeMap<String, Remote>| Json::Object(m.iter().map(|(k, r)| (k.clone(), r.json())).collect());
        let mut object = Map::new();
        object.insert("head".into(), Json::String(self.head.clone()));
        object.insert("remotes".into(), remotes(&self.remotes));
        object.insert("backups".into(), remotes(&self.backups));
        object.insert(
            "branches".into(),
            Json::Object(
                self.branches
                    .iter()
                    .map(|(k, u)| {
                        let mut upstream = Map::new();
                        upstream.insert("head".into(), Json::String(u.merge.clone()));
                        upstream.insert("remote".into(), Json::String(u.remote.clone()));
                        (k.clone(), Json::Object(upstream))
                    })
                    .collect(),
            ),
        );
        let text = serde_json::to_string_pretty(&Json::Object(object)).map_err(PgError::internal)?;
        std::fs::write(dir.join(REPO_STATE), text).map_err(PgError::internal)
    }

    /// default_remote returns the remote that commands use without one named: the only remote, or else `origin`, as
    /// Dolt's GetDefaultRemote picks it.
    fn default_remote(&self) -> Result<Remote> {
        match self.remotes.len() {
            0 => Err(error("no remote")),
            1 => Ok(self.remotes.values().next().cloned().unwrap_or_else(|| Remote::new("", ""))),
            _ => self.remotes.get("origin").cloned().ok_or_else(|| error("unable to determine the default remote")),
        }
    }
}

/// database_dir returns the directory of the session's database.
pub fn database_dir(ctx: &Ctx<'_>) -> PathBuf {
    ctx.session.data_dir.join(&ctx.txn.database)
}

/// go_os_error returns an operating system error's text as Go writes it, in lower case without the error number.
fn go_os_error(err: &std::io::Error) -> String {
    let text = err.to_string();
    let text = text.split(" (os error").next().unwrap_or_default();
    let mut chars = text.chars();
    chars.next().map_or_else(String::new, |c| c.to_lowercase().chain(chars).collect())
}

/// make_dirs creates a directory and its missing parents, failing as Go's MkdirAll does with the first directory it
/// could not create.
fn make_dirs(path: &Path) -> std::result::Result<(), String> {
    let mut missing: Vec<&Path> = path.ancestors().take_while(|p| !p.exists()).collect();
    missing.reverse();
    for dir in missing {
        if let Err(err) = std::fs::create_dir(dir)
            && !dir.is_dir()
        {
            return Err(format!("mkdir {}: {}", dir.display(), go_os_error(&err)));
        }
    }
    Ok(())
}

/// clean_path cleans a path lexically, as Go's filepath.Clean does.
fn clean_path(path: &Path) -> PathBuf {
    let mut out = PathBuf::new();
    for component in path.components() {
        match component {
            std::path::Component::ParentDir => {
                if !out.pop() {
                    out.push("..");
                }
            }
            std::path::Component::CurDir => {}
            other => out.push(other),
        }
    }
    out
}

/// absolute_url returns a remote's URL with a file path made absolute against a directory, creating the directory
/// it names when missing, as Dolt's GetAbsRemoteUrl does. URLs without a scheme name DoltHub databases.
fn absolute_url(base: &Path, url: &str) -> Result<String> {
    let Some((scheme, rest)) = url.split_once("://") else {
        return Ok(format!("https://doltremoteapi.dolthub.com/{}", url.trim_start_matches('/')));
    };
    if scheme != "file" {
        return Ok(url.to_string());
    }
    let path = clean_path(&base.join(rest));
    if !path.exists() {
        make_dirs(&path).map_err(|e| error(format!("failed to create directory '{}': {e}", path.display())))?;
    } else if !path.is_dir() {
        return Err(error("path is a file"));
    }
    Ok(format!("file://{}", path.display()))
}

/// file_path returns the directory that a file remote's URL names.
fn file_path(url: &str) -> Option<PathBuf> {
    url.strip_prefix("file://").map(PathBuf::from)
}

/// open_remote opens the database at a remote's URL, which must be a file remote.
fn open_remote(remote: &Remote) -> Result<Database> {
    let Some(path) = file_path(&remote.url) else {
        return Err(PgError::unsupported(format!("remotes at {}", remote.url)));
    };
    if !path.is_dir() {
        return Err(error(format!(
            "failed to get remote db; the remote: {} '{}' could not be accessed; stat {}: no such file or directory",
            remote.name,
            remote.url,
            path.display()
        )));
    }
    Ok(Database::open_remote(&path)?)
}

/// Ref is a Dolt ref: a branch, a remote-tracking branch, or a tag.
#[derive(Clone, Debug, PartialEq, Eq)]
enum Ref {
    Branch(String),
    Remote(String, String),
    Tag(String),
}

impl Ref {
    /// parse reads a ref, taking a name without a `refs/` prefix as a branch, as Dolt's ref.Parse does.
    fn parse(text: &str) -> Option<Ref> {
        if let Some(path) = text.strip_prefix("refs/heads/") {
            Some(Ref::Branch(path.to_string()))
        } else if let Some(path) = text.strip_prefix("refs/remotes/").or_else(|| text.strip_prefix("remotes/")) {
            let (remote, branch) = path.split_once('/')?;
            Some(Ref::Remote(remote.to_string(), branch.to_string()))
        } else if let Some(path) = text.strip_prefix("refs/tags/") {
            Some(Ref::Tag(path.to_string()))
        } else if text.starts_with("refs/") {
            None
        } else {
            Some(Ref::Branch(text.to_string()))
        }
    }

    /// dataset returns the ref's dataset name.
    fn dataset(&self) -> String {
        match self {
            Ref::Branch(path) => branch_ref(path),
            Ref::Remote(remote, branch) => format!("refs/remotes/{remote}/{branch}"),
            Ref::Tag(path) => format!("refs/tags/{path}"),
        }
    }

    /// path returns the ref's name without its kind.
    fn path(&self) -> String {
        match self {
            Ref::Branch(path) | Ref::Tag(path) => path.clone(),
            Ref::Remote(remote, branch) => format!("{remote}/{branch}"),
        }
    }
}

/// Pattern matches a branch name exactly, or around one `*` that captures the rest.
#[derive(Clone, Debug, PartialEq)]
enum Pattern {
    Exact(String),
    Wildcard(String, String),
}

impl Pattern {
    /// new returns the pattern of a name with at most one `*`.
    fn new(text: &str) -> Pattern {
        match text.split_once('*') {
            Some((prefix, suffix)) => Pattern::Wildcard(prefix.to_string(), suffix.to_string()),
            None => Pattern::Exact(text.to_string()),
        }
    }

    /// matches returns what a name's `*` captured, or an empty string for an exact match, when the name matches.
    fn matches(&self, name: &str) -> Option<String> {
        match self {
            Pattern::Exact(exact) => (exact == name).then(String::new),
            Pattern::Wildcard(prefix, suffix) => {
                name.strip_prefix(prefix.as_str())?.strip_suffix(suffix.as_str()).map(str::to_string)
            }
        }
    }

    /// fill returns the name with a capture in place of its `*`.
    fn fill(&self, captured: &str) -> String {
        match self {
            Pattern::Exact(exact) => exact.clone(),
            Pattern::Wildcard(prefix, suffix) => format!("{prefix}{captured}{suffix}"),
        }
    }

    /// shown returns the pattern as Go prints its branch mapper.
    fn shown(&self) -> String {
        match self {
            Pattern::Exact(exact) => exact.clone(),
            Pattern::Wildcard(prefix, suffix) => format!("{{{prefix} {suffix}}}"),
        }
    }
}

/// RefSpec maps refs between two databases, as Dolt's ref specs do.
#[derive(Clone, Debug, PartialEq)]
enum RefSpec {
    /// A branch to a branch, where an empty source deletes the destination.
    BranchToBranch(String, String),
    TagToTag(String, String),
    /// Branches matching a pattern to remote-tracking branches of a remote.
    Tracking {
        remote: String,
        local: Pattern,
        tracked: Pattern,
    },
}

impl RefSpec {
    /// parse reads a ref spec, which a remote's name constrains when not empty, as Dolt's ParseRefSpecForRemote
    /// does.
    fn parse(remote: &str, text: &str) -> Result<RefSpec> {
        let invalid = || error("invalid ref spec");
        let (from, to) = if let Some(rest) = text.strip_prefix(':') {
            (Ref::Branch(String::new()), Ref::parse(rest).ok_or_else(invalid)?)
        } else {
            let tokens: Vec<&str> = text.split(':').collect();
            if text.is_empty() || tokens.len() > 2 {
                return Err(invalid());
            }
            let to = tokens.get(1).unwrap_or(&tokens[0]);
            (Ref::parse(tokens[0]).ok_or_else(invalid)?, Ref::parse(to).ok_or_else(invalid)?)
        };
        match (from, to) {
            (Ref::Branch(src), Ref::Remote(dest_remote, dest)) => {
                let (src_wildcards, dest_wildcards) = (src.matches('*').count(), dest.matches('*').count());
                if src_wildcards != dest_wildcards || src_wildcards > 1 || (!remote.is_empty() && remote != dest_remote)
                {
                    return Err(invalid());
                }
                Ok(RefSpec::Tracking { remote: dest_remote, local: Pattern::new(&src), tracked: Pattern::new(&dest) })
            }
            (Ref::Branch(src), Ref::Branch(dest)) => Ok(RefSpec::BranchToBranch(src, dest)),
            (Ref::Tag(src), Ref::Tag(dest)) => Ok(RefSpec::TagToTag(src, dest)),
            _ => Err(error("unsupported mapping")),
        }
    }

    /// dest returns the ref that the spec maps a ref to, if it maps it.
    fn dest(&self, from: &Ref) -> Option<Ref> {
        match (self, from) {
            (RefSpec::BranchToBranch(src, dest), Ref::Branch(b)) if src == b => Some(Ref::Branch(dest.clone())),
            (RefSpec::TagToTag(src, dest), Ref::Tag(t)) if src == t => Some(Ref::Tag(dest.clone())),
            (RefSpec::Tracking { remote, local, tracked }, Ref::Branch(b)) => {
                local.matches(b).map(|captured| Ref::Remote(remote.clone(), tracked.fill(&captured)))
            }
            _ => None,
        }
    }

    /// local_name returns what the spec names on the remote side, as Go prints its GetRemRefToLocal mapper.
    fn local_name(&self) -> String {
        match self {
            RefSpec::BranchToBranch(src, _) | RefSpec::TagToTag(src, _) => src.clone(),
            RefSpec::Tracking { local, .. } => local.shown(),
        }
    }
}

/// fetch_specs returns the ref specs that fetch arguments name, each branch name mapped to its remote-tracking
/// branch, as Dolt's ParseRSFromArgs does.
fn fetch_specs(remote: &Remote, args: &[String]) -> Result<Vec<RefSpec>> {
    let mut specs = Vec::new();
    for arg in args {
        let invalid = || error(format!("invalid fetch spec: '{arg}'"));
        let mut spec = RefSpec::parse("", arg).map_err(|_| invalid())?;
        if matches!(spec, RefSpec::BranchToBranch(..))
            && let Ok(tracking) = RefSpec::parse("", &format!("refs/heads/{arg}:remotes/{}/{arg}", remote.name))
        {
            spec = tracking;
        }
        if matches!(spec, RefSpec::BranchToBranch(..)) {
            return Err(invalid());
        }
        specs.push(spec);
    }
    Ok(specs)
}

/// INVALID_NAME_CHARACTERS are the characters that remote and backup names cannot hold.
const INVALID_NAME_CHARACTERS: &[char] = &[
    ' ', '\t', '\n', '\r', '.', '/', '\\', '!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '{', '}', '[', ']', ',',
    '<', '>', '\'', '"', '?', '=', '+', '|',
];

/// REMOTE parses dolt_remote's arguments.
const REMOTE: Parser =
    Parser { command: "remote", options: &[("verbose", "v", Kind::Flag), ("ref", "", Kind::Value)], max_args: None };

/// dolt_remote adds or removes a remote.
pub fn dolt_remote(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    crate::dolt::procedures::require_admin(ctx)?;
    let parsed = REMOTE.parse(&strings(args))?;
    let dir = database_dir(ctx);
    let mut state = RepoState::load(&dir)?;
    match parsed.args.first().map(String::as_str) {
        None => return Err(error("error: invalid argument, use 'dolt_remotes' system table to list remotes")),
        Some("add") => {
            if parsed.args.len() != 3 {
                return Err(error("error: invalid argument"));
            }
            let name = parsed.args[1].trim().to_string();
            let url = absolute_url(&dir, &parsed.args[2])?;
            if state.remotes.contains_key(&name) {
                return Err(error("remote already exists"));
            }
            if name.contains(INVALID_NAME_CHARACTERS) {
                return Err(error("remote name invalid"));
            }
            if let Some(backup) = state.backups.values().find(|b| b.url == url) {
                return Err(error(format!("address conflict with a remote: '{}' -> {}", backup.name, backup.url)));
            }
            state.remotes.insert(name.clone(), Remote::new(&name, &url));
        }
        Some("remove" | "rm") => {
            if parsed.args.len() != 2 {
                return Err(error("error: invalid argument"));
            }
            let name = parsed.args[1].trim().to_string();
            if !state.remotes.contains_key(&name) {
                return Err(error(format!("error: unknown remote: '{name}'")));
            }
            let prefix = format!("refs/remotes/{name}/");
            let doomed: Vec<(String, Option<Hash>)> = ctx
                .db
                .datasets()?
                .into_iter()
                .filter(|(n, _)| n.starts_with(&prefix))
                .map(|(n, _)| (n, None))
                .collect();
            ctx.db.set_heads(&doomed)?;
            for upstream in state.branches.values_mut().filter(|u| u.remote == name) {
                upstream.remote.clear();
            }
            state.remotes.remove(&name);
        }
        Some(_) => return Err(error("error: invalid argument")),
    }
    state.save(&dir)?;
    Ok(Value::Int8(0))
}

/// PushTarget is a ref to push: the local ref, the ref it updates on the remote, the remote-tracking ref that follows
/// it, and whether to set it as the branch's upstream.
struct PushTarget {
    src: Ref,
    dest: Ref,
    tracking: Option<Ref>,
    set_upstream: bool,
}

/// PUSH parses dolt_push's arguments.
const PUSH: Parser = Parser {
    command: "push",
    options: &[
        ("user", "", Kind::Value),
        ("set-upstream", "u", Kind::Flag),
        ("force", "f", Kind::Flag),
        ("all", "", Kind::Flag),
        ("silent", "", Kind::Flag),
    ],
    max_args: None,
};

/// remote_not_found returns the error for pushing to or pulling from a remote that does not exist.
fn remote_not_found(name: &str) -> PgError {
    error(format!("fatal: remote '{name}' not found.\nPlease make sure the remote exists."))
}

/// no_upstream returns the error for pushing a branch without an upstream.
fn no_upstream(branch: &str, remote: &str) -> PgError {
    error(format!(
        "fatal: The current branch {branch} has no upstream branch.\nTo push the current branch and set the remote as \
         upstream, use\n\tdolt push --set-upstream {remote} {branch}\nTo have this happen automatically for branches \
         without a tracking\nupstream, see 'push.autoSetupRemote' in 'dolt config --help'."
    ))
}

/// push_spec returns the ref spec that a push argument names, reading a local branch or tag name as its full ref,
/// as Dolt's getRefSpecFromStr does.
fn push_spec(ctx: &mut Ctx<'_>, text: &str) -> Result<RefSpec> {
    let datasets = ctx.db.datasets()?;
    let full = [format!("refs/heads/{text}"), format!("refs/tags/{text}")]
        .into_iter()
        .find(|name| datasets.iter().any(|(n, _)| n == name))
        .unwrap_or_else(|| text.to_string());
    RefSpec::parse("", &full).map_err(|e| error(format!("{}: '{full}'", e.message)))
}

/// push_target returns what pushing a spec does, as Dolt's getPushTargetFromRefSpec does.
fn push_target(ctx: &Ctx<'_>, spec: &RefSpec, remote: &Remote, set_upstream: bool) -> Result<PushTarget> {
    let current = Ref::Branch(ctx.txn.branch.clone());
    let src = match spec {
        RefSpec::BranchToBranch(src, _) => Ref::Branch(src.clone()),
        RefSpec::TagToTag(src, _) => Ref::Tag(src.clone()),
        RefSpec::Tracking { local, .. } => {
            local.matches(&ctx.txn.branch).map(|_| current).ok_or_else(|| error("invalid ref spec"))?
        }
    };
    let dest = spec.dest(&src).ok_or_else(|| error("invalid ref spec"))?;
    let tracking = match &dest {
        Ref::Branch(path) => {
            if !history::valid_branch_name(path) && !path.is_empty() {
                return Err(error(format!("not a valid user branch name: '{path}'")));
            }
            remote.tracking_ref(path)
        }
        Ref::Tag(_) if set_upstream => return Err(error("cannot set upstream for tag")),
        Ref::Tag(_) => None,
        Ref::Remote(..) => {
            return Err(error(format!("cannot push ref: '{}' of type 'remotes'", src.dataset())));
        }
    };
    Ok(PushTarget { src, dest, tracking, set_upstream })
}

/// push_targets returns the refs that dolt_push's arguments push and the remote they go to, as Dolt's NewPushOpts
/// works them out.
fn push_targets(
    ctx: &mut Ctx<'_>,
    state: &RepoState,
    parsed: &crate::dolt::args::Parsed,
) -> Result<(Vec<PushTarget>, Remote)> {
    let set_upstream = parsed.has("set-upstream");
    let branch = ctx.txn.branch.clone();
    let Some(name) = parsed.args.first() else {
        let remote = state.default_remote().map_err(|_| {
            error(
                "fatal: No configured push destination.\nEither specify the URL from the command-line or configure a \
                 remote repository using\n\n\tdolt remote add <name> <url>\n\nand then push using the remote \
                 name\n\n\tdolt push <name>\n\n",
            )
        })?;
        let upstream = state.branches.get(&branch).filter(|u| !u.remote.is_empty());
        let Some(upstream) = upstream.filter(|_| !set_upstream) else {
            return Err(no_upstream(&branch, &remote.name));
        };
        let remote = state.remotes.get(&upstream.remote).cloned().ok_or_else(|| remote_not_found(&upstream.remote))?;
        let merge = upstream.merge.strip_prefix("refs/heads/").unwrap_or(&upstream.merge).to_string();
        if merge != branch {
            return Err(error(
                "the upstream branch of your current branch does not match the name of your current branch",
            ));
        }
        let target = push_target(ctx, &RefSpec::BranchToBranch(branch.clone(), merge), &remote, false)?;
        return Ok((vec![target], remote));
    };
    let remote = state.remotes.get(name).cloned().ok_or_else(|| remote_not_found(name))?;
    let mut names = parsed.args[1..].to_vec();
    if names.is_empty() {
        if parsed.has("all") {
            names = history::refs(ctx.db, "refs/heads/")?.into_iter().map(|(n, _)| n).collect();
        } else {
            if state.default_remote().is_ok_and(|d| d.name == remote.name) {
                return Err(no_upstream(&branch, &remote.name));
            }
            let spec = push_spec(ctx, &branch)?;
            let target = push_target(ctx, &spec, &remote, true)?;
            return Ok((vec![target], remote));
        }
    } else if parsed.has("all") {
        return Err(error("fatal: --all can't be combined with refspecs"));
    }
    let mut targets = Vec::new();
    for name in &names {
        if name.is_empty() {
            return Err(error(format!("invalid ref spec: '{name}'")));
        }
        let spec = push_spec(ctx, name)?;
        targets.push(push_target(ctx, &spec, &remote, set_upstream)?);
    }
    Ok((targets, remote))
}

/// PushOutcome is how pushing one ref went: a new branch or tag, an update between two commits, a forced update,
/// a deleted branch, a branch already up to date, or a rejected non-fast-forward update.
enum PushOutcome {
    NewBranch,
    NewTag,
    Updated(Hash, Hash),
    Forced(Hash, Hash),
    Deleted,
    UpToDate,
    Rejected,
}

/// working_set_for writes a working set whose roots are a commit's root, as Dolt's SetHead does with a working set
/// path, and returns its address.
fn working_set_for(db: &mut Database, commit: Hash) -> Result<Hash> {
    let root = history::load(db, commit)?.root;
    let fields = WorkingSetFields {
        working_root: root,
        staged_root: Some(root),
        merge_state: None,
        rebase_state: None,
        meta: None,
    };
    Ok(db.write_value(write_working_set(&fields))?)
}

/// push_ref pushes one ref to the remote's database, as Dolt's push does.
fn push_ref(ctx: &mut Ctx<'_>, remote_db: &mut Database, target: &PushTarget, force: bool) -> Result<PushOutcome> {
    let dest = target.dest.dataset();
    if let Ref::Branch(src) = &target.src
        && src.is_empty()
    {
        if remote_db.head(&dest)?.is_none() {
            return Err(error(format!("failed to delete remote; '{dest}' from remote; branch not found")));
        }
        let mut doomed = vec![(dest.clone(), None)];
        if let Ref::Branch(path) = &target.dest
            && remote_db.head(&working_set_ref(path))?.is_some()
        {
            doomed.push((working_set_ref(path), None));
        }
        remote_db.set_heads(&doomed)?;
        if let Some(tracking) = &target.tracking {
            ctx.db.set_heads(&[(tracking.dataset(), None)])?;
        }
        return Ok(PushOutcome::Deleted);
    }
    if let Ref::Tag(_) = &target.src {
        let address = ctx.db.head(&target.src.dataset())?.ok_or_else(|| error("tag not found"))?;
        remote_db.pull(ctx.db, address)?;
        remote_db.set_heads(&[(dest, Some(address))])?;
        return Ok(PushOutcome::NewTag);
    }
    let src = target.src.path();
    let commit = history::resolve(ctx.db, ctx.txn.head, &src)
        .map_err(|e| error(format!("invalid ref spec; refspec not found: '{src}'; {}", e.message)))?;
    let previous = remote_db.head(&dest)?;
    let mut fast_forward = true;
    if let Some(previous) = previous {
        if previous == commit {
            return Ok(PushOutcome::UpToDate);
        }
        remote_db.pull(ctx.db, commit)?;
        fast_forward = history::is_ancestor(remote_db, previous, commit)?;
        if !force && !fast_forward {
            return Ok(PushOutcome::Rejected);
        }
    } else {
        remote_db.pull(ctx.db, commit)?;
    }
    let mut heads = vec![(dest.clone(), Some(commit))];
    if force && let Ref::Branch(path) = &target.dest {
        heads.push((working_set_ref(path), Some(working_set_for(remote_db, commit)?)));
    }
    remote_db.set_heads(&heads)?;
    if let Some(tracking) = &target.tracking {
        ctx.db.set_heads(&[(tracking.dataset(), Some(commit))])?;
    }
    Ok(match previous {
        None => PushOutcome::NewBranch,
        Some(previous) if force || !fast_forward => PushOutcome::Forced(previous, commit),
        Some(previous) => PushOutcome::Updated(previous, commit),
    })
}

/// push_record returns dolt_push's result.
fn push_record(status: i64, message: &str) -> Value {
    Value::Record(vec![Value::Int8(status), Value::Text(message.to_string())])
}

/// dolt_push pushes branches and tags to a remote.
pub fn dolt_push(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    crate::dolt::procedures::require_admin(ctx)?;
    let parsed = PUSH.parse(&strings(args))?;
    let dir = database_dir(ctx);
    let mut state = RepoState::load(&dir)?;
    let (targets, remote) = push_targets(ctx, &state, &parsed)?;
    let mut remote_db = open_remote(&remote)?;
    let (mut pushed, mut upstreams, mut rejected) = (Vec::new(), Vec::new(), Vec::new());
    let mut up_to_date = false;
    for target in &targets {
        let (src, dest) = (target.src.path(), target.dest.path());
        match push_ref(ctx, &mut remote_db, target, parsed.has("force"))? {
            PushOutcome::Rejected => {
                rejected.push(format!(" ! [rejected]            {src} -> {dest} (non-fast-forward)"));
                continue;
            }
            PushOutcome::UpToDate => up_to_date = true,
            PushOutcome::NewBranch => pushed.push(format!(" * [new branch]          {src} -> {dest}")),
            PushOutcome::NewTag => pushed.push(format!(" * [new tag]             {src} -> {dest}")),
            PushOutcome::Deleted => pushed.push(format!(" - [deleted]             {dest}")),
            PushOutcome::Forced(old, new) => pushed.push(format!(" + {old}...{new} {src} -> {dest} (forced update)")),
            PushOutcome::Updated(old, new) => pushed.push(format!("   {old}..{new}  {src} -> {dest}")),
        }
        if target.set_upstream {
            state.branches.insert(src.clone(), Upstream { merge: target.dest.dataset(), remote: remote.name.clone() });
            let tracked = target.tracking.as_ref().map(Ref::path).unwrap_or_default();
            upstreams.push(format!("branch '{src}' set up to track '{tracked}'."));
        }
    }
    remote_db.close()?;
    state.save(&dir)?;
    if pushed.is_empty() && rejected.is_empty() {
        return Ok(push_record(0, if up_to_date { "Everything up-to-date" } else { "" }));
    }
    let mut message = format!("To {}", remote.url);
    for line in pushed.iter().chain(&rejected).chain(&upstreams) {
        message.push('\n');
        message.push_str(line);
    }
    if !rejected.is_empty() {
        return Err(error(format!(
            "{message}\nerror: failed to push some refs to '{}'\nhint: Updates were rejected because the tip of your \
             current branch is behind\nhint: its remote counterpart. Integrate the remote changes (e.g.\nhint: 'dolt \
             pull ...') before pushing again.\n",
            remote.url
        )));
    }
    Ok(push_record(0, &message))
}

/// FETCH parses dolt_fetch's arguments.
const FETCH: Parser = Parser {
    command: "fetch",
    options: &[("user", "", Kind::Value), ("prune", "p", Kind::Flag), ("silent", "", Kind::Flag)],
    max_args: None,
};

/// remote_specs returns the ref specs of a remote's fetch specs, as Dolt's GetRefSpecs does.
fn remote_specs(remote: &Remote) -> Result<Vec<RefSpec>> {
    remote
        .fetch_specs
        .iter()
        .map(|spec| {
            RefSpec::parse(&remote.name, spec)
                .map_err(|_| error(format!("error: for '{}', '{spec}' is not a valid refspec.", remote.name)))
        })
        .collect()
}

/// fetch copies the remote's branches that the specs map to remote-tracking branches, then the tags whose commits it
/// now has, pruning remote-tracking branches the remote no longer has when asked, as Dolt's FetchRefSpecs does.
fn fetch(
    ctx: &mut Ctx<'_>,
    remote_db: &mut Database,
    remote: &Remote,
    specs: &[RefSpec],
    defaults: bool,
    prune: bool,
) -> Result<()> {
    let branches = history::refs(remote_db, "refs/heads/")?;
    if branches.is_empty() {
        if defaults {
            return Ok(());
        }
        return Err(error(format!("no branches found in remote '{}'", remote.name)));
    }
    let mut heads: Vec<(Ref, Hash)> = Vec::new();
    for spec in specs {
        let mut seen = false;
        for (branch, commit) in &branches {
            if let Some(tracking) = spec.dest(&Ref::Branch(branch.clone())) {
                seen = true;
                heads.push((tracking, *commit));
            }
        }
        if !seen {
            return Err(error(format!("invalid ref spec: '{}'", spec.local_name())));
        }
    }
    for (_, commit) in &heads {
        ctx.db.pull(remote_db, *commit)?;
    }
    let mut updates: Vec<(String, Option<Hash>)> = heads.iter().map(|(r, c)| (r.dataset(), Some(*c))).collect();
    if prune {
        let prefix = format!("refs/remotes/{}/", remote.name);
        for (name, _) in ctx.db.datasets()? {
            if name.starts_with(&prefix) && !heads.iter().any(|(r, _)| r.dataset() == name) {
                updates.push((name, None));
            }
        }
    }
    for (name, address) in history::refs(remote_db, "refs/tags/")? {
        if ctx.db.has(&address) {
            continue;
        }
        let commit = history::commit_of(remote_db, address)?;
        if ctx.db.has(&commit) {
            ctx.db.pull(remote_db, address)?;
            updates.push((format!("refs/tags/{name}"), Some(address)));
        }
    }
    ctx.db.set_heads(&updates)?;
    Ok(())
}

/// fetch_remote returns the remote that fetch arguments name, `origin` without one, and the remaining arguments, as
/// Dolt's RemoteForFetchArgs does.
fn fetch_remote<'a>(state: &RepoState, args: &'a [String]) -> Result<(Remote, &'a [String])> {
    if state.remotes.is_empty() {
        return Err(error("no remote"));
    }
    let (name, rest) = match args.split_first() {
        Some((name, rest)) => (name.as_str(), rest),
        None => ("origin", args),
    };
    let remote = state.remotes.get(name).cloned().ok_or_else(|| {
        error(format!(
            "unknown remote; '{name}' does not appear to be a dolt database. could not read from the remote database. \
             please make sure you have the correct access rights and the database exists"
        ))
    })?;
    Ok((remote, rest))
}

/// dolt_fetch fetches branches and tags from a remote.
pub fn dolt_fetch(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    crate::dolt::procedures::require_admin(ctx)?;
    let parsed = FETCH.parse(&strings(args))?;
    let state = RepoState::load(&database_dir(ctx))?;
    let (remote, spec_args) = fetch_remote(&state, &parsed.args)?;
    if !spec_args.is_empty() && parsed.has("prune") {
        return Err(error("--prune option cannot be provided with a ref spec"));
    }
    let (specs, defaults) =
        if spec_args.is_empty() { (remote_specs(&remote)?, true) } else { (fetch_specs(&remote, spec_args)?, false) };
    let mut remote_db = open_remote(&remote)?;
    fetch(ctx, &mut remote_db, &remote, &specs, defaults, parsed.has("prune"))
        .map_err(|e| error(format!("fetch failed: {}", e.message)))?;
    remote_db.close()?;
    Ok(Value::Int8(0))
}

/// PULL parses dolt_pull's arguments.
const PULL: Parser = Parser {
    command: "pull",
    options: &[
        ("squash", "", Kind::Flag),
        ("no-ff", "", Kind::Flag),
        ("ff-only", "", Kind::Flag),
        ("force", "f", Kind::Flag),
        ("commit", "", Kind::Flag),
        ("no-commit", "", Kind::Flag),
        ("no-edit", "", Kind::Flag),
        ("user", "", Kind::Value),
        ("prune", "p", Kind::Flag),
        ("silent", "", Kind::Flag),
        ("rebase", "r", Kind::Flag),
        ("skip-verification", "", Kind::Flag),
    ],
    max_args: Some(2),
};

/// pull_record returns dolt_pull's result.
fn pull_record(fast_forward: i64, conflicts: i64, message: &str) -> Value {
    let message = if message.is_empty() { Value::Null } else { Value::Text(message.to_string()) };
    Value::Record(vec![Value::Int8(fast_forward), Value::Int8(conflicts), message])
}

/// dolt_pull fetches the current branch's upstream from a remote and merges it, as Dolt's doDoltPull does.
pub fn dolt_pull(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    crate::dolt::procedures::require_admin(ctx)?;
    let parsed = PULL.parse(&strings(args)).map_err(|e| {
        if e.message.contains("too many positional arguments") { error("dolt pull takes at most two args") } else { e }
    })?;
    for (a, b) in [("ff-only", "no-ff"), ("ff-only", "squash")] {
        if parsed.has(a) && parsed.has(b) {
            return Err(error(format!("error: Flags '--{a}' and '--{b}' cannot be used together")));
        }
    }
    if parsed.has("rebase") {
        return Err(PgError::unsupported("dolt_pull --rebase"));
    }
    let state = RepoState::load(&database_dir(ctx))?;
    let remote_name = parsed.args.first().cloned().unwrap_or_default();
    let remote = if remote_name.is_empty() {
        state.default_remote().map_err(|_| error("no remote"))?
    } else {
        state.remotes.get(&remote_name).cloned().ok_or_else(|| remote_not_found(&remote_name))?
    };
    let specs = remote_specs(&remote)?;
    let branch = match parsed.args.get(1) {
        Some(name) => name.clone(),
        None => {
            let Some(upstream) = state.branches.get(&ctx.txn.branch) else {
                return Err(error(if parsed.args.len() == 1 {
                    format!(
                        "You asked to pull from the remote '{remote_name}', but did not specify a branch. Because \
                         this is not the default configured remote for your current branch, you must specify a \
                         branch."
                    )
                } else {
                    "There is no tracking information for the current branch.\nPlease specify which branch you want \
                     to merge with.\n\n\tdolt pull <remote> <branch>\n\nIf you wish to set tracking information for \
                     this branch you can do so with:\n\n\t dolt push --set-upstream <remote> <branch>\n"
                        .to_string()
                }));
            };
            upstream.merge.strip_prefix("refs/heads/").unwrap_or(&upstream.merge).to_string()
        }
    };
    let mut remote_db = open_remote(&remote)?;
    if remote_db.head(&branch_ref(&branch))?.is_none() {
        return Err(error(format!("branch \"{branch}\" not found on remote")));
    }
    let fetch_specs_used = match parsed.args.get(1) {
        Some(name) => fetch_specs(&remote, std::slice::from_ref(name))
            .map_err(|e| error(format!("invalid remote ref argument \"{name}\": {}", e.message)))?,
        None => specs.clone(),
    };
    fetch(ctx, &mut remote_db, &remote, &fetch_specs_used, false, parsed.has("prune"))
        .map_err(|e| error(format!("fetch failed: {}", e.message)))?;
    let mut outcome = pull_record(0, 0, "");
    for spec in &specs {
        let Some(tracking) = spec.dest(&Ref::Branch(branch.clone())) else {
            return Err(error(format!("invalid ref spec: '{}'", spec.local_name())));
        };
        let head = Hash::of(&ctx.txn.staged.encode());
        let (staged, working) = (ctx.txn.staged.clone(), ctx.txn.root.clone());
        if head != ctx.txn.head_root || !crate::dolt::revert::changed(ctx, &staged, &working)?.is_empty() {
            return Err(error("cannot merge with uncommitted changes"));
        }
        let message = format!("Merge branch '{branch}' of {} into {}", remote.url, ctx.txn.branch);
        let mut merge_args = vec![Value::Text("-m".into()), Value::Text(message)];
        for flag in ["squash", "no-ff", "ff-only", "no-commit", "force"] {
            if parsed.has(flag) {
                merge_args.push(Value::Text(format!("--{flag}")));
            }
        }
        merge_args.push(Value::Text(tracking.dataset()));
        let Value::Record(fields) = crate::dolt::procedures::dolt_merge(ctx, &merge_args)? else {
            return Err(PgError::internal("dolt_merge's result"));
        };
        let int = |v: &Value| v.output().and_then(|t| t.parse().ok()).unwrap_or(0);
        let conflicts = int(&fields[2]);
        let message = fields[3].output().unwrap_or_default();
        let message =
            if conflicts > 0 { "merge has unresolved conflicts or constraint violations".to_string() } else { message };
        outcome = pull_record(int(&fields[1]), conflicts, &message);
    }
    remote_db.close()?;
    Ok(outcome)
}

/// CLONE parses dolt_clone's arguments.
const CLONE: Parser = Parser {
    command: "clone",
    options: &[
        ("remote", "", Kind::Value),
        ("branch", "b", Kind::Value),
        ("depth", "", Kind::Value),
        ("ref", "", Kind::Value),
        ("user", "u", Kind::Value),
        ("single-branch", "", Kind::Flag),
    ],
    max_args: None,
};

/// default_branch returns the branch a clone checks out without one named: `main`, then `master`, then the first by
/// name, as Dolt's GetDefaultBranch picks it.
fn default_branch(branches: &[String]) -> String {
    for name in ["main", "master"] {
        if branches.iter().any(|b| b == name) {
            return name.to_string();
        }
    }
    branches.iter().min().cloned().unwrap_or_else(|| "main".to_string())
}

/// clone_into writes a new database in a directory from a remote's database, with remote-tracking branches for the
/// remote's branches, its tags, and a local branch checked out, as Dolt's CloneRemote does.
fn clone_into(remote_db: &mut Database, remote: &Remote, dir: &Path, branch: Option<&str>, single: bool) -> Result<()> {
    let datasets = remote_db.datasets()?;
    let branches: Vec<String> =
        datasets.iter().filter_map(|(n, _)| n.strip_prefix("refs/heads/").map(str::to_string)).collect();
    if datasets.is_empty() {
        return Err(error("clone failed; remote at that url contains no Dolt data"));
    }
    let branch = branch.map_or_else(|| default_branch(&branches), str::to_string);
    if datasets.iter().any(|(n, _)| n.strip_prefix("refs/tags/").is_some_and(|t| t.eq_ignore_ascii_case(&branch))) {
        return Err(error("this operation is not supported while in a detached head state"));
    }
    let commit = datasets
        .iter()
        .find(|(n, _)| *n == branch_ref(&branch))
        .map(|(_, h)| *h)
        .ok_or_else(|| error(format!("clone failed; branch not found: {branch}")))?;
    doltdb::create::create_files(dir, &branch)?;
    let mut db = Database::open(&dir.join(".dolt/noms"))?;
    let mut heads = Vec::new();
    for (name, address) in &datasets {
        if let Some(b) = name.strip_prefix("refs/heads/") {
            db.pull(remote_db, *address)?;
            if !single || b == branch {
                heads.push((Ref::Remote(remote.name.clone(), b.to_string()).dataset(), Some(*address)));
            }
            if b == branch {
                heads.push((name.clone(), Some(*address)));
            }
        } else if name.starts_with("refs/tags/") {
            db.pull(remote_db, *address)?;
            heads.push((name.clone(), Some(*address)));
        }
    }
    let root = history::load(&db, commit)?.root;
    let seconds = crate::dolt::procedures::now_millis() as u64 / 1000;
    let fields = WorkingSetFields {
        working_root: root,
        staged_root: Some(root),
        merge_state: None,
        rebase_state: None,
        meta: Some(doltdb::create::environment_meta(seconds)),
    };
    heads.push((working_set_ref(&branch), Some(db.write_value(write_working_set(&fields))?)));
    db.set_heads(&heads)?;
    db.close()?;
    let mut state = RepoState { head: branch_ref(&branch), ..RepoState::default() };
    state.remotes.insert(remote.name.clone(), remote.clone());
    state.branches.insert(branch.clone(), Upstream { merge: branch_ref(&branch), remote: remote.name.clone() });
    state.save(dir)
}

/// dolt_clone clones a remote into a new database.
pub fn dolt_clone(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    crate::dolt::procedures::require_admin(ctx)?;
    let parsed = CLONE.parse(&strings(args))?;
    if parsed.args.is_empty() || parsed.args.len() > 2 {
        return Err(error(
            "error: invalid number of arguments: database URL must be specified and database name is optional",
        ));
    }
    if parsed.has("depth") {
        return Err(PgError::unsupported("shallow clones"));
    }
    let url_arg = &parsed.args[0];
    let name = match parsed.args.get(1) {
        Some(name) => name.clone(),
        None => {
            let base = url_arg.trim_end_matches('/').rsplit('/').next().unwrap_or_default();
            let base = base.strip_suffix(".git").unwrap_or(base);
            if base.is_empty() {
                return Err(error("Could not infer repo name. Please explicitly define a directory for this url"));
            }
            base.to_string()
        }
    };
    let data_dir = ctx.session.data_dir.clone();
    if data_dir.join(&name).exists() {
        return Err(error(format!("can't create database {name}; database exists")));
    }
    let url = absolute_url(&data_dir, url_arg).map_err(|_| error(format!("error: '{url_arg}' is not valid.")))?;
    let remote = Remote::new(parsed.value("remote").unwrap_or("origin"), &url);
    let mut remote_db = open_remote(&remote)?;
    let dir = data_dir.join(&name);
    let cloned = clone_into(&mut remote_db, &remote, &dir, parsed.value("branch"), parsed.has("single-branch"));
    remote_db.close()?;
    if cloned.is_err() && dir.exists() {
        let _ = std::fs::remove_dir_all(&dir);
    }
    cloned?;
    Ok(Value::Int8(0))
}

/// BACKUP parses dolt_backup's arguments.
const BACKUP: Parser = Parser {
    command: "backup",
    options: &[
        ("verbose", "v", Kind::Flag),
        ("force", "f", Kind::Flag),
        ("ref", "", Kind::Value),
        ("aws-region", "", Kind::Value),
        ("aws-creds-type", "", Kind::Value),
        ("aws-creds-file", "", Kind::Value),
        ("aws-creds-profile", "", Kind::Value),
        ("prune-with-grace-period", "", Kind::Value),
    ],
    max_args: None,
};

/// AWS_USAGE are the optional AWS arguments that dolt_backup's usage errors list.
const AWS_USAGE: &[&str] =
    &["--aws-region=<region>", "--aws-creds-type=<type>", "--aws-creds-file=<file>", "--aws-creds-profile=<profile>"];

/// backup_usage returns dolt_backup's usage error for a subcommand, as Dolt's errDoltBackupUsage writes it.
fn backup_usage(command: &str, required: &[&str], optional: &[&str]) -> PgError {
    let mut text = format!("usage: dolt_backup('{command}'");
    for arg in required {
        text.push_str(&format!(", '{arg}'"));
    }
    for arg in optional {
        text.push_str(&format!(", ['{arg}']"));
    }
    text.push(')');
    error(text)
}

/// sync_to makes a backup's store root that of the session's database, copying the chunks it lacks, as Dolt's
/// SyncRoots does.
fn sync_to(ctx: &mut Ctx<'_>, backup: &Remote) -> Result<()> {
    if let Some(path) = file_path(&backup.url) {
        let _ = make_dirs(&path);
    }
    let mut dest = open_remote(backup)?;
    let root = ctx.db.root();
    if dest.root() != root {
        dest.pull(ctx.db, root)?;
        dest.replace_root(root)?;
    }
    Ok(dest.close()?)
}

/// dolt_backup adds, removes, syncs, and restores backups.
pub fn dolt_backup(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    crate::dolt::procedures::require_admin(ctx)?;
    let parsed = BACKUP.parse(&strings(args))?;
    if parsed.args.is_empty() || (parsed.args.len() == 1 && parsed.has("verbose")) {
        return Err(error("use 'dolt_backups' table to list backups"));
    }
    let command = parsed.args[0].clone();
    if parsed.has("prune-with-grace-period") && command != "sync" && command != "sync-url" {
        return Err(error("--prune-with-grace-period is only supported with 'sync' and 'sync-url'"));
    }
    let dir = database_dir(ctx);
    let data_dir = ctx.session.data_dir.clone();
    match command.as_str() {
        "add" => {
            if parsed.args.len() != 3 {
                return Err(backup_usage(&command, &["name", "url"], AWS_USAGE));
            }
            let name = parsed.args[1].clone();
            let url = absolute_url(&data_dir, &parsed.args[2])?;
            let mut state = RepoState::load(&dir)?;
            if state.backups.contains_key(&name) {
                return Err(error(format!("backup '{name}' already exists")));
            }
            if name.contains(INVALID_NAME_CHARACTERS) {
                return Err(error(format!("backup name '{name}' is invalid")));
            }
            if let Some(other) = state.remotes.values().chain(state.backups.values()).find(|r| r.url == url) {
                return Err(error(format!("address conflict with a remote: '{}' -> {}", other.name, other.url)));
            }
            state.backups.insert(name.clone(), Remote::new(&name, &url));
            state.save(&dir)?;
        }
        "remove" | "rm" => {
            if parsed.args.len() != 2 {
                return Err(backup_usage(&command, &["name"], &[]));
            }
            let mut state = RepoState::load(&dir)?;
            if state.backups.remove(&parsed.args[1]).is_none() {
                return Err(error(format!("backup '{}' not found", parsed.args[1])));
            }
            state.save(&dir)?;
        }
        "sync" => {
            if parsed.args.len() != 2 {
                return Err(backup_usage(&command, &["name"], &[]));
            }
            let state = RepoState::load(&dir)?;
            let backup = state
                .backups
                .get(&parsed.args[1])
                .cloned()
                .ok_or_else(|| error(format!("backup '{}' not found", parsed.args[1])))?;
            sync_to(ctx, &backup)?;
        }
        "sync-url" => {
            if parsed.args.len() != 2 {
                return Err(backup_usage(&command, &["remote_url"], AWS_USAGE));
            }
            let url = absolute_url(&data_dir, &parsed.args[1])?;
            sync_to(ctx, &Remote::new("sync-url", &url))?;
        }
        "restore" => {
            if parsed.args.len() != 3 {
                let optional: Vec<&str> = std::iter::once("--force").chain(AWS_USAGE.iter().copied()).collect();
                return Err(backup_usage(&command, &["remote_url", "new_db_name"], &optional));
            }
            let url = absolute_url(&data_dir, &parsed.args[1])?;
            let src = open_remote(&Remote::new("restore", &url))?;
            let name = parsed.args[2].clone();
            let engine = ctx.session.engine.clone();
            if data_dir.join(&name).exists() {
                if !parsed.has("force") {
                    return Err(error(format!("database '{name}' already exists, use '--force' to overwrite")));
                }
                engine.drop_database(&name)?;
            }
            engine.create_database(&name, &ctx.session.user, &ctx.session.host)?;
            let mut dest = Database::open(&data_dir.join(&name).join(".dolt/noms"))?;
            let root = src.root();
            dest.pull(&src, root)?;
            dest.replace_root(root)?;
            dest.close()?;
            src.close()?;
        }
        _ => return Err(error(format!("unrecognized dolt_backup parameter '{command}'"))),
    }
    Ok(Value::Int8(0))
}

/// remote_rows returns the rows of the dolt_remotes table, or of dolt_backups with only names, URLs, and parameters.
pub fn remote_rows(ctx: &mut Ctx<'_>, backups: bool) -> Result<Vec<Vec<Value>>> {
    let state = RepoState::load(&database_dir(ctx))?;
    let params = |r: &Remote| Value::Json(r.json()["params"].to_string());
    Ok(if backups {
        state
            .backups
            .values()
            .map(|r| vec![Value::Text(r.name.clone()), Value::Text(r.url.clone()), params(r)])
            .collect()
    } else {
        state
            .remotes
            .values()
            .map(|r| {
                let specs = Json::Array(r.fetch_specs.iter().cloned().map(Json::String).collect()).to_string();
                vec![Value::Text(r.name.clone()), Value::Text(r.url.clone()), Value::Json(specs), params(r)]
            })
            .collect()
    })
}

/// upstream returns the remote and remote branch that a local branch tracks, which the dolt_branches table shows.
pub fn upstream(ctx: &Ctx<'_>, branch: &str) -> Result<(String, String)> {
    let state = RepoState::load(&database_dir(ctx))?;
    Ok(state
        .branches
        .get(branch)
        .map(|u| (u.remote.clone(), u.merge.strip_prefix("refs/heads/").unwrap_or(&u.merge).to_string()))
        .unwrap_or_default())
}
