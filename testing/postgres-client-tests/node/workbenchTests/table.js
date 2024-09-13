import { dbName } from "../helpers.js";
import {
  doltAddFields,
  doltCommitFields,
  doltStatusFields,
  infoSchemaKeyColumnUsageFields,
} from "../fields.js";

export const tableTests = [
  {
    q: "INSERT INTO test VALUES (0, 0), (1, 1), (2,2)",
    res: {
      command: "INSERT",
      rowCount: 3,
      oid: 0,
      rows: [],
      fields: [],
    },
  },
  {
    skip: true, // TODO: Unique indexes not yet supported
    q: `CREATE UNIQUE INDEX test_idx ON test (pk, value)`,
    res: {
      command: "CREATE",
      rowCount: 3,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: `SELECT ordinal_position, column_name, udt_name as data_type, is_nullable, column_default FROM information_schema.columns WHERE table_catalog=$1 AND table_schema = $2 AND table_name = $3;`,
    p: [`${dbName}/main`, "public", "test"],
    res: {
      command: "SELECT",
      rowCount: 2,
      oid: null,
      rows: [
        {
          ordinal_position: 1,
          column_name: "pk",
          data_type: "int4",
          is_nullable: "NO",
          column_default: null,
        },
        {
          ordinal_position: 2,
          column_name: "value",
          data_type: "int4",
          is_nullable: "YES",
          column_default: null,
        },
      ],
      fields: [
        {
          name: "ordinal_position",
          tableID: 0,
          columnID: 0,
          dataTypeID: 23,
          dataTypeSize: 4,
          dataTypeModifier: -1,
          format: "text",
        },
        {
          name: "column_name",
          tableID: 0,
          columnID: 0,
          dataTypeID: 1043,
          dataTypeSize: -1,
          dataTypeModifier: 68,
          format: "text",
        },
        {
          name: "data_type",
          tableID: 0,
          columnID: 0,
          dataTypeID: 1043,
          dataTypeSize: -1,
          dataTypeModifier: 68,
          format: "text",
        },
        {
          name: "is_nullable",
          tableID: 0,
          columnID: 0,
          dataTypeID: 1043,
          dataTypeSize: -1,
          dataTypeModifier: 7,
          format: "text",
        },
        {
          name: "column_default",
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
    q: `SELECT DOLT_COMMIT('-A', '-m', $1::text)`,
    p: ["Add some rows and a column index"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_commit: "" }],
      fields: doltCommitFields,
    },
  },
  {
    skip: true, // TODO: GROUP BY not implemented
    q: `SELECT 
    table_name, index_name, comment, non_unique, GROUP_CONCAT(column_name ORDER BY seq_in_index) AS COLUMNS 
  FROM information_schema.statistics 
  WHERE table_catalog=$1 AND table_schema=$2 AND table_name=$3 AND index_name!='PRIMARY' 
  GROUP BY index_name;`,
    p: [`${dbName}/main`, "public", "test"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          TABLE_NAME: "test",
          INDEX_NAME: "test_idx",
          COMMENT: "",
          NON_UNIQUE: 0,
          COLUMNS: "pk,value",
        },
      ],
    },
  },
  {
    q: "CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk))",
    res: {
      command: "CREATE",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: "INSERT INTO test_info VALUES (1, 'info about test pk 0', 0)",
    res: {
      command: "INSERT",
      rowCount: 1,
      oid: 0,
      rows: [],
      fields: [],
    },
  },
  {
    q: `SELECT DOLT_COMMIT('-A', '-m', $1::text)`,
    p: ["Add test_info with foreign key"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_commit: "" }],
      fields: doltCommitFields,
    },
  },
  {
    q: `SELECT "table_schema", "table_name" FROM "information_schema"."tables" WHERE "table_schema" = $1 AND "table_catalog" = $2;`,
    p: ["public", `${dbName}/main`],
    res: {
      command: "SELECT",
      rowCount: 2,
      oid: null,
      rows: [
        { table_schema: "public", table_name: "test" },
        { table_schema: "public", table_name: "test_info" },
      ],
      fields: [
        {
          name: "table_schema",
          tableID: 0,
          columnID: 0,
          dataTypeID: 1043,
          dataTypeSize: -1,
          dataTypeModifier: 52,
          format: "text",
        },
        {
          name: "table_name",
          tableID: 0,
          columnID: 0,
          dataTypeID: 1043,
          dataTypeSize: -1,
          dataTypeModifier: 52,
          format: "text",
        },
      ],
    },
  },
  {
    q: `SELECT * FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE WHERE table_name=$1 AND table_schema=$2 AND table_catalog=$3 AND referenced_table_schema IS NOT NULL`,
    p: ["test_info", "public", `${dbName}/main`],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          CONSTRAINT_CATALOG: `${dbName}/main`,
          CONSTRAINT_SCHEMA: "public",
          CONSTRAINT_NAME: "test_info_ibfk_1",
          TABLE_CATALOG: `${dbName}/main`,
          TABLE_SCHEMA: "public",
          TABLE_NAME: "test_info",
          COLUMN_NAME: "test_pk",
          ORDINAL_POSITION: 1,
          POSITION_IN_UNIQUE_CONSTRAINT: 1,
          REFERENCED_TABLE_SCHEMA: dbName,
          REFERENCED_TABLE_NAME: "test",
          REFERENCED_COLUMN_NAME: "pk",
        },
      ],
      fields: infoSchemaKeyColumnUsageFields,
    },
  },
  {
    q: `SELECT * FROM "public"."test_info" "public.test_info" ORDER BY id ASC LIMIT 10;`,
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ id: 1, info: "info about test pk 0", test_pk: 0 }],
      fields: [
        {
          name: "id",
          tableID: 0,
          columnID: 0,
          dataTypeID: 23,
          dataTypeSize: 4,
          dataTypeModifier: -1,
          format: "text",
        },
        {
          name: "info",
          tableID: 0,
          columnID: 0,
          dataTypeID: 1043,
          dataTypeSize: -1,
          dataTypeModifier: 259,
          format: "text",
        },
        {
          name: "test_pk",
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
    q: `USE '${dbName}/main'`,
    res: {
      command: "SET",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },

  // TODO: File upload tests

  // Add and revert load data changes
  {
    q: `SELECT * FROM dolt_status`,
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ table_name: "test_info", staged: 0, status: "modified" }],
      fields: doltStatusFields,
    },
  },
  {
    q: "SELECT DOLT_ADD('.')",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_add: "{0}" }],
      fields: doltAddFields,
    },
  },
  {
    q: `SELECT * FROM dolt_status`,
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ table_name: "test_info", staged: 1, status: "modified" }],
      fields: doltStatusFields,
    },
  },
  // {
  //   q: "SELECT DOLT_RESET('test_info');",
  //   res: {
  //     command: "SELECT",
  //     rowCount: 1,
  //     oid: null,
  //     rows: [{ dolt_reset: "{0}" }],
  //     fields: doltResetFields,
  //   },
  // },
  // {
  //   q: `SELECT * FROM dolt_status`,
  //   res: {
  //     command: "SELECT",
  //     rowCount: 1,
  //     oid: null,
  //     rows: [{ table_name: "test_info", staged: 0, status: "modified" }],
  //     fields: doltStatusFields,
  //   },
  // },
  // {
  //   q: "SELECT DOLT_CHECKOUT('test_info')",
  //   res: {
  //     command: "SELECT",
  //     rowCount: 1,
  //     oid: null,
  //     rows: [{ dolt_checkout: `{0,""}` }],
  //     fields: doltCheckoutFields,
  //   },
  // },
  // {
  //   q: `SELECT * FROM dolt_status`,
  //   res: {
  //     command: "SELECT",
  //     rowCount: 0,
  //     oid: null,
  //     rows: [],
  //     fields: doltStatusFields,
  //   },
  // },
];
