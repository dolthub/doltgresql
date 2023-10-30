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
* Contribute Code for features you want (see [Building From Source](#building-from-source))

Contribution Guide coming soon.

# Getting Started

1. Download the latest release
2. Put the binary on your PATH
3. Navigate to a directory you want your database data stored. The directory name will be the name of your database.
4. Type `doltgresql`
5. Open a new terminal. Connect with the following command: `psql -h localhost -d [DB name]`
6. Create tables. Use text types to show it's Postgres.
7. Make a Dolt Commit
8. View the log
9. Continue with [Dolt Getting Started](https://docs.dolthub.com/introduction/getting-started/database#insert-some-data)

# Building From Source

Due to the rapid pace of development at this early stage, building from source will guarantee that you're always working
with the latest improvement and features.

1. Clone the repository to your local drive
2. Run `./postgres/parser/build.sh` to generate the parser
3. Run `go build .` in the root directory

# Limitations

* No Git-style CLI for version control, only a SQL interface.
* Can't push to DoltHub or DoltLab, only custom remotes.
* Limited support of Postgres-specific types and functions.
* No Postgres system tables.
* No authentication or users.
* Database and schema models are merged.
* Limited support for SSL connections (non-verified connections only).
* No GSSAPI support.

# Performance

Dolt is [1.7X to 1.8X slower than MySQL](https://docs.dolthub.com/sql-reference/benchmarks/latency) as measured by 
a standard suite of Sysnbench tests. 

Similar tests for Doltgres vs Postgres coming soon. 

# Correctness

Dolt is [99.99% compatible](https://docs.dolthub.com/sql-reference/benchmarks/correctness) with MySQL based on a 
standard suite of correctness tests called `sqllogictest`.

A similar comparison for Doltgres coming soon.

# Architecture

Doltgres emulates a Postgres server, including parsing Postgres SQL into an Abstract Syntax Tree (AST). This AST is
converted to a form that can be interpreted by the Dolt engine. Doltgres uses the same SQL engine and storage format as Dolt.

[Dolt has a unique architecure](https://docs.dolthub.com/architecture/architecture) that allows for version control
features at OLTP database performance. Doltgres uses the same architecture.
