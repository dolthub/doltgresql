import {
  countFields,
  doltBranchFields,
  doltCheckoutFields,
  doltStatusFields,
  doltCommitFields,
} from "../fields.js";
import { branchesMatcher } from "./matchers.js";
import { dbName } from "../helpers.js";

export const branchTests = [
  {
    q: `SELECT DOLT_BRANCH($1::text, $2::text)`, // TODO: Should work without casts
    p: ["mybranch", "main"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_branch: ["0"] }],
      fields: doltBranchFields,
    },
  },
  {
    q: `USE '${dbName}/mybranch';`,
    res: {
      command: "SET",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
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
    q: `SELECT * FROM dolt.status;`,
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          table_name: "public.test",
          staged: false,
          status: "new table",
        },
      ],
      fields: doltStatusFields,
    },
  },
  {
    q: `SELECT DOLT_COMMIT('-Am', $1::text, '--author', $2::text);`,
    p: ["Create table test", "Dolt <dolt@dolthub.com>"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_commit: "" }],
      fields: doltCommitFields,
    },
  },
  {
    q: `SELECT * FROM dolt.branches LIMIT 200`,
    res: {
      rows: [
        {
          name: "main",
          hash: "",
          latest_committer: "postgres",
          latest_committer_email: "postgres@127.0.0.1",
          latest_commit_date: "",
          latest_commit_message: "CREATE DATABASE",
          remote: "",
          branch: "",
          dirty: 1,
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
          dirty: 0,
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
      rows: [{ dolt_checkout: ["0", "Switched to branch 'branch-to-delete'"] }],
      fields: doltCheckoutFields,
    },
  },
  {
    q: `SELECT COUNT(*) FROM dolt.branches LIMIT 200`,
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ count: "3" }],
      fields: countFields,
    },
  },
  {
    q: `USE '${dbName}/main';`,
    res: {
      command: "SET",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: `SELECT DOLT_BRANCH('-D', $1::text)`,
    p: ["branch-to-delete"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_branch: ["0"] }],
      fields: doltBranchFields,
    },
  },
  {
    q: `SELECT COUNT(*) FROM dolt.branches LIMIT 200`,
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ count: "2" }],
      fields: countFields,
    },
  },
];
