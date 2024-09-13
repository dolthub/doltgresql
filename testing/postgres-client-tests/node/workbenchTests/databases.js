import { dbName } from "../helpers.js";

export const databaseTests = [
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
    q: `SELECT datname FROM pg_database;`,
    res: {
      command: "SELECT",
      rowCount: 2,
      oid: null,
      rows: [{ datname: dbName }, { datname: `${dbName}/main` }],
      fields: [
        {
          name: "datname",
          tableID: 0,
          columnID: 0,
          dataTypeID: 19,
          dataTypeSize: 252,
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
      rowCount: 3,
      oid: null,
      rows: [
        { datname: dbName },
        { datname: `${dbName}/main` },
        { datname: "new_db" },
      ],
      fields: [
        {
          name: "datname",
          tableID: 0,
          columnID: 0,
          dataTypeID: 19,
          dataTypeSize: 252,
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
