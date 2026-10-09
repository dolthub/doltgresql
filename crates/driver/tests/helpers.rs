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

use driver::process::make_temp_dir;
use driver::runner::prepare_server_args;
use driver::yaml;

#[test]
fn prepare_server_args_keeps_logging() {
    for level in ["", "info", "warn", "trace"] {
        let dir = make_temp_dir("helpers-").unwrap();
        let mut args = Vec::new();
        if !level.is_empty() {
            std::fs::write(dir.join("server.yaml"), format!("log_level: {level}\n")).unwrap();
            args = vec!["--config".to_string(), "server.yaml".to_string()];
        }
        prepare_server_args(&dir, "test", 5432, &args).unwrap();
        let config = yaml::parse(&std::fs::read_to_string(dir.join(".generated-test-config.yaml")).unwrap()).unwrap();
        match level {
            "" => assert!(config.get("log_level").is_none(), "preserve the server's default logging"),
            level => assert_eq!(config.get("log_level").unwrap().string().unwrap(), level),
        }
        std::fs::remove_dir_all(dir).unwrap();
    }
}

#[test]
fn scalars_decode_as_written() {
    let node = yaml::parse("rows: [[1, 1.50, true, ~, 'x'], [\"\", 007]]").unwrap();
    let rows: Vec<Vec<String>> =
        node.get("rows").unwrap().sequence().unwrap().iter().map(|row| row.strings().unwrap()).collect();
    assert_eq!(rows, vec![vec!["1", "1.50", "true", "", "x"], vec!["", "007"]]);
}
