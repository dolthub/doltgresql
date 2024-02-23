# DoltgreSQL is Dolt for Postgres!

From the creators of [Dolt](https://www.doltdb.com), the world's first version controlled SQL database,
DoltgreSQL (aka Doltgres) is a Postgres-flavored version of Dolt. Doltgres offers all the Git-style log, 
diff, branch, and merge functionality of Dolt on your Postgres database schema and data. Instead of 
connecting with a MySQL client and using MySQL SQL, you connect to Doltgres with a Postgres client and 
use Postgres SQL. Doltgres is like Git and Postgres had a baby.

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

* Star this repo to tell us you are interested.
* [Try Doltgres](#getting-started)
* Create [issues](https://github.com/dolthub/doltgresql/issues) if you find bugs
* Create [issues](https://github.com/dolthub/doltgresql/issues) for missing functionality you want
* Contribute code for features you want (see the [Contribution Guide](https://github.com/dolthub/doltgresql/blob/main/CONTRIBUTING.md))

# Getting Started

1. Download the latest release of `doltgres`
   
2. Put `doltgres` on your `PATH`

3. Run `doltgres`. This will create a `doltgres` user and a `doltgres` database in `~/doltgres/databases` (add the `--data-dir` argument or change the `DOLTGRES_DATA_DIR` environment variable to use a different directory).
```bash
$ doltgres
Successfully initialized dolt data repository.
Starting server with Config HP="localhost:5432"|T="28800000"|R="false"|L="info"|S="/tmp/mysql.sock"
```

4. Make sure you have Postgres version 15 or higher installed. I used Homebrew to install Postgres on my Mac.
This requires I manually add `/opt/homebrew/opt/postgresql@15/bin` to my path. On Postgres version 14 or lower,
`\` commands (ie. `\d`, `\l`) do not yet work with Doltgres. 
```
export PATH="/opt/homebrew/opt/postgresql@15/bin:$PATH"
```

5. Open a new terminal. Connect with the following command: `psql -h localhost -U doltgres`. This will connect to the `doltgres` database with the `doltgres` user.
```bash
$ psql -h 127.0.0.1 -U doltgres
psql (15.4 (Homebrew), server 15.0)
Type "help" for help.

doltgres=>
```

6. Create a `getting_started` database. Create the `getting_started` example tables.
```sql
doltgres=> create database getting_started;
--
(0 rows)

doltgres=> \c getting_started;
psql (15.4 (Homebrew), server 15.0)
You are now connected to database "getting_started" as user "doltgres".
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
getting_started=> select * from dolt_status;
   table_name    | staged |  status   
-----------------+--------+-----------
 employees       | 0      | new table
 employees_teams | 0      | new table
 teams           | 0      | new table
(3 rows)

getting_started=> call dolt_add('teams', 'employees', 'employees_teams');
 status 
--------
      0
(1 row)
getting_started=> select * from dolt_status; 
   table_name    | staged |  status   
-----------------+--------+-----------
 employees       | 1      | new table
 employees_teams | 1      | new table
 teams           | 1      | new table
(3 rows)

getting_started=> call dolt_commit('-m', 'Created initial schema');
               hash               
----------------------------------
 peqq98e2dl5gscvfvic71e7j6ne34533
(1 row)
```

8. View the Dolt log.
```
getting_started=> select * from dolt_log;
           commit_hash            | committer |       email        |        date         |          message           
----------------------------------+-----------+--------------------+---------------------+----------------------------
 peqq98e2dl5gscvfvic71e7j6ne34533 | doltgres  | doltgres@127.0.0.1 | 2023-11-01 22:08:04 | Created initial schema
 in7bk735qa6p6rv6i3s797jjem2pg4ru | timsehn   | tim@dolthub.com    | 2023-11-01 22:04:03 | Initialize data repository
(2 rows)
```

9. Continue with [Dolt Getting Started](https://docs.dolthub.com/introduction/getting-started/database#insert-some-data) 
to test out more Doltgres versioning functionality.

# Building From Source

Please follow the [Contributor's Guide](https://github.com/dolthub/doltgresql/blob/main/CONTRIBUTING.md#getting-set-up) to learn how to build from source.

# Limitations

* No [Git-style CLI](https://docs.dolthub.com/cli-reference/cli) for version control like in [Dolt](https://github.com/dolthub/dolt), only a SQL interface.
* Can't push to DoltHub or DoltLab, only custom remotes.
* Limited support of Postgres-specific types and functions.
* No Postgres system tables.
* No authentication or users.
* Database and schema models are merged.
* Limited support for SSL connections (non-verified connections only).
* No GSSAPI support.
* No PostgreSQL functions have been implemented, therefore only MySQL functions may be used.
* No support for replication, clustering, etc.

# Performance

Dolt is [1.7X slower than MySQL](https://docs.dolthub.com/sql-reference/benchmarks/latency) as measured by
a standard suite of Sysbench tests.

We use these same Sysbench tests to benchmark DoltgreSQL and compare the results to PostgreSQL. 

Here are the benchmarks for DoltgreSQL version `0.4.0`.

<!-- START_LATENCY_RESULTS_TABLE -->
|       Read Tests        | PostgreSQL | DoltgreSQL | Multiple |
|-------------------------|------------|------------|----------|
| oltp\_point\_select     |       0.13 |       0.54 |      4.2 |
| oltp\_read\_only        |       2.35 |      12.75 |      5.4 |
| select\_random\_points  |        0.2 |       1.04 |      5.2 |
| select\_random\_ranges  |        0.4 |       1.03 |      2.6 |
| reads\_mean\_multiplier |            |            |      4.4 |

|       Write Tests        | PostgreSQL | DoltgreSQL | Multiple |
|--------------------------|------------|------------|----------|
| oltp\_insert             |       0.78 |       3.02 |      3.9 |
| oltp\_read\_write        |       3.89 |      20.37 |      5.2 |
| oltp\_update\_index      |       0.81 |       3.19 |      3.9 |
| oltp\_update\_non\_index |       0.78 |       3.13 |      4.0 |
| oltp\_write\_only        |       1.37 |       7.56 |      5.5 |
| writes\_mean\_multiplier |            |            |      4.5 |

| Overall Mean Multiple | 4.4 |
|-----------------------|-----|
<!-- END_LATENCY_RESULTS_TABLE -->
<br/>

# Correctness

Dolt is [99.99% compatible](https://docs.dolthub.com/sql-reference/benchmarks/correctness) with MySQL based on a 
standard suite of correctness tests called `sqllogictest`.

We use these same tests to measure the correctness of DoltgreSQL.

Here are DoltgreSQL's sqllogictest results for version `0.4.0`.  Tests that
did not run could not complete due to a timeout earlier in the run.

<!-- START_CORRECTNESS_RESULTS_TABLE -->
| Results |  Count  |
|---------|---------|
| not ok  |  767635 |
| ok      | 4912120 |

| Total Tests | 5679755 |
|-------------|---------|

| Correctness Percentage | 86.484716 |
|------------------------|-----------|
<!-- END_CORRECTNESS_RESULTS_TABLE -->
<br/>

# Architecture

Doltgres emulates a Postgres server, including parsing Postgres SQL into an Abstract Syntax Tree (AST). This AST is
converted to a form that can be interpreted by the Dolt engine. Doltgres uses the same SQL engine and storage format as Dolt.

[Dolt has a unique architecure](https://docs.dolthub.com/architecture/architecture) that allows for version control
features at OLTP database performance. Doltgres uses the same architecture.
