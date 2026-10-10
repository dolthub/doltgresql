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

//! Git remotes: a database's files kept as blobs in the commits of a git ref, which a local bare repository caches and
//! syncs with the git remote through the git command line, as Dolt's GitBlobstore and GitRemoteFactory do.

use std::collections::{BTreeMap, HashMap, VecDeque};
use std::io::Write;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

use sha2::{Digest, Sha256};
use store::{Blob, BlobRange, Blobstore, Error, Result, not_found};

/// DATA_REF is the git ref whose commits hold the database's files, unless the remote's git_ref parameter names
/// another.
const DATA_REF: &str = "refs/dolt/data";

/// REMOTE_NAME is the name of the git remote in the cache repository, unless the git_remote_name parameter names
/// another.
const REMOTE_NAME: &str = "origin";

/// INFO_BRANCH is the branch, pushed after each write, whose DOLT_REMOTE.md says that the repository is a Dolt remote.
const INFO_BRANCH: &str = "__dolt_remote_info__";

/// INFO_BRANCH_ENV overrides the info branch's name, where an empty name turns it off.
const INFO_BRANCH_ENV: &str = "DOLT_REMOTE_INFO_BRANCH";

/// MAX_PART_SIZE is the size of the largest blob a write makes, beyond which a file becomes a tree of numbered parts.
const MAX_PART_SIZE: u64 = 50 * 1024 * 1024;

/// PART_NAME_WIDTH is the width of a part's zero-padded number.
const PART_NAME_WIDTH: usize = 4;

/// MAX_PARENTED_COMMITS is how long a chain of commits grows before a write starts a new one without a parent, which
/// lets git's garbage collection reclaim old objects.
const MAX_PARENTED_COMMITS: usize = 64;

/// SYNC_TTL is how long a fetch keeps answering reads before another fetch.
const SYNC_TTL: Duration = Duration::from_secs(1);

/// WRITE_ATTEMPTS is how many times a write tries again after the remote moved under it.
const WRITE_ATTEMPTS: u32 = 32;

/// MANIFEST is the key of the manifest, the only blob whose content changes in place.
const MANIFEST: &str = "manifest";

/// BLOB_CACHE_BYTES caps the bytes of the blobs that reads keep, since git reads a blob only from its start.
const BLOB_CACHE_BYTES: usize = 512 << 20;

/// NEXT_INSTANCE numbers the blobstores of this process, which name their own refs.
static NEXT_INSTANCE: AtomicU64 = AtomicU64::new(0);

/// open opens the git remote at a `git+<scheme>://` URL, given its parameters, which name the directory of the local
/// cache repositories in git_cache_root, preparing the cache repository and checking that the remote has a branch.
pub fn open(url: &str, params: &BTreeMap<String, String>) -> Result<GitBlobstore> {
    let underlying = url.split_once('+').map(|(_, rest)| rest).unwrap_or(url);
    let underlying = underlying.split(['?', '#']).next().unwrap_or(underlying);
    let git_url = git_url_string(underlying);
    let reference = params.get("git_ref").map(|r| r.trim()).filter(|r| !r.is_empty()).unwrap_or(DATA_REF);
    let remote_name = params.get("git_remote_name").map(|r| r.trim()).filter(|r| !r.is_empty()).unwrap_or(REMOTE_NAME);
    let Some(cache_root) = params.get("git_cache_root").filter(|r| !r.trim().is_empty()) else {
        return Err(other("git_cache_root is required for git remotes"));
    };
    let digest = Sha256::digest(format!("{underlying}|{reference}").as_bytes());
    let hash_dir = Path::new(cache_root).join(digest.iter().map(|b| format!("{b:02x}")).collect::<String>());
    std::fs::create_dir_all(&hash_dir)?;
    let lock = std::fs::File::create(hash_dir.join("init.lock"))?;
    lock.lock()?;
    let git_dir = hash_dir.join("repo.git");
    if !git_dir.exists() {
        std::fs::create_dir_all(&git_dir)?;
        git(&git_dir, &["init", "--bare"], None, None, &[])?;
    }
    if git(&git_dir, &["remote", "add", "--", remote_name, &git_url], None, None, &[]).is_err() {
        git(&git_dir, &["remote", "set-url", "--", remote_name, &git_url], None, None, &[])?;
    }
    let heads = git(&git_dir, &["ls-remote", "--heads", "--", remote_name], None, None, &[])?;
    if heads.trim_ascii().is_empty() {
        return Err(other(format!(
            "git remote has no branches: cannot push to {git_url:?}; initialize the repository with an initial \
             branch/commit first"
        )));
    }
    lock.unlock()?;
    GitBlobstore::new(git_dir, reference, remote_name, params)
}

/// git_url_string returns the URL that git commands reach a remote at, turning an ssh URL whose path starts with
/// `/./` back into the scp-style form relative to the user's home directory that it was written as.
fn git_url_string(url: &str) -> String {
    if let Some(rest) = url.strip_prefix("ssh://")
        && let Some((host, path)) = rest.split_once('/')
        && let Some(relative) = path.strip_prefix("./")
    {
        return format!("{host}:{relative}");
    }
    url.to_string()
}

/// other returns a storage error with a message.
fn other(message: impl Into<String>) -> Error {
    Error::Io(std::io::Error::other(message.into()))
}

/// Failure is a git command that failed, with its exit code and its output.
struct Failure {
    exit: i32,
    output: Vec<u8>,
}

impl Failure {
    /// says reports whether the command's output, in lower case, holds any of the phrases.
    fn says(&self, phrases: &[&str]) -> bool {
        let output = String::from_utf8_lossy(&self.output).to_lowercase();
        phrases.iter().any(|p| output.contains(p))
    }

    /// error returns the storage error that a failed command makes.
    fn error(&self, args: &[&str]) -> Error {
        other(format!(
            "git command failed (exit {})\ncommand: git {}\noutput:\n{}",
            self.exit,
            args.join(" "),
            String::from_utf8_lossy(&self.output).trim_end()
        ))
    }
}

/// run runs a git command against a repository, with an index file and input when given and extra environment
/// variables, in English, returning its output or how it failed.
fn run(
    git_dir: &Path,
    args: &[&str],
    index: Option<&Path>,
    input: Option<&[u8]>,
    env: &[(&str, &str)],
) -> std::result::Result<Vec<u8>, Failure> {
    let mut command = Command::new("git");
    command.args(args).env("GIT_DIR", git_dir).env("LC_ALL", "C").env("GIT_TERMINAL_PROMPT", "0");
    if let Some(index) = index {
        command.env("GIT_INDEX_FILE", index);
    }
    for (key, value) in env {
        command.env(key, value);
    }
    command.stdin(if input.is_some() { Stdio::piped() } else { Stdio::null() });
    command.stdout(Stdio::piped()).stderr(Stdio::piped());
    let mut child = command.spawn().map_err(|err| Failure { exit: -1, output: err.to_string().into_bytes() })?;
    if let (Some(input), Some(mut stdin)) = (input, child.stdin.take()) {
        let written = stdin.write_all(input);
        drop(stdin);
        if let Err(err) = written {
            let _ = child.kill();
            return Err(Failure { exit: -1, output: err.to_string().into_bytes() });
        }
    }
    let output = child.wait_with_output().map_err(|err| Failure { exit: -1, output: err.to_string().into_bytes() })?;
    match output.status.success() {
        true => Ok(output.stdout),
        false => {
            let mut text = output.stdout;
            text.extend(output.stderr);
            Err(Failure { exit: output.status.code().unwrap_or(-1), output: text })
        }
    }
}

/// git runs a git command, failing with its output when it fails.
fn git(
    git_dir: &Path,
    args: &[&str],
    index: Option<&Path>,
    input: Option<&[u8]>,
    env: &[(&str, &str)],
) -> Result<Vec<u8>> {
    run(git_dir, args, index, input, env).map_err(|failure| failure.error(args))
}

/// Build builds a write's commit on the remote's head, returning the commit and the paths it pruned.
type Build<'a> = dyn FnMut(&GitBlobstore, &mut State, Option<&str>) -> Result<(String, Vec<String>)> + 'a;

/// Kind is the kind of a git object at a path.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum Kind {
    Blob,
    Tree,
}

/// Entry is an entry of a git tree: its kind, object, and name.
#[derive(Clone, Debug)]
struct Entry {
    kind: Kind,
    oid: String,
    name: String,
}

/// parse_tree_line reads a line of git ls-tree's output: `<mode> <type> <oid>\t<name>`.
fn parse_tree_line(line: &str) -> Option<Entry> {
    let (left, name) = line.split_once('\t')?;
    let mut fields = left.split_whitespace();
    let (kind, oid) = (fields.nth(1)?, fields.next()?);
    let kind = match kind {
        "blob" => Kind::Blob,
        "tree" => Kind::Tree,
        _ => return None,
    };
    Some(Entry { kind, oid: oid.to_string(), name: name.to_string() })
}

/// Plan is how a write lays a blob out: the blobs it adds at their paths, which are numbered parts of a tree when
/// the blob is chunked.
#[derive(Clone, Debug)]
struct Plan {
    writes: Vec<(String, String)>,
    chunked: bool,
}

/// State is what a git blobstore knows of the remote: the paths of the last commit it merged, with the blobs it has
/// put since, the time it last synced, the writes waiting for the next manifest, and the blobs it has read.
#[derive(Default)]
struct State {
    head: String,
    objects: HashMap<String, (String, Kind)>,
    children: HashMap<String, Vec<Entry>>,
    synced_at: Option<Instant>,
    pending: Vec<(String, Plan)>,
    anchor: String,
    blobs: HashMap<String, Arc<Vec<u8>>>,
    blob_order: VecDeque<String>,
    blob_bytes: usize,
}

/// GitBlobstore keeps blobs in the commits of a git remote's ref, as Dolt's GitBlobstore does: files other than the
/// manifest are content-addressed and wait in the cache repository until a manifest write commits and pushes them
/// along with it, pruning the files that the manifest no longer names.
pub struct GitBlobstore {
    git_dir: PathBuf,
    remote_name: String,
    remote_ref: String,
    tracking_ref: String,
    local_ref: String,
    anchor_ref: String,
    max_history: usize,
    reset_on_prune: bool,
    info_branch: String,
    state: Mutex<State>,
}

impl GitBlobstore {
    /// new returns a blobstore over a cache repository's remote and ref, with history options from the remote's
    /// parameters or their environment variables.
    fn new(git_dir: PathBuf, reference: &str, remote_name: &str, params: &BTreeMap<String, String>) -> Result<Self> {
        let setting = |key: &str, env: &str| std::env::var(env).ok().or_else(|| params.get(key).cloned());
        let max_history = match setting("git_remote_max_history_commits", "DOLT_GIT_REMOTE_MAX_HISTORY_COMMITS") {
            Some(value) => value.trim().parse::<usize>().map_err(|_| {
                other(format!("max history commits must be a non-negative integer (0 means unlimited), got {value:?}"))
            })?,
            None => MAX_PARENTED_COMMITS,
        };
        let reset_on_prune =
            match setting("git_remote_reset_history_on_prune", "DOLT_GIT_REMOTE_RESET_HISTORY_ON_PRUNE") {
                Some(value) if value == "false" => false,
                Some(value) if value == "true" => true,
                Some(value) => {
                    return Err(other(format!("reset history on prune must be true or false, got {value:?}")));
                }
                None => true,
            };
        let trimmed = reference.strip_prefix("refs/").unwrap_or(reference);
        let instance = format!(
            "{:x}-{:x}-{:x}",
            std::process::id(),
            SystemTime::now().duration_since(UNIX_EPOCH).map_or(0, |d| d.as_nanos()),
            NEXT_INSTANCE.fetch_add(1, Ordering::Relaxed)
        );
        let anchor_ref = format!("refs/dolt/remotes/{remote_name}/{trimmed}/last");
        let store = GitBlobstore {
            tracking_ref: format!("refs/dolt/remotes/{remote_name}/{trimmed}/{instance}"),
            local_ref: format!("refs/dolt/blobstore/{remote_name}/{trimmed}/{instance}"),
            anchor_ref,
            git_dir,
            remote_name: remote_name.to_string(),
            remote_ref: reference.to_string(),
            max_history,
            reset_on_prune,
            info_branch: std::env::var(INFO_BRANCH_ENV)
                .map(|b| b.trim().to_string())
                .unwrap_or_else(|_| INFO_BRANCH.to_string()),
            state: Mutex::new(State::default()),
        };
        let anchor = store.resolve(&store.anchor_ref)?.unwrap_or_default();
        store.lock().anchor = anchor;
        Ok(store)
    }

    /// lock returns the blobstore's state.
    fn lock(&self) -> std::sync::MutexGuard<'_, State> {
        self.state.lock().unwrap_or_else(|poisoned| poisoned.into_inner())
    }

    /// git runs a git command against the cache repository.
    fn git(&self, args: &[&str], index: Option<&Path>, input: Option<&[u8]>) -> Result<Vec<u8>> {
        git(&self.git_dir, args, index, input, &[])
    }

    /// git_line runs a git command and returns its output's first line.
    fn git_line(&self, args: &[&str], index: Option<&Path>, input: Option<&[u8]>) -> Result<String> {
        let output = self.git(args, index, input)?;
        Ok(String::from_utf8_lossy(&output).lines().next().unwrap_or_default().trim().to_string())
    }

    /// resolve returns the commit a ref points at, or None when there is no such ref.
    fn resolve(&self, reference: &str) -> Result<Option<String>> {
        let spec = format!("{reference}^{{commit}}");
        let args = ["rev-parse", "--verify", "--quiet", &spec];
        match run(&self.git_dir, &args, None, None, &[]) {
            Ok(output) => Ok(Some(String::from_utf8_lossy(&output).trim().to_string()).filter(|s| !s.is_empty())),
            Err(failure)
                if (failure.exit == 1 && failure.output.trim_ascii().is_empty())
                    || failure.says(&["needed a single revision", "unknown revision", "not a valid object name"]) =>
            {
                Ok(None)
            }
            Err(failure) => Err(failure.error(&args)),
        }
    }

    /// fetch fetches the remote's ref into this blobstore's tracking ref, returning the remote's head, or None when the
    /// remote has no such ref, which stands for an empty store.
    fn fetch(&self) -> Result<Option<String>> {
        let refspec = format!("+{}:{}", self.remote_ref, self.tracking_ref);
        let args = ["fetch", "--no-tags", "--refmap=", &self.remote_name, &refspec];
        match run(&self.git_dir, &args, None, None, &[]) {
            Ok(_) => self.resolve(&self.tracking_ref)?.map(Some).ok_or_else(|| not_found(&self.tracking_ref)),
            Err(failure)
                if failure.says(&[
                    "couldn't find remote ref",
                    "could not find remote ref",
                    "remote ref does not exist",
                ]) =>
            {
                Ok(None)
            }
            Err(failure) => Err(failure.error(&args)),
        }
    }

    /// merge_cache adds the paths of a commit's tree to the cache, overwriting only the manifest, since every other
    /// path is content-addressed.
    fn merge_cache(&self, state: &mut State, head: &str) -> Result<()> {
        state.synced_at = Some(Instant::now());
        if state.head == head {
            return Ok(());
        }
        let tree = format!("{head}^{{tree}}");
        let listing = self.git(&["ls-tree", "-r", "-t", &tree], None, None)?;
        for entry in String::from_utf8_lossy(&listing).lines().filter_map(parse_tree_line) {
            let overwrite = entry.name == MANIFEST;
            if overwrite || !state.objects.contains_key(&entry.name) {
                state.objects.insert(entry.name.clone(), (entry.oid.clone(), entry.kind));
            }
            let (parent, base) = split_path(&entry.name);
            let child = Entry { name: base.to_string(), ..entry.clone() };
            let children = state.children.entry(parent.to_string()).or_default();
            match children.iter_mut().find(|c| c.name == child.name) {
                Some(existing) if overwrite => *existing = child,
                Some(_) => {}
                None => {
                    children.push(child);
                    children.sort_by(|a, b| a.name.cmp(&b.name));
                }
            }
        }
        state.head = head.to_string();
        Ok(())
    }

    /// sync_for_read fetches the remote and merges its head into the cache, unless a fetch did within the last second.
    fn sync_for_read(&self, state: &mut State) -> Result<()> {
        if state.synced_at.is_some_and(|at| at.elapsed() < SYNC_TTL) {
            return Ok(());
        }
        match self.fetch()? {
            Some(head) => self.merge_cache(state, &head),
            None => Ok(()),
        }
    }

    /// absent reports that a content-addressed key is missing on the cache's authority, since the cache holds a whole
    /// commit's listing and a fetch could only add keys that a newer manifest names.
    fn absent(&self, state: &State, key: &str) -> bool {
        key != MANIFEST && !state.head.is_empty() && !state.objects.contains_key(key)
    }

    /// read returns a blob's bytes, keeping recent blobs since git reads them only from their start.
    fn read(&self, state: &mut State, oid: &str) -> Result<Arc<Vec<u8>>> {
        if let Some(bytes) = state.blobs.get(oid) {
            return Ok(bytes.clone());
        }
        let bytes = Arc::new(self.git(&["cat-file", "blob", oid], None, None)?);
        state.blob_bytes += bytes.len();
        state.blobs.insert(oid.to_string(), bytes.clone());
        state.blob_order.push_back(oid.to_string());
        while state.blob_bytes > BLOB_CACHE_BYTES && state.blob_order.len() > 1 {
            let Some(old) = state.blob_order.pop_front() else { break };
            if let Some(evicted) = state.blobs.remove(&old) {
                state.blob_bytes -= evicted.len();
            }
        }
        Ok(bytes)
    }

    /// content returns the bytes of a key that the cache holds, joining a chunked key's numbered parts, with its
    /// version: the key's blob or tree object.
    fn content(&self, state: &mut State, key: &str) -> Result<(Vec<u8>, String)> {
        let Some((oid, kind)) = state.objects.get(key).cloned() else { return Err(not_found(key)) };
        match kind {
            Kind::Blob => Ok((self.read(state, &oid)?.to_vec(), oid)),
            Kind::Tree => {
                let parts = state.children.get(key).cloned().unwrap_or_default();
                if parts.is_empty() {
                    return Err(other("gitblobstore: chunked tree has no parts"));
                }
                let mut bytes = Vec::new();
                for (i, part) in parts.iter().enumerate() {
                    let expected = format!("{:0width$}", i + 1, width = PART_NAME_WIDTH);
                    if part.kind != Kind::Blob || part.name != expected {
                        return Err(other(format!(
                            "gitblobstore: invalid part name {:?} (expected {expected:?})",
                            part.name
                        )));
                    }
                    bytes.extend_from_slice(&self.read(state, &part.oid)?);
                }
                Ok((bytes, oid))
            }
        }
    }

    /// plan hashes a blob into the cache repository, as one blob or as numbered parts of at most the part size.
    fn plan(&self, key: &str, data: &[u8]) -> Result<Plan> {
        if data.len() as u64 <= MAX_PART_SIZE {
            let oid = self.git_line(&["hash-object", "-w", "--stdin"], None, Some(data))?;
            return Ok(Plan { writes: vec![(key.to_string(), oid)], chunked: false });
        }
        let mut writes = Vec::new();
        for (i, part) in data.chunks(MAX_PART_SIZE as usize).enumerate() {
            let oid = self.git_line(&["hash-object", "-w", "--stdin"], None, Some(part))?;
            writes.push((format!("{key}/{:0width$}", i + 1, width = PART_NAME_WIDTH), oid));
        }
        Ok(Plan { writes, chunked: true })
    }

    /// defer queues a plan for the next manifest write and caches its paths, returning the key's version.
    fn defer(&self, state: &mut State, key: &str, plan: Plan) -> String {
        let version = plan.writes[0].1.clone();
        if plan.chunked {
            state.objects.entry(key.to_string()).or_insert_with(|| (version.clone(), Kind::Tree));
            for (path, oid) in &plan.writes {
                state.objects.entry(path.clone()).or_insert_with(|| (oid.clone(), Kind::Blob));
                let (parent, base) = split_path(path);
                let children = state.children.entry(parent.to_string()).or_default();
                if !children.iter().any(|c| c.name == base) {
                    children.push(Entry { kind: Kind::Blob, oid: oid.clone(), name: base.to_string() });
                    children.sort_by(|a, b| a.name.cmp(&b.name));
                }
            }
        } else {
            state.objects.entry(key.to_string()).or_insert_with(|| (version.clone(), Kind::Blob));
        }
        state.pending.push((key.to_string(), plan));
        version
    }

    /// write commits a change built on the remote's head and pushes it with a lease on that head, fetching and building
    /// again when another writer moved the head, and returns the key's version after the push.
    fn write(&self, state: &mut State, key: &str, message: &str, build: &mut Build<'_>) -> Result<String> {
        let mut delay = Duration::from_millis(5);
        for attempt in 0..WRITE_ATTEMPTS {
            let head = self.fetch()?;
            if let Some(head) = &head {
                self.git(&["update-ref", "-m", "gitblobstore: sync write", &self.local_ref, head], None, None)?;
                self.merge_cache(state, head)?;
            }
            let (commit, pruned) = build(self, state, head.as_deref())?;
            self.git(&["update-ref", "-m", message, &self.local_ref, &commit], None, None)?;
            let lease = format!("--force-with-lease={}:{}", self.remote_ref, head.as_deref().unwrap_or(""));
            let refspec = format!("{}:{}", self.local_ref, self.remote_ref);
            let pushed =
                run(&self.git_dir, &["push", "--porcelain", &lease, &self.remote_name, &refspec], None, None, &[]);
            if pushed.is_err() && attempt + 1 < WRITE_ATTEMPTS {
                std::thread::sleep(delay);
                delay = (delay * 2).min(Duration::from_millis(320));
                continue;
            }
            pushed.map_err(|failure| failure.error(&["push", "--porcelain", &lease, &self.remote_name, &refspec]))?;
            self.push_info_branch(&commit);
            self.merge_cache(state, &commit)?;
            for path in pruned {
                state.objects.remove(&path);
                state.children.remove(split_path(&path).0);
            }
            let spec = format!("{commit}:{key}");
            let oid = self.git_line(&["rev-parse", "--verify", &spec], None, None)?;
            let kind = match self.git_line(&["cat-file", "-t", &oid], None, None)?.as_str() {
                "tree" => Kind::Tree,
                _ => Kind::Blob,
            };
            state.objects.insert(key.to_string(), (oid.clone(), kind));
            let (parent, base) = split_path(key);
            let child = Entry { kind, oid: oid.clone(), name: base.to_string() };
            let children = state.children.entry(parent.to_string()).or_default();
            match children.iter_mut().find(|c| c.name == base) {
                Some(existing) => *existing = child,
                None => children.push(child),
            }
            return Ok(oid);
        }
        Err(other("gitblobstore: write retries exhausted"))
    }

    /// build_commit builds the commit of a write on a parent commit: the key's plan and any pending writes, without the
    /// paths that the manifest's table names leave out when it names them, returning the commit and the pruned paths.
    fn build_commit(
        &self,
        parent: Option<&str>,
        key: &str,
        plan: &Plan,
        message: &str,
        extra: &[(String, Plan)],
        allowed: Option<&std::collections::HashSet<String>>,
    ) -> Result<(String, Vec<String>)> {
        let index = TempIndex::new()?;
        let index_path = Some(index.0.as_path());
        match parent {
            Some(parent) => self.git(&["read-tree", &format!("{parent}^{{tree}}")], index_path, None)?,
            None => self.git(&["read-tree", "--empty"], index_path, None)?,
        };
        let mut pruned = Vec::new();
        if let (Some(allowed), Some(parent)) = (allowed, parent) {
            let listing = self.git(&["ls-tree", "-r", "-t", &format!("{parent}^{{tree}}")], None, None)?;
            pruned = String::from_utf8_lossy(&listing)
                .lines()
                .filter_map(parse_tree_line)
                .map(|e| e.name)
                .filter(|path| !referenced(path, allowed))
                .collect();
            self.remove_paths(index_path, &pruned)?;
        }
        for (pending_key, pending) in extra {
            if let Some(parent) = parent {
                self.remove_conflicts(parent, index_path, pending_key, pending.chunked)?;
            }
            for (path, oid) in &pending.writes {
                self.git(&["update-index", "--add", "--cacheinfo", "100644", oid, path], index_path, None)?;
            }
        }
        if let Some(parent) = parent {
            self.remove_conflicts(parent, index_path, key, plan.chunked)?;
        }
        let mut writes = plan.writes.clone();
        writes.sort();
        for (path, oid) in &writes {
            self.git(&["update-index", "--add", "--cacheinfo", "100644", oid, path], index_path, None)?;
        }
        let tree = self.git_line(&["write-tree"], index_path, None)?;
        let parent = match parent {
            Some(_) if self.reset_on_prune && !pruned.is_empty() => None,
            Some(parent) if self.max_history == 0 => Some(parent),
            Some(parent) => {
                let limit = format!("--max-count={}", self.max_history);
                let depth: usize =
                    self.git_line(&["rev-list", "--count", &limit, parent], None, None)?.parse().unwrap_or(0);
                (depth < self.max_history).then_some(parent)
            }
            None => None,
        };
        Ok((self.commit_tree(&tree, parent, message)?, pruned))
    }

    /// commit_tree commits a tree, falling back to Dolt's own identity when git has none configured.
    fn commit_tree(&self, tree: &str, parent: Option<&str>, message: &str) -> Result<String> {
        let mut args = vec!["commit-tree", tree, "-m", message];
        if let Some(parent) = parent {
            args.extend(["-p", parent]);
        }
        match run(&self.git_dir, &args, None, None, &[]) {
            Ok(output) => Ok(String::from_utf8_lossy(&output).trim().to_string()),
            Err(failure)
                if failure.says(&[
                    "author identity unknown",
                    "unable to auto-detect email address",
                    "empty ident name",
                ]) =>
            {
                let env = [
                    ("GIT_AUTHOR_NAME", "dolt gitblobstore"),
                    ("GIT_COMMITTER_NAME", "dolt gitblobstore"),
                    ("GIT_AUTHOR_EMAIL", "gitblobstore@dolt.invalid"),
                    ("GIT_COMMITTER_EMAIL", "gitblobstore@dolt.invalid"),
                ];
                let output = git(&self.git_dir, &args, None, None, &env)?;
                Ok(String::from_utf8_lossy(&output).trim().to_string())
            }
            Err(failure) => Err(failure.error(&args)),
        }
    }

    /// remove_paths removes paths from an index, which a bare repository does by writing them with mode 0.
    fn remove_paths(&self, index: Option<&Path>, paths: &[String]) -> Result<()> {
        if paths.is_empty() {
            return Ok(());
        }
        let input: String =
            paths.iter().map(|p| format!("0 0000000000000000000000000000000000000000 0\t{p}\n")).collect();
        self.git(&["update-index", "--index-info"], index, Some(input.as_bytes()))?;
        Ok(())
    }

    /// remove_conflicts removes what the parent commit has at a key that a new write of the key would conflict with:
    /// a blob that a chunked write replaces, or the parts of a tree.
    fn remove_conflicts(&self, parent: &str, index: Option<&Path>, key: &str, chunked: bool) -> Result<()> {
        let spec = format!("{parent}:{key}");
        let Ok(output) = run(&self.git_dir, &["rev-parse", "--verify", &spec], None, None, &[]) else { return Ok(()) };
        let oid = String::from_utf8_lossy(&output).trim().to_string();
        match self.git_line(&["cat-file", "-t", &oid], None, None)?.as_str() {
            "blob" if chunked => self.remove_paths(index, &[key.to_string()]),
            "tree" => {
                let listing = self.git(&["ls-tree", &spec], None, None)?;
                let paths: Vec<String> = String::from_utf8_lossy(&listing)
                    .lines()
                    .filter_map(parse_tree_line)
                    .map(|e| format!("{key}/{}", e.name))
                    .collect();
                self.remove_paths(index, &paths)
            }
            _ => Ok(()),
        }
    }

    /// push_info_branch force-pushes the branch whose DOLT_REMOTE.md says that the repository is a Dolt remote, as a
    /// best effort that never fails a write.
    fn push_info_branch(&self, head: &str) {
        if self.info_branch.is_empty() {
            return;
        }
        let pushed = || -> Result<()> {
            let now = chrono::Utc::now().to_rfc3339_opts(chrono::SecondsFormat::Secs, true);
            let content = format!(
                "This repository is being used as a Dolt remote.\n\nref={}\n\nhead={head}\n\ntimestamp={now}\n",
                self.remote_ref
            );
            let blob = self.git_line(&["hash-object", "-w", "--stdin"], None, Some(content.as_bytes()))?;
            let index = TempIndex::new()?;
            let index_path = Some(index.0.as_path());
            self.git(&["read-tree", "--empty"], index_path, None)?;
            self.git(&["update-index", "--add", "--cacheinfo", "100644", &blob, "DOLT_REMOTE.md"], index_path, None)?;
            let tree = self.git_line(&["write-tree"], index_path, None)?;
            let commit = self.commit_tree(&tree, None, "dolt remote info")?;
            let local = format!("refs/dolt/info/{}", self.info_branch);
            self.git(&["update-ref", "-m", "dolt remote info", &local, &commit], None, None)?;
            let refspec = format!("{local}:refs/heads/{}", self.info_branch);
            self.git(&["push", "--force", &self.remote_name, &refspec], None, None)?;
            Ok(())
        };
        let _ = pushed();
    }
}

impl Drop for GitBlobstore {
    /// drop keeps the head this blobstore last saw as the shared fetch anchor, unless another blobstore moved it, and
    /// deletes this blobstore's own refs, as Dolt's Teardown does.
    fn drop(&mut self) {
        let state = self.lock();
        let (head, anchor) = (state.head.clone(), state.anchor.clone());
        drop(state);
        if !head.is_empty() {
            let _ = self.git(
                &["update-ref", "-m", "gitblobstore: retain fetch anchor", &self.anchor_ref, &head, &anchor],
                None,
                None,
            );
        }
        for reference in [&self.local_ref, &self.tracking_ref] {
            if self.resolve(reference).ok().flatten().is_some() {
                let _ = self.git(&["update-ref", "-d", reference], None, None);
            }
        }
    }
}

/// TempIndex is a temporary git index file, deleted when dropped.
struct TempIndex(PathBuf);

impl TempIndex {
    /// new returns the path of a fresh temporary index file, which git creates on first use.
    fn new() -> Result<TempIndex> {
        let name = format!("dolt-git-index-{}-{}", std::process::id(), NEXT_INSTANCE.fetch_add(1, Ordering::Relaxed));
        Ok(TempIndex(std::env::temp_dir().join(name)))
    }
}

impl Drop for TempIndex {
    /// drop deletes the index file.
    fn drop(&mut self) {
        let _ = std::fs::remove_file(&self.0);
    }
}

/// split_path splits a tree path into its parent directory and its base name.
fn split_path(path: &str) -> (&str, &str) {
    match path.rsplit_once('/') {
        Some((parent, base)) => (parent, base),
        None => ("", path),
    }
}

/// manifest_names returns the table file names that a manifest's text names, or None when it is not a version 4 or 5
/// manifest.
fn manifest_names(data: &[u8]) -> Option<std::collections::HashSet<String>> {
    let text = String::from_utf8_lossy(data);
    let parts: Vec<&str> = text.split(':').collect();
    let start = match parts.first() {
        Some(&"5") => 5,
        Some(&"4") => 4,
        _ => return None,
    };
    let specs = parts.get(start..)?;
    if specs.len() % 2 != 0 {
        return None;
    }
    Some(specs.iter().step_by(2).filter(|n| !n.is_empty()).map(|n| n.to_string()).collect())
}

/// referenced reports whether a tree path is the manifest or belongs to a table file that a manifest names, ignoring
/// a chunked file's part and the suffixes of a table file's records and tail.
fn referenced(path: &str, allowed: &std::collections::HashSet<String>) -> bool {
    if path == MANIFEST {
        return true;
    }
    let base = path.split('/').next().unwrap_or(path);
    let base = [".darc.records", ".darc.tail", ".records", ".tail", ".darc"]
        .iter()
        .find_map(|suffix| base.strip_suffix(suffix))
        .unwrap_or(base);
    allowed.contains(base)
}

/// normalize checks a key as a git tree path, turning backslashes into slashes.
fn normalize(key: &str) -> Result<String> {
    let key = key.replace('\\', "/");
    let invalid = |why: &str| other(format!("invalid git blobstore key ({why}): {key:?}"));
    if key.contains('\0') {
        return Err(invalid("NUL byte"));
    }
    if key.is_empty() {
        return Err(invalid("empty"));
    }
    if key.starts_with('/') {
        return Err(invalid("absolute path"));
    }
    for part in key.split('/') {
        match part {
            "" => return Err(invalid("empty path segment")),
            "." | ".." => return Err(invalid("path traversal")),
            _ => {}
        }
    }
    Ok(key)
}

impl Blobstore for GitBlobstore {
    fn path(&self) -> String {
        format!("{}@{}", self.git_dir.display(), self.remote_ref)
    }

    fn exists(&self, key: &str) -> Result<bool> {
        let key = normalize(key)?;
        let mut state = self.lock();
        if key != MANIFEST {
            if state.objects.contains_key(&key) {
                return Ok(true);
            }
            if self.absent(&state, &key) {
                return Ok(false);
            }
        }
        self.sync_for_read(&mut state)?;
        Ok(state.objects.contains_key(&key))
    }

    fn get(&self, key: &str, range: BlobRange) -> Result<Blob> {
        let key = normalize(key)?;
        let mut state = self.lock();
        if key == MANIFEST || !state.objects.contains_key(&key) {
            if self.absent(&state, &key) {
                return Err(not_found(&key));
            }
            self.sync_for_read(&mut state)?;
        }
        let (bytes, version) = self.content(&mut state, &key)?;
        let size = bytes.len() as u64;
        let range = range.positive(size as i64);
        let start = (range.offset.max(0) as usize).min(bytes.len());
        let end = (start + range.length.max(0) as usize).min(bytes.len());
        Ok(Blob { data: bytes[start..end].to_vec(), size, version })
    }

    fn put(&self, key: &str, data: &[u8]) -> Result<String> {
        let key = normalize(key)?;
        let mut state = self.lock();
        if key != MANIFEST {
            if let Some((oid, _)) = state.objects.get(&key) {
                return Ok(oid.clone());
            }
            let plan = self.plan(&key, data)?;
            return Ok(self.defer(&mut state, &key, plan));
        }
        let plan = self.plan(&key, data)?;
        let message = format!("gitblobstore: put {key}");
        self.write(&mut state, &key, &message, &mut |store, _, parent| {
            store.build_commit(parent, &key, &plan, &message, &[], None)
        })
    }

    fn check_and_put_manifest(&self, expected: &str, data: &[u8]) -> Result<String> {
        let mut state = self.lock();
        let allowed = manifest_names(data);
        let pending = std::mem::take(&mut state.pending);
        let (flush, deferred): (Vec<_>, Vec<_>) = match &allowed {
            Some(allowed) => pending.iter().cloned().partition(|(key, _)| referenced(key, allowed)),
            None => (pending.clone(), Vec::new()),
        };
        let plan = self.plan(MANIFEST, data)?;
        let message = format!("gitblobstore: checkandput {MANIFEST}");
        let written = self.write(&mut state, MANIFEST, &message, &mut |store, state, parent| {
            let actual = match parent {
                Some(_) => state.objects.get(MANIFEST).map(|(oid, _)| oid.clone()).unwrap_or_default(),
                None => String::new(),
            };
            if actual != expected {
                return Err(Error::VersionMismatch {
                    key: MANIFEST.to_string(),
                    expected: expected.to_string(),
                    actual,
                });
            }
            store.build_commit(parent, MANIFEST, &plan, &message, &flush, allowed.as_ref())
        });
        let mut requeued = match &written {
            Ok(_) => deferred,
            Err(_) => pending,
        };
        requeued.append(&mut state.pending);
        state.pending = requeued;
        written
    }

    fn concatenate(&self, key: &str, sources: &[String]) -> Result<String> {
        let key = normalize(key)?;
        if sources.is_empty() {
            return Err(other("gitblobstore: concatenate requires at least one source"));
        }
        let mut state = self.lock();
        if key != MANIFEST
            && let Some((oid, _)) = state.objects.get(&key)
        {
            return Ok(oid.clone());
        }
        let mut data = Vec::new();
        for source in sources {
            let (bytes, _) = self.content(&mut state, &normalize(source)?)?;
            data.extend_from_slice(&bytes);
        }
        let plan = self.plan(&key, &data)?;
        if key != MANIFEST {
            return Ok(self.defer(&mut state, &key, plan));
        }
        let message = format!("gitblobstore: concatenate {key}");
        self.write(&mut state, &key, &message, &mut |store, _, parent| match parent {
            Some(_) => store.build_commit(parent, &key, &plan, &message, &[], None),
            None => Err(not_found(&key)),
        })
    }
}
