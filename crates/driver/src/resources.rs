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

use std::collections::HashMap;
use std::path::PathBuf;
use std::sync::Mutex;

/// PORTS holds the ports that tests may use, taken from the end and returned when a test finishes.
static PORTS: Mutex<Vec<u16>> = Mutex::new(Vec::new());

/// take_port takes a port from the shared pool, which starts as 5432 through 5687.
fn take_port() -> Result<u16, String> {
    let mut ports = PORTS.lock().unwrap();
    static FILLED: std::sync::Once = std::sync::Once::new();
    FILLED.call_once(|| ports.extend((0..256).map(|i| 5432 + i)));
    ports.pop().ok_or_else(|| "cannot get a port; we are all out.".to_string())
}

/// Resources holds a test's named ports and temporary directories, returning the ports when dropped.
#[derive(Default)]
pub struct Resources {
    ports: HashMap<String, u16>,
    temp_dirs: HashMap<String, PathBuf>,
}

impl Resources {
    /// port returns the named port, taking one from the pool the first time.
    pub fn port(&mut self, name: &str) -> Result<u16, String> {
        if let Some(port) = self.ports.get(name) {
            return Ok(*port);
        }
        let port = take_port()?;
        self.ports.insert(name.to_string(), port);
        Ok(port)
    }

    /// temp_dir returns the named temporary directory, creating it the first time.
    pub fn temp_dir(&mut self, name: &str) -> Result<PathBuf, String> {
        if let Some(dir) = self.temp_dirs.get(name) {
            return Ok(dir.clone());
        }
        let dir = crate::process::make_temp_dir("tempdir-")?;
        self.temp_dirs.insert(name.to_string(), dir.clone());
        Ok(dir)
    }

    /// apply_template replaces `{{get_port "name"}}` and `{{get_tempdir "name"}}` actions.
    pub fn apply_template(&mut self, text: &str) -> Result<String, String> {
        let mut out = String::new();
        let mut rest = text;
        while let Some(start) = rest.find("{{") {
            out.push_str(&rest[..start]);
            let end = rest[start..].find("}}").ok_or_else(|| format!("unclosed action in {text:?}"))? + start;
            let action = rest[start + 2..end].trim();
            let (function, argument) = action.split_once(char::is_whitespace).unwrap_or((action, ""));
            let name = argument.trim().strip_prefix('"').and_then(|a| a.strip_suffix('"'));
            let name = name.ok_or_else(|| format!("unsupported action {{{{{action}}}}} in {text:?}"))?;
            match function {
                "get_port" => out.push_str(&self.port(name)?.to_string()),
                "get_tempdir" => out.push_str(&self.temp_dir(name)?.to_string_lossy()),
                _ => return Err(format!("function {function:?} not defined")),
            }
            rest = &rest[end + 2..];
        }
        out.push_str(rest);
        Ok(out)
    }
}

impl Drop for Resources {
    fn drop(&mut self) {
        PORTS.lock().unwrap().extend(self.ports.values());
        for dir in self.temp_dirs.values() {
            let _ = std::fs::remove_dir_all(dir);
        }
    }
}
