import { logsMatcher } from "./matchers.js";
import { dbName } from "../helpers.js";

export const logTests = [
  {
    q: `SELECT * FROM DOLT_LOG('main', '--parents') LIMIT 10 OFFSET 0;`, // TODO: Prepared not working
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          commit_hash: "",
          committer: "postgres",
          email: "postgres@127.0.0.1",
          date: "",
          message: "CREATE DATABASE",
          parents: ["3orrg69ou1loj2ph21guie3r2lf8bsab"],
        },
        {
          commit_hash: "",
          committer: "Dolt System Account",
          email: "doltuser@dolthub.com",
          date: "",
          message: "Initialize data repository",
          parents: [],
        },
      ],
      fields: [],
    },
    matcher: logsMatcher,
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
    q: `SELECT * FROM dolt.log;`, // TODO: If we decide to implement AS OF, use here instead of USE statement above and below
    res: {
      command: "SELECT",
      rowCount: 2,
      oid: null,
      rows: [
        {
          commit_hash: "",
          committer: "Dolt",
          email: "dolt@dolthub.com",
          date: "",
          message: "Create table test",
        },
        {
          commit_hash: "",
          committer: "postgres",
          email: "postgres@127.0.0.1",
          date: "",
          message: "CREATE DATABASE",
        },
        {
          commit_hash: "",
          committer: "Dolt System Account",
          email: "doltuser@dolthub.com",
          date: "",
          message: "Initialize data repository",
        },
      ],
    },
    matcher: logsMatcher,
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
    q: `SELECT * FROM DOLT_LOG('main..mybranch', '--parents')`, // TODO: Prepared not working
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          commit_hash: "",
          committer: "Dolt",
          email: "dolt@dolthub.com",
          date: "",
          message: "Create table test",
          parents: [""],
        },
      ],
      fields: [],
    },
    matcher: logsMatcher,
  },
];
