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

mod common;

use std::sync::atomic::{AtomicI32, Ordering};
use std::time::Duration;

use common::*;
use driver::client::{Db, Value};
use driver::runner::Env;
use harness::pgx::Arg;
use rand::Rng;
use rand::seq::SliceRandom;

/// query_text returns the single text value of a query.
fn query_text(db: &mut Db, sql: &str) -> String {
    match query_row(db, sql, &[]).as_slice() {
        [Value::Text(v)] => v.clone(),
        other => panic!("{sql}: expected text, got {other:?}"),
    }
}

/// random_sleep sleeps between half a second and a second and a half.
fn random_sleep() {
    std::thread::sleep(Duration::from_millis(rand::thread_rng().gen_range(500..1500)));
}

/// run_racing_amend begins 200 transactions that each update the first row, insert a row, and amend the commit
/// in a random order, checking that exactly one of them wins.
fn run_racing_amend(mut env: Env, insert: &str, insert_args: impl Fn(i64) -> Vec<Arg> + Sync) {
    let mut db = db(&mut env);
    exec(&mut db, "INSERT INTO test_table VALUES (1, 'initial')", &[]);
    exec(&mut db, "SELECT dolt_commit('-A','-m', 'initial commit')", &[]);

    let mut transactions: Vec<(i64, Db)> = (1..=200)
        .map(|tx_num| {
            let mut tx = common::db(&mut env);
            tx.begin().unwrap();
            (tx_num, tx)
        })
        .collect();
    transactions.shuffle(&mut rand::thread_rng());

    let winner = AtomicI32::new(-1);
    let results: Vec<Result<(), String>> = std::thread::scope(|scope| {
        let handles: Vec<_> = transactions
            .iter_mut()
            .map(|(tx_num, tx)| {
                let (tx_num, winner, insert_args) = (*tx_num, &winner, &insert_args);
                scope.spawn(move || -> Result<(), String> {
                    random_sleep();
                    let update = [Arg::Str(format!("tx{tx_num} value"))];
                    tx.exec_args("UPDATE test_table SET value = $1 WHERE id = 1", &update)?;
                    tx.exec_args(insert, &insert_args(tx_num))?;
                    random_sleep();
                    let amend = [Arg::Str(format!("tx{tx_num} amend"))];
                    if tx.exec_args("SELECT dolt_commit('--amend','-a', '-m', $1)", &amend).is_err() {
                        let _ = tx.rollback();
                    } else if winner.compare_exchange(-1, tx_num as i32, Ordering::SeqCst, Ordering::SeqCst).is_err() {
                        return Err(format!("tx{tx_num} also committed"));
                    }
                    Ok(())
                })
            })
            .collect();
        handles.into_iter().map(|h| h.join().unwrap()).collect()
    });
    for result in results {
        result.unwrap();
    }

    let winner = winner.load(Ordering::SeqCst);
    assert_ne!(winner, -1);
    assert_eq!(query_int(&mut db, "SELECT COUNT(*) FROM test_table", &[]), 2);
    assert_eq!(query_text(&mut db, "SELECT value FROM test_table WHERE id = 1"), format!("tx{winner} value"));
    assert_eq!(query_text(&mut db, "SELECT value FROM test_table WHERE id != 1"), format!("tx{winner} new row"));
    let message = query_text(&mut db, "SELECT message FROM dolt_log ORDER BY date DESC LIMIT 1");
    assert_eq!(message, format!("tx{winner} amend"));
    drop(transactions);
    finish(env);
}

#[test]
fn test_commit_concurrency_sql_transaction_with_amend_commit() {
    let mut env = env();
    setup_test_server(&mut env, "commit_concurrency_test");
    let mut db = db(&mut env);
    exec(&mut db, "\nCREATE TABLE test_table (\n  id serial PRIMARY KEY,\n  value VARCHAR(20)\n);", &[]);
    exec(&mut db, "INSERT INTO test_table (value) VALUES ('initial')", &[]);
    exec(&mut db, "SELECT dolt_commit('-A','-m', 'initial commit')", &[]);

    let mut tx1 = common::db(&mut env);
    tx1.begin().unwrap();
    exec(&mut tx1, "UPDATE test_table SET value = 'amended by tx1' WHERE id = 1", &[]);

    let mut tx2 = common::db(&mut env);
    tx2.begin().unwrap();
    exec(&mut tx2, "UPDATE test_table SET value = 'amended by tx2' WHERE id = 1", &[]);
    exec(&mut tx2, "SELECT dolt_commit('--amend', '-m', 'tx2 amended commit')", &[]);

    exec(&mut tx1, "INSERT INTO test_table (value) VALUES ('new row by tx1')", &[]);
    let err = tx1.exec("SELECT dolt_commit('--amend', '-m', 'should fail')", &[]).unwrap_err();
    let expected =
        "this transaction conflicts with a committed transaction from another client, try restarting transaction";
    assert!(err.contains(expected), "{err}");

    assert_eq!(query_text(&mut db, "SELECT value FROM test_table WHERE id = 1"), "amended by tx2");
    let message = query_text(&mut db, "SELECT message FROM dolt_log ORDER BY date DESC LIMIT 1");
    assert_eq!(message, "tx2 amended commit");
    drop((tx1, tx2));
    finish(env);
}

#[test]
fn test_commit_concurrency_sql_racing_amend() {
    let mut env = env();
    setup_test_server(&mut env, "racing_amend_test");
    let mut db = db(&mut env);
    exec(&mut db, "\n\tCREATE TABLE test_table (\n\t  id bigint PRIMARY KEY,\n\t  value VARCHAR(20)\n\t);", &[]);
    db.close();
    run_racing_amend(env, "INSERT INTO test_table (id, value) VALUES ($1, $2)", |tx_num| {
        vec![Arg::Int(tx_num + 1), Arg::Str(format!("tx{tx_num} new row"))]
    });
}

#[test]
#[ignore = "concurrent inserts on the same sequence fail with duplicate keys"]
fn test_commit_concurrency_sql_racing_amend_with_serial_column() {
    let mut env = env();
    setup_test_server(&mut env, "racing_amend_test");
    let mut db = db(&mut env);
    exec(&mut db, "\n\tCREATE TABLE test_table (\n\t  id serial PRIMARY KEY,\n\t  value VARCHAR(20)\n\t);", &[]);
    db.close();
    run_racing_amend(env, "INSERT INTO test_table (value) VALUES ($1)", |tx_num| {
        vec![Arg::Str(format!("tx{tx_num} new row"))]
    });
}
