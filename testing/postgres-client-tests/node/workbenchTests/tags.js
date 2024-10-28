import { doltTagFields, doltTagsFields } from "../fields.js";
import { tagsMatcher } from "./matchers.js";

export const tagsTests = [
  {
    q: "SELECT * FROM dolt.tags ORDER BY date DESC",
    res: {
      command: "SELECT",
      rowCount: 0,
      oid: null,
      rows: [],
      fields: doltTagsFields,
    },
  },
  {
    q: `SELECT DOLT_TAG($1::text, $2::text);`,
    p: ["mytag", "main"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_tag: ["0"] }],
      fields: doltTagFields,
    },
  },
  {
    q: "SELECT * FROM dolt.tags ORDER BY date DESC",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          tag_name: "mytag",
          message: "",
          tagger: "Dolt System Account",
          email: "doltuser@dolthub.com",
          tag_hash: "",
          date: "",
        },
      ],
      fields: doltTagsFields,
    },
    matcher: tagsMatcher,
  },
  {
    q: `SELECT DOLT_TAG('-m', $1::text, $2::text, $3::text)`,
    p: ["latest release", "mytagnew", "main"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_tag: ["0"] }],
      fields: doltTagFields,
    },
  },
  {
    q: "SELECT * FROM dolt.tags ORDER BY date DESC",
    res: {
      command: "SELECT",
      rowCount: 2,
      oid: null,
      rows: [
        {
          tag_name: "mytagnew",
          message: "latest release",
          tagger: "Dolt System Account",
          email: "doltuser@dolthub.com",
          tag_hash: "",
          date: "",
        },
        {
          tag_name: "mytag",
          message: "",
          tagger: "Dolt System Account",
          email: "doltuser@dolthub.com",
          tag_hash: "",
          date: "",
        },
      ],
      fields: doltTagsFields,
    },
    matcher: tagsMatcher,
  },
  {
    q: `SELECT DOLT_TAG('-d', $1::text)`,
    p: ["mytagnew"],
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [{ dolt_tag: ["0"] }],
      fields: doltTagFields,
    },
  },
  {
    q: "SELECT * FROM dolt.tags ORDER BY date DESC",
    res: {
      command: "SELECT",
      rowCount: 1,
      oid: null,
      rows: [
        {
          tag_name: "mytag",
          message: "",
          tagger: "Dolt System Account",
          email: "doltuser@dolthub.com",
          tag_hash: "",
          date: "",
        },
      ],
      fields: doltTagsFields,
    },
    matcher: tagsMatcher,
  },
  {
    q: `SELECT DOLT_COMMIT('-A', '-m', $1::text)`,
    p: ["Add a tag"],
    expectedErr: "nothing to commit",
  },
];
