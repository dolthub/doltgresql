import {
  doltCheckoutFields,
  doltSchemasFields,
  pgTablesFields,
} from "../fields.js";

export const viewsTests = [
  {
    q: "SELECT DOLT_CHECKOUT('-b', $1::text);",
    p: ["more-updates"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_checkout: ["0", "Switched to branch 'more-updates'"] }],
      fields: doltCheckoutFields,
    },
  },
  {
    q: "SELECT * FROM dolt_schemas LIMIT 10 OFFSET 0;",
    res: {
      command: "SELECT",
      rowCount: 0,
      oid: null,
      rows: [],
      fields: doltSchemasFields,
    },
  },
  {
    q: "CREATE VIEW myview AS SELECT * FROM test;",
    res: {
      command: "CREATE",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    skip: true, // TODO: Not returning anything
    q: "SELECT * FROM dolt_schemas LIMIT 10 OFFSET 0",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          type: "view",
          name: "myview",
          fragment: "CREATE VIEW myview AS SELECT * FROM test",
          extra: { CreatedAt: 0 },
          sql_mode:
            "NO_ENGINE_SUBSTITUTION,ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES",
        },
      ],
      fields: doltSchemasFields,
    },
  },
  {
    // Excludes views
    q: "SELECT schemaname, tablename FROM pg_catalog.pg_tables where schemaname=$1;",
    p: ["public"],
    res: {
      command: "SELECT",
      rowCount: 2,
      oid: null,
      rows: [
        {
          schemaname: "public",
          tablename: "test",
        },
        {
          schemaname: "public",
          tablename: "test_info",
        },
      ],
      fields: pgTablesFields,
    },
  },
  {
    // Includes views
    q: "SELECT table_name FROM INFORMATION_SCHEMA.views WHERE table_schema = $1;",
    p: ["public"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ table_name: "myview" }],
      fields: [
        {
          name: "table_name",
          tableID: 0,
          columnID: 0,
          dataTypeID: 1043,
          dataTypeSize: 256,
          dataTypeModifier: -1,
          format: "text",
        },
      ],
    },
  },
  {
    // Includes views
    q: "SELECT viewname FROM pg_catalog.pg_views WHERE schemaname=$1;",
    p: ["public"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ viewname: "myview" }],
      fields: [
        {
          name: "viewname",
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
    // Includes tables and views
    q: "SELECT table_schema, table_name, table_type FROM INFORMATION_SCHEMA.tables WHERE table_schema = $1;",
    p: ["public"],
    res: {
      command: "SELECT",
      rowCount: 3,
      oid: null,
      rows: [
        { table_schema: "public", table_name: "myview", table_type: "VIEW" },
        {
          table_schema: "public",
          table_name: "test",
          table_type: "BASE TABLE",
        },
        {
          table_schema: "public",
          table_name: "test_info",
          table_type: "BASE TABLE",
        },
      ],
      fields: [
        {
          name: "table_schema",
          tableID: 0,
          columnID: 0,
          dataTypeID: 1043,
          dataTypeSize: 256,
          dataTypeModifier: -1,
          format: "text",
        },
        {
          name: "table_name",
          tableID: 0,
          columnID: 0,
          dataTypeID: 1043,
          dataTypeSize: 256,
          dataTypeModifier: -1,
          format: "text",
        },
        {
          name: "table_type",
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
    q: `SELECT pg_get_viewdef($1::regclass, true)`,
    p: ["myview"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ pg_get_viewdef: "SELECT * FROM test" }],
      fields: [
        {
          name: "pg_get_viewdef",
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
];
