#!/usr/bin/env bash
# Stands in for dolt on the oracle tests' Dolt side, which DOLT_REAL names: `init` makes a repository that starts as a
# database that a client creates in doltgres does, with an initial commit by Dolt's system account and then a
# "CREATE DATABASE" commit by the postgres user at the client's address, as its later commits are. Every other command
# runs dolt itself.
if [ "${1:-}" != init ]; then
  exec "$DOLT_REAL" "$@"
fi
shift
args=()
while [ $# -gt 0 ]; do
  case "$1" in
    --name|--email) shift 2 ;;
    --name=*|--email=*) shift ;;
    *) args+=("$1"); shift ;;
  esac
done
"$DOLT_REAL" init --name "Dolt System Account" --email doltuser@dolthub.com "${args[@]}" || exit $?
"$DOLT_REAL" config --local --add user.name postgres >/dev/null
"$DOLT_REAL" config --local --add user.email postgres@127.0.0.1 >/dev/null
"$DOLT_REAL" commit --allow-empty -m "CREATE DATABASE" --author "postgres <postgres@127.0.0.1>" >/dev/null
