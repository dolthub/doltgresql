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

//! Starts the server that a test runs against: a fresh doltgres process, or a fresh Postgres cluster copied from a
//! template, each on its own free port.

use std::net::{TcpListener, TcpStream};
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Stdio};
use std::sync::atomic::{AtomicU64, Ordering};
use std::time::{Duration, Instant};

/// The environment variable that selects the target server.
pub const TARGET_ENV: &str = "DOLTGRES_TEST_TARGET";
/// How long a server may take to accept connections.
const STARTUP_TIMEOUT: Duration = Duration::from_secs(60);

/// Target is the kind of server that tests run against.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Target {
    /// A doltgres binary.
    Doltgres {
        /// The path of the binary.
        binary: PathBuf,
    },
    /// A Postgres installation, whose clusters are copied from an initialized template data directory.
    Postgres {
        /// The directory holding the postgres and pg_ctl binaries.
        bin_dir: PathBuf,
        /// The template data directory.
        template: PathBuf,
    },
}

impl Target {
    /// from_env reads the target from DOLTGRES_TEST_TARGET, which is either "doltgres:<binary>" or
    /// "postgres:<bin dir>:<template data dir>".
    pub fn from_env() -> Result<Target, String> {
        let value = std::env::var(TARGET_ENV).map_err(|_| format!("{TARGET_ENV} is not set"))?;
        Target::parse(&value)
    }

    /// parse parses a target description, making its paths absolute since servers run in their own directories.
    pub fn parse(value: &str) -> Result<Target, String> {
        let absolute =
            |path: &str| std::fs::canonicalize(path).map_err(|err| format!("invalid {TARGET_ENV} path {path}: {err}"));
        if let Some(binary) = value.strip_prefix("doltgres:") {
            return Ok(Target::Doltgres { binary: absolute(binary)? });
        }
        if let Some(rest) = value.strip_prefix("postgres:")
            && let Some((bin_dir, template)) = rest.split_once(':')
        {
            return Ok(Target::Postgres { bin_dir: absolute(bin_dir)?, template: absolute(template)? });
        }
        Err(format!("invalid {TARGET_ENV}: {value}"))
    }

    /// is_postgres reports whether the target is a real Postgres.
    pub fn is_postgres(&self) -> bool {
        matches!(self, Target::Postgres { .. })
    }
}

/// Server is a running server, which is stopped and deleted when dropped.
pub struct Server {
    /// The port the server listens on at 127.0.0.1.
    pub port: u16,
    child: Child,
    /// The command that starts the server's process.
    command: Command,
    directory: PathBuf,
    stop: Option<Command>,
    pg_isready: Option<PathBuf>,
}

impl Server {
    /// start starts a fresh server of the target kind. The extra configuration is YAML appended to a doltgres
    /// config file, and is ignored for Postgres.
    pub fn start(target: &Target, extra_config: &str) -> Result<Server, String> {
        Server::start_with(target, extra_config, &[])
    }

    /// start_tls starts a fresh server of the target kind that also accepts TLS, using the PEM certificate chain and
    /// private key files.
    pub fn start_tls(target: &Target, cert: &Path, key: &Path) -> Result<Server, String> {
        let config = format!("  tls_cert: {}\n  tls_key: {}\n", cert.display(), key.display());
        let settings = [
            "ssl=on".to_string(),
            format!("ssl_cert_file={}", cert.display()),
            format!("ssl_key_file={}", key.display()),
        ];
        Server::start_with(target, &config, &settings)
    }

    /// start_with starts a fresh server with YAML appended to a doltgres config file, or with these settings for
    /// Postgres.
    fn start_with(target: &Target, extra_config: &str, postgres_settings: &[String]) -> Result<Server, String> {
        let directory = fresh_directory()?;
        let port = free_port()?;
        let result = match target {
            Target::Doltgres { binary } => start_doltgres(binary, &directory, port, extra_config),
            Target::Postgres { bin_dir, template } => {
                start_postgres(bin_dir, template, &directory, port, postgres_settings)
            }
        };
        let started = result.and_then(|(mut command, stop)| {
            let child = command
                .spawn()
                .map_err(|err| format!("cannot start {}: {err}", command.get_program().to_string_lossy()))?;
            Ok((child, command, stop))
        });
        let (child, command, stop) = match started {
            Ok(started) => started,
            Err(err) => {
                let _ = std::fs::remove_dir_all(&directory);
                return Err(err);
            }
        };
        let pg_isready = match target {
            Target::Postgres { bin_dir, .. } => Some(bin_dir.join("pg_isready")),
            Target::Doltgres { .. } => None,
        };
        let mut server = Server { port, child, command, directory, stop, pg_isready };
        server.wait_until_listening()?;
        Ok(server)
    }

    /// wait_until_listening waits until the server accepts TCP connections, and for Postgres until it has also
    /// finished starting up.
    fn wait_until_listening(&mut self) -> Result<(), String> {
        let deadline = Instant::now() + STARTUP_TIMEOUT;
        loop {
            if TcpStream::connect(("127.0.0.1", self.port)).is_ok() && self.is_ready() {
                return Ok(());
            }
            if let Ok(Some(status)) = self.child.try_wait() {
                return Err(format!("server exited during startup with {status}: {}", self.log_tail()));
            }
            if Instant::now() > deadline {
                return Err(format!("server did not start listening: {}", self.log_tail()));
            }
            std::thread::sleep(Duration::from_millis(10));
        }
    }

    /// is_ready reports whether pg_isready finds the server accepting connections, which is always true for doltgres.
    fn is_ready(&self) -> bool {
        let Some(pg_isready) = &self.pg_isready else { return true };
        Command::new(pg_isready)
            .args(["-q", "-h", "127.0.0.1", "-p", &self.port.to_string()])
            .status()
            .is_ok_and(|status| status.success())
    }

    /// log_tail returns the end of the server's log.
    pub fn log_tail(&self) -> String {
        let log = std::fs::read_to_string(self.directory.join("server.log")).unwrap_or_default();
        let start = log.len().saturating_sub(2000);
        log[log.floor_char_boundary(start)..].to_string()
    }

    /// directory returns the server's working directory, which holds its data and log.
    pub fn directory(&self) -> &Path {
        &self.directory
    }

    /// is_running reports whether the server's process has not exited.
    pub fn is_running(&mut self) -> bool {
        matches!(self.child.try_wait(), Ok(None))
    }

    /// restart stops the server if it still runs and starts it again in its directory, keeping its data.
    pub fn restart(&mut self) -> Result<(), String> {
        self.halt();
        self.child = self.command.spawn().map_err(|err| format!("cannot restart the server: {err}"))?;
        self.wait_until_listening()
    }

    /// halt stops the server's process.
    fn halt(&mut self) {
        match self.stop.as_mut() {
            Some(stop) => {
                let _ = stop.stdout(Stdio::null()).stderr(Stdio::null()).status();
                let _ = self.child.wait();
            }
            None => {
                let _ = self.child.kill();
                let _ = self.child.wait();
            }
        }
    }
}

impl Drop for Server {
    /// drop implements the interface Drop by stopping the server and deleting its directory.
    fn drop(&mut self) {
        self.halt();
        let _ = std::fs::remove_dir_all(&self.directory);
    }
}

/// start_doltgres returns the command that starts a doltgres binary with a config file that listens on the port. The testdata directory is
/// copied into its working directory, so that relative file paths in statements resolve.
fn start_doltgres(
    binary: &Path,
    directory: &Path,
    port: u16,
    extra_config: &str,
) -> Result<(Command, Option<Command>), String> {
    let config = format!("log_level: warn\n\nlistener:\n  host: 127.0.0.1\n  port: {port}\n{extra_config}");
    let config_path = directory.join("config.yaml");
    std::fs::write(&config_path, config).map_err(|err| err.to_string())?;
    let data_dir = directory.join("data");
    std::fs::create_dir_all(&data_dir).map_err(|err| err.to_string())?;
    let testdata = crate::script::testdata_dir();
    if testdata.is_dir() {
        copy_dir(&testdata, &directory.join("testdata")).map_err(|err| format!("cannot copy testdata: {err}"))?;
    }
    let log = std::fs::File::create(directory.join("server.log")).map_err(|err| err.to_string())?;
    let mut command = Command::new(binary);
    command
        .arg("--config")
        .arg(&config_path)
        .arg("--data-dir")
        .arg(&data_dir)
        .current_dir(directory)
        .env_remove("DOLTGRES_DATA_DIR")
        .stdout(log.try_clone().map_err(|err| err.to_string())?)
        .stderr(log);
    Ok((command, None))
}

/// start_postgres copies the template into a new data directory and returns the commands that start a postmaster on
/// the port with the extra settings and stop it.
fn start_postgres(
    bin_dir: &Path,
    template: &Path,
    directory: &Path,
    port: u16,
    settings: &[String],
) -> Result<(Command, Option<Command>), String> {
    let data_dir = directory.join("data");
    copy_dir(template, &data_dir).map_err(|err| format!("cannot copy {}: {err}", template.display()))?;
    let log = std::fs::File::create(directory.join("server.log")).map_err(|err| err.to_string())?;
    let mut command = Command::new(bin_dir.join("postgres"));
    for setting in settings {
        command.arg("-c").arg(setting);
    }
    command
        .arg("-D")
        .arg(&data_dir)
        .arg("-p")
        .arg(port.to_string())
        .arg("-c")
        .arg("listen_addresses=127.0.0.1")
        .arg("-c")
        .arg(format!("unix_socket_directories={}", directory.display()))
        .arg("-c")
        .arg("fsync=off")
        .env("LC_ALL", "en_US.UTF-8")
        .stdout(log.try_clone().map_err(|err| err.to_string())?)
        .stderr(log);
    let mut stop = Command::new(bin_dir.join("pg_ctl"));
    stop.arg("stop").arg("-D").arg(&data_dir).arg("-m").arg("immediate").arg("-w");
    Ok((command, Some(stop)))
}

/// copy_dir copies a directory tree, keeping file permissions.
fn copy_dir(from: &Path, to: &Path) -> std::io::Result<()> {
    std::fs::create_dir_all(to)?;
    std::fs::set_permissions(to, std::fs::metadata(from)?.permissions())?;
    for entry in std::fs::read_dir(from)? {
        let entry = entry?;
        let destination = to.join(entry.file_name());
        if entry.file_type()?.is_dir() {
            copy_dir(&entry.path(), &destination)?;
        } else {
            std::fs::copy(entry.path(), destination)?;
        }
    }
    Ok(())
}

/// free_port returns a port that was free when it was checked.
fn free_port() -> Result<u16, String> {
    let listener = TcpListener::bind("127.0.0.1:0").map_err(|err| err.to_string())?;
    Ok(listener.local_addr().map_err(|err| err.to_string())?.port())
}

/// fresh_directory creates a new empty directory for a server. Its path is short, since Postgres limits the length
/// of its socket path.
fn fresh_directory() -> Result<PathBuf, String> {
    static COUNTER: AtomicU64 = AtomicU64::new(0);
    let base = std::env::temp_dir().join("dgt");
    std::fs::create_dir_all(&base).map_err(|err| err.to_string())?;
    let name = format!("{}-{}", std::process::id(), COUNTER.fetch_add(1, Ordering::Relaxed));
    let directory = base.join(name);
    let _ = std::fs::remove_dir_all(&directory);
    std::fs::create_dir_all(&directory).map_err(|err| err.to_string())?;
    Ok(directory)
}
