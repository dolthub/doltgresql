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

## Getting started

Add the following to your `Cargo.toml`

```toml
[dependencies]
pg_query = "6.2"
```

## Examples

### Parsing a query

```rust
use pg_query::NodeRef;

let result = pg_query::parse("SELECT * FROM contacts");
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
let result = pg_query::fingerprint("SELECT * FROM contacts.person WHERE id IN (1, 2, 3, 4);").unwrap();
assert_eq!(result.hex, "643d2a3c294ab8a7");
```

### Truncating a query

```rust
let query = "INSERT INTO \"x\" (a, b, c, d, e, f) VALUES (?)";
let result = pg_query::parse(query).unwrap();
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
