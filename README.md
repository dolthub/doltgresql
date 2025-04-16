# Doltgres is Dolt for Postgres!



From the creators of [Dolt](https://www.doltdb.com), the world's first version controlled SQL
database, comes [Doltgres](https://www.doltgres.com), the Postgres-flavored version of Dolt. It's a
SQL database that you can branch and merge, fork and clone, push and pull just like a Git
repository. Connect to your Doltgres server just like any Postgres database to read or modify schema
and data. Version control functionality is exposed in SQL via system tables, functions, and
procedures.

Git versions file, Doltgres versions tables. It's like Git and Postgres had a baby.

# Doltgres is Beta

[Doltgres is now Beta quality](https://dolthub.com/blog/2025-04-16-doltgres-goes-beta/), which means
it's ready for your production use case. There will be bugs and missing features, but [we can fix
most of them in 24 hours](https://www.dolthub.com/blog/2024-05-15-24-hour-bug-fixes/) if you [file
an issue](https://github.com/dolthub/doltgresql/issues).

The wait is over! Now is the time to [try out Doltgres](#getting-started) and let us know what you
think. Import your existing Postgres database into Doltgres with `pg_dump` and `psql`, and let us
know if anything doesn't work.

If you're excited about this project, you can also help speed it along in a few other ways:

- Star this repo
- Create [issues](https://github.com/dolthub/doltgresql/issues) if you find bugs
- Create [issues](https://github.com/dolthub/doltgresql/issues) for missing functionality you want
- Contribute code for features you want (see the [Contribution
  Guide](https://github.com/dolthub/doltgresql/blob/main/CONTRIBUTING.md))
- Tell your friends and colleagues

# Full Documentation

Doltgres has a [documentation website](https://docs.doltgres.com) with extensive documentation.

# Getting Started

1. Download the latest release of `doltgres`
2. Put `doltgres` on your `PATH`
3. Run `doltgres`. This will create a `postgres` user and a `postgres` database in
   `~/doltgres/databases`. The default password will be `password`, just like in Postgres. You can
   use a `config.yaml` file or set the `DOLTGRES_DATA_DIR` environment variable to use a different
   directory for your databases.

```bash
$ doltgres
INFO[0000] Server ready. Accepting connections.
```

4. Install Postgres to get the `psql` tool. I used Homebrew to install Postgres on my Mac.  This
   requires I manually add `/opt/homebrew/opt/postgresql@15/bin` to my path. We only need Postgres
   in order to use `psql`, so feel free to skip this step if you already have `psql`, or if you
   have another Postgres client you use instead.

```
export PATH="/opt/homebrew/opt/postgresql@15/bin:$PATH"
```

5. Open a new terminal. Connect with the following command: `PGPASSWORD=password psql -h localhost
   -U postgres`. This will connect to the `postgres` database with the `postgres` user.

```bash
$ PGPASSWORD=password psql -h localhost
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

# Limitations and differences from Dolt

- No [Git-style CLI](https://docs.dolthub.com/cli-reference/cli) for version control like in
  [Dolt](https://github.com/dolthub/dolt), only a SQL interface.
- Can't push to DoltHub or DoltLab, only custom remotes (such as on the file system or to S3).
- Backup and replication are a work in progress.
- No GSSAPI support.
- No extension support yet.
- Some Postgres syntax, types, functions, and features are not yet implemented. If you encounter a
  missing feature you need for your application, please [file an issue to let us
  know](https://github.com/dolthub/doltgresql/issues).

# Performance

Dolt is [1.1X slower than MySQL](https://docs.dolthub.com/sql-reference/benchmarks/latency) as
measured by a standard suite of Sysbench tests.

We use these same Sysbench tests to benchmark DoltgreSQL and compare the results to PostgreSQL.

Here are the benchmarks for DoltgreSQL version `0.50.0`. All figures are median latency in
milliseconds.

<!-- START_LATENCY_RESULTS_TABLE -->

| Read Tests                   | Postgres | Doltgres | Multiple |
| ---                          | ---      | ---      | ---      |
| covering_index_scan_postgres | 1.89     | 5.28     | 2.8      |
| groupby_scan_postgres        | 5.28     | 46.63    | 8.8      |
| index_join_postgres          | 1.96     | 10.09    | 5.1      |
| index_join_scan_postgres     | 0.67     | 8.9      | 13.3     |
| index_scan_postgres          | 17.95    | 130.13   | 7.2      |
| oltp_point_select            | 0.14     | 0.52     | 3.7      |
| oltp_read_only               | 2.48     | 12.75    | 5.1      |
| select_random_points         | 0.21     | 1.12     | 5.3      |
| select_random_ranges         | 0.41     | 1.39     | 3.4      |
| table_scan_postgres          | 17.95    | 132.49   | 7.4      |
| types_table_scan_postgres    | 43.39    | 292.6    | 6.7      |
| reads_mean_multiplier        |          |          | 6.3      |


| Write Tests                  | Postgres | Doltgres | Multiple |
|------------------------------|----------|----------|----------|
| oltp_delete_insert_postgres  | 2.22     | 6.79     | 3.1      |
| oltp_insert                  | 1.1      | 3.68     | 3.3      |
| oltp_read_write              | 4.25     | 20.37    | 4.8      |
| oltp_update_index            | 1.12     | 3.55     | 3.2      |
| oltp_update_non_index        | 1.12     | 3.43     | 3.1      |
| oltp_write_only              | 1.73     | 7.43     | 4.3      |
| types_delete_insert_postgres | 2.3      | 7.04     | 3.1      |
| write_mean_multiplier        |          |          | 3.6      |

| Overall Mean Multiple | 5.2 |
| --------------------- | --- |

<!-- END_LATENCY_RESULTS_TABLE -->
<br/>

# Correctness

Dolt is [100% compatible](https://docs.dolthub.com/sql-reference/benchmarks/correctness) with MySQL
based on a standard suite of correctness tests called `sqllogictest`.

We use these same tests to measure the correctness of DoltgreSQL.

Here are DoltgreSQL's sqllogictest results for version `0.50.0`. Tests that did not run could not
complete due to a timeout earlier in the run.

<!-- START_CORRECTNESS_RESULTS_TABLE -->

| Results     | Count   |
| --          | --      |
| did not run | 91270   |
| not ok      | 411415  |
| ok          | 5188604 |
| timeout     | 16      |
| Total Tests | 5691305 |

| Correctness Percentage | 91.16721 |
| --                     | --       |

<!-- END_CORRECTNESS_RESULTS_TABLE -->
<br/>

# Architecture

Doltgres emulates a Postgres server, including parsing Postgres SQL into an Abstract Syntax Tree (AST). This AST is
converted to a form that can be interpreted by the Dolt engine. Doltgres uses the same SQL engine and storage format as Dolt.

[Dolt has a unique architecture](https://docs.dolthub.com/architecture/architecture) that allows for version control
features at OLTP database performance. Doltgres uses the same architecture.
