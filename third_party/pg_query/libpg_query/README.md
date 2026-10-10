# libpg_query

C library for accessing the PostgreSQL parser outside of the server.

This library uses the actual PostgreSQL server source to parse SQL queries and return the internal PostgreSQL parse tree.

Note that this is mostly intended as a base library for [pg_query](https://github.com/pganalyze/pg_query) (Ruby), [pg_query.go](https://github.com/pganalyze/pg_query_go) (Go), [pg_query.rs](https://github.com/pganalyze/pg_query.rs) (Rust), [pgsql-parser](https://github.com/launchql/pgsql-parser) (Node), [psqlparse](https://github.com/alculquicondor/psqlparse) (Python) and [pglast](https://pypi.org/project/pglast/) (Python 3).

You can find further background to why a query's parse tree is useful here: https://pganalyze.com/blog/parse-postgresql-queries-in-ruby.html


## Installation

```sh
git clone -b 18-latest git://github.com/pganalyze/libpg_query
cd libpg_query
make
```

Due to compiling parts of PostgreSQL, running `make` will take a bit. Expect up to 3 minutes.

For a production build, its best to use a specific git tag (see CHANGELOG).

When compiling on Windows with Visual Studio, instead use `nmake` with the `Makefile.msvc`:

```sh
nmake /F Makefile.msvc
```


## Usage: Parsing a query

A [full example](https://github.com/pganalyze/libpg_query/blob/master/examples/simple.c) that parses a query looks like this:

```c
#include <pg_query.h>
#include <stdio.h>

int main() {
  PgQueryParseResult result;

  result = pg_query_parse("SELECT 1");

  printf("%s\n", result.parse_tree);

  pg_query_free_parse_result(result);

  return 0;
}
```

Compile it like this:

```
cc -Ilibpg_query -Llibpg_query example.c -lpg_query
```

This will output the parse tree (whitespace adjusted here for better readability):

```json
{
    "version": 180006,
    "stmts": [
        {
            "stmt": {
                "SelectStmt": {
                    "targetList": [
                        {
                            "ResTarget": {
                                "val": {
                                    "A_Const": {
                                        "ival": {
                                            "ival": 1
                                        },
                                        "location": 7
                                    }
                                },
                                "location": 7
                            }
                        }
                    ],
                    "limitOption": "LIMIT_OPTION_DEFAULT",
                    "op": "SETOP_NONE"
                }
            }
        }
    ]
}
```

## Usage: Scanning a query into its tokens using the PostgreSQL scanner/lexer

pg_query also exposes the underlying scanner of Postgres, which is also used in
the very first part in the parsing process. It can be useful on its own for e.g.
syntax highlighting, where one is mostly concerned with differentiating keywords
from identifiers and other parts of the query:

```c
#include <stdio.h>

#include <pg_query.h>

int main() {
  PgQueryScanTokensResult result;
  const char *input = "SELECT update AS left /* comment */ FROM between";

  result = pg_query_scan_tokens(input);
  if (result.error) {
    printf("error: %s at %d\n", result.error->message, result.error->cursorpos);
    pg_query_free_scan_tokens_result(result);
    return 1;
  }

  printf("  tokens: %d\n", result.n_tokens);
  for (int j = 0; j < result.n_tokens; j++) {
    PgQueryScanToken token = result.tokens[j];
    printf("  \"%.*s\" = [ %d, %d, %s, %s ]\n", token.end - token.start, &(input[token.start]), token.start, token.end,
           pg_query_token_name(token.token), pg_query_keyword_kind_name(token.keyword_kind));
  }

  pg_query_free_scan_tokens_result(result);

  return 0;
}
```

If you need the tokens as a Protobuf message instead, e.g. to decode them in another language, use `pg_query_scan`, which returns the same tokens serialized as a `ScanResult`. It can be decoded with the vendored [upb](https://github.com/protocolbuffers/protobuf/tree/main/upb) runtime, which is included in `libpg_query.a` (the upb headers must then be on the include path as well, `-Ilibpg_query/vendor/upb`).

This will output the following:

```
  tokens: 7
  "SELECT" = [ 0, 6, SELECT, RESERVED_KEYWORD ]
  "update" = [ 7, 13, UPDATE, UNRESERVED_KEYWORD ]
  "AS" = [ 14, 16, AS, RESERVED_KEYWORD ]
  "left" = [ 17, 21, LEFT, TYPE_FUNC_NAME_KEYWORD ]
  "/* comment */" = [ 22, 35, C_COMMENT, NO_KEYWORD ]
  "FROM" = [ 36, 40, FROM, RESERVED_KEYWORD ]
  "between" = [ 41, 48, BETWEEN, COL_NAME_KEYWORD ]
```

Where the each element in the token list has the following fields:

1. Start location in the source string
2. End location in the source string
3. Token value - `PgQueryToken` enum (e.g. `PG_QUERY_TOKEN_SELECT`), see `pg_query_scan_tokens.h`, or the Token type in `protobuf/pg_query.proto` for the Protobuf output
4. Keyword type - `PgQueryKeywordKind` enum, see `pg_query_scan_tokens.h`, or the KeywordKind type in `protobuf/pg_query.proto` for the Protobuf output, possible values:
  `PG_QUERY_NO_KEYWORD`: Not a keyword
  `PG_QUERY_UNRESERVED_KEYWORD`: Unreserved keyword (available for use as any kind of unescaped name)
  `PG_QUERY_COL_NAME_KEYWORD`: Unreserved keyword (can be unescaped column/table/etc names, cannot be unescaped function or type name)
  `PG_QUERY_TYPE_FUNC_NAME_KEYWORD`: Reserved keyword (can be unescaped function or type name, cannot be unescaped column/table/etc names)
  `PG_QUERY_RESERVED_KEYWORD`: Reserved keyword (cannot be unescaped column/table/variable/type/function names)

The token values are the token numbers of the Postgres grammar, which change between Postgres versions, so compare them by name rather than by value.

Note that whitespace does not show as tokens.

## Usage: Fingerprinting a query

Fingerprinting allows you to identify similar queries that are different only because
of the specific object that is being queried for (i.e. different object ids in the WHERE clause),
or because of formatting.

Example:

```c
#include <pg_query.h>
#include <stdio.h>

int main() {
  PgQueryFingerprintResult result;

  result = pg_query_fingerprint("SELECT 1");

  printf("%s\n", result.fingerprint_str);

  pg_query_free_fingerprint_result(result);

  return 0;
}
```

This will output:

```
50fde20626009aba
```

In general, fingerprinting intends to be a superset of the Postgres `queryid` mechanism. That means that for a given Postgres `queryid`, there should only be one valid fingerprint - even when varying the inputs that produce that same `queryid`. A simple example of this would be that the Postgres `queryid` ignores constant values (e.g. `... WHERE id = 123` is the same queryid as `... WHERE id = 456`), and thus the fingerprint also behaves that way.

One special case are the changes done in Postgres 18, which significantly changed how schema names and table aliases in `FROM` clauses are handled for `queryid` values. On Postgres 17 and older, the schema was significant, and the alias was always ignored. On Postgres 18, the schema is always ignored, and the table alias (if present) is used instead of the table name.

To support matching behaviour, you can optionally pass fingerprint options. By default fingerprinting behaves like Postgres 18, but you can pass `PG_QUERY_FINGERPRINT_RANGEVAR_PG17_COMPAT` to instead have it behave like Postgres 17 does for `queryid`:

```c
#include <pg_query.h>
#include <stdio.h>

int main() {
  const char *queries[] = {
    "SELECT * FROM public.users u",
    "SELECT * FROM myschema.users u",
  };
  PgQueryFingerprintResult result;

  for (int i = 0; i < 2; i++) {
    result = pg_query_fingerprint_opts(queries[i], PG_QUERY_PARSE_DEFAULT, PG_QUERY_FINGERPRINT_DEFAULT);
    printf("Default:     %s\n", result.fingerprint_str);
    pg_query_free_fingerprint_result(result);

    result = pg_query_fingerprint_opts(queries[i], PG_QUERY_PARSE_DEFAULT, PG_QUERY_FINGERPRINT_RANGEVAR_PG17_COMPAT);
    printf("PG17 compat: %s\n", result.fingerprint_str);
    pg_query_free_fingerprint_result(result);
  }

  return 0;
}
```

This will output:

```
Default:     6640d8be64880eed
PG17 compat: d3198777453aef12
Default:     6640d8be64880eed
PG17 compat: 91e880cf13f4f219
```

With the default options both queries get the same fingerprint, because the schema is ignored and the alias `u` is used. With `PG_QUERY_FINGERPRINT_RANGEVAR_PG17_COMPAT`, the schema-qualified table name counts, so the two queries get different fingerprints.

The options also allow controlling whether relation names that have subsequent digits (often used for daily partitions) should be ignored for fingerprinting, which is the case by default.

See https://github.com/pganalyze/libpg_query/wiki/Fingerprinting for the full fingerprinting rules.

## Usage: Parsing a PL/pgSQL function

A [full example](https://github.com/pganalyze/libpg_query/blob/master/examples/simple_plpgsql.c) that parses a [PL/pgSQL](https://www.postgresql.org/docs/current/static/plpgsql.html) method looks like this:

```c
#include <pg_query.h>
#include <stdio.h>
#include <stdlib.h>

int main() {
  PgQueryPlpgsqlParseResult result;

  result = pg_query_parse_plpgsql(" \
  CREATE OR REPLACE FUNCTION cs_fmt_browser_version(v_name varchar, \
                                                  v_version varchar) \
RETURNS varchar AS $$ \
BEGIN \
    IF v_version IS NULL THEN \
        RETURN v_name; \
    END IF; \
    RETURN v_name || '/' || v_version; \
END; \
$$ LANGUAGE plpgsql;");

  if (result.error) {
    printf("error: %s at %d\n", result.error->message, result.error->cursorpos);
  } else {
    printf("%s\n", result.plpgsql_funcs);
  }

  pg_query_free_plpgsql_parse_result(result);

  return 0;
}
```

This will output (formatted for clarity):

```json
[
    {
        "PLpgSQL_function": {
            "datums": [
                {
                    "PLpgSQL_var": {
                        "refname": "v_name",
                        "datatype": {
                            "PLpgSQL_type": {
                                "typname": "varchar"
                            }
                        }
                    }
                },
                {
                    "PLpgSQL_var": {
                        "refname": "v_version",
                        "datatype": {
                            "PLpgSQL_type": {
                                "typname": "varchar"
                            }
                        }
                    }
                },
                {
                    "PLpgSQL_var": {
                        "refname": "found",
                        "datatype": {
                            "PLpgSQL_type": {
                                "typname": "bool"
                            }
                        }
                    }
                }
            ],
            "action": {
                "PLpgSQL_stmt_block": {
                    "lineno": 1,
                    "body": [
                        {
                            "PLpgSQL_stmt_if": {
                                "lineno": 1,
                                "cond": {
                                    "PLpgSQL_expr": {
                                        "query": "v_version IS NULL",
                                        "parseMode": 2
                                    }
                                },
                                "then_body": [
                                    {
                                        "PLpgSQL_stmt_return": {
                                            "lineno": 1
                                        }
                                    }
                                ]
                            }
                        },
                        {
                            "PLpgSQL_stmt_return": {
                                "lineno": 1,
                                "expr": {
                                    "PLpgSQL_expr": {
                                        "query": "v_name || '/' || v_version",
                                        "parseMode": 2
                                    }
                                }
                            }
                        }
                    ]
                }
            }
        }
    }
]
```

## Versions

For stability, it is recommended you use individual tagged git versions, see CHANGELOG.

Each major version is maintained in a dedicated git branch. Only the latest Postgres stable release receives active updates, with exceptions being made for the prior release branch for critical fixes.

| PostgreSQL Major Version | Branch     | Status              |
|--------------------------|------------|---------------------|
| 18                       | 18-latest  | Active development  |
| 17                       | 17-latest  | Critical fixes only |
| 16                       | 16-latest  | No longer supported |
| 15                       | 15-latest  | No longer supported |
| 14                       | 14-latest  | No longer supported |
| 13                       | 13-latest  | No longer supported |
| 12                       | (n/a)      | Not supported       |
| 11                       | (n/a)      | Not supported       |
| 10                       | 10-latest  | No longer supported |
| 9.6                      | (n/a)      | Not supported       |
| 9.5                      | 9.5-latest | No longer supported |
| 9.4                      | 9.4-latest | No longer supported |

## Updating the vendored upb (Protobuf) runtime

Protobuf serialization uses [upb](https://github.com/protocolbuffers/protobuf/tree/main/upb), vendored in `vendor/upb`. The generated code in `protobuf/pg_query.upb*.{c,h}` is tied to the exact upb version it was generated with, so both are updated together:

```sh
brew upgrade protobuf       # protoc + protoc-gen-upb must match the target release
make -C vendor/upb update TAG=v$(protoc --version | awk '{print $2}')
make clean && make && make test
```

See [vendor/upb/README](vendor/upb/README) for details, including the local patches that are re-applied on each update.

## Resources

pg_query wrappers in other languages:

* Ruby: [pg_query](https://github.com/pganalyze/pg_query)
* Go: [pg_query_go](https://github.com/pganalyze/pg_query_go)
* Javascript (Node): [pgsql-parser](https://github.com/launchql/pgsql-parser)
* Javascript (Browser): [pg-query-emscripten](https://github.com/pganalyze/pg-query-emscripten)
* Python: [psqlparse](https://github.com/alculquicondor/psqlparse), [pglast](https://github.com/lelit/pglast),
[psqlparse2](https://github.com/dani-maarouf/psqlparse2)
* OCaml: [pg_query-ocaml](https://github.com/roddyyaga/pg_query-ocaml)
* Rust: [pg_query.rs](https://github.com/pganalyze/pg_query.rs)

Products, tools and libraries built on pg_query:

* [pganalyze](https://pganalyze.com/)
* [hsql](https://github.com/JackDanger/hsql)
* [sqlint](https://github.com/purcell/sqlint)
* [pghero](https://github.com/ankane/pghero)
* [dexter](https://github.com/ankane/dexter)
* [pgscope](https://github.com/gjalves/pgscope)
* [pg_materialize](https://github.com/aanari/pg-materialize)
* [DuckDB](https://github.com/cwida/duckdb) ([details](https://github.com/cwida/duckdb/tree/master/third_party/libpg_query))
* [pgspot](https://github.com/timescale/pgspot)
* [Postgres Language Server](https://github.com/supabase/postgres_lsp)
* and more

Please feel free to [open a PR](https://github.com/pganalyze/libpg_query/pull/new/master) to add yours! :)

## License

PostgreSQL server source code, used under the [PostgreSQL license](https://www.postgresql.org/about/licence/).<br>
Portions Copyright (c) 1996-2026, The PostgreSQL Global Development Group<br>
Portions Copyright (c) 1994, The Regents of the University of California

upb Protobuf runtime (`vendor/upb`), used under the [3-clause BSD license](vendor/upb/LICENSE).<br>
Copyright 2008 Google Inc.<br>
Includes utf8_range (`vendor/upb/third_party/utf8_range`), used under the [MIT license](vendor/upb/third_party/utf8_range/LICENSE).<br>
Copyright (c) 2019 Yibo Cai<br>
Copyright 2022 Google LLC

xxHash (`vendor/xxhash`), used under the [2-clause BSD license](vendor/xxhash/xxhash.h).<br>
Copyright (C) 2012-2020 Yann Collet

All other parts are licensed under the 3-clause BSD license, see LICENSE file for details.<br>
Copyright (c) 2015, Lukas Fittl <lukas@fittl.com>
Copyright (c) 2016-2026, Duboce Labs, Inc. (pganalyze) <team@pganalyze.com>
