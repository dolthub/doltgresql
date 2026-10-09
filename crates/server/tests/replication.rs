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

//! A port of testing/go/replication_test.go, which replicates from the Postgres primary that the URL in
//! DOLTGRES_REPLICATION_PRIMARY names, whose wal_level must be logical, into the server that DOLTGRES_TEST_TARGET names.

use std::collections::HashMap;
use std::sync::Arc;
use std::thread::JoinHandle;
use std::time::{Duration, Instant};

use harness::pgx::{Conn, ConnConfig};
use harness::server::{Server, Target};
use server::logrepl::{LogicalReplicator, create_publication, drop_publication};

/// CREATE_REPLICATION_SLOT is the pseudo-query that creates the replication slot unless it exists.
const CREATE_REPLICATION_SLOT: &str = "createReplicationSlot";
/// DROP_REPLICATION_SLOT is the pseudo-query that drops the replication slot.
const DROP_REPLICATION_SLOT: &str = "dropReplicationSlot";
/// STOP_REPLICATION is the pseudo-query that stops replication and waits for it to end.
const STOP_REPLICATION: &str = "stopReplication";
/// START_REPLICATION is the pseudo-query that starts replication on its own thread.
const START_REPLICATION: &str = "startReplication";
/// WAIT_FOR_CATCHUP is the pseudo-query that waits until the replica nearly caught up with the primary.
const WAIT_FOR_CATCHUP: &str = "waitForCatchup";
/// SLEEP is the pseudo-query that waits a fifth of a second.
const SLEEP: &str = "sleep";

/// SLOT_NAME names the replication slot and the publication.
const SLOT_NAME: &str = "doltgres_slot";

/// ReplicationTest runs its setup on the primary, or on the replica when a statement's comment says so, then checks
/// the rows that a query on the replica returns.
struct ReplicationTest {
    name: &'static str,
    set_up_script: &'static [&'static str],
    query: &'static str,
    expected: &'static [&'static [&'static str]],
}

#[test]
fn test_replication() {
    let Ok(primary) = std::env::var("DOLTGRES_REPLICATION_PRIMARY") else { return };
    let target = Target::from_env().unwrap();
    drop_publication(&primary, SLOT_NAME).unwrap();
    create_publication(&primary, SLOT_NAME).unwrap();
    std::thread::sleep(Duration::from_millis(500));
    for script in REPLICATION_TESTS {
        run_replication_script(&target, &primary, script);
    }
}

/// Runner is a script's replicator and connections.
struct Runner {
    replicator: Arc<LogicalReplicator>,
    replication: Option<JoinHandle<Result<(), String>>>,
    primary: String,
    connections: HashMap<String, Conn>,
    replica: Conn,
}

/// run_replication_script runs a script against a new server.
fn run_replication_script(target: &Target, primary: &str, script: &ReplicationTest) {
    let server = Server::start(target, "").unwrap();
    let replica_url = format!("postgres://postgres:password@127.0.0.1:{}/", server.port);
    let wal_file = server.directory().join("wal");
    let mut runner = Runner {
        replicator: Arc::new(LogicalReplicator::new(wal_file, primary.to_string(), replica_url.clone())),
        replication: None,
        primary: primary.to_string(),
        connections: HashMap::new(),
        replica: connect(&replica_url),
    };
    for query in script.set_up_script {
        if runner.pseudo_query(query) {
            continue;
        }
        let conn = runner.connection(query);
        if let Some(err) = conn.simple_query(query).unwrap().error {
            panic!("{}: setup query {query} failed: {err}", script.name);
        }
    }
    let mut rows = Vec::new();
    for attempt in 0..3 {
        let result = runner.replica.simple_query_rows(script.query).unwrap();
        assert!(result.error.is_none(), "{}: {:?}", script.name, result.error);
        rows = result
            .rows
            .iter()
            .map(|row| row.iter().map(|v| v.as_deref().map(|v| String::from_utf8_lossy(v).into_owned())).collect())
            .collect::<Vec<Vec<Option<String>>>>();
        let matches = rows.len() == script.expected.len()
            && rows.iter().zip(script.expected).all(|(got, want)| {
                got.len() == want.len() && got.iter().zip(*want).all(|(g, w)| g.as_deref() == Some(*w))
            });
        if matches {
            runner.stop();
            return;
        }
        if attempt < 2 {
            std::thread::sleep(Duration::from_millis(500));
        }
    }
    runner.stop();
    panic!("{}: expected {:?}, got {rows:?}", script.name, script.expected);
}

impl Runner {
    /// connection returns the connection that a query's comment names: the replica, or a client of the primary.
    fn connection(&mut self, query: &str) -> &mut Conn {
        let (target, client) = client_spec_from_query_comment(query);
        if target == "replica" {
            return &mut self.replica;
        }
        let primary = self.primary.clone();
        self.connections.entry(client).or_insert_with(|| connect(&primary))
    }

    /// pseudo_query runs a pseudo-query, reporting whether the query was one.
    fn pseudo_query(&mut self, query: &str) -> bool {
        match query {
            CREATE_REPLICATION_SLOT => self.replicator.create_replication_slot_if_necessary(SLOT_NAME).unwrap(),
            DROP_REPLICATION_SLOT => self.replicator.drop_replication_slot(SLOT_NAME).unwrap(),
            START_REPLICATION => {
                let replicator = self.replicator.clone();
                self.replication = Some(std::thread::spawn(move || replicator.start_replication(SLOT_NAME)));
                let start = Instant::now();
                while !self.replicator.running() {
                    assert!(start.elapsed() < Duration::from_millis(500), "Replication did not start");
                    std::thread::sleep(Duration::from_millis(5));
                }
            }
            STOP_REPLICATION => self.stop(),
            WAIT_FOR_CATCHUP => {
                let start = Instant::now();
                while !self.replicator.caught_up(150).unwrap() {
                    assert!(start.elapsed() < Duration::from_secs(2), "Replication did not catch up");
                    std::thread::sleep(Duration::from_millis(20));
                }
            }
            SLEEP => std::thread::sleep(Duration::from_millis(200)),
            _ => return false,
        }
        true
    }

    /// stop stops replication, failing when it ended with an error.
    fn stop(&mut self) {
        self.replicator.stop();
        if let Some(handle) = self.replication.take() {
            handle.join().unwrap().unwrap();
        }
    }
}

/// client_spec_from_query_comment returns where a query runs from its comment: "replica", or "primary" with the name
/// of the primary's client, which is "a" without a comment.
fn client_spec_from_query_comment(query: &str) -> (&'static str, String) {
    let (Some(start), Some(end)) = (query.find("/*"), query.find("*/")) else { return ("primary", "a".to_string()) };
    let comment = &query[start + 2..end];
    if comment.contains("replica") {
        return ("replica", "a".to_string());
    }
    match comment.find("primary ") {
        Some(i) if i > 0 && i + "primary ".len() < comment.len() => {
            ("primary", comment[i + "primary ".len()..].to_string())
        }
        _ => ("primary", "a".to_string()),
    }
}

/// connect connects to a server by its URL.
fn connect(url: &str) -> Conn {
    Conn::connect(ConnConfig::parse(url).unwrap()).unwrap()
}

/// REPLICATION_TESTS are the scripts of replication_test.go.
const REPLICATION_TESTS: &[ReplicationTest] = &[
    ReplicationTest {
        name: "simple replication, strings and integers",
        set_up_script: &[
            DROP_REPLICATION_SLOT,
            CREATE_REPLICATION_SLOT,
            START_REPLICATION,
            "/* replica */ drop table if exists public.test",
            "/* replica */ create table public.test (id INT primary key, name varchar(100))",
            "drop table if exists public.test",
            "CREATE TABLE public.test (id INT primary key, name varchar(100))",
            "INSERT INTO public.test VALUES (1, 'one')",
            "INSERT INTO public.test VALUES (2, 'two')",
            "UPDATE public.test SET name = 'three' WHERE id = 2",
            "DELETE FROM public.test WHERE id = 1",
            "INSERT INTO public.test VALUES (3, 'one')",
            "INSERT INTO public.test VALUES (4, 'two')",
            "UPDATE public.test SET name = 'five' WHERE id = 4",
            "DELETE FROM public.test WHERE id = 3",
            "INSERT INTO public.test VALUES (5, 'one')",
            "INSERT INTO public.test VALUES (6, 'two')",
            "UPDATE public.test SET name = 'six' WHERE id = 6",
            "DELETE FROM public.test WHERE id = 5",
            WAIT_FOR_CATCHUP,
        ],
        query: "/* replica */ SELECT * FROM public.test order by id",
        expected: &[&["2", "three"], &["4", "five"], &["6", "six"]],
    },
    ReplicationTest {
        name: "stale start",
        set_up_script: &[
            DROP_REPLICATION_SLOT,
            CREATE_REPLICATION_SLOT,
            "/* replica */ drop table if exists public.test",
            "/* replica */ create table public.test (id INT primary key, name varchar(100))",
            "drop table if exists public.test",
            "CREATE TABLE public.test (id INT primary key, name varchar(100))",
            "INSERT INTO public.test VALUES (1, 'one')",
            "INSERT INTO public.test VALUES (2, 'two')",
            "UPDATE public.test SET name = 'three' WHERE id = 2",
            "DELETE FROM public.test WHERE id = 1",
            "INSERT INTO public.test VALUES (3, 'one')",
            "INSERT INTO public.test VALUES (4, 'two')",
            "UPDATE public.test SET name = 'five' WHERE id = 4",
            "DELETE FROM public.test WHERE id = 3",
            "INSERT INTO public.test VALUES (5, 'one')",
            "INSERT INTO public.test VALUES (6, 'two')",
            "UPDATE public.test SET name = 'six' WHERE id = 6",
            "DELETE FROM public.test WHERE id = 5",
            START_REPLICATION,
            WAIT_FOR_CATCHUP,
        ],
        query: "/* replica */ SELECT * FROM public.test order by id",
        expected: &[&["2", "three"], &["4", "five"], &["6", "six"]],
    },
    ReplicationTest {
        name: "stopping and resuming replication",
        set_up_script: &[
            DROP_REPLICATION_SLOT,
            CREATE_REPLICATION_SLOT,
            START_REPLICATION,
            "/* replica */ drop table if exists public.test",
            "/* replica */ create table public.test (id INT primary key, name varchar(100))",
            "drop table if exists public.test",
            "CREATE TABLE public.test (id INT primary key, name varchar(100))",
            "INSERT INTO public.test VALUES (1, 'one')",
            "INSERT INTO public.test VALUES (2, 'two')",
            WAIT_FOR_CATCHUP,
            STOP_REPLICATION,
            "UPDATE public.test SET name = 'three' WHERE id = 2",
            "DELETE FROM public.test WHERE id = 1",
            "INSERT INTO public.test VALUES (3, 'one')",
            "INSERT INTO public.test VALUES (4, 'two')",
            "UPDATE public.test SET name = 'five' WHERE id = 4",
            "DELETE FROM public.test WHERE id = 3",
            START_REPLICATION,
            "INSERT INTO public.test VALUES (5, 'one')",
            "INSERT INTO public.test VALUES (6, 'two')",
            "UPDATE public.test SET name = 'six' WHERE id = 6",
            "DELETE FROM public.test WHERE id = 5",
            WAIT_FOR_CATCHUP,
        ],
        query: "/* replica */ SELECT * FROM public.test order by id",
        expected: &[&["2", "three"], &["4", "five"], &["6", "six"]],
    },
    ReplicationTest {
        name: "extended stop/start",
        set_up_script: &[
            DROP_REPLICATION_SLOT,
            CREATE_REPLICATION_SLOT,
            "/* replica */ drop table if exists public.test",
            "/* replica */ create table public.test (id INT primary key, name varchar(100))",
            "drop table if exists public.test",
            "CREATE TABLE public.test (id INT primary key, name varchar(100))",
            "INSERT INTO public.test VALUES (1, 'one')",
            "INSERT INTO public.test VALUES (2, 'two')",
            "UPDATE public.test SET name = 'three' WHERE id = 2",
            "DELETE FROM public.test WHERE id = 1",
            "INSERT INTO public.test VALUES (3, 'one')",
            "INSERT INTO public.test VALUES (4, 'two')",
            "UPDATE public.test SET name = 'five' WHERE id = 4",
            "DELETE FROM public.test WHERE id = 3",
            "INSERT INTO public.test VALUES (5, 'one')",
            START_REPLICATION,
            "INSERT INTO public.test VALUES (6, 'two')",
            "UPDATE public.test SET name = 'six' WHERE id = 6",
            STOP_REPLICATION,
            "DELETE FROM public.test WHERE id = 5",
            "INSERT INTO public.test VALUES (7, 'one')",
            "INSERT INTO public.test VALUES (8, 'two')",
            START_REPLICATION,
            "UPDATE public.test SET name = 'nine' WHERE id = 8",
            "DELETE FROM public.test WHERE id = 7",
            "INSERT INTO public.test VALUES (9, 'one')",
            STOP_REPLICATION,
            START_REPLICATION,
            "INSERT INTO public.test VALUES (10, 'two')",
            "UPDATE public.test SET name = 'eleven' WHERE id = 10",
            STOP_REPLICATION,
            "DELETE FROM public.test WHERE id = 9",
            "INSERT INTO public.test VALUES (11, 'one')",
            "INSERT INTO public.test VALUES (12, 'two')",
            "UPDATE public.test SET name = 'thirteen' WHERE id = 12",
            "DELETE FROM public.test WHERE id = 11",
            START_REPLICATION,
            "INSERT INTO public.test VALUES (13, 'one')",
            "INSERT INTO public.test VALUES (14, 'two')",
            "UPDATE public.test SET name = 'fifteen' WHERE id = 14",
            "DELETE FROM public.test WHERE id = 13",
            WAIT_FOR_CATCHUP,
            SLEEP,
            STOP_REPLICATION,
            "INSERT INTO public.test VALUES (15, 'one')",
            "INSERT INTO public.test VALUES (16, 'two')",
            "UPDATE public.test SET name = 'seventeen' WHERE id = 16",
            "DELETE FROM public.test WHERE id = 15",
            SLEEP,
        ],
        query: "/* replica */ SELECT * FROM public.test order by id",
        expected: &[
            &["2", "three"],
            &["4", "five"],
            &["6", "six"],
            &["8", "nine"],
            &["10", "eleven"],
            &["12", "thirteen"],
            &["14", "fifteen"],
        ],
    },
    ReplicationTest {
        name: "all supported types",
        set_up_script: &[
            DROP_REPLICATION_SLOT,
            CREATE_REPLICATION_SLOT,
            START_REPLICATION,
            "/* replica */ drop table if exists public.test",
            "/* replica */ create table public.test (id INT primary key, name varchar(100), u_id uuid, age INT, height FLOAT)",
            "drop table if exists public.test",
            "create table public.test (id INT primary key, name varchar(100), u_id uuid, age INT, height FLOAT)",
            "INSERT INTO public.test VALUES (1, 'one', '5ef34887-e635-4c9c-a994-97b1cb810786', 1, 1.1)",
            "INSERT INTO public.test VALUES (2, 'two', '2de55648-76ec-4f66-9fae-bd3d853fb0da', 2, 2.2)",
            "UPDATE public.test SET name = 'three' WHERE id = 2",
            "update public.test set u_id = '3232abe7-560b-4714-a020-2b1a11a1ec65' where id = 2",
            "DELETE FROM public.test WHERE id = 1",
            WAIT_FOR_CATCHUP,
        ],
        query: "/* replica */ SELECT * FROM public.test order by id",
        expected: &[&["2", "three", "3232abe7-560b-4714-a020-2b1a11a1ec65", "2", "2.2"]],
    },
    ReplicationTest {
        name: "concurrent writes",
        set_up_script: &[
            DROP_REPLICATION_SLOT,
            CREATE_REPLICATION_SLOT,
            START_REPLICATION,
            "/* replica */ drop table if exists public.test",
            "/* replica */ create table public.test (id INT primary key, name varchar(100))",
            "drop table if exists public.test",
            "CREATE TABLE public.test (id INT primary key, name varchar(100))",
            "/* primary a */ START TRANSACTION",
            "/* primary a */ INSERT INTO public.test VALUES (1, 'one')",
            "/* primary a */ INSERT INTO public.test VALUES (2, 'two')",
            "/* primary b */ START TRANSACTION",
            "/* primary b */ INSERT INTO public.test VALUES (3, 'one')",
            "/* primary b */ INSERT INTO public.test VALUES (4, 'two')",
            "/* primary a */ UPDATE public.test SET name = 'three' WHERE id > 0",
            "/* primary a */ DELETE FROM public.test WHERE id = 1",
            "/* primary b */ UPDATE public.test SET name = 'five' WHERE id > 0",
            "/* primary b */ DELETE FROM public.test WHERE id = 3",
            "/* primary b */ COMMIT",
            "/* primary a */ COMMIT",
            WAIT_FOR_CATCHUP,
        ],
        query: "/* replica */ SELECT * FROM public.test order by id",
        expected: &[&["2", "three"], &["4", "five"]],
    },
    ReplicationTest {
        name: "concurrent writes with restarts",
        set_up_script: &[
            DROP_REPLICATION_SLOT,
            CREATE_REPLICATION_SLOT,
            START_REPLICATION,
            "/* replica */ drop table if exists public.test",
            "/* replica */ create table public.test (id INT primary key, name varchar(100))",
            "drop table if exists public.test",
            "CREATE TABLE public.test (id INT primary key, name varchar(100))",
            "/* primary a */ START TRANSACTION",
            "/* primary a */ INSERT INTO public.test VALUES (1, 'one')",
            "/* primary a */ INSERT INTO public.test VALUES (2, 'two')",
            STOP_REPLICATION,
            "/* primary b */ START TRANSACTION",
            "/* primary b */ INSERT INTO public.test VALUES (3, 'one')",
            "/* primary b */ INSERT INTO public.test VALUES (4, 'two')",
            "/* primary c */ START TRANSACTION",
            "/* primary c */ INSERT INTO public.test VALUES (5, 'one')",
            "/* primary c */ INSERT INTO public.test VALUES (6, 'two')",
            "/* primary a */ UPDATE public.test SET name = 'three' WHERE id > 0",
            START_REPLICATION,
            "/* primary a */ DELETE FROM public.test WHERE id = 1",
            "/* primary b */ UPDATE public.test SET name = 'five' WHERE id > 0",
            "/* primary b */ DELETE FROM public.test WHERE id = 3",
            "/* primary b */ COMMIT",
            STOP_REPLICATION,
            "/* primary c */ UPDATE public.test SET name = 'seven' WHERE id > 0",
            "/* primary c */ DELETE FROM public.test WHERE id = 5",
            "/* primary a */ COMMIT",
            START_REPLICATION,
            "/* primary c */ COMMIT",
            WAIT_FOR_CATCHUP,
        ],
        query: "/* replica */ SELECT * FROM public.test order by id",
        expected: &[&["2", "three"], &["4", "seven"], &["6", "seven"]],
    },
    ReplicationTest {
        name: "concurrent writes with rollbacks",
        set_up_script: &[
            DROP_REPLICATION_SLOT,
            CREATE_REPLICATION_SLOT,
            START_REPLICATION,
            "/* replica */ drop table if exists public.test",
            "/* replica */ create table public.test (id INT primary key, name varchar(100))",
            "drop table if exists public.test",
            "CREATE TABLE public.test (id INT primary key, name varchar(100))",
            "/* primary a */ START TRANSACTION",
            "/* primary a */ INSERT INTO public.test VALUES (1, 'one')",
            "/* primary a */ INSERT INTO public.test VALUES (2, 'two')",
            STOP_REPLICATION,
            "/* primary b */ START TRANSACTION",
            "/* primary b */ INSERT INTO public.test VALUES (3, 'one')",
            "/* primary b */ INSERT INTO public.test VALUES (4, 'two')",
            "/* primary c */ START TRANSACTION",
            "/* primary c */ INSERT INTO public.test VALUES (5, 'one')",
            "/* primary c */ INSERT INTO public.test VALUES (6, 'two')",
            "/* primary a */ UPDATE public.test SET name = 'three' WHERE id > 0",
            START_REPLICATION,
            "/* primary a */ DELETE FROM public.test WHERE id = 1",
            "/* primary b */ UPDATE public.test SET name = 'five' WHERE id > 0",
            "/* primary b */ DELETE FROM public.test WHERE id = 3",
            "/* primary b */ COMMIT",
            STOP_REPLICATION,
            "/* primary c */ UPDATE public.test SET name = 'seven' WHERE id > 0",
            "/* primary c */ DELETE FROM public.test WHERE id = 5",
            "/* primary a */ ROLLBACK",
            START_REPLICATION,
            "/* primary c */ COMMIT",
            WAIT_FOR_CATCHUP,
        ],
        query: "/* replica */ SELECT * FROM public.test order by id",
        expected: &[&["4", "seven"], &["6", "seven"]],
    },
    ReplicationTest {
        name: "concurrent writes, stale commits",
        set_up_script: &[
            DROP_REPLICATION_SLOT,
            CREATE_REPLICATION_SLOT,
            START_REPLICATION,
            "/* replica */ drop table if exists public.test",
            "/* replica */ create table public.test (id INT primary key, name varchar(100))",
            "drop table if exists public.test",
            "CREATE TABLE public.test (id INT primary key, name varchar(100))",
            "/* primary a */ START TRANSACTION",
            "/* primary a */ INSERT INTO public.test VALUES (1, 'one')",
            "/* primary b */ START TRANSACTION",
            "/* primary b */ INSERT INTO public.test VALUES (2, 'two')",
            "/* primary b */ COMMIT",
            WAIT_FOR_CATCHUP,
            STOP_REPLICATION,
            START_REPLICATION,
            "/* primary a */ COMMIT",
            WAIT_FOR_CATCHUP,
        ],
        query: "/* replica */ SELECT * FROM public.test order by id",
        expected: &[&["1", "one"], &["2", "two"]],
    },
    ReplicationTest {
        name: "concurrent writes, very stale commits",
        set_up_script: &[
            DROP_REPLICATION_SLOT,
            CREATE_REPLICATION_SLOT,
            START_REPLICATION,
            "/* replica */ drop table if exists public.test",
            "/* replica */ create table public.test (id INT primary key, name varchar(100))",
            "drop table if exists public.test",
            "CREATE TABLE public.test (id INT primary key, name varchar(100))",
            "/* primary a */ START TRANSACTION",
            "/* primary a */ INSERT INTO public.test VALUES (1, 'one')",
            "/* primary a */ INSERT INTO public.test VALUES (2, 'two')",
            "/* primary a */ UPDATE public.test SET name = 'three' WHERE id > 0",
            "/* primary a */ DELETE FROM public.test WHERE id = 1",
            "/* primary b */ START TRANSACTION",
            "/* primary b */ INSERT INTO public.test VALUES (3, 'one')",
            "/* primary b */ INSERT INTO public.test VALUES (4, 'two')",
            "/* primary c */ START TRANSACTION",
            "/* primary c */ INSERT INTO public.test VALUES (5, 'one')",
            "/* primary c */ INSERT INTO public.test VALUES (6, 'two')",
            "/* primary c */ UPDATE public.test SET name = 'seven' WHERE id > 0",
            "/* primary c */ DELETE FROM public.test WHERE id = 5",
            "/* primary c */ COMMIT",
            "/* primary b */ UPDATE public.test SET name = 'five' WHERE id > 0",
            "/* primary b */ DELETE FROM public.test WHERE id = 3",
            "/* primary b */ COMMIT",
            WAIT_FOR_CATCHUP,
            STOP_REPLICATION,
            START_REPLICATION,
            "/* primary a */ COMMIT",
            WAIT_FOR_CATCHUP,
        ],
        query: "/* replica */ SELECT * FROM public.test order by id",
        expected: &[&["2", "three"], &["4", "five"], &["6", "five"]],
    },
    ReplicationTest {
        name: "all types",
        set_up_script: &[
            DROP_REPLICATION_SLOT,
            CREATE_REPLICATION_SLOT,
            START_REPLICATION,
            "/* replica */ drop table if exists public.test",
            "/* replica */ create table public.test (id INT primary key, name varchar(100), age INT, is_cool BOOLEAN, height FLOAT, birth_date DATE, birth_timestamp TIMESTAMP)",
            "drop table if exists public.test",
            "create table public.test (id INT primary key, name varchar(100), age INT, is_cool BOOLEAN, height FLOAT, birth_date DATE, birth_timestamp TIMESTAMP)",
            "INSERT INTO public.test VALUES (1, 'one', 1, true, 1.1, '2021-01-01', '2021-01-01 12:00:00')",
            "INSERT INTO public.test VALUES (2, 'two', 2, false, 2.2, '2021-02-02', '2021-02-02 13:00:00')",
            "UPDATE public.test SET name = 'three' WHERE id = 2",
            "DELETE FROM public.test WHERE id = 1",
            WAIT_FOR_CATCHUP,
        ],
        query: "/* replica */ SELECT * FROM public.test order by id",
        expected: &[&["2", "three", "2", "f", "2.2", "2021-02-02", "2021-02-02 13:00:00"]],
    },
];
