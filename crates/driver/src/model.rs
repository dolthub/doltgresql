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

use crate::yaml::Node;

/// TestDef is a file of tests.
#[derive(Clone, Debug, Default)]
pub struct TestDef {
    pub tests: Vec<Test>,
    pub parallel: bool,
}

/// Test creates repos and servers, then runs the connections' queries against them.
#[derive(Clone, Debug, Default)]
pub struct Test {
    pub name: String,
    pub repos: Vec<TestRepo>,
    pub multi_repos: Vec<MultiRepo>,
    pub connections: Vec<Connection>,
    pub skip: String,
}

/// TestRepo is a database in its own store, optionally with a server.
#[derive(Clone, Debug, Default)]
pub struct TestRepo {
    pub name: String,
    pub with_files: Vec<WithFile>,
    pub with_remotes: Vec<WithRemote>,
    pub server: Option<Server>,
}

/// MultiRepo is several databases in one store with one server.
#[derive(Clone, Debug, Default)]
pub struct MultiRepo {
    pub name: String,
    pub repos: Vec<TestRepo>,
    pub with_files: Vec<WithFile>,
    pub server: Option<Server>,
}

/// WithRemote adds a remote to a repo.
#[derive(Clone, Debug, Default)]
pub struct WithRemote {
    pub name: String,
    pub url: String,
}

/// WithFile writes a file into a store, from contents or from a source path.
#[derive(Clone, Debug, Default)]
pub struct WithFile {
    pub name: String,
    pub contents: String,
    pub source_path: String,
}

/// Server is a server process to start.
#[derive(Clone, Debug, Default)]
pub struct Server {
    pub name: String,
    pub args: Vec<String>,
    pub envs: Vec<String>,
    pub port: i64,
    pub dynamic_port: String,
    pub debug_port: i64,
    pub log_matches: Vec<String>,
    pub log_not_matches: Vec<String>,
    pub error_matches: Vec<String>,
}

/// Connection runs queries on one connection to a server, then optionally restarts the server.
#[derive(Clone, Debug, Default)]
pub struct Connection {
    pub on: String,
    pub queries: Vec<Query>,
    pub restart_server: Option<RestartArgs>,
    pub retry_attempts: i64,
    pub user: String,
    pub database: String,
    pub password: String,
    pub password_file: String,
    pub driver_params: Vec<(String, String)>,
}

/// RestartArgs replaces the server's arguments or adds environment variables on a restart.
#[derive(Clone, Debug, Default)]
pub struct RestartArgs {
    pub args: Option<Vec<String>>,
    pub envs: Option<Vec<String>>,
}

/// Query is a query or a statement with its expected result or error.
#[derive(Clone, Debug, Default)]
pub struct Query {
    pub query: String,
    pub exec: String,
    pub args: Vec<String>,
    pub columns: Vec<String>,
    pub rows: Option<Vec<Vec<Vec<String>>>>,
    pub error_match: String,
    pub retry_attempts: i64,
}

/// field returns a mapping value, treating a missing key like null.
fn field<'a>(node: &'a Node, key: &str) -> &'a Node {
    static NULL: Node = Node::Scalar { text: String::new(), plain: true };
    node.get(key).unwrap_or(&NULL)
}

fn optional<'a>(node: &'a Node, key: &str) -> Option<&'a Node> {
    node.get(key).filter(|n| !n.is_null())
}

fn list<T>(node: &Node, key: &str, parse: fn(&Node) -> Result<T, String>) -> Result<Vec<T>, String> {
    field(node, key).sequence()?.iter().map(parse).collect()
}

/// parse_test_def parses a test file.
pub fn parse_test_def(node: &Node) -> Result<TestDef, String> {
    let parallel = matches!(field(node, "parallel"), Node::Scalar { text, plain: true } if text == "true");
    Ok(TestDef { tests: list(node, "tests", parse_test)?, parallel })
}

fn parse_test(node: &Node) -> Result<Test, String> {
    Ok(Test {
        name: field(node, "name").string()?,
        repos: list(node, "repos", parse_repo)?,
        multi_repos: list(node, "multi_repos", parse_multi_repo)?,
        connections: list(node, "connections", parse_connection)?,
        skip: field(node, "skip").string()?,
    })
}

fn parse_repo(node: &Node) -> Result<TestRepo, String> {
    Ok(TestRepo {
        name: field(node, "name").string()?,
        with_files: list(node, "with_files", parse_file)?,
        with_remotes: list(node, "with_remotes", |n| {
            Ok(WithRemote { name: field(n, "name").string()?, url: field(n, "url").string()? })
        })?,
        server: optional(node, "server").map(parse_server).transpose()?,
    })
}

fn parse_multi_repo(node: &Node) -> Result<MultiRepo, String> {
    Ok(MultiRepo {
        name: field(node, "name").string()?,
        repos: list(node, "repos", parse_repo)?,
        with_files: list(node, "with_files", parse_file)?,
        server: optional(node, "server").map(parse_server).transpose()?,
    })
}

fn parse_file(node: &Node) -> Result<WithFile, String> {
    Ok(WithFile {
        name: field(node, "name").string()?,
        contents: field(node, "contents").string()?,
        source_path: field(node, "source_path").string()?,
    })
}

fn parse_server(node: &Node) -> Result<Server, String> {
    Ok(Server {
        name: field(node, "name").string()?,
        args: field(node, "args").strings()?,
        envs: field(node, "envs").strings()?,
        port: field(node, "port").int()?,
        dynamic_port: field(node, "dynamic_port").string()?,
        debug_port: field(node, "debug_port").int()?,
        log_matches: field(node, "log_matches").strings()?,
        log_not_matches: field(node, "log_not_matches").strings()?,
        error_matches: field(node, "error_matches").strings()?,
    })
}

fn parse_connection(node: &Node) -> Result<Connection, String> {
    let user = optional(node, "user").map(Node::string).transpose()?.unwrap_or_else(|| "postgres".to_string());
    Ok(Connection {
        on: field(node, "on").string()?,
        queries: list(node, "queries", parse_query)?,
        restart_server: optional(node, "restart_server")
            .map(|n| -> Result<RestartArgs, String> {
                Ok(RestartArgs {
                    args: optional(n, "args").map(Node::strings).transpose()?,
                    envs: optional(n, "envs").map(Node::strings).transpose()?,
                })
            })
            .transpose()?,
        retry_attempts: field(node, "retry_attempts").int()?,
        user,
        database: field(node, "database").string()?,
        password: field(node, "password").string()?,
        password_file: field(node, "password_file").string()?,
        driver_params: field(node, "driver_params")
            .mapping()?
            .iter()
            .map(|(k, v)| Ok((k.clone(), v.string()?)))
            .collect::<Result<_, String>>()?,
    })
}

fn parse_rows(node: &Node) -> Result<Vec<Vec<String>>, String> {
    node.sequence()?.iter().map(Node::strings).collect()
}

fn parse_query(node: &Node) -> Result<Query, String> {
    let result = field(node, "result");
    let rows = match optional(result, "rows") {
        None => None,
        Some(rows @ Node::Sequence(_)) => Some(vec![parse_rows(rows)?]),
        Some(rows) => optional(rows, "or").map(|or| or.sequence()?.iter().map(parse_rows).collect()).transpose()?,
    };
    Ok(Query {
        query: field(node, "query").string()?,
        exec: field(node, "exec").string()?,
        args: field(node, "args").strings()?,
        columns: field(result, "columns").strings()?,
        rows,
        error_match: field(node, "error_match").string()?,
        retry_attempts: field(node, "retry_attempts").int()?,
    })
}
