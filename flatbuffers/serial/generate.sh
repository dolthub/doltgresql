#!/bin/bash

set -eou pipefail
SRC=$(dirname ${BASH_SOURCE[0]})

GEN_DIR="$SRC/../gen/serial"

# cleanup old generated files
if [ ! -z "$(ls $GEN_DIR)" ]; then
    rm $GEN_DIR/*.go
fi

FLATC=${FLATC:-$SRC/../../../dolt/proto/third_party/flatbuffers/bazel-bin/flatc}

if [ ! -x "$FLATC" ]; then
  echo "$FLATC is not an executable. Did you remember to run 'bazel build //:flatc' in $(dirname $(dirname $FLATC))"
  exit 1
fi

# generate golang (de)serialization package
"$FLATC" -o $GEN_DIR --gen-onefile --filename-suffix "" --gen-mutable --go-namespace "serial" --go \
  collation.fbs \
  rootvalue.fbs

# prefix files with copyright header
for FILE in $GEN_DIR/*.go;
do
  mv $FILE "tmp.go"
  cat "copyright.txt" "tmp.go" >> $FILE
  rm "tmp.go"
done

# format and remove unused imports
goimports -w $GEN_DIR
