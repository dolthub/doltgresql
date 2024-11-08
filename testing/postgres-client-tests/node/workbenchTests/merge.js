import { doltStatusFields } from "../fields.js";
import { logsMatcher, mergeBaseMatcher, mergeMatcher } from "./matchers.js";

export const mergeTests = [
  {
    q: `SELECT DOLT_MERGE_BASE($1::text, $2::text);`,
    p: ["mybranch", "main"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_merge_base: "" }],
      fields: [],
    },
    matcher: mergeBaseMatcher,
  },
  {
    q: `SELECT * FROM dolt.status`,
    res: {
      command: "SELECT",
      rowCount: 0,
      oid: null,
      rows: [],
      fields: doltStatusFields,
    },
  },
  {
    q: `SELECT DOLT_MERGE($1::text, '--no-ff', '-m', $2::text)`,
    p: ["mybranch", "Merge mybranch into main"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          dolt_merge: ["", "0", "0", "merge successful"],
        },
      ],
      fields: [],
    },
    matcher: mergeMatcher,
  },
  {
    q: `SELECT * FROM DOLT_LOG('main', '--parents') LIMIT 10 OFFSET 0;`, // TODO: Prepared not working
    res: {
      command: "SELECT",
      rowCount: 4,
      oid: null,
      rows: [
        {
          commit_hash: "",
          message: "Merge mybranch into main",
          committer: "postgres",
          email: "postgres@127.0.0.1",
          date: "",
          parents: ["", ""],
        },
        {
          commit_hash: "",
          committer: "Dolt",
          email: "dolt@dolthub.com",
          date: "",
          message: "Create table test",
          parents: [""],
        },
        {
          commit_hash: "",
          committer: "postgres",
          email: "postgres@127.0.0.1",
          date: "",
          message: "CREATE DATABASE",
          parents: [""],
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
];
