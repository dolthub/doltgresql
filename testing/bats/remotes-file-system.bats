#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
}

teardown() {
    teardown_common
}

# This test exercises the push/pull/clone workflow across two independent Doltgres server
# processes, each with its own data directory, coordinating only through a shared
# file-system remote.
@test 'remotes-file-system: clone from a fresh server, then push/pull round-trip data, a sequence, an enum type, and a function' {
    mkdir remote
    REMOTE_URL="file://$(pwd)/remote"

    # --- Server A: seed a table, a sequence, a custom enum type, and a user-defined function --
    # (all serialized at the Doltgres layer, not the Dolt layer -- see core/rootobject) -- commit,
    # and push to the file-system remote ---
    mkdir serverA
    cd serverA
    start_sql_server
    query_server <<SQL
CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy');
CREATE FUNCTION double_it(x INT) RETURNS INT AS \$\$ BEGIN RETURN x * 2; END; \$\$ LANGUAGE plpgsql;
CREATE TABLE items (id INT PRIMARY KEY, label TEXT NOT NULL, feeling mood);
INSERT INTO items VALUES (1, 'apple', 'happy'), (2, 'banana', 'sad');
CREATE SEQUENCE counter START 100 INCREMENT 50;
SELECT nextval('counter'); -- advances to 100
SELECT dolt_commit('-Am', 'seed items, counter, mood type, and double_it function');
SELECT dolt_remote('add', 'origin', '$REMOTE_URL');
SELECT dolt_push('origin', 'main');
SQL
    stop_sql_server
    cd ..

    # --- Server B: an entirely separate, freshly-started server with its own data directory;
    # it shares no process, session, or in-memory state with server A. Clone from the remote. ---
    mkdir serverB
    cd serverB
    start_sql_server
    query_server -c "SELECT dolt_clone('$REMOTE_URL', 'cloned');"

    # The enum type definition itself (not just data using it) must have transferred. Checked as
    # separate substrings (rather than one "id | label | feeling" string) because psql's tuples-only
    # output pads each column to its widest value ("apple" pads out to match "banana"'s width), so
    # the exact spacing between columns isn't stable across rows.
    run query_server_for_db cloned -t -c "SELECT id, label, feeling::text FROM items ORDER BY id;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1 | apple" ]] || false
    [[ "$output" =~ "happy" ]] || false
    [[ "$output" =~ "2 | banana" ]] || false
    [[ "$output" =~ "sad" ]] || false

    # The user-defined function must also be callable post-clone.
    run query_server_for_db cloned -c "SELECT double_it(21);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "42" ]] || false

    # Must reflect server A's current value (100), not reset to the sequence's start value. Read
    # via pg_sequences rather than calling nextval() here, since nextval() itself writes the
    # sequence's current value and would leave the clone with an uncommitted change, which the
    # pull below would then reject with "cannot merge with uncommitted changes".
    run query_server_for_db cloned -c "SELECT last_value FROM pg_sequences WHERE sequencename = 'counter';"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "100" ]] || false

    stop_sql_server
    cd ..

    # --- Back on server A: advance the table (using the enum type again) and the sequence
    # further, and push again ---
    cd serverA
    start_sql_server
    query_server <<SQL
INSERT INTO items VALUES (3, 'cherry', 'ok');
SELECT nextval('counter'); -- advances to 150
SELECT dolt_commit('-Am', 'add cherry and advance counter');
SELECT dolt_push('origin', 'main');
SQL
    stop_sql_server
    cd ..

    # --- Back on server B (also restarted fresh): pull the update into the clone ---
    cd serverB
    start_sql_server
    run query_server_for_db cloned -c "SELECT dolt_pull('origin');"
    [ "$status" -eq 0 ]

    run query_server_for_db cloned -t -c "SELECT id, label, feeling::text FROM items ORDER BY id;"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "1 | apple" ]] || false
    [[ "$output" =~ "happy" ]] || false
    [[ "$output" =~ "2 | banana" ]] || false
    [[ "$output" =~ "sad" ]] || false
    [[ "$output" =~ "3 | cherry" ]] || false
    [[ "$output" =~ "ok" ]] || false

    # The function keeps working after an incremental pull too, not just right after clone.
    run query_server_for_db cloned -c "SELECT double_it(10);"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "20" ]] || false

    # The pull must have carried server A's advanced current value (150), not left the clone's own.
    run query_server_for_db cloned -c "SELECT last_value FROM pg_sequences WHERE sequencename = 'counter';"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "150" ]] || false

    # No further pull happens after this, so it's now safe to call nextval() directly: must
    # continue from 150 (next is 200), proving the synced state drives future values correctly too.
    run query_server_for_db cloned -c "SELECT nextval('counter');"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "200" ]] || false

    stop_sql_server
    cd ..
}
