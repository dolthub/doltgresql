import pkg from "pg";
import { getConfig } from "./helpers.js";

const { Client } = pkg;

// Tests that Doltgres reports errors with the correct SQLSTATE and that the
// connection remains usable afterwards. node-postgres exposes the SQLSTATE as
// `err.code`, which is what applications and libraries dispatch on.

async function newClient() {
  const client = new Client(getConfig());
  await client.connect();
  return client;
}

async function expectCode(client, sql, wantCode, label) {
  let err;
  try {
    await client.query(sql);
  } catch (e) {
    err = e;
  }
  if (err === undefined) {
    throw new Error(`${label}: expected SQLSTATE ${wantCode}, but the query succeeded`);
  }
  if (err.code !== wantCode) {
    throw new Error(
      `${label}: expected SQLSTATE ${wantCode}, got ${err.code} (${err.message})`
    );
  }
}

async function main() {
  const client = await newClient();
  await client.query("DROP TABLE IF EXISTS err_test");
  await client.query(
    "CREATE TABLE err_test (id INT PRIMARY KEY, val INT NOT NULL CHECK (val >= 0))"
  );
  await client.query("INSERT INTO err_test VALUES (1, 1)");

  await expectCode(client, "INSERT INTO err_test VALUES (1, 1)", "23505", "unique violation");
  await expectCode(client, "INSERT INTO err_test VALUES (2, NULL)", "23502", "not-null violation");
  await expectCode(client, "INSERT INTO err_test VALUES (2, -1)", "23514", "check violation");
  await expectCode(client, "SELECT * FROM no_such_table", "42P01", "undefined table");
  await expectCode(client, "SELECT no_such_col FROM err_test", "42703", "undefined column");
  await expectCode(client, "SELECT no_such_function(1)", "42883", "undefined function");
  await expectCode(client, "SELECT 1/0", "22012", "division by zero");
  await expectCode(client, "SELEC 1", "42601", "syntax error");
  await expectCode(client, "SELECT 'abc'::int4", "22P02", "invalid text representation");

  // The connection must have survived every error above.
  const res = await client.query("SELECT count(*)::int4 AS n FROM err_test");
  if (res.rows[0].n !== 1) {
    throw new Error(`expected 1 row after failed statements, got ${res.rows[0].n}`);
  }

  // Conflicting concurrent transactions: the losing COMMIT must report 40001
  // (serialization_failure), the code retry logic dispatches on.
  const client2 = await newClient();
  await client.query("BEGIN");
  await client.query("UPDATE err_test SET val = 10 WHERE id = 1");
  await client2.query("UPDATE err_test SET val = 20 WHERE id = 1");
  await expectCode(client, "COMMIT", "40001", "serialization failure");

  // ... and the session must be usable again immediately.
  await client.query("UPDATE err_test SET val = 30 WHERE id = 1");
  const after = await client.query("SELECT val FROM err_test WHERE id = 1");
  if (after.rows[0].val !== 30) {
    throw new Error(`expected val=30 after recovery, got ${after.rows[0].val}`);
  }

  await client.end();
  await client2.end();
  console.log("node error-code tests passed");
  process.exit(0);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
