import { branchTests } from "./branches.js";
import { databaseTests } from "./databases.js";
import { assertQueryResult } from "../helpers.js";

export default async function runWorkbenchTests(database) {
  await runTests(database, databaseTests);
  await runTests(database, branchTests);
  // await runTests(database, logTests);
  // await runTests(database, mergeTests);
  // await runTests(database, tableTests);
  // await runTests(database, docsTests);
  // await runTests(database, tagsTests);
  // await runTests(database, viewsTests);
  // await runTests(database, diffTests);
}

async function runTests(database, tests) {
  await Promise.all(
    tests.map((test) => {
      if (test.skip) return;

      return database
        .query(test.q, test.p)
        .then((data) => {
          assertEqualRows(test, data);
        })
        .catch((err) => {
          if (test.expectedErr) {
            if (err.message.includes(test.expectedErr)) {
              return;
            } else {
              console.log("Query error did not match expected:", test.q);
            }
          } else {
            console.log("Query errored:", test.q);
          }
          console.error(err);
          process.exit(1);
        });
    })
  );
}

function assertEqualRows(test, data) {
  const expected = test.res;
  const resultStr = JSON.stringify(data);
  const result = JSON.parse(resultStr);
  if (!assertQueryResult(test.q, resultStr, expected, data, test.matcher)) {
    console.log("Query:", test.q);
    console.log("Results:", result);
    console.log("Expected:", expected);
    throw new Error("Query failed");
  }
}
