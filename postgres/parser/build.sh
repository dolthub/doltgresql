#!/usr/bin/env bash

# Set the working directory to the directory of the script's location
cd "$(cd -P -- "$(dirname -- "$0")" && pwd -P)"

# Generate sql-gen.y, which is modified from the sql.y file
mkdir -p parser/gen
set -euo pipefail; \
awk '/func.*sqlSymUnion/ {print $(NF - 1)}' parser/sql.y | \
sed -e 's/[]\/$*.^|[]/\\&/g' | \
sed -e "s/^/s_(type|token) <(/" | \
awk '{print $0")>_\\1 <union> /* <\\2> */_"}' > parser/gen/types_regex.tmp; \
sed -E -f parser/gen/types_regex.tmp < parser/sql.y | \
awk -f parser/replace_help_rules.awk | \
sed -Ee 's,//.*$,,g;s,/[*]([^*]|[*][^/])*[*]/, ,g;s/ +$//g' > parser/gen/sql-gen.y.tmp || rm parser/gen/sql-gen.y.tmp
mv -f parser/gen/sql-gen.y.tmp parser/gen/sql-gen.y
#rm parser/gen/types_regex.tmp

# Generate sql.go.tmp, which we will reference in later steps
set -euo pipefail; \
ret=$(cd parser/gen && go run golang.org/x/tools/cmd/goyacc -p sql -o sql.go.tmp sql-gen.y); \
if expr "$ret" : ".*conflicts" >/dev/null; then \
  echo "$ret"; exit 1; \
fi

# Generate tokens.go, which builds on sql.go.tmp
(echo "// GENERATED FILE DO NOT EDIT"; \
 echo; \
 echo "package lex"; \
 echo; \
 grep '^const [A-Z][_A-Z0-9]* ' parser/gen/sql.go.tmp) > lex/tokens.go.tmp || rm lex/tokens.go.tmp
mv -f lex/tokens.go.tmp lex/tokens.go

# Generate reserved_keywords.go, which builds on sql.y
awk -f parser/reserved_keywords.awk < parser/sql.y > lex/reserved_keywords.go.tmp || rm lex/reserved_keywords.go.tmp
mv -f lex/reserved_keywords.go.tmp lex/reserved_keywords.go
gofmt -s -w lex/reserved_keywords.go

# Generate keywords.go, which builds on sql.y
go run -tags all-keywords lex/all_keywords.go < parser/sql.y > lex/keywords.go.tmp || rm lex/keywords.go.tmp
mv -f lex/keywords.go.tmp lex/keywords.go
gofmt -s -w lex/keywords.go

# Generate help_messages.go, which builds on sql.y
awk -f parser/help.awk < parser/sql.y > parser/help_messages.go.tmp || rm parser/help_messages.go.tmp
mv -f parser/help_messages.go.tmp parser/help_messages.go
gofmt -s -w parser/help_messages.go

# Finalize sql.go from sql.go.tmp
(echo "// GENERATED FILE DO NOT EDIT"; \
 cat parser/gen/sql.go.tmp | \
 sed -E 's/^const ([A-Z][_A-Z0-9]*) =.*$/const \1 = lex.\1/g') > parser/sql.go.tmp || rm parser/sql.go.tmp
mv -f parser/sql.go.tmp parser/sql.go
go run golang.org/x/tools/cmd/goimports -w parser/sql.go
