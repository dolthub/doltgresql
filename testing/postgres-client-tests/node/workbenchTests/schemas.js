import {
  doltBranchFields,
  doltStatusFields,
  doltCommitFields,
  schemaNameField,
  pgTablesFields,
} from "../fields.js";
import { dbName } from "../helpers.js";

export const schemaTests = [
  {
    q: `SELECT DOLT_BRANCH($1::text, $2::text)`, // TODO: Should work without casts
    p: ["schemabranch", "main"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_branch: ["0"] }],
      fields: doltBranchFields,
    },
  },
  {
    q: `USE '${dbName}/schemabranch';`,
    res: {
      command: "SET",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: `create table testpub (pk int, "value" int, primary key(pk));`,
    res: {
      command: "CREATE",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: `SELECT DOLT_COMMIT('-Am', $1::text, '--author', $2::text);`,
    p: ["Create table testpub", "Dolt <dolt@dolthub.com>"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_commit: "" }],
      fields: doltCommitFields,
    },
  },
  {
    q: `CREATE SCHEMA testschema;`,
    res: {
      command: "CREATE",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: `SET SEARCH_PATH = 'testschema';`,
    res: {
      command: "SET",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: `create table test2 (pk int, "value" int, primary key(pk));`,
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
      rowCount: 2,
      oid: null,
      rows: [
        {
          table_name: "testschema.test2",
          staged: 0,
          status: "new table",
        },
        { table_name: "testschema", staged: 0, status: "new schema" },
      ],
      fields: doltStatusFields,
    },
  },
  {
    q: `SELECT schema_name FROM information_schema.schemata WHERE catalog_name = $1;`,
    p: [`${dbName}/schemabranch`],
    res: {
      command: "SELECT",
      rowCount: 5,
      oid: null,
      rows: [
        { schema_name: "dolt" },
        { schema_name: "pg_catalog" },
        { schema_name: "public" },
        { schema_name: "testschema" },
        { schema_name: "information_schema" },
      ],
      fields: [schemaNameField],
    },
  },
  {
    q: `SELECT DOLT_COMMIT('-Am', $1::text, '--author', $2::text);`,
    p: ["Create table test2", "Dolt <dolt@dolthub.com>"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_commit: "" }],
      fields: doltCommitFields,
    },
  },
  {
    q: `SELECT schemaname, tablename FROM pg_catalog.pg_tables where schemaname=$1;`,
    p: ["testschema"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ schemaname: "testschema", tablename: "test2" }],
      fields: pgTablesFields,
    },
  },
  {
    q: `SET SEARCH_PATH = 'public';`,
    res: {
      command: "SET",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: `SELECT schemaname, tablename FROM pg_catalog.pg_tables where schemaname=$1;`,
    p: ["public"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ schemaname: "public", tablename: "testpub" }],
      fields: pgTablesFields,
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
    q: `SELECT schema_name FROM information_schema.schemata WHERE catalog_name = $1;`,
    p: [`${dbName}/main`],
    res: {
      command: "SELECT",
      rowCount: 4,
      oid: null,
      rows: [
        { schema_name: "dolt" },
        { schema_name: "pg_catalog" },
        { schema_name: "public" },
        { schema_name: "information_schema" },
      ],
      fields: [schemaNameField],
    },
  },
  {
    q: `SELECT schemaname, tablename FROM pg_catalog.pg_tables where schemaname=$1;`,
    p: ["public"],
    res: {
      command: "SELECT",
      rowCount: 0,
      oid: null,
      rows: [],
      fields: pgTablesFields,
    },
  },
];
