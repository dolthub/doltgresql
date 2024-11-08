const args = process.argv.slice(2);
const user = args[0];
const port = args[1];
const version = args[2];

export const dbName = "doltgres";

export function getArgs() {
  return { user, port };
}

export function getDoltgresVersion() {
  return version;
}

export function getConfig() {
  const { user, port } = getArgs();
  return {
    host: "localhost",
    port: port,
    database: dbName,
    user: user,
    password: "password",
  };
}

export function assertEqualRows(test, data) {
  const expected = test.res;
  const resultStr = JSON.stringify(data);
  const result = JSON.parse(resultStr);
  if (!assertQueryResult(test.q, expected, data, test.matcher)) {
    console.log("Query:", test.q);
    console.log("Results:", result);
    console.log("Expected:", expected);
    throw new Error("Query failed");
  } else {
    console.log("Query succeeded:", test.q);
  }
}

export function assertQueryResult(q, expected, data, matcher) {
  if (matcher) {
    return matcher(data, expected);
  }
  if (q.toLowerCase().includes("dolt_commit")) {
    if (data.rows.length !== 1) return false;
    const hash = data.rows[0].dolt_commit[0];
    if (hash.length !== 32) {
      console.log("Invalid hash for dolt_commit:", hash);
      return false;
    }
    expected.rows[0].dolt_commit = data.rows[0].dolt_commit;
  }

  // Does partial matching of actual and expected results.
  const partialRes = {
    command: data.command,
    rowCount: data.rowCount,
    oid: data.oid,
    rows: data.rows,
    fields: data.fields,
  };
  return JSON.stringify(expected) === JSON.stringify(partialRes);
}
