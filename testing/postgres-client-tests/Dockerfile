# syntax=docker/dockerfile:1
FROM golang:1.26.2-alpine AS golang_cgo
ENV CGO_ENABLED=1
ENV GO_LDFLAGS="-linkmode external -extldflags '-static'"
RUN apk add --no-cache build-base

# --- Build doltgres binary ---
FROM golang_cgo AS doltgres_build
RUN apk add --no-cache icu-dev icu-static
RUN apk update && apk add --no-cache bash
WORKDIR /root/building
COPY go.mod doltgresql/
WORKDIR doltgresql
RUN go mod download
COPY --exclude=testing/postgres-client-tests/ . .
WORKDIR /root/building/doltgresql/postgres/parser
RUN sh ./build.sh
WORKDIR /root/building/doltgresql/cmd/doltgres
RUN go build -ldflags "-linkmode external -extldflags '-static'" -o /build/bin/doltgres .

# --- Build Go postgres clients (pgx, lib/pq) ---
FROM golang:1.26.4 AS go_clients_build
COPY testing/postgres-client-tests/go/pgx/ /build/go/pgx/
WORKDIR /build/go/pgx
RUN go build -o /build/bin/pgx-test .
COPY testing/postgres-client-tests/go/libpq/ /build/go/libpq/
WORKDIR /build/go/libpq
RUN go build -o /build/bin/libpq-test .

# --- Build Rust sqlx client ---
FROM rust:1.96-slim-bookworm AS rust_clients_build
RUN apt-get update && apt-get install -y pkg-config libssl-dev && rm -rf /var/lib/apt/lists/*
COPY testing/postgres-client-tests/rust/ /build/rust/
WORKDIR /build/rust
RUN cargo build --release

# --- Build C libpq client ---
FROM debian:bookworm-slim AS c_clients_build
RUN apt-get update && apt-get install -y gcc make libpq-dev pkg-config && rm -rf /var/lib/apt/lists/*
COPY testing/postgres-client-tests/c/ /build/c/
WORKDIR /build/c
RUN make

# --- Build C++ libpqxx client ---
FROM debian:bookworm-slim AS cpp_clients_build
RUN apt-get update && apt-get install -y g++ make libpqxx-dev pkg-config && rm -rf /var/lib/apt/lists/*
COPY testing/postgres-client-tests/cpp/ /build/cpp/
WORKDIR /build/cpp
RUN make

# --- Build ODBC psqlODBC client ---
FROM debian:bookworm-slim AS odbc_clients_build
RUN apt-get update && apt-get install -y gcc make unixodbc-dev && rm -rf /var/lib/apt/lists/*
COPY testing/postgres-client-tests/odbc/ /build/odbc/
WORKDIR /build/odbc
RUN make

# --- Install Node postgres client deps ---
FROM node:22-bookworm-slim AS node_clients_build
COPY testing/postgres-client-tests/node/package.json /build/node/
COPY testing/postgres-client-tests/node/package-lock.json /build/node/
WORKDIR /build/node
RUN npm install
COPY testing/postgres-client-tests/node/ /build/node/

# --- Install Ruby pg gem ---
FROM ruby:3.4-bookworm AS ruby_clients_build
RUN apt-get update && apt-get install -y libpq-dev && rm -rf /var/lib/apt/lists/*
COPY testing/postgres-client-tests/ruby/Gemfile /build/ruby/
COPY testing/postgres-client-tests/ruby/Gemfile.lock /build/ruby/
WORKDIR /build/ruby
RUN bundle install
COPY testing/postgres-client-tests/ruby/ /build/ruby/

# --- Install Python deps ---
FROM python:3.14-slim-bookworm AS python_clients_build
RUN pip3 install --no-cache-dir --target /build/python-deps sqlalchemy==2.0.46
COPY testing/postgres-client-tests/python/ /build/python/
WORKDIR /build/python/

# --- Build .NET Npgsql client ---
FROM mcr.microsoft.com/dotnet/sdk:9.0-bookworm-slim AS dotnet_clients_build
COPY testing/postgres-client-tests/dotnet/ /build/dotnet/
WORKDIR /build/dotnet
RUN dotnet publish -c Release -o /build/output/

# --- Build Java R2DBC client ---
FROM maven:3.9-eclipse-temurin-17 AS r2dbc_clients_build
COPY testing/postgres-client-tests/r2dbc/ /build/r2dbc/
WORKDIR /build/r2dbc
RUN mvn package -q

# --- Build Elixir Postgrex client ---
FROM hexpm/elixir:1.16.3-erlang-24.2-debian-bookworm-20260610-slim AS elixir_clients_build
ENV ELIXIR_ERL_OPTIONS="+fnu"
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8
RUN apt-get update && apt-get install -y elixir && rm -rf /var/lib/apt/lists/*
COPY testing/postgres-client-tests/elixir/ /build/elixir/
WORKDIR /build/elixir
RUN mix local.hex --force && \
    mix local.rebar --force && \
    mix deps.get && \
    mix escript.build

# --- Build Swift PostgresNIO client ---
FROM swift:6.0-bookworm AS swift_clients_build
COPY testing/postgres-client-tests/swift/ /build/swift/
WORKDIR /build/swift
RUN swift build -c release --static-swift-stdlib

# --- Runtime ---
FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    curl \
    gnupg \
    lsb-release && \
    sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list' && \
    curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | gpg --dearmor -o /etc/apt/trusted.gpg.d/postgresql.gpg && \
    curl -fsSL https://deb.nodesource.com/setup_22.x | bash -

# java JDBC dependencies
COPY testing/postgres-client-tests/java/ /postgres-client-tests/java/
RUN apt-get install -y \
    openjdk-17-jdk && \
    curl -L -o /postgres-client-tests/java/postgresql-42.7.3.jar https://jdbc.postgresql.org/download/postgresql-42.7.3.jar

# perl dependencies
RUN apt-get install -y \
    python3 \
    python3-pip \
    python3-psycopg2 \
    perl \
    cpanminus \
    php \
    php-pgsql \
    r-base \
    libpq-dev \
    odbc-postgresql \
    g++ make libpqxx-dev pkg-config \
    gcc make unixodbc-dev \
    ca-certificates-java \
    bats \
    lsof \
    postgresql-client-15 \
    nodejs \
    elixir \
    libicu-dev \
    git && \
    update-ca-certificates -f && \
    rm -rf /var/lib/apt/lists/*

# install .NET
RUN curl -sSL https://dot.net/v1/dotnet-install.sh -o dotnet-install.sh \
    && chmod +x dotnet-install.sh \
    && ./dotnet-install.sh --channel 9.0 --runtime aspnetcore --install-dir /usr/share/dotnet \
    && ln -s /usr/share/dotnet/dotnet /usr/bin/dotnet \
    && rm dotnet-install.sh

# cpan dependencies
RUN cpanm --force DBD::Pg

# r dependencies
COPY testing/postgres-client-tests/r/ /postgres-client-tests/r/
RUN Rscript -e 'install.packages(c("RPostgres", "RPostgreSQL"), repos="https://cloud.r-project.org")'

ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64/
ENV GEM_HOME="/usr/local/bundle"
ENV PYTHONPATH=/usr/local/lib/python-deps
ENV ELIXIR_ERL_OPTIONS="+fnu"
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8

COPY --from=ruby_clients_build /usr/local/bin/ruby /usr/local/bin/
COPY --from=ruby_clients_build /usr/local/lib/ /usr/local/lib/
COPY --from=ruby_clients_build /usr/local/bundle/ /usr/local/bundle/
RUN ldconfig

COPY --from=doltgres_build /build/bin/doltgres /usr/local/bin/doltgres
COPY --from=go_clients_build /build/bin/pgx-test /build/bin/go/pgx-test
COPY --from=go_clients_build /build/bin/libpq-test /build/bin/go/libpq-test
COPY --from=rust_clients_build /build/rust/target/release/sqlx_exists_demo /build/bin/rust/sqlx_exists_demo
COPY --from=c_clients_build /build/c/postgres-c-connector-test /build/bin/c/postgres-c-connector-test
COPY --from=dotnet_clients_build /build/output /build/bin/dotnet
COPY --from=r2dbc_clients_build /build/r2dbc/target/r2dbc-test-1.0.jar /build/bin/r2dbc/r2dbc-test.jar
COPY --from=elixir_clients_build /build/elixir/postgrex-test /build/bin/elixir/postgrex-test
COPY --from=swift_clients_build /build/swift/.build/release/postgresnio-test /build/bin/swift/postgresnio-test
COPY --from=node_clients_build /build/node/ /postgres-client-tests/node/
COPY --from=python_clients_build /build/python/ /postgres-client-tests/python/
COPY --from=python_clients_build /build/python-deps/ /usr/local/lib/python-deps/
COPY --from=ruby_clients_build /build/ruby/ /postgres-client-tests/ruby/

COPY testing/postgres-client-tests/c/ /postgres-client-tests/c/
COPY testing/postgres-client-tests/cpp/ /postgres-client-tests/cpp/
COPY testing/postgres-client-tests/odbc/ /postgres-client-tests/odbc/
COPY testing/postgres-client-tests/drizzle/ /postgres-client-tests/drizzle/
COPY testing/postgres-client-tests/helpers.bash /postgres-client-tests/
COPY testing/postgres-client-tests/php/ /postgres-client-tests/php/
COPY testing/postgres-client-tests/perl/ /postgres-client-tests/perl/
COPY testing/postgres-client-tests/postgres-client-tests.bats /postgres-client-tests/
COPY testing/postgres-client-tests/postgres-client-tests-entrypoint.sh /postgres-client-tests/entrypoint.sh

WORKDIR /postgres-client-tests
ENTRYPOINT ["/postgres-client-tests/entrypoint.sh"]
