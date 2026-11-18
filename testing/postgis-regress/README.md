# PostGIS regression tests

This command runs the core regression tests of PostGIS 3.6.4 against a running server and reports how many pass. It
reproduces what PostGIS's own `regress/run_test.pl` does for each test:

- Each test runs in a database of its own, in which `CREATE EXTENSION postgis` has been run.
- The test's SQL is run through `psql -tXAq -v VERBOSITY=terse` with the variables that `run_test.pl` sets.
- Blank lines are dropped and the same substitutions are made to the output. The result is compared to the test's
  `_expected` file.

The tests are chosen the way `regress/core/tests.mk.in` chooses them for a PostGIS built against GEOS 3.11.1 with JSON,
SP-GiST, and protobuf support, which are the versions and features that the emulated extension follows. The slow tests
that PostGIS only runs on request are left out.

PostGIS, and its tests, are licensed under the GPL, so they are not checked in. The first run downloads the PostGIS
release, verifies its checksum, and extracts its `regress` directory into the user's cache directory.

## Running

`psql` must be installed. Start a Doltgres server, then run:

```sh
PGPASSWORD=password go run ./testing/postgis-regress -port 5432 -out /tmp/postgis-regress
```

Each test's output, and its diff from the expected output when it fails, are written to the `-out` directory, along
with `summary.txt` and `results.json`. Besides whether each test passes, the summary counts how many of a test's expected
lines its output reproduces in order, which measures progress on tests that do not pass yet. `-run` selects tests with
a regular expression.

## Checking the harness against PostGIS

Every test passes against PostgreSQL 18 with PostGIS 3.6.4, which confirms that the harness runs the tests as PostGIS
does. Tests that include PostGIS's installed `legacy.sql` need `-scriptdir`, and PostgreSQL's databases should be created
as `run_test.pl` creates them:

```sh
PGPASSWORD=postgres go run ./testing/postgis-regress -port 5432 \
    -scriptdir /usr/share/postgresql/18/contrib/postgis-3.6 \
    -create-db-options "ENCODING 'UTF8' TEMPLATE template0 LC_COLLATE 'C' LC_CTYPE 'C'"
```
