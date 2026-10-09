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

use harness::oid::*;
use harness::pgx::Time;
use harness::plan::PlanFact;
use harness::script::Cell::{Any, Null, Text as T};
use harness::script::{A, BindVar, Column, Diagnostic, E, Expected, Flow, N, S, ScriptTest, ScriptTestAssertion, USER_DEFINED, run_scripts, run_scripts_repeated};
use harness::wire::{Datum, F, Field, Fields, PGX_STARTUP, Receive, Send, Step, W, WireTest, run_wire_tests};

#[test]
fn test_multiple_statements() {
    run_wire_tests(&[
        WireTest {
            name: "TestMultipleStatements",
            startup: PGX_STARTUP,
            steps: &[
                Step::Send(&[
                    Send::Query("-- ping"),
                ]),
                Step::Receive(&[
                    Receive::EmptyQueryResponse,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmtcache_6d0ade68bcc5d2976f68d06850d773cde3a5f1d760b45167", query: "BEGIN;DROP TABLE IF EXISTS migrations;DROP TABLE IF EXISTS animals;CREATE TABLE IF NOT EXISTS migrations (file_name TEXT NOT NULL, file_hash TEXT NOT NULL);CREATE TABLE IF NOT EXISTS animals (id SERIAL PRIMARY KEY NOT NULL, name TEXT NOT NULL);SELECT setval(pg_get_serial_sequence('animals', 'id'), 1);;INSERT INTO migrations (file_name, file_hash) VALUES ('2021-09-07T154500-create-animals-table.sql', '42331f4277227d09e9bb32eeaf7e04d9c7fe320160e05372ed0ef010cfbf666b');INSERT INTO animals(name) VALUES('Alpaca');INSERT INTO animals(name) VALUES('Highland cow');INSERT INTO animals(name) VALUES('Aardvark');INSERT INTO migrations (file_name, file_hash) VALUES ('2021-09-07T154700-insert-animals.sql', '3223d0deb6fb7fb2accf6abffc0667ebe4503379987c472d10a585a553f9b3b6');SELECT * FROM migrations ORDER BY file_name;SELECT * FROM animals ORDER BY id;COMMIT;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmtcache_6d0ade68bcc5d2976f68d06850d773cde3a5f1d760b45167"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42601", message: "cannot insert multiple commands into a prepared statement", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Close(b'S', "stmtcache_6d0ade68bcc5d2976f68d06850d773cde3a5f1d760b45167"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "BEGIN;DROP TABLE IF EXISTS migrations;DROP TABLE IF EXISTS animals;CREATE TABLE IF NOT EXISTS migrations (file_name TEXT NOT NULL, file_hash TEXT NOT NULL);CREATE TABLE IF NOT EXISTS animals (id SERIAL PRIMARY KEY NOT NULL, name TEXT NOT NULL);SELECT setval(pg_get_serial_sequence('animals', 'id'), 1);;INSERT INTO migrations (file_name, file_hash) VALUES ('2021-09-07T154500-create-animals-table.sql', '42331f4277227d09e9bb32eeaf7e04d9c7fe320160e05372ed0ef010cfbf666b');INSERT INTO animals(name) VALUES('Alpaca');INSERT INTO animals(name) VALUES('Highland cow');INSERT INTO animals(name) VALUES('Aardvark');INSERT INTO migrations (file_name, file_hash) VALUES ('2021-09-07T154700-insert-animals.sql', '3223d0deb6fb7fb2accf6abffc0667ebe4503379987c472d10a585a553f9b3b6');SELECT * FROM migrations ORDER BY file_name;SELECT * FROM animals ORDER BY id;COMMIT;", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42601", message: "cannot insert multiple commands into a prepared statement", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "BEGIN;DROP TABLE IF EXISTS migrations;DROP TABLE IF EXISTS animals;CREATE TABLE IF NOT EXISTS migrations (file_name TEXT NOT NULL, file_hash TEXT NOT NULL);CREATE TABLE IF NOT EXISTS animals (id SERIAL PRIMARY KEY NOT NULL, name TEXT NOT NULL);SELECT setval(pg_get_serial_sequence('animals', 'id'), 1);;INSERT INTO migrations (file_name, file_hash) VALUES ('2021-09-07T154500-create-animals-table.sql', '42331f4277227d09e9bb32eeaf7e04d9c7fe320160e05372ed0ef010cfbf666b');INSERT INTO animals(name) VALUES('Alpaca');INSERT INTO animals(name) VALUES('Highland cow');INSERT INTO animals(name) VALUES('Aardvark');INSERT INTO migrations (file_name, file_hash) VALUES ('2021-09-07T154700-insert-animals.sql', '3223d0deb6fb7fb2accf6abffc0667ebe4503379987c472d10a585a553f9b3b6');SELECT * FROM migrations ORDER BY file_name;SELECT * FROM animals ORDER BY id;COMMIT;", parameter_oids: &[] },
                    Send::Describe(b'S', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42601", message: "cannot insert multiple commands into a prepared statement", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "", query: "BEGIN;DROP TABLE IF EXISTS migrations;DROP TABLE IF EXISTS animals;CREATE TABLE IF NOT EXISTS migrations (file_name TEXT NOT NULL, file_hash TEXT NOT NULL);CREATE TABLE IF NOT EXISTS animals (id SERIAL PRIMARY KEY NOT NULL, name TEXT NOT NULL);SELECT setval(pg_get_serial_sequence('animals', 'id'), 1);;INSERT INTO migrations (file_name, file_hash) VALUES ('2021-09-07T154500-create-animals-table.sql', '42331f4277227d09e9bb32eeaf7e04d9c7fe320160e05372ed0ef010cfbf666b');INSERT INTO animals(name) VALUES('Alpaca');INSERT INTO animals(name) VALUES('Highland cow');INSERT INTO animals(name) VALUES('Aardvark');INSERT INTO migrations (file_name, file_hash) VALUES ('2021-09-07T154700-insert-animals.sql', '3223d0deb6fb7fb2accf6abffc0667ebe4503379987c472d10a585a553f9b3b6');SELECT * FROM migrations ORDER BY file_name;SELECT * FROM animals ORDER BY id;COMMIT;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "", parameter_formats: &[], parameters: &[], result_formats: &[] },
                    Send::Describe(b'P', ""),
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42601", message: "cannot insert multiple commands into a prepared statement", ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("BEGIN;DROP TABLE IF EXISTS migrations;DROP TABLE IF EXISTS animals;CREATE TABLE IF NOT EXISTS migrations (file_name TEXT NOT NULL, file_hash TEXT NOT NULL);CREATE TABLE IF NOT EXISTS animals (id SERIAL PRIMARY KEY NOT NULL, name TEXT NOT NULL);SELECT setval(pg_get_serial_sequence('animals', 'id'), 1);;INSERT INTO migrations (file_name, file_hash) VALUES ('2021-09-07T154500-create-animals-table.sql', '42331f4277227d09e9bb32eeaf7e04d9c7fe320160e05372ed0ef010cfbf666b');INSERT INTO animals(name) VALUES('Alpaca');INSERT INTO animals(name) VALUES('Highland cow');INSERT INTO animals(name) VALUES('Aardvark');INSERT INTO migrations (file_name, file_hash) VALUES ('2021-09-07T154700-insert-animals.sql', '3223d0deb6fb7fb2accf6abffc0667ebe4503379987c472d10a585a553f9b3b6');SELECT * FROM migrations ORDER BY file_name;SELECT * FROM animals ORDER BY id;COMMIT;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::Notice(Fields { severity: "NOTICE", severity_unlocalized: "NOTICE", code: "00000", message: r#"table "migrations" does not exist, skipping"#, ..F }),
                    Receive::CommandComplete("DROP TABLE"),
                    Receive::Notice(Fields { severity: "NOTICE", severity_unlocalized: "NOTICE", code: "00000", message: r#"table "animals" does not exist, skipping"#, ..F }),
                    Receive::CommandComplete("DROP TABLE"),
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::RowDescription(&[Field { name: "setval", attnum: 0, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("1")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::RowDescription(&[Field { name: "file_name", attnum: 1, type_oid: TEXT, size: -1, typmod: -1, format: 0 }, Field { name: "file_hash", attnum: 2, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("2021-09-07T154500-create-animals-table.sql"), Datum::Text("42331f4277227d09e9bb32eeaf7e04d9c7fe320160e05372ed0ef010cfbf666b")]),
                    Receive::DataRow(&[Datum::Text("2021-09-07T154700-insert-animals.sql"), Datum::Text("3223d0deb6fb7fb2accf6abffc0667ebe4503379987c472d10a585a553f9b3b6")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::RowDescription(&[Field { name: "id", attnum: 1, type_oid: INT4, size: 4, typmod: -1, format: 0 }, Field { name: "name", attnum: 2, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("2"), Datum::Text("Alpaca")]),
                    Receive::DataRow(&[Datum::Text("3"), Datum::Text("Highland cow")]),
                    Receive::DataRow(&[Datum::Text("4"), Datum::Text("Aardvark")]),
                    Receive::CommandComplete("SELECT 3"),
                    Receive::CommandComplete("COMMIT"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("BEGIN;DROP TABLE IF EXISTS migrations;DROP TABLE IF EXISTS animals;CREATE TABLE IF NOT EXISTS migrations (file_name TEXT NOT NULL, file_hash TEXT NOT NULL);CREATE TABLE IF NOT EXISTS animals (id SERIAL PRIMARY KEY NOT NULL, name TEXT NOT NULL);SELECT setval(pg_get_serial_sequence('animals', 'id'), 1);;INSERT INTO migrations (file_name, file_hash) VALUES ('2021-09-07T154500-create-animals-table.sql', '42331f4277227d09e9bb32eeaf7e04d9c7fe320160e05372ed0ef010cfbf666b');INSERT INTO animals(name) VALUES('Alpaca');INSERT INTO animals(name) VALUES('Highland cow');INSERT INTO animals(name) VALUES('Aardvark');INSERT INTO migrations (file_name, file_hash) VALUES ('2021-09-07T154700-insert-animals.sql', '3223d0deb6fb7fb2accf6abffc0667ebe4503379987c472d10a585a553f9b3b6');SELECT * FROM migrations ORDER BY file_name;SELECT * FROM animals ORDER BY id;COMMIT;"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("BEGIN"),
                    Receive::CommandComplete("DROP TABLE"),
                    Receive::CommandComplete("DROP TABLE"),
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::CommandComplete("CREATE TABLE"),
                    Receive::RowDescription(&[Field { name: "setval", attnum: 0, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("1")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::RowDescription(&[Field { name: "file_name", attnum: 1, type_oid: TEXT, size: -1, typmod: -1, format: 0 }, Field { name: "file_hash", attnum: 2, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("2021-09-07T154500-create-animals-table.sql"), Datum::Text("42331f4277227d09e9bb32eeaf7e04d9c7fe320160e05372ed0ef010cfbf666b")]),
                    Receive::DataRow(&[Datum::Text("2021-09-07T154700-insert-animals.sql"), Datum::Text("3223d0deb6fb7fb2accf6abffc0667ebe4503379987c472d10a585a553f9b3b6")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::RowDescription(&[Field { name: "id", attnum: 1, type_oid: INT4, size: 4, typmod: -1, format: 0 }, Field { name: "name", attnum: 2, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("2"), Datum::Text("Alpaca")]),
                    Receive::DataRow(&[Datum::Text("3"), Datum::Text("Highland cow")]),
                    Receive::DataRow(&[Datum::Text("4"), Datum::Text("Aardvark")]),
                    Receive::CommandComplete("SELECT 3"),
                    Receive::CommandComplete("COMMIT"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("INSERT INTO animals(name) VALUES('Pigeon');SELECT * FROM non_existent;INSERT INTO animals(name) VALUES('Elephant');"),
                ]),
                Step::Receive(&[
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::Error(Fields { severity: "ERROR", severity_unlocalized: "ERROR", code: "42P01", message: r#"relation "non_existent" does not exist"#, position: 58, ..F }),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}
