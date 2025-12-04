# Doltgres is Dolt for Postgres!

From the creators of [Dolt](https://www.doltdb.com), the world's first version controlled SQL
database, comes [Doltgres](https://www.doltgres.com), the Postgres-flavored version of Dolt. It's a
SQL database that you can branch and merge, fork and clone, push and pull just like a Git
repository. Connect to your Doltgres server just like any Postgres database to read or modify schema
and data. Version control functionality is exposed in SQL via system tables, functions, and
procedures.

Git versions file, Doltgres versions tables. It's like Git and Postgres had a baby.

[Join us on Discord](https://discord.com/invite/RFwfYpu) to say hi and
ask questions, or [check out our roadmap](https://docs.dolthub.com/other/roadmap)
to see what we're building next.

## What's it for?

Lots of things! Doltgres is a generally useful tool with countless applications. But if you want
some ideas, [here's how people are using it so
far](https://www.dolthub.com/blog/2022-07-11-dolt-case-studies/).

# How to use this image

This image is for the Doltgres server, and is similar to the Postgres Docker image. Running this
image without any arguments is equivalent to running the `doltgres` command inside a Docker
container.

To see all supported options for `doltgres`, you can run the image with `--help` flag.

```shell
$ docker run dolthub/doltgresql:latest --help
```

## Building the image

To build this image, use the `Dockerfile` in the root of the [Doltgres
repository](https://github.com/dolthub/doltgresql/) with an optional build argument:

```shell
# Build the latest Doltgres version (automatically fetches the latest release)
$ docker build -t doltgres:latest .

# Build the latest Doltgres version (fetches the latest release)
$ docker build --build-arg DOLTGRES_VERSION=latest -t doltgres:latest .

# Build with a specific Doltgres version
$ docker build --build-arg DOLTGRES_VERSION=0.55.1 -t doltgres:0.55.1 .

# Build from local source code
$ docker build --build-arg DOLTGRES_VERSION=source -t doltgres:source .
```

## Connect to the server in the container from the host system

From the host system, to connect to a server running in a container, we need to map a port on the
host system to the port our server is running on in the container.

```bash
$ docker run -p 5432:5432 dolthub/doltgresql:latest
```

*Note*: if you have Postgres installed on this machine already, port `5432` will be in use. Either
choose a different port to map, or shut down Postgres.

Now connect with `psql` or another Postgres-compatible client.

```bash
$ PGPASSWORD=password psql --host 127.0.0.1 -U postgres
```

## Define configuration for the server

You can specify server configuration by providing your own `config.yaml` for the server to use as a
mounted volume. The image looks for a `config.yaml` file in the mounted directory
`/etc/doltgres/servercfg.d`. Place your desired `config.yaml` directory in a local file, then
provide it to `docker run` with the `-v` argument like this:

```shell
$ docker run -v ./doltgres_cfg:/etc/doltgres/servercfg.d -p 5432:5432 dolthub/doltgresql:latest
```

The data directory in the container is `/var/lib/doltgresql/`. To change this, provide the `PGDATA`
or `DOLTGRES_DATA` environment variable to `docker run`. This directory can also be a mounted
directory on the local machine.

```shell
$ docker run -e PGDATA=/path/to/doltgres/data -p 5432:5432 dolthub/doltgresql:latest
```

## Specifying a username and password

By default, the server creates a super-user named `postgres` with the password `password`. To change
this, provide the `DOLTGRES_USER` and `DOLTGRES_PASSWORD` environment variables to the `docker run`
command.

```shell
$ docker run -e DOLTGRES_USER=myuser -e DOLTGRES_PASSWORD=mypass -p 5432:5432 dolthub/doltgresql:latest
```

For convenience, `POSTGRES_USER` and `POSTGRES_PASSWORD` are accepted as aliases for these variables.

To create additional users, connect to the running database and issue `CREATE ROLE` and `GRANT`
statements as the super-user.

## Environment Variables

The Doltgres image supports the following environment variables:

- `DOLTGRES_USER`: The name of the super-user (default: `postgres`). `POSTGRES_USER` is an alias.
- `DOLTGRES_PASSWORD`: The password for the super-user (default: `password`). `POSTGRES_PASSWORD` is
  an alias.
- `DOLTGRES_DATA`: Specifies a path in the container to store database data, created if it doesn't
  exist (default: `/var/lib/doltgresql/`). `PGDATA` is an alias.
- `DOLTGRES_DB`: Specifies a database name to be created (default: none). The `postgres` database is
  still created.
