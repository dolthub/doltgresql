import fs from "fs";
import { pipeline } from "stream/promises";
import { from as copyFrom } from "pg-copy-streams";
import path from "path";
import { branchTests } from "./branches.js";
import { databaseTests } from "./databases.js";
import { docsTests } from "./docs.js";
import { logTests } from "./logs.js";
import { assertEqualRows } from "../helpers.js";
import { mergeTests } from "./merge.js";
import { tableTests } from "./table.js";
import { schemaTests } from "./schemas.js";
import { tagsTests } from "./tags.js";
import { viewsTests } from "./views.js";
import { diffTests } from "./diffs.js";

const args = process.argv.slice(2);
const testDataPath = args[3];

export default async function runWorkbenchTests(database) {
  await runTests(database, databaseTests, "database");
  await runTests(database, branchTests, "branches");
  await runTests(database, schemaTests, "schemas");
  await runTests(database, logTests, "logs");
  await runTests(database, mergeTests, "merge");
  await runTests(database, tableTests, "tables");
  await runTests(database, docsTests, "docs");
  await runTests(database, tagsTests, "tags");
  await runTests(database, viewsTests, "views");
  await runTests(database, diffTests, "diffs");
}

async function runTests(database, tests, name) {
  console.log("Running tests for", name);
  await Promise.all(
    tests.map(async (test) => {
      if (test.skip) return;

      if (test.file) {
        const filePath = path.resolve(testDataPath, test.file);
        try {
          // TODO: Is it possible to test the COPY FROM output?
          const ingestStream = database.client.query(copyFrom(test.q));
          const sourceStream = fs.createReadStream(filePath);
          await pipeline([sourceStream, ingestStream]);
        } catch (err) {
          console.log("Query errored:", test.q);
          console.error(err);
          process.exit(1);
        }

        return;
      }

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
