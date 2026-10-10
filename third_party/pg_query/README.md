pg_query.rs &emsp; [![Build Status]][actions] [![Latest Version]][crates.io] [![Docs Badge]][docs]
===========

[Build Status]: https://img.shields.io/endpoint.svg?url=https%3A%2F%2Factions-badge.atrox.dev%2Fpganalyze%2Fpg_query.rs%2Fbadge%3Fref%3Dmain&style=flat&label=build&logo=none
[actions]: https://actions-badge.atrox.dev/pganalyze/pg_query.rs/goto?ref=main
[Latest Version]: https://img.shields.io/crates/v/pg_query.svg
[crates.io]: https://crates.io/crates/pg_query
[Docs Badge]: https://docs.rs/pg_query/badge.svg
[docs]: https://docs.rs/pg_query

This Rust library uses the actual PostgreSQL server source to parse SQL queries and return the internal PostgreSQL parse tree.

It also allows you to normalize queries (replacing constant values with $1, etc.) and parse these normalized queries into a parse tree again.

When you build this library, it builds parts of the PostgreSQL server source (see [libpg_query](https://github.com/pganalyze/libpg_query)), and then statically links it into this library.

You can find further examples and a longer rationale for the original Ruby implementation [here](https://pganalyze.com/blog/parse-postgresql-queries-in-ruby.html). The Rust version tries to have a very similar API.

## Examples

### Parsing a query

```rust
use pg_query::NodeRef;

let result = pg_query::parse("SELECT * FROM contacts", 0);
assert!(result.is_ok());
let result = result.unwrap();
assert_eq!(result.tables(), vec!["contacts"]);
assert!(matches!(result.protobuf.nodes()[0].0, NodeRef::SelectStmt(_)));
```

### Normalizing a query

```rust
let result = pg_query::normalize("SELECT 1 FROM x WHERE y = (SELECT 123 FROM a WHERE z = 'bla')").unwrap();
assert_eq!(result, "SELECT $1 FROM x WHERE y = (SELECT $2 FROM a WHERE z = $3)");
```

### Fingerprinting a query

```rust
let result = pg_query::fingerprint("SELECT * FROM contacts.person WHERE id IN (1, 2, 3, 4);", 0, 0).unwrap();
assert_eq!(result.hex, "5735f5c64dd9f68e");
```

Fingerprinting allows you to identify similar queries that are different only
because of the specific object being queried (e.g. different object ids in the
`WHERE` clause) or because of formatting. It is intended to be a superset of
the Postgres `queryid` mechanism: constant values are ignored, so
`... WHERE id = 123` and `... WHERE id = 456` get the same fingerprint.

Postgres 18 significantly changed how schema names and table aliases in `FROM`
clauses are handled for `queryid`, and libpg_query follows that by default:

- **Postgres 17 and earlier**: the schema name was significant, and table
  aliases were always ignored.
- **Postgres 18 (default)**: schema names are ignored, and a table alias (if
  present) is used instead of the relation name.

```rust
// Postgres 18 behaviour (default): schema is ignored and the alias replaces the relation name
let a = pg_query::fingerprint("SELECT * FROM public.users u", 0, 0).unwrap();
let b = pg_query::fingerprint("SELECT * FROM myschema.users u", 0, 0).unwrap();
assert_eq!(a.hex, b.hex);

// Postgres 17 compatible fingerprints take the schema into account
let a = pg_query::fingerprint("SELECT * FROM public.users u", 0, pg_query::FingerprintOptions::RANGEVAR_PG17_COMPAT).unwrap();
let b = pg_query::fingerprint("SELECT * FROM myschema.users u", 0, pg_query::FingerprintOptions::RANGEVAR_PG17_COMPAT).unwrap();
assert_ne!(a.hex, b.hex);
```

Available fingerprint options (combinable, passed as the third `fingerprint`
argument to `pg_query::fingerprint`):

| `FingerprintOptions`      | effect                                                                  |
| ------------------------- | ----------------------------------------------------------------------- |
| `DEFAULT` (0)             | Postgres 18 query ID behaviour                                          |
| `RANGEVAR_IGNORE_ALIASES` | relation names are always fingerprinted; aliases are ignored           |
| `RANGEVAR_INCLUDE_SCHEMA` | schema names are also fingerprinted in SELECT/DML statements           |
| `RANGEVAR_PG17_COMPAT`    | combination of the two above, matching Postgres 17 (and libpg_query 17) fingerprints |
| `FULL_RELNAME`          | fingerprint the full relation name instead of ignoring 2+ consecutive digits (which by default groups date/number-suffixed partitions together) |

See https://github.com/pganalyze/libpg_query/wiki/Fingerprinting for the
full fingerprinting rules.

### Scanning a query's tokens without protobuf

```rust
let tokens = pg_query::scan_tokens("SELECT 1").unwrap();
assert_eq!(tokens.len(), 2);
assert_eq!(tokens[0].start, 0);
assert_eq!(tokens[0].end, 6);
```

### Detecting utility statements

```rust
let result = pg_query::is_utility_stmt("SELECT 1; SHOW fsync;").unwrap();
assert_eq!(result, vec![false, true]);
```

### Deparsing a query

```rust
use pg_query::DeparseOptions;

let parsed = pg_query::parse("SELECT 1 -- comment\n", 0).unwrap();
let output = pg_query::deparse(&parsed.protobuf, DeparseOptions::default()).unwrap();
assert_eq!(output, "SELECT 1");

// With comments extracted and re-inserted, pretty-printed:
let comments = pg_query::deparse_comments_for_query("SELECT 1 -- comment\n").unwrap();
let output = pg_query::deparse(
    &parsed.protobuf,
    DeparseOptions {
        comments,
        pretty_print: true,
        ..Default::default()
    },
)
.unwrap();
assert!(output.contains("-- comment"));
```

### Truncating a query

```rust
let query = "INSERT INTO \"x\" (a, b, c, d, e, f) VALUES ($1)";
let result = pg_query::parse(query, 0).unwrap();
assert_eq!(result.truncate(32).unwrap(), "INSERT INTO x (...) VALUES (...)");
```

## Credits

Thanks to [Paul Mason](https://github.com/paupino) for his work on [pg_parse](https://github.com/paupino/pg_parse) that this crate is based on.

After version 0.6.0, Paul donated the pg_query crate to the pganalyze team. pg_parse is a lighter alternative that focuses on query parsing, while pg_query aims for feature parity with the Ruby gem.

## License

PostgreSQL server source code, used under the [PostgreSQL license](https://www.postgresql.org/about/licence/).<br>
Portions Copyright (c) 1996-2023, The PostgreSQL Global Development Group<br>
Portions Copyright (c) 1994, The Regents of the University of California

All other parts are licensed under the MIT license, see LICENSE file for details.<br>
Copyright (c) 2021 Paul Mason <paul@form1.co.nz>
Copyright (c) 2021-2023, Duboce Labs, Inc. (pganalyze) <team@pganalyze.com>
