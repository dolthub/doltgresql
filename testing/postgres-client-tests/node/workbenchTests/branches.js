import {
  countFields,
  doltAddFields,
  doltBranchFields,
  doltCheckoutFields,
  doltStatusFields,
} from "../fields.js";
import { branchesMatcher } from "./matchers.js";

export const branchTests = [
  {
    q: `SELECT DOLT_BRANCH($1::text, $2::text)`, // TODO: Should work without casts
    p: ["mybranch", "main"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_branch: "{0}" }],
      fields: doltBranchFields,
    },
  },
  {
    q: `SELECT DOLT_CHECKOUT($1::text)`, // TODO: Should work without casts
    p: ["mybranch"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_checkout: `{0,"Switched to branch 'mybranch'"}` }],
      fields: doltCheckoutFields,
    },
  },
  // TODO: 'public' schema should be accessible by this point. It cannot be
  // created because it already exists, but creating a table fails with a "no
  // schema has been selected to create in" error.
  {
    q: "CREATE SCHEMA test",
    res: { command: "CREATE", rowCount: null, oid: null, rows: [], fields: [] },
  },
  {
    q: "SET search_path = 'test';",
    res: { command: "SET", rowCount: null, oid: null, rows: [], fields: [] },
  },
  {
    q: `create table test (pk int, "value" int, primary key(pk));`,
    res: {
      command: "CREATE",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: `SELECT * FROM dolt_status;`,
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          table_name: "test",
          staged: 0,
          status: "new table",
        },
      ],
      fields: doltStatusFields,
    },
  },
  {
    q: `SELECT DOLT_ADD('-A');`, // TODO: Why does dolt_add not work?
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_add: `{0}` }],
      fields: doltAddFields,
    },
  },
  {
    q: `SELECT * FROM dolt_status;`,
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          table_name: "test",
          staged: 1,
          status: "new table",
        },
      ],
      fields: doltStatusFields,
    },
  },
  {
    q: `SELECT DOLT_COMMIT('-m', $1::text, '--author', $2::text);`,
    p: ["Create table test", "Dolt <dolt@dolthub.com>"],
    res: [{ hash: "" }],
  },
  {
    q: `SELECT * FROM dolt_branches LIMIT 200`,
    res: {
      rows: [
        {
          name: "main",
          hash: "",
          latest_committer: "mysql-test-runner",
          latest_committer_email: "mysql-test-runner@liquidata.co",
          latest_commit_date: "",
          latest_commit_message: "Initialize data repository",
          remote: "",
          branch: "",
        },
        {
          name: "mybranch",
          hash: "",
          latest_committer: "Dolt",
          latest_committer_email: "dolt@dolthub.com",
          latest_commit_date: "",
          latest_commit_message: "Create table test",
          remote: "",
          branch: "",
        },
      ],
    },
    matcher: branchesMatcher,
  },
  {
    q: `SELECT DOLT_CHECKOUT('-b', $1::text)`,
    p: ["branch-to-delete"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_checkout: `{0,"Switched to branch 'branch-to-delete'"}` }],
      fields: doltCheckoutFields,
    },
  },
  {
    q: `SELECT COUNT(*) FROM dolt_branches LIMIT 200`,
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ ["COUNT(*)"]: 3 }],
      fields: countFields,
    },
  },
  {
    q: `SELECT dolt_checkout($1::text)`,
    p: ["main"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_checkout: `{0,"Switched to branch 'main'"}` }],
      fields: doltCheckoutFields,
    },
  },
  {
    q: `SELECT DOLT_BRANCH('-D', $1)`,
    p: ["branch-to-delete"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_branch: "{0}" }],
      fields: doltBranchFields,
    },
  },
  {
    q: `SELECT COUNT(*) FROM dolt_branches LIMIT 200`,
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ ["COUNT(*)"]: 2 }],
      fields: countFields,
    },
  },
];
