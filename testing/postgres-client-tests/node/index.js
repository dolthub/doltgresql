import { Database } from "./database.js";
import { assertQueryResult, getConfig } from "./helpers.js";
import {
  doltAddFields,
  doltCheckoutFields,
  doltCommitFields,
  countFields,
} from "./fields.js";

const tests = [
  {
    q: "create table test (pk int, value int, primary key(pk))",
    res: {
      command: "CREATE",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: "select * from test",
    res: {
      command: "SELECT",
      rowCount: 0,
      oid: null,
      rows: [],
      fields: [
        {
          name: "pk",
          tableID: 0, // TODO: need to be filled? Got 16859 from Postgres
          columnID: 0, // TODO: need to be filled? Got 1 from Postgres
          dataTypeID: 23,
          dataTypeSize: 4,
          dataTypeModifier: -1,
          format: "text",
        },
        {
          name: "value",
          tableID: 0, // TODO: need to be filled? Got 16859 from Postgres
          columnID: 0, // TODO: need to be filled? Got 2 from Postgres
          dataTypeID: 23,
          dataTypeSize: 4,
          dataTypeModifier: -1,
          format: "text",
        },
      ],
    },
  },
  {
    q: "insert into test (pk, value) values (0,0)",
    res: {
      command: "INSERT",
      rowCount: 1,
      oid: 0,
      rows: [],
      fields: [],
    },
  },
  {
    q: "select * from test",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ pk: 0, value: 0 }],
      fields: [
        {
          name: "pk",
          tableID: 0,
          columnID: 0,
          dataTypeID: 23,
          dataTypeSize: 4,
          dataTypeModifier: -1,
          format: "text",
        },
        {
          name: "value",
          tableID: 0,
          columnID: 0,
          dataTypeID: 23,
          dataTypeSize: 4,
          dataTypeModifier: -1,
          format: "text",
        },
      ],
    },
  },
  {
    q: "select dolt_add('-A');",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_add: "{0}" }],
      fields: doltAddFields,
    },
  },
  {
    q: "select dolt_commit('-m', 'my commit')",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_commit: "" }],
      fields: doltCommitFields,
    },
  },
  {
    q: "select COUNT(*) FROM dolt_log",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ count: "3" }],
      fields: countFields,
    },
  },
  {
    q: "select dolt_checkout('-b', 'mybranch')",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_checkout: `{0,"Switched to branch 'mybranch'"}` }],
      fields: doltCheckoutFields,
    },
  },
  {
    q: "insert into test (pk, value) values (1,1),(2,3)",
    res: {
      command: "INSERT",
      rowCount: 2,
      oid: 0,
      rows: [],
      fields: [],
    },
  },
  {
    q: "select dolt_commit('-a', '-m', 'my commit2')",
    res: {
      command: "CALL",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: "select dolt_checkout('main')",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_checkout: `{0,"Switched to branch 'main'"}` }],
      fields: doltCheckoutFields,
    },
  },
  {
    q: "select dolt_merge('mybranch')",
    res: {
      fastForward: "1",
      conflicts: "0",
      message: `"merge successful"`,
    },
  },
  {
    q: "select COUNT(*) FROM dolt_log",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ count: "4" }],
      fields: countFields,
    },
  },
];

async function main() {
  const database = new Database(getConfig());

  await Promise.all(
    tests.map((test) => {
      const expected = test.res;
      return database
        .query(test.q)
        .then((data) => {
          const resultStr = JSON.stringify(data);
          const result = JSON.parse(resultStr);
          if (!assertQueryResult(test.q, resultStr, expected, data)) {
            console.log("Query:", test.q);
            console.log("Results:", result);
            console.log("Expected:", expected);
            throw new Error("Query failed");
          } else {
            console.log("Query succeeded:", test.q);
          }
        })
        .catch((err) => {
          console.error(err);
          process.exit(1);
        });
    })
  );

  database.close();
  process.exit(0);
}

main();
