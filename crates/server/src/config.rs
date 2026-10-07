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
    /// The file that holds the roles and privileges, relative to the working directory.
    pub auth_file: PathBuf,
    /// The file that holds the branch control tables, relative to the working directory, or None to keep them in
    /// memory as the Go server does without one.
    pub branch_control_file: Option<PathBuf>,
}

/// usage is the command line's help.
const USAGE: &str = "usage: doltgres [--config <file>] [--data-dir <dir>]";

impl Config {
    /// from_args reads the configuration from the command line arguments after the program name, the config file
    /// they name, and the environment.
    pub fn from_args(args: impl IntoIterator<Item = String>) -> Result<Config, String> {
        let mut args = args.into_iter();
        let (mut config_path, mut data_dir) = (None, None);
        while let Some(arg) = args.next() {
            let (name, inline) = match arg.split_once('=') {
                Some((name, value)) => (name.to_string(), Some(value.to_string())),
                None => (arg.clone(), None),
            };
            let mut value = || inline.clone().or_else(|| args.next()).ok_or_else(|| format!("{name} needs a value"));
            match name.trim_start_matches('-') {
                "config" => config_path = Some(PathBuf::from(value()?)),
                "data-dir" => data_dir = Some(PathBuf::from(value()?)),
                "help" | "h" => return Err(USAGE.to_string()),
                _ => return Err(format!("unknown argument {arg}\n{USAGE}")),
            }
        }
        let env = |name: &str| std::env::var(name).ok().filter(|v| !v.is_empty());
        let mut config = Config {
            log_level: "info".into(),
            host: "localhost".into(),
            port: 5432,
            tls_cert: None,
            tls_key: None,
            data_dir: data_dir
                .clone()
                .or_else(|| env("DOLTGRES_DATA_DIR").map(PathBuf::from))
                .unwrap_or_else(|| PathBuf::from(env("HOME").unwrap_or_default()).join("doltgres/databases")),
            user: env("DOLTGRES_USER").unwrap_or_else(|| "postgres".into()),
            password: env("DOLTGRES_PASSWORD").unwrap_or_else(|| "password".into()),
            auth_file: PathBuf::from("auth.db"),
            branch_control_file: None,
        };
        if let Some(path) = config_path {
            let text =
                std::fs::read_to_string(&path).map_err(|err| format!("cannot read {}: {err}", path.display()))?;
            config.apply_yaml(&text, data_dir.is_some())?;
        }
        Ok(config)
    }

    /// apply_yaml applies a YAML config file, whose data_dir yields to one given on the command line.
    fn apply_yaml(&mut self, text: &str, data_dir_given: bool) -> Result<(), String> {
        let docs = YamlLoader::load_from_str(text).map_err(|err| format!("invalid config file: {err}"))?;
        let Some(doc) = docs.first() else { return Ok(()) };
        if let Some(level) = doc["log_level"].as_str() {
            self.log_level = level.to_string();
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
        Ok(())
    }
}
