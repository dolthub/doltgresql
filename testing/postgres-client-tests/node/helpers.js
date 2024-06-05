const args = process.argv.slice(2);
const user = args[0];
const port = args[1];
const version = args[2];

export function getArgs() {
  return { user, port };
}

export function getDoltgresVersion() {
  return version;
}

export function getConfig() {
  const { user, port } = getArgs();
  return {
    host: 'localhost',
    port: port,
    database: 'doltgres',
    user: user,
  };
}

export function assertQueryResult(q, resultStr, expected, rows, matcher) {
  if (matcher) {
    return matcher(rows, expected);
  }
  // TODO: figure out a way to get result from stored procedure calls to check these
  // if (q.toLowerCase().includes("dolt_commit")) {
  //   return rows.length === 1 && rows[0].hash.length === 32;
  // }
  // if (q.toLowerCase().includes("dolt_merge")) {
  //   const result = JSON.parse(resultStr);
  //   return (
  //     expected.fast_forward === result.fast_forward &&
  //     expected.conflicts === result.conflicts
  //   );
  // }

  // Does partial matching of actual and expected results.
  const partialRes = {
    command: rows.command,
    rowCount: rows.rowCount,
    oid: rows.oid,
    rows: rows.rows,
    fields: rows.fields,
  }
  return JSON.stringify(expected) === JSON.stringify(partialRes)
}
