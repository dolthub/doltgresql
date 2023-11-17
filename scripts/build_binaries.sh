#!/bin/bash

set -e
set -o pipefail

script_dir=$(dirname "$0")
cd $script_dir/..

[ ! -z "$GO_BUILD_VERSION" ] || (echo "Must supply GO_BUILD_VERSION"; exit 1)

docker run --rm -v `pwd`:/src golang:"$GO_BUILD_VERSION"-bookworm /bin/bash -c '
set -e
set -o pipefail
apt-get update && apt-get install -y p7zip-full pigz
cd /src

BINS="doltgres"
OS_ARCH_TUPLES="windows-amd64 linux-amd64 linux-arm64 darwin-amd64 darwin-arm64"

for tuple in $OS_ARCH_TUPLES; do
  os=`echo $tuple | sed 's/-.*//'`
  arch=`echo $tuple | sed 's/.*-//'`
  o="out/doltgresql-$os-$arch"
  mkdir -p "$o/bin"
  mkdir -p "$o/licenses"
  cp Godeps/LICENSES "$o/"
  cp -r ./licenses "$o/licenses"
  cp LICENSE "$o/licenses"
  for bin in $BINS; do
    echo Building "$o/$bin"
    obin="$bin"
    if [ "$os" = windows ]; then
      obin="$bin.exe"
    fi
    CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags="-s -w" -o "$o/bin/$obin" .
  done
  if [ "$os" = windows ]; then
    (cd out && 7z a "doltgresql-$os-$arch.zip" "doltgresql-$os-$arch" && 7z a "doltgresql-$os-$arch.7z" "doltgresql-$os-$arch")
  else
    tar cf - -C out "doltgresql-$os-$arch" | pigz -9 > "out/doltgresql-$os-$arch.tar.gz"
  fi
done
'
