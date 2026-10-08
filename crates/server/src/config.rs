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

//! The server's configuration: the command line, the YAML config file, and the environment.

use std::path::PathBuf;

use yaml_rust2::{Yaml, YamlLoader};

/// Config is the server's configuration.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Config {
    pub log_level: String,
    pub host: String,
    pub port: u16,
    pub tls_cert: Option<PathBuf>,
    pub tls_key: Option<PathBuf>,
    pub data_dir: PathBuf,
    /// The superuser's name and password, from DOLTGRES_USER and DOLTGRES_PASSWORD.
    pub user: String,
    pub password: String,
    /// The database a first start creates, from DOLTGRES_DB or else the superuser's name.
    pub default_database: String,
    /// The file that holds the roles and privileges, relative to the working directory.
    pub auth_file: PathBuf,
    /// The file that holds the branch control tables, relative to the working directory, or None to keep them in
    /// memory as the Go server does without one.
    pub branch_control_file: Option<PathBuf>,
    /// The file that `-stdout`, `-stderr`, or `-out-and-err` sends the server's log to, or None for standard error.
    pub log_file: Option<PathBuf>,
    /// Whether the server starts without checking its databases for values that earlier releases serialized
    /// incorrectly, from behavior.skip_startup_integrity_check.
    pub skip_integrity_check: bool,
    /// Whether the server collects garbage on its own as stores grow, from behavior.auto_gc_behavior.enable.
    pub auto_gc: bool,
    /// Whether the server refuses every write, from behavior.read_only.
    pub read_only: bool,
    /// The port that serves the remotes API, from remotesapi.port, with whether it refuses writes, from
    /// remotesapi.read_only.
    pub remotesapi: Option<(u16, bool)>,
    /// Whether automatic garbage collection writes archives, from behavior.auto_gc_behavior.archive_level.
    pub auto_gc_archive: bool,
    /// The size of automatic garbage collection's incremental files, or 0 for none, from
    /// behavior.auto_gc_behavior.incremental_file_size.
    pub auto_gc_incremental_file_size: u64,
    /// The cluster replication that the server takes part in, from the `cluster` section.
    pub cluster: Option<sql::cluster::ClusterConfig>,
    /// The Postgres primary that the server replicates from, from the `postgres_replication` section.
    pub postgres_replication: Option<ReplicationConfig>,
    /// The directory that holds the replication position file, from cfg_dir, relative to the working directory.
    pub cfg_dir: PathBuf,
}

/// ReplicationConfig is the `postgres_replication` section, where missing values are empty.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct ReplicationConfig {
    pub server_address: String,
    pub user: String,
    pub password: String,
    pub database: String,
    pub port: i64,
    pub slot_name: String,
}

/// Startup is what a command line asks for: serving with a configuration, or printing text and exiting.
pub enum Startup {
    Serve(Box<Config>),
    Print(String),
}

/// USAGE is the command line's help, as the Go server prints it.
const USAGE: &str = "Usage: doltgres [options]
Options:
  -config string
    \tPath to the config file.
    \tIf not provided, ./config.yaml will be used if it exists.
  -data-dir string
    \tPath to the directory where doltgres databases are stored.
    \tIf not provided, the value in config.yaml will be used. If that's not
    \tprovided either, the value of the DOLTGRES_DATA_DIR environment variable
    \twill be used if set. Otherwise $HOME/doltgres/databases will be used. The
    \tdirectory will be created if it doesn't exist.
  -config-help
    \tprint the config file help
  -version
    \tprint the version
  -chdir string
    \tset the working directory for doltgres
  -stdin string
    \tfile to use as stdin
  -stdout string
    \tfile to use as stdout
  -stderr string
    \tfile to use as stderr
  -out-and-err string
    \tfile to use as stdout and stderr
";

/// config_help returns the help for the config file, with the defaults it lists, as the Go server prints it.
fn config_help(data_dir: &std::path::Path) -> String {
    format!(
        "Supported fields in the config.yaml file, and their default values, are as follows:

log_level: info

encode_logged_query: false

behavior:
  read_only: false
  dolt_transaction_commit: false
  permit_unsupported_locking_statements: false

user:
  name: postgres
  password: password

listener:
  host: localhost
  port: 5432
  read_timeout_millis: 28800000
  write_timeout_millis: 28800000
  allow_cleartext_passwords: false

data_dir: {}

cfg_dir: .doltcfg

privilege_file: .doltcfg/privileges.db

auth_file: .doltcfg/auth.db

branch_control_file: .doltcfg/branch_control.db
",
        data_dir.display()
    )
}

impl Config {
    /// from_args reads what the command line arguments after the program name ask for, reading the configuration from
    /// the config file they name, or ./config.yaml, and the environment.
    pub fn from_args(args: impl IntoIterator<Item = String>) -> Result<Startup, String> {
        let mut args = args.into_iter();
        let (mut config_path, mut data_dir, mut log_file) = (None, None, None);
        let env = |name: &str| std::env::var(name).ok().filter(|v| !v.is_empty());
        let default_data_dir = || PathBuf::from(env("HOME").unwrap_or_default()).join("doltgres/databases");
        while let Some(arg) = args.next() {
            let (name, inline) = match arg.split_once('=') {
                Some((name, value)) => (name.to_string(), Some(value.to_string())),
                None => (arg.clone(), None),
            };
            let mut value = || inline.clone().or_else(|| args.next()).ok_or_else(|| format!("{name} needs a value"));
            match name.trim_start_matches('-') {
                "config" => config_path = Some(PathBuf::from(value()?)),
                "data-dir" => data_dir = Some(PathBuf::from(value()?)),
                "chdir" => {
                    let dir = value()?;
                    std::env::set_current_dir(&dir)
                        .map_err(|err| format!("cannot change directory to {dir}: {err}"))?;
                }
                "stdin" => {
                    value()?;
                }
                "stdout" | "stderr" | "out-and-err" => log_file = Some(PathBuf::from(value()?)),
                "version" => return Ok(Startup::Print(format!("Doltgres version {}\n", sql::DOLTGRES_VERSION))),
                "config-help" => return Ok(Startup::Print(config_help(&default_data_dir()))),
                "help" | "h" => return Ok(Startup::Print(USAGE.to_string())),
                _ => return Err(format!("flag provided but not defined: {arg}\n{USAGE}")),
            }
        }
        if config_path.is_none() && std::path::Path::new("config.yaml").is_file() {
            config_path = Some(PathBuf::from("config.yaml"));
        }
        let mut config = Config {
            log_level: "info".into(),
            host: "localhost".into(),
            port: 5432,
            tls_cert: None,
            tls_key: None,
            data_dir: data_dir
                .clone()
                .or_else(|| env("DOLTGRES_DATA_DIR").map(PathBuf::from))
                .unwrap_or_else(default_data_dir),
            user: env("DOLTGRES_USER").unwrap_or_else(|| "postgres".into()),
            password: env("DOLTGRES_PASSWORD").unwrap_or_else(|| "password".into()),
            default_database: env("DOLTGRES_DB").or_else(|| env("DOLTGRES_USER")).unwrap_or_else(|| "postgres".into()),
            auth_file: PathBuf::from("auth.db"),
            branch_control_file: None,
            log_file,
            skip_integrity_check: false,
            auto_gc: true,
            read_only: false,
            remotesapi: None,
            auto_gc_archive: true,
            auto_gc_incremental_file_size: 0,
            cluster: None,
            postgres_replication: None,
            cfg_dir: PathBuf::new(),
        };
        if let Some(path) = config_path {
            let text =
                std::fs::read_to_string(&path).map_err(|err| format!("cannot read {}: {err}", path.display()))?;
            config.apply_yaml(&text, data_dir.is_some())?;
        }
        Ok(Startup::Serve(Box::new(config)))
    }

    /// apply_yaml applies a YAML config file, whose data_dir yields to one given on the command line.
    fn apply_yaml(&mut self, text: &str, data_dir_given: bool) -> Result<(), String> {
        let docs = YamlLoader::load_from_str(text).map_err(|err| format!("invalid config file: {err}"))?;
        let Some(doc) = docs.first() else { return Ok(()) };
        if let Some(level) = doc["log_level"].as_str() {
            self.log_level = level.to_string();
        }
        if let Some(enable) = doc["behavior"]["auto_gc_behavior"]["enable"].as_bool() {
            self.auto_gc = enable;
        }
        if let Some(level) = doc["behavior"]["auto_gc_behavior"]["archive_level"].as_i64() {
            self.auto_gc_archive = level != 0;
        }
        if let Some(size) = doc["behavior"]["auto_gc_behavior"]["incremental_file_size"].as_i64() {
            self.auto_gc_incremental_file_size = size.max(0) as u64;
        }
        if let Some(port) = doc["remotesapi"]["port"].as_i64() {
            let read_only = doc["remotesapi"]["read_only"].as_bool().unwrap_or(false);
            self.remotesapi = Some((port as u16, read_only));
        }
        if !doc["cluster"].is_badvalue() {
            self.cluster = Some(cluster_config(&doc["cluster"])?);
        }
        let replication = &doc["postgres_replication"];
        if !replication.is_badvalue() {
            let text = |key: &str| replication[key].as_str().unwrap_or_default().to_string();
            self.postgres_replication = Some(ReplicationConfig {
                server_address: text("postgres_server_address"),
                user: text("postgres_user"),
                password: text("postgres_password"),
                database: text("postgres_database"),
                port: replication["postgres_port"].as_i64().unwrap_or(0),
                slot_name: text("slot_name"),
            });
        }
        if let Some(dir) = doc["cfg_dir"].as_str() {
            self.cfg_dir = PathBuf::from(dir);
        }
        if let Some(read_only) = doc["behavior"]["read_only"].as_bool() {
            self.read_only = read_only;
        }
        if let Some(skip) = doc["behavior"]["skip_startup_integrity_check"].as_bool() {
            self.skip_integrity_check = skip;
        }
        if let Some(file) = doc["auth_file"].as_str() {
            self.auth_file = PathBuf::from(file);
        }
        if let Some(file) = doc["branch_control_file"].as_str() {
            self.branch_control_file = Some(PathBuf::from(file));
        }
        if let (Some(dir), false) = (doc["data_dir"].as_str(), data_dir_given) {
            self.data_dir = PathBuf::from(dir);
        }
        let listener = &doc["listener"];
        if let Some(host) = listener["host"].as_str() {
            self.host = host.to_string();
        }
        match &listener["port"] {
            Yaml::Integer(port) => self.port = u16::try_from(*port).map_err(|_| format!("invalid port {port}"))?,
            Yaml::BadValue | Yaml::Null => {}
            other => return Err(format!("invalid port {other:?}")),
        }
        self.tls_cert = listener["tls_cert"].as_str().map(PathBuf::from);
        self.tls_key = listener["tls_key"].as_str().map(PathBuf::from);
        if listener["require_secure_transport"].as_bool() == Some(true)
            && (self.tls_key.is_none() || self.tls_cert.is_none())
        {
            return Err("require_secure_transport can only be `true` when a tls_key and tls_cert are provided.".into());
        }
        Ok(())
    }
}

/// cluster_config reads the `cluster` section of a config file, checking it as Dolt's ValidateClusterConfig does.
fn cluster_config(cluster: &Yaml) -> Result<sql::cluster::ClusterConfig, String> {
    let remotes: Vec<sql::cluster::StandbyRemote> = cluster["standby_remotes"]
        .as_vec()
        .into_iter()
        .flatten()
        .map(|remote| sql::cluster::StandbyRemote {
            name: remote["name"].as_str().unwrap_or_default().to_string(),
            url_template: remote["remote_url_template"].as_str().unwrap_or_default().to_string(),
        })
        .collect();
    if remotes.is_empty() {
        return Err("cluster config: must supply standby_remotes when supplying cluster configuration.".into());
    }
    for (i, remote) in remotes.iter().enumerate() {
        if remote.name.is_empty() {
            return Err(format!("cluster: standby_remotes[{i}]: name: Cannot be empty"));
        }
        if !remote.url_template.contains("{database}") {
            return Err(format!(
                "cluster: standby_remotes[{i}]: remote_url_template: is \"{}\" but must include the {{database}} \
                 template parameter",
                remote.url_template
            ));
        }
    }
    let bootstrap_role = cluster["bootstrap_role"].as_str().unwrap_or_default().to_string();
    if !matches!(bootstrap_role.as_str(), "" | "primary" | "standby") {
        return Err(format!("cluster: boostrap_role: is \"{bootstrap_role}\" but must be \"primary\" or \"standby\""));
    }
    let bootstrap_epoch = cluster["bootstrap_epoch"].as_i64().unwrap_or(0);
    if bootstrap_epoch < 0 {
        return Err(format!("cluster: boostrap_epoch: is {bootstrap_epoch} but must be >= 0"));
    }
    let port = cluster["remotesapi"]["port"].as_i64().unwrap_or(0);
    let remotesapi_port =
        u16::try_from(port).map_err(|_| format!("cluster: remotesapi: port: is not in range 0-65535: {port}"))?;
    Ok(sql::cluster::ClusterConfig { standby_remotes: remotes, bootstrap_role, bootstrap_epoch, remotesapi_port })
}
