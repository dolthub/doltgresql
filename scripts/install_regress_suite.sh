#!/bin/bash
# Installs Postgres' regression suite into <dir>/suite and a psql of the same major version at <dir>/psql, for
# `regress run`: install_regress_suite.sh <dir>
set -euo pipefail

PG_VERSION=18.6
PG_SHA256=555610c24d53e4316da5b7d3fc25c279d96856d5e0e23ee308c328c5fa881d9f
PG_MAJOR=${PG_VERSION%%.*}

dir=$1
mkdir -p "$dir"
archive="$dir/postgresql-$PG_VERSION.tar.bz2"
curl -fsSL -o "$archive" "https://ftp.postgresql.org/pub/source/v$PG_VERSION/postgresql-$PG_VERSION.tar.bz2"
echo "$PG_SHA256  $archive" | shasum -a 256 -c -
tar -xjf "$archive" -C "$dir" "postgresql-$PG_VERSION/src/test/regress"
rm -rf "$dir/suite"
mv "$dir/postgresql-$PG_VERSION/src/test/regress" "$dir/suite"
rm -rf "$dir/postgresql-$PG_VERSION" "$archive"

if [ "$(uname)" = "Linux" ]; then
  sudo apt-get install -y postgresql-common
  sudo /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh -y
  sudo apt-get install -y "postgresql-client-$PG_MAJOR"
  ln -sf "/usr/lib/postgresql/$PG_MAJOR/bin/psql" "$dir/psql"
else
  brew install "libpq@$PG_MAJOR"
  ln -sf "$(brew --prefix "libpq@$PG_MAJOR")/bin/psql" "$dir/psql"
fi
