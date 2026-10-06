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

use std::collections::HashMap;
use std::net::TcpListener;
use std::sync::Arc;
use std::sync::atomic::{AtomicU32, Ordering};

use sql::Engine;

pub use config::Config;
use scram::Verifier;

/// Server is what every connection shares: the engine and the users' password verifiers.
pub struct Server {
    pub engine: Engine,
    users: HashMap<String, Verifier>,
    next_process_id: AtomicU32,
}

impl Server {
    /// new opens the configured data directory and sets up the superuser.
    pub fn new(config: &Config) -> Result<Server, String> {
        let engine = Engine::open(&config.data_dir, &config.user).map_err(|err| err.to_string())?;
        let users = HashMap::from([(config.user.clone(), Verifier::new(&config.password))]);
        Ok(Server { engine, users, next_process_id: AtomicU32::new(1) })
    }

    /// verifier returns the user's password verifier.
    fn verifier(&self, user: &str) -> Option<Verifier> {
        self.users.get(user).cloned()
    }
}

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
        std::thread::spawn(move || {
            if let Err(err) = conn::Conn::new(stream, server, process_id).run()
                && !matches!(&err, conn::ConnError::Io(e) if e.kind() == std::io::ErrorKind::UnexpectedEof)
            {
                eprintln!("connection {process_id} ended: {err}");
            }
        });
    }
    Ok(())
}
