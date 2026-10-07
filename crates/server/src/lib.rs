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

#![forbid(unsafe_code)]

//! The Doltgres server: it listens for Postgres clients and serves each connection on its own thread.

pub mod config;
mod conn;
pub mod scram;

use std::io::Write;
use std::net::TcpListener;
use std::sync::atomic::{AtomicU32, Ordering};
use std::sync::{Arc, Mutex, OnceLock};

use sql::Engine;

pub use config::Config;
use scram::Verifier;

/// Server is what every connection shares: the engine and the numbers that name connections in its log.
pub struct Server {
    pub engine: Engine,
    next_process_id: AtomicU32,
    /// The TLS configuration that encrypts connections, when the listener has a certificate and key.
    tls: Option<Arc<rustls::ServerConfig>>,
}

impl Server {
    /// new opens the configured data directory and auth file.
    pub fn new(config: &Config) -> Result<Server, String> {
        let engine = Engine::open(
            &config.data_dir,
            &config.default_database,
            &config.user,
            &config.password,
            &config.auth_file,
            config.branch_control_file.as_deref(),
        )
        .map_err(|err| err.to_string())?;
        engine.set_port(config.port);
        engine.set_behavior(
            config.read_only,
            config.auto_gc,
            config.auto_gc_archive,
            config.auto_gc_incremental_file_size,
        );
        let tls = match (&config.tls_cert, &config.tls_key) {
            (Some(cert), Some(key)) => Some(Arc::new(tls_config(cert, key)?)),
            _ => None,
        };
        Ok(Server { engine, next_process_id: AtomicU32::new(1), tls })
    }

    /// verifier returns the password verifier of a role with a password.
    fn verifier(&self, user: &str) -> Option<Verifier> {
        let (password, _) = self.engine.login(user)?;
        let password = password?;
        Some(Verifier {
            salt: password.salt,
            iterations: password.iterations,
            stored_key: password.stored_key.try_into().ok()?,
            server_key: password.server_key.try_into().ok()?,
        })
    }
}

/// tls_config reads a certificate chain and private key in PEM files into a TLS configuration.
fn tls_config(cert: &std::path::Path, key: &std::path::Path) -> Result<rustls::ServerConfig, String> {
    use rustls::pki_types::pem::PemObject;
    use rustls::pki_types::{CertificateDer, PrivateKeyDer};
    let chain = CertificateDer::pem_file_iter(cert)
        .and_then(|certs| certs.collect::<Result<Vec<_>, _>>())
        .map_err(|err| format!("cannot read {}: {err}", cert.display()))?;
    let key = PrivateKeyDer::from_pem_file(key).map_err(|err| format!("cannot read {}: {err}", key.display()))?;
    rustls::ServerConfig::builder()
        .with_no_client_auth()
        .with_single_cert(chain, key)
        .map_err(|err| format!("invalid TLS certificate or key: {err}"))
}

/// LOG is the file the server writes its log to, when the command line names one.
static LOG: OnceLock<Mutex<std::fs::File>> = OnceLock::new();

/// log writes a line to the server's log, which is standard error unless the command line names a file.
fn log(line: &str) {
    match LOG.get() {
        Some(file) => {
            let mut file = file.lock().unwrap_or_else(|poisoned| poisoned.into_inner());
            let _ = writeln!(file, "{line}");
        }
        None => eprintln!("{line}"),
    }
}

/// CONNECTION_STACK_SIZE is the stack size of each connection's thread, where function calls nest.
const CONNECTION_STACK_SIZE: usize = 256 << 20;

/// BIND_ATTEMPTS is how many times the server tries to listen on an address that is in use, a hundredth of a second
/// apart, which gives a server stopped just before it time to let the address go.
const BIND_ATTEMPTS: usize = 500;

/// bind listens on the address, waiting for one that is in use to come free.
fn bind(host: &str, port: u16) -> Result<TcpListener, String> {
    let mut attempts = 0;
    loop {
        match TcpListener::bind((host, port)) {
            Ok(listener) => return Ok(listener),
            Err(err) if err.kind() == std::io::ErrorKind::AddrInUse && attempts + 1 < BIND_ATTEMPTS => {
                attempts += 1;
                std::thread::sleep(std::time::Duration::from_millis(10));
            }
            Err(err) => return Err(format!("cannot listen on {host}:{port}: {err}")),
        }
    }
}

/// serve accepts connections on the configured address until the listener fails.
pub fn serve(config: &Config) -> Result<(), String> {
    if let Some(path) = &config.log_file {
        let file = std::fs::File::create(path).map_err(|err| format!("cannot open {}: {err}", path.display()))?;
        let _ = LOG.set(Mutex::new(file));
    }
    std::fs::create_dir_all(&config.data_dir)
        .map_err(|err| format!("failed to make dir '{}': {err}", config.data_dir.display()))?;
    if !config.skip_integrity_check
        && let Some(message) = sql::integrity::check_data_dir(&config.data_dir).map_err(|err| err.to_string())?
    {
        return Err(message);
    }
    let _ = doltdb::database::LOGGER.set(log);
    let server = Arc::new(Server::new(config)?);
    if config.auto_gc {
        let engine = server.engine.clone();
        std::thread::spawn(move || {
            loop {
                std::thread::sleep(std::time::Duration::from_secs(1));
                match engine.auto_gc() {
                    Ok(collected) => {
                        for (database, took) in collected {
                            log(&format!(
                                "sqle/auto_gc: Successfully completed auto GC of database {database} in {took:?}"
                            ));
                        }
                    }
                    Err(err) => log(&format!("sqle/auto_gc: {err}")),
                }
            }
        });
    }
    let engine = server.engine.clone();
    let mut signals = signal_hook::iterator::Signals::new([signal_hook::consts::SIGINT, signal_hook::consts::SIGTERM])
        .map_err(|err| format!("cannot handle signals: {err}"))?;
    std::thread::spawn(move || {
        if signals.forever().next().is_some() {
            if let Err(err) = engine.sync() {
                log(&format!("error writing out databases on shutdown: {err}"));
            }
            std::process::exit(0);
        }
    });
    let host = if config.host == "localhost" { "127.0.0.1" } else { config.host.as_str() };
    let listener = bind(host, config.port)?;
    log(&format!("Server ready. Accepting connections on {host}:{}.", config.port));
    for stream in listener.incoming() {
        let stream = stream.map_err(|err| err.to_string())?;
        let _ = stream.set_nodelay(true);
        let server = server.clone();
        let process_id = server.next_process_id.fetch_add(1, Ordering::Relaxed);
        let spawned = std::thread::Builder::new().stack_size(CONNECTION_STACK_SIZE).spawn(move || {
            if let Err(err) = conn::Conn::new(stream, server).run()
                && !matches!(&err, conn::ConnError::Io(e) if e.kind() == std::io::ErrorKind::UnexpectedEof)
            {
                log(&format!("connection {process_id} ended: {err}"));
            }
        });
        if let Err(err) = spawned {
            log(&format!("connection {process_id} could not start: {err}"));
        }
    }
    Ok(())
}
