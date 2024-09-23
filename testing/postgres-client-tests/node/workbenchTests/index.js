import { branchTests } from "./branches.js";
import { databaseTests } from "./databases.js";
import { logTests } from "./logs.js";
import { assertEqualRows } from "../helpers.js";
import { mergeTests } from "./merge.js";
import { tableTests } from "./table.js";

export default async function runWorkbenchTests(database) {
  await runTests(database, databaseTests, "database");
  await runTests(database, branchTests, "branches");
  await runTests(database, logTests, "logs");
  await runTests(database, mergeTests, "merge");
  await runTests(database, tableTests, "tables");
  // TODO: Move over the rest of the Dolt workbench tests
}

async function runTests(database, tests, name) {
  console.log("Running tests for", name);
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
