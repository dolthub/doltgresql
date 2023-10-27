# DoltgreSQL is Dolt for Postgres!

From the creators of [Dolt](https://www.doltdb.com), the world's first version controlled SQL database,
DoltgreSQL (aka Doltgres) is a Postgres-flavored version of Dolt. Doltgres offers all the Git-style log, 
diff, branch, and merge functionality of Dolt on your Postgres database schema and data. But instead of 
connecting with a MySQL client and using MySQL SQL, you connect with a Postgres client and use Postgres 
SQL. Doltgres is like Git and Postgres had a baby.

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

* Star this repo to tell us you are interested.
* Try Doltgres
* Create issues if you find bugs
* Create issues for missing functionality you want
* Contribute Code for features you want

# Getting Started

1. Download the latest release
2. Put the binary on your PATH
3. Navigate to a directory you want your database data stored
4. Type `doltgresql`
5. Connect `psql`
6. Create database
7. Create tables
8. Make a Dolt Commit
9. View the log
10. Continue with [Dolt Getting Started](https://docs.dolthub.com/introduction/getting-started/database#insert-some-data)

# Configuration

Is here any? Daylon should fill this out.

# Limitations

* No Git-style CLI for version control, only a SQL interface.
* Can't push to DoltHub or DoltLab, only custom remotes
* Limited support of Postgres-specific types.
* No information schema support
* No users and grants

# Architecture

A translation at the AST layer. Then, same Dolt engine.

# Performance

Can we run a sysbench test?
