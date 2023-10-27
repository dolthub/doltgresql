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
connect a client, and do everything with SQL, a familiar experience for Postgres users. Expect 
Doltgres to diverge from Dolt over time and become the preferred database version control solution 
for most customers.

That said, we have a five year head start with core Dolt. 
[Dolt is 1.0](https://www.dolthub.com/blog/2023-05-05-dolt-1-dot-0/). If you are ok with using 
a MySQL-client, we recommend using Dolt for all use cases. Doltgres is more experimental at this point.

# Getting Started



# Limitations

* No Git-style CLI for version control, only a SQL interface.
* Can't push to DoltHub or DoltLab. 
* Limited support of Postgres types.
* No information schema support
* No users and grants
