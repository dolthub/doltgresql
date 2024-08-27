import { doltCheckoutFields } from "../fields.js";
import { dbName } from "../helpers.js";

export const databaseTests = [
  {
    q: `SELECT dolt_checkout($1::text);`, // TODO: All of these should work without type casts
    p: ["main"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_checkout: `{0,"Already on branch 'main'"}` }],
      fields: doltCheckoutFields,
    },
  },
  {
    q: `SELECT datname FROM pg_database;`,
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ datname: dbName }],
      fields: [
        {
          name: "datname",
          tableID: 0,
          columnID: 0,
          dataTypeID: 25,
          dataTypeSize: -1,
          dataTypeModifier: -1,
          format: "text",
        },
      ],
    },
  },
  {
    q: `CREATE DATABASE "new_db";`,
    res: {
      command: "CREATE",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: `SELECT datname FROM pg_database;`,
    res: {
      command: "SELECT",
      rowCount: 2,
      oid: null,
      rows: [{ datname: dbName }, { datname: "new_db" }],
      fields: [
        {
          name: "datname",
          tableID: 0,
          columnID: 0,
          dataTypeID: 25,
          dataTypeSize: -1,
          dataTypeModifier: -1,
          format: "text",
        },
      ],
    },
  },
  {
    q: `SELECT dolt_version();`,
    res: [{ "dolt_version()": "0.0.0" }],
    matcher: (res) => {
      return res.rows[0].dolt_version.length > 0;
    },
  },
];
