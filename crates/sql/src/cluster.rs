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

//! Cluster replication, a port of Dolt's sqle/cluster: a server is the primary or a standby at an epoch, persisted in
//! the global config, and a primary replicates each database to its standby remotes by pushing every chunk of the
//! database's store root to them over the remotes API and moving their roots to the same hash.

use std::collections::HashMap;
use std::path::PathBuf;
use std::sync::atomic::{AtomicI64, Ordering};
use std::sync::{Arc, Condvar, Mutex};
use std::time::{Duration, Instant, SystemTime};

use doltdb::database::Database;
use store::{ChunkStore, Hash};

use crate::error::{PgError, Result, code};

/// ROLE_KEY and EPOCH_KEY name the persisted role and epoch in the global config, as Dolt's PersistentConfigPrefix
/// and variable names do.
const ROLE_KEY: &str = "sqlserver.cluster.dolt_cluster_role";
const EPOCH_KEY: &str = "sqlserver.cluster.dolt_cluster_role_epoch";

/// ACK_TIMEOUT_KEY names the persisted dolt_cluster_ack_writes_timeout_secs in the global config.
const ACK_TIMEOUT_KEY: &str = "sqlserver.global.dolt_cluster_ack_writes_timeout_secs";

/// DATABASE names the read-only database that holds dolt_cluster_status, as Dolt's DoltClusterDbName does.
pub const DATABASE: &str = "dolt_cluster";

/// WAIT is how long a graceful transition to standby waits for the standbys to catch up, as Dolt's
/// waitForHooksToReplicateTimeout.
const WAIT: Duration = Duration::from_secs(10);

/// BROKEN_CONFIG is the error a hook shows once two primaries were found at one epoch.
const BROKEN_CONFIG: &str = "error: more than one server was configured as primary in the same epoch. this server has \
                             stopped accepting writes. choose a primary in the cluster and call dolt_assume_cluster_role() on servers in the \
                             cluster to start replication at a higher epoch";

/// StandbyRemote is a remote that a primary replicates every database to, named by a URL template whose
/// `{database}` stands for the database's name.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct StandbyRemote {
    pub name: String,
    pub url_template: String,
}

/// ClusterConfig is the `cluster:` section of the server's config.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct ClusterConfig {
    pub standby_remotes: Vec<StandbyRemote>,
    pub bootstrap_role: String,
    pub bootstrap_epoch: i64,
    pub remotesapi_port: u16,
}

/// Role is a cluster member's role.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Role {
    Primary,
    Standby,
    DetectedBrokenConfig,
}

impl Role {
    /// name returns the role's name.
    pub fn name(self) -> &'static str {
        match self {
            Role::Primary => "primary",
            Role::Standby => "standby",
            Role::DetectedBrokenConfig => "detected_broken_config",
        }
    }

    /// parse returns the role with the name.
    fn parse(name: &str) -> Option<Role> {
        match name {
            "primary" => Some(Role::Primary),
            "standby" => Some(Role::Standby),
            "detected_broken_config" => Some(Role::DetectedBrokenConfig),
            _ => None,
        }
    }
}

/// Status is a row of dolt_cluster_status: a database's replication to one standby remote.
pub struct Status {
    pub database: String,
    pub remote: String,
    pub role: Role,
    pub epoch: i64,
    pub lag: Option<Duration>,
    pub last_update: Option<SystemTime>,
    pub error: Option<String>,
}

/// Hook replicates one database to one standby remote, as Dolt's commithook does.
pub struct Hook {
    pub database: String,
    pub remote: String,
    pub url: String,
    state: Mutex<HookState>,
}

/// HookState is a hook's progress.
#[derive(Default)]
struct HookState {
    /// The store root to replicate and when it was seen.
    next_head: Hash,
    next_head_time: Option<SystemTime>,
    last_pushed: Hash,
    /// When the root that the standby holds was read, so that a write before it is known to have reached the standby.
    pushed_read: Option<Instant>,
    last_success: Option<SystemTime>,
    next_attempt: Option<Instant>,
    last_heartbeat: Option<Instant>,
    error: Option<String>,
}

/// Replicated is whether a standby caught up on a database, or on the roles or the branch control that the
/// `mysql` and `dolt_branch_control` names stand for, as Dolt's graceTransitionResult reports it.
pub struct Replicated {
    pub caught_up: bool,
    pub database: String,
    pub remote: String,
    pub url: String,
}

/// CaughtUp is whether each standby caught up on each database.
type CaughtUp = Vec<Replicated>;

/// Held is a database whose lock the caller already holds, by name.
pub type Held<'a> = Option<(&'a str, &'a Database)>;

/// with_database runs a function on a database, through the lock that the caller holds when it is the held one.
fn with_database<T>(
    engine: &crate::engine::Engine,
    name: &str,
    held: Held<'_>,
    f: impl FnOnce(&Database) -> T,
) -> Option<T> {
    if let Some((held_name, db)) = held
        && held_name == name
    {
        return Some(f(db));
    }
    let handle = engine.database_handle(name)?;
    let db = handle.write();
    Some(f(&db))
}

/// Cluster is a server's cluster replication state.
pub struct Cluster {
    pub config: ClusterConfig,
    /// The global config file that persists the role and epoch.
    persist: PathBuf,
    state: Mutex<(Role, i64)>,
    hooks: Mutex<Vec<Arc<Hook>>>,
    /// Session ids that a role change ended, which may run no more statements.
    pub ended: Mutex<std::collections::HashSet<u64>>,
    credentials: remotes::cluster::Credentials,
    keys: remotes::cluster::KeySet,
    ack_timeout: AtomicI64,
    /// Whether a write is waiting for the background replication, which it wakes.
    nudged: (Mutex<bool>, Condvar),
    /// What each standby remote received outside of its databases, by name.
    replicas: Mutex<HashMap<String, ReplicaState>>,
}

/// ReplicaState is what a standby received outside of its databases, and the databases it still has to drop.
#[derive(Default)]
struct ReplicaState {
    users: Vec<u8>,
    branch_control: Vec<u8>,
    drops: Vec<String>,
    next_attempt: Option<Instant>,
}

/// global_config_path returns the global config file, under DOLT_ROOT_PATH or else the home directory, as Dolt finds
/// it.
fn global_config_path() -> PathBuf {
    let root = std::env::var_os("DOLT_ROOT_PATH").or_else(|| std::env::var_os("HOME")).unwrap_or_default();
    PathBuf::from(root).join(".dolt").join("config_global.json")
}

/// read_global reads the global config's string entries.
fn read_global(path: &std::path::Path) -> HashMap<String, String> {
    let Ok(text) = std::fs::read_to_string(path) else { return HashMap::new() };
    match crate::json::parse(&text, false) {
        Ok(crate::json::Json::Object(entries)) => entries
            .into_iter()
            .filter_map(|(key, value)| match value {
                crate::json::Json::String(value) => Some((key, value)),
                _ => None,
            })
            .collect(),
        _ => HashMap::new(),
    }
}

/// write_global writes entries into the global config, keeping its other entries.
fn write_global(path: &std::path::Path, entries: &[(&str, String)]) -> Result<()> {
    let mut all = read_global(path);
    for (key, value) in entries {
        all.insert(key.to_string(), value.clone());
    }
    let mut keys: Vec<&String> = all.keys().collect();
    keys.sort();
    let body: Vec<String> = keys
        .iter()
        .map(|key| {
            let mut entry = String::new();
            crate::json::escape(&mut entry, key);
            entry.push(':');
            crate::json::escape(&mut entry, &all[*key]);
            entry
        })
        .collect();
    if let Some(dir) = path.parent() {
        std::fs::create_dir_all(dir).map_err(PgError::internal)?;
    }
    std::fs::write(path, format!("{{{}}}", body.join(","))).map_err(PgError::internal)
}

impl Cluster {
    /// open applies the persisted role and epoch, or else the bootstrap ones, persisting them, as Dolt's
    /// applyBootstrapClusterConfig does.
    pub fn open(config: ClusterConfig) -> std::result::Result<Cluster, String> {
        let persist = global_config_path();
        let persisted = read_global(&persist);
        let (role, from_persisted) = match persisted.get(ROLE_KEY).filter(|r| !r.is_empty()) {
            Some(role) => (role.clone(), true),
            None if config.bootstrap_role.is_empty() => ("primary".to_string(), false),
            None => (config.bootstrap_role.clone(), false),
        };
        let epoch = persisted.get(EPOCH_KEY).cloned().unwrap_or_else(|| config.bootstrap_epoch.to_string());
        let parsed = match Role::parse(&role) {
            Some(Role::DetectedBrokenConfig) if !from_persisted => None,
            parsed => parsed,
        };
        let Some(role) = parsed else {
            return Err(format!(
                "persisted role sqlserver.cluster.dolt_cluster_role = {role} must be \"primary\" or \"secondary\""
            ));
        };
        let epoch: i64 = epoch.parse().map_err(|_| {
            format!("persisted role epoch sqlserver.cluster.dolt_cluster_role_epoch = {epoch} must be an integer")
        })?;
        write_global(&persist, &[(ROLE_KEY, role.name().to_string()), (EPOCH_KEY, epoch.to_string())])
            .map_err(|err| err.message)?;
        let ack_timeout = persisted.get(ACK_TIMEOUT_KEY).and_then(|t| t.parse().ok()).unwrap_or(0);
        let jwks = config.standby_remotes.iter().map(|r| r.url_template.replace("{database}", ".well-known/jwks.json"));
        Ok(Cluster {
            keys: remotes::cluster::KeySet::new(jwks.collect()),
            config,
            persist,
            state: Mutex::new((role, epoch)),
            hooks: Mutex::new(Vec::new()),
            ended: Mutex::default(),
            credentials: remotes::cluster::Credentials::new(),
            ack_timeout: AtomicI64::new(ack_timeout),
            nudged: (Mutex::new(false), Condvar::new()),
            replicas: Mutex::default(),
        })
    }

    /// role returns the role and epoch.
    pub fn role(&self) -> (Role, i64) {
        *self.state.lock().unwrap_or_else(|p| p.into_inner())
    }

    /// add_database makes the hooks that replicate a database to the standby remotes, returning the remotes' names
    /// and URLs for the database's own remote list.
    pub fn add_database(&self, database: &str) -> Vec<(String, String)> {
        let mut hooks = self.hooks.lock().unwrap_or_else(|p| p.into_inner());
        let mut remotes = Vec::new();
        for remote in &self.config.standby_remotes {
            let url = remote.url_template.replace("{database}", database);
            remotes.push((remote.name.clone(), url.clone()));
            if hooks.iter().any(|h| h.database == database && h.remote == remote.name) {
                continue;
            }
            hooks.push(Arc::new(Hook {
                database: database.to_string(),
                remote: remote.name.clone(),
                url,
                state: Mutex::default(),
            }));
        }
        remotes
    }

    /// remove_database drops a database's hooks and has each standby drop it too.
    pub fn remove_database(&self, database: &str) {
        self.remove_hooks(database);
        let mut replicas = self.replicas.lock().unwrap_or_else(|p| p.into_inner());
        for remote in &self.config.standby_remotes {
            replicas.entry(remote.name.clone()).or_default().drops.push(database.to_string());
        }
        drop(replicas);
        self.nudge();
    }

    /// remove_hooks drops a database's hooks.
    fn remove_hooks(&self, database: &str) {
        self.hooks.lock().unwrap_or_else(|p| p.into_inner()).retain(|h| h.database != database);
    }

    /// hooks returns the hooks, in database order.
    pub fn hooks(&self) -> Vec<Arc<Hook>> {
        let mut hooks = self.hooks.lock().unwrap_or_else(|p| p.into_inner()).clone();
        hooks.sort_by(|a, b| (&a.database, &a.remote).cmp(&(&b.database, &b.remote)));
        hooks
    }

    /// status returns the rows of dolt_cluster_status.
    pub fn status(&self, engine: &crate::engine::Engine, held: Held<'_>) -> Vec<Status> {
        let (role, epoch) = self.role();
        self.hooks()
            .iter()
            .map(|hook| {
                let root =
                    engine.is_open(&hook.database).then(|| with_database(engine, &hook.database, held, |db| db.root()));
                let mut state = hook.state.lock().unwrap_or_else(|p| p.into_inner());
                if let Some(root) = root.flatten().filter(|r| *r != state.next_head) {
                    state.next_head = root;
                    state.next_head_time = Some(SystemTime::now());
                }
                let lag = match role == Role::Primary && !state.last_pushed.is_empty() {
                    true if state.next_head == state.last_pushed => Some(Duration::ZERO),
                    true => {
                        let behind = state.next_head_time.and_then(|t| t.elapsed().ok()).unwrap_or_default();
                        Some(behind.max(Duration::from_millis(1)))
                    }
                    false => None,
                };
                Status {
                    database: hook.database.clone(),
                    remote: hook.remote.clone(),
                    role,
                    epoch,
                    lag,
                    last_update: state.last_success,
                    error: state.error.clone(),
                }
            })
            .collect()
    }

    /// set_role changes the role and epoch as Dolt's setRoleAndEpoch does, replicating to every standby first for a
    /// graceful transition to standby, and returns whether the role changed with each hook's result.
    pub fn set_role(
        &self,
        engine: &crate::engine::Engine,
        role: &str,
        epoch: i64,
        graceful: bool,
        min_caught_up: usize,
        held: Held<'_>,
    ) -> Result<(bool, CaughtUp)> {
        let error = |message: String| PgError::new(code::INTERNAL_ERROR, message);
        let (current, current_epoch) = self.role();
        if epoch == current_epoch && role == current.name() {
            return Ok((false, Vec::new()));
        }
        let Some(new_role) = Role::parse(role) else {
            return Err(error(format!("error assuming role '{role}'; valid roles are 'primary' and 'standby'")));
        };
        if epoch < current_epoch {
            return Err(error(format!(
                "error assuming role '{role}' at epoch {epoch}; already at epoch {current_epoch}"
            )));
        }
        if epoch == current_epoch && (graceful || new_role == Role::Primary) {
            return Err(error(format!(
                "error assuming role '{role}' at epoch {epoch}; already at epoch {current_epoch} with different role, '{}'",
                current.name()
            )));
        }
        let changed = new_role != current;
        let mut results = Vec::new();
        if changed && new_role == Role::Standby && graceful {
            results = self.wait_for_standbys(engine, held);
            let databases: std::collections::HashSet<&str> = results.iter().map(|r| r.database.as_str()).collect();
            let mut replicas: HashMap<&str, usize> = HashMap::new();
            for result in &results {
                let host = result.url.split("://").nth(1).and_then(|r| r.split('/').next()).unwrap_or_default();
                *replicas.entry(host).or_default() += usize::from(result.caught_up);
            }
            let caught = replicas.values().filter(|&&n| n == databases.len()).count();
            if min_caught_up == 0 && results.iter().any(|r| !r.caught_up) {
                return Err(error(
                    "cluster/controller: failed to transition from primary to standby gracefully; could not \
                     replicate databases to standby in a timely manner."
                        .into(),
                ));
            }
            if min_caught_up > 0 && caught < min_caught_up {
                return Err(error(format!(
                    "cluster/controller: failed to transition from primary to standby gracefully; could not ensure \
                     {min_caught_up} replicas were caught up on all {} databases. Only caught up {caught} standbys fully.",
                    databases.len()
                )));
            }
        }
        *self.state.lock().unwrap_or_else(|p| p.into_inner()) = (new_role, epoch);
        if changed && new_role == Role::Primary {
            engine.refresh_sequences()?;
        }
        if changed {
            for hook in self.hooks() {
                let mut state = hook.state.lock().unwrap_or_else(|p| p.into_inner());
                *state = HookState::default();
                if new_role == Role::DetectedBrokenConfig {
                    state.error = Some(BROKEN_CONFIG.to_string());
                }
            }
            let mut ended = self.ended.lock().unwrap_or_else(|p| p.into_inner());
            ended.extend(engine.session_ids());
        }
        write_global(&self.persist, &[(ROLE_KEY, new_role.name().to_string()), (EPOCH_KEY, epoch.to_string())])?;
        Ok((changed, results))
    }

    /// wait_for_standbys replicates every database, and the roles and branch control, until each standby has them or
    /// the wait runs out, returning whether each standby caught up on each.
    fn wait_for_standbys(&self, engine: &crate::engine::Engine, held: Held<'_>) -> CaughtUp {
        let deadline = Instant::now() + WAIT;
        let hooks = self.hooks();
        let mut done = vec![false; hooks.len()];
        let mut destinations = HashMap::new();
        let mut clients = HashMap::new();
        let synced = loop {
            for (i, hook) in hooks.iter().enumerate() {
                if !done[i] {
                    done[i] = self.replicate(engine, hook, &mut destinations, true, held);
                }
            }
            let synced = self.sync_replicas(engine, &mut clients, true);
            if Instant::now() >= deadline || (done.iter().all(|d| *d) && synced.iter().all(|(_, s)| *s)) {
                break synced;
            }
            std::thread::sleep(Duration::from_millis(100));
        };
        let mut results: CaughtUp = done
            .into_iter()
            .zip(hooks)
            .map(|(caught_up, hook)| Replicated {
                caught_up,
                database: hook.database.clone(),
                remote: hook.remote.clone(),
                url: hook.url.clone(),
            })
            .collect();
        for (remote, caught_up) in synced {
            let url = remote.url_template.split("://").collect::<Vec<_>>();
            let host = url.get(1).and_then(|r| r.split('/').next()).unwrap_or_default();
            let url = format!("{}://{host}", url[0]);
            for database in ["dolt_branch_control", "mysql"] {
                results.push(Replicated {
                    caught_up,
                    database: database.to_string(),
                    remote: remote.name.clone(),
                    url: url.clone(),
                });
            }
        }
        results
    }

    /// replicate pushes a database's store root to a hook's standby when it moved, or sends a heartbeat when one is due,
    /// and reports whether the standby is caught up.
    fn replicate(
        &self,
        engine: &crate::engine::Engine,
        hook: &Hook,
        destinations: &mut HashMap<(String, String), Database>,
        forced: bool,
        held: Held<'_>,
    ) -> bool {
        if self.role().0 != Role::Primary {
            return true;
        }
        let read = Instant::now();
        let Some(root) = with_database(engine, &hook.database, held, |db| db.root()) else { return true };
        let key = (hook.database.clone(), hook.remote.clone());
        let caught_up = {
            let mut state = hook.state.lock().unwrap_or_else(|p| p.into_inner());
            if root != state.next_head {
                state.next_head = root;
                state.next_head_time = Some(SystemTime::now());
            }
            if state.next_head == state.last_pushed {
                state.pushed_read = Some(read);
                if state.last_heartbeat.is_some_and(|t| t.elapsed() < Duration::from_secs(1)) {
                    return true;
                }
                state.last_heartbeat = Some(Instant::now());
            } else if !forced && state.next_attempt.is_some_and(|t| Instant::now() < t) {
                return false;
            }
            state.next_head == state.last_pushed
        };
        if caught_up {
            if let Some(dest) = destinations.get_mut(&key) {
                let _ = dest.commit_root(root, root);
            }
            return true;
        }
        let result = match destinations.get_mut(&key) {
            Some(dest) => Ok(dest),
            None => match self.open_destination(engine, &hook.url) {
                Ok(dest) => Ok(destinations.entry(key.clone()).or_insert(dest)),
                Err(err) => Err(format!("could not replicate to standby: error fetching destDB: {err}")),
            },
        }
        .and_then(|dest| {
            let pushed =
                with_database(engine, &hook.database, held, |db| dest.pull(db, root).map_err(|e| e.to_string()))
                    .unwrap_or_else(|| Err(format!("database {} is not open", hook.database)));
            pushed
                .and_then(|_| match dest.commit_root(root, dest.root()) {
                    Ok(true) => Ok(()),
                    Ok(false) => Err("root hash moved on the destination".to_string()),
                    Err(err) => Err(err.to_string()),
                })
                .map_err(|err| format!("failed to commit chunks on destDB: {err}"))
        });
        let mut state = hook.state.lock().unwrap_or_else(|p| p.into_inner());
        match result {
            Ok(()) => {
                state.error = None;
                state.last_pushed = root;
                state.pushed_read = Some(read);
                state.last_success = state.next_head_time;
                state.next_attempt = None;
                true
            }
            Err(err) => {
                state.error = Some(err);
                state.next_attempt = Some(Instant::now() + Duration::from_secs(1));
                destinations.remove(&key);
                false
            }
        }
    }

    /// open_destination opens a standby remote as a database whose requests carry this server's role and epoch.
    fn open_destination(&self, engine: &crate::engine::Engine, url: &str) -> std::result::Result<Database, String> {
        let store = remotes::client::RemoteStore::open_as(url, member(engine))?;
        Ok(Database::with_store(Box::new(store) as Box<dyn ChunkStore>))
    }

    /// record_received notes that a standby received a push for a database, which its status shows as the last
    /// update.
    pub fn record_received(&self, database: &str) {
        if self.role().0 != Role::Standby {
            return;
        }
        for hook in self.hooks().iter().filter(|h| h.database == database) {
            let mut state = hook.state.lock().unwrap_or_else(|p| p.into_inner());
            state.last_success = Some(SystemTime::now());
            state.error = None;
        }
    }

    /// ack_timeout returns how long a write waits for the standbys to receive it, from
    /// dolt_cluster_ack_writes_timeout_secs.
    pub fn ack_timeout(&self) -> i64 {
        self.ack_timeout.load(Ordering::Relaxed)
    }

    /// set_ack_timeout sets and persists dolt_cluster_ack_writes_timeout_secs.
    pub fn set_ack_timeout(&self, seconds: i64) -> Result<()> {
        self.ack_timeout.store(seconds, Ordering::Relaxed);
        write_global(&self.persist, &[(ACK_TIMEOUT_KEY, seconds.to_string())])
    }

    /// wait_replicated waits, for up to dolt_cluster_ack_writes_timeout_secs, until every standby holds what a
    /// database held when the wait began, as Dolt's WaitForReplicationController does, and returns how many standbys
    /// it timed out on with how many it waited for.
    pub fn wait_replicated(&self, database: &str) -> (usize, usize) {
        let timeout = self.ack_timeout();
        if timeout <= 0 || self.role().0 != Role::Primary {
            return (0, 0);
        }
        let start = Instant::now();
        let deadline = start + Duration::from_secs(timeout as u64);
        let hooks: Vec<Arc<Hook>> = self.hooks().into_iter().filter(|h| h.database == database).collect();
        loop {
            self.nudge();
            let behind = hooks
                .iter()
                .filter(|h| h.state.lock().unwrap_or_else(|p| p.into_inner()).pushed_read.is_none_or(|t| t < start))
                .count();
            if behind == 0 || Instant::now() >= deadline || self.role().0 != Role::Primary {
                return (behind, hooks.len());
            }
            std::thread::sleep(Duration::from_millis(5));
        }
    }

    /// nudge wakes the background replication.
    fn nudge(&self) {
        *self.nudged.0.lock().unwrap_or_else(|p| p.into_inner()) = true;
        self.nudged.1.notify_all();
    }

    /// sync_replicas sends each standby the roles and privileges, the branch control tables, and the dropped
    /// databases that it has not received yet, retrying a failed standby at once when forced, and returns whether each
    /// standby has them.
    fn sync_replicas<'a>(
        &'a self,
        engine: &crate::engine::Engine,
        clients: &mut HashMap<String, remotes::client::Replica>,
        forced: bool,
    ) -> Vec<(&'a StandbyRemote, bool)> {
        if self.role().0 != Role::Primary {
            return Vec::new();
        }
        let Some(member) = member(engine) else { return Vec::new() };
        let mut synced = Vec::new();
        let users = engine.auth_contents();
        let branch_control = engine.branch_control().lock().map(|c| c.serialize()).unwrap_or_default();
        for remote in &self.config.standby_remotes {
            let mut replicas = self.replicas.lock().unwrap_or_else(|p| p.into_inner());
            let state = replicas.entry(remote.name.clone()).or_default();
            if state.users == users && state.branch_control == branch_control && state.drops.is_empty() {
                synced.push((remote, true));
                continue;
            }
            if !forced && state.next_attempt.is_some_and(|t| Instant::now() < t) {
                synced.push((remote, false));
                continue;
            }
            let drops = state.drops.clone();
            let (sent_users, sent_branch_control) = (state.users != users, state.branch_control != branch_control);
            drop(replicas);
            let client = match clients.get(&remote.name) {
                Some(client) => Ok(client),
                None => {
                    remotes::client::Replica::connect(&remote.url_template.replace("{database}", ""), member.clone())
                        .map(|client| &*clients.entry(remote.name.clone()).or_insert(client))
                }
            };
            let result = client.and_then(|client| {
                if sent_users {
                    client.update_users(users.clone())?;
                }
                if sent_branch_control {
                    client.update_branch_control(branch_control.clone())?;
                }
                drops.iter().try_for_each(|name| client.drop_database(name))
            });
            let mut replicas = self.replicas.lock().unwrap_or_else(|p| p.into_inner());
            let state = replicas.entry(remote.name.clone()).or_default();
            synced.push((remote, result.is_ok()));
            match result {
                Ok(()) => {
                    state.users = users.clone();
                    state.branch_control = branch_control.clone();
                    state.drops.retain(|name| !drops.contains(name));
                    state.next_attempt = None;
                }
                Err(_) => {
                    state.next_attempt = Some(Instant::now() + Duration::from_secs(1));
                    clients.remove(&remote.name);
                }
            }
        }
        synced
    }

    /// run replicates the databases in the background for as long as the server runs.
    pub fn run(self: Arc<Self>, engine: crate::engine::Engine) {
        std::thread::spawn(move || {
            let mut destinations = HashMap::new();
            let mut clients = HashMap::new();
            loop {
                for hook in self.hooks() {
                    self.replicate(&engine, &hook, &mut destinations, false, None);
                }
                self.sync_replicas(&engine, &mut clients, false);
                let nudged = self.nudged.0.lock().unwrap_or_else(|p| p.into_inner());
                let (mut nudged, _) = self
                    .nudged
                    .1
                    .wait_timeout_while(nudged, Duration::from_millis(100), |n| !*n)
                    .unwrap_or_else(|p| p.into_inner());
                *nudged = false;
            }
        });
    }
}

/// Member is the server as a cluster member, as its remotes API server and clients see it.
pub struct Member {
    cluster: Arc<Cluster>,
    engine: crate::engine::Engine,
}

/// member returns the server as a cluster member, when it has cluster replication.
pub fn member(engine: &crate::engine::Engine) -> Option<Arc<dyn remotes::cluster::Member>> {
    let cluster = engine.cluster()?;
    Some(Arc::new(Member { cluster, engine: engine.clone() }))
}

/// databases returns the server as a cluster member that serves its databases to its primary.
pub fn databases(engine: &crate::engine::Engine) -> Option<Arc<dyn remotes::server::Databases>> {
    let cluster = engine.cluster()?;
    Some(Arc::new(Member { cluster, engine: engine.clone() }))
}

impl remotes::cluster::Member for Member {
    fn role(&self) -> (String, i64) {
        let (role, epoch) = self.cluster.role();
        (role.name().to_string(), epoch)
    }

    fn force_role(&self, role: &str, epoch: i64) {
        let _ = self.cluster.set_role(&self.engine, role, epoch, false, 0, None);
    }

    fn credentials(&self) -> &remotes::cluster::Credentials {
        &self.cluster.credentials
    }

    fn keys(&self) -> &remotes::cluster::KeySet {
        &self.cluster.keys
    }

    fn update_users(&self, contents: &[u8]) -> std::result::Result<(), String> {
        self.engine.replace_auth(contents).map_err(|err| err.message)
    }

    fn update_branch_control(&self, contents: &[u8]) -> std::result::Result<(), String> {
        let mut controller = self.engine.branch_control().lock().map_err(|_| "a lock was poisoned".to_string())?;
        controller.replace(contents).map_err(|err| err.message)
    }

    fn drop_database(&self, name: &str) -> std::result::Result<(), String> {
        if !self.engine.database_exists(name) {
            return Ok(());
        }
        self.engine.drop_database(name).map_err(|err| err.message)?;
        self.cluster.remove_hooks(name);
        Ok(())
    }
}

impl remotes::server::Databases for Member {
    fn database(&self, name: &str) -> Option<Arc<doltdb::handle::Handle>> {
        if !self.engine.database_exists(name) && name != DATABASE {
            self.engine.create_database(name, "postgres", "localhost").ok()?;
            self.engine.add_cluster_database(&self.cluster, name).ok()?;
        }
        self.engine.database_handle(name)
    }

    fn committed(&self, name: &str) {
        self.cluster.record_received(name);
    }
}

/// dolt_assume_cluster_role changes the server's role and epoch, as Dolt's procedure of that name does, replicating
/// to the standbys first when it becomes a standby.
pub fn dolt_assume_cluster_role(
    ctx: &mut crate::query::Ctx<'_>,
    args: &[crate::types::Value],
) -> Result<crate::types::Value> {
    let args = crate::dolt::procedures::strings(args);
    let (Some(role), Some(epoch)) = (args.first(), args.get(1)) else {
        return Err(PgError::new(code::INTERNAL_ERROR, "dolt_assume_cluster_role takes a role and an epoch"));
    };
    if role == Role::DetectedBrokenConfig.name() {
        return Err(PgError::new(
            code::INTERNAL_ERROR,
            "cannot set role to detected_broken_config; valid values are 'primary' and 'standby'",
        ));
    }
    let epoch: i64 = epoch
        .parse()
        .map_err(|_| PgError::new(code::INVALID_TEXT_REPRESENTATION, format!("invalid epoch: {epoch}")))?;
    let cluster = cluster_of(ctx)?;
    cluster.set_role(&ctx.session.engine, role, epoch, true, 0, Some((&ctx.txn.database, &*ctx.db)))?;
    Ok(crate::types::Value::Int8(0))
}

/// dolt_cluster_transition_to_standby gracefully makes a primary a standby at an epoch once enough standbys caught
/// up, returning each database's replication to each standby.
pub fn dolt_cluster_transition_to_standby(
    ctx: &mut crate::query::Ctx<'_>,
    args: &[crate::types::Value],
) -> Result<crate::types::Value> {
    let args = crate::dolt::procedures::strings(args);
    let number = |i: usize| -> Result<i64> {
        let text = args.get(i).cloned().unwrap_or_default();
        text.parse().map_err(|_| PgError::new(code::INVALID_TEXT_REPRESENTATION, format!("invalid argument: {text}")))
    };
    let (epoch, min_caught_up) = (number(0)?, number(1)?);
    let cluster = cluster_of(ctx)?;
    let held = Some((ctx.txn.database.as_str(), &*ctx.db));
    let (changed, results) =
        cluster.set_role(&ctx.session.engine, "standby", epoch, true, min_caught_up.max(0) as usize, held)?;
    if !changed {
        return Err(PgError::new(
            code::INTERNAL_ERROR,
            "failed to transition server to standby; it is already standby.",
        ));
    }
    let rows = results
        .into_iter()
        .map(|result| {
            crate::types::Value::Composite(Box::new(crate::types::CompositeValue {
                type_oid: crate::dolt::procedures::RECORD,
                fields: vec![
                    crate::types::Value::Int2(i16::from(result.caught_up)),
                    crate::types::Value::Text(result.database),
                    crate::types::Value::Text(result.remote),
                    crate::types::Value::Text(result.url),
                ],
            }))
        })
        .collect();
    Ok(crate::types::Value::Set(rows))
}

/// cluster_of returns the server's cluster replication, failing on a server without it.
fn cluster_of(ctx: &crate::query::Ctx<'_>) -> Result<Arc<Cluster>> {
    ctx.session
        .engine
        .cluster()
        .ok_or_else(|| PgError::new(code::INTERNAL_ERROR, "cluster replication is not configured on this server"))
}
