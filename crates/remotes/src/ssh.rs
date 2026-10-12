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

//! Ssh remotes: `dolt transfer` run over ssh on the remote host, serving the remotes API's gRPC and HTTP requests over
//! an smux session on its standard input and output, as Dolt's SSHRemoteFactory connects to it.

use std::io::Read;
use std::process::{Child, Command, Stdio};
use std::sync::{Arc, Mutex};
use std::thread::JoinHandle;
use std::time::{Duration, Instant};

use axum::body::Bytes;
use axum::http::{Request, StatusCode, Uri};
use http_body_util::{BodyExt, Full};
use hyper_util::rt::TokioIo;
use tonic::transport::{Channel, Endpoint};

use crate::smux::Session;

/// SSH_COMMAND_ENV names the command, with its arguments, that connects to the remote host, `ssh` by default.
const SSH_COMMAND_ENV: &str = "DOLT_SSH_COMMAND";

/// EXEC_PATH_ENV names the path of the dolt binary on the remote host, `dolt` by default.
const EXEC_PATH_ENV: &str = "DOLT_SSH_EXEC_PATH";

/// Ssh is a running `dolt transfer` and the session over its pipes.
pub(crate) struct Ssh {
    session: Session,
    child: Child,
    /// The remote's error output, which a reader thread collects until the process ends.
    stderr: Arc<Mutex<Vec<u8>>>,
    reader: Option<JoinHandle<()>>,
    /// The repository path on the remote host.
    pub(crate) path: String,
}

/// Address is what an ssh remote's URL names.
#[derive(Debug, PartialEq)]
struct Address {
    host: String,
    port: String,
    path: String,
    user: String,
}

/// parse reads an ssh remote's URL, dropping a trailing `/` or `/.dolt` from its path, as Dolt's parseSSHURL does.
fn parse(url: &str) -> Address {
    let rest = url.strip_prefix("ssh://").unwrap_or(url);
    let (authority, path) = rest.find('/').map_or((rest, ""), |at| rest.split_at(at));
    let path = path.strip_suffix('/').unwrap_or(path);
    let path = path.strip_suffix("/.dolt").unwrap_or(path).to_string();
    let (user, host_port) = match authority.rsplit_once('@') {
        Some((info, host_port)) => (info.split(':').next().unwrap_or(info).to_string(), host_port),
        None => (String::new(), authority),
    };
    let (host, port) = match host_port.strip_prefix('[').and_then(|h| h.split_once(']')) {
        Some((host, port)) => (host, port.strip_prefix(':').unwrap_or("")),
        None => match host_port.rsplit_once(':') {
            Some((host, port)) if port.bytes().all(|b| b.is_ascii_digit()) => (host, port),
            _ => (host_port, ""),
        },
    };
    Address { host: host.to_string(), port: port.to_string(), path, user }
}

/// command builds `ssh [-p port] [user@]host "<dolt> --data-dir <path> transfer"` from an ssh command, with its
/// arguments, and the remote dolt binary's path, as Dolt's buildTransferCommand does.
fn command(address: &Address, ssh: &str, dolt: &str) -> Result<Command, String> {
    let mut words = ssh.split_whitespace();
    let program = words.next().ok_or_else(|| format!("invalid {SSH_COMMAND_ENV}: empty"))?;
    let mut command = Command::new(program);
    command.args(words);
    if !address.port.is_empty() {
        command.args(["-p", &address.port]);
    }
    let target = match address.user.as_str() {
        "" => address.host.clone(),
        user => format!("{user}@{}", address.host),
    };
    command.arg(target).arg(format!("{dolt} --data-dir {} transfer", address.path));
    Ok(command)
}

impl Ssh {
    /// start runs `dolt transfer` for an ssh remote's URL and starts a session over its pipes on the current runtime.
    pub(crate) fn start(url: &str) -> Result<Ssh, String> {
        let address = parse(url);
        let ssh = std::env::var(SSH_COMMAND_ENV).ok().filter(|c| !c.is_empty()).unwrap_or_else(|| "ssh".into());
        let dolt = std::env::var(EXEC_PATH_ENV).ok().filter(|p| !p.is_empty()).unwrap_or_else(|| "dolt".into());
        let mut child = command(&address, &ssh, &dolt)?
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .map_err(|err| format!("failed to start transfer subprocess: {err}"))?;
        let (Some(stdin), Some(stdout), Some(mut errors)) =
            (child.stdin.take(), child.stdout.take(), child.stderr.take())
        else {
            return Err("failed to start transfer subprocess: missing pipes".into());
        };
        let stderr = Arc::new(Mutex::new(Vec::new()));
        let collected = stderr.clone();
        let reader = std::thread::spawn(move || {
            let mut buffer = [0; 4096];
            while let Ok(count @ 1..) = errors.read(&mut buffer) {
                collected.lock().unwrap_or_else(|e| e.into_inner()).extend_from_slice(&buffer[..count]);
            }
        });
        let session = match tokio::process::ChildStdin::from_std(stdin)
            .and_then(|stdin| tokio::process::ChildStdout::from_std(stdout).map(|stdout| Session::new(stdout, stdin)))
        {
            Ok(session) => session,
            Err(err) => {
                let _ = child.kill();
                let _ = child.wait();
                return Err(format!("failed to create SMUX client session: {err}"));
            }
        };
        Ok(Ssh { session, child, stderr, reader: Some(reader), path: address.path })
    }

    /// channel opens the gRPC channel, whose connection is one of the session's streams.
    pub(crate) async fn channel(&self) -> Result<Channel, tonic::transport::Error> {
        let session = self.session.clone();
        let connector = tower::service_fn(move |_: Uri| {
            let stream = session.open().map(TokioIo::new);
            async move { stream }
        });
        Endpoint::from_static("http://stdio").connect_with_connector(connector).await
    }

    /// send sends an HTTP request on a stream of its own, whatever host its URL names, and returns the response.
    pub(crate) async fn send(&self, request: Request<Full<Bytes>>) -> Result<(StatusCode, Bytes), String> {
        let (mut parts, body) = request.into_parts();
        if let Some(authority) = parts.uri.authority().and_then(|a| a.as_str().parse().ok()) {
            parts.headers.insert(axum::http::header::HOST, authority);
        }
        if let Some(origin) = parts.uri.path_and_query().map(|path| path.as_str().parse::<Uri>()) {
            parts.uri = origin.map_err(|e| e.to_string())?;
        }
        let stream = self.session.open().map_err(|err| format!("failed to open SMUX stream for HTTP: {err}"))?;
        let (mut sender, connection) =
            hyper::client::conn::http1::handshake(TokioIo::new(stream)).await.map_err(|e| e.to_string())?;
        tokio::spawn(connection);
        let response = sender
            .send_request(Request::from_parts(parts, body))
            .await
            .map_err(|err| format!("failed to read HTTP response: {err}"))?;
        let status = response.status();
        let body = response.into_body().collect().await.map_err(|e| e.to_string())?.to_bytes();
        Ok((status, body))
    }

    /// failure stops the remote and explains a failed step by its error output, or by the step's error when it wrote
    /// none, as Dolt's sshRemoteError does.
    pub(crate) fn failure(&mut self, step: &str, err: impl std::fmt::Display) -> String {
        self.stop();
        if let Some(reader) = self.reader.take() {
            let _ = reader.join();
        }
        let output = String::from_utf8_lossy(&self.stderr.lock().unwrap_or_else(|e| e.into_inner())).into_owned();
        let lines: Vec<&str> = output
            .lines()
            .map(str::trim)
            .filter(|line| !line.is_empty() && !line.starts_with("Warning: Permanently added"))
            .collect();
        let output = lines.join("\n");
        if output.contains("no such file or directory") || output.contains("failed to load database") {
            format!("repository not found at {}", self.path)
        } else if !output.is_empty() {
            format!("{step}: remote: {output}")
        } else {
            format!("{step}: {err}")
        }
    }

    /// stop closes the remote's input and waits a second for it to exit before killing it.
    fn stop(&mut self) {
        self.session.close();
        let deadline = Instant::now() + Duration::from_secs(1);
        while matches!(self.child.try_wait(), Ok(None)) && Instant::now() < deadline {
            std::thread::sleep(Duration::from_millis(10));
        }
        if matches!(self.child.try_wait(), Ok(None)) {
            let _ = self.child.kill();
            let _ = self.child.wait();
        }
    }
}

impl Drop for Ssh {
    fn drop(&mut self) {
        self.stop();
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn urls_name_the_transfer_command() {
        let address = |host: &str, port: &str, path: &str, user: &str| Address {
            host: host.into(),
            port: port.into(),
            path: path.into(),
            user: user.into(),
        };
        assert_eq!(parse("ssh://host/srv/repo"), address("host", "", "/srv/repo", ""));
        assert_eq!(parse("ssh://me@host:2222/srv/repo/.dolt"), address("host", "2222", "/srv/repo", "me"));
        assert_eq!(parse("ssh://me:secret@[::1]:22/srv/repo/"), address("::1", "22", "/srv/repo", "me"));
        assert_eq!(parse("ssh://host"), address("host", "", "", ""));
        let args = |command: &Command| -> Vec<String> {
            std::iter::once(command.get_program())
                .chain(command.get_args())
                .map(|a| a.to_string_lossy().into_owned())
                .collect()
        };
        let plain = command(&parse("ssh://host/srv/repo"), "ssh", "dolt").unwrap();
        assert_eq!(args(&plain), ["ssh", "host", "dolt --data-dir /srv/repo transfer"]);
        let custom =
            command(&parse("ssh://me@host:2222/srv/repo"), "ssh -i key -o BatchMode=yes", "/opt/dolt").unwrap();
        assert_eq!(
            args(&custom),
            [
                "ssh",
                "-i",
                "key",
                "-o",
                "BatchMode=yes",
                "-p",
                "2222",
                "me@host",
                "/opt/dolt --data-dir /srv/repo transfer"
            ]
        );
        assert_eq!(command(&parse("ssh://host/r"), " ", "dolt").err().unwrap(), "invalid DOLT_SSH_COMMAND: empty");
    }
}
