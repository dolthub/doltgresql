# DoltgreSQL is Dolt for Postgres!

From the creators of [Dolt](https://www.doltdb.com), the world's first version controlled SQL database,
DoltgreSQL (aka [Doltgres](https://www.doltgres.com)) is a Postgres-flavored version of Dolt. Doltgres offers all the Git-style log,
diff, branch, and merge functionality of Dolt on your Postgres database schema and data. Instead of
connecting with a MySQL client and using MySQL SQL, you connect to Doltgres with a Postgres client and
use Postgres SQL. Doltgres is like Git and Postgres had a baby.

# Documentation

Doltgres has a [documentation website](https://docs.doltgres.com) with more extensive documentation.

# Motivation

Dolt [was built MySQL-flavored](https://www.dolthub.com/blog/2022-03-28-have-postgres-want-dolt/#why-is-dolt-mysql-flavored-anyway).
There is no MySQL code in Dolt. In 2019, when we were conceiving of Dolt, MySQL was the most popular
SQL-flavor. Over the past 5 years, the tide has shifted more towards Postgres, especially among
young companies, Dolt's target market. Potential customers
[have been clamoring for a Postgres version of Dolt](https://github.com/dolthub/dolt/issues/4840).

Moreover, Dolt was conceived of and built as [Git for Data](https://www.dolthub.com/blog/2020-03-06-so-you-want-git-for-data/).
Dolt later became a [version controlled database](https://www.dolthub.com/blog/2021-09-17-database-version-control/).
DoltgreSQL gives us the opportunity to strip out some of the Git for Data pieces like the CLI
and build directly for the version controlled database use case. With Doltgres, you start a server,
connect a client, and do everything with SQL, a familiar experience for Postgres users.

Doltgres will diverge from Dolt over time to be a focused database version control solution.
That said, we have a five year head start with Dolt. Dolt is a production-grade version
controlled database today. [Dolt is 1.0](https://www.dolthub.com/blog/2023-05-05-dolt-1-dot-0/).
If you are ok with using a MySQL-client, we recommend using Dolt for all use cases. Doltgres
is [experimental](#limitations).

# How You Can Help

Doltgres is experimental. We need your feedback to understand how much we should invest in it.
If you are interested in using Doltgres now or in the future, please:

- Star this repo to tell us you are interested.
- [Try Doltgres](#getting-started)
- Create [issues](https://github.com/dolthub/doltgresql/issues) if you find bugs
- Create [issues](https://github.com/dolthub/doltgresql/issues) for missing functionality you want
- Contribute code for features you want (see the [Contribution Guide](https://github.com/dolthub/doltgresql/blob/main/CONTRIBUTING.md))

# Getting Started

1. Download the latest release of `doltgres`
2. Put `doltgres` on your `PATH`

3. Run `doltgres`. This will create a `postgres` user and a `postgres` database in `~/doltgres/databases` (add the `--data-dir` argument or change the `DOLTGRES_DATA_DIR` environment variable to use a different directory).

```bash
$ doltgres
Successfully initialized dolt data repository.
Starting server with Config HP="localhost:5432"|T="28800000"|R="false"|L="info"|S="/tmp/mysql.sock"
```

4. Make sure you have Postgres version 15 or higher installed. I used Homebrew to install Postgres on my Mac.
   This requires I manually add `/opt/homebrew/opt/postgresql@15/bin` to my path. On Postgres version 14 or lower,
   `\` commands (ie. `\d`, `\l`) do not yet work with Doltgres. We only need Postgres in order to use PSQL, so feel free to skip this step if you already have a Postgres client. Doltgres does not depend on any Postgres code.

```
export PATH="/opt/homebrew/opt/postgresql@15/bin:$PATH"
```

5. Open a new terminal. Connect with the following command: `psql -h localhost -U postgres`. This will connect to the `postgres` database with the `postgres` user. The default password will be `password`.

```bash
$ psql -h 127.0.0.1 -U postgres
psql (15.4 (Homebrew), server 15.0)
Type "help" for help.

postgres=>
```

6. Create a `getting_started` database. Create the `getting_started` example tables.

```sql
postgres=> create database getting_started;
--
(0 rows)

postgres=> \c getting_started;
psql (15.4 (Homebrew), server 15.0)
You are now connected to database "getting_started" as user "postgres".
getting_started=> create table employees (
    id int8,
    last_name text,
    first_name text,
    primary key(id));
--
(0 rows)

getting_started=> create table teams (
    id int8,
    team_name text,
    primary key(id));
--
(0 rows)

getting_started=> create table employees_teams(
    team_id int8,
    employee_id int8,
    primary key(team_id, employee_id),
    foreign key (team_id) references teams(id),
    foreign key (employee_id) references employees(id));
--
(0 rows)

getting_started=> \d
              List of relations
 Schema |      Name       | Type  |  Owner
--------+-----------------+-------+----------
 public | employees       | table | postgres
 public | employees_teams | table | postgres
 public | teams           | table | postgres
(3 rows)
```

7. Make a Dolt Commit.

```sql
getting_started=> select * from dolt.status;
   table_name           | staged |  status
------------------------+--------+-----------
 public.employees       | f      | new table
 public.employees_teams | f      | new table
 public.teams           | f      | new table
(3 rows)

getting_started=> select dolt_add('teams', 'employees', 'employees_teams');
 dolt_add
----------
 {0}
(1 row)
getting_started=> select * from dolt.status;
   table_name          | staged |  status
-----------------------+--------+-----------
public.employees       | t      | new table
public.employees_teams | t      | new table
public.teams           | t      | new table
(3 rows)

getting_started=> select dolt_commit('-m', 'Created initial schema');
            dolt_commit
------------------------------------
 {peqq98e2dl5gscvfvic71e7j6ne34533}
(1 row)
```

8. View the Dolt log.

```sql
getting_started=> select * from dolt.log;
           commit_hash            | committer |       email        |        date         |          message
----------------------------------+-----------+--------------------+---------------------+----------------------------
 peqq98e2dl5gscvfvic71e7j6ne34533 | postgres  | postgres@127.0.0.1 | 2023-11-01 22:08:04 | Created initial schema
 in7bk735qa6p6rv6i3s797jjem2pg4ru | timsehn   | tim@dolthub.com    | 2023-11-01 22:04:03 | Initialize data repository
(2 rows)
```

9. Continue with [Dolt Getting Started](https://docs.dolthub.com/introduction/getting-started/database#insert-some-data)
   to test out more Doltgres versioning functionality.

# Building From Source

Please follow the [Contributor's Guide](https://github.com/dolthub/doltgresql/blob/main/CONTRIBUTING.md#getting-set-up) to learn how to build from source.

# Limitations

- No [Git-style CLI](https://docs.dolthub.com/cli-reference/cli) for version control like in [Dolt](https://github.com/dolthub/dolt), only a SQL interface.
- Can't push to DoltHub or DoltLab, only custom remotes.
- Limited support of Postgres-specific types and functions.
- No Postgres system tables.
- No authentication or users.
- Database and schema models are merged.
- Limited support for SSL connections (non-verified connections only).
- No GSSAPI support.
- No PostgreSQL functions have been implemented, therefore only MySQL functions may be used.
- No support for replication, clustering, etc.

# Performance

Dolt is [1.1X slower than MySQL](https://docs.dolthub.com/sql-reference/benchmarks/latency) as measured by
a standard suite of Sysbench tests.

We use these same Sysbench tests to benchmark DoltgreSQL and compare the results to PostgreSQL.

Here are the benchmarks for DoltgreSQL version `0.12.0`.

<!-- START_LATENCY_RESULTS_TABLE -->

| Read Tests                   | PostgreSQL | DoltgreSQL | Multiple |
| ---------------------------- | ---------- | ---------- | -------- |
| covering_index_scan_postgres | 1.82       | 4.25       | 2.3      |
| groupby_scan_postgres        | 5.37       | 43.39      | 8.1      |
| index_join_postgres          | 1.96       | 10.65      | 5.4      |
| index_join_scan_postgres     | 0.74       | 9.56       | 12.9     |
| index_scan_postgres          | 18.28      | 106.75     | 5.8      |
| oltp_point_select            | 0.14       | 0.51       | 3.6      |
| oltp_read_only               | 2.52       | 12.98      | 5.2      |
| select_random_points         | 0.21       | 1.12       | 5.3      |
| select_random_ranges         | 0.41       | 1.37       | 3.3      |
| table_scan_postgres          | 18.28      | 106.75     | 5.8      |
| types_table_scan_postgres    | 44.98      | 223.34     | 5.0      |
| reads_mean_multiplier        |            |            | 5.7      |

| Write Tests                  | PostgreSQL | DoltgreSQL | Multiple |
| ---------------------------- | ---------- | ---------- | -------- |
| oltp_delete_insert_postgres  | 2.43       | 6.55       | 2.7      |
| oltp_insert                  | 0.97       | 3.25       | 3.4      |
| oltp_read_write              | 4.25       | 19.29      | 4.5      |
| oltp_update_index            | 1.03       | 3.07       | 3.0      |
| oltp_update_non_index        | 1.03       | 2.97       | 2.9      |
| oltp_write_only              | 1.64       | 6.32       | 3.9      |
| types_delete_insert_postgres | 2.03       | 6.21       | 3.1      |
| writes_mean_multiplier       |            |            | 3.4      |

| Overall Mean Multiple | 4.8 |
| --------------------- | --- |

<!-- END_LATENCY_RESULTS_TABLE -->
<br/>

# Correctness

Dolt is [100% compatible](https://docs.dolthub.com/sql-reference/benchmarks/correctness) with MySQL based on a
standard suite of correctness tests called `sqllogictest`.

We use these same tests to measure the correctness of DoltgreSQL.

Here are DoltgreSQL's sqllogictest results for version `0.12.0`. Tests that
did not run could not complete due to a timeout earlier in the run.

<!-- START_CORRECTNESS_RESULTS_TABLE -->

| Results     | Count   |
| ----------- | ------- |
| did not run | 91270   |
| not ok      | 464029  |
| ok          | 5135990 |
| timeout     | 16      |

| Total Tests | 5691305 |
| ----------- | ------- |

| Correctness Percentage | 90.242747 |
| ---------------------- | --------- |

<!-- END_CORRECTNESS_RESULTS_TABLE -->
<br/>

# Architecture

Doltgres emulates a Postgres server, including parsing Postgres SQL into an Abstract Syntax Tree (AST). This AST is
converted to a form that can be interpreted by the Dolt engine. Doltgres uses the same SQL engine and storage format as Dolt.

[Dolt has a unique architecture](https://docs.dolthub.com/architecture/architecture) that allows for version control
features at OLTP database performance. Doltgres uses the same architecture.
