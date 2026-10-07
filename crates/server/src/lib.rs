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

use std::net::TcpListener;
use std::sync::Arc;
use std::sync::atomic::{AtomicU32, Ordering};

use sql::Engine;

pub use config::Config;
use scram::Verifier;

/// Server is what every connection shares: the engine and the numbers that name connections in its log.
pub struct Server {
    pub engine: Engine,
    next_process_id: AtomicU32,
}

impl Server {
    /// new opens the configured data directory and auth file.
    pub fn new(config: &Config) -> Result<Server, String> {
        let engine = Engine::open(
            &config.data_dir,
            &config.user,
            &config.password,
            &config.auth_file,
            config.branch_control_file.as_deref(),
        )
        .map_err(|err| err.to_string())?;
        engine.set_port(config.port);
        Ok(Server { engine, next_process_id: AtomicU32::new(1) })
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

/// CONNECTION_STACK_SIZE is the stack size of each connection's thread, where function calls nest.
const CONNECTION_STACK_SIZE: usize = 256 << 20;

/// serve accepts connections on the configured address until the listener fails.
pub fn serve(config: &Config) -> Result<(), String> {
    let server = Arc::new(Server::new(config)?);
    let host = if config.host == "localhost" { "127.0.0.1" } else { config.host.as_str() };
    let listener = TcpListener::bind((host, config.port))
        .map_err(|err| format!("cannot listen on {host}:{}: {err}", config.port))?;
    eprintln!("Server ready. Accepting connections on {host}:{}.", config.port);
    for stream in listener.incoming() {
        let stream = stream.map_err(|err| err.to_string())?;
        let _ = stream.set_nodelay(true);
        let server = server.clone();
        let process_id = server.next_process_id.fetch_add(1, Ordering::Relaxed);
        let spawned = std::thread::Builder::new().stack_size(CONNECTION_STACK_SIZE).spawn(move || {
            if let Err(err) = conn::Conn::new(stream, server).run()
                && !matches!(&err, conn::ConnError::Io(e) if e.kind() == std::io::ErrorKind::UnexpectedEof)
            {
                eprintln!("connection {process_id} ended: {err}");
            }
        });
        if let Err(err) = spawned {
            eprintln!("connection {process_id} could not start: {err}");
        }
    }
    Ok(())
}
