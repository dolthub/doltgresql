import {
  doltCommitFields,
  doltDiffStatFields,
  doltDiffSummaryFields,
  doltSchemaDiffFields,
  doltStatusFields,
  pgTablesFields,
} from "../fields.js";
import { diffRowsMatcher, patchRowsMatcher } from "./matchers.js";

export const diffTests = [
  {
    q: "UPDATE test SET value=1 WHERE pk=0",
    res: {
      command: "UPDATE",
      rowCount: 1,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: "DROP TABLE test_info",
    res: {
      command: "DROP",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: `CREATE SCHEMA anotherschema;`,
    res: {
      command: "CREATE",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: `create table anotherschema.testanother (pk int, "value" int, primary key(pk));`,
    res: {
      command: "CREATE",
      rowCount: null,
      oid: null,
      rows: [],
      fields: [],
    },
  },
  {
    q: "INSERT INTO anotherschema.testanother VALUES (1, 2)",
    res: {
      command: "INSERT",
      rowCount: 1,
      oid: 0,
      rows: [],
      fields: [],
    },
  },
  {
    q: `SELECT * FROM dolt_status ORDER BY table_name;`,
    res: {
      command: "SELECT",
      rowCount: 5,
      oid: null,
      rows: [
        { table_name: "anotherschema", staged: 0, status: "new schema" },
        {
          table_name: "anotherschema.testanother",
          staged: 0,
          status: "new table",
        },
        { table_name: "public.dolt_schemas", staged: 0, status: "new table" },
        { table_name: "public.test", staged: 0, status: "modified" },
        { table_name: "public.test_info", staged: 0, status: "deleted" },
      ],
      fields: doltStatusFields,
    },
  },
  {
    q: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING') ORDER BY to_table_name;", // TODO: Prepared not working
    res: {
      command: "SELECT",
      rowCount: 4,
      oid: null,
      rows: [
        {
          from_table_name: "public.test_info",
          to_table_name: "",
          diff_type: "dropped",
          data_change: 1,
          schema_change: 1,
        },
        {
          from_table_name: "",
          to_table_name: "anotherschema.testanother",
          diff_type: "added",
          data_change: 1,
          schema_change: 1,
        },
        {
          from_table_name: "",
          to_table_name: "public.dolt_schemas",
          diff_type: "added",
          data_change: 1,
          schema_change: 1,
        },
        {
          from_table_name: "public.test",
          to_table_name: "public.test",
          diff_type: "modified",
          data_change: 1,
          schema_change: 0,
        },
      ],
      fields: doltDiffSummaryFields,
    },
  },
  {
    q: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING', 'test')",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          from_table_name: "public.test",
          to_table_name: "public.test",
          diff_type: "modified",
          data_change: 1,
          schema_change: 0,
        },
      ],
      fields: doltDiffSummaryFields,
    },
  },
  {
    // TODO: What if a table with same name but different schema exists in different schema?
    q: "SELECT * FROM dolt_diff_summary('HEAD', 'WORKING', 'testanother')",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          from_table_name: "",
          to_table_name: "anotherschema.testanother",
          diff_type: "added",
          data_change: 1,
          schema_change: 1,
        },
      ],
      fields: doltDiffSummaryFields,
    },
  },
  {
    q: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING') ORDER BY table_name",
    res: {
      command: "SELECT",
      rowCount: 4,
      oid: null,
      rows: [
        {
          table_name: "anotherschema.testanother",
          rows_unmodified: "0",
          rows_added: "1",
          rows_deleted: "0",
          rows_modified: "0",
          cells_added: "2",
          cells_deleted: "0",
          cells_modified: "0",
          old_row_count: "0",
          new_row_count: "1",
          old_cell_count: "0",
          new_cell_count: "2",
        },
        {
          table_name: "public.dolt_schemas",
          rows_unmodified: "0",
          rows_added: "1",
          rows_deleted: "0",
          rows_modified: "0",
          cells_added: "5",
          cells_deleted: "0",
          cells_modified: "0",
          old_row_count: "0",
          new_row_count: "1",
          old_cell_count: "0",
          new_cell_count: "5",
        },
        {
          table_name: "public.test",
          rows_unmodified: "2",
          rows_added: "0",
          rows_deleted: "0",
          rows_modified: "1",
          cells_added: "0",
          cells_deleted: "0",
          cells_modified: "1",
          old_row_count: "3",
          new_row_count: "3",
          old_cell_count: "6",
          new_cell_count: "6",
        },
        {
          table_name: "public.test_info",
          rows_unmodified: "0",
          rows_added: "0",
          rows_deleted: "1",
          rows_modified: "0",
          cells_added: "0",
          cells_deleted: "3",
          cells_modified: "0",
          old_row_count: "1",
          new_row_count: "0",
          old_cell_count: "3",
          new_cell_count: "0",
        },
      ],
      fields: doltDiffStatFields,
    },
  },
  {
    q: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING', 'test_info')",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          table_name: "public.test_info",
          rows_unmodified: "0",
          rows_added: "0",
          rows_deleted: "1",
          rows_modified: "0",
          cells_added: "0",
          cells_deleted: "3",
          cells_modified: "0",
          old_row_count: "1",
          new_row_count: "0",
          old_cell_count: "3",
          new_cell_count: "0",
        },
      ],
      fields: doltDiffStatFields,
    },
  },
  {
    q: "SELECT * FROM dolt_diff_stat('HEAD', 'WORKING', 'testanother')",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          table_name: "anotherschema.testanother",
          rows_unmodified: "0",
          rows_added: "1",
          rows_deleted: "0",
          rows_modified: "0",
          cells_added: "2",
          cells_deleted: "0",
          cells_modified: "0",
          old_row_count: "0",
          new_row_count: "1",
          old_cell_count: "0",
          new_cell_count: "2",
        },
      ],
      fields: doltDiffStatFields,
    },
  },
  {
    q: "SELECT * FROM DOLT_DIFF('HEAD', 'WORKING', 'test') ORDER BY to_pk ASC, from_pk ASC LIMIT 10 OFFSET 0",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          to_pk: 0,
          to_value: 1,
          to_commit: "WORKING",
          to_commit_date: "2023-03-09T07:44:47.670Z",
          from_pk: 0,
          from_value: 0,
          from_commit: "HEAD",
          from_commit_date: "2023-03-09T07:44:47.488Z",
          diff_type: "modified",
        },
      ],
      fields: [],
    },
    matcher: diffRowsMatcher,
  },
  {
    q: "SELECT * FROM DOLT_DIFF('HEAD', 'WORKING', 'test_info') ORDER BY to_id ASC, from_id ASC LIMIT 10 OFFSET 0",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          to_id: null,
          to_info: null,
          to_test_pk: null,
          to_commit: "WORKING",
          to_commit_date: "2023-03-09T07:53:48.614Z",
          from_id: 1,
          from_info: "info about test pk 0",
          from_test_pk: 0,
          from_commit: "HEAD",
          from_commit_date: "2023-03-09T07:53:48.284Z",
          diff_type: "removed",
        },
      ],
      fields: [],
    },
    matcher: diffRowsMatcher,
  },
  {
    q: "SELECT * FROM DOLT_DIFF('HEAD', 'WORKING', 'testanother') ORDER BY to_pk ASC, from_pk ASC LIMIT 10 OFFSET 0",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          to_pk: 1,
          to_value: 2,
          to_commit: "WORKING",
          to_commit_date: "2024-10-03T04:33:43.486Z",
          from_pk: null,
          from_value: null,
          from_commit: "HEAD",
          from_commit_date: "2024-10-03T04:33:43.430Z",
          diff_type: "added",
        },
      ],
      fields: [],
    },
    matcher: diffRowsMatcher,
  },
  {
    q: "SELECT * FROM DOLT_DIFF('HEAD', 'WORKING', 'dolt_schemas') ORDER BY to_name ASC, from_name ASC LIMIT 10 OFFSET 0",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          to_type: "view",
          to_name: "myview",
          to_fragment: "CREATE VIEW myview AS SELECT * FROM test",
          to_extra: { CreatedAt: 0 },
          to_sql_mode:
            "NO_ENGINE_SUBSTITUTION,ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES",
          to_commit: "WORKING",
          to_commit_date: "2023-03-09T07:56:29.035Z",
          from_type: null,
          from_name: null,
          from_fragment: null,
          from_extra: null,
          from_sql_mode: null,
          from_commit: "HEAD",
          from_commit_date: "2023-03-09T07:56:28.841Z",
          diff_type: "added",
        },
      ],
      fields: [],
    },
    matcher: diffRowsMatcher,
  },
  {
    skip: true, // TODO: Order is not consistent
    q: "SELECT * FROM DOLT_PATCH('HEAD', 'WORKING') WHERE diff_type = 'schema'",
    res: {
      command: "SELECT",
      rowCount: 3,
      oid: null,
      rows: [
        {
          statement_order: "1",
          from_commit_hash: "",
          to_commit_hash: "WORKING",
          table_name: "test_info",
          diff_type: "schema",
          statement: "DROP TABLE `test_info`;",
        },
        // TODO: Should `CREATE SCHEMA` statement be included here?
        {
          statement_order: "2",
          from_commit_hash: "r6g8g61k89dpgb3cuks70jfnavr6b1q0",
          to_commit_hash: "WORKING",
          table_name: "testanother",
          diff_type: "schema",
          statement:
            "CREATE TABLE `testanother` (\n" +
            "  `pk` integer NOT NULL,\n" +
            "  `value` integer,\n" +
            "  PRIMARY KEY (`pk`)\n" +
            ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;",
        },
        {
          statement_order: "4", // TODO: Why is this 4?
          from_commit_hash: "",
          to_commit_hash: "WORKING",
          table_name: "dolt_schemas",
          diff_type: "schema",
          statement:
            "CREATE TABLE `dolt_schemas` (\n" + // TODO: No backticks
            "  `type` varchar(64) COLLATE utf8mb4_0900_ai_ci NOT NULL,\n" +
            "  `name` varchar(64) COLLATE utf8mb4_0900_ai_ci NOT NULL,\n" +
            "  `fragment` longtext,\n" +
            "  `extra` json,\n" +
            "  `sql_mode` varchar(256) COLLATE utf8mb4_0900_ai_ci,\n" +
            "  PRIMARY KEY (`type`,`name`)\n" +
            ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;",
        },
      ],
      fields: [],
    },
    matcher: patchRowsMatcher,
  },
  {
    q: "SELECT * FROM DOLT_PATCH('HEAD', 'WORKING', 'test_info') WHERE diff_type = 'schema'",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          statement_order: "1",
          from_commit_hash: "",
          to_commit_hash: "WORKING",
          table_name: "public.test_info",
          diff_type: "schema",
          statement: "DROP TABLE `test_info`;", // TODO: No backticks
        },
      ],
    },
    matcher: patchRowsMatcher,
  },
  {
    q: "SELECT * FROM DOLT_SCHEMA_DIFF('HEAD', 'WORKING') ORDER BY to_table_name;",
    res: {
      command: "SELECT",
      rowCount: 3,
      oid: null,
      rows: [
        {
          from_table_name: "public.test_info",
          to_table_name: "",
          from_create_statement:
            "CREATE TABLE `test_info` (\n" + // TODO: No backticks
            "  `id` integer NOT NULL,\n" +
            "  `info` varchar(255),\n" +
            "  `test_pk` integer,\n" +
            "  PRIMARY KEY (`id`),\n" +
            "  KEY `test_pk` (`test_pk`)\n" +
            "  CONSTRAINT `test_info_ibfk_1` FOREIGN KEY (`test_pk`) REFERENCES `test` (`pk`)\n" +
            ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;",
          to_create_statement: "",
        },
        {
          from_table_name: "",
          to_table_name: "anotherschema.testanother",
          from_create_statement: "",
          to_create_statement:
            "CREATE TABLE `testanother` (\n" +
            "  `pk` integer NOT NULL,\n" +
            "  `value` integer,\n" +
            "  PRIMARY KEY (`pk`)\n" +
            ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;",
        },
        {
          from_table_name: "",
          to_table_name: "public.dolt_schemas",
          from_create_statement: "",
          to_create_statement:
            "CREATE TABLE `dolt_schemas` (\n" + // TODO: No backticks
            "  `type` varchar(64) COLLATE utf8mb4_0900_ai_ci NOT NULL,\n" +
            "  `name` varchar(64) COLLATE utf8mb4_0900_ai_ci NOT NULL,\n" +
            "  `fragment` longtext,\n" +
            "  `extra` json,\n" +
            "  `sql_mode` varchar(256) COLLATE utf8mb4_0900_ai_ci,\n" +
            "  PRIMARY KEY (`type`,`name`)\n" +
            ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;",
        },
      ],
      fields: doltSchemaDiffFields,
    },
  },
  {
    q: "SELECT * FROM DOLT_SCHEMA_DIFF('HEAD', 'WORKING', 'test_info')",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          from_table_name: "public.test_info",
          to_table_name: "",
          from_create_statement:
            "CREATE TABLE `test_info` (\n" + // TODO: No backticks
            "  `id` integer NOT NULL,\n" +
            "  `info` varchar(255),\n" +
            "  `test_pk` integer,\n" +
            "  PRIMARY KEY (`id`),\n" +
            "  KEY `test_pk` (`test_pk`)\n" +
            "  CONSTRAINT `test_info_ibfk_1` FOREIGN KEY (`test_pk`) REFERENCES `test` (`pk`)\n" +
            ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;",
          to_create_statement: "",
        },
      ],
      fields: doltSchemaDiffFields,
    },
  },
  {
    q: "SELECT * FROM DOLT_SCHEMA_DIFF('HEAD', 'WORKING', 'testanother')",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          from_table_name: "",
          to_table_name: "anotherschema.testanother",
          from_create_statement: "",
          to_create_statement:
            "CREATE TABLE `testanother` (\n" +
            "  `pk` integer NOT NULL,\n" +
            "  `value` integer,\n" +
            "  PRIMARY KEY (`pk`)\n" +
            ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;",
        },
      ],
      fields: doltSchemaDiffFields,
    },
  },
  {
    q: `SELECT DOLT_COMMIT('-A', '-m', $1::text)`,
    p: ["Make some changes on branch"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_commit: "" }],
      fields: doltCommitFields,
    },
  },

  // Three dot
  {
    q: "SELECT * FROM dolt_diff_summary('main...HEAD') ORDER BY to_table_name",
    res: {
      command: "SELECT",
      rowCount: 4,
      oid: null,
      rows: [
        {
          from_table_name: "public.test_info",
          to_table_name: "",
          diff_type: "dropped",
          data_change: 1,
          schema_change: 1,
        },
        {
          from_table_name: "",
          to_table_name: "anotherschema.testanother",
          diff_type: "added",
          data_change: 1,
          schema_change: 1,
        },
        {
          from_table_name: "",
          to_table_name: "public.dolt_schemas",
          diff_type: "added",
          data_change: 1,
          schema_change: 1,
        },
        {
          from_table_name: "public.test",
          to_table_name: "public.test",
          diff_type: "modified",
          data_change: 1,
          schema_change: 0,
        },
      ],
      fields: doltDiffSummaryFields,
    },
  },
  {
    q: "SELECT * FROM dolt_diff_summary('main...HEAD', 'test')",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          from_table_name: "public.test",
          to_table_name: "public.test",
          diff_type: "modified",
          data_change: 1,
          schema_change: 0,
        },
      ],
      fields: doltDiffSummaryFields,
    },
  },
  {
    q: "SELECT * FROM dolt_diff_stat('main...HEAD') ORDER BY table_name;",
    res: {
      command: "SELECT",
      rowCount: 4,
      oid: null,
      rows: [
        {
          table_name: "anotherschema.testanother",
          rows_unmodified: "0",
          rows_added: "1",
          rows_deleted: "0",
          rows_modified: "0",
          cells_added: "2",
          cells_deleted: "0",
          cells_modified: "0",
          old_row_count: "0",
          new_row_count: "1",
          old_cell_count: "0",
          new_cell_count: "2",
        },
        {
          table_name: "public.dolt_schemas",
          rows_unmodified: "0",
          rows_added: "1",
          rows_deleted: "0",
          rows_modified: "0",
          cells_added: "5",
          cells_deleted: "0",
          cells_modified: "0",
          old_row_count: "0",
          new_row_count: "1",
          old_cell_count: "0",
          new_cell_count: "5",
        },
        {
          table_name: "public.test",
          rows_unmodified: "2",
          rows_added: "0",
          rows_deleted: "0",
          rows_modified: "1",
          cells_added: "0",
          cells_deleted: "0",
          cells_modified: "1",
          old_row_count: "3",
          new_row_count: "3",
          old_cell_count: "6",
          new_cell_count: "6",
        },
        {
          table_name: "public.test_info",
          rows_unmodified: "0",
          rows_added: "0",
          rows_deleted: "1",
          rows_modified: "0",
          cells_added: "0",
          cells_deleted: "3",
          cells_modified: "0",
          old_row_count: "1",
          new_row_count: "0",
          old_cell_count: "3",
          new_cell_count: "0",
        },
      ],
      fields: doltDiffStatFields,
    },
  },
  {
    q: "SELECT * FROM dolt_diff_stat('main...HEAD', 'test_info')",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          table_name: "public.test_info",
          rows_unmodified: "0",
          rows_added: "0",
          rows_deleted: "1",
          rows_modified: "0",
          cells_added: "0",
          cells_deleted: "3",
          cells_modified: "0",
          old_row_count: "1",
          new_row_count: "0",
          old_cell_count: "3",
          new_cell_count: "0",
        },
      ],
      fields: doltDiffStatFields,
    },
  },
  {
    skip: true, // TODO: Order not consistent
    q: "SELECT * FROM DOLT_PATCH('main...HEAD') WHERE diff_type = 'schema'",
    res: {
      command: "SELECT",
      rowCount: 4,
      oid: null,
      rows: [
        {
          statement_order: "1",
          from_commit_hash: "",
          to_commit_hash: "",
          table_name: "public.test_info",
          diff_type: "schema",
          statement: "DROP TABLE `test_info`;",
        },
        {
          statement_order: "2",
          from_commit_hash: "",
          to_commit_hash: "",
          table_name: "anotherschema.testanother",
          diff_type: "schema",
          statement:
            "CREATE TABLE `testanother` (\n" +
            "  `pk` integer NOT NULL,\n" +
            "  `value` integer,\n" +
            "  PRIMARY KEY (`pk`)\n" +
            ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;",
        },
        // TODO: `CREATE SCHEMA` here?
        {
          statement_order: "4", // TODO: why
          from_commit_hash: "",
          to_commit_hash: "",
          table_name: "public.dolt_schemas",
          diff_type: "schema",
          statement:
            "CREATE TABLE `dolt_schemas` (\n" +
            "  `type` varchar(64) COLLATE utf8mb4_0900_ai_ci NOT NULL,\n" +
            "  `name` varchar(64) COLLATE utf8mb4_0900_ai_ci NOT NULL,\n" +
            "  `fragment` longtext,\n" +
            "  `extra` json,\n" +
            "  `sql_mode` varchar(256) COLLATE utf8mb4_0900_ai_ci,\n" +
            "  PRIMARY KEY (`type`,`name`)\n" +
            ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin;",
        },
      ],
      fields: [],
    },
    matcher: patchRowsMatcher,
  },
  {
    q: "SELECT * FROM DOLT_PATCH('main...HEAD', 'test_info') WHERE diff_type = 'schema'",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          statement_order: "1",
          from_commit_hash: "",
          to_commit_hash: "",
          table_name: "public.test_info",
          diff_type: "schema",
          statement: "DROP TABLE `test_info`;",
        },
      ],
      fields: [],
    },
    matcher: patchRowsMatcher,
  },
];
