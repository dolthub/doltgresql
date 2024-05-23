import { mysql as escapeQueryWithParameters } from "yesql";

const args = process.argv.slice(1);
const user = args[0];
const port = args[1];

export function getArgs() {
  return { user, port };
}

export function getConfig() {
  const { user, port } = getArgs();
  return {
    host: "127.0.0.1",
    port: port,
    user: user,
    database: "doltgres",
  };
}

export function assertQueryResult(q, resultStr, expected, rows, matcher) {
  if (matcher) {
    return matcher(rows, expected);
  }
  if (q.toLowerCase().includes("dolt_commit")) {
    return rows.length === 1 && rows[0].hash.length === 32;
  }
  if (q.toLowerCase().includes("dolt_merge")) {
    const result = JSON.parse(resultStr);
    return (
      expected.fast_forward === result.fast_forward &&
      expected.conflicts === result.conflicts
    );
  }
  return resultStr === JSON.stringify(expected);
}

export function getQueryWithEscapedParameters(q, parameters) {
  return escapeQueryWithParameters(q)(parameters || {});
}
