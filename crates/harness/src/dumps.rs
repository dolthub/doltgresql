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

use std::io::{Read, Write};
use std::net::{Shutdown, TcpListener, TcpStream};
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};

use crate::pgx::{Conn, ConnConfig};
use crate::server::{Server, Target};

/// TIMEOUT_ENV names the number of seconds an import may take, three minutes by default.
pub const TIMEOUT_ENV: &str = "DOLTGRES_DUMP_TIMEOUT";

/// PSQL_ENV names the psql binary to import with, which is otherwise found like the Go test finds it.
pub const PSQL_ENV: &str = "DOLTGRES_PSQL";

/// ImportTest imports a dump with psql and passes when the server returns no errors.
pub struct ImportTest {
    pub name: &'static str,
    pub set_up_script: &'static [&'static str],
    pub sql_filename: &'static str,
    pub skip_queries: &'static [&'static str],
}

/// ImportError is an error the server returned, along with the statement that caused it.
#[derive(Clone, Debug)]
pub struct ImportError {
    pub query: String,
    pub error: String,
}

/// dumps_dir returns the directory holding the dump files.
pub fn dumps_dir() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("../../testing/dumps/sql")
}

/// psql returns the psql binary, from the environment, the path, or pg_config's binary directory.
fn psql() -> Result<PathBuf, String> {
    if let Ok(path) = std::env::var(PSQL_ENV) {
        return Ok(PathBuf::from(path));
    }
    let works = |path: &Path| {
        Command::new(path)
            .arg("--version")
            .output()
            .is_ok_and(|o| String::from_utf8_lossy(&o.stdout).contains("PostgreSQL"))
    };
    if works(Path::new("psql")) {
        return Ok(PathBuf::from("psql"));
    }
    let output = Command::new("pg_config").arg("--bindir").output().map_err(|_| "Postgres is not installed")?;
    let path = Path::new(String::from_utf8_lossy(&output.stdout).trim()).join("psql");
    if works(&path) { Ok(path) } else { Err(format!("psql cannot be found at {}", path.display())) }
}

/// run_import_test imports the dump into a fresh server, panicking with the first errors when there are any.
pub fn run_import_test(test: &ImportTest) {
    if let Err(errors) = import(test) {
        panic!("{}", errors);
    }
}

/// import imports the dump, returning a description of the failure.
pub fn import(test: &ImportTest) -> Result<(), String> {
    let psql = psql()?;
    let server = Server::start(&Target::from_env()?, "")?;
    let mut conn = Conn::connect(
        ConnConfig::parse(&format!("postgres://postgres:password@127.0.0.1:{}/postgres", server.port))
            .map_err(|e| e.to_string())?,
    )
    .map_err(|e| e.to_string())?;
    for statement in test.set_up_script {
        conn.exec(statement, &[]).map_err(|e| format!("setup {statement}: {e}"))?;
    }
    let text = std::fs::read_to_string(dumps_dir().join(test.sql_filename)).map_err(|e| e.to_string())?;
    for role in dump_roles(&text) {
        let _ = conn.exec(&format!("CREATE ROLE \"{}\"", role.replace('"', "\"\"")), &[]);
    }
    conn.close();
    let errors = Arc::new(Mutex::new(Vec::new()));
    let proxy = Proxy::start(server.port, test.skip_queries, errors.clone())?;
    let dump = std::fs::File::open(dumps_dir().join(test.sql_filename)).map_err(|e| e.to_string())?;
    let mut child = Command::new(&psql)
        .arg(format!("postgresql://postgres:password@127.0.0.1:{}/postgres?sslmode=disable", proxy.port))
        .stdin(dump)
        .stdout(Stdio::null())
        .stderr(Stdio::piped())
        .spawn()
        .map_err(|e| format!("cannot run {}: {e}", psql.display()))?;
    let timeout = std::env::var(TIMEOUT_ENV).ok().and_then(|t| t.parse().ok()).unwrap_or(180);
    let deadline = Instant::now() + Duration::from_secs(timeout);
    let status = loop {
        if let Some(status) = child.try_wait().map_err(|e| e.to_string())? {
            break status;
        }
        if Instant::now() > deadline {
            let _ = child.kill();
            let _ = child.wait();
            return Err(format!("the import did not finish within {timeout} seconds"));
        }
        std::thread::sleep(Duration::from_millis(50));
    };
    proxy.stop();
    let errors = errors.lock().unwrap().clone();
    if !status.success() {
        let mut stderr = String::new();
        let _ = child.stderr.take().map(|mut s| s.read_to_string(&mut stderr));
        return Err(format!("psql exited with {status}: {stderr}"));
    }
    if errors.is_empty() {
        return Ok(());
    }
    let mut text = format!("COUNT: {}", errors.len());
    for error in errors.iter().take(10) {
        text.push_str(&format!("\nQUERY: {}\nERROR: {}", error.query, error.error));
    }
    Err(text)
}

/// dump_roles returns the roles that a dump names as owners, grantees, or session roles, which a restore into
/// Postgres creates beforehand from the cluster's globals.
fn dump_roles(text: &str) -> Vec<String> {
    let mut roles: Vec<String> = Vec::new();
    for line in text.lines() {
        let upper = line.to_ascii_uppercase();
        let mut starts: Vec<usize> = ["OWNER TO ", "AUTHORIZATION ", "SET ROLE "]
            .iter()
            .flat_map(|keyword| upper.match_indices(keyword).map(|(i, k)| i + k.len()))
            .collect();
        if (upper.starts_with("GRANT ") || upper.starts_with("REVOKE ")) && upper.contains(" TO ") {
            starts.push(upper.rfind(" TO ").map_or(0, |i| i + 4));
        }
        for start in starts {
            for name in line[start..].split(',') {
                let name = name.trim().trim_end_matches(';').trim_start_matches('\'');
                let name = match name.strip_prefix('"') {
                    Some(quoted) => quoted.split('"').next().unwrap_or_default(),
                    None => {
                        name.split(|c: char| !(c.is_alphanumeric() || c == '_' || c == '$')).next().unwrap_or_default()
                    }
                };
                let builtin = matches!(
                    name.to_ascii_lowercase().as_str(),
                    "" | "postgres" | "public" | "current_user" | "session_user" | "current_role" | "default"
                ) || name.starts_with("pg_");
                if !builtin && !roles.iter().any(|r| r == name) {
                    roles.push(name.to_string());
                }
            }
        }
    }
    roles
}

/// Proxy sits between psql and the server, recording every error with the statement that caused it.
struct Proxy {
    port: u16,
    stopped: Arc<AtomicBool>,
}

impl Proxy {
    fn start(
        server_port: u16,
        skip_queries: &'static [&'static str],
        errors: Arc<Mutex<Vec<ImportError>>>,
    ) -> Result<Proxy, String> {
        let listener = TcpListener::bind("127.0.0.1:0").map_err(|e| e.to_string())?;
        let port = listener.local_addr().map_err(|e| e.to_string())?.port();
        let stopped = Arc::new(AtomicBool::new(false));
        let stopping = stopped.clone();
        std::thread::spawn(move || {
            for client in listener.incoming() {
                if stopping.load(Ordering::SeqCst) {
                    return;
                }
                let Ok(client) = client else {
                    return;
                };
                let errors = errors.clone();
                std::thread::spawn(move || {
                    let _ = relay(client, server_port, skip_queries, errors);
                });
            }
        });
        Ok(Proxy { port, stopped })
    }

    fn stop(self) {
        self.stopped.store(true, Ordering::SeqCst);
        let _ = TcpStream::connect(("127.0.0.1", self.port));
    }
}

/// read_exact_or_none reads exactly the buffer's length, returning false at the end of the stream.
fn read_exact_or_none(stream: &mut TcpStream, buffer: &mut [u8]) -> bool {
    stream.read_exact(buffer).is_ok()
}

/// relay forwards one psql connection to the server, answering requests for encryption with a refusal.
fn relay(
    mut client: TcpStream,
    server_port: u16,
    skip_queries: &'static [&'static str],
    errors: Arc<Mutex<Vec<ImportError>>>,
) -> std::io::Result<()> {
    let mut server = TcpStream::connect(("127.0.0.1", server_port))?;
    loop {
        let mut length = [0u8; 4];
        if !read_exact_or_none(&mut client, &mut length) {
            return Ok(());
        }
        let mut body = vec![0u8; (u32::from_be_bytes(length) as usize).saturating_sub(4)];
        if !read_exact_or_none(&mut client, &mut body) {
            return Ok(());
        }
        let code = body.get(..4).map(|c| i32::from_be_bytes(c.try_into().unwrap()));
        if code == Some(pgproto::SSL_REQUEST_CODE) || code == Some(pgproto::GSSENC_REQUEST_CODE) {
            client.write_all(b"N")?;
            continue;
        }
        server.write_all(&length)?;
        server.write_all(&body)?;
        break;
    }
    let last_query = Arc::new(Mutex::new(String::new()));
    let (mut client_reader, mut server_writer) = (client.try_clone()?, server.try_clone()?);
    let queries = last_query.clone();
    let upstream = std::thread::spawn(move || -> std::io::Result<()> {
        loop {
            let mut header = [0u8; 5];
            if !read_exact_or_none(&mut client_reader, &mut header) {
                break;
            }
            let mut body = vec![0u8; (u32::from_be_bytes(header[1..].try_into().unwrap()) as usize).saturating_sub(4)];
            if !read_exact_or_none(&mut client_reader, &mut body) {
                break;
            }
            match header[0] {
                b'Q' => {
                    let query = String::from_utf8_lossy(body.strip_suffix(&[0]).unwrap_or(&body)).into_owned();
                    let mut last = queries.lock().unwrap();
                    if last.is_empty() {
                        *last = query.clone();
                    }
                    drop(last);
                    if skip_queries.iter().any(|skip| query.starts_with(skip)) {
                        body = b";\0".to_vec();
                        header[1..].copy_from_slice(&(body.len() as u32 + 4).to_be_bytes());
                    }
                }
                b'X' => break,
                _ => {}
            }
            server_writer.write_all(&header)?;
            server_writer.write_all(&body)?;
        }
        let _ = server_writer.shutdown(Shutdown::Both);
        Ok(())
    });
    loop {
        let mut header = [0u8; 5];
        if !read_exact_or_none(&mut server, &mut header) {
            break;
        }
        let mut body = vec![0u8; (u32::from_be_bytes(header[1..].try_into().unwrap()) as usize).saturating_sub(4)];
        if !read_exact_or_none(&mut server, &mut body) {
            break;
        }
        match header[0] {
            b'E' => {
                let message = pgproto::BackendMessage::decode(b'E', &body)
                    .ok()
                    .and_then(|m| match m {
                        pgproto::BackendMessage::ErrorResponse(fields) => Some(fields.message),
                        _ => None,
                    })
                    .unwrap_or_default();
                let last = last_query.lock().unwrap();
                let query = if last.is_empty() { "UNKNOWN QUERY HAS ERRORED".to_string() } else { last.clone() };
                errors.lock().unwrap().push(ImportError { query, error: message });
            }
            b'Z' => last_query.lock().unwrap().clear(),
            _ => {}
        }
        if client.write_all(&header).and_then(|_| client.write_all(&body)).is_err() {
            break;
        }
    }
    let _ = client.shutdown(Shutdown::Both);
    let _ = upstream.join();
    Ok(())
}
