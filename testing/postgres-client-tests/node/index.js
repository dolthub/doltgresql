import { Database } from "./database.js";
import { assertQueryResult, getConfig } from "./helpers.js";

const tests = [
  {
    q: "create table test (pk int, value int, primary key(pk))",
    res : {
      command: 'CREATE',
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: "select * from test",
    res: {
      command: 'SELECT',
      rowCount: 0,
      oid: null,
      rows: [],
      fields: [
        {
          name: 'pk',
          tableID: 0, // TODO: need to be filled? Got 16859 from Postgres
          columnID: 0, // TODO: need to be filled? Got 1 from Postgres
          dataTypeID: 23,
          dataTypeSize: 4,
          dataTypeModifier: -1,
          format: 'text'
        },
        {
          name: 'value',
          tableID: 0, // TODO: need to be filled? Got 16859 from Postgres
          columnID: 0, // TODO: need to be filled? Got 2 from Postgres
          dataTypeID: 23,
          dataTypeSize: 4,
          dataTypeModifier: -1,
          format: 'text'
        }
      ],
    }
  },
  {
    q: "insert into test (pk, value) values (0,0)",
    res: {
      command: 'INSERT',
      rowCount: 1,
      oid: 0,
      rows: [],
      fields: [],
    },
  },
  {
    q: "select * from test",
    res: {
      command: 'SELECT',
      rowCount: 1,
      oid: null,
      rows: [{ pk: 0, value: 0 }],
      fields: [
        {
          name: 'pk',
          tableID: 0,
          columnID: 0,
          dataTypeID: 23,
          dataTypeSize: 4,
          dataTypeModifier: -1,
          format: 'text'
        },
        {
          name: 'value',
          tableID: 0,
          columnID: 0,
          dataTypeID: 23,
          dataTypeSize: 4,
          dataTypeModifier: -1,
          format: 'text'
        }
      ],
    }
  },
  {
    q: "call dolt_add('-A');",
    res: {
      command: 'CALL',
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    }
  },
  {
    q: "call dolt_commit('-m', 'my commit')",
    res: {
      command: 'CALL',
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    }
  },
  {
    q: "select COUNT(*) FROM dolt_log",
    res: {
      command: 'SELECT',
      rowCount: 1,
      oid: null,
      rows: [ { 'count(*)': '2' } ], // TODO: Postgres returns 'count' as column name
      fields: [
        {
          name: 'count(*)', // TODO: Postgres returns 'count' as column name
          tableID: 0,
          columnID: 0,
          dataTypeID: 20,
          dataTypeSize: 8,
          dataTypeModifier: -1,
          format: 'text'
        }
      ],
    }
  },
  {
    q: "call dolt_checkout('-b', 'mybranch')",
    res: {
      command: 'CALL',
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    }
  },
  {
    q: "insert into test (pk, value) values (1,1),(2,3)",
    res: {
      command: 'INSERT',
      rowCount: 2,
      oid: 0,
      rows: [],
      fields: [],
    },
  },
  {
    q: "call dolt_commit('-a', '-m', 'my commit2')",
    res: {
      command: 'CALL',
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    }
  },
  {
    q: "call dolt_checkout('main')",
    res: {
      command: 'CALL',
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    }
  },
  {
    q: "call dolt_merge('mybranch')",
    res: {
      command: 'CALL',
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    }
  },
  {
    q: "select COUNT(*) FROM dolt_log",
    res: {
      command: 'SELECT',
      rowCount: 1,
      oid: null,
      rows: [ { 'count(*)': '3' } ],
      fields: [
        {
          name: 'count(*)',
          tableID: 0,
          columnID: 0,
          dataTypeID: 20,
          dataTypeSize: 8,
          dataTypeModifier: -1,
          format: 'text'
        }
      ],
    }
  },
];

async function main() {
  const database = new Database(getConfig());

  await Promise.all(
      tests.map((test) => {
        const expected = test.res;
        return database
            .query(test.q)
            .then((rows) => {
              const resultStr = JSON.stringify(rows);
              const result = JSON.parse(resultStr);
              if (!assertQueryResult(test.q, resultStr, expected, rows)) {
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
